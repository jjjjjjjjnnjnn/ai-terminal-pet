#!/usr/bin/env node
/**
 * Vendor, or check, the overlay binary this plugin drops into place on first run.
 *
 *   node scripts/vendor.mjs          # copy ../../bin/AI-Terminal-Pet.exe into vendor/
 *   node scripts/vendor.mjs --check  # verify what is in vendor/, write nothing
 *
 * Why this package ships the panel
 * --------------------------------
 * A Claude Code plugin can carry the whole install: the hook side is a native
 * binary (vendor/pet-hook.exe), and SessionStart copies the panel out of
 * vendor/ into %LOCALAPPDATA%\AI-Terminal-Pet\ once. Without the panel the
 * plugin reports into a pipe nobody is reading, which looks exactly like a
 * broken install.
 *
 * Staleness is a SOURCE DIGEST, not a byte comparison, for the reason in
 * tools/overlay-vendor.mjs: `go.mod`'s `go 1.23.0` is a floor, so CI and a
 * developer machine compile the same sources to different bytes. The shared
 * helper makes all three vendoring packages reach the same verdict.
 */
import { copyFileSync, existsSync, mkdirSync, readFileSync, statSync, writeFileSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

import { goVersion, sha256File, sourceDigest } from '../../../tools/overlay-vendor.mjs'

const checkOnly = process.argv.includes('--check')
const packageDir = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..')
const repoDir = path.resolve(packageDir, '..', '..')

const EXE_SOURCE = path.join(repoDir, 'bin', 'AI-Terminal-Pet.exe')
const HOOK_SOURCE = path.join(repoDir, 'bin', 'pet-hook.exe')
const VENDOR_DIR = path.join(packageDir, 'vendor')
const EXE_TARGET = path.join(VENDOR_DIR, 'AI-Terminal-Pet.exe')
const HOOK_TARGET = path.join(VENDOR_DIR, 'pet-hook.exe')
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
  if (!existsSync(HOOK_SOURCE)) {
    warn(BUILD_HINT.replace('AI-Terminal-Pet.exe ./cmd/ai-terminal-pet', 'pet-hook.exe ./cmd/pet-hook'))
    return 2
  }
  mkdirSync(VENDOR_DIR, { recursive: true })
  const same = existsSync(EXE_TARGET) && sha256File(EXE_SOURCE) === sha256File(EXE_TARGET)
  if (!same) copyFileSync(EXE_SOURCE, EXE_TARGET)
  const sameHook = existsSync(HOOK_TARGET) && sha256File(HOOK_SOURCE) === sha256File(HOOK_TARGET)
  if (!sameHook) copyFileSync(HOOK_SOURCE, HOOK_TARGET)
  const digest = sourceDigest(repoDir)
  const manifest = {
    schemaVersion: 1,
    generatedBy: 'scripts/vendor.mjs',
    goVersion: goVersion(),
    source: { files: digest.files, sha256: digest.sha256 },
    overlay: { file: 'vendor/AI-Terminal-Pet.exe', bytes: statSync(EXE_TARGET).size, sha256: sha256File(EXE_TARGET) },
    hook: { file: 'vendor/pet-hook.exe', bytes: statSync(HOOK_TARGET).size, sha256: sha256File(HOOK_TARGET) }
  }
  writeFileSync(MANIFEST_TARGET, JSON.stringify(manifest, null, 2) + '\n', 'utf8')
  process.stdout.write(
    'vendor/AI-Terminal-Pet.exe  ' + manifest.overlay.bytes + ' B  sha256 ' + manifest.overlay.sha256.slice(0, 16) +
      '  (' + (same ? 'unchanged' : 'copied') + ')\n'
  )
  process.stdout.write(
    'vendor/pet-hook.exe  ' + manifest.hook.bytes + ' B  sha256 ' + manifest.hook.sha256.slice(0, 16) +
      '  (' + (sameHook ? 'unchanged' : 'copied') + ')\n'
  )
  process.stdout.write('vendor/manifest.json  ' + statSync(MANIFEST_TARGET).size + ' B  (written)\n')
  return 0
}

function checkMode() {
  let failed = false
  for (const [label, target, source] of [
    ['vendor/AI-Terminal-Pet.exe', EXE_TARGET, EXE_SOURCE],
    ['vendor/pet-hook.exe', HOOK_TARGET, HOOK_SOURCE]
  ]) {
    if (!existsSync(target) || statSync(target).size === 0) {
      warn('scripts/vendor.mjs --check: missing or empty ' + label)
      failed = true
    }
  }
  const manifest = readManifest()
  if (!manifest) {
    warn('scripts/vendor.mjs --check: vendor/manifest.json is missing or not schemaVersion 1')
    return 2
  }
  if (!failed) {
    for (const [key, target] of [['overlay', EXE_TARGET], ['hook', HOOK_TARGET]]) {
      const actual = { bytes: statSync(target).size, sha256: sha256File(target) }
      const recorded = manifest[key]
      if (!recorded || recorded.bytes !== actual.bytes || recorded.sha256 !== actual.sha256) {
        warn('scripts/vendor.mjs --check: vendor/manifest.json does not describe ' + target)
        failed = true
      }
    }
  }
  // The staleness question the three packages must agree on.
  const digest = sourceDigest(repoDir)
  if (manifest.source.sha256 !== digest.sha256) {
    warn('scripts/vendor.mjs --check: the vendored overlay predates a change to a Go file it is built from.')
    warn('  manifest ' + manifest.source.sha256 + '  computed ' + digest.sha256)
    warn('  Refresh it with `node packages/claude-ai-terminal-pet/scripts/vendor.mjs` and commit vendor/.')
    failed = true
  }
  // Byte comparison only when this machine's Go matches the one that cut the
  // binaries; different releases emit different bytes (see the header).
  if (!failed && goVersion() === manifest.goVersion) {
    for (const [target, source] of [[EXE_TARGET, EXE_SOURCE], [HOOK_TARGET, HOOK_SOURCE]]) {
      if (existsSync(source) && sha256File(source) !== sha256File(target)) {
        warn('scripts/vendor.mjs --check: ' + target + ' differs from bin/ although both report ' + manifest.goVersion)
        failed = true
      }
    }
  }
  if (!failed) process.stdout.write('vendor/ matches the sources it was built from\n')
  return failed ? 1 : 0
}

process.exitCode = checkOnly ? checkMode() : copyMode()
