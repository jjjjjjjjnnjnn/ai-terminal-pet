/**
 * Rules for the upstream entry sync, pinned so a future edit cannot quietly
 * break them. No network, no git: everything here is a pure function from
 * tools/upstream-entry.mjs.
 *
 * Run with: node --test tools/upstream-entry.test.mjs
 */
import assert from 'node:assert/strict'
import { test } from 'node:test'

import {
  assetNameFor,
  entryFileName,
  entryIsStale,
  expectedTarball,
  listedVersion,
  rebuildVerdict,
  rewriteTarball,
  tarballLine,
  tarballValue,
} from './upstream-entry.mjs'

/** A stand-in for the upstream file, in the shape contributing.md prescribes. */
const ENTRY = `# One entry per file; the generated READMEs are not edited by hand.
url: https://github.com/jjjjjjjjnnjnn/ai-terminal-pet/tree/main/packages/dsh-ai-terminal-pet
name: jjjjjjjjnnjnn/ai-terminal-pet#dsh-ai-terminal-pet
category: notify
tarball: https://github.com/jjjjjjjjnnjnn/ai-terminal-pet/releases/download/dsh-plugin-v1.2.7/dsh-ai-terminal-pet-1.2.7.tgz
icon: https://raw.githubusercontent.com/jjjjjjjjnnjnn/ai-terminal-pet/main/packages/dsh-ai-terminal-pet/icon.png
description:
  en: 'A floating desktop overlay for Windows. Click the arrow at the end of a row to jump to that terminal.'
  zh: 'Windows 上的桌面浮窗。点一下某一行的行尾箭头即可跳到对应终端。'
`

test('the entry file name follows the monorepo convention', () => {
  assert.equal(
    entryFileName('jjjjjjjjnnjnn', 'ai-terminal-pet', 'packages/dsh-ai-terminal-pet'),
    'jjjjjjjjnnjnn__ai-terminal-pet--packages-dsh-ai-terminal-pet.yml',
  )
  // A repository-root plugin has no `--<path>` suffix.
  assert.equal(entryFileName('owner', 'repo', ''), 'owner__repo.yml')
})

test('the tarball URL is built from the release tag and the asset name', () => {
  const asset = assetNameFor('dsh-ai-terminal-pet', '1.2.9')
  assert.equal(asset, 'dsh-ai-terminal-pet-1.2.9.tgz')
  assert.equal(
    expectedTarball({ owner: 'jjjjjjjjnnjnn', repo: 'ai-terminal-pet', version: '1.2.9', assetName: asset }),
    'https://github.com/jjjjjjjjnnjnn/ai-terminal-pet/releases/download/dsh-plugin-v1.2.9/dsh-ai-terminal-pet-1.2.9.tgz',
  )
})

test('the listed version is read from the tarball line, whatever else the file says', () => {
  assert.equal(listedVersion(ENTRY), '1.2.7')
  // A description that happens to contain a version number must not win.
  const decoy = ENTRY.replace('dsh-plugin-v1.2.7', 'dsh-plugin-v9.9.9')
    .replace('category: notify', 'category: notify\n# tarball: dsh-plugin-v1.2.6')
  assert.equal(listedVersion(decoy), '9.9.9')
  // No tarball at all is "no version", not a crash.
  assert.equal(listedVersion('# nothing here'), '')
})

test('staleness is a plain inequality, and "no tarball" counts as stale', () => {
  assert.equal(entryIsStale('1.2.6', '1.2.8'), true)
  assert.equal(entryIsStale('1.2.8', '1.2.8'), false)
  assert.equal(entryIsStale('', '1.2.8'), true)
})

test('a rewrite touches one line and is exactly reversible', () => {
  const url = expectedTarball({
    owner: 'jjjjjjjjnnjnn', repo: 'ai-terminal-pet', version: '1.2.8',
    assetName: assetNameFor('dsh-ai-terminal-pet', '1.2.8'),
  })
  const after = rewriteTarball(ENTRY, url)
  assert.notEqual(after, ENTRY)
  assert.equal(listedVersion(after), '1.2.8')
  // Same byte count, LF only, one changed line: the properties the real push
  // asserts before it is allowed to happen.
  assert.equal(after.length, ENTRY.length + (url.length - tarballValue(ENTRY).length))
  assert.equal(after.split('\n').filter((line) => line.startsWith('tarball:')).length, 1)
  assert.equal(after.split('\n').filter((line, index) => line !== ENTRY.split('\n')[index]).length, 1)
  // Reverting proves nothing else moved.
  assert.equal(rewriteTarball(after, tarballValue(ENTRY)), ENTRY)
})

test('a rewrite refuses every shape that could corrupt the entry', () => {
  // Two tarball keys: YAML would silently keep the last one, so this is not
  // "harmless duplication".
  assert.throws(
    () => rewriteTarball(ENTRY + 'tarball: https://example.com/x.tgz\n', 'https://example.com/new.tgz'),
    /exactly one tarball: line, found 2/,
  )
  // Zero tarball keys: nothing to replace, and inventing one would be a
  // different change than the bump the maintainer was asked to review.
  assert.throws(() => rewriteTarball('# empty\n', 'https://example.com/x.tgz'), /found 0/)
  // A value that would need YAML quoting.
  assert.throws(() => rewriteTarball(ENTRY, "https://example.com/a'b.tgz"), /quoting/)
  assert.throws(() => rewriteTarball(ENTRY, 'https://example.com/a#b.tgz'), /quoting/)
})

test('the rebuild verdict only allows a tarball-only change', () => {
  const url = expectedTarball({
    owner: 'jjjjjjjjnnjnn', repo: 'ai-terminal-pet', version: '1.2.8',
    assetName: assetNameFor('dsh-ai-terminal-pet', '1.2.8'),
  })
  const clean = rewriteTarball(ENTRY, url)
  assert.deepEqual(rebuildVerdict({ before: ENTRY, after: clean }), { changed: true, onlyTarball: true })

  // Something else moved as well (a maintainer's edit we did not intend to
  // carry): the revert no longer reproduces upstream, so the push must be
  // refused.
  const dirty = clean.replace('category: notify', 'category: ui')
  assert.deepEqual(rebuildVerdict({ before: ENTRY, after: dirty }), { changed: true, onlyTarball: false })

  // Nothing changed at all: not stale, not a rebuild.
  assert.deepEqual(rebuildVerdict({ before: ENTRY, after: ENTRY }), { changed: false, onlyTarball: false })

  // The tarball line was removed rather than replaced.
  const removed = ENTRY.replace(tarballLine(ENTRY) + '\n', '')
  assert.deepEqual(rebuildVerdict({ before: ENTRY, after: removed }), { changed: true, onlyTarball: false })
})

test('a CRLF copy is refused rather than rewritten', () => {
  // The tool clones with autocrlf=false so this cannot happen in practice; if a
  // CRLF file ever reached it, the "revert to prove nothing else moved" proof
  // would be meaningless, so it must fail loudly instead of pushing.
  const crlf = ENTRY.split('\n').join('\r\n')
  assert.throws(() => rewriteTarball(crlf, 'https://example.com/x.tgz'), /CRLF/)
  assert.deepEqual(rebuildVerdict({ before: crlf, after: rewriteTarball(ENTRY, 'https://example.com/x.tgz') }),
    { changed: true, onlyTarball: false })
})
