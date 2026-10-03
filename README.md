# LUMERIAN-LUMIX 사용 정보 / Usage Info

LUMIX 카메라의 사용 기록을 확인하는 **Windows x64용 비공식 읽기 전용 도구**입니다.
카메라의 설정과 데이터를 변경하지 않습니다. 한국어와 영어를 지원합니다.

An unofficial, read-only usage reader for LUMIX cameras on Windows x64.
It does not change camera settings or data. Korean and English are supported.

## 다운로드 / Download

**[EXE 다운로드 / Download EXE](https://github.com/BieupHieut/LUMIX-Usage-Info/releases/download/v1.0.0/LUMIX_Usage_Info_v1.0.0.exe)** · [릴리즈 페이지 / Releases](https://github.com/BieupHieut/LUMIX-Usage-Info/releases/latest)

**EXE 하나만 받아 실행하면 됩니다. 별도 설치는 필요 없습니다.** 이미지와 현재 검증 데이터가 내장되어 있습니다.
`camera_profiles.json`은 검증 목록을 별도로 갱신할 때만 사용하는 선택 파일입니다. ZIP도 릴리즈 페이지에서 받을 수 있습니다.

**Download and run the EXE; no installation is required.** Artwork and the current verification data are embedded.
The optional `camera_profiles.json` updates the verification list separately. A ZIP package is also available on the release page.

## 시작하기 / Quick start

1. 컴퓨터의 다른 카메라 앱을 종료하고 카메라 전원을 켭니다.
2. 데이터 전송이 가능한 USB 케이블로 컴퓨터와 연결합니다.
3. 카메라에서 **PC(테더)**를 선택합니다. **DC-L10은 LUMIX Lab**을 선택합니다.
4. 프로그램을 실행하고 **카메라 연결 / 새로고침**을 누릅니다.

Close other camera apps → power on → connect a USB data cable → select **PC(Tether)** (**LUMIX Lab for DC-L10**) → run the EXE → **Connect / Refresh**.

## 주요 기능 / Features

- 모델·펌웨어·시리얼 정보 / Model, firmware and serial information
- 셔터 작동 횟수, 전원·깨우기 횟수 / Shutter and power/wake counts
- 원시 카운터 1–7 및 오류 기록 / Raw counters 1–7 and error history
- PNG·TXT 저장, 요약 복사, 시리얼 공개 범위 선택 / PNG/TXT export, copy summary and serial privacy
- Windows 언어 자동 선택, 한국어·영어 수동 전환 / Windows language default and manual Korean/English selection

검증은 **정확한 모델·펌웨어 조합**에 한정됩니다. 미검증 카운터 명칭은 ‘추정됨’으로 표시합니다.
셔터 작동 횟수는 사진 파일 개수와 다를 수 있으며, 남은 수명이나 고장 시점을 나타내지 않습니다.

Verification applies to **exact model/firmware pairs**. Unverified counter meanings are marked as estimated.
Shutter actuations may differ from the number of photographs and do not predict remaining life or failure dates.

## 안내 / Guides

- [한국어 사용법](docs/USER_GUIDE_KO.md) · [English user guide](docs/USER_GUIDE_EN.md)
- [검증 기종·지원 추정 목록 / Camera support](docs/CAMERA_SUPPORT.md)
- [검증 데이터 업데이트 / Camera data updates](docs/CAMERA_DATA.md)
- [버전 기록 / Changelog](CHANGELOG.md) · [제작·감사 / Credits](CREDITS.md)

프로그램 **v1.0.0 · 2026-10-03** / Version **v1.0.0 · 2026-10-03**

[문제 신고·기종 확인 / Issues](https://github.com/BieupHieut/LUMIX-Usage-Info/issues) · [소스 빌드 / Build from source](docs/BUILD.md)

Panasonic의 공식 프로그램이 아닙니다. / This project is not affiliated with or endorsed by Panasonic.
