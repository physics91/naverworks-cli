---
name: deploy
description: Use when releasing a new version of nw-cli — autonomously builds cross-platform binaries, creates GitHub release, and publishes npm packages. Triggers on "배포", "릴리스", "deploy", "release", "/deploy".
---

# nw-cli 자동 배포

이 스킬은 AI 에이전트가 직접 실행한다. 안내가 아니라 **각 단계를 순서대로 실행**하고, 실패 시 즉시 중단하여 사용자에게 보고한다.

## 실행 규칙

1. **모든 단계를 직접 Bash로 실행한다.** 사용자에게 "이 명령을 실행하세요"라고 안내하지 않는다.
2. 각 단계 실행 후 결과를 확인한다. 실패하면 즉시 중단하고 에러를 보고한다.
3. 사용자 입력이 필요한 것은 **버전 번호**뿐이다. 인자로 주어지지 않으면 물어본다.
4. 위험한 작업(태그 push, npm publish) 전에 사용자에게 확인을 받는다.

## 절차

### Phase 0: 버전 확인

사용자가 인자로 버전을 제공했으면 사용한다. 없으면 묻는다:
- SemVer 형식 (예: `0.1.0`)
- `v` 접두사는 자동으로 붙인다

### Phase 1: 사전 검증 (자동 실행)

아래를 순서대로 실행한다. 하나라도 실패하면 **즉시 중단**:

```
1. go mod tidy
2. git diff --exit-code  (go mod tidy가 뭔가 바꿨으면 커밋하고 계속)
3. go test ./... -count=1
4. go vet ./...
5. git status --porcelain  (비어있지 않으면 중단)
6. goreleaser --version || go version  (goreleaser 없으면 수동 빌드 모드)
7. gh auth status
8. npm whoami
```

각 결과를 간결하게 보고:
```
✓ go mod tidy — 변경 없음
✓ go test — 4 패키지 PASS
✓ go vet — 이상 없음
✓ git status — clean
✓ goreleaser 1.x.x
✓ gh — physics91 인증됨
✓ npm — physics91 인증됨
```

### Phase 2: 빌드 (자동 실행)

goreleaser가 있으면:
```bash
goreleaser release --clean --skip=publish
```

없으면 수동 크로스 빌드:
```bash
VERSION=<VERSION>
COMMIT=$(git rev-parse --short HEAD)
DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS="-s -w -X github.com/physics91/naverworks-cli/cmd.version=$VERSION -X github.com/physics91/naverworks-cli/cmd.commit=$COMMIT -X github.com/physics91/naverworks-cli/cmd.buildDate=$DATE"

mkdir -p dist

GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "$LDFLAGS" -o dist/nw-cli . && \
  tar -czf "dist/nw-cli_${VERSION}_linux_amd64.tar.gz" -C dist nw-cli && rm dist/nw-cli

GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags "$LDFLAGS" -o dist/nw-cli . && \
  tar -czf "dist/nw-cli_${VERSION}_linux_arm64.tar.gz" -C dist nw-cli && rm dist/nw-cli

GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "$LDFLAGS" -o dist/nw-cli . && \
  tar -czf "dist/nw-cli_${VERSION}_darwin_amd64.tar.gz" -C dist nw-cli && rm dist/nw-cli

GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags "$LDFLAGS" -o dist/nw-cli . && \
  tar -czf "dist/nw-cli_${VERSION}_darwin_arm64.tar.gz" -C dist nw-cli && rm dist/nw-cli

GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "$LDFLAGS" -o dist/nw-cli.exe . && \
  (cd dist && zip "nw-cli_${VERSION}_windows_amd64.zip" nw-cli.exe && rm nw-cli.exe)

(cd dist && (sha256sum nw-cli_*.tar.gz nw-cli_*.zip 2>/dev/null || shasum -a 256 nw-cli_*.tar.gz nw-cli_*.zip) > checksums.txt)
```

빌드 완료 후 산출물 목록을 보고:
```
✓ dist/nw-cli_0.1.0_linux_amd64.tar.gz (6.1MB)
✓ dist/nw-cli_0.1.0_linux_arm64.tar.gz (5.9MB)
✓ dist/nw-cli_0.1.0_darwin_amd64.tar.gz (6.3MB)
✓ dist/nw-cli_0.1.0_darwin_arm64.tar.gz (6.0MB)
✓ dist/nw-cli_0.1.0_windows_amd64.zip (6.2MB)
✓ dist/checksums.txt
```

### Phase 3: 태그 + GitHub Release (확인 후 실행)

**사용자에게 확인**: "v<VERSION> 태그를 push하고 GitHub Release를 생성합니다. 진행할까요?"

승인 후 실행:
```bash
git tag v<VERSION>
git push origin v<VERSION>
gh release create v<VERSION> dist/nw-cli_*.tar.gz dist/nw-cli_*.zip dist/checksums.txt \
  --verify-tag \
  --title "v<VERSION>" \
  --generate-notes
```

결과에서 Release URL을 출력한다.

### Phase 4: npm 퍼블리시 (확인 후 실행)

**사용자에게 확인**: "npm에 퍼블리시합니다. 진행할까요?"

승인 후 실행:
```bash
./npm/build-npm.sh <VERSION> dist

for dir in npm/linux-x64 npm/linux-arm64 npm/darwin-x64 npm/darwin-arm64 npm/win32-x64; do
  if [ -f "$dir/nw-cli" ] || [ -f "$dir/nw-cli.exe" ]; then
    (cd "$dir" && npm publish --access public)
  fi
done
(cd npm/cli && npm publish --access public)
```

### Phase 5: 검증 (자동 실행)

```bash
gh release view v<VERSION>
npm view nw-cli version
npm view @nw-cli/linux-x64 version
```

### Phase 6: 정리 (자동 실행)

```bash
git checkout -- npm/cli/package.json
rm -rf dist/ npm/*/nw-cli npm/*/nw-cli.exe
# 플랫폼별 생성된 package.json 정리
for dir in npm/linux-x64 npm/linux-arm64 npm/darwin-x64 npm/darwin-arm64 npm/win32-x64; do
  rm -f "$dir/package.json"
done
```

### 최종 보고

```
배포 완료: v<VERSION>

GitHub Release: https://github.com/physics91/naverworks-cli/releases/tag/v<VERSION>
npm: https://www.npmjs.com/package/nw-cli/v/<VERSION>

설치:
  npm install -g nw-cli@<VERSION>
  npx nw-cli@<VERSION> version
```

## 실패 시 롤백

어느 단계에서든 실패하면, 이미 실행된 단계만 역순으로 롤백한다:

| 실패 지점 | 롤백 범위 |
|-----------|----------|
| Phase 1 (검증) | 없음 |
| Phase 2 (빌드) | `rm -rf dist/` |
| Phase 3 (태그/릴리스) | `gh release delete v<VERSION> --yes && git push origin :refs/tags/v<VERSION> && git tag -d v<VERSION>` |
| Phase 4 (npm) | 위 + `npm unpublish nw-cli@<VERSION>` + 각 플랫폼 패키지 unpublish |

> **중요**: npm unpublish 후 같은 버전 번호 재사용 불가. 롤백 후에는 새 버전을 사용해야 한다.

롤백도 **에이전트가 직접 실행**한다. 사용자에게 명령을 안내하지 않는다.
