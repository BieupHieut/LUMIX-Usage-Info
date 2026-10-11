//go:build windows

package main

import (
	"reflect"
	"testing"
)

func TestStablePublicQueryScope(t *testing.T) {
	req := boardPublicRequest(false, false)
	f := &boardFake{model: "DC-S5", usageRC: PTP_RC_OK}
	r := boardQueryOptions(f, req.Session, req.Extras, func(boardReply) {})
	if !r.Valid || !reflect.DeepEqual(f.calls, []string{"1001:[]", "9414:[15C00010]"}) || f.closeCount != 1 {
		t.Fatalf("stable read queried unexpected fields or lost counters: %v", f.calls)
	}
	oldView, oldBusy, oldPNG, oldReport, oldMenu := currentView, probeBusy, pngBusy, reportBusy, languageMenuOpen
	defer func() {
		currentView, probeBusy, pngBusy, reportBusy, languageMenuOpen = oldView, oldBusy, oldPNG, oldReport, oldMenu
	}()
	currentView = "diagnostics"
	probeBusy, pngBusy, reportBusy, languageMenuOpen = false, false, false, false
	for _, a := range boardActions() {
		if a.ID == "board_extras" || a.ID == "board_card_view" {
			t.Fatal("card experiment is publicly reachable")
		}
	}
}
