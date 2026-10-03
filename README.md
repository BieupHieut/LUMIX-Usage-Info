# LUMERIAN-LUMIX Usage Info

[English](#english) · [한국어](#한국어)

## English

An unofficial, read-only tool for viewing LUMIX camera usage information on **Windows x64**.
It does not change camera settings or data. English and Korean are supported.

### Download

**[Download EXE](https://github.com/BieupHieut/LUMIX-Usage-Info/releases/download/v1.0.0/LUMIX_Usage_Info_v1.0.0.exe)** · [Release page](https://github.com/BieupHieut/LUMIX-Usage-Info/releases/latest)

**Download and run the EXE; no installation is required.** Artwork and the current camera verification data are embedded.
The optional `camera_profiles.json` is only needed to update the verification list separately. A ZIP package is also available on the release page.

### Quick start

1. Close other camera applications on the PC and turn on the camera.
2. Connect the camera to the PC using a USB cable that supports data transfer.
3. Select **PC(Tether)** on the camera. For **DC-L10**, select **LUMIX Lab**.
4. Run the tool and select **Connect / Refresh**.

### Features

- Model, firmware and serial information
- Shutter actuations and power/wake counts
- Raw counters 1–7 and error history
- PNG/TXT export, copy summary and serial visibility options
- Automatic Windows language selection and manual English/Korean switching

### Understanding the counters

Verification applies to **exact model/firmware pairs**. Unconfirmed counter meanings are marked as **estimated** or **unidentified**.
Shutter actuations may differ from the number of photographs and do not indicate remaining shutter life or failure dates.

### Guides

- [English user guide](docs/USER_GUIDE_EN.md)
- [Verified cameras and estimated support list](docs/CAMERA_SUPPORT.md)
- [Camera verification data updates](docs/CAMERA_DATA.md)
- [Changelog](CHANGELOG.md) · [Creator and contributors](CREDITS.md)
- [Bug reports and camera verification](https://github.com/BieupHieut/LUMIX-Usage-Info/issues) · [Build from source](docs/BUILD.md)

**v1.0.0 · Released 2026-10-03**

This is an unofficial community project and is not affiliated with Panasonic.

## 한국어

**Windows x64**에서 LUMIX 카메라의 사용 정보를 확인하는 비공식 읽기 전용 도구입니다.
카메라의 설정과 데이터를 변경하지 않습니다. 영어와 한국어를 지원합니다.

### 다운로드

**[EXE 다운로드](https://github.com/BieupHieut/LUMIX-Usage-Info/releases/download/v1.0.0/LUMIX_Usage_Info_v1.0.0.exe)** · [릴리즈 페이지](https://github.com/BieupHieut/LUMIX-Usage-Info/releases/latest)

**EXE 하나만 받아 실행하면 됩니다. 별도 설치는 필요 없습니다.** 이미지와 현재 기종 검증 데이터가 내장되어 있습니다.
`camera_profiles.json`은 검증 목록을 별도로 업데이트할 때만 사용하는 선택 파일입니다. ZIP도 릴리즈 페이지에서 받을 수 있습니다.

### 시작하기

1. 컴퓨터의 다른 카메라 앱을 종료하고 카메라 전원을 켭니다.
2. 데이터 전송이 가능한 USB 케이블로 카메라와 컴퓨터를 연결합니다.
3. 카메라에서 **PC(테더)**를 선택합니다. **DC-L10**은 **LUMIX Lab**을 선택합니다.
4. 프로그램을 실행하고 **카메라 연결 / 새로고침**을 누릅니다.

### 주요 기능

- 모델·펌웨어·시리얼 정보
- 셔터 작동 횟수와 전원·깨우기 횟수
- 원시 카운터 1–7 및 오류 기록
- PNG·TXT 저장, 요약 복사, 시리얼 공개 범위 선택
- Windows 언어 자동 선택과 영어·한국어 수동 전환

### 카운터 이해하기

검증은 **정확한 모델·펌웨어 조합**에 한정됩니다. 확인되지 않은 카운터의 의미는 **추정됨** 또는 **의미 미확인**으로 표시합니다.
셔터 작동 횟수는 사진 개수와 다를 수 있으며, 남은 셔터 수명이나 고장 시점을 나타내지 않습니다.

### 안내

- [한국어 사용법](docs/USER_GUIDE_KO.md)
- [검증 기종·지원 추정 목록](docs/CAMERA_SUPPORT.md)
- [기종 검증 데이터 업데이트](docs/CAMERA_DATA.md)
- [버전 기록](CHANGELOG.md) · [제작자와 도움을 주신 분들](CREDITS.md)
- [문제 신고·기종 확인](https://github.com/BieupHieut/LUMIX-Usage-Info/issues) · [소스 빌드](docs/BUILD.md)

**v1.0.0 · 릴리즈 2026-10-03**

Panasonic과 무관한 비공식 커뮤니티 프로젝트입니다.
