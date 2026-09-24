"use strict"

const { readStdin, sendStatus, stableProcessId } = require("./protocol.js")

function modelName(value) {
  if (typeof value === "string") return value
  if (!value || typeof value !== "object") return null
  const model = value.model || value.modelID || value.modelId || value.id || value.name
  const provider = value.provider || value.providerID || value.providerId
  return [provider, model].filter(Boolean).join("/") || null
}

function statusFor(input) {
  const hook = String(input.hook_event_name || input.hookEventName || "").toLowerCase()
  const notification = String(input.notification_type || input.notificationType || "").toLowerCase()
  if (hook.includes("sessionend") || notification.includes("session_end")) return ["closed", "session_end"]
  if (input.error) return ["error", "error"]
  if (hook.includes("permission") || /(permission|waiting|elicitation)/.test(notification)) return ["waiting_permission", "permission"]
  if (hook.includes("pretooluse") || notification.includes("tool")) return ["tool_use", "tool"]
  if (hook.includes("posttooluse") || input.stop_hook_active === true) return ["working", "tool_result"]
  if (hook.includes("userprompt")) return ["thinking", "prompt_submitted"]
  if (hook.includes("stop") || /(idle|complete|auth)/.test(notification)) return ["idle", "idle"]
  if (input.tool_name) return ["tool_use", "tool"]
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
  const sessionId = input.session_id || input.sessionId || "global"
  void sendStatus({
    source: "claude-hook",
    pid: input.pid || stableProcessId(`claude-hook:${sessionId}`),
    parentPid: input.parentPid || process.ppid,
    sessionId,
    parentSessionId: input.parent_session_id || input.parentSessionId || null,
    project: input.cwd || null,
    title: null,
    model: modelName(input.model),
    state: mapped[0],
    reason: mapped[1],
    startedAt: input.started_at || input.startedAt || new Date(Date.now() - process.uptime() * 1000).toISOString(),
    tool: input.tool_name || null,
    terminal: "claude"
  }, { retryMs: 750 }).catch(() => {})
  process.stdout.write("{}")
}

main().catch(() => {
  process.stdout.write("{}")
})
