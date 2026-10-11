//go:build windows

package main

import (
	"encoding/binary"
	"fmt"
	"strings"
	"time"
)

// The suite changes ordering, not command values. All reads are already used by
// the app. Each branch is bounded; no reset, settings write or service-mode entry.
var boardSuiteScenarios = map[string]bool{"baseline": true, "session": true, "controls": true, "after_controls": true, "settled": true, "post_session": true, "cleanup": true}

type boardTransferTrace struct {
	Phase          string   `json:"phase"`
	Total          uint64   `json:"expected_bytes"`
	TotalKnown     bool     `json:"size_available"`
	Received       int      `json:"received_bytes"`
	Reads          int      `json:"read_calls"`
	EndAttempted   bool     `json:"end_attempted"`
	EndPTP         uint32   `json:"end_ptp,omitempty"`
	EndCode        string   `json:"end_error,omitempty"`
	ResponseParams []uint32 `json:"response_parameters,omitempty"`
}

func (d *WPDDevice) boardTrace() boardTransferTrace { return d.trace }

func boardDeviceOperations(b []byte) []uint16 {
	if len(b) < 10 {
		return nil
	}
	off := 8
	if _, ok := readPTPString(b, &off); !ok || off+6 > len(b) {
		return nil
	}
	off += 2
	n := int(binary.LittleEndian.Uint32(b[off:]))
	off += 4
	if n > 1024 || n > (len(b)-off)/2 {
		return nil
	}
	a := make([]uint16, n)
	for i := range a {
		a[i] = binary.LittleEndian.Uint16(b[off+i*2:])
	}
	return a
}

func boardSuite(open func(bool) (boardTransport, error), emit func(boardReply), pause func(time.Duration)) (out boardReply) {
	out = boardReply{Kind: "suite_running", Diagnostic: newBoardDiagnostic(false)}
	out.Diagnostic.Outcome = "stopped"
	out.Diagnostic.Schema = 3
	out.Diagnostic.Policy = "s1-comprehensive-v1"
	out.Diagnostic.DeviceAccess = "GENERIC_READ then GENERIC_READ|GENERIC_WRITE (session transport only)"
	var dev boardTransport
	owned := false
	scenario := "baseline"
	send := func() { v := out; v.Kind = "suite_running"; emit(v) }
	step := func(stage, status string, rc uint32, e error, t time.Time, op uint32, args []uint32) {
		out.Diagnostic.step(stage, status, rc, e, t)
		s := &out.Diagnostic.Steps[len(out.Diagnostic.Steps)-1]
		s.Scenario = scenario
		s.Opcode = op
		s.Parameters = append([]uint32(nil), args...)
		if op != 0 {
			if tr, ok := dev.(interface{ boardTrace() boardTransferTrace }); ok {
				v := tr.boardTrace()
				s.Transfer = &v
			}
		}
		send()
	}
	defer func() {
		scenario = "cleanup"
		if owned {
			out.Diagnostic.ActiveStage = "close_session"
			send()
			t := time.Now()
			rc, e := dev.SendSession(0x9103)
			status := "ok"
			if e != nil || rc != PTP_RC_OK {
				status = "unconfirmed"
				out.Diagnostic.Outcome = "stopped"
			}
			step("close_session", status, rc, e, t, 0x9103, nil)
		}
		if dev != nil {
			dev.Close()
			step("release_device", "ok", 0, nil, time.Now(), 0, nil)
		}
		out.Diagnostic.ActiveStage = ""
		out.Kind = "complete"
		emit(out)
	}()
	openPhase := func(session bool) bool {
		out.Diagnostic.ActiveStage = "open_device"
		send()
		t := time.Now()
		var e error
		dev, e = open(session)
		if e != nil {
			step("open_device", "failed", 0, e, t, 0, nil)
			return false
		}
		step("open_device", "ok", 0, nil, t, 0, nil)
		return true
	}
	read := func(stage string, op uint32, args []uint32) ([]byte, uint32, error) {
		out.Diagnostic.ActiveStage = stage
		send()
		t := time.Now()
		b, rc, e := dev.SendReadPTP(op, args)
		status := "ok"
		if e != nil || rc != PTP_RC_OK {
			status = "failed"
		}
		step(stage, status, rc, e, t, op, args)
		return b, rc, e
	}
	identify := func() (*UsageResult, bool) {
		b, rc, e := read("device_info", PTP_GET_DEVICE_INFO, nil)
		if e != nil || rc != PTP_RC_OK {
			return nil, false
		}
		di, e := parsePTPDeviceInfo(b)
		if e != nil {
			step("device_info", "failed", 0, fmt.Errorf("invalid_device_info"), time.Now(), 0, nil)
			return nil, false
		}
		r := &UsageResult{Manufacturer: strings.TrimSpace(di.Manufacturer), Model: strings.TrimSpace(di.Model), Firmware: normalizeFirmware(di.DeviceVersion), Serial: strings.TrimSpace(di.Serial), Timestamp: time.Now()}
		if r.Model != "DC-S1" || firmwareKey(r.Firmware) != "2.4" {
			step("session_policy", "failed", 0, fmt.Errorf("unsupported_trial_pair"), time.Now(), 0, nil)
			return nil, false
		}
		out.Diagnostic.Model = r.Model
		out.Diagnostic.Firmware = "2.4"
		out.Diagnostic.Operations = boardDeviceOperations(b)
		return r, true
	}
	usage := func(r *UsageResult) (bool, uint32) {
		b, rc, e := read("usage", PANASONIC_GET_SETUP_INFO, []uint32{SETUP_INFO_TAG})
		if e != nil {
			return false, 0
		}
		if rc != PTP_RC_OK {
			return true, rc
		}
		c, e := parseUsageRaw(b)
		if e != nil {
			step("parse_usage", "failed", rc, fmt.Errorf("unsupported_structure"), time.Now(), 0, nil)
			return true, rc
		}
		v := *r
		v.Timestamp = time.Now()
		v.PowerWake, v.Shutter = c.PowerWake, c.Shutter
		v.Raw3, v.Raw4, v.Raw5, v.Raw6, v.Raw7 = c.Raw3, c.Raw4, c.Raw5, c.Raw6, c.Raw7
		v.Errors, v.RawLen = c.Errors, len(b)
		applyValidationStatus(&v)
		out.Valid = true
		out.Result = &v
		s := &out.Diagnostic.Steps[len(out.Diagnostic.Steps)-1]
		s.Counters = []uint32{c.PowerWake, c.Shutter, uint32(c.Raw3), uint32(c.Raw4), uint32(c.Raw5), uint32(c.Raw6), uint32(c.Raw7)}
		send()
		return true, rc
	}
	if !openPhase(false) {
		return
	}
	r, ok := identify()
	if !ok {
		return
	}
	out.Result = r
	alive, _ := usage(r)
	if !alive {
		return
	}
	dev.Close()
	step("release_device", "ok", 0, nil, time.Now(), 0, nil)
	dev = nil
	scenario = "session"
	if !openPhase(true) {
		return
	}
	r2, ok := identify()
	if !ok {
		return
	}
	if r.Serial != r2.Serial {
		step("identity", "failed", 0, fmt.Errorf("device_changed"), time.Now(), 0, nil)
		return
	}
	out.Diagnostic.ActiveStage = "open_session"
	send()
	t := time.Now()
	rc, e := dev.SendSession(0x9102)
	if e != nil || rc != PTP_RC_OK {
		step("open_session", "failed", rc, e, t, 0x9102, []uint32{0x10001})
		return
	}
	owned = true
	step("open_session", "ok", rc, nil, t, 0x9102, []uint32{0x10001})
	alive, lastRC := usage(r2)
	if !alive {
		return
	}
	scenario = "settled"
	// One explicitly labelled timing test only after GeneralError/DeviceBusy.
	// It is not a retry loop and does not claim that waiting is a known fix.
	if lastRC == 0x2002 || lastRC == 0x2019 {
		out.Diagnostic.ActiveStage = "settle"
		send()
		pause(750 * time.Millisecond)
		alive, lastRC = usage(r2)
		if !alive {
			return
		}
	} else {
		step("usage", "skipped", 0, fmt.Errorf("not_needed"), time.Now(), 0, nil)
	}
	// A final read-only device reopen tests whether initialization survives a
	// successfully closed vendor session. This is not a physical USB reset.
	scenario = "post_session"
	if lastRC != PTP_RC_OK {
		out.Diagnostic.ActiveStage = "close_session"
		send()
		t = time.Now()
		rc, e = dev.SendSession(0x9103)
		owned = false
		if e != nil || rc != PTP_RC_OK {
			step("close_session", "unconfirmed", rc, e, t, 0x9103, nil)
			return
		}
		step("close_session", "ok", rc, nil, t, 0x9103, nil)
		dev.Close()
		step("release_device", "ok", 0, nil, time.Now(), 0, nil)
		dev = nil
		if !openPhase(false) {
			return
		}
		r3, ok := identify()
		if !ok {
			return
		}
		if r.Serial != r3.Serial {
			step("identity", "failed", 0, fmt.Errorf("device_changed"), time.Now(), 0, nil)
			return
		}
		if alive, _ = usage(r3); !alive {
			return
		}
	} else {
		step("usage", "skipped", 0, fmt.Errorf("not_needed"), time.Now(), 0, nil)
	}
	out.Diagnostic.Outcome = "completed"
	return
}

func boardSuiteSummary(d boardDiagnostic) string {
	// Generated from allowlisted fields only, never private worker data.
	var b strings.Builder
	fmt.Fprintf(&b, "Run outcome: %s (completed means tests finished, not counters verified)\r\n", d.Outcome)
	fmt.Fprintf(&b, "S1 comprehensive diagnostic / S1 종합 진단\r\nVersion: %s\r\nModel: %s / %s\r\nStarted: %s\r\n\r\n", d.Version, d.Model, d.Firmware, d.Started.Format(time.RFC3339))
	for _, s := range d.Steps {
		fmt.Fprintf(&b, "[%s] %s: %s %s %s (%d ms)\r\n", s.Scenario, s.Stage, s.Status, s.PTP, s.Code, s.Milliseconds)
		if s.Transfer != nil {
			t := s.Transfer
			fmt.Fprintf(&b, "  phase=%s size_available=%t expected=%d received=%d reads=%d end=%t\r\n", t.Phase, t.TotalKnown, t.Total, t.Received, t.Reads, t.EndAttempted)
		}
	}
	b.WriteString("\r\nSequential tests; earlier steps can affect later state. No independent cold-start comparison.\r\n순차 검사이므로 앞 단계가 뒤 단계에 영향을 줄 수 있습니다.\r\n0x2002 = GeneralError; it does not establish unsupported hardware.\r\n0x2002만으로 미지원 기종이라고 판단할 수 없습니다.\r\nStopped/interrupted runs do not prove session cleanup. Reconnect USB before another run.\r\n중단된 검사는 세션 종료를 보장하지 않습니다. 다음 검사 전 USB를 다시 연결하세요.\r\n")
	return b.String()
}
