package snapshot

import "testing"

func TestNewMap(t *testing.T) {
	// Create a new map snapshot.
	snap := NewMap[string, int](map[string]int{"a": 1, "b": 2})

	// Check that the state is not nil.
	if snap.State() == nil {
		t.Fatal("expected non-nil state")
	}
	if snap.Generation() != 0 {
		t.Errorf("expected generation 0, got %d", snap.Generation())
	}
	if snap.State()["a"] != 1 {
		t.Errorf("expected value 1, got %d", snap.State()["a"])
	}
	if snap.State()["b"] != 2 {
		t.Errorf("expected value 2, got %d", snap.State()["b"])
	}

	select {
	case <-snap.Expiration():
		t.Error("expected snapshot to still be valid")
	default:
	}
}

func TestNextMap(t *testing.T) {
	// Create a new map snapshot.
	snap := NewMap[string, int](map[string]int{"a": 1, "b": 2, "c": 3})

	// Create a new snapshot with the next state.
	state := snap.State()
	state["a"] = 3
	state["b"] = 4
	delete(state, "c")
	snap2 := snap.Next(state)
	if snap2.State() == nil {
		t.Fatal("expected non-nil state")
	}
	if snap2.Generation() != 1 {
		t.Errorf("expected generation 1, got %d", snap2.Generation())
	}
	if snap2.State()["a"] != 3 {
		t.Errorf("expected value 3, got %d", snap2.State()["a"])
	}
	if snap2.State()["b"] != 4 {
		t.Errorf("expected value 4, got %d", snap2.State()["b"])
	}
	if v, ok := snap2.State()["c"]; ok {
		t.Errorf("expected value c to be deleted, got %d", v)
	}

	if snap.Generation() != 0 {
		t.Errorf("expected generation 0, got %d", snap.Generation())
	}
	if snap.State()["a"] != 1 {
		t.Errorf("expected value 1, got %d", snap.State()["a"])
	}
	if snap.State()["b"] != 2 {
		t.Errorf("expected value 2, got %d", snap.State()["b"])
	}
	if snap.State()["c"] != 3 {
		t.Errorf("expected value 3, got %d", snap.State()["c"])
	}
}

func TestEmptyMap(t *testing.T) {
	// Create a new map snapshot.
	snap := NewEmptyMap[string, int]()

	// Check that the state is not nil.
	if snap.State() == nil {
		t.Fatal("expected non-nil state")
	}
	if snap.Generation() != 0 {
		t.Errorf("expected generation 0, got %d", snap.Generation())
	}
	if len(snap.State()) != 0 {
		t.Errorf("expected empty map, got %d", len(snap.State()))
	}
}
