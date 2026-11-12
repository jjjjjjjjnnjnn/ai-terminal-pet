package ui

// The alert receipt: what the last scan decided, and why.
//
// Why a file at all: every "I did not get notified" report is indistinguishable
// from "the overlay was not running" unless the decision leaves a trace. The
// receipt is one small JSON record, rewritten (not appended) on each scan that
// changes anything, so it cannot grow, and it answers three questions from a
// bug report clipboard:
//
//   - did the scanner run at all, and when;
//   - which sessions were considered, and what kind each one is;
//   - for each of them, why it did or did not speak.
//
// It deliberately records kinds and reasons, not user text: a reason field can
// carry a command or a prompt, and this file must stay safe to paste.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"ai-terminal-pet/internal/config"
	"ai-terminal-pet/internal/session"
)

// alertRecord is the serialised receipt. Field names are stable: `--json`
// prints the same object under "alerts".
type alertRecord struct {
	// At is when this scan ran.
	At time.Time `json:"at"`
	// Sessions records one line per considered session.
	Sessions []alertRecordSession `json:"sessions,omitempty"`
	// Balloon names the conversation the balloon was about, or "".
	Balloon string `json:"balloon,omitempty"`
	// Tone is the timbre played, or "".
	Tone string `json:"tone,omitempty"`
	// Presence says whether the desktop sample was usable.
	Presence string `json:"presence"`
	// Present / Attending are the two facts the policy read.
	Present   bool `json:"present"`
	Attending int  `json:"attendingSessions,omitempty"`
	// Skipped names why a whole scan did nothing ("muted", "quiet", "no-sessions").
	Skipped string `json:"skipped,omitempty"`
}

type alertRecordSession struct {
	// Identity is "<source> · <sessionId>"; it is not the store key because
	// that includes a synthetic pid.
	Identity string `json:"identity"`
	State    string `json:"state"`
	Kind     string `json:"kind"`
	// Decision is the policy's own reason token.
	Decision string `json:"decision"`
	Unread   bool   `json:"unread,omitempty"`
}

// alertReceiptPath is runtime/alerts.json, next to the reporter snapshots so a
// `--json` reader finds it without a second directory convention.
func alertReceiptPath() string {
	base := os.Getenv("LOCALAPPDATA")
	if base == "" {
		return ""
	}
	return filepath.Join(base, "AI-Terminal-Pet", "runtime", "alerts.json")
}

// writeAlertRecord persists the receipt. Every failure is ignored on purpose:
// the receipt is diagnostics, and a machine with a read-only runtime directory
// must still get its balloons.
func writeAlertRecord(record alertRecord) {
	path := alertReceiptPath()
	if path == "" {
		return
	}
	data, err := json.Marshal(record)
	if err != nil {
		return
	}
	_ = config.WriteFileAtomic(path, append(data, '\n'))
}

// readAlertRecord is what `--json` and `--doctor` read. A missing or corrupt
// file is "no evidence", reported as nil rather than as an error: an overlay
// that has never run has no receipt, and that is not a fault.
func readAlertRecord() *alertRecord {
	path := alertReceiptPath()
	if path == "" {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var record alertRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return nil
	}
	return &record
}

// buildAlertRecord turns one scan into the persisted form.
func buildAlertRecord(at time.Time, inputs []AlertInput, plan AlertPlan, sample attendanceState, sampleKnown bool) alertRecord {
	record := alertRecord{
		At:       at.UTC(),
		Presence: "unknown",
	}
	if sampleKnown {
		if sample.Present {
			record.Presence = "present"
		} else {
			record.Presence = "away"
		}
		record.Present = sample.Present
		attending := 0
		for _, value := range sample.Attending {
			if value {
				attending++
			}
		}
		record.Attending = attending
	}
	if len(inputs) == 0 {
		record.Skipped = "no-sessions"
	}
	for index := range inputs {
		in := inputs[index]
		key := sessionAlertKey(in.Session)
		record.Sessions = append(record.Sessions, alertRecordSession{
			Identity: describeMuteKey(session.Identity(in.Session)),
			State:    string(in.Session.State),
			Kind:     in.Kind.Label(),
			Decision: plan.Reasons[key],
			Unread:   plan.Reasons[key] != "" && containsKey(plan.Unread, key),
		})
	}
	if plan.Balloon != nil {
		record.Balloon = describeMuteKey(session.Identity(*plan.Balloon))
	}
	record.Tone = plan.Tone
	return record
}

func containsKey(values []string, key string) bool {
	for _, value := range values {
		if value == key {
			return true
		}
	}
	return false
}
