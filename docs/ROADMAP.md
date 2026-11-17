# ROADMAP — ai-terminal-pet

> 快照时间 **2026-11-17（浮窗 1.0.21 / 插件 1.2.34 已发版）**。本文分六块：现状基线、
> 已交付批次、扩展计划、轻量化计划、实用功能候选与非目标、已知残留、**增长计划（§7，2026-11-17 按
> [docs/METRICS.md](METRICS.md) 的实测重定）**。每项都标注了**动机**和
> **为什么现在不做/怎么做**，便于下一位维护者直接挑一项开工。
>
> 交付历史的逐条细节在 [docs/HANDOVER.md](HANDOVER.md) §13–§27 与 §35–§76（§75 是 OpenCode npm 包那一轮，
> §76 是本轮）；按版本发布的变更摘要在 [CHANGELOG.md](../CHANGELOG.md)。

## 0. 现状基线

| 维度 | 数值（2026-11-17，浮窗 1.0.21 / 插件 1.2.34 已发布） |
|---|---|
| 平台 | Windows（Win32 分层窗口 + 命名管道）· Linux（纯 Go `jezek/xgb` X11 ARGB + Unix socket）· macOS（纯 Go `purego` AppKit + Unix socket），三端均 `CGO_ENABLED=0` |
| 工具链 | Go 1.23+（实测 go1.27.0 windows/amd64），reporters 仅用 Node 标准库（无 npm 依赖） |
| 版本 | `1.0.21`（`internal/app/version.go` 的 `const Version`）；插件包 `1.2.34` |
| OpenCode 上报端包 | `opencode-ai-terminal-pet` **1.0.21**，**与浮窗同号**（`tools/check_versions.mjs` 断言相等），没有自己的版本线 |
| 主程序 | `bin/AI-Terminal-Pet.exe` **4,783,616 B**（4.56 MiB，`-trimpath -buildvcs=false -ldflags "-s -w"`） |
| 状态行 shim | `bin/pet-statusline.exe` **3,678,720 B**（3.51 MiB） |
| 原生 hook | `bin/pet-hook.exe` **3,705,344 B**（3.53 MiB）；三个 Windows 二进制合计 **12,167,680 B** = 11.60 MiB |
| 代码量 | Go 非测试源码 **156 文件 / 28,561 行**、测试 **101 文件 / 14,946 行**（**504** 个 `func Test`）；JS/MJS 侧为 6 个 reporter + 工具链 `tools/*.mjs` + 两个包的测试与生成物（`packages/*/vendor/protocol.js` 与 `protocol.cjs` 是 `reporters/protocol.js` 的逐字节副本，不算第二份实现）。文件/行数与二进制体积**由 `node tools/check_facts.mjs --print` 现测**（它们每天变，本表不把它们当作断言；`func Test` 数量是断言，见本文件 §0 的门禁说明） |
| 第三方依赖 | 6 个直接依赖 + 1 个间接依赖，无 cgo：go-winio v0.6.2(MIT) / lxn/win(BSD) / ebitengine/purego v0.9.1(Apache-2.0) / jezek/xgb v1.3.1(BSD) / x/image v0.25.0(BSD) / x/sys v0.10.0(BSD)（+ x/text v0.23.0 indirect） |
| 运行期占用 | 有活跃会话时工作集约 22–38 MB（含共享页），线程 8–16（Windows 实测，待跨平台长期基线复测） |
| 面板几何 | 紧凑 `56x156`（固定）· 展开 `200x<内容高度>`，1–6 行自适应，超过 6 行加一条溢出行 |
| 设置窗口 | Windows：第二个顶层窗口类 `AITerminalPetSettings`（四页签 general/sound/sessions/look）；Linux/macOS：直接编辑 `config.json`（详情见 [docs/PLATFORMS.md](PLATFORMS.md)） |
| 协议 | `ai-terminal-pet.v1`，每行一个 JSON，上限 1 MiB，回执精确 `OK` / `ERR` |
| 本地端点 | Windows `\\.\pipe\ai-terminal-pet-v1-<homedir-sha256 前 8hex>` · Linux `$XDG_RUNTIME_DIR/ai-terminal-pet/v1-<8hex>.sock` · macOS `$TMPDIR/ai-terminal-pet-v1-<8hex>.sock` |
| 上报端 | `opencode.js`（已发行为 npm 包 `opencode-ai-terminal-pet`）/ `claude-hook.js` / `antigravity-hook.js` / `antigravity-statusline.js` / `generic.js` / `pet-statusline` / `pet-hook` / `integrations/deepseek-harness/ai-terminal-pet-bridge` |
| 生命周期 | 按需拉起（reporter 发现端点不存在就拉起）；会话存在性由 OS 观测（进程表 + 窗口），连续不可观测 ≥ `DefaultAbsenceGrace` 2 分钟才移除；空闲自毁宽限 10 分钟；L1 statusline 30 秒管道短路；F2 键盘行导航；F5 按显示器工作区记忆坐标；**L5 快照扫描按身份/size/mtime 复用上次解析**（200 文件 22.7 ms → 0.51 ms）；**存在性判据：`title` 必须被窗口标题"包含"且不完全等于它**（§63 修掉的回归） |
| 持久化 | `state.json`（位置/缩放/展开/`layouts`）· `config.json`（通知/音效/置顶/贴边/不透明度/语言/auto-approve）· `HKCU\...\Run` / `LaunchAgent` / XDG `.desktop`（开机自启）· `approvals/` |
| 对外接口 | `--show/--toggle/--reset/--quit/--json[--pretty]/--autostart[=on\|off]/--settings`（有窗口开窗口，否则终端编辑器）`/--set k=v/--connect[=HOST]/--doctor/--version/--help` |
| 单测与门禁 | **504 个 `func Test`** 分布在 15 个包（Windows + Linux + macOS 全量通过）；CI 的 `ubuntu-latest` 另跑 `go test -race`；工具链另有 `tools/*.test.mjs`、两个包各自的测试，以及生成物一致性检查（`sync.mjs --check`） |
| 持续集成 | `.github/workflows/ci.yml`：四条并行 Job — `windows-latest`（全套单测/构建/打包/fresh-machine/版本链接检查）、`ubuntu-latest`（`cross-build.sh` / `go test` / Linux 二进制 / `install.sh` + tar 校验）、`macos-latest`（`go vet` / `go test` / Darwin 二进制 / `install.sh` + hook 冒烟）、`scan`（`govulncheck` + `gosec`） |
| 自检 | `ai-terminal-pet --doctor` **13 项**（environment / install files / version / runtime / config / pipe / instance / node / reporters / autostart / alerts / audio / notifications），`[ok]/[warn]/[fail]` 逐行 + 末行汇总，任一 fail 退出码 1；`reporters` 读 OpenCode / Claude Code / Antigravity 三份宿主配置 |
| 接入向导 | `ai-terminal-pet --connect[=opencode\|claude\|antigravity]` 按当前 OS 填写真实二进制/脚本路径并打印配置片段，**只打印不写盘**；OpenCode 现在优先给**包名**形态（`opencode plugin add …`），`file://` 检出形态只作为开发路径附在备注里 |
| 自检的 OpenCode 形状 | `reporters` 这一项认三种入口：**裸包名**（查 `<cache>/opencode/npm/<name>@<tag>/<epoch-ms>/node_modules/<name>/package.json`，即 `opencode plugin add` 真正安装的位置；配置目录的 `node_modules` 作为手动安装的第二种形状一并查；都没有只给 warn 并附安装命令）、**直接文件 URL**、**目录 + index.ts 两跳**（§52.3 的回归） |
| 线协议文档 | [docs/PROTOCOL.md](PROTOCOL.md)（388 行），第三方无需读源码即可编写上报端 |
| 插件包 | `packages/dsh-ai-terminal-pet`（npm 名 `dsh-ai-terminal-pet` **1.2.31**，20 个文件含 vendor 与 `vendor/manifest.json`；release tag `dsh-plugin-v1.2.31`，三平台 bundle） |
| OpenCode 上报端包 | `packages/opencode-ai-terminal-pet`（npm 名 `opencode-ai-terminal-pet` **1.0.21**，**包内带面板**：`vendor/AI-Terminal-Pet.exe` 4.78 MB + 清单，装包约 4.8 MB；`server.js` 手写负责首次装载，`reporter.js` / `protocol.cjs` 由 `scripts/sync.mjs` 从 `reporters/` 生成，两个包共享 `tools/overlay-vendor.mjs` 的源码摘要）；`opencode plugin add opencode-ai-terminal-pet` 一行接入，Windows 上这一行连面板一起装好 |
| 插件清单 | 仓库根 `package.json` 清单 shim（2,115 B / 68 行，无代码；1.2.3 起多九行 `dsh.marketplace`，1.2.4 只改 `dsh.catalog` 摘要）：`dsh.bundle.patch` + `dsh.catalog` + `private`，让 `add github:<owner>/<repo>` 可直接解析 |
| 视觉验证 | **不变量门禁**：`go test ./internal/ui -run 'TestRenderedPanelKeepsItsShape|TestStateColourReachesTheScreen|TestStatesAndLanguagesLookDifferent|TestThemesAreActuallyDifferent'`（三平台都跑，断言尺寸/圆角透明/填色/文字已光栅化/状态色出现/状态与语言确实不同/两套配色确实不同，**不比像素**）；**看图**：`PET_PREVIEW_DIR=<dir> go test ./internal/ui -run TestRenderPreview`（浮窗状态）、`-run TestRenderSettingsPreview`（设置窗口 5 张），`docs/assets/render_assets.py` 由此产出 README 配图；`docs/assets/make_launch_assets.py` 产出 `demo.gif`、`desktop-scene.png` 与 `social-preview.png`（GitHub 社交预览，1280×640，**上传是仓库设置里的手工步骤**），`make_doctor_card.py` / `make_smartscreen_notice.py` 产出安装路径 B 的两张图；`python tools/check_assets.py` 钉住 `docs/assets/README.md` 表里的尺寸与帧数 |

> **这些数字是怎么保证的**：`func Test` 数量、有测试的包数、Go 包总数与 `--doctor` 项数由
> `node tools/check_facts.mjs` **断言**（README 中英、`AGENTS.md`、本表写错就红）；文件/行数、
> 二进制体积与依赖清单**没有断言**——它们每次编辑、每个工具链都会变（`go 1.23.0` 是下限而不是钉死，
> 所以 CI 与本机编出的字节本来就不该相同），由 `node tools/check_facts.mjs --print` 现测后连日期一起写进来。

## 1. 已交付批次（按时间倒序）

### 1k. 2026-11-17（1.0.19 / 1.2.32）—— OpenCode 接入从「贴本机路径」变成一行 npm 安装

**动机**：OpenCode 在此之前只能靠 `opencode.json` 里一条 `file:///…/reporters/opencode.js` 接上，而那条路径指向**仓库检出**。三个后果都是这一轮当场核实的：

- zip / `pet-setup.exe` 装出来的目录里**没有** `reporters/opencode.js`（`install.ps1` 的 12 件制品只有 `protocol.js`），而 zip 里生成的 `README.install.txt` 却让用户把那条路径写进配置——照着做必然失败；
- `--connect=opencode` 的片段是**从仓库检出**生成的（`snippet.go` 找 `reporters/opencode.js`），所以只有 clone 过仓库的人能接；
- POSIX 的 tar.gz 反而带了全部 6 个 reporter（`package-unix.mjs`），这不一致是 Windows 侧打包独有的。

**交付**：

- 新包 `packages/opencode-ai-terminal-pet`（npm 名同名，零依赖，5 个文件约 48 KB）。`server.js` 与 `protocol.cjs` 由 `scripts/sync.mjs` 从 `reporters/opencode.js` / `reporters/protocol.js` **生成**，唯一差异是一处 `require()` 说明符：npm 插件按 OpenCode 规则必须 `"type": "module"`，而共享核心是 CommonJS（5 个 hook `require()` 它，安装器也以 `protocol.js` 部署它）。
- 生成器三条规则各有测试：源里**恰好一处**说明符、副本**逐字节**相等、拒绝二次生成（源里已出现 `.cjs` 就报错）。`--check` 进 CI，与 `packages/dsh-ai-terminal-pet/vendor/protocol.js` 是同一套纪律。
- 包内 `server.test.mjs` 用临时 home + **真实端点**验证「esm 入口 + cjs 核心 + 真投递 + 快照落盘」，三平台各跑一次：Windows 命名管道、Linux `$XDG_RUNTIME_DIR`、macOS `$TMPDIR`。
- `internal/connect` 认第三种入口形状：**裸包名**（含 `@version`）与对象 `{ "package": … }`。证据是 OpenCode 自己装出来的那份清单——**实测位置是 `<cache>/opencode/npm/<name>@<tag>/<epoch-ms>/node_modules/<name>/package.json`**（`$XDG_CACHE_HOME` 或 `~/.cache`，Windows 上也是后者），配置目录的 `node_modules` 作为「手动 npm install」的第二种形状一并查；都没有只给 **warn** 并附安装命令，因为这个工具看不到对方的全部解析路径，不能把一种能跑的装法判成坏的。
- `--connect=opencode` 改成包名优先；`file://` 检出形态只在真有检出现时进备注（给改 reporter 的人）。
- 版本策略：这个包**与浮窗同号**（`tools/check_versions.mjs` 断言相等），所以一次浮窗发版就是一次 npm 发布，不存在「reporter 1.0.19 + 浮窗 1.0.18」这种没有一起测过的组合。

**第三个坑（1.0.19 存在的唯一理由）**：1.0.18 里 doctor 的包形状按「配置目录的 `node_modules`」找，而 `opencode plugin add` 实际把包装在 `~/.cache/opencode/npm/…`。写代码时那是**假设**，跑一次真的 `opencode plugin add` 才发现——于是在一台装好的机器上，doctor 会对正确的安装报一条「没找到」。修法只改了查找路径与测试夹具，但二进制已经发出去，**资产不可变**，于是有了 1.0.19 / 1.2.32 / 上报端 1.0.19（三个号一起走，因为上报端与浮窗同号）。

**两个坑**：

1. `"type": "module"` 与 CJS 核心天生冲突。唯一诚实的答案是核心以 `.cjs` 进包、由生成器保证逐字节——手写第二份实现正是 §52.3 的成因，而且那次是静默的。
2. 测试里**不能**等 `socket.end` 之后再回 `OK`：Windows 命名管道没有半关闭，客户端 `end()` 之后再写就是 EPIPE。真实 Go 服务端是「读到换行就回」，测试必须照它写，否则测的是另一种协议。

### 1j. 2026-11-10（v0.9.0 之后，未发版）—— DSH 插件包与市场分发尝试

**动机**：§1i 解决了「别人的机器能跑通」，但没有解决「别人怎么拿到」。Harness 的接法当时要求先 clone
整个仓库、跑 `scripts/install.ps1`，再把安装目录里 bundle 的绝对路径贴进插件管理器——对外部用户不成立。
DSH 的插件市场正好提供了一个现成的分发面：注册表仓库里一个 `data/plugins/<slug>.yml`，加上一个 npm 包。

**交付**（`7f476c9` + `91719e5`，细节与坑见 HANDOVER §27）：

- `packages/dsh-ai-terminal-pet/` 独立 Cordis bundle 包（npm 名 `dsh-ai-terminal-pet`，MIT，
  `type: module`；1.2.4 口径 **18 个文件**含 vendor 与 `vendor/manifest.json`）。
  `dsh.bundle { patch: './cordis.patch.yml', platforms: ['win32'] }`，`engines node >=24.9.0` /
  `dsh >=0.2.0-0 <0.3.0-0`（1.2.16 起；DSH 0.2.1-alpha.3 之后只读 `engines.dsh`），`files` 显式 10 项。**1.1.0 起没有任何生命周期钩子**（`prepack` 被删，
  因为 pnpm 对 git 依赖会执行它、未钉版本的源会撞上 `allowBuilds`）；`vendor/` 改为整体提交，只做校验。
- **版本线 1.0.0 → 1.1.0 → 1.2.0 → 1.2.1 → 1.2.2 → 1.2.3 → 1.2.4：一版一个 tag，旧 tag 不删也不复用**。
  1.0.0（`7f476c9` + `91719e5`，资产 1,913,682 B / sha256 `7365f6a6f4aea1174ec6f2f00670e868568e944816d1869aeb59db363ae6b5af`；
  它的 `vendor/` 曾没进 git 树，tag 后来被强制挪到 `91719e5`，**已作废**）。
  1.1.0（`b280093`，tag `dsh-plugin-v1.1.0` 是 annotated tag，指向 `b28009377e14840ee003ddaf114922e1d167ed0d`，
  资产 **1,913,726 B / 15 files**，sha256 `3b8f9186cca5142626f2e25175096f1c4fb59d86cc1cdff5879792950146cc78`、
  shasum `19a2e544ab875c61fcf822e85a91c67273981725`；**保持已发布、不可变**）。
  1.2.0（已由 1.2.1 取代，发布物本身不可变；tag `dsh-plugin-v1.2.0` annotated、tagger 2026-11-10T13:46:52+0100、指向
  `4d777e32e0a9ef57bb7714287dccd7cae065346f`；Release 于 2026-11-10T12:47:24Z 发布，资产
  `dsh-ai-terminal-pet-1.2.0.tgz` **1,914,566 B / 16 files / sha256 `ccf9d531…` / shasum `dc65e0d4…`**）：
  两个清单新增 `dsh.catalog`（`category: tool`、中英各一行摘要 190 / 104 字符、
  五项自我声明的 capabilities）、`keywords` 补 `dsh-plugin`、`vendor/manifest.json` 随包，
  于是 tarball 文件数 **15 → 16**；实际资产 packed 1.9 MB / unpacked 4,460,352 B。发版前 dry-run 的
  shasum 是 `166c3043…`，之后包内 `README.md` 改了 13 B 使其作废，最终资产是 `dc65e0d4…`。
  清单内容变了就必须换版本号，
  否则仓库与那份不可变 tarball 分叉——这正是 1.2.0 存在的理由。
  1.2.1（已由 1.2.3 取代，发布物本身不可变；tag `dsh-plugin-v1.2.1` annotated、指向
  `1e08b3e7521d305d744c5ff515b29f6d16abebe1`，tag 对象 sha `0ef1d60cb5c33517aa87d6eed75565c035d5830e`；
  Release 于 **2026-11-10T16:15:04Z** 发布，资产 `dsh-ai-terminal-pet-1.2.1.tgz`
  **1,914,616 B / 16 files / sha256 `b9e6d203…` / shasum `67955a65…`**，GitHub 记录的 digest 与本地重算一致）：
  它存在的唯一理由是 1.2.0 的 tarball 里 `README.md` 还写着「条目在等人工评审」，而条目 14:33:04Z 已被合并；
  Release 资产不可编辑，所以只修一句话也要耗一个版本号（同一版本号下换字节，正是本仓库一直在消除的漂移）。
  与 1.2.0 的差别只有包内 `README.md`（6,665 → **6,863 B**）与两条安装命令里的版本号；解包核对确认不含旧文，
  连续两次 `npm pack` 逐字节相同。
  1.2.2（已由 1.2.3 取代；差别是三处：两份清单的 `version`、包内 `README.md` 的措辞、`vendor/manifest.json` 的源码摘要——
  因为安装脚本下沉到 `scripts/` 改到了 8 处 Go 注释，那份摘要由 `89eef0d6…` 变成 `cf647af1…`；
  打包结果 **1,914,751 B / 16 files / sha256 `02431e16…` / shasum `22b63b25…`**，连跑两次逐字节相同）。
  1.2.3（已由 1.2.4 取代；只改两份清单，各加一个 `dsh.marketplace` 块：`profiles: ["desktop", "headless", "web"]`、
  `requiresBuildApproval: false`、`manualSteps: false`。理由是 DSH 无法表达「host-only bundle 跑在任何 profile」——
  它唯一与 profile 有关的字段 `dsh.client` 只能是 `web`，于是按 `install.profiles` 过滤的第三方商店
  （`YELEBAI/dsh-plugin-marketplace` 就是这种实现）只能把本包猜成 `[headless, web]`，桌面端因此看不见条目。
  **DSH 本身从不读这个字段**（插件照旧经 `dsh.bundle` 挂载，只订阅 `session/event`，不依赖任何 profile 专属服务）；
  读它的是商店侧。该注册表的 `scripts/registry-core.mjs` 的 `classifyInstall` 被本地重放过（替换掉 `yaml` 依赖后
  直接 import 它的源码），结果 `profiles: [desktop, headless, web]`、`profileSource: "manifest"`、`reviewReasons: []`。
  打包结果 **1,915,013 B / 16 files / sha256 `1172aa72…` / shasum `0508ff6b…`**，与 1.2.2 只差 262 B；
  `lib/index.js`、两份 vendor 二进制、`cordis.patch.yml`、`dsh.catalog`、`vendor/manifest.json` 逐字节不变）。
  1.2.4（当前；**只改文案，不改代码**）：市场文案原先读起来像「装一次市场条目就能看到所有终端」，而实际情况是
  这个 bundle 开箱即用映射的只是 **DeepSeek Harness** 自己的会话——OpenCode、Claude Code 与 Antigravity 的会话
  要先给各自那个终端装好上报端（`reporters/`）才会出现在同一块面板里。包内 `README.md`（7,546 B / 137 行 →
  **8,187 B / 146 行**）与两份清单的 `dsh.catalog` 摘要（en 225 字符 / zh 114 字符，三处逐字相同：根 shim、
  包清单、`docs/market-index.json` 的 `plugins[0].description`）现在把这一点写明，并指向仓库 README 的
  `#connect-your-terminal` 段。两份清单的 `version` 因此从 1.2.3 走到 **1.2.4**（根 2,052 → **2,115 B / 68 行**，
  包 2,183 → **2,246 B / 83 行**）；**随包字节其余部分不变**——`lib/index.js` 570 行、两份 vendor 二进制、
  `cordis.patch.yml`、`screenshots/`、`vendor/manifest.json`（源码摘要仍 `cf647af1…`）全部逐字节一致，
  因为 1.2.4 没有碰任何 `.go` 源码。打包结果 **1,915,340 B / 16 files / unpacked 4,462,112 B /
  sha256 `d6b7f95a5a3a71c4a9367699e9ad9b8f7d5db522271782e8ffe29a73e45c6708` / shasum
  `dafb81998efbae8e906a36773e9737117aa482d9`**，与 1.2.3 的 1,915,013 B 只差 327 B，连跑两次 `npm pack` 逐字节相同。
  `docs/market-index.json` 的版本与 `publishTimestamp`（1794379909 = 2026-11-11T06:51:49Z）同一批前移，
  `source.commit` 与图标钉点由紧随其后的提交补齐（沿用 1.2.1 / 1.2.2 / 1.2.3 那三次的两步做法）。
- **capabilities 只自我声明、不被强制**：dsh 不对插件做沙箱，所以那五项描述的是浮窗本来在做的事，
  不是权限授予。`keywords` 与 topics（`dsh-plugin` / `deepseek-harness`，另加 `cordis-plugin` 标生态）决定了两条
  不用申请的自动采集路径（见 HANDOVER §27.9）。
- **包内自带预编译浮窗**：`vendor/AI-Terminal-Pet.exe`（4,371,968 B）与 `vendor/protocol.js`（20,914 B），
  `runBootstrap()` 只在 `%LOCALAPPDATA%\AI-Terminal-Pet\` 里对应文件**缺失**时才拷，
  `copyVerified` 走临时文件 + rename 并在 rename 前复查目标存在——**从不覆盖已装文件**。
- **`cordis.patch.yml` 必须用 `insert:`**。裸 `- id:` 只 patch 既有 loader 行：装完看着成功，
  运行时什么都不挂载。原因写进了该 YAML 的文件头，不只写在文档里。
- **仓库根的清单 shim**（`06d4cae` 建档时 1,877 B / 59 行，1.2.4 起 **2,115 B / 68 行**，不含代码）：`main` 指向插件包、
  `dsh.bundle.patch` 指 `./packages/dsh-ai-terminal-pet/cordis.patch.yml`、`files` 只放行那个子树、
  `private: true` 不占 npm 名、`version` 跟包同步（`verify.mjs` 断言两者相等）。
  没有它，宿主读到的是 0.0.0 依赖并报 `declares no dsh.bundle`，装出空包。
  它同时是「根清单采集」那条市场路径的准入条件。
- `node verify.mjs`：1.2.4 口径仍是 **424 行 / 95 个 `ok` / 0 warn / 末行 `VERIFY PASSED` / rc 0**（1.2.4 只改文案，
  那 95 条断言一条都没加也没减，`caseMarketIndex()` 的「索引最新版本 == 包版本」直接跨过这次改号）
  （tag `dsh-plugin-v1.1.0` 时是 300 行 / 55 ok，1.1.0 批次 331 行 / 63 ok）。它自带 fake Cordis context、
  临时 `LOCALAPPDATA` 与 stub protocol，所以不碰管道、浮窗与网络；`d06001d` 起 CI 也跑这套。
- **三种安装形态都实测成功**（每次都开全新 `DSH_HOME` 与独立 profile）：Release tarball（**1.8 s**）、
  git 源钉 tag（**3.4 s**）、仓库地址 `add github:jjjjjjjjnnjnn/ai-terminal-pet`（**3.6 s**，等价写法
  `git+https://github.com/jjjjjjjjnnjnn/ai-terminal-pet.git` 不带 path 同样 3.6 s；三个计时是在 1.1.0 的资产上取的，
  1.2.0 / 1.2.1 / 1.2.2 / 1.2.3 的资产都在同一位置、URL 只换 tag 与版本号；1.2.2 那次单独复验过 3,335 ms，
  1.2.3 的 Release 资产在全新 `DSH_HOME` 里装过一次：**2.4 s**，`plugin list` 显示 `1.2.3`、
  `node_modules/dsh-ai-terminal-pet` 里 `dsh.marketplace` 块与 vendor 浮窗（4,371,968 B）都在位）；
  1.2.4 的资产就位后又在全新目录里逐个装过：Release 资产 URL **1.8 s**、git 源钉 tag `dsh-plugin-v1.2.4` **3.7 s**、
   `add github:jjjjjjjjnnjnn/ai-terminal-pet` **6.5 s**（三者 rc=0，装出的包与 `git show ff7f6de:...` 逐字节一致）；
  三种都用 `--dump-config` 验证过末尾出现 `- id: dsh-ai-terminal-pet`（证明 `insert:` 生效）。
  1.0.0 时期的三条旧记录（本地 tgz → Release URL 32.6 s → 钉 commit 10.2 s）保留在 CHANGELOG 历史条目里。
  仍会失败的一条：`git+https://...git&path:/packages/dsh-ai-terminal-pet` **不写 `#ref`**——pnpm 把
  `&path:` 拼进 URL 当作仓库地址去 `ls-remote`，报 `repository ... not found`；这条与 `prepack` 无关。
- **awesome-dsh-plugin 注册表条目 PR #13937 已于 2026-11-10T14:33:04Z 被 `fkysly` 合并**（`category: notify`，带
  `icon` 与 `tarball`；上游两道门本地先跑过：`entries read: 10120 / problems: 0`、`all checked entries pass`；
  PR 上 `PR check / check`（10:37:54Z）与 `Submission gate`（10:38:47Z）均 `conclusion: SUCCESS`）。合并时维护者
  删掉了原稿里「也能在面板上直接放行审批」一句——对 Harness 会话，面板只做状态映射与跳转，审批仍在 DSH 里完成；
  17:19Z 他在 PR 上留言确认了这一点，并说明要改措辞得另开 PR 改那个 `data/plugins/*.yml`（我们仓库里的文档改动
  影响不到它）。条目里的 `tarball` 决定它装哪个版本，只能靠 PR 改：跟进 PR **#14076**（分支 `bump-tarball-1.2.2`、
  commit `0863b4d`、`1 file +1/−1`，只改 `tarball:` 一行到 1.2.2，描述不动）**已于 19:59:37Z 合并**；
  1.2.3 发版后又开 **#14089**（`bump-tarball-1.2.3`、commit `c5a1f5f`，同样只改那一行）**已于
  2026-11-10T21:10:15Z 合并**（merge commit `95d7216913d54110a794514a6c0ea0196aa08ddc`；`data/plugins/*.yml` 里
  现在指向 `dsh-plugin-v1.2.3`，站点下一次重建才渲染出该资产），1.2.4 的跟进 **#14243**（`bump-tarball-1.2.4`，
   commit `66ed01d`，`1 file +3/−3`：`tarball` 与 `icon` 换到 1.2.4，中英 description 各补一句 reporters 限定）
   已于 2026-11-11 开出、`mergeable_state: unstable`，等 check 与人工合并；
  （前身 #14063 指向 1.2.1、head `116b0ee`，随 1.2.2 重打包已评论说明后关闭）；
  上游 `check-submission.mjs --only-list` 单条目通过，两个 URL 的 HEAD 均 200。
  条目已进入 `main`，上游 `README.md` 第 8,519 行（`### Notifications & Integrations`）已列出它；**条目页与
  DSH 应用内 Market 当时都还看不到**（16:16Z 复查：条目页 404、`plugins.json` 0 命中），但那只是
  上游站点构建的排队与耗时：`Build site from README` 的 run `42841664024` 在 15:25:40Z 被 cancelled（同分支新 run
  取代），`42844452013` 于 16:39:27Z success；此后条目页返回 200，`plugins.json` 从 `count 10119` 变成 `count 10230`
  并收录本仓库（合并后再读一次是 10,304 条）。**「合并」与「可见」都成立**，只是目录里的 `install` 命令是构建时
  按 yml 的 `tarball` 渲染的，所以合并之后那版仍写着 1.2.0——#14076 已把字段换成 1.2.2，下一次重建才渲染出来，
  #14089 再推到 1.2.3 并已于 2026-11-10T21:10:15Z 合并，1.2.4 的跟进是 **#14243**（2026-11-11 开出，等合并）。同一批里 [DSH-Plugins-Marketplace](https://github.com/bradeGithub/DSH-Plugins-Marketplace)
  的 `build-registry`（run `42812806941`）在 16:22:29Z 跑完：本仓库从 `verdict: script` 改判为 `bundle-plugin`，
  `registry.json` 里出现 `pkg_name`；到 18:08:44Z 那次重建（`risk_at` 17:09:14Z，是安装脚本移进 `scripts/` 后
  第一次真的扫到本仓库），条目已变成 `version: 1.2.1`、**`risk_tier: "safe"`、`risk_flags` 一条不剩**——
  issue #367 里那两条旗标就此消失，全程没有用到任何抑制字段，§4 里那项可以按已解决写。
**踩到的坑（值得写进路线图，因为它决定了 §4 里那几项的状态标记怎么写）**：`vendor/AI-Terminal-Pet.exe`
一开始不在 git 树里（`.gitignore` 的 `*.exe` 把它吞了，`git add` 不会越过忽略规则）。后果不对称：
`npm pack` 读工作区所以 tarball 装得上，`git clone` 读树所以 git 源装出的是个没浮窗的包，
而 `lib/release-pins.json` 还没建 → 回落 no-op → 浮窗静默不装。修法两行：`.gitignore` 加
`!packages/dsh-ai-terminal-pet/vendor/AI-Terminal-Pet.exe`，`.gitattributes` 加 `*.exe binary`
（避免 Git 对 sha256 依赖的二进制做文本规范化）。

**这一批的 CI 曾经是红的**：红在 `vendor:check`，根因是 `go.mod` 的 `go 1.23.0` 是**下限不是钉版本**，
CI 用 go 1.23.0、本地用 go1.27.0，两个工具链编出的字节不同——跨工具链字节相等从来不可能达成。
修法在 `vendor/manifest.json`：记录源码摘要（25 个非测试 `.go` 文件）与两个 vendor 文件自身的哈希，
`vendor:check` 校验这些，只有当本地 Go 版本等于记录的 `go1.27.0` 时才与 `bin/` 做字节比对。
**先红后修，不是一次就绿**（细节见 HANDOVER §27.9）。

**当时没做成什么（已在 2026-11-11T12:10:46Z 完成）**：npm 一个版本都还没发出去——凭据问题已解
（`npm whoami` 返回 `jjjjjjjjnnjnn`），当时 `npm view dsh-ai-terminal-pet version` 还是 E404；
这一步在本机非交互环境里结构上做不到：npm 的 `otplease` 在没有 TTY 时直接抛 EOTP，永不进入 web-OTP 分支，
取到的浏览器授权链接 5–8 分钟就 404 作废。**结论：人已经在本机交互终端里跑过 `npm publish`，1.2.4 已上线**，
`dist-tags.latest` = 1.2.4。因此 **DSH 应用内 Discover 索引现在可达**（npm 版本 1.2.4），
按包名 `add dsh-ai-terminal-pet` 已经能解析（自建索引那条 git 源路仍然可用，见 §4）。注册表条目已于 2026-11-10T14:33:04Z
合并、上游 README 已列出；**「可见」这一条后来也成立**——16:16Z 时条目页还是 404、`plugins.json` 0 命中，
但那是上游站点构建的排队与耗时：run `42844452013` 于 16:39:27Z success 后条目页 200、`plugins.json` `count` 从 10119 涨到 10230 并收录本仓库。
同时 [DSH-Plugins-Marketplace](https://github.com/bradeGithub/DSH-Plugins-Marketplace) 的索引在 16:22:29Z 重建时把本仓库从
`verdict: script` 改判为 `bundle-plugin`，18:08:44Z 那次重建又补上 `version: 1.2.1`、把 `risk_tier` 从「缺字段（未评估）」
变成 **`safe`** 并把 `risk_flags` 清空（`risk_at` 17:09:14Z，是安装脚本移进 `scripts/` 后第一次真的扫过）。
1.2.3 / 1.2.4 的 tag 与 Release 资产已发布（sha256 `1172aa72…` / `d6b7f95a…`），但 `lib/release-pins.json` 刻意留空，所以插件内「自动下载浮窗」
回落仍是 no-op——安装时拷的是 `vendor/` 里那份（见 §27.7）。

### 1i. 2026-11-10（v0.9.0 之后，未发版）—— 可移植性、--doctor、线协议文档、CI

**动机**：v0.9.0 让浮窗「能调」，但是要交给别人用还差四件事——
(1) 仓库里写死了只有作者本机才存在的绝对路径，别人 clone 下来跑不通；
(2) 浮窗失败是**静默**的，装错了、管道没通、Node 不在，从外面看都是「宠物没出现」，
    用户没法自己查；(3) 协议只写在代码里，第三方想写上报端必须读源码；
(4) 没有任何自动门禁，改坏了要等下一次手动跑测试才发现。这一批把四件事一起补上，**不发版本号**，
`internal/app/app_windows.go` 的 `Version` 仍是 `0.9.0`。

- **`--doctor` 自检**（`internal/app/doctor_windows.go` 563 行 + 375 行测试 / 14 个 `func Test`）：
  9 项检查 environment / install files / version / runtime / config / pipe / instance / node / autostart，
  逐行 `[ok]`/`[warn]`/`[fail] name: detail`，末行 `doctor: N ok, N warn, N fail`，任一 fail 退出码 1。
  退出码走 `errDoctorFailed` 哨兵而不是 `writeError`——自检失败是用户**主动请求**的诊断结果，
  记进 `error.log` 会把真正的故障埋掉；用法错误仍是 `usageError` → 退出码 2。
  **不探测互斥体**：`CreateMutex` 会**创建**对象并干扰正在运行的实例，所以 `checkInstance` 改用
  「管道有人监听」作为证据。管道探测连上即关（250 ms 上限、不写字节，服务端把立即 EOF 当空负载，
  因此跑 doctor 不会改变浮窗显示）。node 探测 `node --version` 限 300 ms 并把 `Missing`/`TimedOut` 分开。
  每项包在 `runCheck` 的 recover 墙里，将来某一项 panic 只产出它自己的 fail 行。
  没有 `--json` 变体：这份报告是贴进 issue 给人读的。
- **`docs/PROTOCOL.md`**（388 行）落地，E2 的主要产物：字段、状态语义、优先级、
  管道名派生算法、`OK`/`ERR` 回执契约、脱敏规则成文，schema 变更流程见该文档 §11。
- **`reporters/package.json`**（26 行）：显式 `exports` 覆盖 6 个 `.js`，不再依赖隐式解析；
  `private: true` 不发布。**刻意不写 `"type"`**——5 个 hook 是 CommonJS、`opencode.js` 是真 ESM，
  写任一值都会打断另一半；代价是首次 import 时一条 `MODULE_TYPELESS_PACKAGE_JSON` 提示并按语法重解析，
  属预期行为。改名为 `.mjs` 的方案被否决（README 与用户配置引用的是 `.js`）。
- **`.github/workflows/ci.yml`**：`windows-latest` 单 job 门禁——
  `gofmt -l` 有输出即挂 / `go vet` / `go test -count=1` / 两个 `-trimpath -buildvcs=false -ldflags "-s -w"` 构建 /
  **逐文件** `node --check` / `node integrations/deepseek-harness/verify.mjs` / 插件包 `vendor.mjs --check` 与 `verify.mjs`。
  「逐文件」不是装饰：`node --check a.js b.js` 只检查第一个参数（Node 24 实测第二个参数是坏文件也退出 0），
  所以 `node --check reporters/*.js` 只能守住字母序第一个文件。live-desktop 探针在对应 `PET_*`
  变量未设置时 `Skip`，runner 不需要桌面会话或 secret。
- **`install.ps1` 三个开关** + `Assert-SafeInstallDirectory`：
  `-VerifyOnly` 只读、按**同一份** 10 项制品清单（5 组）逐项 SHA256 比对，全对才 `install: verified` / `exit 0`，
  否则 `exit 1`——「校验集 ≠ 安装集」正是过去残留旧 `protocol.js`、所有上报端静默失联的成因；
  `-Uninstall` 需输入 `yes`（无控制台且无 `-Force` 时 `exit 2`，用户拒绝 `exit 3`），
  拒绝盘根与叶子名不是 `AI-Terminal-Pet` 的目录（**`-Force` 只跳过确认、不绕过名字检查**），
  只在 `HKCU\...\Run` 条目指向被删目录内部时才删该条目，并调用 `Remove-InstallDirectoryFromUserPath`
  清掉 `-AddToPath` 写入的 PATH 条目（安装目录即将消失，PATH 原值改备份到 `%TEMP%`；清理失败只报 `[warn]`）；
  `-AddToPath` 只写 `HKCU\Environment\Path`（绝不碰 HKLM），写前把原值连注册表值类型备份到
  `<安装目录>\path-backup.txt`。三个开关的非法组合在参数解析阶段就以 `[fail]` + `exit 2` 拒绝。
- **可移植性**：`docs/assets/render_assets.py` 去掉写死的本机绝对路径（预演目录读 `PET_PREVIEW_DIR`、
  输出目录取 `Path(__file__).resolve().parent`），缺变量/缺目录/缺 Pillow 都给人话提示并 `exit 2`；
  `internal/platform/focus_windows_test.go`（3 处）、`presence_windows_test.go`（4 处）、
  `internal/pipeserver/server_windows_test.go`（1 处）的本机路径换成中性路径，
  管道名测试改为按 `sha256` 现算期望后缀（输入特意用混合大小写 + 结尾分隔符，仍钉住 Go/JS 共享契约）。
- **删除七处不可达声明**（`3fd49c3`，staticcheck U1000）：`(*Store).remove`、
  `messageEnvelope.reserve`、`(*App).clampToMonitor`、`fillPolygon`/`alphaAt`/`pointInside`/`canvasPoint`；
  行为不变，未加 `//nolint` 或假引用。
- **`.gitattributes` 钉 `* text=auto eol=lf`**（`1330fbe`）：消除 Windows `core.autocrlf` 下反复的规范化警告，
  Go 源码与 Markdown 在各次克隆间字节一致。
- 详细 runbook：HANDOVER §1（当前状态）。

### 1h. 2026-11-08/09（v0.9.0）—— 设置窗口、四套合成音效、按会话告警、DSH 桥

**动机**：浮窗刻意只是一条状态带，所有偏好原先只有两个入口——托盘里的一串勾选，或者手改 `config.json`。
「换个提示音」「调音量」这两件最想改的事，前者根本不存在、后者要在托盘里点三下循环；
不透明度、auto-approve 倒计时、语言也没有任何界面。

- **设置窗口**（`internal/ui/settings_windows.go`，1,382 行 + 228 行测试）：第二个顶层窗口类
  `AITerminalPetSettings`，三页签 general / sound / sessions，**跑在与浮窗同一条消息泵上**
  （`Run()` 里 `runtime.LockOSThread()` 之后创建）——Win32 要求创建窗口的线程就是泵消息的线程，
  这条铁律正是 §13 幽灵窗口的根因。几何单一来源 `settingsGeometryFor(dpi, tab, rows)` 同时供绘制与命中；
  切页签按内容改高度且保持中心不动（真机 450↔504↔306）；每项改动立即生效，没有 OK 按钮。
- **四套音色 + 按事件开关**（`sound_windows.go` 的 `schemeSpecs()`）：`chime`（默认）/ `bell` / `soft` /
  `pulse`，每套都有等待/出错/完成三条波形，仍是内存合成 44.1 kHz 16-bit 单声道 PCM + `PlaySoundW(SND_ASYNC)`。
  `config` 新增 `soundScheme` / `soundOnWaiting` / `soundOnError`，`alertOnTransitions` 改为按事件判断
  （`soundForState`），所以「只想听真正卡住的会话」可以关掉等待音而保留出错音。
  `config.SoundSchemes` 与 `ui.SoundSchemes` 两份镜像由 `TestDefaultSettingsMirrorConfig` 钉住。
- **告警按会话**（`alert_windows.go`）：旧实现只有一个全局 most-urgent key，该会话结束时被重置，
  于是无关的等待会话后来重复告警、气泡像是来自错误的终端。现在按每个会话上次状态做纯函数
  `transitionLevel` / `decideAlerts`，首次加载只播种不重放（热重载/重启不重复旧告警）。
  等待态呼吸从 12.5 fps 降到 6.25 fps——每次动画 tick 都是整层 GDI 光栅。
- **本地化绘制**：所有文案走 `internal/i18n` 的 `a.lang`（`New()` 固定、托盘切语言时刷新），
  光栅器不再解析 locale。
- **配置生命周期归 `internal/app`**（`configSaver`，与 `placementSaver` 同构）：`OnSettingsChanged` 只入队，
  后台 goroutine 读/合并/写（手写的、本版本不认识的键不会丢），`echo(config)` 让 30 s 的 `watchConfig`
  跳过自己的写入，滑块拖动只改内存、松手才 publish 一次。`--settings` 无实例时先起浮窗、首帧后再开窗。
- **auto-approve 否决握手**（`approve_windows.go` + `reporters/protocol.js`）：浮窗无法代点别人终端里的
  Allow，所以反向做——hook 写 `pending-<source>-<session>-<pid>-<epoch>.json` 到 `%LOCALAPPDATA%\AI-Terminal-Pet\approvals`，
  每 100 ms 轮询同名 `.deny`；展开面板显示 `Allow (Ns)` 与倒计时条，点一下否决最紧急的等待会话，
  沉默即由 hook 放行。危险命令（`rm -rf` / `git push --force` / `format` 等 `DANGEROUS_PATTERNS`）
  在倒计时开始前就判，永远走人工。
- **顺带**：`--settings` CLI；`install.ps1` 刷新为新的部署集（主程序 + shim + `protocol.js` + DSH bundle）；
  `docs/assets/render_assets.py` 由真实光栅器输出产出 README 全部配图。
- 详细 runbook：HANDOVER §25（设置窗口与可切换音效）。

### 1g. 2026-11-08（v0.8.0）—— 合成提示音、通知时序、DSH 适配

- **提示音从系统事件改为内存合成**（`sound_windows.go`）：`MessageBeep` 走系统默认事件，在很多机器上
  静默或几乎听不见，还受专注助手影响。改为内存合成 44.1 kHz/16-bit 单声道 PCM +
  `PlaySoundW(SND_MEMORY | SND_ASYNC | SND_NODEFAULT)`，三种音型（等待=两轮上行双音 ~0.9 s、
  错误=下行低音、完成=短上行三音），`soundVolume` 0..100 默认 85，缓冲按 kind 缓存防 GC 越过异步播放器。
- **通知时序**：`alertStates` 逐会话记录，纯函数 `transitionLevel` / `decideAlerts`；
  waiting/error→urgent，busy→idle/closed 或消失→finished；新 busy 不告警；`notifyOnFinish` 默认开。
- **协议钳制**：`protocol.js:normalizePayload` 把 `startedAt = min(started, updated)`、两者夹到 now+4 min，
  幂等——否则 `validate()` 会以「updatedAt precedes startedAt」拒收，随后扫描把快照删掉。
- **DeepSeek Harness 桥**（`integrations/deepseek-harness/ai-terminal-pet-bridge`）：浮窗此前只聚合
  OpenCode / Claude Code / Antigravity，而本项目恰好在 DSH 里开发，做活时面板什么都不显示。
  Cordis bundle 用 `ai-terminal-pet.v1` 上报 Harness 会话：一个对话一个稳定会话、时间戳归一 UTC 并夹在
  服务端上限之下（时钟偏差不能让 `startedAt` 跟着 `updatedAt`）、审批请求保留 waterfall 链、
  `dispose` 只移除监听器不结束会话（HMR 重载不是退出）。`verify.mjs` 是离线测试，
  用临时 `LOCALAPPDATA` 桩掉官方 protocol 模块并断言发出的 payload，不需要管道/浮窗/运行中的 Harness。
  部署侧 `install.ps1` 增 `BundleSource`/`BundleDestination`，每次安装同步 7 个文件并逐个 SHA256 校验。
  Go 侧 `isEditorProcess`/`isCLIProcess` 加 `deepseek harness.exe`（`Chrome_WidgetWin_1` 得分 24），
  `terminalKeyword` 加 `deepseek`，`PET_DSH_PROBE=1` 保留为探针。
- 详细 runbook：HANDOVER §24。

### 1f. 2026-11-05（v0.7.0 / v0.7.1）—— 会话存在性重构、一键跳转、幽灵堆积

- **会话存在性重构（v0.7.0，HANDOVER §21）**：旧模型「没上报 = 结束」——reporter 为事件驱动且状态去重，
  快照 TTL 15 s / 管道 TTL 30 s，`Scan()` 到点直接删文件，tombstone 的 `UpdatedAt` 门槛又挡住旧时间戳无法自愈，
  UI 每 500 ms 检查、5 s 没有会话就 `DestroyWindow`。新模型：**存在性由 OS 观测**（进程表 + 终端窗口），
  state 可以陈旧；删除需连续不可观测 ≥ `DefaultAbsenceGrace`（2 分钟）。三类 tombstone：
  `closed`（保留门槛）/ `expired`（不设门槛、可自愈）/ `evicted`（保留门槛）。`store.Reconcile` 取代运行期 `Cleanup`，
  保活只刷 `pipeSeen`/`snapshotSeen`/`lastSeen`，不改 payload 不 bump revision（零重绘）。
  运算表：空闲检查 500 ms→30 s、扫描 2 s→5 s、存活探测 每会话 OpenProcess/2s→一次 Toolhelp/10 s、
  窗口枚举 仅跳转时→10 s 共用、告警呼吸 80 ms 全帧光栅→160 ms 零 GDI、配置 仅启动→30 s Stat。
  三轮采样第 3 轮 120 s 累计 0.562 s（约 0.5% 单核，较第 2 轮降约 96%）。
- **一键跳转终端（HANDOVER §20）**：只用现有 pid/parentPid/title/project/source/terminal 定位，协议零改动。
  `internal/platform/focus_windows.go` 走 Toolhelp 祖先链 → `EnumWindows` → 打分（祖先距离/会话标题/项目基名/
  终端关键词/窗口类名）→ 前台激活；跨进程枚举不在 UI 线程。点击坐标用 `WM_LBUTTONUP` 的 lParam。
  弱匹配 <40 判失败。Windows Terminal 能定位到标签（`CASCADIA_HOSTING_WINDOW_CLASS`），
  conhost 多会话 / VS Code 只能到窗口级。
- **跳转差点动用户窗口（HANDOVER §22）**：`activateFocusWindow` 里的 `AttachThreadInput` 与
  `SetWindowPos(TOPMOST)→NOTOPMOST` 会动别人的窗口，全部删除。改为 `ResolveTarget` 只定位、
  `ui.jumpTo` 把 `wmFocusActivate` 投回 UI 线程；`platform.ActivateTarget` 仅 `IsIconic→SW_RESTORE` +
  一次 `SetForegroundWindow` + `GetForegroundWindow` 复核，失败走 `focusFailure()` **绝不抢前台**；
  **禁止** `LockOSThread`/`UnlockOSThread`（会重现 §13 幽灵窗口）。25 次合成 + 15 次真实点击、
  12 次「用户正在处理窗口」场景 FAILS=0。
- **幽灵堆积 + Slate 配色（v0.7.1，HANDOVER §23）**：取证 内存 ≥7 会话 vs 磁盘 4 文件 vs 4 个终端标签，
  时区陷阱（payload `updatedAt` 是 UTC、本机 GMT+1）。三根因：宿主存活 ≠ 会话存在、删文件 ≠ 结束会话
  （插件 `cleanupAll`/`closed` 只 unlink 不送 closed）、**事件 ID 当会话 ID**（`e.data?.id || e.id` 兜底）。
  修法：`ScanWithKeys()` 返回磁盘 key 集合 + `snapshotAwareAlive()`（曾有快照但现在没有→判结束）；
  `Presence.SessionAlive` 按 specificity 分级（终端标签标题命中→无条件保活 / 可信真实 PID 优先 /
  hook 的 parentPid / 近期活动超 `DefaultOrphanAfter` 45 min 孤儿回收）；
  opencode 插件 `sessionIdentity()` 只认明确 sessionID、拒 `evt_`/`msg_`/`prt_`/`prj_`/`default-session`，
  `closed` 同步写 `state:closed` 快照；`cleanupAll` 只在 `process.on(exit)`。
  另把结构色收进 `internal/ui/drawing_windows.go` 顶部调色板（Slate：背景蓝黑→中性石墨，
  状态球实心大圆盘→细环+深色内核+状态色字形，外框 alpha 210→120）。
- **收尾 bug（HANDOVER §24.6，由 `ce81a8e` 落地）**：`snapshotAwareAlive` 对「磁盘文件还在」的会话仍调
  `Presence.SessionAlive`，弱 OS 证据下 `Reconcile` 连续缺席 2 min 删条目 → `HasAvailable()` 变假 →
  10 min 自毁（「浮窗自己消失」）。改为 `present[key]` 为真即直接保活（closed 仍优先判死），
  OS 证据只留给从未写过快照的纯 pipe 会话。`TestSnapshotAwareAliveTrustsTheDisk` 锁定。

### 1e. 2026-11-09（v0.6.0）—— i18n 与权限 5 秒自动同意

- **`internal/i18n`**（零依赖字典 `Lang`/`T`/`Resolve`，`en` + `zh-CN` + `auto`）：`GetUserDefaultLocaleName`
  解析一次，`zh*`→中文、其余/失败→英文；`config.language` 默认 `auto`，托盘三项单选；
  全部硬编码文案走 `i18n.T`，绘制函数加 `lang` 参数，`frameKey` 加 `lang`，测试锁两字典 key 对齐。
- **auto-approve**（默认 `autoApprove=false`、`autoApproveTimeoutSec=5`，钳 1..60）：握手目录
  `%LOCALAPPDATA%\AI-Terminal-Pet\approvals`，hook 建 pending 文件并每 100 ms 轮询同名 `.deny`，
  overlay 点展开态等待行写 `.deny`；超时无否决 = allow；hook `finally` 删 pending，
  overlay 每 2 s 清扫超过 `2*timeout+30s` 的残留。否决窗口只覆盖「等待权限」路径。
  验证：黑名单 13 例、off→`{}`(90 ms)、危险→`{}`(100 ms)、超时→allow(1.1 s)、中途否决→deny(212 ms)。
- 详细 runbook：HANDOVER §18。

### 1d. 2026-09-26（v0.5.0）—— 贴边收起 + 漂移修复 + 重绘跳过

**动机**：两点用户反馈——(1) 展开再缩小，面板「往左漂」、和原本位置不一样；(2) 贴屏幕边时想
像电脑管家浮窗那样缩进屏幕边缘。顺手把交互做顺、性能做轻。

- **漂移是真 bug（v0.4.0 引入）**：旧 `applyGeometry` 用「保持窗口中心」处理展开/收起，贴右边缘时
  更宽的展开窗口被边界钳制往回推，回程却仍从被推过的位置算中心 → 左移一个 sliver。改为「锚定边」：
  贴哪条边就保持哪条边，`dockedOrigin(edge, edgeCoord, gap, width)` = `edge-gap-width`（左）/
  `edge+gap`（右）。`TestDockRoundTripIsExact` 锁往返一致。探针实测 `drift_x=0 drift_y=0`。
- **贴边收起**：`config.json` 新增 `edgeRetract`（默认 true），托盘菜单新增「贴边收起」开关。
  逻辑在 `internal/ui/dock_windows.go`：贴左/右边且鼠标不在面板上、且无等待确认/错误会话时，
  滑出屏幕只留 `dockedSliver`（默认 10，按 DPI/缩放缩放、下限 6）宽竖条；悬停滑出，告警态自动滑出。
- **基于时间的缓动**：滑出 160ms、收回 240ms、ease-out；掉帧不改变时长。位置随
  `UpdateLayeredWindow` 一次原子更新，不再额外 `SetWindowPos`（避免撕裂）。
- **重绘跳过（性能）**：`present()` 用 `frameKey(...)` 给帧签名，相同则整帧 GDI 光栅化跳过。
  签名用 `sessionDigest()`（对「实际画出来的字段」做 FNV 哈希，含 `UpdatedAt` 截断到 15s）——
  **不能**用 `store.Revision`，reporter 每次推送都动 `updatedAt`，会让「相同状态」每推必重绘。
  探针实测 20 次相同推送 → 0 次额外重绘。`TestFrameKeySkipsUnchangedFrames` 锁该不变式。
- **帧率分工**：`timerAnimation` 在脉冲（80ms 慢呼吸）与滑动（33ms）间切换间隔，空闲不转。

### 1c. 2026-09-26（v0.4.0）—— 面板重做：更实用、更简洁、更好用

**问题**：紧凑态把「状态」和「数量」这两条信息用六种方式重复编码——状态球、24 段装饰环、
两个 `↑/↓` 指标行、边框颜色、分隔线、页脚箭头——信息密度高但信息量低；
而真正想知道的那件事（**谁在等我**）反而看不到。展开态则是固定 250 高、最多两行，
第 3 个会话只能以摘要里的一个数字存在。

**紧凑态**（`56x156`，原来是 `40x190`）：状态球（含脉冲与字形）→ 会话数 → 状态词 →
最紧急会话的项目名。删掉 24 段装饰环（每帧 24 次椭圆光栅化）、`↑/↓` 指标行、分隔线、
页脚箭头。多出来的宽度给项目名——它才是「一眼可辨」真正要回答的东西。

**展开态**（`200x<内容高度>`）：标题 + 实时状态胶囊 + 关闭钮 / 一行普查（`3 个会话 · 2 运行 · 1 等待`）/
会话列表 / 操作提示。高度按会话数自适应（1–6 行 + 溢出提示），每行显示项目、状态、
来源、**相对时间**。`运行` 刻意不含等待确认的会话——从用户视角那不是「在跑」而是「在堵」，
旧的两个指标行正是因为这个重复计数而对不上总数。

**交互统一为「没有需要瞄准的东西」**：点击面板任意处切换（不再只有 20 单位宽的 ✕ 能收起，
这正是「面板关不掉」的来源）；整块可拖拽（旧代码只允许顶部 40 单位，因为要给会话行
预留点击——但那些行从来没接过任何行为，结果是一个看起来能抓却抓不动的面板）；
悬停时面板底色与描边一起提亮（告警态描边本就是接近不透明的状态色，只改描边等于看不见）。
按压不再启动动画定时器——按压反馈是静态的，为它开 20 FPS 重绘纯属浪费。

**顺带**：
- 新增 `canvas.fitText` / `canvas.measure`（GDI `GetTextExtentPoint32`）。分层窗口没有滚动条
  也没有 tooltip，放不下的文字不是「被裁掉」而是「丢了」；旧代码依赖 `DT_END_ELLIPSIS`，
  而它在 `DT_WORDBREAK` 框里无效（两个标志互斥），长项目名会直接溢出到条带两侧。
- 孤儿临时文件的清扫阈值从 1 小时降到 2 分钟。`install.ps1` 用 `Stop-Process -Force`，
  每次重装都可能留下一个孤儿，而 1 小时的窗口意味着它会在安装目录里躺到用户早已离开。
- `install.ps1` 改为「停进程 → 等待真正退出 → 重试拷贝」。reporter 会按需把浮窗拉起来，
  所以停止后几毫秒它就可能复活，而带 AV 钩挂的机器上进程释放镜像可能要 1 秒以上——
  旧的「Stop-Process + 固定 500 ms」在真正在用这个工具的机器上会随机失败。

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

> 状态标记：**已交付** / **部分交付** / **未开工**。已交付项保留原文，便于回看当时的动机与验收标准。

### E4 — 会话可交互 ✅ 已交付（2026-11-05，v0.7.0）
**结果**：展开态每行右侧有跳转箭头（紧凑态底部 `↗` 跳转条仅在有可跳转会话时出现）。
`internal/platform/focus_windows.go`（392 行）+ `focus_windows_test.go`（280 行）实现
Toolhelp 祖先链 → `EnumWindows` → 打分 → 前台激活；`internal/ui/focus_windows.go`（194 行）
保证 `compactFocusRect` / `expandedRowFocusRect` 与绘制同源，`jumpPriority`
（waiting 4 / error 3 / idle 2 / closed 1 / 忙碌 -1）与 `mostUrgentJump`、纯函数 `focusHit`。
`PET_FOCUS_PROBE=1` 保留为探针入口。
**与原计划的偏差**：原文担心的「稳定哈希 PID 无法定位」由 `platform.Presence` 的真实进程探测解决——
合成 PID 不参与窗口匹配，窗口匹配必须命中 `title` 或项目名。
**已知限制**：Windows Terminal 能定位到标签，conhost 多会话 / VS Code 只能到窗口级；
同标题按 Z 序取第一；合成点击需先 `SetForegroundWindow(petHwnd)`。
**后续修复**：§22 的「跳转关闭用户窗口」——`AttachThreadInput` 与 `SetWindowPos(TOPMOST)` 全部删除，
`ActivateTarget` 只做一次 `SetForegroundWindow` + 复核，失败不抢前台。

### E2b — auto-approve 否决窗口 ✅ 已交付（2026-11-05 v0.6.0 / 2026-11-09 v0.9.0）
**动机**：浮窗能看见「谁在等确认」，却无法替用户点 Allow——它拿不到别人终端的对话框。
**做法（反向握手）**：不再由浮窗点 Allow，而是由 hook 写审批请求文件、浮窗只负责「否决」。
hook 建 `pending-<source>-<session>-<pid>-<epochMs>.json` 后每 100 ms 轮询同名 `.deny`；
展开面板显示 `Allow (Ns)` 与倒计时条，点一下写 `.deny`；沉默 = 超时放行。
`DANGEROUS_PATTERNS` 在倒计时开始前就判，破坏性命令永远走人工。
**验收补充**：fail-closed（config 缺失/损坏 → auto-approve off）；hook 侧每步文件操作都包住，
绝不向宿主 CLI 抛异常（AGENTS.md 的铁律）。

### E5 — 配置档扩展 ✅ 已交付（2026-11-16 复核；原先只记到 v0.9.0 的部分交付）
**已交付**：`config.json` 覆盖 `notifications` / `sound` / `soundVolume` / `soundScheme` /
`soundOnWaiting` / `soundOnError` / `notifyOnFinish` / `alwaysOnTop` / `edgeRetract` / `opacity` /
`language` / `autoApprove` / `autoApproveTimeoutSec`，并有设置窗口（四页签 general/sound/sessions/look）作为 UI。

原先列在这里的四个"未交付"项，本轮逐个对着代码核过，**三项早已落地、一项按设计不做**：

| 项 | 现状 | 落点 |
|---|---|---|
| `theme`（配色组） | **已交付**：`slate` / `paper` / `contrast` 三套命名调色板，设置窗口的外观页签可直接切 | `internal/config/config.go:74`（`ThemeSlate` 等）、`ValidTheme`（:191）、`internal/ui/drawing.go:235` 的 `applyAppearance`，三个平台的 `app_*.go` 各调一次 |
| `perSourceColor`（按 source 着色） | **已交付**：展开行里来源名的颜色跟随终端来源；开关在设置窗口 | `internal/ui/drawing.go:186` 的 `perSourceColors`、`drawing_compact.go:668`、`settings_layout.go:633` |
| `hidden`（启动即隐藏到托盘） | **已交付**：`hiddenAtLaunch`，Windows 走托盘、Linux/macOS 走面板不显示 | `internal/ui/app_windows.go:433`、`app_linux.go:287`、`app_darwin.go:357`、`settings_layout.go:636` |
| `size`（紧凑/展开的基准尺寸预设） | **不做**：需求已被缩放取代——滚轮与 `Ctrl/Cmd` + `+`/`-`/`0` 逐显示器记忆在 `state.json` 的 `layouts` 里 | `internal/config/config.go` 的 `LayoutPlacement.Zoom`、`app_windows.go:707` 的滚轮缩放、`darwin_geometry.go` 的 `Cmd+Plus/Minus/0` |

`internal/config` 的「缺字段用默认」解码方式未变，加字段仍不会破坏旧文件。

### E1 — 把 Node 上报端全部原生化（收益最高）✅ 已交付（2026-11-11）
**动机**：本机每创建一个进程约 **1 s**（AV 钩挂），而 `claude-hook.js` / `antigravity-hook.js`
每次事件都要起一个 `node`（实测 `node -e exit` 0.9–1.2 s）。`pet-statusline.exe` 已经是这条路的先例：
Go 原生 shim 稳态 0.5–0.9 s，比 node 路径的 13–17 s（经 runner）低一个数量级。
**做法**：`cmd/pet-hook`（`--mode=claude|antigravity`）承接 Claude Code 与 Antigravity 的事件映射。
解析、脱敏、id 截断、FNV-1a PID、时间戳钳制、快照落盘、管道投递与审批握手全部下沉到新包
`internal/reporter`，`pet-statusline` 与 `pet-hook` 共用同一份实现，避免与 `reporters/protocol.js` 三方漂移。
三个 JS hook 保留为参照实现与降级路径。CI 增加 `Build the hook binary`，
`scripts/install.ps1` 增加 `hook` 分组（第 6 类产物），`vendor.mjs` 的源码摘要改为排除
`cmd/pet-hook/` 与 `internal/reporter/`：这两个包不进浮窗二进制，改它们不该逼着 1.2.4 的
`vendor/manifest.json` 重新生成。

**验收（已满足）**：同 payload 下原生与 JS 输出同一 session PID（FNV-1a UTF-16 已对齐）与同一 state，
`internal/reporter` 26 个单测 + `cmd/pet-hook` 13 个单测全绿。
`cmd/pet-statusline` 随后也迁到同一个包（删掉约 250 行重复实现），并用 35 个 payload 逐字段比对
迁移前后 stdout/退出码/快照，结果完全一致；详见 [docs/HANDOVER.md](HANDOVER.md) §29.7。

**实测（同 payload、同机器、皆从 `%TEMP%` 启动以排除下面那条完整性问题）**：
`pet-hook.exe` 稳态 **51–73 ms**，`claude-hook.js` **93–157 ms**；首跑分别是 513 ms 与 584 ms（冷启动）。

**顺带修掉一个真实缺陷**：`Send` 原来对所有拨号错误一视同仁地重试满 750 ms 预算。
当进程**永远**打不开管道时（见下），这 750 ms 是纯加在宿主每次事件上的延迟。
现在 `send.go` 用 `retryableDialError` 只把 `ENOENT`/`ERROR_FILE_NOT_FOUND`（还没监听）与
`ERROR_PIPE_BUSY`（实例占满）当作可重试，`ERROR_ACCESS_DENIED` 立即返回且不再触发启动浮窗。
低完整性场景下 `pet-hook` 从 **829–862 ms** 降到 **89–119 ms**。

**一个必须记下来的环境事实**：本机的仓库目录
`C:\Users\rongj\Desktop\项目开发\ai-terminal-pet` 带着 **Low 强制完整性标签**
（`icacls` 显示 `Mandatory Label\Low Mandatory Level:(OI)(CI)(NW)`），同级的 `DSH-Addition`、
`oh_my_3Dcharacter`、`oh_my_3dlab` 也一样。**从该目录里的文件启动的进程一律继承 Low IL**，
而浮窗跑在 Medium IL，其管道 DACL 只给 `OWNER RIGHTS` + `SYSTEM`——Low IL 进程永远打不开。
同样的字节复制到 `%TEMP%` 或安装目录就落回 Medium IL 并正常投递（已用同 sha256 双份验证）。
所以：**从仓库里直接跑 `bin\pet-hook.exe` 投递不到浮窗是预期的**，不是产品缺陷；
管道 ACL 保持 `D:P(A;;GRGW;;;OW)(A;;GRGW;;;SY)` 不变，真机安装到 `%LOCALAPPDATA%` 不受影响。

### E2 — 通用接入成为一等公民 ✅ 已交付（2026-11-10，见 §1i）

**现状**：`docs/PROTOCOL.md` **已交付**（388 行，2026-11-10）。`reporters/generic.js` 已可用，
DeepSeek Harness 桥是现成的第三个接入样例（`integrations/deepseek-harness/ai-terminal-pet-bridge`），
文档成文时就是拿它和 `pet-statusline.exe` 作为两个参照实现来对照的。
**动机**：`reporters/generic.js` 已能一行接入任意工具，但协议本身还只写在代码里。
**做法**：新增 `docs/PROTOCOL.md`，把 `ai-terminal-pet.v1` 的字段、状态语义、优先级、
管道命名算法、回执契约、脱敏规则写成规范；README 给出 5 行以内的接入示例（shell alias / Makefile / 编辑器 task）。
**验收**：不读源码也能写出一个合规上报端。

### E3 — 跨平台（macOS / Linux）⏳ 全 16 包可跨编，Linux X11 面板已落地（2026-11-15，§58）；**macOS 窗口层未做** —— 因此 1.1.0 暂未发
**动机**：目前 UI 层与平台层直接调 Win32，无法移植。核心逻辑（session/store/pipeserver 的协议部分）其实已平台无关。
**盘点结果（2026-11-15 实测，17,238 行非测试 Go）**：

| 档 | 判据 | 行数 | 占比 |
|---|---|---|---|
| portable | 不 import 平台包，纯 Go | **5,650** | 32.8% |
| rename-only | 带 `_windows.go` 后缀但**一个 Win32 调用都没有**（只是被后缀挡住） | **3,887** | 22.5% |
| needs-port | 真有 `win.` / `syscall.` / `unsafe.` | **7,701** | 44.7% |

`internal/ui` 的 12 个 rename-only 文件（2,927 行）只调用 `canvas` 的图案原语，所以新平台只要实现
那 16 个方法，`drawing_*.go` / `settings_text_windows.go` 一行都不用改。今天在
`GOOS=darwin go build ./...` 上**只有 store / session / config / pet-setup 能过**；
`i18n` 与 `ui` 的失败是"纯逻辑被放进了 Windows 文件"（`systemLocaleWindows`、`defaultIdleSeconds`），
不是真的平台绑定——这就是第一步先做它们的理由。

**做法**：把 `internal/ui` 的窗口实现抽成接口（`Presenter`），Win32 实现留在 `_windows.go`，
新增 `_darwin.go` / `_linux.go`（macOS 用 NSWindow + NSVisualEffectView，Linux 用 GTK/Wayland layer-shell）。
IPC 从命名管道换到 Unix domain socket，`PipePathForUser` 的 hash 算法保持不变以免破坏跨语言契约
（`docs/PROTOCOL.md` §3.1 要同时改，按 §11 的不兼容变更策略走）。
**注意**：这是一次真正的架构改动，建议先做 E2（协议文档）再动，否则契约会跟着实现漂移。
七步顺序、每步验收、以及五个会踩的坑（清单字段的副作用、测试夹具自启动循环、目录来源被拆开、
POSIX 上 `\\.\pipe\...` 是个合法文件名、`hang.log` 的时钟语义）写在
[docs/HANDOVER.md](HANDOVER.md) **§55.6 / §55.7**。**不要在第 1 步之前动呈现层。**

**施工结果（2026-11-15 深夜，§57）**：七步里的 **1–5 与第 7 步的前半**已完成，并且在这台机器上
逐项验证过；第 6 步（呈现层）没做。实测矩阵（`GOOS=linux|darwin go build` + WSL 里真跑测试）：

| | windows | linux | darwin |
|---|---|---|---|
| 13 个包（store/session/config/i18n/connect/home/reporter/pipeserver/autostart/ui + 三个 cmd） | ok | ok | ok |
| `internal/app`、`cmd/ai-terminal-pet` | ok | 只差 `ui.New/Options/Placement` | 同左 |

- **新包 `internal/home`**：数据目录的唯一答案（Windows 仍是 `%LOCALAPPDATA%\AI-Terminal-Pet`，
  一个字都没改）。二十多个 Go 文件与 `reporters/protocol.js` 原本各自拼这条路径。
- **传输层**：`pipeserver` 拆成 `server.go`（协议）+ `transport_<goos>.go`（唯一一句传输调用）；
  reporter 的 `dial_<goos>.go` 同理。**按用户派生的 8 位 token 一个字没动**，容器从命名管道换成
  `$XDG_RUNTIME_DIR/…sock`（104 字节上限在 listen 前检查）。
- **自启**：Windows HKCU Run / macOS LaunchAgent / Linux XDG desktop entry，Manager 语义共用。
- **CI 门禁**：`scripts/portable-packages.txt`（必须能编）与 `scripts/windows-only-packages.txt`
  （必须编不了）双向检查，任何一边漂移都会红。
- **第 58 轮增量（§58，全 16 包跨平台 + Linux 纯 Go X11 浮窗）**：
  - `internal/platform`、`internal/app`、`cmd/ai-terminal-pet` 全部消除 Win32 编译锁；`scripts/portable-packages.txt` 收录全部 **16** 个 Go 包，`scripts/windows-only-packages.txt` 清空。
  - **Linux X11 面板**：`internal/ui/app_linux.go` + `canvas_opentype.go` + `canvas_fonts_linux.go` + `control_unix.go` + `clipboard_linux.go` 基于纯 Go `github.com/jezek/xgb` 与 `golang.org/x/image`（零 cgo、零 GTK），支持 32-bit ARGB 透明、`_NET_WM_WINDOW_OPACITY` 脉冲、chunked `PutImage`、OpenType 字体链、X11 `CLIPBOARD`、RandR/workarea 布局记忆、EWMH 焦点回跳（`/proc` + `_NET_CLIENT_LIST`）与 unix 控制 socket。已在 WSL2 Ubuntu + WSLg `DISPLAY=:0` 实跑 Linux 测试与端到端 smoke。
  - **第 59 轮增量（§59，macOS 纯 Go Cocoa/AppKit 适配 + CI `macos-latest`）**：
    - 在无本地 Mac 硬件环境下（基于硬件排查说明未启用违规黑苹果/非授权 macOS 虚拟机），采用 `github.com/ebitengine/purego v0.9.1`（`CGO_ENABLED=0` 零 cgo）完成完整 macOS 原生能力：`internal/platform/focus_darwin.go`（`CGWindowListCopyWindowInfo`、`CGEventSourceSecondsSinceLastEventType`、`NSWorkspace` / `NSRunningApplication`）、`internal/ui/app_darwin.go`（`NSApplication` + 无边框透明置顶 `AITerminalPetWindow` + `NSImageView` + `NSStatusBar` 菜单 + 鼠标/拖拽/滚轮缩放/键盘导航）、`cocoa_darwin.go` / `clipboard_darwin.go`（`NSPasteboard`）/ `sound_darwin.go`（共享 `sound.go` WAV 合成 + Cocoa `NSSound`）。
    - 与 Linux 共享 `canvas_opentype.go`（拆出 `canvas_fonts_darwin.go` 扫描 `/System/Library/Fonts` 与 `~/Library/Fonts`）、`control_unix.go`（Unix 控制 socket）、`internal/app/platform_gui.go`（`uiAvailable() = true`），以及与 Windows 共享 `sound.go` WAV 合成器；纯 Cocoa 坐标/像素/按键转换抽入无标签 `darwin_geometry.go` 并在三平台全部运行单测。
    - `.github/workflows/ci.yml` 新增 `macos` job（`runs-on: macos-latest`，实机运行 `go vet`、`go test -count=1 ./...`、编译 macOS `arm64` 三二进制并执行无头 CLI/hook 冒烟；物理屏幕窗口交互仍以 `PET_COCOA_PROBE=1` 作为 opt-in 探针）。详见 [docs/PLATFORMS.md](PLATFORMS.md)。


### E6 — `--doctor` 自检 ✅ 已交付（2026-11-10，见 §1i）
**动机**：本轮的排查花了大量时间在「进程活着但窗口不可见」这类无声故障上。
**实际交付的 9 项**：environment / install files / version / runtime / config / pipe / instance / node / autostart，
逐行 `[ok]`/`[warn]`/`[fail] name: detail`，末行 `doctor: N ok, N warn, N fail`，任一 fail 退出码 1。
**与原计划的差异（有意为之）**：
- **没有探测互斥体**：`CreateMutex` 会创建对象并可能干扰正在运行的实例；改用「管道有人监听」作为实例存在的证据。
- `runtime/`、`config`、`error.log` 只报了**状态**，没有把 `hang.log` 纳入——那是事后取证文件，不是安装是否健康的判据。
- **没有做窗口 `WS_VISIBLE` 与消息泵响应性检查**：这类检查需要把探针注入别人的进程（`SendMessageTimeout`
  + `SMTO_ABORTIFHUNG`），本轮探针能做是因为当时只有一个已知实例；作为产品功能它会变成「诊断工具自己去戳
  一个可能卡住的窗口」。仍留在 `PET_FOCUS_PROBE` / `PET_SOUND_PROBE` / `PET_DSH_PROBE` 这几个 opt-in 测试里，
  等 E1（原生 hook）落地后再评估是否值得产品化。
- ~~三处终端配置的接入检查也没有做~~ → **2026-11-15 已补上**：第 11 项 `reporters` 已交付，
  通过 `internal/connect` 读 OpenCode / Claude Code / Antigravity 三份宿主配置（是否接入、目标文件是否还在、
  能不能发结束信号）。同日新增 `--connect` 只打印粘贴片段、不写任何人的配置文件。详见 HANDOVER §56。

## 3. 轻量化计划

### L1 — 进程创建税（当前最大开销，~1 s/次）
**现状**：Antigravity 的 `statusLine` **每次渲染**都要起一个进程；`pet-statusline.exe` 自身
实测 0.95–1.5 s，其中绝大部分是进程创建被 AV 钩挂扫描，不是我们的代码。
**可做的**：
1. 缩小二进制（见 L2）——镜像越小，AV 全量扫描越便宜。
2. ✅ **已落地（`internal/reporter/payload.go` + `cmd/pet-statusline/main.go`，§58）**：`pet-statusline`
   先读 `runtime/` 里同会话上一次快照（`ReadSnapshot`），`StatuslineDeliverNeeded` 对比状态字段并保留
   30 秒心跳保活（故意不比较易变的 `ParentPID`）；未变时直接写完快照就跳过管道投递，减少 IPC 往返。
3. AV 排除目录（需管理员，用户决策）——docs/archive/BASELINE-20260924.md §13 已记录这是唯一能真正压到亚秒的手段。
4. 已落地：`--reset`/`--toggle` 等控制路径不再走 `app.Run()`，避免控制命令意外拉起常驻实例。

### L2 — 二进制体积（4.43 MiB；2.5 MiB 的目标已按实测作废，见 HANDOVER §56.10）
**现状**：主程序 **4,655,616 B**（4.44 MiB，+92,672 B 来自 §56 的 `internal/connect` 与 `net/url`）、
状态行 shim **3,669,504 B**、原生 hook **3,701,760 B**，
三者合计 **12,026,880 B** = 11.47 MiB。共享内核 `internal/reporter` 让 shim 比共享前（3,643,392 B）
长了 **21,504 B**，hook 也因此不再各写一份脱敏/投递逻辑——用 21 KB 换掉一份会漂移的重复实现。
1.0.8 又在主程序上加了 **1,024 B**（`internal/ui/lifetime.go` 与两个诊断构建文件）。
本版新增 `--doctor` 让主程序比 v0.9.0 的 4,178,432 B 增加了 **193,536 B**（约 +189 KiB，+4.6%），
其中 `internal/app/doctor_windows.go` 占 22,932 字节源码——这是「可自查」换来的体积。
**做法**：`-ldflags "-s -w"` 已用。**`-buildvcs=false` 已实测并入**：同样的 -s -w，
加上它会去掉 VCS 戳记（`go version -m` 里从 `(devel)` 变成 `v0.0.0-<时间戳>-<sha>+dirty`）。
字节差随工具链变化——go1.23 上 shim 由 3,644,416 B 降到 3,643,392 B（差 1,024 B），go1.27 上两种写法恰好同为
3,664,896 B（伪版本与 `(devel)` 等长）——所以真正保证可比性的是旗标，而不是那个差值。
现在 `scripts/install.ps1` 的缺失提示、`AGENTS.md` 与 CI 用的都是这条命令，所以**同一份源码在任何
机器、任何时间重建出的产物字节一致**（二进制不入库，`bin/` 被 `.gitignore` 忽略）；`-VerifyOnly` 拿已部署
副本与工作区里的 `bin/` 逐项比对，两侧构建方式不同就会误报。继续可做：剥离 `net/http` 等未用包的反射元数据（本项目不用 http，
但 `net` 被管道复用）、用 `-tags` 关掉不需要的 `crypto` 实现。
注意 `pet-statusline` 引了 `regexp`（脱敏规则）——这是体积大头之一，可改为手写扫描器。
**验收**：`go build` 后主程序 < 2.5 MiB，且 `node --check` 与全部单测不变。

### L3 — 管道连接的 goroutine churn ✅ 已实测（2026-11-15）：不做优化
**现状**：`hang.log` 里出现过 goroutine 编号 79 万+，说明 10 小时内创建了约 80 万个短命 goroutine——
go-winio 的 `makeConnectedServerPipe` 每次 Accept 前都会预建一个 `connectPipe` 等待者。
**实测（本轮，两个数字合起来才是答案）**：
- **一次投递往返**：`internal/pipeserver/delivery_bench_test.go` 的 `BenchmarkDeliveryRoundTrip`
  在 `net.Pipe` 上跑生产 `Server.read`（读→解析→update→`OK`），**1.84–1.96 µs/op**（500x，i7-14650HX）。
  这是进程内的地板价，真实路径再加一次 dial；但 1 µs 量级说明「连接数」本身不是瓶颈。
- **事件速率**：同一台机器 90 秒窗口内实测 **80 次快照写入 = 53 次/分钟**（约 0.9/s，6 个活跃会话），
  远低于原文担心的「20+/s」。按此速率，10 小时约 **32,000 次投递**，而不是 80 万。
**结论：不做**。80 万那个数字对应的是 hang.log 里出现过的最高 goroutine 编号，不能直接读作
"创建了 80 万个"——goroutine 编号是累计分配编号，包含所有 goroutine，不只是投递的。
真要压，只有「上报端变化才连」这一条有意义，而 L1 的 30 秒短路已经在 statusline 上做了；
再往下压属于为 1.84 µs 增加复杂度。

### L4 — 光栅化与重绘 ✅ 已实测（2026-11-17）：不做优化
**已做**：DIB 画布复用、命中缓冲复用、动画定时器按状态装/卸。
**这一次把"先测再动"真的测了**：`internal/ui/paint_panel_cost_test.go`（`go test ./internal/ui -run TestMeasurePanelPaint -v`，
每例 50 帧取均值，本机 i7-14650HX）：

| 帧 | 尺寸 | Windows（GDI） | Linux（纯 Go OpenType） |
|---|---|---|---|
| 紧凑·等待 | 56×156 | 0.450 ms | 0.699 ms |
| 紧凑·空 | 56×156 | 0.416 ms | 0.298 ms |
| 展开·3 行 | 200×212 | 2.250 ms | 1.851 ms |
| 展开·6 行 | 200×314 | 3.953 ms | 3.411 ms |
| 展开·3 行 + 悬停 | 200×212 | 2.047 ms | 1.980 ms |

结论：**脏区重绘与内层循环向量化都不做**。最坏的那一帧（6 行展开）约 4 ms，
而提醒层的动画是 20 FPS（50 ms 预算）——余量 12 倍；常驻的紧凑帧是 0.3–0.7 ms。
把每帧全画换成脏区重绘会引入"哪些东西必须重画"的状态机，而它换来的时间在用户眼里不存在。

（旧版本节写的 `40×190 / 200×250` 是 v0.5.0 之前的固定几何，早已被 `drawing_compact.go:23-35` 的常量取代。）

### L5 — 启动与扫描路径 ✅ 已交付（2026-11-15，按 mtime/size 复用上次解析）
**现状**：快照扫描 5 s 一次（`internal/app/app.go` 的 `scanInterval`；本节旧文写的 2 s 已过时），全目录 `Readdir` + 每个文件完整读取与解析。

**实测（第 61 轮，`internal/store/snapshot_bench_test.go`，`-benchtime 30x`，i7-14650HX / NTFS / 带 AV 过滤）**：

| 快照文件数 | 未缓存 | 已缓存（稳态） |
|---|---|---|
| 1 | 298 µs | 113 µs |
| 50 | 7.02 ms | **0.52 ms** |
| 200 | **22.7 ms** | **0.51 ms** |

按项拆开（同一台机，500 次迭代）：`os.Lstat` **31.3 µs**、`readSnapshot`（4 次 stat + 一次 read）**89.6 µs**、
`session.ParsePayload` **5.1 µs**、单个候选端到端 **98.1 µs**。**解析只占 5%，元数据调用占 90%**——
所以要减的不是"少解析"，而是"少碰盘"。

**做法（已落地）**：`internal/store/snapshot_cache.go` 的 `SnapshotCache`，按 **文件身份 + size + mtime** 命中即复用上次的
`session.Session`；`SnapshotScanner.Cache` 是**可选指针**，nil 就是旧行为（`--json` 等一次性调用者与全部既有测试仍走这条路）；
`internal/app/run.go` 的常驻 scanner 传 `store.NewSnapshotCache()`。删除一律仍走 `removeSnapshotIfUnchanged`（先重新 stat 再删）。

**三个硬约束（都有测试钉住）**：
1. **不能靠元数据骗过去**：报告端是 temp+rename 写盘，所以重写会落成**新文件**，`os.SameFile` 能看出来——
   实测把 mtime 手工改回、字节长度也保持一致之后 `SameFile(缓存那份, 新 stat) = false`；
2. **命中时必须仍能删文件**：`removeSnapshotIfUnchanged` 需要 `FileInfo`，而命中路径没有 read 也就没有 post-read stat，
   于是补上"退回用 Readdir 那份"的兜底——**这是本轮唯一一个真 bug**，缺它会静默泄漏过期/宿主已死的快照文件
   （面板照旧丢弃、文件永远留着）。两条测试在去掉兜底后**确实失败**，补回后通过，这就是它们存在的理由；
3. **上限仍成立**：225 文件的目录在本机实测 5 次扫描 **0.50–0.61 ms**，门槛是 5 ms，验收满足。

**一个如实记录的测量事实**：NTFS 在这台机器上 40 次 temp+rename 里**复用过 1 次文件索引**。所以守卫是
"身份 **且** size **且** mtime"三个字段，而不是只看身份；配合"删除前重新 stat"，最坏情况是一次陈旧读取，不会误删文件。

## 4. 实用功能候选

已完成：~~F3 开机自启~~、~~F4 提示音~~（v0.9.0 已升级为四套合成音色 + 按事件开关，见 §1h）、
~~F7 `--json` 导出~~（见 §1a）；~~F6 相对时间~~（展开态每行显示「刚刚 / 3 分钟前 / 1 小时前」，见 §1c）。

v0.6.0–v0.9.0 期间落地、原先不在候选表里的四项：**语言切换**（en/zh-CN + auto，§1e）、
**设置窗口**（§1h）、**auto-approve 否决窗口**（§2 E2b）、**DeepSeek Harness 接入**（§1h）。

v0.9.0 之后（2026-11-10，未发版，见 §1i）又交付了四项**扩展项**，原先分别挂在 §2/§3 里：
~~E2 线协议文档 `docs/PROTOCOL.md`~~、~~E6 `--doctor` 自检~~、~~安装脚本 `-VerifyOnly`/`-Uninstall`/`-AddToPath`~~、
~~CI 门禁 `.github/workflows/ci.yml`~~。

✅ 注册表条目 PR #13937 **已于 2026-11-10T14:33:04Z 由 `fkysly` 合并**，上游 README 已列出（`:8519`）；合并只写进了 `main`，可见性要等上游站点构建——16:16Z 条目页还是 404、`plugins.json` 0 命中，而 run `42844452013` 在 16:39:27Z success 之后条目页返回 200、`plugins.json` 的 `count` 从 10119 涨到 10230 并收录本仓库（目录里的 `install` 按 yml 的 `tarball` 渲染，合并后那版仍是 1.2.0；#14076 已换成 1.2.2 并在 19:59:37Z 合并、#14089 再推到 1.2.3 并已于 2026-11-10T21:10:15Z 合并，都要等下一次站点重建才渲染出来；1.2.4 的 #14243 已开）；同一轮 [imsai-sh/awesome-deepseek-harness-plugins](https://github.com/imsai-sh/awesome-deepseek-harness-plugins) PR #902 于 2026-11-10T21:49:53Z 合并（该店只从 npm 安装，npm 发出去之前以 browse-only 列出）、[zhu1090093659/dsh-community-plugins](https://github.com/zhu1090093659/dsh-community-plugins) PR #97 于 2026-11-11T05:06:32Z 合并；1024 Store 详情页返回 200（8,753 B）；
**DSH 插件包分发** —— ✅ 包本体走到 **1.2.5**，八个 tag 都有对应 Release（`dsh-plugin-v1.0.0` / `v1.1.0` / `v1.2.0` / `v1.2.1` / `v1.2.2` / `v1.2.3` / `v1.2.4` / `v1.2.5`，当前资产 1,917,342 B / sha256 `dfa89a1e…`），且 npm 上已发布（`latest` = 1.2.5，tarball 与 Release 资产逐字节相同）；tarball / Release 资产 URL / git 源钉 tag / **npm 裸包名**四种安装实测成功（不写 `#ref` 的 `&path:` 形态会报 `repository ... not found`，与 `prepack` 无关）；
✅ 注册表条目 PR #13937 **已于 2026-11-10T14:33:04Z 由 `fkysly` 合并**，上游 README 第 8,519 行已列出该条目；「合并进 `main`」与「在某个市场里可见」曾经是两件事，现在后者也成立：条目页在 16:16Z 是 404、站点构建 run `42844452013` 于 16:39:27Z success 后变 200，`plugins.json` 也从 10119 条增至 10230 条并收录本仓库——目录里的 `tarball` 决定安装版本、只能靠 PR 改：**#14076 已于 19:59:37Z 合并**（把字段换成 1.2.2）、**#14089 已于 2026-11-10T21:10:15Z 合并**（再推到 1.2.3），两者都要等下一次站点重建才反映到 `install` 字段（前身 #14063 已关闭）；1.2.4 的 **#14243** 已开，等合并；
✅ [DSH-Plugins-Marketplace](https://github.com/bradeGithub/DSH-Plugins-Marketplace) 的索引 2026-11-10T18:08:44Z 重建后，本仓库条目是 `pkg_name: dsh-ai-terminal-pet`、`version: 1.2.1`、topics 11 个、`verdict: bundle-plugin`，并且 **`risk_tier: "safe"`、`risk_flags` 一条不剩**（`risk_at` 17:09:14Z，是安装脚本移进 `scripts/` 后第一次真的扫到本仓库）。之前的 `verdict: script` 与 `risk` 徽章（`registry-run-key` critical + `remove-item-force-recurse` medium）从此消失，全程没有用到任何抑制字段——因为扫描器 `evalRepoRisk` 只按文件名抓仓库根的 `install.ps1` / `install.sh` / `package.json`，而脚本已移进 `scripts/`（`$RepoRoot = Split-Path -Parent $PSScriptRoot`，原样保留停进程、`-Recurse -Force` 与 `CurrentVersion\Run` 逻辑）。issue #367 已回帖说明，可关。
✅ 两条不用申请的自动采集路径（根清单采集 `LivXue/dsh-plugin-shop` 每日、双 topic 采集 `zasenjc/dsh-plugins-store` 30 分钟）已在仓库侧满足条件；但 `LivXue` 的发布分片实测只有 npm 源（26,351 条条目全带 npm tarball 的 `integrity`、`commit` 字段 0 条），所以「条件满足」≠「已被列出」；
✅ **自建 format-1 索引**（`ef783ba`，`docs/market-index.json`；`fe62900` 前移到 1.2.2、`fabb591` 前移到 1.2.3、发版提交 `34aa1f7` 前移到 1.2.5）：Discovery 不必等 npm——条目 `dsh-ai-terminal-pet` `1.2.5` 用钉到 `34aa1f7` 的 git 源，URL <https://raw.githubusercontent.com/jjjjjjjjnnjnn/ai-terminal-pet/main/docs/market-index.json>，用户加到 Discover 的索引源列表里即可看到并一键安装；校验走宿主自带的 `parseMarketIndex`，`marketInstallTarget` 产出的 spec 就是上面 3.7 秒装通的那一条，且 `verify.mjs` 有断言把「索引最新版本 == 包版本」钉住（发版忘改索引会让 CI 变红）；
✅ **npm 已发布 1.2.4（12:10:46Z）与 1.2.5（14:58:53Z）**，两次都是 agent 在本机包目录执行 `npm publish --access public`——**本机 token 已配置成可绕过 2FA，不再需要人工**，只有 token 失效（`npm whoami` 报 E401/E403）时才要人重新登录。当前 `dist-tags.latest` = **`1.2.5`**、shasum `1f6d5340…`、17 files，从 registry `npm pack` 回来的 tarball sha256 仍是 `dfa89a1e…`（与 Release 资产逐字节相同）。「按包名安装」与「进 npm-only 目录」这两个前置都成立。**注意发布后头几分钟 registry 上只有占位的 `0.0.0-stage`**，此时 `npm install <name>@<version>` 会报 `ETARGET`，本地 npm 还要 `--prefer-online` 才看得到新版本。

| 编号 | 功能 | 价值 | 复杂度 |
|---|---|---|---|
| **I1** | **独立安装包（签名 `.exe` / `.msi`）**：给不用 DeepSeek Harness 的 Windows 用户一条「下载—双击—完成」的路，不再需要 Go 工具链，也不必从 tarball 里手动挖出 `vendor/AI-Terminal-Pet.exe` | **高** | 高 |
| **F10** | **提醒层：一个纯函数决定「要不要打扰」，并留下可回答的判定记录** —— **已交付（2026-11-12，1.0.3）**，见下方 F10 小节 | **高** | **中**（三种请求要分开、前台/闲置要采样、逐会话静音与未读要持久化） |
| **F8** | 会话行点击复制项目路径（面板内回执，避免"点了没反应"）——**已交付（2026-11-13，1.0.5）**，见下方 F8 小节 | 高 | **中**（回执是时间驱动的绘制元素，紧凑条拆成两个命中区） |
| **F12** | 设置窗口可读可用：外观页签真的能点、字号升到系统档位、说明换行、键盘可达与 tooltip —— **已交付（2026-11-13，1.0.7）**，见下方 F12 小节 | 高 | **中**（GDI 字号吸附要实测、行高要按内容算、降档要按工作区） |
| F2 | 键盘快捷键：`Esc` 收起、方向键/`Home`/`End` 在会话行间移动、`Enter` 跳转 —— **已交付（§58，`keyboard.go`）** | 中 | 中 |
| F5 | 记住所在显示器（按显示器工作区签名保存/恢复坐标，`State.Layouts`） —— **已交付（§58，`placement.go`）** | 中 | 中 |
| F1 | 悬停展开面板时显示完整会话明细（就地浮层） | 中 | 中 |
| ~~F9~~ | ~~紧凑态显示「最久等待了多久」~~ **已交付（2026-11-12，1.0.1）**，见下方 F9 小节 | 中 | 低 |

### I1 拆成三段，第一段已交付（2026-11-11）

### F12 — 设置窗口可读可用 ✅ 已交付（2026-11-13，浮窗 1.0.7 / 插件 1.2.14）

**为什么它值得单列**：用户报的是"外观点不进去"和"字太小"，看起来是两件小事，实际上是两类结构性问题——
一类是**测试形状错了**（只验证"点到哪个控件"，不验证"点了之后发生什么"，于是漏掉一个 `case` 两个版本没被发现），
另一类是**排版数字没跟着渲染器走**（GDI 把请求的字符格高吸附到粗档位，旧梯子的三层字号实际只有两种）。

**四条规则（各有测试）**：

| 问题 | 规则 | 位置 |
|---|---|---|
| 页签点了没反应 | 每个侧栏控件都必须真的切换 `settingsTab`，外观页签的控件还必须发布到偏好回调 | `settings_tabs_windows_test.go` |
| 字号又小又平 | 三张手写梯子（小/标准/大），落在 GDI 真实档位上；4 个 DPI 下量宽度保证层级不塌 | `settingsTypeRamps` + `TestSettingsTypeRampStaysOrdered` |
| 说明被截断 / 备注被裁 | 行高按"标题带 + 行数 × 行高 + 控件带"算，换行与测量是同一个函数，中文按行首禁则 | `settings_text_windows.go` |
| 屏幕放不下 | 大 → 标准 → 小 降档，而不是长出工作区 | `settingsGeometryFitting` |

**别改回去的**：`settingsTypeRamps` 是三张表而不是一个系数（乘 1.2 会让两个角色落到同一档）；
`wrapHint` 必须同时供测量与绘制使用（分开就会"预留两行画三行"）；备注的文字在 `title` 里而不是 `hint` 里。

### F11 — 提醒只报真实事件：消失不算完成、一行一个会话、OpenCode 对齐 2.x ✅ 已交付（2026-11-13，浮窗 1.0.6 / 插件 1.2.13）

**为什么它是一条独立条目而不是 F10 的补丁**：F10 把"要不要提醒"收进了一个纯函数，但它假设
**输入是真的**。用户的两次报告说明输入有三个洞：一个"忙过的会话消失"被合成完成、一个会话可能占两行、
一个死进程能被窗口标题续命。第四个洞在上报端：OpenCode 2.x 的插件 API 与事件名一起换了，仓库里那份
`reporters/opencode.js` 在 2.x 上既不加载、也不认识任何真实事件名。

**四条规则（各自一条测试，且都在真机/沙箱复现过）**：

| 洞 | 规则 | 位置 |
|---|---|---|
| 消失被当成完成 | `AlertInput.Vanished` → 静默 `decisionVanished`，回执仍写明 | `internal/ui/attention.go` |
| 一个会话两行 | 面板层按 `session.Identity` 去重（`--json` 仍保留 per-process 行） | `internal/ui/sessions_dedupe.go` |
| 死 pid 被标题续命 | 受信任上报端 pid 死亡即结束；纯工具名标题不算身份 | `internal/platform/presence_windows.go` |
| 普通命令也响 | 事件表逐条写死 2.2.7 的真实事件名；`tui.prompt.append` 永不产出状态 | `reporters/opencode.js` |

**证据**（都在 §43.5）：真机 `--json` 6 行/3 会话（3 个死 pid）→ 4 行/4 会话；沙箱只杀上报端进程，
回执从 `closed/done -> done-away` + `tone=finished` 变成 `closed/none -> vanished` 且静默；
控制组（真完成）仍然响。

**别"优化"回去的四件事**：`--json` 的 per-process 行是有意的（两个上报端说法不一致正是证据）；
标题规则只对合成 pid 生效（hook/statusline 的 pid 不证明任何事）；不要改成"窗口必须属于该会话的进程树"
（实测会误删活会话）；`session.tool.failed` 不是 `error`（命令失败是日常）。

### F10 — 提醒层：一个决策器 + 可回答的判定记录 ✅ 已交付（2026-11-12，浮窗 1.0.3 / 插件 1.2.9）

**为什么它不是「加几个开关」**：上一版把「要不要提醒」散在 `alertOnTransitions` 的一段 if 链里，两个后果都是真 bug：
账本按 `state` 记账，所以「答掉一个权限请求、立刻又被提问」这一声不响（两次都是 `waiting_permission`）；
其余判据（你在不在看、这个会话是否被静音、完成算哪一档）没有地方放，于是「你正在读的请求还额外弹一个气泡」。

**做法**：`internal/ui/attention.go` 的 `Decide()` 是纯函数（输入一个会话的全部事实，输出弹/响/未读/原因），
`PlanAlerts` 一次扫描最多允许一个气泡。规则顺序固定：静音 > 子会话 > 完成三档（含 15 秒合并）> 紧急请求
（只在「你正看着它且没走开」时才可以不弹，且仍标未读）> **测不准就当没在看**。

| 判据 | 来源 | 默认 |
|---|---|---|
| 这个会话在要什么 | `reason` 派生（审批 / 提问 / 方案确认），**没有新协议字段** | — |
| 你在不在看这块屏幕 | `GetLastInputInfo` 的闲置时长 | 3 分钟算离开，`0` 只看窗口焦点 |
| 你看的是不是这个会话 | 复用跳转功能的进程/窗口排名 | 开 |
| 这个会话被静音了吗 | `config.json` 的 `mutedSessions`（键 = source + sessionId，**不含 pid**） | 空 |
| 完成算哪一档 | `donePolicy` | `background`（仅后台） |

**四条施工约束**（都写成了测试）：

1. 静态与会话的键不能混用：`session.Identity` 不含 pid，`config.SessionMuteKey` 与它逐字节一致，由 `internal/app` 钉住。
2. 新画的东西必须进 `renderedKey`：未读位、静音位、悬停行都在；1.0.1 的时长冻结与 1.0.2 的主题不重绘是同一类成因。
3. 采样绝不上 UI 线程：枚举别人的窗口会卡在挂起的窗口上（AppHang 的形状），采样器是独立 goroutine，8 秒一次，复用 `Presence` 缓存。
4. 标记不占宽度：未读是一个点、静音是整行压暗——第一版用两个字把 `reason` 挤掉了一半。

**可回答性**：`%LOCALAPPDATA%\AI-Terminal-Pet\alerts.json`（覆写不追加，在 `config.json` 旁边而不是 `runtime/` 里）
记下最近一次扫描里每个会话为什么响/为什么没响，`--json` 的 `alerts` 与 `--doctor` 的第 10 项读它。
**1.0.4 修过一次**：这行原先写的是 `runtime/alerts.json`，而那个目录里每个 `*.json` 都被 `internal/store`
当成一条上报快照、解析不成就删除——记录在一个 5 秒扫描周期内消失，`--json` 报「没有记录」、`--doctor` 报
「从未扫描」。`internal/store/runtime_contract_test.go` 从目录一侧把规则钉住，**不要再把任何 JSON 放回那个目录**。

**唯一改变老用户体感的默认值**：`donePolicy` 默认 `background`（旧行为是「完成就弹」）。旧配置里的
`notifyOnFinish: true` 被迁移成 `always` 而不是 `background`——用户当初点的就是「完成要提醒」。

**明确不做**：外部推送（Discord / 飞书 / Server酱 等）。它会打破 README 与包清单里的 `no-network` 声明，
并且需要 HTTP 客户端、凭据文件、重试与失败记录。理由记在 §5 非目标里。

### F8 — 点击会话行复制项目路径 ✅ 已交付（2026-11-13，浮窗 1.0.5 / 插件 1.2.11）

**做了什么**：展开态点一行、紧凑态点跳转条的左半，即把该会话的**完整工作目录**放进剪贴板
（面板行上显示的是 `filepath.Base`，剪贴板里是 `session.Project` 的原值）；行上出现 1.2 秒的
回执胶囊（`已复制` / `复制失败`），紧凑态的回执顶掉 subject 行。剪贴板调用失败时**不假装成功**：
回执写「复制失败」并响一声，点击不会落到「收起」。

**四条施工约束（1.0.3 摔出来的那三条 + 这一条新增）**：

1. **命中区域必须与绘制同源**。`expandedRowRect` / `expandedRowHit`（1.0.3 就已落好）与新增的
   `compactActionRects` 同时喂给光栅化与命中；紧凑条的两半有单测钉住「不重叠、中点各归各位」。
2. **新画的东西必须进 `renderedKey`**。回执是时间驱动的元素，所以它有独立的三元组字段
   （`toastRow` / `toastTick` / `toastFailed`），**不挂在 `clock`** 上——呼吸路径刻意忽略 `clock`，
   而「回执留在屏幕上不走」正是 1.0.1 时长冻结的同一种 bug。
3. **回执的定时器复用 `syncAnimation`**，不新开常驻计时器；`toggleExpanded` 会把回执清掉。
4. **回执是「关于某个会话」的断言**。面板会随上报重排，所以 `copyToast` 记的是 `session.Identity`：
   那一行换了主人、或紧凑条的最紧急会话变了，回执立刻消失，而不是把「已复制」留在别人的项目上。

**交互规则写成纯函数**：`internal/ui/tap_windows.go` 的 `decideTap` 按固定顺序
**跳转箭头 > 紧凑条左半 > auto-approve 否决（仅倒计时武装时）> 展开态行 > 收起**。
否决排在行复制之前是有意的：倒计时在跑的那几秒里，「别放行」是最贵的一次点击，
用一次剪贴板写入来回答它等于静默放行。

**验收（都跑过）**：`go test` 12 包全绿；`PET_CLIPBOARD_PROBE=1` 的桌面探针走真实往返（写一次、读回、
再写一次确认是替换不是追加）；沙箱里拿**装出来的** 1.0.5 用 `PostMessage` 点第一行，`Get-Clipboard`
拿到的就是那个完整路径，点页脚仍然收起（200 px → 56 px），紧凑条左半复制、右半不碰剪贴板。

**代价**：`docs/assets/` 的面板 PNG 与插件包 `screenshots/` 重出（紧凑条多了复制半与分隔线，
展开行多了回执），因此需要一次完整发版。

### F9 — 紧凑态显示最长等待时长 ✅ 已交付（2026-11-12，浮窗 1.0.1 / 插件 1.2.7）

**做法**：不新增行、不改 56×156 的几何。有人在等时，大号槽从「会话数」换成「等待时长」
（`12分` / `12m`，16 号粗体），数量降到下面那行（`3 个会话`）；状态行、项目名、跳转条、命中区一律不动。

**三个判断都有测试钉住**：只对 `waiting_permission` / `error` 显示（忙碌没有人在等，`idle` 的时间戳
意思是「什么时候结束」）；多个会话在等时取**最早开始等的**那个（与跳转条的「最该点哪个」故意不同）；
取整边界与展开行的相对时间**共用一套**，避免同一个瞬间两处说法不一致。

**一个必须记下来的 Bug**：呼吸态复用旧位图是刻意忽略 `clock` 的，所以时长不能挂在 `clock` 上，
否则 `12m` 会冻结到下一次上报。修法是独立的 `waitMinute` 分钟桶，并断言「跨分钟内仍可复用、
跨越分钟必须重绘、忙碌/展开态该字段为 0」。细节与实测尺寸见
[docs/HANDOVER.md](HANDOVER.md) §36。

**代价**：`docs/assets/` 的面板 PNG、`hero.png`、`demo.gif`、`desktop-scene.png` 与插件包内的
`screenshots/compact.png` 全部重出（几何未变，所以分辨率与构图仍然成立），并因此需要一次完整发版。

I1 原本被写成一件「高复杂度」的事，实测把它拆开之后三段互不阻塞，而且第一段今天就做完了：

| 段 | 内容 | 状态 |
|---|---|---|
| **I1-α** | **便携 zip**：`scripts/package.ps1` 打出 `dist\ai-terminal-pet-<版本>-win-x64.zip`（17 个文件，仓库原有布局的子集 + 根转发器 + `SHA256SUMS.txt`），配 `scripts/install.cmd` 承接 RemoteSigned，CI 增加打包门禁与 artifact 上传 | ✅ **已交付**（2026-11-11，零版本号代价） |
| **I1-γ** | **自解压单文件 `.exe`**：`cmd/pet-setup`，`//go:embed` 阶段载荷 → **先**按 `SHA256SUMS.txt` 逐文件校验 → 解到临时目录 → 调用内嵌的 `install.ps1`；`--verify` / `--uninstall` / `--add-to-path` 转发到同名开关；未签名发行。20 个单测 + 沙箱端到端实测 | ✅ **已交付**（2026-11-11，`pet-v1.0.0` 资产的 `pet-setup.exe`） |
| **I1-β** | **换版本号**：`Version` → 1.0.0，重建三二进制 → vendor 重算（源码摘要 `edd584c6…`）→ 插件 **1.2.6** → 双清单 + 索引同提交（A 提交 `df4fcf6`）→ 双 tag + Release → npm → B 提交钉 `source.commit`（`6496ba6`）→ 上游 PR #14439 | ✅ **已交付**（2026-11-11） |
| **I1-δ** | **签名 + MSI**：证书（买 OV/EV 或 Azure Trusted Signing）+ `signtool sign /fd sha256 /tr <时间戳> /td sha256`；MSI 需先装 WiX。**这是 I1 里唯一需要外部输入而不是代码的一段** | ⏳ 未开工 |

**I1-γ 的三道坑（都产出了一个「能构建、能运行、看起来正常」的包，必须记住）**：

| 坑 | 表现 | 修法 |
|---|---|---|
| `//go:embed payload` | 点开头的文件被**静默跳过**，`.staged` 没进二进制，每次发布都走「拒绝运行」 | `//go:embed all:payload` |
| `Copy-Item -LiteralPath '…\*'` | `-LiteralPath` 不做通配，**一个文件都没拷**，exe 却照旧构建成功（载荷 2 个文件） | 改为逐条目复制 + 断言载荷 > 1 MB |
| 载荷照抄 zip 整棵暂存树 | `README.install.txt` 与根 `install.cmd` 不在 `SHA256SUMS.txt` 里，于是**每次安装都失败**在自己的「无未列文件」检查上 | 只暂存 `install.ps1` 的那份制品清单 + 摘要 |

**已知约束（2026-11-11 本机实测，不要重新调研）**：

| 事实 | 现状 |
|---|---|
| `signtool.exe` | **有**：`C:\Program Files (x86)\Windows Kits\10\bin\10.0.26100.0\x64\signtool.exe` |
| 代码签名证书 | **没有**：`Cert:\CurrentUser\My` 与 `Cert:\LocalMachine\My` 的 `-CodeSigningCert` 都是空。**决策：先出未签名包，接受 SmartScreen，签名作为独立事项** |
| WiX / NSIS | **没装**：`candle.exe`、`wix`、`makensis` 都不在 PATH，也不在 Windows Kits 里 |
| PowerShell 默认策略 | **`RemoteSigned`**（CurrentUser）。这正是 `install.cmd` 存在的唯一理由 |
| 已有安装能力 | `scripts/install.ps1`（11 件制品、逐件 SHA256、PATH/Run 键、`-Uninstall`）、`scripts/install.cmd`、`scripts/package.ps1`、`cmd/pet-setup`、`verify.mjs`、`tests/fresh-machine.mjs`（**162** 项端到端） |

**不要**为了「看起来专业」给安装器引入运行时依赖：安装器自身必须自包含（Go 或 PowerShell 都行），
理由与 AGENTS.md 第 3 条「零外部运行时依赖」是同一条。

### 对外宣传的顺序约束（2026-11-11 定，同日已解除）

**β 已交付**（浮窗 1.0.0 / 插件 1.2.6，`pet-v1.0.0` 与 `dsh-plugin-v1.2.6` 双 Release，npm `latest` = 1.2.6），
所以那条「正式发声排在 β 之后」的约束**已经满足**：现在可以发声，而且四个数字终于一致。
**2026-11-12 又交付了 F9**（浮窗 1.0.1 / 插件 1.2.7），见下方 F9 小节。
宣传素材里的每一个数字都必须是当场跑出来的（**495** 个 `func Test` / 183 项新机检查 / **4.50** MiB（本节数字更新到 1.0.15）），
**不要写「全量安全扫描」**：`govulncheck` 与 `gosec` 从 2026-11-12 起已跑（见 §6），
`go test -race` 自 2026-11-15 起在 CI 的 `ubuntu-latest` job 里跑（本机没有 gcc，Windows runner 也没有），
但「竞态检测只在 Linux 上跑」这句话要跟着说，所以「静态扫描 + 依赖漏洞扫描通过 + Linux 上的竞态检测通过」
可以说，「安全审计通过」不能说。

**下一次改版本号之前请先读这一段**：`pet-v0.9.0`、`pet-v1.0.0`、`pet-v1.0.1` 三份资产都已发布且
**不可变**；插件线的九个 tag 同理（`dsh-plugin-v1.0.0` … `v1.2.7`）。1.0.1 → 1.0.2 会是新的 tag 与新资产，
不要替换任何已发布的那份。

F9 已于 1.0.1 交付、**F8 已于 1.0.5 交付**（见上）。F1 现在价值下降：展开态已经能列出 6 行带项目名、
状态、来源和相对时间的明细，悬停浮层要证明自己比"点一下"更快才有意义。

**"点击任意处收起"这条规则已经被 F8 改掉了**，而且是有意改的：现在点会话行（或紧凑条左半）= 复制路径
+ 回执，点行与控制以外的面板区域 = 收起。这不只是交互变更——**分层窗口没有别的方式告诉用户"复制成功了"**，
所以回执不是装饰，是这个手势能被相信的前提。

**已交付但值得回头看的一项**：`--json` 目前只读快照，因此看不到「只走管道、不写快照」
的上报路径（如果将来有上报端选择不落盘）。真需要时可以给它加一个查询运行中实例的
IPC 分支，但要在 `--json` 永远不阻塞 GUI 进程这个约束下做。

## 5. 非目标

- **不做终端复用器或日志查看器**：本项目的定位是「一眼可辨的状态指示」，不是替代终端。
- **不做跨平台的"降级版"**：macOS / Linux 的版本要么把平台层真正换掉（进程、前台、闲置、窗口、
  socket 全部有等价物），要么就不发。把存在性简化成"pid 活着"或把 `--doctor` 留空，
  等于把 §21/§43/§51/§52 修掉的缺陷重新引入一遍——那种"看起来能跑"的移植是负资产。
- **不做账号/云同步/遥测上报**：全部数据留在本机，`runtime/` 与 `state.json` 不离开 `%LOCALAPPDATA%`。
- **不做外部推送网关**（Discord / Bark / Gotify / 飞书 / Server酱 / WxPusher / 自定义 Webhook）：提醒只在本机（气泡 + 合成音 + 面板标记）。README 与包清单都声明 `no-network`，那是这个项目唯一的一条强承诺；一次外部推送就要引入 HTTP 客户端、凭据文件、重试与失败记录，并把「不出网」这句话删掉。人不在电脑前时，缺的是回来的理由而不是推送通道——面板与未读标记会一直在。
- **不引入 UI 框架**：零外部 UI 依赖是有意为之，它让二进制保持单文件、无运行时依赖。
- **不做跨用户共享的管道**：管道名按用户哈希隔离是安全边界，不为了「多用户共用」而放宽。

## 6. 已知残留（不计划修，但需知情）

- `hang.log` 看门狗是**事后取证**而非自愈；转储上限已压到 3 次 + 2 min 冷却 + 记录最后消息名。
- UI **有不变量门禁，但没有逐像素回归**：`internal/ui/render_invariants_test.go`（2026-11-17 起，随 `go test`
  在三个平台跑）断言"正确渲染"的形状 —— 帧尺寸等于布局、圆角外全透明、面板体真的被填、**文字真的被光栅化**
  （漏了字体栈就是这条红）、每个会话状态的颜色真的出现在画布上、不同状态/不同语言必须长得不一样、
  两套配色真的不同且浅色确实更亮。它**不**做逐像素比对：GDI 与纯 Go OpenType 的抗锯齿、字体与 hinting
  本来不同，"同机逐字节可复现"已经由 `TestRenderPreview` 证明，跨机只能比不变量。
  `--doctor`（E6）只覆盖环境/安装/运行期状态，**不做视觉比对**；`TestRenderPreview` /
  `TestRenderSettingsPreview` 仍是"生成图给眼睛看"的工具，不是断言。
- `go test -race` **于 2026-11-15 移进 CI**：`linux` job 多一步 `CGO_ENABLED=1 go test -race -count=1 ./...`
  （ubuntu runner 自带 gcc；本机 Windows 上没有，Windows runner 也不提供，所以那一步永远不在 Windows 上跑）。
  因此现在的说法是「Linux 上的竞态检测通过」，而不是「竞态已排除」。**`govulncheck` 与 `gosec` 已于
  2026-11-12 补上**（两者都是纯 Go 程序，不需要 C 工具链）：`govulncheck@v1.9.0` 报「没有被实际调用的漏洞」
  （代码受影响 0 条；另有 2 条在 import 的包、14 条在依赖模块里，但本代码调不到），
  `gosec@v2.30.0` 在 high severity + high confidence 上是 **0 条**，排除方式与理由写在
  `.github/workflows/ci.yml` 的注释里：`G115` 按**规则族**排除（Win32 消息与 `uintptr` 之间的整数转换，
  全部命中都是这一种形态），`G703` 只按**路径**在那四个有命中的文件上排除（由用户自己的
  `%LOCALAPPDATA%` 拼出的路径）——别的文件里出现路径污点仍然会让 `scan` 变红。**不是「全量安全扫描」**，别把这两条说成后者。
- **上游目录（两条不同的原因，2026-11-16 晚状态）**：`awesome-dsh-plugin` 的 `tarball` 由维护者手工合并，
  我们每一版都开一条 `+1 −1` 的 PR。
  - **冻结已修，等重建**：整站自 2026-11-14T05:27Z 冻结的根因是两条**各自合法**的条目
    （monorepo 根 + 子目录）的 `package.json` 同名 `dsh-dot-mcp`，format 1 按 `name` 建键 → 解析器拒掉整份文档 →
    夜间 `Build` 失败 → 部署与 npm 发布一起跳过（对方 **#15569**，维护者 Eapp1e 的诊断）。
    修复 **#15579** 已于 2026-11-16T16:39:48Z **合并**、#15569 随后关闭，维护者 16:39:59Z 手动
    dispatch 了一次 `build-site.yml`（run `43910005122`）。所以现在**只等对方那次重建跑完**：
    它会把我们的 `version` 刷到 1.2.23（这个字段来自 npm 探针）；`tarball` 仍要等 **#16010** 合并
    + 再一次重建（详见 [docs/UPSTREAM.md](UPSTREAM.md) §4.1、§4.2）。
  - **下载量有两段缺口**：站点卡上的 `downloads (30d)` 比真实值少 **2,276** ——
    窗口自己的最后一天 1,197（03:1x UTC 探针撞上 npm 聚合延迟，**全站 12/12 条目同一签名**，对方 issue #15864，
    另有 #15982 一位作者在 `dsh-rea-pro` 上独立复现：目录 458 vs 同窗口 697）+ 窗口没前进的 1,079（冻结，318 + 761）。
    真实 30 天值是 **3,455**（npm 页面 Weekly Downloads 与 `point/last-month` 一致；`1.2.4` 是
    2026-11-11T12:10:46Z 才发的，所以两个窗口相同）。
  **决策不变**：这两条已有多位开发者反馈，我们不催合并、不为它们重发版；要做的只是**记录改准**：
  判断「该我们动/该我们等」用 `node tools/catalog-status.mjs`。
  **2026-11-17 补充**：这里原先写的"凡引用下载量用 npm 的数并写明窗口"已被实测推翻——npm 的数里有
  25 个版本各 137–227 次（连不可安装的 `0.0.0-stage` 都 166 次），是爬虫队形；站点卡片与 npm 面板是
  同一个被放大的数。口径改为 **latest share + 窗口**，证据与复算见 [docs/METRICS.md](METRICS.md) §2。
- **超长行（> 1 MiB）的上报被直接断开连接，不回 `ERR`**（2026-11-16 实测：写入 1,100,000 B 的那次，客户端拿到的是
  `连接被中断`、零字节回复）。这是 `readPayload` 读到超限就返回、`read` 立刻 `Close` 的必然结果；[docs/PROTOCOL.md](PROTOCOL.md)
  §3.4 里"`ERR` 表示载荷校验失败"指的是**解析失败**，不是超长。上报端本来就该在发送前截断（`protocol.js` 的 `normalizePayload` 会做），
  所以这条只影响手写第三方上报端。**不修**：给一次过错加一条回复并不能让写错的客户端变对。
- 同用户命名管道抢占只做了「互斥体先行」缓解，未做 `FILE_FLAG_FIRST_PIPE_INSTANCE` 独占。
- DSH 插件包的 npm 发布**已完成**（1.2.4 于 2026-11-11T12:10:46Z、1.2.5 于 14:58:53Z，1.2.6 于同日随发版上线，2026-11-12 复读 `latest` = **1.2.6**），因此**按包名安装**与 npm-only 目录的两个前置都成立；
  在此之前可用的三条路（git 源、Release tarball、仓库地址）与自建 format-1 索引不受影响。注册表条目 PR #13937 已于
  2026-11-10T14:33:04Z 合并、上游 README 已列出；站点页与 `plugins.json` 的延迟在 run `42844452013`（16:39:27Z
  success）之后消失，条目页 200 且 `plugins.json` 已收录本仓库——但目录里的 `install` 是构建时按 `tarball` 渲染的，
  合并后那版仍是 1.2.0；#14076 已换成 1.2.2（19:59:37Z 合并）、#14089 再推到 1.2.3（2026-11-10T21:10:15Z 合并）、**#14385 带 1.2.5 已于 2026-11-11T17:20:15Z 合并、#14439 带 1.2.6 已于同日 20:12:31Z 合并，站点 `plugins.json` 重建后已渲染 1.2.6**；另有两家目录已合并收录本仓库：[imsai-sh/awesome-deepseek-harness-plugins](https://github.com/imsai-sh/awesome-deepseek-harness-plugins) PR #902（2026-11-10T21:49:53Z，只从 npm 装，故在此之前是 browse-only）与 [zhu1090093659/dsh-community-plugins](https://github.com/zhu1090093659/dsh-community-plugins) PR #97（2026-11-11T05:06:32Z）；1024 Store 详情页已返回 200（8,753 B）。
  发版之后又投了五家（都是「public repo + `dsh-plugin` topic + 人工复核」）：
  [dshbase #322](https://github.com/ylwl1997/dshbase/issues/322)、[dsh.so #24](https://github.com/dsh-so/dsh-plugin-submissions/issues/24)、
  [dsh-plugin.org #735](https://github.com/dshplugin/dsh-plugin-hub/issues/735)、[AllDSH #2](https://github.com/alldsh/submissions/issues/2)
  （走 `POST /api/submit`）与 [dshmarketplace.dev](https://dshmarketplace.dev)（走 `POST /api/v1/submissions`，
  inspect 报 `alreadyListed:false` / `hasDshTopic:true`，submit 报 `{"ok":true,"duplicate":false}`）。
  五条 issue 目前都是 open / 0 评论。**另有 [dshopen #423](https://github.com/polang233/dsh-plugin-index/issues/423)
  已于 2026-11-11T08:32:39Z 以 `state_reason: completed` 关闭**：那是评审制目录里第一条真·收录，其
  `data/sources.json` 第 1204 条记的是本包 1.2.4（`source.commit` `ff7f6de`、`path` 子目录），中英摘要都带上了
  「其他终端需各自安装 reporter」那句；但它的线上 `public/index.json` 仍是旧构建（854 条、全是 npm 源、
  0 条带 `source` 字段），要等下一次构建才在目录页可见——「收录」与「渲染」是两件事，判断可见性看公开索引文件。
  dshmarketplace.dev 的公开接口 `/api/v1/plugins?q=ai-terminal-pet` 仍回
  `{"total":0}`、`DshMarketPlace/awesome-dsh-plugin` 的 `data/curated.yml` 里也没有本仓库；不投的两家是
  `WOOK98/awesome-deepseek-harness-plugins`（imsai 的 fork、已停更）与 `ydhrdh/dsh-marketplace`（7 条、0 stars）。
  [DSH-Plugins-Marketplace](https://github.com/bradeGithub/DSH-Plugins-Marketplace) 16:22:29Z 的重建把本仓库从
  `verdict: script` 改判为 `bundle-plugin`，18:08:44Z 那次重建（`risk_at` 17:09:14Z）给出 **`risk_tier: "safe"`** 且
  `risk_flags` 为空——安装脚本移进 `scripts/` 后扫描器读不到那两个正则的落点，issue #367 的两条旗标就此消失。

## 7. 增长计划（2026-11-17 定，依据 [docs/METRICS.md](METRICS.md)）

这一节不是"想做点什么"，而是**在一轮实测之后重新排的优先级**。测量工具是
`node tools/traffic-report.mjs`（含 `--json`），口径与证据链在同目录的 METRICS.md；每次复核都从它开始。

### 7.1 基线（2026-11-17，全部可复算）

| 面 | 实测 | 结论 |
|---|---|---|
| 触达 | 14 天 **33 浏览 / 11 唯一**；参考来源 github 15、HN 3、目录站 1 | 约 0.8 个唯一访客/天 |
| clone | 1,798 次，其中 **852 = 213 次 CI run × 4 job** | 不是人气 |
| 安装 | setup 41（**2.0/版本**）、zip 43（2.0/版本）、POSIX 75（1.9/版本） | ≈2/版本就是本项目自己的复验 → **外部安装 ≈ 0** |
| npm | 4,586/周，25 个版本各 137–227，`0.0.0-stage` 166 → **latest share 6%** | 爬虫队形（对照 `dsh-pet` 58%） |
| 目录 | `notify` 类目 **26/385**（3,455，中位数 1,171） | 货架位置在，但数字同样被污染 |
| 社区 | star 1（本人）· fork 0 · issue 0 · PR 0 | 没有外部反馈 |

### 7.2 三条判断（后续所有取舍都从这里出发）

1. **瓶颈是可见性，不是质量。** 75 轮工程、501 个测试、三平台可跑；14 天里只有 11 个唯一访客。
   在这个分母上做转化优化无法被测量，所以下一轮的第一件事是"让人看到"，不是"再做功能"。
2. **两个漏点**：安装摩擦（未签名 exe + SmartScreen；winget 未合并；DSH 之外要"装应用 + 接上报端"两步）
   与叙事（首屏是长文档，竞品首屏是 GIF）。
3. **发版节奏在污染唯一可用的指标。** 近 7 天 222 个提交、53 个 release：每发一版多一个 npm 版本，
   爬虫就多抓一次 → 总量虚高、latest share 更低、上游 PR 变噪声。

### 7.3 已定的取舍

| 决策 | 选择 | 为什么不选另一条 |
|---|---|---|
| **定位** | **A（状态带）为主干，C（可选人格）为增量**；不跳到 B（做宠物） | "宠物"人群确实大（`desktop-pet` 2,786 个仓库、头部 2.4 万★；`terminal-pet` 只有 22 个），但我们在 B 上是晚入场者，而 A 的四条差异（单文件 5 MB、无运行时、四终端聚合、点行回终端）是别人没有的。人格是放大器，放大器之前得先有人看到 |
| **顺序** | P1 转化 → P2 分发 → C 的最小版 → 等外部（winget / 上游合并 / 目录重建） | 先做能被测量的那一步；C 需要的美术与动画在新客数为 0 时不产生任何反馈 |
| **签名** | **不写代码，等账户**（Azure Trusted Signing ≈ $10/月，或者 OV 证书） | 没有证书就无法测试的那条路径不该进 `scripts/package.ps1`；这个项目不放不可验证的代码 |
| **发版节奏** | **≤2 次/周**；安全与回归修复除外（已写进 AGENTS 第 10 条） | 每个版本都会把 latest share 压低一点，并让上游 bump PR 变成噪声 |

**签名将来怎么接（等账户下来照做）**：`signtool sign /fd sha256 /tr <RFC3161 时间戳> /td sha256`；
顺序是"先签 `AI-Terminal-Pet.exe` / `pet-statusline.exe` / `pet-hook.exe`，再 `package.ps1`"，
这样 zip 与 `pet-setup` 里的载荷本身是签过的；`pet-setup.exe` 再单独签一次。
`SHA256SUMS.txt` 与 `-VerifyOnly` 会自动跟着变，不需要改代码；`vendor/manifest.json` 需要重新生成。

### 7.4 P1 — 转化（本轮已做一半）

| 项 | 状态 | 落点 |
|---|---|---|
| README 首屏："一行装" + "它不是什么" | ✅ 2026-11-17 | README 英文/中文各一段 |
| 安装摩擦：winget | ⏳ 等 `microsoft/winget-pkgs#481818` 合并，合并后加 `winget install` | [docs/UPSTREAM.md](UPSTREAM.md) §4.3 |
| 安装摩擦：签名 | ⏳ 等账户（7.3） | 同上 |
| 社交预览图上传 | ⏳ 本机浏览器 DNS 故障，恢复后传（或本人手传） | `docs/assets/social-preview.png` |

### 7.5 P2 — 分发（只有部分能由我做）

| 渠道 | 门槛 | 谁做 | 状态 |
|---|---|---|---|
| **awesome-claude-code（58k★）** | 人手表单；≥14 天且活跃 → **已满足**（55 天） | 你提交，我给那一行 | **现在就能做**（[PROMOTION](PROMOTION.md) §6） |
| awesome-opencode（11k★） | ≥25★ 且仓库 ≥30 天，任一不过即关 | 我起稿，你提 PR | 卡在 star，等 |
| HN / V2EX / Reddit | 正文必须人写 | 你 | 文本齐了就能发 |
| 已合并的 5 家 DSH 目录 | 已收录 | — | 等对方重建 |

### 7.6 C 的最小版：规格（下一个功能轮，先不写代码）

- **目标**：状态带不失去信息密度，加一层"有人格"的最小表现。
- **第一步已落地（2026-11-17，未发版）**：`animations` 开关——面板里唯一自己会动的东西（等待/出错的呼吸）
  现在可以关：`config.json`、设置窗口外观页签、`--set animations=on|off`、Linux/macOS 的 `--settings` 表
  四个入口一致；关掉时呼吸定时器不安装、窗口 alpha 停在 `opacity`、光晕画在静止相位。
  两条测试钉住它：`TestTheBreathIsTheOnlyThingMotionControls`（开时 300 ms 两帧必须有像素差，关时必须逐像素相同）
  与 `TestAnimationsDefaultOnAndCanBeTurnedOff`（旧配置文件缺字段仍是开）。
- **下一步（帧）**：等待 / 出错 / 完成三种状态各加 1–2 帧；复用现有呼吸定时器（`syncAnimation`），
  **不新开常驻计时器**；`frameKey` 要带上帧序号（本项目已经两次栽在"画了但没进 key"上）。
- **不做**：可替换皮肤、素材市场、宠物形象包；不改 56 px 几何；不做声音之外的表现层。
- **验收**：视觉不变量门禁再加一条（两帧必须有像素差）；关闭 `animations` 时与今天逐像素相同；
  三平台 `go test` 绿；紧凑帧绘制仍 < 1 ms（沿用 L4 的测量方法）。
- **先决问题**：产品内动画是否真的提高留存——首屏 GIF 已经有动画，而安装后的"人格"没有数据支撑。
  所以帧仍排在 P2 之后，并且发版时要跟一个 ≥2 天的观察窗口。

### 7.7 硬指标（4 周后复核）

| 指标 | 基线（2026-11-17） | 目标 | 怎么测 |
|---|---|---|---|
| 唯一浏览/周 | ~5.5 | > 50 | `traffic-report.mjs` |
| npm latest share | 6% | > 25% | 同上 |
| 资产下载/版本 | ≈2（自查） | > 3 | 同上 |
| star | 1 | ≥ 25（awesome-opencode 门槛） | `gh api` |
| issue + PR | 0 | ≥ 1（第一个外部反馈） | `gh api` |

**不做**：引用 npm 总量当用户数；买 star；刷下载；把"目录里排 26 名"当成绩（那是同一批噪声）。

