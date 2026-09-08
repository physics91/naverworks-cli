---
name: test
description: >
  Use when verifying the naverworks codebase with full or focused Go tests,
  static analysis, and build smoke checks. Triggers on "테스트", "test", "/test".
  For local binaries use build; for publishing a release use deploy.
---

# naverworks 테스트

실행 명령의 기준은 `Makefile`과 `.github/workflows/ci.yml`이다.
테스트 범위를 명시하고, 부분 실행 결과를 전체 통과로 보고하지 않는다.

## 범위

| scope | 실행 | 의미 |
|---|---|---|
| `all` (기본) | `make verify-maintenance` | 모듈·포맷·vet·fast·full·canary·빌드·취약점 검사 |
| `fast` | `make test-fast` | 공통 CLI harness, API/auth, 주요 CLI 계약·여정 |
| `unit` | 아래 단위 범위 | 유틸리티·설정 중심 부분 검사 |
| `integration` | 아래 통합 범위 | API/auth 및 CLI 여정 중심 부분 검사 |
| `e2e` | 아래 CLI 범위 | 명령·여정·메타데이터 및 실제 바이너리 canary |
| `coverage` | 아래 커버리지 절차 | 전체 패키지 테스트와 현재 커버리지 측정 |

## 실행 전

- 저장소 루트와 현재 작업 상태를 확인한다. 소스 수정이 필요하면 상위
  `AGENTS.md`의 작업 트리 규칙을 먼저 적용한다.
- `all`은 `make build`로 루트 `naverworks`를 생성하거나 덮어쓴다.
  실행 전에 `git check-ignore -q -- naverworks`로 무시 경로임을 확인한다.
  기존 바이너리가 있으면 저장소 밖 임시 경로에 보존하고 종료 시 복원한다.
- `go mod tidy`로 파일을 자동 수정하지 않는다. 정합성 검사는
  `go mod tidy -diff`이며, 차이가 있으면 실패와 수정 필요성을 보고한다.
- 테스트는 모의 HTTP 서버를 사용한다. 실제 NAVER WORKS 호출이나 인증 정보는
  필요하지 않다. 도구·모듈 설치 및 취약점 DB 조회에는 네트워크가 필요할 수 있다.

## 기본 전체 검증

각 명령의 종료 코드를 확인한다.

```bash
scripts/check-reuse-guardrails.sh .
```
```bash
make verify-maintenance
```

두 명령이 Linux CI 기준이다. `verify-maintenance`는 최초 실패에서 멈추므로
이후 단계를 통과로 표시하지 않는다. 원인 진단에 필요한 독립 검사만 추가 실행한다.
`make test-full`은 필터 없이 `go test ./... -v -count=1`을 실행하여 새 패키지와
`TestJourney*`도 포함한다. 고정 정규식 목록으로 전체 실행을 대체하지 않는다.

## 부분 검증

`fast`, `unit`, `integration`, `e2e`, `coverage`는 전체 유지보수 검증이 아니다.
부분 범위에도 `go mod tidy -diff`와 `go vet ./...`를 실행하고 별도로 보고한다.

단위 범위 (패키지에 파일 I/O 검증도 포함):

```bash
go test ./internal/output/... ./internal/config/... ./internal/fileutil/... ./internal/httputil/... ./internal/authdoctor/... -v -count=1
```

통합 범위 (각각 실행):

```bash
go test ./internal/api/... ./internal/auth/... ./internal/testkit/cli/... -v -count=1
```
```bash
go test ./cmd/... -run '^TestJourney' -v -count=1
```

CLI 범위 (각각 실행):

```bash
go test ./cmd/... -v -count=1
```
```bash
make test-canary
```

부분 범위의 빌드 확인에는 `make test-canary`를 사용한다. 이미 실행했으면
반복하지 않는다. 이 테스트는 임시 디렉터리에 실제 바이너리를 빌드해 실행한다.

커버리지 파일은 저장소 밖 임시 디렉터리에 생성한다. 아래는 한 Bash 세션에서
실행하며 실패해도 자신이 만든 임시 디렉터리만 정리한다.

```bash
(
  set -e
  coverage_dir=$(mktemp -d)
  trap 'rm -rf -- "$coverage_dir"' EXIT
  go test ./... -count=1 -cover -coverprofile="$coverage_dir/coverage.out"
  go tool cover -func="$coverage_dir/coverage.out"
)
```

## 보고와 한계

- 요청 범위, 실행 명령, PASS/FAIL/미실행, 최초 실패 원인과 다음 조치를 보고한다.
- 커버리지와 실행 시간은 이번 측정값만 보고한다. 과거 수치를 현재값으로 쓰지 않는다.
- `git status --short`로 추적 파일 변경 여부를 확인하고 기존 변경과 구분한다.
- Linux 결과로 Windows ACL 검증을 통과 처리하지 않는다. 자격 증명 파일 보안
  변경은 Windows CI 결과가 필요하며, 확인하지 못했으면 명시한다.
- 모의 API 성공을 실제 테넌트 검증 성공으로 표현하지 않는다.
