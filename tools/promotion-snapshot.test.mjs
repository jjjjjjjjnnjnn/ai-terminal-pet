// Governance for the repo -> knowledge-base promotion bus handoff.
//
//   node --test tools/promotion-snapshot.test.mjs
//
// dsh-remedy established this pattern (R77/R78). The bus is the only channel
// state machine; this repo provides facts and one read-only JSON snapshot. These
// tests pin both halves so a later edit cannot quietly turn the repo into a
// second source of truth or break the contract `kb promo collect` reads.

import assert from 'node:assert/strict'
import { spawnSync } from 'node:child_process'
import { readFileSync } from 'node:fs'
import test from 'node:test'
import { fileURLToPath } from 'node:url'

const root = new URL('../', import.meta.url)
const read = (name) => readFileSync(new URL(name, root), 'utf8')

test('the repo defers channel status to the knowledge-base bus', () => {
  const promotion = read('docs/PROMOTION.md')
  const playbook = read('docs/PROMOTION-PLAYBOOK.md')
  const reach = read('docs/reach/README.md')
  assert.ok(promotion.includes('唯一宣传状态机'), 'PROMOTION.md must name the bus as the state machine')
  assert.ok(promotion.includes('kb promo'), 'the bus commands must be discoverable from the repo')
  assert.ok(playbook.includes('宣传总线'), 'the playbook must keep the bus boundary')
  assert.ok(reach.includes('not the channel state machine'), 'the local reach ledger must not claim to be status')
  assert.ok(reach.includes('kb promo record'), 'real posts must be registered in the bus')
  assert.ok(promotion.includes('agent-promotion'), 'posting must be delegated to the KB agent-promotion')
})

test('promo-version reads the published tag, not the working-tree version', () => {
  const source = read('tools/promo-version.mjs')
  assert.match(source, /['"]release['"]\s*,\s*['"]list['"]/, 'it must ask GitHub for releases')
  assert.match(source, /pet-v/, 'it must filter the overlay tag line')
  assert.doesNotMatch(source, /readFileSync\([^)]*version\.go/, 'a local bump must not be presented as published')
  assert.match(source, /process\.exit\(1\)/, 'an unreadable tag must be reported, not guessed')
})

test('daily-reach --json speaks promotion-snapshot/v1 for the bus', () => {
  const result = spawnSync(process.execPath, ['tools/daily-reach.mjs', '--json'], {
    cwd: fileURLToPath(root),
    // node --test sets NODE_TEST_CONTEXT for child processes; a plain script
    // started with that flag can abort under the test runner on Windows.
    env: { ...process.env, NODE_TEST_CONTEXT: '' },
    encoding: 'utf8',
    maxBuffer: 32 << 20,
    timeout: 240000
  })
  assert.equal(result.status, 0, String(result.stderr ?? '').slice(0, 300))
  const snapshot = JSON.parse(result.stdout)
  assert.equal(snapshot.schema, 'promotion-snapshot/v1')
  assert.equal(snapshot.project_id, 'ai-terminal-pet')
  assert.match(snapshot.release.version, /^\d+\.\d+\.\d+$/)
  assert.equal(snapshot.release.version_source, 'release_tag')
  for (const key of ['github', 'npm', 'catalogue', 'releases', 'social', 'upstream', 'unreachable', 'source']) {
    assert.ok(key in snapshot, `snapshot is missing ${key}`)
  }
  assert.match(snapshot.source.command, /daily-reach\.mjs --json/)
  // A failed source must be named, never turned into a zero. The arrays may be
  // empty on an offline machine, so only the key's existence is pinned here.
  assert.ok(Array.isArray(snapshot.unreachable))
  assert.ok(snapshot.facts && typeof snapshot.facts === 'object', 'the bus needs machine-readable facts')
})
