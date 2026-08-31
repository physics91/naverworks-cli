# naverworks-cli 저장소 현대화 및 API 커버리지 요구사항

- Status: Finalized
- Grade: M
- Quality Validation: Passed

## 개정 이력

| 버전 | 날짜 | 작성자 | 변경 내용 |
|---|---|---|---|
| 1.0 | 2026-08-31 | Codex / 저장소 관리자 | 확정 요구사항 및 품질 검증 결과 |

---

## 1. 개요

### 1.1 목적

장기간 관리 공백이 있었던 `naverworks-cli` 저장소를 현재 지원되는 Go 툴체인과 GitHub Actions 기준으로 복구하고, 미배포된 보안·안정성 수정사항을 릴리스 가능한 상태로 만든다. 이후 NAVER WORKS 2026년 7월 API 변경 중 공식 Developers 문서에 등재된 미지원 기능을 기존 CLI 구조와 계약에 맞게 추가한다.

### 1.2 배경 및 근거

- 사용자 진술: 저장소를 장기간 관리하지 못해 현재 상태를 일괄 점검하고 정상 관리 상태로 복구할 필요가 있다. 증거 강도는 `interview-reported`다.
- 로컬 점검 결과:
  - `go.mod` 기준은 `go1.25.0`이며, 공개 `v0.4.1` Linux 바이너리는 `go1.25.9`로 빌드됐다.
  - `develop`은 `main`보다 13커밋, `v0.4.1`보다 33커밋 앞선다.
  - 직접 Go 모듈은 최신이지만 GitHub Actions 메이저 버전과 한 테스트 파일의 포맷이 오래됐다.
  - Weekly Health 오류 이슈가 정상 회복 후에도 열린 상태로 남는다.
  - NAVER WORKS API 변경 이슈 `#27`의 미지원 항목은 현재 공식 문서에 등재됐다.
- 외부 근거:
  - [Go 릴리스 기록](https://go.dev/doc/devel/release)
  - [Go 1.27 릴리스 노트](https://go.dev/doc/go1.27)
  - [NAVER WORKS API](https://developers.worksmobile.com/kr/docs/api)
  - [Board API](https://developers.worksmobile.com/kr/docs/board)
  - [Note API](https://developers.worksmobile.com/kr/docs/note)
  - [Directory API](https://developers.worksmobile.com/kr/docs/directory)
  - [Drive API](https://developers.worksmobile.com/kr/docs/drive/?lang=ko)

### 1.3 범위

#### 포함

1. 1차 유지보수 릴리스 단위
   - Go 1.26.7 기준 전환
   - GitHub Actions SHA 핀 갱신
   - 포맷·취약점·운영 이슈 수명주기 자동화
   - 기존 `develop` 미배포 변경의 검증 및 로컬 병합 준비
2. 2차 API 기능 릴리스 단위
   - 이슈 `#27`의 검색 API, 구성원 소속 조회, 채널 폴더 전체 공개 API
   - Drive·Monitoring·Directory·Business Support 추가 필드와 파라미터
   - 명령, API 계층, 테스트, 문서 및 커버리지 근거
3. 두 릴리스 단위를 구분하는 검증된 Git 체크포인트 보존

#### 제외

- 실제 NAVER WORKS 테넌트를 대상으로 한 API 호출
- Go 1.27 전환 및 macOS 최소 지원 버전 변경
- 원격 push, 태그 생성, GitHub Release 생성, npm publish
- CLI가 수신 서버가 되어야 하는 Bot Begin/End callback 구현
- 이번 요구사항과 관계없는 구조 개편이나 광범위한 리팩터링
- 새 런타임 의존성 추가

### 1.4 성공 지표

| 지표 | 기준선 | 목표 | 확인 시점 | 소유자 | 측정 근거 |
|---|---|---|---|---|---|
| Go 기준 툴체인 | `go1.25.0`; 공개 바이너리 `go1.25.9` | `go1.26.7` | 1차 체크포인트 | 저장소 관리자 | `go.mod`, `go version -m` |
| 포맷 편차 | `gofmt -l` 1파일 | 0파일 | 각 체크포인트 | 저장소 관리자 | CI 포맷 검사 |
| 도달 가능한 Go 취약점 | 자동 검사 없음 | 0건 | 각 체크포인트 | 저장소 관리자 | `govulncheck ./...` |
| GitHub Action 핀 | v4~v7 혼재 | 합의된 v7 전체 SHA 핀 | 1차 체크포인트 | 저장소 관리자 | workflow 정적 검사 |
| Weekly Health 복구 이슈 | 정상 회복 후 열린 `ops-error` 잔존 | 자동화 소유 이슈만 자동 종료 | 1차 체크포인트 | 저장소 관리자 | helper 단위 테스트, workflow 테스트 |
| 이슈 `#27` API 지원 | 검색·소속·채널 폴더 등 미지원 | 합의된 공식 엔드포인트 100% 매핑 | 2차 체크포인트 | 저장소 관리자 | API 커버리지 매트릭스, smoke test |
| 회귀 검증 | 현재 전체 테스트 통과 | Linux 로컬 및 Windows CI 통과 | 각 체크포인트 | 저장소 관리자 | test/vet/build 결과 |

### 1.5 용어

| 용어 | 정의 |
|---|---|
| 유지보수 체크포인트 | Go·CI·보안·운영 자동화만 포함하는 첫 번째 릴리스 가능 커밋 경계 |
| API 체크포인트 | 이슈 `#27` 기능을 포함하는 두 번째 릴리스 가능 커밋 경계 |
| 채널 폴더 | 조직·그룹·메시지방에 연결된 NAVER WORKS Drive 폴더 리소스 |
| 도달 가능한 취약점 | `govulncheck`가 애플리케이션 호출 경로에서 사용된다고 판정한 취약점 |
| 외부 쓰기 | push, 태그, 이슈 상태 변경, Release 생성, npm publish 등 원격 상태를 바꾸는 작업 |

---

## 2. 이해관계자와 사용자

### 2.1 이해관계자

| 이해관계자 | 역할 | 관심사 |
|---|---|---|
| 저장소 관리자 | 범위·버전·외부 쓰기 승인 | 안전한 복구, 릴리스 가능 상태, 관리 자동화 |
| CLI 사용자 | NAVER WORKS 명령 실행 | 기존 호환성, 새 API 접근, 안정된 배포본 |
| 릴리스 운영자 | GitHub/npm 배포 | 재현 가능한 빌드, 체크포인트, 롤백 가능성 |
| GitHub Actions | 자동 검증·모니터링 | 최소 권한, 고정된 도구, 정확한 이슈 수명주기 |

### 2.2 사용자 유형

| 사용자 유형 | 특성 | 기술 수준 | 핵심 목표 |
|---|---|---|---|
| CLI 운영 사용자 | 스크립트와 터미널에서 NAVER WORKS API 사용 | 중급 이상 | 안정된 명령과 기계 판독 가능한 출력 |
| 저장소 관리자 | Go·GitHub Actions·릴리스 관리 | 고급 | 안전한 변경 검증과 단계별 릴리스 |

---

## 3. 기능 요구사항

### 3.1 1차 유지보수 릴리스 단위

| REQ-ID | 요구사항 | EARS 유형 | 비고 |
|---|---|---|---|
| REQ-001 | 저장소는 Go 언어 및 빌드 기준을 `go1.26.7`로 통일해야 한다. | Ubiquitous | `go.mod`, CI, release, scheduled workflows, 문서 포함 |
| REQ-002 | GitHub Actions가 실행될 때, 시스템은 `actions/checkout@v7.0.1`, `actions/setup-go@v7.0.0`, `actions/upload-artifact@v7.0.1`, `actions/setup-node@v7.0.0`, `goreleaser/goreleaser-action@v7.2.3`을 검증된 전체 커밋 SHA로 참조해야 한다. | Event-driven | 사용 중인 Action만 적용 |
| REQ-003 | Go 소스 검증이 실행될 때, 시스템은 `gofmt -l` 결과가 비어 있지 않으면 검증을 실패시켜야 한다. | Event-driven | 현재 `internal/api/service_test.go` 편차 포함 |
| REQ-004 | CI와 릴리스 사전 검증이 실행될 때, 시스템은 `golang.org/x/vuln/cmd/govulncheck@v1.7.0`으로 `./...`을 검사해야 한다. | Event-driven | 도구 버전 고정 |
| REQ-005 | 취약점 검사에서 애플리케이션 호출 경로에 도달 가능한 취약점이 발견되면, 시스템은 해당 검증과 후속 병합·릴리스 단계를 실패시켜야 한다. | Unwanted | 비도달 정보성 항목은 보고하되 자동 실패 기준과 구분 |
| REQ-006 | Weekly Health가 모든 필수 검사에서 성공하면, 시스템은 자동화가 소유한 정확한 제목과 라벨의 열린 `ops-error` 이슈만 닫아야 한다. | Event-driven | unrelated 이슈 변경 금지 |
| REQ-007 | Weekly Health가 하나 이상의 필수 검사에서 실패하면, 시스템은 기존 동일 제목의 열린 이슈를 갱신하거나 새 이슈를 하나만 생성해야 한다. | Unwanted | 중복 생성 금지 |
| REQ-008 | API 변경 공지가 CLI에 영향을 주지 않는 것으로 판정되면, 시스템은 해당 slug를 baseline에 추가하고 로컬 검증 근거를 남겨야 한다. | Event-driven | `#28`, `#30` 대상 |
| REQ-009 | 이슈 `#27` 기능이 완료되지 않은 동안, 시스템은 해당 slug와 이슈를 미완료 추적 대상으로 유지해야 한다. | State-driven | baseline 조기 종료 금지 |
| REQ-010 | 유지보수 체크포인트를 만들 때, 시스템은 `develop`의 기존 미배포 변경과 유지보수 변경을 하나의 검증 가능한 선형 Git 이력에 보존해야 한다. | Event-driven | 비강제 작업만 허용 |
| REQ-011 | 유지보수 체크포인트가 검증되면, 시스템은 체크포인트 커밋, 변경 내역, 권장 버전, 검증 결과 및 롤백 절차를 기록해야 한다. | Event-driven | 외부 쓰기 전 handoff |
| REQ-012 | 외부 쓰기 승인이 없는 동안, 시스템은 push, 태그 생성, 원격 이슈 상태 변경, GitHub Release 및 npm publish를 수행하지 않아야 한다. | State-driven | 사용자 선택 `a` 반영 |

### 3.2 검색 및 구성원 소속 API

| REQ-ID | 요구사항 | EARS 유형 | 비고 |
|---|---|---|---|
| REQ-013 | 시스템은 공식 문서의 Board 게시글 검색 API를 CLI 명령으로 제공해야 한다. | Ubiquitous | 모든 게시판 검색 |
| REQ-014 | 시스템은 공식 문서의 조직·그룹 Note 게시글 검색 API를 CLI 명령으로 제공해야 한다. | Ubiquitous | `groupId` 경계 유지 |
| REQ-015 | 시스템은 공식 문서의 Calendar 일정 검색 API를 CLI 명령으로 제공해야 한다. | Ubiquitous | 사용자 식별 규칙 유지 |
| REQ-016 | 시스템은 공식 문서의 Contact 검색 API를 CLI 명령으로 제공해야 한다. | Ubiquitous | 기존 contact 출력 규칙 유지 |
| REQ-017 | 시스템은 공식 문서의 User 검색 API를 CLI 명령으로 제공해야 한다. | Ubiquitous | Directory 인증 경계 유지 |
| REQ-018 | 시스템은 공식 문서의 OrgUnit 검색 API를 CLI 명령으로 제공해야 한다. | Ubiquitous | Directory 인증 경계 유지 |
| REQ-019 | 시스템은 공식 문서의 Group 검색 API를 CLI 명령으로 제공해야 한다. | Ubiquitous | Directory 인증 경계 유지 |
| REQ-020 | 시스템은 공식 문서의 Task 검색 API를 CLI 명령으로 제공해야 한다. | Ubiquitous | 사용자 식별 규칙 유지 |
| REQ-021 | 시스템은 공식 문서의 관리자 Approval 문서 검색 API를 CLI 명령으로 제공해야 한다. | Ubiquitous | 관리자 scope 명시 |
| REQ-022 | 시스템은 구성원의 소속 그룹 목록과 소속 조직 목록 조회 API를 각각 CLI 명령으로 제공해야 한다. | Ubiquitous | `/users/{userId}/groups`, `/users/{userId}/orgunits` |
| REQ-023 | 목록 또는 검색 응답에 페이지네이션이 정의된 경우, 시스템은 기존 `--count`, `--cursor`, `--all` 동작을 적용해야 한다. | Optional | 공식 API 계약이 지원하는 범위 |

### 3.3 채널 폴더 API

| REQ-ID | 요구사항 | EARS 유형 | 비고 |
|---|---|---|---|
| REQ-024 | 시스템은 2026-08-31 공식 Drive 문서의 채널 폴더 섹션에 공개된 엔드포인트를 누락 없이 API 커버리지 매트릭스에 기록해야 한다. | Ubiquitous | 문서 기준선 고정 |
| REQ-025 | 시스템은 채널 폴더 목록 조회와 속성 조회 명령을 제공해야 한다. | Ubiquitous | 생성·삭제는 공식 API 미지원 |
| REQ-026 | 시스템은 채널 폴더의 루트 및 하위 파일 목록, 파일 속성, 업로드 URL, 다운로드, 삭제, 복사, 이름 변경, 이동 명령을 제공해야 한다. | Ubiquitous | 파일 크기·URL 보안 규칙 유지 |
| REQ-027 | 시스템은 채널 폴더의 루트 및 하위 폴더 생성 명령을 제공해야 한다. | Ubiquitous | 공식 createfolder 계약 사용 |
| REQ-028 | 시스템은 채널 폴더 파일의 잠금·잠금 해제와 중요 표시·해제 명령을 제공해야 한다. | Ubiquitous | 공식 제공 작업만 포함 |
| REQ-029 | 시스템은 채널 폴더 파일 버전의 목록·속성·복원·다운로드 명령을 제공해야 한다. | Ubiquitous | revision 식별자 검증 |
| REQ-030 | 시스템은 채널 폴더 링크의 설정·속성·생성·수정·삭제 명령을 제공해야 한다. | Ubiquitous | 기존 Drive 링크 패턴 재사용 |
| REQ-031 | 시스템은 채널 폴더 휴지통의 목록·복원·영구 삭제 명령을 제공해야 한다. | Ubiquitous | destructive 동작 명확화 |
| REQ-032 | 시스템은 채널 폴더 접근 권한의 생성·목록·조회·수정·개별 해제·전체 해제·허용·미허용 명령을 제공해야 한다. | Ubiquitous | 공식 권한 계약 적용 |
| REQ-033 | 공식 기준선과 구현 커버리지 매트릭스가 불일치하면, 시스템은 API 체크포인트 완료를 차단해야 한다. | Unwanted | 누락·과잉 구현 모두 보고 |

### 3.4 추가 파라미터와 응답 계약

| REQ-ID | 요구사항 | EARS 유형 | 비고 |
|---|---|---|---|
| REQ-034 | Drive 검색 명령이 실행될 때, 시스템은 선택적 `queryFilters`를 공식 허용 형식으로 전달해야 한다. | Event-driven | 파일명·본문 검색 범위 |
| REQ-035 | Monitoring 메시지 콘텐츠 다운로드가 실행될 때, 시스템은 선택적 `channelId`를 전달할 수 있어야 한다. | Event-driven | 미지정 시 기존 동작 유지 |
| REQ-036 | 구성원 근무 상태 응답에 `delegates`가 존재하면, 시스템은 해당 필드를 손실 없이 출력해야 한다. | Optional | passthrough 계약 테스트 |
| REQ-037 | 관리자 결재 문서 목록이 실행될 때, 시스템은 선택적 `type` 필터를 전달할 수 있어야 한다. | Event-driven | 미지정 시 기존 동작 유지 |

### 3.5 CLI 통합과 문서

| REQ-ID | 요구사항 | EARS 유형 | 비고 |
|---|---|---|---|
| REQ-038 | 신규 API 명령을 추가할 때, 시스템은 기존 `cmd/<domain>.go`와 `internal/api/<domain>.go` 파일 쌍 또는 기존 Drive 세부 모듈 구조를 따라야 한다. | Event-driven | 기술 제약으로 허용 |
| REQ-039 | 신규 명령이 등록될 때, 시스템은 공식 인증 주체와 scope 제약을 도움말 및 사용자 문서에 표시해야 한다. | Event-driven | OAuth/JWT 우회 금지 |
| REQ-040 | 신규 목록 명령이 table 출력을 지원할 수 있는 기존 공통 구조와 일치하면, 시스템은 JSON과 table 출력 계약을 모두 제공해야 한다. | Optional | 기존 도메인 패턴 우선 |
| REQ-041 | 신규 쓰기 또는 다운로드 명령이 실행될 때, 시스템은 기존 `--dry-run`, `--plan-out`, 입력 크기 제한 및 민감 URL 마스킹 규칙을 유지해야 한다. | Event-driven | 적용 가능한 명령만 |
| REQ-042 | 신규 명령이 구현되면, 시스템은 API 경로 계약 테스트, CLI 등록·플래그 smoke test 및 대표 실패 경로 테스트를 추가해야 한다. | Event-driven | 실제 API 호출 없음 |
| REQ-043 | API 체크포인트가 검증되면, 시스템은 유지보수 체크포인트 이후의 별도 커밋 경계와 별도 릴리스 노트를 보존해야 한다. | Event-driven | 두 릴리스 단위 구분 |

---

## 4. 비기능 요구사항

### 4.1 성능 및 네트워크

| REQ-ID | 요구사항 | 측정 기준 |
|---|---|---|
| NFR-001 | 단일 페이지 조회 명령은 인증 갱신과 기존 재시도 정책을 제외하고 하나의 업무 API 요청만 수행해야 한다. | `httptest` 요청 횟수 |
| NFR-002 | `--all` 조회는 기존 cursor 반복 감지와 종료 조건을 유지해야 한다. | 페이지네이션 회귀 테스트 |

### 4.2 보안

| REQ-ID | 요구사항 | 측정 기준 |
|---|---|---|
| NFR-003 | 시스템은 자격 증명, 토큰, presigned URL 및 민감 요청 본문을 로그·오류·dry-run 출력에 노출하지 않아야 한다. | 보안 회귀 테스트, 출력 스냅샷 |
| NFR-004 | 시스템은 기존 인증 방식, 파일 권한, 응답 크기 제한 및 원자적 설정 저장 보안 계약을 유지해야 한다. | 기존 보안 테스트 전체 통과 |
| NFR-005 | 신규 다운로드 명령은 기존 허용 호스트와 HTTPS 검증 경계를 우회하지 않아야 한다. | URL 검증 및 다운로드 테스트 |

### 4.3 품질 및 호환성

| REQ-ID | 요구사항 | 측정 기준 |
|---|---|---|
| NFR-006 | 각 체크포인트는 `go mod tidy -diff`, `gofmt`, `go vet`, 계층별 테스트, 전체 테스트, 빌드 및 취약점 검사를 통과해야 한다. | 명령 종료 코드 0 |
| NFR-007 | 시스템은 Linux 로컬 검증과 GitHub Actions Windows 검증을 모두 통과해야 한다. | 두 플랫폼 CI 성공 |
| NFR-008 | 시스템은 기존 명령명, 플래그, stdout JSON/table 및 stderr JSON 오류 계약을 깨지 않아야 한다. | 기존 smoke/meta/journey 테스트 |
| NFR-009 | 시스템은 새 런타임 의존성을 추가하지 않아야 한다. | `go.mod` direct runtime dependency diff |
| NFR-010 | 시스템은 Go 1.27 전용 문법이나 표준 라이브러리 API를 사용하지 않아야 한다. | `go1.26.7` 빌드·vet |

### 4.4 운영

| REQ-ID | 요구사항 | 측정 기준 |
|---|---|---|
| NFR-011 | 자동화는 정확한 저장소, 이슈 제목, 라벨 및 상태를 검증한 후에만 이슈를 생성·갱신·종료해야 한다. | GitHub API helper 단위 테스트 |
| NFR-012 | 검증이 실패하면, 시스템은 실패 원인과 재현 명령을 기록하고 후속 체크포인트·병합·릴리스를 차단해야 한다. | 실패 시나리오 테스트와 실행 보고서 |
| NFR-013 | 모든 개발 쓰기는 단일 파이프라인 worktree의 비보호 브랜치에서 수행해야 한다. | worktree binding 및 Git 상태 |

---

## 5. 외부 인터페이스

| 인터페이스 | 대상 시스템 | 프로토콜 | 설명 |
|---|---|---|---|
| IF-001 | NAVER WORKS REST API v1.0 | HTTPS/JSON | 검색, Directory, Drive, Monitoring, Business Support 기능 |
| IF-002 | GitHub Actions | Workflow YAML | CI, 릴리스, API monitor, Weekly Health |
| IF-003 | GitHub Issues API | HTTPS/JSON | 자동화 소유 이슈 생성·갱신·종료 |
| IF-004 | Go vulnerability database | HTTPS / `govulncheck` | 호출 경로 기반 취약점 검사 |
| IF-005 | GitHub Releases 및 npm | HTTPS/OIDC | 이번 실행에서는 읽기·준비만 허용; 쓰기는 별도 승인 필요 |

---

## 6. 데이터, 제약 및 가정

### 6.1 데이터

- 프로필 설정, OAuth/JWT 토큰, 사용자 식별자 및 API 응답에는 자격 증명·PII가 포함될 수 있다.
- 신규 API 구현은 기존 passthrough JSON 모델을 우선 사용하고, 테스트에는 합성 데이터만 사용한다.
- 실제 테넌트 데이터는 수집·저장하지 않는다.

### 6.2 제약

- Go 1.26.7, Cobra, 기존 `cmd/` ↔ `internal/api/` 구조를 유지한다.
- 한국어 사용자 메시지와 영어 코드·변수명 규칙을 유지한다.
- 오류는 stderr JSON 계약을 유지한다.
- 새 명령은 `cmd/smoke_test.go` 등록 검증을 포함한다.
- 개발과 계획은 단일 파이프라인 worktree/branch를 사용한다.
- primary `develop`/`main`에는 기능 개발 쓰기를 하지 않는다.
- 원격 상태 변경은 별도 승인 없이는 수행하지 않는다.

### 6.3 가정

| 가정 | 수용자 | 관련 요구사항 | 영향 |
|---|---|---|---|
| 2026-08-31 공식 NAVER WORKS Developers 문서를 API 범위 기준선으로 사용한다. | 저장소 관리자 | REQ-013~REQ-033 | 문서 변경 시 매트릭스 재검토 필요 |
| GitHub-hosted runner는 Action v7의 Node 24 런타임을 지원한다. | 저장소 관리자 | REQ-002 | runner 비호환 시 v6 유지 검토 |
| 실제 테넌트 없이 `httptest`, dry-run 및 계약 테스트로 완료 판정할 수 있다. | 저장소 관리자 | REQ-042, NFR-006 | 운영 권한·데이터 변형 위험 제거 |
| 유지보수 및 API 체크포인트는 한 선형 브랜치에서 별도 커밋 경계로 보존할 수 있다. | 저장소 관리자 | REQ-010, REQ-043 | 외부 릴리스 시 각 경계를 별도 태그 가능 |

### 6.4 열린 질문

차단 또는 비차단 미결정 항목이 없다. 외부 push·태그·배포는 요구사항 불확실성이 아니라 별도 권한 경계이며, 실행 시점에 다시 승인받는다.

---

## 7. 우선순위

| 우선순위 | REQ-ID | 근거 |
|---|---|---|
| Must — 1차 | REQ-001~REQ-012, NFR-003~NFR-013 | 지원 툴체인·보안·릴리스 안전성 복구 |
| Must — 2차 | REQ-013~REQ-043, NFR-001~NFR-002 | 사용자가 선택한 `#27` 전체 API 범위 |
| Should | 없음 | 합의된 항목은 모두 완료 조건 |
| Could | 없음 | 범위 팽창 방지 |
| Won't | 1.3 제외 범위 | 별도 승인 또는 후속 계획 필요 |

---

## 8. 추적성 요약

| 요구사항 영역 | 요구사항 | 우선순위 | 검증 방법 |
|---|---|---|---|
| Go·Action 현대화 | REQ-001~REQ-005 | Must — 1차 | Test, Analysis, Inspection |
| 운영 이슈 수명주기 | REQ-006~REQ-009 | Must — 1차 | Test, Inspection |
| Git·릴리스 경계 | REQ-010~REQ-012 | Must — 1차 | Inspection, Demonstration |
| 검색·소속 API | REQ-013~REQ-023 | Must — 2차 | Test, Inspection |
| 채널 폴더 | REQ-024~REQ-033 | Must — 2차 | Test, Analysis |
| 추가 파라미터·응답 | REQ-034~REQ-037 | Must — 2차 | Test |
| CLI 통합·문서 | REQ-038~REQ-043 | Must — 2차 | Test, Inspection |
| 보안·품질·호환성 | NFR-001~NFR-013 | Must | Test, Analysis, Inspection |

---

## 9. 품질 검증 보고서

### 9.1 EARS 및 문장 수준 검증

| 검사 | 결과 | 조치 |
|---|---|---|
| 요구사항 주체와 의무 표현 | 통과 | 모든 기능 요구사항에 시스템/저장소 주체와 `해야 한다` 사용 |
| 이벤트·상태·실패 조건 | 통과 | CI 실행, 정상 회복, 취약점 발견, 외부 승인 부재 조건 명시 |
| 모호한 최신 버전 표현 | 통과 | Go, Action, `govulncheck` 버전을 정확히 고정 |
| 측정 불가능한 품질 표현 | 통과 | 명령 종료 코드, 요청 횟수, 커버리지 매트릭스, 테스트로 측정 |
| 구현 상세 혼입 | 통과 | Go/Cobra/파일 구조는 사용자와 저장소가 지정한 기술 제약으로 분리 |

### 9.2 문서 수준 검증

| 검사 | 결과 | 근거 |
|---|---|---|
| 중복 | 통과 | 유지보수·검색·채널 폴더·파라미터 영역 분리 |
| 충돌 | 통과 | 두 릴리스 단위는 선형 체크포인트로 분리하고 외부 쓰기는 제외 |
| 누락 | 통과 | 기능, 비기능, 인터페이스, 데이터, 우선순위, 검증, 제외 범위 포함 |
| 실패 경로 | 통과 | 취약점, CI 실패, 페이지네이션 반복, 커버리지 불일치 처리 포함 |
| 보안 경계 | 통과 | 자격 증명·PII·다운로드 URL·외부 쓰기 경계 포함 |

### 9.3 미해결 불확실성 검증

| 항목 | 결과 |
|---|---|
| Needs-clarification marker | 0건 |
| To-be-determined marker | 0건 |
| To-be-reviewed marker | 0건 |
| 차단 항목 | 없음 |
| 수용된 가정 | 4건, 6.3에 수용자·영향 기록 |

### 9.4 최종 판정

**통과** — Scale M 요구사항으로서 완전성, 일관성, 검증 가능성, 인터페이스 및 비기능 범위를 충족한다. `writing-plans`와 후속 plan review를 시작할 수 있으며, 외부 쓰기는 별도 권한 경계로 유지한다.
