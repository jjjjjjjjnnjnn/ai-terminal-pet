# Platform support

Three platforms compile cleanly with `CGO_ENABLED=0` across all 16 Go packages and zero Node.js npm dependencies. Every row below records what is implemented and how it was checked.

| Capability | Windows (Win32) | Linux (X11) | macOS (AppKit / Cocoa) | How verified |
|---|---|---|---|---|
| **Toolchain & runtime dependencies** | Native Win32 (`CGO_ENABLED=0`, no GTK/WebView) | Pure-Go `xgb` X11 client (`CGO_ENABLED=0`, no cgo/libgtk) | Pure-Go `purego` Objective-C / CoreGraphics (`CGO_ENABLED=0`, no cgo) | `scripts/cross-build.sh` + CI on `windows-latest`, `ubuntu-latest`, `macos-latest` |
| **Reporters (`reporters/*.js`, `pet-hook`, `pet-statusline`)** | Named pipe `\\.\pipe\ai-terminal-pet-v1-<8hex>` + snapshots | Unix domain socket `$XDG_RUNTIME_DIR/ai-terminal-pet/v1-<8hex>.sock` + snapshots | Unix domain socket `$TMPDIR/ai-terminal-pet-v1-<8hex>.sock` + snapshots | `go test -count=1 ./...` on Win + Ubuntu CI + macOS CI + WSL2 e2e |
| **Host discovery (`--doctor`, `--connect`)** | `%APPDATA%` / `%USERPROFILE%` + `.exe` shims | `$XDG_CONFIG_HOME` / `$HOME` + POSIX shims | `$HOME/.config` / `$HOME` + POSIX shims | `internal/connect/connect_test.go` + `connect_unix_test.go` + headless CI `--doctor`/`--connect` |
| **Single instance & control IPC** | Per-user named mutex + `WM_USER` / IPC | `flock` lock file + Unix domain control socket | `flock` lock file + Unix domain control socket (`<104B` `sun_path` safe) | Win32 unit tests + Linux X11 + `TestDarwinControlSocketRoundTrip` on `macos-latest` |
| **Floating panel (`internal/ui`)** | Full layered ARGB DIB window + DPI + dock/auto-hide + theme + F2 keyboard | Pure-Go 32-bit ARGB X11 utility window (`PutImage`, `_NET_WORKAREA`, XRandR, `_MOTIF_WM_HINTS`, F2 keyboard) | Pure-Go borderless floating `NSWindow` (`NSImageView`, `NSStatusBar`, F2/`Cmd+±/0` zoom, dock, drag, mute) | Win32 full suite + real WSLg live smoke (`DISPLAY=:0`) + `macos-latest` headless AppKit/geometry tests (`PET_COCOA_PROBE=1` opt-in for WindowServer); **plus `render_invariants_test.go` on all three runners**: the same four assertions run against whichever rasteriser that OS built (frame geometry, transparent corners, body filled, text actually rasterised, state hue present, states/languages/themes distinguishable) - invariants only, never pixel equality |
| **Text rasterization (`internal/ui`)** | Win32 GDI `DrawTextW` / `ClearType` | Pure-Go OpenType glyph rasterizer (`golang.org/x/image/font/opentype`) scanning `/usr/share/fonts` + Windows fallback | Pure-Go OpenType rasterizer scanning `/System/Library/Fonts`, `/Library/Fonts`, `~/Library/Fonts` | `TestLinuxFontRasterization` + OpenType unit tests on Win/Ubuntu/macOS |
| **Settings UI** | Dedicated four-tab Win32 settings window (`AITerminalPetSettings`) | **Terminal editor**: `--settings` prints every option with its value and edits them when a terminal is attached; `--set key=value` for scripts (`internal/settingscli`) | Same terminal editor as Linux (`internal/settingscli`); same `--set` | `internal/settingscli` unit tests (every persisted key must have an entry, refusals, scripted editor runs) + a real pty run in WSL for the interactive path + `--set` verified on Windows and Linux |
| **Tray / menu bar icon** | `Shell_NotifyIconW` tray icon + context menu | Desktop-neutral panel gestures (right-click/middle-click/dock); no XEmbed tray | `NSStatusBar` menu bar item (`Show/Hide`, `Reset Position`, `Quit`) | Win32 live tray + macOS `app_darwin.go` bindings |
| **Process & window presence / focus jump** | Win32 process snapshot + `HWND` enumeration + `SetForegroundWindow` | `/proc` scan + X11 `_NET_CLIENT_LIST` + MIT-SCREEN-SAVER idle + WM activation | `sysctl` `kinfo_proc` + `CGWindowListCopyWindowInfo` + `NSWorkspace` / `NSRunningApplication.activateWithOptions:` | `internal/platform` tests across Win / Linux / macOS CI |
| **Audio cues** | Synthesized WAV PCM via `winmm` `PlaySoundW` (`sound_windows.go`); fires on the alert plan **and** on the six jump/copy/settings confirmation points (`beep(beepExclamation)` / `beep(beepFinished)`) | Synthesized WAV handed to the first player found on `PATH` — `pw-play -`, `paplay` (stdin), `aplay -q -` (`sound_linux.go`); fires on the alert plan and on the copy/jump failure cues. **No player installed = silence**, reported by `--doctor` | Synthesized WAV played async via AppKit `NSSound` (`sound_darwin.go`) from the alert plan and the copy/jump failure cues; the settings-window audition points do not exist (there is no settings window) | `TestToneForTransition` on all 3 OSes (pure synthesis); `TestPlaySoundLive` Win32 live probe; Linux probe/argv/stdin tests (`sound_linux_test.go`); coverage per platform measured in §"Config keys" below |
| **Clipboard copy** | Win32 `OpenClipboard` / `SetClipboardData` | X11 `CLIPBOARD` `SelectionRequest` owner | `NSPasteboard` general pasteboard (`clipboard_darwin.go`) | Win32 + Linux X11 + macOS unit bindings |
| **Autostart** | `HKCU\...\Run` registry key (`--autostart=on/off`) | XDG `~/.config/autostart/ai-terminal-pet.desktop` | LaunchAgent `~/Library/LaunchAgents/io.github.jjjjjjjjnnjnn.ai-terminal-pet.plist` | `internal/autostart` unit tests on Windows + Ubuntu + macOS CI |
| **Installer & release archives** | `pet-setup.exe` (also `pet-setup --uninstall --force`), `ai-terminal-pet-<ver>-win-x64.zip`, `scripts/install.cmd` (`install.ps1`) | `ai-terminal-pet-<ver>-linux-{amd64,arm64}.tar.gz`, `scripts/install.sh` | `ai-terminal-pet-<ver>-darwin-{arm64,amd64}.tar.gz`, `scripts/install.sh` | Packaged & sha256-verified on Windows + WSL2 Ubuntu + `ubuntu-latest` + `macos-latest` CI |
| **Uninstall surface** | `Settings › Apps › AI Terminal Pet`: `install.ps1` is copied into the install directory and `HKCU\...\Uninstall\AI-Terminal-Pet` names it (`UninstallString` / `QuietUninstallString`, `-Uninstall -Force`), so removal works after the zip is deleted. `install.cmd -Uninstall` still asks. | `sh ./install.sh --uninstall` (stops the process, removes the directory, the `~/.local/bin` symlink and the `.desktop` entry) | `sh ./install.sh --uninstall` (same, plus the LaunchAgent) | `tests/fresh-machine.mjs` S5 + S8 run the recorded command, not a hand-built one, and assert the row and the directory both disappear |
| **DSH plugin (`packages/dsh-ai-terminal-pet`)** | Loads (`bundle.platforms` includes `win32`); copies the vendored overlay + `protocol.js` into `%LOCALAPPDATA%\AI-Terminal-Pet`, fully offline | Loads; copies `protocol.js` into `$XDG_DATA_HOME/ai-terminal-pet` and **fetches the pinned release archive once** (SHA-256 verified) for the overlay, installing it `0755` | Loads; same as Linux under `~/Library/Application Support/AI-Terminal-Pet`; the AppKit panel itself is built and unit-tested on `macos-latest` but never driven on Apple hardware | `packages/dsh-ai-terminal-pet/verify.mjs` on all three CI jobs (the Windows job pins the platform seam, Linux/macOS follow the real platform); a real headless session on **DSH 0.2.2-rc.5** (2026-11-19) loaded the bundle, received `session/event`, and wrote the snapshot (HANDOVER §86) |
| **Claude Code plugin (`packages/claude-ai-terminal-pet`)** | Loads; eight hooks call the native `pet-hook.exe`, and `SessionStart` copies the vendored overlay into `%LOCALAPPDATA%` offline (digest verified, never overwrites) | **Not supported**: the hooks are a Windows PE; use `ai-terminal-pet --connect=claude` (JS reporter) | **Not supported**: same; `--connect=claude` remains the path | `claude plugin validate` (plugin and marketplace, both clean) + `bootstrap.test.mjs` / `plugin.test.mjs` in the Windows CI job; a real marketplace install and a real session were measured on Windows 2026-11-18 |

## Config keys and where they actually take effect

`config.json` decodes every key on all three platforms (`internal/config`), and
`internal/app` carries the values into `Options` everywhere. What differs is
whether anything **reads** those values: a key that is accepted and then ignored
is the one shape of gap that reads as "my setting does nothing". This table comes
from the code (which file mentions each field, per build tag), and its last rows
are the backlog for round 62 - it is a record, not a promise.

| Key | Windows | Linux | macOS |
|---|---|---|---|
| `notifications` | balloon layer (`app_windows.go`) | freedesktop notification via `notify-send` (`notify_linux.go`); **no `notify-send` in PATH means no bubble** — the row still goes unread and `alerts.json` records `skipped=no-notification-surface` | `osascript -e 'display notification …'` (`notify_darwin.go`); implemented, **live display unverified** (no Apple hardware, and `UNUserNotificationCenter` needs a bundle identifier a bare binary does not have) |
| `sound` | gates the tone (`attention.go`) | gates the tone, which the probed player plays | gates a tone `NSSound` plays |
| `soundVolume` / `soundScheme` | applied to the synthesizer | applied to the synthesizer | applied (`app_darwin.go`) |
| `soundOnWaiting` / `soundOnError` | per-event gate | same gate, and the tone is played | same gate, and the tone is played |
| `notifyOnFinish` / `donePolicy` | decides the finish tone | plan only | plan only |
| `quietCurrent` / `quietSubagents` | policy (`attention.go`) | same policy (it suppresses marks too, so it works) | same policy |
| `foregroundAware` / `idleMinutes` | presence sampling | same | same |
| `alwaysOnTop`, `edgeRetract`, `opacity`, `theme`, `perSourceColor`, `language`, `hiddenAtLaunch`, `mutedSessions` | full | full (`app_linux.go`) | full (`app_darwin.go`) |
| `autoApprove` / `autoApproveTimeoutSec` | full | full | full |
| `settingsTextSize` | scales the settings window | stored only (there is no settings window) | stored only |
| `settingsKeyboard` / `settingsTooltips` | settings window | stored only | stored only |

Two rows used to be the round-62 backlog, written down rather than only known locally:
`notifications` (no balloon surface on Linux or macOS) and the whole `sound*`
block on Linux. Both were delivered in round 66: Linux plays through a probed
system player and shows alerts through `notify-send`, macOS shows alerts through
`osascript` (best-effort, live display unverified), and `--doctor` now prints
what each machine actually reaches (`audio` / `notifications` lines) instead of
leaving the answer in this table.

## Data directories and IPC endpoints

Implemented identically in Go (`internal/home`, `internal/pipeserver`) and Node.js (`reporters/protocol.js`) and locked by cross-language tests:

- **Windows**: `%LOCALAPPDATA%\AI-Terminal-Pet` · named pipe `\\.\pipe\ai-terminal-pet-v1-<sha256(lower(home))[0..8]>`
- **macOS**: `~/Library/Application Support/AI-Terminal-Pet` · socket `$TMPDIR/ai-terminal-pet-v1-<8hex>.sock`
- **Linux**: `$XDG_DATA_HOME/ai-terminal-pet` (or `~/.local/share/ai-terminal-pet`) · socket `$XDG_RUNTIME_DIR/ai-terminal-pet/v1-<8hex>.sock` (or `/tmp/ai-terminal-pet-<uid>-v1-<8hex>.sock`)

## CI cross-build contract

1. **GitHub Actions runners**: `windows-latest` (`verify` + `scan`), `ubuntu-latest` (`linux` — cross-build gate, `go vet`, `go test`, **`go test -race` with `CGO_ENABLED=1`** (the only job with a C toolchain, so the only place the race detector can run at all), Linux binaries, headless install/hook/CLI smoke, `.tar.gz` SHA-256 verification), and `macos-latest` (`macos` — native `darwin/arm64` Go tests, binary builds, headless install/hook/CLI smoke).
2. **`scripts/cross-build.sh`** asserts that all packages in [scripts/portable-packages.txt](../scripts/portable-packages.txt) (every package in this repository — 16 packages) build cleanly with `CGO_ENABLED=0` for `windows/amd64`, `linux/amd64`, and `darwin/arm64`, while [scripts/windows-only-packages.txt](../scripts/windows-only-packages.txt) stays empty.
3. **Host-windowed gaps still open cleanly**:
   - On Linux & macOS, edit `config.json` directly instead of opening a native settings dialog (`--settings` logs/prints guidance cleanly).
   - Linux audio and notifications depend on what the machine has: the overlay probes `pw-play`/`paplay`/`aplay` and `notify-send`, stays silent/quiet when none exists, and `--doctor` says which one is missing. Nothing has to be installed for the panel itself to work.
   - macOS notifications go through `osascript` and were never observed on real Apple hardware from this project; `--doctor` states that in the same line.
   - Live interactive WindowServer GUI on physical Apple hardware (and non-WSLg desktop environments such as GNOME/KDE/Wayland+XWayland) remains opt-in verified (`PET_COCOA_PROBE=1` on macOS, WSLg X11 live-tested on Linux).
