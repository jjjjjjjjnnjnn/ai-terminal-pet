# FACT SHEET — ai-terminal-pet v0.9.0 (verified from the repository, 2026-11-09)

Everything below was read out of the working tree. Anything not listed here is unknown and must not be written.

## Identity
- Name: ai-terminal-pet
- Version: 0.9.0  (source: internal/app/app_windows.go:23  const Version = "0.9.0")
- Platform: Windows only (Win32 layered window + Windows named pipes). No macOS, no Linux.
- Language: Go 1.23.0 (go.mod), reporters are plain Node.js with zero npm dependencies (standard library only).
- License: MIT. Third-party modules used as libraries only, nothing vendored:
  github.com/Microsoft/go-winio v0.6.2 (MIT), github.com/lxn/win (BSD-style), golang.org/x/sys v0.10.0 (BSD-style, indirect). That is 3 third-party modules total.
- What it is: a tiny always-on-top floating overlay that shows the state of the AI coding terminals you already run.

## Measured numbers (do not round up, do not embellish)
- Go source (excluding tests): 10,201 lines. Test code: 3,831 lines. 140 test functions (grep -c "func Test").
- reporters/*.js: 1,373 lines. All .js/.mjs in the repo including the DeepSeek Harness bridge: 1,742 lines.
- bin/AI-Terminal-Pet.exe: 4,178,432 bytes (~3.98 MiB), built with -trimpath -ldflags "-s -w".
- bin/pet-statusline.exe: 3,644,416 bytes (~3.48 MiB).
- Runtime footprint with live sessions: working set roughly 22–38 MB (shared pages included), 8–16 threads. (ROADMAP §0 measurement, still the documented figure.)
- CPU: a 120-second sampled run of continuous waiting_permission state measured 0.562 s cumulative CPU, about 0.5% of one core. (HANDOVER §21.5, round 3.)
- Idle cadences: session reconcile check every 30 s, snapshot scan every 5 s, presence probe (one Toolhelp snapshot + one EnumWindows pass) every 10 s shared by all sessions.
- Termination: overlay self-destructs only after 10 minutes with no session at all; with terminals open it effectively stays resident.

## States (exactly these seven)
idle · thinking · working · tool_use · waiting_permission · error · closed
The overlay shows the highest-priority state across all sessions. A session that stops reporting is removed, not shown as stale.

## Transport / protocol
- Per-user Windows named pipe: \\.\pipe\ai-terminal-pet-v1-<first 8 hex of SHA-256 of the home directory>. Go (internal/pipeserver.PipePathForUser) and JS (reporters/protocol.js:pipeScopeToken) compute it with the same algorithm; a cross-language golden test locks it.
- One JSON object per line, schemaVersion "ai-terminal-pet.v1", payload cap 1 MiB, server replies exactly OK or ERR, client retries until a deadline and never interprets a partial reply as success.
- Server: 32 concurrent connections, 5 s read timeout.
- Snapshot files under %LOCALAPPDATA%\AI-Terminal-Pet\runtime\ act as a fallback channel and as the on-disk record; the disk snapshot is authoritative for session existence (a session that was on disk and is gone is over).
- Store: max 256 sessions, pipe TTL 30 s, snapshot TTL 15 s, tombstone classes for closed/expired/evicted.

## Session existence model (this is the interesting engineering claim)
- Existence is decided by OS observation (process table + terminal window titles), not by reporter heartbeats. State can be arbitrarily stale.
- Deletion requires the host to be continuously unobservable for 2 minutes (DefaultAbsenceGrace); a single failed probe never deletes.
- A session whose terminal tab title matches its identity stays alive indefinitely, however quiet.
- Synthetic PIDs from hooks never participate in liveness probing.

## The overlay
- Compact panel: 56x156 logical pixels, anchored to a screen edge. Shows state orb (colour + glyph, pulses while something waits), session count, the state word, the project name that is waiting.
- Expanded panel: 200 px wide, height follows content, 1–6 rows then an overflow line ("N more sessions"). Each row: project, state, source, relative time.
- Gestures: click anywhere = compact/expanded toggle; drag anywhere = move 1:1 with the cursor; wheel = zoom 65%–160%; hover = panel brightens; Esc = collapse; right-click = menu.
- Edge retract: when docked to a screen edge, the panel slides mostly off-screen leaving a thin sliver; hovering the sliver brings it back. A session waiting for you or one that errored always stays fully visible.
- Single instance via a named mutex. Starting it again reveals the running instance instead of exiting silently.
- Jump to terminal: when a session needs you, click the "↗ jump" strip (compact, only when a jumpable session exists) or the arrow on an expanded row, and the matching terminal window is brought to the foreground. Windows Terminal tabs are separate top-level windows, so the tab is found by walking the session's process ancestry and matching window titles; a weak match (score < 40) is refused rather than focusing an unrelated window. Supported caveat: classic conhost multi-session windows and VS Code integrated terminals resolve only to the window level. Not documented in the README before this draft, but implemented (internal/platform/focus_windows.go, internal/ui/focus_windows.go, HANDOVER §20).

## Tray menu
show/hide · reset position and size · Settings… · status alerts on/off · sound on/off · always on top · edge retract · Language (Auto (System) / English / 简体中文) · auto-approve permissions after N seconds · run at logon · quit

## Settings window (v0.9.0)
- Second top-level window, same message pump as the overlay, three tabs; every change applies immediately, no OK button.
  - General: status alerts, always on top, edge retract, panel opacity, language, run at logon.
  - Sound: sound on/off, scheme (chime / bell / soft / pulse), volume 0–100, separate gates for waiting / error / finish, preview.
  - Sessions: auto-approve, countdown length (1–60 s), restore defaults.
- Opening a scheme previews it immediately. Writes are merged on a background goroutine; a value this process wrote is not echoed back by the 30 s config hot-reload.

## Sound
- Tones are synthesized in memory as 44.1 kHz 16-bit mono PCM WAV and played with PlaySoundW(SND_MEMORY|SND_ASYNC|SND_NODEFAULT). Four schemes, each with three waveforms (waiting = two rising notes ~0.9 s, error = falling low tone, finish = a short light rising three-note phrase).
- Volume from config (0–100, default 85). Notifications and sound are separate switches because a balloon is read later and a tone is heard now.

## Auto-approve (off by default)
- When enabled, a permission request the hook reports gets a veto window on the expanded panel ("Allow (5s)" plus a draining bar): tap the panel to deny, silence approves after the countdown.
- Never auto-approves dangerous commands. reporters/protocol.js:364 DANGEROUS_PATTERNS covers: rm -r/rm -rf, rd /s, mv ... /dev/null, mkfs, dd of=/dev/..., fork bombs :(){ , git push --force, format C:, del ... \Windows, shutdown /s, Remove-Item ... System32. A match forces the manual path (Claude: falls through to the terminal's own prompt; Antigravity: denies).
- Only covers the permission-prompt path of the Claude and Antigravity hooks; ordinary tool calls are never delayed. If the hook cannot read config.json it treats auto-approve as off.

## CLI (exact, from cmd/ai-terminal-pet/main.go usage)
- ai-terminal-pet                 start the overlay (no-op if already running)
- ai-terminal-pet --show          reveal the running overlay
- ai-terminal-pet --toggle        switch between compact and expanded
- ai-terminal-pet --reset         re-anchor to the screen edge at 100% zoom
- ai-terminal-pet --settings      open the settings window
- ai-terminal-pet --quit          stop the running overlay
- ai-terminal-pet --json          print the current sessions as JSON (add --pretty)
- ai-terminal-pet --autostart     report the run-at-logon state
- ai-terminal-pet --autostart=on / =off
- ai-terminal-pet --version / --help
- --json reads the reporter snapshots on disk, so it works whether or not the overlay is running and can never block on a GUI process.
- Usage errors go to stderr with exit code 2, never into error.log.

## Reporters (reporters/)
- opencode.js — OpenCode plugin (file:// plugin entry)
- claude-hook.js — Claude Code hooks: SessionStart, SessionEnd, UserPromptSubmit, PreToolUse, PostToolUse, Stop, Notification, PermissionRequest
- antigravity-hook.js — Antigravity agent/tool events
- antigravity-statusline.js — Node status line reporter
- generic.js — one-line bridge for any other CLI: node reporters/generic.js --source=aider --state=working --project="$PWD"; also --stdin and --reason
- protocol.js — single protocol source: validation, redaction, retry-until-deadline, fail-fast, never throws into the host CLI
- cmd/pet-statusline — native Go replacement for the Node status line. On machines where AV hooks make every node spawn cost seconds, the native shim answers well under a second (measured ~0.19 s steady state after the v0.7 rebuild).

## DeepSeek Harness integration
- integrations/deepseek-harness/ai-terminal-pet-bridge/ : a Cordis bundle (package.json, cordis.patch.yml, index.js, locale/en.json, locale/zh.json, icon.svg, README.md). Adapter only: it subscribes to session/event, agent/status, agent/error, agent/inbox/claimed, approval/request (waterfall, always calls next()), maps them onto the official state set, source/terminal = deepseek-harness, pid = stableProcessId (synthetic), parentPid = real host.
- install.ps1 syncs the bundle to %LOCALAPPDATA%\AI-Terminal-Pet\bundles\ai-terminal-pet-bridge\ with a SHA256 check. It is installed through DeepSeek Harness's own plugin manager (install_bundle with that absolute directory), not by editing profile files.
- integrations/deepseek-harness/verify.mjs runs 14 assertions and prints VERIFY PASSED.

## Install and build
- Build: go vet ./... ; go test -count=1 ./... ; go build -trimpath -ldflags "-s -w" -o .\bin\AI-Terminal-Pet.exe ./cmd/ai-terminal-pet ; go build -trimpath -ldflags "-s -w" -o .\bin\pet-statusline.exe ./cmd/pet-statusline ; then .\install.ps1
- install.ps1: stops the running overlay, waits for the image to actually unload, copies with retry (a reporter can relaunch the overlay milliseconds after it is stopped), verifies SHA256 of both binaries, and syncs protocol.js plus the DSH bundle and verifies those too.
- No install.ps1 equivalent for other platforms: Windows only.
- The layout preview: $env:PET_PREVIEW_DIR = "$env:TEMP\petpreview"; go test ./internal/ui -run TestRenderPreview — writes compact and expanded panels at 2–3x scale so layout changes can be inspected without a screenshot.

## Screenshots — only these files exist under docs/assets/
- docs/assets/hero.png — 1600x640 hero banner
- docs/assets/overlay-compact.png — 168x468 (compact panel, Chinese labels)
- docs/assets/overlay-expanded.png — 400x424 (expanded panel, Chinese labels, 3 sessions)
- docs/assets/overlay-expanded-many.png — 400x676 (expanded panel with 6 rows plus overflow line)
- docs/assets/overlay-auto-approve.png — 400x424 (expanded panel with the 5 s veto countdown and draining bar)
- Also present but not required: overlay-compact-en.png, overlay-compact-hover.png, overlay-expanded-en.png, and render_assets.py (the generator).
- All of them are rasterized by the real Win32 renderer via the layout preview test, so they cannot drift from the shipped layout. They are not mockups.

## Known limitations (write these honestly)
- Windows only.
- Antivirus scanning makes every process creation expensive on some machines; the Node reporters spawn a process per event. The Go status line shim exists for that; an AV exclusion for the install directory is the only thing that reliably gets it under a second.
- Same-user named-pipe squatting is mitigated (mutex first) but not eliminated; a FILE_FLAG_FIRST_PIPE_INSTANCE or challenge-response upgrade is documented in docs/HANDOVER.md but not implemented.
- go test -race needs a gcc toolchain and is not run. UI geometry, hit testing, layout and text fitting are unit-tested, and the PNG preview renders real layout, but there is no automated pixel regression test.
- Auto-approve countdown length is read at startup: changing it needs an overlay restart, unlike every other setting, which hot-reloads.
- Windows Terminal tabs can be focused; classic conhost multi-session windows and VS Code integrated terminals only resolve to the window level.
- The status line shim answers in well under a second on a machine where node costs seconds, but that is still not zero: the cost is process creation, not our code.

## Roadmap (docs/ROADMAP.md, real, unfinished)
- Native hooks: one pet-hook.exe with --mode=claude|antigravity instead of spawning node per event.
- docs/PROTOCOL.md so a third party can write a compliant reporter without reading the source.
- Cross-platform: extract the window layer behind a Presenter interface, Unix domain socket transport.
- Clickable session rows (focus the terminal), --doctor self-check, per-source colours.
- Binary size: both executables under 5 MiB together; main binary target under 2.5 MiB.
- Non-goals: no terminal multiplexer or log viewer, no account/cloud/telemetry (all data stays in %LOCALAPPDATA%), no UI framework, no cross-user shared pipe.

## Wording rules for the copy (hard bans)
- Banned words/phrases: revolutionary, blazingly fast, blazing, supercharge, game-changer, 10x, seamless, effortless, cutting-edge, next-generation, "unlock", "powerful", "delightful", "magical", "simply", "just works".
- No exclamation marks in headings. At most 6 emoji in the whole English document, 6 in the Chinese one.
- Second person ("you", "your terminal"). Short sentences. Scannable headings. Concrete behaviour instead of adjectives: say "a tap denies the request, silence approves after 5 seconds" rather than "intuitive approval flow".
- Never state a number that is not in this fact sheet. When unsure, omit.
- Never invent a feature, a platform, a benchmark or a user quote.
- Do not claim the overlay is a terminal multiplexer, a log viewer, a cloud service, or cross-platform.
