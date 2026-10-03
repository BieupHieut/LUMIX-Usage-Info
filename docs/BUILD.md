# 소스 빌드 / Build from source

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
