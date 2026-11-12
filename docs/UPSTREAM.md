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
                   等 2–4 分钟再 npm view dsh-ai-terminal-pet version（头几分钟 registry 只有占位 0.0.0-stage）
6. B 提交          把 docs/market-index.json 的 source.commit 与 icon 钉到 A 的 sha
                   （git 无法在一个提交里引用它自己的 sha，所以必须两步）
7. 上游条目        node tools/upstream-entry.mjs --version X.Y.Z --write --branch bump-tarball-X.Y.Z --pr <号>
                   （从上游 main 重建，只换 tarball 行；给了 --pr 就顺手评论）
8. 复验资产        从公网 URL 下载回来核对字节；沙箱 install → -VerifyOnly → --doctor → -Uninstall
```

**顺序上的硬约束**：`2` 的索引版本与 `6` 的索引 sha 必须分在两个提交；`3`/`4`/`5` 在 `2` 之后；
`7` 在 `4` 之后（否则上游指向一个还不存在的资产）。§3 的工具在写之前会先探一次资产 URL。

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

人工复核制的四家（dshbase / dsh.so / dsh-plugin.org / AllDSH）已经投过、0 评论，
**不要再投**：重复投递不会加快，只会让「我们不看规则」变成印象。

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
[ ] Version + 两份清单 + market-index version/publishTimestamp → A 提交
[ ] npm run vendor；verify.mjs；npm pack 记字节与 sha256 → CHANGELOG
[ ] 双 tag → 双 Release（资产齐全，URL 可下载）
[ ] npm publish，等 2–4 分钟核对 dist-tags.latest
[ ] market-index 的 source.commit/icon 钉到 A 的 sha → B 提交
[ ] node tools/upstream-entry.mjs --version <版本> --write --branch bump-tarball-<版本> --pr <号>
[ ] gh pr view <号>：mergeable=MERGEABLE，check 跑起来（不是「没检查」）
[ ] node tools/upstream-entry.mjs --check      → OK 或 OK（in flight）
[ ] 从公网下载三个资产复验；沙箱 install → -VerifyOnly → --doctor → -Uninstall
[ ] HANDOVER 记本轮：发了什么、上游 PR 号与状态、判可见性读的是哪个文件
```

---

## 8. 还没做完的事（诚实清单）

- 上游 PR **合并仍要人工**：我们只能保证「随时可合」（无冲突、检查跑过、只改一行）。
- 站点渲染有延迟：合并后要等对方重建，`plugins.json` 才换新版本。**我们会读到旧值，这是正常的。**
- 其他目录（dshbase / dsh.so / dsh-plugin.org / AllDSH）都是人工复核，已投过，等对方。
- `dshopen.org` 的线上 `public/index.json` 仍是旧构建，等它自己重建。
