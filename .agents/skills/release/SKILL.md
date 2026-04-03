---
name: release
description: >
  Manage GitHub Releases for naverworks — view, edit, delete, and rollback
  published releases using gh CLI. Triggers on "릴리스 관리", "release manage",
  "release edit", "release delete", "release rollback". If you need to create a new release (tag +
  push + publish), use deploy. For local binaries, use build. For running
  tests, use test. For version inspection or tag creation, use version.
---

# Release 관리

GitHub Release 조회·수정·삭제·롤백을 수행한다. 새 릴리스 생성은 `deploy` 스킬이 담당하므로 여기서는 이미 생성된 릴리스를 관리하는 데 집중한다.

## 입출력 계약

### 입력

| 입력 | 필수 여부 | 형식 | 설명 |
|------|----------|------|------|
| `action` | 필수 | enum: `list`, `view`, `edit`, `delete`, `rollback` | 수행할 작업 |
| `tag` | 조건부 | string, SemVer (`v0.2.0`) | `view`, `edit`, `delete` 시 필수. `rollback` 시 문제 릴리스 태그 |
| `previous_tag` | 조건부 | string, SemVer (`v0.1.0`) | `rollback` 시 필수. latest로 되돌릴 이전 릴리스 태그 |
| `edit_flags` | 조건부 | string | `edit` 시 필수. `gh release edit`에 전달할 플래그 문자열. 예: `--notes "수정" --draft=false`. 사용 가능한 플래그: `--notes`, `--notes-file`, `--draft`, `--prerelease`, `--latest` |
| `cleanup_tag` | 선택 | boolean, default: false | `delete` 시 태그까지 삭제할지 여부 |
| `delete_after_rollback` | 선택 | boolean, default: false | `rollback` 후 문제 릴리스를 삭제할지 여부 |

### 출력 (액션별)

AI 에이전트는 `gh` CLI 출력을 사용자에게 보여주고, 아래 필드를 ✓/✗ 요약에 포함한다.

| Action | 보고 필드 |
|--------|----------|
| `list` | 릴리스 목록 (tag, title, latest 여부) |
| `view` | tag, title, 릴리스 노트, 에셋 목록 |
| `edit` | tag, 변경된 필드 |
| `delete` | 삭제된 tag, cleanup_tag 여부 |
| `rollback` | problem_tag, previous_tag, 새 latest tag, 삭제 여부 |

### 성공 기준

- `gh` CLI가 인증된 상태에서 요청된 작업이 완료된다.
- 파괴적 작업(`delete`, `rollback`)은 사용자 확인 후에만 실행한다.

### 사전 조건

```bash
command -v gh >/dev/null 2>&1 || { echo "gh CLI가 설치되어 있지 않습니다"; exit 1; }
command -v jq >/dev/null 2>&1 || { echo "jq가 설치되어 있지 않습니다"; exit 1; }
gh auth status
```

## 비파괴 작업 (list / view / edit)

### list

1. `gh release list --limit 10`
2. 결과 확인: 릴리스 태그, 제목, latest 여부를 사용자에게 보고

### view

1. `gh release view <TAG>`
2. 결과 확인: 릴리스 노트, 에셋 목록, 생성 일시를 사용자에게 보고

### edit

1. 수정 전 확인: `gh release view <TAG>` — 현재 상태 파악
2. 사용자 확인: 변경할 항목과 값을 보여주고 승인 요청 (특히 `--draft`, `--latest`, `--prerelease`는 릴리스 가시성에 영향)
3. 수정 실행: 아래 플래그 중 해당 항목을 `gh release edit <TAG>`에 전달
   - `--notes "..."` — 릴리스 노트 수정
   - `--notes-file ./path` — 파일에서 릴리스 노트 읽기
   - `--draft=true|false` — draft 상태 전환
   - `--prerelease=true|false` — prerelease 플래그 토글
   - `--latest=true|false` — latest 릴리스 지정
4. 수정 후 검증: `gh release view <TAG>` — 변경 사항 반영 확인
5. 결과 보고: 변경된 필드를 ✓ 형식으로 요약

## 파괴 작업 (delete / rollback)

> **자동 호출 금지.** 이 스킬은 사용자의 명시적 릴리스 관리 요청이 있을 때만 실행한다.
> **사용자 확인 필수.** 파괴적 작업이므로 실행 전 반드시 사용자에게 확인한다.

### delete

1. 삭제 대상 확인: `gh release view <TAG>`
2. 사용자 확인 요청: 태그까지 삭제할지 결정 → `cleanup_tag` 값 확정
3. 삭제 실행:
   - `cleanup_tag=false`: `gh release delete <TAG> --yes`
   - `cleanup_tag=true`: `gh release delete <TAG> --yes --cleanup-tag`
4. 삭제 확인: `gh release list --limit 5`
5. 결과 보고: 삭제된 태그와 cleanup 여부를 ✓ 형식으로 요약

- npm에 publish된 패키지는 GitHub Release 삭제만으로 제거되지 않음 (npm unpublish는 72시간 이내만 가능)
- GitHub Actions 워크플로 진행 중 삭제 시 워크플로 실패 가능

### rollback

1. 현재 릴리스 확인: `gh release list --limit 5`
2. 사용자 확인 요청: `<TAG>를 latest에서 해제하고 <PREVIOUS_TAG>를 latest로 지정합니다`
3. 문제 릴리스 latest 해제: `gh release edit <TAG> --latest=false`
4. 이전 릴리스 latest 지정: `gh release edit <PREVIOUS_TAG> --latest=true`
5. 검증: `gh release list --limit 5` — latest 태그 확인
6. npm 패키지 상태 확인: `npm view naverworks versions --json | jq '.[-5:]'`
7. `delete_after_rollback=true`이면: `gh release delete <TAG> --yes --cleanup-tag`
8. 결과 보고: problem_tag, previous_tag, 새 latest, 삭제 여부를 ✓ 형식으로 요약

- `--cleanup-tag`로 태그까지 삭제하면 재사용 가능하지만, npm에 이미 publish된 버전은 재사용 불가 → 새 버전 권장

## 호출 예시

Good:
> "v0.2.0 릴리스 삭제하고 v0.1.0으로 롤백해줘" → `release` 스킬 (rollback action)

Bad:
> "v0.3.0으로 새 릴리스 만들어줘" → `deploy` 스킬 (새 릴리스 생성은 deploy 담당)

<!-- codex-review: APPROVED | round: 6 | date: 2026-04-03 -->
