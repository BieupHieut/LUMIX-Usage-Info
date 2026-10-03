# 검증 데이터 업데이트 / Camera data updates

EXE에는 현재 검증 목록이 포함돼 있어 **JSON 없이도 실행할 수 있습니다**.
새 검증 목록을 따로 적용할 때만 `camera_profiles.json`을 사용합니다.

The EXE includes its verification list and **runs without a JSON file**.
The optional `camera_profiles.json` applies newer verification data separately.

## 적용 / Apply

1. 이 저장소의 최신 `camera_profiles.json`을 내려받습니다.
2. EXE와 같은 폴더에 놓습니다.
3. 프로그램을 종료하고 다시 실행합니다.

Download the updated JSON from this repository, place it beside the EXE, and restart.
Missing or invalid data falls back to the embedded list. The reader has no automatic online update.

파일이 없거나 형식이 잘못되면 내장 목록을 사용합니다. 자동 인터넷 업데이트는 없습니다.
이 파일은 검증 표시만 갱신하며 카메라 명령이나 카운터 해석 방식을 변경하지 않습니다.

## 버전 / Versioning

프로그램 기능·버그 수정은 앱 버전으로, 기종 검증 추가는 데이터 날짜로 관리합니다.
현재 데이터 갱신일은 **2026-10-02**입니다. 실제 기종 검증 날짜와 데이터 발행 날짜는 다를 수 있습니다.

Software changes use application versions; verified-pair updates use data revision dates.
Current revision: **2026-10-02**. A data publication date is not necessarily the test date.

새 기종 결과는 [Issues](https://github.com/BieupHieut/LUMIX-Usage-Info/issues/new/choose)로 보내 주세요.
프로젝트에서 확인한 정확한 모델·펌웨어 조합을 목록에 반영합니다.

## 소스 관리자 참고 / Maintainer note

검증 데이터를 변경할 때 루트와 `source/camera_profiles.json`을 함께 갱신합니다.
스키마 1에서 셔터·전원 카운터가 모두 확인된 조합을 등록합니다. 실제 검증 날짜를 모르면 `validation_date`를 생략합니다.

Update both JSON copies for source builds. Schema 1 holds pairs with both shutter and power/wake meanings confirmed.
Omit `validation_date` when the actual test date is unknown.
