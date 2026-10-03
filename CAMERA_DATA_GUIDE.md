# 기종 검증 데이터 관리 / Camera Data

## 정책

- 프로그램 기능·버그·UI 변경: `1.0.1`, `1.1.0` 등 앱 버전을 변경합니다.
- 검증 기종/펌웨어/검증자 추가: 앱 **1.0.0** 유지, `camera_profiles.json`의 `revision` 날짜를 갱신합니다.
- 같은 날 다시 배포한 데이터는 배포 파일의 SHA-256과 변경 이력으로 구분합니다. 실제 검증 날짜와 데이터 발행 날짜는 별개입니다.
- 공개 배포는 프로젝트 관리자가 확인한 조합만 등록합니다. USB 연결 목록에 있다는 이유로 VERIFIED를 추가하지 않습니다.

## 배포 / Apply

1. `camera_profiles.json`을 실행 파일 옆에 놓습니다. 앱을 종료하고 다시 실행하면 적용됩니다.
2. 정확한 `model` / `firmware` / `validated_by`를 등록합니다. 검증 날짜를 모르면 `validation_date`를 생략합니다. 날짜를 추측하지 않습니다.
3. `revision`은 데이터 발행일 `YYYY-MM-DD`. JSON schema는 1입니다. 카운터 1/2가 모두 확인된 조합만 등록합니다. 부분 검증은 현재 스키마가 지원하지 않습니다.
4. 파일이 없거나 JSON/schema/필드가 잘못되면 내장 데이터로 시작합니다. 잘못된 외부 데이터는 카메라 목록 화면에 안내합니다.
5. 카메라 목록은 검증 조합을 먼저 표시하고, 나머지는 지원 추정 목록으로 이어집니다. 같은 모델의 여러 검증 펌웨어는 각각 표시하며, 검증된 모델은 지원 추정 행과 중복 표시하지 않습니다. 7개씩 좌우 페이지 버튼으로 확인할 수 있습니다. 검증 참여자는 별도의 5개씩 페이지로 표시합니다.

예제는 배포 파일 자체를 참고하세요. 기종/펌웨어는 정확히 구분하고 S1M2의 검증을 S1M2ES로 확대하지 않습니다. 외부 파일은 로컬 관리 데이터이며 제조사 서명이나 공식 인증이 아닙니다. 자동 인터넷 업데이트는 없습니다.

English: Software changes bump the application version; verification updates change the data revision date. Replace the JSON beside the executable and restart. Only exact pairs with both shutter and power/wake confirmed belong in this schema. Unknown test dates are omitted. Missing/invalid data falls back to the embedded registry. This data does not change camera commands, counters 3–7, error decoding or manufacturer certification.

## Change history

| Data revision | Change |
|---|---|
| 2026-10-02 | Initial standalone data: five existing verified pairs; L10 1.2 / S1M2 1.4 community confirmation included. |
