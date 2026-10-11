//go:build windows

package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"runtime"
	"testing"
	"time"
	"unsafe"
)

// Exercise real Windows COM values without opening or querying any device.
func TestBoardWPDSessionParameters(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	hr, _, _ := procCoInitializeEx.Call(0, COINIT_APARTMENTTHREADED)
	if hrFailed(hr) {
		t.Fatalf("COM init: %s", hrString(hr))
	}
	defer procCoUninitialize.Call()
	for _, op := range []uint32{0x9102, 0x9103} {
		p, e := newSessionCommand(op)
		if e != nil {
			t.Fatal(e)
		}
		defer release(p)
		code, e := valuesGetUI4(p, &PK_MTP_OPERATION_CODE)
		if e != nil || code != op {
			t.Fatalf("opcode %X %v", code, e)
		}
		access, e := valuesGetUI4(p, &PK_IOCTL_ACCESS)
		if e != nil || access != 3 {
			t.Fatalf("access %d %v", access, e)
		}
		command, e := valuesGetUI4(p, &PK_COMMON_COMMAND_ID)
		if e != nil || command != 12 {
			t.Fatalf("command %d %v", command, e)
		}
		var args []uint32
		if op == 0x9102 {
			args = []uint32{0x10001}
		}
		var got unsafe.Pointer
		hr = comCall(p, 34, uintptr(unsafe.Pointer(&PK_MTP_OPERATION_PARAMS)), uintptr(unsafe.Pointer(&got)))
		if hrFailed(hr) {
			t.Fatalf("Get collection: %s", hrString(hr))
		}
		defer release(got)
		var count uint32
		hr = comCall(got, 3, uintptr(unsafe.Pointer(&count)))
		if hrFailed(hr) || int(count) != len(args) {
			t.Fatalf("count %d %s", count, hrString(hr))
		}
		for i, expected := range args {
			var v PROPVARIANT
			hr = comCall(got, 4, uintptr(i), uintptr(unsafe.Pointer(&v)))
			if hrFailed(hr) || v.Vt != VT_UI4 || v.Value != uint64(expected) {
				t.Fatalf("param %+v %s", v, hrString(hr))
			}
		}
		t.Logf("op=0x%04X command=12 params=%v: Windows COM round-trip passed", op, args)
	}
	for _, op := range []uint32{0x1002, 0x9414, 0x1016, 0x9104} {
		if p, e := newSessionCommand(op); e == nil || p != nil {
			t.Fatalf("unexpected accepted session command %X", op)
		}
	}
	for _, session := range []bool{false, true} {
		p, e := newValues()
		if e != nil {
			t.Fatal(e)
		}
		defer release(p)
		if e = boardSetDeviceAccess(p, session); e != nil {
			t.Fatal(e)
		}
		access, e := valuesGetUI4(p, &boardReadAccess)
		expected := uint32(0x80000000)
		if session {
			expected = 0xc0000000
		}
		if e != nil || access != expected {
			t.Fatalf("device access %X %v", access, e)
		}
	}
	p, e := setupCommandValues(CMD_EXECUTE_DATA_TO_READ)
	if e != nil {
		t.Fatal(e)
	}
	defer release(p)
	access, e := valuesGetUI4(p, &PK_IOCTL_ACCESS)
	if e != nil || access != 1 {
		t.Fatalf("read IOCTL changed %d %v", access, e)
	}
	if _, e := (&WPDDevice{}).SendSession(0x9102); e == nil {
		t.Fatal("session command allowed without transport permission")
	}
}

func TestBoardWPDErrorEvidence(t *testing.T) {
	d := newBoardDiagnostic(true)
	d.step("open_session", "failed", 0, fmt.Errorf("context: %w", &boardWPDError{Origin: "driver_hresult", Err: fmt.Errorf("WPD command failed: 0x80070057"), PTP: 0x2002}), time.Now())
	if s := d.Steps[0]; s.ErrorSource != "driver_hresult" || s.PTP != "0x2002" || s.Code != "0X80070057" {
		t.Fatalf("lost evidence: %+v", s)
	}
	// Export must preserve known origins but strip arbitrary strings/paths.
	d.Steps = append(d.Steps, boardStep{Stage: "usage", Status: "failed", ErrorSource: "PRIVATE_DEVICE_PATH"})
	b, e := boardDiagnosticZIP(d)
	if e != nil {
		t.Fatal(e)
	}
	if bytes.Contains(b, []byte("PRIVATE_DEVICE_PATH")) {
		t.Fatal("privacy leak")
	}
	z, e := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if e != nil {
		t.Fatal(e)
	}
	for _, f := range z.File {
		if f.Name == "diagnostic.json" {
			rd, e := f.Open()
			if e != nil {
				t.Fatal(e)
			}
			payload, e := io.ReadAll(rd)
			rd.Close()
			if e != nil {
				t.Fatal(e)
			}
			var got boardDiagnostic
			if e = json.Unmarshal(payload, &got); e != nil {
				t.Fatal(e)
			}
			if got.Steps[0].ErrorSource != "driver_hresult" || got.Steps[1].ErrorSource != "" || got.DeviceAccess != d.DeviceAccess {
				t.Fatalf("bad export: %+v", got)
			}
		}
	}
}
