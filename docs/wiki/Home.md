# naverworks User Guide

`naverworks`는 NAVER WORKS REST API v1.0을 터미널에서 다루기 위한 CLI입니다.

- 프로필 기반 인증 지원
- OAuth 2.0 / JWT Service Account 로그인 지원
- JSON/테이블 출력 지원
- `--count`, `--cursor`, `--all` 페이지네이션 지원
- 주요 협업/조직/파일/메일/SCIM 도메인 커맨드 제공

## 가장 먼저 볼 문서

- [Installation](Installation.md)
- [Quick Start](Quick-Start.md)
- [Authentication and Profiles](Authentication-and-Profiles.md)
- [Configuration Keys and Environment Variables](Configuration-Keys-and-Environment-Variables.md)
- [Domain Command Guide](Domain-Command-Guide.md)
- [Troubleshooting](Troubleshooting.md)

## 첫 실행 순서

1. [Installation](Installation.md)에서 원하는 설치 방법 선택
2. [Quick Start](Quick-Start.md) 또는 [Authentication and Profiles](Authentication-and-Profiles.md) 기준으로 인증 설정
3. [Domain Command Guide](Domain-Command-Guide.md)에서 필요한 API 도메인 예시 확인

## 최소 실행 예시

```bash
naverworks auth setup
naverworks auth login
naverworks auth status
naverworks directory list-users --count 20
```

세부 설정 키, 환경변수, SCIM, 페이지네이션은 각 전용 페이지로 분리되어 있습니다.
