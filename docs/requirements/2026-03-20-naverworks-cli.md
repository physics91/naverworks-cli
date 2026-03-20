# 네이버웍스 CLI 요구사항 정의서

## 1. 개요

| 항목 | 내용 |
|------|------|
| 프로젝트명 | `naverworks-cli` (실행 명령: `nw-cli`) |
| 기술 스택 | Go + cobra |
| 초기 버전 | v0.1 |
| 대상 사용자 | 네이버웍스 관리자, 개발자, 자동화 스크립트 |
| 참조 SDK | `naverworks-sdk-kotlin` (API 스펙 참조용) |

---

## 2. 기능 요구사항

### 2.1 인증 (`nw-cli auth`)

| ID | 요구사항 | 우선순위 |
|----|---------|---------|
| FR-01 | `nw-cli auth login` — OAuth 2.0 인증. 로컬에 임시 HTTP 서버(포트 8484, 사용 중이면 8485~8494 순차 탐색)를 띄우고 브라우저를 열어 인증 코드를 수신한다. | P0 |
| FR-02 | `nw-cli auth login --jwt` — config에 설정된 client_id, client_secret, service_account_id, private_key로 JWT Service Account 인증을 수행한다. | P0 |
| FR-03 | `nw-cli auth status` — 현재 인증 상태(인증 방식, 토큰 만료 시각, 사용자 정보)를 JSON으로 stdout에 출력한다. | P0 |
| FR-04 | `nw-cli auth logout` — 저장된 토큰을 삭제하고 로그아웃한다. | P1 |
| FR-05 | access token 만료 300초 전 또는 API 401 응답 시, refresh token으로 자동 갱신한다. 갱신 실패 시 stderr에 에러를 출력하고 종료코드 1을 반환한다. | P0 |

### 2.2 설정 (`nw-cli config`)

| ID | 요구사항 | 우선순위 |
|----|---------|---------|
| FR-06 | `nw-cli config set <key> <value>` — client_id, client_secret, service_account_id, private_key_path, domain_id 등을 설정한다. | P0 |
| FR-07 | `nw-cli config get <key>` — 설정값을 조회한다. 민감 정보(client_secret, private_key)는 마스킹하여 출력한다. | P0 |
| FR-08 | `nw-cli config list` — 전체 설정 목록을 JSON으로 출력한다. 민감 정보는 마스킹한다. | P1 |
| FR-09 | 설정 파일 경로: `~/.config/naverworks/config.json` (파일 권한 0600) | P0 |
| FR-10 | 환경변수 오버라이드: `NW_CLIENT_ID`, `NW_CLIENT_SECRET`, `NW_SERVICE_ACCOUNT_ID`, `NW_PRIVATE_KEY_PATH`, `NW_DOMAIN_ID`. 환경변수가 설정 파일보다 우선한다. | P1 |

### 2.3 Bot (`nw-cli bot`)

| ID | 요구사항 | 우선순위 |
|----|---------|---------|
| FR-11 | `nw-cli bot send --to <userId> --text <message>` — 사용자에게 텍스트 메시지를 전송한다. 성공 시 메시지 ID를 JSON으로 stdout에 출력한다. | P0 |
| FR-12 | `nw-cli bot send --channel <channelId> --text <message>` — 채널에 텍스트 메시지를 전송한다. 성공 시 메시지 ID를 JSON으로 stdout에 출력한다. | P0 |
| FR-13 | `nw-cli bot send --to <userId> --file <path>` — 사용자에게 파일을 첨부하여 전송한다. | P1 |
| FR-14 | `nw-cli bot channels` — 봇의 채널 목록을 조회한다. | P1 |
| FR-15 | `--text -` 플래그 시 stdin에서 메시지 본문을 읽는다. (예: `echo "hello" | nw-cli bot send --to user1 --text -`) | P1 |

### 2.4 Directory (`nw-cli directory`)

| ID | 요구사항 | 우선순위 |
|----|---------|---------|
| FR-16 | `nw-cli directory list-users` — 사용자 목록을 조회한다. | P0 |
| FR-17 | `nw-cli directory get-user <userId>` — 사용자 상세 정보를 조회한다. | P0 |
| FR-18 | `nw-cli directory list-groups` — 그룹 목록을 조회한다. | P1 |
| FR-19 | `nw-cli directory get-group <groupId>` — 그룹 상세 정보를 조회한다. | P1 |
| FR-20 | `--count`, `--cursor` 플래그로 페이지네이션을 제어한다. 기본 count는 API 기본값을 따른다. | P0 |

### 2.5 Calendar (`nw-cli calendar`)

| ID | 요구사항 | 우선순위 |
|----|---------|---------|
| FR-21 | `nw-cli calendar list-events --calendar-id <id>` — 일정 목록을 조회한다. | P0 |
| FR-22 | `nw-cli calendar get-event --calendar-id <id> --event-id <id>` — 일정 상세 정보를 조회한다. | P0 |
| FR-23 | `nw-cli calendar create-event --calendar-id <id> --title <title> --start <time> --end <time>` — 일정을 생성한다. 성공 시 생성된 이벤트 ID를 JSON으로 stdout에 출력한다. | P1 |

### 2.6 공통

| ID | 요구사항 | 우선순위 |
|----|---------|---------|
| FR-24 | `nw-cli version` — 버전 정보를 출력한다. (빌드 시 ldflags로 주입) | P0 |
| FR-25 | `nw-cli --help`, 각 서브커맨드 `--help` — cobra 기본 도움말을 출력한다. | P0 |

---

## 3. 비기능 요구사항

| ID | 요구사항 | 기준 |
|----|---------|------|
| NFR-01 | 시작 시간 | < 10ms |
| NFR-02 | 바이너리 크기 | < 15MB |
| NFR-03 | 출력 형식 | JSON 기본 (`--output table` 옵션으로 테이블 전환) |
| NFR-04 | 에러 출력 | stderr로 에러 메시지 출력, 정상 종료 시 exit 0, 비정상 시 exit 1 |
| NFR-05 | 크로스 플랫폼 | Linux amd64/arm64, macOS amd64/arm64, Windows amd64 |
| NFR-06 | Rate Limit | API 응답 헤더 `RateLimit-*` 존중, HTTP 429 시 `Retry-After` 대기 후 최대 3회 재시도 |
| NFR-07 | 보안 | config.json 파일 권한 0600, 토큰은 `~/.config/naverworks/token.json`에 별도 저장 (권한 0600) |

---

## 4. 프로젝트 구조

```
naverworks-cli/
├── cmd/
│   ├── root.go          # nw-cli 루트 커맨드
│   ├── auth.go          # nw-cli auth *
│   ├── config.go        # nw-cli config *
│   ├── bot.go           # nw-cli bot *
│   ├── directory.go     # nw-cli directory *
│   ├── calendar.go      # nw-cli calendar *
│   └── version.go       # nw-cli version
├── internal/
│   ├── api/             # 네이버웍스 REST API 클라이언트
│   ├── auth/            # JWT, OAuth 토큰 관리
│   ├── config/          # 설정 파일 읽기/쓰기
│   └── output/          # JSON/테이블 출력 포맷터
├── go.mod
├── go.sum
├── main.go
├── Makefile
└── docs/
    └── requirements/
```

---

## 5. 제약조건

| 항목 | 내용 |
|------|------|
| 네이버웍스 API 버전 | v1.0 |
| API Base URL | `https://www.worksapis.com/v1.0` |
| Auth URL | `https://auth.worksmobile.com/oauth2/v2.0` |
| Go 버전 | 1.22+ |
