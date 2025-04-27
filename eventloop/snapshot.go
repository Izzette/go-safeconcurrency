package eventloop

import "github.com/Izzette/go-safeconcurrency/api/types"

// snapshot is an implementation of [types.StateSnapshot].
type snapshot[StateT any] struct {
	gen        types.GenerationID
	state      *StateT
	expiration chan struct{}
}

// Copy creates a copy of the snapshot.
// A shallow copy of the state is made to allow modifications.
// A new expiration channel is created, but the old one is not closed.
func (s *snapshot[StateT]) Copy() *snapshot[StateT] {
	// Make a shallow copy of the snapshot.
	cpy := copyPtr(s)
	// Make a shallow copy of the state.
	cpy.state = copyPtr(s.state)
	// Make a new expiration channel.
	cpy.expiration = make(chan struct{})

	// Return the copy.
	return cpy
}

// State implements [types.StateSnapshot.State].
func (s *snapshot[StateT]) State() *StateT {
	return s.state
}

// Generation implements [types.StateSnapshot.Generation].
func (s *snapshot[StateT]) Generation() types.GenerationID {
	return s.gen
}

// Expiration implements [types.StateSnapshot.Expiration].
func (s *snapshot[StateT]) Expiration() <-chan struct{} {
	return s.expiration
}

// expire closes the expiration channel.
func (s *snapshot[StateT]) expire() {
	close(s.expiration)
}
