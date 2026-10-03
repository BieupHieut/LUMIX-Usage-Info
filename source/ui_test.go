//go:build windows

package main

import (
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
)

func TestViewportCoordinatesAndLimits(t *testing.T) {
	oldScale, x, y := uiScale, scrollX, scrollY
	defer func() { uiScale, scrollX, scrollY = oldScale, x, y }()
	for _, scale := range []float32{1, 1.25, 1.5, 2} {
		uiScale = scale
		scrollX, scrollY = 100, 180
		px := int32(650*scale) - scrollX
		py := int32(550*scale) - scrollY
		logicalX, logicalY := screenToCanvas(px, py)
		// Fractional scale factors quantize physical pixels; allow one logical pixel.
		if logicalX < 649 || logicalX > 650 || logicalY < 549 || logicalY > 550 {
			t.Fatalf("scale %g: hit testing gave %d,%d", scale, logicalX, logicalY)
		}
		content := int32(float32(canvasHeight) * scale)
		if got := clampScroll(99999, content, 480); got != content-480 {
			t.Fatalf("scale %g: last content cannot be reached: %d", scale, got)
		}
	}
	if clampScroll(-100, 780, 480) != 0 || clampScroll(100, 780, 900) != 0 {
		t.Fatal("invalid negative/oversize scroll offset")
	}
}
func TestBusyAndDisconnectedActions(t *testing.T) {
	stateMu.Lock()
	oldState, oldView, oldResult, oldPNG, oldReport, oldProbe := appState, currentView, lastResult, pngBusy, reportBusy, probeBusy
	appState, currentView, lastResult = "connected", "export", UsageResult{Model: "DC-S1RM2"}
	pngBusy, reportBusy, probeBusy = true, false, false
	stateMu.Unlock()
	oldMenu := languageMenuOpen
	languageMenuOpen = false
	defer func() {
		stateMu.Lock()
		appState, currentView, lastResult, pngBusy, reportBusy, probeBusy = oldState, oldView, oldResult, oldPNG, oldReport, oldProbe
		stateMu.Unlock()
		languageMenuOpen = oldMenu
	}()
	for _, a := range visibleActions() {
		if a.Enabled {
			t.Errorf("PNG in progress must disable %s", a.ID)
		}
	}
	if allowActivation(400, 580) {
		t.Fatal("busy save can be invoked")
	}
	stateMu.Lock()
	pngBusy = false
	currentView = "main"
	appState = "disconnected"
	stateMu.Unlock()
	found := map[string]bool{}
	for _, a := range visibleActions() {
		found[a.ID] = a.Enabled
	}
	if found["export"] || found["report"] || !found["refresh"] || !found["details"] {
		t.Fatalf("unexpected disconnected actions: %v", found)
	}
	stateMu.Lock()
	appState = "connected"
	probeBusy = true
	stateMu.Unlock()
	for _, a := range visibleActions() {
		if (a.ID == "refresh" || a.ID == "export" || a.ID == "report") && a.Enabled {
			t.Errorf("refreshing still enables %s", a.ID)
		}
	}
}
func TestKeyboardSkipsDisabledAndWraps(t *testing.T) {
	actions := []uiAction{{ID: "first", Enabled: true}, {ID: "busy", Enabled: false}, {ID: "last", Enabled: true}}
	for _, c := range []struct {
		current string
		back    bool
		want    string
	}{{"", false, "first"}, {"first", false, "last"}, {"last", false, "first"}, {"first", true, "last"}, {"hidden", true, "last"}} {
		if got := nextFocus(actions, c.current, c.back).ID; got != c.want {
			t.Fatalf("focus %q backwards %v: got %q want %q", c.current, c.back, got, c.want)
		}
	}
	if a := nextFocus([]uiAction{{ID: "busy"}}, "", false); a.ID != "" {
		t.Fatal("disabled action received focus")
	}
}
func TestLanguagePreferenceFailureIsReported(t *testing.T) {
	old := atomic.LoadUint32(&languageMode)
	defer atomic.StoreUint32(&languageMode, old)
	blocked := filepath.Join(t.TempDir(), "not-a-directory")
	if err := os.WriteFile(blocked, []byte("block"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("APPDATA", blocked)
	if err := setLanguagePreference(languageKorean); err == nil {
		t.Fatal("language settings failure silently ignored")
	}
	if !isKoreanUI() {
		t.Fatal("manual selection should still work for this session")
	}
	t.Setenv("APPDATA", t.TempDir())
	if err := setLanguagePreference(languageEnglish); err != nil {
		t.Fatal(err)
	}
	loadLanguagePreference()
	if isKoreanUI() {
		t.Fatal("saved English preference not restored")
	}
}
