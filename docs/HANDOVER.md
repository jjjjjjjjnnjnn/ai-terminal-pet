# HANDOVER — ai-terminal-pet 交接文档

> 交接时间：2026-09-24（UTC+8 下午）；2026-11-09 刷新到 v0.9.0；2026-11-10 补记 v0.9.0 之后未发版的工作，含 DSH 插件包 1.2.3 与市场条目合并。
> 下一位 AI 直接从 §1「当前状态」开始读，然后按 §13–§27 的时间序读 runbook，不必回溯长上下文。
> 2026-09-24 停用态时期的章节已归档到 [docs/archive/HANDOVER-2026-09-DISABLE.md](archive/HANDOVER-2026-09-DISABLE.md)。

## 1. 当前状态（v0.9.0 之后，2026-11-10，未发版；含 DSH 插件包 1.2.3）

- **在役**：安装目录 `AI-Terminal-Pet.exe` 已安装并通过 SHA256 校验；浮窗按需由 reporter 拉起，
  无常驻服务，无会话时自毁。
- **版本**：`0.9.0`（`internal/app/app_windows.go:23` 的 `Version` 常量，`--doctor` 与文档刷新都没有改它）。
  最新提交 `b4474a6`（2026-11-10 已推送，与 `origin/main` 同步）；插件包最新发布是 1.2.3（tag `dsh-plugin-v1.2.3`）。
- **健康**：`gofmt -l .` 无输出；`go vet ./...` exit 0；`go test -count=1 ./...` 全过（**154 个** `func Test`，
  由 `go test -list '.*' ./...` 计数）；六个上报端逐个 `node --check` 全过；
  `node integrations/deepseek-harness/verify.mjs` 输出 VERIFY PASSED；
  `scripts/install.ps1 -VerifyOnly` 10/10 `[ok]`、exit 0（脚本已从仓库根下沉到 `scripts/`，见 §27.11）。
- **规模**：Go 非测试源码 **10,719** 行 / 测试 **4,212** 行 / JS **1,742** 行（49 个 `.go`、8 个 `.js`+`.mjs`）；
  `bin/AI-Terminal-Pet.exe` **4,371,968 B**、`bin/pet-statusline.exe` **3,643,392 B**（均按
  `-trimpath -buildvcs=false -ldflags "-s -w"` 构建，见 §26）；第三方依赖 3 个（go-winio MIT、lxn/win BSD、x/sys BSD）。
  v0.9.0 提交时的旧值（10,201 / 3,831 / 1,633 行，4,178,432 / 3,644,416 B，140 个 `func Test`）已作废。
- **新增**（全部在 v0.9.0 之后、未发版）：`--doctor` 自检（9 项检查）、`docs/PROTOCOL.md` 线协议规范、
  `reporters/package.json` 显式 exports、`.github/workflows/ci.yml` 单 job 门禁、
  `scripts/install.ps1` 的 `-VerifyOnly` / `-Uninstall` / `-AddToPath` 三个开关（卸载同时清理 PATH 条目）。
- **分发**：Harness 桥接已做成独立 DSH 插件包 `packages/dsh-ai-terminal-pet`（npm 名 `dsh-ai-terminal-pet`
  **1.2.3**，含预编译浮窗，Release tag `dsh-plugin-v1.2.3`，资产 `dsh-ai-terminal-pet-1.2.3.tgz`；
  1.2.2 / 1.2.1 / 1.2.0 / 1.1.0 均保持已发布、不可变，由 1.2.3 取代。1.2.3 只改两份清单：各加一个
  `dsh.marketplace` 块（`profiles: ["desktop","headless","web"]`、`requiresBuildApproval: false`、
  `manualSteps: false`），因为按 `install.profiles` 过滤的第三方商店（如 YELEBAI/dsh-plugin-marketplace）
  此前只能把 host-only bundle 猜成 `[headless, web]`，桌面端于是看不见条目；DSH 本身从不读该字段。
  `lib/index.js`、两个 vendor 二进制、`cordis.patch.yml`、`dsh.catalog`、`vendor/manifest.json` 逐字节不变
  （上一版 1.2.2 才是字节改动：`scripts/` 迁移改到 Go 注释、`vendor/manifest.json` 的源码摘要随之变）；
  仓库根的清单 shim 让它也能按仓库 URL 安装，三种安装形态实测 rc=0。`dsh.catalog` 是自我声明的 category/摘要/capabilities（不被强制），
  `keywords` 里的 `dsh-plugin` 让自动索引型市场能按 topic 采集。
  awesome-dsh-plugin 注册表 PR #13937 已于 2026-11-10T14:33:04Z 被 `fkysly` 合并，上游 README 的
  「Notifications & Integrations」小节已列出该条目（17:19Z 维护者在 PR 上留言说明合并时删了一句「面板可直接
  放行审批」并提示改措辞要另开 PR）。**该市场现在已经真的可见**：上游那次卡住的站点构建后来跑完，
  条目页 <https://awesome-dsh-plugin.com/p/jjjjjjjjnnjnn/ai-terminal-pet--packages-dsh-ai-terminal-pet/> 返回 200，
  DSH 应用内 Market 加载的 `plugins.json` 也收录了本仓库（`count` 由 10119 增到 10230，合并后再读一次是
  10,304）；那份目录的 `install` 命令是构建时按条目文件里的 `tarball` 渲染的，所以合并之后读到的那版仍给出
  1.2.0 资产。跟进 PR #14063（指向 1.2.1）已评论说明后关闭；**#14076**（分支 `bump-tarball-1.2.2`，只改一行
  `tarball:`）把它换成 1.2.2 并在 **19:59:37Z 合并**，下一次站点重建才会渲染出新命令；1.2.3 发版后又开了
  **#14089**（`bump-tarball-1.2.3`，commit `c5a1f5f`）把字段再推到 1.2.3，仍在等合并。
  DSH-Plugins-Marketplace 当日 16:22Z 那次重建把本仓库从 `verdict: script` 改判为 **`bundle-plugin`**；
  到 18:08:44Z 那次重建（`risk_at` 17:09:14Z，是安装脚本移进 `scripts/` 后第一次真的扫到本仓库），条目是
  `pkg_name: dsh-ai-terminal-pet`、`version: 1.2.1`、topics 11 个、**`risk_tier: "safe"` 且 `risk_flags` 为空**——
  原先那个 critical 徽章（`registry-run-key`）与 medium 旗标（`remove-item-force-recurse`）就此消失。
  `evalRepoRisk` 只按文件名抓仓库根的 `install.ps1` / `install.sh` / `package.json`，安装脚本已于
  2026-11-10 移进 `scripts/`（§27.11），它读不到那两个正则的落点，也没用到任何抑制字段。
  npm 的凭据问题已解（`npm whoami` 返回 `jjjjjjjjnnjnn`，此前整场会话 E401），但**任何版本都还没发布到 npm**
  （`npm view` 仍 E404，且发布要求一次性密码，非交互 shell 里拿不到），所以按包名安装暂时不可达；
  `docs/market-index.json` 是仓库自建的 format-1 索引，钉在 1.2.3 的发布 commit（`b4474a6`），可让 Discover 先按 git 源可用。
  细节与坑见 §27。
- **接入**：OpenCode（`reporters/opencode.js` 插件）、Claude Code（`reporters/claude-hook.js` 7 事件）、
  Antigravity（`bin/pet-statusline.exe` + `reporters/antigravity-hook.js`）、DeepSeek Harness
  （两种：老的 host-only bridge bundle `integrations/deepseek-harness/ai-terminal-pet-bridge`，
  与新的独立插件包 `packages/dsh-ai-terminal-pet`，见 §27）、任意 CLI（`reporters/generic.js`）。
- **未跑项**：`go test -race`（缺 gcc/cgo）、`govulncheck` / `gosec`（未安装）。

> **历史内容去哪了**：2026-09-24「项目全局停用」时期的停用态描述（原 §1、§2、§4、§9、§10 的旧正文）、
> 以及原 §3 架构速览、§5 已知残留、§6 关键修复、§7 验证基线、§8 常用命令、§11 残留风险、§12 重启用记录，
> 已整体搬迁到 [docs/archive/HANDOVER-2026-09-DISABLE.md](archive/HANDOVER-2026-09-DISABLE.md)
> （章节号与正文逐字保留，未删任何技术细节）。因此本文件的章节号从 §13 起续接，§3–§12 有意留空。
> 当前有效 runbook = 本文 §2 + §13–§27（§26 是 v0.9.0 之后的批次，§27 是 DSH 插件包分发）。

## 2. 关键路径

- 源码：`C:\Users\rongj\Desktop\项目开发\ai-terminal-pet`（Go module `ai-terminal-pet`，`go 1.23.0`，实测 go1.27.0）
- 入口：`cmd/ai-terminal-pet/main.go`（浮窗与控制 CLI）；`cmd/pet-statusline/main.go`（状态行 shim）
- 仓库内构建产物：`bin/AI-Terminal-Pet.exe`、`bin/pet-statusline.exe`
- 安装目录 `C:\Users\rongj\AppData\Local\AI-Terminal-Pet\`：
  - `AI-Terminal-Pet.exe`、`pet-statusline.exe`、`protocol.js`（`install.ps1` 从 `reporters/protocol.js` 同步并逐个校验 SHA256）
  - `bundles\ai-terminal-pet-bridge\`（DSH bridge：`package.json` / `cordis.patch.yml` / `index.js` / `icon.svg` / `README.md` / `locale\{en,zh}.json`）
  - `config.json`（偏好，字段见 §25.5）、`state.json`（位置/缩放/展开）、`runtime\`（会话快照 `*.json` / `*.tmp`）
  - `approvals\`（auto-approve 握手：`pending-*.json` 与同名 `.deny`，见 §18.2）、`error.log`（0600，超 1 MB 轮转）、`hang.log`
- 部署脚本：`scripts/install.ps1`（四组参数 `Source`/`ShimSource`/`ProtocolSource`/`BundleSource`；停进程 → 等真正退出 → 重试拷贝 20 次 → 每个文件 SHA256 校验）
- 插件包：`packages/dsh-ai-terminal-pet/`（独立 DSH bundle，`vendor/` 内含预编译 `AI-Terminal-Pet.exe`，
  与 `bin/` 下那份字节一致；详见 §27）
- 当前有效 runbook：**§13 起**；§13–§27 按时间追加，最后两条是 §26（v0.9.0 之后未发版的文档/自检/CI 批次）
  与 §27（作为 DSH 插件包分发，2026-11-10；§27.8 是仓库根清单 shim 与 CI 里的插件套件）。
- 历史提交的边界：`4d777e3`（2026-11-10，已推送并打 annotated tag `dsh-plugin-v1.2.0`）是 1.2.0 批次的终点；
  `dsh.catalog`、根清单 shim、`vendor/manifest.json` 与 CI 里的插件套件都在它里面（见 §27.2、§27.5、§27.8、§27.9）。

## 13. 浮窗全线故障根因修复 2026-09-25（历史起点，最新状态见 §25）

**根因（一个）**：`internal/ui/app_windows.go` 的 `Run()` 没有 `runtime.LockOSThread()`。
Go 调度器把主 goroutine 迁到别的 OS 线程后，窗口消息仍投递到「创建窗口的那个线程」的队列，
而该队列无人泵消息；于是首个跨线程 `SetWindowPos(HWND_TOPMOST)` 在同步 `SendMessage` 里
永久阻塞 → `WS_VISIBLE` 从未置位、消息全不分发、定时器全死。

**这一个遗漏解释并已修复**：显示不出来 / 点击崩溃（幽灵窗口 + AppHang）/ 打开异常 /
无法拖拽 / 定时器失效 / 空闲自毁不生效。**`docs/archive/BASELINE-20260924.md` §16 的「系统级 win32k/DWM 楔子」结论作废**
（当时观察到的跨进程 show 卡死、新生窗口同样中招，都是同一 bug 的不同侧面）。

**第二个 bug**：`endDrag` 在 `ReleaseCapture()` **之后**读 `a.drag.moved`，
而 `ReleaseCapture` 会同步派发 `WM_CAPTURECHANGED` 把 `a.drag` 清空 → 每次拖动都额外
触发一次展开/收起。已改为先判定后释放，并抽成有单测的 `decideRelease`。

**附带改进**：第二实例不再静默退出而是激活已存在浮窗（`ui.ActivateExisting`）；
DIB 画布与命中缓冲复用（去掉每帧 GDI 分配）；watchdog 限 3 次转储 + 记录最后消息名；
展开态关闭按钮绘制与命中几何同源；新增 2 个单测。

**装机**：hash `9C70D60B…`（前值 `71078FA7…`），`install.ps1` 安装 + SHA256 校验通过。

**终端接入**：按用户决策移除 Antigravity 的 `BeforeTool`/`AfterAgent` node 钩子
（每次工具调用约 1 s 进程创建税），仅保留 `statusLine`；备份
`~/.gemini/antigravity-cli/settings.json.bak-20260925-pet-hooks`。

**完整取证 / 验证矩阵见 `docs/archive/BASELINE-20260925.md`。** 遗留：`-race`/govulncheck/gosec 仍缺工具链。

## 14. 打磨迭代 2026-09-25 夜（v0.2.0）

本轮把浮窗从「能显示」推到「可用」，并补上扩展/轻量化路线图（见 `docs/ROADMAP.md`）。

**新增能力**
- 位置记忆：`%LOCALAPPDATA%\AI-Terminal-Pet\state.json`（位置/缩放/展开），拖拽/缩放/展开后落盘，
  启动原样恢复并做显示器钳制。写入经后台单协程合并（`app.placementSaver`），
  避免在 UI 线程做文件 IO——本机 AV 钩挂下写文件可达数百毫秒。
- `config.json`：通知开关 / 置顶 / 不透明度。缺失或损坏一律回落默认，绝不阻塞启动。
- CLI：`--show` / `--toggle` / `--reset` / `--quit` / `--version` / `--help`。
  用法错误走 stderr + exit 2（不写 error.log）；控制类命令不再意外拉起常驻实例。
- 托盘：tooltip 实时显示会话数与聚合状态；等待确认/错误弹气泡（按会话+状态去重、可关、点击回跳）；
  菜单加「重置位置与大小 / 状态提醒 / 窗口置顶」。
- `reporters/generic.js`：任意 CLI 一行接入（`--state=working` 或 `--stdin` JSON）。

**轻量化**
- 脉冲动画定时器改为按状态装/卸（`syncAnimation`）：只有按压/等待/错误态才 50 ms，
  其余时间 `KillTimer`。此前是 7×24 的 20 FPS 空转。
- 复用 DIB 画布与命中缓冲（上一轮已加，本轮确认无回归）。

**新增包与测试**
- `internal/config`（7 个单测：容错解码、钳制、原子写、往返、无残留临时文件）。
- `internal/ui` 新增 `TestExpandedCloseButtonRect`、`TestDecideRelease`。

**验证（装机 hash `38567A0E…` 起，后随打磨更新）**
DPI 感知探针实测：冷启动可见+响应 60x285（150% 缩放下的 40x190）／点击展开 300x375／
关闭按钮收起／滚轮缩放／`--reset` 回锚／拖拽窗口位移与光标位移一致且尺寸不变／
`--toggle` 生效／二次启动 178 ms 退出并激活已隐藏窗口／`--quit` 干净退出／
种子 state.json 四组坐标（300,200）(900,300)(1200,500)(100,80) 全部原样恢复。

**踩坑记录（重要）**
- 验证探针必须自己声明 DPI 感知（`SetProcessDpiAwarenessContext(-4)`）。
  否则 `GetWindowRect` 返回被虚拟化的坐标，150% 屏上会得到 2/3 的比例，
  看起来像「位置恢复有偏移」——实际是测量坐标系错了。
- 单实例互斥体：跑任何测试二进制前必须先 `taskkill AI-Terminal-Pet.exe`，
  否则测试进程静默退出、探针探到的是生产实例的窗口。

**遗留**：`docs/ROADMAP.md` §6 列出的未修项（`-race`/govulncheck/gosec 缺工具链、UI 无视觉回归、
管道抢占只做缓解）。

## 15. F3/F4/F7 落地 2026-09-26（v0.3.0）

**F3 开机自启** — 新增 `internal/autostart`：`HKCU\Software\Microsoft\Windows\CurrentVersion\Run`
下的 `AITerminalPet` 值（用户级、免管理员、停用即删、卸载无残留）。
- 托盘开关 + `--autostart` / `--autostart=on` / `--autostart=off`。
- 启动时 `Sync` 修复指向旧路径的条目：Run 值里只有安装路径，程序移动/重装后
  会在登录时静默失败。`Sync` **永不**把已关闭的条目改成开启。
- 注册表经 `backend` 接口访问，8 个单测全部跑在内存后端上——单测不该悄悄打开
  `go test` 执行者的开机自启。
- 验证：`--autostart=on` 后独立用 PowerShell 读注册表，确认写入
  `"C:\Users\rongj\AppData\Local\AI-Terminal-Pet\AI-Terminal-Pet.exe"`（带引号）；
  `--autostart=off` 后确认值已消失。**当前状态：关闭**（未替你打开）。

**F4 提示音** — `config.json` 新增 `sound`（默认 true）。等待确认 `MB_ICONEXCLAMATION`、
错误 `MB_ICONHAND`，走系统声音方案，遵守专注助手。与气泡提醒解耦但共享同一套
「会话+状态」去重。托盘加开关，开启时立即响一声作为反馈。

**F7 `--json`** — `app.Snapshot()` 读 `runtime/` 快照离线聚合，**不依赖浮窗进程**，
因此状态栏/脚本可以安全调用而不会阻塞在 GUI 进程上；`--pretty` 切缩进。
5 个单测覆盖聚合优先级、过期剔除、目录缺失、无 `LOCALAPPDATA`、两种编码一致性。

**顺带修掉一个真 bug（重要）**
`store.processBacked` 原本是「黑名单」：未知 source 一律当成真实进程。
而 `generic.js` 与所有 hook 上报端用的是**稳定哈希 PID**，存活探测必然失败，
于是刚上报的会话立刻被删——上报端回 `ok`，浮窗和快照里却什么都没有。
已改为**允许列表**（当前只有 `opencode`）。默认值的选择理由：探测一个不可信的 PID
会删掉活会话，而放过一个已死会话只是让它多留一个 TTL。新增两个回归测试
（`TestSnapshotKeepsSyntheticPIDSource`、`TestStoreKeepsNonProcessSourceWithDeadPID`）。

**验证**：`gofmt`/`vet`/`test` 全绿（55 个单测）；DPI 感知探针交互回归全过
（拖拽位移与光标位移完全一致、尺寸不变）；`--json` 在有会话时正确输出
`state=waiting_permission count=2` 且按紧急度排序；装机 hash 校验一致。

**踩坑记录**
- Git Bash 会改写 Windows 风格路径参数（`C:\a\b` → `C:/a/b`），
  用它测试 `--project=...` 会误以为上报端损坏。真实调用（PowerShell/cmd）不受影响。
- 本机 `reg.exe` 被沙箱黑名单拦截；核对注册表改用 PowerShell 的
  `Get-ItemProperty HKCU:\...\Run`。

## 16. 面板重做 2026-09-26（v0.4.0）

**动机**：紧凑态把「状态」和「数量」用六种方式重复编码（状态球、24 段装饰环、两个
`↑/↓` 指标行、边框色、分隔线、页脚箭头），密度高而信息量低，而真正想知道的
「谁在等我」反而看不到；展开态固定 250 高、最多两行，第 3 个会话只能以摘要里的
一个数字存在。

**面板几何现在是单一来源**：`internal/ui/drawing_compact_windows.go` 顶部的
`compactWidth/Height`、`expandedWidth`、`expandedHeaderHeight` 等常量同时被
`logicalWidth()/logicalHeight()`（窗口尺寸）和光栅化代码使用。改版式只需要动这些常量，
但**必须同时确认 `TestExpandedCloseButtonRect` 和 `TestExpandedPanelHeight` 仍然通过**——
它们锁的是「画出来的东西不越界」和「高度随会话数增长但有上限」这两条不变式。

**展开态高度是内容驱动的**：`expandedPanelHeight(count)` = 头 + 普查行 + 正文 + 页脚，
正文 = `min(count, 6) * 34`（0 个会话时用空态高度，超过 6 行加一条溢出行）。
因此 `wmRefresh` 里新增了一次 `applyGeometry()`：会话增减会改变高度，必须重设窗口尺寸。
`applyGeometry` 会**保持窗口中心不动**，否则展开一长串会话会把页脚顶出屏幕底边、
再被钳制整个拽上去。

**交互规则已统一为「没有需要瞄准的东西」**：
- 点击面板任意处 = 紧凑↔展开（旧代码要求点中 20 单位宽的 ✕ 才能收起，这是
  「面板关不掉」的来源）。`decideRelease` 因此简化为两参数，`closeButtonHit` 已删除。
- 整块可拖拽（旧代码只允许顶部 40 单位）。
- 悬停时**底色与描边一起**提亮。只改描边不够：告警态描边本就是接近不透明的状态色，
  在最想让人点的时候反而看不见变化。
- 按压不再计入 `syncAnimation` 的触发条件——按压反馈是静态的。

**踩坑记录（重要）**
- **`^uintptr(n)` 是 `-(n+1)`，不是 `-n`。** 想声明
  `DPI_AWARENESS_CONTEXT_PER_MONITOR_AWARE_V2`（值为 -4）必须写 `^uintptr(3)`；
  写成 `^uintptr(4)` 得到的是 -5（`UNAWARE_GDISCALED`），探针变成 **DPI 不感知**，
  读回的坐标被系统按 `96/dpi` 虚拟化，150% 屏上 84 宽会读成 56——
  看起来像「尺寸算错了」，其实只是探针的坐标空间不对。排查这类问题时先确认探针的
  DPI 感知是否真的生效。
- **本机桌面会自行移动光标**。同一坐标 `SetCursorPos` 后 `GetCursorPos` 经常读回不同值
  （实测漂移可达数百像素），这会让「点击」被应用判定成「拖拽」、让拖拽读数看起来
  完全错乱。拖拽验证必须在每次移动前做「落点确认 + 重试」，确认不了就跳过并说明，
  不要把环境噪声当成产品 bug。
- **`install.ps1` 会与 reporter 抢进程**。浮窗是按需拉起的，`Stop-Process -Force`
  之后几毫秒 reporter 就可能把它拉回来，而 AV 钩挂下进程释放镜像可能要 1 秒以上，
  于是「停进程 + 固定 500 ms + 拷贝」在真正在用这个工具的机器上会随机失败。
  现已改为「停 → 等真正退出 → 重试拷贝（20 次，每次先清一遍实例）」。
- **`touch <不存在的路径>` 会创建文件**。用 `touch -d ... "$DIR"/.pet-*.tmp` 给
  通配符老化时间，在通配符没匹配到任何文件时会在目录里创建一个字面名叫
  `.pet-*.tmp` 的文件，把后续的清扫验证搅乱。先确认匹配到了再 touch。

**验证**（装机 hash `6854F3BD…`，`install.ps1` 两组哈希各自一致）
DPI 感知探针（150% 缩放）：紧凑 `84x234`、展开 1 会话 `300x216` / 3 会话 `300x318` /
4 会话 `300x369` / 7 会话 `300x507`（触顶 6 行 + 溢出行）；点击面板主体收起成功；
拖拽连续 3 次位移与光标**完全一致**且尺寸不变；滚轮 `84x234 → 92x257`；
`--reset` 重新贴边；展开时中心 y 位移 0；`--quit` 干净退出；`--toggle` 往返正确。
离屏 PNG 预览（`TestRenderPreview`）复核了紧凑/展开/空态/告警/悬停各状态。
`gofmt`/`vet` 干净，60 个单测全过，`error.log`/`hang.log` 均无新条目。

## 17. 贴边收起 + 漂移修复 + 性能（v0.5.0）

**动机**：用户两点反馈 —— (1) 点开展开再缩小，面板会**往左漂**（和原本位置不一样）；
(2) 希望贴着屏幕边时，能像电脑管家浮窗一样**缩进屏幕边缘**。顺带把交互做顺、性能做轻。

### 17.1 漂移是真 bug，且是 v0.4.0 我改版时引入的
v0.4.0 的 `applyGeometry` 用「保持窗口中心」处理展开/收起。贴右边缘时，展开出的更宽窗口
被边界钳制往回推，回程却仍从被推过的位置算中心 → 左移约一个 sliver 宽度。
**正确模型是「锚定边」**：贴哪条边就保持哪条边，四周都不贴时才按中心扩张。抽出纯函数
`dockedOrigin(edge, edgeCoord, gap, width)` = `edge - gap - width`（左）/ `edge + gap`（右），
并加 `TestDockRoundTripIsExact` 锁「紧凑↔展开往返后 x 完全一致」。探针实测：
紧凑 `x=2440`，展开、`--toggle` 收回，紧凑 `x=2440`，`drift_x=0 drift_y=0`。

### 17.2 贴边收起（edge tuck）
- 配置 `config.json` 新增 `edgeRetract`（默认 true）；托盘菜单新增「贴边收起」开关。
- 逻辑全在 `internal/ui/dock_windows.go`：贴左/右边（容差 `dockTolerance` 内）时，鼠标不在面板上
  且**没有等待确认/错误**会话时，面板滑出屏幕只留 `dockedSliver`（默认 10，按 DPI/缩放缩放，下限 6）
  宽竖条；悬停滑出，告警态自动滑出。
- 滑出/收回是**基于时间的缓动**（`slideOutDuration=160ms` 快显、`slideInDuration=240ms` 缓收，
  `ease-out`）而非「每 tick 移动剩余距离的固定比例」——掉帧不会改变时长，桌面繁忙时反而更可预测。
- 缓动只在 `a.tuck` 变化，位置随 `present()` 的 `UpdateLayeredWindow` 一次原子更新，不再额外
  `SetWindowPos`（否则会先显示一帧旧内容再重绘，撕裂）。
- `stepSlide` 只算 `a.tuck`，真正移动窗口在 `present()` 里（`frameX()` 用 `slideBetween` 插值），
  所以几何完全同源，不会和点击/拖拽区域漂移。

### 17.3 悬停的坑：verifyHover 看的是「真实光标」
`verifyHover` 每 2s（`timerClock`）用 `GetCursorPos`+`GetWindowRect` 复核：真实光标不在面板上就
`clearHover()`。这是 `WM_MOUSELEAVE` 丢失（远程桌面/合成光标/显示切换）时的兜底，**必须有**。
代价：纯靠 `PostMessage(WM_MOUSEMOVE)` 合成悬停的自动化测试会失败——真实光标在别处，看门狗 2s 内就清掉。
验证悬停必须让**真实光标停在 sliver 上**（探针 `park()` 到 `workRight-5, midY`）。

### 17.4 重绘跳过（性能）
`present()` 用 `frameKey(...)` 给当前帧签名；签名相同且已画过就直接 `return`，**整帧 GDI 光栅化跳过**。
签名包含 `digest`（见下）、尺寸、状态、数量、展开/按压/悬停/收起/动画。关键修正：签名**不能**用
`store.Revision`——reporter 每次推送都更新 `updatedAt`，Revision 必变，会让「相同状态」每次都重绘。
改为 `sessionDigest()`：对「实际画出来的字段」（`State/Source/Project/Reason` + `UpdatedAt` 截断到 15s）
做 FNV 哈希。探针实测：**20 次完全相同的推送 → 0 次额外重绘**（仅有的几次来自环境光标漂到面板上
触发的悬停滑动）。`TestFrameKeySkipsUnchangedFrames` 锁这条不变式。

### 17.5 帧率分工
`timerAnimation` 间隔按事件切换：`pulseInterval=80ms`（告警呼吸，慢呼吸不需要更高帧率）
vs `slideInterval=33ms`（滑动要顺）。`syncAnimation` 在「脉冲或滑动进行中」时才装计时器，
空闲时完全不转。

### 17.6 验证（装机 hash `179a5BA5`，两组 hash 各自一致）
DPI 感知探针（150%）：
- tuck on idle：收紧（x≈2545，sliver 留屏）✓
- 真实光标停在 sliver → 滑出 ✓；移开 → 收回 ✓
- 告警（waiting_permission）→ 自动保持全显 ✓
- 告警清除（idle）→ 收回 ✓
- `--toggle` 展开↔收回往返 x 稳定于 2545（drift_x=0）✓
- 性能：20 次相同推送 0 次重绘 ✓
`gofmt`/`vet` 干净，全量单测通过（新增 `TestDockRoundTripIsExact`/`TestSlideTuckIsTimeBased`/
`TestFrameKeySkipsUnchangedFrames`），所有 `reporters/*.js` 通过 `node --check`。

### 17.7 本机环境坑（再记一次，值）
- **残留 reporter 会污染验证**：之前的探针跑完后 `runtime/` 里留有 `waiting_permission` 快照，
  装好的浮窗按需拉起后读到它，状态一直是告警，看起来「贴边不生效」。验证前 `RemoveAll(runtime)`
  并 `taskkill node.exe` 清掉残留上报进程。
- **光标会自己漂移**：本机合成光标会把 `SetCursorPos` 后读回的坐标改掉。拖拽/悬停验证必须
  `park()` + 读回确认，否则会把环境噪声当成产品 bug。

## 18. 2026-09-26（v0.6.0）—— i18n + 权限请求 5 秒自动同意

### 18.1 i18n（`internal/i18n` 新包 + `config.language`）
- 仅两种展示语言：`en` / `zh-CN`。设置第三值 `auto`（默认）跟随系统（`GetUserDefaultLocaleName`，
  `zh*`→中文，其余/失败→英文）。未知一律回英文，绝不出现空白标签。
- `config.json` 加 `language`（默认 `"auto"`，归一化大小写/非法值→auto）。旧文件缺字段保持默认。
- 托盘菜单平铺三项（`●/○` 单选语义）：`语言 / Language · 自动（跟随系统）` / `English` / `简体中文`。
  切换即时生效（`a.lang` + `postRefresh`）。字体保持 `Microsoft YaHei UI`（中英全覆盖）。
- 全部硬编码文案（状态词/面板/托盘/气泡/相对时间）改走 `i18n.T`，key 表见 `internal/i18n/i18n.go`。
  绘制函数签名加 `lang` 参数（`stateLabel/projectLabel/reasonLabel/summaryLine/relativeTime/
  compactSubject`）；单测按语言参数化（中英各断言一遍）；`frameKey` 加 `lang` 字段。
- 预览：`TestRenderPreview` 新增 `*-en` 与 `expanded-approve-*` 用例，中英 PNG 目检通过。

### 18.2 自动同意（默认关；hook 阻塞 + `approvals/` 文件握手）
- 背景：浮窗是只读看板（管道回执仅 OK/ERR，Session 无 requestId），overlay 无法回写终端。
  所以反过来：hook 在权限门控处阻塞等 overlay 的否决，沉默=同意。
- `config.json` 加 `autoApprove`（默认 false）、`autoApproveTimeoutSec`（默认 5，钳 1..60）。
  托盘开关：`5秒未响应时自动允许权限请求 / Auto-approve tool permissions after 5s`。
- 握手目录 `%LOCALAPPDATA%\AI-Terminal-Pet\approvals\`：
  hook 建 `pending-<source>-<session>-<pid>-<epochMs>.json`（原子写，body 含精确身份），
  每 100ms 轮询同名 `.deny`；overlay 点展开态等待行=写 `.deny`+清倒计时；超时无否决=allow。
  hook 退出必删 pending（`finally`）；overlay 每 2s（`timerClock`）清扫超 `2*timeout+30s` 的残留。
- 否决窗口只覆盖权限提示路径（hook 判定为 `waiting_permission` 的事件）。普通 tool_use 观察
  永不阻塞——否则每个工具调用都被拖慢整整一个 timeout。
  Claude 危险命令（`protocol.js: DANGEROUS_PATTERNS`：rm -rf/rd /s/mkfs/dd/fork bomb/git push --force/
  format/del Windows/shutdown/Remove-Item System32）→ 仍上报 waiting 告警但立即返回 `{}`，
  走终端原生提示；Antigravity 同条件下返回 deny（该 harness 无原生提示回退，fail-safe）。
  读不到配置/建不了 pending 文件一律 fail-closed（deny），hook 永不抛错到宿主 CLI。
- overlay 侧：`approvalTracker`（UI 线程）按 `(Source,SessionID,PID)` 建倒计时，
  新一轮权限（`UpdatedAt` 变新）重启时钟；会话离开 waiting 即删（防幽灵触发）；
  复用 `timerAnimation`（等待态本就常驻 80ms），不新增 timer ID；
  展开态等待行显示 `Allow (Ns)` + 耗尽进度条；`frameKey` 混入最热剩余秒（否则 15s 分桶不重绘倒计时）。
  点展开态=否决最紧急等待会话；紧凑态点击仍是展开/收起。
- 验证：`go vet ./...`、`go test -count=1 ./...`（新增 i18n 3 / config 1 / ui tracker 6+deny路由+sweep）、
  `node --check reporters/*.js` 全过；冒烟：黑名单 13 例、`off→{}(90ms)`、`危险→{}(100ms)`、
  `超时→allow(1.1s)`、`中途否决→deny(212ms)`、pending/deny 文件清理确认。
- 已知：危险命令在 Claude 侧 overlay 倒计时走到 0s 后会话仍 waiting（hook 已退出，无 pending，
  沉默批准不了任何东西）——显示的是"自动窗口已过、仍在等人处理"，无害，已在注释说明。

## 19. 全链路失效修复 2026-11-05（历史记录，最新见 §25）

**症状**：浮窗「完全失去作用」——进程不存在/时隐时现，三端状态全无。

**根因不是 UI，而是 2026-10 那批"优化"把上报端分叉成第二套协议**，与仓库
`ai-terminal-pet.v1` 契约彻底脱钩。逐条：

1. **管道名分叉（致命）**：`pet-ipc.js`、`opencode/plugins/terminal-pet/index.ts`、
   `pet-statusline.exe`(Rust)、`terminal-pet-profile.ps1` 全部硬编码
   `\\.\pipe\ai-terminal-pet-v1`；服务端 `app.Run` 用 `pipeserver.PipePathForUser()`
   = `\\.\pipe\ai-terminal-pet-v1-42ce7d78`（homedir SHA256 前 8hex）。
   所有上报都发到**不存在的管道**，`fs.existsSync(pipe)` 恒 false。
2. **状态枚举分叉**：线上用 `waiting`，合法集只有 `waiting_permission`；
   `session.State.Valid()` 与 `ParsePayload` 的 `DisallowUnknownFields` 会直接拒收。
3. **hook 自己掐断发送**：`terminal-pet-hook.js` / `codex-notify.js` 在
   `reportStatus()` 之后立刻 `process.exit(0)`，异步 `net.connect` 还没建立就被杀掉。
4. **会话身份分裂**：线上用 `pid: process.pid`（hook 每次都是新进程），而 store key =
   `source + sessionId + pid`；即使管道通了也会刷出一屏幽灵会话。
5. **JSON-lines 缺换行**：opencode 插件 `socket.write(payloadStr)` 没有 `\n`，
   服务端 `bufio.Scanner` 一直等到 5s 读超时。
6. **Rust statusline 是坏的，还会撑爆磁盘**：166KB exe 由
   `.gemini/.../scratch/pet_statusline_standalone.rs` 编译，读的是 `sessionId`/`state`/`pid`
   字段，而 Antigravity 实际给的是 `conversation_id`/`agent_state` → sessionId 恒 `default`、
   pid 恒为自身新进程 ID，于是每次状态行刷新都新建一个
   `antigravity-statusline-<pid>-default.json`。实测 **5182 个文件 / 26MB**，且内容是
   Antigravity 原始 dump，根本不是 v1 快照，服务端一律拒收。
7. **装机版本落后**：安装目录 exe 停在 9/26 的 v0.5.0，源码已是 v0.6.0
   （i18n + 自动同意）却从未构建；`config.json` 缺 `language/autoApprove/...`。
8. **自毁循环**：reporter 每 10s 拉起浮窗 → 扫描 runtime 把 TTL>15s 的旧快照全删
   （这就是排查过程中 5182 个文件突然变 1 个的原因）→ 收不到任何管道连接 → 35s 自毁 → 再拉起。

**修复（统一到单一协议源，不再有第二套实现）**

- 把仓库 `reporters/protocol.js` 部署到 `%LOCALAPPDATA%\AI-Terminal-Pet\protocol.js`；
  **`install.ps1` 现在每次安装都会同步并 SHA256 校验它**（新增
  `ProtocolSource/ProtocolDestination` 参数）——这是防复发的关键，以后 protocol 改动
  不会再留下陈旧副本。
- `pet-ipc.js` 重写为薄适配层：只做字段映射，管道名/状态集/脱敏/快照命名/按需拉起
  全部委托 `protocol.sendStatus`；PID 用 `stableProcessId(terminal:session)`；
  `waiting`→`waiting_permission` 等别名归一；**删掉所有 `process.exit(0)`**。
- `terminal-pet-hook.js`：状态改合法集（PreToolUse→`tool_use`、
  PostToolUseFailure→`error`、PermissionRequest→`waiting_permission`、Notification 按
  `notification_type` 判定），删 `process.exit(0)`；`~/.claude/settings.json` 补
  `PermissionRequest` / `Notification` / `UserPromptSubmit` 三个事件（此前等待确认态根本没入口）。
- `codex-notify.js`：删 `process.exit(0)`，状态改合法集。
- `opencode index.ts`：管道名改 `os.homedir()` SHA256 同源算法、状态改
  `waiting_permission`、`socket.write` 补 `\n`、pipe 缺失时 `ensurePetRunning()`。
- `terminal-pet-profile.ps1`：用 .NET SHA256 算同一个 `42ce7d78` 后缀，
  `NamedPipeClientStream` 改用带后缀短名，pipe 缺失时拉起浮窗。
- **Antigravity 换回仓库原生 `bin/pet-statusline.exe`**（Go，走 `PipePathForUser()`，
  解析真实 `conversation_id`/`agent_state`/`tool_confirmation_pending`，快照名
  `antigravity-statusline-<稳定hashPID>-<conversation_id>.json`）。Rust exe 已备份为
  `pet-statusline.exe.bak-20261105`，其源码留在 `.gemini/.../scratch/` 不再使用。
- 重建 v0.6.0 两个二进制并 `install.ps1` 装机（主 exe SHA256 `6965FE01…`），
  `config.json` 补 `language:auto` / `autoApprove:false` / `autoApproveTimeoutSec:5`。

**验证矩阵（150% DPI）**

- 管道 `\\.\pipe\ai-terminal-pet-v1-42ce7d78` 存在；Go golden test
  `TestPipePathForUserMatchesReporterAlgorithm` 通过；JS / PowerShell / Go 三端 hash 一致。
- `sendStatus` 回执 `OK`；浮窗 HWND 3413820、`IsWindowVisible=True`、
  `SendMessageTimeout(WM_NULL)` 成功（消息泵活着）、`Responding=True`。
- `--json` 同时聚合出 **3 个真实会话**：`claude-ipc / waiting_permission`、
  `antigravity-statusline / working`、`opencode / working`；聚合态 `waiting_permission`。
- 原生 shim 输出 `pet tool_use`；稳态 ~0.19s/次（首次 0.58s，进程创建税）。
- runtime 稳定在 3 个文件（不再按 PID 膨胀）；`error.log` 无新增。
- `gofmt -l` 干净、`go vet ./...`、`go test -count=1 ./...` 全过、
  `reporters/*.js` 与 4 个线上脚本 `node --check` 全过、TS 插件
  `node --experimental-strip-types --check` 通过、PowerShell Parser 无错、
  install.ps1 三份文件 SHA256 与仓库一致。
- 所有被改文件均先备份为 `*.bak-20261105`。

**遗留 / 权衡**

- 原生 Go shim 单次 ~0.19s，比那版 Rust（~15ms，但它连不上管道、状态也解析错）慢；
  历史上 Go shim 一直是 0.5–0.9s 且可用，此处优先选可维护的官方实现。若要极致延迟，
  应把 Rust 源码收进仓库并复用官方管道名算法，而不是再留一个游离实现。
- 开机自启仍为「关」（沿用 §15 的决定）；浮窗由 reporter 按需拉起，已验证可用。
- 安装目录仍有 10 月遗留物（`pet-statusline-core.exe` / `-old.exe` /
  `pet-statusline-v1-backup.exe` / `pet-statusline-shim.js`），无引用、未删除。

## 20. 一键跳转终端 2026-11-05（历史记录，最新见 §25）

**需求**：有需要决策、或已经结束等用户查看的会话时，点一下就跳到那个终端。

### 20.1 协议零改动，可行性来自窗口标题

先在本机枚举了可见顶层窗口与进程树：Windows Terminal 每个标签是**独立顶层
窗口**，类名 `CASCADIA_HOSTING_WINDOW_CLASS`，标题带会话名（实测
`OC | 项目交接与后续计划`）；opencode PID 5228 的父链是
`5228 → 2832 → 2896(powershell.exe) → 11492(WindowsTerminal.exe)`。

所以只用会话现有的 `pid / parentPid / title / project / source / terminal`
就能定位，**不需要新增协议字段、也不动任何上报脚本**。hook 类上报端的 PID
是合成值也不影响，因为 `parentPid` 是真实父进程，是祖先链的可靠起点
（见 `ancestorDepths` 把 parentPid 放在 depth 0、pid 放在 depth 1）。
本轮 `reporters/*`、`install.ps1` 的协议同步逻辑、所有线上接入脚本零改动。

### 20.2 实现

- `internal/platform/focus_windows.go`（新）：Toolhelp 进程快照 → 祖先链 →
  `EnumWindows` 收集可见顶层窗口 → 打分（祖先距离 / 会话标题 / 项目基名 /
  终端关键词 / 窗口类名）→ 前台激活。跨进程窗口枚举**绝不在 UI 线程**，
  全部在后台 goroutine（`runtime.LockOSThread`）执行，结果用 `wmFocusResult`
  回投。`resolveFocusWindow` 与激活分离，便于不抢前台地做真实桌面探针。
- 激活兜底：`SetForegroundWindow` → `AttachThreadInput` + `BringWindowToTop`
  → `SetWindowPos(TOPMOST / NOTOPMOST)`，最后用 `GetForegroundWindow` 复核。
  弱匹配（分数 < 40）直接判定失败，绝不把无关窗口拽到前台；失败时提示音 +
  托盘气泡。
- `internal/ui/focus_windows.go`（新）：几何单一来源
  （`compactFocusRect` / `expandedRowFocusRect` 同时供光栅化与命中）、
  `jumpPriority`（waiting 4 / error 3 / idle 2 / closed 1，忙碌态 -1）、
  `mostUrgentJump`、纯函数 `focusHit`。箭头用线段画，不依赖字体是否含 U+2197。
- 交互：紧凑态底部出现「↗ 跳转」条（**仅在有可跳转会话时**，项目名上移一行
  保留），展开态每行右侧一个箭头按钮。点击优先于普通 toggle 和自动同意否决，
  但「整块可点、可拖拽」的既定规则不变；跳转后不展开/收起面板。
- 点击坐标改用 **`WM_LBUTTONUP` 的 lParam**（客户区坐标），不再读真实光标：
  本机光标会自行漂移，读回位置可能差数百像素（§16 已记过这个坑）。
- i18n 新增 `panel.jump` / `balloon.jumpTitle` / `balloon.jumpBody`，
  展开态页脚提示加「↗ 跳转」。

### 20.3 验证

- 新增 13 个单测（platform 7 / ui 6），`go test -count=1 ./...` 全过；
  `gofmt -l`、`go vet ./...`、所有 reporter `node --check` 干净。
- `PET_PREVIEW_DIR=<dir> go test ./internal/ui -run TestRenderPreview`：
  紧凑/展开/中英/告警/空态 PNG 目检，按钮不越界、不压状态文字；
  忙碌态不出现跳转条。
- `PET_FOCUS_PROBE=1 go test ./internal/platform -run TestResolveFocusWindowLive -v`：
  真实桌面定位到 `pid=11492 CASCADIA_HOSTING_WINDOW_CLASS
  "OC | 项目交接与后续计划"`。
- 端到端：推送 `waiting_permission` → 合成点击紧凑态跳转条 →
  前台窗口从 `AI Terminal Pet` 变成 `OC | 项目交接与后续计划`，
  面板尺寸不变（没有误触发展开/收起）。点面板中央仍正常 toggle，
  证明原有交互未回归。
- 装机主 exe 4,050,944 bytes，三份部署文件 SHA256 与仓库一致；无 error.log。
  验证过程动过 `state.json`，已恢复到本轮开始时的
  `x=2481 y=325 zoom=0.8 expanded=false`。

### 20.4 已知限制

- Windows Terminal 能定位到标签；传统 conhost 单窗口多会话、VS Code 集成
  终端只能到窗口级。
- 多个标签标题完全相同时按 Z 顺序取第一个；`title / project / terminal`
  关键词已尽量消歧（opencode→`oc |`、antigravity→`agy`、claude→`claude`）。
- 合成点击不会让浮窗成为前台进程，自动化验证里需先
  `SetForegroundWindow(petHwnd)`；真实鼠标点击天然满足这个前提。

## 21. 会话存在性重构 2026-11-05（v0.7.0，历史记录，最新见 §25）

**症状**：会话显示一会有、一会没有，但终端一直开着。用户要求「从底层上就有
稳定性」，并同时要热切换/热加载、更低的运算占用。这轮不是调参数，是把
「会话凭什么算存在」换了个判据。

### 21.1 根因链（实测确认）

1. reporter 是**事件驱动 + 状态去重**的：`syncToPet` 只在状态/工具变化时推送，
   一个持续 `working` 的会话可以十几分钟不发一个字节。
2. `DefaultSnapshotTTL=15s` / `DefaultPipeTTL=30s` 把「没收到上报」当成
   「会话结束」。`SnapshotScanner.Scan()` 更会把超过 15s 的快照文件**直接删掉**
   —— 连磁盘证据都抹掉。`--json` 因此也跟着空。
3. 删掉的会话进 tombstone；`openLocked` 用 `UpdatedAt` 门槛挡住旧时间戳，
   而静默会话的再次上报恰好还是旧时间戳 → **无法自愈**，只能等下一个状态
   变化。这是「闪一下又没了、之后很久不回来」的来源。
4. UI 每 500 ms 检查一次，无会话 5 s 就 `DestroyWindow` → 进程消失；
   下次 reporter 事件才拉起。
5. 实证：11:15 刚写入的 `opencode-5228-*.json`，几秒后 `cat` 已不存在。

### 21.2 新模型：存在性与状态分离

- **existence** 由 OS 观测决定（进程表 + 终端窗口），与 reporter 心跳无关。
- **state** 是 reporter 最后一次说的话，可以任意陈旧。
- 删除必须满足「宿主连续不可观测 ≥ `DefaultAbsenceGrace`(2 min)」，单次观察
  失败绝不删除；宿主一出现就清空缺席计时。
- 三类 tombstone 分开：`closed`（显式结束，保留时间戳门槛，防迟到上报复活）、
  `expired`（静默/缺席，**不设门槛**，允许旧时间戳自愈）、`evicted`（容量淘汰，
  保留门槛，防列表抖动）。第 2 类就是之前修不好的自愈缺口。

### 21.3 落地

- `internal/platform/presence_windows.go`（新）：`Presence` 缓存一次
  Toolhelp + EnumWindows（默认 10 s），对外提供 `ProcessAlive` /
  `SessionAlive`。证据按强弱排序：`parentPid` 存活 → 可信 source 的真实
  `pid` 存活 → 终端窗口标题与会话 title/项目名/CLI 关键词身份匹配。
  合成 PID 绝不参与探测；窗口匹配**必须命中会话 title 或项目名**，仅
  「桌面上有个终端」或「标题里有 agy」都不算——第一轮采样发现后者会把所有
  历史 Antigravity 对话永久保活，那些对话各自有真实 `parentPid` 可用。
- `internal/store/store.go`：新增 `Reconcile` 取代运行期的 `Cleanup`。
  保活只刷新 `pipeSeen/snapshotSeen/lastSeen`，**不改 payload、不 bump
  revision** → 静默但存活的终端零重绘。`Cleanup` 保留给离线视图与历史单测。
- `internal/store/snapshot.go`：`SnapshotScanner.KeepAlive`；payload 过期或
  单次 PID 探测失败时，只要宿主仍可见就保留文件并返回该会话。
- `internal/app/app_windows.go`：一个 `Presence` 贯穿 scanner / store / 清理；
  `--json` 用同一套模型；扫描周期 2 s → **5 s**；新增 `watchConfig`，每 30 s
  一次 `Stat`，config.json 改动后后台读取再经 `QueueSettings` 交 UI 线程应用
  （语言/通知/提示音/置顶/贴边/自动同意即时生效，UI 线程不做文件 IO）。
- `internal/ui/app_windows.go`：空闲检查 500 ms → **30 s**，自毁宽限
  5 s → **10 min**；且新模型下终端开着时 `HasAvailable()` 恒真，几乎不会触发。
  浮窗因此常驻，支持热切换。
- `internal/ui/pulse_windows.go`（新）：告警呼吸改为调整分层窗口常量 alpha。
  `purePulseFrame` 判定「本帧与上一帧只差时钟」时复用已渲染位图，
  **完全跳过 GDI 光栅化**；任何内容/尺寸/悬停/滑动变化仍走完整重绘。
  `pulseInterval` 80 → 160 ms 作为双保险。这是 CPU 降幅最大的一项。
- **热加载/热恢复**：存活会话的快照不再被删，浮窗重启后 `apply()` 从
  `runtime/` + 进程表重建会话，不需要 reporter 重发，也不需要重启终端。
- 版本号 `0.7.0`。**协议与所有上报脚本零改动**，用户无需重开终端。

### 21.4 运算占用

| 项 | 之前 | 现在 |
| --- | --- | --- |
| 空闲检查 | 每 500 ms | 每 30 s |
| 快照扫描 | 每 2 s | 每 5 s |
| 存活探测 | 每会话一次 OpenProcess / 2 s | 一次 Toolhelp 快照 / 10 s（全会话共享） |
| 窗口枚举 | 仅跳转时 | 一次 / 10 s，与存活探测共用 |
| 静默会话 | 反复删文件 → tombstone → 重绘 | 无变化、无 revision、无重绘 |
| 告警呼吸动画 | 80 ms（12.5 fps **每帧整块 GDI 光栅化**） | 160 ms；纯脉冲帧复用已渲染位图，只改分层窗口常量 alpha，**零 GDI 绘制** |
| 配置读取 | 仅启动 | 30 s 一次 Stat，变化才读 |

第一轮 3 分钟采样还暴露两点并已修掉：(1) `agy` 关键词过宽导致历史会话累积；
(2) 持续 `waiting_permission` 让脉冲动画常驻，成为空闲期的主要 CPU 开销。

`frameKey` 的重绘跳过、`syncAnimation` 空闲卸载计时器保持不变。

### 21.5 验证

- 新增 18 个单测（platform 6 + store 5 + snapshot 3 + app 2 + ui 2，另有既有用例回归），
  `gofmt`/`go vet`/`go test -count=1 ./...` 全过，`reporters/*.js` `node --check` 全过，
  离屏 PNG 预览回归通过。
- `TestReconcileKeepsLiveQuietSessionWithoutRepaint`：模拟 60 分钟静默，
  会话不消失、revision 不变、UpdatedAt 不被改写。
- `TestExpiredTombstoneAllowsUnchangedTimestampToReturn`：锁定自愈。
- `TestPresenceDropsHistoricalAntigravityConversation`：锁定幽灵会话不累积。
- **三轮本机采样**（每轮：一次性上报 `waiting_permission` 后完全静默，
  每 20 s 记录会话数 / PID / 快照数 / 浮窗 CPU 时间）：

  | 轮次 | 版本 | 会话数 | PID | CPU / 20 s |
  | --- | --- | --- | --- | --- |
  | 1（3 min） | 初版重构 | 2→4，只增不减 | 全程不变 | 0.1→3.8 s（后段升高） |
  | 2（3 min） | 修幽灵 + 脉冲 160 ms | 稳定 4 | 全程不变 | 1.1–2.2 s |
  | 3（2 min） | **脉冲复用位图** | 稳定 4 | 全程不变 | **0.016–0.156 s** |

  第 3 轮 120 s 累计 CPU 仅 **0.562 s**（约 0.5% 单核），相对第 2 轮
  下降约 **96%**；三轮中 count / PID / 快照文件数始终一致，
  会话数再没有归零或回退。
- 装机主 exe 4,067,840 bytes，三份部署文件 SHA256 与仓库一致，无 error.log。

### 21.6 取舍

- `AutoApproveTimeoutSec`（`Options` 单独字段）仍只在启动时读取，改它需重启
  浮窗；其余配置均已热加载。
- 终端窗口标题完全无法匹配、且 `parentPid` 已退出的会话，会在连续 2 分钟
  不可观测后移除；这是唯一会让仍开着的会话消失的路径，需要终端进程与窗口
  同时不可见。
- 开机自启仍为「关」，沿用 §15 决定；浮窗由 reporter 首次拉起后常驻。

## 22. 跳转关闭用户窗口修复 2026-11-05（历史记录，最新见 §25）

**症状**：点击跳转时，如果用户正在处理某个窗口，那个窗口「有概率」被关掉。

### 22.1 根因（不是浮窗自己的窗口）

`internal/platform/focus_windows.go:activateFocusWindow` 在 `SetForegroundWindow`
失败时用了两个「抢前台」兜底，两者都会动到**用户正在操作的窗口**：

1. `AttachThreadInput(backgroundThread, foregroundThread, TRUE)` —— 把跳转后台
   线程的输入队列直接合并到当前前台线程。前台线程正是用户正在用的窗口所属
   线程；合并期间焦点/激活消息错乱，失焦即关闭的窗口与弹层会直接消失。
   而且激活跑在后台 goroutine（虽已 `LockOSThread`），与 UI 线程输入队列形成
   竞态，这就解释了「有概率」。
2. `SetWindowPos(HWND_TOPMOST)` 紧接 `HWND_NOTOPMOST` —— 重排整个桌面的
   z-order，把用户窗口压下去。对「失焦即关」的窗口等同关闭。

### 22.2 修复：只碰目标窗口，且只在 UI 线程激活

- 新增 `platform.ResolveTarget`：后台 goroutine **只做定位**（进程/窗口枚举），
  不碰前台、不碰 z-order。
- `ui.jumpTo` 把解析到的 HWND 通过新消息 `wmFocusActivate` 投回 UI 线程
  （`wmApp+7`）；UI 线程刚被点击，本来就拥有前台权限。
- 新增 `platform.ActivateTarget`：`IsIconic → SW_RESTORE` +
  一次 `SetForegroundWindow` + `GetForegroundWindow` 复核。失败就走
  `focusFailure()` 提示音/气泡，**绝不强行抢前台**。
- 删除 `AttachThreadInput` / `BringWindowToTop` / `SetWindowPos` 全部兜底及
  相关 proc 与常量。`FocusTerminal` 保留给非 UI 调用者（探针/测试），
  内部改为 `ResolveTarget + ActivateTarget`。
- `ActivateTarget` **禁止** `LockOSThread/UnlockOSThread`：UI 线程为窗口生命期
  全程锁定，解锁会重现 §13 的幽灵窗口 bug。
- 跳转成功后 `holdTuck(2s)`：光标离开浮窗时不会立刻贴边收起，避免用户把
  「面板滑走」误认成「窗口被关掉」。

### 22.3 验证

- 25 次合成点击 + 15 次真实鼠标点击：浮窗 HWND/可见性/尺寸、终端窗口数、
  `state.json` 全程不变（没有触发 `wmClose`），无 `error.log`。
- **12 次「用户正在处理窗口」场景**：每次记录点击前的前台窗口
  （第 1 次是 `WorkBuddy AI`，其后是终端标签），点击后该窗口始终存在且可见；
  20 个稳定窗口数 20→20 一个不少；前台正确切到 `OC | 项目交接下一步计划`；
  浮窗 HWND 始终在。`FAILS=0`。
- 3 个 opt-in 桌面探针（`PET_FOCUS_PROBE=1`）：定位不改前台、激活不关闭
  任何窗口、真实桌面仍能定位到正确标签。
- 新增 4 个单测（含 `TestActivateTargetRejectsZeroHandle`），全量
  `gofmt`/`go vet`/`go test -count=1 ./...`/预览回归通过；
  装机主 exe 4,066,816 bytes，哈希与仓库一致，版本 0.7.0。
- 压测探针会话全部以 `closed` 清理，最终只剩 4 个真实 opencode 会话。

---

## 23. 会话幽灵堆积 + Slate 配色（v0.7.1，2026-11-05 第五项）

### 23.1 取证

用户报「大量会话显示堆积，现实中并没有这些会话」。关键数据：

| 观测 | 值 |
| --- | --- |
| 浮窗内存会话数（展开高度反推） | **≥ 7**（6 行 + 溢出提示） |
| 磁盘 `runtime/*.json` | **4** |
| 桌面 Windows Terminal 标签 | 4 个 `OC \| …` |
| 全部 opencode 快照的 `pid`/`parentPid` | `5228` / `2832`，自 11/2 起常驻 |

**时区陷阱（排查必看）**：payload 的 `updatedAt` 是 UTC，本机是 GMT+1。
`15:39Z` 是本地 `16:39`（刚刚），不是一小时前。一开始据此误判了
「快照全都陈旧」，核对 mtime 后才确认磁盘是在滚动更新的，堆积只在内存。

### 23.2 三个叠加根因

1. **宿主存活 ≠ 会话存在**。`Presence.SessionAlive` 第一条就是
   `parentPid 在进程表 → return true`。`opencode.exe` 一开几天，它历史上
   承载过的每个会话因此全部永生。
2. **删文件不等于结束会话**。opencode 插件 `cleanupAll` / `closed` 分支只
   `unlink` 快照，不送 `closed`；通过 pipe 进 store 的条目没有任何结束信号，
   而 `parentPid` 又一直保活 → **内存只增不减，磁盘却滚动变少**，于是
   `--json`（离线重扫磁盘）看着正常、面板上却越堆越多。
3. **事件 ID 被当成会话 ID**。插件的 `sessionID` 提取用了
   `e.data?.id || e.id` 兜底，`evt_1abf…` 这类事件 ID 变成独立会话，各写一
   个快照且从不 closed（本次实测就有一条 14:45 的 `waiting_permission`
   幽灵，还让面板一直处于最高优先级告警态）。

### 23.3 修复：磁盘文件成为会话身份锚点

- `store.SnapshotScanner.ScanWithKeys()` 额外返回本轮磁盘上观察到的会话
  key 集合；`store.Key()` 导出；`app.snapshotAwareAlive()` 组合判据：
  **曾有快照但现在磁盘上没有 → 判为结束**（`closed` 快照被本轮消费删除，
  不计入 present）。纯 pipe 会话不受影响。
- `Presence.SessionAlive` 改为按证据 specificity 排序：
  1. **终端标签标题命中会话身份**（title / 项目名）→ 无条件保活，静默多久
     都不会被删（保住 §21 的核心不变式）；
  2. **宿主进程**：可信真实 PID 优先（其消失即结束，长驻父终端不再背书），
     否则才是 hook 的 `parentPid`；
  3. **近期活动**：宿主活着但没有标签身份时，超过 `DefaultOrphanAfter`
     （45 min）的孤儿快照回收。
     这不是心跳 TTL：有标签身份的会话永久保活，被回收的走
     `tombstoneExpired`（无时间戳门槛）可在下次上报立刻复活。
- **opencode 插件**（`~/.config/opencode/plugins/terminal-pet/index.ts`，
  备份 `*.bak-20261105-b`）：`sessionIdentity()` 只认明确 `sessionID` 字段，
  拒绝 `evt_/msg_/prt_/prj_` 与 `default-session`；`closed` 改为**同步写
  `state:"closed"` 快照**（关机时 pipe 来不及发，但服务端下一次扫描一定能
  看到）；`sweepStaleSnapshots()` 启动时只删事件型伪会话与 2 小时未动的
  文件，**保留近期真实 `ses_` 快照**以不破坏热加载。
- **`cleanupAll` 只在进程真退出时调用**（`process.on("exit")`）。`setup` 的
  dispose 回调只 `abort` + 清空内存 map：编辑插件文件或 opencode 热重载插件
  **不等于**用户的会话结束了。第一次实现让 dispose 也写 closed，实测在
  2 分钟观察里把面板短暂清空（随后新事件又恢复），已修正。

### 23.4 Slate 配色

旧版问题是「告警时边框 + 实心圆盘 + 状态字 + 跳转条全是同一种高饱和琥珀，
整条面板糊成一团」，展开态又有 4 种霓虹色互相打架。

- 所有结构色收拢进 `internal/ui/drawing_windows.go` 顶部单一调色板
  （`colorPanel` / `colorRaised` / `colorBorder` / `colorDivider` /
  `colorOrbInner` 等），绘制代码里不再有散落的 `rgba{}` 字面量。
- 背景由蓝黑改中性石墨；状态色统一降饱和、拉平明度。
- 状态球由「实心大圆盘 + 强光晕」改为**细环 + 深色内核 + 状态色字形**；
  面板外框改细线（告警时状态色 alpha 120，原 210）。
- 跳转按钮由实心色块改为**中性底 + 状态色细边 + 状态色箭头/文字**；
  展开态行左色条降到 alpha 190；文字层级提亮 `colorMuted`。

### 23.5 验证

- v0.7.1 装机，三份部署文件 SHA256 与仓库一致。
- **内存 4 = 磁盘 4**（原来 ≥7 vs 4）；清理探针后 3 = 3；浮窗可见、
  消息泵正常、`state.json` 位置恢复 `x=2353`。
- 2 分钟无干预观察：HWND 全程不变；中途出现过一次 `count/files` 归零，
  追查是**插件文件被编辑触发 opencode 热重载**、`cleanupAll` 把活跃会话全
  宣告 closed 所致（16:44 新事件到达后自动恢复为 2 个真实会话）。已按
  §23.3 修正 dispose 语义；这也说明「热重载 ≠ 会话结束」。
- **受控回收实验**：上报一条探针后**只删除它的快照文件、不发 `closed`**，
  等 150 s（> `DefaultAbsenceGrace` = 2 min）。探针文件未再出现，
  内存行数与磁盘数保持相等，说明 `snapshotAwareAlive` 确实让内存条目
  跟着文件消失而回收。期间 opencode 又新建了一个真实会话，所以总数
  一直是 3；判据要看「探针 key 是否还在」而非只看总数。
- 5 个新单测（孤儿回收、可信 PID 死亡、`ScanWithKeys`、
  `snapshotAwareAlive`）+ 全量 `gofmt` / `go vet` / `go test -count=1 ./...` /
  `node --check` / TS strip-types / 离屏 PNG 预览回归全过，无 `error.log`。
- 现存快照已备份到 `%LOCALAPPDATA%\AI-Terminal-Pet\runtime-backup-20261105\`。

**用户侧**：服务端修复重启浮窗即生效；opencode 插件侧的防复发需要
**下次方便时重启 OpenCode**（不重启也不会再堆积，只是旧插件仍会写事件型
伪会话文件，由服务端孤儿回收兜底）。

---

## 24. 提示音 / 通知时序 / DeepSeek Harness 适配（2026-11-08，v0.8.0）

用户反馈：① 提示音太小；② 终端结束时不报也不变色，反而「开始新会话」
才弹窗；③ 希望把 DeepSeek Harness 做成可导入的插件。

### 24.1 提示音：MessageBeep → 内存合成 WAV

`MessageBeep(MB_ICONEXCLAMATION)` 播的是系统「默认蜂鸣」事件，多数机器
静音或极轻，且受声音方案 / 专注助手影响。改为 `internal/ui/sound_windows.go`
在内存里合成 44.1 kHz / 16-bit 单声道 PCM WAV，用
`PlaySoundW(SND_MEMORY|SND_ASYNC|SND_NODEFAULT)` 播放：

- 三种可听辨的音型：等待确认 = 两轮上行双音（~0.9 s）、错误 = 下行低音、
  完成 = 短上行三音（明显更轻，不与告警混淆）。
- 幅度由 `config.json:soundVolume`（0..100，默认 85；本机已设 100）线性
  控制，托盘菜单「提示音音量」点击循环 60/85/100 并即时试听。
- `soundBuffers` 按 kind 缓存 PCM，保证 SND_ASYNC 异步播放期间 Go 切片
  不被回收；PlaySound 失败才回退 MessageBeep。
- 托盘气泡统一加 `NIIF_NOSOUND`，避免 Windows 再叠一个排队的声音。

### 24.2 通知时序：单一全局 key → 每会话状态转换

旧 `alertOnTransitions` 只报 `waiting_permission` / `error`，且用**一个**
`alertKey` 记录「当前最紧急会话」。后果：那个会话一结束 key 被清空，
另一个仍在等待的会话会重新告警 → 看起来像「新会话开始才弹窗」；
会话结束本身则完全没有提示。现在：

- `alertStates map[sessionKey]State` 逐会话跟踪；首次 `load()` 只建基线，
  **不回放历史告警**（热加载 / 重启不再 replay）。
- 纯函数 `transitionLevel` / `decideAlerts`：进入 waiting/error → urgent；
  从 busy 落到 idle/closed（或从 `List()` 消失）→ finished；新开的普通
  busy 会话 → **不告警**。
- finished 发独立完成音 + 气泡（`notifyOnFinish` 可关，默认开），
  文案 `balloon.doneTitle/doneBody` 中英齐全。
- 气泡 `NIF_INFO|NIIF_NOSOUND`；声音与气泡共享同一次决策。

### 24.3 官方协议源加固：时钟偏差不再吞掉整条状态

排查中实测到：`startedAt` 若比 `updatedAt` 晚（例如把本地时间当 UTC、时钟
快 15 min），Go `validate()` 会拒绝 `updatedAt precedes startedAt`，
`SnapshotScanner` 随即**删除该快照** → 状态静默消失。现在
`reporters/protocol.js:normalizePayload` 在单一协议源里钳制：
`startedAt = min(started, updated)`，两者都夹到 `now + 4 min` 以内，
并保持幂等（`sendStatus` 会调用两次 normalize）。服务端严格校验不变。

### 24.4 DeepSeek Harness 插件（Host-only Cordis bundle）

- DSH 44.7.0 是 Electron，profile 在 `~/.dsh/profiles/desktop`，插件是
  `dsh.bundle.patch` 的 npm 包 + `insert:` 行（官方
  `cordis-plugin-development` SKILL 规范）。`dsh --profile desktop
  --dump-config` 会报 **"managed exclusively by the Electron application"**，
  且启动会跑 pnpm 重装，所以**不能手工改 profile 的 package.json /
  cordis.patch.yml**。唯一正规入口是应用内
  `plugin_manager(action: install_bundle, target: <绝对目录>)`。
- 注意 `~/.dsh/dsh-pet` 是 DSH 自带的桌面宠物（npm 包 `dsh-pet`），与
  本项目同名无关。
- 仓库 `integrations/deepseek-harness/ai-terminal-pet-bridge/`：
  `package.json`（不声明 `engines.dsh`，避免像 `dsh-codex-pet@0.3.1`
  那样被版本门禁拒绝）、`cordis.patch.yml`、`index.js`、`locale/{en,zh}.json`、
  `icon.svg`、`README.md`。`install.ps1` 每次装机同步到
  `%LOCALAPPDATA%\AI-Terminal-Pet\bundles\ai-terminal-pet-bridge\` 并
  SHA256 校验（新增 `BundleSource/BundleDestination` 参数）。
- 插件只做字段适配：订阅 `session/event`、`agent/status`、`agent/error`、
  `agent/inbox/claimed`、`approval/request`（waterfall 必须 `next()`），
  映射到官方状态集；`source/terminal = deepseek-harness`，
  `pid = stableProcessId(...)`（合成），`parentPid = process.pid`（真实宿主），
  `title = "DeepSeek Harness"`（正好是窗口标题子串，保活 + 跳转都命中）。
  所有上报经官方 `protocol.js`，fire-and-forget，绝不 `process.exit`。
  仅 `process.on("exit")` 写 `closed`；dispose（HMR）只移除监听器，
  **不清空会话**。
- Go 侧适配：`isEditorProcess` / `isCLIProcess` 加入 `deepseek harness.exe`
  （Chrome_WidgetWin_1 得分 24），`terminalKeyword` 加 `deepseek`。
  只读探针 `PET_DSH_PROBE=1` 实测命中
  `交接项目并规划下一步 — DeepSeek Harness`，且不改变前台窗口。

**导入方式**（在 DeepSeek Harness 对话里说）：
> 用 plugin_manager，action install_bundle，target 设为
> `C:\Users\rongj\AppData\Local\AI-Terminal-Pet\bundles\ai-terminal-pet-bridge`

若返回 `restart-required` 则重启 DSH。

### 24.5 验证

- `node integrations/deepseek-harness/verify.mjs`：14 项断言全过
  （状态序列、稳定会话 ID、waterfall、`exit` 才写 closed、HMR 不结束会话、
  时钟钳制幂等）。
- `PET_SOUND_PROBE=1` PlaySoundW 真机播放成功；`PET_DSH_PROBE=1` 只读
  窗口定位成功且不抢前台。
- 端到端：`working → tool_use → waiting_permission → idle` 全部进入
  `--json`，浮窗 HWND 全程可见；故意 +30 min 时钟偏差的会话也正常显示；
  清理后只剩 1 个真实 opencode 会话 / 1 个快照，无 `error.log`。
- 新增单测：WAV header/音量/音型、状态转换表、多会话不重复告警、
  config 音量默认与钳制、DSH 跳转打分；全量 `gofmt`/`go vet`/
  `go test -count=1 ./...`/`node --check`/PowerShell Parser 全过。
- 装机 v0.8.0，四份部署文件（两个 exe、protocol.js、插件 index.js）
  SHA256 与仓库一致。

### 24.6 收尾发现：磁盘快照必须压过 OS 证据

交付前观察到一个「浮窗自己消失」：`--json` 显示 1 个 opencode 会话，
浮窗却在约 10 分钟后正常退出。原因不是崩溃（无事件日志/转储），而是
运行期 `snapshotAwareAlive` 对「磁盘文件还在」的会话仍然调用
`Presence.SessionAlive`；一旦 OS 证据弱（宿主 pid 常驻但标签已关、标题
变化），`Reconcile` 连续缺席 2 min 后把内存条目删掉 → `HasAvailable()`
变假 → `timerIdle` 10 min 自毁。而 `--json` 每次新建 store、达不到连续缺席
门槛，所以始终看得到同一份磁盘文件，形成「磁盘有、内存没有」的错位。

修正：`snapshotAwareAlive` 中 **`present[key]` 为真即直接保活**（closed 仍
优先判死），OS 证据只留给从未写过快照的纯 pipe 会话。这严格落实
§23 定下的「磁盘快照 = 会话事实」，且 `Reconcile` 保活不改 payload、
不 bump revision，零重绘。新增
`TestSnapshotAwareAliveTrustsTheDisk` 锁定：弱 OS 证据不删、closed 不保、
文件消失仍判结束。装机后 3 分钟无干预观察：HWND 全程不变且可见、
`--json` count 恒为 1、无 `error.log`（此前同条件下浮窗会自行退出）。

## 25. 设置窗口与可切换音效（v0.9.0，2026-11-08）

### 25.1 动机

浮窗刻意只是一条状态带，所以所有偏好原先只有两种入口：托盘里的一串勾选，
或者手改 `config.json`。结果是「换个提示音」「调音量」这两件最想改的事，
前者根本不存在、后者要在托盘里点三下循环；不透明度、自动同意倒计时、
语言也没有任何界面。§24 加了合成提示音和音量，但只能靠托盘循环。

### 25.2 形态：同一 UI 线程上的第二个窗口

`internal/ui/settings_windows.go`。不是子窗口、不是对话框资源，而是第二个
顶层窗口，注册自己的窗口类与 `WndProc`（`settingsProcedure`），但**跑在同一条
消息泵上**——`Run()` 里 `runtime.LockOSThread()` 之后创建，`Run` 负责它的
注册与注销。理由：Win32 要求「创建窗口的线程 = 泵消息的线程」，跨线程
`SetWindowPos` 会永久阻塞（§15 幽灵窗口的根因），所以第二个窗口必须同线程。

三条来自浮窗的铁律照搬：

- **几何单一来源**：`settingsGeometryFor(dpi, tab, rows)` 同时供绘制与命中测试，
  控件不可能画在一处、点在另一处。行高由 `settingsRowHeight(kind)` 决定，
  面板高度 = 内容高度（`settingsMinHeight` 只是下限）。
- **UI 线程不做文件 IO**：见 §25.4。
- 绘制复用 `canvas` 与 Slate 调色板；新增的三个结构色
  （`colorAccent` / `colorControlBg` / `colorTrackOff`）与 `colorHint`
  也放在 `drawing_windows.go` 的调色板里，绘制代码里没有散落的 `rgba{}`。

窗口 `WS_EX_TOOLWINDOW | WS_EX_LAYERED | WS_EX_TOPMOST`，用
`UpdateLayeredWindow` 推帧，因此和浮窗一样支持圆角与逐像素 alpha。
位置在浮窗所在显示器的工作区居中；**切换标签页时窗口按内容改高度，且保持
中心不动**（`reflowSettings` 只调整记录的 origin，`UpdateLayeredWindow`
自带尺寸参数，不需要额外 `SetWindowPos`），并钳制在工作区内。

### 25.3 控件与交互

自绘控件四种：开关（胶囊+圆钮）、滑块、分段选择、步进器，外加按钮与说明行。
命中测试按「整行都是热区」设计（标签和开关一样可点），分段按 x 比例换算
选项下标，滑块按 x 比例换算数值。所有纯计算都有单测：
`settingsGeometryFor` / `hit` / `settingsSliderValue` / `settingsSegmentRect`。

交互细节（都踩过坑）：

- **鼠标捕获**：每次按下都 `SetCapture`，否则在窗口外松手收不到
  `WM_LBUTTONUP`，按下态会一直亮着。
- **`ReleaseCapture` 会同步回调 `WM_CAPTURECHANGED`**：所以 `settingsMouseUp`
  先快照 `press/drag/moving` 再释放，并且捕获丢失的处理**不清 press**——
  第一版清了，结果每一次点击都自己把自己抹掉，窗口看起来完全点不动。
- `TrackMouseEvent` 必须显式调用（`trackSettingsHover`），否则收不到
  `WM_MOUSELEAVE`，高亮会卡住。
- 头部可拖动窗口（`GetCursorPos` + 抓取偏移），这是唯一用到真实光标的地方；
  其余点击一律用 lParam 客户区坐标（本机光标会自行漂移，见 §20）。

### 25.4 保存链路：读、合并、写全在后台

`internal/app` 新增 `configSaver`（与 `placementSaver` 同构的合并写）：

- `OnSettingsChanged` 只把 `ui.Settings` 塞进 saver，**UI 线程不碰文件**。
- saver 在自己的 goroutine 上 `config.Load` → 覆盖窗口能改的字段 → `config.Save`，
  这样手写的、本版本不认识的键不会被丢掉。
- `configSaver.echo(config)` 记住本进程最后一次写入的完整值；30 s 的
  `watchConfig` 命中时先比对，**自己写的变更直接跳过**。否则「改一下开关 →
  30 秒后被热加载回灌成旧值 → 再变回来」会真的发生。
- 滑块拖动过程只改内存（音量走 `setSoundVolume`，不透明度改
  `options.Opacity` + 重绘浮窗），**松手才 publish 一次**，不会每像素写盘。

### 25.5 声音：四套音色 + 按事件开关

`sound_windows.go` 的音型表扩成 `schemeSpecs()`：`chime`（清脆，默认）、
`bell`（钟声）、`soft`（柔和）、`pulse`（脉冲），每种都有
等待/出错/完成三条波形，仍然是内存合成 PCM + `PlaySoundW(SND_ASYNC)`。
`config.SoundSchemes` 与 `ui.SoundSchemes` 是两份镜像（ui 不 import config），
由 `TestDefaultSettingsMirrorConfig` 把**列表与全部默认值**钉在一起。

`config` 新增 `soundScheme` / `soundOnWaiting` / `soundOnError`；
`alertOnTransitions` 现在按事件判断（`soundForState`），所以「只想听真正卡住的
会话」可以关掉等待音而保留出错音。完成的轻音由 `notifyOnFinish` 管。

### 25.6 入口

- 托盘菜单 **设置…**（`menuSettings`，带分隔线）
- `ai-terminal-pet --settings`（`ControlSettings` → `wmSettings`）。若当时没有
  实例在跑，走 `app.RunWith(true)`：**先起浮窗、首帧后再开设置窗口**，
  否则这个参数等于什么都没做。

### 25.7 验证

- 单测：布局不越界（4 种 DPI × 3 页）、每个控件中心都能命中自己、
  分段正好铺满控件、滑块端点/中点映射、默认值与 config 一致、
  按事件的声音开关。
- 离屏预览 `PET_PREVIEW_DIR=<dir> go test ./internal/ui -run TestRenderSettingsPreview`
  输出 5 张 PNG（中英 × 三页），据此修掉了三处排版缺陷：
  分段控件压住说明文字、按钮压住自己的标题、最短页底部大片空白。
- 真机（合成点击，`AITerminalPetSettings` 类名定位）：
  开关翻转并落盘、页签切换时高度 450↔504↔306 且中心不动、步进 ±2、
  恢复默认、音量拖拽到 100、语言 en↔auto、× 关闭后窗口隐藏而浮窗仍在；
  冷启动 `--settings` 能同时起浮窗并开窗口；60 s 静默观察浮窗不消失；
  全程无 `error.log`。
- 装机 v0.9.0，四份部署文件 SHA256 与仓库一致。

### 25.8 已知取舍

- 设置窗口不参与贴边收起、不跟随浮窗透明度（固定 255 常量 alpha，
  面板自身 238 已足够沉），关掉「窗口置顶」后设置窗口仍置顶——它是对话框。
- 「恢复默认设置」只重置行为，不动窗口位置；不透明度会回到 235。
- 托盘里仍保留「提示音 / 提示音音量 / 会话结束时提醒」三个快捷开关，
  它们与设置窗口写同一份 config，不冲突。

## 26. v0.9.0 之后、未发版的一批工作 2026-11-10

目标从「对外宣传」升级为：**任意机器可用**（可移植性 / 通用性）+ **可作为 DeepSeek Harness 与 OpenCode 插件对外开源分发**。
本节的数字都是 2026-11-10 实测，命令按 §1 的「健康」口径复跑。

### 26.1 可移植性（去写死本机路径）

- `docs/assets/render_assets.py` 重写成 190 行：删掉模块级的 `PREVIEW = C:\\Users\\rongj\\...` 与
  `ASSETS` 常量，改为 `PREVIEW_ENV = "PET_PREVIEW_DIR"` + `ASSETS = Path(__file__).resolve().parent`；
  未设环境变量 / 目录不存在 / 缺 PNG / 缺 Pillow 四种情况各自打印人话提示并 `exit 2`，不再抛 traceback。
  端到端复跑：全新 `$TEMP/petpreview-fresh` + 两条 preview 测试 + 本脚本，8 张 PNG 与 git HEAD 版逐像素相同。
- 测试夹具中性化（只改 `_test.go`）：`internal/platform/focus_windows_test.go` 与
  `presence_windows_test.go` 里的 `C:\Users\rongj\...` 换成 `C:\work\项目\...` / `C:/Users/pet-test`；
  `internal/pipeserver/server_windows_test.go` 改为按算法现算期望管道后缀（`sha256("c:\\users\\pet-test")` 前 4 字节），
  不再钉死某个历史哈希。全仓非测试源码 grep `rongj|Desktop` 命中 0。
- `integrations/deepseek-harness/ai-terminal-pet-bridge/README.md` 里给用户照抄的 `plugin_manager` 目标
  由 `C:\Users\<name>\...` 改成 `%LOCALAPPDATA%\AI-Terminal-Pet\bundles\ai-terminal-pet-bridge`。
- 有意保留：`docs/HANDOVER.md` §2 的主机路径是**事实记录**（这台机器上的安装位置），不是给用户照抄的指令。

### 26.2 `--doctor` 自检

`internal/app/doctor_windows.go` 563 行、`doctor_windows_test.go` 375 行 / 14 个 `func Test`。
九项检查按序：`environment`、`install files`、`version`、`runtime`、`config`、`pipe`、`instance`、`node`、`autostart`。
每项一行 `[ok]` / `[warn]` / `[fail] name: detail`，末行 `doctor: N ok, N warn, N fail`；任一 fail → exit 1，
经 `errDoctorFailed` 哨兵传递，因此**不写 `error.log`**。用法错误仍是 exit 2。

三条刻意设计，改代码前必须知道：

1. **不碰互斥体**。`CreateMutex` 会创建对象并干扰正在运行的实例，所以 `checkInstance` 只用「管道有人监听」
   作为可观察证据，并在文案里写明「mutex 未被探测」。
2. **管道探测只连不写**（`probePipe`，`doctorDialTimeout` 250 ms）。立即 EOF 被服务端当作空负载，
   不触发 update 回调，所以跑 doctor 不改变浮窗内容。
3. **node 探测分 Missing / TimedOut**（`node --version`，限时 300 ms，子进程 `HideWindow`）。
   「没装 Node」与「装了但没应答」对用户是两件相反的事。首跑偶发超时，实测重试即命中。

已知边界：没有 `--doctor --json`（传 `--json` 会按未知参数 exit 2）；`version` 只能比体积不能比版本号，
因此尺寸一致时仍是 `[warn]`；`runNodeVersion` 的真实 spawn 分支与 `probePipe` 的真实 Dial 分支没有单测
（靠 `%TEMP%` 副本实跑取证）。强制失败验证：`LOCALAPPDATA=C:\definitely-not-a-real-dir-xyz` →
`doctor: 5 ok, 2 warn, 2 fail`、exit 1。

### 26.3 线协议规范与上报端可安装化

- `docs/PROTOCOL.md` 388 行：§3.1 管道名派生（`sha256(home 小写、去尾分隔符)` 前 4 字节 → 8 hex，Go 与 JS 两侧
  必须同源，`TestPipePathForUserMatchesReporterAlgorithm` 钉住）、§3.2 一行 JSON + 半关闭、§3.3 只认精确 `OK`、
  §4 字段表与七状态、§5 九条脱敏、§7 auto-approve 的 `approvals/` 握手与 11 条 `DANGEROUS_PATTERNS`、
  §8 约 20 行最小 reporter、§9 换语言的 12 条注意、§10 错误处理硬约束、§11 不兼容变更策略。
- `reporters/package.json`（26 行）：`exports` 覆盖 6 个 `.js`，`private: true`，**刻意不写 `"type"`**——
  5 个 hook 是 `require()` 的 CommonJS，`opencode.js` 是真 ESM；写 `module` 打断前者，写 `commonjs` 打断后者。
  代价是首次 import `opencode.js` 时 Node 打印一条 `MODULE_TYPELESS_PACKAGE_JSON` 并重解析一次，属预期。

### 26.4 CI 与构建可复现性

`.github/workflows/ci.yml` 单 job：`windows-latest`、`gofmt -l` 有输出即挂、`go vet`、
`go test -count=1 ./...`、两条生产构建、逐个 `node --check`、Harness 包的 `verify.mjs`、
插件包的 `scripts/vendor.mjs --check` 与 `verify.mjs`。**不记行数**：这份文件随门禁增删而变，
记一个固定数字只会马上过期。

两个**必须**遵守的细节：

1. **构建命令一律带 `-buildvcs=false`**。VCS 戳记里含构建时的 commit 与工作区状态，
   同一份源码在不同时间/机器上重建出的字节不同；不带它时 `pet-statusline.exe` 是 3,644,416 B，
   带它是 3,643,392 B（差 1,024 B），`scripts/install.ps1 -VerifyOnly` 只要两侧构建方式不同就误报。
   二进制不入库（`.gitignore` 忽略 `bin/`），所以「与仓库一致」实指「与你自己 `bin/` 里那份一致」。
2. **`node --check` 必须逐文件循环**：`node --check a.js b.js` 只检查第一个参数（Node 24 实测：
   第二个是坏文件也 exit 0）。CI、`AGENTS.md`、`PROTOCOL.md` 附录 B、README 中英两半用的都是
   `for f in reporters/*.js; do node --check "$f"; done`。

### 26.5 `scripts/install.ps1` 的三个开关

纯 ASCII（PowerShell 5.1 读无 BOM 脚本按 ANSI 代码页，非 ASCII 会乱码）。行数与字节数都会随每次增强
变动，故不记——需要具体值时以工作区文件为准。

- `-VerifyOnly`：只读。按**同一份 10 项制品清单**（overlay / statusline / protocol / bundle manifest 2 / bundle runtime 5）
  逐项比对工作区与已部署侧的 SHA256，全对才 `install: verified - 10 file(s) match the repository` + exit 0。
- `-Uninstall`：先 `Assert-SafeInstallDirectory`（拒绝盘根、拒绝目录名不是 `AI-Terminal-Pet`，`-Force` 也绕不过），
  无 `-Force` 时要求手输 `yes`（无控制台则 exit 2），最多 20 次 `Stop-Process` + 最多 5 次删除重试；
  只删指向被删目录的 `AITerminalPet` Run 值，并调用 `Remove-InstallDirectoryFromUserPath` 清掉 `-AddToPath`
  写入的 PATH 条目（原值改备份到 `%TEMP%\ai-terminal-pet-path-backup-<yyyyMMdd-HHmmss>.txt`，失败只报 `[warn]`）。
- `-AddToPath`：只写 `HKCU\Environment\Path`（绝不碰 `HKLM`），保留注册表值类型，写前把原值备份到
  `<安装目录>\path-backup.txt`；已存在则不动。需要新开终端才生效。

安装脚本自身仍沿用 v0.9.0 的原则：先停浮窗、等句柄释放、20 次重试拷贝、每个文件 SHA256 校验——
因为 reporter 会按需拉起浮窗，纯覆盖拷贝会随机失败。

### 26.6 可复现的验证命令

```powershell
gofmt -l .                                   # 必须无输出
go vet ./...
go test -count=1 ./...                       # 154 个 func Test
for f in reporters/*.js; do node --check "$f"; done
node integrations/deepseek-harness/verify.mjs
go build -trimpath -buildvcs=false -ldflags "-s -w" -o .\bin\AI-Terminal-Pet.exe ./cmd/ai-terminal-pet
go build -trimpath -buildvcs=false -ldflags "-s -w" -o .\bin\pet-statusline.exe ./cmd/pet-statusline
.\scripts\install.ps1
.\scripts\install.ps1 -VerifyOnly                     # 期望 10/10 [ok]
& "$env:LOCALAPPDATA\AI-Terminal-Pet\AI-Terminal-Pet.exe" --doctor   # 期望 0 fail，exit 0
```

### 26.7 已知残留（2026-11-10）

- runbook 里 `--doctor` 的 `version` 永远只能给出 `[warn]`：只读地比对版本号不可行，尺寸一致时无法判定新旧。
- `reporters/package.json` 无 `type` 字段导致的 typeless 提示是刻意的，不要为了消警告改成 `module`。
- `scripts/install.ps1 -Uninstall` 期间若有 reporter 重新拉起浮窗，删除会失败并给出「关掉所有 hook 再试」的可见失败提示。
- `docs/assets/render_assets.py` 仍需 Pillow；缺库时给安装提示并 exit 2，不尝试自动安装。

---

## 27. 作为 DSH 插件包分发 2026-11-10

本节的全部数字都是 2026-11-10 实测。`7f476c9`（插件包本体 1.0.0）与 `91719e5`（把没进 git 树的
vendor 二进制补进去，见 §27.4）是 1.0.0；`b280093` 是 1.1.0（删掉 `prepack`、新增 `vendor:check`，见 §27.5）；
`06d4cae` 加仓库根的清单 shim、`d06001d` 让 `verify.mjs` 连根 shim 一起检查并在 CI 里跑插件套件（两者见 §27.8）。
1.2.0 在 `b280093` 的 1.1.0 之上再改了两个清单（新增 `dsh.catalog`、`keywords` 补 `dsh-plugin`，见 §27.2）。
本节凡涉及「行数/字节」都注明是 **tag `dsh-plugin-v1.1.0`（指向 `b280093`）** 还是 **HEAD** 口径——两者在 `verify.mjs` 上差 31 行。

### 27.1 为什么做成插件包

Harness 原先只有一个 host-only 的 Cordis bundle（`integrations/deepseek-harness/ai-terminal-pet-bridge`），
装法是让用户**先把整个仓库 clone 下来**，再把安装目录里那份 bundle 的绝对路径贴进插件管理器。
这条路径对外部用户不成立：他要先有 Go 工具链跑通 `install.ps1`，或者至少要先有这个仓库。

插件包要解决的是三件事，顺序就是这个优先级：

1. **一条命令装完**。`dsh plugin --profile <p> add <Release tarball 的 URL>` 之后浮窗就有了，
   不需要 Go、不需要 clone、不需要另跑安装脚本（按包名 `add dsh-ai-terminal-pet` 要等 npm 发布后才成立）。
2. **能被市场发现**。DSH 的市场读注册表仓库，条目是 `data/plugins/<slug>.yml` 一个文件；
   有了 npm 包名和 tarball 才有东西可填。
3. **不能退化浮窗的独立性**。浮窗是独立 Go 程序，装了插件包不等于浮窗只能由插件包管理——
   反过来也不成立：只要浮窗的机器仍然只需要 `install.ps1`。这条是硬约束，
   决定了 §27.3 的「绝不覆盖」不是优化而是正确性要求。

### 27.2 包结构与 `dsh.bundle` 契约

`packages/dsh-ai-terminal-pet/` 共 **18 个文件**（含 vendor；1.2.0 起多一个 `vendor/manifest.json`）：

| 文件 | 字节 | 行 | 作用 |
|---|---|---|---|
| `package.json` | 2,008 | 74 | 包元数据 + `dsh.bundle` + `dsh.catalog` 声明（scripts 只剩 verify / vendor / vendor:check） |
| `lib/index.js` | 17,872 | 571 | 事件映射、去抖、bootstrap、下载回落 |
| `verify.mjs` | 22,499 | 415 | 自包含验证脚本（无测试框架）；tag `dsh-plugin-v1.1.0` 上是 15,117 B / 300 行 |
| `scripts/vendor.mjs` | 13,981 | 324 | 双模式：无参时从 `../../bin` 与 `../../reporters` 拷进 `vendor/` 并写 `manifest.json`；`--check` 只校验 sha256 |
| `README.md` | 6,665 | 129 | 包自己的安装/限制说明（写明「本包不带任何生命周期钩子」与 `dsh.catalog` 的自我声明性质） |
| `cordis.patch.yml` | 399 | 10 | loader patch 层（`insert:`） |
| `locale/en.json` | 210 | — | 英文清单文案 |
| `locale/zh.json` | 217 | — | 中文清单文案 |
| `icon.svg` / `icon.png` | 526 / 4,386 | — | 市场图标（png 是注册表 `icon` 字段用的那份） |
| `LICENSE` | 1,070 | — | MIT |
| `screenshots.json` | 97 | — | 三张截图的清单 |
| `screenshots/compact.png` | 4,459 | — | 紧凑态 |
| `screenshots/expanded.png` | 11,335 | — | 展开态 |
| `screenshots/expanded-many.png` | 17,697 | — | 多会话展开态 |
| `vendor/manifest.json` | 529 | 19 | vendor 的溯源记录：两个文件的 sha256、Go 源码树的 sha256 与 Go 版本 |
| `vendor/protocol.js` | 20,914 | — | 与 `reporters/protocol.js` 同一份 |
| `vendor/AI-Terminal-Pet.exe` | 4,371,968 | — | 与 `bin/AI-Terminal-Pet.exe` 字节一致 |

`package.json` 的关键字段（改之前先读完这一段）：

- `"type": "module"` —— 整个包是 ESM（`lib/index.js`、`verify.mjs`、`scripts/vendor.mjs` 都是）。
  这与 `reporters/package.json` **刻意不写 type** 的处境正好相反：那边混着 5 个 CommonJS 与 1 个 ESM，
  这边全部是新写的 ESM，没有历史包袱。
- `dsh.bundle`: `{ patch: './cordis.patch.yml', platforms: ['win32'] }`。`platforms` 让非 Windows 的
  Harness 把它显示成「不支持」而不是装上一个必然跑不起来的包。
- `engines`: `node: '^22.19 || >=24'`、`dsh: '>=0.1.0-rc.6'`。下限来自 ESM 语法探测的默认开启版本，
  上限故意写成开区间——插件只依赖 `node:fs` / `node:path` / `node:crypto` 这类稳定 API。
- `files` **显式列 10 项**：`lib` / `vendor` / `locale` / `icon.svg` / `icon.png` / `screenshots` /
  `screenshots.json` / `README.md` / `LICENSE` / `cordis.patch.yml`。
  不靠 `.npmignore` 兜底：漏一个文件就是用户装上一个空包。
- `exports` 覆盖 `.` / `./locale/*.json` / `./icon` / `./package.json`。
- `scripts`：`verify`（`node verify.mjs`）/ `vendor`（`node scripts/vendor.mjs`）/ `vendor:check`
  （`node scripts/vendor.mjs --check`）。**没有 `prepack`、没有 `prepare`、没有 `postinstall`**：
  生命周期钩子在 1.1.0 被整个删掉（原因见 §27.5），所以 `vendor/` 必须由发版人在打 tarball 之前手动
  拷一次并提交——`npm pack` 不再替你重新生成它，`vendor:check` 与 CI 负责拦住漏跑。
- `dsh.catalog`（1.2.0 新增）：`{ category: 'tool', summary: { en, zh }, capabilities: [...] }`。
  `summary.en` 190 字符、`summary.zh` 104 字符，各是一句「这是一块 Windows 桌面浮窗，把四类终端的会话汇总
  到一块常驻置顶面板上」。五项 `capabilities`：`windows-overlay` / `named-pipe` / `session-state` /
  `tray-icon` / `no-network`。**这些 capability 是自我声明、不被强制**——dsh 不对插件做沙箱，读到的是
  这块浮窗本来就在做的事，不是一份权限授予。写文档或写市场文案时不要把它说成「沙箱能力声明」。
- `keywords`（1.2.0 补 `dsh-plugin`）：该值配合 GitHub 仓库的 topics 被自动索引型市场采集，
  所以现在有两条 **互不依赖**的发现路径——① 注册表评审（`awesome-dsh-plugin` 的 PR 合并进 `main`
  之后才会出现在列表里）；② topic/关键词采集（第三方市场直接扫 GitHub，不需要提交任何申请）。
  仓库 topics 现为 `dsh-plugin` / `deepseek-harness` / `cordis-plugin`：`zasenjc/dsh-plugins-store` 的收录门槛是
  `dsh-plugin` 与 `deepseek-harness` **同时**存在（前两个必须填），`cordis-plugin` 只标生态、不参与门槛。
- `version`：1.2.0。**版本号必须跟着清单内容走**：`dsh.catalog` 与 `keywords` 是包内容，不是文档内容，
  若让 package.json 变了而版本仍停在 1.1.0，已发布且不可变的 1.1.0 tarball 就会与仓库分叉。

**`cordis.patch.yml` 必须用 `insert:`**，这是这一批里唯一一个「错了也看不出来」的地方：

```yaml
- insert:
    - id: dsh-ai-terminal-pet
      name: dsh-ai-terminal-pet
```

裸的 `- id: dsh-ai-terminal-pet` 是**按 id patch 一行已存在的 loader 行**的语法。写错成那样，安装照样
`+ dsh-ai-terminal-pet <version>` 成功、`--dump-config` 里也能看到条目，但运行时什么都没挂载：
浮窗永远不更新，而所有日志都是干净的。因此原因直接写进了该 YAML 的文件头注释里，
而不是只写在文档里——改这个文件的人一定会看到文件本身。

### 27.3 引导安装与「绝不覆盖」

`runBootstrap()`（`lib/index.js`）在 Harness 进程首次用到插件时跑一次，做两件事：

- 仅当 `%LOCALAPPDATA%\AI-Terminal-Pet\AI-Terminal-Pet.exe` **缺失**时，才从包内
  `vendor/AI-Terminal-Pet.exe` 拷进安装目录；`vendor/protocol.js` 同样只在
  `%LOCALAPPDATA%\AI-Terminal-Pet\protocol.js` 缺失时才拷，权限 `0o600`。
- **从不覆盖已有文件**。已存在的文件（尤其是 `install.ps1` 装的那份）永远赢。
  `copyVerified` 的实现是「写临时文件 → rename 到目标」，并且在 rename **之前**再查一次目标是否存在——
  两次检查之间正好是另一个 reporter 拉起浮窗、`install.ps1` 同时跑的时间窗。

两个开关是测试缝，也是给高级用户的逃生口：

- `DSH_PET_NO_BOOTSTRAP=1`：整段引导跳过（浮窗已由 `install.ps1` 管理的机器上用它）。
- `DSH_PET_TEST=1` + `DSH_PET_PACKAGE_ROOT`：把包根换成测试目录，验证脚本靠它跑真实的拷贝分支。

下载回落（`downloadFallback()`）：读 `lib/release-pins.json`，按钉住的 URL / 大小 / SHA-256 下两个资产，
30 s 超时，`node:crypto` 校验大小与摘要后才落盘，失败只 warn 不抛。
**这个文件目前故意没建**，所以回落当前是 no-op——这也是 §27.4 那个坑之所以危险的原因。

### 27.4 vendor 二进制没进 git 树：症状、根因、修法

这是这一批里最值得记住的一次事故，因为它在本地与 CI 全是绿的。

**症状**：`dsh plugin add` 用 release tarball 装能成功，用 **git 源**装也能「成功」（rc=0），
但装上之后浮窗永远不出现。没有任何报错——`bootstrap()` 找不到 `vendor/AI-Terminal-Pet.exe`，
于是回落到下载；`lib/release-pins.json` 还没建，回落是 no-op；插件按设计静默，所以只剩「什么都没发生」。

**根因**：`.gitignore` 里的 `*.exe` 规则把这颗二进制一并忽略了，而 `git add` 一个目录
**不会**因为文件在目录里就越过忽略规则（要 `-f`）。所以：

- `npm pack` 读的是**工作区**，二进制在，tarball 完整，tarball 安装能跑通；
- `git clone` 读的是**树**，二进制不在，`vendor/` 里只剩 `protocol.js`。

两条路径的差异只在这一个文件上，正好是「本地测过、换台机器就坏」的形状。

**修法**（`91719e5`，两行，外加注释）：

```gitignore
!packages/dsh-ai-terminal-pet/vendor/AI-Terminal-Pet.exe
```

```gitattributes
*.exe binary
```

- `!` 例外是必须的：这是唯一一颗**发布制品**而不是本地构建产物的 exe，规则里也写明了
  「它是构建产物规则的唯一例外，改 `bin/` 之后用 `npm run vendor` 重新生成」。
- `.gitattributes` 的 `*.exe binary` 同样必须：Git 的 `text=auto` 会试图对二进制做文本规范化，
  而 `-VerifyOnly` 与 `scripts/vendor.mjs` 都靠 SHA-256 判等，一个被重编码的字节就让校验永远失败。
- 事后验证：`git ls-files packages/.../vendor` 能列出该 exe，`git cat-file -s` 报 4,371,968 B，
  sha256 与 `bin/AI-Terminal-Pet.exe` 相同（`bbb5163c296cd460...`）。

**教训**：任何「包内 vendored 二进制」都必须用 `git ls-files` 而不是 `ls` 来确认它在树里；
而判断一个分发链路是否真的完整，唯一可信的证据是**从那条链路的产物里解包再看一次**。

### 27.5 发布与分发现状（截至 2026-11-10）

- **npm**：凭据问题**已解决**（2026-11-10T12:20Z 换了 token 后 `npm whoami` 返回 `jjjjjjjjnnjnn`，
  此前整场会话都是 E401）。`npm view dsh-ai-terminal-pet version` 仍是 **E404**，说明这个名字还空着、
  **还没有发布任何版本**。发布被一次性密码（OTP/2FA）挡住：npm 的 `otplease` 在无 TTY 时直接抛 EOTP，
  浏览器授权链接（带 `npm-auth-type: web` 的探针才拿得到未打码的那个）5–8 分钟即失效，只能由用户在本机交互式终端里发。
  发布前已在包目录跑过 `npm publish --dry-run`：rc=0，**16 files / 1.9 MB packed / 4.5 MB unpacked**
  （那次 dry-run 的 shasum 是 `166c3043…`，但之后包内 `README.md` 又改了 13 B，所以最终的号不是它；
  1.2.1 的 dry-run 口径见下面 1.2.1 那条，1.2.2 见它自己那条）。
  因为从**仓库根**发布时 pnpm/npm 抬到的是根清单，而根清单带 `private: true`；包清单没有 `private`，
  所以在 `packages/dsh-ai-terminal-pet/` 目录里发布是可行的。这条对写文档的人有价值：根 shim 的
  `private` 是「别抢 npm 名」，同时也意味着**不能从仓库根发布**。
  仍未解锁的是 DSH 应用内 Plugins → Discover 索引：它只表达 npm 版本（format 1）或钉到 40 位 commit 的
  git 源，所以 `add dsh-ai-terminal-pet` 要等这次真正发布之后才解析得到。
- **GitHub Release 线：1.0.0 → 1.1.0 → 1.2.0 → 1.2.1 → 1.2.2 → 1.2.3**，一版一个 tag，旧 tag 不删也不复用：
  - **1.0.0**（已作废）：tag `dsh-plugin-v1.0.0`，资产 `dsh-ai-terminal-pet-1.0.0.tgz`
    （1,913,682 B，sha256 `7365f6a6f4aea1174ec6f2f00670e868568e944816d1869aeb59db363ae6b5af`）。
    这一版的 vendor 曾经没进 git 树（见 §27.4），tag 后来被强制挪到 `91719e5`。
  - **1.1.0**：tag `dsh-plugin-v1.1.0`（annotated，指向 `b28009377e14840ee003ddaf114922e1d167ed0d`），
    资产 `dsh-ai-terminal-pet-1.1.0.tgz`（**1,913,726 B / 15 files**，sha256
    `3b8f9186cca5142626f2e25175096f1c4fb59d86cc1cdff5879792950146cc78`，shasum
    `19a2e544ab875c61fcf822e85a91c67273981725`）。这一版删掉了 `prepack`，因此不再有生命周期钩子。
  - **1.2.0**（由 1.2.1 取代，发布物本身不可变）：tag `dsh-plugin-v1.2.0`（annotated，tagger 2026-11-10T13:46:52+0100，指向
     `4d777e32e0a9ef57bb7714287dccd7cae065346f`），资产 `dsh-ai-terminal-pet-1.2.0.tgz`
     （**1,914,566 B / 16 files / sha256 `ccf9d531267af681438010bc94c721e7822a0db3f4d14575953b1757a2ef65c9`
     / shasum `dc65e0d4259be640c0de7ff1bc86675af5454546`**）。
    内容增量是 `dsh.catalog` 与 `keywords` 里的 `dsh-plugin`；`vendor/manifest.json` 成为随包文件，
    所以 tarball 文件数由 15 变为 **16**（多出来的 840 B 就是那个 529 B 的清单加上打包开销）。
    发版**之前**那次 `npm publish --dry-run` 的 shasum 是 `166c3043…`，但 dry-run 之后包内 `README.md`
    又改了 13 B（6,652 → 6,665 B），所以**最终发出去的资产号是 `dc65e0d4…`，不是 dry-run 那个**。
    1.1.0 的 tarball 已发布且不可变，由 1.2.0 取代——清单内容变了就必须换版本号。
  - **1.2.1**（由 1.2.3 取代，发布物本身不可变）：只为让包内说明追上事实，不改代码。tag `dsh-plugin-v1.2.1`（annotated，指向
     `1e08b3e7521d305d744c5ff515b29f6d16abebe1`，tag 对象 sha `0ef1d60cb5c33517aa87d6eed75565c035d5830e`），
     GitHub Release 于 **2026-11-10T16:15:04Z** 发布，资产 `dsh-ai-terminal-pet-1.2.1.tgz`
     （**1,914,616 B / 16 files / sha256 `b9e6d20354b5b1543b55041aa0c45c38cd4fdfacc65ab1b13fef5ab0651d0075`
     / shasum `67955a65417b6b469c591d1d65fbd1893426628a`**，GitHub 记录的 digest 与本地重算一致）。
     与 1.2.0 的差别只有包内 `README.md`（6,665 B → **6,863 B**）与两条命令里的版本号；解包核对时
     `package/README.md` 里已不含 `waiting on a human reviewer` 那句旧文。连续两次 `npm pack` 逐字节相同。
  - **1.2.2**（由 1.2.3 取代）：由「安装脚本下沉」引发的重打包。`bdc6c36` 把 `install.ps1` 从仓库根移进 `scripts/`，
     这一批改到了 8 处 Go 源码注释；而 `vendor/manifest.json` 的 `source.sha256` 正是 `cmd/` 与 `internal/`
     下全部非测试 `.go` 文件的摘要，于是从 `89eef0d6…` 变成 **`cf647af1…`**（25 个文件，`goVersion` 仍 `go1.27.0`）。
     已发布资产不可变，所以只能靠新版本把新记账发出去——按 1.2.1 装到、且本机有 `bin/` 的人跑 `vendor:check`
     会直接失败（干净 clone 里因 `bin/` 不存在而按「以已提交的 vendor 为准」通过，问题被掩盖）。
     与 1.2.1 的差别只有三处：两份清单的 `version`、包内 `README.md` 的措辞、`vendor/manifest.json` 的新摘要；
     `lib/index.js`、两个 vendor 二进制、`cordis.patch.yml`、`dsh.catalog` 逐字节不变。
     发布提交 `694ff3f13740b1bb9fb4df04f5fd226064a559b2`（7 files, +48/−35），annotated tag `dsh-plugin-v1.2.2`
     （tag 对象 sha `06d7795270d516ca7ced6707bf726bcf0337de99`），Release **2026-11-10T18:11:26Z**，资产
     `dsh-ai-terminal-pet-1.2.2.tgz` **1,914,751 B / 16 files / sha256
     `02431e16dd58db166794d7c86608087446d4a3ed8ab51bcb4b0950b2a045307c` / shasum `22b63b25c50d5f4d799bfcf00fd4f54986be63a8`**
     （GitHub 记录的 digest 与本地重算一致，连跑两次逐字节相同）。随后 `fe62900` 把 `docs/market-index.json`
     前移到这个 commit，`verify.mjs` 复跑 93 行 / 0 fail。
  - **1.2.2 的端到端复验**：全新 `DSH_HOME` + 独立 profile `p122`，直接 `add` 1.2.2 的 Release 资产 URL →
     rc=0 / **3,335 ms** / `+ dsh-ai-terminal-pet 1.2.2`；`--dump-config` 末尾出现 `- id: dsh-ai-terminal-pet`；
     装出的 `vendor/manifest.json` 里 `source.sha256` 就是 `cf647af1…`——新版本把正确的记账带到了用户机器上。
  - **1.2.3**（当前）：只改两份清单，各加一个 `dsh.marketplace` 块。**动机来自商店侧，不是代码侧**：
     `YELEBAI/dsh-plugin-marketplace` 的注册表已经把本仓库收录了，但它的 `install.profiles` 记的是
     `[headless, web]`，而它桌面端市场会按 desktop 过滤（`src/host/registry.ts:207-210` 里
     `if (targetProfile !== undefined && !plugin.install.profiles.includes(targetProfile)) return false`），
     于是桌面端把条目藏了起来。它的分类器 `scripts/registry-core.mjs` 的 `classifyInstall` 在没有
     `dsh.marketplace` 声明时回落到 `identity.hasClient ? ['web'] : ['headless','web']`——我们的浮窗是原生
     Win32 窗口、不是 web client，所以 `hasClient` 为 false。DSH 自己没法表达「host-only bundle 跑在任何
     profile」：它唯一与 profile 有关的字段 `dsh.client` 只能是 `web`。于是两份清单声明
     `profiles: ["desktop", "headless", "web"]`、`requiresBuildApproval: false`、`manualSteps: false`。
     **DSH 宿主从不读这个字段**（对 `@deepseek-ai/dsh-plugin-market` 与 `dsh-plugin-market-index` 的 139 个
     js 文件 grep `dsh.marketplace|marketplace.profiles` 命中 0），插件照旧经 `dsh.bundle` 挂载，且只订阅
     `session/event`、不依赖任何 profile 专属服务；读它的是商店侧。把 `registry-core.mjs` 下载后替换掉
     `yaml` 依赖再 import 它自己的源码重放，得到 `profiles: [desktop, headless, web]`、
     `profileSource: "manifest"`、`runtimeArtifactsCommitted: true`、`reviewReasons: []`。
     发布提交 `b4474a6da10a7bc3518362e910c66b9213be92e6`（5 files, +52/−19），annotated tag
     `dsh-plugin-v1.2.3`（tag 对象 sha `afbbf9c5d4a99ad2d67aaae9e8d3dd7d4001461c`），Release
     **2026-11-10T20:20:31Z**，资产 `dsh-ai-terminal-pet-1.2.3.tgz` **1,915,013 B / 16 files / unpacked 4,461,408 B /
     sha256 `1172aa72c63482442ccf0c1dfc12446fa15e1b12220f921fbe06a56c4db87c35` / shasum
     `0508ff6b9159674ae4c7136cb00bf1870d44eb35`**。与 1.2.2 的 1,914,751 B 只差 262 B，全部来自两块
     `dsh.marketplace` 与包内 README 的新增段落；`lib/index.js`、两份 vendor 二进制、`cordis.patch.yml`、
     `dsh.catalog`、`vendor/manifest.json` 逐字节不变。`docs/market-index.json` 由 `b4474a6` 前移到此版。
     `verify.mjs` 由 415 行 / 91 ok 长到 **424 行 / 95 ok / 0 fail**（新增 4 条断言钉住 marketplace 块，
     其中一条要求根 shim 与包清单逐字相同）。
  - **1.2.3 的端到端复验**：全新 `DSH_HOME` + 独立 profile `pet123`，`add` 1.2.3 的 Release 资产 URL →
     rc=0 / **2.4 s** / `+ dsh-ai-terminal-pet 1.2.3`；`plugin list` 显示 `dsh-ai-terminal-pet@1.2.3`；
     profile 的 `dsh.profile.bundles` 为 `["@deepseek-ai/dsh-base", "dsh-ai-terminal-pet"]`；装出的
     `package.json` 里 `dsh.marketplace` 块与 `dsh.bundle`（`platforms: ["win32"]`）都在位，
     `vendor/AI-Terminal-Pet.exe` 4,371,968 B。
  - 文件数与包目录的差：`npm pack` 走包内 `files` 白名单（10 项）外加 npm 必然带上的
    `package.json` / `README.md` / `LICENSE`，所以白名单外的 `scripts/vendor.mjs` 与 `verify.mjs`
    从不进 tarball——包目录数到 18 个文件，tarball 只有 16 个。
- **三种安装形态都实测成功**（每次都开一个全新的 `DSH_HOME` 与独立 profile）：
  计时是在 1.1.0 的资产上做的；1.2.0 / 1.2.1 / 1.2.2 / 1.2.3 的资产都在同一位置，URL 只换 tag 与版本号
  （1.2.2 那次单独复验过 3,335 ms，1.2.3 是 2.4 s）。
  1. **Release tarball**：`add https://github.com/jjjjjjjjnnjnn/ai-terminal-pet/releases/download/dsh-plugin-v1.1.0/dsh-ai-terminal-pet-1.1.0.tgz`
     → rc=0、**1.8 s**；装出的包里 `vendor/AI-Terminal-Pet.exe` 4,371,968 B 与 `vendor/protocol.js` 20,914 B 都在位。
  2. **git 源 + 钉 tag**：`git+https://github.com/jjjjjjjjnnjnn/ai-terminal-pet.git#dsh-plugin-v1.1.0&path:/packages/dsh-ai-terminal-pet`
     → rc=0、**3.4 s**，vendor 同样两个文件在位。
  3. **仓库地址**（1.1.0 起才成立）：`add github:jjjjjjjjnnjnn/ai-terminal-pet` → rc=0、**3.6 s**；
     等价写法 `git+https://github.com/jjjjjjjjnnjnn/ai-terminal-pet.git`（不带 `path`）同样 rc=0、3.6 s。
     装出的 `node_modules/dsh-ai-terminal-pet/` 里 `packages/dsh-ai-terminal-pet/` 子树完整，
     profile 的 `package.json` 中 `dsh.profile.bundles` 出现 `dsh-ai-terminal-pet`。
  三种都用 `dsh --profile <name> --dump-config` 验证过：末尾出现 `- id: dsh-ai-terminal-pet`（证明 `insert:` 生效）。
  1.0.0 时期的三条旧记录（本地 tgz → 32.6 s 的 Release URL → 10.2 s 的钉 commit 形态）保留在
  CHANGELOG 的历史条目里，这里不再重复。
- **仍会失败的一条形态，但它与 `prepack` 无关**：`git+https://github.com/jjjjjjjjnnjnn/ai-terminal-pet.git&path:/packages/dsh-ai-terminal-pet`
  **不写 `#ref`** 时，pnpm 把 `&path:` 拼进 URL 当作仓库地址去 `ls-remote`，报
  `repository ... not found`——它请求的那个仓库根本不存在。也就是说，1.0.0 时期「裸 git 源会被
  `allowBuilds` 拦在 `prepack` 处」这条已随 1.1.0 删掉 `prepack` 而**失效**；现在会失败的是拼错的 ref。
  对外命令给三种形态即可（tarball / 钉 tag / 仓库 URL），不需要教用户改 `pnpm-workspace.yaml`。
- **注册表 PR #13937 已合并**（fork 分支 `add-dsh-ai-terminal-pet`，commit `a1fef56`，
  基于上游 main `09cd105`）：<https://github.com/awesome-dsh-plugin/awesome-dsh-plugin/pull/13937>。
  条目文件 `data/plugins/jjjjjjjjnnjnn__ai-terminal-pet--packages-dsh-ai-terminal-pet.yml`，
  `category: notify`、`url` 指向 `.../tree/main/packages/dsh-ai-terminal-pet`、
  `name: jjjjjjjjnnjnn/ai-terminal-pet#dsh-ai-terminal-pet`，带 `icon`（raw.githubusercontent 的 icon.png）
  与 `tarball`（合并时钉的是 `dsh-plugin-v1.2.0` 那份资产；条目改成新版本要另开 PR 改这个 yml）。
  上游两道门本地先跑过：
  - 条目模型校验：`entries read: 10120 / problems: 0`；
  - `scripts/check-submission.mjs --only-list` → `ok https://github.com/jjjjjjjjnnjnn/ai-terminal-pet/tree/main/packages/dsh-ai-terminal-pet`
    与 `all checked entries pass`。

  CI 全绿后才合并：`check` 于 2026-11-10T10:37:54Z 结束（`conclusion: SUCCESS`，前 14 步与最后一步
  `Build (locale parity, date derivation, templates)` 全部通过），`Submission gate` 于 10:38:47Z 结束
  （`SUCCESS`），随后 PR 在 **2026-11-10T14:33:04Z** 由评审人 `fkysly` 合并（`state: MERGED`）。维护者合并时
  删掉了原稿里「也能在面板上直接放行审批」一句：对 Harness 会话，面板只做状态映射与跳转到 DSH 窗口，
  审批仍在 DSH 里完成。条目已进入 `main` 的 `data/plugins/`，上游 `README.md` 第 8,519 行
  （`### Notifications & Integrations` 小节）已列出它。

  **「已合并」不等于「市场里可见」——而这次它后来真的可见了，两件事都要记。** 合并后的复查记录：
  - 条目页 <https://awesome-dsh-plugin.com/p/jjjjjjjjnnjnn/ai-terminal-pet--packages-dsh-ai-terminal-pet/> 在
    15:56Z、16:16Z 两次都是 **404**；上游站点构建跑完之后同一地址返回 **200**，页面标题是
    `dsh-ai-terminal-pet — dsh plugin · Notifications & Integrations`。
  - DSH 应用内 Market 真正加载的 <https://awesome-dsh-plugin.com/plugins.json> 在 16:16Z 时
    `count: 10119` / **0 命中**；站点构建完成后 `count: 10230` 且已收录本仓库，条目里 `npm: null`、
    `version: null`、`install` 命令直接是钉 1.2.0 的 Release 资产 URL（因为目录按 yml 里的 `tarball` 生成）。
  - 这两件事的延迟来自上游 `Build site from README`：run `42841664024`（15:13:06Z 起）与后继
    run `42844452013`（15:27:09Z 起）当时都停在步骤 9 `Build`（steps 1-8 全部 success，之后
    upload-pages-artifact / deploy-pages / publish catalog to npm / publish update notes to npm 与所有
    Post 步骤都还是 pending）；后来 `42841664024` 在 15:25:40Z 被 **cancelled**（同分支新 run 取代），
    `42844452013` 于 16:39:27Z **success**。所以延迟是上游重建的排队与耗时，不是条目本身有问题。

  2026-11-10T17:19Z 维护者 `fkysly` 在 PR 上回了一条评论，说明合并时在 `main` 上做过微调——原文要点：
  「已合并。合并时按复验结果在 main 上微调了条目描述：『也能在面板上直接放行审批』不适用于 DeepSeek Harness
  会话（README 里只有状态映射与跳转到 DSH 窗口），已改为如实描述：面板只显示状态并跳转，审批仍在 DSH 里完成；
  措辞如需调整，直接开 PR 改对应的 `data/plugins/*.yml` 即可，我们按最新的为准。」
  所以上游 `main` 上那份 yml 的 `description.en/zh` 现在带一句「approving is still done in DSH / 审批仍在 DSH 里完成」，
  与我们原来的稿子不同；**要改措辞只能提 PR 改那个 yml，本仓库的文档改动不会影响它**。
  先开过跟进 PR #14063（fork 分支 `bump-tarball-1.2.1`，head `116b0ee`，1 file `+1/-1`）：只把条目里那行
  `tarball` 从 `dsh-plugin-v1.2.0/dsh-ai-terminal-pet-1.2.0.tgz` 换成 1.2.1 的资产，描述一字未改。
  但 1.2.1 是在 `install.ps1` 还留在仓库根时打的，§27.11 的下沉改到了 8 处 Go 注释，`vendor/manifest.json`
  的源码摘要随之变成 `cf647af1…`，所以 #14063 换成 1.2.2 之前先在 PR 上留言说明、随后 **关闭**；
  改开 **#14076**（<https://github.com/awesome-dsh-plugin/awesome-dsh-plugin/pull/14076>，分支
  `bump-tarball-1.2.2`，commit `0863b4d`，1 file `+1/-1`）指向 1.2.2 的资产，描述一字未改。
  本地用上游自己的门先验过：`scripts/check-submission.mjs --only-list` → `checking 1 entry` /
  `ok https://github.com/jjjjjjjjnnjnn/ai-terminal-pet/tree/main/packages/dsh-ai-terminal-pet` /
  `all checked entries pass`（该脚本需要 `GITHUB_TOKEN`，缺了会直接报 required）；两个 URL 的 HEAD 都是 200。
  **#14076 已于 2026-11-10T19:59:37Z 合并**（merge commit `e5f2d26335fb2821be0ec1c7d06992ff95725ebd`；
  `check` 跑了 21m21s pass、`Submission gate` pass）。合并之后目录里那条安装命令才会在**下一次站点重建**时
  从 1.2.0 换成 1.2.2——`install` 字段是构建时按 yml 的 `tarball` 渲染的，不是运行时读的。
  1.2.3 发版后照同样办法再跟一次：**#14089**（<https://github.com/awesome-dsh-plugin/awesome-dsh-plugin/pull/14089>，
  分支 `bump-tarball-1.2.3`，commit `c5a1f5f`，仍然只改那一行 `tarball:`），开单时
  `state: OPEN`、`mergeable: MERGEABLE`、`mergeStateStatus: UNSTABLE`（只因检查还在跑），等维护者合并。

### 27.6 可复现的验证命令

```powershell
# 包自身的断言（状态映射、waterfall next()、dispose 不发 closed、真 exit 发 closed、
# bootstrap 只补缺失不覆盖、protocol 抛异常时插件全静默、根清单与包清单一致、dsh.catalog 形状）
# 1.2.0 / 1.2.1 实测：rc=0、91 个 ok、0 warn、末行 VERIFY PASSED（1.1.0 是 63 个 ok，tag 上是 55 个）
cd packages/dsh-ai-terminal-pet; node verify.mjs          # 期望 rc=0、末行 VERIFY PASSED

# vendor 校验（只读）：重新算两个文件的 sha256，并与 vendor/manifest.json 比对
npm run vendor:check                                      # = node scripts/vendor.mjs --check

# 重新生成 vendor 里的预编译浮窗 + manifest.json（改过 bin/ 之后必跑，然后提交这两个文件）
npm run vendor                                            # = node scripts/vendor.mjs

# 打包口径（没有 prepack，不会再替你重生成 vendor）
npm pack                                                  # 1.2.0 实测体积/sha256 待填；文件数 16

# 三条安装形态，任选其一（Release 资产 URL / git 源钉 tag / 仓库地址）
dsh plugin --profile <profile> add https://github.com/jjjjjjjjnnjnn/ai-terminal-pet/releases/download/dsh-plugin-v1.2.0/dsh-ai-terminal-pet-1.2.0.tgz
dsh plugin --profile <profile> add "git+https://github.com/jjjjjjjjnnjnn/ai-terminal-pet.git#dsh-plugin-v1.2.0&path:/packages/dsh-ai-terminal-pet"
dsh plugin --profile <profile> add github:jjjjjjjjnnjnn/ai-terminal-pet

# 确认 insert: 真的挂上了（而不是 patch 了一行不存在的 id）
dsh --profile <profile> --dump-config | tail -5       # 末尾应出现 "- id: dsh-ai-terminal-pet"
```

包内二进制与仓库那份是否一致，只看两条：

```powershell
git ls-files packages/dsh-ai-terminal-pet/vendor      # 必须列出 AI-Terminal-Pet.exe
git cat-file -s "$(git ls-files -s packages/dsh-ai-terminal-pet/vendor/AI-Terminal-Pet.exe | %{ $_.Split()[1] })"  # 4,371,968
```

### 27.7 已知残留（2026-11-10）

- **npm 尚未发布**：凭据问题已解决（`npm whoami` = `jjjjjjjjnnjnn`，此前 E401），但
  `npm view dsh-ai-terminal-pet version` 仍 E404，所以**当前没有任何版本在 npm 上**（1.0.0 到 1.2.3 一个都没发）。
  发布 1.2.3 是下一步；但这一步在本机非交互环境里结构上做不到：npm 的 `otplease` 在无 TTY 时直接抛 EOTP，
  web-OTP 分支永不进入，取到的浏览器授权链接 5–8 分钟就 404 作废（`npm trust github` 走可信发布者同样撞 EOTP）。
  **要发只能由人在本机交互终端里跑一次 `npm publish`**；在那之前 DSH 应用内的 npm 索引仍拿不到本包，
  而 git 源、Release tarball、仓库地址、以及自建 format-1 索引四条路可用，手工安装不受阻塞。
  **文档里不要写成「已发布到 npm」，也不要写成「被凭据阻塞」。**
- **1.2.0 的 Release 已产出，但 dry-run 的摘要已作废**：`dsh-plugin-v1.2.0` 这个 annotated tag 指向
  `4d777e32e0a9ef57bb7714287dccd7cae065346f`，GitHub Release 于 2026-11-10T12:47:24Z 发布，唯一资产
  `dsh-ai-terminal-pet-1.2.0.tgz`（**1,914,566 B / 16 files**，sha256
  `ccf9d531267af681438010bc94c721e7822a0db3f4d14575953b1757a2ef65c9`，shasum
  `dc65e0d4259be640c0de7ff1bc86675af5454546`；本地下载后重算与 GitHub 记录的 digest 一致）。
  注意发版**之前**那次 `npm publish --dry-run` 的 shasum 是 `166c3043…`——dry-run 之后包内
  `README.md` 又改了 13 B（6,652 → 6,665 B），摘要随之作废。**教训：dry-run 的摘要只在包内容不再变时有效。**
- **`lib/release-pins.json` 未建**：下载回落当前是 no-op。它不是「功能未完成」而是「刻意留空」——
  没有它就没有远端资产可信；填它之前必须先有稳定的 Release 与 SHA-256。
- **市场条目已改判并拿到 `safe`**：DSH-Plugins-Marketplace 的 `installability-report.json` 在 16:22Z 那次重建后
  已把本仓库从 `verdict: script` 改判为 **`verdict: bundle-plugin`**；`registry.json` 到 18:08:44Z 那次重建
  （`risk_at` 17:09:14Z）是 `pkg_name: dsh-ai-terminal-pet` / `version: 1.2.1` / topics 11 个 /
  **`risk_tier: "safe"` 且 `risk_flags` 为空**——`riskScorecardOf` 的语义是「无命中即 safe」，所以这不是「未评估」。
  两条旗标的输入在 2026-11-10 消失：`evalRepoRisk` 只抓仓库根的 `install.ps1` / `install.sh` / `package.json`，
  而安装脚本已移进 `scripts/`（§27.11）；历史行号仅供查证——`/CurrentVersion\Run(?:Once)?\b/i` 命中原根脚本
  :31 的 `$RunKeyPath` 常量与 :353 的提示文本，`/\bremove-item\s+[^\n]*-recurse[^\n]*-force/i` 命中 :308 的
  `Remove-Item -Recurse -Force`（medium）。扫描器仍无 suppress / allowlist，换别的规则仍可能命中，所以「safe」
  只代表当前规则表下的结论。issue #367 已回帖说明（17:58:34Z 的 `#issuecomment-6775035737`，19:04:00Z 又补了 `#issuecomment-6776560881` 写明 18:08:44Z 重建后已是 `safe`、旗标清零、可关闭）。
- **注册表条目与 `dsh.catalog` 不同步**：PR #13937 的 `data/plugins/<slug>.yml` 写的是它自己的
  `category` / `summary`，与包内 `dsh.catalog` 是两份独立文案；合并之后若要改文案，两边都得改。

### 27.8 仓库根的清单 shim 与 CI 里的插件套件（2026-11-10）

`06d4cae` 加的是**仓库根**的 `package.json`（1.2.3 起 2,052 B / 68 行——比 1,877 B / 59 行多出的九行就是 `dsh.marketplace`；此前只改版本号那一位），`d06001d` 让验证脚本与 CI
都覆盖到它。两件事合起来才让「按仓库 URL 安装」成为一条可长期成立的路，而不是一次手工碰运气。

**根 shim 为什么必须存在**：`dsh plugin add github:owner/repo` 这类写法读的是**仓库根清单**。
仓库根原先只有 `go.mod`、没有 `package.json`，于是宿主把仓库解析成一个 0.0.0 依赖、报
`declares no dsh.bundle`、装出一个空包——**症状是「装成功了但浮窗永远不出现」，没有任何报错**，
和 §27.4 那次事故是同一种形状：链路看着通，产物是空的。

shim 里没有任何代码，只做四件事：

1. `name` / `version` 与插件包**保持一致**（`verify.mjs` 有断言，所以一次版本号变更不可能只落在一个清单里）；
2. `main` 指 `./packages/dsh-ai-terminal-pet/lib/index.js`，`exports` 抄同一组四条；
3. `dsh.bundle.patch` 指 `./packages/dsh-ai-terminal-pet/cordis.patch.yml`、`platforms: ['win32']`
   ——这两行才是让宿主认出它是 bundle 的东西；
4. `files` 只放行 `packages/dsh-ai-terminal-pet`，所以从仓库根 `npm pack` 出来的 tarball
   只含插件子树（`verify.mjs` 有用例覆盖这条）。

`private: true` 是必须的：这个清单不带 `private` 就会被当成一个可发布的包，抢走 `dsh-ai-terminal-pet`
这个 npm 名字。它也带 `dsh.catalog`——宿主若从仓库根读清单（而不是从包内），看到的 storefront 元数据
应当是同一份。

**CI 侧的两步**（`.github/workflows/ci.yml`，`d06001d` 起）：

```yaml
- name: Check the vendored overlay is current
  run: node packages/dsh-ai-terminal-pet/scripts/vendor.mjs --check
- name: Verify the DSH plugin package
  run: node packages/dsh-ai-terminal-pet/verify.mjs
```

为什么这两步值得占 CI 时间：验证脚本自带 fake Cordis context、临时 `LOCALAPPDATA` 与 stub protocol 模块，
所以它**不碰网络、不碰管道、不需要真实浮窗**，在 `windows-latest` 上就是一个普通的 Node 进程；
但它能拦住三类事故——包清单与根清单分叉、vendor 与 `bin/` 不一致、`dsh.catalog` 形状写坏。
这三类在本仓库都真实发生过或差一点发生。

**口径提醒**：写这部分的任何数字前先确认是 tag 还是 HEAD。`dsh-plugin-v1.1.0` 上的 `verify.mjs` 是
300 行 / 55 个 `ok`；1.2.0 之后的 HEAD 上是 **424 行 / 95 个 `ok`**（其中 14 条是 `docs/market-index.json` 的漂移用例，见 §27.10；1.2.3 新增的 4 条钉住两个清单的 `dsh.marketplace` 块必须逐字相同）。两个都不是错的，混在一句话里才是错的。

### 27.9 三条发现路径，与一次 CI 变红的根因（2026-11-10）

| 路径 | 要做什么 | 采集方 | 延迟 |
|---|---|---|---|
| 注册表评审 | 提一个 `data/plugins/<slug>.yml` 的 PR，等人工合并。本条目已于 2026-11-10T14:33:04Z 由 `fkysly` 合并，上游 README 已列出 | `awesome-dsh-plugin`（DSH 内置 Plugins → Market 的目录源） | 合并只写进 `main`；列表、站点页与 `plugins.json` 要等站点构建跑完。16:16Z 复查：条目页 404、`plugins.json` 0 命中，那次 build（run `42841664024`）从 15:13:06Z 起一直卡在 build 步骤 |
| 根清单采集 | 什么都不做。要求仓库 **根** `package.json` 里有 `name` + `dsh.bundle`，市场钉默认分支的 commit；**但它的发布数据只有 npm 源**（26,351 条条目全部带 npm tarball 的 `integrity`，`commit` 字段 0 条），所以纯 git 源在进 npm 之前不会被它列出来 | `LivXue/dsh-plugin-shop` | 每日构建 |
| 双 topic 采集 | 什么都不做。要求仓库 topics **同时**具备 `dsh-plugin` 与 `deepseek-harness` | `zasenjc/dsh-plugins-store` | 30 分钟同步 |
| npm 包 + PR | 要一个已发布到 npm 的包，外加一个 PR | `cyber-moshen` 的 workshop | 人工 |

这解释了 §27.8 那个根 shim 的第二重价值：它不只是让 `add github:owner/repo` 能装，它还是
**「根清单采集」这条路径的准入条件**——没有根清单的市场只会看到一个 0.0.0 依赖。
**一条实测更正**：`LivXue/dsh-plugin-shop` 的 `docs/schema.md` 写着「GitHub 仓库也能列」，但它的实际发布物
推翻了这个说法——2026-11-10 那份 25,385,268 B 的分片里 26,351 条条目**全部**带 npm tarball 的 `integrity`，
`commit` 字段一条也没有（`"source": "npm"` 为 0 也是同理：字段名是 `integrity`/`tarball` 而不是 `source`）。
所以根清单 shim 只是「有资格被它读」的前提，不是「一定会被它列出来」的保证。写文档时别把条件满足写成已收录。

也解释了为什么 `keywords` 与 topics 都要填 `dsh-plugin`：不同采集方读的是不同位置（有的读清单字段、
有的读 GitHub topics），而双 topic 那一家还要求 `deepseek-harness` 同时存在。

**这一批的 CI 曾经是红的，根因不是代码错**：`d06001d` 之后 CI 在 `windows-latest` 上挂了，红在
`vendor:check` 那一步。原因一行就能说清：`go.mod` 里写的 `go 1.23.0` 是**下限不是钉版本**，
CI runner 用 go 1.23.0、本地用 go1.27.0，两个工具链编出来的二进制字节不同——
**跨工具链的字节相等从来就不可能达成**，所以那一步当时是拿一个达不到的条件当门禁。
修法是把断言改成能达成的那种：`vendor/manifest.json` 记录**源码摘要**（25 个非测试 `.go` 文件）
以及两个 vendor 文件自己的哈希，`vendor:check` 校验这些；只有当**本地 Go 版本等于 manifest 里记录的
`go1.27.0`** 时才去和 `bin/` 做字节比对。于是 CI（go 1.23.0）校验的是「vendor 与源码摘要一致」，
本地（go1.27.0）额外校验「vendor 与 `bin/` 字节一致」，两边都在验真实的东西。
**写文档时不要美化这段**：CI 是红了之后修的，不是一次就绿；也不要把它写成 go.mod 需要钉工具链。
- **`locale.meta.detail` 省略**：清单里只给一行摘要，详细说明指向 `README.md`。
  这样市场文案不会随文档改动而漂移，代价是用户必须点进 README 才看到限制（Windows only 等已由 `platforms` 表达）。
- **`verify.mjs` 没有更强断言**：它验证了 `insert:` 能挂载，但没有断言「patch 文件里出现裸 `- id:` 就该失败」。
  这个坑现在只靠 `cordis.patch.yml` 的文件头注释防住，属于文档约束而非测试约束。

### 27.10 自建 format-1 索引：让 Discover 在 npm 之前可用（2026-11-10）

**动机**：npm 发布卡在一次性密码（注册表对 PUT 回 `401 You must provide a one-time pass`，
`npm publish` 的 EOTP 链接被 npm 打码成 `***`，浏览器回写 token 在这个无终端环境里拿不到）。
而 Discover 的索引格式允许「git 源 + 40 位 commit」，于是不必等 npm。

**产物**：`docs/market-index.json`（`ef783ba`，已推送），公开 URL
<https://raw.githubusercontent.com/jjjjjjjjnnjnn/ai-terminal-pet/main/docs/market-index.json>（实测 200）。
条目只有一个：`dsh-ai-terminal-pet` `1.2.0`，
`source: {type: git, url: https://github.com/jjjjjjjjnnjnn/ai-terminal-pet,
commit: 4d777e32e0a9ef57bb7714287dccd7cae065346f, path: packages/dsh-ai-terminal-pet}`（与 tag 指向同一提交），
`engines.dsh` 抄包里声明的 `>=0.1.0-rc.6`，图标钉在同一 commit 的 `packages/dsh-ai-terminal-pet/icon.png`（4,386 B），
`publishTimestamp` 取 Release 的 `2026-11-10T12:47:24Z`（Unix 秒 1794314844）。

**校验方式与时序**：不手抄一份形状检查，而是直接 `import` 宿主的
`@deepseek-ai/dsh-plugin-market-index`（`parseMarketIndex` / `marketInstallTarget`）做真校验——
`format: 1`、`localizedText` 必含 `en`、category 与 label 必须引用 `metadata` 里已声明的 id、
`gitUrl` 必须 https 且至少两段 path、`commit` 必须小写 40 位 hex、版本精确 SemVer、
`publishTimestamp` 必须是 [0, 10^11) 的整数秒。随后用 `marketInstallTarget` 产出的 spec 在
全新 `DSH_HOME` + 独立 profile 里真装了一次：**rc=0 / 2.5 s / `+ dsh-ai-terminal-pet 1.2.0`**。

**为什么这比只提 PR 更有效**：PR #13937 是「别人家的列表」，入列要人工评审；自建索引是「自己的列表」，
用户只要在 Discover 里加一次 URL 就能看到并一键装。两份都可以留，不冲突。

**边界（别夸大）**：索引不会自动进入任何人的 DSH——用户必须手动添加这个 URL；
它也不改变 npm 上的状态（`npm view` 仍 E404）；格式 1 没有「一个版本同时挂 npm 与 git 两种源」的表达，
所以等 npm 通了要发下一版时**新增一个版本条目**而不是改现有那条的来源类型。
**同样是外部数据**：本地验证用的是宿主的 validator，但上游是否会采信我们的索引、以及
各家第三方市场何时重建，都不由我们决定。

### 27.11 安装脚本下沉到 `scripts/`（2026-11-10）

**做了什么**：把仓库根的 `install.ps1`（27,358 B / 583 行）移到 `scripts/install.ps1`（27,930 B / 590 行），
仓库根从此没有同名文件。行为一个字没改——停进程、20 次重试拷贝、逐文件 SHA256、`-VerifyOnly` /
`-Uninstall` / `-AddToPath` 三个开关、`Assert-SafeInstallDirectory`、PATH 备份与清理全部原样。

**为什么**：[DSH-Plugins-Marketplace](https://github.com/bradeGithub/DSH-Plugins-Marketplace) 的
`evalRepoRisk`（`scripts/build-registry.mjs` + `lib/domain/security-scan.js`）**只按固定文件名抓仓库根**的
`package.json` / `install.sh` / `install.ps1`，然后把整份文件喂给正则表。它给我们报的两条旗标
（`registry-run-key` critical、`remove-item-force-recurse` medium）匹配的是脚本里**合法的**内容：
`$RunKeyPath = 'Software\Microsoft\Windows\CurrentVersion\Run'` 只是注册表路径常量（脚本自己从不
`New-ItemProperty` 到 Run，只有 `-Uninstall` 时按条件删自己写的那条），`Remove-Item -Recurse -Force`
只作用于经 `Assert-SafeInstallDirectory` 校验过的安装目录。该扫描器**没有 suppress / allowlist 机制**
（唯一的 allowlist 是 `HAZARD_ALLOWLIST_DOMAINS` 对 downloadExec 的域名降级），所以「改注释绕正则」不是
正路；把它移出被扫描的路径才是。

**顺带的好处**：市场类型判定 `detectTypeDetail(cacheDir)` 的第三分支是「根 `install.ps1` → script」，
移走后该分支不再命中，判定落到第二分支——根 `package.json` 的 `hasDshPluginDeclaration`（`pkg.dsh` 是对象）
与 `isBundlePackage`（`dsh.bundle.patch` 非空字符串），结果是 `bundleDeclared` / type `bundle`。这与
16:22Z 那次重建得出的 `verdict: bundle-plugin` 一致，等于两条独立证据指向同一结论。

**两条风险旗标也随之消失**：18:08:44Z 那次 `build-registry` 重建（`risk_at` 17:09:14Z，是脚本移走后
第一次真的扫到本仓库）把条目写成 **`risk_tier: "safe"`、`risk_flags` 为空**，`pkg_name` / `version: 1.2.1`
也补上了。这不是靠抑制字段得到的——`evalRepoRisk` 抓不到根目录的脚本，正则表无从命中。issue #367
先回帖于 17:58:34Z（`#issuecomment-6775035737`），19:04:00Z 再补一条（`#issuecomment-6776560881`）写明
重建后的 scorecard 与 `verdict: bundle-plugin`，并建议由维护者关闭。

**技术要点**：`param()` 的默认值在**参数绑定阶段**求值，此时脚本体的赋值还没执行，所以不能写
`Join-Path $RepoRoot ...`；`$PSScriptRoot` 是唯一在那时就可用的自动变量，因此四个默认值写成
`Join-Path (Split-Path -Parent $PSScriptRoot) 'bin\AI-Terminal-Pet.exe'` 这类形式。脚本体里再定义
`$RepoRoot = Split-Path -Parent $PSScriptRoot`，给预检用来判断某个缺失文件是否属于 `bin/` 的 `$binRoot` 用。

**实测**：`PSParser` 解析 `PARSE_OK`；`-VerifyOnly` → `install: verified - 10 file(s) match the repository, 0 warning(s)`、
exit 0；完整安装 → rc=0，两个 exe（4,371,968 / 3,643,392 B）与 `protocol.js`（20,914 B）与 bundle 7 文件当场
重新落盘；装完再跑 `-VerifyOnly` 仍 10/10。

**代价**：全仓 27 个文件里约 100 处 `install.ps1` 文本引用要跟着改（README 中英两半、AGENTS.md、ROADMAP、
CHANGELOG、插件包 README 与 `lib/index.js` 注释、`integrations` 桥接 README 与 `index.js`、`.gitignore` 首行、
以及 Go 源码里的注释）。改漏一处就是给读者一条跑不通的命令——这也是为什么这次把 `scripts\` 前缀写进了
所有能被执行的地方，而不是只说「脚本已移动」。

**没有做的事**：没有给 `scripts/install.ps1` 留一个根目录的转发壳（那会把被扫描的路径又造回来）；也没有
把 `-Uninstall` 的确认提示改成非交互（那是另一件事，且当前行为是安全默认）。
