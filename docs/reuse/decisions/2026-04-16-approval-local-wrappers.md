# approval-local-wrappers

## Context

`cmd/approval.go` repeated the same `newSvc(api.NewApprovalService)`, `newAPIClientWithUser(cmd)`, `runListCmd(...)`, `readJSONFlagRaw(cmd)`, `printBody(resp.Body)`, and `printResponse(resp)` orchestration across approval document, category, form, and linkage-code CRUD commands. The duplication was obvious, but the signatures were still tightly coupled to `ApprovalService`, raw JSON payloads, and approval-specific user-context flows.

## Decision

`keep-local`

## Reason

We extracted a local wrapper family inside `cmd/approval.go` instead of creating another repo-wide helper tier. The approval commands now share local wrappers for plain CRUD, paginated lists, and `--user-id`-aware body/list flows, while file upload commands keep their explicit orchestration.

## Consequences

- Reuse the local wrapper family for future ApprovalService CRUD and paginated list commands
- Keep `cmd/helpers.go` focused on repo-wide helpers such as `runListCmd`, `fetchAndPrint`, and raw JSON parsing
- Leave attachment upload flows explicit until a second approval-domain upload path appears that truly justifies another local extraction
