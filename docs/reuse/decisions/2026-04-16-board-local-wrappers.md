# board-local-wrappers

## Context

`cmd/board.go` repeated the same `newSvc(api.NewBoardService)`, `runListCmd(...)`, `readJSONFlag(cmd)`, `requireTitleBodyPost(cmd)`, `printBody(resp.Body)`, and `printResponse(resp)` orchestration across board, post, comment, and attachment CRUD commands. The duplication was high, but the signatures were still tightly bound to `BoardService` and to board-specific id arity.

## Decision

`keep-local`

## Reason

We extracted a local wrapper family inside `cmd/board.go` instead of adding another repo-wide helper layer. The board commands now share `boardIDRunE`, `boardTwoIDRunE`, `boardThreeIDRunE`, `boardFourIDRunE`, and the matching body/list wrappers, while upload flows keep their own explicit orchestration.

## Consequences

- Reuse the local wrapper family for future BoardService CRUD and paginated list commands
- Keep `cmd/helpers.go` focused on repo-wide helpers such as `runListCmd`, `fetchAndPrint`, and JSON parsing primitives
- Update `docs/reuse/catalog.yaml` and `docs/reuse/scorecard.md` to reflect the approved board-local wrapper family and the lower hotspot count
