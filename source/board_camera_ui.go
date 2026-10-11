//go:build windows

package main

import (
	"fmt"
	"strings"
)

var boardStoragePage int
var boardSerialMode int // Screen only: hidden, last four, full. Reset for every query.

func boardScreenSerial(serial string) string {
	if !knownSerial(serial) {
		return boardText("unavailable")
	}
	mode := "masked"
	switch boardSerialMode {
	case 1:
		mode = "last4"
	case 2:
		mode = "full"
	}
	return serialForMode(serial, mode)
}

func boardCardSummary(c *boardCameraStatus) string {
	if c == nil || !c.CardsRead {
		return boardText("unavailable")
	}
	if len(c.Cards) == 0 {
		return boardText("no_cards_reported")
	}
	parts := []string{}
	for i, v := range c.Cards {
		if i == 2 {
			parts = append(parts, "…")
			break
		}
		parts = append(parts, fmt.Sprintf("%d: %s", i+1, boardText(boardCardStateKey(v))))
	}
	return strings.Join(parts, " · ")
}

func boardCenter(hdc uintptr, text string, r Rect, font uintptr) {
	setFont(hdc, font)
	drawText(hdc, text, r, DT_CENTER|DT_VCENTER|DT_SINGLELINE|0x8000, rgb(229, 234, 244))
}
func boardWelcomeMascot(hdc uintptr) {
	b := mascotBitmap
	if b.Handle == 0 || b.Width <= 0 || b.Height <= 0 {
		return
	}
	w := int32(200)
	h := w * b.Height / b.Width
	mem, _, _ := procCreateCompatibleDC.Call(hdc)
	if mem == 0 {
		return
	}
	old, _, _ := procSelectObject.Call(mem, b.Handle)
	mode, _, _ := procSetStretchBltMode.Call(hdc, 4)
	procStretchBlt.Call(hdc, uintptr((980-w)/2), uintptr(246), uintptr(w), uintptr(h), mem, 0, 0, uintptr(b.Width), uintptr(b.Height), SRCCOPY)
	procSetStretchBltMode.Call(hdc, mode)
	procSelectObject.Call(mem, old)
	procDeleteDC.Call(mem)
}
func boardPaintCamera(hdc uintptr, r UsageResult, latest boardReply) {
	if r.Model == "" {
		boardCenter(hdc, boardText("welcome_title"), Rect{60, 182, 920, 223}, fontTitle)
		boardCenter(hdc, boardText("welcome_subtitle"), Rect{60, 227, 920, 256}, fontSmall)
		boardWelcomeMascot(hdc)
		boardCenter(hdc, boardText("welcome_usb"), Rect{60, 445, 920, 473}, fontSmall)
		boardCenter(hdc, boardText("welcome_l10"), Rect{60, 472, 920, 494}, fontTiny)
		key := "welcome_connect"
		if probeBusy {
			key = "reading"
		}
		boardButton(hdc, "welcome_read", key)
		boardCenter(hdc, boardText("safe"), Rect{60, 565, 920, 590}, fontSmall)
		boardCenter(hdc, boardText("close_apps"), Rect{60, 589, 920, 611}, fontTiny)

		return
	}
	boardPageTitle(hdc, boardText("overview"))
	boardButton(hdc, "read", "refresh_read")
	for i, title := range []string{"my_camera", "shutter", "power"} {
		x := int32(60 + i*294)
		fillRound(hdc, Rect{x, 260, x + 272, 447}, rgb(34, 40, 50), 12)
		boardCenter(hdc, boardText(title), Rect{x + 12, 270, x + 244, 306}, fontSmall)
		if i == 0 {
			boardCenter(hdc, r.Model+" · "+r.Firmware, Rect{x + 10, 319, x + 262, 355}, fontLabel)
			sr := boardRects["serial"]
			fillRound(hdc, sr, actionColor(sr, rgb(43, 48, 57)), 8)
			boardCenter(hdc, boardText("serial_change"), Rect{sr.L + 6, sr.T + 3, sr.R - 6, sr.T + 25}, fontTiny)
			v := boardScreenSerial(r.Serial)
			boardCenter(hdc, v, Rect{sr.L + 6, sr.T + 26, sr.R - 6, sr.B - 3}, serialDisplayFont(v, boardSerialMode == 2))
			decorateAction(hdc, sr)
		} else {
			v := "—"
			if latest.Valid {
				if i == 1 {
					v = formatNumber(uint64(r.Shutter))
				} else {
					v = formatNumber(uint64(r.PowerWake))
				}
			}
			numberFont := fontCounter
			if len(v) > 10 {
				numberFont = fontTitle
			}
			boardCenter(hdc, v, Rect{x + 12, 319, x + 260, 380}, numberFont)
			boardCenter(hdc, boardText("recorded_total"), Rect{x + 10, 382, x + 262, 410}, fontTiny)
			if !latest.Valid {
				boardCenter(hdc, boardText("unavailable"), Rect{x + 10, 411, x + 262, 438}, fontTiny)
			} else if !r.SemanticsVerified {
				boardCenter(hdc, boardText("estimated"), Rect{x + 10, 411, x + 262, 438}, fontTiny)
			}
		}
	}
	drawOutlineButton(hdc, boardRects["shutter_help"], "!")
	boardCenter(hdc, boardText("safe"), Rect{60, 580, 920, 610}, fontSmall)
	if !r.Timestamp.IsZero() {
		boardLabel(hdc, boardText("snapshot")+" · "+r.Timestamp.Format("2006-01-02 15:04:05"), Rect{60, 458, 920, 489}, fontTiny)
	}
	if latest.Valid {
		for _, b := range []struct{ id, key string }{{"home_export", "share"}, {"home_details", "details_short"}, {"home_errors", "errors_short"}} {
			boardButton(hdc, b.id, b.key)
		}
	}

}
func boardPaintCardTrial(hdc uintptr, c *boardCameraStatus) {
	boardPageTitle(hdc, boardText("card_trial"))
	boardLabel(hdc, boardCardSummary(c), Rect{60, 245, 815, 279}, fontSmall)
	if c != nil && len(c.Stores) > 1 {
		drawOutlineButton(hdc, boardRects["next_store"], fmt.Sprintf("%d/%d  ›", boardStoragePage%len(c.Stores)+1, len(c.Stores)))
	}
	if c == nil || !c.StorageRead {
		boardLabel(hdc, boardText("space_unavailable"), Rect{60, 320, 920, 355}, fontLabel)
	} else if len(c.Stores) == 0 {
		boardLabel(hdc, boardText("no_storage"), Rect{60, 320, 920, 355}, fontLabel)
	} else {
		s := c.Stores[boardStoragePage%len(c.Stores)]
		boardLabel(hdc, fmt.Sprintf("%s %d", boardText("camera_store"), boardStoragePage%len(c.Stores)+1), Rect{60, 303, 920, 335}, fontSmall)
		free, total := "—", "—"
		if s.spaceKnown() {
			free = fmt.Sprintf("%.1f GB", float64(s.Free)/1e9)
		}
		if s.capacityKnown() {
			total = fmt.Sprintf("%.1f GB", float64(s.Capacity)/1e9)
		}
		boardLabel(hdc, boardText("free")+" "+free, Rect{60, 351, 920, 397}, fontTitle)
		if s.InfoSkipped {
			boardParagraph(hdc, boardText("storage_unexposed"), Rect{60, 412, 920, 476})
		} else {
			boardLabel(hdc, boardText("capacity")+" "+total, Rect{60, 412, 920, 446}, fontLabel)
		}
	}
	boardButton(hdc, "diag_raw", "raw_view")
	boardButton(hdc, "card_back", "back")
}
