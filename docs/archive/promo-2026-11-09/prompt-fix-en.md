You are revising an existing draft, NOT writing a new one. The draft is at drafts/README.md — read it first in full. It is already good. Your job is exactly EIGHT surgical fixes. Do not restructure, do not rewrite sections that are not named below, do not change wording that is already correct. The revision must stay recognisably the same document.

FACT CORRECTIONS (verified by the repo owner — these are authoritative, trust them over anything else):
- The canvas is plain GDI, NOT GDI+. internal/ui/canvas_windows.go:13 imports only syscall + github.com/lxn/win. :61 CreateCompatibleDC, :75 CreateDIBSection (32bpp, BI_RGB, negative biHeight = top-down DIB), :80 SelectObject, then UpdateLayeredWindow for the layered window. There is no gdiplus anywhere in the repo. AGENTS.md:25 still says "GDI+ canvas rasterization" — that line is stale and wrong, do not copy it.
- The named-pipe DACL is literally "D:P(A;;GRGW;;;OW)(A;;GRGW;;;SY)" (internal/pipeserver/server_windows.go:68). That grants read+write to Owner Rights (OW, i.e. the Windows user running the app) and to LocalSystem (SY) only. Other local users cannot connect to the pipe and cannot squat it.
- internal/session/session.go:17 MaxPayloadBytes = 1 << 20 (1 MiB). SchemaVersion "ai-terminal-pet.v1" at :16. Seven states idle, thinking, working, tool_use, waiting_permission, error, closed at :21-29.
- internal/pipeserver/server_windows.go:23 — 32 concurrent connections. cmd/pet-statusline/main.go:375 — the ACK is the exact string "OK".
- store.DefaultAbsenceGrace = 2 minutes (internal/store/store.go:22). No-session self-destruct at internal/app/app_windows.go:830: time.Since(a.lastAvailable) >= 10*time.Minute.

THE EIGHT FIXES

FIX 1 — The positioning sentence on line 3 is one 28-word monster that also stuffs in "3.98 MiB executable". Split it into TWO sentences, each under 15 words. Sentence one: the pain it removes (you stop glancing across terminal windows). Sentence two: what it costs you (one small executable, no account, no cloud). Move the exact size out of the sentence — it belongs in the badges row or the body, not in the first line a reader sees.

FIX 2 — In the Features table, the last row "Single instance and lifecycle" crams four unrelated things into one cell. Delete that row and add THREE rows in its place, each with one job:
  - "Single instance" — a named mutex makes a second launch activate the running overlay instead of starting another one.
  - "Connection capacity" — the pipe server accepts 32 concurrent connections with a 5 s read timeout, so IPC never blocks the UI thread.
  - "Idle lifecycle" — a session is dropped only after a 2-minute absence grace; with no session left, the process removes itself after 10 minutes.
Keep the table's existing column shape and row wording style.

FIX 3 — In the ASCII architecture diagram the UI line says "GDI+ topmost canvas". Change it to "GDI topmost canvas". Then, in the "How it works" bullets right under the diagram, add one bullet on how it is actually drawn: a top-down 32bpp DIB section created with CreateDIBSection, painted through a memory DC, and handed to the layered window with UpdateLayeredWindow. One sentence is enough; state the sequence, do not advertise it.

FIX 4 — In the Install section, delete the vet/test step entirely. A user installing the product is not a contributor. The install stays three steps and nothing more: build both binaries, run .\\install.ps1, prove it with ai-terminal-pet --json. Then add a NEW section at the very end of the document, after License, called "## Develop from source", and put the verification commands there: go vet ./..., go test -count=1 ./..., node --check reporters/*.js, node integrations/deepseek-harness/verify.mjs, plus the PET_PREVIEW_DIR preview command. That section is for people changing code, so one line of framing is fine.

FIX 5 — In the "Connect your terminal" table, the DeepSeek Harness row's "What you get" cell says "Event subscriptions, agent status updates, and approval request forwarding". That is vague and partly invented. Replace it with the real event mapping, in this shape: turn/start to working, tool/call to tool_use, approval/asked to waiting_permission, tool/result with an error to error, turn/end to idle, process exit to closed. Also add a sentence under the table saying the Harness side is a host-only Cordis bundle: it holds no pipe-name, state-set or snapshot logic of its own, it loads the protocol from %LOCALAPPDATA%\\AI-Terminal-Pet\\protocol.js and delegates everything to it.

FIX 6 — In Safety, the "Per-user ACL protection" bullet says only "tightened Windows security descriptors". Make it concrete: quote the DACL string "D:P(A;;GRGW;;;OW)(A;;GRGW;;;SY)" and say plainly what it grants — read/write for the owner of the pipe and for LocalSystem, and nothing for any other account on the machine. Say that other local users therefore cannot connect and cannot occupy the pipe name.

FIX 7 — In Roadmap, the binary-size bullet claims "under 5 MiB combined, targeting under 2.5 MiB for the main binary". The "5 MiB combined" figure is not in the roadmap and must go. docs/ROADMAP.md section L2 states 3.7 MiB now, target under 2.5 MiB for the main binary. Keep exactly one target: main binary under 2.5 MiB. Describe the means briefly (strip reflection metadata, drop unused packages, replace the regexp-based redaction in pet-statusline with a hand-written scanner).

FIX 8 — The document ends with two separate bullets both starting "**Non-goals**". That is a duplicated heading in bullet clothing. Merge them into one subsection: a short "## Non-goals" heading (or a single "**Non-goals**" lead-in followed by two sub-bullets) with exactly two items:
  - no accounts, no cloud sync, no telemetry; everything stays in %LOCALAPPDATA%;
  - not a terminal multiplexer and not a replacement for your terminal — it only tells you which session wants attention.
Do not keep any other bullet under this heading.

HARD RULES FOR THIS REVISION
- Change nothing else. Same sections, same order, same image filenames, same table columns.
- Keep the banned-word list enforced: revolutionary, blazingly, blazing, supercharge, game-changer, 10x, seamless, effortless, cutting-edge, next-generation, unlock, powerful, delightful, magical, simply, "just works", robust, comprehensive, elegant.
- No exclamation marks. No emoji in headings.
- Every number you write must be one of the verified numbers above or already present in the draft. Invent nothing.
- Second person, present tense, short sentences. Do not add adjectives that add no information ("lightweight", "fast", "smart").
- Do not add a Chinese section.
- Target length 200-260 lines. If a section not named in the eight fixes would get shorter or longer, you have made a mistake.

OUTPUT FORMAT: print the complete revised Markdown document to stdout and nothing else. No preamble, no "Here is the revised version", no explanation, no code fence around the document. Start with the line "# ai-terminal-pet".