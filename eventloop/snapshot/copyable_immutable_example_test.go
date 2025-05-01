package snapshot_test

import (
	"fmt"

	"github.com/Izzette/go-safeconcurrency/eventloop/snapshot"
	"github.com/benbjohnson/immutable"
)

// Demonstrates how to use the Copyable interface to wrap immutable data structures.
func ExampleNewCopyable_immutable() {
	// Create a new snapshot with an immutableState
	state := newImmutableState()
	state.Map = state.Set("foo", "bar")
	snap := snapshot.NewCopyable(state)

	fmt.Printf("gen %d: %v\n", snap.Generation(), snap.State())

	// Set the value of foo in the state.
	state = snap.State()
	state.Map = state.Set("foo", "quz")
	nextSnap := snap.Next(state)
	fmt.Printf("gen %d: %v\n", nextSnap.Generation(), nextSnap.State())

	// The previous snapshot is not modified.
	fmt.Printf("gen %d: %v\n", snap.Generation(), snap.State())
	// Output:
	// gen 0: map[foo:bar]
	// gen 1: map[foo:quz]
	// gen 0: map[foo:bar]
}

// immutableState is used as the state for the event loop, it wraps an immutable data structure.
type immutableState struct {
	*immutable.Map[string, string]
}

// newImmutableState creates a new immutable state.
func newImmutableState() *immutableState {
	return &immutableState{
		immutable.NewMap[string, string](nil),
	}
}

// Copy implements [types.Copyable.Copy].
// As the other map is immutable, we can just return a copy of the wrapper struct.
func (s *immutableState) Copy() *immutableState {
	if s == nil {
		return nil
	}

	return snapshot.CopyPtr(s)
}

// String implements [fmt.Stringer.String].
func (s *immutableState) String() string {
	m := make(map[string]string, s.Len())
	it := s.Iterator()
	for !it.Done() {
		k, v, _ := it.Next()
		m[k] = v
	}

	return fmt.Sprintf("%v", m)
}
