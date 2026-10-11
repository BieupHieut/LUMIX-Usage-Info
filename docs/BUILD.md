# Build from source / 소스 빌드

## English

For normal use, download the [EXE from Releases](https://github.com/BieupHieut/LUMIX-Usage-Info/releases/latest). To build v1.1.0, install Go on Windows, open PowerShell in `source`, and run:

```powershell
$env:GO111MODULE='off'
$env:GOOS='windows'
$env:GOARCH='amd64'
$env:CGO_ENABLED='0'
go test .
go vet .
go build -trimpath -ldflags='-H=windowsgui -s -w' -o ../LUMERIAN_LUMIX_Usage_Board_v1.1.0.exe .
```

No third-party Go dependencies are needed. Artwork in `source/assets/` and `source/camera_profiles.json` are embedded in the EXE. External camera data is optional. Tests use synthetic data and do not establish camera compatibility.

When updating verification data, update both the root and source copies of `camera_profiles.json`. Schema 1 records exact model/firmware pairs with shutter and power/wake meanings confirmed. Omit `validation_date` when the test date is unknown.

## 한국어

일반 사용자는 [Releases의 EXE](https://github.com/BieupHieut/LUMIX-Usage-Info/releases/latest)를 받으면 됩니다. v1.1.0을 직접 빌드하려면 Windows에 Go를 설치하고 `source` 폴더에서 PowerShell을 열어 위 명령을 실행하세요.

외부 Go 의존성은 없습니다. `source/assets/`의 이미지와 `source/camera_profiles.json`은 EXE에 내장됩니다. 외부 기종 데이터 파일은 선택 사항입니다. 테스트는 합성 데이터를 사용하며 카메라 호환성을 입증하는 실기 검사는 아닙니다.

검증 데이터를 갱신할 때는 루트와 source의 `camera_profiles.json`을 함께 수정합니다. 스키마 1에는 셔터·전원/깨우기 의미가 확인된 정확한 모델·펌웨어 조합을 등록합니다. 실제 시험 날짜를 모르면 `validation_date`를 생략하세요.
