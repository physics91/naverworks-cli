# naverworks

네이버웍스(NAVER WORKS) REST API v1.0용 CLI 도구

`naverworks`는 프로필 기반 인증(OAuth 2.0, JWT Service Account), 자동 토큰 갱신, JSON/테이블 출력, 페이지네이션 순회를 지원하는 NAVER WORKS CLI입니다.

## 핵심 기능

- 인증: `auth setup`, `auth login`, `auth refresh`, `auth status`, `auth logout`
- 설정/프로필: `config set|get|list`, `--profile`, `NW_PROFILE`
- 주요 도메인: `bot`, `calendar`, `directory`, `drive`, `mail`, `approval`, `task`, `board`, `contact`, `attendance`, `audit`, `monitoring`, `scim`
- 출력: pretty JSON 기본 출력, 일부 목록형 명령의 `--output table`
- 페이지네이션: `--count`, `--cursor`, `--all`
- API 요청 미리보기: `--dry-run`, `--plan-out`, `--generate-input`

## 설치

### npm

```bash
npm install -g naverworks
```

### bun

```bash
bun add -g naverworks
```

전역 명령은 현재 Node 호환 런처를 사용합니다.
Node 없이 Bun만 쓰는 환경이면 아래 `bunx --bun` 또는 설치 스크립트를 권장합니다.

### npx

```bash
npx naverworks version
```

### bunx

```bash
bunx --bun naverworks version
```

### 설치 스크립트

```bash
curl -sSL https://raw.githubusercontent.com/physics91/naverworks-cli/main/install.sh | sh
```

기본 설치 경로는 `/usr/local/bin`이며, `INSTALL_DIR`로 변경할 수 있습니다.

추가 설치 방법과 플랫폼별 바이너리 정보는 [Installation](https://github.com/physics91/naverworks-cli/wiki/Installation) 문서를 보면 됩니다.

## 30초 시작

```bash
# 대화형 설정
naverworks auth setup

# 로그인
naverworks auth login

# 상태 확인
naverworks auth status

# 첫 API 호출
naverworks directory list-users --count 20
```

바로 다른 도메인도 써먹을 수 있습니다.

```bash
naverworks bot send --to USER_ID --text "배포 완료"
naverworks drive list --user-id me
naverworks mail send --user-id me --to user@example.com --subject "배포 완료" --body "운영 반영 끝"

# API 요청 미리보기
naverworks --dry-run bot send --bot-id BOT_ID --to USER_ID --text "배포 완료"
naverworks --dry-run directory list-users --count 20
```

### Directory 검색·소속 조회

Directory 조회 명령은 구성원 계정 또는 서비스 계정으로 발급한 Access Token을 사용합니다. 최소 권한으로 구성원 검색·소속 조회는 `user.read` 또는 `directory.read`, 그룹 검색은 `group.read` 또는 `directory.read`, 조직 검색은 `orgunit.read` 또는 `directory.read` scope가 필요합니다.

```bash
naverworks directory search-users "홍길동" --domain-id 10000001 --order-by "userName asc"
naverworks directory search-groups "개발" --order-by "groupName asc"
naverworks directory search-orgunits "플랫폼" --order-by "orgUnitName asc"
naverworks directory list-user-groups USER_ID --membership-type DIRECT
naverworks directory list-user-orgunits USER_ID --membership-type ALL
```

### Board·Note·Calendar·Contact 검색

Board·Calendar·Contact 검색은 구성원 계정 또는 서비스 계정 Access Token을 사용할 수 있으며 최소 `board.read`, `calendar.read`, `contact.read` scope가 각각 필요합니다. Note 검색은 구성원 계정 Access Token 전용이며 최소 `group.note.read` scope가 필요합니다.

```bash
naverworks board search-posts "회의록" --board-ids "1,2" --has-attachment
naverworks note search-posts GROUP_ID "회의록"
naverworks calendar search-events "주간 회의" --user-id me --query-filters "summary,attendee"
naverworks contact search "홍길동" --user-id me --query-filters "contactName,emails" --order-by "name asc"
```

Calendar 검색은 query 없이 기간만 지정할 수 있지만 `--query-filters`를 쓰려면 query가 필요합니다. 각 검색 명령은 `--cursor`, `--count`, `--all` 페이지네이션을 지원합니다.

### Task 검색·관리자 Approval 조회

Task 검색은 구성원 계정 Access Token 전용이며 최소 `task.read` scope가 필요합니다. query를 생략하려면 요청자·담당자·기간 조건 중 하나를 지정해야 합니다.

```bash
naverworks task search "주간 회의" --user-id me --status TODO --order-by "createdTime desc"
naverworks task search --user-id me --assignee-id USER_ID --has-due-date=false
```

관리자용 Approval 문서 조회는 구성원 계정 또는 서비스 계정 Access Token과 관리자 권한, 최소 `businessSupport.approval.read` scope가 필요합니다. `--from`과 `--until`은 필수이며 최대 1개월 범위에서 `--type pending|upcoming|approved|completed`를 선택할 수 있습니다.

```bash
naverworks approval list-all --from 2026-08-01 --until 2026-08-31 --type approved
```

### Drive 검색·채널 폴더·Monitoring 채널 다운로드

Drive 검색은 구성원 계정 Access Token 전용이며 서비스 계정 토큰은 사용할 수 없습니다. 최소 `file.read` scope가 필요하고, 채널 폴더를 포함하면 `group.folder.read` scope도 필요합니다. `--query-filters fileName,content`로 파일명과 본문 검색 범위를 지정할 수 있습니다.

```bash
naverworks drive search "분기 보고서" --user-id me --query-filters "fileName,content" --drive-type-filters "MY_DRIVE,CHANNEL_FOLDER"
```

채널 폴더는 구성원 계정 Access Token 전용입니다. 조회·다운로드에는 최소 `file.read`, `group.folder.read` scope가, 업로드·폴더 생성·파일 변경·복원·권한 변경에는 `file`, `group.folder` scope가 필요합니다. 파일 목록(`files`), 버전 목록(`revision list`), 휴지통 목록(`trash-list`)은 `--cursor`, `--count`, `--all`과 JSON·table 출력을 지원합니다. 공식 API에 페이지네이션이 없는 채널 목록(`channel list`)과 권한 목록(`permission list`)은 JSON·table 출력만 지원합니다.

```bash
naverworks drive channel list
naverworks drive channel files CHANNEL_FOLDER_ID --all --output table
naverworks drive channel revision list CHANNEL_FOLDER_ID FILE_ID --count 20
naverworks drive channel permission list CHANNEL_FOLDER_ID FILE_ID
naverworks drive channel upload CHANNEL_FOLDER_ID --folder FOLDER_ID --file ./report.pdf

# 쓰기 요청을 보내지 않고 method·path·body 확인
naverworks --dry-run drive channel delete CHANNEL_FOLDER_ID FILE_ID
```

업로드 URL은 HTTPS·허용 호스트 검증을 거치며 stdout에 노출하지 않습니다. 다운로드 명령은 API redirect를 자동으로 따라가지 않고 URL만 반환합니다. 실제 쓰기 전에 `--dry-run` 또는 `--plan-out`으로 요청을 확인할 수 있습니다.

Monitoring 메시지 콘텐츠 다운로드는 관리자 또는 Service Account 권한과 `monitoring.read` scope가 필요합니다. 기존 기간 조회에 `--channel-id`를 지정하면 특정 메시지방만 대상으로 다운로드 URL을 요청합니다.

```bash
naverworks monitoring download-messages --start-time "2026-08-01T00:00:00+09:00" --end-time "2026-08-31T23:59:59+09:00" --channel-id CHANNEL_ID
```

## 문서

- [User Guide Home](https://github.com/physics91/naverworks-cli/wiki)
- [Installation](https://github.com/physics91/naverworks-cli/wiki/Installation)
- [Quick Start](https://github.com/physics91/naverworks-cli/wiki/Quick-Start)
- [Authentication and Profiles](https://github.com/physics91/naverworks-cli/wiki/Authentication-and-Profiles)
- [Configuration Keys and Environment Variables](https://github.com/physics91/naverworks-cli/wiki/Configuration-Keys-and-Environment-Variables)
- [Output and Pagination](https://github.com/physics91/naverworks-cli/wiki/Output-and-Pagination)
- [Domain Command Guide](https://github.com/physics91/naverworks-cli/wiki/Domain-Command-Guide)
- [SCIM](https://github.com/physics91/naverworks-cli/wiki/SCIM)
- [Troubleshooting](https://github.com/physics91/naverworks-cli/wiki/Troubleshooting)
- [Releases](https://github.com/physics91/naverworks-cli/releases)

상세 문서는 GitHub wiki에서 읽고, 원본은 `docs/wiki/`에서 관리합니다.

전체 명령은 아래처럼 확인할 수 있습니다.

```bash
naverworks --help
naverworks <command> --help
```

## API 호환성 노트

### 코어 정기 업데이트 (`core_20260723`)

- **CLI 반영 (스펙 확정분)**
  - `mail move <mailId> --folder <folderId>`: 메일 폴더 이동 (`PATCH .../mail/{mailId}`, body `folderId` 정수)
  - `task list`: `--category-id`, `--status`, `--search-filter-type` (`ALL|ASSIGNEE|ASSIGNOR`)
  - `approval list-all`: `--from`, `--until`, `--document-form-id`, `--order-by`
  - `directory search-users|search-groups|search-orgunits`: Directory 리소스 검색
  - `directory list-user-groups|list-user-orgunits`: 구성원 소속 그룹·조직 조회
  - 구성원 profile-status 응답의 `delegates` 필드 passthrough 검증
  - `board search-posts`, `note search-posts`, `calendar search-events`, `contact search`: 도메인 검색 API
  - `task search`, `approval list-all --type`: Task 검색과 관리자 결재 문서 필터
  - `drive search --query-filters`, `monitoring download-messages --channel-id`: Drive 검색과 채널별 메시지 콘텐츠 다운로드
  - `drive channel ...`: 채널 폴더 목록·파일·버전·링크·휴지통·권한 전체 명령군
- **응답 변경 (passthrough, 코드 변경 없음)**
  - 구성원/연락처 messenger type: 응답 값이 `X`로 통일될 수 있음 (기존 `TWITTER` 포함)
  - 공용 드라이브 목록 `quota.trash` 필드 제거 가능
  - 휴지통 목록에서 종료된 `orderBy` 값(`deletedDate`, `name`) 사용 불가 — CLI는 orderBy 미노출
- **미반영**
  - 없음 (2026-08-31 공식 Developers 문서 기준)

## 개발 검증

빠른 회귀 확인과 전체 검증을 분리해서 돌릴 수 있습니다.

```bash
make test-fast   # 핵심 unit/contract + meta + 대표 journey
make test-full   # 전체 테스트 스위트
make build
go vet ./...
```

## 자동 점검

GitHub Actions로 주기 점검을 돌립니다. 두 workflow 모두 `gh` CLI에 의존하지 않고 `${{ secrets.GITHUB_TOKEN }}` 기반 GitHub API 호출로 이슈를 생성합니다.

- `API Change Monitor` (`.github/workflows/api-monitor.yml`): NAVER WORKS 릴리즈 노트를 매일 확인하고 `docs/baselines/api-monitor-baseline.json`에 없는 새 공지를 `api-monitor` 이슈로 등록
- `Weekly Health` (`.github/workflows/weekly-health.yml`): `make test-full`, `go vet ./...`, `make build`, `go list -m -u -json all` 결과를 주간 점검하고 실패 시 `ops-error`, 업데이트 가능 모듈이 있으면 `health-check` 이슈로 등록

## 라이선스

MIT
