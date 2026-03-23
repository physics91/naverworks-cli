# naverworks

네이버웍스(NAVER WORKS) REST API v1.0 명령줄 도구

## 설치

### npm (권장)

```bash
npm install -g naverworks
```

또는 npx로 바로 실행:

```bash
npx naverworks version
```

### 스크립트 (Linux/macOS)

```bash
curl -sSL https://raw.githubusercontent.com/physics91/naverworks-cli/master/install.sh | sh
```

### Go install

```bash
go install github.com/physics91/naverworks-cli@latest
```

### GitHub Releases

[Releases](https://github.com/physics91/naverworks-cli/releases)에서 플랫폼에 맞는 바이너리를 다운로드하세요.

| 플랫폼 | 아키텍처 |
|--------|---------|
| Linux | amd64, arm64 |
| macOS | amd64 (Intel), arm64 (Apple Silicon) |
| Windows | amd64 |

## 빠른 시작

### 1. 설정

```bash
# 필수 설정
naverworks config set client_id YOUR_CLIENT_ID
naverworks config set client_secret --stdin <<< "YOUR_CLIENT_SECRET"
naverworks config set bot_id YOUR_BOT_ID

# JWT 인증 시 추가 설정
naverworks config set service_account_id YOUR_SA_ID
naverworks config set private_key_path /path/to/private.pem
```

### 2. 로그인

```bash
# OAuth 2.0 (브라우저 인증)
naverworks auth login

# JWT Service Account
naverworks auth login --jwt
```

### 3. 사용

```bash
# 메시지 전송
naverworks bot send --to user@example.com --text "안녕하세요"

# 사용자 목록
naverworks directory list-users

# 일정 조회
naverworks calendar list-events \
  --user-id me \
  --calendar-id CAL_ID \
  --from 2026-03-01T00:00:00Z \
  --until 2026-03-31T23:59:59Z
```

## 명령어

| 명령어 | 설명 |
|--------|------|
| `naverworks auth login` | OAuth 2.0 로그인 |
| `naverworks auth login --jwt` | JWT 인증 |
| `naverworks auth status` | 인증 상태 확인 |
| `naverworks auth logout` | 로그아웃 |
| `naverworks config set <key> <value>` | 설정 저장 |
| `naverworks config get <key>` | 설정 조회 |
| `naverworks config list` | 전체 설정 목록 |
| `naverworks bot send` | 메시지 전송 |
| `naverworks bot get-channel <id>` | 채널 조회 |
| `naverworks bot channel-members <id>` | 채널 멤버 목록 |
| `naverworks directory list-users` | 사용자 목록 |
| `naverworks directory get-user <id>` | 사용자 상세 |
| `naverworks directory list-groups` | 그룹 목록 |
| `naverworks directory get-group <id>` | 그룹 상세 |
| `naverworks calendar list-calendars` | 캘린더 목록 |
| `naverworks calendar list-events` | 일정 목록 |
| `naverworks calendar get-event` | 일정 상세 |
| `naverworks calendar create-event` | 일정 생성 |
| `naverworks version` | 버전 정보 |

## 출력 형식

```bash
# JSON (기본)
naverworks directory list-users

# 테이블
naverworks directory list-users --output table
```

## 페이지네이션

```bash
# 첫 페이지
naverworks directory list-users --count 10

# 다음 페이지 (nextCursor 사용)
naverworks directory list-users --cursor "CURSOR_VALUE"

# 전체 자동 순회
naverworks directory list-users --all
```

## 파이프라인

```bash
# stdin에서 메시지 읽기
echo "배포 완료" | naverworks bot send --to user@example.com --text -

# jq와 조합
naverworks directory list-users | jq '.users[].userName'
```

## 환경변수

설정 파일 대신 환경변수를 사용할 수 있습니다 (환경변수가 우선):

| 환경변수 | 설정 키 |
|---------|---------|
| `NW_CLIENT_ID` | client_id |
| `NW_CLIENT_SECRET` | client_secret |
| `NW_SERVICE_ACCOUNT_ID` | service_account_id |
| `NW_PRIVATE_KEY_PATH` | private_key_path |
| `NW_DOMAIN_ID` | domain_id |
| `NW_BOT_ID` | bot_id |
| `NW_SCOPE` | scope |
| `NW_DEFAULT_CALENDAR_USER_ID` | default_calendar_user_id |

## 릴리스

```bash
git tag v0.1.0
git push origin v0.1.0
# GitHub Actions가 자동으로 크로스 플랫폼 빌드 + Release 생성
```

## 라이선스

MIT
