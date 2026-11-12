# dsh-ai-terminal-pet

A DeepSeek Harness (DSH) Cordis bundle that mirrors Harness session state onto
the [AI Terminal Pet](https://github.com/jjjjjjjjnnjnn/ai-terminal-pet) desktop
overlay: a small always-on-top Win32 pet that shows which sessions are working,
which one is waiting for your approval, which failed, and which finished.

The package ships the overlay binary and the official reporter protocol module
inside the tarball, so a single market install is enough to get a pet on screen
that follows **this Harness's** sessions. It does not attach to any other
terminal by itself: sessions from OpenCode, Claude Code and Antigravity appear
only after that terminal's own reporter is installed into it. The repository
README section [Connect your terminal](https://github.com/jjjjjjjjnnjnn/ai-terminal-pet#connect-your-terminal)
has one hook per terminal; the panel, the pipe and the protocol are shared, so
once a terminal reports it shows up next to the Harness sessions.

No npm dependencies, no install scripts, no network calls at runtime (unless the
maintainer pins a release in lib/release-pins.json, see below).

## What it does

| Harness signal | Overlay state |
| --- | --- |
| `turn/start`, `agent/inbox/claimed`, `agent/status` running | working |
| `tool/call` | tool_use |
| `approval/request` (waterfall listener) | waiting_permission |
| `tool/result` with `error`, `agent/error` | error |
| `turn/end`, `agent/status` idle | idle |
| Harness process exit | closed |

Events this table does not list are not part of the Harness session log: there is
no `command/done` or `working-directory/change` in `SessionEventMap`. The project
path, the parent session and the session start time do not come from events at
all - they are read from `session.header` (`cwd`, `parentSession`, `createdAt`),
which is the only place the host states them. That is also why a session shows
the right project even though the desktop Harness itself runs from its install
directory, and why a subagent conversation is nested under the session it was
forked from.

The overlay then plays an alert tone, shows a tray balloon on those transitions
and its one-click jump button targets the Harness window. Plugin reload (HMR)
disposes the listeners but deliberately does not close sessions: only a real
process exit means the conversation is over.

## Where it is listed, and what each market indexes

Three independent paths exist today, and each one reads a different field:

| Market | What it indexes | What this package offers it |
| --- | --- | --- |
| npm registry | package name, version, `dist.integrity` | published as `dsh-ai-terminal-pet`; a market that installs by name gets a verified tarball |
| dshmarket (the market tab most Harness Desktop users run) | `https://awesome-dsh-plugin.com/plugins.json`, which is generated from the [awesome-dsh-plugin](https://github.com/awesome-dsh-plugin/awesome-dsh-plugin) repository | entry merged via PR [#13937](https://github.com/awesome-dsh-plugin/awesome-dsh-plugin/pull/13937); the `install` command it renders comes from that entry's pinned `tarball`, so bumping it is always a pull request *there*, never an edit here |
| GitHub-harvesting markets | repositories carrying the `dsh-plugin` topic plus a root `package.json` declaring `name` and `dsh.bundle` | this repository has the topic and the root shim manifest, which is what makes `dsh plugin add github:owner/repo` resolve instead of mounting a 0.0.0 dependency |

"Listed" and "visible" are different states, and only the second one helps a
user: the awesome entry existed before its `plugins.json` entry did, because the
site builds the catalog on a schedule. Check the market's own public index, not
the merged pull request. The package README cannot promise what another
repository's build will render; it can only promise that the fields those builds
read are present and correct.

The package describes itself to storefronts through `dsh.catalog`: a category, a
one-line summary in both languages, and the capabilities it uses. Capabilities
are self-declared and never checked — DSH does not sandbox plugins — so read them
as a description, not a permission grant.
The bundle is Windows only and declares `dsh.bundle.platforms: ["win32"]`, so
other platforms show it as unsupported instead of installing it.

It also declares `dsh.marketplace` (`profiles: ["desktop", "headless", "web"]`,
`requiresBuildApproval: false`, `manualSteps: false`). DSH itself never reads
that field - the bundle mounts through `dsh.bundle` either way - but storefronts
that sort entries by profile do: without it a registry falls back to a
headless-or-web guess and a desktop client hides the entry.

Manually, from a terminal, into a profile you own — by package name, by release
tarball, or by pinning the release tag. All three are verified end to end:

```sh
# by package name: published on npm, no clone, no Go toolchain
dsh plugin --profile <profile> add dsh-ai-terminal-pet

# release tarball: same bytes as the npm release, straight from the GitHub Release
dsh plugin --profile <profile> add https://github.com/jjjjjjjjnnjnn/ai-terminal-pet/releases/download/dsh-plugin-v1.2.8/dsh-ai-terminal-pet-1.2.8.tgz

# git source, pinned to the release tag
dsh plugin --profile <profile> add "git+https://github.com/jjjjjjjjnnjnn/ai-terminal-pet.git#dsh-plugin-v1.2.8&path:/packages/dsh-ai-terminal-pet"
```

Those three differ only in transport. The npm path is the shortest and the one a
market uses when it installs by name; the tarball path works when a market pins a
release asset instead of a registry entry; the git path is the fallback for a
machine that cannot reach npm at all. Every one of them lands the same
prebuilt overlay in `vendor/`.

Do not hand pnpm a bare repository URL. pnpm resolves a git dependency by
running the package's lifecycle scripts, and it blocks those for unpinned
git sources until you copy the `allowBuilds` key it prints into your profile's
`pnpm-workspace.yaml`; the failure message never mentions AI Terminal Pet. The
package therefore ships no lifecycle hook, and its vendored overlay is a
committed file rather than a build product. `vendor/manifest.json` records the
SHA-256 of that file, the SHA-256 of the Go sources it was built from, and the
Go version that produced it, so a stale or hand-edited vendor directory fails a
check instead of shipping quietly.

The GitHub and npm paths above need nothing from this repository at install time.
`dsh plugin add dsh-ai-terminal-pet` resolves through the npm registry; the
tarball and git forms resolve through GitHub. A marketplace that indexes npm
sees the package under its own name and can verify the published integrity
digest; a marketplace that indexes repositories sees the `dsh-plugin` topic and
the root shim manifest.

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
(Claude Code, OpenCode, Antigravity, generic) talks to the same named pipe.
Those reporters are separate installs, not part of this package: a terminal that
has none of them hooked reports nothing, in this install or in any other. A
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
