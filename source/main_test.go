package main

import (
	"os"
	"strings"
	"sync/atomic"
	"testing"
)

func TestVerifiedCounterLabels(t *testing.T) {
	cases := []struct {
		model, firmware, validator string
		verified                   bool
	}{
		{"DC-S1RM2", "Ver. 1.5", "@bieup_hieut", true},
		{"DC-S5M2", "Ver. 3.7", "엘가, 제비동선", true},
		{"DC-S5", "Ver. 2.9", "아름프로", true},
		{"DC-L10", "Ver. 1.2", "잠이든", true},
		{"DC-S1M2", "Ver. 1.4", "잠이든", true},
		{"DC-L10", "Ver. 1.3", "Not verified", false},
		{"DC-S1M2", "Ver. 1.5", "Not verified", false},
		{"DC-S1M2ES", "Ver. 1.4", "Not verified", false},
		{"DC-S5M2", "Ver. 3.8", "Not verified", false},
		{"DC-S5M2X", "Ver. 3.7", "Not verified", false},
	}
	for _, tc := range cases {
		t.Run(tc.model+"/"+tc.firmware, func(t *testing.T) {
			r := UsageResult{Model: tc.model, Firmware: tc.firmware, Shutter: 12793, PowerWake: 2969}
			applyValidationStatus(&r)
			if r.SemanticsVerified != tc.verified {
				t.Fatalf("SemanticsVerified = %v, want %v", r.SemanticsVerified, tc.verified)
			}
			validator, _ := validationDetails(r)
			if validator != tc.validator {
				t.Fatalf("validator = %q, want %q", validator, tc.validator)
			}
			summary := buildSummary(r, "last4")
			if tc.verified {
				if !strings.Contains(summary, "Shutter Actuations: 12,793") || !strings.Contains(summary, "Power / Wake Activations: 2,969") || strings.Contains(summary, "Raw Counter 2:") {
					t.Fatalf("verified summary has wrong labels: %s", summary)
				}
			} else if !strings.Contains(summary, "Shutter Actuations (estimated): 12,793") || !strings.Contains(summary, "Power / Wake Activations (estimated): 2,969") || !strings.Contains(summary, "shutter = Counter 2, power/wake = Counter 1") {
				t.Fatalf("unverified summary should use provisional names and retain raw mapping: %s", summary)
			}
		})
	}
}

func TestAutomaticLanguageAndProvisionalReports(t *testing.T) {
	if !isKoreanLanguageID(0x0412) || isKoreanLanguageID(0x0409) {
		t.Fatal("Windows UI language ID classification is incorrect")
	}
	defer atomic.StoreUint32(&languageMode, languageAuto)
	defer atomic.StoreUint32(&windowsKoreanUI, 0)
	r := UsageResult{Model: "DC-S5M2X", Firmware: "Ver. 3.7", Serial: "TESTSERIAL0230", Shutter: 12793, PowerWake: 2969}
	applyValidationStatus(&r)
	atomic.StoreUint32(&windowsKoreanUI, 1)
	summary := buildSummary(r, "last4")
	if !strings.Contains(summary, "셔터 작동 횟수 (추정됨): 12,793") || !strings.Contains(summary, "전원 / 깨우기 횟수 (추정됨): 2,969") {
		t.Fatalf("Korean provisional summary is incorrect: %s", summary)
	}
	status, _ := validationLabel(r)
	report := buildReport(r, status, "Not verified", "Not available")
	for _, want := range []string{"LUMIX 사용 정보 리포트", "셔터 작동 횟수", "원본 카운터 2 · 추정됨 / 미검증", "전원 / 깨우기 횟수", "원본 카운터 1 · 추정됨 / 미검증", "********0230"} {
		if !strings.Contains(report, want) {
			t.Errorf("Korean report missing %q", want)
		}
	}
	atomic.StoreUint32(&windowsKoreanUI, 0)
	report = buildReport(r, status, "Not verified", "Not available")
	if !strings.Contains(report, "Shutter Actuations       12,793  (Raw Counter 2; estimated / unverified)") || !strings.Contains(report, "Power / Wake Activations 2,969  (Raw Counter 1; estimated / unverified)") {
		t.Fatalf("English provisional report is incorrect: %s", report)
	}
}

func TestManualLanguageOverride(t *testing.T) {
	defer atomic.StoreUint32(&languageMode, languageAuto)
	defer atomic.StoreUint32(&windowsKoreanUI, 0)
	atomic.StoreUint32(&windowsKoreanUI, 1)
	atomic.StoreUint32(&languageMode, languageAuto)
	if !isKoreanUI() || languageButtonLabel() != "한국어 · 자동  ▾" {
		t.Fatal("Windows Korean default was not applied")
	}
	atomic.StoreUint32(&languageMode, languageEnglish)
	if isKoreanUI() || languageButtonLabel() != "English  ▾" {
		t.Fatal("English override was not applied")
	}
	atomic.StoreUint32(&windowsKoreanUI, 0)
	atomic.StoreUint32(&languageMode, languageKorean)
	if !isKoreanUI() || languageButtonLabel() != "한국어  ▾" {
		t.Fatal("Korean override was not applied")
	}
}

func TestLanguagePreferencePersists(t *testing.T) {
	t.Setenv("APPDATA", t.TempDir())
	defer atomic.StoreUint32(&languageMode, languageAuto)
	setLanguagePreference(languageEnglish)
	data, err := os.ReadFile(languageSettingPath())
	if err != nil || strings.TrimSpace(string(data)) != "2" {
		t.Fatalf("language preference was not saved: %q, %v", data, err)
	}
	atomic.StoreUint32(&languageMode, languageAuto)
	loadLanguagePreference()
	if atomic.LoadUint32(&languageMode) != languageEnglish {
		t.Fatal("language preference was not restored")
	}
}

func TestSerialMissingDoesNotIdentifyCamera(t *testing.T) {
	a := UsageResult{Model: "DC-S5M2", Serial: "Not available", Shutter: 12793}
	b := UsageResult{Model: "DC-S5M2", Serial: "Not available", Shutter: 200}
	if sameCamera(a, b) {
		t.Fatal("missing serials must not be treated as the same camera for delta calculations")
	}
	a.Serial, b.Serial = "TESTSERIAL0230", "TESTSERIAL0230"
	if !sameCamera(a, b) {
		t.Fatal("matching known serials should identify the same camera")
	}
	if sameCamera(a, UsageResult{Model: "DC-S5M2", Serial: "TESTSERIAL9999"}) {
		t.Fatal("different serials should not identify the same camera")
	}
}

func TestMissingSerialDisplay(t *testing.T) {
	for _, mode := range []string{"full", "last4", "masked"} {
		if got := serialForMode("Not available", mode); got != "Not available" {
			t.Fatalf("mode %q rendered missing serial as %q", mode, got)
		}
	}
	if got := serialForMode("TESTSERIAL0230", "last4"); got != "********0230" {
		t.Fatalf("known serial was not masked as expected: %q", got)
	}
}
