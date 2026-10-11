//go:build windows

package main

import (
	"reflect"
	"testing"
)

func TestBoardNavigationActualBackRoutes(t *testing.T) {
	oldEnabled, oldView, oldState, oldResult := boardEnabled, currentView, appState, lastResult
	oldStack := boardNavigationStack
	oldMenu, oldBusy, oldPNG, oldReport := languageMenuOpen, probeBusy, pngBusy, reportBusy
	oldLatest := boardLatest
	defer func() {
		boardEnabled, currentView, appState, lastResult = oldEnabled, oldView, oldState, oldResult
		boardNavigationStack = oldStack
		languageMenuOpen, probeBusy, pngBusy, reportBusy = oldMenu, oldBusy, oldPNG, oldReport
		boardLatest = oldLatest
	}()
	boardEnabled = true
	languageMenuOpen = false
	probeBusy = false
	pngBusy = false
	reportBusy = false
	boardLatest = boardReply{Valid: true, Diagnostic: newBoardDiagnostic(false)}
	click := func(id string) {
		t.Helper()
		for _, a := range visibleActions() {
			if a.ID == id && a.Enabled {
				handleClick(0, (a.Rect.L+a.Rect.R)/2, (a.Rect.T+a.Rect.B)/2)
				return
			}
		}
		t.Fatal("action unavailable", currentView, id)
	}
	for _, connected := range []bool{false, true} {
		lastResult = UsageResult{}
		appState = "idle"
		if connected {
			lastResult = UsageResult{Model: "DC-S1RM2", Firmware: "1.5", Serial: "TEST1234"}
			appState = "connected"
		}
		for _, from := range []string{"main", "usage_history", "storage", "diagnostics", "card_info", "query_raw", "validation"} {
			currentView = from
			boardNavigationStack = nil
			click("credits")
			if currentView != "credits" {
				t.Fatal("about not opened")
			}
			click("credits") // Reopening the current page must not create a loop.
			click("versions")
			click("back")
			if currentView != "credits" {
				t.Fatal("nested back", currentView)
			}
			click("back")
			if currentView != from || len(boardNavigationStack) != 0 {
				t.Fatal("wrong return", from, currentView, boardNavigationStack)
			}
		}
	}
	appState = "connected"
	currentView = "diagnostics"
	boardNavigationStack = nil
	click("board_diag_raw")
	click("credits")
	click("back")
	if currentView != "query_raw" {
		t.Fatal("lost raw view")
	}
	click("board_raw_back")
	if currentView != "diagnostics" {
		t.Fatal("lost diagnostic parent")
	}
	currentView = "usage_history"
	boardNavigationStack = nil
	click("board_history_settings")
	click("credits")
	click("back")
	click("board_settings_back")
	if currentView != "usage_history" {
		t.Fatal("settings parent")
	}
	currentView = "main"
	boardNavigationStack = nil
	for _, id := range []string{"board_home_export", "board_home_details", "board_home_errors"} {
		click(id)
		click("credits")
		click("back")
		click("back")
		if currentView != "main" {
			t.Fatal("camera child return", id, currentView)
		}
	}
	click("credits")
	click("tab_validation")
	if currentView != "validation" || len(boardNavigationStack) != 0 {
		t.Fatal("tab did not reset ancestry")
	}
	for _, a := range visibleActions() {
		if a.ID == "back" || a.ID == "creditsPage" || a.ID == "versionsPage" {
			t.Fatal("duplicate validation navigation", a.ID)
		}
	}
}

func TestBoardNavigationFixedTabs(t *testing.T) {
	oldEnabled, oldView, oldResult, oldState := boardEnabled, currentView, lastResult, appState
	oldMenu := languageMenuOpen
	defer func() {
		boardEnabled, currentView, lastResult, appState = oldEnabled, oldView, oldResult, oldState
		languageMenuOpen = oldMenu
	}()
	boardEnabled = true
	languageMenuOpen = false
	appState = "connected"
	expected := map[string]Rect{}
	for _, tab := range boardTabs {
		expected["tab_"+tab.view] = tab.rect
	}
	for _, model := range []string{"", "DC-S1RM2"} {
		lastResult = UsageResult{Model: model}
		for _, view := range []string{"main", "usage_history", "diagnostics", "validation", "storage", "credits", "versions"} {
			currentView = view
			got := map[string]Rect{}
			for _, a := range visibleActions() {
				if _, ok := expected[a.ID]; ok {
					got[a.ID] = a.Rect
				}
				if a.ID == "board_welcome_history" || a.ID == "board_home_compat" {
					t.Fatal("duplicate lower navigation")
				}
			}
			if !reflect.DeepEqual(got, expected) {
				t.Fatalf("tabs changed: %s / %s: %v", model, view, got)
			}
		}
	}
}
