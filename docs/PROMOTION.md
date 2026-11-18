# 让人看见它

> 目标一句话：**你的 agent 在等你，而你在另一个标签页里。**
>
> 这份文件不是可粘贴的稿子。调研下来，几条硬规则决定了分工：HN 与 V2EX 明禁 AI 文本，
> awesome 列表要求**人手提交**。所以下面的顺序是「先让人能搜到、能一键装」，再谈发声。

---

## 0. 分工（规则决定的，不是懒）

| 事情 | 谁做 | 依据 |
|---|---|---|
| README 首屏、仓库描述、topics、社交预览图、issue 模板、图与 GIF | **我做** | 属于文件与设置，不是社区发言 |
| **npm 上的 OpenCode 插件包**（`opencode plugin add …`） | **我做** | 这是一条真实的安装路径，不是宣传 |
| awesome 列表的收录申请 | **你点表单**，我写那一行描述 | awesome-claude-code：「resource recommendations must be created by human beings」，且只能用网页表单 |
| Show HN / V2EX 正文 | **你写** | HN：“don't put generated text in HN posts”；V2EX：「请不要把 AI 生成的内容发送到这里」 |
| X / 即刻 / 小红书 | 我起稿，**你改语气** | 没有禁 AI 规则，但读起来还得像你 |
| 发帖、回评论 | 你 | 账号与语气都是你的 |

HN 创始人 dang 的原话（2020，2026-03 更新）值得记住：

> Write your text by hand. Don't use an LLM to generate any of it (not even a tiny bit, including
> to edit or spruce it up). … LLM language leaves imprints on your text which are generating quite
> some backlash. **This is a big dividing line at present!**

所以：**不给成稿，只给事实与结构**。这不是保守，是这条渠道的实际规则。

---

## 1. 渠道、门槛、顺序

> **先读 [docs/METRICS.md](METRICS.md)**：截至 2026-11-17，外部用户数为 0，目录排名与 npm 下载量
> 都是机器流量的红利。下面的顺序是按"每单位努力换来的可见性"排的，不是按"已经有多少人"排的。

| 顺序 | 渠道 | 真实门槛（读来的） | 内容 | 谁写 |
|---|---|---|---|---|
| **1** | **awesome-claude-code**（58k★） | **入口已关闭**：2026-11-18 实测，该仓库把 issue 创建限制为 **collaborators only**，表单页返回 `Action restricted`。资格（≥14 天且活跃）满足也没用——普通账号连表单都打不开。规则另写明推荐必须由人创建 | 那一行描述在 §6，**等对方开放入口**再用；不要用 PR / `gh` 绕过 | 等外部（原先的"你提交"已无入口） |
| 2 | GitHub 仓库页 | 无 | 描述/topics/社交预览/README 首屏 —— 决定"点进来 30 秒懂不懂"；首屏的"一行装 + 它不是什么"2026-11-17 已改 | 我 |
| 3 | **npm 上的 OpenCode 插件** | ✅ 已发布：`opencode-ai-terminal-pet`（与浮窗同号，1.0.19） | `opencode plugin add opencode-ai-terminal-pet` 一行装 —— 对目标人群最有用的一步 | 我 |
| 4 | **awesome-opencode**（11k★） | **实际门槛（它的 contributing.md 原文）：≥25 stars，且仓库满 30 天；任一规则不过，PR 会被 close**。我们 1★ → **现在投了就是被关**。这一条等的是 star，不是等包 | 等过线再投，一条 `<details>` | 我起稿，你提 PR |
| 5 | DSH 插件目录 / npm | 已有 | 生态内先看到 | 我 |
| 6 | winget | 等合并 | `winget install` | 我 |
| 7 | X / 即刻 | 你的号 | 5 秒动图 + 一句话 | 我起稿你改 |
| 8 | **V2EX 分享创造** | 需账号；禁 AI 文本 | 讲你自己那次白等 | **你** |
| 9 | **Show HN** | 账号要有历史（新号发链接基本沉）；要能直接跑；不许求赞/删了重发 | 讲你怎么做的、哪里不一样 | **你** |
| 10 | Reddit | **先读每个版侧栏**；r/SideProject 的站级 spam filter 会移除低 karma 账号的推广帖（2026-11-18 实测 `removed_by_category: reddit`），r/coolgithubprojects 允许直接发 GitHub 链接 | r/coolgithubprojects 已发一条；r/SideProject 转 modmail 等放行。**不要同一天连发第三个版** | 我（账号 `Visible_Swim471`），评论由你回 |

> 顺序的道理：**先让"能被搜到"和"点进来 30 秒懂"成立，再引人来**。反过来就是把半成品摊在门口。
> 另：一周内不要在多个社区发同一段话，这些人互相都看得到。
>
> **2026-11-17 调整**：awesome-claude-code 从第 4 位提到第 1 —— 它现在是"只差一次人手提交"的状态，
> 而 awesome-opencode 卡在 25★（不是我们能单方面推动的）。顺序按"这一步还缺谁"排，不按名气排。

---

## 2. 事实卡（写帖时从这里取，别超过）

- 是什么：一条 56px 置顶窄条，显示哪个 AI 终端卡住、在等什么、等了多久；点行尾箭头回到那个终端。
- 支持：DeepSeek Harness（装上即接）、OpenCode（`opencode plugin add opencode-ai-terminal-pet` 一行）、Claude Code、Antigravity（后两个是配置钩子，`--connect` 会打印片段）。
- 平台：Windows / Linux（X11）/ macOS；单文件 Go 二进制，无运行时。
- 不联网：无账号、无遥测、无云；审批仍在你的终端里点。
- MIT；约 4.5 MiB；只在真事件上响（等待确认 / 出错 / 你不在看时完成）。
- 小细节（读者会喜欢）：点行内容复制项目路径；右键行静音；设置界面底部有 Star 与反馈入口。

**必须自己说的三条边界**（说了加分，不说会被当吹牛）
1. macOS 面板在 CI 里构建和测试，**没在苹果硬件上开过窗**。
2. Windows 安装包**未签名**，第一次双击会看到 SmartScreen。
3. 没有遥测，所以没有任何"多少人在用"的数字。

---

## 3. Show HN：结构给你，句子你写

标题（改一个，别照抄）：

- `Show HN: A 56 px always-on-top strip that shows which AI coding agent is waiting`
- `Show HN: I kept leaving Claude Code waiting 15 minutes for a permission prompt`
- `Show HN: ai-terminal-pet – see which agent is blocked, click, get back to it`

正文按这个顺序，各 1–3 句，**不要写成小标题列表**：

1. 你遇到的真实场景（几个终端、哪次白等了多久、当时在干什么）
2. 它具体是什么（多宽的条、每行有什么、点哪里会发生什么）
3. 哪里不一样（只报真事件；行尾箭头直接跳转；审批不离开终端）
4. 怎么试（一条命令/一个安装包，不要注册）
5. 诚实边界（macOS 未上真机、安装包未签名，各一句）

会被问到的（提前想好答案，别现编）：

- Windows 之外真的能用吗？→ 面板三平台，Linux 需 X11，macOS 面板未上真机。
- 为什么要单独一个进程？→ 上报端是 hook；浮窗按需拉起，无会话时自毁。
- 会不会偷看我的 prompt？→ 只发状态/项目路径/一行 reason；9 类机密在落盘前脱敏。
- 和终端自带通知的区别？→ 行级跳转 + 未读/静音 + 只在真事件响。
- 能支持别的终端吗？→ `generic.js` 与 `--connect` 就是给这个的。

---

## 4. V2EX / 即刻：结构同上，口吻更日常

- 标题用问题，不用产品名：`你的 AI 在等你，你在另一个标签页`
- 第一段写你那次真实的等待；不要"随着 AI 编程的普及"这种开头。
- 第二段写它长什么样、点了会发生什么（一两句）。
- 第三段给地址，加一句 macOS 未上真机、安装包未签名，结束。
- 不求 star，不放"欢迎体验"。

---

## 5. X / 即刻：我起的稿（你改语气再用）

```
你的 AI 在等你，你在另一个标签页。

四个终端开着的时候，最浪费时间的不是它写错，是它卡在"要不要允许"上，而你在看别的窗口。
一条 56px 的置顶窄条：哪个终端、等什么、等了多久；点行尾箭头直接回去。

Windows / Linux / macOS · MIT · 不联网、不要账号
<链接>
```

配 `docs/assets/demo.gif`（5 秒：呼吸 → 展开 → 点箭头）。

---

## 6. awesome-claude-code 那**一行**（单行、无 emoji、描述不推销）

> Always-on-top strip showing which AI coding agent is waiting, and a one-click jump back to that terminal; no account, no network, no telemetry.

提交方式：仓库 README 里的 **网页表单**（不是 PR、不是 `gh`），一次只荐这一个。
表单地址在 `hesreallyhim/awesome-claude-code` 的 README / CONTRIBUTING 里（2026-11-17 读到的规则：
资源推荐必须由人创建，且只有网页表单入口）。

**2026-11-17 的资格复核**：门槛是「≥14 天且活跃」或「≥100★」。本仓库 2026-09-24 建立（**55 天**），
当天仍有提交 → **第一档已满足，不需要 star**。填表时如实写：单人维护、无遥测、Windows/Linux/macOS、
MIT；不要写下载量（见 §9）。

**唯一需要你做的**：把上面那一行粘进表单的 description 字段，其余按表单问什么答什么。不要让它"更
有感染力"——它就是要一行、无 emoji、不推销。

---

## 7. 「一眼看出是 AI 写的」检查表

中三条以上就重写——这不是修辞问题，是可信度问题：

- 每段都以总结句收尾；三段排比工整得像广告。
- 「不是 X，而是 Y」「不仅…而且…」连用。
- 破折号 —— 出现三次以上（这份文档自己就超标，所以别当范文）。
- 形容词堆叠：无缝、丝滑、强大、极致、优雅、赋能。
- emoji 当小标题；每条都以 ✅ / 🚀 / ✨ 开头。
- "在当今快速发展的 AI 时代"、"让我们一起"、"深入探讨"。
- 全是能力描述，没有具体数字、没有具体场景。
- 名词都加"能力"：xx 能力、闭环、抓手。
- 通篇没有一句"它现在做不到什么"。

---

## 8. 发完的 24 小时

- 每条评论都回；挑刺的先承认，再给改法或 issue 链接。
- 当天把 bug 变成 issue；下个版本 release notes 点名感谢。
- 不催、不转自己的帖、不在别的社区二次刷同一链接。
- 一周后回看：哪个渠道来的 issue 最多，下一轮就往那里加力气。

## 8.1 2026-11-18 的 Reddit 实测

- **r/coolgithubprojects 已发布并可见**（`/new/` 里能找到，无 removed 标记）：
  https://www.reddit.com/r/coolgithubprojects/comments/1xzj9zc/ai_terminal_pet_a_56px_alwaysontop_strip_that/
  标题：`AI Terminal Pet - a 56px always-on-top strip that shows which coding agent is waiting`
  链接直指 GitHub 仓库；版规只要求 GitHub 托管、标题带描述，语言 flair 自动分配。
- **r/SideProject 同文被 Reddit 站级 spam filter 移除**（帖子 JSON 的
  `removed_by_category: "reddit"`，`approved_at_utc: null`），不出现在 `/new/`；已给版主发 modmail
  请求从 mod queue 放行（发送后表单清空），等待回复。帖子本身仍可在作者 profile 看到：
  https://www.reddit.com/r/SideProject/comments/1xzj589/ai_terminal_pet_a_56px_alwaysontop_strip_that/
- 账号 `Visible_Swim471` 发布时只有 **6 karma**。同一天不要再去第三个版；先回评论、在相关帖子里
  正常参与，等这波过滤/放行结果出来再决定下一条。

## 9. 不要做的事

- 不要请人点赞/转发/顶帖（HN 明文禁止；V2EX 会判定推广）。
- 不要引用 npm 下载总量：2026-11-17 实测 4,586/周里有 25 个版本各 137–227 次（连不可安装的
  `0.0.0-stage` 都 166 次），是爬虫队形，不是用户——证据链在 [docs/METRICS.md](METRICS.md) §2。
  要引用就引用 **latest share** 并写窗口与日期。
- 不要用 AI 写 HN / V2EX 正文，也不要"润色"。
- 不要用公司/项目名当 HN 用户名（dang 的 tips：显得像品牌在推销）。
- 不要在 winget 合并前写 `winget install`。
- 不要写无法证明的使用量。
- 不要把直链 exe 发到帖子里，链接给 README。

---

## 10. 调研来源

- <https://news.ycombinator.com/showhn.html>（Show HN 规则：能跑、不要注册、别求赞、别删重发）
- <https://news.ycombinator.com/newsguidelines.html>（"don't put generated text in HN posts"）
- <https://news.ycombinator.com/item?id=22336638>（dang 的 present-your-work tips，含"别用 LLM 写/润色"）
- <https://www.v2ex.com/about>（「请不要把 AI 生成的内容发送到这里」）
- `awesome-claude-code` 的 CONTRIBUTING（人手网页表单、一次一个、14 天/100★、描述单行无 emoji）
- `awesome-opencode` 的 `contributing.md`（YAML + PR，一条一个文件；**门槛：≥25 stars、仓库满 30 天、90 天内有提交**；规则明写"任一不过就关 PR"）
