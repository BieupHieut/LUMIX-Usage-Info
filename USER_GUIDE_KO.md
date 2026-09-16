# LUMIX Usage Info v1.0.0-beta.6 사용자 가이드

## 1. 연결
1. 카메라 전원을 켭니다.
2. USB 케이블로 PC와 연결합니다.
3. 카메라 USB 모드를 **PC(Tether)** 로 설정합니다.
4. LUMIX Tether가 실행 중이면 종료합니다.
5. 프로그램을 실행합니다.

## 2. 메인 화면
DC-S1RM2 / FW 1.5에서 검증된 값은 다음과 같습니다.
- **Shutter Actuations**: 실제 물리 셔터 작동 횟수
- **Power / Wake Activations**: 전원 ON 및 절전 복귀와 관련된 활성화 횟수

Serial Number는 기본적으로 `********1234`처럼 마지막 4자리만 보입니다.

## 3. 연결이 끊겼을 때
beta.6에서는 이전 정상 읽기 결과가 있으면 값을 지우지 않습니다.
- 상태: **Disconnected · last data**
- 안내: 마지막 정상 읽기 값을 보여주고 있음을 표시
- USB 재연결 후 **REFRESH**하면 현재 값을 다시 읽습니다.
- Windows 장치 변경 이벤트가 감지되면 프로그램이 별도 Worker에서 자동으로 연결 상태를 재확인할 수 있습니다.

중요: Disconnected 상태의 값은 “마지막 정상 읽기 값”이지 실시간 연결 상태의 새 값이 아닙니다.

## 4. VALIDATION & INFO
카메라가 연결되지 않아도 접근할 수 있습니다.
- **Camera List**: 검증 완료 / 검증 필요 모델
- **Version History**: 버전별 변경 사항
- **Developer Credits**: 제작자 및 프로젝트 정보

현재 실제 카운터 의미까지 검증된 조합은 **DC-S1RM2 / FW 1.5**입니다.

## 5. Save PNG
- **Public Share**: 공개 공유용
- **Device Verification**: 중고 거래 / 소유 기기 인증용
- Serial 표시: Masked / Last 4 / Full 선택 가능

PNG 저장 중에는 관련 옵션을 중복 조작하지 못하도록 처리됩니다.

## 6. Save Report
TXT 리포트의 Serial Number는 기본적으로 **마지막 4자리만 표시**됩니다.

## 7. Error History
기록이 없으면 다음 메시지를 표시합니다.
`No recorded non-zero error codes found.`

오류 코드 설명은 구형 Panasonic 서비스 코드 자료를 참고한 부분 해석이며 S1RM2 공식 오류표로 단정하지 않습니다.

## 8. 알려진 카운터 동작
### Shutter Actuations
- 기계식 셔터 1회: +1 확인
- 전자셔터 1회: +0 확인
- 전원 차단 시 셔터 동작 = CLOSE: 전원 OFF 시 +1 확인
- 고해상도 촬영 테스트: +0 확인
- Sensor Cleaning + 요구된 재시작: 총 +1 관찰
- Pixel Refresh + 요구된 재시작: 총 +3 관찰(정확한 내부 구성은 분리 검증되지 않음)

### Power / Wake Activations
- OFF → ON: +1 확인
- Sleep → Wake 사이클: +1 확인
- USB 분리/재연결만: +0 확인

## 9. 문의 / 제보
GitHub Issues를 권장합니다. 간단한 문의나 스크린샷 제보는 Instagram DM **@bieup_hieut**로도 가능합니다.

공개 스크린샷에는 전체 Serial Number가 포함되지 않도록 확인해 주세요.

## 카메라 검증 상태
- **DC-S1RM2 / FW 1.5:** FULLY VERIFIED — 카운터 의미까지 실기 검증 완료.
- **DC-S5 / FW 2.9:** READ COMPATIBLE — **아름프로**님이 v1.0.0-beta.3으로 실제 데이터 읽기에 성공. 단, 카운터 의미는 아직 검증 전이므로 Raw Counter로만 표시합니다.
- 그 외 목록의 카메라: NEEDS VALIDATION.
