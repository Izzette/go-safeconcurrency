package snapshot

import "testing"

func TestNewSlice(t *testing.T) {
	// Create a new slice snapshot.
	snap := NewSlice([]int{1, 2, 3})

	// Check that the state is not nil.
	if snap.State() == nil {
		t.Fatal("expected non-nil state")
	}
	if snap.Generation() != 0 {
		t.Errorf("expected generation 0, got %d", snap.Generation())
	}
	if snap.State()[0] != 1 {
		t.Errorf("expected value 1, got %d", snap.State()[0])
	}
	if snap.State()[1] != 2 {
		t.Errorf("expected value 2, got %d", snap.State()[1])
	}
	if snap.State()[2] != 3 {
		t.Errorf("expected value 3, got %d", snap.State()[2])
	}

	select {
	case <-snap.Expiration():
		t.Error("expected snapshot to still be valid")
	default:
	}
}

func TestNextSlice(t *testing.T) {
	// Create a new slice snapshot.
	snap := NewSlice([]int{1, 2, 3})
	if snap.State() == nil {
		t.Fatal("expected non-nil state")
	}

	// Create a new snapshot with the next state.
	state := snap.State()
	state[0] = 3
	state = state[:2]
	snap2 := snap.Next(state)
	if snap2.State() == nil {
		t.Fatal("expected non-nil state")
	}
	if snap2.Generation() != 1 {
		t.Errorf("expected generation 1, got %d", snap2.Generation())
	}
	if snap2.State()[0] != 3 {
		t.Errorf("expected value 3, got %d", snap2.State()[0])
	}
	if snap2.State()[1] != 2 {
		t.Errorf("expected value 2, got %d", snap2.State()[1])
	}
	if len(snap2.State()) != 2 {
		t.Errorf("expected length to be 2, got %d", len(snap2.State()))
	}

	if snap.Generation() != 0 {
		t.Errorf("expected generation 0, got %d", snap.Generation())
	}
}

func TestExpireSlice(t *testing.T) {
	// Create a new slice snapshot.
	snap := NewSlice([]int{1, 2, 3})
	if snap.State() == nil {
		t.Fatal("expected non-nil state")
	}

	// Expire the snapshot.
	snap.Expire()

	// Check that the snapshot is expired.
	select {
	case <-snap.Expiration():
	default:
		t.Error("expected snapshot to be expired")
	}
}

func TestEmptySlice(t *testing.T) {
	// Create a new slice snapshot.
	snap := NewEmptySlice[int]()

	// Check that the state is not nil.
	if snap.State() == nil {
		t.Fatal("expected non-nil state")
	}
	if snap.Generation() != 0 {
		t.Errorf("expected generation 0, got %d", snap.Generation())
	}
	if len(snap.State()) != 0 {
		t.Errorf("expected empty slice, got %v", snap.State())
	}
}
