# AGENTS.md

naverworks-cli — A Go CLI for NAVER WORKS REST API v1.0.

## Project Context

- Go 1.25, Cobra CLI framework, minimal external dependencies
- Profile-based multi-auth (OAuth + JWT) with automatic token refresh
- `cmd/` for Cobra commands, `internal/` for business logic separation
- `npm/` wrapper for cross-platform distribution (goreleaser + GitHub Actions)

## Coding Standards

- Domain file pairs: `cmd/<domain>.go` ↔ `internal/api/<domain>.go`
- Korean user-facing messages, English code/variable names
- Conventional Commits format, Korean commit messages
- `go vet ./...` must pass, `go test ./... -v` must pass
- When adding new commands, add registration smoke tests in `cmd/smoke_test.go`

## Build & Test

```bash
make build          # Build ./naverworks binary
make test           # go test ./... -v
go vet ./...        # Static analysis
```

## Architecture Rules

- `internal/api/client.go`: HTTP requests, token refresh, error parsing
- `internal/auth/`: Token issuance (OAuth/JWT), storage, refresh — no dependency on config
- `internal/config/`: Profile loading, env var overrides — no dependency on auth
- `cmd/helpers.go`: Common flag handling, API client creation, pagination utilities
- Errors are written to stderr as JSON: `{"error":{"code":"...","description":"..."}}`

## Local Skill Invocation Rules

In Claude Code, when a trigger condition below is matched, the corresponding skill **must** be invoked via the Skill tool. Do not run the commands directly.

| Skill | Triggers | Description |
|-------|----------|-------------|
| `test` | "테스트", "test", `/test` | Run go test, go vet, and local build smoke check |
| `build` | "빌드", "build", `/build` | Build local or cross-platform binaries with ldflags version metadata |
| `version` | "버전", "version", `/version`, "bump" | Inspect version state or create/push release tags |
| `deploy` | "배포", "릴리스", "deploy", "release", `/deploy` | Preflight checks → release tag → GitHub Actions verification |
| `release` | "릴리스 관리", "release manage", "release edit", "release delete", "release rollback" | View, edit, delete, rollback GitHub Releases |
| `naverworks-profile` | "프로필", "인증", "NW_PROFILE", "auth setup" | Multi-profile setup, auth, and troubleshooting |
| `commit-work` | Any commit request | Create commits (never run `git commit` directly) |

## Skill Invocation Order

- `deploy` includes tests — no need to invoke `test` separately before deploy
- After a version bump, suggest deploy if appropriate
- `build`, `test`, and `version` can be invoked independently
- "릴리스 관리", "release manage", "release edit", "release delete", "release rollback" 요청은 `release`를 `deploy`보다 우선한다
