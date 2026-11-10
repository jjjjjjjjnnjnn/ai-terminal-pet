# dsh-ai-terminal-pet

A DeepSeek Harness (DSH) Cordis bundle that mirrors Harness session state onto
the [AI Terminal Pet](https://github.com/jjjjjjjjnnjnn/ai-terminal-pet) desktop
overlay: a small always-on-top Win32 pet that shows which sessions are working,
which one is waiting for your approval, which failed, and which finished.

The package ships the overlay binary and the official reporter protocol module
inside the tarball, so a single market install is enough to get a pet on screen.
No npm dependencies, no install scripts, no network calls at runtime (unless the
maintainer pins a release in lib/release-pins.json, see below).

## What it does

| Harness event | Overlay state |
| --- | --- |
| turn/start, agent/inbox/claimed, agent/status running | working |
| tool/call | tool_use |
| approval/asked, approval/request | waiting_permission |
| tool/result with error, command/done error, agent/error | error |
| turn/end, command/done success | idle |
| Harness process exit | closed |

The overlay then plays an alert tone, shows a tray balloon on those transitions
and its one-click jump button targets the Harness window. Plugin reload (HMR)
disposes the listeners but deliberately does not close sessions: only a real
process exit means the conversation is over.

The bundle's entry in the [awesome-dsh-plugin](https://github.com/awesome-dsh-plugin/awesome-dsh-plugin)
registry ([PR #13937](https://github.com/awesome-dsh-plugin/awesome-dsh-plugin/pull/13937)) was
merged on 2026-11-10, and that repository's README lists it. Being merged is not the same as
being visible in a market, and initially it was not: the entry page answered 404 and the catalog
file behind DSH's own Plugins -> Market did not contain this repository while the upstream site
build was still queued. That build finished (run 42844452013, 16:39:27Z), the entry page now
answers 200, and the catalog (`count` 10119 -> 10230) lists this repository - so it can be found
in that market. The entry's `install` command, however, is generated from the pinned `tarball`
and pointed at the 1.2.0 asset; a follow-up PR moves it to the current release, which is 1.2.3.
Changing that file always means a pull request against that repository, never an edit here.

The other route needs no application at all: third-party markets harvest GitHub
repositories carrying the `dsh-plugin` topic (this repository has it) together with
a root `package.json` that declares `name` and `dsh.bundle`. That is exactly why
the repository root carries a shim manifest — `dsh plugin add github:owner/repo`
reads the root manifest, and without one the install resolves to a 0.0.0
dependency that mounts nothing.

The package describes itself to those storefronts through `dsh.catalog`: a
category, a one-line summary in both languages, and the capabilities it uses.
Capabilities are self-declared and never checked — DSH does not sandbox plugins —
so read them as a description, not a permission grant.
The bundle is Windows only and declares `dsh.bundle.platforms: ["win32"]`, so
other platforms show it as unsupported instead of installing it.

It also declares `dsh.marketplace` (`profiles: ["desktop", "headless", "web"]`,
`requiresBuildApproval: false`, `manualSteps: false`). DSH itself never reads
that field - the bundle mounts through `dsh.bundle` either way - but storefronts
that sort entries by profile do: without it a registry falls back to a
headless-or-web guess and a desktop client hides the entry.

Manually, from a terminal, into a profile you own — use the release tarball or
pin the release tag. Both are verified end to end:

```sh
# release tarball: no clone, no Go toolchain
dsh plugin --profile <profile> add https://github.com/jjjjjjjjnnjnn/ai-terminal-pet/releases/download/dsh-plugin-v1.2.3/dsh-ai-terminal-pet-1.2.3.tgz

# git source, pinned to the release tag
dsh plugin --profile <profile> add "git+https://github.com/jjjjjjjjnnjnn/ai-terminal-pet.git#dsh-plugin-v1.2.3&path:/packages/dsh-ai-terminal-pet"
```

Do not hand pnpm a bare repository URL. pnpm resolves a git dependency by
running the package's lifecycle scripts, and it blocks those for unpinned
git sources until you copy the `allowBuilds` key it prints into your profile's
`pnpm-workspace.yaml`; the failure message never mentions AI Terminal Pet. The
package therefore ships no lifecycle hook, and its vendored overlay is a
committed file rather than a build product. `vendor/manifest.json` records the
SHA-256 of that file, the SHA-256 of the Go sources it was built from, and the
Go version that produced it, so a stale or hand-edited vendor directory fails a
check instead of shipping quietly.

The GitHub paths above need no npm. Installing by package name
(`add dsh-ai-terminal-pet`) starts working once the package is published to the
npm registry; until then use one of the two commands above.

`add` is the whole install: the profile records the dependency, and the bundle
declares itself through `dsh.bundle.patch`, so no separate `enable` step exists.
The profile you add it to has to be the one Harness boots.

Two profiles are special. The reserved `desktop` profile is rejected by the
standalone `dsh` CLI (`profile "desktop" is managed exclusively by the Electron
application`); the Desktop app's own bundled CLI does accept it, and the in-app
plugin manager is the supported path there. Any other profile name is yours.

If the app reports restart-required after enabling the bundle, restart Harness.

## What it writes on disk

First use in a Harness process copies two files, and only when they are missing:

| File | Origin | Purpose |
| --- | --- | --- |
| %LOCALAPPDATA%\AI-Terminal-Pet\protocol.js | vendor/protocol.js | pipe name, state set, redaction, snapshot |
| %LOCALAPPDATA%\AI-Terminal-Pet\AI-Terminal-Pet.exe | vendor/AI-Terminal-Pet.exe | the Win32 overlay |

Nothing is ever overwritten: an existing file, including one placed by the
standalone installer, wins. Set DSH_PET_NO_BOOTSTRAP=1 to skip this step
entirely (for example when the overlay is managed by `scripts/install.ps1` already).

Reports also write the overlay's normal runtime snapshot under the same
directory, exactly like every other reporter.

### Optional download fallback

lib/release-pins.json is absent from the published package by default. When the
maintainer fills it in with pinned release asset URLs, sizes and SHA-256
digests, a machine that has no vendored binary downloads the two assets over
HTTPS with a 30 s timeout, verifies size and digest with node:crypto, and only
then installs them. Any failure is silent apart from one warning; the plugin
keeps working without an overlay.

## The overlay itself is still standalone

This bundle is a convenience. The overlay is a Go application that installs
normally with `scripts/install.ps1` from the repository, and every other reporter
(Claude Code, OpenCode, Antigravity, generic) talks to the same named pipe. A
machine that only wants the overlay does not need this package; a machine that
only wants the plugin gets one installed automatically.

## Requirements and limits

- Windows only (named pipes plus a Win32 layered window).
- Node 22.19+ or 24+ in the Harness runtime.
- Nothing in this package can fail the host: every entry point is wrapped,
  every report is fire-and-forget, and there is no process.exit.
- Session id is the Harness session.header.id; pid is a synthetic per-session
  id and parentPid is the real Harness process, which is how the overlay ties a
  session to its terminal window.
