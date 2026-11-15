package ui

import (
	"sync/atomic"
	"time"

	"ai-terminal-pet/internal/platform"
	"ai-terminal-pet/internal/session"
)

const attentionSampleInterval = 8 * time.Second

type attendanceState struct {
	Present   bool
	Attending map[string]bool
	Known     bool
}

type attendanceSampler struct {
	presence *platform.Presence
	sessions atomic.Pointer[[]session.Session]
	state    atomic.Pointer[attendanceState]
}

func newAttendanceSampler(presence *platform.Presence) *attendanceSampler {
	return &attendanceSampler{presence: presence}
}

func (s *attendanceSampler) publish(items []session.Session) {
	if s == nil {
		return
	}
	snapshot := make([]session.Session, len(items))
	copy(snapshot, items)
	s.sessions.Store(&snapshot)
}

func (s *attendanceSampler) Current() (attendanceState, bool) {
	if s == nil {
		return attendanceState{}, false
	}
	if value := s.state.Load(); value != nil {
		return *value, true
	}
	return attendanceState{}, false
}

func (s *attendanceSampler) run(stop <-chan struct{}, idleAfter func() time.Duration, wake func()) {
	if s == nil || s.presence == nil {
		return
	}
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
