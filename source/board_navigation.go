//go:build windows

package main

import "strings"

// Protected by stateMu. Tabs select roots; secondary pages remember their caller.
var boardNavigationStack []string

func boardEnterLocked(target string, root bool) {
	if root {
		boardNavigationStack = nil
	} else if currentView != target {
		boardNavigationStack = append(boardNavigationStack, currentView)
	}
	currentView = target
	boardHelpPinned = false
	uiNotice = ""
}
func boardBackLocked() {
	target := "main"
	if n := len(boardNavigationStack); n > 0 {
		target = boardNavigationStack[n-1]
		boardNavigationStack = boardNavigationStack[:n-1]
	}
	currentView = target
	boardHelpPinned = false
	uiNotice = ""
}
func boardNavigationClick(id string) bool {
	stateMu.Lock()
	defer stateMu.Unlock()
	if strings.HasPrefix(id, "tab_") {
		target := strings.TrimPrefix(id, "tab_")
		for _, tab := range boardTabs {
			if target == tab.view {
				boardEnterLocked(target, true)
				return true
			}
		}
		return false
	}
	switch id {
	case "back", "board_settings_back", "board_raw_back", "board_card_back", "board_analysis_back":
		boardBackLocked()
		return true
	}
	targets := map[string]string{
		"credits": "credits", "creditsPage": "credits", "versions": "versions", "versionsPage": "versions",
		"board_history_analysis": "analytics", "board_analysis_body": "history_bodies",
		"board_history_settings": "storage", "board_filter": "history_bodies",
		"board_home_export": "export", "board_home_details": "details", "board_home_errors": "history",
		"board_diag_raw": "query_raw", "board_home_raw": "query_raw", "board_card_view": "card_info",
	}
	if target, ok := targets[id]; ok {
		if target == "analytics" {
			if boardBodyIndex == 0 {
				_, bodies := boardHistoryRows(boardData, 0)
				if len(bodies) > 1 {
					boardBodyIndex = 1
					boardSelectedRecords = map[string]bool{}
				}
			}
			boardAnalysisOffset = 0
		}
		if target == "query_raw" {
			boardRawReturn = currentView
			boardRawPage = 0
		}
		if target == "history_bodies" {
			boardBodyPage = boardBodyIndex / 6
		}
		boardEnterLocked(target, false)
		return true
	}
	return false
}
func boardTabForView(view string) string {
	switch view {
	case "main", "export", "details", "history", "guide":
		return "main"
	case "usage_history", "storage", "history_bodies", "analytics":
		return "usage_history"
	case "diagnostics", "query_raw", "card_info":
		return "diagnostics"
	case "validation", "cameras":
		return "validation"
	}
	return ""
}
func boardSelectedTab(view string) string {
	if tab := boardTabForView(view); tab != "" {
		return tab
	}
	stateMu.Lock()
	defer stateMu.Unlock()
	for i := len(boardNavigationStack) - 1; i >= 0; i-- {
		if tab := boardTabForView(boardNavigationStack[i]); tab != "" {
			return tab
		}
	}
	return ""
}
func init() {
	boardStrings["compat_tab"] = [4]string{"Compatibility", "기종 검증", "機種の検証", "Compatibilidad"}
}
