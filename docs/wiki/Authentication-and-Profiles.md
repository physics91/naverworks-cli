# Authentication and Profiles

## 인증 방식

`naverworks`는 두 가지 인증 흐름을 지원합니다.

- OAuth 2.0: 일반 사용자 기준으로 가장 무난한 방식
- JWT Service Account: 서비스 계정 기반 자동화용

가장 쉬운 시작은 아래 순서입니다.

```bash
naverworks auth setup
naverworks auth login
naverworks auth status
```

## 수동 설정

직접 설정할 때는 `config set`을 사용합니다.

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

로그인:

```bash
# OAuth 2.0
naverworks auth login

# JWT Service Account
naverworks auth login --jwt
```

토큰 갱신/상태/로그아웃:

```bash
naverworks auth refresh
naverworks auth status
naverworks auth logout
```

## 프로필 우선순위

활성 프로필은 아래 순서로 결정됩니다.

1. `--profile`
2. `NW_PROFILE`
3. 설정 파일의 `current_profile`
4. `default`

## 프로필 예시

```bash
# staging 프로필에 설정 저장
naverworks --profile staging config set client_id STAGING_CLIENT_ID
naverworks --profile staging config set client_secret --stdin <<< "STAGING_SECRET"
naverworks --profile staging auth login

# staging 프로필로 실행
naverworks --profile staging directory list-users

# 환경변수로 기본 프로필 지정
export NW_PROFILE=staging
naverworks bot send --to USER_ID --text "hello"
```

설정 키와 환경변수 표는 [Configuration Keys and Environment Variables](Configuration-Keys-and-Environment-Variables.md)에 정리되어 있습니다.
