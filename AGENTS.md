# AGENTS.md

Agent instructions for the naverworks project.

## Local Skill Invocation Rules

When a trigger condition below is matched, the corresponding skill **must** be invoked via the Skill tool. Do not run the commands directly.

| Skill | Triggers | Description |
|-------|----------|-------------|
| `test` | "테스트", "test", `/test` | Run go test, go vet and report results |
| `build` | "빌드", "build", `/build` | Build local or cross-platform binaries |
| `version` | "버전", "version", `/version`, "bump" | Check, bump, or manage version tags |
| `deploy` | "배포", "릴리스", "deploy", "release", `/deploy` | Cross-build → GitHub Release → npm publish |

## Skill Invocation Order

- Tests run automatically as part of the deploy skill — no need to invoke separately before deploy.
- After a version bump, suggest deploy if appropriate.
- build, test, and version can be invoked independently.

## Commit Rules

- Use the `commit-work` skill for all commits.
- Never run `git commit` directly.
- Commit messages follow Conventional Commits format in Korean.
