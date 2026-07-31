# Configuration Keys and Environment Variables

## 기본 명령

```bash
naverworks config set <key> <value>
naverworks config get <key>
naverworks config list
```

민감한 값은 `--stdin`으로 넣습니다. 값을 인자로 주면 셸 히스토리와 프로세스
목록에 남습니다.

```bash
# 시크릿 매니저나 파일에서 파이프로 전달 (명령문에 값이 남지 않음)
pass show naverworks/client-secret | naverworks config set client_secret --stdin
cat ~/secure/scim-token | naverworks config set scim_access_token --stdin
```

값을 직접 입력하려면 인자 없이 실행해 터미널에서 붙여넣으세요. here-string
(`<<< "값"`)은 명령문의 일부이므로 히스토리에 그대로 기록됩니다.

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

## 자격 증명 보관

NAVER WORKS 토큰 발급 API는 `client_secret`을 필수로 요구합니다. 그래서 CLI는
어떤 형태로든 이 값을 확보해야 하며, 보관 위치를 고를 수 있습니다.

**설정 파일에 저장 (기본)**

`config.json`에 평문으로 저장됩니다. CLI는 이 파일을 소유자만 접근할 수 있게
유지합니다.

- 저장 시 권한을 강제합니다 (POSIX `0600`, Windows는 현재 사용자 전용 ACL).
- 읽을 때도 권한을 확인하고, 느슨하면 조여준 뒤 경고를 출력합니다. 이전 버전이나
  백업에서 넘어온 파일이 그대로 노출되는 것을 막습니다.
- 권한이 느슨했다는 경고를 봤다면 자격 증명이 이미 노출됐을 수 있으니 Developer
  Console에서 재발급을 검토하세요.

**환경변수로만 전달 (디스크에 남기지 않음)**

`client_secret`을 파일에 두고 싶지 않으면 설정하지 말고 환경변수만 쓰면 됩니다.
환경변수가 설정 파일보다 우선합니다.

```bash
export NW_CLIENT_SECRET="$(pass show naverworks/client-secret)"
naverworks directory list-users --count 20
```

CI나 외부 시크릿 매니저(Vault, AWS Secrets Manager, GitHub Actions secrets 등)를
쓰는 환경에서는 이 방식을 권장합니다. 값이 프로세스 환경에만 존재하므로 디스크에
평문이 남지 않습니다.

다만 환경변수도 완전히 안전한 것은 아닙니다. 자식 프로세스로 상속되고, 같은
사용자 권한이면 프로세스 환경을 조회할 수 있으며(Linux `/proc/<pid>/environ`),
크래시 덤프나 디버그 로그에 포함될 수 있습니다. 값을 셸 프로필에 하드코딩하지
말고, 필요한 순간에 시크릿 매니저에서 읽어 해당 명령에만 전달하는 방식이 가장
안전합니다.

```bash
# 이 명령의 환경에만 값이 존재
NW_CLIENT_SECRET="$(pass show naverworks/client-secret)" \
  naverworks directory list-users --count 20
```

**셸 히스토리 주의**

`config set`에 값을 인자로 넘기면 셸 히스토리와 프로세스 목록에 남습니다.
`--stdin`으로 파이프해서 넣으면 둘 다 피할 수 있습니다. here-string은
명령문의 일부라 히스토리에 남으므로 피하세요.

```bash
pass show naverworks/client-secret | naverworks config set client_secret --stdin
```

## 설정 파일 위치

- Linux/macOS
  - `~/.config/naverworks/config.json`
  - `~/.config/naverworks/token.json`
- Windows
  - `%APPDATA%\\naverworks\\config.json`
  - `%APPDATA%\\naverworks\\token.json`

프로필 동작 방식은 [Authentication and Profiles](Authentication-and-Profiles.md), SCIM 전용 사용법은 [SCIM](SCIM.md)을 보면 됩니다.
