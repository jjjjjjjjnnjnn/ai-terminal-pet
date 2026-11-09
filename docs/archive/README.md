# docs/archive — 历史文档索引

本目录存放**已冻结**的历史文档：它们记录的是某一时刻的取证与状态，不再随代码更新。
当前有效的 runbook 在 [docs/HANDOVER.md](../HANDOVER.md)，计划在 [docs/ROADMAP.md](../ROADMAP.md)，
按版本的变更摘要在 [CHANGELOG.md](../../CHANGELOG.md)。

约定：归档文件**不做内容修订**（拼写、过时结论、当时的错误判断都原样保留），只在文件顶部加归档头。
若要更正历史结论，写在现行文档里并在此索引注明。

---

## 1. [BASELINE-20260924.md](BASELINE-20260924.md) — Phase 0 冻结（停用态）

**覆盖什么**：2026-09-24 项目「全局停用」当日的一次完整取证与随后四阶段重新启用的记录。

| 小节 | 内容 |
|---|---|
| §1–§2 | 进程与安装态（`AI-Terminal-Pet.exe.disabled` 3,485,696 B、SHA256 `02A14DD8…`）、`runtime/` 残留 89 个 `*.json` / 0 `*.tmp` |
| §3 | 上报入口停用态干净表格（opencode / claude settings + hooks / agy settings / notifier hooks 六份配置的 SHA256） |
| §4 | 五个 reporter 的 SHA256 基线 |
| §5 | Phase 0 审核（可维护性 / 可交接性 / 合规 / 安全） |
| §6–§7 | 需用户完全退出终端的步骤；Phase 1 清空 `runtime/`（89 → 0，30 s 复查仍 0） |
| §8 | Phase 2 加固（per-user 管道名 + 抢占缓解、`read()` 半关闭契约、`TestStatePriorityOrder` 锁定优先级并纠正 `tool_use` 笔误） |
| §9 | Phase 3 构建与安装（装机 hash `3204889A…`、冒烟 PID 27788 / 管道 `42ce7d78`） |
| §10 | Phase 3 配置恢复（四份 `.bak-20260924-reenable`；Claude 7 事件全落 `settings.json` 而非 `hooks.json` 的偏差记录） |
| §11 | Phase 4 隔离 harness 9/9；harness 误触生产管道拉起 PID 4332 |
| §12 | Antigravity `statusLine` 引号故障（runner 不做 shell 解析，内层双引号被当路径字面量） |
| §13 | **原生 statusline shim 的立项依据**（`node -e exit` 4.3 s / `cmd /c exit` 7.7 s / `powershell` 15 s，McAfee+Defender 钩挂；`cmd/pet-statusline` 的设计与验收） |
| §14 | 浮窗点击/悬停 4 缺陷（`hitTest` 越界 panic、展开/缩放/DPI 不改尺寸、输入同步 `present` 致 AppHang、滚轮方向反转） |
| §15 | 看门狗 `watchHang` / `writeHangLog` 的设计（6 s 无分发写 `hang.log`，0600，512 KB 轮转，60 s 节流，pprof goroutine 栈） |
| §16 | **「窗口不可显示」事件：系统级 win32k/DWM 楔子怀疑**——⚠️ **该结论已被推翻**，见下 |
| §17 | 开源交接（MIT/BSD 库引用、无复制代码、密钥扫描零命中、`*.exe` 忽略） |

**被谁取代**：
- §1–§11 的停用态与重新启用流程 → [HANDOVER-2026-09-DISABLE.md](HANDOVER-2026-09-DISABLE.md)（原 HANDOVER §1–§12 逐字归档）；
  其中 §9–§11 仍被 [docs/HANDOVER.md](../HANDOVER.md) §12 的原文（现已归档）直接引用。
- §13–§15 的设计与验收 → [docs/HANDOVER.md](../HANDOVER.md) §13–§15（`cmd/pet-statusline` 仍在服役，
  `sound_windows.go` 已在 v0.8.0 取代 `MessageBeep`，看门狗参数仍是当前值）。
- **§16 的结论被 [BASELINE-20260925.md](BASELINE-20260925.md) §3 推翻**：真因是 `internal/ui/app_windows.go` 的
  `Run()` 缺少 `runtime.LockOSThread()`，不是系统级楔子。看到 §16 时请以 BASELINE-20260925 为准。

---

## 2. [BASELINE-20260925.md](BASELINE-20260925.md) — 浮窗全线故障根因定位与修复

**覆盖什么**：2026-09-25 对「浮窗显示不出来 / 点击像崩溃 / 拖不动 / 定时器全死」的完整定位过程。
开篇即声明推翻 BASELINE-20260924 §16 的系统级楔子结论。

| 小节 | 内容 |
|---|---|
| §1 | 报障清单 |
| §2 | 现场取证（PID 56840、hash `71078FA7…`、hwnd `0x3d0852`、`IsWindowVisible` false、`SendMessageTimeout` 立即失败、`hang.log` 370 KB / 6,544 行同一条 `goroutine 1 [syscall]`→`win.GetMessage` 栈） |
| §3 | **根因**：`runtime.LockOSThread()` 缺失；以及定位手法（埋点二分、`GetMessage` 栈会误导、真正阻塞点是 `SetWindowPos`） |
| §4 | **第二 bug**：`endDrag` 在 `ReleaseCapture()` 之后读 `a.drag.moved` |
| §5 | 本轮代码变更表（5 个文件，含 `hitMap`→`hitMapInto` 去掉每帧 50 KB 分配） |
| §6 | 验证表（装机 hash `9C70D60B…`；40×190→200×250、滚轮 40×190→44×209、拖拽 1:1、二次启动 202 ms、无会话干净移除） |
| §7 | 终端接入收敛（移除 Antigravity 的 `BeforeTool`/`AfterAgent` node 钩子，备份 `settings.json.bak-20260925-pet-hooks`） |
| §8 | 遗留 4 条 |

**被谁取代**：[docs/HANDOVER.md](../HANDOVER.md) **§13** 是它的摘要版（`cbc043e` 提交）。两份并存：
HANDOVER §13 给结论与后续 runbook，本文件给逐条证据。§4 的 `decideRelease` 与 §6 的验证矩阵至今有效。

---

## 3. [HANDOVER-2026-09-DISABLE.md](HANDOVER-2026-09-DISABLE.md) — 原 HANDOVER §1–§12（停用态）

**覆盖什么**：2026-09-24「项目全局停用」时期 HANDOVER 的前 12 节，**逐字搬迁、未删任何技术细节**，
章节号沿用原文。归档于 2026-11-09，因为 §1/§2/§4/§9/§10 描述的是 `AI-Terminal-Pet.exe.disabled`
时代的状态，与当前 v0.9.0 运行态不符。

| 原节 | 内容 | 现状 |
|---|---|---|
| §1 | 停用态一句话状态（浮窗进程不存在、三处上报入口已摘除） | 已被 HANDOVER §1（v0.9.0 现状）取代 |
| §2 | 关键路径（安装路径「当前不存在」、部署脚本「停用期间禁止运行」） | 已被 HANDOVER §2 取代（当前安装目录内容见新 §2） |
| §3 | 架构速览（管道名算法、回执契约、store 上限、UI 默认几何 40×190 / 200×250） | 几何已过时（v0.4.0 起 56×156 / 200×内容高度）；管道与回执契约**仍有效** |
| §4 | 停用态配置触点（三处上报入口的移除清单，可逆条目） | 已在 2026-09-24 夜按 §12 加回，当前重新启用流程见 HANDOVER §1 |
| §5 | 已知残留（旧终端内存里的旧 reporter 会继续写惰性快照） | 现象仍可复现，处理办法（终端重启后清 `runtime/`）仍适用 |
| §6 | 本轮已落地的关键修复（按文件：protocol.js / pipeserver / store / session / platform / app / ui / main.go / 用户侧 hook） | 大部分代码仍在服役，是理解协议边界的**最好入口**之一 |
| §7 | 验证基线（覆盖率 session 76.8% / store 75.3% / pipeserver 53.4%；隔离 harness 9/9；未跑项） | 数字是 2026-09-24 的；未跑项（`-race`/`govulncheck`/`gosec`）至今仍缺工具链 |
| §8 | 常用命令（PowerShell + `rtk` 前缀） | 仍可用；`rtk` 是本机的外部包装，非仓库内容 |
| §9 | **禁忌（停用期间）** | ⚠️ 已失效：当前项目在役，这些约束不再适用 |
| §10 | **重新启用步骤（仅在用户明确要求时）** | ⚠️ 已失效：当前项目已启用，正常安装见 `install.ps1` |
| §11 | 残留风险（管道抢占只缓解、UI 无自动化测试、`socket.end` 半关闭语义弱、node 冷启动 1.4 s、沙盒豁免） | 五条**全部仍然成立**（详见 HANDOVER §11 原文与本文件） |
| §12 | 重启用记录 2026-09-24 夜 + 当日五次补充 | 是停用态 → 在役的转折点；后续演进见 HANDOVER §13 起 |

**「§16」在本文件里的含义**：§12 与 §13 附近出现的「§16」指的是
[BASELINE-20260924.md](BASELINE-20260924.md) §16（系统级楔子结论），不是 HANDOVER 自己的 §16（面板重做）。
该结论已被 [BASELINE-20260925.md](BASELINE-20260925.md) 推翻。

---

## 4. [promo-2026-11-09/](promo-2026-11-09/README.md) — 对外宣传文案的生产记录

**覆盖什么**：2026-11-09 用本机 agy（`gemini-3.8-flash-high`）重写对外 README 与发布文案的全部中间产物：
事实底座 `facts.md`（134 行、每条带 `文件:行号`）、三轮 prompt（初稿 / 定点修订 / SOCIAL 手术）、
`agy-notes.md`（耗时表与自评）、`SOCIAL.md`（三组社交文案，尚未发布）。

**为什么归档**：这批稿子的结论已经进了 [README.md](../../README.md)，但「每条事实从哪来、哪几句是 agy 写的、
哪几句是人改的」需要留证据。产物本身不再修订。

---
## 5. 目录里还有两份非归档文件

[docs/assets/README.md](../assets/README.md) 与 `render_assets.py` 不在本索引里——它们不在 `docs/archive/`。
`docs/assets/` 下 8 张 PNG 由二者的脚本产出，随代码更新，因此不属于冻结文档。
说明一下以免混淆：README 里的全部配图都是 `TestRenderPreview` / `TestRenderSettingsPreview` 的真实
Win32 光栅化输出，经 `render_assets.py` 裁掉平背景后生成，因此不会与实际布局漂移。
