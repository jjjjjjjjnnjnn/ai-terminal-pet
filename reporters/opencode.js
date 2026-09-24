import { createRequire } from "node:module"
const require = createRequire(import.meta.url)
const { removeRuntimeSnapshot, sendStatus } = require("./protocol.js")
const instances = new Map()
const terminal = process.env.TERM_PROGRAM || (process.env.WT_SESSION ? "Windows Terminal" : "opencode")
let lifecycleBound = false

function modelName(value) {
  if (typeof value === "string") return value
  if (!value || typeof value !== "object") return null
  const model = value.modelID || value.modelId || value.id || value.name
  const provider = value.providerID || value.providerId || value.provider
  return [provider, model].filter(Boolean).join("/") || null
}

function isoTime(value) {
  if (typeof value !== "number" && typeof value !== "string") return null
  const number = typeof value === "string" ? Date.parse(value) : value
  if (!Number.isFinite(number)) return null
  const date = new Date(number < 100000000000 ? number * 1000 : number)
  return Number.isNaN(date.getTime()) ? null : date.toISOString()
}

function eventState(type, properties, previous) {
  const status = properties.status
  const raw = typeof status === "string" ? status : status && (status.type || status.state || status.status)
  const normalized = String(raw || "").toLowerCase()
  if (type === "session.error" || type === "error") return ["error", "session_error"]
  if (type === "session.deleted") return ["closed", "session_deleted"]
  if (type === "session.idle") return ["idle", "session_idle"]
  if (type === "session.status") {
    if (normalized.includes("idle")) return ["idle", "status_idle"]
    if (normalized.includes("error") || normalized.includes("fail")) return ["error", "status_error"]
    if (normalized.includes("retry")) return ["working", "status_retry"]
    if (normalized.includes("busy") || normalized.includes("run") || normalized.includes("work")) return ["working", "status_busy"]
    return [previous || "idle", "status"]
  }
  if (type === "permission.asked" || type === "permission.updated") return ["waiting_permission", "permission"]
  if (type === "permission.replied") {
    const response = String(properties.response || "").toLowerCase()
    return [/(deny|reject|cancel)/.test(response) ? "idle" : "thinking", "permission_replied"]
  }
  if (type === "question.asked") return ["waiting_permission", "question"]
  if (type === "question.replied") return ["thinking", "question_replied"]
  if (type === "tool.execute.before") return ["tool_use", "tool_before"]
  if (type === "tool.execute.after") return ["working", "tool_after"]
  if (type === "session.created") return ["idle", "session_created"]
  if (type === "session.updated") return [previous || "idle", "session_updated"]
  return null
}

function responseData(response) {
  return response && typeof response === "object" && "data" in response ? response.data : response
}

function statusEntries(value) {
  const data = responseData(value)
  if (Array.isArray(data)) {
    return data.map((item) => {
      if (!item || typeof item !== "object") return null
      const id = item.sessionID || item.sessionId || item.id
      return id ? [id, item.status || item] : null
    }).filter(Boolean)
  }
  return data && typeof data === "object" ? Object.entries(data) : []
}

function createReporter(context, onClose = () => {}) {
  const sessions = new Map()
  const contextSessionID = context.sessionID || context.sessionId || context.session?.id || context.session?.sessionID || null
  const fallbackProject = context.directory || context.worktree || context.project?.worktree || context.project?.id || null
  let closed = false
  let closing = null

  function report(input, state, reason, tool, updatedAt) {
    if (closed) return Promise.resolve(false)
    const properties = input?.properties || input || {}
    const info = properties.info || input?.info || {}
    const sessionId = properties.sessionID || properties.sessionId || info.id || input?.sessionID || input?.sessionId || contextSessionID
    if (!sessionId) return Promise.resolve(false)
    if (!sessions.has(sessionId) && sessions.size >= 64) {
      sessions.delete(sessions.keys().next().value)
    }
    const previous = sessions.get(sessionId) || {}
    const current = {
      source: "opencode",
      sessionId,
      parentSessionId: info.parentID || properties.parentSessionId || properties.parentID || previous.parentSessionId || null,
      project: info.directory || properties.cwd || properties.project || previous.project || fallbackProject,
      title: info.title || properties.title || previous.title || null,
      model: modelName(properties.model || info.model) || previous.model || null,
      state,
      startedAt: isoTime(info.time?.created) || previous.startedAt || null
    }
    sessions.set(sessionId, current)
    return sendStatus({
      ...current,
      reason,
      updatedAt: updatedAt || new Date().toISOString(),
      tool: tool || null,
      terminal
    })
  }

  function handleEvent(input) {
    if (closed) return
    const event = input?.event || input || {}
    const type = String(event.type || "").toLowerCase()
    const properties = event.properties || {}
    const info = properties.info || {}
    const sessionId = properties.sessionID || properties.sessionId || info.id
    const previous = sessionId ? sessions.get(sessionId)?.state : null
    const mapped = eventState(type, properties, previous)
    if (!mapped) return
    const permission = properties.permission || properties.question || {}
    const tool = type.startsWith("tool.execute.") ? properties.tool : permission.type || permission.name || null
    const cleanup = () => {
      if (type === "session.deleted" && sessionId) {
        const item = sessions.get(sessionId)
        if (item) removeRuntimeSnapshot(item)
        sessions.delete(sessionId)
      }
    }
    void report({ properties, info }, mapped[0], mapped[1], tool).then(cleanup, cleanup)
  }

  async function initialize() {
    for (let attempt = 0; attempt < 3; attempt += 1) {
      const call = async (method) => {
        try {
          return method ? await method() : null
        } catch {
          return null
        }
      }
      const [listed, statuses] = await Promise.all([
        call(context.client?.session?.list ? () => context.client.session.list() : null),
        call(context.client?.session?.status ? () => context.client.session.status() : null)
      ])
      const list = responseData(listed)
      const sessionMap = new Map((Array.isArray(list) ? list : []).filter(Boolean).map((item) => [item.id, item]))
      let candidates = statusEntries(statuses)
      if (!candidates.length) {
        candidates = (Array.isArray(list) ? list : [])
          .filter(Boolean)
          .sort((left, right) => Number(right.time?.updated || 0) - Number(left.time?.updated || 0))
          .slice(0, 16)
          .map((item) => [item.id, { type: "idle" }])
      }
      if (!candidates.length && contextSessionID) {
        candidates = [[contextSessionID, { type: "idle" }]]
      }
      if (candidates.length) {
        await Promise.all(candidates.map(([id, status]) => {
          const info = sessionMap.get(id) || { id }
          const mapped = eventState("session.status", { status }, null) || ["idle", "initial_snapshot"]
          return report({ properties: { info, sessionID: id } }, mapped[0], "initial_snapshot", null, isoTime(info.time?.updated))
        }))
        return
      }
      await new Promise((resolve) => setTimeout(resolve, 250))
    }
  }

  function close() {
    if (closing) return closing
    closed = true
    const closingItems = [...sessions.values()]
    closing = (async () => {
      try {
        await Promise.all(closingItems.map((item) => sendStatus({
          ...item,
          state: "closed",
          reason: "process_exit",
          updatedAt: new Date().toISOString(),
          terminal
        }).catch(() => false)))
      } finally {
        closingItems.forEach((item) => removeRuntimeSnapshot(item))
        sessions.clear()
        onClose()
      }
    })()
    return closing
  }

  return {
    close,
    hooks: {
      event: handleEvent,
      "tool.execute.before": (input) => {
        void report(input, "tool_use", "tool_before", input?.tool)
      },
      "tool.execute.after": (input) => {
        void report(input, "working", "tool_after", input?.tool)
      }
    },
    initialize,
    sessions
  }
}

export const OpenCodeStatusReporter = async (context) => {
  if (instances.has(context)) return instances.get(context).hooks
  let reporter
  reporter = createReporter(context, () => {
    if (instances.get(context) === reporter) instances.delete(context)
  })
  instances.set(context, reporter)
  if (!lifecycleBound) {
    lifecycleBound = true
    process.once("beforeExit", () => {
      void Promise.all([...instances.values()].map((item) => item.close()))
    })
    process.once("exit", () => {
      for (const reporterInstance of instances.values()) {
        for (const item of reporterInstance.sessions.values()) {
          void sendStatus({ ...item, state: "closed", reason: "process_exit", updatedAt: new Date().toISOString(), terminal })
          removeRuntimeSnapshot(item)
        }
      }
    })
  }
  void reporter.initialize()
  return reporter.hooks
}

export const plugin = OpenCodeStatusReporter
export default OpenCodeStatusReporter
