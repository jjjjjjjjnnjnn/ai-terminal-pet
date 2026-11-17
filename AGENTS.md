# ai-terminal-pet — AI Maintainer & Agent Guide (AGENTS.md)

> **Project Identity**: Lightweight, native floating companion overlay for AI coding terminals
> (OpenCode, Claude Code, Antigravity CLI, DeepSeek Harness). Aggregates multi-agent session states via
> one per-user endpoint (a Windows named pipe, a unix socket elsewhere) into a glanceable layered overlay.
> **Language**: Go 1.23+ / Win32 API (the window) / Node.js (Reporters, zero npm dependencies)
> **License**: MIT
>
> **The reporting path and the window layer are cross-platform; the depth of the verification is not.**
> What works where, and how each claim was measured, is [docs/PLATFORMS.md](docs/PLATFORMS.md) — read it before assuming a package
> builds everywhere, and before promising anything to a macOS or Linux user.

---

## 🎯 Architectural Blueprint & Core Invariants

```
Host terminal (Claude Code / OpenCode / Antigravity / DeepSeek Harness)
       │
       ▼ (fire-and-forget hooks)
reporters/*.js  ──[pipe / unix socket, JSON lines]──►  cmd/ai-terminal-pet (Go)
                                                          │
                                                          ├─► internal/pipeserver (32 conns, 5 s read timeout)
                                                          ├─► internal/store (max 256 sessions, TTL + tombstones)
                                                          ├─► internal/store (disk snapshots, internal/home)
                                                          └─► internal/ui (layered window, topmost: Win32 today)
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
  `UpdateLayeredWindow`, **no GDI+**), panel layout, tray icon, click/drag/wheel gestures, a four-tab
  settings window, synthesized alert tones, auto-approve countdown, jump-to-terminal, and the alert layer:
  `attention.go` is the pure policy (`Decide` / `PlanAlerts`, no window, file or clock), `attention_keys.go`
  holds its types and constants (untagged, so both can be checked off Windows), `attention_windows.go`
  samples the desktop off the UI thread, `mute.go` and `receipt_windows.go` hold the per-session
  marks and the decision record. `internal/session/attention.go` derives *what* a blocked session is asking
  for (approval / question / plan) from the reason reporters already send — the wire protocol did not move.
  **The buffer and all 16 drawing primitives live in `canvas.go` (untagged)**; `canvas_windows.go` is GDI
  behind them and `canvas_posix.go` is the compile-only placeholder (real buffer, no text). See
  [docs/PLATFORMS.md](docs/PLATFORMS.md) for what that means per OS.
- `cmd/pet-setup/`: the single-file installer (`pet-setup.exe`) the release ships. It embeds the
  staged payload, verifies it against `SHA256SUMS.txt` **before** writing anything, then runs the
  `scripts/install.ps1` that ships inside it — so `install.ps1` stays the only implementation of
  what an install is. `payload/` holds only `.gitkeep`; the real files are staged by
  `scripts/package.ps1` and deleted again before it returns.
- `internal/home/`: **the only answer to "where does this project keep its files"** (`%LOCALAPPDATA%\AI-Terminal-Pet`
  on Windows — unchanged — `~/Library/Application Support` on macOS, `$XDG_DATA_HOME/ai-terminal-pet` on Linux).
  Twenty-odd files used to spell that path themselves and `reporters/protocol.js` spelled it again in JavaScript.
- `internal/pipeserver/`: the transport-neutral protocol (`server.go`: one JSON line, 5 s deadline, exact
  `OK`/`ERR`, 32 connections) plus one platform file per transport — a named pipe with a per-user DACL on
  Windows, a `$XDG_RUNTIME_DIR` unix socket elsewhere. **The per-user token is a cross-language contract and
  must not move**; only the container changes.
- `internal/store/`: in-memory state store with tombstone cleanup plus snapshot-file fallback under
  `<data directory>/runtime/`.
- `internal/platform/`: single-instance mutex, process/window presence probing, focus activation
  (Windows-only: the POSIX presence rule is deliberately narrower — see `internal/app/platform_posix.go`).
- `internal/config/`, `internal/i18n/`, `internal/session/`, `internal/autostart/` (one backend per OS:
  HKCU Run / LaunchAgent / XDG desktop entry), `internal/connect/` (the host-configuration reader behind
  `--doctor`'s `reporters` line and `--connect`), and `internal/app/` (`run.go` is the lifecycle, `app.go`
  the neutral half, `doctor_windows.go` the `--doctor` implementation).
- `reporters/`: host CLI hooks (`claude-hook.js`, `opencode.js`, `antigravity-hook.js`,
  `antigravity-statusline.js`, `generic.js`, shared `protocol.js`). **This is the only implementation**; every
  packaged copy is generated from it.
- `packages/opencode-ai-terminal-pet/`: the OpenCode reporter as a publishable npm plugin. `server.js` and
  `protocol.cjs` are **generated** by `scripts/sync.mjs` from `reporters/opencode.js` and `reporters/protocol.js`
  (the one difference is the `require()` specifier: an npm plugin must be `"type": "module"`, the shared core
  is CommonJS). `--check` plus its tests are a CI gate: a fix that lands in `reporters/` and never reaches
  this package is the drift handover §52.3 paid for. The package is **versioned with the overlay**
  (`tools/check_versions.mjs` asserts the equality), so an overlay release publishes it in the same step.
- `integrations/deepseek-harness/ai-terminal-pet-bridge/`: host-only Cordis bundle plus `verify.mjs`.

---

## ⚡ Standard Verification & Build Commands

Any AI modifying this repository must run the following steps. They are exactly what GitHub Actions runs
in [.github/workflows/ci.yml](.github/workflows/ci.yml), so a green local run means a green CI run.

```powershell
# 1. Static analysis and all unit tests (499 func Test across 15 packages)
gofmt -l .          # must print nothing
go vet ./...
go test -count=1 ./...

# 1b. The two scanners that need no C compiler (so they run here). Both are pinned,
#     and both declare a Go newer than this module's floor, so CI runs them in a
#     separate `scan` job. `go test -race` needs cgo/gcc, which this machine
#     does not have: CI runs it in the `linux` job on ubuntu-latest (ROADMAP §6).
go run golang.org/x/vuln/cmd/govulncheck@v1.9.0 ./...      # "No vulnerabilities found"
go run github.com/securego/gosec/v2/cmd/gosec@v2.30.0 -severity high -confidence high -exclude G115 \
  -exclude-rules "(cmd/ai-terminal-pet/main|internal/app/app_windows|internal/ui/app_windows|internal/reporter/approval)\.go:G703" ./...

# 2. Reporter syntax and the DeepSeek Harness bundle (zero npm dependencies)
for f in reporters/*.js; do node --check "$f"; done   # per file: node --check only guards its first argument
node --check tests/fresh-machine.mjs
node --test reporters/opencode-events.test.mjs   # the OpenCode event -> state table
node integrations/deepseek-harness/verify.mjs
node packages/dsh-ai-terminal-pet/verify.mjs
node packages/opencode-ai-terminal-pet/scripts/sync.mjs --check   # the npm reporter is generated from reporters/
node --test packages/opencode-ai-terminal-pet/scripts/sync.test.mjs
node --test packages/opencode-ai-terminal-pet/server.test.mjs      # loads the package the way npm install does
node tools/upstream-entry.mjs --check   # catalogue entry points at this package version (SKIP is not a pass)
node --test tools/upstream-entry.test.mjs
node tools/check_versions.mjs           # every documented version equals the code's (badge, links, install commands)
node --test tools/check_versions.test.mjs
node tools/check_facts.mjs              # every documented count equals the tree's (func Test, packages, --doctor)
node --test tools/check_facts.test.mjs
python tools/check_docs_links.py   # every relative path and #anchor in the docs we maintain
python tools/check_assets.py       # every size/frame count in docs/assets/README.md matches the file
node --test tools/check-tarball.test.mjs   # the .tgz reader/writer rules (no tarball needed)
node tools/traffic-report.mjs       # reach / installs / npm, with the machine traffic separated (docs/METRICS.md)
node --test tools/traffic-report.test.mjs  # latest share and the asset baseline rules

# 2b. Doesn't run in CI (a clean checkout has no artifact) but must run before every
#     plugin release, after `npm pack` in the package directory: it unpacks the
#     tarball and compares it file for file with the working tree. 1.2.28 shipped a
#     packaged README that still said 1.2.27 because the pack ran before the bump.
node tools/check-tarball.mjs                                         # DSH plugin
node tools/check-tarball.mjs --package packages/opencode-ai-terminal-pet   # OpenCode reporter

# 3. Production build
go build -trimpath -buildvcs=false -ldflags "-s -w" -o .\bin\AI-Terminal-Pet.exe ./cmd/ai-terminal-pet
go build -trimpath -buildvcs=false -ldflags "-s -w" -o .\bin\pet-statusline.exe ./cmd/pet-statusline
go build -trimpath -buildvcs=false -ldflags "-s -w" -o .\bin\pet-hook.exe ./cmd/pet-hook

# 3a. The cross-platform gate. 16 packages must build for linux/amd64 and darwin/arm64,
#     and the two that must not are named so a port that silently regresses one is seen
#     (docs/PLATFORMS.md is the measured table this list has to agree with). Same logic
#     as CI's "Cross-platform build gate"; on Windows run it from Git Bash or WSL.
bash scripts/cross-build.sh

# 3b. Fresh-machine end-to-end suite (CI runs it with --no-overlay; locally drop
#     the flag to also boot the overlay in a sandbox and check delivery)
node tests/fresh-machine.mjs --no-overlay

# 3c. The portable release zip. Refuses when bin\ and vendor\ disagree on the
#     overlay; CI sets PET_PACKAGE_SKIP_VENDOR_CHECK=1 because its Go release differs.
.\scripts\package.ps1

# 4. Install, then let the shipped binary prove the install
.\scripts\install.cmd                 # cmd wrapper: RemoteSigned refuses a downloaded .ps1
.\scripts\install.cmd -VerifyOnly     # per-file SHA256; exit 0 only when everything matches
& "$env:LOCALAPPDATA\AI-Terminal-Pet\AI-Terminal-Pet.exe" --doctor   # 13 checks; exit 1 on any failure
.\scripts\install.cmd -AddToPath      # optional: adds the install dir to HKCU Path
.\scripts\install.cmd -Uninstall      # needs `yes` at the prompt, or -Force

# 5. Optional UI Layout Preview (renders PNGs into PET_PREVIEW_DIR)
$env:PET_PREVIEW_DIR = "$env:TEMP\petpreview"
go test ./internal/ui -run TestRenderPreview
go test ./internal/ui -run TestRenderSettingsPreview
python docs/assets/render_assets.py        # hero.png and the panel PNGs
python docs/assets/make_launch_assets.py   # demo.gif and desktop-scene.png
```

`scripts\install.ps1` stays the single implementation; `install.cmd` only starts one `powershell.exe`
with `-ExecutionPolicy Bypass` and forwards every argument and exit code, because a stock Windows install
runs RemoteSigned and that refuses a `.ps1` carrying the "downloaded from the internet" marker — which is
exactly what a user who unzips a release has.

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
| **A release is out — how does it reach the catalogue, and why does the entry keep conflicting?** | **[docs/UPSTREAM.md](docs/UPSTREAM.md)** — the upstream rules, the eight-step release-to-listing order, `tools/upstream-entry.mjs`, and what "listed" vs "visible" actually means |
| I am picking this up from someone else. What is pending, what must not be edited, what looks redundant but is not? | [docs/HANDOVER.md](docs/HANDOVER.md) §28 (handover checklist) |
| I am working on GitHub email, commit identity, signing or account settings | [docs/HANDOVER.md](docs/HANDOVER.md) §34 — the measured identity table, and the one irreversible decision (rewriting history) |
| I am writing my own reporter | [docs/PROTOCOL.md](docs/PROTOCOL.md) — the wire protocol contract |
| **What works on which OS, and how was each claim measured?** | **[docs/PLATFORMS.md](docs/PLATFORMS.md)** — the per-capability table (the reporting path is three platforms, the panel is one), what is unverified and why, and the two CI lists that pin it |
| I am announcing this somewhere (X, Hacker News, a marketplace) | [docs/PROMOTION.md](docs/PROMOTION.md) has the paste-ready copy; [docs/LAUNCH.md](docs/LAUNCH.md) has the measured numbers it must not exceed and the posting history |
| **How many people actually use this, and is that download count real?** | **[docs/METRICS.md](docs/METRICS.md)** — the 2026-11-17 measurement (33 views, 0 external installs), the crawler-fleet proof, and the bot-resistant metric; re-measure with `node tools/traffic-report.mjs` |
| **What is the growth plan, and why in that order?** | [docs/ROADMAP.md](docs/ROADMAP.md) §7 — the four decisions (positioning A+C, sequence, no untestable signing path, ≤2 releases a week), P1–P2 with owners, and the four-week targets |
| What does the DSH plugin package ship, and how is it versioned? | [packages/dsh-ai-terminal-pet/README.md](packages/dsh-ai-terminal-pet/README.md) |
| How does an OpenCode user install the reporter in one command? | [packages/opencode-ai-terminal-pet/README.md](packages/opencode-ai-terminal-pet/README.md) |
| How does Discover find this plugin without npm? | [docs/market-index.json](docs/market-index.json) (format 1, pinned to the release commit) |
| Frozen historical documents | [docs/archive/README.md](docs/archive/README.md) |

---

## 🛑 Strict Rules for AI Agents

1. **Host Terminal Safety (Iron Law)**:
   - Hook scripts under `reporters/` MUST ALWAYS fail fast and silently. Never allow an unhandled
     exception or crash in `reporters/*.js` to block, delay, or kill the developer's host CLI
     (`claude`, `opencode`, `antigravity`, or the DeepSeek Harness).
2. **Single-Instance**:
   - The app relies on one exclusive handle per logon session. Do not bypass or weaken its acquisition:
     a named mutex on Windows (`internal/app/platform_windows.go`), an `flock` on a file in the data
     directory on POSIX (`internal/app/platform_posix.go` — **not** "does the file exist", because only
     flock survives a hard kill).
   - `--doctor` deliberately does **not** open it: on Windows `CreateMutex` would create the object and
     disturb a live instance, so it uses the pipe listener as its evidence instead (see
     `internal/app/doctor_windows.go`, `checkInstance`).
3. **No External Runtime Bloat**:
   - The Go binary must remain statically compiled with zero runtime DLL dependencies outside standard
     Windows system libraries (`user32.dll`, `gdi32.dll`, `shell32.dll`).
   - On POSIX the equivalent is: **no cgo, no libgtk, no shelling out for anything the process does
     per frame.** A window layer that needs cgo is a decision the whole project has to make together
     (§57.7), not something a port slips in.
4. **Encoding & Line Endings**:
   - Preserve UTF-8 encoding across all files. `.gitattributes` pins `* text=auto eol=lf`; do not
     introduce CRLF. `scripts/install.ps1`, `scripts/install.cmd` and `scripts/package.ps1` must stay
     pure ASCII, because Windows PowerShell 5.1 reads a BOM-less script as ANSI.
5. **Cross-language protocol contract**:
   - The per-user endpoint's token is the first 4 bytes of `sha256(home lowercased, trailing separators
     trimmed)` rendered as 8 hex chars. It is implemented twice — Go (`internal/pipeserver/server.go`) and
     JS (`reporters/protocol.js`) — and pinned by `TestPipePathForUserMatchesReporterAlgorithm`.
     **The token is the contract; the container around it is not** (a Windows named pipe, a
     `$XDG_RUNTIME_DIR` socket elsewhere). Changing either side without the other breaks every reporter
     silently. Schema changes follow `docs/PROTOCOL.md` §11.
   - The data directory has the same shape of contract, one level up: `internal/home` answers it for Go,
     `reporters/protocol.js` mirrors it for JS, and `docs/PLATFORMS.md` records the per-OS answer
     (`%LOCALAPPDATA%\AI-Terminal-Pet` on Windows is **unchanged on purpose** — moving it would orphan
     every existing install).
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
   - **The npm publish step now covers two packages**: `dsh-ai-terminal-pet` (its own 1.2.x line) and
     `opencode-ai-terminal-pet` (**the overlay's own number** — `tools/check_versions.mjs` fails when the
     two disagree). Publishing the reporter is not optional in an overlay release: the package is generated
     from `reporters/`, so its bytes and the binary's bytes are the same code, and "reporter 1.0.18 +
     overlay 1.0.17" is a pair that was never tested together. Pack *after* the bump, then
     `node tools/check-tarball.mjs --package packages/opencode-ai-terminal-pet`.
7. **Portability**:
   - Never commit machine-specific absolute paths into source, tests, scripts or docs. Use
     `internal/home.Directory()` (or `internal/home.Join(...)`) for this project's own paths,
     `os.UserHomeDir()` for anything else, `Path(__file__).resolve().parent` in Python, or an environment
     variable that fails with a human-readable message.
   - **A test that needs the data directory calls its package's one helper** (e.g.
     `dataDirectory(t)` in `internal/reporter`, `cmd/pet-hook`) which points LOCALAPPDATA,
     XDG_DATA_HOME, HOME and USERPROFILE at one temp tree. Setting LOCALAPPDATA alone was correct while
     the path was spelled by hand and is wrong three platforms later: on linux it writes into the tester's
     real `~/.local/share` (§57.3). One helper per package, so a caller cannot pass the wrong root.
8. **Review the work yourself, with the model that wrote it**:
   - Do not shell out to `agy`, `gemini`, or any other model CLI for a second opinion. The whole
     chain — implementation, review, and the wording of anything outward-facing — runs in the session
     that is doing the work, on the model the user selected for it.
   - The substance of the old rule stays, because it was never about which model answered: read the
     claim back against the code before publishing it. A review that cannot point at the line
     implementing a sentence in the README is not a review. Where a claim depends on another
     repository's behaviour, read that repository's own files rather than recalling them.
   - **Every outward-facing text is verified before it is posted or published**: the launch copy in
     [docs/LAUNCH.md](docs/LAUNCH.md), a marketplace description, a README paragraph that makes a promise,
     an issue reply that explains a design decision. Drafting is cheap; a claim the project cannot
     support, published under its own name, is not. The `docs/LAUNCH.md` round found exactly that: a
     draft that read as if all four terminals worked with zero setup, when only DeepSeek Harness does.

9. **"It didn't write the file" starts with how the process exited**:
   - `config.json` is written from a goroutine on purpose (`configSaver`), because a synchronous write can
     stall the message pump on a machine with an AV write hook. A staged value reaches the disk on its own,
     and `defer settingsSaver.flush()` is only the last resort at a **graceful** exit.
   - So a `taskkill`, a `Stop-Process -Force` or a sandbox that gets torn down skips that flush. Verifying
     persistence by killing the process answers a different question than the one you asked - it proves
     nothing about the click handler. Close the window, `--quit`, wait a moment, then read the file.
   - This was recorded as an unexplained bug for a whole round (HANDOVER §44.7) and cost a re-verification
     in §45.1. `TestConfigSaverPersistsWithoutAFlush` / `TestConfigSaverFlushPersistsAStagedValue` pin both
     halves of the contract so the next reading starts from the mechanism.
10. **A usage number in public text is measured, and machine traffic is subtracted**:
   - `node tools/traffic-report.mjs` answers reach / installs / npm in one page, and
     [docs/METRICS.md](docs/METRICS.md) records what the numbers mean. **Never quote npm download totals**:
     on 2026-11-17 the package showed 4,586/week spread evenly over 25 versions - including 166 for the
     uninstallable `0.0.0-stage` - while the comparable `dsh-pet` put 58% of its downloads on its two newest
     versions. That flatness is a crawler fleet. The honest derived metric is `latest share` (humans ≈ 50%+,
     a fleet ≈ 3-10%), always with its window.
   - Release-asset downloads have a baseline too: each release downloads ~2 of its own assets back for
     verification, so anything near 2 per release is us. Subtract before calling it a user.
   - **Release cadence is a metric input**: every release adds an npm version, which the fleet fetches, which
     raises the total and lowers the latest share. Batch changes into one release (aim for ≤ 2 a week;
     security and regression fixes excepted) instead of shipping on every commit.
