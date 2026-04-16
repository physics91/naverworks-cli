# drive-shared-folder-list

## Context

`cmd/drive.go` was reviewed during the simplify2 reuse cleanup. Most list commands were migrated to `addListFlags()` and `runListCmd()`, but the shared-folder file listing path still kept folder-conditional branching before fetch execution. The same `folder/cursor/count` branching shape also already existed in `listFilesWithFolder()` for MyDrive, SharedDrive, and GroupFolder flows.

## Decision

`extract`

## Reason

The current `runListCmd()` contract assumes a straight cursor/count/all pagination flow, so forcing the shared-folder path into that helper would still be misleading. Instead of carrying a long-lived waiver, we extracted `listFilesWithOptionalFolder()` so the drive-family commands can share the `folder/cursor/count` branch contract without overloading `runListCmd()`.

## Consequences

- Reuse `listFilesWithOptionalFolder()` for MyDrive, SharedDrive, GroupFolder, and SharedFolder list commands
- Remove the waiver entry from `docs/reuse/waivers.yaml`
- Keep `runListCmd()` as the preferred helper for plain cursor/count/all pagination, and use the new helper only for folder-aware list branching
