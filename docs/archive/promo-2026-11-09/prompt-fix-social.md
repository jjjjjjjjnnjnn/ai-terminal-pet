你在修订一份已经写好的发布文案草稿 drafts/_raw-social.md（请先完整读它，里面的 @@FILE 标记行要去掉、不要带到输出里）。整体可用，只改下面五处。其余文字、结构、小节顺序一个字不动。

【必改】
1）Room 1 正文第 3 段出现 "the GDI+ renderer"。事实是：这个画布是**纯 GDI**，全仓没有任何 gdiplus；只有 AGENTS.md 里一句过时描述写了 GDI+。把 "GDI+ renderer" 改成 "GDI renderer"。
2）全文里的 "eleven minutes" 改成 "fourteen minutes"，中文稿已有的"十四分钟"保持不变。原因：仓库自己的文案统一用十四分钟，两个房间讲同一个场景不要出现两个时长。（Room 1 标题、Room 1 正文第一段、Room 3 Post 1 都要改。）
3）Room 3 Post 3 结尾 "If an agent runs quietly for half an hour, its session does not vanish from the screen." 里的 "half an hour" 是一个凭空的具体时长。改成 "for hours"。
4）Room 2 正文现在 437 字（不含空白），超出 300–400 字的要求，请压缩到 380 字左右。压缩时优先删掉重复的铺垫，不要删掉这些信息：56 像素状态带、点一下跳到那个终端、5 秒否决窗口、rm -rf 与 git push --force 永不自动批准、存活靠操作系统观测而非心跳、仅 Windows、装完要接上报端、仓库地址、欢迎提 issue。改完请在正文末尾单独一行用 "(NNN 字)" 标注你实际算出的不含空白字符数。
5）Room 3 每条 Post 后面的 "(NNN chars)" 标注必须重算：改过节时长后长度变了，按可见字符（含空格、含链接占位符）重新数，标成真实数值。Post 5 的标注同样要重算。

【不许做】
- 不许出现感叹号；不许用 revolutionary / blazingly / supercharge / game-changer / seamless / effortless / cutting-edge / next-generation / unlock / powerful / delightful / magical / simply / "just works" / robust / comprehensive / elegant，也不许出现 神器 / 革命性 / 丝滑 / 极致 / 赋能 / 一键搞定 / 遥遥领先。
- 不许出现 emoji（现在 0 个，保持 0 个）。
- 不许编造任何数字、star 数、用户反馈、假引用。仓库地址占位符保持 https://github.com/<owner>/ai-terminal-pet。
- 不许声称支持 macOS / Linux，不许说它替代终端、是终端复用器、有云端功能。
- 不许新增小节，Hashtags / Posting notes / Numbers you may cite 三个小节保持原样（除非其中也写了 "eleven minutes"）。

【输出】只输出修订后的完整 Markdown，第一行就是 "# Launch Copy: ai-terminal-pet v0.9.0"。不要 @@FILE 标记，不要前言，不要解释，不要代码围栏。篇幅 150–220 行。