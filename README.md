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
  plus `generic.js` — a one-line bridge for any other CLI —
  sharing `protocol.js` (validation, redaction, retry-until-deadline,
  fail-fast, never throws into the host CLI).
- **pet-statusline** (`cmd/pet-statusline`): dependency-free native
  replacement for the Node statusline reporter. On machines where AV
  hooks make every `node`/shell spawn cost seconds, the native shim
  answers in well under a second.

## Control the overlay

The overlay is single-instance and removes itself once no terminal session
is left, so it is normally invisible in a process list. Every command below
is safe to bind to a hotkey or shell alias:

```powershell
ai-terminal-pet --show      # reveal the running overlay
ai-terminal-pet --toggle    # switch between compact and expanded
ai-terminal-pet --reset     # re-anchor to the screen edge at 100% zoom
ai-terminal-pet --quit      # stop the overlay
ai-terminal-pet --version
```

Launching the executable with no arguments while an instance is already
running reveals that instance instead of exiting silently.

## Tray menu

Right-click the tray icon:

- **显示浮窗 / 隐藏到托盘**
- **重置位置与大小** — escape hatch for a pet dragged somewhere awkward
- **✓ 状态提醒** — tray balloons when a session starts waiting for
  permission or fails
- **✓ 窗口置顶**
- **退出**

## Settings and remembered state

Two files under `%LOCALAPPDATA%\AI-Terminal-Pet\`:

- `config.json` — yours. Read at startup, rewritten only when you flip a
  tray switch, so editing it by hand while the app is stopped is safe.
  Missing or malformed values fall back to defaults; a broken file never
  stops the overlay from starting.

  ```json
  { "notifications": true, "alwaysOnTop": true, "opacity": 235 }
  ```

- `state.json` — the app's. Window position, zoom and expanded flag, so the
  pet comes back where you left it. Delete it (or run `--reset`) to get the
  default right-edge anchor back.

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
- **Anything else**: one line, no plugin needed. Any tool that can run a
  shell command on state changes can report:

  ```powershell
  node reporters/generic.js --source=aider --state=working --project="$PWD"
  node reporters/generic.js --source=codex --state=waiting_permission --reason=approve
  echo '{"state":"error","reason":"build failed"}' | node reporters/generic.js --source=ci --stdin
  ```

  States: `idle thinking working tool_use waiting_permission error closed`.
  `--help` lists every option.

Reporters are fire-and-forget: any failure returns silently and never
breaks the host CLI. Restart the terminal after changing its config so
the new reporter loads.

## Layout

```
cmd/ai-terminal-pet    main exe entry (CLI control flags + error.log rotation)
cmd/pet-statusline     native statusline shim
internal/app           wiring: mutex, pipe server, scanner, window, state saver
internal/config        config.json + state.json (atomic writes, fail-safe reads)
internal/pipeserver    named-pipe server (user-isolated name, ACK/ERR)
internal/store         session store (256 cap, tombstones, TTLs, snapshots)
internal/session       schema validation + priority/aggregation/sort
internal/platform      single-instance mutex, process-liveness probe
internal/ui            layered overlay, canvas rasterizer, tray, notifications
reporters/             JS push side (incl. generic.js)
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

See [HANDOVER.md](HANDOVER.md) for the operational runbook and
[ROADMAP.md](ROADMAP.md) for the extension and lightweighting plans.

## License

MIT — see [LICENSE](LICENSE). Third-party modules used as libraries only:
`github.com/Microsoft/go-winio` (MIT), `github.com/lxn/win` (BSD-style),
`golang.org/x/sys` (BSD-style). No third-party code is vendored.
