## Review Round 1

| Criterion | Verdict | Evidence Strength | Rationale |
|-----------|---------|-------------------|-----------|
| Completeness | PASS | HIGH | 목표인 `일 1회 API 변경 모니터링`과 `주 1회 보안+의존성 체크`는 각각 Task 3, Task 4에 대응되고, 전용 baseline 브랜치/파일은 Task 1, Task 2, label 준비는 Task 2.5, 최종 검증은 Task 5에서 다룬다. 목표 수준의 요구사항 누락은 보이지 않는다. |
| Feasibility | PASS | MEDIUM | 계획은 `RemoteTrigger`, `gh`, `Go toolchain`을 명시하고, 주요 단계마다 구체 명령(`git checkout --orphan`, `gh label create`, `RemoteTrigger run`, `go install`, `go list -m -u -json all`, `go test -cover`)을 적었다. 원격 샌드박스 가용성 가정은 남아 있지만, 문서상 팀 역량 밖의 기능을 요구하지는 않는다. |
| Risk Identification | FAIL | HIGH | Risk 표는 위험을 나열하지만, 마지막 행의 `동일 에이전트 중복 실행은 cron 최소 간격(1시간)으로 방지`는 실제 완화책이 아니다. 긴 실행 시간이나 수동 재실행이 있으면 cron 간격과 무관하게 중복 실행될 수 있다. |
| Consistency | FAIL | HIGH | `Execution Rule`은 모든 원격 에이전트가 `work-main`과 `work-monitor`를 분리하고 `work-monitor`에서는 문서 수집을 하지 말라고 규정한다. 그런데 Task 3 프롬프트는 `Check out the monitoring-api branch` 후 같은 흐름에서 문서 fetch, diff, baseline update를 수행하게 되어 있어 자기모순이다. |
| Step Decomposition | FAIL | HIGH | Task 3/4의 bootstrap 항목 2는 `collect current state ... and exit`라고 쓰여 있는데, 실제 수집 방법은 뒤 단계에서 정의되어 있어 실행 순서가 모호하다. 또 Task 5 Step 3은 label 전체 목록을 조회한 뒤 `모두 0건이면 정상`이라고 판정하는데, 이번 bootstrap run과 무관한 기존 issue가 있어도 실패하게 된다. |

## Improvement Suggestions
- `Risk Identification`: `Risks & Mitigations` 마지막 행을 다음으로 교체하라. `| 동일 에이전트 중복 실행 또는 수동 재실행 중첩 | cron 간격만으로는 중복 실행을 막을 수 없다. 각 에이전트 시작 시 work-monitor에 ".run-lock.json"을 기록하고, 최근 2시간 이내 unfinished lock이 있으면 "[Ops Error] overlapping monitor run" 이슈를 생성한 뒤 종료한다. 정상 종료 시 lock의 finished_at을 갱신한다. |` 그리고 두 Remote Trigger 프롬프트 맨 앞에 다음 문장을 추가하라: `0. Before starting work, check work-monitor/.run-lock.json. If a recent unfinished lock exists, create issue titled "[Ops Error] overlapping monitor run" with label "ops-error" and exit. Otherwise create/update the lock and mark it finished at the end of the run.`
- `Consistency`: Task 3 프롬프트의 1, 5, 8, 9를 다음으로 교체하라. `1. Prepare two working directories: work-main checked out at origin/main for documentation fetch/extraction and diff generation only, and work-monitor checked out at monitoring-api for baseline read/write and git push only. Read baselines/api-snapshot.json from work-monitor.` `5. Compare the extracted snapshot from work-main with the previous snapshot in work-monitor/baselines/api-snapshot.json.` `8. Write the new JSON only to work-monitor/baselines/api-snapshot.json (always write "_bootstrap": false).` `9. In work-monitor, stage baselines/api-snapshot.json, create a commit with message "chore: update API snapshot YYYY-MM-DD", then push: git fetch origin monitoring-api && git rebase origin/monitoring-api && git push. Retry once on failure.` 또한 `## Output Rules` 아래에 `- Do NOT fetch documents or parse API fields in work-monitor`를 추가하라.
- `Step Decomposition`: Task 3/4 프롬프트의 bootstrap 항목 2를 둘 다 다음 문장으로 교체하라. `2. If "_bootstrap" is true, still execute the collection steps below, save the collected state with "_bootstrap": false, and exit without creating issues after the save completes.` 그리고 Task 5 Step 3은 다음으로 교체하라:

```markdown
**Step 3: GitHub issues 확인**

Task 3/4 테스트 실행 직전에 각각 `api_run_started_at` / `health_run_started_at`를 UTC로 기록한다. 아래 조회는 그 시각 이후에 생성된 issue만 대상으로 한다.

```bash
gh issue list --repo physics91/naverworks-cli --search "label:api-monitor created:>=$api_run_started_at"
gh issue list --repo physics91/naverworks-cli --search "label:health-check created:>=$health_run_started_at"
gh issue list --repo physics91/naverworks-cli --search "label:ops-error created:>=$api_run_started_at"
gh issue list --repo physics91/naverworks-cli --search "label:ops-error created:>=$health_run_started_at"
```

**Completion Criteria:** 위 네 명령이 모두 0건을 반환한다.
```

- `Additional item (outside rubric, advisory)`: Task 1/2의 `git checkout --orphan ...` + `git rm -rf .`는 현재 작업 트리에서 직접 실행하면 위험하다. 각 Task 첫 문장을 `브랜치 초기화는 현재 작업 트리가 아니라 임시 worktree에서 수행한다.`로 바꾸고, Step 1은 `git worktree add /tmp/naverworks-monitoring-api --detach origin/main` / `git worktree add /tmp/naverworks-monitoring-health --detach origin/main` 방식으로 분리하는 편이 안전하다.

## Verdict
NEEDS_REVISION
