# Configuration Keys and Environment Variables

## 기본 명령

```bash
naverworks config set <key> <value>
naverworks config get <key>
naverworks config list
```

민감한 값은 `--stdin`으로 넣는 편이 안전합니다.

```bash
naverworks config set client_secret --stdin <<< "YOUR_CLIENT_SECRET"
naverworks config set scim_access_token --stdin <<< "YOUR_SCIM_TOKEN"
```

## 설정 키

| 키 | 설명 | 언제 필요한가 |
| --- | --- | --- |
| `client_id` | OAuth/JWT 공통 Client ID | 대부분의 일반 API 호출 |
| `client_secret` | Client Secret | OAuth/JWT 공통 |
| `service_account_id` | Service Account ID | JWT 로그인 |
| `private_key_path` | JWT 개인키 경로 | JWT 로그인 |
| `domain_id` | 도메인 ID | 도메인/봇/관리 계열 설정 시 |
| `bot_id` | 기본 Bot ID | Bot API 호출 시 |
| `scope` | OAuth/JWT scope | 기본 scope를 바꾸고 싶을 때 |
| `default_calendar_user_id` | Calendar 기본 `--user-id` | Calendar에서 `me`를 기본값으로 쓰고 싶을 때 |
| `scim_access_token` | SCIM 전용 액세스 토큰 | SCIM API 호출 시 |

## 환경변수

환경변수는 설정 파일보다 우선합니다.

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

## 설정 파일 위치

- Linux/macOS
  - `~/.config/naverworks/config.json`
  - `~/.config/naverworks/token.json`
- Windows
  - `%APPDATA%\\naverworks\\config.json`
  - `%APPDATA%\\naverworks\\token.json`

프로필 동작 방식은 [Authentication and Profiles](Authentication-and-Profiles.md), SCIM 전용 사용법은 [SCIM](SCIM.md)을 보면 됩니다.
