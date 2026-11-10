# AI Terminal Pet bridge for DeepSeek Harness

A Host-only Cordis bundle that reports DeepSeek Harness session state to the
AI Terminal Pet desktop overlay.

## What it does

| Harness event | Overlay state |
| --- | --- |
| `turn/start`, `agent/inbox/claimed`, `agent/status` running | `working` |
| `tool/call` | `tool_use` |
| `approval/asked`, `approval/request` | `waiting_permission` |
| `tool/result` with error, `command/done` error, `agent/error` | `error` |
| `turn/end`, `command/done` success | `idle` |
| Harness process exit | `closed` |

The overlay then plays an alert tone and shows a tray balloon on those
transitions, and its one-click jump button targets the Harness window.

## Protocol source

This bundle contains **no** pipe-name, state-set or snapshot logic of its own.
It loads the official protocol module from
`%LOCALAPPDATA%\AI-Terminal-Pet\protocol.js`, which `install.ps1` copies from
`reporters/protocol.js` and verifies by SHA256 on every install. All reports
are fire-and-forget; a missing overlay never affects Harness.

Session id = Harness `session.header.id`; `pid` is a synthetic
`stableProcessId`, `parentPid` is the real Harness process id. Sessions are
closed explicitly on exit so the overlay does not accumulate ghosts.

## Install into DeepSeek Harness

The `desktop` profile is managed exclusively by the Electron application (the
`dsh` CLI refuses it), so install through the in-app plugin manager. In a
Harness conversation, ask:

```
Use plugin_manager with action install_bundle and target
"%LOCALAPPDATA%\AI-Terminal-Pet\bundles\ai-terminal-pet-bridge"
```

That string is pasted into the in-app manager, so the environment variable form is intended:
%LOCALAPPDATA% resolves per user, and a hard-coded C:\Users\<name>\... would be wrong on
every other machine. If the manager does not expand variables, substitute your own profile path.

That directory is a copy of this package, refreshed by `install.ps1`.
Installing links the local directory into the profile and applies
`cordis.patch.yml`; restart Harness if the manager reports `restart-required`.
Do not hand-edit the profile's `package.json` or `cordis.patch.yml`: the
desktop profile runs pnpm on startup and would discard manual edits.

To verify, start a Harness task and run:

```
"%LOCALAPPDATA%\AI-Terminal-Pet\AI-Terminal-Pet.exe" --json
```

A `deepseek-harness` session should appear and change state as the task runs.
`plugin_manager` with `remove_bundle` removes it.

## Files

- `index.js` — Host plugin (`apply(ctx, config)` + cleanup disposer)
- `cordis.patch.yml` — loader `insert:` row
- `locale/{en,zh}.json` — display metadata
- `icon.svg` — plugin-manager icon
