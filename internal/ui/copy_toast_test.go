package ui

// The receipt now speaks for two gestures, not one. These tests pin the part
// that made the third state necessary: a failed jump must not borrow the copy
// sentence, and a pill that says it must stay up long enough to be read.

import (
	"testing"
	"time"

	"ai-terminal-pet/internal/i18n"
)

func TestToastKindsHaveDifferentSentences(t *testing.T) {
	labels := map[string]bool{}
	for _, lang := range []i18n.Lang{i18n.LangEn, i18n.LangZhCN} {
		for _, kind := range []toastKind{toastCopied, toastCopyFailed, toastJumpFailed} {
			label := copyToastLabel(panelToast{row: 0, tick: 10, kind: kind}, lang)
			if label == "" {
				t.Fatalf("%s: kind %d has no label", lang, kind)
			}
			labels[string(lang)+"/"+label] = true
		}
	}
	// Six distinct sentences: three kinds times two languages.
	if len(labels) != 6 {
		t.Fatalf("the three kinds share copy: %v", labels)
	}
}

func TestJumpFailedIsNotTheCopySentence(t *testing.T) {
	got := copyToastLabel(panelToast{tick: 10, kind: toastJumpFailed}, i18n.LangEn)
	if want := i18n.T(i18n.LangEn, i18n.KeyJumpFailed); got != want {
		t.Fatalf("jump label = %q, want %q", got, want)
	}
	if got == i18n.T(i18n.LangEn, i18n.KeyCopyFailed) {
		t.Fatal("a failed jump must not say the copy failed")
	}
}

func TestToastKindsMoveTheFrameKey(t *testing.T) {
	keys := []renderedKey{}
	for _, kind := range []toastKind{toastCopied, toastCopyFailed, toastJumpFailed} {
		var key renderedKey
		key.withToast(panelToast{row: 0, tick: 5, kind: kind})
		keys = append(keys, key)
	}
	for i := range keys {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] == keys[j] {
				t.Fatalf("kinds %d and %d produce the same frame key", i, j)
			}
		}
	}
}

// The jump pill gets the longer lifetime: it is three words and it arrives only
// after the gesture already failed, so there is no second signal behind it.
func TestJumpToastOutlivesTheCopyToast(t *testing.T) {
	start := time.Unix(1000, 0)
	flagship := copyToastDuration + 100*time.Millisecond

	var copyPill copyToast
	copyPill.show(0, "claude-hook · ses_1", toastCopied, start)
	if copyPill.active(start.Add(flagship)) {
		t.Fatal("the copy pill should be gone by now")
	}

	var jumpPill copyToast
	jumpPill.show(0, "claude-hook · ses_1", toastJumpFailed, start)
	if !jumpPill.active(start.Add(flagship)) {
		t.Fatal("the jump pill must still be readable after the copy one expired")
	}
	if jumpPill.active(start.Add(jumpToastDuration)) {
		t.Fatal("the jump pill outlived its own duration")
	}
}

func TestOnlyFailedKindsUseTheErrorColour(t *testing.T) {
	if toastColour(panelToast{kind: toastCopied}) != colors.Primary {
		t.Fatal("a successful copy is not an error")
	}
	for _, kind := range []toastKind{toastCopyFailed, toastJumpFailed} {
		if toastColour(panelToast{kind: kind}) != colors.Error {
			t.Fatalf("kind %d must be drawn in the error colour", kind)
		}
	}
}

func TestCopyToastKindFollowsTheClipboard(t *testing.T) {
	if got := copyToastKind(true); got != toastCopyFailed {
		t.Fatalf("copyToastKind(true) = %d", got)
	}
	if got := copyToastKind(false); got != toastCopied {
		t.Fatalf("copyToastKind(false) = %d", got)
	}
}

// Arming a receipt with no sentence must clear it rather than leave the previous
// pill counting down to a claim nobody made.
func TestShowingNothingClearsTheReceipt(t *testing.T) {
	start := time.Unix(1000, 0)
	var toast copyToast
	toast.show(0, "claude-hook · ses_1", toastCopied, start)
	if !toast.active(start) {
		t.Fatal("the receipt was not armed")
	}
	toast.show(0, "claude-hook · ses_1", toastNone, start)
	if toast.active(start) {
		t.Fatal("showing nothing must clear the receipt")
	}
}
