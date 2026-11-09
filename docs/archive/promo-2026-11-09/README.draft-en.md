# ai-terminal-pet

You stop glancing across terminal windows to track agent states. It costs one small executable with no account and no cloud.

![ai-terminal-pet](docs/assets/hero.png)

![Platform](https://img.shields.io/badge/platform-Windows-blue) ![Go](https://img.shields.io/badge/go-1.23-blue) ![Version](https://img.shields.io/badge/version-0.9.0-blue)
![License](https://img.shields.io/badge/license-MIT-green) ![Size](https://img.shields.io/badge/size-3.98%20MiB-blue) ![Dependencies](https://img.shields.io/badge/dependencies-3-blue) ![Tests](https://img.shields.io/badge/tests-140-brightgreen)

## The problem

You run Claude Code, OpenCode, or Antigravity in multiple terminal windows at once. A background task prompts for permission to execute a tool, but you are focused in another window and never notice the prompt.

Fourteen minutes later, you switch back only to discover that your build has been sitting idle the entire time. Glancing across multiple terminal tabs wastes focus and breaks your concentration.

ai-terminal-pet anchors a 56px strip to the screen edge. It aggregates your sessions in real time, pulses when an agent needs permission, and displays which project is waiting and for how long.

The table below contrasts the everyday workflow before and after using the overlay:

| Context | Before | After |
|---|---|---|
| Multiple terminals | Four terminal windows open; you lose track of which window needs attention. | One 56px strip on the edge of the screen aggregates every active session. |
| Blocked build | You find out a build was blocked on a permission prompt fourteen minutes ago. | The overlay pulses immediately and shows which project is waiting and for how long. |
| Context switching | You click across tabs to check whether builds and tasks are still running. | A single glance reveals current states without touching your keyboard or mouse. |

## What it looks like

![Compact strip](docs/assets/overlay-compact.png)
*Compact strip: pulsing state orb with status glyph, session count, state word, and waiting project name.*

![Expanded list](docs/assets/overlay-expanded.png)
*Expanded list: active sessions showing per-row project name, state, source hook, and relative time.*

![Many sessions](docs/assets/overlay-expanded-many.png)
*Expanded list with many sessions: six session rows followed by the overflow line showing remaining sessions.*

![Auto-approve countdown](docs/assets/overlay-auto-approve.png)
*Auto-approve countdown: draining bar with "Allow (5s)" prompt allowing manual veto before automatic permission approval.*

---

## Features

The table below outlines the core capabilities and exact runtime parameters of the application:

| Feature | What it actually does |
|---|---|
| State aggregation | Aggregates sessions across 7 states (`idle`, `thinking`, `working`, `tool_use`, `waiting_permission`, `error`, `closed`), showing the highest-priority state across all active sessions. |
| Compact and expanded panels | Compact panel occupies 56x156 logical pixels anchored to a screen edge; clicking expands to 200 px wide, displaying 1–6 rows before adding an overflow line. |
| Whole-panel gestures | Click anywhere toggles compact and expanded views, drag anywhere moves 1:1 with the cursor, and mouse wheel adjusts zoom between 65% and 160%. |
| Edge retract | Slides mostly off-screen when docked at an edge leaving a thin sliver; hovering restores it, while waiting_permission or error states override retraction to remain visible. |
| Tray menu | Right-click tray menu toggles visibility, resets position and zoom, opens settings, switches language, toggles autostart, and configures auto-approve. |
| Settings window | Native three-tab window (General, Sound, Sessions) where every change persists immediately to `config.json` without an OK button. |
| Synthesized alert tones | Generates 44.1 kHz 16-bit mono PCM WAV tones in memory across four sound schemes (`chime`, `bell`, `soft`, `pulse`), with separate gates for waiting, error, and finish. |
| Per-session alerts | Fires Windows tray notification balloons and audio alerts on transitions into `waiting_permission`, `error`, or finish states. |
| Auto-approve and denylist | Provides a 5-second veto window on permission prompts where a tap denies and silence approves; blocks commands matching the dangerous-command denylist. |
| Jump to terminal | Walks session process ancestry and matches window titles to bring the waiting terminal window or tab to the foreground. |
| Scriptable CLI | Outputs structured session state via `ai-terminal-pet --json` by reading disk snapshots, avoiding blocking on GUI processes. |
| Single instance | A named mutex makes a second launch activate the running overlay instead of starting another one. |
| Connection capacity | The pipe server accepts 32 concurrent connections with a 5 s read timeout, so IPC never blocks the UI thread. |
| Idle lifecycle | A session is dropped only after a 2-minute absence grace; with no session left, the process removes itself after 10 minutes. |

## How it works

```text
Host CLI (Claude Code / OpenCode / Antigravity)
       │
       ▼ (Fire-and-forget hooks)
reporters/*.js
       │
       ▼ (JSON line over Windows named pipe: \\.\pipe\ai-terminal-pet-v1-<8hex>)
AI-Terminal-Pet.exe (Go 1.23, single-instance mutex)
       │
       ├─► internal/pipeserver (32 connections, 5s timeout, OK/ERR response)
       ├─► internal/store (in-memory state, TTL, disk snapshots fallback)
       └─► internal/ui (Win32 layered window, GDI topmost canvas)
```

- **Per-user named pipe**: IPC flows through `\\.\pipe\ai-terminal-pet-v1-<8hex>`, where `<8hex>` is derived from the SHA-256 hash of the user home directory. The exact same pipe-name algorithm is implemented in Go and JavaScript, locked by a cross-language golden test.

- **JSON line protocol**: Sessions transmit one JSON line per message with a 1 MiB payload cap. The server replies with an exact `OK` or `ERR` acknowledgment, and reporters retry until a deadline without treating partial reads as success.

- **On-disk snapshots**: Snapshot files under `%LOCALAPPDATA%\AI-Terminal-Pet\runtime\` provide a fallback channel and serve as the authoritative on-disk record of active sessions.

- **OS observation over heartbeats**: Session existence is decided by operating system observation of the process table and window titles rather than reporter heartbeats. A session remains active as long as its terminal window exists, and removal requires the host to be unobservable for a continuous 2-minute grace period.

- **Disks beat heartbeats**: The disk snapshot is authoritative for session existence: a session that was on disk and is gone is over, while synthetic PIDs from hooks never participate in liveness probing.

- **Connection handling**: The server accommodates 32 concurrent connections with a 5s read timeout, keeping IPC processing off the UI thread.

- **GDI rendering**: A top-down 32bpp DIB section created with CreateDIBSection is painted through a memory DC and handed to the layered window with UpdateLayeredWindow.

## Install

Installation takes about 30 seconds on Windows using PowerShell:

1. Build both executables:
   ```powershell
   go build -trimpath -ldflags "-s -w" -o .\bin\AI-Terminal-Pet.exe ./cmd/ai-terminal-pet
   go build -trimpath -ldflags "-s -w" -o .\bin\pet-statusline.exe ./cmd/pet-statusline
   ```

2. Run the installer to copy binaries and sync assets:
   ```powershell
   .\install.ps1
   ```

3. Verify that the installation works:
   ```powershell
   ai-terminal-pet --json
   ```

Installing a terminal hook is required for that terminal to report; see the next section to connect your tools.

## Connect your terminal

| Terminal | How to connect | What you get |
|---|---|---|
| OpenCode | Add `file:///C:/path/to/ai-terminal-pet/reporters/opencode.js` to `plugin` in `opencode.json` | Automatic session tracking for tool use, prompt handling, and task completion |
| Claude Code | Add `node <repo>/reporters/claude-hook.js` for `SessionStart`, `SessionEnd`, `UserPromptSubmit`, `PreToolUse`, `PostToolUse`, `Stop`, and `Notification` in `settings.json` | Real-time status updates, tool execution tracking, and permission prompt alerts |
| Antigravity | Set `statusLine.command` to `pet-statusline.exe` and configure `antigravity-hook.js` for agent and tool events | Fast statusline updates, agent state changes, and permission request forwarding |
| DeepSeek Harness | Install the Cordis bundle from `%LOCALAPPDATA%\AI-Terminal-Pet\bundles\ai-terminal-pet-bridge\` via the plugin manager (`install_bundle`) | turn/start to working, tool/call to tool_use, approval/asked to waiting_permission, tool/result with an error to error, turn/end to idle, process exit to closed |
| Anything else | Run `node reporters/generic.js` with `--source`, `--state`, and `--project` flags | Instant status reporting from any custom script or CLI tool |

The Harness side is a host-only Cordis bundle: it holds no pipe-name, state-set or snapshot logic of its own, it loads the protocol from `%LOCALAPPDATA%\AI-Terminal-Pet\protocol.js` and delegates everything to it.

Run `node reporters/generic.js` directly to report from any other CLI or script:

```powershell
node reporters/generic.js --source=aider --state=working --project="$PWD"
node reporters/generic.js --source=codex --state=waiting_permission --reason=approve
echo '{"state":"error","reason":"build failed"}' | node reporters/generic.js --source=ci --stdin
```

- `--source`: sets the reporter identity name.
- `--state`: sets one of the seven official session states.
- `--stdin`: parses a JSON payload piped directly from standard input.

## Control it from a script

| Command | Effect |
|---|---|
| `ai-terminal-pet` | Start the overlay (no-op if already running) |
| `ai-terminal-pet --show` | Reveal the running overlay |
| `ai-terminal-pet --toggle` | Switch between compact and expanded views |
| `ai-terminal-pet --reset` | Re-anchor to the screen edge at 100% zoom |
| `ai-terminal-pet --settings` | Open the settings window |
| `ai-terminal-pet --quit` | Stop the running overlay |
| `ai-terminal-pet --json` | Print the current sessions as JSON (add `--pretty` for formatted output) |
| `ai-terminal-pet --autostart` | Report the run-at-logon state |
| `ai-terminal-pet --autostart=on` | Enable run at logon |
| `ai-terminal-pet --autostart=off` | Disable run at logon |
| `ai-terminal-pet --version` | Print version information |
| `ai-terminal-pet --help` | Print CLI usage help |

Run `--json` to inspect active sessions directly from scripts or shell status lines:

```powershell
ai-terminal-pet --json --pretty
```

```json
{
  "schemaVersion": "ai-terminal-pet.v1",
  "state": "waiting_permission",
  "count": 2,
  "generatedAt": "2026-11-09T21:30:00Z",
  "sessions": [{"project": "ai-terminal-pet", "state": "waiting_permission", "source": "claude-hook"}]}
```

The `--json` command never blocks on a GUI process because it reads session snapshots directly from disk.

## Safety

Terminal hooks run inside your development environment, so their boundaries are strict:

- **Fail-fast reporters**: Reporters run fire-and-forget, catch errors internally, and never throw exceptions or block your host CLI. Any reporter failure terminates silently without interrupting user keystrokes or host terminal operations.

- **Auto-approve disabled by default**: The auto-approve feature starts turned off; activating it requires explicit configuration. A terminal hook that cannot read the configuration file automatically defaults to off.

- **Configurable veto window**: When auto-approve is active, permission prompts display a draining bar on the expanded panel; a single tap denies the prompt, while silence approves after the countdown. Ordinary tool calls without permission prompts are never delayed.

- **Dangerous-command denylist**: Destructive commands never auto-approve and always fall back to manual prompts, blocking patterns such as `rm -rf`, `git push --force`, `format C:`, `rd /s`, and `shutdown /s`. A matched command forces manual review in the terminal.

- **Zero network activity**: The application makes no network calls, contains no telemetry, and stores all configuration and runtime snapshots under `%LOCALAPPDATA%\AI-Terminal-Pet\`. No code or logs leave your local machine.

- **Per-user ACL protection**: The named pipe DACL `"D:P(A;;GRGW;;;OW)(A;;GRGW;;;SY)"` grants read/write access to the owner of the pipe and to LocalSystem only, and nothing for any other account on the machine. Other local users therefore cannot connect and cannot occupy the pipe name.

## Known limits

- **Windows only**: The overlay relies directly on Win32 layered windows and Windows named pipes; macOS and Linux are not supported.

- **Process creation overhead**: Antivirus scanning hooks can make each Node.js process launch take roughly 1s; using the native `pet-statusline.exe` shim or adding an AV exclusion for the install folder reduces this latency. The delay comes from process scanning rather than application code.

- **Named-pipe squatting**: Same-user pipe preemption is mitigated by acquiring a named mutex first, but it is not completely eliminated. A full challenge-response handshake remains documented for future hardening.

- **Auto-approve countdown restart**: Changing the auto-approve countdown duration requires an overlay restart, while all other settings hot-reload automatically.

- **No pixel regression tests**: Pure UI logic, hit testing, and layout bounding boxes are unit-tested, but automated visual pixel regression testing is not in place. Visual verification relies on the PNG preview generator.

- **Window focus granularity**: Session jump focusing resolves tabs in Windows Terminal, but classic conhost windows and VS Code integrated terminals resolve only to the parent window level.

## Roadmap

- **Native hook binary**: A unified `pet-hook.exe` executable to eliminate per-event Node.js spawns for Claude Code and Antigravity (planned).

- **Protocol specification**: A formal `docs/PROTOCOL.md` document defining the wire protocol for custom reporter implementations (planned).

- **Cross-platform support**: Extracting the UI layer behind a presenter interface with Unix domain socket transport for macOS and Linux (not started).

- **Binary size reduction**: Bring both executables under 5 MiB combined, targeting under 2.5 MiB for the main binary (currently 3.7 MiB) by stripping reflection metadata, dropping unused packages, and replacing the regexp-based redaction in pet-statusline with a hand-written scanner (planned).

## Non-goals

- No accounts, no cloud sync, no telemetry; everything stays in `%LOCALAPPDATA%`.
- Not a terminal multiplexer and not a replacement for your terminal — it only tells you which session wants attention.

## License

Distributed under the MIT License. See [LICENSE](LICENSE) for details. Third-party dependencies are used as libraries only with zero vendored code: github.com/Microsoft/go-winio (MIT), github.com/lxn/win (BSD-style), and golang.org/x/sys (BSD-style).

## Develop from source

Run verification commands when modifying code:

```powershell
go vet ./...
go test -count=1 ./...
node --check reporters/*.js
node integrations/deepseek-harness/verify.mjs
$env:PET_PREVIEW_DIR = "$env:TEMP\petpreview"
go test ./internal/ui -run TestRenderPreview
```
