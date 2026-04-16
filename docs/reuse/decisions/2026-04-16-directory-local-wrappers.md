# directory-local-wrappers

## Context

`cmd/directory.go` had a large pile of repeated `newSvc(api.NewDirectoryService)`, `readJSONFlagRaw(cmd)`, `printBody(resp.Body)`, and `printResponse(resp)` orchestration across CRUD-style commands. The duplication pressure was real, but the signatures were still tightly coupled to `DirectoryService` and the command family's raw body/response printing conventions.

## Decision

`keep-local`

## Reason

We extracted a local wrapper family inside `cmd/directory.go` instead of promoting another generic helper tier into `cmd/helpers.go`. The directory commands now reuse `directoryRunE`, `directoryBodyRunE`, `directoryIDRunE`, `directoryIDBodyRunE`, `directoryTwoIDRunE`, and `directoryTwoIDBodyRunE`, while list commands, upload/download flows, and other custom paths keep their existing specialized logic.

## Consequences

- Reuse the local wrapper family when adding more CRUD-style directory commands
- Keep `cmd/helpers.go` focused on repo-wide helpers such as `runListCmd`, `fetchAndPrint`, and `readJSONFlagRaw`
- Update `docs/reuse/catalog.yaml` and `docs/reuse/scorecard.md` to reflect the approved directory-local helper family and the reduced hotspot count
