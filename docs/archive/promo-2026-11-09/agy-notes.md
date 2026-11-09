# agy 出稿记录（promo-writer 自留档）

本文件记录：喂给 agy 的输入、每一轮要它解决的问题、哪一版更好，以及我人工改过的地方。供 Lead 复核"文案到底是谁写的、事实从哪来"。

## 0. 工具与调用

```bash
/c/Users/rongj/AppData/Local/agy/bin/agy.exe -p "$(cat drafts/prompt-XX.md)" \
  --model gemini-3.8-flash-high --effort high \
  --dangerously-skip-permissions --output-format text
```

agy v1.5.3。单次 2–8 分钟。**输出只在进程退出时才落盘**，所以中间过程中 `wc -c` 一直是 0 并不代表卡死——这一条踩过坑（见 §5）。

## 1. 事实底座：drafts/facts.md

agy 没有仓库上下文，不能让它自由发挥。第一件事是把仓库事实读成一份清单（`drafts/facts.md`，134 行），所有 prompt 都以"facts.md 是唯一事实来源，不许编"开头。清单里的每条都带出处（文件:行号），例如：

- `internal/app/app_windows.go:23` → `const Version = "0.9.0"`
- `internal/session/session.go:17` → `MaxPayloadBytes = 1 << 20`
- `internal/pipeserver/server_windows.go:23` → 32 并发连接；`:68` → 管道 DACL
- `internal/store/store.go:22` → `DefaultAbsenceGrace` 2 分钟
- `internal/app/app_windows.go:830` → 10 分钟无会话自毁
- `reporters/protocol.js:364` → 11 条危险命令黑名单

## 2. 三轮 agy 各解决什么

### 第 1 轮：初稿（prompt-en.md / prompt-zh.md / prompt-social.md）
- 输入：facts.md + 逐节大纲（14 节，每节写什么、用哪张图、图表列名）+ 风格禁令（禁用词表、禁感叹号、emoji ≤ 6）+ 输出格式。
- 产出：`drafts/README.md`（219 行 / 1,954 词）、`drafts/README.zh.md`（213 行 / 17,414 B）、`drafts/_raw-social.md`（6,553 B）。
- 结果：骨架一次成型，可用。问题集中在"事实细节"而不是"文风"——说明长 prompt + 事实清单这条路是对的。

### 第 2 轮：按 Lead 的 8 条定点修订（prompt-fix-en.md / prompt-fix-zh.md）
Lead 逐条实测后给了 8 处必改（其中第 ⑦ 条后来被 Lead 自己撤回，见 §4）。prompt 的写法是"**这是修订不是重写**，只做这 8 处，没被点名的段落一个字不动"，并把每条改成"现状 → 改成什么 → 为什么"三段式，附上核实过的常量。
- EN 产出：`drafts/_raw-en2.md` 233 行，8 条全部落实。
- ZH 产出：`drafts/_raw-zh2.md` 233 行，7 条（含撤回的那条不用改）全部落实。
- 这一轮的价值：agy 在"局部手术"模式下比"再写一遍"稳得多——没出现大片措辞漂移。

### 第 3 轮：SOCIAL.md 修订（prompt-fix-social.md）
自查发现 5 处问题写进 prompt：`GDI+ renderer` 事实错误、"eleven minutes" 与中文稿"十四分钟"场景不一致、"half an hour" 是编造的时长、中文正文 437 字超了 300–400 字、X 各条的 (NNN chars) 标注要重算。
- agy 把其中 4 条一次改到位（GDI 措辞、fourteen minutes、for hours、X 字数重算），中文正文仍剩 433 字超限，我自己手改到 376 字。
- Lead 复核时又追加两件事：`https://github.com/<owner>/ai-terminal-pet` 占位符换成真实 owner `jjjjjjjjnnjnn`（5 处全换，换完 X 各条字数随之变动，重新跑脚本核了一遍：237/213/210/221/202）；以及危险命令条数确认（`reporters/protocol.js:364-376` 逐行数是 11 条正则，我原先写的 11 条是对的，Lead 一度说 10 条后自行更正）。
- 结果：`drafts/SOCIAL.md` 158 行 / 6,264 B。

### 耗时（墙钟，含我的人工复核）

| 轮次 | 启动 | 落盘 | 用时 | 备注 |
|---|---|---|---|---|
| 1 初稿 EN | 20:36 | 21:46 | ~70 min | 中途我误判卡死 kill 了一次，重跑；实际首轮稿子是两次尝试后落盘的 |
| 1 初稿 ZH | 20:39 | 21:45 | ~66 min | 同上，与 EN 并行跑 |
| 1 SOCIAL | 21:39 | 21:55 | ~16 min | 与两版 README 串行，机器上同时还有其他仓库的 agy 任务 |
| 2 修订 EN | 21:52 | 21:58 | ~6 min | 定点修订模式，prompt 短、任务窄，快很多 |
| 2 修订 ZH | 21:52 | 22:00 | ~8 min | 同上 |
| 3 SOCIAL 修订 | 21:56 | 22:05 | ~9 min | |

规律：**任务越窄越快**。"从零写一篇 200+ 行 README" 要 60 分钟量级，"只改 5 处" 10 分钟以内。所以第 2 轮起我把 prompt 从"给我重写"改成"只动这几行"，这既是质量策略也是时间策略。

## 3. 人工复核做了什么（不是 agy 的功劳）

- **图片**：把 5 张 PNG 全部读进来看像素，逐条比对 caption 写的东西是否真的在图里。`overlay-expanded-many.png` 的溢出文案在中文图里是"还有 3 个会话"（对应 i18n `KeyPanelMore` = "还有 %d 个会话"，`internal/i18n/i18n.go:345`），英文版是 "and %d more sessions"（`:234`）。
- **事实核查**：`grep -rn "gdiplus" --include="*.go" .` 零命中 → 确认是纯 GDI，`AGENTS.md:25` 那句 "GDI+ canvas rasterization" 是过时描述。
- **数字**：`drafts/lint.py` 扫全量数字 token，白名单外的每一个都人工回查。
- **计数**：X 各条 (NNN chars) 用脚本按可见字符重算，五条全部吻合。
- **零容忍项**：感叹号、emoji、禁用词、坏链、不存在的图片路径。

## 4. 被撤回的一条

Lead 最初要求把 Roadmap 的 "under 5 MiB combined" 改成单一目标 <2.5 MiB。复核 `docs/ROADMAP.md` 后发现原文 §L2 标题是"（3.7 MiB → 目标 <2.5 MiB）"，:174 验收标准写"两个 exe 合计 < 5 MiB"——**两个数都在原文里**，不是编的。Lead 随后撤回该条。最终三份稿子都保留了 "两 exe 合计 <5 MiB，主程序 <2.5 MiB" 的原表述。

## 5. 踩过的坑

- `ops_bash` 后台任务偶发 `EXECUTOR_UNAVAILABLE: Niubash execution tree link count mismatch`，**同一条命令重试一次就过**，不是命令本身的问题。
- agy 输出写到文件时，运行期间文件恒为 0 字节。曾据此误判"卡死"并 kill 掉三个进程；实际第一次的 README 稿子在 21:45/21:46 正常落盘。教训：判断 agy 是否活着看进程，不看输出文件大小。
- 中文稿曾有 GBK 输出问题：`python drafts/lint.py` 打印 `►` 会 `UnicodeEncodeError`。用 `PYTHONIOENCODING=utf-8` 或重定向到文件解决。

## 6. 哪一版更好

- **中文稿整体优于英文初稿**（Lead 也这么判断）：第 12–20 行的场景铺陈更具体，第 200 行路线图措辞更自然。英文稿在第 2 轮之后已经按中文稿的叙事节奏补过一轮。
- 结构上三份稿子共用同一套 14 节骨架，中英不是互译关系：中文按中文技术写作习惯重排了句式（短句、破折号少、数字与单位之间加空格），没有出现"通过……来……"式的翻译腔。

## 7. 自评（promo-writer 视角，打分留证据不留情面）

- **drafts/README.md — 8/10。** 事实层干净：所有数字有出处、5 张图全部真实存在且 caption 与像素对得上、没有禁用词与感叹号。扣分在两处：① 英文初稿的叙事节奏偏"功能清单"，第 12–20 行的痛点铺陈不如中文稿具体，靠第 2 轮按中文节奏回补才拉起来；② "Develop from source" 一节是 Lead 要求加的，位置在文末合理，但它与 Install 一节的边界仍有点模糊（读者可能不知道哪三条才是必需的）。
- **drafts/README.zh.md — 8.5/10。** 中文是重写不是翻译，短句、少破折号、数字与单位加空格，没有"通过……来……"式翻译腔（Lead 也认为中文优于英文初稿）。扣分：个别表格单元格为了对齐牺牲了口语感，读起来像文档不像人话。
- **drafts/SOCIAL.md — 8/10。** 三组文案都有具体钩子（"等十四分钟"这个数字比任何形容词都有效），HN 那条明确写了取舍与寻求反馈，不是单向宣传。扣分：X 五条里 Post 5 偏"免责声明"风格，转发欲一般；中文那条 376 字仍然略长于 V2EX 典型首楼。
- **共同扣分项（−1 级别，已记录未解决）**：三份稿子都没有真人跑过一遍"照做能不能装成功"。我核了每一条命令都来自仓库原文，但没有在干净机器上执行 —— 这是事实边界，不假装验过。
- **一句话结论**：agy 负责把事实组织成可读的结构，事实准确性完全靠 facts.md + 人工回查兜底；凡是没进 facts.md 的东西，这次三份稿子里都没有出现。
