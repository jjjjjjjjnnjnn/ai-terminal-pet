// The rules behind tools/traffic-report.mjs, which are the only judgement calls
// in it: what "the newest versions" means, and where the line between a human
// distribution and a crawler fleet sits.
//
//   node --test tools/traffic-report.test.mjs

import assert from 'node:assert/strict'
import test from 'node:test'

import { assetSummary, latestShare, versionRank } from './traffic-report.mjs'

test('versions order numerically, not lexicographically', () => {
  // "1.2.9" > "1.2.10" as strings, which would pick the wrong versions as newest.
  assert.ok(versionRank('1.2.10') > versionRank('1.2.9'))
  assert.ok(versionRank('1.10.0') > versionRank('1.9.0'))
  // A prerelease sits at its release's rank rather than above or below the line.
  assert.equal(versionRank('1.2.0-stage'), versionRank('1.2.0'))
})

test('the latest share is measured on the two newest versions', () => {
  // A human distribution: the newest two carry most of it, the long tail is old
  // pins. These are dsh-pet's real numbers from 2026-11-17.
  const human = { '0.4.7': 29383, '0.4.8': 29240, '0.4.9': 21312, '0.3.8': 1998, '0.1.5': 3 }
  const share = latestShare(human)
  assert.deepEqual(share.newest.sort(), ['0.4.8', '0.4.9'])
  assert.ok(share.share > 0.5, `humans should clear half, got ${share.share}`)

  // A crawler fleet: every version fetched once per run, including a version
  // nobody can install. These are this project's real shape from the same day:
  // 25 versions, 137-227 downloads each, and 166 for `0.0.0-stage`.
  const fleet = { '0.0.0-stage': 166 }
  for (let minor = 4; minor <= 27; minor++) fleet[`1.2.${minor}`] = 137 + ((minor * 7) % 90)
  assert.ok(latestShare(fleet).share < 0.1, 'a flat distribution is not a human one')
})

test('an empty or zero distribution has no share rather than a division by zero', () => {
  assert.equal(latestShare({}), null)
  assert.equal(latestShare({ '1.0.0': 0 }), null)
})

test('release assets are grouped by what they install, with a per-release rate', () => {
  const summary = assetSummary([
    { name: 'dsh-ai-terminal-pet-1.2.31.tgz', downloads: 24 },
    { name: 'dsh-ai-terminal-pet-1.2.32.tgz', downloads: 26 },
    { name: 'pet-setup.exe', downloads: 2 },
    { name: 'ai-terminal-pet-1.0.19-win-x64.zip', downloads: 2 },
    { name: 'ai-terminal-pet-1.0.19-linux-amd64.tar.gz', downloads: 1 },
    { name: 'ai-terminal-pet-1.0.19-darwin-arm64.tar.gz', downloads: 1 }
  ])
  const dsh = summary.grouped.find((kind) => kind.name === 'DSH plugin tarball')
  assert.equal(dsh.downloads, 50)
  assert.equal(dsh.releases, 2)
  assert.equal(dsh.perRelease, 25)
  // The per-release rate is the whole point: ~2 means our own verification.
  const setup = summary.grouped.find((kind) => kind.name === 'Windows setup')
  assert.equal(setup.perRelease, 2)
  assert.equal(summary.total, 56)
})
