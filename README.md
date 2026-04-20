# naverworks

네이버웍스(NAVER WORKS) REST API v1.0용 CLI 도구

`naverworks`는 프로필 기반 인증(OAuth 2.0, JWT Service Account), 자동 토큰 갱신, JSON/테이블 출력, 페이지네이션 순회를 지원하는 NAVER WORKS CLI입니다.

## 핵심 기능

- 인증: `auth setup`, `auth login`, `auth refresh`, `auth status`, `auth logout`
- 설정/프로필: `config set|get|list`, `--profile`, `NW_PROFILE`
- 주요 도메인: `bot`, `calendar`, `directory`, `drive`, `mail`, `approval`, `task`, `board`, `contact`, `attendance`, `audit`, `monitoring`, `scim`
- 출력: pretty JSON 기본 출력, 일부 목록형 명령의 `--output table`
- 페이지네이션: `--count`, `--cursor`, `--all`
- API 요청 미리보기: `--dry-run`, `--plan-out`, `--generate-input`

## 설치

### npm

```bash
npm install -g naverworks
```

### npx

```bash
npx naverworks version
```

### 설치 스크립트

```bash
curl -sSL https://raw.githubusercontent.com/physics91/naverworks-cli/main/install.sh | sh
```

기본 설치 경로는 `/usr/local/bin`이며, `INSTALL_DIR`로 변경할 수 있습니다.

추가 설치 방법과 플랫폼별 바이너리 정보는 [Installation](https://github.com/physics91/naverworks-cli/wiki/Installation) 문서를 보면 됩니다.

## 30초 시작

```bash
# 대화형 설정
naverworks auth setup

# 로그인
naverworks auth login

# 상태 확인
naverworks auth status

# 첫 API 호출
naverworks directory list-users --count 20
```

바로 다른 도메인도 써먹을 수 있습니다.

```bash
naverworks bot send --to USER_ID --text "배포 완료"
naverworks drive list --user-id me
naverworks mail send --user-id me --to user@example.com --subject "배포 완료" --body "운영 반영 끝"

# API 요청 미리보기
naverworks --dry-run bot send --bot-id BOT_ID --to USER_ID --text "배포 완료"
naverworks --dry-run directory list-users --count 20
```

## 문서

- [User Guide Home](https://github.com/physics91/naverworks-cli/wiki)
- [Installation](https://github.com/physics91/naverworks-cli/wiki/Installation)
- [Quick Start](https://github.com/physics91/naverworks-cli/wiki/Quick-Start)
- [Authentication and Profiles](https://github.com/physics91/naverworks-cli/wiki/Authentication-and-Profiles)
- [Configuration Keys and Environment Variables](https://github.com/physics91/naverworks-cli/wiki/Configuration-Keys-and-Environment-Variables)
- [Output and Pagination](https://github.com/physics91/naverworks-cli/wiki/Output-and-Pagination)
- [Domain Command Guide](https://github.com/physics91/naverworks-cli/wiki/Domain-Command-Guide)
- [SCIM](https://github.com/physics91/naverworks-cli/wiki/SCIM)
- [Troubleshooting](https://github.com/physics91/naverworks-cli/wiki/Troubleshooting)
- [Releases](https://github.com/physics91/naverworks-cli/releases)

상세 문서는 GitHub wiki에서 읽고, 원본은 `docs/wiki/`에서 관리합니다.

전체 명령은 아래처럼 확인할 수 있습니다.

```bash
naverworks --help
naverworks <command> --help
```

## 개발 검증

빠른 회귀 확인과 전체 검증을 분리해서 돌릴 수 있습니다.

```bash
make test-fast   # 핵심 unit/contract + meta + 대표 journey
make test-full   # 전체 테스트 스위트
make build
go vet ./...
```

## 자동 점검

GitHub Actions로 주기 점검을 돌립니다. 두 workflow 모두 `gh` CLI에 의존하지 않고 `${{ secrets.GITHUB_TOKEN }}` 기반 GitHub API 호출로 이슈를 생성합니다.

- `API Change Monitor` (`.github/workflows/api-monitor.yml`): NAVER WORKS 릴리즈 노트를 매일 확인하고 `docs/baselines/api-monitor-baseline.json`에 없는 새 공지를 `api-monitor` 이슈로 등록
- `Weekly Health` (`.github/workflows/weekly-health.yml`): `make test-full`, `go vet ./...`, `make build`, `go list -m -u -json all` 결과를 주간 점검하고 실패 시 `ops-error`, 업데이트 가능 모듈이 있으면 `health-check` 이슈로 등록

## 라이선스

MIT
