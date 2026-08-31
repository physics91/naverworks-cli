# API monitor baseline decisions

이 문서는 자동 감지된 릴리즈 노트를 `api-monitor-baseline.json`에 포함하거나 계속 추적하기로 한 판단 근거를 기록한다.

| Slug | GitHub issue | Decision | Evidence |
|---|---:|---|---|
| `core_20260514` | #28 | Baseline | 2026-05-14 공지는 iOS 앱의 외부 그룹 생성·수정 오류 수정만 포함하며 REST API 계약 변경이 없다. |
| `management_20260821` | #30 | Baseline | 2026-08-21 공지는 경영지원 데이터 내보내기의 파일 분할 및 포함 항목 변경이며 이 CLI가 호출하는 REST API 계약 변경이 없다. |
| `core_20260723` | #27 | Keep tracking | Developers 섹션에 검색, 구성원 소속, 채널 폴더와 추가 파라미터·응답 필드가 포함되어 현재 API 커버리지 작업이 완료될 때까지 baseline에 넣지 않는다. |

## Sources checked

- `https://naver.worksmobile.com/release-notes/core_20260514/`
- `https://naver.worksmobile.com/release-notes/management_20260821/`
- `https://naver.worksmobile.com/release-notes/core_20260723/`
- `https://github.com/physics91/naverworks-cli/issues/27`
- `https://github.com/physics91/naverworks-cli/issues/28`
- `https://github.com/physics91/naverworks-cli/issues/30`

확인 기준일: 2026-08-31
