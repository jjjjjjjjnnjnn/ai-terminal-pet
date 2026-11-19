// Host-backed reconciliation for the OpenCode reporter.
//
// The panel treats a snapshot file as a conversation's existence. That is the
// right rule for a *pipe* reporter, whose file is deleted when its process
// exits, but wrong for a plugin that lives in `opencode serve --service` for
// days: the service keeps every conversation it has ever seen, so idle rows
// from yesterday stay on the overlay until the next reboot. The user sees
// "16 sessions" while five terminals are open.
//
// This module answers one narrow question: given what the HOST told us about a
// conversation, should the reporter stop showing it? It never guesses from a
// file's age and never closes anything on an uncertain answer.
//
//   - archived  -> the host's own record says the conversation is archived.
//   - settled child -> a subagent session is a step of its parent's turn, not
//     a conversation the user can switch to in a terminal.
//   - stale idle -> idle and not viewed in a terminal for STALE_IDLE_MS.
//
// A busy/waiting session is never closed here, however old its view is: that
// is exactly the case where the user still needs the row.

export const STALE_IDLE_MS = 2 * 60 * 60 * 1000
export const SWEEP_INTERVAL_MS = 2 * 60 * 1000

// viewedAtFrom reads the host's "last viewed" time. OpenCode stores it as
// `time_viewed` and exposes it on the session info as `time.viewed`; a couple
// of other shapes are accepted because this is diagnostic metadata, and a
// mis-shaped value must mean "unknown" (keep the row), never "stale".
export function viewedAtFrom(info) {
  if (!info || typeof info !== "object") return 0
  const time = info.time && typeof info.time === "object" ? info.time : {}
  for (const value of [time.viewed, time.viewedAt, info.viewedAt]) {
    if (typeof value === "number" && Number.isFinite(value) && value > 0) return value
    if (typeof value === "string") {
      const parsed = Date.parse(value)
      if (Number.isFinite(parsed) && parsed > 0) return parsed
    }
  }
  return 0
}

// staleReason returns a diagnostic token when a tracked session should be
// closed, or null when it must stay. `now`, `viewedAt` and `staleAfter` are
// supplied by the caller so every branch is table-testable without a clock.
export function staleReason(meta, now, options = {}) {
  if (!meta || typeof meta !== "object") return null
  if (meta.archived) return "session_archived"
  if (meta.state !== "idle") return null
  if (meta.parentSessionId) return "subagent_settled"
  const viewedAt = meta.viewedAt || 0
  if (!viewedAt) return null
  const staleAfter = options.staleAfter ?? STALE_IDLE_MS
  if (now - viewedAt > staleAfter) return "stale_idle"
  return null
}

// collectStale is the sweep over one tracker map. It returns [sessionId,
// reason] pairs and mutates nothing; the caller reports `closed` and removes
// the entries only after this function has decided.
export function collectStale(sessions, now, options = {}) {
  const result = []
  if (!sessions || typeof sessions.forEach !== "function") return result
  sessions.forEach((meta, sessionId) => {
    const reason = staleReason(meta, now, options)
    if (reason) result.push([sessionId, reason])
  })
  return result
}

// applyViewedAt folds a host session record into the tracker's freshness mark.
// It returns true only when the mark actually moved, so a metadata refresh
// does not rewrite a snapshot for no reason.
export function applyViewedAt(meta, info) {
  if (!meta || typeof meta !== "object") return false
  const viewedAt = viewedAtFrom(info)
  if (!viewedAt || viewedAt === meta.viewedAt) return false
  meta.viewedAt = viewedAt
  return true
}
