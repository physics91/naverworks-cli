---
name: version
description: Use when inspecting naverworks version state or creating/pushing release tags directly. Triggers on "버전", "version", "/version", "bump". If you want guarded preflight checks plus release monitoring, use deploy. If you only need local binaries, use build.
---

# naverworks 버전 관리

이 스킬은 AI 에이전트가 직접 실행한다. 버전 조회, 태그 목록 확인, SemVer 태그 생성/푸시를 수행한다.

## 입출력 계약

### 입력

| 입력 | 필수 여부 | 설명 |
|------|----------|------|
| 서브커맨드 | 선택 | `show`(기본), `list`, `bump <major|minor|patch>` |
| bump 종류 | `bump` 시 필수 | `major`, `minor`, `patch` 중 하나 |

### 출력

| 필드 | 설명 |
|------|------|
| `latest_tag` | 가장 최근 Git 태그 (`v*`) |
| `local_build_version` | 현재 워킹트리에서 `make build` 후 `./naverworks version` 결과 |
| `npm_package_version` | `npm/cli/package.json`에 기록된 버전 |
| `new_tag` | `bump` 실행 시 생성한 새 태그 |
| `push_status` | 원격 push 여부 |

### 성공 기준

- `show`: 최신 태그, 로컬 빌드 메타데이터, npm 패키지 버전을 보고한다.
- `list`: `v*` 태그 목록을 정렬하여 보고한다.
- `bump`: 새 태그를 계산하고, 사용자 승인 범위까지 로컬 생성 또는 원격 push를 완료한다.

## 실행 규칙

1. 모든 명령을 직접 Bash로 실행한다.
2. 각 명령은 개별로 실행하고 종료 코드를 확인한다.
3. 태그 생성 전과 원격 push 전에는 사용자에게 확인을 받는다.
4. `v*` 태그를 원격에 push하면 GitHub Actions Release 워크플로가 자동 시작된다는 점을 명시한다.

## 인자별 동작

### 인자 없음 또는 `show`: 현재 버전 상태 조회

최신 Git 태그 확인:
```bash
git describe --tags --abbrev=0 2>/dev/null
```
→ 태그가 없으면 "태그 없음"으로 보고

현재 워킹트리 빌드 메타데이터 확인:
```bash
make build
```
```bash
./naverworks version
```
```bash
rm -f naverworks
```
→ `Makefile` 기본값 때문에 `VERSION`을 넘기지 않으면 `version` 필드는 보통 `dev`다.

npm 패키지 버전 확인:
```bash
grep '"version"' npm/cli/package.json
```

보고 예시:
```
현재 버전 상태:
  latest_tag: v0.1.0
  local_build_version: {"version":"dev","commit":"abc1234","build_date":"..."}
  npm_package_version: 0.1.0
```

### `list`: 태그 목록

```bash
git tag -l 'v*' --sort=-v:refname
```

### `bump <major|minor|patch>`: 태그 범프

#### Phase 1: 현재 버전 확인

```bash
git describe --tags --abbrev=0 2>/dev/null
```
→ 태그가 없으면 `v0.0.0`에서 시작
→ `v` prefix를 제거하여 현재 SemVer 추출

#### Phase 2: 새 버전 계산

에이전트가 SemVer를 파싱하여 지정된 부분을 증가:

| 현재 | bump | 결과 |
|------|------|------|
| 0.1.0 | patch | 0.1.1 |
| 0.1.0 | minor | 0.2.0 |
| 0.1.0 | major | 1.0.0 |

- patch: PATCH +1
- minor: MINOR +1, PATCH = 0
- major: MAJOR +1, MINOR = 0, PATCH = 0

#### Phase 3: 사전 검증

```bash
go test ./... -count=1
```
```bash
go vet ./...
```
```bash
git status --porcelain
```
→ `git status --porcelain` 출력이 비어있지 않으면 "워킹 트리가 clean하지 않습니다"로 중단

#### Phase 4: 태그 생성

사용자에게 확인: `v<NEW_VERSION> 태그를 로컬에 생성합니다. 진행할까요?`

승인 후:
```bash
git tag v<NEW_VERSION>
```

#### Phase 5: 선택적 원격 push

사용자에게 확인: `원격에 push하면 GitHub Actions Release 워크플로가 자동 시작됩니다. push할까요?`

승인 후:
```bash
git push origin v<NEW_VERSION>
```

### 보고

push 안 함:
```
버전 범프 완료:
  previous_tag: v0.1.0
  new_tag: v0.2.0
  push_status: local only

다음 단계: 릴리스까지 진행하려면 /deploy 또는 git push origin v0.2.0
```

push 함:
```
버전 범프 완료:
  previous_tag: v0.1.0
  new_tag: v0.2.0
  push_status: pushed

다음 단계: GitHub Actions Release 워크플로 상태를 확인하세요.
```
