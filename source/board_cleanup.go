//go:build windows

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
	"unsafe"
)

// Retain the first/latest observation for every body and all change boundaries.
// Compare chronological neighbours, never records from different cameras.
func boardCompactRecords(records []boardRecord, body string) ([]boardRecord, int) {
	groups := map[string][]int{}
	for i, r := range records {
		if body == "" || r.Body == body {
			groups[r.Body] = append(groups[r.Body], i)
		}
	}
	remove := map[int]bool{}
	for _, indices := range groups {
		sort.SliceStable(indices, func(i, j int) bool { return records[indices[i]].Time.Before(records[indices[j]].Time) })
		for j := 1; j < len(indices)-1; j++ {
			a, b, c := records[indices[j-1]], records[indices[j]], records[indices[j+1]]
			sh, p, ok := boardDifference(a, b)
			// Preserve both sides of firmware/verification/model/counter-reset boundaries.
			nextSh, nextP, nextOK := boardDifference(b, c)
			// Keep the last observation before an increase and daily endpoints.
			// Otherwise cleanup can turn a known daily/monthly/yearly increase
			// into an unallocated interval, or erase an observed zero day.
			sameDay := boardPeriodStart(a.Time, 0, time.Local).Equal(boardPeriodStart(b.Time, 0, time.Local)) && boardPeriodStart(b.Time, 0, time.Local).Equal(boardPeriodStart(c.Time, 0, time.Local))
			if ok && nextOK && sh == 0 && p == 0 && nextSh == 0 && nextP == 0 && sameDay && a.Time.Before(b.Time) && b.Time.Before(c.Time) {
				remove[indices[j]] = true
			}
		}
	}
	out := make([]boardRecord, 0, len(records)-len(remove))
	for i, r := range records {
		if !remove[i] {
			out = append(out, r)
		}
	}
	return out, len(remove)
}

func boardCommitCleanup(path string, before, after boardStore) (string, error) {
	b, e := json.MarshalIndent(before, "", "  ")
	if e != nil {
		return "", e
	}
	dir := filepath.Join(filepath.Dir(path), "history-backups")
	if e = os.MkdirAll(dir, 0700); e != nil {
		return "", e
	}
	f, e := os.CreateTemp(dir, "before-cleanup_"+time.Now().Format("20060102_150405")+"_*.json")
	if e != nil {
		return "", e
	}
	backup := f.Name()
	if _, e = f.Write(b); e != nil {
		f.Close()
		return backup, e
	}
	if e = f.Sync(); e != nil {
		f.Close()
		return backup, e
	}
	if e = f.Close(); e != nil {
		return backup, e
	}
	return backup, boardPersist(path, after)
}

func boardCleanHistory() {
	stateMu.Lock()
	if probeBusy || boardStoreError != nil {
		stateMu.Unlock()
		return
	}
	before := boardData
	before.Records = append([]boardRecord(nil), before.Records...)
	_, bodies := boardHistoryRows(before, 0)
	body := ""
	if boardBodyIndex > 0 && boardBodyIndex < len(bodies) {
		body = bodies[boardBodyIndex]
	}
	label := boardBodyLabel(before, boardBodyIndex)
	stateMu.Unlock()
	records, n := boardCompactRecords(before.Records, body)
	if n == 0 {
		stateMu.Lock()
		boardNotice = "clean_none"
		stateMu.Unlock()
		return
	}
	msg := fmt.Sprintf(boardText("clean_confirm"), label, n)
	answer, _, _ := procMessageBoxW.Call(hwndMain, uintptr(unsafe.Pointer(wstr(msg))), uintptr(unsafe.Pointer(wstr(boardText("clean_history")))), 0x4|0x100|0x30)
	if answer != 6 {
		return
	}
	stateMu.Lock()
	defer stateMu.Unlock()
	oldBytes, _ := json.Marshal(before)
	nowBytes, _ := json.Marshal(boardData)
	if probeBusy || string(oldBytes) != string(nowBytes) {
		boardNotice = "clean_changed"
		return
	}
	after := before
	after.Records = records
	// Preserve external copies until the user chooses to update them.
	after.ExternalAuto = false
	if _, e := boardCommitCleanup(boardStorePath(), before, after); e != nil {
		boardNotice = "save_failed"
		return
	}
	boardSelectedRecords = map[string]bool{}
	boardData = after
	boardPage = 0
	boardNotice = "clean_done"
}

var boardHelpPinned bool
var boardHelpRect = Rect{330, 302, 918, 466}

func boardPaintShutterHelp(hdc uintptr) {
	fillRound(hdc, boardHelpRect, rgb(44, 55, 72), 10)
	boardLabel(hdc, boardText("shutter_help_title"), Rect{348, 310, 901, 339}, fontLabel)
	boardParagraph(hdc, boardText("shutter_help_body"), Rect{348, 347, 901, 455})
}
