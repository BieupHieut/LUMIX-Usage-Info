# GitHub 업로드 안내 — v1.0.0

## 1. 소스 업로드

업로드 대상은 `repository` 폴더의 내용입니다. 최상위에 README.md, LICENSE, camera_profiles.json, source/, docs/, branding/, .github/가 보이게 올립니다. 작업 폴더 전체나 이전 버전 폴더를 올리지 마세요.

이 폴더는 로컬 Git 저장소로 준비하며 main 브랜치를 사용합니다. 아직 커밋이나 원격 업로드를 수행하지 않았습니다. 커밋 작성자 이름/이메일과 업로드할 저장소 주소는 사용자 정보로 설정해야 합니다.

새 빈 저장소에 올릴 경우 PowerShell에서 다음 순서로 진행합니다. 기존 저장소라면 해당 저장소를 clone한 뒤 이 소스를 반영해 이력을 유지하세요. 기존 이력을 강제로 덮어쓰지 않습니다.

```powershell
# repository 폴더에서 실행. 실제 본인 정보와 저장소 주소로 바꿉니다.
git config user.name 'YOUR_NAME'
git config user.email 'YOUR_GITHUB_EMAIL_OR_NOREPLY_EMAIL'
git add .
git commit -m 'Release LUMIX Usage Info v1.0.0'
git remote add origin 'https://github.com/YOUR_ACCOUNT/YOUR_REPOSITORY.git'
git push -u origin main
```

웹 업로드를 사용하면 .gitignore는 자동 제외 기능으로 작동하지 않습니다. 이 업로드용 소스 폴더에는 EXE/ZIP을 실제로 제외했고, .github와 .gitignore도 포함해야 합니다. .git 폴더는 올리지 않습니다.

## 2. 사용자 다운로드용 릴리즈

GitHub 저장소의 Releases에서 새 릴리즈를 작성합니다.

- 태그: `v1.0.0`
- 대상: 업로드한 소스가 있는 main 브랜치
- 제목: `LUMERIAN-LUMIX Usage Info v1.0.0`
- 본문: 별도 `RELEASE_DESCRIPTION_v1.0.0.md`의 내용을 붙여 넣습니다.
- 첨부: `release-assets`의 Windows x64 ZIP, 실행 파일, camera_profiles.json, SHA256SUMS.txt

사용자에게는 Windows x64 ZIP을 우선 안내하세요. EXE를 단독으로 받는 경우 JSON도 같은 폴더에 둡니다. 릴리즈 본문에 첨부한 실행 파일과 소스 버전이 같은지 확인합니다.

## 3. 기종 데이터 업데이트

기종 검증만 추가하면 앱 버전은 유지하고 camera_profiles.json의 revision 날짜를 갱신합니다. 정확한 모델/펌웨어/확인자를 등록하고 실제 검증 날짜를 모르면 생략합니다. root와 source의 JSON 복사본을 함께 갱신합니다. 앱/UI/버그 수정은 이후 프로그램 버전으로 관리합니다.

## 공식 참고

- [소스 파일 업로드](https://docs.github.com/en/repositories/working-with-files/managing-files/adding-a-file-to-a-repository)
- [릴리즈 작성](https://docs.github.com/en/repositories/releasing-projects-on-github/managing-releases-in-a-repository)

이 자료는 업로드 준비본입니다. 원격 저장소 생성/커밋/태그/게시를 자동 수행하지 않습니다.
