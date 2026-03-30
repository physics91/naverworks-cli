# naverworks

네이버웍스(NAVER WORKS) REST API v1.0용 CLI 도구

프로필 기반 인증(OAuth 2.0, JWT Service Account), 자동 토큰 갱신, JSON/테이블 출력, 페이지네이션 순회를 지원합니다.

## 지원 기능

- 인증: `auth setup`, `auth login`, `auth refresh`, `auth status`, `auth logout`
- 설정/프로필: `config set|get|list`, `--profile`, `NW_PROFILE`
- 협업 API: `bot`, `calendar`, `board`, `note`, `task`, `approval`, `form`
- 조직/관리 API: `directory`, `contact`, `attendance`, `hr`, `business-place`, `audit`, `monitoring`
- 파일/메일 API: `drive`, `drive shared`, `drive group`, `drive shared-folder`, `mail`
- SCIM: `scim` 도메인 전체 CRUD
- 기타: `version`, `completion`

## 설치

### npm

```bash
npm install -g naverworks
```

또는 `npx`로 바로 실행:

```bash
npx naverworks version
```

### 설치 스크립트

```bash
curl -sSL https://raw.githubusercontent.com/physics91/naverworks-cli/master/install.sh | sh
```

기본 설치 경로는 `/usr/local/bin`이며, `INSTALL_DIR`로 변경할 수 있습니다.

### Go install

```bash
go install github.com/physics91/naverworks-cli@latest
```

### GitHub Releases

[Releases](https://github.com/physics91/naverworks-cli/releases)에서 바이너리를 직접 받을 수 있습니다.

| 플랫폼 | 아키텍처 |
| --- | --- |
| Linux | amd64, arm64 |
| macOS | amd64, arm64 |
| Windows | amd64 |

## 빠른 시작

### 1. 대화형 설정

가장 덜 귀찮은 방법은 이거임

```bash
naverworks auth setup
```

설정 저장 후 바로 로그인까지 이어서 진행할 수 있습니다.

### 2. 수동 설정

```bash
# 공통
naverworks config set client_id YOUR_CLIENT_ID
naverworks config set client_secret --stdin <<< "YOUR_CLIENT_SECRET"

# Bot API를 쓸 때
naverworks config set bot_id YOUR_BOT_ID

# JWT 인증을 쓸 때
naverworks config set service_account_id YOUR_SERVICE_ACCOUNT_ID
naverworks config set private_key_path /path/to/private.pem

# Calendar에서 기본 user-id를 쓰고 싶을 때
naverworks config set default_calendar_user_id me
```

유효한 설정 키 목록:

- `client_id`
- `client_secret`
- `service_account_id`
- `private_key_path`
- `domain_id`
- `bot_id`
- `scope`
- `default_calendar_user_id`
- `scim_access_token`

### 3. 로그인

```bash
# OAuth 2.0
naverworks auth login

# JWT Service Account
naverworks auth login --jwt

# 현재 상태 확인
naverworks auth status
```

### 4. 바로 써먹기

```bash
# Bot 메시지 전송
naverworks bot send --to USER_ID --text "배포 완료"

# 사용자 목록
naverworks directory list-users --count 20

# 일정 조회
naverworks calendar list-events \
  --user-id me \
  --calendar-id CALENDAR_ID \
  --from 2026-03-01T00:00:00Z \
  --until 2026-03-31T23:59:59Z

# 내 드라이브 파일 목록
naverworks drive list --user-id me

# 메일 전송
naverworks mail send \
  --user-id me \
  --to user@example.com \
  --subject "배포 완료" \
  --body "운영 반영 끝"
```

## 프로필과 설정 파일

프로필 우선순위는 아래 순서입니다.

1. `--profile`
2. `NW_PROFILE`
3. 설정 파일의 `current_profile`
4. `default`

예시:

```bash
# staging 프로필에 설정 저장
naverworks --profile staging config set client_id STAGING_CLIENT_ID
naverworks --profile staging config set client_secret --stdin <<< "STAGING_SECRET"
naverworks --profile staging auth login

# staging 프로필로 명령 실행
naverworks --profile staging directory list-users

# 환경변수로 기본 프로필 지정
export NW_PROFILE=staging
naverworks bot send --to USER_ID --text "hello"
```

기본 저장 위치:

- Linux/macOS: `~/.config/naverworks/config.json`, `~/.config/naverworks/token.json`
- Windows: `%APPDATA%\\naverworks\\config.json`, `%APPDATA%\\naverworks\\token.json`

## 환경변수

환경변수가 설정 파일보다 우선합니다.

| 환경변수 | 설명 |
| --- | --- |
| `NW_PROFILE` | 활성 프로필명 |
| `NW_CLIENT_ID` | Client ID |
| `NW_CLIENT_SECRET` | Client Secret |
| `NW_SERVICE_ACCOUNT_ID` | Service Account ID |
| `NW_PRIVATE_KEY_PATH` | Private Key 경로 |
| `NW_DOMAIN_ID` | 도메인 ID |
| `NW_BOT_ID` | Bot ID |
| `NW_SCOPE` | OAuth/JWT scope |
| `NW_DEFAULT_CALENDAR_USER_ID` | Calendar 기본 user-id |
| `NW_SCIM_ACCESS_TOKEN` | SCIM 전용 액세스 토큰 |

## 출력과 페이지네이션

기본 출력은 pretty JSON입니다.

```bash
naverworks directory list-users
```

일부 목록형 명령은 테이블 출력도 지원합니다.

```bash
naverworks directory list-users --output table
```

페이지네이션 플래그:

```bash
# 첫 페이지
naverworks directory list-users --count 10

# 다음 페이지
naverworks directory list-users --cursor "NEXT_CURSOR"

# 가능한 페이지 전부 자동 순회
naverworks directory list-users --all
```

## SCIM

SCIM은 일반 로그인 토큰이 아니라 별도 `scim_access_token` 설정을 사용합니다.

```bash
naverworks config set scim_access_token --stdin <<< "YOUR_SCIM_TOKEN"
naverworks scim list-users
naverworks scim get-user USER_ID
```

## 주요 도메인 예시

```bash
# 결재
naverworks approval list --user-id me

# 태스크
naverworks task list --user-id me --all

# 게시판
naverworks board list

# 연락처
naverworks contact list

# 근태
naverworks attendance status --user-id me

# 감사 로그
naverworks audit download-logs \
  --start-time 2026-03-01T00:00:00Z \
  --end-time 2026-03-31T23:59:59Z
```

전체 명령 목록은 아래로 확인하면 됩니다.

```bash
naverworks --help
naverworks <command> --help
```

자동완성 스크립트도 생성할 수 있습니다.

```bash
naverworks completion bash
naverworks completion zsh
```

## 개발

```bash
make build
make test
go vet ./...
```

## 릴리스

```bash
git tag v0.1.0
git push origin v0.1.0
```

태그 푸시 후 GitHub Actions + GoReleaser로 바이너리와 릴리스 자산이 생성됩니다.

## 라이선스

MIT
