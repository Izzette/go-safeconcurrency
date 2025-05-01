package snapshot_test

import (
	"encoding/json"
	"fmt"

	"github.com/Izzette/go-safeconcurrency/eventloop/snapshot"
)

// Demonstrates how to use the [snapshot.NewCopyable] function to create a snapshot with a state that can be deep
// copied.
func ExampleNewCopyable_deepCopy() {
	// Create a new snapshot with a deep copyable struct as the state.
	snap := snapshot.NewCopyable(&DeepCopyableState{})

	// Print the initial state of the snapshot.
	fmt.Printf("gen %d: %v\n", snap.Generation(), snap.State())

	// Update the state of the snapshot
	state := snap.State()
	state.Count++
	state.Key = "foo"
	state.Value = "bar"
	state.Nested = &NestedState{
		Map: map[string]string{"a": "b"},
		People: []*MarshalablePerson{
			{Name: "John", Age: 30},
			{Name: "Jane", Age: 25, Address: "123 Main St"},
		},
	}
	nextSnap := snap.Next(state)
	fmt.Printf("gen %d: %v\n", nextSnap.Generation(), nextSnap.State())

	// The previous snapshot is not modified.
	fmt.Printf("gen %d: %v\n", snap.Generation(), snap.State())
	//nolint:lll
	// Output:
	// gen 0: count: 0, key: "", value: "", nested: <nil>
	// gen 1: count: 1, key: "foo", value: "bar", nested: map: map[a:b], people: [name: "John", age: 30 years old, address: "" name: "Jane", age: 25 years old, address: "123 Main St"]
	// gen 0: count: 0, key: "", value: "", nested: <nil>
}

// DeepCopyableState is used as the state for the event loop, it has nested reference types.
type DeepCopyableState struct {
	// Count, Key, and Value are all value type, so we don't need to worry about copying them.
	Count int    `json:"count"`
	Key   string `json:"key"`
	Value string `json:"value"`

	// Nested is a struct that contains a map and a slice of pointers to person structs.
	Nested *NestedState `json:"nested"`
}

// Copy implements [types.Copyable.Copy].
func (s *DeepCopyableState) Copy() *DeepCopyableState {
	if s == nil {
		return nil
	}

	cpy := &DeepCopyableState{}
	if marshaled, err := json.Marshal(s); err != nil {
		panic(err)
	} else if err := json.Unmarshal(marshaled, &cpy); err != nil {
		panic(err)
	}

	return cpy
}

// String implements [fmt.Stringer.String].
func (s *DeepCopyableState) String() string {
	if s == nil {
		return "<nil>"
	}

	return fmt.Sprintf("count: %d, key: %#v, value: %#v, nested: %v", s.Count, s.Key, s.Value, s.Nested)
}

// NestedState is a struct that contains a map and a slice of pointers to person structs.
type NestedState struct {
	Map    map[string]string    `json:"map"`
	People []*MarshalablePerson `json:"people"`
}

// String implements [fmt.Stringer.String].
func (s *NestedState) String() string {
	if s == nil {
		return "<nil>"
	}

	return fmt.Sprintf("map: %v, people: %v", s.Map, s.People)
}

// MarshalablePerson is a structure that can be marshaled to and from JSON.
type MarshalablePerson struct {
	Name    string `json:"name"`
	Age     int    `json:"age"`
	Address string `json:"address"`
}

// String implements [fmt.Stringer.String].
func (p *MarshalablePerson) String() string {
	if p == nil {
		return "<nil>"
	}

	return fmt.Sprintf("name: %#v, age: %d years old, address: %#v", p.Name, p.Age, p.Address)
}
