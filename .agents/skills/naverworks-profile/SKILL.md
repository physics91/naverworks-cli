---
name: naverworks-profile
description: Use when setting up or troubleshooting naverworks CLI multi-profile authentication, including profile creation, OAuth/JWT login, and CI/CD auth configuration. Triggers on "naverworks 프로필", "naverworks 인증", "NW_PROFILE". If the task is executing NAVER WORKS API commands, use the naverworks-cli skill. If the task is build or release automation, use the build or deploy skill.
---

# 네이버웍스 CLI 멀티 프로필

인증이랑 프로필 전용 스킬임. 여러 네이버웍스 환경(개발/스테이징/운영)을 프로필로 분리하고 OAuth/JWT 로그인 문제를 다룸

> **부작용 있는 명령어** (`auth setup`, `config set`, `auth login`, `auth refresh`, `auth logout`)는 사용자 확인 후 실행한다.

## 이 스킬을 쓸 때

- 새 프로필 만들기, 기존 프로필 재설정
- `auth setup`, `auth login`, `auth status`, `auth refresh` 흐름 점검
- `NW_PROFILE` 기반 CI/CD 인증 구성
- OAuth/JWT 인증 오류 트러블슈팅

## 이 스킬을 쓰지 말아야 할 때

- 메일/드라이브/디렉토리 같은 API 명령 실행은 `naverworks-cli`
- 이 저장소의 빌드/테스트/배포는 `build`, `test`, `deploy`, `version`

## 입출력 계약

### 입력

- 공통: `profile`(생략 시 활성 프로필), 문제 증상 또는 목표
- 프로필 생성/로그인: `auth_mode` (`oauth`/`jwt`), `client_id`, `client_secret`
- JWT 설정: `service_account_id`, `private_key_path` 추가
- 선택 입력: `bot_id`, `domain_id`, `scope`
- 상태 점검/갱신/로그아웃: 기존 프로필이나 환경변수만 있어도 진행 가능

### 성공 기준

- `naverworks --profile <name> auth status` 종료 코드 0
- `auth_method`가 의도한 인증 방식(`oauth` 또는 `jwt`)과 일치
- `expires_at`이 현재 시각 이후

### 출력 형식

다음 5개를 사용자에게 보고한다

- `profile`
- `auth_mode`
- `config_source` (`interactive` / `manual` / `env`)
- `verification` (`auth_method`, `expires_at`, 필요 시 `scopes`)
- `next_action` (없으면 성공)

### 실패 시 대응

아래 트러블슈팅 참조. 해결 불가 시 `auth setup`으로 재설정.

## 기본 절차

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

## 참고

- 주요 설정 키는 `client_id`, `client_secret`, `service_account_id`, `private_key_path`, `domain_id`, `bot_id`, `scope`, `default_calendar_user_id`, `scim_access_token`
- 환경변수는 `NW_CLIENT_ID`, `NW_CLIENT_SECRET`, `NW_SERVICE_ACCOUNT_ID`, `NW_PRIVATE_KEY_PATH`, `NW_DOMAIN_ID`, `NW_BOT_ID`, `NW_SCOPE`, `NW_DEFAULT_CALENDAR_USER_ID`, `NW_SCIM_ACCESS_TOKEN`
- 설정 파일은 Linux/macOS에서 `~/.config/naverworks/config.json`, Windows에서 `%APPDATA%\\naverworks\\config.json`
- 토큰 파일은 Linux/macOS에서 `~/.config/naverworks/token.json`, Windows에서 `%APPDATA%\\naverworks\\token.json`
- 레거시 단일 설정은 자동으로 `default` 프로필로 마이그레이션됨
