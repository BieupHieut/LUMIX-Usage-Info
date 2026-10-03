package main

import "testing"

func TestScrollbarLayoutUsesSpaceWithoutExistingBars(t *testing.T) {
	for _, tc := range []struct {
		name                 string
		cw, ch, w, h         int32
		horizontal, vertical bool
	}{
		{"exact fit", 980, 780, 980, 780, false, false},
		{"large", 980, 780, 1200, 1000, false, false},
		{"horizontal only", 980, 780, 970, 810, true, false},
		{"vertical only", 980, 780, 1010, 770, false, true},
		{"vertical forces horizontal", 980, 780, 980, 770, true, true},
		{"horizontal forces vertical", 980, 780, 970, 780, true, true},
		{"small", 980, 780, 640, 480, true, true},
		{"125 percent exact", 1225, 975, 1225, 975, false, false},
		{"200 percent exact", 1960, 1560, 1960, 1560, false, false},
	} {
		h, v := scrollbarLayout(tc.cw, tc.ch, tc.w, tc.h, 17, 17)
		if h != tc.horizontal || v != tc.vertical {
			t.Errorf("%s: bars=%v,%v", tc.name, h, v)
		}
	}
}
