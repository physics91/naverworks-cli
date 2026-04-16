# Journey Testing Stability Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a reusable CLI journey-testing layer that catches real
user-flow regressions across config, auth, API, output, and side effects.

**Architecture:** Extend the existing `cmd/smoke_test.go` patterns into a
shared harness under `internal/testkit/cli` for env setup, stdout/stderr
capture, and scripted HTTP recording. Keep actual `rootCmd` execution inside
`cmd` test helpers and inject it into the harness via callbacks so the test
support code does not create import cycles. Add a command-meta test layer in
`cmd`, onboard high-value user journeys in priority order, and wire
`fast/full/canary` entry points into GitHub Actions so the CI strategy matches
the design.

**Tech Stack:** Go 1.25, `testing`, `net/http/httptest`, Cobra, repository
fixtures under `testdata/`

---

### Task 1: Baseline the current regression surface

**Files:**
- Read: `cmd/smoke_test.go`
- Read: `cmd/e2e_security_test.go`
- Read: `internal/api/service_test.go`
- Create: `docs/2026-04-16-journey-testing-baseline-notes.md`

**Step 1: Capture the current test-layer roles**

Write baseline notes that list what each existing test file protects today and
what user-flow gaps remain.

**Step 2: Run the current focused test packages**

Run: `go test ./cmd ./internal/api ./internal/auth -count=1`
Expected: PASS with the current suite behavior recorded for reference

**Step 3: Save baseline notes**

Document the current gaps in
`docs/2026-04-16-journey-testing-baseline-notes.md`.

**Step 4: Commit**

```bash
git add docs/2026-04-16-journey-testing-baseline-notes.md
git commit -m "docs(plan): 여정 테스트 베이스라인 정리"
```

### Task 2: Introduce the shared journey harness skeleton

**Files:**
- Create: `internal/testkit/cli/harness.go`
- Create: `internal/testkit/cli/harness_test.go`
- Read: `cmd/smoke_test.go`
- Read: `cmd/e2e_security_test.go`

**Step 1: Write the failing harness test**

Create `internal/testkit/cli/harness_test.go` with a focused test that expects
the harness to:

- create an isolated fake home
- set baseline NAVER WORKS env vars
- capture stdout and stderr independently

**Step 2: Run the new harness test**

Run: `go test ./internal/testkit/cli -run TestHarnessCreatesIsolatedEnv -v`
Expected: FAIL because the package and harness do not exist yet

**Step 3: Write the minimal harness implementation**

Create `internal/testkit/cli/harness.go` with:

- a harness struct
- environment setup helpers
- stdout/stderr capture helpers
- cleanup hooks using `t.Cleanup`

**Step 4: Re-run the harness test**

Run: `go test ./internal/testkit/cli -run TestHarnessCreatesIsolatedEnv -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/testkit/cli/harness.go internal/testkit/cli/harness_test.go
git commit -m "test(cli): 여정 테스트 하네스 뼈대 추가"
```

### Task 3: Add request recording and API script support

**Files:**
- Modify: `internal/testkit/cli/harness.go`
- Modify: `internal/testkit/cli/harness_test.go`

**Step 1: Write the failing request-recorder test**

Add a test that starts a scripted `httptest.Server`, executes a sample request,
and expects:

- request method/path/query capture
- optional body capture
- deterministic scripted response delivery

**Step 2: Run the focused recorder test**

Run: `go test ./internal/testkit/cli -run TestHarnessRecordsRequests -v`
Expected: FAIL because request recording is not implemented yet

**Step 3: Implement request recording**

Extend the harness with:

- a request-log struct
- response-script registration
- helpers to expose captured requests for assertions

**Step 4: Re-run the recorder test**

Run: `go test ./internal/testkit/cli -run TestHarnessRecordsRequests -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/testkit/cli/harness.go internal/testkit/cli/harness_test.go
git commit -m "test(cli): 요청 기록과 API 스크립트 추가"
```

### Task 4: Add command-meta contract coverage

**Files:**
- Create: `cmd/meta_contract_test.go`
- Read: `cmd/root.go`
- Read: `cmd/helpers.go`
- Read: representative domain commands under `cmd/`

**Step 1: Write the failing meta-contract test**

Create a test that walks the Cobra tree and fails when representative command
policy is violated, including:

- missing `Use` or `Short`
- missing pagination flag policy on list-style commands
- inconsistent list-style command naming or parent registration

**Step 2: Run the new meta-contract test**

Run: `go test ./cmd -run TestCommandMetaContracts -v`
Expected: FAIL on at least one unchecked policy gap or missing scaffolding

**Step 3: Implement the meta-contract walker**

Add reusable helpers in `cmd/meta_contract_test.go` to traverse commands and
check policy rules with actionable failures.

**Step 4: Re-run the meta-contract test**

Run: `go test ./cmd -run TestCommandMetaContracts -v`
Expected: PASS

**Step 5: Commit**

```bash
git add cmd/meta_contract_test.go
git commit -m "test(cmd): 커맨드 메타 계약 검사 추가"
```

### Task 5: Move CLI execution into the shared harness

**Files:**
- Modify: `internal/testkit/cli/harness.go`
- Create: `cmd/test_runner_test.go`
- Modify: `cmd/smoke_test.go`

**Step 1: Write the failing integration test for shared execution**

Add a new test in `cmd/test_runner_test.go` that passes a `rootCmd` runner
callback into the shared harness, runs a harmless command such as `version`,
and expects valid stdout capture without touching the existing smoke tests yet.

**Step 2: Run the targeted command test**

Run: `go test ./cmd -run TestSharedCLIRunnerVersion -v`
Expected: FAIL because the harness does not yet accept a runner callback

**Step 3: Refactor smoke execution onto the harness**

Implement the callback-based runner wiring in `internal/testkit/cli`, then
replace duplicated environment and output-capture logic in `cmd/smoke_test.go`
with calls into the shared harness through `cmd/test_runner_test.go`.

**Step 4: Re-run smoke tests**

Run: `go test ./cmd -run 'TestSmoke_(Version|Help|ConfigGetInvalidKey)' -v`
Expected: PASS with behavior unchanged

**Step 5: Commit**

```bash
git add internal/testkit/cli/harness.go cmd/test_runner_test.go cmd/smoke_test.go
git commit -m "refactor(test): 공용 CLI 실행 하네스 연결"
```

### Task 6: Add the first journey suite for auth and config

**Files:**
- Create: `cmd/journey_auth_config_test.go`
- Create: `testdata/journey/auth/status/`
- Create: `testdata/journey/config/set-get-list/`
- Read: `cmd/auth.go`
- Read: `cmd/config_cmd.go`

**Step 1: Write the failing auth/status journey test**

Add a scenario that sets up fake profile and token state, runs
`naverworks auth status`, and asserts:

- expected stdout or stderr contract
- no unexpected API requests
- correct fixture-driven state interpretation

**Step 2: Run the auth/status journey test**

Run: `go test ./cmd -run TestJourneyAuthStatus -v`
Expected: FAIL until the harness fixture API is complete

**Step 3: Implement the auth/status scenario**

Use the shared harness and `testdata/journey/auth/status/` fixtures to make the
scenario pass.

**Step 4: Add the config set/get/list journey**

Write a second scenario covering config persistence and visible output across:

- `config set`
- `config get`
- `config list`

**Step 5: Run the focused auth/config journey suite**

Run: `go test ./cmd -run 'TestJourney(AuthStatus|ConfigLifecycle)' -v`
Expected: PASS

**Step 6: Commit**

```bash
git add cmd/journey_auth_config_test.go testdata/journey/auth/status testdata/journey/config/set-get-list
git commit -m "test(journey): 인증과 설정 흐름 추가"
```

### Task 7: Add the first API-backed journey suite

**Files:**
- Create: `cmd/journey_directory_bot_test.go`
- Create: `testdata/journey/directory/list-users/`
- Create: `testdata/journey/bot/send-text/`
- Read: `cmd/directory.go`
- Read: `cmd/bot.go`

**Step 1: Write the failing directory journey**

Add a scenario for `directory list-users --count 20` that asserts:

- request method/path/query
- pagination output behavior
- stable stdout contract

**Step 2: Run the directory journey test**

Run: `go test ./cmd -run TestJourneyDirectoryListUsers -v`
Expected: FAIL until scripted API assertions are wired fully

**Step 3: Implement the directory journey**

Back the scenario with scripted API responses and request-log assertions.

**Step 4: Add the bot send journey**

Add a scenario for `bot send --to USER_ID --text hello` that asserts:

- request payload shape
- expected success output
- no stray side-effect files

**Step 5: Run the focused API-backed journey suite**

Run: `go test ./cmd -run 'TestJourney(DirectoryListUsers|BotSendText)' -v`
Expected: PASS

**Step 6: Commit**

```bash
git add cmd/journey_directory_bot_test.go testdata/journey/directory/list-users testdata/journey/bot/send-text
git commit -m "test(journey): 디렉토리와 봇 대표 흐름 추가"
```

### Task 8: Add failure-category aware assertions

**Files:**
- Modify: `internal/testkit/cli/harness.go`
- Modify: `cmd/journey_auth_config_test.go`
- Modify: `cmd/journey_directory_bot_test.go`

**Step 1: Write the failing category assertion test**

Add a harness test that expects assertion helpers to emit one of the standard
failure categories:

- `SetupFailure`
- `RequestShapeFailure`
- `ResponseHandlingFailure`
- `SideEffectFailure`
- `UXContractFailure`

**Step 2: Run the focused category test**

Run: `go test ./internal/testkit/cli -run TestHarnessFailureCategories -v`
Expected: FAIL because categorized assertion helpers do not exist yet

**Step 3: Implement categorized failure helpers**

Extend the harness with helper functions that prepend failure-category labels to
assertion errors and journey failures.

**Step 4: Update journey tests to use categorized assertions**

Replace plain `t.Fatalf` mismatch reports in the journey suites with the shared
category-aware helpers.

**Step 5: Re-run focused harness and journey tests**

Run: `go test ./internal/testkit/cli ./cmd -run 'TestHarnessFailureCategories|TestJourney' -v`
Expected: PASS

**Step 6: Commit**

```bash
git add internal/testkit/cli/harness.go cmd/journey_auth_config_test.go cmd/journey_directory_bot_test.go
git commit -m "test(journey): 실패 분류 헬퍼 추가"
```

### Task 9: Add fast/full test entry points

**Files:**
- Modify: `Makefile`
- Modify: `.github/workflows/ci.yml`
- Modify: `README.md`
- Modify: `docs/wiki/Troubleshooting.md`

**Step 1: Write the failing workflow expectation**

Document and codify the expected command entry points:

- fast suite
- full repository suite with all journey coverage
- existing release-time canary hook

**Step 2: Add Make targets**

Introduce explicit targets such as:

- `make test-fast`
- `make test-full`

while keeping existing commands intact.

**Step 3: Update GitHub Actions to use the split**

Modify `.github/workflows/ci.yml` so:

- pull requests run `make test-fast`
- pushes to `main` or `master` run `make test-full` plus existing build/vet

**Step 4: Run the new fast target**

Run: `make test-fast`
Expected: PASS with unit/contract, meta, and the first journey subset

**Step 5: Run the new full target**

Run: `make test-full`
Expected: PASS with the full repository test suite, including all journey tests

**Step 6: Document the workflow**

Update user-facing or contributor-facing docs with the new verification split.

**Step 7: Commit**

```bash
git add Makefile .github/workflows/ci.yml README.md docs/wiki/Troubleshooting.md
git commit -m "docs(test): 빠른 검증과 여정 검증 진입점 추가"
```

### Task 10: Add binary canary scaffolding

**Files:**
- Create: `cmd/canary_binary_test.go`
- Modify: `Makefile`
- Modify: `.github/workflows/release.yml`
- Read: build entry points in the repository

**Step 1: Write the failing canary test**

Create a subprocess-driven test that builds a temporary `naverworks` binary and
executes at least one harmless command such as `version`, asserting:

- subprocess stdout contract
- zero exit status
- no stderr noise

**Step 2: Run the canary test**

Run: `go test ./cmd -run TestBinaryCanaryVersion -v`
Expected: FAIL until build-and-run scaffolding exists

**Step 3: Implement the binary canary helper**

Add helper logic to build the binary once per test run and execute subprocess
commands with controlled env and working directory.

**Step 4: Wire the canary into release verification**

Modify `.github/workflows/release.yml` to run the binary canary test before the
goreleaser step so release tags fail fast on process-level regressions.

**Step 5: Re-run the canary test**

Run: `go test ./cmd -run TestBinaryCanaryVersion -v`
Expected: PASS

**Step 6: Commit**

```bash
git add cmd/canary_binary_test.go Makefile .github/workflows/release.yml
git commit -m "test(canary): 바이너리 검증 뼈대 추가"
```
