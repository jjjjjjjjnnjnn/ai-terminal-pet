# HANDOVER — ai-terminal-pet 交接文档

> 交接时间：2026-09-24（UTC+8 下午）。下一位 AI 直接从“当前状态”开始读，不必回溯长上下文。

## 1. 当前状态（一句话）

> **本节已过期，最新状态见 §13（2026-09-25 根因修复）。** 下面保留作历史记录。

**项目全局停用中**：浮窗进程不存在；安装目录可执行文件已改名为 `AI-Terminal-Pet.exe.disabled`；
OpenCode / Antigravity / Claude 三处上报入口已全部摘除并校验通过；源码保留、可随时重新启用。

## 2. 关键路径

- 源码：`C:\Users\rongj\Desktop\项目开发\ai-terminal-pet`（Go module `ai-terminal-pet`，`go 1.23.0`）
- 入口：`cmd/ai-terminal-pet/main.go`
- 安装路径：`C:\Users\rongj\AppData\Local\AI-Terminal-Pet\AI-Terminal-Pet.exe`（**当前不存在**，只有 `AI-Terminal-Pet.exe.disabled`）
- 构建产物：`bin/AI-Terminal-Pet.exe`
- 快照目录：`C:\Users\rongj\AppData\Local\AI-Terminal-Pet\runtime`（*.json / *.tmp）
- 错误日志：`C:\Users\rongj\AppData\Local\AI-Terminal-Pet\error.log`（0600，超 1MB 轮转）
- 部署脚本：`install.ps1`（只停 `AI-Terminal-Pet` 进程 → 拷贝 → SHA256 校验；**停用期间禁止运行**）

## 3. 架构速览

- 单实例（`Local\AI-Terminal-Pet-v1-SingleInstance` mutex）→ Win32 分层置顶浮窗（`internal/ui`）。
- IPC：Windows Named Pipe，名称按用户隔离：`\\.\pipe\ai-terminal-pet-v1-<homedir-sha256前8hex>`。
  Go（`internal/pipeserver.PipePathForUser`）与 JS（`reporters/protocol.js: pipeScopeToken`）同算法，
  已有跨语言 golden 测试锁定（`TestPipePathForUserMatchesReporterAlgorithm`）。
- 服务端回执：合法 `OK\n` / 非法 `ERR\n`；客户端只接受精确 `OK`，否则 `EBADREPLY` 重试至 deadline。
- 状态聚合：`internal/store`（上限 256 会话 / tombstone 10 分钟 / pipe TTL 30s / 快照 TTL 15s / 快照文件上限 256）。
- 推送端：`reporters/opencode.js`、`claude-hook.js`、`antigravity-hook.js`、`antigravity-statusline.js`，经 `reporters/protocol.js`（schema `ai-terminal-pet.v1`，payload 上限 1MiB）。
- UI：默认竖条 `40x190`（logical），展开 `200x250`，屏幕右侧居中；等待/错误呼吸提醒，按压反馈，动画仅提醒态 20FPS。

## 4. 配置触点（当前值 = 停用态）

- `C:\Users\rongj\.config\opencode\opencode.json`：`plugin` 仅剩 `@leo000001/opencode-notify-native`。
- `C:\Users\rongj\.gemini\antigravity-cli\settings.json`：无 `statusLine` 块；保留安全沙盒（`enableTerminalSandbox:true`，`toolPermission:proceed-in-sandbox`）。
- `C:\Users\rongj\.gemini\config\plugins\antigravity-windows-notifier\hooks.json`：仅剩原生 `notify.js`（PreToolUse attention / Stop）。
- `C:\Users\rongj\.claude\settings.json`：无 `hooks` 块（仅 env + includeCoAuthoredBy）。
- `C:\Users\rongj\.claude\hooks\hooks.json`：仅 ECC 官方条目，无 `ai-terminal-pet` 引用。
- 本轮移除的可逆条目（重启用时按原样加回）：
  OpenCode 的 `file:///.../reporters/opencode.js` 插件行；
  Antigravity 的 `statusLine.command = node .../reporters/antigravity-statusline.js`；
  Antigravity hooks 中两处 `antigravity-hook.js`；
  Claude settings + hooks.json 中 7 处 `claude-hook.js`。
- 注意：只有 Antigravity 有 `.bak-20260923*` 备份；opencode/claude 的 settings 是直接改的，恢复靠本文第 4 节描述。

## 5. 已知残留（非故障，预期内）

- 仍在运行的旧终端（尤其 OpenCode PID 39796 一系）内存里还载着旧 reporter，会继续向 `runtime/` 写**惰性快照**直到这些终端完全退出重开。
  这些文件无消费端（进程无、可执行文件缺失），不会复活浮窗；终端重启后停止产生；之后可安全删除 `runtime/*.json,*.tmp`。
- Antigravity 运行中进程沿用启动时加载的旧 `statusLine` 配置直到重启；但源文件已改为 fail-fast（见第 6 节），报错计数应停止增长。

## 6. 本轮已落地的关键修复（按文件）

- `reporters/protocol.js`：stdin 上限 1MiB/2s；`closed` 不自动拉起；exe 路径限制在安装根内；`stableProcessId`；超长 ID hash 化；C1 控制字符剔除；导出 `executablePath`。
- `reporters/antigravity-statusline.js`：exe 缺失时跳过 `sendStatus` 只做通知转发；转发改进程内 `require`（渲染 3.7s→1.7s）。
- `reporters/antigravity-hook.js`、`claude-hook.js`：PID 改用 `stableProcessId(source:session)`，不再依赖短命 shell PPID。
- `reporters/opencode.js`：初始化重试+服务端时间锚定；sessions 上限 64 并清理注册表。
- `internal/pipeserver/server_windows.go`：连接上限 32、读超时 5s、payload 界、ACK/ERR、shutdown 等待、ACL 收紧为 `GRGW/OW+SY`、per-user 管道名。
- `internal/store/{store,snapshot}.go`：tombstone 时序、pipe TTL/上限、快照预检大小+`LimitReader`+`SameFile` 防替换删除、symlink 拒绝、stale tmp 清理（TTL，≤64/轮）、hook 源跳过 PID 存活检查。
- `internal/session/session.go`：PID/ParentPID 32 位界。
- `internal/platform/process_windows.go`：`ACCESS_DENIED` 视为存活（保守不清掉）。
- `internal/app/app_windows.go`：新 pipe 事件唤醒隐藏窗口（`wmActivate` 合并）；scan 监听 ctx；server 错误透出到 `Run` 返回。
- `internal/ui/*`：`wmActivate` 合并、`showFromTray` 先显示后 post-refresh（避免同步 present 死锁）、托盘图标 Explorer 重启后重建、DPI 钳制 72–600、画布上限 4096px/16M 像素、动画仅提醒态、展开行收进 250px、删死代码。
- `cmd/ai-terminal-pet/main.go`：错误日志追加 + 1MB 轮转。
- 用户侧 `hooks/notify.js`：stdin 3s 兜底 + 干净退出；toast 子进程 detached+unref（Stop 耗时 3.7s→1.45s）。
- `hooks/statusline-notify.js`：抽出 `handleRawInput` 并导出（保持 CLI 行为不变）。

## 7. 验证基线（停用约束下全部通过）

- `gofmt -d` clean；`go vet ./...` exit 0；`go test -count=1 ./...` 全过。
- 覆盖率：session 76.8% / store 75.3% / pipeserver 53.4%；app/platform/ui 无测试。
- `node --check`（5 个 reporter + notify.js + statusline-notify.js）全过；5 份接入配置 JSON 解析全过。
- 隔离 harness（`C:\Users\rongj\AppData\Local\Temp\opencode\pet-harness-check.js`，用测试管道名 + 临时 LOCALAPPDATA）：9/9 通过
  （管道命名隔离 / 稳定PID / 长ID / 脱敏 / ACK-OK / 垃圾回复拒绝 / 挂起有界 / 缺失管道不拉起 / 快照隔离）。
- `opencode debug config` ok；`agy plugin validate` ok。
- 未跑项：`go test -race`（缺 gcc/cgo）、`govulncheck`/`gosec`（未安装）、真实终端联调（按停用约束推迟）。

## 8. 常用命令（Windows PowerShell，前缀 rtk 以省 token）

```powershell
rtk "C:\Program Files\Go\bin\go.exe" test -count=1 ./...
rtk "C:\Program Files\Go\bin\go.exe" vet ./...
rtk "C:\Program Files\Go\bin\go.exe" build ./cmd/ai-terminal-pet
rtk node --check .\reporters\protocol.js
Get-Process AI-Terminal-Pet -ErrorAction SilentlyContinue
Get-ChildItem "$env:LOCALAPPDATA\AI-Terminal-Pet\runtime" -ErrorAction SilentlyContinue
```

## 9. 禁忌（停用期间）

- 不要启动 `AI-Terminal-Pet.exe`，不要运行 `install.ps1`，不要构建后覆盖安装。
- 不要用生产管道名做测试；JS 测试一律临时 `LOCALAPPDATA` + 测试管道名；Go 用 `net.Pipe`。
- 不要为验证而接入真实终端；不要杀用户/当前会话的 OpenCode、Antigravity 进程。

## 10. 重新启用步骤（仅在用户明确要求时）

1. `go build -trimpath -ldflags "-s -w" -o .\bin\AI-Terminal-Pet.exe ./cmd/ai-terminal-pet`
   + `go build -trimpath -ldflags "-s -w" -o .\bin\pet-statusline.exe ./cmd/pet-statusline`
2. 运行 `.\install.ps1`（停浮窗进程 → 双二进制拷贝 → 逐个 SHA256 校验；若 `AI-Terminal-Pet.exe.disabled` 仍在可删，避免混淆）。
3. 按第 4 节把三处上报入口加回。
4. 完全退出并重开所有 AI 终端（卸载内存中的旧 reporter/hook）。
5. 按第 7 节跑全量验证 + 真实终端矩阵（各 reporter × idle/working/tool/waiting/error/closed）。

## 11. 残留风险（如重启用需继续处理）

- 同用户管道抢占只能缓解：建议未来加 `FILE_FLAG_FIRST_PIPE_INSTANCE` 独占或挑战应答。
- UI 零自动化测试，视觉回归靠人工 soak + GDI 计数。
- Windows 命名管道 `socket.end` 半关闭语义弱：服务端必须在 FIN 到达前回 ACK，客户端靠重试兜底（已验证可行）。
- 本机 node 冷启动约 1.4s（疑似 AV 扫描），statusline/hook 延迟下限受此约束；再优化需走 AV 排除（用户决策）。
- 沙盒豁免提示是 Antigravity 原生策略行为，与上报链无关；默认拒绝、逐条审查命令后再放行。

## 12. 重启用记录 2026-09-24 夜（本文件 §1“全局停用”已过期，以本节为准）

- 状态反转：停用 → **已启用**。生产 exe `AI-Terminal-Pet.exe`（SHA256 `3204889A…`，go1.27 构建，
  源码含本轮注释+`TestStatePriorityOrder`）已安装；`.disabled` 已删；冒烟（拉起→建管→干净退出）通过；
  harness 误触生产管道另拉起一例（PID 4332）并常驻，属正常服役态。
- 三处上报已加回（均先 `.bak-20260924-reenable` 备份）：opencode plugin file URI 行；
  antigravity `statusLine.command` + AfterAgent/BeforeTool 两处 hook；claude settings.json 7 事件 hook。
  **偏差**：claude 7 处全部落在 settings.json，`hooks.json`（ECC 托管）未动；行为等价。
- 基线与细节见 `BASELINE-20260924.md` §9–§11。
- 待办（需用户重开全部 AI 终端后）：真实终端矩阵 3×6；`go test -race`/govulncheck/gosec（缺工具链）；
  harness 加 `launch:false` 彻底隔离生产管道；`Session.Priority` 文档以代码为准
  （waiting6/error5/working4/tool_use3/thinking2/idle1，本轮已修笔误）。
- 2026-09-24 补：Antigravity statusLine 已由 node 版切为原生 `pet-statusline.exe`
  （McAfee 钩挂致 node 冷启动 4.3s/cmd 跳 7.7s，runner 超时击杀；shim 稳态 0.5–0.9s）。
  settings 指向安装路径 ASCII 无引号。代价：状态行显示降级为 `pet <state>` 单 token。
- 2026-09-24 补：修浮窗交互 4 缺陷（主 exe `23C1DA12…`）：hitTest 越界 panic（缩小+边距悬停杀进程）、
  展开/缩放/DPI 不改窗口尺寸（点击无反应）、输入处理同步 present（AppHang 忙光标）、滚轮方向反转；
  另加字体缓存 + ui 首个单测。旧 AppHang 报告见事件查看器（13:50–13:51）；若复发取 procdump -h 栈再定。
- 2026-09-24 补（§16 详述）：出现“窗口不可显示”——进程/线程/管道/配置全正常，
  但 `WS_VISIBLE` 永不清零且跨进程 show 卡死，指向 win32k/DWM 系统级楔子（+McAfee 钩挂嫌疑），
  应用层 show 路径已验无误。首选恢复=重启系统；重启后按 BASELINE §16 清单验证；
  若仍现，升级 procdump 完整 dump + WinDbg。另：主 exe 已迭代至看门狗版（装机 hash `71078FA7…`）。
- 2026-09-24 补：项目已开源（MIT，holder `rongj`，公开仓库 `ai-terminal-pet`）。
  合规结论：依赖均为宽松许可库引用（go-winio MIT / lxn/win BSD-style / x-sys BSD-style），
  无第三方代码复制、无密钥、无二进制进仓（`*.exe` 全忽略）；新增 `LICENSE`/`README.md`/`.gitignore`。
- 回滚：删三处新增条目（或 bak 覆盖恢复）→ `Stop-Process AI-Terminal-Pet` →
  安装目录 exe 改名 `.disabled` → 清 `runtime/` → 回到 §1 停用态。

## 13. 浮窗全线故障根因修复 2026-09-25（本节为最新状态）

**根因（一个）**：`internal/ui/app_windows.go` 的 `Run()` 没有 `runtime.LockOSThread()`。
Go 调度器把主 goroutine 迁到别的 OS 线程后，窗口消息仍投递到「创建窗口的那个线程」的队列，
而该队列无人泵消息；于是首个跨线程 `SetWindowPos(HWND_TOPMOST)` 在同步 `SendMessage` 里
永久阻塞 → `WS_VISIBLE` 从未置位、消息全不分发、定时器全死。

**这一个遗漏解释并已修复**：显示不出来 / 点击崩溃（幽灵窗口 + AppHang）/ 打开异常 /
无法拖拽 / 定时器失效 / 空闲自毁不生效。**§16 的「系统级 win32k/DWM 楔子」结论作废**
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

**完整取证 / 验证矩阵见 `BASELINE-20260925.md`。** 遗留：`-race`/govulncheck/gosec 仍缺工具链。

## 14. 打磨迭代 2026-09-25 夜（v0.2.0）

本轮把浮窗从「能显示」推到「可用」，并补上扩展/轻量化路线图（见 `ROADMAP.md`）。

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

**遗留**：`ROADMAP.md` §6 列出的未修项（`-race`/govulncheck/gosec 缺工具链、UI 无视觉回归、
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
