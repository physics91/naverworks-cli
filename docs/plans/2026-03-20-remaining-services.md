# 나머지 13개 서비스 추가 구현 계획

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** naverworks에 나머지 13개 네이버웍스 서비스의 주요 CLI 명령을 추가한다 (MVP 범위). 전체 SDK API 노출은 phase 2에서 다룬다.

**Architecture:** 기존 패턴을 반복하되, 선행 인프라 확장이 필요하다:
1. HTTP client에 `Put`/`Patch`/`Delete` 메서드 추가
2. 공통 `resolveUserID` 헬퍼 (calendar의 `resolveCalendarUserID` 일반화)
3. 공통 페이지네이션 패턴 (`--cursor`/`--count`/`--all`) 재사용
4. SCIM 전용 config 키 (`scim_access_token`) + 별도 API client
5. Drive 전용 raw upload helper (resumable upload 지원)

---

## Task 0: HTTP client 확장 (선행 필수)

**Files:**
- Modify: `internal/api/client.go`
- Modify: `internal/api/client_test.go`

**추가 메서드:**

```go
func (c *Client) Put(path string, body []byte) (*Response, error)
func (c *Client) Patch(path string, body []byte) (*Response, error)
func (c *Client) Delete(path string) (*Response, error)
```

모두 `c.do(method, path, body)` 위임. 기존 `Get`/`Post`와 동일 패턴.

**테스트:** PUT 성공, PATCH 성공, DELETE 성공 케이스.

---

## Task 1: 공통 헬퍼 확장

**Files:**
- Modify: `cmd/helpers.go`

**변경:**
- `resolveUserID(cmd, defaultUID, authMethod)` — calendar/task/mail/attendance 등에서 공유
- SCIM용 `loadScimClient()` — config에서 `scim_access_token` 읽어 별도 Client 생성 (baseURL: `https://www.worksapis.com/scim/v2`)

**Config 확장:**
- `internal/config/config.go`에 `ScimAccessToken` 필드 추가
- `NW_SCIM_ACCESS_TOKEN` 환경변수 추가

---

## Task 2: Directory 확장

**Files:**
- Modify: `internal/api/directory.go`
- Modify: `cmd/directory.go`

**추가 커맨드 (SDK 경로 검증 완료):**

| 커맨드 | API |
|--------|-----|
| `directory list-orgunits` | GET `/orgunits` |
| `directory get-orgunit <id>` | GET `/orgunits/{orgUnitId}` |
| `directory list-levels` | GET `/directory/levels` |
| `directory list-positions` | GET `/directory/positions` |
| `directory list-user-types` | GET `/directory/user-types` |
| `directory list-employment-types` | GET `/directory/employment-types` |

---

## Task 3: Contact 서비스

**Files:**
- Create: `internal/api/contact.go`
- Create: `cmd/contact.go`

| 커맨드 | API |
|--------|-----|
| `contact list` | GET `/contacts` |
| `contact list-user <userId>` | GET `/users/{userId}/contacts` |
| `contact get <id>` | GET `/contacts/{contactId}` |
| `contact create --name <n> --email <e>` | POST `/contacts` |
| `contact update <id> --name <n>` | PATCH `/contacts/{contactId}` |
| `contact delete <id>` | DELETE `/contacts/{contactId}` |
| `contact list-tags` | GET `/contact-tags` |
| `contact list-user-tags --user-id <id>` | GET `/users/{userId}/contact-tags` |

---

## Task 4: Board 서비스

**Files:**
- Create: `internal/api/board.go`
- Create: `cmd/board.go`

| 커맨드 | API |
|--------|-----|
| `board list` | GET `/boards` |
| `board get <boardId>` | GET `/boards/{boardId}` |
| `board list-posts <boardId>` | GET `/boards/{boardId}/posts` |
| `board get-post <boardId> <postId>` | GET `/boards/{boardId}/posts/{postId}` |
| `board create-post <boardId> --title <t> --body <b>` | POST `/boards/{boardId}/posts` |
| `board update-post <boardId> <postId> --title <t>` | PUT `/boards/{boardId}/posts/{postId}` |
| `board delete-post <boardId> <postId>` | DELETE `/boards/{boardId}/posts/{postId}` |
| `board list-comments <boardId> <postId>` | GET `/boards/{boardId}/posts/{postId}/comments` |

---

## Task 5: Task 서비스

**Files:**
- Create: `internal/api/task.go`
- Create: `cmd/task_cmd.go` (cobra의 task와 이름 충돌 방지)

| 커맨드 | API |
|--------|-----|
| `task list --user-id <id>` | GET `/users/{userId}/tasks` |
| `task get <taskId>` | GET `/tasks/{taskId}` |
| `task create --user-id <id> --title <t>` | POST `/users/{userId}/tasks` |
| `task update <taskId> --title <t>` | PATCH `/tasks/{taskId}` |
| `task delete <taskId>` | DELETE `/tasks/{taskId}` |
| `task list-categories --user-id <id>` | GET `/users/{userId}/task-categories` |

`--user-id`는 `resolveUserID` 공통 헬퍼 사용.

---

## Task 6: Note 서비스

**Files:**
- Create: `internal/api/note.go`
- Create: `cmd/note.go`

| 커맨드 | API |
|--------|-----|
| `note create <groupId>` | POST `/groups/{groupId}/note` |
| `note delete <groupId>` | DELETE `/groups/{groupId}/note` |
| `note list-posts <groupId>` | GET `/groups/{groupId}/note/posts` |
| `note get-post <groupId> <postId>` | GET `/groups/{groupId}/note/posts/{postId}` |
| `note create-post <groupId> --title <t> --body <b>` | POST `/groups/{groupId}/note/posts` |
| `note update-post <groupId> <postId> --title <t>` | PUT `/groups/{groupId}/note/posts/{postId}` |
| `note delete-post <groupId> <postId>` | DELETE `/groups/{groupId}/note/posts/{postId}` |

---

## Task 7: Mail 서비스

**Files:**
- Create: `internal/api/mail.go`
- Create: `cmd/mail.go`

| 커맨드 | API |
|--------|-----|
| `mail send --user-id <id> --to <addr> --subject <s> --body <b>` | POST `/users/{userId}/mail` |
| `mail get --user-id <id> <mailId>` | GET `/users/{userId}/mail/{mailId}` |
| `mail delete --user-id <id> <mailId>` | DELETE `/users/{userId}/mail/{mailId}` |
| `mail list-folders --user-id <id>` | GET `/users/{userId}/mail/mailfolders` |
| `mail get-folder --user-id <id> <folderId>` | GET `/users/{userId}/mail/mailfolders/{folderId}` |
| `mail list --user-id <id> <folderId>` | GET `/users/{userId}/mail/mailfolders/{folderId}/children` |

`--user-id`는 `resolveUserID` 공통 헬퍼 사용.

---

## Task 8: Form 서비스

**Files:**
- Create: `internal/api/form.go`
- Create: `cmd/form.go`

| 커맨드 | API |
|--------|-----|
| `form list-responses <formId>` | GET `/forms/{formId}/responses` |
| `form download-attachment <formId> <responseId> <attachmentId>` | GET `/forms/{formId}/responses/{responseId}/attachments/{attachmentId}` |

---

## Task 9: Approval 서비스

**Files:**
- Create: `internal/api/approval.go`
- Create: `cmd/approval.go`

| 커맨드 | API |
|--------|-----|
| `approval list --user-id <id>` | GET `/business-support/approval/users/{userId}/documents` |
| `approval list-all` | GET `/business-support/approval/documents` |
| `approval get <documentId>` | GET `/business-support/approval/documents/{documentId}` |
| `approval list-categories` | GET `/business-support/approval/categories` |
| `approval get-category <id>` | GET `/business-support/approval/categories/{categoryId}` |
| `approval list-forms` | GET `/business-support/approval/document-forms` |

---

## Task 10: Attendance 서비스

**Files:**
- Create: `internal/api/attendance.go`
- Create: `cmd/attendance.go`

| 커맨드 | API |
|--------|-----|
| `attendance status --user-id <id>` | GET `/business-support/attendance/users/{userId}/status` |
| `attendance clock-in --user-id <id> --date <YYYY-MM-DD> --time <HH:mm>` | POST `/business-support/attendance/users/{userId}/clock-in` |
| `attendance clock-out --user-id <id> --date <YYYY-MM-DD> --time <HH:mm>` | POST `/business-support/attendance/users/{userId}/clock-out` |
| `attendance list-absences` | GET `/business-support/attendance/absences` |
| `attendance list-annual-leaves` | GET `/business-support/attendance/annual-leaves` |

clock-in/clock-out body: `{"baseDate": "YYYY-MM-DD", "clockInTime": "HH:mm"}` (또는 `clockOutTime`)

---

## Task 11: Audit + Monitoring

**Files:**
- Create: `internal/api/audit.go`
- Create: `cmd/audit.go`

| 커맨드 | API |
|--------|-----|
| `audit download-logs --from <date> --until <date>` | GET `/audits/logs/download` |
| `audit list-policy-groups` | GET `/audits/policy-groups` |
| `monitoring download-messages --from <date> --until <date>` | GET `/monitoring/message-contents/download` |

---

## Task 12: Human Resource

**Files:**
- Create: `internal/api/hr.go`
- Create: `cmd/hr.go`

| 커맨드 | API |
|--------|-----|
| `hr list-extension-properties` | GET `/business-support/human-resource/extension-properties` |
| `hr get-user-properties <userId>` | GET `/business-support/human-resource/user/{userId}/extension-properties` |
| `hr list-leave-types` | GET `/business-support/human-resource/leave-of-absences` |
| `hr list-on-leave` | GET `/business-support/human-resource/on-leave-users` |

---

## Task 13: Business Place

**Files:**
- Create: `internal/api/businessplace.go`
- Create: `cmd/businessplace.go`

| 커맨드 | API |
|--------|-----|
| `business-place list` | GET `/business-support/business-places` |
| `business-place get <id>` | GET `/business-support/business-places/{id}` |
| `business-place create --name <n>` | POST `/business-support/business-places` |
| `business-place update <id> --name <n>` | PATCH `/business-support/business-places/{id}` |
| `business-place delete <id>` | DELETE `/business-support/business-places/{id}` |

---

## Task 14: SCIM 서비스

**특이사항:**
- Base URL: `https://www.worksapis.com/scim/v2`
- 인증: SCIM 전용 long-lived access token (OAuth/JWT와 별도)
- Config 키: `scim_access_token` / `NW_SCIM_ACCESS_TOKEN`

**Files:**
- Create: `internal/api/scim.go`
- Create: `cmd/scim.go`
- Modify: `cmd/helpers.go` (`loadScimClient` 추가)

| 커맨드 | API |
|--------|-----|
| `scim list-users` | GET `/Users` |
| `scim get-user <id>` | GET `/Users/{id}` |
| `scim create-user --userName <n>` | POST `/Users` |
| `scim update-user <id>` | PUT `/Users/{id}` |
| `scim patch-user <id>` | PATCH `/Users/{id}` |
| `scim delete-user <id>` | DELETE `/Users/{id}` |
| `scim list-groups` | GET `/Groups` |
| `scim get-group <id>` | GET `/Groups/{id}` |
| `scim create-group --displayName <n>` | POST `/Groups` |
| `scim update-group <id>` | PUT `/Groups/{id}` |
| `scim patch-group <id>` | PATCH `/Groups/{id}` |
| `scim delete-group <id>` | DELETE `/Groups/{id}` |

---

## Task 15: Drive — 공통 인프라 + MyDrive

**선행:** raw upload helper

**Files:**
- Create: `internal/api/drive.go` (공통 drive 유틸: upload, download URL 처리)
- Create: `internal/api/drive_mydrive.go`
- Create: `cmd/drive.go`

Raw upload helper:
```go
func (c *Client) UploadFile(uploadURL string, filePath string) error
// - uploadURL에 직접 PUT (Bearer 토큰 없음, Content-Type: application/octet-stream)
// - resumable upload offset 지원
```

**MyDrive 커맨드 (SDK 경로 검증 완료):**

| 커맨드 | API |
|--------|-----|
| `drive info --user-id <id>` | GET `/users/{userId}/drive` |
| `drive list --user-id <id> [--folder <folderId>]` | GET `/users/{userId}/drive/files` 또는 `GET .../files/{folderId}/children` |
| `drive get --user-id <id> <fileId>` | GET `/users/{userId}/drive/files/{fileId}` |
| `drive download --user-id <id> <fileId>` | GET `/users/{userId}/drive/files/{fileId}/download` |
| `drive upload --user-id <id> [--folder <folderId>] <localPath>` | POST `/users/{userId}/drive/files` (루트) 또는 POST `.../files/{folderId}` (하위) → PUT uploadUrl |
| `drive mkdir --user-id <id> [--parent <folderId>] --name <n>` | POST `/users/{userId}/drive/files/createfolder` (루트) 또는 POST `.../files/{folderId}/createfolder` (하위) |
| `drive delete --user-id <id> <fileId>` | DELETE `/users/{userId}/drive/files/{fileId}` |
| `drive trash-list --user-id <id>` | GET `/users/{userId}/drive/trash-files` |
| `drive trash-restore --user-id <id> <fileId>` | POST `/users/{userId}/drive/trash-files/{fileId}/restore` |

---

## Task 16: Drive — SharedDrive

**Files:**
- Create: `internal/api/drive_shared.go`
- Modify: `cmd/drive.go`

| 커맨드 | API |
|--------|-----|
| `drive shared list-drives` | GET `/sharedrives` |
| `drive shared get-drive <driveId>` | GET `/sharedrives/{driveId}` |
| `drive shared list <driveId> [--folder <folderId>]` | GET `/sharedrives/{driveId}/files` 또는 `.../files/{folderId}/children` |
| `drive shared get <driveId> <fileId>` | GET `/sharedrives/{driveId}/files/{fileId}` |
| `drive shared download <driveId> <fileId>` | GET `.../files/{fileId}/download` |
| `drive shared upload <driveId> [--folder <folderId>] <path>` | POST `.../files` → PUT uploadUrl |

---

## Task 17: Drive — GroupFolder + SharedFolder

**Files:**
- Create: `internal/api/drive_group.go`
- Create: `internal/api/drive_sharedfolder.go`
- Modify: `cmd/drive.go`

**GroupFolder (SDK: DriveGroupFolderApi — groupId 기반):**

| 커맨드 | API |
|--------|-----|
| `drive group get-folder <groupId>` | GET `/groups/{groupId}/folder` |
| `drive group list <groupId> [--folder <folderId>]` | GET `/groups/{groupId}/folder/files` 또는 `.../files/{folderId}/children` |
| `drive group get <groupId> <fileId>` | GET `/groups/{groupId}/folder/files/{fileId}` |

**SharedFolder (SDK: DriveSharedFolderApi — sharedfolders 세그먼트):**

| 커맨드 | API |
|--------|-----|
| `drive shared-folder list --user-id <id>` | GET `/users/{userId}/drive/sharedfolders` |
| `drive shared-folder files <sharedFolderId> --user-id <id>` | GET `/users/{userId}/drive/sharedfolders/{id}/files` |

---

## Task 18: 통합 테스트 + README 업데이트

- `go test ./...` 전체 통과
- `go vet ./...`
- 바이너리 크기 확인 (목표: < 15MB)
- 모든 서브커맨드 `--help` 확인
- README에 새 커맨드 추가

---

## 태스크 의존성

```
Task 0 (HTTP verb 확장) ─┬→ Task 2~13 (각 서비스, 병렬 가능)
Task 1 (공통 헬퍼)       ─┘
                              ├→ Task 14 (SCIM, 별도 auth)
                              └→ Task 15 (Drive MyDrive + upload helper)
                                    → Task 16 (Drive SharedDrive)
                                    → Task 17 (Drive Group/SharedFolder)
                                        → Task 18 (통합)
```

**실행 순서:**
1. Task 0 + 1 (선행 인프라)
2. Task 2~13 (병렬 가능)
3. Task 14 (SCIM)
4. Task 15 → 16 → 17 (Drive 순차)
5. Task 18 (통합)
