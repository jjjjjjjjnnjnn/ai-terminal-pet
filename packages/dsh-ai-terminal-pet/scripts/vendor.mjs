/**
 * Vendor the two files this package ships, right before npm packs it.
 *
 *   node scripts/vendor.mjs
 *
 * Copies ../../bin/AI-Terminal-Pet.exe    -> vendor/AI-Terminal-Pet.exe
 * and     ../../reporters/protocol.js     -> vendor/protocol.js
 *
 * The protocol copy is checked byte-for-byte against its source, because the
 * Go server and the reporters must agree on the wire contract; a drifting
 * copy would silently break every report.
 * Idempotent: a destination that already matches is left alone.
 */
import { createHash } from 'node:crypto'
import { copyFileSync, existsSync, mkdirSync, readFileSync, statSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

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

function sha256(file) {
  return createHash('sha256').update(readFileSync(file)).digest('hex')
}

function report(label, file, action) {
  const bytes = statSync(file).size
  const digest = sha256(file)
  process.stdout.write(label + '  ' + bytes + ' B  sha256 ' + digest.slice(0, 16) + '  (' + action + ')\n')
}

/** Copy when the destination is missing or differs; true when a write happened. */
function sync(source, target, label) {
  if (!existsSync(source)) {
    process.stderr.write('scripts/vendor.mjs: missing ' + source + '\n')
    return false
  }
  mkdirSync(vendorDir, { recursive: true })
  const same = existsSync(target) && sha256(source) === sha256(target)
  if (!same) copyFileSync(source, target)
  report(label, target, same ? 'unchanged' : 'copied')
  return true
}

if (!existsSync(EXE_SOURCE)) {
  process.stderr.write(BUILD_HINT)
  process.exitCode = 2
} else if (sync(EXE_SOURCE, EXE_TARGET, 'vendor/AI-Terminal-Pet.exe') === false) {
  process.exitCode = 2
} else {
  if (!sync(PROTOCOL_SOURCE, PROTOCOL_TARGET, 'vendor/protocol.js')) process.exitCode = 2
  else {
    const vendorDigest = sha256(PROTOCOL_TARGET)
    const sourceDigest = sha256(PROTOCOL_SOURCE)
    if (vendorDigest !== sourceDigest) {
      process.stderr.write('scripts/vendor.mjs: vendor/protocol.js is not byte-identical to reporters/protocol.js\n')
      process.exitCode = 2
    } else {
      process.stdout.write('vendor/protocol.js matches reporters/protocol.js (sha256 ' + vendorDigest.slice(0, 16) + ')\n')
    }
  }
}
