<!-- plan-workflow: task-detail-v1 -->
# TASK-001 Detail Plan

- Task ID: TASK-001
- Big plan SHA-256: 319235e0e9e6e4b16e57de8b941ee43a35f7e0253cb39d4690a55816a64f30be
- Repository commit: b9af677fdaf721e7fcc7095df958de1bc97c9af1
- Worktree status SHA-256: 0a715d28186a2c28ec78917bc3d196fcb51535dbf031f0d33a6e58493c0430b8
- Requirements: REQ-001, REQ-002, REQ-003, REQ-004, REQ-005, NFR-006, NFR-009, NFR-010
- Mapped requirement text: REQ-001 — Go 언어 및 빌드 기준을 go1.26.7로 통일한다; REQ-002 — checkout v7.0.1, setup-go v7.0.0, upload-artifact v7.0.1, setup-node v7.0.0, goreleaser-action v7.2.3을 검증된 전체 SHA로 참조한다; REQ-003 — gofmt 편차가 검증을 실패시킨다; REQ-004 — CI와 릴리스 사전 검증이 govulncheck v1.7.0으로 전체 패키지를 검사한다; REQ-005 — 도달 가능한 취약점이 검증과 후속 단계를 차단한다; NFR-006 — tidy diff, format, vet, 계층별·전체 테스트, 빌드, 취약점 검사를 통과한다; NFR-009 — 새 런타임 의존성을 추가하지 않는다; NFR-010 — Go 1.27 전용 기능 없이 go1.26.7 빌드·vet을 통과한다.
- Predecessor evidence: none — no predecessors
- Code evidence: `go.mod` :: language directive — 현재 `go 1.25.0`; `.github/workflows/ci.yml` :: Linux/Windows jobs — Go 1.25와 v6 Action 핀 사용; `.github/workflows/release.yml` :: release job — Go 1.25, 이전 setup-node 및 goreleaser 핀 사용; `.github/workflows/api-monitor.yml` 및 `.github/workflows/weekly-health.yml` :: scheduled jobs — v4/v5 Action 핀 사용; `Makefile` :: test/build targets — 포맷·tidy diff·취약점 전용 검증 없음; `.goreleaser.yml` :: before hooks — 변경을 일으키는 tidy와 테스트만 수행; `AGENTS.md` :: Project Context — Go 1.25로 기록됨.
- Test evidence: `internal/api/service_test.go` :: file formatting — 현재 gofmt 출력에 잡히는 유일한 파일; `.github/workflows/ci.yml` :: validation jobs — Linux와 Windows의 test/vet/build는 있으나 포맷·취약점 검사가 없음; `Makefile` :: test-fast/test-full/test-canary — 기존 계층별 및 전체 회귀 검증 진입점.
- Target: `.gitignore` :: repository root worktree exclusion; `go.mod` :: language directive; `.github/workflows/ci.yml`, `.github/workflows/release.yml`, `.github/workflows/api-monitor.yml`, `.github/workflows/weekly-health.yml` :: Go setup and Action pins; `Makefile` :: deterministic maintenance checks; `.goreleaser.yml` :: release preflight hooks; `internal/api/service_test.go` :: formatting; `AGENTS.md` :: active project toolchain guidance.
- Exact changes: 1) 다른 구현 파일보다 먼저 `.gitignore`에 정확한 루트 고정 `/.worktrees/` 항목을 복원한다; 2) Go 기준과 프로젝트 안내를 1.26.7로 맞춘다; 3) 사용 중인 모든 Action을 공식 태그가 가리키는 검증된 전체 SHA와 정확한 버전 주석으로 교체한다; 4) Make 검증 진입점 `verify-maintenance`에 변경 없는 tidy 검사, 빈 gofmt 목록, vet, 계층별·전체 테스트, build 및 고정 govulncheck v1.7.0을 모두 실패 전파 조건으로 추가한다; 5) Linux 및 Windows CI의 모든 검증 job과 릴리스 사전 검증이 동일한 `make verify-maintenance` 진입점을 호출하도록 하고 scheduled workflow도 동일 Go/Action 기준으로 통일한다; 6) 기존 포맷 편차를 gofmt로 정리한다; 7) 새 런타임 직접 의존성이 없고 도달 가능한 취약점이 없음을 검증한다.
- External side effects: local-only
- Non-goals: Windows 원격 workflow 실행, 원격 push·태그·배포, Weekly Health 이슈 수명주기 변경, API 기능 구현, Go 1.27 전환, 새 런타임 의존성 추가.
- Failure paths: Action 태그 SHA가 공식 원격 참조와 다르면 변경을 중단한다; go1.26.7에서 tidy·format·vet·test·build가 실패하면 체크포인트를 만들지 않는다; govulncheck가 도달 가능한 취약점을 보고하면 자동화는 실패 상태를 보존하고 후속 병합·릴리스를 차단한다; 도구 다운로드 또는 취약점 DB networking 실패는 취약점 없음으로 오인하지 않고 검증 실패로 처리한다; dependency execution이 `go.mod` 직접 의존성을 바꾸면 원인을 확인하고 승인 없는 의존성 변경을 되돌린다.
- Verification: `focused-TASK-001` => go1.26.7 기준, 정확한 Action SHA, gofmt, tidy diff, 고정 취약점 도구 및 Linux·Windows CI 두 job과 릴리스 사전 단계의 공통 `make verify-maintenance` 호출을 정적으로 확인; `integration-TASK-001` => 프로젝트 test 스킬의 계층별 테스트·전체 테스트·vet·build와 종합 유지보수 검증 통과; `risk-networking-TASK-001` => 공식 Action 태그와 로컬 SHA가 일치하고 govulncheck가 네트워크 실패 없이 결과를 생성; `risk-dependency-execution-TASK-001` => 고정 버전 도구 실행 뒤 런타임 직접 의존성 diff가 없고 취약점 결과가 성공.
- Rollback: 이 task의 단일 체크포인트 커밋을 revert하면 Go·Action·검증 자동화와 포맷 변경을 모두 이전 상태로 되돌릴 수 있으며, 외부 상태는 변경되지 않는다.
- Residual risks: GitHub-hosted Windows 실행은 TASK-012의 외부 승인 후 증거로 남는다; GitHub-hosted runner의 Action v7 호환성은 저장소 관리자가 소유하며 첫 승인된 CI run이 재시도 신호다; 취약점 DB의 일시적 가용성 문제는 성공으로 간주하지 않는다.

## Implementation Sequence

1. 필수 worktree 제외 항목을 첫 구현 변경으로 복원하고 Go 언어 기준 및 프로젝트 지침을 1.26.7로 갱신한다.
2. 공식 태그 조회 결과인 checkout `3d3c42e5aac5ba805825da76410c181273ba90b1`, setup-go `b7ad1dad31e06c5925ef5d2fc7ad053ef454303e`, upload-artifact `043fb46d1a93c77aae656e7c1c64a875d1fc6a0a`, setup-node `820762786026740c76f36085b0efc47a31fe5020`, goreleaser-action `f06c13b6b1a9625abc9e6e439d9c05a8f2190e94`를 모든 해당 workflow에 적용한다.
3. Makefile에 `go mod tidy -diff`, 빈 `gofmt -l`, vet, 계층별·전체 테스트, build, govulncheck v1.7.0을 실패 전파하는 공통 `verify-maintenance` 목표를 추가한다.
4. Linux 및 Windows CI의 모든 검증 job과 릴리스 사전 검증이 `make verify-maintenance`를 호출하도록 연결하고 release hook도 동일 구성으로 강화한 뒤 기존 포맷 편차를 정리한다.
5. Go 1.26.7에서 두 CI job 및 릴리스 연결을 포함한 집중 검증과 통합·networking·dependency execution 검증을 실행하고 결과를 task evidence로 보존한다.

## Progress Boundaries

- Implementation-entry prerequisites: 공식 Git 태그에서 다섯 Action 버전의 전체 SHA를 읽기 전용으로 확인할 것 — 확인 완료.
- Completion-critical evidence: go1.26.7의 tidy diff, 빈 gofmt 목록, vet, 계층별·전체 테스트, build, govulncheck v1.7.0 무취약점 결과, Action 핀 정적 일치, 런타임 직접 의존성 무변경.
- Deferrable environment/manual evidence: GitHub-hosted Windows workflow의 실제 성공은 TASK-012가 소유하며 사용자 외부 쓰기 승인 후 수집한다.

## Verification Evidence Contract

| Command identity | Expected success signal | Evidence destination |
|---|---|---|
| focused-TASK-001 | exit=0; output-contains=TASK_001_FOCUSED_OK | evidence/TASK-001/verification/focused.txt |
| integration-TASK-001 | exit=0; output-contains=TASK_001_INTEGRATION_OK | evidence/TASK-001/verification/integration.txt |
| risk-networking-TASK-001 | exit=0; output-contains=TASK_001_NETWORKING_OK | evidence/TASK-001/verification/risk-networking.txt |
| risk-dependency-execution-TASK-001 | exit=0; output-contains=TASK_001_DEPENDENCY_EXECUTION_OK | evidence/TASK-001/verification/risk-dependency-execution.txt |

## Plan Review Handoff

- Reviewer: plan-review
- Loop ID: 2026-08-31-repository-modernization-api-coverage-TASK-001-plan
- Raw report destination: evidence/TASK-001/plan-review/round-001.md
