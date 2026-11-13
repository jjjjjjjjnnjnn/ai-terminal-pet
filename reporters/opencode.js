// AI-Terminal-Pet reporter for OpenCode.
//
// Load it by adding this file to `plugins` in opencode.json:
//   "file:///C:/path/to/ai-terminal-pet/reporters/opencode.js"
//
// Two things in here were wrong for a long time, and both are fixed by being
// explicit rather than clever:
//
//  1. The plugin API. This file used to export a function returning
//     `{ event, "tool.execute.before", "tool.execute.after" }`, which is the
//     1.x hook shape. OpenCode 2 loads a plugin by calling `setup(context)` and
//     reads events from `context.event.subscribe(...)`; its loader refuses a
//     default export without an `id` and a `setup` ("Plugin must export a
//     default definition with an id and an effect or setup function"). On 2.x
//     the old file therefore did nothing at all.
//
//  2. The event names. The old mapping listed `tool.execute.before/after`,
//     `question.asked/replied` and `permission.updated` - none of which exist
//     in 2.2.7 - and guessed the rest from substrings. A substring matcher is
//     exactly how `tui.prompt.append` (fired when text is appended to the TUI
//     prompt box, i.e. when the user types or pastes an ordinary command)
//     became `waiting_permission`, and the overlay rang for a command that no
//     agent had asked about.
//
// The catalog below is the one OpenCode 2.2.7 actually emits, read from the
// installed @opencode/schema (session-event.js, session-status-event.js,
// permission.js, form.js). Every event is named exactly; an event that is not
// listed is ignored on purpose, because an event this file cannot interpret is
// not evidence about the agent's state.
//
// The iron law still holds: nothing in this file may delay or break OpenCode.
// Every handler is wrapped, every promise is fire-and-forget, and a failure to
// report is silent.

import { createRequire } from "node:module"

const require = createRequire(import.meta.url)
const { sendStatus } = require("./protocol.js")

const SOURCE = "opencode"
const TERMINAL = "opencode"
const MAX_TRACKED = 64

// EVENT_STATE is event type -> [state, reason].
//
// The reasons are free text, but they are not arbitrary: internal/session
// derives *what* a blocked session is asking for from them, so "question" in a
// waiting_permission reason means "the agent asked a question" rather than
// "may I run this?".
//
// The choice of state matters more than it looks:
//
//   - `session.execution.succeeded` is the real completion. It is the runner's
//     terminal event for one turn, so it is the only thing mapped to idle.
//   - `session.tool.failed` is NOT an error state. A tool failing is ordinary -
//     a grep with no match, a command with a non-zero exit - and the agent
//     reads the failure and continues. Only `session.execution.failed` is
//     terminal.
//   - `session.step.failed` is not terminal either: a failed model call is
//     retried (`session.retry.scheduled`), and the runner reports the final
//     failure as `session.execution.failed`.
//   - `session.execution.interrupted` is idle with its own reason: the turn did
//     end, and saying so is honest even though the user is the one who ended
//     it.
//
// Token-level events (`session.text.delta`, `session.reasoning.delta`,
// `session.tool.input.delta`, `session.tool.progress`) are deliberately absent.
// They change nothing the panel draws, and they fire hundreds of times per
// answer; forwarding them would repaint the overlay continuously.
const EVENT_STATE = new Map(Object.entries({
  // --- the turn ---------------------------------------------------------
  "session.execution.started": ["working", "execution_started"],
  "session.execution.succeeded": ["idle", "execution_succeeded"],
  "session.execution.failed": ["error", "execution_failed"],
  "session.execution.interrupted": ["idle", "execution_interrupted"],

  // --- the agent's own loop ---------------------------------------------
  "session.step.started": ["thinking", "step_started"],
  "session.step.streamed": ["working", "step_streamed"],
  "session.step.ended": ["working", "step_ended"],
  "session.step.failed": ["working", "step_failed"],
  "session.text.started": ["working", "text_started"],
  "session.text.ended": ["working", "text_ended"],
  "session.reasoning.started": ["thinking", "reasoning_started"],
  "session.reasoning.ended": ["thinking", "reasoning_ended"],
  "session.retry.scheduled": ["working", "retry_scheduled"],
  "session.compaction.started": ["working", "compaction_started"],
  "session.compaction.ended": ["working", "compaction_ended"],
  "session.compaction.failed": ["working", "compaction_failed"],
  "session.shell.started": ["working", "shell_started"],
  "session.shell.ended": ["working", "shell_ended"],
  "session.inbox.enqueued": ["working", "prompt_queued"],
  "session.inbox.delivered": ["working", "prompt_delivered"],

  // --- tools ------------------------------------------------------------
  "session.tool.input.started": ["tool_use", "tool_input"],
  "session.tool.input.ended": ["tool_use", "tool_input_ready"],
  "session.tool.called": ["tool_use", "tool_called"],
  "session.tool.success": ["working", "tool_succeeded"],
  "session.tool.failed": ["working", "tool_failed"],

  // --- lifecycle --------------------------------------------------------
  "session.created": ["idle", "session_created"],
  "session.deleted": ["closed", "session_deleted"],
}))

// eventData reads one event's payload.
//
// OpenCode 2 frames events as `{ id, created, type, durable?, location?, data }`
// (@opencode/schema event.js). `properties` is the 1.x name and is kept as a
// fallback so a host that still sends the older envelope is not silently
// ignored - the type is checked exactly, so a mis-framed event can never be
// mistaken for another.
function eventData(event) {
  if (!event || typeof event !== "object") return {}
  if (event.data && typeof event.data === "object") return event.data
  if (event.properties && typeof event.properties === "object") return event.properties
  return {}
}

// sessionIdOf finds the conversation id. Only an explicit sessionID counts:
// event, message and part ids (evt_/msg_/prt_) are not conversations, and using
// them once created a ghost row per event. form.created nests it one level
// deeper, because a form belongs to the session that asked it.
function sessionIdOf(data) {
  if (!data || typeof data !== "object") return ""
  const form = data.form && typeof data.form === "object" ? data.form : {}
  const candidates = [data.sessionID, data.sessionId, form.sessionID, data.part?.sessionID]
  for (const candidate of candidates) {
    if (typeof candidate !== "string") continue
    const id = candidate.trim()
    if (!id || id.length > 200) continue
    if (/^(evt|msg|prt|prj)_/i.test(id)) continue
    return id
  }
  return ""
}

// modelName renders Model.Ref ({ providerID, id }) as "provider/model".
function modelName(value) {
  if (typeof value === "string") return value || null
  if (!value || typeof value !== "object") return null
  const model = value.id || value.modelID || value.modelId || value.name
  const provider = value.providerID || value.providerId || value.provider
  return [provider, model].filter(Boolean).join("/") || null
}

// mapEvent answers "what does this event say about the session", or null when
// the event says nothing this reporter acts on.
function mapEvent(type, data) {
  const name = String(type || "").toLowerCase()
  const payload = data && typeof data === "object" ? data : {}

  // session.status carries the current status, not a transition, so it is read
  // rather than looked up. The three values are exact (SessionStatusEvent.Info):
  // idle / retry / busy.
  if (name === "session.status") {
    const status = payload.status
    const kind = String((status && typeof status === "object" ? status.type : status) || "").toLowerCase()
    if (kind === "busy") return { state: "working", reason: "status_busy" }
    if (kind === "retry") return { state: "working", reason: "status_retry" }
    if (kind === "idle") return { state: "idle", reason: "status_idle" }
    return null
  }

  // A permission request is the one state that must never be missed: the agent
  // is blocked until a human answers. `action` is the tool the request is
  // about ("bash", "edit"), which is what the row shows.
  if (name === "permission.asked") {
    return { state: "waiting_permission", reason: "permission_required", tool: payload.action }
  }
  if (name === "permission.replied") {
    const reply = String(payload.reply || "").toLowerCase()
    if (reply === "reject") return { state: "working", reason: "permission_rejected" }
    return { state: "thinking", reason: "permission_granted" }
  }

  // A form is a question, not an approval: same waiting state, different reason,
  // and internal/session turns "question" into the question kind.
  if (name === "form.created") return { state: "waiting_permission", reason: "question_asked" }
  if (name === "form.replied") return { state: "thinking", reason: "question_replied" }
  if (name === "form.cancelled") return { state: "working", reason: "question_cancelled" }

  const entry = EVENT_STATE.get(name)
  if (!entry) return null
  const mapped = { state: entry[0], reason: entry[1] }
  // The tool name is only carried by tool events, and it is the model's name for
  // the tool, never its arguments: the protocol redacts and bounds it anyway,
  // but nothing here should ever send a command line.
  if (name.startsWith("session.tool.") && typeof payload.name === "string") {
    mapped.tool = payload.name
  }
  return mapped
}

// applyMetadata folds the parts of an event that describe the conversation
// rather than its state (title, parent, model, project) into the tracked entry.
// Returns true when something actually changed, so the row can be refreshed
// without inventing a state transition.
function applyMetadata(meta, data) {
  const payload = data && typeof data === "object" ? data : {}
  let changed = false
  if (typeof payload.title === "string" && payload.title.trim() && payload.title !== meta.title) {
    meta.title = payload.title
    changed = true
  }
  if (typeof payload.parentID === "string" && payload.parentID && payload.parentID !== meta.parentSessionId) {
    meta.parentSessionId = payload.parentID
    changed = true
  }
  const model = modelName(payload.model)
  if (model && model !== meta.model) {
    meta.model = model
    changed = true
  }
  const directory = payload.location?.directory || payload.directory
  if (typeof directory === "string" && directory && directory !== meta.project) {
    meta.project = directory
    changed = true
  }
  return changed
}

// sessions is process-wide on purpose: a plugin reload creates a new instance
// while the conversations it was reporting are still running, and the exit
// handler has to be able to close everything the process ever reported.
const sessions = new Map()

function createMeta(sessionId) {
  return {
    sessionId,
    parentSessionId: null,
    project: null,
    title: null,
    model: null,
    startedAt: null,
    state: "",
    reason: "",
    tool: null,
    hydrated: false
  }
}

// report sends one state for one conversation.
//
// Identical reports are dropped. The overlay repaints on every store revision,
// so a reporter that re-sends "working" on each step would keep the panel busy
// for no new information; `force` exists only for a metadata refresh, where the
// state is the same but the row's title or model is not.
//
// Dropping a repeat cannot make a row disappear: internal/store keeps a
// snapshot whose host process is still alive regardless of how old its
// timestamp is (SnapshotScanner.KeepAlive).
function report(meta, state, reason, tool, force) {
  const nextTool = tool ? String(tool) : meta.tool
  if (!force && meta.state === state && meta.reason === reason && meta.tool === nextTool) return
  meta.state = state
  meta.reason = reason
  meta.tool = nextTool
  void sendStatus({
    source: SOURCE,
    sessionId: meta.sessionId,
    parentSessionId: meta.parentSessionId,
    project: meta.project,
    title: meta.title,
    model: meta.model,
    state,
    reason,
    startedAt: meta.startedAt,
    updatedAt: new Date().toISOString(),
    tool: meta.tool,
    terminal: TERMINAL
  }).catch(() => {})
}

function createTracker(context) {
  const client = context && context.session && typeof context.session.get === "function" ? context.session : null
  const fallbackProject = typeof context?.location?.directory === "string" ? context.location.directory : null

  // hydrate fills in what an event does not carry - the title, the model, the
  // parent session, the project directory - by asking the host once per
  // conversation. It never throws and never blocks the event loop.
  function hydrate(meta) {
    if (meta.hydrated || !client) return
    meta.hydrated = true
    void (async () => {
      try {
        const info = await client.get({ sessionID: meta.sessionId })
        if (!info || typeof info !== "object") return
        const changed = applyMetadata(meta, {
          title: info.title,
          parentID: info.parentID,
          model: info.model,
          location: info.location
        })
        if (typeof info.time?.created === "number" && !meta.startedAt) {
          meta.startedAt = new Date(info.time.created).toISOString()
        }
        if (meta.state) report(meta, meta.state, meta.reason, null, changed)
      } catch {
        // A host that cannot answer costs a title, not a session.
      }
    })()
  }

  function track(sessionId) {
    let meta = sessions.get(sessionId)
    if (meta) return { meta, known: true }
    if (sessions.size >= MAX_TRACKED) {
      sessions.delete(sessions.keys().next().value)
    }
    meta = createMeta(sessionId)
    meta.project = fallbackProject
    sessions.set(sessionId, meta)
    hydrate(meta)
    return { meta, known: false }
  }

  function handle(event) {
    try {
      const type = String(event?.type || "").toLowerCase()
      if (!type) return
      const data = eventData(event)
      const sessionId = sessionIdOf(data)
      if (!sessionId) return
      const { meta, known } = track(sessionId)
      const metadataChanged = applyMetadata(meta, data)
      const mapped = mapEvent(type, data)
      if (!mapped) {
        // A metadata-only event (session.created, session.renamed,
        // session.model.selected): refresh the row if there is a row to
        // refresh, but never invent a state for a conversation we have not
        // seen do anything yet.
        if (known && metadataChanged && meta.state) report(meta, meta.state, meta.reason, null, true)
        return
      }
      report(meta, mapped.state, mapped.reason, mapped.tool, metadataChanged)
      if (mapped.state === "closed") {
        // The closed snapshot sendStatus just wrote is the durable end signal:
        // the overlay consumes it and tombstones the session. Nothing is
        // deleted here, because deleting the file would leave the overlay with
        // no explanation for the row that disappears.
        sessions.delete(sessionId)
      }
    } catch {
      // A malformed event costs one update, never the session.
    }
  }

  return { handle }
}

// close is the process-exit path and nothing else. A plugin reload is not the
// end of the user's conversations (see setup below), but a real exit is, and a
// `closed` snapshot is how the overlay learns it: sendStatus writes the file
// synchronously before its first await, so this works even when the process is
// already on its way out and the pipe write never completes.
function closeAll() {
  const now = new Date().toISOString()
  for (const meta of sessions.values()) {
    void sendStatus({
      source: SOURCE,
      sessionId: meta.sessionId,
      parentSessionId: meta.parentSessionId,
      project: meta.project,
      title: meta.title,
      model: meta.model,
      state: "closed",
      reason: "process_exit",
      startedAt: meta.startedAt,
      updatedAt: now,
      tool: meta.tool,
      terminal: TERMINAL
    }).catch(() => {})
  }
  sessions.clear()
}

let lifecycleBound = false

export const plugin = {
  id: "ai-terminal-pet",
  setup(context) {
    if (!lifecycleBound) {
      lifecycleBound = true
      process.once("exit", closeAll)
    }
    const tracker = createTracker(context)
    const controller = new AbortController()
    void (async () => {
      try {
        if (typeof context?.event?.subscribe !== "function") return
        for await (const event of context.event.subscribe({ signal: controller.signal })) {
          tracker.handle(event)
        }
      } catch {
        // The subscription ends when the plugin is unloaded; that is normal.
      }
    })()
    return () => {
      // Abort only. Marking everything closed here would empty the panel every
      // time this file is edited, because a reload is not the end of the
      // conversations OpenCode is still running.
      controller.abort()
    }
  }
}

export default plugin

// The mapping helpers are exported so reporters/opencode-events.test.mjs can pin
// the event catalog without a pipe, a window or an OpenCode process.
export { eventData, mapEvent, sessionIdOf }
