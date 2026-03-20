---
name: version
description: Use when managing nw-cli versions — bump version, list tags, check current version. Triggers on "버전", "version", "/version", "bump".
---

# nw-cli 버전 관리

이 스킬은 AI 에이전트가 직접 실행한다. 버전 조회, 범프, 태그 관리를 수행한다.

## 실행 규칙

1. 모든 명령을 직접 Bash로 실행한다.
2. 인자에 따라 동작이 달라진다.
3. 태그 생성/push 전에 사용자에게 확인을 받는다.

## 인자별 동작

### 인자 없음 또는 `show`: 현재 버전 조회

```bash
git describe --tags --abbrev=0 2>/dev/null || echo "태그 없음"
```
```bash
go build -o /tmp/nw-cli-ver .
```
```bash
/tmp/nw-cli-ver version
```
```bash
rm -f /tmp/nw-cli-ver
```
```bash
cat npm/cli/package.json | grep '"version"'
```

보고:
```
현재 버전:
  git 태그: v0.1.0
  바이너리: {"version":"0.1.0","commit":"abc1234","build_date":"..."}
  npm package.json: 0.1.0
```

### `list`: 태그 목록

```bash
git tag -l 'v*' --sort=-v:refname
```

### `bump <major|minor|patch>`: 버전 범프

현재 최신 태그에서 SemVer를 추출하고 지정된 부분을 증가:

Phase 1: 현재 버전 확인
```bash
git describe --tags --abbrev=0 2>/dev/null
```
→ 태그가 없으면 `v0.0.0`에서 시작

Phase 2: 새 버전 계산

| 현재 | bump | 결과 |
|------|------|------|
| 0.1.0 | patch | 0.1.1 |
| 0.1.0 | minor | 0.2.0 |
| 0.1.0 | major | 1.0.0 |

Phase 3: 사전 검증
```bash
go test ./... -count=1
```
```bash
go vet ./...
```
```bash
git status --porcelain
```
→ 하나라도 실패하면 중단

Phase 4: 태그 생성

**사용자에게 확인**: "v<NEW_VERSION> 태그를 생성합니다. 진행할까요?"

승인 후:
```bash
git tag v<NEW_VERSION>
```

**사용자에게 확인**: "원격에 push할까요?"

승인 후:
```bash
git push origin v<NEW_VERSION>
```

### 보고

```
버전 범프 완료:
  이전: v0.1.0
  현재: v0.2.0
  태그: v0.2.0 (로컬 + 원격 push 완료)

다음 단계: /deploy로 배포하세요.
```
