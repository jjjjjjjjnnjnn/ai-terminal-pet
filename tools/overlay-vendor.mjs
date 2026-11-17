#!/usr/bin/env node
/**
 * The two facts both npm packages need about the overlay they ship: what its
 * source digest is, and what the binary's own hash is.
 *
 * Why this is shared
 * ------------------
 * Two packages now vendor the same `bin/AI-Terminal-Pet.exe`:
 * `packages/dsh-ai-terminal-pet` (Cordis bundle) and
 * `packages/opencode-ai-terminal-pet` (OpenCode plugin). Each writes its own
 * `vendor/manifest.json`, but "is this binary older than the sources it was
 * built from?" must have exactly one answer, or the two packages disagree about
 * staleness and one of them ships a stale panel.
 *
 * The digest is not a byte comparison, and the reason is measured: `go.mod`
 * declares `go 1.23.0` as a *floor*, so CI (go1.23.0) and a developer machine
 * (go1.27.0) compile identical sources into different bytes. Byte equality
 * across toolchains is unattainable and must never gate anything. What can be
 * answered is "do the sources under cmd/ and internal/ still hash to the same
 * value?", which is what `sourceDigest` computes: a sorted list of
 * `<slash path>\0<sha256 of the file>\n` folded into one sha256.
 *
 * `NON_OVERLAY_PACKAGES` keeps that question honest: the overlay links neither
 * `cmd/pet-hook`, `cmd/pet-statusline`, `cmd/pet-setup` nor `internal/reporter`,
 * so an edit to any of them cannot change a byte of the panel - and demanding a
 * re-vendor for it would mean re-writing a released package's manifest, which
 * immutability forbids. Add new non-overlay packages here; the list is explicit
 * so that a new package shows up as a digest change and a reviewer decides.
 *
 * Usage:
 *   import { sourceDigest, sha256File, goVersion } from '../../tools/overlay-vendor.mjs'
 */
import { execFileSync } from 'node:child_process'
import { createHash } from 'node:crypto'
import { existsSync, readFileSync, readdirSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

export const repoRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..')

export const SOURCE_ROOTS = ['cmd', 'internal']

/** Directories that hold build output, never Go sources. */
export const SKIP_DIRS = new Set(['bin', 'vendor', 'node_modules', '.git'])

/** Packages the overlay binary does not link; see the header. */
export const NON_OVERLAY_PACKAGES = [
  'cmd/pet-hook/',
  'cmd/pet-setup/',
  'cmd/pet-statusline/',
  'internal/reporter/'
]

export function sha256File(file) {
  return createHash('sha256').update(readFileSync(file)).digest('hex')
}

/** The Go release this machine runs, or '' when no toolchain is installed. */
export function goVersion() {
  try {
    return execFileSync('go', ['env', 'GOVERSION'], { encoding: 'utf8', stdio: ['ignore', 'pipe', 'ignore'] }).trim()
  } catch {
    return ''
  }
}

function walkGoSources(dir, found, root) {
  for (const entry of readdirSync(dir, { withFileTypes: true })) {
    const full = path.join(dir, entry.name)
    if (entry.isDirectory()) {
      if (!SKIP_DIRS.has(entry.name.toLowerCase())) walkGoSources(full, found, root)
    } else if (entry.isFile() && entry.name.endsWith('.go') && !entry.name.endsWith('_test.go')) {
      found.push(path.relative(root, full).split(path.sep).join('/'))
    }
  }
}

function isOverlaySource(relative) {
  return !NON_OVERLAY_PACKAGES.some((prefix) => relative.startsWith(prefix))
}

/**
 * sourceDigest returns { files, sha256 } over every Go file the overlay links.
 * `root` is the repository; tests pass a fixture tree.
 */
export function sourceDigest(root = repoRoot) {
  const found = []
  for (const tree of SOURCE_ROOTS) {
    const dir = path.join(root, tree)
    if (!existsSync(dir)) continue
    const collected = []
    walkGoSources(dir, collected, root)
    for (const relative of collected) if (isOverlaySource(relative)) found.push(relative)
  }
  found.sort((left, right) => (left < right ? -1 : left > right ? 1 : 0))
  const fold = createHash('sha256')
  for (const relative of found) {
    fold.update(relative + '\0' + sha256File(path.join(root, relative)) + '\n', 'utf8')
  }
  return { files: found.length, sha256: fold.digest('hex') }
}
