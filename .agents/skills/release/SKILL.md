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

| 입력 | 필수 여부 | 설명 |
|------|----------|------|
| `action` | 필수 | `list`, `view`, `edit`, `delete`, `rollback` 중 하나 |
| `tag` | 조건부 | `view`, `edit`, `delete`, `rollback` 시 필수 (예: `v0.2.0`) |
| `edit_flags` | 조건부 | `edit` 시 수정할 항목 (예: `--notes`, `--draft`, `--prerelease`, `--latest`) |
| `previous_tag` | 조건부 | `rollback` 시 필수. latest로 되돌릴 이전 릴리스 태그 (예: `v0.1.0`) |

### 출력

| 필드 | 설명 |
|------|------|
| `result` | 실행 결과 요약 |
| `release_url` | 해당 릴리스 GitHub URL |

### 성공 기준

- `gh` CLI가 인증된 상태에서 요청된 작업이 완료된다.
- 파괴적 작업(`delete`, `rollback`)은 사용자 확인 후에만 실행한다.

## 사전 조건

```bash
command -v gh >/dev/null 2>&1 || { echo "gh CLI가 설치되어 있지 않습니다"; exit 1; }
command -v jq >/dev/null 2>&1 || { echo "jq가 설치되어 있지 않습니다"; exit 1; }
gh auth status
```

## 절차

### 1. list — 릴리스 목록 조회

```bash
gh release list --limit 10
```

### 2. view — 특정 릴리스 상세 조회

```bash
gh release view <TAG>
```

릴리스 노트, 에셋 목록, 생성 일시를 확인한다.

### 3. edit — 릴리스 수정

릴리스 메타데이터를 수정한다. 수정 가능 항목:

```bash
# 릴리스 노트 수정
gh release edit <TAG> --notes "새로운 릴리스 노트"

# 파일에서 릴리스 노트 읽기
gh release edit <TAG> --notes-file ./RELEASE_NOTES.md

# draft ↔ published 전환
gh release edit <TAG> --draft=false

# prerelease 플래그 토글
gh release edit <TAG> --prerelease=true

# latest 릴리스 지정
gh release edit <TAG> --latest=true
```

### 4. delete — 릴리스 삭제

**파괴적 작업**: 반드시 사용자 확인 후 실행한다.

```bash
# 릴리스만 삭제 (태그 유지)
gh release delete <TAG> --yes

# 릴리스 + 태그 삭제
gh release delete <TAG> --yes --cleanup-tag
```

삭제 후 npm 패키지는 이미 publish된 상태이므로 자동 롤백되지 않는다. npm unpublish가 필요하면 별도로 안내한다.

### 5. rollback — 이전 버전으로 롤백

롤백은 여러 단계로 구성된다:

1. **현재 릴리스 확인**
   ```bash
   gh release list --limit 5
   ```

2. **문제 릴리스를 latest에서 해제**
   ```bash
   gh release edit <PROBLEM_TAG> --latest=false
   ```

3. **이전 릴리스를 latest로 지정**
   ```bash
   gh release edit <PREVIOUS_TAG> --latest=true
   ```

4. **npm 패키지 확인** (이미 publish된 버전은 수동 대응 필요)
   ```bash
   npm view naverworks versions --json | jq '.[-5:]'
   ```

5. 필요시 문제 릴리스 삭제
   ```bash
   gh release delete <PROBLEM_TAG> --yes --cleanup-tag
   ```

## Best Practices

- goreleaser가 생성한 릴리스 노트는 Conventional Commits 기반 자동 생성이므로, 수동 수정 시 형식을 유지한다
- `--cleanup-tag`로 태그까지 삭제하면 같은 버전 태그를 재사용할 수 있지만, npm에 이미 publish된 버전은 재사용 불가하므로 새 버전을 권장한다
- prerelease 태그(`v1.0.0-rc.1`)는 `--prerelease` 플래그와 함께 사용한다
- 릴리스 에셋(바이너리)을 개별 교체해야 하면 `gh release upload <TAG> <FILE> --clobber`를 사용한다

## Caveats

- npm에 publish된 패키지는 GitHub Release 삭제만으로 제거되지 않음. npm unpublish는 72시간 이내에만 가능하고, 의존하는 패키지가 있으면 불가함
- goreleaser가 생성한 체크섬 파일(`checksums.txt`)은 릴리스 에셋과 함께 관리됨. 에셋을 개별 교체하면 체크섬 불일치 발생 가능
- GitHub Actions Release 워크플로가 진행 중인 상태에서 릴리스를 삭제하면 워크플로가 실패할 수 있음
