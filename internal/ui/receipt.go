package ui

import (
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

const alertReceiptName = "alerts.json"

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
