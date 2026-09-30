# 메시지방 노트·할 일

메시지방 노트·할 일은 `--room-id`로 지정한다. 방 ID는 채널 UUID이며
웹 내부의 숫자 방 번호와 다르다. `--room-id`를 생략하면 기존 공식 API 명령을 실행한다.

```bash
./naverworks auth login --method browser
./naverworks auth status --method browser
./naverworks note list-posts --room-id "$NW_ROOM_ID" --all
./naverworks task list --room-id "$NW_ROOM_ID" --all
./naverworks task list-categories --room-id "$NW_ROOM_ID"
```

비밀번호와 추가 인증은 전용 Chrome/Chromium 창에서 직접 입력한다.
Chrome 실행 경로를 지정해야 하는 환경에서는 `NW_BROWSER_PATH`를 설정한다.
브라우저를 찾지 못하면 종료 코드 1과 `CLI_ERROR` JSON 오류로 종료한다.
자동으로 다운로드하거나 설치하지 않으며, headless 실행에도 Chrome/Chromium이 필요하다.
기존 API 명령과 방 명령의 미리보기에는 브라우저가 필요하지 않다.
브라우저 세션은 CLI 프로필별로 격리되며 API 토큰과 별도로 저장한다.
일반 브라우저 프로필을 가져오거나 쿠키를 외부 HTTP 클라이언트로 복사하지 않는다.
세션이 만료되면 같은 프로필에서 다시 로그인한다. 한 프로필의 명령은 순서대로 실행한다.
로그인할 때만 브라우저 창이 열리고, 노트·할 일 명령은 항상 headless로 실행한다.
일반 명령에는 화면 서버가 필요하지 않으며 실패 시 창을 띄우는 대체 경로도 없다.

`auth status --method browser`는 웹 로그인 상태를 실제로 확인한다.
`auth logout --method browser`는 선택한 프로필의 로컬 전용 브라우저 세션을 삭제한다.
서버 세션을 철회하는 명령은 아니며 API 토큰과 다른 프로필은 보존한다.
브라우저 실행 파일이 없어도 로컬 로그아웃은 가능하다. 이 프로필의 브라우저와 CLI를
모두 종료한 뒤 실행한다. 사용 중인 잠금 파일이나 소유 정보 불일치를 발견하면 삭제를 중단한다.
`auth status`와 `auth logout`의 기본 대상은 API 토큰이다.

```bash
./naverworks note get-post "$POST_ID" --room-id "$NW_ROOM_ID"
./naverworks note create-post --room-id "$NW_ROOM_ID" --title "제목" --body '<p>본문</p>'
./naverworks note update-post "$POST_ID" --room-id "$NW_ROOM_ID" --title "수정 제목" --body '<p>수정 본문</p>'
./naverworks note delete-post "$POST_ID" --room-id "$NW_ROOM_ID"

./naverworks task get "$TASK_ID" --room-id "$NW_ROOM_ID"
./naverworks task create --room-id "$NW_ROOM_ID" --title "제목" --description "설명"
./naverworks task update "$TASK_ID" --room-id "$NW_ROOM_ID" --description "수정 설명"
./naverworks task complete "$TASK_ID" --room-id "$NW_ROOM_ID"
./naverworks task incomplete "$TASK_ID" --room-id "$NW_ROOM_ID"
./naverworks task delete "$TASK_ID" --room-id "$NW_ROOM_ID"
```

노트 명령에서는 `groupId` 인수를 생략한다. 노트 ID는 정밀도 손실을 막기 위해
문자열로 출력한다. 새 노트는 공지로 지정하지 않으며 알림 옵션은 끈다.
수정 시 기존 첨부파일·공지·공동 편집 설정을 보존하고 수정 시각이 달라졌으면 중단한다.
첨부파일 추가·삭제, 노트 검색·부분 수정 명령에는 아직 메시지방 옵션이 없다.

할 일 생성 기본 요청자·담당자는 로그인 사용자이며 완료 조건은 `ANY_ONE`이다.
`--assignor-id`, `--assignee-ids`, `--completion-condition ANY_ONE|MUST_ALL`로 변경할 수 있다.
이 ID는 웹 내부 사용자 번호다. `--due-date YYYY-MM-DD`는 마감일이며 빈 문자열로 지울 수 있다.
`--data`에는 `title`, `content`, `dueDate`, `assignorId`, `assigneeIdList`, `taskStatusOption`만 허용한다.
`--data`와 개별 입력 플래그는 함께 사용할 수 없다.

완료·미완료 명령은 기본적으로 내 담당 상태를 바꾼다. 모든 담당자의 상태를 바꾸려면
`--all-assignees`를 명시한다. 목록은 방의 모든 카테고리를 순회하며
`--category-id`, `--status TODO|DONE`, `--count 1..100`, `--all`을 지원한다.
첫 페이지에서는 일부 카테고리만 반환될 수 있으므로 방 전체 조회에는 `--all`을 사용한다.
`list-categories`도 `--count`, `--cursor`, `--all`, `--output table`을 지원한다.
카테고리는 서버 목록을 조회한 뒤 로컬에서 페이지로 나눈다. 할 일과 카테고리 커서는
각 목록에서 반환된 값을 그대로 사용하고 다른 방이나 다른 목록에 재사용하지 않는다.
`--all`과 `--cursor`를 함께 지정하면 해당 커서부터 순회한다. 기본 페이지 크기는 20이다.
카테고리 생성·수정·삭제와 할 일 검색에는 아직 메시지방 옵션이 없다.

`--dry-run`, `--plan-out`, `--generate-input`은 브라우저를 시작하지 않는다.
방 번호·저장 위치·프로젝트·로그인 사용자·생성될 노트 ID는 실행 시 조회해야 하므로
미리보기에는 `${...}` 자리표시자와 `resolution_required`를 포함한다.
수정에 필요한 기존 필드도 실제 실행 시 조회한다.
미리보기의 `headless: true`와 `headers_template`에서도 실행 방식과 미해결 사용자 번호를 확인할 수 있다.

이 기능은 공식 API가 보장하는 계약 대신 웹에서 확인한 내부 요청을 사용한다.
서비스 변경이나 권한 변경으로 실패할 수 있다. 실패한 쓰기는 자동 재시도하지 않는다.
결과를 확인하지 못했다는 오류가 나오면 웹에서 반영 여부를 먼저 확인한다.
노트·할 일이 활성화되지 않은 방에서는 자동으로 공간을 생성하지 않는다.

2026-09-30 Linux 실제 로그인 계정에서 조회·페이지네이션·상태 필터와 임시 항목의
생성·수정·완료·미완료·삭제를 확인했다. 화면 환경 변수 없이 headless로 수정·상태 변경·삭제와
재조회를 통과했고, 기존 노트 5개·할 일 4개의 조회 결과가 그대로임을 확인했다.
할 일 생성은 생성 전후 목록에서 기존 ID를 제외하고 입력과 일치하는 새 항목 1개를
확인한 뒤 재조회한다. 일치 항목이 없거나 2개 이상이면 결과 불명으로 중단한다.
이 처리로 보완한 생성 명령도 추가 승인된 임시 할 일 1개의 headless 생성·조회·삭제에서
정상 종료했고, 기존 항목이 그대로임을 다시 확인했다. Windows Chrome·ACL 실행은 미검증이다.

후속 점검에서 카테고리 페이지 옵션·표 출력과 시작 커서 이후 전체 순회를 보완했다.
실제 서비스에서 headless 인증 상태, 카테고리 1개 조회·나머지 2개 순회·전체 3개·표 출력,
노트 시작 위치 3 이후 3개 조회를 확인했다. 이 점검은 조회만 실행했다.
로컬 로그아웃은 격리된 임시 디렉터리의 실제 파일 삭제로 검증했으며,
사용 중인 세션·심볼릭 링크·다른 프로필의 소유 정보는 거부하고 API 토큰은 보존했다.
실제 로그인 프로필에는 로그아웃을 실행하지 않았다.
