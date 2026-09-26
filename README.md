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
  overlay (compact `56x156`, expanded `200x<content>`, tray icon).
- **Reporters** (`reporters/`): `opencode.js` (plugin),
  `claude-hook.js`, `antigravity-hook.js`, `antigravity-statusline.js`,
  plus `generic.js` — a one-line bridge for any other CLI —
  sharing `protocol.js` (validation, redaction, retry-until-deadline,
  fail-fast, never throws into the host CLI).
- **pet-statusline** (`cmd/pet-statusline`): dependency-free native
  replacement for the Node statusline reporter. On machines where AV
  hooks make every `node`/shell spawn cost seconds, the native shim
  answers in well under a second.

## The overlay

**Compact** answers three questions in the order a glance asks them:

```
   ◉      state orb - colour plus glyph, pulses while something waits on you
   3      how many sessions are behind it
 会话
  ────
等待确认   what that state is, in words
ai-term…  which project is the one waiting
```

**Expanded** is the list itself. It grows with the number of sessions
(up to six rows, then "还有 N 个会话"), and each row shows the project, its
state, where it came from and how long ago it last spoke:

```
终端会话                            [等待确认]  ✕
──────────────────────────────────────────────
3 个会话 · 2 运行 · 1 等待
▍ai-terminal-pet          等待确认   2 分钟前
  claude-hook · Bash(go test)
▍webapp                   执行中     刚刚
  opencode · editing src/app.tsx
──────────────────────────────────────────────
      拖拽移动 · 滚轮缩放 · 点击收起
```

Interaction is deliberately uniform — there is nothing to aim at:

| gesture | effect |
|---|---|
| click anywhere | compact ↔ expanded |
| drag anywhere | move, 1:1 with the cursor |
| wheel | zoom 65 %–160 % |
| hover | the panel lights up, so it reads as a control |
| `Esc` | collapse |
| right-click | menu (hide, reset, notifications, sound, topmost, autostart, quit) |

The overlay is single-instance and removes itself once no terminal session
is left, so it is normally invisible in a process list.

## Control the overlay

Every command below is safe to bind to a hotkey or shell alias:

```powershell
ai-terminal-pet --show      # reveal the running overlay
ai-terminal-pet --toggle    # switch between compact and expanded
ai-terminal-pet --reset     # re-anchor to the screen edge at 100% zoom
ai-terminal-pet --quit      # stop the overlay
ai-terminal-pet --version
```

Launching the executable with no arguments while an instance is already
running reveals that instance instead of exiting silently.

## Read the state from a script

`--json` prints the current sessions and exits. It reads the reporter
snapshots on disk rather than talking to the overlay, so it works whether or
not the overlay is running and can never block on a GUI process:

```powershell
ai-terminal-pet --json            # one line, for jq / status lines
ai-terminal-pet --json --pretty   # indented, for humans
```

```json
{"schemaVersion":"ai-terminal-pet.v1","state":"waiting_permission","count":2,
 "generatedAt":"2026-09-26T05:31:37Z","sessions":[{"source":"opencode", ...}]}
```

`state` is the highest-priority state across all sessions; `sessions` is the
same list, already sorted by urgency. A reporter that stopped writing is
absent, not stale.

## Tray menu

Right-click the tray icon:

- **显示浮窗 / 隐藏到托盘**
- **重置位置与大小** — escape hatch for a pet dragged somewhere awkward
- **✓ 状态提醒** — tray balloons when a session starts waiting for
  permission or fails
- **✓ 提示音** — a system alert tone on the same transitions
- **✓ 窗口置顶**
- **✓ 开机自启** — run at logon, via `HKCU\...\CurrentVersion\Run`
- **退出**

## Settings and remembered state

Two files under `%LOCALAPPDATA%\AI-Terminal-Pet\`:

- `config.json` — yours. Read at startup, rewritten only when you flip a
  tray switch, so editing it by hand while the app is stopped is safe.
  Missing or malformed values fall back to defaults; a broken file never
  stops the overlay from starting.

  ```json
  { "notifications": true, "sound": true, "alwaysOnTop": true, "opacity": 235 }
  ```

  `sound` is separate from `notifications` because the two reach you
  differently: a balloon is something you read when you look, a tone is
  something you hear when you are not looking.

- `state.json` — the app's. Window position, zoom and expanded flag, so the
  pet comes back where you left it. Delete it (or run `--reset`) to get the
  default right-edge anchor back.

Autostart is not in either file — the registry is its source of truth, so it
can also be driven from a script:

```powershell
ai-terminal-pet --autostart        # report the current state
ai-terminal-pet --autostart=on     # run at logon
ai-terminal-pet --autostart=off    # stop running at logon
```

## Build from source (Windows, Go 1.23+)

```powershell
go vet ./...
go test -count=1 ./...
go build -trimpath -ldflags "-s -w" -o .\bin\AI-Terminal-Pet.exe ./cmd/ai-terminal-pet
go build -trimpath -ldflags "-s -w" -o .\bin\pet-statusline.exe ./cmd/pet-statusline
.\install.ps1   # stops the running pet, copies both binaries, SHA256-verifies
```

`install.ps1` retries the copy: the reporters relaunch the overlay on demand,
so it can be alive again milliseconds after being stopped, and a plain
copy-over fails at random on a machine that is actually using the tool.

`node --check reporters/*.js` covers the reporters (stdlib only).

### Seeing a layout change without a window

Layered-window geometry is easy to get subtly wrong and hard to look at
because there is nothing to screenshot. The rasterizer can be driven straight
to PNG instead:

```powershell
$env:PET_PREVIEW_DIR = "$env:TEMP\petpreview"
go test ./internal/ui -run TestRenderPreview
```

That writes the compact and expanded panel in several states at 2–3x scale.
It is a development aid, not an assertion, so it skips unless
`PET_PREVIEW_DIR` is set.

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
internal/ui            layered overlay, canvas rasterizer, panel layout, tray
reporters/             JS push side (incl. generic.js)
install.ps1            stop → copy (retrying) → SHA256 verify (both binaries)
```

## Known limitations

- Windows only.
- Per-launch AV scanning can make Node-based reporters slow; use the
  native shim and/or add an AV exclusion for the install directory.
- Same-user pipe preemption is mitigated (mutex-first ordering), not
  eliminated — a future `FILE_FLAG_FIRST_PIPE_INSTANCE` or
  challenge-response upgrade is tracked in HANDOVER.md.
- `go test -race` needs a gcc toolchain. The UI has unit tests for pure
  logic (hit testing, panel sizing, release decisions, text fitting) and a
  PNG preview renderer for layout, but no automated pixel regression.

See [HANDOVER.md](HANDOVER.md) for the operational runbook and
[ROADMAP.md](ROADMAP.md) for the extension and lightweighting plans.

## License

MIT — see [LICENSE](LICENSE). Third-party modules used as libraries only:
`github.com/Microsoft/go-winio` (MIT), `github.com/lxn/win` (BSD-style),
`golang.org/x/sys` (BSD-style). No third-party code is vendored.
