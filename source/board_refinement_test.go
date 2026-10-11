//go:build windows

package main

import (
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestBoardReceivedValuesAndExportPrivacy(t *testing.T) {
	old := atomic.LoadUint32(&languageMode)
	defer atomic.StoreUint32(&languageMode, old)
	d := newBoardDiagnostic(false)
	d.Camera = &boardCameraStatus{CardsRead: true, Cards: []uint16{0}, Stores: []boardCameraStorage{{ID: 0x10001, Read: true, Capacity: 128000000000, Free: 96000000000}}}
	lines := strings.Join(boardRawLines(d), "\n")
	for _, want := range []string{"[1]  0   (0x0000)", "0x12000015", "128000000000 bytes", "96000000000 bytes"} {
		if !strings.Contains(lines, want) {
			t.Fatal("missing received field", want)
		}
	}
	d.Camera.Cards = nil
	if !strings.Contains(strings.Join(boardRawLines(d), "\n"), "0x12000015 · "+boardText("unavailable")) {
		t.Fatal("no data conflated with zero")
	}
	r := UsageResult{Model: "DC-S1RM2", Firmware: "1.5", Serial: "PRIVATE7777"}
	for _, lang := range []uint32{languageKorean, languageEnglish, languageJapanese, languageSpanish} {
		atomic.StoreUint32(&languageMode, lang)
		for _, mode := range []string{"masked", "last4", "full"} {
			report := boardExportReport(r, mode)
			if !strings.Contains(report, serialForMode(r.Serial, mode)) {
				t.Fatal("missing chosen serial representation", lang, mode)
			}
			if mode != "full" && strings.Contains(report, r.Serial) {
				t.Fatal("full serial leaked", lang, mode)
			}
			if mode == "masked" && strings.Contains(report, "7777") {
				t.Fatal("last four leaked into masked export", lang)
			}
		}
	}
}

func TestBoardRefinedNavigation(t *testing.T) {
	oldEnabled, oldView, oldState := boardEnabled, currentView, appState
	oldData, oldLatest, oldFailure := boardData, boardLatest, boardFailure
	oldBody, oldPage, oldRaw, oldReturn := boardBodyIndex, boardBodyPage, boardRawPage, boardRawReturn
	oldMenu, oldBusy, oldPNG, oldReport := languageMenuOpen, probeBusy, pngBusy, reportBusy
	defer func() {
		boardEnabled, currentView, appState = oldEnabled, oldView, oldState
		boardData, boardLatest, boardFailure = oldData, oldLatest, oldFailure
		boardBodyIndex, boardBodyPage, boardRawPage, boardRawReturn = oldBody, oldPage, oldRaw, oldReturn
		languageMenuOpen, probeBusy, pngBusy, reportBusy = oldMenu, oldBusy, oldPNG, oldReport
	}()
	boardEnabled = true
	appState = "connected"
	languageMenuOpen, probeBusy, pngBusy, reportBusy = false, false, false, false
	boardData = boardStore{Records: []boardRecord{{Body: "A", Model: "DC-S5", SerialSuffix: "1234", Time: time.Now()}, {Body: "B", Model: "DC-S5", SerialSuffix: "1234", Time: time.Now()}}}
	boardLatest = boardReply{Valid: true, Diagnostic: newBoardDiagnostic(false)}
	boardBodyIndex, boardBodyPage, boardRawPage = 0, 0, 0
	for _, view := range []string{"main", "diagnostics", "query_raw", "history_bodies", "export"} {
		currentView = view
		a := visibleActions()
		for i, x := range a {
			for _, y := range a[i+1:] {
				if x.Rect.L < y.Rect.R && y.Rect.L < x.Rect.R && x.Rect.T < y.Rect.B && y.Rect.T < x.Rect.B {
					t.Fatalf("overlap %s: %s / %s", view, x.ID, y.ID)
				}
			}
		}
	}
	currentView = "usage_history"
	if !boardClick(100, 260) || currentView != "history_bodies" {
		t.Fatal("picker not opened")
	}
	if boardBodyLabel(boardData, 1) == boardBodyLabel(boardData, 2) {
		t.Fatal("same model/suffix bodies indistinguishable")
	}
	r := boardBodyRect(2)
	if !boardClick(r.L+10, r.T+10) || currentView != "usage_history" || boardBodyIndex != 2 {
		t.Fatal("direct body selection failed")
	}
	rows, _ := boardHistoryRows(boardData, boardBodyIndex)
	if len(rows) != 1 || rows[0].Body != "B" {
		t.Fatal("wrong body selected")
	}
	currentView = "diagnostics"
	boardLatest.Valid = false
	if !boardClick(boardRects["diag_raw"].L+10, boardRects["diag_raw"].T+10) || currentView != "query_raw" {
		t.Fatal("failed query raw view unavailable")
	}
	if !boardClick(backRect.L+15, backRect.T+15) || currentView != "diagnostics" {
		t.Fatal("wrong return destination")
	}
}
