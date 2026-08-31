<!-- plan-workflow: task-detail-v1 -->
# TASK-002 Detail Plan

- Task ID: TASK-002
- Big plan SHA-256: 319235e0e9e6e4b16e57de8b941ee43a35f7e0253cb39d4690a55816a64f30be
- Repository commit: 439f77c03b4fc849055a530eacb991d664b76773
- Worktree status SHA-256: e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
- Requirements: REQ-006, REQ-007, REQ-008, REQ-009, NFR-011
- Mapped requirement text: REQ-006 — Weekly Health가 모든 필수 검사에서 성공하면 자동화가 소유한 정확한 제목과 라벨의 열린 `ops-error` 이슈만 닫아야 한다; REQ-007 — 하나 이상의 필수 검사가 실패하면 동일 제목의 열린 이슈를 갱신하거나 하나만 생성해야 한다; REQ-008 — CLI에 영향이 없는 API 변경 공지의 slug를 baseline에 추가하고 로컬 검증 근거를 남겨야 한다; REQ-009 — 이슈 `#27` 기능이 완료될 때까지 해당 slug와 이슈를 미완료 추적 대상으로 유지해야 한다; NFR-011 — 저장소, 제목, 라벨, 상태를 검증한 뒤에만 이슈를 생성·갱신·종료해야 한다.
- Predecessor evidence: `evidence/TASK-001/checkpoints/439f77c03b4fc849055a530eacb991d664b76773.json` — Go 1.26.7 및 공통 유지보수 검증 진입점이 확정된 선행 체크포인트.
- Code evidence: `scripts/github_issue/main.go` :: `main`, `findOpenIssueByExactTitle`, `updateIssueBody` — 열린 이슈를 제목으로만 선택해 본문을 갱신하거나 생성하며 라벨 소유권 검증과 종료 동작이 없다; `.github/workflows/weekly-health.yml` :: `weekly-health` job — 실패/의존성 이슈 생성 단계만 있고 정상 복구 시 `ops-error` 종료 단계가 없다; `scripts/weekly_health/main.go` :: `buildReport` — 실패 이슈 제목은 `[Ops Error] weekly health check failed`, 라벨은 `ops-error`로 고정되어 있다; `docs/baselines/api-monitor-baseline.json` — `core_20260514`, `management_20260821`, `core_20260723`이 모두 아직 기준선에 없다.
- Test evidence: `scripts/github_issue/main_test.go` :: `TestFindOpenIssueByExactTitle`, `TestUpdateIssueBody`, `TestCreateIssue` — 검색·갱신·생성 기본 경로만 검사하고 저장소/라벨/상태 불일치와 중복·종료를 검사하지 않는다; `scripts/weekly_health/main_test.go` :: `TestBuildReportClassifiesOpsError` — 실패 보고서의 고정 제목과 라벨 계약을 확인한다; `scripts/api_monitor/main_test.go` :: `TestExtractReleaseNoteSlugs` — slug 추출만 검사하고 기준선 결정을 고정하지 않는다.
- Target: `scripts/github_issue/main.go` :: 이슈 선택 및 reconcile 흐름; `scripts/github_issue/main_test.go` :: 합성 GitHub API 수명주기 테스트; `.github/workflows/weekly-health.yml` :: 정상 복구 종료 단계; `docs/baselines/api-monitor-baseline.json` :: 영향 없는 공지 기준선; `docs/baselines/api-monitor-decisions.md` :: 기준선 판단 근거; `scripts/api_monitor/main_test.go` :: 처리/미처리 slug 회귀 테스트.
- Exact changes: 1. `owner/repo` 형식, 정확한 제목, 기대 라벨, 열린 상태를 검증하는 이슈 선택기를 만들고 동일한 자동화 소유 이슈가 둘 이상이거나 동일 제목의 라벨 불일치 대상만 존재하면 mutation 없이 실패시킨다. 2. open reconcile은 소유 이슈 하나의 본문을 갱신하거나 없을 때 하나만 생성하고, close reconcile은 소유 이슈 하나만 `state=closed`로 갱신하며 없으면 멱등 성공한다. 3. CLI에 `--state open|closed`를 추가하고 open에서는 body file을 요구하되 closed에서는 고정 제목·라벨만으로 안전하게 종료하게 한다. 4. Weekly Health가 정상일 때 `[Ops Error] weekly health check failed`/`ops-error`만 close reconcile하고 실패일 때 기존 open reconcile을 사용한다. 5. `core_20260514`와 `management_20260821`을 영향 없는 공지로 baseline에 추가하고 공식 공지 내용, GitHub 이슈 번호, 판단 사유를 로컬 결정 기록에 남긴다. 6. 미구현 API 묶음인 `core_20260723`은 baseline에서 제외하고 회귀 테스트로 보존한다.
- External side effects: local-only
- Non-goals: 원격 GitHub 이슈 `#27`, `#28`, `#29`, `#30`의 상태를 이번 작업에서 변경하지 않는다; 자동화가 소유하지 않은 동일 제목/다른 라벨 이슈를 변경하지 않는다; 일반 사용자가 만든 이슈를 일괄 정리하지 않는다; `core_20260723` API 기능을 이 작업에서 구현하거나 baseline 처리하지 않는다; 새 런타임 의존성을 추가하지 않는다.
- Failure paths: 잘못된 저장소 형식, 빈/중복 기대 라벨, 검색 API 실패, 동일 제목 라벨 불일치, 복수 소유 이슈, 비정상 HTTP 상태, 응답 파싱 실패는 mutation 전에 오류로 반환한다; close 대상이 없으면 멱등 성공한다; open body file이 없으면 실행 전에 실패한다; 기준선에서 `core_20260723`이 발견되면 테스트를 실패시킨다.
- Verification: `go test ./scripts/github_issue ./scripts/api_monitor ./scripts/weekly_health -count=1` => exit=0; output-contains=TASK_002_FOCUSED_OK; `make verify-maintenance` => exit=0; output-contains=TASK_002_INTEGRATION_OK; 합성 GitHub API 테스트는 실제 토큰이나 원격 mutation 없이 생성·갱신·종료·중복·대상 불일치를 검사한다.
- Rollback: TASK-002 체크포인트 커밋을 비강제 `git revert`하여 helper, workflow, baseline과 결정 기록을 함께 이전 상태로 되돌린다; 원격 상태는 변경하지 않으므로 별도 원격 롤백은 없다.
- Residual risks: GitHub Search API의 일시적 인덱싱 지연은 존재하며 자동화 소유 이슈의 중복 생성 가능성을 완전히 제거할 수 없다; 저장소 관리자 소유로, workflow 로그에서 동일 제목 중복이 보이면 자동 mutation을 중지하고 이슈 목록 API 기반 전략을 재검토한다. 실제 GitHub Actions 실행과 원격 이슈 수명주기 확인은 외부 쓰기 승인이 없는 동안 수행하지 않는다.

## Implementation Sequence
1. 저장소·제목·기대 라벨·열린 상태를 단일 소유권 계약으로 묶고 합성 API 테스트에서 정확한 검색 query와 mutation 이전 검증을 고정한다.
2. create/update/close를 하나의 멱등 reconcile 흐름으로 연결하고 open/closed CLI 입력 검증을 추가한다.
3. Weekly Health 정상/실패 분기를 reconcile CLI에 연결하고 정적 workflow 테스트로 정확한 제목·라벨·상태 인자를 고정한다.
4. `core_20260514`, `management_20260821` 기준선 결정과 공식 근거를 기록하고 `core_20260723` 제외 회귀 테스트를 추가한다.
5. focused, 보안/네트워크 표적 테스트, 전체 유지보수 검증을 실행하고 실제 외부 쓰기가 없었음을 확인한다.

## Progress Boundaries
- Implementation-entry prerequisites: TASK-001 체크포인트가 현재 HEAD이며 작업 워크트리가 비보호 파이프라인 브랜치여야 한다; GitHub 이슈 조회는 읽기 전용 근거 수집으로 끝나야 하며 원격 mutation은 금지한다.
- Completion-critical evidence: 합성 GitHub API가 성공·실패·중복·라벨 불일치·저장소 불일치·멱등 close를 결정적으로 통과해야 한다; Weekly Health workflow가 정확한 고정 제목/라벨로 정상 close와 실패 open을 호출해야 한다; 두 무영향 slug만 baseline에 있고 `core_20260723`은 없어야 한다; `make verify-maintenance`가 통과해야 한다.
- Deferrable environment/manual evidence: 실제 GitHub Actions 실행과 원격 `#29` 종료는 외부 쓰기 승인 후 저장소 관리자가 확인한다; 재시도 신호는 승인된 브랜치 push 후 Weekly Health workflow 성공 및 `#29` 자동 종료다.

## Verification Evidence Contract
| Command identity | Expected success signal | Evidence destination |
|---|---|---|
| focused-TASK-002 | exit=0; output-contains=TASK_002_FOCUSED_OK | evidence/TASK-002/verification/focused.txt |
| integration-TASK-002 | exit=0; output-contains=TASK_002_INTEGRATION_OK | evidence/TASK-002/verification/integration.txt |
| risk-networking-TASK-002 | exit=0; output-contains=TASK_002_NETWORKING_OK | evidence/TASK-002/verification/risk-networking.txt |

## Plan Review Handoff
- Reviewer: plan-review
- Loop ID: 2026-08-31-repository-modernization-api-coverage-TASK-002-plan
- Raw report destination: evidence/TASK-002/plan-review/round-001.md
