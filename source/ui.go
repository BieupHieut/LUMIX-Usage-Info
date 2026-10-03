//go:build windows

package main

import (
	"os"
	"path/filepath"
	"syscall"
	"unsafe"
)

const windowStyle = 0x00CE0000 // Caption, resize, system menu, minimize; no maximize box.
const canvasWidth, canvasHeight int32 = 980, 780

var (
	procSaveDC                                          = gdi32.NewProc("SaveDC")
	procRestoreDC                                       = gdi32.NewProc("RestoreDC")
	procSetGraphicsMode                                 = gdi32.NewProc("SetGraphicsMode")
	procSetWorldTransform                               = gdi32.NewProc("SetWorldTransform")
	procCreatePen                                       = gdi32.NewProc("CreatePen")
	procStretchBlt                                      = gdi32.NewProc("StretchBlt")
	procSetStretchBltMode                               = gdi32.NewProc("SetStretchBltMode")
	procSetBrushOrgEx                                   = gdi32.NewProc("SetBrushOrgEx")
	procGetCursorPos                                    = user32.NewProc("GetCursorPos")
	procScreenToClient                                  = user32.NewProc("ScreenToClient")
	procSetScrollInfo                                   = user32.NewProc("SetScrollInfo")
	procGetScrollInfo                                   = user32.NewProc("GetScrollInfo")
	procSetWindowPos                                    = user32.NewProc("SetWindowPos")
	procGetDpiForWindow                                 = user32.NewProc("GetDpiForWindow")
	procSetProcessDpiAwarenessContext                   = user32.NewProc("SetProcessDpiAwarenessContext")
	procGetMonitorInfo                                  = user32.NewProc("GetMonitorInfoW")
	procMonitorFromWindow                               = user32.NewProc("MonitorFromWindow")
	procAdjustWindowRectExForDpi                        = user32.NewProc("AdjustWindowRectExForDpi")
	procAdjustWindowRectEx                              = user32.NewProc("AdjustWindowRectEx")
	procGetKeyState                                     = user32.NewProc("GetKeyState")
	procSetCapture                                      = user32.NewProc("SetCapture")
	procReleaseCapture                                  = user32.NewProc("ReleaseCapture")
	procTrackMouseEvent                                 = user32.NewProc("TrackMouseEvent")
	procShellExecute                                    = syscall.NewLazyDLL("shell32.dll").NewProc("ShellExecuteW")
	uiScale                                     float32 = 1
	scrollX, scrollY, clientWidth, clientHeight int32
	viewportUpdating                            bool
	hoverAction, pressedAction, focusAction     string
	keyboardFocus                               bool
	languageSaveError                           error
	lastSavedPath                               string // Guarded by stateMu, like export state.
	openFileRect                                = Rect{290, 690, 480, 719}
	openFolderRect                              = Rect{500, 690, 690, 719}
	exportOpenFileRect                          = Rect{75, 552, 285, 580}
	exportOpenFolderRect                        = Rect{75, 589, 285, 617}
)

type scrollInfo struct {
	Size, Mask    uint32
	Min, Max      int32
	Page          uint32
	Pos, TrackPos int32
}
type worldTransform struct{ M11, M12, M21, M22, DX, DY float32 }
type monitorInfo struct {
	Size          uint32
	Monitor, Work WinRect
	Flags         uint32
}
type trackMouse struct {
	Size, Flags uint32
	Hwnd        uintptr
	HoverTime   uint32
}
type uiAction struct {
	ID      string
	Rect    Rect
	Enabled bool
}

func enableDPIAwareness() {
	if procSetProcessDpiAwarenessContext.Find() == nil {
		if ok, _, _ := procSetProcessDpiAwarenessContext.Call(^uintptr(3)); ok != 0 {
			return
		}
	}
	procSetProcessDPIAware.Call()
}
func windowDPI(hwnd uintptr) uint32 {
	if procGetDpiForWindow.Find() == nil {
		if dpi, _, _ := procGetDpiForWindow.Call(hwnd); dpi != 0 {
			return uint32(dpi)
		}
	}
	return 96
}
func fitInitialWindow(hwnd uintptr) {
	dpi := windowDPI(hwnd)
	uiScale = float32(dpi) / 96
	mon, _, _ := procMonitorFromWindow.Call(hwnd, 2)
	info := monitorInfo{Size: uint32(unsafe.Sizeof(monitorInfo{}))}
	if ok, _, _ := procGetMonitorInfo.Call(mon, uintptr(unsafe.Pointer(&info))); ok == 0 {
		updateViewport(hwnd)
		return
	}
	r := WinRect{0, 0, int32(float32(canvasWidth) * uiScale), int32(float32(canvasHeight) * uiScale)}
	if procAdjustWindowRectExForDpi.Find() == nil {
		procAdjustWindowRectExForDpi.Call(uintptr(unsafe.Pointer(&r)), windowStyle, 0, 0, uintptr(dpi))
	} else {
		procAdjustWindowRectEx.Call(uintptr(unsafe.Pointer(&r)), windowStyle, 0, 0)
	}
	w, h := r.Right-r.Left, r.Bottom-r.Top
	// Leave enough room for the taskbar and window borders on small displays.
	w = min32(w, info.Work.Right-info.Work.Left)
	h = min32(h, info.Work.Bottom-info.Work.Top)
	x := info.Work.Left + (info.Work.Right-info.Work.Left-w)/2
	y := info.Work.Top + (info.Work.Bottom-info.Work.Top-h)/2
	procSetWindowPos.Call(hwnd, 0, uintptr(x), uintptr(y), uintptr(w), uintptr(h), 0x14)
	updateViewport(hwnd)
}
func min32(a, b int32) int32 {
	if a < b {
		return a
	}
	return b
}
func max32(a, b int32) int32 {
	if a > b {
		return a
	}
	return b
}
func clampScroll(pos, content, page int32) int32 { return max32(0, min32(pos, max32(0, content-page))) }
func screenToCanvas(x, y int32) (int32, int32) {
	return int32(float32(x+scrollX) / uiScale), int32(float32(y+scrollY) / uiScale)
}
func beginViewport(hdc uintptr) func() {
	saved, _, _ := procSaveDC.Call(hdc)
	procSetGraphicsMode.Call(hdc, 2)
	t := worldTransform{M11: uiScale, M22: uiScale, DX: -float32(scrollX), DY: -float32(scrollY)}
	procSetWorldTransform.Call(hdc, uintptr(unsafe.Pointer(&t)))
	return func() { procRestoreDC.Call(hdc, saved) }
}

func visibleActions() []uiAction {
	stateMu.Lock()
	st, view, r := appState, currentView, lastResult
	pb, rb, busy := pngBusy, reportBusy, probeBusy
	page, saved := historyPage, lastSavedPath != ""
	stateMu.Unlock()
	l10Active := l10View(view)
	l10Busy := false
	if l10Active {
		d := l10Snapshot()
		l10Busy = d.Busy || d.ExportBusy
	}
	a := []uiAction{{"language", languageButtonRect, !pb && !rb && !l10Busy}}
	if languageMenuOpen {
		return append(a, uiAction{"langAuto", languageAutoRect, !pb && !rb && !l10Busy}, uiAction{"langKo", languageKoreanRect, !pb && !rb && !l10Busy}, uiAction{"langEn", languageEnglishRect, !pb && !rb && !l10Busy})
	}
	a = append(a, uiAction{"versions", footerVersionRect, !pb && !l10Active}, uiAction{"credits", footerCreditsRect, !pb && !l10Active})
	if l10Active {
		return append(a, l10Actions(view)...)
	}
	add := func(id string, r Rect, enabled bool) { a = append(a, uiAction{id, r, enabled && !pb}) }
	readable := st == "connected" || (st == "disconnected" && r.Model != "")
	if !readable && !isProjectInfoView(view) {
		add("connect", connectRect, !busy)
		add("validationWelcome", welcomeValidationRect, true)
		return a
	}
	switch view {
	case "export":
		add("public", exportPublicRect, true)
		add("verify", exportVerifyRect, true)
		add("masked", exportMaskedRect, true)
		add("last4", exportLast4Rect, true)
		add("full", exportFullRect, true)
		add("png", exportSaveRect, st == "connected" && !busy)
		add("copy", exportCopyRect, true)
		if saved {
			add("file", exportOpenFileRect, true)
			add("folder", exportOpenFolderRect, true)
		}
		add("back", backRect, true)
	case "validation":
		if cameraListPages() > 1 {
			add("cameraPrevious", cameraPreviousRect, cameraPage > 0)
			add("cameraNext", cameraNextRect, cameraPage+1 < cameraListPages())
		}
		add("versionsPage", validationVersionsRect, true)
		add("creditsPage", validationCreditsRect, true)

		add("back", validationBackRect, true)
	case "history":
		pages := (len(r.Errors) + 4) / 5
		if pages > 1 {
			add("previous", prevRect, page > 0)
			add("next", nextRect, page+1 < pages)
		}
		add("back", backRect, true)
	case "credits":
		if creditPages() > 1 {
			add("creditPrevious", cameraPreviousRect, creditsPage > 0)
			add("creditNext", cameraNextRect, creditsPage+1 < creditPages())
		}
		add("back", backRect, true)
	case "details", "guide", "cameras", "versions":
		add("back", backRect, true)
	default:
		add("serial", serialToggleRect, true)
		add("guide", guideRect, true)
		add("refresh", refreshRect, !busy)
		add("export", exportRect, st == "connected" && !busy)
		add("report", saveRect, st == "connected" && !busy && !rb)
		add("history", historyRect, true)
		add("details", detailsRect, true)
		add("validation", validationRect, true)
		if saved {
			add("file", openFileRect, true)
			add("folder", openFolderRect, true)
		}
	}
	return a
}
func actionAt(x, y int32) (uiAction, bool) {
	for _, a := range visibleActions() {
		if pointIn(a.Rect, x, y) {
			return a, true
		}
	}
	return uiAction{}, false
}
func allowActivation(x, y int32) bool {
	if a, ok := actionAt(x, y); ok {
		return a.Enabled
	}
	return languageMenuOpen // Clicking outside a menu dismisses it.
}
func settleNavigation(hwnd uintptr, before string) {
	stateMu.Lock()
	changed := currentView != before
	stateMu.Unlock()
	if changed {
		scrollX, scrollY = 0, 0
		focusAction, hoverAction = "", ""
		updateViewport(hwnd)
	}
}

func actionForRect(r Rect) (uiAction, bool) {
	for _, a := range visibleActions() {
		if a.Rect == r {
			return a, true
		}
	}
	return uiAction{}, false
}
func actionColor(r Rect, normal uintptr) uintptr {
	a, ok := actionForRect(r)
	if !ok {
		return normal
	}
	if !a.Enabled {
		rr, gg, bb := (normal & 255), (normal >> 8 & 255), (normal >> 16 & 255)
		return rgb(byte((rr+42)/2), byte((gg+42)/2), byte((bb+46)/2))
	}
	return normal
}
func decorateAction(hdc uintptr, r Rect) {
	a, ok := actionForRect(r)
	if !ok || !a.Enabled {
		return
	}
	color := uintptr(0)
	width := uintptr(1)
	if a.ID == hoverAction {
		color = rgb(123, 153, 191)
	}
	if a.ID == pressedAction {
		color = rgb(241, 163, 169)
		width = 2
	}
	if keyboardFocus && a.ID == focusAction {
		color = rgb(167, 200, 240)
		width = 2
	}
	if color == 0 {
		return
	}
	pen, _, _ := procCreatePen.Call(0, width, color)
	oldP, _, _ := procSelectObject.Call(hdc, pen)
	nullBrush, _, _ := procGetStockObject.Call(5)
	oldB, _, _ := procSelectObject.Call(hdc, nullBrush)
	procRoundRect.Call(hdc, uintptr(r.L+2), uintptr(r.T+2), uintptr(r.R-2), uintptr(r.B-2), 10, 10)
	procSelectObject.Call(hdc, oldB)
	procSelectObject.Call(hdc, oldP)
	procDeleteObject.Call(pen)
}
func paintSavedActions(hdc uintptr, export bool) {
	stateMu.Lock()
	saved := lastSavedPath != ""
	stateMu.Unlock()
	if !saved {
		return
	}
	f, d := openFileRect, openFolderRect
	if export {
		f, d = exportOpenFileRect, exportOpenFolderRect
	}
	drawOutlineButton(hdc, f, uiText("OPEN LAST FILE", "최근 저장 파일"))
	drawOutlineButton(hdc, d, uiText("OPEN SAVE FOLDER", "저장 폴더 열기"))
}
func openSavedAt(x, y int32) bool {
	a, ok := actionAt(x, y)
	if !ok || (a.ID != "file" && a.ID != "folder") {
		return false
	}
	stateMu.Lock()
	path := lastSavedPath
	stateMu.Unlock()
	if a.ID == "folder" {
		path = filepath.Dir(path)
	}
	if _, err := os.Stat(path); err != nil {
		stateMu.Lock()
		uiNotice = uiText("Saved item was moved or deleted.", "저장한 파일 또는 폴더가 이동되었거나 삭제되었습니다.")
		stateMu.Unlock()
	} else {
		result, _, _ := procShellExecute.Call(hwndMain, uintptr(unsafe.Pointer(wstr("open"))), uintptr(unsafe.Pointer(wstr(path))), 0, 0, SW_SHOW)
		if result <= 32 {
			stateMu.Lock()
			uiNotice = uiText("Could not open the saved item.", "저장한 파일 또는 폴더를 열 수 없습니다.")
			stateMu.Unlock()
		}
	}
	procInvalidateRect.Call(hwndMain, 0, 0)
	return true
}

func ensureActionVisible(hwnd uintptr, r Rect) {
	left, right := int32(float32(r.L)*uiScale), int32(float32(r.R)*uiScale)
	top, bottom := int32(float32(r.T)*uiScale), int32(float32(r.B)*uiScale)
	if left < scrollX {
		scrollX = left
	}
	if right > scrollX+clientWidth {
		scrollX = right - clientWidth
	}
	if top < scrollY {
		scrollY = top
	}
	if bottom > scrollY+clientHeight {
		scrollY = bottom - clientHeight
	}
	updateViewport(hwnd)
}
func nextFocus(actions []uiAction, current string, backward bool) uiAction {
	enabled := []uiAction{}
	index := -1
	for _, a := range actions {
		if a.Enabled {
			if a.ID == current {
				index = len(enabled)
			}
			enabled = append(enabled, a)
		}
	}
	if len(enabled) == 0 {
		return uiAction{}
	}
	if backward {
		if index < 0 {
			index = 0
		}
		index = (index - 1 + len(enabled)) % len(enabled)
	} else {
		index = (index + 1) % len(enabled)
	}
	return enabled[index]
}
func handleUIMessage(hwnd uintptr, msg uint32, wp, lp uintptr) (bool, uintptr) {
	if msg == 0x0102 && l10Input(uint32(wp)) {
		procInvalidateRect.Call(hwnd, 0, 0)
		return true, 0
	}
	if msg == WM_KEYDOWN && wp == 0x1B {
		stateMu.Lock()
		view := currentView
		stateMu.Unlock()
		if l10View(view) {
			s := l10Snapshot()
			if view == "l10report" && !s.ExportBusy {
				l10Click((l10PreviewBackRect.L+l10PreviewBackRect.R)/2, 700)
			} else if !s.ExportBusy {
				l10Click((l10CloseRect.L+l10CloseRect.R)/2, 700)
			}
			return true, 0
		}
	}
	switch msg {
	case 0x0014: // WM_ERASEBKGND: paintScene fills the complete back buffer.
		return true, 1
	case 0x0112: // WM_SYSCOMMAND: suppress maximize via title bar / system commands.
		if wp&0xFFF0 == 0xF030 {
			return true, 0
		}
		return false, 0
	case 0x0005:
		updateViewport(hwnd)
		return true, 0 // WM_SIZE
	case 0x02E0: // WM_DPICHANGED
		uiScale = float32(uint16(wp)) / 96
		if lp == 0 {
			return true, 0
		}
		var r WinRect
		procRtlMoveMemory.Call(uintptr(unsafe.Pointer(&r)), lp, unsafe.Sizeof(r))
		procSetWindowPos.Call(hwnd, 0, uintptr(r.Left), uintptr(r.Top), uintptr(r.Right-r.Left), uintptr(r.Bottom-r.Top), 0x14)
		updateViewport(hwnd)
		return true, 0
	case 0x0114, 0x0115: // Scrollbars: GetScrollInfo avoids 16-bit thumb truncation.
		axis := uintptr(0)
		pos, content, page := scrollX, int32(float32(canvasWidth)*uiScale), clientWidth
		if msg == 0x0115 {
			axis = 1
			pos, content, page = scrollY, int32(float32(canvasHeight)*uiScale), clientHeight
		}
		switch uint16(wp) {
		case 0:
			pos -= int32(32 * uiScale)
		case 1:
			pos += int32(32 * uiScale)
		case 2:
			pos -= page
		case 3:
			pos += page
		case 4, 5:
			si := scrollInfo{Size: uint32(unsafe.Sizeof(scrollInfo{})), Mask: 0x10}
			procGetScrollInfo.Call(hwnd, axis, uintptr(unsafe.Pointer(&si)))
			pos = si.TrackPos
		case 6:
			pos = 0
		case 7:
			pos = content
		}
		pos = clampScroll(pos, content, page)
		if axis == 0 {
			scrollX = pos
		} else {
			scrollY = pos
		}
		updateViewport(hwnd)
		return true, 0
	case 0x020A, 0x020E: // Wheel and horizontal wheel.
		delta := int32(int16(uint16(wp >> 16)))
		shift := wp&4 != 0
		amount := int32(float32(delta) * uiScale * 96 / 120)
		if msg == 0x020E {
			scrollX += amount
		} else if shift {
			scrollX -= amount
		} else {
			scrollY -= amount
		}
		updateViewport(hwnd)
		return true, 0
	case 0x0200: // Pointer movement, including mouse capture while pressed.
		x, y := screenToCanvas(int32(int16(uint16(lp))), int32(int16(uint16(lp>>16))))
		id := ""
		if a, ok := actionAt(x, y); ok && a.Enabled {
			id = a.ID
		}
		if hoverAction != id {
			previous := hoverAction
			hoverAction = id
			invalidateAction(hwnd, previous)
			invalidateAction(hwnd, id)
		}
		tm := trackMouse{Size: uint32(unsafe.Sizeof(trackMouse{})), Flags: 2, Hwnd: hwnd}
		procTrackMouseEvent.Call(uintptr(unsafe.Pointer(&tm)))
		return true, 0
	case 0x02A3:
		previous := hoverAction
		hoverAction = ""
		invalidateAction(hwnd, previous)
		return true, 0
	case 0x0201: // Press feedback starts on down and activates only on matching up.
		x, y := screenToCanvas(int32(int16(uint16(lp))), int32(int16(uint16(lp>>16))))
		keyboardFocus = false
		if a, ok := actionAt(x, y); ok && a.Enabled {
			pressedAction = a.ID
			focusAction = a.ID
			procSetCapture.Call(hwnd)
		} else if languageMenuOpen {
			handleClick(hwnd, x, y)
		}
		procInvalidateRect.Call(hwnd, 0, 0)
		return true, 0
	case WM_LBUTTONUP:
		x, y := screenToCanvas(int32(int16(uint16(lp))), int32(int16(uint16(lp>>16))))
		pressed := pressedAction
		pressedAction = ""
		procReleaseCapture.Call()
		if a, ok := actionAt(x, y); ok && a.Enabled && a.ID == pressed {
			handleClick(hwnd, x, y)
		}
		procInvalidateRect.Call(hwnd, 0, 0)
		return true, 0
	case 0x0215:
		pressedAction = ""
		procInvalidateRect.Call(hwnd, 0, 0)
		return true, 0 // WM_CAPTURECHANGED
	case 0x0008:
		keyboardFocus = false
		pressedAction = ""
		procInvalidateRect.Call(hwnd, 0, 0)
		return true, 0 // WM_KILLFOCUS
	case WM_KEYDOWN:
		shift, _, _ := procGetKeyState.Call(0x10)
		switch wp {
		case 9:
			a := nextFocus(visibleActions(), focusAction, int16(shift) < 0)
			focusAction = a.ID
			keyboardFocus = true
			ensureActionVisible(hwnd, a.Rect)
		case 0x0D, 0x20:
			if lp&(1<<30) != 0 {
				return true, 0
			} // Ignore held-key repeats.
			for _, a := range visibleActions() {
				if a.Enabled && a.ID == focusAction {
					handleClick(hwnd, (a.Rect.L+a.Rect.R)/2, (a.Rect.T+a.Rect.B)/2)
					break
				}
			}
		case 0x1B:
			if languageMenuOpen {
				languageMenuOpen = false
			} else {
				for _, a := range visibleActions() {
					if a.ID == "back" && a.Enabled {
						handleClick(hwnd, (a.Rect.L+a.Rect.R)/2, (a.Rect.T+a.Rect.B)/2)
						break
					}
				}
			}
		case 0x74:
			for _, a := range visibleActions() {
				if (a.ID == "refresh" || a.ID == "connect") && a.Enabled {
					handleClick(hwnd, (a.Rect.L+a.Rect.R)/2, (a.Rect.T+a.Rect.B)/2)
					break
				}
			}
		default:
			return false, 0
		}
		procInvalidateRect.Call(hwnd, 0, 0)
		return true, 0
	case WM_SETCURSOR:
		if uint16(lp) != 1 {
			return false, 0
		} // Preserve Windows resize cursors outside client area.
		cursor := uintptr(IDC_ARROW)
		var pt struct{ X, Y int32 }
		procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
		procScreenToClient.Call(hwnd, uintptr(unsafe.Pointer(&pt)))
		x, y := screenToCanvas(pt.X, pt.Y)
		if action, ok := actionAt(x, y); ok && action.Enabled {
			cursor = 32649
		}
		c, _, _ := procLoadCursorW.Call(0, cursor)
		procSetCursor.Call(c)
		return true, 1
	}
	return false, 0
}
