# Contributing to ai-terminal-pet

Thanks for looking. This project has one rule that shapes everything else: **it must never make the developer's terminal slower or less reliable.** The overlay is a convenience; your `claude`, `opencode` or `antigravity` session is the actual work.

## The short version

- **Bug reports**: open an issue with the `--doctor` output (see below). Most reports resolve in one look when it is attached.
- **Reporter for another terminal**: welcome, and you do not need to read the source. [`docs/PROTOCOL.md`](docs/PROTOCOL.md) is the complete wire contract.
- **Code changes**: run the gates in the next section. CI runs exactly the same ones, so a green local run means a green CI run.
- **Wording and market-copy changes**: also a code change here. Everything that ships in the plugin package is release-immutable, so a changed sentence costs a version number.

## Before you report "the pet does not appear"

Run this and paste the whole output:

```powershell
& "$env:LOCALAPPDATA\AI-Terminal-Pet\AI-Terminal-Pet.exe" --doctor
```

Ten checks answer the usual causes (missing install files, no runtime directory, corrupt config, nobody listening on the pipe, Node missing, no autostart entry, and the last alert decision). A failing line names the thing to fix; a green report means the overlay is healthy and the problem is on the terminal side — in that case say which terminal and which hook you configured.

Two environment traps that look like product bugs and are not:

- **A directory carrying a Low mandatory integrity label** makes every executable launched from it inherit that label, and a Low-integrity process cannot open the overlay's pipe (the overlay runs at Medium). Copy the binaries to `%TEMP%` or into `%LOCALAPPDATA%\AI-Terminal-Pet\` and try again.
- **`--json` is not a liveness probe.** It reads disk snapshots, so it prints `{"count":0}` whether or not the overlay is running. To check for a live instance, look for the pipe.

## Building and running the gates

```powershell
gofmt -l .        # must print nothing
go vet ./...
go test -count=1 ./...

# The two scanners need no C compiler. They declare a Go newer than this
# module's floor, which is why CI runs them in their own `scan` job:
go run golang.org/x/vuln/cmd/govulncheck@v1.9.0 ./...
go run github.com/securego/gosec/v2/cmd/gosec@v2.30.0 -severity high -confidence high -exclude G115,G703 ./...

# node --check validates only the FIRST file it is given, so check each one:
foreach ($f in Get-ChildItem reporters\*.js) { node --check $f.FullName }

node integrations/deepseek-harness/verify.mjs
node packages/dsh-ai-terminal-pet/scripts/vendor.mjs --check
node packages/dsh-ai-terminal-pet/verify.mjs
node tools/upstream-entry.mjs --check
node --test tools/upstream-entry.test.mjs
node tests/fresh-machine.mjs --no-overlay
.\scripts\install.cmd -VerifyOnly
```

Two things about that list are worth knowing before you trust a red run:

- **`-race` is not here.** On Windows it needs gcc/cgo, which this project does not require (see
  `docs/ROADMAP.md` §6). Do not call the scans above a "full security audit"; they are a static scan and a
  dependency-vulnerability scan.
- **`tests/fresh-machine.mjs` without `--no-overlay` only makes sense on an idle machine.** The GUI scenarios
  boot the overlay, and the single-instance mutex is `Local\AI-Terminal-Pet-v1-SingleInstance` - per Windows
  session, not per sandbox. If a host CLI with this project's reporter is running in the same session, its next
  hook event relaunches the real overlay, the sandbox copy loses the mutex and exits cleanly without creating
  its pipe, and S4 looks like a product failure. Kill every `AI-Terminal-Pet.exe` first if you want the 182-check
  form; CI deliberately runs the 162-check form (`docs/HANDOVER.md` §35.8 has the measurement).

`tests/fresh-machine.mjs` is the gate that answers "does this work on a machine that has never seen it" — it packs the plugin, extracts it into a DSH-style `node_modules`, installs through the installer and drives all three binaries inside a sandbox with its own `LOCALAPPDATA`, `USERPROFILE` and named pipe. Drop `--no-overlay` to also boot the overlay inside that sandbox.

The three builds use `-trimpath -buildvcs=false -ldflags "-s -w"` on purpose: without `-buildvcs=false` the binary embeds the commit and working-tree state, so two builds of identical source differ and `.\scripts\install.cmd -VerifyOnly` reports a false failure afterwards.

## Rules the code follows

These are load-bearing; a pull request that breaks one will be asked to change:

1. **Reporters fail fast and silently.** Nothing under `reporters/` may throw into the host CLI, block it, or make it wait. That includes error paths and parse failures.
2. **No external runtime dependencies.** The binaries must stay statically compiled against Windows system libraries only (`user32.dll`, `gdi32.dll`, `shell32.dll`). No GDI+, no UI framework, no Node for the overlay itself.
3. **The pipe name is a cross-language contract.** It is derived twice — Go (`internal/pipeserver`) and JS (`reporters/protocol.js`) — and pinned by `TestPipePathForUserMatchesReporterAlgorithm`. Changing one side without the other silently disconnects every reporter on every machine.
4. **UTF-8, LF line endings.** `.gitattributes` pins `* text=auto eol=lf`. `scripts/install.ps1` and `scripts/install.cmd` must stay pure ASCII, because Windows PowerShell 5.1 reads a BOM-less script in the ANSI code page.
5. **Protocol changes follow [`docs/PROTOCOL.md`](docs/PROTOCOL.md) §11.** A schema change is a versioned decision, not a field rename.
6. **No machine-specific absolute paths** in source, tests, scripts or docs. Use `%LOCALAPPDATA%`, `os.UserHomeDir()`, `Path(__file__).resolve().parent`, or an environment variable that fails with a human-readable message.

The full architecture notes, release procedure and the reason each of these exists are in [`docs/HANDOVER.md`](docs/HANDOVER.md) and [`AGENTS.md`](AGENTS.md). The catalogue side — the `awesome-dsh-plugin` rules, the release-to-listing checklist, and why the entry is rebuilt from that repository's `main` instead of merged — is in [`docs/UPSTREAM.md`](docs/UPSTREAM.md).

## Release-immutable files

Published npm tarballs and GitHub Release assets cannot be edited. These files change only by shipping a new version:

- `packages/dsh-ai-terminal-pet/README.md` and its two `package.json` files
- `packages/dsh-ai-terminal-pet/vendor/` (both files and `manifest.json`)
- `docs/market-index.json` → `plugins[0].versions[0]`
- any tag under `dsh-plugin-v*`

## Commit messages

Conventional-commit prefixes (`feat:`, `fix:`, `docs:`, `test:`, `refactor:`, `chore:`), a subject in the imperative mood, and a body that explains *why*. If a change fixes something a user would notice, say what a user would have seen before it. Numbers in a commit message should be something you measured, not something you remember.

## License

By contributing you agree that your contribution is licensed under the MIT License, the same terms as [`LICENSE`](LICENSE).
