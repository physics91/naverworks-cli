---
name: build
description: Use when building nw-cli binaries — single platform or cross-platform builds with version injection. Triggers on "빌드", "build", "/build".
---

# nw-cli 빌드

이 스킬은 AI 에이전트가 직접 실행한다. 로컬 또는 크로스 플랫폼 바이너리를 빌드한다.

## 실행 규칙

1. 모든 명령을 직접 Bash로 실행한다.
2. 각 명령을 개별로 실행하고 종료 코드를 확인한다.
3. 인자가 없으면 현재 플랫폼용 빌드, `all`이면 5개 플랫폼 크로스 빌드.

## 절차

### Phase 0: 인자 확인

- 인자 없음 또는 `local`: 현재 플랫폼 빌드 → Phase 2a
- `all`: 5개 플랫폼 크로스 빌드 → Phase 2b

### Phase 1: 사전 검증

```bash
go vet ./...
```
```bash
go test ./... -count=1
```

하나라도 실패하면 중단한다.

### Phase 2a: 로컬 빌드 (기본)

```bash
make build
```
```bash
ls -lh nw-cli
```
```bash
./nw-cli version
```

### Phase 2b: 크로스 빌드 (all)

버전 추출 — `v` prefix를 제거하여 goreleaser/deploy와 일치시킨다:

```bash
git describe --tags --abbrev=0 2>/dev/null
```
→ 태그가 있으면 `v` prefix를 제거하여 VERSION에 저장 (예: `v0.1.0` → `0.1.0`)
→ 태그가 없으면 `dev`를 사용

goreleaser 확인:
```bash
command -v goreleaser
```

**goreleaser가 있으면:**
```bash
goreleaser release --clean --skip=publish --snapshot
```

**goreleaser가 없으면** 각 플랫폼을 개별 명령으로 빌드:

```bash
COMMIT=$(git rev-parse --short HEAD)
```
```bash
DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
```
```bash
mkdir -p dist
```

각 플랫폼별 빌드 + 아카이브 (deploy/npm과 동일한 산출물 형식):

linux-amd64:
```bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "$LDFLAGS" -o dist/nw-cli .
```
```bash
tar -czf "dist/nw-cli_${VERSION}_linux_amd64.tar.gz" -C dist nw-cli
```
```bash
rm dist/nw-cli
```

linux-arm64:
```bash
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags "$LDFLAGS" -o dist/nw-cli .
```
```bash
tar -czf "dist/nw-cli_${VERSION}_linux_arm64.tar.gz" -C dist nw-cli
```
```bash
rm dist/nw-cli
```

darwin-amd64:
```bash
GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "$LDFLAGS" -o dist/nw-cli .
```
```bash
tar -czf "dist/nw-cli_${VERSION}_darwin_amd64.tar.gz" -C dist nw-cli
```
```bash
rm dist/nw-cli
```

darwin-arm64:
```bash
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags "$LDFLAGS" -o dist/nw-cli .
```
```bash
tar -czf "dist/nw-cli_${VERSION}_darwin_arm64.tar.gz" -C dist nw-cli
```
```bash
rm dist/nw-cli
```

windows-amd64:
```bash
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "$LDFLAGS" -o dist/nw-cli.exe .
```
```bash
zip dist/nw-cli_${VERSION}_windows_amd64.zip -j dist/nw-cli.exe
```
```bash
rm dist/nw-cli.exe
```

> LDFLAGS는 Phase 2b 시작 시 아래와 같이 구성한다:
> `-s -w -X github.com/physics91/naverworks-cli/cmd.version=$VERSION -X github.com/physics91/naverworks-cli/cmd.commit=$COMMIT -X github.com/physics91/naverworks-cli/cmd.buildDate=$DATE`

### 보고

로컬 빌드:
```
빌드 완료:
  nw-cli (6.1MB)
  버전: {"version":"dev","commit":"abc1234","build_date":"..."}
```

크로스 빌드:
```
빌드 완료: 5개 플랫폼

  dist/nw-cli_0.1.0_linux_amd64.tar.gz     2.6MB
  dist/nw-cli_0.1.0_linux_arm64.tar.gz     2.4MB
  dist/nw-cli_0.1.0_darwin_amd64.tar.gz    2.6MB
  dist/nw-cli_0.1.0_darwin_arm64.tar.gz    2.5MB
  dist/nw-cli_0.1.0_windows_amd64.zip      2.7MB
```
