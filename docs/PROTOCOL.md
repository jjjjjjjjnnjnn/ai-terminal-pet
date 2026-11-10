# PROTOCOL — ai-terminal-pet 上报协议 v1

> 目标读者：想给 **别的 CLI** 写一个第三方 reporter 的人。
> 本文所有内容都取自真实代码，行号即出处：
> 唯一事实来源是 [reporters/protocol.js](../reporters/protocol.js)（570 行，CommonJS），
> 服务端契约在 [internal/pipeserver/server_windows.go](../internal/pipeserver/server_windows.go)
> 与 [internal/session/session.go](../internal/session/session.go)，
> 快照回退在 [internal/store/snapshot.go](../internal/store/snapshot.go)，
> auto-approve 浮窗侧在 [internal/ui/approve_windows.go](../internal/ui/approve_windows.go)。
> 写实现前请先读 §3 的传输契约和 §4 的字段表，其余章节按需查。

## 1. 一句话模型

host CLI →（fire-and-forget）→ reporter → 一行 JSON → 命名管道 → 浮窗。
管道没人监听时，reporter 把同一份载荷写进快照文件，浮窗稍后扫描时补读。
reporter **永远不能让宿主 CLI 失败**：不抛异常、不阻塞、不改变退出码（§10）。

## 2. schema 版本

| 项 | 值 | 出处 |
| --- | --- | --- |
| 版本字符串 | `ai-terminal-pet.v1` | `reporters/protocol.js:10`、`internal/session/session.go:16` |
| 载荷上限 | 1 MiB（1048576 B，含换行符前的内容） | `session.go:17` `MaxPayloadBytes = 1 << 20`；`protocol.js:28` |
| 传输编码 | UTF-8，一行以 `\n` 结束 | `server_windows.go:164` `ReadSlice('\n')` |

`schemaVersion` 在线上是可选的：Go 端收到空值会补成 `ai-terminal-pet.v1`（`session.go:133-135`），
收到**别的**值直接 `unsupported schema version` → `ERR`。第三方实现请总是显式带上。

## 3. 传输层：命名管道 + JSON-lines + 半关闭

### 3.1 管道名算法

```
\\.\pipe\ai-terminal-pet-v1-<8 hex>
```

`<8 hex>` = **SHA-256( 小写化后的 home 目录，去掉尾部所有 `\` 和 `/` ) 的前 4 字节**（16 进制小写）。

| 端 | 函数 | 出处 |
| --- | --- | --- |
| Go（服务端） | `PipePathForUser()` | `internal/pipeserver/server_windows.go:37-45` |
| JS（客户端） | `pipeScopeToken()` | `reporters/protocol.js:13-24` |

Go 侧：`strings.ToLower(strings.TrimRight(os.UserHomeDir(), \`\\/\`))` → `sha256.Sum256` → `hex(sum[:4])`；
JS 侧：`os.homedir().toLowerCase().replace(/[\\/]+$/, "")` → `sha256` hex → `.slice(0, 8)`。
home 为空时**两边都不加后缀**，退化成 `\\.\pipe\ai-terminal-pet-v1`。

跨语言锁定：`internal/pipeserver/server_windows_test.go:120-130` 的
`TestPipePathForUserMatchesReporterAlgorithm` 用 `USERPROFILE=C:\Users\Pet-Test\` 断言
`sha256("c:\\users\\pet-test")` 的前 4 字节——也就是说"小写 + 去尾部分隔符"是契约的一部分，
不是实现细节。第三方实现必须逐字复刻这两步，否则连不上。

服务端用 `winio.ListenPipe`，安全描述符 `D:P(A;;GRGW;;;OW)(A;;GRGW;;;SY)`
（仅当前用户 + LocalSystem），InputBuffer 64 KiB / OutputBuffer 4 KiB，最多 32 条并发连接
（`server_windows.go:23`）。

### 3.2 请求：一行 JSON，然后半关闭

1. 客户端 `connect` 到上面的管道名。
2. 连接建立后**立即**写入一行 JSON（末尾必须带 `\n`），随后调用 `socket.end()`（半关闭，不再写）。
3. 服务端在 FIN 之前把应答写回，然后关闭连接。

`protocol.js:291-327` 的 `sendOnce()` 就是这条契约的实现（`socket.once("connect", () => socket.end(line, "utf8"))`）。
**不要**先等 `drain`、不要关闭前再发第二行、不要握手——服务端每连接只读一行
（`server_windows.go:137-158` `read()` → `readPayload()`）。

### 3.3 应答：只认精确的 `OK`

| 应答正文 | 含义 | 客户端应做的事 |
| --- | --- | --- |
| `OK\n` | 载荷已解析、已写入内存 store | 成功，本次上报结束 |
| `ERR\n` | 载荷被 `session.ParsePayload` 拒绝（字段越界 / 未知字段 / 时间戳非法 / 超过 1 MiB） | 视为失败；**不要重发同一份载荷**，它还会被拒 |
| 无应答（连接被关） | 读超时（5 s）、载荷超限、空行 | 视为失败（`protocol.js` 归为 `ETIMEDOUT`/`EBADREPLY`） |

`sendOnce` 的成功判据是 `response.trim() === "OK"`（`protocol.js:308`），其它任何回执都算失败
（`EBADREPLY`，`protocol.js:312`）。这一点对第三方实现同样成立：**只有 `OK` 是成功**。

### 3.4 失败模式速查

| 现象 | 原因 | 出处 |
| --- | --- | --- |
| `ENOENT` | 没有进程在监听（浮窗未运行） | 客户端 `net.createConnection` 错误 |
| `ETIMEDOUT` | 服务端 5 s 内没读到完整一行 | `server_windows.go:24` `connectionReadTimeout = 5 * time.Second` |
| `EBADREPLY` | 回了非 `OK` 的内容，或连接提前断开 | `protocol.js:308-313` |
| 连接被直接丢弃 | 并发连接已达 32 条 | `server_windows.go:23,178` `maxConnections = 32` |
| `ERR` | 载荷校验失败 | `session.go:131-190` |

`sendStatus` 的重试策略（`protocol.js:519-547`）：默认总预算 `retryMs = 5000`、
单次连接超时 `connectTimeoutMs = 150`；首次 `ENOENT` 且状态不是 `closed` 时，尝试
`launchApplication()` 拉起浮窗（`protocol.js:277-289`，只认
`%LOCALAPPDATA%\AI-Terminal-Pet\AI-Terminal-Pet.exe`，可用 `AI_TERMINAL_PET_EXE` 覆盖但必须落在该目录内，
除非 `AI_TERMINAL_PET_ALLOW_UNSAFE_EXE=1`）。`closed` 状态的报表**不会**拉起浮窗。

## 4. 载荷字段

除最后一行说明外，每个字段都在 `session.Session`（`session.go:31-47`）里有同名 JSON tag。
"JS 上限"指 `normalizePayload`（`protocol.js:91-127`）在发送前截断的长度，"Go 上限"指
`validate`（`session.go:131-190`）会拒绝的长度；**两者都按字符（rune）计**（`session.go:200`），
JS 侧用的是 `.slice()`（UTF-16 code unit），所以 JS 只会截得更短，不会超限。

| 字段 | 类型 | 必填 | 可空 | JS 上限 | Go 上限 | 含义 / 规范化 |
| --- | --- | --- | --- | --- | --- | --- |
| `schemaVersion` | string | 否（空即补默认） | 否 | 固定 `ai-terminal-pet.v1` | 必须等于该值 | 协议版本 |
| `source` | string | **是** | 否 | 48（`token()` 白名单） | 48，非空 | 上报方标识，如 `claude-hook`/`opencode`/`generic` |
| `sessionId` | string | **是** | 否 | 200（截断后加 `-<8hex>` 摘要） | 200 | 一次会话的稳定 ID，同一会话每次上报必须相同 |
| `parentSessionId` | string | 否 | **是** | 200 | 200 | 子 agent 指向父会话（用于把嵌套会话合并展示） |
| `pid` | int | 是（JS 缺省 `process.pid`） | 否 | — | 1..0xFFFFFFFF，必须 > 0 | 宿主进程 PID；浮窗据此判断"宿主还活着" |
| `parentPid` | int | 否 | **是** | — | 0..0xFFFFFFFF（0 = 无） | 父进程 PID |
| `project` | string | 否 | 是 | 1024 | 1024 | 项目路径 / 工作目录 |
| `title` | string | 否 | 是 | 200 | 200 | 人类可读标题 |
| `model` | string | 否 | 是 | 200 | 200 | 模型名 |
| `state` | string | 是（非法值 JS 会降级成 `idle`） | 否 | 枚举见 §4.1 | 必须在 7 个状态内 | 当前状态 |
| `reason` | string | 否 | 是 | 120（`token()`） | 120 | 状态下方的一行短原因 |
| `tool` | string | 否 | 是 | 120 | 120 | 正在使用的工具名 |
| `terminal` | string | 否 | 是 | 120 | 120 | 终端名 |
| `startedAt` | string (RFC3339/ISO-8601) | **是** | 否 | 解析失败则用进程启动时间 | 不能是零值 | 会话开始时间 |
| `updatedAt` | string (RFC3339/ISO-8601) | **是** | 否 | 解析失败则用 `now` | 不能是零值 | 本状态的时间戳，**必须随状态刷新** |

`parentSessionId` 可空这一点在线上是"省略或 `null`"：JS 侧 `boundedId(..., null)`
（`protocol.js:99`）在缺省时给 `null`；Go 侧 `text()` 对空串返回空串，两者等价。

> **未知字段 = `ERR`**：`ParsePayload` 用 `json.Decoder.DisallowUnknownFields()`
> （`session.go:58`），并且只允许**一个** JSON 值（`session.go:63-65`）。
> 第三方实现不要塞自定义字段——想扩展请走 §11 的版本变更流程。

### 4.1 七个状态

`session.go:21-29` 的常量，正好 7 个：

| 状态 | 含义 | 展示优先级（`session.go:78-95`） |
| --- | --- | --- |
| `waiting_permission` | 等用户批准（auto-approve 相关，见 §7） | 6（最高） |
| `error` | 出错 | 5 |
| `working` | 干活中 | 4 |
| `tool_use` | 正在调用工具（观察性事件） | 3 |
| `thinking` | 模型在思考 | 2 |
| `idle` | 空闲/就绪 | 1 |
| `closed` | 会话结束（tombstone，浮窗消费后删除快照） | 0（最低） |

聚合规则 `Aggregate`（`session.go:97-110`）：`closed` 与非法状态被跳过，取优先级最高者；
并列时按 `updatedAt` 新者优先（`Sort`，`session.go:112-129`）。
所以 `closed` 不是"关掉浮窗"，写好它等于告诉浮窗"这个会话我真的结束了"。

### 4.2 文本规范化

JS 侧 `text(value, limit)`（`protocol.js:45-49`）依次做：
`replace(/[\u0000-\u001f\u007f\u0080-\u009f]/g, " ")` 把控制字符换成空格 →
`replace(/\s+/g, " ")` 压成单空格 → `trim()` → **过一遍 `redact()`（§5）** → `slice(0, limit)` → 空串变 `null`。
`token()`（`protocol.js:60-65`）在 `text()` 之后再套白名单 `[A-Za-z0-9_.:@/-]`，其余字符变 `-`。

Go 侧 `text()`（`session.go:192-209`）**不**替换控制字符，而是**直接拒绝**（`unicode.IsControl`）；
非 UTF-8 也拒绝。也就是说 JS 侧已经清洗过一遍，Go 侧是最后一道闸。

### 4.3 时间戳钳制

| 规则 | JS（发送前，`protocol.js:112-126`） | Go（接收时，`session.go:175-186`） |
| --- | --- | --- |
| 时区 | 一律转 UTC 的 `ISOString` | 接收后转 `UTC()` |
| 两者必填 | 缺省用"进程启动时间"/`now` | 零值 → `timestamps are required` |
| `updatedAt ≥ startedAt` | `startedAt = min(startedAt, updatedAt)` | 违反 → `updatedAt precedes startedAt` |
| 未来上限 | **钳到 `now + 4 分钟`** | 超过 `now + 5 分钟` → 拒绝 |

Go 端是 5 分钟、JS 钳到 4 分钟，中间那 1 分钟是给时钟漂移/传输延迟留的余量。
**这就是为什么不要自己拼时间戳**：直接调 `normalizePayload` 能让时钟快几秒的机器也稳定上报。
钳制是幂等的，因为 `sendStatus` 会对同一份载荷调用两次 `normalizePayload`（`protocol.js:117`）。

## 5. 脱敏（redact）

`redact()` 在 `protocol.js:32-43`，共 **9 条**替换规则。它在 `text()`（`protocol.js:47`）内部被调用，
所以**每个文本字段**（`source`/`sessionId`/`project`/`title`/`model`/`reason`/`tool`/`terminal` …）都会过一遍，
命中就替换（不是拒绝上报）：

| # | 目标 | 正则（简化表述） | 替换为 |
| --- | --- | --- | --- |
| 1 | Bearer token | `/\bBearer\s+[A-Za-z0-9._~+/=-]+/gi` | `Bearer [REDACTED]` |
| 2 | `sk-`/`pk-`/`rk-`/`ak-` 开头的 ≥8 字符 key | `/\b(?:sk\|pk\|rk\|ak)-[A-Za-z0-9_-]{8,}\b/gi` | `[REDACTED]` |
| 3 | GitHub token | `/\b(?:ghp\|gho\|ghu\|ghs\|ghr)_[A-Za-z0-9_]{8,}\b/gi` | `[REDACTED]` |
| 4 | AWS access key id | `/\bAKIA[0-9A-Z]{16}\b/g` | `[REDACTED]` |
| 5 | PEM 私钥块（可跨行） | `/-----BEGIN [^-]+-----…-----END [^-]+-----/g` | `[REDACTED PRIVATE KEY]` |
| 6 | URL 里的 `user:pass@` | `/\b([a-z][a-z0-9+.-]*:\/\/)([^\/\s:@]+):([^\/\s@]+)@/gi` | `$1[REDACTED]@`（scheme 保留） |
| 7 | JWT | `/\beyJ[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}\b/g` | `[REDACTED]` |
| 8 | `api_key`/`access_token`/`auth*`/`password`/`secret` 键值 | `/\b((?:api[_-]?key\|access[_-]?token\|auth(?:orization)?\|password\|secret)\s*[:=]\s*)(…)/gi` | `$1[REDACTED]`（键保留） |
| 9 | CLI 参数 `--api-key`/`--token`/`--password`/`--secret` | `/(--?(?:api-key\|token\|password\|secret)\s+)([^\s,;]+)/gi` | `$1[REDACTED]` |

第三方实现**不强制**做脱敏（服务端不校验），但请务必照做：这些字段最终会显示在
悬浮窗上、写进快照文件、出现在 `--json` 输出里。漏一个就等于把用户的 token 抄到 `%LOCALAPPDATA%`。

## 6. 快照文件回退

管道连不上时（浮窗没运行），reporter 仍然把同一份载荷写进快照文件，浮窗启动后扫描补读。
实现在 `protocol.js:211-260` 与 `internal/store/snapshot.go`。

| 项 | 值 / 规则 | 出处 |
| --- | --- | --- |
| 目录 | `%LOCALAPPDATA%\AI-Terminal-Pet\runtime\` | `protocol.js:211-215` |
| 文件名 | `<source>-<pid>-<sessionId>.json` | `protocol.js:226-232` |
| 文件名安全化 | 非 `[A-Za-z0-9._-]` 的字符 → `-`，保留 48 字符；若与原值不等则取前 39 字符 + `-<8hex>` 摘要 | `protocol.js:217-224` |
| 权限 | `0o600`（仅当前用户可读写） | `protocol.js:240` |
| 原子写 | 先写 `<target>.<pid>.<12hex>.tmp`（`flag: "wx"`），再 `rename` 覆盖 | `protocol.js:237-241` |
| 写失败 | 返回 `false`，**不抛** | `protocol.js:243-248` |
| TTL | 15 s（`DefaultSnapshotTTL`） | `snapshot.go:15` |
| keepalive | 时间戳超 TTL 但宿主进程仍活 → **保留并返回** | `snapshot.go:31-38,128` |
| 文件数上限 | 256（`DefaultMaxSnapshotFiles`），可配到 4096 | `snapshot.go:16-17` |
| `.tmp` 清理 | 超过 TTL 的 `.tmp` 每次扫描最多清 64 个 | `snapshot.go:94-99` |
| 坏文件 | 超限 / 解析失败 / 时间戳非法 / 过期且宿主已死 → 内容未变则删除 | `snapshot.go:116-145` |
| `closed` | 被消费后删除，**不进 key 集合** | `snapshot.go:132-139,59-61` |

keepalive 的意义：reporter 只在**状态变化**时写快照，一个跑了十分钟的 `working` 会话
时间戳会很旧但完全活着。只按 TTL 删文件曾导致面板和 `--json` 丢会话，所以
`KeepAlive`（`internal/app/app_windows.go:74,537` 传 `presence.SessionAlive`）把"宿主还在"作为第二判据。

## 7. auto-approve 的 `approvals/` 握手

浮窗是只读面板，点不到别人终端里的"允许"。所以自动批准是**反向**的：hook 阻塞一小段时间
等浮窗的否决，**沉默即批准**。实现在 `protocol.js:333-517` 与 `internal/ui/approve_windows.go`。

| 项 | 值 | 出处 |
| --- | --- | --- |
| 目录 | `%LOCALAPPDATA%\AI-Terminal-Pet\approvals\` | `protocol.js:406-409` |
| 请求文件名 | `pending-<sanitizeStem(source)>-<sanitizeStem(sessionId)>-<pid>-<epochms>.json` | `protocol.js:428-453` |
| 请求体 | `{"source","sessionId","pid","tool","createdAt"}` | `protocol.js:438-444` |
| 否决文件名 | **同名文件 + `.deny`** | `protocol.js:459`；`approve_windows.go:257` |
| 否决体 | `{"decision":"deny","at":"<RFC3339>"}` | `approve_windows.go:240` |
| 轮询间隔 | 100 ms（`APPROVAL_POLL_MS`） | `protocol.js:355` |
| 默认超时 | 5 s（`DEFAULT_APPROVE_TIMEOUT_SEC`），配置范围 1–60 s | `protocol.js:357,394-399` |
| 文件权限 | hook 侧 `0o600`，先写 `.tmp` 再 `rename` | `protocol.js:446-448` |
| 配置来源 | `%LOCALAPPDATA%\AI-Terminal-Pet\config.json` 的 `autoApprove` / `autoApproveTimeoutSec`；**读失败一律按关闭** | `protocol.js:384-404`、`internal/config/config.go:127-130` |

决策树 `approvalPlan()`（`protocol.js:487-492`）三态：

- `off` —— 配置没开，维持宿主原行为（不介入）。
- `manual` —— 命中危险模式，**永不自动批准**（Claude 走终端原生提示，Antigravity 直接拒绝）。
- `auto` —— 写 pending 文件，等完否决窗口。

`DANGEROUS_PATTERNS` 共 **11 条**（`protocol.js:364-376`，大小写不敏感，匹配"工具名 + 工具输入"）：

| # | 匹配 | 拦截的命令 |
| --- | --- | --- |
| 1 | `\brm\s+[^;\|&]*-[a-z]*r` | `rm -r` / `rm -rf` |
| 2 | `\brd\s+\/s` | `rd /s` |
| 3 | `\bmv\b[^;\|&]*\/dev\/null` | `mv … /dev/null` |
| 4 | `\bmkfs\b` | `mkfs` |
| 5 | `\bdd\b[^;\|&]*\bof\s*=\s*\/dev\/` | `dd of=/dev/…` |
| 6 | `:\(\)\s*\{` | fork bomb |
| 7 | `\bgit\s+push\b[^;\|&]*--force` | `git push --force` |
| 8 | `\bformat\s+[a-zA-Z]:` | `format C:` |
| 9 | `\bdel\b[^;\|&]*\\windows` | `del … \Windows` |
| 10 | `\bshutdown\s+\/s` | `shutdown /s` |
| 11 | `remove-item\b[^;\|&]*system32` | `Remove-Item … System32` |

失败即拒绝（fail-closed）：`writeApprovalPending` 失败、`pollApprovalDeny` 异常、
任何未预期错误 → `waitForApprovalDecision` 返回 `"deny"` 且**永不抛**
（`protocol.js:499-517`）。第三方实现照此办理：宁可多问一次，也不要静默放行。

浮窗侧的 `approvalTracker`（`approve_windows.go:124-144`）只给 `waiting_permission` 的会话计时，
`updatedAt` 变化即重启倒计时（新的一轮权限请求），会话离开该状态就立刻丢计时器（不会"幽灵批准"）。

## 8. 最小 reporter 实现（Node.js，约 20 行）

```js
// reporter.js - 最小可用的第三方 reporter（Node.js >= 16，零依赖）
"use strict"
const net = require("node:net")
const crypto = require("node:crypto")
const os = require("node:os")

const PIPE = "\\\\.\\pipe\\ai-terminal-pet-v1-" + crypto
  .createHash("sha256")
  .update(os.homedir().toLowerCase().replace(/[\\/]+$/, ""), "utf8")
  .digest("hex").slice(0, 8)

function report(fields) {
  const now = new Date().toISOString()
  const line = JSON.stringify({
    schemaVersion: "ai-terminal-pet.v1",
    source: "my-cli", sessionId: "my-cli:" + (process.argv[2] || "default"),
    pid: process.pid, state: fields.state, startedAt: fields.startedAt || now,
    updatedAt: now, project: process.cwd(), reason: fields.reason || null,
  }) + "\n"
  const socket = net.createConnection({ path: PIPE })
  let reply = ""
  socket.on("data", (chunk) => { reply += chunk.toString("utf8") })
  socket.once("connect", () => socket.end(line, "utf8"))   // 半关闭：一行后立刻 end
  socket.once("end", () => { /* reply.trim() === "OK" 才算成功 */ })
  socket.on("error", () => {})                             // 管道不在 = 正常，静默
  setTimeout(() => socket.destroy(), 2000).unref()         // 别拖住宿主 CLI
}

report({ state: process.argv[3] || "idle" })                // node reporter.js <sess> <state>
```

想省事就直接用仓库里的 `reporters/generic.js`（外壳别名级接入）：

```
node reporters/generic.js --source=aider --state=working --project="$PWD"
node reporters/generic.js --source=codex --state=waiting_permission --reason=approve
node reporters/generic.js --source=mybot --stdin      # stdin 上给一个 JSON 对象
```

## 9. 用别的语言实现的注意清单

| # | 要点 | 为什么 |
| --- | --- | --- |
| 1 | 管道名必须自己算 `sha256(home 小写、去尾部分隔符)` 的前 4 字节 | 服务端不广播名字，算错就是永久 `ENOENT`；§3.1 有 golden test 钉住 |
| 2 | Windows 上 `home` 取 `USERPROFILE`（`os.UserHomeDir()` 的 Go 语义） | 别用 `HOMEDRIVE`+`HOMEPATH` 的拼接变体，大小写或分隔符一差就 mismatch |
| 3 | 写完一行**立刻半关闭**，然后等应答 | 服务端每连接只读一行且在读超时后关连接（§3.2） |
| 4 | 成功判据是 `trim() == "OK"`，不是"连接成功" | 连接成功但回 `ERR` 仍是失败（§3.3） |
| 5 | 别发未知字段 | `DisallowUnknownFields` → `ERR`（§4） |
| 6 | 长度按**字符**数，不按字节 | Go 用 `utf8.RuneCountInString`（`session.go:200`） |
| 7 | 时间戳用 UTC（RFC3339），且 `updatedAt ≥ startedAt`、不超过 `now+5min` | 否则被拒并被快照清理器删掉（§4.3） |
| 8 | 同一会话的 `sessionId` 必须稳定 | 它是 store 的键，抖动一次就多一条会话 |
| 9 | 连不上管道时**照常写快照** | 浮窗未运行时这是唯一的传递通道（§6） |
| 10 | 无论成败：不抛、不阻塞、exit 0 | 见 §10，这是硬约束 |
| 11 | 没有真实 PID 的宿主（如某些 IDE 插件）可以用 FNV-1a 派生一个稳定正整数 | 参考 `stableProcessId()`（`protocol.js:72-80`） |
| 12 | 想自动批准就走 `approvals/` 握手，且自己实现危险命令拦截 | 只靠浮窗的 11 条模式不够，钩子侧也应有兜底（§7） |

## 10. 错误处理契约（硬约束）

| 约束 | 做法 | 出处 |
| --- | --- | --- |
| 永不抛进宿主 CLI | 所有 I/O 包在 `try/catch`；`main().catch(() => {})` | `reporters/generic.js` 等全部 5 个 hook |
| 永不阻塞 | 单次连接 150 ms、总预算 5 s；stdin 读取上限 1 MiB / 2 s | `protocol.js:519-547`、`:28-29` |
| 退出码恒为 0 | 除非纯用法错误（exit 2） | `generic.js:21-23` |
| 静默降级 | 写快照失败返回 `false`；读配置失败按"关闭"；审批失败按"拒绝" | `protocol.js:243-248,401-403,513-515` |

一句话：**reporter 坏了不能连累宿主**。这条比"上报成功"更重要。

## 11. 不兼容变更策略

`schemaVersion` 的规则：

| 变更类型 | 是否需要改版本 | 例子 |
| --- | --- | --- |
| 新增**可选**字段，服务端忽略未知字段 | 需要，但可以先在 Go 端放开 | 见下方说明 |
| 修改已有字段语义 / 类型 / 单位 | **必须**，`v1 → v2` | `pid` 从进程号改成别的标识 |
| 收紧长度上限或校验（原本合法的载荷开始被拒） | **必须** | `source` 48 → 32 |
| 删除字段 | **必须** | 去掉 `terminal` |
| 改管道名、改半关闭契约、改 `OK` 判据 | **必须**（等同于换传输协议） | 管道名改成 session 作用域 |
| 新增状态枚举值 | **必须**（老浮窗会把它当非法状态） | 加 `paused` |
| 内部实现优化（脱敏正则加一条、清理策略调整） | 否 | 加一个新 token 前缀的脱敏规则 |
| 放宽校验（原本被拒的现在收） | 否 | `title` 200 → 300 的过渡期兼容 |

操作建议：

1. **加字段**要分两步：先让 Go 端接受（这一步本身就已经是 breaking，因为当前是
   `DisallowUnknownFields`），再改 JS，最后才让第三方使用。当前 `v1` 下没有"安全加字段"的空间——
   任何新字段都必须伴随一次服务端侧的前置放开。
2. 版本号是`ai-terminal-pet.vN`，服务端**只接受一个**版本（`session.go:136-138`），
没有协商、没有降级。跨版本过渡只能靠发布说明和 `scripts/install.ps1` 一起升级。
3. 升级顺序：**浮窗先、reporter 后**。老 reporter + 新浮窗（v1 继续被接受）安全；
   新 reporter + 老浮窗会被 `unsupported schema version` 顶回 `ERR`（上报静默丢失，但不影响宿主）。
4. 破坏性变更要在 `CHANGELOG.md` 里显式标 `BREAKING`，并更新本文的版本表。

## 附录 A：一份完整载荷（真实形态）

```json
{
  "schemaVersion": "ai-terminal-pet.v1",
  "source": "claude-hook",
  "sessionId": "9f3a1c2e-...",
  "parentSessionId": null,
  "pid": 27788,
  "parentPid": 14420,
  "project": "C:\\Users\\rongj\\Desktop\\项目开发\\ai-terminal-pet",
  "title": "refactor pipeserver",
  "model": "deepseek-v4",
  "state": "waiting_permission",
  "reason": "approve",
  "startedAt": "2026-11-09T15:02:11.000Z",
  "updatedAt": "2026-11-09T15:11:40.000Z",
  "tool": "Bash",
  "terminal": "Windows Terminal"
}
```

（`null` 的字段在线上可以省略；反斜杠按 JSON 规则转义。）

## 附录 B：怎么验证你的实现

| 检查 | 命令 / 位置 |
| --- | --- |
| 管道名是否算对 | `go test ./internal/pipeserver -run TestPipePathForUserMatchesReporterAlgorithm` |
| JS 侧规范化行为 | `node integrations/deepseek-harness/verify.mjs`（含时钟偏移、幂等、UTC 等 14 条断言） |
| 全部 reporter 语法 | 逐文件循环：`for f in reporters/*.js; do node --check "$f"; done`（`node --check a.js b.js` 只检查第一个参数，Node 24 实测） |
| 手动打一发 | `node reporters/generic.js --source=ci --state=error --reason=test` |
| 浮窗是否收到 | `AI-Terminal-Pet.exe --json`（无会话时为空数组，有则含你刚发的 `sessionId`） |

最后更新：2026-11（对应 v0.9.0，`reporters/protocol.js` 570 行 / 20,914 B）。
