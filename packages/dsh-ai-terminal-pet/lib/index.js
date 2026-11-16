/**
 * dsh-ai-terminal-pet - a DeepSeek Harness Cordis bundle that reports session
 * state to the AI Terminal Pet desktop overlay.
 *
 * The plugin is a field adapter only. It never computes the state set, the
 * snapshot path or the redaction rules: those come from the official protocol
 * module, resolved at first use from
 *   <data directory>/protocol.js   (placed there by the installer, or by this
 *                                   plugin's own bootstrap from vendor/)
 * and, when that file is missing, from the copy shipped in this package
 * (vendor/protocol.js).
 *
 * The data directory itself is the one path rule it cannot borrow from that
 * module - protocol.js lives inside the directory it would have to compute - so
 * this file carries its own reading of the same contract (Windows
 * %LOCALAPPDATA%\AI-Terminal-Pet, macOS Application Support, XDG on Linux) and
 * verify.mjs pins it to protocol.js's answer on every OS in CI.
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
  readFileSync,
  renameSync,
  rmSync,
  writeFileSync,
} from 'node:fs'
import { createRequire } from 'node:module'
import os from 'node:os'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { gunzipSync } from 'node:zlib'

const require = createRequire(import.meta.url)

const SOURCE = 'deepseek-harness'
const TERMINAL = 'deepseek-harness'
// A short, stable title that is also a substring of the Harness window title.
// That lets the overlay identify the session's window for both liveness and
// one-click jump, while the session itself stays keyed by its Harness id.
const TITLE = 'DeepSeek Harness'
const KEEP_ALIVE_MS = 60_000

const PROTOCOL_NAME = 'protocol.js'
const PET_DIRECTORY = 'AI-Terminal-Pet'
// Linux follows XDG, where the convention is lower-case; internal/home makes the
// same split on the Go side ("a stray capitalised directory in ~/.local/share is
// not what any other Linux application does").
const PET_DIRECTORY_XDG = 'ai-terminal-pet'
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

/**
 * Which platform and architecture this process is, with a test-only seam.
 *
 * verify.mjs has to exercise the Linux and macOS bootstrap from whichever
 * machine runs it, and the three CI jobs that run it are themselves three
 * different machines. The seam makes both possible; a shipped install never
 * sets either variable.
 */
function hostPlatform() {
  try {
    if (process.env.DSH_PET_TEST === '1' && process.env.DSH_PET_PLATFORM) return process.env.DSH_PET_PLATFORM
  } catch {
    // A broken environment falls back to the real platform.
  }
  return process.platform
}

function hostArch() {
  try {
    if (process.env.DSH_PET_TEST === '1' && process.env.DSH_PET_ARCH) return process.env.DSH_PET_ARCH
  } catch {
    // Fall through.
  }
  return process.arch
}

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

function homeDirectory() {
  try {
    const home = os.homedir()
    if (typeof home === 'string' && home.length > 0) return home
  } catch {
    // Fall through.
  }
  return ''
}

/**
 * This project's data directory, on all three platforms.
 *
 * The contract is owned by reporters/protocol.js (dataDirectory) and mirrored by
 * internal/home in Go. This is the plugin's reading of it - it needs the answer
 * before it can load protocol.js, which is what lives at that path - and
 * verify.mjs pins the two JavaScript answers together on every OS.
 */
function petDirectory() {
  const platform = hostPlatform()
  if (platform === 'win32') {
    const base = localAppData()
    return base ? path.join(base, PET_DIRECTORY) : ''
  }
  if (platform === 'darwin') {
    const home = homeDirectory()
    return home ? path.join(home, 'Library', 'Application Support', PET_DIRECTORY) : ''
  }
  const xdg = process.env.XDG_DATA_HOME
  if (typeof xdg === 'string' && path.isAbsolute(xdg)) return path.join(xdg, PET_DIRECTORY_XDG)
  const home = homeDirectory()
  return home ? path.join(home, '.local', 'share', PET_DIRECTORY_XDG) : ''
}

/** The overlay binary's file name. The Go side strips .exe on POSIX too. */
function overlayBinaryName() {
  return hostPlatform() === 'win32' ? 'AI-Terminal-Pet.exe' : 'AI-Terminal-Pet'
}

/** os/arch as the release assets spell it: linux-amd64, darwin-arm64, ... */
function platformKey() {
  const arch = hostArch()
  const goarch = { x64: 'amd64', arm64: 'arm64', ia32: '386', arm: 'arm' }[arch] || arch
  return hostPlatform() + '-' + goarch
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

/** A pinned release archive { url, sha256, bytes, member } or null. */
function pinnedAsset(entry) {
  if (!entry || typeof entry !== 'object') return null
  const url = typeof entry.url === 'string' ? entry.url : ''
  const sha256 = typeof entry.sha256 === 'string' ? entry.sha256.toLowerCase() : ''
  const member = typeof entry.member === 'string' ? entry.member : ''
  const bytes = Number.isSafeInteger(Number(entry.bytes)) && Number(entry.bytes) > 0 ? Number(entry.bytes) : 0
  if (!url || !/^[0-9a-f]{64}$/.test(sha256) || !member) return null
  return { url, sha256, bytes, member }
}

/**
 * lib/release-pins.json's `platforms`, or null when it is absent.
 *
 * Schema 2, keyed by `platformKey()`. Windows never reads it: its overlay ships
 * in vendor/ and installs offline. Linux and macOS have no bundled binary - one
 * tarball serves every platform, and four ~2 MB binaries inside it would make
 * every Windows install pay for three platforms it cannot run - so they fetch the
 * pinned archive once, on first use, and verify it against this file.
 */
function releasePins() {
  try {
    const file = path.join(moduleDirectory, 'release-pins.json')
    if (!existsSync(file)) return null
    const parsed = JSON.parse(readFileSync(file, 'utf8'))
    if (!parsed || typeof parsed !== 'object') return null
    if (!parsed.platforms || typeof parsed.platforms !== 'object') return null
    return parsed.platforms
  } catch {
    return null
  }
}

/** The pinned archive's bytes, or null when the fetch, size or digest fails. */
async function fetchPinnedBytes(pin) {
  if (typeof fetch !== 'function') return null
  const options = { redirect: 'follow', headers: { 'user-agent': 'dsh-ai-terminal-pet' } }
  try {
    if (typeof AbortSignal !== 'undefined' && typeof AbortSignal.timeout === 'function') {
      options.signal = AbortSignal.timeout(30000)
    }
  } catch {
    // Without timeout support the request still settles on its own.
  }
  try {
    const response = await fetch(pin.url, options)
    if (!response || !response.ok) return null
    // A hostile redirect must not be able to make the host allocate without
    // bound: the pin's size is the ceiling, and it is small.
    const declared = Number(response.headers?.get?.('content-length'))
    if (pin.bytes && Number.isFinite(declared) && declared > pin.bytes) return null
    const bytes = Buffer.from(await response.arrayBuffer())
    if (pin.bytes && bytes.length !== pin.bytes) return null
    if (createHash('sha256').update(bytes).digest('hex') !== pin.sha256) return null
    return bytes
  } catch {
    return null
  }
}

/**
 * Reads one file out of a gzipped USTAR archive - the shape
 * scripts/package-unix.mjs writes, and the shape the release ships.
 *
 * It is not a general tar reader: no pax, no GNU long names, no symlinks. It
 * walks the header sequence, compares the full name (ustar prefix included) and
 * returns that member's bytes. A member it cannot parse ends the walk with null
 * rather than guessing.
 */
function readTarMember(archive, member) {
  let raw
  try {
    raw = gunzipSync(archive)
  } catch {
    return null
  }
  let offset = 0
  while (offset + 512 <= raw.length) {
    const header = raw.subarray(offset, offset + 512)
    if (header.every((byte) => byte === 0)) return null
    const name = header.subarray(0, 100).toString('utf8').replace(/\0.*$/, '')
    const magic = header.subarray(257, 262).toString('ascii')
    const prefix = magic.startsWith('ustar') ? header.subarray(345, 345 + 155).toString('utf8').replace(/\0.*$/, '') : ''
    const sizeText = header.subarray(124, 136).toString('ascii').replace(/\0.*$/, '').trim()
    const size = parseInt(sizeText || '0', 8)
    const type = header[156] === 0 ? '0' : String.fromCharCode(header[156])
    const full = prefix ? prefix + '/' + name : name
    if (!Number.isFinite(size) || size < 0) return null
    const dataStart = offset + 512
    const dataEnd = dataStart + size
    if (dataEnd > raw.length) return null
    if (type === '0' && full === member) return raw.subarray(dataStart, dataEnd)
    offset = dataStart + Math.ceil(size / 512) * 512
  }
  return null
}

/**
 * Writes an executable atomically. The temporary name ends in .tmp so a crash
 * leaves something verify.mjs can find, and renameSync never clobbers a file a
 * user (or a newer install) put there first.
 */
function writeExecutable(destination, bytes) {
  const temp = destination + '.' + String(process.pid) + '.' + String(Date.now()) + '.tmp'
  try {
    mkdirSync(path.dirname(destination), { recursive: true })
    writeFileSync(temp, bytes, { mode: 0o755 })
    try {
      chmodSync(temp, 0o755)
    } catch {
      // Windows maps this to the read-only bit; a refusal is fine.
    }
    if (existsSync(destination)) {
      rmSync(temp, { force: true })
      return false
    }
    renameSync(temp, destination)
    return true
  } catch {
    try {
      rmSync(temp, { force: true })
    } catch {
      // Best effort.
    }
    return false
  }
}

/**
 * Installs the overlay for this platform from the pinned release archive.
 *
 * One { ctx.logger.warn } per process when it cannot: the user can still install
 * the overlay from the release archive (scripts/install.sh), and a plugin that
 * shouts on every load is worse than one that says it once.
 */
async function installFromPin(ctx, dir, outcome) {
  const platforms = releasePins()
  const key = platformKey()
  const pin = platforms ? pinnedAsset(platforms[key]) : null
  if (!pin) {
    outcome.skipped = 'no-pinned-overlay'
    warnOnce(ctx, 'ai-terminal-pet: no pinned overlay for ' + key + ' - install it from the release archive (scripts/install.sh) and the panel will appear.')
    return outcome
  }
  const archive = await fetchPinnedBytes(pin)
  if (!archive) {
    outcome.skipped = 'download-failed'
    warnOnce(ctx, 'ai-terminal-pet: the pinned overlay download failed; the panel stays uninstalled (reports still go to the pipe when it runs).')
    return outcome
  }
  const binary = readTarMember(archive, pin.member)
  if (!binary) {
    outcome.skipped = 'archive-member-missing'
    warnOnce(ctx, 'ai-terminal-pet: the pinned archive does not contain ' + pin.member + '; nothing was installed.')
    return outcome
  }
  outcome.copiedExe = writeExecutable(path.join(dir, overlayBinaryName()), binary)
  outcome.downloaded = true
  return outcome
}

/**
 * One-click install of the overlay binary for this platform.
 *
 * Runs once per process, fire-and-forget, and never overwrites a file. Windows
 * copies the vendored binary (offline); Linux and macOS fetch the pinned release
 * archive, because the package deliberately ships no binary for them.
 */
async function runBootstrap(ctx) {
  const outcome = { skipped: '', copiedProtocol: false, copiedExe: false, downloaded: false }
  try {
    const dir = petDirectory()
    if (!dir) {
      outcome.skipped = 'no-data-directory'
      return outcome
    }
    const targetExe = path.join(dir, overlayBinaryName())
    if (existsSync(targetExe)) {
      outcome.skipped = 'installed'
      return outcome
    }
    mkdirSync(dir, { recursive: true })
    // protocol.js ships with the package on every platform, and the reporters
    // read the installed copy first, so it is always the one that lands.
    outcome.copiedProtocol = copyVerified(
      path.join(packageRoot, VENDOR_DIRECTORY, PROTOCOL_NAME),
      path.join(dir, PROTOCOL_NAME),
      0o600,
    )
    const vendorExe = path.join(packageRoot, VENDOR_DIRECTORY, overlayBinaryName())
    if (existsSync(vendorExe)) {
      outcome.copiedExe = copyVerified(vendorExe, targetExe, 0o755)
      return outcome
    }
    return await installFromPin(ctx, dir, outcome)
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
export const __testing = {
  bootstrap,
  officialProtocol,
  packageRoot,
  petDirectory,
  overlayBinaryName,
  platformKey,
  readTarMember,
}
