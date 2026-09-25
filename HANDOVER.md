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
