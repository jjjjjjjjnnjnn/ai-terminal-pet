# ai-terminal-pet

A tiny always-on-top floating companion for AI coding terminals on Windows.
It aggregates session states (idle / thinking / working / tool_use /
waiting_permission / error / closed) from OpenCode, Claude Code and
Antigravity CLI into one glanceable overlay.

![platform](https://img.shields.io/badge/platform-Windows-blue)
![license](https://img.shields.io/badge/license-MIT-green)

## How it works

```
AI terminals ── reporters/*.js ── named pipe ── AI-Terminal-Pet.exe ── layered overlay
                                  (JSON line,   (Go, single instance)   (Win32, topmost)
                                   ACK = OK)
```

- **Transport**: per-user Windows named pipe
  `\\.\pipe\ai-terminal-pet-v1-<8hex>`, one JSON object per line
  (`schemaVersion: ai-terminal-pet.v1`, 1 MiB cap), server replies
  exact `OK` / `ERR`. Snapshot files under
  `%LOCALAPPDATA%\AI-Terminal-Pet\runtime\` act as a fallback channel.
- **App** (`cmd/ai-terminal-pet`): single instance (mutex), pipe server
  (32 conns, 5 s read timeout), snapshot scanner (2 s), Win32 layered
  overlay (collapsed `40x190`, expanded `200x250`, tray icon).
- **Reporters** (`reporters/`): `opencode.js` (plugin),
  `claude-hook.js`, `antigravity-hook.js`, `antigravity-statusline.js`,
  sharing `protocol.js` (validation, redaction, retry-until-deadline,
  fail-fast, never throws into the host CLI).
- **pet-statusline** (`cmd/pet-statusline`): dependency-free native
  replacement for the Node statusline reporter. On machines where AV
  hooks make every `node`/shell spawn cost seconds, the native shim
  answers in well under a second.

## Build from source (Windows, Go 1.23+)

```powershell
go vet ./...
go test -count=1 ./...
go build -trimpath -ldflags "-s -w" -o .\bin\AI-Terminal-Pet.exe ./cmd/ai-terminal-pet
go build -trimpath -ldflags "-s -w" -o .\bin\pet-statusline.exe ./cmd/pet-statusline
.\install.ps1   # stops the running pet, copies both binaries, SHA256-verifies
```

`node --check reporters/*.js` covers the reporters (stdlib only).

## Wire up a terminal

- **OpenCode**: add the file plugin to `plugin` in `opencode.json`:
  `file:///C:/path/to/ai-terminal-pet/reporters/opencode.js`
- **Claude Code**: add `node <repo>/reporters/claude-hook.js` commands for
  `SessionStart`, `SessionEnd`, `UserPromptSubmit`, `PreToolUse`,
  `PostToolUse`, `Stop`, `Notification` in `settings.json`.
- **Antigravity**: point `statusLine.command` at the installed
  `pet-statusline.exe` (or the `.js` reporter if your Node is fast),
  and register `antigravity-hook.js` for agent/tool events.

Reporters are fire-and-forget: any failure returns silently and never
breaks the host CLI. Restart the terminal after changing its config so
the new reporter loads.

## Layout

```
cmd/ai-terminal-pet    main exe entry (error.log + 1 MB rotation)
cmd/pet-statusline     native statusline shim
internal/app           wiring: mutex, pipe server, scanner, window
internal/pipeserver    named-pipe server (user-isolated name, ACK/ERR)
internal/store         session store (256 cap, tombstones, TTLs, snapshots)
internal/session       schema validation + priority/aggregation/sort
internal/platform      single-instance mutex, process-liveness probe
internal/ui            layered overlay, canvas rasterizer, tray
reporters/             JS push side (see above)
install.ps1            stop → copy → SHA256 verify (both binaries)
```

## Known limitations

- Windows only.
- Per-launch AV scanning can make Node-based reporters slow; use the
  native shim and/or add an AV exclusion for the install directory.
- Same-user pipe preemption is mitigated (mutex-first ordering), not
  eliminated — a future `FILE_FLAG_FIRST_PIPE_INSTANCE` or
  challenge-response upgrade is tracked in HANDOVER.md.
- `go test -race` needs a gcc toolchain; UI has unit tests for pure
  logic only, visual behavior is verified manually.

See [HANDOVER.md](HANDOVER.md) for the operational runbook.

## License

MIT — see [LICENSE](LICENSE). Third-party modules used as libraries only:
`github.com/Microsoft/go-winio` (MIT), `github.com/lxn/win` (BSD-style),
`golang.org/x/sys` (BSD-style). No third-party code is vendored.
