package store

import (
	"strconv"
	"sync"
	"time"

	"ai-terminal-pet/internal/session"
)

const (
	DefaultMaxSessions  = 256
	DefaultPipeTTL      = 2 * DefaultSnapshotTTL
	DefaultTombstoneTTL = 10 * time.Minute
	maxSessionsLimit    = 4096

	// DefaultAbsenceGrace is how long a session's host may be unobservable
	// before it is finally removed. A single failed observation never removes
	// anything: a process can be briefly invisible while a terminal restarts,
	// a tab is re-parented, or a permission-protected process refuses a
	// snapshot. Two minutes of continuous absence is the actual signal.
	DefaultAbsenceGrace = 2 * time.Minute
)

type Origin uint8

const (
	OriginPipe Origin = iota
	OriginSnapshot
)

type entry struct {
	value        session.Session
	snapshotSeen time.Time
	pipeSeen     time.Time
	lastSeen     time.Time
	// unseenSince is the start of the current "host not observable" streak.
	// It is cleared the moment the host is seen again, so a transient miss
	// cannot evict a live session.
	unseenSince time.Time
	hasSnapshot bool
	hasPipe     bool
}

type tombstoneReason uint8

const (
	// tombstoneClosed comes from an explicit closed event. Its timestamp
	// gate must stay: a late, stale push from a session that really ended
	// must not resurrect it.
	tombstoneClosed tombstoneReason = iota
	// tombstoneExpired comes from silence (TTL or a long unobservable host).
	// It must NOT gate re-entry: a quiet terminal that reappears with its
	// unchanged timestamp has to be able to come straight back. Blocking that
	// was the reason flickering sessions could not self-heal.
	tombstoneExpired
	// tombstoneEvicted comes from the session cap. It keeps the timestamp
	// gate: allowing an unchanged old event to immediately re-admit a session
	// we just evicted for capacity would make the list oscillate.
	tombstoneEvicted
)

type tombstone struct {
	updatedAt time.Time
	seenAt    time.Time
	reason    tombstoneReason
}

type Store struct {
	mutex       sync.RWMutex
	entries     map[string]entry
	tombstones  map[string]tombstone
	revision    uint64
	maxSessions int
	pipeTTL     time.Duration
}

func New() *Store {
	return NewWithLimits(DefaultMaxSessions, DefaultPipeTTL)
}

func NewWithMaxSessions(maxSessions int) *Store {
	return NewWithLimits(maxSessions, DefaultPipeTTL)
}

func NewWithLimits(maxSessions int, pipeTTL time.Duration) *Store {
	return &Store{
		entries:     make(map[string]entry),
		tombstones:  make(map[string]tombstone),
		maxSessions: boundedMaxSessions(maxSessions),
		pipeTTL:     positiveDuration(pipeTTL, DefaultPipeTTL),
	}
}

func (s *Store) UpsertPipe(value session.Session, now time.Time) bool {
	return s.upsert(value, OriginPipe, now)
}

func (s *Store) UpsertSnapshot(value session.Session, now time.Time) bool {
	return s.upsert(value, OriginSnapshot, now)
}

func (s *Store) List() []session.Session {
	s.mutex.RLock()
	items := make([]session.Session, 0, len(s.entries))
	for _, current := range s.entries {
		if current.value.State != session.StateClosed {
			items = append(items, current.value)
		}
	}
	s.mutex.RUnlock()
	session.Sort(items)
	return items
}

func (s *Store) Revision() uint64 {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.revision
}

func (s *Store) HasAvailable() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return len(s.entries) > 0
}

// Reconcile is the stable lifecycle primitive used by the running overlay.
//
// Cleanup (below) is the legacy heartbeat model: silence for one TTL means
// "gone". That is the wrong question for event-driven reporters, which stay
// quiet for as long as a session keeps the same state. Reconcile instead asks
// whether the host terminal is still observable. While it is, the session is
// kept indefinitely; only a continuous, grace-long absence removes it.
//
// Refreshing liveness deliberately does not change the payload (so UpdatedAt,
// the frame digest and relative timestamps stay put) and does not bump the
// revision (so no repaint is scheduled). A live but quiet terminal costs
// literally nothing between OS snapshots.
func (s *Store) Reconcile(now time.Time, alive func(session.Session) bool, grace time.Duration) bool {
	now = normalizedTime(now)
	grace = positiveDuration(grace, DefaultAbsenceGrace)

	// Observations are made outside the store lock: the liveness callback
	// consults a shared cached OS view and must never run while we hold it.
	type candidate struct {
		key   string
		value session.Session
	}
	s.mutex.RLock()
	candidates := make([]candidate, 0, len(s.entries))
	for entryKey, current := range s.entries {
		candidates = append(candidates, candidate{key: entryKey, value: current.value})
	}
	s.mutex.RUnlock()

	observed := make(map[string]bool, len(candidates))
	for _, item := range candidates {
		if alive != nil && item.value.State != session.StateClosed {
			observed[item.key] = alive(item.value)
		}
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.ensureLocked()
	changed := false
	for _, item := range candidates {
		current, exists := s.entries[item.key]
		if !exists {
			continue
		}
		if current.value.State == session.StateClosed {
			s.tombstoneActiveLocked(item.key, current, now)
			changed = true
			continue
		}
		if observed[item.key] {
			if now.After(current.pipeSeen) {
				current.pipeSeen = now
			}
			if now.After(current.snapshotSeen) {
				current.snapshotSeen = now
			}
			current.lastSeen = latestTime(current.lastSeen, now)
			current.unseenSince = time.Time{}
			s.entries[item.key] = current
			continue
		}
		if current.unseenSince.IsZero() {
			current.unseenSince = now
			s.entries[item.key] = current
			continue
		}
		if now.Sub(current.unseenSince) >= grace {
			s.tombstoneActiveLocked(item.key, current, now)
			changed = true
		}
	}
	if s.pruneTombstonesLocked(now) {
		changed = true
	}
	if s.enforceLimitLocked(now) {
		changed = true
	}
	if changed {
		s.revision++
	}
	return changed
}

func (s *Store) Cleanup(now time.Time, alive func(int) bool, snapshotTTL time.Duration) bool {
	s.mutex.RLock()
	pipeTTL := s.pipeTTL
	s.mutex.RUnlock()
	return s.cleanup(now, alive, snapshotTTL, pipeTTL)
}

func (s *Store) CleanupWithPipeTTL(now time.Time, alive func(int) bool, snapshotTTL, pipeTTL time.Duration) bool {
	return s.cleanup(now, alive, snapshotTTL, pipeTTL)
}

func (s *Store) cleanup(now time.Time, alive func(int) bool, snapshotTTL, pipeTTL time.Duration) bool {
	now = normalizedTime(now)
	snapshotTTL = positiveDuration(snapshotTTL, DefaultSnapshotTTL)
	pipeTTL = positiveDuration(pipeTTL, DefaultPipeTTL)

	s.mutex.RLock()
	candidates := make(map[string]entry, len(s.entries))
	for entryKey, current := range s.entries {
		candidates[entryKey] = current
	}
	pipeTTL = positiveDuration(pipeTTL, s.pipeTTL)
	s.mutex.RUnlock()

	dead := make(map[string]struct{})
	if alive != nil {
		for entryKey, current := range candidates {
			if processBacked(current.value.Source) && !alive(current.value.PID) {
				dead[entryKey] = struct{}{}
			}
		}
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.ensureLocked()
	changed := false
	for entryKey, current := range s.entries {
		remove := current.value.State == session.StateClosed
		if _, isDead := dead[entryKey]; isDead {
			remove = true
		}
		if !remove {
			switch {
			case current.hasSnapshot && current.snapshotSeen.After(current.pipeSeen):
				remove = expired(now, current.snapshotSeen, snapshotTTL)
			case current.hasPipe:
				remove = expired(now, current.pipeSeen, pipeTTL)
			}
		}
		if remove {
			s.tombstoneActiveLocked(entryKey, current, now)
			changed = true
		}
	}

	if s.pruneTombstonesLocked(now) {
		changed = true
	}
	if s.enforceLimitLocked(now) {
		changed = true
	}
	if changed {
		s.revision++
	}
	return changed
}

func (s *Store) upsert(value session.Session, origin Origin, now time.Time) bool {
	now = normalizedTime(now)
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.ensureLocked()
	if value.State == session.StateClosed {
		return s.closeLocked(value, now)
	}
	return s.openLocked(value, origin, now)
}

func (s *Store) openLocked(value session.Session, origin Origin, now time.Time) bool {
	entryKey := key(value)
	current, exists := s.entries[entryKey]
	closed, hasTombstone := s.tombstones[entryKey]
	if exists && value.UpdatedAt.Before(current.value.UpdatedAt) {
		return false
	}
	// Only an explicit close gates re-entry by timestamp. A silence-induced
	// tombstone must not: otherwise a session whose reporter went quiet can
	// never return, because its next push still carries the same UpdatedAt.
	if hasTombstone && closed.reason != tombstoneExpired && !value.UpdatedAt.After(closed.updatedAt) {
		return false
	}

	roomChanged := false
	if !exists && !hasTombstone {
		roomChanged = s.makeActiveRoomLocked(entryKey, now)
	}
	valueChanged := !exists || current.value != value
	next := current
	if !exists {
		next = entry{}
	}
	if hasTombstone {
		delete(s.tombstones, entryKey)
	}
	next.value = value
	switch origin {
	case OriginPipe:
		if !next.hasPipe || next.pipeSeen.IsZero() || now.After(next.pipeSeen) {
			next.pipeSeen = now
		}
		next.hasPipe = true
	case OriginSnapshot:
		if !next.hasSnapshot || next.snapshotSeen.IsZero() || valueChanged {
			next.snapshotSeen = latestTime(next.snapshotSeen, now)
		}
		next.hasSnapshot = true
	}
	next.lastSeen = latestTime(next.lastSeen, now)
	s.entries[entryKey] = next

	changed := valueChanged || hasTombstone || roomChanged
	if changed {
		s.revision++
	}
	return changed
}

func (s *Store) closeLocked(value session.Session, now time.Time) bool {
	entryKey := key(value)
	current, exists := s.entries[entryKey]
	closed, hasTombstone := s.tombstones[entryKey]
	if exists && value.UpdatedAt.Before(current.value.UpdatedAt) {
		return false
	}
	if hasTombstone && value.UpdatedAt.Before(closed.updatedAt) {
		return false
	}

	roomChanged := false
	if !hasTombstone {
		roomChanged = s.makeTombstoneRoomLocked(entryKey, now)
	}
	changed := exists || roomChanged
	if exists {
		delete(s.entries, entryKey)
	}
	if hasTombstone {
		if value.UpdatedAt.After(closed.updatedAt) {
			closed.updatedAt = value.UpdatedAt
			changed = true
		}
		closed.seenAt = latestTime(closed.seenAt, now)
		closed.reason = tombstoneClosed
		s.tombstones[entryKey] = closed
	} else {
		s.tombstones[entryKey] = tombstone{updatedAt: value.UpdatedAt, seenAt: now, reason: tombstoneClosed}
		changed = true
	}
	if changed {
		s.revision++
	}
	return changed
}

func (s *Store) tombstoneActiveLocked(entryKey string, current entry, now time.Time) {
	delete(s.entries, entryKey)
	if closed, exists := s.tombstones[entryKey]; exists {
		if current.value.UpdatedAt.After(closed.updatedAt) {
			closed.updatedAt = current.value.UpdatedAt
		}
		closed.seenAt = latestTime(closed.seenAt, now)
		s.tombstones[entryKey] = closed
		return
	}
	s.makeTombstoneRoomLocked(entryKey, now)
	s.tombstones[entryKey] = tombstone{updatedAt: current.value.UpdatedAt, seenAt: now, reason: tombstoneExpired}
}

func (s *Store) makeActiveRoomLocked(protected string, now time.Time) bool {
	changed := false
	for len(s.entries) >= s.maxSessions {
		entryKey, _, ok := s.oldestEntryLocked(protected)
		if !ok {
			break
		}
		current := s.entries[entryKey]
		delete(s.entries, entryKey)
		if s.makeTombstoneRoomLocked(entryKey, now) {
			changed = true
		}
		s.tombstones[entryKey] = tombstone{updatedAt: current.value.UpdatedAt, seenAt: now, reason: tombstoneEvicted}
		changed = true
	}
	return changed
}

func (s *Store) makeTombstoneRoomLocked(protected string, now time.Time) bool {
	changed := false
	for len(s.tombstones) >= s.maxSessions {
		entryKey, _, ok := s.oldestTombstoneLocked(protected)
		if !ok {
			break
		}
		delete(s.tombstones, entryKey)
		changed = true
	}
	return changed
}

func (s *Store) enforceLimitLocked(now time.Time) bool {
	changed := false
	for len(s.entries) > s.maxSessions {
		entryKey, _, ok := s.oldestEntryLocked("")
		if !ok {
			break
		}
		current := s.entries[entryKey]
		delete(s.entries, entryKey)
		if s.makeTombstoneRoomLocked(entryKey, now) {
			changed = true
		}
		s.tombstones[entryKey] = tombstone{updatedAt: current.value.UpdatedAt, seenAt: now, reason: tombstoneEvicted}
		changed = true
	}
	for len(s.tombstones) > s.maxSessions {
		entryKey, _, ok := s.oldestTombstoneLocked("")
		if !ok {
			break
		}
		delete(s.tombstones, entryKey)
		changed = true
	}
	return changed
}

func (s *Store) pruneTombstonesLocked(now time.Time) bool {
	changed := false
	for entryKey, current := range s.tombstones {
		if expired(now, current.seenAt, DefaultTombstoneTTL) {
			delete(s.tombstones, entryKey)
			changed = true
		}
	}
	return changed
}

func (s *Store) oldestEntryLocked(protected string) (string, time.Time, bool) {
	var selectedKey string
	var selectedSeen time.Time
	found := false
	for entryKey, current := range s.entries {
		if entryKey == protected {
			continue
		}
		seen := current.lastSeen
		if seen.IsZero() {
			seen = current.value.UpdatedAt
		}
		if !found || recordBefore(entryKey, seen, selectedKey, selectedSeen) {
			selectedKey = entryKey
			selectedSeen = seen
			found = true
		}
	}
	return selectedKey, selectedSeen, found
}

func (s *Store) oldestTombstoneLocked(protected string) (string, time.Time, bool) {
	var selectedKey string
	var selectedSeen time.Time
	found := false
	for entryKey, current := range s.tombstones {
		if entryKey == protected {
			continue
		}
		if !found || recordBefore(entryKey, current.seenAt, selectedKey, selectedSeen) {
			selectedKey = entryKey
			selectedSeen = current.seenAt
			found = true
		}
	}
	return selectedKey, selectedSeen, found
}

func (s *Store) ensureLocked() {
	if s.entries == nil {
		s.entries = make(map[string]entry)
	}
	if s.tombstones == nil {
		s.tombstones = make(map[string]tombstone)
	}
	if s.maxSessions <= 0 {
		s.maxSessions = DefaultMaxSessions
	}
	if s.maxSessions > maxSessionsLimit {
		s.maxSessions = maxSessionsLimit
	}
	if s.pipeTTL <= 0 {
		s.pipeTTL = DefaultPipeTTL
	}
}

func recordBefore(leftKey string, leftSeen time.Time, rightKey string, rightSeen time.Time) bool {
	if leftSeen.Before(rightSeen) {
		return true
	}
	if rightSeen.Before(leftSeen) {
		return false
	}
	return leftKey < rightKey
}

func boundedMaxSessions(value int) int {
	if value <= 0 {
		return DefaultMaxSessions
	}
	if value > maxSessionsLimit {
		return maxSessionsLimit
	}
	return value
}

func positiveDuration(value, fallback time.Duration) time.Duration {
	if value <= 0 {
		return fallback
	}
	return value
}

func normalizedTime(value time.Time) time.Time {
	if value.IsZero() {
		return time.Now()
	}
	return value
}

func latestTime(current, candidate time.Time) time.Time {
	if current.IsZero() || candidate.After(current) {
		return candidate
	}
	return current
}

func expired(now, seen time.Time, ttl time.Duration) bool {
	return !seen.IsZero() && ttl > 0 && now.Sub(seen) > ttl
}

func key(value session.Session) string {
	return value.Source + "\x00" + value.SessionID + "\x00" + strconv.Itoa(value.PID)
}

// Key exposes the identity shared by the pipe, the snapshot files and the
// store. The app layer needs it to compare what the scanner observed on disk
// with what the store still holds in memory.
func Key(value session.Session) string {
	return key(value)
}

// processBackedSources lists the reporters whose PID is a real, long-lived
// process, where probing liveness is worth doing so a dead session disappears
// at once instead of at the next TTL. Everything else relies on the snapshot
// TTL.
//
// This is deliberately an allow-list, and the default is the whole point.
// Probing a PID we cannot trust *deletes live sessions*: generic.js and every
// hook reporter key their session on a stable hash rather than a real process
// id, so a liveness check against that hash always fails and the session would
// vanish the instant it was reported - silently, because the removal is a
// best-effort unlink of a file nobody is watching.
//
// The cost of the other default is a session lingering for one TTL after its
// reporter died, which is free.
var processBackedSources = map[string]struct{}{
	"opencode": {},
}

func processBacked(source string) bool {
	_, known := processBackedSources[source]
	return known
}
