# Quick Start

가장 덜 귀찮은 시작 순서는 `auth setup` → 첫 API 호출입니다. OAuth 셋업은 Client ID/Secret만 묻고 바로 로그인까지 이어지므로, 그 다음에 `auth login`을 다시 할 필요는 없습니다.

## 1. 대화형 설정

```bash
naverworks auth setup
```

OAuth는 Client ID/Secret만 입력하면 바로 로그인까지 이어집니다. Redirect URL을 하나만 쓰려면 `--callback-port 8484`를 붙이세요.

## 2. 로그인 (setup에서 건너뛴 경우)

JWT Service Account, 또는 OAuth 셋업에서 로그인을 건너뛴 경우에만 따로 로그인합니다.

```bash
naverworks auth login --jwt
naverworks auth login
naverworks auth login --callback-port 8484
```

JWT 개인키는 Linux/macOS에서 `0600`, Windows에서 현재 사용자 전용 ACL이어야 합니다.

현재 상태 확인:

```bash
naverworks auth status
```

## 3. 첫 호출

조직 사용자 목록:

```bash
naverworks directory list-users --count 20
```

Bot 메시지 전송:

```bash
naverworks bot send --to USER_ID --text "배포 완료"
```

일정 조회:

```bash
naverworks calendar list-events \
  --user-id me \
  --calendar-id CALENDAR_ID \
  --from 2026-03-01T00:00:00Z \
  --until 2026-03-31T23:59:59Z
```

드라이브 파일 목록:

```bash
naverworks drive list --user-id me
```

메일 전송:

```bash
naverworks mail send \
  --user-id me \
  --to user@example.com \
  --subject "배포 완료" \
  --body "운영 반영 끝"
```

프로필 분리, 수동 설정, 환경변수 우선순위는 [Authentication and Profiles](Authentication-and-Profiles.md)와 [Configuration Keys and Environment Variables](Configuration-Keys-and-Environment-Variables.md)를 보면 됩니다.
