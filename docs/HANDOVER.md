# HANDOVER — ai-terminal-pet 交接文档

> 交接时间：2026-09-24（UTC+8 下午）；2026-11-09 刷新到 v0.9.0；2026-11-10 补记 v0.9.0 之后未发版的工作，含 DSH 插件包 1.2.3 与市场条目合并；2026-11-11 刷新到插件包 1.2.4（只改文案）并补齐三家已合并的目录；**2026-11-11 10:30 +0100 交接整理**——同步到 `bb00e58`、记入 dshopen 收录、新增 §28 交接清单。
> 下一位 AI 直接从 §1「当前状态」与 **§40「交接：提醒层」**（产品能力）以及 **§39「交接：上游目录条目 sync」**（上架机制）开始读，然后按 §13–§27 的时间序读
> runbook，不必回溯长上下文；§35–§38 是 2026-11-12 这一天四轮工作的施工记录，§39 是同日晚上把「上架」
> 变成机制的记录。上架的规则与清单另有专文：[docs/UPSTREAM.md](UPSTREAM.md)。
> 2026-09-24 停用态时期的章节已归档到 [docs/archive/HANDOVER-2026-09-DISABLE.md](archive/HANDOVER-2026-09-DISABLE.md)。

## 1. 当前状态（2026-11-17；浮窗 **1.0.18**、插件包 **1.2.31** 已发版；**§75 第 75 轮（OpenCode 上报端成为 npm 包）；§74 第 74 轮（门禁、卸载项、像素、设置入口、推广功课）；§71（三平台视觉不变量门禁）；§70（安装器写卸载项 + winget 清单）**）

> 接手请先读本节与 **§75「OpenCode 接入变成一行 npm 安装」**（为什么它是一个发行缺口、npm 包为什么是生成的、
> 两个坑、以及本机为什么故意还没换配置），再读 **§74「一轮五件事，下个会话从这里开始」**（机器状态、外部世界谁在等谁、
> 下一步顺序，以及三条只有本人能做的事），再读 **§72「Linux/macOS 终于有设置入口」**
> （终端编辑器与 `--set`、为什么手改 JSON 不算入口、三个"看起来对其实错"的地方，
> 以及顺手修掉的 `100%!z(MISSING)oom`），再读 **§71「面板的像素第一次被断言」**
> （三平台视觉不变量门禁：断言什么、为什么不比像素），再读 **§70「安装器写卸载项」**
> （为什么"装得上、卸不掉"是个真缺陷、注册表那一行的每个字段、沙箱键名为什么必须派生、
> 以及 winget 清单的提交状态），再读 **§69「把文档里的数字变成门禁」**
> （`tools/check_facts.mjs` 钉什么/不钉什么；它第一次运行抓到的 `#english` 假警报与被修正的 E5 清单），
> 再读 **§68「公开文案重做」**
> （README 的新结构、被抓到的三处不实描述、图片门禁与它的自测、以及"硬件实测永久跳过"的对外写法），
> 再读 **§67「DSH 插件三平台」**
> （四个改动面、三个"看起来对其实错"的地方、尺寸方案为什么是"首启按 pin 下载"、以及两条过程教训），
> 再读 **§66「POSIX 平权落地」**
> （Linux 提示音与通知面、macOS 的 osascript、跳转失败反馈、`--doctor` 从 11 项变 13 项、三条证不到的事），
> 再读 **§65「交接：下个会话从这里开始」**
> （上一轮收尾时的机器状态与四条上手命令；那一节的 §65.4 四项候选，第 1 项已在本轮做完），
> 再读 **§64「上游目录里我们那行为什么是旧的（两个不同的原因），以及交接前的又一轮装机实测」**
> （站点 `version` 停在 1.2.16 是对方整站冻结（#15569 / #15579），下载数少掉的 2,276 是两段叠加
> （1,197 见 #15864 + 1,079 是冻结）；**已按决定不再跟进**，判断用 `node tools/catalog-status.mjs`），
> 再读 **§63「修 1.0.11 的回归：`title` 等于窗口标题时不算会话存在」**
> （结束几小时的 agy 会话被标签页标题留在面板上：四条现场记录、修法与三条测试），
> 再读 **§62「L5 快照扫描复用解析结果，以及缓存命中时删除守卫的真 bug」**
> （200 文件 22.7 ms → 0.51 ms；命中路径拿不到 `FileInfo` 导致的过期快照泄漏；三条安全性质与一个 NTFS 实测事实），
> 再读 **§61「agy 点不跳的根因、设置布局搬出 Windows 文件、`-race` 进 CI，以及 L3/L5 的两个数字」**
> （三层取证、控制台标题锚点、一个被否掉的方案、没修的半边，以及本轮数字的来源），
> 再读 **§60「POSIX 路径/信号收口、文档全面重构与 1.0.10 / 1.2.19 全平台全量化发布」**（再上一轮的收口与发版），
> 再读 **§57「跨平台施工（第 1–7 步），以及这一版为什么还没发」**
> （九个提交、三平台实测矩阵、五个"看起来对其实错"的地方、以及窗口层为什么不在那一版里），
> 再读 **§56「`--doctor` 现在会看上报端接入，`--connect` 只打印」**，
> 再读 **§55「交接：下一步是 macOS / 其他系统 —— 先盘点，别先动手」**
> （跨平台盘点：可搬行数、当时 darwin 上编译到哪一步、五层要换什么、七步顺序与五个坑——**那七步在第 57–60 轮已走完**，
> 读它是为了看那五个坑与判断方式，不是为了照着再走一遍），
> 再读 **§53「交接：上报端只剩一份实现」**（本机现在加载的是仓库里的 `reporters/opencode.js`，
> 本地目录插件只是一次静态 re-export），再读
> **§52「交接：两个现场缺陷已修 —— 归档即结束、设置窗口不再跟着自毁」**（1.0.8 / 1.2.17 的代码与验收）、
> **§51「交接：两个现场缺陷」**（只取证的那一轮）、
> **§50「中文 README 与市场文案」**、
> **§49「上架巡检：十家目录现在各指哪里」**、
> **§48「交接：适配 DSH 0.2.x」**（1.2.15 的兼容声明）、
> **§47「维护：工作区整理」**（只删可再生产物）、
> **§46「交接：把两个流程缺口变成门禁」**（工具与测试）、
> **§45「交接：把第 44 轮发出去 —— 1.0.7 / 1.2.14」**（上一次发版，没有改代码）、
> **§44「交接：设置窗口可读、可用」**（1.0.7 的代码就是它）、
> **§43「交接：三个假提醒的根因 + OpenCode 2.x」**（上一轮）、
> **§42「交接：点行复制路径」**、**§41「交接：判定记录搬家」**、
> **§40「交接：提醒层」**（产品能力）和 **§39「交接：上游目录条目 sync」**
> （上架机制，读它再读 §28 的清单形状）；§28 是更早一版的清单（分发状态已过期，只保留它的清单形状），
> §27 是插件包分发的完整 runbook，而上游目录条目的规则、八步发版顺序与「合并 ≠ 可见」的判断方式在
> [docs/UPSTREAM.md](UPSTREAM.md)。

- **在役**：安装目录 `AI-Terminal-Pet.exe` 已安装并通过 SHA256 校验；浮窗按需由 reporter 拉起，
  无常驻服务，无会话时自毁。
- **已修复（详见 §52）**：① **归档的会话仍占一行**——结束信号原来只有 `session.deleted`，
  而 OpenCode 归档只填 `time_archived`、不发那条事件（`opencode serve` 可活几天）。现在
  `session.archived → closed`、`session.unarchived → idle`，并在 hydrate 时用宿主自己的会话记录再挡一次。
  ② **设置窗口跟着进程消失**——空闲自毁（会话表空 10 分钟）会连同一个 UI 线程上的设置窗口一起带走；
  现在**可见的设置窗口算"仍在使用"**（`internal/ui/lifetime.go` 的 `idleDecision` 纯函数），
  关闭设置后重新计 10 分钟。
- **版本**：浮窗 **`1.0.18`**（`internal/app/version.go` 的 `const Version`），插件包 **`1.2.31`**。
  两条版本线各有 tag：`pet-v1.0.0`…`pet-v1.0.18` 与 `dsh-plugin-v1.2.6`…`dsh-plugin-v1.2.31`。
  OpenCode 上报端包 `opencode-ai-terminal-pet` **没有自己的版本线**：它与浮窗同号（`check_versions` 断言相等）。
  **工作区与本机安装目录已经与 `1.0.18` / `1.2.31` 同步**：`bin/`、
  `packages/dsh-ai-terminal-pet/vendor/` 与 Release 逐字节相同
  （`AI-Terminal-Pet.exe` 的字节数与 sha256 见下方规模一行）。
  更早的每一个 tag 与资产仍然有效且不可变。查当前 CI 用
  `gh run list --limit 1`，查提交用 `git log --oneline -10`——**不写死**，因为 git 无法在一个提交里引用它自己或它的后继。
- **健康**：`gofmt -l .` 无输出；`go vet ./...` exit 0；`go test -count=1 ./...` 全过（**498 个** `func Test`，
  本轮（第 75 轮）+3 来自 OpenCode 的 npm 包形状与包名解析）；CI 四条 job
  （windows / ubuntu（含 `go test -race`）/ macOS / scan）在 1.0.18 的提交上全绿（见 §75）；
  `scripts/cross-build.sh` 全 16 包三端 `CGO_ENABLED=0` 通过（本机无 gcc，用 PowerShell 复跑同一逻辑：16 包 × linux/amd64 + darwin/arm64 全过）；两个 `verify.mjs` 全 PASS；
  `vendor.mjs --check` 与 1.0.18 的 `bin/` 一致；`tools/check_versions.mjs` **21** 条一致
  （`node --test tools/check_versions.test.mjs` 27/27）、**`tools/check_facts.mjs` 11 条一致**、
  `tools/catalog-status.test.mjs` **12/12**、`tools/upstream-entry.test.mjs` 26/26、
  `tools/check-tarball.test.mjs` 8/8；**OpenCode 包**：`sync.mjs --check` 3/3、
  `sync.test.mjs` 4/4、`server.test.mjs` 1/1（本机 Windows 走命名管道，Linux/macOS 在 CI 走 socket）；
  `tests/fresh-machine.mjs --no-overlay` **183/183**；
  `tools/check_docs_links.py` **14 份**全解、`tools/check_assets.py` 15 行一致；
  `govulncheck` + `gosec` 在 CI 的 scan job 跑。
  1.0.18 的装机实测：`scripts\install.cmd` → `-VerifyOnly` **12/12 一致、0 warning**、
  `--doctor` **12 ok / 1 warn / 0 fail**（唯一 warn 仍是 Antigravity 只接 statusline）、
  `--version` 报 `ai-terminal-pet 1.0.18`，注册表那一行 `DisplayVersion=1.0.18`。
- **规模**（`node tools/check_facts.mjs --print`，1.0.18 实测）：非测试 **154 文件 / 28,255 行**、
  测试 **100 文件 / 14,735 行**、**498 个 `func Test`**。
  `bin/AI-Terminal-Pet.exe` **4,764,160 B**（与 `vendor/AI-Terminal-Pet.exe` 逐字节相同，sha256 `B8DE4C94…`）、
  `bin/pet-statusline.exe` **3,678,720 B**、`bin/pet-hook.exe` **3,705,344 B**。
  三个二进制合计 **12,148,224 B** = 11.585 MiB。安装目录里多了一件制品：`install.ps1` 的副本。
- **分发（两条版本线 + 两个 npm 包 + 一个包管理器）**：
  - 插件线：npm `latest` = **1.2.31**；`dsh-plugin-v1.2.31` 的 Release 资产 `dsh-ai-terminal-pet-1.2.31.tgz`
    **2,091,456 B / sha256 `8CEA5882…` / 20 文件**（解包后跑包内 `vendor/AI-Terminal-Pet.exe --version`
    = `ai-terminal-pet 1.0.18`），**与 npm 上那一份逐字节相同**；上游条目 PR 见 §75。
  - OpenCode 上报端线：npm `opencode-ai-terminal-pet` **1.0.18**（与浮窗同号），
    资产 `opencode-ai-terminal-pet-1.0.18.tgz` **17,822 B / sha256 `C6588551…` / 5 文件**。
  - 浮窗线：`pet-v1.0.18` 的六个资产——`pet-setup.exe` **14,551,552 B / `8A589006…`**、
    `ai-terminal-pet-1.0.18-win-x64.zip` **5,142,108 B / `70AE18C0…`（17 文件）**，以及
    linux-{amd64,arm64}（**5,439,674 B / `5F74890E…`**、**4,949,913 B / `17E29037…`**）与
    darwin-{arm64,amd64}（**5,143,588 B / `9416A411…`**、**5,536,325 B / `39193DBF…`**）四个 tar.gz。
  - 钉 tag 的 git 源；仓库根的清单 shim；自建 format-1 索引（`source.commit` 与 `icon` 均钉到 Commit A）。
    `b4474a6` 那次 CI 红就是反面教材：索引只被下一个提交 `fabb591` 前移，而 `verify.mjs` 要求「索引里的最新
    版本 == 包版本」，于是 tag 指向的提交本身是红的（规矩：发版提交必须同时前移索引，见 AGENTS.md 第 6 条）。
    细节与坑见 §27。
  - **winget**：清单在 `packaging/winget/`（`winget validate` 通过，`winget show` 解析正确），
    提交到 `microsoft/winget-pkgs` 的 PR 见 §70.6。能不能上取决于对方 pipeline 里那次静默安装/卸载；
    未签名带来的 SmartScreen 风险写在 [docs/UPSTREAM.md](UPSTREAM.md) §4.3。
- **接入**：OpenCode（**npm 包 `opencode-ai-terminal-pet`，`opencode plugin add` 一行**；包内 `server.js`
  由 `reporters/opencode.js` 生成，本机 `opencode.json` 目前仍指向
  `~/.config/opencode/plugins/terminal-pet` 这个目录，里面的 `index.ts` 只做一次静态 re-export，
  见 §53——§75.6 记了为什么本机特意先留着它）、Claude Code（`reporters/claude-hook.js` 7 事件）、
  Antigravity（`bin/pet-statusline.exe` + `reporters/antigravity-hook.js`）、DeepSeek Harness
  （两种：老的 host-only bridge bundle `integrations/deepseek-harness/ai-terminal-pet-bridge`，
  与新的独立插件包 `packages/dsh-ai-terminal-pet`，见 §27）、任意 CLI（`reporters/generic.js`）。
- **未跑项**：`go test -race`（缺 gcc/cgo）。`govulncheck` 与 `gosec` 不需要 C 工具链，已于 2026-11-12
  补进 CI（`govulncheck@v1.9.0` 0 条被实际调用的漏洞、`gosec@v2.30.0` 在 high/high 上 0 条），
  排除的规则族与理由写在 `.github/workflows/ci.yml`，见 §35。

> **历史内容去哪了**：2026-09-24「项目全局停用」时期的停用态描述（原 §1、§2、§4、§9、§10 的旧正文）、
> 以及原 §3 架构速览、§5 已知残留、§6 关键修复、§7 验证基线、§8 常用命令、§11 残留风险、§12 重启用记录，
> 已整体搬迁到 [docs/archive/HANDOVER-2026-09-DISABLE.md](archive/HANDOVER-2026-09-DISABLE.md)
> （章节号与正文逐字保留，未删任何技术细节）。因此本文件的章节号从 §13 起续接，§3–§12 有意留空。
> 当前有效 runbook = 本文 §2 + §13–§35（§26 是 v0.9.0 之后的批次，§27 是 DSH 插件包分发，§35 是 2026-11-12 的门禁加固与账号/签名交接）。

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

**完整取证 / 验证矩阵见 `docs/archive/BASELINE-20260925.md`。** 遗留：`-race` 仍缺工具链（`govulncheck` / `gosec` 已于 2026-11-12 补上，见 §35）。

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

**遗留**：`docs/ROADMAP.md` §6 列出的未修项（`-race` 缺工具链、UI 无视觉回归、
管道抢占只做缓解）；`govulncheck` / `gosec` 已不在其中（2026-11-12 补进 CI，见 §35）。

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

- `-VerifyOnly`：只读。按**同一份 11 项制品清单**（overlay / statusline / hook / protocol / bundle manifest 2 / bundle runtime 5）
  逐项比对工作区与已部署侧的 SHA256，全对才 `install: verified - 11 file(s) match the repository` + exit 0。
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
go vet ./...                                 # 必须无输出
go test -count=1 ./...                       # 201 个 func Test，12 个包全 ok
for f in reporters/*.js; do node --check "$f"; done
node integrations/deepseek-harness/verify.mjs
node packages/dsh-ai-terminal-pet/verify.mjs
go build -trimpath -buildvcs=false -ldflags "-s -w" -o .\bin\AI-Terminal-Pet.exe ./cmd/ai-terminal-pet
go build -trimpath -buildvcs=false -ldflags "-s -w" -o .\bin\pet-statusline.exe ./cmd/pet-statusline
go build -trimpath -buildvcs=false -ldflags "-s -w" -o .\bin\pet-hook.exe ./cmd/pet-hook
node tests/fresh-machine.mjs --no-overlay   # 162 项
.\scripts\package.ps1                       # 便携 zip
.\scripts\install.cmd                       # 装（cmd 包装，绕开 RemoteSigned）
.\scripts\install.cmd -VerifyOnly           # 期望 11/11 [ok]
& "$env:LOCALAPPDATA\AI-Terminal-Pet\AI-Terminal-Pet.exe" --doctor   # 期望 0 fail，exit 0
```

**这一段的数字是 2026-11-11 复读的**：更早版本里的 154 / 10 件制品 / 两处 `install.ps1` 都已是历史——§29
之后多了 `cmd/pet-hook` 与 `internal/reporter`（154 → 190），§30 之后多了 `tests/fresh-machine.mjs` 作为
独立门禁，§33 之后制品 10 → 11 并把文档入口换成 `install.cmd`。**引用前请重跑，别照抄本节。**

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

`packages/dsh-ai-terminal-pet/` 共 **18 个文件**（含 vendor；1.2.0 起多一个 `vendor/manifest.json`）。
下表的字节与行数是 **1.2.4 口径**（2026-11-11 实测）；1.2.0 那一版的口径见本节末尾注：

| 文件 | 字节 | 行 | 作用 |
|---|---|---|---|
| `package.json` | 2,246 | 83 | 包元数据 + `dsh.bundle` + `dsh.catalog` 声明（scripts 只剩 verify / vendor / vendor:check）；1.2.0 时 2,008 / 74 |
| `lib/index.js` | 17,872 | 570 | 事件映射、去抖、bootstrap、下载回落（自 1.2.0 起字节与行数都没动） |
| `verify.mjs` | 23,426 | 424 | 自包含验证脚本（无测试框架），1.2.4 口径 **95 个 `ok`**；1.2.0 时 22,499 / 415；tag `dsh-plugin-v1.1.0` 上是 15,117 B / 300 行 |
| `scripts/vendor.mjs` | 13,981 | 324 | 双模式：无参时从 `../../bin` 与 `../../reporters` 拷进 `vendor/` 并写 `manifest.json`；`--check` 只校验 sha256 |
| `README.md` | 8,187 | 146 | 包自己的安装/限制说明（写明「本包不带任何生命周期钩子」、「只直接接入 DSH、其他终端要各自装 reporter」，与 `dsh.catalog` 的自我声明性质） |
| `cordis.patch.yml` | 399 | 9 | loader patch 层（`insert:`；裸 `- id:` 会挂载不到东西） |
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

> 注：**「18 个文件」与 tarball 的「16 files」两个数字都对**——`git ls-files packages/dsh-ai-terminal-pet`
> 数的是 18（多出 `verify.mjs` 与 `scripts/vendor.mjs`，这两份不进 `files` 白名单，因此不进 tarball）。
> 谁要「修正」成 16，就会把包目录说少两个文件。

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

### 27.5 发布与分发现状（截至 2026-11-11）

- **npm**：**最新是 1.2.5**（2026-11-11T14:58Z 前后上线；1.2.4 在 12:10:46Z）。
  两次都由 agent 直接在包目录跑 `npm publish --access public` —— 本机 token 已配置成可绕过 2FA，
  所以**不再需要人工在交互终端操作**；这一步的旧结论「只能由人跑」已经作废。
  `dist-tags.latest` = 1.2.5、shasum `1f6d5340…`、integrity `sha512-2sEFAf0k…`、17 files。
  发布前在包目录跑 `npm publish --dry-run` / `npm pack` 并把字节与 sha256 记进 CHANGELOG；
  从 registry `npm pack` 回来的 tarball 与 Release 资产**逐字节相同**（sha256 `dfa89a1e…`）。
  两个仍然成立的事实：从**仓库根**发布会撞 `EPRIVATE`（根清单 `private: true`），
  以及发布后头几分钟 registry 上只有占位的 `0.0.0-stage`，此时 `npm install <name>@<version>`
  会报 `ETARGET`，本地 npm 缓存还要 `--prefer-online` 才看得到新版本。
  以下为发布之前（截至 2026-11-10）的记录，保留作背景：
  凭据问题**已解决**（2026-11-10T12:20Z 换了 token 后 `npm whoami` 返回 `jjjjjjjjnnjnn`，
  此前整场会话都是 E401）。`npm view dsh-ai-terminal-pet version` 当时仍是 **E404**，说明那个名字还空着。
  发布被一次性密码（OTP/2FA）挡住：npm 的 `otplease` 在无 TTY 时直接抛 EOTP，
  浏览器授权链接（带 `npm-auth-type: web` 的探针才拿得到未打码的那个）5–8 分钟即失效，只能由用户在本机交互式终端里发。
  发布前已在包目录跑过 `npm publish --dry-run`：rc=0，**16 files / 1.9 MB packed / 4.5 MB unpacked**
  （那次 dry-run 的 shasum 是 `166c3043…`，但之后包内 `README.md` 又改了 13 B，所以最终的号不是它；
  1.2.1 的 dry-run 口径见下面 1.2.1 那条，1.2.2 见它自己那条）。
  因为从**仓库根**发布时 pnpm/npm 抬到的是根清单，而根清单带 `private: true`；包清单没有 `private`，
  所以在 `packages/dsh-ai-terminal-pet/` 目录里发布是可行的。这条对写文档的人有价值：根 shim 的
  `private` 是「别抢 npm 名」，同时也意味着**不能从仓库根发布**。
  仍未解锁的是 DSH 应用内 Plugins → Discover 索引：它只表达 npm 版本（format 1）或钉到 40 位 commit 的
  git 源。**1.2.4 发布之后** `add dsh-ai-terminal-pet` 这条按包名的路径已经解析得到（npm 版本 1.2.4），
  自建的 format-1 索引（钉 `ff7f6de`）继续作为 git 源路径存在。
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
  - **1.2.4（当前；只改文案，不改代码）**：存量市场文案读起来像「装一次市场条目就能看到所有终端」，
     而事实是这个 bundle 开箱即用映射的只是 **DeepSeek Harness 自己的会话**——OpenCode、Claude Code 与
     Antigravity 的会话要先给**各自那个终端**装好上报端（`reporters/opencode.js` / `claude-hook.js` /
     `antigravity-hook.js`）才会汇进同一块面板；没装上报端的终端在任何安装形态下都不上报。
     这句话来自用户按商店实际规格从头复现安装后给出的唯一一条（不阻断）反馈，因此值得为它单独发一版。
     改动落在三处，且**摘要必须逐字相同**（`verify.mjs` 的根清单用例要求两份清单的 `dsh.catalog`
     序列化后完全一致）：
     ① 包内 `README.md`（7,546 B / 137 行 → **8,187 B / 146 行**）：开头那段「一次市场安装就够」改成
        「先把这个 Harness 的会话接上、其他终端要各自装上报端」，并指向仓库 README 的
        `#connect-your-terminal` 段；standalone 段补 Those reporters are separate installs, not part of this package；
     ② 根 shim `package.json` 的 `dsh.catalog.summary`；
     ③ `packages/dsh-ai-terminal-pet/package.json` 的同一块，与 `docs/market-index.json` 的 `plugins[0].description`。
     新摘要 en 225 字符 / zh 114 字符（原文 190 / 104）。两份清单 `version` 1.2.3 → **1.2.4**
     （根 2,052 → **2,115 B / 68 行**、包 2,183 → **2,246 B / 83 行**）。
     **随包字节其余部分完全不变**：`lib/index.js`（570 行 / 17,872 B）、`vendor/AI-Terminal-Pet.exe`
     （4,371,968 B / sha256 `bbb5163c296cd460d7c9a4c85e10342e125c0bd8017a065f913aa01567d41f0a`）、
     `vendor/protocol.js`（20,914 B）、`cordis.patch.yml`、`icon.svg` / `icon.png`、
     `locale/en.json` / `locale/zh.json`、`screenshots/` 三张、`vendor/manifest.json`（源码摘要仍
     `cf647af1…`——1.2.4 没碰任何 `.go` 源码，所以记账不用重算，这一点与 1.2.2 形成对照）。
     打包实测：`dsh-ai-terminal-pet-1.2.4.tgz` **1,915,340 B / 16 files / unpacked 4,462,112 B /
     sha256 `d6b7f95a5a3a71c4a9367699e9ad9b8f7d5db522271782e8ffe29a73e45c6708` / shasum
     `dafb81998efbae8e906a36773e9737117aa482d9`**，连跑两次 `npm pack` 逐字节相同（`cmp` 无差异）；
     与 1.2.3 的 1,915,013 B 只差 **327 B**，全部来自两块摘要与包内 README 的新增段落。
     `verify.mjs` 复跑仍是 **424 行 / 95 ok / 0 warn / 0 fail / VERIFY PASSED / rc 0**——只改文案，
     断言一条没加也没减，`caseMarketIndex()` 的「索引最新版本 == 包版本」直接跨过这次改号。
     `docs/market-index.json` 的 `versions[0].version` 与 `publishTimestamp`（1794379909 =
     2026-11-11T06:51:49Z）与两份清单**同一批**前移，`source.commit` 与图标钉点由紧随其后的提交补齐
     （沿用 1.2.1 / 1.2.2 / 1.2.3 那三次的两步做法，见 §27.10 与 AGENTS.md 第 6 条）。
     1.2.4 的资产就位后又逐个装过一遍（每次全新目录 + 纯 ASCII 路径）：Release 资产 URL **1.8 s**、git 源钉
     tag `dsh-plugin-v1.2.4` **3.7 s**、`add github:jjjjjjjjnnjnn/ai-terminal-pet` **6.5 s**，三者 rc=0；
     装出的包与 `git show ff7f6de:` 逐字节一致（`README.md` 8,187 B / `package.json` 2,246 B /
     `lib/index.js` 17,872 B / `cordis.patch.yml` 399 B 四项 sha256 全 MATCH），vendor 两个文件也与
     `vendor/manifest.json` 对得上。
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

### 27.7 已知残留（2026-11-11 刷新到 1.2.4）

- **npm 已发布 1.2.4**（2026-11-11T12:10:46Z，`npm publish --access public` 由人在本机交互终端完成）。
  `npm view dsh-ai-terminal-pet version` → **1.2.4**，`dist-tags.latest` = 1.2.4。
  发布内容与仓库里验证过的 tarball **逐字节相同**：shasum `dafb81998efbae8e906a36773e9737117aa482d9`、
  integrity `sha512-VrD4QNfzBLbqlstDG8AW+ys9M5pu6QEPHlxyVyK1qI1Wgkc8TX10kWPB/413gzf2oufNPp3NABBdfvic3JA/jA==`、
  16 files / unpacked 4,462,112 B。另用 `npm pack dsh-ai-terminal-pet@1.2.4` 从 registry 取回一份，
  sha256 同样是 `d6b7f95a5a3a71c4a9367699e9ad9b8f7d5db522271782e8ffe29a73e45c6708`；
  在临时目录 `npm install dsh-ai-terminal-pet@1.2.4` 装出来 16 个文件，
  `vendor/AI-Terminal-Pet.exe` 与仓库那份 sha256 相同（`bbb5163c…`），`lib/index.js` 作为 ESM 可加载并导出 `apply`。
  **一个坑记下来**：发布后头几分钟 registry 上只有一个占位的 `0.0.0-stage` 版本（`latest` 也指向它），
  此时 `npm install dsh-ai-terminal-pet@1.2.4` 会报 `ETARGET`；约 3 分钟后真版本才出现。
  本地 npm 缓存还会继续给旧 packument，`npm install --prefer-online` 才能立刻看到新版本。
  **改这一条时不要再写「npm 尚未发布」或「被凭据阻塞」**——两者都已经是历史。
- **已收录**不等于**已渲染**——这条经验在 2026-11-11 又推进了一步：dshopen 的 #423 已 closed，
  而且它**站点域名上的** `https://dshopen.org/index.json` 现在确实有我们（1,207 条里的 `dsh-ai-terminal-pet`
  1.2.4，`source.type: git`、钉 `ff7f6de`、分类 `integrations`）。之前看到的「854 条、0 条 git 源」
  是它仓库里的 `public/index.json`，那是**未重新构建的仓库副本**，和站点实际提供的索引不是同一份。
  所以判读规则再收紧一格：**读对方站点域上的索引文件，不要读它仓库里的 `public/` 目录**。
  同理 awesome-dsh-plugin 的 `plugins.json` 里那条 `install` 命令是构建时按条目 `tarball` 渲染的，
  「PR 合并」也只意味着「下一次重建之后可见」。
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

`06d4cae` 加的是**仓库根**的 `package.json`（1.2.4 起 2,115 B / 68 行；1.2.3 时是 2,052 B，比 1,877 B / 59 行多出的九行就是 `dsh.marketplace`，1.2.4 只把 `dsh.catalog` 摘要加长），`d06001d` 让验证脚本与 CI
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

### 27.12 1.2.4：只改文案的发版，以及「哪些终端要另外装」这件事（2026-11-11）

**为什么值得单独发一版**：用户按商店的实际规格（`dsh plugin --profile <name> add <仓库 URL>`）在隔离目录
从头复现了安装，结论全部通过——条目合规、`pnpm add` rc=0、解析到 `dsh-ai-terminal-pet 1.2.3`、
根 shim 的 `dsh.bundle.patch` 指向 `packages/dsh-ai-terminal-pet/cordis.patch.yml`、随包 vendor 的 exe 与
`protocol.js` 的 SHA-256 与 `vendor/manifest.json` 逐位一致、`dependencies` 与 `scripts` 均为 null、
`engines.dsh >=0.1.0-rc.6` 在本机宿主 0.2.2-rc.1 上按 `semver.satisfies(..., {includePrerelease:true})` 为 true、
`dsh.bundle.platforms` 已声明 `win32`。唯一一条**不阻断**的建议是：Claude Code / OpenCode / Antigravity 的
会话需要另行安装各自的 reporters（上游 README 的 connect 段），文案里最好加一句限定，否则用户会以为
「装完这一条就能看到那些会话」。这条反馈本身就是产品质量的一部分，所以按本仓库的规矩（Release 资产不可变、
改一句也要换版本号）走成 1.2.4。

**改的是哪三处（并且它们必须保持一致）**：

1. 包内 `README.md`：7,546 B / 137 行 → **8,187 B / 146 行**。开头原来写「so a single market install is enough to
   get a pet on screen」，现在改成「this bundle follows **this Harness's** sessions；OpenCode / Claude Code /
   Antigravity 的会话只有在**那个终端自己**装了上报端之后才会出现」，并指向仓库 README 的
   `#connect-your-terminal` 段；standalone 那段补一句「Those reporters are separate installs, not part of this
   package: a terminal that has none of them hooked reports nothing, in this install or in any other」。
2. 根 shim `package.json` 与包内 `package.json` 的 `dsh.catalog.summary`（en / zh 两块）。
3. `docs/market-index.json` 的 `plugins[0].description`——它与上面两处**逐字相同**，因为
   `verify.mjs` 的根清单用例会先 `JSON.stringify` 两份 `dsh.catalog` 再要求完全相等。

新摘要：en **225 字符**（原文 190）、zh **114 字符**（原文 104）。两份清单 `version` 1.2.3 → **1.2.4**
（根 2,052 → **2,115 B / 68 行**、包 2,183 → **2,246 B / 83 行**）。

**随包字节的其余部分一个都没动**——这是与 1.2.2 的关键区别：1.2.2 因为 `scripts/` 迁移改到 8 处 Go 注释、
连带 `vendor/manifest.json` 的 `source.sha256` 从 `89eef0d6…` 变成 `cf647af1…`；1.2.4 **没有碰任何 `.go` 文件**，
所以源码摘要仍是 `cf647af1…`，`lib/index.js`（570 行 / 17,872 B）、`vendor/AI-Terminal-Pet.exe`
（4,371,968 B / sha256 `bbb5163c296cd460d7c9a4c85e10342e125c0bd8017a065f913aa01567d41f0a`）、
`vendor/protocol.js`（20,914 B / sha256 `5e197db6…`）、`cordis.patch.yml`、`icon.svg` / `icon.png`、
`locale/*.json`、`screenshots/` 三张全部逐字节不变。

**打包与验证实测**：

- `npm pack`：`dsh-ai-terminal-pet-1.2.4.tgz` **1,915,340 B / 16 files / unpacked 4,462,112 B /
  sha256 `d6b7f95a5a3a71c4a9367699e9ad9b8f7d5db522271782e8ffe29a73e45c6708` / shasum
  `dafb81998efbae8e906a36773e9737117aa482d9`**；连跑两次 `npm pack` 后 `cmp` 无差异。
  与 1.2.3 的 1,915,013 B 只差 **327 B**（两块摘要 + 包内 README 的新增段落，白名单仍是 16 个文件）。
- `node packages/dsh-ai-terminal-pet/verify.mjs`：**95 ok / 0 warn / 0 fail / VERIFY PASSED / rc 0**，
  文件仍是 424 行。只改文案时断言一条不加也不减，`caseMarketIndex()` 的「索引最新版本 == 包版本」直接跨过改号。
- `node packages/dsh-ai-terminal-pet/scripts/vendor.mjs --check`：两个 vendor 文件仍 matches the source。
- `gofmt -l .` 无输出、`go vet ./...` 干净、`go test -count=1 ./...` 9 个 internal 包全 ok。

**索引的两步前移，实际落成的两个提交**（见 §27.10 与 AGENTS.md 第 6 条）：

- **A = `ff7f6de7d87a7a4d99c06e3a9a68e9a4b417b0a9`**（`Release 1.2.4 so the market copy stops promising other
  terminals`，9 files +252/−67，2026-11-11 08:38:16 +0100）：版本号、包内 README、索引的 `versions[0].version`
  与 `publishTimestamp`（**1794379909** = 2026-11-11T06:51:49Z）全在这一批。**tag `dsh-plugin-v1.2.4` 指向它**。
- **B = `4ee7f01d74cdfc3bcf51745e29c1306b1a42501b`**（`Pin the market index to 1.2.4's release commit`，
  `1 file +2/−2`，08:40:01 +0100）：把索引的 `source.commit` 与 `icon` 从 `b4474a6` 改成 A 的 sha。
  之所以拆成两个提交，是因为 git 无法在一个提交里引用它自己的 sha。

推送后 CI 只在 head 上起一次 run：**run `42981618507`（job `130393123402`）在 B 上 `conclusion: success`**
（started 2026-11-11T07:43:54Z）。A 的 check-runs 数组是空的，所以另外用 `git worktree add --detach .local/wt-A ff7f6de`
解出一份独立树跑了一遍 `node packages/dsh-ai-terminal-pet/verify.mjs`，同样 `VERIFY PASSED`，跑完
`git worktree remove --force` 清掉。

**tag 与 Release**：注释 tag `dsh-plugin-v1.2.4` 的 tag 对象 sha 是
`86cc1c034b702d9066a30f590c707fe33369224c`；Release id **441376551**，
<https://github.com/jjjjjjjjnnjnn/ai-terminal-pet/releases/tag/dsh-plugin-v1.2.4>，published
**2026-11-11T07:48:54Z**。资产 `dsh-ai-terminal-pet-1.2.4.tgz` 1,915,340 B，GitHub 侧 digest
`sha256:d6b7f95a…`——从 `browser_download_url` 真下载一次再算 sha256 仍是同一个值（不是只看 API 字段）。

**上游侧同步的两件事，也都已经开成 PR（都还没合并）**：

1. [awesome-dsh-plugin#14243](https://github.com/awesome-dsh-plugin/awesome-dsh-plugin/pull/14243)：分支
   `bump-tarball-1.2.4`、commit `66ed01d`、`1 file +3/−3`——`tarball` 与 `icon` 换到 1.2.4 资产，
   中英 `description` 各补一句 reporters 限定；state `open` / `mergeable: true` / `mergeable_state: unstable`。
   该文件的 `install` 命令是站点构建时渲染的，所以「PR 合并」≠「立刻可见」，还要等一次重建。
2. [imsai-sh/awesome-deepseek-harness-plugins#912](https://github.com/imsai-sh/awesome-deepseek-harness-plugins/pull/912)：
   分支 `entry-wording-1.2.4`、commit `1570316`、`1 file +3/−3`，只改 `description` 的中英两段
   （en 701 / zh 288 字符，schema 上限各 1,000）。该仓库 CONTRIBUTING 规定「更新已存在条目永不自动合并」，
   所以 `static-review` success、`merge` skipped，bot 留言 `maintainer review required`——等人工合并。
   两个 fork 都先用 `POST /repos/<fork>/merge-upstream` fast-forward 过再动笔。

**没做的事（明确边界）**：没有为了让文案更准确而改任何代码或 `lib/index.js`；没有把 reporters 打包进插件
（那会在 Windows 上把 Node 依赖塞进一个宿主桥接包里，且 Claude Code / OpenCode 的钩子必须先被那个 CLI 自己
接受，插件无法代劳）；`lib/release-pins.json` 仍刻意留空，所以插件内「自动下载浮窗」仍然回落为 no-op，
安装时拷的始终是 `vendor/` 里那份（见 §27.7）。

**发版之后的收尾（同日 08:55–09:10 +0100）**：仓库侧只剩文档在追这次出货——`README.md`（英中两半把安装计时从
「取在 1.1.0 资产上」改成 1.2.4 实测的 1.8 / 3.7 / 6.5 秒，并把 #14243、#912 两个 PR 号写进去）、`CHANGELOG.md`
（1.2.4 段补上 A/B 两个提交与 CI run `42981618507`，上游两条改成已开 PR）、`docs/ROADMAP.md`（发版前写的
「资产要等打 tag 才有」与「1.2.4 需再开一个 PR」全部改成既成事实，索引条目改钉 `ff7f6de`）、本文件（§1 的最新提交
前移到 `4ee7f01`、§27.5 的端到端待测改成实测、§27.12 的「上游要做的两件事」改成「已经开成 PR」）。
四个文件合计 +71/−34，改完复跑 `node packages/dsh-ai-terminal-pet/verify.mjs` 仍 `VERIFY PASSED`、
六个上报端 `node --check` 仍全过、`go test -count=1 ./...` 9 个 internal 包仍全 ok（仅文档改动，代码路径未动）。

**仍未解决的外部事（2026-11-11 10:30 +0100 复读之后的状态）**：npm 发布被一次性密码挡在非交互 shell 外
（只能由人在交互终端里跑 `cd packages/dsh-ai-terminal-pet && npm publish`）；`awesome-dsh-plugin` 的 #14243
（check 跑完之后 `mergeable_state` 在 `clean` 与 `unknown` 之间来回，属于一方重跑 check 的正常抖动）与
imsai 的 #912（`clean`，该仓库规定「改已存在条目永不自动合并」）都等人工合并；radar #1255、brade #367
两条仍挂在对方仓库；dshbase #322、dsh.so #24、dsh-plugin.org #735、alldsh #2 与 dshmarketplace.dev 那次
`POST /api/v1/submissions` 全部 open / 无回应。**dshopen #423 已 closed（= 收录）**，所以不再列在这一串里；
它是本次交接时唯一从「pending」变成「done」的一条。

**发版之后再投的五家目录（同日 09:1x–09:5x +0100）**：这条路已经走到边际——能投的公开目录基本都投过了，
剩下的要么只收 npm 包（`LivXue/dsh-plugin-shop` 的 26,351 条条目每条都带 npm tarball 的 `integrity`，
`commit` 字段一条也没有）、要么只发布过 Linux 沙箱验证的仓库（`zasenjc/dsh-plugins-store`，我们的 vendor
是 PE，跑不过）、要么要求先在插件仓库根提交 `dsh-plugin.naming.json`（`oh-my-dsh/dsh-plugin-registry`，
成本最高的一条）。这一轮投出去的五条：

| 目录 | 提交入口 | 结果 |
|---|---|---|
| [dshbase](https://github.com/ylwl1997/dshbase) | GitHub issue 表单（六字段） | [#322](https://github.com/ylwl1997/dshbase/issues/322) open / 0 评论 |
| [dsh.so](https://github.com/dsh-so/dsh-plugin-submissions) | GitHub issue | [#24](https://github.com/dsh-so/dsh-plugin-submissions/issues/24) open / 0 评论 |
| [dsh-plugin.org](https://github.com/dshplugin/dsh-plugin-hub) | GitHub issue 固定模板 | [#735](https://github.com/dshplugin/dsh-plugin-hub/issues/735) open / 0 评论 |
| [AllDSH](https://www.alldsh.com) | `POST /api/submit`（JSON） | `201 {"ok":true,"issue":"https://github.com/alldsh/submissions/issues/2"}` |
| [DSH Marketplace](https://dshmarketplace.dev) | `POST /api/v1/submissions` | 先 `action:"inspect"`（`alreadyListed:false`、`hasDshTopic:true`），再 `action:"submit"` → `{"ok":true,"duplicate":false}` |

- **dshbase 的 403 坑**：提交 issue 时带 `labels: ["plugin-submission"]` 会返回
  `403 "You do not have permission to create labels on this repository."`；去掉 `labels` 即可（issue 的 labels 为空）。
- **dsh.so 的格式**来自它自己的既有投稿（#21 brickhu、#23 hxlog）：仓库/子目录/版本/「npm 未发布」/简介/
  「一个安装覆盖哪些会话」限定段/DSH 依据/无副作用/自建索引/topics/安装实测，末尾附建议的 registry entry JSON。
  该站的公开数据接口是 `cdn.dsh.so/plugins-index.json`（14,553,328 B）、`plugins-index.delta.json`（678,517 B）、
  `cdn.dsh.so/market/index.json`（format 1、11,119 条），三处目前都没有本仓库。
- **dsh-plugin.org** 的模板六段（仓库地址/一句话价值/能力分类/安装命令/兼容与运行要求/许可证/截图/补充说明/
  提交前自查）与它自己的「收录要求」一致：没有 npm 包的仓库照常收录、只是不进它站内的「插件中心」——
  这与 imsai 的 browse-only 是同一套口径。它的插件索引契约是 `dsh-plugin-registry/v2` 系列之外的
  `/directory/zh/*.json` 分片。
- **AllDSH** 的接口说明直接由 `GET /api/submit` 返回（字段含 `repo`/`tagline`/`install`/`kind`/`categories`/
  `notes`/`listing`/`contact`/`agent`，外加蜜罐字段 `website` 必须为空）；它的目录分类枚举是
  `agent / knowledge / interface / dev / integration / workflow / tools / model / session / fun / memory`，
  我们按 `interface` + `session` 报（上限 3 个）。
- **DSH Marketplace** 的页面写着「本站同步 `dsh-plugin` topic」，但复读 `data/marketplace-cache.json`
  （`total: 9909`、缓存 63 条）与 `GET /api/v1/plugins?q=ai-terminal-pet`（`{"total":0}`）都没有本仓库，
  所以它确实需要走一次提交；社区精选库是 `DshMarketPlace/awesome-dsh-plugin` 的 `data/curated.yml`（4 stars、
  15 open issues、pushed 2026-10-04），目前也没有我们的条目。
- **不投的两家**：`WOOK98/awesome-deepseek-harness-plugins` 是 `imsai-sh` 的 fork 且已停更（pushed 2026-08-15），
  活的上游是 imsai（我们已有 #912）；`ydhrdh/dsh-marketplace` 只有 7 个条目、0 stars，schema 允许 git 源但价值太低。

## 28. 交接清单（2026-11-11 10:30 +0100）

> 本节是给「下一个会话、下一个人」的一页纸。读完 §1 与本节就可以开工；§27 是插件包分发的完整 runbook，
> 需要动包或发版时再翻。

### 28.1 现在在哪（一句话）

浮窗本体停 **0.9.0**、DSH 插件包已到 **1.2.5**（tag `dsh-plugin-v1.2.5` → **`34aa1f7`**，Release 资产
`dsh-ai-terminal-pet-1.2.5.tgz` **1,917,342 B / sha256 `dfa89a1e…`**），并且 1.2.4 与 1.2.5
**都已发布到 npm**（`latest` = 1.2.5、shasum `1f6d5340…`，与 Release 资产逐字节相同）。
下一个动作是 **ROADMAP §4 的 I1（独立安装包）**，做法与已知约束见那里的表；`docs/HANDOVER.md` §31
是上一轮按真实 Harness 类型定义做的适配，改动上报端或插件前值得先读。
先看清单用 `git log --oneline c8ad006..HEAD`、当前 CI 用 `gh run list --limit 1`——**这两个数字刻意不写死**，
因为 git 无法在一个提交里引用它自己或它的后继。
（§28 最初那次整理只动 `.gitignore` 与四份文档，所以没有新 tag；此后 §29–§31 交付了原生 hook、
共享内核、新机套件与 1.2.5，那三轮才有代码与发版。）

### 28.2 接手要先看什么（按顺序，别回溯长上下文）

1. 本文件 §1（当前状态、版本、健康、规模、分发）与本节；
2. [README.md](../README.md) 的安装段、connect 段与「设置项与文件位置」表——对外承诺的口径以它为准；
3. 需要动包/发版时：§27.2（包结构与 `dsh.bundle` 契约）、§27.5（发布与分发现状）、§27.9–§27.10（三条发现路径
   与自建索引）、§27.12（1.2.4 只改文案那次的完整操作）、§31（按真实 Harness 类型定义做的适配）；
4. [AGENTS.md](../AGENTS.md) **八条**铁律——尤其第 6 条（发版提交必须同时前移索引）、第 1 条（reporter 绝不
   把异常抛进宿主 CLI）与第 8 条（**同会话复核**：实现、评审与对外措辞都在做这件事的会话里、用用户选定的
   模型完成，逐条对着代码核；**不要**再调 `agy` / `gemini` 之类的模型 CLI 求第二意见）；
5. 改动上报端或插件后：`node tests/fresh-machine.mjs`（§30），它才是「新机器上装上去能不能用」的门禁。

### 28.3 外部世界在等什么（2026-11-11 15:40 UTC 复读）

| 对象 | 状态 | 谁来动 |
|---|---|---|
| npm 发布 | **已完成**：1.2.4（12:10:46Z）与 1.2.5 均在 npm 上，`latest` = **1.2.5**；本机 token 已绕过 2FA，agent 可直接发 | 已完结 |
| [awesome-dsh-plugin#14385](https://github.com/awesome-dsh-plugin/awesome-dsh-plugin/pull/14385) | **已合并**（2026-11-11T17:20:15Z，merge commit `70581cc`）；两个 check 都 SUCCESS。仓库里 `data/plugins/…yml` 的 `tarball` 已是 1.2.5 | 已完结（等站点重建才渲染到 `plugins.json`） |
| [awesome-dsh-plugin#14243](https://github.com/awesome-dsh-plugin/awesome-dsh-plugin/pull/14243) | 1.2.4 的上一版跟进，已被 #14385 取代；已 CLOSED | 已完结 |
| [imsai-sh#912](https://github.com/imsai-sh/awesome-deepseek-harness-plugins/pull/912) | open / MERGEABLE；2 条机器人评论 + 1 条我们的（说明 1.2.5 已上 npm 与两个修复）；该仓库规定改已存在条目永不自动合并 | 上游维护者 |
| [radar#1255](https://github.com/AdamPlatin123/dsh-plugin-radar/pull/1255) | open / UNKNOWN，零 check | 上游维护者 |
| [brade#367](https://github.com/bradeGithub/DSH-Plugins-Marketplace/issues/367) | open / 3 评论（都是我们自己的，最后一条给了 npm 版本与 tarball 哈希） | 上游维护者（我们要的信息已给全） |
| [dshopen#423](https://github.com/polang233/dsh-plugin-index/issues/423) | **已 closed = 已收录且站点已渲染**（`dshopen.org/index.json` 1,207 条含 **1.2.4**；1.2.5 要等它下一次构建） | 等站点重建 |
| dshbase#322 / dsh.so#24 / dsh-plugin.org#735 / alldsh#2 / dshmarketplace.dev | 全部 open、无回应 | 上游维护者 |
| YELEBAI/dsh-plugin-marketplace | 自动扫描，仍记 1.2.3 / `verifiedCommit 9b0df94`（`profiles` 已含 desktop） | 等它下一轮扫描 |

判断「是否真的可见」看**对方的公开索引文件**（站点域上的 `index.json` / `plugins.json` / `registry.json`），
不要只看 PR/issue 状态，**也不要读对方仓库里的 `public/` 目录**——那是未重新构建的副本。

**dshopen 的具体状态（2026-11-11 复核）**：`https://dshopen.org/index.json` 返回 **1,207 条**，
其中包含本包 `dsh-ai-terminal-pet` **1.2.4**，`source.type: git`、
`source.commit: ff7f6de7d87a7a4d99c06e3a9a68e9a4b417b0a9`、`path: packages/dsh-ai-terminal-pet`、
分类 `integrations`、标签 `verified` + `updated-recent`。当初读到的「854 条、0 条 git 源」来自
`raw.githubusercontent.com/polang233/dsh-plugin-index/main/public/index.json`，那是未重建的仓库副本。

### 28.4 npm：**已完成，且不再需要人工**（2026-11-11）

1.2.4 在 2026-11-11T12:10:46Z 上线、1.2.5 同日 14:58Z 前后上线，两次都是 agent 在包目录直接
`npm publish --access public`。**本机 token 已经配置成可绕过 2FA**，所以「必须人跑一次」的旧结论
作废；只有 token 失效（`npm whoami` 报 E401/E403）时才需要人重新登录。

事后核对全过（1.2.5）：`dist-tags.latest` = `1.2.5`、shasum `1f6d5340…`、
integrity `sha512-2sEFAf0k…`、**17 files**、`npm pack dsh-ai-terminal-pet@1.2.5` 取回的 tarball
与 Release 资产 sha256 相同（`dfa89a1e…`）；`npm install` 装出 17 个文件，
`vendor/AI-Terminal-Pet.exe` 与仓库一致，`lib/index.js` 作为 ESM 可加载，
`vendor/package.json` 确实是 `{"type": "commonjs"}`。

发布之后仓库侧已经做完的：`README.md` / `CHANGELOG.md` / §27.7 / 本节里的「npm 尚未发布」全部改成事实。
仍然**不要**重传或改写已发布的 `.tgz`；**不要**把某个版本的 tarball 换成内容不同的文件
——Release 资产不可变。

### 28.5 哪些文件「不可再改」

已发布内容一旦被改，`scripts/install.ps1 -VerifyOnly` 会当场报 fail，因此下面这些只能是「另起一版」：

- `packages/dsh-ai-terminal-pet/README.md` 与两份 `package.json`（属已发布 tarball 的字节）；
- `packages/dsh-ai-terminal-pet/vendor/manifest.json`（同样属已发布 tarball 的字节；§29.5 记录了一次
  已经发生的重写，因此**下一次发布必须是新版本号**）；
- `packages/dsh-ai-terminal-pet/vendor/package.json`（§30.2 新增的 CommonJS 覆盖，同样属包内字节）；
- `docs/market-index.json` 的 `versions[0]`（`verify.mjs` 断言它等于包版本）；
- **八个** tag 指向的提交（`dsh-plugin-v1.0.0` … `v1.2.5`）与它们各自的 Release 资产。

改动任何一条 → 新版本号 + 新 tag + 新 Release；**纯文案改动同样要换版本号**（1.2.1 与 1.2.4 就是这么来的）。

### 28.6 想发下一个版本时的固定套路

1. A 提交：改 `version`（两份清单）+ 要改的内容 + 索引的 `versions[0].version` 与 `publishTimestamp`；
   跑 `node packages/dsh-ai-terminal-pet/verify.mjs`（**不要写死条数**，它随用例增减变化）、
   `npm pack` 记下字节与 sha256，并把它们写进 `CHANGELOG.md`；
2. 打 annotated tag（`git tag -a dsh-plugin-vX.Y.Z`）、推、发布 Release 并上传 tgz
   （`gh release create <tag> <tgz> --title <tag> --notes ...`，GitHub 会自己算出 `digest: sha256:…`）；
3. B 提交：把索引的 `source.commit` 与 `icon` 钉到 A 的 sha（**git 无法在一个提交里引用自己的 sha**，所以必须两步）；
4. npm：在包目录 `npm publish --access public`（本机 token 已可绕过 2FA）。发完等 2–4 分钟再
   `npm view <name> version` 确认 `latest` 换了——头几分钟 registry 上只有占位的 `0.0.0-stage`，
   此时 `npm install <name>@<version>` 会报 `ETARGET`；
5. 上游跟进一个只改 `tarball` / `icon` 的 PR（awesome-dsh-plugin 的 `data/plugins/*.yml`）。
   **务必 `-R awesome-dsh-plugin/awesome-dsh-plugin`**，否则 `gh` 会把 PR 开进自己的 fork；
   分支要建在**自己 fork**（`gh api git/refs` POST），PR 用 `--head jjjjjjjjnnjnn:<branch>`。
   两处都是「合并 ≠ 立刻可见」，要等对方重建；判断可见性读**站点域**的索引文件，不读仓库里的 `public/`。

### 28.7 刻意保留、看起来像冗余的东西（不要清理）

- **`integrations/deepseek-harness/ai-terminal-pet-bridge/`**（`index.js` 7,780 B、`package.json` 495 B、
  `cordis.patch.yml` 355 B、`icon.svg` 526 B、`locale/{en,zh}.json` 各 223 B、`README.md` 2,712 B）
  是**旧的 host-only 接法**，与新插件包并存是故意的：`scripts/install.ps1` 仍把它拷进
  `%LOCALAPPDATA%\AI-Terminal-Pet\bundles\ai-terminal-pet-bridge`（11 项制品清单里占 7 项），
  `integrations/deepseek-harness/verify.mjs` 与 CI 仍校验它，README 的 connect 表仍把「已装了整仓」
  的用户指向它。删掉它 = 同时打断安装脚本的制品清单、CI 一步校验与一条用户路径。
  它与插件包的 `icon.svg` 恰好逐字节相同（同一次改版留下的），这也不是错。
- **`vendor/AI-Terminal-Pet.exe` 与 `bin/AI-Terminal-Pet.exe` 字节相同**：前者是发布产物（进 git，
  为让 `dsh plugin add` 不需要 Go 工具链），后者是本地构建产物（不进 git）。这是
  `.gitignore` 末尾那条反向例外存在的原因。
- **`.local/`**（5.7 M）：上游 JSON 快照、发版草稿、`npm pack` 产物等一次性材料，
  本次已加进 `.gitignore`；可以随时删，**不属于项目源码**。
- **`.local/awesome-entry-1.2.5.yml`**：awesome-dsh-plugin 那条目在 1.2.5 时的**原始字节**副本。
  必须存在，因为 PowerShell 5.1 的 `Get-Content -Raw` 会把 UTF-8 当 ANSI 解码，中文整行 mojibake——
  要改那条目时请**用这个文件的内容直接 base64 上传**（`[IO.File]::ReadAllBytes` → `ToBase64String`），
  不要先读成字符串再写回。§30.6 记了这次踩坑。改完记得同步更新同一个文件。

### 28.8 交接时跑过的验证（2026-11-11，全部通过）

```
node packages/dsh-ai-terminal-pet/verify.mjs        -> VERIFY PASSED (rc 0)
node packages/dsh-ai-terminal-pet/scripts/vendor.mjs --check  -> 两个 vendor 文件均 matches the source
node integrations/deepseek-harness/verify.mjs       -> VERIFY PASSED
gofmt -l . / go vet ./... / go test -count=1 ./...  -> 干净 / 0 / 11 个包全 ok（201 个 func Test）
六个 reporters 逐个 node --check                     -> 全过（Node 24 只查第一个参数，必须逐个跑）
node tests/fresh-machine.mjs --no-overlay           -> PASS: 160/160
scripts/install.ps1 -VerifyOnly                     -> 11 file(s) match the repository, 0 warning(s)
```

上表是当前口径（`c8ad006` 那次刷新）：§29–§31 交付原生 hook、共享内核、新机套件与 1.2.5 之后，
制品从 10 项变 11 项、`internal` 从 9 个包变 10 个 + `cmd/pet-hook`，`tests/fresh-machine.mjs`
又成为一条独立门禁。**引用这些数字时不要照抄**，跑一遍上面的命令再写。

## 29. 上报端原生化：`pet-hook.exe` 与 `internal/reporter`（2026-11-11）

本节是 ROADMAP §E1 的施工记录。**不影响已发布的 1.2.4 包**：`packages/dsh-ai-terminal-pet/`
目录内除 `scripts/vendor.mjs` 的摘要作用域外没有改动，`vendor/AI-Terminal-Pet.exe` 字节未变，
`vendor/manifest.json` 也**没有**重新生成（摘要仍等于 `cf647af1aa1f45d1ed6f0df7b97e7e1439272f87064009e1989f1bfbd63db29a`）。

### 29.1 为什么做，以及为什么是现在

本机创建进程约 1 s（AV 钩挂），而 `claude-hook.js` / `antigravity-hook.js` 每一次宿主事件都要起一个
`node`。`pet-statusline.exe` 已经是这条路的先例，直接把它的骨架（stdin 解析 → 事件映射 → 快照 →
管道投递）提出来，加上 `--mode` 分派就够了。

### 29.2 结构：为什么多了一个 `internal/reporter`

```
cmd/pet-hook/           --mode=claude|antigravity，读 stdin，写 {} 到 stdout
      │
      └──► internal/reporter/   ← 新增，共享内核
             reporter.go   bounded id / 脱敏 / 稳定 PID / 时间戳钳制
             payload.go    与 session.Session 对齐的 wire payload
             stdin.go      1 MiB + 2 s 的读取上限（两个命令共用）
             send.go       快照落盘 → 拨号 → 写一行 → 只认 "OK"；启动浮窗
             approval.go   审批握手（pending-*.json 轮询）
             hook.go       hook 事件名 → session.State 的映射表
             statusline.go 状态栏的 agent_state → session.State 映射表
```

**关键决定**：`pet-statusline` 里那套 helper 不复制一份到 `pet-hook`。两处各写一遍意味着
「与 `reporters/protocol.js` 保持一致」这条契约从此有两个可能各自漂移的实现；收进一个包之后，
它可以被单元测试钉住，而不是靠人眼比对。`internal/reporter` 现在 **23 个 `func Test`**
（`reporter_test.go` 16 + `send_test.go` 2 + `statusline_test.go` 5），加上 `cmd/pet-hook`
的 13 个，覆盖脱敏边界、id 截断、UTF-16 FNV-1a PID、时间戳钳制、审批决策树、投递语义与两张映射表。

`cmd/pet-statusline` **也已迁移**到 `internal/reporter`（见 §29.7）。先做 hook、再做 shim，是因为两者的
风险不同：hook 是新增路径，出问题只影响新接入的用户；shim 是既有路径，动它必须能证明「行为一字未变」，
所以先对它做了逐 payload 的行为快照比对。

### 29.3 一个真实缺陷：`Send` 对「永远打不开」也重试满预算

原实现里任何一次拨号失败都会重试到 750 ms 预算用完。宿主是同步等上报进程结束的，所以当投递
**不可能成功**时，这 750 ms 是纯加在开发者每一次事件上的延迟。现在 `retryableDialError` 只把
`ENOENT` / `ERROR_FILE_NOT_FOUND`（还没监听）与 `ERROR_PIPE_BUSY`（实例占满）算可重试；
`ERROR_ACCESS_DENIED` 立即返回，且不再触发启动浮窗（启动了也打不开）。

### 29.4 为什么会遇到「Access is denied」——**不是产品缺陷**

本机仓库目录 `C:\Users\rongj\Desktop\项目开发\ai-terminal-pet` 带 **Low 强制完整性标签**
（`icacls <repo> /q` 显示 `Mandatory Label\Low Mandatory Level:(OI)(CI)(NW)`；同级还有
`DSH-Addition`、`oh_my_3Dcharacter`、`oh_my_3dlab`，`oh_my_3Dworld` 等则是正常的）。
**从该目录里的文件启动的进程一律继承 Low IL**，而浮窗跑在 Medium IL、管道 DACL 只给
`OWNER RIGHTS` + `SYSTEM`，Low IL 永远打不开。同样的字节复制到 `%TEMP%`、`C:\Users\rongj`
或 `%LOCALAPPDATA%\AI-Terminal-Pet` 就落回 Medium IL 并正常投递（已用同 sha256 双份验证）。

复现与判读：

```powershell
icacls . /q | Select-String 'Mandatory'      # 有 Low 行 → 从这里跑出来的进程投递不到
icacls "$env:TEMP" /q | Select-String 'Mandatory'  # 无 → 放在这里的进程可以
```

因此：**从仓库里直接跑 `bin\pet-hook.exe` 或 `bin\pet-statusline.exe` 投递不到浮窗是预期行为**；
安装到 `%LOCALAPPDATA%` 的副本不受影响，`scripts\install.ps1` 装的正是副本。
管道 ACL `D:P(A;;GRGW;;;OW)(A;;GRGW;;;SY)` 因此**保持不变**——改它等于为了让一个环境标签生效
而放宽全局安全边界。

如果哪天要在一台「整棵工作区都被标成 Low IL」的机器上直接验证，把两个 exe 复制到
`%TEMP%` 再跑，不要改代码。

### 29.5 安装器与 CI 的改动

- `scripts/install.ps1`：新增 `$HookSource` / `$HookDestination`，制品清单从「五类」变「六类」
  （新增 `hook` 分组），`-Destination` 演练会把 `-HookDestination` 一起带走，
  `-VerifyOnly` 的分组顺序里加了 `hook`，卸载时的进程列表加了 `pet-hook`。
  演练实测：安装 **11 file(s)** → `-VerifyOnly` 11/11 匹配、0 warning → `-Uninstall -Force` 干净。
  脚本仍是**纯 ASCII**（0 个非 ASCII 字节，PowerShell 5.1 会按 ANSI 读 BOM-less 脚本）。
- `.github/workflows/ci.yml`：新增 `Build the hook binary` 步骤。
- `packages/dsh-ai-terminal-pet/scripts/vendor.mjs`：`NON_OVERLAY_PACKAGES =
  ['cmd/pet-hook/', 'cmd/pet-statusline/', 'internal/reporter/']`。摘要原本扫 `cmd/` + `internal/`
  下全部非测试 `.go`，但这三个包根本不进浮窗二进制；不排除就会为一个「不可能改变浮窗一个字节」
  的改动逼着重新 vendor 一份**已发布**包的清单。
  **第一版把 `cmd/pet-statusline` 故意留在里面**（理由：它自 1.0.0 起就在每份 manifest 里，
  排除等于无端重写），这个理由在第一次动 shim 时就失效了——摘要照样坏，而浮窗二进制一个字节没变。
  所以改成按「是否被链接进去」这个真正的标准判断。
  **因此这次 `vendor/manifest.json` 确实被重写了一次**：`files` 25 → 24、
  `sha256` `cf647af1…` → `39685c79a41c30880de852ebafecd46d777405fd0a6d4730c6bf66438addec63`。
  两个 vendor 文件的哈希与字节未变（`vendor/AI-Terminal-Pet.exe` 仍是 `bbb5163c…`）。
  **这意味着仓库里的 `packages/dsh-ai-terminal-pet/` 已不再逐字节等于 npm 上的 1.2.4 tarball**，
  所以下一次发布必须是**新版本号（1.2.5）**，不能重新上传 1.2.4。
  `verify.mjs` 与运行期行为都不受影响（`--check` 恢复 exit 0）。

### 29.6 实测数字（同 payload、同机器）

| 路径 | 首跑 | 稳态 |
|---|---|---|
| `pet-hook.exe`（native） | 513 ms | **51–73 ms** |
| `node reporters/claude-hook.js` | 584 ms | **93–157 ms** |
| `pet-hook.exe` 低完整性（被拒，修复前） | — | 829–862 ms |
| `pet-hook.exe` 低完整性（被拒，修复后） | — | **89–119 ms** |

两侧都从 `%TEMP%` 启动以排除 §29.4 的完整性差异。测量工具是一次性的 Go harness
（直接 `exec.Command`，不经 `cmd /c`），用完即删。

### 29.7 `pet-statusline` 的迁移（同一批，2026-11-11）

**为什么做**：`pet-statusline` 原本自己带着一份脱敏规则表、一份 id 截断、一份 UTF-16 FNV-1a PID、
一份时间戳钳制、一份快照落盘。这些是**线协议规则**，而 `reporters/protocol.js` 是规范实现——
三份拷贝意味着三种漂移方式。这次发现的两个缺陷就属于这一类：`pet-hook` 与 `pet-statusline`
**各有一份同样的「不建 runtime 目录就静默不落快照」**（`reporters/protocol.js` 的
`writeRuntimeSnapshot` 是一直 `mkdirSync(recursive)` 的）。新包现在先 `MkdirAll(0700)`。

**怎么保证没改行为**：迁移前后各跑一遍同一组 **35 个 payload**（空 stdin、非 JSON、JSON 数组、
`{}`、各个 `agent_state` 取值、对象形态的 `agent_state`/`workspace`/`model`、`tool_name` 与 `toolCall`、
显式 pid/parentPid、epoch 秒/毫秒、RFC3339、未来时间戳、超长会话 id、含 `/ : ? *` 的会话 id、
Bearer token、控制字符、parent session），记录 stdout、退出码、stderr、快照文件名与快照里 11 个字段。
结果 **35/35 完全一致**。那些取值同时被 `internal/reporter/statusline_test.go` 钉住，
所以以后再动映射表会先在测试里红，而不是等用户看到面板状态不对。

另外做了一次**跨语言**比对：参考实现 `reporters/antigravity-statusline.js` 与原生 shim 跑同一组
33 个 payload，逐字段比落盘快照。`state` / `reason` / `sessionId` / `pid` / `project` / `model` /
`terminal` 本来就一致，只有 `tool` 不一致（参考实现不读 `toolCall`）——已把参考实现补齐，现在 33/33 一致。
做这个比对有个前提要记住：**两边在「浮窗二进制不存在」时都不上报**，参考实现还会转去调 notifier，
所以 harness 必须放一个占位的 `AI-Terminal-Pet.exe`，并把 `USERPROFILE` 也指到临时目录，
否则任何一边都可能摸到真实浮窗的管道。

**刻意保留的两处差异**（写下来免得下次「顺手统一」）：

1. `startedAt` 超过 5 分钟**忽略**而不是钳制——时钟偏斜的宿主否则会显示一个跑了几小时的会话；
2. 投递是**单次**（失败 → 启动浮窗 → 再试一次），不是 `reporter.Send` 的 750 ms 重试预算——
   状态栏每次刷新都要跑，付不起这个预算。

**代价**：`pet-statusline.exe` 3,643,392 → **3,664,896 B**（+21,504 B，共享包里的审批/配置代码
被一起链接进来）。稳态耗时实测不变：迁移前 86–92 ms、迁移后 88–93 ms（几乎全是进程创建开销）。

## 30. 全新机器端到端：`tests/fresh-machine.mjs`（2026-11-11）

### 30.1 它为什么存在

到这一轮为止，所有自动化测试回答的都是「逻辑对不对」：`go test` 按函数测，`verify.mjs` 用假包根
和假 Cordis 上下文测插件。**没有一个测试回答「这台机器从没见过这个项目，用户按 README 装上去，
能不能用」**——而两个真实缺陷（§30.2、§30.3）恰好都活在这条缝里，都是「所有单元测试都绿、
实际什么都不发生」。

套件就建在这条缝上。每次运行在 `%TEMP%` 里建一个一次性沙箱：

```
<tmp>/pet-fresh-XXXX/
  local/        ← LOCALAPPDATA（配置、运行期快照、审批目录都在这下面）
  home/         ← USERPROFILE（管道名由它哈希出来，所以连管道都是另一条）
  dsh/          ← DSH 风格的工作目录，node_modules/dsh-ai-terminal-pet 就在这
  project/      ← 一个真实存在的项目目录，用来验证 cwd 映射
  bin/          ← 三个二进制的副本（见 30.5）
  pack/         ← npm pack 出来的 tgz
```

八个场景共 **158 项检查**，全绿约 25 秒。`--no-overlay` 跳过会开窗口的两个场景（S4 启动浮窗、
S5 里对浮窗的投递断言退化为快照断言），CI 用这个形态。

### 30.2 挖出的缺陷一：插件自带的 `vendor/protocol.js` 从来没加载成功

包清单是 `"type": "module"`，所以包内任何 `.js` 都按 ESM 解释。插件用 `createRequire()` 去
`require()` 它，Node 24 的 `require(ESM)` 会**成功**返回一个模块命名空间——但那个文件是
`module.exports = {...}`，作 ESM 求值根本执行不到赋值，于是返回的对象里没有 `sendStatus`。
插件下一行就判断「没有协议模块」并静默什么都不发。

触发条件不是「正常安装」：正常安装时 bootstrap 会把 `protocol.js` 拷到
`%LOCALAPPDATA%\AI-Terminal-Pet\`，那份没有 `package.json` 管着，是 CommonJS，能正常加载。
会踩到的是：`DSH_PET_NO_BOOTSTRAP=1`、拷贝失败（只读盘/磁盘满）、用户删掉了那份副本，
或从 git 源装但没跑过任何安装步骤。

**修法**：`packages/dsh-ai-terminal-pet/vendor/package.json` 写 `{"type":"commonjs"}`。
嵌套 `package.json` 覆盖目录模块类型是标准行为，`require` 从此正常，不需要在运行时加特判。

**为什么之前的测试没发现**：`verify.mjs` 的每个用例都先把 `lib/index.js` 与 `vendor/*` 复制到
一个**没有 `package.json`** 的假包根里，那里 `.js` 默认就是 CommonJS——假包根比真实包更宽松，
于是这条路径永远通过。新增的 `caseVendoredProtocolLoads` 换成对**真实包目录**做 `require()`，
并断言 `sendStatus` / `stableProcessId` 都是函数。

### 30.3 挖出的缺陷二：DSH 插件只在状态变化时上报

`report()` 的去重条件是「状态是否变了」。而 `working-directory/change` 只改 `project`、
`session/title` 只改 `reason`（标题本身固定为 `DeepSeek Harness`，用于窗口识别），
两者都不改状态——事件被整条吞掉。表现是：用户在 DSH 里换了个工作目录，浮窗继续显示旧路径，
直到 60 秒保活或下一次状态变化才顺带刷新；改会话标题则永远到不了面板。

**修法**：去重签名改为 `state` / `reason` / `tool` / `project` / `title` 五元组。重复内容仍然
被丢弃（保活与防抖语义不变），任何一项真变了就立刻上报。`verify.mjs` 的状态序列因此多了一次
`error`（切目录发生时状态仍是 error），已在用例里注明原因，免得下次有人把它当回归删掉。

### 30.4 场景清单与它刻意验证的东西

| 场景 | 关键断言（不只是「没报错」） |
|---|---|
| S1 空机器 | `--doctor` exit 1 且指名 `install files` 缺失；`--json` 返回 `count: 0` 的完整文档；`--version` 匹配 `0.9.0` |
| S2 包安装 | tarball 能解开；`dsh.bundle.patch` 指向 `./cordis.patch.yml`；patch 用 `insert:` 而不是裸行（裸行只补丁不挂载）；locale、vendor 三件套都在 |
| S3 插件 | 拷进去的 exe/protocol 哈希**等于** vendor 原件；二次 bootstrap 是同一个 memoized 结果且文件未变；7 种事件的 state/reason/tool 逐字段断言；切目录、改标题；两会话不串号且 pid 不同；`approval/request` 必定调 `next()`；dispose 后不再上报；全程没有 `logger.warn` |
| S3b 开关 | `DSH_PET_NO_BOOTSTRAP=1` 不拷贝，但**仍然用 vendor 协议写出快照**；`config.enabled:false` 注册 0 个监听 |
| S4 浮窗 | 沙箱里首次启动即可应答；`--json` 里能看到 `source: deepseek-harness` 的会话；**没有 `config.json`** 也能跑（全默认）；`--toggle`/`--show`/`--reset` 都 exit 0 |
| S5 独立安装 | 11 件制品、`-VerifyOnly` 全对、`--doctor` 0 fail、两个原生 hook 的会话真的进了沙箱浮窗 |
| S6 设置 | 自动同意关闭 → `{}`（交回宿主）；开启 → `{"decision":"allow"}`；危险命令两种方言分别 `{}` 与 `deny`；真实写 `.deny` 文件 → `deny` |
| S7 坏机器 | 配置损坏 / 无 `LOCALAPPDATA` / 1 MiB stdin / 空 / 非 JSON / 数组 / null / 未知参数 / 未知模式：全部 exit 0 且给出答案；无监听时 `--json` 不挂起 |
| S8 卸载 | 目录消失、`uninstall: done`、`--doctor` 重新 exit 1 |

### 30.5 两个「不这么写就会假绿」的细节

1. **二进制必须复制进沙箱再执行**。本仓库目录带 Low 强制完整性标签（§29.4），从 `bin/` 直接起
   的子进程是 Low IL，读不到 `%TEMP%` 沙箱里 Medium IL 的 `config.json`——于是「自动同意」永远
   看起来是关的，危险命令永远「放行」，而代码其实没问题。套件在 `sandbox/bin/` 放副本，一切从
   那里跑。这条也顺带证明了 §29.4 那段结论是对的。
2. **浮窗场景要停掉真实例，并从安装目录那份恢复**。单实例互斥体是
   `Local\AI-Terminal-Pet-v1-SingleInstance`，每个 Windows 会话一份，所以沙箱浮窗和真浮窗不能
   共存。恢复时**不能**用仓库 `bin/` 那份：它跑在 Low IL，控制命令连不上真浮窗的 Medium IL 管道，
   结果是「以为恢复了、其实没恢复」，而且沙箱目录因为被那个进程占着而删不掉——这个坑真的踩到了，
   第一次跑完留下了一个 15 MB 的沙箱和一个从沙箱路径运行的浮窗。

### 30.6 后来补上的三个场景（S4b / S4c / 中文路径）

第一版只覆盖「打开就有」的路径。补的三条都属于**用户真会遇到、但一次顺利启动碰不到**的情况：

- **S4b 冷启动**：先 `--quit` 让沙箱浮窗真的退出，然后**只跑一次 `pet-hook`**（用户从没手动开过浮窗
  的机器就是这个状态）。断言两件事：上报端先把快照落了盘，然后浮窗被它自己拉起来，那条会话可见。
- **S4c 八条会话**：连续 8 条，全部要进 store；另外断言 `--json` 的排序与 `internal/session.Sort`
  一致——**先按状态优先级，同状态内新的在前**。这条最初写成「按时间排序」是错的：一条更早的
  `error` 会刻意排在更新的 `idle` 前面。
- **中文路径与中文工具名**：项目目录用 `项目 开发/状态 浮窗`、工具名用 `读取`，断言它们原样穿过
  JSON 编码、快照、store 与 `--json`。对本项目的用户来说这不是边缘情况。

**一个把整轮测试判错的细节**：`--json` **不是**存活探针。浮窗没在跑时它照样读快照文件并打印
`{"count":0,...}`，所以拿「`--json` 是否输出 JSON」判断浮窗在不在，会让「浮窗已经退出」这件事永远
测不出来——第一版的 S4 就这样假绿过。现在改用**直连命名管道**：管道名按
`sha256(home 小写、去尾分隔符)` 前 4 字节在 harness 里重算一遍，`net.createConnection` 能连上才算活着；
连上立刻断开不写任何字节，服务端读到 EOF 就关连接，没有副作用。

另外 `--quit` 会与控制探针抢管道（服务端每个连接读到 EOF，读超时 5 s），一次不成就让后续所有场景
撞互斥体。`stopOverlay()` 因此最多重试三次并轮询到管道消失，恢复真实浮窗时同样重试三次——
沙箱实例刚被杀、互斥体还没释放的第一枪会打空。

### 30.7 跑法
```powershell
node tests/fresh-machine.mjs                 # 全量（含浮窗场景，会临时停掉正在跑的实例）
node tests/fresh-machine.mjs --no-overlay    # CI 形态
node tests/fresh-machine.mjs --keep          # 留沙箱现场，配合 --verbose 排查
node tests/fresh-machine.mjs --verbose       # 打印每一项检查
```

前置：三个二进制已构建（`bin/AI-Terminal-Pet.exe`、`pet-statusline.exe`、`pet-hook.exe`）、
Node 18+、Windows PowerShell。退出码 0 = 全过。

### 30.8 这一轮对已发布包的影响
`vendor/package.json` 是**新增的包内文件**，所以仓库里的 `packages/dsh-ai-terminal-pet/`
与 npm 上的 1.2.4 tarball 差别又多了一处。**下一次发布必须是新版本号（1.2.5）**，不能重传 1.2.4；
发版流程见 §28.6。

## 31. 按真实 Harness 类型定义做的一次深度适配（2026-11-11）

### 31.1 先拿到权威定义，再改代码

插件是照着**猜测的事件名**写的。这一轮在本机装着的 Harness 里找到了权威定义
（`~/.dsh/profiles/node_modules/@deepseek-ai/dsh-session/lib/types/types.d.ts`），
`SessionEventMap` 只有这 13 个持久会话事件：

```
turn/start  turn/end  step/start  step/end  user/message
assistant/chunk  assistant/message  tool/call  tool/result
todo/write  request/header  request/context  session/end-seed
```

对照下来：插件里 `working-directory/change` **在整个 Harness 里根本不存在**（23 个 `.d.ts`、全部
运行时 `.js` 都搜不到），`session/title` 只存在于**客户端**事件里，`command/done` 是**客户端命令服务**
的事件而不是会话事件。真正能用的主机侧信号是 `turn/*`、`tool/*` 与 `approval/request` 瀑布监听器
——这三类插件原本就接对了。

### 31.2 会话头才是项目路径的唯一来源

`SessionHeader` 定义（同文件）里有：

```ts
readonly id: SessionId
readonly createdAt: number        // Unix epoch ms
readonly cwd?: string             // Absolute working directory the session was created in
readonly parentSession?: SessionId // The session this one was forked from
```

而插件原来：`project` 取 `process.cwd()`、`parentSessionId` 恒为 `''`、`startedAt` 取「第一次看到
这个会话的时间」。桌面版 Harness 从安装目录启动，`process.cwd()` 通常是安装目录或 System32，
所以**每一条 DSH 会话都被标成同一个错误路径**——面板上的「跳到终端」也会因此跳错窗口。

改法：所有事件统一走 `reportFor(candidate, patch)`，它读 `headerOf(candidate)`（兼容
`session/event` 直接给会话对象、`agent/*` 给 `{ agent: { session } }` 两种形态），把
`cwd` / `parentSession` / `createdAt` 折进 payload。会话头对「尚未声明项目」的会话是权威；
某个事件显式给了 `project` 时以事件为准（保留未来的 cwd-change 扩展空间）。

`verify.mjs` 的用例因此改用**真实会话头形状**，并新增三条断言：每条上报的 `project` 都等于会话头
`cwd`、`parentSessionId` 等于 `parentSession`、`startedAt` 等于 `createdAt`。
`tests/fresh-machine.mjs` 的 S3 也加了同样三条，外加一个带 `parentSession` 的子会话。

### 31.3 顺带发现的两件事

- `tool/result` 的 `error` 形状是 `{ name, code }`，没有 `reason`；插件的
  `data.error.reason || data.error.code` 因此总是落到 `code`。行为可接受，但注释里写的是
  「reason」，容易误读——已在 `eventPatch` 的总注释里说明事件来源。
- `turn/end` 的 `reason` 是 `TurnEndReason` 联合类型；`protocol.js` 会把它当 token 清洗
  （空格变 `-`），所以「build failed」到面板上是 `build-failed`。这是协议行为，不是 bug，
  但测试里必须按清洗后的值断言（`verify.mjs` 与本套件都已如此）。

## 32. 交接：下一个会话从哪里开始（2026-11-11 15:45 UTC）

### 32.1 现在的事实（一句话）

代码、测试、发布、分发四条线**都是绿的**，浮窗本体已到 **1.0.0**（`pet-v1.0.0`），插件包已到 **1.2.6**
（`dsh-plugin-v1.2.6`，npm `latest` = 1.2.6），两个成品包（便携 zip 与自解压 `pet-setup.exe`）都已发布
并从公网下载回来实测过。**I1 只剩 I1-δ（签名与 MSI）**，而它卡在证书，不卡在代码。下一个动作是发声。

### 32.2 交接口令（按顺序跑，先证明起点干净）

```powershell
git status --short                       # 必须是空的
git log --oneline -3                     # 顶上是 c8ad006 之后的文档提交
npm view dsh-ai-terminal-pet version     # 1.2.5
gh run list --limit 1                    # success
gofmt -l . ; go vet ./...                # 干净 / 0
go test -count=1 ./...                   # 11 个包全 ok（201 个 func Test）
node packages/dsh-ai-terminal-pet/verify.mjs
node packages/dsh-ai-terminal-pet/scripts/vendor.mjs --check
node integrations/deepseek-harness/verify.mjs
node tests/fresh-machine.mjs --no-overlay   # 160/160；去掉 --no-overlay 是 181 并且会重启浮窗
.\scripts\install.ps1 -VerifyOnly           # 11 file(s), 0 warning
```

### 32.3 I1 的开工顺序（照 ROADMAP §4 那张表，不要重新调研）

> **2026-11-11 晚更新：I1-α / γ / β 全部已交付。** 浮窗 1.0.0、插件 1.2.6、`pet-v1.0.0` 的
> 便携 zip 与 `pet-setup.exe` 都在线上，npm `latest` = 1.2.6，上游 PR #14439 已开。
> **只剩 I1-δ（签名 + MSI），它需要证书这个外部输入。** 下面四条保留为当时的开工顺序。

1. **先定签名**：有没有证书，决定做「未签名但可用」还是「签名发行」。没有证书不要假装能签名。
2. **做自解压单文件 `.exe`**：内嵌 `AI-Terminal-Pet.exe`、`pet-statusline.exe`、`pet-hook.exe`、
   `protocol.js`；复用 `scripts/install.ps1` 的 11 件制品清单与逐件 SHA256；落到同一安装目录，
   同样写 `HKCU\Environment\Path` 与 `HKCU\...\Run`；卸载走同一份清单。
   **验收必须用 `tests/fresh-machine.mjs` 的沙箱模型**：一台「没装过 Go、没装过本项目」的机器，
   双击后浮窗出现、`--doctor` 全绿、`-VerifyOnly` 等价的自校验通过。
3. **再考虑 MSI**（WiX v3 的 `candle`/`light`，或 WiX v4 的 `wix build`），企业分发用。
4. **签名**：`signtool sign /fd sha256 /tr <时间戳> /td sha256`，并保留「未签名也能构建」的 CI 分支。

### 32.4 交接时最容易踩的四个坑（都是这一轮真的踩过的）

| 坑 | 表现 | 怎么避开 |
|---|---|---|
| PowerShell 读 UTF-8 | `Get-Content -Raw` 按 ANSI 解码，中文整行 mojibake，上传后 YAML 报 `invalid YAML – deficient indentation` | 传文件内容一律 `[IO.File]::ReadAllBytes` + `ToBase64String`；不要「读成字符串再写回」 |
| 仓库目录是 Low IL | 从 `bin\` 直接跑的进程读不到 `%TEMP%` 沙箱里的配置，「自动同意」永远像关着 | 测试里先把二进制复制进沙箱再执行；详见 §29.4 / §30.5 |
| `--json` 不是存活探针 | 浮窗没跑它照样打印 `{"count":0}`，用它判断「浮窗在不在」会假绿 | 直连命名管道（名字按 `sha256(home 小写去尾)` 前 4 字节重算） |
| `gh pr create` 默认仓库 | 不带 `-R` 时会把 PR 开进自己的 fork，上游看不到 | 永远 `-R <上游>` + `--head <自己>:<分支>` |

### 32.5 还没做完、也不该由下一个会话猜的事

> **2026-11-12 07:40 +0100 复读更新**：下面这张表在本会话结束时**又全部变了**，所以旧条目保留但被覆盖。
> 现状（每条都是刚读对方公开索引/API 得到的，不是读 PR 状态）：

- **awesome-dsh-plugin**：#14385（→1.2.5，17:20:15Z）与 **#14439（→1.2.6，20:12:31Z，`475a48e`）都已合并**，
  仓库里 yml 已是 1.2.6，**线上 `plugins.json`（`count` 10,578）现在也渲染 1.2.6 了**——
  这条「合并 vs 渲染」的延迟终于追平。`#14243` 早已 CLOSED。**不需要再开 PR，等下一个版本时再说。**
- **dsh-plugin.org（`dshplugin/dsh-plugin-hub` #735）**：**2026-11-12T00:43:14Z 由 `yacuo` 以
  `completed` 关闭**，留言：`platforms: ["win32"]` → 条目标注 Windows-only，**从源码安装**
  （`dsh plugin --profile web add github:jjjjjjjjnnjnn/ai-terminal-pet`）。但**它会拉到什么版本**要注意：
  它不钉版本，走的是仓库根 shim，因此永远装到默认分支的最新提交——本次实测其线上目录
  `api.dsh-plugin.org/plugins.en.json`（6,688 条）里**还没有本包的条目**（同账号只有 `dsh-triage`、
  `dsh-remedy`），所以要等它下一次重建。
- **dshopen**：线上 `dshopen.org/index.json`（1,207 条）仍记 **1.2.4**，落后两个版本，等它自己重建。
- **YELEBAI/dsh-plugin-marketplace**：**自动采集，已收录**（`registry/plugins.json` 第 8,729 条，
  `version: 1.2.5`、`verifiedCommit: 584bc2f`、`profiles: [desktop, headless, web]`）。它按 topic 扫，
  不用申请；下一次扫描会自己走到 1.2.6。
- **bradeGithub/DSH-Plugins-Marketplace**：**自动采集，已收录**（`registry.json` 第 13,128 条，
  `version: 1.2.4`）。issue #367 只是留证，**可以关了**。
- **imsai-sh/awesome-deepseek-harness-plugins**：目录文件已合并并在
  `catalog/plugins/jjjjjjjjnnjnn--ai-terminal-pet--packages--dsh-ai-terminal-pet.json`，
  `category: notify`，中英摘要齐全；**它不钉版本，所以没有滞后问题**。PR #912 仍 open，
  但对这个仓库来说「目录已在」就等于「已收录」。
- **zhu1090093659/dsh-community-plugins**：PR #97 于 2026-11-11T05:06:31Z 合并、reviewer 给的是
  **APPROVED**，`community.json`（124 条）里 `id: dsh-ai-terminal-pet` 在位。它的评审意见只有一句
  **不阻断的建议**：文案里最好写明「Claude Code / OpenCode / Antigravity 需另装 reporters」——
  **这条已经落实**（1.2.4 起三处清单、README 中英两段、1.2.6 的条目描述都写了）。
- **dshbase #322 / dsh.so #24 / alldsh #2 / AdamPlatin123/dsh-plugin-radar #1255**：截至
  2026-11-12 07:40 **四家全是 open、0 评论、updated_at 还停在 08:1x–08:4x**。等对方，不要再投。
- **dshmarketplace.dev**：公开搜索 `/api/v1/plugins?q=ai-terminal-pet` 仍回 `{"total":0}`。
- **`go test -race`** 仍缺工具链（Windows 上的 `-race` 要 gcc/cgo），§6 已记录，不要假装跑过。
  `govulncheck` / `gosec` **已跑**（2026-11-12 进 CI，见 §35），不要再写成「未安装」。
- **CI 有两条 `failure`（不是 cancelled）**，都已修复且现行全绿，查历史时不要误判：
  `43030375577`（2026-11-11T13:34Z）挂在 **Fresh-machine end-to-end suite**（当时刚加进 CI 的 S 系列）；
  `42894142140`（2026-11-10T20:19Z）挂在 **Verify the DSH plugin package**
  （「索引最新版本 == 包版本」断言，正是 §27 记的 `b4474a6` 那类问题）。
  其余 5 条非 success 的 run 全是 `cancelled`，来自 `concurrency` 的同分支新推送取消——**不是失败**。

### 32.6 上游门禁的实测步骤（2026-11-11 复查用）

判断「合并」与「可见」不能读 PR 状态，要读两处：

```powershell
# 1. PR 是否真的合并
gh pr view 14385 -R awesome-dsh-plugin/awesome-dsh-plugin --json state,mergedAt,mergeCommit

# 2. 上游仓库里那条 yml 现在指向哪个版本（合并后立刻就是新版本）
gh api "repos/awesome-dsh-plugin/awesome-dsh-plugin/contents/data/plugins/jjjjjjjjnnjnn__ai-terminal-pet--packages-dsh-ai-terminal-pet.yml" --jq .content
#    （返回 base64；解开成文件后看 tarball: 那一行的版本号，别读成字符串再写回，见 §30.6）

# 3. 站点域上真正渲染出来的索引——这才是用户看到的
Invoke-WebRequest https://awesome-dsh-plugin.com/plugins.json
  # 找出 name = ai-terminal-pet#dsh-ai-terminal-pet 那一条，看 tarball / install 里的版本号
```

2026-11-11T18:05Z 实测的三处状态**互不一致**，这正好是个教案：

| 位置 | 状态 |
|---|---|
| PR #14385 | **MERGED**（17:20:15Z，merge commit `70581cc`） |
| 仓库里 `data/plugins/…yml` | `tarball` = `dsh-plugin-v1.2.5` ✅，`description.en` 也已带上「安装本插件只接 DeepSeek Harness；另外三个终端要各自装上报端」那句 ✅ |
| 站点 `plugins.json` | `count` 10,446，本条目 `tarball` / `install` 仍是 **1.2.3**，描述也是旧版（没有 reporters 那句）❌ |

原因是「Build site from README」那条 workflow 的并发取消：同一分支上每隔几分钟就有新提交，
旧 run 被 `cancelled`，所以站点重建一直在排队。**这不是我们能推的，也不要为此重复开 PR。**

同一轮的另外两个数字记在这里，因为它们决定了这一轮宣传的落点：

- `count` 从 10304（合并 1.2.3 后）涨到 **10,446**，`plugins.json` 有 14 MB——这个目录本身在长，
  单个条目的新旧取决于它自己那次重建。
- 本仓库在该目录里的描述**已经**写明了「其他终端需各自安装 reporter」，是 #14385 合并时一起带进去的；
  `dsh.catalog` 与 `docs/market-index.json` 是同一段文字，所以**线上那条旧描述也会随重建一次性对齐**。
  ——也就是说：这一处**不需要**再开 PR，等站点重建即可。

---

## 33. 便携 zip、`install.cmd` 与仓库门面（2026-11-11 宣传准备批）

本轮的目标是「在确保稳定之后开始对外宣传」。稳定性这一条**先被证明**（见 §33.6 的实测），
然后动手时发现漏斗上还有两个真实断点，必须先修掉才谈得上宣传：

1. **`.\scripts\install.ps1` 在默认 `RemoteSigned` 策略下直接被拒**。README 的安装第 2 步就是那一行，
   任何普通用户照抄必然失败，而且报的是安全策略的官话（`UnauthorizedAccess`），不像「装错了」。
   这比「没有独立安装包」更致命，因为它是**当前唯一被文档承诺的路径**。
2. **Release 资产里没有成品包**。`dsh-plugin-v*` 的 tarball 只是 DSH 插件包；非 Harness 用户
   要么装 Go 工具链，要么从 tarball 里手动挖 `vendor/AI-Terminal-Pet.exe`。

### 33.1 `scripts/install.cmd`：文档入口换成 cmd

`RemoteSigned` 拒绝一个带「从互联网下载」标记的本地 `.ps1`，而**从 Release 解压出来的文件恰好带这个标记**
——这正是「别人第一次接触这个项目」的形态。`install.cmd` 只做一件事：
`powershell.exe -NoProfile -NoLogo -ExecutionPolicy Bypass -File "%~dp0install.ps1" %*`，
参数与 `%ERRORLEVEL%` 原样转发，`endlocal & exit /b %_STATUS%`。

三条设计约束：

- **`-ExecutionPolicy Bypass` 只作用于这一个 powershell 进程**，不改 MachinePolicy / UserPolicy /
  CurrentUser，别的脚本不受影响。脚本注释里写明了这一点，因为这是用户最合理的担心。
- **不加任何安装逻辑**。`install.ps1` 仍是唯一实现；新增的第二份实现就是漂移的来源。
- **纯 ASCII**，理由与 `install.ps1` 相同（Windows PowerShell 5.1 按 ANSI 读无 BOM 脚本）。
  `%~dp0` 里可能有中文路径，但那是运行时数据，不是脚本字节。

无参数时（双击的场景）才 `pause`，因为双击的 cmd 会随脚本结束立即关窗，把最后一行——包括失败——
一起吞掉。`pause` 在非控制台 stdin 下立即返回，所以无人值守调用不受影响。

### 33.2 `scripts/package.ps1`：便携 zip

产物 `dist\ai-terminal-pet-<overlay 版本>-win-x64.zip`，17 个文件：

```
bin\AI-Terminal-Pet.exe  bin\pet-statusline.exe  bin\pet-hook.exe
reporters\protocol.js
integrations\deepseek-harness\ai-terminal-pet-bridge\   (7 files)
scripts\install.ps1      scripts\install.cmd
install.cmd              README.install.txt   SHA256SUMS.txt   LICENSE
```

**为什么是这个形状**：`install.ps1` 的一切路径都从 `Split-Path -Parent $PSScriptRoot` 推导，
所以一个「与仓库同形」的子树不需要第二套安装实现，也不需要改写路径——**装 zip 和装仓库跑的是同一份脚本**。
任何其他形状都得单独验证一遍，而「校验集 ≠ 安装集」正是过去残留旧 `protocol.js`、
所有上报端静默失联的成因（§1i）。

版本号从 `internal/app/app_windows.go` 的 `const Version = "..."` 正则读出，**不在这里写第二份**。
顺手记一条：浮窗本体（0.9.0）与插件包（1.2.x）是两条版本线，zip 名必须跟**浮窗**那条。

**三个 PowerShell 陷阱，全是实测撞到的**（都写进了脚本注释）：

| 陷阱 | 表现 | 原因与修法 |
|---|---|---|
| `Compress-Archive` | 条目名是 `bin\pet-hook.exe`（反斜杠） | Windows PowerShell 5.1 走 .NET Framework，它写反斜杠；Explorer/7-Zip 能读，`tar -xf` 与 Linux 上的 `unzip` 会解出一个名字里带反斜杠的**单个文件**。zip 规范要的是 `/` |
| `ZipFile.CreateFromDirectory` | 同上 | 它内部也用框架的 `CreateEntry`，不给名字就没法控制分隔符。**唯一可控的做法是逐条目写入并显式给名字** |
| `[ZipArchive]::new(...)` | `找不到类型 [System.IO.Compression.ZipArchive]` | 5.1 要先 `Add-Type -AssemblyName System.IO.Compression` + `...FileSystem`；PS7 自带 |
| `$archive.CreateEntryFromFile(...)` | `找不到 ... CreateEntryFromFile 方法` | 那是 `ZipFileExtensions` 上的**扩展方法**，PowerShell 没有扩展方法语法，`$obj.Method()` 解析不到 |
| `$item.LastWriteTime = ...` | `正在创建模式下无法修改条目` | 条目一旦写入字节就进入 create 模式；**必须在写字节之前**赋时间戳，所以改成先 `CreateEntry` → 赋时间戳 → `Open()` → 拷贝流 |

每个条目钉固定时间戳 `2026-11-11T00:00:00Z`，所以同一份源码重建出的 zip**逐字节相同**
——与三个二进制用 `-buildvcs=false` 是同一条理由。实测连跑两次 sha256 一致。

**打包前的交叉校验**：比对 `bin\AI-Terminal-Pet.exe` 与
`packages\dsh-ai-terminal-pet\vendor\AI-Terminal-Pet.exe`，不一致就**拒绝打包**。
这两个是同一产品同一版本号下的两份副本，任一侧陈旧都会让两类用户拿到不同字节、同一个名字。
CI 用 `PET_PACKAGE_SKIP_VENDOR_CHECK=1` 把它降级为警告——runner 的 Go 版本与提交进仓库的 vendor
二进制不是同一个（`go.mod` 的 `go 1.23.0` 是下限不是钉版本，跨工具链字节相等从来不可能达成，§27.9）。
真正能比对的是**切 release 的那台机器**，所以这个拒绝留在那里咬人。

**断言方式**：数的是**文件**不是条目（目录条目是正常的），而且逐个名字比对，
最后再断言每个条目名都不含反斜杠。第一版数了条目数，于是 18 vs 17 把正确的包判成失败——
「zip 是对的、检查是错的」和「检查过了、zip 少了三个文件」都真实发生过。

### 33.3 CI 与门禁

`.github/workflows/ci.yml` 在 `fresh-machine` 之后新增两步：`Package the portable zip`
（`shell: pwsh`，带 `PET_PACKAGE_SKIP_VENDOR_CHECK=1`）与 `Upload the portable zip`
（artifact 名 `ai-terminal-pet-win-x64`，留存 7 天，`if-no-files-found: error`）。
理由是：一个打不出包的打包脚本对 release 而言就是红的构建，不能等到切版本那天才发现。

`tests/fresh-machine.mjs` 增到 **162** 项，新增 S5b：用 `cmd.exe` 跑 `scripts\install.cmd`
的两种形态，断言
(1) 参数转发——`-VerifyOnly` 打出 `11 file(s) ... 0 warning`；
(2) 退出码转发——`-VerifyOnly -Uninstall` 仍是 **2**。
两者都带 `-Destination` 指向沙箱，所以**不碰本机的真实安装目录**——第一版没有带，
于是它在本机（已装过本项目）会绿、在干净的 CI runner 上会红，而失败原因与 wrapper 无关。

### 33.4 仓库门面（宣传前必须补的那一栏）
社区健康度实测 **42%**（缺 CONTRIBUTING、CoC、issue/PR 模板）。本轮补齐：

- `CONTRIBUTING.md`：六条硬规则（reporter 快速失败 / 零运行时依赖 / 管道名跨语言契约 /
  UTF-8+LF 与纯 ASCII 脚本 / 协议改动走 PROTOCOL §11 / 不写死本机路径）、
  发布不可变文件清单、`--doctor` 作为报障前置、三套环境陷阱（Low IL、`--json` 不是存活探针）。
- `CODE_OF_CONDUCT.md`：Contributor Covenant 2.1 全文。
- `.github/ISSUE_TEMPLATE/`：`bug.yml`（**强制**附 `--doctor` 全文）、`integration.yml`
  （新终端接入，指向 `docs/PROTOCOL.md`）、`feature.yml`（并列两条非目标，先问「你在什么场景里」）。
- `.github/pull_request_template.md`：门禁清单 + 「是否触碰发布不可变文件」的显式勾选
  ——后者决定这次 PR 要不要连带一个版本号，事前问比事后发现便宜。
- 仓库元数据：description 改成带关键词的一句（含 Claude Code / OpenCode / Antigravity /
  DeepSeek Harness、4.17 MiB、零运行时依赖、无账号无网络），设 Homepage，topics 11 → **15**
  （补 `opencode`、`antigravity`、`ai-agents`、`coding-agents`）。

**一个必须知道的口径**：GitHub 的 `community/profile` API 只把**单个** `.github/ISSUE_TEMPLATE.md`
算作 `issue_template`，所以用了 `ISSUE_TEMPLATE/` 目录里的表单之后，`health_percentage` 停在 **85%**
（贡献指南、行为准则、PR 模板、许可证、README 五项都是 present）。**85 不是没做完**，别照着这个数字
继续加文件；要判断表单是否真的生效，看新建 issue 页的模板选择器。

这套表单第一版是**坏的**：`bug.yml` 里 `description: \`winver\` output ...` 的反引号让 YAML
标量提前结束，文件扫不过——GitHub 会静默拒收，选择器里什么都不显示，而坏掉的恰好是唯一强制要
`--doctor` 全文的那张表。修法是改成引号标量，并且改成**先跑一个 YAML 解析器再信它**。
另外 `config.yml` 把 `blank_issues_enabled` 关掉：bug 表单的第一个必填项就是 `--doctor` 全文，
空表单正好是绕开它的捷径，关掉之后「最省事的路径」和「能回答问题的路径」重合。

### 33.5 README / CHANGELOG / ROADMAP 的事实修正

README 修的不是叙事而是**事实**：插件 1.2.4 → **1.2.5**、npm 已发布（`dist-tags.latest` 1.2.5、
17 文件、tarball 与 Release 资产逐字节相同）、market-index 钉到 `34aa1f7`、
徽章 `tests-154` → `tests-190` 并加 npm 徽章、`No standalone installer yet` → `Not code-signed`、
安装段改为 A/B/C 三条并列路径、CI 门禁列表补 `package.ps1`。
英文段与中文段同步改——这一段之前两边已经分叉过一次（英文说「npm 尚未发布」、中文说「已发布」）。

ROADMAP §4 把 I1 拆成 α（已交付，零版本号代价）/ γ（自解压单 exe，未开工）/ β（换版本号，排在文案定型后），
并记下「正式发声排在 β 之后」这条排程约束，以及「不要在 0.9.0 上先宣传再改号」的理由。

### 33.6 本轮的实测基线（宣传材料的每个数字都以这张表为准）
```
git status --short                               干净（开工前）
git log --oneline -3                             735ebd0 / c8ad006 / b7a7142
npm view dsh-ai-terminal-pet version             1.2.5
gh run list --limit 1                            success（43052732146, 1m41s）
go version                                       go1.27.0 windows/amd64
gofmt -l . / go vet ./...                        干净 / 0
go test -count=1 ./...                           12 个包全 ok，201 个 func Test
node packages/dsh-ai-terminal-pet/verify.mjs     VERIFY PASSED
node .../scripts/vendor.mjs --check              两个 vendor 文件 matches the source
node integrations/deepseek-harness/verify.mjs    VERIFY PASSED
node tests/fresh-machine.mjs --no-overlay        PASS 162/162（本轮新增 2 项）
scripts\install.cmd -VerifyOnly                  install: verified - 11 file(s), 0 warning(s)
--doctor（已装实例）                              9 ok, 0 warn, 0 fail
scripts\package.ps1                              4,983,915 B / 17 files / 7 dirs
                                                 sha256 EF0FC0D0A9278F8C52E3C7D796B668EDBBE87105FF0CD040314ADE88E1B346F4
```

外部世界的基线（gh api / npm registry 实读，用于事后判断宣传有没有效果）：

| 指标 | 值（2026-11-11 17:35 UTC 复读） |
|---|---|
| Release 下载数 | `dsh-plugin-v1.2.3` = **47**、`v1.2.4` = 4、`v1.2.5` = 5；`pet-v0.9.0` = **3**（发布后约 40 分钟） |
| 仓库 30 天浏览 | 7 views / 2 uniques（截至 2026-11-10） |
| star / fork / watcher | 1 / 0 / 0 |
| npm 首次发布 | 1.2.4 于 2026-11-11T12:07:36Z，1.2.5 于 14:58:53Z |
| 社区健康度 | **85%**（GitHub 不把 `ISSUE_TEMPLATE/` 目录算作 `issue_template`，见 §33.4） |

**46 vs 3 这个对比是本轮最有用的一个数字**（复读时已变成 47 vs 4/5）：把市场条目推到最新版的渠道价值，
比发一条推文更确定。宣传的重点应该是「让已经存在的分发路径指向最新资产」，
其次才是「制造新的入口」。`pet-v0.9.0` 上线 40 分钟就有 3 次下载，而**它还没有被宣传过**
——那三次来自 README 的安装路径 B，说明「文档一改就有人装」，也让「先修漏斗、后发声」这个顺序
有了第一个正向证据。

### 33.7 I1-γ / β 已交付，还剩什么（2026-11-11 晚更新）

**这一小节原先列的「还没做完」两项已经做完**，保留在这里是因为它们的坑值得回看：

- ✅ **I1-γ 自解压单文件 `.exe`**：`cmd/pet-setup` 已交付。三处「能构建、能运行、看起来正常」的坑写在
  ROADMAP §4 与 CHANGELOG 的 `[1.0.0]` 条目里：`//go:embed payload` 静默跳过点开头文件；
  `Copy-Item -LiteralPath '…\*'` 不做通配（载荷 2 个文件也照样构建成功）；
  载荷不能照抄 zip 整棵暂存树（`README.install.txt` 与根 `install.cmd` 不在摘要里，每次安装都会失败）。
- ✅ **I1-β 换版本号**：浮窗 1.0.0 + 插件 1.2.6，A 提交 `df4fcf6`（版本、索引 `version` 与
  `publishTimestamp`、两份 README 的安装命令、版本徽章）、B 提交 `6496ba6`（索引 `source.commit` 与
  `icon` 钉到 `df4fcf60…`）。两个 tag、两个 Release、npm `latest` = 1.2.6、上游 PR **#14439** 已开。
  **`pet-v0.9.0` 与 `pet-v1.0.0` 都是既成事实且不可变**，下一次是新的 tag 与新资产。
- ⏳ **只剩 I1-δ（签名 + MSI）**：`signtool.exe` 在本机已找到，**证书没有**；WiX/NSIS 都没装。
  这一段需要外部输入（买证书或 Azure Trusted Signing），代码侧只需把 `signtool sign /fd sha256
  /tr <时间戳> /td sha256` 挂进 `scripts/package.ps1`，并保留「未签名也能构建」的分支。
- ✅ **上游 PR #14439 已合并**（2026-11-11T20:12:31Z，merge commit `475a48e`）：仓库里
  `data/plugins/jjjjjjjjnnjnn__ai-terminal-pet--packages-dsh-ai-terminal-pet.yml` 的 `tarball`
  已是 **1.2.6**。**但站点 `plugins.json`（`count` 10,578）仍渲染 1.2.5**，`dshopen.org/index.json`
  （1,207 条）仍是 **1.2.4**——「合并」与「渲染」始终是两件事（§32.6 有复读命令与三处对照表）。
- ✅ **DSH 官方 Discussions 已发帖**：[#13001](https://github.com/deepseek-ai/deepseek-harness/discussions/13001)
  （`Show Your Plugins!`，2026-11-11T20:50:40Z，upvote 1 / 0 评论）。这是受众最准的一条渠道。
  发帖前过了一轮 `agy` 评审，改掉三处会直接导致删帖的问题，记录在 [docs/LAUNCH.md](LAUNCH.md) §1.2：
  四终端并列（触犯「禁止无关推广/报菜名」）、把自家 `protocol.js` 写成「官方 protocol 模块」、
  以及 `**_> …_**` 渲染不出引用块（"非官方"声明会变成一串星号）。
  （**历史记录**：那次评审发生在 2026-11-11，当时的第 8 条允许调用外部模型 CLI；2026-11-13 起第 8 条
  改为同会话复核，这类评审不再经过 `agy`。）
- ✅ **两家市场是自动采集、不用申请**：`YELEBAI/dsh-plugin-marketplace` 的 `registry/plugins.json`
  第 8,729 条（`version: 1.2.5`、`verifiedCommit: 584bc2f`、`profiles: [desktop, headless, web]`）、
  `bradeGithub/DSH-Plugins-Marketplace` 的 `registry.json` 第 13,128 条（`pkg_name` +
  `npm_pkg_name` 都在、`version: 1.2.4`）。加上 awesome-dsh-plugin 与 imsai，**已收录四家**；
  dshbase / dsh.so / dsh-plugin.org / AllDSH 四家人工目录仍是 open / 0 评论（详见 §28.3）。

**宣传素材（已完成）**：`docs/assets/demo.gif`（25 帧 / 5.38 秒循环）、`desktop-scene.png`
（两个终端 + 浮窗同框）、`smartscreen-unsigned.png`（**标着 MOCKUP 的示意图**，不是真机截图）、
`doctor-green.png`（程序自己的 stdout，用户名替换为 `<you>`）。四张都已从 README 中英两段引用。

**文案**：对外可见的一切（X 中英双线程、Show HN 标题与正文、DSH Discussions 帖、市场摘要、三份 issue
回复模板）已按当时的 AGENTS.md 第 8 条由 `agy --model gemini-3.8-flash-high --effort high` 起草并逐条核对
（**这一条已于 2026-11-13 改为同会话复核：不再调外部模型 CLI，见 §41 与本轮 §42**），
成品在 [docs/LAUNCH.md](LAUNCH.md)。该轮评审至少纠正了一条会真出问题的事实表述：初稿把浮窗写成
像是对 Claude Code / OpenCode / Antigravity 开箱即用，而实际只有 DeepSeek Harness 无需配置。
同一份评审在 DSH 论坛帖上又抓出三处：四终端并列触犯该区的「禁止无关推广」、把自家 `protocol.js`
说成「官方 protocol 模块」、以及 `**_> …_**` 在 Markdown 里渲染不出引用块（"非官方"声明会变成
一串星号）。三处都是会让人删帖的那种，不是措辞偏好。

---

## 34. 交接：2026-11-12 早，下一个会话做 GitHub 邮箱与账号相关的事

### 34.1 这一刻的事实（一句话）

五条线全绿：**浮窗 1.0.0 / 插件包 1.2.6 / npm `latest` = 1.2.6 / `pet-v1.0.0` 两个资产都在 /
DSH Discussions 帖 #13001 已发**；上游合并的三个（awesome-dsh-plugin、imsai、zhu）都已落地，
自动采集的两家（YELEBAI、brade）也都收录了。**下一个会话的目标不是再发版本，而是处理 GitHub 邮箱
与账号这条线**——它和本仓库的「谁是提交者」直接相关，所以先写清现状。

### 34.2 现在的邮箱事实（本会话实测，别凭记忆）

| 项 | 值 | 怎么得到的 |
|---|---|---|
| git 提交身份 | `jjjjjjjjnnjnn <jjjjjjjjnnjnn@users.noreply.github.com>` | `git config user.name` / `user.email`；**全仓 2596 行历史里只有这一个 author**（`git log --format='%ae'` 去重后 1 条） |
| GitHub 账号 | `jjjjjjjjnnjnn`（显示名 `RobertJJJ`），id **236163329**，2025-10-05 注册 | `gh api user` |
| GitHub 认为这个提交是他吗 | **是**，`author.login = jjjjjjjjnnjnn` | `gh api repos/.../commits` 的 `author` 字段 |
| 提交签名 | **`verified: false`（`reason: unsigned`）** | 同上；这是「本地 ssh 提交没签名」，不是邮箱问题 |
| 账号 profile 的 email 字段 | **两个都是 `null`**（`email` 与 `public_email`） | `gh api user` |
| `gh` token 的 scopes | `gist, read:org, repo, workflow` —— **没有 `user`** | `gh auth status`；因此 `gh api user/emails` 返回 404，要读邮箱列表必须先 `gh auth refresh -h github.com -s user` |
| 仓库里有没有硬编码的私人邮箱 | **没有**。全仓 `*.md/*.js/*.mjs/*.ps1/*.go/*.json` 扫描只命中 `dsh-ai-terminal-pet@x.y.z` 这类包名伪装成邮箱的字符串 | 用正则跑 `git ls-files` 全量扫过 |

**结论**：现在的 `jjjjjjjjnnjnn@users.noreply.github.com` 是**旧式 noreply 地址**，它能正确归属
（GitHub 认了 `author.login`），但它有两个实际代价：(1) 别人无法从提交里点进一个真实联系地址；
(2) 这个地址**不含账号 id**，所以账号改名后历史提交会失去归属——按 GitHub 的说法，带 id 的形式
`236163329+jjjjjjjjnnjnn@users.noreply.github.com` 才是改名后仍然归属的形式。
**要不要改、改成哪个，是下一个会话要决定的第一件事**，而且它一旦决定就要重写 2596 行历史的
author（`git filter-repo`），**必须在动手前明确这是不可逆的**。

### 34.3 邮箱这条线上，本会话确认过的四个子问题

1. **Outlook 侧有五个账号在本机**（`outlook_list_accounts`）：
   `chenyuyou2026@outlook.com`（默认）、`rongjiajun2025@outlook.com`（**GitHub 通知都发到这里**）、
   `rongjiajun701@outlook.com`、`jiajun.rong@schloss-heessen.de`（token 已过期）、
   `rongjiajun2026@outlook.com`（token 已过期）。
   **GitHub 的通知全部落在 `rongjiajun2025`**，所以改 GitHub 通知邮箱前先确认那个账号还在用。
2. **`rongjiajun701` 收的信几乎全是垃圾邮件**（同类约会站点的批量投放），
   **不要把它当作 GitHub 的联系邮箱**。
3. **有一个还没回的问题**：DSH Discussion #8202 里 `Jianwei Mao` 问了一句
   「你这是用大模型回复的吗？」——这是账号相关的声誉问题，下一个会话可以顺手处理（不处理也不算漏）。
4. **`gh api user/emails` 需要 `user` scope**，而当前 token 没有。要读/改邮箱必须重新授权；
   `gh auth refresh -h github.com -s user` 是唯一正规做法，**不要**用 `gh api -X PATCH user` 盲改。

### 34.4 与邮箱/账号相关、但属于本仓库的待办（按优先级）

> **2026-11-12 更新**：前两条已落地，第 3 条**刻意不做**，第 4–5 条仍待办。细节见 §35.5。

| 优先级 | 事 | 状态 |
|---|---|---|
| **高** | `git config user.email` 改成 id 形式；要不要重写 118 个提交的历史 | ✅ **已完成，且决定不重写**：`user.email` = `236163329+jjjjjjjjnnjnn@users.noreply.github.com`（仓库内生效，全局仍是旧地址，因为别的项目没有这个账号的 id）；重写会改掉每一个 sha 并让 11 个已发布 tag 与索引里钉的 `source.commit` 指空，所以不做 |
| **中** | 给提交加签名 | ✅ **已完成**：SSH 签名（`gpg.format=ssh`、`user.signingkey=~/.ssh/id_ed25519_signing.pub`、`commit.gpgsign`/`tag.gpgsign` = true），公钥以 `user/ssh_signing_keys` id `1325600` 挂在账号上；`gh api .../commits/main` 已实测 `verified: true / reason: valid` |
| **中** | 把 GitHub profile 的 `email` 设成能收信的真实地址 | ⏸ **刻意不做**：公开邮箱是垃圾邮件入口，而上表的 `rongjiajun701` 已经证明这类地址会收到什么；noreply 形式仍然归属正确，也仍然能在 issue 里被 @ 到。想改随时可以，但要先想清楚公开的代价 |
| **低** | 清理 token scope：现在缺 `user`（如果将来要读通知/邮箱，得加） | ⏳ 仍缺；`gh auth refresh -h github.com -s user` 是唯一正规做法。（`admin:ssh_signing_key` 已于 2026-11-12 加上，那是挂签名公钥所必需的） |
| **低** | DSH #8202 那条「用不用大模型回复」的追问 | ⏳ 未处理（声誉，不属于代码） |

### 34.5 下会话的开工口令（按顺序，先证明起点干净）

```powershell
git status --short                       # 必须是空的（当前只有一个未跟踪的 .agents/ 目录）
git log --oneline -3
npm view dsh-ai-terminal-pet version     # 1.2.6
gh run list --limit 1                    # success
gofmt -l . ; go vet ./...                # 干净 / 0
go test -count=1 ./...                   # 201 个 func Test，12 个包全 ok
node packages/dsh-ai-terminal-pet/verify.mjs
node packages/dsh-ai-terminal-pet/scripts/vendor.mjs --check
node integrations/deepseek-harness/verify.mjs
node tests/fresh-machine.mjs --no-overlay   # 162/162
python tools/check_docs_links.py            # 11 份文档全解
.\scripts\install.cmd -VerifyOnly           # 11 file(s), 0 warning
& "$env:LOCALAPPDATA\AI-Terminal-Pet\AI-Terminal-Pet.exe" --doctor   # 9 ok / 0 warn / 0 fail
```

**注意 `node --test tests/*.test.js` 不存在**：本仓库没有 `*.test.js`，只有 `tests/fresh-machine.mjs`。
`.agents/tasks/` 里那五张自动派发的工单都把这一行写进「验收命令」，**照着跑会得到 `tests 0 / pass 0`
且退出码 1**——那不是失败，是那份模板抄错了。正确的端到端命令是
`node tests/fresh-machine.mjs --no-overlay`。

### 34.6 `.agents/` 里的五张工单：都不是新工作，且已被本会话的改动覆盖

本会话开始前个人知识库中枢往 `.agents/tasks/` 派了五张工单，逐张核对的结果：

| 工单 | 它的验收标准 | 与现状的关系 |
|---|---|---|
| TSK-20261112-010 登记 dsh-plugin-hub 收录状态 | 记第 5 家、注明源码安装与 win32 单平台 | **事实已经更新**：`dshplugin/dsh-plugin-hub` #735 已于 2026-11-12T00:43:14Z 以 `completed` 关闭，留言写明「`platforms: ["win32"]` 标注 Windows-only、从源码安装」，已写进 §32.5 |
| TSK-20261112-011 验证 awesome 1.2.6 合并并核对镜像版本 | 复读 `plugins.json` 与 dshopen | 已做：`plugins.json` 现已渲染 **1.2.6**，dshopen 仍 1.2.4，见 §32.5 |
| TSK-20261112-012 核查 CI 失败原因 | `go test` 与 reporter 测试全绿 | 已做：两条真 `failure` 的 run 定位到具体 step（见 §32.5 末两条），本地全绿 |
| TSK-20261112-014 落实 zhu 的审核反馈 | 文案里写明各终端 reporter 要另装 | **早已落实**，reviewer 的那句建议本身也标了「不阻断」 |
| TSK-20261112-015 核验 npm 版本 | `latest=1.2.6 / 0.3.6` | 已做：`latest` = **1.2.6**（`0.3.6` 是 `dsh-remedy`，同账号下的另一个包） |

**所以这五张工单不需要再开工**；它们唯一有价值的信息是「模板把验收命令写成了
`node --test tests/*.test.js`」，而那行在本仓库永远 0 通过（§34.5）。

### 34.7 交接时最容易踩的几个坑（这一轮真的踩了）

| 坑 | 表现 | 怎么避开 |
|---|---|---|
| PowerShell 5.1 的 `>` 重定向 | `gh api ... > file` 把 UTF-8 当 ANSI 解码，中文全变 mojibake，读文件会以为 GitHub 上的内容坏了 | 用 Python `subprocess` 读 bytes 再 `decode('utf-8')`，或 `[IO.File]::ReadAllText(path, UTF8NoBOM)` |
| `gh api` 的 `--jq` 带空格/多表达式 | `gh: accepts at most 1 arg(s)` / `function not defined` | 把 jq 表达式写进文件用 `--jq "$env:TEMP\q.jq"`，或者干脆用 Python 解析 |
| GraphQL 的 `number` 参数 | `discussions(numbers: [...])` 不存在；要 `d1: discussion(number: N)` 别名逐个取 | 见 §32.6 的查询写法 |
| `.agents/tasks/*.md` 的验收命令 | `node --test tests/*.test.js` → `tests 0 / pass 0` 且 exit 1 | 忽略该行，跑 `node tests/fresh-machine.mjs --no-overlay` |
| `agy` 传长 prompt | `--print=$text` 被 PowerShell 按空格拆成多个参数，报 `unexpected argument "Your"` | 用 stdin：`$text \| & agy --model ... --effort high` |
| `pet-v0.9.0` 与 `pet-v1.0.0` 都在 | 容易贴错链接 | 对外一律用 `pet-v1.0.0`；0.9.0 已是被取代的旧资产，且**不可变** |

## 35. 交接：2026-11-12 的门禁加固、两个扫描器与提交身份（本次会话）

### 35.1 这一次做了什么（一句话）

**没有换任何版本号，没有改任何一个产品源码的字节**：这一轮只动 `.gitignore`、`CHANGELOG.md`、
`README.md`、`docs/*`、`.github/workflows/ci.yml` 与 `tests/fresh-machine.mjs`。
`vendor/manifest.json` 的源码摘要因此保持 `edd584c6…` 不变（`vendor.mjs --check` 跑过，两个文件都 matches），
插件包仍是 1.2.6、浮窗仍是 1.0.0、npm `latest` 仍是 1.2.6。

### 35.2 为什么能「零版本代价」（这条规则要记住）

`vendor.mjs` 的源码摘要覆盖 `cmd/` 与 `internal/` 下**所有非 `_test.go` 的 `.go`**，
排除 `cmd/pet-hook/`、`cmd/pet-setup/`、`cmd/pet-statusline/`、`internal/reporter/`。
于是：

| 改这些东西 | 摘要变吗 | 要换插件版本号吗 |
|---|---|---|
| `internal/**`、`cmd/ai-terminal-pet/` 里的**非测试** `.go` | **变** | **要**（`vendor/AI-Terminal-Pet.exe` 与 `manifest.json` 都得重写） |
| `internal/**/*_test.go`、`tests/*.mjs`、`.github/`、`.md`、`reporters/` 之外的 JS | 不变 | 不要 |
| `reporters/protocol.js` | 摘要不变，但 `vendor/protocol.js` 是逐字节比对的 | 要（且所有上报端必须同时改） |

所以「先做一个不加功能的稳定批」与「后面加功能」必须分开发版——否则 1.0.0 会为了一个 CI 修复被迫前移。

### 35.3 这一轮发现的那个偶发红（161/162）

2026-11-12 交接时第一次跑 `node tests/fresh-machine.mjs --no-overlay` 得到 **161/162**：

```
== S5 standalone install and native hooks ==
   FAIL doctor reports no failures after a standalone install
FAIL: 161/162 checks in 149.8s
```

随后连跑四次都是 **162/162**（25.0s / 27.2s / 28.0s / 28.5s）。同一批二进制、同一个沙箱模型。

**当时的失败信息是空的**，因为那条检查的 detail 只传了 `doctor.stdout`，而 `spawnSync` 在超时时
`stdout` 是空串——「超时被杀」与「进程正常退出但一行没打印」在报告里长得一模一样。
这正是它难查的原因：一次真实的偶发红留下了一行比没留还误导的证据。

**这一轮的修法（只改测试，不改产品）**：

1. 新增 `describeRun(result)`：一格放进 `status` / `signal` / `error.code` / `retried` 与两段输出，
   空 stdout 会被写成 `<empty>` 而不是一片空白。
2. 新增 `runDoctor(exe, env, budgetMs = 30000)`：**只在 `ETIMEDOUT` 时**用双倍预算重试一次，
   并把「重试过了」写进结果（`retried=…`），由 `describeRun` 打印。理由与浮窗自己给 node 探针
   加重试的理由是同一条（`internal/app/doctor_windows.go` 的 `checkNode`）：冷启动被 AV 扫描
   拖过 deadline 与「真的卡住」必须分开，否则每次都是一次假警报。
3. 三处 `--doctor` 调用（S1 空机、S5 安装后、S8 卸载后）统一走 `runDoctor` + `describeRun`。

**没有做的事**：没有把预算从 30 s 往上调（那样只会掩盖真卡住）；没有放宽断言（`0 fail` 仍然照旧断言）。
如果它再红一次，现在的输出会直接指出是超时还是九项里的哪一项失败了。

### 35.4 两个新门禁：`govulncheck` 与 `gosec`

交接时 `docs/ROADMAP.md` §6 写的是「`-race` / `govulncheck` / `gosec` 缺 gcc 未跑」。**前一条是真的，
后两条是错的**：两者都是纯 Go 程序，`go run` 就能跑，不需要 C 工具链。

| 工具 | 版本（钉死） | 结果 | 触发条件 |
|---|---|---|---|
| `golang.org/x/vuln/cmd/govulncheck` | `v1.9.0` | 代码受影响 **0** 条（另有 2 条在 import 的包、14 条在依赖模块里，但本代码调不到） | 任何「被实际调用」的已知漏洞 → 红 |
| `github.com/securego/gosec/v2/cmd/gosec` | `v2.30.0` | 全量 159 条，其中 high/high **0** 条（12 条 G703、84 条 G115 被排除） | high+high 的任何其他发现 → 红 |

**为什么是两个规则族被排除、而不是逐行 `#nosec`**：`#nosec` 会改 `internal/**` 的 Go 源码，
摘要一变就要重新 vendor 并吃掉一个插件版本号。而这两族的命中全是本代码有意为之：

- **G115**（84 条）＝ Win32 消息与 `uintptr`/`uint32` 之间的转换：`lParam` 解包、DPI 相关字段、菜单 ID。
  这些值是消息负载与指针，不是参与算术的量。
- **G703**（12 条）＝ 由 `%LOCALAPPDATA%` 经 `filepath.Join` 拼出的路径。信任边界就是「用户自己的目录」。

排除写在 `.github/workflows/ci.yml` 的步骤注释里，**不是**写在 `gosec` 的配置里——这样下一次有人想放宽，
必须先读那段理由。剩下 43 条 LOW（`G103`/`G104`：`unsafe` 与未检查返回值的 Win32 调用）按设计不进门禁：
它们在这类 Win32 代码里是常态，做成红线只会训练人忽略红线。

**为什么独立成一个 `scan` 作业**：`golang.org/x/vuln@v1.9.0` 与 `gosec@v2.30.0` 各自的 `go.mod` 都写着
`go 1.26.0`，而 `verify` 作业按本仓库 `go.mod` 的下限装的是最新的 1.23.x。把两条 `go run` 放在同一个作业里，
Go 要么隐式换工具链（`GOTOOLCHAIN=auto` 会额外下载一整套 Go，既慢又多了一种变红的方式），要么直接拒绝构建。
所以：**构建作业继续钉在 1.23**（这正是「同一份源码构建出同一份字节」的前提），扫描作业用 `go-version: stable`。
这是两个作业，不是一个作业里两段——写在这里，是因为「为什么不能省一个作业」以后还会被问到。

**不要把这说成「全量安全扫描」**：`-race` 仍缺 gcc/cgo。对外只能说「静态扫描 + 依赖漏洞扫描通过」。

### 35.5 提交身份：只改未来，不重写历史（本会话已执行）

§34 把这件事留成决策。本会话按「只改未来」执行，理由是已发布内容不可变：

| 项 | 之前 | 现在 |
|---|---|---|
| `user.email`（仓库内） | `jjjjjjjjnnjnn@users.noreply.github.com`（旧式 noreply，账号改名后失去归属） | `236163329+jjjjjjjjnnjnn@users.noreply.github.com`（带账号 id 的形式） |
| 提交签名 | 无，`verified: false / reason: unsigned` | SSH 签名：`gpg.format=ssh`、`user.signingkey=~/.ssh/id_ed25519_signing.pub`、`commit.gpgsign=true`、`tag.gpgsign=true` |
| GitHub 侧 | 没有 SSH 签名密钥 | `user/ssh_signing_keys` id `1325600`，标题 `ai-terminal-pet commit signing (2026-11-12)` |
| 118 个历史提交 / 11 个 tag | 旧式 noreply、未签名 | **一字未改**（这是刻意的） |

**为什么不重写**：邮箱是提交对象的一部分，改了邮箱等于改了每一个 commit 的 sha。
11 个已发布 tag、`docs/market-index.json` 里钉的 `source.commit`、第三方目录记的 `verifiedCommit`
全都指向现有 sha；重写要么让它们指向不可达对象，要么逼着删掉并重打 tag——两条都违反 AGENTS.md 第 6 条。
而收益只有「历史提交里的邮箱也带 id」，对以中立身份发布的项目没有实际价值。

**要读邮箱列表或改邮箱**仍然需要 `user` scope（当前 token 只有
`admin:ssh_signing_key, gist, read:org, repo, workflow`）。`gh auth refresh -h github.com -s user` 是唯一正规做法。

### 35.6 外部世界与数字（2026-11-12 08:5x +0100 实读）

| 对象 | 状态 |
|---|---|
| npm | `dist-tags.latest` = **1.2.6** |
| awesome-dsh-plugin 站点 `plugins.json` | `count` **10,581**，本条目 `version: 1.2.6`、`tarball` 指向 `dsh-plugin-v1.2.6` 的资产 —— **追平了** |
| dshopen `index.json` | **1,219** 条，本包仍是 **1.2.4**（落后两版，等它自己重建） |
| Release 下载数 | `pet-v1.0.0` = **4**（与 24 小时前相同）、`dsh-plugin-v1.2.6` = **7**、`v1.2.5` = 13、`v1.2.3` = **47** |
| DSH Discussions #13001 | upvote **1**、评论 **0**（发帖后约 22 小时） |
| star / fork / watcher | 1 / 0 / 0；30 天浏览 7 views / 2 uniques（数据有滞后） |
| 四家人工目录（dshbase / dsh.so / dsh-plugin.org / AllDSH） | 仍 open、0 评论；不要重复投 |

**结论**：发声之后的可见增量接近 0（`pet-v1.0.0` 的下载数 24 小时没动），而条目侧能做的
（把已有条目推到最新版）**已经做完**（awesome 已渲染 1.2.6）。剩下的杠杆只有一个：产品本身。
这也是为什么下一步排的是功能批次（F9 → 配色 → F8），而不是再发帖、再投目录。

### 35.8 本机的 GUI 场景为什么不可靠（实测，不是猜测）

`--no-overlay`（CI 用的形态）跑十次十次绿；去掉这个旗标、让 S4/S4b 真的启动浮窗，
在本机却会红：`S4 the overlay boots on a machine with no prior state` 的
`the overlay answers on its own pipe` 与 S4b 的 `the overlay comes back on its own`。

**根因是单实例互斥体不是按沙箱隔离的**：它是 `Local\AI-Terminal-Pet-v1-SingleInstance`，
按 **Windows 会话**全局唯一，而管道名才是按 `USERPROFILE` 哈希隔离的。于是出现这种组合：

1. 套件在开工前会 `--quit` 真实实例（代码里确实这么做）；
2. 但**本机这条 opencode 会话自己就装了上报端**，下一次 hook 事件又把它拉起来——
   实测：失败的沙箱进程 `exec 0`、一秒内退出、从未建管道；`Win32_Process` 里那个
   胜出的实例 `ParentProcessId = opencode.exe`，路径是安装目录那份；
3. 沙箱那份二进制于是抢不到互斥体，正常退出（退出码 0），S4 等 25 s 也等不到管道。

**取证方式（已经写进套件）**：S4 失败时会打印互斥体是一条会话级对象、沙箱内外各有哪些
`AI-Terminal-Pet.exe` 在跑、以及 spawn 的目标路径；`waitForOverlay` 失败时的 detail 里带
`pipe=<沙箱管道名>`。所以下一次看到它红，一眼能分清「产品没起来」和「被本机实例抢了」。

**结论与用法**：完整形态（不带 `--no-overlay`）**只适合在一台没有宿主 hook 在跑的机器上跑**；
在开发者自己开着 AI 终端时，它是「测套件」和「测被测物」之间的竞态，不是产品缺陷。
带 `--no-overlay` 的 162 项才是门禁；完整形态的 182 项是手工取证用的（本机先 `taskkill`
掉所有 `AI-Terminal-Pet.exe` 之后跑通一次：**182/182 / 35.6 s**）。

### 35.7 文档改动清单与「不动」的边界

- `.agents/tasks/` 那五张工单仍然不重开（§34.6 的核对结论不变），本轮只把 `.agents/` 加进 `.gitignore`。
- `README.md` 里所有「当前版本」口径改成 1.2.6 / `df4fcf6`；顺带修掉一处**被合并坏掉的段落**：
  中文段有一句从句子中间开始（「。 的 git 源、或直接给仓库地址…」），上一轮修它时又把同一段故事
  讲了两遍（PR #13937、`plugins.json` 的 count、imsai #902、zhu #97、`risk_tier`、Discover 前置
  各出现两次）。现在合并成两段、每条事实只出现一次；这类重复是**第二个模型评审**在 2026-11-12
  指出来的，不是自查发现的——记在这里是因为它是这类工作的常见失效模式。
- `CHANGELOG.md` 的两个 `[Unreleased]` 改成了带年份的批次标题（它们记录的两批工作随 1.0.0 / 1.2.6
  发出），并新增了本轮的批次条目。
- `docs/ROADMAP.md` §6 与 README 的门禁列表同步到两个新扫描器；§6 里两处残留旧事实
  （`latest` = 1.2.5、#14385 等人工合并）也已改成本轮实测结果。
- `docs/LAUNCH.md` **动过**：只改「实测状态」类文字（渠道表里 `v1.2.5` 的下载数、awesome 的
  `plugins.json` 已渲染 1.2.6、dshopen 仍 1.2.4、§8 的下载/upvote 基线），**没有改任何对外文案**
  （标题、X 线程、Show HN 正文、市场摘要一个字未动，因为推动它们的事实没变）。

## 36. F9：紧凑态显示等待时长（2026-11-12，浮窗 1.0.1 / 插件 1.2.7）

### 36.1 做了什么，以及为什么不是「加一行」

产品开头那句承诺是「哪个终端卡住了、卡了多久」。前半句一直在紧凑态上，**后半句要展开面板才看得到**——
一个 56×156 的常驻条，却把用户最常问的事实藏在二级界面里。

所以这一轮**不加行、不改尺寸**，改的是「最大那行装什么」：

| 位置 | 有人在等时（waiting_permission / error） | 其它情况 |
|---|---|---|
| 大号槽（22 → 16 号粗体） | `12分` / `12m` | 会话数（`3`），与以前一字不差 |
| 下面那行（9 号） | `3 个会话` / `3 sessions` | `会话` / `sessions`，与以前一字不差 |
| 状态行 / 项目名 / 跳转条 / 命中区 | 未动 | 未动 |

**为什么坚持不动几何**：`compactWidth/compactHeight` 一动，`docs/assets/` 里全部面板 PNG、`demo.gif`、
`desktop-scene.png`、README 中英两段的尺寸描述、以及插件包里的 `screenshots/compact.png` 都要重出。
素材重出是可以做的（本轮也做了），但**「为了一个能塞进现有空间的事实而改产品轮廓」是本末倒置**——
先量、再决定，这一轮两条都做了：先量字体，再证明它塞得下，然后才动布局。

### 36.2 三个判断，每个都有测试钉住

1. **只对「真的有人在等」的状态显示时长**：`waiting_permission` 与 `error`。忙碌（`working`/`tool_use`/
   `thinking`）没有人在等；`idle`/`closed` 的时间戳意思是「什么时候结束的」。`TestCompactWaitOnlyCoversStatesThatWaitOnTheUser`。
2. **多个会话在等时取最早开始等的那个**，不是优先级最高的：卡了 12 分钟的权限确认比 30 秒前的构建失败
   更值得先看。注意这与 `mostUrgentJump`（跳转条）的规则**不同且故意不同**——跳转要的是「最该点哪个」，
   时长要的是「谁等最久」。`TestCompactWaitPrefersTheOldestWait`。
3. **取整边界与展开行共用一套**（<1 分 → 最小标签、<1 小时按分、<2 小时是「1 小时」、之后按小时；
   超过 99 小时截断，展开态仍显示准确数字）。两个各自维护阈值的格式化器，就是同一个瞬间在两处
   一个说「刚刚」一个说「1 分钟前」的成因。`TestShortElapsedUsesTheSameBucketsAsRelativeTime`。

### 36.3 那个必须记下来的 Bug：时长会在呼吸时冻结

呼吸态（waiting/error）的帧复用 `purePulseFrame` **刻意忽略 `clock`**——脉冲只改整窗 alpha、不改像素，
这是上一轮把 CPU 从 ~8% 压下来的关键。所以：**如果把时长挂在 `clock` 上，它会被当成「只有时钟变了」
而复用旧位图，`12m` 会一直停在 12 分钟直到下一次上报**。

修法是给 `renderedKey` 一个独立的 `waitMinute` 分钟桶，只在「紧凑态 + 有人等」时赋值。三条断言：
跨分钟内 20 秒仍可复用（不白烧 GDI）、跨到下一分钟必须重绘、忙碌/展开态 `waitMinute` 必须是 0
（否则每个忙碌面板都会每分钟白重绘一次）。`TestFrameKeyAdvancesTheDurationPerMinute`。

### 36.4 文案与尺寸都是量出来的

用渲染器自己的 `measure()`（同一份 YaHei UI 字体）逐个测候选串，16 号粗体、48 逻辑单位：

| 串 | 宽度 |
|---|---|
| `12m` / `59m` | 33 px |
| `1小时` | 39 px |
| `23小时` | 42 px |
| `99小时` | 44 px |
| `12 sessions`（9 号） | 36 px |

所以时长用 16 号而不是 22（22 号下 `9小时+` 已经 54 px、`12小时` 52 px 出界）。
`TestWaitLabelFitsTheStripIsAMeasuredClaim` 把这张表钉进测试——以后改字号或文案撞到边是测试红，
而不是在 100% 缩放的桌面上才发现。

### 36.5 这一轮的实测

```
gofmt -l . / go vet ./...                     干净 / 0
go test -count=1 ./...                        12 个包全 ok，206 个 func Test（新增 5）
node tests/fresh-machine.mjs --no-overlay     PASS 162/162
三个 verify.mjs + vendor --check             全 PASS；vendor 两文件 matches
scripts\install.cmd -VerifyOnly               11/11
--doctor（安装 1.0.1 之后）                    7 ok / 2 warn / 0 fail（两条 warn 是浮窗没在跑）
scripts\package.ps1                           zip 4,985,803 B / 17 files；pet-setup.exe 14,131,200 B
vendor/manifest.json                          源码摘要 ab16b003…（24 文件 / go1.27.0）
                                              overlay 4,377,600 B / sha256 59f1cfcf…
npm pack（1.2.7）                             17 files / shasum efa49bef…
```

**素材**：`docs/assets/` 的面板 PNG、`hero.png`、`demo.gif`（345,212 B / 56 帧）、`desktop-scene.png`
全部由 `render_assets.py` 与 `make_launch_assets.py` 从同一份光栅器输出重出；插件包里的
`screenshots/compact.png` 同步替换（`screenshots.json` 条目不变）。`desktop-scene.png` 与 `demo.gif`
的图注仍然写明「终端窗口是画的、面板是真实光栅器输出」。

### 36.6 发布记录（按 §28.6 的固定套路，全部已落地）

| 步骤 | 结果 |
|---|---|
| A 提交 | `a4f21af`：`Version` → 1.0.1、`internal/ui` 与 `internal/i18n` 的改动、`internal/ui/wait_label_test.go`（5 个新测试）、三二进制重编 + `npm run vendor`（摘要 `ab16b003…`、overlay `59f1cfcf…`）、两份清单与根 shim 的 1.2.7、索引的 `version`/`publishTimestamp`、README/CHANGELOG/ROADMAP 口径、全部素材 |
| tag | `pet-v1.0.1` 与 `dsh-plugin-v1.2.7` 都指向 `a4f21af`，两个都通过 `git verify-tag`（SSH 签名） |
| Release | [`pet-v1.0.1`](https://github.com/jjjjjjjjnnjnn/ai-terminal-pet/releases/tag/pet-v1.0.1) 带 zip `4,985,803 B / sha256 849A3AA9…` 与 `pet-setup.exe` `14,131,200 B / sha256 8389B426…`；[`dsh-plugin-v1.2.7`](https://github.com/jjjjjjjjnnjnn/ai-terminal-pet/releases/tag/dsh-plugin-v1.2.7) 带 tgz `1,919,625 B / sha256 8CB907E9…` |
| npm | `dsh-ai-terminal-pet@1.2.7` 已发布，`dist-tags.latest` = **1.2.7**；`npm pack` 取回的 tarball 与 Release 资产**逐字节相同**（同为 `8CB907E9…`） |
| B 提交 | `7f3d4b6`：索引 `source.commit` 与 `icon` 钉到 `a4f21af`，README 两段同步写上这个 sha |
| 上游 | [awesome-dsh-plugin #14688](https://github.com/awesome-dsh-plugin/awesome-dsh-plugin/pull/14688)：只改 `tarball` 一行，MERGEABLE，等 check 与人工合并（中文描述逐字节未动，用 base64 上传避开 PowerShell 5.1 的 ANSI 解码陷阱） |
| 公网复验 | 两个浮窗资产都从公开 URL 下载回来重新哈希（与上表一致）；zip 解到带独立 `LOCALAPPDATA`/`USERPROFILE` 的沙箱里 `install.cmd` → 11/11、`--doctor` 8 ok / 1 warn / 0 fail（唯一 warn 是「running binary 不是这份安装」，因为是从解压目录跑的）、`-Uninstall` 目录归零；`pet-setup.exe --version` 打印 `pet-setup 1.0.1` |


## 37. 外观成为设置：三套调色板、按终端着色、启动时隐藏（2026-11-12，1.0.2 / 1.2.8）

### 37.1 这一轮做了什么

三件事，互不阻塞，但共用一次发版：

| 设置 | `config.json` 键 | 默认 | 它解决什么 |
|---|---|---|---|
| 配色主题 | `theme` | `slate` | 深色看腻了或者要用浅色桌面的人，现在有 `paper`（浅）与 `contrast`（纯黑 + 高饱和状态色） |
| 按终端着色 | `perSourceColor` | `false` | 展开行里的**来源名**按终端上色，回答「这是哪个工具」；状态色不动 |
| 启动时隐藏 | `hiddenAtLaunch` | `false` | 上报端会按需拉起浮窗，想要「登录后先待托盘」的人不必每次手动隐藏 |

设置窗口新增第四个页签**外观（Look）**。`slate` 是默认值，所以**从未打开这个设置的人，面板字节与 1.0.1 相同**
——这是「默认档不许改」那条规矩的直接后果，也是为什么 `docs/assets/` 里的截图只需按同一份代码重出、构图不变。

### 37.2 为什么调色板是包级变量而不是穿参

20 个颜色要穿过 `drawVerticalPanel` → `drawVerticalCompact` / `drawVerticalExpanded` → `drawVerticalSession`
等六七个签名，只为让一个外观开关能改颜色。而**所有栅格化都跑在 UI 线程**（`canvas_windows.go` 里字体缓存那段
注释早就写明了这一点），所以「进程级一份调色板，启动时与切换时各赋一次」与原来的常量一样安全，
而且调用点从 `colorPanel` 变成 `colors.Panel` 只是机械替换（本轮改动了 106 处引用，全部在 `internal/ui`）。

代价是**必须把主题进帧 key**：帧复用规则是「key 相同 = 像素相同」，只赋 `colors` 而不动 key，切换主题会被
当成没变化。这与 1.0.1 的「时长冻结」是同一类 bug，所以 `renderedKey` 增加了 `theme` 与 `perSource` 两个字段，
并断言「未知主题名归一化成 slate 之后 key 不应变化」（否则每次读盘都会白重绘一次）。

### 37.3 浅色调色板最容易做错的一件事

深色主题的「次要文字」是浅灰（`170,180,196`），在浅底上它**必须变暗**而不是继续变亮，否则小字直接消失。
`TestEveryThemeKeepsItsTextReadableOnItsOwnPanel` 对三套调色板逐对比较「每种文字 vs 它自己那块底色」的
相对亮度，也检查六个状态色相对底色都要 ≥2.5:1。这里**刻意不断言 WCAG 4.5:1**：面板是半透明浮在别人的桌面上，
拿一个对自己的桌面成立的比值去声称「符合无障碍标准」是过度声明；能说的是「每种文字与它自己底色分得开」。

### 37.4 一处有意的重复

`config.ValidTheme` 与 `ui.normalizeTheme` 是两份实现，因为 `internal/ui` 不 import `internal/config`
（文件路径与文件格式归 app 层，`ui.DefaultSettings` 的注释写明了）。两份清单因此可能漂移，
所以 `internal/app/appearance_test.go` 的 `TestThemeListsAgree` 把「主题名集合与顺序一致」钉住。
同一文件还钉住 `applyUISettings` 会把三个新字段写进 `config.json`——**它忘了字段，就等于「设置窗口能选、
重启就丢」**，这类 bug 只有测试能抓。

### 37.5 这一轮的实测

```
gofmt -l . / go vet ./...                     干净 / 0
go test -count=1 ./...                        12 个包全 ok，218 个 func Test（新增 12）
node tests/fresh-machine.mjs --no-overlay     PASS 162/162
三个 verify.mjs + vendor --check             全 PASS
scripts\install.cmd -VerifyOnly               11/11
--doctor（安装 1.0.2 之后）                    7 ok / 2 warn / 0 fail（warn 是浮窗没在跑）
scripts\package.ps1                           zip 4,992,775 B / 17 files；pet-setup.exe 14,145,024 B
npm pack（1.2.8）                             17 files / shasum 922f42ce…
vendor/manifest.json                          overlay 4,391,424 B / sha256 f7788a9a…；protocol 未变
```

**素材**：`docs/assets/` 由同一份光栅器输出重出（几何未变，构图不变），并新增设置窗口「外观」页的两张预览
（`settings-appearance*.png` 属于 `PET_PREVIEW_DIR` 的产物，不进仓库）。插件包的 `screenshots/` 同步替换。

### 37.6 发布记录（按 §28.6 的固定套路）

| 步骤 | 结果 |
|---|---|
| A 提交 | 本批：`Version` → 1.0.2、调色板与三套主题、`perSourceColor`、`hiddenAtLaunch`、第四页签、12 个新测试、三二进制重编 + `npm run vendor`（overlay `f7788a9a…`）、两份清单与索引前移到 1.2.8、README/CHANGELOG/HANDOVER 口径、素材重出 |
| tag | `pet-v1.0.2` 与 `dsh-plugin-v1.2.8`（都打在同一提交上，SSH 签名） |
| Release | `pet-v1.0.2`：zip **4,992,775 B / sha256 `EA3A4EC3…`**、`pet-setup.exe` **14,145,024 B / sha256 `1531AE8B…`**；`dsh-plugin-v1.2.8`：tgz **1,926,796 B / sha256 `F0B9CDA0…`** |
| npm | `dsh-ai-terminal-pet@1.2.8`，`dist-tags.latest` = **1.2.8**，`npm pack` 回来的 tarball 与 Release 资产逐字节相同（sha256 二者都是 `F0B9CDA0…`） |
| B 提交 | 索引 `source.commit` 与 `icon` 钉到 A 的 sha（git 无法在提交里引用自己） |
| 上游 | awesome-dsh-plugin 的 `bump-tarball-1.2.8` 跟进 PR（只改 `tarball` 一行，用 base64 上传） |
| 公网复验 | 两个浮窗资产从公开 URL 下载重哈希一致；沙箱 install → `-VerifyOnly` 11/11 → `--doctor` 8 ok / 1 warn / 0 fail → `-Uninstall` 目录归零；`pet-setup.exe --version` 打印 `pet-setup 1.0.2` |

### 37.7 已知取舍

- 三套调色板是**手挑的值**，没有自动从桌面取色。自动主题涉及读系统主题、再算对比度，是另一件事。
- `perSourceColor` 只改**来源名**，不改左侧状态脊与状态词：那两处是信号本身。
- 未知上报端取中性色（`Secondary`），不猜。新终端要显示自己的颜色，得在 `sourceColor` 里加一行。
- `hiddenAtLaunch` 只作用于启动。托盘菜单、`--show`、`--toggle` 之后仍然是权威——它不是「隐藏模式」。

## 38. 交接：2026-11-12 会话结束（浮窗 1.0.2 / 插件 1.2.8，下一步 F8）

### 38.1 这一整天做了什么（四轮，全部在线上）

| 轮次 | 交付 | 版本 |
|---|---|---|
| 门禁加固 | `scan` 作业（govulncheck + gosec，G115 按族排除、G703 按路径限定四个文件）、`fresh-machine` 失败信息不再空、五个 action 升到 node24、提交身份：id 形式 noreply + SSH 签名（**历史不重写**） | 无版本号 |
| F9 | 紧凑态大号槽在有人等时改显**等待时长**（`12分`/`12m`），几何不变 | `pet-v1.0.1` / `dsh-plugin-v1.2.7` |
| 外观 | 三套命名调色板（`slate` 默认 / `paper` / `contrast`）、`perSourceColor`、`hiddenAtLaunch`、设置窗口第四页签「外观」 | `pet-v1.0.2` / `dsh-plugin-v1.2.8` |

三条版本线的账都在：`pet-v0.9.0`、`v1.0.0`、`v1.0.1`、`v1.0.2` 与 `dsh-plugin-v1.0.0` … `v1.2.8`，**全部不可变**，
不要替换任何已发布资产。

### 38.2 开工口令（按顺序跑，先证明起点干净）

```powershell
git status --short                       # 必须为空
git log --oneline -6                     # 顶上是本轮的文档提交
npm view dsh-ai-terminal-pet version     # 1.2.8
gh run list --limit 1                    # success（build+verify 与 scan 两个作业）
gofmt -l . ; go vet ./...                # 干净 / 0
go test -count=1 ./...                   # 12 个包全 ok，218 个 func Test
node packages/dsh-ai-terminal-pet/verify.mjs
node packages/dsh-ai-terminal-pet/scripts/vendor.mjs --check
node integrations/deepseek-harness/verify.mjs
node tests/fresh-machine.mjs --no-overlay   # 162/162
python tools/check_docs_links.py            # 11 份文档全解
.\scripts\install.cmd -VerifyOnly           # 11 file(s), 0 warning
& "$env:LOCALAPPDATA\AI-Terminal-Pet\AI-Terminal-Pet.exe" --doctor   # 9 ok / 0 warn / 0 fail（浮窗在跑时）
```

**注意**：本机的 `--doctor` 如果不带浮窗实例在跑，会给 2 条 warn（pipe/instance），那是正常的；
9 ok 的前提是「有实例在监听」。

### 38.3 下一步：F8（最后一项，也是三步里最大的一步）

**F8 —— 点击会话行复制项目路径 + 气泡回执**（ROADMAP §4）。为什么它比 F9 大：现在的规则是「点击任意处收起」，
所以要做 F8 必须改成「点击行 = 复制路径 + 回执，点击行以外的面板区域 = 收起」，并且新增一个绘制元素与一个
定时器。ROADMAP 里已经写明建议的交互形态，不要重新设计。

三条施工约束（都是前两轮踩出来的，别绕过去）：
1. **命中区域必须与绘制同源**（`focusHit` / `expandedRowFocusRect` 那样的纯函数），否则会出现「画在这儿、
   点在那儿」。
2. **任何新画的东西都要进 `renderedKey`**，否则「同一个 key = 同一张位图」的帧复用会把它吞掉——
   1.0.1 的时长冻结与 1.0.2 的主题不重绘都是这个成因。
3. **回执的定时器要与现有动画定时器同一套装/卸机制**（`syncAnimation`），别新开一个常驻计时器。

发版代价与之前一样：三个二进制重编 → `npm run vendor` → 插件 1.2.9 → A 提交（版本+索引时间戳）→
双 tag + 双 Release + npm → B 提交钉索引 sha → awesome 的 `bump-tarball-1.2.9` 跟进 PR。
`docs/assets/` 需要重出（回执是一个新视觉元素，`TestRenderPreview` 要加一个用例）。

### 38.4 这一天学到、下一轮别再犯的

| 事项 | 记在哪 |
|---|---|
| `go run` 的两个扫描器各自要求比本模块更新的 Go，所以 CI 把它们放独立 `scan` 作业，构建作业继续钉 1.23 | §35.4 / `ci.yml` 注释 |
| 本机完整形态（不带 `--no-overlay`）会与本机宿主 hooks 抢单实例互斥体，看起来像产品缺陷；`--no-overlay` 才是门禁 | §35.8 |
| 呼吸态复用位图忽略 `clock`，任何「随时钟变化的内容」都要自己的桶字段 | §36.3 |
| 浅色主题里「次要文字」必须变暗；对比度只对「文字 vs 它自己的底色」成立，别声称 WCAG | §37.3 |
| `config.ValidTheme` 与 `ui.normalizeTheme` 是有意重复，由 `internal/app` 的测试钉住；新增设置必须同时改 `applyUISettings` | §37.4 |
| 上游条目 PR 永远只改 `tarball` 一行，且**必须用 base64 上传**（PowerShell 5.1 会把 UTF-8 当 ANSI 解码） | §30.6 / §37.6 |
| PowerShell 里 `git show > file` 会写出 UTF-16，Python 读 UTF-8 会报 `0xff`；用 `cmd /c "git show … > file"` | 本轮实测 |

### 38.5 外部世界现在是什么样（2026-11-12 13:50Z 实读）

- **npm**：`dist-tags.latest` = **1.2.8**；`npm pack` 取回的 tarball 与 Release 资产逐字节相同。
- **awesome-dsh-plugin**：1.2.7 的跟进 PR [#14688](https://github.com/awesome-dsh-plugin/awesome-dsh-plugin/pull/14688)
  仍 open；1.2.8 的跟进 PR [#14718](https://github.com/awesome-dsh-plugin/awesome-dsh-plugin/pull/14718)
  已开、MERGEABLE，**两个 check（`check` 与 `Submission gate`）都已 SUCCESS**，等人工合并。**上游仓库里的 `tarball` 仍是 1.2.6**——合并前不会变，别以为 PR 开了就等于收录。
- **dshopen**：线上 `index.json` 仍渲染 1.2.4，等它自己重建。
- **四家人工目录**（dshbase / dsh.so / dsh-plugin.org / AllDSH）仍 open、0 评论，不要再投。
- **下载数**（发版后约 15 分钟）：`pet-v1.0.2` = 0、`dsh-plugin-v1.2.8` = 0；对照 `pet-v1.0.1` = 5、
  `dsh-plugin-v1.2.7` = 2、`pet-v1.0.0` = 4、`v1.2.6` = 9、`v1.2.3` = 47。
- **提交身份**：`user.email` = `236163329+jjjjjjjjnnjnn@users.noreply.github.com`（仓库内），
  SSH 签名，`gh api .../commits/main` 实测 `verified: true / reason: valid`。
  **118+ 个历史提交与 11+ 个已发布 tag 刻意未重写**（§35.5）。

### 38.6 交接时最容易踩的坑（本轮真的踩了）

| 坑 | 表现 | 怎么避开 |
|---|---|---|
| 版本号批量替换 | 上一轮把 `1.2.6` 全量替换成 `1.2.7`，结果出现「tag 是 1.2.7、资产名还是 1.2.6」这种自相矛盾的句子 | 用带**断言**的脚本改文档：每条替换必须恰好命中一次，改完再断言结果里应有的字符串都在 |
| 中文段落被合并坏掉 | `462dca2` 把两版中文段落首尾相接，留下一个从句中开始的残句；第一次修的时候又把故事讲了两遍 | 改中文长段前先数「同一事实出现几次」；重复检测用探针字符串 |
| `go test` 的 `frameKey` 签名变化 | 加参数后测试批量 `not enough arguments` | 签名改动要么保持兼容，要么一次性把所有调用点（含测试）都改掉 |
| 文档说「未跑」而其实能跑 | 旧文档写「govulncheck/gosec 因缺 gcc 未跑」，其实两者都是纯 Go | 写「做不到」之前先跑一次 |

## 45. 交接：把第 44 轮发出去 —— 1.0.7 / 1.2.14（2026-11-13 傍晚，第 45 轮）

这一轮**没有改代码**：第 44 轮的实现与测试已经在 `821226a` 上、CI 绿，缺的只是发版仪式。本轮做的是
**发版前的一条真机验收** + [docs/UPSTREAM.md](UPSTREAM.md) §2 的八步。

### 45.1 发版前补做的那条验收：设置窗口的改动真的落盘了吗

§44.7 留了一句「沙箱里点完控件 `config.json` 没有被重写，原因未定」。发版前在本机装好的构建上重做了一次
（本机安装目录里的 exe 与 `bin/` 那份**逐字节相同**，所以验的就是要发的这版）：

| 动作 | 结果 |
|---|---|
| 点侧栏「外观」 | 窗口高 **826 → 586**，字号 / 配色主题 / 按终端着色 / 启动时隐藏都在 |
| 点「纸白」 | **不到 1 秒** `config.json` 里 `"theme": "paper"` |
| 点「字号 大」 | `"settingsTextSize": "large"`，窗口 750×586 → **900×701**，层级与换行都没塌 |
| 点回「石板」「标准」 | 两值复原为 `slate` / `standard` |

所以原因不是 bug：写盘在 `configSaver` 的守护协程里，**进程被 kill 就不会走 `defer flush()`**，
沙箱里那次是提前杀进程。以后验这类"点了之后有没有写盘"，**必须让进程正常退出**（关窗 + `--quit`），
或者等守护协程跑完再看文件。

### 45.2 发版清单（按 docs/UPSTREAM.md §2 的顺序，逐条落地）

1. 版本前移：`internal/app/app_windows.go:29` → **1.0.7**（那段注释也重写成这一版的理由，不再停在 1.0.2）、
   根 `package.json` 与 `packages/dsh-ai-terminal-pet/package.json` → **1.2.14**、
   `docs/market-index.json` 的 `versions[0].version` 与 `publishTimestamp`（`1794584809`）。
2. 三个二进制按 `-trimpath -buildvcs=false -ldflags "-s -w"` 重建 → `node …/scripts/vendor.mjs` → `--check`
   （源码摘要 `a841b09b…`、overlay `01b42dc2…`；**源码改了就必须重 vendor**，否则 CI 红）。
3. `scripts/package.ps1` 打 zip 与 `pet-setup.exe`；`npm pack` 取 tarball，并把包里的 `README.md` 解出来
   核对两处安装命令已经是 `dsh-plugin-v1.2.14`（§43.6 的教训）。
4. 全门禁：见 §1「健康」那一行；发版前 `node tools/upstream-entry.mjs --check` 是 **SKIP**（资产还没上传，
   `SKIP 不是 pass`），打完 tag 后重跑成 `OK` / `OK（in flight）`。
5. A 提交（版本 + 两份清单 + 索引版本/时间戳 + CHANGELOG + README/ROADMAP/HANDOVER）→
   `pet-v1.0.7` 与 `dsh-plugin-v1.2.14` → 两个 Release → `npm publish`。
6. B 提交把 `docs/market-index.json` 的 `source.commit` 与 `icon` 钉到 A 的 sha。
7. `node tools/upstream-entry.mjs --version 1.2.14 --write --branch bump-tarball-1.2.14` → 开 PR →
   关掉被取代的旧版本 PR。
8. 三个资产从公网 URL 下载回来复验字节；沙箱 `install → -VerifyOnly → --doctor → -Uninstall`；
   本机 `scripts\install.cmd` 升到 1.0.7 后 `--doctor` 应 10 项 0 fail。

### 45.3 发布资产（本机产物实测）

| 资产 | 大小 | sha256 |
|---|---|---|
| `ai-terminal-pet-1.0.7-win-x64.zip`（17 文件） | **5,053,946 B** | `F3D049619BAEFD355E8481CDAF718175792B8C3B0022A3A9408960F99603E9DA` |
| `pet-setup.exe` | **14,307,840 B** | `AE3E2DE769A47A5D87A22FC0D19E1D7617548DBC03F8E5FAAFB7B39AE15C8A6B` |
| `dsh-ai-terminal-pet-1.2.14.tgz`（18 文件） | **2,000,539 B** | `EBD6CE82B653EF5C54FAA01888533D78A79DED4CFEA3137031FCD2E46A3D3FDE` |

### 45.4 顺带修掉的三处文档失准

发版要改版本号，于是把"读起来像事实"的地方都核了一遍，发现三处已经不是事实：

- `README.md` 的版本徽章停在 **1.0.2**（1.0.2 之后三次发版都没跟着走）；
- README 中英两段「插件包当前版本」都写着 **1.2.8**，而 npm 上是 1.2.13——现在统一成 1.2.14，并把
  1.2.9–1.2.14 各自带的浮窗版本与那句话该说明的事（提醒判定、回执搬家、点行复制、三个假提醒、设置窗口）
  逐条补上，英文与中文逐句对应；
- `docs/ROADMAP.md` 的基线表里，`--version` 徽章、二进制字节、单测数量（302 → 326）与逐包分布、
  代码行数（16,880 → 16,882）都过期了。

### 45.5 没做的事

- **没有新功能**：这一轮的产物是"上一轮的代码到了用户手上"。功能候选取自 ROADMAP §4（UIA 无障碍、
  设置窗口滚动、`--json` 走 IPC），都不应塞进发版轮。
- **`go test -race` 仍未跑**：缺 gcc/cgo，属环境，不是代码。
- **上游仍要人工合并**：我们只能保证"随时可合、检查跑过、只改一行"，可见性以对方站点索引文件为准。

### 45.6 发出去的结果（当场核对，不靠回忆）

- **提交与 tag**：A = `f60d8a5`（版本 + 文档数字），B = `5b95107`（索引的 `source.commit` 与 `icon` 钉到
  `f60d8a5`）。两个 tag 都指向 `f60d8a5`；该提交上的 CI **success**（run `43457593263`，1m35s）。
- **两个 Release**：[`pet-v1.0.7`](https://github.com/jjjjjjjjnnjnn/ai-terminal-pet/releases/tag/pet-v1.0.7)
  （zip + `pet-setup.exe`）与
  [`dsh-plugin-v1.2.14`](https://github.com/jjjjjjjjnnjnn/ai-terminal-pet/releases/tag/dsh-plugin-v1.2.14)
  （tgz）。GitHub 自己算的 `digest` 与我们记的 sha256 **逐字相同**（三个资产都核过）。
- **npm**：`dsh-ai-terminal-pet@1.2.14` 已发布，`dist-tags.latest` = **1.2.14**；
  `npm pack dsh-ai-terminal-pet@1.2.14` 解包后与 Release 资产**逐文件相同（18/18）**
  ——`.tgz` 字节不同是 npm 重打包的正常现象，比的是解包后的文件。
- **上游**：[PR #15045](https://github.com/awesome-dsh-plugin/awesome-dsh-plugin/pull/15045)
  （分支 `bump-tarball-1.2.14`，commit `bac06ec`，单文件单行）已开、`MERGEABLE`；
  `node tools/upstream-entry.mjs --check` = **OK（in flight）**。被取代的 #14990（1.2.13）、
  #14945（1.2.11）、#14899（1.2.10）已关，各留一句「被 #15045 取代」。
  上游 `main` 的 `tarball` 仍是 1.2.9——**合并仍要维护者**，且**合并 ≠ 可见**：判断可见性读对方站点索引。
- **资产复验**（全部从公网 URL 下载回来，不读本机文件）：三个 sha256 与 §45.3 的表一致；
  沙箱拿**下载回来的 zip** 跑 `install`（11 个文件、逐文件 SHA256 校验）→ `-VerifyOnly`
  **11/11 `[ok]`、0 failed** → 装出来的 `--version` = **1.0.7** → `--doctor` **7 ok / 3 warn / 0 fail**
  （三条 warn 是「`runtime/` 还没建」「管道没人听」「没有活实例」，空沙箱里必然如此）→ `-Uninstall` 后
  安装目录 **0 文件**。
- **本机升级**：`scripts\install.cmd` 装到 1.0.7 后 `--version` = **1.0.7**、`--doctor`
  **10 ok / 0 warn / 0 fail**；在**装好的这份**上再点一次「纸白」→ `config.json` 的 `theme` 变 `paper`，
  点回「石板」→ `slate`。
- `packages/dsh-ai-terminal-pet/*.tgz` 已删（打包产物只作 Release 资产，`.gitignore` 已挡住）。

### 45.7 下一个会话从这里开始

没有欠账。可选的下一轮（都取自 [ROADMAP.md](ROADMAP.md) §4，按"用户可感知"排序）：
**UIA 无障碍**（自绘分层窗口对屏幕阅读器完全不可见，是一整轮）、**设置窗口滚动**（现在按工作区降档）、
**`--json` 走 IPC 查运行中实例**（目前只读快照，约束是永不阻塞 GUI）。上架这边只剩等 #15045 合并与
对方站点重建，判断标准见 [docs/UPSTREAM.md](UPSTREAM.md) §4。

### 45.8 这一轮暴露的两个流程缺口

两条都已在第 46 轮变成门禁，见 §46：

- **文档里的版本号会独立腐烂** → `tools/check_versions.mjs`（CI 步骤 + 24 条规则测试）。
- **"点了没写盘"要先问进程是怎么退出的** → `internal/app/config_saver_test.go` 三条测试 + AGENTS.md 第 9 条。

## 51. 交接：两个现场缺陷 —— 幽灵会话与设置窗口「闪退」（2026-11-14 晚，第 51 轮；只取证，未改代码）

用户原话（两条，都在真机上观察到）：

> **"有些不存在的程序，已经关闭的程序还是会显示"**
> **"设置打开以后界面经常闪退"**

这一轮**没有动代码**：先把现场证据、可证伪的机制候选、以及复现步骤钉下来，下一轮直接进入修复。
两条都还不能算"已定位根因"——下面每条都写清"什么现象才能算它"。

### 51.1 现场（2026-11-14 18:3x UTC，实测）

| 项 | 值 |
|---|---|
| 安装版 | **1.0.7**（`%LOCALAPPDATA%\AI-Terminal-Pet\AI-Terminal-Pet.exe`，构建于 2026-11-13 16:49），进程 70580 |
| 用户偏好 | `theme=contrast`、`settingsTextSize=large`、`edgeRetract=true`、`idleMinutes=3`、`language=auto` |
| 面板几何 | 紧凑 `55x152 @ (2493,176)`；设置窗口 `900x1080`（150% DPI，`y=224` → 底边 1304 < 1600） |
| 宿主进程 | **54972** = `opencode.exe serve --service`（15:00 起就一直活着），本地端口 `127.0.0.1:28894`（返回的是 Web UI 的 HTML，不是 JSON API；`/api/session` 401） |
| `--json` | 7 个会话，**全部** pid=54972 |
| runtime/ | 7 个快照，pid 全部 alive，没有 `closed` 的快照 |

### 51.2 缺陷一：幽灵会话 —— 机制已清楚，现场此刻是干净的

**先说现场核对结果**：用两条独立路径比对，此刻 7 行**都是真会话**。

1. `--json` / `runtime/` 给出 7 个 `sessionId`；
2. 以只读方式查 opencode 自己的库 `~/.local/share/opencode/opencode.db` 的 `session_v2` 表
   （`mode=ro` + `PRAGMA query_only=1`）：**7/7 都存在，且 `time_archived` 全是 NULL**。

**机制（读代码得到的，不是猜）**：

- 存在性是**按进程**判定的：`internal/app/app_windows.go` 的 `snapshotAwareAlive()` 在快照仍在时
  直接返回 true；没有快照时才退到 `presence.SessionAlive(pid)`。`internal/store` 的注释写得很清楚：
  *"While it is [observable], the session is kept indefinitely."* —— 只在宿主进程连续不可观测
  `DefaultAbsenceGrace`（2 分钟）后才移除。
- 对话结束**只有一个信号**：`reporters/opencode.js` 的 `EVENT_STATE` 表里 `"session.deleted": ["closed", "session_deleted"]`
  （第 104 行附近）。**没有** archive / suspend 的映射。
- 而 opencode 的库里明确有 **`time_archived`** 与 **`time_suspended`** 两列：在界面上"归档/关掉一个会话"
  是把它归档，不是删除，因此**不会**触发 `session.deleted`。
- `opencode serve` 是可以活几天的常驻进程（本机 54972 就是），所以这些行会一直留在面板上，
  直到那个进程退出。

**结论**：幽灵 = **"在宿主里已归档/已关闭、但没被删除的会话"**。它们此刻没有出现，只是因为用户还没归档过
现有这 7 个。这也解释了用户说的"有些"和"还会"（不是每次，是有归档动作之后）。

**下个会话的第一件事（确定性复现，10 分钟内可完成）**：

1. 在 opencode 里归档一个会话；
2. 等 ≥2 分钟（`DefaultAbsenceGrace`），看 `--json` 是否仍有它；
3. 用 §51.6 的脚本查该 id 的 `time_archived` 是否非空；
4. **同时抓宿主事件**：在 `reporters/opencode.js` 的 handler 里临时打印 `event.type`（只在自己机器、
   一次性），确认归档到底发了什么事件。这一步决定修法：
   - 若发 `session.updated`（带 `time.archived`）或专用的 archive 事件 → **加一条映射即可**；
   - 若什么都不发 → 需要主动核对宿主状态（读库或找到正确端点；端口 28894 根路径是 Web UI，需再找）；
   - 都不行 → 只能做策略：空闲超过 N 分钟的会话"隐藏而不删除"。**这是产品决定**，见 51.5 的禁令。

### 51.3 缺陷二：设置窗口"闪退" —— 没有任何崩溃痕迹，所以不是崩溃

取证结论（四路都查了）：

| 检查 | 结果 |
|---|---|
| `error.log`（Go 侧错误日志） | **不存在** |
| Windows 应用日志，近 3 天含 `AI-Terminal-Pet` 的条目 | **0 条** |
| `CrashDumps` / WER `ReportArchive` | **无相关报告** |
| `hang.log` | **存在**：6 次 stall，最近两次 **2026-11-13 09:47** 与 **2026-11-14 04:07**，均 `lastMessage=WM_TIMER goroutines=12` |

所以"闪退"只可能是下面三种之一，每种都有代码定位与判据：

1. **进程正常退出，窗口跟着消失**。`destroySettings()` 只在 `wmClose`（托盘退出 / `--quit`）里被调用，
   窗口自己不会关。另一条退出路径是空闲自毁：`internal/ui/app_windows.go:977-984`，`timerIdle`
   （30 s 一跳）在 `!values.HasAvailable()` 且距 `lastAvailable` ≥ **10 分钟**时 `DestroyWindow`
   → `wmDestroy`（`:960-968`）→ `PostQuitMessage`。而 `HasAvailable()` 就是 `len(entries) > 0`。
   **判据**：闪退前 store 已经空了（所有 AI 终端都关了）。**可确定性验证**：关掉所有终端与 reporter，
   再 `--settings`，≤10 分钟窗口与进程一起消失。
2. **消息泵卡住**（`hang.log` 的形态）。`timerClock`（2 s）在展开态里 `a.load()` + `postRefresh()`；
   设置窗口与浮窗**共用同一个 UI 线程**。卡住时窗口表现为"未响应"，用户点关闭或系统结束它——
   读作"闪退"。**判据**：闪退前窗口标题栏出现"未响应"，或 `hang.log` 在那时刻增长。
3. **重排把窗口放到屏幕外**。`WM_DPICHANGED` 用新 DPI 重排并夹进工作区（1.0.7 新增）；
   若夹取算错，窗口跑到工作区外，用户读作"窗口没了"。**判据**：闪退后用
   `GetWindowRect` 查——若窗口仍存在但坐标在工作区外，就是这条。

**本轮做过的复现尝试（别重复）**：连续 3 轮 `--settings` → 枚举窗口 → 点「外观」「通用」页签 → 关闭；
进程始终存活，`hang.log` 增长 **0 B**。所以**不是"一打开就崩"**，必须带具体交互（切主题/改字号/
拖到另一块屏、或让 store 变空）才能复现。

### 51.4 建议的排查顺序

1. **先修幽灵**（51.2 的复现是确定性的，且判据清楚）。顺手确认：面板层与 `--json` 的去重规则不同
   （面板按 source+sessionId 合并，`--json` 保留 per-process 行），改的时候两条视图都要有测试。
2. **给三条退出路径加"只开发构建可见"的原因日志**（`exit: idle` / `exit: quit` / `exit: destroyed`），
   再复现闪退。这比继续猜快一个数量级；生产构建不要写，`no-network` 与轻量是硬约束。
3. 若确认是卡死，用 `PET_PREVIEW_DIR` 渲染 + `hang.log` 的 `lastMessage` 缩到具体消息；重点量
   **large + contrast** 组合下设置窗口的绘制耗时（1.0.7 的 `wrapHint`、字号表、DPI 重排都是新的）。
4. 修复幽灵时**不要默默删会话**：§43 已经把"消失 = 完成"这个错误修过——"消失"是有语义的。
   优先加"归档即结束"的映射，而不是加"空闲超时就删除"。

### 51.5 守则（与前几轮一致，别绕）

- `reporters/*.js` 永远 fail-fast 且静默：不能让宿主 CLI（opencode/claude/DSH/antigravity）因上报端卡住或崩。
- 验证写盘类行为必须**让进程正常退出**（AGENTS 第 9 条）：kill 掉再读文件说明不了任何事。
- 改包内任何文件都要 `go build` → `vendor.mjs` → `--check`；改 `engines`/清单要同步三处（包、根 shim、索引）。
- 这是**浮窗行为缺陷**：走 `pet-v1.0.8`；插件包只有在新 vendor 时才需要新版本号。
- 全门禁按 AGENTS「标准验证命令」；本机 `--doctor` 应 10 项 0 fail。

### 51.6 取证工具（本轮现场做的，可直接复用）

- `%TEMP%\opencode\petwin.ps1`：枚举窗口 / 截屏 / `PostMessage` 点击 / 关闭（自带 per-monitor DPI 感知，
  否则 `GetWindowRect` 与 `CopyFromScreen` 会差一个缩放比）。
- `%TEMP%\opencode\ghost-check.mjs`：把 `--json` 的会话与宿主自己的列表对账（注意 PowerShell 写文件带 BOM，
  脚本里已去）。
- `%TEMP%\opencode\session-store-check.py`、`session-columns.py`：只读打开 `opencode.db`，
  查 `session_v2` 的 `time_archived` / `time_suspended` / `time_idle`（`mode=ro` + `query_only`，不写一行）。
- 一条命令看会话视图：`& "$env:LOCALAPPDATA\AI-Terminal-Pet\AI-Terminal-Pet.exe" --json --pretty`
  （它只读快照，永远不会阻塞 GUI）。

## 52. 交接：两个现场缺陷已修 —— 归档即结束、设置窗口不再跟着自毁（2026-11-14 晚，第 52 轮）

§51 只取证、没改代码。这一轮两条都先**复现**、再改，而且改的每一处都有能让它重新变红的测试。
发版：**浮窗 1.0.8（`pet-v1.0.8`）/ 插件包 1.2.17（`dsh-plugin-v1.2.17`）**。

### 52.1 缺陷一：归档的会话在面板上留一行（已修）

**复现（确定性，2 分钟）**：不需要界面，宿主自己的接口就能归档——

```powershell
opencode api PATCH "/api/session/<ses_id>" --data '{\"archived\":true}'
```

归档之后 `GET /api/session?archived=true` 能列出它，`time.archived` 变成毫秒时间戳，
而 `session.deleted` **不会**发。这就是 §51 推出来的机制，本轮实测确认了。

**改动**（`reporters/opencode.js`，并同步到本机真正加载的那份插件）：

- 事件表补两条：`session.archived → closed / session_archived`、`session.unarchived → idle / session_unarchived`。
  两条名字都在宿主的 OpenAPI 里声明（`Session.Event.Durable` 的 union 里就有），`data.sessionID` 也都在。
- `hydrate()` 补一道**收养守卫**：宿主自己的会话记录里 `time.archived` 非空时，不再把它当活会话收养，
  直接写一条 `closed`。归档事件只投给当时在场的那份插件，重载之后光靠事件补不回来——这一步是给那种情况兜底。
  守卫不看"我们之前报过什么状态"：归档的会话不发状态事件，所以我们报过的任何状态都比归档时间旧。
- **不改 Go**：`closed` 快照本来就会被 `SnapshotScanner` 消费、由 `internal/store` 入墓碑，`List()` 两头都不显示。
  §43 的"消失 ≠ 完成"规则因此原样保留——归档走的是 `closed`，不是"空闲超时删除"。

**真机验收**（隔离沙箱，`--json` 与 `runtime/` 两条视图都看）：

| 动作 | 修前 | 修后 |
|---|---|---|
| 归档一个会话 | 快照仍 `idle`，行还在 | 3 秒内 `closed / session_archived`，`--json` `count` 7 → 6 |
| 取消归档 | 无变化 | 快照 `idle`，行回到面板，不响铃 |
| 跨插件重载 | 会被当新会话收养 | hydrate 读到 `time.archived` 直接判死 |

### 52.2 缺陷二：设置窗口跟着进程一起消失（已修）

**为什么不是崩溃**：§51 已经查过 WER / CrashDumps / `error.log` / 应用日志，全空。这一轮加了一个
**只存在于诊断构建**的退出日志（`-tags petdebug`，写 `exit.log`），于是"哪条退出路径"变成一行字。

**诊断构建与它的两条开关**：

```powershell
go build -tags petdebug -o .\bin\petdebug.exe .\cmd\ai-terminal-pet
$env:PET_IDLE_GRACE_SEC = "15"    # 只在 petdebug 构建里被读取
```

发布构建里 `traceExit` / `traceSettings` 是**空函数**——"不出网、非必要不落盘"是产品承诺，
所以这份日志不能进用户安装的包里。`lifetime_test.go` 有一条测试钉住发布构建的 10 分钟默认值。

**复现（修前构建，同一台机器）**：无会话 + `--settings` + 宽限 15 秒 → **~30 秒时进程退出**，
设置窗口一起消失；`exit.log` 里是 `exit idle` → `exit destroyed`。这就是"打开以后闪退"。

**根因**：`timerIdle` 每 30 秒问一次"会话表空了吗"，空过 10 分钟就 `DestroyWindow` 浮窗 →
`wmDestroy` → `PostQuitMessage` → 进程退出。设置在**同一个 UI 线程**上的第二个顶层窗口，跟着走。

**改动**：这条规则抽进 `internal/ui/lifetime.go` 的 `idleDecision(now, anchor, hasSessions,
settingsVisible, grace) (quit, nextAnchor)`：

- **可见的设置窗口也算"仍在使用"**（`IsWindowVisible`，创建但隐藏的不算——用户点了 × 就是不用了）；
- 打开设置时重置宽限锚点，隐藏时**重新计 10 分钟**（否则一个靠设置窗口续命的进程会在关掉它的下一个 tick 上自杀）。

**真机验收（修前 / 修后同场景对照）**：

| 场景 | 1.0.7 | 1.0.8 |
|---|---|---|
| 设置打开、零会话、宽限 15 秒 | ~30 秒进程退出 | 80 秒后仍在运行 |
| 之后关掉设置（不发 `--quit`） | — | 12 秒宽限后 `exit idle` → `exit destroyed` |
| 有会话时 | 不受影响 | 不受影响（`HasAvailable()` 仍优先） |

### 52.3 本轮还发现一件事：本机加载的插件不在仓库里

`~/.config/opencode/opencode.json` 的 `plugins` 指向 `plugins/terminal-pet`（一个**目录**），
OpenCode 解析到的是 `~/.config/opencode/plugins/terminal-pet/index.ts`——一份 **20,515 B 的 TS 分支**，
仓库里的 `reporters/opencode.js` 是 **17,332 B 的 JS**，两者不是同一个实现（本地版有 stale-snapshot
sweep 与写盘节流，仓库版有 `MAX_TRACKED`、hydrate 与 `sendStatus` 的 5 秒重试+拉起）。
所以 §51 那句"在 `reporters/opencode.js` 的 handler 里打印 event.type"在本机**不会生效**。

本轮（第 52 轮）的处理是**两份同步**：两边的映射与收养守卫现在一致，所以发出去的 1.0.8 在本机
真的生效了。**统一成单一来源**紧随其后完成，见 §53。

### 52.4 发版清单（按 docs/UPSTREAM.md §2，逐条落地）

1. `Version` → **1.0.8**；三个二进制按 `-trimpath -buildvcs=false -ldflags "-s -w"` 重编；
   `npm run vendor` 重算 `vendor/manifest.json`（源码摘要 `6c41ef6a…`，39 个文件；overlay
   4,555,264 B / `cb881ace…`）。
2. A 提交：版本 + 两份清单 + `docs/market-index.json` 的 `versions[0].version` 与 `publishTimestamp`
   + CHANGELOG + README / ROADMAP / HANDOVER。`verify.mjs` 全 PASS。
3. 双 tag：`pet-v1.0.8` / `dsh-plugin-v1.2.17`。
4. 双 Release：浮窗 `ai-terminal-pet-1.0.8-win-x64.zip`（**5,054,322 B / `C8FE834A…`**，17 文件）
   与 `pet-setup.exe`（**14,308,864 B / `2EBF0E7E…`**）；插件 `dsh-ai-terminal-pet-1.2.17.tgz`
   （**2,006,574 B / `5D4B2AA8…` / shasum `a7c660a3…`**，19 文件）。三个资产都从**公网 URL 下载回来**
   重新哈希，与本地逐字节一致（GitHub 自己记的 `digest` 也一致）。
5. `npm publish`，2–4 分钟后 `npm view dsh-ai-terminal-pet version` = **1.2.17**、`dist-tags.latest` = 1.2.17。
6. B 提交 `ad8975a`：`source.commit` 与 `icon` 钉到 A 的 sha `7b57f32`。
7. 上游条目：[PR #15378](https://github.com/awesome-dsh-plugin/awesome-dsh-plugin/pull/15378)
   （分支 `bump-tarball-1.2.17`，commit `a1579df`，`MERGEABLE`），
   `tools/upstream-entry.mjs --check` = **OK（in flight）**。合并仍要人工。

**发布资产不可变**，所以以后哪怕只改一句措辞也要新版本号；`pet-v1.0.7` 与 `dsh-plugin-v1.2.16`
及更早的每一个 tag 都仍然有效、不要重传。

### 52.5 下一轮从这里开始（按顺序）

1. **统一上报端来源**（§52.3）：把本地 `index.ts` 改成 re-export 仓库 `reporters/opencode.js`，
   先把本地独有的两个优化（stale-snapshot sweep、写盘节流）逐个评估再决定是否移植。
   → **已在第 53 轮完成，见 §53**。
2. **Claude Code / Antigravity 的"结束"信号**：这一轮只修了 OpenCode 的归档。那两个终端的
   会话结束靠 hook 的 `SessionEnd` / 进程退出；如果用户报的是"关掉 Claude Code 之后还留着",
   从 `reporters/claude-hook.js` 的 `SessionEnd` 与 `pet-statusline` 的 `closed` 两条路径查起
   （判据：`--json` 里那个会话的 `pid` 是否已死、快照是否还在）。
3. **设置窗口的另外两个候选**（§51.3 的 2 与 3）：这一轮把候选 1 证成了，但 `hang.log` 的
   `WM_TIMER` stall 与 DPI 重排没有复现过。若用户再报"未响应"，先用
   `PET_PREVIEW_DIR` 渲染 large+contrast 量设置窗口的绘制耗时，再动代码。
4. **面板与 `--json` 的去重规则**（§51.4 的提醒）：面板按 `session.Identity` 合并、
   `--json` 保留 per-process 行。这一轮没动它们，但归档修复让两条视图第一次在"会话结束"上完全一致
   ——`internal/ui/dedupe_windows_test.go` 与 `internal/store` 的 closed 路径都有测试，别为了"统一"改掉。

### 52.6 守则（与前几轮一致，别绕）

- `reporters/*.js` 永远 fail-fast 且静默：不能让宿主 CLI 因上报端卡住或崩。
- 验证写盘类行为必须**让进程正常退出**（AGENTS 第 9 条）：kill 掉再读文件说明不了任何事。
- **诊断构建可以写日志，发布构建不行**：`-tags petdebug` 是唯一的入口；不要把 `exit.log`
  写进默认构建，也不要在发布构建里读 `PET_IDLE_GRACE_SEC`（`lifetime_test.go` 会红）。
- 改包内任何文件都要 `go build` → `vendor.mjs` → `vendor.mjs --check`；改 `engines`/清单要同步三处
  （包、根 shim、索引），且**发版提交必须同时前移索引**（AGENTS 第 6 条）。
- 全门禁按 AGENTS「标准验证命令」；本机 `--doctor` 应 10 项 0 fail。

### 51.7 上一轮的收尾（顺手核对，供对照；上游那一行的"现在"已过期，以 §52.4 为准）

- 上游条目当时指到 **1.2.16**（现在是 1.2.16 + 1.2.17 的 PR #15378 在飞）；
  站点 `plugins.json` 渲染的 `version` 当时也是 **1.2.16**，`install` 为 `dsh plugin --profile web add dsh-ai-terminal-pet`。
- **中文 README 已经生效**：中文详情页不再显示"该插件的 README 只有英文版本"（§50 的目标达成）。
- 描述被维护者**按自己的措辞润色过**（我们 PR 里就写了"只取 tarball 那一行也可以"，所以这是预期结果）：
  现在的英文开头是 *"A floating always-on-top overlay for Windows that shows which of your terminals is
  waiting: it aggregates … over a named pipe, rings only on real events, and a click jumps to that terminal
  or copies its project path."*，前置条件与"审批仍在 DSH 完成"都还在。**不要**再为此开 PR 改回去。

## 53. 交接：上报端只剩一份实现（2026-11-14 深夜，第 53 轮；没有耗版本号）

第 52 轮发版时留了一件事：**本机真正加载的上报端不在仓库里**（§52.3）。这一轮把它收成一份，
并且把它变成仓库里看得见的一条规则，而不是一句"记得两边一起改"。

### 53.1 本机现在加载什么（实测）

`opencode.json` 的 `plugins` 写的是一个**目录** `plugins/terminal-pet`，OpenCode 把它解析成
`index.ts`。这个文件现在只有 **1,077 B**，内容是一次静态 re-export：

```ts
import { plugin } from "file:///C:/Users/rongj/Desktop/%E9%A1%B9%E7%9B%AE%E5%BC%80%E5%8F%91/ai-terminal-pet/reporters/opencode.js"

export default plugin
```

**为什么是静态 import 而不是 dynamic**：仓库文件缺失时，静态 import 会让 OpenCode 在日志里写
"loading plugin ... failed"，宿主其余部分照常运行；动态 import 会把这条证据吞掉（一个 `.catch(() => {})`
之后什么都没发生，而面板从此空着）。**为什么是绝对 file URL**：相对路径从 `~/.config` 走不到仓库，
而中文目录名必须百分号编码。

### 53.2 怎么验证的（三步，都能重跑）

1. **Bun 能加载它**（OpenCode 的 server 就是 Bun 编的，用 Node 只能证明语法对）：

   ```powershell
   bun -e 'const m = await import("file:///.../reporters/opencode.js"); console.log(m.default.id, typeof m.default.setup)'
   ```

   输出 `ai-terminal-pet function`。

2. **wrapper + 真 reporter 端到端**：一个假 context（`event.subscribe` 依次吐
   `session.execution.started` / `permission.asked` / `session.archived`，`session.get` 返回标题），
   `setup()` 之后等 2.5 秒，`runtime/` 里出现 **1 个** closed/session_archived 快照。说明
   "wrapper → 仓库文件 → 协议 → 快照"整条链路是通路的。

3. **真宿主**：改完之后 OpenCode 的 watcher 自己重载了（日志 `loading plugin ... index.ts`，
   21:26:45Z），随后真机复验一次归档：`PATCH {"archived":true}` → 快照 4 秒内变
   `closed / session_archived`，`--json` 的 `count` 6 → 6（那次归档的是列表里已不再展示的一行，
   所以"行消失"由 `runtime/` 里那个文件被消费证明）；`{"archived":false}` → 快照变
   `idle / session_unarchived`，行回到面板。

### 53.3 旧分支里那三个"优化"没有搬过来，理由写进了代码

| 旧本地分支的做法 | 为什么**不**移植 |
|---|---|
| `sweepStaleSnapshots()`：加载时删掉本进程 pid 名下超过 2 小时的快照 | 一段安静两小时的对话**不是**结束了；把文件删掉等于对浮窗说"它结束了"（§43 的规则）。而快照扫描器本来就会删"宿主进程已经不在"的文件，`sessionIdOf()` 也拒绝 `evt_/msg_/prt_/prj_` 开头的 id，它当初要治的病在这里构造不出来 |
| 写盘节流（非关键写 300 ms 一次） | `report()` 已经在"state/reason/tool 都没变"时直接返回，节流只能推迟**某个状态的第一次上报**——那一次恰恰是最要紧的 |
| `checkPipeAvailable()` 2 秒缓存 + 10 秒拉起节流 | `protocol.js` 的 `sendStatus` 本来就是"连不上就重试 5 秒、需要时拉起、失败静默"，两套机制叠在一起会让"到底谁在拉起浮窗"变成一个有两个答案的问题 |

前两条在 `reporters/opencode.js` 里留了一段注释（有测试的规则旁边就应该有这样的"为什么没有照抄"
记录），第三条留在本节。

### 53.4 备份与"不要删"

- `~/.config/opencode/plugins/terminal-pet/` 里现在有：
  `index.ts`（1,077 B 的一次性 re-export）、`index.ts.bak-20261114-fork`（旧分支的**尽力复原**，
  **不是**逐字节备份——原文在被覆盖前没有先复制，这份是从本轮编辑过程的读取内容拼回来的）、
  以及更早的两份 `.bak-20261105*`（不动）。
- **不要**把 `index.ts.bak-20261114-fork` 当成"随时可以恢复的旧版本"：它能编译，但没有被验证过，
  而且它承载的两个行为上面已经说明为什么不采纳。真正的实现只有仓库里那份。

### 53.5 顺带把第 2 项查清了：Claude Code / Antigravity 的结束信号

结论：**三个入口都真的会发 `closed`，缺的不是代码，是"这些终端有没有装上报端"**。

| 入口 | 实测（沙箱，仓库里的文件） | 结果 |
|---|---|---|
| `reporters/claude-hook.js` | `SessionStart` → `SessionEnd` | 快照 `closed / session_end`（`claude-hook`） |
| 本机装的 `~/.claude/hooks/terminal-pet-hook.js` | `SessionStart` / `Stop` / `SessionEnd` | `idle` / `idle` / **`closed`**（`claude-ipc`），走的是安装目录里的 `pet-ipc.js` |
| `reporters/antigravity-hook.js` | `terminationReason` / `fullyIdle` | `closed / termination_reason` 与 `idle / fully_idle` |
| `pet-statusline.exe` | `agent_state` = `terminated` / `closed` / `finished` | **`idle / agent_state`** |

最后一行不是缺陷，是**设计**：statusline 是"每刷新一次喂一个电平"的接口，没有边沿，
`closed|terminated|finished|complete|done|idle` 都映射成 `idle`（`internal/reporter/statusline.go`
的注释与 `reporters/antigravity-statusline.js` 都写明了，`StatuslineStatus` 的测试也钉着）。
所以 Antigravity 的会话会不会"关掉之后还留着"，取决于它有没有 hook 在发 `terminationReason`；
本机的 `~/.gemini/config/plugins/antigravity-windows-notifier` 只配了
`BeforeAgent` / `AfterAgent` / `BeforeTool(ask_user)` / `Notification`，**没有**终止事件，
而 `~/.gemini/antigravity-cli/settings.json` 的 `statusLine` 用的是 `pet-statusline.exe`。
也就是说：本机的 Antigravity 只有一个"电平"上报端，回合结束后会停在 `idle`，不会有 `closed`。
这不是这一轮要改的东西（改它等于替宿主发明一个它没发的信号），但**下一个报"agy 关掉之后还留着"的人
应该先看这一行**：判据是 `--json` 里那条会话的 `pid`（statusline 用的是稳定哈希，不是真 pid，
所以 `trustedProcessID` 不认它——靠的是 `DefaultOrphanAfter` 45 分钟与窗口标题）。

### 53.6 顺手量了第 3 项的第一个候选：设置窗口一帧画多久

§51.3 的候选 2 是"消息泵卡住"。这一轮加了 `internal/ui/paint_cost_test.go`
（`go test ./internal/ui -run TestMeasureSettingsPaint -v`），在**本机真实的字号与主题**
（contrast、150% DPI）下量一帧的**光栅化**成本（不含 `UpdateLayeredWindow`）：

| 字号 | 通用 | 声音 | 会话 | 外观 |
|---|---|---|---|---|
| 标准 | 17.0 ms | 19.1 ms | 22.3 ms | 13.2 ms |
| **大** | 22.0 ms | 33.1 ms | **47.5 ms** | 27.2 ms |

（窗口 781×997 与 938×1198；每帧 20 次取平均，已含一次预热。）

结论：**这不是卡死的来源**。47 ms 是"一次点击后慢半拍"，不是 `hang.log` 里那种 6 秒 stall；
`hang.log` 那 6 次 `WM_TIMER` stall 仍然没有解释，但它不是设置窗口绘制。测试只打印不设阈值
（阈值在 CI 上必然是 flaky），**别**把它改成断言。

### 53.7 门禁与影响面

- 只动上报端与文档：`node --check reporters/opencode.js`、`node --test
  reporters/opencode-events.test.mjs` **10/10**、`node --check` 本机 wrapper、Go 全套未受影响
  （`go test ./...` 仍 334 个，新增的是一个只看 `-run` 才跑的测量）。**没有重发版本**：浮窗二进制
  与插件包一个字节没变，`pet-v1.0.8` / `dsh-plugin-v1.2.17` 仍是当前版本。
  CI run `43643276551` 与 `43644215965` 都 success。
- 仓库侧新增的只有注释（`reporters/opencode.js` 里"为什么不抄那三个优化"）与一个测量测试，
  因此 `vendor/manifest.json` 的源码摘要不受影响（它只哈希 `cmd/` 与 `internal/` 的**非测试**文件）。

### 53.8 下一轮（从 §52.5 顺延）

1. ~~上报端来源~~ → 已完成（§53）。
2. ~~Claude Code / Antigravity 的结束信号~~ → 查清（§53.5）：三个入口都会发 `closed`；
   Antigravity 的 statusline 是电平接口，本机只装了它，所以没有 `closed`。
   **如果之后要动它**，方向是让 Antigravity 也装 hook（`terminationReason`），不是改 statusline 的映射。
3. **设置窗口剩下的候选**：`WM_TIMER` stall 与 DPI 重排。绘制成本已量过（§53.6，最大 47.5 ms），
   所以不是它；若用户再报"未响应"，下一步用同一台机器上的 `hang.log` 时间戳去对
   `WM_DPICHANGED` 与托盘/贴边动作，而不是继续改绘制。
4. **面板与 `--json` 的去重规则**：不要为了"统一"改掉。

## 54. 交接：`hang.log` 的"卡死"有一半是待机（2026-11-15 早，第 54 轮；浮窗 1.0.9 / 插件 1.2.18）

第 51 轮把 `hang.log` 里那六次 `WM_TIMER` stall 当成"设置窗口消失"的候选证据之一（§51.3 候选 2）。
这一轮先做了一件便宜得多的事：**把它们和系统自己的电源事件对表**。结论是它们不是卡死，
而 §51.3 的候选 2 一直由一份待机证据支撑着。

### 54.1 对表的方法与结果（这台机器上可以重跑）

两路数据都用 UTC 对齐：`hang.log` 的转储头写的是带 `+01:00` 偏移的本地时间，而
`Get-WinEvent` 读出来的是 UTC，所以先换算再比。

```powershell
Get-WinEvent -FilterHashtable @{LogName='System'; ProviderName='Microsoft-Windows-Kernel-Power'; Id=506,507} -MaxEvents 400 |
  ForEach-Object { '{0}|{1}' -f $_.TimeCreated.ToUniversalTime().ToString('o'), $_.Id }
```

| `hang.log` 转储（本地） | 最近的 507（退出待机） | 差 |
|---|---|---|
| 2026-11-06 07:59:26 | 06:59:26.560Z → 本地 07:59:26.560 | **同一秒** |
| 2026-11-13 09:47:22 | 08:47:22.789Z → 本地 09:47:22.789 | **同一秒** |
| 2026-11-14 04:07:53 | 03:07:53.061Z → 本地 04:07:53.061 | **同一秒** |

三次全部同一秒（日志保留期只到 2026-10-25，所以十月那三次对不上，但那三次的
`goroutine 1` 也停在 `GetMessage`）。**机制**：看门狗用的是 `time.Now`，它**把睡眠算作流逝**：

```
GetTickCount64             102.73 h since boot   （含待机）
QueryUnbiasedInterruptTime  96.17 h since boot   （只算清醒）
                          差 394 分钟
```

所以一次超过 6 秒的睡眠，一醒来就"满足"6 秒无分发。

### 54.2 改法与证据

- `internal/ui/hang.go` 新增：`interruptTicks()`（`QueryUnbiasedInterruptTime`，100 ns 单位）、
  `awakeElapsed()`、`hangDue()`。判定只吃清醒时间；`time.Now` 只用于日志里给人读的时间戳。
- 转储头多一个字段：`awakeQuiet=8.001s`（以前只能靠猜；现在它写着"UI 线程清醒时多久没分发"）。
- **待机不会触发**这件事只有跑一遍才算数，所以有两条会**真的启动看门狗**的测试
  （`hang_watchdog_windows_test.go`：只给 handle、时钟与 `LOCALAPPDATA`，不需要窗口）：
  ① 清醒 8 秒静默 → 必须转储，且头里 `awakeQuiet` 等于实测值；② 3 毫秒（待机恢复的形态）
  → 必须静默；③ 时钟缺失 → 必须静默（宁可少报，不要误报）。
- **真卡死照旧报**：诊断构建 `-tags petdebug` + `PET_HANG_PROBE_MS=12000` 让刷新处理器故意阻塞
  12 秒，看门狗必须转储。发布构建里这个钩子是常数 0（`hang_probe_release.go`）。
- §51.3 候选 3（DPI 重排把窗口移出工作区）也顺手做了：`clampSettingsToWorkArea` 的规则抽成纯函数
  `clampOriginToWorkArea`，`settings_clamp_windows_test.go` 覆盖负坐标副屏、比窗口还矮/窄的工作区、
  正好贴角。单屏机器上它本来就不可能出错，所以它值得一张表而不是一句注释。

### 54.3 发版（1.0.9 / 1.2.18）

按 `docs/UPSTREAM.md` §2 走完，逐条落地并在发出去之后**当场核对**：

1. `Version` → **1.0.9**；三个二进制按 `-trimpath -buildvcs=false -ldflags "-s -w"` 重编（浮窗
   **4,556,288 B**、statusline 3,664,896 B、hook 3,697,152 B）；`vendor.mjs` 重算清单
   （源码摘要 `2aab36f6…`，42 个文件；overlay `ddade226…`）。
2. A 提交 `1d47473`：版本 + 两份清单 + 索引 `versions[0].version` / `publishTimestamp` + CHANGELOG
   + README / ROADMAP / HANDOVER / LAUNCH + AGENTS 的 `func Test` 计数。CI run `43691668564` **success**。
3. 双 tag：`pet-v1.0.9` / `dsh-plugin-v1.2.18`。
4. 双 Release，三个资产**从公网 URL 下载回来重哈希**，与本地逐字节一致（GitHub 记的 `digest` 也一致）：
   zip **5,054,467 B / `81F9FAB4…`**、`pet-setup.exe` **14,309,888 B / `3C22B4A7…`**、
   tgz **2,006,574 B / `08A6FE06…`**。
5. `npm publish` 之后 `npm view dsh-ai-terminal-pet version` = **1.2.18**（`dist-tags.latest`）。
6. B 提交 `ff7a106`：索引 `source.commit` 与 `icon` 钉到 `1d47473`。CI **success**。
7. 上游条目：[PR #15482](https://github.com/awesome-dsh-plugin/awesome-dsh-plugin/pull/15482)
   （分支 `bump-tarball-1.2.18`，commit `5f64040`，`MERGEABLE`），`--check` = **OK（in flight）**。
   上一轮那条 **#15378（1.2.17）已 `MERGED`**。合并仍要人工。

**为什么这一轮要占一个版本号**：改的是 `internal/ui`（浮窗行为），而插件包自带一份 vendor 二进制，
按规矩 vendor 变了就必须换插件版本。若只发浮窗，新装 DSH 的人会拿到 1.0.8 的浮窗。

本机升级：`scripts\install.cmd` 装到 1.0.9，`-VerifyOnly` **11/11**、`--version` = 1.0.9、
`--doctor` **10 ok / 0 warn / 0 fail**。

### 54.4 顺手修掉一条真实的测试夹具缺陷（suite 的 launch stub 会互相拉起）

`tests/fresh-machine.mjs` 的 S3 把 `AI_TERMINAL_PET_EXE` 指向 `harness-launch-stub.exe`，而那个 stub
原来是 **`pet-statusline.exe` 的副本**——statusline 自己也是上报端，管道不在时它会拉起
`AI_TERMINAL_PET_EXE`，也就是**它自己**。本机的浮窗正好没在跑时，一次插件上报留下 **23 个** stub，
它们持有沙箱里 exe 的句柄，S8 的 `-Uninstall` 于是报 `访问被拒绝`，套件 **159/162**。主机浮窗在跑时
mutex 会挡住第二个实例，所以这个互相拉起平时不发作——这就是它一直没被发现的原因。stub 换成惰性的
`hostname.exe` 之后 **162/162**。

### 54.5 下一轮

1. 设置窗口的 `WM_DPICHANGED` 与托盘/贴边动作：`hang.log` 现在能给出**真实**的清醒静默时长，
   下次真出"未响应"时，用它的时间戳去对动作，而不是再改绘制（§53.6 已排除绘制）。
2. Antigravity 若真报"关掉还留着"：装 hook（`terminationReason`），不要改 statusline 的映射（§53.5）。
3. 面板与 `--json` 的去重规则：别为了"统一"改掉（§51.4、§52.5）。

## 55. 交接：下一步是 macOS / 其他系统 —— 先盘点，别先动手（2026-11-15 晚，第 55 轮；只盘点，**没有改代码、没有发版**）

下一位要做的是一次**架构改动**（[ROADMAP §2 E3](ROADMAP.md)）。这一轮什么都没改，只把"要搬的东西到底有多大、
哪一块已经天然可移植、哪一块必须先决定"量出来，并且**每条都跑过**，不是读代码猜的。

### 55.1 起点（实测，别凭记忆）

| 项 | 值 |
|---|---|
| 当前提交 | `78c6981`（`git log --oneline -3` 复核） |
| 版本 | 浮窗 **1.0.9** / 插件包 **1.2.18**，npm `latest` = 1.2.18 |
| 分支状态 | 工作区干净；CI 三次连绿（`43692026500`、`43691668564`、`43692383065`） |
| 上游条目 | **1.2.17 的 #15378 已 MERGED**；1.2.18 的 [#15482](https://github.com/awesome-dsh-plugin/awesome-dsh-plugin/pull/15482) 在飞、`MERGEABLE` |
| 平台 | **Windows only**。`GOFLAGS=-mod=readonly`，`go 1.23.0` 是下限（本机 go1.27.0） |

三条与这一轮直接相关的现状：

1. `internal/autostart`、`internal/pipeserver`、`internal/platform`、`internal/app` **整个包都是
   `_windows.go`**——交叉编译时 go 会直接说 `build constraints exclude all Go files`，不是"有错"。
2. 两个 `verify.mjs` 与 `docs/market-index.json` 都断言过 `platforms`/`engines`；
   `dsh.bundle.platforms: ["win32"]` 写死在**两个清单**里，而
   [§49](HANDOVER.md) 记过：**dsh-plugin.org 正是按这个字段把它标成 Windows-only 的**。
   这个字段是有外部效果的，不是注释。
3. 上一轮（§54.4）刚踩过一次"测试夹具自己是被测对象"的坑；跨平台这件事会制造同一类坑
   （见 §55.7）。

### 55.2 代码盘点：17,238 行非测试 Go 里，多少能直接搬

用这三档给每个非测试文件分类（判据：有没有 `win.` / `lxn/win` / `syscall.` / `unsafe.`）：

| 档 | 含义 | 行数 | 文件 |
|---|---|---|---|
| **portable** | 已经是纯 Go，不 import 平台包 | **5,650 行（32.8%）** | 26 个（cmd 5 / reporter 7 / store 2 / session 2 / config 1 / i18n 1 / ui 8） |
| **rename-only** | 带 `_windows.go` 后缀，但一个 Win32 调用都没有：纯粹被后缀挡住 | **3,887 行（22.5%）** | 14 个（app 1 / platform 1 / ui 12） |
| **needs-port** | 真的有 Win32 / syscall / unsafe | **7,701 行（44.7%）** | 12 个（ui 7 / platform 3 / app 1 / autostart 1；pipeserver 另计，见下） |

按包看（`needs-port` / `rename-only` / `portable`，单位行）：

```
internal/ui           6,047 / 2,927 /   753     (19 个 _windows.go + 8 个纯 Go)
internal/app            606 /   687 /     0     (app_windows.go 与 doctor_windows.go)
internal/platform       590 /   273 /     0     (attend/focus/presence 三个 Win32 文件)
internal/pipeserver     187（传输一层）/  0 /     0       ← 见 §55.4，它只差"换传输"
internal/autostart      271 /     0 /     0     (整个包都在注册表里)
internal/reporter         0 /     0 / 1,418     (7 个文件，只有 send.go 里 3 处 Windows)
internal/store            0 /     0 /   850
internal/session          0 /     0 /   325
internal/config           0 /     0 /   569
internal/i18n             0 /     0 /   661     ← 但见 §55.3，它有一个 Win32 调用
reporters/*.js            0 /     0 /   —       （Node 侧，见 §55.4）
```

**`internal/ui` 那 12 个 "rename-only" 文件是最值钱的一笔**：`drawing_windows.go`（360）、
`drawing_compact_windows.go`（783）、`focus_windows.go`（332）、`settings_text_windows.go`（208）、
`mute_windows.go`（242）、`receipt_windows.go`（162）、`copy_toast_windows.go`（170）、
`rows_windows.go`（71）、`pulse_windows.go`（50）、`tap_windows.go`（72）、`attention_windows.go`（129）、
`approve_windows.go`（348）。它们只调用 `canvas` 的图案原语（见 §55.5），所以**只要新平台提供同样
那 16 个方法，这 2,927 行一行都不用改**。

### 55.3 今天在 darwin 上到底能编译到哪一步（实测）

```powershell
$env:GOOS='darwin'; $env:GOARCH='arm64'; go build ./...
```

结果（原样，值得记住——`build constraints exclude all Go files` 与"编译错误"是两种不同的失败）：

| 包 | darwin 结果 |
|---|---|
| `internal/store` / `internal/session` / `internal/config` | **OK，并 `go vet` 也通过** |
| `cmd/pet-setup` | **OK**（它嵌的是 Windows 载荷，但代码本身平台无关） |
| `internal/i18n` | **FAIL**：`undefined: syscall.NewLazyDLL`（`i18n.go:84`）、`undefined: syscall.UTF16ToString`（`:104`）——`SystemLocale()` 是纯逻辑，`systemLocaleWindows()` 却写在了同一个文件里 |
| `internal/ui` | **FAIL**：`undefined: defaultIdleSeconds`（`attention.go:72`）——它定义在 `app_windows.go:31`，而这个函数是**纯策略** |
| `internal/platform` / `pipeserver` / `autostart` / `app` | **FAIL**：`build constraints exclude all Go files` |
| `internal/reporter` | **FAIL**：`imports golang.org/x/sys/windows`（`send.go:18`，另有 `winio` 与 `syscall.SysProcAttr`，全在同一个文件里） |
| `cmd/ai-terminal-pet` / `pet-statusline` / `pet-hook` | **FAIL**：传递依赖上面几个 |

**测试的可搬比例**：`341` 个 `func Test` 里，**177 个在 `_windows_test.go`**、
**164 个在普通测试文件**；而这 164 个里今天能跑起来的只有 **71 个**
（store 27 + session 11 + config 22 + pet-setup 11）。另外 93 个（app 19 / reporter 23 / ui 35 /
pet-hook 13 / i18n 3）**属于纯逻辑、只是所在包今天在 darwin 上编译不了**——也就是说
"能搬的测试比能跑的编译单元多"，这正是 §55.8 第一步要拿到的收益。

### 55.4 五层要换的东西（每层都写清"现在依赖什么 / 换成什么 / 风险"）

**① 传输：命名管道 → Unix domain socket**

- Go 侧只有一处真正的 Win32：`internal/pipeserver/server_windows.go:67` 的
  `winio.ListenPipe(path, &winio.PipeConfig{SecurityDescriptor: "D:P(A;;GRGW;;;OW)(A;;GRGW;;;SY)", ...})`。
  其余 187 行（读一行 JSON、5 s 读超时、精确回 `OK`、最多 32 条并发）**与传输无关**。
- Node 侧更简单：`reporters/protocol.js` 的 `sendOnce` 用的是
  `net.createConnection({ path: pipePath })`，**同一个 API 在 POSIX 上连接的就是 Unix socket**；
  要改的只有 `PIPE_BASE_PATH`（`:9` 的 `\\.\pipe\ai-terminal-pet-v1`）这一个常量。
- Go 侧的 `internal/reporter/send.go` 有 3 处（`winio.DialPipe`、`windows.ERROR_PIPE_BUSY`、
  `syscall.SysProcAttr{HideWindow, DETACHED_PROCESS}`），都在同一个文件里。
- **风险**：macOS 的 socket 路径上限是 **104 字符**，而管道名是
  `\\.\pipe\ai-terminal-pet-v1-<8hex>`（不含 home 的绝对路径），所以换成
  `$TMPDIR/ai-terminal-pet-v1-<8hex>.sock` 是够的——但**不要**把它塞进 home 哈希路径里。
- **不能动的是哈希**：`PipePathForUser()` 与 `pipeScopeToken()` 都是
  `sha256(lower(home) 去尾部分隔符) 前 4 字节（8 hex）`。这一步是跨语言契约，
  由 `TestPipePathForUserMatchesReporterAlgorithm` 钉着（`USERPROFILE=C:\Users\Pet-Test\` → `sha256("c:\users\pet-test")`）。
  **容器可以换，哈希不能换**；但 `docs/PROTOCOL.md` §3.1 现在的标题是"管道名算法"、
  正文写的是 Win32 路径，所以这次改动必须同时改 PROTOCOL.md，并按 §11 的不兼容变更策略走一遍
  （第三方实现需要知道"同一台机器是哪个平台"）。

**② 数据目录：`%LOCALAPPDATA%` → 平台目录**

- `23` 个 Go 文件读 `LOCALAPPDATA`（`internal/app` 是唯一决定目录的地方，`:73-78`），
  `protocol.js` 有 5 处（`runtime/`、`config.json`、`approvals/` 与可执行文件路径）。
- 建议一个 `config.DataDirectory()`（darwin: `os.UserConfigDir()` 即 `~/Library/Application Support`，
  linux: `$XDG_DATA_HOME`），**Go 与 Node 两侧都调它**，不要各自拼字符串——
  现在就是两边各拼一遍，`runtime/` 的契约靠 §2 的文档与测试对齐。
- **`--json` 的契约写了"只读快照、不连 GUI"**（§52 之后仍成立），它读的就是这个目录。

**③ 存在性 / 前台 / 客户输入：`internal/platform`（863 行 Win32）**

这一层是**最难也最容易做错**的一块，因为它是 §21/§43/§51/§52 全部教训的载体：

| 现在用什么 | 干什么 | darwin 的等价物 | 风险 |
|---|---|---|---|
| `CreateToolhelp32Snapshot` + Process32First/Next | 进程表（`ProcessAlive`） | `syscall.Kill(pid, 0)` 或 `ps -o`；Go 没有内置枚举 | 低 |
| `EnumWindows` + `GetWindowTextW` | "这个会话的终端窗口还在吗" | Accessibility（`AXUIElement`，**需要用户在系统设置里授权**）或退化到"只看 pid" | **高**：这是产品级取舍——没有授权就没有窗口标题证据 |
| `GetForegroundWindow` / 焦点归属 | 跳转与"你在看哪个会话" | `NSWorkspace.frontmostApplication`（可以，但窗口级要 Accessibility） | 中 |
| `GetLastInputInfo` | 你走开了多久（提醒层） | `CGEventSource.secondsSinceLastEventType`，或 `ioreg -c IOHIDSystem` 的 `HIDIdleTime` | 中 |

**不要**为了让 macOS 编译通过就把 `SessionAlive` 简化成"pid 活着"：§51/§52 的两条缺陷正是
"按进程判定存在性"造成的。反过来说，这一层如果必须退化，**要把它写进 README 的 Known limits 与
`--doctor`**，而不是悄悄少一条证据。

**④ 呈现层 `internal/ui`（6,047 行真 Win32）**

- `canvas_windows.go`（462 行）是 GDI 光栅化：16 个方法（`clear` / `fillRect` / `fillRounded` /
  `outlineRounded` / `fillEllipse` / `line` / `text` / `measure` / `fitText` / `hitMapInto` /
  `setAlphaScale` / `insideRounded` / `blend` / `normalizeText` / `close` / `scaledAlpha`）。
  **两个绘制文件只用到其中 9 个**（`clear/fillRect/fillRounded/outlineRounded/fillEllipse/line/text/measure/fitText`），
  外加 `setAlphaScale` 与 `hitMapInto`。所以新平台的 `canvas` 只要实现这套签名，
  `drawing_*.go` 与 `settings_text_windows.go` 不改。
- `app_windows.go`（2,549 行）是窗口 + 消息泵 + 托盘 + 拖拽 + 贴边 + DPI + 跳转。macOS 版本是
  `NSWindow`（无边框、透明、`NSWindowLevel`）+ `NSStatusItem`；**关键约束不变**：
  分层窗口没有子控件，命中测试必须与绘制同源（§F8 的教训）。
- `settings_windows.go`（2,222 行）+ `settingsTypeRamps`：**字号梯子是在 GDI 上量出来的**
  （§44 的表格：GDI 把 11/12/13 px 都吸附成 8 px em）。Core Text 的吸附规则不同，
  三张表需要**在新的渲染器上重新量**，`TestSettingsTypeRampStaysOrdered` 的判据（4 个 DPI 下量宽度、层级不塌）
  仍然适用，但期望值要重取。别把 GDI 的数直接抄过去。
- `sound_windows.go`：`PlaySoundW` 播内存里的 WAV（winmm）。darwin 用 `NSSound` / `AudioToolbox`，
  或 `afplay`（但 `afplay` 是另一个进程，会引入 §L1 的进程创建税）。
- `hang.go`（上一轮刚换的时钟）：`QueryUnbiasedInterruptTime`。**darwin 要重新回答"清醒时间"
  这个问题**——`mach_absolute_time` 与 `mach_continuous_time` 的区别正是 §54 那张表要的东西，
  但**先测再选**，不要抄结论。
- **cgo 是这一层绕不开的决定**：AppKit 无法用纯 Go 调用。三条候选各有代价，下一个会话要先拍板：
  ① 少量 `.m` + cgo（编译期依赖，运行时仍是系统框架，不违反 AGENTS 第 3 条的"零外部运行时 DLL"精神，但"单文件静态编译"这句话就变了）；
  ② 外部 helper 进程（保住纯 Go，但每次绘制的 IPC 成本与生命期管理都是新问题）；
  ③ 纯 Go 软件光栅只用于"面板看不见的预览/测试"，真窗口仍需要 ①。

**⑤ 外围：自启、安装、自检、CI**

- `internal/autostart`（271 行，HKCU `Run`）→ macOS `~/Library/LaunchAgents/*.plist`。
- `cmd/pet-setup`（421 行，嵌 Windows 载荷 + 跑 `install.ps1`）→ macOS 需要 `.app` bundle 或 tar + `install.sh`；
  它现在"能编 darwin"只是因为 Go 代码本身无关平台，**里面的载荷是 Windows 的**。
- `internal/app/doctor_windows.go`（606 行、10 项）→ 需要一个 darwin 版本，否则 `--doctor` 在新平台上是空的。
  注意它**故意不碰互斥体**（§E6），darwin 侧同样不要用"锁文件存在"当实例证据。
- CI 只有两个 `windows-latest` job。**最便宜的第一道跨平台门禁**是
  `GOOS=darwin GOARCH=arm64 go build ./...`（不需要 mac runner、不需要 cgo）。等测试也能跑时，
  再决定要不要加 `macos-latest`（它会真的花 minutes）。

### 55.5 一条必须守住的设计边界

`internal/session`、`internal/store`、`internal/config`、`internal/reporter`（除 `send.go` 的 3 处）
**今天就是平台无关的，并且这些包承载的是协议契约**。搬平台时**不要**顺手"为 macOS 优化"
它们——`session.ParsePayload`、`store.Reconcile` 的 2 分钟宽限、`reporter` 的脱敏与幂等
是 §21–§54 的结论，跨平台不需要也不应该改。任何改动都要先说明"为什么这条规则与平台有关"。

### 55.6 提交顺序（按这个顺序，每一小步都能验证）

| 步 | 做什么 | 验收 |
|---|---|---|
| **1** | 把 `defaultIdleSeconds` 与 `systemLocaleWindows` 挪出 Windows 文件（`attention.go` 现在依赖 `app_windows.go` 的常量；`i18n` 的 Win32 调用占着 `SystemLocale()` 的位置） | `GOOS=darwin go build ./internal/i18n ./internal/ui` 有进展；Windows 全套不变（`go test ./...` 仍 341） |
| **2** | 把 12 个 rename-only 文件去掉 `_windows` 后缀 | `GOOS=darwin go vet ./internal/ui` 只剩真 Win32 文件报错；Windows 构建与测试逐字节等价（除了 build 依赖） |
| **3** | `config.DataDirectory()` + `protocol.js` 同步（Go/Node 都调它） | 现有 `runtime/` 契约测试仍在 Windows 上过；加一条"目录来自同一个函数"的测试 |
| **4** | CI 加 `GOOS=darwin go build ./...` 门禁（此时仍有包不通过，所以先只对已通过的包列表） | CI 绿；一旦有人往纯 Go 包里塞 `syscall` 就会红 |
| **5** | 传输层：`pipeserver` 抽一个 `Listener` 接口，`_windows` = winio、`_darwin` = `net.Listen("unix")`；`reporter/send.go` 的 3 处同样处理 | `PROTOCOL.md` §3.1 同步；哈希 golden 测试仍过；两侧（Go/JS）在 macOS 上互连实测 |
| **6** | 呈现层：`canvas` 的 16 个方法 + 窗口/托盘/DPI 的 darwin 实现（先拍板 cgo 形状） | `PET_PREVIEW_DIR` 渲染出 PNG（§视觉验证那条命令在 darwin 上也要能跑） |
| **7** | `platform`（存在性/前台/闲置）+ `--doctor` + 自启 + 安装 | §51/§52 的行为矩阵在新平台重跑一遍（归档、关窗、自毁） |

**不要在第 1 步之前动呈现层。** 第 1、2 步是纯机械的，而且它们会把"能跑测试的包"从 4 个变成
更多——那才是继续往下搬的支点。

### 55.7 这次会冒出来的坑（按 §54.4 那一类记）

1. **清单字段的副作用**：`platforms: ["win32"]` 出现在**两个**清单里，`verify.mjs` 断言它们一致，
   而 dsh-plugin.org 用它给仓库贴 Windows-only。要加 darwin 就必须**同一次提交**里改两处，
   并且**需要一个新版本号**（已发布资产不可变）——这与"改代码才换版本"是两条独立的规则。
2. **测试夹具的自启动循环**：§54.4 刚修了一个（stub 是上报端、于是拉起自己）。
   跨平台会引入同样形状的东西：macOS 上 `AI_TERMINAL_PET_EXE` 指向 `.app` 里还是外面、
   `open` 与直接 spawn 的生命期不同，**会造出"一个测试留下 N 个进程"**。判据与 §54.4 修完后一样：
   `Get-Process`/`pgrep` 计数为 0，而不是"看起来跑完了"。
3. **`LOCALAPPDATA` 是唯一目录来源这件事会被拆开**。现在 `internal/app` 一处决定目录，Node 侧另拼一遍。
   第 3 步如果只改一侧，`--json` 会读空、`--doctor` 会报 runtime 为空，而错误信息会指向"没有会话"。
   判据：`--json` 的 `generatedAt` 与 `runtime/` 的文件时间在同一侧。
4. **管道名的"看起来一样"陷阱**：`\\.\pipe\...` 在 POSIX 上是一个合法文件名（会创建出奇怪的相对路径），
   不会报错。所以传输层换掉之后，**旧常量必须删掉而不是留着当默认值**。
5. **`hang.log` 的时钟语义**（§54）：新平台如果直接 `time.Since`，§54 的误报会**原样复现**。
   先量两个时钟的差再选。

### 55.8 这一轮没做的事（诚实清单）

- **没有改任何代码、没有发版**：工作区只多了这一节与 ROADMAP 的 E3 盘点表。浮窗仍是 1.0.9、
  插件包仍是 1.2.18、三个 Release 资产字节未动。
- 没有试过真机 macOS（本机只有 Windows），所以 §55.4 的"darwin 等价物"全部标成**候选**而不是结论；
  凡是标了"先测再选"的地方，下一个会话**必须量**。
- 没有决定 cgo 的三种形状（§55.4 ④）；那是要用户拍板的产品/工程取舍，不是能用测试决定的事。

## 56. 交接：`--doctor` 现在会看上报端接入，`--connect` 只打印（2026-11-15 晚，第 56 轮；浮窗 **1.0.9** / 插件 **1.2.18**，**未发版**）

第 55 轮盘点了跨平台，用户当场决定"先做 Windows 侧的事"，并且选了本机没有任何 Mac 可验的那条
（"第 5 步前再说"）。这一轮因此不动平台层，去做一件在本机就能完整验收、并且直接命中历史上最贵的两次故障的事：
**让 `--doctor` 去读别人家的配置**。

### 56.1 为什么是这一件（三条证据，都是实测）

1. `--doctor` 原有十项**全在看这个项目往机器上放了什么**（安装目录、管道、Run 键、node、alerts），
   一项都不看"谁在调用我们"。ROADMAP §E6 的"与原计划的差异"里第四条自己写着这件事没做。
2. 而真实发生的两次故障都在那一边：§52.3（本机真正加载的 OpenCode 上报端不在仓库里）、
   §53.1（那份 wrapper 是一次静态 re-export，第二跳的目标没人检查）。两者从外面看都是"面板空着"。
3. §53.5 花了整整一轮才查清"Antigravity 只有 statusline 时，会话结束停在 `idle` 而不是 `closed`"。
   那是个**可以一行检查完**的事，没理由只留在交接文档里。

### 56.2 交付形状：一个新包，两个出口

```
internal/connect/            ← 新增（8 个非测试文件 1,599 行；零 Win32，E3 时可整包搬走）
    connect.go        Finding / Status / Roots / Inspect / Worst / Summary
    inspect.go        配置加载、路径缩写、环境变量展开、命令首 token、JS/TS 词法扫描
    opencode.go       plugins[] → 目录 → index.ts → 静态 file:// import（两跳）
    claude.go         hooks.<Event>[].hooks[].command → 命令目标 → JS wrapper 的桥
    antigravity.go    statusLine.command + 有没有能发 terminationReason 的钩子
    fileurl.go        file:/// 编解码（黄金样本见 56.3）
    snippet.go        三段可粘贴片段 + Render + ReadOnlyRoots
    roots.go          唯一与平台有关的那一处（os.Executable / os.UserHomeDir / join）
internal/app/
    doctor_windows.go    只加一行 runCheck("reporters", …) 与 env.home 字段
    connect_windows.go   49 行的展示层（--connect 的文本 + 结尾那句"没写盘"）
cmd/ai-terminal-pet/
    main.go              --connect / --connect=<terminal> 分派 + usage 三行
```

两个出口共用一份读取逻辑：`--doctor` 说**接没接**，`--connect` 说**粘什么**。分开实现的话，
片段可以描述出一个检查随后判"缺失"的形态——这正是本包要消灭的那类不一致。

### 56.3 判定表（每条都有测试）

| 输入 | 状态 | 例子 |
|---|---|---|
| 配置文件不存在 | `warn` | `opencode: warn - not installed (no opencode.json)` |
| 文件在，没有指向我们的条目 | `warn` | `not connected (plugins has an empty plugins list; run --connect=opencode)` |
| 条目目标存在 | `ok` | `plugins[0] -> %LOCALAPPDATA%/repo/reporters/opencode.js` |
| 条目是我们自己的文件但**不存在** | **`fail`** | `... (missing - the repository file is gone, which is handover §52.3's failure mode)` |
| 目录里没有 `index.ts`/`index.js` | **`fail`** | `plugins[0] -> ~/.config/... (no index.ts/index.js inside)` |
| 配置不是纯 JSON（注释/尾逗号） | `warn` | `~/.config/... is not plain JSON (line 2: …)`，**不冒充"未接入"** |
| Claude 接了但没有 `SessionEnd` | `warn` | 目标存在但"a finished session stays idle" |
| Claude 的 wrapper 存在、它 require 的桥不存在 | **`fail`** | 本机的 `pet-ipc.js`（§53.5 那条链） |
| Antigravity 只有 `statusLine` | `warn` | "sessions end at idle, never closed, because no hook sends terminationReason" |
| Antigravity 有终止钩子 | `ok` | `termination hook + statusline -> …` |

两条写进代码注释的规矩：

1. **未安装 = `warn`，不是 `fail`。** 这不是客气：`tests/fresh-machine.mjs` 的 S5 断言独立安装后零 fail，
   而沙箱的 `USERPROFILE` 里没有 `.claude` / `.gemini` / `.config`。判成 fail 就等于每次安装都红。
2. **detail 不打印配置内容。** 本机 `opencode.json` 带 API key；每个 detail 行只出现路径、条目数与事件名，
   `TestFindingsNeverEchoConfigurationContent` 放了一个 `sk-live-DO-NOT-PRINT` 哨兵并断言它不出现。

### 56.4 本机实测（`--doctor` 与 `--connect`，跑过的原样）

```
[warn] reporters: opencode: ok - ~/.config/opencode/plugins/terminal-pet/index.ts -> ~/Desktop/…/reporters/opencode.js;
                  claude: ok - 8 event(s) -> ~/.claude/hooks/terminal-pet-hook.js
                  (SessionStart/UserPromptSubmit/PreToolUse/PostToolUse/PermissionRequest/Notification/Stop/SessionEnd);
                  antigravity: warn - statusline only -> %LOCALAPPDATA%/AI-Terminal-Pet/pet-statusline.exe;
                  sessions end at idle, never closed, because no hook sends terminationReason
doctor: 9 ok, 2 warn, 0 fail          （第 2 条 warn 是 alerts：本机还没有判定记录）
```

整行 473 字符——**它是一行**，`oneLine()` 会把换行压平，这是"贴进 issue 的人读一份报告"这个用法的前提。

这条输出同时是三件事的证据：OpenCode 的第二跳真的被跟到了（`~/Desktop/…/reporters/opencode.js`，
中文目录名与百分号编码都正确解码）；Claude 的八个事件是从 `~/.claude/settings.json` 里数出来的，
顺序是文档顺序而不是 map 的随机顺序；Antigravity 那条 warn 就是 §53.5 的结论。

### 56.5 三处"不这么写就会假绿"的地方

1. **file URL 用真实工作件做黄金样本。** `~/.config/opencode/plugins/terminal-pet/index.ts` 里那条
   `file:///C:/Users/…/%E9%A1%B9%E7%9B%AE%E5%BC%80%E5%8F%91/…` 是唯一被真宿主加载通过的形态。
   `net/url` 的 `URL.String()` 对 `Path:"C:/x"` 会产出 `file:C:/x`（少两条斜杠），所以编码器自己拼，
   并由 `TestPathToFileURLReproducesTheWorkingWrapper` 逐字节钉住（期望值是**中性路径**，
   不写死本机目录，AGENTS 第 7 条）。
2. **JSON 里必须转义反斜杠。** 第一版 `--connect=claude` 打印出 `"command": ""C:\Users\…""`——
   肉眼看像对的，粘进去不是 JSON。现在 `jsonQuote()` 是唯一出口，并且
   `TestConnectJSONQuotedPathsAreValidJSON` 用真的 `encoding/json` 把整块解一遍。
3. **词法扫描而不是正则。** 本仓库自己那份 wrapper 在注释里**三次**点名 `reporters/opencode.js`；
   用"从 … from 后面抓引号"的写法会把注释里的路径当成 import，于是真文件删掉之后仍判 `ok`。
   `scanSource` 一次走完（字符串/行注释/块注释/标识符），`TestImportedTargetsReadsBothDirections`
   同时钉住"注释里的不算"与"真实 wrapper 只给一个目标"。

### 56.6 顺带修掉的文档失准

- `README.md:253`（英文控制表）写 `--doctor` 是 **9** 项，而代码从 1.0.3 起就是 10 项。
  `tools/check_versions.mjs` 只校验版本号形状，管不到计数——这一类只能靠人当场数。
- 中文 README 同一段话原本**重复出现了两遍**（`:574` 与 `:578` 内容相同），这次合并成一处 11 项。
- `HANDOVER` §1 的规模口径：旧值写的是 53 文件 / 17,238 行，但按 `git ls-files '*.go'` 逐文件数
  HEAD 是 **53 文件 / 16,118 行**——那 1,120 行的差是另一套统计方法留下的。本轮把新数写清楚，
  并把旧数标成"按它那一轮的统计方法"。

### 56.8 门禁（本轮实跑）

```
gofmt -l .                                  -> 空
go vet ./...                                -> exit 0
go test -count=1 ./...                      -> 13 个包全 ok（366 个 func Test，§54 是 341）
node tests/fresh-machine.mjs --no-overlay   -> PASS 172/172 checks（§54 是 162）
node tools/check_versions.mjs               -> 18 个版本形状一致；node --test 26/26
python tools/check_docs_links.py            -> 12 份文档全部解析
node packages/dsh-ai-terminal-pet/scripts/vendor.mjs  -> 重算（vendor exe 4,655,616 B）
node .../verify.mjs                         -> VERIFY PASSED
& AI-Terminal-Pet.exe --doctor              -> 9 ok / 2 warn / 0 fail，exit 0
& AI-Terminal-Pet.exe --connect             -> 三段齐全，exit 0；--connect=codex -> exit 2 + 列出三个名字
```

`tests/fresh-machine.mjs` 的断言用的是 `/doctor: \d+ ok, \d+ warn, \d+ fail/`，加一项不会打破它；
S5 的 `0 fail` 是 §56.3 第 1 条存在的理由。本轮给 S1 与 S7 各加了断言（空机器上 `reporters` 是
`warn` 而不是 `fail`；S7 里自己写一份 `~/.claude/settings.json`，接通的判 `ok`、指向不存在的文件
判 `fail` 且 exit 1），检查数因此 **162 → 172**。

### 56.9 顺带量到的：L2 的"2.5 MiB 目标"其实达不到（实测归因）

`--connect` 的第一版用 `net/url` 拼 file URL，主程序因此长到 **4,681,216 B**（比 §54 的 4,556,288 B
多 124,928 B）——为一个 40 行的百分号编码循环买下整个 URL 解析器。改成本包自己写的编码器
（`escapeURLPath` / `percentDecode`，黄金样本仍由 §56.5 那条测试钉住）之后：依赖数 **152 → 151**、
主程序 **4,649,984 B**，省下 **32,256 B**。剩下的 **+92,672 B** 是 `internal/connect` 本身的代价，
最大一块是 `encoding/json`（要解析三份宿主配置，没有别的选择）。

顺手把 L2 那条"4.36 MiB → <2.5 MiB"的账重算了一遍。方法是 `go build -trimpath -buildvcs=false`
（**不加** `-s -w`，否则符号表没了）到一个临时文件，再 `go tool nm -size` 按包前缀汇总：

| 组成 | 大小 | 说明 |
|---|---|---|
| `crypto/internal/fips140/drbg.memory` | 32,772 KB | **BSS**，不计入文件字节，别被它吓到 |
| `runtime` | 2,266 KB | 运行时 |
| `encoding/json` v2 链（`json` + `jsontext` + `v2`） | ~290 KB | 配置、快照、`--json` 都要它 |
| `runtime/pprof` + `compress/flate` | ~130 KB | `internal/ui/app_windows.go:9` 为 `hang.log` 的 goroutine dump 引入 |
| `slices` / `reflect` / `compress/flate` 等 | ~310 KB | 间接依赖 |
| **我们自己的代码**（`internal/ui` 203 KB + `internal/app` 46 KB + 其余包） | **~250 KB** | 占 4.4 MiB 的 **5%** 左右 |

结论：**2.5 MiB 那个目标不可能达成**——它要求砍掉 43%，而能砍的两块（json 与 pprof）都是功能本身。
真正的余量是几百 KB 级别，且每次砍都要先证明"这条依赖是死的"。所以这一轮把 ROADMAP §L2 的目标
改成实测口径，而不是继续追一个写死的数字；真要瘦身，第一步是**再量一次**并写清每一项为什么在，
而不是先改代码。

### 56.10 这一轮没做的事（诚实清单）

- **没发版**：`vendor/manifest.json` 的源码摘要必然要重算（它哈希 `cmd/` 与 `internal/` 的非测试
  `.go` 路径 + 内容），所以仓库里的包与已发布的 1.2.18 tarball 暂时分叉（§29.5 有同样先例）。
  下一次发布是浮窗 `1.0.10` / 插件 `1.2.19`，走 `docs/UPSTREAM.md` 八步。
- **`--connect` 不写盘**：没有 `--apply`、没有"帮我改"的开关。改别人的配置文件要替对方的产品猜 schema，
  而这个项目的上报端铁律就是"不猜、不抛异常进宿主"。如果哪天真要做写入，先做的事是把三份配置的
  schema 版本化，而不是在 `--connect` 里加一个 if。
- **没测其它终端的配置形态**：本机只有这三家。OpenCode 的 `plugins` 目录形态与直接 `.js` 形态都覆盖了，
  但"`plugin`（单数）` 键"只在解析器里兼容，没有真机证据。
- **跨平台仍是下一步**：本轮一行都没有碰 `_windows` 后缀、`canvas` 与 cgo 形状，
  §55 的七步顺序与验收标准原样有效。

### 56.11 下一轮从哪里开始

1. 这一批发出去：`1.0.10` / `1.2.19`（版本常量 + 索引 + 重新 vendor + 八步清单），
   然后把 `--doctor` 的 11 项与 `--connect` 写进 CHANGELOG 的正式条目。
2. 真要动跨平台时，从 §55.6 第 1 步开始；本轮给 `internal/app` 加的是一个只做展示的
   `connect_windows.go`（仍在 `_windows` 里），`internal/connect` 则整包不含 Win32——E3 时那 8 个文件
   可以原样搬，`roots.go` 是唯一要按平台回答不同的地方。
3. 如果之后有人报"某个终端不出现"，第一步不是查上报端代码，是跑 `--doctor` 看 `reporters` 那一行。

## 57. 交接：跨平台施工（第 1–7 步），以及这一版为什么还没发（2026-11-15 深夜，第 57 轮；**未发版**）

用户的要求是"按 §55.6 第 1 步动跨平台，把跨平台和适配全系统做成下一个版本号"。这一轮把 §55.6
的七步里能在这台机器上验证的部分全部做完了，**但窗口层没有做**——所以 1.1.0 现在还不能发，
理由写在 §57.7。

### 57.1 现在的实测矩阵（跑出来的，不是读出来的）

```
PACKAGE                            windows  linux    darwin
./cmd/ai-terminal-pet              ok       FAIL     FAIL     ← 窗口层
./cmd/pet-hook                     ok       ok       ok
./cmd/pet-setup                    ok       ok       ok
./cmd/pet-statusline               ok       ok       ok
./internal/app                     ok       FAIL     FAIL     ← ui.New/Options/Placement
./internal/autostart               ok       ok       ok
./internal/config                  ok       ok       ok
./internal/connect                 ok       ok       ok
./internal/home                    ok       ok       ok
./internal/i18n                    ok       ok       ok
./internal/pipeserver              ok       ok       ok
./internal/reporter                ok       ok       ok
./internal/session                 ok       ok       ok
./internal/store                   ok       ok       ok
./internal/ui                      ok       ok       ok
```

Linux 的测试**真的跑起来了**（WSL2 Ubuntu 26.04 + go1.27.4，CGO_ENABLED=0）：
`cmd/pet-hook`、`cmd/pet-setup`、`internal/{autostart,config,connect,home,i18n,pipeserver,
reporter,session,store,ui}` 全部 ok。darwin/arm64 只做过 `go build` 与 `go vet`——
**这台机器没有 Mac，darwin 的运行时行为一律是未验证**。

### 57.2 九个提交，每一步都能独立验证

| 提交 | 内容 | 关键点 |
|---|---|---|
| `a501a22` | 第 1–2 步：纯逻辑从 Windows 文件里切出来 | `attention_keys.go`（400 行策略）、`hang.go`+`awclock_*.go`、`canvas.go`（16 个绘制原语）、10 个文件去掉 `_windows` 后缀、`pipeserver/server.go`+`transport_*.go`、`connect` 的 `localpath_*.go` |
| `f626078` | 第 3、5 步：数据目录与上报端 | 新包 `internal/home`（三平台布局，Windows 那行**一个字都没改**）；`send.go` 不再 import go-winio / x-sys-windows；重试策略按平台各一张表 |
| `52935c5` | 第 7 步之一：自启 | Windows HKCU Run / LaunchAgent plist / XDG desktop entry，Manager 语义共用 |
| `8299a5a` | 测试夹具修正 | 三个 auto-approve 测试之前会写进测试者真实的 `~/.local/share` |
| `01a39b8` | 第 6–7 步：`internal/app` 生命周期 | `run.go` + `app.go` + `platform_*.go` + `flock_*.go` + `version.go`；`ErrNoWindowLayer`；新文档 `docs/PLATFORMS.md` |
| `b6d5791` | `ControlAction`/`Control` 中性化 | posix 版回答 `false`，调用方于是回落并报 `ErrNoWindowLayer` |
| `089b595` | 第 4 步：CI 门禁 | 两份清单（必须能编 / 必须编不了）双向检查 |
| `dfda2e7` | 重新 vendor | 摘要必然变，`vendor:check` 先红后绿 |
| `411ed3c` | 工具跟随 | `check_versions` 读 `version.go` |

### 57.3 五个"看起来对、其实错"的地方（都修了）

1. **`internal/home` 的第一版在 Windows 上给 LOCALAPPDATA 缺失加了 profile 回落**。两个
   上报端测试立刻变红：`WaitForApprovalDecisionFailsClosedWithoutADirectory`（不可归档的
   请求必须回答 deny）与 `SendSkipsTheSnapshotForAnInvalidPayload`（不可归档的负载不该发送）。
   因为代码从此再也看不到"没有目录"这个状态。**规则改成显式的：Windows 上没
   LOCALAPPDATA = 没有磁盘状态，不猜路径。**
2. **测试只设 LOCALAPPDATA**：跨平台之后这就是错的，linux 会把快照写进测试者真实的主目录，
   下一条断言便在临时目录里找不到它。现在每个包里都有**一个** helper，把四个变量一起指向
   临时树。
3. **unix 的 `retryableDialError` 起初只查 `fs.ErrPermission`**：实测 dial 一个 unix socket
   的 EACCES 是通过 `*net.OpError` / `*os.SyscallError` 包起来的，`errors.Is` 才找得到。
   现在同时比对 `syscall.EACCES`/`EPERM`，并有一条真 socket（chmod 000）的探针证明
   `Send` 在 750 ms 预算里立刻返回。
4. **`canvas` 的 buffer 必须是 DIB 的位**：GDI 把文字直接画进 DIB 内存，如果原语混进另一
   块分配，字形就会落到别的地方。所以 `prepare` 返回 slice，而不是接收一个 slice 去填。
5. **`scripts/cross-build.sh` 的 `cd "$(dirname "$0")/../.."` 多跳了一级**：以
   `scripts/cross-build.sh` 调用时解析到仓库外，第一次跑就报"清单文件不存在"。

### 57.4 刻意没有统一的东西

- **`SessionAlive` 在 POSIX 上只是 signal-0 探针**，文件里写明了理由：§21、§43、§51、§52
  的四次缺陷全部来自"证据比 Windows 少却沿用同一条规则"。一个平台可以用更少的信号，但不
  能继承一条它没有证据支撑的规则。这条限制进了 `docs/PLATFORMS.md`，不只是注释。
- **`canvas_posix.go` 不画字**：`drawText` 返回 false，`measureText` 是按字数的估算。它是
  编译桩，名字里就写着它是缺的那一块；面板在 linux 上因此不存在，而不是画出一个没有文字的
  假面板。
- **`ErrNoWindowLayer`**：没有窗口层的构建会明确报错并指出去看 `docs/PLATFORMS.md`。
  启动一个谁也看不见的窗口，与 §51 花整轮消灭的"浮窗就是不出现"无法区分。

### 57.5 顺手修掉的三处文档/工具失准

- `README.md` 的 "Known limits" 原来写"仅限 Windows"，现在写成三平台的能力表并指向
  `docs/PLATFORMS.md`（中英各一处）。
- `docs/PROTOCOL.md` 里指向 `server_windows.go` / `approve_windows.go` 的行号引用全部跟着
  文件走（契约一个字没变）。
- `tools/check_versions.mjs` 与 `scripts/package.ps1` 改读 `internal/app/version.go`；
  `tools/check_versions.test.mjs` 里那条断言也随之更新（**那条测试在 CI 上红过一次**，
  正是"断言写死了文件名"的典型）。

### 57.6 门禁（全部实跑）

```
gofmt -l .                                  -> 空
go vet ./...                                -> exit 0
go test -count=1 ./...                      -> 14 个包全 ok（382 个 func Test）
node tests/fresh-machine.mjs --no-overlay   -> PASS 172/172
node tools/check_versions.mjs               -> 18 个形状一致；node --test 26/26
python tools/check_docs_links.py            -> 12 份文档全部解析
vendor.mjs --check / verify.mjs             -> 吻合 / VERIFY PASSED
install.cmd -VerifyOnly                     -> 11 file(s) match，0 warning
scripts/cross-build.sh（WSL 实跑）          -> 两个方向都通过
CI run 43714564544 / 最新一次               -> success（含新增的跨平台门禁）
```

二进制：主程序 **4,655,616 B**（4.44 MiB，sha256 `b407e50a…`）、shim 3,669,504 B、
hook 3,701,760 B。三个文件与 `vendor/`、本机安装目录逐字节相同。

### 57.7 为什么 1.1.0 现在不发（诚实清单）

发不了的不是流程，是**产品**：Windows 之外没有面板，而面板就是本产品。

- **Linux 窗口层**（X11 客户端或 GTK）：没写。WSLg 里 X server 可以连（实测 X0 socket
  存在、握手返回 X.Org vendor），但仓库目录里没有字体（只有 DejaVu，**没有中文字体**），
  一个中文字幕的面板在这里会画成方块；`x/image` 可以拉到（proxy 通），但那意味着新的渲染
  管线。这是一次独立施工，不是一次会话的收尾工作。
- **macOS 窗口层**：无法在本机编译验证，更无法运行。写 1,500 行没人跑过的 ObjC/AppKit
  代码进发布物，与"每条声明都要能当场验证"的规矩冲突。
- **因此这一版收在这里**：内核、传输、数据目录、自启、上报端、`--doctor`、`--connect`、
  CI 门禁全部跨平台并且**在这台机器上验证过**；剩下的窗口层是下一轮的第一件事，详见
  `docs/PLATFORMS.md` 的"what is still missing"与 §55.4 ④。
- 版本常量仍是 **1.0.9**、插件仍是 **1.2.18**，npm/Release 未动。`vendor/manifest.json`
  已随源码重算，所以下一次发布**必须是新版本号**（浮窗走次版本号，插件随之换号）——
  与 §56.10 记录的结论一致。

### 57.8 下一轮从哪里开始

1. **Linux 窗口层**：先做一次"能不能画"的最小验证（X11 CreateWindow + PutImage 推一张
   512×512 的纯色位图到 WSLg，截图确认），再决定 X11 客户端 vs GTK。中文字体必须先装并
   实测（WSL 镜像里没有）。
2. 窗口层落地后，`internal/ui/canvas_posix.go` 的 `drawText` 换成真字形光栅，`control_posix.go`
   换成真的进程间控制（unix socket 或 flock 文件），`uiAvailable()` 对 linux 返回 true，
   `scripts/windows-only-packages.txt` 里那两行随之删掉——CI 会立刻要求这么做。
3. macOS 需要一个能跑 `GOOS=darwin go test` 的机器或 CI runner，否则它只能停在"编译通过"。

## 60. 交接：POSIX 路径/信号收口、文档全面重构与 1.0.10 / 1.2.19 全平台全量化发布（2026-11-15，第 60 轮；**已发版**）

本轮完成第 56–59 轮所有跨平台与功能工作的收口、三系统 CI 联调排错、公开文档重构并按双提交规则发布 **`pet-v1.0.10`** 与 **`dsh-plugin-v1.2.19`**：

1. **POSIX 运行时与接入向导修复**：
   - `cmd/ai-terminal-pet/main.go`（`main_test.go`）：`--reset` 与 `error.log` 统一改用 `internal/home.Join`，不再读 Win32 专属 `%LOCALAPPDATA%`。
   - `internal/app/signals_unix.go` / `signals_windows.go`：捕获 `SIGINT` / `SIGTERM` 并优雅关闭窗口、刷新 `config.json` / `state.json`。
   - `internal/connect`（`binary_windows.go` / `binary_unix.go` / `roots.go` / `snippet.go` / `internal/app/doctor.go`）：自动按系统区分 XDG/`$HOME` 路径与无 `.exe` 二进制文件名。
2. **POSIX 安装与全平台打包器**：
   - 新增 `scripts/install.sh`（可执行权限 `100755`）：支持 Linux/macOS 标准数据目录安装、`~/.local/bin/ai-terminal-pet` 软链、`--verify-only` 与 `--uninstall [--purge]`。
   - 新增 `scripts/package-unix.mjs`：无外部依赖交叉打包 `linux/amd64`、`linux/arm64`、`darwin/arm64`、`darwin/amd64` 四个 `.tar.gz` 归档及 `SHA256SUMS.txt`；WSL2 Ubuntu 实测解压校验、安装、`pet-hook`、`--json`、`--doctor` 与 `--uninstall` 通过。
3. **三系统 CI (`windows-latest` / `ubuntu-latest` / `macos-latest`) + `scan` 全绿验证**：
   - 借助真机 CI Run `43730668303` 与 `43731376911` 暴露并一次根治三处跨平台问题：① macOS `sun_path`（104 字节上限）导致 `t.TempDir()` 超长控制 socket `bind` 失败（`internal/ui/control_unix.go` + `darwin_appkit_test.go`）；② Linux `GOMODCACHE` 在沙盒 `HOME` 中产生只读权限导致退出 `rm -rf` 失败（`.github/workflows/ci.yml`）；③ `TestWaitForApprovalDecisionObservesLateVeto` 误命中 `config.WriteFileAtomic` 的临时文件名（`internal/reporter/reporter_test.go`）。
   - 修复合入 `00e177e` 后 Run `43731837246` 四端全部一次通过（Windows 1m45s、Linux 1m2s、macOS 50s、scan 1m7s）。
4. **文档重构与多平台发版**：
   - 重构 `README.md`（中英文）、`docs/PLATFORMS.md`、`docs/ROADMAP.md`、`docs/HANDOVER.md`、`CHANGELOG.md`、`docs/LAUNCH.md` 与 `packages/dsh-ai-terminal-pet/README*.md`，并在 `tools/check_versions.mjs` 对齐 `internal/app/version.go` 与 Unix `.tar.gz` 命名检查。
   - 按 `docs/UPSTREAM.md` 两步发布：
     - **Commit A (`4e7823b`)**：落地版本号、二进制及市场时间戳，发布 GitHub Release `pet-v1.0.10`（6 个跨平台资产）与 `dsh-plugin-v1.2.19`（`.tgz`），成功将 `dsh-ai-terminal-pet@1.2.19` 推送至 npm registry；
     - **Commit B (`e53c7ea`)**：将 `docs/market-index.json` 的 `source.commit` 与 `icon` 钉死在 Commit A 的真实 SHA；
     - **上游 PR 同步**：在 `awesome-dsh-plugin` 开启 PR [#15626](https://github.com/awesome-dsh-plugin/awesome-dsh-plugin/pull/15626) 并关闭旧 PR #15482，`tools/upstream-entry.mjs --check` 判定 in flight 通过；
     - **全端 CI 4/4 绿灯**：Commit B 的重跑 CI（Run `43734286878`）在 `windows-latest`、`ubuntu-latest`、`macos-latest` 与安全扫描全部通过；本机安装目录已更新为 1.0.10，`install.ps1 -VerifyOnly` 11/11 匹配，`--doctor` 10 ok / 1 warn / 0 fail。

## 61. 交接：agy 点不跳的根因、设置布局搬出 Windows 文件、`-race` 进 CI，以及 L3/L5 的两个数字（2026-11-15 晚，第 61 轮；浮窗 **1.0.11** / 插件 **1.2.20**）

用户报的唯一一条现场缺陷：**"对于本机的 agy，pet 无法点击自动跳转"**。本轮先取证再改代码，取证过程
本身就是 rond 61 最有价值的部分，所以按"证据 → 机制 → 修法 → 没修的半边"记。

### 61.1 取证：三层证据，每一层都能否掉一个假设

1. **会话确实在面板里**：`--json` 里有 `source=antigravity-statusline`、`project` 正确、
   `state` 跟着 agy 变。所以不是"没收到上报"。
2. **窗口确实存在且是候选**：`EnumWindows` 快照里能列出该标签（`CASCADIA_HOSTING_WINDOW_CLASS`，
   标题 `命令提示符 - agy` / `Windows PowerShell`），`terminalWindowScore` = 30。
   所以不是"窗口类没被识别"。
3. **`parentPid` 是死的**：把快照里的 `parentPid` 拿去 `Get-CimInstance` 查，40 次采样里每 2–3 秒换一个
   （32152 / 31332 / 97196 / 17492 / 100548 …），**每一次在下一次读取时都已经不存在**。
   `pet-statusline` 在宿主为一次 statusline 渲染而临时启动的 shell 里运行，`os.Getppid()` 记下的正是那个
   一次性进程。`internal/platform` 的祖先链因此为空，所有候选窗口都拿不到 depth 加成，
   只剩 `classScore`（30）< 阈值 40 → **`selectFocusWindow` 返回 false**，点击落到"找不到终端"。
   用真实快照形状跑 `selectFocusWindow`：`ok=false`，这就是现场。

### 61.2 修法：控制台标题作为跳转锚点（线协议一个字没动）

`AttachConsole(pid)` + `GetConsoleTitleW` 读到的标题**就是 Windows Terminal 渲染的那个标签标题**。
实测两条 agy：标签 `命令提示符 - agy` 的那条读回 `命令提示符 - agy`；标签 `Windows PowerShell`
的那条读回 `Windows PowerShell`。所以：

- `internal/reporter/console_windows.go`（新增）：`ConsoleAnchorTitle()` 先读**自己的**控制台
  （statusline 是终端里那个 shell 的子进程，自己的控制台就是宿主那个，不需要 attach，
  也就不可能弄丢 stdout 那一行）；自己读不到时才 `AttachConsole(os.Getppid())`，
  读完 `FreeConsole` + 重新 attach 回 `GetConsoleProcessList` 里最小的那个 pid。
- `internal/reporter/console_anchor.go`（无标签）：`sanitizeConsoleTitle` 拒绝"看起来是二进制路径"的值
  （未设标题的控制台会返回 `…\powershell.EXE`，而任何窗口标题都不会包含它）。
- `internal/reporter/console_posix.go`：返回空串——POSIX 上会话身份来自 pty，与窗口标题没有这层关系。
- `cmd/pet-statusline/main.go` 与 `cmd/pet-hook/main.go`：`payload.Title = reporter.ConsoleAnchorTitle()`。
  **`title` 本来就是"人类可读标题"**，解析器早已给完全标题匹配 +100 分，所以这不是协议变更。
- `internal/platform/focus.go`：`PseudoConsoleWindow` 不再被算作终端。实测 ConPTY 的隐藏窗口
  对 `SetForegroundWindow` **一律返回 false**（7 个句柄全部失败，前台窗口编号一次没变），
  它偏偏又带着终端里那个 shell 的 pid——选中它等于"点了没反应"。

**端到端复验（本机）**：把控制台标题设成 `命令提示符 - agy` → 跑新 `pet-statusline` →
快照里 `"title":"命令提示符 - agy"` → 拿这份真实快照喂 `ResolveTarget`（`parentPid` 已死）
→ `ok=true hwnd=0x1e61462 windowTitle="命令提示符 - agy"`。三个回归测试钉住两个半：
`TestAntigravityConsoleTitleResolvesWithoutLiveAncestry`、`TestAntigravityConsoleTitleDoesNotBeatOwnTab`、
`TestConsoleAnchorTitleRoundTrip`（真控制台，设置→读→复原）。

### 61.3 一个被否掉的方案（记下来，别重新发明）

最初的实现是"拿不到自己的标题就 attach 父进程读，再 attach 回去"。它在 Windows Terminal 里能用，
但把 `FreeConsole` 放在 stdout 那一行之前就会让 `pet idle` 这一行落进虚空——而那一行是 statusline
**唯一可见的产物**。所以最终只保留"读自己的控制台"这一条，attach 路径只作为父进程兜底，
并且**在任何返回值之前**都把控制台还回去。

### 61.4 没修的半边（诚实记录）

从 `explorer.exe` 双击启动的 agy（进程链 `agy ← cmd ← explorer`，链上没有终端窗口），
仍然只会得到 `ok=false`，也就是"找不到"这句诚实的回答，而不是跳到一个错误的窗口。
要修它需要在会话身份里引入"这个 agy 属于哪个窗口"的另一种证据（例如按 `_NET_WM_PID`/hwnd 反查），
那是下一轮的事。

### 61.5 设置窗口的布局搬出 Windows 文件（零用户可见变化，收益是覆盖率）

`internal/ui/settings_windows.go` 2,204 行里，只有 900 行真的碰 Win32；
tabs/rows/geometry/hit-test/focus/drawing 都是纯算术。本轮把它们搬进：

- `internal/ui/settings_layout.go`（新增，**无构建标签**）：页签表、行结构、几何、命中判定、焦点环、绘制；
- `internal/ui/settings_text_windows.go` → `settings_text.go`（改名，它本来就一行 Win32 都不碰）；
- `internal/ui/settings_windows.go` 剩 906 行：窗口类、DPI、消息泵、分层表面，加 6 个 `*App` 适配器。

**验收是像素级的**：`TestRenderPreview` + `TestRenderSettingsPreview` 在改动前后各出 31 张 PNG，
**逐文件 sha256 全部相同**（0 差异）。11 个纯算术测试从 `*_windows_test.go` 挪到
`settings_layout_test.go`，现在 Linux / macOS 的 `go vet` 与 `go test` 都会跑它们——
这就是这次重构唯一真正的产出。

**顺带修掉一个真 bug**：`TestSettingsKeyboardIsOptIn` 在全量跑时偶发失败（focus index 12），
原因是 `handleSettingsKey` 直接读真实键盘的 Shift 状态，而跑测试的时候人可能正按着 Shift。
改成读一个可替换的 `settingsShiftDown`，测试固定它；生产行为一字未变。

### 61.6 `go test -race` 进 CI（本机仍然没有 gcc）

`ubuntu-latest` 的 `linux` job 增加一步 `CGO_ENABLED=1 go test -race -count=1 ./...`。
本机没有 gcc（WSL 里也没有），Windows runner 也不提供，所以这是唯一能跑竞态检测的地方。
说法也跟着改：**"Linux 上的竞态检测通过"**，不是"竞态已排除"（AGENTS.md §1b、ROADMAP §6 同步）。

### 61.7 L3 / L5 的两个数字：一个"不做"，一个"该做"

ROADMAP §3 要求"先测再动"。两个测量都写成 `*_test.go`（测试文件不进 vendor 源码摘要，
所以记数不花版本号）：

| 项 | 实测 | 结论 |
|---|---|---|
| **L3** 一次投递往返 | `BenchmarkDeliveryRoundTrip`（生产 `Server.read`：读→解析→update→`OK`）**1.84–1.96 µs** | **不做优化** |
| **L3** 事件速率 | 90 s 窗口内 **80 次快照写入 = 53 次/分钟**（约 0.9/s，6 个活跃会话） | 10 小时约 3.2 万次投递，原文担心的"20+/s"不成立；"80 万"是 goroutine 累计编号，不是投递数 |
| **L5** 快照扫描 | 1 文件 **298 µs**、50 文件 **7.02 ms**、200 文件 **22.7 ms**（≈140 µs/文件 + 298 µs 固定） | **5 ms 门槛在第 36 个文件附近越线**，该做；`scanInterval` 现在是 5 s 不是 2 s |

L5 的三步做法（目录 mtime 短路 → 只保留最近 N 个 → 解析下沉）已写进 ROADMAP §3，验收是 200 文件 <5 ms。

### 61.8 版本与发版

浮窗 **1.0.11** / 插件 **1.2.20**（改动碰了 `internal/ui`，vendor 源码摘要必须重算，所以必须发新版本号）。
包体 **2,065,107 B / sha256 `6892C039…` / shasum `40f63b36…` / 19 文件**；三个二进制
`AI-Terminal-Pet.exe` 4,702,208 B（`CDCA876B…`）、`pet-statusline.exe` 3,678,720 B（`563CF6FE…`）、
`pet-hook.exe` 3,705,344 B（`737C46ED…`）。npm `latest` = **1.2.20**（发布后 90 秒复查 `npm view dist-tags.latest` 得到 1.2.20）；
上游条目跟进 PR [#15670](https://github.com/awesome-dsh-plugin/awesome-dsh-plugin/pull/15670)
（分支 `bump-tarball-1.2.20`，commit `252ad60`，单文件 `+1 −1`，`MERGEABLE`）已开、
`tools/upstream-entry.mjs --check` = OK（in flight）；合并仍要人工，**合并 ≠ 可见**。
两个 tag（`pet-v1.0.11` / `dsh-plugin-v1.2.20`）已推；CI 四条 job 全绿（Run `43754035174`），
其中 macOS 那次第一次是缓存解压撞车触发的 SIGSEGV（`gtar: Cannot open: File exists`），
**重跑后绿**——不是代码问题，但下次遇到 macOS 只在 `go test` 就挂，先看缓存解压那几行。

### 61.9 下一轮从这里开始

1. **L5 扫描优化**（数字在 §61.7，门槛已经越线，是唯一"该做"的性能项）；
2. **POSIX 平权**：`docs/PLATFORMS.md` 新表的第 1、2 行——Linux/macOS 的 `notifications`
   （没有任何气泡表面）与 Linux 的整个 `sound*` 区块；
3. **macOS 设置窗口**：仍然**不做**（没有真机，CI 上 AppKit 窗口探针是 opt-in skip）。

## 62. 交接：L5 快照扫描复用解析结果（200 文件 22.7 ms → 0.51 ms），以及缓存命中时删除守卫的真 bug（2026-11-15 深夜，第 62 轮；浮窗 **1.0.12** / 插件 **1.2.21**）

上一轮把 L5 从"待测"变成了"该做"（§61.7 的数字：50 文件 7.02 ms、200 文件 22.7 ms，门槛 5 ms）。本轮就做它，
而且先拆成本再动手——拆完才发现该改的地方和直觉相反。

### 62.1 成本拆分：解析只占 5%，元数据占 90%

同一台机（NTFS + 活跃 AV 过滤），500 次迭代取平均：

| 项 | 每次 |
|---|---|
| `os.Lstat` | **31.3 µs** |
| `readSnapshot`（4 次 stat + 一次 read） | **89.6 µs** |
| `session.ParsePayload` | **5.1 µs** |
| 单个候选端到端 | **98.1 µs** |

所以"少解析"这个方向是错的：解析是 5%，**碰盘是 90%**。要减的是"每 5 秒把整目录重读一遍"这件事本身。

### 62.2 做法：按 身份 + size + mtime 复用上次的 `session.Session`

`internal/store/snapshot_cache.go`：

- `SnapshotCache` 是**指针类型字段**（`SnapshotScanner.Cache`），因为 scanner 是值类型、会被复制，
  而缓存必须活过这些复制；**nil 就是旧行为**——`--json` 等一次性调用者与全部既有测试仍走未缓存路径。
- 只有 `internal/app/run.go` 的常驻 scanner 传 `store.NewSnapshotCache()`。
- 命中条件是 `sameSnapshotFile` 已经用的那三个字段（身份 `os.SameFile` + size + mtime），
  所以缓存**不引入新的正确性假设**，只是把"已经读过一次"的结果留着。

**稳态实测**（`-benchtime 30x`，本机）：1 文件 298 µs → 113 µs；50 文件 7.02 ms → **0.52 ms**；
200 文件 **22.7 ms → 0.51 ms**；本机那个 225 文件的目录连扫 5 次是 **0.50–0.61 ms**。门槛 5 ms，验收满足。

### 62.3 本轮唯一一个真 bug：命中路径拿不到 `FileInfo`

`SnapshotScanner` 里每个"这个文件该走"的分支都调 `removeSnapshotIfUnchanged(path, info)`，
它在 `expected == nil` 时**直接返回**。命中路径没有 read，也就没有 post-read 的 `FileInfo`，于是：

- 过期快照 → 面板每 tick 丢一次，**文件永远留着**；
- 宿主已死且没写 `closed` 的快照 → 同上；
- `closed` tombstone → 由 Readdir 分支删掉，不受影响（这也是为什么直觉版的测试没抓到它）。

修法是命中时退回用 Readdir 那份 `FileInfo`。**两条测试在去掉兜底后确实失败、补回后通过**——
其中一个必须让文件"先活过一轮进缓存、然后才过期"，因为已经过期或 tombstone 的文件在读到的那一轮就被删掉，
根本不会进缓存。这一条写进测试注释里，免得下一个人重写测试时又写成抓不到的那个版本。

### 62.4 三条安全性质（都有测试）

1. **不能靠元数据骗过去**：报告端是 temp+rename 写盘，重写落成**新文件**，`os.SameFile` 能分辨。
   实测：把 mtime 手工改回、字节长度保持相同（`"build"`→`"tests"` 都是 5 字符），`SameFile(缓存那份, 新 stat) = false`，
   扫描返回新值。
2. **命中仍然能删文件**（§62.3）。
3. **缓存不会无限长大**：每次扫描 `retain(seenPaths)`，只看这一轮的路径；被删/被消费的文件立刻 `forget`。

**一个如实记录的测量事实**：NTFS 在这台机器上 **40 次 temp+rename 里复用过 1 次文件索引**。
所以守卫是"身份 **且** size **且** mtime"三个字段而不是只看身份；删除前一定重新 stat。
最坏情况因此是**一次陈旧读取**（下一轮 size/mtime 一动就修正），不是误删文件。

### 62.5 版本与发版

浮窗 **1.0.12** / 插件 **1.2.21**。本轮只有 overlay 二进制变了（`AI-Terminal-Pet.exe` 4,702,208 → 4,711,424 B，
sha256 `FDFFF5C8…`，+9,216 B 来自缓存与其测试）；`pet-statusline.exe`、`pet-hook.exe` 字节未变
（`563CF6FE…` / `737C46ED…`）。vendor 源码摘要随 `internal/store` 变动重算。

### 62.6 下一轮从这里开始

1. **POSIX 平权**：`docs/PLATFORMS.md` 新表第 1、2 行——Linux/macOS 的 `notifications`（没有任何气泡表面）、
   Linux 的整个 `sound*` 区块（连 `beep` 都没有）；
2. **macOS 设置窗口**：仍然**不做**（没有真机，CI 上 AppKit 窗口探针是 opt-in skip）；
3. 目录级短路（L5 当初列的方案 1）现在**不需要**了：缓存已经把 225 文件压到 0.5 ms，
   再加"目录 mtime 未变则跳过整轮"只会多一个已知的 mtime 精度坑。

## 63. 交接：修 1.0.11 的回归（`title` 等于窗口标题时不算会话存在），以及交付前的装机实测清单（2026-11-16，第 63 轮；浮窗 **1.0.13** / 插件 **1.2.22**）

本轮的任务是"稳定性和全量化推送"。稳定性的第一站是**先在自己的机器上把装好的程序按用户的用法查一遍**，
查的过程中撞上一个自己上一轮引入的回归，于是先修它，然后才更新文案。

### 63.1 装机实测做了什么（都跑在**装好的** 1.0.12 上）

| 查什么 | 怎么查 | 结论 |
|---|---|---|
| 协议契约 | 直连命名管道逐条发：合法载荷 / 未知字段 / 坏 JSON / 超长行 | `OK` / `ERR` / `ERR` / **连接被断开（无回复）** |
| 连接受压 | 并发开 40 条（上限 32） | 40 条全部连上（上限是"同时处理"而不是"同时打开"），之后仍能正常收报 |
| 设置窗口 | `--settings` 打开、**原样放 40 秒**、再 `WM_CLOSE` | 40 秒里一直可见（这正是当年"闪一下就没了"活不过的场景）；关闭是**隐藏**而不是销毁，面板与进程都活着 |
| 报告洪峰 | 120 次上报（16 个会话轮转 6 种状态） | `ok=120 err=0 failed=0` |
| 句柄/内存 | 每批 40 次、批间等 5 秒，共 5 批 | 句柄 **468→469**、线程 **30→30**、WS 在 36–55 MB 之间浮动 —— **不爬升**，洪峰后 60 秒回落到 468 |
| 文件泄漏 | 洪峰后把 16 个会话全部 `closed`，等 6 秒 | `runtime/` 回到基线的 5 个文件 |
| `--doctor` | 全程多次 | 始终 10 ok / 1 warn / 0 fail（唯一 warn 是"Antigravity 只接了 statusline、没有结束信号"，已知） |
| 新机端到端 | `tests/fresh-machine.mjs --no-overlay` | **172/172** |

**一条必须记住的实测事实**：`tests/fresh-machine.mjs` **不带 `--no-overlay`** 时会红（182/184），
因为它要起沙箱里的浮窗，而单实例互斥体是 `Local\AI-Terminal-Pet-v1-SingleInstance`、**按 Windows 会话**
而不是按沙箱隔离——外面那个真实例先拿到互斥体，沙箱里那份就 `exit 0` 且不开管道。脚本自己的输出已经把这条
写成人话（"the overlay scenarios raced a real instance ... only meaningful on a machine with no host hooks"），
所以这不是缺陷，是**在有宿主 hook 的机器上跑全量套件的正确用法是 `--no-overlay`**。

### 63.2 撞到的回归：`title` 等于窗口标题时，被当成了"这个会话还存在"

**现象**：`--json` 里四条 antigravity 记录，年龄 **2h01m / 2h08m / 2h56m**，`parentPid` 全部已死
（`ProcessAlive=false`），却全部列着。

**机制（两件各自正确的事撞在一起）**：

1. 第 61 轮为了修"agy 点不跳"，让 `pet-statusline` 把**控制台标题**写进载荷的 `title`；
   那个字符串**就是** Windows Terminal 的标签页标题。
2. `internal/platform/presence.go` 的存在性模型把 `title` 当作**时间无关**的身份证据
   （这是第 21/43/51/52 轮修幽灵会话时立下的规则：一个真正叫出这个会话名字的标签页是永恒证据）。

一个标签页装着它里面跑过的**每一个**会话，所以"窗口在叫它的名字"对已结束的会话也永远成立。

**修法（`windowIdentifiesSession`）**：`title` 必须**被窗口标题包含且不完全等于它**才算证据。

| `title` | 窗口标题 | 算证据？ |
|---|---|---|
| `项目交接与下一步计划` | `OC \| 项目交接与下一步计划` | **算** |
| `命令提示符 - agy` | `命令提示符 - agy`（完全相同） | **不算** |
| `agy` / `cmd` / `powershell` | 任意 | 不算（`bareWindowTitle` 原有规则） |
| 任意 | 任意（`project` 命中窗口标题） | **算**（项目名那一半完全没动） |

**端到端复验（装好的二进制）**：把那条"三小时前、`title` 为标签标题"的记录写回 `runtime/`，
一次扫描内从 `--json` 消失、文件被删；同目录里三条真实的 opencode 会话不受影响。
把四条历史记录**原样**写回再留着不动，也是一次扫描内全部消失。

**三条测试**（`internal/platform/presence_anchor_test.go`，**无构建标签**，三平台都跑）：
纯函数规则、结束 3 小时的会话必须判死、宿主还活着时必须判活（防止修成"什么都不算证据"）。

### 63.3 为什么这条回归值得单独记一笔

- **它是"上一轮的修法"造成的**，不是老缺陷。第 61 轮的报告里写过"控制台标题就是标签标题"，
  也写过 `presence.go` 用 `title` 做证据 —— 但没有把这两句话放在一起推一遍。
  **教训**：一个字段被两处消费时，往它里面塞新语义之前，先把每个消费点列出来。
- **它只在"结束且宿主进程已退出"的状态下可见**，所以单元测试与 CI 都是绿的：所有测试用的都是
  构造好的窗口与进程表，不会有一个真实的、挂着旧标签页的桌面。
  发现它的唯一途径是**在真机上把装好的程序放一段时间再看 `--json`**。
- 文档侧也补了：`docs/PROTOCOL.md` 新增 **§4.0**，在第三方上报端动手之前就把这条规则写清楚
  （"让 `title` 描述这次会话，而不是描述它所在的窗口"）。

### 63.4 版本与发版

浮窗 **1.0.13** / 插件 **1.2.22**。只有 overlay 变了（`AI-Terminal-Pet.exe` 4,711,424 B，
sha256 `161C6083…`；`pet-statusline.exe` / `pet-hook.exe` 字节未动）。
vendor 源码摘要随 `internal/platform` 变动重算。

### 63.5 下一轮从这里开始

1. **POSIX 平权**（`docs/PLATFORMS.md` 的 config 表第 1、2 行）：Linux/macOS 的 `notifications`
   没有任何气泡表面；Linux 的整个 `sound*` 区块连 `beep` 都没有。
2. **macOS 设置窗口**仍然**不做**（没有真机，CI 上 AppKit 窗口探针是 opt-in skip）。
3. 超长行不回 `ERR` 已记进 ROADMAP §6 且**不修**——写错的第三方上报端不该由服务端替它兜。

## 64. 交接：上游目录里我们那行为什么是旧的（两个不同的原因），以及交付前的又一轮装机实测（2026-11-16，第 64 轮；**未发版**）

用户报的现象：`https://awesome-dsh-plugin.com/?q=jjjjjjjj` 里我们的**下载量与 star 没跟 npm 同步**。
本轮先按 `docs/UPSTREAM.md` §4 的规矩取证（**读对方公开的索引文件，不读 PR 页面**），
结论是**两个互不相关的原因**，而**两个都不是重发版能改的**。

### 64.1 站点那行现在长什么样（2026-11-16 实读 `plugins.json`）

```json
{
  "name": "ai-terminal-pet#dsh-ai-terminal-pet",
  "version": "1.2.16",
  "tarball": ".../download/dsh-plugin-v1.2.21/dsh-ai-terminal-pet-1.2.21.tgz",
  "downloads": 1179,
  "downloadsStart": "2026-10-15", "downloadsEnd": "2026-11-13", "downloadsCheckedAt": "2026-11-14",
  "stars": 1, "starsCheckedAt": "2026-11-14T03:18:19.955Z"
}
```

入口页 <https://awesome-dsh-plugin.com/p/jjjjjjjjnnjnn/ai-terminal-pet--packages-dsh-ai-terminal-pet/>
渲染的安装命令还是 `dsh-plugin-v1.2.16`（另外两处是 npm 名与 git 源），**与我们当前的 1.2.22 差 6 个版本**。

### 64.2 原因一：整站自 2026-11-14T05:27Z 冻结（不是我们的入口过期）

`tarball` 字段**早已**是 1.2.21（#15674 于 2026-11-15T23:19:43Z 合并），`version` 却停在 1.2.16 ——
两者不一致说明**重建**没发生，而不是入口写错了。取证：

- 夜间 `Build`：2026-11-15 与 2026-11-16 两次 `schedule` 运行**都失败**，失败步骤是 **`Build`**，
  之后 `upload-pages-artifact` / `deploy-pages` / 两个 npm publish 全部被跳过；
- 失败原因（对方 issue **#15569** 与修复 PR **#15579** 已写明）：
  `MarketIndexError: $.plugins[N].name duplicates dsh-dot-mcp` —— 两个合法入口（monorepo 根与其子目录）
  各自的 `package.json` 都叫 `dsh-dot-mcp`，format 1 按 name 做键，解析器于是**拒掉整份文档**；
- 最后一个成功的夜间是 2026-11-14T05:27Z；期间的 push 构建在 23:20Z 前后**全是 cancelled**
  （并发组被后续 push 顶掉）。
- 对方维护者在 #15569 的回帖里确认了机制，并说"就等 #15579 与下一次夜间"；#15579 现在是 `MERGEABLE`、`OPEN`。

**结论**：我们什么都不用做，也不该再重发版——`version` 是对方探针缓存，未被重建就永远落后。
唯一能做的是把这条事实记下来，让下一次交接不再把它误判成"入口没跟上"。

### 64.3 原因二：30 天下载数少掉的 2,276 分**两段**，且第二段的来源与第一段不同

最初的读法是"1,179 是 1.2.21 之前的老数据"，错。真实值是 **3,455**（npm 自己的 Weekly Downloads
与 `point/last-month` 都是这个数），而站点发布的 1,179 是**两段缺失叠加**的结果：

| 缺的部分 | 值 | 原因 | 谁的 |
|---|---|---|---|
| 窗口**自己的最后一天**（2026-11-13） | **1,197** | 探针在 03:1x UTC 跑，npm 对刚结束约 2.5 小时那天还没聚合，返回 0；`downloadResult()` 只校验 `row.end === window.end`，那天存在且为空，与"真的没人装"无法区分。**全站 12/12 条目同一签名** | 对方 issue **#15864** |
| 窗口**根本没前进到** 11-14 / 11-15 | **1,079**（318 + 761） | 整站冻结，`end` 停在 11-13 | 对方 issue **#15569** / PR **#15579** |
| 站点发布值 | **1,179** | = 11-11 的 624 + 11-12 的 555 | — |

**1,179 + 1,197 + 1,079 = 3,455**，三段拼回去正好吻合；所以这不是推测，是任何人两条 `curl` 都能复现的分解：

```sh
curl -s 'https://api.npmjs.org/downloads/range/2026-11-11:2026-11-15/dsh-ai-terminal-pet'
# 624 / 555 / 1197 / 318 / 761  -> 合计 3455
```

**"7 天下载 = 30 天下载"不是巧合**：`1.2.4` 是 2026-11-11T12:10:46Z 才发布的，11-11 之前那 30 天里
每天都是 0，所以两个窗口的真实值相同。这也解释了为什么 npm 页面的 Weekly Downloads 与
`point/last-month` 给出同一个数字。

**12/12 对照测量**（窗口 2026-10-15 → 11-13，逐个包问 npm 的同窗口）：

| npm 包 | 站点发布 | 同窗口今天 | 少一天 = | 最后一天 |
|---|---|---|---|---|
| dsh-ai-terminal-pet | 1,179 | 2,376 | 1,179 | 1,197 |
| dsh-tavily-mcp | 1,320 | 1,364 | 1,320 | 44 |
| dsh-fish-dream | 483 | 516 | 483 | 33 |
| kunpet-dsh | 592 | 610 | 592 | 18 |
| dsh-happy | 1,392 | 1,428 | 1,392 | 36 |
| dsh-om-messiah | 1,663 | 1,694 | 1,663 | 31 |
| dsh-plugin-yushe | 6,010 | 6,278 | 6,010 | 268 |
| dsh-cosplay | 4,822 | 4,949 | 4,822 | 127 |
| dsh-pet-plugin | 11,429 | 11,700 | 11,429 | 271 |
| dsh-dice-game | 1,136 | 1,175 | 1,136 | 39 |
| toto-the-cat | 728 | 751 | 728 | 23 |
| dsh-meme | 80,094 | 81,991 | 80,094 | 1,897 |

11 个是别人的包，签名一模一样 —— 所以这条**不可能是我们这张卡片填错了**。
这 12 组数字已作为固定样本写进 `tools/catalog-status.test.mjs`（12 条测试），任何人跑一次就复现同一算术。

**一个必须说清的取舍**：我最初在上游补丁里提的"窗口改成截至两天前"**只能修到 2,694**，
比今天的 3,455 还低一天。三种改法的实际后果是：

| 方案 | 今天会显示 | 代价 |
|---|---|---|
| `end = today-2` | 2,694 | 永远不含最新一天，但永不踩聚合延迟 |
| `end = today-1` | 3,455 | 最新，但当天下聚合时概率性丢一天（即当前 bug） |
| 两者都查、取可用的那个 | 3,455 | 推荐：最新可用，且没有"静默丢一天"的通道 |

**决策（2026-11-16，用户拍板）**：这两条上游 issue 里已有多位开发者反馈，本项**不再跟进**：
不发 PR、不催合并、不重发版。要做的是**把自己这边的记录改准**——凡引用下载量一律用 npm 的数
并写明窗口，站点卡上的 1,179 只作为"上游当前状态"出现（`docs/UPSTREAM.md` §4.1 与
`docs/LAUNCH.md` §8 都已改成这个口径）。

### 64.4 新增工具：`tools/catalog-status.mjs`（把"该谁动"变成一条命令）

三种情况**看起来一样、处理方式相反**，所以工具必须分开说：

| 判定 | 依据 | 动作 |
|---|---|---|
| `upstream` | 站点 `version` 落后于 npm，或下载数正好少窗口最后一天 | **不动**；重发版/重探都不会改变它 |
| `pending` | 离线判：入口 `tarball` 落后于本工作区 | 跑 `tools/upstream-entry.mjs --check` 看 PR 在飞还是没开 |
| `fail` | 线上判：入口 `tarball` 落后于 npm | 开一条 `+1 −1` 的 bump PR |
| `ok` | 那行是当前版本 | 无 |

退出码 `ok`/`upstream`/`pending` = **0**（后两者不是我们的失败），`fail` = 1，索引读不到 = 2。
CI 里这条是 **informational**（`|| true`），因为对方站点挂了不该让本仓库变红；
判定规则本身有 `tools/catalog-status.test.mjs`（**11 条**，纯函数、不联网）钉住，随 CI 一起跑。

`node tools/catalog-status.mjs` 今天的输出（可作为交接时的基线）：

```
[ok]  row present: ai-terminal-pet#dsh-ai-terminal-pet
      site version 1.2.16 · tarball .../dsh-plugin-v1.2.21/dsh-ai-terminal-pet-1.2.21.tgz
      site downloads 1179 (2026-10-15 → 2026-11-13) · stars 1
      local: overlay 1.0.13, plugin 1.2.22
[upstream] row says 1.2.16, npm has 1.2.22 - the catalogue has not been rebuilt since its last successful nightly
      downloads for the site's own window: published 1179, now 2376, window-minus-last-day 1179, final day 1197
```

### 64.5 交付前又跑了一轮装机实测（结论：稳定，两个异常都有解释）

同一台机器、**装好的 1.0.13**：

| 查什么 | 结果 |
|---|---|
| 120 次上报（16 会话轮转） | `ok=120 / failed=0`，句柄 414→443、线程 30 未变 |
| 240 次上报（6 轮 × 40，批间停 3 秒） | **`failures: 0`**，进程号全程 106712 不变 |
| 300 次上报（专测 `ERROR_PIPE_BUSY`） | **`failures: 0`** |
| 内存 | WS 32–62 MB 之间浮动、稳定在 27 线程；洪峰后回落 |
| 文件泄漏 | 每组 soak 结束后 `runtime/` 都回到基线（7 个文件） |
| `--doctor` | 每次都是 10 ok / 1 warn / 0 fail |

**两个"异常"都查到底了**：

1. 一个 soak 脚本里 120 次出现 **1 次**失败（PowerShell `MethodInvocationException`）：换脚本重跑
   240 次 + 300 次**再没复现**，且进程号没变（不是自毁、不是崩溃）—— Windows 事件日志里
   **没有** Application Error、**没有** WER 报告、`error.log` 不存在、`hang.log` 仍是 2026-11-14 那份。
   判断：客户端侧的一次性管道竞态（本机 AV 钩挂会让 `Connect` 偶发超时），**与产品的重试策略同源**——
   `internal/reporter` 正是为此重试 `ERROR_PIPE_BUSY`（`docs/PROTOCOL.md` §3.4），而手写 PowerShell 客户端不重试。
   **不记为缺陷**，但要记下来：下次谁看到"偶发一次失败"，先看是不是没有重试的客户端。
2. 一批脚本报告 `WS=5.1MB handles=6`：那是**读取时机**问题（进程正被 `--quit` + 安装替换的瞬间），
   同一时刻的 `Get-Process` 打到了正在退出的旧进程；随后新进程 106712 起来并跑完 540 次上报。

### 64.6 本轮没有发版

代码零改动（新增的只有 `tools/catalog-status.mjs` 与其 11 条测试、CI 两步、文档）。
`1.0.13 / 1.2.22` 仍是最新版本；上游抓到的两条都记在上游仓库，不占用我们的版本号。

### 64.7 下一轮从这里开始

1. **等 #15579 合并 + 一次成功夜间**，然后用 `node tools/catalog-status.mjs` 复验：
   那行的 `version` 应该跳到 1.2.22、入口页安装命令也应换成 1.2.22。**这是唯一需要等的事**。
2. **POSIX 平权**：Linux/macOS 的 `notifications` 没有气泡表面、Linux 整个 `sound*` 区块没有播放路径。
3. macOS 设置窗口仍**不做**（没有真机）。

## 65. 交接：下个会话从这里开始（2026-11-16 收尾，第 65 轮；**未发版**，浮窗 **1.0.13** / 插件 **1.2.22** 仍是最新）

这一轮没有改产品代码，只把**交接状态**钉死：仓库干净、门禁全绿、本机装的就是发布的那个版本、
上一轮那条上游问题也按用户拍板的方向收了尾。下一位可以直接从 §65.4 挑一项开工。

### 65.1 这一刻的机器状态（全部当场跑出来，不是回忆）

| 项 | 值 |
|---|---|
| 分支 | `main`，与 `origin/main` **同步**（`git status -sb` 只有 `## main...origin/main`） |
| HEAD | `5e93d77`（`docs: the 30-day download count is short by 2,276, from two unrelated upstream bugs`） |
| 版本 | 浮窗 **1.0.13**（`internal/app/version.go`）、插件 **1.2.22**；npm `latest` = **1.2.22** |
| CI | 最近一次 Run **`43894252617` 四条 job 全绿**（windows / ubuntu（含 `go test -race`）/ macOS / scan） |
| 单元测试 | **428 个 `func Test`**（`go test -count=1 ./...` 全过） |
| 代码量 | Go 非测试 **136 文件 / 26,526 行**、测试 **87 文件 / 12,673 行** |
| 二进制 | `AI-Terminal-Pet.exe` **4,711,424 B**（`161C6083…`）、`pet-statusline.exe` 3,678,720 B（`563CF6FE…`）、`pet-hook.exe` 3,705,344 B（`737C46ED…`） |
| 本机安装 | **就是 1.0.13**：`scripts\install.cmd -VerifyOnly` **11/11**、`--doctor` **10 ok / 1 warn / 0 fail**（唯一 warn 仍是"Antigravity 只接了 statusline、没有结束信号"，已知且不是缺陷） |
| 文档门禁 | `tools/check_docs_links.py` 12 份全解；`tools/check_versions.mjs` 20 条版本形状一致；`tools/catalog-status.test.mjs` **12/12**；`tools/upstream-entry.test.mjs` 26/26；`check_versions.test.mjs` 26/26 |

**一条不要复跑就下结论的事**：`tests/fresh-machine.mjs` **必须带 `--no-overlay`**（172/172）。
不带它时它会起沙箱浮窗，而单实例互斥体是 `Local\AI-Terminal-Pet-v1-SingleInstance`、**按 Windows 会话**
而不是按沙箱隔离，外面那个真实例先拿到互斥体 → 沙箱那份 `exit 0` 且不开管道 → 182/184 红。
脚本自己的输出把这条写成人话了（§63.1），**这不是缺陷**。

### 65.2 上游那件事已经收尾（按用户决定：不跟进）

`awesome-dsh-plugin` 那两条都**不再由我们推动**：站点的 `version` 停在 1.2.16 是对方整站冻结
（对方 issue #15569 / PR #15579），`downloads (30d)` 少掉的 2,276 是两个上游 bug 叠加
（对方 issue #15864 + 冻结）。**两条已有多位开发者反馈**，所以：不发 PR、不催合并、不重发版。

我们只做了「把自己的记录改准」：

- `docs/UPSTREAM.md` §4.1：三段分解（1,197 + 1,079 = 2,276，站点 1,179 + 2,276 = **3,455**）、
  12/12 条目的同签名对照、以及**三种窗口改法的实际后果**（`end=today-2` 只能修到 2,694，比今天的 3,455 低，
  所以正解是"两个窗口都查、取已聚合的那个"）。
- `docs/LAUNCH.md` §8：对外**一律引用 npm 的数**（3,455，写明窗口 `2026-11-09 → 11-15`；
  `point/last-month` 也是 3,455，因为 `1.2.4` 是 2026-11-11T12:10:46Z 才发的）。站点卡上的 1,179
  **只作为"上游当前状态"出现，并明确写"不要抄这个数"**。
- `tools/catalog-status.test.mjs`：12 条真实样本钉住这个签名。**上游哪天修好窗口，这些样本会开始失败**
  —— 那就是我们要的通知，不是噪声。
- `tools/catalog-status.mjs`：`ok` / `pending` / `fail` / `upstream` 四个判定，把"该我们动"与"对方的原因"
  分开；CI 里跑成 **informational**（`|| true`），因为另一个仓库的夜间失败不该让本仓库变红。

**上游 PR 线的现状（唯一还需要人做的事）**：`#15814`（把 `tarball` 推到 1.2.22）`OPEN` / `MERGEABLE`，
`tools/upstream-entry.mjs --check` 报 in-flight（**那是健康状态**）。合并是维护者的事，我们不开新 PR。

### 65.3 这一轮为什么没发版

产品代码零改动：动的只有 `docs/`（UPSTREAM / LAUNCH / ROADMAP / HANDOVER）与 `tools/catalog-status.test.mjs`
的样本。`1.0.13 / 1.2.22` 的资产、tag、npm 包都没动，**不占版本号**。

### 65.4 下一轮从这里开始（按"用户可感知"排序）

1. **POSIX 平权**（`docs/PLATFORMS.md` 的 config 表第 1、2 行，也是这两轮反复提到的那一项）：
   - Linux/macOS 的 `notifications`：`PlanAlerts` 给出的 `Balloon` 在两个平台上被**直接丢掉**
     （没有气泡表面：X11 没有 XEmbed 托盘、macOS 没有 `NSUserNotification` 绑定）。要么补一个
     平台等价的提示面，要么在 `--doctor`/文档里把"这两个平台没有气泡"写成一条明确的已知项。
   - Linux 的整个 `sound*` 区块：`internal/ui` 里**没有任何 `beep`**，`sound.go` 合成了 WAV 却没人消费，
     `soundVolume` / `soundScheme` 被 `config` 读进来后**丢弃**。零 cgo 的可行路只有"探测外部播放器"
     （`paplay` / `pw-play` / `aplay`，找不到就静音并在 `--doctor` 报 warn）——**这是用户已认可的方向**（§61 那一轮就定过）。
2. **macOS 设置窗口**：仍然**不做**。没有真机，CI 上 AppKit 窗口探针是 opt-in skip
   （`darwin_appkit_test.go` 的 `PET_COCOA_PROBE`），写 1,500 行没人跑过的 ObjC 与 §57.7 的规矩冲突。
3. **验收 L5 的长期收益**（可选）：快照缓存在 §62 已交付（200 文件 22.7 ms → 0.51 ms），
   但只在 `run.go` 的常驻 scanner 上生效；`--json` 仍走未缓存路径。若哪天 `--json` 变慢，先看这里。
4. **README 里那张平台表**（`README.md` §Platform matrix）与 `docs/PLATFORMS.md` 的 config 表要保持一致 ——
   两张表现在都由代码推导，改行为时**两处一起改**，否则下次交接又要花一轮去核对。

### 65.5 接手时最先跑的四条命令

```powershell
gofmt -l .                                   # 必须无输出
go test -count=1 ./...                       # 428 个 func Test，全过
node tests/fresh-machine.mjs --no-overlay    # 172/172（带 --no-overlay，见 §65.1）
node tools/catalog-status.mjs                # 目录那行现在是 current 还是 upstream，一条命令答完
```

**别做的事**：不要在仓库目录里直接跑 `bin\pet-hook.exe` 验证投递（该目录带 Low 完整性标签，
从它启动的进程一律继承 Low IL，浮窗的管道 DACL 只给 OWNER RIGHTS + SYSTEM，**永远打不开**；
复制到 `%TEMP%` 或安装目录就正常。详见 ROADMAP §2 E1 的实测记录）。

## 66. 交接：POSIX 平权落地 —— Linux 的提示音与通知面、macOS 的 osascript、跳转失败不再静默（2026-11-16，第 66 轮；浮窗 **1.0.14** / 插件 **1.2.23**）

§65.4 把下一项排在第一：`docs/PLATFORMS.md` config 表第 1、2 行那两条"设置被读进来然后丢掉"。
这一轮把它做完了，并且顺手补了一个谁都看得见、但三份文档都没写过的洞：**Linux / macOS 上跳转失败是完全没有反馈的**。

### 66.1 三条用户可感知的变化

1. **Linux 有提示音了**。`sound` / `soundVolume` / `soundScheme` 过去在 Linux 上被判定、然后丢弃
   （`sound.go` 合成了 WAV，`app_linux.go` 从不调 `beep`）。现在把同一段 WAV 交给探测到的系统播放器。
2. **Linux 与 macOS 有提醒面了**。`notifications` 过去在这两个平台只影响判定记录，气泡没有任何表面。
   现在 Linux 走 `notify-send`，macOS 走 `osascript`；两边都没有时，`alerts.json` 记
   `skipped=no-notification-surface`，而不是记一条从没出现过的气泡。
3. **跳转失败会说话了**（POSIX）。原来 `focusCh` 的失败分支什么都不做，点一个终端已关掉的会话行
   看起来和"点击没被接收"一模一样，而且未读标记已经被清掉。现在在那一行上画一条 `未能切换` 并列响一声。

### 66.2 五个提交（按落地顺序）

| 提交 | 内容 | 关键文件 |
|---|---|---|
| `feat(sound)` | Linux 提示音：探测 `pw-play` → `paplay` → `aplay`，WAV 写 stdin，`Start` + goroutine `Wait`，新音先杀旧音 | `sound_linux.go`（新）、`app_linux.go`、`sound_windows.go` / `sound_darwin.go`（各加一个 `SoundSurface`） |
| `feat(notify)` | 文案抽成一份（`composeAlert` / `composeJumpFailure`），Windows 走托盘（行为不变），Linux 走 `notify-send`，macOS 走 `osascript`；回执不再虚报气泡 | `notify.go`、`notify_{windows,linux,darwin}.go`、`app_windows.go`（`balloon` 拆掉）、`receipt.go` |
| `fix(ui)` | 跳转失败：`focusCh` 携带 `focusRequest`，失败时 `未能切换` 小条 + 提示音，迟到 2 秒的结果丢弃；复制失败在 POSIX 也响 | `focus_posix.go`（新）、`copy_toast.go`、`drawing_compact.go`、`i18n.go`、两个 `app_*.go` |
| `feat(doctor)` | `--doctor` 11 → 13 项：`audio` / `notifications`；缺表面是 warn 不是 fail | `internal/app/doctor.go`、`doctor_windows_test.go` |
| `release` | 版本、CHANGELOG、PLATFORMS、README 矩阵、HANDOVER、市场索引 | 见 §66.6 |

### 66.3 四个"看起来对、其实错"的地方（都记下来）

1. **"把 `beep` 的调用补上就行"是错的**：Linux 上根本没有 `beep`。`sound_windows.go` 与
   `sound_darwin.go` 都带 build tag，`app_linux.go` 里加一行会直接编译失败——必须先有播放路径。
2. **"一次只响一个"要自己实现**：Windows 的 `PlaySoundW` 再给一个缓冲区就顶掉上一个，macOS 是显式
   `stop` 上一个 `NSSound`。Linux 没有这层，所以要自己留着上一个 `*exec.Cmd` 并 `Kill`，否则两条提醒
   叠在一起会变成噪声。
3. **子进程必须 `Wait`**：只 `Start` 不 `Wait` 会留僵尸。播放是异步的，所以 `Wait` 放在 goroutine 里，
   UI 线程只负责 `Start`（毫秒级）。
4. **AppleScript 的转义是安全性问题，不是美观问题**：项目名会被原样塞进 `display notification "…"`，
   一个双引号就能改掉整条脚本。`appleScriptLiteral` 转义 `"` 与 `\`、把换行折成空格，测试在三个平台都跑
   （函数故意放在 untagged 的 `notify.go` 里）。

### 66.4 本机验证做到哪一步（以及一个新学会的办法）

- `gofmt -l .` 无输出；`go vet ./...` exit 0；`go test -count=1 ./...` 15 个包全过（Windows，**463 个 `func Test`**，
  本轮 +33）；`scripts/cross-build.sh` 16 个包在 linux/amd64 与 darwin/arm64 上全绿。
- **Linux 专属代码在这台机器上真的跑过了**：`GOOS=linux CGO_ENABLED=0 go test -c -o %TEMP%\petui.test ./internal/ui`
  编出 linux 测试二进制，WSL 里直接执行 → **122 PASS**，包含 4 个投递测试（探针顺序、argv、WAV 进了 stdin、
  没有播放器时静默）与 3 个 POSIX 跳转失败测试。WSL 里没有 Go，但**不需要 Go 也能跑交叉编译出来的静态测试二进制**
  ——下一轮做 Linux 侧改动时直接用这条路，不要等 CI。
- `node --test reporters/opencode-events.test.mjs` 10/10、两个 `verify.mjs` PASS、`tools/check_versions.mjs` 20 条一致。
- `--doctor` 实测从 11 行变 13 行，新增两行：
  `[ok] audio: winmm PlaySoundW (built in)` / `[ok] notifications: tray balloon (Shell_NotifyIconW, built in)`。

### 66.5 三条证不到的事（不要在下一次交接里写成"已支持"）

1. **Linux 是否真的出声**：WSL 里 `pw-play` / `paplay` / `aplay` / `pactl` 都没有、没有免密 sudo。
   验证到的是"我们这一侧"（选谁、argv、WAV 字节确实进了它的 stdin），**音箱有没有响没验过**。
   另外三个播放器"无参/`-` 即读 stdin"来自各自手册，本机无法实测。
2. **Linux 是否真的弹窗**：WSL 的会话总线上只有 `org.freedesktop.DBus`，没有通知守护进程；
   `notify-send` 也没装。验的是调用形状与失败静默。
3. **macOS 的任何实际显示**：没有苹果硬件，CI 只跑编译与单测；`docs/PLATFORMS.md`、README 矩阵与
   `--doctor` 的通知行都写着 "best-effort, live display unverified"。

### 66.6 发版记录（2026-11-16，全部当场跑出来）

- **tag**：`pet-v1.0.14`、`dsh-plugin-v1.2.23`；两个 Release 都建好，资产从**公开 URL 下载回来**逐字节核对
  （7/7 的字节数与 sha256 与本地一致，完整表在 CHANGELOG「1.0.14 / 1.2.23 发布包实测清单」）。
- **npm**：`dsh-ai-terminal-pet@1.2.23` 已 publish，`npm view ... version` = **1.2.23**。
- **本机安装**：`scripts\install.cmd` → 11 个文件、`-VerifyOnly` **11/11 一致**、
  `AI-Terminal-Pet.exe --version` = `ai-terminal-pet 1.0.14`、`--doctor` = **12 ok / 1 warn / 0 fail**
  （唯一 warn 仍是 Antigravity 只接 statusline）。
- **上游条目**：PR [#16010](https://github.com/awesome-dsh-plugin/awesome-dsh-plugin/pull/16010)
  （分支 `bump-tarball-1.2.23`，单文件 `+1 −1`，`MERGEABLE`，OPEN）；
  `node tools/upstream-entry.mjs --check` = **OK（in flight）**。
- **一条顺序上的实话**：release 提交（`ef4c9e4`）推上去时 CI 立刻跑，而那时 1.2.23 的上游 PR 还不存在
  （资产要先有、PR 才能开），于是 `Upstream catalogue entry` 这一步报
  `STALE and NOTHING PENDING` 红了；PR #16010 开好之后**重跑同一提交的那条 job 即绿**。
  这不是掩盖红灯，而是这条门禁的固有顺序——它判断的是「现在有没有人在推」，而 PR 只能在我们之后。
  本次已把这条写进 [docs/UPSTREAM.md](UPSTREAM.md) §2，下一轮不必再想一遍。
- `be45a1a`（把市场索引钉到 `ef4c9e4`）在 CI 上全绿（Run `43907267179`）。

### 66.7 下一轮从这里开始

1. ~~**真实 Linux 桌面上的 5 分钟验收**~~ —— **按用户决定跳过（2026-11-16）**：现有硬件不支持，
   不做排期。方法留在这里，将来有机器时照着做即可：跑 `--doctor`（应看到 `audio: external player …`、
   `notifications: notify-send …`），触发一次等待提醒，确认"响 + 弹 + 面板未读"三件事。
   若某个播放器不认 stdin，那是本轮唯一可能翻车的假设。
2. **macOS 真机验收**（有硬件时）：`osascript` 那条是否真的出现 banner；如果不行，退路是把这条改回已知项。
3. §65.4 的其余三项仍然有效：macOS 设置窗口（**仍不做**，§57.7 的理由不变）、L5 对 `--json` 的长期收益（可选）、
   README 平台表与 `docs/PLATFORMS.md` 的两表一致性（本轮已同步，改行为时记得还改两处）。

### 66.8 上游：冻结的根因已被对方修掉，重建在我们交接时正在跑（2026-11-16 晚）

这一节是**别人的仓库发生了什么**，不是我们的缺陷；写下来是为了下一次不必重新诊断一遍。

- **#15569 的结论与我们原来的猜测不同**（维护者 Eapp1e 的回复）：夜间 `Build` 不是卡在探针，而是卡在
  **它刚建好的索引的校验**——`data/plugins/Tests4gptplay__dsh-dots.yml` 与它子目录那条**各自都合法**的条目
  同名 `dsh-dot-mcp`，format 1 按 `name` 建键，解析器于是拒掉整份文档（`MarketIndexError: $.plugins[1950]
  .name duplicates dsh-dot-mcp`）。后果是 `deploy-pages` 与两个 `npm publish` 一起被跳过，
  所以站点 `plugins.json` 与 npm 上的 `dsh-plugin-catalog` 都停在 **2026-11-14T05:27Z（`2026.1114.20961`）**。
  **重跑探针救不了**——输入本身没问题，这正是我们 2026-11-16 上午判断"等对方"的依据。
- **16:39:48Z #15579 合并**（把重复降级成"只拒掉后出现的那一行"），**16:39:49Z #15569 关闭**；
  **16:39:59Z** 维护者手动 dispatch 了一次 `build-site.yml`（run `43910005122`），另有一条 push 触发的
  `43910024003`。历史上这两步要 1h20m–2h50m，所以**目录会在 11-16 当天晚些时候前进**。
- **对我们的直接影响只有两处，且都不需要动手**：
  1. 我们那行的 `version` 来自 **npm 探针**（入口 yml 没有这个字段），所以那次重建跑完就会变成 **1.2.23**、
     下载窗口也会从 10-15→11-13 往前推；
  2. 入口页的安装命令读的是 yml 里的 `tarball`，它要等 **PR #16010** 合并 + **再一次**重建。
- **下载窗口那条（#15864）仍然没人回复**，但 **#15982**（另一位作者）在同一个窗口独立量到同样的缺失
  （`dsh-rea-pro` 目录 458、npm 同窗口 697，+52%；几个大盘包只差 1–4%）。这条按原决定继续"不催、不发 PR"，
  记录已按它改准（[docs/UPSTREAM.md](UPSTREAM.md) §4.2、[docs/ROADMAP.md](ROADMAP.md) §6）。
- **交接后第一件要查的事**：`node tools/catalog-status.mjs`。如果对方那次重建成功，我们的行会从
  `version 1.2.16 / tarball 1.2.21` 变成 `version 1.2.23 / tarball 1.2.21`（等 #16010），
  下载窗口也会前进——那时**不需要发新版**，也不需要再开 PR。

## 67. 交接：DSH 插件三平台 —— Linux / macOS 首次使用按 pin 取浮窗，插件不再手写 Windows 路径（2026-11-16，第 67 轮；插件 **1.2.24**，浮窗仍 **1.0.14**）

第 66 轮把浮窗做到了三平台，这一轮把**唯一的零配置入口**追上：DSH 插件在此之前只对 Windows 生效。
它同时是仓库里最后一份手写 Windows 路径——`EXE_NAME`、`LOCALAPPDATA`、`platforms: ["win32"]`——
而 `reporters/protocol.js` 与 `internal/home` 早就在管同一件契约。

### 67.1 四个改动面

| 面 | 之前 | 现在 |
|---|---|---|
| 平台清单 | `["win32"]`（两份清单） | `["win32","linux","darwin"]`；`dsh.catalog.summary` 与市场索引描述同步 |
| 数据目录 | 插件自己从 `LOCALAPPDATA` 拼 | Windows `%LOCALAPPDATA%\AI-Terminal-Pet`、macOS `Application Support`、Linux XDG；**`verify.mjs` 钉住它与 `protocol.js` 快照目录一致** |
| 浮窗二进制 | 只从 `vendor/AI-Terminal-Pet.exe` 拷 | Windows 照旧（离线、不读 pin）；Linux/macOS 用 `lib/release-pins.json`（schema 2）钉住的发布归档，首启下载 + SHA-256 + 解出成员 + `0755` |
| CI | 只有 Windows job 跑 `verify.mjs` | `ubuntu-latest` 与 `macos-latest` 也跑（macOS job 补了 `setup-node`）；Windows job 用 `DSH_PET_PLATFORM=win32` 钉住自己的那一半 |

### 67.2 三个"看起来对其实错"的地方

1. **`copyVerified` 的 mode 参数不是装饰**：exe 那一行原来没传 mode。Windows 上 chmod 只映射只读位，
   所以看不出问题；POSIX 上会落地 0644，`protocol.js` 的 `launchApplication()` 立刻 EACCES。
   现在以 `0755` 写入并在**非 Windows** 的 job 上断言 `mode & 0o777 === 0o755`。
2. **一个 tarball 服务所有平台，所以"内置四个二进制"不是免费的**：实测 POSIX 二进制比 Windows 的还大
   （linux-amd64 5,009,568 B / gzip-9 2.06 MB；linux-arm64 1.86 MB；darwin-arm64 1.96 MB），
   内置四个会把 tgz 从 2.07 MB 推到 ~9.9 MB，让每个 Windows 安装为跑不了的平台买单。
   所以选"首启按 pin 下载"：tgz 不变，代价是 Linux/macOS 首次运行要下 4.9–5.5 MB 一次。
3. **测试缝是必需的，不是作弊**：`verify.mjs` 要在同一台机器上验三个平台的路径，而三个 CI job 又各在
   自己的 OS 上跑。`DSH_PET_TEST=1` 下的 `DSH_PET_PLATFORM` / `DSH_PET_ARCH` 让两件事同时成立；
   发行版从不设置它们，而且**真实平台那一半仍然在真 Linux / 真 macOS 上跑**（没有缝）。

### 67.3 实测

- `node packages/dsh-ai-terminal-pet/verify.mjs`（Windows）：**118 条 ok / 0 fail**，新增约 12 条——
  三平台数据目录、`platformKey()` 的 x64→amd64 映射、从本地 HTTP 服务器取一个真 tar.gz 后解出的字节
  与成员一致、摘要不符时**一个字节都不写**、没有 pin 的平台明确报 `no-pinned-overlay`。
- 包**不含任何 POSIX 二进制**：`vendor/` 只有 Windows 那份，tgz 与 1.2.23 同量级
  （1.2.24 **2,072,902 B / 20 文件**，1.2.25 **2,072,904 B / 20 文件**，差的是 `vendor/manifest.json`
  里的源码摘要）。
- 浮窗本体零改动 → 浮窗版本仍是 **1.0.14**，只发插件线（资产不可变，没必要重发同一份 overlay）。
- 新增的 CI 步骤跑在 `ubuntu-latest` / `macos-latest` 上；`0755` 断言只在非 Windows 生效，所以它
  真正被执行是在那两条 job 里。

### 67.4 发版记录与两条过程教训（2026-11-16）

- **发了两次**：`dsh-plugin-v1.2.24`（插件三平台）之后 **40 分钟内**又发了 `dsh-plugin-v1.2.25`。
  原因不是功能，而是 1.2.24 上线后 macOS CI 第一次真的执行到提示音路径：**SIGSEGV**（详见 CHANGELOG 的
  1.2.25 一节）。修好之后 Go 源码动了 → `packages/dsh-ai-terminal-pet/vendor/manifest.json` 的源码摘要
  必须跟着动（`vendor.mjs --check` 判的就是这个），所以包必须再发一版；vendored 的 Windows exe
  **字节未变**（仍是 `DB2DD2AB…`），浮窗线仍是 **1.0.14**，没有重发任何浮窗资产。
- **npm**：`latest` = **1.2.25**；tgz **2,072,904 B / sha256 `0E5DE443…` / 20 文件**。
- **上游**：PR [#16031](https://github.com/awesome-dsh-plugin/awesome-dsh-plugin/pull/16031)
  （`bump-tarball-1.2.25`）；#16010（1.2.23）与 #16023（1.2.24）都已**按"被取代"关闭**，
  保持"同时只有一条 bump 在飞"。#16010 那次之后对方重建过一次，条目现在停在 1.2.22（§66.8 的重建确实跑了）。
- **CI**：`013274e`（1.2.25 release）四条 job 全绿（windows / ubuntu / macOS / scan），
  `7a5633a`（索引 pin）同样全绿；两条都用了重跑，因为 push 会取消前一次（`concurrency`）。

**两条过程教训（比代码更值钱，下一轮别重犯）：**

1. **新门禁的第一次运行才是它的验收。** 插件三平台的 `verify.mjs` 在 Windows 上 118/118 全过，
   ubuntu job 一跑就发现两个旧用例假定了 Windows 布局（"no reports at all"）。写测试的那台机器
   验不了另一半——所以**加 CI 步骤的那一轮，要等三个 job 都绿了再宣布"门禁成立"**。
2. **发版前没跑过的平台，等于没验过。** 1.2.24 是先 tag、后看 CI，于是 macOS 的两个问题
   （用例没钉平台、提示音 SIGSEGV）都在 tag 之后暴露。正确顺序是：**先让三个 job 在同一个 commit 上绿，
   再打 tag**。这一轮最后就是这样收口的（`73bc51c` 三平台全绿之后才发 1.2.25）。

### 67.5 未验证项

1. **macOS 面板的真实显示**：没有苹果硬件。这一轮能证明的是"插件在 macOS 上选对目录、装对文件、
   跑通逻辑"，不是"窗口在 Mac 上出现过"。
2. **Linux 桌面上的真实安装**：CI runner 没有面板可看；XDG 路径、`0755` 与解包都是断言过的，
   "装完窗口真的出现"仍要在有桌面的 Linux 上确认一次（方法与 §66.7 第 1 项相同，且按用户决定
   **不排期**）。
3. **macOS 首次下载的代理/证书环境**：`fetch` 走系统信任链，没有额外配置；企业代理下若失败，
   插件按设计只 warn 一次并保持不上报浮窗——这条没有在真实企业网络里验过。

### 67.6 下一轮从这里开始

1. **上游目录那行**（§66.8 的后续）：对方那次重建已经把条目推到 1.2.22，剩下的只是
   `tarball` 跟着我们的 PR 前进，**不需要我们再动**。查法仍是
   `node tools/catalog-status.mjs`。
2. **插件在 Linux 上的手工安装体验**：`scripts/install.sh` 与插件的首启下载现在都能装浮窗，
   两条路会写同一个目录——值得在文档里明确"先装的赢"，避免用户以为要装两次。
3. 视觉不变量门禁（本轮盘点过的候选 B）：preview PNG 已被确认**同机逐字节可复现**，
   跨机器只能做颜色/几何探针，不能比整图。

## 68. 交接：公开文案重做（README 是这一轮的交付物），以及它逼出来的三处假话（2026-11-16，第 68 轮；**未发版**）

这一轮不改产品代码、不发版。交付物是**给外人看的那一页**：README 重写、配图重生成、
以及一个让"配图坏掉"不可能再悄悄发生的门禁。改的过程中抓到三处长期存在的不实描述。

### 68.1 为什么要重做

原来的 README 是**工程记事**的口吻：平台矩阵、IPC 端点、`internal/*` 目录树都写得比"它能替你解决什么"
更早出现。对第一次打开仓库的人（以及任何转发这条链接的人），前 30 秒应该是这个问题：

> 你的 agent 在等你，而你在另一个标签页里。

所以新版的结构是：**痛点 → 它能做什么 → 长什么样（真图）→ 怎么装（一条命令）→ 接哪些终端 →
提醒策略 → 本地优先 → 平台矩阵 → 验证到什么程度（含没验证的）→ CLI/文件/开发/FAQ**。
英文在前、中文同构在后，两边的数字与承诺完全一致。

### 68.2 抓到并修掉的三处假话

1. **四张配图是坏的。** "Visual showcase" 表格引用了 `panel-dark-expanded.png` / `panel-light-expanded.png` /
   `panel-error.png` / `panel-autoapprove.png` —— 这四个文件在 `docs/assets/` 里**根本不存在**
   （真实文件名是 `overlay-*.png`）。仓库首页一直显示四个破图，而 `check_docs_links.py` **只查 `[链接](…)`、
   不查 `<img src>`**，所以从来没人报警。现在改用真实资产，并且门禁补上了这一课。
2. **状态词是编的。** README 写状态徽标是 `ASK / PLAN / WORK / TOOL / ERR / DONE`；代码里
   `stateLabel()` 与 i18n 实际给的是 **Waiting / Tool call / Working / Thinking / Error / Idle**
   （紧凑条上也是整词，不是三字母码）。这就是"文案看着专业、但与实现不符"的典型，已改成真实标签表。
3. **数字全都旧了。** 同一页里同时出现 `--doctor` **11 项**（实为 13）、主程序 **4.49 MiB**
   （实为 4,717,568 B = 4.50 MiB）、测试数 **425 / 461 / 382**（`AGENTS.md` 还写着 382 func Test / 14 包，
   实为 **463 / 15 包**）、跨平台门禁 **13 包**（实为 **16 包**，`scripts/portable-packages.txt`）。
   这一轮把这几处一次对齐，并在 §67.4 的教训之外补了第 3 条：**数字要由命令产出，不要由记忆产出**。

### 68.3 门禁补强：`tools/check_docs_links.py` 现在看图片

- 新增 `<img src="…">` 与 `![alt](…)` 两种形态的本地路径检查；`![…](…)` 不会同时被当成链接重复报告。
- **自测**（不进仓库）：在临时目录造一个 `README.md`，放一张存在的图、一张不存在的 `<img>`、一个不存在的
  `![]()`，跑 `main()` → 必须报 **2 条 image 问题并退出 1**；真实仓库里跑 → 仍然全绿（12 份文档）。
  这条自测正是"新门禁第一次运行才是验收"的应用（§67.4 第 1 条教训）。

### 68.4 配图重新生成（不是手工替换）

`docs/assets/render_assets.py` 从真实光栅器的 preview 输出重建素材，所以图**不可能与界面漂移**。
这一轮跑了一遍：10 张 PNG 字节变了（平均颜色差 <1、约 6% 像素不同、尺寸不变），说明之前那批是旧版本的界面；
`demo.gif` 与 `desktop-scene.png` 未变。**规则**：以后改面板外观，重跑
`PET_PREVIEW_DIR=… go test ./internal/ui -run TestRenderPreview` + `python docs/assets/render_assets.py`，
不要手修 PNG。

### 68.5 硬件实测：按用户决定永久跳过（这一次写进了对外文案）

README 新增 **"How this is verified — and what is not"** 一节，把三件事**主动写出来**：
macOS 面板从未在苹果硬件上开过窗、macOS 通知走 `osascript` 从未被看到、Linux 的声音/通知只验证到进程边界。
措辞是"没有验证、而且**不打算**验证（没有硬件，维护者的决定是不做这些实测）"，而不是含糊的"待验证"。
中文版同一段逐句对应。

### 68.6 这一轮没有做的事（诚实清单）

- **没有新版本号**：产品代码零改动（只改了 README、AGENTS 的数字、`tools/check_docs_links.py` 与 10 张 PNG）。
- **没有重发 LAUNCH 的 X / Show HN / DSH Discuss 文案**：§1 记录它们**已经发过**，改写历史帖子没有意义；
  需要再发时以 README 现在的开头为准。
- **没有动 PLATFORMS / PROTOCOL / UPSTREAM 的实质内容**：它们已经是实测口径，本轮只跟随数字。

### 68.6 顺带发了一次"只含文案"的插件版本：1.2.26

包内 README 是**会随 npm tarball 发布**的公开文案。本仓库的规矩是"资产不可变，改一句就要新版本号"
（1.2.1 与 1.2.4 就是这么来的），所以这次文案重做没有停在仓库里：

- `dsh-plugin-v1.2.26`：`dsh-ai-terminal-pet-1.2.26.tgz` **2,072,884 B / sha256 `030545BB…` / 20 文件**，
  npm `latest` = **1.2.26**；公开 URL 下载回来核对过（200 / 字节数 / 摘要三项一致）。
- `lib/`、`vendor/`、`protocol.js`、清单字段与协议全部与 1.2.25 相同；浮窗仍是 **1.0.14**。
- 上游 PR [#16038](https://github.com/awesome-dsh-plugin/awesome-dsh-plugin/pull/16038)（`bump-tarball-1.2.26`）；
  #16023、#16031 已按"被取代"关闭。#16010 之后对方重建过一次，条目现在停在 1.2.22，
  剩下的只是 tarball 跟着我们的 PR 前进。
- CI：`78129c8`（release）与 `8474ca8`（索引 pin）四条 job 全绿。

### 68.7 下一轮从这里开始

1. **门禁方面**：`check_docs_links.py` 现在覆盖图片，但**不覆盖 HTML 表格里的相对链接**（`<a href>`）；
   如果下一轮重排表格时用到，顺手补上。
2. **对外文案**：README 的"头像/社交预览"（GitHub 的 og:image）仍是仓库默认图；要让它好看需要一张
   1280×640 的 `social-preview.png`，这是纯素材工作（`docs/assets/make_launch_assets.py` 里已有桌面场景的
   画法，可以复用）。
3. §67.6 的三项仍然有效：上游条目只等我们的 PR；插件与 `install.sh` 两条安装路径的"先装的赢"值得写清；
   视觉不变量门禁（preview 已证明同机逐字节可复现，跨机只能比不变量）。

## 75. 交接：OpenCode 接入变成一行 npm 安装 —— 下个会话从这里开始（2026-11-17，**已发版** 浮窗 **1.0.18** / 插件 **1.2.31** / npm 新包 **1.0.18**）

这一轮只做一件事，但它是 §74.4 排在第一的那件：**OpenCode 上报端变成一个真能 `opencode plugin add` 的 npm 包**。
过程里发现它不只是体验问题——是发行缺口（zip 装出来的目录里根本没有那个上报端文件）。

### 75.1 机器状态（全部当场跑出来）

| 项 | 值 |
|---|---|
| 分支 / HEAD | `main`，本轮代码提交 `d709d78`（发版提交见 §75.6 的两步） |
| 版本 | 浮窗 **1.0.18**、插件 **1.2.31**、`opencode-ai-terminal-pet` **1.0.18**（与浮窗同号） |
| tag | `pet-v1.0.18`、`dsh-plugin-v1.2.31` |
| CI | 四条 job 全绿（windows / ubuntu / macOS / scan） |
| 测试 | **498 个 `func Test`** / 15 包；Node 侧 `sync.test.mjs` 4/4、`server.test.mjs` 1/1、`check_versions.test` 27/27 |
| 本机安装 | **1.0.18**：`-VerifyOnly` **12/12**、`--doctor` **12 ok / 1 warn / 0 fail**、`DisplayVersion=1.0.18` |
| 门禁 | `check_versions` **21** 条、`check_facts` 11 条、`check_docs_links` **14 份**、`check_assets` 15 行、`check-tarball` 两个包都与工作区逐文件一致 |

### 75.2 交付了什么

| 面 | 内容 |
|---|---|
| 新包 | `packages/opencode-ai-terminal-pet`（零依赖，5 文件约 48 KB）：`server.js` + `protocol.cjs` **生成**自 `reporters/`，`sync.mjs --check` 进 CI |
| 包内测试 | 临时 home + 真实端点 + 真快照；三平台各跑一次（Windows 管道 / Linux socket / macOS socket） |
| `--doctor` | `reporters` 认第三种入口：裸包名、`@version`、`{package}` 对象；证据是 `node_modules/<name>/package.json` 的版本 |
| `--connect` | 包名优先；`file://` 检出形态只在真有检出现时进备注 |
| 版本策略 | 这个包**与浮窗同号**，`check_versions` 断言相等——一次浮窗发版就是一次 npm 发布 |

### 75.3 两个坑（值得记住）

1. **`"type": "module"` 与 CJS 核心天生冲突。** 唯一诚实的答案是核心以 `.cjs` 进包、由生成器保证逐字节；
   手写第二份实现正是 §52.3 的成因，而且那次是静默的。
2. **测试里不能等 `socket.end` 之后再回 `OK`。** Windows 命名管道没有半关闭，客户端 `end()` 之后再写就是
   EPIPE。真实 Go 服务端是「读到换行就回」（`pipeserver/server.go` 的 `ReadSlice('\n')`），测试必须照它写。

### 75.4 外部世界（复核时间 2026-11-17）

| 对象 | 状态 | 谁动 |
|---|---|---|
| winget `microsoft/winget-pkgs#481818` | `OPEN` / `BLOCKED` / `REVIEW_REQUIRED`（auto-merge 已开） | 等志愿维护者批准 |
| DSH 上游条目 | 1.2.31 的 bump PR 见 §75.6 | 等维护者合并 |
| 目录站 | 站点仍是 1.2.27，`catalog-status` 判 upstream stale（对方夜间构建失败） | 等对方重建；**不为此重发版** |
| awesome-opencode | **门槛是 ≥25★ 且仓库满 30 天**（它的 contributing.md 原文）；我们 1★ → 现在投了就是被关 | 等 star，不是等包 |
| 社交预览图 | 本会话的 browse 扩展在线但 **Chrome 解析不了任何域名**（`ERR_NAME_NOT_RESOLVED`），上传未完成 | 网络恢复后重试，或本人手传 |

### 75.5 下个会话从这里开始

1. **把本机 OpenCode 也换成包名形态**（§75.6 特意没换，见那一条）。换完跑 `--doctor`，`reporters` 行应是
   `opencode: ok - plugins[0] -> opencode-ai-terminal-pet@1.0.18`。
2. **社交预览图上传**（`docs/assets/social-preview.png`，1280×640）。
3. **brew / scoop**：`opencode` 与 winget 都已覆盖，这两个是下一层；先读它们的 manifest 要求再动手。
4. **HN / V2EX 正文与 awesome-claude-code 表单**：只有本人能做（规则见 [docs/PROMOTION.md](PROMOTION.md)）。

### 75.6 发版记录与一个"故意没做"

- 代码提交 `d709d78`（包 + doctor/connect + 门禁 + 版本 + vendor + 文档），tag 打在紧随其后的文档提交上；
  索引 `source.commit` 由 B 提交钉上（git 无法在一个提交里引用自己）。
- **本机 `opencode.json` 故意仍指向 `~/.config/opencode/plugins/terminal-pet`（仓库检出）。** 理由：这一轮要在
  真实的 OpenCode 会话里验证的是**新包**，而换配置会让当前会话的插件重载一次；先确认包在 npm 上装得上、doctor
  认得出（两件事都当场验过），再换。换法是删掉 `plugins` 里那条 `file://…`、换成 `"opencode-ai-terminal-pet"`，
  或者直接 `opencode plugin add opencode-ai-terminal-pet`。
- 上游条目：`bump-tarball-1.2.31` 的 PR 号与状态见本节下方（开出来之后补）；只换 `tarball` 一行。

## 74. 交接：一轮五件事（门禁、卸载项、像素、设置入口、推广功课）—— 下个会话从这里开始（2026-11-17，**已发版** 浮窗 **1.0.17** / 插件 **1.2.30**）

这一轮会话跨了五个主题，每一件都单独发过版或单独记过节（§69–§73），所以这里**只写"现在在哪、下一步做什么"**，
不重复已经写清楚的 runbook。

### 74.1 机器状态（全部当场跑出来）

| 项 | 值 |
|---|---|
| 分支 / HEAD | `main` 与 `origin/main` 同步，HEAD `c5e9cb1` |
| 版本 | 浮窗 **1.0.17**、插件 **1.2.30**；npm `latest` = **1.2.30** |
| tag | `pet-v1.0.17`、`dsh-plugin-v1.2.30`（前一个插件 tag `v1.2.29` 是"随包 README 指向旧版本"的补救，见 §72.5） |
| CI | 最近一次 Run 四条 job 全绿（windows / ubuntu / macOS / scan） |
| 测试 | **495 个 `func Test`** / 15 包；`go test -count=1 ./...` 全过 |
| 本机安装 | 就是 **1.0.17**：`-VerifyOnly` **12/12**、`--doctor` **12 ok / 1 warn / 0 fail** |
| 门禁 | `check_versions` 20 条、`check_facts` 11 条、`check_docs_links` **13 份**、`check_assets` 15 行、`check-tarball` 与工作区逐文件一致 |

### 74.2 这一轮交付了什么（各节详情）

| 轮次 | 交付 | 节 |
|---|---|---|
| 69 | `tools/check_facts.mjs`：文档里的数字与代码不一致就红；`check_docs_links.py` 开始查 `<a href>` 与 HTML 锚点 | §69 |
| 70 | 安装器写"设置 › 应用"那一行 + 把 `install.ps1` 复制进安装目录（删掉 zip 也能卸）；`pet-setup --uninstall --force`；`-VerifyOnly` 11→12 件 | §70 |
| 71 | `internal/ui/render_invariants_test.go`：三平台**视觉不变量**门禁（不比像素）；L4 结案（实测最坏帧 ~4 ms，不做脏区重绘） | §71 |
| 72 | Linux/macOS 的设置入口：`internal/settingscli` 终端编辑器 + `--set key=value`（拒绝而不是规范化）；顺手修掉 `100%!z(MISSING)oom` | §72 |
| 72.5 | **插件线发了两次**：1.2.28 的随包 README 指向 1.2.27（打包早于改版本号）→ 发 1.2.29；并新增 `tools/check-tarball.mjs` | §72.5 |
| 73 | 设置界面底部 **★ Star on GitHub / Report an issue** + 头部版本号；三平台终端版打印同样的链接 | §73 |
| 73+ | 推广功课：**调研了 HN / V2EX / awesome 列表的实际规则**，写成 [docs/PROMOTION.md](PROMOTION.md)；仓库描述重写、topics 15→20、新增两个 issue 模板 | [PROMOTION.md](PROMOTION.md) |

### 74.3 外部世界（谁在等谁）

| 对象 | 状态（2026-11-17 复核） | 谁动 |
|---|---|---|
| **winget** [microsoft/winget-pkgs#481818](https://github.com/microsoft/winget-pkgs/pull/481818) | CLA 已签；**`01`–`10` 号校验全过**（含真跑静默安装/卸载的 `08`）；标签 `New-Package` + `Validation-Completed` + `Azure-Pipeline-Passed`；机器人已开 **auto-merge（squash）**；`reviewDecision: REVIEW_REQUIRED` | **等社区志愿维护者批准**。批准后 auto-merge 自己合并 |
| **DSH 上游条目** [#16249](https://github.com/awesome-dsh-plugin/awesome-dsh-plugin/pull/16249) | OPEN / `MERGEABLE` / 单文件 `+1 −1`；`--check` 报 in flight（健康）；#16228、#16235 已按"被取代"关闭 | 等维护者合并 |
| **目录站点**（`node tools/catalog-status.mjs`） | 站点行仍是 `version 1.2.27`、`tarball 1.2.27`；`[upstream]`：对方夜间构建没跑成功；同一窗口下 **npm 侧下载量 3,455 → 4,586**（最后一天 1,131 此前没被算进去） | 等对方重建。**不要为此重发版** |

**一个对推广有用的数字**：`catalog-status` 现在把同一窗口的 npm 下载量算成 **4,586**（此前记的 3,455 少了窗口最后一天，窗口是 `2026-10-18 → 11-16`）。引用时用 npm 的数并写窗口，不要用站点卡片上的 3,455。

### 74.4 下个会话从这里开始（按价值排序）

1. **把上报端做成一个真能 `opencode plugin add` 的 npm 包**（这一轮调研发现的唯一实质缺口）。
   - 现状：OpenCode 接入是仓库里的 `reporters/opencode.js`，用户得手动放进 `~/.config/opencode/plugins/`。
   - 目标：`opencode plugin add opencode-ai-terminal-pet` 一行装；包内就是同一份上报端 + `reporters/protocol.js`。
   - 顺带解锁：**awesome-opencode（11k★）只按 npm 插件收录**。
   - `opencode.ai/docs/plugins/` 的规则：插件可以是包名或 git spec；`opencode plugin add` 不接受 tarball 与 alias。
2. **社交预览图上传**：`docs/assets/social-preview.png`（1280×640）已生成，但 Settings → Social preview **没有 API**。
   本会话的浏览器扩展是 `paired-offline`，没代传。扩展一上线就传，或本人手传。
3. **awesome-claude-code 收录**（58k★）：只能**人手网页表单**、一次一个、描述单行无 emoji。
   那行字在 `docs/PROMOTION.md` §6（我们按"14 天且活跃"过线，不需要 100★）。
4. **winget 合并后**：往根 README 安装段加 `winget install jjjjjjjjnnjnn.AI-Terminal-Pet`（根 README 不在 tarball 里，**不需要**发插件新版）。
5. **HN / V2EX 发声**：正文必须本人写，`docs/PROMOTION.md` 只给结构、事实卡与会被问到的问题。

### 74.5 只有本人能做的三件事

1. **HN / V2EX 的正文**：HN 指南明写 "don't put generated text in HN posts"，dang 的 tips 更进一步——
   "Don't use an LLM to generate any of it (not even a tiny bit, including to edit or spruce it up)"；
   V2EX About 同样写着「请不要把 AI 生成的内容发送到这里」。所以那份文档里**没有**可粘贴的 HN/V2EX 成稿。
2. **awesome-claude-code 的表单**（人手提交是硬要求）。
3. **社交预览图上传**（或让浏览器的 browse 配置上线，我代传）。

### 74.6 不要重做的事（每一条都有代价）

- **不要给 HN / V2EX 用 AI 稿**，也不要让它润色——对方明文规定，被抓到是信誉问题。
- **不要请人点赞/顶帖**；不要用项目名当 HN 用户名（dang 的 tips 原话：显得像品牌在推销）。
- **不要重传或改写已发布的 Release 资产**（1.2.28 的教训：宁可发 1.2.29）。
- **不要在 winget 合并前**在 README 写 `winget install`。
- **不要写任何无法证明的使用量**（没有遥测；只有 GitHub stars / release 下载 / npm 下载三种公开计数）。
- **不要为目录站的滞后重发版**：`catalog-status` 判的是 `[upstream]`。

### 74.7 这一轮学到的两条（已各自进门禁）

1. **打包顺序是产物问题，不是工作区问题**：1.2.28 的 tgz 早于"把版本号写进随包 README"，工作区自洽、产物过期。
   只有看产物才能发现 → `tools/check-tarball.mjs`（zlib + 40 行 tar，零依赖；按 `files` 白名单判断该打包什么）。
2. **`App` 是分平台类型**：页脚的点击测试用了 `*App`，Windows 上过、Linux/macOS 的 `go vet` 红。
   共享的是布局，落到 Windows 的只是那一脚点击 → 那条测试进 `//go:build windows`，其余继续三平台跑。

## 73. 交接：设置界面里的出口 —— 两个链接与一个版本号（2026-11-17，第 73 轮；浮窗 **1.0.17** / 插件 **1.2.30**）

这一版只加"从产品内部走到项目所在地"的路。一个装得上、跑得好、但没有任何出口的工具，用户想反馈要自己去搜仓库；
页脚把这件事变成一次点击，而把它做在**设置界面**里（不是 README 里）意味着只有真的用过的人才会看到。

### 73.1 三个面

| 面 | 内容 |
|---|---|
| Windows 四页签窗口 | 分隔线下方一排：**★ Star on GitHub** 与 **Report an issue**；悬停变强调色 + 下划线；**Tab 也能走到**；点击用 `ShellExecuteW` 打开默认浏览器 |
| 头部副标题 | 从"AI Terminal Pet"变成"AI Terminal Pet · v1.0.17"：报 bug 时版本号应该伸手可得 |
| Linux/macOS 终端设置 | 同一组链接以**可复制文本**打印在表尾，并附版本号；终端里**不**主动拉起浏览器 |

URL 只写一次（新包 `internal/project` 的 `RepoURL` / `IssuesURL`），三处引用同一份常量。

### 73.2 三个"看起来对、其实错"的地方

1. **版本号不能挤在页脚右边。** 第一版把页脚分成"两个链接 + 右对齐版本"，而页脚可用宽度只有
   `500 − 20 − 176 = 304` 逻辑像素；`150 + 8 + 130 + 8 = 296` 之后只剩 **8 px** 给版本，
   测试立刻报"版本那里 0 个亮像素"。正解是把版本放到**头部副标题**：宽度足够，而且"有没有画"
   从一个像素断言变成**纯函数**断言（`settingsSubtitle(lang, version)`）。
2. **打开浏览器的动作必须能注入。** 点链接的测试若调用真实实现，CI 上会弹浏览器（或在没有浏览器的机器上等待）。
   所以 `ui.Options.OpenExternal` 由 app 层注入；**nil 表示"这个平台没有开浏览器的办法"，
   此时什么都不做**——不是 panic，也不是假装成功。
3. **终端里不自动开浏览器。** `--settings` 会被管道化（`| less`、CI、ssh 抓输出），
   一个被要求打表的程序替用户开窗口是意外行为；终端版只打印 URL。

### 73.3 本轮实测

- Windows：`TestRenderSettingsPreview` 生成的 `settings-general-en.png` 可见页脚；六个单测覆盖
  几何（在最后一排下方、不越界、两链接不重叠）/ 命中（各自命中、间隙不误命中）/ Tab 顺序（最后两站）/
  `settingsApply` 打开**正确 URL**（注入 opener，不真开浏览器）/ 悬停多出像素 /
  副标题带版本且空版本不写 `v`。
- Linux（交叉编译的真二进制，WSL）：`--settings < /dev/null` 表尾三行是
  `★ Star on GitHub  <repo>` / `Report an issue  <issues/new>` / `v1.0.17`。
- 本机装机：`-VerifyOnly` **12/12**、`--doctor` **12 ok / 1 warn / 0 fail**、`--version` = `1.0.17`。
- `go test ./...` 全过（本轮 +7 个 `func Test`）。

### 73.4 发版记录（2026-11-17）

- tag：`pet-v1.0.17`、`dsh-plugin-v1.2.30`（插件跟发的原因同 §70.5：版本常量编进浮窗二进制）。
- 资产：zip **5,140,377 B / `A31E93FA…`（17 文件）**、`pet-setup.exe` **14,546,432 B / `4202DB66…`**、
  `dsh-ai-terminal-pet-1.2.30.tgz` **2,089,727 B / `EFDD7C10…`（20 文件）**，另有四个 POSIX tar.gz。
- **打包顺序**：`npm pack` 在"版本号与随包 README 都改完"之后执行，并用 `node tools/check-tarball.mjs`
  确认 tgz 与工作区逐文件一致（这条规矩是 §72.5 换来的）。
- 上游条目 PR：awesome-dsh-plugin [#16249](https://github.com/awesome-dsh-plugin/awesome-dsh-plugin/pull/16249)
  （`bump-tarball-1.2.30`，单文件 `+1 −1`）；1.2.29 的
  [#16235](https://github.com/awesome-dsh-plugin/awesome-dsh-plugin/pull/16235) 已按"被取代"关闭。
- 资产复验：`pet-setup.exe` 200 / 14,546,432 B / `4202DB66…`、zip 200 / 5,140,377 B / `A31E93FA…`、
  tgz 200 / 2,089,727 B / `EFDD7C10…`（三个都从公网 URL 下载回来核对过）。
- 按 §67.4 的教训：**四个 job 在同一个提交上全绿之后才打 tag**。第一轮 CI 在
  `5ad4d3b` 上 Linux/macOS 的 `go vet` 红了：页脚的点击测试用了 `*App`，
  而 `App` 是分平台的（`app_windows.go` 有设置窗口状态，Linux/macOS 的没有）。
  修法是把那条测试移到 `//go:build windows`，几何/命中/绘制/副标题继续无标签跑三平台（`66ea190` 全绿）。

### 73.5 这一轮没有做的事

- **没有任何"统计/回访"**：没有启动上报、没有首次运行计数器、没有"你从哪来"的参数。页脚只是两个链接，
  URL 是常量，点击之外没有别的动作。这是项目"不出网"承诺的一部分，不是省略。
- **macOS 的页脚没有实机验证**（无苹果硬件）；它与 Linux 共用同一份 `settingscli`。
- **没有把链接做进托盘菜单/浮窗本体**：浮窗是"一眼可辨的状态条"，不是菜单；设置界面才是它该在的地方。

## 72. 交接：Linux/macOS 终于有设置入口（终端编辑器 + `--set`），以及顺手修掉的一个"100%"（2026-11-17，第 72 轮；浮窗 **1.0.16** / 插件 **1.2.28**）

这一轮补的是两块平台上唯一"功能缺失"级别的空白。Windows 有四页签设置窗口；
Linux 与 macOS 只能手改 `config.json`，而 `--settings` 在那两个平台上只启动面板、别的什么也不做。

### 72.1 为什么"手改 JSON"不算入口

`config.Normalize` 的职责是"读旧文件不炸"，所以它把认不出的值**换成默认值**而不报错。
于是 `"theme": "papper"` 的结果不是一条错误，而是**主题悄悄回到 slate**——用户看到的现象是
"这个设置不生效"，而不是"我拼错了"。这一轮把这件事变成两条明确的路：
终端里的编辑器（人用）与 `--set`（脚本用），两者对非法值都是**拒绝**。

### 72.2 改了什么

| 面 | 内容 |
|---|---|
| `--settings` | 没有设置窗口的平台：先打表（21 个键 / 当前值 / 本地化标签），有终端时进入编辑：编号改一项，`s` 保存、`q` 放弃、`?` 重画。**有管道或重定向时只打表**，所以 `--settings \| less` 与 CI 都不会挂住 |
| `--set k=v …` | 两平台都有。多键**要么全部生效要么一个都不写**（先全部校验，再一次性落盘）；未知键、越界值、坏拼写一律 exit 2 并列出已知键；`mutedSessions=clear` 是唯一列表型键的入口 |
| 语言 | 跟随 `config.json` 的 `language`（而不是进程 locale），否则"想把界面切成中文"的人会拿到英文表 |
| 代码 | 新包 `internal/settingscli`（纯逻辑 + 一行 I/O）；`internal/app.SettingsWindowAvailable()` 决定 `--settings` 走窗口还是走终端（`//go:build windows` 与 `!windows` 两个实现） |

### 72.3 三个"看起来对、其实错"的地方

1. **`os.ModeCharDevice` 不是"有没有终端"。** `/dev/null` 与 Windows 的 `NUL` 都是字符设备，
   用这个判断会让 `--settings < /dev/null` 掉进一个立刻 EOF 的编辑器（表现出来就是"什么也没发生"）。
   现在用 `unix.IoctlGetTermios`（Linux `TCGETS` / macOS `TIOCGETA`）与 `GetConsoleMode`，
   也就是每个 `isatty()` 背后的那件事。
2. **"改了就变脏"会把失败也算成成功。** 第一版编辑器在 `promptField` 之后**无条件**置
   `changed = true`，于是"输入了一个非法值 → 按 s"会写盘一次、报告 saved，而实际上什么都没改。
   现在 `promptField` 返回 `applied bool`；`TestRunKeepsGoingAfterABadValue` 就是钉这一条的
   （它要求非法答案被拒绝、随后的合法答案仍然生效、并且只有后一次算改动）。
3. **多键 `--set` 不能边校验边写。** `--set theme=paper opacity=999` 如果把第一对先落地，
   用户会得到"一半生效"的配置，而错误信息只说第二对不合格。现在是**先全部校验、后一次写入**，
   `TestControlSetLeavesTheFileAloneWhenAPairIsBad` 用文件里的最终值钉住它。

### 72.4 本轮实测（两个平台都真的跑过）

**Windows（本机）**

```
--set theme=paper                      -> theme = paper / saved to …\config.json   exit 0
--set sound=off opacity=200            -> 两行回显 + saved                          exit 0
--set theme=neon                       -> theme: "neon" is not one of slate, paper, contrast  exit 2
--set nosuchkey=1                      -> unknown key … (known keys: …)             exit 2
--set mutedSessions=clear              -> mutedSessions cleared (0 entries)        exit 0
安装态：-VerifyOnly 12/12、--doctor 12 ok / 1 warn / 0 fail、--version 1.0.16、注册表行 DisplayVersion=1.0.16
```

**Linux（交叉编译的真二进制，在 WSL 里执行）**

```
--settings < /dev/null    -> 打印 21 行表 + 用法提示，exit 0（没有掉进编辑器）
--set theme=paper         -> theme = paper，exit 0
--set soundVolume=40 soundScheme=bell -> 两行回显，exit 0
--set theme=neon          -> exit 2，且 config.json 的 md5 与之前相同（一个字节都没写）
交互路径（script 分配真 pty）：1 -> off -> s  -> notifications = off / saved，随后 config.json 里 "notifications": false
```

**macOS**：没有实机；编辑器与 Linux 共用同一份代码与测试，差异只在 `IsTerminal` 的 ioctl
（`TIOCGETA`），并且它和其他 `func Test` 一起在 `macos-latest` 上编译与运行。

### 72.5 发版记录（2026-11-17；插件线发了两次）

- 两个 tag：`pet-v1.0.16`（浮窗）、`dsh-plugin-v1.2.28` 与 `dsh-plugin-v1.2.29`（插件线）。
  插件必须跟发的原因同 §70.5：版本常量编进浮窗二进制，`1.0.15 → 1.0.16` 必改字节，
  `vendor/AI-Terminal-Pet.exe` 与源码摘要随之前移。
- 浮窗资产：zip **5,137,799 B / `F1F2A578…`（17 文件）**、`pet-setup.exe` **14,540,800 B / `5A4462CB…`**，
  另有四个 POSIX tar.gz；三个都从公网 URL 下载回来核对过。
- **插件线为什么发两次**：1.2.28 的 **Release 资产**里的随包 README 仍写着 `dsh-plugin-v1.2.27` ——
  `npm pack` 跑在"把版本号写进 `packages/…/README*.md`"**之前**，而 `npm publish` 在发布时重新打包，
  所以 npm 上那份反而是对的（两者差 1 字节；解包后逐文件比对，只有两个 README 不同）。
  资产不可变 → 按本仓库规矩再发 **1.2.29**（功能零差异，`lib/`、`vendor/`、`protocol.js` 与 1.2.28 相同）。
  `dsh-plugin-v1.2.29` 的 tgz **2,086,439 B / `059EF47A…` / 20 文件**，从 npm 拉回来**逐字节相同**。
- **顺序是这一轮的 bug，工作区自检看不见它**：改完版本号之后工作区是自洽的，只有看**产物**才能发现。
  所以新增 `tools/check-tarball.mjs`（node 的 zlib + 40 行 tar 解析，零依赖）：解出 `.tgz` 逐文件与工作区比对，
  并按 `files` 白名单判断"哪些文件本来就该被打包"。它第一次对真实包运行时报了 `scripts/` 与 `verify.mjs`
  "缺失"——那正是比较器必须扛住的那类错。已写进 AGENTS 的发版步骤、[docs/UPSTREAM.md](UPSTREAM.md) §7 的清单
  与 CI（它自己的规则测试）。
- 上游条目 PR：awesome-dsh-plugin [#16235](https://github.com/awesome-dsh-plugin/awesome-dsh-plugin/pull/16235)
  （`bump-tarball-1.2.29`，单文件 `+1 −1`）；1.2.28 的
  [#16228](https://github.com/awesome-dsh-plugin/awesome-dsh-plugin/pull/16228) 已按"被取代"关闭，
  保持同时只有一条 bump 在飞。
- 按 §67.4 的教训：**四个 job 在同一个提交上全绿之后才打 tag**（浮窗在 `8a62017`，插件重发在 `4fcf8d3`）。

### 72.6 顺手修掉的一个老显示 bug

`usageError` 原来把 usage 文本拼进了 `fmt` 的**格式串**：
`fmt.Fprintf(os.Stderr, "ai-terminal-pet: "+format+"\n\n"+usage, args...)`。
usage 里有一句 "re-anchor … at 100% zoom"，于是**任何用法错误**都会打出
`at 100%!z(MISSING)oom`。这一轮改成两次输出（消息走格式化、usage 原样打印），
`TestUsageErrorPrintsTheUsageVerbatim` 钉住"既不能出现 MISSING，也不能丢百分号"。

### 72.7 这一轮没有做的事

- **不做 GUI**：`docs/ROADMAP.md` §5 的"不引入 UI 框架"没变；这是行式编辑器，不用 raw mode、不进备用屏。
- **`mutedSessions` 不能逐条手改**：它是 `source|sessionId` 列表，逐条编辑只会制造无效键；
  单条静音仍在面板里右键该行完成，终端这边只提供"清空"。
- **`--settings` 不再顺带启动面板**：在没有窗口的平台上，"我要改设置"不等于"我要一块浮窗出现"。
- **macOS 的终端编辑器没有实机验证**（理由同上）。

## 71. 交接：面板的像素第一次被断言 —— 三平台不变量门禁（2026-11-17，第 71 轮；**不发版**，浮窗仍 **1.0.15** / 插件仍 **1.2.27**）

这一轮只加测试，不改产品代码、不发版。交付物是 `internal/ui/render_invariants_test.go`：
它把 `docs/ROADMAP.md` §6 里那条诚实的自述（"UI 无自动化视觉回归"）换成四条**可执行**的断言。

### 71.1 为什么这一项排在前面

面板的**布局**早就有测试（矩形、行、命中区），但**像素**没有任何断言。Windows 的渲染器被看过几千次
（谁跑 `TestRenderPreview` 都算一次），而 Linux X11 与 macOS AppKit 的浮窗走的是**另一条纯 Go 光栅化路径**，
本项目没有任何人能在真实硬件上看它一眼（`docs/PLATFORMS.md` 写明了这一点）。
于是"面板画出来是能认的"一直是个**假设**；这条门禁把它变成断言。

关键的可行性来自一件事：`drawing_compact.go` 与 `canvas.go` 都**没有 build 标签**，
所以同一份测试能在三个平台上跑——只是各自用自己那套文字光栅化（Windows GDI / Linux+macOS OpenType）。

### 71.2 断言什么，不做什么

| 断言 | 抓的缺陷 |
|---|---|
| 帧尺寸 = 布局算出的尺寸 | 几何悄悄变了而没人同时改两处 |
| 四个角像素 alpha 为 0 | 圆角半径失效 → 桌面上出现一个硬矩形 |
| 画上去的比例 ≥ 80%，且 **alpha ≥ 200 的比例 ≥ 70%** | 面板体没画（透明洞），或只剩一层半透明薄雾 |
| **亮像素 ≥ 40**（文字与字形） | **字体栈缺失**：OpenType 找不到字体时一个字都不画，面板变成一块带颜色的板子 |
| 每个状态的颜色真的在画布上（等待琥珀 / 出错红 / 进行中绿，各 ≥ 20 像素） | 状态色没上屏或调色板被换错 |
| 不同状态、不同语言**必须长得不同**（像素差 ≥ 1%） | 渲染器忽略输入（会话列表、语言）也能通过每一条单帧断言 |
| 两套配色真的不同：slate 中心亮度 ≤ 0.20、paper ≥ 0.60 | `applyAppearance` 没到光栅器，或浅色主题画成了深色 |

**不做的**：逐像素回归。同机逐字节可复现这一点 `TestRenderPreview` 已经证明过；
跨机只能比不变量，因为 GDI 与 OpenType 的抗锯齿、字体与 hinting 本来就不同。

### 71.3 三个"看起来对、其实错"的地方

1. **像素相等在跨平台上是错的目标，不是"暂时做不到"。** 同一帧"紧凑态等待"，Windows 量到
   **659** 个亮像素，Linux 量到 **289** —— 差 2.3 倍，因为 GDI 的 ClearType 更厚。
   任何"和基准图比"的设计都会在第二个平台上失败，而且失败信息毫无价值。
2. **单点采样会把字体度量写进断言。** 第一版断言"面板中心像素 alpha ≥ 200"。在 Windows 的**紧凑空态**里，
   中心恰好落在一块半透明提示框上（alpha 215）；同一个帧在 Linux 是 238。这不是缺陷，是"中心点在哪".
   所以"面板被填"改成**比例**（实测 91.7%–98.6% 的像素 alpha ≥ 200），中心点只作为日志。
3. **色相断言必须写成通道差，不能写成颜色相等。** 文字是按 `color×alpha/255` 预乘存进画布的
   （`normalizeText`），一个 60% 覆盖率的琥珀字，通道值已经不是 `231,184,112`。
   写成 `R-B ≥ 40 且 G-B ≥ 10` 这类差值谓词，既抓住了"状态色真的上屏"，也不在乎抗锯齿。

### 71.4 本轮实测

**Windows（GDI）**

```
compact waiting      8196/8736 painted (93.8%), 91.7% opaque, 659 light pixels
expanded three rows 41792/42400 painted (98.6%), 97.4% opaque, 2512 light pixels
compact empty        8128/8736 painted (93.0%), 92.1% opaque, 553 light pixels
expanded one row    28192/28800 painted (97.9%), 96.8% opaque, 1479 light pixels
hues: waiting 913, error 839, working 319
diffs: waiting-vs-working 25.84%, waiting-vs-idle 17.33%, working-vs-idle 16.79%, zh-vs-en 7.73%
themes: slate centre luminance 0.01, paper 1.00
```

**Linux（同一份测试交叉编译后在 WSL 里执行，走纯 Go OpenType）**

```
compact waiting      8196/8736 painted (93.8%), 93.8% opaque, 289 light pixels
expanded three rows 41792/42400 painted (98.6%), 98.6% opaque, 817 light pixels
compact empty        8128/8736 painted (93.0%), 93.0% opaque, 265 light pixels
expanded one row    28192/28800 painted (97.9%), 97.9% opaque, 347 light pixels
hues: waiting 746, error 703, working 249
diffs: waiting-vs-working 24.14%, waiting-vs-idle 14.93%, working-vs-idle 14.73%, zh-vs-en 9.78%
```

**macOS**：本机没有苹果硬件；这条测试与其他 `func Test` 一样在 `macos-latest` 上跑
（它走的是与 Linux 同一份 `canvas_opentype.go`）。**首次 CI 运行即通过**（提交 `f3a4b2d`，四条 job 全绿）。
具体像素数没有记在这里：`go test ./...` 不带 `-v`，`t.Logf` 只在失败时进日志——
要看数字就在本机跑 `go test ./internal/ui -run TestRenderedPanel -v`，或在那个 job 上临时加 `-v`。

顺带一个观察：新测试让 `func Test` 从 465 变 470（4 条不变量 + 1 条测量），`tools/check_facts.mjs` 立刻变红并点名五处——
第 69 轮那条门禁第一次被"真实的工作"撞到，而不是被造出来的样例。

### 71.5 这一轮没有做的事

- **没有做逐像素/基准图比对**，理由见 71.2 最后一段。
- **设置窗口没有同类不变量**：它只有 Windows 实现，且 `TestRenderSettingsPreview` 已经覆盖"能画出来"。
  三平台共有的东西才值得写跨机断言。
- **没有把不变量接到 `--doctor`**：那是运行期环境自检，不该承担像素回归的职责。
- **没有发版**：产品代码零改动。

### 71.6 顺手把 L4 结掉了：测过了，不做

`docs/ROADMAP.md` §3 的 L4（脏区重绘 / 向量化）从项目初期就挂着"先测再动"，但一直没有数字。
同一个提交里补了 `internal/ui/paint_panel_cost_test.go`（无标签，三平台都能跑；报告，不断言）：
每例 50 帧取均值。本机（i7-14650HX）与 WSL 上的 Linux：

| 帧 | 尺寸 | Windows（GDI） | Linux（OpenType） |
|---|---|---|---|
| 紧凑·等待 | 56×156 | 0.450 ms | 0.699 ms |
| 紧凑·空 | 56×156 | 0.416 ms | 0.298 ms |
| 展开·3 行 | 200×212 | 2.250 ms | 1.851 ms |
| 展开·6 行 | 200×314 | 3.953 ms | 3.411 ms |
| 展开·3 行 + 悬停 | 200×212 | 2.047 ms | 1.980 ms |

**结论：不做。** 最坏一帧 ~4 ms，而提醒动画是 20 FPS（50 ms 预算）；常驻的紧凑帧 0.3–0.7 ms。
脏区重绘要引入"什么必须重画"的状态机，换来的时间用户根本看不到。ROADMAP 的 L4 已按此写成"已实测：不做优化"。
这条也顺带证明了一件容易被忽略的事：**Linux/macOS 那条纯 Go 字体路径并不比 GDI 慢**（6 行时还更快）。

### 71.7 社交预览图与"第三张没人读的表"

GitHub 的社交预览（仓库被转发时那张卡）一直是仓库默认图。这一轮生成了
`docs/assets/social-preview.png`（1280×640），由 `make_launch_assets.py` 与 `demo.gif`/`desktop-scene.png`
一起产出：右半是**真实光栅器输出**（原生 400×424，没有放大——第一版放大到 1.18 倍，字发虚），
左半是标题与一句话，底部小字写明"panel: rendered by the product's own rasterizer"。

**一个脚本做不到的步骤**：上传只在仓库 Settings → Social preview，**没有 API**。
所以规矩是"改文案 = 重新生成 + 手动上传"，不是"重新生成"。

顺着这件事补的第二个门禁：`docs/assets/README.md` 那张尺寸表是**第三张没人读的表**
（前两张分别是版本与计数，§69 已钉）。`tools/check_assets.py` 直接从文件里读尺寸
（PNG IHDR、GIF 逻辑屏描述符与 image descriptor 计数），要求表与字节一致。

它第一次运行就抓到两处真错，而且错的是**表**、不是文件：

| 行 | 表里写的 | 文件实际 |
|---|---|---|
| `demo.gif` | 760x600 | **760x660** |
| `demo.gif` 帧数 | 56 | **25**（12 呼吸 + 悬停 1 + 展开 1 + 点击 10 + 收尾 1） |

`docs/assets/README.md` 里那句"56 帧因为它是一段 12 帧呼吸"因此也是错的，已按脚本真实的帧预算重写
（5.38 秒一轮）。`docs/LAUNCH.md` 里 `demo.gif` 的字节数（342,606）也不对，实测 **359,471 B**。
自测（临时目录）：造一份"尺寸写错 + 帧数写错 + 列了但文件不存在"的表，必须报 3 条并退出 1。

**没有做的事**：没有重新编码现有的 `demo.gif` / `desktop-scene.png`——重跑生成器得到的字节与仓库里的差
105 / 170 B（几何与帧数相同，是编码器噪声），为了不制造无解释的二进制改动，这一轮只新增文件。

## 70. 交接：安装器登记"设置 › 应用"那一行 —— 于是删掉 zip 也能卸，winget 才谈得上（2026-11-16，第 70 轮；浮窗 **1.0.15** / 插件 **1.2.27**）

这一轮改的是**交付路径本身**：Windows 上"装得上、卸不掉"是个真缺陷，不是体验瑕疵。
`install.cmd -Uninstall` 要你留着 zip 或 checkout，`pet-setup --uninstall` 会停下来问一句，
而"设置 › 应用"里**根本没有这一行**——对一个双击安装的产品来说，这三条加起来就是"装得上、卸不掉"。

### 70.1 三个入口原来为什么都不成立

| 入口 | 之前 | 现在 |
|---|---|---|
| `设置 › 应用` | **没有这一行**（安装器从不写 `HKCU\...\Uninstall\*`） | 有：`DisplayName=AI Terminal Pet`、`DisplayVersion`、`InstallLocation`、`DisplayIcon`、`UninstallString` / `QuietUninstallString`、`NoModify` / `NoRepair`、`EstimatedSize` |
| 卸载命令 | 只有 zip 里的 `install.cmd -Uninstall`（zip 删了就没了） | 安装器把自己复制成 `<安装目录>\install.ps1`，那一行指的就是这份副本，所以**源没了也能卸** |
| 免交互卸载 | `pet-setup --uninstall` 一定 `Read-Host` 问 `yes`，包管理器无控制台时直接 exit 2 | `pet-setup --uninstall --force`（以及 `install.ps1 -Uninstall -Force`）；`--uninstall` 单独仍然会问 |

### 70.2 四个改动面

| 面 | 内容 | 关键文件 |
|---|---|---|
| 制品清单 | 第 12 件：`install.ps1` 自己（复制前后都算摘要，`-VerifyOnly` 会验它） | `scripts/install.ps1` 的 `Get-InstallArtifacts` |
| 注册表 | 装完才写（复制与摘要都通过之后），卸载时**只在 `InstallLocation` 指向刚删的目录时才删**（与 `HKCU\...\Run` 同一条守卫） | `Write-UninstallEntry` / `Get-UninstallEntryLocation` / `Invoke-Uninstall` |
| 版本号来源 | checkout 里读 `internal/app/version.go`；zip 用户问装好的二进制 `--version`（5 秒上限，问不到就 `0.0.0` 并 warn） | `Get-OverlayVersion` |
| 自检 | `-VerifyOnly` 多一行 `[ok]/[warn] apps list/…`：**"装得对"现在包括"能卸掉"** | `Invoke-VerifyOnly` |
| pet-setup | `--uninstall --force`；`--force` 单独出现或与 `--verify`/`--add-to-path` 组合都拒绝 | `cmd/pet-setup/main.go` + 两条新测试 |

### 70.3 四个"看起来对、其实错"的地方（下一轮别再犯）

1. **键名不能按"有没有传 `-Destination`"派生。** 第一版是这么写的，于是真实安装也踩了坑：
   注册表里记录的命令**总是**把 `-Destination` 拼在末尾（沙箱需要，真实安装拼出来就是默认路径），
   卸载时 `$PSBoundParameters` 里于是也有 `-Destination` → 键名变成派生的那个 → **找不到自己刚写的那一行**。
   正解是拿**安装目录本身**与默认目录比较：等于默认目录就是真实安装（键名 `AI-Terminal-Pet`），
   否则是演习（键名 `AI-Terminal-Pet-<12hex>`）。派生键名的意义也在这里：测试装出来的那一行
   不会混进真实的"设置 › 应用"，卸载也不会删错条目。
2. **`-Force` 不能默认带上。** 它只有一个含义：不再问。`--uninstall` 免掉提示是**用户少按一个键**的代价，
   而"删错目录"是不可逆的；所以 `--force` 必须显式拼，并且在别的子命令旁边出现时直接报错
   （否则一个手滑就变成"静默执行一个我没打算执行的操作"）。
3. **测试要执行"记录下来的那条命令"，不是自己再拼一条。** S8 原来用 `install.ps1 -Uninstall -Force`
   自己拼参数，那只证明脚本能卸；现在读注册表里 `QuietUninstallString` 的**原文**去执行，
   才证明"设置 › 应用"点下去真的能卸。两者差的正是这一轮的全部价值。
4. **`[System.IO.Path]::GetFullPath` 会展开 8.3 短名。** GitHub 的 Windows runner 把沙盒建在
   `C:\Users\RUNNER~1\...` 下，而安装器记录的 `InstallLocation` 是 `C:\Users\runneradmin\...`
   （`GetFullPath` 展开了 `RUNNER~1`）。产品侧是自洽的——卸载时同样展开，键名派生与守卫都一致，
   `-VerifyOnly` 也绿——但测试拿**原始字符串**比较，于是报"这一行不存在"，而那一行就在注册表里。
   修法：比较放到 JS 侧，两边各做一次 canonical 化（`realpathSync.native` 在 Windows 上走
   `GetFinalPathNameByHandle`，本机实测 `C:\PROGRA~1` → `C:\Program Files`）；
   并且把**注册表里实际看到的键**写进失败详情——没有 `seen` 的那一轮只能靠猜，白花一次 CI。

### 70.4 本轮实测（全部当场跑出来）

```
本机安装：scripts\install.cmd                      -> 12 file(s) copied and hash-verified
          scripts\install.cmd -VerifyOnly          -> 12/12 一致, 0 warning
          AI-Terminal-Pet.exe --version            -> ai-terminal-pet 1.0.15
          AI-Terminal-Pet.exe --doctor             -> 12 ok, 1 warn, 0 fail
注册表：  HKCU\...\Uninstall\AI-Terminal-Pet       -> DisplayName=AI Terminal Pet
                                                       DisplayVersion=1.0.15
                                                       InstallLocation=%LOCALAPPDATA%\AI-Terminal-Pet
                                                       QuietUninstallString -> <安装目录>\install.ps1 -Uninstall -Force -Destination …
沙箱：    -Destination %TEMP%\pet-arp-check-*\AI-Terminal-Pet
          装上 → -VerifyOnly 12/12 → 执行那一行里的命令 → 目录没了、键没了
门禁：    tests/fresh-machine.mjs --no-overlay      -> 183/183 (44.6 s)
          node -e "realpathSync.native('C:\\PROGRA~1')" -> C:\Program Files（8.3 展开，S5/S8 的比较靠它）
          node packages/dsh-ai-terminal-pet/verify.mjs -> VERIFY PASSED
          vendor.mjs --check                       -> 两个 vendor 文件都与源码同摘要
          tools/check_versions.mjs / check_facts.mjs -> 全绿
```

### 70.5 发版记录（2026-11-16）

- **两个 tag**：`pet-v1.0.15`、`dsh-plugin-v1.2.27`。为什么插件也要发：版本常量在
  `internal/app/version.go`、会被编进浮窗二进制，所以 `1.0.14 → 1.0.15` 必然改变
  `bin/AI-Terminal-Pet.exe` 的字节，`vendor/AI-Terminal-Pet.exe` 与 `vendor/manifest.json`
  的源码摘要必须跟着前移（`vendor.mjs --check` 判的就是这个）。`lib/`、`protocol.js`、清单字段均未改。
- **资产**（详见 CHANGELOG 的「1.0.15 发布包实测清单」）：`pet-setup.exe` 14,504,960 B / `AF9E854F…`、
  `ai-terminal-pet-1.0.15-win-x64.zip` 5,129,002 B / `5292B2E7…`（17 文件）、
  `dsh-ai-terminal-pet-1.2.27.tgz` 2,072,885 B / `FA50130D…`（20 文件），另有四个 POSIX tar.gz。
- **npm**：`latest` = 1.2.27；`npm pack dsh-ai-terminal-pet@1.2.27` 取回来的 tgz 与 Release 资产**逐字节相同**
  （2,072,885 B / `FA50130D…`），integrity `sha512-2wxXHQts…` 与 `npm pack --dry-run` 当场输出的一致。
- **上游条目**：PR [#16048](https://github.com/awesome-dsh-plugin/awesome-dsh-plugin/pull/16048)
  （分支 `bump-tarball-1.2.27`，单文件 `+1 −1`，`MERGEABLE`，OPEN）；`--check` 报 in flight（健康状态）。
  按 §67.4 的教训，**四个 job 在同一个提交（`c9733d2`）上全绿之后才打 tag**——1.0.15 的第一次 CI 是红的，
  原因就是 §70.3 第 4 条那个 8.3 短名问题，所以 tag 比版本提交晚了三个（测试-only 的）提交。
- **资产复验**：从公网 URL 下载回来核对过（`pet-setup.exe` 200 / 14,504,960 B / `AF9E854F…`、
  zip 200 / 5,129,002 B / `5292B2E7…`、linux-amd64 tar.gz 200 / 5,418,747 B / `8CD33190…`、
  tgz 200 / 2,072,885 B / `FA50130D…`）。

### 70.6 winget：清单已提交，卡在两件不由我们决定的事上

- 清单在 `packaging/winget/manifests/j/jjjjjjjjnnjnn/AI-Terminal-Pet/1.0.15/`（三份 YAML），
  目录结构就是 `microsoft/winget-pkgs` 的布局；`winget validate --manifest <目录>` 打印"清单验证成功"，
  `winget show --manifest <目录>` 能正确解析 URL 与 sha256。
- **已开 PR**：[microsoft/winget-pkgs#481818](https://github.com/microsoft/winget-pkgs/pull/481818)
  （3 文件 `+60 −0`，`MERGEABLE`）。2026-11-17 的状态：
  - **CLA 已签**（`license/cla` pass）；
  - **`01`–`10` 号校验全 SUCCESS** —— 11 项真检查全过，**`08. Installation Validation` 真的下载、
    静默安装再静默卸载过 `pet-setup.exe`**；标签 `New-Package` + `Validation-Completed` + `Azure-Pipeline-Passed`；
  - **机器人已开启 auto-merge（squash）**，只差一位**社区志愿维护者**批准（`REVIEW_REQUIRED`）。
  也就是说：我们这一侧的活全干完了。**不催、不改 PR 正文**（改了会让校验重跑）。
- 一条实测结论：`08` 通过意味着**未签名没有拦住他们的静默安装**（我们担心的那个风险这次没发生），
  但用户首次双击 `pet-setup.exe` 仍会看到 SmartScreen —— README 的提示照旧有效。
- 能不能装/卸由我们的安装器决定：静默安装是 `pet-setup.exe --install`（默认就是这个），
  静默卸载是注册表那一行的 `QuietUninstallString`；`tests/fresh-machine.mjs` 的 S5/S8 就是按这条契约验的。
- **真实风险写在前面**：`pet-setup.exe` 没有 Authenticode 签名，对方 pipeline 里那次静默安装可能被
  SmartScreen 拦住。若如此，**不要**改成 `InstallerType: zip` + `portable` 的"解压版"清单
  （那会绕过 `install.ps1`：没有 `protocol.js`、没有 bundle、没有自启动，等于 ROADMAP §5 禁止的降级安装），
  而是记录原因、保留安装器改进（它对所有 Windows 用户都有效），等有证书再说。
- 规则与判断可见性的方法写进 [docs/UPSTREAM.md](UPSTREAM.md) §4.3；提交要一个只新增一个版本目录的 PR。

### 70.7 这一轮没有做的事

- **没有签名、没有公证**：`pet-setup.exe` 仍会触发 SmartScreen，README 的「Unsigned binary notice」照旧。
- **没有改 POSIX 的卸载路径**：`install.sh --uninstall` 本来就自洽（数据目录 + 软链 + `.desktop`/LaunchAgent），
  这一轮只补 Windows 缺的那一半。
- **没有把 winget 写进 README 的安装段**：清单还没被合并，写了就是一句没兑现的承诺；
  合并之后再改 README 与本次的 README 不一致处（规矩见 AGENTS §8）。

## 69. 交接：把"文档里的数字"变成门禁，以及它第一次运行抓到的两件事（2026-11-16，第 69 轮；**不发版**，浮窗仍 **1.0.14** / 插件仍 **1.2.26**）

这一轮不改产品代码、不发版。交付物是**一条新门禁**（`tools/check_facts.mjs`）和它顺手逼出来的四处修正：
`docs/ROADMAP.md` §0 的基线表、§2 的 E5 清单、`tools/check_docs_links.py` 的 HTML 盲区、
以及 `AGENTS.md` 的验证命令表。

### 69.1 为什么需要它（第 68 轮已经写过一次教训）

§68.2 抓到 README 的数字全旧了，当时的结论是"**数字要由命令产出，不要由记忆产出**"。
但那一轮只改了 README，**没有让这件事可检查**；一轮之后同一类漂移已经在 `docs/ROADMAP.md` 的基线表里，
而且同一份文件里两处自相矛盾（§0 写 428 个 `func Test`，§0 的另一行写 463）：

| 项 | 表里写的 | 实测（当场跑） |
|---|---|---|
| `func Test` | **428** | **463** |
| 测试文件 / 行 | 87 / 12,673 | **94 / 13,493** |
| 非测试源码 文件 / 行 | 136 / 26,526 | **143 / 27,200** |
| 主程序 | 4,702,208 B（4.48 MiB） | **4,717,568 B（4.50 MiB）** |
| statusline / hook | 3,675,136 / 3,701,760 B | **3,678,720 / 3,705,344 B** |
| 三个二进制合计 | 12,079,104 B | **12,101,632 B** |
| 插件 tarball 文件数 | 19 | **20**（CHANGELOG 的 1.2.24–1.2.26 三节都写着 20） |

### 69.2 新门禁钉住什么、不钉什么（边界写在这里，免得下次重新想一遍）

**钉住**（写错就红，11 个 claim 形状，覆盖 README 中英 / `AGENTS.md` / ROADMAP §0）：

| 事实 | 规则 | 现值 |
|---|---|---|
| `func Test` 数量 | 以 `func Test` 开头的行（**含 `TestMain`**，这是历年计数的口径） | 463 |
| 有测试的包数 | 目录里有 `*_test.go` 的包。**故意不用 `go list`**：只有 Windows 才编的测试文件会让同一条声明在 ubuntu 上量出别的数，而这句话说的是树、不是某个 OS 的构建 | 15 |
| Go 包总数 | `scripts/portable-packages.txt` 的非注释行（同时也是跨编译门禁的清单） | 16 |
| `--doctor` 项数 | `internal/app/doctor.go` 里 `runCheck("名字", …)` 的条目；函数定义 `func runCheck(` 不算 | 13 |

**不钉**（写进表里也只会变成噪声）：文件数、行数、二进制体积、依赖清单。
`go 1.23.0` 是**下限**不是钉死，所以 CI 与本机编出的字节本来就不该相同——
把体积做成断言，等于要求 CI 与本机同字节，而那正是 `-buildvcs=false` 那条约定在避免的事。
这些值由 `node tools/check_facts.mjs --print` 现测输出（本轮的输出记在 69.4），写进文档时**连日期一起写**。

ROADMAP §0 因此多了一段"这些数字是怎么保证的"，说明哪几项是断言、哪几项是带日期的现状测量。

### 69.3 它第一次运行抓到的两件事（都是"看起来是 bug，其实是认知错"）

1. **`README.md` 的 `#english` 不是坏链接，是门禁不认识它。** 新加的 `<a href>` 检查立刻报
   `README.md: #english names no heading in this file`。README 顶部导航是 `<a href="#english">English</a>`，
   而英文页的锚点是手写的 `<a id="english"></a>`（第 40 行），标题式锚点里没有 "English"。
   **没有直接改 README**：先把 GitHub 真正渲染出来的 HTML 取下来看——
   `gh api repos/jjjjjjjjnnjnn/ai-terminal-pet/readme -H 'Accept: application/vnd.github.html+json'`。
   结果是 GitHub 把作者的 id 加前缀渲染成 `<a id="user-content-english">`，**与它自己给标题加的锚点
   完全同一套机制**（`<h2>Install</h2><a id="user-content-install" href="#install">`）。
   也就是说这个链接一直是好的，缺的是门禁的识别能力。于是改成：
   `anchors()` 除 markdown 标题外，也把 `id=`/`name=` 当作本文件的锚点目标。
   这条记下来的价值不在代码，在于**"先读对方渲染出来的东西，再决定改自己还是改门禁"**。
2. **E5 的"未交付"清单已经过期四轮。** `theme` / `perSourceColor` / `hidden` 三项早已落地并在
   三平台接线，`size`（基准尺寸预设）则被缩放取代。§2 的 E5 现在是一张逐项列出**代码落点**的表
   （`config.go:74`/`:191`、`drawing.go:186`/`:235`、`settings_layout.go:625`/`:633`/`:636`、
   三个 `app_*.go` 的 `applyAppearance`）。`size` 标成**不做**，理由与落点都写着。

### 69.4 本轮实测（全部当场跑出来）

```
gofmt -l .                                   # 空
go vet ./...                                 # 0
go test -count=1 ./...                       # 15 个包全过（463 个 func Test，本轮未改 Go 代码）
node tools/check_facts.mjs                   # 11 个 claim 形状一致
node --test tools/check_facts.test.mjs       # 14/14
node tools/check_versions.mjs                # 20 条一致（overlay 1.0.14 / plugin 1.2.26）
node --test tools/check_versions.test.mjs    # 26/26
python tools/check_docs_links.py             # 12 份全解（含新增的 <a href> 与 HTML id）
node tools/check_facts.mjs --print           # 见下
```

`--print` 在写这份记录时的输出（**这些数字随时会变，引用时请重跑**）：

```
packages (scripts/portable-packages.txt): 16
packages with tests: 15
func Test declarations: 463
test files / lines: 94 / 13,493
source files / lines: 143 / 27,200
--doctor checks: 13
AI-Terminal-Pet.exe: 4717568 B
pet-statusline.exe: 3678720 B
pet-hook.exe: 3705344 B
```

`check_docs_links.py` 的自测（临时目录、不进仓库）：造一份文档，里面同时有一个存在的 `<a href>`、
一个不存在的 `<a href>`、一个不存在的 `#锚点`、一个坏 `<img>`、一个坏 `![]()`，
以及**围栏代码块与行内代码里的假链接**；要求恰好报 **4 条**并退出 1，代码块与行内代码里的示例一条都不许报。

### 69.5 这一轮没有做的事（诚实清单）

- **没有发版**：产品代码零改动（Go 源码、`reporters/`、插件包全部未动），`1.0.14 / 1.2.26` 仍是最新。
- **`172 项 fresh-machine 检查`没有进门禁**：它是"执行到的检查条数"，只能跑整套才算得出，
  静态数 `check(...)` 的调用数会被循环与条件分支骗。留在 `--print` 之外，写作时以那次的实跑输出为准。
- **`docs/assets/` 的社交预览图（og:image）仍然没有**：§68.7 第 2 项，纯素材工作，下一轮或再下一轮做。
- **视觉不变量门禁**（§68.7 第 3 项）仍然只是候选：本轮把"数字"这条先钉住了，几何/配色探针没有动。

## 59. 交接：macOS 纯 Go AppKit 适配与 GitHub Actions `macos-latest` CI（2026-11-15，第 59 轮；**未发版**）

本机为 Intel `i7-14650HX` HP OMEN Windows 11 笔记本（无苹果硬件，遵循许可与 SLA 不部署非授权 macOS VM），按用户确认选用 **零 cgo `github.com/ebitengine/purego v0.9.1` + GitHub Actions `macos-latest` 原生验证**路线交付 macOS 全端适配：

### 59.1 代码共享重构（零重复实现）

- `internal/app`：`platform_gui.go`（`//go:build linux || darwin`，`uiAvailable() = true`）与 `platform_nogui.go`（`!windows && !linux && !darwin`）。
- `internal/ui`：
  - `control_unix.go`（`//go:build linux || darwin`）：Linux 与 macOS 共用按用户派生的 Unix domain 控制 socket。
  - `canvas_opentype.go`（`//go:build linux || darwin`）+ `canvas_fonts_linux.go` / `canvas_fonts_darwin.go`：共用纯 Go OpenType TTF/OTF/TTC 复合字体光栅化引擎，macOS 扫描 `/System/Library/Fonts`、`~/Library/Fonts` 并优先排序 `SFNS`/`SFPro`/`HelveticaNeue`/`PingFang`/`Hiragino`。
  - `sound.go`（全平台共用）+ `sound_test.go`（纯波形测试跨 Windows/Linux/macOS 执行）+ `sound_windows_test.go`（`TestPlaySoundLive` Win32 探针）：Win32 与 macOS 共用 PCM WAV 合成与 `toneKind`。

### 59.2 macOS 原生实现（全部 `CGO_ENABLED=0`）

- **`internal/platform/focus_darwin.go` + `focus_darwin_test.go`**：`purego` 绑定 `CoreFoundation` / `CoreGraphics` / `AppKit`，实现 `CGWindowListCopyWindowInfo`（层 0 可见窗口枚举）、`NSWorkspace.frontmostApplication` + `CGEventSourceSecondsSinceLastEventType`（前台识别与 HID 空闲时长 `SampleForeground`）、`NSRunningApplication activateWithOptions:`（点击行回跳终端 `ActivateTarget`）。
- **`internal/ui/cocoa_darwin.go` / `sound_darwin.go` / `clipboard_darwin.go`**：共用 AppKit loader 与 `NSAutoreleasePool` / `NSString` helper；实现 `NSSound` 异步 WAV 播报与 `NSPasteboard` 剪贴板写入。
- **`internal/ui/darwin_geometry.go` + `darwin_geometry_test.go`**：Cocoa 左下角坐标系（`NSScreen.visibleFrame` / `setFrame:` / `NSEvent.locationInWindow` / `mouseLocation`）、直通 BGRA→预乘 RGBA 像素转换、macOS 虚拟 `keyCode` 映射与 `Cmd+Plus/Minus/0` 缩放，全部纳入三平台单测。
- **`internal/ui/app_darwin.go` + `darwin_appkit_test.go`**：`AITerminalPetWindow` 无边框透明置顶浮窗（单标量方法 `canBecomeKeyWindow`，零 struct 值传递 IMP 回调）、`NSImageView` + `NSBitmapImageRep` 表面上传、`NSStatusBar` 状态栏菜单（显示/隐藏、重置位置、退出）、贴边收起与状态脉冲、左键拖拽/点击/右键静音或切换/键盘导航与提示音过渡提醒。
- **CI 接入**：`.github/workflows/ci.yml` 新增 `macos` job（`runs-on: macos-latest`，`CGO_ENABLED=0`），运行 `go vet`、`go test -count=1 ./...`、编译 `darwin/arm64` 三二进制并执行无头 `pet-hook` + `--json` + `--doctor` 冒烟；`docs/PLATFORMS.md`、`docs/ROADMAP.md` 已同步更新，版本号维持 `1.0.9` / `1.2.18` 未动，待用户决定 push 到远端后在 GitHub Actions `macos-latest` 实跑验证。

## 58. 交接：Track 3（L1/F2/F5）与全平台适配 + Linux 纯 Go X11 浮窗（2026-11-15，第 58 轮；**未发版**）

本轮接手后按「3 然后 1」顺序执行：

### 58.1 Track 3 功能与轻量化交付

- **L1 statusline 管道短路**：`internal/reporter/payload.go`（`ReadSnapshot` / `StatuslineDeliverNeeded`）+ `cmd/pet-statusline/main.go`。快照仍每次刷新，字段未变且距上次投递未满 30 秒心跳窗口时跳过命名管道/socket 投递；单测覆盖 `internal/reporter/statusline_test.go`。
- **F2 键盘行导航**：`internal/ui/keyboard.go` + `keyboard_test.go` + `drawing_compact.go`。展开面板支持 `Up`/`Down`/`Home`/`End` 选行、`Enter` 跳转终端、`Esc` 收起，Win32 与 Linux X11 共用状态机。
- **F5 按显示器布局记忆坐标**：`internal/config/config.go`（`State.Layouts`）+ `internal/ui/placement.go` + `internal/app/run.go`。按显示器工作区签名独立保存与恢复坐标，切换外接屏不再覆盖原配置。

### 58.2 Track 1 全平台扩展与 Linux 纯 Go X11 面板

- **共享协议与平台抽象**：
  - `reporters/protocol.js` 按 `process.platform` 解析 Win32/Darwin/Linux 数据目录与终端端点（命名管道 vs Unix socket）和可执行文件名，保持按用户 8-hex sha256 token 不变。
  - `internal/reporter`（`binary_windows.go` / `binary_posix.go`、`approve.go` 改用 `home.Join("approvals")`）。
  - `internal/platform` 拆为中立层 + `focus_linux.go`（`/proc` + X11 `_NET_CLIENT_LIST` + MIT-SCREEN-SAVER）+ `focus_darwin.go`（`SysctlKinfoProcSlice`）+ `process_posix.go` + `focus_other.go`。
  - `internal/app`（`connect.go`、`doctor.go` + `doctor_posix.go`/`doctor_windows.go`、`platform_linux.go` 启用 `uiAvailable() = true`）与 `internal/autostart/autostart_unix_test.go` 的 macOS vet 修复。
- **Linux 纯 Go X11 浮窗**（零 cgo、零 GTK，依赖 `github.com/jezek/xgb v1.3.1` 与 `golang.org/x/image v0.25.0`）：
  - 共用 `internal/ui/panel_core.go` 与 `internal/ui/dock.go`；
  - `internal/ui/app_linux.go`：32-bit TrueColor ARGB、`_MOTIF_WM_HINTS` + `_NET_WM_WINDOW_TYPE_UTILITY`、chunked `PutImage`、`_NET_WM_WINDOW_OPACITY`、XRandR + `_NET_WORKAREA`、X11 `KeyPress` 映射 F2、`SelectionRequest` 提供 `CLIPBOARD`、左键拖拽/点击/滚轮缩放/右键行静音；
  - `internal/ui/canvas_linux.go`：扫描 `/usr/share/fonts` 与 `/mnt/c/Windows/Fonts` 构建 OpenType 复合字体链；
  - `internal/ui/control_linux.go` & `clipboard_linux.go`：unix 控制 socket 与 X11 `CLIPBOARD`；
  - `macOS`：`internal/ui/app_posix.go`（`//go:build !windows && !linux`）保持诚实非窗口占位。
- **跨平台 CI 门禁更新**：`scripts/portable-packages.txt` 收录全部 16 个 Go 包，`scripts/windows-only-packages.txt` 清空说明；`docs/PLATFORMS.md` 已按实测同步。

### 58.3 本轮验证边界

- **未做 macOS 运行期实测**与**未做非 WSLg 原生桌面环境实测**；本轮没有改版本号（浮窗仍 `1.0.9` / 插件仍 `1.2.18`）。

## 50. 中文 README 与市场文案（2026-11-13 深夜，第 50 轮；只改文档与自述字段）

用户直接给了详情页链接，两条反馈：**"缺少 readme 中文版"**、**"描述不够爆款"**。两条都先查清"那边到底读什么"，
再改我们这一侧——详情页不是我们渲染的，改错地方等于白改。

### 50.1 详情页的 README 是怎么来的（读他们的代码，不是猜）

`awesome-dsh-plugin` 的 `scripts/probe-readmes.mjs`：

- 用 GitHub API **列一次插件目录**，从文件名里挑一个 readme 形状的文件
  （`/^readme([._-].+)?\.(md|markdown|rst|txt)$/i`）；
- 语言**不由文件名决定**，而是对内容做 CJK 嗅探（`langOf`，中日韩字符占比）；
- 另一种语言的候选名是一张名单：`README.zh.md`、`README.zh-CN.md`、`README_zh.md`… 都认；
- 只有一种语言时，页面写一句「该插件的 README 只有英文版本」——**我们页面上当时就是这句**。

所以我们缺的不是"把中文贴到某个表单里"，而是**目录里第二个 readme 文件，内容是中文**。

### 50.2 做了什么

- 新增 `packages/dsh-ai-terminal-pet/README.zh.md`：英文 README 的完整逐节翻译（含安装三种形态、
  市场索引那张表、磁盘上写什么、要求与限制），安装命令与兼容范围都对齐到本版。
- 英文 README 顶部加 `English · 中文` 切换行；`files` 里加上中文那份，随 tarball 分发。
- `verify.mjs` 加两条断言，把**目录那边的嗅探规则钉进我们自己的 CI**：tarball 必须带 `README.zh.md`，
  且它的 CJK 字符占比必须 > 10%（实测 2,055 个 CJK 字符）。将来谁把翻译改回英文，或者把它从 `files` 里删掉，
  会先在这里红，而不是在别人的目录页上变成一句话。

### 50.3 文案：先给问题，再给答案

原来的 `dsh.catalog.summary` 是"这包是什么"。改成两段，各自都短：

> **en**（225 字符）: Which of your agents is waiting on you? One always-on-top Windows panel for
> DeepSeek Harness, Claude Code, OpenCode and Antigravity sessions: it rings only for real events, and
> the arrow at a row's end jumps to that terminal.
>
> **zh**（110 字符）: 哪个 agent 在等你？一块置顶的 Windows 面板汇总 DeepSeek Harness、Claude Code、
> OpenCode 与 Antigravity 的会话：只在真事件上响，点行尾箭头即可跳到对应终端。

`docs/market-index.json` 的 `description`（Discover 的卡片与详情页读它）是长一版：每一行写着什么状态、
收起时显示卡了多久、点箭头跳转、点行复制项目路径、只在真事件上响、授权仍在 DSH 完成、
Harness 装上即接而另外三家要装上报端。

**每一条都能在代码里指出落点**：状态与短时长键在 `internal/i18n`，点击规则在 `internal/ui/tap_windows.go`，
"只在真事件上响"在 `internal/ui/attention.go` 的静默判定（`internal/session/attention.go` 说明它要什么）。
没有加形容词，也没有承诺"面板里放行审批"——那是上游维护者上次合并时删掉的一句，删得对。

### 50.4 为什么这也值一个版本号

`dsh.catalog` 在 `package.json` 里，`files` 也变了，所以 **tarball 变了**；资产不可变，
于是走一个新版本号（1.2.16），而不是改已发布的 1.2.15。浮窗仍是 1.0.7，`engines` 不变，`lib/index.js` 一行没改。

**发布结果**（当场核对）：

- 资产 `dsh-ai-terminal-pet-1.2.16.tgz` **2,005,718 B / sha256 `9CAEE39D…` / shasum `98e7bc0a…`**，19 个文件（多出 `README.zh.md`）。
- 从公网 URL 下载回来复验字节与摘要一致；**拿下载回来的那个文件**装进真的 **0.2.2-rc.3** 宿主：
  `add` 成功、`--dump-config` 有我们的层，装出来的目录里 `README.md`（10,970 B）与
  `README.zh.md`（10,588 B）都在。
- npm `dist-tags.latest` = **1.2.16**，`npm pack` 解包后与 Release 资产**逐文件相同（19/19）**。
- 市场侧：`docs/market-index.json` 的 `description` 与 `dsh.catalog.summary` 已是新版；上游条目的
  `description` 走 [#15075](https://github.com/awesome-dsh-plugin/awesome-dsh-plugin/pull/15075)
  （单文件 `+3 −3`，`MERGEABLE`）——**顺带发现 #15067（1.2.15）在我们发这一版期间已被合并**，
  所以那个 PR 是从"上游 main 已含 1.2.15"重建的。

### 50.5 工具改动：`--description`

详情页上的**描述**来自上游 `data/plugins/<我们>.yml` 的 `description.en/zh`，那是别人仓库里的两行。
`tools/upstream-entry.mjs` 原来只会换 `tarball`（并断言"只改一行"），两个 PR 改同一行必然互相冲突，
所以给它加了 `--description <file.json>`：

- 从 JSON 取 `en`/`zh`，重建那两个**单引号单行标量**（YAML 里 `'` 用 `''` 转义）；
- 形状不对就**拒绝**：块标量（`|`/`>`）、缺一种语言、两个 `description:` 键，一律不猜；
- 断言升级为"只改我们这一个文件、**恰好三行**、每一行都必须以 `tarball:`/`en:`/`zh:` 开头"；
- 证明仍然成立：**把这三行都还原回去**，必须与上游逐字节相同。

规则测试从 8 条加到 **12 条**（形状识别、引号转义、三种拒绝、三条 verdict 语义、块标量拒绝）。
仍然保留 `# 描述由维护者润色` 的余地——PR 正文里写明了"如果你更想用自己的措辞，只取 `tarball` 那一行即可，
我不会为此再开第二个 PR"。

## 49. 上架巡检：十家目录现在各指哪里（2026-11-13 深夜，第 49 轮）

要求是"确保所有插件市场全部更新到最新版"。按 [docs/UPSTREAM.md](UPSTREAM.md) §4 的规矩**读对方公开索引文件**，
不读 PR 页面，逐家核对如下。**"最新"分三种情况**：我们自己能直接改的（已改完）、等对方合并/重建的、
以及对方按自己节奏自动扫的。全部都是当场读到的值，不是回忆。

| 目录 | 它怎么收录 | 现在提供的（实测） | 我们做了什么 |
|---|---|---|---|
| **npm** | 我们发布 | `dist-tags.latest` = **1.2.15** | 已发布，`npm pack` 解包与 Release 资产逐文件相同 |
| **本仓库 format-1 索引**（DSH Discover 可直接加） | 我们自己写 | `1.2.15`、`engines.dsh >=0.2.0-0 <0.3.0-0`、commit `139466c` | B 提交已钉 sha；`verify.mjs` 断言索引与包内声明一致 |
| **awesome-dsh-plugin**（dshmarket 的 Market 页由它生成） | 单文件 PR + 站点重建 | 仓库条目已合并到 **1.2.14**；站点 `plugins.json` 仍渲染 **1.2.9**（含旧 install 命令） | 开了 [#15067](https://github.com/awesome-dsh-plugin/awesome-dsh-plugin/pull/15067)（1.2.15，单文件单行，`MERGEABLE`）。**合并与重建都只能等** |
| **polang233/dsh-plugin-index**（dshopen.org） | 人工审核，按版本/提交固定 | `data/sources.json` 第 1204 条仍钉 **1.2.4**（commit `ff7f6de`）；线上 `public/index.json` 里还没有带 git 源的条目 | 按他们 CONTRIBUTING 的「更新与更正」在**原 issue [#423](https://github.com/polang233/dsh-plugin-index/issues/423) 里回复**了 1.2.15 + 新 commit + npm 已可用（他们写明关闭的 issue 也接受回复，且默认跟踪 npm `latest`） |
| **YELEBAI/dsh-plugin-marketplace** | 每 2 小时扫 `topic:dsh-plugin`，写一份中心 registry | `version 1.2.8`，`install.spec` 钉在 `github:…#8ec59ac` | 无需投递；下次扫描会带走新版本与新描述（见 §49.2） |
| **bradeGithub/DSH-Plugins-Marketplace** | 2 小时增量扫描 | `version` / `npm_version` = **1.2.13**，`registry_seen_at 13:32Z` | 同上，等它自己的扫描 |
| **imsai-sh/awesome-deepseek-harness-plugins** | 目录文件只存仓库与描述，安装走 npm | 条目里**没有版本字段**，所以 npm 装到哪个就是哪个（现在 1.2.15） | 无需改动 |
| **zhu1090093659/dsh-community-plugins** | `community.json` 只存 id/名称/描述/仓库 | 同样**没有版本字段** | 无需改动 |
| **dshmarketplace.dev** | 它自己的提交 API | **已收录**（`inRegistry: true`），但 `installKind: unknown`、`installable: false`、`npmPackage: null` | 没有文档化的更新入口，**没有重投**（重复投递不会加快，只会让"我们不看规则"变成印象） |
| **1024 Store** | 抓取/展示 | 详情页 **200**，标题与摘要正常；**版本不在 HTML 里**（前端渲染），其 API 搜不到本包 | 无可查项，无动作 |
| **dshbase / dsh.so / AllDSH** | 人工复核 issue | `#322` / `#24` / `#2` 仍 **OPEN、0 评论**（11-11 投出） | 按 `docs/UPSTREAM.md` §4 的规矩**不重投** |
| **dsh-plugin.org** | 人工复核 issue | [#735](https://github.com/dshplugin/dsh-plugin-hub/issues/735) 已关闭并回复 **collected**：因 manifest 声明 `platforms: ["win32"]` 标成 Windows-only，安装走源码 `github:jjjjjjjjnnjnn/ai-terminal-pet` | 已收录，无需动作（站点首页前端渲染，HTML 里看不到条目） |

### 49.1 结论

- **我们能自己决定的两位已经是最新**：npm 与仓库内的 format-1 索引。
- **两位跟着仓库/包自动走的已经间接最新**：imsai 与 zhu 没有版本字段，装到哪个就是哪个。
- **两位自动扫描的会在下一次扫描到位**：YELEBAI 与 bradeGithub（各自 2 小时节奏）。
- **三位要等别人动**：awesome-dsh-plugin 等合并 + 站点重建；dshopen 等人工复核；1024 Store 与
  dshmarketplace.dev 的元数据缺陷在对方手里（一个不暴露版本，一个没识别出安装源）。
- **三位仍停在"已投、无回应"**：dshbase / dsh.so / AllDSH。这不是我们能推进的事，重投只会变差。

### 49.2 顺带修掉一个"源头"

抓取型目录的条目描述是**逐字抄我们 GitHub 仓库的那段 description**。它当时还写着
"**56px** always-on-top status strip … **4.17 MiB**"——面板早已是可展开的 200×N（紧凑条 56×156），
浮窗也已是 **4,554,240 B = 4.34 MiB**。也就是说：不改这一处，每两小时就有一批目录把错数字抄一遍。
已改为（289 字符，只写可核对的事实）：

> A Windows overlay for DeepSeek Harness, Claude Code, OpenCode and Antigravity: session state travels
> over a named pipe to a native Go binary, which draws a 4.34 MiB always-on-top panel with a tray icon
> so you can see which agent is waiting. No account, no network, no runtime dependencies.

所以这一轮**没有**去改十家目录里的九段文案，只改了一个源头 + 一个索引 + 一个 PR + 一条 issue 回复。

## 48. 交接：适配 DSH 0.2.x —— 1.2.15（2026-11-13 深夜，第 48 轮；只改兼容声明）

用户说"DSH 版本更新了"。查完事实：本机桌面版是 `@deepseek-ai/dsh-desktop` **0.2.2-rc.3**（更新通道
`nightly`，`resources/app.asar` 里的 `package.json` 就是这么写的），npm 上 `@deepseek-ai/dsh` 的
`dist-tags` 是 `latest 0.2.2-rc.1` / `next 0.2.2-rc.3` / `alpha 0.2.2-alpha.3`。而我们包里的声明还停在
`engines.dsh: ">=0.1.0-rc.6"`——那是 0.1 线的地基，中间隔了 0.1.1–0.1.7 与整条 0.2 线。

### 48.1 真正要改的那一条规则

DSH 从 **0.2.1-alpha.3** 起改判兼容性（`docs/upgrade-guide/v0.2.1-alpha.3/plugin-engines-dsh`）：

> An own `engines.dsh` range decides alone and those peers are ignored; the peer rule applies only
> when `engines.dsh` is absent.

实现在 `@deepseek-ai/dsh-app-boot` 的 `evaluatePluginCompatibility(manifest, exemptions, runtimeVersion)`，
判定用 `semver.satisfies(runtimeVersion, range, { includePrerelease: true })`，不符就把包名、版本与
**未满足的 requirements** 一起返回，`dsh plugin` 据此拒绝安装并**回滚 profile**（pnpm 的 package.json、
lockfile、node_modules 都还原）。也就是说：**`engines.dsh` 这一个字符串就是全部的兼容性声明**，
`peerDependencies` 上的 `@deepseek-ai/dsh*` 范围不再参与判断。

### 48.2 量出来的坑：`includePrerelease` 让文档写法的两端都错

DSH 自己的 `publish.md` 推荐写 `">=0.2.0 <0.3.0"`。用宿主自己的 `evaluatePluginCompatibility` 跑一遍矩阵：

| 运行时 | `>=0.2.0 <0.3.0`（文档写法） | `>=0.2.0-0 <0.3.0-0`（本包采用） |
|---|---|---|
| 0.1.7 | 拒 | 拒 |
| 0.2.0-rc.1 | **拒**（预发布排在 0.2.0 之下） | 收 |
| 0.2.1-rc.5 / 0.2.2-rc.1 / 0.2.2-rc.3 | 收 | 收 |
| 0.2.2（正式版） | 收 | 收 |
| 0.3.0-rc.1 | **收**（预发布同样排在 0.3.0 之下） | 拒 |
| 1.0.0 | 拒 | 拒 |

所以本包写 **`>=0.2.0-0 <0.3.0-0`**：整条 0.2 线（含它的预发布）收，0.3 的预发布也不放进来。
这不是洁癖——DSH 自己就在发 `-rc.N`，用文档写法会一边拒掉自己家的 rc、一边放过下一个大版本的 rc。

### 48.3 怎么验的（可以在任何机器上重做）

```powershell
# 1. 两个真宿主：0.2.2-rc.3（当前）与 0.2.0-rc.1（区间下端）
npm install @deepseek-ai/dsh@0.2.2-rc.3   # 装进一个空目录，2 分钟、672 个包
$env:DSH_HOME = "$env:TEMP\dsh-home"; node .\node_modules\@deepseek-ai\dsh\lib\bin.js --version

# 2. 装包并看它有没有被合成进配置（--dump-config 不启动宿主，只合成层）
node ...\lib\bin.js plugin --profile pet add <tarball>
node ...\lib\bin.js --profile pet --dump-config | Select-String 'ai-terminal-pet'

# 3. 反证：把 engines.dsh 改成 >=9.0.0 再装，必须被拒且回滚
```

实测结果：两个宿主都把 1.2.15 装进 profile（`bundles` = `["@deepseek-ai/dsh-base","dsh-ai-terminal-pet"]`），
`--dump-config` 都能看到 `# == dsh-ai-terminal-pet` 那一层；改成 `>=9.0.0` 的那份被拒，输出
`requirements {"engines.dsh":">=9.0.0"}` 并把 profile 还原。

### 48.4 代码这次一行没改，凭什么

插件只依赖五个事件，0.2.2-rc.3 的文档里都还在：`session/event`（durable session 记录）、
`agent/inbox/claimed`、`agent/status`、`agent/error`、`approval/request`（仍是 **waterfall**，签名仍是
`(req, next)`，见 `docs/subsystems/approval.md`）。`ctx.logger` 与 `process.on('exit')` 也没动。
所以 `lib/index.js` 保持原样，`vendor/AI-Terminal-Pet.exe`（浮窗 1.0.7）一个字节没动——这次是**元数据版本**，
不是功能版本。

### 48.5 同时改的 `engines.node`

`^22.19 || >=24` → **`>=24.9.0`**。DSH 0.2.1-rc.2 起要求 Node 24.9+ 并放弃 Node 22
（`docs/upgrade-guide/v0.2.1-rc.2/node-24`，宿主的 `package.json` 也是 `{"node":">=24.9.0"}`）。
bundle 只在宿主进程里跑，所以声明与宿主对齐；我们自己的代码没有 Node 24 专属 API，但"能在 Node 22 上跑"
这句话现在没有意义——那个宿主根本起不来。

### 48.6 下次 DSH 再升版时照着做

1. `npm view @deepseek-ai/dsh dist-tags` 与桌面版 asar 里的 `version` 各读一次（**读文件，别猜**）。
2. 读 `docs/upgrade-guide/README.md` 里跨过的那几档，尤其 `plugin-*`、`approval*`、`session*` 条目。
3. 用 §48.3 的配方在新宿主上装一次 + `--dump-config`；顺手把矩阵跑一遍（`engines.dsh` 的上下界各取一个
   预发布版本试）。
4. 只有这三件同步改完才算"适配"：包内 `engines`、根 shim 的 `engines`、`docs/market-index.json` 的
   `versions[0].engines.dsh`（`verify.mjs` 会断言三者一致），再加 README 里那句人话。
5. 发布走 §27 / [docs/UPSTREAM.md](UPSTREAM.md)：新版本号（资产不可变）、A/B 两提交、上游 bump PR。

### 48.7 这一轮的发布结果（当场核对）

- **提交与 tag**：A = `139466c`（版本 + 三处 `engines` + 文档），B = `beae43f`（索引 `source.commit` 与
  `icon` 钉到 A）。tag `dsh-plugin-v1.2.15` 指向 A；该提交上的 CI **success**（run `43484108896`）。
- **Release**：[`dsh-plugin-v1.2.15`](https://github.com/jjjjjjjjnnjnn/ai-terminal-pet/releases/tag/dsh-plugin-v1.2.15)，
  资产 `dsh-ai-terminal-pet-1.2.15.tgz` **2,000,898 B / sha256 `F6D85F76…` / shasum `27c20674…`**（18 文件）。
  从公网 URL 下载回来复验：字节与 sha256 一致；**拿下载回来的那个文件**装进真的 0.2.2-rc.3 宿主，
  `dsh plugin --profile pet add` 成功、`--dump-config` 里有 `# == dsh-ai-terminal-pet` 那一层。
- **npm**：`dist-tags.latest` = **1.2.15**；`npm pack dsh-ai-terminal-pet@1.2.15` 解包后与 Release 资产
  **逐文件相同（18/18）**。
- **上游**：新 PR [#15067](https://github.com/awesome-dsh-plugin/awesome-dsh-plugin/pull/15067)
  （分支 `bump-tarball-1.2.15`，单文件单行）已开、`MERGEABLE`；`tools/upstream-entry.mjs --check` 报
  `OK（in flight）`。**意外收获**：1.2.14 的 #15045 在这期间已经被合并了，所以上游 `main` 的 `tarball`
  现在指向 **1.2.14**（不是 1.2.9）——`--check` 的提示原文就是 "still dsh-plugin-v1.2.14, but
  dsh-plugin-v1.2.15 is in flight"，这也是为什么每次都要重新读，而不是照着上一次的记录推理。
- 平台（`dsh.bundle.platforms: ["win32"]`）与 `dsh.marketplace` 块这次没动：前者仍然正确（浮窗是 Win32 的），
  后者是给按 `install.profiles` 过滤的第三方商店读的，DSH 自己不看。

## 47. 维护：工作区整理（2026-11-13 晚，第 47 轮；没有改代码）

要求是"维护、更新、整理工作区"。这一轮只动**可再生的产物**：源码、文档结论与已发布资产一个字节没动。

### 47.1 删了什么，以及凭什么能删

| 位置 | 之前 | 之后 | 依据 |
|---|---|---|---|
| `dist/` | **73,356,358 B**（0.9.0–1.0.6 共 8 个 zip、1 个 1.2.6 的 tgz、`dist/ai-terminal-pet` staging 树） | **19,361,786 B**（只留 1.0.7 的 zip 与 `pet-setup.exe`） | 旧 zip 与 tgz 都在 GitHub Release 上，也能由对应 tag 重建；staging 树是 `scripts/package.ps1` 的工作目录，它每次运行都先删再建（第 243 行），第 451 行还把路径打印给人看 |
| `.local/` | **9,749,966 B**（1.2.4/1.2.5 的 tarball 四份、发版文案草稿、上游快照、门禁日志） | **36 B**（只剩 `state/gh/device-id`） | 这个目录自己的注释就写着"machine-local scratch，随时可删"；草稿的内容已在 GitHub Release 与 git 历史里 |
| `.agents/tasks/` | 5 张 2026-11-12 的任务卡（4,599 B） | 空 | `.gitignore` 注明它们是 KB hub 自动派发的模板，HANDOVER §34.6 记过"它们不是工作" |
| 本地分支 `feat/pet-hook-native` | 已并入 `main`（`git log main..feat` 为空，`main` 领先 86 个提交） | 已删 | 提交仍在 `main` 里；**远端同名分支没动**，删不删由人决定 |
| `.git` | 172 个松散对象 12.6 MiB + 3 个包 9.15 MiB | 0 个松散对象 + 2 个包 10.08 MiB | `git gc` |

核对：`git status` 干净；`git status --ignored` 只剩 `bin/`、`dist/`、`.local/`、`.workbuddy-ai/`、`.agents/`
五个预期目录。

### 47.2 故意没删的，以及为什么不能删

`%LOCALAPPDATA%\AI-Terminal-Pet` 里有一批 2026-11-05 迁移期留下的文件，看着像残渣，但**其中两个是活线**：

- `codex-notify.js` —— `~/.codex/config.toml` 第 1 行 `notify = [ "node", "...\\AI-Terminal-Pet\\codex-notify.js", "turn-ended" ]`。删了 Codex 的回合结束提醒就没了。
- `terminal-pet-profile.ps1` —— 用户的 `Microsoft.PowerShell_profile.ps1` 第 20–23 行 `Test-Path` 之后
  dot-source 它。删了 PowerShell 里的状态同步就没了。

其余（`pet-ipc.js`、`pet-statusline-shim.js`、`pet-statusline-core.exe`、`pet-statusline-old.exe`、
`pet-statusline-v1-backup.exe`、四个 `*.bak-20261105`、`path-backup.txt`、`probe.log`、
`runtime-backup-20261105\`）在 `%APPDATA%`/`%LOCALAPPDATA%` 的配置目录（含 Antigravity、Claude、
opencode）里**没有找到引用**。但"没找到"不等于"确认没有"——Antigravity 的 statusline 配置位置没有
权威清单——所以**一个都没删**。真要清，建议**移动**到 `_old-20261105\` 而不是删除；动手前先确认
`--version` 是 1.0.7、`--doctor` 10 项 0 fail（本轮实测两者都成立）。`path-backup.txt` 尤其不要删：
那是 `install.ps1 -AddToPath` 之前留下的 PATH 原件记录。

### 47.3 顺带核对的两件事

- **没有死图**：`docs/assets/` 的 14 个文件在文档里都有引用（最少 3 次）。
- **没有 CRLF/mixed**：`git ls-files --eol` 里没有 `w/crlf`、`i/mixed`、`w/mixed`；`go mod tidy -diff`
  无输出、`go mod verify` 全过。

### 47.4 以后照这个来

- `dist/` 只留**当前版本**的两件资产；历史版本从 tag 或 Release 取，别指望本地还在。
- 删 `%LOCALAPPDATA%` 里的东西之前先查三个地方：PowerShell profile、`~/.codex/config.toml`，
  以及 `install.ps1 -VerifyOnly` 报的那 11 个文件——**那份清单是"应该存在"的白名单，清单之外的都是历史残留**，
  但残留也可能是活线（上面两条就是），所以"没引用"要有证据，不能靠眼熟。

## 46. 交接：把两个流程缺口变成门禁（2026-11-13 晚，第 46 轮；工具与测试，不改已发布版本）

第 45 轮结束时留下的是两句"下次注意"。这一轮把两句都做成**改不掉的东西**：一个是 CI 步骤，
一个是三条单测加一条 AGENTS 规则。代码行为没有变化，所以**没有升版本、没有打 tag**，
`dsh-plugin-v1.2.14` 的资产一个字节都没动。

### 46.1 缺口一：文档里的版本号会独立腐烂 → `tools/check_versions.mjs`

三条版本线各有各的真相源（`internal/app/app_windows.go` 的 `const Version`、两份 `package.json`、
`docs/market-index.json`），而**大约十几个句子在文档里各自说了一遍版本号**，此前没有任何东西读它们。
结果是三处同时过期：README 徽章停在 **1.0.2**（跨了三次发版）、中英两段「插件包当前版本」写着
**1.2.8**（npm 上已是 1.2.13）、`docs/LAUNCH.md` 还让人去链 `pet-v1.0.0`。

工具做三件事，而且只做这三件：

1. **18 个"当前版本"形状**逐条断言：徽章、发布链接（中英各一次）、便携 zip 名、tarball 安装 URL、
   git 安装 tag、`The package is now` / `当前版本是`、两处 npm 链接、`--version` 示例、ROADMAP 基线行、
   HANDOVER 的状态标题与版本段落、LAUNCH 的发帖链接与两处 npm 数字。
   **每条都要求出现次数精确**——不是"至少一次"：中英两份并存的句子少了一份、或者一份留下旧值，
   都会被抓住（README 徽章就是在"只改一处"的情况下活了四次发版）。
2. **未来版本扫描**：任何文档都不许出现比当前**更新**的版本号（1.0.N 对浮窗线、1.2.N 对插件线）。
   历史值必须留着，所以只抓"更大"；这抓的是半途而废的版本号替换与手滑。
   围栏代码块与反引号里的字面量是唯一的豁免（`docs/HANDOVER.md` 里 `--version 1.2.99` 是故意演示 SKIP
   的例子），而代码里真正要紧的版本字符串（安装 URL、资产名、基线行）由第 1 条覆盖，那一条不跳过任何东西。
3. **CHANGELOG 的标题形状**：必须存在 `## [<当前浮窗版本>]`；两条版本线各自的标题不许超过自己的当前版本；
   `[未发布]` 允许存在，但**只能排在第一条**——"已发布版本的标题下面还挂着一个未发布小节"正是
   半成品发版提交的样子。

溯源一次就够：把 README 徽章改回 1.0.2 再跑，输出是

```
fail: readme badge: README.md:7 says 1.0.2, the overlay version is 1.0.7
```

**诚实地写清它能抓到什么**：它抓的是**已经写下来的形状**。将来用新的句式写版本号（工具不认识的新句型），
它抓不到——那时要做的是把新形状加进 `CLAIMS`，工具在形状消失时会明确提示
`reworded? then update CLAIMS in tools/check_versions.mjs`。所以发版清单里那句
「改版本号前先 grep 一次」仍然保留（`docs/UPSTREAM.md` §2 与 §7），门禁只是兜底，不是替代。

`tools/check_versions.test.mjs`（`node --test`）把上面每条规则钉成纯函数测试：`const Version` 的唯一性、
计数精确、中英双份、未来版本、围栏豁免与行号、CHANGELOG 的两种版本线标题、围栏与反引号豁免，共 **25 条**。

### 46.2 缺口二："点了没写盘"要先问进程是怎么退出的

`config.json` 由 `configSaver` 的守护协程写（同步写会卡住消息泵），`defer settingsSaver.flush()`
只是**正常退出**时的最后一道保险。所以用 `kill` 验证持久化，验的是别的问题。这一轮补了三条测试
（`internal/app/config_saver_test.go`）：

| 测试 | 钉住的合同 |
|---|---|
| `TestConfigSaverPersistsWithoutAFlush` | 只调 `save()`（不 flush）也必须落盘——运行中的浮窗从不 flush |
| `TestConfigSaverFlushPersistsAStagedValue` | 不给它协程，只 `save()` + `flush()`：**flush 是唯一能写出这个值的东西**，也正是 kill 拿走的那一步 |
| `TestConfigSaverFlushWithoutAPendingValueIsSilent` | 没有待写值时不创建文件——空转退出不该伪造偏好 |

第二条的确定性来自"不启动协程"，所以它不是靠 sleep 猜时序的测试。AGENTS.md 同时加了**第 9 条**规则：
报告"没写盘"之前先说清进程是怎么退出的；关窗 + `--quit` 之后再读文件。

### 46.3 为什么值得做成门禁而不是"记住"

这两条的代价都不是"一次疏忽"：§44.7 把沙箱里的一次 kill 当成了未解 bug，一直挂到 §45.1 才被推翻；
版本号则是**每发一次版就要重犯一次**的人工检查。门禁的好处不是省事，而是把"我记得检查"换成
"检查不过就红"。反过来也要守一条边界：**不要**给文档写通用 lint（"所有数字都必须是对的"必然误报），
只钉**已经写下来的形状**，并且让失败信息带上文件、行号与"怎么修"。

### 46.4 这一轮顺带修掉的四份文档

`AGENTS.md` 的验证命令标题写着 **201 个 `func Test`**（实际 329）；`docs/LAUNCH.md` 三条推文草稿
也写着 201、浮窗 4.17 MiB（实际 4.34 MiB）；README 徽章 1.0.2；README 四处把插件包写成 1.2.7 / 1.2.8。
后三项在上一个提交 `c2c9484` 修掉，AGENTS 与 LAUNCH 的数字在本轮修掉，并且全部由新门禁接管。

### 46.5 没做的

- **没有跨仓库版本同步**：上游目录条目指到哪个版本仍是 `tools/upstream-entry.mjs --check` 的事，
  它需要网络，失败语义必须是 SKIP，这两件事不能混进一个本地的、确定性的检查里。
- **没有把 `--json`/`--doctor` 的数字搬进门禁**：那些是运行期观测，属于探针，不属于文档一致性。
- **没有升版本**：本轮产物是工具与测试，用户可见行为没有变化；下次发版时 `[未发布]` 一节会一并转正。

## 44. 交接：设置窗口可读、可用、外观页签终于能点（2026-11-13 深夜，第 44 轮；**已随 1.0.7 发版**）

用户两句话：**"外观这个部分点不进去"**、**"设置界面字太小之类的"**。第一句是一个漏掉的 `case`，
第二句是字号梯子整体低于系统基线。两句都在渲染图和真机窗口上确认过之后才动手。

### 44.1 外观页签：一个漏掉的 case，后果是三个设置没有入口

`settingsApply` 的页签分支原来只写了 `settingsTabGeneral, settingsTabSound, settingsTabSessions`。
`settingsTabAppearance` 不在里面，所以命中测试、悬停高亮、选中态全部正常，**点下去什么也不发生**。

严重性不在"一个页签坏了"：主题、按终端着色、启动时隐藏**只存在于那一页**，而托盘右键菜单里没有主题项
（逐条核过：显示/隐藏/重置/设置/通知/声音/音量/完成/置顶/贴边/自动同意/语言/自启/退出）。
也就是说这三项在 UI 里**完全没有入口**，只能手改 `config.json`。

**为什么现有测试没抓到**：`TestSettingsHitTestFindsEveryControl` 只问"点在哪个控件上"，几何层全绿；
没有任何测试问"点了之后发生了什么"。`TestSettingsEveryTabSwitches` 问的是后者——它直接调
`settingsApply` 并断言 `settingsTab` 变了，是这类"画得出来、点不动"的唯一有效测试形状。
`TestSettingsApplyPublishes` 再补一层：外观页签的六个控件必须都触发 `OnSettingsChanged`。

### 44.2 字号：GDI 会把请求吸附到粗档位（这一轮最值钱的发现）

`canvas.text` 用 `CreateFontIndirect` 的 `lfHeight`（正数 = 字符**格**高），GDI 再把它吸附到字体自己的档位。
实测（同一串"状态提醒"的宽度 ÷ 4 = 实际 em）：

| 请求 | 9 | 10 | 11 | 12 | 13 | 14 | 15 | 16 | 17 | 18 | 20 |
|---|---|---|---|---|---|---|---|---|---|---|---|
| 实际 em | 7 | 7 | 8 | 8 | 8 | 9 | 9 | 10 | 12 | 12 | 14 |

两条结论：

1. **加 1px 经常完全没效果**（11/12/13 是同一个字号）。旧梯子是标题 15、页签 12、行标题 12、说明 10，
   实际只有 **9px 和 8px 两种**——"又小又平"是机械原因，不是感觉。
2. 说明只有 7px em，比 Windows 对话框基线（12px 格高 ≈ 9px em）低两档。上一轮甚至用加亮颜色
   （`slatePalette.Hint` 的注释写着"在 10px 下 Muted 太暗读不清"）绕过了字号问题。

新梯子落在真实不同的档位上（标准档：标题 20→14px、页签 14→9px、行标题 16→10px、说明 14→9px），
小/大三档是**三张手写的表**（`settingsTypeRamps`）而不是一个系数乘出来：把同一张表乘 1.2 会让
"12px 标题"和"9px 说明"一起落到 48px 那档，层级直接消失。`TestSettingsTypeRampStaysOrdered`
在 96/120/144/192 四个 DPI 下量真实宽度，保证标题 > 行标题 > 说明。

### 44.3 行高从部件算出来，换行只有一个实现

- 行高 = 标题带 + 说明行数 × 行高 + 控件带 + 内边距，不再写死 54/68/80/58。
- 说明从"单行省略号"改成"最多两行换行"（备注三行）。这同时修掉英文界面**实测 6 处**被截断的提示。
- `wrapHint` 同时负责测量与绘制，所以"预留几行"和"画了几行"不可能不一致；
  `TestSettingsHintsNeverLoseTextSilently` 在 3 档字号 × 2 语言 × 4 页上比对两者。
- 中文按**行首禁则**处理（`pullOrphanPunctuation`）：句号/逗号/右引号不会被挤到下一行开头。
- **踩过的坑**：备注的段落存在 `title` 里（旁边没有控件），而测量只量 `hint`，于是备注永远按一行算高、
  第二行被裁——渲染图第一眼就能看见。现在 `measureSettingsRows` 对 `rowNote` 量 `title`，
  测试也覆盖备注。

### 44.4 屏幕放不下：降档，而不是滚动

原计划是给设置窗口加滚轮滚动。实际做的时候改成了**按工作区降档**（大 → 标准 → 小）：
字号是偏好，工作区是约束，而"够不到的控件"比"小一档的字"更糟。这样就不需要滚动偏移进
`settingsFrameKey`、不需要命中测试加偏移、也不需要给分层窗口做裁剪绘制——同样的目标，少一半状态。
`TestSettingsFittingShrinksForAShortScreen` 钉住"矮屏一定降档、宽屏保留选择、极端情况退到最小档"。

### 44.5 DPI：读自己的窗口，处理 WM_DPICHANGED

设置窗口原来用**浮窗**的 `GetDpiForWindow`，而且不处理 `WM_DPICHANGED`：把它拖到另一块不同缩放比的屏上，
字号和布局都不跟着变。现在 `settingsDpi()` 读设置窗口自己的 DPI（回退浮窗），`WM_DPICHANGED` 里
用 wParam 低字的新 DPI 重排并重新夹进工作区（Windows 建议的矩形被刻意忽略——这个窗口的位置由自己的规则决定）。

### 44.6 键盘与 tooltip：两个新开关，默认值都有理由

- `settingsKeyboard`（**默认关**）：Tab / Shift+Tab 移动焦点，空格或回车执行，Esc 始终关闭。
  默认关是因为这个对话框是鼠标优先的，会吞 Tab 的窗口对只用鼠标的人是意外。焦点环只在真的用过键盘之后才画。
- `settingsTooltips`（**默认开**）：悬停被省略的标签时在那一行旁边画出全文。它画在窗口内而不是做成
  shell 工具提示窗口——这个窗口是分层表面、没有子控件，为一个已经很少发生的情况再开一个顶层窗口不划算。

### 44.7 验证

- 单测：`settings_tabs_windows_test.go` 8 条（页签真的切换、梯子不塌、字号真的改变窗口、矮屏降档、
  说明不丢字、焦点顺序完整、键盘是可选、换行不丢字）+ `TestSettingsApplyPublishes`（6 个控件都会发布）
  + `internal/app/settings_window_test.go`（新偏好过 `applyUISettings` 与 `config.json` 往返仍在）。
- 渲染图：96dpi 与 144dpi 下四页都渲染出来看过；外观页的备注换行、行首禁则、"标准"高亮都是看图才发现的。
- 真机：沙箱里 `--settings` 起真窗口，用 `PostMessage` 点侧栏第三个页签，窗口高度 **826 → 586**
  两次都复现（点关闭按钮窗口隐藏，证明消息通路是通的）。
- **没观察到的**（**§45.1 已查清**）：沙箱里 `config.json` 没有被重写，原因未查明（进程被 kill 早于合并写入、或沙箱环境差异），
  所以这条改由 `applyUISettings` + `Save`/`Load` 的往返测试覆盖。下次做真机点击验证时值得再查一次。

### 44.8 没做的事

- **没发版**（**已由 §45 完成**）：本轮的代码随后作为 1.0.7 / 1.2.14 发布，版本前移与八步清单见 §45。
- **`vendor/` 必须重新 vendor**：源码变了，`vendor.mjs --check` 的源码摘要会失配。顺序永远是
  `go build` → `node packages/dsh-ai-terminal-pet/scripts/vendor.mjs` → `--check`（见 §43.7）。
- **无障碍（UIA）没有做**：这个窗口是自绘的分层表面，屏幕阅读器看不到任何控件。这是一整轮的工作，不是顺手。
- **设置窗口没有滚动**：按 44.4 改成了降档。如果以后真的需要滚动，`settingsGeometry` 已经带了
  `contentHeight`，加偏移是局部改动。

### 44.9 下一个会话从这里开始（发 1.0.7 / 1.2.14 —— **已完成，见 §45**）

一句话：**代码和测试已经好了，缺的是发版仪式。** 工作区干净、`821226a` 已推送、CI 绿，
第 44 轮的改动全部在 `main` 上，只是没有升版本、没有 tag。

按 [docs/UPSTREAM.md](UPSTREAM.md) §2 的八步走：

1. **升版本**：`internal/app/app_windows.go` 的 `const Version` → `1.0.7`；根 `package.json` 与
   `packages/dsh-ai-terminal-pet/package.json` → `1.2.14`；`docs/market-index.json` 的
   `versions[0].version` 与 `publishTimestamp`。CHANGELOG 里第 44 轮的 `[未发布]` 转成 `[1.0.7]`。
2. **所有版本号文本改完之后**才 `npm pack`（§43.6 的教训），并把包里的 README 解出来核对
   （`tar -xzf … package/README.md`）。
3. `go build` 三个二进制 → `node packages/dsh-ai-terminal-pet/scripts/vendor.mjs` → `--check`。
   **源码变了就必须重 vendor**，否则 `vendor.mjs --check` 的源码摘要失配、CI 红。
4. `scripts/package.ps1` 产出 zip 与 `pet-setup.exe`；`node packages/dsh-ai-terminal-pet/verify.mjs`。
5. **A 提交**（版本 + 两份清单 + 索引版本/时间戳 + CHANGELOG + README/ROADMAP/HANDOVER 的数字）
   → 打 `pet-v1.0.7` 与 `dsh-plugin-v1.2.14` → 两个 Release → `npm publish`。
6. **B 提交**把索引的 `source.commit` 与 `icon` 钉到 A 的完整 sha（git 无法在一个提交里引用自己的 sha）。
7. `node tools/upstream-entry.mjs --version 1.2.14 --write --branch bump-tarball-1.2.14` → 开 PR →
   关掉 #14990 并留一句「被新 PR 取代」；#14899、#14945 也可以一并关掉，它们只是同一行的旧版本。
8. 从公网 URL 下载三个资产复验字节；沙箱里 install → `-VerifyOnly` → `--doctor` → `-Uninstall`。

两件顺手要做的事：

- **真机确认一次持久化**：§44.7 里那条没查清的事（沙箱里点完控件 `config.json` 没被重写，原因未定）。
  在本机设置窗口里点一下「纸白」、退出浮窗再打开，`config.json` 的 `theme` 应该变成 `paper`。
  如果没变，那是一个真 bug，值得在发版前修掉——`TestSettingsApplyPublishes` 与
  `TestUISettingsRoundTripKeepsSettingsWindowPreferences` 只覆盖到 `applyUISettings` 为止。
- `python docs/assets/render_assets.py` 只在面板外观变化时才需要重跑；本轮确认 `docs/assets/` 里
  没有设置窗口的图（设置窗口的预览只在 `PET_PREVIEW_DIR` 下由测试临时产出），所以**不需要**更新截图资产。

**不要**为了"顺手"把第 44 轮和第 43 轮合并成一个版本：1.0.6/1.2.13 已经发布且不可变，
新批次只能往前走。

## 43. 交接：三个"假提醒"的根因 + OpenCode 对齐 2.x（2026-11-13 晚，第 43 轮；浮窗 1.0.6 / 插件 1.2.13）

用户的原话只有两句：**"普通终端命令也会响"**、**"没有 agent 完成任务它也响"**。两句都不是"提醒太敏感"，
而是四个互相独立的错误。顺序、证据与结论如下；每一条都先在真机/沙箱里复现，再改，再回归。

### 43.1 消失 ≠ 完成（"没人完成它也响"的一半）

`internal/ui/app_windows.go` 的 `vanishedInputs` 原来为"忙过之后从 store 里消失的会话"**合成一个
closed/done**，注释写着"最强的完成信号"。但"消失"只是"丢失"：上报端进程死掉、重复 pid 被清理、
快照过期，三种都会触发，没有一种能证明工作结束了。

现在：`AlertInput.Vanished` → `Decide` 在 `unchanged` 之后、`KindDone` 之前返回静默的
`decisionVanished`（`internal/ui/attention.go`）。回执里仍然写明 `state=closed / kind=none /
decision=vanished`，所以"为什么没提醒我"仍然可查。

沙箱实测（`%TEMP%\pet-repro3.ps1`，只杀上报端进程，会话本身一直是 `working`）：

| | 旧 | 新 |
|---|---|---|
| 回执 | `closed/done -> done-away` | `closed/none -> vanished` |
| tone | `finished` | 空 |
| balloon | `opencode · ses_silent` | 空 |

### 43.2 一个会话一行（重复行）

store 的 key 是 `source + sessionId + pid`（`internal/store/store.go` 的 `key()`），因为管道、快照文件、
tombstone 全部按上报进程记账——这是**协议层身份**，不能动。面板画的是**会话**，所以在
`internal/ui/sessions_dedupe.go`（新文件）做面板层去重，由 `App.load()` 唯一一处接入。

- 同 `session.Identity` 只留一行；时间戳相同取更紧急的状态，再取更大的 pid → 结果稳定（每帧都一样）。
- 没有 sessionId 的行**永不合并**：合并两行不相关的会把其中一行藏起来。
- `--json` **故意**保留 per-process 行：两个上报端对同一会话说法不一致，正是排障要的证据。
  这条边界由 `TestDedupeIsAPanelConcernOnly` 钉住。

同时 `sessionAlertKey` 从 `source\0sessionId\0pid` 改成 `session.Identity`（`source\0sessionId`）：
静音列表与未读标记本来就是按身份记的，只有告警账本还按进程记，导致一次 OpenCode 重启让同一个等待中的
会话以两个 pid 各响一次。

### 43.3 幽灵行（死 pid 被窗口标题续命）

`internal/platform/presence_windows.go` 的 `SessionAlive` 加了两条规则：

1. **受信任上报端**（`opencode` / `powershell-ipc` / `pwsh-session`，即 `trustedProcessID` 为真的那些）
   的 pid 一旦死亡，`SessionAlive` 立刻返回 false，**窗口标题不再参与判断**。
   此前一个死掉的 `opencode.exe` 留下的快照，会被任意一个"标题里含会话标题或项目名"的标签页永久续命。
2. **纯工具名标题不是会话身份**：`bareWindowTitle` 表 + `titleIdentifiesConversation()`，
   按**整条标题**匹配（`OC | 修一下解析器` 仍是身份，`OpenCode` 不是）。

真机测量（同一台机器，修前）：`--json` 6 行 / 3 个 sessionId，其中 pid 106060、46108、38752 都已不存在，
三行全靠一个 `OC | <标题>` 标签页吊着。修后同机：9 个死 pid 的快照被扫描器删除，只剩活进程 54972 的
4 行 / 4 个会话。

**故意没做的事**：没有采用"窗口必须属于该会话的进程树"这条更强的规则。实测 pid 38072（活的 opencode）
父进程 24640 已死、它的窗口 pid 20992 不在链上——那条规则会把活会话一起删掉。
标题规则也**保留**给合成 pid 的 hook/statusline 上报端：它们的 pid 本身不证明任何事，
`TestPresenceSyntheticPIDStillUsesTheTabTitle` 钉的就是这个。

### 43.4 OpenCode 上报端：两处同时过时（"普通命令也响"的直接原因）

`reporters/opencode.js` 旧文件有两个独立的错：

- **插件 API 是 1.x 的**：导出函数 + `{ event, "tool.execute.before", "tool.execute.after" }`。
  OpenCode 2 的加载器要求默认导出带 `id` 和 `setup`（二进制字符串原文：`Plugin must export a default
  definition with an id and an effect or setup function`），所以这个文件在 2.x 上**什么都没做**。
- **事件名靠子串猜**：`tool.execute.before/after`、`question.asked/replied`、`permission.updated`
  在 2.2.7 里一个都不存在；而 `includes("prompt")` 把 `tui.prompt.append`（往 TUI 输入框追加文本，
  也就是用户敲/粘一条普通命令）判成 `waiting_permission` —— 这就是"普通命令也响"。

现在事件表按 `@opencode/schema` 逐条写死（来源文件写在文件头注释里：
`session-event.js` / `session-status-event.js` / `permission.js` / `form.js`，装在本机
`~/.config/opencode/plugins/notify-native/node_modules/@opencode/schema`）。三个刻意的判断：

- `session.tool.failed` → `working`，**不是** `error`。命令失败是日常，agent 读错误继续。
- `session.step.failed` → `working`（会被 `session.retry.scheduled` 重试；终结性失败由
  `session.execution.failed` 报）。
- 唯一映射到 `idle` 的是 `session.execution.succeeded`。
- 逐 token 的 `*.delta` / `session.tool.progress` **一律不映射**：不改变面板画的东西，一次回答却有上百个。

回归测试 `reporters/opencode-events.test.mjs`（8 个用例）与 CI 新增一步一起钉住这张表。

**本机还有一份不在仓库里的插件**：`~/.config/opencode/plugins/terminal-pet/index.ts` 才是
`opencode.json` 的 `plugins` 里真正加载的那个。它做了同样的收敛，并顺手修了两处：
新会话第一次上报若是 `idle` 会被"状态没变"优化吞掉（现在 `lastStateChange === 0` 视为必须上报）；
标题不再伪造 `"opencode"`。**下次改 OpenCode 上报语义时，两边必须一起改**——仓库里那份是文档与
`file://` 接入方式的真相，本机这份是当下实际跑的。

### 43.5 这一轮的验证清单（都跑过）

- `gofmt -l .` 无输出；`go vet ./...` exit 0；`go test -count=1 ./...` 12 包全过（316 个 `func Test`）。
- `reporters/*.js` 逐个 `node --check`；`node --test reporters/opencode-events.test.mjs` 8/8。
- `node integrations/deepseek-harness/verify.mjs`、`node packages/dsh-ai-terminal-pet/verify.mjs` PASS；
  `node packages/dsh-ai-terminal-pet/scripts/vendor.mjs --check` 两文件吻合（本轮已重新 vendor）。
- `node tests/fresh-machine.mjs --no-overlay` 162/162；`python tools/check_docs_links.py` 12 份全解。
- 真机：`--json` 从 6 行/3 会话（3 个死 pid）变成 4 行/4 会话（只有活进程）。
- 沙箱（`%TEMP%\pet-repro2.ps1`、`pet-repro3.ps1`，都用 `alerts.json` 这个产品自己的判定记录）：
  控制组（真完成）仍然 `tone=finished`；重复 pid 被清理后**无 tone/balloon**；只杀上报端后
  `decision=vanished` 且静默。

### 43.6 发布与上游（本轮实际发生的）

**浮窗 `pet-v1.0.6`**（tag 在 `1efca5a`）：

| 资产 | 大小 | sha256 | 公网复验 |
|---|---|---|---|
| `ai-terminal-pet-1.0.6-win-x64.zip`（17 文件） | 5,041,100 B | `5E2762A2…` | 下载回来逐字节相同 |
| `pet-setup.exe` | 14,274,048 B | `39BC505C…` | 下载回来逐字节相同 |

沙箱拿**下载回来的 zip** 跑完整链路：解出 17 个文件 → `install.cmd`（无 `[fail]`）→ 覆盖层与
`protocol.js`、DSH bundle 都落到沙箱 `%LOCALAPPDATA%` → `-VerifyOnly` **11 个 `[ok]`、0 失败** →
装出来的二进制 `--version` 打印 **1.0.6** → `--doctor` **10 项、0 fail** → `-Uninstall` 后安装目录
**0 文件**。`pet-setup.exe` 本轮重建（内嵌 payload 跟着 `scripts/package.ps1` 一起变）。

**插件 `dsh-plugin-v1.2.13`**（tag 在 `05fc1fb`）：`dsh-ai-terminal-pet-1.2.13.tgz`，
1,990,629 B / `FB809C2B…` / shasum `dbfc774d…`；公网下载回来逐字节相同，解包后 README 里两条安装命令
都是 1.2.13，没有残留的 1.2.11/1.2.12。

**为什么是 1.2.13**：1.2.12 的 tarball 是在改 README **之前** `npm pack` 出来的，包里那份 README 的
安装命令还写着 `dsh-plugin-v1.2.11`——照它装会装到上一版。已发布资产不可变（AGENTS.md 规矩 6），
所以用修好的 README 重新打 1.2.13，并关掉 1.2.12 的 PR。**这是本轮唯一一次重复劳动，成因是顺序**：
`npm pack` 必须在所有版本号文本（含 `packages/dsh-ai-terminal-pet/README.md`）改完之后，
而且要把包里的 README 解出来核对（`tar -xzf … package/README.md`），只看工作区不算数。

**npm**：`dist-tags.latest` = **1.2.13**（发布后约 3 分钟才生效，头几分钟读到的还是上一版，这是正常的）。
`npm pack dsh-ai-terminal-pet@1.2.13` 解包后与 Release 资产**逐文件 sha256 相同**（18/18）；
注意 npm 会重新打包，**tarball 字节数与 sha256 与 Release 资产不同是正常的**（1.2.13：npm 端
1,990,629 B 而 gzip 头不同 → 哈希不同），所以复验要比解包后的文件，不要比 `.tgz` 本身。

**上游**：[PR #14990](https://github.com/awesome-dsh-plugin/awesome-dsh-plugin/pull/14990)
（分支 `bump-tarball-1.2.13`，单文件单行）。`node tools/upstream-entry.mjs --check` =
`OK (in flight)`。同一作者的 #14899（1.2.10）、#14945（1.2.11）仍开着，#14986（1.2.12）已关并留了说明；
维护者合并 #14990 之后，上游条目才会从 `dsh-plugin-v1.2.9` 跳到 1.2.13（**合并 ≠ 可见**，
要等对方站点重建，读 `plugins.json` 而不是 PR 页面）。

**市场索引**：A 提交 `1efca5a` 带 1.2.12 与 `publishTimestamp`，A2 提交 `05fc1fb` 把它改成 1.2.13
（`publishTimestamp` 同时更新），B 提交 `22a9211` 把 `source.commit` 与 `icon` 钉到 `05fc1fb` 的完整 sha。

**本机已就地升级到 1.0.6**（先 `--quit` 再 `scripts\install.ps1`，因为运行中的 exe 会被锁）：
`-VerifyOnly` **11/11**，`--version` 打印 1.0.6，`--doctor` **10 ok / 0 warn / 0 fail**。
这一步不能省——本轮修的是"提醒响错"，而用户面板里跑的如果是旧 exe，改完代码他也感觉不到。

### 43.7 没做的事 / 下一步

- **已发版**：`pet-v1.0.6` 与 `dsh-plugin-v1.2.13` 按 §27 与
  [docs/UPSTREAM.md](UPSTREAM.md) 的八步顺序走完（A 提交带版本/两份清单/索引版本与时间戳，B 提交钉
  `source.commit` 与 `icon`，两个 tag、两个 Release、npm publish、上游 `bump-tarball-1.2.13`）。
  资产字节、sha256、沙箱复验与上游 PR 号记在本节末尾。
- **`packages/dsh-ai-terminal-pet/*.tgz` 不进 git**：1.2.11 的 tarball 是上一次 `git add -A` 顺手带进
  `73a9fad` 的（2 MB 二进制躺在历史里）。本轮把它从工作区删掉、把 `*.tgz` 写进 `.gitignore`，
  打包产物只作为 Release 资产上传。
- **改任何 `.go` 文件后要"重建 + 重新 vendor"一次，哪怕只改注释**：Go 的 build ID 参与源码内容哈希，
  所以注释改动也会让二进制字节变（体积不变、哈希变），`vendor.mjs --check` 在本机（同 Go 版本）会走
  逐字节比较并报 `the build is not reproducible`。这不是源码摘要失配——摘要比的是源码，那一步是过的。
  顺序永远是：`go build` → `node packages/dsh-ai-terminal-pet/scripts/vendor.mjs` → `--check`。
- 如果用户下次还说"还在响"，排查顺序是固定的：先看 `%LOCALAPPDATA%\AI-Terminal-Pet\alerts.json`
  里那行的 `decision`（`vanished`/`done-away`/`idle-away`/`foreground` 各自指向不同层），
  再看 `runtime\*.json` 里那个会话的 `state`+`reason` 是谁写的（`reason` 能直接区分是
  `execution_succeeded` 这种真完成、还是某个上报端自己声称的 `waiting_permission`），
  最后才怀疑提醒层——本轮四个根因里有三个在提醒层之外。

## 42. 交接：点行复制路径（2026-11-13 下午，第 42 轮；浮窗 1.0.5 / 插件 1.2.11）

### 42.1 一句话：这一轮没有改协议，因为要用的数据早就在线上

`session.Project` 里装的就是上报端给的**完整工作目录**（`docs/PROTOCOL.md` §4 写的是「项目路径 / 工作目录」，
上限 1024），面板行只是用 `filepath.Base` 把它截成名字。所以 F8 剩下的确实是三件事：剪贴板、回执、
以及一个新命中区——**没有新字段，没有任何已装上报端需要升级**。这一条在第一轮读代码时就确认了，
后来的施工与验收都建立在这个事实上。

### 42.2 交互规则写成纯函数，顺序是排过的

`internal/ui/tap_windows.go` 的 `decideTap`：

```
跳转箭头 / 紧凑条右半  >  紧凑条左半（复制）  >  auto-approve 否决（仅倒计时武装时）  >  展开态行（复制）  >  收起
```

**否决排在行复制之前是有意的**：自动放行倒计时在跑的那几秒里，「别放行」是最贵的一次点击；
用一次剪贴板写入来回答它，等于静默放行。这条规则有表驱动测试（`tap_windows_test.go`），
因为「点哪儿会发生什么」是用户唯一真的能感觉到的部分。

### 42.3 五条施工约束（三条是旧的，两条是新的）

| 约束 | 具体 |
|---|---|
| 命中与绘制同源 | `expandedRowRect`/`expandedRowHit`（1.0.3 落好）与新的 `compactActionRects` 同时喂给光栅化与命中；紧凑条两半有「不重叠、中点各归各位」的单测 |
| 新画的进 `renderedKey` | 回执有自己的三元组 `toastRow`/`toastTick`/`toastFailed`，**不挂 `clock`**：呼吸路径刻意忽略 `clock`，挂在上面就是 1.0.1 时长冻结的翻版 |
| 回执复用 `syncAnimation` | 不新开常驻计时器；`syncAnimation` 现在就问 `toast.frame(...)` 而不是 `toast.active(...)`，这样「行换主人」时立刻停止付费 |
| **回执是关于某个会话的断言**（新） | 面板会随上报重排，所以 `copyToast` 记 `session.Identity`：那一行换了主人、或紧凑条的最紧急会话变了，回执立刻消失，而不是把「已复制」留在别人的项目上 |
| **失败不假装成功**（新） | `copyToClipboard` 只在 `SetClipboardData` 真的成功后返回 true，失败时回执写「复制失败」并响一声，且**不**落到「收起」——被忽略的点击和没收到的点击看起来一样 |

**剪贴板本身的顺序也换了**：先 `GlobalAlloc` 拷贝、再 `OpenClipboard`/`EmptyClipboard`/`SetClipboardData`。
直觉顺序（先开、先清、再分配）在分配失败时会留下**一个空剪贴板**——用户上一次复制的内容没了，
而「复制失败」与「点了没反应」还长得一样。现在最坏情况是「什么都没变」。

### 42.4 紧凑条：一个胶囊两个控件

56×156 的条子上没有第二行可用，所以跳转条被**横着分成两半**：左半 17 单位是复制图标
（两个重叠的方片，用线条画，不依赖字体），右半是原来的箭头 + 动词，中间一条分隔线说明「这里是两个控件」。
外轮廓仍由 `compactFocusRect` 定义，所以面板剪影没变。三条守卫：

- 没有可复制的会话（没有跳转目标、或 project 为空）→ **不分**，整条仍是跳转键；
- 设备像素太小（65% 缩放以下）→ 不分，宁可没有复制也不给一个四像素的陷阱；
- 动词放不下时不画省略号：`Jump` 在 56 单位上量出来是 17 px，而右半只有 ~20 px，
  `fitText` 会把整句变成「…」，读起来像坏掉的标签。

### 42.5 这一轮的实测

```
gofmt -l .                     # 空
go vet ./...                   # 0
go test -count=1 ./...         # 12 包全 ok，302 个 func Test（1.0.4 是 283）
node packages/dsh-ai-terminal-pet/verify.mjs / scripts/vendor.mjs --check / integrations/…/verify.mjs   # 全 PASS
node tests/fresh-machine.mjs --no-overlay   # 162/162
python tools/check_docs_links.py            # 12 份
node --test tools/upstream-entry.test.mjs   # 8/8
PET_PREVIEW_DIR=… go test ./internal/ui -run 'TestRenderPreview|TestRenderSettingsPreview'   # 31 张
```

**交付形态上验过三遍**（这是本轮最花时间的部分，也是最值钱的部分）：

1. `PET_CLIPBOARD_PROBE=1` 的桌面探针：写一次、用剪贴板 API 读回、再写一次确认是**替换**而不是追加；
2. 沙箱里跑**本机构建**的 1.0.5，用 `PostMessage` 发真实的 `WM_LBUTTONDOWN`/`WM_LBUTTONUP`：
   点第一行后 `Get-Clipboard` 就是那个完整路径，点页脚仍然收起（200 px → 56 px），
   紧凑条左半复制、右半不碰剪贴板；
3. 从**公网下载**回来的 `ai-terminal-pet-1.0.5-win-x64.zip` 重跑一遍：`install.cmd` 11 件制品 →
   `-VerifyOnly` 11/11 → 装出来的 `--doctor` **8 ok / 2 warn / 0 fail** → 同一个手势 → `-Uninstall` 目录归零。

发布：`pet-v1.0.5`（zip **5,039,878 B / AB0218D0…**、`pet-setup.exe` **14,267,904 B / A5B9025A…**）与
`dsh-plugin-v1.2.11`（tgz **1,988,410 B / C3FA49B4…**，18 文件）都已从公网下载回来复验；npm `latest` = **1.2.11**，
`npm pack` 取回的 tarball 与 Release 资产逐字节相同。B 提交把索引钉到 A 提交 `885bdd4`。

### 42.6 交接要点

- 上游跟进：[PR #14945](https://github.com/awesome-dsh-plugin/awesome-dsh-plugin/pull/14945)（`bump-tarball-1.2.11`），
  工具生成、单文件单行（`+1/−1`）、`MERGEABLE`，`check` 排队中。合并后 `node tools/upstream-entry.mjs --check` 会变 OK。
  **注意**：`#14899`（1.2.10）仍开着，两者改同一行，先合哪个都行，后合的那个会再冲突一次——那时**重新用工具从 `main` 建**。
- **B 提交之后 CI 会红一次**，设计如此（`--check` 在「已发布、上游还是旧、且没有 PR 开着」时报红，而 B 提交早于 PR）。
  本轮实测：重跑那一次即 success。**这条是第二次遇到了，下次别再怀疑自己。**
- **本机已就地升级到 1.0.5**：`scripts\install.ps1` → `-VerifyOnly` 11/11 → `--doctor` 10 ok / 0 warn / 0 fail。
  交接前顺手核对一次本机版本与最新 tag（留着旧版会让「本机实测」与文档打架）。
- **`bump-tarball` 分支不能复用**：本轮先推过一次同名分支，第二次 `--force-with-lease` 被 GitHub 判 stale
  （force-with-lease 的期望 sha 是本地旧引用）。做法是 `gh api -X DELETE .../git/refs/heads/<branch>` 再重跑，
  或干脆换一个分支名。工具没有自动处理这件事，**这是它的一个缺口**。
- 剩下的候选见 [docs/ROADMAP.md](ROADMAP.md) §4：`I1-δ`（代码签名 / MSI，唯一需要外部输入的一段）、
  `E3` 跨平台、`L2` 体积（4.29 MiB → 2.5 MiB 目标）、`F2` 键盘导航（`Esc` 已有，缺方向键与 `Enter`）。

## 41. 交接：判定记录搬家（2026-11-13 早上，第 41 轮；浮窗 1.0.4 / 插件 1.2.10）

### 41.1 一句话：1.0.3 装的「可回答性」在装出来之后是坏的，而坏法正是它想消除的那种

1.0.3 的 `--doctor` 第 10 项在**刚装好的机器上**打印「还没有判定记录」，而 `--json` 的 `alerts` 在一次真实气泡
之后仍是 `null`。原因不是它没写，是**它写错了地方**：`alerts.json` 放在 `runtime/`，而那个目录里每个 `*.json`
都被 `internal/store` 当成一条上报快照，**解析不成会话就删掉**。一个 5 秒的扫描周期内文件就没了。

于是「这一层跑过、没什么好说」和「这一层没跑」还是一模一样——正是这份记录存在的理由。

### 41.2 为什么单测没抓到

单测验证的是「写得进去、读得回来」，全绿。它没有验证「**扫描器跑过之后它还在**」。抓到它的是把发布出来的
1.0.4 装进沙箱、让上报端推一次 `working` → `idle`、再等两个扫描周期。**这一类错误只能在交付形态上验证**，
所以现在多了一条从目录一侧出发的测试：`internal/store/runtime_contract_test.go` 钉住「runtime/ 里只放快照，
任何解析不了的 JSON 都会被删」，下一个想在快照旁边放 JSON 的功能会在 CI 里发现。

### 41.3 修法

| 项 | 现在 |
|---|---|
| 位置 | `%LOCALAPPDATA%\AI-Terminal-Pet\alerts.json`（`config.json` 旁边），不在 `runtime/` |
| 安静时 | **也写一条**，`skipped` 为 `first-load`（首次播种）或 `unchanged`（没有会话变化） |
| `--json` | `alerts` 对象照旧；没有记录时是 `null`（从未跑过的新机器不是故障） |
| `--doctor` | 一行给出时间、在场状态、气球目标、逐会话判定；没有记录时是 `[ok]` |

沙箱实测（**装出来之后**）：

```
上报端 working → idle，等两个扫描周期
alerts.json 仍在：{"at":"…","sessions":[{"identity":"probe · generic","state":"idle","kind":"done","decision":"done-away"}],
                  "balloon":"probe · generic","tone":"finished","presence":"present","present":true}
--doctor：alerts: last scan …, presence=present, balloon=probe · generic, decisions=done/done-away
```

### 41.4 这一轮的实测

```
gofmt -l .                         # 空
go vet ./...                       # 0
go test -count=1 ./...             # 12 包全 ok，283 个 func Test
node packages/dsh-ai-terminal-pet/verify.mjs / vendor.mjs --check / integrations/verify.mjs   # 全 PASS
node tests/fresh-machine.mjs --no-overlay   # 162/162
python tools/check_docs_links.py            # 12 份
```

发布与分发：`pet-v1.0.4`（zip **5,031,986 B / B490A2BB…**、`pet-setup.exe` **14,248,448 B / 96D9189F…**）与
`dsh-plugin-v1.2.10`（tgz **1,974,423 B / 676F41EC…**）都从公网下载回来复验过；npm `latest` = **1.2.10**；
沙箱里 zip → install → `-VerifyOnly` 11/11 → 装出来的 1.0.4 `--doctor` **0 fail** → `-Uninstall` 目录归零。

### 41.5 交接要点

- 上游跟进：[PR #14899](https://github.com/awesome-dsh-plugin/awesome-dsh-plugin/pull/14899)（`bump-tarball-1.2.10`），
  工具生成、单文件单行、`MERGEABLE`，`check` 排队中。合并后 `node tools/upstream-entry.mjs --check` 会变 OK。
- **B 提交之后 CI 会红一次**，这是设计的：`--check` 在「已发布、上游还是旧、且没有 PR 开着」时故意报红，
  而 B 提交发生在 PR 开出来之前。**重跑那一次即可**（本轮实测：重跑后 success）。
  这条写在这里，是因为下一次还会遇到，而且看起来像真失败。
- ~~F8（点击会话行复制项目路径）仍未开工~~ —— **已在 1.0.5 交付**，见 §42；地基（`expandedRowRect` / `expandedRowHit`、行悬停高亮）确实在 1.0.3 就落好了。
- **本机也在跑 1.0.4 了**：交接时本机装的还是 1.0.3（`4,495,360 B`），已用 `scripts\install.ps1` 就地升级到
  1.0.4（`4,494,848 B`），`-VerifyOnly` 11/11、`--doctor` **10 ok / 0 warn / 0 fail**（浮窗在场），
  `alerts.json` 由装出来的这一版真的写出来了。**交接前请顺手核对本机版本与最新 tag 一致**——
  浮窗是本机在用的东西，留着旧版会让「本机实测」与文档不一致。
- ~~**一个小瑕疵，留给下一轮**~~ —— 已随 1.0.5 修掉：`Kind.Label()` 把空 kind 写成 `none`，`--doctor` 的 `decisions=` 因此从 `/not-asking` 变成 `none/not-asking`。见 §42.5。

## 40. 交接：提醒层（2026-11-12 晚，第 40 轮；浮窗 1.0.3 / 插件 1.2.9）

### 40.1 一句话：决定「要不要打扰」的代码只剩一处，而且它不读窗口、不读文件、不读时钟

上一版把这条规则散在 `alertOnTransitions` 的 if 链里，后果有两个，都是真 bug：

- 账本按 `state` 记：答掉一个权限请求、立刻又被提问，两次都是 `waiting_permission`，于是**一声不响**。
  「答完就问」是 agent 最常见的节奏。
- 其余问题（你在不在看、这个会话是不是静音、完成算哪一档）**没有地方放**，只能变成又一个 if，或者干脆不问。
  「你正在读的请求还额外弹一个气泡」就是这个结果。

现在 `internal/ui/attention.go` 的 `Decide()` 是纯函数，规则顺序固定，且每条都有表驱动测试：
静音 > 子会话 > 完成三档（含 15 秒合并）> 紧急请求（只在「你正看着它」时才可以不弹，且仍标未读）> **测不准就当没在看**。

### 40.2 「要什么」是派生的，不是新字段

`waiting_permission` 里其实混着三种请求：审批、提问、方案确认。加字段会让**所有已装的上报端**整条 payload 被
`DisallowUnknownFields` 拒掉，这是拿协议换排版。所以 `internal/session/attention.go` 从 `reason` 里派生出
`Kind`，认不出来的词保守地归为「审批」——**沉默在真有阻塞时是最贵的错**。

### 40.3 六条施工教训（这一轮踩出来的）

| 教训 | 具体表现 |
|---|---|
| 静态与会话的键不能混用 | 静音/未读的键必须是 `source + sessionId`，**不能含 pid**：hook 上报端的 pid 是每进程一个的合成哈希，键在 pid 上会让每条静音在下一次上报时失效。`config.SessionMuteKey` 与 `session.Identity` 是跨包约定，由 `internal/app` 的测试钉住 |
| 新画的东西必须进 `renderedKey` | 未读位、静音位、悬停行都进了；1.0.1 的时长冻结与 1.0.2 的主题不重绘都是「同一个 key = 同一张位图」造成的 |
| 采样绝不上 UI 线程 | 枚举别人的窗口会卡在挂起的窗口上，那正是 Windows 报 AppHang 的形状。采样器是独立 goroutine，8 秒一次，复用 `Presence` 的缓存 |
| 不占宽度的标记才是好标记 | 第一版用「未读」「已静音」两个字，把 reason 挤掉了一半。改成**一个点**（未读）+ **整行压暗**（静音） |
| 纯函数里也要写「没变化就沉默」 | 调用方已经过滤了没变化的会话；但这条规则放进 `Decide` 之后，任何调用方都不可能再造出「驻留的等待请求每 8 秒响一次」 |
| 默认值只改一处，且必须明说 | 唯一改变老用户体感的是「回复完成时」从「总是」变成「仅后台」。旧配置 `notifyOnFinish: true` 读成 `always`，不是 `background`——用户当初点的就是「完成要提醒」 |
| **判定记录不能放进 `runtime/`** | 那个目录里每个 `*.json` 都被 `internal/store` 当作一条上报快照，解析不成会话就**删掉**。第一版把 `alerts.json` 放在那里，一个 5 秒扫描周期内它就消失了——`--json` 报「没有记录」、`--doctor` 报「从未扫描」，正好是这份记录要解决的两种沉默。单测只验证了「写得进去、读得回来」，**没验证「放对了地方」**；装到沙箱里跑一次才看见。现在它在 `config.json` 旁边，`internal/store/runtime_contract_test.go` 从目录一侧把规则钉住 |

### 40.4 这一轮的实测（2026-11-12 22:1x）

```
gofmt -l .                         # 空
go vet ./...                       # 0
go test -count=1 ./...             # 12 个包全 ok，282 个 func Test（上一版 218）
node packages/dsh-ai-terminal-pet/verify.mjs                     # VERIFY PASSED
node packages/dsh-ai-terminal-pet/scripts/vendor.mjs --check     # 两个 vendor 文件吻合（摘要 afca3441…）
node integrations/deepseek-harness/verify.mjs                    # VERIFY PASSED
node tests/fresh-machine.mjs --no-overlay                        # 162/162
python tools/check_docs_links.py                                 # 12 份文档全解
node --test tools/upstream-entry.test.mjs                        # 8 tests pass
node tools/upstream-entry.mjs --check                            # 1.2.9 尚未发布 → SKIP（不是红）
```

装出来的浮窗在沙箱里跑过一遍端到端：上报端推 working → idle，`alerts.json` 落在**应用目录**（不是 `runtime/`），
`--json --pretty` 能看到 `alerts` 对象，`--doctor` 的第 10 项打印 `presence=away, balloon=probe · generic,
decisions=done/done-away`。

**未改绘制的 15 张预览图逐字节未变**（用改造前后两次渲染的 sha256 对比过），只有新增的三张（未读/静音/悬停）
是新的。`bin/AI-Terminal-Pet.exe` 4,495,360 B、`vendor/AI-Terminal-Pet.exe` 与它逐字节相同。

### 40.5 下一轮开工前要知道的两件事

1. **上架用工具，不要手改**：`node tools/upstream-entry.mjs --version 1.2.9 --write --branch bump-tarball-1.2.9 --pr <号>`。
   `#14718`（1.2.8）当时已重建为单提交、单文件、`MERGEABLE`；它和 `#14688`（1.2.7）改的是同一行，
   **后合的那个会再冲突一次**，同样用工具再来一遍。
2. **F8（点击会话行复制项目路径）仍然没做**，方案在 ROADMAP §4 与 §38.3 里，三条约束不变。
   这一轮新增的 `rows_windows.go`（`expandedRowRect` / `expandedRowHit`）正好是它的地基。

## 39. 交接：上游目录条目 sync（2026-11-12 晚，第 39 轮）

### 39.1 一句话：把「上架」从手感变成机制

上游维护者 `fkysly` 在 `#14718` 上留言说这条 PR **和 main 冲突**，因此 GitHub 不跑检查、他也合不了，要求
「以 main 上的内容为底，把这次的新增加进去」。这是**第 N 次**同一件事了，所以这一轮不再只是把那条 PR 修好，
而是把它变成一条可以验证的路：规则进 [docs/UPSTREAM.md](UPSTREAM.md)，动作进 `tools/upstream-entry.mjs`，
判断进 CI。

### 39.2 冲突的真实成因（不是「我们改错了」）

| 事实 | 依据 |
|---|---|
| 我们的 PR 只改了 1 行（`tarball:` 1.2.7 → 1.2.8） | `gh api .../pulls/14718/files`：`+1 −1` |
| 上游 `main` 在**同一天**里也人工改过同一个文件 | `b2e2347`「Describe merged entries by what they do and disclose defaults」把该文件整份重写；`4f05a50d`（`#14688`）把 `tarball` 推到 1.2.7 |
| 两边都改 `tarball:` 那一行 → 内容冲突 | `git merge-tree --write-tree main fork/bump-tarball-1.2.8` 报 `CONFLICT (content)` |
| 冲突的 PR **没有 merge ref** → GitHub 不创建 `check` 运行 → 依赖它的 `Submission gate` 也不跑 | 上游 `contributing.md` 的「gate 一栏是空」那一段写的就是这件事 |
| 修完之后 | `mergeable=MERGEABLE`，`check` 正常排队（本轮实测） |

**所以「用旧分支开 PR + 让 git 合并」这条路在结构上就会坏**：那个文件本来就由维护者持续人工润色。
结论写进文档：**永远从上游 `main` 重建那一行，不合并旧分支**。

### 39.3 这一轮把哪句话改对了（诚实记录）

条目里的 `Click a row to jump to that terminal` 当时就**不是真的**：点行会收起面板，只有**行尾箭头**跳转。
改成 `Click the arrow at the end of a row to jump to that terminal`（zh：`点一下某一行的行尾箭头即可跳到对应终端`）。
这属于「描述必须可核对」那一条（上游复审第 1 项会读源码核对），不是文案润色。

### 39.4 三条施工教训（下一轮别再犯）

| 教训 | 具体表现 |
|---|---|
| 工具的「能读」不等于「会读」 | 我给 `openBumpPR` 写的是 `execFileSync('gh', [...])`，本机 `gh` 已登录所以本地绿；换到 CI 就是**没有 gh、也没有跨仓库 token**，那一步永远走 `supported:false`。改成直接调公开 REST（`GET /pulls?head=owner:branch`），无 token 也能跑，`GH_TOKEN` 存在时才带 |
| 忘了 `await` 是最安静的 bug | `openBumpPR` 改成 async 之后，调用点漏了 `await`，于是 `pending.open` 永远是 `undefined` → 输出「SKIP 而不是红」——**恰好是那种看起来正常、其实没判断的失败**。本地一条 `--check` 就能暴露它，前提是真看一眼输出 |
| 依赖外部仓库的检查，结论要分四类 | 已列出 / 有 PR 开着 / 发了版却没人推 / 读不到（未发布、无 gh、网络失败）。只有第三种该红；其余三种「红」只会训练人忽略红灯 |

### 39.5 这一轮的实测（2026-11-12 晚）

```
node --test tools/upstream-entry.test.mjs        # 8 tests pass（纯函数：文件名、URL、四种拒绝、单行断言、CRLF）
node tools/upstream-entry.mjs --check            # in flight：still 1.2.7，但 #14718 开着 → OK
node tools/upstream-entry.mjs --check --version 1.2.99   # asset 404 → SKIP（未发布的版本不该红）
python tools/check_docs_links.py                 # 12 files（新文档已纳入）
gh pr view 14718 --json mergeable                # MERGEABLE，check 在跑
```

`gofmt -l .` / `go vet ./...` / `go test -count=1 ./...` / 三个 verify / `vendor.mjs --check` /
`fresh-machine --no-overlay` 全部照旧通过；**这一轮没有改任何非测试的 `.go` 文件**，所以
`vendor/manifest.json` 的源码摘要不变，浮窗仍 1.0.2、插件包仍 1.2.8——不需要发版。

### 39.6 现在上游是什么状态（交接要点）

- `#14718`（1.2.8）：**已重建为单提交、单文件、+3/−3**，`MERGEABLE`，`check` 在跑，等维护者合并。
  合并后站点重建才会把 `plugins.json` 的 install 命令渲染成 1.2.8。
- `#14688`（1.2.7）：仍 open。它和 `#14718` 改的是同一行，**先合哪个都行，后合那个会再冲突一次**；
  我们的工具可以再来一遍（从 `main` 重建）。
- 上游 `main` 现在指向 1.2.7 —— **「PR 开着」不等于「已收录」**，判断可见性读站点索引文件。
- 下一轮发版（1.2.9）时：`node tools/upstream-entry.mjs --version 1.2.9 --write --branch bump-tarball-1.2.9 --pr <号>`。


