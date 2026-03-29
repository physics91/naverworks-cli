# 구현 계획: 보안 이슈 수정 (Tier 1)

> **Decision**: `/tmp/decisions/2026-03-29-naverworks-cli-security-remediation.md`
> **Date**: 2026-03-29
> **Scope**: Tier 1 즉시 수정 3건 (코드 5개 파일 + 관련 *_test.go, 총 ~90 LOC)

## 배경

보안 리뷰에서 발견된 12건 중 ROI가 높은 3건을 즉시 수정한다.

## 태스크

### Task 1: 원자적 파일 쓰기 (M1 + L1)

**목적**: 프로세스 크래시 시 토큰/설정 파일 손상 방지

**대상 파일**:
- `internal/auth/token.go` — `writeSecureJSON()` 함수 (line 138)
- `internal/config/config.go` — `saveSecureJSON()` 함수 (line 81)
- 관련 `*_test.go` 파일 — 기존 파일 overwrite/fallback 및 임시 파일 정리 검증

**변경 내용**:

두 함수 모두 동일한 패턴으로 교체:

```go
// Before (비원자적)
os.WriteFile(path, data, 0600)

// After (원자적)
tmp, err := os.CreateTemp(dir, ".naverworks-*.tmp")
// ... write to tmp ...
tmp.Close()
os.Chmod(tmpPath, 0600)  // non-Windows only
os.Rename(tmpPath, path)
defer os.Remove(tmpPath)  // 실패 시 정리
```

**핵심 주의사항**:
- `os.CreateTemp`의 디렉토리는 `filepath.Dir(path)`를 사용하여 cross-device rename 방지
- `defer os.Remove(tmpPath)`로 실패 시 임시 파일 정리
- `tmp.Write` 후 `tmp.Sync()`를 호출한 다음 `tmp.Close()`하여 디스크 내구성 보장
- Windows에서는 대상 파일이 이미 존재할 때 `os.Rename`이 실패할 수 있으므로, Windows 경로는 기존 `os.WriteFile` 동작을 유지한다. non-Windows에서만 temp + sync + rename 경로를 적용하고, Windows 원자적 교체는 후속 이슈로 분리한다.
- `config.go`의 `saveSecureJSON`도 동일 패턴이므로 함께 수정

**테스트**: `go test ./internal/auth/... -v` + `go test ./internal/config/... -v`
- non-Windows: temp 파일 생성 후 `os.Rename`으로 교체되는지 확인
- Windows: 기존 파일이 있는 경로에서 fallback 저장이 성공하는지 확인

---

### Task 2: API 응답 크기 제한 (M4)

**목적**: 비정상 API 응답에 의한 OOM 방지

**대상 파일**: `internal/api/client.go`

**변경 위치** (2곳):
1. `doWithRetry()` 내 `io.ReadAll(resp.Body)` (line 106)
2. `getDownloadURLWithRetry()` 내 `io.ReadAll(resp.Body)` (line 157)

**변경 내용**:

```go
// Before
respBody, err := io.ReadAll(resp.Body)

// After
const maxAPIResponseSize = 10 << 20 // 10MB
respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxAPIResponseSize+1))
if err != nil {
    return nil, fmt.Errorf("응답 읽기 실패: %w", err)
}
if int64(len(respBody)) > maxAPIResponseSize {
    return nil, fmt.Errorf("API 응답 크기 초과: > %d bytes", maxAPIResponseSize)
}
```

**핵심 주의사항**:
- 상수 `maxAPIResponseSize`를 파일 상단에 정의
- `maxSize+1` 패턴으로 초과 시 조용히 잘리지 않고 명시적 오류 반환
- `UploadFile()`은 업로드 전용이므로 변경 불필요
- auth 엔드포인트(`requestToken`, `FetchUserName`)는 신뢰된 서버이므로 이번 수정에서 제외

**테스트**: `go test ./internal/api/... -v`

---

### Task 3: GitHub Actions SHA 고정 (L5)

**목적**: GitHub Actions 공급망 공격 표면 제거

**대상 파일**:
- `.github/workflows/ci.yml`
- `.github/workflows/release.yml`

**변경 내용**:

각 `uses:` 태그를 현재 major 버전의 최신 SHA로 교체:

```yaml
# Before
- uses: actions/checkout@v4
- uses: actions/setup-go@v5

# After (major 태그가 가리키는 commit SHA를 조회해 고정)
- uses: actions/checkout@<commit-sha> # v4 태그가 가리키는 commit
- uses: actions/setup-go@<commit-sha> # v5 태그가 가리키는 commit
```

**대상 uses 항목 (6회, 4개 고유 Action)**:
1. `actions/checkout@v4` → SHA (ci.yml, release.yml)
2. `actions/setup-go@v5` → SHA (ci.yml, release.yml)
3. `goreleaser/goreleaser-action@v6` → SHA (release.yml)
4. `actions/setup-node@v4` → SHA (release.yml)

**SHA 조회 방법** (commit SHA 기준):
```bash
gh api repos/actions/checkout/commits/v4 --jq '.sha'
gh api repos/actions/setup-go/commits/v5 --jq '.sha'
gh api repos/goreleaser/goreleaser-action/commits/v6 --jq '.sha'
gh api repos/actions/setup-node/commits/v4 --jq '.sha'
```

- 조회 결과는 태그 ref object가 아니라 commit SHA여야 하며, YAML에는 반드시 40자리 commit SHA를 사용한다.

**테스트**: CI 워크플로우가 정상 실행되는지 확인 (push 후 Actions 탭 확인)

---

## 리스크 및 대응

- **리스크 1: `gh` CLI 인증 의존** — Task 3은 `gh` CLI 인증과 GitHub 네트워크 접근이 필요하다.
  - 대응: 작업 시작 전에 `gh auth status`로 인증 상태를 확인한다. 인증이 없으면 GitHub 웹 UI에서 각 Action 릴리스 커밋 SHA를 수동 확인한다.

- **리스크 2: Windows 파일 동작 차이** — `os.Chmod`와 `os.Rename`의 동작이 플랫폼별로 다를 수 있다.
  - 대응: `runtime.GOOS == "windows"` 분기를 유지하고, 기존 파일이 이미 존재하는 경로에 대한 저장 테스트를 추가해 overwrite 동작을 확인한다.

- **리스크 3: 응답 크기 초과 시 조용한 잘림** — `io.LimitReader`만 사용하면 초과 응답이 정상 응답처럼 잘려 읽힐 수 있다.
  - 대응: `maxSize+1` 패턴으로 읽고, `len(respBody) > maxSize`이면 명시적 오류를 반환한다.

- **리스크 4: Actions 검증의 외부 의존** — SHA 고정 검증은 push 권한과 GitHub Actions 대기 시간에 의존한다.
  - 대응: 로컬 완료 기준(SHA 치환 + YAML 확인)과 원격 완료 기준(push 후 CI/release 실행 성공)을 분리한다.

## 실행 순서

1. Task 1 (atomic write) — 가장 높은 실질 위험, 핵심 수정
2. Task 2 (LimitReader) — 1줄 변경, 빠른 적용
3. Task 3 (SHA 고정) — SHA 조회 필요, 마지막 수행
4. 전체 검증: `make test && go vet ./... && make build`
5. 커밋

## 검증 체크리스트

- [ ] `go test ./... -v` 전체 통과
- [ ] `go vet ./...` 경고 없음
- [ ] `make build` 성공
- [ ] non-Windows에서 token.go의 writeSecureJSON이 os.CreateTemp + tmp.Sync + os.Rename 경로를 사용한다
- [ ] non-Windows에서 config.go의 saveSecureJSON이 os.CreateTemp + tmp.Sync + os.Rename 경로를 사용한다
- [ ] Windows에서 기존 파일이 있는 경로에 대해 token/config 저장이 fallback(os.WriteFile)으로 성공한다
- [ ] client.go의 io.ReadAll이 io.LimitReader로 래핑되고, 초과 시 명시적 오류를 반환한다
- [ ] ci.yml, release.yml의 각 uses: 값이 40자리 commit SHA이며, 원래 major 버전을 주석으로 보존한다
