package autostart

import (
	"errors"
	"testing"
)

// memoryBackend stands in for the registry so these tests never touch real
// machine state.
type memoryBackend struct {
	values  map[string]string
	failOn  string
	readErr error
}

func newMemory() *memoryBackend {
	return &memoryBackend{values: map[string]string{}}
}

func (m *memoryBackend) read(name string) (string, bool, error) {
	if m.readErr != nil {
		return "", false, m.readErr
	}
	value, found := m.values[name]
	return value, found, nil
}

func (m *memoryBackend) write(name, value string) error {
	if m.failOn == "write" {
		return errors.New("write refused")
	}
	m.values[name] = value
	return nil
}

func (m *memoryBackend) remove(name string) error {
	if m.failOn == "remove" {
		return errors.New("remove refused")
	}
	delete(m.values, name)
	return nil
}

func TestCommandQuotesThePath(t *testing.T) {
	// Windows parses the Run value as a command line, so a path with spaces
	// must be quoted or the shell would try to run "C:\Program".
	got := Command(`C:\Users\a b\AppData\Local\AI-Terminal-Pet\AI-Terminal-Pet.exe`)
	want := `"C:\Users\a b\AppData\Local\AI-Terminal-Pet\AI-Terminal-Pet.exe"`
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	// Quoting is unconditional so a space-free path today cannot break after a
	// move to one with spaces.
	if Command(`C:\pet.exe`) != `"C:\pet.exe"` {
		t.Fatal("quoting must not depend on whitespace")
	}
}

func TestStatusReportsTarget(t *testing.T) {
	backend := newMemory()
	manager := newWith(backend)

	status, err := manager.Status()
	if err != nil || status.Enabled {
		t.Fatalf("fresh backend must be disabled, got %+v err=%v", status, err)
	}

	if err := manager.Enable(`C:\a\pet.exe`); err != nil {
		t.Fatalf("enable: %v", err)
	}
	status, err = manager.Status()
	if err != nil || !status.Enabled {
		t.Fatalf("must be enabled, got %+v err=%v", status, err)
	}
	if status.Target != `C:\a\pet.exe` {
		t.Fatalf("target must be unquoted, got %q", status.Target)
	}
}

func TestStatusToleratesHandEditedValue(t *testing.T) {
	// Someone may have written the value by hand without quotes; Status must
	// still report the path rather than a quoted blob.
	backend := newMemory()
	backend.values[ValueName] = `C:\a\pet.exe`
	status, err := newWith(backend).Status()
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if status.Target != `C:\a\pet.exe` {
		t.Fatalf("unquoted value must be tolerated, got %q", status.Target)
	}
}

func TestDisableIsIdempotent(t *testing.T) {
	backend := newMemory()
	manager := newWith(backend)
	// "make sure it is off" must succeed when it is already off.
	if err := manager.Disable(); err != nil {
		t.Fatalf("disable on empty must succeed: %v", err)
	}
	if err := manager.Enable(`C:\a\pet.exe`); err != nil {
		t.Fatalf("enable: %v", err)
	}
	if err := manager.Disable(); err != nil {
		t.Fatalf("disable: %v", err)
	}
	if status, _ := manager.Status(); status.Enabled {
		t.Fatal("must be disabled after Disable")
	}
}

func TestEnableRejectsEmptyPath(t *testing.T) {
	if err := newWith(newMemory()).Enable("   "); err == nil {
		t.Fatal("empty executable must be rejected")
	}
}

func TestSyncRepairsMovedExecutable(t *testing.T) {
	backend := newMemory()
	manager := newWith(backend)
	if err := manager.Enable(`C:\old\pet.exe`); err != nil {
		t.Fatalf("enable: %v", err)
	}
	changed, err := manager.Sync(`C:\new\pet.exe`)
	if err != nil {
		t.Fatalf("sync: %v", err)
	}
	if !changed {
		t.Fatal("a stale target must be repaired")
	}
	status, _ := manager.Status()
	if status.Target != `C:\new\pet.exe` {
		t.Fatalf("target must be repaired, got %q", status.Target)
	}
	// Second sync is a no-op.
	if changed, err = manager.Sync(`C:\new\pet.exe`); err != nil || changed {
		t.Fatalf("second sync must be a no-op, changed=%v err=%v", changed, err)
	}
}

func TestSyncNeverEnablesADisabledEntry(t *testing.T) {
	// Turning autostart on behind the user's back would be a nasty surprise.
	backend := newMemory()
	manager := newWith(backend)
	changed, err := manager.Sync(`C:\new\pet.exe`)
	if err != nil {
		t.Fatalf("sync: %v", err)
	}
	if changed {
		t.Fatal("sync must not change anything when disabled")
	}
	if status, _ := manager.Status(); status.Enabled {
		t.Fatal("sync must not enable autostart")
	}
}

func TestSyncSurfacesReadErrors(t *testing.T) {
	backend := newMemory()
	backend.readErr = errors.New("registry unavailable")
	if _, err := newWith(backend).Sync(`C:\new\pet.exe`); err == nil {
		t.Fatal("a read error must be surfaced, not swallowed")
	}
}
