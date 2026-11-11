# ai-terminal-pet — AI Maintainer & Agent Guide (AGENTS.md)

> **Project Identity**: Lightweight, native Windows floating companion overlay for AI coding terminals
> (OpenCode, Claude Code, Antigravity CLI, DeepSeek Harness). Aggregates multi-agent session states via
> Windows named pipes into a glanceable Win32 layered overlay.
> **Language**: Go 1.23+ / Win32 API / Node.js (Reporters, zero npm dependencies)
> **License**: MIT

---

## 🎯 Architectural Blueprint & Core Invariants

```
Host terminal (Claude Code / OpenCode / Antigravity / DeepSeek Harness)
       │
       ▼ (fire-and-forget hooks)
reporters/*.js  ──[named pipe / JSON-lines]──►  cmd/ai-terminal-pet (Go)
                                                    │
                                                    ├─► internal/pipeserver (32 conns, 5 s read timeout)
                                                    ├─► internal/store (max 256 sessions, TTL + tombstones)
                                                    ├─► internal/store (disk snapshots, %LOCALAPPDATA%)
                                                    └─► internal/ui (Win32 layered window, topmost)
```

### Key Subsystems

- `cmd/ai-terminal-pet/`: primary entry point. CLI dispatch (`--show`, `--toggle`, `--reset`, `--settings`,
  `--quit`, `--json [--pretty]`, `--autostart[=on|off]`, `--doctor`, `--version`, `--help`), error-log
  rotation, bootstrap.
- `cmd/pet-statusline/`: high-performance native statusline shim (avoids the Node.js startup penalty
  under Windows AV hooks).
- `cmd/pet-hook/`: native hook reporter, `--mode=claude|antigravity`. Reads one hook payload on stdin
  and reports it without spawning Node. The JavaScript hooks in `reporters/` remain the reference
  implementation and the fallback.
- `internal/reporter/`: the reporter core shared by `cmd/pet-hook` and `cmd/pet-statusline`:
  redaction, bounded ids, stable PID (FNV-1a over UTF-16), timestamp clamps, stdin limits, snapshot
  write, pipe delivery, approval handshake, and both hook-event→state and statusline→state mappings.
  Anything that must match `reporters/protocol.js` belongs here so it can be unit-tested instead of
  eyeballed. `cmd/pet-statusline` deliberately keeps single-shot delivery and a 5-minute ignore on a
  future `startedAt`; both are documented in its header.
- `internal/ui/`: Win32 layered overlay — GDI canvas rasterization (`CreateDIBSection` +
  `UpdateLayeredWindow`, **no GDI+**), panel layout, tray icon, click/drag/wheel gestures, three-tab
  settings window, synthesized alert tones, per-session alerts, auto-approve countdown, jump-to-terminal.
- `internal/pipeserver/`: named pipe server (`\\.\pipe\ai-terminal-pet-v1-<8hex>`) with `OK`/`ERR` replies
  and a per-user DACL.
- `internal/store/`: in-memory state store with tombstone cleanup plus snapshot-file fallback under
  `%LOCALAPPDATA%\AI-Terminal-Pet\runtime\`.
- `internal/platform/`: single-instance mutex, process/window presence probing, focus activation.
- `internal/config/`, `internal/i18n/`, `internal/session/`, `internal/autostart/`, and `internal/app/`
  (which also holds `doctor_windows.go`, the `--doctor` implementation).
- `reporters/`: host CLI hooks (`claude-hook.js`, `opencode.js`, `antigravity-hook.js`,
  `antigravity-statusline.js`, `generic.js`, shared `protocol.js`).
- `integrations/deepseek-harness/ai-terminal-pet-bridge/`: host-only Cordis bundle plus `verify.mjs`.

---

## ⚡ Standard Verification & Build Commands

Any AI modifying this repository must run the following steps. They are exactly what GitHub Actions runs
in [.github/workflows/ci.yml](.github/workflows/ci.yml), so a green local run means a green CI run.

```powershell
# 1. Static analysis and all unit tests (184 func Test across 10 internal packages + cmd/pet-hook)
gofmt -l .          # must print nothing
go vet ./...
go test -count=1 ./...

# 2. Reporter syntax and the DeepSeek Harness bundle (zero npm dependencies)
for f in reporters/*.js; do node --check "$f"; done   # per file: node --check only guards its first argument
node --check tests/fresh-machine.mjs
node integrations/deepseek-harness/verify.mjs

# 3. Production build
go build -trimpath -buildvcs=false -ldflags "-s -w" -o .\bin\AI-Terminal-Pet.exe ./cmd/ai-terminal-pet
go build -trimpath -buildvcs=false -ldflags "-s -w" -o .\bin\pet-statusline.exe ./cmd/pet-statusline
go build -trimpath -buildvcs=false -ldflags "-s -w" -o .\bin\pet-hook.exe ./cmd/pet-hook

# 3b. Fresh-machine end-to-end suite (CI runs it with --no-overlay; locally drop
#     the flag to also boot the overlay in a sandbox and check delivery)
node tests/fresh-machine.mjs --no-overlay

# 4. Install, then let the shipped binary prove the install
.\scripts\install.ps1
.\scripts\install.ps1 -VerifyOnly     # per-file SHA256; exit 0 only when everything matches
& "$env:LOCALAPPDATA\AI-Terminal-Pet\AI-Terminal-Pet.exe" --doctor   # 9 checks; exit 1 on any failure
.\scripts\install.ps1 -AddToPath      # optional: adds the install dir to HKCU Path
.\scripts\install.ps1 -Uninstall      # needs `yes` at the prompt, or -Force

# 5. Optional UI Layout Preview (renders PNGs into PET_PREVIEW_DIR)
$env:PET_PREVIEW_DIR = "$env:TEMP\petpreview"
go test ./internal/ui -run TestRenderPreview
go test ./internal/ui -run TestRenderSettingsPreview
python docs/assets/render_assets.py   # composes hero.png and the panel PNGs from that directory
```

Opt-in desktop probes (their tests `Skip` when the variable is unset — that is why CI stays headless):
`PET_PREVIEW_DIR`, `PET_SOUND_PROBE`, `PET_FOCUS_PROBE`, `PET_DSH_PROBE`, `PET_PRESENCE_PROBE`.

Two conventions in those commands are load-bearing, not style:

- **`-buildvcs=false`** in all three builds. The VCS stamp records the commit and the working-tree state at
  build time — `go version -m` on an unstamped build shows `(devel)`, on a stamped one a pseudo-version like
  `v0.0.0-20261111122731-3529bffc12e1+dirty` — so two builds of the same source differ across machines, dates
  and working-tree states. That is enough to make `scripts/install.ps1 -VerifyOnly` report a false failure;
  the byte-count delta varies by toolchain (it was 1,024 B on go1.23 and is currently 0 B on go1.27, because
  the pseudo-version happens to be the same length as `(devel)`), which is exactly why the flag matters more
  than the number. The binaries are not committed (`.gitignore` ignores `bin/`), so "matches the
  repository" means "matches the copy in your working tree".
- **`node --check` one file at a time.** `node --check a.js b.js` checks only the first argument and
  exits 0 even when the second file has a syntax error (verified on Node 24).

---

## 📚 Where the documentation lives

| Question | File |
|---|---|
| How do I install, use, and wire up a terminal? | [README.md](README.md) (bilingual: English then 中文) |
| What changed in which version? | [CHANGELOG.md](CHANGELOG.md) |
| What is planned, and why? | [docs/ROADMAP.md](docs/ROADMAP.md) |
| How did we get here? (runbooks §13–§27) | [docs/HANDOVER.md](docs/HANDOVER.md) |
| I am picking this up from someone else. What is pending, what must not be edited, what looks redundant but is not? | [docs/HANDOVER.md](docs/HANDOVER.md) §28 (handover checklist) |
| I am writing my own reporter | [docs/PROTOCOL.md](docs/PROTOCOL.md) — the wire protocol contract |
| What does the DSH plugin package ship, and how is it versioned? | [packages/dsh-ai-terminal-pet/README.md](packages/dsh-ai-terminal-pet/README.md) |
| How does Discover find this plugin without npm? | [docs/market-index.json](docs/market-index.json) (format 1, pinned to the release commit) |
| Frozen historical documents | [docs/archive/README.md](docs/archive/README.md) |

---

## 🛑 Strict Rules for AI Agents

1. **Host Terminal Safety (Iron Law)**:
   - Hook scripts under `reporters/` MUST ALWAYS fail fast and silently. Never allow an unhandled
     exception or crash in `reporters/*.js` to block, delay, or kill the developer's host CLI
     (`claude`, `opencode`, `antigravity`, or the DeepSeek Harness).
2. **Single-Instance Mutex**:
   - The app relies on a named global mutex to enforce single-instance semantics. Do not bypass or
     weaken mutex acquisition in `internal/platform`.
   - `--doctor` deliberately does **not** open the mutex: `CreateMutex` would create the object and
     disturb a live instance, so it uses the pipe listener as its evidence instead (see
     `internal/app/doctor_windows.go`, `checkInstance`).
3. **No External Runtime Bloat**:
   - The Go binary must remain statically compiled with zero runtime DLL dependencies outside standard
     Windows system libraries (`user32.dll`, `gdi32.dll`, `shell32.dll`).
4. **Encoding & Line Endings**:
   - Preserve UTF-8 encoding across all files. `.gitattributes` pins `* text=auto eol=lf`; do not
     introduce CRLF. `scripts/install.ps1` must stay pure ASCII, because Windows PowerShell 5.1 reads a
     BOM-less script as ANSI.
5. **Cross-language protocol contract**:
   - The pipe name is the first 4 bytes of `sha256(home lowercased, trailing separators trimmed)`
     rendered as 8 hex chars. It is implemented twice — Go (`internal/pipeserver`) and JS
     (`reporters/protocol.js`) — and pinned by `TestPipePathForUserMatchesReporterAlgorithm`.
     Changing one side without the other breaks every reporter silently. Schema changes follow
     `docs/PROTOCOL.md` §11.
6. **Release commit must carry the market index**:
   - `docs/market-index.json` states the version each release should install, and
     `verify.mjs` asserts that its newest entry equals the packaged version. Bumping the
     version without moving the index in the same commit turns the tagged commit's CI red
     (that is exactly what happened to `b4474a6`: run `42894142140`, step
     `Verify the DSH plugin package`, `fail: the newest indexed version is the packaged version`).
     Move the index, the two manifests and the packaged README together, then tag.
   - Because git cannot embed a commit's own sha, a release takes two commits: **A** moves the
     version, the packaged README and the index's `version` / `publishTimestamp` (and is what the tag
     points at); **B** rewrites the index's `source.commit` and `icon` to A's real sha. Neither may be
     skipped: 1.2.1 → 1.2.3 all shipped this way, and 1.2.4 is the first one where the two halves were
     planned before the version bump rather than patched afterwards.
   - Any text-only release still costs a version number. Release assets are immutable, so editing one
     sentence in the packaged README or in `dsh.catalog` means a new tag (that is why 1.2.1 and 1.2.4
     exist); never re-upload or rewrite a published `.tgz`.
7. **Portability**:
   - Never commit machine-specific absolute paths into source, tests, scripts or docs. Use
     `os.Getenv("LOCALAPPDATA")`, `os.UserHomeDir()`, `Path(__file__).resolve().parent`, or an
     environment variable that fails with a human-readable message.
8. **Second opinion before a release, and every review request goes through `agy`**:
   - When a change needs an outside look — a release, a marketplace submission, a README that makes
     claims about other people's indexes — ask Antigravity's CLI, not a hand-written checklist:
     `agy --model gemini-3.8-flash-high --effort high --print='<prompt>'` (the prompt must be attached
     to `--print=`, never as a trailing argument; `agy models` lists what is installed).
   - This is a standing preference for **every project in this workspace**, not just this one: prefer
     `gemini-3.8-flash-high` when a second model's opinion would help, and say in the commit message
     which finding came from it. Treat its output as a reviewer, not as an authority — verify each
     claim against the installed packages before acting on it.
