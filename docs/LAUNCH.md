# LAUNCH — 对外发布素材与渠道（2026-11-11 建立，2026-11-16 刷新）

> 本文件管**数字与历史**：对外文案里可以用哪些实测值（不许超过）、发过什么、对方回了什么。
> **可直接粘贴的短文案**在 [PROMOTION.md](PROMOTION.md)：一句话版、Show HN、X、V2EX 各一套。
> 分工：PROMOTION 管「怎么说才有人看」，本文件管「能说的数是多少」。
> 这份文案最初（2026-11-11）由 `agy --model gemini-3.8-flash-high --effort high` 起草后逐条核对；
> **自 2026-11-13 起，AGENTS.md 第 8 条要求同会话复核**：实现、评审与措辞都在做这件事的会话里、用
> 用户选定的模型完成，不再调任何外部模型 CLI。**每个数字都以 [HANDOVER §33.6](HANDOVER.md) 那张基线表为准**，
> 改动数字前先跑一遍那里的命令，不要凭记忆写。

## 0. 发布顺序（**已全部就绪**）

```
P0/I1-α 已交付 ──► I1-γ 自解压 exe 已交付 ──► I1-β 版本号已付清（当前 1.0.21 / 1.2.31）──► ✅ 现在可以发声
```

**前提条件全部满足，可以发了。** 浮窗 1.0.21（Windows / Linux X11 / macOS AppKit）、插件 1.2.34、npm `latest` = 1.2.34、`pet-v1.0.21` 的
安装包与各平台压缩包都已构建并在发布时传回复验。
发帖时用 `https://github.com/jjjjjjjjnnjnn/ai-terminal-pet/releases/tag/pet-v1.0.21` 这个链接。

**唯一还需要人做的一步**：`docs/assets/smartscreen-unsigned.png` 是**示意图**（图左下角写着 MOCKUP）。
真机截图只有一台「从没构建过本项目」的机器才拍得到；如果能在发帖前补一张真的，更好；补不到就用示意图，
**但帖子的说明里必须写明它是重建的**，不要冒充截图。

## 1. 渠道优先级

| 渠道 | 权重 | 形态 | 备注 |
|---|---|---|---|
| **DSH 官方 Discussions** | **最高（受众最准）** | 「Show Your Plugins!」专区的单帖 | ✅ **已发：[#13001](https://github.com/deepseek-ai/deepseek-harness/discussions/13001)**（2026-11-11T20:50:40Z）。受众全部是 DSH 用户，且该区有 1.2 万帖的活跃度；规则见 §1.2 |
| **Show HN** | 高 | 一条 URL 提交 + 作者首楼技术补充 | ✅ 你已发布。周一–周四 08:00–10:00 ET；周五与周日几乎必沉 |
| **X（英文线程）** | 高 | 5 条，链接放最后一条 | ✅ 你已发布。不做「RT/关注」诱导；GIF 挂在第 1 条 |
| **X（中文线程）** | 中高 | 5 条，同一批事实，中文原生表达 | ✅ 你已发布。与英文线程**同一天内**发，不要隔周 |
| **DSH 生态（GitHub 目录）** | 中但最确定 | 注册表条目跟进 PR、topic 采集、自建索引 | 实测证据：`v1.2.3` 的 tarball 被下载 **47** 次，而 `v1.2.4`/`v1.2.5` 只有 4/5 次（2026-11-12 复读：`v1.2.5` 已追到 13、`v1.2.6` 是 7）——**把已有条目推到最新版比发帖更有效**。现状见 §1.1 |
| r/ClaudeAI / r/ChatGPTCoding | 中 | 英文帖改写为问答体 | 你已说明不发 |
| V2EX / 掘金 / 少数派 | 中 | 中文长文（带 GIF） | 你已说明不发 |

**不要投的**：`WOOK98/awesome-deepseek-harness-plugins`（imsai 的 fork、已停更）、
`ydhrdh/dsh-marketplace`（7 条、0 stars）。详见 HANDOVER §28.3。

### 1.2 DSH 官方 Discussions 的发帖规则与我们的合规情况

该区置顶了两帖：`#2004 Plugin Category Guidelines` 与 `#1797 Welcome`。六条规则逐条对照：

| 规则 | 我们怎么满足的 |
|---|---|
| 1. 一贴一项目 | 一帖只讲 ai-terminal-pet |
| 2. 必须与 DSH 真实集成，禁止无关推广与报菜名 | 标题与正文都以 DSH 为主体；OpenCode/Claude Code/Antigravity 被降级为「可选能力」，且写明「这是可选能力，不是对 Harness 的前提」 |
| 3. 标题格式 `DSH \| 项目名 \| 一句话` | `DSH \| ai-terminal-pet \| Windows 置顶状态带：一眼看出哪个 DSH 会话在等审批`（55 字符） |
| 4. 正文含项目地址、简介、截图、与 DSH 的集成方式 | 四项齐备；集成方式给了**事件→状态映射表**，并说明 `approval/request` 是无条件 `next()` 透传 |
| 5. 显著标注非官方 | 首行 blockquote 加粗：「非官方项目…与 DeepSeek 官方没有隶属关系」 |
| 6. 建议模板 | 按模板顺序：非官方声明 → 项目地址 → 简介 → 集成方式 → 截图 |

**评审改了三处才会过门**（`agy --model gemini-3.8-flash-high --effort high` 逐条挑出来的，都是真的）：

1. **原稿把四个终端并列**，在 DSH 自己的插件区会被读成「借 DSH 流量推广竞品」——规则 2 明确禁止。
   改成 DSH 为主线，其余三个是可选聚合。
2. **「官方 protocol 模块」这个说法是错的**，在那个论坛里「官方」默认指 DeepSeek。`protocol.js`
   是本项目自己的模块，已改成「ai-terminal-pet 的统一协议模块」。
3. **`**_> …_**` 这种写法在 Markdown 里渲染不出引用块**，改成标准 `> **_…_**`。

另外两处补充是评审加的、也确实该有：`56×156`（原稿只写了 56）、以及 Release 直链。

### 1.1 已经投出去的那批，现在的真实状态（2026-11-11T21:05Z 复查，2026-11-12 08:5xZ 更新）

**「投了」不等于「有人看」**，下面这批是同一轮提交，状态差别很大（四家已收录，四家仍在排队）：

| 目标 | 状态 | 说明 |
|---|---|---|
| [awesome-dsh-plugin](https://github.com/awesome-dsh-plugin/awesome-dsh-plugin) | ✅ **1.2.6 的两次合并都完成**（#14385 → 1.2.5、#14439 → 1.2.6）；1.2.7 与 1.2.8 的跟进 PR [#14688](https://github.com/awesome-dsh-plugin/awesome-dsh-plugin/pull/14688)、[#14718](https://github.com/awesome-dsh-plugin/awesome-dsh-plugin/pull/14718) 都已开、MERGEABLE | 仓库里 `data/plugins/*.yml` 的 `tarball` 仍是 **1.2.6**（合并前不会变），站点 `plugins.json` 也已渲染 1.2.6（2026-11-12 实测 `count` 10,581）。两个跟进 PR 只改 `tarball` 一行，中文描述逐字节未动（base64 上传）；**#14718 是最新的那个**，等 check 与人工合并 |
| [dshopen](https://github.com/polang233/dsh-plugin-index/issues/423) | ✅ **已收录并已渲染** | `dshopen.org/index.json` **1,207 条**含本包 **1.2.4**（commit `ff7f6de`），`labels: verified + updated-recent`；**落后两个版本**，要等它下一次构建 |
| [YELEBAI/dsh-plugin-marketplace](https://github.com/YELEBAI/dsh-plugin-marketplace) | ✅ **自动采集，已收录** | `registry/plugins.json` 第 8,729 条：`version: 1.2.5`、`verifiedCommit: 584bc2f`、`install.spec` = `github:jjjjjjjjnnjnn/ai-terminal-pet#584bc2f`、`profiles: [desktop, headless, web]`、`verifiedAt` 17:54:34Z。**无需申请**，它按 topic 自动扫 |
| [bradeGithub/DSH-Plugins-Marketplace](https://github.com/bradeGithub/DSH-Plugins-Marketplace) | ✅ **自动采集，已收录** | `registry.json` 第 13,128 条：`pkg_name`/`npm_pkg_name` = `dsh-ai-terminal-pet`、`version: 1.2.4`、topics 15 个、`stars_delta_30d` 也在记。issue #367 只是留证，可关 |
| [dshbase #322](https://github.com/ylwl1997/dshbase/issues/322) | ⏳ open / **0 评论** | 08:35Z 提交，至今无人回应 |
| [dsh.so #24](https://github.com/dsh-so/dsh-plugin-submissions/issues/24) | ⏳ open / **0 评论** | 08:36Z |
| [dsh-plugin.org #735](https://github.com/dshplugin/dsh-plugin-hub/issues/735) | ⏳ open / **0 评论** | 08:46Z；这个库的 `Submission gate` 若会回应通常几分钟内就有，0 评论意味着没进队列 |
| [AllDSH #2](https://github.com/alldsh/submissions/issues/2) | ⏳ open / **0 评论** | 08:44Z |
| [imsai #912](https://github.com/imsai-sh/awesome-deepseek-harness-plugins/pull/912) | ✅ **已合并并已收录** | 目录里 `catalog/plugins/jjjjjjjjnnjnn--ai-terminal-pet--packages--dsh-ai-terminal-pet.json` 在，`category: notify`，中英摘要齐全；README 也列了。它是「目录文件」而不是「版本钉」，所以没有版本滞后问题 |
| [radar #1255](https://github.com/AdamPlatin123/dsh-plugin-radar/pull/1255) | ⏳ open / 零 check | 无回应 |

**结论（也是发给人的那句）**：**已收录的有四家**——awesome-dsh-plugin（仓库侧已 1.2.6）、
imsai（目录文件，无版本滞后）、YELEBAI 与 brade（两家都是按 topic 自动采的，不用申请）。
**版本号还落后的只剩一家**：dshopen 仍是 1.2.4（2026-11-12 复读 `index.json` 1,219 条，本包版本未动）——awesome-dsh-plugin 的站点已经渲染 1.2.6。「合并」与「渲染」、以及「采到了」与「渲染成最新版」，始终是两件事。
四家人工提交的目录（dshbase / dsh.so / dsh-plugin.org / AllDSH）截至 21:05Z 仍是 **open、0 评论**，
`dshmarketplace.dev` 的公开搜索仍回 `{"total":0}`。

**发帖时的措辞边界**：可以说「已在 X 个 DSH 目录里」，`X` 用**上面四家**的数；不要把「已提交」
写成「已被收录」，也不要把「收录」写成「市场推荐」——后两者在这几家都是别人的判断，不是我们能声明的。


## 2. 发布前必须就位的东西

- [x] **六个跨平台资产**（`pet-v1.0.21`）：`pet-setup.exe` / `ai-terminal-pet-1.0.21-win-x64.zip` /
      `linux-{amd64,arm64}.tar.gz` / `darwin-{arm64,amd64}.tar.gz`；字节与 sha256 见
      [CHANGELOG](../CHANGELOG.md) 的「1.0.21 / 1.2.34 发布包实测清单」。
      **已从公开 URL 下载回来逐字节核对，六个资产全部与本地相同**
- [x] 独立 tag `pet-v1.0.21` + Release；`scripts\install.cmd -VerifyOnly` **12/12**，
      `--doctor` 12 ok / 1 warn / 0 fail，`--version` = 1.0.21
- [x] 插件包 `dsh-ai-terminal-pet` **1.2.34**（npm `latest`、`dsh-plugin-v1.2.34` Release、
      上游 `+1 −1` 的跟进 PR **#16391**）；npm 上的 `.tgz` 与本地打包逐字节相同
- [x] OpenCode 上报端 `opencode-ai-terminal-pet` **1.0.21**（npm `latest`，与浮窗同号；`.tgz` 同样逐字节核对）
- [x] **GIF 动图**：`docs/assets/demo.gif`（25 帧 / 5.38 秒循环 / 359,471 B），三段真实动作
      —— 等待呼吸 → 悬停展开 → 点击等待行。循环时长 5.4 秒而不是 15 秒：GIF 是循环播放的，
      15 秒的循环第一遍没人看完，5 秒刚好一遍读完三句话。由 `docs/assets/make_launch_assets.py`
      从同一份光栅器输出合成，`README.md` 中英两段都已引用
- [x] **「两个终端 + 浮窗同框」实景图**：`docs/assets/desktop-scene.png`（1500x860，59,029 B）。
      面板那一半是真实光栅器输出，两个终端是画出来的，图注里写明了这一点
- [ ] **GitHub 社交预览图上传**：`docs/assets/social-preview.png`（1280x640）已生成，
      但仓库 Settings → Social preview **没有 API**，重生成文件不等于换上去。
      2026-11-17 尝试代传时本机浏览器解析不了 DNS（`ERR_NAME_NOT_RESOLVED`），等网络恢复或本人手传
- [x] `--doctor` 全绿（12 ok, 1 warn, 0 fail），截图 `docs/assets/doctor-green.png`
      （1798x358 / 63,335 B）已生成：那是程序自己的 stdout，只加了着色与窗口边框，
      并把捕获机器上的用户名换成 `<you>`
- [x] `docs/assets/smartscreen-unsigned.png`（1320x560 / 30,204 B）：**按官方对话框文字重建的示意图**，
      不是真机截图，图左下角写着 MOCKUP。真机截图需要一台"从没构建过本项目"的机器，
      而且会带上别人的桌面与机器名；这两张卡片把"你会看到什么、点哪里"说清楚就够了
- [ ] **awesome-claude-code 的表单**：仓库已满 55 天、当天仍有提交 → 它「≥14 天且活跃」这一档已满足。
      那一行描述在 [docs/PROMOTION.md](PROMOTION.md) §6，**提交只能由人做**（网页表单）
- [x] 基线数字记下来（[docs/METRICS.md](METRICS.md) §1，2026-11-17 实测）：14 天 33 浏览 / 11 唯一、
      star 1、issue 0、release 资产 ≈2 件/版本（= 自己的复验）、npm **latest share 6%**（爬虫队形）。
      **这一版之后对外只说 latest share，不再说下载总量**

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
highest-priority state across every session: which project waits, what for, how long.

Two choices worth explaining:

State is pushed as one JSON line per event over a per-user Windows named pipe. Session
lifecycle is not pushed at all: it is verified against the OS process table and window
titles. A reporter going silent does not end a session; one idle for an hour is not dead.
Dropping it takes two minutes of unobservability. Clicking a row activates that terminal
window, or the right tab in Windows Terminal.

Permission prompts get an optional 5 second veto window: a draining bar, tap to deny,
silence to allow. Off by default. Destructive commands (rm -rf, rd /s, git push --force,
format) are checked before the countdown and always go to manual review.

Windows, Linux (X11) and macOS (AppKit), one CGO-free Go codebase. Out of the box only
DeepSeek Harness sessions appear, through the DSH plugin (bundled overlay on Windows,
one pinned download on Linux and macOS); the other three
terminals need their reporter wired once (one line, README table). Not code-signed, so
first run shows SmartScreen. Three native executables: a 4.50 MiB overlay, a statusline
shim, a hook. MIT, no account, no network, no telemetry. 463 Go tests plus a 172-check
fresh-machine suite.

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

### 3.1 下一次再发时用的开头（2026-11-16，依据 README 重写；**未发布**）

上面 §3/§4/§5 是**已经发过**的版本，保留原文作为历史。如果还有下一次（新版本、新渠道、或别人转载时
要一句准确的介绍），用下面这段：它把痛点放在第一句，而不是先讲架构。

**一句话介绍（仓库描述/目录条目/转发时用）**

```text
Your agents are waiting. You are in another tab. A 56px always-on-top strip shows which AI
coding agent is blocked on you, for what, and for how long - then jumps you back in one click.
```

**帖子开头（≤120 词，可接 §3 正文其余部分）**

```text
Four agents, four tabs. One of them asks for permission, gets no answer, and waits - because
you are reading a diff in another window, or on another monitor, or that tab is simply behind
this one. Nothing crashed. You just paid fifteen minutes of an agent's life for nothing, and you
will only notice when you switch back.

A bell does not fix that. Neither does a spinner inside each terminal: the information exists,
but it is inside the thing you are not looking at. So I put it where you cannot miss it - a
56px strip, always on top, showing every local agent session, what it is asking for, and how
long it has been asking.
```

**发布时仍然必须守住的三条**：数字只引用末尾那套实测（461→**463** 个 Go 测试、**172** 项 fresh-machine、
**4.50 MiB** 浮窗、`--doctor` **13** 项）；不写"已收录"只写"已提交/已被渲染"（§1.1 的分界线）；
不写任何我们没验证过的东西——macOS 面板与通知、Linux 的真实出声/弹窗，一律按 README 的
"How this is verified — and what is not" 原样说。

## 4. X 线程（英文，5 条，每条 ≤270 字符）```text
1/5 Running Claude Code, OpenCode and Antigravity in three terminals means one of them eventually sits blocked on a permission prompt while you are focused in another window. I built a 56px strip that shows which project is waiting, and for how long.

2/5 State arrives as one JSON line per event over a per-user Windows named pipe. Lifecycle is not heartbeat-based: the overlay checks the OS process table and window titles, so a reporter that goes quiet never ends a session. Click a row to jump back to that terminal.

3/5 Auto-approve is off by default and opt-in per terminal. Before the 5-second countdown starts a denylist is checked: rm -rf, rd /s, format, git push --force. Those always go to manual review; silence only approves an ordinary tool prompt you chose to delegate.

4/5 Three native exes, no runtime at install: the 4.50 MiB overlay, an Antigravity statusline shim, a Claude Code hook. DeepSeek Harness connects with one command; the other three each get one line in their config, from a README table.

5/5 Windows / Linux (X11) / macOS (AppKit), one CGO-free Go codebase. MIT, no account, no network, no telemetry. 463 Go tests plus a 172-check fresh-machine suite. Not code-signed yet, so first run shows SmartScreen. https://github.com/jjjjjjjjnnjnn/ai-terminal-pet
```

字数实测（含链接、含空格、含 `n/5` 前缀）：250 / 268 / 263 / 235 / 265。
**第 1 条挂 `docs/assets/demo.gif`**，第 2 条挂 `docs/assets/overlay-expanded.png`，
第 4 条挂 `docs/assets/smartscreen-unsigned.png`（未签名这件事先说在前面，
而不是等人来问），第 5 条挂 `docs/assets/desktop-scene.png`。GIF 在 X 上要 ≤5 MB / ≤15 MB（移动端另算），
342 KB 远在限内；HN 不支持 GIF，Show HN 帖里改贴 `desktop-scene.png` 的直链。

**Show HN 的正文不发图**，HN 的帖子是纯文本，图只能放在链接的仓库里——所以 README 里的
这四张图是 HN 那条帖子的实际配图位置。

## 5. X 线程（中文，5 条，每条 ≤270 字符）

```text
1/5 同时开着 Claude Code、OpenCode 和 Antigravity，总有一个会卡在权限确认里，而你在另一个窗口忙。我做了个 56 像素宽的状态带，告诉你哪个项目在等、等了多久。

2/5 状态以每行一条 JSON 经按用户隔离的 Windows 命名管道送达。会话存活不靠心跳：浮窗直接看进程表和窗口标题，所以上报端沉默也不会误关会话。点一行就跳回那个终端。

3/5 自动同意默认关闭、按终端单独开启。倒计时开始前先过黑名单：rm -rf、rd /s、format、git push --force 一律走人工确认；沉默只放行你主动委托的普通工具调用。

4/5 三个原生 exe，装完不需要任何运行时：4.50 MiB 浮窗、Antigravity 状态栏 shim、Claude Code hook。Harness 一条命令接入，另外三个各加一行配置，README 里有表。

5/5 支持 Windows / Linux (X11) / macOS (AppKit)，同一份零 cgo 的 Go 代码。MIT，无账号、不联网、无遥测。465 个 Go 单测 + 183 项全新机器端到端。还没代码签名，首次运行会看到 SmartScreen。https://github.com/jjjjjjjjnnjnn/ai-terminal-pet
```

字数实测（Unicode 字符）：98 / 89 / 96 / 111 / 181。
中文帖比英文帖短得多，这是**故意的**：中文读者更愿意点进去看长图，而不是在时间线上读长句。

## 6. DSH 市场条目描述

市场列表只显示摘要，所以它必须是**事实**而不是卖点，而且**它已经存在，不能直接换**：
`docs/market-index.json` 的 `plugins[0].description`、`packages/dsh-ai-terminal-pet/package.json`
的 `dsh.catalog.summary`、以及仓库根 `package.json` 的同一块，三处是**逐字相同的一段文字**
（`verify.mjs` 有断言钉住它们），并且它属于已发布的 1.2.30 tarball 的字节。

**1.2.34 起用的（中英文各一段）**

```text
Which of your agents is waiting on you? One always-on-top panel for DeepSeek Harness, Claude Code, OpenCode and Antigravity sessions: it rings only for real events, and the arrow at a row's end jumps to that terminal. Harness sessions appear on install; OpenCode is one command; Claude Code and Antigravity need their reporter.
```

```text
哪个 agent 在等你？一块置顶面板汇总 DeepSeek Harness、Claude Code、OpenCode 与 Antigravity 的会话：只在真事件上响，点行尾箭头即可跳到对应终端。装上即接上 Harness 的会话；OpenCode 只需一行命令；Claude Code 与 Antigravity 各需接一次上报端。
```

**这两段为什么仍然要交代"谁还要接上报端"**：这条声明是市场条目页的摘要，读的人会据此判断
"装上就能看到全部终端吗"。这个 bundle **开箱即用映射的仍然只有 Harness 自己的会话**；区别只在
OpenCode 现在是一条命令（`opencode plugin add opencode-ai-terminal-pet`，Windows 上连面板一起装），
而 Claude Code 与 Antigravity 仍要走各自的 `reporters/*.js`（或 `pet-hook`）。1.2.4 就是因为原稿读起来像
"全都开箱可用"才只改文案发了一版；**这条底线不能为了更短的文案而丢**。1.2.34 只把"另外三个"改成
"一个一行、两个要接"，没有把 OpenCode 说成零配置。

想换措辞时：三处清单 + `docs/market-index.json` 的同一段一起改，再跑
`node packages/dsh-ai-terminal-pet/verify.mjs`。纯文案改动的代价仍是一个新版本号（tag + Release + npm + 上游 PR）。

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

没有遥测，所以只能看代理指标。**全部指标用 `node tools/traffic-report.mjs` 一次量完**
（口径与判读见 [docs/METRICS.md](METRICS.md)）：

| 指标 | 怎么读 | 基线（2026-11-17 实测） |
|---|---|---|
| Release 资产下载 | `traffic-report.mjs` 的 INSTALLS 段 | 按类别：DSH tarball 431（12.3/版本）、setup 41（**2.0/版本**）、win zip 43（2.0/版本）、POSIX 75（1.9/版本）。**≈2 件/版本 = 本项目自己的发版复验**，不是用户 |
| 仓库浏览 | GitHub traffic（14 天窗口） | **33 views / 11 uniques**；referrer 只有 github.com 15、HN 3、目录站 1 |
| clone | 同上 | 1,798 次里 **852** 直接由 213 次 CI run × 4 job 解释；**不要把 clone 当人气** |
| star / fork | `gh repo view` | **1 / 0**（那 1 个是作者本人） |
| npm | `npm view dsh-ai-terminal-pet version` | **1.2.34**（`dist-tags.latest`，2026-11-17 发版同步）；OpenCode 上报端 `opencode-ai-terminal-pet` 同日 **1.0.21** |
| DSH Discussions | 帖子 #13001 的 upvote 与回复数 | **1 / 0**（发帖后已数日未动） |
| **npm 下载量（会被爬虫放大，不能当用户数）** | `node tools/traffic-report.mjs` | 2026-11-17 实测 **4,586/周**，但它分布在 **25 个版本、每个 137–227 次**，连不可安装的 `0.0.0-stage` 都有 166 次 → **latest share 6%**，是爬虫队形不是用户（对照 `dsh-pet` 的 58%）。证据链见 [docs/METRICS.md](METRICS.md) §2；**任何对外文本都不要引用这个总量** |
| **目录那行（只作参考，不当数据源）** | `node tools/catalog-status.mjs` | 站点写 `downloads: 3,455 (30d)`、`version: 1.2.27`、安装命令解析到最新版。**这三个都落后于现实，且都不归我们修**：站点数字来自同一个被爬虫放大的 npm 探针；版本停在 1.2.27 是对方夜间构建未成功（见 [docs/UPSTREAM.md](UPSTREAM.md) §4.1）。**对外引用一律不引用下载量；要引用就引用 latest share 与窗口** |
| 社区健康度 | `gh api .../community/profile` | 85%（原因见 HANDOVER §33.4） |
| issue 类型分布 | `gh issue list --json labels` | 0 条（本仓库；上游报的两条是该仓库的 issue，不是这里的） |

**判断渠道是否真的起效，看"减去基线后的安装增量"，不看点赞数。** 发帖后 24 小时内的
Release 资产下载增量是唯一能说明「有人真的装了」的信号——**但要先扣掉本项目自己的复验下载
（≈2 件/版本）**，否则你会把自己的核对当成用户。star 与转发都可能只是路过。

**判断「目录里那行是不是当前版本」，跑 `node tools/catalog-status.mjs`，不要开浏览器**：
它把"该我们动"与"对方的原因"分开写清楚（`docs/UPSTREAM.md` §4.1），
而 PR 页面上的 `merged` 既不等于 `published`，也不等于 `rebuilt`。

**判断"有人真的在用"，看 npm 的 latest share，不看任何下载总量**：总量两千多次里绝大部分是
按版本爬取的机器流量（[docs/METRICS.md](METRICS.md) §2 有三条证据）。真要对外说一句，就说
"最近一周 downloads 里 X% 落在最新两个版本"，并且写上窗口与日期。
