# Platform support

Three platforms compile cleanly with `CGO_ENABLED=0` across all 16 Go packages and zero Node.js npm dependencies. Every row below records what is implemented and how it was checked.

| Capability | Windows (Win32) | Linux (X11) | macOS (AppKit / Cocoa) | How verified |
|---|---|---|---|---|
| **Toolchain & runtime dependencies** | Native Win32 (`CGO_ENABLED=0`, no GTK/WebView) | Pure-Go `xgb` X11 client (`CGO_ENABLED=0`, no cgo/libgtk) | Pure-Go `purego` Objective-C / CoreGraphics (`CGO_ENABLED=0`, no cgo) | `scripts/cross-build.sh` + CI on `windows-latest`, `ubuntu-latest`, `macos-latest` |
| **Reporters (`reporters/*.js`, `pet-hook`, `pet-statusline`)** | Named pipe `\\.\pipe\ai-terminal-pet-v1-<8hex>` + snapshots | Unix domain socket `$XDG_RUNTIME_DIR/ai-terminal-pet/v1-<8hex>.sock` + snapshots | Unix domain socket `$TMPDIR/ai-terminal-pet-v1-<8hex>.sock` + snapshots | `go test -count=1 ./...` on Win + Ubuntu CI + macOS CI + WSL2 e2e |
| **Host discovery (`--doctor`, `--connect`)** | `%APPDATA%` / `%USERPROFILE%` + `.exe` shims | `$XDG_CONFIG_HOME` / `$HOME` + POSIX shims | `$HOME/.config` / `$HOME` + POSIX shims | `internal/connect/connect_test.go` + `connect_unix_test.go` + headless CI `--doctor`/`--connect` |
| **Single instance & control IPC** | Per-user named mutex + `WM_USER` / IPC | `flock` lock file + Unix domain control socket | `flock` lock file + Unix domain control socket (`<104B` `sun_path` safe) | Win32 unit tests + Linux X11 + `TestDarwinControlSocketRoundTrip` on `macos-latest` |
| **Floating panel (`internal/ui`)** | Full layered ARGB DIB window + DPI + dock/auto-hide + theme + F2 keyboard | Pure-Go 32-bit ARGB X11 utility window (`PutImage`, `_NET_WORKAREA`, XRandR, `_MOTIF_WM_HINTS`, F2 keyboard) | Pure-Go borderless floating `NSWindow` (`NSImageView`, `NSStatusBar`, F2/`Cmd+±/0` zoom, dock, drag, mute) | Win32 full suite + real WSLg live smoke (`DISPLAY=:0`) + `macos-latest` headless AppKit/geometry tests (`PET_COCOA_PROBE=1` opt-in for WindowServer) |
| **Text rasterization (`internal/ui`)** | Win32 GDI `DrawTextW` / `ClearType` | Pure-Go OpenType glyph rasterizer (`golang.org/x/image/font/opentype`) scanning `/usr/share/fonts` + Windows fallback | Pure-Go OpenType rasterizer scanning `/System/Library/Fonts`, `/Library/Fonts`, `~/Library/Fonts` | `TestLinuxFontRasterization` + OpenType unit tests on Win/Ubuntu/macOS |
| **Settings UI** | Dedicated four-tab Win32 settings window (`AITerminalPetSettings`) | Edit `config.json` directly (CLI `--settings` prints guidance until native settings UI arrives) | Edit `config.json` directly (`internal/ui/settings_windows.go` is Win32-only so far) | Verified in Win32 preview suite & POSIX headless behavior |
| **Tray / menu bar icon** | `Shell_NotifyIconW` tray icon + context menu | Desktop-neutral panel gestures (right-click/middle-click/dock); no XEmbed tray | `NSStatusBar` menu bar item (`Show/Hide`, `Reset Position`, `Quit`) | Win32 live tray + macOS `app_darwin.go` bindings |
| **Process & window presence / focus jump** | Win32 process snapshot + `HWND` enumeration + `SetForegroundWindow` | `/proc` scan + X11 `_NET_CLIENT_LIST` + MIT-SCREEN-SAVER idle + WM activation | `sysctl` `kinfo_proc` + `CGWindowListCopyWindowInfo` + `NSWorkspace` / `NSRunningApplication.activateWithOptions:` | `internal/platform` tests across Win / Linux / macOS CI |
| **Audio cues** | Synthesized WAV PCM via `winmm` `PlaySoundW` (`sound_windows.go`); fires on the alert plan **and** on the six jump/copy/settings confirmation points (`beep(beepExclamation)` / `beep(beepFinished)`) | **No playback path at all**: `sound.go` synthesises the WAV, nothing consumes it, and `internal/ui/app_linux.go` never calls `beep`. `soundVolume` / `soundScheme` are read by `internal/config` and then ignored | Synthesized WAV played async via AppKit `NSSound` (`sound_darwin.go`), but only from the **alert plan** (`app_darwin.go`, one call site): finished/error/waiting transition tones and the UI confirmation beeps do not exist here | `TestToneForTransition` on all 3 OSes (pure synthesis); `TestPlaySoundLive` Win32 live probe; coverage per platform measured in §"Config keys" below |
| **Clipboard copy** | Win32 `OpenClipboard` / `SetClipboardData` | X11 `CLIPBOARD` `SelectionRequest` owner | `NSPasteboard` general pasteboard (`clipboard_darwin.go`) | Win32 + Linux X11 + macOS unit bindings |
| **Autostart** | `HKCU\...\Run` registry key (`--autostart=on/off`) | XDG `~/.config/autostart/ai-terminal-pet.desktop` | LaunchAgent `~/Library/LaunchAgents/io.github.jjjjjjjjnnjnn.ai-terminal-pet.plist` | `internal/autostart` unit tests on Windows + Ubuntu + macOS CI |
| **Installer & release archives** | `pet-setup.exe`, `ai-terminal-pet-<ver>-win-x64.zip`, `scripts/install.cmd` (`install.ps1`) | `ai-terminal-pet-<ver>-linux-{amd64,arm64}.tar.gz`, `scripts/install.sh` | `ai-terminal-pet-<ver>-darwin-{arm64,amd64}.tar.gz`, `scripts/install.sh` | Packaged & sha256-verified on Windows + WSL2 Ubuntu + `ubuntu-latest` + `macos-latest` CI |

## Config keys and where they actually take effect

`config.json` decodes every key on all three platforms (`internal/config`), and
`internal/app` carries the values into `Options` everywhere. What differs is
whether anything **reads** those values: a key that is accepted and then ignored
is the one shape of gap that reads as "my setting does nothing". This table comes
from the code (which file mentions each field, per build tag), and its last rows
are the backlog for round 62 - it is a record, not a promise.

| Key | Windows | Linux | macOS |
|---|---|---|---|
| `notifications` | balloon layer (`app_windows.go`) | plan only, then dropped: nothing renders a balloon on X11 | plan only, then dropped |
| `sound` | gates the tone (`attention.go`) | gates a tone that nothing plays | gates a tone `NSSound` can play |
| `soundVolume` / `soundScheme` | applied to the synthesizer | stored, never read (only `config.go` mentions them) | applied (`app_darwin.go`) |
| `soundOnWaiting` / `soundOnError` | per-event gate | same gate, no playback; the checks are in `attention.go`, which is untagged, so the *decision* is identical everywhere and only the playback is missing | same gate, no playback |
| `notifyOnFinish` / `donePolicy` | decides the finish tone | plan only | plan only |
| `quietCurrent` / `quietSubagents` | policy (`attention.go`) | same policy (it suppresses marks too, so it works) | same policy |
| `foregroundAware` / `idleMinutes` | presence sampling | same | same |
| `alwaysOnTop`, `edgeRetract`, `opacity`, `theme`, `perSourceColor`, `language`, `hiddenAtLaunch`, `mutedSessions` | full | full (`app_linux.go`) | full (`app_darwin.go`) |
| `autoApprove` / `autoApproveTimeoutSec` | full | full | full |
| `settingsTextSize` | scales the settings window | stored only (there is no settings window) | stored only |
| `settingsKeyboard` / `settingsTooltips` | settings window | stored only | stored only |

Two rows are the round-62 backlog, written down rather than only known locally:
`notifications` (no balloon surface on Linux or macOS) and the whole `sound*`
block on Linux.

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
   - Linux audio has no playback path at all (`internal/ui` never calls `beep` there; the synthesiser in `sound.go` builds a WAV nothing consumes). Nothing to install, nothing to break - and nothing to hear.
   - Live interactive WindowServer GUI on physical Apple hardware (and non-WSLg desktop environments such as GNOME/KDE/Wayland+XWayland) remains opt-in verified (`PET_COCOA_PROBE=1` on macOS, WSLg X11 live-tested on Linux).
