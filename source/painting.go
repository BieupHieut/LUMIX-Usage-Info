//go:build windows

package main

import (
	"math"
	"unsafe"
)

func productTitle() string {
	return uiText("LUMERIAN-LUMIX Usage Info", "LUMERIAN-LUMIX 사용 정보")
}

// Render a complete frame off screen. Windows clips the single final transfer
// to the invalid region, so no background clearing or intermediate card paint
// is exposed when the pointer crosses a button.
func paintBuffered(hdc uintptr, width, height int32) bool {
	if width <= 0 || height <= 0 {
		return true
	}
	mem, _, _ := procCreateCompatibleDC.Call(hdc)
	if mem == 0 {
		paintScene(hdc, width, height)
		return false
	}
	defer procDeleteDC.Call(mem)
	bmp, _, _ := procCreateCompatibleBitmap.Call(hdc, uintptr(width), uintptr(height))
	if bmp == 0 {
		paintScene(hdc, width, height)
		return false
	}
	defer procDeleteObject.Call(bmp)
	old, _, _ := procSelectObject.Call(mem, bmp)
	if old == 0 || old == ^uintptr(0) {
		paintScene(hdc, width, height)
		return false
	}
	defer procSelectObject.Call(mem, old)
	paintScene(mem, width, height)
	ok, _, _ := procBitBlt.Call(hdc, 0, 0, uintptr(width), uintptr(height), mem, 0, 0, SRCCOPY)
	if ok == 0 {
		paintScene(hdc, width, height)
		return false
	}
	return true
}

func actionDamageRect(r Rect) WinRect {
	// Round outwards so fractional DPI and the action border are fully updated.
	return WinRect{
		Left:   int32(math.Floor(float64(r.L)*float64(uiScale))) - 2 - scrollX,
		Top:    int32(math.Floor(float64(r.T)*float64(uiScale))) - 2 - scrollY,
		Right:  int32(math.Ceil(float64(r.R)*float64(uiScale))) + 2 - scrollX,
		Bottom: int32(math.Ceil(float64(r.B)*float64(uiScale))) + 2 - scrollY,
	}
}

var requestActionRepaint = func(hwnd uintptr, r WinRect) {
	procInvalidateRect.Call(hwnd, uintptr(unsafe.Pointer(&r)), 0)
}

func invalidateAction(hwnd uintptr, id string) {
	if id == "" {
		return
	}
	for _, a := range visibleActions() {
		if a.ID == id {
			r := actionDamageRect(a.Rect)
			requestActionRepaint(hwnd, r)
			return
		}
	}
	// Navigation can make a previously hovered action disappear.
	procInvalidateRect.Call(hwnd, 0, 0)
}
