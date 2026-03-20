# 나머지 13개 서비스 추가 구현 계획

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** nw-cli에 나머지 13개 네이버웍스 서비스를 추가하여 SDK 전체 API를 CLI로 노출한다.

**Architecture:** 기존 패턴을 반복한다. 각 서비스당 `internal/api/<service>.go` + `cmd/<service>.go`. 공통 헬퍼(`loadConfigAndToken`, `buildAPIClient`, `resolveBotID` 패턴)를 재사용한다. userId가 필요한 서비스는 calendar과 동일하게 `--user-id` 플래그를 사용한다.

**구현 순서:** 복잡도 낮은 것부터. Drive는 가장 복잡하므로 마지막.

---

## Task 1: Directory 확장 (orgunit, level, position 등)

기존 directory 커맨드에 서브커맨드 추가.

**Files:**
- Modify: `internal/api/directory.go`
- Modify: `cmd/directory.go`

**추가 커맨드:**

| 커맨드 | API | 메서드 |
|--------|-----|--------|
| `directory list-orgunits` | GET `/orgunits` | 조직 목록 |
| `directory get-orgunit <id>` | GET `/orgunits/{orgunitId}` | 조직 상세 |
| `directory list-levels` | GET `/levels` | 직급 목록 |
| `directory list-positions` | GET `/positions` | 직책 목록 |
| `directory list-user-types` | GET `/usertypes` | 사용자 유형 |
| `directory list-employment-types` | GET `/employmenttypes` | 고용 유형 |

**패턴:** 모두 GET + 페이지네이션. `list-users`와 동일 구조.

---

## Task 2: Contact 서비스

**Files:**
- Create: `internal/api/contact.go`
- Create: `cmd/contact.go`

**커맨드:**

| 커맨드 | API | 메서드 |
|--------|-----|--------|
| `contact list` | GET `/contacts` | 연락처 목록 |
| `contact get <id>` | GET `/contacts/{contactId}` | 연락처 상세 |
| `contact create --name <n> --email <e>` | POST `/contacts` | 연락처 생성 |
| `contact update <id> --name <n>` | PATCH `/contacts/{contactId}` | 연락처 수정 |
| `contact delete <id>` | DELETE `/contacts/{contactId}` | 연락처 삭제 |
| `contact list-tags` | GET `/contacts/tags` | 태그 목록 |

---

## Task 3: Board 서비스

**Files:**
- Create: `internal/api/board.go`
- Create: `cmd/board.go`

**커맨드:**

| 커맨드 | API | 메서드 |
|--------|-----|--------|
| `board list` | GET `/boards` | 게시판 목록 |
| `board get <boardId>` | GET `/boards/{boardId}` | 게시판 상세 |
| `board list-posts <boardId>` | GET `/boards/{boardId}/posts` | 게시글 목록 |
| `board get-post <boardId> <postId>` | GET `/boards/{boardId}/posts/{postId}` | 게시글 상세 |
| `board create-post <boardId> --title <t> --body <b>` | POST `/boards/{boardId}/posts` | 게시글 생성 |
| `board delete-post <boardId> <postId>` | DELETE `/boards/{boardId}/posts/{postId}` | 게시글 삭제 |
| `board list-comments <boardId> <postId>` | GET `/boards/{boardId}/posts/{postId}/comments` | 댓글 목록 |

---

## Task 4: Task 서비스

**Files:**
- Create: `internal/api/task.go`
- Create: `cmd/task.go`

**커맨드:**

| 커맨드 | API | 메서드 |
|--------|-----|--------|
| `task list --user-id <id>` | GET `/users/{userId}/tasks` | 할일 목록 |
| `task get <taskId>` | GET `/tasks/{taskId}` | 할일 상세 |
| `task create --user-id <id> --title <t>` | POST `/users/{userId}/tasks` | 할일 생성 |
| `task update <taskId> --title <t>` | PATCH `/tasks/{taskId}` | 할일 수정 |
| `task delete <taskId>` | DELETE `/tasks/{taskId}` | 할일 삭제 |
| `task list-categories --user-id <id>` | GET `/users/{userId}/task-categories` | 카테고리 목록 |

---

## Task 5: Note 서비스

**Files:**
- Create: `internal/api/note.go`
- Create: `cmd/note.go`

**커맨드:**

| 커맨드 | API | 메서드 |
|--------|-----|--------|
| `note create <groupId>` | POST `/groups/{groupId}/note` | 노트 생성 |
| `note delete <groupId>` | DELETE `/groups/{groupId}/note` | 노트 삭제 |
| `note list-posts <groupId>` | GET `/groups/{groupId}/note/posts` | 포스트 목록 |
| `note get-post <groupId> <postId>` | GET `/groups/{groupId}/note/posts/{postId}` | 포스트 상세 |

---

## Task 6: Mail 서비스

**Files:**
- Create: `internal/api/mail.go`
- Create: `cmd/mail.go`

**커맨드:**

| 커맨드 | API | 메서드 |
|--------|-----|--------|
| `mail send --user-id <id> --to <addr> --subject <s> --body <b>` | POST `/users/{userId}/mail` | 메일 발송 |
| `mail get --user-id <id> <mailId>` | GET `/users/{userId}/mail/{mailId}` | 메일 상세 |
| `mail delete --user-id <id> <mailId>` | DELETE `/users/{userId}/mail/{mailId}` | 메일 삭제 |
| `mail list-folders --user-id <id>` | GET `/users/{userId}/mail/mailfolders` | 폴더 목록 |
| `mail list --user-id <id> <folderId>` | GET `/users/{userId}/mail/mailfolders/{folderId}/children` | 폴더 내 메일 목록 |

---

## Task 7: Form 서비스

**Files:**
- Create: `internal/api/form.go`
- Create: `cmd/form.go`

**커맨드:**

| 커맨드 | API | 메서드 |
|--------|-----|--------|
| `form list-responses <formId>` | GET `/forms/{formId}/responses` | 설문 응답 목록 |

---

## Task 8: Approval 서비스

**Files:**
- Create: `internal/api/approval.go`
- Create: `cmd/approval.go`

**커맨드:**

| 커맨드 | API | 메서드 |
|--------|-----|--------|
| `approval list --user-id <id>` | GET `/business-support/approval/users/{userId}/documents` | 결재 문서 목록 |
| `approval get <documentId>` | GET `/business-support/approval/documents/{documentId}` | 결재 문서 상세 |
| `approval list-categories` | GET `/business-support/approval/categories` | 카테고리 목록 |
| `approval get-category <id>` | GET `/business-support/approval/categories/{categoryId}` | 카테고리 상세 |
| `approval list-forms` | GET `/business-support/approval/forms` | 양식 목록 |

---

## Task 9: Attendance 서비스

**Files:**
- Create: `internal/api/attendance.go`
- Create: `cmd/attendance.go`

**커맨드:**

| 커맨드 | API | 메서드 |
|--------|-----|--------|
| `attendance status --user-id <id>` | GET `/business-support/attendance/users/{userId}/status` | 근태 상태 |
| `attendance clock-in --user-id <id>` | POST `/business-support/attendance/users/{userId}/clock-in` | 출근 |
| `attendance clock-out --user-id <id>` | POST `/business-support/attendance/users/{userId}/clock-out` | 퇴근 |
| `attendance list-absences` | GET `/business-support/attendance/absences` | 부재 항목 |
| `attendance list-annual-leaves --user-id <id>` | GET `/business-support/attendance/users/{userId}/annual-leaves` | 연차 목록 |

---

## Task 10: Audit + Monitoring 서비스

**Files:**
- Create: `internal/api/audit.go`
- Create: `cmd/audit.go`

**커맨드:**

| 커맨드 | API | 메서드 |
|--------|-----|--------|
| `audit download-logs --from <date> --until <date>` | GET `/audits/logs/download` | 감사 로그 다운로드 URL |
| `audit list-policy-groups` | GET `/audits/policy-groups` | 정책 그룹 목록 |
| `monitoring download-messages --from <date> --until <date>` | GET `/monitoring/message-contents/download` | 메시지 콘텐츠 다운로드 URL |

> audit과 monitoring은 커맨드가 적으므로 하나의 task에서 처리.

---

## Task 11: Human Resource 서비스

**Files:**
- Create: `internal/api/hr.go`
- Create: `cmd/hr.go`

**커맨드:**

| 커맨드 | API | 메서드 |
|--------|-----|--------|
| `hr list-extension-properties` | GET `/human-resource/extension-properties` | 확장 속성 목록 |
| `hr get-user-properties <userId>` | GET `/human-resource/users/{userId}/extension-properties` | 사용자 확장 속성 |
| `hr list-leave-types` | GET `/human-resource/leave-of-absences` | 휴직 유형 |
| `hr list-on-leave` | GET `/human-resource/on-leave-users` | 휴직 중 사용자 |

---

## Task 12: Business Place 서비스

**Files:**
- Create: `internal/api/businessplace.go`
- Create: `cmd/businessplace.go`

**커맨드:**

| 커맨드 | API | 메서드 |
|--------|-----|--------|
| `business-place list` | GET `/business-support/business-places` | 사업장 목록 |
| `business-place get <id>` | GET `/business-support/business-places/{id}` | 사업장 상세 |
| `business-place create --name <n>` | POST `/business-support/business-places` | 사업장 생성 |
| `business-place update <id> --name <n>` | PATCH `/business-support/business-places/{id}` | 사업장 수정 |
| `business-place delete <id>` | DELETE `/business-support/business-places/{id}` | 사업장 삭제 |

---

## Task 13: SCIM 서비스

**특이사항:** Base URL이 `https://www.worksapis.com/scim/v2`로 다름. 별도 API client 또는 baseURL 오버라이드 필요.

**Files:**
- Create: `internal/api/scim.go`
- Create: `cmd/scim.go`

**커맨드:**

| 커맨드 | API | 메서드 |
|--------|-----|--------|
| `scim list-users` | GET `/Users` | SCIM 사용자 목록 |
| `scim get-user <id>` | GET `/Users/{id}` | SCIM 사용자 상세 |
| `scim create-user --userName <n>` | POST `/Users` | SCIM 사용자 생성 |
| `scim update-user <id>` | PUT `/Users/{id}` | SCIM 사용자 수정 |
| `scim delete-user <id>` | DELETE `/Users/{id}` | SCIM 사용자 삭제 |
| `scim list-groups` | GET `/Groups` | SCIM 그룹 목록 |
| `scim get-group <id>` | GET `/Groups/{id}` | SCIM 그룹 상세 |
| `scim create-group --displayName <n>` | POST `/Groups` | SCIM 그룹 생성 |
| `scim update-group <id>` | PUT `/Groups/{id}` | SCIM 그룹 수정 |
| `scim delete-group <id>` | DELETE `/Groups/{id}` | SCIM 그룹 삭제 |

---

## Task 14: Drive 서비스 (MyDrive)

Drive는 가장 복잡하므로 MyDrive/SharedDrive/GroupFolder/SharedFolder를 나눈다.

**Files:**
- Create: `internal/api/drive.go`
- Create: `cmd/drive.go`

**커맨드 (MyDrive):**

| 커맨드 | API | 메서드 |
|--------|-----|--------|
| `drive info --user-id <id>` | GET `/users/{userId}/drive/info` | 드라이브 정보 |
| `drive list --user-id <id> [folderId]` | GET `/users/{userId}/drive/files[/{folderId}/children]` | 파일 목록 |
| `drive get --user-id <id> <fileId>` | GET `/users/{userId}/drive/files/{fileId}` | 파일 상세 |
| `drive download --user-id <id> <fileId>` | GET `.../download` | 파일 다운로드 URL |
| `drive upload --user-id <id> [--folder <folderId>] <localPath>` | POST `.../files` | 업로드 URL 생성 → PUT 업로드 |
| `drive mkdir --user-id <id> [--parent <folderId>] --name <n>` | POST `.../createfolder` | 폴더 생성 |
| `drive delete --user-id <id> <fileId>` | DELETE `.../files/{fileId}` | 파일 삭제 |
| `drive trash-list --user-id <id>` | GET `.../trash` | 휴지통 |
| `drive trash-restore --user-id <id> <fileId>` | POST `.../trash/{fileId}/restore` | 복원 |

---

## Task 15: Drive 서비스 (SharedDrive)

**추가 커맨드:**

| 커맨드 | API | 메서드 |
|--------|-----|--------|
| `drive shared list-drives` | GET `/sharedrives` | 공유 드라이브 목록 |
| `drive shared get-drive <driveId>` | GET `/sharedrives/{driveId}` | 공유 드라이브 상세 |
| `drive shared list <driveId> [folderId]` | GET `/sharedrives/{driveId}/files/...` | 파일 목록 |
| `drive shared get <driveId> <fileId>` | GET `.../files/{fileId}` | 파일 상세 |
| `drive shared download <driveId> <fileId>` | GET `.../download` | 다운로드 |
| `drive shared upload <driveId> [--folder <folderId>] <path>` | POST `.../files` | 업로드 |
| `drive shared permissions <driveId>` | GET `.../permissions` | 권한 목록 |

---

## Task 16: Drive 서비스 (GroupFolder + SharedFolder)

**추가 커맨드:**

| 커맨드 | API | 메서드 |
|--------|-----|--------|
| `drive group list-folders` | GET `/drive/group-folders` | 그룹 폴더 목록 |
| `drive group list <folderId>` | GET `/drive/group-folders/{folderId}/files` | 파일 목록 |
| `drive group get <folderId> <fileId>` | GET `.../files/{fileId}` | 파일 상세 |
| `drive shared-folder list` | GET `/drive/shared-folders` | 공유 폴더 목록 |
| `drive shared-folder files <folderId>` | GET `/drive/shared-folders/{folderId}/files` | 파일 목록 |

---

## Task 17: 통합 테스트 + 최종 빌드

- `go test ./...` 전체 통과
- `go vet ./...`
- 바이너리 크기 확인 (목표: < 15MB)
- 모든 서브커맨드 `--help` 확인
- README 업데이트

---

## 태스크 의존성

```
Task 1 (Directory 확장) — 독립
Task 2~12 (Contact ~ BusinessPlace) — 모두 독립, 병렬 가능
Task 13 (SCIM) — API client baseURL 분기 필요
Task 14~16 (Drive) — 14 → 15 → 16 순차
Task 17 (통합) — 모두 완료 후
```

**병렬 실행 가능:** Task 1~12는 서로 독립. 서브에이전트로 병렬 디스패치 가능.
