# Operating Model

## Decision Classes

| Decision | Use when | Avoid when |
|----------|----------|------------|
| `reuse` | Existing helper already fits the call shape and preserves clarity | You need per-call branching that the helper hides |
| `extract` | Same pattern appears in 2+ places and the abstraction has a crisp contract | Only one site exists or the helper name would be vague |
| `keep-local` | The logic is domain-specific and reads better inline | The same block has already spread to multiple domains |
| `waive` | Shared abstraction is premature today but worth revisiting later | You are using waiver to avoid obvious cleanup with no expiry |

## Lifecycle

Use these helper lifecycle states in `catalog.yaml`.

| State | Meaning |
|-------|---------|
| `experimental` | Newly extracted helper; shape may still change |
| `approved` | Valid shared helper with at least one confirmed reuse case |
| `default` | Preferred path; new code should use this unless waived |
| `deprecated` | Do not introduce new usages; migrate toward replacement |
| `removed` | Historical record only; no live call sites should remain |

## Waiver Policy

Every waiver must include:

- `id`
- `path`
- `helper`
- `reason`
- `owner`
- `created_on`
- `expires_on`
- `review_after`

Recommended defaults:

- short-lived cleanup waiver: 30-45 days
- larger refactor boundary waiver: next scheduled refactor date or release milestone

## Known naverworks-cli Heuristics

- Check `cmd/helpers.go` before inventing new command-side helpers
- Prefer existing pagination flow (`addListFlags` + `runListCmd`) when the fetch signature is `(cursor, count)`
- Prefer `readJSONFlagRaw` over ad-hoc JSON flag parsing
- Prefer `readStdinLimited` over custom stdin readers
- Treat `cmd/drive.go` and `cmd/directory.go` as hotspot files; require extra skepticism before adding one-off helpers there

## Decision Note Triggers

Write a short decision note when:

- extracting a new shared helper
- deprecating or removing a helper
- granting a waiver that survives beyond the current change
- rejecting an apparently obvious abstraction because the divergence is intentional
