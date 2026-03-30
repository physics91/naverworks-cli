# Upload Spec Matrix

NAVER WORKS REST API v1.0에서 파일 업로드가 필요한 모든 endpoint의 업로드 방식을 정리한 문서.

> **조사 기준**: `naverworks-sdk-kotlin` 소스 코드 + NAVER WORKS 공식 문서

## 업로드 방식 요약

NAVER WORKS API의 파일 업로드는 **2-step presigned URL** 패턴을 공통으로 사용한다:

1. **Step 1** — NAVER WORKS API에 JSON POST 요청 → 응답으로 `uploadUrl` (presigned URL)을 받음
2. **Step 2** — 받은 `uploadUrl`에 실제 파일을 PUT (Authorization 헤더 불필요, URL 자체가 서명됨)

Bot Rich Menu Image는 예외적으로 파일을 직접 업로드하지 않고, 이미 업로드된 bot attachment의 `fileId`를 JSON으로 참조한다.

## Endpoint 매트릭스

| # | Endpoint | HTTP | Path | 업로드 방식 | Step 1 Content-Type | Step 1 Request Body | Step 1 응답 형식 | Step 2 Method | Step 2 Content-Type | 비고 |
|---|----------|------|------|-----------|-------------------|-------------------|----------------|-------------|-------------------|------|
| 1 | Board post attachment | POST | `/boards/{boardId}/posts/{postId}/attachments` | presigned URL | `application/json` | `{fileName, fileSize, contentType}` | `{uploadUrl}` | PUT | `application/octet-stream` | fileName ≤ 200자, fileSize ≤ 2GB |
| 2 | Board comment attachment | POST | `/boards/{boardId}/posts/{postId}/comments/{commentId}/attachments` | presigned URL | `application/json` | `{fileName, fileSize, contentType}` | `{uploadUrl}` | PUT | `application/octet-stream` | Board post attachment와 동일 스펙 |
| 3 | Contact photo | POST | `/contacts/{contactId}/photo` | presigned URL | `application/json` | `{fileName, fileSize}` | `{uploadUrl}` | PUT | `application/octet-stream` | fileSize ≤ 10MB, contentType 필드 없음 |
| 4 | Note attachment | POST | `/groups/{groupId}/note/posts/{postId}/attachments` | presigned URL | `application/json` | `{fileName, fileSize, contentType}` | `{uploadUrl}` | PUT | `application/octet-stream` | fileName ≤ 200자 |
| 5 | Approval user document attachment | POST | `/business-support/approval/users/{userId}/documents/attachments` | presigned URL | `application/json` | `{documentFormId, componentId, fileName, fileSize}` | `{documentFormId, componentId, fileId, uploadUrl}` | PUT | `application/octet-stream` | 결재 서식별 componentId 필요 |
| 6 | Approval imported document attachment | POST | `/business-support/approval/imported-documents/attachments` | presigned URL | `application/json` | `{fileName, fileSize}` | `{fileId, uploadUrl}` | PUT | `application/octet-stream` | 이관 문서용 |
| 7 | Bot attachment | POST | `/bots/{botId}/attachments` | presigned URL | `application/json` | `{fileName}` | `{fileId, uploadUrl}` | PUT | `application/octet-stream` | fileSize 필드 없음 (fileName만 필요) |
| 8 | Bot rich menu image | POST | `/bots/{botId}/richmenus/{richmenuId}/image` | fileId 참조 (업로드 아님) | `application/json` | `{fileId, i18nFileIds?}` | 200 OK (본문 없음) | — | — | 직접 파일 업로드 아님. Bot attachment (#7)로 먼저 업로드 후 fileId 참조 |
| 9 | User photo | POST | `/users/{userId}/photo` | presigned URL | `application/json` | `{fileName, fileSize}` | `{uploadUrl}` | PUT | `application/octet-stream` | fileSize ≤ 10MB, contentType 필드 없음 |
| 10 | Drive file upload | POST | `/users/{userId}/drive/files` 등 | presigned URL | `application/json` | `{fileName, fileSize, modifiedTime?, overwrite?, resume?, suffixOnDuplicate?}` | `{uploadUrl, offset}` | PUT | `application/octet-stream` | fileName ≤ 200자, fileSize ≤ 10GB, resume 지원 |

## 상세 스펙

### 1–2. Board Attachment (게시글/댓글)

- **SDK 참조**: `BoardPostsApi.createPostAttachment()`, `BoardCommentsApi.createCommentAttachment()`
- **Request Model**: `BoardAttachmentUploadUrlRequest(fileName, fileSize, contentType)`
- **Response Model**: `BoardAttachmentUploadUrlResponse(uploadUrl)`
- **제약**: fileName ≤ 200자, 1 ≤ fileSize ≤ 2,147,483,648 (2GB)
- **공식 문서**: https://developers.worksmobile.com/kr/docs/board-post-attachment-create

### 3. Contact Photo (연락처 사진)

- **SDK 참조**: `ContactContactsApi.createPhoto()`
- **Request Model**: `ContactPhotoUploadUrlRequest(fileName, fileSize)`
- **Response Model**: `ContactPhotoUploadUrlResponse(uploadUrl)`
- **제약**: 1 ≤ fileSize ≤ 10,485,760 (10MB)
- **참고**: `contentType` 필드 없음 — 서버가 파일 확장자로 추론

### 4. Note Attachment (노트 첨부)

- **SDK 참조**: `NotePostsApi.createAttachment()`
- **Request Model**: `NoteAttachmentUploadUrlRequest(fileName, fileSize, contentType)`
- **Response Model**: `NoteAttachmentUploadUrlResponse(uploadUrl)`
- **제약**: fileName ≤ 200자, fileSize ≥ 1
- **공식 문서**: https://developers.worksmobile.com/kr/docs/group-note-attachment-create

### 5. Approval User Document Attachment (결재 서식 첨부)

- **SDK 참조**: `ApprovalAttachmentsApi.createUserDocumentAttachment()`
- **Request Model**: `ApprovalUserDocumentAttachmentCreateRequest(documentFormId, componentId, fileName, fileSize)`
- **Response Model**: `ApprovalUserDocumentAttachmentCreateResponse(documentFormId, componentId, fileId, uploadUrl)`
- **참고**: 결재 서식의 `documentFormId`와 첨부 컴포넌트 `componentId`가 추가로 필요

### 6. Approval Imported Document Attachment (이관 문서 첨부)

- **SDK 참조**: `ApprovalAttachmentsApi.createImportedDocumentAttachment()`
- **Request Model**: `ApprovalImportedDocumentAttachmentCreateRequest(fileName, fileSize)`
- **Response Model**: `ApprovalAttachmentUploadUrlResponse(fileId, uploadUrl)`

### 7. Bot Attachment (봇 첨부)

- **SDK 참조**: `BotApi.createAttachment()`
- **Request Model**: `BotAttachmentCreateRequest(fileName)`
- **Response Model**: `BotAttachmentCreateResponse(fileId, uploadUrl)`
- **참고**: `fileSize` 필드 없이 `fileName`만으로 요청
- **공식 문서**: https://developers.worksmobile.com/kr/docs/bot-attachment-create

### 8. Bot Rich Menu Image (리치 메뉴 이미지)

- **SDK 참조**: `BotRichMenuApi.setRichMenuImage()`
- **Request Model**: `BotRichMenuImage(fileId, i18nFileIds?)`
- **Response**: 200 OK (본문 없음)
- **참고**: 실제 파일 업로드가 아님. Bot attachment (#7)로 먼저 이미지를 업로드한 후 받은 `fileId`를 이 endpoint에 JSON으로 전달
- **공식 문서**: https://developers.worksmobile.com/kr/docs/bot-richmenu-image-set

### 9. User Photo (구성원 사진)

- **SDK 참조**: `DirectoryUserProfileApi.createPhoto()`
- **Request Model**: `UserPhotoUploadUrlRequest(fileName, fileSize)`
- **Response Model**: `UserPhotoUploadUrlResponse(uploadUrl)`
- **제약**: 1 ≤ fileSize ≤ 10,485,760 (10MB)
- **참고**: Contact photo와 동일 패턴

### 10. Drive File Upload (드라이브 파일)

- **SDK 참조**: `AbstractDriveFilesApi.createRootUploadUrlTypedInternal()` 등
- **Request Model**: `DriveFileUploadUrlRequest(fileName, fileSize, modifiedTime?, overwrite?, resume?, suffixOnDuplicate?)`
- **Response Model**: `DriveFileUploadUrlResponse(uploadUrl, offset)`
- **제약**: fileName ≤ 200자, 0 ≤ fileSize ≤ 10,737,418,240 (10GB)
- **참고**: `resume=true` 시 이어올리기 지원 (`offset` 반환), **이미 naverworks-cli에 구현됨**
- **Drive 종류별 경로**:
  - 내 드라이브: `POST /users/{userId}/drive/files` 또는 `POST /users/{userId}/drive/files/{folderId}`
  - 공용 드라이브: `POST /drives/{driveId}/files` 또는 `POST /drives/{driveId}/files/{folderId}`
  - 조직/그룹 폴더: `POST /group-folders/{groupFolderId}/files` 또는 `POST /group-folders/{groupFolderId}/files/{folderId}`
  - 초대받은 폴더: `POST /shared-folders/{sharedFolderId}/files` 또는 `POST /shared-folders/{sharedFolderId}/files/{folderId}`

## CLI 구현 시 참고

### 공통 패턴 (presigned URL)

```
[CLI] --JSON POST--> [NAVER WORKS API] --{uploadUrl}--> [CLI] --PUT file--> [presigned URL]
```

1. `application/json`으로 메타데이터 전송
2. 응답에서 `uploadUrl` 추출
3. `uploadUrl`로 파일 바이너리를 PUT (Content-Type: `application/octet-stream`)
4. PUT 시 Authorization 헤더 불필요

### 기존 CLI 코드 재활용

- `client.UploadFile(uploadURL, filePath)` — presigned URL PUT 업로드 (Drive에서 사용 중)
- `client.UploadMultipart(path, fieldName, fileName, data)` — multipart 업로드 (현재 사용처 없음, 향후 필요 시)

### Bot Rich Menu Image 특수 케이스

Bot rich menu image는 2-step이 아니라 3-step:
1. `POST /bots/{botId}/attachments` → `{fileId, uploadUrl}`
2. PUT으로 이미지 파일을 `uploadUrl`에 업로드
3. `POST /bots/{botId}/richmenus/{richmenuId}/image` → `{fileId}` 전달
