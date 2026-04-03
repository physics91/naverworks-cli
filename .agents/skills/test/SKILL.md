---
name: test
description: >
  Use when verifying the naverworks codebase — runs unit, integration, and
  e2e test layers with go test, go vet, and build smoke check. Triggers on
  "테스트", "test", "/test". Covers test pyramid (unit/integration/e2e) with
  real functionality verification. For local binaries, use build. For releasing,
  use deploy. For managing existing releases, use release. For version
  inspection or tag creation, use version.
---

# naverworks 테스트

이 스킬은 AI 에이전트가 직접 실행한다. 모든 명령을 순서대로 실행하고 결과를 보고한다.

## 실행 규칙

1. 모든 명령을 직접 Bash로 실행한다.
2. 각 단계의 결과를 ✓/✗ 형식으로 보고한다.
3. 실패한 테스트가 있으면 실패 내용을 상세히 보고한다.
4. 사용자 입력 없이 자동 실행한다.

## Test Pyramid Strategy

### Layer 1: Unit Tests
- Target: 순수 함수, 데이터 모델, 유틸리티
  - internal/output — 포맷팅 로직
  - internal/auth/jwt — JWT assertion 빌드
  - internal/config — 설정 저장/로드/마이그레이션
  - internal/auth/token — 토큰 저장소
  - cmd/task_cmd — body 빌더
  - cmd/smoke_test.go 내 헬퍼/파서 (TestResolveUserID, TestRequireTitleBodyPost 등)
- Isolation: 파일시스템은 t.TempDir(), HTTP 없음
- Speed: <1초, 매 커밋마다 실행
- Commands (각 명령 개별 실행 후 결과를 모두 보고):
  1. `go test ./internal/output/... ./internal/config/... -v -count=1`
  2. `go test ./internal/auth/... -run "Test(BuildJWT|CheckKey|Token|ProfileToken|WriteSecure|SaveSecure)" -v -count=1`
  3. `go test ./cmd/... -run "Test(BuildTask|ResolveUserID|RequireTitleBodyPost|ParseOptionalJSONData|ResolveBotID|ResolveOrCreateProfile)" -v -count=1`

### Layer 2: Integration Tests
- Target: 모듈 경계 — HTTP 클라이언트, OAuth 흐름, API 서비스 엔드포인트
  - internal/api/client — 인증 헤더, 재시도, 에러 파싱, 업로드/다운로드
  - internal/auth/oauth — 토큰 교환/리프레시/폐기
  - internal/api/* — 서비스별 엔드포인트 경로·메서드 검증 (100+ 테이블 기반 케이스)
- Isolation: httptest.NewServer로 로컬 HTTP 서버, 실제 외부 API 호출 없음
- Speed: 1~3초, PR 전/CI 전 실행
- Fixtures: 테이블 기반 테스트 케이스 (100+ 서비스 메서드)
- Commands (각 명령 개별 실행 후 결과를 모두 보고):
  1. `go test ./internal/api/... -v -count=1`
  2. `go test ./internal/auth/... -run "Test(BuildAuthorizationURL|ExchangeCode|RefreshToken|RevokeToken|FindAvailableListener|HasScope|GenerateState|RequestToken)" -v -count=1`

### Layer 3: E2E Tests
- Target: CLI 명령 워크플로우 검증, 보안 검증, 워크플로우 파일 무결성
  - cmd/smoke_test — 46개 CLI 워크플로우 (버전, 도움말, 플래그 검증, JSON 입력)
  - cmd/e2e_security — atomic write, 동시성, 응답 크기 제한, SHA 핀닝
- Isolation: setupTestEnv로 HOME/환경변수 격리, httptest로 API 모킹
- Speed: 수초~수십초, 릴리스 전 또는 보안 변경 후 실행
- Entry Points:
  - naverworks version / help
  - naverworks bot send (--to, --channel, 플래그 충돌)
  - naverworks config get/set
  - naverworks auth status
  - naverworks calendar (create/delete event 인자 검증)
  - .github/workflows/*.yml SHA 핀닝 검증
- Commands:
  1. `go test ./cmd/... -run "Test(Smoke|E2E)" -v -count=1`

## Procedure

> Run by Layer 섹션의 명령을 레이어 실행의 단일 기준(source of truth)으로 사용하며, Layer 설명과 Procedure는 이 명령과 동일해야 한다.

### Run All Tests (피라미드 순서)

1. Phase 1: 모듈 정합성
   ```bash
   go mod tidy
   ```
   ```bash
   git diff --exit-code go.mod go.sum
   ```
   → 변경이 있으면 ✗ 보고하고 이후 단계는 계속 진행 (중단하지 않음)

2. Phase 2: 정적 분석
   ```bash
   go vet ./...
   ```
   → 실패 시 실패 패키지와 대표 오류를 보고하고 이후 단계는 계속 진행

3. Phase 3: Unit
   - Run by Layer의 Unit only 명령을 그대로 실행
   → 실패 시 FAIL 패키지와 테스트 함수명 보고

4. Phase 4: Integration
   - Run by Layer의 Integration only 명령을 그대로 실행
   → 실패 시 FAIL 패키지와 테스트 함수명 보고

5. Phase 5: E2E
   - Run by Layer의 E2E only 명령을 그대로 실행
   → 실패 시 FAIL 패키지와 테스트 함수명 보고

6. Phase 6: 빌드 확인
   ```bash
   go build -o /tmp/naverworks-test .
   ```
   ```bash
   /tmp/naverworks-test version
   ```
   ```bash
   rm -f /tmp/naverworks-test
   ```

### Run by Layer

- **Unit only** (각 명령 개별 실행 후 결과를 모두 보고):
  1. `go test ./internal/output/... ./internal/config/... -v -count=1`
  2. `go test ./internal/auth/... -run "Test(BuildJWT|CheckKey|Token|ProfileToken|WriteSecure|SaveSecure)" -v -count=1`
  3. `go test ./cmd/... -run "Test(BuildTask|ResolveUserID|RequireTitleBodyPost|ParseOptionalJSONData|ResolveBotID|ResolveOrCreateProfile)" -v -count=1`

- **Integration only** (각 명령 개별 실행 후 결과를 모두 보고):
  1. `go test ./internal/api/... -v -count=1`
  2. `go test ./internal/auth/... -run "Test(BuildAuthorizationURL|ExchangeCode|RefreshToken|RevokeToken|FindAvailableListener|HasScope|GenerateState|RequestToken)" -v -count=1`

- **E2E only**:
  1. `go test ./cmd/... -run "Test(Smoke|E2E)" -v -count=1`

### Run with Coverage

```bash
go test ./... -cover -coverprofile=coverage.out
```
```bash
go tool cover -func=coverage.out
```

### 보고

성공 시:
```
테스트 결과:

✓ go mod tidy — 완료
✓ git diff go.mod go.sum — 변경 없음
✓ go vet — 이상 없음
✓ Unit — N 패키지 PASS
✓ Integration — N 패키지 PASS
✓ E2E — N 패키지 PASS
✓ go build — 빌드 성공
```

실패 시:
```
✗ Integration — 1/4 패키지 FAIL
  FAIL github.com/physics91/naverworks-cli/internal/api
    TestClient_401_Retry: expected 200, got 401

다음 단계: 실패한 테스트를 수정하세요.
```

## Test Environment Setup

- 외부 의존성: 없음 (Docker Compose 불필요)
- 환경 격리: t.TempDir() + t.Setenv() (NW_* 변수 클리어)
- 테스트 헬퍼:
  - cmd/smoke_test.go: setupTestEnv(), writeTestConfig(), runCLI(), captureStdout()
  - internal/auth/jwt_test.go: generateTestKey()

## What to Test vs What to Mock

| Layer | Real | Mocked |
|-------|------|--------|
| Unit | 함수 로직, 파일 I/O (TempDir) | HTTP, 외부 서비스, 시간 |
| Integration | HTTP 클라이언트 + httptest 서버 | 실제 NAVER WORKS API |
| E2E | CLI 실행, Cobra 명령 라우팅, 보안 검증 | API 서버 (httptest) |

## Best Practices

- 테스트는 -count=1로 캐시를 끄고 실행한다
- 파일시스템 의존 테스트는 t.TempDir와 t.Setenv로 격리한다
- 네트워크 의존 테스트는 실제 외부 호출 대신 httptest.NewServer를 사용한다
- 테이블 기반 테스트(table-driven tests)로 엔드포인트 검증을 구조화한다

## Caveats

- CI에서 -race 플래그 미사용
- t.Parallel() 미사용 (순차 실행)
- cmd/ 커버리지 12.9%로 낮음
- Docker Compose 없음 (순수 Go 프로젝트라 불필요)
- e2e_security_test.go는 //go:build !windows 태그 — Windows에서 일부 건너뜀
