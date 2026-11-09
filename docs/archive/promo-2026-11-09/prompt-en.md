You are writing the front page of a GitHub repository. The repository is real, at C:\\Users\\rongj\\Desktop\\项目开发\\ai-terminal-pet. Read the FACT SHEET at drafts/facts.md first — it is the only source of truth. You may also read README.md, AGENTS.md, docs/ROADMAP.md and reporters/protocol.js to check anything, but the fact sheet already covers what you need.

TASK: write ONE complete Markdown file, drafts/README.md, in ENGLISH, for a developer who lands on this repo for the first time and has never heard of it.

AUDIENCE AND THE 3-SECOND TEST
- The reader runs Claude Code, OpenCode, or Antigravity in two or three terminals at once, on Windows. They lose track of which terminal is waiting for permission. Their first thought when they land is "what does this do for me, and is it junk?".
- The first screen (title, one line, hero image, badges) must answer: what it is, what it replaces (glancing across terminal windows), and what it costs (a ~4 MB exe, no account, no cloud).
- Answer "what problem does this solve" before "how does it work". No lore, no backstory, no philosophy paragraph.

STRUCTURE — follow this outline exactly, section by section:
1. # ai-terminal-pet
2. One sentence under the title (max 25 words), then the hero image: ![ai-terminal-pet](docs/assets/hero.png)
3. Badge row (shields.io static badges only, no dynamic workflow badges): platform Windows, Go 1.23, version 0.9.0, license MIT, dependencies 3, tests 140. Two lines maximum.
4. ## The problem — three short paragraphs or a compact before/after table. Before: four terminal windows, you find out a build was blocked on a permission prompt fourteen minutes ago. After: one 56px strip on the edge of the screen tells you which project is waiting and for how long.
5. ## What it looks like — the screenshots. Each image on its own line with an italic caption underneath. Use exactly these four, in this order, and keep their filenames:
   docs/assets/overlay-compact.png — compact strip
   docs/assets/overlay-expanded.png — expanded list
   docs/assets/overlay-expanded-many.png — many sessions
   docs/assets/overlay-auto-approve.png — the 5-second veto countdown
   Captions must describe what is actually visible in a real Win32 render: state orb, session count, state word, project name, per-row project/state/source/relative time, the overflow line, the "Allow (5s)" bar. Do not caption a feature that is not in the picture.
6. ## Features — a table with 8–12 rows. Columns: Feature | What it actually does. Keep each cell a concrete behaviour with a number where the fact sheet has one (65–160% zoom, 56x156 compact, 6 rows then overflow, 2-minute absence grace, 32 connections). Include: state aggregation over 7 states; compact and expanded panels; whole-panel gestures; edge retract with alert override; tray menu; three-tab settings window; synthesized alert tones in four schemes; per-session alerts (waiting/error/finish); auto-approve with a veto window and a dangerous-command denylist; jump to the waiting terminal window; --json for scripts; single instance; self-removal when no session is left.
7. ## How it works — ONE fenced ASCII diagram, max 12 lines, showing host CLI -> reporters -> named pipe -> Go overlay -> Win32 layered window, then 4–6 bullets: per-user pipe name (do not print a real hash, write <8hex>), JSON line protocol with exact OK/ERR, 1 MiB cap, snapshot files as the on-disk record, existence decided by the OS not by heartbeats, disks beat heartbeats. Mention that the same pipe-name algorithm is implemented in Go and JS and locked by a cross-language test.
8. ## Install — 30 seconds, Windows only, PowerShell. Numbered steps, max 6 commands in total. Use the real build commands from the fact sheet, then .\\install.ps1, then one line to prove it works (ai-terminal-pet --json). Add a one-line note that installing a terminal hook is required for that terminal to report, and point at the next section.
9. ## Connect your terminal — a table: Terminal | How to connect | What you get. Rows: OpenCode, Claude Code, Antigravity, DeepSeek Harness (the Cordis bundle via the app's plugin manager), Anything else (reporters/generic.js one-liner). Keep each "how" to one or two short lines with the real command shape. Then a fenced block with 3 generic.js examples exactly in the shape the README shows.
10. ## Control it from a script — the CLI table. Columns: Command | Effect. Every command from the fact sheet. Then a 6-line JSON example of --json output (invent plausible project names, keep the real field names: schemaVersion, state, count, generatedAt, sessions). One sentence on why --json never blocks: it reads snapshots on disk.
11. ## Safety — what the hook can and cannot do. Cover: reporters fail fast and never throw into the host CLI; auto-approve is off by default; the veto window; the dangerous-command denylist (list 4–5 of the real patterns, not all 11); no telemetry, no network, all state under %LOCALAPPDATA%; the pipe is per-user and ACL-restricted.
12. ## Known limits — a bulleted list of 5–7, honest, from the fact sheet. Do not soften them. Include: Windows only; AV makes process creation expensive; pipe squatting mitigated not eliminated; auto-approve timeout needs a restart; no pixel regression tests.
13. ## Roadmap — 4–6 bullets from the real roadmap, phrased as work not promises ("planned", "not started"). Include the two non-goals that matter to a reader: no cloud, no telemetry, and it is not a terminal replacement.
14. ## License — one short paragraph, MIT, name the three third-party modules.

STYLE RULES (violating any of these makes the draft unusable)
- Banned words: revolutionary, blazingly fast, blazing, supercharge, game-changer, 10x, seamless, effortless, cutting-edge, next-generation, unlock, powerful, delightful, magical, simply, "just works", robust, comprehensive, elegant.
- No exclamation marks anywhere. At most 6 emoji in the whole document, and none in headings.
- Second person, present tense, short sentences (mostly under 20 words). Prefer the concrete over the abstract: "a tap denies it, five seconds of silence approves it" not "a flexible approval workflow".
- Every number must appear in the fact sheet. If you are not sure, leave the number out.
- Do not invent features, platforms, benchmarks, stars, users, testimonials, or a comparison with products that are not in the fact sheet.
- Do not call it a terminal multiplexer, a log viewer, a cloud service, cross-platform, or an "AI agent".
- Keep headings at ## level, no deeper than ###. No horizontal rules except one between the screenshots section and the features section.
- Do not add a Chinese section. This file is English only; a separate Chinese file is written by someone else.

OUTPUT FORMAT — follow this literally, it is parsed by a script:
Write nothing before the first line "@@FILE drafts/README.md@@". Then the whole Markdown document. Then nothing after the last line of that document. Do not wrap the document in a code fence. Do not add commentary, preamble, or a summary of what you did.

If you need to check a detail, read the files first, then emit the marker and the document in one go. Target length: 550–750 lines is too long; aim for 220–320 lines of Markdown.
