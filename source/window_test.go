//go:build windows

package main

import (
	"runtime"
	"syscall"
	"testing"
	"unsafe"
)

// Exercise our real Win32 resize/input handlers in a hidden test window. No
// camera probe, native save dialog or external file viewer is invoked here.
func TestNativeViewportAndKeyboardNavigation(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	oldScale, oldX, oldY, oldW, oldH, oldHwnd := uiScale, scrollX, scrollY, clientWidth, clientHeight, hwndMain
	oldFocus, oldHover, oldPress, oldKeyboard, oldMenu := focusAction, hoverAction, pressedAction, keyboardFocus, languageMenuOpen
	stateMu.Lock()
	oldState, oldView, oldResult, oldBusy := appState, currentView, lastResult, pngBusy
	appState, currentView, lastResult, pngBusy = "waiting", "validation", UsageResult{}, false
	stateMu.Unlock()
	defer func() {
		uiScale, scrollX, scrollY, clientWidth, clientHeight, hwndMain = oldScale, oldX, oldY, oldW, oldH, oldHwnd
		focusAction, hoverAction, pressedAction, keyboardFocus, languageMenuOpen = oldFocus, oldHover, oldPress, oldKeyboard, oldMenu
		stateMu.Lock()
		appState, currentView, lastResult, pngBusy = oldState, oldView, oldResult, oldBusy
		stateMu.Unlock()
	}()
	uiScale = 1.25
	scrollX, scrollY = 0, 0
	languageMenuOpen = false
	inst, _, _ := procGetModuleHandleW.Call(0)
	cls := wstr("LumerianBeta15HiddenViewportTest")
	callback := syscall.NewCallback(func(hwnd uintptr, msg uint32, wp, lp uintptr) uintptr {
		if handled, result := handleUIMessage(hwnd, msg, wp, lp); handled {
			return result
		}
		result, _, _ := procDefWindowProcW.Call(hwnd, uintptr(msg), wp, lp)
		return result
	})
	wc := WndClassEx{CbSize: uint32(unsafe.Sizeof(WndClassEx{})), LpfnWndProc: callback, HInstance: inst, LpszClassName: cls}
	if atom, _, _ := procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc))); atom == 0 {
		t.Fatal("cannot register test window")
	}
	defer user32.NewProc("UnregisterClassW").Call(uintptr(unsafe.Pointer(cls)), inst)
	hwnd, _, _ := procCreateWindowExW.Call(0, uintptr(unsafe.Pointer(cls)), uintptr(unsafe.Pointer(wstr("hidden UI test"))), windowStyle, 0, 0, 650, 500, 0, 0, inst, 0)
	if hwnd == 0 {
		t.Fatal("cannot create test window")
	}
	hwndMain = hwnd
	defer user32.NewProc("DestroyWindow").Call(hwnd)
	updateViewport(hwnd)
	if clientWidth < 1 || clientHeight < 1 {
		t.Fatal("empty native client area")
	}
	scrollX, scrollY = 100000, 100000
	updateViewport(hwnd)
	if scrollX != int32(980*uiScale)-clientWidth || scrollY != int32(780*uiScale)-clientHeight {
		t.Fatalf("native scrollbars cannot reach edge: %d,%d", scrollX, scrollY)
	}
	focusAction = "versionsPage"
	send := user32.NewProc("SendMessageW")
	send.Call(hwnd, WM_KEYDOWN, 0x0D, 0)
	stateMu.Lock()
	view := currentView
	stateMu.Unlock()
	if view != "versions" || scrollX != 0 || scrollY != 0 {
		t.Fatalf("Enter navigation failed: %s scroll=%d,%d", view, scrollX, scrollY)
	}
	// Focus traversal scrolls the footer into view instead of focusing hidden text.
	focusAction = "versions"
	send.Call(hwnd, WM_KEYDOWN, 9, 0)
	if focusAction != "credits" || !keyboardFocus || scrollY == 0 {
		t.Fatal("Tab did not reveal footer action")
	}
	send.Call(hwnd, WM_KEYDOWN, 0x1B, 0)
	stateMu.Lock()
	view = currentView
	stateMu.Unlock()
	if view != "validation" {
		t.Fatalf("Escape did not return to validation: %s", view)
	}
	// Exercise the production activation handler on our own hidden window.
	oldPage, oldCredits, oldRegistry := cameraPage, creditsPage, cameraData
	cameraPage = 0
	for p := 1; p < cameraListPages(); p++ {
		handleClick(hwnd, cameraNextRect.L+2, cameraNextRect.T+2)
		if cameraPage != p {
			t.Fatal("next camera page inaccessible")
		}
	}
	handleClick(hwnd, cameraNextRect.L+2, cameraNextRect.T+2)
	if cameraPage != cameraListPages()-1 {
		t.Fatal("camera pager exceeded last page")
	}
	for cameraPage > 0 {
		handleClick(hwnd, cameraPreviousRect.L+2, cameraPreviousRect.T+2)
	}
	cameraData.Profiles = append(append([]cameraProfile{}, cameraData.Profiles...), cameraProfile{Model: "DC-S9", Firmware: "2.0", By: "TEST FIXTURE"})
	handleClick(hwnd, validationCreditsRect.L+2, validationCreditsRect.T+2)
	creditsPage = 0
	handleClick(hwnd, cameraNextRect.L+2, cameraNextRect.T+2)
	if creditsPage != 1 {
		t.Fatal("future contributor page inaccessible")
	}
	handleClick(hwnd, cameraPreviousRect.L+2, cameraPreviousRect.T+2)
	if creditsPage != 0 {
		t.Fatal("previous contributor page inaccessible")
	}
	handleClick(hwnd, backRect.L+2, backRect.T+2)
	cameraPage, creditsPage, cameraData = oldPage, oldCredits, oldRegistry
	// A wider viewport removes both bars and resets out-of-range positions.
	procSetWindowPos.Call(hwnd, 0, 0, 0, 1600, 1200, 0x16)
	updateViewport(hwnd)
	if scrollX != 0 || scrollY != 0 {
		t.Fatal("larger window retained stale scroll offsets")
	}
	// At an exact content fit, legacy visible bars must not keep each other alive.
	uiScale = 1
	fitRect := WinRect{0, 0, canvasWidth, canvasHeight}
	procAdjustWindowRectEx.Call(uintptr(unsafe.Pointer(&fitRect)), windowStyle, 0, 0)
	procSetWindowPos.Call(hwnd, 0, 0, 0, uintptr(fitRect.Right-fitRect.Left), uintptr(fitRect.Bottom-fitRect.Top), 0x16)
	procShowScrollBar.Call(hwnd, 3, 1)
	updateViewport(hwnd)
	style, _, _ := procGetWindowStyle.Call(hwnd, ^uintptr(15))
	if style&(wsHScroll|wsVScroll) != 0 || clientWidth != canvasWidth || clientHeight != canvasHeight {
		t.Fatalf("exact fit retained bars: style=%x client=%dx%d", style, clientWidth, clientHeight)
	}
	procSetWindowPos.Call(hwnd, 0, 120, 130, 0, 0, 0x15) // Move only: no resize.
	updateViewport(hwnd)
	style, _, _ = procGetWindowStyle.Call(hwnd, ^uintptr(15))
	if style&(wsHScroll|wsVScroll) != 0 || scrollX != 0 || scrollY != 0 {
		t.Fatal("moving a fitting window introduced scrollbars")
	}
	// Verify WM_DPICHANGED consumes Windows' suggested rectangle and DPI.
	suggested := WinRect{10, 20, 1210, 920}
	send.Call(hwnd, 0x02E0, 192|(192<<16), uintptr(unsafe.Pointer(&suggested)))
	if uiScale != 2 || clientWidth < 1 || clientHeight < 1 {
		t.Fatal("DPI change did not update the viewport")
	}
}
