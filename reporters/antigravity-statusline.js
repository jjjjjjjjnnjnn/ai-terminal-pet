"use strict"

const fs = require("node:fs")
const path = require("node:path")
const { spawn } = require("node:child_process")
const { executablePath, readStdin, sendStatus, stableProcessId } = require("./protocol.js")

const NOTIFIER_TIMEOUT_MS = 1000

function notifierPath() {
  if (process.env.ANTIGRAVITY_NOTIFIER_PATH) return process.env.ANTIGRAVITY_NOTIFIER_PATH
  const home = process.env.USERPROFILE || process.env.HOME
  return home ? path.join(home, ".gemini", "config", "plugins", "antigravity-windows-notifier", "hooks", "statusline-notify.js") : null
}

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
  if (input.tool_confirmation_pending === true) return ["waiting_permission", "tool_confirmation_pending"]
  const raw = input.agent_state
  const value = String(raw && typeof raw === "object" ? raw.type || raw.state || raw.status || raw.name : raw || "").toLowerCase()
  if (/(closed|terminated|finished|complete|done|idle)/.test(value)) return ["idle", "agent_state"]
  if (/(error|failed|failure)/.test(value)) return ["error", "agent_state"]
  if (/(permission|confirm|waiting)/.test(value)) return ["waiting_permission", "agent_state"]
  if (/(tool)/.test(value)) return ["tool_use", "agent_state"]
  if (/(working|executing|busy|running|processing)/.test(value)) return ["working", "agent_state"]
  if (/(thinking|reasoning|generat|stream)/.test(value)) return ["thinking", "agent_state"]
  return ["idle", "agent_state"]
}

function forwardNotifier(raw) {
  try {
    const target = notifierPath()
    if (!target || !fs.existsSync(target)) return Promise.resolve()
    try {
      const notifier = require(target)
      if (notifier && typeof notifier.handleRawInput === "function") {
        notifier.handleRawInput(raw)
        return Promise.resolve()
      }
    } catch {}
  } catch {
    return Promise.resolve()
  }
  let child
  try {
    const target = notifierPath()
    if (!target || !fs.existsSync(target)) return Promise.resolve()
    child = spawn(process.execPath, [target], { stdio: ["pipe", "pipe", "ignore"], windowsHide: true })
  } catch {
    return Promise.resolve()
  }
  if (!child || typeof child.once !== "function") return Promise.resolve()
  return new Promise((resolve) => {
    let settled = false
    let timer
    let outputErrorHandler

    function finish(terminate = false) {
      if (settled) return
      settled = true
      if (timer) clearTimeout(timer)
      if (outputErrorHandler) process.stdout.removeListener("error", outputErrorHandler)
      if (terminate) {
        if (typeof child.kill === "function") {
          try {
            child.kill()
          } catch {}
        }
        if (typeof child.unref === "function") {
          try {
            child.unref()
          } catch {}
        }
      }
      resolve()
    }

    function forward(chunk) {
      if (process.stdout.destroyed) {
        finish(true)
        return
      }
      try {
        process.stdout.write(chunk)
      } catch {
        finish(true)
      }
    }

    try {
      outputErrorHandler = () => finish(true)
      process.stdout.once("error", outputErrorHandler)
      timer = setTimeout(() => finish(true), NOTIFIER_TIMEOUT_MS)
      if (child.stdout && typeof child.stdout.on === "function") {
        child.stdout.on("data", forward)
        child.stdout.once("error", () => finish(true))
      }
      child.once("error", () => finish(true))
      child.once("close", () => finish())
      if (!child.stdin || typeof child.stdin.end !== "function") {
        finish(true)
        return
      }
      child.stdin.once("error", () => finish(true))
      child.stdin.end(raw)
    } catch {
      finish(true)
    }
  })
}

async function main() {
  const raw = await readStdin()
  let input = {}
  try {
    input = raw.length ? JSON.parse(raw.toString("utf8")) : {}
    if (!input || typeof input !== "object" || Array.isArray(input)) input = {}
  } catch {}
  const mapped = statusFor(input)
  const sessionId = input.conversation_id || input.session_id || input.sessionId || "global"
  let applicationPresent = false
  try {
    const executable = executablePath()
    applicationPresent = !!executable && fs.existsSync(executable)
  } catch {}
  if (!applicationPresent) {
    await forwardNotifier(raw)
    return
  }
  try {
    await sendStatus({
      source: "antigravity-statusline",
      pid: input.pid || stableProcessId(`antigravity-statusline:${sessionId}`),
      parentPid: input.parentPid || process.ppid,
      sessionId,
      parentSessionId: input.parent_session_id || input.parentSessionId || null,
      project: workspacePath(input.workspace) || workspacePath(input.cwd),
      title: null,
      model: modelName(input.model),
      state: mapped[0],
      reason: mapped[1],
      startedAt: input.started_at || input.startedAt || input.session_start_time || null,
      tool: input.tool_name || null,
      terminal: "antigravity"
    }, { retryMs: 1500 })
  } catch {}
  await forwardNotifier(raw)
}

main().catch(() => {
  process.exitCode = 0
})
