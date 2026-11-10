/**
 * AI Terminal Pet bridge for DeepSeek Harness.
 *
 * Host-only Cordis plugin. It subscribes to Harness session events and turns
 * them into ai-terminal-pet.v1 status reports. It never computes a pipe name,
 * state set, snapshot layout or redaction rule on its own: all of that comes
 * from the single official protocol module installed at
 * %LOCALAPPDATA%\AI-Terminal-Pet\protocol.js (shipped and checksum-verified
 * by scripts/install.ps1). This file is only a field adapter.
 *
 * Everything is fire-and-forget. No error is ever thrown back into Harness,
 * no process.exit, no blocking wait on the pet application.
 */
import { createRequire } from 'node:module'
import os from 'node:os'
import path from 'node:path'

const require = createRequire(import.meta.url)

const SOURCE = 'deepseek-harness'
const TERMINAL = 'deepseek-harness'
// A short, stable title that is also a substring of the Harness window title.
// That lets the overlay identify the session's window for both liveness and
// one-click jump, while the session itself stays keyed by its Harness id.
const TITLE = 'DeepSeek Harness'
const KEEP_ALIVE_MS = 60_000
const PROTOCOL_FILE = path.join(
  process.env.LOCALAPPDATA || path.join(os.homedir(), 'AppData', 'Local'),
  'AI-Terminal-Pet',
  'protocol.js',
)

let protocol = null
let protocolTried = false

function officialProtocol() {
  if (!protocolTried) {
    protocolTried = true
    try {
      protocol = require(PROTOCOL_FILE)
    } catch {
      protocol = null
    }
  }
  return protocol
}

/** Per-conversation state. Kept at module scope so plugin HMR does not lose it. */
const sessions = new Map()

function metaFor(sessionId) {
  let item = sessions.get(sessionId)
  if (!item) {
    item = {
      state: 'idle',
      project: safeCwd(),
      title: TITLE,
      tool: '',
      reason: '',
      startedAt: new Date().toISOString(),
      stateAt: 0,
      sentAt: 0,
    }
    sessions.set(sessionId, item)
  }
  return item
}

function safeCwd() {
  try {
    return process.cwd()
  } catch {
    return ''
  }
}

function short(value, limit) {
  return String(value == null ? '' : value).slice(0, limit)
}

function sessionIdOf(...candidates) {
  for (const candidate of candidates) {
    const value = candidate?.session?.header?.id
      ?? candidate?.session?.id
      ?? candidate?.header?.id
      ?? candidate?.id
    if (value != null && String(value).length > 0) return String(value)
  }
  return 'dsh-main'
}

/**
 * Publish one status. The protocol module writes the durable snapshot
 * synchronously before its first await, so calling this from process.on('exit')
 * still leaves a correct `state: "closed"` snapshot even when the pipe is gone.
 */
function report(sessionId, patch, force = false) {
  const p = officialProtocol()
  if (!p || typeof p.sendStatus !== 'function') return
  const item = metaFor(sessionId)
  const previous = item.state
  Object.assign(item, patch)
  if (!p.STATES.has(item.state)) item.state = 'idle'

  const now = Date.now()
  const changed = item.state !== previous
  const stale = now - item.sentAt >= KEEP_ALIVE_MS
  if (!force && !changed && !stale) return
  item.sentAt = now
  if (changed) item.stateAt = now

  const stablePid = typeof p.stableProcessId === 'function'
    ? p.stableProcessId(`${SOURCE}:${sessionId}`)
    : 1

  const payload = {
    schemaVersion: p.SCHEMA_VERSION,
    source: SOURCE,
    sessionId,
    parentSessionId: '',
    // Synthetic per-conversation pid (never used for OS liveness) plus the
    // real Harness host pid as the ancestor-chain anchor.
    pid: stablePid,
    parentPid: process.pid,
    project: item.project,
    title: item.title,
    model: '',
    state: item.state,
    reason: item.reason,
    startedAt: item.startedAt,
    updatedAt: new Date(now).toISOString(),
    tool: item.tool,
    terminal: TERMINAL,
  }
  try {
    void p.sendStatus(payload, {
      retryMs: item.state === 'closed' ? 0 : 900,
      launch: item.state !== 'closed',
    }).catch(() => {})
  } catch {
    // Never let a missing pipe or a dead overlay disturb the host CLI.
  }
}

/** Map the durable session events documented in Harness' SessionEventMap. */
function eventPatch(type, data) {
  switch (type) {
    case 'approval/asked':
      return { state: 'waiting_permission', reason: short(data?.reason || 'permission required', 120), tool: short(data?.toolName, 120) }
    case 'approval/decided':
      return { state: 'working', reason: '', tool: '' }
    case 'turn/start':
      return { state: 'working', reason: '', tool: '' }
    case 'turn/end':
      return { state: 'idle', reason: short(data?.reason, 120), tool: '' }
    case 'tool/call':
      return { state: 'tool_use', tool: short(data?.name, 120), reason: '' }
    case 'tool/result':
      return data?.error
        ? { state: 'error', reason: short(data.error.reason || data.error.code, 120), tool: '' }
        : { state: 'working', reason: '', tool: '' }
    case 'command/done':
      return data?.kind === 'error'
        ? { state: 'error', reason: short(data?.text, 120), tool: '' }
        : { state: 'idle', reason: '', tool: '' }
    case 'working-directory/change':
      return { project: short(data?.cwd, 1024) }
    case 'session/title': {
      const name = short(data?.title ?? data?.text ?? data?.name, 120)
      return name ? { title: TITLE, reason: name } : null
    }
    default:
      return null
  }
}

function quiet(fn) {
  return (...args) => {
    try {
      return fn(...args)
    } catch {
      return undefined
    }
  }
}

export function apply(ctx, config) {
  if (config && config.enabled === false) return () => {}

  const off = []
  const on = (name, handler) => {
    try {
      const disposer = ctx.on(name, quiet(handler))
      if (typeof disposer === 'function') off.push(disposer)
    } catch {
      // An event that this Harness version does not expose is simply skipped.
    }
  }

  on('session/event', (session, event) => {
    const sessionId = sessionIdOf(session)
    const patch = eventPatch(event?.type, event?.data)
    if (patch) report(sessionId, patch)
  })

  on('agent/inbox/claimed', (payload) => {
    report(sessionIdOf(payload?.agent), { state: 'working', reason: '', tool: '' })
  })

  on('agent/status', (payload) => {
    const status = String(payload?.status ?? '').toLowerCase()
    if (/run|work|busy|think|stream/.test(status)) {
      report(sessionIdOf(payload?.agent), { state: 'working' })
    } else if (/idle|done|wait|stop/.test(status)) {
      report(sessionIdOf(payload?.agent), { state: 'idle' })
    }
  })

  on('agent/error', (payload) => {
    report(sessionIdOf(payload?.agent), {
      state: 'error',
      reason: short(payload?.error?.message ?? payload?.error, 120),
    })
  })

  // `approval/request` is a waterfall listener: it must always call next(),
  // otherwise other plugins' approval handling is silently dropped.
  on('approval/request', (request, next) => {
    report(sessionIdOf(request?.agent), {
      state: 'waiting_permission',
      reason: 'permission required',
      tool: short(request?.toolName, 120),
    })
    if (typeof next === 'function') return next()
    return undefined
  })

  // Only a real process exit means "all conversations are over". Plugin HMR
  // calls the dispose function below instead, and must NOT clear the panel.
  const onExit = quiet(() => {
    for (const sessionId of sessions.keys()) {
      report(sessionId, { state: 'closed', reason: 'harness exited' }, true)
    }
  })
  try {
    process.on('exit', onExit)
  } catch {}

  return () => {
    try {
      process.off('exit', onExit)
    } catch {}
    for (const disposer of off) {
      try {
        disposer()
      } catch {}
    }
    off.length = 0
  }
}
