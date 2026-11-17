#!/usr/bin/env node
/**
 * Vendor, or check, the overlay binary this package drops into place on first run.
 *
 *   node scripts/vendor.mjs          # copy ../../bin/AI-Terminal-Pet.exe into vendor/
 *   node scripts/vendor.mjs --check  # verify what is in vendor/, write nothing
 *
 * Why this package ships the panel
 * --------------------------------
 * The first version of this package was a thin reporter: `opencode plugin add`
 * wired the reporting side and told the user to install the panel separately.
 * Measured against what a user actually does, that is a dead end - the plugin
 * loads, reports, and nothing appears on screen, because
 * `protocol.launchApplication()` refuses to start a binary that is not there.
 * One command has to mean one command.
 *
 * The DSH bundle made the same call the same way: vendor the Windows binary
 * (offline, no SmartScreen, no download), and fetch a pinned archive only where
 * a bundled copy cannot work (Linux/macOS - not in this package yet; those users
 * install from the release tarball and the README says so).
 *
 * Staleness is a SOURCE DIGEST, not a byte comparison, and the reasoning is in
 * tools/overlay-vendor.mjs: `go.mod`'s `go 1.23.0` is a floor, so CI and a
 * developer machine compile the same sources to different bytes. The shared
 * helper makes this package and the DSH bundle reach the same verdict.
 */
import { copyFileSync, existsSync, mkdirSync, readFileSync, statSync, writeFileSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

import { goVersion, sha256File, sourceDigest } from '../../../tools/overlay-vendor.mjs'

const checkOnly = process.argv.includes('--check')
const packageDir = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..')
const repoDir = path.resolve(packageDir, '..', '..')

const EXE_SOURCE = path.join(repoDir, 'bin', 'AI-Terminal-Pet.exe')
const VENDOR_DIR = path.join(packageDir, 'vendor')
const EXE_TARGET = path.join(VENDOR_DIR, 'AI-Terminal-Pet.exe')
const MANIFEST_TARGET = path.join(VENDOR_DIR, 'manifest.json')

const BUILD_HINT = [
  'scripts/vendor.mjs: the overlay binary is missing at bin/AI-Terminal-Pet.exe.',
  'Build it first, then re-run this script:',
  '',
  '  go build -trimpath -buildvcs=false -ldflags "-s -w" -o bin\\AI-Terminal-Pet.exe ./cmd/ai-terminal-pet',
  '',
].join('\n')

function warn(text) {
  process.stderr.write(text + '\n')
}

function readManifest() {
  try {
    const manifest = JSON.parse(readFileSync(MANIFEST_TARGET, 'utf8'))
    if (manifest === null || typeof manifest !== 'object' || manifest.schemaVersion !== 1) return null
    return manifest
  } catch {
    return null
  }
}

function copyMode() {
  if (!existsSync(EXE_SOURCE)) {
    warn(BUILD_HINT)
    return 2
  }
  mkdirSync(VENDOR_DIR, { recursive: true })
  const same = existsSync(EXE_TARGET) && sha256File(EXE_SOURCE) === sha256File(EXE_TARGET)
  if (!same) copyFileSync(EXE_SOURCE, EXE_TARGET)
  const digest = sourceDigest(repoDir)
  const manifest = {
    schemaVersion: 1,
    generatedBy: 'scripts/vendor.mjs',
    goVersion: goVersion(),
    source: { files: digest.files, sha256: digest.sha256 },
    overlay: { file: 'vendor/AI-Terminal-Pet.exe', bytes: statSync(EXE_TARGET).size, sha256: sha256File(EXE_TARGET) }
  }
  writeFileSync(MANIFEST_TARGET, JSON.stringify(manifest, null, 2) + '\n', 'utf8')
  process.stdout.write(
    'vendor/AI-Terminal-Pet.exe  ' + manifest.overlay.bytes + ' B  sha256 ' + manifest.overlay.sha256.slice(0, 16) +
      '  (' + (same ? 'unchanged' : 'copied') + ')\n'
  )
  process.stdout.write('vendor/manifest.json  ' + statSync(MANIFEST_TARGET).size + ' B  (written)\n')
  return 0
}

function checkMode() {
  let failed = false
  if (!existsSync(EXE_TARGET) || statSync(EXE_TARGET).size === 0) {
    warn('scripts/vendor.mjs --check: missing or empty vendor/AI-Terminal-Pet.exe')
    failed = true
  }
  const manifest = readManifest()
  if (!manifest) {
    warn('scripts/vendor.mjs --check: vendor/manifest.json is missing or not schemaVersion 1')
    return 2
  }
  if (!failed) {
    const actual = { bytes: statSync(EXE_TARGET).size, sha256: sha256File(EXE_TARGET) }
    if (manifest.overlay.bytes !== actual.bytes || manifest.overlay.sha256 !== actual.sha256) {
      warn('scripts/vendor.mjs --check: vendor/manifest.json does not describe vendor/AI-Terminal-Pet.exe')
      failed = true
    }
  }
  // The staleness question the two packages must agree on.
  const digest = sourceDigest(repoDir)
  if (manifest.source.sha256 !== digest.sha256) {
    warn('scripts/vendor.mjs --check: the vendored overlay predates a change to a Go file it is built from.')
    warn('  manifest ' + manifest.source.sha256 + '  computed ' + digest.sha256)
    warn('  Refresh it with `node packages/opencode-ai-terminal-pet/scripts/vendor.mjs` and commit vendor/.')
    failed = true
  }
  // Byte comparison only when this machine's Go matches the one that cut the
  // binary; different releases emit different bytes (see the header).
  if (!failed && existsSync(EXE_SOURCE) && goVersion() === manifest.goVersion) {
    if (sha256File(EXE_SOURCE) !== sha256File(EXE_TARGET)) {
      warn('scripts/vendor.mjs --check: vendor/AI-Terminal-Pet.exe differs from bin/ although both report ' + manifest.goVersion)
      failed = true
    }
  }
  if (!failed) process.stdout.write('vendor/AI-Terminal-Pet.exe matches the sources it was built from\n')
  return failed ? 1 : 0
}

process.exitCode = checkOnly ? checkMode() : copyMode()
