//go:build windows

package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

type l10Session struct {
	Firmware                    string
	Fixture                     bool
	BaselineStatus              string
	Baseline                    []l10Candidate
	USBStatus, USBReason        string
	Candidates                  []l10Candidate
	Selected, Page, PreviewPage int
	PTP                         l10PTPResult
	Read                        l10ReadResult
	Identity                    string
	BConsent, CConsent          bool
	Busy, ExportBusy            bool
	Running                     string
	Cancel                      context.CancelFunc
	Notice                      string
	Preview                     *l10Report
}

var (
	l10State           = l10NewSession()
	l10EntryRect       = Rect{290, 600, 690, 640}
	l10FirmwareRect    = Rect{270, 283, 410, 317}
	l10ModeRect        = Rect{710, 283, 910, 317}
	l10ARect           = Rect{70, 330, 450, 370}
	l10BaselineRect    = Rect{530, 330, 910, 370}
	l10BRect           = Rect{70, 523, 450, 565}
	l10CRect           = Rect{530, 523, 910, 565}
	l10PrevRect        = Rect{70, 485, 175, 515}
	l10NextRect        = Rect{805, 485, 910, 515}
	l10PreviewRect     = Rect{70, 675, 450, 724}
	l10CloseRect       = Rect{530, 675, 910, 724}
	l10PreviewSaveRect = Rect{355, 675, 675, 724}
	l10PreviewBackRect = Rect{70, 675, 285, 724}
	l10PreviewNextRect = Rect{805, 607, 910, 640}
	l10PreviewPrevRect = Rect{70, 607, 175, 640}
)

func l10NewSession() l10Session {
	return l10Session{Selected: -1, BaselineStatus: "NOT_RUN", USBStatus: "NOT_RUN", PTP: l10PTPResult{Status: "NOT_RUN", Operations: []uint16{}}, Read: l10ReadResult{Status: "NOT_RUN"}}
}
func l10View(view string) bool { return view == "l10" || view == "l10report" }
func l10RowRect(i int) Rect    { y := int32(383 + i*33); return Rect{70, y, 910, y + 29} }
func l10Snapshot() l10Session  { stateMu.Lock(); defer stateMu.Unlock(); return l10State }
func l10CanB(s l10Session) bool {
	return !s.Busy && !s.ExportBusy && s.PTP.Status == "NOT_RUN" && s.USBStatus == "PASS" && s.Selected >= 0 && s.Selected < len(s.Candidates) && l10AllowedCandidate(s.Candidates[s.Selected]) && s.Candidates[s.Selected].USB.WPDStatus == "PASS" && len(s.Candidates[s.Selected].WPDIDs) == 1
}
func l10CanC(s l10Session) bool {
	return !s.Busy && !s.ExportBusy && s.Selected >= 0 && s.Selected < len(s.Candidates) && l10AllowedCandidate(s.Candidates[s.Selected]) && s.BConsent && s.PTP.Status == "PASS" && s.PTP.Model == "DC-L10" && (len(s.Identity) == 64 || s.Fixture && s.Identity == "FIXTURE_ONLY") && l10HasOperation(s.PTP.Operations, uint16(PANASONIC_GET_SETUP_INFO)) && s.Read.Status == "NOT_RUN"
}
func l10Actions(view string) []uiAction {
	s := l10Snapshot()
	idle := !s.Busy && !s.ExportBusy
	a := []uiAction{}
	add := func(id string, r Rect, on bool) { a = append(a, uiAction{id, r, on}) }
	if view == "l10report" {
		add("l10ReportPrevious", l10PreviewPrevRect, idle && s.PreviewPage > 0)
		pages := l10PreviewPages(s)
		add("l10ReportNext", l10PreviewNextRect, idle && s.PreviewPage+1 < len(pages))
		add("l10ReportBack", l10PreviewBackRect, idle)
		add("l10ReportSave", l10PreviewSaveRect, idle && s.Preview != nil)

		return a
	}
	add("l10Firmware", l10FirmwareRect, idle)
	add("l10Fixture", l10ModeRect, idle)
	add("l10A", l10ARect, idle && l10VersionPattern.MatchString(s.Firmware))
	add("l10Baseline", l10BaselineRect, idle)
	for i := 0; i < 3; i++ {
		index := s.Page*3 + i
		if index < len(s.Candidates) {
			add(fmt.Sprintf("l10Candidate%d", i), l10RowRect(i), idle && l10AllowedCandidate(s.Candidates[index]))
		}
	}
	add("l10Previous", l10PrevRect, idle && s.Page > 0)
	add("l10Next", l10NextRect, idle && (s.Page+1)*3 < len(s.Candidates))
	add("l10B", l10BRect, l10CanB(s))
	add("l10C", l10CRect, l10CanC(s))
	add("l10Preview", l10PreviewRect, idle)
	add("l10Close", l10CloseRect, !s.ExportBusy)
	return a
}
func l10StatusText(status string) string {
	ko := map[string]string{"NOT_RUN": "미실행", "PASS": "확인됨", "DEVICE_NOT_FOUND": "장치 없음", "UNSUPPORTED": "미지원", "INACCESSIBLE": "접근 불가", "TIMEOUT": "시간 초과", "CANCELLED": "중단됨", "ERROR": "오류"}
	if isKoreanUI() {
		if t, ok := ko[status]; ok {
			return t
		}
	}
	return status
}

// Notices are retained across language changes. Translate either stored language
// at paint time so the selected language always controls the current screen.
func l10NoticeText(text string) string {
	pairs := [][2]string{
		{"Checking… Cancel stops the worker.", "확인 중… 중단 버튼으로 워커를 종료할 수 있습니다."},
		{"No candidate. Check power, data cable and LUMIX Lab mode; you can export this result.", "장치 후보가 없습니다. 전원·데이터 케이블·LUMIX Lab 모드를 확인하세요. 이 결과도 저장할 수 있습니다."},
		{"Baseline saved in memory. Connect the L10, choose LUMIX Lab, then run step 1.", "연결 전 목록을 메모리에 기록했습니다. L10 연결 후 LUMIX Lab을 선택하고 1번을 실행하세요."},
		{"Select your L10 candidate. USB discovery alone does not prove Lab/PTP compatibility.", "본인의 L10 장치 후보를 선택하세요. USB 발견만으로 Lab·PTP 호환성이 확인되지는 않습니다."},
		{"DC-L10 answered, but device identity is unavailable. Step 3 stays locked.", "DC-L10은 응답했지만 기기 식별 정보를 얻지 못했습니다. 3번은 잠금 상태입니다."},
		{"DC-L10 confirmed; the existing read operation is advertised. Step 3 still needs consent.", "DC-L10과 기존 읽기 명령 지원 목록을 확인했습니다. 3번은 별도 동의 후 실행합니다."},
		{"PTP answered, but the existing read operation was not advertised. Step 3 stays locked.", "PTP는 응답했지만 기존 읽기 명령이 지원 목록에 없습니다. 3번은 잠금 상태입니다."},
		{"Raw counters received; all meanings remain UNVERIFIED. Review the report.", "원시 카운터를 받았습니다. 모든 의미는 미검증입니다. 보고서에서 확인하세요."},
		{"Target selected by you. Windows has not confirmed LUMIX Lab mode.", "대상을 선택했습니다. LUMIX Lab 모드는 Windows에서 확인한 값이 아닙니다."},
		{"Choose a location for both reports.", "두 보고서를 저장할 위치를 선택하세요."},
		{"Could not save reports. Check the destination and permissions.", "보고서를 저장할 수 없습니다. 저장 위치와 권한을 확인하세요."},
		{"Report save cancelled.", "보고서 저장을 취소했습니다."},
		{"JSON + TXT saved: ", "JSON + TXT 저장됨: "},
	}
	for _, pair := range pairs {
		for _, prefix := range pair {
			if strings.HasPrefix(text, prefix) {
				return uiText(pair[0], pair[1]) + strings.TrimPrefix(text, prefix)
			}
		}
	}
	for _, prefix := range []string{"Test stopped: ", "테스트 중단: "} {
		if strings.HasPrefix(text, prefix) {
			parts := strings.SplitN(strings.TrimPrefix(text, prefix), " · ", 2)
			status := parts[0]
			for code, ko := range map[string]string{"NOT_RUN": "미실행", "PASS": "확인됨", "DEVICE_NOT_FOUND": "장치 없음", "UNSUPPORTED": "미지원", "INACCESSIBLE": "접근 불가", "TIMEOUT": "시간 초과", "CANCELLED": "중단됨", "ERROR": "오류"} {
				if status == ko {
					status = code
					break
				}
			}
			out := uiText("Test stopped: ", "테스트 중단: ") + l10StatusText(status)
			if len(parts) > 1 {
				out += " · " + parts[1]
			}
			return out
		}
	}
	return text
}
func l10ResetRead(s *l10Session) {
	s.PTP = l10PTPResult{Status: "NOT_RUN", Operations: []uint16{}}
	s.Read = l10ReadResult{Status: "NOT_RUN"}
	s.Identity = ""
	s.BConsent, s.CConsent = false, false
	s.Preview = nil
}
func l10Start(stage string) {
	stateMu.Lock()
	if l10State.Busy || l10State.ExportBusy || probeBusy {
		stateMu.Unlock()
		return
	}
	req := l10WorkerRequest{Stage: stage}
	if stage == "B" && !l10CanB(l10State) || stage == "C" && !l10CanC(l10State) {
		stateMu.Unlock()
		return
	}
	if stage == "B" || stage == "C" {
		req.Target = l10State.Candidates[l10State.Selected]
		req.Consent = true
		req.Identity = l10State.Identity
	}
	ctx, cancel := context.WithTimeout(context.Background(), workerTimeout)
	l10State.Busy = true
	l10State.Running = stage
	l10State.Cancel = cancel
	l10State.Preview = nil
	l10State.Notice = uiText("Checking… Cancel stops the worker.", "확인 중… 중단 버튼으로 워커를 종료할 수 있습니다.")
	fixture := l10State.Fixture
	if stage == "A" {
		l10ResetRead(&l10State)
		l10State.Candidates = nil
		l10State.Selected = -1
		l10State.Page = 0
		l10State.USBStatus = "NOT_RUN"
	}
	if stage == "B" {
		l10State.BConsent = true
	}
	if stage == "C" {
		l10State.CConsent = true
	}
	stateMu.Unlock()
	procInvalidateRect.Call(hwndMain, 0, 0)
	go func() {
		defer cancel()
		var result l10WorkerReply
		if fixture {
			result = l10Fixture(stage)
		} else {
			result = l10Worker(ctx, req)
		}
		stateMu.Lock()
		l10State.Busy = false
		l10State.Running = ""
		l10State.Cancel = nil
		switch stage {
		case "baseline":
			l10State.BaselineStatus = result.Status
			l10State.Baseline = result.Candidates
		case "A":
			l10State.USBStatus, l10State.USBReason = result.Status, result.Reason
			l10State.Candidates = result.Candidates
			if l10State.BaselineStatus == "PASS" || l10State.BaselineStatus == "DEVICE_NOT_FOUND" {
				seen := map[string]bool{}
				for _, c := range l10State.Baseline {
					seen[c.USBID] = true
				}
				for i := range l10State.Candidates {
					added := !seen[l10State.Candidates[i].USBID]
					l10State.Candidates[i].USB.NewSinceBaseline = &added
				}
			}
		case "B":
			l10State.PTP = result.PTP
			l10State.Identity = result.Identity
			if l10State.PTP.Status == "" || l10State.PTP.Status == "NOT_RUN" {
				l10State.PTP.Status, l10State.PTP.Reason = result.Status, result.Reason
			}
			l10State.PTP.Uncertain = result.Reason == "WORKER_TIMEOUT" || result.Reason == "USER_CANCELLED" || result.Reason == "WORKER_FAILED"
		case "C":
			l10State.Read = result.Read
			if l10State.Read.Status == "" || l10State.Read.Status == "NOT_RUN" {
				l10State.Read.Status, l10State.Read.Reason = result.Status, result.Reason
			}
			l10State.Read.Uncertain = result.Reason == "WORKER_TIMEOUT" || result.Reason == "USER_CANCELLED" || result.Reason == "WORKER_FAILED"
		}
		l10State.Notice = l10ResultNotice(stage, result)
		stateMu.Unlock()
		procPostMessageW.Call(hwndMain, WM_APP_RESULT, 0, 0)
	}()
}
func l10ResultNotice(stage string, r l10WorkerReply) string {
	if r.Status == "DEVICE_NOT_FOUND" {
		return uiText("No candidate. Check power, data cable and LUMIX Lab mode; you can export this result.", "장치 후보가 없습니다. 전원·데이터 케이블·LUMIX Lab 모드를 확인하세요. 이 결과도 저장할 수 있습니다.")
	}
	if r.Status != "PASS" {
		return uiText("Test stopped: ", "테스트 중단: ") + l10StatusText(r.Status) + " · " + r.Reason
	}
	switch stage {
	case "baseline":
		return uiText("Baseline saved in memory. Connect the L10, choose LUMIX Lab, then run step 1.", "연결 전 목록을 메모리에 기록했습니다. L10 연결 후 LUMIX Lab을 선택하고 1번을 실행하세요.")
	case "A":
		return uiText("Select your L10 candidate. USB discovery alone does not prove Lab/PTP compatibility.", "본인의 L10 장치 후보를 선택하세요. USB 발견만으로 Lab·PTP 호환성이 확인되지는 않습니다.")
	case "B":
		if r.Identity == "" {
			return uiText("DC-L10 answered, but device identity is unavailable. Step 3 stays locked.", "DC-L10은 응답했지만 기기 식별 정보를 얻지 못했습니다. 3번은 잠금 상태입니다.")
		}
		if l10HasOperation(r.PTP.Operations, 0x9414) {
			return uiText("DC-L10 confirmed; the existing read operation is advertised. Step 3 still needs consent.", "DC-L10과 기존 읽기 명령 지원 목록을 확인했습니다. 3번은 별도 동의 후 실행합니다.")
		}
		return uiText("PTP answered, but the existing read operation was not advertised. Step 3 stays locked.", "PTP는 응답했지만 기존 읽기 명령이 지원 목록에 없습니다. 3번은 잠금 상태입니다.")
	default:
		return uiText("Raw counters received; all meanings remain UNVERIFIED. Review the report.", "원시 카운터를 받았습니다. 모든 의미는 미검증입니다. 보고서에서 확인하세요.")
	}
}
func l10Confirm(stage string) bool {
	text := uiText("Confirm that the selected USB device is your DC-L10. Read standard camera information once? Compatibility is unverified. No service counter request is sent in this step.", "선택한 USB 장치가 본인의 DC-L10인지 확인해 주세요. 표준 카메라 정보를 1회 읽을까요? 호환성은 미검증이며 이 단계에서는 서비스 카운터를 조회하지 않습니다.")
	if stage == "C" {
		text = uiText("Reconfirm DC-L10 identity and advertised operations, then try the existing Panasonic read once? This model is unverified. Raw values will not be labelled as shutter or power counts.", "DC-L10 식별과 지원 명령을 다시 확인한 뒤 기존 Panasonic 읽기를 1회 시도할까요? 이 기종은 미검증입니다. 반환값을 셔터·전원 횟수로 표시하지 않습니다.")
	}
	answer, _, _ := procMessageBoxW.Call(hwndMain, uintptr(unsafe.Pointer(wstr(text))), uintptr(unsafe.Pointer(wstr(uiText("L10 read-only test — consent", "L10 읽기 전용 테스트 · 동의")))), 0x4|0x100|0x30)
	return answer == 6
}
func l10Click(x, y int32) bool {
	stateMu.Lock()
	view := currentView
	stateMu.Unlock()
	if !l10View(view) {
		return false
	}
	a, ok := actionAt(x, y)
	if !ok || !a.Enabled {
		return true
	}
	switch a.ID {
	case "l10Firmware":
		focusAction = a.ID
		keyboardFocus = true
	case "l10Fixture":
		stateMu.Lock()
		fixture := !l10State.Fixture
		fw := l10State.Firmware
		l10State = l10NewSession()
		l10State.Fixture = fixture
		l10State.Firmware = fw
		if fixture {
			l10State.Firmware = "1.2"
		}
		stateMu.Unlock()
	case "l10Baseline":
		l10Start("baseline")
	case "l10A":
		l10Start("A")
	case "l10B":
		if l10Confirm("B") {
			l10Start("B")
		}
	case "l10C":
		if l10Confirm("C") {
			l10Start("C")
		}
	case "l10Previous", "l10Next":
		stateMu.Lock()
		if a.ID == "l10Previous" {
			l10State.Page--
		} else {
			l10State.Page++
		}
		stateMu.Unlock()
	case "l10Candidate0", "l10Candidate1", "l10Candidate2":
		row := int(a.ID[len(a.ID)-1] - '0')
		stateMu.Lock()
		l10State.Selected = l10State.Page*3 + row
		l10ResetRead(&l10State)
		l10State.Notice = uiText("Target selected by you. Windows has not confirmed LUMIX Lab mode.", "대상을 선택했습니다. LUMIX Lab 모드는 Windows에서 확인한 값이 아닙니다.")
		stateMu.Unlock()
	case "l10Preview":
		stateMu.Lock()
		report := l10BuildReport(l10State)
		l10State.Preview = &report
		l10State.PreviewPage = 0
		currentView = "l10report"
		stateMu.Unlock()
	case "l10ReportBack":
		stateMu.Lock()
		currentView = "l10"
		stateMu.Unlock()
	case "l10ReportPrevious", "l10ReportNext":
		stateMu.Lock()
		if a.ID == "l10ReportPrevious" {
			l10State.PreviewPage--
		} else {
			l10State.PreviewPage++
		}
		stateMu.Unlock()
	case "l10ReportSave":
		l10Export()
	case "l10Close":
		stateMu.Lock()
		cancel := l10State.Cancel
		if !l10State.Busy {
			currentView = "validation"
		}
		stateMu.Unlock()
		if cancel != nil {
			cancel()
		}
	}
	procInvalidateRect.Call(hwndMain, 0, 0)
	return true
}
func l10Input(char uint32) bool {
	stateMu.Lock()
	defer stateMu.Unlock()
	if currentView != "l10" || focusAction != "l10Firmware" || l10State.Busy || l10State.ExportBusy {
		return false
	}
	if char == 8 {
		if len(l10State.Firmware) > 0 {
			l10State.Firmware = l10State.Firmware[:len(l10State.Firmware)-1]
		}
	} else if (char >= '0' && char <= '9' || char == '.') && len(l10State.Firmware) < 16 {
		l10State.Firmware += string(rune(char))
	} else {
		return true
	}
	// A changed firmware entry cannot inherit B/C gates.
	l10ResetRead(&l10State)
	return true
}
func paintL10(hdc uintptr) {
	s := l10Snapshot()
	fillRound(hdc, Rect{40, 180, 940, 655}, rgb(28, 28, 31), 18)
	setFont(hdc, fontSection)
	drawText(hdc, uiText("L10 USB CONNECTION TEST · BETA", "L10 USB 연결 테스트 · BETA"), Rect{70, 199, 910, 234}, DT_VCENTER|DT_SINGLELINE, rgb(242, 242, 245))
	setFont(hdc, fontSmall)
	drawText(hdc, uiText("DC-L10 only · Select LUMIX Lab on the camera and connect a USB data cable to Windows.", "DC-L10 전용 · 카메라에서 LUMIX Lab을 선택하고 USB 데이터 케이블로 Windows에 연결하세요."), Rect{70, 239, 910, 263}, DT_VCENTER|DT_SINGLELINE, rgb(185, 198, 216))
	drawText(hdc, uiText("Experimental and unverified. USB discovery ≠ Lab communication ≠ verified shutter count.", "실험 기능 · 호환성 미검증. USB 발견 ≠ Lab 통신 성공 ≠ 셔터 횟수 검증."), Rect{70, 264, 910, 285}, DT_VCENTER|DT_SINGLELINE, rgb(241, 163, 169))
	drawText(hdc, uiText("Firmware on camera:", "본체에서 확인한 펌웨어:"), Rect{70, 286, 263, 315}, DT_VCENTER|DT_SINGLELINE, rgb(198, 198, 208))
	fw := s.Firmware
	if fw == "" {
		fw = uiText("Type e.g. 1.2", "입력: 예) 1.2")
	}
	if focusAction == "l10Firmware" && keyboardFocus {
		fw += " |"
	}
	drawOutlineButton(hdc, l10FirmwareRect, fw)
	setFont(hdc, fontTiny)
	drawText(hdc, uiText("Mode is user-reported", "모드는 사용자 선택값"), Rect{435, 287, 694, 315}, DT_VCENTER|DT_SINGLELINE, rgb(150, 150, 160))
	modeLabel := uiText("REAL USB · SWITCH", "실제 USB · 전환")
	if s.Fixture {
		modeLabel = uiText("SAMPLE · SWITCH", "가상 샘플 · 전환")
	}
	drawChoiceButton(hdc, l10ModeRect, modeLabel, s.Fixture)
	drawButton(hdc, l10ARect, uiText("1. CHECK USB DEVICE", "1. USB 장치 확인"))
	drawOutlineButton(hdc, l10BaselineRect, uiText("BEFORE CONNECTION (OPTIONAL)", "연결 전 목록 기록 (선택)"))
	for i := 0; i < 3; i++ {
		index := s.Page*3 + i
		if index >= len(s.Candidates) {
			continue
		}
		c := s.Candidates[index]
		label := fmt.Sprintf("%02d  %s   VID %s / PID %s", index+1, c.USB.Name, c.USB.VID, c.USB.PID)
		if c.USB.WPDExposed {
			label += " · WPD"
		} else {
			label += uiText(" · no WPD path", " · WPD 경로 없음")
		}
		if !l10AllowedCandidate(c) {
			label += uiText(" · not L10", " · L10 아님")
		}
		drawChoiceButton(hdc, l10RowRect(i), label, index == s.Selected)
	}
	if len(s.Candidates) == 0 {
		setFont(hdc, fontSmall)
		placeholder := uiText("Enter firmware, then run step 1. A negative result can also be exported.", "펌웨어를 입력하고 1번을 실행하세요. 장치가 발견되지 않은 결과도 저장할 수 있습니다.")
		if s.Busy {
			placeholder = uiText("Checking USB device metadata…", "USB 장치 정보를 확인하고 있습니다…")
		} else if s.USBStatus != "NOT_RUN" {
			placeholder = uiText("No candidate available. Review and export the result below.", "선택할 장치 후보가 없습니다. 아래 보고서 미리보기에서 결과를 저장할 수 있습니다.")
		}
		drawText(hdc, placeholder, Rect{70, 390, 910, 453}, DT_CENTER|0x10, rgb(150, 157, 169))
	}
	drawOutlineButton(hdc, l10PrevRect, uiText("PREVIOUS", "이전"))
	drawOutlineButton(hdc, l10NextRect, uiText("NEXT", "다음"))
	setFont(hdc, fontSmall)
	drawText(hdc, fmt.Sprintf("A: %s     B: %s     C: %s", l10StatusText(s.USBStatus), l10StatusText(s.PTP.Status), l10StatusText(s.Read.Status)), Rect{185, 486, 795, 514}, DT_VCENTER|DT_CENTER|DT_SINGLELINE, rgb(167, 200, 240))
	drawOutlineButton(hdc, l10BRect, uiText("2. STANDARD PTP (CONSENT)", "2. 표준 PTP 확인 (동의)"))
	drawOutlineButton(hdc, l10CRect, uiText("3. EXISTING READ (CONSENT)", "3. 기존 읽기 조회 (동의)"))
	notice := l10NoticeText(s.Notice)
	if notice == "" {
		notice = uiText("Step 2 needs an OS WPD path. Step 3 needs confirmed DC-L10 and an advertised read operation.", "2번은 WPD 경로가 필요합니다. 3번은 DC-L10 식별과 지원 명령 확인 후 활성화됩니다.")
	}
	if s.Fixture {
		notice = uiText("SYNTHETIC SAMPLE — NO USB ACCESS. ", "가상 샘플 · 실제 USB 접근 없음. ") + notice
	}
	setFont(hdc, fontSmall)
	drawText(hdc, notice, Rect{70, 580, 910, 642}, DT_LEFT|0x10|0x8000, rgb(185, 198, 216))
	drawOutlineButton(hdc, l10PreviewRect, uiText("REVIEW REDACTED REPORT", "개인정보 제거 보고서 미리보기"))
	close := uiText("CLOSE TEST", "테스트 닫기")
	if s.Busy {
		close = uiText("CANCEL WORKER", "워커 중단")
	}
	drawOutlineButton(hdc, l10CloseRect, close)
	drawFooter(hdc, UsageResult{}, false)
}
func l10PreviewPages(s l10Session) [][]string {
	if s.Preview == nil {
		return [][]string{{"NOT RUN"}}
	}
	var lines []string
	for _, line := range strings.Split(l10ReportText(*s.Preview), "\n") {
		// Technical report text uses ASCII fields; wrapping preserves exact content.
		runes := []rune(line)
		for len(runes) > 100 {
			lines = append(lines, string(runes[:100]))
			runes = runes[100:]
		}
		lines = append(lines, string(runes))
	}
	var pages [][]string
	for i := 0; i < len(lines); i += 17 {
		end := i + 17
		if end > len(lines) {
			end = len(lines)
		}
		pages = append(pages, lines[i:end])
	}
	return pages
}
func paintL10Report(hdc uintptr) {
	s := l10Snapshot()
	pages := l10PreviewPages(s)
	page := s.PreviewPage
	if page >= len(pages) {
		page = len(pages) - 1
	}
	if page < 0 {
		page = 0
	}
	fillRound(hdc, Rect{40, 180, 940, 655}, rgb(28, 28, 31), 18)
	setFont(hdc, fontSection)
	drawText(hdc, uiText("REDACTED REPORT PREVIEW", "개인정보 제거 보고서 미리보기"), Rect{70, 199, 910, 234}, DT_VCENTER|DT_SINGLELINE, rgb(242, 242, 245))
	setFont(hdc, fontTiny)
	drawText(hdc, uiText("TXT preview · JSON contains the same facts · No identifiers or full RAW reply", "TXT 미리보기 · JSON에도 동일한 정보 저장 · 고유 식별자와 전체 RAW 응답 제외"), Rect{70, 238, 910, 260}, DT_VCENTER|DT_SINGLELINE, rgb(167, 200, 240))
	setFont(hdc, fontSmall)
	for i, line := range pages[page] {
		top := int32(268 + i*19)
		drawText(hdc, line, Rect{70, top, 910, top + 20}, DT_SINGLELINE|0x8000, rgb(205, 211, 220))
	}
	drawOutlineButton(hdc, l10PreviewPrevRect, uiText("PREVIOUS", "이전"))
	drawOutlineButton(hdc, l10PreviewNextRect, uiText("NEXT", "다음"))
	setFont(hdc, fontSmall)
	drawText(hdc, fmt.Sprintf("%d / %d", page+1, len(pages)), Rect{400, 607, 580, 640}, DT_VCENTER|DT_CENTER|DT_SINGLELINE, rgb(167, 200, 240))
	drawOutlineButton(hdc, l10PreviewBackRect, uiText("BACK TO TEST", "테스트로 돌아가기"))
	label := uiText("SAVE JSON + TXT", "JSON + TXT 저장")
	if s.ExportBusy {
		label = uiText("SAVING…", "저장 중…")
	}
	drawButton(hdc, l10PreviewSaveRect, label)
	setFont(hdc, fontSmall)
	drawText(hdc, l10NoticeText(s.Notice), Rect{70, 647, 910, 674}, DT_CENTER|DT_SINGLELINE|0x8000, rgb(167, 200, 240))
	drawFooter(hdc, UsageResult{}, false)
}
func l10SaveDialog() (string, bool, error) {
	initial := syscall.StringToUTF16(preferredReportDir())
	name := "L10_USB_Test_" + time.Now().Format("20060102_150405") + ".json"
	file := make([]uint16, 32768)
	copy(file, syscall.StringToUTF16(name))
	filter := utf16MultiString(uiText("JSON report (*.json)", "JSON 보고서 (*.json)"), "*.json")
	title := syscall.StringToUTF16(uiText("Save L10 JSON + TXT reports", "L10 JSON + TXT 보고서 저장"))
	ext := syscall.StringToUTF16("json")
	ofn := OPENFILENAMEW{LStructSize: uint32(unsafe.Sizeof(OPENFILENAMEW{})), HwndOwner: hwndMain, LpstrFilter: &filter[0], LpstrFile: &file[0], NMaxFile: uint32(len(file)), LpstrInitialDir: &initial[0], LpstrTitle: &title[0], LpstrDefExt: &ext[0], Flags: OFN_OVERWRITEPROMPT | OFN_NOCHANGEDIR | OFN_PATHMUSTEXIST}
	ok, _, _ := procGetSaveFileNameW.Call(uintptr(unsafe.Pointer(&ofn)))
	if ok == 0 {
		code, _, _ := procCommDlgExtendedError.Call()
		if code != 0 {
			return "", false, fmt.Errorf("dialog error %X", code)
		}
		return "", false, nil
	}
	path := syscall.UTF16ToString(file)
	if !strings.EqualFold(filepath.Ext(path), ".json") {
		path += ".json"
	}
	txt := strings.TrimSuffix(path, filepath.Ext(path)) + ".txt"
	if _, err := os.Stat(txt); err == nil {
		msg := uiText("A matching TXT report already exists. Replace both JSON and TXT?", "같은 이름의 TXT가 이미 있습니다. JSON과 TXT를 모두 덮어쓸까요?")
		answer, _, _ := procMessageBoxW.Call(hwndMain, uintptr(unsafe.Pointer(wstr(msg))), uintptr(unsafe.Pointer(wstr(uiText("Replace report pair", "보고서 덮어쓰기")))), 0x4|0x100|0x30)
		if answer != 6 {
			return "", false, nil
		}
	}
	return path, true, nil
}
func l10Export() {
	stateMu.Lock()
	if l10State.Busy || l10State.ExportBusy || l10State.Preview == nil {
		stateMu.Unlock()
		return
	}
	report := *l10State.Preview
	l10State.ExportBusy = true
	l10State.Notice = uiText("Choose a location for both reports.", "두 보고서를 저장할 위치를 선택하세요.")
	stateMu.Unlock()
	procInvalidateRect.Call(hwndMain, 0, 0)
	go func() {
		runtime.LockOSThread()
		path, ok, err := l10SaveDialog()
		runtime.UnlockOSThread()
		if err == nil && ok {
			err = l10SavePair(path, report)
		}
		stateMu.Lock()
		l10State.ExportBusy = false
		if err != nil {
			l10State.Notice = uiText("Could not save reports. Check the destination and permissions.", "보고서를 저장할 수 없습니다. 저장 위치와 권한을 확인하세요.")
		} else if !ok {
			l10State.Notice = uiText("Report save cancelled.", "보고서 저장을 취소했습니다.")
		} else {
			l10State.Notice = uiText("JSON + TXT saved: ", "JSON + TXT 저장됨: ") + filepath.Base(path)
		}
		stateMu.Unlock()
		procPostMessageW.Call(hwndMain, WM_APP_RESULT, 0, 0)
	}()
}
