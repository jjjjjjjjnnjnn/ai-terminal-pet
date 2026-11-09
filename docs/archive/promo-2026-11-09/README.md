# promo-2026-11-09 — 对外宣传文案的生产记录

本目录冻结 2026-11-09 那次「用 agy（gemini-3.8-flash-high）重写对外 README 与发布文案」的**全部中间产物**，
以便日后质疑「这段文案是 AI 写的还是编的」时能逐句回到事实来源。

约定同 [../README.md](../README.md)：归档文件不做内容修订，只在顶部加归档头。

## 最终落地在哪

| 产物 | 位置 |
|---|---|
| 英文正文 | [README.md](../../../README.md) `## English` 之下（v0.9.0 起替换了旧的 304 行英文版） |
| 中文正文 | [README.md](../../../README.md) `## 中文` 之下（重写而非直译） |
| 社交文案 | 本目录 [SOCIAL.md](SOCIAL.md)（HN/r/programming、V2EX/掘金、X 三组，尚未发布） |

合并进 README 的**定稿原文**另存两份，便于比对「agy 交稿」与「最终合并」的差异（实际只差标题降级、语言切换条与中文导语一行）：

- [README.draft-en.md](README.draft-en.md)（233 行 / 15,462 B）：英文定稿，`## The problem` 起到文末
- [README.draft-zh.md](README.draft-zh.md)（232 行 / 18,708 B）：中文定稿，`## 它解决什么问题` 起到文末

未归档的中间产物：`_raw-en.md` / `_raw-zh.md` / `_raw-social.md`（首轮 agy 输出，含 3 处后被发现的事实错误）、
`_raw-en2.md` / `_raw-zh2.md` / `_raw-social2.md`（第 2、3 轮输出）、`_lint-*.txt`（lint 日志）、`prompt-fix-zh.md` 以外的临时稿。
它们可由 [facts.md](facts.md) 与各轮 prompt 重新生成，且内容已被上方定稿与 [agy-notes.md](agy-notes.md) 完整记录，故不再入库。

## 事实底座 [`facts.md`](facts.md)

134 行，每条都带 `文件:行号` 出处。喂给 agy 之前先由人把仓库读成清单，
所有 prompt 都以「facts.md 是唯一事实来源，不许编」开头。**凡没进 facts.md 的东西，三份稿子里都没出现。**
几个关键锚点：`internal/app/app_windows.go:23` → `Version = "0.9.0"`；
`internal/session/session.go:17` → `MaxPayloadBytes = 1 << 20`；
`internal/pipeserver/server_windows.go:23` → 32 并发连接、`:68` → 管道 DACL；
`internal/store/store.go:22` → `DefaultAbsenceGrace` 2 分钟；
`internal/app/app_windows.go:830` → 10 分钟无会话自毁；
`reporters/protocol.js:364-376` → 11 条危险命令黑名单正则。

## 三轮 prompt 与产出

| 轮次 | prompt | 要 agy 做什么 | 产出 |
|---|---|---|---|
| 1 | [prompt-en.md](prompt-en.md) / [prompt-zh.md](prompt-zh.md) / [prompt-social.md](prompt-social.md) | 从 facts.md 出初稿：14 节大纲、每节用哪张图、禁用词表（revolutionary / blazingly / seamless / robust 等）、禁感叹号 | README 219 行、中文 213 行、SOCIAL 6,553 B |
| 2 | [prompt-fix-en.md](prompt-fix-en.md) / [prompt-fix-zh.md](prompt-fix-zh.md) | 定点修订 7 条（「这是修订不是重写，没被点名的段落一个字不动」，每条写成「现状 → 改成什么 → 为什么」） | 两版各 233 行，7 条全落实 |
| 3 | [prompt-fix-social.md](prompt-fix-social.md) | SOCIAL 局部手术 | 改到 4 处，中文 433 字仍超限，由人压到 376 字 |

耗时规律（见 [agy-notes.md](agy-notes.md) §6）：初稿约 70 / 66 / 16 分钟，定点修订只要 6 / 8 / 9 分钟——
任务越窄越快，agy 在「局部手术」模式下措辞漂移最少。

## 三轮之后人改了什么（不是 agy 的功劳）

- **GDI 不是 GDI+**：`grep -rn "gdiplus" --include="*.go" .` 零命中，画布是纯 GDI（`canvas_windows.go:61` `CreateCompatibleDC`、`:75` `CreateDIBSection`、`:80` `SelectObject`，提交走 `UpdateLayeredWindow`）。初稿写的 "GDI+ topmost canvas" / "GDI+ renderer" 全部改掉；`AGENTS.md` 与 `CHANGELOG` 里同样的错误描述也在同一轮一并修正。
- **危险命令是 11 条不是 10 条**：`reporters/protocol.js:364-376` 逐行数，第 365–375 行各一条正则。
- **时间数字统一**：全篇统一为 fourteen minutes / 十四分钟（初稿英文写 eleven）。
- **仓库链接**：占位符 `<owner>` 一律换成真实 owner `jjjjjjjjnnjnn`；换完 X 五条推文字数全变，已重算为 237/213/210/221/202。
- **图片 caption 逐张核过**：`overlay-expanded-many.png` 里的中文溢出行是「还有 3 个会话」（对应 `internal/i18n/i18n.go:345` `KeyPanelMore` = "还有 %d 个会话"，英文版 `:234`），不是想当然写的。
- **数字 token 全量回查**：[lint.py](lint.py) 扫每一种数字 token，白名单外的逐一对照 facts.md。它同时检查 emoji 数量、禁用词、感叹号、图片路径是否存在。

## 为什么最后没用「把草稿整篇搬进 README」的做法

旧 README 是 26691 B 的双语版（English 1–304 / 中文 312–588），但**通篇零图片引用**——
`docs/assets/` 下的 hero 与 7 张 overlay 截图生成于 `832cfd5` 后一次也没被用上，这是宣传缺口的核心。
新 README 把 5 张真实渲染截图插进正文，并把旧版里靠文字描述才能传达的交互（整块可点可拖、贴边收起、否决倒计时）改成「图 + 一句话 caption」。
