# 네이버웍스 CLI 요구사항 정의서

## 1. 개요

| 항목 | 내용 |
|------|------|
| 프로젝트명 | `naverworks-cli` (실행 명령: `naverworks`) |
| 기술 스택 | Go + cobra |
| 초기 버전 | v0.1 |
| 대상 사용자 | 네이버웍스 관리자, 개발자, 자동화 스크립트 |
| 참조 SDK | `naverworks-sdk-kotlin` (API 스펙 참조용) |

---

## 2. 기능 요구사항

### 2.1 인증 (`naverworks auth`)

| ID | 요구사항 | 우선순위 |
|----|---------|---------|
| FR-01 | `naverworks auth login` — OAuth 2.0 인증. 로컬에 임시 HTTP 서버(포트 8484, 사용 중이면 8485~8494 순차 탐색)를 띄우고 브라우저를 열어 인증 코드를 수신한다. `state` 파라미터를 crypto/rand로 생성하여 CSRF를 방지한다. 콜백 타임아웃은 120초이며, 초과 시 서버를 종료하고 에러를 반환한다. 브라우저 실행 실패 시 URL을 stderr에 출력하여 수동 입력을 허용한다. `redirect_uri`는 `http://localhost:{port}/callback`이며, 네이버웍스 Developer Console에 사전 등록이 필요하다. | P0 |
| FR-02 | `naverworks auth login --jwt` — config에 설정된 client_id, client_secret, service_account_id, private_key_path, scope로 JWT Service Account 인증을 수행한다. private_key 파일 권한 검증: Linux/macOS에서는 0600이 아니면 경고를 stderr에 출력하고, Windows에서는 현재 사용자 외 ACE가 있으면 경고한다. 암호화된 키나 잘못된 PEM 형식은 명확한 에러 메시지를 반환한다. | P0 |
| FR-03 | `naverworks auth status` — 현재 인증 상태를 JSON으로 stdout에 출력한다. 출력 필드: `auth_method` (oauth\|jwt), `expires_at` (RFC3339), `scopes` (문자열 배열). OAuth 모드에서 scope에 `openid profile`이 포함되어 있으면 OIDC userinfo로 `user_name`을 추가하고, 미포함 시 `user_name` 필드를 생략한다. JWT 모드에서는 `service_account_id`를 출력한다. access token/refresh token 값은 절대 출력하지 않는다. | P0 |
| FR-04 | `naverworks auth logout` — 로컬 토큰 파일을 삭제한다. OAuth 모드에서는 서버측 token revoke(`POST /oauth2/v2.0/revoke`)를 먼저 시도한다. revoke 순서: (1) refresh_token을 `token_type_hint=refresh_token`으로 폐기, (2) access_token을 `token_type_hint=access_token`으로 폐기. 각 단계 실패 시 경고를 stderr에 출력하되 다음 단계와 로컬 삭제는 계속 진행한다. | P1 |
| FR-05 | access token 만료 60초 전 또는 API 401 응답 시(1회에 한해), refresh token으로 자동 갱신한다. 갱신 실패 시 stderr에 에러를 출력하고 종료코드 1을 반환한다. 401 재시도는 무한 루프 방지를 위해 요청당 최대 1회로 제한한다. | P0 |

### 2.2 설정 (`naverworks config`)

| ID | 요구사항 | 우선순위 |
|----|---------|---------|
| FR-06 | `naverworks config set <key> <value>` — 설정 가능 키: `client_id`, `client_secret`, `service_account_id`, `private_key_path`, `domain_id`, `bot_id`, `scope`, `default_calendar_user_id`. 민감 키(`client_secret`)는 `naverworks config set client_secret --stdin`으로 stdin 입력을 지원하여 쉘 히스토리 노출을 방지한다. | P0 |
| FR-07 | `naverworks config get <key>` — 설정값을 조회한다. 민감 정보(`client_secret`, `private_key_path`의 내용)는 마스킹(`****`)하여 출력한다. | P0 |
| FR-08 | `naverworks config list` — 전체 설정 목록을 JSON으로 출력한다. 민감 정보는 마스킹한다. | P1 |
| FR-09 | 설정 파일 경로: Linux/macOS는 `~/.config/naverworks/config.json` (권한 0600), Windows는 `%APPDATA%\naverworks\config.json` (부모 디렉토리 포함 현재 사용자만 접근 가능하도록 ACL 설정). | P0 |
| FR-10 | 환경변수 오버라이드: `NW_CLIENT_ID`, `NW_CLIENT_SECRET`, `NW_SERVICE_ACCOUNT_ID`, `NW_PRIVATE_KEY_PATH`, `NW_DOMAIN_ID`, `NW_BOT_ID`, `NW_SCOPE`, `NW_DEFAULT_CALENDAR_USER_ID`. 환경변수 > 설정 파일 순으로 우선한다. | P1 |

### 2.3 Bot (`naverworks bot`)

Bot API 호출에는 `bot_id`가 필수이며, `config set bot_id <id>` 또는 `NW_BOT_ID`로 설정한다. 미설정 시 stderr에 에러를 출력하고 종료코드 1을 반환한다.

| ID | 요구사항 | 우선순위 |
|----|---------|---------|
| FR-11 | `naverworks bot send --to <userId> --text <message>` — 사용자에게 텍스트 메시지를 전송한다. 성공 시 API 응답 전체를 JSON으로 stdout에 출력한다 (응답에 메시지 ID가 포함되지 않을 수 있으며, 이 경우 HTTP 201 상태와 빈 객체를 출력한다). | P0 |
| FR-12 | `naverworks bot send --channel <channelId> --text <message>` — 채널에 텍스트 메시지를 전송한다. 성공 시 API 응답 전체를 JSON으로 stdout에 출력한다. | P0 |
| FR-13 | `naverworks bot send --to <userId> --file <path>` — 파일 첨부 전송. 3단계로 수행한다: (1) 첨부 URL 생성 API 호출 → `fileId`와 `uploadUrl` 수신, (2) `uploadUrl`에 파일 바이너리 PUT 업로드 (최대 파일 크기: API 제한에 따름, 초과 시 에러), (3) `fileId`를 포함한 파일 메시지 전송 API 호출. 각 단계 실패 시 해당 단계의 에러를 stderr에 출력한다. | P1 |
| FR-14 | `naverworks bot get-channel <channelId>` — 특정 채널의 상세 정보를 조회한다 (채널 목록 조회 API는 존재하지 않으므로 단건 조회만 지원). | P1 |
| FR-15 | `naverworks bot channel-members <channelId>` — 채널 멤버 목록을 조회한다. | P1 |
| FR-16 | `--text -` 플래그 시 stdin에서 메시지 본문을 읽는다. (예: `echo "hello" | naverworks bot send --to user1 --text -`) | P1 |

### 2.4 Directory (`naverworks directory`)

| ID | 요구사항 | 우선순위 |
|----|---------|---------|
| FR-17 | `naverworks directory list-users` — 사용자 목록을 조회한다. | P0 |
| FR-18 | `naverworks directory get-user <userId>` — 사용자 상세 정보를 조회한다. | P0 |
| FR-19 | `naverworks directory list-groups` — 그룹 목록을 조회한다. | P1 |
| FR-20 | `naverworks directory get-group <groupId>` — 그룹 상세 정보를 조회한다. | P1 |
| FR-21 | `--count`, `--cursor` 플래그로 페이지네이션을 제어한다. `count` 유효 범위는 엔드포인트별로 다르며(예: Directory 1~100, Calendar personals 1~50), 검증은 API 응답에 위임한다. 기본값은 API 기본값을 따른다. | P0 |

### 2.5 Calendar (`naverworks calendar`)

Calendar API 호출에는 `userId`가 필수이다. `--user-id` 플래그 또는 config의 `default_calendar_user_id`로 지정한다. OAuth 모드에서는 `--user-id me`를 허용한다. JWT 모드에서는 `me`를 사용할 수 없으므로 명시적 userId가 필수이며, 미지정 시 에러를 반환한다.

| ID | 요구사항 | 우선순위 |
|----|---------|---------|
| FR-22 | `naverworks calendar list-calendars --user-id <userId>` — 사용자의 캘린더 개인 속성 목록을 조회한다 (API: `GET /users/{userId}/calendar-personals`). 각 항목의 `calendarId`를 알아내기 위한 진입점이다. 기본 캘린더 정보는 `GET /users/{userId}/calendar`로 별도 조회 가능하며, `--default` 플래그 시 기본 캘린더만 반환한다. | P0 |
| FR-23 | `naverworks calendar list-events --calendar-id <id> --user-id <userId> --from <RFC3339> --until <RFC3339>` — 일정 목록을 조회한다. `--from`과 `--until`은 필수이며, 간격은 최대 31일이다. 초과 시 에러를 반환한다. | P0 |
| FR-24 | `naverworks calendar get-event --calendar-id <id> --event-id <id> --user-id <userId>` — 일정 상세 정보를 조회한다. | P0 |
| FR-25 | `naverworks calendar create-event --calendar-id <id> --user-id <userId> --title <title> --start <RFC3339> --end <RFC3339>` — 일정을 생성한다. 성공 시 생성된 이벤트를 JSON으로 stdout에 출력한다. 추가 옵션: `--description`, `--location`, `--is-all-day`. | P1 |

### 2.6 공통

| ID | 요구사항 | 우선순위 |
|----|---------|---------|
| FR-26 | `naverworks version` — 버전 정보를 출력한다. (빌드 시 `ldflags`로 주입: version, commit, build date) | P0 |
| FR-27 | `naverworks --help`, 각 서브커맨드 `--help` — cobra 기본 도움말을 출력한다. | P0 |

### 2.7 페이지네이션 공통 정책

cursor 기반 페이지네이션을 지원하는 목록 명령(`directory list-users`, `directory list-groups`, `calendar list-calendars`, `bot channel-members`)에 `--count`와 `--cursor` 플래그를 제공한다. `calendar list-events`는 cursor가 아닌 시간 범위(`--from`/`--until`) 기반이므로 페이지네이션 대상에서 제외하며, 지정된 시간 범위 내 전체 결과를 반환한다. 기본 동작은 **첫 페이지만 반환**이며, 응답에 `nextCursor`가 있으면 JSON 출력에 포함하여 사용자가 다음 페이지를 수동으로 요청할 수 있도록 한다. `--all` 플래그 시 자동으로 모든 페이지를 순회하여 결과를 합쳐 반환한다.

---

## 3. 비기능 요구사항

| ID | 요구사항 | 기준 |
|----|---------|------|
| NFR-01 | 시작 시간 | < 50ms (config 로드 포함, `naverworks version` 기준 측정) |
| NFR-02 | 바이너리 크기 | < 15MB |
| NFR-03 | 출력 형식 | JSON 기본. `--output table` 옵션 시 테이블 형식. 테이블 컬럼은 각 명령별로 정의한다 (예: `directory list-users` → `userId, userName, email, department`). |
| NFR-04 | 에러 출력 | stderr로 에러 메시지 출력, 정상 종료 시 exit 0, 비정상 시 exit 1. API 에러는 `{"error": {"code": "...", "description": "..."}}` 형식으로 stderr에 출력. |
| NFR-05 | 크로스 플랫폼 | Linux amd64/arm64, macOS amd64/arm64, Windows amd64 |
| NFR-06 | Rate Limit | API 응답 헤더 `RateLimit-Limit`, `RateLimit-Remaining`, `RateLimit-Reset` (또는 레거시 `X-RateLimit-*`)를 파싱한다. `RateLimit-Reset`은 "기준 시간 갱신까지 남은 시간(초)"으로 해석한다. HTTP 429 시 `RateLimit-Reset` 값(초)만큼 대기 후 재시도, 최대 3회. 헤더가 없으면 지수 백오프(1초, 2초, 4초)를 적용한다. 429가 동시 접속 제한(concurrent limit)인 경우 요청을 직렬화하여 순차 재시도한다. 3회 초과 시 에러 반환. |
| NFR-07 | 보안 | config.json/token.json 파일 권한: Linux/macOS 0600, Windows는 현재 사용자 전용 ACL. 토큰은 `token.json`에 별도 저장. 로그/에러 출력에서 access token, refresh token, client_secret은 절대 노출하지 않는다. |

---

## 4. 인증 Scope 정의

| 인증 방식 | 필요 scope | 비고 |
|-----------|-----------|------|
| OAuth 2.0 | `openid`, `profile`, `bot`, `directory`, `calendar` | `openid profile`은 OIDC userinfo용. 사용자 동의 화면에 표시 |
| JWT | `bot`, `directory`, `calendar` | JWT assertion의 scope 클레임에 포함. JWT에서는 OIDC scope 불필요 |

사용자가 `config set scope`로 커스텀 scope를 설정할 수 있다. 미설정 시 위 기본값을 사용한다.

---

## 5. 토큰 저장 스키마

`~/.config/naverworks/token.json` (또는 Windows `%APPDATA%\naverworks\token.json`):

```json
{
  "auth_method": "oauth",
  "access_token": "...",
  "refresh_token": "...",
  "token_type": "Bearer",
  "expires_at": "2026-03-20T12:00:00Z",
  "scope": "openid profile bot directory calendar"
}
```

---

## 6. 프로젝트 구조

```
naverworks-cli/
├── cmd/
│   ├── root.go          # 루트 커맨드, 글로벌 플래그 정의
│   ├── auth.go          # naverworks auth *
│   ├── config_cmd.go    # naverworks config * (config 패키지와 이름 충돌 방지)
│   ├── bot.go           # naverworks bot *
│   ├── directory.go     # naverworks directory *
│   ├── calendar.go      # naverworks calendar *
│   └── version.go       # naverworks version
├── internal/
│   ├── api/             # 네이버웍스 REST API 클라이언트
│   │   ├── client.go    # HTTP client (인증 헤더, 401 재시도, 429 백오프)
│   │   ├── bot.go       # Bot API 호출
│   │   ├── directory.go # Directory API 호출
│   │   └── calendar.go  # Calendar API 호출
│   ├── auth/            # JWT, OAuth 토큰 관리
│   │   ├── jwt.go       # JWT assertion 생성 및 토큰 발급
│   │   ├── oauth.go     # OAuth 플로우 (로컬 서버, 코드 교환)
│   │   └── token.go     # TokenStore (읽기/쓰기/갱신)
│   ├── config/          # 설정 파일 읽기/쓰기, 환경변수 오버라이드
│   └── output/          # JSON/테이블 출력 포맷터
├── go.mod
├── go.sum
├── main.go
├── Makefile
└── docs/
    └── requirements/
```

### 공통 초기화 (per-command 방식)

각 서비스 커맨드(`bot`, `directory`, `calendar`)의 `RunE`에서 다음을 순서대로 수행한다:
1. `loadConfigAndToken()` 호출 — config 파일 로드 + 환경변수 오버라이드 병합 + TokenStore에서 토큰 로드
2. `buildAPIClient()` 호출 — API client 생성 (인증 헤더, 401 재시도 1회, 429 백오프 미들웨어 포함)
3. 출력 포맷터 초기화 (`--output` 글로벌 플래그 확인)

`auth`, `config`, `version` 서브커맨드는 위 초기화를 수행하지 않는다.

---

## 7. 제약조건

| 항목 | 내용 |
|------|------|
| 네이버웍스 API 버전 | v1.0 |
| API Base URL | `https://www.worksapis.com/v1.0` |
| Auth URL | `https://auth.worksmobile.com/oauth2/v2.0` |
| Go 버전 | 1.22+ |
| Bot API | `botId` 필수 — 네이버웍스 Developer Console에서 생성 |
| Calendar API | `userId` 필수 — OAuth 시 `me` 허용, JWT 시 명시적 ID 필요 |
| 페이지네이션 | cursor 기반. count 최댓값은 엔드포인트별로 다름 (Directory: 1~100, Calendar personals: 1~50 등). CLI는 API 응답의 에러로 검증을 위임하고 클라이언트 측 검증은 하지 않는다. |
| 시간 형식 | RFC3339 (`2006-01-02T15:04:05Z07:00`) |
