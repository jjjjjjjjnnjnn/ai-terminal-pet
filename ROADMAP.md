# ROADMAP — ai-terminal-pet

> 快照时间 2026-09-26。本文分四块：现状基线、本轮已完成、扩展计划、轻量化计划，
> 以及一组实用功能候选与非目标。每项都标注了**动机**和**为什么现在不做/怎么做**，
> 便于下一位维护者直接挑一项开工。

## 0. 现状基线

| 维度 | 数值 |
|---|---|
| 平台 | Windows only（Win32 分层窗口 + 命名管道） |
| 工具链 | Go 1.23+（实测 1.27.0），reporters 仅用 Node 标准库 |
| 主程序 | `bin/AI-Terminal-Pet.exe` 3,954,688 B（`-trimpath -ldflags "-s -w"`） |
| 状态行 shim | `bin/pet-statusline.exe` 3,644,416 B |
| 代码量 | Go 6.3k 行 / JS 1.2k 行 |
| 第三方依赖 | 3 个，均为库引用、无 vendor：go-winio(MIT) / lxn/win(BSD) / x/sys(BSD) |
| 运行期占用 | 有活跃会话时工作集约 22–38 MB（含共享页），线程 8–16 |
| 协议 | `ai-terminal-pet.v1`，每行一个 JSON，上限 1 MiB，回执精确 `OK` / `ERR` |
| 管道 | `\\.\pipe\ai-terminal-pet-v1-<homedir-sha256 前 8hex>`（按用户隔离，跨语言 golden 测试锁定） |
| 上报端 | `opencode.js` / `claude-hook.js` / `antigravity-hook.js` / `antigravity-statusline.js` / `generic.js` / `pet-statusline.exe` |
| 生命周期 | 按需拉起（reporter 发现管道不存在就拉起）+ 无会话自毁（5 s 宽限，会话 TTL 30 s） |
| 持久化 | `state.json`（位置/缩放/展开）· `config.json`（通知/提示音/置顶/不透明度）· `HKCU\...\Run`（开机自启） |
| 对外接口 | `--show/--toggle/--reset/--quit/--json[--pretty]/--autostart[=on\|off]/--version/--help` |
| 单测 | 55 个（config 9 / autostart 8 / app 5 / store 10 / ui 3 / session 8 / pipeserver 12） |

## 1. 本轮已完成

### 1a. 2026-09-26（v0.3.0）—— F3 / F4 / F7

- **F3 开机自启**：新增 `internal/autostart`，写 `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`
  （用户级、免管理员、停用即删除，卸载无残留）。托盘开关 + `--autostart[=on|off]` CLI。
  启动时自动修复指向旧路径的条目——安装路径是 Run 值里唯一的内容，程序被移动或重装后
  会在登录时静默失败，那是最糟的失败方式。`Sync` 永不把「关闭」的条目改成开启。
  注册表经 `backend` 接口访问，8 个单测全部跑在内存后端上——单测绝不该悄悄打开
  `go test` 执行者的开机自启。
- **F4 提示音**：`config.json` 新增 `sound`。等待确认用 `MB_ICONEXCLAMATION`、错误用
  `MB_ICONHAND`，走系统声音方案，因此与桌面其余部分一致并遵守专注助手/静默时段。
  与气泡提醒解耦但共享同一套「会话+状态」去重，关掉任一个都不会改变另一个的触发时机。
- **F7 `--json` 导出**：`app.Snapshot()` 直接读 `runtime/` 快照离线聚合，**不依赖浮窗进程**，
  因此可以被状态栏或脚本安全调用而不会阻塞在 GUI 进程上；`--pretty` 切换缩进输出。
  代价是新鲜度受快照 TTL（15 s）约束，这一点与浮窗完全一致。

### 1b. 2026-09-25（v0.2.0）—— 打磨

- **记住位置**：拖拽/缩放/展开后落盘 `state.json`，下次启动原样恢复（含多显示器钳制）。
  写入走后台单协程合并，避免在 UI 线程上做文件 IO（本机 AV 钩挂下写文件可达数百毫秒）。
- **CLI 控制**：`--show` / `--toggle` / `--reset` / `--quit` / `--version` / `--help`。
  第二实例不再是「静默退出」，而是激活已存在浮窗。用法错误走 stderr + exit 2，不写 error.log。
- **托盘升级**：tooltip 实时显示「N 会话 · 状态」；进入等待确认/错误时弹气泡提醒（按会话+状态去重，可关）；
  气泡点击回跳浮窗；菜单新增「重置位置与大小 / 状态提醒 / 窗口置顶」开关。
- **配置档** `config.json`：通知开关、置顶、不透明度。缺失或损坏一律回落默认值，绝不影响启动。
- **轻量化**：脉冲动画定时器改为按状态自适应——只有按压/等待/错误态才装 50 ms 定时器，
  其余时间直接 `KillTimer`（此前是 7×24 的 20 FPS 空转）；DIB 画布与命中缓冲复用，去掉每帧 GDI 分配。
- **通用上报器** `reporters/generic.js`：任意 CLI 一条命令接入（见 §2 E2）。
- **修复**：`store.processBacked` 由「黑名单」改为「允许列表」。原实现把未知 source 当成真实进程，
  而 `generic.js` 与所有 hook 上报端用的是稳定哈希 PID，存活探测必然失败，
  于是**刚上报的会话立刻被删**——上报端回 `ok`，浮窗里却永远看不到。
  允许列表的默认值更安全：探测一个不可信的 PID 会删掉活会话，而放过一个已死会话
  只是让它多留一个 TTL。
- **新增单测**：`internal/config` 7 个、`internal/ui` 3 个、`internal/app` 5 个（离线聚合）、
  `internal/autostart` 8 个、`internal/store` 2 个（合成 PID 保留 + 非进程源保留）。

## 2. 扩展计划

### E1 — 把 Node 上报端全部原生化（收益最高）
**动机**：本机每创建一个进程约 **1 s**（AV 钩挂），而 `claude-hook.js` / `antigravity-hook.js`
每次事件都要起一个 `node`（实测 `node -e exit` 0.9–1.2 s）。`pet-statusline.exe` 已经是这条路的先例：
Go 原生 shim 稳态 0.5–0.9 s，比 node 路径的 13–17 s（经 runner）低一个数量级。
**做法**：在 `cmd/` 下加 `pet-hook.exe`，复用 `pet-statusline` 的 stdin 解析 + 状态映射 + 管道投递骨架，
用 `--mode=claude|antigravity` 区分事件映射。三个 JS hook 保留为参照实现与降级路径。
**验收**：同 payload 下原生与 JS 输出同一 session PID（FNV-1a UTF-16 已对齐）与同一 state。

### E2 — 通用接入成为一等公民
**动机**：`reporters/generic.js` 已能一行接入任意工具，但协议本身还只写在代码里。
**做法**：新增 `docs/PROTOCOL.md`，把 `ai-terminal-pet.v1` 的字段、状态语义、优先级、
管道命名算法、回执契约、脱敏规则写成规范；README 给出 5 行以内的接入示例（shell alias / Makefile / 编辑器 task）。
**验收**：不读源码也能写出一个合规上报端。

### E3 — 跨平台（macOS / Linux）
**动机**：目前 UI 层与平台层直接调 Win32，无法移植。核心逻辑（session/store/pipeserver 的协议部分）其实已平台无关。
**做法**：把 `internal/ui` 的窗口实现抽成接口（`Presenter`），Win32 实现留在 `_windows.go`，
新增 `_darwin.go` / `_linux.go`（macOS 用 NSWindow + NSVisualEffectView，Linux 用 GTK/Wayland layer-shell）。
IPC 从命名管道换成 Unix domain socket，`PipePathForUser` 的 hash 算法保持不变以免破坏跨语言契约。
**注意**：这是一次真正的架构改动，建议先做 E2（协议文档）再动，否则契约会跟着实现漂移。

### E4 — 会话可交互
**动机**：现在展开面板只能「看」，看到某个会话在等待却无法跳过去。
**做法**：展开态会话行做命中检测（几何与绘制同源，参照 `expandedCloseButtonRect` 的做法），
点击时用 `EnumWindows` + `GetWindowThreadProcessId` 找到该 PID 的顶层窗口并 `SetForegroundWindow`。
**限制**：`claude-hook` / `antigravity-*` 的 PID 是稳定哈希而非真实进程号，这些行不响应点击；
要么保持「无窗口则无动作」，要么顺带给这些上报端补上真实 PID（与 E1 一起做更自然）。

### E5 — 配置档扩展
**做法**：`config.json` 增加 `theme`（配色组）、`size`（紧凑/展开的基准尺寸预设）、
`perSourceColor`（按 source 指定强调色）、`hidden`（启动即隐藏到托盘）。
`internal/config` 的解码已是「缺字段用默认」，加字段不会破坏旧文件。

### E6 — `--doctor` 自检
**动机**：本轮的排查花了大量时间在「进程活着但窗口不可见」这类无声故障上。
**做法**：`--doctor` 打印并校验：进程/互斥体状态、窗口 `WS_VISIBLE` 与消息泵响应性
（`SendMessageTimeout` + `SMTO_ABORTIFHUNG`，本轮探针已实现）、管道名与 ACK 往返、三处终端配置是否接入、
`runtime/` 快照与 `error.log`/`hang.log` 现状。把本轮的临时探针固化成产品功能。

## 3. 轻量化计划

### L1 — 进程创建税（当前最大开销，~1 s/次）
**现状**：Antigravity 的 `statusLine` **每次渲染**都要起一个进程；`pet-statusline.exe` 自身
实测 0.95–1.5 s，其中绝大部分是进程创建被 AV 钩挂扫描，不是我们的代码。
**可做的**：
1. 缩小二进制（见 L2）——镜像越小，AV 全量扫描越便宜。
2. 上报端在**状态未变化时直接短路**：`pet-statusline` 先读 `runtime/` 里自己的上一个快照，
   若 `state`/`reason`/`project` 全同则跳过管道投递（快照仍写，保证 TTL）。可省掉每次渲染的 IPC 往返。
3. AV 排除目录（需管理员，用户决策）——BASELINE §13 已记录这是唯一能真正压到亚秒的手段。
4. 已落地：`--reset`/`--toggle` 等控制路径不再走 `app.Run()`，避免控制命令意外拉起常驻实例。

### L2 — 二进制体积（3.7 MiB → 目标 <2.5 MiB）
**做法**：`-ldflags "-s -w"` 已用；继续可做 `-buildvcs=false`、剥离 `net/http` 等未用包的
反射元数据（本项目不用 http，但 `net` 被管道复用）、用 `-tags` 关掉不需要的 `crypto` 实现。
注意 `pet-statusline` 引了 `regexp`（脱敏规则）——这是体积大头之一，可改为手写扫描器。
**验收**：`go build` 后两个 exe 合计 < 5 MiB，且 `node --check` 与全部单测不变。

### L3 — 管道连接的 goroutine churn
**现状**：`hang.log` 里出现过 goroutine 编号 79 万+，说明 10 小时内创建了约 80 万个短命 goroutine——
go-winio 的 `makeConnectedServerPipe` 每次 Accept 前都会预建一个 `connectPipe` 等待者。
单次开销极小，但频率高（statusline 高频上报时可达 20+/s）。
**做法**：先量化（在 `Server` 加计数器），确认是否需要处理；若需要，考虑给服务端加连接复用
或把上报端改为「变化才连」。

### L4 — 光栅化与重绘
**已做**：DIB 画布复用、命中缓冲复用、动画定时器按状态装/卸。
**可做**：脏区重绘（当前每帧全画），或把 `fillRounded`/`fillEllipse` 的内层循环向量化。
面板只有 40×190 / 200×250 逻辑像素，收益有限——**建议先测再动**，别为 0.3 ms 增加复杂度。

### L5 — 启动与扫描路径
**现状**：快照扫描 2 s 一次，全目录 `ReadDir`。会话多、`runtime/` 残留多时会变贵。
**做法**：`runtime/` 里按 mtime 排序 + 只扫最近 N 个；或给扫描加「目录 mtime 未变则跳过」。
仅在实测扫描耗时 >5 ms 时才值得做。

## 4. 实用功能候选

已完成：~~F3 开机自启~~、~~F4 提示音~~、~~F7 `--json` 导出~~（见 §1a）。

| 编号 | 功能 | 价值 | 复杂度 |
|---|---|---|---|
| F1 | 悬停展开面板时显示完整会话明细（tooltip 或就地浮层） | 高 | 低 |
| F2 | 键盘快捷键：`Esc` 收起（已有）、方向键在会话行间移动、`Enter` 跳转 | 中 | 中 |
| F5 | 记住所在显示器（多屏场景下按显示器恢复，而不是全局坐标） | 中 | 中 |
| F6 | 会话时长/最近活动时间显示（数据已在 `StartedAt`/`UpdatedAt`） | 中 | 低 |
| F8 | 会话行右键：复制项目路径 / 在资源管理器中打开 | 低 | 低 |

F6 和 F8 都是「一个下午」的量级，且不触碰渲染核心，适合作为下一次迭代的起点。
F1 需要 tooltip 或多行文本布局，是展开面板的第一次真正的版式改动，建议单独一轮做。

**已交付但值得回头看的一项**：`--json` 目前只读快照，因此看不到「只走管道、不写快照」
的上报路径（如果将来有上报端选择不落盘）。真需要时可以给它加一个查询运行中实例的
IPC 分支，但要在 `--json` 永远不阻塞 GUI 进程这个约束下做。

## 5. 非目标

- **不做终端复用器或日志查看器**：本项目的定位是「一眼可辨的状态指示」，不是替代终端。
- **不做账号/云同步/遥测上报**：全部数据留在本机，`runtime/` 与 `state.json` 不离开 `%LOCALAPPDATA%`。
- **不引入 UI 框架**：零外部 UI 依赖是有意为之，它让二进制保持单文件、无运行时依赖。
- **不做跨用户共享的管道**：管道名按用户哈希隔离是安全边界，不为了「多用户共用」而放宽。

## 6. 已知残留（不计划修，但需知情）

- `hang.log` 看门狗是**事后取证**而非自愈；本轮已把转储上限压到 3 次并记录最后消息名。
- UI 无自动化视觉回归，靠 `--doctor`（待做）+ 探针脚本人工验证。
- `go test -race` / `govulncheck` / `gosec` 因缺工具链（gcc 等）仍未跑。
- 同用户命名管道抢占只做了「互斥体先行」缓解，未做 `FILE_FLAG_FIRST_PIPE_INSTANCE` 独占。
