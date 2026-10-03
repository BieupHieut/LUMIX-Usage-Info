//go:build windows

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

// Windows UI language is the default. A manual selection is stored per user.
const (
	languageAuto uint32 = iota
	languageKorean
	languageEnglish
)

var windowsKoreanUI uint32
var languageMode uint32

func isKoreanUI() bool {
	switch atomic.LoadUint32(&languageMode) {
	case languageKorean:
		return true
	case languageEnglish:
		return false
	default:
		return atomic.LoadUint32(&windowsKoreanUI) != 0
	}
}

func languageSettingPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "Lumerian Usage Info", "language.txt")
}

func loadLanguagePreference() {
	path := languageSettingPath()
	if path == "" {
		return
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	mode, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err == nil && mode >= int(languageAuto) && mode <= int(languageEnglish) {
		atomic.StoreUint32(&languageMode, uint32(mode))
	}
}

func setLanguagePreference(mode uint32) error {
	if mode > languageEnglish {
		return fmt.Errorf("invalid language mode")
	}
	atomic.StoreUint32(&languageMode, mode)
	path := languageSettingPath()
	if path == "" {
		return fmt.Errorf("configuration folder unavailable")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(strconv.Itoa(int(mode))), 0600)
}

func languageButtonLabel() string {
	switch atomic.LoadUint32(&languageMode) {
	case languageKorean:
		return "한국어  ▾"
	case languageEnglish:
		return "English  ▾"
	default:
		return uiText("English · Auto  ▾", "한국어 · 자동  ▾")
	}
}

func isKoreanLanguageID(id uint16) bool { return id&0x03ff == 0x12 }

func refreshUILanguage() bool {
	wasKorean := isKoreanUI()
	value, _, _ := procGetUserDefaultUILanguage.Call()
	var next uint32
	if isKoreanLanguageID(uint16(value)) {
		next = 1
	}
	atomic.StoreUint32(&windowsKoreanUI, next)
	return wasKorean != isKoreanUI()
}

func uiText(english, korean string) string {
	if isKoreanUI() {
		return korean
	}
	return english
}

var koreanText = map[string]string{
	"Camera response timed out. Close other camera apps and check USB mode: PC(Tether), or LUMIX Lab for DC-L10.": "카메라 응답 시간이 초과됐습니다. 다른 카메라 앱을 종료하고 USB 모드(DC-L10: LUMIX Lab)를 확인하세요.",
	"No LUMIX camera found. Connect USB and select PC(Tether), or LUMIX Lab for DC-L10.":                          "카메라를 찾지 못했습니다. USB 연결 후 PC(테더), DC-L10은 LUMIX Lab을 선택하세요.",
	"READ REPORTED · COUNTERS UNVERIFIED":                                             "읽기 성공 제보 · 카운터 미검증",
	"L10 standard device information could not be confirmed.":                         "L10의 표준 기기 정보를 확인하지 못했습니다.",
	"UI LAYOUT & TETHER GUIDE":                                                        "화면 배치 및 테더 안내",
	"Fixed text bounds, stable header, character variants and USB tether model list.": "글자 영역·헤더 배치 수정, 캐릭터 변형 및 USB 테더 기종 안내.",
	"Cameras without USB tether / PC(Tether) mode are not supported.":                 "USB 테더 / PC(테더) 기능이 없는 카메라는 지원하지 않습니다.",
	"CAMERA LIST & VALIDATION":                                                        "카메라 목록과 검증",
	"Verified counters and USB tether models":                                         "검증된 횟수와 USB 테더 지원 기종",
	"USB TETHER MODELS":                                                               "USB 테더 지원 기종",
	"Counter verification: the three rows above only":                                 "횟수 검증: 위의 세 조합에만 적용",
	"S SERIES":            "S 시리즈",
	"G / GH / BGH SERIES": "G / GH / BGH 시리즈",
	"Tether support does not verify this tool's readings or counter meanings.": "테더 지원은 이 도구의 읽기 성공이나 카운터 의미 검증을 뜻하지 않습니다.",
	"LUMIX Camera Usage Reader":                     "LUMIX 카메라 사용 정보",
	"●  USB CONNECTION GUIDE":                       "●  USB 연결 안내",
	"Connect your camera":                           "카메라 연결하기",
	"Turn on the camera.":                           "카메라 전원을 켜세요.",
	"Connect the camera to this computer with USB.": "USB 케이블로 카메라와 컴퓨터를 연결하세요.",
	"On the camera, choose USB Mode → PC(Tether).":  "카메라에서 USB 모드 → PC(테더)를 선택하세요.",
	"Waiting for a camera":                          "카메라 연결을 기다리는 중",
	"Checking camera connection...":                 "카메라 연결 확인 중...",
	"No LUMIX camera found. Turn on the camera, connect USB, and select PC(Tether).": "카메라를 찾지 못했습니다. 전원과 USB 연결, PC(테더) 선택을 확인하세요.",
	"No readable LUMIX camera found.":                                                "읽을 수 있는 LUMIX 카메라를 찾지 못했습니다.",
	"Camera disconnected · reconnect USB":                                            "카메라 연결이 끊어졌습니다 · USB를 다시 연결하세요",
	"CAMERA DATA":                                                                    "카메라 정보",
	"Model, firmware and serial":                                                     "모델 · 펌웨어 · 시리얼",
	"USAGE COUNTERS":                                                                 "사용 횟수",
	"Shutter and power / wake":                                                       "셔터 · 전원/깨우기",
	"SAVE & SHARE":                                                                   "저장 및 공유",
	"PNG and text reports":                                                           "PNG · 텍스트 리포트",
	"CONNECT CAMERA":                                                                 "카메라 연결",
	"CONNECTING...":                                                                  "연결 중...",
	"VALIDATION & INFO":                                                              "검증 및 정보",
	"Camera menu may vary: Setup → IN/OUT → USB → USB Mode → PC(Tether)": "카메라별 메뉴가 다를 수 있습니다: 설정 → 입출력 → USB → USB 모드 → PC(테더)",
	"Close LUMIX Tether on this computer before connecting.":             "연결 전 컴퓨터의 LUMIX Tether 앱을 종료하세요.",
	"●  CAMERA CONNECTED":                  "●  카메라 연결됨",
	"●  DISCONNECTED · LAST READ":          "●  연결 끊김 · 마지막 읽기값",
	"●  CAMERA READ · COUNTERS UNVERIFIED": "●  카메라 읽음 · 카운터 미검증",
	"CAMERA MODEL":                         "카메라 모델",
	"FIRMWARE":                             "펌웨어",
	"SERIAL NUMBER":                        "시리얼 번호",
	"HIDE":                                 "숨기기",
	"SHOW":                                 "보기",
	"SHUTTER ACTUATIONS":                   "셔터 작동 횟수",
	"POWER / WAKE ACTIVATIONS":             "전원 / 깨우기 횟수",
	"UNVERIFIED FOR THIS MODEL / FIRMWARE": "이 모델 / 펌웨어에서 미검증",
	"WHAT COUNTS?":                         "횟수 기준",
	"REFRESH CAMERA":                       "카메라 새로고침",
	"REFRESHING...":                        "새로고침 중...",
	"Read current camera data":             "카메라 정보 다시 읽기",
	"ERROR HISTORY":                        "오류 기록",
	"TECHNICAL DETAILS":                    "기술 세부 정보",
	"Raw values and status":                "원본 값과 상태",
	"Supported camera pairs":               "검증된 기종 / 펌웨어",
	"SAVE PNG":                             "PNG 저장",
	"Privacy options and share":            "공개 범위 및 공유",
	"SAVE REPORT":                          "리포트 저장",
	"SAVING...":                            "저장 중...",
	"Text report":                          "텍스트 리포트",
	"USB is disconnected. Last successfully read values are shown.": "USB 연결이 끊어졌습니다. 마지막 읽기값을 표시합니다.",
	"Refreshing camera data...":                                     "카메라 정보를 새로 읽는 중...",
	"Read-only service data · No camera data modified":              "읽기 전용 서비스 정보 · 카메라 데이터는 변경하지 않습니다",
	"UNKNOWN":                       "알 수 없음",
	"Power / Wake · VERIFIED":       "전원 / 깨우기 · 검증됨",
	"Shutter Actuations · VERIFIED": "셔터 작동 횟수 · 검증됨",
	"Power / Wake · NOT VERIFIED for this model / firmware":                                                 "전원 / 깨우기 · 이 모델/펌웨어에서 미검증",
	"Shutter Actuations · NOT VERIFIED for this model / firmware":                                           "셔터 작동 횟수 · 이 모델/펌웨어에서 미검증",
	"Power / Wake · NOT VERIFIED for this firmware":                                                         "전원 / 깨우기 · 이 펌웨어에서 미검증",
	"Shutter Actuations · NOT VERIFIED for this firmware":                                                   "셔터 작동 횟수 · 이 펌웨어에서 미검증",
	"Counters 3–7 remain unidentified; legacy candidate names are unverified reference hints.":              "카운터 3~7의 의미는 확인되지 않았습니다. 기존 명칭은 검증되지 않은 참고 정보입니다.",
	"Usage names are provisional for this model / firmware; raw Counter 1/2 values remain available above.": "이 모델/펌웨어에서는 사용 횟수 명칭이 잠정적입니다. 위에서 원본 카운터 1/2 값을 확인할 수 있습니다.",
	"BACK":                                "뒤로",
	"Observed on DC-S1RM2 · Firmware 1.5": "DC-S1RM2 · 펌웨어 1.5에서 관찰",
	"EXPORT / SHARE":                      "내보내기 / 공유",
	"Create a clean PNG or copy a text summary": "PNG를 저장하거나 텍스트 요약을 복사합니다",
	"1. CHOOSE PURPOSE":                         "1. 용도 선택",
	"PUBLIC SHARE":                              "공개 공유",
	"DEVICE VERIFICATION":                       "기기 검증",
	"Public Share masks the serial. Device Verification starts with Last 4 digits; Full Serial is optional.": "공개 공유는 시리얼을 숨깁니다. 기기 검증은 뒤 4자리로 시작하며 전체 공개도 선택할 수 있습니다.",
	"2. SERIAL NUMBER IN EXPORT": "2. 내보낼 시리얼 번호",
	"MASKED":                     "숨김",
	"LAST 4 DIGITS":              "뒤 4자리",
	"FULL SERIAL":                "전체 시리얼",
	"Full serial number will be included. Avoid posting the exported image publicly.": "전체 시리얼이 포함됩니다. 이 이미지를 공개 게시하지 마세요.",
	"WORKING...":   "작업 중...",
	"COPY SUMMARY": "요약 복사",
	"PNG files are saved to Documents\\LUMIX Usage Info Reports when available.": "PNG는 가능하면 문서\\LUMIX Usage Info Reports 폴더에 저장됩니다.",
	"VALIDATION & PROJECT INFO":                                 "검증 및 프로젝트 정보",
	"Hardware validation · Compatibility · Project information": "하드웨어 검증 · 호환성 · 프로젝트 정보",
	"VERIFIED COMBINATION":                                      "검증된 조합",
	"● HARDWARE VERIFIED":                                       "● 하드웨어 검증됨",
	"VALIDATION SCOPE":                                          "검증 범위",
	"EXPLORE PROJECT INFO":                                      "프로젝트 정보 보기",
	"CAMERA LIST":                                               "카메라 목록",
	"VERSION HISTORY":                                           "버전 기록",
	"DEVELOPER CREDITS":                                         "개발자 정보",
	"CAMERA VALIDATION MATRIX":                                  "카메라 검증 목록",
	"Verified hardware vs. models that still need real-camera testing": "검증된 기종과 실제 카메라 테스트가 필요한 기종",
	"● VERIFIED":                    "● 검증됨",
	"✓ VERIFIED":                    "✓ 검증됨",
	"Shutter ✓   Power / Wake ✓":    "셔터 ✓   전원/깨우기 ✓",
	"Counter meanings not verified": "카운터 의미 미검증",
	"Validated by ":                 "검증자 ",
	"Shutter Actuations ✓    Power / Wake Activations ✓":          "셔터 작동 횟수 ✓    전원 / 깨우기 횟수 ✓",
	"NEEDS HARDWARE VALIDATION":                                   "하드웨어 검증 필요",
	"Candidate list: real-camera field mapping not yet verified.": "후보 기종: 실제 카메라의 필드 의미는 아직 검증되지 않았습니다.",
	"NEEDS VALIDATION": "검증 필요",
	"Verification applies only to the listed model / firmware combinations.": "검증은 표시된 모델 / 펌웨어 조합에만 적용됩니다.",
	"Public beta evolution":                                                       "공개 베타 변경 내역",
	"CREATOR & CONTRIBUTORS":                                                      "제작자와 도움을 주신 분들",
	"Creator · Product design · Camera validation":                                "제작 · 제품 디자인 · 카메라 검증",
	"Feedback: GitHub Issues · Instagram @bieup_hieut":                            "문의: GitHub Issues · Instagram @bieup_hieut",
	"Unofficial read-only community tool · Not affiliated with Panasonic":         "비공식 읽기 전용 도구 · Panasonic과 무관합니다",
	"MANUAL LANGUAGE & UI POLISH":                                                 "언어 선택 및 화면 개선",
	"HOME TEXT CLIPPING FIX":                                                      "홈 카드 글자 잘림 수정",
	"Card titles and descriptions now fit the selected font height.":              "글꼴 높이에 맞춰 카드 제목과 설명 영역을 수정.",
	"Windows language default, Korean / English switch, clearer layout and copy.": "Windows 언어 기본값, 한·영 전환, 배치와 문구 개선.",
	"Unofficial community project":                                                "비공식 커뮤니티 프로젝트",
	"Lumerian · Good Photos, Better Days!":                                        "Lumerian · 좋은 사진, 더 좋은 하루!",
	"Created & hardware-validated by @bieup_hieut":                                "제작 및 하드웨어 검증 @bieup_hieut",
	"LUMIX Usage Info is an unofficial read-only community tool and is not affiliated with or endorsed by Panasonic.": "LUMIX Usage Info는 비공식 읽기 전용 커뮤니티 도구이며 Panasonic과 제휴하거나 후원받지 않습니다.",
	"Camera data is read only. No write / EEPROM / ROM / firmware-modification path is implemented.":                  "카메라 정보는 읽기만 합니다. 쓰기·EEPROM·ROM·펌웨어 변경 기능은 없습니다.",
	"ERROR HISTORY DETAILS":                                               "오류 기록 상세",
	"● FOCUSED VIEW · Error History opened":                               "● 오류 기록 상세 화면",
	"Up to 16 service-history records · legacy Panasonic reference codes": "최대 16개 서비스 기록 · 기존 Panasonic 참고 코드",
	"No recorded non-zero error codes found.":                             "0이 아닌 오류 코드가 기록되지 않았습니다.",
	"A slot may contain a date marker while the error code is 00000000; this app does not treat that as a recorded error.": "날짜가 있어도 오류 코드가 00000000이면 오류 기록으로 간주하지 않습니다.",
	"DATE / TIME": "날짜 / 시간",
	"CODE":        "코드",
	"DESCRIPTION": "설명",
	"PREVIOUS":    "이전",
	"NEXT":        "다음",
	"Read-only access · No camera data modified": "읽기 전용 · 카메라 데이터 변경 없음",
	"Made by @bieup_hieut · CREDITS":             "제작 @bieup_hieut · 정보",
	"Model":                                      "모델",
	"Firmware":                                   "펌웨어",
	"Serial Number":                              "시리얼 번호",
	"Shutter Actuations":                         "셔터 작동 횟수",
	"Power / Wake Activations":                   "전원 / 깨우기 횟수",
	"CAMERA INFORMATION":                         "카메라 정보",
	"Read directly from the connected camera · Read-only": "연결된 카메라에서 읽음 · 읽기 전용",
	"Not available":         "사용 불가",
	"Not supplied":          "미제공",
	"Not verified":          "미검증",
	"Serial number shown.":  "시리얼 번호를 표시합니다.",
	"Serial number masked.": "시리얼 번호를 숨깁니다.",
	"Camera disconnected or unavailable. Showing last known data.":         "카메라 연결이 끊겼거나 사용할 수 없습니다. 마지막 읽기값을 표시합니다.",
	"Camera not connected. Project information remains available offline.": "카메라가 연결되지 않았습니다. 프로젝트 정보는 오프라인에서도 볼 수 있습니다.",
	"Camera read completed.": "카메라 정보 읽기를 마쳤습니다.",
	"Saving report...":       "리포트 저장 중...",
	"USB device change detected. Checking camera connection...":                            "USB 기기 변경을 감지했습니다. 카메라 연결 확인 중...",
	"Developer Credits opened.":                                                            "개발자 정보를 열었습니다.",
	"Version History opened.":                                                              "버전 기록을 열었습니다.",
	"Validation & Info opened.":                                                            "검증 및 정보를 열었습니다.",
	"Camera Validation Matrix opened.":                                                     "카메라 검증 목록을 열었습니다.",
	"PNG operation in progress...":                                                         "PNG 작업 중...",
	"Public Share selected. Serial number is masked by default.":                           "공개 공유를 선택했습니다. 시리얼 번호는 기본으로 숨깁니다.",
	"Device Verification selected. Choose serial visibility, then SAVE PNG.":               "기기 검증을 선택했습니다. 시리얼 표시 방식을 고른 뒤 PNG를 저장하세요.",
	"Choose a save location...":                                                            "저장 위치를 선택하세요...",
	"PNG save cancelled.":                                                                  "PNG 저장을 취소했습니다.",
	"Saving PNG...":                                                                        "PNG 저장 중...",
	"Copying summary...":                                                                   "요약 복사 중...",
	"✓ Summary copied to clipboard.":                                                       "✓ 요약을 클립보드에 복사했습니다.",
	"Error History opened.":                                                                "오류 기록을 열었습니다.",
	"Reconnect and REFRESH before exporting a verification PNG.":                           "검증 PNG를 내보내려면 다시 연결하고 새로고침하세요.",
	"Choose export privacy, then click SAVE PNG.":                                          "공개 범위를 선택한 뒤 PNG 저장을 누르세요.",
	"Reconnect and REFRESH before saving a new report.":                                    "새 리포트를 저장하려면 다시 연결하고 새로고침하세요.",
	"✓ MECH exposure: +1 / physical actuation":                                             "✓ 기계식 셔터: 실제 작동당 +1",
	"✓ Power OFF + shutter CLOSE: +1":                                                      "✓ 전원 끔 + 셔터 닫기: +1",
	"✓ Electronic shutter exposure: +0":                                                    "✓ 전자식 셔터 촬영: +0",
	"✓ High Resolution test: +0 (tested setup)":                                            "✓ 고해상도 촬영 시험: +0",
	"• Sensor Cleaning + restart: +1 total":                                                "• 센서 청소 + 재시작: 총 +1",
	"! Pixel Refresh + restart: +3 total*":                                                 "! 픽셀 리프레시 + 재시작: 총 +3*",
	"✓ Camera OFF -> ON: +1":                                                               "✓ 카메라 전원 끔 → 켬: +1",
	"✓ Sleep -> wake cycle: +1":                                                            "✓ 절전 → 깨우기: +1",
	"✓ USB reconnect only: +0":                                                             "✓ USB 재연결만 수행: +0",
	"• Sensor Cleaning restart: +1":                                                        "• 센서 청소 재시작: +1",
	"• Pixel Refresh restart: +1":                                                          "• 픽셀 리프레시 재시작: +1",
	"* Sleep cycle verified; exact increment moment not isolated":                          "* 절전 주기는 검증, 정확한 증가 시점은 미확인",
	"* Sleep cycle verified; increment timing unknown":                                     "* 절전 주기 검증 · 증가 시점은 미확인",
	"Pixel Refresh: +3 shutter counts observed with Power-off Shutter=CLOSE.":              "픽셀 리프레시: 전원 끔 셔터=닫기에서 셔터 +3 관찰",
	"Exact cause unknown · DC-S1RM2 / 1.5 observations · Not official counter definitions": "정확한 원인 미확인 · DC-S1RM2 / 1.5 관찰값 · 공식 정의 아님",
	"Pixel Refresh note: +3 shutter actuations were observed with [Power-off Shutter] = CLOSE and the required OFF/ON.": "픽셀 리프레시: [전원 끔 셔터]=닫기와 필수 재시작에서 셔터 +3을 관찰했습니다.",
	"The exact +3 breakdown was not isolated. These are behavior tests, not official Panasonic counter definitions.":    "+3의 세부 원인은 분리되지 않았습니다. Panasonic 공식 정의가 아닌 동작 시험 결과입니다.",
	"Camera Model":                           "카메라 모델",
	"Camera Firmware":                        "카메라 펌웨어",
	"Tool Version":                           "도구 버전",
	"Validated By":                           "검증자",
	"Last Validated":                         "마지막 검증일",
	"✓ Shutter Actuations":                   "✓ 셔터 작동 횟수",
	"✓ Power / Wake Activations":             "✓ 전원 / 깨우기 횟수",
	"✓ SetupInfo transport / raw structure":  "✓ SetupInfo 전송 / 원본 구조",
	"△ Error descriptions: legacy / partial": "△ 오류 설명: 기존 자료 / 일부만 확인",
	"? Counters 3–7: unidentified":           "? 카운터 3~7: 의미 미확인",
	"No camera connected · Compatibility / project info is available offline · GitHub Issues preferred · Instagram DM: @bieup_hieut": "카메라 미연결 · 호환성/프로젝트 정보는 오프라인에서 확인 가능 · 제보: GitHub Issues / Instagram @bieup_hieut",
	"RED / BLUE & USB GUIDE":        "빨강·파랑 색상 및 USB 안내",
	"AUTO LANGUAGE & COUNTER NAMES": "언어 자동 선택 및 카운터 명칭",
	"Windows language auto-selection; named counters with clear unverified status.": "Windows 언어 자동 선택; 미검증 상태를 표시하며 카운터 명칭 제공.",
	"Deep red and blue accents; clear USB connection and PC(Tether) steps.":         "진한 빨강·파랑 색상과 USB 연결·PC(테더) 단계 안내.",
	"PALETTE & MASCOT PLACEMENT": "색상과 캐릭터 배치",
	"Warm and teal accents unified; Lumerian moved into dashboard content.": "따뜻한 색·청록을 통일하고 Lumerian을 대시보드 안에 배치.",
	"DARK DASHBOARD DESIGN": "어두운 대시보드 디자인",
	"Dark canvas retained; connection, camera data and action cards redesigned.": "검정 배경을 유지하고 연결·카메라 정보·기능 카드를 재구성.",
	"LUMERIAN BRAND & CHARACTER": "Lumerian 브랜드와 캐릭터",
	"Lumerian identity, four consistent mascot states and branded product surfaces.": "Lumerian 정체성과 네 가지 캐릭터 상태를 제품에 적용.",
	"VERIFIED CAMERA EXPANSION": "검증된 카메라 확대",
	"Three verified camera pairs; report labels, page navigation and serial handling improved.": "세 기종/펌웨어 조합 검증, 리포트·화면 이동·시리얼 표시 개선.",
	"CONNECTION, UX & COMMUNITY VALIDATION":                                                     "연결·사용성·커뮤니티 검증",
	"Disconnected state, validation info, single-instance protection and privacy / mascot UX.":  "연결 끊김·검증 정보·중복 실행 방지·개인정보·캐릭터 개선.",
	"PROJECT INFO & QA": "프로젝트 정보 및 QA",
	"Developer Credits, Version History, Camera Validation Matrix, documentation and regression-review hardening.": "개발자 정보·버전 기록·카메라 검증 목록·문서·회귀 점검 보강.",
	"PNG / UI STABILITY": "PNG·화면 안정성",
	"Save PNG hotfix, Serial Number layout, focused Error History behavior, pixel-mascot rendering and stronger UI feedback.": "PNG 저장·시리얼 배치·오류 기록·캐릭터 렌더링·화면 반응 개선.",
	"SHARING & PRIVACY": "공유와 개인정보",
	"Serial privacy controls, PNG share / verification modes, Copy Summary and refresh-delta display.": "시리얼 공개 설정·PNG 공유/검증·요약 복사·증가량 표시.",
	"METADATA / RESEARCH": "메타데이터 및 연구",
	"Public-beta metadata, read-only transport, worker isolation and early UI foundations.": "공개 베타 메타데이터·읽기 전용 전송·작업 분리·초기 화면 구성.",
	"beta.2 & earlier": "beta.2 및 이전",
	"Lumerian brand & character direction / four-state product mascot system":         "Lumerian 브랜드·캐릭터 기획 / 네 가지 상태의 제품 캐릭터",
	"Concept / product direction / real-camera validation / photography-user testing": "콘셉트·제품 기획·실제 카메라 검증·사진 사용자 시험",
	"Primary validation hardware: Panasonic LUMIX DC-S1RM2 · Firmware 1.5":            "주요 검증 기기: Panasonic LUMIX DC-S1RM2 · 펌웨어 1.5",
	"Instagram DM: @bieup_hieut": "Instagram DM: @bieup_hieut",
	"Bug reports & reproducible validation results: GitHub Issues preferred": "오류 제보 및 재현 가능한 검증 결과: GitHub Issues 권장",
}

func localizeText(text string) string {
	if strings.HasPrefix(text, "L10 response structure not supported: ") {
		return uiText(text, "L10 응답 구조 미지원: "+strings.TrimPrefix(text, "L10 response structure not supported: "))
	}
	if strings.HasPrefix(text, "Could not read a compatible LUMIX camera. Close other camera apps and check USB mode (L10: LUMIX Lab). ") && isKoreanUI() {
		return "호환 카메라를 읽지 못했습니다. 다른 카메라 앱을 종료하고 USB 모드(DC-L10: LUMIX Lab)를 확인하세요. " + strings.TrimPrefix(text, "Could not read a compatible LUMIX camera. Close other camera apps and check USB mode (L10: LUMIX Lab). ")
	}
	if !isKoreanUI() {
		return text
	}
	if translated, ok := koreanText[text]; ok {
		return translated
	}
	if strings.HasPrefix(text, "PUBLIC SHARE · v") {
		return "공개 공유 · v" + strings.TrimPrefix(text, "PUBLIC SHARE · v")
	}
	if strings.HasPrefix(text, "DEVICE VERIFICATION · v") {
		return "기기 검증 · v" + strings.TrimPrefix(text, "DEVICE VERIFICATION · v")
	}
	if strings.HasPrefix(text, "Generated ") && strings.Contains(text, " · Made by @bieup_hieut") {
		return strings.Replace(strings.Replace(text, "Generated ", "생성 시각 ", 1), " · Made by @bieup_hieut", " · 제작 @bieup_hieut", 1)
	}
	if strings.HasPrefix(text, "Last camera read: ") {
		return strings.Replace(strings.Replace(text, "Last camera read: ", "마지막 읽은 카메라: ", 1), " · Bugs: GitHub Issues preferred", " · 오류 제보: GitHub Issues 권장", 1)
	}
	if strings.HasPrefix(text, "Raw service response: ") {
		return strings.Replace(strings.Replace(text, "Raw service response: ", "원본 서비스 응답: ", 1), " bytes · Error records: ", "바이트 · 오류 기록: ", 1)
	}
	if strings.HasPrefix(text, "Raw Counter ") {
		return "원본 카운터 " + strings.TrimPrefix(text, "Raw Counter ")
	}
	if strings.HasPrefix(text, "Ver. ") {
		return "버전 " + strings.TrimPrefix(text, "Ver. ")
	}
	for _, item := range []struct{ prefix, translated string }{
		{"● VERIFIED · ", "● 검증됨 · "},
		{"VERIFIED · ", "검증됨 · "},
		{"● MODEL VERIFIED · FW UNVERIFIED", "● 모델 검증됨 · 펌웨어 미검증"},
		{"MODEL VERIFIED · FIRMWARE UNVERIFIED", "모델 검증됨 · 펌웨어 미검증"},
		{"● UNVERIFIED MODEL · ", "● 미검증 모델 · "},
		{"UNVERIFIED MODEL / FIRMWARE", "모델 / 펌웨어 미검증"},
		{"Read ", "읽은 시각 "},
		{"Preview: ", "미리보기: "},
		{"Generated ", "생성 시각 "},
		{"Validated by ", "검증자 "},
		{"Last camera read: ", "마지막 읽은 카메라: "},
		{"✓ PNG saved: ", "✓ PNG 저장됨: "},
		{"✓ Report saved: ", "✓ 리포트 저장됨: "},
		{"✕ PNG save failed: ", "✕ PNG 저장 실패: "},
		{"✕ Could not save report: ", "✕ 리포트 저장 실패: "},
		{"✕ Could not open Save PNG dialog: ", "✕ PNG 저장 창을 열 수 없음: "},
		{"✕ Could not copy summary: ", "✕ 요약 복사 실패: "},
		{"Raw service response: ", "원본 서비스 응답: "},
	} {
		if strings.HasPrefix(text, item.prefix) {
			suffix := strings.TrimPrefix(text, item.prefix)
			suffix = strings.Replace(suffix, " / Ver. ", " / 버전 ", 1)
			return item.translated + suffix
		}
	}
	if strings.Contains(text, " · FW ") {
		return strings.Replace(text, " · FW ", " · 펌웨어 ", 1)
	}
	if strings.HasSuffix(text, " recorded codes") {
		return strings.TrimSuffix(text, " recorded codes") + "개 기록된 코드"
	}
	if strings.HasPrefix(text, "Page ") {
		var page, total int
		if _, err := fmt.Sscanf(text, "Page %d / %d", &page, &total); err == nil {
			return fmt.Sprintf("%d / %d페이지", page, total)
		}
	}
	return text
}

func buildReport(r UsageResult, status, validator, verifiedDate string) string {
	if isKoreanUI() {
		return buildKoreanReport(r, status, validator, verifiedDate)
	}
	return buildEnglishReport(r, status, validator, verifiedDate)
}

func buildEnglishReport(r UsageResult, status, validator, verifiedDate string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "LUMIX Usage Info Report\r\n")
	fmt.Fprintf(&b, "Brand / Character        Lumerian\r\n")
	fmt.Fprintf(&b, "Tool Version             %s\r\nCamera Data Revision     %s\r\n", appVersion, cameraData.Revision)
	fmt.Fprintf(&b, "========================================\r\n")
	fmt.Fprintf(&b, "Model                    %s\r\n", r.Model)
	fmt.Fprintf(&b, "Firmware                 %s\r\n", r.Firmware)
	fmt.Fprintf(&b, "Serial Number            %s\r\n", serialForMode(r.Serial, "last4"))
	fmt.Fprintf(&b, "Verification             %s\r\n", status)
	fmt.Fprintf(&b, "Validated By             %s\r\n", validator)
	fmt.Fprintf(&b, "Validation Date          %s\r\n\r\n", verifiedDate)
	writeReadSuccess(&b, r, false)
	if r.SemanticsVerified {
		fmt.Fprintf(&b, "VERIFIED USAGE\r\n")
		fmt.Fprintf(&b, "Shutter Actuations       %s\r\n", formatNumber(uint64(r.Shutter)))
		fmt.Fprintf(&b, "Power / Wake Activations %s\r\n\r\n", formatNumber(uint64(r.PowerWake)))
	} else {
		fmt.Fprintf(&b, "USAGE COUNTERS (MEANINGS NOT VERIFIED FOR THIS MODEL / FIRMWARE)\r\n")
		fmt.Fprintf(&b, "Shutter Actuations       %s  (Raw Counter 2; estimated / unverified)\r\n", formatNumber(uint64(r.Shutter)))
		fmt.Fprintf(&b, "Power / Wake Activations %s  (Raw Counter 1; estimated / unverified)\r\n\r\n", formatNumber(uint64(r.PowerWake)))
	}
	fmt.Fprintf(&b, "RAW COUNTERS / ESTIMATED MEANINGS\r\n")
	for i, value := range []uint16{r.Raw3, r.Raw4, r.Raw5, r.Raw6, r.Raw7} {
		fmt.Fprintf(&b, "Counter %-16d %s   %s\r\n", i+3, formatNumber(uint64(value)), counterName(i+3, r))
	}
	fmt.Fprintf(&b, "\r\nCOUNTER BEHAVIOR REFERENCE\r\n")
	fmt.Fprintf(&b, "Tested on DC-S1RM2 / Firmware 1.5. These are behavior observations, not official Panasonic counter definitions.\r\n")
	fmt.Fprintf(&b, "Shutter Actuations:\r\n")
	fmt.Fprintf(&b, "- Mechanical shutter exposure: +1 per physical actuation (verified).\r\n")
	fmt.Fprintf(&b, "- Power OFF with [Power-off Shutter] = CLOSE: +1 (verified).\r\n")
	fmt.Fprintf(&b, "- Electronic shutter exposure: +0 (verified).\r\n")
	fmt.Fprintf(&b, "- High Resolution test: +0 on the tested settings (verified observation).\r\n")
	fmt.Fprintf(&b, "- Sensor Cleaning + required restart: +1 total observed; no separate cleaning count was identified.\r\n")
	fmt.Fprintf(&b, "- Pixel Refresh + required restart: +3 total observed with [Power-off Shutter] = CLOSE; exact breakdown not isolated.\r\n")
	fmt.Fprintf(&b, "Power / Wake Activations:\r\n")
	fmt.Fprintf(&b, "- Camera OFF -> ON cycle: +1 (verified).\r\n")
	fmt.Fprintf(&b, "- Sleep -> wake cycle: +1 (verified as a cycle; exact increment moment not isolated).\r\n")
	fmt.Fprintf(&b, "- USB disconnect / reconnect only: +0 (verified).\r\n\r\n")
	fmt.Fprintf(&b, "ERROR HISTORY\r\n")
	if len(r.Errors) == 0 {
		fmt.Fprintf(&b, "No non-zero error codes found in 16 service-history slots.\r\n")
	} else {
		for _, e := range r.Errors {
			fmt.Fprintf(&b, "%02d  %-18s  %s  %-20s  %s\r\n", e.Index, e.DateText, e.CodeText, e.Category, e.Description)
		}
	}
	fmt.Fprintf(&b, "\r\nNOTES\r\n")
	fmt.Fprintf(&b, "- Read-only access. No camera data is modified.\r\n")
	fmt.Fprintf(&b, "- Error descriptions are decoded from legacy Panasonic service-code references and may not be official mappings for this model.\r\n")
	fmt.Fprintf(&b, "- Counters 3-7 are not behavior-verified for this model / firmware.\r\n- Counter 3/4 names are legacy STBCNT / PSVCNT candidates, not confirmed modern field mappings.\r\n")
	fmt.Fprintf(&b, "- First activation date is not reported because no verified field has been identified.\r\n")
	fmt.Fprintf(&b, "- This report masks the serial number except for the last four digits by default.\r\n")
	fmt.Fprintf(&b, "- Bug reports and additional model/firmware validation: GitHub Issues preferred; Instagram DM %s is also welcome.\r\n", validatedBy)
	fmt.Fprintf(&b, "\r\nGenerated: %s\r\nMade by @bieup_hieut\r\n", time.Now().Format("2006-01-02 15:04:05"))
	return b.String()
}

func buildKoreanReport(r UsageResult, status, validator, verifiedDate string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "LUMIX 사용 정보 리포트\r\n")
	fmt.Fprintf(&b, "브랜드 / 캐릭터       Lumerian\r\n")
	fmt.Fprintf(&b, "도구 버전             %s\r\n기종 데이터 갱신일    %s\r\n", appVersion, cameraData.Revision)
	fmt.Fprintf(&b, "========================================\r\n")
	fmt.Fprintf(&b, "모델                  %s\r\n", r.Model)
	fmt.Fprintf(&b, "펌웨어                %s\r\n", r.Firmware)
	fmt.Fprintf(&b, "시리얼 번호           %s\r\n", serialForMode(r.Serial, "last4"))
	fmt.Fprintf(&b, "검증 상태             %s\r\n", localizeText(status))
	fmt.Fprintf(&b, "검증자                %s\r\n", localizeText(validator))
	fmt.Fprintf(&b, "검증 날짜             %s\r\n\r\n", localizeText(verifiedDate))
	writeReadSuccess(&b, r, true)
	if r.SemanticsVerified {
		fmt.Fprintf(&b, "검증된 사용 횟수\r\n")
		fmt.Fprintf(&b, "셔터 작동 횟수         %s\r\n", formatNumber(uint64(r.Shutter)))
		fmt.Fprintf(&b, "전원 / 깨우기 횟수     %s\r\n\r\n", formatNumber(uint64(r.PowerWake)))
	} else {
		fmt.Fprintf(&b, "사용 횟수 (이 모델 / 펌웨어에서 의미 미검증)\r\n")
		fmt.Fprintf(&b, "셔터 작동 횟수         %s  (원본 카운터 2 · 추정됨 / 미검증)\r\n", formatNumber(uint64(r.Shutter)))
		fmt.Fprintf(&b, "전원 / 깨우기 횟수     %s  (원본 카운터 1 · 추정됨 / 미검증)\r\n\r\n", formatNumber(uint64(r.PowerWake)))
	}
	fmt.Fprintf(&b, "원본 카운터 / 추정 명칭\r\n")
	for i, value := range []uint16{r.Raw3, r.Raw4, r.Raw5, r.Raw6, r.Raw7} {
		fmt.Fprintf(&b, "카운터 %-14d %s   %s\r\n", i+3, formatNumber(uint64(value)), counterName(i+3, r))
	}
	fmt.Fprintf(&b, "\r\n카운터 동작 참고\r\n")
	fmt.Fprintf(&b, "DC-S1RM2 / 펌웨어 1.5에서 관찰했습니다. Panasonic의 공식 카운터 정의가 아닙니다.\r\n")
	fmt.Fprintf(&b, "셔터 작동 횟수:\r\n")
	fmt.Fprintf(&b, "- 기계식 셔터 촬영: 실제 작동 1회당 +1 (검증).\r\n")
	fmt.Fprintf(&b, "- [전원 끔 셔터] = 닫기 상태에서 전원 끄기: +1 (검증).\r\n")
	fmt.Fprintf(&b, "- 전자식 셔터 촬영: +0 (검증).\r\n")
	fmt.Fprintf(&b, "- 고해상도 촬영 시험: 해당 설정에서 +0 (관찰).\r\n")
	fmt.Fprintf(&b, "- 센서 청소 및 필수 재시작: 총 +1 관찰; 별도 청소 횟수는 확인되지 않음.\r\n")
	fmt.Fprintf(&b, "- 픽셀 리프레시 및 필수 재시작: [전원 끔 셔터] = 닫기에서 총 +3 관찰; 세부 원인은 분리되지 않음.\r\n")
	fmt.Fprintf(&b, "전원 / 깨우기 횟수:\r\n")
	fmt.Fprintf(&b, "- 전원 끔 → 켬: +1 (검증).\r\n")
	fmt.Fprintf(&b, "- 절전 → 깨우기: +1 (주기 검증, 증가 시점은 미확인).\r\n")
	fmt.Fprintf(&b, "- USB 연결 해제 / 재연결만 수행: +0 (검증).\r\n\r\n")
	fmt.Fprintf(&b, "오류 기록\r\n")
	if len(r.Errors) == 0 {
		fmt.Fprintf(&b, "16개 서비스 기록에서 0이 아닌 오류 코드가 없습니다.\r\n")
	} else {
		for _, e := range r.Errors {
			fmt.Fprintf(&b, "%02d  %-18s  %s  %-20s  %s\r\n", e.Index, e.DateText, e.CodeText, localizeText(e.Category), localizeText(e.Description))
		}
	}
	fmt.Fprintf(&b, "\r\n참고 사항\r\n")
	fmt.Fprintf(&b, "- 읽기 전용입니다. 카메라 데이터는 변경되지 않습니다.\r\n")
	fmt.Fprintf(&b, "- 오류 설명은 기존 Panasonic 서비스 코드 참고 자료에서 가져왔으며 이 기종의 공식 정의가 아닐 수 있습니다.\r\n")
	fmt.Fprintf(&b, "- 카운터 3~7의 동작은 이 모델 / 펌웨어에서 검증되지 않았습니다.\r\n- 카운터 3/4 명칭은 기존 STBCNT / PSVCNT 후보이며 최신 기종의 필드 대응은 미검증입니다.\r\n")
	fmt.Fprintf(&b, "- 검증된 필드가 없어 최초 사용 날짜를 표시하지 않습니다.\r\n")
	fmt.Fprintf(&b, "- 시리얼 번호는 기본적으로 마지막 4자리만 표시합니다.\r\n")
	fmt.Fprintf(&b, "- 오류 제보와 추가 기종 검증: GitHub Issues 또는 Instagram DM %s.\r\n", validatedBy)
	fmt.Fprintf(&b, "\r\n생성 시각: %s\r\n제작 @bieup_hieut\r\n", time.Now().Format("2006-01-02 15:04:05"))
	return b.String()
}
