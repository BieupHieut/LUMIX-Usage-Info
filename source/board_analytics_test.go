//go:build windows

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func analysisRecord(id, body, date string, sh, p uint32) boardRecord {
	t, _ := time.Parse(time.RFC3339, date)
	return boardRecord{ID: id, Body: body, Model: "DC-S1RM2", Firmware: "1.5", Verified: true, Time: t, Shutter: sh, Power: p}
}
func TestBoardAnalyticsPeriodsAndUnknown(t *testing.T) {
	rows := []boardRecord{
		analysisRecord("b0", "B", "2026-10-01T08:00:00Z", 800, 100),
		analysisRecord("a1", "A", "2026-10-01T10:00:00Z", 110, 21),
		analysisRecord("a0", "A", "2026-10-01T08:00:00Z", 100, 20),
		analysisRecord("a2", "A", "2026-10-02T08:00:00Z", 130, 22),
		analysisRecord("a3", "A", "2026-10-02T10:00:00Z", 130, 22),
		analysisRecord("a4", "A", "2026-11-01T08:00:00Z", 200, 25),
		analysisRecord("a5", "A", "2027-01-01T08:00:00Z", 250, 28),
		analysisRecord("b1", "B", "2026-10-01T10:00:00Z", 900, 200),
	}
	day := boardBuildAnalysis(rows, "A", 0, time.UTC)
	b := day.Buckets["2026-10-01"]
	if b.Shutter != 10 || b.Power != 1 || b.Pairs != 1 {
		t.Fatal("same-day values", b)
	}
	zero := day.Buckets["2026-10-02"]
	if zero.Pairs != 1 || zero.Shutter != 0 {
		t.Fatal("known zero lost")
	}
	unknown := day.Buckets["2026-11-01"]
	if unknown.Pairs != 0 || unknown.Reads != 1 {
		t.Fatal("single read fabricated increase")
	}
	if day.UnallocatedShutter != 140 || day.UnallocatedPower != 7 {
		t.Fatal("cross-period allocation", day)
	}
	month := boardBuildAnalysis(rows, "A", 1, time.UTC)
	if month.Buckets["2026-10-01"].Shutter != 30 || month.UnallocatedShutter != 120 {
		t.Fatal("monthly", month)
	}
	year := boardBuildAnalysis(rows, "A", 2, time.UTC)
	if year.Buckets["2026-01-01"].Shutter != 100 || year.UnallocatedShutter != 50 {
		t.Fatal("yearly", year)
	}
	if len(boardBuildAnalysis(rows, "", 0, time.UTC).Buckets) != 0 {
		t.Fatal("mixed bodies")
	}
	win := boardAnalysisWindow(day, 0, 0)
	if len(win) != 8 || win[0].Pairs != 0 || win[7].Start.Year() != 2027 {
		t.Fatal("window gaps")
	}
	if rows[1].ID != "a1" {
		t.Fatal("input sorted in place")
	}
}
func TestBoardAnalyticsBoundariesAndTimezone(t *testing.T) {
	rows := []boardRecord{analysisRecord("0", "A", "2026-10-01T14:00:00Z", 100, 20), analysisRecord("1", "A", "2026-10-01T16:00:00Z", 110, 21)}
	kst := time.FixedZone("KST", 9*3600)
	if len(boardBuildAnalysis(rows, "A", 0, kst).Buckets) != 2 {
		t.Fatal("local midnight ignored")
	}
	for _, kind := range []string{"reset", "firmware", "verification", "deletion", "same_time"} {
		r := append([]boardRecord(nil), rows...)
		switch kind {
		case "reset":
			r[1].Shutter = 1
		case "firmware":
			r[1].Firmware = "2.0"
		case "verification":
			r[1].Verified = false
		case "deletion":
			r[1].BreakBefore = true
		case "same_time":
			r[1].Time = r[0].Time
		}
		a := boardBuildAnalysis(r, "A", 1, kst)
		if a.Skipped != 1 || a.Buckets["2026-10-01"].Pairs != 0 {
			t.Fatal("boundary joined", kind, a)
		}
	}
	dec := time.Date(2026, 12, 1, 0, 0, 0, 0, kst)
	if boardPeriodAdd(dec, 1, 1).Month() != time.January {
		t.Fatal("month wrap")
	}
	leap := time.Date(2024, 2, 28, 0, 0, 0, 0, kst)
	if boardPeriodAdd(leap, 0, 2).Day() != 1 {
		t.Fatal("leap day")
	}
}
func TestBoardSelectionDeletionDoesNotJoinIntervals(t *testing.T) {
	rows := []boardRecord{analysisRecord("a", "A", "2026-10-01T08:00:00Z", 100, 5), analysisRecord("b", "B", "2026-10-01T08:00:00Z", 100, 5), analysisRecord("c", "A", "2026-10-01T09:00:00Z", 120, 6), analysisRecord("d", "A", "2026-10-01T10:00:00Z", 130, 7), analysisRecord("e", "A", "2026-10-01T11:00:00Z", 140, 8)}
	out, n := boardRemoveSelected(rows, map[string]bool{"c": true, "missing": true})
	if n != 1 || len(out) != 4 || !out[2].BreakBefore || out[1].BreakBefore || rows[3].BreakBefore {
		t.Fatal("wrong deletion/gap", out)
	}
	a := boardBuildAnalysis(out, "A", 0, time.UTC)
	if a.Buckets["2026-10-01"].Shutter != 10 || a.Skipped != 1 {
		t.Fatal("deleted interval joined", a)
	}
	all := map[string]bool{}
	for _, r := range rows {
		all[r.ID] = true
	}
	out, n = boardRemoveSelected(rows, all)
	if n != 5 || len(out) != 0 {
		t.Fatal("delete all failed")
	}
}
func TestBoardSelectionConfirmAndBackup(t *testing.T) {
	oldData, oldSelection, oldPath, oldConfirm, oldBusy, oldError := boardData, boardSelectedRecords, boardDeletionPath, boardConfirmDelete, probeBusy, boardStoreError
	oldBody, oldPage, oldOffset := boardBodyIndex, boardPage, boardAnalysisOffset
	defer func() {
		boardData, boardSelectedRecords, boardDeletionPath, boardConfirmDelete, probeBusy, boardStoreError = oldData, oldSelection, oldPath, oldConfirm, oldBusy, oldError
		boardBodyIndex, boardPage, boardAnalysisOffset = oldBody, oldPage, oldOffset
	}()
	dir := t.TempDir()
	path := filepath.Join(dir, "usage-board.json")
	boardDeletionPath = func() string { return path }
	probeBusy = false
	boardStoreError = nil
	boardData = boardStore{Schema: 1, Key: strings.Repeat("01", 32), ExternalAuto: true, Records: []boardRecord{analysisRecord("a", "A", "2026-10-01T08:00:00Z", 10, 1), analysisRecord("b", "A", "2026-10-01T09:00:00Z", 20, 2)}}
	boardPersist(path, boardData)
	before, _ := os.ReadFile(path)
	boardSelectedRecords = map[string]bool{"a": true}
	boardConfirmDelete = func(string) bool { return false }
	boardDeleteSelected()
	got, _ := os.ReadFile(path)
	if string(got) != string(before) || len(boardData.Records) != 2 {
		t.Fatal("cancel modified data")
	}
	boardConfirmDelete = func(string) bool { return true }
	boardDeleteSelected()
	if len(boardData.Records) != 1 || !boardData.Records[0].BreakBefore || boardData.ExternalAuto || len(boardSelectedRecords) != 0 {
		t.Fatal("commit wrong")
	}
	backups, _ := filepath.Glob(filepath.Join(dir, "history-backups", "*.json"))
	if len(backups) != 1 {
		t.Fatal("backup missing")
	}
	backup, _ := os.ReadFile(backups[0])
	if string(backup) != string(before) {
		t.Fatal("backup mismatch")
	}
	got, _ = os.ReadFile(path)
	var persisted boardStore
	json.Unmarshal(got, &persisted)
	if persisted.Schema != 2 || !persisted.Records[0].BreakBefore {
		t.Fatal("gap not persisted")
	}
	loaded, e := boardLoad(path)
	if e != nil || loaded.Schema != 2 || !loaded.Records[0].BreakBefore {
		t.Fatal("cannot reopen new history", e)
	}
	old, e := boardLoad(backups[0])
	if e != nil || old.Schema != 1 || len(old.Records) != 2 {
		t.Fatal("cannot restore old history", e)
	}
	exported, e := boardExportBytes(loaded)
	if e != nil || !strings.Contains(string(exported), `"schema": 2`) || !strings.Contains(string(exported), `"break_before": true`) || strings.Contains(string(exported), "local_key") {
		t.Fatal("export version/gap/privacy", e)
	}
}

func TestBoardSelectionUIAndAnalyticsNavigation(t *testing.T) {
	oldState, oldResult := appState, lastResult
	defer func() { appState, lastResult = oldState, oldResult }()
	appState = "connected"
	lastResult = UsageResult{Model: "DC-S1RM2"}
	oldData, oldSelected, oldView, oldEnabled, oldBusy, oldMenu := boardData, boardSelectedRecords, currentView, boardEnabled, probeBusy, languageMenuOpen
	oldBody, oldPage, oldOffset, oldPeriod, oldPower := boardBodyIndex, boardPage, boardAnalysisOffset, boardAnalysisPeriod, boardAnalysisPower
	oldStack, oldPNG, oldReport, oldError := boardNavigationStack, pngBusy, reportBusy, boardStoreError
	defer func() {
		boardData, boardSelectedRecords, currentView, boardEnabled, probeBusy, languageMenuOpen = oldData, oldSelected, oldView, oldEnabled, oldBusy, oldMenu
		boardBodyIndex, boardPage, boardAnalysisOffset, boardAnalysisPeriod, boardAnalysisPower = oldBody, oldPage, oldOffset, oldPeriod, oldPower
		boardNavigationStack, pngBusy, reportBusy, boardStoreError = oldStack, oldPNG, oldReport, oldError
	}()
	boardData = boardStore{Schema: 1}
	boardSelectedRecords = map[string]bool{}
	currentView = "usage_history"
	boardEnabled = true
	probeBusy = false
	languageMenuOpen = false
	pngBusy = false
	reportBusy = false
	boardStoreError = nil
	boardBodyIndex = 0
	boardPage = 0
	for i := 0; i < 12; i++ {
		r := analysisRecord(fmt.Sprintf("a%d", i), "A", "2026-10-01T08:00:00Z", uint32(100+i), uint32(20+i))
		r.Time = r.Time.AddDate(0, 0, i)
		boardData.Records = append(boardData.Records, r)
	}
	boardData.Records = append(boardData.Records, analysisRecord("b", "B", "2026-10-12T08:00:00Z", 10, 1))
	click := func(id string) {
		t.Helper()
		for _, a := range visibleActions() {
			if a.ID == id && a.Enabled {
				handleClick(0, (a.Rect.L+a.Rect.R)/2, (a.Rect.T+a.Rect.B)/2)
				return
			}
		}
		t.Fatal("unavailable", id, currentView)
	}
	click("board_record_b")
	click("board_hnext")
	click("board_record_a7")
	if boardSelectionCount(boardData) != 2 {
		t.Fatal("selection lost between pages")
	}
	click("board_history_analysis")
	if currentView != "analytics" || boardBodyIndex != 1 || boardSelectionCount(boardData) != 0 {
		t.Fatal("analysis/filter transition")
	}
	click("board_analysis_prev")
	if boardAnalysisOffset != 1 {
		t.Fatal("chart paging")
	}
	click("board_analysis_month")
	if boardAnalysisPeriod != 1 || boardAnalysisOffset != 0 {
		t.Fatal("period transition")
	}
	click("board_analysis_power")
	if !boardAnalysisPower {
		t.Fatal("metric selection")
	}
	click("board_analysis_body")
	click("board_body_2")
	if currentView != "analytics" || boardBodyIndex != 2 {
		t.Fatal("chart picker return")
	}
	click("board_analysis_back")
	if currentView != "usage_history" {
		t.Fatal("chart return")
	}
	for _, view := range []string{"usage_history", "analytics", "storage", "diagnostics", "export", "credits", "versions", "validation"} {
		currentView = view
		actions := visibleActions()
		for i, a := range actions {
			for _, b := range actions[i+1:] {
				if a.Rect.L < b.Rect.R && b.Rect.L < a.Rect.R && a.Rect.T < b.Rect.B && b.Rect.T < a.Rect.B {
					t.Fatalf("overlap %s: %s / %s", view, a.ID, b.ID)
				}
			}
		}
	}
}
