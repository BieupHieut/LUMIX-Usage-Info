//go:build windows

package main

import "unsafe"

const wsHScroll, wsVScroll = 0x00100000, 0x00200000

var (
	procGetWindowStyle         = user32.NewProc("GetWindowLongW")
	procShowScrollBar          = user32.NewProc("ShowScrollBar")
	procGetSystemMetrics       = user32.NewProc("GetSystemMetrics")
	procGetSystemMetricsForDpi = user32.NewProc("GetSystemMetricsForDpi")
)

// Start from the available area without bars. Existing bars must not create
// a circular dependency where each makes the other appear necessary.
func scrollbarLayout(cw, ch, bw, bh, vbar, hbar int32) (horizontal, vertical bool) {
	for i := 0; i < 3; i++ {
		pw, ph := bw, bh
		if vertical {
			pw -= vbar
		}
		if horizontal {
			ph -= hbar
		}
		nextH, nextV := cw > pw, ch > ph
		if nextH == horizontal && nextV == vertical {
			break
		}
		horizontal, vertical = nextH, nextV
	}
	return
}

func scrollbarMetric(index uintptr) int32 {
	var n uintptr
	if procGetSystemMetricsForDpi.Find() == nil {
		n, _, _ = procGetSystemMetricsForDpi.Call(index, uintptr(96*uiScale))
	} else {
		n, _, _ = procGetSystemMetrics.Call(index)
	}
	return int32(n)
}

func updateViewport(hwnd uintptr) {
	if viewportUpdating {
		return
	}
	viewportUpdating = true
	defer func() { viewportUpdating = false }()
	var cr WinRect
	procGetClientRect.Call(hwnd, uintptr(unsafe.Pointer(&cr)))
	clientWidth, clientHeight = cr.Right, cr.Bottom
	if clientWidth <= 0 || clientHeight <= 0 {
		return
	}
	style, _, _ := procGetWindowStyle.Call(hwnd, ^uintptr(15)) // GWL_STYLE = -16
	baseW, baseH := clientWidth, clientHeight
	vbar, hbar := scrollbarMetric(2), scrollbarMetric(3)
	if style&wsVScroll != 0 {
		baseW += vbar
	}
	if style&wsHScroll != 0 {
		baseH += hbar
	}
	cw, ch := int32(float32(canvasWidth)*uiScale), int32(float32(canvasHeight)*uiScale)
	needH, needV := scrollbarLayout(cw, ch, baseW, baseH, vbar, hbar)
	for _, a := range []struct {
		axis            uintptr
		needed, visible bool
	}{{0, needH, style&wsHScroll != 0}, {1, needV, style&wsVScroll != 0}} {
		if a.needed != a.visible {
			show := uintptr(0)
			if a.needed {
				show = 1
			}
			procShowScrollBar.Call(hwnd, a.axis, show)
		}
	}
	procGetClientRect.Call(hwnd, uintptr(unsafe.Pointer(&cr)))
	clientWidth, clientHeight = cr.Right, cr.Bottom
	scrollX = clampScroll(scrollX, cw, clientWidth)
	scrollY = clampScroll(scrollY, ch, clientHeight)
	for _, a := range []struct {
		axis               uintptr
		content, page, pos int32
	}{{0, cw, clientWidth, scrollX}, {1, ch, clientHeight, scrollY}} {
		si := scrollInfo{Size: uint32(unsafe.Sizeof(scrollInfo{})), Mask: 7, Max: a.content - 1, Page: uint32(max32(0, a.page)), Pos: a.pos}
		procSetScrollInfo.Call(hwnd, a.axis, uintptr(unsafe.Pointer(&si)), 1)
	}
	hoverAction = ""
	procInvalidateRect.Call(hwnd, 0, 0)
}
