"use strict"

const { readStdin, sendStatus, stableProcessId } = require("./protocol.js")

function modelName(value) {
  if (typeof value === "string") return value
  if (!value || typeof value !== "object") return null
  const model = value.model || value.modelID || value.modelId || value.id || value.name
  const provider = value.provider || value.providerID || value.providerId
  return [provider, model].filter(Boolean).join("/") || null
}

function workspacePath(value) {
  if (typeof value === "string") return value
  if (!value || typeof value !== "object") return null
  return value.cwd || value.current_dir || value.project_dir || value.path || null
}

function statusFor(input) {
  const termination = input.terminationReason || input.termination_reason
  if (typeof termination === "string" && termination.trim()) return ["closed", "termination_reason"]
  if (input.error) return ["error", "error"]
  if (input.fullyIdle === true || input.fully_idle === true) return ["idle", "fully_idle"]
  const tool = input.toolCall && typeof input.toolCall === "object" ? input.toolCall.name : null
  if (tool) return ["tool_use", "tool_call"]
  return ["idle", "hook"]
}

async function main() {
  const raw = (await readStdin()).toString("utf8")
  let input = {}
  try {
    input = raw.trim() ? JSON.parse(raw) : {}
    if (!input || typeof input !== "object" || Array.isArray(input)) input = {}
  } catch {}
  const mapped = statusFor(input)
  const sessionId = input.conversation_id || input.session_id || input.sessionId || "global"
  void sendStatus({
    source: "antigravity-hook",
    pid: input.pid || stableProcessId(`antigravity-hook:${sessionId}`),
    parentPid: input.parentPid || process.ppid,
    sessionId,
    parentSessionId: input.parent_session_id || input.parentSessionId || null,
    project: workspacePath(input.workspace) || workspacePath(input.cwd),
    title: null,
    model: modelName(input.model),
    state: mapped[0],
    reason: mapped[1],
    startedAt: input.started_at || input.startedAt || input.session_start_time || null,
    tool: input.toolCall?.name || null,
    terminal: "antigravity"
  }, { retryMs: 750 }).catch(() => {})
  process.stdout.write('{"decision":"allow"}')
}

main().catch(() => {
  process.stdout.write('{"decision":"allow"}')
})
