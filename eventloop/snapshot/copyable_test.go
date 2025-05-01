package snapshot

import (
	"testing"
)

type testCopyableState struct {
	value     int
	copyCalls int
}

func (s *testCopyableState) Copy() *testCopyableState {
	if s == nil {
		return nil
	}

	newT := CopyPtr(s)
	newT.copyCalls += 1

	return newT
}

func TestNewCopyableState(t *testing.T) {
	// Create a new copyable value.
	snap := NewCopyable(&testCopyableState{value: 69})

	if snap.Generation() != 0 {
		t.Errorf("expected generation 0, got %d", snap.Generation())
	}
	if snap.State().value != 69 {
		t.Errorf("expected value 69, got %d", snap.State().value)
	}
	if snap.State().copyCalls != 2 {
		t.Errorf("expected copyCalls 2, got %d", snap.State().copyCalls)
	}
}

func TestNextCopyableState(t *testing.T) {
	snap := NewCopyable(&testCopyableState{value: 69})

	// Create a new snapshot with the next state.
	snap2 := snap.Next(&testCopyableState{value: 70, copyCalls: snap.State().copyCalls})
	if snap2.Generation() != 1 {
		t.Errorf("expected generation 1, got %d", snap2.Generation())
	}
	if snap2.State().value != 70 {
		t.Errorf("expected value 70, got %d", snap2.State().value)
	}
	if snap2.State().copyCalls != 4 {
		t.Errorf("expected copyCalls 4, got %d", snap2.State().copyCalls)
	}
}

func TestExpireCopyableState(t *testing.T) {
	// Create a new copyable value.
	snap := NewCopyable(&testCopyableState{value: 69})

	select {
	case <-snap.Expiration():
		t.Error("expected snapshot to still be valid")
	default:
	}

	// Expire the snapshot.
	snap.Expire()

	// Check that the snapshot is expired.
	select {
	case _, ok := <-snap.Expiration():
		if ok {
			t.Error("expected expiration channel to be closed")
		}
	default:
		t.Error("expected snapshot to be expired")
	}
}
