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
naverworks directory get-user USER_ID
naverworks directory list-groups
```

## Drive

```bash
naverworks drive list --user-id me
naverworks drive upload --user-id me ./report.pdf
naverworks drive shared-folder list
naverworks drive shared-folder list-files SHARED_FOLDER_ID
```

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
naverworks approval get DOCUMENT_ID --user-id me
```

## Task

```bash
naverworks task list --user-id me --all
naverworks task create --user-id me --title "주간 점검"
```

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
  --end-time 2026-03-31T23:59:59Z
```

SCIM은 토큰 체계가 따로라서 [SCIM](SCIM.md) 페이지를 별도로 보는 게 낫습니다.
