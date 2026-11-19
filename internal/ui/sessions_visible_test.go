package ui

import (
	"testing"

	"ai-terminal-pet/internal/session"
)

func TestPanelRowsHideSettledSubagents(t *testing.T) {
	items := []session.Session{
		{SessionID: "parent", Source: "opencode", State: session.StateWorking},
		{SessionID: "child-done", Source: "opencode", ParentSessionID: "parent", State: session.StateIdle},
		{SessionID: "child-working", Source: "opencode", ParentSessionID: "parent", State: session.StateWorking},
		{SessionID: "child-asking", Source: "opencode", ParentSessionID: "parent", State: session.StateWaitingPermission, Reason: "permission_required"},
		{SessionID: "plain-idle", Source: "opencode", State: session.StateIdle},
	}
	got := panelRows(items)
	if len(got) != 3 {
		t.Fatalf("panelRows kept %d sessions, want 3: parent, urgent child, plain idle", len(got))
	}
	want := []string{"parent", "child-asking", "plain-idle"}
	for index, id := range want {
		if got[index].SessionID != id {
			t.Errorf("row %d = %q, want %q", index, got[index].SessionID, id)
		}
	}
}

func TestPanelRowsKeepsUrgentChildWithUnknownReason(t *testing.T) {
	items := []session.Session{
		{SessionID: "child-error", Source: "opencode", ParentSessionID: "parent", State: session.StateError},
	}
	if got := panelRows(items); len(got) != 1 {
		t.Fatal("a failed child must stay visible: its error needs a human")
	}
}
