# UPSTREAM — 上架与同步：一次发版要走完的每一条路

> 这份文件回答一个问题：**发完版之后，东西是怎么到用户手上的，哪一步最容易坏，坏了怎么修。**
> 它把散在 [HANDOVER.md](HANDOVER.md) §27 / §28.6 / §38、[ROADMAP.md](ROADMAP.md) §4 与
> 上游 `awesome-dsh-plugin` 的 `contributing.md` 里的规则收成一条可执行的路。
> 上游规则以他们的 `contributing.md` 为准；本文只记录**我们这一侧的动作**，并在每条规则后面写明
> 它是从哪条要求来的。上游改了规则，改这一节，不要凭记忆。

---

## 0. 一句话为什么需要这份文件

**「发版」与「上架」是两件事，而且后者会给前者记账。** 我们每发一个插件版本（`dsh-plugin-v*`），
上游总目录里的**同一个 YAML 文件**必须跟着改一行 `tarball:`；那个文件同时在**别人手里被人工润色**
（描述补前置条件、补默认行为、改类别）。于是「用旧分支开 PR、让 git 三方合并」这条路必然撞车：
`tarball:` 那一行在两边都改了 → 内容冲突 → **冲突的 PR 没有 merge ref → GitHub 连 `check` 都不跑** →
 依赖 `PR check` 完成的 `Submission gate` 也不跑 → PR 看起来「在等检查」，实际是**卡死**，
维护者既没法合，也看不到任何结论。

`#14718` 就是这样卡住的（2026-11-12），维护者 `fkysly` 的原话给了解法：

> 请 rebase 到最新 main（或点 “Resolve conflicts”）：冲突只在 `data/plugins/` 下这个 yml 的
> `description`/`tarball`。**以 main 上的内容为底，把你这次的新增加进去，不要删掉 main 里已写明的
> 前置条件和默认行为。**

所以规矩只有一条：**每次同步都从上游 `main` 重建这一行，永远不合并旧分支。**
这件事已经机械化（§3 的 `tools/upstream-entry.mjs`），不要再手工做。

---

## 1. 上游要什么（`awesome-dsh-plugin` 的硬要求）

逐条抄自 `contributing.md`（2026-11-12 实读），括号里是我们这一侧对应的现状：

| 要求 | 我们怎么满足 |
|---|---|
| 一个插件一个文件：`data/plugins/<owner>__<repo>.yml`；monorepo 子包是 `data/plugins/<owner>__<repo>--<子路径>.yml` | 我们的文件是 `data/plugins/jjjjjjjjnnjnn__ai-terminal-pet--packages-dsh-ai-terminal-pet.yml`，由工具按规则**算出来**（`entryFileName`），不手写 |
| 只提交**这一个文件**，生成的两份 README 不要手改（它们在 `main` 合并后由脚本重生成） | 我们的 PR 永远是 `1 file changed` |
| `description.en` 必填，`zh` 可选（缺了维护者会补） | 我们有中英两份，改动时两边一起改 |
| 描述里有 `: `（冒号加空格）**必须加引号**，否则 YAML 解析失败 | 我们的值整段用单引号包住 |
| 仓库要声明 `dsh.bundle`（只有 `dsh.client` 不算，装不上） | 根 `package.json` 与 `packages/dsh-ai-terminal-pet/package.json` 都有 `dsh.bundle` |
| 仓库要有真实可用的代码、非占位 | 是 |
| 仓库至少 **1 天**（这个 CI 自动查） | 是 |
| 仓库活跃维护（长期沉寂会被扫进 tracking issue 后移除） | 每轮都发版 |
| 仓库要有 `dsh-plugin` topic | 已有（`gh api repos/jjjjjjjjnnjnn/ai-terminal-pet --jq .topics` 里含 `dsh-plugin`） |
| 描述**必须是可核对的事实**，不许营销词，描述里提到的数字/API/命令都要真的存在 | 见 §5「描述怎么写」 |
| `category` 选最接近的（选偏了维护者直接改，不驳回） | 我们是 `notify` |
| `tarball:` 必须是 **GitHub 自己托管的 https `.tgz`**；`latest/download/` 的文件名**不能带版本号**，要带就**钉 tag** | 我们**钉 tag**：`releases/download/dsh-plugin-v<版本>/dsh-ai-terminal-pet-<版本>.tgz`——资产名带版本号，所以必须钉 tag，这也正是工具默认的形态 |
| `icon:` 必须是 **GitHub 托管的位图**（png/jpg/jpeg/webp，约 96px 方形），SVG 与第三方图床一律拒 | 我们用仓库里的 `icon.png`（raw.githubusercontent.com） |
| 截图放在**我们自己仓库**的 `screenshots.json`（1–8 张，相对路径，不能 `..`、不能以 `/` 开头） | `packages/dsh-ai-terminal-pet/screenshots.json` |
| npm 可选；`repository` 字段必须指回**这里列出的仓库**；yml 里**不许**写 `npm:` 键 | 包 `repository` 已指回本仓库；yml 里没有任何 npm 字段 |
| 一个 PR 最多 3 条 | 我们永远 1 条 |
| PR **不许碰别人的条目**（CI 会把改到的每一条列出来质问） | 只改自己的文件，且工具断言「只有一个文件、只有一行」 |
| 接受 PR 的对象：描述修正、换类别、移除失效项目 | 我们的跟进 PR 就是描述修正 + `tarball` |

**CI 会查什么**（`check` 作业）：PR 条目数 ≤3、`dsh.bundle`、仓库年龄、`awesome-lint` 与站点构建
（双语对齐、分隔符、日期、截图）。**`check` 通过只是前提，不是结论**：维护者会**读我们的源码**核对
描述里的每一句。

**`held: author` 标签**：维护者要求改动时打上，PR 停在合并队列里（不是拒绝、不是坏掉）。
修法只有一条：**把改动推到同一个分支，然后留一条评论说明改了什么**；标签**不会自动消失**，
只能由人复审后摘掉。反复推、「怕万一」再评论一次都快不了。

---

## 2. 一次发版的完整顺序（A/B 两提交 + 四条分发）

版本常量在 `internal/app/app_windows.go`（浮窗线 `pet-v*`）与
`packages/dsh-ai-terminal-pet/package.json`（插件线 `dsh-plugin-v*`）。两条线各自发版、各自打 tag，
**资产不可变**：连改一句措辞都要耗一个新版本号。

```
1. 代码与文档      Version → X.Y.Z；三二进制重编（-trimpath -buildvcs=false -ldflags "-s -w"）
                  → npm run vendor（重算 vendor/manifest.json 的源码摘要）
2. A 提交          版本 + 两份清单 + docs/market-index.json 的 versions[0].version 与 publishTimestamp
                  + CHANGELOG + README/ROADMAP/HANDOVER
                  跑 node packages/dsh-ai-terminal-pet/verify.mjs（不写死条数）
                  npm pack 取 tarball，记下字节与 sha256，写进 CHANGELOG
3. 双 tag          git tag -a pet-vX.Y.Z / git tag -a dsh-plugin-vX.Y.Z，推上去
4. 双 Release      gh release create <tag> <资产> --title … --notes …（GitHub 自己算 digest）
                   浮窗：ai-terminal-pet-X.Y.Z-win-x64.zip + pet-setup.exe（scripts/package.ps1 产出）
                   插件：dsh-ai-terminal-pet-X.Y.Z.tgz
5. npm             cd packages/dsh-ai-terminal-pet && npm publish --access public
                   cd packages/opencode-ai-terminal-pet && npm publish --access public   ← 与浮窗同号
                   等 2–4 分钟再 npm view <name> version（头几分钟 registry 只有占位 0.0.0-stage）
                   先 npm pack，再 node tools/check-tarball.mjs --package packages/opencode-ai-terminal-pet
                   （包内 server.js / protocol.cjs 由 reporters/ 生成，所以要 bump 后重跑 sync.mjs --write）
6. B 提交          把 docs/market-index.json 的 source.commit 与 icon 钉到 A 的 sha
                   （git 无法在一个提交里引用它自己的 sha，所以必须两步）
7. 上游条目        node tools/upstream-entry.mjs --version X.Y.Z --write --branch bump-tarball-X.Y.Z --pr <号>
                   （从上游 main 重建，只换 tarball 行；给了 --pr 就顺手评论）
8. 复验资产        从公网 URL 下载回来核对字节；沙箱 install → -VerifyOnly → --doctor → -Uninstall
```

**顺序上的硬约束**：`2` 的索引版本与 `6` 的索引 sha 必须分在两个提交；`3`/`4`/`5` 在 `2` 之后；
`7` 在 `4` 之后（否则上游指向一个还不存在的资产）。§3 的工具在写之前会先探一次资产 URL。

**批次发版（2026-11-17 起）**：同一主题攒一次发布，**每周 ≤2 次**（安全与回归修复除外）。
理由不是流程洁癖：每次发布都会给 npm 加一个版本，而抓取队列会按版本各取一次，
于是"下载总量"被自己的发版节奏推高、`latest share`（唯一诚实的指标，见
[docs/METRICS.md](METRICS.md)）被压低，上游目录的 bump PR 也变成噪声。近 7 天 222 个提交、
53 个 release 就是这条规矩的来源。

**`4` 一推上去，CI 就会先红一次，这是顺序决定的，不是回归**：`verify` 作业里的
`node tools/upstream-entry.mjs --check` 判断的是"此刻有没有人开 PR"，而 PR 只能在本版本资产存在之后才开
（`7` 依赖 `4`）。所以 release 提交那一次运行会是 `STALE and NOTHING PENDING`；
`7` 做完之后**重跑同一条 job**，它会因为 PR 在飞而报 `OK（in flight）` 并转绿。
1.2.23 就是这样走的（HANDOVER §66.6）：Run 红在 upstream 那一步，PR #16010 开好后重跑即绿。
不要为了让它一次变绿而把这一步改成 `|| true`——那正好删掉了"发了版却没人推上游"的唯一警报。

---

## 3. 上游条目：为什么用工具，不用手

```powershell
# 只问状态：上游条目是不是已经指到我们最新的 tag？（离线/无 gh 时是 SKIP，不是通过）
node tools/upstream-entry.mjs --check

# 看一眼它要写的 URL 与资产探测结果
node tools/upstream-entry.mjs --version 1.2.9

# 真正重建：从上游 main 取文件 → 只换 tarball 行 → 断言 → 签名提交 → force-push 分支 → 评论 PR
node tools/upstream-entry.mjs --version 1.2.9 --write `
  --branch bump-tarball-1.2.9 --pr <PR 号> --comment-out .local/comment.md
```

它做了什么，每一步都是**断言**而不是「应该没问题」：

| 步骤 | 拒绝条件（拒绝就停下，不推） |
|---|---|
| 资产探测 | 目标 `.tgz` 不是 200/302 → 停（钉的 tag 或资产名写错了） |
| 取上游文件 | 上游 `main` 读不到 → 停（不是「大概能用」） |
| 换行 | 文件里 `tarball:` 行 **不是恰好一行** → 停（0 行 = 没得换；2 行 = YAML 会静默取最后一行） |
| 换行 | 值里有 `:` `#` 引号 → 停（那需要加引号，真实 release URL 不会这样） |
| 换行 | 文件是 CRLF → 停（clone 用 `core.autocrlf=false`，CRLF 说明拿到的不是 git 会哈希的那份字节） |
| 回滚证明 | 把那一个字符串**换回去**必须与上游**逐字节相同**；换不回去 = 带了别人的改动 → 停 |
| diff 断言 | 改动文件数必须 = 1、numstat 必须 = `1 1 <我们的文件>` → 否则停 |
| 提交 | 用本机身份，本机配了 SSH 签名就签（`gh api` 复核过 `verified: true / valid`） |
| 推送 | `--force-with-lease`，只推到自己 fork 的 `bump-tarball-<版本>` 分支 |

`tools/upstream-entry.test.mjs`（`node --test tools/upstream-entry.test.mjs`）把上面这些拒绝条件
全部钉成单测，纯函数、不联网、不碰 git。

**CI 里怎么跑**：`.github/workflows/ci.yml` 的 `verify` 作业跑 `node tools/upstream-entry.mjs --check`
（版本默认取 `packages/dsh-ai-terminal-pet/package.json`）。它读三个外部来源：本版本的 Release 资产 URL、
上游 `main` 的条目文件、以及「有没有针对这个版本的 `bump-tarball-*` PR 开着」。四种结论都是**故意的**：

- `OK`：上游已经指到我们这个版本；
- `OK（in flight）`：上游还旧，但有对应 PR 开着——计入成功，因为**上架要等维护者合并，不是我们能这一侧
  控制的事**，把它做成红灯只会训练人忽略红灯；
- `STALE and NOTHING PENDING`：**发了版、上游还旧、而且没人开 PR** → **红灯**，这正是「忘了同步」；
- `SKIP`：版本还没发布（资产 404）/ 读不到上游 / 读不到 GitHub API（网络或权限）。
  **skip 不是 pass**，输出里明写，不会假装通过。

判断「有没有 PR 开着」用的是公开 REST（`GET /pulls?head=<owner>:<branch>`），**不依赖本机 `gh`**：
CI 里没有 gh 登录，一个仓库范围的 token 也不保证能读别人的仓库；有 `GH_TOKEN` / `GITHUB_TOKEN` 就带上
（只是提高限速额度），没有也照常工作。

---

## 4. 上架之后：什么叫「已经可见」

三类路径读的字段不同，**「合并」≠「可见」**，判断可见性要读对方**站点域**的索引文件：

| 路径 | 它读什么 | 状态查看方式 |
|---|---|---|
| npm | 包名、版本、`dist.integrity` | `npm view dsh-ai-terminal-pet version` |
| awesome-dsh-plugin（dshmarket 的 Market 页由它生成） | `data/plugins/*.yml` → 站点 `plugins.json` | PR 合并后等它的站点重建，再读 `https://awesome-dsh-plugin.com/plugins.json` |
| 本仓库的 format-1 索引（Discover 可直接加） | `docs/market-index.json` | 读 raw 的线上文件 |
| 采集型市场（topic + 根清单） | 仓库根 `package.json` 的 `name` / `dsh.bundle` | 无需 PR，它们自己扫 |

我们碰过的目录与状态**记在 [HANDOVER.md](HANDOVER.md) §1 与 §38.5**，不在这里抄：那些数字会变，
抄一遍就是下一次的过期文档。判断标准只有一条：**读对方公开的索引文件，不读 PR 页面**。

### 4.1 一条命令回答"我们那行现在到底怎么样"：`tools/catalog-status.mjs`

```powershell
node tools/catalog-status.mjs            # 线上：站点索引 + npm + 本工作区
node tools/catalog-status.mjs --json     # 机器可读
node tools/catalog-status.mjs --file f.json   # 离线：只判我们自己能改的那一半
```

它把三种**看起来一样、处理方式相反**的情况分开，这是它存在的唯一理由（2026-11-16 实测）：

| 判定 | 含义 | 我们该做什么 |
|---|---|---|
| `upstream` | 站点那行落后于 npm（`row says 1.2.16, npm has 1.2.26`），或 30 天下载数正好少了窗口最后一天 | **什么都不做**。前者是对方站点构建失败，后者是对方探针在 03:1x UTC 跑、npm 还没聚合完前一天（issue #15864）；重发版、重探都不会改变它 |
| `pending` | 离线判定：入口的 `tarball` 落后于本工作区 | 跑 `node tools/upstream-entry.mjs --check` 看 PR 是在飞还是根本没开 |
| `fail` | 入口 `tarball` 落后于 npm（线上判定） | 开一条 `+1 −1` 的 bump PR |
| `ok` | 那行是当前版本 | 无 |

退出码：`ok` / `upstream` / `pending` 都是 **0**（后两者不是我们的失败），只有 `fail` 是 1，
站点索引读不到是 2。CI 里这条是 **informational**（`|| true`），因为对方站点挂了不该让本仓库变红。

**两次真实测量（2026-11-16，都用它跑出来的）**：

- 我们的 `tarball` 早已是 1.2.21（#15674 已合并），而站点 `version` 停在 **1.2.16**、
  入口页的安装命令还写着 `dsh-plugin-v1.2.16` —— 原因是对方**整站在 2026-11-14T05:27Z 之后冻结**：
  2026-11-15 与 11-16 两次夜间 `Build` 都死在
  `MarketIndexError: $.plugins[N].name duplicates dsh-dot-mcp`（对方 issue **#15569**，修复 PR **#15579**）。
  所以「合并」与「重建」是两件事，**重建**现在也还不是我们能催的。
- 30 天下载数 `1179` 不是"1.2.21 之前的老数据"，而是**两段缺失之和**，而且两段原因不同：

  | 缺的部分 | 值 | 原因 |
  |---|---|---|
  | 窗口**自己的最后一天**（2026-11-13） | **1,197** | 探针在 03:1x UTC 跑，npm 对刚结束 2.5 小时那天还没聚合 → 那天返回 0。**全站 12/12 条目都是同一签名**（对方 issue **#15864**） |
  | 窗口**根本没前进到** 11-14 / 11-15 | **1,079**（318 + 761） | 整站冻结，`end` 停在 11-13（#15569 那一项） |
  | 站点发布值 | **1,179** | = 11-11 的 624 + 11-12 的 555 |

  **1,179 + 1,197 + 1,079 = 3,455**，而 3,455 正是 npm 自己页面上的 Weekly Downloads 与
  `point/last-month` 的返回值（`2026-10-17 → 11-15`）—— 三段拼回去正好吻合，所以这不是推测，
  是可以用两条 `curl` 复现的分解：

  ```sh
  curl -s 'https://api.npmjs.org/downloads/range/2026-11-11:2026-11-15/dsh-ai-terminal-pet'
  # 624 / 555 / 1197 / 318 / 761  -> 合计 3455
  ```

  **"7 天下载 = 30 天下载"不是巧合**：首个版本 `1.2.4` 是 2026-11-11T12:10:46Z 才发布的，
  11-11 之前那 30 天窗口里每天都是 0，所以两个窗口的真实值相同。

- **决策（2026-11-16）**：这两条上游 bug 已有多位开发者在那两条 issue 下反馈，
  我们**不再跟进、不发 PR、不催合并**。要做的是**把自己这边的记录改准**：
  任何引用下载量的地方都用 npm 自己的数（并写清窗口），不用站点那个被两条 bug 压低的数。

人工复核制的四家（dshbase / dsh.so / dsh-plugin.org / AllDSH）已经投过、0 评论，
**不要再投**：重复投递不会加快，只会让「我们不看规则」变成印象。

### 4.2 2026-11-16 晚：冻结的根因已被上游修掉，重建正在跑（我们仍然什么都不做）

- **16:39:48Z**：修复 PR **#15579** 被 **MERGED**；**16:39:49Z**：issue **#15569** 被 **CLOSED**。
- **根因与我们记的不完全一样**（维护者 Eapp1e 在 #15569 里的回复，逐字摘要）：不是探针失败，而是
  `data/plugins/Tests4gptplay__dsh-dots.yml` 与 `...--packages-dsh-dot-mcp.yml` 这两条**各自都合法**的条目
  （monorepo 根 + 它的一个子目录）的 `package.json` 都叫 `dsh-dot-mcp`。format 1 按 `name` 建键，
  解析器于是**拒掉整份文档** → 夜间 `Build` 失败 → 后面 `deploy-pages` 与两个 `npm publish` 步骤全部跳过，
  于是站点与 npm 目录一起停在 2026-11-14T05:27Z。**重跑探针救不了**，输入本身没问题。
  #15579 把这条冲突降级成「只拒掉那一行（后出现的那条），其余照常构建」，`skipped.refused` 里记是谁占了名字。
- **16:39:59Z**：维护者手动 `workflow_dispatch` 了一次 `build-site.yml`（run **`43910005122`**），另有一条
  push 触发的 `43910024003`。历史耗时 1h20m–2h50m，所以目录与 npm 包会在当天晚些时候前进到 2026-11-16 的戳。
- **这对我们意味着什么（关键的一条）**：那行的 `version` 来自 **npm 探针**，不是入口 yml 里的字段，
  所以**重建本身**就会把 `version` 刷到 1.2.26、把下载窗口往前推；而 `tarball`（也就是入口页那条安装命令）
  仍要等 **#16010** 合并 + 再来一次重建才会从 1.2.21 前进。
- **下载窗口那条（#15864）仍无维护者回复**，但 **#15982** 里另一位作者在同一窗口独立量到同样的缺失
  （`dsh-rea-pro`：目录记 458，npm 同窗口实际 697，+52%；另有几个大盘包只差 1–4%），
  说明这条不是我们一个条目的偶然。**决策不变**：不催合并、不发 PR，只把记录写准。
- **下一次怎么查（一行）**：`node tools/catalog-status.mjs` 会给出 current / upstream / pending / fail
  和「该谁动」；要确认对方那两次重建有没有跑完，用
  `gh run list --repo awesome-dsh-plugin/awesome-dsh-plugin --workflow build-site.yml --limit 3`。

### 4.3 winget：它不是一个"目录"，而是 Windows 自己的那一行

`microsoft/winget-pkgs` 不是市场目录，是**清单仓库**；真正的可见性来自用户机器上的
「设置 › 应用」那一行。所以这一条与前面的目录有一个根本区别：**能不能装、能不能卸，由我们的安装器决定。**

- 清单在 [`packaging/winget/`](../packaging/winget/README.md)，布局就是 winget-pkgs 的
  `manifests/j/jjjjjjjjnnjnn/AI-Terminal-Pet/<版本>/`（三份 YAML）。
- 1.0.15 起 `scripts/install.ps1` 写 `HKCU\...\Uninstall\AI-Terminal-Pet` 并把 `install.ps1`
  复制进安装目录，因此静默安装（`pet-setup.exe --install`）与静默卸载
  （那一行的 `QuietUninstallString`）都成立。**这是 winget 能接这个包的前提，不是附加项。**
- 本地校验用 `winget validate --manifest <目录>`（必须打印"清单验证成功"），
  `winget show --manifest <目录>` 可以肉眼核对解析结果。
- 提交是一个只**新增**一个版本目录的 PR；他们的 pipeline 会下载 exe、核对 sha256、跑静默安装与卸载。
- **2026-11-17 的实际状态**：PR [microsoft/winget-pkgs#481818](https://github.com/microsoft/winget-pkgs/pull/481818)
  （`jjjjjjjjnnjnn.AI-Terminal-Pet` 1.0.15，3 文件 `+60 −0`，`MERGEABLE`）。
  - **CLA 已签**（`license/cla` pass）；
  - **`01`–`10` 号校验全部 SUCCESS**（30 项 skipped 是他们的 fork 骨架，11 项真检查全过），
    其中 **`08. Installation Validation` 真的下载并静默安装了 `pet-setup.exe`、再静默卸载**；
  - 标签：`New-Package` + `Validation-Completed` + `Azure-Pipeline-Passed`；
  - **机器人已开启 auto-merge（squash）**，`reviewDecision: REVIEW_REQUIRED` ——
    只等一位**社区志愿维护者**批准。他们的机器人原话："Our moderators are community volunteers,
    please be patient and allow them sufficient time to review your submission."
  **所以现在什么都别做**：不催、不重开、不改 PR 正文（改了会让校验重跑）。
- **一条实测结论（值得记下来）**：`08` 通过说明**未签名并没有拦住他们的静默安装**——
  `docs/UPSTREAM.md` §4.3 里那个"SmartScreen 可能让校验失败"的风险在这一版**没有发生**。
  这不等于"用户不会看到 SmartScreen"：用户第一次双击 `pet-setup.exe` 时仍然会看到提示，
  README 的 Unsigned binary notice 仍然有效。
- **判断可见性**：`winget search jjjjjjjjnnjnn.AI-Terminal-Pet`
  （合并后要等索引同步，通常几小时）；**不要**拿 PR 状态当"已上架"。
- **一个已知风险，写在这里免得下次重新发现**：`pet-setup.exe` 未做 Authenticode 签名，
  自动校验时 SmartScreen 有可能拦住 — 若如此，记录原因、保留安装器改进，**不要**改成
  `InstallerType: zip` + `portable` 的降级清单（那会绕过 `install.ps1`，没有 `protocol.js`、
  没有 bundle、没有自启动，见 [docs/ROADMAP.md](ROADMAP.md) §5）。

---

## 5. 描述怎么写（这条最容易让我们被退回）

上游维护者的第 6 条复审项是**「描述是否属实」**，他们会读源码核对；第 4 条是**「是否已被重复收录」**；
第 6 条还有一条历史（[#1348](https://github.com/awesome-dsh-plugin/awesome-dsh-plugin/issues/1348)）：
一个 PR 改了别人的条目却全绿通过。

所以：

1. **只写现在为真的事**，且必须能从代码里核出来。写完问一句：这句话在仓库里对应哪个函数？
2. **不写营销词**（fast / powerful / better），写行为。
3. **前置条件与默认值必须写明**，因为读者是按这一行决定装不装的：
   - 装上这个 bundle，**DeepSeek Harness 的会话立刻汇入**；
   - OpenCode / Claude Code / Antigravity **要先装各自的上报端**；
   - 对 DSH 会话，面板**只显示状态并跳窗口**，**授权仍在 DSH 里完成**；
   - auto-approve **默认关**。
4. **手势要写准**。我们这一轮就抓到一句错的：`Click a row to jump to that terminal` ——
   实际是**点行会收起面板**，只有**行尾箭头**才跳转。已经改成
   `Click the arrow at the end of a row to jump to that terminal`（zh：`点一下某一行的行尾箭头即可跳到对应终端`）。
   以后面板真的加了「点行做某件事」，再改这一句，**不要同时留两种手势的说法**。
5. 中英两份**逐句对应**，改一边就必须改另一边（上游会把双语不一致当排版缺陷）。

---

## 6. 踩过的坑（同一天内真实发生过，别再踩）

| 坑 | 表现 | 怎么避开 |
|---|---|---|
| 用旧分支开 PR 让 git 三方合并 | `tarball` 行两边都改 → 冲突 → **没有 merge ref** → `check` 与 `Submission gate` 都不跑 → PR 卡死 | 从上游 `main` 重建（§3 的工具）；**永远不合并旧分支** |
| 以为「PR 开着 = 已收录」 | 上游 `tarball` 还是旧版本，站点渲染的 install 命令也是旧版本 | 读对方站点索引；PR 只是请求 |
| PowerShell 5.1 读写中文 | `Get-Content -Raw` 把 UTF-8 当 ANSI，整行 mojibake，写回去就毁了中文 | **用 base64 上传原始字节**（`[IO.File]::ReadAllBytes` → `ToBase64String`），不要先读成字符串再拼 |
| `latest/download/` + 带版本号的资产名 | 提交当天有效，下次发版立刻 404 | **钉 tag**（我们本来就钉），或资产名去掉版本号 |
| yml 里写 `npm:` 键 | 上游校验直接拒 | npm 映射他们从 registry 自动取，别写 |
| 手改生成的两份 README | 冲突、改到邻居行 | 只改 `data/plugins/<自己那一个>.yml` |
| 改了别人的条目 | 全绿也能过，但会被维护者质问（#1348） | 只改自己的文件；工具断言「一个文件一行」 |

---

## 7. 一页清单（发版当天照着打勾）

```
[ ] 代码/测试/素材/文档就绪，全门禁绿（AGENTS.md「标准验证命令」）
[ ] Version + 两份清单 + opencode 包清单（同号）+ market-index version/publishTimestamp → A 提交
[ ] 版本号与随包 README **都改完**之后再 `npm pack`；`node tools/check-tarball.mjs` 必须与工作区逐文件一致
[ ] `scripts/sync.mjs --write` 已在 reporters/ 改动之后跑过（CI 的 --check 会拦）
[ ] npm run vendor；verify.mjs；npm pack 记字节与 sha256 → CHANGELOG
[ ] npm pack（opencode 包）+ check-tarball --package …；三平台 server.test.mjs 绿
[ ] 双 tag → 双 Release（资产齐全，URL 可下载）
[ ] npm publish 两个包，等 2–4 分钟核对 dist-tags.latest
[ ] market-index 的 source.commit/icon 钉到 A 的 sha → B 提交
[ ] node tools/upstream-entry.mjs --version <版本> --write --branch bump-tarball-<版本> --pr <号>
[ ] gh pr view <号>：mergeable=MERGEABLE，check 跑起来（不是「没检查」）
[ ] node tools/upstream-entry.mjs --check      → OK 或 OK（in flight）
[ ] 从公网下载三个资产复验；沙箱 install → -VerifyOnly → --doctor → -Uninstall
[ ] 真机装一次 opencode plugin add（临时 XDG_CONFIG_HOME 先试，再换本机），确认 --doctor 的 reporters 行变 ok
[ ] packaging/winget 加新版本目录；winget validate 通过；向 microsoft/winget-pkgs 开 PR（只加这一个目录）
[ ] HANDOVER 记本轮：发了什么、上游 PR 号与状态、判可见性读的是哪个文件
[ ] node tools/check_versions.mjs 绿（徽章/发布链接/安装命令/基线行都与代码里的版本一致）
```

---

## 8. 还没做完的事（诚实清单）

- **winget（1.0.15 起）**：清单已写在 `packaging/winget/`、`winget validate` 通过，
  PR 已开：[microsoft/winget-pkgs#481818](https://github.com/microsoft/winget-pkgs/pull/481818)（见 §4.3）。
  **`01`–`10` 号校验全过（含真跑静默安装/卸载的 `08`），auto-merge 已由机器人开启**；
  现在只等一位社区志愿维护者批准。在他的批准之前，"能被 winget 安装"仍不成立。
- **1.0.23 的上游条目**：[#16719](https://github.com/awesome-dsh-plugin/awesome-dsh-plugin/pull/16719)
  已开（只改 `tarball` 到 1.2.36；上一版 #16703 已关闭），等维护者合并；合并后还要等对方站点重建才在目录页可见。
- 上游 PR **合并仍要人工**：我们只能保证「随时可合」（无冲突、检查跑过、只改一行）。
- 站点渲染有延迟：合并后要等对方重建，`plugins.json` 才换新版本。**我们会读到旧值，这是正常的。**
- 其他目录（dshbase / dsh.so / dsh-plugin.org / AllDSH）都是人工复核，已投过，等对方。
- `dshopen.org` 的线上 `public/index.json` 仍是旧构建，等它自己重建。
