package snapshot_test

import (
	"fmt"

	"github.com/Izzette/go-safeconcurrency/eventloop/snapshot"
)

func ExampleNewCopyable_bare() {
	// Create a new snapshot with a bare struct as the state.
	snap := snapshot.NewCopyable(BareState{})

	// Print the initial state of the snapshot.
	fmt.Printf("gen %d: %v\n", snap.Generation(), snap.State())

	// Update the state of the snapshot
	state := snap.State()
	state.Count++
	state.Key = "foo"
	state.Value = "bar"
	nextSnap := snap.Next(state)
	fmt.Printf("gen %d: %v\n", nextSnap.Generation(), nextSnap.State())

	// The previous snapshot is not modified.
	fmt.Printf("gen %d: %v\n", snap.Generation(), snap.State())
	// Output:
	// gen 0: count: 0, key: "", value: ""
	// gen 1: count: 1, key: "foo", value: "bar"
	// gen 0: count: 0, key: "", value: ""
}

// BareState is used as the state for the event loop, it has no mutable reference types.
type BareState struct {
	// Count, Key, and Value are all value type, so we don't need to worry about copying them.
	Count int
	Key   string
	Value string
}

// Copy implements [types.Copyable.Copy].
// As the receiver is not a pointer, and the struct has no reference types, we can just return the original struct.
func (s BareState) Copy() BareState {
	return s
}

// String implements [fmt.Stringer.String].
func (s BareState) String() string {
	return fmt.Sprintf("count: %d, key: %#v, value: %#v", s.Count, s.Key, s.Value)
}
