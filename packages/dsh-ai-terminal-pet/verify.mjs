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
import { createRequire } from 'node:module'
import os from 'node:os'
import path from 'node:path'
import { fileURLToPath, pathToFileURL } from 'node:url'

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

  const session = { header: { id: 'ses_abc' }, id: 'ses_abc' }
  const agent = { session: { header: { id: 'ses_abc' }, id: 'ses_abc' } }

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
  // 'error' appears twice on purpose: working-directory/change carries no state,
  // so a report is emitted only because the project changed. De-duplicating on
  // state alone used to swallow that event and leave the panel showing the old
  // directory until something else moved the state.
  const expected = [
    'working',
    'tool_use',
    'waiting_permission',
    'idle',
    'error',
    'error',
    'waiting_permission',
    'closed',
  ]

  check(JSON.stringify(states) === JSON.stringify(expected), 'state sequence ' + states.join(' -> '))
  check(nextResult === 'next-called', 'approval/request preserves the waterfall chain')
  check(stub.calls.length > 0 && stub.calls.every((item) => item.source === 'deepseek-harness' && item.terminal === 'deepseek-harness'), 'source/terminal are stable')
  check(stub.calls.every((item) => item.sessionId === 'ses_abc'), 'one stable session id, no event-id ghosts')
  check(stub.calls.every((item) => item.pid === 1701354615 && item.parentPid === process.pid), 'synthetic pid + real host parentPid')
  check(stub.calls.some((item) => item.tool === 'Bash'), 'tool name is forwarded')
  check(stub.calls.some((item) => item.project === 'C:\\work\\demo'), 'working directory is forwarded')
  check(stub.calls.at(-1)?.state === 'closed', 'process exit writes a closed report')
  check(handlers.size === 0, 'dispose removes every listener')
  check(afterCleanup === beforeCleanup, 'dispose does NOT end conversations (HMR is not an exit)')
  check(stub.calls.every((item) => item.updatedAt.endsWith('Z')), 'timestamps are UTC')
  check(stub.calls.every((item) => item.parentSessionId === ''), 'an empty parentSessionId is still set')
  check(typeof cleanup === 'function', 'apply returns a disposer')
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

  process.env.LOCALAPPDATA = local
  process.env.DSH_PET_TEST = '1'
  process.env.DSH_PET_PACKAGE_ROOT = fakeRoot

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
  check(manifest.dsh?.bundle?.patch === './cordis.patch.yml', 'dsh.bundle.patch is declared')
  check(JSON.stringify(manifest.dsh?.bundle?.platforms) === '["win32"]', 'the bundle is Windows only')
  // Third-party registries (YELEBAI/dsh-plugin-marketplace, and any other that
  // reads the same convention) filter a plugin out unless install.profiles
  // contains the profile the client is running. Without this declaration the
  // registry falls back to [headless, web] and the desktop storefront hides the
  // entry; DSH itself never reads this field.
  check(JSON.stringify(manifest.dsh?.marketplace?.profiles) === '["desktop","headless","web"]', 'dsh.marketplace.profiles covers the desktop, headless and web profiles')
  check(manifest.dsh.marketplace.requiresBuildApproval === false, 'the package ships prebuilt runtime files, so no build approval is needed')
  check(manifest.dsh.marketplace.manualSteps === false, 'installing needs no manual step')
  check(manifest.engines?.node === '^22.19 || >=24', 'engines.node covers Node 22.19 and 24+')
  check(typeof manifest.engines?.dsh === 'string' && manifest.engines.dsh.length > 0, 'engines.dsh declares a dsh requirement')
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
  check(JSON.stringify(rootManifest.dsh?.bundle?.platforms) === '["win32"]', 'the root shim keeps the Windows platform gate')
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
