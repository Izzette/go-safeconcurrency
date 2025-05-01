package snapshot

import "github.com/Izzette/go-safeconcurrency/api/types"

// NewMap returns a new [types.StateSnapshot] with the given initial state.
// Keep in mind that while the map is copied, neither the keys nor values inside the slice are copied.
// If the values are references, they will be shared between the snapshots!
func NewMap[KeyT comparable, ValueT any](initialState map[KeyT]ValueT) types.StateSnapshot[map[KeyT]ValueT] {
	initialState = CopyMap(initialState)

	return &mapSnapshot[KeyT, ValueT]{
		abstract: newAbstract(),
		state:    initialState,
	}
}

// NewEmptyMap returns a new [types.StateSnapshot] with an empty map.
// It is equivalent to calling [NewMap] with an empty map.
func NewEmptyMap[KeyT comparable, ValueT any]() types.StateSnapshot[map[KeyT]ValueT] {
	return NewMap[KeyT, ValueT](make(map[KeyT]ValueT))
}

// mapSnapshot implements [types.StateSnapshot] for a map.
type mapSnapshot[KeyT comparable, ValueT any] struct {
	*abstract
	state map[KeyT]ValueT
}

// State implements [types.StateSnapshot.State].
func (s *mapSnapshot[KeyT, ValueT]) State() map[KeyT]ValueT {
	return CopyMap(s.state)
}

// Next implements [types.StateSnapshot.Next].
func (s *mapSnapshot[KeyT, ValueT]) Next(state map[KeyT]ValueT) types.StateSnapshot[map[KeyT]ValueT] {
	return &mapSnapshot[KeyT, ValueT]{
		abstract: s.abstract.Next(),
		state:    CopyMap(state),
	}
}
