//go:build windows

package main

import (
	"fmt"
	"strconv"
	"strings"
)

var boardTabs = []struct {
	view, key string
	rect      Rect
}{
	{"main", "camera", Rect{40, 688, 258, 728}}, {"usage_history", "history", Rect{267, 688, 485, 728}}, {"diagnostics", "diagnostic", Rect{494, 688, 712, 728}}, {"validation", "compat_tab", Rect{721, 688, 940, 728}},
}
var boardRects = map[string]Rect{
	"shutter_help":     {594, 270, 622, 298},
	"history_settings": {735, 195, 920, 230}, "history_analysis": {535, 195, 720, 230},
	"history_delete": {60, 631, 260, 670}, "history_clear": {270, 631, 440, 670},
	"history_clean": {450, 631, 690, 670}, "history_folder": {700, 631, 920, 670},
	"settings_back":   {40, 136, 80, 176},
	"session_option":  {60, 244, 610, 280},
	"welcome_read":    {330, 500, 650, 552},
	"welcome_history": {60, 615, 330, 656}, "welcome_diag": {350, 615, 625, 656}, "welcome_compat": {645, 615, 920, 656},
	"card_view": {277, 570, 484, 610}, "card_back": {40, 136, 80, 176},
	"serial": {78, 367, 314, 427},
	"read":   {700, 197, 920, 235}, "trial": {60, 232, 610, 273}, "extras": {60, 306, 610, 344}, "diag_read": {700, 197, 920, 235}, "suite": {650, 306, 920, 344},
	"diag_save": {353, 570, 627, 610}, "issues": {646, 570, 920, 610}, "home_diag": {40, 624, 480, 666}, "home_compat": {500, 563, 920, 610},
	"home_export": {150, 515, 360, 553}, "home_details": {385, 515, 595, 553}, "home_errors": {620, 515, 830, 553},
	"home_raw": {60, 530, 480, 566}, "diag_raw": {60, 570, 334, 610},
	"raw_prev": {60, 553, 350, 589}, "raw_next": {630, 553, 920, 589}, "raw_save": {60, 610, 480, 655}, "raw_back": {40, 136, 80, 176},
	"filter": {60, 250, 600, 282}, "hprev": {620, 250, 765, 282}, "hnext": {775, 250, 920, 282},
	"auto": {60, 254, 920, 300}, "external_auto": {60, 410, 920, 456}, "choose_file": {60, 474, 480, 520}, "update_file": {500, 474, 920, 520},
	"choose_device": {620, 342, 920, 382},
	"next_store":    {824, 250, 920, 282},
}

func boardView(view string) bool {
	return view == "analytics" || view == "card_info" || view == "main" || view == "usage_history" || view == "diagnostics" || view == "storage" || view == "query_raw" || view == "history_bodies"
}
func boardActions() []uiAction {
	stateMu.Lock()
	view, busy, pb, rb := currentView, probeBusy, pngBusy, reportBusy
	valid := boardLatest.Valid
	welcome := view == "main" && lastResult.Model == ""
	hasSerial := lastResult.Model != "" && knownSerial(lastResult.Serial)

	hasDiag := !boardLatest.Diagnostic.Started.IsZero() || !boardFailure.Diagnostic.Started.IsZero()
	s := boardData
	d := boardCurrentDiagnostic(boardLatest, boardFailure)
	devices := len(boardDevices)
	stores := 0
	if boardLatest.Camera != nil {
		stores = len(boardLatest.Camera.Stores)
	}
	stateMu.Unlock()
	a := []uiAction{{"language", languageButtonRect, !pb && !rb}, {"credits", footerCreditsRect, !pb && !rb}, {"versions", footerVersionRect, !pb && !rb}}
	if languageMenuOpen {
		return append(a, uiAction{"langAuto", languageAutoRect, true}, uiAction{"langKo", languageKoreanRect, true}, uiAction{"langEn", languageEnglishRect, true}, uiAction{"langJa", languageJapaneseRect, true}, uiAction{"langEs", languageSpanishRect, true})
	}
	{
		for _, t := range boardTabs {
			a = append(a, uiAction{"tab_" + t.view, t.rect, !pb && !rb})
		}
	}
	add := func(id string, on bool) { a = append(a, uiAction{"board_" + id, boardRects[id], on && !pb && !rb}) }
	switch view {
	case "main":
		if welcome {
			add("welcome_read", !busy)

			break
		}
		add("read", !busy)
		if hasSerial {
			add("serial", !busy)
		}
		add("shutter_help", true)

		if valid {
			add("home_export", !busy)
			add("home_details", true)
			add("home_errors", true)
		}
	case "diagnostics":

		add("diag_read", !busy)
		if busy {
			add("suite", true)
		}
		add("diag_raw", hasDiag && !busy)
		add("diag_save", hasDiag && !busy)
		add("issues", true)
		if devices > 0 {
			add("choose_device", !busy)
		}
	case "card_info":
		add("diag_raw", hasDiag && !busy)
		add("card_back", true)
		if stores > 1 {
			add("next_store", true)
		}
	case "usage_history":
		rows, bodies := boardHistoryRows(s, boardBodyIndex)
		add("history_settings", !busy)
		add("history_analysis", !busy)
		add("history_delete", !busy && boardStoreError == nil && boardSelectionCount(s) > 0)
		add("history_clear", len(boardSelectedRecords) > 0)
		for i := 0; i < 5 && boardPage*5+i < len(rows); i++ {
			a = append(a, uiAction{"board_record_" + rows[boardPage*5+i].ID, boardHistoryCheckRect(i), !busy && !pb && !rb})
		}
		add("history_clean", !busy && len(rows) > 2 && boardStoreError == nil)
		add("history_folder", true)
		add("filter", len(bodies) > 1)
		add("hprev", boardPage > 0)
		add("hnext", (boardPage+1)*5 < len(rows)) // Export lives on Storage.
	case "query_raw":
		add("raw_prev", boardRawPage > 0)
		add("raw_next", (boardRawPage+1)*8 < len(boardRawLines(d)))
		add("raw_save", hasDiag && !busy)
		add("raw_back", true)
	case "history_bodies":
		_, bodies := boardHistoryRows(s, 0)
		for i := 0; i < 6 && boardBodyPage*6+i < len(bodies); i++ {
			a = append(a, uiAction{"board_body_" + strconv.Itoa(boardBodyPage*6+i), boardBodyRect(i), !pb && !rb})
		}
		add("raw_prev", boardBodyPage > 0)
		add("raw_next", (boardBodyPage+1)*6 < len(bodies))
		add("raw_back", true)
	case "analytics":
		for _, id := range []string{"analysis_body", "analysis_day", "analysis_month", "analysis_year", "analysis_shutter", "analysis_power", "analysis_back"} {
			add(id, !busy)
		}
		add("analysis_prev", !busy && boardAnalysisCanPrevious(s))
		add("analysis_next", !busy && boardAnalysisOffset > 0)
	case "storage":
		add("settings_back", true)
		add("auto", !busy)
		add("external_auto", !busy && s.External != "")
		add("choose_file", !busy)
		add("update_file", !busy && s.External != "")
	}
	return a
}
func boardLabel(hdc uintptr, text string, r Rect, font uintptr) {
	setFont(hdc, font)
	drawText(hdc, text, r, DT_LEFT|DT_VCENTER|DT_SINGLELINE|0x8000, rgb(229, 234, 244))
}
func boardParagraph(hdc uintptr, text string, r Rect) {
	setFont(hdc, fontSmall)
	drawText(hdc, text, r, DT_LEFT|0x10, rgb(170, 186, 206))
}
func boardButton(hdc uintptr, id, key string) {
	if id == "read" || id == "diag_read" || id == "welcome_read" {
		r := boardRects[id]
		fillRound(hdc, r, actionColor(r, rgb(161, 62, 79)), 10)
		setFont(hdc, fontButton)
		drawText(hdc, boardText(key), r, DT_SINGLELINE|DT_CENTER|DT_VCENTER, actionColor(r, rgb(250, 245, 246)))
		decorateAction(hdc, r)
		return
	}
	drawOutlineButton(hdc, boardRects[id], boardText(key))
}
func boardHeader(hdc uintptr, view string) {
	if view == "query_raw" {
		view = boardRawReturn
	}
	if view == "history_bodies" || view == "storage" {
		view = "usage_history"
	}
	stateMu.Lock()
	welcome := view == "main" && lastResult.Model == ""
	stateMu.Unlock()
	if view == "card_info" {
		view = "diagnostics"
	}
	drawBitmap(hdc, logoBitmap, 40, 25)
	if !welcome {
		drawBitmap(hdc, mascotInfoBitmap, 802, 12)
	}
	boardLabel(hdc, "LUMIX Usage Board", Rect{43, 86, 260, 120}, fontSection)
	boardTextLink(hdc, "v"+appVersion, footerVersionRect, false)
	view = boardSelectedTab(view)
	for _, t := range boardTabs {
		c := rgb(36, 39, 46)
		if t.view == view {
			c = rgb(43, 67, 95)
		}
		fillRound(hdc, t.rect, actionColor(t.rect, c), 10)
		setFont(hdc, fontSmall)
		drawText(hdc, boardText(t.key), t.rect, DT_CENTER|DT_VCENTER|DT_SINGLELINE, rgb(227, 236, 248))
	}
}
func boardFooter(hdc uintptr) {
	boardLabel(hdc, boardText("readonly"), Rect{40, 737, 365, 767}, fontTiny)
	drawOutlineButton(hdc, languageButtonRect, languageButtonLabel())
	setFont(hdc, fontTiny)
	boardTextLink(hdc, "@bieup_hieut · "+boardText("about"), footerCreditsRect, true)
}
func boardPaint(hdc uintptr, view string) {
	stateMu.Lock()
	r := lastResult
	latest, failure := boardLatest, boardFailure
	s := boardData
	status, notice := boardStatus, boardNotice
	legacyNotice := uiNotice
	busy := probeBusy
	devices := append([]DeviceInfo(nil), boardDevices...)
	storeErr := boardStoreError
	stateMu.Unlock()
	if view != "main" || r.Model != "" {
		boardPanel(hdc)
	}
	switch view {
	case "main":
		boardPaintCamera(hdc, r, latest)
		if notice == "" {
			notice = boardText(status)

		}
	case "card_info":
		boardPaintCardTrial(hdc, latest.Camera)
	case "diagnostics":
		boardPageTitle(hdc, boardText("diagnostic"))
		boardLabel(hdc, boardText("connection_policy"), Rect{60, 254, 920, 278}, fontTiny)
		boardParagraph(hdc, boardText("safe"), Rect{60, 279, 920, 302})
		boardButton(hdc, "diag_read", "connect")
		if busy {
			boardButton(hdc, "suite", "stop")
		}
		if len(devices) > 0 {
			boardLabel(hdc, boardText("select")+" "+fmt.Sprintf("%d/%d", boardDeviceIndex+1, len(devices)), Rect{60, 348, 615, 369}, fontSmall)
			boardLabel(hdc, devices[boardDeviceIndex%len(devices)].FriendlyName, Rect{60, 368, 610, 390}, fontSmall)
			boardButton(hdc, "choose_device", "choose")
		} else {
			boardParagraph(hdc, boardText("diagnostic_ready"), Rect{60, 319, 920, 370})
		}
		d := boardCurrentDiagnostic(latest, failure)
		boardPaintDiagnosticSummary(hdc, d, status, busy)
		saveKey := "zip_short"

		boardButton(hdc, "diag_save", saveKey)
		boardButton(hdc, "issues", "contact_short")
		if notice == "" {
			notice = ""
		}
	case "usage_history":
		boardPaintHistory(hdc, s)
	case "analytics":
		boardPaintAnalysis(hdc, s)
	case "query_raw":
		boardPaintRaw(hdc, latest, failure)
	case "history_bodies":
		boardPaintBodies(hdc, s)
	case "storage":
		boardPageTitle(hdc, boardText("settings"))
		boardButton(hdc, "settings_back", "back")
		toggle := func(on bool) string {
			if on {
				return boardText("on")
			}
			return boardText("off")
		}
		drawOutlineButton(hdc, boardRects["auto"], boardText("auto_history")+" · "+toggle(s.Auto))
		boardParagraph(hdc, boardText("local_note"), Rect{60, 311, 920, 361})
		name := boardText("external_no_file")
		if s.External != "" {
			name = s.External
		}
		boardLabel(hdc, name, Rect{60, 370, 920, 399}, fontSmall)
		drawOutlineButton(hdc, boardRects["external_auto"], boardText("auto_external")+" · "+toggle(s.ExternalAuto))
		boardButton(hdc, "choose_file", "export_history")
		boardButton(hdc, "update_file", "update_history")
		boardParagraph(hdc, boardText("external_note"), Rect{60, 540, 920, 590})
		if storeErr != nil {
			boardParagraph(hdc, boardText("store_error"), Rect{60, 602, 920, 665})
		}
	}
	if busy && status == "" {
		status = "reading"
	}
	_ = status
	if notice == "" {
		notice = legacyNotice
	}
	boardLabel(hdc, boardText(notice), Rect{100, 140, 940, 173}, fontSmall)
	boardFooter(hdc)
	if view == "main" && r.Model != "" && (hoverAction == "board_shutter_help" || keyboardFocus && focusAction == "board_shutter_help" || boardHelpPinned) {
		boardPaintShutterHelp(hdc)
	}
}
func boardVersionHistory(hdc uintptr) {
	boardPanel(hdc)
	boardPageTitle(hdc, boardText("versions"))
	boardLabel(hdc, "v"+appVersion+" · "+releaseDate+" · "+boardText("stable_release"), Rect{60, 252, 920, 291}, fontLabel)
	keys := []string{"release_history", "release_analysis", "release_diagnostic", "release_privacy"}
	for i, k := range keys {
		boardLabel(hdc, "• "+boardText(k), Rect{75, 310 + int32(i*36), 920, 344 + int32(i*36)}, fontSmall)
	}
	boardLabel(hdc, "한국어 · English · 日本語 · Español", Rect{75, 490, 920, 525}, fontSmall)
	drawHLine(hdc, 60, 548, 920, rgb(65, 68, 75))
	boardLabel(hdc, "v1.0.0 · 2026-10-03", Rect{60, 570, 920, 607}, fontLabel)
	drawOutlineButton(hdc, backRect, "BACK")
	boardFooter(hdc)
}
func boardClick(x, y int32) bool {
	a, ok := actionAt(x, y)
	if !ok || !a.Enabled {
		return false
	}
	if boardNavigationClick(a.ID) {
		procInvalidateRect.Call(hwndMain, 0, 0)
		return true
	}
	if strings.HasPrefix(a.ID, "tab_") {
		stateMu.Lock()
		boardHelpPinned = false
		currentView = strings.TrimPrefix(a.ID, "tab_")
		stateMu.Unlock()
		procInvalidateRect.Call(hwndMain, 0, 0)
		return true
	}
	if !strings.HasPrefix(a.ID, "board_") {
		return false
	}
	id := strings.TrimPrefix(a.ID, "board_")
	if strings.HasPrefix(id, "record_") {
		key := strings.TrimPrefix(id, "record_")
		boardSelectedRecords[key] = !boardSelectedRecords[key]
		if !boardSelectedRecords[key] {
			delete(boardSelectedRecords, key)
		}
		procInvalidateRect.Call(hwndMain, 0, 0)
		return true
	}
	if strings.HasPrefix(id, "body_") {
		i, e := strconv.Atoi(strings.TrimPrefix(id, "body_"))
		stateMu.Lock()
		_, b := boardHistoryRows(boardData, 0)
		if e == nil && i >= 0 && i < len(b) {
			if boardBodyIndex != i {
				boardSelectedRecords = map[string]bool{}
				boardAnalysisOffset = 0
			}
			boardBodyIndex = i
			boardPage = 0
			boardBackLocked()
		}
		stateMu.Unlock()
		procInvalidateRect.Call(hwndMain, 0, 0)
		return true
	}
	switch id {
	case "history_delete":
		boardDeleteSelected()
	case "history_clear":
		boardSelectedRecords = map[string]bool{}
	case "analysis_day":
		boardAnalysisPeriod = 0
		boardAnalysisOffset = 0
	case "analysis_month":
		boardAnalysisPeriod = 1
		boardAnalysisOffset = 0
	case "analysis_year":
		boardAnalysisPeriod = 2
		boardAnalysisOffset = 0
	case "analysis_shutter":
		boardAnalysisPower = false
	case "analysis_power":
		boardAnalysisPower = true
	case "analysis_prev":
		boardAnalysisOffset++
	case "analysis_next":
		if boardAnalysisOffset > 0 {
			boardAnalysisOffset--
		}
	case "shutter_help":
		boardHelpPinned = !boardHelpPinned
	case "history_settings":
		currentView = "storage"
	case "settings_back":
		currentView = "usage_history"
	case "history_folder":
		boardDataFolder()
	case "history_clean":
		boardCleanHistory()
	case "session_option":
		stateMu.Lock()
		boardSession = !boardSession
		stateMu.Unlock()
	case "suite":
		stateMu.Lock()
		busy := probeBusy
		cancel := boardCancel
		stateMu.Unlock()
		if busy {
			if cancel != nil {
				cancel()
			}
		} else {
			boardStartOptions(true)
		}
	case "read", "diag_read", "welcome_read":
		boardStart()
	case "home_raw", "diag_raw":
		stateMu.Lock()
		boardRawReturn = currentView
		boardRawPage = 0
		currentView = "query_raw"
		stateMu.Unlock()
	case "raw_back":
		stateMu.Lock()
		if currentView == "history_bodies" {
			currentView = "usage_history"
		} else {
			currentView = boardRawReturn
		}
		stateMu.Unlock()
	case "raw_prev":
		stateMu.Lock()
		if currentView == "history_bodies" {
			boardBodyPage--
		} else {
			boardRawPage--
		}
		stateMu.Unlock()
	case "raw_next":
		stateMu.Lock()
		if currentView == "history_bodies" {
			boardBodyPage++
		} else {
			boardRawPage++
		}
		stateMu.Unlock()
	case "home_diag", "welcome_diag", "card_back":
		stateMu.Lock()
		currentView = "diagnostics"
		stateMu.Unlock()
	case "welcome_history":
		stateMu.Lock()
		currentView = "usage_history"
		stateMu.Unlock()
	case "welcome_compat":
		stateMu.Lock()
		currentView = "validation"
		stateMu.Unlock()
	case "card_view":
		stateMu.Lock()
		currentView = "card_info"
		stateMu.Unlock()
	case "next_store":
		boardStoragePage++
	case "serial":
		boardSerialMode = (boardSerialMode + 1) % 3
	case "trial":
		stateMu.Lock()
		boardSession = !boardSession
		stateMu.Unlock()
	case "extras":
		stateMu.Lock()
		boardExtras = !boardExtras
		stateMu.Unlock()
	case "diag_save", "raw_save":
		boardSaveDiagnostic()
	case "issues":
		boardOpenIssues()
	case "auto":
		boardToggleSetting(false)
	case "external_auto":
		boardToggleSetting(true)
	case "choose_file":
		boardSaveHistory(true)
	case "update_file":
		boardSaveHistory(false)
	case "choose_device":
		stateMu.Lock()
		boardDeviceIndex++
		stateMu.Unlock()
	case "filter":
		stateMu.Lock()
		boardBodyPage = boardBodyIndex / 6
		currentView = "history_bodies"
		stateMu.Unlock()
	case "hprev":
		boardPage--
	case "hnext":
		boardPage++
	default:
		views := map[string]string{"home_diag": "diagnostics", "home_compat": "validation", "home_export": "export", "home_details": "details", "home_errors": "history"}
		if v, ok := views[id]; ok {
			stateMu.Lock()
			currentView = v
			stateMu.Unlock()
		}
	}
	procInvalidateRect.Call(hwndMain, 0, 0)
	return true
}
