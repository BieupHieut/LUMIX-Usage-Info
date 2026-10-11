//go:build windows

package main

import "unsafe"

var boardResponseParamsKey = PROPERTYKEY{WPD_MTP_GUID, 1004}
var boardPropVariantClear = ole32.NewProc("PropVariantClear")

// Optional numeric response parameters, capped at the PTP maximum of five.
// An absent property is normal; never turn telemetry failure into query failure.
func boardResponseParameters(v unsafe.Pointer) []uint32 {
	var coll unsafe.Pointer
	hr := comCall(v, 34, uintptr(unsafe.Pointer(&boardResponseParamsKey)), uintptr(unsafe.Pointer(&coll)))
	if hrFailed(hr) || coll == nil {
		release(coll)
		return nil
	}
	defer release(coll)
	var n uint32
	if hrFailed(comCall(coll, 3, uintptr(unsafe.Pointer(&n)))) || n > 5 {
		return nil
	}
	values := make([]uint32, 0, n)
	for i := uint32(0); i < n; i++ {
		var pv PROPVARIANT
		hr = comCall(coll, 4, uintptr(i), uintptr(unsafe.Pointer(&pv)))
		ok := !hrFailed(hr) && pv.Vt == VT_UI4
		value := uint32(pv.Value)
		boardPropVariantClear.Call(uintptr(unsafe.Pointer(&pv)))
		if !ok {
			return nil
		}
		values = append(values, value)
	}
	return values
}

func boardSanitizeStep(s boardStep) boardStep {
	if !boardSuiteScenarios[s.Scenario] {
		s.Scenario = ""
	}
	allowed := false
	switch s.Opcode {
	case 0x1001, 0x1004, 0x9103:
		allowed = len(s.Parameters) == 0
	case 0x1005:
		allowed = len(s.Parameters) == 1 && s.Parameters[0] != 0 && s.Parameters[0] != 0xffffffff
	case 0x9102:
		allowed = len(s.Parameters) == 1 && s.Parameters[0] == 0x10001
	case 0x9414:
		allowed = len(s.Parameters) == 1 && (s.Parameters[0] == 0x15c00010 || s.Parameters[0] == 0x12000015)
	}
	if !allowed {
		s.Opcode = 0
		s.Parameters = nil
	}
	if s.Stage != "usage" || s.Status != "ok" || len(s.Counters) != 7 {
		s.Counters = nil
	}
	if s.Transfer != nil {
		tr := *s.Transfer
		phases := map[string]bool{"request_build": true, "begin_transfer": true, "transfer_context": true, "transfer_size": true, "read_data": true, "end_transfer": true, "complete": true, "session_request": true}
		if !phases[tr.Phase] {
			tr.Phase = "unknown"
		}
		if !boardResponsePattern.MatchString(tr.EndCode) && !boardSafeCode[tr.EndCode] {
			tr.EndCode = ""
		}
		if (s.Opcode != 0x9102 && s.Opcode != 0x9103) || len(tr.ResponseParams) > 5 {
			tr.ResponseParams = nil
		}
		s.Transfer = &tr
	}
	return s
}
