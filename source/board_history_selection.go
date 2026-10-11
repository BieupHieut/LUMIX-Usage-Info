//go:build windows

package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"unsafe"
)

var boardSelectedRecords = map[string]bool{}
var boardDeletionPath = boardStorePath
var boardConfirmDelete = func(message string) bool {
	answer, _, _ := procMessageBoxW.Call(hwndMain, uintptr(unsafe.Pointer(wstr(message))), uintptr(unsafe.Pointer(wstr(boardText("delete_selected")))), 0x4|0x100|0x30)
	return answer == 6
}

func boardSelectionCount(s boardStore) int {
	n := 0
	for _, r := range s.Records {
		if boardSelectedRecords[r.ID] {
			n++
		}
	}
	return n
}
func boardRemoveSelected(records []boardRecord, selected map[string]bool) ([]boardRecord, int) {
	groups := map[string][]int{}
	removed := 0
	for i, r := range records {
		groups[r.Body] = append(groups[r.Body], i)
		if selected[r.ID] {
			removed++
		}
	}
	out := append([]boardRecord(nil), records...)
	for _, indices := range groups {
		sort.SliceStable(indices, func(i, j int) bool { return records[indices[i]].Time.Before(records[indices[j]].Time) })
		gap := false
		for _, i := range indices {
			if selected[records[i].ID] {
				gap = true
				continue
			}
			if gap {
				out[i].BreakBefore = true
				gap = false
			}
		}
	}
	kept := make([]boardRecord, 0, len(records)-removed)
	for _, r := range out {
		if !selected[r.ID] {
			kept = append(kept, r)
		}
	}
	return kept, removed
}
func boardDeleteSelected() {
	stateMu.Lock()
	if probeBusy || boardStoreError != nil {
		stateMu.Unlock()
		return
	}
	before := boardData
	before.Records = append([]boardRecord(nil), before.Records...)
	selected := map[string]bool{}
	for id, on := range boardSelectedRecords {
		selected[id] = on
	}
	records, n := boardRemoveSelected(before.Records, selected)
	stateMu.Unlock()
	if n == 0 {
		return
	}
	if !boardConfirmDelete(fmt.Sprintf(boardText("delete_confirm"), n)) {
		return
	}
	stateMu.Lock()
	defer stateMu.Unlock()
	a, _ := json.Marshal(before)
	b, _ := json.Marshal(boardData)
	if probeBusy || string(a) != string(b) {
		boardNotice = "clean_changed"
		return
	}
	after := before
	// Schema 2 prevents older builds from dropping deletion-boundary metadata.
	after.Schema = 2
	after.Records = records
	after.ExternalAuto = false
	if _, err := boardCommitCleanup(boardDeletionPath(), before, after); err != nil {
		boardNotice = "save_failed"
		return
	}
	_, oldBodies := boardHistoryRows(before, 0)
	body := ""
	if boardBodyIndex > 0 && boardBodyIndex < len(oldBodies) {
		body = oldBodies[boardBodyIndex]
	}
	boardData = after
	boardSelectedRecords = map[string]bool{}
	boardPage = 0
	boardAnalysisOffset = 0
	boardBodyIndex = 0
	_, bodies := boardHistoryRows(after, 0)
	for i, id := range bodies {
		if id == body {
			boardBodyIndex = i
			break
		}
	}
	boardNotice = "delete_done"
}
func boardHistoryCheckRect(i int) Rect { y := int32(288 + i*68); return Rect{72, y + 16, 104, y + 47} }
func boardPaintHistory(hdc uintptr, s boardStore) {
	boardPageTitle(hdc, boardText("history"))
	boardButton(hdc, "history_analysis", "analysis")
	boardButton(hdc, "history_settings", "settings_short")
	rows, bodies := boardHistoryRows(s, boardBodyIndex)
	if boardPage*5 >= len(rows) {
		boardPage = 0
	}
	drawOutlineButton(hdc, boardRects["filter"], boardBodyLabel(s, boardBodyIndex)+"  ▾")
	boardButton(hdc, "hprev", "previous")
	boardButton(hdc, "hnext", "next")
	if len(rows) == 0 {
		boardLabel(hdc, boardText("none"), Rect{60, 310, 920, 360}, fontLabel)
	}
	for i := 0; i < 5 && boardPage*5+i < len(rows); i++ {
		n := boardPage*5 + i
		v := rows[n]
		y := int32(288 + i*68)
		color := rgb(34, 40, 50)
		if boardSelectedRecords[v.ID] {
			color = rgb(45, 65, 85)
		}
		fillRound(hdc, Rect{60, y, 920, y + 61}, color, 10)
		check := "□"
		if boardSelectedRecords[v.ID] {
			check = "✓"
		}
		drawOutlineButton(hdc, boardHistoryCheckRect(i), check)
		bodyNumber := 0
		for j, b := range bodies {
			if b == v.Body {
				bodyNumber = j
				break
			}
		}
		badge := ""
		if !v.Verified {
			badge = "? "
		}
		boardLabel(hdc, fmt.Sprintf("%s%s · ****%s · %s %d", badge, v.Model, v.SerialSuffix, boardText("body"), bodyNumber), Rect{116, y + 3, 515, y + 29}, fontSmall)
		boardLabel(hdc, v.Time.Format("2006-01-02 15:04:05"), Rect{535, y + 3, 905, y + 29}, fontSmall)
		delta := boardText("boundary")
		for j := n + 1; j < len(rows); j++ {
			if rows[j].Body == v.Body {
				sh, p, ok := boardDifference(rows[j], v)
				if ok {
					delta = fmt.Sprintf("Δ %s +%d · %s +%d", boardText("shutter_short"), sh, boardText("power_short"), p)
				}
				break
			}
		}
		if v.BreakBefore {
			delta = boardText("deleted_gap")
		}
		boardLabel(hdc, fmt.Sprintf("%s %s · %s %s · %s", boardText("shutter_short"), formatNumber(uint64(v.Shutter)), boardText("power_short"), formatNumber(uint64(v.Power)), v.Firmware), Rect{116, y + 29, 515, y + 57}, fontTiny)
		boardLabel(hdc, delta, Rect{535, y + 29, 905, y + 57}, fontTiny)
	}
	drawOutlineButton(hdc, boardRects["history_delete"], fmt.Sprintf("%s (%d)", boardText("delete_selected"), boardSelectionCount(s)))
	boardButton(hdc, "history_clear", "clear_selection")
	boardButton(hdc, "history_clean", "clean_short")
	boardButton(hdc, "history_folder", "folder_short")
}
