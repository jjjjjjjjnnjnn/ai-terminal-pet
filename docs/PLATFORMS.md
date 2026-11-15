# PLATFORMS — what runs where, measured

> Snapshot: **2026-11-15**, overlay `1.0.9`, plugin `1.2.18`.
> Everything below was produced by running the command named next to it, on the
> machine that wrote this file (Windows 11 + WSL2 Ubuntu 26.04 + a darwin/arm64
> cross-build). Numbers that were not measured are marked **unverified**.

## The short version

| Capability | Windows | Linux | macOS |
|---|---|---|---|
| `pet-hook.exe` (native hook reporter) | ✅ | ✅ | ✅ |
| `pet-statusline.exe` (statusline shim) | ✅ | ✅ | ✅ |
| `--json` | ✅ | ✅ | ✅ |
| `--doctor` (11 checks) | ✅ | ⚠️ no window layer | ⚠️ no window layer |
| `--connect` (prints, never writes) | ✅ | ✅ | ✅ |
| `--autostart` | ✅ HKCU Run | ✅ XDG desktop entry | ✅ LaunchAgent |
| named pipe / unix socket transport | ✅ | ✅ | ✅ |
| **the floating panel (the product)** | ✅ | ❌ not yet | ❌ not yet |
| settings window, tray icon, jump-to-terminal | ✅ | ❌ not yet | ❌ not yet |

So: **the reporting path works on all three, the window exists on one.** That is
deliberate and it is stated out loud rather than shipped as a panel that silently
never appears — see `internal/app/run.go` (`ErrNoWindowLayer`) and the `reporters`
line of `--doctor`.

## Why the window is the hold-up

It is 6,047 lines of Win32 in `internal/ui` (`app_windows.go` 2,549 +
`settings_windows.go` 2,222 + `canvas_windows.go` 462 + the rest). The buffer and
all 16 drawing primitives are now platform-neutral (`internal/ui/canvas.go`), so a
new platform needs a backend that can:

1. put RGBA bytes on screen (GDI `UpdateLayeredWindow` today),
2. rasterise text (`DrawTextEx` today — the one primitive with no portable
   implementation),
3. deliver pointer/keyboard events.

`internal/ui/canvas_posix.go` is the honest placeholder: a real buffer, no text,
and a comment naming itself as the missing piece. It exists so the compiler can
check the 2,927 lines of drawing code on every platform, not to pretend a panel
works there.

## What was measured, and how

```
# Windows: everything
gofmt -l . ; go vet ./... ; go test -count=1 ./...
# 382 func Test, 14 packages, all green

# Linux (WSL2, go1.27.4): build + test, per package
GOOS=linux go build ./...      # fails only on ./internal/app and ./cmd/ai-terminal-pet
GOOS=linux go test ./...       # 12 of 14 packages run and pass

# darwin: build + vet only (no machine here, so no test run)
GOOS=darwin GOARCH=arm64 go build ./...   # same two packages fail
GOOS=darwin GOARCH=arm64 go vet  ./...    # ditto
```

Per package on linux and darwin, as of this snapshot:

| Package | windows | linux | darwin |
|---|---|---|---|
| `internal/{config,connect,home,i18n,pipeserver,reporter,session,store}` | ok | ok | ok |
| `internal/autostart` | ok | ok | ok |
| `internal/ui` | ok | ok | ok |
| `cmd/{pet-hook,pet-setup,pet-statusline}` | ok | ok | ok |
| `internal/platform`, `internal/app`, `cmd/ai-terminal-pet` | ok | **fails** | **fails** |

The three that fail are exactly the three that need the window layer; their
failure is `build constraints exclude all Go files` or `undefined: App`, not a
logic error.

## The Linux tests that run today

12 packages, and the interesting ones are not the trivially pure ones:

- `internal/reporter` — the retry policy against this platform's errors (a socket
  nobody listens on is ENOENT/ECONNREFUSED and is retried; a socket owned by
  another user is EACCES and returns immediately), the snapshot-before-delivery
  ordering, and a payload that must not leave a file behind.
- `internal/pipeserver` — read a line with a deadline, `OK`/`ERR`, a payload the
  server refuses, and the per-user token both implementations agree on.
- `internal/connect` — `--connect` and the `reporters` check with POSIX path
  spelling, including "a backslash is a character here, not a separator".
- `internal/autostart` — a real desktop-entry round trip: Enable, Status, Sync
  repairing a moved install, Disable idempotent, Sync never enabling.
- `internal/ui` — the alert policy, the ledger keys, the clock, the drawing
  primitives. No window.

## Unverified, and why

- **macOS behaviour**: no Mac here. The code compiles and vets for
  `darwin/arm64`; whether a LaunchAgent actually starts the overlay at login, and
  whether `--doctor` reads the right places, is **unverified**.
- **Linux with a real desktop**: the tests run headless in WSL2. X11 rendering,
  GTK/portal tray icons, and font availability are **unverified** — the WSL
  image has DejaVu only (no CJK font), which is why a future X11 backend must not
  promise CJK text before it measures it.
- **`--doctor` on POSIX**: the checks that read the install directory work, but the
  `reporters` line and the window-layer warning are the only parts that have been
  exercised, and the install-directory layout there follows `internal/home` rather
  than the Windows installer's.
