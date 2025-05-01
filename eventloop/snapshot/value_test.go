package snapshot

import "testing"

func TestNewValue(t *testing.T) {
	// Create a new value snapshot.
	snap := NewValue[string]("test")

	if snap.State() != "test" {
		t.Fatalf("expected state to be \"test\", got %#v\n", snap.State())
	}
	if snap.Generation() != 0 {
		t.Errorf("expected generation 0, got %d", snap.Generation())
	}

	select {
	case <-snap.Expiration():
		t.Error("expected snapshot to still be valid")
	default:
	}
}

func TestNextValue(t *testing.T) {
	// Create a new value snapshot.
	snap := NewValue("test")
	if snap.State() != "test" {
		t.Errorf("expected state to be \"test\", got %#v\n", snap.State())
	}

	// Create a new snapshot with the next state.
	snap2 := snap.Next("new value")
	if snap2.State() != "new value" {
		t.Fatalf("expected state to be \"new value\", got %#v\n", snap2.State())
	}
	if snap2.Generation() != 1 {
		t.Errorf("expected generation 1, got %d", snap2.Generation())
	}

	if snap.Generation() != 0 {
		t.Errorf("expected original snapshot to be generation 0, got %d", snap.Generation())
	}
}

func TestZeroValue(t *testing.T) {
	// Create a new value snapshot.
	snap := NewZeroValue[string]()

	if snap.State() != "" {
		t.Fatalf("expected state to be \"\", got %#v\n", snap.State())
	}
	if snap.Generation() != 0 {
		t.Errorf("expected generation 0, got %d", snap.Generation())
	}
}
