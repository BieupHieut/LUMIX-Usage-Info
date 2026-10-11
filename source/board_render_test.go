//go:build windows

package main

import (
	"fmt"
	"image"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"
)

func TestBoardRender(t *testing.T) {
	dir := os.Getenv("LUMERIAN_RENDER_DIR")
	if dir == "" {
		t.Skip("Set LUMERIAN_RENDER_DIR for native GDI screenshots")
	}
	if e := os.MkdirAll(dir, 0700); e != nil {
		t.Fatal(e)
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	oldLanguage := atomic.LoadUint32(&languageMode)
	oldScale := uiScale
	oldSerial := boardSerialMode
	defer func() { boardSerialMode = oldSerial }()
	defer func() { atomic.StoreUint32(&languageMode, oldLanguage); uiScale = oldScale }()
	boardEnabled = true
	defer func() { boardEnabled = false; probeBusy = false }()
	creatorBitmap = loadBitmap("assets/creator_character.bmp")
	defer procDeleteObject.Call(creatorBitmap.Handle)
	logoBitmap = loadBitmap("assets/lumerian_wordmark.bmp")
	mascotBitmap = loadBitmap("assets/mascot_default.bmp")
	defer procDeleteObject.Call(logoBitmap.Handle)
	defer procDeleteObject.Call(mascotBitmap.Handle)
	mascotInfoBitmap = loadBitmap("assets/mascot_info.bmp")
	defer procDeleteObject.Call(mascotInfoBitmap.Handle)
	fonts := []*uintptr{&fontCounter, &fontTitle, &fontSubtitle, &fontSection, &fontLabel, &fontValue, &fontSmall, &fontTiny, &fontMicro, &fontButton}
	sizes := []int32{-40, -28, -15, -19, -16, -18, -14, -12, -11, -15}
	for i, p := range fonts {
		weight := int32(FW_NORMAL)
		if i == 3 || i == 4 || i == 5 || i == 9 {
			weight = FW_SEMIBOLD
		}
		if i == 0 || i == 1 {
			weight = FW_BOLD
		}
		*p = createFont(sizes[i], weight)
		defer procDeleteObject.Call(*p)
	}
	lastResult = UsageResult{Model: "DC-S1RM2", Firmware: "Ver. 1.5", Serial: "TEST1519", Timestamp: time.Date(2026, 10, 6, 12, 30, 15, 0, time.Local), Shutter: 39143, PowerWake: 3996, SemanticsVerified: true}
	appState = "connected"
	boardLatest = boardReply{Valid: true, Result: &lastResult, Diagnostic: newBoardDiagnostic(false)}
	state := uint16(0)
	boardLatest.Camera = &boardCameraStatus{State: &state, Images: boardKnown32(9745), Seconds: boardKnown32(9960), CardsRead: true, Cards: []uint16{1, 2}, StorageRead: true, Stores: []boardCameraStorage{{ID: 0x10001, Read: true, Capacity: 128000000000, Free: 96000000000, Images: 9745}, {ID: 0x20001, Read: true, Capacity: 64000000000, Free: 12000000000, Images: 1450}}}
	boardStatus = "complete"
	boardData = boardStore{Auto: true, Schema: 1, Key: "0101010101010101010101010101010101010101010101010101010101010101"}
	for i := 0; i < 7; i++ {
		r := lastResult
		r.Shutter += uint32(i)
		r.Timestamp = r.Timestamp.Add(time.Duration(i) * time.Minute)
		if e := boardAppend(&boardData, r); e != nil {
			t.Fatal(e)
		}
	}
	boardLatest.Diagnostic.step("device_info", "ok", 0x2001, nil, time.Now())
	boardLatest.Diagnostic.step("usage", "ok", 0x2001, nil, time.Now())
	boardLatest.Diagnostic.Model = lastResult.Model
	boardLatest.Diagnostic.Firmware = lastResult.Firmware
	boardLatest.Diagnostic.Camera = boardLatest.Camera
	baseResult, baseLatest := lastResult, boardLatest
	baseData := boardData
	for _, mode := range []uint32{languageKorean, languageEnglish, languageJapanese, languageSpanish} {
		atomic.StoreUint32(&languageMode, mode)
		for _, view := range []string{"export_saved", "history_pages", "details", "history", "credits", "version_hover", "credits_hover", "back_focus", "main", "shutter_help", "storage_skipped", "storage_skipped_raw", "unverified", "serial_last4", "serial_full", "usage_history", "history_selected", "analytics_day", "analytics_month", "analytics_year", "analytics_power", "analytics_empty", "history_bodies", "query_raw", "raw_storage", "diag_failed", "diag_partial", "diagnostics", "storage", "export", "versions", "validation", "waiting", "failure", "language_menu", "status_partial"} {
			if filter := os.Getenv("LUMERIAN_RENDER_VIEWS"); filter != "" && !strings.Contains(","+filter+",", ","+view+",") {
				continue
			}
			lastResult, boardLatest = baseResult, baseLatest
			lastSavedPath = ""
			boardData = baseData
			boardSelectedRecords = map[string]bool{}
			boardBodyIndex = 0
			boardPage = 0
			boardAnalysisOffset = 0
			boardAnalysisPeriod = 0
			boardAnalysisPower = false
			boardStatus = "complete"
			boardNotice = ""
			uiNotice = ""
			languageMenuOpen = false
			currentView = view
			boardHelpPinned = false
			hoverAction = ""
			focusAction = ""
			keyboardFocus = false
			if view == "version_hover" {
				currentView = "main"
				hoverAction = "versions"
			}
			if view == "credits_hover" {
				currentView = "credits"
				hoverAction = "credits"
			}
			if view == "back_focus" {
				currentView = "details"
				keyboardFocus = true
				focusAction = "back"
			}
			if view == "export_saved" {
				currentView = "export"
				lastSavedPath = "sample-report.txt"
			}
			if view == "history_pages" {
				currentView = "history"
				for i := 0; i < 8; i++ {
					lastResult.Errors = append(lastResult.Errors, ErrorRecord{Index: i + 1, DateText: "2026-10-07", CodeText: "TEST", Description: "Sample error record"})
				}
			}
			if view == "history_selected" {
				currentView = "usage_history"
				boardSelectedRecords[boardData.Records[1].ID] = true
				boardSelectedRecords[boardData.Records[4].ID] = true
			}
			if strings.HasPrefix(view, "analytics_") {
				currentView = "analytics"
				boardBodyIndex = 1
				period := 0
				if view == "analytics_month" {
					period = 1
				}
				if view == "analytics_year" {
					period = 2
				}
				boardAnalysisPeriod = period
				boardAnalysisPower = view == "analytics_power"
				boardData.Records = nil
				anchor := time.Date(2026, 10, 7, 0, 0, 0, 0, time.Local)
				for i := 0; i < 8; i++ {
					if i == 3 {
						continue
					}
					start := boardPeriodAdd(boardPeriodStart(anchor, period, time.Local), period, i-7)
					r := baseResult
					r.Timestamp = start.Add(8 * time.Hour)
					r.Shutter = uint32(1000 + i*100)
					r.PowerWake = uint32(100 + i*10)
					boardAppend(&boardData, r)
					if i != 5 {
						r.Timestamp = start.Add(11 * time.Hour)
						r.Shutter += uint32(i * 10)
						r.PowerWake += uint32(i)
						boardAppend(&boardData, r)
					}
				}
				if view == "analytics_empty" {
					boardData.Records = nil
					boardBodyIndex = 0
				}
			}
			if view == "shutter_help" {
				currentView = "main"
				boardHelpPinned = true
			}
			boardSerialMode = 0
			boardRawPage, boardBodyPage = 0, 0
			if view == "raw_storage" {
				currentView = "query_raw"
				boardRawPage = 1
			}
			if view == "diag_failed" || view == "diag_partial" {
				currentView = "diagnostics"
				boardStatus = "failed"
				boardLatest.Diagnostic.FailedStage = "usage"
				if view == "diag_partial" {
					boardStatus = "partial"
					boardLatest.Diagnostic.FailedStage = "storage_info"
				}
			}
			if view == "unverified" {
				currentView = "main"
				lastResult.SemanticsVerified = false
			}
			if view == "serial_last4" || view == "serial_full" {
				currentView = "main"
				boardSerialMode = 1
				if view == "serial_full" {
					boardSerialMode = 2
				}
			}
			if view == "storage_skipped" || view == "storage_skipped_raw" {
				currentView = "card_info"
				f := &storageIDFake{statusFake: statusFake{boardFake: boardFake{model: "DC-S1RM2", usageRC: 0x2001}}, ids: []uint32{0x10000, 0x20000}}
				boardLatest = boardQueryOptions(f, false, true, func(boardReply) {})
				boardLatest.Result = &lastResult
				if view == "storage_skipped_raw" {
					currentView = "query_raw"
					boardRawPage = 0
				}
			}
			if view == "status_partial" {
				currentView = "main"
				boardLatest.Valid = false
				boardStatus = "failed"
			}
			if view == "waiting" {
				currentView = "main"
				lastResult = UsageResult{}
				boardLatest = boardReply{}
				boardStatus = ""
			}
			if view == "failure" {
				currentView = "main"
				lastResult = UsageResult{}
				boardLatest = boardReply{Diagnostic: newBoardDiagnostic(false)}
				boardLatest.Diagnostic.FailedStage = "usage"
				boardFailure = boardLatest
				boardStatus = "failed"
			}
			probeBusy = false
			if view == "suite_done" || view == "suite_stopped" || view == "suite_running" {
				currentView = "diagnostics"
				a, b := suiteCamera(), suiteCamera()
				boardLatest, _ = runSuiteFake(a, b)
				lastResult = *boardLatest.Result
				boardFailure = boardLatest
				boardStatus = view
				if view == "suite_running" {
					probeBusy = true
					boardLatest.Diagnostic.ActiveStage = "usage"
				}
			}
			if view == "language_menu" {
				currentView = "main"
				languageMenuOpen = true
			}
			for _, scale := range []float32{1, 1.5} {
				uiScale = scale
				w, h := int32(float32(canvasWidth)*scale), int32(float32(canvasHeight)*scale)
				dc, _, _ := procGetDC.Call(0)
				mem, _, _ := procCreateCompatibleDC.Call(dc)
				bmi := BITMAPINFO{BmiHeader: BITMAPINFOHEADER{BiSize: uint32(unsafe.Sizeof(BITMAPINFOHEADER{})), BiWidth: w, BiHeight: -h, BiPlanes: 1, BiBitCount: 32}}
				var bits unsafe.Pointer
				bmp, _, _ := procCreateDIBSection.Call(dc, uintptr(unsafe.Pointer(&bmi)), 0, uintptr(unsafe.Pointer(&bits)), 0, 0)
				if bmp == 0 || bits == nil {
					t.Fatal("DIB creation failed")
				}
				old, _, _ := procSelectObject.Call(mem, bmp)
				paintScene(mem, w, h)
				procSelectObject.Call(mem, old)
				raw := unsafe.Slice((*byte)(bits), int(w*h*4))
				img := image.NewRGBA(image.Rect(0, 0, int(w), int(h)))
				for i := 0; i < len(raw); i += 4 {
					img.Pix[i], img.Pix[i+1], img.Pix[i+2], img.Pix[i+3] = raw[i+2], raw[i+1], raw[i], 255
				}
				if e := writePNGFile(filepath.Join(dir, fmt.Sprintf("%d_%s_%.1f.png", mode, view, scale)), img); e != nil {
					t.Fatal(e)
				}
				procDeleteObject.Call(bmp)
				procDeleteDC.Call(mem)
				procReleaseDC.Call(0, dc)
			}
		}
	}
}
