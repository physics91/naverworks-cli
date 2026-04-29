# naverworks User Guide

`naverworks`는 NAVER WORKS REST API v1.0을 터미널에서 다루기 위한 CLI입니다. 이 페이지에서 설치, 인증, 명령 실행, 출력, 문제 해결 문서로 이동할 수 있습니다.

- 프로필 기반 인증 지원
- OAuth 2.0 / JWT Service Account 로그인 지원
- JSON/테이블 출력 지원
- `--count`, `--cursor`, `--all` 페이지네이션 지원
- 주요 협업/조직/파일/메일/SCIM 도메인 커맨드 제공

## 무엇을 하려는가요?

- 처음 설치하려면 [Installation](Installation.md)을 봅니다.
- 바로 실행해보려면 [Quick Start](Quick-Start.md)를 봅니다.
- 인증을 설정하려면 [Authentication and Profiles](Authentication-and-Profiles.md)를 봅니다.
- 환경변수와 설정 키를 확인하려면 [Configuration Keys and Environment Variables](Configuration-Keys-and-Environment-Variables.md)를 봅니다.
- 출력 형식과 페이지네이션을 이해하려면 [Output and Pagination](Output-and-Pagination.md)를 봅니다.
- 도메인별 명령을 찾으려면 [Domain Command Guide](Domain-Command-Guide.md)를 봅니다.
- SCIM 명령을 사용하려면 [SCIM](SCIM.md)을 봅니다.
- 문제가 발생했다면 [Troubleshooting](Troubleshooting.md)을 봅니다.

## 첫 실행 순서

1. [Installation](Installation.md)에서 설치 방법을 선택합니다.
2. [Quick Start](Quick-Start.md)에서 기본 실행 흐름을 확인합니다.
3. [Authentication and Profiles](Authentication-and-Profiles.md)에서 인증과 프로필을 설정합니다.
4. 아래 최소 실행 예시를 실행합니다.

## 최소 실행 예시

```bash
naverworks auth setup
naverworks auth login
naverworks auth status
naverworks directory list-users --count 20
```

세부 설정 키, 환경변수, SCIM, 페이지네이션은 각 전용 페이지로 분리되어 있습니다.

## 위키 유지보수

- 새 문서를 쓰거나 큰 구조를 바꾸기 전에 [Wiki Writing Guide](Writing-Guide.md)를 확인합니다.
- 퍼블리시 전에는 [Wiki Review Checklist](Wiki-Review-Checklist.md)로 변경된 문서를 점검합니다.
