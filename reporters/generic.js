#!/usr/bin/env node
"use strict"

// Generic reporter: a one-line bridge from any CLI, script or editor hook to
// the overlay.
//
// The other reporters in this directory are written against one specific
// tool's event shape. This one is deliberately shape-free - it takes the state
// on the command line, so supporting a new tool is a shell alias rather than a
// new plugin:
//
//   node reporters/generic.js --source=aider --state=working --project="$PWD"
//   node reporters/generic.js --source=codex --state=waiting_permission --reason=approve
//   node reporters/generic.js --source=mybot --stdin        # JSON object on stdin
//
// Contract (same as every other reporter): fire and forget. A missing pipe, a
// dead overlay or a malformed payload never throws into the caller and never
// changes the exit code - a status reporter must not be able to break the tool
// it is reporting on.
//
// Exit codes: 0 for anything that was attempted, 2 for bad usage only. Bad
// usage is a human typo, not a runtime failure, and no host tool passes
// arguments this script does not recognise.

const { STATES, sendStatus, stableProcessId } = require("./protocol.js")

const USAGE = `generic reporter for ai-terminal-pet

usage:
  node generic.js --state=<state> [options]
  node generic.js --stdin [options]

states: idle thinking working tool_use waiting_permission error closed

options:
  --source <name>     reporting tool name (default: generic)
  --session <id>      session id (default: generic)
  --project <path>    project or working directory
  --title <text>      human readable title
  --reason <text>     short reason shown under the state
  --tool <name>       tool currently in use
  --model <name>      model name
  --terminal <name>   terminal name
  --pid <n>           process id to attribute the session to
  --stdin             read a JSON object from stdin and merge it over the flags
  --quiet             print nothing on success
  --help              print this message
`

// parseArgs is a tiny flag reader. Both `--name value` and `--name=value` are
// accepted, because shell one-liners overwhelmingly use the `=` form and
// getting that wrong is a confusing failure at exactly the moment the user is
// wiring up a new tool.
//
// Unknown flags are a hard error: they are always a typo at a shell prompt,
// never something a host tool emits.
function parseArgs(argv) {
  const values = {}
  const flags = new Set()
  for (let index = 0; index < argv.length; index += 1) {
    const token = argv[index]
    if (!token.startsWith("--")) {
      return { error: `unexpected argument ${JSON.stringify(token)}` }
    }
    const separator = token.indexOf("=")
    const name = separator === -1 ? token.slice(2) : token.slice(2, separator)
    if (name === "help" || name === "h") {
      flags.add("help")
      continue
    }
    if (name === "stdin" || name === "quiet") {
      flags.add(name)
      continue
    }
    if (separator !== -1) {
      values[name] = token.slice(separator + 1)
      continue
    }
    const value = argv[index + 1]
    if (value === undefined || value.startsWith("--")) {
      return { error: `option --${name} needs a value` }
    }
    values[name] = value
    index += 1
  }
  return { values, flags }
}

function readStdin() {
  return new Promise((resolve) => {
    if (process.stdin.isTTY) {
      resolve(null)
      return
    }
    const chunks = []
    let size = 0
    let settled = false
    const finish = () => {
      if (settled) return
      settled = true
      clearTimeout(timer)
      resolve(Buffer.concat(chunks, size).toString("utf8"))
    }
    const timer = setTimeout(finish, 2000)
    process.stdin.on("data", (chunk) => {
      const buffer = Buffer.isBuffer(chunk) ? chunk : Buffer.from(String(chunk))
      if (size < 1 << 20) chunks.push(buffer)
      size += buffer.length
      if (size > 1 << 20) finish()
    })
    process.stdin.once("end", finish)
    process.stdin.once("error", finish)
    process.stdin.once("close", finish)
    process.stdin.resume()
  })
}

function positiveInteger(value) {
  const number = Number(value)
  return Number.isSafeInteger(number) && number > 0 && number <= 0xffffffff ? number : null
}

async function main() {
  const parsed = parseArgs(process.argv.slice(2))
  if (parsed.error) {
    process.stderr.write(`generic.js: ${parsed.error}\n\n${USAGE}`)
    process.exitCode = 2
    return
  }
  const { values, flags } = parsed
  if (flags.has("help")) {
    process.stdout.write(USAGE)
    return
  }

  let payload = {}
  if (flags.has("stdin")) {
    const raw = await readStdin()
    if (raw && raw.trim()) {
      try {
        const decoded = JSON.parse(raw)
        if (decoded && typeof decoded === "object" && !Array.isArray(decoded)) payload = decoded
      } catch {
        // Malformed stdin is ignored rather than fatal: the caller is a host
        // tool and must never see an error from a status reporter.
      }
    }
  }

  const source = values.source || payload.source || "generic"
  const sessionId = values.session || payload.sessionId || "generic"
  const state = values.state || payload.state || "idle"
  if (!STATES.has(state)) {
    process.stderr.write(`generic.js: unknown state ${JSON.stringify(state)}\n\n${USAGE}`)
    process.exitCode = 2
    return
  }

  const pid =
    positiveInteger(values.pid) ||
    positiveInteger(payload.pid) ||
    stableProcessId(`generic:${source}:${sessionId}`)

  const now = new Date().toISOString()
  const sent = await sendStatus(
    {
      source,
      sessionId,
      pid,
      project: values.project || payload.project || null,
      title: values.title || payload.title || null,
      model: values.model || payload.model || null,
      state,
      reason: values.reason || payload.reason || "generic",
      tool: values.tool || payload.tool || null,
      terminal: values.terminal || payload.terminal || null,
      startedAt: payload.startedAt || now,
      updatedAt: now,
    },
    { retryMs: 1500 }
  )

  if (!flags.has("quiet")) {
    process.stdout.write(`${source} ${state} ${sent ? "ok" : "not delivered"}\n`)
  }
}

main().catch(() => {
  // Never surface an error to the host: a reporter that can break the tool it
  // reports on is worse than no reporter.
})
