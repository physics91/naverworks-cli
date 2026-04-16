# Data Files

The skill can operate without repository data files, but if reuse work becomes recurring, create these files under `docs/reuse/`.

## `docs/reuse/catalog.yaml`

Purpose: inventory shared helpers and their lifecycle.

Suggested shape:

```yaml
helpers:
  - id: run-list-cmd
    path: cmd/helpers.go
    symbol: runListCmd
    lifecycle: default
    owner: Code Reuse Agent
    use_when:
      - list command uses cursor/count/all pagination
    avoid_when:
      - fetch signature requires extra domain-specific branching
    replacement: ""
```

## `docs/reuse/waivers.yaml`

Purpose: time-bounded exceptions for known divergence.

Suggested shape:

```yaml
waivers:
  - id: drive-shared-folder-list
    path: cmd/drive.go
    helper: runListCmd
    reason: folder-conditional branching does not fit the generic list helper contract
    owner: Code Reuse Agent
    created_on: 2026-04-16
    review_after: 2026-05-16
    expires_on: 2026-07-31
```

## `docs/reuse/decisions/YYYY-MM-DD-<topic>.md`

Purpose: explain why a helper became shared, stayed local, or received a waiver.

Use the template in `assets/decision-template.md`.

## `docs/reuse/scorecard.md`

Purpose: snapshot helper adoption, hotspots, and overdue waivers. The script `scripts/reuse-scorecard.sh` can generate the raw metrics for this file.
