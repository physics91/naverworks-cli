# Project Guidance Evidence

- Evidence fingerprint: `9b2e74272496719ec647bbe41c41216c41fac064ba242ad163932972f5c53ede`
- Generator format: `3`

## Applicability

| Path | Stack | Evidence |
|---|---|---|
| `.` | Go, Shell | `@source-stack:Go`, `@source-stack:Shell`, `go.mod` |
| `.agents/skills/reuse-governor-local/scripts` | Shell | `@source-stack:Shell` |
| `cmd` | Go | `@source-stack:Go` |
| `internal` | Go | `@source-stack:Go` |
| `npm` | Shell | `@source-stack:Shell` |
| `npm/cli` | Node.js | `@source-stack:Node.js`, `npm/cli/package.json` |
| `scripts` | Go, Shell | `@source-stack:Go`, `@source-stack:Shell` |

## Repository Evidence

- `@source-imports` — `d72eb6bbdd54789ac39ebafd9fdba9cde6345e66e5c7f2e28016d429d806ae83`
- `@source-layout` — `415d7b3974223c583719ac6b3d46ac537f95d1d73eef00826ae52189e270f031`
- `@source-module:.:Go` — `bf5f10148ff09907859aebd63cee1711d9d473e7d4ee2b2e425095b9b71a3e7c`
- `@source-module:.:Shell` — `043df5bdbf6639d7a77e1d44c5226fd7371e5259a1e4df3a0dd5d64c30dca44f`
- `@source-module:.agents%2Fskills%2Freuse-governor-local%2Fscripts:Shell` — `6bb7af3470c76a1e932377d984f75771861930fe6b40e4670f86dbe3f7748f40`
- `@source-module:cmd:Go` — `ae76bc31382905a384422b4552d61deae831a8b71f0ba3a255d865e75e7207ab`
- `@source-module:internal:Go` — `b24738a44b4f6a79717ef76bfc23e437379c29b2b7c2359b55c4543a4ab90966`
- `@source-module:npm%2Fcli:Node.js` — `c9cd190dd977eb6895acd9f3b8c4341d3528ac1cf826f7e4d11219a15fb91bdf`
- `@source-module:npm:Shell` — `4729a314b272a3a71f474aec0c9831f839783b7dbbf1edce8755a3edf3d56704`
- `@source-module:scripts:Go` — `91e9587ecb792450fa31e09c03084b05a8ae60af26202e828969f920e72c1e60`
- `@source-module:scripts:Shell` — `5bd68ce941b81623e54a9a205c7345522706636e1223e5f4c49c85bc8f9826c1`
- `@source-stack:Go` — `643679ff6987e2d75ddd0de8fcdd77807705e8c67ed03cab939c81d9f02fc587`
- `@source-stack:Node.js` — `c9cd190dd977eb6895acd9f3b8c4341d3528ac1cf826f7e4d11219a15fb91bdf`
- `@source-stack:Shell` — `9e1a10d36dafb36262e66bf587d970d46aa1cd41642da89845253245e1c28b1e`
- `AGENTS.md` — `9699d56d3175846443adbce3bd648432ed53bce0898c0961b3127f2861dbf207`
- `Makefile` — `38edca85786ffc22c547c68ace41a3bb556d358bcb543c6866a4f46dc155e4af`
- `README.md` — `033ca590f48216654698b4f1d16a0b497699813c57f9d112bbd999984b4fdb92`
- `cmd/config_profile_test.go` — `08e4dc93cb5505b52e94b82091ca675927cef52c0e6036e0cda7b319e7c7ae4f`
- `go.mod` — `033d19226aca78f2dd2a1befd806752c025bce0df41665c208a149b005206c57`
- `go.sum` — `b51bfc522536e74c229135c3ed9cba939cb146f6e681c537324bbb972beecf40`
- `npm/cli/package.json` — `e642d4e1a6140241f6392cb8a61a00bf5ce257b203c99923680f3c23f12d5d3c`

## Official Version Sources

- No version-sensitive external guidance was required

## Merge Diagnostics

- No keyed manual-rule conflict detected

## Analysis Diagnostics

- No evidence-analysis omission detected

## Repository-Native Commands

- `go test ./...` [evidence: `go.mod`]
- `npm --prefix npm/cli run postinstall` [evidence: `npm/cli/package.json`]
