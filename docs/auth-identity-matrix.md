# Auth/Identity Matrix

NAVER WORKS REST API v1.0 전체 endpoint에 대한 인증/식별 요건 매트릭스.
CLI 구현 시 각 도메인 커맨드의 인증 분기, userId 처리, scope 설정의 SSOT로 사용한다.

## 범례

| 기호 | 의미 |
|------|------|
| ✓ | 지원함 |
| ✗ | 지원하지 않음 |
| N/A | 해당 없음 (userId 파라미터 자체가 없음) |

**CLI identity 처리 방식:**
- `--user-id (profile default)` — `resolveUserID()` 헬퍼로 처리. `--user-id` 플래그 → `default_calendar_user_id` config → 에러. JWT 모드에서 `me` 사용 불가.
- `--bot-id (config fallback)` — `resolveBotID()` 헬퍼로 처리. `--bot-id` 플래그 → `bot_id` config → 에러.
- `(none)` — identity 파라미터 불필요. API 클라이언트 토큰만으로 호출.

---

## Bot

| Operation | Scope | OAuth | JWT | userId=me | CLI identity |
|-----------|-------|-------|-----|-----------|--------------|
| sendMessageToUser | `bot.message` | ✓ | ✓ | N/A | `--bot-id` (config fallback), `--to` |
| sendMessageToChannel | `bot.message` | ✓ | ✓ | N/A | `--bot-id` (config fallback), `--channel` |
| getChannel | `bot.read` | ✓ | ✓ | N/A | `--bot-id` (config fallback) |
| listChannelMembers | `bot.read` | ✓ | ✓ | N/A | `--bot-id` (config fallback) |
| createChannel | `bot` | ✓ | ✓ | N/A | `--bot-id` (config fallback) |
| leaveChannel | `bot` | ✓ | ✓ | N/A | `--bot-id` (config fallback) |
| listBots | `bot.read` | ✓ | ✓ | N/A | (none) |
| getBot | `bot.read` | ✓ | ✓ | N/A | (none) |
| createBot | `bot` | ✓ | ✓ | N/A | (none) |
| updateBot | `bot` | ✓ | ✓ | N/A | (none) |
| deleteBot | `bot` | ✓ | ✓ | N/A | (none) |
| getRichMenu | `bot.read` | ✓ | ✓ | N/A | `--bot-id` (config fallback) |
| listRichMenus | `bot.read` | ✓ | ✓ | N/A | `--bot-id` (config fallback) |
| createRichMenu | `bot` | ✓ | ✓ | N/A | `--bot-id` (config fallback) |
| deleteRichMenu | `bot` | ✓ | ✓ | N/A | `--bot-id` (config fallback) |
| getPersistentMenu | `bot.read` | ✓ | ✓ | N/A | `--bot-id` (config fallback) |
| upsertPersistentMenu | `bot` | ✓ | ✓ | N/A | `--bot-id` (config fallback) |
| deletePersistentMenu | `bot` | ✓ | ✓ | N/A | `--bot-id` (config fallback) |
| registerDomain | `bot` | ✓ | ✓ | N/A | `--bot-id` (config fallback) |
| listDomains | `bot.read` | ✓ | ✓ | N/A | `--bot-id` (config fallback) |

**Scope 패턴:** `bot`(읽기/쓰기), `bot.read`(읽기 전용), `bot.message`(메시지 전송 전용)

---

## Calendar

| Operation | Scope | OAuth | JWT | userId=me | CLI identity |
|-----------|-------|-------|-----|-----------|--------------|
| listCalendars | `calendar.read` | ✓ | ✓ | ✓ | `--user-id` (profile default) |
| getDefaultCalendar | `calendar.read` | ✓ | ✓ | ✓ | `--user-id` (profile default) |
| getCalendarPersonal | `calendar.read` | ✓ | ✓ | ✓ | `--user-id` (profile default) |
| patchCalendarPersonal | `calendar` | ✓ | ✓ | ✓ | `--user-id` (profile default) |
| createCalendar | `calendar` | ✓ | ✓ | N/A | (none) |
| getCalendar | `calendar.read` | ✓ | ✓ | N/A | (none) |
| patchCalendar | `calendar` | ✓ | ✓ | N/A | (none) |
| deleteCalendar | `calendar` | ✓ | ✓ | N/A | (none) |
| removeUserFromCalendar | `calendar` | ✓ | ✓ | ✓ | `--user-id` (profile default) |
| listEvents | `calendar.read` | ✓ | ✓ | ✓ | `--user-id` (profile default) |
| getEvent | `calendar.read` | ✓ | ✓ | ✓ | `--user-id` (profile default) |
| createEvent | `calendar` | ✓ | ✓ | ✓ | `--user-id` (profile default) |
| updateEvent | `calendar` | ✓ | ✓ | ✓ | `--user-id` (profile default) |
| deleteEvent | `calendar` | ✓ | ✓ | ✓ | `--user-id` (profile default) |

**Scope 패턴:** `calendar`(읽기/쓰기), `calendar.read`(읽기 전용)
**userId=me:** OAuth 로그인한 사용자 본인. JWT에서는 명시적 userId 필수.

---

## Directory

| Operation | Scope | OAuth | JWT | userId=me | CLI identity |
|-----------|-------|-------|-----|-----------|--------------|
| listUsers | `user.read` / `directory.read` | ✓ | ✓ | N/A | (none) |
| getUser | `user.read` / `directory.read` | ✓ | ✓ | N/A | (none) |
| createUser | `user` / `directory` | ✓ | ✓ | N/A | (none) |
| updateUser | `user` / `directory` | ✓ | ✓ | N/A | (none) |
| patchUser | `user` / `directory` | ✓ | ✓ | N/A | (none) |
| deleteUser | `user` / `directory` | ✓ | ✓ | N/A | (none) |
| listGroups | `group.read` / `directory.read` | ✓ | ✓ | N/A | (none) |
| getGroup | `group.read` / `directory.read` | ✓ | ✓ | N/A | (none) |
| createGroup | `group` / `directory` | ✓ | ✓ | N/A | (none) |
| deleteGroup | `group` / `directory` | ✓ | ✓ | N/A | (none) |
| listOrgUnits | `orgunit.read` / `directory.read` | ✓ | ✓ | N/A | (none) |
| getOrgUnit | `orgunit.read` / `directory.read` | ✓ | ✓ | N/A | (none) |
| createOrgUnit | `orgunit` / `directory` | ✓ | ✓ | N/A | (none) |
| deleteOrgUnit | `orgunit` / `directory` | ✓ | ✓ | N/A | (none) |
| listLevels | `directory.read` | ✓ | ✓ | N/A | (none) |
| listPositions | `directory.read` | ✓ | ✓ | N/A | (none) |
| listUserTypes | `directory.read` | ✓ | ✓ | N/A | (none) |
| listEmploymentTypes | `directory.read` | ✓ | ✓ | N/A | (none) |
| getUserProfile | `user.profile.read` / `user.read` | ✓ | ✓ | N/A | (none) |
| getUserEmail | `user.email.read` / `user.read` | ✓ | ✓ | N/A | (none) |

**Scope 패턴:**
- 전체: `directory`(읽기/쓰기), `directory.read`(읽기 전용)
- 구성원: `user`, `user.read`, `user.profile.read`, `user.email.read`
- 그룹: `group`, `group.read`
- 조직: `orgunit`, `orgunit.read`

---

## Board

| Operation | Scope | OAuth | JWT | userId=me | CLI identity |
|-----------|-------|-------|-----|-----------|--------------|
| listBoards | `board.read` | ✓ | ✓ | N/A | (none) |
| getBoard | `board.read` | ✓ | ✓ | N/A | (none) |
| createBoard | `board` | ✓ | ✓ | N/A | (none) |
| updateBoard | `board` | ✓ | ✓ | N/A | (none) |
| deleteBoard | `board` | ✓ | ✓ | N/A | (none) |
| listPosts | `board.read` | ✓ | ✓ | N/A | (none) |
| getPost | `board.read` | ✓ | ✓ | N/A | (none) |
| createPost | `board` | ✓ | ✓ | N/A | (none) |
| updatePost | `board` | ✓ | ✓ | N/A | (none) |
| deletePost | `board` | ✓ | ✓ | N/A | (none) |
| listComments | `board.read` | ✓ | ✓ | N/A | (none) |
| createComment | `board` | ✓ | ✓ | N/A | (none) |
| deleteComment | `board` | ✓ | ✓ | N/A | (none) |

**Scope 패턴:** `board`(읽기/쓰기), `board.read`(읽기 전용)

---

## Drive

> **중요:** Drive API는 **OAuth(구성원 계정) Access Token만 사용 가능**. JWT(서비스 계정) 불가.

### MyDrive

| Operation | Scope | OAuth | JWT | userId=me | CLI identity |
|-----------|-------|-------|-----|-----------|--------------|
| getDriveInfo | `file.read` | ✓ | ✗ | ✓ | `--user-id` (profile default) |
| listFiles | `file.read` | ✓ | ✗ | ✓ | `--user-id` (profile default) |
| getFile | `file.read` | ✓ | ✗ | ✓ | `--user-id` (profile default) |
| getDownloadURL | `file.read` | ✓ | ✗ | ✓ | `--user-id` (profile default) |
| createUploadURL | `file` | ✓ | ✗ | ✓ | `--user-id` (profile default) |
| createFolder | `file` | ✓ | ✗ | ✓ | `--user-id` (profile default) |
| deleteFile | `file` | ✓ | ✗ | ✓ | `--user-id` (profile default) |
| listTrashFiles | `file.read` | ✓ | ✗ | ✓ | `--user-id` (profile default) |
| restoreTrashFile | `file` | ✓ | ✗ | ✓ | `--user-id` (profile default) |

### SharedDrive

| Operation | Scope | OAuth | JWT | userId=me | CLI identity |
|-----------|-------|-------|-----|-----------|--------------|
| listDrives | `file.read` | ✓ | ✗ | N/A | (none) |
| getDrive | `file.read` | ✓ | ✗ | N/A | (none) |
| listFiles | `file.read` | ✓ | ✗ | N/A | (none) |
| getFile | `file.read` | ✓ | ✗ | N/A | (none) |
| getDownloadURL | `file.read` | ✓ | ✗ | N/A | (none) |
| createUploadURL | `file` | ✓ | ✗ | N/A | (none) |

### GroupFolder

| Operation | Scope | OAuth | JWT | userId=me | CLI identity |
|-----------|-------|-------|-----|-----------|--------------|
| getFolder | `group.folder.read` | ✓ | ✗ | N/A | (none) |
| listFiles | `group.folder.read` | ✓ | ✗ | N/A | (none) |
| getFile | `group.folder.read` | ✓ | ✗ | N/A | (none) |

### SharedFolder

| Operation | Scope | OAuth | JWT | userId=me | CLI identity |
|-----------|-------|-------|-----|-----------|--------------|
| listSharedFolders | `file.read` | ✓ | ✗ | ✓ | `--user-id` (profile default) |
| listSharedFolderFiles | `file.read` | ✓ | ✗ | ✓ | `--user-id` (profile default) |

**Scope 패턴:** `file`(읽기/쓰기), `file.read`(읽기 전용), `group.folder`(조직/그룹 폴더), `group.folder.read`

---

## Contact

| Operation | Scope | OAuth | JWT | userId=me | CLI identity |
|-----------|-------|-------|-----|-----------|--------------|
| listContacts | `contact.read` | ✓ | ✓ | N/A | (none) |
| listUserContacts | `contact.read` | ✓ | ✓ | ✓ | `--user-id` (profile default) |
| getContact | `contact.read` | ✓ | ✓ | N/A | (none) |
| createContact | `contact` | ✓ | ✓ | N/A | (none) |
| updateContact | `contact` | ✓ | ✓ | N/A | (none) |
| deleteContact | `contact` | ✓ | ✓ | N/A | (none) |
| listTags | `contact.read` | ✓ | ✓ | N/A | (none) |
| listUserTags | `contact.read` | ✓ | ✓ | ✓ | `--user-id` (profile default) |

**Scope 패턴:** `contact`(읽기/쓰기), `contact.read`(읽기 전용)

---

## Mail

> **중요:** Mail API는 **OAuth(구성원 계정) Access Token만 사용 가능**. JWT(서비스 계정) 불가.

| Operation | Scope | OAuth | JWT | userId=me | CLI identity |
|-----------|-------|-------|-----|-----------|--------------|
| sendMail | `mail` | ✓ | ✗ | ✓ | `--user-id` (profile default) |
| getMail | `mail.read` | ✓ | ✗ | ✓ | `--user-id` (profile default) |
| deleteMail | `mail` | ✓ | ✗ | ✓ | `--user-id` (profile default) |
| listFolders | `mail.read` | ✓ | ✗ | ✓ | `--user-id` (profile default) |
| getFolder | `mail.read` | ✓ | ✗ | ✓ | `--user-id` (profile default) |
| listMails | `mail.read` | ✓ | ✗ | ✓ | `--user-id` (profile default) |

**Scope 패턴:** `mail`(읽기/쓰기), `mail.read`(읽기 전용)

---

## Note

> **중요:** Note API는 **OAuth(구성원 계정) Access Token만 사용 가능**. JWT(서비스 계정) 불가.

| Operation | Scope | OAuth | JWT | userId=me | CLI identity |
|-----------|-------|-------|-----|-----------|--------------|
| createNote | `group.note` | ✓ | ✗ | N/A | (none) |
| deleteNote | `group.note` | ✓ | ✗ | N/A | (none) |
| listPosts | `group.note.read` | ✓ | ✗ | N/A | (none) |
| getPost | `group.note.read` | ✓ | ✗ | N/A | (none) |
| createPost | `group.note` | ✓ | ✗ | N/A | (none) |
| updatePost | `group.note` | ✓ | ✗ | N/A | (none) |
| deletePost | `group.note` | ✓ | ✗ | N/A | (none) |

**Scope 패턴:** `group.note`(읽기/쓰기), `group.note.read`(읽기 전용)

---

## Task

> **중요:** Task API는 **OAuth(구성원 계정) Access Token만 사용 가능**. JWT(서비스 계정) 불가.

| Operation | Scope | OAuth | JWT | userId=me | CLI identity |
|-----------|-------|-------|-----|-----------|--------------|
| listTasks | `task.read` | ✓ | ✗ | ✓ | `--user-id` (profile default) |
| getTask | `task.read` | ✓ | ✗ | N/A | (none) |
| createTask | `task` | ✓ | ✗ | ✓ | `--user-id` (profile default) |
| updateTask | `task` | ✓ | ✗ | N/A | (none) |
| deleteTask | `task` | ✓ | ✗ | N/A | (none) |
| listCategories | `task.read` | ✓ | ✗ | ✓ | `--user-id` (profile default) |

**Scope 패턴:** `task`(읽기/쓰기), `task.read`(읽기 전용)

---

## Form

> **중요:** Form API는 **OAuth(구성원 계정) Access Token만 사용 가능**. JWT(서비스 계정) 불가.

| Operation | Scope | OAuth | JWT | userId=me | CLI identity |
|-----------|-------|-------|-----|-----------|--------------|
| listResponses | `form.read` | ✓ | ✗ | N/A | (none) |
| downloadAttachment | `form.read` | ✓ | ✗ | N/A | (none) |

**Scope 패턴:** `form`(읽기/쓰기), `form.read`(읽기 전용)

---

## Approval

| Operation | Scope | OAuth | JWT | userId=me | CLI identity |
|-----------|-------|-------|-----|-----------|--------------|
| listUserDocuments | `businessSupport.approval.read` | ✓ | ✓ | ✓ | `--user-id` (profile default) |
| listDocuments | `businessSupport.approval.read` | ✓ | ✓ | N/A | (none) |
| getDocument | `businessSupport.approval.read` | ✓ | ✓ | N/A | (none) |
| createUserDocument | `businessSupport.approval` | ✓ | ✓ | ✓ | `--user-id` (profile default) |
| createImportedDocument | `businessSupport.approval` | ✓ | ✓ | N/A | (none) |
| listCategories | `businessSupport.approval.read` | ✓ | ✓ | N/A | (none) |
| getCategory | `businessSupport.approval.read` | ✓ | ✓ | N/A | (none) |
| listDocumentForms | `businessSupport.approval.read` | ✓ | ✓ | N/A | (none) |

**Scope 패턴:** `businessSupport.approval`(읽기/쓰기), `businessSupport.approval.read`(읽기 전용), `businessSupport.read`(전체 읽기)

---

## Attendance

| Operation | Scope | OAuth | JWT | userId=me | CLI identity |
|-----------|-------|-------|-----|-----------|--------------|
| getStatus | `businessSupport.attendance.read` | ✓ | ✓ | ✓ | `--user-id` (profile default) |
| clockIn | `businessSupport.attendance` | ✓ | ✓ | ✓ | `--user-id` (profile default) |
| clockOut | `businessSupport.attendance` | ✓ | ✓ | ✓ | `--user-id` (profile default) |
| listAbsences | `businessSupport.attendance.read` | ✓ | ✓ | N/A | (none) |
| listAnnualLeaves | `businessSupport.attendance.read` | ✓ | ✓ | N/A | (none) |
| listAbsenceSchedules | `businessSupport.attendance.read` | ✓ | ✓ | N/A | (none) |
| listTimecards | `businessSupport.attendance.read` | ✓ | ✓ | N/A | (none) |

**Scope 패턴:** `businessSupport.attendance`(읽기/쓰기), `businessSupport.attendance.read`(읽기 전용)

---

## Human Resource

| Operation | Scope | OAuth | JWT | userId=me | CLI identity |
|-----------|-------|-------|-----|-----------|--------------|
| listExtensionProperties | `businessSupport.humanResource.read` | ✓ | ✓ | N/A | (none) |
| getUserExtensionProperties | `businessSupport.humanResource.read` | ✓ | ✓ | N/A | (none) |
| listLeaveOfAbsences | `businessSupport.humanResource.read` | ✓ | ✓ | N/A | (none) |
| listOnLeaveUsers | `businessSupport.humanResource.read` | ✓ | ✓ | N/A | (none) |

**Scope 패턴:** `businessSupport.humanResource`(읽기/쓰기), `businessSupport.humanResource.read`(읽기 전용)

---

## Business Place

| Operation | Scope | OAuth | JWT | userId=me | CLI identity |
|-----------|-------|-------|-----|-----------|--------------|
| listBusinessPlaces | `businessSupport.businessPlace.read` | ✓ | ✓ | N/A | (none) |
| getBusinessPlace | `businessSupport.businessPlace.read` | ✓ | ✓ | N/A | (none) |
| createBusinessPlace | `businessSupport.businessPlace` | ✓ | ✓ | N/A | (none) |
| updateBusinessPlace | `businessSupport.businessPlace` | ✓ | ✓ | N/A | (none) |
| deleteBusinessPlace | `businessSupport.businessPlace` | ✓ | ✓ | N/A | (none) |

**Scope 패턴:** `businessSupport.businessPlace`(읽기/쓰기), `businessSupport.businessPlace.read`(읽기 전용)

---

## Audit

| Operation | Scope | OAuth | JWT | userId=me | CLI identity |
|-----------|-------|-------|-----|-----------|--------------|
| downloadLogs | `audit.read` | ✓ | ✓ | N/A | (none) |
| listPolicyGroups | `audit.read` | ✓ | ✓ | N/A | (none) |
| getPolicyGroup | `audit.read` | ✓ | ✓ | N/A | (none) |
| createPolicyGroup | `audit` | ✓ | ✓ | N/A | (none) |
| updatePolicyGroup | `audit` | ✓ | ✓ | N/A | (none) |
| deletePolicyGroup | `audit` | ✓ | ✓ | N/A | (none) |

**Scope 패턴:** `audit`(읽기/쓰기), `audit.read`(읽기 전용)

---

## Monitoring

| Operation | Scope | OAuth | JWT | userId=me | CLI identity |
|-----------|-------|-------|-----|-----------|--------------|
| downloadMessages | `monitoring.read` | ✓ | ✓ | N/A | (none) |

**Scope 패턴:** `monitoring.read`(읽기 전용만 존재)

---

## Security

| Operation | Scope | OAuth | JWT | userId=me | CLI identity |
|-----------|-------|-------|-----|-----------|--------------|
| getExternalBrowser | `security.external-browser.read` | ✓ | ✓ | N/A | (none) |
| enableExternalBrowser | `security.external-browser` | ✓ | ✓ | N/A | (none) |
| disableExternalBrowser | `security.external-browser` | ✓ | ✓ | N/A | (none) |

**Scope 패턴:** `security.external-browser`(읽기/쓰기), `security.external-browser.read`(읽기 전용)

---

## SCIM

> **별도 인증 방식.** OAuth/JWT Access Token이 아닌 **SCIM 전용 Long-Lived Token**을 사용한다.
> CLI에서는 `scim_access_token` config 값으로 관리하며, `buildScimClient()`로 별도 클라이언트를 생성한다.

| Operation | Auth | userId=me | CLI identity |
|-----------|------|-----------|--------------|
| listUsers | SCIM Token | N/A | (none) |
| getUser | SCIM Token | N/A | (none) |
| createUser | SCIM Token | N/A | (none) |
| updateUser (PUT) | SCIM Token | N/A | (none) |
| patchUser (PATCH) | SCIM Token | N/A | (none) |
| deleteUser | SCIM Token | N/A | (none) |
| listGroups | SCIM Token | N/A | (none) |
| getGroup | SCIM Token | N/A | (none) |
| createGroup | SCIM Token | N/A | (none) |
| updateGroup (PUT) | SCIM Token | N/A | (none) |
| patchGroup (PATCH) | SCIM Token | N/A | (none) |
| deleteGroup | SCIM Token | N/A | (none) |

**Base URL:** `https://www.worksapis.com/scim/v2` (일반 API의 `/v1.0`과 다름)
**Rate Limit:** API당 240 requests/min
**필터:** `eq` 연산자만 지원 (User→`userName`, Group→`displayName`)

---

## 도메인별 Scope 요약

| 도메인 | 읽기/쓰기 | 읽기 전용 |
|--------|-----------|-----------|
| Bot | `bot`, `bot.message` | `bot.read` |
| Calendar | `calendar` | `calendar.read` |
| Directory (전체) | `directory` | `directory.read` |
| Directory (구성원) | `user` | `user.read`, `user.profile.read`, `user.email.read` |
| Directory (그룹) | `group` | `group.read` |
| Directory (조직) | `orgunit` | `orgunit.read` |
| Board | `board` | `board.read` |
| Drive (파일) | `file` | `file.read` |
| Drive (그룹 폴더) | `group.folder` | `group.folder.read` |
| Contact | `contact` | `contact.read` |
| Mail | `mail` | `mail.read` |
| Note | `group.note` | `group.note.read` |
| Task | `task` | `task.read` |
| Form | `form` | `form.read` |
| Approval | `businessSupport.approval` | `businessSupport.approval.read` |
| Attendance | `businessSupport.attendance` | `businessSupport.attendance.read` |
| Human Resource | `businessSupport.humanResource` | `businessSupport.humanResource.read` |
| Business Place | `businessSupport.businessPlace` | `businessSupport.businessPlace.read` |
| Audit | `audit` | `audit.read` |
| Monitoring | — | `monitoring.read` |
| Security | `security.external-browser` | `security.external-browser.read` |
| SCIM | (별도 토큰) | (별도 토큰) |

---

## OAuth-Only 도메인 (JWT 불가)

아래 도메인은 구성원 계정(OAuth) Access Token만 지원하며, 서비스 계정(JWT) 토큰으로는 호출 불가:

| 도메인 | 이유 |
|--------|------|
| **Drive** | 사용자 개인 저장소에 접근하는 API |
| **Mail** | 사용자 개인 메일함에 접근하는 API |
| **Note** | 구성원 계정 전용 |
| **Task** | 구성원 계정 전용 |
| **Form** | 구성원 계정 전용 |

---

## CLI userId 처리 흐름

```
resolveUserID(cmd, defaultUID, authMethod):
  1. --user-id 플래그 확인
  2. 없으면 config의 default_calendar_user_id 사용
  3. 둘 다 없으면 에러
  4. "me" + JWT 조합이면 에러 ("JWT 모드에서는 --user-id me를 사용할 수 없습니다")
```

**`userId=me` 동작:**
- OAuth 로그인 시: 로그인한 사용자 본인으로 해석
- JWT 로그인 시: 사용 불가 (서비스 계정에 "본인"이라는 개념이 없음)

---

## CLI 기본 Scope

| 인증 방식 | 기본 scope (config 미설정 시) |
|-----------|------------------------------|
| OAuth | `openid profile bot directory calendar` |
| JWT | `bot directory calendar` |

필요한 도메인이 기본 scope에 포함되지 않으면, `naverworks config set scope "..."` 으로 추가해야 한다.
