//go:build windows

package main

import "testing"

func TestBoardSerialPrivacyCycle(t *testing.T) {
	oldEnabled, oldView, oldResult := boardEnabled, currentView, lastResult
	oldMode, oldExport, oldMenu := boardSerialMode, exportSerialMode, languageMenuOpen
	oldBusy, oldPNG, oldReport := probeBusy, pngBusy, reportBusy
	defer func() {
		boardEnabled, currentView, lastResult = oldEnabled, oldView, oldResult
		boardSerialMode, exportSerialMode, languageMenuOpen = oldMode, oldExport, oldMenu
		probeBusy, pngBusy, reportBusy = oldBusy, oldPNG, oldReport
	}()
	boardEnabled, currentView, languageMenuOpen = true, "main", false
	probeBusy, pngBusy, reportBusy = false, false, false
	lastResult = UsageResult{Model: "DC-S1RM2", Serial: "PRIVATE1519"}
	boardSerialMode, exportSerialMode = 0, "masked"
	for _, want := range []string{"************", "********1519", "PRIVATE1519", "************"} {
		if got := boardScreenSerial(lastResult.Serial); got != want {
			t.Fatalf("serial screen: got %q, want %q", got, want)
		}
		if !boardClick(120, 420) {
			t.Fatal("serial area not clickable")
		}
		if exportSerialMode != "masked" {
			t.Fatal("screen visibility changed export privacy")
		}
	}
	lastResult.Serial = ""
	if boardClick(120, 420) {
		t.Fatal("unavailable serial should not be clickable")
	}
}
