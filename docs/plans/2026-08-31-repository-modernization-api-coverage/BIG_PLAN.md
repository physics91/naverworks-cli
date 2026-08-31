<!-- plan-workflow: hierarchical-v1 -->
# naverworks-cli 저장소 현대화 및 API 커버리지 Big Plan

- Plan ID: 2026-08-31-repository-modernization-api-coverage
- Grade: M
- Requirements Source: 2026-08-31-repository-modernization-api-coverage.md
- Requirements SHA-256: 623b5180ab4bf6edd9c83493e018355002de4519b273c1054a52805ba6c43b00
- Quality Status: Passed
- Repository Base: b9af677fdaf721e7fcc7095df958de1bc97c9af1

## Goal

장기간 관리 공백이 있었던 저장소를 Go 1.26.7과 검증된 자동화 기준으로 복구하고, 유지보수 체크포인트 뒤에 공식 NAVER WORKS API 미지원 범위를 구현한 별도 API 체크포인트를 만들어 로컬 `develop` 병합까지 안전하게 인계한다.

## Scope

- In: Go 및 GitHub Actions 현대화, 포맷·취약점 검증, Weekly Health와 API monitor 수명주기 보강, 기존 미배포 변경 검증, 공식 검색·소속·채널 폴더·추가 파라미터 API, CLI 계약·문서·합성 테스트, 두 개의 선형 릴리스 체크포인트, 검증된 로컬 병합.
- Out: Go 1.27 전환, 새 런타임 의존성, 실제 NAVER WORKS 테넌트 호출, Bot callback 서버, 원격 push·태그·이슈 변경·Release·npm publish, 범위 밖 구조 개편.
- Security boundary: 자격 증명·토큰·PII·presigned URL과 민감 요청 본문은 CLI와 외부 API 사이의 신뢰 경계를 넘으며, 인증 scope 우회·민감정보 출력·안전하지 않은 다운로드·파괴적 호출·Action 공급망 변조를 주요 위협으로 다룬다.
- Security controls: 전체 SHA Action 핀과 고정 취약점 도구는 TASK-001, GitHub 이슈 대상 제한은 TASK-002, 인증·응답·다운로드 계약은 TASK-004부터 TASK-010에서 검증한다. 실제 테넌트 미검증 위험은 저장소 관리자가 소유하며 외부 승인 후 제한된 운영 검증으로 해소한다.
- Compatibility boundary: 유지보수와 API 변경은 선형 체크포인트로 분리하고 기존 명령·출력·오류 계약을 회귀 검증한다. 중단 시 마지막 검증 체크포인트로 되돌리고, 원격 배포 전까지 변경은 로컬 이력으로만 보존한다.
- Production boundary: 로컬 구현 수용은 합성 테스트·정적 검사·빌드로 판정한다. Windows GitHub Actions, 실제 릴리스 및 운영 관찰은 외부 쓰기 승인 이후의 배포 수용 단계이며, 실패 시 push·태그·배포를 중단하고 해당 체크포인트를 수정 또는 되돌린다.
- Material risks: 도구 실행 취약점은 스캔 실패와 의존성 diff로 감지해 체크포인트를 중단하고 버전 고정 또는 변경 철회로 복구한다. 인증·PII·삭제 위험은 계약 테스트와 민감 출력 검증 실패 시 중단하고 관련 API 변경을 되돌린다. API 문서 드리프트는 커버리지 매트릭스 불일치 시 API 체크포인트를 중단하고 기준선과 구현을 재조정한다.

## Completion Criteria

- 유지보수 체크포인트가 Go 1.26.7, Action SHA 핀, 포맷·취약점 검사, 운영 이슈 수명주기와 기존 `develop` 변경을 포함하고 모든 로컬 검증을 통과한다.
- 공식 2026-08-31 API 기준선의 검색·소속·채널 폴더·추가 파라미터가 명령, API 계약, 도움말, 테스트 및 커버리지 매트릭스로 연결된다.
- 기존 CLI 호환성, 인증·보안·다운로드·페이지네이션·오류 계약이 합성 테스트에서 유지되고 새 런타임 의존성이 추가되지 않는다.
- 유지보수와 API 체크포인트의 커밋 식별자, 검증 결과, 릴리스 노트, 롤백 경계가 각각 기록된다.
- 정확한 최종 브랜치가 로컬 `develop`에 비강제 병합되고 primary에는 병합 전용 쓰기만 발생한다.
- 원격 push·태그·이슈 변경·Release·npm publish는 승인 전 수행하지 않으며, Windows GitHub Actions 성공 증거는 별도 외부 쓰기 승인 후 TASK-012에서 수집해야 전체 계획이 완료된다.

## Task Graph

### TASK-001 — 툴체인과 검증 자동화 현대화
- Requirements: REQ-001, REQ-002, REQ-003, REQ-004, REQ-005, NFR-006, NFR-009, NFR-010
- Predecessors: none
- Priority: must
- Risk: high
- Deliverable: Go 1.26.7 기준, 검증된 Action 핀, 포맷 및 도달 가능 취약점 검사를 포함하는 재현 가능한 유지보수 자동화.
- Verification: 고정된 툴체인과 자동화 정의가 로컬 품질 계층을 통과하고 Windows 실행 준비 상태와 의존성 무변경 근거가 확인된다.

### TASK-002 — 운영 이슈 수명주기와 API 모니터 기준선 정비
- Requirements: REQ-006, REQ-007, REQ-008, REQ-009, NFR-011
- Predecessors: TASK-001
- Priority: must
- Risk: high
- Deliverable: 자동화 소유 이슈만 생성·갱신·종료하고 영향 없는 공지는 기준선 처리하되 미완료 API 공지는 유지하는 운영 동작.
- Verification: 합성 GitHub API 시나리오에서 성공·실패·중복·대상 불일치 동작과 공지 기준선 판단이 결정적으로 확인된다.

### TASK-003 — 유지보수 릴리스 체크포인트 확정
- Requirements: REQ-010, REQ-011, REQ-012, NFR-012, NFR-013
- Predecessors: TASK-002
- Priority: must
- Risk: high
- Deliverable: 기존 미배포 변경과 유지보수 변경을 보존한 검증 가능한 첫 번째 릴리스 경계 및 외부 쓰기 전 인계 기록.
- Verification: 전용 비보호 브랜치의 선형 커밋 경계가 전체 로컬 검증, 실패 차단, 변경·롤백 기록을 갖추고 원격 상태를 바꾸지 않았음이 확인된다.

### TASK-004 — Directory 검색·소속·근무 상태 계약 확장
- Requirements: REQ-017, REQ-018, REQ-019, REQ-022, REQ-036, NFR-004
- Predecessors: TASK-003
- Priority: must
- Risk: high
- Deliverable: User·OrgUnit·Group 검색, 구성원 그룹·조직 소속 조회, 근무 상태 delegates 출력을 기존 Directory 인증 경계로 제공하는 API와 CLI 동작.
- Verification: 합성 요청·응답에서 경로, 식별자, 인증 경계, 소속 목록, delegates 보존과 기존 보안 계약이 확인된다.

### TASK-005 — Board·Note·Calendar·Contact 검색 제공
- Requirements: REQ-013, REQ-014, REQ-015, REQ-016
- Predecessors: TASK-003
- Priority: must
- Risk: high
- Deliverable: 네 도메인의 공식 검색 엔드포인트를 기존 사용자·그룹 식별 및 출력 규칙에 맞춘 API와 CLI 동작.
- Verification: 각 도메인의 합성 요청·응답에서 공식 경로, 필수·선택 입력, 출력 계약 및 대표 오류가 확인된다.

### TASK-006 — Task·Approval 검색과 결재 유형 필터 제공
- Requirements: REQ-020, REQ-021, REQ-037
- Predecessors: TASK-003
- Priority: must
- Risk: high
- Deliverable: Task 검색과 관리자 Approval 검색 및 선택적 결재 유형 필터를 올바른 권한 경계로 제공하는 API와 CLI 동작.
- Verification: 합성 요청에서 검색 조건과 선택 필터가 정확히 전달되고 관리자 scope, 미지정 호환성 및 오류 계약이 확인된다.

### TASK-007 — Drive·Monitoring 선택 파라미터 확장
- Requirements: REQ-034, REQ-035
- Predecessors: TASK-003
- Priority: must
- Risk: high
- Deliverable: Drive 검색의 queryFilters와 Monitoring 콘텐츠 다운로드의 channelId를 선택적으로 전달하는 하위 호환 동작.
- Verification: 파라미터 지정·미지정 합성 요청에서 직렬화, 기존 동작 유지 및 민감 다운로드 경계가 확인된다.

### TASK-008 — 채널 폴더 공식 기준선과 API 계층 완성
- Requirements: REQ-024, REQ-033
- Predecessors: TASK-003
- Priority: must
- Risk: high
- Deliverable: 2026-08-31 공식 채널 폴더 엔드포인트 전부를 열거하고 구현 상태와 연결하는 커버리지 기준선 및 API 계층.
- Verification: 공식 기준선과 API 계층의 누락·과잉이 없는 일대일 매핑이 자동 검사로 확인되고 불일치가 체크포인트를 차단한다.

### TASK-009 — 채널 폴더 전체 CLI 동작 제공
- Requirements: REQ-025, REQ-026, REQ-027, REQ-028, REQ-029, REQ-030, REQ-031, REQ-032
- Predecessors: TASK-008
- Priority: must
- Risk: high
- Deliverable: 채널 폴더 속성, 파일·폴더, 잠금·중요 표시, 버전, 링크, 휴지통, 권한의 공식 전체 동작을 노출하는 CLI 명령군.
- Verification: 읽기·쓰기·다운로드·복원·영구 삭제·권한 변경의 대표 합성 계약이 공식 경로와 안전 경계를 만족한다.

### TASK-010 — 신규 API의 공통 CLI 계약과 보안·호환성 통합
- Requirements: REQ-023, REQ-038, REQ-039, REQ-040, REQ-041, REQ-042, NFR-001, NFR-002, NFR-003, NFR-005, NFR-008
- Predecessors: TASK-004, TASK-005, TASK-006, TASK-007, TASK-009
- Priority: must
- Risk: high
- Deliverable: 신규 API 전체에 일관된 구조, 페이지네이션, 도움말, 출력, dry-run, 다운로드, 오류 및 테스트 계약을 적용한 통합 상태.
- Verification: 명령 등록·플래그·경로·페이지네이션·요청 횟수·출력·민감정보·다운로드·실패 경로의 통합 회귀 검증이 통과한다.

### TASK-011 — API 릴리스 체크포인트와 로컬 병합 인계
- Requirements: REQ-043
- Predecessors: TASK-010
- Priority: must
- Risk: high
- Deliverable: 유지보수 경계 이후의 독립된 API 릴리스 경계, 릴리스 노트, 전체 검증 증거 및 비강제 로컬 병합 결과.
- Verification: 두 체크포인트가 식별 가능하고 최종 정확한 이력이 로컬 `develop`에 병합되며 원격 쓰기와 Windows 배포 수용은 승인 대기로 명확히 분리된다.

### TASK-012 — 승인 후 Windows CI 수용
- Requirements: NFR-007
- Predecessors: TASK-011
- Priority: must
- Risk: high
- Deliverable: 사용자 외부 쓰기 승인 후 유지보수 및 API 체크포인트의 정확한 커밋에 대해 수집한 GitHub-hosted Windows CI 성공 증거.
- Verification: 각 체크포인트의 workflow run URL, commit SHA 및 성공 conclusion이 기록되며, 승인·실행·증거 중 하나라도 없으면 NFR-007과 전체 계획이 미완료로 유지된다.

## Traceability

| Requirement | Task | Completion signal |
|---|---|---|
| REQ-001 | TASK-001 | Go 기준이 1.26.7로 일치한다. |
| REQ-002 | TASK-001 | 사용 중인 Action이 합의 버전의 전체 SHA를 사용한다. |
| REQ-003 | TASK-001 | 포맷 편차가 검증 실패로 처리된다. |
| REQ-004 | TASK-001 | 고정 버전 취약점 검사가 CI와 릴리스 사전에 포함된다. |
| REQ-005 | TASK-001 | 도달 가능한 취약점이 후속 단계를 차단한다. |
| REQ-006 | TASK-002 | 정상 회복 시 정확한 자동화 소유 이슈만 종료된다. |
| REQ-007 | TASK-002 | 실패 시 동일 이슈가 중복 없이 생성 또는 갱신된다. |
| REQ-008 | TASK-002 | 비영향 공지 slug와 근거가 baseline에 반영된다. |
| REQ-009 | TASK-002 | 미완료 API 공지는 열린 추적 상태로 유지된다. |
| REQ-010 | TASK-003 | 유지보수 변경이 기존 미배포 이력 뒤에 선형 보존된다. |
| REQ-011 | TASK-003 | 체크포인트와 검증·롤백 인계가 기록된다. |
| REQ-012 | TASK-003 | 승인 전 외부 쓰기가 발생하지 않는다. |
| REQ-013 | TASK-005 | Board 검색 계약과 명령이 검증된다. |
| REQ-014 | TASK-005 | 조직·그룹 Note 검색 계약과 명령이 검증된다. |
| REQ-015 | TASK-005 | Calendar 검색 계약과 명령이 검증된다. |
| REQ-016 | TASK-005 | Contact 검색 계약과 명령이 검증된다. |
| REQ-017 | TASK-004 | User 검색 계약과 명령이 검증된다. |
| REQ-018 | TASK-004 | OrgUnit 검색 계약과 명령이 검증된다. |
| REQ-019 | TASK-004 | Group 검색 계약과 명령이 검증된다. |
| REQ-020 | TASK-006 | Task 검색 계약과 명령이 검증된다. |
| REQ-021 | TASK-006 | 관리자 Approval 검색 계약과 명령이 검증된다. |
| REQ-022 | TASK-004 | 구성원 그룹·조직 소속 조회가 각각 검증된다. |
| REQ-023 | TASK-010 | 지원되는 검색·목록에 기존 페이지네이션 동작이 적용된다. |
| REQ-024 | TASK-008 | 공식 채널 폴더 엔드포인트가 매트릭스에 전부 기록된다. |
| REQ-025 | TASK-009 | 채널 폴더 목록·속성 명령이 검증된다. |
| REQ-026 | TASK-009 | 채널 파일 전체 동작이 검증된다. |
| REQ-027 | TASK-009 | 루트·하위 폴더 생성이 검증된다. |
| REQ-028 | TASK-009 | 잠금·중요 표시 전환이 검증된다. |
| REQ-029 | TASK-009 | 파일 버전 전체 동작이 검증된다. |
| REQ-030 | TASK-009 | 링크 전체 동작이 검증된다. |
| REQ-031 | TASK-009 | 휴지통 목록·복원·영구 삭제가 검증된다. |
| REQ-032 | TASK-009 | 접근 권한 전체 동작이 검증된다. |
| REQ-033 | TASK-008 | 기준선 불일치가 API 체크포인트를 차단한다. |
| REQ-034 | TASK-007 | Drive queryFilters 지정·미지정 동작이 검증된다. |
| REQ-035 | TASK-007 | Monitoring channelId 지정·미지정 동작이 검증된다. |
| REQ-036 | TASK-004 | 근무 상태 delegates가 손실 없이 출력된다. |
| REQ-037 | TASK-006 | Approval type 필터가 선택적으로 전달된다. |
| REQ-038 | TASK-010 | 신규 구현이 기존 도메인 구조를 따른다. |
| REQ-039 | TASK-010 | 도움말과 문서에 인증 주체와 scope가 표시된다. |
| REQ-040 | TASK-010 | 적용 가능한 목록 명령이 JSON과 table 출력을 제공한다. |
| REQ-041 | TASK-010 | 쓰기·다운로드 명령이 기존 안전 계약을 유지한다. |
| REQ-042 | TASK-010 | 경로·등록·플래그·실패 테스트가 신규 명령을 덮는다. |
| REQ-043 | TASK-011 | API 체크포인트가 별도 경계와 릴리스 노트를 갖는다. |
| NFR-001 | TASK-010 | 단일 페이지 요청 횟수가 계약 범위에 머문다. |
| NFR-002 | TASK-010 | 전체 페이지 조회의 반복 감지와 종료가 유지된다. |
| NFR-003 | TASK-010 | 민감정보가 로그·오류·dry-run에 노출되지 않는다. |
| NFR-004 | TASK-004 | 기존 인증·저장·크기 제한 보안 계약이 통과한다. |
| NFR-005 | TASK-010 | 다운로드가 HTTPS와 허용 호스트 경계를 지킨다. |
| NFR-006 | TASK-001 | 각 로컬 품질 검증 계층이 성공한다. |
| NFR-007 | TASK-012 | 승인 후 두 체크포인트의 Windows CI 성공 증거가 기록된다. |
| NFR-008 | TASK-010 | 기존 명령·출력·오류 호환성이 유지된다. |
| NFR-009 | TASK-001 | 새 런타임 직접 의존성이 없다. |
| NFR-010 | TASK-001 | Go 1.26.7에서 빌드와 정적 검사가 통과한다. |
| NFR-011 | TASK-002 | GitHub 이슈 대상 검증이 합성 테스트를 통과한다. |
| NFR-012 | TASK-003 | 실패 원인·재현 정보가 기록되고 경계가 차단된다. |
| NFR-013 | TASK-003 | 모든 개발 쓰기가 바인딩된 비보호 worktree에 남는다. |

## Handoff

- Plan status: canonical
- Review: plan-review offered and accepted; round 1 blocker resolved locally
- Implementation authority: 승인된 요구사항과 이 계획 범위의 로컬 코드·테스트·문서 변경, task commit 및 검증된 로컬 `develop` 병합까지 허용한다.
- External authority boundary: TASK-012의 push 및 Windows 원격 CI, 태그, GitHub 이슈 상태 변경, Release, npm publish는 수행 전에 사용자 승인을 다시 받아야 하며 승인 전 계획 상태는 외부 쓰기 대기로 유지한다.
