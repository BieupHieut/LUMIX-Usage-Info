//go:build windows

package main

import (
	"runtime"
	"sync/atomic"
	"syscall"
	"testing"
	"unsafe"
)

func TestProductTitleBothLanguages(t *testing.T) {
	prior := atomic.LoadUint32(&languageMode)
	defer atomic.StoreUint32(&languageMode, prior)
	for _, tc := range []struct {
		mode uint32
		want string
	}{{languageKorean, "LUMERIAN-LUMIX 사용 정보"}, {languageEnglish, "LUMERIAN-LUMIX Usage Info"}} {
		atomic.StoreUint32(&languageMode, tc.mode)
		if productTitle() != tc.want {
			t.Fatal(productTitle())
		}
	}
	if windowStyle&0x10000 != 0 {
		t.Fatal("maximize box still present")
	}
	if handled, result := handleUIMessage(0, 0x0112, 0xF030, 0); !handled || result != 0 {
		t.Fatal("maximize command not blocked")
	}
	if handled, _ := handleUIMessage(0, 0x0112, 0xF020, 0); handled {
		t.Fatal("minimize command blocked")
	}
	if handled, result := handleUIMessage(0, 0x0014, 0, 0); !handled || result != 1 {
		t.Fatal("background erasure not suppressed")
	}
}

func TestNativeHoverInvalidatesOnlyChangedButtons(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	oldScale, oldX, oldY, oldW, oldH, oldHwnd := uiScale, scrollX, scrollY, clientWidth, clientHeight, hwndMain
	oldHover, oldMenu := hoverAction, languageMenuOpen
	stateMu.Lock()
	oldState, oldView, oldResult, oldProbe, oldPNG, oldReport := appState, currentView, lastResult, probeBusy, pngBusy, reportBusy
	appState, currentView, lastResult, probeBusy, pngBusy, reportBusy = "connected", "main", UsageResult{Model: "DC-S1RM2", Firmware: "Ver. 1.5"}, false, false, false
	stateMu.Unlock()
	defer func() {
		uiScale, scrollX, scrollY, clientWidth, clientHeight, hwndMain = oldScale, oldX, oldY, oldW, oldH, oldHwnd
		hoverAction, languageMenuOpen = oldHover, oldMenu
		stateMu.Lock()
		appState, currentView, lastResult, probeBusy, pngBusy, reportBusy = oldState, oldView, oldResult, oldProbe, oldPNG, oldReport
		stateMu.Unlock()
	}()
	uiScale = 1
	scrollX, scrollY = 0, 0
	hoverAction = ""
	languageMenuOpen = false
	inst, _, _ := procGetModuleHandleW.Call(0)
	cls := wstr("LumerianHiddenHoverPaintingTest")
	eraseCount, paintCount := 0, 0
	callback := syscall.NewCallback(func(hwnd uintptr, msg uint32, wp, lp uintptr) uintptr {
		if msg == 0x0014 {
			eraseCount++
		}
		if msg == WM_PAINT {
			paintWindow(hwnd)
			paintCount++
			return 0
		}
		if handled, result := handleUIMessage(hwnd, msg, wp, lp); handled {
			return result
		}
		result, _, _ := procDefWindowProcW.Call(hwnd, uintptr(msg), wp, lp)
		return result
	})
	wc := WndClassEx{CbSize: uint32(unsafe.Sizeof(WndClassEx{})), LpfnWndProc: callback, HInstance: inst, LpszClassName: cls}
	if atom, _, _ := procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc))); atom == 0 {
		t.Fatal("register failed")
	}
	defer user32.NewProc("UnregisterClassW").Call(uintptr(unsafe.Pointer(cls)), inst)
	hwnd, _, _ := procCreateWindowExW.Call(0, uintptr(unsafe.Pointer(cls)), uintptr(unsafe.Pointer(wstr("hidden hover test"))), windowStyle, 0, 0, 1100, 900, 0, 0, inst, 0)
	if hwnd == 0 {
		t.Fatal("create failed")
	}
	hwndMain = hwnd
	defer user32.NewProc("DestroyWindow").Call(hwnd)
	style, _, _ := procGetWindowStyle.Call(hwnd, ^uintptr(15))
	if style&0x10000 != 0 {
		t.Fatal("native maximize box still present")
	}
	validate := user32.NewProc("ValidateRect")
	priorRequest := requestActionRepaint
	defer func() { requestActionRepaint = priorRequest }()
	damage := WinRect{}
	invalidations := 0
	requestActionRepaint = func(hwnd uintptr, r WinRect) { damage = r; invalidations++; priorRequest(hwnd, r) }
	send := user32.NewProc("SendMessageW")
	validate.Call(hwnd, 0)
	eraseCount = 0
	move := func(r Rect) {
		x, y := (r.L+r.R)/2, (r.T+r.B)/2
		send.Call(hwnd, 0x0200, 0, uintptr(uint16(x))|uintptr(uint16(y))<<16)
	}
	move(refreshRect)
	if invalidations != 1 || hoverAction != "refresh" {
		t.Fatalf("hover invalidations=%d action=%q", invalidations, hoverAction)
	}
	expected := actionDamageRect(refreshRect)
	if damage != expected {
		t.Fatalf("hover invalidated wrong region: got %+v want %+v", damage, expected)
	}
	send.Call(hwnd, WM_PAINT, 0, 0)
	move(refreshRect)
	if invalidations != 1 {
		t.Fatal("same button movement repainting")
	}
	for i := 0; i < 20; i++ {
		move(exportRect)
		send.Call(hwnd, WM_PAINT, 0, 0)
		move(refreshRect)
		send.Call(hwnd, WM_PAINT, 0, 0)
	}
	if eraseCount != 0 || paintCount < 40 {
		t.Fatalf("hover erasures=%d paints=%d", eraseCount, paintCount)
	}
	send.Call(hwnd, 0x02A3, 0, 0)
	if hoverAction != "" {
		t.Fatal("mouse leave did not clear hover")
	}
	if damage != expected {
		t.Fatal("mouse leave did not invalidate old button only")
	}

}
