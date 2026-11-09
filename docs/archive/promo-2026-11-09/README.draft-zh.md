# ai-terminal-pet

不必再逐个终端窗口来回扫。代价只是一个可执行文件，不要账号，也不联网。

![ai-terminal-pet](docs/assets/hero.png)

![platform](https://img.shields.io/badge/platform-Windows-blue) ![Go](https://img.shields.io/badge/Go-1.23-00ADD8) ![version](https://img.shields.io/badge/version-v0.9.0-blue)
![license](https://img.shields.io/badge/license-MIT-green) ![dependencies](https://img.shields.io/badge/dependencies-3-informational) ![tests](https://img.shields.io/badge/tests-140-brightgreen)

## 它解决什么问题

日常使用 AI 编程终端时，很多开发者习惯同时开启两三个会话：一个终端运行 Claude Code 重构模块，另一个终端跑 OpenCode 执行测试，第三个终端用 Antigravity 生成文档。

终端在后台自主运行时，你可能会切换到浏览器查阅技术文档，或者在编辑器中编写业务逻辑。然而，AI 终端随时可能因为调用外部工具或执行脚本而暂停，等待你的权限确认。

这类权限提示不会抢占操作系统焦点，也不会主动发出系统通知。四个终端窗口重叠在桌面后台，某个窗口静静等待确认；当你十多分钟后切回终端时，才发现进度早在十四分钟前就已经停滞。

ai-terminal-pet 在屏幕边缘贴上一条 56 像素宽的状态带。你无需来回切换窗口，扫一眼就能看到哪一个项目正在等待确认、当前已经等待了多久。

所有会话正常执行时，面板仅占用极窄的屏幕边沿，贴边收起后只留一条细边；出现等待确认或报错时，浮窗会自动完全展开并脉动提醒；当全部终端关闭后，浮窗在无会话状态下静默等待 10 分钟便会自行退出，不长期占用桌面与后台资源。

## 长什么样

以下截图均由真实的 Win32 渲染引擎在版式测试中直接光栅化生成，与程序运行时的像素布局完全一致：

![紧凑态](docs/assets/overlay-compact.png)
*紧凑态：展示呼吸脉动的状态球、当前纳管会话总数、最高紧急度的状态词，以及正在等待确认的项目名。*

![展开态](docs/assets/overlay-expanded.png)
*展开态：展示全部会话列表，每一行清晰标明项目名、执行状态、终端来源以及距上次活跃的相对时间。*

![会话很多时](docs/assets/overlay-expanded-many.png)
*会话很多时：超过 6 行后自适应纵向拉长面板，并在列表末尾清晰呈现"还有 N 个会话"溢出行。*

![5 秒否决倒计时](docs/assets/overlay-auto-approve.png)
*5 秒否决倒计时：权限请求行展示"允许（5 秒）"倒计时与渐退进度条，在倒计时结束前点击面板即可立即否决。*

---

## 功能

浮窗采用中性石墨底色与低饱和度状态色，所有界面交互遵循"整块可操作"的设计原则，无需刻意瞄准特定微小控件：

| 功能 | 具体行为 |
|---|---|
| 状态聚合 | 聚合 `idle`、`thinking`、`working`、`tool_use`、`waiting_permission`、`error`、`closed` 七种状态，始终展示所有会话中优先级最高的状态。 |
| 紧凑与展开两态 | 紧凑态固定为 56×156 逻辑像素；点击展开后宽度为 200 像素，自适应展示 1–6 行会话，会话超过 6 行后展示溢出提示。 |
| 整块可点可拖 | 面板任意位置点击切换展开/收起，整块区域均可按住拖拽并与光标 1:1 跟随；滚轮可在 65%–160% 之间无级缩放，按 `Esc` 键直接收起。 |
| 贴边收起机制 | 面板贴靠屏幕边缘时自动滑出屏幕并保留细边，鼠标悬停即滑出复原；有会话处于等待确认或出错状态时保持完全可见，不执行收起。 |
| 终端窗口跳转 | 出现需要处理的会话时，点击紧凑态跳转条或展开态行尾箭头，沿进程树与窗口标题匹配并将对应终端窗口置顶；匹配置信度不足时拒绝误切。 |
| 自动同意与否决 | 开启后为权限请求留出默认 5 秒的否决窗口，面板显示"允许（5 秒）"倒计时条；点击面板立即拒绝，倒计时结束仍无操作则自动同意。 |
| 危险命令黑名单 | 自动同意严格校验命令内容，遇到匹配 `rm -rf`、`rd /s`、`git push --force`、`format`、`del ... \Windows` 等规则的高危命令时强制退回人工确认。 |
| 合成提示音方案 | 内存合成 44.1 kHz 16 位单声道 PCM 音效，提供清脆、钟声、柔和、脉冲四套方案；等待确认、出错、完成三种状态可独立设置声音开关。 |
| 逐会话系统通知 | 进入等待确认或出错状态时弹出系统托盘气泡；气泡与提示音共用会话与状态去重逻辑，点击气泡可快速聚焦浮窗。 |
| 托盘与三页设置 | 托盘菜单包含显示隐藏、重置位置、置顶和语言切换；三页设置窗口（通用、声音、会话）调整参数即刻生效，无需点击确认按钮。 |
| 存活判定 | 终端是否还在，靠操作系统观测进程表和窗口标题，不依赖上报端心跳；静默的会话不会被误判成已关闭。 |
| 单实例运行 | 具名互斥体保证第二次启动只是唤醒已经运行的浮窗，不会再起一个进程。 |
| 连接容量 | 管道服务最多接受 32 路并发连接，读超时 5 秒，IPC 不占用 UI 线程。 |
| 空闲生命周期 | 宿主连续 2 分钟不可观测才移除会话；一个会话都不剩时，进程在 10 分钟后自行退出。 |

## 它是怎么工作的

宿主终端与浮窗通过本地 Windows 命名管道进行进程间通信，整体架构如下：

```
AI 编码终端 (Claude Code / OpenCode / Antigravity)
       │
       ▼ (快速失败的上报脚本)
 reporters/*.js ──[命名管道 / JSON-lines]──► AI-Terminal-Pet.exe (Go)
                                                    │
                                                    ├─► 命名管道服务 (32 连接, 5 秒读超时)
                                                    ├─► 会话状态存储 (上限 256 会话)
                                                    ├─► 磁盘快照回退 (%LOCALAPPDATA%)
                                                    └─► Win32 分层浮窗 (GDI 置顶画布)
```

系统核心设计遵循以下原则：

- **按用户隔离的管道路径**：管道名为 `\\.\pipe\ai-terminal-pet-v1-<8hex>`，后缀为当前用户主目录 SHA-256 哈希的前 8 位十六进制字符。Go 端主程序与 JS 上报端各自独立实现哈希计算逻辑，通过跨语言测试严格钉住契约。
- **单行 JSON 与精确回执**：通信采用每行一个独立 JSON 对象（协议版本号 `ai-terminal-pet.v1`），单条数据载荷上限 1 MiB。服务端处理完毕后精确回复 `OK` 或 `ERR`，客户端未收到完整确认前绝不视作成功。
- **系统级存活性观测模型**：会话是否存在由操作系统直接观测决定（扫描系统进程表与终端窗口标题），而非依赖上报端的心跳包。即使终端长时间无状态更新，只要进程或窗口仍在就不会被误判关闭；只有宿主连续 2 分钟在系统中不可观测时才会判定会话结束。
- **权威磁盘证据优先于心跳**：存储在 `%LOCALAPPDATA%\AI-Terminal-Pet\runtime\` 下的快照文件既是管道断开时的回退通道，也是会话存在的落盘证据。磁盘证据的判定优先级高于易失的心跳包，已落盘快照消失即代表该会话正式终结。
- **单实例互斥与按需拉起**：主程序通过具名互斥体保证桌面全局单实例。上报端在发现管道不存在时可按需拉起浮窗，二次启动命令行则会激活已存在的窗口。
- **脏帧跳过与按状态控频**：光栅化模块对实际绘制字段进行哈希签名，前后两帧签名相同时直接跳过底层 GDI 重绘；脉冲动画与滑动动画按需装卸定时器，空闲时完全停止重绘。
- **纯 GDI 内存画布**：用 CreateDIBSection 建立自顶向下的 32bpp DIB，通过内存 DC 绘制，再交给 UpdateLayeredWindow 贴到分层窗口。

## 装起来

本项目仅支持 Windows 平台。打开 PowerShell，按步骤执行构建与安装：

```powershell
# 1. 编译主程序与原生状态栏 shim
go build -trimpath -ldflags "-s -w" -o .\bin\AI-Terminal-Pet.exe ./cmd/ai-terminal-pet
go build -trimpath -ldflags "-s -w" -o .\bin\pet-statusline.exe ./cmd/pet-statusline

# 2. 安装到系统本地目录并做哈希校验
.\install.ps1

# 3. 验证安装输出
ai-terminal-pet --json
```

`install.ps1` 在安装时会自动停止正在运行的浮窗，等待系统进程释放文件句柄，并在拷贝后对主程序、状态栏 shim、协议脚本和插件包逐一执行 SHA256 哈希校验。

安装完成后浮窗与状态栏 shim 已部署就绪，但终端默认不会自动推送事件；你必须按照下一节的说明，为正在使用的终端环境配置对应的上报端。

## 接入你的终端

上报端脚本均存放在 `reporters/` 目录下，基于 Node.js 标准库编写，零 npm 第三方依赖，通过 `protocol.js` 统一封装通信协议。

各个终端的具体接入方式如下表所示：

| 终端 | 怎么接 | 得到什么 |
|---|---|---|
| OpenCode | 在 `opencode.json` 的 `plugin` 列表中加入 `file:///C:/path/to/ai-terminal-pet/reporters/opencode.js`。修改配置后重启 OpenCode 即可生效。 | 执行中、工具调用、思考中等状态实时上报。 |
| Claude Code | 在 `settings.json` 的 hooks 配置中，为 `SessionStart`、`PreToolUse`、`Stop`、`PermissionRequest` 等挂载 `node <repo>/reporters/claude-hook.js`。 | 权限确认拦截、工具执行状态与 5 秒否决倒计时。 |
| Antigravity | 将 `statusLine.command` 指向 `pet-statusline.exe`，并在事件配置中注册 `antigravity-hook.js`。 | 亚秒级状态栏响应、Agent 思考与权限确认聚合。 |
| DeepSeek Harness | 在官方插件管理器中运行 `install_bundle`，传入 `%LOCALAPPDATA%\AI-Terminal-Pet\bundles\ai-terminal-pet-bridge\` 的绝对路径。 | 真实事件映射：turn/start→working、tool/call→tool_use、approval/asked→waiting_permission、tool/result 出错→error、turn/end→idle、进程退出→closed。 |
| 其他任意 CLI | 在状态变化时调用 `reporters/generic.js`。支持通过命令行参数或标准输入管道传递状态。 | 一行脚本接入任意外部 CLI 或自定义脚本任务。 |

Harness 那一侧是 host-only 的 Cordis bundle，自身不含管道名、状态集与快照逻辑，它从 `%LOCALAPPDATA%\AI-Terminal-Pet\protocol.js` 加载协议，全部委托给官方 protocol.js。

针对 Antigravity CLI，推荐将状态栏命令直接指向编译好的原生二进制 `pet-statusline.exe`。在开启杀毒软件的机器上，原生 shim 能够消除频繁启动 Node 运行时的进程创建延迟。

对于表格中未列出的其他终端或自动化脚本，可以使用 `reporters/generic.js` 进行一行接入：

```powershell
node reporters/generic.js --source=aider --state=working --project="$PWD"
node reporters/generic.js --source=codex --state=waiting_permission --reason=approve
echo '{"state":"error","reason":"build failed"}' | node reporters/generic.js --source=ci --stdin
```

所有上报端均遵循快速失败机制，如果浮窗未启动或管道繁忙，脚本会静默退出，绝不向宿主终端抛出未捕获异常。修改终端配置后，请重启对应的终端以加载新的上报配置。

## 用脚本控制它

浮窗提供完整的命令行控制参数，各命令行为如下：

| 命令 | 作用 |
|---|---|
| `ai-terminal-pet` | 启动浮窗（若已在运行则激活已有实例） |
| `ai-terminal-pet --show` | 唤出并显示正在运行的浮窗 |
| `ai-terminal-pet --toggle` | 在紧凑态与展开态之间切换 |
| `ai-terminal-pet --reset` | 重置位置并以 100% 缩放重新贴靠屏幕边缘 |
| `ai-terminal-pet --settings` | 打开三页设置窗口 |
| `ai-terminal-pet --quit` | 退出正在运行的浮窗进程 |
| `ai-terminal-pet --json` | 打印当前会话状态 JSON（可加 `--pretty`） |
| `ai-terminal-pet --autostart` | 查询当前的开机自启状态 |
| `ai-terminal-pet --autostart=on` | 设置为登录时自动启动 |
| `ai-terminal-pet --autostart=off` | 取消开机自启设置 |
| `ai-terminal-pet --version` | 打印版本信息（v0.9.0） |
| `ai-terminal-pet --help` | 输出命令行帮助信息 |

通过 `--json` 参数读取当前状态时，返回的会话数据结构示例如下：

```json
{"schemaVersion": "ai-terminal-pet.v1",
 "state": "waiting_permission",
 "count": 1,
 "generatedAt": "2026-11-09T20:30:00Z",
 "sessions": [{"source": "claude-hook", "project": "C:\\Users\\me\\work\\my-service", "state": "waiting_permission"}]
}
```

`--json` 命令直接读取本地磁盘上的快照文件，不与 GUI 进程进行任何同步通信，因此无论浮窗是否运行都能立刻返回，绝不会因界面卡死而阻塞。

系统的状态持久化分为两份文件：
- `%LOCALAPPDATA%\AI-Terminal-Pet\config.json`：存放用户偏好设置，包括通知开关、音效配置、置顶状态、贴边收起、不透明度、界面语言以及自动同意开关等。文件在浮窗启动时读取，修改设置时由后台协程异步合并写入；缺失或格式错误项会自动回退到默认值，损坏的文件绝不阻碍程序启动。
- `%LOCALAPPDATA%\AI-Terminal-Pet\state.json`：存放浮窗的屏幕坐标、当前缩放比例以及是否处于展开态。浮窗移动或缩放后自动保存，重启后原样恢复窗口几何位置。

开机自启配置独立于上述两份文件，直接通过 Windows 注册表 `HKCU\Software\Microsoft\Windows\CurrentVersion\Run` 管理，无需管理员权限，卸载时清理无残留。

## 安全边界

上报端与浮窗的设计以不干扰宿主终端为第一原则，界定了明确的安全边界：

- **快速失败与异常隔离**：上报端脚本遵循快速失败原则，任何网络异常、解析错误或超时均静默返回，绝不把异常抛回宿主 CLI，确保宿主终端正常执行。
- **自动同意默认关闭**：自动同意功能默认处于关闭状态。开启后仅在收到权限请求时触发，且普通工具调用不会被拖慢。若上报端无法读取配置文件，则自动回退至关闭状态。
- **5 秒否决窗口语义**：权限请求到达后开启 5 秒倒计时，面板展示递减进度条。在倒计时期间点击面板立即拒绝执行，5 秒内无任何操作才会自动同意。
- **危险命令黑名单**：自动同意内置黑名单检测，对匹配高危模式的命令一律拒绝自动放行并转交人工确认，包括 `rm -rf`（递归删除）、`rd /s`（Windows 目录递归删除）、`git push --force`（强制推送）、`format C:`（磁盘格式化）以及 `Remove-Item ... System32`（系统核心路径删除）。
- **零遥测与本地存储**：无任何遥测代码，不进行任何网络请求，全部运行配置与会话状态均保存在本地 `%LOCALAPPDATA%\AI-Terminal-Pet\` 目录下。
- **命名管道按用户隔离**：Windows 命名管道基于当前用户主目录哈希隔离，管道 DACL 为 `"D:P(A;;GRGW;;;OW)(A;;GRGW;;;SY)"`，只授予管道属主（运行浮窗的那个 Windows 用户）与本机 LocalSystem 读写权限，本机其他账户既不能连接也不能抢占这个管道名。

## 已知限制

- **仅限 Windows**：仅基于 Win32 分层窗口与 Windows 命名管道构建，不支持 macOS 或 Linux。
- **杀毒软件增加进程启动开销**：在部分开启杀毒软件钩子的机器上，频繁启动 Node 进程会导致可感知的延迟；建议使用原生状态栏 shim 并为安装目录配置杀毒排除项。
- **同用户管道抢占未根除**：通过互斥体优先排序缓解了同用户命名管道被抢占的风险，但尚未实现基于独占标志的彻底根除。
- **倒计时时长修改需重启**：自动同意倒计时时长在程序启动时一次性载入，修改该配置项需要重启浮窗进程方可生效，其余配置项均支持热加载。
- **缺少像素级视觉回归测试**：纯逻辑、几何计算与命中测试具备完整的单元测试覆盖，并提供了 PNG 预览渲染，但目前尚未接入像素级别的自动化视觉回归测试。
- **终端标签页定位限制**：Windows Terminal 标签页支持沿进程树精准识别与置顶，但经典 conhost 多会话窗口与 VS Code 内置终端目前仅能定位并跳转到主窗口层级。

## 路线图

以下为当前计划进行的改进项：

- **上报端原生化**：计划开发统一的原生二进制 `pet-hook.exe`（支持 `--mode=claude|antigravity`），替代每次事件拉起 Node 运行时的开销。
- **协议规范文档**：计划补充完善 `docs/PROTOCOL.md`，使第三方开发者无需查阅源码即可编写合规的上报端。
- **跨平台抽象探索**：计划将窗口渲染层抽象为通用接口，为后续探索基于 Unix 域套接字的平台移植奠定结构基础。
- **自检诊断工具**：计划增加 `--doctor` 命令，用于自动排查互斥体状态、窗口可见性、管道连通性与配置异常。
- **二进制体积压缩**：计划将主程序体积从现状 3.7 MiB 降至 2.5 MiB 以内，两个可执行文件合计控制在 5 MiB 以内。手段包括剥离无用反射元数据、去掉未用包，以及把 pet-statusline 里基于 regexp 的脱敏规则改成手写扫描器。

## 非目标（明确不做）

- 不做账号体系、不做云同步、不做遥测，数据只落在 `%LOCALAPPDATA%`。
- 不是终端复用器，也不替代你的终端，它只回答“哪个会话在等你”。

## 许可证

本项目基于 [MIT](LICENSE) 许可证开源。仅以库形式引入以下三个第三方模块，未内嵌任何第三方源码：

- `github.com/Microsoft/go-winio`（MIT 许可证）
- `github.com/lxn/win`（BSD 风格许可证）
- `golang.org/x/sys`（BSD 风格许可证）

## 从源码开发

如果你需要修改代码或向本项目提交贡献，可以在修改后运行以下命令执行静态分析、单元测试、上报端脚本语法检查、Harness 适配校验以及 UI 布局预览：

```powershell
# 静态分析与全量单测
go vet ./...
go test -count=1 ./...

# 上报端脚本语法与适配校验
node --check reporters/*.js
node integrations/deepseek-harness/verify.mjs

# UI 布局光栅化预览（在临时目录生成 PNG）
$env:PET_PREVIEW_DIR = "$env:TEMP\petpreview"
go test ./internal/ui -run TestRenderPreview
```
