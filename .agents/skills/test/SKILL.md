---
name: test
description: Use when running tests for nw-cli — runs go test, go vet, and reports results. Triggers on "테스트", "test", "/test".
---

# nw-cli 테스트 실행

이 스킬은 AI 에이전트가 직접 실행한다. 모든 명령을 순서대로 실행하고 결과를 보고한다.

## 실행 규칙

1. 모든 명령을 직접 Bash로 실행한다.
2. 각 단계의 결과를 ✓/✗ 형식으로 보고한다.
3. 실패한 테스트가 있으면 실패 내용을 상세히 보고한다.
4. 사용자 입력 없이 자동 실행한다.

## 절차

### Phase 1: 모듈 정합성 확인

```bash
go mod tidy
```
```bash
git diff --exit-code go.mod go.sum
```
→ 변경이 있으면 ✗ 보고하고 계속 진행 (중단하지 않음)

### Phase 2: 정적 분석

```bash
go vet ./...
```

### Phase 3: 단위 테스트

```bash
go test ./... -v -count=1
```

결과에서 PASS/FAIL 패키지 수를 집계한다.

### Phase 4: 빌드 확인

```bash
go build -o /tmp/nw-cli-test .
```
```bash
/tmp/nw-cli-test version
```
```bash
rm -f /tmp/nw-cli-test
```

### 보고

```
테스트 결과:

✓ go mod tidy — 변경 없음
✓ go vet — 이상 없음
✓ go test — N 패키지 PASS (N.NNs)
✓ go build — 빌드 성공
```

실패 시:
```
✗ go test — 1/4 패키지 FAIL
  FAIL github.com/physics91/naverworks-cli/internal/api
    TestClient_401_Retry: expected 200, got 401

다음 단계: 실패한 테스트를 수정하세요.
```
