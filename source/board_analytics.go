//go:build windows

package main

import (
	"fmt"
	"sort"
	"time"
)

var boardAnalysisPeriod, boardAnalysisOffset int
var boardAnalysisPower bool

// The UI reads immutable record snapshots. Reuse aggregation during hover paints;
// append/delete/load publish a new slice or length, invalidating this cache.
var boardAnalysisCache struct {
	records      []boardRecord
	body, period int
	valid        bool
	result       boardAnalysis
}

func boardCachedAnalysis(s boardStore) boardAnalysis {
	c := &boardAnalysisCache
	same := len(c.records) == len(s.Records)
	if same && len(s.Records) > 0 {
		same = &c.records[0] == &s.Records[0]
	}
	if c.valid && same && c.body == boardBodyIndex && c.period == boardAnalysisPeriod {
		return c.result
	}
	c.records = s.Records
	c.body = boardBodyIndex
	c.period = boardAnalysisPeriod
	c.valid = true
	c.result = boardBuildAnalysis(s.Records, boardAnalysisBody(s), boardAnalysisPeriod, time.Local)
	return c.result
}

type boardBucket struct {
	Start          time.Time
	Shutter, Power uint64
	Pairs, Reads   int
	Estimated      bool
}
type boardAnalysis struct {
	Buckets                              map[string]boardBucket
	First, Last                          time.Time
	UnallocatedShutter, UnallocatedPower uint64
	Skipped                              int
}

func boardPeriodStart(t time.Time, period int, loc *time.Location) time.Time {
	t = t.In(loc)
	y, m, d := t.Date()
	if period == 1 {
		d = 1
	}
	if period == 2 {
		m = 1
		d = 1
	}
	return time.Date(y, m, d, 0, 0, 0, 0, loc)
}
func boardPeriodAdd(t time.Time, period, n int) time.Time {
	if period == 1 {
		return t.AddDate(0, n, 0)
	}
	if period == 2 {
		return t.AddDate(n, 0, 0)
	}
	return t.AddDate(0, 0, n)
}
func boardPeriodKey(t time.Time) string { return t.Format("2006-01-02") }
func boardBuildAnalysis(records []boardRecord, body string, period int, loc *time.Location) boardAnalysis {
	out := boardAnalysis{Buckets: map[string]boardBucket{}}
	if body == "" {
		return out
	}
	rows := []boardRecord{}
	for _, r := range records {
		if r.Body == body {
			rows = append(rows, r)
		}
	}
	sort.SliceStable(rows, func(i, j int) bool { return rows[i].Time.Before(rows[j].Time) })
	for i, r := range rows {
		start := boardPeriodStart(r.Time, period, loc)
		key := boardPeriodKey(start)
		bucket := out.Buckets[key]
		bucket.Start = start
		bucket.Reads++
		bucket.Estimated = bucket.Estimated || !r.Verified
		if out.First.IsZero() {
			out.First = start
		}
		out.Last = start
		if i > 0 {
			prev := rows[i-1]
			sh, p, ok := boardDifference(prev, r)
			if !r.Time.After(prev.Time) {
				ok = false
			}
			if !ok {
				out.Skipped++
			} else if boardPeriodStart(prev.Time, period, loc).Equal(start) {
				bucket.Shutter += uint64(sh)
				bucket.Power += uint64(p)
				bucket.Pairs++
			} else {
				out.UnallocatedShutter += uint64(sh)
				out.UnallocatedPower += uint64(p)
			}
		}
		out.Buckets[key] = bucket
	}
	return out
}
func boardAnalysisBody(s boardStore) string {
	_, b := boardHistoryRows(s, 0)
	if boardBodyIndex > 0 && boardBodyIndex < len(b) {
		return b[boardBodyIndex]
	}
	return ""
}
func boardAnalysisWindow(a boardAnalysis, period, offset int) []boardBucket {
	if a.Last.IsZero() {
		return nil
	}
	end := boardPeriodAdd(a.Last, period, -offset*8)
	start := boardPeriodAdd(end, period, -7)
	out := make([]boardBucket, 8)
	for i := range out {
		t := boardPeriodAdd(start, period, i)
		out[i] = a.Buckets[boardPeriodKey(t)]
		out[i].Start = t
	}
	return out
}
func boardAnalysisCanPrevious(s boardStore) bool {
	a := boardCachedAnalysis(s)
	if a.Last.IsZero() {
		return false
	}
	start := boardPeriodAdd(a.Last, boardAnalysisPeriod, -boardAnalysisOffset*8-7)
	return a.First.Before(start)
}
func boardAnalysisValue(b boardBucket) uint64 {
	if boardAnalysisPower {
		return b.Power
	}
	return b.Shutter
}
func boardPaintAnalysis(hdc uintptr, s boardStore) {
	boardPageTitle(hdc, boardText("analysis"))
	boardButton(hdc, "analysis_back", "back")
	drawOutlineButton(hdc, boardRects["analysis_body"], boardBodyLabel(s, boardBodyIndex)+"  ▾")
	boardButton(hdc, "analysis_prev", "previous")
	boardButton(hdc, "analysis_next", "next")
	for i, key := range []string{"day", "month", "year"} {
		r := boardRects["analysis_"+key]
		drawChoiceButton(hdc, r, boardText(key), boardAnalysisPeriod == i)
	}
	drawChoiceButton(hdc, boardRects["analysis_shutter"], boardText("shutter_short"), !boardAnalysisPower)
	drawChoiceButton(hdc, boardRects["analysis_power"], boardText("power_short"), boardAnalysisPower)
	body := boardAnalysisBody(s)
	a := boardCachedAnalysis(s)
	if body == "" || a.Last.IsZero() {
		boardParagraph(hdc, boardText("analysis_empty"), Rect{85, 382, 895, 485})
		return
	}
	buckets := boardAnalysisWindow(a, boardAnalysisPeriod, boardAnalysisOffset)
	var max, total uint64
	known := 0
	estimated := false
	for _, b := range buckets {
		v := boardAnalysisValue(b)
		if v > max {
			max = v
		}
		if b.Pairs > 0 {
			known++
			total += v
		}
		estimated = estimated || b.Estimated
	}
	sum := "—"
	if known > 0 {
		sum = "+" + formatNumber(total)
	}
	boardLabel(hdc, boardText("observed_total")+" "+sum, Rect{60, 337, 620, 365}, fontLabel)
	if estimated {
		boardLabel(hdc, boardText("estimated"), Rect{640, 337, 920, 365}, fontTiny)
	}
	color := rgb(193, 92, 109)
	if boardAnalysisPower {
		color = rgb(106, 166, 220)
	}
	drawHLine(hdc, 75, 523, 905, rgb(72, 80, 94))
	for i, b := range buckets {
		x := int32(80 + i*104)
		v := boardAnalysisValue(b)
		label := "—"
		if b.Pairs > 0 {
			label = formatNumber(v)
			height := int32(0)
			if max > 0 {
				height = int32(float64(v) / float64(max) * 130)
			}
			if height < 2 {
				height = 2
			}
			fillRound(hdc, Rect{x + 20, 522 - height, x + 66, 522}, color, 4)
		}
		boardCenter(hdc, label, Rect{x - 5, 365, x + 92, 390}, fontTiny)
		date := b.Start.Format("01-02")
		if boardAnalysisPeriod == 1 {
			date = b.Start.Format("2006-01")
		}
		if boardAnalysisPeriod == 2 {
			date = b.Start.Format("2006")
		}
		boardCenter(hdc, date, Rect{x - 5, 531, x + 92, 552}, fontTiny)
	}
	boardCenter(hdc, buckets[0].Start.Format("2006-01-02")+" — "+boardPeriodAdd(buckets[7].Start, boardAnalysisPeriod, 1).AddDate(0, 0, -1).Format("2006-01-02"), Rect{60, 555, 920, 579}, fontTiny)
	boardLabel(hdc, fmt.Sprintf(boardText("unallocated"), formatNumber(a.UnallocatedShutter), formatNumber(a.UnallocatedPower)), Rect{60, 584, 920, 609}, fontTiny)
	boardParagraph(hdc, fmt.Sprintf(boardText("analysis_note"), a.Skipped), Rect{60, 614, 920, 668})
}
func init() {
	for key, r := range map[string]Rect{
		"analysis_back": {40, 136, 80, 176}, "analysis_body": {60, 250, 660, 284}, "analysis_prev": {680, 250, 795, 284}, "analysis_next": {805, 250, 920, 284},
		"analysis_day": {60, 292, 160, 328}, "analysis_month": {170, 292, 270, 328}, "analysis_year": {280, 292, 380, 328},
		"analysis_shutter": {480, 292, 685, 328}, "analysis_power": {695, 292, 920, 328},
	} {
		boardRects[key] = r
	}
}
