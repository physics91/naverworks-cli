# contact-local-wrappers

## Context

`cmd/contact.go` repeated the same `newSvc(api.NewContactService)`, `newAPIClientWithUser(cmd)`, `runListCmd(...)`, `readJSONFlag(cmd)`, `parseOptionalJSONData(cmd)`, `printBody(resp.Body)`, and `printResponse(resp)` orchestration across contact, custom-property, and tag CRUD commands. The duplication was high, but the signatures were still tied to `ContactService`, contact-specific optional body builders, and `--user-id`-aware flows.

## Decision

`keep-local`

## Reason

We extracted a local wrapper family inside `cmd/contact.go` instead of creating another repo-wide helper tier. The contact commands now share local wrappers for plain CRUD, paginated lists, `--user-id`-aware flows, and the contact-specific create/update body builders, while photo upload/download flows keep their explicit orchestration.

## Consequences

- Reuse the local wrapper family for future ContactService CRUD and paginated list commands
- Keep `cmd/helpers.go` focused on repo-wide helpers such as `runListCmd`, JSON parsing primitives, and generic output helpers
- Leave photo upload and download URL flows explicit until a second contact-domain upload path appears that truly justifies another local extraction
