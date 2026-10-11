//go:build windows

package main

import (
	"fmt"
	"strings"
)

var boardRawPage, boardBodyPage int
var boardRawReturn = "main"

func boardCurrentDiagnostic(latest, failure boardReply) boardDiagnostic {
	if latest.Diagnostic.Started.IsZero() {
		return failure.Diagnostic
	}
	return latest.Diagnostic
}

// Show only already parsed numeric fields, never device paths, serials or
// unknown binary payloads. An empty response is distinct from a returned zero.
func boardRawLines(d boardDiagnostic) []string {
	lines := []string{}
	words := func(label string, values []uint16) {
		if len(values) == 0 {
			lines = append(lines, label+" · "+boardText("unavailable"))
			return
		}
		lines = append(lines, label+" · "+boardText("raw_short"))
		for i, v := range values {
			lines = append(lines, fmt.Sprintf("  [%d]  %d   (0x%04X)", i+1, v, v))
		}
	}
	c := d.Camera
	if c != nil {
		if c.CardsRead {
			words(boardText("cards_space")+" · 0x12000015", c.Cards)
		}
		for i, s := range c.Stores {
			lines = append(lines, fmt.Sprintf("%s %d · ID 0x%08X", boardText("camera_store"), i+1, s.ID))
			if !s.Read {
				key := "unavailable"
				if s.InfoSkipped {
					key = "storage_skipped"
				}
				lines = append(lines, "  "+boardText(key))
				continue
			}
			lines = append(lines, fmt.Sprintf("  %s: %d bytes", boardText("capacity"), s.Capacity), fmt.Sprintf("  %s: %d bytes", boardText("free"), s.Free))
			lines = append(lines, fmt.Sprintf("  Type: %d · Filesystem: %d · Access: %d", s.Type, s.Filesystem, s.Access))
		}
	}
	lines = append(lines, "— "+boardText("trace_title")+" —")
	for _, s := range d.Steps {
		lines = append(lines, strings.TrimSpace(s.Stage+" · "+s.Status+" · "+s.PTP+" "+s.Code+" "+s.ErrorSource))
	}
	return lines
}

func boardPaintRaw(hdc uintptr, latest, failure boardReply) {
	d := boardCurrentDiagnostic(latest, failure)
	boardPageTitle(hdc, boardText("raw_view"))
	stamp := "—"
	if !d.Started.IsZero() {
		stamp = d.Started.Format("2006-01-02 15:04:05")
	}
	boardLabel(hdc, d.Model+" · "+d.Firmware+" · "+stamp, Rect{60, 250, 920, 275}, fontSmall)
	boardParagraph(hdc, boardText("raw_compare"), Rect{60, 280, 920, 317})
	lines := boardRawLines(d)
	pages := (len(lines) + 7) / 8
	if boardRawPage >= pages {
		boardRawPage = 0
	}
	for i := 0; i < 8 && boardRawPage*8+i < len(lines); i++ {
		y := int32(324 + i*28)
		boardLabel(hdc, lines[boardRawPage*8+i], Rect{70, y, 910, y + 27}, fontSmall)
	}
	boardButton(hdc, "raw_prev", "previous")
	boardButton(hdc, "raw_next", "next")
	boardLabel(hdc, fmt.Sprintf("%d / %d", boardRawPage+1, pages), Rect{465, 556, 565, 588}, fontSmall)
	boardButton(hdc, "raw_save", "save_diag")
	boardButton(hdc, "raw_back", "back")
}

func boardBodyLabel(s boardStore, body int) string {
	rows, _ := boardHistoryRows(s, body)
	if body == 0 {
		return boardText("all")
	}
	if len(rows) == 0 {
		return boardText("none")
	}
	return fmt.Sprintf("%s · ****%s · %s %d", rows[0].Model, rows[0].SerialSuffix, boardText("body"), body)
}

func boardPaintBodies(hdc uintptr, s boardStore) {
	boardPageTitle(hdc, boardText("pick_body"))
	_, bodies := boardHistoryRows(s, 0)
	pages := (len(bodies) + 5) / 6
	if boardBodyPage >= pages {
		boardBodyPage = 0
	}
	for i := 0; i < 6 && boardBodyPage*6+i < len(bodies); i++ {
		index := boardBodyPage*6 + i
		label := boardBodyLabel(s, index)
		if index == boardBodyIndex {
			label = "✓ " + label
		}
		drawOutlineButton(hdc, boardBodyRect(i), label)
	}
	boardButton(hdc, "raw_prev", "previous")
	boardButton(hdc, "raw_next", "next")
	boardLabel(hdc, fmt.Sprintf("%d / %d", boardBodyPage+1, pages), Rect{465, 556, 565, 588}, fontSmall)
	boardButton(hdc, "raw_back", "back")
}

func boardBodyRect(i int) Rect { y := int32(252 + i*48); return Rect{60, y, 920, y + 40} }

func boardStageLabel(stage string) string {
	key := map[string]string{"device_info": "stage_identity", "identity": "stage_identity", "open_device": "stage_identity", "open_session": "stage_session", "session_policy": "stage_session", "usage": "stage_counters", "parse_usage": "stage_counters", "card_status": "cards_space", "storage_ids": "cards_space", "storage_info": "cards_space", "close_session": "stage_close", "release_device": "stage_close", "worker": "stage_process"}[stage]
	if key == "" {
		return boardText("stage_process")
	}
	return boardText(key)
}

func boardPaintDiagnosticSummary(hdc uintptr, d boardDiagnostic, status string, busy bool) {
	groups := [][]string{{"device_info"}, {"usage", "parse_usage"}, {"release_device", "close_session"}}
	titles := []string{"stage_identity", "stage_counters", "stage_close"}
	for i, stages := range groups {
		value := "—"
		failed := false
		for _, step := range d.Steps {
			for _, stage := range stages {
				if step.Stage == stage {
					if step.Status == "failed" || step.Status == "unconfirmed" {
						failed = true
					}
					if step.Status == "ok" {
						value = "✓"
					}
				}
			}
		}
		if failed {
			value = "!"
		}
		x := int32(60 + i*293)
		fillRound(hdc, Rect{x, 393, x + 274, 449}, rgb(34, 40, 50), 8)
		boardLabel(hdc, boardText(titles[i]), Rect{x + 12, 398, x + 232, 443}, fontSmall)
		boardCenter(hdc, value, Rect{x + 235, 398, x + 266, 443}, fontLabel)
	}
	boardLabel(hdc, boardText(status), Rect{60, 455, 920, 479}, fontSmall)
	if busy {
		boardLabel(hdc, boardStageLabel(d.ActiveStage), Rect{60, 486, 920, 524}, fontSmall)
	} else if d.FailedStage != "" {
		boardParagraph(hdc, boardText("failed_step")+" · "+boardStageLabel(d.FailedStage)+". "+boardText("diag_retry_short"), Rect{60, 486, 920, 527})
	} else {
		boardParagraph(hdc, boardText("diag_send_short"), Rect{60, 486, 920, 527})
	}
	boardButton(hdc, "diag_raw", "raw_short")
}

func boardExportReport(r UsageResult, mode string) string {
	validator, date := validationDetails(r)
	status := "UNVERIFIED MODEL / FIRMWARE"
	if r.SemanticsVerified {
		status = "VERIFIED " + r.Model + " / " + r.Firmware
	}
	if !r.SemanticsVerified && r.VerifiedModel {
		status = "MODEL VERIFIED / FIRMWARE UNVERIFIED"
	}
	return strings.ReplaceAll(buildReport(r, status, validator, date), serialForMode(r.Serial, "last4"), serialForMode(r.Serial, mode))
}
