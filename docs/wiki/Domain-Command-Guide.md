# Domain Command Guide

전체 하위 명령은 아래처럼 확인하는 편이 제일 빠릅니다.

```bash
naverworks --help
naverworks <command> --help
```

이 페이지는 자주 쓰는 도메인 예시만 모아둔 사용자 가이드입니다.

## Bot

```bash
naverworks bot send --to USER_ID --text "배포 완료"
naverworks bot list
```

## Calendar

```bash
naverworks calendar list-calendars --user-id me
naverworks calendar list-events \
  --user-id me \
  --calendar-id CALENDAR_ID \
  --from 2026-03-01T00:00:00Z \
  --until 2026-03-31T23:59:59Z
```

## Directory

```bash
naverworks directory list-users --count 20
naverworks directory search-users "홍길동" --domain-id DOMAIN_ID
naverworks directory search-groups "개발" --order-by "groupName asc"
naverworks directory search-orgunits "플랫폼"
naverworks directory list-user-groups USER_ID --membership-type DIRECT
naverworks directory get-user USER_ID
naverworks directory list-groups
```

Directory 검색·소속 조회는 구성원 또는 서비스 계정 Access Token을 지원합니다. 리소스별 `user.read`, `group.read`, `orgunit.read` 또는 상위 `directory.read` scope가 필요합니다.

## Drive

```bash
naverworks drive list --user-id me
naverworks drive search "분기 보고서" --user-id me --query-filters "fileName,content"
naverworks drive upload --user-id me ./report.pdf
naverworks drive shared-folder list
naverworks drive shared-folder list-files SHARED_FOLDER_ID
naverworks drive channel list
naverworks drive channel files CHANNEL_FOLDER_ID --all --output table
naverworks drive channel upload CHANNEL_FOLDER_ID --folder FOLDER_ID --file ./report.pdf
naverworks --dry-run drive channel delete CHANNEL_FOLDER_ID FILE_ID
```

Drive 검색과 채널 폴더는 구성원 계정 Access Token 전용입니다. 채널 폴더 조회에는 `file.read`, `group.folder.read`, 쓰기에는 `file`, `group.folder` scope가 필요합니다. 실제 쓰기 전에는 `--dry-run` 또는 `--plan-out`으로 요청을 확인할 수 있고 presigned 업로드 URL은 출력에서 마스킹됩니다.

## Mail

```bash
naverworks mail list-folders --user-id me
naverworks mail list FOLDER_ID --user-id me --count 20
naverworks mail get MAIL_ID --user-id me --has-threads
naverworks mail send --user-id me --to user@example.com --subject "배포 완료" --body "운영 반영 끝"
```

## Approval

```bash
naverworks approval list --user-id me
naverworks approval list-all --from 2026-03-01 --until 2026-03-31 --type approved
naverworks approval get DOCUMENT_ID --user-id me
```

`approval list-all`은 관리자 권한과 `businessSupport.approval.read` scope가 필요합니다.

## Task

```bash
naverworks task list --user-id me --all
naverworks task search "주간 점검" --user-id me --status TODO
naverworks task create --user-id me --title "주간 점검"
```

Task 검색은 구성원 계정 Access Token 전용이며 `task.read` scope가 필요합니다.

## Board

```bash
naverworks board list
naverworks board list-posts BOARD_ID
```

## Contact

```bash
naverworks contact list
naverworks contact get CONTACT_ID
```

## Attendance

```bash
naverworks attendance status --user-id me
naverworks attendance list-timecards --user-id me
```

## Audit

```bash
naverworks audit download-logs \
  --start-time 2026-03-01T00:00:00Z \
  --end-time 2026-03-31T23:59:59Z

naverworks audit download-logs \
  --service approval \
  --start-time 2026-03-01T00:00:00Z \
  --end-time 2026-03-31T23:59:59Z
```

## Monitoring

```bash
naverworks monitoring download-messages \
  --start-time 2026-03-01T00:00:00Z \
  --end-time 2026-03-31T23:59:59Z \
  --channel-id CHANNEL_ID
```

Monitoring 메시지 콘텐츠 다운로드는 관리자 또는 Service Account 권한과 `monitoring.read` scope가 필요합니다. CLI는 API redirect를 자동으로 따라가지 않고 다운로드 URL만 반환합니다.

SCIM은 토큰 체계가 따로라서 [SCIM](SCIM.md) 페이지를 별도로 보는 게 낫습니다.
