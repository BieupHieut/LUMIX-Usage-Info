# Camera data updates / 검증 데이터 업데이트

## English

Current verification data is embedded in the EXE. **The reader runs without a JSON file.**
Use the optional `camera_profiles.json` only to apply newer verification data separately.

### Apply an update

1. Download [camera_profiles.json](../camera_profiles.json) from this repository.
2. Place it in the same folder as the EXE.
3. Close and restart the reader.

If the file is missing or invalid, the reader uses its embedded list. If an external-data warning appears, replace the file with the version from this repository or remove it and restart.
Updates are not downloaded automatically. The file updates the verification list, not the numbers read from the camera.

### Versions

Software changes use application versions; camera verification additions use data revision dates.
Current data revision: **2026-10-02**. A data publication date is not necessarily the camera test date.

## 한국어

현재 검증 데이터는 EXE 안에 포함돼 있습니다. **JSON 없이도 실행할 수 있습니다.**
새 검증 데이터만 따로 적용할 때 선택 파일인 `camera_profiles.json`을 사용합니다.

### 업데이트 적용

1. 이 저장소의 [camera_profiles.json](../camera_profiles.json)을 내려받습니다.
2. EXE와 같은 폴더에 놓습니다.
3. 프로그램을 종료하고 다시 실행합니다.

파일이 없거나 형식이 잘못되면 내장 목록을 사용합니다. 외부 데이터 경고가 뜨면 이 저장소의 파일로 교체하거나 파일을 제거한 뒤 다시 실행하세요.
업데이트를 자동으로 내려받지는 않습니다. 이 파일은 검증 목록을 갱신하며 카메라에서 읽은 숫자는 변경하지 않습니다.

### 버전

프로그램 변경은 앱 버전으로, 기종 검증 추가는 데이터 날짜로 관리합니다.
현재 데이터 갱신일은 **2026-10-02**입니다. 데이터 발행 날짜와 실제 카메라 검증 날짜는 다를 수 있습니다.
