"use strict"

const crypto = require("node:crypto")
const fs = require("node:fs")
const net = require("node:net")
const os = require("node:os")
const path = require("node:path")
const { spawn } = require("node:child_process")

const SCHEMA_VERSION = "ai-terminal-pet.v1"
const PIPE_BASE_PATH = "\\\\.\\pipe\\ai-terminal-pet-v1"

function pipeScopeToken() {
  let home = ""
  try {
    home = os.homedir() || ""
  } catch {
    home = ""
  }
  if (!home) return ""
  const normalized = home.toLowerCase().replace(/[\\/]+$/, "")
  if (!normalized) return ""
  return crypto.createHash("sha256").update(normalized, "utf8").digest("hex").slice(0, 8)
}

const PIPE_SCOPE_TOKEN = pipeScopeToken()
const PIPE_PATH = PIPE_SCOPE_TOKEN ? `${PIPE_BASE_PATH}-${PIPE_SCOPE_TOKEN}` : PIPE_BASE_PATH
const MAX_STDIN_BYTES = 1 << 20
const STDIN_TIMEOUT_MS = 2000
const STATES = new Set(["idle", "thinking", "working", "tool_use", "waiting_permission", "error", "closed"])

function redact(value) {
  return value
    .replace(/\bBearer\s+[A-Za-z0-9._~+/=-]+/gi, "Bearer [REDACTED]")
    .replace(/\b(?:sk|pk|rk|ak)-[A-Za-z0-9_-]{8,}\b/gi, "[REDACTED]")
    .replace(/\b(?:ghp|gho|ghu|ghs|ghr)_[A-Za-z0-9_]{8,}\b/gi, "[REDACTED]")
    .replace(/\bAKIA[0-9A-Z]{16}\b/g, "[REDACTED]")
    .replace(/-----BEGIN [^-]+-----[\s\S]*?-----END [^-]+-----/g, "[REDACTED PRIVATE KEY]")
    .replace(/\b([a-z][a-z0-9+.-]*:\/\/)([^\/\s:@]+):([^\/\s@]+)@/gi, "$1[REDACTED]@")
    .replace(/\beyJ[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}\b/g, "[REDACTED]")
    .replace(/\b((?:api[_-]?key|access[_-]?token|auth(?:orization)?|password|secret)\s*[:=]\s*)("[^"]*"|'[^']*'|[^\s,;]+)/gi, "$1[REDACTED]")
    .replace(/(--?(?:api-key|token|password|secret)\s+)([^\s,;]+)/gi, "$1[REDACTED]")
}

function text(value, limit) {
  if (typeof value !== "string" && typeof value !== "number" && typeof value !== "boolean") return null
  const clean = redact(String(value).replace(/[\u0000-\u001f\u007f\u0080-\u009f]/g, " ").replace(/\s+/g, " ").trim()).slice(0, limit)
  return clean || null
}

function boundedId(value, limit, fallback) {
  const raw = value == null ? "" : String(value)
  if (!raw.trim()) return fallback
  if (raw.length <= limit) return text(raw, limit) || fallback
  const digest = crypto.createHash("sha256").update(raw, "utf8").digest("hex").slice(0, 8)
  const stem = text(raw, limit - 9) || fallback
  return `${stem}-${digest}`
}

function token(value, limit, fallback) {
  const clean = text(value, limit)
  if (!clean) return fallback
  const safe = clean.replace(/[^A-Za-z0-9_.:@/-]+/g, "-").replace(/^[. -]+|[. -]+$/g, "").slice(0, limit)
  return safe || fallback
}

function positiveInteger(value, fallback) {
  const number = Number(value)
  return Number.isSafeInteger(number) && number > 0 && number <= 0xffffffff ? number : fallback
}

function stableProcessId(value) {
  const input = String(value ?? "")
  let hash = 0x811c9dc5
  for (let index = 0; index < input.length; index++) {
    hash ^= input.charCodeAt(index)
    hash = Math.imul(hash, 0x01000193)
  }
  return ((hash >>> 0) % 0x7fffffff) + 1
}

function timestamp(value, fallback) {
  let number = value
  if (typeof value === "string") number = Date.parse(value)
  if (typeof number !== "number" || !Number.isFinite(number)) return fallback
  if (number > 0 && number < 100000000000) number *= 1000
  const date = new Date(number)
  return Number.isNaN(date.getTime()) ? fallback : date.toISOString()
}

function normalizePayload(payload) {
  const input = payload && typeof payload === "object" ? payload : {}
  const now = new Date().toISOString()
  const processStartedAt = new Date(Date.now() - process.uptime() * 1000).toISOString()
  const result = {
    schemaVersion: SCHEMA_VERSION,
    source: token(input.source, 48, "unknown"),
    sessionId: boundedId(input.sessionId, 200, "unknown"),
    parentSessionId: boundedId(input.parentSessionId, 200, null),
    pid: positiveInteger(input.pid, process.pid),
    parentPid: input.parentPid == null ? positiveInteger(process.ppid, null) : positiveInteger(input.parentPid, null),
    project: text(input.project, 1024),
    title: text(input.title, 200),
    model: text(input.model, 200),
    state: STATES.has(input.state) ? input.state : "idle",
    reason: token(input.reason, 120, null),
    startedAt: timestamp(input.startedAt, processStartedAt),
    updatedAt: timestamp(input.updatedAt, now),
    tool: text(input.tool, 120),
    terminal: text(input.terminal, 120)
  }
  // The server rejects "updatedAt precedes startedAt" and timestamps more
  // than 5 minutes in the future, and the snapshot scanner then deletes the
  // file. A reporter whose clock is a little fast - or that passed a local
  // time as UTC - would therefore lose its whole status silently. Clamp it
  // here, in the single protocol source, so every adapter stays immune.
  // This stays idempotent: sendStatus() calls normalizePayload twice.
  const started = Date.parse(result.startedAt)
  const updated = Date.parse(result.updatedAt)
  if (Number.isFinite(started) && Number.isFinite(updated)) {
    const ceiling = Date.now() + 4 * 60 * 1000
    const clamp = (millis) => new Date(Math.min(millis, ceiling)).toISOString()
    result.startedAt = clamp(Math.min(started, updated))
    result.updatedAt = clamp(updated)
  }
  return result
}

function stdinError(message, code) {
  return Object.assign(new Error(message), { code })
}

function readStdin(options = {}) {
  options = options && typeof options === "object" ? options : {}
  const stream = options.stream && typeof options.stream.on === "function" ? options.stream : process.stdin
  const maxBytes = Math.min(MAX_STDIN_BYTES, positiveInteger(options.maxBytes, MAX_STDIN_BYTES))
  const timeoutMs = Math.min(STDIN_TIMEOUT_MS, positiveInteger(options.timeoutMs, STDIN_TIMEOUT_MS))
  return new Promise((resolve, reject) => {
    const chunks = []
    let size = 0
    let settled = false
    let timer

    function cleanup() {
      if (timer) clearTimeout(timer)
      if (typeof stream.removeListener === "function") {
        stream.removeListener("data", onData)
        stream.removeListener("end", onEnd)
        stream.removeListener("close", onClose)
        stream.removeListener("error", onError)
      }
      if (typeof stream.pause === "function") {
        try {
          stream.pause()
        } catch {}
      }
    }

    function finish(error) {
      if (settled) return
      settled = true
      cleanup()
      if (error) {
        reject(error)
        return
      }
      resolve(Buffer.concat(chunks, size))
    }

    function onData(chunk) {
      const buffer = Buffer.isBuffer(chunk) ? chunk : Buffer.from(String(chunk))
      const remaining = maxBytes - size
      if (remaining > 0) chunks.push(buffer.subarray(0, remaining))
      size += buffer.length
      if (size > maxBytes) finish(stdinError("stdin input too large", "E2BIG"))
    }

    function onEnd() {
      finish()
    }

    function onClose() {
      finish(stdinError("stdin closed before end", "ESTREAMCLOSED"))
    }

    function onError(error) {
      finish(error || stdinError("stdin read failed", "ESTREAMFAILED"))
    }

    stream.on("data", onData)
    stream.once("end", onEnd)
    stream.once("close", onClose)
    stream.once("error", onError)
    if (stream.readableEnded) {
      finish()
      return
    }
    if (stream.destroyed) {
      finish(stdinError("stdin is unavailable", "ESTREAMCLOSED"))
      return
    }
    timer = setTimeout(() => finish(stdinError("stdin input timed out", "ETIMEDOUT")), timeoutMs)
    try {
      stream.resume()
    } catch (error) {
      finish(error)
    }
  })
}

function runtimeDirectory(options) {
  if (options && typeof options.runtimeDir === "string" && options.runtimeDir) return options.runtimeDir
  if (!process.env.LOCALAPPDATA) return null
  return path.join(process.env.LOCALAPPDATA, "AI-Terminal-Pet", "runtime")
}

function filenamePart(value, fallback) {
  const input = String(value || fallback)
  const safe = input.replace(/[^A-Za-z0-9._-]+/g, "-").replace(/^[. -]+|[. -]+$/g, "").slice(0, 48)
  if (safe && safe === input) return safe
  const stem = safe.slice(0, 39) || fallback
  const digest = crypto.createHash("sha256").update(input).digest("hex").slice(0, 8)
  return `${stem}-${digest}`
}

function snapshotPath(payload, options = {}) {
  const directory = runtimeDirectory(options)
  if (!directory) return null
  const item = normalizePayload(payload)
  const name = `${filenamePart(item.source, "unknown")}-${item.pid}-${filenamePart(item.sessionId, "unknown")}.json`
  return path.join(directory, name)
}

function writeRuntimeSnapshot(payload, options = {}) {
  const target = snapshotPath(payload, options)
  if (!target || options.writeSnapshot === false) return false
  const temporary = `${target}.${process.pid}.${crypto.randomBytes(6).toString("hex")}.tmp`
  try {
    fs.mkdirSync(path.dirname(target), { recursive: true })
    fs.writeFileSync(temporary, `${JSON.stringify(normalizePayload(payload))}\n`, { encoding: "utf8", flag: "wx", mode: 0o600 })
    fs.renameSync(temporary, target)
    return true
  } catch {
    try {
      fs.unlinkSync(temporary)
    } catch {}
    return false
  }
}

function removeRuntimeSnapshot(payload, options = {}) {
  const target = snapshotPath(payload, options)
  if (!target) return false
  try {
    fs.unlinkSync(target)
    return true
  } catch {
    return false
  }
}

function executablePath() {
  const root = process.env.LOCALAPPDATA ? path.join(process.env.LOCALAPPDATA, "AI-Terminal-Pet") : null
  const override = process.env.AI_TERMINAL_PET_EXE
  if (override) {
    if (!root) return null
    if (process.env.AI_TERMINAL_PET_ALLOW_UNSAFE_EXE === "1") return path.resolve(override)
    const resolvedRoot = path.resolve(root)
    const resolved = path.resolve(override)
    const relative = path.relative(resolvedRoot, resolved)
    if (relative.startsWith("..") || path.isAbsolute(relative)) return null
    return resolved
  }
  return root ? path.join(root, "AI-Terminal-Pet.exe") : null
}

function launchApplication() {
  const executable = executablePath()
  if (!executable) return false
  try {
    if (!fs.statSync(executable).isFile()) return false
    const child = spawn(executable, [], { detached: true, stdio: "ignore", windowsHide: true })
    child.once("error", () => {})
    child.unref()
    return true
  } catch {
    return false
  }
}

function sendOnce(pipePath, line, timeoutMs) {
  // Half-close contract: send one JSON line then half-close (socket.end);
  // the Go server replies OK/ERR before FIN. Only exact "OK" counts —
  // anything else is EBADREPLY so sendStatus retries until its deadline.
  return new Promise((resolve) => {
    let settled = false
    let socket
    let timer
    let response = ""
    const finish = (error) => {
      if (settled) return
      settled = true
      clearTimeout(timer)
      if (socket) socket.destroy()
      resolve({ ok: !error, code: error && error.code })
    }
    const complete = () => {
      if (response.trim() === "OK") {
        finish()
        return
      }
      finish(Object.assign(new Error("pipe did not acknowledge status"), { code: "EBADREPLY" }))
    }
    try {
      socket = net.createConnection({ path: pipePath })
      timer = setTimeout(() => finish(Object.assign(new Error("pipe timeout"), { code: "ETIMEDOUT" })), Math.max(1, timeoutMs))
      socket.on("data", (chunk) => {
        if (response.length < 32) response += chunk.toString("utf8")
      })
      socket.once("connect", () => socket.end(line, "utf8"))
      socket.once("end", complete)
      socket.once("error", finish)
    } catch (error) {
      finish(error)
    }
  })
}

function delay(milliseconds) {
  return new Promise((resolve) => setTimeout(resolve, milliseconds))
}

// ---- Auto-approve handshake (hook side of approvals/) ----
//
// The overlay is a read-only panel: it cannot click "allow" inside a
// terminal. So auto-approve works the other way round - the hook blocks
// for up to the configured timeout waiting for a veto from the overlay,
// and silence means approval:
//
//   hook    writes approvals/pending-<...>.json, polls for a sibling
//           <same>.deny every 100 ms, deletes its pending file on exit.
//   overlay while a waiting_permission session is on screen with
//           auto-approve on, runs a countdown per session. A tap on the
//           expanded panel writes the .deny file; the countdown reaching
//           zero writes nothing - the hook treats the missing veto as
//           approval.
//
// The veto window applies to the permission-prompt path only (events the
// hook classifies as waiting_permission). Plain tool observations
// (PreToolUse/tool_use that will not prompt) never block: stalling every
// tool call by the full timeout would make the agent unusable.

// APPROVAL_POLL_MS is the veto check interval. 100 ms keeps the deny
// snappy without spinning on the filesystem.
const APPROVAL_POLL_MS = 100
const APPROVAL_PENDING_PREFIX = "pending-"
const DEFAULT_APPROVE_TIMEOUT_SEC = 5

// DANGEROUS_PATTERNS never auto-approve: a match forces the manual path
// (Claude: fall through to the terminal's native prompt; Antigravity:
// deny). Case-insensitive, matched against "tool name + tool input".
// Extend freely - a missed pattern auto-approves, so err on the broad
// side for genuinely destructive commands.
const DANGEROUS_PATTERNS = [
  /\brm\s+[^;|&]*-[a-z]*r/i, // rm -r / rm -rf
  /\brd\s+\/s/i, // rd /s (Windows recursive delete)
  /\bmv\b[^;|&]*\/dev\/null/i, // mv ... /dev/null
  /\bmkfs\b/i, // make filesystem
  /\bdd\b[^;|&]*\bof\s*=\s*\/dev\//i, // dd of=/dev/...
  /:\(\)\s*\{/, // fork bomb :(){ :|:& };
  /\bgit\s+push\b[^;|&]*--force/i, // git push --force
  /\bformat\s+[a-zA-Z]:/i, // format C:
  /\bdel\b[^;|&]*\\windows/i, // del ... \Windows
  /\bshutdown\s+\/s/i, // shutdown /s
  /remove-item\b[^;|&]*system32/i // Remove-Item ... System32
]

function isDangerousCommand(...parts) {
  const text = parts.filter((part) => part != null && String(part).trim()).join("\n")
  if (!text) return false
  return DANGEROUS_PATTERNS.some((pattern) => pattern.test(text))
}

// loadPetConfig reads the overlay's config.json for the auto-approve
// switch. Any failure (missing file, bad JSON, no LOCALAPPDATA) means
// "off": a hook that cannot read the user's intent must not approve.
function loadPetConfig() {
  const fallback = { autoApprove: false, autoApproveTimeoutSec: DEFAULT_APPROVE_TIMEOUT_SEC }
  try {
    if (!process.env.LOCALAPPDATA) return fallback
    const raw = fs.readFileSync(path.join(process.env.LOCALAPPDATA, "AI-Terminal-Pet", "config.json"), "utf8")
    const parsed = JSON.parse(raw)
    if (!parsed || typeof parsed !== "object") return fallback
    const timeout = Number(parsed.autoApproveTimeoutSec)
    return {
      autoApprove: parsed.autoApprove === true,
      autoApproveTimeoutSec: Number.isInteger(timeout)
        ? Math.min(60, Math.max(1, timeout))
        : DEFAULT_APPROVE_TIMEOUT_SEC
    }
  } catch {
    return fallback
  }
}

function approvalsDirectory() {
  if (!process.env.LOCALAPPDATA) return null
  return path.join(process.env.LOCALAPPDATA, "AI-Terminal-Pet", "approvals")
}

// sanitizeStem mirrors sanitizeApprovalStem in internal/ui: only
// [A-Za-z0-9_-], truncated to 64 chars. Stems are cosmetic (the JSON
// body carries the exact identity), so byte-vs-char differences for
// non-ASCII session ids do not matter.
function sanitizeStem(value) {
  const input = String(value ?? "")
  let out = ""
  for (const char of input) {
    if (out.length >= 64) break
    out += /[A-Za-z0-9_-]/.test(char) ? char : "_"
  }
  return out || "session"
}

// writeApprovalPending files one request for the overlay to veto.
// Returns the pending path, or null when it cannot be filed. Atomic
// (temp + rename) so the overlay never reads a half-written file.
function writeApprovalPending(info) {
  try {
    const directory = approvalsDirectory()
    if (!directory) return null
    const item = info && typeof info === "object" ? info : {}
    const source = String(item.source || "hook")
    const sessionId = String(item.sessionId || "global")
    const pid = Number.isSafeInteger(item.pid) ? item.pid : process.pid
    const name = `${APPROVAL_PENDING_PREFIX}${sanitizeStem(source)}-${sanitizeStem(sessionId)}-${pid}-${Date.now()}.json`
    const target = path.join(directory, name)
    const body = JSON.stringify({
      source,
      sessionId,
      pid,
      tool: item.tool == null ? null : String(item.tool),
      createdAt: new Date().toISOString()
    })
    fs.mkdirSync(directory, { recursive: true })
    const temporary = `${target}.${process.pid}.tmp`
    fs.writeFileSync(temporary, `${body}\n`, { encoding: "utf8", mode: 0o600 })
    fs.renameSync(temporary, target)
    return target
  } catch {
    return null
  }
}

// pollApprovalDeny waits up to timeoutMs for the overlay's veto file.
// True = the user tapped deny on the panel. Never throws.
function pollApprovalDeny(pendingPath, timeoutMs) {
  const deadline = Date.now() + Math.max(0, Number(timeoutMs) || 0)
  const veto = `${pendingPath}.deny`
  return (async () => {
    try {
      while (true) {
        try {
          if (fs.existsSync(veto)) return true
        } catch {}
        if (Date.now() >= deadline) return false
        await delay(Math.min(APPROVAL_POLL_MS, Math.max(1, deadline - Date.now())))
      }
    } catch {
      return false
    }
  })()
}

function removeApprovalPending(pendingPath) {
  if (!pendingPath) return
  try {
    fs.unlinkSync(pendingPath)
  } catch {}
}

// approvalPlan is the hook's decision tree for one permission gate:
// "off"    auto-approve is disabled - keep the legacy behaviour.
// "manual" dangerous command - never auto-approve (Claude falls through
//          to the terminal's native prompt, Antigravity denies).
// "auto"   file a pending request and wait out the veto window.
function approvalPlan(toolText) {
  const config = loadPetConfig()
  if (!config.autoApprove) return { mode: "off", timeoutMs: 0 }
  if (isDangerousCommand(toolText)) return { mode: "manual", timeoutMs: 0 }
  return { mode: "auto", timeoutMs: config.autoApproveTimeoutSec * 1000 }
}

// waitForApprovalDecision files a pending request and blocks for the
// veto window. Returns "allow" (silence) or "deny" (the user tapped the
// panel). Fail-closed: when the request cannot even be filed, or
// anything unexpected happens, it returns "deny". Never throws - a hook
// must never crash its host CLI.
function waitForApprovalDecision(info, options = {}) {
  return (async () => {
    try {
      const opts = options && typeof options === "object" ? options : {}
      const timeoutMs = Number.isFinite(Number(opts.timeoutMs)) && Number(opts.timeoutMs) > 0
        ? Number(opts.timeoutMs)
        : DEFAULT_APPROVE_TIMEOUT_SEC * 1000
      const pendingPath = writeApprovalPending(info)
      if (!pendingPath) return "deny"
      try {
        return (await pollApprovalDeny(pendingPath, timeoutMs)) ? "deny" : "allow"
      } finally {
        removeApprovalPending(pendingPath)
      }
    } catch {
      return "deny"
    }
  })()
}

async function sendStatus(payload, options = {}) {
  options = options && typeof options === "object" ? options : {}
  let item
  try {
    item = normalizePayload(payload)
    writeRuntimeSnapshot(item, options)
  } catch {
    return false
  }
  const line = `${JSON.stringify(item)}\n`
  const pipePath = typeof options.pipePath === "string" && options.pipePath ? options.pipePath : PIPE_PATH
  const retryMs = Number.isFinite(Number(options.retryMs)) ? Math.max(0, Number(options.retryMs)) : 5000
  const connectTimeoutMs = Number.isFinite(Number(options.connectTimeoutMs)) ? Math.max(1, Number(options.connectTimeoutMs)) : 150
  const deadline = Date.now() + retryMs
  const launchAllowed = item.state !== "closed" && options.launch !== false
  let launchAttempted = false
  while (true) {
    const remaining = Math.max(1, deadline - Date.now())
    const result = await sendOnce(pipePath, line, Math.min(connectTimeoutMs, remaining))
    if (result.ok) return true
    if (result.code === "ENOENT" && !launchAttempted && launchAllowed) {
      launchAttempted = true
      launchApplication()
    }
    const wait = Math.min(100, deadline - Date.now())
    if (wait <= 0) return false
    await delay(wait)
  }
}

module.exports = {
  APPROVAL_POLL_MS,
  DANGEROUS_PATTERNS,
  PIPE_PATH,
  SCHEMA_VERSION,
  STATES,
  approvalPlan,
  executablePath,
  isDangerousCommand,
  loadPetConfig,
  normalizePayload,
  pollApprovalDeny,
  readStdin,
  removeApprovalPending,
  removeRuntimeSnapshot,
  sanitizeStem,
  sendStatus,
  snapshotPath,
  stableProcessId,
  waitForApprovalDecision,
  writeApprovalPending
}
