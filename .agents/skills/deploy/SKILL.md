---
name: deploy
description: Use when releasing a new version of naverworks — runs preflight checks, creates and pushes a SemVer tag, and verifies the GitHub Actions release workflow that builds artifacts and publishes npm packages. Triggers on "배포", "릴리스", "deploy", "release", "/deploy". If you only need to inspect or create tags, use version. If you only need local binaries, use build.
---

# naverworks 자동 배포

이 저장소는 `v*` 태그가 원격에 push되면 GitHub Actions가 크로스 플랫폼 빌드, GitHub Release 생성, npm publish를 수행한다. 이 스킬은 그 전단계 검증과 태그 push, 가능한 범위의 사후 검증을 자동으로 수행한다.

## 입출력 계약

### 입력

| 입력 | 필수 여부 | 설명 |
|------|----------|------|
| `version` | 선택 | 릴리스할 SemVer (`0.1.1` 형식). 없으면 사용자에게 묻는다 |

### 출력

| 필드 | 설명 |
|------|------|
| `tag` | 생성/푸시한 태그 (`v<version>`) |
| `preflight` | `go mod tidy`, `go test`, `go vet`, `git status` 결과 |
| `push_status` | 원격 push 여부 |
| `release_status` | GitHub Release 확인 결과 |
| `npm_status` | npm 패키지 버전 확인 결과 |

### 성공 기준

- 태그 push 전 검증이 모두 통과한다.
- `git push origin v<VERSION>`까지 성공한다.
- `gh`가 사용 가능하면 GitHub Release 또는 워크플로 상태를 확인해서 보고한다.

## 실행 규칙

1. 모든 단계를 직접 Bash로 실행한다.
2. 각 명령은 개별로 실행하고 종료 코드를 확인한다.
3. 사용자 입력이 필요한 것은 버전 번호와 태그 push 승인뿐이다.
4. 원격 태그 push는 GitHub Actions Release 워크플로를 트리거하므로 반드시 사용자 확인 후 실행한다.
5. 태그 push 이후에는 workflow가 외부에서 진행되므로, npm unpublish 같은 파괴적 롤백을 자동 수행하지 않는다.

## 절차

### Phase 0: 버전 확인

사용자가 인자로 버전을 제공했으면 사용한다. 없으면 묻는다:
- SemVer 형식 (예: `0.1.0`)
- `v` 접두사는 자동으로 붙인다

### Phase 1: 사전 검증

아래를 개별 명령으로 순서대로 실행한다. 하나라도 실패하면 태그를 만들지 않고 즉시 중단한다.

```bash
go mod tidy
```
```bash
git diff --exit-code go.mod go.sum
```
→ 종료 코드 != 0이면 `go.mod/go.sum이 변경되었습니다. 먼저 검토 후 커밋하세요`로 중단

```bash
go test ./... -count=1
```
```bash
go vet ./...
```
```bash
git status --porcelain
```
→ 출력이 있으면 `워킹 트리가 clean하지 않습니다`로 중단

간결한 보고 예시:
```
✓ go mod tidy — 변경 없음
✓ go test — PASS
✓ go vet — 이상 없음
✓ git status — clean
```

### Phase 2: 로컬 태그 생성

사용자에게 확인: `v<VERSION> 태그를 생성합니다. 진행할까요?`

승인 후:
```bash
git tag v<VERSION>
```

### Phase 3: 원격 태그 push

사용자에게 확인: `v<VERSION> 태그를 push하면 GitHub Actions Release 워크플로가 시작됩니다. 진행할까요?`

승인 후:
```bash
git push origin v<VERSION>
```

### Phase 4: 릴리스 검증

`gh`가 설치되어 있고 인증되어 있으면 아래를 사용해 검증한다.

설치 여부:
```bash
command -v gh
```

인증 여부:
```bash
gh auth status
```

Release 확인:
```bash
gh release view v<VERSION>
```
→ 즉시 보이지 않으면 10초 간격으로 몇 차례 재시도하고, 여전히 없으면 "워크플로 진행 중일 수 있음"으로 보고

npm 확인:

6개 패키지 버전이 `<VERSION>`과 **모두 일치**하는지 검증한다. 하나라도 일치하지 않으면 workflow가 부분 성공했음을 의미하므로 실패로 보고한다.

```bash
TARGET=<VERSION>
MISMATCH=0
for pkg in naverworks \
           @physics91org/linux-x64 @physics91org/linux-arm64 \
           @physics91org/darwin-x64 @physics91org/darwin-arm64 \
           @physics91org/win32-x64; do
  v=$(npm view "$pkg" version 2>/dev/null)
  if [ "$v" = "$TARGET" ]; then
    printf "✓ %-32s %s\n" "$pkg" "$v"
  else
    printf "✗ %-32s %s (expected %s)\n" "$pkg" "$v" "$TARGET"
    MISMATCH=1
  fi
done
```

- 모두 일치: `npm_status: 6개 패키지 @<VERSION> 확인 완료`
- 일부 불일치: `npm_status: partial — N개 실패`로 보고하고, **트러블슈팅 섹션 참조** 안내

`gh`를 사용할 수 없으면 태그 push 완료까지만 확정 보고하고, Release/npm 검증은 생략 사유를 함께 보고한다.

### 최종 보고

Release 확인 성공 시:
```
배포 시작 완료: v<VERSION>

push_status: pushed
GitHub Release: https://github.com/physics91/naverworks-cli/releases/tag/v<VERSION>
npm: naverworks@<VERSION> 및 플랫폼 패키지 확인 완료
```

워크플로만 시작 확인한 경우:
```
배포 시작 완료: v<VERSION>

push_status: pushed
release_status: GitHub Actions Release 워크플로 시작됨
npm_status: 아직 확인되지 않음
```

## 실패 시 대응

- 태그 push 전 실패: 필요하면 로컬 태그만 삭제한다.
  ```bash
  git tag -d v<VERSION>
  ```
- 태그 push 후 실패: workflow가 이미 Release 생성 또는 npm publish를 시작했을 수 있으므로 자동 롤백하지 않는다.
- 태그 push 후에는 같은 버전을 재사용하지 못할 수 있으니, 실제 게시 상태를 확인한 뒤 새 버전으로 재시도할지 결정한다.

## 트러블슈팅

npm publish가 `ENEEDAUTH`, `OIDC token exchange error - package not found`, goreleaser `already_exists` 등으로 실패할 경우 [`references/oidc-troubleshooting.md`](references/oidc-troubleshooting.md)를 참조한다. 증상별 원인, 디버깅 단계(`--loglevel=verbose`, `npm trust list`, OIDC JWT claim 덤프), 필수 전제 조건 체크리스트를 포함한다.
