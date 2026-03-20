---
name: deploy
description: Use when releasing a new version of nw-cli — builds cross-platform binaries, creates GitHub release, and publishes npm packages. Triggers on "배포", "릴리스", "deploy", "release", "/deploy".
---

# nw-cli 수동 배포

로컬에서 크로스 플랫폼 빌드 → GitHub Release → npm 퍼블리시를 수행한다.

## 사전 조건

배포 전 다음을 확인한다:

1. `go test ./...` 전체 PASS
2. `go vet ./...` 통과
3. `go mod tidy` 후 변경 없음
4. 작업 디렉토리 clean (`git status` 확인)
5. 도구 설치: `goreleaser`, `gh` (GitHub CLI), `npm`
6. 인증 상태: `gh auth status`, `npm whoami`

## 배포 절차

### Step 1: 버전 결정

사용자에게 버전을 물어본다. SemVer 형식 (예: `0.1.0`, `0.2.0`, `1.0.0`).

### Step 2: 사전 검증

```bash
# 코드 검증
go mod tidy
go test ./... -count=1
go vet ./...

# 워킹 트리 clean 확인
git status

# 인증 확인
gh auth status
npm whoami
```

하나라도 실패하면 중단한다. `go mod tidy`가 파일을 변경하면 커밋한 뒤 진행한다.

### Step 3: Git 태그 생성 및 push

태그를 먼저 원격에 push하여 GitHub Release와 소스 커밋을 일치시킨다.

```bash
git tag v<VERSION>
git push origin v<VERSION>
```

### Step 4: goreleaser로 크로스 플랫폼 빌드

```bash
goreleaser release --clean --skip=publish
```

`dist/` 디렉토리에 5개 플랫폼 아카이브가 생성된다:
- `nw-cli_<VERSION>_linux_amd64.tar.gz`
- `nw-cli_<VERSION>_linux_arm64.tar.gz`
- `nw-cli_<VERSION>_darwin_amd64.tar.gz`
- `nw-cli_<VERSION>_darwin_arm64.tar.gz`
- `nw-cli_<VERSION>_windows_amd64.zip`
- `checksums.txt`

### Step 5: GitHub Release 생성

`--verify-tag`로 원격 태그와 일치하는지 확인한다.

```bash
gh release create v<VERSION> dist/nw-cli_*.tar.gz dist/nw-cli_*.zip dist/checksums.txt \
  --verify-tag \
  --title "v<VERSION>" \
  --generate-notes
```

### Step 6: npm 패키지 빌드

```bash
./npm/build-npm.sh <VERSION> dist
```

> 주의: `build-npm.sh`는 `npm/cli/package.json`의 버전을 수정한다. 배포 후 Step 9에서 정리한다.

### Step 7: npm 퍼블리시

플랫폼 패키지를 먼저, 메인 패키지를 마지막에 퍼블리시한다:

```bash
for dir in npm/linux-x64 npm/linux-arm64 npm/darwin-x64 npm/darwin-arm64 npm/win32-x64; do
  if [ -f "$dir/nw-cli" ] || [ -f "$dir/nw-cli.exe" ]; then
    (cd "$dir" && npm publish --access public)
  fi
done
(cd npm/cli && npm publish --access public)
```

### Step 8: 검증

```bash
# GitHub Release 확인
gh release view v<VERSION>

# npm 메인 패키지 확인
npm view nw-cli version

# 현재 플랫폼 패키지 확인
npm view @nw-cli/linux-x64 version   # Linux 예시

# 설치 테스트
npx nw-cli@<VERSION> version
```

### Step 9: 배포 후 정리

`build-npm.sh`가 수정한 `npm/cli/package.json`을 원래대로 되돌린다:

```bash
git checkout -- npm/cli/package.json
rm -rf dist/ npm/*/nw-cli npm/*/nw-cli.exe npm/*/package.json
```

또는 변경된 버전을 그대로 커밋할 수도 있다:

```bash
git add npm/cli/package.json
git commit -m "chore: npm 패키지 버전을 v<VERSION>으로 업데이트"
```

## 롤백

문제 발생 시:

```bash
# GitHub Release 삭제
gh release delete v<VERSION> --yes

# npm unpublish (72시간 이내만 가능)
npm unpublish nw-cli@<VERSION>
for pkg in linux-x64 linux-arm64 darwin-x64 darwin-arm64 win32-x64; do
  npm unpublish @nw-cli/$pkg@<VERSION>
done

# Git 태그 삭제
git tag -d v<VERSION>
git push origin :refs/tags/v<VERSION>
```

> **중요**: npm은 unpublish 후 같은 버전 번호를 재사용할 수 없다. 롤백 후 재배포 시 반드시 새 버전을 사용해야 한다 (예: `0.1.0` → `0.1.1`).

## goreleaser 미설치 시

goreleaser 없이 수동 빌드:

```bash
VERSION=<VERSION>
COMMIT=$(git rev-parse --short HEAD)
DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS="-s -w -X github.com/physics91/naverworks-cli/cmd.version=$VERSION -X github.com/physics91/naverworks-cli/cmd.commit=$COMMIT -X github.com/physics91/naverworks-cli/cmd.buildDate=$DATE"

mkdir -p dist

# Linux amd64
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "$LDFLAGS" -o dist/nw-cli . && \
  tar -czf "dist/nw-cli_${VERSION}_linux_amd64.tar.gz" -C dist nw-cli && rm dist/nw-cli

# Linux arm64
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags "$LDFLAGS" -o dist/nw-cli . && \
  tar -czf "dist/nw-cli_${VERSION}_linux_arm64.tar.gz" -C dist nw-cli && rm dist/nw-cli

# macOS amd64
GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "$LDFLAGS" -o dist/nw-cli . && \
  tar -czf "dist/nw-cli_${VERSION}_darwin_amd64.tar.gz" -C dist nw-cli && rm dist/nw-cli

# macOS arm64
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags "$LDFLAGS" -o dist/nw-cli . && \
  tar -czf "dist/nw-cli_${VERSION}_darwin_arm64.tar.gz" -C dist nw-cli && rm dist/nw-cli

# Windows amd64
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "$LDFLAGS" -o dist/nw-cli.exe . && \
  (cd dist && zip "nw-cli_${VERSION}_windows_amd64.zip" nw-cli.exe && rm nw-cli.exe)

# 체크섬 (macOS 호환)
(cd dist && (sha256sum nw-cli_*.tar.gz nw-cli_*.zip 2>/dev/null || shasum -a 256 nw-cli_*.tar.gz nw-cli_*.zip) > checksums.txt)
```

이후 Step 5부터 동일하게 진행한다.
