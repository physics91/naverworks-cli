---
name: deploy
description: Use when releasing a new version of naverworks — autonomously builds cross-platform binaries, creates GitHub release, and publishes npm packages. Triggers on "배포", "릴리스", "deploy", "release", "/deploy".
---

# naverworks 자동 배포

이 스킬은 AI 에이전트가 직접 실행한다. 안내가 아니라 **각 단계를 순서대로 실행**하고, 실패 시 즉시 중단하여 롤백 후 사용자에게 보고한다.

## 실행 규칙

1. **모든 단계를 직접 Bash로 실행한다.** 사용자에게 "이 명령을 실행하세요"라고 안내하지 않는다.
2. 각 명령은 개별로 실행하고, 종료 코드를 확인한다. **0이 아니면 즉시 중단**하고 롤백으로 진입한다.
3. 사용자 입력이 필요한 것은 **버전 번호**뿐이다. 인자로 주어지지 않으면 물어본다.
4. 위험한 작업(태그 push, npm publish) 전에 사용자에게 확인을 받는다.
5. **상태 추적**: 에이전트는 내부적으로 `COMPLETED_PHASES` 목록과 `PUBLISHED_PACKAGES` 목록을 유지한다. 롤백 시 이 목록만 역순 정리한다.

## 절차

### Phase 0: 버전 확인

사용자가 인자로 버전을 제공했으면 사용한다. 없으면 묻는다:
- SemVer 형식 (예: `0.1.0`)
- `v` 접두사는 자동으로 붙인다

### Phase 1: 사전 검증 (자동 실행)

아래를 **개별 명령으로** 순서대로 실행한다. 하나라도 실패하면 **즉시 중단** (롤백 불필요):

```
1. go mod tidy
2. git diff --exit-code
   → 종료 코드 != 0이면 "go.mod/go.sum이 변경되었습니다. 먼저 커밋해주세요"로 중단
3. go test ./... -count=1
4. go vet ./...
5. git status --porcelain
   → 출력이 있으면 "워킹 트리가 clean하지 않습니다"로 중단
6. command -v goreleaser
   → 실패하면 HAS_GORELEASER=false로 표시 (중단하지 않음)
7. command -v tar
8. command -v zip
9. gh auth status
10. npm whoami
```

각 결과를 간결하게 보고:
```
✓ go mod tidy — 변경 없음
✓ go test — 4 패키지 PASS
✓ go vet — 이상 없음
✓ git status — clean
✓ goreleaser 1.x.x (또는 ✗ goreleaser 미설치 → 수동 빌드 모드)
✓ gh — physics91 인증됨
✓ npm — physics91 인증됨
```

### Phase 2: 태그 생성 (로컬만)

goreleaser가 태그를 요구하므로 **빌드 전에 로컬 태그를 생성**한다. push는 아직 하지 않는다.

```bash
git tag v<VERSION>
```

### Phase 3: 빌드 (자동 실행)

`HAS_GORELEASER=true`이면:
```bash
goreleaser release --clean --skip=publish
```

`HAS_GORELEASER=false`이면 (각 명령을 개별 실행, 실패 시 즉시 중단):
```bash
VERSION=<VERSION>
COMMIT=$(git rev-parse --short HEAD)
DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS="-s -w -X github.com/physics91/naverworks-cli/cmd.version=$VERSION -X github.com/physics91/naverworks-cli/cmd.commit=$COMMIT -X github.com/physics91/naverworks-cli/cmd.buildDate=$DATE"

mkdir -p dist
```

그 다음 **각 플랫폼을 개별 명령으로** 실행:
각 플랫폼을 **개별 명령 3개씩** 실행 (linux-amd64 예시):
```bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "$LDFLAGS" -o dist/naverworks .
```
```bash
tar -czf "dist/naverworks_${VERSION}_linux_amd64.tar.gz" -C dist naverworks
```
```bash
rm dist/naverworks
```

linux-arm64, darwin-amd64, darwin-arm64도 **동일하게 3개 명령으로 분리**한다.

Windows:
```bash
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "$LDFLAGS" -o dist/naverworks.exe .
```
```bash
zip dist/naverworks_${VERSION}_windows_amd64.zip -j dist/naverworks.exe
```
```bash
rm dist/naverworks.exe
```

체크섬 (sha256sum 또는 shasum 중 사용 가능한 것):
```bash
sha256sum dist/naverworks_*.tar.gz dist/naverworks_*.zip > dist/checksums.txt
```
macOS에서 sha256sum이 없으면:
```bash
shasum -a 256 dist/naverworks_*.tar.gz dist/naverworks_*.zip > dist/checksums.txt
```

빌드 완료 후 산출물 목록과 크기를 보고한다.

### Phase 4: 태그 push + GitHub Release (확인 후 실행)

**사용자에게 확인**: "v<VERSION> 태그를 push하고 GitHub Release를 생성합니다. 진행할까요?"

승인 후 **개별 명령으로** 실행:
```bash
git push origin v<VERSION>
```
```bash
gh release create v<VERSION> dist/naverworks_*.tar.gz dist/naverworks_*.zip dist/checksums.txt \
  --verify-tag \
  --title "v<VERSION>" \
  --generate-notes
```

결과에서 Release URL을 출력한다.

### Phase 5: npm 퍼블리시 (확인 후 실행)

**사용자에게 확인**: "npm에 퍼블리시합니다. 진행할까요?"

승인 후 실행:
```bash
./npm/build-npm.sh <VERSION> dist
```

플랫폼 패키지를 **하나씩 개별 명령으로** 퍼블리시하고, 성공한 패키지를 `PUBLISHED_PACKAGES`에 기록:

```bash
cd npm/linux-x64
```
```bash
npm publish --access public
```
```bash
cd ../..
```
→ 성공 시 PUBLISHED_PACKAGES에 `@physics91/linux-x64` 추가

linux-arm64, darwin-x64, darwin-arm64, win32-x64도 **동일하게 3개 명령으로 분리**한다.

마지막에 메인 패키지:
```bash
cd npm/cli
```
```bash
npm publish --access public
```
```bash
cd ../..
```
→ 성공 시 PUBLISHED_PACKAGES에 `naverworks` 추가

**어느 패키지에서든 실패하면** 즉시 중단하고 롤백으로 진입한다.

### Phase 6: 검증 (자동 실행)

`PUBLISHED_PACKAGES`의 모든 패키지를 검증:
```bash
gh release view v<VERSION>
npm view naverworks version
npm view @physics91/linux-x64 version
npm view @physics91/linux-arm64 version
npm view @physics91/darwin-x64 version
npm view @physics91/darwin-arm64 version
npm view @physics91/win32-x64 version
```

### Phase 7: 정리 (자동 실행)

```bash
git checkout -- npm/cli/package.json
rm -rf dist/
for dir in npm/linux-x64 npm/linux-arm64 npm/darwin-x64 npm/darwin-arm64 npm/win32-x64; do
  rm -f "$dir/naverworks" "$dir/naverworks.exe" "$dir/package.json"
done
```

### 최종 보고

```
배포 완료: v<VERSION>

GitHub Release: https://github.com/physics91/naverworks-cli/releases/tag/v<VERSION>
npm: https://www.npmjs.com/package/naverworks/v/<VERSION>

설치:
  npm install -g naverworks@<VERSION>
  npx naverworks@<VERSION> version
```

## 실패 시 롤백

에이전트는 `COMPLETED_PHASES`와 `PUBLISHED_PACKAGES`를 기반으로 **실제로 완료된 작업만** 역순으로 정리한다.

### 롤백 로직 (에이전트가 조건부 실행)

```
if "npm" in COMPLETED_PHASES:
    for pkg in reversed(PUBLISHED_PACKAGES):
        npm unpublish <pkg>@<VERSION>  # 실패해도 계속
    git checkout -- npm/cli/package.json

if "github_release" in COMPLETED_PHASES:
    gh release delete v<VERSION> --yes  # 실패해도 계속

if "tag_pushed" in COMPLETED_PHASES:
    git push origin :refs/tags/v<VERSION>  # 실패해도 계속

# 로컬 태그는 항상 정리
git tag -d v<VERSION> 2>/dev/null

# 빌드 산출물 정리
rm -rf dist/
for dir in npm/linux-x64 npm/linux-arm64 npm/darwin-x64 npm/darwin-arm64 npm/win32-x64; do
  rm -f "$dir/naverworks" "$dir/naverworks.exe" "$dir/package.json"
done
```

> **중요**: npm unpublish 후 같은 버전 번호 재사용 불가. 롤백 후에는 새 버전을 사용해야 한다.

롤백 각 명령은 **실패해도 다음 정리를 계속**한다 (정리 작업은 best-effort).

롤백 완료 후 사용자에게 무엇이 정리되었고 무엇이 실패했는지 보고한다.
