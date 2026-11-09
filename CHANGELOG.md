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

未发布版本：无。

---

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
