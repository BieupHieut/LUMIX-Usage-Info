# 1.0.0 final QA — 2026-10-03

## Passed locally

- Go test suite and go vet (Windows x64, Go 1.27.1); production GUI binary build.
- Existing exact five pairs, unknown firmware and variant rejection; regular mocked L10/S1M2 reads.
- Korean/English auto/manual language, remembered settings, serial masking and snapshot identity.
- Requested S9/S5M2 unknown firmware/G9M2/GH7: estimated 1/2 names; 3/4 estimates; 5–7 unidentified. Report data revision and privacy checks.
- Camera registry schema/date/size/duplicates/unknown fields/trailing JSON rejection; data-only addition and pagination. Runtime executable metadata smoke checks use isolated synthetic files.
- Same-second TXT saves retain both files; file-as-directory failure does not claim success. PNG replacement produces valid image and removes temporary files.
- Truncated headers/payload and duplicate service frames rejected; normal field offsets preserved.
- GDI rendering/text metrics in Korean/English: welcome, verified L10/S1M2 home, requested unverified homes, technical details, behavior reference, error history, verification, future registry page, credits, history, export controls and PNG card. No unintended single-line clipping found.
- Real saveExportPNGToPath rendered Korean/English 1000×650 PNGs with estimated labels. Images inspected locally.
- 125/150/200% welcome/verification/credits rendering; hidden native Win32 tests for fit/small/resize/move/forced-bar cleanup/keyboard navigation/DPI-change handling.

## Boundaries

No developer camera connected in this session. Community evidence establishes existing meanings; mock protocol tests do not confirm this binary on physical cameras. Native interactive save-dialog behavior, deliberate Defender blocking, organization policy and movement between physical monitors with different DPI were not reproduced. No claim that all firmware variants work or counters 3–7 / modern error codes are fully decoded.

Packaging: root/ZIP SHA-256 manifests and installed files must be verified after packaging; see the local verification script/output. No GitHub posting occurred.


## 2026-10-02 배포 전 UI 수정

- 사용자가 기존 1.0.0의 카메라 연결은 정상이라고 확인했습니다(기종/펌웨어 미제공). 검증 기종 데이터는 변경하지 않았습니다.
- 호버 때 전체 배경을 지우고 중간 그리기 상태를 노출하던 방식 대신, 메모리에서 완성한 화면을 한 번에 표시합니다. 호버 갱신은 이전/새 버튼 영역으로 제한합니다.
- 최대화 버튼 및 최대화 시스템 명령을 제거합니다. 창 이동·최소화·작은 화면의 크기 조절/스크롤은 유지합니다.
- 화면/창 제목/PNG의 브랜드와 설명을 한 줄로 통합합니다: 한국어 LUMERIAN-LUMIX 사용 정보 / 영어 LUMERIAN-LUMIX Usage Info.
- 한국어 PNG의 미검증 상태가 영어로 남던 항목도 번역했습니다.
- 배포 전 같은 1.0.0의 파일을 다시 만들었습니다. 기종 데이터 갱신일은 2026-10-02 그대로입니다. 이미 받은 실행 파일/ZIP은 새 파일로 교체하세요.

확인: 전체 Go test/vet, 반복 호버의 국소 갱신 및 배경 삭제 방지, 최대화 차단, 메모리 렌더와 직접 렌더의 픽셀 일치, 한/영 화면·PNG·배율 검사를 수행했습니다. 사용자 PC에서 수정 빌드의 실제 화면 깜빡임 재확인은 필요합니다.

## 2026-10-03 배포 전 목록 및 버전 기록 정리

- 버전 기록 화면은 프로그램 버전과 릴리즈 날짜만 표시합니다. 정식 1.0.0의 릴리즈 날짜는 2026-10-03입니다. 기종 데이터 항목 및 개발 이력 카드는 화면에서 제외했습니다.
- 카메라 목록은 검증/미검증 모두 같은 카드 구조입니다. 정확한 모델·펌웨어 검증 조합을 먼저, 나머지는 ‘지원 추정’으로 표시합니다. 한 페이지에 7개씩 표시하며 현재 총 20개 항목은 3페이지입니다.
- 한국어 표기는 ‘확인: 이름’, 영어는 ‘Validated by 이름’입니다. ‘DC-L10: LUMIX Lab · 기종별 지원 펌웨어 필요’ 하단 문구와 별도의 작은 USB 기종 표는 제거했습니다. 연결 방법 안내는 시작 화면과 사용 안내에서 확인합니다.
- 신규 검증 데이터는 같은 모델의 지원 추정 행을 대체합니다. 여러 펌웨어가 검증되면 조합별 행을 유지합니다. 다른 펌웨어까지 검증한 것으로 간주하지 않습니다.
- 검증 참여자 페이지를 카메라 목록 페이지와 독립적으로 관리해 향후 참여자도 모두 확인할 수 있습니다. 잘못된 외부 데이터 안내는 카메라 목록에 표시합니다.
- 배포 전 수정으로 앱 버전은 1.0.0을 유지합니다. 카메라 데이터는 추가하지 않았으며 revision은 2026-10-02입니다. 이전 실행 파일 및 ZIP을 새 파일로 교체하세요.

확인: 전체 Go test/vet, 실제 Win32 입력 처리의 목록/참여자 페이지 이동, 신규 조합 추가·중복 후보 제거·펌웨어 검증 범위, 한/영 화면의 글자 크기 및 125/150/200% 렌더링. 물리 카메라 재시험은 수행하지 않았습니다.

## 2026-10-03 상태 색상 및 하단 정렬 개선

- 검증됨: 파란 카드와 밝은 파란 표시. 지원 추정: 차분한 회갈색 카드와 황갈색 표시. 두 상태의 문구와 검증 범위는 계속 함께 표시합니다. 통합 목록과 검증 우선 순서를 유지합니다.
- 버전을 상단 언어 메뉴 아래에서 좌측 하단으로 이동했습니다. 버전 클릭으로 버전 기록을 열 수 있습니다.
- 제작 정보의 중심을 논리 화면의 x=490으로 보정했습니다(기존 x=522.5). 버전/제작 정보/검증 상태의 기준선과 글자 크기를 통일했습니다. 하단 클릭 영역도 함께 이동했습니다.
- 기본 글꼴은 맑은 고딕입니다. 제목/정보/보조 문구의 위계는 유지하고 하단은 모두 14 논리 픽셀 크기로 통일했습니다. Windows 배율로 글자와 이미지, 클릭 위치가 함께 확대됩니다.
- 150%에서 제작 정보의 Special Thanks 및 내보내기 제목, 200%에서 홈 버튼 제목/연결 안내의 글자 영역 높이가 부족한 부분을 확인해 높이와 줄 간격을 보정했습니다.
- 한·영 전체 주요 화면을 100/125/150/175/200%로 렌더링하고 글자 크기, 잘림, 하단 정렬을 검사했습니다. 로컬 Go test/vet 및 숨김 네이티브 창의 키보드/페이지 이동 검사도 수행했습니다.
- 배포 전 1.0.0 수정본입니다. 릴리즈 날짜는 2026-10-03, 기종 데이터는 2026-10-02를 유지합니다. 실물 카메라 및 서로 다른 실제 모니터 간 이동을 새로 검증한 것은 아닙니다.

## 2026-10-03 제작자 / 도움을 주신 분들 50:50 배치

제작 정보 화면을 같은 너비의 두 카드로 나눴습니다. 왼쪽에는 제작자 제목·로고·계정·소개를, 오른쪽에는 검증 참여자와 Special Thanks / 카페 감사 문구를 배치했습니다. 향후 참여자가 늘면 오른쪽 명단을 페이지로 확인할 수 있습니다. 한·영 및 100/125/150/175/200% 렌더링으로 글자 배치와 잘림을 확인했습니다. 앱 버전과 릴리즈 날짜는 1.0.0 / 2026-10-03이며 검증 데이터는 2026-10-02 그대로입니다.

## 2026-10-03 제작자 Bieup_Hieut 캐릭터 반영

제작 정보는 좌우 50:50 배치를 유지합니다. 왼쪽 제작자 로고를 사용자가 제공한 `assets/Bieup_Hieut/Bieup_Hieut_In_App.png`의 3D 캐릭터로 교체했습니다. 원본은 변경 없이 branding 폴더에 보관하고, 네이티브 화면에서는 원래 비율을 유지한 BMP로 표시합니다. 상단 루메리안과 오른쪽 참여자 명단/카페 감사 문구는 기존 배치입니다. 한·영 및 100/125/150/175/200% 화면을 확인했습니다.

## 2026-10-03 연결 전/후 읽기 전용 안내

- 연결 전: 시작 화면의 연결 버튼 아래에 ‘읽기 전용 — 카메라의 설정과 데이터를 변경하지 않습니다.’를 표시합니다. 중복된 USB 모드 설명 한 줄을 이 안내로 정리했습니다.
- 연결 후: 카메라 모델/펌웨어/시리얼 아래에도 같은 문구를 표시합니다. 미검증 기종, 연결 해제 후 마지막 값 화면에도 유지합니다.
- 영어: “Read-only — camera settings and data are not changed.”
- 한·영 및 100/125/150/175/200% 렌더링으로 안내 위치와 잘림을 확인했습니다. 카메라 통신 명령은 기존 읽기 전용 구현을 유지합니다.
