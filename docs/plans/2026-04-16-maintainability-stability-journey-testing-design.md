# Maintainability Stability Journey Testing Design

## Goal

Strengthen maintainability-focused stability in `naverworks-cli` by adding a
test architecture that catches real CLI user-flow regressions before release.

## Current Problem

- Existing `cmd` smoke tests focus mostly on command registration, help output,
  and flag validation.
- Existing `internal/api` tests validate method/path contracts but do not prove
  that user-facing CLI flows still work end to end.
- Existing end-to-end style tests cover a few security and file-write paths, but
  they do not form a reusable layer for representative user journeys.
- As a result, refactors can keep low-level tests green while still breaking
  actual CLI behavior across config, auth, API, output, and file side effects.

## Design Decisions

### 1. Add a dedicated journey-test layer

Keep the current unit and contract tests, but add a new test layer whose target
is the real CLI user flow.

The journey layer validates one user goal per scenario, including:

- fake home and profile/config/token setup
- command execution through the CLI entry path
- mock API interaction
- stdout/stderr behavior
- expected side effects on files or state

This layer is the primary guardrail for refactoring safety.

### 2. Keep a four-layer test architecture

#### Layer 1: Unit/Contract

Retain the existing focused tests in packages such as:

- `internal/api/*_test.go`
- `internal/auth/*_test.go`
- `cmd/helpers*_test.go`

This layer protects low-level logic and request-shape assumptions.

#### Layer 2: Command Meta

Add a structural contract layer for the Cobra tree and common CLI policies.

This layer should verify:

- command tree registration
- required metadata and command conventions
- common flag policy such as pagination behavior
- list-style command naming and registration conventions

Its purpose is to prevent command-structure drift across domains.

#### Layer 3: Journey

Add reusable scenario-driven tests for representative user flows.

This layer becomes the main regression shield for changes that cross multiple
subsystems.

#### Layer 4: Black-box Canary

Add a small set of subprocess-driven checks against the built binary for release
or nightly verification.

This layer exists to catch process-level regressions that in-process Cobra tests
may miss, including argument parsing, env handling, stdio behavior, and exit
semantics.

### 3. Build the journey layer around a reusable harness

Create a shared harness instead of ad-hoc scenario code.

Recommended layout:

- `internal/testkit/cli/` for reusable harness code
- `cmd/journey_*_test.go` for scenario suites
- `cmd/meta_contract_test.go` for structural command checks
- `testdata/journey/...` for stable fixtures

The harness should provide four capabilities:

- environment fixtures
- API stubbing and request recording
- CLI execution helpers
- reusable assertions for output, requests, and side effects

To avoid import cycles, the harness must not import `cmd` directly. The `cmd`
test package should own a thin runner adapter that passes `rootCmd` execution
into the shared harness through function injection.

The harness should extend the current ideas already visible in
`cmd/smoke_test.go` rather than replacing them with a new framework.

### 4. Use strict scenario-selection rules

Journey tests should target high-value flows, not every endpoint.

Priority rules:

- favor commands users run often
- favor flows that cross config/auth/client/api/output boundaries
- favor write paths with larger blast radius
- favor commands with branching behavior such as `--all`, `--resume`,
  attachments, or stdin/json/file modes
- deprioritize simple CRUD operations already covered well by lower layers

### 5. Start with ten representative user journeys

Initial priority list:

1. `auth status`
2. `config set -> config get -> config list`
3. `directory list-users --count / --all`
4. `bot send --to ... --text ...`
5. `drive list --user-id me`
6. `drive upload --resume`
7. `mail send --attachment`
8. `approval list-documents` or `create-user-document`
9. `scim list-users`
10. `version` plus minimal `--help` canary coverage

These flows provide strong regression coverage without turning the test suite
into an unbounded endpoint mirror.

### 6. Standardize failure categories

Journey failures must identify the broken layer, not just report a generic
golden mismatch.

Use these categories:

- `SetupFailure`
- `RequestShapeFailure`
- `ResponseHandlingFailure`
- `SideEffectFailure`
- `UXContractFailure`

Every journey test should report:

- the executed command
- a concise request log summary
- the failure category

This keeps diagnosis fast during refactors.

### 7. Split CI execution by cost and confidence level

Do not run every layer on every change with the same frequency.

Recommended tiers:

- `fast`: unit/contract + command-meta + top journey subset for PR gating
- `full`: full journey suite after merge to `main`
- `canary`: built-binary subprocess scenarios for release or nightly gates

These tiers should be wired into real GitHub Actions workflow changes rather
than remaining Makefile-only conventions.

Supporting policy:

- every new domain gets at least one journey test
- commands with `--all`, `--resume`, attachments, or stdin/json/file branches
  require journey coverage before merge
- any flow that caused a real regression becomes a canary candidate

## Verification Targets

The design is successful when the repository can support the following:

- low-level contract tests remain fast and focused
- command-structure regressions fail in a dedicated meta-test layer
- representative user journeys fail with actionable categories when behavior
  changes unintentionally
- release verification includes a small subprocess-driven binary check set

## Risks And Mitigations

- Risk: journey tests become brittle because they snapshot too much output
  - Mitigation: normalize dynamic values and assert on stable, meaningful fields

- Risk: harness complexity grows faster than scenario coverage
  - Mitigation: keep the harness small and scenario-oriented; prefer extending
    the current `cmd/smoke_test.go` patterns

- Risk: the shared harness introduces `cmd` import cycles
  - Mitigation: keep `rootCmd` execution in `cmd` test adapters and inject the
    runner into `internal/testkit/cli`

- Risk: the suite becomes too slow for normal PR use
  - Mitigation: separate `fast`, `full`, and `canary` execution tiers

- Risk: contributors skip journey coverage for new commands
  - Mitigation: codify minimum journey requirements in CI and review norms
