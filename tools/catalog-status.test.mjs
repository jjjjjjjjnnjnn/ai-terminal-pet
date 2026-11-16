import assert from 'node:assert/strict'
import test from 'node:test'
import {
  compareVersions,
  describeEntry,
  downloadShortfall,
  findEntry,
  tarballVersion,
  verdict,
} from './catalog-status.mjs'

// The three cases this tool exists to tell apart, taken from the live data of
// 2026-11-16 (handover §64). A test that only knew "out of date" would have
// produced the wrong action for two of them: a bump PR, a rebuild we cannot
// trigger, and a re-probe that would change nothing.

const OUR_TARBALL_1221 = 'https://github.com/jjjjjjjjnnjnn/ai-terminal-pet/releases/download/dsh-plugin-v1.2.21/dsh-ai-terminal-pet-1.2.21.tgz'

const siteRow = {
  name: 'ai-terminal-pet#dsh-ai-terminal-pet',
  npm: 'dsh-ai-terminal-pet',
  version: '1.2.16',
  tarball: OUR_TARBALL_1221,
  downloads: 1179,
  downloadsStart: '2026-10-15',
  downloadsEnd: '2026-11-13',
}

test('finds our row by npm name rather than by position', () => {
  const index = { plugins: [{ npm: 'other' }, siteRow] }
  assert.equal(findEntry(index)?.name, siteRow.name)
  assert.equal(findEntry(index, 'not-listed'), null)
  assert.equal(findEntry({ plugins: 'nonsense' }), null)
  assert.equal(findEntry(null), null)
})

test('reads the version out of the tarball URL, and refuses to invent one', () => {
  assert.equal(tarballVersion(OUR_TARBALL_1221), '1.2.21')
  assert.equal(tarballVersion('https://example.invalid/latest.tgz'), null)
  assert.equal(tarballVersion(null), null)
})

test('version comparison is numeric, not lexical', () => {
  assert.equal(compareVersions('1.2.9', '1.2.10'), -1)
  assert.equal(compareVersions('1.2.10', '1.2.9'), 1)
  assert.equal(compareVersions('1.2.22', '1.2.22'), 0)
})

test('the download shortfall is the final day of the window, exactly', () => {
  // npm's real rows for 2026-10-15..2026-11-13, queried 2026-11-16: only the
  // last three days carried traffic, and the published 1,179 is the first two.
  const days = [
    { day: '2026-11-11', downloads: 624 },
    { day: '2026-11-12', downloads: 555 },
    { day: '2026-11-13', downloads: 1197 },
  ]
  const result = downloadShortfall(1179, days)
  assert.deepEqual(result, {
    site: 1179,
    now: 2376,
    withoutLastDay: 1179,
    lastDay: 1197,
    lagging: true,
    shortfall: 1197,
  })
})

test('a genuine zero day is not called lagging', () => {
  // npm legitimately reports 0 for days with no installs. The published total
  // then equals the full window and there is nothing short.
  const days = [
    { day: '2026-11-11', downloads: 0 },
    { day: '2026-11-12', downloads: 0 },
  ]
  const result = downloadShortfall(0, days)
  assert.equal(result.lagging, false)
  assert.equal(result.shortfall, 0)
})

test('verdict: a frozen catalogue is upstream, never ours', () => {
  const result = verdict({
    entry: siteRow,
    npmVersion: '1.2.22',
    publishedVersion: tarballVersion(siteRow.tarball),
    downloads: null,
    localVersion: '1.2.22',
  })
  assert.equal(result.level, 'upstream')
  assert.equal(result.fixableByUs, false)
  assert.match(result.reason, /1\.2\.16/)
})

test('verdict: lagging downloads outlive a rebuild, so they are the next cause', () => {
  const result = verdict({
    entry: { ...siteRow, version: '1.2.22' },
    npmVersion: '1.2.22',
    publishedVersion: '1.2.22',
    downloads: { site: 1179, now: 2376, withoutLastDay: 1179, lastDay: 1197, lagging: true, shortfall: 1197 },
    localVersion: '1.2.22',
  })
  assert.equal(result.level, 'upstream')
  assert.match(result.reason, /#15864/)
})

test('verdict: our tarball behind npm IS ours to fix', () => {
  const result = verdict({
    entry: { ...siteRow, version: '1.2.22' },
    npmVersion: '1.2.22',
    publishedVersion: '1.2.21',
    downloads: null,
    localVersion: '1.2.22',
  })
  assert.equal(result.level, 'fail')
  // Opening the bump PR is ours; merging it is not. The field answers "does a
  // re-release fix this", and for this branch the answer is a PR, not a release.
  assert.equal(result.fixableByUs, true)
  assert.match(result.reason, /1\.2\.21/)
})

test('verdict: current row, and a row that is missing entirely', () => {
  assert.equal(verdict({
    entry: { ...siteRow, version: '1.2.22' },
    npmVersion: '1.2.22',
    publishedVersion: '1.2.22',
    downloads: null,
    localVersion: '1.2.22',
  }).level, 'ok')
  assert.equal(verdict({ entry: null, npmVersion: '1.2.22', publishedVersion: null, downloads: null, localVersion: '1.2.22' }).level, 'fail')
})

test('verdict offline: a tarball behind the working tree is pending, not a failure', () => {
  // No npmVersion: we cannot say whether the bump is in flight, so the honest
  // answer is "check upstream-entry --check", not "ours to fix".
  assert.equal(verdict({
    entry: siteRow,
    npmVersion: null,
    publishedVersion: '1.2.21',
    downloads: null,
    localVersion: '1.2.22',
  }).level, 'pending')
})

test('describeEntry carries exactly the fields the verdict and the report use', () => {
  assert.deepEqual(describeEntry(siteRow), {
    name: siteRow.name,
    version: '1.2.16',
    tarball: OUR_TARBALL_1221,
    publishedVersion: '1.2.21',
    downloads: 1179,
    window: '2026-10-15:2026-11-13',
  })
  assert.equal(describeEntry(null), null)
})
