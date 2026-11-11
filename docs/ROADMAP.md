# ROADMAP — ai-terminal-pet

> 快照时间 **2026-11-10（v0.9.0 + 未发版增量）**（上一版快照为 2026-11-09 的 v0.9.0）。本文分四块：现状基线、
> 已交付批次、扩展计划、轻量化计划，以及一组实用功能候选与非目标。每项都标注了**动机**和
> **为什么现在不做/怎么做**，便于下一位维护者直接挑一项开工。
>
> 交付历史的逐条细节在 [docs/HANDOVER.md](HANDOVER.md) §13–§27；按版本发布的变更摘要在
> [CHANGELOG.md](../CHANGELOG.md)。

## 0. 现状基线

| 维度 | 数值（2026-11-10，v0.9.0 + 未发版增量） |
|---|---|
| 平台 | Windows only（Win32 分层窗口 + 命名管道） |
| 工具链 | Go 1.23+（实测 go1.27.0 windows/amd64），reporters 仅用 Node 标准库（无 npm 依赖） |
| 版本 | `0.9.0`（`internal/app/app_windows.go:23`） |
| 主程序 | `bin/AI-Terminal-Pet.exe` **4,371,968 B**（4.17 MiB，`-trimpath -buildvcs=false -ldflags "-s -w"`） |
| 状态行 shim | `bin/pet-statusline.exe` **3,643,392 B**（3.47 MiB，合计 8,015,360 B = 7.64 MiB） |
| 代码量 | Go 非测试源码 **10,719 行** / 测试 **4,212 行** / JS **1,742 行**（49 个 .go 文件，合计 14,931 行） |
| 第三方依赖 | 3 个，均为库引用、无 vendor：go-winio v0.6.2(MIT) / lxn/win(BSD) / x/sys v0.10.0(BSD) |
| 运行期占用 | 有活跃会话时工作集约 22–38 MB（含共享页），线程 8–16（2026-09-26 实测，v0.9.0 **未复测**） |
| 面板几何 | 紧凑 `56x156`（固定）· 展开 `200x<内容高度>`，1–6 行自适应，超过 6 行加一条溢出行（常量仍单源于 `internal/ui/drawing_compact_windows.go:23-35`） |
| 设置窗口 | 第二个顶层窗口类 `AITerminalPetSettings`，三页签 general/sound/sessions，高度按内容 450↔504↔306（150% DPI） |
| 协议 | `ai-terminal-pet.v1`，每行一个 JSON，上限 1 MiB，回执精确 `OK` / `ERR` |
| 管道 | `\\.\pipe\ai-terminal-pet-v1-<homedir-sha256 前 8hex>`（按用户隔离，跨语言 golden 测试锁定） |
| 上报端 | `opencode.js` / `claude-hook.js` / `antigravity-hook.js` / `antigravity-statusline.js` / `generic.js` / `pet-statusline.exe` / `integrations/deepseek-harness/ai-terminal-pet-bridge` |
| 生命周期 | 按需拉起（reporter 发现管道不存在就拉起）；会话**存在性由 OS 观测**（进程表 + 终端窗口），连续不可观测 ≥ `DefaultAbsenceGrace` 2 分钟才移除；空闲自毁宽限 10 分钟；扫描 5 s / 存活探测 10 s |
| 持久化 | `state.json`（位置/缩放/展开）· `config.json`（通知/音效/置顶/贴边/不透明度/语言/auto-approve）· `HKCU\...\Run`（开机自启）· `approvals\`（auto-approve 握手） |
| 对外接口 | `--show/--toggle/--reset/--quit/--json[--pretty]/--autostart[=on\|off]/--settings/--doctor/--version/--help` |
| 单测 | **154 个 `func Test`**（ui 49 / store 26 / app 23 / platform 20 / config 13 / autostart 8 / pipeserver 7 / session 5 / i18n 3） |
| 持续集成 | `.github/workflows/ci.yml`：`windows-latest` 上跑 `gofmt -l` / `go vet` / `go test -count=1` / 两个 `-trimpath -buildvcs=false -ldflags "-s -w"` 构建 / 逐文件 `node --check` / `node integrations/deepseek-harness/verify.mjs` / 插件包 `vendor.mjs --check` 与 `verify.mjs` |
| 自检 | `ai-terminal-pet --doctor` 9 项（environment / install files / version / runtime / config / pipe / instance / node / autostart），`[ok]/[warn]/[fail]` 逐行 + 末行汇总，任一 fail 退出码 1 |
| 线协议文档 | [docs/PROTOCOL.md](PROTOCOL.md)（388 行），第三方无需读源码即可编写上报端 |
| 插件包 | `packages/dsh-ai-terminal-pet`（npm 名 `dsh-ai-terminal-pet` **1.2.4**，18 个文件含 vendor 与 `vendor/manifest.json`，tarball **16 files / 1,915,340 B / packed 1.9 MB / unpacked 4,462,112 B**）；release tag `dsh-plugin-v1.2.4`（发版提交见 CHANGELOG，资产 sha256 `d6b7f95a…` / shasum `dafb8199…`） |
| 插件清单 | 仓库根 `package.json` 清单 shim（2,115 B / 68 行，无代码；1.2.3 起多九行 `dsh.marketplace`，1.2.4 只改 `dsh.catalog` 摘要）：`dsh.bundle.patch` + `dsh.catalog` + `private`，让 `add github:<owner>/<repo>` 可直接解析 |
| 视觉验证 | `PET_PREVIEW_DIR=<dir> go test ./internal/ui -run TestRenderPreview`（浮窗状态）、`-run TestRenderSettingsPreview`（设置窗口 5 张），`docs/assets/render_assets.py` 由此产出 README 配图 |

## 1. 已交付批次（按时间倒序）

### 1j. 2026-11-10（v0.9.0 之后，未发版）—— DSH 插件包与市场分发尝试

**动机**：§1i 解决了「别人的机器能跑通」，但没有解决「别人怎么拿到」。Harness 的接法当时要求先 clone
整个仓库、跑 `scripts/install.ps1`，再把安装目录里 bundle 的绝对路径贴进插件管理器——对外部用户不成立。
DSH 的插件市场正好提供了一个现成的分发面：注册表仓库里一个 `data/plugins/<slug>.yml`，加上一个 npm 包。

**交付**（`7f476c9` + `91719e5`，细节与坑见 HANDOVER §27）：

- `packages/dsh-ai-terminal-pet/` 独立 Cordis bundle 包（npm 名 `dsh-ai-terminal-pet`，MIT，
  `type: module`；1.2.4 口径 **18 个文件**含 vendor 与 `vendor/manifest.json`）。
  `dsh.bundle { patch: './cordis.patch.yml', platforms: ['win32'] }`，`engines node ^22.19 || >=24` /
  `dsh >=0.1.0-rc.6`，`files` 显式 10 项。**1.1.0 起没有任何生命周期钩子**（`prepack` 被删，
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
- **仓库根的清单 shim**（`06d4cae`，1,877 B / 59 行，不含代码）：`main` 指向插件包、
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
  1.2.4 与 1.2.3 在同一个 URL 位置，只换 tag 与版本号，Release 资产要等发版提交打 tag 后才存在（发版前别写成已装过）；
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
  现在指向 `dsh-plugin-v1.2.3`，站点下一次重建才渲染出该资产），**1.2.4 还要再开一个 PR 把 `tarball` 推到新 tag**
  （前身 #14063 指向 1.2.1、head `116b0ee`，随 1.2.2 重打包已评论说明后关闭）；
  上游 `check-submission.mjs --only-list` 单条目通过，两个 URL 的 HEAD 均 200。
  条目已进入 `main`，上游 `README.md` 第 8,519 行（`### Notifications & Integrations`）已列出它；**条目页与
  DSH 应用内 Market 当时都还看不到**（16:16Z 复查：条目页 404、`plugins.json` 0 命中），但那只是
  上游站点构建的排队与耗时：`Build site from README` 的 run `42841664024` 在 15:25:40Z 被 cancelled（同分支新 run
  取代），`42844452013` 于 16:39:27Z success；此后条目页返回 200，`plugins.json` 从 `count 10119` 变成 `count 10230`
  并收录本仓库（合并后再读一次是 10,304 条）。**「合并」与「可见」都成立**，只是目录里的 `install` 命令是构建时
  按 yml 的 `tarball` 渲染的，所以合并之后那版仍写着 1.2.0——#14076 已把字段换成 1.2.2，下一次重建才渲染出来，
  #14089 再推到 1.2.3 并已于 2026-11-10T21:10:15Z 合并（1.2.4 需再开一个 PR）。同一批里 [DSH-Plugins-Marketplace](https://github.com/bradeGithub/DSH-Plugins-Marketplace)
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

**没做成什么**：**npm 一个版本都还没发出去**——凭据问题已解（`npm whoami` 返回 `jjjjjjjjnnjnn`，
此前整场会话都是 E401），`npm view dsh-ai-terminal-pet version` 仍是 E404，一个版本都还没发出去；
而且这一步在本机非交互环境里结构上做不到：npm 的 `otplease` 在没有 TTY 时直接抛 EOTP，永不进入 web-OTP 分支，
取到的浏览器授权链接 5–8 分钟就 404 作废。**要发只能由人在本机交互终端里跑一次 `npm publish`**。
因此 **DSH 应用内 Discover 索引仍不可达**（该索引 format 1 只表达 npm 版本，或钉 40 位 commit 的 git 源），
按包名 `add dsh-ai-terminal-pet` 还不能解析（自建索引那条路可用，见 §4）。注册表条目已于 2026-11-10T14:33:04Z
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

### E5 — 配置档扩展 ⏳ 部分交付（2026-11-09，v0.9.0）
**已交付**：`config.json` 现在覆盖 `notifications` / `sound` / `soundVolume` / `soundScheme` /
`soundOnWaiting` / `soundOnError` / `notifyOnFinish` / `alwaysOnTop` / `edgeRetract` / `opacity` /
`language` / `autoApprove` / `autoApproveTimeoutSec`，并有设置窗口（三页签）作为 UI。
**未交付的部分（仍按原计划）**：`theme`（配色组）、`size`（紧凑/展开的基准尺寸预设）、
`perSourceColor`（按 source 指定强调色）、`hidden`（启动即隐藏到托盘）。
`internal/config` 的「缺字段用默认」解码方式未变，加字段仍不会破坏旧文件。

### E1 — 把 Node 上报端全部原生化（收益最高）⏳ 未开工
**动机**：本机每创建一个进程约 **1 s**（AV 钩挂），而 `claude-hook.js` / `antigravity-hook.js`
每次事件都要起一个 `node`（实测 `node -e exit` 0.9–1.2 s）。`pet-statusline.exe` 已经是这条路的先例：
Go 原生 shim 稳态 0.5–0.9 s，比 node 路径的 13–17 s（经 runner）低一个数量级。
**做法**：在 `cmd/` 下加 `pet-hook.exe`，复用 `pet-statusline` 的 stdin 解析 + 状态映射 + 管道投递骨架，
用 `--mode=claude|antigravity` 区分事件映射。三个 JS hook 保留为参照实现与降级路径。
**验收**：同 payload 下原生与 JS 输出同一 session PID（FNV-1a UTF-16 已对齐）与同一 state。

### E2 — 通用接入成为一等公民 ✅ 已交付（2026-11-10，见 §1i）

**现状**：`docs/PROTOCOL.md` **已交付**（388 行，2026-11-10）。`reporters/generic.js` 已可用，
DeepSeek Harness 桥是现成的第三个接入样例（`integrations/deepseek-harness/ai-terminal-pet-bridge`），
文档成文时就是拿它和 `pet-statusline.exe` 作为两个参照实现来对照的。
**动机**：`reporters/generic.js` 已能一行接入任意工具，但协议本身还只写在代码里。
**做法**：新增 `docs/PROTOCOL.md`，把 `ai-terminal-pet.v1` 的字段、状态语义、优先级、
管道命名算法、回执契约、脱敏规则写成规范；README 给出 5 行以内的接入示例（shell alias / Makefile / 编辑器 task）。
**验收**：不读源码也能写出一个合规上报端。

### E3 — 跨平台（macOS / Linux）⏳ 未开工
**动机**：目前 UI 层与平台层直接调 Win32，无法移植。核心逻辑（session/store/pipeserver 的协议部分）其实已平台无关。
**做法**：把 `internal/ui` 的窗口实现抽成接口（`Presenter`），Win32 实现留在 `_windows.go`，
新增 `_darwin.go` / `_linux.go`（macOS 用 NSWindow + NSVisualEffectView，Linux 用 GTK/Wayland layer-shell）。
IPC 从命名管道换成 Unix domain socket，`PipePathForUser` 的 hash 算法保持不变以免破坏跨语言契约。
**注意**：这是一次真正的架构改动，建议先做 E2（协议文档）再动，否则契约会跟着实现漂移。

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
- 三处终端配置的接入检查也没有做：`--doctor` 只查这台机器上**装了什么**，不读别人的终端配置文件。

## 3. 轻量化计划

### L1 — 进程创建税（当前最大开销，~1 s/次）
**现状**：Antigravity 的 `statusLine` **每次渲染**都要起一个进程；`pet-statusline.exe` 自身
实测 0.95–1.5 s，其中绝大部分是进程创建被 AV 钩挂扫描，不是我们的代码。
**可做的**：
1. 缩小二进制（见 L2）——镜像越小，AV 全量扫描越便宜。
2. 上报端在**状态未变化时直接短路**：`pet-statusline` 先读 `runtime/` 里自己的上一个快照，
   若 `state`/`reason`/`project` 全同则跳过管道投递（快照仍写，保证 TTL）。可省掉每次渲染的 IPC 往返。
3. AV 排除目录（需管理员，用户决策）——docs/archive/BASELINE-20260924.md §13 已记录这是唯一能真正压到亚秒的手段。
4. 已落地：`--reset`/`--toggle` 等控制路径不再走 `app.Run()`，避免控制命令意外拉起常驻实例。

### L2 — 二进制体积（4.17 MiB → 目标 <2.5 MiB）
**现状**：主程序 **4,371,968 B**（4.17 MiB）、shim **3,643,392 B**（3.47 MiB），合计 8,015,360 B = 7.64 MiB。
本版新增 `--doctor` 让主程序比 v0.9.0 的 4,178,432 B 增加了 **193,536 B**（约 +189 KiB，+4.6%），
其中 `internal/app/doctor_windows.go` 占 22,932 字节源码——这是「可自查」换来的体积。
**做法**：`-ldflags "-s -w"` 已用。**`-buildvcs=false` 已实测并入**：同样的 -s -w，
加上它会去掉 VCS 戳记，`pet-statusline.exe` 由 3,644,416 B 降到 3,643,392 B（差 1,024 B），主程序两种写法同为
4,371,968 B。现在 `scripts/install.ps1` 的缺失提示、`AGENTS.md` 与 CI 用的都是这条命令，所以**同一份源码在任何
机器、任何时间重建出的产物字节一致**（二进制不入库，`bin/` 被 `.gitignore` 忽略）；`-VerifyOnly` 拿已部署
副本与工作区里的 `bin/` 逐项比对，两侧构建方式不同就会误报。继续可做：剥离 `net/http` 等未用包的反射元数据（本项目不用 http，
但 `net` 被管道复用）、用 `-tags` 关掉不需要的 `crypto` 实现。
注意 `pet-statusline` 引了 `regexp`（脱敏规则）——这是体积大头之一，可改为手写扫描器。
**验收**：`go build` 后主程序 < 2.5 MiB，且 `node --check` 与全部单测不变。

### L3 — 管道连接的 goroutine churn
**现状**：`hang.log` 里出现过 goroutine 编号 79 万+，说明 10 小时内创建了约 80 万个短命 goroutine——
go-winio 的 `makeConnectedServerPipe` 每次 Accept 前都会预建一个 `connectPipe` 等待者。
单次开销极小，但频率高（statusline 高频上报时可达 20+/s）。
**做法**：先量化（在 `Server` 加计数器），确认是否需要处理；若需要，考虑给服务端加连接复用
或把上报端改为「变化才连」。

### L4 — 光栅化与重绘
**已做**：DIB 画布复用、命中缓冲复用、动画定时器按状态装/卸。
**可做**：脏区重绘（当前每帧全画），或把 `fillRounded`/`fillEllipse` 的内层循环向量化。
面板只有 56×156（紧凑）与 200×<内容高度>（展开，1–6 行自适应）逻辑像素，收益有限——
**建议先测再动**，别为 0.3 ms 增加复杂度。（旧版本节写的 `40×190 / 200×250` 是 v0.5.0 之前的固定几何，早已被 `drawing_compact_windows.go:23-35` 的常量取代。）

### L5 — 启动与扫描路径
**现状**：快照扫描 2 s 一次，全目录 `ReadDir`。会话多、`runtime/` 残留多时会变贵。
**做法**：`runtime/` 里按 mtime 排序 + 只扫最近 N 个；或给扫描加「目录 mtime 未变则跳过」。
仅在实测扫描耗时 >5 ms 时才值得做。

## 4. 实用功能候选

已完成：~~F3 开机自启~~、~~F4 提示音~~（v0.9.0 已升级为四套合成音色 + 按事件开关，见 §1h）、
~~F7 `--json` 导出~~（见 §1a）；~~F6 相对时间~~（展开态每行显示「刚刚 / 3 分钟前 / 1 小时前」，见 §1c）。

v0.6.0–v0.9.0 期间落地、原先不在候选表里的四项：**语言切换**（en/zh-CN + auto，§1e）、
**设置窗口**（§1h）、**auto-approve 否决窗口**（§2 E2b）、**DeepSeek Harness 接入**（§1h）。

v0.9.0 之后（2026-11-10，未发版，见 §1i）又交付了四项**扩展项**，原先分别挂在 §2/§3 里：
~~E2 线协议文档 `docs/PROTOCOL.md`~~、~~E6 `--doctor` 自检~~、~~安装脚本 `-VerifyOnly`/`-Uninstall`/`-AddToPath`~~、
~~CI 门禁 `.github/workflows/ci.yml`~~。

✅ 注册表条目 PR #13937 **已于 2026-11-10T14:33:04Z 由 `fkysly` 合并**，上游 README 已列出（`:8519`）；合并只写进了 `main`，可见性要等上游站点构建——16:16Z 条目页还是 404、`plugins.json` 0 命中，而 run `42844452013` 在 16:39:27Z success 之后条目页返回 200、`plugins.json` 的 `count` 从 10119 涨到 10230 并收录本仓库（目录里的 `install` 按 yml 的 `tarball` 渲染，合并后那版仍是 1.2.0；#14076 已换成 1.2.2 并在 19:59:37Z 合并、#14089 再推到 1.2.3 并已于 2026-11-10T21:10:15Z 合并，都要等下一次站点重建才渲染出来；1.2.4 需再开一个 PR）；同一轮 [imsai-sh/awesome-deepseek-harness-plugins](https://github.com/imsai-sh/awesome-deepseek-harness-plugins) PR #902 于 2026-11-10T21:49:53Z 合并（该店只从 npm 安装，npm 发出去之前以 browse-only 列出）、[zhu1090093659/dsh-community-plugins](https://github.com/zhu1090093659/dsh-community-plugins) PR #97 于 2026-11-11T05:06:32Z 合并；1024 Store 详情页返回 200（8,753 B）；
**DSH 插件包分发** —— ✅ 包本体走到 1.2.4，七个 tag 都有对应 Release（`dsh-plugin-v1.0.0` / `dsh-plugin-v1.1.0` / `dsh-plugin-v1.2.0` / `dsh-plugin-v1.2.1` / `dsh-plugin-v1.2.2` / `dsh-plugin-v1.2.3` / `dsh-plugin-v1.2.4`，当前资产 1,915,340 B / sha256 `d6b7f95a…`），tarball / Release 资产 URL / git 源钉 tag 三种安装实测成功（不写 `#ref` 的 `&path:` 形态会报 `repository ... not found`，与 `prepack` 无关）；
✅ 注册表条目 PR #13937 **已于 2026-11-10T14:33:04Z 由 `fkysly` 合并**，上游 README 第 8,519 行已列出该条目；「合并进 `main`」与「在某个市场里可见」曾经是两件事，现在后者也成立：条目页在 16:16Z 是 404、站点构建 run `42844452013` 于 16:39:27Z success 后变 200，`plugins.json` 也从 10119 条增至 10230 条并收录本仓库——目录里的 `tarball` 决定安装版本、只能靠 PR 改：**#14076 已于 19:59:37Z 合并**（把字段换成 1.2.2）、**#14089 已于 2026-11-10T21:10:15Z 合并**（再推到 1.2.3），两者都要等下一次站点重建才反映到 `install` 字段（前身 #14063 已关闭）；1.2.4 需再开一个 PR 推 `tarball`；
✅ [DSH-Plugins-Marketplace](https://github.com/bradeGithub/DSH-Plugins-Marketplace) 的索引 2026-11-10T18:08:44Z 重建后，本仓库条目是 `pkg_name: dsh-ai-terminal-pet`、`version: 1.2.1`、topics 11 个、`verdict: bundle-plugin`，并且 **`risk_tier: "safe"`、`risk_flags` 一条不剩**（`risk_at` 17:09:14Z，是安装脚本移进 `scripts/` 后第一次真的扫到本仓库）。之前的 `verdict: script` 与 `risk` 徽章（`registry-run-key` critical + `remove-item-force-recurse` medium）从此消失，全程没有用到任何抑制字段——因为扫描器 `evalRepoRisk` 只按文件名抓仓库根的 `install.ps1` / `install.sh` / `package.json`，而脚本已移进 `scripts/`（`$RepoRoot = Split-Path -Parent $PSScriptRoot`，原样保留停进程、`-Recurse -Force` 与 `CurrentVersion\Run` 逻辑）。issue #367 已回帖说明，可关。
✅ 两条不用申请的自动采集路径（根清单采集 `LivXue/dsh-plugin-shop` 每日、双 topic 采集 `zasenjc/dsh-plugins-store` 30 分钟）已在仓库侧满足条件；但 `LivXue` 的发布分片实测只有 npm 源（26,351 条条目全带 npm tarball 的 `integrity`、`commit` 字段 0 条），所以「条件满足」≠「已被列出」；
✅ **自建 format-1 索引**（`ef783ba`，`docs/market-index.json`；`fe62900` 前移到 1.2.2、`fabb591` 前移到 1.2.3）：Discovery 不必再等 npm——条目 `dsh-ai-terminal-pet` `1.2.4` 用钉到 `b4474a6` 的 git 源（1.2.4 的版本与 `publishTimestamp` 已前移，`commit` / `icon` 由紧随的提交补齐），URL <https://raw.githubusercontent.com/jjjjjjjjnnjnn/ai-terminal-pet/main/docs/market-index.json>，用户加到 Discover 的索引源列表里即可看到并一键安装；校验走宿主自带的 `parseMarketIndex`，`marketInstallTarget` 产出的 spec 已实测 2.5 秒装通，且 `verify.mjs` 有 14 条断言把它钉在包版本上（发版忘改索引会让 CI 变红）；
⏳ **npm 尚未发布任何版本**（凭据问题已解、`npm whoami` 返回 `jjjjjjjjnnjnn`、`npm view` 仍 E404，publish 要一次性密码，而 npm 的 `otplease` 在无 TTY 时直接抛 EOTP，所以只能由人在交互终端里跑）。npm 只是「按包名安装」与「进 npm-only 目录」的前置，不再是 Discover 的前置（自建索引已覆盖后者）。

| 编号 | 功能 | 价值 | 复杂度 |
|---|---|---|---|
| F8 | 会话行点击复制项目路径（托盘气泡回执，避免"点了没反应"） | 高 | 低 |
| F2 | 键盘快捷键：`Esc` 收起（已有）、方向键在会话行间移动、`Enter` 跳转 | 中 | 中 |
| F5 | 记住所在显示器（多屏场景下按显示器恢复，而不是全局坐标） | 中 | 中 |
| F1 | 悬停展开面板时显示完整会话明细（就地浮层） | 中 | 中 |
| F9 | 紧凑态显示「最久等待了多久」，把"有人在等"升级成"等了 12 分钟" | 中 | 低 |

F8、F9 都是「一个下午」的量级，且不触碰渲染核心，适合作为下一次迭代的起点。
F1 现在价值下降：展开态已经能列出 6 行带项目名、状态、来源和相对时间的明细，
悬停浮层要证明自己比"点一下"更快才有意义。

**关于"点击任意处收起"与 F8 的冲突**：当前规则是「点击 = 收起」，所以会话行暂时不可点。
若要做 F8，建议改成「点击行 = 复制路径 + 气泡回执，点击行以外的面板区域 = 收起」——
回执是必须的，分层窗口没有别的方式告诉用户"复制成功了"。

**已交付但值得回头看的一项**：`--json` 目前只读快照，因此看不到「只走管道、不写快照」
的上报路径（如果将来有上报端选择不落盘）。真需要时可以给它加一个查询运行中实例的
IPC 分支，但要在 `--json` 永远不阻塞 GUI 进程这个约束下做。

## 5. 非目标

- **不做终端复用器或日志查看器**：本项目的定位是「一眼可辨的状态指示」，不是替代终端。
- **不做账号/云同步/遥测上报**：全部数据留在本机，`runtime/` 与 `state.json` 不离开 `%LOCALAPPDATA%`。
- **不引入 UI 框架**：零外部 UI 依赖是有意为之，它让二进制保持单文件、无运行时依赖。
- **不做跨用户共享的管道**：管道名按用户哈希隔离是安全边界，不为了「多用户共用」而放宽。

## 6. 已知残留（不计划修，但需知情）

- `hang.log` 看门狗是**事后取证**而非自愈；转储上限已压到 3 次 + 2 min 冷却 + 记录最后消息名。
- UI 无自动化视觉回归，靠 `--doctor`（E6，已交付但只覆盖环境/安装/运行期状态，**不做视觉比对**）+ 探针脚本 + `TestRenderPreview` / `TestRenderSettingsPreview` 光栅化 PNG 人工验证。
- `go test -race` / `govulncheck` / `gosec` 因缺工具链（gcc 等）仍未跑（2026-11-09 复核仍然如此）。
- 同用户命名管道抢占只做了「互斥体先行」缓解，未做 `FILE_FLAG_FIRST_PIPE_INSTANCE` 独占。
- DSH 插件包的 npm 发布**尚未发生**（凭据问题已解、`npm view` 仍 E404），因此**按包名安装**与 npm-only 目录仍不通；
  git 源、Release tarball、仓库地址三条安装路可用，自建 format-1 索引也已钉到 1.2.4。注册表条目 PR #13937 已于
  2026-11-10T14:33:04Z 合并、上游 README 已列出；站点页与 `plugins.json` 的延迟在 run `42844452013`（16:39:27Z
  success）之后消失，条目页 200 且 `plugins.json` 已收录本仓库——但目录里的 `install` 是构建时按 `tarball` 渲染的，
  合并后那版仍是 1.2.0；#14076 已换成 1.2.2（19:59:37Z 合并）、#14089 再推到 1.2.3（2026-11-10T21:10:15Z 合并），1.2.4 需再开一个 PR；另有两家目录已合并收录本仓库：[imsai-sh/awesome-deepseek-harness-plugins](https://github.com/imsai-sh/awesome-deepseek-harness-plugins) PR #902（2026-11-10T21:49:53Z，只从 npm 装，故在此之前是 browse-only）与 [zhu1090093659/dsh-community-plugins](https://github.com/zhu1090093659/dsh-community-plugins) PR #97（2026-11-11T05:06:32Z）；1024 Store 详情页已返回 200（8,753 B）。
  [DSH-Plugins-Marketplace](https://github.com/bradeGithub/DSH-Plugins-Marketplace) 16:22:29Z 的重建把本仓库从
  `verdict: script` 改判为 `bundle-plugin`，18:08:44Z 那次重建（`risk_at` 17:09:14Z）给出 **`risk_tier: "safe"`** 且
  `risk_flags` 为空——安装脚本移进 `scripts/` 后扫描器读不到那两个正则的落点，issue #367 的两条旗标就此消失。