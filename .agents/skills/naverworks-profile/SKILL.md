---
name: naverworks-profile
description: Use when setting up or troubleshooting naverworks CLI multi-profile authentication. Covers profile creation, OAuth/JWT login, and auth configuration for CI/CD environments. Triggers on "naverworks 프로필", "naverworks 인증", "NW_PROFILE". If the task is build or release automation, use the build or deploy skill. If only running API commands, refer to naverworks --help instead.
---

# 네이버웍스 CLI 멀티 프로필

여러 네이버웍스 환경(개발/스테이징/운영)을 프로필로 분리하여 관리한다.

> **부작용 있는 명령어** (`auth setup`, `config set`, `auth login`, `auth refresh`, `auth logout`)는 사용자 확인 후 실행한다.

## 입출력 계약

### 입력 (시나리오별)

**프로필 생성/재설정 시 (OAuth):**

| 입력 | 필수 여부 | 설명 |
|------|----------|------|
| `profile` | 필수 | 생성할 프로필명 (예: `dev`, `prod`) |
| `client_id` | 필수 | 네이버웍스 Developer Console에서 발급 |
| `client_secret` | 필수 | 네이버웍스 Developer Console에서 발급 |
| `bot_id` | 선택 | 메시지 전송 시 필요 |
| `domain_id` | 선택 | 도메인 ID |
| `scope` | 선택 | 기본값: `openid profile bot directory calendar` |

**프로필 생성/재설정 시 (JWT) — OAuth 항목에 추가:**

| 입력 | 필수 여부 | 설명 |
|------|----------|------|
| `service_account_id` | 필수 | JWT Service Account ID |
| `private_key_path` | 필수 | Private Key 파일 경로 (권한 600) |
| `scope` | 선택 | 기본값: `bot directory calendar` |

**트러블슈팅/검증 시:**

| 입력 | 필수 여부 | 설명 |
|------|----------|------|
| `profile` | 선택 | 대상 프로필. 생략 시 활성 프로필 사용 |
| 증상 | 필수 | 에러 메시지 또는 문제 설명 |

### 성공 기준

`naverworks --profile <name> auth status` 실행 시:
- 종료 코드 0 반환
- `auth_method`가 의도한 인증 방식(`oauth` 또는 `jwt`)과 일치
- `expires_at`이 현재 시각 이후

### 출력 형식

스킬 완료 시 다음을 사용자에게 보고한다:

| 필드 | 예시 |
|------|------|
| `profile` | `prod` |
| `auth_mode` | `jwt` |
| `config_source` | `interactive` 또는 `manual` 또는 `env` |
| `verification` | `auth_method: jwt, expires_at: 2026-03-25T12:00:00Z` |
| `next_action` | 없음 (성공) 또는 트러블슈팅 항목 |

### 실패 시 대응

아래 트러블슈팅 참조. 해결 불가 시 `auth setup`으로 재설정.

## 프로필 설정 절차

### OAuth 설정

1. 사용자에게 프로필명과 인증 방식 확인
2. `naverworks --profile <name> auth setup` 실행 (대화형)
   - 또는 수동:
     ```bash
     naverworks --profile <name> config set client_id YOUR_ID
     naverworks --profile <name> config set client_secret --stdin <<< "SECRET"
     ```
3. `auth setup` 마지막 단계에서 즉시 로그인하지 않았다면 `naverworks --profile <name> auth login` 실행
4. 브라우저에서 네이버웍스 로그인 완료
5. `naverworks --profile <name> auth status`로 검증
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

### CI/CD (환경변수 방식)

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
→ `auth status` 출력에는 `authenticated` 필드가 없고, 성공 시 `auth_method`, `expires_at`, `scopes`가 출력된다.

### 트러블슈팅

| 증상 | 원인 | 해결 |
|------|------|------|
| "프로필 'X'을(를) 찾을 수 없습니다" | config.json에 프로필 없음 | `naverworks --profile X auth setup` |
| 토큰 만료 | access_token 유효기간 초과 | `auth refresh` 또는 재로그인 |
| JWT 로그인 실패 | private key 경로/권한 오류 | `private_key_path` 확인, 파일 권한 600 |
| 환경변수가 무시됨 | `--profile` 플래그가 우선 | 플래그 제거 또는 값 변경 |

## references/

<details>
<summary>설정 키 & 환경변수 전체 목록</summary>

| 키 | 환경변수 |
|----|---------|
| `client_id` | `NW_CLIENT_ID` |
| `client_secret` | `NW_CLIENT_SECRET` |
| `service_account_id` | `NW_SERVICE_ACCOUNT_ID` |
| `private_key_path` | `NW_PRIVATE_KEY_PATH` |
| `domain_id` | `NW_DOMAIN_ID` |
| `bot_id` | `NW_BOT_ID` |
| `scope` | `NW_SCOPE` |
| `default_calendar_user_id` | `NW_DEFAULT_CALENDAR_USER_ID` |
| `scim_access_token` | `NW_SCIM_ACCESS_TOKEN` |

환경변수는 config.json보다 우선한다.

</details>

<details>
<summary>파일 위치</summary>

| 파일 | Linux/macOS | Windows |
|------|------------|---------|
| 설정 | `~/.config/naverworks/config.json` | `%APPDATA%\naverworks\config.json` |
| 토큰 | `~/.config/naverworks/token.json` | `%APPDATA%\naverworks\token.json` |

레거시 형식(profiles 키 없는 단일 설정)은 자동으로 `"default"` 프로필로 마이그레이션된다.
토큰은 `token.json` 내 `tokens.<profile>` 구조로 프로필별 저장된다.

</details>
