"use strict"

const { approvalPlan, readStdin, sendStatus, stableProcessId, waitForApprovalDecision } = require("./protocol.js")

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

// toolText gathers what the blacklist sees: the tool name plus its input.
// The input shape varies by Claude version, so stringify the lot rather
// than guessing field names - a missed field could auto-approve a
// destructive command.
function toolText(input) {
  const parts = []
  if (input.tool_name) parts.push(String(input.tool_name))
  if (input.tool_input != null) {
    try {
      parts.push(typeof input.tool_input === "string" ? input.tool_input : JSON.stringify(input.tool_input))
    } catch {
      parts.push(String(input.tool_input))
    }
  }
  return parts.join("\n")
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
  const pid = input.pid || stableProcessId(`claude-hook:${sessionId}`)
  const base = {
    source: "claude-hook",
    pid,
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
  }

  // The veto window applies to the permission-prompt path only. Plain
  // tool observations never block (see protocol.js).
  if (mapped[0] === "waiting_permission") {
    const plan = approvalPlan(toolText(input))
    if (plan.mode === "auto") {
      // Report first (unawaited): the overlay needs the session now to
      // start its countdown while this hook waits out the veto window.
      void sendStatus(base, { retryMs: 750 }).catch(() => {})
      const decision = await waitForApprovalDecision(
        { source: "claude-hook", sessionId: String(sessionId), pid, tool: input.tool_name || null },
        { timeoutMs: plan.timeoutMs }
      )
      process.stdout.write(decision === "deny" ? '{"decision":"deny"}' : '{"decision":"allow"}')
      return
    }
    if (plan.mode === "manual") {
      // Dangerous command with auto-approve on: alarm the panel, but
      // decide nothing here - the terminal's native prompt owns the
      // verdict. The overlay countdown (if any) expires harmlessly: this
      // hook is already gone and filed no pending request, so silence
      // approves nothing.
      void sendStatus({ ...base, reason: "permission_dangerous" }, { retryMs: 750 }).catch(() => {})
      process.stdout.write("{}")
      return
    }
  }

  void sendStatus(base, { retryMs: 750 }).catch(() => {})
  process.stdout.write("{}")
}

main().catch(() => {
  process.stdout.write("{}")
})
