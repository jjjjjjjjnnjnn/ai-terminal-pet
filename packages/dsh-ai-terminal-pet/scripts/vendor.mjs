/**
 * Vendor, or check, the two files this package ships.
 *
 *   node scripts/vendor.mjs          # copy ../../bin + ../../reporters into vendor/
 *   node scripts/vendor.mjs --check  # verify what is in vendor/, write nothing
 *
 * Copies ../../bin/AI-Terminal-Pet.exe -> vendor/AI-Terminal-Pet.exe
 * and     ../../reporters/protocol.js   -> vendor/protocol.js
 *
 * The protocol copy is checked byte-for-byte against its source, because the Go
 * server and the reporters must agree on the wire contract; a drifting copy
 * would silently break every report.
 *
 * Copy mode is idempotent: a destination that already matches is left alone.
 *
 * Check mode is what `npm run verify` runs, and it exists because this package
 * deliberately has no `prepack` hook. A pnpm `prepack`/`prepare` step makes
 * installing from git fail with an `allowBuilds` error that never mentions this
 * package, so vendor/ is committed instead and merely verified here. Check mode
 * therefore treats the build output as optional: on a machine that only cloned
 * the repository there is no bin/, and the committed overlay is authoritative on
 * its own. It is only compared with bin/ when bin/ happens to be there, which is
 * how a stale vendor copy gets caught before a release is tagged.
 */
import { createHash } from 'node:crypto'
import { copyFileSync, existsSync, mkdirSync, readFileSync, statSync } from 'node:fs'
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

const BUILD_HINT = [
  'scripts/vendor.mjs: the overlay binary is missing at bin/AI-Terminal-Pet.exe.',
  'Build it first, then re-run npm pack:',
  '',
  '  run: go build -trimpath -buildvcs=false -ldflags "-s -w" -o bin\\AI-Terminal-Pet.exe ./cmd/ai-terminal-pet',
  '',
].join('\n')

const CHECK_HINT = [
  'scripts/vendor.mjs --check: vendor/ is incomplete.',
  'vendor/ is committed, so this means the checkout itself is broken.',
  'Restore it with: git checkout -- packages/dsh-ai-terminal-pet/vendor',
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

/** True when the vendored copy is present and, where comparable, still matches. */
function checkVendored(target, source, label) {
  if (!existsSync(target)) {
    warn('scripts/vendor.mjs --check: missing ' + target)
    return false
  }
  if (statSync(target).size === 0) {
    warn('scripts/vendor.mjs --check: ' + target + ' is empty')
    return false
  }
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
  return 0
}

function checkMode() {
  let ok = true
  if (!checkVendored(EXE_TARGET, EXE_SOURCE, 'vendor/AI-Terminal-Pet.exe')) ok = false
  if (!checkVendored(PROTOCOL_TARGET, PROTOCOL_SOURCE, 'vendor/protocol.js')) ok = false
  if (!ok) {
    warn(CHECK_HINT)
    return 2
  }
  if (!existsSync(EXE_SOURCE)) {
    process.stdout.write('bin/ is absent, so the committed overlay is authoritative; rebuild it before tagging a release.\n')
  }
  return 0
}

const code = checkOnly ? checkMode() : copyMode()
if (code !== 0) process.exitCode = code