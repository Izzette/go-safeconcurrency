package snapshot_test

import (
	"fmt"

	"github.com/Izzette/go-safeconcurrency/eventloop/snapshot"
)

// Demonstrates how to use a Copy-on-Write strategy with a snapshot state.
// This is particularly useful for large, complex data structures that are expensive to do an eager deep-copy.
func ExampleNewCopyable_coW() {
	// Create a new Copyable snapshot with a cowState
	snapshot := snapshot.NewCopyable(&CowState{
		Count: 0,
		People: map[string]*Person{
			"Amy": {
				Name:    "Amy Smith",
				Age:     25,
				Address: "456 Elm St",
			},
		},
	})

	// Print the current state
	fmt.Printf("gen %d: Count %d\n", snapshot.Generation(), snapshot.State().Count)
	fmt.Printf("gen %d: Amy is %v\n", snapshot.Generation(), snapshot.State().GetPerson("Amy"))
	fmt.Printf("gen %d: John is %v\n", snapshot.Generation(), snapshot.State().GetPerson("John"))

	// .State() is calling .Copy() on the cowState struct, so we can modify it without affecting the original.
	nextState := snapshot.State()
	nextState.Count++
	nextState.SetPerson("John", &Person{
		Name:    "John Brown",
		Age:     30,
		Address: "123 Main St",
	})
	nextSnapshot := snapshot.Next(nextState)

	// Print the next state
	fmt.Printf("gen %d: Count %d\n", nextSnapshot.Generation(), nextSnapshot.State().Count)
	fmt.Printf("gen %d: John is %v\n", nextSnapshot.Generation(), nextSnapshot.State().GetPerson("John"))

	// The original snapshot has not changed
	fmt.Printf("gen %d: Count %d\n", snapshot.Generation(), snapshot.State().Count)
	fmt.Printf("gen %d: John is %v\n", snapshot.Generation(), snapshot.State().GetPerson("John"))

	// Normally we don't want to access the underlying people map directly, but we can do it here to demonstrate that the
	// original snapshot shares the same pointer for Amy, a person who was not modified.
	if snapshot.State().People["Amy"] == nextSnapshot.State().People["Amy"] {
		fmt.Printf("Amy is shared with CoW between gen %d and %d\n", snapshot.Generation(), nextSnapshot.Generation())
	}
	// Output:
	// gen 0: Count 0
	// gen 0: Amy is name: Amy Smith, age: 25 years old, address: 456 Elm St
	// gen 0: John is <nil>
	// gen 1: Count 1
	// gen 1: John is name: John Brown, age: 30 years old, address: 123 Main St
	// gen 0: Count 0
	// gen 0: John is <nil>
	// Amy is shared with CoW between gen 0 and 1
}

// Person is a struct that implements [types.Copyable].
type Person struct {
	Name    string
	Age     int
	Address string
}

// Copy implements [types.Copyable.Copy].
func (p *Person) Copy() *Person {
	if p == nil {
		return nil
	}

	return &Person{
		Name:    p.Name,
		Age:     p.Age,
		Address: p.Address,
	}
}

// String implements [fmt.Stringer.String].
func (p *Person) String() string {
	if p == nil {
		return "<nil>"
	}

	return fmt.Sprintf("name: %+v, age: %d years old, address: %+v", p.Name, p.Age, p.Address)
}

// CowState is used as the state for the event loop, it has reference types but exposes copy-on-write methods.
type CowState struct {
	// Count is a value type, so we don't need to worry about copying it.
	Count int

	// people is a map of string to *Person.
	// This is a reference type, so we need to use a copy-on-write strategy to avoid modifying the original.
	// This map is copied on write, so that each event can modify it without affecting the original.
	People map[string]*Person
}

// GetPerson returns the person at the given key in the map.
// This prevents exposing the original map, preventing unintended modifications.
func (s *CowState) GetPerson(key string) *Person {
	return s.People[key].Copy()
}

// SetPerson sets the person at the given key in the map.
// Allows modifying the map without updating the original present in snapshots.
func (s *CowState) SetPerson(key string, value *Person) {
	// Copy the map to avoid modifying the original
	// This will not copy the values themselves, it will just copy the map.
	// All the pointers to Person will still point to the same object.
	people := snapshot.CopyMap(s.People)

	// Update the map, copying the value to ensure that no modifications to the pointer made outside the *State
	// struct affect the stored value.
	people[key] = value.Copy()

	// Replace the map with the updated one
	s.People = people
}

// A Copy-on-Write strategy is used for updates to avoid modifying the original state, so it's safe to use a shallow.
func (s *CowState) Copy() *CowState {
	if s == nil {
		return nil
	}

	return &CowState{
		Count:  s.Count,  // As Count is a value type, we can just copy it.
		People: s.People, // As the map is copied on write, we can use the original.
	}
}
