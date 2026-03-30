# NAVER WORKS API 100% 커버리지 구현 마스터 플랜

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** naverworks-sdk-kotlin 기준 538개 전체 API를 CLI에 100% 구현하여 완전한 API 커버리지를 달성한다.

**Architecture:** 기존 cmd/<domain>.go ↔ internal/api/<domain>.go 쌍 패턴을 유지하며, Phase 0(공통 인프라)을 선행 필수 단계로 두고 Phase 1~5는 각 Task의 Dependencies를 만족하는 범위에서 순차 또는 제한적 병렬로 진행한다.

**Tech Stack:** Go 1.22, Cobra CLI, net/http, encoding/json

**현재 상태:** 116/538 API 구현 (21.6%)
**목표:** 538/538 API 구현 (100%)
**추가 구현 필요:** 422개 API (Task 합산 410 + Coverage Gap 12)

---

## Coverage Ledger 규칙

- Source of truth: `naverworks-sdk-kotlin`의 538개 endpoint 목록
- 각 신규 endpoint는 정확히 하나의 Task에 매핑하고, 기존 116개 endpoint는 Baseline Ledger에 정확히 하나씩 매핑한다
- 각 Task 섹션은 반드시 `Files`, `Dependencies`, `API 메서드`, `CLI 커맨드`, `Definition of Done`을 포함한다
- 축약 표현(`동일 패턴`, `~30개`)만으로는 완료로 인정하지 않는다
- 모든 수치(Phase별 API 수, 총 태스크)는 본 문서 내 Task 합산으로만 갱신한다

### Source Snapshot

- 기준 SDK: `naverworks-sdk-kotlin` `main` branch
- 추출 일자: `2026-03-30`
- 538개 endpoint 집계 방법: SDK 내 모든 API service class의 method를 열거하여 HTTP method + path 쌍으로 전수 목록화
- 이후 기준 변경은 별도 PR에서만 허용한다

### Baseline Ledger (기존 구현 116개)

- 현재 이미 구현된 116개 endpoint는 별도 `docs/coverage-ledger-existing.md`에 `도메인 / HTTP / 경로 / 구현 파일 / smoke test` 형식으로 전수 기록한다
- 본 문서의 구현 Task는 Phase 1~5의 44개이며, 이들이 신규 422개 endpoint를 담당한다. Phase 0의 6개 Task는 인프라/ledger 작업이다
- 최종 538/538 판정은 `기존 116개 ledger + 본 문서의 신규 410개 Task + Coverage Reconciliation에서 확인된 12개 누락분`을 합산해 검증한다

## Phase 개요

| Phase | 범위 | 신규 API | 난이도 | 태스크 수 |
|-------|------|---------|--------|----------|
| 0 | 공통 인프라 | 0 | ★☆☆ | 6 |
| 1 | 기존 도메인 API 보완 (CUD 중심, 누락 GET/LIST 포함) | 125 | ★★☆ | 10 |
| 2 | Security 신규 도메인 | 3 | ★☆☆ | 1 |
| 3 | Bot 전체 구현 | 36 | ★★★ | 6 |
| 4 | Directory 전체 구현 | 121 | ★★★ | 12 |
| 5 | Drive 전체 구현 + Gap | 125+12 | ★★★ | 15 |
| **합계** | | **410+12** | | **50** |

---

## 공통 구현 패턴 가이드

### API 서비스 메서드 패턴

```go
// GET (단건)
func (s *DomainService) GetItem(id string) (*Response, error) {
    return s.client.Get(fmt.Sprintf("/items/%s", url.PathEscape(id)))
}

// GET (목록 + 페이지네이션)
func (s *DomainService) ListItems(cursor string, count int) (*Response, error) {
    return s.client.Get("/items" + BuildPaginationQuery(cursor, count))
}

// POST (생성 - JSON body)
func (s *DomainService) CreateItem(body map[string]interface{}) (*Response, error) {
    return s.client.PostJSON("/items", body)
}

// PUT (전체 수정)
func (s *DomainService) UpdateItem(id string, body map[string]interface{}) (*Response, error) {
    return s.client.PutJSON(fmt.Sprintf("/items/%s", url.PathEscape(id)), body)
}

// PATCH (부분 수정)
func (s *DomainService) PatchItem(id string, body map[string]interface{}) (*Response, error) {
    return s.client.PatchJSON(fmt.Sprintf("/items/%s", url.PathEscape(id)), body)
}

// DELETE
func (s *DomainService) DeleteItem(id string) (*Response, error) {
    return s.client.Delete(fmt.Sprintf("/items/%s", url.PathEscape(id)))
}

// POST (액션 - body 없음)
func (s *DomainService) ActionItem(id string) (*Response, error) {
    return s.client.Post(fmt.Sprintf("/items/%s/action", url.PathEscape(id)), nil)
}
```

### Cobra 커맨드 패턴

```go
// 생성 커맨드 (--json 플래그로 body 전달)
var domainCreateCmd = &cobra.Command{
    Use:   "create",
    Short: "항목 생성",
    RunE: func(cmd *cobra.Command, args []string) error {
        svc, err := newSvc(api.NewDomainService)
        if err != nil { return err }
        body, err := readJSONFlag(cmd)
        if err != nil { return err }
        resp, err := svc.CreateItem(body)
        if err != nil { return err }
        printResponse(resp)
        return nil
    },
}

// 수정 커맨드 (위치 인자 + --json)
var domainUpdateCmd = &cobra.Command{
    Use:   "update <id>",
    Short: "항목 수정",
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        svc, err := newSvc(api.NewDomainService)
        if err != nil { return err }
        body, err := readJSONFlag(cmd)
        if err != nil { return err }
        resp, err := svc.PatchItem(args[0], body)
        if err != nil { return err }
        printResponse(resp)
        return nil
    },
}

// 삭제 커맨드
var domainDeleteCmd = &cobra.Command{
    Use:   "delete <id>",
    Short: "항목 삭제",
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        svc, err := newSvc(api.NewDomainService)
        if err != nil { return err }
        resp, err := svc.DeleteItem(args[0])
        if err != nil { return err }
        printResponse(resp)
        return nil
    },
}
```

### 선행 인프라: readJSONFlag 헬퍼 (Phase 0)

CUD 커맨드가 대량으로 추가되므로, JSON body를 공통으로 읽는 헬퍼가 필요하다.

```go
// cmd/helpers.go
func readJSONFlag(cmd *cobra.Command) (map[string]interface{}, error) {
    jsonStr, _ := cmd.Flags().GetString("json")
    if jsonStr == "" {
        return nil, fmt.Errorf("--json 플래그가 필요합니다")
    }
    if jsonStr == "-" {
        data, err := io.ReadAll(os.Stdin)
        if err != nil { return nil, err }
        jsonStr = string(data)
    }
    var body map[string]interface{}
    if err := json.Unmarshal([]byte(jsonStr), &body); err != nil {
        return nil, fmt.Errorf("JSON 파싱 실패: %w", err)
    }
    return body, nil
}

// raw bytes 버전 (PUT/POST에서 직접 사용)
func readJSONFlagRaw(cmd *cobra.Command) ([]byte, error) {
    jsonStr, _ := cmd.Flags().GetString("json")
    if jsonStr == "" {
        return nil, fmt.Errorf("--json 플래그가 필요합니다")
    }
    if jsonStr == "-" {
        return io.ReadAll(os.Stdin)
    }
    // JSON 유효성 검증
    var js json.RawMessage
    if err := json.Unmarshal([]byte(jsonStr), &js); err != nil {
        return nil, fmt.Errorf("JSON 파싱 실패: %w", err)
    }
    return []byte(jsonStr), nil
}
```

### 스모크 테스트 패턴

```go
func TestSmoke_Domain_Create_MissingJSON(t *testing.T) {
    tmpDir := setupTestEnv(t)
    writeTestConfig(t, tmpDir)
    _, err := runCLI(t, "domain", "create")
    if err == nil || !strings.Contains(err.Error(), "--json") {
        t.Errorf("expected --json error, got: %v", err)
    }
}
```

---

## Phase 0: 공통 인프라 확장

### Task 0-0: Baseline Ledger 작성 (기존 116개)

**Files:**
- Create: `docs/coverage-ledger-existing.md`
- Create: `scripts/verify-coverage-ledger.go`

**Dependencies:** 없음

**API 메서드:** 없음 (기존 구현 전수 목록화 작업)

**CLI 커맨드:** 없음

**Definition of Done:**
1. 기존 구현 116개 endpoint가 `도메인 / HTTP / 경로 / 구현 파일 / smoke test` 형식으로 전수 기록됨
2. 중복/누락 없이 총 116개가 검증됨
3. 기존 116개 ledger, 본 문서에 명시된 신규 410개 endpoint, Coverage Reconciliation으로 식별되는 누락 12개를 합산해 총 538개 판정이 가능함
4. `scripts/verify-coverage-ledger.go`가 SDK endpoint 목록 + ledger 문서를 입력받아 `총계/누락/중복` 리포트를 출력함
5. Task 0-0 완료 시점에 자동 검증 결과를 근거로 Coverage Gap 후보 목록을 산출함

---

### Task 0-1: readJSONFlag / readJSONFlagRaw 헬퍼 추가

**Files:**
- Modify: `cmd/helpers.go`
- Test: `cmd/smoke_test.go`

**Dependencies:** 없음

**API 메서드:** 없음 (공통 헬퍼 태스크)

**CLI 커맨드:** 없음

**구현:**
- `readJSONFlag(cmd)` → `(map[string]interface{}, error)` — PATCH/POST JSON body 용
- `readJSONFlagRaw(cmd)` → `([]byte, error)` — PUT raw body 용
- `--json` 플래그: string, stdin(`-`) 지원, JSON 유효성 검증

**Definition of Done:**
1. `readJSONFlag`, `readJSONFlagRaw`가 `cmd/helpers.go`에 추가되고 컴파일 성공
2. `cmd/smoke_test.go`에 `--json` 미지정, stdin(`-`) 입력, JSON 파싱 실패 케이스 테스트 추가
3. 이후 CUD 커맨드는 공통 헬퍼를 사용하고 중복 JSON 파싱 코드를 추가하지 않음

---

### Task 0-2: raw body 전송 전략 확정

**Files:**
- Modify: `internal/api/client.go` (필요 시)
- Modify: 본 문서의 raw body 사용 Task 설명

**Dependencies:** Task 0-1

**API 메서드:** 없음 (전송 전략 확정 태스크)

**CLI 커맨드:** 없음

**구현:**
- 기본 원칙: raw JSON body가 필요한 endpoint는 새 `PostJSONRaw`를 만들지 않고 기존 `Post`/`Put`/`Patch`에 `readJSONFlagRaw` 결과를 전달한다
- 예외가 확인되면 그때만 `PostRaw`/`PutRaw`/`PatchRaw`를 추가하고 적용 Task 번호를 문서에 명시한다

**Definition of Done:**
1. raw body 전송 표준이 위 원칙으로 문서에 확정됨
2. `internal/api/client.go` 변경 필요 여부가 결정되고 후속 Task가 그 결정을 따름

### Task 0-3: 파일 업로드/다운로드 공통 헬퍼 추가

**Files:**
- Modify: `cmd/helpers.go`
- Modify: `internal/api/client.go`
- Test: `cmd/smoke_test.go`

**Dependencies:** 없음 (Task 0-1과 병렬 가능)

**API 메서드:** 없음 (공통 파일 입출력 헬퍼 태스크)

**CLI 커맨드:** 없음

**구현:**
- `readFileFlag(cmd *cobra.Command, flagName string) ([]byte, string, error)` — 파일 경로를 읽어 바이트 + 파일명 반환
- `client.UploadMultipart(path, fieldName, fileName string, data []byte) (*Response, error)` — multipart/form-data 업로드
- `client.DownloadFile(path string) ([]byte, http.Header, error)` — 바이너리 다운로드 (Content-Disposition 파싱)
- 첨부/이미지 API는 아래 3가지 방식 중 하나로만 구현한다:
  1. **presigned URL 업로드**: 업로드 URL 발급 API 호출 후 `client.UploadFile(url, filePath)`
  2. **multipart 업로드**: `client.UploadMultipart(path, field, name, data)` 사용
  3. **raw binary 업로드**: `client.Post(path, data)` + Content-Type 직접 지정
- 각 `--file` 커맨드 Task에는 `업로드 방식`, `Content-Type`을 반드시 명시한다
- 대상 도메인: Board attachment, Comment attachment, Contact photo, Note attachment, Approval attachment, Bot attachment, RichMenu image, User photo

**Definition of Done:**
1. `readFileFlag`, `UploadMultipart`, `DownloadFile` 컴파일 성공
2. 스모크 테스트에 `--file` 미지정 시 에러 검증 추가
3. 기존 `drive upload`의 presigned URL 방식과 공존 확인

---

### Task 0-4: 업로드 스펙 매트릭스 확정

**Files:**
- Create: `docs/upload-spec-matrix.md`
- Modify: 본 문서의 업로드 관련 Task 섹션

**Dependencies:** 없음

**API 메서드:** 없음 (사전 분석 태스크)

**CLI 커맨드:** 없음

**구현:**
- `naverworks-sdk-kotlin`과 공식 REST 문서를 대조해 Board attachment, Comment attachment, Contact photo, Note attachment, Approval attachment, Bot attachment, RichMenu image, User photo의 업로드 방식을 endpoint별로 확정한다
- 각 endpoint마다 `presigned URL | multipart/form-data | raw binary`, `Content-Type`, 필요 헤더, 응답 형식을 `docs/upload-spec-matrix.md`에 기록한다
- 관련 Task의 `업로드 사양` 섹션을 확정값으로 갱신하고 `착수 전 확정 필요` 문구를 제거한다

**Definition of Done:**
1. 업로드 관련 endpoint 전부가 `docs/upload-spec-matrix.md`에 기록됨
2. 업로드 관련 Task의 `업로드 사양` 섹션이 모두 확정값으로 치환됨
3. 미확정 업로드 Task가 0개임

---

### Task 0-5: Auth/Identity 매트릭스 확정

**Files:**
- Create: `docs/auth-identity-matrix.md`
- Modify: 본 문서의 전 구현 Task 섹션

**Dependencies:** Task 0-0

**API 메서드:** 없음 (인증/식별 규칙 확정 태스크)

**CLI 커맨드:** 없음

**Definition of Done:**
1. 538개 endpoint 전부에 대해 `Required scopes`, `OAuth/JWT 지원 여부`, `userId=me 허용 여부`, `CLI identity 처리 방식`을 `docs/auth-identity-matrix.md`에 기록함
2. 모든 구현 Task에 `Auth/Identity` 섹션을 추가하고 matrix 값으로 채움
3. `Auth/Identity` 미기입 Task가 0개임

---

## Phase 1: 기존 도메인 API 보완 (125개 API)

이미 list/get이 구현된 도메인에 create/update/delete 및 누락된 GET/LIST를 추가한다.

### Task 1-1: Calendar 완성 (14개 추가)

**Files:**
- Modify: `internal/api/calendar.go`
- Modify: `cmd/calendar.go`
- Modify: `cmd/smoke_test.go`

**Dependencies:** Phase 0 완료

**추가 API 메서드:**

| # | 메서드 | HTTP | 경로 |
|---|--------|------|------|
| 1 | CreateCalendar | POST | `/calendars` |
| 2 | GetCalendar | GET | `/calendars/{calendarId}` |
| 3 | PatchCalendar | PATCH | `/calendars/{calendarId}` |
| 4 | DeleteCalendar | DELETE | `/calendars/{calendarId}` |
| 5 | GetCalendarPersonal | GET | `/users/{userId}/calendar-personals/{calendarId}` |
| 6 | PatchCalendarPersonal | PATCH | `/users/{userId}/calendar-personals/{calendarId}` |
| 7 | RemoveUserFromCalendar | DELETE | `/users/{userId}/calendars/{calendarId}` |
| 8 | UpdateEvent | PUT | `/users/{userId}/calendars/{calendarId}/events/{eventId}` |
| 9 | DeleteEvent | DELETE | `/users/{userId}/calendars/{calendarId}/events/{eventId}` |
| 10 | CreateDefaultEvent | POST | `/users/{userId}/calendar/events` |
| 11 | ListDefaultEvents | GET | `/users/{userId}/calendar/events` |
| 12 | GetDefaultEvent | GET | `/users/{userId}/calendar/events/{eventId}` |
| 13 | UpdateDefaultEvent | PUT | `/users/{userId}/calendar/events/{eventId}` |
| 14 | DeleteDefaultEvent | DELETE | `/users/{userId}/calendar/events/{eventId}` |

**추가 CLI 커맨드:**

```
calendar create-calendar --json '{...}'
calendar get-calendar <calendarId>
calendar update-calendar <calendarId> --json '{...}'
calendar delete-calendar <calendarId>
calendar get-personal <calendarId>
calendar update-personal <calendarId> --json '{...}'
calendar remove-user <calendarId>
calendar update-event <calendarId> <eventId> --json '{...}'
calendar delete-event <calendarId> <eventId>
calendar default list-events [--from, --until]
calendar default get-event <eventId>
calendar default create-event --json '{...}'
calendar default update-event <eventId> --json '{...}'
calendar default delete-event <eventId>
```

**Definition of Done:**
1. API 메서드 구현 및 컴파일 성공
2. Cobra 커맨드 등록 및 `--help` 출력 확인
3. `cmd/smoke_test.go`에 등록 테스트 추가

---

### Task 1-2: Board 완성 (19개 추가)

**Files:**
- Modify: `internal/api/board.go`
- Modify: `cmd/board.go`
- Modify: `cmd/smoke_test.go`

**Dependencies:** Tasks 0-3, 0-4

**업로드 사양:** 착수 전 확정 필요 — `presigned URL | multipart/form-data | raw binary` 중 택 1, Content-Type 명시

**추가 API 메서드:**

| # | 메서드 | HTTP | 경로 |
|---|--------|------|------|
| 1 | CreateBoard | POST | `/boards` |
| 2 | UpdateBoard | PUT | `/boards/{boardId}` |
| 3 | DeleteBoard | DELETE | `/boards/{boardId}` |
| 4 | ListPostReaders | GET | `/boards/{boardId}/posts/{postId}/readers` |
| 5 | ListRecentPosts | GET | `/boards/recent/posts` |
| 6 | ListMyPosts | GET | `/boards/my/posts` |
| 7 | ListMustPosts | GET | `/boards/must/posts` |
| 8 | CreatePostAttachment | POST | `/boards/{boardId}/posts/{postId}/attachments` |
| 9 | ListPostAttachments | GET | `/boards/{boardId}/posts/{postId}/attachments` |
| 10 | GetPostAttachment | GET | `/boards/{boardId}/posts/{postId}/attachments/{attachmentId}` |
| 11 | DeletePostAttachment | DELETE | `/boards/{boardId}/posts/{postId}/attachments/{attachmentId}` |
| 12 | CreateComment | POST | `/boards/{boardId}/posts/{postId}/comments` |
| 13 | GetComment | GET | `/boards/{boardId}/posts/{postId}/comments/{commentId}` |
| 14 | UpdateComment | PUT | `/boards/{boardId}/posts/{postId}/comments/{commentId}` |
| 15 | DeleteComment | DELETE | `/boards/{boardId}/posts/{postId}/comments/{commentId}` |
| 16 | CreateCommentAttachment | POST | `/boards/{boardId}/posts/{postId}/comments/{commentId}/attachments` |
| 17 | ListCommentAttachments | GET | `/boards/{boardId}/posts/{postId}/comments/{commentId}/attachments` |
| 18 | GetCommentAttachment | GET | `/boards/{boardId}/posts/{postId}/comments/{commentId}/attachments/{id}` |
| 19 | DeleteCommentAttachment | DELETE | `/boards/{boardId}/posts/{postId}/comments/{commentId}/attachments/{id}` |

**추가 CLI 커맨드:**

```
board create --json '{...}'
board update <boardId> --json '{...}'
board delete <boardId>
board list-readers <boardId> <postId>
board list-recent
board list-my
board list-must
board create-attachment <boardId> <postId> --file <path>
board list-attachments <boardId> <postId>
board get-attachment <boardId> <postId> <attachmentId>
board delete-attachment <boardId> <postId> <attachmentId>
board create-comment <boardId> <postId> --json '{...}'
board get-comment <boardId> <postId> <commentId>
board update-comment <boardId> <postId> <commentId> --json '{...}'
board delete-comment <boardId> <postId> <commentId>
board create-comment-attachment <boardId> <postId> <commentId> --file <path>
board list-comment-attachments <boardId> <postId> <commentId>
board get-comment-attachment <boardId> <postId> <commentId> <attachmentId>
board delete-comment-attachment <boardId> <postId> <commentId> <attachmentId>
```

**Definition of Done:**
1. API 메서드 구현 및 컴파일 성공
2. Cobra 커맨드 등록 및 `--help` 출력 확인
3. `cmd/smoke_test.go`에 등록 테스트 추가

---

### Task 1-3: Mail 완성 (17개 추가)

**Files:**
- Modify: `internal/api/mail.go`
- Modify: `cmd/mail.go`
- Modify: `cmd/smoke_test.go`

**Dependencies:** Phase 0 완료

**추가 API 메서드:**

| # | 메서드 | HTTP | 경로 |
|---|--------|------|------|
| 1 | PatchMail | PATCH | `/users/{userId}/mail/{mailId}` |
| 2 | GetUnreadCount | GET | `/users/{userId}/mail/unread-count` |
| 3 | GetAttachment | GET | `/users/{userId}/mail/{mailId}/attachments/{attachmentId}` |
| 4 | ListFavoriteContactsFolders | GET | `/users/{userId}/mail/mailfolders/favorite-contacts` |
| 5 | CreateMailFolder | POST | `/users/{userId}/mail/mailfolders` |
| 6 | UpdateMailFolder | PUT | `/users/{userId}/mail/mailfolders/{folderId}` |
| 7 | DeleteMailFolder | DELETE | `/users/{userId}/mail/mailfolders/{folderId}` |
| 8 | CreateFilter | POST | `/users/{userId}/mail/filters` |
| 9 | ListFilters | GET | `/users/{userId}/mail/filters` |
| 10 | GetFilter | GET | `/users/{userId}/mail/filters/{filterId}` |
| 11 | DeleteFilter | DELETE | `/users/{userId}/mail/filters/{filterId}` |
| 12 | CreateImapMigration | POST | `/users/{userId}/mail/migration/imap` |
| 13 | GetImapMigration | GET | `/users/{userId}/mail/migration/imap` |
| 14 | DeleteImapMigration | DELETE | `/users/{userId}/mail/migration/imap` |
| 15 | CreatePop3Migration | POST | `/users/{userId}/mail/migration/pop3` |
| 16 | CreateForwarding | POST | `/users/{userId}/mail/settings/forwarding` |
| 17 | DeleteForwarding | DELETE | `/users/{userId}/mail/settings/forwarding` |

**추가 CLI 커맨드:**

```
mail update <mailId> --json '{...}'
mail unread-count
mail get-attachment <mailId> <attachmentId>
mail list-favorite-folders
mail create-folder --json '{...}'
mail update-folder <folderId> --json '{...}'
mail delete-folder <folderId>
mail filter list
mail filter get <filterId>
mail filter create --json '{...}'
mail filter delete <filterId>
mail migration create-imap --json '{...}'
mail migration get-imap
mail migration delete-imap
mail migration create-pop3 --json '{...}'
mail forwarding create --json '{...}'
mail forwarding delete
```

**Definition of Done:**
1. API 메서드 구현 및 컴파일 성공
2. Cobra 커맨드 등록 및 `--help` 출력 확인
3. `cmd/smoke_test.go`에 등록 테스트 추가

---

### Task 1-4: Task 완성 (9개 추가)

**Files:**
- Modify: `internal/api/task.go`
- Modify: `cmd/task_cmd.go`
- Modify: `cmd/smoke_test.go`

**Dependencies:** Phase 0 완료

**추가 API 메서드:**

| # | 메서드 | HTTP | 경로 |
|---|--------|------|------|
| 1 | CreateCategory | POST | `/users/{userId}/task-categories` |
| 2 | GetCategory | GET | `/users/{userId}/task-categories/{categoryId}` |
| 3 | PatchCategory | PATCH | `/users/{userId}/task-categories/{categoryId}` |
| 4 | DeleteCategory | DELETE | `/users/{userId}/task-categories/{categoryId}` |
| 5 | MoveTask | POST | `/users/{userId}/tasks/{taskId}/move` |
| 6 | CompleteTask | POST | `/tasks/{taskId}/complete` |
| 7 | IncompleteTask | POST | `/tasks/{taskId}/incomplete` |
| 8 | CompleteAssigneeTask | POST | `/tasks/{taskId}/assignees/{userId}/complete` |
| 9 | IncompleteAssigneeTask | POST | `/tasks/{taskId}/assignees/{userId}/incomplete` |

**추가 CLI 커맨드:**

```
task create-category --json '{...}'
task get-category <categoryId>
task update-category <categoryId> --json '{...}'
task delete-category <categoryId>
task move <taskId> --category <categoryId>
task complete <taskId>
task incomplete <taskId>
task complete-assignee <taskId> <userId>
task incomplete-assignee <taskId> <userId>
```

**Definition of Done:**
1. API 메서드 구현 및 컴파일 성공
2. Cobra 커맨드 등록 및 `--help` 출력 확인
3. `cmd/smoke_test.go`에 등록 테스트 추가

---

### Task 1-5: Contact 완성 (16개 추가)

**Files:**
- Modify: `internal/api/contact.go`
- Modify: `cmd/contact.go`
- Modify: `cmd/smoke_test.go`

**Dependencies:** Tasks 0-3, 0-4

**업로드 사양:** 착수 전 확정 필요 — `presigned URL | multipart/form-data | raw binary` 중 택 1, Content-Type 명시

**추가 API 메서드:**

| # | 메서드 | HTTP | 경로 |
|---|--------|------|------|
| 1 | FullUpdateContact | PUT | `/contacts/{contactId}` |
| 2 | ForceDeleteContact | DELETE | `/contacts/{contactId}/forcedelete` |
| 3 | CreatePhoto | POST | `/contacts/{contactId}/photo` |
| 4 | GetPhoto | GET | `/contacts/{contactId}/photo` |
| 5 | DeletePhoto | DELETE | `/contacts/{contactId}/photo` |
| 6 | CreateCustomProperty | POST | `/contacts/custom-properties` |
| 7 | ListCustomProperties | GET | `/contacts/custom-properties` |
| 8 | GetCustomProperty | GET | `/contacts/custom-properties/{id}` |
| 9 | PatchCustomProperty | PATCH | `/contacts/custom-properties/{id}` |
| 10 | DeleteCustomProperty | DELETE | `/contacts/custom-properties/{id}` |
| 11 | CreateTag | POST | `/contact-tags` |
| 12 | GetTag | GET | `/contact-tags/{tagId}` |
| 13 | UpdateTag | PUT | `/contact-tags/{tagId}` |
| 14 | PatchTag | PATCH | `/contact-tags/{tagId}` |
| 15 | DeleteTag | DELETE | `/contact-tags/{tagId}` |
| 16 | CreateUserTags | POST | `/users/{userId}/contact-tags` |

**추가 CLI 커맨드:**

```
contact full-update <contactId> --json '{...}'
contact force-delete <contactId>
contact upload-photo <contactId> --file <path>
contact get-photo <contactId>
contact delete-photo <contactId>
contact custom-property list
contact custom-property get <id>
contact custom-property create --json '{...}'
contact custom-property update <id> --json '{...}'
contact custom-property delete <id>
contact tag create --json '{...}'
contact tag get <tagId>
contact tag update <tagId> --json '{...}'
contact tag patch <tagId> --json '{...}'
contact tag delete <tagId>
contact tag create-user-tags --json '{...}'
```

**Definition of Done:**
1. API 메서드 구현 및 컴파일 성공
2. Cobra 커맨드 등록 및 `--help` 출력 확인
3. `cmd/smoke_test.go`에 등록 테스트 추가

---

### Task 1-6: Note 완성 (5개 추가)

**Files:**
- Modify: `internal/api/note.go`
- Modify: `cmd/note.go`
- Modify: `cmd/smoke_test.go`

**Dependencies:** Tasks 0-3, 0-4

**업로드 사양:** 착수 전 확정 필요 — `presigned URL | multipart/form-data | raw binary` 중 택 1, Content-Type 명시

**추가 API 메서드:**

| # | 메서드 | HTTP | 경로 |
|---|--------|------|------|
| 1 | PatchPost | PATCH | `/groups/{groupId}/note/posts/{postId}` |
| 2 | CreateAttachment | POST | `/groups/{groupId}/note/posts/{postId}/attachments` |
| 3 | ListAttachments | GET | `/groups/{groupId}/note/posts/{postId}/attachments` |
| 4 | GetAttachment | GET | `/groups/{groupId}/note/posts/{postId}/attachments/{id}` |
| 5 | DeleteAttachment | DELETE | `/groups/{groupId}/note/posts/{postId}/attachments/{id}` |

**추가 CLI 커맨드:**

```
note patch-post <groupId> <postId> --json '{...}'
note create-attachment <groupId> <postId> --file <path>
note list-attachments <groupId> <postId>
note get-attachment <groupId> <postId> <attachmentId>
note delete-attachment <groupId> <postId> <attachmentId>
```

**Definition of Done:**
1. API 메서드 구현 및 컴파일 성공
2. Cobra 커맨드 등록 및 `--help` 출력 확인
3. `cmd/smoke_test.go`에 등록 테스트 추가

---

### Task 1-7: Attendance 완성 (10개 추가)

**Files:**
- Modify: `internal/api/attendance.go`
- Modify: `cmd/attendance.go`
- Modify: `cmd/smoke_test.go`

**Dependencies:** Phase 0 완료

**추가 API 메서드:**

| # | 메서드 | HTTP | 경로 |
|---|--------|------|------|
| 1 | CreateTimecard | POST | `/business-support/attendance/timecards` |
| 2 | ListTimecards | GET | `/business-support/attendance/timecards` |
| 3 | GetTimecard | GET | `/business-support/attendance/timecards/{timecardId}` |
| 4 | PatchTimecard | PATCH | `/business-support/attendance/timecards/{timecardId}` |
| 5 | AdjustAnnualLeave | POST | `/business-support/attendance/annual-leaves/adjust` |
| 6 | ListAbsenceSchedules | GET | `/business-support/attendance/absence-schedule` |
| 7 | CreateAbsence | POST | `/business-support/attendance/absences` |
| 8 | GetAbsence | GET | `/business-support/attendance/absences/{absenceId}` |
| 9 | PatchAbsence | PATCH | `/business-support/attendance/absences/{absenceId}` |
| 10 | DeleteAbsence | DELETE | `/business-support/attendance/absences/{absenceId}` |

**추가 CLI 커맨드:**

```
attendance create-timecard --json '{...}'
attendance list-timecards [--cursor, --count, --all]
attendance get-timecard <timecardId>
attendance update-timecard <timecardId> --json '{...}'
attendance adjust-annual-leave --json '{...}'
attendance list-absence-schedules [--cursor, --count, --all]
attendance create-absence --json '{...}'
attendance get-absence <absenceId>
attendance update-absence <absenceId> --json '{...}'
attendance delete-absence <absenceId>
```

**Definition of Done:**
1. API 메서드 구현 및 컴파일 성공
2. Cobra 커맨드 등록 및 `--help` 출력 확인
3. `cmd/smoke_test.go`에 등록 테스트 추가

---

### Task 1-8: HR 완성 (10개 추가)

**Files:**
- Modify: `internal/api/hr.go`
- Modify: `cmd/hr.go`
- Modify: `cmd/smoke_test.go`

**Dependencies:** Phase 0 완료

**추가 API 메서드:**

| # | 메서드 | HTTP | 경로 |
|---|--------|------|------|
| 1 | CreateExtensionProperty | POST | `/business-support/human-resource/extension-properties` |
| 2 | GetExtensionProperty | GET | `/business-support/human-resource/extension-properties/{id}` |
| 3 | PatchExtensionProperty | PATCH | `/business-support/human-resource/extension-properties/{id}` |
| 4 | DeleteExtensionProperty | DELETE | `/business-support/human-resource/extension-properties/{id}` |
| 5 | GetUserExtensionProperty | GET | `/business-support/human-resource/user/{userId}/extension-properties/{id}` |
| 6 | PatchUserExtensionProperty | PATCH | `/business-support/human-resource/user/{userId}/extension-properties/{id}` |
| 7 | CreateLeaveOfAbsence | POST | `/business-support/human-resource/leave-of-absences` |
| 8 | GetLeaveOfAbsence | GET | `/business-support/human-resource/leave-of-absences/{id}` |
| 9 | PatchLeaveOfAbsence | PATCH | `/business-support/human-resource/leave-of-absences/{id}` |
| 10 | DeleteLeaveOfAbsence | DELETE | `/business-support/human-resource/leave-of-absences/{id}` |

**추가 CLI 커맨드:**

```
hr create-extension-property --json '{...}'
hr get-extension-property <id>
hr update-extension-property <id> --json '{...}'
hr delete-extension-property <id>
hr get-user-property <userId> <id>
hr update-user-property <userId> <id> --json '{...}'
hr create-leave-of-absence --json '{...}'
hr get-leave-of-absence <id>
hr update-leave-of-absence <id> --json '{...}'
hr delete-leave-of-absence <id>
```

**Definition of Done:**
1. API 메서드 구현 및 컴파일 성공
2. Cobra 커맨드 등록 및 `--help` 출력 확인
3. `cmd/smoke_test.go`에 등록 테스트 추가

---

### Task 1-9: Audit 완성 (7개 추가)

**Files:**
- Modify: `internal/api/audit.go`
- Modify: `cmd/audit.go`
- Modify: `cmd/smoke_test.go`

**Dependencies:** Phase 0 완료

**추가 API 메서드:**

| # | 메서드 | HTTP | 경로 |
|---|--------|------|------|
| 1 | CreatePolicyGroup | POST | `/audits/policy-groups` |
| 2 | GetPolicyGroup | GET | `/audits/policy-groups/{policyGroupId}` |
| 3 | UpdatePolicyGroup | PUT | `/audits/policy-groups/{policyGroupId}` |
| 4 | DeletePolicyGroup | DELETE | `/audits/policy-groups/{policyGroupId}` |
| 5 | AddPolicyGroupMembers | POST | `/audits/policy-groups/{policyGroupId}/members` |
| 6 | ListPolicyGroupMembers | GET | `/audits/policy-groups/{policyGroupId}/members` |
| 7 | RemovePolicyGroupMember | DELETE | `/audits/policy-groups/{policyGroupId}/members/{userId}` |

**추가 CLI 커맨드:**

```
audit create-policy-group --json '{...}'
audit get-policy-group <policyGroupId>
audit update-policy-group <policyGroupId> --json '{...}'
audit delete-policy-group <policyGroupId>
audit add-policy-members <policyGroupId> --json '{...}'
audit list-policy-members <policyGroupId>
audit remove-policy-member <policyGroupId> <userId>
```

**Definition of Done:**
1. API 메서드 구현 및 컴파일 성공
2. Cobra 커맨드 등록 및 `--help` 출력 확인
3. `cmd/smoke_test.go`에 등록 테스트 추가

---

### Task 1-10: Approval 완성 (18개 추가)

**Files:**
- Modify: `internal/api/approval.go`
- Modify: `cmd/approval.go`
- Modify: `cmd/smoke_test.go`

**Dependencies:** Tasks 0-3, 0-4

**업로드 사양:** 착수 전 확정 필요 — `presigned URL | multipart/form-data | raw binary` 중 택 1, Content-Type 명시

**추가 API 메서드:**

| # | 메서드 | HTTP | 경로 |
|---|--------|------|------|
| 1 | CreateCategory | POST | `/business-support/approval/categories` |
| 2 | PatchCategory | PATCH | `/business-support/approval/categories/{categoryId}` |
| 3 | DeleteCategory | DELETE | `/business-support/approval/categories/{categoryId}` |
| 4 | CreateUserDocument | POST | `/business-support/approval/users/{userId}/documents` |
| 5 | CreateImportedDocument | POST | `/business-support/approval/imported-documents` |
| 6 | CreateDocumentLink | POST | `/business-support/approval/users/{userId}/documents/create-document-link` |
| 7 | GetDocumentForm | GET | `/business-support/approval/document-forms/{documentFormId}` |
| 8 | CreateUserDocumentAttachment | POST | `/business-support/approval/users/{userId}/documents/attachments` |
| 9 | CreateImportedDocumentAttachment | POST | `/business-support/approval/imported-documents/attachments` |
| 10 | CreateLinkageCode | POST | `/business-support/approval/linkage-codes` |
| 11 | ListLinkageCodes | GET | `/business-support/approval/linkage-codes` |
| 12 | GetLinkageCode | GET | `/business-support/approval/linkage-codes/{key}` |
| 13 | PatchLinkageCode | PATCH | `/business-support/approval/linkage-codes/{key}` |
| 14 | CreateLinkageCodeItem | POST | `/business-support/approval/linkage-codes/{key}/linkage-code-items` |
| 15 | ListLinkageCodeItems | GET | `/business-support/approval/linkage-codes/{key}/linkage-code-items` |
| 16 | GetLinkageCodeItem | GET | `/business-support/approval/linkage-codes/{key}/linkage-code-items/{id}` |
| 17 | PatchLinkageCodeItem | PATCH | `/business-support/approval/linkage-codes/{key}/linkage-code-items/{id}` |
| 18 | DeleteLinkageCodeItem | DELETE | `/business-support/approval/linkage-codes/{key}/linkage-code-items/{id}` |

**추가 CLI 커맨드:**

```
approval create-category --json '{...}'
approval update-category <categoryId> --json '{...}'
approval delete-category <categoryId>
approval create-document --json '{...}'
approval create-imported-document --json '{...}'
approval create-document-link --json '{...}'
approval get-form <documentFormId>
approval upload-attachment --file <path>
approval upload-imported-attachment --file <path>
approval linkage-code list
approval linkage-code get <key>
approval linkage-code create --json '{...}'
approval linkage-code update <key> --json '{...}'
approval linkage-code-item list <key>
approval linkage-code-item get <key> <id>
approval linkage-code-item create <key> --json '{...}'
approval linkage-code-item update <key> <id> --json '{...}'
approval linkage-code-item delete <key> <id>
```

**Definition of Done:**
1. API 메서드 구현 및 컴파일 성공
2. Cobra 커맨드 등록 및 `--help` 출력 확인
3. `cmd/smoke_test.go`에 등록 테스트 추가

---

## Phase 2: Security 신규 도메인 (3개 API)

### Task 2-1: Security 도메인 추가

**Files:**
- Create: `internal/api/security.go`
- Create: `cmd/security.go`
- Modify: `cmd/smoke_test.go`

**Dependencies:** Phase 0 완료

**API 메서드:**

| # | 메서드 | HTTP | 경로 |
|---|--------|------|------|
| 1 | GetExternalBrowser | GET | `/security/external-browser` |
| 2 | EnableExternalBrowser | POST | `/security/external-browser/enable` |
| 3 | DisableExternalBrowser | POST | `/security/external-browser/disable` |

**CLI 커맨드:**

```
security get-external-browser
security enable-external-browser
security disable-external-browser
```

**Definition of Done:**
1. API 메서드 구현 및 컴파일 성공
2. Cobra 커맨드 등록 및 `--help` 출력 확인
3. `cmd/smoke_test.go`에 등록 테스트 추가

---

## Phase 3: Bot 전체 구현 (36개 추가)

### Task 3-1: Bot 관리 (7개)

**Files:**
- Modify: `internal/api/bot.go`
- Modify: `cmd/bot.go`
- Modify: `cmd/smoke_test.go`

**Dependencies:** Phase 0 완료

**추가 API:**

| # | 메서드 | HTTP | 경로 |
|---|--------|------|------|
| 1 | CreateBot | POST | `/bots` |
| 2 | ListBots | GET | `/bots` |
| 3 | GetBot | GET | `/bots/{botId}` |
| 4 | UpdateBot | PUT | `/bots/{botId}` |
| 5 | PatchBot | PATCH | `/bots/{botId}` |
| 6 | DeleteBot | DELETE | `/bots/{botId}` |
| 7 | RegenerateSecret | POST | `/bots/{botId}/secret` |

**CLI:**

```
bot list
bot get <botId>
bot create --json '{...}'
bot update <botId> --json '{...}'
bot patch <botId> --json '{...}'
bot delete <botId>
bot regenerate-secret <botId>
```

**Definition of Done:**
1. API 메서드 구현 및 컴파일 성공
2. Cobra 커맨드 등록 및 `--help` 출력 확인
3. `cmd/smoke_test.go`에 등록 테스트 추가

---

### Task 3-2: Bot 구조화 메시지 + 첨부 (4개)

**Files:**
- Modify: `internal/api/bot.go`
- Modify: `cmd/bot.go`
- Modify: `cmd/smoke_test.go`

**Dependencies:** Tasks 0-3, 0-4

**업로드 사양:** 착수 전 확정 필요 — `presigned URL | multipart/form-data | raw binary` 중 택 1, Content-Type 명시

**추가 API:**

| # | 메서드 | HTTP | 경로 |
|---|--------|------|------|
| 1 | SendMessageToUser | POST | `/bots/{botId}/users/{userId}/messages` |
| 2 | SendMessageToChannel | POST | `/bots/{botId}/channels/{channelId}/messages` |
| 3 | CreateAttachment | POST | `/bots/{botId}/attachments` |
| 4 | GetAttachmentDownloadUrl | GET | `/bots/{botId}/attachments/{fileId}` |

**CLI:**

```
bot send --to <userId> --json '{...}'    # 구조화 메시지 (기존 --text와 구분)
bot send --channel <channelId> --json '{...}'
bot create-attachment --file <path>
bot get-attachment <fileId>
```

**Definition of Done:**
1. API 메서드 구현 및 컴파일 성공
2. Cobra 커맨드 등록 및 `--help` 출력 확인
3. `cmd/smoke_test.go`에 등록 테스트 추가

---

### Task 3-3: Bot 채널 추가 (2개)

**Files:**
- Modify: `internal/api/bot.go`
- Modify: `cmd/bot.go`
- Modify: `cmd/smoke_test.go`

**Dependencies:** Phase 0 완료

**추가 API:**

| # | 메서드 | HTTP | 경로 |
|---|--------|------|------|
| 1 | CreateChannel | POST | `/bots/{botId}/channels` |
| 2 | LeaveChannel | DELETE | `/bots/{botId}/channels/{channelId}` |

**CLI:**

```
bot create-channel --json '{...}'
bot leave-channel <channelId>
```

**Definition of Done:**
1. API 메서드 구현 및 컴파일 성공
2. Cobra 커맨드 등록 및 `--help` 출력 확인
3. `cmd/smoke_test.go`에 등록 테스트 추가

---

### Task 3-4: Bot 도메인 (8개)

**Files:**
- Modify: `internal/api/bot.go`
- Modify: `cmd/bot.go`
- Modify: `cmd/smoke_test.go`

**Dependencies:** Phase 0 완료

**추가 API:**

| # | 메서드 | HTTP | 경로 |
|---|--------|------|------|
| 1 | RegisterDomain | POST | `/bots/{botId}/domains/{domainId}` |
| 2 | ListDomains | GET | `/bots/{botId}/domains` |
| 3 | UpdateDomain | PUT | `/bots/{botId}/domains/{domainId}` |
| 4 | PatchDomain | PATCH | `/bots/{botId}/domains/{domainId}` |
| 5 | DeleteDomain | DELETE | `/bots/{botId}/domains/{domainId}` |
| 6 | AddDomainMembers | POST | `/bots/{botId}/domains/{domainId}/members` |
| 7 | ListDomainMembers | GET | `/bots/{botId}/domains/{domainId}/members` |
| 8 | RemoveDomainMember | DELETE | `/bots/{botId}/domains/{domainId}/members/{userId}` |

**CLI:**

```
bot domain register <domainId> --json '{...}'
bot domain list
bot domain update <domainId> --json '{...}'
bot domain patch <domainId> --json '{...}'
bot domain delete <domainId>
bot domain add-members <domainId> --json '{...}'
bot domain list-members <domainId>
bot domain remove-member <domainId> <userId>
```

**Definition of Done:**
1. API 메서드 구현 및 컴파일 성공
2. Cobra 커맨드 등록 및 `--help` 출력 확인
3. `cmd/smoke_test.go`에 등록 테스트 추가

---

### Task 3-5: Bot 고정메뉴 (3개)

**Files:**
- Modify: `internal/api/bot.go`
- Modify: `cmd/bot.go`
- Modify: `cmd/smoke_test.go`

**Dependencies:** Phase 0 완료

**추가 API:**

| # | 메서드 | HTTP | 경로 |
|---|--------|------|------|
| 1 | UpsertPersistentMenu | POST | `/bots/{botId}/persistentmenu` |
| 2 | GetPersistentMenu | GET | `/bots/{botId}/persistentmenu` |
| 3 | DeletePersistentMenu | DELETE | `/bots/{botId}/persistentmenu` |

**CLI:**

```
bot persistent-menu set --json '{...}'
bot persistent-menu get
bot persistent-menu delete
```

**Definition of Done:**
1. API 메서드 구현 및 컴파일 성공
2. Cobra 커맨드 등록 및 `--help` 출력 확인
3. `cmd/smoke_test.go`에 등록 테스트 추가

---

### Task 3-6: Bot 리치메뉴 (12개)

**Files:**
- Modify: `internal/api/bot.go`
- Modify: `cmd/bot.go`
- Modify: `cmd/smoke_test.go`

**Dependencies:** Tasks 0-3, 0-4

**업로드 사양:** 착수 전 확정 필요 — `presigned URL | multipart/form-data | raw binary` 중 택 1, Content-Type 명시

**추가 API:**

| # | 메서드 | HTTP | 경로 |
|---|--------|------|------|
| 1 | CreateRichMenu | POST | `/bots/{botId}/richmenus` |
| 2 | ListRichMenus | GET | `/bots/{botId}/richmenus` |
| 3 | GetRichMenu | GET | `/bots/{botId}/richmenus/{richmenuId}` |
| 4 | DeleteRichMenu | DELETE | `/bots/{botId}/richmenus/{richmenuId}` |
| 5 | SetRichMenuImage | POST | `/bots/{botId}/richmenus/{richmenuId}/image` |
| 6 | GetRichMenuImage | GET | `/bots/{botId}/richmenus/{richmenuId}/image` |
| 7 | SetUserRichMenu | POST | `/bots/{botId}/richmenus/{richmenuId}/users/{userId}` |
| 8 | GetUserRichMenu | GET | `/bots/{botId}/richmenus/users/{userId}` |
| 9 | DeleteUserRichMenu | DELETE | `/bots/{botId}/richmenus/users/{userId}` |
| 10 | SetDefaultRichMenu | POST | `/bots/{botId}/richmenus/{richmenuId}/set-default` |
| 11 | GetDefaultRichMenu | GET | `/bots/{botId}/richmenus/default` |
| 12 | DeleteDefaultRichMenu | DELETE | `/bots/{botId}/richmenus/default` |

**CLI:**

```
bot richmenu create --json '{...}'
bot richmenu list
bot richmenu get <richmenuId>
bot richmenu delete <richmenuId>
bot richmenu set-image <richmenuId> --file <path>
bot richmenu get-image <richmenuId>
bot richmenu set-user <richmenuId> <userId>
bot richmenu get-user <userId>
bot richmenu delete-user <userId>
bot richmenu set-default <richmenuId>
bot richmenu get-default
bot richmenu delete-default
```

**Definition of Done:**
1. API 메서드 구현 및 컴파일 성공
2. Cobra 커맨드 등록 및 `--help` 출력 확인
3. `cmd/smoke_test.go`에 등록 테스트 추가

---

## Phase 4: Directory 전체 구현 (121개 추가)

가장 큰 도메인. 서브도메인별로 태스크를 나눈다.

### Phase 4 실행 규칙

- Phase 4 태스크는 기능 의존성은 낮지만 `internal/api/directory.go`, `cmd/directory.go`, `cmd/smoke_test.go`를 공유하므로 완전 병렬 실행하지 않는다
- 기본 실행 순서: `4-1 → 4-2 → 4-3 → 4-4 → 4-5 → 4-6 → 4-7 → 4-8 → 4-9 → 4-10 → 4-11 → 4-12`
- 병렬이 필요하면 선행 리팩터링으로 `internal/api/directory_*.go`, `cmd/directory_*.go`로 파일을 분리한 뒤 후속 Task를 병렬 진행한다

### Task 4-1: User CUD (12개)

**Files:**
- Modify: `internal/api/directory.go`
- Modify: `cmd/directory.go`
- Modify: `cmd/smoke_test.go`

**Dependencies:** Phase 0 완료

**추가 API:**

| # | 메서드 | HTTP | 경로 |
|---|--------|------|------|
| 1 | CreateUser | POST | `/users` |
| 2 | UpdateUser | PUT | `/users/{userId}` |
| 3 | PatchUser | PATCH | `/users/{userId}` |
| 4 | DeleteUser | DELETE | `/users/{userId}` |
| 5 | ForceDeleteUser | DELETE | `/users/{userId}/forcedelete` |
| 6 | UndeleteUser | POST | `/users/{userId}/undelete` |
| 7 | SuspendUser | POST | `/users/{userId}/suspend` |
| 8 | UnsuspendUser | POST | `/users/{userId}/unsuspend` |
| 9 | ForceLogoutUser | POST | `/users/{userId}/force-logout` |
| 10 | MoveUser | POST | `/users/{userId}/move` |
| 11 | SetLeaveOfAbsence | POST | `/users/{userId}/set-leave-of-absence` |
| 12 | ClearLeaveOfAbsence | POST | `/users/{userId}/clear-leave-of-absence` |

**CLI:**

```
directory create-user --json '{...}'
directory update-user <userId> --json '{...}'
directory patch-user <userId> --json '{...}'
directory delete-user <userId>
directory force-delete-user <userId>
directory undelete-user <userId>
directory suspend-user <userId>
directory unsuspend-user <userId>
directory force-logout-user <userId>
directory move-user <userId> --json '{...}'
directory set-leave <userId> --json '{...}'
directory clear-leave <userId>
```

**Definition of Done:**
1. API 메서드 구현 및 컴파일 성공
2. Cobra 커맨드 등록 및 `--help` 출력 확인
3. `cmd/smoke_test.go`에 등록 테스트 추가

---

### Task 4-2: User Profile (9개)

**Files:**
- Modify: `internal/api/directory.go`
- Modify: `cmd/directory.go`
- Modify: `cmd/smoke_test.go`

**Dependencies:** Tasks 0-3, 0-4

**업로드 사양:** 착수 전 확정 필요 — `presigned URL | multipart/form-data | raw binary` 중 택 1, Content-Type 명시

**추가 API:**

| # | 메서드 | HTTP | 경로 |
|---|--------|------|------|
| 1 | CreateUserPhoto | POST | `/users/{userId}/photo` |
| 2 | GetUserPhoto | GET | `/users/{userId}/photo` |
| 3 | DeleteUserPhoto | DELETE | `/users/{userId}/photo` |
| 4 | CreateProfileStatus | POST | `/users/{userId}/user-profile-statuses` |
| 5 | ListProfileStatuses | GET | `/users/{userId}/user-profile-statuses` |
| 6 | GetProfileStatus | GET | `/users/{userId}/user-profile-statuses/{id}` |
| 7 | UpdateProfileStatus | PUT | `/users/{userId}/user-profile-statuses/{id}` |
| 8 | PatchProfileStatus | PATCH | `/users/{userId}/user-profile-statuses/{id}` |
| 9 | DeleteProfileStatus | DELETE | `/users/{userId}/user-profile-statuses/{id}` |

**CLI:**

```
directory upload-photo <userId> --file <path>
directory get-photo <userId>
directory delete-photo <userId>
directory profile-status list <userId>
directory profile-status get <userId> <id>
directory profile-status create <userId> --json '{...}'
directory profile-status update <userId> <id> --json '{...}'
directory profile-status patch <userId> <id> --json '{...}'
directory profile-status delete <userId> <id>
```

**Definition of Done:**
1. API 메서드 구현 및 컴파일 성공
2. Cobra 커맨드 등록 및 `--help` 출력 확인
3. `cmd/smoke_test.go`에 등록 테스트 추가

---

### Task 4-3: User Email + Invitations + Links (12개)

**Files:**
- Modify: `internal/api/directory.go`
- Modify: `cmd/directory.go`
- Modify: `cmd/smoke_test.go`

**Dependencies:** Phase 0 완료

**추가 API:**

| # | 메서드 | HTTP | 경로 |
|---|--------|------|------|
| 1 | AddAliasEmail | POST | `/users/{userId}/alias-emails/{email}` |
| 2 | DeleteAliasEmail | DELETE | `/users/{userId}/alias-emails/{email}` |
| 3 | SendInvitationEmail | POST | `/users/{userId}/send-invitation-email` |
| 4 | SendInvitationEmailToAll | POST | `/users/send-invitation-email` |
| 5 | LinkAllUsersToWorks | POST | `/users/link-to-works` |
| 6 | LinkUserToWorks | POST | `/users/{userId}/link-to-works` |
| 7 | UnlinkUserToWorks | POST | `/users/{userId}/unlink-to-works` |
| 8 | LinkAllUsersToLine | POST | `/users/link-to-line` |
| 9 | LinkUserToLine | POST | `/users/{userId}/link-to-line` |
| 10 | UnlinkUserToLine | POST | `/users/{userId}/unlink-to-line` |
| 11 | GetLinkUrl | GET | `/users/{userId}/link-url` |
| 12 | ResetLinkUrl | POST | `/users/{userId}/link-url/reset` |

**CLI:**

```
directory add-alias-email <userId> <email>
directory delete-alias-email <userId> <email>
directory send-invitation <userId>
directory send-invitation-all
directory link-to-works <userId>
directory link-all-to-works
directory unlink-to-works <userId>
directory link-to-line <userId>
directory link-all-to-line
directory unlink-to-line <userId>
directory get-link-url <userId>
directory reset-link-url <userId>
```

**Definition of Done:**
1. API 메서드 구현 및 컴파일 성공
2. Cobra 커맨드 등록 및 `--help` 출력 확인
3. `cmd/smoke_test.go`에 등록 테스트 추가

---

### Task 4-4: User External Keys + Custom Properties (7개)

**Files:**
- Modify: `internal/api/directory.go`
- Modify: `cmd/directory.go`
- Modify: `cmd/smoke_test.go`

**Dependencies:** Phase 0 완료

**추가 API:**

| # | 메서드 | HTTP | 경로 |
|---|--------|------|------|
| 1 | UpsertUserExternalKeys | POST | `/users/external-keys` |
| 2 | ListUserExternalKeys | GET | `/users/external-keys` |
| 3 | CreateUserCustomProperty | POST | `/directory/users/custom-properties` |
| 4 | ListUserCustomProperties | GET | `/directory/users/custom-properties` |
| 5 | GetUserCustomProperty | GET | `/directory/users/custom-properties/{id}` |
| 6 | PatchUserCustomProperty | PATCH | `/directory/users/custom-properties/{id}` |
| 7 | DeleteUserCustomProperty | DELETE | `/directory/users/custom-properties/{id}` |

**CLI:**

```
directory upsert-external-keys --json '{...}'    # users
directory list-external-keys                     # users
directory user-custom-property list
directory user-custom-property get <id>
directory user-custom-property create --json '{...}'
directory user-custom-property update <id> --json '{...}'
directory user-custom-property delete <id>
```

**Definition of Done:**
1. API 메서드 구현 및 컴파일 성공
2. Cobra 커맨드 등록 및 `--help` 출력 확인
3. `cmd/smoke_test.go`에 등록 테스트 추가

---

### Task 4-5: Group CUD + Members + Admins + External Keys (12개)

**Files:**
- Modify: `internal/api/directory.go`
- Modify: `cmd/directory.go`
- Modify: `cmd/smoke_test.go`

**Dependencies:** Phase 0 완료

**추가 API:**

| # | 메서드 | HTTP | 경로 |
|---|--------|------|------|
| 1 | CreateGroup | POST | `/groups` |
| 2 | UpdateGroup | PUT | `/groups/{groupId}` |
| 3 | PatchGroup | PATCH | `/groups/{groupId}` |
| 4 | DeleteGroup | DELETE | `/groups/{groupId}` |
| 5 | ListGroupMembers | GET | `/groups/{groupId}/members` |
| 6 | AddGroupMembers | POST | `/groups/{groupId}/members` |
| 7 | RemoveGroupMember | DELETE | `/groups/{groupId}/members/{id}` |
| 8 | ListGroupAdministrators | GET | `/groups/{groupId}/administrators` |
| 9 | AddGroupAdministrator | POST | `/groups/{groupId}/administrators` |
| 10 | RemoveGroupAdministrator | DELETE | `/groups/{groupId}/administrators/{userId}` |
| 11 | UpsertGroupExternalKeys | POST | `/groups/external-keys` |
| 12 | ListGroupExternalKeys | GET | `/groups/external-keys` |

**CLI:**

```
directory create-group --json '{...}'
directory update-group <groupId> --json '{...}'
directory patch-group <groupId> --json '{...}'
directory delete-group <groupId>
directory list-group-members <groupId>
directory add-group-members <groupId> --json '{...}'
directory remove-group-member <groupId> <memberId>
directory list-group-admins <groupId>
directory add-group-admin <groupId> --json '{...}'
directory remove-group-admin <groupId> <userId>
directory upsert-group-external-keys --json '{...}'
directory list-group-external-keys
```

**Definition of Done:**
1. API 메서드 구현 및 컴파일 성공
2. Cobra 커맨드 등록 및 `--help` 출력 확인
3. `cmd/smoke_test.go`에 등록 테스트 추가

---

### Task 4-6: OrgUnit CUD + Members + AccessRestrict + External Keys (12개)

**Files:**
- Modify: `internal/api/directory.go`
- Modify: `cmd/directory.go`
- Modify: `cmd/smoke_test.go`

**Dependencies:** Phase 0 완료

**추가 API:**

| # | 메서드 | HTTP | 경로 |
|---|--------|------|------|
| 1 | CreateOrgUnit | POST | `/orgunits` |
| 2 | UpdateOrgUnit | PUT | `/orgunits/{orgUnitId}` |
| 3 | PatchOrgUnit | PATCH | `/orgunits/{orgUnitId}` |
| 4 | DeleteOrgUnit | DELETE | `/orgunits/{orgUnitId}` |
| 5 | MoveOrgUnit | POST | `/orgunits/{orgUnitId}/move` |
| 6 | ListOrgUnitMembers | GET | `/orgunits/{orgUnitId}/members` |
| 7 | CreateOrgUnitAccessRestrict | POST | `/orgunits/{orgUnitId}/orgunit-access-restrict` |
| 8 | GetOrgUnitAccessRestrict | GET | `/orgunits/{orgUnitId}/orgunit-access-restrict` |
| 9 | UpdateOrgUnitAccessRestrict | PUT | `/orgunits/{orgUnitId}/orgunit-access-restrict` |
| 10 | DeleteOrgUnitAccessRestrict | DELETE | `/orgunits/{orgUnitId}/orgunit-access-restrict` |
| 11 | UpsertOrgUnitExternalKeys | POST | `/orgunits/external-keys` |
| 12 | ListOrgUnitExternalKeys | GET | `/orgunits/external-keys` |

**CLI:**

```
directory create-orgunit --json '{...}'
directory update-orgunit <orgUnitId> --json '{...}'
directory patch-orgunit <orgUnitId> --json '{...}'
directory delete-orgunit <orgUnitId>
directory move-orgunit <orgUnitId> --json '{...}'
directory list-orgunit-members <orgUnitId>
directory orgunit-access-restrict create <orgUnitId> --json '{...}'
directory orgunit-access-restrict get <orgUnitId>
directory orgunit-access-restrict update <orgUnitId> --json '{...}'
directory orgunit-access-restrict delete <orgUnitId>
directory upsert-orgunit-external-keys --json '{...}'
directory list-orgunit-external-keys
```

**Definition of Done:**
1. API 메서드 구현 및 컴파일 성공
2. Cobra 커맨드 등록 및 `--help` 출력 확인
3. `cmd/smoke_test.go`에 등록 테스트 추가

---

### Task 4-7: Positions CRUD + External Keys (9개)

**Files:**
- Modify: `internal/api/directory.go`
- Modify: `cmd/directory.go`
- Modify: `cmd/smoke_test.go`

**Dependencies:** Phase 0 완료

**추가 API:**

| # | 메서드 | HTTP | 경로 |
|---|--------|------|------|
| 1 | CreatePosition | POST | `/directory/positions` |
| 2 | GetPosition | GET | `/directory/positions/{positionId}` |
| 3 | UpdatePosition | PUT | `/directory/positions/{positionId}` |
| 4 | PatchPosition | PATCH | `/directory/positions/{positionId}` |
| 5 | DeletePosition | DELETE | `/directory/positions/{positionId}` |
| 6 | EnablePositions | POST | `/directory/positions/enable` |
| 7 | DisablePositions | POST | `/directory/positions/disable` |
| 8 | UpsertPositionExternalKeys | POST | `/directory/positions/external-keys` |
| 9 | ListPositionExternalKeys | GET | `/directory/positions/external-keys` |

**CLI:**

```
directory get-position <positionId>
directory create-position --json '{...}'
directory update-position <positionId> --json '{...}'
directory patch-position <positionId> --json '{...}'
directory delete-position <positionId>
directory enable-positions
directory disable-positions
directory upsert-position-external-keys --json '{...}'
directory list-position-external-keys
```

**Definition of Done:**
1. API 메서드 구현 및 컴파일 성공
2. Cobra 커맨드 등록 및 `--help` 출력 확인
3. `cmd/smoke_test.go`에 등록 테스트 추가

---

### Task 4-8: Levels CRUD + External Keys (9개)

**Files:**
- Modify: `internal/api/directory.go`
- Modify: `cmd/directory.go`
- Modify: `cmd/smoke_test.go`

**Dependencies:** 직전 Directory Task 완료 또는 공유 파일 분리 선행 PR 완료

**추가 API:**

| # | 메서드 | HTTP | 경로 |
|---|--------|------|------|
| 1 | CreateLevel | POST | `/directory/levels` |
| 2 | GetLevel | GET | `/directory/levels/{levelId}` |
| 3 | UpdateLevel | PUT | `/directory/levels/{levelId}` |
| 4 | PatchLevel | PATCH | `/directory/levels/{levelId}` |
| 5 | DeleteLevel | DELETE | `/directory/levels/{levelId}` |
| 6 | EnableLevels | POST | `/directory/levels/enable` |
| 7 | DisableLevels | POST | `/directory/levels/disable` |
| 8 | UpsertLevelExternalKeys | POST | `/directory/levels/external-keys` |
| 9 | ListLevelExternalKeys | GET | `/directory/levels/external-keys` |

**CLI:**

```
directory get-level <levelId>
directory create-level --json '{...}'
directory update-level <levelId> --json '{...}'
directory patch-level <levelId> --json '{...}'
directory delete-level <levelId>
directory enable-levels
directory disable-levels
directory upsert-level-external-keys --json '{...}'
directory list-level-external-keys
```

**Definition of Done:**
1. API 메서드 구현 및 컴파일 성공
2. Cobra 커맨드 등록 및 `--help` 출력 확인
3. `cmd/smoke_test.go`에 등록 테스트 추가

---

### Task 4-9: Employment Types CRUD + External Keys + Access Restrict (13개)

**Files:**
- Modify: `internal/api/directory.go`
- Modify: `cmd/directory.go`
- Modify: `cmd/smoke_test.go`

**Dependencies:** Phase 0 완료

**추가 API:**

| # | 메서드 | HTTP | 경로 |
|---|--------|------|------|
| 1 | CreateEmploymentType | POST | `/directory/employment-types` |
| 2 | GetEmploymentType | GET | `/directory/employment-types/{id}` |
| 3 | UpdateEmploymentType | PUT | `/directory/employment-types/{id}` |
| 4 | PatchEmploymentType | PATCH | `/directory/employment-types/{id}` |
| 5 | DeleteEmploymentType | DELETE | `/directory/employment-types/{id}` |
| 6 | EnableEmploymentTypes | POST | `/directory/employment-types/enable` |
| 7 | DisableEmploymentTypes | POST | `/directory/employment-types/disable` |
| 8 | UpsertEmploymentTypeExternalKeys | POST | `/directory/employment-types/external-keys` |
| 9 | ListEmploymentTypeExternalKeys | GET | `/directory/employment-types/external-keys` |
| 10 | CreateEmploymentTypeAccessRestrict | POST | `/directory/employment-types/{id}/orgunit-access-restrict` |
| 11 | GetEmploymentTypeAccessRestrict | GET | `/directory/employment-types/{id}/orgunit-access-restrict` |
| 12 | UpdateEmploymentTypeAccessRestrict | PUT | `/directory/employment-types/{id}/orgunit-access-restrict` |
| 13 | DeleteEmploymentTypeAccessRestrict | DELETE | `/directory/employment-types/{id}/orgunit-access-restrict` |

**CLI:**

```
directory get-employment-type <id>
directory create-employment-type --json '{...}'
directory update-employment-type <id> --json '{...}'
directory patch-employment-type <id> --json '{...}'
directory delete-employment-type <id>
directory enable-employment-types
directory disable-employment-types
directory upsert-employment-type-external-keys --json '{...}'
directory list-employment-type-external-keys
directory employment-type-access-restrict create <id> --json '{...}'
directory employment-type-access-restrict get <id>
directory employment-type-access-restrict update <id> --json '{...}'
directory employment-type-access-restrict delete <id>
```

**Definition of Done:**
1. API 메서드 구현 및 컴파일 성공
2. Cobra 커맨드 등록 및 `--help` 출력 확인
3. `cmd/smoke_test.go`에 등록 테스트 추가

---

### Task 4-10: User Types CRUD + External Keys + Access Restrict (13개)

**Files:**
- Modify: `internal/api/directory.go`
- Modify: `cmd/directory.go`
- Modify: `cmd/smoke_test.go`

**Dependencies:** 직전 Directory Task 완료 또는 공유 파일 분리 선행 PR 완료

**추가 API:**

| # | 메서드 | HTTP | 경로 |
|---|--------|------|------|
| 1 | CreateUserType | POST | `/directory/user-types` |
| 2 | GetUserType | GET | `/directory/user-types/{id}` |
| 3 | UpdateUserType | PUT | `/directory/user-types/{id}` |
| 4 | PatchUserType | PATCH | `/directory/user-types/{id}` |
| 5 | DeleteUserType | DELETE | `/directory/user-types/{id}` |
| 6 | EnableUserTypes | POST | `/directory/user-types/enable` |
| 7 | DisableUserTypes | POST | `/directory/user-types/disable` |
| 8 | UpsertUserTypeExternalKeys | POST | `/directory/user-types/external-keys` |
| 9 | ListUserTypeExternalKeys | GET | `/directory/user-types/external-keys` |
| 10 | CreateUserTypeAccessRestrict | POST | `/directory/user-types/{id}/orgunit-access-restrict` |
| 11 | GetUserTypeAccessRestrict | GET | `/directory/user-types/{id}/orgunit-access-restrict` |
| 12 | UpdateUserTypeAccessRestrict | PUT | `/directory/user-types/{id}/orgunit-access-restrict` |
| 13 | DeleteUserTypeAccessRestrict | DELETE | `/directory/user-types/{id}/orgunit-access-restrict` |

**CLI:**

```
directory get-user-type <id>
directory create-user-type --json '{...}'
directory update-user-type <id> --json '{...}'
directory patch-user-type <id> --json '{...}'
directory delete-user-type <id>
directory enable-user-types
directory disable-user-types
directory upsert-user-type-external-keys --json '{...}'
directory list-user-type-external-keys
directory user-type-access-restrict create <id> --json '{...}'
directory user-type-access-restrict get <id>
directory user-type-access-restrict update <id> --json '{...}'
directory user-type-access-restrict delete <id>
```

**Definition of Done:**
1. API 메서드 구현 및 컴파일 성공
2. Cobra 커맨드 등록 및 `--help` 출력 확인
3. `cmd/smoke_test.go`에 등록 테스트 추가

---

### Task 4-11: Profile Statuses CRUD (8개)

**Files:**
- Modify: `internal/api/directory.go`
- Modify: `cmd/directory.go`
- Modify: `cmd/smoke_test.go`

**Dependencies:** 직전 Directory Task 완료 또는 공유 파일 분리 선행 PR 완료

**추가 API:**

| # | 메서드 | HTTP | 경로 |
|---|--------|------|------|
| 1 | CreateDirectoryProfileStatus | POST | `/directory/profile-statuses` |
| 2 | ListDirectoryProfileStatuses | GET | `/directory/profile-statuses` |
| 3 | GetDirectoryProfileStatus | GET | `/directory/profile-statuses/{id}` |
| 4 | UpdateDirectoryProfileStatus | PUT | `/directory/profile-statuses/{id}` |
| 5 | PatchDirectoryProfileStatus | PATCH | `/directory/profile-statuses/{id}` |
| 6 | DeleteDirectoryProfileStatus | DELETE | `/directory/profile-statuses/{id}` |
| 7 | EnableDirectoryProfileStatuses | POST | `/directory/profile-statuses/enable` |
| 8 | DisableDirectoryProfileStatuses | POST | `/directory/profile-statuses/disable` |

**CLI:**

```
directory profile-status-def list
directory profile-status-def get <id>
directory profile-status-def create --json '{...}'
directory profile-status-def update <id> --json '{...}'
directory profile-status-def patch <id> --json '{...}'
directory profile-status-def delete <id>
directory profile-status-def enable
directory profile-status-def disable
```

**Definition of Done:**
1. API 메서드 구현 및 컴파일 성공
2. Cobra 커맨드 등록 및 `--help` 출력 확인
3. `cmd/smoke_test.go`에 등록 테스트 추가

---

### Task 4-12: Custom Fields CRUD (5개)

**Files:**
- Modify: `internal/api/directory.go`
- Modify: `cmd/directory.go`
- Modify: `cmd/smoke_test.go`

**Dependencies:** 직전 Directory Task 완료 또는 공유 파일 분리 선행 PR 완료

**추가 API:**

| # | 메서드 | HTTP | 경로 |
|---|--------|------|------|
| 1 | CreateCustomField | POST | `/directory/custom-fields` |
| 2 | ListCustomFields | GET | `/directory/custom-fields` |
| 3 | GetCustomField | GET | `/directory/custom-fields/{id}` |
| 4 | PatchCustomField | PATCH | `/directory/custom-fields/{id}` |
| 5 | DeleteCustomField | DELETE | `/directory/custom-fields/{id}` |

**CLI:**

```
directory custom-field list
directory custom-field get <id>
directory custom-field create --json '{...}'
directory custom-field update <id> --json '{...}'
directory custom-field delete <id>
```

**Definition of Done:**
1. API 메서드 구현 및 컴파일 성공
2. Cobra 커맨드 등록 및 `--help` 출력 확인
3. `cmd/smoke_test.go`에 등록 테스트 추가

---

## Phase 5: Drive 전체 구현 (125개 추가)

Drive는 4개 드라이브 타입에 같은 패턴이 반복됨. 공통 인터페이스로 코드 중복을 최소화한다.

### Drive 공통 기능 매트릭스

| 기능 | MyDrive | GroupFolder | SharedDrive | SharedFolder |
|------|---------|-------------|-------------|-------------|
| 관리/파일보완 | N/A | Task 5-4 | Task 5-8 | Task 5-12 |
| 파일조작 (7) | Task 5-1 | Task 5-5 | Task 5-9 | Task 5-13 |
| 리비전 (4) | Task 5-2 | Task 5-6 | Task 5-10 | Task 5-13 |
| 휴지통 (1-3) | Task 5-3 | Task 5-6 | Task 5-10 | N/A |
| 링크/공유 (5-11) | Task 5-3 | Task 5-7 | Task 5-11 | Task 5-14 |
| 권한 (6-21) | N/A | Task 5-7 | Task 5-11 | N/A |

### Task 5-1: MyDrive 파일조작 (7개)

**Files:**
- Modify: `internal/api/drive.go`
- Modify: `cmd/drive.go`
- Modify: `cmd/smoke_test.go`

**Dependencies:** Phase 0 완료

**추가 API:**

| # | 메서드 | HTTP | 경로 |
|---|--------|------|------|
| 1 | CopyFile | POST | `/users/{userId}/drive/files/{fileId}/copy` |
| 2 | RenameFile | POST | `/users/{userId}/drive/files/{fileId}/rename` |
| 3 | MoveFile | POST | `/users/{userId}/drive/files/{fileId}/move` |
| 4 | ProtectFile | POST | `/users/{userId}/drive/files/{fileId}/protect` |
| 5 | UnprotectFile | POST | `/users/{userId}/drive/files/{fileId}/unprotect` |
| 6 | LockFile | POST | `/users/{userId}/drive/files/{fileId}/lock` |
| 7 | UnlockFile | POST | `/users/{userId}/drive/files/{fileId}/unlock` |

**CLI:**

```
drive copy <fileId> --json '{...}'
drive rename <fileId> --json '{...}'
drive move <fileId> --json '{...}'
drive protect <fileId>
drive unprotect <fileId>
drive lock <fileId>
drive unlock <fileId>
```

**Definition of Done:**
1. API 메서드 구현 및 컴파일 성공
2. Cobra 커맨드 등록 및 `--help` 출력 확인
3. `cmd/smoke_test.go`에 등록 테스트 추가

---

### Task 5-2: MyDrive 리비전 (4개)

**Files:**
- Modify: `internal/api/drive.go`
- Modify: `cmd/drive.go`
- Modify: `cmd/smoke_test.go`

**Dependencies:** Phase 0 완료

**추가 API:**

| # | 메서드 | HTTP | 경로 |
|---|--------|------|------|
| 1 | ListRevisions | GET | `/users/{userId}/drive/files/{fileId}/revisions` |
| 2 | GetRevision | GET | `/users/{userId}/drive/files/{fileId}/revisions/{revisionId}` |
| 3 | RestoreRevision | POST | `/users/{userId}/drive/files/{fileId}/revisions/{revisionId}/restore` |
| 4 | GetRevisionDownloadUrl | GET | `/users/{userId}/drive/files/{fileId}/revisions/{revisionId}/download` |

**CLI:**

```
drive revision list <fileId>
drive revision get <fileId> <revisionId>
drive revision restore <fileId> <revisionId>
drive revision download <fileId> <revisionId>
```

**Definition of Done:**
1. API 메서드 구현 및 컴파일 성공
2. Cobra 커맨드 등록 및 `--help` 출력 확인
3. `cmd/smoke_test.go`에 등록 테스트 추가

---

### Task 5-3: MyDrive 휴지통 보완 + 링크/공유 (11개)

**Files:**
- Modify: `internal/api/drive.go`
- Modify: `cmd/drive.go`
- Modify: `cmd/smoke_test.go`

**Dependencies:** Phase 0 완료

**추가 API:**

| # | 메서드 | HTTP | 경로 |
|---|--------|------|------|
| 1 | DeleteTrashFile | DELETE | `/users/{userId}/drive/trash-files/{fileId}` |
| 2 | GetLinkSetting | GET | `/users/{userId}/drive/link-setting` |
| 3 | GetLink | GET | `/users/{userId}/drive/files/{fileId}/link` |
| 4 | CreateLink | POST | `/users/{userId}/drive/files/{fileId}/link` |
| 5 | PatchLink | PATCH | `/users/{userId}/drive/files/{fileId}/link` |
| 6 | DeleteLink | DELETE | `/users/{userId}/drive/files/{fileId}/link` |
| 7 | GetShare | GET | `/users/{userId}/drive/files/{fileId}/share` |
| 8 | CreateShare | POST | `/users/{userId}/drive/files/{fileId}/share` |
| 9 | PatchShare | PATCH | `/users/{userId}/drive/files/{fileId}/share` |
| 10 | DeleteShare | DELETE | `/users/{userId}/drive/files/{fileId}/share` |
| 11 | ListShareSubFolders | GET | `/users/{userId}/drive/files/{fileId}/share-sub-folders` |

**CLI:**

```
drive trash-delete <fileId>
drive link-setting
drive link get <fileId>
drive link create <fileId> --json '{...}'
drive link update <fileId> --json '{...}'
drive link delete <fileId>
drive share get <fileId>
drive share create <fileId> --json '{...}'
drive share update <fileId> --json '{...}'
drive share delete <fileId>
drive share list-sub-folders <fileId>
```

**Definition of Done:**
1. API 메서드 구현 및 컴파일 성공
2. Cobra 커맨드 등록 및 `--help` 출력 확인
3. `cmd/smoke_test.go`에 등록 테스트 추가

---

### Task 5-4: GroupFolder 관리 + 파일 보완 (7개)

**Files:**
- Modify: `internal/api/drive_group.go`
- Modify: `cmd/drive.go`
- Modify: `cmd/smoke_test.go`

**Dependencies:** Task 0-3 (파일 업로드 헬퍼)

**추가 API:**

| # | 메서드 | HTTP | 경로 |
|---|--------|------|------|
| 1 | CreateGroupFolder | POST | `/groups/{groupId}/folder` |
| 2 | DeleteGroupFolder | DELETE | `/groups/{groupId}/folder` |
| 3 | CreateGroupFolderInRoot | POST | `/groups/{groupId}/folder/files/createfolder` |
| 4 | CreateGroupSubFolder | POST | `/groups/{groupId}/folder/files/{fileId}/createfolder` |
| 5 | DeleteGroupFile | DELETE | `/groups/{groupId}/folder/files/{fileId}` |
| 6 | CreateGroupUploadUrl | POST | `/groups/{groupId}/folder/files/{fileId}` |
| 7 | GetGroupDownloadUrl | GET | `/groups/{groupId}/folder/files/{fileId}/download` |

**CLI:**

```
drive group create-folder <groupId>
drive group delete-folder <groupId>
drive group mkdir <groupId> --json '{...}'
drive group mkdir <groupId> --parent <fileId> --json '{...}'
drive group delete <groupId> <fileId>
drive group upload <groupId> <fileId> --file <path>    # 업로드 방식: presigned URL
drive group download <groupId> <fileId>
```

**Definition of Done:**
1. API 메서드 구현 및 컴파일 성공
2. Cobra 커맨드 등록 및 `--help` 출력 확인
3. `cmd/smoke_test.go`에 등록 테스트 추가

---

### Task 5-5: GroupFolder 파일조작 (7개)

**Files:**
- Modify: `internal/api/drive_group.go`
- Modify: `cmd/drive.go`
- Modify: `cmd/smoke_test.go`

**Dependencies:** Task 5-4

**추가 API:**

| # | 메서드 | HTTP | 경로 |
|---|--------|------|------|
| 1 | CopyGroupFile | POST | `/groups/{groupId}/folder/files/{fileId}/copy` |
| 2 | RenameGroupFile | POST | `/groups/{groupId}/folder/files/{fileId}/rename` |
| 3 | MoveGroupFile | POST | `/groups/{groupId}/folder/files/{fileId}/move` |
| 4 | ProtectGroupFile | POST | `/groups/{groupId}/folder/files/{fileId}/protect` |
| 5 | UnprotectGroupFile | POST | `/groups/{groupId}/folder/files/{fileId}/unprotect` |
| 6 | LockGroupFile | POST | `/groups/{groupId}/folder/files/{fileId}/lock` |
| 7 | UnlockGroupFile | POST | `/groups/{groupId}/folder/files/{fileId}/unlock` |

**CLI:**

```
drive group copy <groupId> <fileId> --json '{...}'
drive group rename <groupId> <fileId> --json '{...}'
drive group move <groupId> <fileId> --json '{...}'
drive group protect <groupId> <fileId>
drive group unprotect <groupId> <fileId>
drive group lock <groupId> <fileId>
drive group unlock <groupId> <fileId>
```

**Definition of Done:**
1. API 메서드 구현 및 컴파일 성공
2. Cobra 커맨드 등록 및 `--help` 출력 확인
3. `cmd/smoke_test.go`에 등록 테스트 추가

---

### Task 5-6: GroupFolder 리비전 + 휴지통 (7개)

**Files:**
- Modify: `internal/api/drive_group.go`
- Modify: `cmd/drive.go`
- Modify: `cmd/smoke_test.go`

**Dependencies:** Task 5-4

**추가 API:**

| # | 메서드 | HTTP | 경로 |
|---|--------|------|------|
| 1 | ListGroupRevisions | GET | `/groups/{groupId}/folder/files/{fileId}/revisions` |
| 2 | GetGroupRevision | GET | `/groups/{groupId}/folder/files/{fileId}/revisions/{revisionId}` |
| 3 | RestoreGroupRevision | POST | `/groups/{groupId}/folder/files/{fileId}/revisions/{revisionId}/restore` |
| 4 | GetGroupRevisionDownloadUrl | GET | `/groups/{groupId}/folder/files/{fileId}/revisions/{revisionId}/download` |
| 5 | ListGroupTrashFiles | GET | `/groups/{groupId}/folder/trash-files` |
| 6 | RestoreGroupTrashFile | POST | `/groups/{groupId}/folder/trash-files/{fileId}/restore` |
| 7 | DeleteGroupTrashFile | DELETE | `/groups/{groupId}/folder/trash-files/{fileId}` |

**CLI:**

```
drive group revision list <groupId> <fileId>
drive group revision get <groupId> <fileId> <revisionId>
drive group revision restore <groupId> <fileId> <revisionId>
drive group revision download <groupId> <fileId> <revisionId>
drive group trash-list <groupId>
drive group trash-restore <groupId> <fileId>
drive group trash-delete <groupId> <fileId>
```

**Definition of Done:**
1. API 메서드 구현 및 컴파일 성공
2. Cobra 커맨드 등록 및 `--help` 출력 확인
3. `cmd/smoke_test.go`에 등록 테스트 추가

---

### Task 5-7: GroupFolder 링크 + 권한 (11개)

**Files:**
- Modify: `internal/api/drive_group.go`
- Modify: `cmd/drive.go`
- Modify: `cmd/smoke_test.go`

**Dependencies:** Task 5-4

**추가 API:**

| # | 메서드 | HTTP | 경로 |
|---|--------|------|------|
| 1 | GetGroupLinkSetting | GET | `/groups/{groupId}/folder/link-setting` |
| 2 | GetGroupLink | GET | `/groups/{groupId}/folder/files/{fileId}/link` |
| 3 | CreateGroupLink | POST | `/groups/{groupId}/folder/files/{fileId}/link` |
| 4 | PatchGroupLink | PATCH | `/groups/{groupId}/folder/files/{fileId}/link` |
| 5 | DeleteGroupLink | DELETE | `/groups/{groupId}/folder/files/{fileId}/link` |
| 6 | ListGroupPermissions | GET | `/groups/{groupId}/folder/files/{fileId}/permissions` |
| 7 | CreateGroupPermission | POST | `/groups/{groupId}/folder/files/{fileId}/permissions` |
| 8 | GetGroupPermission | GET | `/groups/{groupId}/folder/files/{fileId}/permissions/{permissionId}` |
| 9 | PatchGroupPermission | PATCH | `/groups/{groupId}/folder/files/{fileId}/permissions/{permissionId}` |
| 10 | DeleteGroupPermission | DELETE | `/groups/{groupId}/folder/files/{fileId}/permissions/{permissionId}` |
| 11 | DeleteAllGroupPermissions | DELETE | `/groups/{groupId}/folder/files/{fileId}/permissions` |

**CLI:**

```
drive group link-setting <groupId>
drive group link get <groupId> <fileId>
drive group link create <groupId> <fileId> --json '{...}'
drive group link update <groupId> <fileId> --json '{...}'
drive group link delete <groupId> <fileId>
drive group permission list <groupId> <fileId>
drive group permission create <groupId> <fileId> --json '{...}'
drive group permission get <groupId> <fileId> <permissionId>
drive group permission update <groupId> <fileId> <permissionId> --json '{...}'
drive group permission delete <groupId> <fileId> <permissionId>
drive group permission delete-all <groupId> <fileId>
```

**Definition of Done:**
1. API 메서드 구현 및 컴파일 성공
2. Cobra 커맨드 등록 및 `--help` 출력 확인
3. `cmd/smoke_test.go`에 등록 테스트 추가

---

### Task 5-8: SharedDrive 관리 + 파일 보완 (8개)

**Files:**
- Modify: `internal/api/drive_shared.go`
- Modify: `cmd/drive.go`
- Modify: `cmd/smoke_test.go`

**Dependencies:** Task 0-3 (파일 업로드 헬퍼)

**추가 API:**

| # | 메서드 | HTTP | 경로 |
|---|--------|------|------|
| 1 | CreateSharedDrive | POST | `/sharedrives` |
| 2 | PatchSharedDrive | PATCH | `/sharedrives/{driveId}` |
| 3 | DeleteSharedDrive | DELETE | `/sharedrives/{driveId}` |
| 4 | CreateSharedDriveFolderInRoot | POST | `/sharedrives/{driveId}/files/createfolder` |
| 5 | CreateSharedDriveSubFolder | POST | `/sharedrives/{driveId}/files/{fileId}/createfolder` |
| 6 | DeleteSharedDriveFile | DELETE | `/sharedrives/{driveId}/files/{fileId}` |
| 7 | CreateSharedDriveUploadUrlInFolder | POST | `/sharedrives/{driveId}/files/{fileId}` |
| 8 | CreateSharedDriveRootUploadUrl | POST | `/sharedrives/{driveId}/files` |

**CLI:**

```
drive shared create-drive --json '{...}'
drive shared update-drive <driveId> --json '{...}'
drive shared delete-drive <driveId>
drive shared mkdir <driveId> --json '{...}'
drive shared mkdir <driveId> --parent <fileId> --json '{...}'
drive shared delete <driveId> <fileId>
drive shared upload <driveId> --file <path>              # 업로드 방식: presigned URL
drive shared upload <driveId> --folder <fileId> --file <path>
```

**Definition of Done:**
1. API 메서드 구현 및 컴파일 성공
2. Cobra 커맨드 등록 및 `--help` 출력 확인
3. `cmd/smoke_test.go`에 등록 테스트 추가

---

### Task 5-9: SharedDrive 파일조작 (7개)

**Files:**
- Modify: `internal/api/drive_shared.go`
- Modify: `cmd/drive.go`
- Modify: `cmd/smoke_test.go`

**Dependencies:** Task 5-8

**추가 API:**

| # | 메서드 | HTTP | 경로 |
|---|--------|------|------|
| 1 | CopySharedDriveFile | POST | `/sharedrives/{driveId}/files/{fileId}/copy` |
| 2 | RenameSharedDriveFile | POST | `/sharedrives/{driveId}/files/{fileId}/rename` |
| 3 | MoveSharedDriveFile | POST | `/sharedrives/{driveId}/files/{fileId}/move` |
| 4 | ProtectSharedDriveFile | POST | `/sharedrives/{driveId}/files/{fileId}/protect` |
| 5 | UnprotectSharedDriveFile | POST | `/sharedrives/{driveId}/files/{fileId}/unprotect` |
| 6 | LockSharedDriveFile | POST | `/sharedrives/{driveId}/files/{fileId}/lock` |
| 7 | UnlockSharedDriveFile | POST | `/sharedrives/{driveId}/files/{fileId}/unlock` |

**CLI:**

```
drive shared copy <driveId> <fileId> --json '{...}'
drive shared rename <driveId> <fileId> --json '{...}'
drive shared move <driveId> <fileId> --json '{...}'
drive shared protect <driveId> <fileId>
drive shared unprotect <driveId> <fileId>
drive shared lock <driveId> <fileId>
drive shared unlock <driveId> <fileId>
```

**Definition of Done:**
1. API 메서드 구현 및 컴파일 성공
2. Cobra 커맨드 등록 및 `--help` 출력 확인
3. `cmd/smoke_test.go`에 등록 테스트 추가

---

### Task 5-10: SharedDrive 리비전 + 휴지통 (7개)

**Files:**
- Modify: `internal/api/drive_shared.go`
- Modify: `cmd/drive.go`
- Modify: `cmd/smoke_test.go`

**Dependencies:** Task 5-8

**추가 API:**

| # | 메서드 | HTTP | 경로 |
|---|--------|------|------|
| 1 | ListSharedDriveRevisions | GET | `/sharedrives/{driveId}/files/{fileId}/revisions` |
| 2 | GetSharedDriveRevision | GET | `/sharedrives/{driveId}/files/{fileId}/revisions/{revisionId}` |
| 3 | RestoreSharedDriveRevision | POST | `/sharedrives/{driveId}/files/{fileId}/revisions/{revisionId}/restore` |
| 4 | GetSharedDriveRevisionDownloadUrl | GET | `/sharedrives/{driveId}/files/{fileId}/revisions/{revisionId}/download` |
| 5 | ListSharedDriveTrashFiles | GET | `/sharedrives/{driveId}/trash-files` |
| 6 | RestoreSharedDriveTrashFile | POST | `/sharedrives/{driveId}/trash-files/{fileId}/restore` |
| 7 | DeleteSharedDriveTrashFile | DELETE | `/sharedrives/{driveId}/trash-files/{fileId}` |

**CLI:**

```
drive shared revision list <driveId> <fileId>
drive shared revision get <driveId> <fileId> <revisionId>
drive shared revision restore <driveId> <fileId> <revisionId>
drive shared revision download <driveId> <fileId> <revisionId>
drive shared trash-list <driveId>
drive shared trash-restore <driveId> <fileId>
drive shared trash-delete <driveId> <fileId>
```

**Definition of Done:**
1. API 메서드 구현 및 컴파일 성공
2. Cobra 커맨드 등록 및 `--help` 출력 확인
3. `cmd/smoke_test.go`에 등록 테스트 추가

---

### Task 5-11: SharedDrive 링크 + 권한 (21개)

**Files:**
- Modify: `internal/api/drive_shared.go`
- Modify: `cmd/drive.go`
- Modify: `cmd/smoke_test.go`

**Dependencies:** Task 5-8

**추가 API:**

| # | 메서드 | HTTP | 경로 |
|---|--------|------|------|
| 1 | GetSharedDriveLinkSetting | GET | `/sharedrives/{driveId}/link-setting` |
| 2 | GetSharedDriveLink | GET | `/sharedrives/{driveId}/files/{fileId}/link` |
| 3 | CreateSharedDriveLink | POST | `/sharedrives/{driveId}/files/{fileId}/link` |
| 4 | PatchSharedDriveLink | PATCH | `/sharedrives/{driveId}/files/{fileId}/link` |
| 5 | DeleteSharedDriveLink | DELETE | `/sharedrives/{driveId}/files/{fileId}/link` |
| 6 | ListSharedDrivePermissions | GET | `/sharedrives/{driveId}/permissions` |
| 7 | CreateSharedDrivePermission | POST | `/sharedrives/{driveId}/permissions` |
| 8 | GetSharedDrivePermission | GET | `/sharedrives/{driveId}/permissions/{permissionId}` |
| 9 | PatchSharedDrivePermission | PATCH | `/sharedrives/{driveId}/permissions/{permissionId}` |
| 10 | DeleteSharedDrivePermission | DELETE | `/sharedrives/{driveId}/permissions/{permissionId}` |
| 11 | DeleteAllSharedDrivePermissions | DELETE | `/sharedrives/{driveId}/permissions` |
| 12 | EnableSharedDrivePermissions | POST | `/sharedrives/{driveId}/permissions/enable` |
| 13 | DisableSharedDrivePermissions | POST | `/sharedrives/{driveId}/permissions/disable` |
| 14 | ListSharedDriveFilePermissions | GET | `/sharedrives/{driveId}/files/{fileId}/permissions` |
| 15 | CreateSharedDriveFilePermission | POST | `/sharedrives/{driveId}/files/{fileId}/permissions` |
| 16 | GetSharedDriveFilePermission | GET | `/sharedrives/{driveId}/files/{fileId}/permissions/{permissionId}` |
| 17 | PatchSharedDriveFilePermission | PATCH | `/sharedrives/{driveId}/files/{fileId}/permissions/{permissionId}` |
| 18 | DeleteSharedDriveFilePermission | DELETE | `/sharedrives/{driveId}/files/{fileId}/permissions/{permissionId}` |
| 19 | DeleteAllSharedDriveFilePermissions | DELETE | `/sharedrives/{driveId}/files/{fileId}/permissions` |
| 20 | EnableSharedDriveFilePermissions | POST | `/sharedrives/{driveId}/files/{fileId}/permissions/enable` |
| 21 | DisableSharedDriveFilePermissions | POST | `/sharedrives/{driveId}/files/{fileId}/permissions/disable` |

**CLI:**

```
drive shared link-setting <driveId>
drive shared link get <driveId> <fileId>
drive shared link create <driveId> <fileId> --json '{...}'
drive shared link update <driveId> <fileId> --json '{...}'
drive shared link delete <driveId> <fileId>
drive shared permission list <driveId>
drive shared permission create <driveId> --json '{...}'
drive shared permission get <driveId> <permissionId>
drive shared permission update <driveId> <permissionId> --json '{...}'
drive shared permission delete <driveId> <permissionId>
drive shared permission delete-all <driveId>
drive shared permission enable <driveId>
drive shared permission disable <driveId>
drive shared file-permission list <driveId> <fileId>
drive shared file-permission create <driveId> <fileId> --json '{...}'
drive shared file-permission get <driveId> <fileId> <permissionId>
drive shared file-permission update <driveId> <fileId> <permissionId> --json '{...}'
drive shared file-permission delete <driveId> <fileId> <permissionId>
drive shared file-permission delete-all <driveId> <fileId>
drive shared file-permission enable <driveId> <fileId>
drive shared file-permission disable <driveId> <fileId>
```

**Definition of Done:**
1. API 메서드 구현 및 컴파일 성공
2. Cobra 커맨드 등록 및 `--help` 출력 확인
3. `cmd/smoke_test.go`에 등록 테스트 추가

---

### Task 5-12: SharedFolder 관리 + 파일 (12개)

**Files:**
- Create: `internal/api/drive_sharedfolder.go`
- Modify: `cmd/drive.go`
- Modify: `cmd/smoke_test.go`

**Dependencies:** Task 0-3 (파일 업로드 헬퍼)

**추가 API:**

| # | 메서드 | HTTP | 경로 |
|---|--------|------|------|
| 1 | GetSharedFolder | GET | `/users/{userId}/drive/sharedfolders/{sharedFolderId}` |
| 2 | LeaveSharedFolder | DELETE | `/users/{userId}/drive/sharedfolders/{sharedFolderId}` |
| 3 | ListSharedFolderMembers | GET | `/users/{userId}/drive/sharedfolders/{sharedFolderId}/members` |
| 4 | ListSharedFolderRootFiles | GET | `/users/{userId}/drive/sharedfolders/{sharedFolderId}/files` |
| 5 | CreateSharedFolderRootUploadUrl | POST | `/users/{userId}/drive/sharedfolders/{sharedFolderId}/files` |
| 6 | CreateSharedFolderInRoot | POST | `/users/{userId}/drive/sharedfolders/{sharedFolderId}/files/createfolder` |
| 7 | CreateSharedSubFolder | POST | `/users/{userId}/drive/sharedfolders/{sharedFolderId}/files/{fileId}/createfolder` |
| 8 | GetSharedFolderFile | GET | `/users/{userId}/drive/sharedfolders/{sharedFolderId}/files/{fileId}` |
| 9 | ListSharedFolderChildren | GET | `/users/{userId}/drive/sharedfolders/{sharedFolderId}/files/{fileId}/children` |
| 10 | DeleteSharedFolderFile | DELETE | `/users/{userId}/drive/sharedfolders/{sharedFolderId}/files/{fileId}` |
| 11 | CreateSharedFolderUploadUrl | POST | `/users/{userId}/drive/sharedfolders/{sharedFolderId}/files/{fileId}` |
| 12 | GetSharedFolderDownloadUrl | GET | `/users/{userId}/drive/sharedfolders/{sharedFolderId}/files/{fileId}/download` |

**CLI:**

```
drive shared-folder get <sharedFolderId>
drive shared-folder leave <sharedFolderId>
drive shared-folder list-members <sharedFolderId>
drive shared-folder list <sharedFolderId>
drive shared-folder list <sharedFolderId> --folder <fileId>
drive shared-folder get-file <sharedFolderId> <fileId>
drive shared-folder mkdir <sharedFolderId> --json '{...}'
drive shared-folder mkdir <sharedFolderId> --parent <fileId> --json '{...}'
drive shared-folder delete <sharedFolderId> <fileId>
drive shared-folder upload <sharedFolderId> --file <path>            # 업로드 방식: presigned URL
drive shared-folder upload <sharedFolderId> --folder <fileId> --file <path>
drive shared-folder download <sharedFolderId> <fileId>
```

**Definition of Done:**
1. API 메서드 구현 및 컴파일 성공
2. Cobra 커맨드 등록 및 `--help` 출력 확인
3. `cmd/smoke_test.go`에 등록 테스트 추가

---

### Task 5-13: SharedFolder 파일조작 + 리비전 (11개)

**Files:**
- Modify: `internal/api/drive_sharedfolder.go`
- Modify: `cmd/drive.go`
- Modify: `cmd/smoke_test.go`

**Dependencies:** Task 5-12

**추가 API:**

| # | 메서드 | HTTP | 경로 |
|---|--------|------|------|
| 1 | CopySharedFolderFile | POST | `/users/{userId}/drive/sharedfolders/{sharedFolderId}/files/{fileId}/copy` |
| 2 | RenameSharedFolderFile | POST | `/users/{userId}/drive/sharedfolders/{sharedFolderId}/files/{fileId}/rename` |
| 3 | MoveSharedFolderFile | POST | `/users/{userId}/drive/sharedfolders/{sharedFolderId}/files/{fileId}/move` |
| 4 | ProtectSharedFolderFile | POST | `/users/{userId}/drive/sharedfolders/{sharedFolderId}/files/{fileId}/protect` |
| 5 | UnprotectSharedFolderFile | POST | `/users/{userId}/drive/sharedfolders/{sharedFolderId}/files/{fileId}/unprotect` |
| 6 | LockSharedFolderFile | POST | `/users/{userId}/drive/sharedfolders/{sharedFolderId}/files/{fileId}/lock` |
| 7 | UnlockSharedFolderFile | POST | `/users/{userId}/drive/sharedfolders/{sharedFolderId}/files/{fileId}/unlock` |
| 8 | ListSharedFolderRevisions | GET | `/users/{userId}/drive/sharedfolders/{sharedFolderId}/files/{fileId}/revisions` |
| 9 | GetSharedFolderRevision | GET | `/users/{userId}/drive/sharedfolders/{sharedFolderId}/files/{fileId}/revisions/{revisionId}` |
| 10 | RestoreSharedFolderRevision | POST | `/users/{userId}/drive/sharedfolders/{sharedFolderId}/files/{fileId}/revisions/{revisionId}/restore` |
| 11 | GetSharedFolderRevisionDownloadUrl | GET | `/users/{userId}/drive/sharedfolders/{sharedFolderId}/files/{fileId}/revisions/{revisionId}/download` |

**CLI:**

```
drive shared-folder copy <sharedFolderId> <fileId> --json '{...}'
drive shared-folder rename <sharedFolderId> <fileId> --json '{...}'
drive shared-folder move <sharedFolderId> <fileId> --json '{...}'
drive shared-folder protect <sharedFolderId> <fileId>
drive shared-folder unprotect <sharedFolderId> <fileId>
drive shared-folder lock <sharedFolderId> <fileId>
drive shared-folder unlock <sharedFolderId> <fileId>
drive shared-folder revision list <sharedFolderId> <fileId>
drive shared-folder revision get <sharedFolderId> <fileId> <revisionId>
drive shared-folder revision restore <sharedFolderId> <fileId> <revisionId>
drive shared-folder revision download <sharedFolderId> <fileId> <revisionId>
```

**Definition of Done:**
1. API 메서드 구현 및 컴파일 성공
2. Cobra 커맨드 등록 및 `--help` 출력 확인
3. `cmd/smoke_test.go`에 등록 테스트 추가

---

### Task 5-14: SharedFolder 링크 (5개)

**Files:**
- Modify: `internal/api/drive_sharedfolder.go`
- Modify: `cmd/drive.go`
- Modify: `cmd/smoke_test.go`

**Dependencies:** Task 5-12

**추가 API:**

| # | 메서드 | HTTP | 경로 |
|---|--------|------|------|
| 1 | GetSharedFolderLinkSetting | GET | `/users/{userId}/drive/sharedfolders/{sharedFolderId}/link-setting` |
| 2 | GetSharedFolderLink | GET | `/users/{userId}/drive/sharedfolders/{sharedFolderId}/files/{fileId}/link` |
| 3 | CreateSharedFolderLink | POST | `/users/{userId}/drive/sharedfolders/{sharedFolderId}/files/{fileId}/link` |
| 4 | PatchSharedFolderLink | PATCH | `/users/{userId}/drive/sharedfolders/{sharedFolderId}/files/{fileId}/link` |
| 5 | DeleteSharedFolderLink | DELETE | `/users/{userId}/drive/sharedfolders/{sharedFolderId}/files/{fileId}/link` |

**CLI:**

```
drive shared-folder link-setting <sharedFolderId>
drive shared-folder link get <sharedFolderId> <fileId>
drive shared-folder link create <sharedFolderId> <fileId> --json '{...}'
drive shared-folder link update <sharedFolderId> <fileId> --json '{...}'
drive shared-folder link delete <sharedFolderId> <fileId>
```

**Definition of Done:**
1. API 메서드 구현 및 컴파일 성공
2. Cobra 커맨드 등록 및 `--help` 출력 확인
3. `cmd/smoke_test.go`에 등록 테스트 추가

---

### Task 5-15: Coverage Gap 12개 해소

**Files:**
- Create: `docs/coverage-gap-12.md`
- Modify: 해당 도메인 API 파일
- Modify: 해당 도메인 CLI 파일
- Modify: `cmd/smoke_test.go`
- Modify: coverage ledger 문서

**Dependencies:** Task 0-0 완료, 관련 선행 도메인 Task 완료

**선행 산출물:**
- `docs/coverage-gap-12.md`에 누락 12개 endpoint를 `도메인 / HTTP / 경로 / 구현 파일 / CLI 커맨드 / Auth/Identity` 형식으로 먼저 확정한다
- 이 문서가 채워지기 전에는 Task 5-15 구현 착수 금지

**API 메서드:** `docs/coverage-gap-12.md`에 확정된 12개 endpoint를 1:1로 기입

**CLI 커맨드:** `docs/coverage-gap-12.md`에 확정된 12개 endpoint 대응 명령을 1:1로 기입

**Auth/Identity:**
- Required scopes: `docs/auth-identity-matrix.md` 참조
- OAuth/JWT 지원: `docs/auth-identity-matrix.md` 참조
- userId=me 허용 여부: `docs/auth-identity-matrix.md` 참조
- CLI identity 처리: 각 endpoint별 명시

**Definition of Done:**
1. 누락 12개 endpoint가 `docs/coverage-gap-12.md`에 전수 확정됨
2. 12개 endpoint가 도메인별 하위 작업 또는 체크리스트로 분해됨
3. 각 endpoint별 구현 파일, CLI 커맨드, Auth/Identity, smoke test가 1:1로 문서화됨
4. 문서 내 총합과 `scripts/verify-coverage-ledger.go` 결과가 모두 538/538로 일치함

**하위 태스크 분해:**
- Task 5-15는 12개 하위 태스크(5-15a ~ 5-15l)로 분해한다
- 각 하위 태스크는 `Files / Dependencies / API 메서드 / CLI 커맨드 / Auth/Identity / Definition of Done`를 모두 포함한다
- Task 0-0 완료 직후 아래 표를 확정값으로 채우기 전까지 구현 착수 금지

| Subtask | Domain | HTTP | Path | Files | CLI Command | Dependency |
|---------|--------|------|------|-------|-------------|------------|
| 5-15a | TBD | TBD | TBD | TBD | TBD | Task 0-0 + 관련 도메인 Task |
| 5-15b | TBD | TBD | TBD | TBD | TBD | Task 0-0 + 관련 도메인 Task |
| 5-15c | TBD | TBD | TBD | TBD | TBD | Task 0-0 + 관련 도메인 Task |
| 5-15d | TBD | TBD | TBD | TBD | TBD | Task 0-0 + 관련 도메인 Task |
| 5-15e | TBD | TBD | TBD | TBD | TBD | Task 0-0 + 관련 도메인 Task |
| 5-15f | TBD | TBD | TBD | TBD | TBD | Task 0-0 + 관련 도메인 Task |
| 5-15g | TBD | TBD | TBD | TBD | TBD | Task 0-0 + 관련 도메인 Task |
| 5-15h | TBD | TBD | TBD | TBD | TBD | Task 0-0 + 관련 도메인 Task |
| 5-15i | TBD | TBD | TBD | TBD | TBD | Task 0-0 + 관련 도메인 Task |
| 5-15j | TBD | TBD | TBD | TBD | TBD | Task 0-0 + 관련 도메인 Task |
| 5-15k | TBD | TBD | TBD | TBD | TBD | Task 0-0 + 관련 도메인 Task |
| 5-15l | TBD | TBD | TBD | TBD | TBD | Task 0-0 + 관련 도메인 Task |

---

## 실행 순서 요약

```
Phase 0: 공통 인프라 (6개 태스크)
    Task 0-0: Baseline Ledger 작성 (기존 116개)
    Task 0-1: readJSONFlag / readJSONFlagRaw
    Task 0-2: raw body 전송 전략 확정
    Task 0-3: 파일 업로드/다운로드 공통 헬퍼
    Task 0-4: 업로드 스펙 매트릭스 확정
    Task 0-5: Auth/Identity 매트릭스 확정
    ↓
Phase 1: 기존 도메인 API 보완 (125개, 10개 태스크)
    Task 1-1: Calendar (14)
    Task 1-2: Board (19)
    Task 1-3: Mail (17)
    Task 1-4: Task (9)
    Task 1-5: Contact (16)
    Task 1-6: Note (5)
    Task 1-7: Attendance (10)
    Task 1-8: HR (10)
    Task 1-9: Audit (7)
    Task 1-10: Approval (18)
    ↓
Phase 2: Security (3개, 1개 태스크)
    Task 2-1: Security (3)
    ↓
Phase 3: Bot 전체 (36개, 6개 태스크)
    Task 3-1: Bot 관리 (7)
    Task 3-2: 구조화 메시지 + 첨부 (4)
    Task 3-3: 채널 (2)
    Task 3-4: 도메인 (8)
    Task 3-5: 고정메뉴 (3)
    Task 3-6: 리치메뉴 (12)
    ↓
Phase 4: Directory 전체 (121개, 12개 태스크)
    Task 4-1: User CUD (12)
    Task 4-2: User Profile (9)
    Task 4-3: Email + Invitations + Links (12)
    Task 4-4: External Keys + Custom Properties (7)
    Task 4-5: Group CUD + Members + Admins + ExtKeys (12)
    Task 4-6: OrgUnit CUD + Members + AccessRestrict + ExtKeys (12)
    Task 4-7: Positions CRUD + ExtKeys (9)
    Task 4-8: Levels CRUD + ExtKeys (9)
    Task 4-9: Employment Types CRUD + ExtKeys + AccessRestrict (13)
    Task 4-10: User Types CRUD + ExtKeys + AccessRestrict (13)
    Task 4-11: Profile Statuses CRUD (8)
    Task 4-12: Custom Fields CRUD (5)
    ↓
Phase 5: Drive 전체 + Gap (125+12개, 15개 태스크)
    Task 5-1: MyDrive 파일조작 (7)
    Task 5-2: MyDrive 리비전 (4)
    Task 5-3: MyDrive 휴지통+링크+공유 (11)
    Task 5-4: GroupFolder 관리+파일보완 (7)
    Task 5-5: GroupFolder 파일조작 (7)
    Task 5-6: GroupFolder 리비전+휴지통 (7)
    Task 5-7: GroupFolder 링크+권한 (11)
    Task 5-8: SharedDrive 관리+파일보완 (8)
    Task 5-9: SharedDrive 파일조작 (7)
    Task 5-10: SharedDrive 리비전+휴지통 (7)
    Task 5-11: SharedDrive 링크+권한 (21)
    Task 5-12: SharedFolder 관리+파일 (12)
    Task 5-13: SharedFolder 파일조작+리비전 (11)
    Task 5-14: SharedFolder 링크 (5)
    Task 5-15: Coverage Gap 12개 해소
```

**총 태스크:** 50개 (Phase 0 포함)
**총 신규 API:** 422개 (410 + Gap 12) (현 Task 합산 기준)

### Coverage Reconciliation

- 초기 분석에서 신규 구현 대상은 422개로 추정했으나, Task 상세화 후 합산 기준 410개임
- 차이 12개는 SDK endpoint 집계와 Task 세분화 과정에서의 카운트 차이 (중복 제거, 경계 판단 등)
- Phase 0의 Task 0-0 (Baseline Ledger 작성) 시점에 SDK 538개 endpoint를 전수 재확인하고, 기존 116개 + 신규 410개 = 526개와 538개의 차이 12개를 식별한다
- 식별된 누락 12개 endpoint는 `Task 5-15: Coverage Gap 12개 해소`에 편성한다. Task 0-0 완료 후 endpoint 상세를 Task 5-15에 기입한다
- Task 0-0 완료 산출물에는 누락 12개 endpoint의 `도메인 / HTTP / 경로 / 예정 Task` 목록을 반드시 포함하며, 이 목록은 즉시 `Task 5-15` 본문에 반영한다
- Coverage Reconciliation이 완료되기 전에는 본 문서를 `100% 커버리지 달성` 근거 문서로 사용하지 않는다
- Baseline Ledger와 신규 Task endpoint 합계는 수동 계산만 사용하지 않고, `scripts/verify-coverage-ledger.go`로 자동 검증한다. 입력은 SDK endpoint 목록 + ledger 문서, 출력은 `총계/누락/중복` 리포트

---

## Task 공통 규칙

각 Task는 아래 항목을 반드시 포함한다:

- **Files:** 생성/수정 대상 파일 경로
- **Dependencies:** 선행 Task 번호 (없으면 "없음")
- **API 메서드 표:** HTTP method + 경로 1:1 매핑 (인프라 Task는 "없음")
- **CLI 커맨드:** 사용자가 실행할 명령어 (인프라 Task는 "없음")
- **Auth/Identity:** required scope, OAuth/JWT 지원 여부, `userId=me` 허용 여부, CLI에서 `<userId>` 위치 인자 또는 `--user-id` 플래그 처리 방식을 명시 (인프라 Task는 생략 가능)

아래 5개 항목은 모든 Task에 자동 적용되는 공통 완료 기준이다. 각 Task 하위 `Definition of Done`의 3개 항목은 요약본이며, 공통 완료 기준 5개를 대체하지 않는다.

### Task Definition of Done (각 Task 완료 기준)

1. API 메서드 구현 및 컴파일 성공
2. Cobra 커맨드 등록 및 `--help` 출력 확인
3. `cmd/smoke_test.go`에 등록 테스트와 입력 검증 테스트 추가
4. 파일 업로드/다운로드 endpoint는 성공/실패 케이스 테스트 추가
5. 변경 도메인 관련 패키지 테스트와 패키지 범위 `go vet` 통과

### Phase Definition of Done (각 Phase 완료 기준)

1. `go vet ./...` 통과
2. `go test ./... -v` 전체 통과
3. `make build` 성공
4. 스모크 테스트에 신규 커맨드 등록 테스트 전수 확인
5. 커버리지 카운트 검증 (Phase별 목표 API 수 달성 여부)

### 파일 업로드 사양 규칙

`--file` 플래그를 사용하는 Task는 착수 전 `naverworks-sdk-kotlin`과 공식 REST 문서를 대조해 `presigned URL | multipart/form-data | raw binary` 중 하나를 확정하고, 확정된 방식과 `Content-Type`을 해당 Task의 `업로드 사양` 섹션에 기입한다. 미기입 상태에서는 구현 시작 금지.

### Task 템플릿 강제 규칙

문서에 `Files`, `Dependencies`, `Auth/Identity`, `Definition of Done` 섹션이 빠진 Task는 착수 전 반드시 보강한다. 생략 상태로는 구현 시작하지 않는다.

`Auth/Identity` 섹션 기본 형식:
```
**Auth/Identity:**
- Required scopes: <scope 목록 또는 없음>
- OAuth/JWT 지원: <OAuth | JWT | 둘 다 | 미확인>
- userId=me 허용 여부: <허용 | 불가 | 해당 없음>
- CLI identity 처리: <userId> 위치 인자 | --user-id 플래그 | 프로필 기본값
```

현재 본문의 Task에는 Auth/Identity가 일괄 미기입 상태이다. Task 0-5 (Auth/Identity 매트릭스 확정)에서 전 endpoint 인증 요건을 조사하고, 모든 구현 Task에 Auth/Identity 섹션을 채운다. Task 0-5 완료 전에는 구현 Task 착수 금지.

**문서 일관성 규칙:**
- 모든 구현 Task의 `Dependencies`에는 최소 `Task 0-5`를 포함한다
- `Dependencies: Phase 0 완료` 표기는 `Phase 0 완료 (특히 Task 0-5 필수)`를 의미한다
- 업로드 관련 Task의 `Dependencies: Tasks 0-3, 0-4` 표기는 `Tasks 0-3, 0-4, 0-5`를 의미한다

업로드 관련 Task(1-2, 1-5, 1-6, 1-10, 3-2, 3-6, 4-2)는 모두 `Dependencies: Tasks 0-3, 0-4`로 표기한다.

---

## 리스크 및 대응

| 리스크 | 근거 | 영향 | 대응 |
|--------|------|------|------|
| API 스펙 드리프트 | 기준이 `naverworks-sdk-kotlin`이라 공식 REST 문서와 차이 가능 | 잘못된 경로/필드 구현 | 각 Phase 시작 전 SDK 정의와 공식 문서 대조, 차이점은 `docs/api-diffs/<domain>.md`에 기록 |
| 인증/권한 스코프 차이 | OAuth/JWT별 허용 scope와 `userId=me` 제약이 다름 | 런타임 실패 | 각 Task에 required scopes, OAuth/JWT 제약, `--user-id` 필요 여부 구현 시 확인 |
| 파일 업로드 방식 불확실성 | 도메인별로 presigned URL, multipart, raw binary 방식이 다를 수 있음 | 구현 지연 및 중복 코드 | Task 0-3 선행 완료 전 `--file` 커맨드 구현 시작 금지 |
| 병합 충돌 | `cmd/drive.go`, `internal/api/directory.go`에 태스크 집중 | 리드타임 증가 | 도메인별 브랜치 분리, PR 크기 제한, Phase별 freeze 적용 |
| 검증 비용 증가 | 매 Task마다 전체 `go vet`, `go test`, `make build` 실행 | 일정 지연 | Task 완료 시 도메인 단위 검증, Phase 완료 시 전체 검증으로 2단계 운영 |
| 권한/검증 환경 부족 | 일부 API는 관리자 권한 필요 | 실 API 검증 실패 | 권한별 테스트 프로필 확보, smoke 테스트와 실 API 검증 분리 |
| 일정 초과 | 410+개 endpoint 추가 | Phase 지연 | Phase별 종료 기준 정의, 미완료 endpoint는 다음 Phase로 이월 금지 |

---

## 검증 체크리스트

**Task 완료 시 (도메인 단위):**
1. 변경 패키지 범위에서 `go test` 실행 후 통과 (예: `go test ./cmd ./internal/api -run 'TestSmoke_(Calendar|Board)' -v`)
2. 변경 패키지 범위에서 `go vet` 실행 후 통과
3. 신규 커맨드 `--help` 출력 확인
4. 전체 검증(`go vet ./...`, `go test ./... -v`, `make build`)은 Phase 완료 시에만 수행한다

**Phase 완료 시 (전체):**
1. `go vet ./...` 통과
2. `go test ./... -v` 전체 통과
3. `make build` 성공
4. 스모크 테스트에 신규 커맨드 등록 테스트 전수 확인
5. 커버리지 카운트 검증 (Phase별 목표 API 수 달성 여부)

---

## 커밋 전략

- 태스크 단위로 커밋한다
- 커밋 메시지는 Conventional Commits 형식을 유지하되 한국어로 작성한다
- 커밋 작업은 `commit-work` 스킬로 수행한다
- 예: `feat(calendar): 캘린더 CUD 커맨드 추가`
- Phase 단위로 태그한다 (예: `v0.2.0-phase1`)
- 각 커밋에 추가된 API 수를 명시한다
