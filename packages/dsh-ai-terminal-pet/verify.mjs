/**
 * Offline verification for dsh-ai-terminal-pet.
 *
 *   node verify.mjs        (or: npm run verify)
 *
 * No network, no named pipe, no overlay. A temporary LOCALAPPDATA, a temporary
 * package root and a stub protocol module stand in for the real ones, and a
 * fake Cordis context drives the plugin.
 *
 * Covered:
 *   - the full state mapping (turn/start, tool/call, approval/asked, turn/end,
 *     error, approval/request, closed)
 *   - the approval/request waterfall still calls next()
 *   - dispose removes every listener and does NOT emit closed (HMR is not exit)
 *   - a real process exit emits closed
 *   - bootstrap copies protocol.js then the exe into the fake LOCALAPPDATA,
 *     only when missing, and never overwrites an existing file
 *   - a protocol module that throws on import leaves the plugin completely silent
 *   - vendor/manifest.json still describes the two files committed under vendor/
 *   - docs/market-index.json still pins the packaged version to a commit
 */
import { createHash } from 'node:crypto'
import { existsSync, mkdirSync, mkdtempSync, readFileSync, readdirSync, rmSync, statSync, writeFileSync } from 'node:fs'
import { createServer } from 'node:http'
import { createRequire } from 'node:module'
import os from 'node:os'
import path from 'node:path'
import { fileURLToPath, pathToFileURL } from 'node:url'
import { gzipSync } from 'node:zlib'

const require = createRequire(import.meta.url)
const packageDir = path.dirname(fileURLToPath(import.meta.url))
const source = path.join(packageDir, 'lib', 'index.js')
const root = mkdtempSync(path.join(os.tmpdir(), 'dsh-ai-terminal-pet-verify-'))

const STUB_PROTOCOL = [
  'const calls = []',
  'module.exports = {',
  '  calls,',
  '  SCHEMA_VERSION: "ai-terminal-pet.v1",',
  '  STATES: new Set(["waiting_permission","error","working","tool_use","thinking","idle","closed"]),',
  '  stableProcessId: () => 1701354615,',
  '  sendStatus: async (payload) => { calls.push(payload); return true },',
  '}',
  '',
].join('\n')

const VENDOR_PROTOCOL = [
  '// stub shipped in vendor/protocol.js',
  'module.exports = {',
  '  SCHEMA_VERSION: "ai-terminal-pet.v1",',
  '  STATES: new Set(["idle", "working"]),',
  '  stableProcessId: () => 1,',
  '  sendStatus: async () => true,',
  '}',
  '',
].join('\n')

const VENDOR_EXE = 'MZ fake overlay payload for the offline verify run\n'

let failed = false
function check(condition, message) {
  if (!condition) {
    failed = true
    console.log('fail: ' + message)
  } else {
    console.log('ok  : ' + message)
  }
}

function sha256(file) {
  return createHash('sha256').update(readFileSync(file)).digest('hex')
}

let generation = 0
/** Import a fresh module instance so module-level caches start clean. */
async function load(entry) {
  generation += 1
  return await import(pathToFileURL(entry).href + '?gen=' + String(generation))
}

function fakeContext(handlers) {
  return {
    on(name, handler) {
      handlers.set(name, handler)
      return () => handlers.delete(name)
    },
    logger: { warn: () => {} },
  }
}

/** Case 1: state mapping, waterfall, dispose and exit. */
async function caseMapping() {
  const local = path.join(root, 'local-mapping')
  const petDir = path.join(local, 'AI-Terminal-Pet')
  mkdirSync(petDir, { recursive: true })
  writeFileSync(path.join(petDir, 'protocol.js'), STUB_PROTOCOL, 'utf8')

  const fakeRoot = path.join(root, 'pkg-mapping')
  mkdirSync(path.join(fakeRoot, 'lib'), { recursive: true })
  writeFileSync(path.join(fakeRoot, 'lib', 'index.js'), readFileSync(source, 'utf8'), 'utf8')

  process.env.LOCALAPPDATA = local
  process.env.DSH_PET_TEST = '1'
  process.env.DSH_PET_PACKAGE_ROOT = fakeRoot
  process.env.DSH_PET_NO_BOOTSTRAP = '1'
  // The stub is filed where Windows would look for it, so the seam pins the
  // platform here too: without it this case silently reports nothing on Linux
  // and macOS (a real finding from the first Linux CI run).
  process.env.DSH_PET_PLATFORM = 'win32'
  process.env.DSH_PET_ARCH = 'x64'

  const stub = require(path.join(petDir, 'protocol.js'))
  const plugin = await load(path.join(fakeRoot, 'lib', 'index.js'))

  const handlers = new Map()
  const ctx = fakeContext(handlers)
  const cleanup = plugin.apply(ctx, { enabled: true })

  const fire = (name, ...args) => {
    const handler = handlers.get(name)
    if (!handler) throw new Error('plugin did not subscribe to ' + name)
    return handler(...args)
  }

  const session = {
    header: { id: 'ses_abc', cwd: 'C:\\work\\demo', parentSession: 'ses_parent', createdAt: 1763000000000 },
    id: 'ses_abc',
  }
  const agent = { session: { header: { id: 'ses_abc', cwd: 'C:\\work\\demo', parentSession: 'ses_parent', createdAt: 1763000000000 }, id: 'ses_abc' } }

  fire('session/event', session, { type: 'turn/start', data: {} })
  fire('session/event', session, { type: 'tool/call', data: { name: 'Bash' } })
  fire('session/event', session, { type: 'approval/asked', data: { toolName: 'Bash' } })
  fire('session/event', session, { type: 'turn/end', data: {} })
  fire('agent/error', { agent, error: { message: 'boom' } })
  fire('session/event', session, { type: 'working-directory/change', data: { cwd: 'C:\\work\\demo' } })
  const nextResult = fire('approval/request', { agent, toolName: 'Bash' }, () => 'next-called')

  process.emit('exit', 0)
  const beforeCleanup = stub.calls.length
  cleanup()
  process.emit('exit', 0)
  const afterCleanup = stub.calls.length

  const states = stub.calls.map((item) => item.state)
  // Seven reports for seven state changes. working-directory/change adds none:
  // it names the same cwd the session header already carries, so the de-dupe
  // signature does not move. A project that *does* differ would produce its own
  // report - that is what the signature change is for.
  const expected = ['working', 'tool_use', 'waiting_permission', 'idle', 'error', 'waiting_permission', 'closed']

  check(JSON.stringify(states) === JSON.stringify(expected), 'state sequence ' + states.join(' -> '))
  check(nextResult === 'next-called', 'approval/request preserves the waterfall chain')
  check(stub.calls.length > 0 && stub.calls.every((item) => item.source === 'deepseek-harness' && item.terminal === 'deepseek-harness'), 'source/terminal are stable')
  check(stub.calls.every((item) => item.sessionId === 'ses_abc'), 'one stable session id, no event-id ghosts')
  check(stub.calls.every((item) => item.pid === 1701354615 && item.parentPid === process.pid), 'synthetic pid + real host parentPid')
  check(stub.calls.some((item) => item.tool === 'Bash'), 'tool name is forwarded')
  // The project is the session header's cwd, not the host process's: a desktop
  // Harness runs from its install directory, so process.cwd() would label every
  // session with the same wrong path.
  check(
    stub.calls.every((item) => item.project === 'C:\\work\\demo'),
    'every report carries the session header cwd (' + JSON.stringify([...new Set(stub.calls.map((item) => item.project))]) + ')',
  )
  check(
    stub.calls.every((item) => item.parentSessionId === 'ses_parent'),
    'the header lineage becomes parentSessionId',
  )
  check(
    stub.calls.every((item) => Date.parse(item.startedAt) === 1763000000000),
    'startedAt is the header createdAt, not the first event seen (' + stub.calls[0]?.startedAt + ')',
  )
  check(stub.calls.at(-1)?.state === 'closed', 'process exit writes a closed report')
  check(handlers.size === 0, 'dispose removes every listener')
  check(afterCleanup === beforeCleanup, 'dispose does NOT end conversations (HMR is not an exit)')
  check(stub.calls.every((item) => item.updatedAt.endsWith('Z')), 'timestamps are UTC')
  check(typeof cleanup === 'function', 'apply returns a disposer')
  delete process.env.DSH_PET_PLATFORM
  delete process.env.DSH_PET_ARCH
}

/** Case 2: bootstrap copies both files, once, and never overwrites. */
async function caseBootstrap() {
  delete process.env.DSH_PET_NO_BOOTSTRAP
  const local = path.join(root, 'local-bootstrap')
  const petDir = path.join(local, 'AI-Terminal-Pet')
  mkdirSync(local, { recursive: true })

  const fakeRoot = path.join(root, 'pkg-bootstrap')
  mkdirSync(path.join(fakeRoot, 'lib'), { recursive: true })
  mkdirSync(path.join(fakeRoot, 'vendor'), { recursive: true })
  writeFileSync(path.join(fakeRoot, 'lib', 'index.js'), readFileSync(source, 'utf8'), 'utf8')
  writeFileSync(path.join(fakeRoot, 'vendor', 'protocol.js'), VENDOR_PROTOCOL, 'utf8')
  writeFileSync(path.join(fakeRoot, 'vendor', 'AI-Terminal-Pet.exe'), VENDOR_EXE, 'utf8')

  // This case is about the Windows path (a vendored binary, offline), so it
  // pins the platform instead of following the machine running it: the same
  // assertions then hold on the Linux and macOS CI jobs.
  process.env.LOCALAPPDATA = local
  process.env.DSH_PET_TEST = '1'
  process.env.DSH_PET_PACKAGE_ROOT = fakeRoot
  process.env.DSH_PET_PLATFORM = 'win32'
  process.env.DSH_PET_ARCH = 'x64'

  check(!existsSync(petDir), 'bootstrap starts from a machine with no AI-Terminal-Pet directory')

  const handlers = new Map()
  const plugin = await load(path.join(fakeRoot, 'lib', 'index.js'))
  plugin.apply(fakeContext(handlers), { enabled: true })
  const outcome = await plugin.__testing.bootstrap({ logger: { warn: () => {} } })

  const exe = path.join(petDir, 'AI-Terminal-Pet.exe')
  const protocol = path.join(petDir, 'protocol.js')
  check(outcome.copiedProtocol === true, 'bootstrap copies protocol.js into LOCALAPPDATA')
  check(outcome.copiedExe === true, 'bootstrap copies AI-Terminal-Pet.exe into LOCALAPPDATA')
  check(existsSync(exe) && statSync(exe).size === Buffer.byteLength(VENDOR_EXE), 'the installed exe matches the vendored bytes')
  check(existsSync(protocol) && sha256(protocol) === sha256(path.join(fakeRoot, 'vendor', 'protocol.js')), 'the installed protocol.js matches the vendored bytes')
  const strays = readdirSync(petDir).filter((name) => name.endsWith('.tmp'))
  check(strays.length === 0, 'no temporary files are left behind')

  // A second run sees the installed exe and must not touch anything.
  const sentinel = 'sentinel: a user-edited file the plugin must never replace\n'
  writeFileSync(exe, sentinel, 'utf8')
  writeFileSync(protocol, sentinel, 'utf8')
  const again = await load(path.join(fakeRoot, 'lib', 'index.js'))
  const outcomeAgain = await again.__testing.bootstrap({ logger: { warn: () => {} } })
  check(outcomeAgain.skipped === 'installed', 'a second bootstrap skips an installed overlay')
  check(readFileSync(exe, 'utf8') === sentinel, 'bootstrap never overwrites an existing exe')
  check(readFileSync(protocol, 'utf8') === sentinel, 'bootstrap never overwrites an existing protocol.js')

  // DSH_PET_NO_BOOTSTRAP=1 must keep the machine untouched.
  const localOptOut = path.join(root, 'local-optout')
  mkdirSync(localOptOut, { recursive: true })
  process.env.LOCALAPPDATA = localOptOut
  process.env.DSH_PET_NO_BOOTSTRAP = '1'
  const optOut = await load(path.join(fakeRoot, 'lib', 'index.js'))
  const outcomeOptOut = await optOut.__testing.bootstrap({})
  check(outcomeOptOut.skipped === 'disabled', 'DSH_PET_NO_BOOTSTRAP=1 disables the bootstrap')
  check(!existsSync(path.join(localOptOut, 'AI-Terminal-Pet')), 'the opt-out writes nothing to disk')
  delete process.env.DSH_PET_NO_BOOTSTRAP
  delete process.env.DSH_PET_PLATFORM
  delete process.env.DSH_PET_ARCH
}

/**
 * Case 2b: the data directory, the binary name and the platform key, on all
 * three platforms, plus the cross-file consistency that matters - the plugin's
 * answer must equal protocol.js's own data directory on whichever OS runs this.
 */
async function casePlatformDirectories() {
  const home = path.join(root, 'home-platforms')
  const xdg = path.join(root, 'xdg-platforms')
  mkdirSync(home, { recursive: true })
  mkdirSync(xdg, { recursive: true })
  const previous = {
    HOME: process.env.HOME,
    USERPROFILE: process.env.USERPROFILE,
    XDG_DATA_HOME: process.env.XDG_DATA_HOME,
    DSH_PET_PLATFORM: process.env.DSH_PET_PLATFORM,
    DSH_PET_ARCH: process.env.DSH_PET_ARCH,
  }
  process.env.DSH_PET_TEST = '1'
  process.env.HOME = home
  process.env.USERPROFILE = home
  delete process.env.XDG_DATA_HOME

  const fakeRoot = path.join(root, 'pkg-platforms')
  mkdirSync(path.join(fakeRoot, 'lib'), { recursive: true })
  writeFileSync(path.join(fakeRoot, 'lib', 'index.js'), readFileSync(source, 'utf8'), 'utf8')
  process.env.DSH_PET_PACKAGE_ROOT = fakeRoot
  const plugin = await load(path.join(fakeRoot, 'lib', 'index.js'))

  process.env.DSH_PET_PLATFORM = 'darwin'
  check(
    plugin.__testing.petDirectory() === path.join(home, 'Library', 'Application Support', 'AI-Terminal-Pet'),
    'macOS data directory is ~/Library/Application Support/AI-Terminal-Pet',
  )
  check(plugin.__testing.overlayBinaryName() === 'AI-Terminal-Pet', 'macOS binary has no .exe suffix')
  process.env.DSH_PET_ARCH = 'arm64'
  check(plugin.__testing.platformKey() === 'darwin-arm64', 'arm64 maps to darwin-arm64')

  process.env.DSH_PET_PLATFORM = 'linux'
  check(
    plugin.__testing.petDirectory() === path.join(home, '.local', 'share', 'ai-terminal-pet'),
    'Linux without XDG_DATA_HOME falls back to ~/.local/share/ai-terminal-pet',
  )
  process.env.XDG_DATA_HOME = xdg
  check(
    plugin.__testing.petDirectory() === path.join(xdg, 'ai-terminal-pet'),
    'Linux honours an absolute XDG_DATA_HOME',
  )
  process.env.XDG_DATA_HOME = 'relative/ignored'
  check(
    plugin.__testing.petDirectory() === path.join(home, '.local', 'share', 'ai-terminal-pet'),
    'a relative XDG_DATA_HOME is ignored, as the spec says',
  )
  delete process.env.XDG_DATA_HOME
  process.env.DSH_PET_ARCH = 'x64'
  check(plugin.__testing.platformKey() === 'linux-amd64', 'x64 maps to linux-amd64')

  process.env.DSH_PET_PLATFORM = 'win32'
  check(
    plugin.__testing.petDirectory() === path.join(process.env.LOCALAPPDATA, 'AI-Terminal-Pet'),
    'Windows data directory is %LOCALAPPDATA%\\AI-Terminal-Pet',
  )
  check(plugin.__testing.overlayBinaryName() === 'AI-Terminal-Pet.exe', 'Windows binary keeps the .exe suffix')

  // Cross-file consistency: only meaningful on the platform actually running,
  // because protocol.js reads process.platform rather than the test seam.
  delete process.env.DSH_PET_PLATFORM
  const real = await load(path.join(fakeRoot, 'lib', 'index.js'))
  const requireFromSource = createRequire(source)
  const realProtocol = requireFromSource(path.join(packageDir, 'vendor', 'protocol.js'))
  const snapshot = realProtocol.snapshotPath({ source: 'deepseek-harness', sessionId: 'ses_probe' })
  check(
    typeof snapshot === 'string' && snapshot.startsWith(real.__testing.petDirectory() + path.sep),
    'the plugin data directory is the parent protocol.js writes under (' + process.platform + ')',
  )

  for (const [key, value] of Object.entries(previous)) {
    if (value === undefined) delete process.env[key]
    else process.env[key] = value
  }
  delete process.env.DSH_PET_ARCH
}

/** A minimal USTAR writer, so the pinned-install case needs no fixture file. */
function tarHeader(name, size) {
  const header = Buffer.alloc(512)
  header.write(name, 0, 100, 'utf8')
  header.write('0000644\0', 100, 8, 'ascii')
  header.write('0000000\0', 108, 8, 'ascii')
  header.write('0000000\0', 116, 8, 'ascii')
  header.write(size.toString(8).padStart(11, '0') + '\0', 124, 12, 'ascii')
  header.write('00000000000\0', 136, 12, 'ascii')
  header.write('        ', 148, 8, 'ascii')
  header.write('0', 156, 1, 'ascii')
  header.write('ustar\0', 257, 6, 'ascii')
  header.write('00', 263, 2, 'ascii')
  let sum = 0
  for (const byte of header) sum += byte
  header.write(sum.toString(8).padStart(6, '0') + '\0 ', 148, 8, 'ascii')
  return header
}

function makeTarGz(members) {
  const chunks = []
  for (const member of members) {
    const data = Buffer.from(member.data)
    chunks.push(tarHeader(member.name, data.length), data, Buffer.alloc((512 - (data.length % 512)) % 512))
  }
  chunks.push(Buffer.alloc(1024))
  return gzipSync(Buffer.concat(chunks))
}

/**
 * Case 2c: the pinned install. Linux and macOS ship no binary, so the plugin
 * fetches the pinned release archive and extracts one member. A local HTTP
 * server serves the archive, so nothing here touches the network.
 */
async function casePinnedInstall() {
  const payload = Buffer.from('#!/bin/sh\n# fake overlay for the pinned-install case\n')
  const archive = makeTarGz([
    { name: 'ai-terminal-pet-9.9.9-linux-amd64/README.install.txt', data: 'read me\n' },
    { name: 'ai-terminal-pet-9.9.9-linux-amd64/bin/AI-Terminal-Pet', data: payload },
  ])
  const server = createServer((request, response) => {
    if (request.url !== '/overlay.tar.gz') {
      response.writeHead(404).end()
      return
    }
    response.writeHead(200, { 'content-type': 'application/octet-stream', 'content-length': archive.length })
    response.end(archive)
  })
  await new Promise((resolve) => server.listen(0, '127.0.0.1', resolve))
  const port = server.address().port
  const url = 'http://127.0.0.1:' + String(port) + '/overlay.tar.gz'

  const home = path.join(root, 'home-pinned')
  mkdirSync(home, { recursive: true })
  const fakeRoot = path.join(root, 'pkg-pinned')
  mkdirSync(path.join(fakeRoot, 'lib'), { recursive: true })
  mkdirSync(path.join(fakeRoot, 'vendor'), { recursive: true })
  writeFileSync(path.join(fakeRoot, 'lib', 'index.js'), readFileSync(source, 'utf8'), 'utf8')
  writeFileSync(path.join(fakeRoot, 'vendor', 'protocol.js'), VENDOR_PROTOCOL, 'utf8')

  const pin = {
    url,
    sha256: createHash('sha256').update(archive).digest('hex'),
    bytes: archive.length,
    member: 'ai-terminal-pet-9.9.9-linux-amd64/bin/AI-Terminal-Pet',
  }
  writeFileSync(
    path.join(fakeRoot, 'lib', 'release-pins.json'),
    JSON.stringify({ schemaVersion: 2, overlayVersion: '9.9.9', platforms: { 'linux-amd64': pin } }),
    'utf8',
  )

  const previous = { HOME: process.env.HOME, USERPROFILE: process.env.USERPROFILE, XDG_DATA_HOME: process.env.XDG_DATA_HOME }
  process.env.HOME = home
  process.env.USERPROFILE = home
  delete process.env.XDG_DATA_HOME
  process.env.DSH_PET_TEST = '1'
  process.env.DSH_PET_PACKAGE_ROOT = fakeRoot
  process.env.DSH_PET_PLATFORM = 'linux'
  process.env.DSH_PET_ARCH = 'x64'
  delete process.env.DSH_PET_NO_BOOTSTRAP

  try {
    const target = path.join(home, '.local', 'share', 'ai-terminal-pet', 'AI-Terminal-Pet')
    const first = await load(path.join(fakeRoot, 'lib', 'index.js'))
    const outcome = await first.__testing.bootstrap({ logger: { warn: () => {} } })
    check(outcome.downloaded === true && outcome.copiedExe === true, 'Linux installs the overlay from the pinned archive')
    check(existsSync(target), 'the extracted overlay lands in the XDG data directory')
    check(existsSync(target) && readFileSync(target).equals(payload), 'the extracted bytes are the archive member')
    if (process.platform !== 'win32') {
      check((statSync(target).mode & 0o777) === 0o755, 'the installed overlay is executable (0755)')
    }

    // A wrong digest must install nothing at all.
    const badHome = path.join(root, 'home-pinned-bad')
    mkdirSync(badHome, { recursive: true })
    process.env.HOME = badHome
    process.env.USERPROFILE = badHome
    writeFileSync(
      path.join(fakeRoot, 'lib', 'release-pins.json'),
      JSON.stringify({
        schemaVersion: 2,
        overlayVersion: '9.9.9',
        platforms: { 'linux-amd64': { ...pin, sha256: 'f'.repeat(64) } },
      }),
      'utf8',
    )
    const bad = await load(path.join(fakeRoot, 'lib', 'index.js'))
    const badOutcome = await bad.__testing.bootstrap({ logger: { warn: () => {} } })
    check(badOutcome.skipped === 'download-failed', 'a digest mismatch is reported, not installed')
    check(
      !existsSync(path.join(badHome, '.local', 'share', 'ai-terminal-pet', 'AI-Terminal-Pet')),
      'a digest mismatch writes no binary',
    )

    // A platform with no pin says so (and still copies protocol.js).
    process.env.DSH_PET_ARCH = 'riscv64'
    const missing = await load(path.join(fakeRoot, 'lib', 'index.js'))
    const missingOutcome = await missing.__testing.bootstrap({ logger: { warn: () => {} } })
    check(missingOutcome.skipped === 'no-pinned-overlay', 'an unpinned platform is reported as such')
  } finally {
    await new Promise((resolve) => server.close(resolve))
    for (const [key, value] of Object.entries(previous)) {
      if (value === undefined) delete process.env[key]
      else process.env[key] = value
    }
    delete process.env.DSH_PET_PLATFORM
    delete process.env.DSH_PET_ARCH
  }
}

/** Case 3: a hostile protocol module leaves the plugin silent. */
async function caseThrowingProtocol() {
  const local = path.join(root, 'local-throwing')
  const petDir = path.join(local, 'AI-Terminal-Pet')
  mkdirSync(petDir, { recursive: true })
  writeFileSync(path.join(petDir, 'protocol.js'), 'throw new Error("protocol import boom")\n', 'utf8')

  const fakeRoot = path.join(root, 'pkg-throwing')
  mkdirSync(path.join(fakeRoot, 'lib'), { recursive: true })
  writeFileSync(path.join(fakeRoot, 'lib', 'index.js'), readFileSync(source, 'utf8'), 'utf8')

  process.env.LOCALAPPDATA = local
  process.env.DSH_PET_TEST = '1'
  process.env.DSH_PET_PACKAGE_ROOT = fakeRoot
  process.env.DSH_PET_NO_BOOTSTRAP = '1'
  process.env.DSH_PET_PLATFORM = 'win32'
  process.env.DSH_PET_ARCH = 'x64'

  const handlers = new Map()
  const plugin = await load(path.join(fakeRoot, 'lib', 'index.js'))
  let threw = false
  try {
    plugin.apply(fakeContext(handlers), { enabled: true })
    for (const event of ['turn/start', 'tool/call', 'approval/asked', 'turn/end']) {
      handlers.get('session/event')({ header: { id: 'ses_x' } }, { type: event, data: { name: 'Bash' } })
    }
    handlers.get('approval/request')({ agent: { id: 'ses_x' }, toolName: 'Bash' }, () => 'next-called')
  } catch {
    threw = true
  }
  check(threw === false, 'a protocol module that throws on import never reaches the host')
  check(handlers.size > 0, 'event subscription survives a broken protocol module')
  check(plugin.__testing.officialProtocol() === null, 'a throwing protocol module is treated as absent')
  delete process.env.DSH_PET_NO_BOOTSTRAP
  delete process.env.DSH_PET_PLATFORM
  delete process.env.DSH_PET_ARCH
}

/** Case 4: the package can be loaded the way DSH loads it, from package.json. */
async function caseManifest() {
  const manifest = JSON.parse(readFileSync(path.join(packageDir, 'package.json'), 'utf8'))
  check(manifest.name === 'dsh-ai-terminal-pet', 'package name is dsh-ai-terminal-pet')
  // Any released 1.x will do: the version is bumped per release, and pinning the
  // assertion to one number just makes every release edit it.
  check(/^1\.[0-9]+\.[0-9]+$/.test(manifest.version), 'version is a released 1.x')
  check(manifest.type === 'module', 'the package is ESM')
  check(manifest.main === 'lib/index.js', 'main is lib/index.js')
  check(manifest.exports['.'] === './lib/index.js', 'exports["."] resolves lib/index.js')
  check(manifest.dependencies === undefined, 'zero runtime dependencies')
  check(manifest.scripts?.postinstall === undefined && manifest.scripts?.preinstall === undefined, 'no install lifecycle scripts')
  // vendor/ is committed and verified instead of regenerated at pack time. A
  // prepack hook makes installing from an unpinned git source fail with an
  // allowBuilds error that never mentions this package (see README, Install).
  check(manifest.scripts?.prepack === undefined, 'no prepack hook that would block a git install')
  check(typeof manifest.scripts?.['vendor:check'] === 'string', 'vendor:check verifies the committed vendor/ files')
  // The storefront entry accepts raster icons only (an SVG is rejected), so the
  // manifest icon and the one the registry renders must be the same PNG.
  check(manifest.icon === './icon.png', 'icon points at the raster icon.png')
  check(manifest.exports?.['./icon'] === './icon.png', 'exports["./icon"] is that same PNG')
  check(manifest.files.includes('screenshots.json'), 'the screenshot declaration ships in the tarball')
  check(manifest.files.includes('screenshots'), 'the screenshots themselves ship in the tarball')
  // Catalogues that render a plugin README pick the file from a directory
  // listing and decide its language by sniffing the text for CJK; a second
  // readme-shaped file is what turns a catalogue's "only an English version
  // exists" note into a real Chinese page. Both the shipping and the language
  // are pinned here, so a translation that quietly turns back into English, or
  // stops shipping, fails a check instead of the marketplace page.
  check(manifest.files.includes('README.zh.md'), 'the tarball ships the Chinese README')
  const zhReadme = readFileSync(path.join(packageDir, 'README.zh.md'), 'utf8')
  const cjkCount = (zhReadme.match(/[\u4e00-\u9fff]/g) ?? []).length
  check(
    cjkCount > zhReadme.length / 10,
    `README.zh.md reads as Chinese under a CJK sniff (${cjkCount} CJK characters)`
  )
  check(manifest.dsh?.bundle?.patch === './cordis.patch.yml', 'dsh.bundle.patch is declared')
  check(
    JSON.stringify(manifest.dsh?.bundle?.platforms) === '["win32","linux","darwin"]',
    'the bundle loads on Windows, Linux and macOS',
  )
  // Third-party registries (YELEBAI/dsh-plugin-marketplace, and any other that
  // reads the same convention) filter a plugin out unless install.profiles
  // contains the profile the client is running. Without this declaration the
  // registry falls back to [headless, web] and the desktop storefront hides the
  // entry; DSH itself never reads this field.
  check(JSON.stringify(manifest.dsh?.marketplace?.profiles) === '["desktop","headless","web"]', 'dsh.marketplace.profiles covers the desktop, headless and web profiles')
  check(manifest.dsh.marketplace.requiresBuildApproval === false, 'the package ships prebuilt runtime files, so no build approval is needed')
  check(manifest.dsh.marketplace.manualSteps === false, 'installing needs no manual step')
  // Since DSH 0.2.1-alpha.3 an own `engines.dsh` range decides alone and DSH peer
  // ranges are ignored (app-boot's evaluatePluginCompatibility, which evaluates
  // it with semver's includePrerelease), so this string is the entire
  // compatibility gate: a range the running host does not satisfy makes
  // `dsh plugin add` refuse the package and roll the profile back. Verified
  // against a real host on both ends of the range: 0.2.0-rc.1 and 0.2.2-rc.3
  // installed the tarball and composed its layer in `dsh --dump-config`.
  //
  // The `-0` bounds are not decoration. Under includePrerelease the documented
  // idiom ">=0.2.0 <0.3.0" refuses 0.2.0-rc.1 (a prerelease sorts below 0.2.0)
  // and admits 0.3.0-rc.1 (a prerelease sorts below 0.3.0) - measured with the
  // host's own evaluatePluginCompatibility. ">=0.2.0-0 <0.3.0-0" admits the whole
  // 0.2 line including its prereleases and nothing of 0.3.
  check(
    manifest.engines?.dsh === '>=0.2.0-0 <0.3.0-0',
    'engines.dsh targets the DSH 0.2 line (verified on 0.2.0-rc.1 and 0.2.2-rc.3)'
  )
  // DSH 0.2.x requires Node 24.9 or newer (v0.2.1-rc.2 dropped Node 22), and a
  // bundle only ever runs inside that host, so the package states the same floor
  // instead of advertising a runtime the host refuses to start on.
  check(manifest.engines?.node === '>=24.9.0', 'engines.node states the Node floor DSH 0.2.x requires')
  check(manifest.files.includes('vendor'), 'the vendored binary ships in the tarball')

  const patch = readFileSync(path.join(packageDir, 'cordis.patch.yml'), 'utf8')
  const code = patch.split('\n').filter((line) => line.trim() !== '' && !line.trim().startsWith('#'))
  check(code[0]?.trim() === '- insert:', 'cordis.patch.yml mounts with an insert list')
  // A row inside an `insert:` list is an indented `- id:` mapping; a bare row that
  // would patch an existing loader row instead of mounting anything starts at column 0.
  const rows = code.map((line) => line.trim())
  check(rows.some((line) => line === '- id: dsh-ai-terminal-pet'), 'the inserted row has the plugin id')
  check(rows.some((line) => line === 'name: dsh-ai-terminal-pet'), 'the inserted row names the package')
  check(!code.some((line) => /^-\s+id:/.test(line)), 'no top-level bare id row that would patch instead of insert')

  for (const locale of ['en', 'zh']) {
    const dictionary = JSON.parse(readFileSync(path.join(packageDir, 'locale', locale + '.json'), 'utf8'))
    check(typeof dictionary.meta?.title === 'string' && dictionary.meta.title.length > 0, 'locale/' + locale + '.json has meta.title')
    check(typeof dictionary.meta?.description === 'string' && dictionary.meta.description.length > 0, 'locale/' + locale + '.json has meta.description')
    check(dictionary.meta?.detail === undefined, 'locale/' + locale + '.json leaves meta.detail to the README (a path, not prose)')
  }
}

/**
 * Case 5: the repository-root shim manifest.
 *
 * `dsh plugin add github:owner/repo` reads the manifest at the REPOSITORY root,
 * not at packages/dsh-ai-terminal-pet. Without a root manifest that declares
 * dsh.bundle the host registers nothing and the install is a no-op with a
 * `declares no dsh.bundle` warning. The shim carries no code: it points at this
 * package and limits its `files` list to the plugin subtree, so a tarball built
 * from the root contains exactly what this package ships.
 */
async function caseRootManifest() {
  const rootManifestPath = path.join(packageDir, '..', '..', 'package.json')
  if (!existsSync(rootManifestPath)) {
    check(false, 'the repository root has a package.json shim')
    return
  }
  const rootManifest = JSON.parse(readFileSync(rootManifestPath, 'utf8'))
  const manifest = JSON.parse(readFileSync(path.join(packageDir, 'package.json'), 'utf8'))
  check(rootManifest.name === manifest.name, 'the root shim reuses the package name')
  check(rootManifest.version === manifest.version, 'the root shim tracks the package version')
  check(rootManifest.private === true, 'the root shim is private so it never claims the npm name')
  check(rootManifest.dsh?.bundle?.patch === './packages/dsh-ai-terminal-pet/cordis.patch.yml', 'the root shim points dsh.bundle.patch at this package')
  check(
    JSON.stringify(rootManifest.dsh?.bundle?.platforms) === '["win32","linux","darwin"]',
    'the root shim agrees about the platform list',
  )
  check(JSON.stringify(rootManifest.dsh?.marketplace) === JSON.stringify(manifest.dsh?.marketplace), 'the root shim copies the marketplace declaration verbatim')
  const patched = path.join(packageDir, '..', '..', rootManifest.dsh?.bundle?.patch ?? '')
  check(rootManifest.dsh?.bundle?.patch !== undefined && existsSync(patched), 'the patch file the root shim names exists')
  const entry = path.join(packageDir, '..', '..', rootManifest.main ?? '')
  check(typeof rootManifest.main === 'string' && existsSync(entry), 'the root shim main entry exists')
  check(JSON.stringify(rootManifest.files) === '["packages/dsh-ai-terminal-pet"]', 'a tarball built at the root contains only the plugin subtree')
}

/**
 * Case 6: vendor/manifest.json describes what is actually committed.
 *
 * The overlay cannot be compared byte-for-byte against a fresh CI build (the
 * go.mod line is a floor, not a pin, and different Go releases emit different
 * bytes), so scripts/vendor.mjs records a digest of the Go sources the binary
 * was built from. This case only checks the manifest is present, parses, and
 * still describes the two vendored files on disk; whether the sources have
 * moved on since is scripts/vendor.mjs --check's job, in CI.
 */
async function caseVendorManifest() {
  const vendorDir = path.join(packageDir, 'vendor')
  const manifestPath = path.join(vendorDir, 'manifest.json')
  check(existsSync(manifestPath), 'vendor/manifest.json is committed')
  if (!existsSync(manifestPath)) return
  let manifest
  try {
    manifest = JSON.parse(readFileSync(manifestPath, 'utf8'))
  } catch (error) {
    check(false, 'vendor/manifest.json is valid JSON (' + error.message + ')')
    return
  }
  check(manifest.schemaVersion === 1, 'vendor/manifest.json has schemaVersion 1')
  check(manifest.generatedBy === 'scripts/vendor.mjs', 'vendor/manifest.json names the script that wrote it')
  check(/^go1\.[0-9]+/.test(String(manifest.goVersion)), 'vendor/manifest.json records the Go release that built the overlay')
  check(typeof manifest.source?.sha256 === 'string' && /^[0-9a-f]{64}$/.test(manifest.source.sha256), 'vendor/manifest.json carries a source digest')
  check(typeof manifest.source?.files === 'number' && manifest.source.files > 0, 'vendor/manifest.json counts the Go files behind that digest')
  for (const [kind, entry, fileName] of [['overlay', manifest.overlay, 'AI-Terminal-Pet.exe'], ['protocol', manifest.protocol, 'protocol.js']]) {
    const file = path.join(vendorDir, fileName)
    check(entry?.file === 'vendor/' + fileName, 'vendor/manifest.json names ' + entry?.file + ' as the ' + kind)
    check(existsSync(file), 'the ' + kind + ' the manifest describes exists at vendor/' + fileName)
    if (!existsSync(file)) continue
    check(entry?.bytes === statSync(file).size, 'the ' + kind + ' size still matches the manifest')
    check(entry?.sha256 === sha256(file), 'the ' + kind + ' sha256 still matches the manifest')
  }
}

/**
 * Case 7: docs/market-index.json still points at this release.
 *
 * The repository publishes its own format-1 index so Discover works before npm
 * does. That index names one exact version and pins it to a commit, so it drifts
 * the moment either side changes: bump the package without touching the index and
 * Discover keeps offering the old build; move the tag forward and the pinned
 * commit installs a tree the manifests no longer describe. This case is the gate
 * between the two.
 */
async function caseMarketIndex() {
  const indexPath = path.join(packageDir, '..', '..', 'docs', 'market-index.json')
  if (!existsSync(indexPath)) {
    check(false, 'docs/market-index.json is committed')
    return
  }
  let index
  try {
    index = JSON.parse(readFileSync(indexPath, 'utf8'))
  } catch (error) {
    check(false, 'docs/market-index.json is valid JSON (' + error.message + ')')
    return
  }
  const manifest = JSON.parse(readFileSync(path.join(packageDir, 'package.json'), 'utf8'))
  check(index.format === 1, 'the market index declares format 1')
  check(typeof index.title?.en === 'string' && index.title.en.length > 0, 'the market index has an English title')
  check(Array.isArray(index.plugins) && index.plugins.length > 0, 'the market index lists at least one plugin')
  const entry = index.plugins?.find((plugin) => plugin.name === manifest.name)
  check(entry !== undefined, 'the market index lists ' + manifest.name)
  if (entry === undefined) return
  check(Array.isArray(entry.versions) && entry.versions.length > 0, 'the indexed plugin lists a version')
  const newest = entry.versions?.[0]
  check(newest?.version === manifest.version, 'the newest indexed version is the packaged version (' + manifest.version + ')')
  check(typeof newest?.publishTimestamp === 'number' && Number.isInteger(newest.publishTimestamp) && newest.publishTimestamp >= 0, 'the indexed version carries a Unix-seconds publish timestamp')
  const from = newest?.source ?? {}
  check(from.type === 'git', 'the indexed version installs from git until npm has the package')
  check(from.url === 'https://github.com/jjjjjjjjnnjnn/ai-terminal-pet', 'the indexed version names this repository as a canonical https URL')
  check(typeof from.commit === 'string' && /^[0-9a-f]{40}$/.test(from.commit), 'the indexed version pins a lowercase 40-hex commit')
  check(from.path === 'packages/dsh-ai-terminal-pet', 'the indexed version installs this package directory')
  check(newest?.engines?.dsh === manifest.engines?.dsh, 'the indexed engines.dsh matches the packaged requirement')
  const icon = typeof entry.icon === 'string' ? entry.icon : ''
  check(icon.endsWith('/packages/dsh-ai-terminal-pet/icon.png'), 'the index icon points at this package icon')
  check(icon.includes('/' + (from.commit ?? ' ') + '/'), 'the index icon is pinned to the same commit as the install source')
}
/**
 * Case 8: the vendored protocol must be loadable from the real package tree.
 *
 * Every other case builds a fake package root, and a fake root has no
 * package.json - so `require()` treats its vendor/protocol.js as CommonJS and
 * the test passes. The shipped package declares `"type": "module"`, which makes
 * every .js file inside it ESM, and `require()` of that file then returns a
 * module namespace whose `sendStatus` is undefined. The plugin treats that as
 * "no protocol" and reports nothing at all, silently. That is exactly the
 * fallback path used when the installed %LOCALAPPDATA% copy is missing, so it
 * has to be tested against the real tree, not a synthetic one.
 */
async function caseVendoredProtocolLoads() {
  const protocolPath = path.join(packageDir, 'vendor', 'protocol.js')
  const requireFromPackage = createRequire(path.join(packageDir, 'lib', 'index.js'))
  let loaded = null
  let error = ''
  try {
    loaded = requireFromPackage(protocolPath)
  } catch (thrown) {
    error = thrown && thrown.message ? thrown.message.split('\n')[0] : String(thrown)
  }
  check(error === '', 'the vendored protocol can be required from the shipped tree (' + error + ')')
  check(typeof loaded?.sendStatus === 'function', 'requiring the vendored protocol yields sendStatus')
  check(typeof loaded?.stableProcessId === 'function', 'requiring the vendored protocol yields stableProcessId')
  check(
    existsSync(path.join(packageDir, 'vendor', 'package.json')),
    'vendor/package.json marks the vendored directory CommonJS',
  )
}
try {
  await caseMapping()
  await caseBootstrap()
  await casePlatformDirectories()
  await casePinnedInstall()
  await caseThrowingProtocol()
  await caseManifest()
  await caseRootManifest()
  await caseVendorManifest()
  await caseMarketIndex()
  await caseVendoredProtocolLoads()
} finally {
  rmSync(root, { recursive: true, force: true })
}

console.log(failed ? '\nVERIFY FAILED' : '\nVERIFY PASSED')
process.exitCode = failed ? 1 : 0
