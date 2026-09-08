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
| `preflight` | `make verify-maintenance`, 재사용 guardrail, `git status` 결과 |
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

태그를 붙일 정확한 커밋과 작업 상태를 먼저 확인한다. 변경이 있으면 중단한다.

```bash
git rev-parse HEAD
```
```bash
git status --porcelain
```

현재 Linux CI와 Release의 검증 기준을 실행한다. 각 명령의 종료 코드를 확인하고
하나라도 실패하면 태그를 만들지 않는다.

```bash
scripts/check-reuse-guardrails.sh .
```
```bash
make verify-maintenance
```

- `verify-maintenance`는 `go mod tidy -diff`, 포맷, vet, fast/full/canary 테스트,
  로컬 빌드, 취약점 검사를 포함한다. `go mod tidy`로 추적 파일을 수정하지 않는다.
- 루트 `naverworks` 빌드 경로가 Git 무시 대상인지 먼저 확인한다. 기존 바이너리는
  저장소 밖 임시 경로에 보존한 뒤 종료 시 복원한다. 소스 수정은 작업 트리 규칙을 따른다.
- Linux 결과는 Windows ACL 검증을 대신하지 않는다. 자격 증명 저장 관련 변경은
  대상 커밋의 Windows CI 결과를 확인하고, 미확인 상태면 게시 전에 해결한다.
- 검사 후 `git status --porcelain`과 `git rev-parse HEAD`를 다시 확인한다.
  변경이 생겼거나 검증한 커밋과 다르면 중단하고 원인을 보고한다.

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
FAIL_COUNT=0
LOOKUP_FAIL=0
NPM_ERR=$(mktemp)
trap 'rm -f "$NPM_ERR"' EXIT
for pkg in naverworks \
           @physics91org/linux-x64 @physics91org/linux-arm64 \
           @physics91org/darwin-x64 @physics91org/darwin-arm64 \
           @physics91org/win32-x64; do
  if v=$(npm --loglevel=error view "$pkg" version 2>"$NPM_ERR"); then
    if [ "$v" = "$TARGET" ]; then
      printf "✓ %-32s %s\n" "$pkg" "$v"
    else
      printf "✗ %-32s %s (expected %s)\n" "$pkg" "$v" "$TARGET"
      FAIL_COUNT=$((FAIL_COUNT+1))
    fi
  else
    err=$(cat "$NPM_ERR")
    case "$err" in
      *E404*|*"404 Not Found"*)
        printf "✗ %-32s not published (expected %s)\n" "$pkg" "$TARGET"
        FAIL_COUNT=$((FAIL_COUNT+1))
        ;;
      *)
        printf "! %-32s lookup failed: %s\n" "$pkg" "$err"
        LOOKUP_FAIL=$((LOOKUP_FAIL+1))
        ;;
    esac
  fi
done
```

상태 판정은 아래 우선순위를 따른다 (첫 번째로 매칭되는 항목 하나만 보고). 확정 실패(`FAIL_COUNT`)와 조회 실패(`LOOKUP_FAIL`)가 동시에 발생하면 둘 다 표시한다:

1. `FAIL_COUNT>0 && LOOKUP_FAIL>0`: `npm_status: partial_with_verification_failures — <FAIL_COUNT>개 실패, <LOOKUP_FAIL>개 조회 실패`. 확정 실패와 조회 실패가 섞여 있으므로 **트러블슈팅 섹션 참조** 후 두 유형을 각각 처리한다.
2. `FAIL_COUNT>0`: `npm_status: partial — <FAIL_COUNT>개 실패`. 일부 패키지가 미게시(E404) 또는 버전 불일치 — **트러블슈팅 섹션 참조** 안내.
3. `LOOKUP_FAIL>0`: `npm_status: verification_failed — <LOOKUP_FAIL>개 조회 실패`. 레지스트리/네트워크/인증 문제이므로 publish 성공 여부와 무관 — 재확인 필요.
4. 그 외: `npm_status: 6개 패키지 @<VERSION> 확인 완료`.

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
