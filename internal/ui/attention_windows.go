package ui

// Wiring for the alert policy: who samples the desktop, and how the pure
// Decide/PlanAlerts pair is fed.
//
// The expensive part - "which window is in front, and is that window the one
// this session lives in" - cannot run on the UI thread. Enumerating other
// processes' windows can stall behind a hung window, and a stalled window
// procedure is exactly what Windows reports as AppHang. So the sample runs on
// its own goroutine at a deliberately slow cadence, and the UI thread only ever
// reads the last published answer.
//
// The cadence is 8 seconds rather than the 50 ms animation tick for one reason:
// two enumerations per second would be pure heat. Attention does not change
// faster than a person does, and the freshness that actually matters - "the
// user just Alt-Tabbed away" - also arrives as a reporter push, which triggers
// its own scan.

import (
	"sync/atomic"
	"time"

	"ai-terminal-pet/internal/platform"
	"ai-terminal-pet/internal/session"
)

const attentionSampleInterval = 8 * time.Second

// attendanceState is one published sample. It is written by the sampler and
// read by the UI thread, so it is a value type held in an atomic pointer: no
// lock, and no torn reads.
type attendanceState struct {
	// Present is true when the user has touched the machine recently.
	Present bool
	// Attending maps session.Identity to "this session's window is in front".
	Attending map[string]bool
	// Known is false when the desktop could not be sampled at all. The policy
	// reads unknown as "not looking", never as "present".
	Known bool
}

// attendanceSampler owns the goroutine. It exists as a struct so the ticker can
// be stopped deterministically when the window goes away, rather than leaking a
// goroutine per overlay instance.
type attendanceSampler struct {
	presence *platform.Presence
	sessions atomic.Pointer[[]session.Session]
	state    atomic.Pointer[attendanceState]
}

func newAttendanceSampler(presence *platform.Presence) *attendanceSampler {
	return &attendanceSampler{presence: presence}
}

// publish stores the session list the sampler should evaluate. Called on the UI
// thread from load(), so the sampler never touches a.sessions itself.
func (s *attendanceSampler) publish(items []session.Session) {
	if s == nil {
		return
	}
	snapshot := make([]session.Session, len(items))
	copy(snapshot, items)
	s.sessions.Store(&snapshot)
}

// Current returns the last published sample, or false when nothing has been
// sampled yet - which the policy treats as unknown.
func (s *attendanceSampler) Current() (attendanceState, bool) {
	if s == nil {
		return attendanceState{}, false
	}
	if value := s.state.Load(); value != nil {
		return *value, true
	}
	return attendanceState{}, false
}

// run samples until stop closes. Each tick publishes one attendanceState and
// then calls wake, so the UI thread can repaint and plan an alert.
func (s *attendanceSampler) run(stop <-chan struct{}, idleAfter func() time.Duration, wake func()) {
	if s == nil || s.presence == nil {
		return
	}
	// One sample immediately: the first refresh after launch would otherwise
	// decide with "unknown", and an overlay that starts while a session is
	// already blocked should know whether the user is looking at it.
	s.sample(idleAfter())
	if wake != nil {
		wake()
	}
	ticker := time.NewTicker(attentionSampleInterval)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			s.sample(idleAfter())
			if wake != nil {
				wake()
			}
		}
	}
}

func (s *attendanceSampler) sample(idleAfter time.Duration) {
	sample := platform.SampleForeground()
	state := attendanceState{
		// Present means "the machine has been touched recently". An idle time
		// past the threshold means the user walked away from a window that is
		// still in front, and both halves have to be known for Present to be
		// true - so an unknown sample can never silence an alert.
		Present:   sample.IdleKnown && !sample.IdleAtLeast(idleAfter),
		Known:     sample.Known && sample.IdleKnown,
		Attending: map[string]bool{},
	}
	if items := s.sessions.Load(); items != nil && sample.Known {
		for _, value := range *items {
			identity := session.Identity(value)
			if identity == "" {
				continue
			}
			if s.presence.Attends(value, sample.Handle) {
				state.Attending[identity] = true
			}
		}
	}
	s.state.Store(&state)
}
