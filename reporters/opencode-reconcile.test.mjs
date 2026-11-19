// Tests for the host-backed stale-row sweep in the OpenCode reporter.
//
//   node --test reporters/opencode-reconcile.test.mjs
//
// The field report (handover §88) was "the panel says 16 sessions, my terminal
// does not have 10". All 16 still exist in OpenCode's database and none is
// archived, so "ask the host if the session exists" cannot fix it. What makes
// nine of them invisible to the user is that five are subagents and four were
// last viewed yesterday. Every branch below is pinned with a synthetic clock.

import assert from 'node:assert/strict'
import test from 'node:test'

import {
  STALE_IDLE_MS,
  applyViewedAt,
  collectStale,
  staleReason,
  viewedAtFrom
} from './opencode-reconcile.js'

const NOW = Date.UTC(2026, 10, 19, 8, 0, 0)

test('viewedAtFrom accepts the host shape and never guesses', () => {
  assert.equal(viewedAtFrom({ time: { viewed: 1234 } }), 1234)
  assert.equal(viewedAtFrom({ time: { viewed: '2026-11-19T07:00:00.000Z' } }), Date.UTC(2026, 10, 19, 7))
  assert.equal(viewedAtFrom({ time: { archived: 1234 } }), 0)
  assert.equal(viewedAtFrom({ time: { viewed: 'not a date' } }), 0)
  assert.equal(viewedAtFrom(null), 0)
})

test('an archived conversation is always closed', () => {
  assert.equal(staleReason({ archived: true, state: 'working' }, NOW), 'session_archived')
})

test('a busy or waiting session is never closed however old its view is', () => {
  for (const state of ['working', 'thinking', 'tool_use', 'waiting_permission', 'error']) {
    assert.equal(staleReason({ state, viewedAt: 1 }, NOW), null, state)
  }
})

test('a settled subagent is not a conversation the user can switch to', () => {
  assert.equal(
    staleReason({ state: 'idle', parentSessionId: 'ses_parent', viewedAt: NOW }, NOW),
    'subagent_settled'
  )
  assert.equal(
    staleReason({ state: 'working', parentSessionId: 'ses_parent', viewedAt: NOW }, NOW),
    null
  )
})

test('an idle conversation is stale only past the threshold', () => {
  const viewedAt = NOW - STALE_IDLE_MS + 1
  assert.equal(staleReason({ state: 'idle', viewedAt }, NOW), null)
  assert.equal(staleReason({ state: 'idle', viewedAt: viewedAt - 2 }, NOW), 'stale_idle')
  assert.equal(staleReason({ state: 'idle', viewedAt: 0 }, NOW), null, 'unknown view = keep the row')
})

test('collectStale returns pairs and mutates nothing', () => {
  const sessions = new Map([
    ['live', { state: 'idle', viewedAt: NOW - 1000 }],
    ['old', { state: 'idle', viewedAt: NOW - STALE_IDLE_MS - 1 }],
    ['child', { state: 'idle', parentSessionId: 'p', viewedAt: NOW }],
    ['busy', { state: 'working', viewedAt: 1 }]
  ])
  assert.deepEqual(collectStale(sessions, NOW), [
    ['old', 'stale_idle'],
    ['child', 'subagent_settled']
  ])
  assert.equal(sessions.size, 4)
})

test('replay of the field snapshot: nine of sixteen rows are not on a terminal', () => {
  // Shape of the 2026-11-19 06:47 UTC snapshot (handover §87): 16 opencode
  // sessions, 11 idle, 3 working, 2 thinking. Five rows have a parent session
  // (subagent steps, no time_viewed in the host database) and four idle rows
  // were last viewed on 11-18. No personal paths or titles are reproduced.
  const now = Date.UTC(2026, 10, 19, 8, 0, 0)
  const yesterday = Date.UTC(2026, 10, 18, 21, 0, 0)
  const today = Date.UTC(2026, 10, 19, 6, 45, 0)
  const rows = []
  const add = (id, state, extra = {}) => rows.push({ id, state, viewedAt: today, ...extra })
  for (let i = 0; i < 3; i++) add(`work-${i}`, 'working')
  for (let i = 0; i < 2; i++) add(`think-${i}`, 'thinking')
  for (let i = 0; i < 2; i++) add(`live-idle-${i}`, 'idle')
  for (let i = 0; i < 5; i++) add(`child-${i}`, 'idle', { parentSessionId: 'parent', viewedAt: 0 })
  for (let i = 0; i < 4; i++) add(`old-${i}`, 'idle', { viewedAt: yesterday })
  const sessions = new Map(rows.map((row) => [row.id, row]))
  assert.equal(sessions.size, 16)

  const closed = collectStale(sessions, now)
  assert.equal(closed.length, 9)
  assert.equal(closed.filter(([, reason]) => reason === 'subagent_settled').length, 5)
  assert.equal(closed.filter(([, reason]) => reason === 'stale_idle').length, 4)
  // The remaining seven are exactly the sessions with real terminal activity:
  // the three busy ones, two thinking ones and two idle ones viewed today.
  assert.equal(sessions.size - closed.length, 7)
})

test('viewed events refresh the mark without inventing a state', () => {
  const meta = { state: 'idle', viewedAt: NOW - STALE_IDLE_MS - 10 }
  assert.equal(applyViewedAt(meta, { time: { viewed: NOW } }), true)
  assert.equal(meta.viewedAt, NOW)
  assert.equal(staleReason(meta, NOW), null)
  assert.equal(applyViewedAt(meta, { time: { viewed: NOW } }), false)
  assert.equal(meta.state, 'idle')
})
