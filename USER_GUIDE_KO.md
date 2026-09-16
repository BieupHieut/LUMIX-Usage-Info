# LUMIX Usage Info v1.0.0-beta.3 사용자 가이드

## 1. 이 프로그램은 무엇인가요?

Panasonic LUMIX 카메라를 USB로 연결해 카메라 내부의 일부 사용/진단 정보를 **읽기 전용**으로 확인하는 비공식 커뮤니티 도구입니다.

현재 실기 검증 조합은 **DC-S1RM2 / 펌웨어 Ver. 1.5**입니다. 다른 기종 또는 다른 펌웨어에서는 Raw 값이 읽히더라도 그 의미를 검증 완료로 표시하지 않습니다.

## 2. 연결 방법

1. 카메라 전원을 켭니다.
2. USB 케이블로 Windows PC에 연결합니다.
3. 카메라 USB Mode를 **[PC(Tether)]**로 설정합니다.
4. LUMIX Tether가 실행 중이면 종료합니다.
5. `LUMIX_Usage_Info_v1.0.0-beta.3.exe`를 실행합니다.

## 3. 메인 화면

### Shutter Actuations
실제 **물리 셔터 작동 횟수**로 검증된 값입니다.

사진 파일 개수와 동일하지 않습니다. 예를 들어 전자셔터 촬영은 사진 파일이 생겨도 이 값이 증가하지 않았습니다.

### Power / Wake Activations
전원 ON 및 절전 복귀와 관련해 증가하는 값입니다.

- OFF → ON: +1 확인
- Sleep → Wake 사이클: +1 확인
- USB 재연결만 수행: +0 확인

### Refresh 후 변화량
같은 카메라를 연속해서 읽은 경우 숫자 옆에 `+1`, `+2`처럼 **직전 Refresh 대비 변화량**을 표시합니다. 다른 카메라로 바뀌면 이전 바디와 비교하지 않습니다.

## 4. 셔터 수가 촬영 외에 증가할 수 있는 경우

DC-S1RM2 / FW 1.5에서 실제로 관찰된 내용입니다.

- 기계식 셔터 촬영: 물리 작동 1회당 +1
- `[전원 차단 시 셔터 동작] = CLOSE`: 전원을 끌 때 셔터가 닫히므로 +1이 발생할 수 있음
- 전자셔터 촬영: +0 확인
- 고해상도 촬영 테스트: +0 확인
- Pixel Refresh 후 요구되는 재시작 과정에서는 Shutter Count 증가가 관찰됨. 다만 Pixel Refresh 자체와 전원 종료 셔터 동작의 정확한 분리는 아직 검증 중이므로 고정 증가량으로 해석하면 안 됩니다.

## 5. Serial Number 개인정보 보호

메인 화면에서는 Serial Number를 기본적으로 일부 가려서 표시합니다.

`SHOW / HIDE` 버튼으로 전체 시리얼 번호 표시 여부를 전환할 수 있습니다.

## 6. Save PNG

`SAVE PNG`를 누르면 공유 목적을 선택할 수 있습니다.

### Public Share
SNS, 커뮤니티, GitHub Issue 등에 공유하기 위한 개인정보 우선 프리셋입니다.

### Device Verification
중고거래 또는 본인 기기 인증처럼 해당 바디의 Serial Number 확인이 필요한 경우를 위한 프리셋입니다.

### Serial 표시 방식

- **Masked** — 전체 가림
- **Last 4 digits** — 마지막 4자리만 표시
- **Full Serial** — 전체 시리얼 번호 표시

Full Serial이 포함된 PNG는 공개 게시 전에 시리얼 노출 여부를 다시 확인하세요.

PNG에는 Windows 창 테두리 대신 결과 정보만 정리된 공유용 카드가 저장됩니다. 기본 저장 위치는 가능한 경우:

`문서\LUMIX Usage Info Reports`

입니다.

## 7. Copy Summary

Export 화면의 `COPY SUMMARY`를 누르면 선택한 Serial 표시 방식을 유지한 상태로 모델, 펌웨어, 사용 카운터, 검증 상태를 클립보드에 복사합니다.

GitHub Issue, 커뮤니티 글, Instagram DM 등에 바로 붙여넣을 수 있습니다.

## 8. Error History

최대 16개의 서비스 Error History 슬롯을 읽습니다.

- `00000000`은 기록된 에러로 세지 않습니다.
- 설명 문구는 구형 Panasonic 서비스 코드 자료를 참고해 부분 해석한 것이며 **S1RM2 공식 오류코드 표라는 의미는 아닙니다.**

## 9. Validation Info

현재 앱에서 실제 하드웨어로 확인된 조합을 보여줍니다.

- Camera Model: DC-S1RM2
- Firmware: Ver. 1.5
- Tool Version: v1.0.0-beta.3
- Validated By: @bieup_hieut

## 10. 문의 / 테스트 결과 공유

가능하면 GitHub Issues를 이용해 주세요. 결과가 공개적으로 축적되어 다른 사용자 검증에도 도움이 됩니다.

Instagram DM: **@bieup_hieut**

제보 시 아래 정보를 포함하면 좋습니다.

- 카메라 모델
- 펌웨어 버전
- 툴 버전
- Windows 버전
- 어떤 동작을 했는지
- Refresh 전/후 변화량

공개 스크린샷이나 리포트에서는 의도적으로 인증하려는 경우가 아니라면 Serial Number를 가리는 것을 권장합니다.

## 11. 안전성

이 프로그램은 카메라 데이터를 수정하지 않는 **Read-only** 도구입니다. 공개 버전에는 Panasonic 쓰기 명령이 구현되어 있지 않습니다.
