---
name: reuse-governor-local
description: Use when adding or refactoring `naverworks-cli` commands or internals and you need to manage project-specific code reuse, helper extraction, duplication cleanup, or reuse exceptions. Reviews hotspot-heavy `cmd/` and `internal/` code, decides whether to reuse an existing helper, extract a new shared helper, keep logic local, or grant a time-bounded waiver, and records catalog/lifecycle/scorecard updates. If you only need build, test, deploy, release, version, profile, or commit workflows, use the existing project-local skills instead.
---

# naverworks 재사용 거버넌스

이 저장소는 이미 공용 헬퍼가 많은데도 도메인별 보일러플레이트가 다시 생기기 쉬움. 이 스킬은 `공용화가 맞는지`, `로컬 유지가 맞는지`, `예외를 줘야 하는지`를 먼저 판단하고, 그 결정을 `catalog`, `waiver`, `decision note`, `scorecard`에 남기게 한다.

## 입출력 계약

### 입력

| 입력 | 필수 여부 | 설명 |
|------|----------|------|
| `target_files` | 필수 | 바꾸는 파일 또는 패키지 |
| `change_intent` | 필수 | 왜 수정하는지 (`새 커맨드`, `중복 제거`, `helper 추출`, `예외 유지` 등) |
| `scope` | 선택 | `cmd`, `internal`, `docs`, `mixed` |
| `force_decision` | 선택 | 이미 방향이 정해졌다면 `reuse`, `extract`, `keep-local`, `waive` 중 하나 |

### 출력

| 필드 | 설명 |
|------|------|
| `decision` | `reuse`, `extract`, `keep-local`, `waive` 중 하나 |
| `reason` | 결정 근거 1-3줄 |
| `existing_helper` | 재사용 대상 공용 헬퍼 또는 `none` |
| `follow_up` | 필수 후속 작업 목록 |
| `records` | 갱신할 catalog/waiver/decision/scorecard 파일 |

### 성공 기준

- 기존 공용 헬퍼를 먼저 검토했다.
- 무의미한 shared helper 증가를 막았다.
- 예외는 `owner`, `reason`, `expires_on` 없이 남기지 않았다.
- 새 명령을 추가하면 `cmd/smoke_test.go` 반영 여부를 함께 검토했다.

## 기본 원칙

1. `cmd/helpers.go`, `internal/api/client.go`, `internal/api/pagination.go`, `internal/fileutil/write.go` 같은 기존 헬퍼 축부터 본다.
2. 한 번만 쓰는 추상화나 동작을 가리는 추상화는 shared로 올리지 않는다.
3. false positive를 허용한다. 반복처럼 보여도 분기 구조가 다르면 `keep-local` 또는 `waive`가 더 낫다.
4. hotspot부터 정리한다. 특히 `cmd/directory.go`, `cmd/drive.go`처럼 호출량이 큰 파일을 우선한다.
5. 공용화나 예외 결정은 기록으로 남긴다. "왜 이렇게 했는지"가 없으면 다음 리팩터에서 또 같은 싸움 반복됨.

## 실행 절차

### 1. Inventory

아래를 먼저 확인한다.

- [AGENTS.md](../../../AGENTS.md)
- [cmd/helpers.go](../../../cmd/helpers.go)
- [docs/plans/2026-04-09-134538-simplify2-plan.md](../../../docs/plans/2026-04-09-134538-simplify2-plan.md)
- [docs/reviews/code/2026-04-09-134538-implementation.md](../../../docs/reviews/code/2026-04-09-134538-implementation.md)
- `docs/reuse/catalog.yaml`, `docs/reuse/waivers.yaml`가 있으면 함께 읽는다.

필요하면 `scripts/reuse-scorecard.sh`로 현재 helper adoption과 hotspot을 먼저 뽑는다.

### 2. Classify

세부 규칙은 [references/operating-model.md](references/operating-model.md)를 따른다.

| 상황 | 결정 |
|------|------|
| 기존 helper가 동작/입력 구조를 거의 그대로 덮음 | `reuse` |
| 동일 패턴이 2개 이상이고 shared로 올려도 의미가 선명함 | `extract` |
| 호출부 특수 분기나 도메인 문맥이 커서 shared helper가 오히려 흐림 | `keep-local` |
| 지금은 공용화가 과하고, 나중에 다시 볼 가치가 있음 | `waive` |

### 3. Record

`extract` 또는 `waive`면 기록을 남긴다.

- data file 스키마: [references/data-files.md](references/data-files.md)
- 템플릿 asset: `assets/catalog.yaml`, `assets/waivers.yaml`, `assets/decision-template.md`

기록 규칙:

- `extract`: catalog 엔트리 추가 또는 lifecycle 갱신
- `waive`: waiver 엔트리 추가, `expires_on` 필수
- 큰 결정: `docs/reuse/decisions/YYYY-MM-DD-<topic>.md` 생성

### 4. Enforce

다음을 같이 점검한다.

- 새 command면 `cmd/smoke_test.go` 등록 필요 여부
- 기존 helper rename/삭제면 call-site 영향 범위
- `deprecated` helper면 replacement가 catalog에 있는지
- waiver가 만료 예정이면 이번 변경에서 같이 해소 가능한지

### 5. Verify

변경 범위에 맞게 최소 검증을 고른다.

- `go test ./... -v`
- `go vet ./...`
- 특정 패키지 집중 테스트
- `scripts/reuse-scorecard.sh` 재실행

build/test/deploy 자체가 목적이면 이 스킬로 계속 밀지 말고 해당 전용 스킬로 넘긴다.

## 작업 산출물

산출물 형식은 아래처럼 짧고 구조적으로 남긴다.

```text
Decision: reuse
Existing helper: runListCmd
Reason: list pagination pattern matches current helper and avoids new closure boilerplate
Follow-up:
- addListFlags on the new list command
- verify smoke registration
Records:
- docs/reuse/catalog.yaml unchanged
- docs/reuse/waivers.yaml unchanged
```

또는

```text
Decision: waive
Existing helper: runListCmd
Reason: folder-conditional branching makes generic list helper misleading
Follow-up:
- add waiver entry with owner and expiry
- revisit during next drive refactor
Records:
- docs/reuse/waivers.yaml
- docs/reuse/decisions/YYYY-MM-DD-drive-shared-folder-list.md
```

## 리소스

- 운영 규칙: [references/operating-model.md](references/operating-model.md)
- data file 스키마: [references/data-files.md](references/data-files.md)
- 점수표 스크립트: `scripts/reuse-scorecard.sh`
- 초기 템플릿: `assets/`
