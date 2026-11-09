You are writing launch copy for a real Windows tool. Read drafts/facts.md first — it is the only source of truth. Do not invent anything.

The product: ai-terminal-pet, v0.9.0, MIT, Windows only, a ~4 MB Go overlay that shows the state of all your AI coding terminals (Claude Code, OpenCode, Antigravity, DeepSeek Harness) in one always-on-top strip on the edge of the screen.

TASK: write ONE complete Markdown file, drafts/SOCIAL.md. It contains three launch posts — not three variants of the same sentence, but three different registers for three different rooms. A reader of one should not feel they already read the other two.

ROOM 1 — Hacker News / r/programming, English, ~220 words.
- Register: sober, technical, slightly self-deprecating, zero marketing. HN readers punish hype instantly and reward "here is the specific problem and the specific fix".
- Lead with the concrete failure mode, not the product: you have four terminal tabs, one of them has been sitting on a permission prompt for eleven minutes while you debugged something else. Then say what you built, in one sentence, and what it is made of (Go, Win32 layered window, per-user named pipe, JSON lines, no electron, no account).
- Include exactly two numbers that a skeptic would want: the binary size (4,178,432 B for the main exe) and the fact that 140 tests exist. Then the one genuinely interesting engineering decision: session existence is decided by OS observation, not by reporter heartbeats, so a quiet session does not vanish.
- Then the honest cost: Windows only, antivirus makes process spawning expensive, no pixel regression tests, the auto-approve timeout needs a restart. Do not hide them; the caveats are the credibility.
- End with an ask that invites critique rather than stars ("I would like to know where this breaks for you" style, but write your own sentence).
- First line must be a hook under 90 characters that works as a submission title. Suggested shape: the failure mode, not the product name.

ROOM 2 — 中文, V2EX / 掘金风, 300–400 字.
- 语气：平实、具体、带一点自嘲和踩坑复盘的味道，像一个开发者在 V2EX 分享自己做的东西，不是厂商发稿。
- 开头别喊"开源了一个神器"，直接抛场景：同时开着 Claude Code、OpenCode、Antigravity，某个终端早就在等你点确认，你却在另一个窗口里改 bug，十四分钟后才发现。
- 中段讲它做了什么（贴边一条状态带 / 点一下就能跳到那个终端 / 自动同意有 5 秒否决窗口 / 危险命令永远不自动批准），再讲一个技术判断：会话在不在，靠操作系统观测，不靠心跳，所以静默的会话不会莫名消失。
- 结尾给读者一个明确的下一步（仓库地址占位、装完要接上报端、欢迎提 issue 说哪里不对），并且点明限制：只有 Windows。
- 中文要短句，不要翻译腔，不要"赋能/丝滑/革命性/神器/一键"。中英文之间加空格。不要感叹号。
- 首句钩子必须能在信息流里一眼看懂，不超过 40 字。

ROOM 3 — X / Twitter, English, 3–5 posts in a short thread, each under 280 characters.
- Register: terse, dry, one idea per post. No thread-boy cadence ("🧵 1/9"), no emoji spam (max 2 in the whole thread), no "just shipped".
- Post 1: the failure mode in one line plus what it is.
- Post 2: what the compact strip actually shows.
- Post 3: the engineering decision (existence from OS observation, not heartbeats).
- Post 4: the safety angle (auto-approve off by default, 5 s veto, dangerous patterns never auto-approved).
- Post 5: the honest limit (Windows only) plus a link placeholder.
- Mark each post with a character count you actually computed for that post, in the form "(NNN chars)". Be accurate: count the visible characters including spaces and the link placeholder.

AFTER THE THREE POSTS
- ## Hashtags — one block: which hashtags to attach to which room, 2–4 per post, no more. Explain in one line why fewer is better here.
- ## Posting notes — 4–6 bullets: what to post as the first comment (e.g. the honest limitation, pre-empting the obvious objection), what not to do (do not lead with the ASCII diagram, do not paste the whole README), the best time/ordering to post, and which room to post in first and why.
- ## Numbers you may cite — a short list, at most 6 bullets, of exactly the verifiable claims used above, each with a one-line note on where it comes from (fact sheet). This is for fact-checking before publishing.

STYLE RULES (violating any makes the draft unusable)
- Banned: revolutionary, blazingly fast, supercharge, game-changer, 10x, seamless, effortless, cutting-edge, next-generation, unlock, powerful, delightful, magical, simply, "just works", 神器, 革命性, 丝滑, 极致, 赋能, 一键搞定, 遥遥领先.
- No exclamation marks anywhere, in any language.
- Never invent a number, a user reaction, a star count, a benchmark, or a quote from a real person.
- Do not claim macOS or Linux support. Do not claim it replaces the terminal, works as a multiplexer, or does anything in the cloud.
- Repository URL placeholder: use https://github.com/<owner>/ai-terminal-pet consistently. Do not invent an owner name.

OUTPUT FORMAT — parsed by a script, follow literally:
First line must be exactly "@@FILE drafts/SOCIAL.md@@". Then the whole document. Nothing after the last line. No code fence around the document, no preamble, no "here is what I wrote".
Target 150–220 lines.
