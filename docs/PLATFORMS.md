# PLATFORMS — what runs where, measured

> Snapshot: **2026-11-15**, overlay `1.0.9`, plugin `1.2.18` (Track 3 + cross-platform Windows Win32 + Linux X11 + pure-Go macOS AppKit).
> Capabilities marked **local-verified** were executed on this machine (Windows 11 + WSL2 Ubuntu + WSLg `DISPLAY=:0`) plus `CGO_ENABLED=0` cross-builds (`linux/amd64`, `darwin/arm64`, `darwin/amd64`); macOS runtime execution runs in GitHub Actions `macos-latest` (`ci.yml` `macos` job, `PET_COCOA_PROBE=1` remains opt-in for interactive desktop windows).

## The short version

| Capability | Windows | Linux | macOS |
|---|---|---|---|
| `pet-hook` (native hook reporter) | ✅ | ✅ | ✅ (CI `macos-latest`) |
| `pet-statusline` (statusline shim) | ✅ | ✅ | ✅ (CI `macos-latest`) |
| `--json` | ✅ | ✅ | ✅ |
| `--doctor` (11 checks) | ✅ | ✅ | ✅ (`doctor_posix.go`) |
| `--connect` (prints, never writes) | ✅ | ✅ | ✅ |
| `--autostart` | ✅ HKCU Run | ✅ XDG desktop entry | ✅ LaunchAgent |
| named pipe / unix socket transport | ✅ | ✅ | ✅ |
| **the floating panel (the product)** | ✅ (Win32) | ✅ (pure-Go X11 ARGB) | ✅ (pure-Go Cocoa/AppKit via `purego`) |
| alert sound synthesis | ✅ (PCM `PlaySoundW`) | ❌ not yet | ✅ (shared PCM WAV + Cocoa `NSSound`) |
| clipboard copy | ✅ (Win32) | ✅ (X11 `CLIPBOARD`) | ✅ (Cocoa `NSPasteboard`) |
| jump-to-terminal focus & idle detection | ✅ | ✅ (X11 EWMH + `/proc`) | ✅ (CoreGraphics + `NSWorkspace` + `NSRunningApplication`) |
| tray menu / status item | ✅ (Win32 tray + balloon) | ❌ not yet | ✅ (`NSStatusBar` toggle / reset / quit) |
| settings window | ✅ | ❌ (edits config / tray pending) | ❌ (edits config; dedicated Cocoa settings UI pending) |

So: **all 16 Go packages compile cleanly with `CGO_ENABLED=0` across Windows, Linux, and macOS** with **zero cgo dependencies anywhere in the repository**. Shared code covers the pixel buffer & 16 drawing primitives (`canvas.go`), PCM WAV synthesis (`sound.go`), OpenType glyph rasterizer (`canvas_opentype.go` on `linux || darwin`), per-user Unix domain control sockets (`control_unix.go` on `linux || darwin`), and panel geometry/dock/alert/tap/keyboard state machines.

## Architecture of the window backends

- **Windows (`internal/ui/*_windows.go`, `sound_windows.go`, `platform/focus_windows.go`)**: Win32 layered window (`UpdateLayeredWindow` + GDI text/DIB, `settings_windows.go`, tray icon, PCM `PlaySoundW` audio).
- **Linux (`internal/ui/app_linux.go`, `canvas_opentype.go` + `canvas_fonts_linux.go`, `control_unix.go`, `clipboard_linux.go`)**: Pure-Go X11 (`github.com/jezek/xgb` + `randr` + `bigreq`, zero cgo / zero GTK), 32-bit TrueColor ARGB visual, `_NET_WM_WINDOW_OPACITY` pulse/opacity, chunked `PutImage` upload, OpenType TTF/OTF/TTC glyph rendering across `/usr/share/fonts` and `/mnt/c/Windows/Fonts`, per-user unix control socket, X11 `CLIPBOARD` selection owner, and EWMH jump-to-terminal (`_NET_CLIENT_LIST` + `/proc`).
- **macOS (`internal/ui/app_darwin.go`, `cocoa_darwin.go`, `darwin_geometry.go`, `canvas_fonts_darwin.go`, `sound_darwin.go`, `clipboard_darwin.go`, `platform/focus_darwin.go`)**: Pure-Go Objective-C runtime (`github.com/ebitengine/purego v0.9.1`, zero cgo) driving AppKit `NSApplication` + borderless transparent floating `NSWindow` (`AITerminalPetWindow`) + `NSImageView` (`NSBitmapImageRep` premultiplied RGBA upload), `NSScreen.visibleFrame` multi-display coordinate transforms, `NSStatusBar` menu bar item, `NSPasteboard` clipboard copy, shared WAV synthesis played via `NSSound`, OpenType TTF/OTF/TTC glyph rendering across `/System/Library/Fonts` and `~/Library/Fonts`, `CGWindowListCopyWindowInfo` / `CGEventSourceSecondsSinceLastEventType` / `NSRunningApplication` focus & idle queries, and shared Unix domain control sockets (`control_unix.go`).
- **Other POSIX (`*_posix.go` / `*_nogui.go` with `!windows && !linux && !darwin`)**: Explicit compile-safe non-window stubs returning `ErrNoWindowLayer`.

## What was measured, and how

```text
# Windows: full repository verification
gofmt -l . ; go vet ./... ; go test -count=1 ./...

# Linux (cross-compiled with CGO_ENABLED=0 on Windows, executed inside WSL2 Ubuntu)
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build ./cmd/... ./internal/...
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go vet ./...
# WSL2 Ubuntu (DISPLAY=:0 PET_X11_PROBE=1): Linux test suite + live X11 E2E smoke: OK

# macOS: zero-cgo arm64 + amd64 cross-vet/build & test compilation on Windows + CI job on macos-latest
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go vet ./...
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build ./cmd/... ./internal/...
GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go vet ./...
GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build ./cmd/... ./internal/...
# Native darwin/arm64 execution & headless CLI/socket/ObjC tests: GitHub Actions `.github/workflows/ci.yml` (`macos` job on `macos-latest`)
```

Per package across platforms (`CGO_ENABLED=0`):

| Package | windows | linux | darwin |
|---|---|---|---|
| `internal/{config,connect,home,i18n,pipeserver,reporter,session,store}` | ok | ok (tested in WSL2) | ok (CI `macos-latest` + local arm64/amd64 build+vet) |
| `internal/{autostart,platform,ui,app}` | ok | ok (tested in WSL2) | ok (CI `macos-latest` + local arm64/amd64 build+vet) |
| `cmd/{ai-terminal-pet,pet-hook,pet-setup,pet-statusline}` | ok | ok (tested on WSLg) | ok (CI `macos-latest` + local arm64/amd64 build+vet) |

## Remaining platform gaps

- **Live interactive desktop probing on a physical Mac display (`PET_COCOA_PROBE=1`)**: `ci.yml` `macos-latest` exercises `go test -count=1 ./...` (Objective-C class registration, AppKit loading, Unix control socket round-trip, pure Cocoa coordinate/pixel/keycode tests) and headless native `arm64` CLI/hook binaries; live physical monitor WindowServer visuals / notch offset / multi-Space gestures require an interactive macOS desktop session (`PET_COCOA_PROBE=1`).
- **Native Linux DEs outside WSLg**: the X11 panel, OpenType renderer, unix control socket, and E2E pipeline ran cleanly against WSLg XWayland (`DISPLAY=:0`), whereas native GNOME/KDE/XFCE compositor interactions, tray protocols (SNI/AppIndicator), and Linux desktop sound playback remain unverified.
- **Dedicated Cocoa/X11 settings GUI & POSIX installer**: `settings_windows.go` provides the native four-tab settings dialog on Windows; Linux & macOS edit `config.json` and use CLI/tray controls until native settings dialogs and `.tar.gz`/`.pkg` installers are added.
