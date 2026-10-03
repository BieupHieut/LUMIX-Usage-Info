package main

import "testing"

func TestCameraListPromotesExactProfilesWithoutDuplicateCandidates(t *testing.T) {
	old := cameraData
	defer func() { cameraData = old }()
	cameraData = mustBundledCameraData()
	initial := buildCameraList()
	if len(initial) != 20 {
		t.Fatalf("expected 20 initial rows, got %d", len(initial))
	}
	cameraData.Profiles = append(cameraData.Profiles,
		cameraProfile{Model: "DC-S9", Firmware: "2.0", By: "TEST FIXTURE"},
		cameraProfile{Model: "DC-S9", Firmware: "2.1", By: "TEST FIXTURE"})
	items := buildCameraList()
	if len(items) != 21 {
		t.Fatalf("lost firmware rows or duplicate candidate: %d", len(items))
	}
	candidates := map[string]bool{}
	s9 := 0
	for i, item := range items {
		if item.Verified != (i < len(cameraData.Profiles)) {
			t.Fatal("verified-first order broken")
		}
		if item.Model == "DC-S9" {
			s9++
			if !item.Verified {
				t.Fatal("verified model also listed as candidate")
			}
		}
		if !item.Verified {
			if candidates[item.Model] || item.Firmware != "" || item.By != "" {
				t.Fatal("invented or duplicated candidate metadata")
			}
			candidates[item.Model] = true
		}
	}
	if s9 != 2 {
		t.Fatal("exact firmware pairs not preserved")
	}
	sample := UsageResult{Model: "DC-S9", Firmware: "2.2"}
	applyValidationStatus(&sample)
	if sample.SemanticsVerified {
		t.Fatal("list promotion leaked to other firmware")
	}
}

func TestCameraListPagingMakesEveryRowReachable(t *testing.T) {
	oldPage, oldData := cameraPage, cameraData
	stateMu.Lock()
	oldView, oldState := currentView, appState
	currentView, appState = "validation", "waiting"
	stateMu.Unlock()
	defer func() {
		cameraPage, cameraData = oldPage, oldData
		stateMu.Lock()
		currentView, appState = oldView, oldState
		stateMu.Unlock()
	}()
	cameraData = mustBundledCameraData()
	items := buildCameraList()
	count := 0
	for page := 0; page < cameraListPages(); page++ {
		cameraPage = page
		start, end := page*cameraRowsPerPage, (page+1)*cameraRowsPerPage
		if end > len(items) {
			end = len(items)
		}
		count += len(items[start:end])
		found := 0
		for _, a := range visibleActions() {
			if a.ID == "cameraPrevious" {
				found++
				if a.Enabled != (page > 0) {
					t.Fatal("incorrect previous boundary")
				}
			}
			if a.ID == "cameraNext" {
				found++
				if a.Enabled != (page+1 < cameraListPages()) {
					t.Fatal("incorrect next boundary")
				}
			}
		}
		if found != 2 {
			t.Fatal("pager missing")
		}
	}
	if count != len(items) {
		t.Fatal("inaccessible camera rows")
	}
}
