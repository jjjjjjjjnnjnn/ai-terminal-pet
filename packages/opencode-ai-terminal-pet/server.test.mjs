// What this file proves, and why it is a *packaged* test
// ------------------------------------------------------
// The generator (scripts/sync.mjs) makes server.js byte-identical to
// reporters/opencode.js apart from one specifier, and the event catalog is
// already pinned by reporters/opencode-events.test.mjs. Neither of those would
// notice the failure this file exists for: that the generated entry point, once
// loaded from inside a "type": "module" package, still resolves the CommonJS
// core and still puts a real payload on the endpoint.
//
// That is the whole packaged shape - ESM entry, require() of a .cjs file - and
// it is exactly the thing that broke every time the reporter was copied by hand.
//
// The endpoint is a temporary one, not the user's: pipeScopeToken() derives the
// endpoint from the home directory *at module load*, so the environment is
// redirected before protocol.cjs is imported. The test therefore cannot reach a
// running overlay, and asserts that it did not.
//
//   node --test packages/opencode-ai-terminal-pet/server.test.mjs

import assert from 'node:assert/strict'
import { existsSync, mkdirSync, mkdtempSync, readFileSync, rmSync } from 'node:fs'
import net from 'node:net'
import os from 'node:os'
import path from 'node:path'
import test from 'node:test'

const OVERRIDES = ['USERPROFILE', 'HOME', 'LOCALAPPDATA', 'XDG_DATA_HOME', 'XDG_RUNTIME_DIR', 'TMPDIR']

function redirectHome(t) {
  const root = mkdtempSync(path.join(os.tmpdir(), 'atp-opencode-'))
  // The endpoint is a unix socket on POSIX, and sun_path is 104 bytes (108 on
  // Linux, 104 on macOS) including the terminator. os.tmpdir() on macOS is a
  // per-user path under /var/folders/..., and three directory levels on top of
  // it is enough to cross that line - CI found this the first time the test ran
  // on a Mac (listen EINVAL). So the runtime directory gets a short root of its
  // own, and the length is asserted below rather than trusted.
  const shortRoot =
    process.platform === 'win32'
      ? root
      : mkdtempSync(path.join(existsSync('/tmp') ? '/tmp' : os.tmpdir(), 'atp-'))
  const directories = {
    home: path.join(root, 'home'),
    appData: path.join(root, 'appdata'),
    runtime: path.join(shortRoot, 'run')
  }
  for (const directory of Object.values(directories)) mkdirSync(directory, { recursive: true })
  const saved = {}
  const values = {
    USERPROFILE: directories.home,
    HOME: directories.home,
    LOCALAPPDATA: directories.appData,
    XDG_DATA_HOME: path.join(root, 'xdg'),
    XDG_RUNTIME_DIR: directories.runtime,
    TMPDIR: directories.runtime
  }
  for (const name of OVERRIDES) {
    saved[name] = process.env[name]
    process.env[name] = values[name]
  }
  t.after(() => {
    for (const name of OVERRIDES) {
      if (saved[name] === undefined) delete process.env[name]
      else process.env[name] = saved[name]
    }
    rmSync(root, { recursive: true, force: true })
    if (shortRoot !== root) rmSync(shortRoot, { recursive: true, force: true })
  })
  return { root, ...directories }
}

async function waitFor(predicate, timeoutMs, describe) {
  const deadline = Date.now() + timeoutMs
  while (Date.now() < deadline) {
    if (predicate()) return
    await new Promise((resolve) => setTimeout(resolve, 20))
  }
  throw new Error(`timed out after ${timeoutMs} ms waiting for ${describe}`)
}

test('the packaged entry point reports real payloads over a real endpoint', async (t) => {
  const { runtime } = redirectHome(t)

  // Imported after the redirect: protocol.cjs computes PIPE_PATH once, at load.
  const protocol = (await import('./protocol.cjs')).default
  const endpoint = protocol.PIPE_PATH

  // Safety, and the reason this test is worth having on three platforms: the
  // endpoint must belong to this temporary home, never the developer's.
  if (process.platform === 'win32') {
    assert.match(endpoint, /^\\\\\.\\pipe\\ai-terminal-pet-v1-[0-9a-f]{8}$/)
  } else {
    assert.ok(
      endpoint.startsWith(path.join(runtime, 'ai-terminal-pet-v1-')),
      `endpoint ${endpoint} escaped the temporary runtime directory`
    )
    // sun_path is 104 bytes on macOS and 108 on Linux, terminator included; a
    // longer path fails at listen() with EINVAL, which is how CI caught this.
    assert.ok(endpoint.length < 104, `endpoint path is ${endpoint.length} bytes, too long for sun_path`)
  }

  const lines = []
  const server = net.createServer((socket) => {
    // Answer as soon as the line is complete, not on 'end': on Windows a named
    // pipe has no half-close, so the client's socket.end() disconnects and a
    // reply written afterwards is EPIPE. The overlay's Go server reads until
    // EOF because it owns both ends; here the newline is the natural boundary.
    let buffer = ''
    socket.on('error', () => {}) // a disconnected client is the next delivery's problem
    socket.on('data', (chunk) => {
      buffer += chunk.toString('utf8')
      if (!buffer.includes('\n')) return
      lines.push(buffer)
      buffer = ''
      try {
        socket.end('OK') // the exact acknowledgement the overlay's server sends
      } catch {
        /* the test's own write raced a disconnect; the line is already counted */
      }
    })
    socket.on('end', () => {})
  })
  await new Promise((resolve, reject) => {
    server.once('error', reject)
    server.listen(endpoint, resolve)
  })
  t.after(() => new Promise((resolve) => server.close(resolve)))

  const { plugin } = await import('./server.js')
  assert.equal(plugin.id, 'ai-terminal-pet', 'the id OpenCode shows in its plugin list')
  assert.equal(typeof plugin.setup, 'function')

  const events = [
    {
      type: 'session.execution.started',
      data: { sessionID: 'ses_pkg_1', title: 'Refactor the parser', location: { directory: '/work/pet' } }
    },
    // The same state twice: report() must drop the repeat, because every
    // identical line repaints the panel for no new information.
    { type: 'session.execution.started', data: { sessionID: 'ses_pkg_1' } },
    { type: 'permission.asked', data: { sessionID: 'ses_pkg_1', action: 'bash' } },
    { type: 'permission.asked', data: { sessionID: 'ses_pkg_1', action: 'bash' } },
    { type: 'session.archived', data: { sessionID: 'ses_pkg_1' } }
  ]
  const context = {
    location: { directory: '/work/pet' },
    event: {
      subscribe() {
        return (async function* stream() {
          for (const event of events) yield event
        })()
      }
    }
  }

  const dispose = plugin.setup(context)
  t.after(() => dispose())

  await waitFor(() => lines.length >= 3, 8000, `3 deliveries on ${endpoint}`)

  // Three lines out of five events, and they stay three: a line the endpoint
  // refused would be retried within 100 ms, so a quiet window is what proves
  // the acknowledgement was accepted rather than merely attempted.
  await new Promise((resolve) => setTimeout(resolve, 300))
  assert.equal(lines.length, 3, 'repeats are dropped and every delivered line was acknowledged')

  const payloads = lines.map((line) => JSON.parse(line.trim()))
  assert.equal(payloads.length, 3, 'one line per state change, and no repeats')
  for (const payload of payloads) {
    assert.equal(payload.schemaVersion, 'ai-terminal-pet.v1')
    assert.equal(payload.source, 'opencode')
    assert.equal(payload.terminal, 'opencode')
    assert.equal(payload.sessionId, 'ses_pkg_1')
    assert.equal(payload.title, 'Refactor the parser')
    assert.equal(payload.project, '/work/pet')
  }
  assert.deepEqual(payloads.map((payload) => payload.state).sort(), ['closed', 'waiting_permission', 'working'])

  const ask = payloads.find((payload) => payload.state === 'waiting_permission')
  assert.equal(ask.reason, 'permission_required')
  assert.equal(ask.tool, 'bash')

  // The snapshot is the durable half of the contract: the panel reads it even
  // when the pipe delivery is missed, and `closed` is how a row ends.
  const snapshot = protocol.snapshotPath({ source: 'opencode', sessionId: 'ses_pkg_1' })
  const stored = JSON.parse(readFileSync(snapshot, 'utf8'))
  assert.equal(stored.state, 'closed')
  assert.equal(stored.reason, 'session_archived')
})
