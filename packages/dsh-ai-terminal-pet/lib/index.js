/**
 * dsh-ai-terminal-pet - a DeepSeek Harness Cordis bundle that reports session
 * state to the AI Terminal Pet desktop overlay.
 *
 * The plugin is a field adapter only. It never computes the pipe name, the
 * state set, the snapshot path or the redaction rules: those come from the
 * official protocol module, resolved at first use from
 *   %LOCALAPPDATA%\AI-Terminal-Pet\protocol.js   (placed there by install.ps1)
 * and, when that file is missing, from the copy shipped in this package
 * (vendor/protocol.js).
 *
 * Everything is fire-and-forget. Nothing here may throw into the Harness host
 * process, nothing calls process.exit, and nothing waits on the overlay.
 */
import { createHash } from 'node:crypto'
import {
  chmodSync,
  copyFileSync,
  existsSync,
  mkdirSync,
  mkdtempSync,
  readFileSync,
  renameSync,
  rmSync,
  writeFileSync,
} from 'node:fs'
import { createRequire } from 'node:module'
import os from 'node:os'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const require = createRequire(import.meta.url)

const SOURCE = 'deepseek-harness'
const TERMINAL = 'deepseek-harness'
// A short, stable title that is also a substring of the Harness window title.
// That lets the overlay identify the session's window for both liveness and
// one-click jump, while the session itself stays keyed by its Harness id.
const TITLE = 'DeepSeek Harness'
const KEEP_ALIVE_MS = 60_000

const EXE_NAME = 'AI-Terminal-Pet.exe'
const PROTOCOL_NAME = 'protocol.js'
const PET_DIRECTORY = 'AI-Terminal-Pet'
const VENDOR_DIRECTORY = 'vendor'

const moduleDirectory = path.dirname(fileURLToPath(import.meta.url))

/**
 * The package root that ships vendor/. DSH_PET_TEST plus DSH_PET_PACKAGE_ROOT
 * is a test-only seam used by verify.mjs; a shipped install never sets either.
 */
function packageDirectory() {
  try {
    if (process.env.DSH_PET_TEST === '1') {
      const override = process.env.DSH_PET_PACKAGE_ROOT
      if (typeof override === 'string' && override.length > 0) return path.resolve(override)
    }
  } catch {
    // A malformed environment must never break the host.
  }
  return path.resolve(moduleDirectory, '..')
}

const packageRoot = packageDirectory()

function localAppData() {
  try {
    const value = process.env.LOCALAPPDATA
    if (typeof value === 'string' && value.length > 0) return value
  } catch {
    // Fall through to the home directory guess.
  }
  try {
    const home = os.homedir()
    if (home) return path.join(home, 'AppData', 'Local')
  } catch {
    // No home directory: nothing can be written.
  }
  return ''
}

/** LOCALAPPDATA/AI-Terminal-Pet, or an empty string when unknowable. */
function petDirectory() {
  const base = localAppData()
  return base ? path.join(base, PET_DIRECTORY) : ''
}

let protocol = null
let protocolTried = false

/** The installed module first (install.ps1 owns it), the shipped copy second. */
function protocolCandidates() {
  const candidates = []
  const installed = petDirectory()
  if (installed) candidates.push(path.join(installed, PROTOCOL_NAME))
  candidates.push(path.join(packageRoot, VENDOR_DIRECTORY, PROTOCOL_NAME))
  return candidates
}

/**
 * Resolve the official protocol module lazily, once, and never throw.
 * A candidate that fails to load (missing, unparsable, hostile) is skipped so
 * the next candidate still gets a chance.
 */
function officialProtocol() {
  if (!protocolTried) {
    protocolTried = true
    for (const candidate of protocolCandidates()) {
      try {
        if (!existsSync(candidate)) continue
        const loaded = require(candidate)
        if (loaded && typeof loaded.sendStatus === 'function') {
          protocol = loaded
          break
        }
      } catch {
        protocol = null
      }
    }
  }
  return protocol
}

/** Per-conversation state. Kept at module scope so plugin HMR does not lose it. */
const sessions = new Map()

/**
 * The session header, whichever shape the caller was handed.
 *
 * `session/event` hands over the session object itself; other events hand over
 * a wrapper with `.agent`. The header is the only place the Harness states a
 * session's cwd, lineage and creation time, so everything below reads it from
 * here instead of guessing.
 */
function headerOf(candidate) {
  return candidate?.session?.header ?? candidate?.header ?? candidate?.agent?.session?.header ?? null
}

/** What the header says about one candidate, with safe fallbacks. */
function sessionInfo(candidate) {
  const header = headerOf(candidate)
  const cwd = header && typeof header.cwd === 'string' ? header.cwd : ''
  const parent = header && header.parentSession != null ? String(header.parentSession) : ''
  const createdAt = header && Number.isFinite(header.createdAt) ? header.createdAt : 0
  return { id: sessionIdOf(candidate), cwd, parent, createdAt }
}

function metaFor(sessionId, info) {
  let item = sessions.get(sessionId)
  if (!item) {
    item = {
      state: 'idle',
      // The session's own cwd, never the host process's: a desktop Harness runs
      // from its install directory, so process.cwd() would label every session
      // with the same wrong path.
      project: info?.cwd || safeCwd(),
      title: TITLE,
      tool: '',
      reason: '',
      parentSessionId: info?.parent || '',
      startedAt: new Date(info?.createdAt || Date.now()).toISOString(),
      stateAt: 0,
      sentAt: 0,
    }
    sessions.set(sessionId, item)
  } else if (info) {
    // The header is authoritative for a session that has not stated a project
    // yet. An explicit project from an event wins for that report, which is
    // what keeps a future cwd-change extension working.
    if (!item.project && info.cwd) item.project = info.cwd
    if (info.parent) item.parentSessionId = info.parent
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
 * What makes one report different from the previous one.
 *
 * The state alone is not enough: a patch can move the project, the reason, the
 * tool or the lineage without touching the state, and the panel shows all of
 * them. De-duplicating on state alone silently dropped those updates. Identical
 * repeats are still dropped, which is what keeps a tool that fires fifty events
 * per second from flooding the pipe.
 */
function signatureOf(item) {
  return [
    item.state,
    item.reason || '',
    item.tool || '',
    item.project || '',
    item.title || '',
    item.parentSessionId || '',
  ].join('\u0000')
}

/**
 * Publish one status. The protocol module writes the durable snapshot
 * synchronously before its first await, so calling this from process.on('exit')
 * still leaves a correct closed-state snapshot even when the pipe is gone.
 */
function report(sessionId, patch, force = false, info = null) {
  const p = officialProtocol()
  if (!p || typeof p.sendStatus !== 'function') return
  const item = metaFor(sessionId, info)
  const previousState = item.state
  const previousSignature = signatureOf(item)
  Object.assign(item, patch)
  if (!p.STATES.has(item.state)) item.state = 'idle'

  const now = Date.now()
  const changed = signatureOf(item) !== previousSignature
  const stateChanged = item.state !== previousState
  const stale = now - item.sentAt >= KEEP_ALIVE_MS
  if (!force && !changed && !stale) return
  item.sentAt = now
  if (stateChanged) item.stateAt = now

  const stablePid = typeof p.stableProcessId === 'function'
    ? p.stableProcessId(SOURCE + ':' + sessionId)
    : 1

  const payload = {
    schemaVersion: p.SCHEMA_VERSION,
    source: SOURCE,
    sessionId,
    parentSessionId: item.parentSessionId || '',
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

/**
 * Report for whatever object an event handler was given, after folding in what
 * its session header says. Every handler goes through here so the project path
 * and the parent link are never taken from the host process or invented per
 * event type.
 */
function reportFor(candidate, patch, force = false) {
  const info = sessionInfo(candidate)
  return report(info.id, patch, force, info)
}

/** Copy one file into place through a temp sibling, never replacing a file. */
function copyVerified(source, destination, mode) {
  let temp = ''
  try {
    if (!existsSync(source)) return false
    if (existsSync(destination)) return false
    mkdirSync(path.dirname(destination), { recursive: true })
    temp = destination + '.' + String(process.pid) + '.' + String(Date.now()) + '.tmp'
    copyFileSync(source, temp)
    if (typeof mode === 'number') {
      try {
        chmodSync(temp, mode)
      } catch {
        // On Windows this only maps to the read-only bit; a refusal is fine.
      }
    }
    // rename() replaces on Windows, so re-check immediately before it.
    if (existsSync(destination)) {
      rmSync(temp, { force: true })
      return false
    }
    renameSync(temp, destination)
    return true
  } catch {
    try {
      if (temp) rmSync(temp, { force: true })
    } catch {
      // Cleanup is best effort.
    }
    return false
  }
}

let warned = false

function warnOnce(ctx, message) {
  if (warned) return
  warned = true
  try {
    const logger = ctx && typeof ctx === 'object' ? ctx.logger : undefined
    if (logger && typeof logger.warn === 'function') logger.warn(message)
  } catch {
    // Logging differences must never break the host.
  }
}

/** A pinned release asset { url, sha256, bytes } or null. */
function pinnedAsset(parsed, flat, group) {
  const block = parsed && typeof parsed[group] === 'object' && parsed[group] !== null ? parsed[group] : {}
  const url = typeof parsed[flat.url] === 'string' && parsed[flat.url] ? parsed[flat.url]
    : typeof block.url === 'string' && block.url ? block.url : ''
  const sha256 = typeof parsed[flat.sha256] === 'string' && parsed[flat.sha256] ? parsed[flat.sha256]
    : typeof block.sha256 === 'string' && block.sha256 ? block.sha256 : ''
  const rawBytes = parsed[flat.bytes] === undefined ? block.bytes : parsed[flat.bytes]
  const bytes = Number.isSafeInteger(Number(rawBytes)) && Number(rawBytes) > 0 ? Number(rawBytes) : 0
  if (!url || !sha256) return null
  return { url, sha256: sha256.toLowerCase(), bytes }
}

/** lib/release-pins.json, or null when the lead has not filled it in yet. */
function releasePins() {
  try {
    const file = path.join(moduleDirectory, 'release-pins.json')
    if (!existsSync(file)) return null
    const parsed = JSON.parse(readFileSync(file, 'utf8'))
    if (!parsed || typeof parsed !== 'object') return null
    const exe = pinnedAsset(parsed, { url: 'exeUrl', sha256: 'exeSha256', bytes: 'exeBytes' }, 'exe')
    const protocolPin = pinnedAsset(parsed, { url: 'protocolUrl', sha256: 'protocolSha256', bytes: 'protocolBytes' }, 'protocol')
    if (!exe || !protocolPin) return null
    return { exe, protocol: protocolPin }
  } catch {
    return null
  }
}

function temporaryDirectory() {
  const bases = []
  try {
    const temp = os.tmpdir()
    if (typeof temp === 'string' && path.isAbsolute(temp)) bases.push(temp)
  } catch {
    // tmpdir is broken in some sandboxes; try the next base.
  }
  try {
    const local = localAppData()
    if (local) bases.push(path.join(local, 'Temp'))
  } catch {
    // Ignore.
  }
  for (const base of bases) {
    try {
      mkdirSync(base, { recursive: true })
      return mkdtempSync(path.join(base, 'dsh-ai-terminal-pet-'))
    } catch {
      // Try the next base.
    }
  }
  return ''
}

async function fetchAsset(url, destination, pin) {
  if (typeof fetch !== 'function') return false
  const options = { redirect: 'follow', headers: { 'user-agent': 'dsh-ai-terminal-pet' } }
  try {
    if (typeof AbortSignal !== 'undefined' && typeof AbortSignal.timeout === 'function') {
      options.signal = AbortSignal.timeout(30000)
    }
  } catch {
    // Without timeout support the request still settles on its own.
  }
  try {
    const response = await fetch(url, options)
    if (!response || !response.ok) return false
    const buffer = await response.arrayBuffer()
    const bytes = new Uint8Array(buffer)
    if (pin.bytes && bytes.byteLength !== pin.bytes) return false
    const digest = createHash('sha256').update(bytes).digest('hex')
    if (digest !== pin.sha256) return false
    mkdirSync(path.dirname(destination), { recursive: true })
    writeFileSync(destination, bytes)
    return true
  } catch {
    return false
  }
}

/**
 * Optional, OFF unless lib/release-pins.json exists. Downloads the two pinned
 * release assets, verifies size and SHA-256, then installs them. Every failure
 * is silent except one ctx.logger.warn per process.
 */
async function downloadFallback(ctx, dir, outcome) {
  const pins = releasePins()
  if (!pins) {
    outcome.skipped = 'no-pinned-release'
    return outcome
  }
  const work = temporaryDirectory()
  if (!work) {
    outcome.skipped = 'no-temp-directory'
    return outcome
  }
  try {
    const exeTemp = path.join(work, EXE_NAME)
    const protocolTemp = path.join(work, PROTOCOL_NAME)
    const exeOk = await fetchAsset(pins.exe.url, exeTemp, pins.exe)
    const protocolOk = exeOk && await fetchAsset(pins.protocol.url, protocolTemp, pins.protocol)
    if (!exeOk || !protocolOk) {
      warnOnce(ctx, 'ai-terminal-pet: the pinned release download failed; the overlay stays uninstalled.')
      outcome.skipped = 'download-failed'
      return outcome
    }
    outcome.copiedProtocol = copyVerified(protocolTemp, path.join(dir, PROTOCOL_NAME), 0o600)
    outcome.copiedExe = copyVerified(exeTemp, path.join(dir, EXE_NAME))
    outcome.downloaded = true
    return outcome
  } catch {
    return outcome
  } finally {
    try {
      rmSync(work, { recursive: true, force: true })
    } catch {
      // Best effort.
    }
  }
}

/**
 * One-click install of the overlay binary that this package ships.
 * Runs once per process, fire-and-forget, and never overwrites a file.
 */
async function runBootstrap(ctx) {
  const outcome = { skipped: '', copiedProtocol: false, copiedExe: false, downloaded: false }
  try {
    const dir = petDirectory()
    if (!dir) {
      outcome.skipped = 'no-localappdata'
      return outcome
    }
    const targetExe = path.join(dir, EXE_NAME)
    if (existsSync(targetExe)) {
      outcome.skipped = 'installed'
      return outcome
    }
    const vendorExe = path.join(packageRoot, VENDOR_DIRECTORY, EXE_NAME)
    if (existsSync(vendorExe)) {
      mkdirSync(dir, { recursive: true })
      outcome.copiedProtocol = copyVerified(
        path.join(packageRoot, VENDOR_DIRECTORY, PROTOCOL_NAME),
        path.join(dir, PROTOCOL_NAME),
        0o600,
      )
      outcome.copiedExe = copyVerified(vendorExe, targetExe)
      return outcome
    }
    return await downloadFallback(ctx, dir, outcome)
  } catch {
    return outcome
  }
}

let bootstrapPromise = null

function bootstrap(ctx) {
  if (bootstrapPromise) return bootstrapPromise
  try {
    if (process.env.DSH_PET_NO_BOOTSTRAP === '1') {
      bootstrapPromise = Promise.resolve({ skipped: 'disabled', copiedProtocol: false, copiedExe: false, downloaded: false })
      return bootstrapPromise
    }
  } catch {
    // Ignore a broken environment.
  }
  bootstrapPromise = runBootstrap(ctx).catch(() => ({ skipped: 'failed', copiedProtocol: false, copiedExe: false, downloaded: false }))
  return bootstrapPromise
}

/**
 * Map one durable session event onto a state patch.
 *
 * The authoritative list lives in @deepseek-ai/dsh-session's SessionEventMap
 * (`turn/start`, `turn/end`, `step/start`, `step/end`, `user/message`,
 * `assistant/chunk`, `assistant/message`, `tool/call`, `tool/result`,
 * `todo/write`, `request/header`, `request/context`, `session/end-seed`), which
 * this switch covers for every event that says something about activity.
 *
 * `working-directory/change` and `session/title` are NOT in that map: the first
 * does not exist anywhere in the host, the second is a client-side event. Both
 * branches are kept as cheap forward-compatibility - if a future version emits
 * them, they map correctly - but the project path and the lineage come from
 * session.header (see reportFor), which is the only place the host states them.
 * `approval/asked` / `approval/decided` likewise come from the approval service
 * rather than the session log; the approval/request waterfall listener in
 * apply() is the path that actually fires.
 */
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
  try {
    if (config && config.enabled === false) return () => {}

    // Fire-and-forget: this only makes sure the overlay binary exists.
    void bootstrap(ctx)

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
      const patch = eventPatch(event?.type, event?.data)
      if (patch) reportFor(session, patch)
    })

    on('agent/inbox/claimed', (payload) => {
      reportFor(payload?.agent, { state: 'working', reason: '', tool: '' })
    })

    on('agent/status', (payload) => {
      const status = String(payload?.status ?? '').toLowerCase()
      if (/run|work|busy|think|stream/.test(status)) {
        reportFor(payload?.agent, { state: 'working' })
      } else if (/idle|done|wait|stop/.test(status)) {
        reportFor(payload?.agent, { state: 'idle' })
      }
    })

    on('agent/error', (payload) => {
      reportFor(payload?.agent, {
        state: 'error',
        reason: short(payload?.error?.message ?? payload?.error, 120),
      })
    })

    // approval/request is a waterfall listener: it must always call next(),
    // otherwise other plugins' approval handling is silently dropped.
    on('approval/request', (request, next) => {
      reportFor(request?.agent, {
        state: 'waiting_permission',
        reason: 'permission required',
        tool: short(request?.toolName, 120),
      })
      if (typeof next === 'function') return next()
      return undefined
    })

    // Only a real process exit means all conversations are over. Plugin HMR
    // calls the dispose function below instead, and must NOT clear the panel.
    const onExit = quiet(() => {
      for (const sessionId of sessions.keys()) {
        report(sessionId, { state: 'closed', reason: 'harness exited' }, true)
      }
    })
    try {
      process.on('exit', onExit)
    } catch {
      // A host without process events is fine; reports simply stop.
    }

    return () => {
      try {
        process.off('exit', onExit)
      } catch {
        // Ignore.
      }
      for (const disposer of off) {
        try {
          disposer()
        } catch {
          // Ignore.
        }
      }
      off.length = 0
    }
  } catch {
    return () => {}
  }
}

/** Test-only seam used by verify.mjs; not part of the plugin surface. */
export const __testing = { bootstrap, officialProtocol, packageRoot, petDirectory }
