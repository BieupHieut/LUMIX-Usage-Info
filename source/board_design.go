//go:build windows

package main

import "unsafe"

// Shared geometry: all content pages start with the same title baseline and rule.
var boardPanelRect = Rect{40, 184, 940, 675}
var boardTitleRect = Rect{60, 198, 520, 236}
var procMoveToEx = gdi32.NewProc("MoveToEx")
var procLineTo = gdi32.NewProc("LineTo")

func boardPanel(hdc uintptr) { fillRound(hdc, boardPanelRect, rgb(28, 28, 31), 16) }
func boardPageTitle(hdc uintptr, title string) {
	boardLabel(hdc, title, boardTitleRect, fontSection)
	drawHLine(hdc, 60, 244, 920, rgb(50, 53, 59))
}
func boardPrimaryButton(hdc uintptr, r Rect, text string) {
	fillRound(hdc, r, actionColor(r, rgb(161, 62, 79)), 10)
	setFont(hdc, fontButton)
	drawText(hdc, text, r, DT_CENTER|DT_VCENTER|DT_SINGLELINE, actionColor(r, rgb(250, 245, 246)))
	decorateAction(hdc, r)
}
func boardBackIcon(hdc uintptr) {
	r := backRect
	fillRound(hdc, r, actionColor(r, rgb(36, 39, 46)), 40)
	pen, _, _ := procCreatePen.Call(0, 2, actionColor(r, rgb(221, 230, 242)))
	old, _, _ := procSelectObject.Call(hdc, pen)
	x, y := (r.L+r.R)/2, (r.T+r.B)/2
	procMoveToEx.Call(hdc, uintptr(x+8), uintptr(y), 0)
	procLineTo.Call(hdc, uintptr(x-8), uintptr(y))
	procMoveToEx.Call(hdc, uintptr(x-1), uintptr(y-7), 0)
	procLineTo.Call(hdc, uintptr(x-8), uintptr(y))
	procLineTo.Call(hdc, uintptr(x-1), uintptr(y+7))
	procSelectObject.Call(hdc, old)
	procDeleteObject.Call(pen)
	decorateAction(hdc, r)
}
func boardTextLink(hdc uintptr, text string, r Rect, right bool) {
	setFont(hdc, fontTiny)
	flags := uint32(DT_LEFT | DT_SINGLELINE | DT_VCENTER)
	if right {
		flags = DT_RIGHT | DT_SINGLELINE | DT_VCENTER
	}
	color := rgb(170, 186, 206)
	a, ok := actionForRect(r)
	active := ok && a.Enabled && (a.ID == hoverAction || a.ID == pressedAction || keyboardFocus && a.ID == focusAction)
	if active {
		color = rgb(225, 236, 250)
	}
	drawText(hdc, text, r, flags, color)
	if active {
		wr := WinRect{0, 0, 1000, 100}
		ptr := wstr(localizeText(text))
		procDrawTextW.Call(hdc, uintptr(unsafe.Pointer(ptr)), ^uintptr(0), uintptr(unsafe.Pointer(&wr)), DT_SINGLELINE|DT_NOPREFIX|0x400)
		x := r.L
		if right {
			x = r.R - wr.Right
		}
		drawHLine(hdc, x, (r.T+r.B)/2+9, x+wr.Right, color)
	}
}

func boardIsBackAction(id string) bool {
	switch id {
	case "back", "board_settings_back", "board_raw_back", "board_card_back", "board_analysis_back":
		return true
	}
	return false
}
