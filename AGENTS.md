# ai-terminal-pet — AI Maintainer & Agent Guide (AGENTS.md)

> **Project Identity**: Lightweight, native Windows floating companion overlay for AI coding terminals (OpenCode, Claude Code, Antigravity CLI). Aggregates multi-agent session states via Windows Named Pipes into a glanceable Win32 layered overlay.
> **Language**: Go 1.23+ / Win32 API / Node.js (Reporters)
> **License**: MIT

---

## 🎯 Architectural Blueprint & Core Invariants

```
Host Terminal (Claude Code / OpenCode / Antigravity)
       │
       ▼ (Fire-and-forget hooks)
reporters/*.js  ──[Named Pipe / JSON-lines]──►  cmd/ai-terminal-pet (Go)
                                                    │
                                                    ├─► internal/pipeserver (32 conns, 5s timeout)
                                                    ├─► internal/store (max 256 sessions, TTL)
                                                    └─► internal/ui (Win32 layered window, topmost)
```

### Key Subsystems:
- `cmd/ai-terminal-pet/`: Primary application entry point. Handles CLI arguments (`--show`, `--toggle`, `--reset`, `--json`), error logging rotation, and bootstrap.
- `cmd/pet-statusline/`: High-performance native statusline shim (avoids Node.js startup penalty under Windows AV hooks).
- `internal/ui/`: Win32 layered overlay, GDI+ canvas rasterization, panel layout, system tray icon, and click-drag-wheel gestures.
- `internal/pipeserver/`: Named pipe server (`\\.\pipe\ai-terminal-pet-v1-<hash>`) with ACK/ERR response.
- `internal/store/`: In-memory state store with tombstone cleanup and snapshot file fallback under `%LOCALAPPDATA%\AI-Terminal-Pet\runtime\`.
- `reporters/`: Host CLI hooks (`claude-hook.js`, `opencode.js`, `antigravity-hook.js`, `generic.js`).

---

## ⚡ Standard Verification & Build Commands

Any AI modifying this repository must run the following verification steps:

```powershell
# 1. Run all unit tests
go test -count=1 ./...

# 2. Static analysis
go vet ./...

# 3. Syntax check on all hook reporters (Standard JS, zero npm dependencies)
node --check reporters/*.js

# 4. Production build
go build -trimpath -ldflags "-s -w" -o .\bin\AI-Terminal-Pet.exe ./cmd/ai-terminal-pet
go build -trimpath -ldflags "-s -w" -o .\bin\pet-statusline.exe ./cmd/pet-statusline

# 5. Optional UI Layout Preview (generates PNGs in Temp)
$env:PET_PREVIEW_DIR = "$env:TEMP\petpreview"
go test ./internal/ui -run TestRenderPreview
```

---

## 🛑 Strict Rules for AI Agents

1. **Host Terminal Safety (Iron Law)**:
   - Hook scripts under `reporters/` MUST ALWAYS fail fast and silently. Never allow an unhandled exception or crash in `reporters/*.js` to block, delay, or kill the developer's host CLI (`claude`, `opencode`, or `antigravity`).
2. **Single-Instance Mutex**:
   - The app relies on a named global mutex to enforce single-instance semantics. Do not bypass or weaken mutex acquisition in `internal/platform`.
3. **No External Runtime Bloat**:
   - The Go binary must remain statically compiled with zero runtime DLL dependencies outside standard Windows system libraries (`user32.dll`, `gdi32.dll`, `shell32.dll`).
4. **Encoding & Line Endings**:
   - Preserve UTF-8 encoding across all files.
