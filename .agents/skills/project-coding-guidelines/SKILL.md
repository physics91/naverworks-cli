---
name: project-coding-guidelines
description: "Project-local coding guidance generated from repository evidence. Use when writing, changing, or reviewing code in this repository to preserve module boundaries, reuse existing design, run project-native verification, and gate performance complexity on measurements."
---

<!-- clean-code:managed:start -->
<!-- clean-code:sync format=3 evidence=9b2e74272496719ec647bbe41c41216c41fac064ba242ad163932972f5c53ede manual=a0f35af64c54d745fc4c0fdff113b85192b13832c851371ccbb849fb1bb0fe4a ledger=695be87be387a3e93efbc3f42867c1f7a526122b1c60268f163f80619c4b4a04 managed=83395285160e8e3536958a6cc3554e958045ad7237ee52ff1e92c5dce4f7d512 -->
# Project Coding Guidelines

Use these rules with the controlling repository instructions. Active instructions and current repository evidence win every conflict.

## Applicability Map

| Path | Stack | Evidence |
|---|---|---|
| `.` | Go, Shell | `@source-stack:Go`, `@source-stack:Shell`, `go.mod` |
| `.agents/skills/reuse-governor-local/scripts` | Shell | `@source-stack:Shell` |
| `cmd` | Go | `@source-stack:Go` |
| `internal` | Go | `@source-stack:Go` |
| `npm` | Shell | `@source-stack:Shell` |
| `npm/cli` | Node.js | `@source-stack:Node.js`, `npm/cli/package.json` |
| `scripts` | Go, Shell | `@source-stack:Go`, `@source-stack:Shell` |

## Architecture and Maintainability

- [EVIDENCE-001] Verify a project-specific claim against the linked evidence ledger before treating it as mandatory [generic baseline: base clean-code workflow]
- [ARCH-001] Keep responsibilities inside the applicable module boundary; preserve public contracts, dependency direction, validation, error, concurrency, and persistence behavior [generic baseline: base clean-code workflow; project applicability only from the evidence map]
- [MAINT-001] Split code when mixed responsibilities, repeated coupled changes, obscured control flow, or duplication with real change cost makes maintenance harder; treat numeric size limits as sourced soft signals only [generic baseline: base clean-code workflow]
- [REUSE-001] Search existing modules and abstractions before adding a helper; extract only for demonstrated reuse, domain meaning, or duplicated change cost [generic baseline: base clean-code workflow]

## Stack-Aware Rules

- [STACK-GO] For `.`, `cmd`, `internal`, `scripts`, Keep packages cohesive and define interfaces at consumer boundaries; benchmark allocations and concurrency before pooling or goroutine tuning [stack evidence: `@source-stack:Go`, `go.mod`]
- [STACK-NODE-JS] For `npm/cli`, Keep synchronous CPU or filesystem work out of latency-sensitive request paths; measure event-loop or CPU pressure before adding workers, queues, or caches [stack evidence: `@source-stack:Node.js`, `npm/cli/package.json`]
- [STACK-SHELL] For `.`, `.agents/skills/reuse-governor-local/scripts`, `npm`, `scripts`, Keep orchestration steps small, quote data-bearing expansions, and isolate filesystem or process side effects; measure startup and subprocess cost before adding parallelism or caching [stack evidence: `@source-stack:Shell`]

## Project-Native Verification

- [VERIFY-001] Run `go test ./...` when its module and change type apply [evidence: `go.mod`]
- [VERIFY-002] Run `npm --prefix npm/cli run postinstall` when its module and change type apply [evidence: `npm/cli/package.json`]

## Performance

- [PERF-001] Use the repository measurement sources before optimizing, require measured benefit, and run regression verification [evidence: `cmd/config_profile_test.go`]

## Evidence and Precedence

Read the [evidence ledger](references/evidence.md) before applying a project-specific rule
<!-- clean-code:managed:end -->

<!-- clean-code:manual:start -->
## NAVER WORKS CLI Rules

These repository-specific rules refine the managed baseline. Where a rule below
explicitly marks a managed rule as inapplicable, follow the repository evidence
as required by `EVIDENCE-001`.

### Command and API Contracts

- [NW-ARCH-001] Pair each NAVER WORKS domain command in `cmd/<domain>.go` with its REST service implementation in `internal/api/<domain>.go`. Keep Cobra flags, argument validation, orchestration, and presentation in `cmd`; keep endpoint paths, request bodies, and service calls in `internal/api` (`AGENTS.md`, `cmd/bot.go`, `internal/api/bot.go`).
- [NW-ARCH-002] Route HTTP execution, token refresh, retry behavior, response-size limits, and API error decoding through `internal/api/client.go` and `internal/api/errors.go`; do not reimplement transport behavior in domain services (`AGENTS.md`, `internal/api/client.go`, `internal/api/errors.go`).
- [NW-API-001] Escape every user-controlled URL path segment with `url.PathEscape`. Build query strings with `BuildPaginationQuery`, `BuildListQuery`, or `url.Values`; do not concatenate raw identifiers or query values. Use `runListCmd` for the standard `cursor`/`count`/`all` flow and cover method and path changes in `internal/api/service_test.go` or a focused API path test (`internal/api/bot.go`, `internal/api/pagination.go`, `internal/api/api_path_test.go`, `internal/api/service_test.go`).
- [NW-PREVIEW-001] Preserve `--dry-run`, `--plan-out`, and `--generate-input` for every API-backed command. Preview execution must not send network requests, must use the same final method/path/body construction as live execution, and must write plan files through `fileutil.WriteSecureJSON`. Add a no-network preview test when a command introduces custom request or upload orchestration (`cmd/root.go`, `cmd/helpers.go`, `internal/api/client.go`, `cmd/journey_*_test.go`).
- [NW-AUTH-001] Preserve the dependency boundary between `internal/auth` and `internal/config`: neither package may import the other. Put shared secure filesystem behavior in `internal/fileutil` and shared transport hardening in `internal/httputil` (`AGENTS.md`, current package imports).
- [NW-CLI-001] Write user-facing command text and errors in Korean, keep Go identifiers in English, and preserve the stderr JSON error envelope `{"error":{"code":"...","description":"..."}}` for non-interactive failures. Keep `Use` and `Short` non-empty for visible commands, describe list-style commands with `목록`, declare `count` whenever `cursor` or `all` is present, and avoid duplicate sibling command names (`AGENTS.md`, `main.go`, `cmd/meta_contract_test.go`).
- [NW-OUTPUT-001] Send API response bodies through the existing `internal/output` formatter and sanitization boundary. Do not add a direct output path that can bypass response masking (`cmd/helpers.go`, `internal/output`).

### Reuse, Testing, and Security

- [NW-REUSE-001] Check `cmd/helpers.go` and `docs/reuse/catalog.yaml` before adding command helpers. Use cataloged helpers only when their documented contract matches; keep domain-specific wrapper families local when promotion would create a generic one-use abstraction (`docs/reuse/catalog.yaml`, `scripts/check-reuse-guardrails.sh`).
- [NW-TEST-001] Match tests to the changed contract: update `cmd/smoke_test.go` for command registration, command metadata tests for flag/help structure, API service/path tests for method or endpoint changes, and journey tests for multi-step CLI, output, or preview behavior. For Go changes, require `gofmt`, `go vet ./...`, and `go test ./... -v`; use `make verify-maintenance` for the full local CI/release baseline (`AGENTS.md`, `Makefile`, `cmd/smoke_test.go`, `cmd/meta_contract_test.go`, `internal/api/service_test.go`).
- [NW-TEST-002] Run `scripts/check-reuse-guardrails.sh .` separately when changing `cmd` pagination/helper use or `docs/reuse` lifecycle data; `make verify-maintenance` does not include this check. Treat the reuse guardrail plus `make verify-maintenance` as the Linux CI baseline (`scripts/check-reuse-guardrails.sh`, `.github/workflows/ci.yml`, `Makefile`).
- [NW-SEC-001] Treat authentication, credential persistence, uploads, redirects, and output masking as security-sensitive boundaries. Preserve secure file permissions and atomic JSON writes, explicit upload-host validation, redirect controls, and targeted security tests when changing those paths (`internal/fileutil`, `internal/api/client.go`, `cmd/*security_test.go`, `internal/*/*security_test.go`).
- [NW-WINDOWS-001] Treat Windows ACL behavior as a separate platform contract. Changes under `internal/fileutil`, or auth/config changes that persist credentials, require the Windows CI path or an explicit statement that ACL behavior remains unverified; Linux-only results do not validate `icacls` handling (`internal/fileutil/*windows*.go`, `.github/workflows/ci.yml`).

### Packaging, Automation, and Performance

- [NW-NPM-001] Treat `npm/cli` as an install-time distribution wrapper, not a latency-sensitive request service. The managed `STACK-NODE-JS` latency guidance and `VERIFY-002` postinstall command do not apply here. Do not run `npm --prefix npm/cli run postinstall` as validation because it copies and chmods a sidecar binary and depends on installed optional packages; validate wrapper behavior with `go test . -run '^TestNpmWrapper'` or the broader Go test suite (`npm/cli/install.js`, `npm/cli/platform.js`, `npm_wrapper_test.go`).
- [NW-NPM-002] Keep the CLI package version and every platform optional dependency version identical, preserve the platform/package/binary mapping, and verify archive names against `.goreleaser.yml` when changing release packaging (`npm/cli/package.json`, `npm/cli/platform.js`, `npm/build-npm.sh`, `.goreleaser.yml`).
- [NW-CI-001] Preserve full commit-SHA pins for GitHub Actions together with readable version comments, keep workflow permissions minimal, and retain OIDC trusted publishing for npm releases. Treat changes to workflow permissions, action execution, or release publishing as security-sensitive (`.github/workflows/*.yml`).
- [NW-PERF-001] No repository benchmark or production profile is currently recorded. `cmd/config_profile_test.go` is a correctness test, not a performance measurement source, so the managed `PERF-001` citation is not sufficient by itself. Add or identify a targeted benchmark/profile and record before/after evidence before accepting complexity-increasing optimization (`cmd/config_profile_test.go`, absence of `Benchmark*` functions in current tests).
<!-- clean-code:manual:end -->
