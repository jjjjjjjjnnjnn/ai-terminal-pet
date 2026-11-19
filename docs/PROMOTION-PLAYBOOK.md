# Promotion Playbook（可复用流程）

> **2026-11-19 起**：可执行流程以本地知识库 `Promotion/PLAYBOOK.md` 为准（`C:\Users\rongj\Desktop\本地知识库\知识库内容\Promotion\`）。
> 本文件保留为方法参考与历史；渠道状态、排期与发布台账已迁到**宣传总线**，这里不再新增。

> 适用：本地优先的开发者工具、桌面工具、CLI、插件。本文只写**流程、判据和验证方法**，不写某个项目的数字。
> 换项目时整份复制，只替换「事实卡」和链接；数字必须从该项目的代码/构建现场取，不能沿用本文。
>
> 本项目的具体文案与发布记录在 [PROMOTION.md](PROMOTION.md)；渠道内容与截图清单在 [LAUNCH.md](LAUNCH.md)。

---

## 0. 开始前的四个前置条件

1. **事实卡是现取的。** 版本、二进制体积、测试数、平台限制，用项目自己的命令打印。本文不提供数字。
2. **每个渠道先读当期规则。** 规则会变；`about/rules`、`CONTRIBUTING.md`、置顶帖三者都要看。
3. **登录/2FA/验证码由本人完成。** 自动化可以填表、上传、点发布，但不能替人过验证码，也不该替人写平台明令禁止的 AI 文本。
4. **网络先验证。** 能打开不等于能发：有些网络会按域名阻断 TLS，有些 VPN 只在浏览器或部分应用生效。见 §7.1。

---

## 1. 渠道矩阵（硬规则）

| 渠道 | 能做什么 | 不能做什么 | 验证方式 |
|---|---|---|---|
| GitHub 社交预览 | 上传 1280×640 PNG | 没有 API，只能走 Settings UI | DOM 里的 `repository-images.githubusercontent.com/...`，再下载比对尺寸 |
| X / Twitter | 单帖或线程；GIF 或最多 4 图 | 同一条不能又 GIF 又图片 | 登录页 + `publish.twitter.com/oembed`；未登录访客可能有登录门槛 |
| Reddit | 按版规发链接/文本；回复评论 | 同一天同一文案连发多版；绕过站级过滤 | 帖子 JSON 的 `removed_by_category`；`/new/` 是否出现 |
| awesome 列表 | 按 CONTRIBUTING 用网页表单 | 多数要求**人类提交**；有的只允许 collaborator 开 issue | 表单页是否出现 `Action restricted` |
| Hacker News | 本人写、本人发 | **明令禁止 AI 生成正文**（含润色） | 不代写；只提供事实卡 |
| V2EX | 本人写、本人发 | **明令禁止 AI 生成内容** | 不代写；只提供事实卡 |
| npm / 市场目录 | 包版本、README、目录条目 | 把下载总量当用户数 | `npm view`、对方站点的公开索引 |

---

## 2. 去 AI 化的写作规则

**结构（中文 150–250 字 / 英文 120–200 词）**

1. 一个具体场景：哪个工具、哪次等待、等了多久。
2. 它具体是什么：尺寸、交互、数据从哪来。
3. 一条可复现的数字：体积、测试数、二进制数。
4. 一条它现在做不到的事：未签名、未上真机、只支持某平台。
5. 一个具体问题：请读者判断这个交互是否有用，不要求 star/upvote。

**删掉这些**

- 破折号 `—` 连用；三段排比；每段都用总结句收尾。
- `强大 / 无缝 / 优雅 / 极致 / 赋能 / 一站式`。
- `不是 X，而是 Y`、`不仅…而且…`、`在当今 AI 时代`、`让我们一起`。
- emoji 当小标题；`✅ 🚀 ✨` 列表。
- 只列功能、没有场景；没有一句限制；没有数字。
- 把 npm 下载总量、目录排名、star 数当成绩。

**保留这些**

- 第一人称、短句、具体名词。
- 承认边界：未签名会弹 SmartScreen；macOS 只在 CI 构建；Linux 需要 X11。
- 一条链接，放在正文靠后。

---

## 3. 一次宣传的固定顺序

```
0. 事实卡（命令现取）
1. 素材：社交预览 1280×640、演示 GIF（≤5 MB）、2–3 张截图
2. GitHub 社交预览
3. X：先根帖，再一次一条回复
4. Reddit：读规则 → 发一个版 → 查 removed_by_category → 必要时 modmail
5. awesome 列表：先查入口是否对人开放
6. HN / V2EX：只交事实卡，本人写
7. 24 小时 / 1 周后测量
```

同一主题一周内不要重复发；同一文案不要在同一天进两个社区。

---

## 4. 每个渠道的步骤与验证

### 4.1 GitHub 社交预览

1. 打开 `https://github.com/<owner>/<repo>/settings`。
2. Social preview → Edit → Upload an image…。
3. 文件输入是隐藏的：直接对 `#repo-image-file-input` 上传，不要点菜单项去开系统文件框。
4. 上传后页面可能仍显示旧态；重新加载，检查 `.repository-og-image` 的 `background-image`。
5. 下载该 URL，确认尺寸/字节与本地一致。

判据：DOM 里出现 `repository-images.githubusercontent.com/<repo-id>/<uuid>`，并且取回成功。

### 4.2 X

1. 先验证登录：`https://x.com/compose/post` 能打开且右侧有账号。
2. **多帖编辑器有一个坑**：后加的附件会被挂到第一条。可靠做法是：
   - 先发根帖（只带 GIF，或只带图片）；
   - 然后在已发布的帖子上逐条 Reply，每条单独上传媒体。
3. 每条 ≤270 字符（X 上限 280；URL 按 23 字符计）。
4. 发布后检查登录页；再用公开 oEmbed 验证：
   `https://publish.twitter.com/oembed?url=<tweet-url>`。
5. 未登录访客看到 `This Post is unavailable` 不一定是被删；X 对未登录访问有门槛。不要因此重发；等 24–48 小时，或让本人登录查看。

判据：根帖 URL 能打开，回复在同一会话里；oEmbed 返回该文本。

### 4.3 Reddit

**发之前**

1. 读 `https://old.reddit.com/r/<sub>/about/rules` 和右侧栏；看置顶帖是不是当期主题。
2. 看标题格式要求。例如有的版要求 `[Project name] - [Short description]`。
3. 看账号状态：karma、注册时间。低 karma 账号的推广帖很容易进站级过滤器。
4. 选择帖子类型：
   - 版块就是“分享 GitHub 项目” → 链接帖；
   - 版块要故事/反馈 → 文本帖，链接放正文。

**发之后（必须验证）**

```
GET https://www.reddit.com/r/<sub>/comments/<id>.json?raw_json=1
```

看第一个 child 的：

- `removed_by_category: null` 且 `approved_at_utc` 有值 → 正常；
- `removed_by_category: "reddit"` → **站级 spam filter**，不是版规；去 modmail；
- `removed_by_category: "moderator"` → 版主移除，读 `mod_reason_title`。

同时确认帖子出现在 `/r/<sub>/new/`。

**被站级过滤时**

- 不要重发，不要换标题再发。
- 给版主发 modmail：`https://www.reddit.com/message/compose/?to=/r/<sub>`。
- 内容只写：帖子 URL、`removed_by_category: "reddit"`、请他们从 mod queue 看一下、愿意按要求改。
- 同一天不要再发第三个版；先去回评论、攒一点 karma。

**节奏**

- 一个版一天最多一条；同一天最多两个版，且文案不能逐字相同。
- 发完 24 小时内回每条评论；被指出问题时先承认，再给 issue/改法。

### 4.4 awesome 列表

1. 读该仓库的 `CONTRIBUTING.md`，确认入口是 **issue 表单 / PR / 网页表单** 中的哪一种。
2. 如果页面出现 `Action restricted — an owner has limited the ability to create an issue to users that are collaborators`，说明普通账号没有入口；记录下来，不要绕过。
3. 如果规则写明“recommendations must be created by human beings”，只准备一行描述，由本人提交。
4. 一行描述：单行、无 emoji、不推销、说清它是什么。

### 4.5 Hacker News / V2EX

- 两个平台都明文禁止 AI 生成正文。**不代写、不发。**
- 提供：事实卡、结构提纲、会被问到的问题与答案、诚实边界。
- 正文由本人写；发完本人回评论。

---

## 5. 测量与复盘

24 小时看：

- 渠道来的访问（项目自带的流量脚本 / referrer）；
- star、issue、PR；
- 评论里问得最多的一个问题。

1 周看：

- npm 的 `latest share`，不是下载总量；
- 哪个渠道带来第一个外部 issue；
- 下一轮把力气加到那个渠道，停掉没有回声的渠道。

---

## 6. 换一个项目怎么复用

复制本文件，按顺序替换：

1. **事实卡**：项目名、一句话、版本、体积、平台、限制、仓库链接。
2. **素材**：社交预览、GIF、截图，重新生成，不沿用旧图。
3. **文案**：按 §2 重写；不要复制上一项目的句子和数字。
4. **渠道**：先确认目标社区是否接受这个品类；AI 工具进“Not-AI”主题的版要谨慎。
5. **记录**：把实际发布 URL、过滤结果、modmail、评论结论写进该项目的 `PROMOTION.md`。

**不要复制**：版本号、二进制体积、测试数、下载量、star 数、任何“已验证”的措辞。

---

## 7. 本次实战教训（可迁移）

### 7.1 浏览器“什么都打不开”先查 Secure DNS

Chrome 的 `Local State` 里 `dns_over_https.mode = "secure"` 且模板指向一个当前网络不可达的 DoH 时，**所有站点**都是 `ERR_NAME_NOT_RESOLVED`，而系统 DNS 正常。

- 路径：`%LOCALAPPDATA%\Google\Chrome\User Data\Local State`
- 先备份，再把 `dns_over_https.mode` 改成 `"off"`，重启 Chrome。
- 不要先怀疑网站或账号。

### 7.2 网络“通了”不等于目标站点通了

系统能开 Google/GitHub，不代表 Reddit 能开。按域名阻断会在 TLS 握手阶段重置连接：

```
curl -v https://<site>   # schannel: failed to receive handshake
```

如果本机没有代理、VPN 会员过期、Tailscale 没有 exit node，就不要承诺“马上发”；先让本人开全局代理/VPN，或提供代理地址。

### 7.3 X 的两件事

- 多帖编辑器把后加的媒体挂到首条；GIF 与图片不能同条。分步发回复最稳。
- 未登录访问可能显示 `This Post is unavailable`；登录后能看就不是被删，不要重发。

### 7.4 Reddit 的两层过滤

- 版规（人）与站级 spam filter（机器）是两件事。
- `removed_by_category: "reddit"` = 机器；modmail 是唯一正规恢复路径。
- 新账号 6 karma 发外链，进过滤器的概率很高。先评论、后发帖。

### 7.5 入口可能对普通账号关闭

awesome 列表可能不只是“必须人类提交”，而是**只允许 collaborator 开 issue**。看到 `Action restricted` 就停止，记录状态，不要用 PR/API 绕过。

---

## 8. 交接清单（给下一个人/下一个项目）

- [ ] 事实卡已现取，数字与代码一致。
- [ ] 每个渠道的当期规则已读，入口可用。
- [ ] 浏览器登录状态、网络、代理已验证。
- [ ] 素材已生成并与文档尺寸一致。
- [ ] 文案按 §2 自检：无破折号堆叠、无营销词、有限制、有一个具体问题。
- [ ] 发布后验证：社交预览 URL、X oEmbed、Reddit `removed_by_category`。
- [ ] 记录写进项目文档：URL、时间、过滤/放行、评论结论。
- [ ] 下一轮只加一个渠道，并设 24 小时/1 周复盘。
