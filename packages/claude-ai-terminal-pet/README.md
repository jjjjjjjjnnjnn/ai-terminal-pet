# AI Terminal Pet — Claude Code plugin

Shows every Claude Code session in the AI Terminal Pet panel: which one is
working, which one is blocked on you and for how long, and lets the panel answer
a permission gate before Claude Code's own 5-second countdown runs out. The
panel is the same 56 px always-on-top strip the standalone app shows for
DeepSeek Harness, OpenCode and Antigravity.

**Windows only in this release.** The hooks run the native `pet-hook.exe` and the
SessionStart step installs the bundled panel into
`%LOCALAPPDATA%\AI-Terminal-Pet\`. On macOS and Linux use
`ai-terminal-pet --connect=claude` instead; the README explains what that wires.

## Install

```sh
claude plugin marketplace add jjjjjjjjnnjnn/ai-terminal-pet
claude plugin install ai-terminal-pet@ai-terminal-pet-market
```

The first command adds this repository as a marketplace; the second installs the
hooks. Plugin installation copies the package into Claude Code's cache, so the
two commands work offline afterwards.

If you already wired Claude Code with `ai-terminal-pet --connect=claude`, remove
those `hooks` entries from `~/.claude/settings.json` first: two reporters would
send the same session twice.

## What it changes

- Eight Claude Code hooks (`SessionStart`, `UserPromptSubmit`, `PreToolUse`,
  `PostToolUse`, `PermissionRequest`, `Notification`, `Stop`, `SessionEnd`) run
  `vendor/pet-hook.exe --mode=claude`, the native reporter.
- The first `SessionStart` runs `bootstrap.cjs`, which copies the bundled panel
  into `%LOCALAPPDATA%\AI-Terminal-Pet\` **once**: it verifies the copy against
  `vendor/manifest.json`, writes through a temporary file and renames, and never
  overwrites an existing panel. Node is needed for that one step; the per-event
  path is the native binary.
- Nothing is sent anywhere. The reporter writes to the local named pipe and the
  runtime snapshots the standalone app already uses.

## Uninstall

```sh
claude plugin uninstall ai-terminal-pet
claude plugin marketplace remove ai-terminal-pet-market
```

Claude Code removes the hook wiring. It does **not** delete its plugin cache
(`~/.claude/plugins/cache/...`) and this plugin does not delete the panel it
installed in `%LOCALAPPDATA%\AI-Terminal-Pet\`: an overlay you installed yourself
must not be removed by a plugin. Delete both by hand if you want them gone.

## Layout

| Path | What |
|---|---|
| `.claude-plugin/plugin.json` | plugin manifest, versioned with the overlay |
| `hooks/hooks.json` | the eight events, all native |
| `bootstrap.cjs` | one-time panel install (digest verified, never overwrites) |
| `vendor/pet-hook.exe` | native reporter, built from `cmd/pet-hook` |
| `vendor/AI-Terminal-Pet.exe` | the panel this plugin installs |
| `vendor/manifest.json` | overlay + source digests, written by `scripts/vendor.mjs` |

Regenerate `vendor/` after a Go change with
`node packages/claude-ai-terminal-pet/scripts/vendor.mjs`; `--check` is the CI
gate. The version tracks the overlay (`internal/app/version.go`); the next
overlay release bumps this manifest too, so a plugin number always names the
pair that was tested together.
