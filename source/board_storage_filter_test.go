//go:build windows

package main

import (
	"archive/zip"
	"bytes"
	"encoding/binary"
	"io"
	"strings"
	"testing"
)

type storageIDFake struct {
	statusFake
	ids []uint32
}

func (f *storageIDFake) SendReadPTP(op uint32, args []uint32) ([]byte, uint32, error) {
	if op == 0x1004 {
		f.calls = append(f.calls, "1004")
		b := make([]byte, 4+4*len(f.ids))
		binary.LittleEndian.PutUint32(b, uint32(len(f.ids)))
		for i, id := range f.ids {
			binary.LittleEndian.PutUint32(b[4+4*i:], id)
		}
		return b, 0x2001, nil
	}
	return f.statusFake.SendReadPTP(op, args)
}
func TestStorageObservedIDsSkipAndKeepCounters(t *testing.T) {
	for _, ids := range [][]uint32{{0x10000, 0x20000}, {0x10000, 0x20001}, {0x10001, 0x20001}} {
		f := &storageIDFake{statusFake: statusFake{boardFake: boardFake{model: "DC-S1RM2", usageRC: 0x2001}}, ids: ids}
		out := boardQueryOptions(f, false, true, func(boardReply) {})
		if !out.Valid || out.Diagnostic.FailedStage != "" || f.closeCount != 1 || out.Camera == nil || !out.Camera.CardsRead {
			t.Fatal(out, f.calls)
		}
		wantQueries, skipped := 0, 0
		for i, id := range ids {
			s := out.Camera.Stores[i]
			if s.ID != id {
				t.Fatal("returned ID changed", s.ID, id)
			}
			if id&0xffff == 0 {
				skipped++
				if !s.InfoSkipped || s.Read || s.spaceKnown() {
					t.Fatal("invalid store reported as capacity", s)
				}
			} else {
				wantQueries++
				if s.InfoSkipped || !s.spaceKnown() || s.Free != 96000000000 {
					t.Fatal("valid storage stopped working", s)
				}
			}
		}
		gotQueries := 0
		for _, call := range f.calls {
			if strings.HasPrefix(call, "1005:") {
				gotQueries++
			}
			for _, tag := range []string{"16000010", "18000010", "12000010", "12000016"} {
				if strings.Contains(call, tag) {
					t.Fatal("removed optional request sent", call)
				}
			}
		}
		if gotQueries != wantQueries {
			t.Fatal("unexpected storage queries", f.calls)
		}
		skipSteps := 0
		for _, s := range out.Diagnostic.Steps {
			if s.Stage == "storage_info" && s.Status == "skipped" {
				skipSteps++
				if s.Code != "storage_not_addressable" || s.PTP != "" || s.Transfer != nil || s.Opcode != 0x1005 {
					t.Fatal("skip fabricated camera result", s)
				}
			}
		}
		if skipSteps != skipped {
			t.Fatal("missing skip evidence")
		}
		payload, e := boardDiagnosticZIP(out.Diagnostic)
		if e != nil {
			t.Fatal(e)
		}
		z, e := zip.NewReader(bytes.NewReader(payload), int64(len(payload)))
		if e != nil {
			t.Fatal(e)
		}
		for _, entry := range z.File {
			if entry.Name == "diagnostic.json" {
				r, _ := entry.Open()
				b, _ := io.ReadAll(r)
				r.Close()
				if skipped > 0 && (!bytes.Contains(b, []byte(`"info_skipped": true`)) || !bytes.Contains(b, []byte("storage_not_addressable"))) {
					t.Fatal("skip reason lost on export")
				}
				for _, removed := range []string{"usb_raw", "battery_raw", "camera_state_code", "detailed_free_spaces", "PRIVATE_"} {
					if bytes.Contains(b, []byte(removed)) {
						t.Fatal("unexpected diagnostic field", removed)
					}
				}
			}
		}
	}
}
func TestStorageFailureStillVisible(t *testing.T) {
	f := &storageIDFake{statusFake: statusFake{boardFake: boardFake{model: "DC-S1RM2", usageRC: 0x2001}, storageFail: true}, ids: []uint32{0x10001}}
	out := boardQueryOptions(f, false, true, func(boardReply) {})
	if !out.Valid || out.Diagnostic.FailedStage != "storage_info" || out.Camera.Stores[0].InfoSkipped {
		t.Fatal("real failure hidden", out)
	}
}
func TestBoardWelcomeAndDiagnosticNavigation(t *testing.T) {
	oldEnabled, oldView, oldResult, oldLatest := boardEnabled, currentView, lastResult, boardLatest
	oldMenu, oldBusy, oldPNG, oldReport := languageMenuOpen, probeBusy, pngBusy, reportBusy
	defer func() {
		boardEnabled, currentView, lastResult, boardLatest = oldEnabled, oldView, oldResult, oldLatest
		languageMenuOpen, probeBusy, pngBusy, reportBusy = oldMenu, oldBusy, oldPNG, oldReport
	}()
	boardEnabled = true
	currentView = "main"
	lastResult = UsageResult{}
	boardLatest = boardReply{}
	languageMenuOpen, probeBusy, pngBusy, reportBusy = false, false, false, false
	for _, a := range visibleActions() {
		if strings.HasPrefix(a.ID, "board_welcome_") && a.ID != "board_welcome_read" || a.ID == "board_serial" {
			t.Fatal("hidden dashboard control on welcome", a)
		}
	}
	if !boardClick(300, 705) || currentView != "usage_history" {
		t.Fatal("offline history inaccessible")
	}
	currentView = "main"
	if !boardClick(540, 705) || currentView != "diagnostics" {
		t.Fatal("failure help inaccessible")
	}
	boardLatest.Diagnostic = newBoardDiagnostic(false)
	if !boardClick(boardRects["diag_raw"].L+10, boardRects["diag_raw"].T+10) || currentView != "query_raw" {
		t.Fatal("diagnostic details inaccessible")
	}
	if !boardClick(backRect.L+15, backRect.T+15) || currentView != "diagnostics" {
		t.Fatal("diagnostic details back failed")
	}
	currentView = "main"
	lastResult = UsageResult{Model: "DC-S1RM2", Serial: "PRIVATE1234"}
	boardLatest.Valid = true
	for _, view := range []string{"main", "diagnostics", "card_info", "query_raw"} {
		currentView = view
		a := visibleActions()
		for i, x := range a {
			for _, y := range a[i+1:] {
				if x.Rect.L < y.Rect.R && y.Rect.L < x.Rect.R && x.Rect.T < y.Rect.B && y.Rect.T < x.Rect.B {
					t.Fatal("overlapping controls", view, x.ID, y.ID)
				}
			}
		}
	}
}
