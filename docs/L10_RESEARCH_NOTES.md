> Historical reference only. For the current v1.0.0 flow and verification, see ../USER_GUIDE_EN.md and ../VALIDATION_MATRIX.md.

> Historical beta.16 research/reference. beta.17 uses the regular L10 reader based on the new user report; see current release notes and L10_TEST_GUIDE_KO.md. Old A/B/C descriptions are not the current user flow.

# LUMIX 사용 이력 연구 노트

확인일: 2026-10-02 · 적용 배포: beta.16 · 목적: 읽기 전용 진단의 근거와 미확인 사항 구분

## 1. 공식 Data Act 단서

[Panasonic 사용 이력 안내](https://www.panasonic.com/uk/consumer/eu-data-act/dsc.html)는 서비스 진단에 활용할 기록을 설명합니다. USB 명령·필드 위치는 공개하지 않습니다.

| 모델군 예시 | 주요 기록 | 예상량 |
| --- | --- | --- |
| G100D/G97/TZ99/TZ300 | 전원·셔터·플래시·절전 횟수 | 20 bytes |
| S1RM2/S1M2/S1M2ES/S5M2X/S5M2/S9/GH7/G9M2/L10 | 같은 종류의 횟수 | <16 Kbytes |
| S5D/DC-GHM2/AW-UB10/AW-UB50 | 같은 종류의 횟수 | 52 bytes |
| X2/X20/CX20 등 캠코더군 | 사용·녹화·표시·팬 시간, 촬영·모터·필터·조작 횟수 | 68 bytes |

모두 rawdata로 안내됩니다. DC-GHM2는 원문 표기이며 다른 모델명으로 임의 정정하지 않습니다. 이 표는 기종별 예상 기록량을 요약한 것이며 현재 SetupInfo의 146바이트 응답과 동일하다는 뜻은 아닙니다.

### 개발 판단

- 플래시·절전은 Counter 3–7의 동작을 조사할 후보입니다. 의미를 확정할 필드 근거가 없어 UNKNOWN을 유지합니다.
- 전원 on/off 문구만으로 모든 기종의 Power/Wake 동작이 같다고 판단하지 않습니다.
- 크기 차이는 구조 차이의 단서일 수 있으나, 바이트 배치나 통신 경로를 추정해 구현할 근거는 아닙니다.
- 캠코더 누적 시간은 향후 연구 대상입니다. 현재 제품에 사용 시간·건강도·수명 점수를 추가하지 않습니다.

향후 동작 검증은 한 동작씩 전후 차이를 비교해야 합니다. 촬영 모드·전원 종료 셔터·재시작·절전·플래시 조건을 기록하고, 전원 켜짐/꺼짐과 절전 진입/깨우기를 분리해 관찰합니다. 이번 배포는 이러한 행동을 자동으로 실행하지 않습니다.

## 2. LUMIX Lab / S9 / SDK 공식 자료

- [DC-L10 manual — LUMIX Lab USB connection](https://eww.pavc.panasonic.co.jp/dscoi/DC-L10/html/DC-L10_DVQP3497_eng/0132.html): 스마트폰 Lab USB 연결 절차. Windows WPD/PTP 노출 또는 서비스 읽기 성공을 증명하지 않습니다.
- [DC-S9 firmware changelog](https://av.jpn.support.panasonic.com/support/dsc/download/ff/dl/s9.html): FW2.0에서 Lab 3.0.0 이상 유선 연결 지원. S9의 정상 사용은 기존 PC(Tether) 방식으로 유지합니다. Lab 비교 테스트를 추가하지 않습니다.
- [LUMIX Lab model features](https://av.jpn.support.panasonic.com/support/software/lumix_lab/index.html): 기종·펌웨어별 유선 기능 확인용.
- [LUMIX Tether supported cameras](https://av.jpn.support.panasonic.com/support/global/cs/soft/download/d_lumixtether.html): L10의 실험 경로를 일반 테더 지원 목록과 구분합니다.
- [LUMIX SDK](https://av.jpn.support.panasonic.com/support/software/tool/sdk.html): 공개 지원 목록에 S9은 있고 L10은 없습니다. 공개 페이지 확인만으로 SDK 내부 전체 API를 조사했다고 주장하지 않습니다. SDK를 통한 L10 셔터 조회 근거는 확보하지 못했습니다.

## 3. 사용자 제공 서비스 소프트웨어 조사

사용자가 공유한 조사에는 과거 DMC-F3의 DscCalDi 조정 소프트웨어와 Panasonic TSN 배포, 다른 카메라/렌즈의 PC 서비스 조정 사례가 포함됩니다. 이 작업에서는 해당 서비스 매뉴얼 원문·페이지를 직접 확인하지 않았으므로 **사용자 제공 연구 단서**로 보관합니다.

역사적 서비스 사례가 확인되더라도 2026년 DC-L10의 프로그램·USB 모드·읽기 명령을 특정하지 못합니다. 전용 서비스 모드나 인터페이스 가능성은 가설입니다. 자료 확보 시 모델/문서 버전/페이지/공개 출처를 붙여 현재 A단계의 VID/PID·인터페이스·드라이버 자료와 비교할 수 있습니다.

## 4. 이번 구현의 경계

A: OS 열거 → B: 동의 후 표준 PTP → C: DC-L10 식별·지원 명령·동일 기기 확인·새 동의 후 기존 읽기 1회. 보고서는 모든 원시 값의 의미를 UNVERIFIED로 유지합니다.

`0x9414` 응답의 태그와 146바이트 검사는 기존 툴의 알려진 구조와 호환되는지 보는 엄격한 조건입니다. Data Act의 <16Kbytes를 근거로 길이를 확대하거나 다른 서비스 명령을 시도하지 않습니다. 서비스 메뉴 진입, 조정 프로그램 실행, USB 모드 변경, 펌웨어 변경, 드라이버 교체 기능은 구현하지 않습니다.

## 5. Panasonic 문의 초안 — 전송하지 않음

제목: LUMIX DC-L10 사용 이력 및 읽기 전용 진단 인터페이스 문의

안녕하세요. 비공식 LUMIX 사용 이력 조회 프로그램을 개발하며 DC-L10 지원 가능성을 검토하고 있습니다. 귀사의 Data Act 안내에서 DC-L10의 전원·셔터·플래시·절전 기록과 서비스 진단 활용을 확인했습니다.

1. 공식 서비스센터에서 DC-L10의 누적 셔터와 기타 사용 이력을 조회할 수 있나요?
2. 일반 USB 단자, 별도 연결 모드 또는 전용 장비 중 어떤 경로를 사용하는지 공개 가능한 범위에서 안내받을 수 있나요?
3. LUMIX Lab 또는 PC(Storage) 유선 연결로 사용 기록을 읽는 공식 방법이 있나요?
4. 일반 개발자에게 제공 가능한 읽기 전용 인터페이스/API/기술 문서가 있나요?
5. 공식 표의 기록 항목과 실제 데이터 필드·단위·동작 조건을 연결하는 공개 설명이 있나요?

공개 가능한 범위의 안내를 부탁드립니다. 감사합니다.

Data Act 페이지에 기재된 `DataAct@eu.panasonic.com`은 데이터 문의 단서입니다. 한국 제품 기술지원 처리 여부는 미확인입니다. 이 파일은 문의 초안이며 이메일이나 게시글을 전송한 기록이 아닙니다.
