# Launch Copy: ai-terminal-pet v0.9.0

## Room 1 — Hacker News / r/programming

### Title

One terminal tab sat on a permission prompt for fourteen minutes while you debugged

### Submission Body

You have four terminal tabs open. One of them has been sitting on a
permission prompt for fourteen minutes while you debugged something else in
another window.

To fix this for my own workflow, I built ai-terminal-pet: an edge overlay
written in Go with Win32 layered windows that aggregates terminal state
over a per-user named pipe with JSON lines, without Electron, frameworks,
or any user account.

The main executable compiles to 4,178,432 B, and the repository has 140 tests
covering geometry, hit testing, and protocol framing. The most relevant
engineering decision is that session existence is determined by OS
observation rather than reporter heartbeats. By periodically scanning the
process tree and matching window titles, the overlay avoids dropping quiet
sessions when an agent stops broadcasting.

The trade-offs are real. It runs only on Windows. Antivirus scanning makes
process spawning expensive when hooks run on every event. There are no
automated pixel regression tests for the GDI renderer, though geometry and
text layout are unit-tested. Also, unlike other settings that hot-reload
immediately, changing the auto-approve countdown length requires an
overlay restart.

I would like to know where the window matching breaks down on your terminal
setup or what crashes on unusual display scale factors.

## Room 2 — V2EX / 掘金

### 帖子标题

同时开几个 AI 终端，某个窗口卡在授权确认等了十四分钟，你切过去才看见

### 正文内容

平时开着 Claude Code、OpenCode、Antigravity，某个终端早就在等确认，你却在另一个窗口改 bug，十四分钟后才发现。

我做了个叫 ai-terminal-pet 的屏幕边缘 56 像素状态带：哪个终端在跑、在报错、在等权限，一眼看到；点一下跳到那个终端。

嫌确认烦可以开启自动同意，每次权限请求留 5 秒否决窗口，点一下拒绝，5 秒没操作才放行；rm -rf、rd /s、mkfs、dd of=/dev/、fork bomb、shutdown /s、git push --force 等 11 条规则永不自动批准。

会话存活靠操作系统观测而非心跳，静默的会话不会莫名消失。

限制：仅 Windows；装完要接上报端才能抓到状态。

仓库在 https://github.com/jjjjjjjjnnjnn/ai-terminal-pet ，标题匹配不上或用着别扭，欢迎提 issue。

(376 字)

## Room 3 — X / Twitter

### Post 1

You have four terminal tabs open. One sat on a permission prompt for fourteen minutes while you debugged in another window.

ai-terminal-pet is a 4 MB Go overlay that keeps the state of your AI coding sessions visible on the screen edge.

(237 chars)

### Post 2

The 56 px compact strip shows the highest-priority state across Claude Code, OpenCode, and Antigravity. It pulses when a prompt waits. Clicking the jump strip brings the matching terminal window to the foreground.

(213 chars)

### Post 3

Session existence is decided by OS observation, not reporter heartbeats. The overlay inspects the process tree and window titles. If an agent runs quietly for hours, its session does not vanish from the screen.

(210 chars)

### Post 4

Auto-approve is off by default. When enabled, permission prompts get a 5 s veto countdown with a draining bar. Destructive patterns like rm -rf and git push --force are never auto-approved and require manual confirmation.

(221 chars)

### Post 5

The honest cost: Windows only (Win32 layered window and per-user named pipe). Process spawning can feel slow under aggressive antivirus. Zero telemetry.

https://github.com/jjjjjjjjnnjnn/ai-terminal-pet

(202 chars)

## Hashtags

- Hacker News / r/programming:
  No hashtags on Hacker News; use the `Project` or `Tool` post flair on r/programming.

- V2EX / 掘金:
  `#Windows` `#Go` `#开源`

- X / Twitter:
  `#golang` `#devtools` `#windows`

Keeping to two or three targeted tags prevents algorithmic demotion as spam and keeps the focus directly on developers dealing with multi-terminal state.

## Posting notes

- Post the honest limitations immediately as the first comment on technical forums:
  State clearly that the project is Windows-only, explain why antivirus software makes
  process spawning expensive during hook execution, and note that auto-approve timeout
  adjustments require an application restart.

- Do not lead with ASCII architecture diagrams, marketing slogans, or verbatim dumps
  of the README file:
  Technical communities reward leading with the specific friction point and the precise
  solution rather than unprompted architectural diagrams.

- Launch order:
  Submit to Hacker News first, where technical interrogation is most rigorous; follow
  with V2EX to gather localized Windows desktop feedback, and publish the X thread last
  with links back to the active discussions.

- Timing:
  Post to Hacker News during morning engineering hours (around 8 AM US Eastern on a
  Tuesday or Wednesday) to reach peak developer traffic; post to V2EX around mid-morning
  during developer breaks.

- Ground all comment responses in verified facts:
  If commenters report issues with specific terminal emulators or multi-monitor DPI scaling,
  collect their terminal name and Windows build without defensiveness.

## Numbers you may cite

- `4,178,432 B` (~3.98 MiB):
  The exact size of the main executable `bin/AI-Terminal-Pet.exe` built with `-trimpath -ldflags "-s -w"`
  (source: facts.md § Measured numbers).

- `140`:
  Number of Go test functions covering layout, geometry, hit testing, and protocol framing
  (source: facts.md § Measured numbers).

- `56 px`:
  Logical pixel width of the compact overlay panel docked against the screen edge
  (source: facts.md § The overlay).

- `5 s` (5 秒):
  Length of the auto-approve countdown veto window before an unvetoed permission is granted
  (source: facts.md § Auto-approve).

- `2 minutes`:
  Duration of continuous absence required before an unobservable session is removed
  (source: facts.md § Session existence model).

- `0.9.0`:
  Current release version of the application
  (source: facts.md § Identity).
