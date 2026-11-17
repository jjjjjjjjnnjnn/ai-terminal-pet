// The rules of bootstrap.mjs: the panel is dropped into place once, never
// replaced, never trusted without its manifest digest, and never at the cost of
// the host process.
//
//   node --test packages/opencode-ai-terminal-pet/bootstrap.test.mjs
//
// The platform is injected so the Windows logic is testable on Linux and macOS,
// where the real thing returns skipped-platform before touching a file.

import assert from 'node:assert/strict'
import { createHash } from 'node:crypto'
import { existsSync, mkdirSync, mkdtempSync, readFileSync, readdirSync, rmSync, writeFileSync } from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import test from 'node:test'

import { installPanel } from './bootstrap.mjs'

function fixture(t) {
  const root = mkdtempSync(path.join(os.tmpdir(), 'atp-bootstrap-'))
  t.after(() => rmSync(root, { recursive: true, force: true }))
  const source = path.join(root, 'vendor', 'AI-Terminal-Pet.exe')
  mkdirSync(path.dirname(source), { recursive: true })
  writeFileSync(source, 'stub overlay payload')
  const digest = createHash('sha256').update(readFileSync(source)).digest('hex')
  const manifest = path.join(root, 'vendor', 'manifest.json')
  writeFileSync(manifest, JSON.stringify({ schemaVersion: 1, overlay: { sha256: digest } }))
  return { root, source, manifest, target: path.join(root, 'data', 'AI-Terminal-Pet.exe') }
}

test('a missing panel is installed from the vendored copy', (t) => {
  const { source, manifest, target } = fixture(t)
  const result = installPanel({ platform: 'win32', source, manifestPath: manifest, target })
  assert.equal(result.status, 'installed')
  assert.equal(readFileSync(target, 'utf8'), readFileSync(source, 'utf8'))
  // Through a temp sibling: no half-written file is left behind.
  assert.deepEqual(readdirSync(path.dirname(target)), ['AI-Terminal-Pet.exe'])
})

test('an existing panel is never replaced or downgraded', (t) => {
  const { source, manifest, target } = fixture(t)
  mkdirSync(path.dirname(target), { recursive: true })
  writeFileSync(target, 'the newer panel the user installed')
  const result = installPanel({ platform: 'win32', source, manifestPath: manifest, target })
  assert.equal(result.status, 'present')
  assert.equal(readFileSync(target, 'utf8'), 'the newer panel the user installed')
})

test('a vendored copy that does not match its manifest is refused', (t) => {
  const { source, manifest, target } = fixture(t)
  writeFileSync(manifest, JSON.stringify({ schemaVersion: 1, overlay: { sha256: 'not-the-digest' } }))
  const result = installPanel({ platform: 'win32', source, manifestPath: manifest, target })
  assert.equal(result.status, 'failed')
  assert.match(result.reason, /does not match its manifest/)
  assert.equal(existsSync(target), false, 'nothing may be written from an unverified copy')
})

test('no manifest means no installation, not a guess', (t) => {
  const { source, target, root } = fixture(t)
  const result = installPanel({ platform: 'win32', source, manifestPath: path.join(root, 'missing.json'), target })
  assert.equal(result.status, 'failed')
  assert.equal(existsSync(target), false)
})

test('other platforms and the opt-out are skipped before any file is touched', (t) => {
  const { source, manifest, target } = fixture(t)
  assert.equal(installPanel({ platform: 'linux', source, manifestPath: manifest, target }).status, 'skipped-platform')
  assert.equal(installPanel({ platform: 'darwin', source, manifestPath: manifest, target }).status, 'skipped-platform')
  const previous = process.env.AI_TERMINAL_PET_NO_PANEL_INSTALL
  process.env.AI_TERMINAL_PET_NO_PANEL_INSTALL = '1'
  try {
    assert.equal(installPanel({ platform: 'win32', source, manifestPath: manifest, target }).status, 'skipped-env')
  } finally {
    if (previous === undefined) delete process.env.AI_TERMINAL_PET_NO_PANEL_INSTALL
    else process.env.AI_TERMINAL_PET_NO_PANEL_INSTALL = previous
  }
  assert.equal(existsSync(target), false)
})

test('a missing vendored copy is a skip, and a bad target path is a failure that does not throw', (t) => {
  const { manifest, target, root } = fixture(t)
  assert.equal(
    installPanel({ platform: 'win32', source: path.join(root, 'nope.exe'), manifestPath: manifest, target }).status,
    'skipped-unsupported'
  )
  // A file where a directory has to be: the copy must fail, and must not throw
  // into the host's plugin loader.
  const blocked = path.join(root, 'blocked')
  writeFileSync(blocked, 'not a directory')
  const result = installPanel({ platform: 'win32', source: path.join(root, 'vendor', 'AI-Terminal-Pet.exe'), manifestPath: manifest, target: path.join(blocked, 'AI-Terminal-Pet.exe') })
  assert.equal(result.status, 'failed')
  assert.equal(typeof result.reason, 'string')
})
