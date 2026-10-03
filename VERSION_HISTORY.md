# Version History / 버전 기록

## v1.0.0 — 2026-10-03

Initial stable release / 첫 정식 버전.

- Read-only camera information, shutter/power counts and error history.
- PNG / TXT / clipboard exports with serial privacy options.
- Korean / English, Windows language detection, Lumerian design.
- Verified-first unified camera list, estimated support labels and pagination.
- Buffered rendering, resizing, DPI and keyboard navigation.

배포 전 1.0.0의 화면 수정 내용을 포함합니다. 이전 로컬 실행 파일과 ZIP은 최신 파일로 교체하세요.

[Development archive / 개발 이력 보관](docs/development_archive/VERSION_HISTORY.md)

## 2026-10-03 상태 색상 및 하단 정렬 개선

- 검증됨: 파란 카드와 밝은 파란 표시. 지원 추정: 차분한 회갈색 카드와 황갈색 표시. 두 상태의 문구와 검증 범위는 계속 함께 표시합니다. 통합 목록과 검증 우선 순서를 유지합니다.
- 버전을 상단 언어 메뉴 아래에서 좌측 하단으로 이동했습니다. 버전 클릭으로 버전 기록을 열 수 있습니다.
- 제작 정보의 중심을 논리 화면의 x=490으로 보정했습니다(기존 x=522.5). 버전/제작 정보/검증 상태의 기준선과 글자 크기를 통일했습니다. 하단 클릭 영역도 함께 이동했습니다.
- 기본 글꼴은 맑은 고딕입니다. 제목/정보/보조 문구의 위계는 유지하고 하단은 모두 14 논리 픽셀 크기로 통일했습니다. Windows 배율로 글자와 이미지, 클릭 위치가 함께 확대됩니다.
- 150%에서 제작 정보의 Special Thanks 및 내보내기 제목, 200%에서 홈 버튼 제목/연결 안내의 글자 영역 높이가 부족한 부분을 확인해 높이와 줄 간격을 보정했습니다.
- 한·영 전체 주요 화면을 100/125/150/175/200%로 렌더링하고 글자 크기, 잘림, 하단 정렬을 검사했습니다. 로컬 Go test/vet 및 숨김 네이티브 창의 키보드/페이지 이동 검사도 수행했습니다.
- 배포 전 1.0.0 수정본입니다. 릴리즈 날짜는 2026-10-03, 기종 데이터는 2026-10-02를 유지합니다. 실물 카메라 및 서로 다른 실제 모니터 간 이동을 새로 검증한 것은 아닙니다.
