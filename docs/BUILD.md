# Build from source / 소스 빌드

일반 사용자는 [EXE](https://github.com/BieupHieut/LUMIX-Usage-Info/releases/latest)만 받으면 됩니다.
이 문서는 직접 소스를 빌드하려는 개발자를 위한 안내입니다.

For normal use, download the EXE from Releases. This guide is for developers building the source.

## Windows x64

Go가 설치된 Windows에서 `source` 폴더를 열고 PowerShell로 실행합니다.
Run these commands in the `source` directory on Windows with Go installed.

```powershell
$env:GO111MODULE='off'
$env:GOOS='windows'
$env:GOARCH='amd64'
$env:CGO_ENABLED='0'
go test .
go vet .
go build -trimpath -ldflags='-H=windowsgui -s -w' -o ../LUMIX_Usage_Info_v1.0.0.exe .
```

외부 Go 의존성은 없습니다. `source/assets/`와 `source/camera_profiles.json`이 EXE에 포함됩니다.
별도의 JSON은 최신 검증 데이터를 덮어 적용할 때만 필요합니다.

No third-party Go dependencies. Assets and camera data are embedded in the EXE; external JSON is optional.

## Maintaining verification data / 검증 데이터 관리

Update both the root `camera_profiles.json` and `source/camera_profiles.json` for source builds.
Schema 1 holds exact model/firmware pairs with both shutter and power/wake meanings confirmed.
Omit `validation_date` if the actual test date is unknown.

소스를 빌드할 때 루트와 `source/camera_profiles.json`을 함께 갱신합니다.
스키마 1에는 셔터·전원 카운터 의미가 모두 확인된 정확한 모델·펌웨어 조합을 등록합니다.
실제 검증 날짜를 모르면 `validation_date`를 생략합니다.
