---
name: build
description: Use when building naverworks binaries — single-platform or cross-platform artifacts with ldflags version metadata. Triggers on "빌드", "build", "/build". If you need test results only, use the test skill. If you need to publish a release tag, use deploy or version.
---

# naverworks 빌드

이 스킬은 AI 에이전트가 직접 실행한다. 로컬 또는 크로스 플랫폼 바이너리를 빌드한다.

## 입출력 계약

### 입력

| 입력 | 필수 여부 | 형식 | 설명 |
|------|----------|------|------|
| 인자 | 선택 | enum: `local`(기본), `all` | `local`은 현재 플랫폼, `all`은 5개 플랫폼 크로스 빌드 |

### 출력

| 필드 | 설명 |
|------|------|
| `mode` | `local` 또는 `all` |
| `artifacts` | 생성된 바이너리/아카이브 경로 목록 |
| `version_metadata` | `./naverworks version` 결과 (JSON) |

### 성공 기준

- `go vet` 및 `go test`가 통과한다.
- `local`: `./naverworks` 실행 파일이 생성되고 `naverworks version`이 동작한다.
- `all`: 5개 플랫폼 아카이브(4 tar.gz + 1 zip)가 `dist/`에 생성된다.

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
ls -lh naverworks
```
```bash
./naverworks version
```

### Phase 2b: 크로스 빌드 (all)

버전 추출 — `v` prefix를 제거하여 goreleaser/deploy와 일치시킨다:

```bash
git describe --tags --abbrev=0 2>/dev/null
```
→ 태그가 있으면 `v` prefix 제거하여 VERSION에 저장 (예: `v0.1.0` → `0.1.0`)
→ 태그가 없으면 `dev` 사용

goreleaser 확인:
```bash
command -v goreleaser
```

**goreleaser가 있으면** (권장):
```bash
goreleaser release --clean --skip=publish --snapshot
```

**goreleaser가 없으면** 각 플랫폼을 수동으로 빌드한다. LDFLAGS 구성, 플랫폼별 `GOOS/GOARCH` 명령, 아카이브 포맷은 [`references/cross-build-commands.md`](references/cross-build-commands.md)를 참조한다.

### 보고

로컬 빌드:
```
빌드 완료:
  naverworks (6.1MB)
  버전: {"version":"dev","commit":"abc1234","build_date":"..."}
```

크로스 빌드:
```
빌드 완료: 5개 플랫폼

  dist/naverworks_0.1.0_linux_amd64.tar.gz     2.6MB
  dist/naverworks_0.1.0_linux_arm64.tar.gz     2.4MB
  dist/naverworks_0.1.0_darwin_amd64.tar.gz    2.6MB
  dist/naverworks_0.1.0_darwin_arm64.tar.gz    2.5MB
  dist/naverworks_0.1.0_windows_amd64.zip      2.7MB
```
