//go:build windows

package main

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestReviewCleanupPreservesPeriodIncreases(t *testing.T) {
	for _, period := range []int{0, 1, 2} {
		rows := []boardRecord{
			analysisRecord("a", "A", "2025-12-31T08:00:00Z", 10, 1),
			analysisRecord("b", "A", "2026-01-01T08:00:00Z", 10, 1),
			analysisRecord("c", "A", "2026-01-01T09:00:00Z", 20, 2),
		}
		before := boardBuildAnalysis(rows, "A", period, time.UTC)
		kept, _ := boardCompactRecords(rows, "A")
		after := boardBuildAnalysis(kept, "A", period, time.UTC)
		if before.UnallocatedShutter != after.UnallocatedShutter || before.UnallocatedPower != after.UnallocatedPower || after.Buckets["2026-01-01"].Shutter != 10 || after.Buckets["2026-01-01"].Power != 1 {
			t.Errorf("period %d: cleanup lost a confirmed increase: before=%+v after=%+v", period, before, after)
		}
	}
}

func TestReviewCleanupKeepsObservedZeroDays(t *testing.T) {
	rows := []boardRecord{
		analysisRecord("a", "A", "2026-01-01T08:00:00Z", 10, 1),
		analysisRecord("b", "A", "2026-01-02T08:00:00Z", 10, 1),
		analysisRecord("c", "A", "2026-01-02T09:00:00Z", 10, 1),
		analysisRecord("d", "A", "2026-01-02T10:00:00Z", 10, 1),
		analysisRecord("e", "A", "2026-01-03T08:00:00Z", 10, 1),
	}
	for i := range rows {
		stamp := rows[i].Time
		rows[i].Time = time.Date(stamp.Year(), stamp.Month(), stamp.Day(), stamp.Hour(), 0, 0, 0, time.Local)
	}
	kept, removed := boardCompactRecords(rows, "A")
	a := boardBuildAnalysis(kept, "A", 0, time.Local)
	if removed != 1 || a.Buckets["2026-01-02"].Pairs == 0 || a.Buckets["2026-01-02"].Shutter != 0 {
		t.Fatalf("observed zero day disappeared: removed=%d %+v", removed, a)
	}
}

func TestReviewSingleInstanceNative(t *testing.T) {
	name := fmt.Sprintf("Local\\LumerianReview_%d_%d", os.Getpid(), time.Now().UnixNano())
	h, exists, err := boardAcquireInstance(name)
	if err != nil || exists || h == 0 {
		t.Fatalf("first: %v %v %v", h, exists, err)
	}
	defer func() {
		if h != 0 {
			procCloseHandle.Call(h)
		}
	}()
	for i := 0; i < 10; i++ {
		h2, exists, err := boardAcquireInstance(name)
		if h2 != 0 {
			procCloseHandle.Call(h2)
		}
		if err != nil || !exists || h2 != 0 {
			t.Fatalf("duplicate: %v %v %v", h2, exists, err)
		}
	}
	procCloseHandle.Call(h)
	h = 0
	h3, exists, err := boardAcquireInstance(name)
	if h3 != 0 {
		defer procCloseHandle.Call(h3)
	}
	if err != nil || exists || h3 == 0 {
		t.Fatalf("reopen: %v %v %v", h3, exists, err)
	}
	if h4, _, err := boardAcquireInstance("Local\\invalid\\name"); err == nil || h4 != 0 {
		t.Fatal("mutex failure was ignored")
	}
}

func TestReviewNewCameraExportDefaults(t *testing.T) {
	oldView, oldStack := currentView, boardNavigationStack
	oldPreset, oldSerial, oldPage := exportPreset, exportSerialMode, historyPage
	defer func() {
		currentView, boardNavigationStack = oldView, oldStack
		exportPreset, exportSerialMode, historyPage = oldPreset, oldSerial, oldPage
	}()
	currentView = "main"
	exportPreset, exportSerialMode = "verify", "full"
	boardNavigationClick("board_home_export")
	if exportPreset != "public" || exportSerialMode != "masked" {
		t.Error("a newly opened export page reused an earlier full serial disclosure")
	}
	currentView = "main"
	historyPage = 3
	boardNavigationClick("board_home_errors")
	if historyPage != 0 {
		t.Error("a new camera error view starts on the previous camera's last page")
	}
}
