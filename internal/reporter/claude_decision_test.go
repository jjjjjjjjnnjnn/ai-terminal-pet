package reporter

import "testing"

func TestClaudeApprovalAnswerMatchesTheAcceptedShape(t *testing.T) {
	cases := []struct {
		name      string
		eventName string
		behavior  string
		want      string
	}{
		{
			name:      "allow names the event",
			eventName: "PermissionRequest",
			behavior:  "allow",
			want:      `{"hookSpecificOutput":{"hookEventName":"PermissionRequest","decision":{"behavior":"allow"}}}`,
		},
		{
			name:      "empty event defaults to PermissionRequest",
			eventName: "",
			behavior:  "allow",
			want:      `{"hookSpecificOutput":{"hookEventName":"PermissionRequest","decision":{"behavior":"allow"}}}`,
		},
		{
			name:      "deny carries a reason the terminal can show",
			eventName: "PermissionRequest",
			behavior:  "deny",
			want:      `{"hookSpecificOutput":{"hookEventName":"PermissionRequest","decision":{"behavior":"deny","message":"Denied from the AI Terminal Pet panel."}}}`,
		},
		{
			name:      "another event keeps its own name",
			eventName: "Notification",
			behavior:  "allow",
			want:      `{"hookSpecificOutput":{"hookEventName":"Notification","decision":{"behavior":"allow"}}}`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ClaudeApprovalAnswer(tc.eventName, tc.behavior); got != tc.want {
				t.Fatalf("got %s, want %s", got, tc.want)
			}
		})
	}
}

// The legacy shape is what this project used until 2026-11-18. It is kept here
// as a test rather than a comment so nobody "simplifies" the answer back to it:
// Claude Code 2.1.312 ignored it, which meant auto-approve looked enabled and
// silently did nothing.
func TestClaudeApprovalAnswerIsNotTheLegacyDialect(t *testing.T) {
	allow := ClaudeApprovalAnswer("PermissionRequest", "allow")
	if allow == `{"decision":"allow"}` {
		t.Fatal("the legacy decision dialect is ignored by Claude Code 2.1.312")
	}
	deny := ClaudeApprovalAnswer("PermissionRequest", "deny")
	if deny == `{"decision":"deny"}` {
		t.Fatal("the legacy decision dialect is ignored by Claude Code 2.1.312")
	}
}
