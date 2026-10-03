package main

import (
	"bytes"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
)

func TestCommunityVerificationIsExactModelFirmware(t *testing.T) {
	for _, tc := range []struct {
		model, fw string
		known     bool
	}{
		{"DC-L10", "Ver. 1.2", true}, {"DC-S1M2", "V1.4", true},
		{"DC-L10", "1.3", false}, {"DC-S1M2", "1.5", false}, {"DC-S1M2ES", "1.4", false}, {"DMC-L10", "1.2", false},
	} {
		r := UsageResult{Model: tc.model, Firmware: tc.fw}
		applyValidationStatus(&r)
		report, ok := readSuccessDetails(r)
		if ok != tc.known || r.SemanticsVerified != tc.known || r.VerifiedFirmware != tc.known {
			t.Fatalf("incorrect scope: %+v, report %+v", r, report)
		}
		if ok && (report.By != "잠이든" || report.Date != "2026-10-02") {
			t.Fatal(report)
		}
	}
}

func TestL10UsesRegularReadTransportWithStrictResponse(t *testing.T) {
	for _, length := range []int{145, 146, 147} {
		dev := &l10FakeTransport{info: l10TestPTP("DC-L10", "TEST_PRIVATE", []uint16{0x1001, 0x9414}), raw: l10TestReply(length, SETUP_INFO_REPLY), rc: PTP_RC_OK}
		r, err := probeUsageTransport(DeviceInfo{FriendlyName: "DC-L10"}, dev)
		if (err == nil) != (length == 146) {
			t.Fatalf("length %d, error %v", length, err)
		}
		if !reflect.DeepEqual(dev.ops, []uint32{0x1001, 0x9414}) || !reflect.DeepEqual(dev.args[1], []uint32{SETUP_INFO_TAG}) {
			t.Fatal("unexpected regular read operations", dev.ops, dev.args)
		}
		if err == nil && (r.Model != "DC-L10" || r.Shutter != 100 || r.PowerWake != 42 || !r.SemanticsVerified) {
			t.Fatal(r)
		}
	}
	dev := &l10FakeTransport{info: []byte{0}, rc: PTP_RC_OK}
	if _, err := probeUsageTransport(DeviceInfo{FriendlyName: "DC-L10"}, dev); err == nil || len(dev.ops) != 1 {
		t.Fatal("unconfirmed L10 reached service read")
	}
	dev = &l10FakeTransport{info: l10TestPTP("DC-S1M2", "TEST_PRIVATE", []uint16{0x1001, 0x9414}), raw: l10TestReply(146, SETUP_INFO_REPLY), rc: PTP_RC_OK}
	dev.info = bytes.Replace(dev.info, []byte{'1', 0, '.', 0, '2', 0}, []byte{'1', 0, '.', 0, '4', 0}, 1)
	r, err := probeUsageTransport(DeviceInfo{FriendlyName: "DC-S1M2"}, dev)
	if err != nil || r.Model != "DC-S1M2" || r.Firmware != "Ver. 1.4" || !r.SemanticsVerified {
		t.Fatal(r, err)
	}
}

func TestCommunityAttributionInBothLanguagesAndNoDiagnosticEntry(t *testing.T) {
	old := atomic.LoadUint32(&languageMode)
	defer atomic.StoreUint32(&languageMode, old)
	for _, mode := range []uint32{languageEnglish, languageKorean} {
		atomic.StoreUint32(&languageMode, mode)
		for _, model := range []string{"DC-L10", "DC-S1M2"} {
			fw := "1.2"
			if model == "DC-S1M2" {
				fw = "1.4"
			}
			r := UsageResult{Model: model, Firmware: fw, Serial: "PRIVATE_SERIAL1519", Shutter: 1387, PowerWake: 404}
			applyValidationStatus(&r)
			status, _ := validationLabel(r)
			report := buildReport(r, status, "Not verified", "Not available")
			for _, text := range []string{report, buildSummary(r, "last4")} {
				if !strings.Contains(text, "잠이든") || !strings.Contains(text, "cafe.naver.com/panalumix") || strings.Contains(text, "PRIVATE_SERIAL") {
					t.Fatal("missing attribution or serial exposure", text)
				}
			}
			if !strings.Contains(report, "VERIFIED USAGE") && !strings.Contains(report, "검증된 사용 횟수") {
				t.Fatal("community verification missing from report")
			}
		}
	}
	stateMu.Lock()
	oldView, oldState := currentView, appState
	currentView, appState = "validation", "waiting"
	stateMu.Unlock()
	defer func() { stateMu.Lock(); currentView, appState = oldView, oldState; stateMu.Unlock() }()
	for _, a := range visibleActions() {
		if a.ID == "l10Entry" {
			t.Fatal("separate diagnostic entry still exposed")
		}
	}
}
