//go:build windows

package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type boardFake struct {
	calls                     []string
	startRC, usageRC, closeRC uint32
	model                     string
	closeCount                int
	extraFail                 bool
}

func (f *boardFake) SendReadPTP(op uint32, args []uint32) ([]byte, uint32, error) {
	f.calls = append(f.calls, fmt.Sprintf("%X:%X", op, args))
	if op == PTP_GET_DEVICE_INFO {
		b := l10TestPTP(f.model, "PRIVATE_SERIAL1234", nil)
		b = bytes.ReplaceAll(b, []byte{'1', 0, '.', 0, '2', 0}, []byte{'2', 0, '.', 0, '4', 0})
		return b, PTP_RC_OK, nil
	}
	if op == 0x1004 || op == 0x1005 {
		return nil, 0x2005, nil
	}
	if args[0] == SETUP_INFO_TAG {
		return l10TestReply(146, SETUP_INFO_REPLY), f.usageRC, nil
	}
	if f.extraFail {
		return nil, 0x2005, nil
	}
	n := 2
	return l10TestReply(n, args[0]+1), PTP_RC_OK, nil
}
func (f *boardFake) SendSession(op uint32) (uint32, error) {
	f.calls = append(f.calls, fmt.Sprintf("%X", op))
	if op == 0x9102 {
		return f.startRC, nil
	}
	return f.closeRC, nil
}
func (f *boardFake) Close() { f.closeCount++ }
func TestBoardSessionOwnershipAndPartialResults(t *testing.T) {
	for _, tc := range []struct {
		name                string
		start, usage, close uint32
		valid               bool
		want                []string
	}{
		{"success", 0x2001, 0x2001, 0x2001, true, []string{"1001:[]", "9102", "9414:[15C00010]", "9103"}},
		{"usage_error", 0x2001, 0x2002, 0x2001, false, []string{"1001:[]", "9102", "9414:[15C00010]", "9103"}},
		{"start_error", 0x2002, 0x2001, 0x2001, false, []string{"1001:[]", "9102"}},
		{"close_error", 0x2001, 0x2001, 0x2002, true, []string{"1001:[]", "9102", "9414:[15C00010]", "9103"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := &boardFake{model: "DC-S1", startRC: tc.start, usageRC: tc.usage, closeRC: tc.close}
			sentBeforeClose := false
			r := boardQuery(f, true, func(v boardReply) {
				if v.Kind == "usage_complete" {
					sentBeforeClose = v.Valid && len(f.calls) == 3
				}
			})
			if r.Valid != tc.valid || !reflect.DeepEqual(f.calls, tc.want) || f.closeCount != 1 || r.Result == nil {
				t.Fatalf("%+v / %v", r, f.calls)
			}
			if tc.valid && !sentBeforeClose {
				t.Fatal("valid counters were not streamed before cleanup")
			}
		})
	}
	f := &boardFake{model: "DC-S5", usageRC: 0x2001}
	r := boardQuery(f, true, func(boardReply) {})
	if r.Valid || len(f.calls) != 1 {
		t.Fatal("session trial escaped model gate")
	}
	f = &boardFake{model: "DC-S5", usageRC: 0x2001}
	r = boardQuery(f, false, func(boardReply) {})
	if !r.Valid || len(f.calls) != 2 {
		t.Fatal("legacy path changed")
	}
}
func TestBoardOptionalQueriesDoNotLoseCounters(t *testing.T) {
	f := &boardFake{model: "DC-S5", usageRC: 0x2001, extraFail: true}
	streamedFirst := false
	r := boardQueryOptions(f, false, true, func(v boardReply) {
		if v.Kind == "usage_complete" {
			streamedFirst = v.Valid && len(f.calls) == 2
		}
	})
	if !streamedFirst {
		t.Fatal("optional queries blocked initial counter delivery")
	}
	if !r.Valid || r.Diagnostic.FailedStage != "card_status" {
		t.Fatal(r)
	}
	f = &boardFake{model: "DC-S5", usageRC: 0x2002}
	r = boardQueryOptions(f, false, true, func(boardReply) {})
	if r.Valid || r.Camera == nil || r.Diagnostic.FailedStage != "usage" {
		t.Fatal("metadata incorrectly coupled to counters", r)
	}
	for _, call := range f.calls {
		if call == "9414:[16000010]" || call == "9414:[18000010]" || call == "9414:[12000010]" || call == "9414:[12000016]" {
			t.Fatal("removed optional query was sent")
		}
	}

}

func TestBoardOfflineNavigationAndFailedQueryActions(t *testing.T) {
	oldEnabled, oldView, oldState, oldLatest := boardEnabled, currentView, appState, boardLatest
	defer func() { boardEnabled = oldEnabled; currentView = oldView; appState = oldState; boardLatest = oldLatest }()
	boardEnabled = true
	appState = "error"
	boardLatest = boardReply{Diagnostic: newBoardDiagnostic(false)}
	boardLatest.Diagnostic.FailedStage = "usage"
	for _, view := range []string{"main", "usage_history", "diagnostics", "storage"} {
		currentView = view
		a := visibleActions()
		tabs := 0
		for i, x := range a {
			if strings.HasPrefix(x.ID, "tab_") {
				tabs++
			}
			for j := i + 1; j < len(a); j++ {
				y := a[j]
				if x.Rect.L < y.Rect.R && y.Rect.L < x.Rect.R && x.Rect.T < y.Rect.B && y.Rect.T < x.Rect.B {
					t.Fatalf("overlapping actions %s %s on %s", x.ID, y.ID, view)
				}
			}
		}
		wantTabs := 4
		if tabs != wantTabs {
			t.Fatal("offline navigation missing", view)
		}
		if view == "diagnostics" {
			found := false
			for _, x := range a {
				if x.ID == "board_diag_save" && x.Enabled {
					found = true
				}
			}
			if !found {
				t.Fatal("diagnostic export unavailable after failure")
			}
		}
	}
	for _, s := range []string{"", "unknown", "********1234", "00000000", "N/A"} {
		if boardReliableSerial(s) {
			t.Fatal("unreliable serial accepted", s)
		}
	}
}
func TestBoardHistoryIsolationAndExternalGuard(t *testing.T) {
	path := filepath.Join(t.TempDir(), "store.json")
	s, e := boardLoad(path)
	if e != nil {
		t.Fatal(e)
	}
	r := UsageResult{Model: "DC-S5", Firmware: "2.9", Serial: "PRIVATE_FULL1234", Timestamp: time.Now(), Shutter: 100, PowerWake: 20}
	for i := 0; i < 2; i++ {
		if e = boardAppend(&s, r); e != nil {
			t.Fatal(e)
		}
	}
	if s.Records[0].Body != s.Records[1].Body || s.Records[0].ID == s.Records[1].ID {
		t.Fatal("body or timestamp collision")
	}
	r.Serial = "ANOTHER_BODY1234"
	if e = boardAppend(&s, r); e != nil {
		t.Fatal(e)
	}
	if s.Records[2].Body == s.Records[0].Body {
		t.Fatal("same-model bodies merged")
	}
	if e = boardPersist(path, s); e != nil {
		t.Fatal(e)
	}
	b, _ := os.ReadFile(path)
	if bytes.Contains(b, []byte("PRIVATE_FULL")) {
		t.Fatal("full serial persisted")
	}
	export := filepath.Join(t.TempDir(), "external.json")
	if e = boardExternalSave(&s, export, true); e != nil {
		t.Fatal(e)
	}
	b, _ = os.ReadFile(export)
	if bytes.Contains(b, []byte(s.Key)) || bytes.Contains(b, []byte(s.Records[0].Body)) {
		t.Fatal("local identity exported")
	}
	if e = boardExternalSave(&s, export, false); e != nil {
		t.Fatal(e)
	}
	os.WriteFile(export, []byte("external edit"), 0600)
	if e = boardExternalSave(&s, export, false); e == nil {
		t.Fatal("external edits overwritten")
	}
	b, _ = os.ReadFile(export)
	if string(b) != "external edit" {
		t.Fatal("external data lost")
	}
	a, c := s.Records[0], s.Records[1]
	c.Shutter += 3
	sh, _, ok := boardDifference(a, c)
	if !ok || sh != 3 {
		t.Fatal("delta")
	}
	c.Firmware = "3.0"
	if _, _, ok = boardDifference(a, c); ok {
		t.Fatal("firmware boundary ignored")
	}
	c.Firmware = a.Firmware
	c.Shutter = 1
	if _, _, ok = boardDifference(a, c); ok {
		t.Fatal("counter rollback ignored")
	}
	os.WriteFile(path, []byte("corrupt"), 0600)
	if _, e = boardLoad(path); e == nil {
		t.Fatal("corrupt history accepted")
	}
	b, _ = os.ReadFile(path)
	if string(b) != "corrupt" {
		t.Fatal("corrupt original overwritten")
	}
}
func TestBoardDiagnosticPrivacyAndLanguages(t *testing.T) {
	old := atomic.LoadUint32(&languageMode)
	defer atomic.StoreUint32(&languageMode, old)
	for _, m := range []uint32{languageEnglish, languageKorean, languageJapanese, languageSpanish} {
		atomic.StoreUint32(&languageMode, m)
		for k, row := range boardStrings {
			if boardText(k) == "" || row[0] == "" {
				t.Fatal(k)
			}
		}
	}
	d := newBoardDiagnostic(false)
	d.Model = "PRIVATE_SERIAL"
	d.Firmware = "C:\\Users\\Secret"
	d.Policy = "secret"
	d.FailedStage = "SECRET"
	d.Steps = []boardStep{{Stage: "usage", Status: "failed", PTP: "0x2002", Code: "PRIVATE_SERIAL"}, {Stage: "SECRET", Status: "ok"}}
	b, e := boardDiagnosticZIP(d)
	if e != nil {
		t.Fatal(e)
	}
	z, e := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if e != nil {
		t.Fatal(e)
	}
	for _, f := range z.File {
		r, _ := f.Open()
		data, _ := io.ReadAll(r)
		r.Close()
		for _, bad := range []string{"PRIVATE_SERIAL", "Secret", "SECRET"} {
			if strings.Contains(string(data), bad) {
				t.Fatal("private diagnostic field escaped", f.Name)
			}
		}
		if f.Name == "diagnostic.json" {
			var out boardDiagnostic
			if e = json.Unmarshal(data, &out); e != nil || len(out.Steps) != 1 {
				t.Fatal(e)
			}
		}
	}
}
