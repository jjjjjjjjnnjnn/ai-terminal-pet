// Regression tests for the OpenCode reporter's event mapping.
//
//   node --test reporters/opencode-events.test.mjs
//
// No pipe, no window, no OpenCode process: the mapping is pure, so the questions
// the two shipped bugs asked can be answered directly.
//
// The bugs, both measured on the installed plugin before they were fixed:
//
//   - `tui.prompt.append` ("text was appended to the prompt box", i.e. the user
//     typed or pasted an ordinary command) was mapped to waiting_permission by
//     a substring matcher, so the overlay rang for a command no agent had asked
//     about. The same matcher turned `session.prompt.*` and
//     `session.queued_prompts` into permission prompts.
//   - The events that do carry a real ask (`permission.asked`) and the ones that
//     carry real work (`session.tool.called`, `session.execution.*`) were not
//     mapped at all under their 2.2.7 names, so the panel showed a session as
//     busy or idle with no relation to what the agent was doing.

import assert from 'node:assert/strict'
import test from 'node:test'

import { eventData, mapEvent, plugin, sessionIdOf } from './opencode.js'

// Events OpenCode 2.2.7 emits that say nothing about a session's state, or that
// describe the TUI rather than the agent. None of them may ever produce a
// waiting_permission - that is the phantom ring.
const IGNORED = [
  'tui.prompt.append',
  'tui.command.execute',
  'tui.toast.show',
  'tui.session.select',
  'session.viewed',
  'session.renamed',
  'session.model.selected',
  'session.agent.selected',
  'session.metadata.updated',
  'session.permissions',
  'session.usage.updated',
  'session.usage.recorded',
  'session.instructions.updated',
  'session.message.content.updated',
  'session.revert.staged',
  'session.revert.cleared',
  'session.synthetic',
  'session.skill.activated',
  'session.moved',
  'session.forked',
  'session.text.delta',
  'session.reasoning.delta',
  'session.tool.input.delta',
  'session.tool.progress',
  'permission.updated',
  'question.asked',
  'question.replied',
  'tool.execute.before',
  'tool.execute.after'
]

test('TUI and metadata events never ask for permission', () => {
  for (const type of IGNORED) {
    assert.equal(mapEvent(type, {}), null, `${type} must not map to a state`)
  }
})

test('a permission request and a question both wait, with different reasons', () => {
  const permission = mapEvent('permission.asked', { sessionID: 'ses_1', action: 'bash' })
  assert.equal(permission.state, 'waiting_permission')
  assert.equal(permission.tool, 'bash')
  // internal/session turns a reason containing "question" into the question
  // kind and everything else into an approval, so these two strings are the
  // contract, not decoration.
  assert.doesNotMatch(permission.reason, /question|plan|ask/)

  const form = mapEvent('form.created', { form: { id: 'frm_1', sessionID: 'ses_1', title: 'Which branch?' } })
  assert.equal(form.state, 'waiting_permission')
  assert.match(form.reason, /question/)

  assert.equal(mapEvent('permission.replied', { reply: 'reject' }).state, 'working')
  assert.equal(mapEvent('permission.replied', { reply: 'once' }).state, 'thinking')
  assert.equal(mapEvent('form.replied', {}).state, 'thinking')
  assert.equal(mapEvent('form.cancelled', {}).state, 'working')
})

test('a failed tool is not a failed turn', () => {
  // The commonest thing an agent does is run a command that fails. Mapping it
  // to the error state made the overlay announce an error for ordinary work.
  assert.equal(mapEvent('session.tool.failed', { name: 'bash' }).state, 'working')
  assert.equal(mapEvent('session.tool.success', { name: 'bash' }).state, 'working')
  assert.equal(mapEvent('session.step.failed', {}).state, 'working')
  assert.equal(mapEvent('session.retry.scheduled', {}).state, 'working')
  // Only the runner's own terminal failure is an error.
  assert.equal(mapEvent('session.execution.failed', {}).state, 'error')
})

test('a turn ends exactly when the runner says it ended', () => {
  assert.equal(mapEvent('session.execution.started', {}).state, 'working')
  assert.equal(mapEvent('session.execution.succeeded', {}).state, 'idle')
  assert.equal(mapEvent('session.execution.interrupted', {}).state, 'idle')
  assert.equal(mapEvent('session.idle', {}), null, 'session.idle is deprecated in 2.2.7')
  assert.equal(mapEvent('session.status', { status: { type: 'idle' } }).state, 'idle')
  assert.equal(mapEvent('session.status', { status: { type: 'busy' } }).state, 'working')
  assert.equal(mapEvent('session.status', { status: { type: 'retry' } }).state, 'working')
})

test('tool events carry the tool name and nothing else', () => {
  const called = mapEvent('session.tool.called', { name: 'grep', input: { pattern: 'secret' } })
  assert.equal(called.state, 'tool_use')
  assert.equal(called.tool, 'grep')
  assert.equal(mapEvent('session.tool.called', { input: {} }).tool, undefined)
})

test('a session id is a conversation, never an event id', () => {
  assert.equal(sessionIdOf({ sessionID: 'ses_abc' }), 'ses_abc')
  assert.equal(sessionIdOf({ form: { sessionID: 'ses_form' } }), 'ses_form')
  assert.equal(sessionIdOf({ part: { sessionID: 'ses_part' } }), 'ses_part')
  // Event, message, part and project ids are not conversations. Using them once
  // created one ghost row per event.
  for (const id of ['evt_1', 'msg_2', 'prt_3', 'prj_4']) {
    assert.equal(sessionIdOf({ sessionID: id }), '', `${id} is not a conversation`)
  }
  assert.equal(sessionIdOf({}), '')
  assert.equal(sessionIdOf({ sessionID: '   ' }), '')
})

test('the payload is read from data, and from properties as a fallback', () => {
  // OpenCode 2 frames events as { type, data }; 1.x used { type, properties }.
  assert.deepEqual(eventData({ type: 'x', data: { sessionID: 'ses_1' } }), { sessionID: 'ses_1' })
  assert.deepEqual(eventData({ type: 'x', properties: { sessionID: 'ses_2' } }), { sessionID: 'ses_2' })
  assert.deepEqual(eventData({ type: 'x', data: { sessionID: 'ses_1' }, properties: { sessionID: 'ses_2' } }), {
    sessionID: 'ses_1'
  })
  assert.deepEqual(eventData(null), {})
})

test('the export is a 2.x plugin, not a 1.x hook function', () => {
  // OpenCode 2 refuses a default export without these two fields
  // ("Plugin must export a default definition with an id and an effect or setup
  // function"), which is why the previous shape reported nothing at all.
  assert.equal(typeof plugin, 'object')
  assert.equal(plugin.id, 'ai-terminal-pet')
  assert.equal(typeof plugin.setup, 'function')
})
