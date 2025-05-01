package snapshot

import "github.com/Izzette/go-safeconcurrency/api/types"

// NewSlice returns a new [types.StateSnapshot] with the given initial state.
// Keep in mind that while the slice is copied, the elements inside the slice are not.
// If the slice elements are references, they will be shared between the snapshots!
func NewSlice[T any](initialState []T) types.StateSnapshot[[]T] {
	return &sliceSnapshot[T]{
		abstract: newAbstract(),
		state:    CopySlice[T](initialState),
	}
}

// NewEmptySlice returns a new [types.StateSnapshot] with an empty slice.
// It is equivalent to calling [NewSlice] with an empty slice.
func NewEmptySlice[T any]() types.StateSnapshot[[]T] {
	return NewSlice[T](make([]T, 0))
}

// sliceSnapshot implements [types.StateSnapshot] for a slice.
type sliceSnapshot[T any] struct {
	*abstract
	state []T
}

// State implements [types.StateSnapshot.State].
func (s *sliceSnapshot[T]) State() []T {
	return CopySlice(s.state)
}

// Next implements [types.StateSnapshot.Next].
func (s *sliceSnapshot[T]) Next(state []T) types.StateSnapshot[[]T] {
	cpy := CopyPtr(s)
	cpy.abstract = s.abstract.Next()
	cpy.state = CopySlice(state)

	return cpy
}
