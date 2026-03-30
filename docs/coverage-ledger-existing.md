# Baseline Coverage Ledger (기존 구현)

> 자동 생성일: 2026-03-30
> 총 endpoint: **116개**

## 요약

| 도메인 | 수량 |
|--------|------|
| Bot | 4 |
| Calendar | 5 |
| Board | 8 |
| Drive MyDrive | 12 |
| Drive SharedFolders | 2 |
| Drive SharedDrive | 7 |
| Drive GroupFolder | 4 |
| Mail | 6 |
| Contact | 8 |
| Directory | 10 |
| Approval | 6 |
| Attendance | 5 |
| HR | 4 |
| Business Place | 5 |
| Task | 6 |
| Note | 7 |
| Form | 2 |
| Audit | 2 |
| Monitoring | 1 |
| SCIM Users | 6 |
| SCIM Groups | 6 |
| **합계** | **116** |

## 전수 목록

| # | 도메인 | HTTP | 경로 | API 파일 | CLI 커맨드 | Smoke Test |
|---|--------|------|------|----------|-----------|------------|
| 1 | Bot | POST | /bots/{botId}/users/{userId}/messages | internal/api/bot.go | bot send --to | ✓ |
| 2 | Bot | POST | /bots/{botId}/channels/{channelId}/messages | internal/api/bot.go | bot send --channel | ✓ |
| 3 | Bot | GET | /bots/{botId}/channels/{channelId} | internal/api/bot.go | bot get-channel | - |
| 4 | Bot | GET | /bots/{botId}/channels/{channelId}/members | internal/api/bot.go | bot channel-members | - |
| 5 | Calendar | GET | /users/{userId}/calendar-personals | internal/api/calendar.go | calendar list-calendars | - |
| 6 | Calendar | GET | /users/{userId}/calendar | internal/api/calendar.go | calendar list-calendars --default | - |
| 7 | Calendar | GET | /users/{userId}/calendars/{calendarId}/events | internal/api/calendar.go | calendar list-events | - |
| 8 | Calendar | GET | /users/{userId}/calendars/{calendarId}/events/{eventId} | internal/api/calendar.go | calendar get-event | - |
| 9 | Calendar | POST | /users/{userId}/calendars/{calendarId}/events | internal/api/calendar.go | calendar create-event | - |
| 10 | Board | GET | /boards | internal/api/board.go | board list | - |
| 11 | Board | GET | /boards/{boardId} | internal/api/board.go | board get | - |
| 12 | Board | GET | /boards/{boardId}/posts | internal/api/board.go | board list-posts | - |
| 13 | Board | GET | /boards/{boardId}/posts/{postId} | internal/api/board.go | board get-post | - |
| 14 | Board | POST | /boards/{boardId}/posts | internal/api/board.go | board create-post | - |
| 15 | Board | PUT | /boards/{boardId}/posts/{postId} | internal/api/board.go | board update-post | - |
| 16 | Board | DELETE | /boards/{boardId}/posts/{postId} | internal/api/board.go | board delete-post | - |
| 17 | Board | GET | /boards/{boardId}/posts/{postId}/comments | internal/api/board.go | board list-comments | - |
| 18 | Drive MyDrive | GET | /users/{userId}/drive | internal/api/drive.go | drive info | - |
| 19 | Drive MyDrive | GET | /users/{userId}/drive/files | internal/api/drive.go | drive list | - |
| 20 | Drive MyDrive | GET | /users/{userId}/drive/files/{folderId}/children | internal/api/drive.go | drive list --folder | - |
| 21 | Drive MyDrive | GET | /users/{userId}/drive/files/{fileId} | internal/api/drive.go | drive get | - |
| 22 | Drive MyDrive | GET | /users/{userId}/drive/files/{fileId}/download | internal/api/drive.go | drive download | - |
| 23 | Drive MyDrive | POST | /users/{userId}/drive/files | internal/api/drive.go | drive upload | - |
| 24 | Drive MyDrive | POST | /users/{userId}/drive/files/{folderId} | internal/api/drive.go | drive upload --folder | - |
| 25 | Drive MyDrive | POST | /users/{userId}/drive/files/createfolder | internal/api/drive.go | drive mkdir | - |
| 26 | Drive MyDrive | POST | /users/{userId}/drive/files/{parentId}/createfolder | internal/api/drive.go | drive mkdir --parent | - |
| 27 | Drive MyDrive | DELETE | /users/{userId}/drive/files/{fileId} | internal/api/drive.go | drive delete | - |
| 28 | Drive MyDrive | GET | /users/{userId}/drive/trash-files | internal/api/drive.go | drive trash-list | - |
| 29 | Drive MyDrive | POST | /users/{userId}/drive/trash-files/{fileId}/restore | internal/api/drive.go | drive trash-restore | - |
| 30 | Drive SharedFolders | GET | /users/{userId}/drive/sharedfolders | internal/api/drive.go | drive shared-folder list | - |
| 31 | Drive SharedFolders | GET | /users/{userId}/drive/sharedfolders/{sharedFolderId}/files | internal/api/drive.go | drive shared-folder files | - |
| 32 | Drive SharedDrive | GET | /sharedrives | internal/api/drive_shared.go | drive shared list-drives | - |
| 33 | Drive SharedDrive | GET | /sharedrives/{driveId} | internal/api/drive_shared.go | drive shared get-drive | - |
| 34 | Drive SharedDrive | GET | /sharedrives/{driveId}/files | internal/api/drive_shared.go | drive shared list | - |
| 35 | Drive SharedDrive | GET | /sharedrives/{driveId}/files/{folderId}/children | internal/api/drive_shared.go | drive shared list --folder | - |
| 36 | Drive SharedDrive | GET | /sharedrives/{driveId}/files/{fileId} | internal/api/drive_shared.go | drive shared get | - |
| 37 | Drive SharedDrive | GET | /sharedrives/{driveId}/files/{fileId}/download | internal/api/drive_shared.go | drive shared download | - |
| 38 | Drive SharedDrive | POST | /sharedrives/{driveId}/files | internal/api/drive_shared.go | drive shared upload | - |
| 39 | Drive GroupFolder | GET | /groups/{groupId}/folder | internal/api/drive_group.go | drive group get-folder | - |
| 40 | Drive GroupFolder | GET | /groups/{groupId}/folder/files | internal/api/drive_group.go | drive group list | - |
| 41 | Drive GroupFolder | GET | /groups/{groupId}/folder/files/{folderId}/children | internal/api/drive_group.go | drive group list --folder | - |
| 42 | Drive GroupFolder | GET | /groups/{groupId}/folder/files/{fileId} | internal/api/drive_group.go | drive group get | - |
| 43 | Mail | POST | /users/{userId}/mail | internal/api/mail.go | mail send | - |
| 44 | Mail | GET | /users/{userId}/mail/{mailId} | internal/api/mail.go | mail get | - |
| 45 | Mail | DELETE | /users/{userId}/mail/{mailId} | internal/api/mail.go | mail delete | - |
| 46 | Mail | GET | /users/{userId}/mail/mailfolders | internal/api/mail.go | mail list-folders | - |
| 47 | Mail | GET | /users/{userId}/mail/mailfolders/{folderId} | internal/api/mail.go | mail get-folder | - |
| 48 | Mail | GET | /users/{userId}/mail/mailfolders/{folderId}/children | internal/api/mail.go | mail list | - |
| 49 | Contact | GET | /contacts | internal/api/contact.go | contact list | - |
| 50 | Contact | GET | /users/{userId}/contacts | internal/api/contact.go | contact list-user | - |
| 51 | Contact | GET | /contacts/{contactId} | internal/api/contact.go | contact get | - |
| 52 | Contact | POST | /contacts | internal/api/contact.go | contact create | - |
| 53 | Contact | PATCH | /contacts/{contactId} | internal/api/contact.go | contact update | - |
| 54 | Contact | DELETE | /contacts/{contactId} | internal/api/contact.go | contact delete | - |
| 55 | Contact | GET | /contact-tags | internal/api/contact.go | contact list-tags | - |
| 56 | Contact | GET | /users/{userId}/contact-tags | internal/api/contact.go | contact list-user-tags | - |
| 57 | Directory | GET | /users | internal/api/directory.go | directory list-users | - |
| 58 | Directory | GET | /users/{userId} | internal/api/directory.go | directory get-user | - |
| 59 | Directory | GET | /groups | internal/api/directory.go | directory list-groups | - |
| 60 | Directory | GET | /groups/{groupId} | internal/api/directory.go | directory get-group | - |
| 61 | Directory | GET | /orgunits | internal/api/directory.go | directory list-orgunits | - |
| 62 | Directory | GET | /orgunits/{orgUnitId} | internal/api/directory.go | directory get-orgunit | - |
| 63 | Directory | GET | /directory/levels | internal/api/directory.go | directory list-levels | - |
| 64 | Directory | GET | /directory/positions | internal/api/directory.go | directory list-positions | - |
| 65 | Directory | GET | /directory/user-types | internal/api/directory.go | directory list-user-types | - |
| 66 | Directory | GET | /directory/employment-types | internal/api/directory.go | directory list-employment-types | - |
| 67 | Approval | GET | /business-support/approval/users/{userId}/documents | internal/api/approval.go | approval list | - |
| 68 | Approval | GET | /business-support/approval/documents | internal/api/approval.go | approval list-all | - |
| 69 | Approval | GET | /business-support/approval/documents/{documentId} | internal/api/approval.go | approval get | - |
| 70 | Approval | GET | /business-support/approval/categories | internal/api/approval.go | approval list-categories | - |
| 71 | Approval | GET | /business-support/approval/categories/{categoryId} | internal/api/approval.go | approval get-category | - |
| 72 | Approval | GET | /business-support/approval/document-forms | internal/api/approval.go | approval list-forms | - |
| 73 | Attendance | GET | /business-support/attendance/users/{userId}/status | internal/api/attendance.go | attendance status | - |
| 74 | Attendance | POST | /business-support/attendance/users/{userId}/clock-in | internal/api/attendance.go | attendance clock-in | - |
| 75 | Attendance | POST | /business-support/attendance/users/{userId}/clock-out | internal/api/attendance.go | attendance clock-out | - |
| 76 | Attendance | GET | /business-support/attendance/absences | internal/api/attendance.go | attendance list-absences | - |
| 77 | Attendance | GET | /business-support/attendance/annual-leaves | internal/api/attendance.go | attendance list-annual-leaves | - |
| 78 | HR | GET | /business-support/human-resource/extension-properties | internal/api/hr.go | hr list-extension-properties | - |
| 79 | HR | GET | /business-support/human-resource/user/{userId}/extension-properties | internal/api/hr.go | hr get-user-properties | - |
| 80 | HR | GET | /business-support/human-resource/leave-of-absences | internal/api/hr.go | hr list-leave-types | - |
| 81 | HR | GET | /business-support/human-resource/on-leave-users | internal/api/hr.go | hr list-on-leave | - |
| 82 | Business Place | GET | /business-support/business-places | internal/api/businessplace.go | business-place list | - |
| 83 | Business Place | GET | /business-support/business-places/{businessPlaceId} | internal/api/businessplace.go | business-place get | - |
| 84 | Business Place | POST | /business-support/business-places | internal/api/businessplace.go | business-place create | - |
| 85 | Business Place | PATCH | /business-support/business-places/{businessPlaceId} | internal/api/businessplace.go | business-place update | - |
| 86 | Business Place | DELETE | /business-support/business-places/{businessPlaceId} | internal/api/businessplace.go | business-place delete | - |
| 87 | Task | GET | /users/{userId}/tasks | internal/api/task.go | task list | - |
| 88 | Task | GET | /tasks/{taskId} | internal/api/task.go | task get | - |
| 89 | Task | POST | /users/{userId}/tasks | internal/api/task.go | task create | - |
| 90 | Task | PATCH | /tasks/{taskId} | internal/api/task.go | task update | - |
| 91 | Task | DELETE | /tasks/{taskId} | internal/api/task.go | task delete | - |
| 92 | Task | GET | /users/{userId}/task-categories | internal/api/task.go | task list-categories | - |
| 93 | Note | POST | /groups/{groupId}/note | internal/api/note.go | note create | - |
| 94 | Note | DELETE | /groups/{groupId}/note | internal/api/note.go | note delete | - |
| 95 | Note | GET | /groups/{groupId}/note/posts | internal/api/note.go | note list-posts | - |
| 96 | Note | GET | /groups/{groupId}/note/posts/{postId} | internal/api/note.go | note get-post | - |
| 97 | Note | POST | /groups/{groupId}/note/posts | internal/api/note.go | note create-post | - |
| 98 | Note | PUT | /groups/{groupId}/note/posts/{postId} | internal/api/note.go | note update-post | - |
| 99 | Note | DELETE | /groups/{groupId}/note/posts/{postId} | internal/api/note.go | note delete-post | - |
| 100 | Form | GET | /forms/{formId}/responses | internal/api/form.go | form list-responses | - |
| 101 | Form | GET | /forms/{formId}/responses/{responseId}/attachments/{attachmentId} | internal/api/form.go | form download-attachment | - |
| 102 | Audit | GET | /audits/logs/download | internal/api/audit.go | audit download-logs | - |
| 103 | Audit | GET | /audits/policy-groups | internal/api/audit.go | audit list-policy-groups | - |
| 104 | Monitoring | GET | /monitoring/message-contents/download | internal/api/audit.go | monitoring download-messages | - |
| 105 | SCIM Users | GET | /Users | internal/api/scim.go | scim list-users | - |
| 106 | SCIM Users | GET | /Users/{id} | internal/api/scim.go | scim get-user | - |
| 107 | SCIM Users | POST | /Users | internal/api/scim.go | scim create-user | - |
| 108 | SCIM Users | PUT | /Users/{id} | internal/api/scim.go | scim update-user | - |
| 109 | SCIM Users | PATCH | /Users/{id} | internal/api/scim.go | scim patch-user | - |
| 110 | SCIM Users | DELETE | /Users/{id} | internal/api/scim.go | scim delete-user | - |
| 111 | SCIM Groups | GET | /Groups | internal/api/scim.go | scim list-groups | - |
| 112 | SCIM Groups | GET | /Groups/{id} | internal/api/scim.go | scim get-group | - |
| 113 | SCIM Groups | POST | /Groups | internal/api/scim.go | scim create-group | - |
| 114 | SCIM Groups | PUT | /Groups/{id} | internal/api/scim.go | scim update-group | - |
| 115 | SCIM Groups | PATCH | /Groups/{id} | internal/api/scim.go | scim patch-group | - |
| 116 | SCIM Groups | DELETE | /Groups/{id} | internal/api/scim.go | scim delete-group | - |

## Smoke Test 현황

`cmd/smoke_test.go`에 등록된 API 관련 테스트:
- `TestSmoke_BotSend_MissingTarget` — bot send (flag 검증)
- `TestSmoke_BotSend_ConflictingFlags` — bot send (flag 충돌 검증)

그 외는 version, help, config, auth 등 인프라 테스트만 존재.
API endpoint 116개 중 smoke test가 있는 것은 bot send 관련 2건(#1, #2)만 해당됨.
