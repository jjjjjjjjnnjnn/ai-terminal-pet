/**
 * Vendor, or check, the two files this package ships.
 *
 *   node scripts/vendor.mjs          # copy ../../bin + ../../reporters into vendor/
 *   node scripts/vendor.mjs --check  # verify what is in vendor/, write nothing
 *
 * Copies ../../bin/AI-Terminal-Pet.exe -> vendor/AI-Terminal-Pet.exe
 * and     ../../reporters/protocol.js   -> vendor/protocol.js
 * and then records what it copied in vendor/manifest.json.
 *
 * The protocol copy is checked byte-for-byte against its source, because the Go
 * server and the reporters must agree on the wire contract; a drifting copy
 * would silently break every report.
 *
 * Copy mode is idempotent: a destination that already matches is left alone.
 *
 * Why staleness is a SOURCE DIGEST and not a byte comparison
 * --------------------------------------------------------
 * A committed overlay cannot be compared byte-for-byte against a fresh CI build.
 * go.mod's `go 1.23.0` is a floor, not a pin: GitHub's setup-go installs go1.23.0
 * while a developer machine may run go1.27.0, and the same sources compile to
 * different bytes under different compilers (4,371,968 B under go1.27.0 versus
 * 3,443,200 B under go1.23.0 for the very same tree). Byte equality across
 * toolchains is unattainable, so it must never gate CI.
 *
 * The question that actually matters is "does the committed overlay predate a
 * change under cmd/ or internal/?". Copy mode therefore records, in
 * vendor/manifest.json, one sha256 folded over every non-test *.go file it can
 * see (path + content hash, sorted) plus the toolchain that produced the binary.
 * `*_test.go` files are excluded on purpose: they are never compiled into the
 * binary, so editing a test must not demand a re-vendor.
 * Check mode recomputes that digest: a mismatch means the sources moved on and
 * the overlay is stale, whatever compiler built it. The byte comparison against
 * bin/ is kept as a bonus, but only when the manifest names the same Go release
 * as the machine running the check -- otherwise it is skipped on purpose.
 *
 * Check mode is what `npm run verify` runs, and it exists because this package
 * deliberately has no `prepack` hook. A pnpm `prepack`/`prepare` step makes
 * installing from git fail with an `allowBuilds` error that never mentions this
 * package, so vendor/ is committed instead and merely verified here. Check mode
 * therefore treats the build output as optional: on a machine that only cloned
 * the repository there is no bin/, and the committed overlay is authoritative on
 * its own.
 */
import { execFileSync } from 'node:child_process'
import { createHash } from 'node:crypto'
import { copyFileSync, existsSync, mkdirSync, readFileSync, readdirSync, statSync, writeFileSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const checkOnly = process.argv.includes('--check')

const packageDir = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..')
const repoDir = path.resolve(packageDir, '..', '..')
const vendorDir = path.join(packageDir, 'vendor')

const EXE_SOURCE = path.join(repoDir, 'bin', 'AI-Terminal-Pet.exe')
const PROTOCOL_SOURCE = path.join(repoDir, 'reporters', 'protocol.js')
const EXE_TARGET = path.join(vendorDir, 'AI-Terminal-Pet.exe')
const PROTOCOL_TARGET = path.join(vendorDir, 'protocol.js')
const MANIFEST_TARGET = path.join(vendorDir, 'manifest.json')

const MANIFEST_LABEL = 'vendor/manifest.json'
const SOURCE_ROOTS = ['cmd', 'internal']
// bin/ and vendor/ hold build output, never Go sources; skipping them keeps the
// digest stable whether or not somebody has built the tree.
const SKIP_DIRS = new Set(['bin', 'vendor', 'node_modules', '.git'])

const BUILD_HINT = [
  'scripts/vendor.mjs: the overlay binary is missing at bin/AI-Terminal-Pet.exe.',
  'Build it first, then re-run npm pack:',
  '',
  '  run: go build -trimpath -buildvcs=false -ldflags "-s -w" -o bin\\AI-Terminal-Pet.exe ./cmd/ai-terminal-pet',
  '',
].join('\n')

const CHECK_HINT = [
  'scripts/vendor.mjs --check: vendor/ is incomplete or its manifest is damaged.',
  'vendor/ is committed, so this means the checkout itself is broken.',
  'Restore it with: git checkout -- packages/dsh-ai-terminal-pet/vendor',
  '',
].join('\n')

const STALE_HINT = [
  'scripts/vendor.mjs --check: the vendored overlay predates a change under cmd/ or internal/.',
  'Refresh it with `node packages/dsh-ai-terminal-pet/scripts/vendor.mjs` (copy mode) and commit the regenerated vendor/manifest.json.',
  '',
].join('\n')

function warn(text) {
  process.stderr.write(text + '\n')
}

function sha256(file) {
  return createHash('sha256').update(readFileSync(file)).digest('hex')
}

function report(label, file, action) {
  const bytes = statSync(file).size
  process.stdout.write(label + '  ' + bytes + ' B  sha256 ' + sha256(file).slice(0, 16) + '  (' + action + ')\n')
}

/** The Go release this machine runs, or '' when no toolchain is installed. */
function currentGoVersion() {
  try {
    return execFileSync('go', ['env', 'GOVERSION'], { encoding: 'utf8', stdio: ['ignore', 'pipe', 'ignore'] }).trim()
  } catch {
    return ''
  }
}

function walkGoSources(dir, found) {
  for (const entry of readdirSync(dir, { withFileTypes: true })) {
    const full = path.join(dir, entry.name)
    if (entry.isDirectory()) {
      if (!SKIP_DIRS.has(entry.name.toLowerCase())) walkGoSources(full, found)
    } else if (entry.isFile() && entry.name.endsWith('.go') && !entry.name.endsWith('_test.go')) {
      found.push(path.relative(repoDir, full).split(path.sep).join('/'))
    }
  }
}

/**
 * One digest over every Go file the overlay is built from: a sorted list of
 * `<slash path>\0<sha256 of the file>\n` folded into a single sha256. Sorting
 * makes the digest independent of directory order, and hashing the path as well
 * as the bytes means a rename or a move between packages counts as a change.
 */
function sourceDigest() {
  const found = []
  for (const root of SOURCE_ROOTS) {
    const dir = path.join(repoDir, root)
    if (existsSync(dir)) walkGoSources(dir, found)
  }
  found.sort((left, right) => (left < right ? -1 : left > right ? 1 : 0))
  const fold = createHash('sha256')
  for (const relative of found) {
    fold.update(relative + '\0' + sha256(path.join(repoDir, relative)) + '\n', 'utf8')
  }
  return { files: found.length, sha256: fold.digest('hex') }
}

/** Copy when the destination is missing or differs; true when a write happened. */
function sync(source, target, label) {
  if (!existsSync(source)) {
    warn('scripts/vendor.mjs: missing ' + source)
    return false
  }
  mkdirSync(vendorDir, { recursive: true })
  const same = existsSync(target) && sha256(source) === sha256(target)
  if (!same) copyFileSync(source, target)
  report(label, target, same ? 'unchanged' : 'copied')
  return true
}

/** True when the vendored file is present and non-empty; warns otherwise. */
function presentVendored(target) {
  if (!existsSync(target)) {
    warn('scripts/vendor.mjs --check: missing ' + target)
    return false
  }
  if (statSync(target).size === 0) {
    warn('scripts/vendor.mjs --check: ' + target + ' is empty')
    return false
  }
  return true
}

/** True when the vendored copy is present and, where comparable, still matches. */
function checkVendored(target, source, label) {
  if (!presentVendored(target)) return false
  if (!existsSync(source)) {
    report(label, target, 'checked, no source to compare against')
    return true
  }
  const same = sha256(source) === sha256(target)
  if (!same) {
    warn('scripts/vendor.mjs --check: ' + label + ' differs from ' + source)
    warn('  vendor ' + sha256(target))
    warn('  source ' + sha256(source))
    return false
  }
  report(label, target, 'matches the source')
  return true
}

/** Parse vendor/manifest.json, or null with a warning when it is unusable. */
function readManifest() {
  if (!existsSync(MANIFEST_TARGET)) {
    warn('scripts/vendor.mjs --check: missing ' + MANIFEST_TARGET)
    warn('  it records which sources the committed overlay was built from')
    return null
  }
  let manifest
  try {
    manifest = JSON.parse(readFileSync(MANIFEST_TARGET, 'utf8'))
  } catch (error) {
    warn('scripts/vendor.mjs --check: ' + MANIFEST_TARGET + ' is not valid JSON (' + error.message + ')')
    return null
  }
  if (manifest === null || typeof manifest !== 'object' || manifest.schemaVersion !== 1) {
    warn('scripts/vendor.mjs --check: ' + MANIFEST_LABEL + ' has schemaVersion ' + JSON.stringify(manifest === null || typeof manifest !== 'object' ? undefined : manifest.schemaVersion) + ', expected 1')
    return null
  }
  return manifest
}

/** True when the manifest describes the bytes actually sitting in vendor/. */
function manifestDescribesVendor(manifest) {
  let ok = true
  for (const [entry, target, label] of [[manifest.overlay, EXE_TARGET, 'overlay'], [manifest.protocol, PROTOCOL_TARGET, 'protocol']]) {
    const actualBytes = statSync(target).size
    const actualSha = sha256(target)
    if (entry === null || typeof entry !== 'object' || entry.bytes !== actualBytes || entry.sha256 !== actualSha) {
      warn('scripts/vendor.mjs --check: ' + MANIFEST_LABEL + ' does not describe the ' + label + ' in vendor/')
      warn('  manifest ' + String(entry && entry.sha256) + '  ' + String(entry && entry.bytes) + ' B')
      warn('  vendor   ' + actualSha + '  ' + actualBytes + ' B')
      ok = false
    }
  }
  return ok
}

/** True when the committed overlay was built from the sources as they are now. */
function manifestDescribesSources(manifest) {
  const digest = sourceDigest()
  return manifest.source !== null && typeof manifest.source === 'object' && manifest.source.sha256 === digest.sha256
}

/**
 * The optional byte comparison. Skipped, not failed, when bin/ is absent or when
 * this machine's Go differs from the Go named in the manifest: different releases
 * compile the same sources to different bytes, so the comparison would always
 * shout at a contributor whose toolchain differs from the one that cut the tag.
 */
function compareOverlayBytes(manifest) {
  if (!existsSync(EXE_SOURCE)) {
    process.stdout.write('bin/ is absent, so the committed overlay is authoritative; rebuild it before tagging a release.\n')
    return true
  }
  const current = currentGoVersion()
  if (current !== manifest.goVersion) {
    process.stdout.write('vendor/AI-Terminal-Pet.exe  byte comparison skipped: the committed overlay was built with ' +
      (manifest.goVersion || 'an unknown toolchain') + ', this Go is ' + (current || 'no usable toolchain') +
      '. Different Go releases emit different bytes and the go.mod line is a floor, not a pin; the manifest source digest is the staleness signal.\n')
    return true
  }
  if (sha256(EXE_SOURCE) !== sha256(EXE_TARGET)) {
    warn('scripts/vendor.mjs --check: vendor/AI-Terminal-Pet.exe differs from ' + EXE_SOURCE + ' although both were built with ' + manifest.goVersion)
    warn('  vendor ' + sha256(EXE_TARGET))
    warn('  source ' + sha256(EXE_SOURCE))
    warn('  the source digest matched, so the build is not reproducible; re-run the copy mode from a clean checkout')
    return false
  }
  report('vendor/AI-Terminal-Pet.exe', EXE_TARGET, 'matches the source')
  return true
}

/** Record what was just vendored: toolchain, source digest, both file digests. */
function writeManifest() {
  const digest = sourceDigest()
  const manifest = {
    schemaVersion: 1,
    generatedBy: 'scripts/vendor.mjs',
    goVersion: currentGoVersion(),
    source: { files: digest.files, sha256: digest.sha256 },
    overlay: { file: 'vendor/AI-Terminal-Pet.exe', bytes: statSync(EXE_TARGET).size, sha256: sha256(EXE_TARGET) },
    protocol: { file: 'vendor/protocol.js', bytes: statSync(PROTOCOL_TARGET).size, sha256: sha256(PROTOCOL_TARGET) },
  }
  writeFileSync(MANIFEST_TARGET, JSON.stringify(manifest, null, 2) + '\n', 'utf8')
  process.stdout.write(MANIFEST_LABEL + '  ' + statSync(MANIFEST_TARGET).size + ' B  (written)\n')
}

function copyMode() {
  if (!existsSync(EXE_SOURCE)) {
    warn(BUILD_HINT)
    return 2
  }
  if (!sync(EXE_SOURCE, EXE_TARGET, 'vendor/AI-Terminal-Pet.exe')) return 2
  if (!sync(PROTOCOL_SOURCE, PROTOCOL_TARGET, 'vendor/protocol.js')) return 2
  if (sha256(PROTOCOL_TARGET) !== sha256(PROTOCOL_SOURCE)) {
    warn('scripts/vendor.mjs: vendor/protocol.js is not byte-identical to reporters/protocol.js')
    return 2
  }
  process.stdout.write('vendor/protocol.js matches reporters/protocol.js (sha256 ' + sha256(PROTOCOL_TARGET).slice(0, 16) + ')\n')
  writeManifest()
  return 0
}

function checkMode() {
  // Two very different failures, so they get two different remedies: a damaged
  // vendor/ means the checkout is broken (git restore it), while a stale overlay
  // means the sources moved on and the vendor copy simply needs regenerating.
  let damaged = false
  let stale = false

  const exePresent = presentVendored(EXE_TARGET)
  const protocolPresent = presentVendored(PROTOCOL_TARGET)
  if (!exePresent || !protocolPresent) damaged = true
  if (protocolPresent && !checkVendored(PROTOCOL_TARGET, PROTOCOL_SOURCE, 'vendor/protocol.js')) damaged = true

  const manifest = readManifest()
  if (manifest === null) {
    damaged = true
  } else {
    if (exePresent && protocolPresent && !manifestDescribesVendor(manifest)) damaged = true
    if (!manifestDescribesSources(manifest)) stale = true
  }

  if (stale) {
    const digest = sourceDigest()
    warn(STALE_HINT)
    warn('  manifest source ' + String(manifest.source && manifest.source.sha256) + '  ' + String(manifest.source && manifest.source.files) + ' file(s)')
    warn('  current  source ' + digest.sha256 + '  ' + digest.files + ' file(s)')
  }
  if (damaged) warn(CHECK_HINT)
  if (stale || damaged) return 2

  if (!compareOverlayBytes(manifest)) return 2
  return 0
}

const code = checkOnly ? checkMode() : copyMode()
if (code !== 0) process.exitCode = code
