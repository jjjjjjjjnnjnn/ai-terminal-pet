# LAUNCH — 对外发布素材与渠道（2026-11-11）

> 本文件是**对外可见文案的唯一存放处**：X 线程、Show HN 帖、市场条目描述、issue 回复模板。
> 文案由 `agy --model gemini-3.8-flash-high --effort high` 起草，再逐条按本仓库实测事实核对过
> （AGENTS.md 第 8 条）。**每个数字都以 [HANDOVER §33.6](HANDOVER.md) 那张基线表为准**，
> 改动数字前先跑一遍那里的命令，不要凭记忆写。

## 0. 发布顺序（不要打乱）

```
P0/I1-α 已交付 ──► I1-γ 自解压 exe（可选，静默上线） ──► I1-β 版本号一次性付清 ──► P4 正式发声
```

**正式发声排在 I1-β 之后。** 在 0.9.0 上先宣传再改号，评论区第一条就会是
「README 说 1.2.5、npm 是 1.2.5、exe 却是 0.9.0」。已经存在的分发路径（市场条目、topic 采集）
可以先静默更新，那不算发声。

## 1. 渠道优先级

| 渠道 | 权重 | 形态 | 备注 |
|---|---|---|---|
| **Show HN** | 高 | 一条 URL 提交 + 作者首楼技术补充 | 周一–周四 08:00–10:00 ET；周五与周日几乎必沉 |
| **X（英文线程）** | 高 | 5 条，链接放最后一条 | 不做「RT/关注」诱导；GIF 挂在第 1 条 |
| **X（中文线程）** | 中高 | 5 条，同一批事实，中文原生表达 | 与英文线程**同一天内**发，不要隔周 |
| **DSH 生态（GitHub）** | 中但最确定 | 注册表条目跟进 PR、topic 采集、自建索引 | 实测证据：1.2.3 的 tarball 被下载 **46** 次，1.2.4/1.2.5 只有 3/4 次——**把已有条目推到最新版比发帖更有效** |
| r/ClaudeAI / r/ChatGPTCoding | 中 | 英文帖改写为问答体 | HN 之后 4 小时再发，避免同一批人重复看到 |
| V2EX / 掘金 / 少数派 | 中 | 中文长文（带 GIF） | 第三日再发，不要与 X 同日 |

**不要投的**：`WOOK98/awesome-deepseek-harness-plugins`（imsai 的 fork、已停更）、
`ydhrdh/dsh-marketplace`（7 条、0 stars）。详见 HANDOVER §28.3。

## 2. 发布前必须就位的东西

- [x] `ai-terminal-pet-0.9.0-win-x64.zip`（`scripts/package.ps1`，17 文件，SHA256 已算）
- [x] 独立的 `pet-v0.9.0` tag + Release：资产 `ai-terminal-pet-0.9.0-win-x64.zip`，4,983,915 B，sha256 `EF0FC0D0…`。已从公开 URL 下载回来、解压、`install` → `-VerifyOnly`（11 件全对）→ `-Uninstall -Force`（rc=0）各跑过一遍。
- [ ] **SmartScreen 那张截图**：帖子里必须提前放「更多信息 → 仍要运行」，否则第一条评论就是「未签名」
- [ ] **15 秒 GIF**：等待脉冲 → 展开看到项目名与「等了 3 分钟」→ 点击跳回终端。
      现有 `docs/assets/*.png` 是静态面板，缺动图与「两个终端 + 浮窗同框」的实景图
- [ ] `--doctor` 全绿截图（9 ok, 0 warn, 0 fail），用来回答「我装了没反应」
- [ ] 基线数字记下来（§33.6）：views 7 / stars 1 / Release 下载 46:3:4 / npm 首日

## 3. Show HN

**标题（61 字符，≤80）**

```
Show HN: A 56px strip that shows which AI terminal is blocked
```

**正文（259 词，≤260）**

```text
I run three AI coding terminals at once - Claude Code, OpenCode, Antigravity - and the
failure mode is always the same: one asks for permission, I am focused in another window,
and twelve minutes later I notice the build sat idle.

So I built ai-terminal-pet: a 56 by 156 pixel strip at the screen edge showing the
highest-priority state across every session - which project waits, what for, how long.

Two choices worth explaining:

State is pushed as one JSON line per event over a per-user Windows named pipe. Session
lifecycle is not pushed at all: it is verified against the OS process table and window
titles. A reporter going silent does not end a session, and one idle for an hour is not
dead. Dropping it takes two minutes of unobservability. Clicking a row activates that
terminal window, or the right tab in Windows Terminal.

Permission prompts get an optional 5 second veto window: a draining bar, tap to deny,
silence to allow. Off by default. Destructive commands - rm -rf, rd /s, git push --force,
format - are checked before the countdown and always go to manual review.

Limits: Windows only. Out of the box only DeepSeek Harness sessions appear; the other
three need their reporter wired once (one line per terminal, table in the README). Not
code-signed, so first run shows SmartScreen.

Three native executables, nothing to install first: a 4.17 MiB overlay, a statusline shim,
a hook binary. No account, no network, no telemetry, MIT. 190 Go tests plus a 162-check
suite that installs into a throwaway sandbox.

https://github.com/jjjjjjjjnnjnn/ai-terminal-pet
```

**作者首楼补充（116 词，≤120）**

```text
What was tried first: sessions reported on events only, and a session was treated as ended
when reports stopped - a 15 second snapshot TTL, a 30 second pipe TTL, the file deleted on a
timer. It looked fine and was wrong constantly: reporters are event driven and de-duplicate
identical states, so a twenty minute build looks exactly like a dead terminal, and a quiet
user looks exactly like a closed one.

The fix inverted the question. State can be stale; existence comes from the OS - process
table and window titles - and it takes two minutes of being unobservable to drop a session.
Three tombstone kinds keep closed, expired and evicted apart, so a late report can heal an
entry.
```

## 4. X 线程（英文，5 条，每条 ≤270 字符）

```text
1/5 Running Claude Code, OpenCode and Antigravity in three terminals means one of them eventually sits blocked on a permission prompt while you are focused in another window. I built a 56px strip that shows which project is waiting, and for how long.

2/5 State arrives as one JSON line per event over a per-user Windows named pipe. Lifecycle is not heartbeat-based: the overlay checks the OS process table and window titles, so a reporter that goes quiet never ends a session. Click a row to jump back to that terminal.

3/5 Auto-approve is off by default and opt-in per terminal. Before the 5-second countdown starts a denylist is checked: rm -rf, rd /s, format, git push --force. Those always go to manual review; silence only approves an ordinary tool prompt you chose to delegate.

4/5 Three native exes, no runtime at install: the 4.17 MiB overlay, an Antigravity statusline shim, a Claude Code hook. DeepSeek Harness connects with one command; the other three each get one line in their config, from a README table.

5/5 Windows only. MIT, no account, no network, no telemetry. 190 Go tests plus a 162-check fresh-machine suite. Not code-signed yet, so first run shows SmartScreen. https://github.com/jjjjjjjjnnjnn/ai-terminal-pet
```

字数实测（含链接、含空格、含 `n/5` 前缀）：250 / 268 / 263 / 235 / 213。
**第 1 条挂 GIF**，第 2 条挂 `docs/assets/overlay-expanded.png`。

## 5. X 线程（中文，5 条，每条 ≤270 字符）

```text
1/5 同时开着 Claude Code、OpenCode 和 Antigravity，总有一个会卡在权限确认里，而你在另一个窗口忙。我做了个 56 像素宽的状态带，告诉你哪个项目在等、等了多久。

2/5 状态以每行一条 JSON 经按用户隔离的 Windows 命名管道送达。会话存活不靠心跳：浮窗直接看进程表和窗口标题，所以上报端沉默也不会误关会话。点一行就跳回那个终端。

3/5 自动同意默认关闭、按终端单独开启。倒计时开始前先过黑名单：rm -rf、rd /s、format、git push --force 一律走人工确认；沉默只放行你主动委托的普通工具调用。

4/5 三个原生 exe，装完不需要任何运行时：4.17 MiB 浮窗、Antigravity 状态栏 shim、Claude Code hook。Harness 一条命令接入，另外三个各加一行配置，README 里有表。

5/5 仅 Windows。MIT，无账号、不联网、无遥测。190 个 Go 单测 + 162 项全新机器端到端。还没代码签名，首次运行会看到 SmartScreen。https://github.com/jjjjjjjjnnjnn/ai-terminal-pet
```

字数实测（Unicode 字符）：98 / 89 / 96 / 111 / 132。
中文帖比英文帖短得多，这是**故意的**：中文读者更愿意点进去看长图，而不是在时间线上读长句。

## 6. DSH 市场条目描述（两版，入库前必须与实际条目一致）

市场列表只显示摘要，所以它必须是**事实**而不是卖点。

**English（211 字符，≤220）**

```text
Windows only. Native floating overlay tracking DeepSeek Harness session states out of the box. Other terminals (Claude Code, OpenCode, Antigravity) require their own reporter wired. Zero telemetry, MIT licensed.
```

**中文（114 字符，≤120）**

```text
仅限 Windows。原生桌面悬浮窗，安装后直接显示 DeepSeek Harness 会话状态；接入 Claude Code、OpenCode 或 Antigravity 需单独配置各自的上报器。零网络遥测，MIT 开源协议。
```

这条摘要与 `docs/market-index.json` 的 `plugins[0].description`、
`packages/dsh-ai-terminal-pet/package.json` 的 `dsh.catalog`（以及仓库根 shim 的那份）
**必须是同一份文字**——`verify.mjs` 有断言把它们钉在一起。

## 7. Issue / PR 回复模板

「为什么每个终端都要单独装上报端？」

```text
Because no AI coding CLI exposes a common IPC surface. Each one reports through its own
mechanism: a hook callback, a statusline command, or a plugin API. A small per-terminal
reporter translates that into one JSON-lines protocol over a local named pipe, so the
overlay stays independent of every host and adding a terminal never touches the renderer.
```

「为什么未签名？」/「SmartScreen 报警怎么办？」

```text
There is no code-signing certificate yet, so the release is unsigned and Windows shows its
unknown-publisher prompt: click More info, then Run anyway. What you can verify instead is
SHA256SUMS.txt inside the zip - one hash per file - and the whole thing builds from source
with go build in a few seconds if you would rather not trust a binary.
```

「装了但什么都不显示？」

```text
Run this and paste the whole output: & "$env:LOCALAPPDATA\AI-Terminal-Pet\AI-Terminal-Pet.exe" --doctor

Nine checks cover the nine usual causes. If it is all green, the overlay is fine and the
problem is on the terminal side: that terminal needs its reporter wired, and installing the
files alone does not wire it. The connect table in the README has one line per terminal.
```

## 8. 发完之后看什么

没有遥测，所以只能看代理指标：

| 指标 | 怎么读 | 基线（2026-11-11） |
|---|---|---|
| Release 下载数 | `gh api repos/.../releases` | 1.2.3 = 46 / 1.2.4 = 3 / 1.2.5 = 4 |
| 仓库浏览 | `gh api .../traffic/views` | 30 天 7 views / 2 uniques |
| star / fork | `gh repo view` | 1 / 0 |
| npm | `npm view dsh-ai-terminal-pet version` | 1.2.5 |
| issue 类型分布 | `gh issue list --json labels` | 0 条 |

**判断渠道是否真的起效，看下载数的增量，不看点赞数。** 发帖后 24 小时内的 Release 下载增量
是唯一能说明「有人真的装了」的信号；star 与转推都可能只是路过。
