package ui

// The per-session silence list and the unread marks.
//
// Both are "what the user decided about a conversation", which is why neither
// is keyed on the store's key: that key includes the pid, and a hook reporter
// hands out a synthetic pid per process. Keying on it would lose every mute at
// the next hook invocation. session.Identity (source + session id) is the key,
// and config.SessionMuteKey builds the same string for the file.
//
// Muting is intentionally total for alerts: no balloon, no tone, no unread
// mark. The one thing it never suppresses is the auto-approve countdown drawn
// on the panel, because that is a safety control, not a notification - a user
// who muted a session still needs to be able to see that it is about to
// approve itself and veto it.

import (
	"sort"
	"strings"

	"ai-terminal-pet/internal/session"
)

// muteSet is the mute list plus the lookup the panel draws from.
type muteSet struct {
	keys map[string]bool
	// cached is the sorted slice handed to the settings window and the tray,
	// rebuilt only when the set changes.
	cached []string
}

func newMuteSet(keys []string) *muteSet {
	set := &muteSet{keys: make(map[string]bool, len(keys))}
	for _, key := range keys {
		if key == "" {
			continue
		}
		set.keys[key] = true
	}
	set.rebuild()
	return set
}

func (m *muteSet) rebuild() {
	m.cached = m.cached[:0]
	for key := range m.keys {
		m.cached = append(m.cached, key)
	}
	sort.Strings(m.cached)
}

// has reports whether this session is muted. An empty identity - a report with
// no source or no session id - is never muted: it cannot be keyed, so it cannot
// have been muted on purpose.
func (m *muteSet) has(value session.Session) bool {
	if m == nil {
		return false
	}
	identity := session.Identity(value)
	if identity == "" {
		return false
	}
	return m.keys[identity]
}

// set toggles one session and returns the new state.
func (m *muteSet) set(value session.Session, muted bool) bool {
	identity := session.Identity(value)
	if identity == "" {
		return false
	}
	if muted {
		m.keys[identity] = true
	} else {
		delete(m.keys, identity)
	}
	m.rebuild()
	return muted
}

// toggle flips one session and returns the state it now has.
func (m *muteSet) toggle(value session.Session) bool {
	if m == nil {
		return false
	}
	return m.set(value, !m.has(value))
}

// clear empties the list.
func (m *muteSet) clear() int {
	removed := len(m.keys)
	m.keys = make(map[string]bool, 0)
	m.rebuild()
	return removed
}

// list returns the stored keys in stable order.
func (m *muteSet) list() []string {
	if m == nil {
		return nil
	}
	keys := make([]string, len(m.cached))
	copy(keys, m.cached)
	return keys
}

func (m *muteSet) size() int {
	if m == nil {
		return 0
	}
	return len(m.keys)
}

// describeMuteKey renders a key for the settings window and the tray. It is
// the only place the separator is turned back into prose, so the encoding stays
// invisible everywhere else.
func describeMuteKey(key string) string {
	parts := strings.SplitN(key, "\x00", 2)
	if len(parts) != 2 {
		return key
	}
	return parts[0] + " · " + parts[1]
}

// unreadSet is the set of conversations with an unread alert.
//
// It is cleared by three events, and all three mean "the user has now seen
// this": the session is jumped to, it starts running again, or it disappears
// from the store. Anything else - collapsing the panel, a repaint, a reload -
// leaves it alone, because the mark is the only thing that survives an alert
// being suppressed by foreground awareness.
type unreadSet struct {
	keys map[string]bool
}

func newUnreadSet() *unreadSet {
	return &unreadSet{keys: make(map[string]bool, 4)}
}

func (u *unreadSet) mark(values []string) {
	if u == nil {
		return
	}
	for _, key := range values {
		if key != "" {
			u.keys[key] = true
		}
	}
}

func (u *unreadSet) clearOne(value session.Session) bool {
	if u == nil {
		return false
	}
	identity := session.Identity(value)
	if identity == "" {
		return false
	}
	if !u.keys[identity] {
		return false
	}
	delete(u.keys, identity)
	return true
}

func (u *unreadSet) has(value session.Session) bool {
	if u == nil {
		return false
	}
	identity := session.Identity(value)
	if identity == "" {
		return false
	}
	return u.keys[identity]
}

func (u *unreadSet) size() int {
	if u == nil {
		return 0
	}
	return len(u.keys)
}

// retain drops marks for conversations that are no longer in the list, so a
// closed session cannot leave a badge behind forever. It returns whether
// anything changed, which is what decides if a repaint is needed.
func (u *unreadSet) retain(items []session.Session) bool {
	if u == nil || len(u.keys) == 0 {
		return false
	}
	present := make(map[string]bool, len(items))
	for _, value := range items {
		if identity := session.Identity(value); identity != "" {
			present[identity] = true
		}
	}
	changed := false
	for key := range u.keys {
		if !present[key] {
			delete(u.keys, key)
			changed = true
		}
	}
	return changed
}

// unreadBits packs the unread marks into the frame key. Bit i is set when the
// i-th session in the drawn list is unread, so adding a mark changes the key and
// therefore the bitmap - the rule 1.0.1's frozen duration and 1.0.2's stale
// theme both learned the hard way.
func unreadBits(items []session.Session, unread *unreadSet) uint64 {
	if unread == nil || len(unread.keys) == 0 {
		return 0
	}
	var bits uint64
	for index, value := range items {
		if index >= 64 {
			break
		}
		if unread.has(value) {
			bits |= 1 << uint(index)
		}
	}
	return bits
}

// mutedBits is the same for the mute list: a muted row is drawn differently.
func mutedBits(items []session.Session, mutes *muteSet) uint64 {
	if mutes == nil || len(mutes.keys) == 0 {
		return 0
	}
	var bits uint64
	for index, value := range items {
		if index >= 64 {
			break
		}
		if mutes.has(value) {
			bits |= 1 << uint(index)
		}
	}
	return bits
}
