# Lumerian · LUMIX Usage Info v1.0.0

Unofficial read-only LUMIX usage reader for Windows x64. Korean / English with Windows language default, manual switch and saved preference. Dark interface with Lumerian character artwork.

## 다운로드 및 실행 / Download

저장소의 **Releases**에서 `LUMIX_Usage_Info_v1.0.0_Windows_x64.zip`을 내려받아 압축을 풀고 실행하세요. 이 소스 폴더에는 실행 파일이 포함되지 않습니다.

카메라 전원 켜기 → 데이터 USB 케이블로 PC 연결 → **PC(테더)** 선택(DC-L10은 **LUMIX Lab**) → 프로그램의 카메라 연결 / 새로고침.

**읽기 전용: 카메라의 설정과 데이터를 변경하지 않습니다.** 검증은 표시된 모델·펌웨어 조합에 한정됩니다. 한/영은 Windows 언어를 따르며 직접 변경할 수도 있습니다.

Download the Windows x64 ZIP from this repository's **Releases**, extract it, and run the executable. This source checkout contains no executable. The reader does not change camera settings or data.

## Start

Close other camera apps → turn on camera → USB data cable to this PC → select **PC(Tether)**, or **LUMIX Lab for DC-L10** → run `LUMIX_Usage_Info_v1.0.0.exe` → Connect / Refresh.

## Features

- Model, firmware, masked serial; shutter and power/wake counts.
- Raw counters 1–7; estimated names distinguished from verified meanings.
- Up to 16 service-history entries with legacy reference descriptions.
- PNG save with a location dialog; TXT report save; copy summary; open last file / folder.
- Korean / English automatic selection and manual switch; DPI, scrolling and keyboard navigation.
- Community verification list, USB connection models, version history and Special Thanks.

## Independent camera data

App version **1.0.0** tracks software changes. Camera data revision **2026-10-02** tracks model / firmware verification updates. `camera_profiles.json` next to the executable can be updated independently; the bundled data is used if the file is absent or invalid. Restart to apply. See [camera data guide](CAMERA_DATA_GUIDE.md). The unified camera list shows exact verified pairs first, then estimated support models, seven rows per page. A new verified model replaces its estimated row; other firmware remains unverified. Release history shows app versions and release dates only: v1.0.0 / 2026-10-03.

## Verified pairs

| Model | Firmware | Shutter / Power-Wake | Verified by | Validation date |
|---|---|---|---|---|
| DC-S1RM2 | 1.5 | VERIFIED / VERIFIED | @bieup_hieut | 2026-09-16 |
| DC-S5M2 | 3.7 | VERIFIED / VERIFIED | 엘가, 제비동선 | Not supplied |
| DC-S5 | 2.9 | VERIFIED / VERIFIED | 아름프로 | Not supplied |
| DC-L10 | 1.2 | VERIFIED / VERIFIED | 잠이든 | Not supplied |
| DC-S1M2 | 1.4 | VERIFIED / VERIFIED | 잠이든 | Not supplied |

Verification applies to the exact pair. DC-S5M2 3.7 remains verified; other firmware, S9, G9M2 and GH7 are estimated unless added after confirmation. Community confirmation is distinct from developer testing of this build on hardware. Detailed S1RM2 behavior rules are not automatically extended to other models.

## Saving

Default: `%LOCALAPPDATA%\Lumerian\LUMIX Usage Info\Reports`. TXT saves use unique names. PNG lets you choose a location. Windows security policies can still restrict a chosen folder; see the [user guide](USER_GUIDE_EN.md).

## Guides / build

[한국어](USER_GUIDE_KO.md) · [English](USER_GUIDE_EN.md) · [Release notes](RELEASE_NOTES_v1.0.0.md) · [Final review](FINAL_REVIEW.md).

Windows Go: from `source`, set GO111MODULE=off, GOOS=windows, GOARCH=amd64, CGO_ENABLED=0. Run `go test .`, `go vet .`, then `go build -trimpath -ldflags="-H=windowsgui -s -w" -o ../LUMIX_Usage_Info_v1.0.0.exe .`. Keep source/camera_profiles.json for embedding; the root copy is the editable runtime data. No third-party Go dependencies.

Special Thanks: **Panasonic LUMIX Café Forum community members** / 파나소닉 루믹스 카페 포럼 회원 여러분 · https://cafe.naver.com/panalumix

## Source distribution

The repository contains source, build assets, documentation and historical development references. User download files are attached to releases. See [upload guide](GITHUB_PUBLISHING_GUIDE_KO.md).
