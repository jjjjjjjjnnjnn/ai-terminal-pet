package ui

import "ai-terminal-pet/internal/session"

// panelKey is the platform-neutral set of keyboard navigation keys handled by
// the floating panel.
//
// Each window backend maps its OS key code (Win32 VK_*, X11 KeySym, macOS
// keyCode) onto one of these values and hands it to keyboardNav, so every
// platform gets identical navigation, wrap-around, identity tracking, and
// Enter-to-jump rules.
type panelKey int

const (
	panelKeyNone panelKey = iota
	panelKeyEscape
	panelKeyUp
	panelKeyDown
	panelKeyHome
	panelKeyEnd
	panelKeyEnter
)

// panelKeyOutcome describes the effect one key press should have on the panel.
type panelKeyOutcome struct {
	toggleExpanded bool
	jump           bool
	jumpRow        int
	repainted      bool
}

// keyboardNav tracks the keyboard-selected session row across refreshes.
//
// The zero value has no active selection: the focus ring never appears merely
// because the window opened, only after the user actually presses a navigation
// key. Storing the session identity alongside the index keeps an incoming
// refresh from silently rebinding Enter to a different terminal when rows
// reorder.
type keyboardNav struct {
	active   bool
	index    int
	identity string
}

// clear drops any active keyboard selection.
func (k *keyboardNav) clear() bool {
	if !k.active && k.identity == "" {
		return false
	}
	k.active = false
	k.index = 0
	k.identity = ""
	return true
}

// selected returns the currently selected visible row index, or (0, false) when
// no keyboard selection is active.
func (k *keyboardNav) selected() (int, bool) {
	if !k.active || k.index < 0 {
		return 0, false
	}
	return k.index, true
}

// sync reconciles the remembered selection against the latest sorted sessions
// and expanded state. Collapsing clears the selection; reordering follows the
// remembered session identity within the visible window.
func (k *keyboardNav) sync(expanded bool, items []session.Session) bool {
	if !expanded {
		return k.clear()
	}
	if !k.active {
		return false
	}
	limit := minInt(len(items), expandedMaxRows)
	if limit <= 0 {
		return k.clear()
	}
	if k.identity != "" {
		for i := 0; i < limit; i++ {
			if session.Identity(items[i]) == k.identity {
				changed := k.index != i
				k.index = i
				return changed
			}
		}
	}
	return k.clear()
}

// handle applies one key press to the panel navigation state.
func (k *keyboardNav) handle(key panelKey, expanded bool, items []session.Session) panelKeyOutcome {
	limit := minInt(len(items), expandedMaxRows)
	switch key {
	case panelKeyEscape:
		if !expanded {
			return panelKeyOutcome{}
		}
		k.clear()
		return panelKeyOutcome{toggleExpanded: true, repainted: true}
	case panelKeyDown:
		if !expanded {
			if limit <= 0 {
				return panelKeyOutcome{}
			}
			k.selectAt(0, items)
			return panelKeyOutcome{toggleExpanded: true, repainted: true}
		}
		if limit <= 0 {
			return panelKeyOutcome{}
		}
		next := 0
		if k.active {
			next = (k.index + 1) % limit
		}
		k.selectAt(next, items)
		return panelKeyOutcome{repainted: true}
	case panelKeyUp:
		if !expanded || limit <= 0 {
			return panelKeyOutcome{}
		}
		prev := limit - 1
		if k.active {
			prev = (k.index - 1 + limit) % limit
		}
		k.selectAt(prev, items)
		return panelKeyOutcome{repainted: true}
	case panelKeyHome:
		if !expanded || limit <= 0 {
			return panelKeyOutcome{}
		}
		changed := !k.active || k.index != 0
		k.selectAt(0, items)
		return panelKeyOutcome{repainted: changed}
	case panelKeyEnd:
		if !expanded || limit <= 0 {
			return panelKeyOutcome{}
		}
		last := limit - 1
		changed := !k.active || k.index != last
		k.selectAt(last, items)
		return panelKeyOutcome{repainted: changed}
	case panelKeyEnter:
		if expanded && k.active && k.index >= 0 && k.index < limit {
			return panelKeyOutcome{jump: true, jumpRow: k.index}
		}
		if _, ok := mostUrgentJump(items); ok {
			return panelKeyOutcome{jump: true, jumpRow: -1}
		}
		return panelKeyOutcome{}
	default:
		return panelKeyOutcome{}
	}
}

func (k *keyboardNav) selectAt(index int, items []session.Session) {
	k.active = true
	k.index = index
	k.identity = session.Identity(items[index])
}
