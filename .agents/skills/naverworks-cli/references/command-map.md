# 주요 리소스 명령과 인증 기준

현재 명령 구조는 `naverworks <domain> --help`가 최종 기준이다. 이 문서는 검색과
채널 폴더·메시지방 웹 기능처럼 인증 방식이 혼동되기 쉬운 기능을 요약한다.

## 메시지방 노트·할 일 (`v0.8.0`부터)

모든 방 명령에 `--room-id <channelUUID>`를 붙인다. 채널 UUID는 방 서랍의
'채널 ID'이며 웹 내부 숫자 방 번호나 공식 API의 `groupId`와 다르다.
이 옵션이 있으면 프로필의 웹 로그인 세션으로 실행하고, 생략하면 기존 공식 API를 사용한다.
OAuth/JWT 토큰 scope가 아니라 로그인 사용자의 방 접근 권한으로 실행하므로
API 인증과 혼동하지 않는다. 인증은 `naverworks-profile` 스킬로 처리한다.

| 목적 | 명령 (`--room-id <channelUUID>` 필수) | 분류 |
|---|---|---|
| 노트 목록 | `note list-posts` | read |
| 노트 상세 | `note get-post <postId>` | read |
| 노트 생성 | `note create-post --title <title> --body <body>` | write |
| 노트 수정 | `note update-post <postId> --title <title> --body <body>` | write |
| 노트 삭제 | `note delete-post <postId>` | write |
| 할 일 목록·카테고리 | `task list`, `task list-categories` | read |
| 할 일 상세 | `task get <taskId>` | read |
| 할 일 생성 | `task create --title <title> --description <description>` | write |
| 할 일 수정 | `task update <taskId> --description <description>` | write |
| 할 일 완료 | `task complete <taskId>` | write |
| 할 일 미완료 | `task incomplete <taskId>` | write |
| 할 일 삭제 | `task delete <taskId>` | write |

노트 명령에서는 `groupId` 위치 인수를 생략하며 `postId`는 정밀도 손실을 막기 위해
문자열 그대로 사용한다. 할 일 ID는 `t-UUID` 형식이다. 생성 기본 요청자·담당자는
로그인 사용자, 완료 조건은 `ANY_ONE`이며 노트·할 일 생성의 알림 옵션은 꺼져 있다.
`--assignor-id`·`--assignee-ids`에는 웹 내부 사용자 번호를 사용한다.
`--room-id`와 API 전용 `--user-id`·`--search-filter-type`은 함께 사용할 수 없다.

완료·미완료는 기본적으로 내 담당 상태만 바꾼다. `--all-assignees`는 모든 담당자의
상태를 변경하므로 사용자의 승인 범위에 포함된 경우에만 명시한다.
실행 후 `task get`으로 전체 `taskStatus`와 `assigneeList`의 담당 상태를 함께 확인한다.

노트·할 일·카테고리 목록은 `--count 1..100`(기본 20), `--cursor`, `--all`,
`--output table`을 지원한다. 할 일은 `--category-id`·`--status TODO|DONE`으로
필터링하며 첫 페이지가 방 전체를 나타내지 않을 수 있다. 빈 페이지에도
`responseMetaData.nextCursor`가 있으면 계속 순회한다. `--all --cursor <cursor>`는
지정 위치부터 순회하고, 할 일과 카테고리 커서는 반환된 값 그대로 사용하며
다른 방이나 다른 목록의 커서와 섞지 않는다.

Chrome/Chromium이 필요하며 실행 파일은 `NW_BROWSER_PATH`로 지정할 수 있다.
로그인할 때만 창이 열리고 상태 확인·방 명령은 항상 headless다. 같은 프로필의
브라우저 명령은 순서대로 실행한다. 브라우저가 없으면 종료 코드 1의 `CLI_ERROR`로
종료하고 자동 설치하지 않는다. 방 명령의 `--dry-run`·`--generate-input`은
네트워크나 브라우저 없이 가능하며 미해결 값은 `${...}`와 `resolution_required`로 표시한다.

웹 내부 요청은 서비스 변경에 영향을 받을 수 있다. 브라우저 쓰기는 자동 재시도하지
않고 결과 불명 시 get/list로 반영 여부부터 확인한다. 노트·할 일이 활성화되지 않은
방을 자동으로 초기화하지 않는다. 노트 첨부파일 추가·삭제·검색·부분 수정,
할 일 검색·카테고리 생성·수정·삭제는 아직 방 옵션을 지원하지 않는다.

```bash
# 실제 조회: 동일 프로필의 auth status --method browser로 로그인 상태부터 확인
naverworks --profile default note list-posts --room-id "$NW_ROOM_ID" --count 2
naverworks --profile default task list --room-id "$NW_ROOM_ID" --all --status TODO

# 완료 요청의 미리보기만 수행; 실제 완료 실행에는 해당 write 승인이 필요
naverworks --profile default task complete "$TASK_ID" --room-id "$NW_ROOM_ID" --dry-run
```

세션과 입력 옵션의 상세 동작은
[메시지방 노트·할 일](../../../../docs/browser-room-management.md)을 참조한다.

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
| 메시지방 단건 조회 | `bot get-channel <channelId>` | OAuth 또는 JWT; `bot`, `bot.read` 또는 `bot.message` |

공식 Bot API에 메시지방 목록·이름 검색은 없다. `channelId`는 메시지방 서랍 메뉴의
'채널 ID' 또는 봇 콜백 `source.channelId`에서 확인한다. `directory search-groups`와
`directory search-orgunits`는 그룹·조직 방만 찾고, 일반(`MULTI_USERS`) 방은 이름만으로
조회할 수 없다. `drive channel list`는 드라이브 메시지방 폴더 목록이며 채팅방 목록이
아니다.

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
