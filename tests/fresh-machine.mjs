/**
 * Fresh-machine end-to-end suite.
 *
 * Every scenario runs inside a brand-new sandbox: its own LOCALAPPDATA, its own
 * USERPROFILE (so even the named pipe is a different one), its own DSH-style
 * node_modules and its own project directory. Nothing outside the sandbox is
 * read or written, with two deliberate exceptions that are restored:
 *
 *   - the single-instance mutex is `Local\AI-Terminal-Pet-v1-SingleInstance`,
 *     which is global to this Windows session, so a live overlay is stopped
 *     before the run and started again afterwards;
 *   - `scripts/install.ps1 -Uninstall` stops processes by name.
 *
 * Run it with:
 *
 *   node tests/fresh-machine.mjs                # full run
 *   node tests/fresh-machine.mjs --keep         # keep the sandbox for a look
 *   node tests/fresh-machine.mjs --no-overlay   # skip the scenarios that spawn a window
 *   node tests/fresh-machine.mjs --verbose      # print every check
 *
 * Exit code 0 means every check passed. The script requires the three binaries
 * under bin/ (build them first), Node 18+ and Windows PowerShell.
 */
import { spawn, spawnSync } from 'node:child_process'
import { createHash } from 'node:crypto'
import {
  existsSync,
  mkdirSync,
  mkdtempSync,
  readdirSync,
  readFileSync,
  writeFileSync,
} from 'node:fs'
import { rmSync } from 'node:fs'
import { tmpdir } from 'node:os'
import path from 'node:path'
import { fileURLToPath, pathToFileURL } from 'node:url'

const repoDir = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..')
const args = new Set(process.argv.slice(2))
const keepSandbox = args.has('--keep')
const verbose = args.has('--verbose')
const runGui = !args.has('--no-overlay')

let failures = 0
let checks = 0
const started = Date.now()

function sha256(file) {
  return createHash('sha256').update(readFileSync(file)).digest('hex')
}

function check(label, condition, detail = '') {
  checks++
  if (condition) {
    if (verbose) console.log('   ok   ' + label)
    return true
  }
  failures++
  console.log('   FAIL ' + label + (detail ? '  <- ' + detail : ''))
  return false
}

function equal(label, actual, want) {
  return check(label, actual === want, 'got ' + JSON.stringify(actual) + ', want ' + JSON.stringify(want))
}

function section(name) {
  console.log('\n== ' + name + ' ==')
}

function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms))
}

/** Synchronous nap for the cleanup paths. */
function nap(ms) {
  const until = Date.now() + ms
  while (Date.now() < until) {
    // Intentionally empty: this script has no event loop work to do here.
  }
}

/**
 * Kill anything still running from inside the sandbox. `--quit` is the polite
 * route and is always tried first, but a control command such as `--settings`
 * legitimately starts an instance when none is listening, and that instance
 * cannot always be asked to leave again by pipe (the sandbox pipe is derived
 * from the sandbox USERPROFILE). Without this, the sandbox directory stays
 * locked and the run looks like it leaked a process.
 */
function killSandboxProcesses(sandboxRoot) {
  const script =
    'Get-CimInstance Win32_Process -ErrorAction SilentlyContinue | ' +
    'Where-Object { $_.ExecutablePath -and $_.ExecutablePath.StartsWith(' + JSON.stringify(sandboxRoot) + ', [System.StringComparison]::OrdinalIgnoreCase) } | ' +
    'ForEach-Object { try { Stop-Process -Id $_.ProcessId -Force -ErrorAction Stop } catch {} }'
  spawnSync('powershell', ['-NoProfile', '-NonInteractive', '-Command', script], { encoding: 'utf8', timeout: 60000 })
}

/** rmSync with retries: Windows keeps handles for a moment after an exit. */
function removeWithRetry(target) {
  for (let attempt = 0; attempt < 12; attempt++) {
    try {
      rmSync(target, { recursive: true, force: true })
      return true
    } catch {
      nap(250)
    }
  }
  return false
}

/** Sandbox layout: one root, and nothing shared with the real machine. */
function makeSandbox() {
  const root = mkdtempSync(path.join(tmpdir(), 'pet-fresh-'))
  const sandbox = {
    root,
    localAppData: path.join(root, 'local'),
    userProfile: path.join(root, 'home'),
    dshHome: path.join(root, 'dsh'),
    project: path.join(root, 'project'),
    petDir: path.join(root, 'local', 'AI-Terminal-Pet'),
    runtime: path.join(root, 'local', 'AI-Terminal-Pet', 'runtime'),
    nodeModules: path.join(root, 'dsh', 'node_modules'),
  }
  for (const dir of [sandbox.localAppData, sandbox.userProfile, sandbox.dshHome, sandbox.project, sandbox.nodeModules]) {
    mkdirSync(dir, { recursive: true })
  }
  writeFileSync(path.join(sandbox.project, 'README.md'), 'sandbox project\n', 'utf8')

  /** Environment for every child process in this scenario. */
  sandbox.env = {
    ...process.env,
    LOCALAPPDATA: sandbox.localAppData,
    USERPROFILE: sandbox.userProfile,
    HOME: sandbox.userProfile,
    DSH_HOME: sandbox.dshHome,
    APPDATA: sandbox.localAppData,
    // Never let a test inherit a real install's overrides.
    AI_TERMINAL_PET_EXE: '',
    DSH_PET_TEST: '',
    DSH_PET_PACKAGE_ROOT: '',
    DSH_PET_NO_BOOTSTRAP: '',
  }
  return sandbox
}

/**
 * Run `fn` with the sandbox environment installed in this process, then put the
 * real one back. The plugin reads LOCALAPPDATA at call time, so a scenario that
 * loads it in-process has to swap the variables instead of passing them to a
 * child. Restoration happens in finally so a failed assertion cannot leak a
 * sandbox path into the next scenario - a later "pass" would then mean nothing.
 */
async function withSandboxEnv(env, fn) {
  const keys = [
    'LOCALAPPDATA',
    'USERPROFILE',
    'HOME',
    'DSH_HOME',
    'APPDATA',
    'AI_TERMINAL_PET_EXE',
    'DSH_PET_TEST',
    'DSH_PET_PACKAGE_ROOT',
    'DSH_PET_NO_BOOTSTRAP',
  ]
  const saved = new Map()
  for (const key of keys) saved.set(key, process.env[key])
  try {
    for (const key of keys) {
      if (env[key] === undefined || env[key] === '') delete process.env[key]
      else process.env[key] = env[key]
    }
    return await fn()
  } finally {
    for (const [key, value] of saved) {
      if (value === undefined) delete process.env[key]
      else process.env[key] = value
    }
  }
}

function runExe(exe, argv, env, timeout = 60000) {
  return spawnSync(exe, argv, { encoding: 'utf8', timeout, env })
}

/** True when a live overlay answers on this sandbox's pipe. */
function overlayAnswers(exe, env) {
  const result = runExe(exe, ['--json'], env, 15000)
  return result.status === 0 && String(result.stdout || '').trim().startsWith('{')
}

async function waitForOverlay(exe, env, timeoutMs = 25000) {
  const deadline = Date.now() + timeoutMs
  while (Date.now() < deadline) {
    if (overlayAnswers(exe, env)) return true
    await sleep(250)
  }
  return false
}

function readOverlaySessions(exe, env) {
  const result = runExe(exe, ['--json'], env, 15000)
  if (result.status !== 0) return null
  try {
    return JSON.parse(String(result.stdout || '').trim())
  } catch {
    return null
  }
}

/** The newest snapshot written for one session id, or null. */
function snapshotFor(sandbox, sessionId) {
  try {
    const files = readdirSync(sandbox.runtime)
      .filter((name) => name.endsWith('.json') && name.includes(sessionId))
      .sort()
    if (files.length === 0) return null
    return JSON.parse(readFileSync(path.join(sandbox.runtime, files[files.length - 1]), 'utf8'))
  } catch {
    return null
  }
}

/**
 * A small but faithful Cordis context: `on` registers and returns a disposer,
 * `emit` dispatches, and `logger.warn` is recorded so a silent failure can be
 * asserted against.
 */
function makeCordis() {
  const handlers = new Map()
  const warnings = []
  const ctx = {
    on(name, handler) {
      const list = handlers.get(name) || []
      list.push(handler)
      handlers.set(name, list)
      return () => {
        const current = handlers.get(name) || []
        const index = current.indexOf(handler)
        if (index >= 0) current.splice(index, 1)
      }
    },
    logger: { warn: (message) => warnings.push(String(message)) },
  }
  return {
    ctx,
    warnings,
    emit(name, ...payload) {
      for (const handler of [...(handlers.get(name) || [])]) handler(...payload)
    },
    count(name) {
      return (handlers.get(name) || []).length
    },
  }
}

async function main() {
  const builtOverlay = path.join(repoDir, 'bin', 'AI-Terminal-Pet.exe')
  const builtHook = path.join(repoDir, 'bin', 'pet-hook.exe')
  const builtStatusline = path.join(repoDir, 'bin', 'pet-statusline.exe')
  for (const file of [builtOverlay, builtHook, builtStatusline]) {
    if (!existsSync(file)) {
      console.error('missing build output: ' + file + '\nrun the three go build commands first')
      process.exit(2)
    }
  }

  const sandbox = makeSandbox()
  console.log('sandbox: ' + sandbox.root)

  // The binaries must be executed from a copy inside the sandbox, not from the
  // repository. On a machine whose checkout carries a Low mandatory integrity
  // label (this one does), a process started from bin/ runs at Low integrity and
  // cannot read the Medium-integrity %TEMP% sandbox - every config read would
  // look like "no config" and every write would fail. Copying them first makes
  // the sandbox self-contained and makes the run independent of where the
  // repository happens to live.
  sandbox.bin = path.join(sandbox.root, 'bin')
  mkdirSync(sandbox.bin, { recursive: true })
  const overlayExe = path.join(sandbox.bin, 'AI-Terminal-Pet.exe')
  const hookExe = path.join(sandbox.bin, 'pet-hook.exe')
  const statuslineExe = path.join(sandbox.bin, 'pet-statusline.exe')
  for (const [from, to] of [
    [builtOverlay, overlayExe],
    [builtHook, hookExe],
    [builtStatusline, statuslineExe],
  ]) {
    writeFileSync(to, readFileSync(from))
  }

  let sandboxOverlay = null
  const realEnv = { ...process.env }
  // Control commands talk to the real user's pipe, and a Low-integrity binary
  // (anything started from this checkout) cannot open a Medium-integrity pipe.
  // The installed copy is the one that can, and it is also the copy the user
  // would expect to come back afterwards.
  const installedRealOverlay = path.join(realEnv.LOCALAPPDATA || '', 'AI-Terminal-Pet', 'AI-Terminal-Pet.exe')
  const realOverlay = existsSync(installedRealOverlay) ? installedRealOverlay : builtOverlay
  const realWasRunning = runGui && overlayAnswers(realOverlay, realEnv)
  if (realWasRunning) {
    console.log('stopping the running overlay for the duration of the run')
    runExe(realOverlay, ['--quit'], realEnv, 15000)
    await sleep(1500)
    if (overlayAnswers(realOverlay, realEnv)) {
      console.log('warning: the overlay did not stop; the overlay scenarios may be skipped')
    }
  }

  try {
    scenarioFreshMachine(sandbox, overlayExe)
    const pkg = scenarioPackageInstall(sandbox)
    await scenarioPluginOnEmptyMachine(sandbox, pkg)
    await scenarioPluginDisabledPaths(sandbox, pkg)
    if (runGui) sandboxOverlay = await scenarioOverlayBoots(sandbox)
    scenarioStandaloneInstall(sandbox, overlayExe, hookExe, statuslineExe)
    await scenarioSettingsAndConfig(sandbox, overlayExe, hookExe)
    scenarioRobustness(sandbox, overlayExe, hookExe)
    scenarioUninstall(sandbox, overlayExe)
  } finally {
    // Ask the sandbox overlay to leave first: a graceful shutdown releases its
    // file handles, which is what decides whether the sandbox can be deleted.
    runExe(overlayExe, ['--quit'], sandbox.env, 15000)
    await sleep(1200)
    if (sandboxOverlay && sandboxOverlay.pid) {
      try {
        process.kill(sandboxOverlay.pid)
      } catch {
        // Already gone.
      }
    }
    killSandboxProcesses(sandbox.root)
    await sleep(600)
    if (realWasRunning) {
      console.log('\nrestoring the real overlay')
      const child = spawn(realOverlay, [], { detached: true, stdio: 'ignore', env: realEnv })
      child.unref()
      await waitForOverlay(realOverlay, realEnv)
    }
    if (!keepSandbox) {
      if (!removeWithRetry(sandbox.root)) console.log('could not remove the sandbox: ' + sandbox.root)
    } else {
      console.log('sandbox kept at ' + sandbox.root)
    }
  }

  const seconds = ((Date.now() - started) / 1000).toFixed(1)
  console.log(
    '\n' + (failures === 0 ? 'PASS' : 'FAIL') + ': ' + (checks - failures) + '/' + checks + ' checks in ' + seconds + 's',
  )
  process.exit(failures === 0 ? 0 : 1)
}

/** S1: a machine with nothing installed must say so, not crash or hang. */
function scenarioFreshMachine(sandbox, overlayExe) {
  section('S1 empty machine: CLI and doctor')
  check('no AI-Terminal-Pet directory yet', !existsSync(sandbox.petDir))

  const doctor = runExe(overlayExe, ['--doctor'], sandbox.env, 30000)
  equal('doctor exits 1 on an uninstalled machine', doctor.status, 1)
  check('doctor names the missing install', /install files/.test(doctor.stdout || ''), doctor.stdout)
  check('doctor prints a summary', /doctor: \d+ ok, \d+ warn, \d+ fail/.test(doctor.stdout || ''), doctor.stdout)

  const info = runExe(overlayExe, ['--json'], sandbox.env, 15000)
  equal('--json works without a running overlay', info.status, 0)
  let empty = null
  try {
    empty = JSON.parse(String(info.stdout || '').trim())
  } catch {
    empty = null
  }
  check('--json still returns a document', empty !== null, info.stdout)
  equal('the document reports no sessions', empty?.count, 0)
  equal('--pretty is accepted too', runExe(overlayExe, ['--json', '--pretty'], sandbox.env, 15000).status, 0)

  const version = runExe(overlayExe, ['--version'], sandbox.env, 15000)
  check('--version works before install', /0\.9\.0/.test(version.stdout || ''), version.stdout)
}

/** S2: pack the repository package and unpack it the way an installer would. */
function scenarioPackageInstall(sandbox) {
  section('S2 package install (npm pack then extract)')
  const packDir = path.join(sandbox.root, 'pack')
  mkdirSync(packDir, { recursive: true })
  // One command string with shell:true: `npm` is a .cmd shim on Windows, and
  // spawning a .cmd needs a shell. Passing an argv array alongside shell:true is
  // deprecated in Node 22+, hence the single string.
  const packCommand = 'npm pack --pack-destination "' + packDir + '" --silent'
  const packed = spawnSync(packCommand, {
    cwd: path.join(repoDir, 'packages', 'dsh-ai-terminal-pet'),
    encoding: 'utf8',
    timeout: 300000,
    env: { ...sandbox.env },
    shell: true,
  })
  const lines = String(packed.stdout || '')
    .trim()
    .split(/\r?\n/)
    .filter(Boolean)
  const tarball = (lines[lines.length - 1] || '').trim()
  check('npm pack produced a tarball', Boolean(tarball) && existsSync(path.join(packDir, tarball)), tarball + ' ' + packed.stderr)

  const target = path.join(sandbox.nodeModules, 'dsh-ai-terminal-pet')
  mkdirSync(target, { recursive: true })
  const extracted = spawnSync('tar', ['-xzf', path.join(packDir, tarball), '-C', target, '--strip-components=1'], {
    encoding: 'utf8',
    timeout: 120000,
  })
  check('the tarball extracts', extracted.status === 0, extracted.stderr)

  const pkg = {
    dir: target,
    manifest: JSON.parse(readFileSync(path.join(target, 'package.json'), 'utf8')),
    entry: path.join(target, 'lib', 'index.js'),
    vendorExe: path.join(target, 'vendor', 'AI-Terminal-Pet.exe'),
    vendorProtocol: path.join(target, 'vendor', 'protocol.js'),
    vendorManifest: path.join(target, 'vendor', 'manifest.json'),
  }
  equal('package name', pkg.manifest.name, 'dsh-ai-terminal-pet')
  check('entry point exists', existsSync(pkg.entry))
  check('vendored overlay ships with it', existsSync(pkg.vendorExe))
  check('vendored protocol ships with it', existsSync(pkg.vendorProtocol))
  check('vendored manifest ships with it', existsSync(pkg.vendorManifest))
  check('bundle declaration present', Boolean(pkg.manifest.dsh?.bundle), JSON.stringify(pkg.manifest.dsh))
  equal('bundle patch is the loader patch', pkg.manifest.dsh?.bundle?.patch, './cordis.patch.yml')
  check('locale files ship', existsSync(path.join(target, 'locale', 'en.json')) && existsSync(path.join(target, 'locale', 'zh.json')))
  const patch = readFileSync(path.join(target, 'cordis.patch.yml'), 'utf8')
  check('loader patch uses insert:', /\binsert:/.test(patch), patch)
  check('loader patch mounts this package', /name:\s*dsh-ai-terminal-pet/.test(patch), patch)
  return pkg
}

/** S3: load the plugin in a fake host on an empty machine and drain events. */
async function scenarioPluginOnEmptyMachine(sandbox, pkg) {
  section('S3 plugin bootstrap and event mapping on an empty machine')
  check('still nothing installed before the plugin loads', !existsSync(sandbox.petDir))

  await withSandboxEnv(sandbox.env, async () => {
    // A cache-busting query gives this scenario its own module instance, so
    // bootstrapPromise and the protocol cache are cold even when another
    // scenario already imported the same file.
    const plugin = await import(pathToFileURL(pkg.entry).href + '?scenario=' + Date.now())
    check('plugin exports apply()', typeof plugin.apply === 'function')
    const cordis = makeCordis()
    const previousCwd = process.cwd()
    process.chdir(sandbox.project)
    try {
      const dispose = plugin.apply(cordis.ctx, {})
      check('apply() returns a disposer', typeof dispose === 'function')

      const bootstrapOutcome = await plugin.__testing.bootstrap(cordis.ctx)
      check('bootstrap ran', bootstrapOutcome.skipped !== 'disabled', JSON.stringify(bootstrapOutcome))
      check('overlay exists after bootstrap', existsSync(path.join(sandbox.petDir, 'AI-Terminal-Pet.exe')), JSON.stringify(bootstrapOutcome))
      check('protocol exists after bootstrap', existsSync(path.join(sandbox.petDir, 'protocol.js')), JSON.stringify(bootstrapOutcome))
      check('the install landed inside the sandbox', existsSync(path.join(sandbox.localAppData, 'AI-Terminal-Pet', 'AI-Terminal-Pet.exe')))
      if (existsSync(path.join(sandbox.petDir, 'AI-Terminal-Pet.exe'))) {
        equal('copied overlay is the vendored binary', sha256(path.join(sandbox.petDir, 'AI-Terminal-Pet.exe')), sha256(pkg.vendorExe))
        equal('copied protocol is the vendored copy', sha256(path.join(sandbox.petDir, 'protocol.js')), sha256(pkg.vendorProtocol))
      }

      // bootstrap() is memoized: the second call must return the very same
      // promise result and must not touch the file again.
      const before = sha256(path.join(sandbox.petDir, 'AI-Terminal-Pet.exe'))
      const first = await plugin.__testing.bootstrap(cordis.ctx)
      const again = await plugin.__testing.bootstrap(cordis.ctx)
      check('bootstrap is memoized', first === again, JSON.stringify(again))
      equal('second bootstrap leaves the binary alone', sha256(path.join(sandbox.petDir, 'AI-Terminal-Pet.exe')), before)

      // Drain one event at a time and read back what hit the disk.
      const session = { session: { header: { id: 'ses-fresh-1' } } }
      const events = [
        ['turn/start', {}, { state: 'working' }],
        ['tool/call', { name: 'Bash' }, { state: 'tool_use', tool: 'Bash' }],
        ['tool/result', { error: { code: 'E_TOOL' } }, { state: 'error', reason: 'E_TOOL' }],
        ['approval/asked', { reason: 'needs approval', toolName: 'Shell' }, { state: 'waiting_permission', tool: 'Shell' }],
        ['approval/decided', {}, { state: 'working' }],
        ['turn/end', { reason: 'done' }, { state: 'idle', reason: 'done' }],
        ['command/done', { kind: 'error', text: 'build failed' }, { state: 'error', reason: 'build-failed' }],
      ]
      for (const [type, data, want] of events) {
        cordis.emit('session/event', session, { type, data })
        await sleep(40)
        const snapshot = snapshotFor(sandbox, 'ses-fresh-1')
        if (!check(type + ' wrote a snapshot', Boolean(snapshot))) continue
        equal(type + ' source', snapshot.source, 'deepseek-harness')
        equal(type + ' terminal', snapshot.terminal, 'deepseek-harness')
        equal(type + ' sessionId', snapshot.sessionId, 'ses-fresh-1')
        equal(type + ' project', snapshot.project, sandbox.project)
        equal(type + ' title', snapshot.title, 'DeepSeek Harness')
        if (want.state) equal(type + ' state', snapshot.state, want.state)
        if (want.reason !== undefined) equal(type + ' reason', snapshot.reason, want.reason)
        if (want.tool) equal(type + ' tool', snapshot.tool, want.tool)
      }

      // working-directory/change retargets the project, and session/title keeps
      // the stable window title while carrying the human title in reason.
      cordis.emit('session/event', session, { type: 'working-directory/change', data: { cwd: sandbox.dshHome } })
      await sleep(40)
      equal('working-directory/change retargets project', snapshotFor(sandbox, 'ses-fresh-1')?.project, sandbox.dshHome)
      cordis.emit('session/event', session, { type: 'session/title', data: { title: 'fix the parser' } })
      await sleep(40)
      const titled = snapshotFor(sandbox, 'ses-fresh-1')
      equal('session/title keeps the window title stable', titled?.title, 'DeepSeek Harness')
      equal('session/title carries the human title in reason', titled?.reason, 'fix-the-parser')

      // A second conversation must not collide with the first.
      const second = { session: { header: { id: 'ses-fresh-2' } } }
      cordis.emit('session/event', second, { type: 'tool/call', data: { name: 'Read' } })
      await sleep(40)
      const firstSnapshot = snapshotFor(sandbox, 'ses-fresh-1')
      const other = snapshotFor(sandbox, 'ses-fresh-2')
      check('two conversations file two snapshots', Boolean(firstSnapshot) && Boolean(other))
      equal('the second conversation keeps its own session id', other?.sessionId, 'ses-fresh-2')
      equal('the first conversation is untouched', firstSnapshot?.sessionId, 'ses-fresh-1')
      check('the two pids differ', firstSnapshot?.pid !== other?.pid, firstSnapshot?.pid + ' vs ' + other?.pid)

      // The waterfall listener must always call next().
      let nextCalls = 0
      cordis.emit('approval/request', { agent: session, toolName: 'Shell' }, () => {
        nextCalls++
      })
      await sleep(40)
      equal('approval/request calls next exactly once', nextCalls, 1)

      check('unknown session events are ignored', snapshotFor(sandbox, 'ses-unknown') === null)

      // Disposing must unhook without clearing the panel.
      const beforeDispose = snapshotFor(sandbox, 'ses-fresh-1')?.state
      dispose()
      equal('dispose removes every listener', cordis.count('session/event'), 0)
      cordis.emit('session/event', session, { type: 'turn/end', data: {} })
      await sleep(40)
      equal('nothing is reported after dispose', snapshotFor(sandbox, 'ses-fresh-1')?.state, beforeDispose)

      check('plugin never warned', cordis.warnings.length === 0, cordis.warnings.join(' | '))
    } finally {
      process.chdir(previousCwd)
    }
  })
}

/**
 * S3b: the documented switches. DSH_PET_NO_BOOTSTRAP must stop the copy while
 * reporting still works from the vendored protocol; config.enabled = false must
 * register nothing at all.
 */
async function scenarioPluginDisabledPaths(sandbox, pkg) {
  section('S3b plugin switches that must stay inert')
  const freshDir = path.join(sandbox.root, 'local-nobootstrap')
  await withSandboxEnv({ ...sandbox.env, LOCALAPPDATA: freshDir, DSH_PET_NO_BOOTSTRAP: '1' }, async () => {
    const plugin = await import(pathToFileURL(pkg.entry).href + '?scenario=nobootstrap')
    const cordis = makeCordis()
    const dispose = plugin.apply(cordis.ctx, {})
    const outcome = await plugin.__testing.bootstrap(cordis.ctx)
    equal('DSH_PET_NO_BOOTSTRAP disables the copy', outcome.skipped, 'disabled')
    equal('nothing was copied', existsSync(path.join(freshDir, 'AI-Terminal-Pet', 'AI-Terminal-Pet.exe')), false)

    cordis.emit('session/event', { session: { header: { id: 'ses-noboot' } } }, { type: 'tool/call', data: { name: 'Read' } })
    await sleep(80)
    const runtime = path.join(freshDir, 'AI-Terminal-Pet', 'runtime')
    const files = existsSync(runtime) ? readdirSync(runtime).filter((name) => name.endsWith('.json')) : []
    check('the vendored protocol still writes a snapshot', files.length > 0, runtime)
    if (files.length > 0) {
      const snapshot = JSON.parse(readFileSync(path.join(runtime, files[0]), 'utf8'))
      equal('the snapshot carries the session', snapshot.sessionId, 'ses-noboot')
      equal('the snapshot carries the state', snapshot.state, 'tool_use')
    }
    dispose()
  })

  const listeners = await withSandboxEnv(sandbox.env, async () => {
    const plugin = await import(pathToFileURL(pkg.entry).href + '?scenario=disabled')
    const cordis = makeCordis()
    const dispose = plugin.apply(cordis.ctx, { enabled: false })
    const count = cordis.count('session/event')
    dispose()
    return count
  })
  equal('config.enabled = false registers nothing', listeners, 0)
}

/** S4: the copied overlay actually boots in the sandbox and reports sessions. */
async function scenarioOverlayBoots(sandbox) {
  section('S4 the overlay boots on a machine with no prior state')
  const installed = path.join(sandbox.petDir, 'AI-Terminal-Pet.exe')
  check('the sandbox overlay is the file bootstrap wrote', existsSync(installed))

  const child = spawn(installed, [], { detached: true, stdio: 'ignore', env: sandbox.env })
  child.unref()
  const up = await waitForOverlay(installed, sandbox.env)
  check('the overlay answers on its own pipe', up)
  if (!up) return child

  const report = readOverlaySessions(installed, sandbox.env)
  check('--json returns a document', report !== null)
  check('the plugin sessions reached the overlay', (report?.count || 0) >= 2, 'count=' + report?.count)
  const sources = new Set((report?.sessions || []).map((item) => item.source))
  check('the DSH session is attributed to deepseek-harness', sources.has('deepseek-harness'), [...sources].join(','))

  equal('no config.json was needed to start', existsSync(path.join(sandbox.petDir, 'config.json')), false)

  equal('--toggle exits 0', runExe(installed, ['--toggle'], sandbox.env, 15000).status, 0)
  equal('--show exits 0', runExe(installed, ['--show'], sandbox.env, 15000).status, 0)
  equal('--reset exits 0', runExe(installed, ['--reset'], sandbox.env, 15000).status, 0)
  return child
}

/** S5: the standalone path - install.ps1 plus the two hook binaries. */
function scenarioStandaloneInstall(sandbox, overlayExe, hookExe, statuslineExe) {
  section('S5 standalone install and native hooks')
  const destination = path.join(sandbox.petDir, 'AI-Terminal-Pet.exe')
  const installScript = path.join(repoDir, 'scripts', 'install.ps1')
  const install = spawnSync(
    'powershell',
    ['-NoProfile', '-ExecutionPolicy', 'Bypass', '-File', installScript, '-Destination', destination],
    { encoding: 'utf8', timeout: 300000 },
  )
  check('install.ps1 succeeds', install.status === 0, install.stdout + install.stderr)
  check('installer reports 11 artifacts', /install: ok - 11 file/.test(install.stdout || ''), install.stdout)

  for (const name of ['AI-Terminal-Pet.exe', 'pet-statusline.exe', 'pet-hook.exe', 'protocol.js']) {
    check('installed ' + name, existsSync(path.join(sandbox.petDir, name)))
  }
  equal('installed hook is the built one', sha256(path.join(sandbox.petDir, 'pet-hook.exe')), sha256(hookExe))
  equal('installed statusline is the built one', sha256(path.join(sandbox.petDir, 'pet-statusline.exe')), sha256(statuslineExe))

  const verify = spawnSync(
    'powershell',
    ['-NoProfile', '-ExecutionPolicy', 'Bypass', '-File', installScript, '-Destination', destination, '-VerifyOnly'],
    { encoding: 'utf8', timeout: 300000 },
  )
  check(
    'verify-only matches everything',
    /install: verified - 11 file\(s\) match the repository, 0 warning/.test(verify.stdout || ''),
    verify.stdout,
  )

  const doctor = runExe(overlayExe, ['--doctor'], sandbox.env, 30000)
  check('doctor reports no failures after a standalone install', /doctor: \d+ ok, \d+ warn, 0 fail/.test(doctor.stdout || ''), doctor.stdout)

  const hook = spawnSync(path.join(sandbox.petDir, 'pet-hook.exe'), ['--mode=claude'], {
    input: JSON.stringify({
      hook_event_name: 'PreToolUse',
      session_id: 'ses-hook',
      cwd: sandbox.project,
      tool_name: 'Bash',
      tool_input: { command: 'ls' },
    }),
    encoding: 'utf8',
    timeout: 30000,
    env: sandbox.env,
  })
  equal('pet-hook exits 0', hook.status, 0)
  equal('pet-hook answers the Claude dialect', String(hook.stdout || '').trim(), '{}')

  const statusline = spawnSync(path.join(sandbox.petDir, 'pet-statusline.exe'), [], {
    input: JSON.stringify({ agent_state: 'working', conversation_id: 'ses-sl', workspace: sandbox.project }),
    encoding: 'utf8',
    timeout: 30000,
    env: sandbox.env,
  })
  equal('pet-statusline exits 0', statusline.status, 0)
  equal('pet-statusline prints the status line', String(statusline.stdout || '').trim(), 'pet working')

  const report = readOverlaySessions(overlayExe, sandbox.env)
  const ids = new Set((report?.sessions || []).map((item) => item.sessionId))
  if (report !== null) {
    check('the hook session reached the overlay', ids.has('ses-hook'), [...ids].join(','))
    check('the statusline session reached the overlay', ids.has('ses-sl'), [...ids].join(','))
  } else {
    console.log('   note: no running sandbox overlay, delivery asserted through snapshots instead')
    equal('the hook wrote a snapshot', snapshotFor(sandbox, 'ses-hook')?.sessionId, 'ses-hook')
    equal('the statusline wrote a snapshot', snapshotFor(sandbox, 'ses-sl')?.sessionId, 'ses-sl')
  }
}

/** S6: config defaults and the auto-approve handshake. */
async function scenarioSettingsAndConfig(sandbox, overlayExe, hookExe) {
  section('S6 settings, config and the approval handshake')
  const configPath = path.join(sandbox.petDir, 'config.json')
  rmSync(configPath, { force: true })

  // With auto-approve off (the default), a Claude permission gate must defer to
  // the terminal's own prompt: an empty object, not a decision.
  const gate = JSON.stringify({
    hook_event_name: 'PermissionRequest',
    session_id: 'ses-gate',
    cwd: sandbox.project,
    tool_name: 'Read',
    tool_input: { file_path: 'README.md' },
  })
  const deferred = spawnSync(hookExe, ['--mode=claude'], { input: gate, encoding: 'utf8', timeout: 30000, env: sandbox.env })
  equal('auto-approve off defers to the host', String(deferred.stdout || '').trim(), '{}')

  const antigravity = spawnSync(hookExe, ['--mode=antigravity'], {
    input: JSON.stringify({ event: 'BeforeTool', session_id: 'ses-ag', toolCall: { name: 'Bash' } }),
    encoding: 'utf8',
    timeout: 30000,
    env: sandbox.env,
  })
  equal('antigravity answers with an explicit decision', String(antigravity.stdout || '').trim(), '{"decision":"allow"}')

  writeFileSync(configPath, JSON.stringify({ autoApprove: true, autoApproveTimeoutSec: 1 }), 'utf8')
  const approved = spawnSync(hookExe, ['--mode=claude'], { input: gate, encoding: 'utf8', timeout: 30000, env: sandbox.env })
  equal('auto-approve with no veto allows', String(approved.stdout || '').trim(), '{"decision":"allow"}')

  const dangerous = spawnSync(hookExe, ['--mode=claude'], {
    input: JSON.stringify({
      hook_event_name: 'PermissionRequest',
      session_id: 'ses-gate',
      cwd: sandbox.project,
      tool_name: 'Bash',
      tool_input: { command: 'rm -rf build' },
    }),
    encoding: 'utf8',
    timeout: 30000,
    env: sandbox.env,
  })
  equal('a dangerous command falls through even when auto-approve is on', String(dangerous.stdout || '').trim(), '{}')

  const dangerousAg = spawnSync(hookExe, ['--mode=antigravity'], {
    input: JSON.stringify({ tool_confirmation_pending: true, session_id: 'ses-ag', command: 'rm -rf build' }),
    encoding: 'utf8',
    timeout: 30000,
    env: sandbox.env,
  })
  equal('antigravity denies a dangerous command', String(dangerousAg.stdout || '').trim(), '{"decision":"deny"}')

  // A veto written while the hook waits must turn the approval into a deny.
  const approvals = path.join(sandbox.petDir, 'approvals')
  mkdirSync(approvals, { recursive: true })
  const vetoChild = spawn(hookExe, ['--mode=claude'], { env: sandbox.env, stdio: ['pipe', 'pipe', 'pipe'] })
  let vetoStdout = ''
  vetoChild.stdout.on('data', (chunk) => {
    vetoStdout += chunk.toString('utf8')
  })
  vetoChild.stdin.end(gate)
  const vetoDeadline = Date.now() + 5000
  let vetoed = false
  while (Date.now() < vetoDeadline && !vetoed) {
    const entries = readdirSync(approvals).filter((name) => /^pending-.*\.json$/.test(name))
    if (entries.length > 0) {
      writeFileSync(path.join(approvals, entries[0] + '.deny'), '{"decision":"deny"}', 'utf8')
      vetoed = true
      break
    }
    await sleep(25)
  }
  await new Promise((resolve) => vetoChild.on('close', resolve))
  check('the overlay veto file is observed', vetoed)
  equal('a veto turns the decision into deny', vetoStdout.trim(), '{"decision":"deny"}')

  equal('--settings exits 0 while an instance is running', runExe(overlayExe, ['--settings'], sandbox.env, 20000).status, 0)
  // --settings starts an instance when none is listening; close whatever it left
  // so the next scenario starts from a known state.
  runExe(overlayExe, ['--quit'], sandbox.env, 15000)
  await sleep(600)
}

/** S7: hostile environments must fail quietly and fast, never hang the host. */
function scenarioRobustness(sandbox, overlayExe, hookExe) {
  section('S7 robustness on a broken machine')
  const configPath = path.join(sandbox.petDir, 'config.json')
  const good = readFileSync(configPath, 'utf8')

  writeFileSync(configPath, '{ this is not json', 'utf8')
  const corrupt = spawnSync(hookExe, ['--mode=claude'], {
    input: JSON.stringify({ hook_event_name: 'PermissionRequest', session_id: 'ses-corrupt', tool_name: 'Read' }),
    encoding: 'utf8',
    timeout: 30000,
    env: sandbox.env,
  })
  equal('a corrupt config falls back to off', String(corrupt.stdout || '').trim(), '{}')
  equal('a corrupt config does not fail the hook', corrupt.status, 0)
  writeFileSync(configPath, good, 'utf8')

  const bare = { ...sandbox.env, LOCALAPPDATA: '' }
  const noLocal = spawnSync(hookExe, ['--mode=antigravity'], {
    input: JSON.stringify({ event: 'BeforeTool', session_id: 'ses-nolocal' }),
    encoding: 'utf8',
    timeout: 30000,
    env: bare,
  })
  equal('no LOCALAPPDATA still answers', String(noLocal.stdout || '').trim(), '{"decision":"allow"}')
  equal('no LOCALAPPDATA still exits 0', noLocal.status, 0)

  const huge = 'x'.repeat(1 << 20)
  const oversized = spawnSync(hookExe, ['--mode=claude'], {
    input: JSON.stringify({ hook_event_name: 'PreToolUse', session_id: 'ses-huge', tool_input: { command: huge } }),
    encoding: 'utf8',
    timeout: 30000,
    env: sandbox.env,
  })
  equal('oversized stdin still exits 0', oversized.status, 0)
  equal('oversized stdin still answers', String(oversized.stdout || '').trim(), '{}')

  for (const [label, input] of [
    ['empty', ''],
    ['not json', 'hello'],
    ['array', '[1,2,3]'],
    ['null', 'null'],
  ]) {
    const result = spawnSync(hookExe, ['--mode=claude'], { input, encoding: 'utf8', timeout: 30000, env: sandbox.env })
    equal(label + ' stdin exits 0', result.status, 0)
    equal(label + ' stdin still answers', String(result.stdout || '').trim(), '{}')
  }

  equal(
    'an unknown mode still exits 0',
    spawnSync(hookExe, ['--mode=nonsense'], { input: '', encoding: 'utf8', timeout: 30000, env: sandbox.env }).status,
    0,
  )

  const missingRoot = path.join(sandbox.root, 'vendor', 'AI-Terminal-Pet.exe')
  equal(
    'an unknown flag still exits 0',
    spawnSync(hookExe, ['--frobnicate'], { input: '', encoding: 'utf8', timeout: 30000, env: sandbox.env }).status,
    0,
  )
  void missingRoot

  const stopped = runExe(overlayExe, ['--json'], { ...sandbox.env, USERPROFILE: path.join(sandbox.root, 'other-home') }, 15000)
  check('--json with no listener returns instead of hanging', stopped.status !== null, 'timeout')
}

/** S8: uninstall must leave the machine as clean as it found it. */
function scenarioUninstall(sandbox, overlayExe) {
  section('S8 uninstall')
  const destination = path.join(sandbox.petDir, 'AI-Terminal-Pet.exe')
  const uninstall = spawnSync(
    'powershell',
    [
      '-NoProfile',
      '-ExecutionPolicy',
      'Bypass',
      '-File',
      path.join(repoDir, 'scripts', 'install.ps1'),
      '-Destination',
      destination,
      '-Uninstall',
      '-Force',
    ],
    { encoding: 'utf8', timeout: 300000 },
  )
  check('uninstall succeeds', uninstall.status === 0, uninstall.stdout + uninstall.stderr)
  check('the install directory is gone', !existsSync(sandbox.petDir), 'still present')
  check('uninstall reports completion', /uninstall: done/.test(uninstall.stdout || ''), uninstall.stdout)
  equal('doctor fails again on the cleaned machine', runExe(overlayExe, ['--doctor'], sandbox.env, 30000).status, 1)
}

main().catch((error) => {
  console.error(error)
  process.exit(1)
})
