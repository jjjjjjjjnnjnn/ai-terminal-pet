# ai-terminal-pet

[![platform](https://img.shields.io/badge/platform-Windows-blue)](https://github.com/jjjjjjjjnnjnn/ai-terminal-pet)
[![license](https://img.shields.io/badge/license-MIT-green)](LICENSE)

> 🇬🇧 [English](#english) &nbsp;·&nbsp; 🇨🇳 [中文](#中文)

---

## English

A tiny always-on-top floating companion for AI coding terminals on Windows.
It aggregates session states (idle / thinking / working / tool_use /
waiting_permission / error / closed) from OpenCode, Claude Code and
Antigravity CLI into one glanceable overlay.

![platform](https://img.shields.io/badge/platform-Windows-blue)
![license](https://img.shields.io/badge/license-MIT-green)

### How it works

```
AI terminals ── reporters/*.js ── named pipe ── AI-Terminal-Pet.exe ── layered overlay
                                  (JSON line,   (Go, single instance)   (Win32, topmost)
                                   ACK = OK)
```

- **Transport**: per-user Windows named pipe
  `\\.\pipe\ai-terminal-pet-v1-<8hex>`, one JSON object per line
  (`schemaVersion: ai-terminal-pet.v1`, 1 MiB cap), server replies
  exact `OK` / `ERR`. Snapshot files under
  `%LOCALAPPDATA%\AI-Terminal-Pet\runtime\` act as a fallback channel.
- **App** (`cmd/ai-terminal-pet`): single instance (mutex), pipe server
  (32 conns, 5 s read timeout), snapshot scanner (2 s), Win32 layered
  overlay (compact `56x156`, expanded `200x<content>`, tray icon).
- **Reporters** (`reporters/`): `opencode.js` (plugin),
  `claude-hook.js`, `antigravity-hook.js`, `antigravity-statusline.js`,
  plus `generic.js` — a one-line bridge for any other CLI —
  sharing `protocol.js` (validation, redaction, retry-until-deadline,
  fail-fast, never throws into the host CLI).
- **pet-statusline** (`cmd/pet-statusline`): dependency-free native
  replacement for the Node statusline reporter. On machines where AV
  hooks make every `node`/shell spawn cost seconds, the native shim
  answers in well under a second.

### The overlay

**Compact** answers three questions in the order a glance asks them:

```
   ◉      state orb - colour plus glyph, pulses while something waits on you
   3      how many sessions are behind it
 会话
 ────
等待确认   what that state is, in words
ai-term…  which project is the one waiting
```

**Expanded** is the list itself. It grows with the number of sessions
(up to six rows, then "还有 N 个会话"), and each row shows the project, its
state, where it came from and how long ago it last spoke:

```
终端会话                            [等待确认]  ✕
──────────────────────────────────────────────
3 个会话 · 2 运行 · 1 等待
▍ai-terminal-pet          等待确认   2 分钟前
  claude-hook · Bash(go test)
▍webapp                   执行中     刚刚
  opencode · editing src/app.tsx
──────────────────────────────────────────────
      拖拽移动 · 滚轮缩放 · 点击收起
```

Interaction is deliberately uniform — there is nothing to aim at:

| gesture | effect |
|---|---|
| click anywhere | compact ↔ expanded |
| drag anywhere | move, 1:1 with the cursor |
| wheel | zoom 65 %–160 % |
| hover | the panel lights up, so it reads as a control |
| `Esc` | collapse |
| right-click | menu (hide, reset, notifications, sound, topmost, autostart, quit) |

The overlay is single-instance and removes itself once no terminal session
is left, so it is normally invisible in a process list.

### Control the overlay

Every command below is safe to bind to a hotkey or shell alias:

```powershell
ai-terminal-pet --show      # reveal the running overlay
ai-terminal-pet --toggle    # switch between compact and expanded
ai-terminal-pet --reset     # re-anchor to the screen edge at 100% zoom
ai-terminal-pet --quit      # stop the overlay
ai-terminal-pet --version
```

Launching the executable with no arguments while an instance is already
running reveals that instance instead of exiting silently.

### Read the state from a script

`--json` prints the current sessions and exits. It reads the reporter
snapshots on disk rather than talking to the overlay, so it works whether or
not the overlay is running and can never block on a GUI process:

```powershell
ai-terminal-pet --json            # one line, for jq / status lines
ai-terminal-pet --json --pretty   # indented, for humans
```

```json
{"schemaVersion":"ai-terminal-pet.v1","state":"waiting_permission","count":2,
 "generatedAt":"2026-09-26T05:31:37Z","sessions":[{"source":"opencode", ...}]}
```

`state` is the highest-priority state across all sessions; `sessions` is the
same list, already sorted by urgency. A reporter that stopped writing is
absent, not stale.

### Tray menu

Right-click the tray icon:

- **显示浮窗 / 隐藏到托盘**
- **重置位置与大小** — escape hatch for a pet dragged somewhere awkward
- **✓ 状态提醒** — tray balloons when a session starts waiting for
  permission or fails
- **✓ 提示音** — a system alert tone on the same transitions
- **✓ 窗口置顶**
- **✓ 贴边收起** — when docked to a screen edge, the panel slides most of the
  way off, leaving a thin sliver; hover the sliver to bring it back. A session
  waiting for you (or one that errored) always stays fully visible.
- **✓ 开机自启** — run at logon, via `HKCU\...\CurrentVersion\Run`
- **退出**

### Settings and remembered state

Two files under `%LOCALAPPDATA%\AI-Terminal-Pet\`:

- `config.json` — yours. Read at startup, rewritten only when you flip a
  tray switch, so editing it by hand while the app is stopped is safe.
  Missing or malformed values fall back to defaults; a broken file never
  stops the overlay from starting.

  ```json
  { "notifications": true, "sound": true, "alwaysOnTop": true, "opacity": 235 }
  ```

  `sound` is separate from `notifications` because the two reach you
  differently: a balloon is something you read when you look, a tone is
  something you hear when you are not looking.

- `state.json` — the app's. Window position, zoom and expanded flag, so the
  pet comes back where you left it. Delete it (or run `--reset`) to get the
  default right-edge anchor back.

Autostart is not in either file — the registry is its source of truth, so it
can also be driven from a script:

```powershell
ai-terminal-pet --autostart        # report the current state
ai-terminal-pet --autostart=on     # run at logon
ai-terminal-pet --autostart=off    # stop running at logon
```

### Build from source (Windows, Go 1.23+)

```powershell
go vet ./...
go test -count=1 ./...
go build -trimpath -ldflags "-s -w" -o .\bin\AI-Terminal-Pet.exe ./cmd/ai-terminal-pet
go build -trimpath -ldflags "-s -w" -o .\bin\pet-statusline.exe ./cmd/pet-statusline
.\install.ps1   # stops the running pet, copies both binaries, SHA256-verifies
```

`install.ps1` retries the copy: the reporters relaunch the overlay on demand,
so it can be alive again milliseconds after being stopped, and a plain
copy-over fails at random on a machine that is actually using the tool.

`node --check reporters/*.js` covers the reporters (stdlib only).

#### Seeing a layout change without a window

Layered-window geometry is easy to get subtly wrong and hard to look at
because there is nothing to screenshot. The rasterizer can be driven straight
to PNG instead:

```powershell
$env:PET_PREVIEW_DIR = "$env:TEMP\petpreview"
go test ./internal/ui -run TestRenderPreview
```

That writes the compact and expanded panel in several states at 2–3x scale.
It is a development aid, not an assertion, so it skips unless
`PET_PREVIEW_DIR` is set.

### Wire up a terminal

- **OpenCode**: add the file plugin to `plugin` in `opencode.json`:
  `file:///C:/path/to/ai-terminal-pet/reporters/opencode.js`
- **Claude Code**: add `node <repo>/reporters/claude-hook.js` commands for
  `SessionStart`, `SessionEnd`, `UserPromptSubmit`, `PreToolUse`,
  `PostToolUse`, `Stop`, `Notification` in `settings.json`.
- **Antigravity**: point `statusLine.command` at the installed
  `pet-statusline.exe` (or the `.js` reporter if your Node is fast),
  and register `antigravity-hook.js` for agent/tool events.
- **Anything else**: one line, no plugin needed. Any tool that can run a
  shell command on state changes can report:

  ```powershell
  node reporters/generic.js --source=aider --state=working --project="$PWD"
  node reporters/generic.js --source=codex --state=waiting_permission --reason=approve
  echo '{"state":"error","reason":"build failed"}' | node reporters/generic.js --source=ci --stdin
  ```

  States: `idle thinking working tool_use waiting_permission error closed`.
  `--help` lists every option.

Reporters are fire-and-forget: any failure returns silently and never
breaks the host CLI. Restart the terminal after changing its config so
the new reporter loads.

### Layout

```
cmd/ai-terminal-pet    main exe entry (CLI control flags + error.log rotation)
cmd/pet-statusline     native statusline shim
internal/app           wiring: mutex, pipe server, scanner, window, state saver
internal/config        config.json + state.json (atomic writes, fail-safe reads)
internal/pipeserver    named-pipe server (user-isolated name, ACK/ERR)
internal/store         session store (256 cap, tombstones, TTLs, snapshots)
internal/session       schema validation + priority/aggregation/sort
internal/platform      single-instance mutex, process-liveness probe
internal/ui            layered overlay, canvas rasterizer, panel layout, tray
reporters/             JS push side (incl. generic.js)
install.ps1            stop → copy (retrying) → SHA256 verify (both binaries)
```

### Known limitations

- Windows only.
- Per-launch AV scanning can make Node-based reporters slow; use the
  native shim and/or add an AV exclusion for the install directory.
- Same-user pipe preemption is mitigated (mutex-first ordering), not
  eliminated — a future `FILE_FLAG_FIRST_PIPE_INSTANCE` or
  challenge-response upgrade is tracked in HANDOVER.md.
- `go test -race` needs a gcc toolchain. The UI has unit tests for pure
  logic (hit testing, panel sizing, release decisions, text fitting) and a
  PNG preview renderer for layout, but no automated pixel regression.

See [HANDOVER.md](HANDOVER.md) for the operational runbook and
[ROADMAP.md](ROADMAP.md) for the extension and lightweighting plans.

### License

MIT — see [LICENSE](LICENSE). Third-party modules used as libraries only:
`github.com/Microsoft/go-winio` (MIT), `github.com/lxn/win` (BSD-style),
`golang.org/x/sys` (BSD-style). No third-party code is vendored.

---

## 中文

一个常驻置顶的迷你浮窗，专门陪伴 Windows 上的 AI 编程终端。它把
OpenCode、Claude Code、Antigravity CLI 的会话状态（idle / thinking /
working / tool_use / waiting_permission / error / closed）聚合成一块
一眼可辨的悬浮面板。

![platform](https://img.shields.io/badge/platform-Windows-blue)
![license](https://img.shields.io/badge/license-MIT-green)

### 工作原理

```
AI 终端 ── reporters/*.js ── 命名管道 ── AI-Terminal-Pet.exe ── 分层浮层
                               (每行一个 JSON,  (Go, 单实例)        (Win32, 置顶)
                                回执 = OK)
```

- **传输**：按用户隔离的 Windows 命名管道
  `\\.\pipe\ai-terminal-pet-v1-<8hex>`，每行一个 JSON 对象
  （`schemaVersion: ai-terminal-pet.v1`，上限 1 MiB），服务端回精确
  `OK` / `ERR`。`%LOCALAPPDATA%\AI-Terminal-Pet\runtime\` 下的快照文件
  作为兜底通道。
- **主程序**（`cmd/ai-terminal-pet`）：单实例（互斥体）、管道服务端
  （32 连接，5 秒读超时）、快照扫描（2 秒）、Win32 分层浮层
  （紧凑态 `56×156`，展开态 `200×<内容>`，含托盘图标）。
- **上报端**（`reporters/`）：`opencode.js`（插件）、`claude-hook.js`、
  `antigravity-hook.js`、`antigravity-statusline.js`，以及 `generic.js`——
  给其他任意 CLI 用的「一行桥接」——它们共用 `protocol.js`
  （校验、脱敏、重试到截止、快速失败、绝不把错误抛回宿主 CLI）。
- **pet-statusline**（`cmd/pet-statusline`）：无依赖的原生状态栏上报端，
  替代 Node 版本。在那些被杀毒软件钩子拖慢、每次起 `node`/shell 都要
  好几秒的机器上，原生 shim 能在不到一秒内响应。

### 浮层

**紧凑态**按「视线扫过」的顺序回答三个问题：

```
   ◉      状态球——颜色加字形；有人在等你时脉动
   3      背后有几路会话
 会话
 ────
等待确认   这个状态用文字怎么说
ai-term…  正在等的那个项目是哪个
```

**展开态**就是会话列表本身。它随会话数量增长（最多六行，之后显示
「还有 N 个会话」），每行展示项目、状态、来源，以及它上次说话是多久前：

```
终端会话                            [等待确认]  ✕
──────────────────────────────────────────────
3 个会话 · 2 运行 · 1 等待
▍ai-terminal-pet          等待确认   2 分钟前
  claude-hook · Bash(go test)
▍webapp                   执行中     刚刚
  opencode · editing src/app.tsx
──────────────────────────────────────────────
      拖拽移动 · 滚轮缩放 · 点击收起
```

交互刻意保持「整块可操作」——没有需要瞄准的小按钮：

| 手势 | 效果 |
|---|---|
| 点击任意处 | 紧凑 ↔ 展开 |
| 拖拽任意处 | 移动，与光标 1:1 跟随 |
| 滚轮 | 缩放 65%–160% |
| 悬停 | 面板提亮，看起来像可操作控件 |
| `Esc` | 收起 |
| 右键 | 菜单（隐藏、重置、提醒、提示音、置顶、开机自启、退出） |

浮层是单实例的，当没有任何终端会话时它会自行移除，所以在进程列表里
通常看不到它。

### 控制浮层

下面每条命令都可以安全地绑到快捷键或 shell 别名：

```powershell
ai-terminal-pet --show      # 显示正在运行的浮层
ai-terminal-pet --toggle    # 在紧凑与展开之间切换
ai-terminal-pet --reset     # 以 100% 缩放重新贴回屏幕边缘
ai-terminal-pet --quit      # 停止浮层
ai-terminal-pet --version
```

如果已有实例在运行，不带参数启动可执行文件会显示那个实例，而不是
静默退出。

### 从脚本读取状态

`--json` 会打印当前会话并退出。它读取磁盘上的上报端快照，而不是去问
浮层，所以无论浮层是否运行都能工作，也绝不会阻塞在某个 GUI 进程上：

```powershell
ai-terminal-pet --json            # 单行，给 jq / 状态栏用
ai-terminal-pet --json --pretty   # 缩进美化，给人看
```

```json
{"schemaVersion":"ai-terminal-pet.v1","state":"waiting_permission","count":2,
 "generatedAt":"2026-09-26T05:31:37Z","sessions":[{"source":"opencode", ...}]}
```

`state` 是所有会话里优先级最高的那个状态；`sessions` 是同一份列表，
已按紧急程度排好序。一个停止写入的上报端会「消失」，而不是「变旧」。

### 托盘菜单

右键托盘图标：

- **显示浮窗 / 隐藏到托盘**
- **重置位置与大小** — 当浮窗被拖到别扭位置时的一键逃生口
- **✓ 状态提醒** — 有会话开始等你批准或失败时弹出托盘气泡
- **✓ 提示音** — 同样的状态切换时发出系统提示音
- **✓ 窗口置顶**
- **✓ 贴边收起** — 当浮窗贴在屏幕边缘时，会滑出大半、只留一条细边；
  把光标悬停在细边上即可唤回。正在等你（或出错）的会话永远保持完整可见。
- **✓ 开机自启** — 通过 `HKCU\...\CurrentVersion\Run` 在登录时运行
- **退出**

### 设置与记忆的状态

`%LOCALAPPDATA%\AI-Terminal-Pet\` 下有两个文件：

- `config.json` — 属于你。启动时读取，只在你切换托盘开关时才被重写，
  所以应用停止时手动编辑它是安全的。缺失或格式错误的值会回退到默认；
  一个坏掉的文件绝不会让浮层起不来。

  ```json
  { "notifications": true, "sound": true, "alwaysOnTop": true, "opacity": 235 }
  ```

  `sound` 与 `notifications` 分开，是因为两者触达你的方式不同：气泡是
  「你去看时读到的」，提示音是「你没看时听到的」。

- `state.json` — 属于应用的。记录窗口位置、缩放与展开标记，让浮窗下次
  回到你离开时的样子。删掉它（或运行 `--reset`）即可回到默认的右边缘
  锚点。

开机自启不在这两个文件里——注册表才是它的唯一真源，所以也能用脚本驱动：

```powershell
ai-terminal-pet --autostart        # 报告当前状态
ai-terminal-pet --autostart=on     # 登录时运行
ai-terminal-pet --autostart=off    # 登录时不运行
```

### 从源码构建（Windows, Go 1.23+）

```powershell
go vet ./...
go test -count=1 ./...
go build -trimpath -ldflags "-s -w" -o .\bin\AI-Terminal-Pet.exe ./cmd/ai-terminal-pet
go build -trimpath -ldflags "-s -w" -o .\bin\pet-statusline.exe ./cmd/pet-statusline
.\install.ps1   # 停止运行中的浮窗，拷贝两个二进制，做 SHA256 校验
```

`install.ps1` 会重试拷贝：上报端会按需重新拉起浮窗，所以它可能在被
停止后几毫秒就又活过来了，在真正在用这个工具的机器上，普通的覆盖拷贝
会随机失败。

`node --check reporters/*.js` 覆盖全部上报端（仅用标准库）。

#### 没有窗口也能看版式变化

分层窗口的几何尺寸很容易被改得微妙出错，又因「没法截图」而难以查看。
可以让光栅化器直接输出 PNG：

```powershell
$env:PET_PREVIEW_DIR = "$env:TEMP\petpreview"
go test ./internal/ui -run TestRenderPreview
```

这会把紧凑态和展开态在若干状态下以 2–3 倍缩放写入 PNG。它是个开发辅助
而非断言，所以除非设置了 `PET_PREVIEW_DIR`，否则会跳过。

### 接入终端

- **OpenCode**：在 `opencode.json` 的 `plugin` 里加上文件插件：
  `file:///C:/path/to/ai-terminal-pet/reporters/opencode.js`
- **Claude Code**：在 `settings.json` 里为 `SessionStart`、`SessionEnd`、
  `UserPromptSubmit`、`PreToolUse`、`PostToolUse`、`Stop`、`Notification`
  添加 `node <repo>/reporters/claude-hook.js` 命令。
- **Antigravity**：把 `statusLine.command` 指向安装好的 `pet-statusline.exe`
  （如果你的 Node 够快，也可以用 `.js` 上报端），并注册 `antigravity-hook.js`
  处理 agent/工具事件。
- **其他任意工具**：一行搞定，无需插件。任何能在状态变化时跑 shell
  命令的工具都能上报：

  ```powershell
  node reporters/generic.js --source=aider --state=working --project="$PWD"
  node reporters/generic.js --source=codex --state=waiting_permission --reason=approve
  echo '{"state":"error","reason":"build failed"}' | node reporters/generic.js --source=ci --stdin
  ```

  状态：`idle thinking working tool_use waiting_permission error closed`。
  `--help` 列出所有选项。

上报端是 fire-and-forget（发了就走）：任何失败都静默返回，绝不会搞崩
宿主 CLI。改完终端配置后请重启终端，让新的上报端加载进来。

### 目录结构

```
cmd/ai-terminal-pet    主程序入口（CLI 控制参数 + error.log 轮转）
cmd/pet-statusline     原生状态栏 shim
internal/app           装配：互斥体、管道服务端、扫描器、窗口、状态保存
internal/config        config.json + state.json（原子写入，容错读取）
internal/pipeserver    命名管道服务端（用户隔离命名，ACK/ERR）
internal/store         会话存储（上限 256，墓碑、TTL、快照）
internal/session        schema 校验 + 优先级/聚合/排序
internal/platform       单实例互斥体、进程存活性探测
internal/ui            分层浮层、画布光栅化、面板版式、托盘
reporters/             JS 上报端（含 generic.js）
install.ps1            停止 → 拷贝（重试）→ SHA256 校验（两个二进制）
```

### 已知限制

- 仅支持 Windows。
- 每次启动的杀毒扫描会让 Node 上报端变慢；请使用原生 shim，和/或给
  安装目录加杀毒排除项。
- 同用户管道抢占已缓解（互斥体优先排序），但未被根除——未来的
  `FILE_FLAG_FIRST_PIPE_INSTANCE` 或挑战-应答升级记录在 HANDOVER.md。
- `go test -race` 需要 gcc 工具链。UI 对纯逻辑（命中测试、面板尺寸、
  释放判定、文本适配）有单元测试，以及用于版式的 PNG 预览渲染器，但
  没有自动化的像素回归测试。

运行手册见 [HANDOVER.md](HANDOVER.md)，扩展与轻量化计划见
[ROADMAP.md](ROADMAP.md)。

### 许可证

MIT —— 见 [LICENSE](LICENSE)。仅作为库使用的第三方模块：
`github.com/Microsoft/go-winio` (MIT)、`github.com/lxn/win` (BSD 风格)、
`golang.org/x/sys` (BSD 风格)。未内联任何第三方代码。
