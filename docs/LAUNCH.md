# LAUNCH — 对外发布素材与渠道（2026-11-11）

> 本文件是**对外可见文案的唯一存放处**：X 线程、Show HN 帖、市场条目描述、issue 回复模板。
> 文案由 `agy --model gemini-3.8-flash-high --effort high` 起草，再逐条按本仓库实测事实核对过
> （AGENTS.md 第 8 条）。**每个数字都以 [HANDOVER §33.6](HANDOVER.md) 那张基线表为准**，
> 改动数字前先跑一遍那里的命令，不要凭记忆写。

## 0. 发布顺序（**已全部就绪**）

```
P0/I1-α 已交付 ──► I1-γ 自解压 exe 已交付 ──► I1-β 版本号已付清（1.0.0 / 1.2.6）──► ✅ 现在可以发声
```

**前提条件全部满足，可以发了。** 浮窗 1.0.0、插件 1.2.6、npm `latest` = 1.2.6、`pet-v1.0.0` 的
zip 与 `pet-setup.exe` 都在线上并且都从公网下载回来实测过，四个数字终于一致。
发帖时用 `https://github.com/jjjjjjjjnnjnn/ai-terminal-pet/releases/tag/pet-v1.0.0` 这个链接。

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
| **DSH 生态（GitHub 目录）** | 中但最确定 | 注册表条目跟进 PR、topic 采集、自建索引 | 实测证据：`v1.2.3` 的 tarball 被下载 **47** 次，而 `v1.2.4`/`v1.2.5` 只有 4/5 次——**把已有条目推到最新版比发帖更有效**。现状见 §1.1 |
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

### 1.1 已经投出去的那批，现在的真实状态（2026-11-11T21:05Z 复查）

**「投了」不等于「有人看」**，下面这批是同一轮提交，状态差别很大（四家已收录，四家仍在排队）：

| 目标 | 状态 | 说明 |
|---|---|---|
| [awesome-dsh-plugin](https://github.com/awesome-dsh-plugin/awesome-dsh-plugin) | ✅ **两次合并都完成**：#14385（17:20:15Z，`70581cc`）→ 1.2.5；**#14439（20:12:31Z，`475a48e`）→ 1.2.6** | 仓库里 `data/plugins/*.yml` 的 `tarball` 已是 **1.2.6**；**站点 `plugins.json`（`count` 10,578）仍渲染 1.2.5**，等它下一次构建 |
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
**两家的版本号还落后**：awesome-dsh-plugin 的站点 `plugins.json` 仍渲染 1.2.5，dshopen 仍是 1.2.4——
「合并」与「渲染」、以及「采到了」与「渲染成最新版」，始终是两件事。
四家人工提交的目录（dshbase / dsh.so / dsh-plugin.org / AllDSH）截至 21:05Z 仍是 **open、0 评论**，
`dshmarketplace.dev` 的公开搜索仍回 `{"total":0}`。

**发帖时的措辞边界**：可以说「已在 X 个 DSH 目录里」，`X` 用**上面四家**的数；不要把「已提交」
写成「已被收录」，也不要把「收录」写成「市场推荐」——后两者在这几家都是别人的判断，不是我们能声明的。


## 2. 发布前必须就位的东西

- [x] `ai-terminal-pet-1.0.0-win-x64.zip`（`scripts/package.ps1`，17 文件 / 4,983,903 B / sha256 `70D992B1…`）
- [x] **`pet-setup.exe`**（14,125,568 B / sha256 `87D5E60B…`）：自解压单文件，双击即装；**先**校验载荷再动手
- [x] 独立的 `pet-v1.0.0` tag + Release（两个资产都已从公开 URL 下载回来实测：`--version` → `pet-setup 1.0.0`，沙箱里 install → `--verify` 11/11 → `--uninstall` 归零）
- [x] 插件包 `dsh-ai-terminal-pet` **1.2.6**（npm `latest`、`dsh-plugin-v1.2.6` Release、tarball sha256 `671CDC95…`）
- [x] **GIF 动图**：`docs/assets/demo.gif`（25 帧 / 5.38 秒循环 / 342,606 B），三段真实动作
      —— 等待呼吸 → 悬停展开 → 点击等待行。循环时长 5.4 秒而不是 15 秒：GIF 是循环播放的，
      15 秒的循环第一遍没人看完，5 秒刚好一遍读完三句话。由 `docs/assets/make_launch_assets.py`
      从同一份光栅器输出合成，`README.md` 中英两段都已引用
- [x] **「两个终端 + 浮窗同框」实景图**：`docs/assets/desktop-scene.png`（1500x860，59,029 B）。
      面板那一半是真实光栅器输出，两个终端是画出来的，图注里写明了这一点
- [x] `--doctor` 全绿（9 ok, 0 warn, 0 fail），截图 `docs/assets/doctor-green.png`
      （1798x358 / 63,335 B）已生成：那是程序自己的 stdout，只加了着色与窗口边框，
      并把捕获机器上的用户名换成 `<you>`
- [x] `docs/assets/smartscreen-unsigned.png`（1320x560 / 30,204 B）：**按官方对话框文字重建的示意图**，
      不是真机截图，图左下角写着 MOCKUP。真机截图需要一台"从没构建过本项目"的机器，
      而且会带上别人的桌面与机器名；这两张卡片把"你会看到什么、点哪里"说清楚就够了
- [x] 基线数字记下来（§33.6 已复读）：views 7 / stars 1 / 社区健康度 85% /
      Release 下载 `v1.2.3`=47、`v1.2.4`=4、`v1.2.5`=5、`pet-v0.9.0`=3（上线 40 分钟，未宣传）

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
a hook binary. No account, no network, no telemetry, MIT. 201 Go tests plus a 162-check
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

5/5 Windows only. MIT, no account, no network, no telemetry. 201 Go tests plus a 162-check fresh-machine suite. Not code-signed yet, so first run shows SmartScreen. https://github.com/jjjjjjjjnnjnn/ai-terminal-pet
```

字数实测（含链接、含空格、含 `n/5` 前缀）：250 / 268 / 263 / 235 / 213。
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

4/5 三个原生 exe，装完不需要任何运行时：4.17 MiB 浮窗、Antigravity 状态栏 shim、Claude Code hook。Harness 一条命令接入，另外三个各加一行配置，README 里有表。

5/5 仅 Windows。MIT，无账号、不联网、无遥测。201 个 Go 单测 + 162 项全新机器端到端。还没代码签名，首次运行会看到 SmartScreen。https://github.com/jjjjjjjjnnjnn/ai-terminal-pet
```

字数实测（Unicode 字符）：98 / 89 / 96 / 111 / 132。
中文帖比英文帖短得多，这是**故意的**：中文读者更愿意点进去看长图，而不是在时间线上读长句。

## 6. DSH 市场条目描述

市场列表只显示摘要，所以它必须是**事实**而不是卖点，而且**它已经存在，不能直接换**：
`docs/market-index.json` 的 `plugins[0].description`、`packages/dsh-ai-terminal-pet/package.json`
的 `dsh.catalog.summary`、以及仓库根 `package.json` 的同一块，三处是**逐字相同的一段文字**
（`verify.mjs` 有断言钉住它们），并且它属于已发布 1.2.5 tarball 的字节。

**当前线上用的（225 字符英文 / 114 字符中文，已在 1.2.5 里）**

```text
A floating overlay for Windows. DeepSeek Harness sessions show up as soon as you install this bundle; OpenCode, Claude Code and Antigravity sessions join the same panel once each of those terminals has its reporter installed.
```

```text
Windows 桌面浮窗。装上本插件后 DeepSeek Harness 的会话立刻出现在置顶面板里；OpenCode、Claude Code 与 Antigravity 的会话要先各自装好该终端的上报端，才会汇总进同一个面板。
```

它已经说清了「开箱即用只有 Harness」，所以 **1.2.6 不必为文案而发**。

**下一版（1.2.6）的候选措辞**（agy 起草，211 / 114 字符，更短、更像条目摘要）：

```text
Windows only. Native floating overlay tracking DeepSeek Harness session states out of the box. Other terminals (Claude Code, OpenCode, Antigravity) require their own reporter wired. Zero telemetry, MIT licensed.
```

```text
仅限 Windows。原生桌面悬浮窗，安装后直接显示 DeepSeek Harness 会话状态；接入 Claude Code、OpenCode 或 Antigravity 需单独配置各自的上报器。零网络遥测，MIT 开源协议。
```

**用哪一段取决于 1.2.6 那天**：如果那时没有别的东西要发，保持现有文字、跳过这次改号更划算
——纯文案改动的代价是 tag + Release + npm + 一个上游 PR（1.2.1 与 1.2.4 都是这么来的）。
真正改的时候，上面这份候选要同时落进三处清单与 `docs/market-index.json` 的同一段，
再跑 `node packages/dsh-ai-terminal-pet/verify.mjs`。

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

| 指标 | 怎么读 | 基线（2026-11-11 21:10Z） |
|---|---|---|
| Release 下载数 | `gh api repos/.../releases` | `pet-v1.0.0` = **4**（zip 2 + setup 2）／`pet-v0.9.0` = 3／`v1.2.6` = 1／`v1.2.5` = 13／`v1.2.4` = 4 |
| 仓库浏览 | `gh api .../traffic/views` | 30 天 7 views / 2 uniques（发帖前，尚未刷新） |
| star / fork | `gh repo view` | 1 / 0 |
| npm | `npm view dsh-ai-terminal-pet version` | **1.2.6**（`dist-tags.latest`） |
| DSH Discussions | 帖子 #13001 的 upvote 与回复数 | 0 / 0（刚发） |
| 社区健康度 | `gh api .../community/profile` | 85%（原因见 HANDOVER §33.4） |
| issue 类型分布 | `gh issue list --json labels` | 0 条 |

**判断渠道是否真的起效，看下载数的增量，不看点赞数。** 发帖后 24 小时内的 Release 下载增量
是唯一能说明「有人真的装了」的信号；star 与转推都可能只是路过。
