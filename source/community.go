//go:build windows

package main

import (
	"fmt"
	"strings"
)

type readSuccessReport struct {
	Model, Firmware, By, Date, Mode, Tool string
}

var validationBackRect = Rect{652, 675, 910, 724}

var readSuccessReports = []readSuccessReport{
	{"DC-L10", "1.2", "잠이든", "2026-10-02", "LUMIX Lab", "beta.6"},
	{"DC-S1M2", "1.4", "잠이든", "2026-10-02", "Not supplied", "Not supplied"},
}

func readSuccessDetails(r UsageResult) (readSuccessReport, bool) {
	model := strings.ToUpper(strings.NewReplacer("-", "", " ", "", "_", "").Replace(r.Model))
	fw := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(r.Firmware), " ", ""))
	fw = strings.TrimPrefix(strings.TrimPrefix(fw, "VER."), "V")
	for _, report := range readSuccessReports {
		key := strings.ReplaceAll(report.Model, "-", "")
		if (model == key || model == strings.TrimPrefix(key, "DC")) && (fw == report.Firmware || fw == report.Firmware+"0") {
			return report, true
		}
	}
	return readSuccessReport{}, false
}

func writeReadSuccess(b *strings.Builder, r UsageResult, korean bool) {
	report, ok := readSuccessDetails(r)
	if !ok {
		return
	}
	mode, tool := report.Mode, report.Tool
	if korean {
		if mode == "Not supplied" {
			mode = "미제공"
		}
		if tool == "Not supplied" {
			tool = "미제공"
		}
		fmt.Fprintf(b, "\r\n읽기 성공 제보         확인자 %s · 접수일 %s\r\n제보 연결 모드         %s · 제보 툴 %s\r\n제보 출처             파나소닉 루믹스 카페 포럼 (회원 제보)\r\n                      https://cafe.naver.com/panalumix\r\n위 연결 모드는 과거 성공 사례의 제보값이며 현재 USB 모드의 자동 감지값이 아닙니다.\r\n커뮤니티 확인으로 셔터·전원/깨우기 의미를 검증했습니다. 카운터 3/4는 추정 명칭이며 5~7은 의미 미확인입니다.\r\n\r\n", report.By, report.Date, mode, tool)
	} else {
		fmt.Fprintf(b, "\r\nRead success reported by %s / Received %s\r\nReported USB mode        %s / Reported tool %s\r\nCommunity                Panasonic LUMIX Café Forum (member report)\r\n                         https://cafe.naver.com/panalumix\r\nReported mode describes the prior success case, not automatic detection of the current USB mode.\r\nShutter and power/wake meanings verified by community confirmation; counters 3/4 are estimated and 5–7 unidentified.\r\n\r\n", report.By, report.Date, mode, tool)
	}
}

// A verified row represents one exact model/firmware pair. Candidates are
// model-level connection possibilities, never additional verified profiles.
const cameraRowsPerPage = 7

var creditsPage int

func creditPages() int { return (len(cameraData.Profiles) + 4) / 5 }

type cameraListItem struct {
	cameraProfile
	Verified bool
}

var estimatedSupportModels = []string{
	"DC-BS1H", "DC-S1", "DC-S1R", "DC-S1M2", "DC-S1M2ES", "DC-S1RM2",
	"DC-S1H", "DC-S5", "DC-S5M2", "DC-S5M2X", "DC-S9", "DC-BGH1",
	"DC-GH5", "DC-GH5S", "DC-GH5M2", "DC-G9", "DC-G9M2", "DC-GH6", "DC-GH7",
}

func buildCameraList() []cameraListItem {
	items := make([]cameraListItem, 0, len(cameraData.Profiles)+len(estimatedSupportModels))
	represented := make(map[string]bool)
	for _, profile := range cameraData.Profiles {
		items = append(items, cameraListItem{profile, true})
		represented[profile.Model] = true
	}
	for _, model := range estimatedSupportModels {
		if !represented[model] {
			items = append(items, cameraListItem{cameraProfile{Model: model}, false})
			represented[model] = true
		}
	}
	return items
}

func cameraListPages() int {
	return (len(buildCameraList()) + cameraRowsPerPage - 1) / cameraRowsPerPage
}

func paintCameraMatrix(hdc uintptr, r UsageResult) {
	items := buildCameraList()
	pages := cameraListPages()
	if cameraPage < 0 || cameraPage >= pages {
		cameraPage = 0
	}
	fillRound(hdc, Rect{40, 180, 940, 655}, rgb(28, 28, 31), 18)
	setFont(hdc, fontSection)
	drawText(hdc, "CAMERA LIST & VALIDATION", Rect{70, 204, 430, 239}, DT_VCENTER|DT_SINGLELINE, rgb(242, 242, 245))
	setFont(hdc, fontTiny)
	drawText(hdc, fmt.Sprintf(uiText("%d verified · %d estimated", "검증 %d · 지원 추정 %d"), len(cameraData.Profiles), len(items)-len(cameraData.Profiles)), Rect{430, 210, 720, 239}, DT_RIGHT|DT_VCENTER|DT_SINGLELINE, rgb(174, 197, 225))
	if pages > 1 {
		drawChoiceButton(hdc, cameraPreviousRect, "‹", false)
		drawChoiceButton(hdc, cameraNextRect, "›", false)
	}
	drawText(hdc, fmt.Sprintf("%d/%d", cameraPage+1, pages), Rect{855, 210, 908, 239}, DT_CENTER|DT_VCENTER|DT_SINGLELINE, rgb(174, 197, 225))
	drawHLine(hdc, 70, 244, 910, rgb(50, 50, 55))
	start := cameraPage * cameraRowsPerPage
	end := start + cameraRowsPerPage
	if end > len(items) {
		end = len(items)
	}
	for i, item := range items[start:end] {
		y := int32(258 + i*48)
		style := cameraRowStyle(item.Verified)
		fillRound(hdc, Rect{70, y, 910, y + 46}, style.Background, 10)
		fillRound(hdc, Rect{78, y + 10, 81, y + 36}, style.Accent, 3)
		badge := uiText("ESTIMATED", "지원 추정")
		title := item.Model
		detail := uiText("PC(Tether) · Counter meanings unverified", "PC(테더) · 카운터 의미 미검증")
		by := uiText("Not yet verified", "확인 제보 대기")
		badgeColor := style.Accent
		if item.Verified {
			badge = uiText("✓ VERIFIED", "✓ 검증됨")
			title += " · " + uiText("FW ", "펌웨어 ") + item.Firmware
			detail = uiText("Shutter ✓   Power / Wake ✓", "셔터 ✓   전원/깨우기 ✓")
			by = uiText("Validated by ", "확인: ") + item.By
			badgeColor = style.Accent
		}
		setFont(hdc, fontSmall)
		drawText(hdc, badge, Rect{84, y + 10, 195, y + 37}, DT_CENTER|DT_VCENTER|DT_SINGLELINE, badgeColor)
		setFont(hdc, fontLabel)
		drawText(hdc, title, Rect{214, y + 2, 585, y + 26}, DT_VCENTER|DT_SINGLELINE|0x8000, rgb(245, 245, 249))
		setFont(hdc, fontTiny)
		drawText(hdc, detail, Rect{214, y + 26, 605, y + 44}, DT_VCENTER|DT_SINGLELINE, style.Detail)
		setFont(hdc, fontSmall)
		drawText(hdc, by, Rect{610, y + 11, 887, y + 36}, DT_RIGHT|DT_VCENTER|DT_SINGLELINE|0x8000, rgb(220, 224, 234))
	}
	setFont(hdc, fontTiny)
	note := uiText("Verified for the listed firmware only. Other entries: estimated support list.", "검증은 표시된 펌웨어에 한정됩니다. 나머지 항목은 지원 추정 목록입니다.")
	color := rgb(146, 168, 195)
	if cameraDataWarning {
		note = uiText("Invalid external camera data. Using bundled verification data.", "외부 기종 데이터가 잘못되어 내장 검증 데이터를 사용합니다.")
		color = rgb(240, 160, 160)
	}
	drawText(hdc, note, Rect{72, 603, 906, 625}, DT_CENTER|DT_VCENTER|DT_SINGLELINE, color)
	drawText(hdc, uiText("Counters 3/4 estimated · 5–7 and error mappings unverified.", "카운터 3/4는 추정 · 5~7과 오류 해석은 미검증"), Rect{72, 628, 906, 650}, DT_CENTER|DT_VCENTER|DT_SINGLELINE, rgb(146, 168, 195))
	drawOutlineButton(hdc, validationVersionsRect, "VERSION HISTORY")
	drawOutlineButton(hdc, validationCreditsRect, "DEVELOPER CREDITS")
	drawOutlineButton(hdc, validationBackRect, "BACK")
	drawFooter(hdc, r, isLiveConnection())
}

func paintCredits(hdc uintptr, r UsageResult) {
	if creditsPage < 0 || creditsPage >= creditPages() {
		creditsPage = 0
	}
	fillRound(hdc, Rect{40, 180, 940, 655}, rgb(28, 28, 31), 18)
	setFont(hdc, fontSection)
	drawText(hdc, "CREATOR & CONTRIBUTORS", Rect{70, 204, 540, 235}, DT_VCENTER|DT_SINGLELINE, rgb(242, 242, 245))
	if creditPages() > 1 {
		drawChoiceButton(hdc, cameraPreviousRect, "‹", false)
		drawChoiceButton(hdc, cameraNextRect, "›", false)
		setFont(hdc, fontTiny)
		drawText(hdc, fmt.Sprintf("%d/%d", creditsPage+1, creditPages()), Rect{855, 210, 908, 239}, DT_CENTER|DT_VCENTER|DT_SINGLELINE, rgb(174, 197, 225))
	}
	drawHLine(hdc, 70, 244, 910, rgb(50, 50, 55))
	// Equal-width columns; creator information stays entirely in the left half.
	fillRound(hdc, Rect{70, 263, 480, 622}, rgb(34, 34, 38), 14)
	fillRound(hdc, Rect{500, 263, 910, 622}, rgb(34, 34, 38), 14)
	setFont(hdc, fontSection)
	drawText(hdc, uiText("CREATOR", "제작자 정보"), Rect{88, 276, 462, 308}, DT_CENTER|DT_VCENTER|DT_SINGLELINE, rgb(167, 200, 240))
	drawBitmap(hdc, creatorBitmap, 143, 310)
	setFont(hdc, fontTitle)
	drawText(hdc, "@bieup_hieut", Rect{88, 550, 462, 588}, DT_CENTER|DT_VCENTER|DT_SINGLELINE, rgb(245, 245, 247))
	setFont(hdc, fontSmall)
	drawText(hdc, uiText("Creator · Product design", "제작 · 제품 디자인"), Rect{88, 590, 462, 616}, DT_CENTER|DT_VCENTER|DT_SINGLELINE, rgb(173, 181, 194))

	setFont(hdc, fontSection)
	drawText(hdc, uiText("CONTRIBUTORS", "도움을 주신 분들"), Rect{518, 276, 892, 308}, DT_CENTER|DT_VCENTER|DT_SINGLELINE, rgb(167, 200, 240))
	setFont(hdc, fontTiny)
	drawText(hdc, uiText("Camera tests and feedback", "카메라 확인 및 피드백"), Rect{518, 310, 892, 332}, DT_CENTER|DT_VCENTER|DT_SINGLELINE, rgb(167, 180, 203))
	start := creditsPage * 5
	end := start + 5
	if end > len(cameraData.Profiles) {
		end = len(cameraData.Profiles)
	}
	for i, item := range cameraData.Profiles[start:end] {
		y := int32(342 + i*30)
		setFont(hdc, fontSmall)
		drawText(hdc, item.Model+" / "+item.Firmware, Rect{518, y, 688, y + 28}, DT_VCENTER|DT_SINGLELINE|0x8000, rgb(166, 177, 195))
		drawText(hdc, item.By, Rect{698, y, 892, y + 28}, DT_RIGHT|DT_VCENTER|DT_SINGLELINE|0x8000, rgb(242, 242, 246))
	}
	drawHLine(hdc, 518, 505, 892, rgb(55, 55, 61))
	setFont(hdc, fontLabel)
	drawText(hdc, "SPECIAL THANKS", Rect{518, 515, 892, 545}, DT_CENTER|DT_VCENTER|DT_SINGLELINE, rgb(167, 200, 240))
	setFont(hdc, fontSmall)
	drawText(hdc, uiText("Panasonic LUMIX Café Forum", "파나소닉 루믹스 카페 포럼 회원 여러분"), Rect{518, 548, 892, 572}, DT_CENTER|DT_VCENTER|DT_SINGLELINE, rgb(230, 230, 236))
	setFont(hdc, fontTiny)
	drawText(hdc, uiText("Thank you to community members for tests and feedback.", "카메라 확인과 피드백에 감사드립니다."), Rect{518, 573, 892, 595}, DT_CENTER|DT_VCENTER|DT_SINGLELINE, rgb(188, 195, 207))
	drawText(hdc, "cafe.naver.com/panalumix", Rect{518, 596, 892, 616}, DT_CENTER|DT_VCENTER|DT_SINGLELINE, rgb(167, 180, 203))
	drawText(hdc, "Unofficial read-only community tool · Not affiliated with Panasonic", Rect{72, 628, 906, 652}, DT_CENTER|DT_VCENTER|DT_SINGLELINE, rgb(145, 153, 166))
	drawOutlineButton(hdc, backRect, "BACK")
	drawFooter(hdc, r, isLiveConnection())
}
