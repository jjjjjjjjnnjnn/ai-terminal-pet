import { test } from 'node:test'
import assert from 'node:assert/strict'

import {
  CLAIMS,
  check,
  checkChangelog,
  checkClaims,
  compareVersions,
  extractOverlayVersion,
  stripCode,
  sweepFutureVersions
} from './check_versions.mjs'

const versions = { overlay: '1.0.7', plugin: '1.2.14' }

// ------------------------------------------------------------ version source

test('the overlay version comes from the one const Version line', () => {
  const source = 'package app\n\nconst Version = "1.0.7"\n\nfunc Run() error {}\n'
  assert.deepEqual(extractOverlayVersion(source), { version: '1.0.7' })
})

test('a missing const Version is an error, not a silent pass', () => {
  const result = extractOverlayVersion('const Version = Version\n')
  assert.match(result.error, /expected exactly 1/)
})

test('two const Version lines are an error', () => {
  const source = 'const Version = "1.0.7"\nconst Version = "1.0.8"\n'
  assert.match(extractOverlayVersion(source).error, /found 2/)
})

test('compareVersions orders numerically, not lexically', () => {
  assert.equal(compareVersions('1.0.7', '1.0.7'), 0)
  assert.equal(compareVersions('1.0.7', '1.0.10'), -1)
  assert.equal(compareVersions('1.0.10', '1.0.9'), 1)
  assert.equal(compareVersions('1.2.0', '1.0.99'), 1)
})

// ------------------------------------------------------------------- claims

const claim = {
  id: 'test claim',
  file: 'doc.md',
  pattern: /version-(\d+\.\d+\.\d+)/g,
  expect: ['overlay'],
  count: 1
}

test('a matching claim passes', () => {
  assert.deepEqual(checkClaims({ 'doc.md': 'the badge is version-1.0.7' }, versions, [claim]), [])
})

test('a stale value fails, with the file and line', () => {
  const failures = checkClaims({ 'doc.md': 'skip\nskip\nthe badge is version-1.0.2' }, versions, [claim])
  assert.equal(failures.length, 1)
  assert.match(failures[0], /doc\.md:3 says 1\.0\.2, the overlay version is 1\.0\.7/)
})

test('a missing shape fails instead of passing quietly', () => {
  const failures = checkClaims({ 'doc.md': 'the sentence was reworded' }, versions, [claim])
  assert.equal(failures.length, 1)
  assert.match(failures[0], /found 0 match\(es\).*expected 1/)
  assert.match(failures[0], /update CLAIMS/)
})

test('the expected count is exact: a dropped English copy cannot pass', () => {
  const twice = { ...claim, count: 2 }
  const failures = checkClaims({ 'doc.md': 'version-1.0.7' }, versions, [twice])
  assert.equal(failures.length, 1)
  assert.match(failures[0], /found 1 match\(es\).*expected 2/)
})

test('one group per expected kind, both compared', () => {
  const pair = {
    id: 'install pair',
    file: 'doc.md',
    pattern: /dsh-plugin-v(\d+\.\d+\.\d+)\/dsh-ai-terminal-pet-(\d+\.\d+\.\d+)\.tgz/g,
    expect: ['plugin', 'plugin'],
    count: 1
  }
  assert.deepEqual(
    checkClaims({ 'doc.md': 'dsh-plugin-v1.2.14/dsh-ai-terminal-pet-1.2.14.tgz' }, versions, [pair]),
    []
  )
  const failures = checkClaims({ 'doc.md': 'dsh-plugin-v1.2.13/dsh-ai-terminal-pet-1.2.13.tgz' }, versions, [pair])
  assert.equal(failures.length, 2)
})

test('an unread document fails rather than being skipped', () => {
  const failures = checkClaims({}, versions, [claim])
  assert.equal(failures.length, 1)
  assert.match(failures[0], /doc\.md was not read/)
})

test('every shipped claim names a file the checker reads and one or two kinds', () => {
  for (const entry of CLAIMS) {
    assert.ok(entry.id, 'a claim without an id cannot be reported')
    assert.match(entry.file, /\.(md|json)$/, `${entry.id} must point at a document`)
    assert.ok(entry.count >= 1, `${entry.id} must expect at least one occurrence`)
    assert.ok(
      entry.expect.length >= 1 && entry.expect.every((kind) => kind === 'overlay' || kind === 'plugin'),
      `${entry.id} must expect overlay/plugin kinds`
    )
    assert.ok(entry.pattern.global, `${entry.id} needs the g flag or matchAll throws`)
  }
})

// ------------------------------------------------------- future-version sweep

test('history is allowed, a future version is not', () => {
  const doc = { 'doc.md': 'shipped 1.0.6 then 1.0.7; the plugin was 1.2.6' }
  assert.deepEqual(sweepFutureVersions(doc, versions), [])

  const failures = sweepFutureVersions({ 'doc.md': 'line\nbumped to 1.0.9 by mistake' }, versions)
  assert.equal(failures.length, 1)
  assert.match(failures[0], /doc\.md:2 names 1\.0\.9.*newer than the overlay version 1\.0\.7/)
})

test('a future plugin version is judged against the plugin line, not the overlay', () => {
  assert.deepEqual(sweepFutureVersions({ 'doc.md': 'plugin 1.2.14' }, versions), [])
  const failures = sweepFutureVersions({ 'doc.md': 'plugin 1.2.15' }, versions)
  assert.equal(failures.length, 1)
  assert.match(failures[0], /newer than the plugin version 1\.2\.14/)
})

test('fenced examples are exempt but keep the line count', () => {
  const doc = {
    'doc.md': ['prose 1.0.7', '```powershell', 'node tools/upstream-entry.mjs --check --version 1.2.99', '```', 'prose 1.0.8'].join('\n')
  }
  const failures = sweepFutureVersions(doc, versions)
  assert.equal(failures.length, 1)
  assert.match(failures[0], /doc\.md:5 names 1\.0\.8/)
})

test('a backticked literal is exempt too, because that is how a command is quoted in prose', () => {
  const doc = { 'doc.md': 'see `--version 1.2.99` for the skip case, and prose 1.2.15 is a typo' }
  const failures = sweepFutureVersions(doc, versions)
  assert.equal(failures.length, 1)
  assert.match(failures[0], /1\.2\.15/)
})

test('stripCode blanks fenced content and inline spans, and preserves the number of lines', () => {
  const text = 'a\n```\nhidden\n```\n`also hidden`\nb'
  const stripped = stripCode(text)
  assert.equal(stripped.split('\n').length, text.split('\n').length)
  assert.equal(stripped.includes('hidden'), false)
  assert.equal(stripped.startsWith('a'), true)
  assert.equal(stripped.endsWith('b'), true)
})

// ---------------------------------------------------------------- changelog

test('both release lines may carry headings of their own', () => {
  const text = ['## [1.0.7] — today', '## [1.2.14] — today', '## [1.0.6] — yesterday'].join('\n')
  assert.deepEqual(checkChangelog(text, versions), [])
})

test('a release without a heading for the current overlay version fails', () => {
  const failures = checkChangelog('## [1.0.6] — yesterday', versions)
  assert.equal(failures.length, 1)
  assert.match(failures[0], /no "## \[1\.0\.7\]" heading/)
})

test('an unreleased section below a released overlay version fails', () => {
  const text = ['## [1.0.7] — today', '## [未发布] — tomorrow'].join('\n')
  const failures = checkChangelog(text, versions)
  assert.equal(failures.length, 1)
  assert.match(failures[0], /lists \[未发布\] below 1\.0\.7/)
})

test('a released plugin version above [未发布] is fine: the lines ship independently', () => {
  const text = ['## [1.2.14] — today', '## [未发布] — also today', '## [1.0.7] — last week'].join('\n')
  assert.deepEqual(checkChangelog(text, versions), [])
})

test('an unreleased section at the top is fine - that is what a working tree looks like', () => {
  const text = ['## [未发布] — today', '## [1.0.7] — last week'].join('\n')
  assert.deepEqual(checkChangelog(text, versions), [])
})

test('a heading newer than either line fails', () => {
  const text = ['## [1.0.8] — typo', '## [1.2.15] — typo', '## [1.0.7] — today'].join('\n')
  const failures = checkChangelog(text, versions)
  assert.equal(failures.length, 2)
  assert.match(failures.join('\n'), /1\.0\.8/)
  assert.match(failures.join('\n'), /1\.2\.15/)
})

// ------------------------------------------------------------------ pipeline

function docs(overrides = {}) {
  const base = {
    'doc.md': 'version-1.0.7 and dsh-plugin-v1.2.14',
    'CHANGELOG.md': '## [1.0.7] — today'
  }
  return { ...base, ...overrides }
}

const claims = [claim]

test('check() passes when the manifests, the index and the documents agree', () => {
  const manifests = { root: '1.2.14', package_: '1.2.14', index: '1.2.14' }
  assert.deepEqual(check(docs(), manifests, 'const Version = "1.0.7"', claims), [])
})

test('check() reports two manifests that disagree', () => {
  const manifests = { root: '1.2.14', package_: '1.2.13', index: '1.2.14' }
  const failures = check(docs(), manifests, 'const Version = "1.0.7"', claims)
  assert.match(failures.join('\n'), /two plugin manifests disagree/)
})

test('check() reports an index that is behind the package', () => {
  const manifests = { root: '1.2.14', package_: '1.2.14', index: '1.2.13' }
  const failures = check(docs(), manifests, 'const Version = "1.0.7"', claims)
  assert.match(failures.join('\n'), /lists 1\.2\.13, the packaged version is 1\.2\.14/)
})

test('check() stops at a broken overlay source instead of guessing', () => {
  const manifests = { root: '1.2.14', package_: '1.2.14', index: '1.2.14' }
  const failures = check(docs(), manifests, 'var Version = 1', claims)
  assert.equal(failures.length, 1)
  assert.match(failures[0], /internal\/app\/app_windows\.go/)
})
