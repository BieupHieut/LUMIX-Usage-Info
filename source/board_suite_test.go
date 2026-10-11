//go:build windows

package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"
)

type suiteFake struct {
	boardFake
	usageCodes    []uint32
	fatalTag      uint32
	differentBody bool
}

func (f *suiteFake) SendReadPTP(op uint32, args []uint32) ([]byte, uint32, error) {
	if len(args) > 0 && args[0] == f.fatalTag {
		f.calls = append(f.calls, "disconnect")
		return nil, 0, fmt.Errorf("transport_error")
	}
	if len(args) > 0 && args[0] == SETUP_INFO_TAG && len(f.usageCodes) > 0 {
		f.usageRC = f.usageCodes[0]
		f.usageCodes = f.usageCodes[1:]
	}
	b, rc, e := f.boardFake.SendReadPTP(op, args)
	if f.differentBody && op == PTP_GET_DEVICE_INFO {
		b = bytes.ReplaceAll(b, []byte{'1', 0, '2', 0, '3', 0, '4', 0}, []byte{'9', 0, '9', 0, '9', 0, '9', 0})
	}
	return b, rc, e
}
func suiteCamera() *suiteFake {
	return &suiteFake{boardFake: boardFake{model: "DC-S1", startRC: 0x2001, closeRC: 0x2001, usageRC: 0x2002}}
}
func runSuiteFake(a, b *suiteFake) (boardReply, int) {
	pauses := 0
	opens := 0
	out := boardSuite(func(session bool) (boardTransport, error) {
		opens++
		if opens == 3 {
			return suiteCamera(), nil
		}
		if session {
			return b, nil
		}
		return a, nil
	}, func(boardReply) {}, func(d time.Duration) {
		pauses++
		if d != 750*time.Millisecond {
			panic("unbounded wait")
		}
	})
	return out, pauses
}
func TestBoardSuiteAdaptiveBranches(t *testing.T) {
	for _, tc := range []struct {
		name        string
		codes       []uint32
		extraFail   bool
		wantQueries int
		pause       int
		valid       bool
	}{
		{"all_general_error", []uint32{0x2002, 0x2002, 0x2002}, true, 2, 1, false},
		{"settled_success", []uint32{0x2002, 0x2001}, false, 2, 1, true},
		{"immediate_success", []uint32{0x2001}, false, 1, 0, true},
		{"unsupported_no_controls", []uint32{0x2005}, true, 1, 0, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a, b := suiteCamera(), suiteCamera()
			b.usageCodes = tc.codes
			b.extraFail = tc.extraFail
			out, pauses := runSuiteFake(a, b)
			n := 0
			for _, c := range b.calls {
				if c == "9414:[16000010]" || c == "9414:[18000010]" {
					t.Fatal("removed optional query was sent by suite")
				}
				if c == "9414:[15C00010]" {
					n++
				}
			}
			if n != tc.wantQueries || pauses != tc.pause || out.Valid != tc.valid || out.Diagnostic.Outcome != "completed" {
				t.Fatalf("%+v / %v pauses=%d", out, b.calls, pauses)
			}
			if a.closeCount != 1 || b.closeCount != 1 || b.calls[len(b.calls)-1] != "9103" {
				t.Fatal("cleanup", a.calls, b.calls)
			}
			if !reflect.DeepEqual(a.calls, []string{"1001:[]", "9414:[15C00010]"}) {
				t.Fatal("baseline mutated", a.calls)
			}
		})
	}
}
func TestBoardSuiteStopAndOwnership(t *testing.T) {
	a, b := suiteCamera(), suiteCamera()
	b.startRC = 0x2002
	out, _ := runSuiteFake(a, b)
	if out.Diagnostic.Outcome != "stopped" || !reflect.DeepEqual(b.calls, []string{"1001:[]", "9102"}) || b.closeCount != 1 {
		t.Fatal(out, b.calls)
	}
	a, b = suiteCamera(), suiteCamera()
	b.fatalTag = SETUP_INFO_TAG
	out, _ = runSuiteFake(a, b)
	if !reflect.DeepEqual(b.calls, []string{"1001:[]", "9102", "disconnect", "9103"}) || out.Diagnostic.Outcome != "stopped" {
		t.Fatal(out, b.calls)
	}
	a, b = suiteCamera(), suiteCamera()
	b.differentBody = true
	out, _ = runSuiteFake(a, b)
	if out.Diagnostic.Outcome != "stopped" || len(b.calls) != 1 {
		t.Fatal("wrong body queried", out, b.calls)
	}
	a, b = suiteCamera(), suiteCamera()
	a.model = "DC-S5"
	out, _ = runSuiteFake(a, b)
	if len(a.calls) != 1 || len(b.calls) != 0 || a.closeCount != 1 {
		t.Fatal("model gate", out, a.calls, b.calls)
	}
	a, b = suiteCamera(), suiteCamera()
	b.closeRC = 0x2002
	out, _ = runSuiteFake(a, b)
	if out.Diagnostic.Outcome != "stopped" {
		t.Fatal("cleanup false success", out)
	}
	a, b = suiteCamera(), suiteCamera()
	a.usageRC = 0x2001
	b.fatalTag = SETUP_INFO_TAG
	out, _ = runSuiteFake(a, b)
	if !out.Valid || out.Result == nil {
		t.Fatal("lost baseline success")
	}
}
func TestBoardSuiteZIPPrivacyAndCompleteness(t *testing.T) {
	a, b := suiteCamera(), suiteCamera()
	out, _ := runSuiteFake(a, b)
	out.Diagnostic.Steps[0].Scenario = "PRIVATE_PATH"
	out.Diagnostic.Steps[0].Parameters = []uint32{0xdeadbeef}
	out.Diagnostic.Steps[0].Transfer = &boardTransferTrace{Phase: "PRIVATE_SERIAL", EndCode: "PRIVATE_PATH", ResponseParams: []uint32{0xdeadbeef}}
	out.Diagnostic.ActiveStage = "PRIVATE_PATH"
	payload, e := boardDiagnosticZIP(out.Diagnostic)
	if e != nil {
		t.Fatal(e)
	}
	z, e := zip.NewReader(bytes.NewReader(payload), int64(len(payload)))
	if e != nil {
		t.Fatal(e)
	}
	if len(z.File) != 3 {
		t.Fatal("missing summary")
	}
	for _, f := range z.File {
		r, _ := f.Open()
		v, _ := io.ReadAll(r)
		r.Close()
		if strings.Contains(string(v), "PRIVATE_") {
			t.Fatal("privacy", f.Name)
		}
		if f.Name == "diagnostic.json" {
			var d boardDiagnostic
			if e = json.Unmarshal(v, &d); e != nil {
				t.Fatal(e)
			}
			if len(d.Steps) != len(out.Diagnostic.Steps) || d.Policy != "s1-comprehensive-v1" || d.Outcome != "completed" {
				t.Fatal("lost suite results", d)
			}
			if len(d.Steps[0].Parameters) > 0 || len(d.Steps[0].Transfer.ResponseParams) > 0 {
				t.Fatal("unexpected params exported")
			}
		}
	}
}
func TestBoardResponseParametersCOM(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	hr, _, _ := procCoInitializeEx.Call(0, COINIT_APARTMENTTHREADED)
	if hrFailed(hr) {
		t.Fatal(hrString(hr))
	}
	defer procCoUninitialize.Call()
	p, e := newValues()
	if e != nil {
		t.Fatal(e)
	}
	defer release(p)
	if v := boardResponseParameters(p); v != nil {
		t.Fatal("missing parameter not optional")
	}
	for _, n := range []int{0, 1, 5, 6} {
		a := make([]uint32, n)
		for i := range a {
			a[i] = uint32(0x10001 + i)
		}
		c, e := newUI4Collection(a)
		if e != nil {
			t.Fatal(e)
		}
		e = valuesSetPVC(p, &boardResponseParamsKey, c)
		release(c)
		if e != nil {
			t.Fatal(e)
		}
		got := boardResponseParameters(p)
		if n > 5 {
			if got != nil {
				t.Fatal("unbounded collection")
			}
		} else if len(got) != len(a) || n > 0 && !reflect.DeepEqual(got, a) {
			t.Fatal(got, a)
		}
	}
}

func TestBoardDiagnosticBundlePreservesPriorAttempts(t *testing.T) {
	a, b := suiteCamera(), suiteCamera()
	first, _ := runSuiteFake(a, b)
	a, b = suiteCamera(), suiteCamera()
	b.usageRC = 0x2001
	last, _ := runSuiteFake(a, b)
	first.Diagnostic.Model = "PRIVATE_SERIAL"
	payload, e := boardDiagnosticBundle(last.Diagnostic, []boardDiagnostic{first.Diagnostic, last.Diagnostic})
	if e != nil {
		t.Fatal(e)
	}
	z, e := zip.NewReader(bytes.NewReader(payload), int64(len(payload)))
	if e != nil {
		t.Fatal(e)
	}
	if len(z.File) != 7 {
		t.Fatal("expected latest + prior + index", len(z.File))
	}
	found := false
	for _, f := range z.File {
		r, _ := f.Open()
		v, _ := io.ReadAll(r)
		r.Close()
		if bytes.Contains(v, []byte("PRIVATE_SERIAL")) {
			t.Fatal("prior attempt privacy")
		}
		if f.Name == "previous_attempts/01/diagnostic.json" {
			found = true
			var d boardDiagnostic
			json.Unmarshal(v, &d)
			if !d.Started.Equal(first.Diagnostic.Started) {
				t.Fatal("overwrote first attempt")
			}
		}
	}
	if !found {
		t.Fatal("missing prior attempt")
	}
}
