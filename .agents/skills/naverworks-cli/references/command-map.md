# 주요 API 명령과 인증 기준

현재 명령 구조는 `naverworks <domain> --help`가 최종 기준이다. 이 문서는 검색과
채널 폴더처럼 인증 방식이 혼동되기 쉬운 기능만 요약한다.

## 검색·조회

| 기능 | 명령 | 인증과 최소 scope |
|---|---|---|
| 사용자 검색 | `directory search-users <query>` | OAuth 또는 JWT; `user.read` 또는 `directory.read` |
| 그룹 검색 | `directory search-groups <query>` | OAuth 또는 JWT; `group.read` 또는 `directory.read` |
| 조직 검색 | `directory search-orgunits <query>` | OAuth 또는 JWT; `orgunit.read` 또는 `directory.read` |
| 사용자 소속 | `directory list-user-groups|list-user-orgunits <userId>` | 해당 Directory read scope |
| 게시판 검색 | `board search-posts <query>` | OAuth 또는 JWT; `board.read` |
| 노트 검색 | `note search-posts <groupId> <query>` | 구성원 OAuth; `group.note.read` |
| 캘린더 검색 | `calendar search-events [query]` | OAuth 또는 JWT; `calendar.read` |
| 연락처 검색 | `contact search <query>` | OAuth 또는 JWT; `contact.read` |
| Task 검색 | `task search [query]` | 구성원 OAuth; `task.read` |
| 관리자 결재 조회 | `approval list-all --from <date> --until <date>` | OAuth 또는 JWT, 관리자 권한; `businessSupport.approval.read` |
| Drive 검색 | `drive search <query>` | 구성원 OAuth; `file.read`, 채널 폴더 포함 시 `group.folder.read` |
| 메시지 콘텐츠 URL | `monitoring download-messages ... [--channel-id <id>]` | 관리자 또는 JWT Service Account; `monitoring.read` |
| 메일 조회 | `mail list|get|list-folders|get-folder ...` | 구성원 OAuth; `mail.read` |
| 메일 전송·변경 | `mail send|delete|move|update ...` | 구성원 OAuth; `mail` |

`approval list-all`의 기간은 최대 1개월이며 `--type`은
`pending|upcoming|approved|completed`다. 검색 명령의 `--query-filters`,
`--order-by`, 날짜 형식은 해당 명령의 `--help`로 검증한다.

## 채널 폴더

모든 채널 폴더 명령은 구성원 OAuth Access Token 전용이다. JWT Service Account를
사용하면 CLI가 네트워크 요청 전에 거부해야 한다.

### Read

최소 `file.read group.folder.read` scope가 필요하다.

| 목적 | 명령 |
|---|---|
| 채널 폴더 목록·속성 | `drive channel list`, `drive channel get <channelFolderId>` |
| 파일 목록·속성 | `drive channel files <channelFolderId>`, `drive channel get-file <channelFolderId> <fileId>` |
| 다운로드 URL | `drive channel download <channelFolderId> <fileId>` |
| 버전 | `drive channel revision list|get|download ...` |
| 링크 설정·링크 | `drive channel link-setting <channelFolderId>`, `drive channel link get ...` |
| 휴지통 목록 | `drive channel trash-list <channelFolderId>` |
| 권한 목록·상세 | `drive channel permission list|get ...` |

`files`, `revision list`, `trash-list`는 `--cursor`, `--count`, `--all`을 지원한다.
채널 목록과 권한 목록은 공식 API에 페이지네이션이 없으므로 임의의 `--all`을
추가하지 않는다.

### Write

최소 `file group.folder` scope가 필요하며 exact command 승인을 받은 뒤 실행한다.

| 목적 | 명령군 |
|---|---|
| 파일·폴더 | `upload`, `mkdir`, `delete`, `copy`, `rename`, `move` |
| 파일 상태 | `protect`, `unprotect`, `lock`, `unlock` |
| 버전·휴지통 | `revision restore`, `trash-restore`, `trash-delete` |
| 공유 링크 | `link create|update|delete` |
| 접근 권한 | `permission create|update|delete|delete-all|enable|disable` |

업로드 전에 로컬 파일 경로와 크기, 대상 폴더를 확인한다. `trash-delete`,
`permission delete-all`, 일반 `delete`는 복구 가능성을 별도로 명시한다.

## 안전한 예시

```bash
# 실제 조회
naverworks --profile member drive channel list --output table

# 쓰기 전 요청 확인만 수행
naverworks --profile member --dry-run drive channel delete \
  CHANNEL_FOLDER_ID FILE_ID
```

전체 사용 예시는 [README](../../../../README.md)와
[Domain Command Guide](../../../../docs/wiki/Domain-Command-Guide.md)를 참고한다.
도메인별 OAuth/JWT 지원 여부와 scope는
[Auth Identity Matrix](../../../../docs/auth-identity-matrix.md)를 대조한다.
