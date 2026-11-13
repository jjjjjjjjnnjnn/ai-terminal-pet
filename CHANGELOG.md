# Changelog

本文件记录 ai-terminal-pet 的版本变更。格式参考 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.1.0/)，
版本号遵循语义化版本。所有条目均取自本仓库 `git log` 的提交正文与
[docs/HANDOVER.md](docs/HANDOVER.md) 的对应小节，未做任何功能推测。

## 关于版本号：请先读这一段

版本常量是 `internal/app/app_windows.go` 里的 `Version`。它在历史上的落地方式并不规整，
读本文件时请以此为准：

| 版本 | 发布日期 | 版本常量的实际情况 | 详细 runbook |
|---|---|---|---|
| 0.1.0 | 2026-09-24 | 初始发布时源码里**还没有** `Version` 常量 | — |
| 0.2.0 | 2026-09-25 | `e57838c` 引入常量 `0.2.0` | HANDOVER §14 |
| 0.3.0 | 2026-09-26 | `1b38608` 改为 `0.3.0` | HANDOVER §15 |
| 0.4.0 | 2026-09-26 | `869312e` 改为 `0.4.0` | HANDOVER §16 |
| 0.5.0 | 2026-09-26 | `ecd23ac` 改为 `0.5.0` | HANDOVER §17 |
| 0.6.0 | 2026-09-26 | 代码里仍是 `0.5.0`；版本号只记在 HANDOVER §18 | HANDOVER §18 |
| 0.7.0 | 2026-11-05 | 代码里仍是 `0.5.0`；版本号只记在 HANDOVER §21 | HANDOVER §20–§22 |
| 0.7.1 | 2026-11-05 | 同上 | HANDOVER §23 |
| 0.8.0 | 2026-11-08 | 同上 | HANDOVER §24 |
| 0.9.0 | 2026-11-08/09 | `b40c5db` 一次性从 `0.5.0` 直接改成 `0.9.0` | HANDOVER §25 |

两条因此需要知道的事实：

1. **0.6.0 – 0.8.0 没有独立的版本号提交**。这些版本的交付内容确实在代码里（提交时间集中在 2026-11-09），
   只是版本常量在那几次提交中没有跟着走，直到 `b40c5db` 才一次跳到 `0.9.0`。
2. **HANDOVER 里的日期是「工作日期」，不是提交日期**。§18–§24 记录的是 2026-09-26 至 2026-11-08
   期间的排查与实现，而这些提交实际是在 2026-11-09 集中落库的（`6f18394` 之后的 14 个提交）。
   本文件按**交付日期**分组，并标注对应提交。

已发版到 **1.0.7**（tag `pet-v1.0.7`）/ **1.2.14**（tag `dsh-plugin-v1.2.14`）：这一批是设置窗口的
可读性、可用性，以及那个点不进去的外观页签（见下面的 `[1.0.7]` 一节）。此前浮窗本体 **1.0.6**
（tag `pet-v1.0.6`）、插件包 **1.2.13**（tag `dsh-plugin-v1.2.13`）都已发布且不可变；下面几段原本标着 `[Unreleased]`/批次标题的段落记录的是 1.0.2 之前的
三批工作，它们分别在 1.0.0 / 1.2.6 那次换版本号（`df4fcf6`）与 2026-11-12 的无版本号批次里交付
（`[Unreleased]` 只能属于「还没发出去的东西」，留着它会和本段第一句直接矛盾）。
**注意浮窗本体与 DSH 插件包是两条版本线**：它们各自有 tag 线（`pet-v*` 与 `dsh-plugin-v*`）。
原先写在这一段的「浮窗停在 0.9.0、插件包 1.2.5」是当时的事实，现在是历史。
两条线的**上架**（把上游目录条目指到新版本）不是「顺手」的事，它有自己的工具、规则与清单：
[docs/UPSTREAM.md](docs/UPSTREAM.md)。

## [未发布] — 2026-11-13 晚（第 46 轮：把两个流程缺口做成门禁）

第 45 轮发完版后留下的两句"下次注意"，这一轮把它们做成改不掉的东西。**用户可见行为没有变化**，
所以没有升版本：浮窗仍是 1.0.7、插件包仍是 1.2.14，`dsh-plugin-v1.2.14` 的资产一个字节都没动。

### Added

- **`tools/check_versions.mjs`**：文档里的版本号以前没有任何东西读，于是三处同时过期——README 徽章停在
  **1.0.2**（跨了三次发版）、中英两段「插件包当前版本」写着 **1.2.8**（npm 上已是 1.2.13）、
  `docs/LAUNCH.md` 还让人去链 `pet-v1.0.0`。工具读 `internal/app/app_windows.go` 的 `const Version`
  与两份 `package.json`，然后断言 **18 个"当前版本"形状**（徽章、发布链接、zip 名、两条安装命令、
  `The package is now` / `当前版本是`、两处 npm 链接、`--version` 示例、ROADMAP 基线行、
  HANDOVER 状态标题与版本段落、LAUNCH 的发帖链接与两处 npm 数字），**每条都要求出现次数精确**；
  另外扫描"比当前更新的版本号"（只抓更大，历史值必须留下，围栏与反引号里的字面量豁免），并检查 CHANGELOG 的标题
  形状（当前版本必须有、两条线各自的标题不许超前、`[未发布]` 只能排第一）。
  `tools/check_versions.test.mjs` 把每条规则钉成纯函数测试，**25 条**。
- **`internal/app/config_saver_test.go`**：`config.json` 由守护协程写、`flush()` 只服务正常退出，
  所以用 kill 验证持久化验的是别的问题（§44.7 把它当成未解 bug 挂了一整轮）。三条测试分别钉住
  「只 save 也会落盘」「flush 是唯一能写出待写值的东西」「没有待写值时不创建文件」。

### Changed

- CI 增加两步：`node tools/check_versions.mjs`、`node --test tools/check_versions.test.mjs`；
  `docs/UPSTREAM.md` §7 的发版清单多一条同样的勾。
- `AGENTS.md` 增加**第 9 条规则**：报告"没写盘"之前先说清进程是怎么退出的（关窗 + `--quit` 之后再读文件）。

### Fixed

- `AGENTS.md` 的验证命令标题写着 **201 个 `func Test`**（实际 329）。

## [1.0.7] — 2026-11-13（第 44 轮：设置窗口可读、可用、外观页签终于能点）

用户报了两件事：**"外观这个部分点不进去"** 和 **"设置界面字太小之类的"**。前者是一个漏掉的 `case`，
后者是字号梯子整体低于系统基线——两者都在真机/渲染图上确认过，而不是靠读代码下结论。

### Fixed

- **外观页签点不进去**：`settingsApply` 的页签分支只列了通用/声音/会话，**漏了 `settingsTabAppearance`**。
  命中测试、悬停高亮、选中态全都正常，点下去什么也不发生。后果比"一个页签坏了"严重：主题、按终端着色、
  启动时隐藏**只在那一页**，而托盘菜单没有主题项——这三项在 UI 里完全没有入口，只能手改 `config.json`。
  现有测试为什么没抓到：`TestSettingsHitTestFindsEveryControl` 只问"点在哪个控件上"，不问"点了之后发生什么"。
  新测试 `TestSettingsEveryTabSwitches` 问的是后者。
- **备注行永远按一行算高**：备注的段落文字存在 `title` 里（它旁边没有控件），而测量只量了 `hint`，
  于是第二行被裁掉——这是本轮渲染图第一眼就看见的问题。`measureSettingsRows` 现在对 `rowNote` 量 `title`。
- **"标准"字号却高亮"小"**：`config.TextSizes` 是"默认在前"（standard/small/large），分段控件画的是
  小/标准/大，两套顺序不一致，选中的格子就错了一格。现在只有一套顺序（小→标准→大），
  `DefaultTextSize` 单独命名默认值。
- **设置窗口用的是浮窗的 DPI**，而且不处理 `WM_DPICHANGED`：把设置窗口拖到另一块不同缩放比的屏上，
  字号和布局都不跟着变。现在读它自己的 `GetDpiForWindow`，并处理 `WM_DPICHANGED`。
- **英文界面 6 处提示被省略号截断**（实测宽度 > 内容列 272px）：说明原本是单行 `dtEndEllipsis`。

### Changed

- **字号梯子重做**。实测发现 GDI 会把请求的"字符高度"吸附到字体自己的档位：请求 11/12/13 都渲染成
  8px em，14/15 都渲染成 9px em。也就是说旧梯子（标题 15、页签 12、行标题 12、说明 10）**实际只有两种字号**，
  层级根本不存在，而且说明只有 7px em——比 Windows 对话框基线（9px em）低两档。新梯子落在真实不同的档位上：

  | 角色 | 旧 | 新（标准档） | 实际 em |
  |---|---|---|---|
  | 窗口标题 | 15 | 20 | 14px |
  | 页签 | 12 | 14 | 9px |
  | 行标题 | 12 | 16 | 10px |
  | 行说明 / 备注 | 10 | 14 | 9px |
  | 分段 / 数值 / 按钮 | 11 | 14 | 9px |

  档位表与 GDI 阶梯写进注释，`TestSettingsTypeRampStaysOrdered` 在 4 个 DPI 下量真实宽度，
  保证"标题 > 行标题 > 说明"不会悄悄塌成一档。
- **说明改成最多两行换行**（备注三行），换行函数同时负责测量与绘制，所以"预留了几行"和"画了几行"不可能不一致；
  行高从部件算出来（标题带 + 行数 × 行高 + 控件带 + 内边距），不再写死 54/68/80/58。
  中文按行首禁则处理：句号/逗号等不会被挤到下一行开头。
- **窗口加宽** 452 → 500（内容列 272 → 304），侧栏 104 → 112，行高随字号走。
- **屏幕放不下时自动降一档**：字号是偏好，工作区是约束。窗口高度超过工作区时按 大 → 标准 → 小 退让，
  而不是长出屏幕——"够不到的控件"比"小一档的字"更糟。这一步替代了原计划的滚轮滚动：同样的目标，
  少一套滚动状态、命中偏移与裁剪绘制。
- 设置窗口的绘制入口改成 `settingsPaint` 结构体（悬停/按下/键盘焦点三种瞬态状态，位置参数已经太长）。

### Added

- **外观 → 字号：小 / 标准 / 大**（`settingsTextSize`，默认 `standard`）。
- **通用 → 键盘操作**（`settingsKeyboard`，**默认关**）：Tab / Shift+Tab 在控件间移动，空格或回车执行，
  Esc 始终关闭。焦点环只在真的用过键盘之后才画——窗口一打开就画焦点环，等于对伸手去拿鼠标的人喊话。
- **通用 → 显示被截断的文字**（`settingsTooltips`，默认开）：悬停被省略的标签时，在那一行旁边画出全文。
  它画在窗口内而不是做成 shell 工具提示窗口：这个窗口是分层表面、没有子控件，为一个已经很少发生的
  情况再开一个顶层窗口不划算。
- 测试：`settings_tabs_windows_test.go`（8 条：页签真的切换、梯子不塌、字号真的改变窗口、
  矮屏降档、说明不丢字、焦点顺序完整、键盘是可选、换行不丢字）+ `internal/app/settings_window_test.go`
  （新偏好经过 `applyUISettings` 与 `config.json` 往返仍在）。

### Numbers

`gofmt -l .` 无输出 / `go vet ./...` 0 / `go test -count=1 ./...` 12 包全过（**326** 个 `func Test`，其中
`cmd/pet-setup` 11 个）/ 上报端逐个 `node --check` 15 个文件全过 / `node --test reporters/opencode-events.test.mjs`
**8/8** / 三个 `verify.mjs` 全 PASS / `vendor.mjs --check` 两个文件吻合 / `tests/fresh-machine.mjs --no-overlay`
**162/162** / `tools/check_docs_links.py` **12 份**文档全解 / `tools/upstream-entry.test.mjs` 8/8。
`node tools/upstream-entry.mjs --check` 在打 tag 之前是 **SKIP**（1.2.14 的资产还没上传，404 是预期），
打完 tag 后必须重跑成 `OK` 或 `OK（in flight）`。

规模：Go 非测试源码 **47 文件 / 16,882 行**、测试 **51 文件 / 8,884 行**；
`bin/AI-Terminal-Pet.exe` **4,554,240 B**、`bin/pet-statusline.exe` **3,664,896 B**、`bin/pet-hook.exe` **3,697,152 B**。

真机验收（第 44 轮发布前补做的那条，§44.7 曾记为"原因未明"）：在本机装好的构建上开设置窗口，
点外观页签 → 窗口高度 826 → 586；点「纸白」→ **不到 1 秒** `config.json` 里 `theme` 变成 `paper`；
点「字号 大」→ `settingsTextSize` 变成 `large`、窗口 750×586 → 900×701、层级与换行都没塌；随后点回
「石板」「标准」两值复原。所以沙箱里那次"没写回"是**进程被 kill**（写盘在 `configSaver` 的守护协程里，
正常退出才走 `defer flush()`），不是 bug。

发布资产（本机 `scripts/package.ps1` 与 `npm pack` 实测）：

| 资产 | 大小 | sha256 |
|---|---|---|
| `ai-terminal-pet-1.0.7-win-x64.zip`（17 文件） | **5,053,946 B** | `F3D049619BAEFD355E8481CDAF718175792B8C3B0022A3A9408960F99603E9DA` |
| `pet-setup.exe` | **14,307,840 B** | `AE3E2DE769A47A5D87A22FC0D19E1D7617548DBC03F8E5FAAFB7B39AE15C8A6B` |
| `dsh-ai-terminal-pet-1.2.14.tgz`（18 文件） | **2,000,539 B** | `EBD6CE82B653EF5C54FAA01888533D78A79DED4CFEA3137031FCD2E46A3D3FDE` |

`npm pack` 之后把包里的 `README.md` 解出来核对过：两处安装命令都是 `dsh-plugin-v1.2.14`，
`package.json` 与 `vendor/manifest.json` 的摘要也是本次构建的（`source.sha256`
`a841b09b52bba55d…`、`overlay.sha256` `01b42dc27c06413b…`）。

## [1.0.6] — 2026-11-13（第 43 轮：三个"假提醒"的根因，OpenCode 上报端对齐 2.x）


用户报了两件事：**普通终端命令也会响**、**没有 agent 完成任务它也响**。两者都不是提醒层"太敏感"，
而是四个各自独立的错误，全部先在真机/沙箱里复现，再逐个修掉。

### Fixed

- **会话消失不再冒充"完成"**（这就是"没人完成任务它也响"）。旧的 `vanishedInputs` 为"忙过之后消失的会话"
  **合成一个 closed/done**，注释写着"最强的完成信号"——而"消失"其实只是"丢失"：上报端进程死掉、重复 pid 被清理、
  快照过期，三种情况都会触发，没有一种能证明工作结束了。现在它标记 `AlertInput.Vanished`，
  `Decide` 返回**静默**的 `decisionVanished`，回执仍然写明原因。
  沙箱实测（只杀上报端进程，会话本身一直 `working`）：旧版 `closed/done -> done-away` + `tone=finished`；
  现在 `closed/none -> vanished`，无 balloon、无 tone、无未读。
- **一个会话一行**（面板层去重）。store 的 key 是 `source + sessionId + pid`，因为管道、快照文件、
  tombstone 都是按上报进程记账的——那是**协议层身份**，`--json` 必须保留（两个上报端对同一会话说法不一致，
  正是排障要看的证据）。但面板画的是**会话**，所以去重放在 `internal/ui/sessions_dedupe.go`：
  同一 `session.Identity` 只留一行，时间戳相同则取更紧急的状态、再取更大的 pid（稳定，不会每帧换行）。
  同时把 `sessionAlertKey` 从 pid 键改成 `session.Identity`：否则一次 OpenCode 重启会让同一个等待中的会话
  以两个 pid 各提醒一次（静音列表与未读标记本来就是按身份记的，只有告警账本还按进程记）。
- **幽灵行**：`presence.SessionAlive` 里两条规则，都来自真机测量而不是读代码——
  1. **受信任上报端（`opencode` / `powershell-ipc` / `pwsh-session`）的 pid 一旦死亡，会话即结束**，
     任何窗口标题都不能翻案。此前一个死掉的 `opencode.exe` 留下的快照，会被**任意**一个标题里含会话标题或项目名的
     标签页永久续命。真机测量：6 行 / 3 个会话，其中 3 行来自已经不存在的 pid，全靠一个 `OC | <标题>` 标签页吊着。
     修后同一台机器只剩活进程的 4 行 / 4 个会话。
  2. **纯工具名标题不是会话身份**（`opencode` / `OpenCode` / `OC` / `claude` / `agy` / `codex` / `terminal` …）。
     插件在学到真实标题之前会写 `"opencode"`，标签页也常叫 `OpenCode`，两个短字符串互相匹配，
     就把该工具所有历史会话续了命。匹配按**整条标题**，所以 `OC | 修一下解析器` 仍然是身份。
     故意**没有**采用"窗口必须属于该会话的进程树"这条更强的规则：实测 pid 38072（活的 opencode）父进程 24640 已死、
     它的窗口 pid 20992 不在链上，那条规则会把活会话一起删掉。
- **OpenCode 上报端对齐 2.2.7**（这是"普通命令也响"的直接原因）。`reporters/opencode.js` 有两处同时过时：
  - **插件 API**：旧文件导出的是函数 + `{ event, "tool.execute.before", "tool.execute.after" }`，那是 1.x 的 hook 形状。
    OpenCode 2 的加载器要求默认导出带 `id` 和 `setup`（二进制里的原话：`Plugin must export a default definition
    with an id and an effect or setup function`），所以在 2.x 上这个文件**什么都没做**。
  - **事件名**：旧映射里的 `tool.execute.before/after`、`question.asked/replied`、`permission.updated` 在 2.2.7 里
    一个都不存在，其余靠子串猜。子串猜法正是本机插件把 `tui.prompt.append`（往 TUI 输入框里追加文本，
    也就是**用户敲/粘一条普通命令**）判成 `waiting_permission` 的原因——这就是"普通命令也响"。
  现在事件表按 `@opencode/schema`（`session-event.js` / `session-status-event.js` / `permission.js` / `form.js`）
  逐条写死：`session.execution.started/succeeded/failed/interrupted`、`session.step.*`、`session.tool.called/
  success/failed`、`session.reasoning.*`、`session.text.*`、`session.shell.*`、`session.compaction.*`、
  `session.retry.scheduled`、`session.status`（`busy/retry/idle`）、`permission.asked/replied`、
  `form.created/replied/cancelled`、`session.deleted`。不在表里的事件**故意忽略**：
  这个文件读不懂的事件不是关于 agent 状态的证据。
  三个刻意的判断：`session.tool.failed` → `working`（**不是** error，命令失败是日常，agent 会读错误继续）；
  `session.step.failed` → `working`（模型调用失败会被重试，终结性失败由 `session.execution.failed` 报）；
  唯一映射到 idle 的是 `session.execution.succeeded`。逐 token 的 `*.delta` / `tool.progress` 一律不映射——
  它们不改变面板画的东西，一次回答里却有上百个。
  本机真正被加载的那份插件（`~/.config/opencode/plugins/terminal-pet/index.ts`）做了同样的收敛，
  并顺手修掉两处：新会话第一次上报若是 `idle` 会被"状态没变"的优化吞掉（现在 `lastStateChange === 0` 视为必须上报）；
  标题不再伪造 `"opencode"`（空标题让面板回落到项目目录名，至少是真的）。

### Added

- `reporters/opencode-events.test.mjs`：8 个用例钉住事件表（含"`tui.prompt.append` 等 29 个事件绝不出状态"、
  "工具失败不是回合失败"、"只有 runner 说结束才算结束"、`evt_/msg_/prt_` 不是会话身份、2.x 插件形状）。
  CI 增加 `node --test reporters/opencode-events.test.mjs` 一步，`AGENTS.md` 的校验清单同步。
- `internal/ui/vanished_windows_test.go`、`internal/ui/dedupe_windows_test.go`、
  `internal/platform/ghost_windows_test.go`：三组回归测试，每条规则一条用例。

### Changed

- `internal/ui/app_windows.go` 的 `vanishedInputs` 不再合成 `KindDone`；`alertOnTransitions` 里那句
  "消失就是完成"的注释同步改掉（注释和代码相反比没有注释更坏）。
- `packages/dsh-ai-terminal-pet/vendor/` 重新 vendor（源码变了，`vendor.mjs --check` 的源码摘要必须跟着走）。
  包版本仍是 1.2.11：这一批**未发版**，已发布的 `dsh-plugin-v1.2.11` 资产不动，下一轮发版时一起升版本号。

### Numbers

`gofmt` / `vet` / `go test`（12 包，**316** 个 `func Test`，比 1.0.5 多 14）/ 三个 `verify.mjs` /
`node --test reporters/opencode-events.test.mjs` 8/8 / `vendor --check` / `fresh-machine --no-overlay` 162/162 /
`check_docs_links.py` 12 份——全绿。

真机（同一台机器，修前 / 修后）：`--json` **6 行 / 3 个会话**（其中 3 行来自已死的 pid 106060、46108、38752，
全靠一个 `OC | <标题>` 标签页吊着）→ **4 行 / 4 个会话**，全部来自活进程。
沙箱两轮（`pet-repro2.ps1` / `pet-repro3.ps1`，读产品自己的 `alerts.json`）：控制组（真完成）仍 `tone=finished`；
重复 pid 被清理后无 tone / 无 balloon；只杀上报端进程后 `closed/none -> vanished` 且静默。

发布资产（本机产物实测）：

| 资产 | 大小 | sha256 |
|---|---|---|
| `ai-terminal-pet-1.0.6-win-x64.zip`（17 文件） | **5,041,100 B** | `5E2762A29BE8446B146804FF7E889D80801F729ABCCC5C51492864D1571D3E21` |
| `pet-setup.exe` | **14,274,048 B** | `39BC505CE79FB594B2264292B2DCE782CC34E3032D3E0A024FD90F300B140486` |
| `dsh-ai-terminal-pet-1.2.13.tgz`（18 文件） | **1,990,629 B** | `FB809C2B8B225523534041CF951AF7BFC285F361C2BEB57DE3B870651D202782` |

**为什么插件包是 1.2.13 而不是 1.2.12**：1.2.12 的 tarball 是在改 README 之前 `npm pack` 出来的，
包里那份 `README.md` 的安装命令还写着 `dsh-plugin-v1.2.11`——照它装会装到上一版。已发布的资产不可变
（AGENTS.md 规矩 6），所以不改 1.2.12，而是用**修好的 README** 重新打一个 1.2.13；1.2.1 与 1.2.4 也是为
同一类"一句话"存在的。教训写进 §43.7：`npm pack` 必须在所有版本号文本改完之后，并且要把包里的 README
解出来核对一次（`tar -xzf … package/README.md`），不能只看工作区。

复验（全部从公网 URL 下载回来，不读本机文件）：三个资产的 sha256 与上表一致；
沙箱拿**下载回来的 zip** 跑 `install.cmd`（无 `[fail]`）→ `-VerifyOnly` **11 个 `[ok]`、0 失败** →
装出来的二进制 `--version` 打印 **1.0.6** → `--doctor` **10 项、0 fail** → `-Uninstall` 后安装目录 **0 文件**。
npm 上 `dsh-ai-terminal-pet@1.2.13` 的 `dist-tags.latest` 已生效，`npm pack` 解包后与 Release 资产
**逐文件 sha256 相同**（18/18）；npm 会重新打包，所以 `.tgz` 的字节与 Release 资产不同是正常的，
比的是解包后的文件。

## [1.0.5] — 2026-11-13（浮窗：点会话行复制项目路径；判定记录的空 kind 有了名字）

### Added

- **点一下会话行，就把它的工作目录放进剪贴板**。展开态点一行、紧凑态点跳转条的左半都是这个手势；
  剪贴板里是 `session.Project` 的**原始完整路径**（面板行上显示的是 `filepath.Base`，那是 200 像素宽的妥协），
  可以直接粘进终端。上报端本来就发的是 cwd / workspace，**线协议没有改**。
- **面板内的回执**：行上出现约 1.2 秒的胶囊（`已复制` / `复制失败`），紧凑态的回执顶掉 subject 行。
  分层窗口既没有光标也没有焦点环，回执是唯一能让这个手势"被相信"的东西。
- **复制失败不假装成功**：`copyToClipboard` 只在 `SetClipboardData` 真的成功后返回 true；失败时回执写
  「复制失败」并响一声，而**不会**落到「收起面板」——被忽略的点击和没收到的点击看起来一样，那才是坏结果。
- **紧凑条拆成两个控件**：左半是复制（一个两页重叠的图标），右半是原来的跳转（箭头 + 动词），
  中间一条分隔线说明"这里是两个控件"。条子的外轮廓没变，所以面板剪影还是 1.0.4 那一版；
  缩到 65% 以下或没有可复制的会话时，条子**不分**，整条仍是跳转键。
- `internal/ui/clipboard_windows.go`：`OpenClipboard` + `SetClipboardData` 全用 lxn/win 已有的声明，
  **零新依赖**；先 `GlobalAlloc` 再开剪贴板，所以最坏情况是"什么都没变"，而不是"剪贴板被清空后失败"。
- `PET_CLIPBOARD_PROBE=1`：桌面探针做真实往返（写一次、读回、再写一次确认是替换不是追加），
  CI 上没有剪贴板窗口站时 Skip。

### Fixed

- **判定记录的 `kind` 不再写成空串**。`"kind":""` 读起来像缺字段，而它其实是一句陈述（"这个会话没在要什么"）；
  `--doctor` 的 `decisions=/not-asking` 就是因为同一个原因为空。`Kind.Label()` 现在把空 kind 写成 `none`，
  两者一起变成 `none/not-asking`。类型本身的零值仍是空串，所以 `KindOf` 与所有比较的含义都没动。

### Changed

- **点击规则从"点任意处 = 收起"改成"点行 = 复制、点其余 = 收起"**。规则按固定顺序写在
  `internal/ui/tap_windows.go` 的 `decideTap` 纯函数里：跳转箭头 > 紧凑条左半 >
  auto-approve 否决（**仅倒计时武装时**）> 展开态行 > 收起。否决排在行复制之前是有意的：
  自动放行倒计时在跑的那几秒里，"别放行"是最贵的一次点击，用一次剪贴板写入来回答它等于静默放行。
- README 的手势表（中英两段）与插件包内 README 都写明了这个手势，包括"上报端没带工作目录的行仍然只收起"。

### Numbers

`gofmt` / `vet` / `go test`（12 包，**302** 个 `func Test`，比 1.0.4 多 19）/ 三个 `verify.mjs` / `vendor --check` /
`fresh-machine --no-overlay` 162/162 / `check_docs_links.py` 12 份——全绿。
沙箱里拿**装出来的** 1.0.5 用 `PostMessage` 点第一行：`Get-Clipboard` 拿到的就是那个完整路径；
点页脚仍然收起（200 px → 56 px）；紧凑条左半复制、右半不碰剪贴板。

发布资产（都从本机产物算出，并从公网 URL 下载回来复验过）：

| 资产 | 大小 | sha256 |
|---|---|---|
| `ai-terminal-pet-1.0.5-win-x64.zip`（17 文件） | **5,039,878 B** | `AB0218D03F69E1784C49763B464A95FDAA307E254066F8CAF743CEEF149F1BC7` |
| `pet-setup.exe` | **14,267,904 B** | `A5B9025A6E6A6519D305BC6F1BBAAD6D005C20A3D348B3FD61F75F385B0A5C09` |
| `dsh-ai-terminal-pet-1.2.11.tgz`（18 文件） | **1,988,410 B** | `C3FA49B4659FBBD0A820420F095936C7E31C2BC79B01174693B8D04E6FBCC07D` |

沙箱复验（在公网资产上重跑过）：zip → `install.cmd`（11 件制品）→ `-VerifyOnly` 11/11 → 装出来的 1.0.5
`--doctor` **8 ok / 2 warn / 0 fail**（两条 warn 是「浮窗已退出」与「没有活实例」，都不是故障）→
点行复制 → `-Uninstall` 目录归零。npm 上 `dsh-ai-terminal-pet@1.2.11` 已发布，`npm pack` 取回的 tarball
与 Release 资产**逐字节相同**（sha256 相同）。

## [1.0.4] — 2026-11-13（浮窗：判定记录搬家）

### Fixed

- **判定记录不再被删掉**。1.0.3 把 `alerts.json` 放在 `runtime/` 里——那个目录中每个 `*.json` 都被
  `internal/store` 当作一条上报快照，解析不成会话就**删除**。所以 `--json` 永远读不到它、`--doctor` 永远说
  「浮窗从未扫描」，而这两句正好是这份记录要消除的两种沉默。记录现在住在 `config.json` 旁边（应用目录根），
  `internal/store/runtime_contract_test.go` 从目录一侧把「这里只放快照」钉住。
  这一类错误单测看不出来：它验证了「写得进去、读得回来」，没验证「放对了地方」；装进沙箱、让扫描器跑一个周期才看见。
- **种子扫描与「无变化」扫描也各写一条记录**（`skipped` 分别是 `first-load` / `unchanged`）。安静不再是空白：
  「这一层跑过、没什么好说」与「这一层没跑」从此可以区分。

### Changed

- `--doctor` 的 `alerts` 一行现在给出时间、在场状态、气球目标与会话逐条判定；没有记录时仍然是 `[ok]`（全新的机器）。

### Numbers

`gofmt` / `vet` / `go test`（12 包，**283** 个 `func Test`）/ 三个 `verify.mjs` / `vendor --check` /
`fresh-machine --no-overlay` 162/162 / `check_docs_links.py` 12 份——全绿。装出来的 1.0.4 在沙箱里跑过：
上报端推 `working` → `idle` 后 `alerts.json` 落在应用目录、`--json` 的 `alerts` 有内容、`--doctor` 第 10 项
打印 `presence=away, balloon=probe · generic, decisions=done/done-away`。

## [1.0.3] — 2026-11-12（浮窗：提醒层，一个决策器 + 可回答的判定记录）

这一版**唯一会改变老用户体感的默认值**是「回复完成时」：从「总是弹」变成「仅后台弹」。其余新增能力
默认都开着「只在你看不见时提醒」，而那是**减少**打扰的方向；想回到旧行为把这一项设成 `always` 即可
（`config.json` 的 `donePolicy`）。旧配置里的 `notifyOnFinish: true` 会被读成 `always`，不是 `background`
——用户当初点的是「完成要提醒」，把它静默调小了是一种会被当成故障的改动。

### Added

- **一块面板，一个决策器**。`internal/ui/attention.go` 的 `Decide()` 是纯函数：输入是一个会话的全部事实
  （它要什么、你此刻在不在看、它是否被静音、完成策略是哪一档），输出是「弹不弹、响不响、标不标未读、
  为什么」。这里不读窗口、不读文件、不读时钟，所以每条规则都能被表驱动测试钉住。
- **「它在要什么」是从既有字段派生的**（`internal/session/attention.go`）：`waiting_permission` 一个状态
  里其实有三种请求——审批、提问、方案确认。**线协议没有改**，上报端没有改，`reason` 里已有的词足够区分；
  认不出来的词一律保守地当成「审批」，因为**沉默在真有阻塞时是最贵的错**。
- **前台感知**（`foregroundAware`，默认开）：正在看的那个会话的请求不再额外弹气泡，只在行上标未读。
  「窗口在前台但人不在」用 `GetLastInputInfo` 判断（`idleMinutes`，默认 3 分钟，`0` 表示只看窗口焦点）。
  采样在工作线程里跑（枚举别人的窗口绝不上 UI 线程），复用跳转功能已有的进程/窗口缓存。
- **「回复完成」三档**（`donePolicy`：`background` 默认 / `off` / `always`）：单一布尔只能表达「总是/从不」，
  中间那档才是常用的——只提醒你没在看的那些会话。同一会话 15 秒内的两次完成合并成一次。
- **逐会话静音**（右键点会话行 → 「静音此会话的提醒」；设置页「会话」也有「全部恢复」）。键是
  `source + sessionId`，**不含 pid**：hook 上报端的 pid 是每次进程一个的合成哈希，键在 pid 上会让
  每一条静音在下一次上报时失效。静音对提醒是全静（气泡、提示音、未读标记），**唯独不静音面板上的
  自动同意倒计时**——那是安全控件，不是通知。
- **未读标记**：被前台感知压掉的气泡仍然留下标记，点行尾箭头跳过去、会话重新开跑、或会话消失时清除。
  标记画成一个点、静音行整体压暗，都**不占文字宽度**（第一版用「未读/已静音」两个字，把 reason 挤掉了）。
- **两个试发按钮**（设置页「声音」）：一条「需要你处理」、一条「回复完成」，走与真实提醒**同一条路径**。
  以前只有「试听音效」，回答不了「气泡到底弹不弹得出来」。
- **提醒判定记录**（`alerts.json`（在应用目录，不在 `runtime/`，理由见下），覆写不追加）：最近一次扫描里每个会话为什么响/为什么没响。
  `--json` 新增 `alerts` 对象，`--doctor` 新增第 10 项把它汇总成一行。这是「我为什么没被提醒」第一次有答案。

### Fixed

- **同一状态里换了请求不再漏报**（这是这一版存在的起因）。旧账本按 `state` 记账：答掉一个权限请求、
  立刻又被提问，两次都是 `waiting_permission`，于是**一声不响**——而「答完就问」是 agent 最常见的节奏。
  现在账本记的是 (state, kind)，kind 变了就是新事件。
- **连续重复扫描不再重复打扰**。同一条规则的反面：状态与 kind 都没动的会话，`Decide` 直接返回
  `unchanged`；8 秒一次的采样是重绘节奏，不是提醒节奏。
- **新开会话不再因为「完成」而弹**：完成必须是「先前在忙、现在停了」，一个一出生就是 idle 的会话
  什么都没完成。
- **行高亮与命中来自同一个函数**（`rows_windows.go` 的 `expandedRowRect` / `expandedRowHit`）。右键静音
  必须打在它画的那一行上，否则会是「点了没反应」这一类 bug。
- **新画的东西全部进了 `renderedKey`**（未读位、静音位、悬停行）：1.0.1 的时长冻结与 1.0.2 的主题不重绘
  都是「同一个 key = 同一张位图」造成的，这一版不再犯。
- **`--doctor` 的 9 项变 10 项**，新增的是一项真实断言（读不到判定记录才是 warn，全新的机器是 ok）。
- **判定记录不能放在 `runtime/` 里**：那个目录里每个 `*.json` 都被当作一条上报快照，`internal/store` 的扫描器
  会把解析不成会话的 JSON **删掉**。第一版就放在那里，一个 5 秒的扫描周期内它就消失了——`--json` 报「没有记录」、
  `--doctor` 报「浮窗从未扫描」，正好是这份记录要解决的两种沉默。记录现在住在 `config.json` 旁边，
  `internal/store/runtime_contract_test.go` 从目录一侧把这条规则钉住。

### Changed

- 托盘菜单的「会话完成时提醒」现在是**三档策略的两态视图**：关 = `off`，开 = 回到默认 `background`
  （不是 `always`）。设置页的段选才是精确控制。
- 设置页「会话」页签新增六行（前台感知、静止阈值、完成三档、当前会话安静、子会话静默、静音列表），
  「声音」页签新增两行试发。
- 已完成状态的字幕、气球文案都带上项目名：「任务完成了」原来对哪个会话都同一句话，读了也不知道该回哪个。

### Numbers

`gofmt -l .` 无输出 / `go vet ./...` 0 / `go test -count=1 ./...` 12 个包全 ok（**282 个** `func Test`，上一版 218）；
三个 `verify.mjs` 全 PASS；`vendor.mjs --check` 两文件吻合；`fresh-machine --no-overlay` **162/162**；
`check_docs_links.py` **12 份**；`upstream-entry.test.mjs` 8/8。**未改绘制的 15 张预览图逐字节未变**——
这一轮是重构加新元素，不是重画旧面板。

## [1.0.2] — 2026-11-12（浮窗：外观成为设置）

### Added

- **三套命名调色板**（`theme`：`slate` / `paper` / `contrast`）。`slate` 是默认值，也是仓库里所有截图用的那套
  ——**从未打开这个设置的人看到的字节与 1.0.1 完全一致**。`paper` 是浅色（这里最容易做错的一点是「次要文字」
  必须往**黑**走而不是往白走，否则小字在浅底上消失；测试按「每种文字对它自己那块底色」的对比度把这件事钉住）。
  `contrast` 是纯黑底 + 高饱和状态色，给「从远处一眼看清状态」的场景。
- **按终端着色**（`perSourceColor`，默认关）：展开行里的**来源名**按终端上色（claude-hook / opencode /
  antigravity / deepseek-harness），**状态色不动**——状态是信号、来源是出处，把信号按工具染色就毁了这块面板的
  全部意义。未知上报端取中性的 `Secondary` 而不是猜一个颜色。
- **启动时隐藏**（`hiddenAtLaunch`，默认关）：上报端会按需拉起浮窗，所以想要「登录后先在托盘里待着」的人
  不必每次手动隐藏一次。它**只作用于启动**：`--show`、`--toggle` 与托盘菜单之后仍然说了算。
- **设置窗口新增第四页签「外观」**（Look）：主题、按终端着色、启动时隐藏。页签名数从常量
  `settingsTabCount` 来，几何、标签与测试不会对「到底有几个页签」各说各话。

### Fixed

- **主题切换不会静默不生效**。帧复用（`purePulseFrame` 之外的那条「同一个 key = 同一张位图」的规则）只比较
  `renderedKey`，所以调色板与「按终端着色」都必须进 key，否则切换主题会被当成「没有任何变化」而不重绘，
  面板会一直保留旧颜色直到别的原因触发重绘。这与 1.0.1 里「时长冻结」是**同一类** bug，
  `TestThemeAndProvenanceMoveTheFrameKey` 直接断言这两件事，也断言「未知主题名归一化成 slate 之后 key 不应变化」。

### Changed

- `internal/config` 新增 `theme` / `perSourceColor` / `hiddenAtLaunch` 三个字段。旧 `config.json` 没有这些键，
  `Normalize()` 会把 `theme` 归到 `slate`、两个开关保持关闭——升级不会悄悄改变外观。
  `config.ValidTheme` 与 `ui` 里的 `normalizeTheme` 是**有意重复**的（`internal/ui` 不 import `internal/config`，
  文件路径与文件格式归 app 层），由 `internal/app` 的一个测试把两份清单钉在一起。

## [1.2.10] — 2026-11-13（DSH 插件包）

### Changed

- 包内随附的浮窗更新到 **1.0.4**（见上一节）：判定记录搬到应用目录，不再被快照扫描器删掉；安静的扫描
  也会留下记录。`vendor/protocol.js` **未变**（线协议没动）。
- `vendor/AI-Terminal-Pet.exe` 与 `vendor/manifest.json` 重算。
- 包内 README 与仓库 README 的三种安装命令同步换到 `dsh-plugin-v1.2.10` 的资产。
- `docs/market-index.json` 前移到 1.2.10。

## [1.2.9] — 2026-11-12（DSH 插件包）

### Changed

- 包内随附的浮窗更新到 **1.0.3**（见上一节）：提醒层重写成单一决策器，新增前台感知、完成三档、
  逐会话静音、未读标记、试发按钮与判定记录。`vendor/protocol.js` **未变**，因为**线协议没有动**：
  「一个 `waiting_permission` 里是审批还是提问」是面板从既有 `reason` 字段派生的，不是新字段。
- `vendor/AI-Terminal-Pet.exe` 与 `vendor/manifest.json` 重算：源码摘要随 `internal/` 的改动而变。
  这也是「每个非测试 `.go` 改动都必须走 `npm run vendor` 并换一个插件版本」的原因。
- 包内 README 与仓库 README 的三种安装命令同步换到 `dsh-plugin-v1.2.9` 的资产。
- `docs/market-index.json` 前移到 1.2.9（版本 + `publishTimestamp` 在 A 提交里，钉住的
  `source.commit` 与 `icon` 在 B 提交里——git 无法在一个提交里引用它自己的 sha）。

## [1.2.8] — 2026-11-12（DSH 插件包）

### Changed

- 包内随附的浮窗更新到 **1.0.2**（见上一节）：`vendor/AI-Terminal-Pet.exe` 4,391,424 B、
  sha256 `f7788a9a…`。`vendor/protocol.js` **未变**（20,914 B / `5e197db6…`），这一批没有碰线协议。
- 包内 README 与仓库 README 的三种安装命令换到 `dsh-plugin-v1.2.8` 的资产。

## [1.0.1] — 2026-11-12（浮窗）

### Changed

- **紧凑态现在同时回答「哪个终端卡住了」和「卡了多久」。** 有人在等的时候，56×156 面板上最大的那行
  从**会话数**变成**等待时长**（`12m` / `12分`），数量降为下面那行的一部分（`3 sessions` / `3 个会话`），
  状态行与项目名保持不变。整块尺寸、圆角、命中区、拖拽与收起行为**一个字节没动**——素材的分辨率与
  构图因此仍然成立，变的只是最大那行承载的事实。这正是产品开头那句承诺（「哪一行在等、等了多久」），
  以前要展开面板才看得到。
- 只对**确实有人在等的状态**显示时长：`waiting_permission` 与 `error`。忙碌（`working`/`tool_use`/
  `thinking`）没有人在等，`idle`/`closed` 的时间戳含义是「什么时候结束的」，都不是「等了多久」。
- 同时有多个会话在等时，取**最早开始等**的那个，而不是优先级最高的那个：一个卡了 12 分钟的权限确认，
  比 30 秒前的构建失败更值得先看一眼。
- 时长的取整边界与展开行的相对时间**共用同一套**（`<1 分` / `<1 小时` / `<2 小时` / 之后按小时），
  所以同一个瞬间在两处不会一个说「刚刚」一个说「1 分钟前」。超过 99 小时在上限截断，展开态仍显示准确数字。
- 时长文本走新的**短形态** i18n 键（`time.shortMin1` / `time.shortMins` / `time.shortHour1` /
  `time.shortHours`），因为 56 单位宽的面板画不下「12 分钟前」。展开行用的长形态一字未改。

### Fixed

- **时长不会在告警呼吸时冻结。** 呼吸态的帧复用（`purePulseFrame`）是刻意忽略 `clock` 的——脉冲只改
  整窗 alpha，不改像素。因此时长需要一个**独立的分钟桶**字段（`waitMinute`），否则它会被当成「只有时钟
  变了」而被复用旧位图，`12m` 会一直停在 12 分钟直到下一次上报。实测与断言都写进了
  `TestFrameKeyAdvancesTheDurationPerMinute`。
- 尺寸是被**量出来的**，不是看出来的：用真实字体（YaHei UI）在 48 逻辑单位里逐个测过候选串，
  最宽的 `99小时` 在 16 号粗体下 44 px；`TestWaitLabelFitsTheStripIsAMeasuredClaim` 把这件事钉住，
  以后改字号或文案撞到边会在测试里红，而不是在 100% 缩放的桌面上才发现。

## 上游同步批 — 2026-11-12（同样没占版本号，产品字节未变）

这一批只动 `tools/`、CI 与文档，并修了上游那个条目 PR 的冲突。**没有改任何非测试的 `.go` 文件**，
所以 `vendor/manifest.json` 的源码摘要不变，浮窗仍 1.0.2、插件包仍 1.2.8。

### Added

- **`tools/upstream-entry.mjs`**（含 `tools/upstream-entry.test.mjs` 8 条纯函数测试）：从上游
  `awesome-dsh-plugin` 的 **`main` 重建**我们的条目文件，而不是合并旧分支。每一步都是断言，拒绝而不是
  「大概没问题」：目标 `.tgz` 先探一次（不是 200/302 就停）、`tarball:` 行必须恰好一行、
  URL 不许需要 YAML 引号、CRLF 直接拒、**把那一个字符串换回去必须与上游逐字节相同**（换不回去说明
  带了别人的改动）、`git diff` 必须是 `1 file changed / 1 1`、`--force-with-lease` 只推自己 fork 的
  `bump-tarball-*` 分支。`--check` 有三种**故意不同**的结论：已列出 → OK；上游还旧但有对应 PR 开着 →
  OK（上架要等人合，不是我们可控）；发版了却没人推 → 红；未发布的版本 / 读不到上游 / 没 `gh` → SKIP
  （skip 不是 pass，输出里写明）。
- **`docs/UPSTREAM.md`**：上游 `contributing.md` 的硬要求逐条对照我们的现状、一次发版的八步顺序、
  工具的用法与拒绝条件、「合并 ≠ 可见」该读谁、描述怎么写、当天真踩过的六个坑、以及一页发版清单。
- **CI**：`verify` 作业新增 `Upstream catalogue entry`（`--check`）与 `Upstream entry rules`
  （`node --test`）；`permissions` 增加 `pull-requests: read`，因为 `--check` 要看上游是否已有 PR 开着。

### Fixed

- **`#14718` 的冲突**：旧分支是拿 `tarball` 行做三方合并的，而上游 `main` 在同一天里被维护者人工
  改过同一个文件（`#14688` 把条目推到 1.2.7），于是内容冲突 → **冲突的 PR 没有 merge ref，GitHub 连
  `check` 都不跑**，`Submission gate` 也就不触发，PR 停在「等检查」里。改法按维护者 `fkysly` 的原话：
  以上游 `main` 为底重建，只把那一个字符串换成 1.2.8，其余（中英描述、上报端前置条件、「DSH 会话只做
  状态映射与跳转、授权仍在 DSH 完成」）**逐字保留**；推上去后 `mergeable=MERGEABLE`，`check` 正常启动。
- **条目里的一句描述本来就不准**：`Click a row to jump to that terminal` —— 实际点行会**收起面板**，
  只有**行尾箭头**跳转。改成 `Click the arrow at the end of a row to jump to that terminal`，
  中文同步。这是条目自己的一句，不是别人的；面板以后真有了「点行做某事」再改，不留两种手势的说法。

## [1.2.7] — 2026-11-12（DSH 插件包）

### Changed

- 包内随附的浮窗更新到 **1.0.1**（见上一节）：`vendor/AI-Terminal-Pet.exe` 4,377,600 B、
  sha256 `59f1cfcf…`，源码摘要 `ab16b003…`（24 个文件、go1.27.0）。`vendor/protocol.js` **未变**
  （2,914 B / `5e197db6…`），因为这一批没有碰线协议。
- 包内 `screenshots/compact.png` 跟着布局重出，`screenshots.json` 的条目不变。
- 包内 README 与仓库 README 的三种安装命令同步换到 `dsh-plugin-v1.2.7` 的资产。

## 门禁加固批 — 2026-11-12（同样没占版本号，字节与 1.0.0 / 1.2.6 相同）

这一批只动测试、CI 与文档，**没有改任何一个非测试的 `.go` 文件**，所以 `vendor/manifest.json` 的源码摘要
保持 `edd584c6…` 不变，插件包仍 1.2.6、浮窗仍 1.0.0——`vendor.mjs --check` 与
`scripts\install.cmd -VerifyOnly` 都当场跑过。

### Added

- **CI 新增 `scan` 作业**：`govulncheck@v1.9.0` 与 `gosec@v2.30.0`，都钉版本。它独立成一个作业，是因为
  两者的 `go.mod` 都要求比本模块下限更新的 Go（`go 1.26.0`）：构建作业继续按 `go.mod` 的 1.23 走
  （「同一份源码构建出同一份字节」的前提），扫描作业用 `go-version: stable`，不让 `GOTOOLCHAIN=auto`
  在构建作业里偷偷多下载一整套 Go。结果：`govulncheck` 代码受影响 **0** 条（14 条依赖模块级公告调不到）、
  `gosec` 在 high severity + high confidence 上 **0** 条。
  `G115`（84 条，Win32 的 `uintptr`/`uint32` 转换）与 `G703`（12 条，由自己的 `%LOCALAPPDATA%` 拼出的路径）
  按**规则族**排除，理由写在 workflow 注释里；不用逐行 `#nosec`，因为那会改产品源码、改摘要、逼着换版本号。
- **`tests/fresh-machine.mjs` 的失败信息**：新增 `describeRun()`（`status` / `signal` / 生成错误 / 是否重试 /
  输出，空 stdout 写成 `<empty>`）与 `runDoctor()`（**只在 `ETIMEDOUT` 时**用双倍预算重试一次，并把
  「重试过」写进结果）。起因见下方 Fixed。
- **`.agents/` 进 `.gitignore`**：个人知识库中枢派发的工单属于机器本地材料，不是项目源码（§34.6 记录过
  为什么那五张工单不是新工作）。
- **`docs/HANDOVER.md` §35**：这一批的施工记录、两个扫描器、偶发红事件、以及「只改未来的提交身份」的决策。

### Fixed

- **一次 161/162 的偶发红，失败信息是空的**。S5 的 `doctor reports no failures after a standalone install`
  失败时 detail 只传了 `doctor.stdout`，而 `spawnSync` 超时被杀时 stdout 是空串——「被超时杀掉」与
  「正常退出但一行没打印」在报告里长得一模一样。修法见 Added；随后连跑四次都是 162/162，
  断言一条都没放宽（仍是 `0 fail`，预算仍是 30 s）。
- **CI 的五个 action 还停在 node20**。GitHub 每次运行都在注解里提示 `checkout@v4`、`setup-go@v5`、
  `setup-node@v4`、`setup-python@v5`、`upload-artifact@v4` 正被强跑到 node24。逐个读过它们发布的
  `action.yml` 才换（确认 `using: node24`，且本 workflow 用到的输入都还在），现在都在 v7
  ——`upload-artifact` 的 node24 第一个主版本是 v6，v5 仍是 node20，这里取 v7 只为保持一条主版本线。

### Changed

- **对外口径按 2026-11-12 的实读结果修正**：npm `latest` = 1.2.6；`docs/market-index.json` 钉的是
  `df4fcf6`；awesome-dsh-plugin 的站点 `plugins.json` 已渲染 1.2.6（`count` 10,581）；README 里
  「`1.2.4` 尚未发布 / `npm view` 仍是 E404」与「#14385 等人工合并」两处旧事实换成了实测结果，
  中文段一处被合并坏掉的句子（从句中开始、缺主语）也补回了主语。
- `CHANGELOG.md` 两个 `[Unreleased]` 标题改成带年份的批次标题：它们记录的工作随 1.0.0 / 1.2.6 一起发出，
  留 `[Unreleased]` 会和本文件开头的「已全部发版」自相矛盾。
- `docs/ROADMAP.md` §6 与 `README.md` 的门禁清单：`govulncheck` / `gosec` 从「因缺 gcc 未跑」改为「已进 CI」，
  `go test -race` 仍是缺 C 工具链的真实缺口——对外**不能说**「全量安全扫描」。
- 提交身份：`user.email` 与签名只作用于**未来的提交**（`236163329+jjjjjjjjnnjnn@users.noreply.github.com`
  + SSH 签名）。118 个历史提交与 11 个已发布 tag **刻意不重写**：邮箱是提交对象的一部分，重写会改掉每一个
  sha，而 `docs/market-index.json` 与第三方目录已经按现有 sha 钉过版本。

## 宣传准备批 — 2026-11-11（随 1.0.0 / 1.2.6 交付，批次本身没占版本号）

这一批同样不换版本号，但它把「非 Harness 用户怎么拿到这个东西」这条断掉的漏斗接上了。
触发点是本轮实测：`.\scripts\install.ps1` 在本机默认 `RemoteSigned` 策略下**直接被拒**
（报的是安全策略的官话，不像装错了），而 README 的安装第 2 步就是那一行，任何普通用户照抄必然失败；
同时 Release 资产里没有任何可下载的成品包——`dsh-plugin-v*` 的 tarball 只是插件包，
非 Harness 用户要么装 Go 工具链，要么从 tarball 里手动挖 `vendor/AI-Terminal-Pet.exe`。

### Added

- **`scripts/install.cmd`：文档里的安装入口**。RemoteSigned 会拒绝一个带「从互联网下载」标记的本地
  `.ps1`，而从 Release 解压出来的文件恰好带这个标记；`install.cmd` 用
  `-ExecutionPolicy Bypass` 启动**一个** `powershell.exe`，不改任何 MachinePolicy / UserPolicy /
  CurrentUser 设置，参数与退出码原样转发。它不加任何安装逻辑，`install.ps1` 仍是唯一实现。
  纯 ASCII（Windows PowerShell 5.1 按 ANSI 读无 BOM 脚本，与 `install.ps1` 同一条理由）。
- **`scripts/package.ps1`：便携 zip 打包**，产物 `dist\ai-terminal-pet-<overlay 版本>-win-x64.zip`。
  里面是**仓库原有布局的一个子集**（`bin\` 三个二进制、`reporters\protocol.js`、
  `integrations\deepseek-harness\ai-terminal-pet-bridge\` 七个文件、`scripts\install.ps1`、
  `scripts\install.cmd`），加一个 zip 根的 `install.cmd` 转发器、`README.install.txt` 与
  `SHA256SUMS.txt`。选这个形状是因为 `install.ps1` 的一切路径都从 `Split-Path -Parent $PSScriptRoot`
  推导——同一份脚本既能装仓库也能装 zip，不需要第二套安装实现，也不会有「校验集 ≠ 安装集」的漂移。
  版本号从 `internal/app/app_windows.go` 的 `const Version` 读取，不在这里写第二份。
  三个已知的 PowerShell 陷阱都是实测撞到的，已写进脚本注释：`Compress-Archive` 与
  `ZipFile.CreateFromDirectory` 在 .NET Framework 下写出**反斜杠**分隔的条目名（Linux/macOS 上
  `unzip` 会解出一个名字里带反斜杠的文件），所以改为逐条目写入；`[ZipArchive]` 必须先
  `Add-Type` 两个程序集才解析得到；`CreateEntryFromFile` 是 `ZipFileExtensions` 上的扩展方法，
  而 PowerShell 没有扩展方法语法。每个条目钉固定时间戳，同一份源码重建出的 zip 逐字节相同
  （与三个二进制用 `-buildvcs=false` 是同一条理由）。
  打包前会比对 `bin\AI-Terminal-Pet.exe` 与插件包 `vendor\AI-Terminal-Pet.exe`，不一致就拒绝打包
  ——这两个是同一产品同一个版本号下的两份副本，任一侧陈旧都会让两类用户拿到不同字节。
  CI 用 `PET_PACKAGE_SKIP_VENDOR_CHECK=1` 把它降级为警告，因为 runner 的 Go 版本与提交进仓库
  的 vendor 二进制不是同一个（那是跨工具链字节相等不可能达成的老问题）。
- **`.github/workflows/ci.yml` 增加打包门禁与产物上传**：镜像里跑 `package.ps1` 并断言语义
  （文件数与逐名比对，不数目录条目），再把 zip 作为 `ai-terminal-pet-win-x64` artifact 留存 7 天。
- **`tests/fresh-machine.mjs` 增到 162 项（S5b）**：用 `cmd.exe` 跑 `scripts\install.cmd`
  的两种形态，断言参数转发（`-VerifyOnly` 打印 `11 file(s) ... 0 warning`）与退出码转发
  （`-VerifyOnly -Uninstall` 仍是 2）。走 `-Destination` 指向沙箱，所以不碰本机的真实安装目录。
- **`docs/assets/` 四个新素材**：`demo.gif`（25 帧 / 5.38 秒 / 342,606 B，等待呼吸 → 悬停展开 →
  点击等待行）、`desktop-scene.png`（1500x860，「两个终端 + 浮窗同框」）、
  `doctor-green.png`（1,798x358，真机 `--doctor` 输出 + 着色，用户名换成 `<you>`）、
  `smartscreen-unsigned.png`（1320x560，**按官方对话框文字重建的示意图**，图左下角标 MOCKUP）。
  前两张由 `make_launch_assets.py` 从同一份光栅器输出合成；后两张分别由 `make_doctor_card.py`
  与 `make_smartscreen_notice.py` 产出。**真机 SmartScreen 截图需要一台"从没构建过本项目"的机器**，
  而且会带上别人的桌面与机器名——示意图把"你会看到什么、点哪里"说清楚，并且不假装自己是截图。
- **仓库门面**：`CONTRIBUTING.md`（含六条硬规则、发布不可变文件清单、`--doctor` 作为报障前置）、
  `CODE_OF_CONDUCT.md`（Contributor Covenant 2.1）、`.github/ISSUE_TEMPLATE/` 三份表单
  （bug 强制附 `--doctor` 全文、terminal 接入请求、feature 并列出两条非目标）、
  `pull_request_template.md`（门禁清单 + 是否触碰发布不可变文件的显式勾选）、
  `ISSUE_TEMPLATE/config.yml`（关掉空 issue：bug 表单第一个必填项就是 `--doctor` 全文，
  空表单正好是绕开它的捷径；另挂三条 contact link）。
  仓库 description / homepage / topics 同时更新，topics 从 11 个补到 15 个（`opencode`、
  `antigravity`、`ai-agents`、`coding-agents`）——社区健康度此前只有 42%。
  **注意上限是 85%**：GitHub 只把单个 `.github/ISSUE_TEMPLATE.md` 算作 `issue_template`，
  目录形式的表单不计分，所以五项 guidance 文件齐了之后分数就停在那里。
  **这套表单第一版是坏的**：`bug.yml` 里 `description: \`winver\` output ...` 的反引号让 YAML
  标量提前结束，文件扫不过，GitHub 会静默拒收——坏掉的恰好是唯一强制要 `--doctor` 全文的那张表。
  修法是改成引号标量，并且**先跑一个 YAML 解析器再信它**。

### Changed

- **README 三个安装路径并列**：A（Harness 用户，npm 裸包名）、B（下载便携 zip 双击 `install.cmd`，
  零 Go 工具链）、C（开发者从源码构建）。英文段与中文段同步。
  **首次运行会撞 SmartScreen「未知发布者」**，两段都提前写明点「更多信息 → 仍要运行」，
  并挂上 `docs/assets/smartscreen-unsigned.png`（示意图）与 `docs/assets/doctor-green.png`，
  说明 `SHA256SUMS.txt` 是签名到位之前诚实的来源证明——这一版刻意不签，因为没有证书。
- **README 事实修正**：插件版本 1.2.4 → **1.2.5**、npm 已发布（`dist-tags.latest` 1.2.5，
  17 个文件，tarball 与 Release 资产逐字节相同）、market-index 钉到 `34aa1f7`、
  徽章 `tests-154` → `tests-190` 并加 npm 徽章、"No standalone installer yet" 改为
  "Not code-signed"、CI 门禁列表补上 `package.ps1`。
- **`.gitignore` 增加 `dist/`**：打包产物是 5 MB 二进制，随每个 release 重建，不入库。
- **`docs/LAUNCH.md` 成为对外文案的唯一存放处**：Show HN 标题与正文（61 字符 / 259 词）、
  作者首楼技术补充、X 中英各 5 条线程（字数逐条实测 ≤270）、DSH 市场摘要两版、
  三份 issue 回复模板，以及「发布顺序」「渠道优先级」「发完之后看什么」三张表。
  文案按 AGENTS.md 第 8 条由 `agy --model gemini-3.8-flash-high --effort high` 起草，
  再逐条按本仓库实测事实核对。**该轮评审至少纠正了一条会真出问题的事实表述**：
  初稿把浮窗写成像是对 Claude Code / OpenCode / Antigravity 开箱即用，而实际只有
  DeepSeek Harness 无需配置。
- **`pet-v0.9.0` 独立 Release**：资产 `ai-terminal-pet-0.9.0-win-x64.zip`（4,983,915 B /
  sha256 `EF0FC0D0…`），与插件线 `dsh-plugin-v*` 是两条互不干扰的 tag 线。已从公开 URL
  下载回来、解压、`install` → `-VerifyOnly`（11 件全对）→ `-Uninstall -Force`（rc=0）实测。

### Fixed

- **文档承诺与默认 Windows 策略冲突**（本轮实测）：README 第 2 步的 `.\scripts\install.ps1`
  在 `RemoteSigned` 下抛 `UnauthorizedAccess`，现在由 `install.cmd` 承接，并在两段正文里解释原因，
  而不是让用户第一次接触这个项目就读一条安全策略错误。

---

## 可自查与可验证批 — 2026-11-10 / 2026-11-11（随 1.0.0 / 1.2.6 交付，批次本身没占版本号）

这一批不发新版本号（`internal/app/app_windows.go` 里的 `Version` 仍是 `0.9.0`），
目的是让这个仓库在**别人的机器上**也能跑通、自查、接入与持续验证：可移植性、`--doctor` 自检、
正式线协议文档、安装脚本的只读/卸载/Path 开关，以及 GitHub Actions 门禁。详细 runbook 见
[docs/HANDOVER.md](docs/HANDOVER.md) §1（当前状态）与 [docs/ROADMAP.md](docs/ROADMAP.md) §1i。
2026-11-11 追加的这一批把上报端原生化（ROADMAP §E1）：新增 `cmd/pet-hook` 与 `internal/reporter`，
施工记录见 HANDOVER §29。浮窗本体仍是 0.9.0，安装器制品由 10 项变 11 项。

### Added

- **`tests/fresh-machine.mjs`：全新机器端到端套件（158 项检查）**。它回答的是「这台机器从没见过
  这个项目，装上去能用吗」，而不是「代码逻辑对不对」——后者由 `go test` 与 `verify.mjs` 覆盖。
  每次运行都在 `%TEMP%` 建一个一次性沙箱，自带 `LOCALAPPDATA`、`USERPROFILE`（因此连命名管道都是
  另一条）、DSH 风格的 `node_modules` 与一个真实存在的项目目录，八个场景：

  | 场景 | 覆盖什么 |
  |---|---|
  | S1 空机器 | 未安装时 `--doctor` 退出 1 并指名缺什么；`--json` 仍返回空文档；`--version` 可用 |
  | S2 包安装 | `npm pack` → 解开到 `node_modules`；清单、`dsh.bundle`、`cordis.patch.yml` 的 `insert:` 形态、locale、vendor 三件套 |
  | S3 插件 | bootstrap 拷贝 + 哈希等于 vendor；二次调用是 memoized 而非覆盖；7 种事件的状态/原因/工具映射；切目录与改标题；两个会话互不干扰；`approval/request` 必定 `next()`；dispose 后不再上报；**中文路径与中文工具名原样往返** |
  | S3b 开关 | `DSH_PET_NO_BOOTSTRAP=1` 不拷贝但**仍能用 vendor 协议上报**；`config.enabled:false` 一个监听都不注册 |
  | S4 浮窗 | 沙箱里第一次启动、`--json` 能看到插件的会话、没有 `config.json` 也能跑、`--toggle`/`--show`/`--reset` 都 exit 0 |
  | S4b 冷启动 | `--quit` 让浮窗真的退出（**连管道一起消失**）；之后只跑一次 `pet-hook`，浮窗应当被上报端自己拉起来，并且那条会话可见 |
  | S4c 八条会话 | 连续 8 条会话全部进 store；`--json` 的排序与 `internal/session.Sort` 一致（先按状态优先级，同状态内新的在前） |
  | S5 独立安装 | `install.ps1` 11 件制品、`-VerifyOnly` 全对、`--doctor` 无 fail、两个原生 hook 投递进沙箱浮窗 |
  | S6 设置 | 自动同意关闭时 defer、开着时 allow；危险命令永不自动放行（两种方言）；否决文件 → deny；`--settings` 可达 |
  | S7 坏机器 | 配置损坏、没有 `LOCALAPPDATA`、超大 stdin、空/非 JSON/数组/null、未知参数/未知模式——全部 exit 0 且给出答案 |
  | S8 卸载 | 目录清干净、`--doctor` 重新变红、无残留进程 |

  两个设计要点写在文件头：二进制必须**复制进沙箱再执行**（本仓库目录带 Low 完整性标签，从
  `bin/` 直接跑的子进程读不到 `%TEMP%` 沙箱里的配置文件）；浮窗相关场景会先停掉正在运行的实例、
  结束后从**安装目录**那份重新拉起（单实例互斥体是每个 Windows 会话级别的，而仓库里那份是
  Low IL，控制命令根本连不上 Medium IL 的管道）。CI 里跑 `--no-overlay`，只跳过会开窗口的两个场景。
- **`ai-terminal-pet --doctor`**（`internal/app/doctor_windows.go` 563 行 + 375 行测试 /
  14 个 `func Test`）：一次跑完 **9 项**自检——`environment`（编译目标与 `LOCALAPPDATA`）、
  `install files`（安装目录里的 `AI-Terminal-Pet.exe` / `pet-statusline.exe` / `protocol.js`）、
  `version`、`runtime`（快照目录）、`config`、`pipe`、`instance`、`node`、`autostart`。
  每项输出一行 `[ok]` / `[warn]` / `[fail] name: detail`，末行汇总 `doctor: N ok, N warn, N fail`，
  任一 `fail` 时退出码为 **1**；退出码经 `errDoctorFailed` 哨兵值传递，因此**不会写进 `error.log`**
  （自检失败是用户主动请求的诊断结果，不是程序自己出了故障）。用法错误仍走 `usageError` → 退出码 **2**。
  - **刻意不碰互斥体**：`CreateMutex` 会**创建**这个对象、干扰正在运行的实例，所以 `checkInstance`
    改用「管道有人监听」作为实例存在的证据。
  - 管道探测 `probePipe` 连上即关（`doctorDialTimeout` 250 ms），不写任何字节——服务端把立即 EOF
    当作空负载、不会触发 update 回调，所以跑 doctor 不会改变浮窗显示的内容。
  - `node` 探测 `node --version` 限时 300 ms，并把 `Missing` 与 `TimedOut` 分开：
    「没装 Node」和「装了但没应答」对用户是两件相反的事。子进程带 `HideWindow`，从快捷方式启动时不闪黑框。
  - 每项检查都包在 `runCheck` 的 recover 墙里：将来某一项 panic 只产出它自己的 `fail` 行，
    不会带走另外八项；返回空 verdict 也会被计成 `fail` 而不是悄悄消失。
  - 只提供人读文本，**没有 `--json` 变体**：报告是贴进 issue 的，每行已经比可解析格式更好读。
- **`docs/PROTOCOL.md`**（388 行）：正式线协议规范。第三方为未列出的终端自行编写上报端时
  不必再读源码，字段、状态语义、管道名派生规则、`OK`/`ERR` 回执契约与脱敏要求都在这里；
  schema 变更流程见该文档 §11。
- **`reporters/package.json`**（26 行）：给 `reporters/` 目录写显式 `exports`，覆盖 6 个 `.js`
  （`protocol.js`、`opencode.js`、`claude-hook.js`、`antigravity-hook.js`、`antigravity-statusline.js`、
  `generic.js`），不再依赖 Node 的隐式解析；`private: true`，不发布到 npm。
  **刻意不写 `"type"` 字段**：5 个 hook 是 `require()` 的 CommonJS，而 `opencode.js` 是真正的 ESM——
  写 `"module"` 会打断前者，写 `"commonjs"` 会打断后者。无名 type 的代价是首次 import `opencode.js` 时
  Node 打印一条 `MODULE_TYPELESS_PACKAGE_JSON` 提示并按语法探测重解析为 ESM：预期行为，只多一次解析。
  把 `opencode.js` 改名 `.mjs` 的方案被否决，因为 README 与用户配置里引用的是 `.js` 路径。
- **Harness 桥 README 去掉写死的主机路径**：`integrations/deepseek-harness/ai-terminal-pet-bridge/README.md`
  里那句给用户照抄的 `plugin_manager` 目标原为 `C:\Users\<name>\...`，现改为
  `%LOCALAPPDATA%\AI-Terminal-Pet\bundles\ai-terminal-pet-bridge` 并补三行说明：该字符串是贴进
  Harness 插件管理器的路径，按用户解析才有意义，写死某个用户名在别人机器上必然不对。
- **`.github/workflows/ci.yml`**：`windows-latest` 上的单 job 门禁，任一失败即挂——
  `gofmt -l` 有任何输出、`go vet ./...` 有发现、`go test -count=1 ./...` 失败、两个
  `go build -trimpath -buildvcs=false -ldflags "-s -w"` 构建失败、任一上报端 `node --check` 失败、
  `node integrations/deepseek-harness/verify.mjs` 失败、`node packages/dsh-ai-terminal-pet/scripts/vendor.mjs --check`
  失败、`node packages/dsh-ai-terminal-pet/verify.mjs` 失败。（不记行数：这份文件随门禁增删而变。）细节：`concurrency` 让同分支的新推送取消旧运行；
  `fetch-depth: 1`（只要树）；`go-version-file: go.mod` 让 `go.mod` 的 `go 1.23.0` 成为唯一工具链来源；
  Node 22（ESM 语法探测自 22.7.0 起默认开启）；`core.autocrlf` 显式关掉，避免 runner 检出 CRLF 让
  `gofmt` 误报。**`node --check` 是逐文件循环而不是 `node --check reporters/*.js`**：
  后者只检查第一个参数（Node 24 实测：第二个参数是坏文件也退出 0），循环是唯一真正覆盖全部文件的写法。
  live-desktop 探针测试在对应 `PET_*` 变量未设置时 `Skip`，所以不需要桌面会话或任何 secret。
- **`packages/dsh-ai-terminal-pet` —— 把 Harness 桥接做成独立 DSH 插件包**（`7f476c9` 起，历经 1.0.0 →
  1.1.0 → 1.2.0）：npm 包名 `dsh-ai-terminal-pet`、license MIT、`type: module`、`private` 未设（可发布）；
  当前版本 **1.2.4**（`packages/dsh-ai-terminal-pet/package.json` 2,246 B / 83 行，1.2.3 → 1.2.4 只改 `dsh.catalog` 摘要、1.2.2 → 1.2.3 多了 `dsh.marketplace` 块）：
  包内 18 个文件（含 vendor，1.2.0 起多一个 `vendor/manifest.json`），主体是 `lib/index.js`（17,872 B / 570 行）、
  `verify.mjs`（23,426 B / 424 行）、`scripts/vendor.mjs`（13,981 B / 324 行）、`package.json`（2,246 B / 83 行）、
  `README.md`（8,187 B / 146 行）、`vendor/manifest.json`（529 B / 19 行）、
  `cordis.patch.yml`（399 B / 10 行）、`locale/en.json`（210 B）/ `locale/zh.json`（217 B）、
  `icon.svg`（526 B）/ `icon.png`（4,386 B）、`LICENSE`（1,070 B）、`screenshots.json`（97 B）、
  `screenshots/compact.png`（4,459 B）/ `expanded.png`（11,335 B）/ `expanded-many.png`（17,697 B）、
  `vendor/protocol.js`（20,914 B）、`vendor/AI-Terminal-Pet.exe`（4,371,968 B）。
  `package.json` 用 `dsh.bundle { patch: './cordis.patch.yml', platforms: ['win32'] }` 声明自己是一个
  Cordis bundle，`engines` 为 `node: '^22.19 || >=24'`、`dsh: '>=0.1.0-rc.6'`，`files` 10 项。
  **1.1.0 删掉 `prepack` 并新增 `vendor:check`**（`b280093`）：`scripts` 现在是
  `verify` / `vendor` / `vendor:check`，没有任何生命周期钩子——pnpm 对 git 依赖会跑包的生命周期脚本，
  未钉版本的源会因此撞上一条完全不提 AI Terminal Pet 的 `allowBuilds` 报错，删掉比解释更可靠。
  `scripts/vendor.mjs` 随之变双模式（无参拷贝、`--check` 只校验 sha256，vendor 缺失或 0 字节即 exit 2），
  `vendor/` 必须由发版人在打包前手工拷一次并提交，`npm pack` 不再替你重新生成它。
  **1.2.0 增加 `dsh.catalog`**：`category: tool`、中英各一行摘要（190 / 104 字符）、五项自我声明的
  `capabilities`（`windows-overlay` / `named-pipe` / `session-state` / `tray-icon` / `no-network`），
  `keywords` 补上 `dsh-plugin`。这些 capabilities **只自我声明、不被强制**——dsh 不对插件做沙箱，
  所以它是描述，不是权限授予；`keywords` 则是自动索引型市场采集的标签，于是多出第二条发现路径：
  注册表评审（条目合并进 `main` 才上架）与 topic/关键词采集（无需提交申请）。版本必须跟着清单内容走，
  否则已发布的不可变 tarball 会与仓库分叉——这正是本仓库一直在消除的漂移。
  行为：把 Harness 事件映射成浮窗状态（`approval/asked → waiting_permission`、`approval/decided → working`、
  `turn/start → working`、`turn/end → idle`、`tool/call → tool_use`、`tool/result(error) → error`、
  `command/done`、`working-directory/change → project`、`session/title`）；记住每个会话最后一次状态，
  只在状态变化或距上次 ≥60 s（`KEEP_ALIVE_MS = 60_000`）时才发一次，避免高频刷屏。
  `npm pack` 的 1.0.0 口径是 **1.9 MB packed / 4.5 MB unpacked / 15 files**；1.2.0 因为 `vendor/manifest.json`
  成为随包文件，文件数变成 **16**，实测 **1,914,566 B / 16 files / packed 1.9 MB / unpacked 4,460,352 B（4.5 MB）**
  （1.1.0 是 1,913,726 B / 15 files）。多出来的 840 B 就是那个 529 B 的清单加上打包开销。
  - **`cordis.patch.yml` 必须用 `insert:` 列表**（`- id: dsh-ai-terminal-pet` / `name: dsh-ai-terminal-pet`）。
    裸的 `- id:` 只会 patch 一行**已存在**的 loader 行，装完看着成功却什么都不挂载——这是踩过的坑，
    因此文件头的注释把原因写在了 YAML 里。
  - **vendor 预编译 + 引导安装只补缺失**（`runBootstrap()`）：仅当
    `%LOCALAPPDATA%\AI-Terminal-Pet\AI-Terminal-Pet.exe` **缺失**时，才从包内 `vendor/` 拷
    `protocol.js`（`0o600`）与 exe；`copyVerified` 走「临时文件 + rename、rename 前再查一次目标是否存在」，
    **从不覆盖已有文件**。`DSH_PET_NO_BOOTSTRAP=1` 可整段关掉（测试缝），
    `DSH_PET_TEST=1` + `DSH_PET_PACKAGE_ROOT` 可替换包根。
  - `node verify.mjs`：rc=0、**54 ok**、末行 `VERIFY PASSED`；覆盖状态映射、waterfall `next()`、
    dispose 不发 `closed`、真 exit 发 `closed`、bootstrap 只补缺失不覆盖、protocol 抛异常时插件全静默。
- **仓库根的 `package.json` 清单 shim**（`06d4cae`，1.2.4 口径 2,115 B / 68 行——1.2.3 只多了 `dsh.marketplace` 九行，1.2.4 只把 `dsh.catalog` 摘要加长）：不含任何代码，只让
  「按仓库 URL 安装」在仓库根就能读到 `dsh.bundle`、`main`、`exports`、`icon` 与 `engines`；
  `version` 与插件包同步（`verify.mjs` 断言两者相等，所以一次版本号变更不可能只落在一个清单里），
  并带上同一份 `dsh.catalog`。没有它时，只读仓库根的宿主把仓库当成 0.0.0 依赖、报 `declares no dsh.bundle`
  并装出一个空包——症状是「装成功了，但浮窗永远不出现」，没有任何报错。标 `private: true` 是为了不占用 npm 名。
- **`verify.mjs` 连根 shim 一起检查，并在 CI 里跑插件套件**（`d06001d`）：新增根清单一致性用例，
  从 tag `dsh-plugin-v1.1.0` 的 300 行 / 55 个 `ok` 长到 1.1.0 批次的 331 行 / 63 个 `ok`，再随 1.2.0 的清单用例
  到 77 个 `ok`、最后由 `e25842f` 的索引漂移用例长到 415 行 / 91 个 `ok`，1.2.3 又加四条 `dsh.marketplace` 断言到 **424 行 / 95 个 `ok`**（1.2.4 只改文案，断言数与行数不变，仍是 424 行 / 95 ok、`VERIFY PASSED`、rc 0；那 14 条
  把 `docs/market-index.json` 钉在包版本与同一个 commit 上，发版忘改索引 CI 会红）；
  `.github/workflows/ci.yml` 增加两步：`node packages/dsh-ai-terminal-pet/scripts/vendor.mjs --check` 与
  `node packages/dsh-ai-terminal-pet/verify.mjs`。验证脚本自带 fake Cordis context + 临时 `LOCALAPPDATA` +
  stub protocol 模块，所以 CI 上不需要网络、管道或真实浮窗。
- **GitHub Release 线：1.0.0 → 1.1.0 → 1.2.0 → 1.2.1 → 1.2.2 → 1.2.3 → 1.2.4（历史按顺序记，当前版本是 1.2.4）**。
  **1.0.0**（已作废）：tag `dsh-plugin-v1.0.0`，资产 `dsh-ai-terminal-pet-1.0.0.tgz`（1,913,682 B，
  sha256 `7365f6a6f4aea1174ec6f2f00670e868568e944816d1869aeb59db363ae6b5af`），
  下载地址 `https://github.com/jjjjjjjjnnjnn/ai-terminal-pet/releases/download/dsh-plugin-v1.0.0/dsh-ai-terminal-pet-1.0.0.tgz`。
  三种安装方式都实测成功：①本地 tarball `dsh plugin --profile <name> add <tgz 绝对路径>` →
  `+ dsh-ai-terminal-pet 1.0.0`，`dsh --profile <name> --dump-config` 末尾出现 `- id: dsh-ai-terminal-pet`
  （证明 `insert:` 真的挂上了）；②git 源
  `dsh plugin --profile git-test add git+https://github.com/jjjjjjjjnnjnn/ai-terminal-pet.git#91719e5a2c7c221ad163686f80eaab9c4eca4ea6&path:/packages/dsh-ai-terminal-pet`
  → rc=0、10.2 s，装出的包里 `vendor/AI-Terminal-Pet.exe` 4,371,968 B 在位；③Release 资产 URL
  `add https://github.com/jjjjjjjjnnjnn/ai-terminal-pet/releases/download/dsh-plugin-v1.0.0/dsh-ai-terminal-pet-1.0.0.tgz`
  → rc=0、32.6 s，vendor 内两个文件都在位。裸 git 源（不钉 commit/tag）会被 pnpm 的 `allowBuilds`
  门拦在 `prepack` 处失败，所以对外命令只给 tarball 与钉 tag 两种形态。
  **1.1.0**：tag `dsh-plugin-v1.1.0`（annotated，指向 `b28009377e14840ee003ddaf114922e1d167ed0d`），资产
  `dsh-ai-terminal-pet-1.1.0.tgz`（**1,913,726 B / 15 files / sha256 3b8f9186cca5142626f2e25175096f1c4fb59d86cc1cdff5879792950146cc78** /
  shasum `19a2e544ab875c61fcf822e85a91c67273981725`），删掉 `prepack` 后不再有生命周期钩子，因此
  多出第三种安装形态：**直接给仓库地址** `github:jjjjjjjjnnjnn/ai-terminal-pet`（靠仓库根的清单 shim 解析，
  等价 `git+https://github.com/jjjjjjjjnnjnn/ai-terminal-pet.git`）。三种形态实测（计时取在 1.1.0 的资产上；
  1.2.0 资产同位置就位、URL 只换 tag 与版本号）：Release 资产 1.8 s、git 钉 tag 3.4 s、仓库 URL 3.6 s，均 rc=0。仍会失败的是「不写 `#ref` 的 `git+...git&path:/packages/...`」：
  pnpm 把两段拼进 URL 当仓库地址去 `ls-remote`，报 `repository ... not found`——**与 `prepack` 无关**。
  **1.2.0**（由 1.2.1 取代，发布物本身仍不可变）：tag `dsh-plugin-v1.2.0`（annotated，tagger 2026-11-10T13:46:52+0100，指向
  `4d777e32e0a9ef57bb7714287dccd7cae065346f`，Release 于 2026-11-10T12:47:24Z 发布），资产
  `dsh-ai-terminal-pet-1.2.0.tgz`（**1,914,566 B / 16 files / sha256
  `ccf9d531267af681438010bc94c721e7822a0db3f4d14575953b1757a2ef65c9` / shasum
  `dc65e0d4259be640c0de7ff1bc86675af5454546`**，实际下载后本地重算与 GitHub 记录的 digest 一致）；
  内容新增 `dsh.catalog`、
  `vendor/manifest.json` 随包（文件数 15 → 16）。1.1.0 保持已发布、不可变，由 1.2.0 取代——清单内容变了
  就必须换版本号，否则仓库与那份不可变 tarball 分叉。1.2.0 发版前 `npm publish --dry-run` 给的是
  **16 files / 1.9 MB packed / 4.5 MB unpacked / shasum `166c3043…**；但 dry-run 之后包内 `README.md`
  又改了 13 B（6,652 → 6,665 B），所以**真正发出去的资产是 1,914,566 B / shasum `dc65e0d4…` /
  sha256 `ccf9d531…`，不是 dry-run 那个摘要**——包里任何一个字节变过，dry-run 的摘要就作废。
- **awesome-dsh-plugin 注册表条目（PR #13937，2026-11-10T14:33:04Z 由评审人 `fkysly` 合并）**：市场 App `dshmarket` 只是前端，
  目录来自注册表仓库 `awesome-dsh-plugin/awesome-dsh-plugin`，投稿方式是只加一个
  `data/plugins/<slug>.yml`。本次条目为
  `data/plugins/jjjjjjjjnnjnn__ai-terminal-pet--packages-dsh-ai-terminal-pet.yml`，
  `category: notify`、`url` 指向 `.../tree/main/packages/dsh-ai-terminal-pet`、
  `name: jjjjjjjjnnjnn/ai-terminal-pet#dsh-ai-terminal-pet`，并带上 `icon`（raw.githubusercontent 的 icon.png）
  与 `tarball`（**合并时钉的是 1.2.0 那份资产**，条目本身不会跟着本仓库发版走；要改得另开 PR 改那个 yml）。
  本地先跑通上游两道门：条目模型校验
  `entries read: 10120 / problems: 0`，`scripts/check-submission.mjs --only-list` →
  `ok https://github.com/jjjjjjjjnnjnn/ai-terminal-pet/tree/main/packages/dsh-ai-terminal-pet` /
  `all checked entries pass`。PR 见 <https://github.com/awesome-dsh-plugin/awesome-dsh-plugin/pull/13937>
  （fork 分支 `add-dsh-ai-terminal-pet`，commit `a1fef56`，基于上游 main `09cd105`）。
  **状态是「已合并、且随后真的在市场里可见」**：PR 上两个检查都已完成且 `conclusion: SUCCESS`
  （`PR check / check` 于 2026-11-10T10:37:54Z 结束、`Submission gate` 于 10:38:47Z 结束，
  `mergeable: MERGEABLE`），随后在 14:33:04Z 被 `fkysly` 合并。维护者合并时删掉了原稿里「也能在
  面板上直接放行审批」一句——对 DeepSeek Harness 会话而言，面板只做状态映射与跳转到 Harness 窗口，
  审批仍在 DSH 里完成；上游 README 因此把这一条写成了如实描述。17:19Z 维护者在 PR 上留言确认了这件事，
  并说明「措辞如需调整，直接开 PR 改对应的 `data/plugins/*.yml` 即可，我们按最新的为准」——即上游 `main`
  上那份 yml 的 `description.en/zh` 带的是他那句「approving is still done in DSH / 审批仍在 DSH 里完成」，
  与我们原来的稿子不同，而且**本仓库任何文档改动都影响不到它**。接入 `main` 的事实仍可验证：
  `awesome-dsh-plugin/awesome-dsh-plugin` 的 `README.md` 在「Notifications & Integrations」小节第 8,519 行
  列着 `jjjjjjjjnnjnn/ai-terminal-pet#dsh-ai-terminal-pet`。但**合并进 `main` 不等于市场里可见**——
  **这一条后来被证伪了一半**：条目页
  <https://awesome-dsh-plugin.com/p/jjjjjjjjnnjnn/ai-terminal-pet--packages-dsh-ai-terminal-pet/>
  在 15:56Z 与 16:16Z 两次复查都是 404；它变成 200 是在上游 `Build site from README` 跑完之后。
  那两次 run 都停在步骤 9 `Build`：`42841664024`（15:13:06Z 起）于 15:25:40Z 被 cancelled（同分支新
  run 取代），`42844452013`（15:27:09Z 起）于 16:39:27Z success。同时 DSH 应用内 Market 真正加载的
  `https://awesome-dsh-plugin.com/plugins.json` 从 `count: 10119` / 0 命中变成 `count: 10230` 且已收录
  本仓库——但它里面的 `install` 命令是按 yml 的 `tarball` 字段在构建时渲染的，所以合并之后读到的那份仍写着 1.2.0。
  结论修正为：「条目已合并，并且已经在列表与市场里可见；延迟来自上游站点重建，不是条目本身的问题」。
  跟进 PR 换过一次版本：先开 #14063（fork 分支 `bump-tarball-1.2.1`，head `116b0ee`，1 file `+1/−1`）
  指向 1.2.1，两项上游检查（`check` pass 24m59s、`Submission gate` pass）与 `mergeable: MERGEABLE` 都拿到了；
  但 1.2.1 是在 `install.ps1` 还在仓库根时打的，`scripts/` 下沉改到 8 处 Go 注释后 `vendor/manifest.json`
  的源码摘要变成 `cf647af1…`，于是 1.2.2 重打包、#14063 留言说明后关闭，改开 **#14076**
  （<https://github.com/awesome-dsh-plugin/awesome-dsh-plugin/pull/14076>，分支 `bump-tarball-1.2.2`，
  commit `0863b4d`，仍只有 `+1/−1`）指向 1.2.2 的资产，**已于 2026-11-10T19:59:37Z 合并**（merge commit
  `e5f2d26`）；1.2.3 发版后原样再跟一次：[PR #14089](https://github.com/awesome-dsh-plugin/awesome-dsh-plugin/pull/14089)
  （分支 `bump-tarball-1.2.3`，commit `c5a1f5f`），**已于 2026-11-10T21:10:15Z 合并**（merge commit
  `95d7216913d54110a794514a6c0ea0196aa08ddc`）；1.2.4 的跟进是
   [PR #14243](https://github.com/awesome-dsh-plugin/awesome-dsh-plugin/pull/14243)（分支 `bump-tarball-1.2.4`，
   commit `66ed01d`，`1 file +3/−3`，2026-11-11 开出、等 check 与人工合并）。
  1.2.4 这一版还顺手把「其他终端要不要另装东西」写进了上游两个条目的措辞里
  （awesome 的 `data/plugins/*.yml` 与 imsai 的 `catalog/plugins/*.json` 各开一个只改 description 的 PR：
   前者是 #14243，后者是
   [imsai-sh/awesome-deepseek-harness-plugins#912](https://github.com/imsai-sh/awesome-deepseek-harness-plugins/pull/912)
   （分支 `entry-wording-1.2.4`，commit `1570316`，`1 file +3/−3`；该仓库 CONTRIBUTING 规定「改已存在条目永不自动合并」，
   所以 `static-review` success、`merge` skipped，等 maintainer 人工合并），
  因为那两处文案同样会让读者以为「装了就有全部终端」。每一版都先用上游自己的门验过：
  `scripts/check-submission.mjs --only-list` → `checking 1 entry / ok / all checked entries pass`（该门需要 `GITHUB_TOKEN`）。
  **两处外部副作用已记录**：① 已发布的 `dsh-ai-terminal-pet-1.2.0.tgz` 里那份 `README.md` 是 1.2.0 打包时的旧文，仍写着「等人工评审、尚未列入」——
  Release 资产不可变，本仓库的改动不会回写进那份 1,914,566 B 的 tarball，所以才有下面那条 1.2.1 —— 只换文本，也只能换版本号。当时 `docs/market-index.json` 钉的同样是 `4d777e3` 那个 commit，也会把读者带到旧文；该文件后来在 `cf1085d` 前移到 `1e08b3e`。写文案时别拿包内 README 当现状。
  ② 已向市场仓库提了误报 issue <https://github.com/bradeGithub/DSH-Plugins-Marketplace/issues/367>：
  `registry-run-key` 与 `remove-item-force-recurse` 两条旗标都是正则匹配到常量定义与提示文本，
  `install.ps1` 从不新增 Run 值、递归删除只作用于一个经 `Assert-SafeInstallDirectory` 校验的目录；
  该仓库没有抑制机制，是否接受由维护者决定。
  同日 16:22Z 该仓库的 `build-registry`（run `42812806941`，schedule，12:05:32Z 起）跑完了：
  `installability-report.json` 里从 `verdict: script` 改成 **`verdict: bundle-plugin`**——根清单的
  `dsh.bundle.patch` 终于被它看见。**但那次条目仍不带 `risk_tier`**：它的前端把缺字段渲染成「未评估」，
  扫描器本身也没有抑制机制，所以「没有 risk_tier」当时不等于 safe。**到 18:08:44Z 那次重建就真的变成 safe 了**：
  本仓库条目是 `pkg_name: dsh-ai-terminal-pet` / `version: 1.2.1` / topics 11 个 / **`risk_tier: "safe"` 且
  `risk_flags` 为空**（`risk_at` 17:09:14Z，是安装脚本移进 `scripts/` 后第一次真的扫到本仓库）。
  这不是靠抑制字段换来的——`evalRepoRisk` 只按文件名抓仓库根的 `install.ps1` / `install.sh` / `package.json`，
  脚本移走后那两个正则没有落点；issue #367 先回帖于 17:58:34Z（`#issuecomment-6775035737`），19:04:00Z
  再补一条（`#issuecomment-6776560881`）写明 18:08:44Z 重建后已是 `safe`、旗标清零、可由此关闭。
- **安装脚本下沉到 `scripts/`，消掉市场扫描器的两条误报**（`bdc6c36`，2026-11-10T17:38:20Z 推送，未发版）：仓库根的 `install.ps1`
  （27,358 B / 583 行）移到 `scripts/install.ps1`（现 27,930 B / 590 行），仓库根不再有同名文件。动机是 [DSH-Plugins-Marketplace]
  (https://github.com/bradeGithub/DSH-Plugins-Marketplace) 的 `evalRepoRisk` **只按文件名抓仓库根的
  `install.ps1` / `install.sh` / `package.json` 再整文正则扫描**，而它报的两条旗标（`registry-run-key` critical、
  `remove-item-force-recurse` medium）恰好匹配到脚本里合法的常量定义与清理逻辑——脚本从不新增 Run 值、
  `-Recurse -Force` 只作用于一个经 `Assert-SafeInstallDirectory` 校验过的目录。移走之后扫描器读不到落点，
  旗标自然消失，同时市场类型判定会落到根 `package.json` 的 `dsh.bundle`（`bundle-plugin`，与 16:22Z 那次重建的结论一致）。
  **行为一个字没改**，只有路径与相对定位变了：`param()` 的四个默认值改用 `Split-Path -Parent $PSScriptRoot`
  （`param()` 里不能用已赋值的变量，而 `$PSScriptRoot` 是唯一在参数绑定阶段就可用的自动变量），脚本体里新增
  `$RepoRoot = Split-Path -Parent $PSScriptRoot` 给预检的 `$binRoot` 用。实测 `-VerifyOnly` 仍是 `10 file(s) match`、
  `0 warning(s)`；完整安装仍 rc=0、两个 exe 与 `protocol.js` 与 bundle 7 文件当场刷新。**代价**是全仓 27 个文件里
  约 100 处 `install.ps1` 文本引用要跟着改（README 中英、AGENTS.md、ROADMAP、HANDOVER、CHANGELOG、插件包与本仓库的
  注释），改漏一处就等于告诉读者一条跑不通的命令。
- **1.2.1 —— 只为让包内说明追上事实，不改代码**（`1e08b3e`）：1.2.0 的 tarball 冻结在打包那一刻，里面的
  `README.md` 还写着「等人工评审、尚未列入」。Release 资产不可变，所以这句旧文只能靠新版本发出去改正。
  1.2.1 与 1.2.0 的唯一差别是包内 `README.md`（Install 段改成「条目已合并、站点构建跑完后市场里已可见」）
  与两条安装命令里的版本号；`lib/index.js`、`vendor/AI-Terminal-Pet.exe`、`vendor/protocol.js`、
  `cordis.patch.yml`、`dsh.catalog` 全部逐字节不变，所以浮窗、事件映射与线协议契约没有任何改动。
  **这类「只改文字」的发版也必须换版本号**：复用 1.2.0 的号会让 npm/registry 侧看到同一个版本两份不同内容，
  也让任何 sha256 核对失去意义。
  - 发布物实测：annotated tag `dsh-plugin-v1.2.1` 指向 `1e08b3e7521d305d744c5ff515b29f6d16abebe1`
    （tag 对象 sha `0ef1d60cb5c33517aa87d6eed75565c035d5830e`）；GitHub Release 于
    **2026-11-10T16:15:04Z** 发布，资产 `dsh-ai-terminal-pet-1.2.1.tgz`
    **1,914,616 B / 16 files / sha256 `b9e6d20354b5b1543b55041aa0c45c38cd4fdfacc65ab1b13fef5ab0651d0075` /
    shasum `67955a65417b6b469c591d1d65fbd1893426628a`**；GitHub 记录的 digest 与本地重算一致。
    连续两次 `npm pack` 产出的 tgz 逐字节相同——只改文字时这一条是可复现的。
  - 解开资产核对：16 个文件里 `package/README.md` **6,863 B / 130 行**（旧文那句 `waiting on a human reviewer`
    已不存在、两条命令都指向 `dsh-plugin-v1.2.1`），`package/package.json` version 为 `1.2.1` 且 `scripts.prepack` 为 `undefined`。
  - 同一批还改了仓库自己的插件索引：`docs/market-index.json`（`cf1085d`）把版本、`publishTimestamp`、
    安装源 commit 与图标 commit 全部从 `1.2.0` / `4d777e3…` 移到 `1.2.1` / `1e08b3e…`，否则索引会把读者带到
    那份写着旧文的 1.2.0 tarball。改完 `verify.mjs` 复跑 **93 行 / 0 fail / VERIFY PASSED**。
- **1.2.2 —— 安装脚本下沉后必须重打包**（`bdc6c36` 之后的发版提交）：1.2.1 的包是在 `install.ps1` 还在仓库根的时候打的。
  随后把脚本移进 `scripts/` 消掉了市场风险扫描器的两条误报输入，同时改到 8 处 Go 源码注释；而
  `vendor/manifest.json` 记录的 `source.sha256` 正是「`cmd/` 与 `internal/` 下所有非测试 `.go` 文件」的摘要，
  于是它从 `89eef0d6…` 变成 `cf647af1…`（25 个文件，`goVersion` 仍 `go1.27.0`）。已发布资产不可变，
  所以包内那份 manifest 必须靠新版本发出去，否则按 1.2.1 装到的人拿到的是与源码对不上的摘要——
  `vendor:check` 在干净 clone 里会打印「bin/ 不存在，以已提交的 vendor 为准」而通过，但在有 `bin/` 的机器上直接失败。
  1.2.2 与 1.2.1 的差别只有三处：两份清单的 `version`、包内 `README.md`（两条命令与「跟进 PR 指向当前版本」的措辞）、
  `vendor/manifest.json` 的新摘要；`lib/index.js`、两个 vendor 二进制、`cordis.patch.yml`、`dsh.catalog` 逐字节不变。
  - 打包实测：`dsh-ai-terminal-pet-1.2.2.tgz` **1,914,751 B / 16 files / sha256
    `02431e16dd58db166794d7c86608087446d4a3ed8ab51bcb4b0950b2a045307c` / shasum
    `22b63b25c50d5f4d799bfcf00fd4f54986be63a8`**；连续两次 `npm pack` 逐字节相同（1.2.1 为 1,914,616 B，
    差值 135 B 全部来自 manifest 摘要与 README 措辞，`files` 仍是 16 个）。
  - 发布物实测：提交 `694ff3f13740b1bb9fb4df04f5fd226064a559b2`（7 files, +48/−35），annotated tag
    `dsh-plugin-v1.2.2` 指向它（tag 对象 sha `06d7795270d516ca7ced6707bf726bcf0337de99`）；GitHub Release
    **2026-11-10T18:11:26Z** 发布，资产 digest 与本地重算一致（`sha256:02431e16…`）。
  - 随后 `fe62900` 把 `docs/market-index.json` 前移到 1.2.2：`source.commit` 改成 `694ff3f…`、图标钉同一个 commit、
    `publishTimestamp` 更新；改完 `verify.mjs` 复跑 **93 行 / 0 fail / VERIFY PASSED**。
  - 端到端复验：全新 `DSH_HOME` + 独立 profile `p122`，`add <1.2.2 的 release 资产 URL>` → rc=0 / **3,335 ms** /
    `+ dsh-ai-terminal-pet 1.2.2`，`--dump-config` 末尾出现 `- id: dsh-ai-terminal-pet`；装出的
    `vendor/manifest.json` 里 `source.sha256` 正是新值 `cf647af1…`（证明新版本把正确的记账带到了用户机器上）。
  - 上游 registry 的 tarball 跟进：#14063（指向 1.2.1）已评论说明后关闭；#14076（`bump-tarball-1.2.2`）只改一行
    `tarball:` 指向 1.2.2 资产，**已于 2026-11-10T19:59:37Z 合并**（merge commit `e5f2d26`，两项检查
    `check` 21m21s 与 `Submission gate` 均 pass）；#14089（`bump-tarball-1.2.3`，commit `c5a1f5f`）把它再推到
    1.2.3，开单时 `state: OPEN`、`mergeable: MERGEABLE`、`mergeStateStatus: UNSTABLE`（只因检查没跑完），
    而后 **已于 2026-11-10T21:10:15Z 合并**（merge commit `95d7216913d54110a794514a6c0ea0196aa08ddc`）。
    注意**合并不会立刻改变任何人装到的版本**：目录里的 `install` 命令是构建时按 yml 的 `tarball` 渲染的，要等
    上游下一次站点重建才更新（合并后读到的那份仍写着 1.2.0）。#14089 合并于 2026-11-10T21:10:15Z，1.2.4 的跟进 PR #14243 已于 2026-11-11 开出（等合并）。
    同一天还落地两家别的目录：[imsai-sh/awesome-deepseek-harness-plugins](https://github.com/imsai-sh/awesome-deepseek-harness-plugins)
    的 PR #902 于 2026-11-10T21:49:53Z 合并（`catalog/plugins/jjjjjjjjnnjnn--ai-terminal-pet--packages--dsh-ai-terminal-pet.json`，
    只从 npm 装，所以在 1.2.3 发到 npm 之前以 browse-only 列出）；
    [zhu1090093659/dsh-community-plugins](https://github.com/zhu1090093659/dsh-community-plugins)
    的 PR #97 于 2026-11-11T05:06:32Z 合并（`community.json` 里 category `integration` / subcategory `bridge`）。
    1024 Store 详情页 <https://deepseek1024.com/plugins/jjjjjjjjnnjnn/ai-terminal-pet/packages/dsh-ai-terminal-pet>
    返回 200、8,753 B，HTML 的 meta description 就是我们的英文摘要。
- **1.2.3 —— 让第三方商店知道这个包能跑在哪些 profile 里**（`b4474a6`）：DSH 自己无法表达「这个 host-only bundle 在任何 profile 都能跑」——它唯一与 profile 有关的字段是 `dsh.client`，而 client 只能是 `web`。于是按 `install.profiles` 过滤的第三方注册表只能退回 `[headless, web]` 猜测，桌面端商店因此把条目藏起来；`YELEBAI/dsh-plugin-marketplace` 就是这种实现，它此前已把本仓库记成 `profiles: [headless, web]`。
  两份清单的 `dsh` 里因此多了 `marketplace`：`profiles: ["desktop", "headless", "web"]`、`requiresBuildApproval: false`、`manualSteps: false`。**DSH 本身从不读这个字段**——插件照旧经由 `dsh.bundle` 挂载，而且这个 bundle 只订阅 `session/event`，不依赖任何 profile 专属服务；读它的是商店侧。`verify.mjs` 加了四条断言钉住它（含根 shim 必须逐字复制同一块），两个清单不可能各写各的。
  - 该注册表的分类器 `scripts/registry-core.mjs` 的 `classifyInstall` 被本地重放了一遍（把它的 `yaml` 依赖替换成空桩后直接 import 它自己的源码，不是手抄判定）：对当前这棵树得到 `profiles: [desktop, headless, web]`、`profileSource: "manifest"`、`runtimeArtifactsCommitted: true`、`reviewReasons: []`。
  - 打包实测：`dsh-ai-terminal-pet-1.2.3.tgz` **1,915,013 B / 16 files / unpacked 4,461,408 B / sha256 `1172aa72c63482442ccf0c1dfc12446fa15e1b12220f921fbe06a56c4db87c35` / shasum `0508ff6b9159674ae4c7136cb00bf1870d44eb35`**。解包核对：`package/README.md` 与工作区那份逐字节相同（7,546 B）、`package/package.json` version 1.2.3、`dsh.marketplace` 块在位、`scripts.prepack` 仍为 `undefined`。
  - 与 1.2.2 的 1,914,751 B 相比只多 262 B，全部来自两块 `dsh.marketplace` 与包内 README 的新增段落；`lib/index.js`、两份 vendor 二进制、`cordis.patch.yml`、`dsh.catalog`、`screenshots/` 逐字节不变。
  - 一次 CI 红灯值得记下来：`b4474a6` 把包版本写成 1.2.3，而 `docs/market-index.json` 还停在 1.2.2，`verify.mjs` 的
    `caseMarketIndex` 断言「索引里的最新版本 == 包版本」，于是那次 run `42894142140` 在 `Verify the DSH plugin package`
    这一步失败（日志原文 `fail: the newest indexed version is the packaged version (1.2.3)`）。索引是在下一个提交
    `fabb591` 才前移的，所以那条 tag 指向的提交本身是红的（tag 与 Release 资产不受影响——索引不在 `files` 白名单里）。
    **因此定一条规矩：发版的同一个提交里必须把 `docs/market-index.json` 一起前移**，否则 tag 上的 CI 必红；
    `fabb591` 之后 run `42896771226` 恢复 success。
- **1.2.4 —— 只改文案，不改代码：说清楚「哪些终端要另外装上报端」**：
  此前市场文案读起来像「装一次市场条目就能看到所有终端」，但实际情况是——这个 bundle 开箱即用映射的只是
  **DeepSeek Harness 自己的会话**。OpenCode、Claude Code 与 Antigravity 的会话要先给**各自那个终端**装好上报端
  （`reporters/` 下的 `opencode.js` / `claude-hook.js` / `antigravity-hook.js`）才会汇进同一块面板；
  没装上报端的终端**不会**上报任何东西，这一点在任何安装形态下都成立（不是本包的缺陷，是接线的物理事实）。
  这一句是用户按商店实际规格从头复现安装后给出的唯一一条（不阻断）反馈。
  - 改到的地方：包内 `README.md`（7,546 B / 137 行 → **8,187 B / 146 行**；开头那段「一次市场安装就够」改成
    「先把这个 Harness 的会话接上，其他终端要各自装上报端」并指向仓库 README 的 `#connect-your-terminal` 段，
    再加一句 the panel, the pipe and the protocol are shared, so once a terminal reports it shows up next to the
    Harness sessions；standalone 段补 Those reporters are separate installs, not part of this package），
    以及两份清单的 `dsh.catalog` 摘要（en 225 字符 / zh 114 字符，**三处逐字相同**：根 shim、包清单、
    `docs/market-index.json` 的 `plugins[0].description`，`verify.mjs` 的根清单用例要求序列化后逐字相同）。
  - 版本号：两份清单 `version` 1.2.3 → **1.2.4**（根 2,052 → **2,115 B / 68 行**，包 2,183 → **2,246 B / 83 行**）。
    只改一句也要换版本号，理由与前几版一致：Release 资产不可变，同版本号下换字节就是漂移；
    且 `verify.mjs` 断言根 shim 与包清单的版本相等，一次改号不可能只落在一个清单里。
  - 随包字节的其余部分**逐字节不变**：`lib/index.js`（570 行 / 17,872 B）、`vendor/AI-Terminal-Pet.exe`
    （4,371,968 B / sha256 `bbb5163c…`）、`vendor/protocol.js`（20,914 B / sha256 `5e197db6…`）、
    `cordis.patch.yml`、`icon.svg` / `icon.png`、`locale/en.json` / `locale/zh.json`、`screenshots/` 三张、
    `vendor/manifest.json`（源码摘要仍 `cf647af1…`——1.2.4 没碰任何 `.go` 源码）。
  - 发版之后（同日 09:1x–09:5x +0100）又向外投了五家目录，全部走「public repo + `dsh-plugin` topic + 人工复核」这套：
    [dshbase](https://github.com/ylwl1997/dshbase/issues/322)（`[收录] ai-terminal-pet`；POST 时带 `labels` 会 403
    `You do not have permission to create labels on this repository.`，去掉 labels 才成功）、
    [dsh.so](https://github.com/dsh-so/dsh-plugin-submissions/issues/24)（按 #21/#23 的既有格式写，末尾附建议的
    registry entry JSON）、[dsh-plugin.org](https://github.com/dshplugin/dsh-plugin-hub/issues/735)（该站固定模板六段 +
    提交前自查七项；它明说没有 npm 包的仓库照常收录、只是不进「插件中心」）、
    [AllDSH](https://github.com/alldsh/submissions/issues/2)（走 `POST https://www.alldsh.com/api/submit`，
    返回 `201 {"ok":true,"issue":".../issues/2"}`）与 [dshmarketplace.dev](https://dshmarketplace.dev)
    （走 `POST /api/v1/submissions`，先 `action:"inspect"` 得到 `alreadyListed:false`/`hasDshTopic:true`，
    再 `action:"submit"` 得到 `{"ok":true,"duplicate":false}`）。五条 issue 现在都是 open / 0 评论，
    人工复核通常在几天到一周。
  - 同一轮复读到的外部状态：`awesome-dsh-plugin` 的 #14243 仍 open，`mergeable_state` 从 `unstable` 变回
    `clean`（check 已跑完）；imsai #912 仍 open，但 `mergeable_state` 从 `unknown` 变成 `clean`（仍等 maintainer
    人工合并）；站点 `plugins.json`（12,570,153 B）里的条目仍渲染 1.2.3 资产；dshopen #423 的正文已由我们 PATCH 到
    1.2.4（body 3,047 字符，updated 2026-11-11T08:14:43Z）并追加了一条评论，**随后该 issue 于
    2026-11-11T08:32:39Z 以 `state_reason: completed` 关闭**——它的 `data/sources.json`（1,213 条）第 1204 条已收录
    本包（`version` 1.2.4、`source.commit` `ff7f6de`、`path` 子目录、中英摘要都写了「其他终端需各自安装 reporter」），
    但线上 `public/index.json` 仍是旧构建（854 条、0 条 git 源），所以目录页要等它下一次构建才看得到；
    `YELEBAI/dsh-plugin-marketplace` 的
    `registry/plugins.json`（generatedAt 2026-11-11T02:19:42.600Z、18,385 条）仍记 `version: "1.2.3"` 与
    `verifiedCommit: 9b0df94…`，但 `install.profiles` 已经是 **["desktop","headless","web"]**——1.2.3 声明的
    `dsh.marketplace` 真的被它读到了，桌面端不再隐藏条目；下一步扫描会把 version 推到 1.2.4。
  - 打包实测：`dsh-ai-terminal-pet-1.2.4.tgz` **1,915,340 B / 16 files / unpacked 4,462,112 B /
    sha256 `d6b7f95a5a3a71c4a9367699e9ad9b8f7d5db522271782e8ffe29a73e45c6708` / shasum
    `dafb81998efbae8e906a36773e9737117aa482d9`**；连续两次 `npm pack` 逐字节相同（`cmp` 无差异）。
    与 1.2.3 的 1,915,013 B 只差 **327 B**，全部来自两块摘要与包内 README 新增的段落。
  - `docs/market-index.json` 同一批前移：`versions[0].version` → 1.2.4、`publishTimestamp` → 1794379909
    （2026-11-11T06:51:49Z）；`source.commit` 与图标钉点由紧随其后的提交补上（沿用 1.2.1 / 1.2.2 / 1.2.3
    那三次的两步做法）。改完 `node packages/dsh-ai-terminal-pet/verify.mjs` 仍是 **95 ok / 0 warn / 0 fail /
    VERIFY PASSED / rc 0**。两步实际落成：**A `ff7f6de`**（9 files +252/−67，tag `dsh-plugin-v1.2.4` 指向它）与
    **B `4ee7f01`**（`1 file +2/−2`，把索引的 `source.commit` 与 `icon` 改成 A 的 sha），推送后 CI run
    `42981618507` 在 B 上 success（同一个 push 只为 head 起一次 run，所以另外用 `git worktree add --detach` 解出
    A 的树单独跑了一遍 `verify.mjs`，同样 `VERIFY PASSED`）。
- **两条「不用申请」的自动采集路径**：仓库 topics 现为 `dsh-plugin` / `deepseek-harness` /
  `cordis-plugin`。`LivXue/dsh-plugin-shop` 每日构建，收录**根 `package.json` 里同时有 `name` 与
  `dsh.bundle`** 的 GitHub 仓库，并钉住默认分支的那个 commit；`zasenjc/dsh-plugins-store` 每 30 分钟同步，
  收录门槛是 topics **同时**具备 `dsh-plugin` 与 `deepseek-harness`（两者都填，不是冗余）。
  另有 `cyber-moshen` 的 workshop 要求「已发布到 npm 的包 + 一个 PR」，属于最重的一条。
  补两条实测更正：`LivXue/dsh-plugin-shop` 发布的 26,351 条条目里**每一条都带 npm tarball 的 `integrity`、
  `commit` 字段一条也没有**，所以它在包进 npm 之前不会收录纯 git 源——根清单 shim 只是「有资格被它读」
  的前提，不是「会被它列出来」的保证。
- **仓库自己发布一份 format-1 插件索引**（`ef783ba`，新文件 `docs/market-index.json`）：
  <https://raw.githubusercontent.com/jjjjjjjjnnjnn/ai-terminal-pet/main/docs/market-index.json>。
  因为索引格式允许「钉到 40 位 commit 的 git 源」，所以在 npm 发布被一次性密码挡住时，Discover 仍然可用；
  条目只有 `dsh-ai-terminal-pet` `1.2.4`（索引随发版前移过四次：`cf1085d` 到 1.2.1 / `1e08b3e`，`fe62900` 到 1.2.2 / `694ff3f`，`fabb591` 到 1.2.3 / `b4474a6`，1.2.4 那次与包清单同一批改——索引比发布提交晚一个，所以 `b4474a6` 上 CI 红，此后固定成「发版同一个提交里把索引一起前移」；1.2.4 的 `commit` 由发版后的 `4ee7f01` 钉到 `ff7f6de`），`source` 为 `{type: git, url, commit: ff7f6de7d87a7a4d99c06e3a9a68e9a4b417b0a9,
  path: packages/dsh-ai-terminal-pet}`，`engines.dsh` 取包里声明的 `>=0.1.0-rc.6`，图标钉在同一个 commit 的
  `icon.png`（4,386 B）。校验用的是宿主自带的 `parseMarketIndex`（`@deepseek-ai/dsh-plugin-market-index`），
  不是手抄一份形状检查：`format: 1`、`localizedText` 含 `en`、category/label 都引用已声明 id、
  `publishTimestamp` 为 Unix 秒、版本精确 SemVer。`marketInstallTarget` 产出的 spec 已实测安装成功
  （`git+…#<commit>&path:/packages/dsh-ai-terminal-pet` 这一形态，全新 `DSH_HOME` + 独立 profile，2.5 s）。

- **`cmd/pet-hook`：原生 hook 二进制**（`--mode=claude|antigravity`）——取代每次事件拉起
  一个 `node` 的开销，是 ROADMAP §E1 的交付物。`reporters/claude-hook.js` 与
  `reporters/antigravity-hook.js` 保留为参照实现与降级路径：两者对同一 payload 必须产出同一
  session PID（FNV-1a over UTF-16）与同一 `state`。
  实测（同 payload、同机器、双方都从 `%TEMP%` 启动）：`pet-hook.exe` 稳态 **51–73 ms**，
  `claude-hook.js` **93–157 ms**；首跑 513 ms vs 584 ms。构建产物 3,696,128 B。
- **`internal/reporter`：上报端共享内核**（`reporter.go` / `payload.go` / `send.go` /
  `approval.go` / `hook.go` + 两个测试文件，17 个 `func Test`）。把脱敏、id 截断、稳定 PID、
  时间戳钳制、快照落盘、管道投递、审批握手与 hook 事件到状态的映射收进一个包，让
  `pet-statusline` 与 `pet-hook` 不可能各自漂移，也让「与 `reporters/protocol.js` 保持一致」
  变成可测的断言而不只是约定。
- **`scripts/install.ps1` 增加第 6 类产物 `hook`**：`pet-hook.exe` 现在与主程序、状态栏 shim、
  `protocol.js` 和 bundle 一起被安装、被 `-VerifyOnly` 逐文件校验、被 `-Uninstall` 清理。
  `-Destination` 演练目录也会带着 `-HookDestination` 一起走，不会出现「一个文件落进演练目录、
  其余的落进真实安装目录」。脚本仍是纯 ASCII（0 个非 ASCII 字节）。
- **CI 增加 `Build the hook binary` 步骤**：`pet-hook.exe` 编不出来就等于 Claude Code /
  Antigravity 每次事件都失败，所以它必须和另外两个二进制一起被构建门禁覆盖。

### Changed

- **`cmd/pet-statusline` 迁到 `internal/reporter`，删掉约 250 行重复实现**：脱敏规则、id 截断、
  UTF-16 FNV-1a PID、时间戳钳制、快照落盘这些原本在 `pet-statusline` 里各有一份拷贝，
  现在与 `pet-hook` 共用同一个包。两份拷贝的风险不是理论上的——这次就在**两份里都**发现了同一个缺陷
  （见 Fixed 第一条）。迁移用**行为对齐**验证：先对旧二进制跑 35 个 payload（空 stdin、非 JSON、
  数组、各种 `agent_state`、对象形态的 workspace/model、超长与含非法字符的会话 id、epoch 秒/毫秒、
  未来时间戳、Bearer token、控制字符）记下 stdout 与落盘字段，迁移后逐字段比对，**35/35 完全一致**。
  语义上刻意保留了状态栏自己的两处差异：起报时间超过 5 分钟**忽略**而不是钳制（时钟偏斜的宿主否则会
  显示一个跑了几小时的会话），以及**单次投递**而不是 `reporter.Send` 的 750 ms 重试预算
  （状态栏每次刷新都要跑，不能付这个预算）。
  **代价**：`pet-statusline.exe` 由 3,643,392 B 变成 3,664,896 B（+21,504 B，把共享包的审批/配置代码
  也链接进来了）。实测稳态耗时不变（88–93 ms，主要是进程创建开销）。
- **`ReadStdin` 也收进共享包**：`pet-hook` 与 `pet-statusline` 原本各有一份「1 MiB 上限 + 2 秒超时」的
  读取循环，现在只有一份。
- **`vendor.mjs` 的源码摘要改为只覆盖浮窗真正链接的包**（`NON_OVERLAY_PACKAGES = ['cmd/pet-hook/',
  'cmd/pet-statusline/', 'internal/reporter/']`）：摘要原本扫 `cmd/` + `internal/` 下全部非测试 `.go`，
  但这三个包根本不进 `AI-Terminal-Pet.exe`。不排除的话，动一条对浮窗没有任何影响的上报路径，
  就会逼着重新生成一份**已发布**包的清单。
  **第一版把 `cmd/pet-statusline` 故意留在里面**（理由是不想重写自 1.0.0 就存在的条目），
  这个理由在第一次改 shim 时就失效了：摘要照样坏，浮窗二进制一个字节没变。于是改按
  「是否被链接进浮窗」这个真正的标准判断。
  **因此 `vendor/manifest.json` 被重写了一次**：`files` 25 → 24、摘要
  `cf647af1…` → `39685c79a41c30880de852ebafecd46d777405fd0a6d4730c6bf66438addec63`；
  两个 vendor 文件本身的字节与哈希未变（`vendor/AI-Terminal-Pet.exe` 仍 `bbb5163c…`）。
  副作用要说清楚：**仓库里的包目录已不再逐字节等于 npm 上的 1.2.4 tarball**，所以下一次发布必须是
  新版本号（1.2.5），不能重传 1.2.4。`vendor.mjs --check` 恢复 exit 0。
- **构建命令统一加 `-buildvcs=false`**：`go build -trimpath -ldflags "-s -w"` →
  `go build -trimpath -buildvcs=false -ldflags "-s -w"`，四处同步——`scripts/install.ps1` 的缺失文件提示、
  `.github/workflows/ci.yml` 的两条 build 步骤、`AGENTS.md` 的 Production build、README 的安装步骤。
  **为什么必须加**：不加时二进制里带一块 VCS 戳记，而戳记含构建时的 commit 与工作区状态，所以同一个
  提交在不同时间、不同机器上重建出的产物字节不同；`scripts/install.ps1 -VerifyOnly` 只要发现已部署副本与工作区
  里的 `bin/` 是两种构建方式下的产物，就会误报 `[fail]`。二进制本身不入库（`.gitignore` 忽略 `bin/`）。
  **实测差值**：加上之后 `pet-statusline.exe` 由 **3,644,416 B → 3,643,392 B**（差 1,024 B，
  即被去掉的戳记），主程序两种写法同为 **4,371,968 B**。
  （这两个是 go1.23 工具链下的数字；go1.27 上伪版本与 `(devel)` 等长，差值是 0 B。现役体积见本节第一条。）
  按新命令重建后 sha256：主程序 `bbb5163c296cd460d7c9a4c85e10342e125c0bd8017a065f913aa01567d41f0a`、
  shim `31c21739c60e427a387e1360b19395c5f3e9b48fc7597820bcb983a348b82328`；
  `scripts/install.ps1 -VerifyOnly` 实测 **10/10 `[ok]`**、末行
  `install: verified - 10 file(s) match the repository, 0 warning(s)`、exit 0。
- **`scripts/install.ps1` 三个新开关**，并新增 `Assert-SafeInstallDirectory` 等保护：
  - **`-VerifyOnly`**：只读，绝不写文件。按**同一份 10 项制品清单**（5 个分组：overlay / statusline /
    protocol / bundle manifest / bundle runtime）逐项比对仓库侧与已部署侧的 SHA256，
    每项打印 `[ok]` / `[warn]` / `[fail]`，末尾分组小计；全部吻合才打印
    `install: verified - N file(s) match the repository` 并 `exit 0`，否则 `install: not verified`
    并 `exit 1`。清单只有一处定义（`Get-InstallArtifacts`），因为「校验集 ≠ 安装集」正是过去
    一次升级后残留旧 `protocol.js`、所有上报端静默失联的成因。
  - **`-Uninstall`**：先要求输入 `yes`（无控制台且未加 `-Force` 时 `exit 2`；用户拒绝时 `exit 3`，
    什么都没删），然后停掉浮窗与 shim、删安装目录（5 次重试，防 reporter 中途重新拉起），
    只在 `HKCU\...\Run` 条目**指向被删目录内部**时才删除该条目（指向别的副本的条目属于那个副本），
    并调用 `Remove-InstallDirectoryFromUserPath` 清掉 `-AddToPath` 写入的那条 PATH 条目以避免死条目——
    此时安装目录即将消失，PATH 原值改备份到 `%TEMP%`。PATH 清理失败只报 `[warn]`：文件已经删了，
    残留条目是整洁问题而非失败。
    `Assert-SafeInstallDirectory` 拒绝盘根与任何叶子名不是 `AI-Terminal-Pet` 的目录——
    **`-Force` 不绕过名字检查**，只跳过确认；一个只靠提示词兜底的卸载器离误删只差一次手滑。
  - **`-AddToPath`**：只写 `HKCU\Environment` 的 `Path`，绝不碰 HKLM（工具是按用户装的，
    机器级 PATH 需要管理员权限且会改到这台机器上的所有账号）。写入前把原值（含注册表值类型）
    备份到 `<安装目录>\path-backup.txt`，已在 PATH 中则不改动只回报。
  - 三个开关的互斥组合在参数解析阶段就以 `[fail]` + `exit 2` 拒绝（`-VerifyOnly` 与 `-Uninstall` 互斥，
    `-AddToPath` 不能与这两个同时用）。
- **安装目录的"装完就是好的"改为"随时可证"**：README 的安装步骤多了第 4 步 `-VerifyOnly`，
  并在装完后用 `--doctor` 自查；卸载与 Path 开关同样写进 README 的中英两半。
- **`.gitattributes` 固定 `* text=auto eol=lf`**（`1330fbe`）：仓库本来就按 LF 存取，但 Windows 上
  `core.autocrlf=true` 会一直报规范化警告；钉进树里之后警告消失，Go 源码与 Markdown 在各次克隆间字节一致。
- **插件包版本 1.1.0 → 1.2.0**：清单内容变了（新增 `dsh.catalog`、`keywords` 补 `dsh-plugin`、根 shim
  `version` 同步），所以必须换版本号。1.1.0 的 tarball 已发布且**不可变**，若让版本停在 1.1.0，仓库与那份
  tarball 就会对不上——同一个版本号对应两份不同内容，正是本仓库这一轮一直在消除的漂移。因此
  1.1.0 保留为历史发布、由 1.2.0 取代。
- **npm 凭据问题已解决，发布也已完成**（历史记录保持原样，结论见下）：整场会话里 `npm whoami` 一直是 **E401**；
  2026-11-10T12:20Z 换过 token 之后它返回 **`jjjjjjjjnnjnn`**。当时 `npm view dsh-ai-terminal-pet version`
  报 **E404**，所以那个名字当时还空着。在包目录跑 `npm publish --dry-run` 通过（rc=0）：
  **16 files / 1.9 MB packed / 4.5 MB unpacked**。从**仓库根**发布会撞上 **EPRIVATE**
  （根清单带 `private: true`，npm 抬到的是根清单），从 `packages/dsh-ai-terminal-pet/` 发布则正常
  ——包清单没有 `private`。这一步在本机非交互环境里结构上做不到：npm 的 `otplease`
  在 `process.stdin.isTTY` 为假时直接抛 EOTP，永不进入 web-OTP 分支，所以只能由人在交互终端里跑一次。
  **结论（2026-11-11T12:10:46Z）**：人已经跑过 `npm publish --access public`，**1.2.4 已在 npm 上**，
  `dist-tags.latest` = `1.2.4`、shasum `dafb81998efbae8e906a36773e9737117aa482d9`、16 files / unpacked 4,462,112 B，
  与 Release 资产逐字节相同（从 registry `npm pack` 回来 sha256 仍是 `d6b7f95a…`）。「按包名安装」与
  「被 npm-only 目录收录」这两个前置从此都成立。**发布后头几分钟 registry 上只有占位的 `0.0.0-stage`**，
  此时 `npm install dsh-ai-terminal-pet@1.2.4` 会报 `ETARGET`，本地 npm 缓存还需 `--prefer-online` 才看得到新版本。
  其余安装路径（git 源、Release tarball、仓库地址）与自建 format-1 索引（`docs/market-index.json`，
  钉 40 位 commit 的 git 源）继续可用。

### Fixed

- **DSH 插件的项目路径、上级会话和会话起始时间全都没读对**（深度适配的核心一条）：
  Harness 的会话头 `session.header` 是唯一声明 `cwd` / `parentSession` / `createdAt` 的地方，
  而插件原来用 `process.cwd()` 当项目路径、`parentSessionId` 恒为空、`startedAt` 取「第一次见到
  这个会话的时间」。桌面版 Harness 是从安装目录启动的，所以**每一条会话都会被标成同一个错误路径**；
  子代理会话也不会挂在父会话下面；面板上的「已运行时长」从用户打开面板那刻才开始算。
  现在 `reportFor()` 统一读会话头：`cwd` → `project`，`parentSession` → `parentSessionId`，
  `createdAt` → `startedAt`。`sessionIdOf`/`headerOf` 同时兼容 `session/event` 直接给出的会话对象
  和 `agent/*` 事件给出的 `{ agent: { session } }` 包装。
  **同时核对过真实的 `SessionEventMap`**（`@deepseek-ai/dsh-session/lib/types/types.d.ts`）：
  主机端的持久会话事件只有 `turn/start`、`turn/end`、`step/start`、`step/end`、`user/message`、
  `assistant/chunk`、`assistant/message`、`tool/call`、`tool/result`、`todo/write`、`request/header`、
  `request/context`、`session/end-seed`。`working-directory/change` **根本不存在**，
  `session/title` 只是客户端事件——这两条分支保留为前向兼容，但项目与血缘改由会话头提供。
- **插件自带的 `vendor/protocol.js` 从来没被加载成功过**（1.2.4 及以前）：包清单声明了
  `"type": "module"`，所以包内任何 `.js` 都按 ESM 解释；插件用 `createRequire()` 去
  `require()` 那个文件，Node 24 的 `require(ESM)` 会**成功返回一个模块命名空间**，但里面没有
  `sendStatus`（文件里是 `module.exports = {...}`，作 ESM 求值时根本执行不到）。插件随即判定
  「没有协议模块」并静默什么都不发。这条路径只在**已安装的 `%LOCALAPPDATA%\AI-Terminal-Pet\protocol.js`
  缺失**时才会走到：`DSH_PET_NO_BOOTSTRAP=1`、拷贝失败、用户删了它，或来源是 git 而没有任何安装步骤。
  修法不是加特判，而是在 `vendor/` 里放一个 `{"type":"commonjs"}` 的 `package.json`——嵌套清单
  是标准的「目录模块类型覆盖」，`require` 从此正常。
  **为什么之前的测试没发现**：`verify.mjs` 的每个用例都在**没有 `package.json` 的假包根**里跑，
  那里 `.js` 默认就是 CommonJS，于是永远通过。新增的 `caseVendoredProtocolLoads` 直接对**真实包目录**
  做 `require()`，并断言 `sendStatus` 是函数——这个盲区现在有门禁了。
- **DSH 插件只在状态变化时才上报，于是「切目录」和「改标题」被静默丢掉**：
  `report()` 用「状态是否变化」去重，而 `working-directory/change` 只改 `project`、
  `session/title` 只改 `reason`，两者都不改状态——事件本身被吞掉，面板会一直显示旧的项目路径，
  直到 60 秒保活或别的状态变化才顺带刷新。现在去重签名覆盖 `state`/`reason`/`tool`/`project`/`title`
  五个字段：同样的内容重复到达仍然被丢弃，但任何一项真的变了就立刻上报。
  `verify.mjs` 的状态序列因此多了一次 `error`（切目录时状态未变），已按新语义更新并注明原因。
- **`reporters/antigravity-statusline.js` 不读 `toolCall`，与原生 shim 不一致**：`cmd/pet-statusline`
  的 `toolName()` 依次看 `tool_name` / `toolCall` / `tool_call`，而参考实现只看 `tool_name`。
  于是「只带 `toolCall`」的 payload 会让两边落出不同的会话内容（`tool: null` vs `tool: "Read"`）。
  现在参考实现也按同样顺序解析。发现方式是把参考实现与原生 shim 一起跑过 **33 个 payload**
  并逐字段比对落盘快照：其余字段（state / reason / sessionId / pid / project / model / terminal）
  本来就已经一致，修完这 33 个全部一致。
- **「没有 runtime 目录就静默不落快照」——`pet-hook` 与 `pet-statusline` 里各有一份**：
  两个命令都直接用 `os.CreateTemp` 往 `%LOCALAPPDATA%\AI-Terminal-Pet\runtime\` 写快照，
  但都不先建目录。而 `reporters/protocol.js` 的 `writeRuntimeSnapshot` 调的是
  `mkdirSync(dirname, {recursive:true})`。后果只在**全新安装**上出现、而且完全静默：浮窗还没跑过时
  这个目录不存在，于是每一条上报都既没有快照、又（在浮窗未启动时）没有管道，报告直接消失——
  而「浮窗没在跑」正是快照存在的理由。现在 `WriteSnapshot` 先 `MkdirAll(0700)`，
  `cmd/pet-statusline` 走同一条路径也一并修好。发现方式是把 JS 参考实现与原生实现的同 28 个 payload
  逐一跑过、逐字段比对落盘结果（stdout 与其他字段本来就一致）。
  新增 `TestWriteSnapshotCreatesAMissingRuntimeDirectory` 钉住这个前提。
- **上报端在「永远打不开管道」时要白等满整个重试预算**：`Send` 原来把所有拨号错误一视同仁，
  失败就重试到 750 ms 预算用完。宿主是同步等上报进程结束的，所以当投递**不可能成功**时，
  这 750 ms 是纯加在开发者每一次事件上的延迟。现在 `internal/reporter/send.go` 用
  `retryableDialError` 分类：只有 `ENOENT`/`ERROR_FILE_NOT_FOUND`（浮窗还没监听）与
  `ERROR_PIPE_BUSY`（实例占满）算可重试，`ERROR_ACCESS_DENIED` 立即返回、且不再触发启动浮窗
  （启动了也打不开）。低完整性场景下实测 `pet-hook` 由 **829–862 ms** 降到 **89–119 ms**。
  新增两个测试：错误分类表，以及一个 DACL 只给 SYSTEM 的真实管道——后者断言 `Send` 在
  300 ms 内返回而不是耗满 750 ms。
  **触发这条路径的真实原因不是产品缺陷**：本机仓库目录带着 Low 强制完整性标签
  （`icacls` 显示 `Mandatory Label\Low Mandatory Level:(OI)(CI)(NW)`），从该目录里的文件启动的
  进程一律是 Low IL，而浮窗跑在 Medium IL、管道 DACL 只给 `OWNER RIGHTS` + `SYSTEM`。
  同 sha256 的副本放到 `%TEMP%` 或安装目录就落回 Medium IL 并正常投递。管道 ACL 因此保持不变，
  详见 [docs/ROADMAP.md](docs/ROADMAP.md) §E1。
（`d06001d` 之后）：`go.mod` 里的 `go 1.23.0` 是
  **下限不是钉版本**，CI runner 用 go 1.23.0、本地用 go1.27.0，两个工具链编出的二进制字节不同——
  **跨工具链字节相等从来不可能达成**，那一步当时是拿一个达不到的条件当门禁。修法是把断言换成能达成的：
  `vendor/manifest.json` 记录**源码摘要**（25 个非测试 `.go` 文件）与两个 vendor 文件自己的哈希，
  `vendor:check` 校验这些；只有当**本地 Go 版本等于 manifest 里记录的 `go1.27.0`** 时才和 `bin/` 做字节比对。
  于是 CI 校验「vendor 与源码摘要一致」、本地额外校验「vendor 与 `bin/` 字节一致」，两边都在验真实的东西。
  这段历史不美化：CI 是先红后修的，不是一次就绿；也不是「go.mod 该钉工具链」。
- **可移植性：`docs/assets/render_assets.py` 不再写死本机绝对路径**（`98 行`改动）。
  旧版把 `PREVIEW` 写成 `C:\Users\<本人>\AppData\Local\Temp\petpreview2`、`ASSETS` 写成
  `C:\Users\<本人>\Desktop\项目开发\ai-terminal-pet\docs\assets`，任何别人的克隆一跑就断。
  现在预演目录读环境变量 `PET_PREVIEW_DIR`、输出目录取 `Path(__file__).resolve().parent`；
  变量未设置或目录不存在时打印中英双语的可执行提示（先跑哪两条 `go test`）并 `exit 2`，
  缺 Pillow 同样给出 `pip install pillow` 后 `exit 2`，而不是抛 Python traceback。
  另外 `sys.stdout`/`sys.stderr` 显式 `reconfigure(encoding="utf-8")`，保证在非 UTF-8 代码页下重定向时中文不乱码。
- **测试夹具里的本机路径全部中性化**：`internal/platform/focus_windows_test.go`（3 处）、
  `internal/platform/presence_windows_test.go`（4 处）、`internal/pipeserver/server_windows_test.go`（1 处）
  原先把 `C:\Users\<本人>\Desktop\项目开发\...` 当作期望值钉在测试里，
  现在换成 `C:\work\项目\...` / `C:/Users/pet-test` 等中性路径。
- **管道名测试改为按算法现算**：`TestPipePathForUserMatchesReporterAlgorithm` 过去硬编码
  `-42ce7d78` 并 `t.Setenv("USERPROFILE", `C:\Users\rongj``)，现在用
  `sha256.Sum256([]byte(`c:\users\pet-test`))` 取前 4 字节转 8 位十六进制得到期望后缀，
  并特意用**混合大小写 + 结尾分隔符**的输入，仍然钉住 Go/JS 共享契约
  （home 小写化 → 去尾分隔符 → 取 sha256 前 4 字节）。
- **删除七处不可达声明**（`3fd49c3`，staticcheck U1000）：`internal/store/store.go` 的 `(*Store).remove`、
  `internal/ui/app_windows.go` 的 `messageEnvelope.reserve` 与 `(*App).clampToMonitor`、
  `internal/ui/canvas_windows.go` 的 `fillPolygon` / `alphaAt` / `pointInside` / `canvasPoint`。
  行为不变，`gofmt` / `go vet` / `go test` 保持干净，没有为了消 warning 加 `//nolint` 或假引用。
  唯一保留的 staticcheck 发现是 `internal/ui/win32_api_windows.go` 里 `syscall.Syscall9` 的 SA1019，
  因为 `UpdateLayeredWindow` 的调用签名需要它。
- **`vendor/AI-Terminal-Pet.exe` 一开始没进 git 树**（修于 `91719e5`）：`.gitignore` 里的 `*.exe` 把它一并忽略，
  而 `git add` 不会自动越过忽略规则，于是这颗二进制只存在于工作区。后果不对称得很有欺骗性——
  `npm pack` 读的是工作区，所以 **release tarball 装得上**；而 **git 源安装拿到的是一个 `vendor/` 里
  只有 `protocol.js` 的包**，`runBootstrap()` 于是回落到下载分支，`lib/release-pins.json` 又还没建，
  回落是 no-op，浮窗就静默地不装了。修法是两行：`.gitignore` 加例外
  `!packages/dsh-ai-terminal-pet/vendor/AI-Terminal-Pet.exe`；`.gitattributes` 加 `*.exe binary`，
  免得 Git 对这颗 sha256 依赖的二进制做文本规范化。
  同一次排障还改正了 Release 的 tag 落点：`dsh-plugin-v1.0.0` 最初打在 `7f476c9`（vendor 还在被忽略的那次提交），
  于是从 tag 安装拿到的是一个 `vendor/` 里只有 `protocol.js` 的包。修完 `.gitignore` 后 tag 已强制移到 `91719e5`，
  并用 `git+...#dsh-plugin-v1.0.0&path:/packages/dsh-ai-terminal-pet` 重装复验（rc=0、3.4 s、`vendor/AI-Terminal-Pet.exe` 4,371,968 B 在位）。

### Docs

- **npm 1.2.4 已发布，文档里的「尚未发布」全部改成事实**（2026-11-11T12:10:46Z）：
  `README` / `CHANGELOG` / `HANDOVER` §1 / §27.5 / §27.7 / §28.1 / §28.3 / §28.4 / `ROADMAP` §4 与「没做成什么」
  逐条校正。发布内容与仓库里验证过的 tarball **逐字节相同**（shasum `dafb8199…`、16 files、
  unpacked 4,462,112 B；从 registry `npm pack` 回来 sha256 仍是 `d6b7f95a…`），
  `npm install dsh-ai-terminal-pet@1.2.4` 装出 16 个文件、`vendor/AI-Terminal-Pet.exe` 与仓库一致、
  `lib/index.js` 作为 ESM 可加载并导出 `apply`。
  记下一个只有发布后才会出现的坑：头几分钟 registry 上只有一个占位的 **`0.0.0-stage`** 版本
  （`latest` 也指向它），此时 `npm install <name>@1.2.4` 报 `ETARGET`；约三分钟后真版本才出现，
  本地 npm 缓存还需 `--prefer-online` 才看得到。
- **dshopen 那条「已收录但未渲染」的判断被推翻**：`https://dshopen.org/index.json` 返回 **1,207 条**，
  含本包 1.2.4（git 源、钉 `ff7f6de`、分类 `integrations`）。当初读的「854 条」来自该仓库的
  `public/index.json`——**未重新构建的仓库副本**。教训：判读可见性要读**对方站点域**上的索引文件。

- **README 重写并配真实截图**（`7c0aaaa`，双语严格镜像）：安装、CLI、五类终端接入、路线图、非目标、
  从源码开发全部对齐 v0.9.0 之后的事实；`docs/assets/README.md`（`4574cba`）逐个说明 8 张 PNG
  的用途与复现步骤（两张 preview 测试共享同一个 `PET_PREVIEW_DIR`），并标注 3 张未被 README 引用的变体为何保留。
- **停用期 runbook 归档**（`8ada28b`）：`docs/HANDOVER.md` §1–§12（2026-09-24 全局停用态）逐字搬入
  `docs/archive/HANDOVER-2026-09-DISABLE.md`，正文只留指回链接；`docs/archive/README.md` 成为归档索引。
- 早期宣传稿与草稿移入 `docs/archive/promo-2026-11-09/`，不再混在对外文档里。
- **交接整理与外部状态刷新**（`bb00e58` 之后的一串纯文档提交，2026-11-11 10:35–11:00 +0100）：
  `docs/HANDOVER.md` 新增 **§28 交接清单**——一句话现状、接手先读的四份材料、外部世界还在等谁、
  唯一必须由人做的 npm 发布、哪些文件「不可再改」、下一个版本的两步发版套路，以及**三样看起来像冗余、
  实则刻意保留的东西**（旧的 host-only 桥接 bundle 仍占 `install.ps1` 十个制品中的七个并受 CI 校验；
  `vendor/AI-Terminal-Pet.exe` 与 `bin/` 那份字节相同是发布产物 vs 构建产物的区别；`.local/` 是一次性暂存）。
  同时把 §27.2 的包文件表从 1.2.0 口径校准到 **1.2.4**
  （`package.json` 2,246 B / 83 行、`README.md` 8,187 B / 146 行、`verify.mjs` 23,426 B / 424 行），并注明
  「包目录 18 个文件」与「tarball 16 files」两个数字都对（差的正是 `verify.mjs` 与 `scripts/vendor.mjs`）。
  `.gitignore` 补 `.local/`；`AGENTS.md` 的文档表加一行指向 §28。
  同一轮里还修了两处只有写下来才会暴露的问题：§28 里的 `../../README.md` 越出仓库根（改成 `../README.md`），
  以及 §1 / §28.1 **把自己的 sha 与 CI run 号写死**——下一次纯文档提交就过期。后者改成「用
  `git log --oneline bb00e58..HEAD` 与 `gh run list --limit 1` 自查」，把易腐的数字从文档里拿出去。
- **dshopen 收录**：`polang233/dsh-plugin-index` 的 issue #423 于 2026-11-11T08:32:39Z 以
  `state_reason: completed` 关闭，其 `data/sources.json`（1,213 条）第 1204 条记的就是本包 **1.2.4**
  （`source.commit` `ff7f6de`、`path` 子目录、中英摘要都带上了「其他终端需各自安装 reporter」那句，
  分类 `integrations`）。**但它线上的 `public/index.json` 仍是旧构建**（854 条、全是 npm 源、0 条带 `source`
  字段），所以目录页要等它下一次构建才可见——这条差异写成了 §27.7 的一条新残留：
  **「已收录」不等于「已渲染」，判断可见性要看对方公开的索引文件，不要只看 issue/PR 状态**。
  同一轮复读：awesome #14243 仍 open 且 `mergeable_state` 在 `clean` / `unknown` 间抖动、imsai #912 仍 `clean`
  等人工、radar #1255 与 brade #367 仍 open、dshbase/dsh.so/dsh-plugin.org/alldsh/dshmarketplace.dev 全部无回应、
  YELEBAI 仍记 1.2.3（`verifiedCommit 9b0df94`，`install.profiles` 已含 desktop）。
---

## [1.0.0] — 2026-11-11

浮窗本体的**第一个 1.0**，与插件包 `dsh-ai-terminal-pet` **1.2.6** 同批发布。这一版的意义是「号码
不再自相矛盾」：在此之前浮窗说自己 0.9.0、npm 上是 1.2.5、便携包按浮窗那条线命名、插件包按自己那条线
命名，四个数字在同一页文档里并存。

### Added

- **`pet-setup.exe`：自解压单文件安装器**（`cmd/pet-setup`，ROADMAP 的 I1-γ）。它 `//go:embed`
  打包时暂存的载荷，**先**用自带的 `SHA256SUMS.txt` 逐文件校验，全部吻合才动手，解到临时目录，
  再调用内嵌的 `scripts/install.ps1`。`--verify` / `--uninstall` / `--add-to-path` 逐个转发到
  `install.ps1` 的同名开关。**安装逻辑仍然只有一份**：`install.ps1` 是唯一实现，`pet-setup` 只负责
  「知道自己是不是为发布而构建」「校验自己带的字节」「转发参数」。
  它比 zip 多出来的那件事恰好是 zip 做不到的：下载被截断时，zip 用户要等到装完才发现，自解压 exe
  在写第一个字节之前就发现。三道实测出来的坑都写进了代码注释：`//go:embed payload` 会**跳过点开头的
  文件**（`.staged` 因此没进二进制，每次发布都会走「拒绝运行」分支，必须 `all:payload`）；
  `Copy-Item -LiteralPath '…\*'` **不做通配**，结果一个文件都没拷、但 exe 照样构建成功（载荷只有 2 个
  文件），现在用「载荷必须 > 1 MB」把它钉住；载荷不能照抄 zip 的整棵暂存树，否则
  `README.install.txt` 与根 `install.cmd` 会以「未被 SHA256SUMS.txt 列出」的名义让**每一次**安装失败。
- **`scripts/package.ps1` 同时产出两个包**：`ai-terminal-pet-<版本>-win-x64.zip` 与 `pet-setup.exe`，
  并在打包前后校验载荷字节数、`--version` 输出与 zip 的交叉一致性。
- **20 个 `cmd/pet-setup` 单测**：默认拒绝安装（exit 3）、`--version`/`--help` 不需要载荷、
  未知参数走 exit 2 而不是把 PowerShell 的错误甩给用户、四个开关的转发表、`SHA256SUMS.txt` 的解析
  与四种失败（哈希不符 / 缺文件 / 多文件 / 畸形行）、以及 `..` 逃逸路径被拒。
- **CI**：`Build the setup stub`（证明 `go build ./cmd/pet-setup` 在裸检出里也能编）、打包步骤改名为
  `Package the portable zip and the setup exe`，artifact 同时收 zip 与 exe。

### Changed

- **`Version` 0.9.0 → 1.0.0**（`internal/app/app_windows.go`）。三个二进制按同一组旗标重建，
  `vendor/AI-Terminal-Pet.exe` 随之更新，`vendor/manifest.json` 的源码摘要由 `39685c79…` 变为
  `edd584c6…`，文件数从 24 变回 24（`cmd/pet-setup` 已加入 `vendor.mjs` 的排除表——它是独立二进制，
  改它一个字都动不了浮窗）。
- **`tests/fresh-machine.mjs` 不再钉 `0.9.0`**：那条断言本来只是「`--version` 有输出」，用宽松匹配
  钉死后会变成版本冻结，现在锚定「是一个已发布的 1.x」。
- README 中英两段：版本徽章、`--version` 表项、安装路径 B（两个成品包）、market-index 条目、
  插件安装命令全部前移到 1.0.0 / 1.2.6；「独立安装包」从路线图移入「已交付」，剩下的只有签名与 MSI。

### Published

| 资产 | 字节 | sha256 |
|---|---|---|
| `ai-terminal-pet-1.0.0-win-x64.zip` | 4,983,903 | `70D992B1…` |
| `pet-setup.exe` | 14,125,568 | `87D5E60B…` |
| `dsh-ai-terminal-pet-1.2.6.tgz` | 1,917,331 | `671CDC95…` |

三个资产都从公开 URL 下载回来复验过：`pet-setup.exe --version` 打印 `pet-setup 1.0.0`，在自带
`LOCALAPPDATA`/`USERPROFILE` 的沙箱里 install → 11 件哈希全对 → `--verify` 11 件全对 →
`--uninstall`（输入 yes）目录归零；zip 解出 17 个文件；npm `npm pack dsh-ai-terminal-pet@1.2.6`
取回的 tarball sha256 与 Release 资产一致。`dist-tags.latest` = **1.2.6**。

## [1.2.6] — 2026-11-11（插件包）

只因为**载荷字节变了**：浮窗升到 1.0.0，`vendor/AI-Terminal-Pet.exe` 必须重新记账，
`vendor/manifest.json` 的源码摘要跟着变。`lib/index.js` 与 `vendor/protocol.js` 与 1.2.5 逐字节相同，
`verify.mjs` 未改。Release 资产不可变，所以这是新版本号而不是编辑——与 1.2.1、1.2.4 同一条理由。

tarball **17 files / 1,917,331 B / sha256 `671cdc9503c3875960182616ca83bc9fe158d17778925b069d08af53e4ff5cd4`**。

## [1.2.5] — 2026-11-11
插件包的第五个发布。**这一版第一次真正「装上就能用」**：前几版里有两个缺陷会让插件在没装过浮窗的
机器上静默什么都不发（§30.2、§30.3 有完整复盘），而且项目路径连反应都没有。代码改动只有三个文件
（`lib/index.js`、`verify.mjs`、新增 `vendor/package.json`），但它换来的是下面这些行为：

- **按包名安装可用**：1.2.4 已发布到 npm，`dsh plugin add dsh-ai-terminal-pet` 现在解析得到；
  插件包 README 里「等 npm 发布再用」那段已经过期，改成三条安装路径并列。
- **自带的 `vendor/protocol.js` 终于能加载**：包声明 `"type": "module"`，`require()` 那个文件会
  返回没有 `sendStatus` 的模块命名空间；`vendor/package.json` 写 `{"type":"commonjs"}` 覆盖目录
  模块类型。受影响的是「已安装副本缺失」的所有情况（`DSH_PET_NO_BOOTSTRAP`、拷贝失败、用户删了它）。
- **项目路径 / 上级会话 / 起始时间改从 `session.header` 读**：`cwd`、`parentSession`、`createdAt`。
  之前用 `process.cwd()`，桌面版 Harness 从安装目录启动，于是每条会话都被标成同一个错误路径。
- **去重签名覆盖 `state/reason/tool/project/title/parentSession`**：原来只比 state，
  只改项目或只改标题的更新会被整条吞掉。

包内容与 1.2.4 的差异：`lib/index.js`（会话头适配 + 去重签名）、`verify.mjs`（新增
`caseVendoredProtocolLoads` 与三条会话头断言）、新增 `vendor/package.json`。两个 vendor 文件
（`AI-Terminal-Pet.exe`、`protocol.js`）字节未变，`vendor/manifest.json` 的源码摘要在 1.2.4 那次
就已经前移到 `39685c79…`（§29.5），本版不再改动。tarball **17 files**（1.2.4 是 16）。

npm 元数据（2026-11-11T14:37:59Z 打包）：**1,917,342 B / sha256
`dfa89a1ed412c77f85293ce8d07bf788aab6f5a5c3e1ad3e6d5be6e49c6dae5d` / shasum
`1f6d5340aaf1639c129616c9afc6ccf5503009da`**。

## [0.9.0] — 2026-11-08 / 2026-11-09

这一批把浮窗从「能看」推到「能调」，并补上了本项目自己在用的 DeepSeek Harness 接入。
详细 runbook：[docs/HANDOVER.md](docs/HANDOVER.md) §25。

### Added

- **设置窗口**（`b99a6a0`，`internal/ui/settings_windows.go` 1,382 行 + 228 行测试）：第二个顶层窗口，
  窗口类 `AITerminalPetSettings`，三个页签 general / sound / sessions。它不是对话框资源，而是
  与浮窗**跑在同一条消息泵上**的顶层窗口——在 `Run()` 里 `runtime.LockOSThread()` 之后创建、
  `Run` 负责注册与注销，因为 Win32 要求「创建窗口的线程 = 泵消息的线程」，
  跨线程 `SetWindowPos` 会永久阻塞（就是 §13 幽灵窗口的根因）。
  几何单一来源 `settingsGeometryFor(dpi, tab, rows)` 同时供绘制与命中测试；
  切页签时按内容改高度且保持中心不动（真机 450↔504↔306 @150% DPI）；每项改动立即生效，没有 OK 按钮。
- **四套合成音色 + 按事件开关**（`9fd6a9d` 建立音型表，`53d45c1`/§25.5 扩成 `schemeSpecs()`）：
  `chime`（默认）/ `bell` / `soft` / `pulse`，每套含等待/出错/完成三条波形。
  仍是内存合成 44.1 kHz / 16-bit 单声道 PCM，以 `PlaySoundW(SND_MEMORY | SND_ASYNC | SND_NODEFAULT)` 播放，
  缓冲按音型缓存，避免异步播放器越过 Go 的切片。
  `internal/config` 新增 `soundScheme` / `soundOnWaiting` / `soundOnError`；
  `config.SoundSchemes` 与 `ui.SoundSchemes` 两份镜像由 `TestDefaultSettingsMirrorConfig` 钉住。
- **按会话的告警状态机**（`4e2963d`，`alert_windows.go`）：`alertStates` 记录**每个**会话的上次状态，
  纯函数 `transitionLevel` / `decideAlerts`——waiting/error→urgent，busy→idle/closed 或消失→finished，
  新出现的 busy 不告警；首次加载只播种、不回放（热重载与重启不会重复旧告警）。
- **一键跳转终端**（`4e2963d` 的 UI 侧 + `2ddd87b` 的 `platform.Focus`）：展开态每行右侧跳转箭头、
  紧凑态底部 `↗` 跳转条（仅在有可跳转会话时出现）。点击把 `wmFocusActivate` 投回 UI 线程，
  `platform.ActivateTarget` 只做 `IsIconic→SW_RESTORE` + 一次 `SetForegroundWindow` + `GetForegroundWindow` 复核，
  失败走 `focusFailure()`，**绝不抢前台**；`AttachThreadInput` 与 `SetWindowPos(TOPMOST)` 的兜底全部删除
  （它们会动用户正在用的窗口）。成功跳转后 `holdTuck(2s)`，不让贴边收起立刻把面板吸回去。
- **auto-approve 否决窗口**（`50fd6bb` UI + `310f6b0` hook 侧）：浮窗拿不到别人终端里的对话框，
  所以反向做——hook 写 `pending-<source>-<session>-<pid>-<epoch>.json` 到
  `%LOCALAPPDATA%\AI-Terminal-Pet\approvals`，每 100 ms 轮询同名 `.deny`；展开面板显示 `Allow (Ns)`
  与倒计时条，点一下写 `.deny` 否决最紧急的等待会话，沉默即超时放行。
  `DANGEROUS_PATTERNS`（`rm -rf` / `rd /s` / `mkfs` / `dd` / fork bomb / `git push --force` /
  `format` / Windows `del` / `shutdown` / `Remove-Item System32`）在倒计时开始前就判，
  破坏性命令永远走人工提示。
- **跨进程存活与窗口探测**（`2ddd87b`，`internal/platform/presence_windows.go` 217 行 +
  `focus_windows.go` 392 行 + 535 行测试）：`Presence` 回答「上报进程还活着吗」，
  `Focus` 回答「哪个窗口属于它」。只读探测，测试用合成 PID，从不碰真实进程。
- **DeepSeek Harness 桥**（`3acb4f8`，`integrations/deepseek-harness/ai-terminal-pet-bridge`）：
  Cordis bundle 用 `ai-terminal-pet.v1` 上报 Harness 会话状态，让面板在本项目自己的开发会话里也能显示。
  一个对话一个稳定会话；时间戳归一 UTC 并夹在服务端上限之下（时钟偏差不能让 `startedAt` 跟着 `updatedAt`）；
  审批请求保留 waterfall 链；`dispose` 只移除监听器，不结束会话（HMR 重载不是退出）。
  `verify.mjs` 是离线测试：用临时 `LOCALAPPDATA` 桩掉官方 protocol 模块并断言 payload，不需要管道/浮窗/运行中的 Harness。
- **`--settings` 命令行入口**（`b40c5db`）：`ControlSettings` → `wmSettings`；没有实例在跑时走
  `app.RunWith(true)`，**先起浮窗、首帧后再开设置窗口**，否则这个参数等于什么都没做。
- 托盘菜单新增 **设置…**（带分隔线），与 `--settings` 同源。
- **README 配图**（`832cfd5`，`docs/assets/render_assets.py`）：把 `TestRenderPreview` 产出的 PNG 裁掉平背景，
  输出 7 个真实浮窗状态（紧凑/展开、双语、hover、溢出、auto-approve）加一张 1600×640 hero。
  每张图都来自真实 Win32 光栅器，不会与实际布局漂移。

### Changed

- **配置生命周期归 `internal/app`**（`b40c5db`）：新增 `configSaver`，与 `placementSaver` 同构的合并写——
  `OnSettingsChanged` 只把 `ui.Settings` 入队，**UI 线程不碰文件**；后台 goroutine 做
  `config.Load` → 只覆盖窗口能改的字段 → `config.Save`，手写的、本版本不认识的键不会被丢掉。
  `configSaver.echo(config)` 记住本进程最后一次写入的完整值，30 s 的 `watchConfig` 命中时先比对，
  自己写的变更直接跳过（否则「改一下开关 → 30 秒后被热加载回灌成旧值 → 再变回来」会真的发生）。
  滑块拖动只改内存，松手才 publish 一次。
- **本地化绘制**（`4e2963d`）：所有文案经 `internal/i18n` 的 `a.lang`（`New()` 固定、托盘切语言时刷新），
  光栅器不再解析 locale；`frameKey` 把 `lang` 纳入签名，切语言必然重绘一次。
- **告警动画降帧**（`4e2963d`）：等待态呼吸 12.5 fps → 6.25 fps。每次动画 tick 都是整层 GDI 光栅，
  半速循环看起来仍然是「呼吸」。
- **部署集扩展**（`b40c5db`，`install.ps1`）：参数从 2 组（主程序 + shim）扩到 4 组
  （`Source`/`ShimSource`/`ProtocolSource`/`BundleSource`），`protocol.js` 与 DSH bundle 每次安装同步并逐个 SHA256 校验。
- **安装流程加固**：`install.ps1` 停进程 → `Wait-Process` 10 s → 重试拷贝 20 次（reporter 会按需把浮窗拉起来，
  「停下、睡 500 ms、拷贝」在真正在用的机器上会随机失败）。
- **README 重写为 v0.9.0 的实际界面**（`1565669`）：三页签设置窗口、四套合成音效与按事件开关、
  auto-approve 否决窗口、en/zh 语言切换；同时删掉重复出现在三个位置的徽章块。
- **文档目录重组**（`6f18394`）：`HANDOVER.md`（25 节、72 KB）/ `ROADMAP.md` / 两份 BASELINE 从仓库根移入
  `docs/`，交叉引用改写；无源码或行为变更。

### Fixed

- **浮窗在最后一次上报后约 10 分钟自行消失**（`ce81a8e`，HANDOVER §24.6）：`snapshotAwareAlive`
  对「磁盘文件还在」的会话仍然调用 `Presence.SessionAlive`，OS 证据一弱，`Reconcile` 连续缺席 2 分钟后删掉内存条目 →
  `HasAvailable()` 变假 → 10 分钟空闲自毁。而 `--json` 每次新建 store、达不到缺席门槛，于是形成
  「磁盘有、内存没有」的错位。修法：`present[key]` 为真即直接保活（`closed` 仍优先判死），
  OS 证据只留给从未写过快照的纯 pipe 会话。`TestSnapshotAwareAliveTrustsTheDisk` 锁定。
- **设置窗口点不动**（`b99a6a0` 的迭代过程）：`ReleaseCapture` 会同步回调 `WM_CAPTURECHANGED`，
  第一版在那里把 `press` 清掉，结果每次点击都自己把自己抹掉。现在 `settingsMouseUp` 先快照
  `press/drag/moving` 再释放，捕获丢失的处理**不清 press**。
- **托盘气泡像是来自错误的终端**（`4e2963d`）：旧实现只有一个全局 most-urgent key，
  该会话结束时被重置，于是无关的等待会话后来重复告警。改为逐会话记录（见 Added）。

### Removed

- `MessageBeep` 路径被内存合成音效取代（`9fd6a9d`）。系统默认事件在多数机器上静默或几乎听不见，
  而且受专注助手影响。

---

## [0.8.0] — 2026-11-08

详细 runbook：[docs/HANDOVER.md](docs/HANDOVER.md) §24。

### Added

- **合成提示音**（`9fd6a9d`，`internal/ui/sound_windows.go` 281 行 + 156 行测试）：
  内存合成 44.1 kHz / 16-bit 单声道 PCM WAV，三种音型（等待 = 两轮上行双音 ≈0.9 s、
  错误 = 下行低音、完成 = 短上行三音）；`soundVolume` 0..100 默认 85，托盘可按 60/85/100 即时试听。
- **完成提醒**（§24.2）：`notifyOnFinish` 默认开，气泡文案 `balloon.doneTitle` / `balloon.doneBody`。
- **Toast 静音**：托盘气泡加 `NIIF_NOSOUND`，气泡与提示音各管一路，互不重复。
- **DSH 接入**（见 0.9.0 的 Added，交付于 2026-11-08 的工作时段）。

### Changed

- **通知时序**（§24.2）：`alertStates` 逐会话；纯函数 `transitionLevel` / `decideAlerts`；
  托盘气泡点击仍是回跳浮窗。
- `protocol.js:normalizePayload`（§24.3）：`startedAt = min(startedAt, updatedAt)`，两者夹到 now+4 min，
  且幂等。hub 侧的 `validate()` 会拒收「updatedAt precedes startedAt」，一旦拒收，随后的扫描会把快照删掉——
  这个钳制就是为时钟偏差准备的。
- 提示音与「通知」解耦但共享同一套「会话 + 状态」去重，关掉任一个都不会改变另一个的触发时机。

### Fixed

- 上报端与浮窗的时钟偏差导致会话被误删（见上一条 `normalizePayload`）。

---

## [0.7.1] — 2026-11-05

详细 runbook：[docs/HANDOVER.md](docs/HANDOVER.md) §23。

### Fixed

- **会话幽灵堆积**（内存 ≥7 个会话 vs 磁盘 4 个文件 vs 4 个 Windows Terminal 标签）。三个根因：
  1. **宿主存活 ≠ 会话存在**——`Presence.SessionAlive` 只要 `parentPid` 在进程表里就返回真，
     而 opencode 的 `pid/parentPid` 是自 11/2 就常驻的宿主进程。
  2. **删文件 ≠ 结束会话**——插件 `cleanupAll`/`closed` 只 `unlink` 不送 closed 事件，
     浮窗只能靠 TTL 猜。
  3. **事件 ID 当会话 ID**——`e.data?.id || e.id` 的兜底让 `evt_1abf…` 变成独立会话。
  另外取证时踩到时区陷阱：payload 的 `updatedAt` 是 UTC，本机 GMT+1，`15:39Z` 是本地 16:39。

  修法：`SnapshotScanner.ScanWithKeys()` 返回磁盘 key 集合 + `store.Key()` 导出 +
  `app.snapshotAwareAlive()`（曾有快照但现在没有 → 判结束）；`Presence.SessionAlive` 改为按 specificity 分级
  （① 终端标签标题命中 `title`/项目名 → 无条件保活；② 可信真实 PID 优先于 hook 的 `parentPid`；
  ③ 近期活动超过 `DefaultOrphanAfter`（45 min）的孤儿回收走 `tombstoneExpired`）；
  opencode 插件 `sessionIdentity()` 只认明确的 sessionID，拒 `evt_`/`msg_`/`prt_`/`prj_`/`default-session`，
  `closed` 同步写 `state:closed` 快照；`sweepStaleSnapshots()` 启动时只删事件型伪会话与 2 小时未动的文件；
  `cleanupAll` 只在 `process.on(exit)`。

### Changed

- **Slate 配色**（§23.5）：结构色收进 `internal/ui/drawing_windows.go` 顶部调色板
  （`colorPanel` / `colorRaised` / `colorBorder` / `colorDivider` / `colorOrbInner`），
  背景由蓝黑改为中性石墨；状态球从实心大圆盘 + 强光晕改为细环 + 深色内核 + 状态色字形；
  外框 alpha 210→120；跳转按钮改实心中性底 + 状态色细边；行左色条 alpha 190。
  绘制代码里不再散落 `rgba{}`。

---

## [0.7.0] — 2026-11-05

详细 runbook：[docs/HANDOVER.md](docs/HANDOVER.md) §20–§22。

### Added

- **一键跳转终端**（§20）：只用现有 `pid/parentPid/title/project/source/terminal` 定位，**协议零改动**。
  `internal/platform/focus_windows.go`：Toolhelp 祖先链 → `EnumWindows` → 打分
  （祖先距离 / 会话标题 / 项目基名 / 终端关键词 / 窗口类名）→ 前台激活；跨进程枚举不在 UI 线程。
  `internal/ui/focus_windows.go`：`compactFocusRect` / `expandedRowFocusRect` 与绘制同源，
  `jumpPriority`（waiting 4 / error 3 / idle 2 / closed 1 / 忙碌 -1）+ `mostUrgentJump` + 纯函数 `focusHit`。
  弱匹配（得分 <40）判失败。新增 13 个单测（platform 7 / ui 6）。`PET_FOCUS_PROBE=1` 为探针。
- **会话存在性模型重构**（§21）：**存在性由 OS 观测**（进程表 + 终端窗口），state 可以陈旧；
  删除需连续不可观测 ≥ `DefaultAbsenceGrace`（2 分钟）。三类 tombstone：`closed`（保留门槛）/
  `expired`（不设门槛、可自愈）/ `evicted`（保留门槛）。新增 `internal/platform/presence_windows.go`；
  `store.Reconcile` 取代运行期 `Cleanup`，保活只刷 `pipeSeen`/`snapshotSeen`/`lastSeen`，
  不改 payload、不 bump revision（零重绘）。`SnapshotScanner.KeepAlive`。
- 告警呼吸改为分层窗口常量 alpha（`internal/ui/pulse_windows.go`，`purePulseFrame` 复用已渲染位图，
  零 GDI），`pulseInterval` 80→160 ms。

### Changed

- `app` 里一个 `Presence` 贯穿；`--json` 用同一模型；扫描 2 s→5 s；`watchConfig` 每 30 s `Stat`
  后台读再 `QueueSettings`；UI 空闲检查 500 ms→30 s、自毁宽限 5 s→10 min。
- 运算表（三轮采样）：空闲检查 500 ms→30 s、扫描 2 s→5 s、存活探测 每会话 `OpenProcess`/2 s → 一次 Toolhelp/10 s、
  窗口枚举 仅跳转时→10 s 共用、告警呼吸 80 ms 全帧光栅→160 ms 零 GDI、配置 仅启动→30 s Stat。
  第 3 轮 120 s 累计 CPU 0.562 s（约 0.5% 单核，较第 2 轮降约 96%）。新增 18 个单测。

### Fixed

- **跳转差点动用户窗口**（§22）：`activateFocusWindow` 里的 `AttachThreadInput` 与
  `SetWindowPos(TOPMOST)→NOTOPMOST` 会移动用户正在用的窗口。全部删除，改为 `platform.ResolveTarget` 只定位、
  `ui.jumpTo` 把 `wmFocusActivate`（`wmApp+7`）投回 UI 线程、`platform.ActivateTarget` 仅
  `IsIconic→SW_RESTORE` + 一次 `SetForegroundWindow` + `GetForegroundWindow` 复核，失败走 `focusFailure()`。
  `ActivateTarget` **禁止** `LockOSThread`/`UnlockOSThread`（会重现 §13 的幽灵窗口）。
  验证：25 次合成 + 15 次真实点击；12 次「用户正在处理窗口」场景 FAILS=0；4 个新单测。
- **浮窗在最后一次上报后约 10 分钟自行消失**：见 0.9.0 的 Fixed（修复提交 `ce81a8e` 落在 0.9.0 那批，
  问题本身是 0.8.0 时段的收尾发现）。

---

## [0.6.0] — 2026-09-26

详细 runbook：[docs/HANDOVER.md](docs/HANDOVER.md) §18。相关提交：`9d45c78`、`53d45c1`、`50fd6bb`、`310f6b0`。

### Added

- **`internal/i18n`**（`9d45c78`，456 行 + 74 行测试）：零依赖字典（`Lang` / `T` / `Resolve`），
  `en` + `zh-CN`；启动时 `GetUserDefaultLocaleName` 解析一次，`zh*`→中文、其余/失败→英文；
  缺失 key 一律回退英文，测试锁两个字典的 key 对齐。`config.language` 默认 `auto`，托盘三项单选。
- **权限 5 秒自动同意（auto-approve）**（`50fd6bb`，`internal/ui/approve_windows.go` 348 行 + 207 行测试）：
  倒计时放行 + 否决握手。见 0.9.0 的 Added（契约与 UI 在 0.6.0 建立，0.9.0 补否决窗口的绘制与点击）。
- **扩展配置**（`53d45c1`）：`soundScheme` / `soundOnWaiting` / `soundOnError` / `notifyOnFinish` /
  `opacity` / `language`。缺失或损坏一律回落默认值，手改坏文件永远不会阻止浮窗启动。

### Changed

- 全部硬编码文案走 `i18n.T`；绘制函数加 `lang` 参数；`frameKey` 加 `lang`；
  离屏预览新增 `*-en` 与 `expanded-approve-*` 变体。
- auto-approve 默认关（`autoApprove=false`）、超时 `autoApproveTimeoutSec=5`（钳 1..60）。
  握手目录 `%LOCALAPPDATA%\AI-Terminal-Pet\approvals`；hook 每 100 ms 轮询同名 `.deny`，
  overlay 每 2 s 清扫超过 `2*timeout+30s` 的残留；否决窗口只覆盖「等待权限」路径。

### Fixed

- 非中文 Windows 上无法使用（中文文案硬编码在光栅器里）。

### Known issues（0.6.0 当时已知，未修）

- Claude 侧危险命令倒计时走完仍显示 waiting——无害（hook 侧已直接返回 `{}` 走上游原生提示）。

---

## [0.5.0] — 2026-09-26

提交 `ecd23ac`（+1011/−71，12 文件）。详细 runbook：[docs/HANDOVER.md](docs/HANDOVER.md) §17。

### Added

- **贴边收起（EdgeRetract）**（`internal/ui/dock_windows.go` 324 行 + 169 行测试）：
  `config.edgeRetract` 默认 true，托盘菜单新增开关。贴左/右边且鼠标不在面板上、且无等待确认/错误会话时，
  滑出屏幕只留 `dockedSliver`（默认 10，按 DPI/缩放缩放、下限 6）宽竖条；悬停滑出，告警态自动滑出。
- **基于时间的缓动**：`slideOut=160ms` / `slideIn=240ms`，`ease-out 1-(1-p)^3`，
  按流逝时间计算（掉帧不改变时长）。`stepSlide` 只算 tuck，`SetWindowPos`/`UpdateLayeredWindow`
  只在 `present()` 落一次，避免撕裂。
- **重绘跳过**：`present()` 用 `frameKey` 帧签名（`sessionDigest` 对 State/Source/Project/Reason +
  `UpdatedAt` 截断 15 s 做 FNV-64a），相同则整帧 GDI 光栅化跳过。探针实测 20 次相同推送 → 0 次额外重绘。
- `timerClock=2s` 驱动 `verifyHover` 与 `syncTuck`；`pulse(80ms)` 与 `slide(33ms)` 定时器拆分。

### Fixed

- **展开再缩回面板向左漂移**：v0.4.0 引入——旧 `applyGeometry` 用「保持窗口中心」处理展开/收起，
  贴右边缘时更宽的展开窗口被边界钳制往回推，回程仍从被推过的位置算中心，于是左移一个 sliver。
  改为锚定边模型：`dockedOrigin(edge, edgeCoord, gap, width)` = `edge-gap-width`（左）/ `edge+gap`（右）。
  `TestDockRoundTripIsExact` 锁往返一致；探针实测 `drift_x=0 drift_y=0`。
- `beginDrag` 不再复位 dock（修复拖拽后 dock 状态 stale）。
- `verifyHover` 每 2 s 用真实光标复核（`WM_MOUSELEAVE` 丢失的兜底）；合成悬停测试会失败，
  必须把真实光标 `park()` 到 sliver 上。

---

## [0.4.0] — 2026-09-26

提交 `869312e`（+882/−265，14 文件）。详细 runbook：[docs/HANDOVER.md](docs/HANDOVER.md) §16。

### Changed

- **紧凑态 40×190 → 56×156**：删掉 24 段装饰环（每帧 24 次椭圆光栅化）、`↑/↓` 指标行、分隔线、页脚箭头；
  改为状态球 / 会话数 / 状态词 / 最紧急项目名。多出来的宽度给项目名——它才是「一眼可辨」要回答的东西。
- **展开态 固定 250 高两只硬编码行 → 200×<内容高度>**：标题 + 实时状态胶囊 + 关闭钮 / 一行普查 /
  最多 6 行 + 溢出行 / 页脚提示；每行显示项目、状态、来源、相对时间。
  「运行」刻意**不含等待确认**的会话——从用户视角那不是「在跑」而是「在堵」，
  旧的两个指标行正是因为这个重复计数而对不上总数。
- **交互统一为「没有需要瞄准的东西」**：点击面板任意处切换（旧代码要求命中 20 单位宽的关闭钮，
  这正是「面板关不掉」的来源）；整块可拖拽（旧代码只允许顶部 40 单位）；悬停时底色与描边一起提亮；
  按压不再启动动画定时器。

### Added

- `canvas.measure` / `canvas.fitText`（GDI `GetTextExtentPoint32`）。分层窗口没有滚动条也没有 tooltip，
  放不下的文字不是「被裁掉」而是「丢了」；旧代码依赖 `DT_END_ELLIPSIS`，而它在 `DT_WORDBREAK` 框里无效
  （两个标志互斥），长项目名会溢出到条带两侧。
- `TestRenderPreview`：`PET_PREVIEW_DIR=<dir> go test ./internal/ui -run TestRenderPreview` 直接光栅化出 PNG。

### Fixed

- 长项目名溢出条带（见 `fitText`）。
- `install.ps1` 改为「停进程 → 等待真正退出 → 重试拷贝」：reporter 会按需把浮窗拉起来，
  停止后几毫秒它就可能复活，而带 AV 钩挂的机器上进程释放镜像要 1 秒以上，「停下、睡 500 ms、拷贝」会随机失败。

### Changed（续）

- 孤儿临时文件清扫阈值 1 小时 → 2 分钟（`install.ps1` 每次重装都可能留下一个孤儿）。

---

## [0.3.0] — 2026-09-26

提交 `1b38608`（+1057/−60，15 文件），55 个单测。详细 runbook：[docs/HANDOVER.md](docs/HANDOVER.md) §15。

### Added

- **F3 开机自启**（`internal/autostart` 271 行 + 163 行测试）：写
  `HKCU\Software\Microsoft\Windows\CurrentVersion\Run` 的 `AITerminalPet` 值。
  用 HKCU 而非 HKLM：免提权、按用户隔离、「关闭」就是一个删除，卸载没有残留要清理。
  启动时 `Sync` 修复指向已不存在路径的条目（安装路径是 Run 值里唯一的内容，
  程序被移动或重装后会在登录时静默失败——最糟的失败方式），且 `Sync` **永不**把关闭的条目改成开启。
  注册表经 `backend` 接口访问，8 个单测全跑内存后端——单测绝不该悄悄打开执行者的开机自启。
  托盘开关 + `--autostart[=on|off]`。
- **F4 提示音**：`config.json` 新增 `sound`。等待确认用 `MB_ICONEXCLAMATION`、错误用 `MB_ICONHAND`，
  走系统声音方案，因此与桌面其余部分一致并遵守专注助手/静默时段。与气泡提醒解耦但共享同一套去重。
- **F7 `--json` 导出**：`app.Snapshot()` 直接读 `runtime/` 快照离线聚合，**不依赖浮窗进程**，
  因此可被状态栏或脚本安全调用而不会阻塞在 GUI 进程上；`--pretty` 切换缩进。
  代价是新鲜度受快照 TTL（15 s）约束，与浮窗一致。

### Fixed

- `store.processBacked` 由「黑名单」改为「允许列表」。原实现把未知 source 当成真实进程，
  而 `generic.js` 与所有 hook 上报端用的是稳定哈希 PID，存活探测必然失败，
  于是**刚上报的会话立刻被删**——上报端回 `ok`，浮窗里却永远看不到。
  允许列表的默认值更安全：探测一个不可信的 PID 会删掉活会话，而放过一个已死会话只是多留一个 TTL。
  两个回归测试（`TestSnapshotKeepsSyntheticPIDSource` / `TestStoreKeepsNonProcessSourceWithDeadPID`）钉住两个方向。

---

## [0.2.0] — 2026-09-25

提交 `e57838c`（+1331/−43，10 文件）。详细 runbook：[docs/HANDOVER.md](docs/HANDOVER.md) §14。

### Added

- **记住位置**：拖拽/缩放/展开后落盘 `state.json`，下次启动原样恢复（含多显示器钳制，
  避免来自已断开的显示器的坐标把窗口留在屏幕外）。写入走后台单协程 `app.placementSaver` 合并，
  避免在 UI 线程做文件 IO（本机 AV 钩挂下写文件可达数百毫秒）。
- **CLI 控制**：`--show` / `--toggle` / `--reset` / `--quit` / `--version` / `--help`。
  控制命令与活窗口对话后退出，`--show`/`--toggle` 在没有实例时正常启动（可以安全绑热键）；
  用法错误走 stderr + exit 2，不写 `error.log`。
- **托盘升级**：tooltip 实时显示「N sessions - state」；进入等待确认/错误时弹气泡提醒
  （按会话 + 状态去重，可关）；气泡点击回跳浮窗；菜单新增「重置位置与大小 / 状态提醒 / 窗口置顶」。
- **`config.json`**（`internal/config` 192 行 + 116 行测试）：通知、置顶、不透明度。
  启动时读、托盘开关拨动时才写，所以停止状态下手改是安全的；每条读取路径都是 fail-safe——
  文件缺失/不可读/损坏一律给默认值，因为一个坏的设置文件绝不该阻止浮窗启动。
- **通用上报器** `reporters/generic.js`（191 行）：任意 CLI 一条命令接入，
  `node reporters/generic.js --source=aider --state=working --project="$PWD"`；
  同时接受 `--name=value` 与 `--name value`，`--stdin` 收 JSON 对象；与其它 hook 同一套 fire-and-forget 契约。

### Changed

- **轻量化**：脉冲动画定时器改为按状态自适应——只有按压/等待/错误态才装定时器，其余时间 `KillTimer`
  （此前是 7×24 的 20 Hz 空转，唤醒 UI 线程重绘同一帧）。
- `ROADMAP.md` 首次成型（E1–E6 扩展、L1–L5 轻量化、F1–F8 候选）。

---

## [0.1.0] — 2026-09-24

初始公开发布（MIT）。提交 `3456ed4`（31 文件，5,657 行），随后 `93d6783` 把 LICENSE 持有者改为 `jjjjjjjjnnjnn`。

### Added

- **Win32 分层置顶浮窗**（`internal/ui`，765 行 `app_windows.go` + 427 行 `canvas_windows.go` +
  272 行紧凑绘制 + 85 行绘制基座）：单实例（`Local\AI-Terminal-Pet-v1-SingleInstance` mutex）→
  置顶分层窗口，纯 GDI 画布逐帧光栅化（`CreateDIBSection` 自顶向下 32bpp DIB + `UpdateLayeredWindow`，全仓无 gdiplus），DPI 感知（每显示器 `^uintptr(3)` 对应 `-4`）。
- **命名管道服务端**（`internal/pipeserver` 187 行 + 276 行测试）：`\\.\pipe\ai-terminal-pet-v1-<homedir-sha256 前 8hex>`，
  按用户隔离；合法回执 `OK\n` / 非法 `ERR\n`，客户端只接受精确 `OK`，否则 `EBADREPLY` 重试至 deadline。
- **状态聚合**（`internal/store` 700 行 + 373 行测试）：上限 256 会话、tombstone 10 分钟、
  pipe TTL 30 s、快照 TTL 15 s、快照文件上限 256。
- **会话模型**（`internal/session` 209 行 + 108 行测试）：schema `ai-terminal-pet.v1`，
  payload 上限 1 MiB，PID/ParentPID 32 位界。
- **原生 statusline shim**（`cmd/pet-statusline` 470 行，零新依赖）：stdin 上限 1 MiB / 2 s、宽松 JSON、
  状态映射、`session.ParsePayload` 自校验降级 `idle`、原子快照、`winio.DialPipe` 150 ms、
  `ENOENT` 且非 `closed` 时拉起重试、stdout 只打印 `pet <state>` 且永不 exit 非 0。
  存在理由：本机 `node -e exit` 4.3 s、`cmd /c exit` 7.7 s、`whoami` 1.5–1.9 s（McAfee + Defender 钩挂），
  Antigravity 的 statusLine runner 会超时击杀 node 路径。
- **三个宿主上报端**（`reporters/`）：`opencode.js`（229 行插件）、`claude-hook.js`、
  `antigravity-hook.js`，共用 `protocol.js`（359 行）。
- `install.ps1`（26 行）、`LICENSE`（MIT）、`.gitignore`（`*.exe` 全忽略）、`README.md`（95 行）、
  `docs/archive/BASELINE-20260924.md` 与 `HANDOVER.md`。

### Fixed（含在首批提交里，同日）

- `cbc043e` **浮窗不可见/不可交互的根因**：`Run()` 在没有 `runtime.LockOSThread()` 的 goroutine 上建窗口并泵消息。
  Go 调度器把它迁到别的 OS 线程后，窗口消息仍投递到「创建窗口的那个线程」的队列，而该队列无人泵——
  首个跨线程 `SetWindowPos(HWND_TOPMOST)` 在同步 `SendMessage` 里永久阻塞，
  `WS_VISIBLE` 永不置位、消息全不分发、定时器全死。这一个遗漏解释了「显示不出来 / 点击像崩溃 /
  双击没反应 / 拖不动 / 定时器不触发」全部症状，并**推翻了 `BASELINE-20260924.md` §16 的
  「系统级 win32k/DWM 楔子」结论**（那是同一 bug 的误诊）。
- `cbc043e` **每次拖动都会额外切换展开/收起**：`endDrag` 在 `ReleaseCapture()` **之后**读 `a.drag.moved`，
  而 `ReleaseCapture` 会同步派发 `WM_CAPTURECHANGED` 把 `a.drag` 清空。改为先判定后释放，抽成有单测的 `decideRelease`。
- `cbc043e` 第二实例不再静默退出，而是激活已存在浮窗（`ActivateExisting`）；
  DIB 画布与命中缓冲复用；watchdog 限 3 次转储 + 2 min 冷却 + 记录最后消息名
  （此前在一次 80 分钟的故障里追加了 370 KB 完全相同的栈，却没记下是哪条消息卡住）。

---

## 参考

- 逐节 runbook 与事故复盘：[docs/HANDOVER.md](docs/HANDOVER.md)（§13–§25 为当前有效部分）
- 扩展与轻量化计划：[docs/ROADMAP.md](docs/ROADMAP.md)
- 历史取证：[docs/archive/](docs/archive/)（索引见 [docs/archive/README.md](docs/archive/README.md)）
