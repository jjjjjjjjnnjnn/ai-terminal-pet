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
| **Audio cues** | Synthesized WAV PCM via `winmm` `PlaySoundW` (`sound_windows.go`) | Silent placeholder (`sound_linux.go`) | Synthesized WAV PCM played async via AppKit `NSSound` (`sound_darwin.go`) | `TestToneForTransition` on all 3 OSes; `TestPlaySoundLive` Win32 live probe |
| **Clipboard copy** | Win32 `OpenClipboard` / `SetClipboardData` | X11 `CLIPBOARD` `SelectionRequest` owner | `NSPasteboard` general pasteboard (`clipboard_darwin.go`) | Win32 + Linux X11 + macOS unit bindings |
| **Autostart** | `HKCU\...\Run` registry key (`--autostart=on/off`) | XDG `~/.config/autostart/ai-terminal-pet.desktop` | LaunchAgent `~/Library/LaunchAgents/io.github.jjjjjjjjnnjnn.ai-terminal-pet.plist` | `internal/autostart` unit tests on Windows + Ubuntu + macOS CI |
| **Installer & release archives** | `pet-setup.exe`, `ai-terminal-pet-<ver>-win-x64.zip`, `scripts/install.cmd` (`install.ps1`) | `ai-terminal-pet-<ver>-linux-{amd64,arm64}.tar.gz`, `scripts/install.sh` | `ai-terminal-pet-<ver>-darwin-{arm64,amd64}.tar.gz`, `scripts/install.sh` | Packaged & sha256-verified on Windows + WSL2 Ubuntu + `ubuntu-latest` + `macos-latest` CI |

## Data directories and IPC endpoints

Implemented identically in Go (`internal/home`, `internal/pipeserver`) and Node.js (`reporters/protocol.js`) and locked by cross-language tests:

- **Windows**: `%LOCALAPPDATA%\AI-Terminal-Pet` · named pipe `\\.\pipe\ai-terminal-pet-v1-<sha256(lower(home))[0..8]>`
- **macOS**: `~/Library/Application Support/AI-Terminal-Pet` · socket `$TMPDIR/ai-terminal-pet-v1-<8hex>.sock`
- **Linux**: `$XDG_DATA_HOME/ai-terminal-pet` (or `~/.local/share/ai-terminal-pet`) · socket `$XDG_RUNTIME_DIR/ai-terminal-pet/v1-<8hex>.sock` (or `/tmp/ai-terminal-pet-<uid>-v1-<8hex>.sock`)

## CI cross-build contract

1. **GitHub Actions runners**: `windows-latest` (`verify` + `scan`), `ubuntu-latest` (`linux` — builds & tests Linux binaries, headless install/hook/CLI smoke, `.tar.gz` SHA-256 verification), and `macos-latest` (`macos` — native `darwin/arm64` Go tests, binary builds, headless install/hook/CLI smoke).
2. **`scripts/cross-build.sh`** asserts that all packages in [scripts/portable-packages.txt](../scripts/portable-packages.txt) (every package in this repository — 16 packages) build cleanly with `CGO_ENABLED=0` for `windows/amd64`, `linux/amd64`, and `darwin/arm64`, while [scripts/windows-only-packages.txt](../scripts/windows-only-packages.txt) stays empty.
3. **Host-windowed gaps still open cleanly**:
   - On Linux & macOS, edit `config.json` directly instead of opening a native settings dialog (`--settings` logs/prints guidance cleanly).
   - Linux audio notifications stay silent (`internal/ui/sound_linux.go`) so zero ALSA/PulseAudio/cgo dependencies are required.
   - Live interactive WindowServer GUI on physical Apple hardware (and non-WSLg desktop environments such as GNOME/KDE/Wayland+XWayland) remains opt-in verified (`PET_COCOA_PROBE=1` on macOS, WSLg X11 live-tested on Linux).
