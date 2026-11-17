# opencode-ai-terminal-pet

Report OpenCode session state to the [AI Terminal Pet](https://github.com/jjjjjjjjnnjnn/ai-terminal-pet)
overlay: one always-on-top strip that shows which of your agent sessions is
waiting on you, with the project, the reason it is blocked and how long it has
been blocked — and one click back to that terminal.

This package is the reporter only. The panel is a separate, small desktop app
that you install once; this package feeds it.

## Requirements

| | |
|---|---|
| OpenCode | **V2** (validated against 2.3.0; the reporter uses V2's `context.event.subscribe` and V2 event names) |
| Node | 22 or newer (the one OpenCode already ships with) |
| The AI Terminal Pet overlay | **1.0.18 or newer**, installed separately — see [Install the overlay](#install-the-overlay). This package is versioned with the overlay: `1.0.18` of the reporter and `1.0.18` of the panel are the pair that was tested together |
| Platform | anything OpenCode runs on; the reporting path is Windows, Linux and macOS |

## Install

```sh
opencode plugin add opencode-ai-terminal-pet
```

That is the whole integration. It adds the package to `plugins` in your
`opencode.json(c)` and OpenCode loads it on the next start (or reload).

Or pin it in `opencode.jsonc` yourself:

```jsonc
{
  "$schema": "https://opencode.ai/config.json",
  "plugins": ["opencode-ai-terminal-pet"]
}
```

### Install the overlay

The package reports; it does not draw. Install the overlay once:

- **Windows** — `pet-setup.exe` (or the portable zip) from
  [Releases](https://github.com/jjjjjjjjnnjnn/ai-terminal-pet/releases/latest).
- **Linux / macOS** — the `linux-*` / `darwin-*` tarball from the same release,
  then `./install.sh`.

You do not have to start it by hand: the first report launches it if it is
missing, and it exits by itself when there is nothing to show.

## What it reports

One state per conversation: `idle`, `thinking`, `working`, `tool_use`,
`waiting_permission`, `error` and `closed`, plus the session's title, project
directory, model, tool name and timestamps. A permission request and a question
both arrive as `waiting_permission` with different reasons, which is what lets
the panel say *approval* or *question* rather than guessing.

It deliberately does **not** report:

- tool arguments, prompts, file contents or command lines — the tool *name*
  (`bash`, `edit`, `grep`) is the only part of a tool call that leaves the plugin;
- token-level events (`session.text.delta` and friends), because they fire
  hundreds of times per answer and would repaint the panel for nothing;
- anything to a network. There is no HTTP client in this package and no
  telemetry: the reporter writes to local files and to a local endpoint owned by
  the overlay. That is the project's one strong promise, and it is why the
  package has zero dependencies.

Reporting is fire-and-forget by contract: every handler is wrapped, every
promise is caught, and a failure to report is silent. A reporter that can delay
or break the host CLI is a worse bug than a missing row.

## Verify it is working

With an OpenCode session running:

```sh
ai-terminal-pet --json --pretty     # the sessions the overlay can see
ai-terminal-pet --doctor            # `reporters: opencode: ok` when it is wired
```

The second one is the interesting one: it reads your OpenCode config and tells
you whether the package is listed, installed and still loadable, instead of
leaving you with a panel that never lights up.

## Uninstall

```sh
opencode plugin remove opencode-ai-terminal-pet
```

The plugin keeps no state of its own. Snapshots written under the overlay's data
directory are pruned by the overlay when the session's process is gone; if you
want them gone immediately, quit OpenCode and empty
`%LOCALAPPDATA%\AI-Terminal-Pet\runtime\` (Linux: `$XDG_DATA_HOME/ai-terminal-pet/runtime/`,
macOS: `~/Library/Application Support/AI-Terminal-Pet/runtime/`).

## What is inside

| File | What it is |
|---|---|
| `server.js` | The OpenCode plugin. **Generated** from [`reporters/opencode.js`](https://github.com/jjjjjjjjnnjnn/ai-terminal-pet/blob/main/reporters/opencode.js) — the reference implementation — with one line changed: the shared core is required as `./protocol.cjs` |
| `protocol.cjs` | The shared core, byte for byte the same as `reporters/protocol.js`: payload normalisation, redaction, snapshot writing, endpoint delivery, acknowledgement |

Two files instead of one because an npm plugin package is `"type": "module"`,
while the shared core is CommonJS — it is `require()`d by five hooks and is
deployed to the install directory as `protocol.js` by the overlay's own
installer, so it cannot become ESM. A generator plus a checked-in copy keeps the
two from drifting; that drift is a bug this project has already paid for once.

The wire contract, including the endpoint's name derivation and the exact `OK` /
`ERR` acknowledgement, is [docs/PROTOCOL.md](https://github.com/jjjjjjjjnnjnn/ai-terminal-pet/blob/main/docs/PROTOCOL.md).
The package speaks `ai-terminal-pet.v1`.

## License

MIT.

---

## 中文速览

`opencode plugin add opencode-ai-terminal-pet` 一行接入。这个包只负责**上报**：
OpenCode 里哪个会话在等你、卡在什么原因上、等了多久，交给
[AI Terminal Pet](https://github.com/jjjjjjjjnnjnn/ai-terminal-pet) 那块置顶小面板显示。

面板本体要单独装一次（Windows 用 `pet-setup.exe`，Linux/macOS 用对应 tar.gz），
装上之后**不需要手动启动**：第一次上报会把它拉起来，没有会话时它自己退出。

不上报工具参数、命令内容或文件内容，只上报工具**名字**；不出网，无遥测，零依赖。
包内 `server.js` 由仓库里的 `reporters/opencode.js` 生成，`protocol.cjs` 与
`reporters/protocol.js` 逐字节相同——这样同一个上报端不会在 npm 上悄悄变成第二份实现。
