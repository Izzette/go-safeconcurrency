package snapshot

import "github.com/Izzette/go-safeconcurrency/api/types"

// value implements [types.StateSnapshot] a non-reference type.
type value[StateT types.Value] struct {
	*abstract
	state StateT
}

// NewValue returns a new [types.StateSnapshot] with the given initial state.
func NewValue[StateT types.Value](initialState StateT) types.StateSnapshot[StateT] {
	return &value[StateT]{
		abstract: newAbstract(),
		state:    initialState,
	}
}

// NewZeroValue returns a new [types.StateSnapshot] with the zero value of the type.
// It is equivalent to calling [NewValue] with the zero value of the type.
func NewZeroValue[StateT types.Value]() types.StateSnapshot[StateT] {
	var zero StateT

	return NewValue(zero)
}

// State implements [types.StateSnapshot.State].
func (s *value[StateT]) State() StateT {
	return s.state
}

// Next implements [types.StateSnapshot.Next].
func (s *value[StateT]) Next(state StateT) types.StateSnapshot[StateT] {
	cpy := CopyPtr(s)
	cpy.abstract = s.abstract.Next()
	cpy.state = state

	return cpy
}
