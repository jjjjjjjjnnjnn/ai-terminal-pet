/**
 * Offline verification for the DeepSeek Harness bridge.
 *
 *   node integrations/deepseek-harness/verify.mjs
 *
 * It substitutes a stub for the official protocol module (via a temporary
 * LOCALAPPDATA), drives the plugin with a fake Cordis context, and asserts the
 * emitted ai-terminal-pet.v1 payloads. No pipe, no overlay, no real Harness.
 */
import { createRequire } from 'node:module'
import fs from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import { pathToFileURL } from 'node:url'

const require = createRequire(import.meta.url)
const root = path.resolve(import.meta.dirname, '..', '..')
const bundle = path.join(root, 'integrations', 'deepseek-harness', 'ai-terminal-pet-bridge')
const stubDir = fs.mkdtempSync(path.join(os.tmpdir(), 'dsh-pet-verify-'))
const petDir = path.join(stubDir, 'AI-Terminal-Pet')
fs.mkdirSync(petDir, { recursive: true })
fs.writeFileSync(path.join(petDir, 'protocol.js'), `
const calls = []
module.exports = {
  calls,
  SCHEMA_VERSION: 'ai-terminal-pet.v1',
  STATES: new Set(['waiting_permission','error','working','tool_use','thinking','idle','closed']),
  stableProcessId: () => 1701354615,
  sendStatus: async (payload) => { calls.push(payload); return true },
}
`)
// The plugin resolves PROTOCOL_FILE at import time, so set this first.
process.env.LOCALAPPDATA = stubDir
const stub = require(path.join(petDir, 'protocol.js'))
const { apply } = await import(pathToFileURL(path.join(bundle, 'index.js')).href)

const handlers = new Map()
const ctx = {
  on(name, handler) {
    handlers.set(name, handler)
    return () => handlers.delete(name)
  },
}
const cleanup = apply(ctx, { enabled: true })
const fire = (name, ...args) => {
  const handler = handlers.get(name)
  if (!handler) throw new Error(`plugin did not subscribe to ${name}`)
  return handler(...args)
}

const session = { header: { id: 'ses_abc' }, id: 'ses_abc' }
const agent = { session: { header: { id: 'ses_abc' }, id: 'ses_abc' } }

fire('session/event', session, { type: 'turn/start', data: {} })
fire('session/event', session, { type: 'tool/call', data: { name: 'Bash' } })
fire('session/event', session, { type: 'approval/asked', data: { toolName: 'Bash' } })
fire('session/event', session, { type: 'turn/end', data: {} })
fire('agent/error', { agent, error: { message: 'boom' } })
fire('session/event', session, { type: 'working-directory/change', data: { cwd: 'C:\\work\\demo' } })
const nextResult = fire('approval/request', { agent, toolName: 'Bash' }, () => 'next-called')

// A real process exit is the only event that ends every conversation.
process.emit('exit', 0)
const beforeCleanup = stub.calls.length
cleanup()
process.emit('exit', 0)
const afterCleanup = stub.calls.length

const states = stub.calls.map((item) => item.state)
const expected = ['working', 'tool_use', 'waiting_permission', 'idle', 'error', 'waiting_permission', 'closed']
let failed = false
function check(condition, message) {
  if (!condition) {
    failed = true
    console.error('FAIL:', message)
  } else {
    console.log('ok  :', message)
  }
}

check(JSON.stringify(states) === JSON.stringify(expected), `state sequence ${states.join(' -> ')}`)
check(nextResult === 'next-called', 'approval/request preserves the waterfall chain')
check(stub.calls.every((item) => item.source === 'deepseek-harness' && item.terminal === 'deepseek-harness'), 'source/terminal are stable')
check(stub.calls.every((item) => item.sessionId === 'ses_abc'), 'one stable session id, no event-id ghosts')
check(stub.calls.every((item) => item.pid === 1701354615 && item.parentPid === process.pid), 'synthetic pid + real host parentPid')
check(stub.calls.some((item) => item.tool === 'Bash'), 'tool name is forwarded')
check(stub.calls.some((item) => item.project === 'C:\\work\\demo'), 'working directory is forwarded')
check(stub.calls.at(-1).state === 'closed', 'process exit writes a closed report')
check(handlers.size === 0, 'dispose removes every listener')
check(afterCleanup === beforeCleanup, 'dispose does NOT end conversations (HMR is not an exit)')
check(stub.calls.every((item) => item.updatedAt.endsWith('Z')), 'timestamps are UTC')

// The real protocol module must repair clock skew instead of letting the
// server reject and delete the whole snapshot. A Harness plugin may pass a
// session start time taken from a different clock.
const official = require(path.join(root, 'reporters', 'protocol.js'))
const future = new Date(Date.now() + 30 * 60 * 1000).toISOString()
const skewed = official.normalizePayload({
  source: 'deepseek-harness', sessionId: 'clock-skew',
  state: 'working', startedAt: future, updatedAt: new Date().toISOString(),
})
const reparsed = official.normalizePayload(skewed)
check(Date.parse(skewed.startedAt) <= Date.parse(skewed.updatedAt), 'clock skew cannot make startedAt follow updatedAt')
check(Date.parse(skewed.updatedAt) <= Date.now() + 5 * 60 * 1000, 'timestamps are clamped below the server limit')
check(JSON.stringify(skewed) === JSON.stringify(reparsed), 'normalization is idempotent (sendStatus calls it twice)')

fs.rmSync(stubDir, { recursive: true, force: true })
console.log(failed ? '\nVERIFY FAILED' : '\nVERIFY PASSED')
process.exitCode = failed ? 1 : 0
