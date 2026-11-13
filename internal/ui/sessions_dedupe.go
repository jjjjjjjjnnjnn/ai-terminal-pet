package ui

// One conversation, one row.
//
// The store keys a session on source + sessionId + pid, because the pipe, the
// snapshot files and the tombstone ledger are all per reporter process. That is
// correct for bookkeeping and wrong for the panel: a conversation reported by
// two pids - an OpenCode restart, a plugin that reports its own pid, a stale
// snapshot from a previous run - produced two rows for one conversation, and
// the panel is a picture of conversations, not of processes.
//
// The dedupe is deliberately here rather than in internal/store: the store's
// key is a protocol-level identity (docs/PROTOCOL.md §4), the panel's is a
// user-level one, and --json keeps the per-process rows on purpose so a bug
// report can still show that two reporters disagreed.

import (
	"ai-terminal-pet/internal/session"
)

// dedupeSessions keeps one row per conversation, the most recently updated one.
//
// Ties are broken by state priority and then by pid, so the result is stable:
// dedupe runs on every load, and a row that flips between two identical
// candidates on every scan would repaint forever.
func dedupeSessions(items []session.Session) []session.Session {
	if len(items) < 2 {
		return items
	}
	best := make(map[string]int, len(items))
	kept := make([]session.Session, 0, len(items))
	for _, value := range items {
		identity := session.Identity(value)
		if identity == "" {
			// No conversation identity: never merge, because merging two
			// unrelated rows would hide one of them.
			kept = append(kept, value)
			continue
		}
		index, seen := best[identity]
		if !seen {
			best[identity] = len(kept)
			kept = append(kept, value)
			continue
		}
		if betterCandidate(value, kept[index]) {
			kept[index] = value
		}
	}
	return kept
}

// betterCandidate reports whether the newer row should replace the older one.
func betterCandidate(candidate, current session.Session) bool {
	if !candidate.UpdatedAt.Equal(current.UpdatedAt) {
		return candidate.UpdatedAt.After(current.UpdatedAt)
	}
	if candidate.State.Priority() != current.State.Priority() {
		// Same timestamp, different story: the more urgent state is the one
		// the user needs to see. Two reporters disagreeing about one
		// conversation must not hide a blocked one behind a busy one.
		return candidate.State.Priority() > current.State.Priority()
	}
	// Everything else equal: a larger pid is the later process, which is the
	// safer guess for "the reporter that is still running".
	return candidate.PID > current.PID
}
