# Journey Testing Baseline Notes

## Current Focused Test Status

Command run on 2026-04-16:

```bash
go test ./cmd ./internal/api ./internal/auth -count=1
```

Observed result:

- `ok github.com/physics91/naverworks-cli/cmd 0.764s`
- `ok github.com/physics91/naverworks-cli/internal/api 9.237s`
- `ok github.com/physics91/naverworks-cli/internal/auth 0.401s`

## What Each Existing Layer Protects

### `cmd/smoke_test.go`

Primary role:

- validates command registration and `--help` visibility
- validates representative flag conflicts and required-flag behavior
- exercises a small amount of in-process CLI execution through `rootCmd`

Strengths:

- catches Cobra tree regressions and obvious CLI guardrail mistakes quickly
- already contains reusable pieces for fake home setup, config fixture writing,
  stdout capture, and flag reset

Current gaps:

- stdout is captured, but stderr is not modeled as a first-class assertion path
- mock API interaction is not part of the general smoke layer
- scenarios are mostly single-command checks, not multi-step user journeys
- side effects are only validated in a few targeted cases

### `cmd/e2e_security_test.go`

Primary role:

- validates specific real-behavior scenarios for file persistence,
  oversized-response handling, and security-sensitive edge cases

Strengths:

- proves some meaningful behavior across config save, service layer usage, and
  filesystem side effects
- already demonstrates that `httptest.Server`-backed scenarios fit this repo

Current gaps:

- coverage is narrow and security/file-write oriented
- there is no shared scenario harness for broader CLI workflows
- test names and organization do not yet form a general journey-testing layer

### `internal/api/service_test.go`

Primary role:

- validates request-shape contracts for representative service methods
- validates SCIM token/base URL usage at the client/service boundary

Strengths:

- broad method/path coverage across many domains
- cheap regression protection for endpoint wiring changes

Current gaps:

- proves request construction, not user-facing CLI behavior
- does not validate config/auth/profile/env integration
- does not validate output shaping or filesystem side effects

### `internal/auth/*_test.go`

Primary role:

- validates token issuance, storage, and refresh logic in isolation

Strengths:

- good unit-level protection for auth behavior

Current gaps:

- does not validate auth behavior through CLI entry points and profile setup

## Main User-Flow Gaps

The repository still lacks a reusable layer for representative CLI journeys
that cross all of these boundaries together:

- fake user environment and profile/config/token state
- CLI invocation through the command tree
- API request scripting and request-log assertions
- stdout and stderr contract assertions
- file and state side-effect verification

The highest-value missing journeys remain:

- `auth status`
- `config set -> config get -> config list`
- `directory list-users --count / --all`
- `bot send --to ... --text ...`
- `drive upload --resume`
- `mail send --attachment`
- `scim list-users`

## Immediate Implementation Implication

The next layer should not replace the current tests. It should reuse the
existing `cmd` helpers as the seed for a shared harness, then add scenario
tests that cover real user flows end to end without duplicating the lower-level
request-shape and auth unit coverage that already exists.
