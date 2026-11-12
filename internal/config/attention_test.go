package config

import (
	"encoding/json"
	"testing"
)

// The three-way completion policy replaced a boolean. The migration has to be
// exact in both directions, because getting it wrong either silences an alert
// the user asked for or invents one they turned off.
func TestDonePolicyMigration(t *testing.T) {
	cases := []struct {
		name string
		data string
		want string
	}{
		{"absent with old true means always", `{"notifyOnFinish":true}`, DonePolicyAlways},
		{"absent with old false means off", `{"notifyOnFinish":false}`, DonePolicyOff},
		{"explicit policy wins over the old boolean", `{"donePolicy":"background","notifyOnFinish":true}`, DonePolicyBackground},
		{"explicit always", `{"donePolicy":"always"}`, DonePolicyAlways},
		{"explicit off", `{"donePolicy":"off"}`, DonePolicyOff},
		{"unknown value falls back to the default", `{"donePolicy":"nonsense"}`, DonePolicyBackground},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := DecodeConfig([]byte(tc.data))
			if got.DonePolicy != tc.want {
				t.Fatalf("donePolicy = %q, want %q", got.DonePolicy, tc.want)
			}
			// The legacy field is kept in sync so an older build reading the same
			// file still sees a coherent answer.
			if wantOn := tc.want != DonePolicyOff; got.NotifyOnFinish != wantOn {
				t.Fatalf("notifyOnFinish = %v, want %v", got.NotifyOnFinish, wantOn)
			}
		})
	}
}

func TestDonePolicyDefaults(t *testing.T) {
	if Default().DonePolicy != DonePolicyBackground {
		t.Fatalf("the default policy must be background, got %q", Default().DonePolicy)
	}
	if Default().IdleMinutes != DefaultIdleMinutes {
		t.Fatalf("the default idle threshold must be %d, got %d", DefaultIdleMinutes, Default().IdleMinutes)
	}
	if len(DonePolicies) != 3 || DonePolicies[0] != DonePolicyBackground {
		t.Fatalf("the segment must offer three options, default first: %v", DonePolicies)
	}
}

func TestIdleMinutesClamp(t *testing.T) {
	for _, value := range []int{-1, MaxIdleMinutes + 1, 100000} {
		if got := (Config{IdleMinutes: value}).Normalize().IdleMinutes; got != DefaultIdleMinutes {
			t.Fatalf("idleMinutes %d should clamp to the default, got %d", value, got)
		}
	}
	if got := (Config{IdleMinutes: 0}).Normalize().IdleMinutes; got != 0 {
		t.Fatalf("0 means window focus only and must survive, got %d", got)
	}
	if got := (Config{IdleMinutes: 10}).Normalize().IdleMinutes; got != 10 {
		t.Fatalf("in-range value changed to %d", got)
	}
}

// A mute list is hand-editable, so every shape that is not a key is dropped
// rather than allowed to poison the whole list.
func TestNormalizeMutes(t *testing.T) {
	clean := NormalizeMutes([]string{
		SessionMuteKey("claude-hook", "ses_1"),
		"  " + SessionMuteKey("opencode", "ses_2") + " ",
		SessionMuteKey("claude-hook", "ses_1"), // duplicate
		"garbage-without-separator",
		"",
		SessionMuteKey("", "ses_3"), // no source, cannot be a key
	})
	if len(clean) != 2 {
		t.Fatalf("want 2 surviving keys, got %d: %v", len(clean), clean)
	}
	if clean[0] != SessionMuteKey("claude-hook", "ses_1") {
		t.Fatalf("first key was rewritten: %q", clean[0])
	}
	if clean[1] != SessionMuteKey("opencode", "ses_2") {
		t.Fatalf("trimming did not happen: %q", clean[1])
	}
}

func TestNormalizeMutesCaps(t *testing.T) {
	values := make([]string, 0, muteListMax*2)
	for index := 0; index < muteListMax*2; index++ {
		values = append(values, SessionMuteKey("src", "id"+itoaTest(index)))
	}
	clean := NormalizeMutes(values)
	if len(clean) != muteListMax {
		t.Fatalf("want the cap %d, got %d", muteListMax, len(clean))
	}
}

func itoaTest(value int) string {
	if value == 0 {
		return "0"
	}
	digits := make([]byte, 0, 6)
	for value > 0 {
		digits = append(digits, byte('0'+value%10))
		value /= 10
	}
	for left, right := 0, len(digits)-1; left < right; left, right = left+1, right-1 {
		digits[left], digits[right] = digits[right], digits[left]
	}
	return string(digits)
}

// The mute key must survive a round trip through config.json, because that is
// how it is stored.
func TestMuteListSurvivesJSON(t *testing.T) {
	value := Default()
	value.MutedSessions = []string{SessionMuteKey("claude-hook", "ses_1")}
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	got := DecodeConfig(data)
	if len(got.MutedSessions) != 1 || got.MutedSessions[0] != value.MutedSessions[0] {
		t.Fatalf("mute list did not survive: %v", got.MutedSessions)
	}
}
