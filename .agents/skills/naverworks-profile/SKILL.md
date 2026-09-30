---
name: naverworks-profile
description: Use when setting up or troubleshooting naverworks CLI profiles and authentication, including OAuth/JWT API tokens, browser login/status/logout for message-room sessions, and NW_PROFILE for CI/CD. Use naverworks-cli for resource operations and project workflow skills for builds, tests, and releases.
---

# 네이버웍스 CLI 멀티 프로필

여러 네이버웍스 환경을 프로필로 분리하고 OAuth/JWT API 토큰과 메시지방 전용
브라우저 세션의 로그인·상태 확인·로그아웃을 다룬다.

> **부작용 있는 명령어** (`auth setup`, `config set`, `auth login`, `auth refresh`, `auth logout`)는 승인 범위에서 실행한다. 같은 프로필·작업·범위의 기존 승인은 재사용한다. `auth status`는 read이며 요청된 상태 확인에 추가 승인이 필요하지 않다.

## 이 스킬을 쓸 때

- 새 프로필 만들기, 기존 프로필 재설정
- `auth setup`, `auth login`, `auth status`, `auth refresh` 흐름 점검
- `NW_PROFILE` 기반 CI/CD 인증 구성
- OAuth/JWT 인증 및 메시지방 브라우저 세션 오류 트러블슈팅

## 이 스킬을 쓰지 말아야 할 때

- 메일/드라이브/디렉토리 API와 방 노트·할 일 명령 실행은 `naverworks-cli`
- 이 저장소의 빌드/테스트/배포는 `build`, `test`, `deploy`, `version`

## 입출력 계약

### 입력

- 공통: `profile`(생략 시 활성 프로필), 문제 증상 또는 목표
- API 프로필 생성/로그인: `auth_mode` (`oauth`/`jwt`), `client_id`, `client_secret`
- JWT 설정: `service_account_id`, `private_key_path` 추가
- 브라우저 로그인: `auth_mode: browser`, 사용할 프로필, 설치된 Chrome/Chromium
  (필요 시 `NW_BROWSER_PATH`). 비밀번호와 추가 인증은 사용자가 브라우저에서 입력한다.
- 선택 입력: `bot_id`, `domain_id`, `scope`
- 상태 점검/갱신/로그아웃: 기존 프로필이나 환경변수만 있어도 진행 가능

### 성공 기준

- API: `naverworks --profile <name> auth status` 종료 코드 0,
  `auth_method`가 `oauth` 또는 `jwt`로 일치하고 `expires_at`이 미래 시각이다.
- 브라우저: `naverworks --profile <name> auth status --method browser` 종료 코드 0,
  `auth_method: browser`, 선택한 `profile`, `authenticated: true`, `headless: true`다.
  브라우저 응답에 없는 `expires_at`이나 OAuth scope를 성공 조건으로 요구하지 않는다.

### 출력 형식

다음 5개를 사용자에게 보고한다

- `profile`
- `auth_mode`
- `config_source` (`interactive` / `manual` / `env` / `browser_session`)
- `verification` (API: `auth_method`, `expires_at`, 필요 시 `scopes`;
  브라우저: `auth_method`, `profile`, `authenticated`, `headless`)
- `next_action` (없으면 성공)

### 실패 시 대응

아래 트러블슈팅을 참조한다. API 설정 문제는 `auth setup`, 브라우저 세션 만료는
같은 프로필의 `auth login --method browser`로 처리하며 실행 전 승인 범위를 확인한다.

## 기본 절차

### OAuth 설정

1. 사용자에게 프로필명과 인증 방식 확인
2. `naverworks --profile <name> auth setup` 실행 (대화형)
   - OAuth는 Client ID/Secret만 묻고 바로 로그인한다. Scope·Calendar·사전 Bot 질문은 생략한다.
   - Redirect URL을 하나만 등록했다면 `--callback-port 8484`를 붙인다.
   - 또는 수동:
     ```bash
     naverworks --profile <name> config set client_id YOUR_ID
     naverworks --profile <name> config set client_secret --stdin <<< "SECRET"
     ```
3. `auth setup`에서 즉시 로그인하지 않았다면 `naverworks --profile <name> auth login` 실행. 이미 셋업이 OAuth 로그인을 끝냈으면 다시 로그인하지 않는다. OAuth의 인증 창은 `--method browser` 전용 웹 세션과 별개다.
4. `naverworks --profile <name> auth status`로 검증
   - `auth_method: oauth`와 미래 시각의 `expires_at` 확인 → 완료
   - 실패 → 트러블슈팅 참조

### JWT 설정

1. 사용자에게 프로필명 확인
2. `naverworks --profile <name> auth setup` 실행 (대화형)
   - 또는 수동:
     ```bash
     naverworks --profile <name> config set client_id YOUR_ID
     naverworks --profile <name> config set client_secret --stdin <<< "SECRET"
     naverworks --profile <name> config set service_account_id YOUR_SA_ID
     naverworks --profile <name> config set private_key_path /path/to/key.pem
     ```
3. `auth setup` 마지막 단계에서 즉시 로그인하지 않았다면 `naverworks --profile <name> auth login --jwt` 실행
4. `naverworks --profile <name> auth status`로 검증
   - `auth_method: jwt`와 미래 시각의 `expires_at` 확인 → 완료
   - 실패 → private_key_path 경로/권한(600) 확인

### 메시지방 브라우저 세션

1. 사용할 프로필과 Chrome/Chromium을 확인한다. 브라우저 실행 경로를 직접 지정하려면
   `NW_BROWSER_PATH`를 사용한다. 자동 설치하지 않으며 로그인·상태 확인·방 명령에
   브라우저가 필요하다. 브라우저 인증을 위해 OAuth/JWT 자격 증명을 새로 발급하지 않는다.
2. 승인된 프로필에서 `naverworks --profile <name> auth login --method browser`를 실행한다.
   로그인할 때만 전용 브라우저 창이 열린다. 비밀번호·추가 인증은 사용자가 직접 입력한다.
3. `naverworks --profile <name> auth status --method browser`로 실제 웹 로그인 상태를
   headless로 확인하고 위 브라우저 성공 기준을 적용한다. 방 노트·할 일 명령도 항상
   headless로 실행하며 같은 프로필의 브라우저 명령은 순서대로 실행한다.
4. 세션이 만료되면 같은 프로필로 재로그인한다. API `auth refresh`로 웹 세션을 갱신하지 않는다.
5. 로그아웃 요청과 프로필을 확인한 뒤 해당 프로필의 브라우저·CLI를 모두 종료하고
   `naverworks --profile <name> auth logout --method browser`를 실행한다.
   종료 코드 0과 `local_session_deleted: true`를 확인한다. 이 명령은 로컬 전용 세션만
   삭제하며 서버 세션을 철회하지 않는다. API 토큰과 다른 프로필은 보존하고,
   브라우저 실행 파일이 없어도 가능하다. 사용 중인 세션 잠금이 있으면 강제로 삭제하지 않는다.

`auth status`·`auth logout`은 기본적으로 API 토큰을 대상으로 하므로 웹 세션에는
항상 `--method browser`를 붙인다. 브라우저 인증 명령에는 미리보기 옵션을 사용할 수 없다.
API 토큰과 웹 세션은 별도로 저장되며 일반 브라우저의 프로필·쿠키를 가져오거나
외부 HTTP 클라이언트로 복사하지 않는다. 상세 동작은
[메시지방 노트·할 일](../../../docs/browser-room-management.md)을 참조한다.

### CI/CD (API 환경변수 방식)

1. 환경변수 설정 (프로필 지정 포함):
   ```bash
   export NW_PROFILE="<name>"
   export NW_CLIENT_ID="$CI_CLIENT_ID"
   export NW_CLIENT_SECRET="$CI_CLIENT_SECRET"
   export NW_SERVICE_ACCOUNT_ID="$CI_SA_ID"
   export NW_PRIVATE_KEY_PATH="/secrets/private.pem"
   export NW_BOT_ID="$CI_BOT_ID"
   export NW_SCOPE="bot directory calendar"
   ```
2. `naverworks auth login --jwt` (NW_PROFILE로 프로필 결정)
3. `naverworks auth status`로 검증
   - `auth_method: jwt`와 미래 시각의 `expires_at` 확인 → 완료
   - 실패 → 트러블슈팅 참조

## 검증 & 트러블슈팅

### 프로필 우선순위

1. `--profile` 플래그 (최우선)
2. `NW_PROFILE` 환경변수
3. `current_profile` (config.json 내 저장값)
4. `"default"` (기본값)

### 주요 명령어

```bash
naverworks --profile <name> auth status      # 인증 상태 확인
naverworks --profile <name> auth refresh     # 토큰 갱신
naverworks --profile <name> auth logout      # 로그아웃 (확인 필요)
naverworks --profile <name> config list      # 전체 설정 (민감값 마스킹)
naverworks --profile <name> config get <key> # 개별 조회
```
API `auth status` 출력에는 `authenticated` 필드가 없고 `auth_method`, `expires_at`,
`scopes`가 출력된다. 웹 세션은 `auth status --method browser`를 사용하며
`auth_method`, `profile`, `authenticated`, `headless`가 출력된다.

### 트러블슈팅

| 증상 | 원인 | 해결 |
|------|------|------|
| "프로필 'X'을(를) 찾을 수 없습니다" | config.json에 프로필 없음 | `naverworks --profile X auth setup` |
| 토큰 만료 | access_token 유효기간 초과 | `auth refresh` 또는 재로그인 |
| 브라우저 세션 없음·만료 | 선택한 프로필에 유효한 웹 세션이 없음 | 같은 프로필의 `auth login --method browser`로 재로그인 |
| Chrome/Chromium 없음·시작 실패 | 실행 파일 없음·지정 경로 오류·브라우저 시작 시간 초과 | 설치 상태와 `NW_BROWSER_PATH` 확인; 종료 코드 1과 `CLI_ERROR` JSON을 보고하고 자동 설치하지 않음 |
| 브라우저 세션 잠금 | 같은 프로필의 브라우저·CLI가 실행 중 | 해당 프로필의 실행을 끝내고 명령을 순서대로 실행; 강제 세션 삭제 금지 |
| JWT 로그인 실패 | private key 경로/권한 오류 | `private_key_path` 확인, 파일 권한 600 |
| OAuth `redirect_uri_mismatch` | 콘솔 Redirect URL과 콜백 포트 불일치 | `http://127.0.0.1:8484/callback`~`8494` 등록 또는 `--callback-port`로 고정 |
| 환경변수가 무시됨 | `--profile` 플래그가 우선 | 플래그 제거 또는 값 변경 |

## 참고

- 주요 설정 키는 `client_id`, `client_secret`, `service_account_id`, `private_key_path`, `domain_id`, `bot_id`, `scope`, `default_calendar_user_id`, `scim_access_token`
- 환경변수는 `NW_CLIENT_ID`, `NW_CLIENT_SECRET`, `NW_SERVICE_ACCOUNT_ID`, `NW_PRIVATE_KEY_PATH`, `NW_DOMAIN_ID`, `NW_BOT_ID`, `NW_SCOPE`, `NW_DEFAULT_CALENDAR_USER_ID`, `NW_SCIM_ACCESS_TOKEN`
- `NW_BROWSER_PATH`는 Chrome/Chromium 실행 파일 경로이며 웹 세션은 프로필별 전용 브라우저 디렉터리에 저장한다.
- 설정 파일은 Linux/macOS에서 `~/.config/naverworks/config.json`, Windows에서 `%APPDATA%\\naverworks\\config.json`
- 토큰 파일은 Linux/macOS에서 `~/.config/naverworks/token.json`, Windows에서 `%APPDATA%\\naverworks\\token.json`
- 레거시 단일 설정은 자동으로 `default` 프로필로 마이그레이션됨
