package ui

import (
	"bytes"
	"encoding/json"
	"os"
	"time"

	"ai-terminal-pet/internal/config"
	"ai-terminal-pet/internal/home"
	"ai-terminal-pet/internal/session"
)

type alertRecord struct {
	At        time.Time            `json:"at"`
	Sessions  []alertRecordSession `json:"sessions,omitempty"`
	Balloon   string               `json:"balloon,omitempty"`
	Tone      string               `json:"tone,omitempty"`
	Presence  string               `json:"presence"`
	Present   bool                 `json:"present"`
	Attending int                  `json:"attendingSessions,omitempty"`
	Skipped   string               `json:"skipped,omitempty"`
}

type alertRecordSession struct {
	Identity string `json:"identity"`
	State    string `json:"state"`
	Kind     string `json:"kind"`
	Decision string `json:"decision"`
	Unread   bool   `json:"unread,omitempty"`
}

const (
	alertReceiptName = "alerts.json"
	// alertLogName is the append-only history of scans that actually spoke.
	// alerts.json stays a one-scan snapshot because --json reads it; the two
	// answer different questions and must not share a shape.
	alertLogName = "alerts-log.jsonl"
	alertLogMax  = 500
)

// Why a scan said nothing. The receipt is diagnostics, so an absent balloon has
// to be explained rather than merely missing: "no-sessions" is the empty tick,
// "no-notification-surface" is a real alert on a machine where nothing could
// show it (see notificationSurfaceAvailable).
const (
	skipNoSessions            = "no-sessions"
	skipNoNotificationSurface = "no-notification-surface"
)

func alertReceiptPath() string {
	return home.Join(alertReceiptName)
}

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

// appendAlertLog keeps the last alertLogMax scans that actually spoke. A
// one-snapshot receipt cannot answer "why did it ring five minutes ago"; this
// can, without changing alerts.json's shape for --json.
func appendAlertLog(record alertRecord) {
	path := home.Join(alertLogName)
	if path == "" {
		return
	}
	encoded, err := json.Marshal(record)
	if err != nil {
		return
	}
	line := append(encoded, '\n')
	existing, err := os.ReadFile(path)
	if err == nil && len(existing) > 0 {
		kept := make([][]byte, 0, alertLogMax)
		for _, old := range bytes.Split(existing, []byte{'\n'}) {
			if len(bytes.TrimSpace(old)) > 0 {
				kept = append(kept, old)
			}
		}
		if len(kept) >= alertLogMax {
			kept = kept[len(kept)-alertLogMax+1:]
		}
		var buffer bytes.Buffer
		for _, old := range kept {
			buffer.Write(old)
			buffer.WriteByte('\n')
		}
		buffer.Write(line)
		line = buffer.Bytes()
	}
	_ = config.WriteFileAtomic(path, line)
}

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
		record.Skipped = skipNoSessions
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
		// The receipt says what happened, not what the policy wanted: a balloon
		// nobody could render is recorded as skipped, so --doctor cannot report
		// an alert surface the machine never had.
		if alertSurfaceAvailable() {
			record.Balloon = describeMuteKey(session.Identity(*plan.Balloon))
		} else {
			record.Skipped = joinSkipped(record.Skipped, skipNoNotificationSurface)
		}
	}
	record.Tone = plan.Tone
	return record
}

// joinSkipped keeps both reasons when a scan managed to have two.
func joinSkipped(existing, next string) string {
	if existing == "" {
		return next
	}
	return existing + "+" + next
}

func containsKey(values []string, key string) bool {
	for _, value := range values {
		if value == key {
			return true
		}
	}
	return false
}
