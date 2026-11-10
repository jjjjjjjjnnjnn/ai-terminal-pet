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

未发布版本：见下方 `[Unreleased]` 段——2026-11-10 在 v0.9.0 之上完成的文档、自检工具、安装脚本与 CI 改动，
版本常量仍是 `0.9.0`，尚未发版。

---

## [Unreleased] — 2026-11-10

这一批不发新版本号（`internal/app/app_windows.go` 里的 `Version` 仍是 `0.9.0`），
目的是让这个仓库在**别人的机器上**也能跑通、自查、接入与持续验证：可移植性、`--doctor` 自检、
正式线协议文档、安装脚本的只读/卸载/Path 开关，以及 GitHub Actions 门禁。详细 runbook 见
[docs/HANDOVER.md](docs/HANDOVER.md) §1（当前状态）与 [docs/ROADMAP.md](docs/ROADMAP.md) §1i。

### Added

- **`ai-terminal-pet --doctor`**（`internal/app/doctor_windows.go` 563 行 + 375 行测试 /
  14 个 `func Test`）：一次跑完 **9 项**自检——`environment`（编译目标与 `LOCALAPPDATA`）、
  `install files`（安装目录里的 `AI-Terminal-Pet.exe` / `pet-statusline.exe` / `protocol.js`）、
  `version`、`runtime`（快照目录）、`config`、`pipe`、`instance`、`node`、`autostart`。
  每项输出一行 `[ok]` / `[warn]` / `[fail] name: detail`，末行汇总 `doctor: N ok, N warn, N fail`，
  任一 `fail` 时退出码为 **1**；退出码经 `errDoctorFailed` 哨兵值传递，因此**不会写进 `error.log`**
  （自检失败是用户主动请求的诊断结果，不是程序自己出了故障）。用法错误仍走 `usageError` → 退出码 **2**。
  - **刻意不碰互斥体**：`CreateMutex` 会**创建**这个对象、干扰正在运行的实例，所以 `checkInstance`
    改用「管道有人监听」作为实例存在的证据。
  - 管道探测 `probePipe` 连上即关（`doctorDialTimeout` 250 ms），不写任何字节——服务端把立即 EOF
    当作空负载、不会触发 update 回调，所以跑 doctor 不会改变浮窗显示的内容。
  - `node` 探测 `node --version` 限时 300 ms，并把 `Missing` 与 `TimedOut` 分开：
    「没装 Node」和「装了但没应答」对用户是两件相反的事。子进程带 `HideWindow`，从快捷方式启动时不闪黑框。
  - 每项检查都包在 `runCheck` 的 recover 墙里：将来某一项 panic 只产出它自己的 `fail` 行，
    不会带走另外八项；返回空 verdict 也会被计成 `fail` 而不是悄悄消失。
  - 只提供人读文本，**没有 `--json` 变体**：报告是贴进 issue 的，每行已经比可解析格式更好读。
- **`docs/PROTOCOL.md`**（388 行）：正式线协议规范。第三方为未列出的终端自行编写上报端时
  不必再读源码，字段、状态语义、管道名派生规则、`OK`/`ERR` 回执契约与脱敏要求都在这里；
  schema 变更流程见该文档 §11。
- **`reporters/package.json`**（26 行）：给 `reporters/` 目录写显式 `exports`，覆盖 6 个 `.js`
  （`protocol.js`、`opencode.js`、`claude-hook.js`、`antigravity-hook.js`、`antigravity-statusline.js`、
  `generic.js`），不再依赖 Node 的隐式解析；`private: true`，不发布到 npm。
  **刻意不写 `"type"` 字段**：5 个 hook 是 `require()` 的 CommonJS，而 `opencode.js` 是真正的 ESM——
  写 `"module"` 会打断前者，写 `"commonjs"` 会打断后者。无名 type 的代价是首次 import `opencode.js` 时
  Node 打印一条 `MODULE_TYPELESS_PACKAGE_JSON` 提示并按语法探测重解析为 ESM：预期行为，只多一次解析。
  把 `opencode.js` 改名 `.mjs` 的方案被否决，因为 README 与用户配置里引用的是 `.js` 路径。
- **Harness 桥 README 去掉写死的主机路径**：`integrations/deepseek-harness/ai-terminal-pet-bridge/README.md`
  里那句给用户照抄的 `plugin_manager` 目标原为 `C:\Users\<name>\...`，现改为
  `%LOCALAPPDATA%\AI-Terminal-Pet\bundles\ai-terminal-pet-bridge` 并补三行说明：该字符串是贴进
  Harness 插件管理器的路径，按用户解析才有意义，写死某个用户名在别人机器上必然不对。
- **`.github/workflows/ci.yml`**：`windows-latest` 上的单 job 门禁，任一失败即挂——
  `gofmt -l` 有任何输出、`go vet ./...` 有发现、`go test -count=1 ./...` 失败、两个
  `go build -trimpath -buildvcs=false -ldflags "-s -w"` 构建失败、任一上报端 `node --check` 失败、
  `node integrations/deepseek-harness/verify.mjs` 失败、`node packages/dsh-ai-terminal-pet/scripts/vendor.mjs --check`
  失败、`node packages/dsh-ai-terminal-pet/verify.mjs` 失败。（不记行数：这份文件随门禁增删而变。）细节：`concurrency` 让同分支的新推送取消旧运行；
  `fetch-depth: 1`（只要树）；`go-version-file: go.mod` 让 `go.mod` 的 `go 1.23.0` 成为唯一工具链来源；
  Node 22（ESM 语法探测自 22.7.0 起默认开启）；`core.autocrlf` 显式关掉，避免 runner 检出 CRLF 让
  `gofmt` 误报。**`node --check` 是逐文件循环而不是 `node --check reporters/*.js`**：
  后者只检查第一个参数（Node 24 实测：第二个参数是坏文件也退出 0），循环是唯一真正覆盖全部文件的写法。
  live-desktop 探针测试在对应 `PET_*` 变量未设置时 `Skip`，所以不需要桌面会话或任何 secret。
- **`packages/dsh-ai-terminal-pet` —— 把 Harness 桥接做成独立 DSH 插件包**（`7f476c9` 起，历经 1.0.0 →
  1.1.0 → 1.2.0）：npm 包名 `dsh-ai-terminal-pet`、license MIT、`type: module`、`private` 未设（可发布）；
  当前版本 **1.2.0**（`packages/dsh-ai-terminal-pet/package.json` 2,008 B / 74 行）：
  包内 18 个文件（含 vendor，1.2.0 起多一个 `vendor/manifest.json`），主体是 `lib/index.js`（17,872 B / 570 行）、
  `verify.mjs`（22,499 B / 415 行）、`scripts/vendor.mjs`（13,981 B / 324 行）、`package.json`（2,008 B / 74 行）、
  `README.md`（6,665 B / 129 行）、`vendor/manifest.json`（529 B / 19 行）、
  `cordis.patch.yml`（399 B / 10 行）、`locale/en.json`（210 B）/ `locale/zh.json`（217 B）、
  `icon.svg`（526 B）/ `icon.png`（4,386 B）、`LICENSE`（1,070 B）、`screenshots.json`（97 B）、
  `screenshots/compact.png`（4,459 B）/ `expanded.png`（11,335 B）/ `expanded-many.png`（17,697 B）、
  `vendor/protocol.js`（20,914 B）、`vendor/AI-Terminal-Pet.exe`（4,371,968 B）。
  `package.json` 用 `dsh.bundle { patch: './cordis.patch.yml', platforms: ['win32'] }` 声明自己是一个
  Cordis bundle，`engines` 为 `node: '^22.19 || >=24'`、`dsh: '>=0.1.0-rc.6'`，`files` 10 项。
  **1.1.0 删掉 `prepack` 并新增 `vendor:check`**（`b280093`）：`scripts` 现在是
  `verify` / `vendor` / `vendor:check`，没有任何生命周期钩子——pnpm 对 git 依赖会跑包的生命周期脚本，
  未钉版本的源会因此撞上一条完全不提 AI Terminal Pet 的 `allowBuilds` 报错，删掉比解释更可靠。
  `scripts/vendor.mjs` 随之变双模式（无参拷贝、`--check` 只校验 sha256，vendor 缺失或 0 字节即 exit 2），
  `vendor/` 必须由发版人在打包前手工拷一次并提交，`npm pack` 不再替你重新生成它。
  **1.2.0 增加 `dsh.catalog`**：`category: tool`、中英各一行摘要（190 / 104 字符）、五项自我声明的
  `capabilities`（`windows-overlay` / `named-pipe` / `session-state` / `tray-icon` / `no-network`），
  `keywords` 补上 `dsh-plugin`。这些 capabilities **只自我声明、不被强制**——dsh 不对插件做沙箱，
  所以它是描述，不是权限授予；`keywords` 则是自动索引型市场采集的标签，于是多出第二条发现路径：
  注册表评审（条目合并进 `main` 才上架）与 topic/关键词采集（无需提交申请）。版本必须跟着清单内容走，
  否则已发布的不可变 tarball 会与仓库分叉——这正是本仓库一直在消除的漂移。
  行为：把 Harness 事件映射成浮窗状态（`approval/asked → waiting_permission`、`approval/decided → working`、
  `turn/start → working`、`turn/end → idle`、`tool/call → tool_use`、`tool/result(error) → error`、
  `command/done`、`working-directory/change → project`、`session/title`）；记住每个会话最后一次状态，
  只在状态变化或距上次 ≥60 s（`KEEP_ALIVE_MS = 60_000`）时才发一次，避免高频刷屏。
  `npm pack` 的 1.0.0 口径是 **1.9 MB packed / 4.5 MB unpacked / 15 files**；1.2.0 因为 `vendor/manifest.json`
  成为随包文件，文件数变成 **16**，实测 **1,914,566 B / 16 files / packed 1.9 MB / unpacked 4,460,352 B（4.5 MB）**
  （1.1.0 是 1,913,726 B / 15 files）。多出来的 840 B 就是那个 529 B 的清单加上打包开销。
  - **`cordis.patch.yml` 必须用 `insert:` 列表**（`- id: dsh-ai-terminal-pet` / `name: dsh-ai-terminal-pet`）。
    裸的 `- id:` 只会 patch 一行**已存在**的 loader 行，装完看着成功却什么都不挂载——这是踩过的坑，
    因此文件头的注释把原因写在了 YAML 里。
  - **vendor 预编译 + 引导安装只补缺失**（`runBootstrap()`）：仅当
    `%LOCALAPPDATA%\AI-Terminal-Pet\AI-Terminal-Pet.exe` **缺失**时，才从包内 `vendor/` 拷
    `protocol.js`（`0o600`）与 exe；`copyVerified` 走「临时文件 + rename、rename 前再查一次目标是否存在」，
    **从不覆盖已有文件**。`DSH_PET_NO_BOOTSTRAP=1` 可整段关掉（测试缝），
    `DSH_PET_TEST=1` + `DSH_PET_PACKAGE_ROOT` 可替换包根。
  - `node verify.mjs`：rc=0、**54 ok**、末行 `VERIFY PASSED`；覆盖状态映射、waterfall `next()`、
    dispose 不发 `closed`、真 exit 发 `closed`、bootstrap 只补缺失不覆盖、protocol 抛异常时插件全静默。
- **仓库根的 `package.json` 清单 shim**（`06d4cae`，1.2.0 口径 1,877 B / 59 行）：不含任何代码，只让
  「按仓库 URL 安装」在仓库根就能读到 `dsh.bundle`、`main`、`exports`、`icon` 与 `engines`；
  `version` 与插件包同步（`verify.mjs` 断言两者相等，所以一次版本号变更不可能只落在一个清单里），
  并带上同一份 `dsh.catalog`。没有它时，只读仓库根的宿主把仓库当成 0.0.0 依赖、报 `declares no dsh.bundle`
  并装出一个空包——症状是「装成功了，但浮窗永远不出现」，没有任何报错。标 `private: true` 是为了不占用 npm 名。
- **`verify.mjs` 连根 shim 一起检查，并在 CI 里跑插件套件**（`d06001d`）：新增根清单一致性用例，
  从 tag `dsh-plugin-v1.1.0` 的 300 行 / 55 个 `ok` 长到 1.1.0 批次的 331 行 / 63 个 `ok`、再随 1.2.0 到
  415 行 / **91 个 `ok`**（均 `VERIFY PASSED`、rc 0；多出的 14 条是新增的索引漂移用例）；
  `.github/workflows/ci.yml` 增加两步：`node packages/dsh-ai-terminal-pet/scripts/vendor.mjs --check` 与
  `node packages/dsh-ai-terminal-pet/verify.mjs`。验证脚本自带 fake Cordis context + 临时 `LOCALAPPDATA` +
  stub protocol 模块，所以 CI 上不需要网络、管道或真实浮窗。
- **GitHub Release 线：1.0.0 → 1.1.0 → 1.2.0（历史按顺序记，当前版本是 1.2.0）**。
  **1.0.0**（已作废）：tag `dsh-plugin-v1.0.0`，资产 `dsh-ai-terminal-pet-1.0.0.tgz`（1,913,682 B，
  sha256 `7365f6a6f4aea1174ec6f2f00670e868568e944816d1869aeb59db363ae6b5af`），
  下载地址 `https://github.com/jjjjjjjjnnjnn/ai-terminal-pet/releases/download/dsh-plugin-v1.0.0/dsh-ai-terminal-pet-1.0.0.tgz`。
  三种安装方式都实测成功：①本地 tarball `dsh plugin --profile <name> add <tgz 绝对路径>` →
  `+ dsh-ai-terminal-pet 1.0.0`，`dsh --profile <name> --dump-config` 末尾出现 `- id: dsh-ai-terminal-pet`
  （证明 `insert:` 真的挂上了）；②git 源
  `dsh plugin --profile git-test add git+https://github.com/jjjjjjjjnnjnn/ai-terminal-pet.git#91719e5a2c7c221ad163686f80eaab9c4eca4ea6&path:/packages/dsh-ai-terminal-pet`
  → rc=0、10.2 s，装出的包里 `vendor/AI-Terminal-Pet.exe` 4,371,968 B 在位；③Release 资产 URL
  `add https://github.com/jjjjjjjjnnjnn/ai-terminal-pet/releases/download/dsh-plugin-v1.0.0/dsh-ai-terminal-pet-1.0.0.tgz`
  → rc=0、32.6 s，vendor 内两个文件都在位。裸 git 源（不钉 commit/tag）会被 pnpm 的 `allowBuilds`
  门拦在 `prepack` 处失败，所以对外命令只给 tarball 与钉 tag 两种形态。
  **1.1.0**：tag `dsh-plugin-v1.1.0`（annotated，指向 `b28009377e14840ee003ddaf114922e1d167ed0d`），资产
  `dsh-ai-terminal-pet-1.1.0.tgz`（**1,913,726 B / 15 files / sha256 3b8f9186cca5142626f2e25175096f1c4fb59d86cc1cdff5879792950146cc78** /
  shasum `19a2e544ab875c61fcf822e85a91c67273981725`），删掉 `prepack` 后不再有生命周期钩子，因此
  多出第三种安装形态：**直接给仓库地址** `github:jjjjjjjjnnjnn/ai-terminal-pet`（靠仓库根的清单 shim 解析，
  等价 `git+https://github.com/jjjjjjjjnnjnn/ai-terminal-pet.git`）。三种形态实测（计时取在 1.1.0 的资产上；
  1.2.0 资产同位置就位、URL 只换 tag 与版本号）：Release 资产 1.8 s、git 钉 tag 3.4 s、仓库 URL 3.6 s，均 rc=0。仍会失败的是「不写 `#ref` 的 `git+...git&path:/packages/...`」：
  pnpm 把两段拼进 URL 当仓库地址去 `ls-remote`，报 `repository ... not found`——**与 `prepack` 无关**。
  **1.2.0**（当前）：tag `dsh-plugin-v1.2.0`（annotated，tagger 2026-11-10T13:46:52+0100，指向
  `4d777e32e0a9ef57bb7714287dccd7cae065346f`，Release 于 2026-11-10T12:47:24Z 发布），资产
  `dsh-ai-terminal-pet-1.2.0.tgz`（**1,914,566 B / 16 files / sha256
  `ccf9d531267af681438010bc94c721e7822a0db3f4d14575953b1757a2ef65c9` / shasum
  `dc65e0d4259be640c0de7ff1bc86675af5454546`**，实际下载后本地重算与 GitHub 记录的 digest 一致）；
  内容新增 `dsh.catalog`、
  `vendor/manifest.json` 随包（文件数 15 → 16）。1.1.0 保持已发布、不可变，由 1.2.0 取代——清单内容变了
  就必须换版本号，否则仓库与那份不可变 tarball 分叉。1.2.0 发版前 `npm publish --dry-run` 给的是
  **16 files / 1.9 MB packed / 4.5 MB unpacked / shasum `166c3043…**；但 dry-run 之后包内 `README.md`
  又改了 13 B（6,652 → 6,665 B），所以**真正发出去的资产是 1,914,566 B / shasum `dc65e0d4…` /
  sha256 `ccf9d531…`，不是 dry-run 那个摘要**——包里任何一个字节变过，dry-run 的摘要就作废。
- **awesome-dsh-plugin 注册表条目（PR #13937，2026-11-10T14:33:04Z 由评审人 `fkysly` 合并）**：市场 App `dshmarket` 只是前端，
  目录来自注册表仓库 `awesome-dsh-plugin/awesome-dsh-plugin`，投稿方式是只加一个
  `data/plugins/<slug>.yml`。本次条目为
  `data/plugins/jjjjjjjjnnjnn__ai-terminal-pet--packages-dsh-ai-terminal-pet.yml`，
  `category: notify`、`url` 指向 `.../tree/main/packages/dsh-ai-terminal-pet`、
  `name: jjjjjjjjnnjnn/ai-terminal-pet#dsh-ai-terminal-pet`，并带上 `icon`（raw.githubusercontent 的 icon.png）
  与 `tarball`（上面的 Release 资产）。本地先跑通上游两道门：条目模型校验
  `entries read: 10120 / problems: 0`，`scripts/check-submission.mjs --only-list` →
  `ok https://github.com/jjjjjjjjnnjnn/ai-terminal-pet/tree/main/packages/dsh-ai-terminal-pet` /
  `all checked entries pass`。PR 见 <https://github.com/awesome-dsh-plugin/awesome-dsh-plugin/pull/13937>
  （fork 分支 `add-dsh-ai-terminal-pet`，commit `a1fef56`，基于上游 main `09cd105`）。
  **状态是「已合并、尚未在任何市场可见」**：PR 上两个检查都已完成且 `conclusion: SUCCESS`
  （`PR check / check` 于 2026-11-10T10:37:54Z 结束、`Submission gate` 于 10:38:47Z 结束，
  `mergeable: MERGEABLE`），随后在 14:33:04Z 被 `fkysly` 合并。维护者合并时删掉了原稿里「也能在
  面板上直接放行审批」一句——对 DeepSeek Harness 会话而言，面板只做状态映射与跳转到 Harness 窗口，
  审批仍在 DSH 里完成；上游 README 因此把这一条写成了如实描述。接入 `main` 的事实已可验证：
  `awesome-dsh-plugin/awesome-dsh-plugin` 的 `README.md` 在「Notifications & Integrations」小节第 8,519 行
  列着 `jjjjjjjjnnjnn/ai-terminal-pet#dsh-ai-terminal-pet`。但**合并进 `main` 不等于市场里可见**：
  条目页 <https://awesome-dsh-plugin.com/p/jjjjjjjjnnjnn/ai-terminal-pet--packages-dsh-ai-terminal-pet/>
  在 15:56Z 与 16:16Z 两次复查都是 404，而 DSH 应用内 Market 真正加载的
  `https://awesome-dsh-plugin.com/plugins.json`（`count: 10119`、`updated: 2026-11-10`）里
  同样 0 命中——上游 `Build site from README` 工作流（run `42841664024`）自 15:13:06Z 起停在
  build 步骤没动，后面还排着两次重跑。所以现在的正确说法是「条目已合并、列表与市场待上游站点构建落地」。
  **两处外部副作用已记录**：① 已发布的 `dsh-ai-terminal-pet-1.2.0.tgz` 里那份 `README.md` 是 1.2.0 打包时的旧文，
  仍写着「等人工评审、尚未列入」——Release 资产不可变，本仓库的改动不会回写进那份 1,914,566 B 的 tarball，
  所以**下一版（≥1.3.0）发版前必须重新打包**才能让包内说明与事实一致；`docs/market-index.json` 钉的也是
  `4d777e3` 那个 commit，同样读到旧文。写文案时别拿包内 README 当现状。
  ② 已向市场仓库提了误报 issue <https://github.com/bradeGithub/DSH-Plugins-Marketplace/issues/367>：
  `registry-run-key` 与 `remove-item-force-recurse` 两条旗标都是正则匹配到常量定义与提示文本，
  `install.ps1` 从不新增 Run 值、递归删除只作用于一个经 `Assert-SafeInstallDirectory` 校验的目录；
  该仓库没有抑制机制，是否接受由维护者决定。
- **1.2.1 —— 只为让包内说明追上事实，不改代码**：1.2.0 的 tarball 冻结在打包那一刻，里面的
  `README.md` 还写着「等人工评审、尚未列入」。Release 资产不可变，所以这句旧文只能靠新版本发出去改正。
  1.2.1 与 1.2.0 的唯一差别是包内 `README.md`（Install 段改成「条目已合并、市场可见性待上游站点构建」）
  与两条安装命令里的版本号；`lib/index.js`、`vendor/` 两个文件、`cordis.patch.yml`、`dsh.catalog` 全部逐字节不变，
  所以浮窗与协议契约没有任何改动。**这类「只改文字」的发版也必须换版本号**：复用 1.2.0 的号会让 npm/registry
  侧看到同一个版本两份不同内容，也让 `-VerifyOnly` 之外的任何 sha256 比对失去意义。
- **两条「不用申请」的自动采集路径**：仓库 topics 现为 `dsh-plugin` / `deepseek-harness` /
  `cordis-plugin`。`LivXue/dsh-plugin-shop` 每日构建，收录**根 `package.json` 里同时有 `name` 与
  `dsh.bundle`** 的 GitHub 仓库，并钉住默认分支的那个 commit；`zasenjc/dsh-plugins-store` 每 30 分钟同步，
  收录门槛是 topics **同时**具备 `dsh-plugin` 与 `deepseek-harness`（两者都填，不是冗余）。
  另有 `cyber-moshen` 的 workshop 要求「已发布到 npm 的包 + 一个 PR」，属于最重的一条。
  补两条实测更正：`LivXue/dsh-plugin-shop` 发布的 26,351 条条目里**每一条都带 npm tarball 的 `integrity`、
  `commit` 字段一条也没有**，所以它在包进 npm 之前不会收录纯 git 源——根清单 shim 只是「有资格被它读」
  的前提，不是「会被它列出来」的保证。
- **仓库自己发布一份 format-1 插件索引**（`ef783ba`，新文件 `docs/market-index.json`）：
  <https://raw.githubusercontent.com/jjjjjjjjnnjnn/ai-terminal-pet/main/docs/market-index.json>。
  因为索引格式允许「钉到 40 位 commit 的 git 源」，所以在 npm 发布被一次性密码挡住时，Discover 仍然可用；
  条目只有 `dsh-ai-terminal-pet` `1.2.0`，`source` 为 `{type: git, url, commit: 4d777e32e0a9ef57bb7714287dccd7cae065346f,
  path: packages/dsh-ai-terminal-pet}`，`engines.dsh` 取包里声明的 `>=0.1.0-rc.6`，图标钉在同一个 commit 的
  `icon.png`（4,386 B）。校验用的是宿主自带的 `parseMarketIndex`（`@deepseek-ai/dsh-plugin-market-index`），
  不是手抄一份形状检查：`format: 1`、`localizedText` 含 `en`、category/label 都引用已声明 id、
  `publishTimestamp` 为 Unix 秒、版本精确 SemVer。`marketInstallTarget` 产出的 spec 已实测安装成功
  （`git+…#4d777e3…&path:/packages/dsh-ai-terminal-pet`，全新 `DSH_HOME` + 独立 profile，2.5 s）。

### Changed

- **构建命令统一加 `-buildvcs=false`**：`go build -trimpath -ldflags "-s -w"` →
  `go build -trimpath -buildvcs=false -ldflags "-s -w"`，四处同步——`install.ps1` 的缺失文件提示、
  `.github/workflows/ci.yml` 的两条 build 步骤、`AGENTS.md` 的 Production build、README 的安装步骤。
  **为什么必须加**：不加时二进制里带一块 VCS 戳记，而戳记含构建时的 commit 与工作区状态，所以同一个
  提交在不同时间、不同机器上重建出的产物字节不同；`install.ps1 -VerifyOnly` 只要发现已部署副本与工作区
  里的 `bin/` 是两种构建方式下的产物，就会误报 `[fail]`。二进制本身不入库（`.gitignore` 忽略 `bin/`）。
  **实测差值**：加上之后 `pet-statusline.exe` 由 **3,644,416 B → 3,643,392 B**（差 1,024 B，
  即被去掉的戳记），主程序两种写法同为 **4,371,968 B**。
  按新命令重建后 sha256：主程序 `bbb5163c296cd460d7c9a4c85e10342e125c0bd8017a065f913aa01567d41f0a`、
  shim `31c21739c60e427a387e1360b19395c5f3e9b48fc7597820bcb983a348b82328`；
  `install.ps1 -VerifyOnly` 实测 **10/10 `[ok]`**、末行
  `install: verified - 10 file(s) match the repository, 0 warning(s)`、exit 0。
- **`install.ps1` 三个新开关**，并新增 `Assert-SafeInstallDirectory` 等保护：
  - **`-VerifyOnly`**：只读，绝不写文件。按**同一份 10 项制品清单**（5 个分组：overlay / statusline /
    protocol / bundle manifest / bundle runtime）逐项比对仓库侧与已部署侧的 SHA256，
    每项打印 `[ok]` / `[warn]` / `[fail]`，末尾分组小计；全部吻合才打印
    `install: verified - N file(s) match the repository` 并 `exit 0`，否则 `install: not verified`
    并 `exit 1`。清单只有一处定义（`Get-InstallArtifacts`），因为「校验集 ≠ 安装集」正是过去
    一次升级后残留旧 `protocol.js`、所有上报端静默失联的成因。
  - **`-Uninstall`**：先要求输入 `yes`（无控制台且未加 `-Force` 时 `exit 2`；用户拒绝时 `exit 3`，
    什么都没删），然后停掉浮窗与 shim、删安装目录（5 次重试，防 reporter 中途重新拉起），
    只在 `HKCU\...\Run` 条目**指向被删目录内部**时才删除该条目（指向别的副本的条目属于那个副本），
    并调用 `Remove-InstallDirectoryFromUserPath` 清掉 `-AddToPath` 写入的那条 PATH 条目以避免死条目——
    此时安装目录即将消失，PATH 原值改备份到 `%TEMP%`。PATH 清理失败只报 `[warn]`：文件已经删了，
    残留条目是整洁问题而非失败。
    `Assert-SafeInstallDirectory` 拒绝盘根与任何叶子名不是 `AI-Terminal-Pet` 的目录——
    **`-Force` 不绕过名字检查**，只跳过确认；一个只靠提示词兜底的卸载器离误删只差一次手滑。
  - **`-AddToPath`**：只写 `HKCU\Environment` 的 `Path`，绝不碰 HKLM（工具是按用户装的，
    机器级 PATH 需要管理员权限且会改到这台机器上的所有账号）。写入前把原值（含注册表值类型）
    备份到 `<安装目录>\path-backup.txt`，已在 PATH 中则不改动只回报。
  - 三个开关的互斥组合在参数解析阶段就以 `[fail]` + `exit 2` 拒绝（`-VerifyOnly` 与 `-Uninstall` 互斥，
    `-AddToPath` 不能与这两个同时用）。
- **安装目录的"装完就是好的"改为"随时可证"**：README 的安装步骤多了第 4 步 `-VerifyOnly`，
  并在装完后用 `--doctor` 自查；卸载与 Path 开关同样写进 README 的中英两半。
- **`.gitattributes` 固定 `* text=auto eol=lf`**（`1330fbe`）：仓库本来就按 LF 存取，但 Windows 上
  `core.autocrlf=true` 会一直报规范化警告；钉进树里之后警告消失，Go 源码与 Markdown 在各次克隆间字节一致。
- **插件包版本 1.1.0 → 1.2.0**：清单内容变了（新增 `dsh.catalog`、`keywords` 补 `dsh-plugin`、根 shim
  `version` 同步），所以必须换版本号。1.1.0 的 tarball 已发布且**不可变**，若让版本停在 1.1.0，仓库与那份
  tarball 就会对不上——同一个版本号对应两份不同内容，正是本仓库这一轮一直在消除的漂移。因此
  1.1.0 保留为历史发布、由 1.2.0 取代。
- **npm 凭据问题已解决，但发布仍未发生**：整场会话里 `npm whoami` 一直是 **E401**；2026-11-10T12:20Z
  换过 token 之后它返回 **`jjjjjjjjnnjnn`**（收尾复核时仍然有效）。`npm view dsh-ai-terminal-pet version`
  仍旧报 **E404**，所以这个名字还空着、**一个版本都没发出去**。在包目录跑 `npm publish --dry-run`
  通过（rc=0）：**16 files / 1.9 MB packed / 4.5 MB unpacked**。从**仓库根**发布会撞上 **EPRIVATE**
  （根清单带 `private: true`，npm 抬到的是根清单），从 `packages/dsh-ai-terminal-pet/` 发布则正常
  ——包清单没有 `private`。下一步是发 `1.2.0`。仍未解锁的是 DSH 应用内 Plugins → Discover 索引
  （只表达 npm 版本或钉 40 位 commit 的 git 源）；手工安装靠 git 源、Release tarball、仓库地址三条路。

### Fixed

- **CI 曾经变红在 `vendor:check`，根因不是代码错**（`d06001d` 之后）：`go.mod` 里的 `go 1.23.0` 是
  **下限不是钉版本**，CI runner 用 go 1.23.0、本地用 go1.27.0，两个工具链编出的二进制字节不同——
  **跨工具链字节相等从来不可能达成**，那一步当时是拿一个达不到的条件当门禁。修法是把断言换成能达成的：
  `vendor/manifest.json` 记录**源码摘要**（25 个非测试 `.go` 文件）与两个 vendor 文件自己的哈希，
  `vendor:check` 校验这些；只有当**本地 Go 版本等于 manifest 里记录的 `go1.27.0`** 时才和 `bin/` 做字节比对。
  于是 CI 校验「vendor 与源码摘要一致」、本地额外校验「vendor 与 `bin/` 字节一致」，两边都在验真实的东西。
  这段历史不美化：CI 是先红后修的，不是一次就绿；也不是「go.mod 该钉工具链」。
- **可移植性：`docs/assets/render_assets.py` 不再写死本机绝对路径**（`98 行`改动）。
  旧版把 `PREVIEW` 写成 `C:\Users\<本人>\AppData\Local\Temp\petpreview2`、`ASSETS` 写成
  `C:\Users\<本人>\Desktop\项目开发\ai-terminal-pet\docs\assets`，任何别人的克隆一跑就断。
  现在预演目录读环境变量 `PET_PREVIEW_DIR`、输出目录取 `Path(__file__).resolve().parent`；
  变量未设置或目录不存在时打印中英双语的可执行提示（先跑哪两条 `go test`）并 `exit 2`，
  缺 Pillow 同样给出 `pip install pillow` 后 `exit 2`，而不是抛 Python traceback。
  另外 `sys.stdout`/`sys.stderr` 显式 `reconfigure(encoding="utf-8")`，保证在非 UTF-8 代码页下重定向时中文不乱码。
- **测试夹具里的本机路径全部中性化**：`internal/platform/focus_windows_test.go`（3 处）、
  `internal/platform/presence_windows_test.go`（4 处）、`internal/pipeserver/server_windows_test.go`（1 处）
  原先把 `C:\Users\<本人>\Desktop\项目开发\...` 当作期望值钉在测试里，
  现在换成 `C:\work\项目\...` / `C:/Users/pet-test` 等中性路径。
- **管道名测试改为按算法现算**：`TestPipePathForUserMatchesReporterAlgorithm` 过去硬编码
  `-42ce7d78` 并 `t.Setenv("USERPROFILE", `C:\Users\rongj``)，现在用
  `sha256.Sum256([]byte(`c:\users\pet-test`))` 取前 4 字节转 8 位十六进制得到期望后缀，
  并特意用**混合大小写 + 结尾分隔符**的输入，仍然钉住 Go/JS 共享契约
  （home 小写化 → 去尾分隔符 → 取 sha256 前 4 字节）。
- **删除七处不可达声明**（`3fd49c3`，staticcheck U1000）：`internal/store/store.go` 的 `(*Store).remove`、
  `internal/ui/app_windows.go` 的 `messageEnvelope.reserve` 与 `(*App).clampToMonitor`、
  `internal/ui/canvas_windows.go` 的 `fillPolygon` / `alphaAt` / `pointInside` / `canvasPoint`。
  行为不变，`gofmt` / `go vet` / `go test` 保持干净，没有为了消 warning 加 `//nolint` 或假引用。
  唯一保留的 staticcheck 发现是 `internal/ui/win32_api_windows.go` 里 `syscall.Syscall9` 的 SA1019，
  因为 `UpdateLayeredWindow` 的调用签名需要它。
- **`vendor/AI-Terminal-Pet.exe` 一开始没进 git 树**（修于 `91719e5`）：`.gitignore` 里的 `*.exe` 把它一并忽略，
  而 `git add` 不会自动越过忽略规则，于是这颗二进制只存在于工作区。后果不对称得很有欺骗性——
  `npm pack` 读的是工作区，所以 **release tarball 装得上**；而 **git 源安装拿到的是一个 `vendor/` 里
  只有 `protocol.js` 的包**，`runBootstrap()` 于是回落到下载分支，`lib/release-pins.json` 又还没建，
  回落是 no-op，浮窗就静默地不装了。修法是两行：`.gitignore` 加例外
  `!packages/dsh-ai-terminal-pet/vendor/AI-Terminal-Pet.exe`；`.gitattributes` 加 `*.exe binary`，
  免得 Git 对这颗 sha256 依赖的二进制做文本规范化。
  同一次排障还改正了 Release 的 tag 落点：`dsh-plugin-v1.0.0` 最初打在 `7f476c9`（vendor 还在被忽略的那次提交），
  于是从 tag 安装拿到的是一个 `vendor/` 里只有 `protocol.js` 的包。修完 `.gitignore` 后 tag 已强制移到 `91719e5`，
  并用 `git+...#dsh-plugin-v1.0.0&path:/packages/dsh-ai-terminal-pet` 重装复验（rc=0、3.4 s、`vendor/AI-Terminal-Pet.exe` 4,371,968 B 在位）。

### Docs

- **README 重写并配真实截图**（`7c0aaaa`，双语严格镜像）：安装、CLI、五类终端接入、路线图、非目标、
  从源码开发全部对齐 v0.9.0 之后的事实；`docs/assets/README.md`（`4574cba`）逐个说明 8 张 PNG
  的用途与复现步骤（两张 preview 测试共享同一个 `PET_PREVIEW_DIR`），并标注 3 张未被 README 引用的变体为何保留。
- **停用期 runbook 归档**（`8ada28b`）：`docs/HANDOVER.md` §1–§12（2026-09-24 全局停用态）逐字搬入
  `docs/archive/HANDOVER-2026-09-DISABLE.md`，正文只留指回链接；`docs/archive/README.md` 成为归档索引。
- 早期宣传稿与草稿移入 `docs/archive/promo-2026-11-09/`，不再混在对外文档里。

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
