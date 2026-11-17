// The rules of scripts/sync.mjs, and the one that matters most: the package's
// copies still match the reference reporter in reporters/.
//
//   node --test packages/opencode-ai-terminal-pet/scripts/sync.test.mjs
//
// A generator nobody checks is a second implementation with extra steps. This
// file is what makes "generated" mean "provably the same bytes as the source".

import assert from 'node:assert/strict'
import { mkdtempSync, readFileSync, rmSync, writeFileSync } from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import test from 'node:test'

import {
  BANNER,
  COPIES,
  REPLACEMENT,
  SERVER,
  SPECIFIER,
  compare,
  packageDir,
  renderServer,
  write
} from './sync.mjs'

const SOURCE = 'const core = require("./protocol.js")\nconsole.log(core)\n'

test('the only edited line is the protocol specifier', () => {
  const rendered = renderServer(SOURCE)
  assert.ok(rendered.startsWith(`${BANNER}\n`), 'the generated file says where it came from')
  const lines = rendered.split('\n')
  const original = SOURCE.split('\n')
  // Banner + source, one for one, so a hand edit cannot hide in a shifted line.
  assert.equal(lines.length, original.length + BANNER.split('\n').length)
  const body = lines.slice(BANNER.split('\n').length)
  assert.deepEqual(body, SOURCE.replace(SPECIFIER, REPLACEMENT).split('\n'))
  // Byte for byte apart from the one specifier: 19 chars become 20.
  assert.equal(rendered.length - (BANNER.length + 1) - SOURCE.length, REPLACEMENT.length - SPECIFIER.length)
})

test('a source shape this tool does not recognise is refused, not guessed at', () => {
  assert.throws(() => renderServer('const core = require("./elsewhere.js")\n'), /exactly one/, 'no require of the core')
  assert.throws(
    () => renderServer(`a(${SPECIFIER}) b(${SPECIFIER})`),
    /exactly one/,
    'two requires: which one moves is a question, not an inference'
  )
  assert.throws(
    () => renderServer(`a()\nb(${REPLACEMENT})`),
    /already mentions/,
    'a reference copy that already says .cjs would be generated twice'
  )
})

test('a missing or edited generated file is reported by name', (t) => {
  const dir = mkdtempSync(path.join(os.tmpdir(), 'atp-sync-'))
  t.after(() => rmSync(dir, { recursive: true, force: true }))

  const written = write(dir)
  assert.deepEqual(written.sort(), [SERVER.target, ...COPIES.map((copy) => copy.target)].sort())
  assert.deepEqual(compare(dir), [], 'a fresh write is clean')

  const target = path.join(dir, SERVER.target)
  writeFileSync(target, `${readFileSync(target, 'utf8')}// hand edit\n`)
  const problems = compare(dir)
  assert.equal(problems.length, 1)
  assert.match(problems[0], /server\.js: differs from reporters\/opencode\.js/)

  rmSync(target)
  assert.match(compare(dir)[0], /server\.js: missing/)
})

test('the package in this checkout is in sync with reporters/', () => {
  // This is the gate. It fails exactly when somebody edits reporters/opencode.js
  // or reporters/protocol.js and forgets to run scripts/sync.mjs.
  assert.deepEqual(compare(packageDir), [])
})
