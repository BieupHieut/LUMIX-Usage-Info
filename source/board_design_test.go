//go:build windows

package main

import "testing"

func TestBoardDesignSecondaryControls(t *testing.T) {
	oldEnabled, oldView, oldState, oldResult, oldSaved := boardEnabled, currentView, appState, lastResult, lastSavedPath
	oldMenu, oldPNG, oldReport, oldBusy := languageMenuOpen, pngBusy, reportBusy, probeBusy
	defer func() {
		boardEnabled, currentView, appState, lastResult, lastSavedPath = oldEnabled, oldView, oldState, oldResult, oldSaved
		languageMenuOpen, pngBusy, reportBusy, probeBusy = oldMenu, oldPNG, oldReport, oldBusy
	}()
	boardEnabled = true
	appState = "connected"
	languageMenuOpen = false
	pngBusy = false
	reportBusy = false
	probeBusy = false
	lastResult = UsageResult{Model: "DC-S1RM2", Errors: make([]ErrorRecord, 8)}
	lastSavedPath = "sample-report.txt"
	for _, v := range []string{"export", "history", "details", "credits", "versions", "storage", "analytics", "card_info", "query_raw", "history_bodies"} {
		currentView = v
		actions := visibleActions()
		found := false
		for i, a := range actions {
			if boardIsBackAction(a.ID) {
				found = true
				if a.Rect != backRect {
					t.Fatalf("back position %s", v)
				}
			}
			for _, b := range actions[i+1:] {
				if a.Rect.L < b.Rect.R && b.Rect.L < a.Rect.R && a.Rect.T < b.Rect.B && b.Rect.T < a.Rect.B {
					t.Fatalf("overlap %s: %s / %s", v, a.ID, b.ID)
				}
			}
		}
		if !found {
			t.Fatalf("missing back %s", v)
		}
	}
}
