# bot-local-wrappers

## Context

`cmd/bot.go` repeated the same `newSvc(api.NewBotService)`, `newBotSvc(cmd)`, `runListCmd(...)`, `readJSONFlagRaw(cmd)`, `printBody(resp.Body)`, and `printResponse(resp)` orchestration across top-level bot CRUD, channel/domain commands, persistent menus, and rich menu CRUD commands. The duplication was high, but the signatures were still tied to `BotService`, `resolveBotID`-based scoped flows, and bot-domain upload/download behavior.

## Decision

`keep-local`

## Reason

We extracted a local wrapper family inside `cmd/bot.go` instead of creating another repo-wide helper tier. The bot commands now share local wrappers for top-level bot CRUD and for `resolveBotID`-aware scoped CRUD/list commands, while send validation, attachment upload/download, and file-based rich menu image flows stay explicit.

## Consequences

- Reuse the local wrapper family for future BotService CRUD and paginated list commands
- Keep `cmd/helpers.go` focused on repo-wide helpers such as `runListCmd`, raw JSON parsing, and generic upload/download utilities
- Leave send, attachment, and rich menu image flows explicit until another bot-domain path appears that truly justifies a separate local extraction
