# Troubleshooting

## `auth login`이 안 됨

먼저 현재 설정과 인증 상태부터 확인합니다.

```bash
naverworks config list
naverworks auth status
```

체크할 것:

- `client_id`, `client_secret`가 현재 프로필에 들어가 있는지
- JWT를 쓰는 경우 `service_account_id`, `private_key_path`가 맞는지
- 환경변수(`NW_CLIENT_ID`, `NW_CLIENT_SECRET`, `NW_PROFILE`)가 다른 값을 덮어쓰고 있지 않은지

## 엉뚱한 프로필로 호출됨

우선순위는 `--profile` → `NW_PROFILE` → `current_profile` → `default`입니다.

```bash
echo "$NW_PROFILE"
naverworks --profile staging auth status
naverworks config list
```

프로필별로 저장된 값을 다시 확인하려면 명령마다 `--profile`을 명시하는 편이 덜 헷갈립니다.

## `--user-id` 때문에 자꾸 실패함

일부 커맨드는 사용자 컨텍스트가 필요합니다.

- Calendar: `--user-id me` 또는 `default_calendar_user_id`
- Mail: `--user-id me`
- Drive: `--user-id me`
- Attendance: `--user-id me`

예시:

```bash
naverworks calendar list-calendars --user-id me
naverworks mail list-folders --user-id me
```

## SCIM이 일반 로그인으로 안 됨

SCIM은 일반 OAuth/JWT 토큰이 아니라 `scim_access_token`을 따로 써야 합니다.

```bash
naverworks config get scim_access_token
naverworks scim list-users --count 10
```

`scim_access_token`이 비어 있으면 먼저 설정해야 합니다.

## 필요한 커맨드를 못 찾겠음

도메인별 도움말을 직접 보는 게 가장 정확합니다.

```bash
naverworks --help
naverworks directory --help
naverworks drive shared-folder --help
naverworks scim --help
```

대표 예시는 [Domain Command Guide](Domain-Command-Guide.md)에 정리돼 있습니다.

## 수정 후 뭐부터 돌려야 할지 모르겠음

빠르게 회귀만 보고 싶으면:

```bash
make test-fast
```

전체 테스트를 다시 확인하려면:

```bash
make test-full
go vet ./...
make build
```
