package snapshot

import "testing"

func TestCopyMap(t *testing.T) {
	// Create a map with some values.
	m := map[string]int{
		"a": 1,
		"b": 2,
		"c": 3,
	}

	// Copy the map.
	cpy := CopyMap(m)

	// Check that the copied map has the same values.
	if cpy["a"] != 1 {
		t.Errorf("expected value 1, got %d", cpy["a"])
	}
	if cpy["b"] != 2 {
		t.Errorf("expected value 2, got %d", cpy["b"])
	}
	if cpy["c"] != 3 {
		t.Errorf("expected value 3, got %d", cpy["c"])
	}

	// Check that the copied map is not the same as the original map.
	cpy["a"] = 4
	if m["a"] != 1 {
		t.Error("expected copied map to be different from original map")
	}
}

func TestCopySlice(t *testing.T) {
	// Create a slice with some values.
	s := []int{1, 2, 3}

	// Copy the slice.
	cpy := CopySlice(s)

	// Check that the copied slice has the same values.
	if cpy[0] != 1 {
		t.Errorf("expected value 1, got %d", cpy[0])
	}
	if cpy[1] != 2 {
		t.Errorf("expected value 2, got %d", cpy[1])
	}
	if cpy[2] != 3 {
		t.Errorf("expected value 3, got %d", cpy[2])
	}

	// Check that the copied slice is not the same as the original slice.
	cpy[0] = 4
	if s[0] != 1 {
		t.Error("expected copied slice to be different from original slice")
	}
}

func TestCopyPtr(t *testing.T) {
	// Create a struct with some values.
	type testStruct struct {
		a int
		b int
	}

	// Create a pointer to the struct.
	s := &testStruct{a: 1, b: 2}

	// Copy the pointer.
	cpy := CopyPtr(s)

	// Check that the copied pointer has the same values.
	if cpy.a != 1 {
		t.Errorf("expected value 1, got %d", cpy.a)
	}
	if cpy.b != 2 {
		t.Errorf("expected value 2, got %d", cpy.b)
	}

	// Check that the copied pointer is not the same as the original pointer.
	cpy.a = 3
	if s.a != 1 {
		t.Error("expected copied pointer to be different from original pointer")
	}
}

func TestCopyNilPtr(t *testing.T) {
	// Create a nil pointer.
	var s *int

	// Copy the nil pointer.
	cpy := CopyPtr(s)

	// Check that the copied pointer is nil.
	if cpy != nil {
		t.Error("expected copied pointer to be nil")
	}
}
