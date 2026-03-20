---
name: build
description: Use when building nw-cli binaries — single platform or cross-platform builds with version injection. Triggers on "빌드", "build", "/build".
---

# nw-cli 빌드

이 스킬은 AI 에이전트가 직접 실행한다. 로컬 또는 크로스 플랫폼 바이너리를 빌드한다.

## 실행 규칙

1. 모든 명령을 직접 Bash로 실행한다.
2. 인자가 없으면 현재 플랫폼용 빌드, `--all`이면 5개 플랫폼 크로스 빌드.
3. 빌드 결과를 바이너리 크기와 함께 보고한다.

## 절차

### Phase 0: 인자 확인

- 인자 없음 또는 `local`: 현재 플랫폼 빌드
- `all` 또는 `--all`: 5개 플랫폼 크로스 빌드
- 특정 플랫폼: `linux-amd64`, `darwin-arm64` 등

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

### Phase 2b: 크로스 빌드 (--all)

버전은 git 태그에서 추출하거나 `dev`를 사용:
```bash
git describe --tags --abbrev=0 2>/dev/null || echo "dev"
```

goreleaser가 있으면:
```bash
goreleaser release --clean --skip=publish --snapshot
```

없으면 각 플랫폼을 개별 명령으로 빌드:

```bash
VERSION=<추출된 버전>
COMMIT=$(git rev-parse --short HEAD)
DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS="-s -w -X github.com/physics91/naverworks-cli/cmd.version=$VERSION -X github.com/physics91/naverworks-cli/cmd.commit=$COMMIT -X github.com/physics91/naverworks-cli/cmd.buildDate=$DATE"
```
```bash
mkdir -p dist
```

각 플랫폼별 **개별 명령**으로 빌드:

linux-amd64:
```bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "$LDFLAGS" -o dist/nw-cli_linux_amd64 .
```

linux-arm64:
```bash
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags "$LDFLAGS" -o dist/nw-cli_linux_arm64 .
```

darwin-amd64:
```bash
GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "$LDFLAGS" -o dist/nw-cli_darwin_amd64 .
```

darwin-arm64:
```bash
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags "$LDFLAGS" -o dist/nw-cli_darwin_arm64 .
```

windows-amd64:
```bash
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "$LDFLAGS" -o dist/nw-cli_windows_amd64.exe .
```

### 보고

로컬 빌드:
```
빌드 완료:
  nw-cli (6.1MB)
  버전: {"version":"dev","commit":"abc1234","build_date":"2026-03-20T..."}
```

크로스 빌드:
```
빌드 완료: 5개 플랫폼

  dist/nw-cli_linux_amd64     6.0MB
  dist/nw-cli_linux_arm64     5.8MB
  dist/nw-cli_darwin_amd64    6.3MB
  dist/nw-cli_darwin_arm64    6.0MB
  dist/nw-cli_windows_amd64   6.2MB
```
