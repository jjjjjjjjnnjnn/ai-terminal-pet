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
  return {
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
  PIPE_PATH,
  SCHEMA_VERSION,
  STATES,
  executablePath,
  normalizePayload,
  readStdin,
  removeRuntimeSnapshot,
  sendStatus,
  snapshotPath,
  stableProcessId
}
