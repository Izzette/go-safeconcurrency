package snapshot

import (
	"github.com/Izzette/go-safeconcurrency/api/types"
)

// NewCopyable returns a new [types.StateSnapshot] with the given initial state.
// When [types.StateSnapshot.State] and [types.StateSnapshot.Next] is called, the [types.Copyable] interface is used to
// create a copy of the state.
// You're state should be a persistent data structure, modifications to the state returned by Copy will not affect the
// original.
// There are several ways to implement this:
//
//   - Use a "bare" structure that is neither a pointer nor contains any reference types.
//     In this case the [types.Copyable.Copy] method can just return the original structure.
//   - Use a copy-on-write strategy, where only the relevant state is copied when it is modified.
//     In this case you typically return a new root state, which is a shallow copy of the top-level structure.
//     This is the most complicated solution, but it is very efficient when handing large, complex data structures that
//     need small, scattered modifications.
//   - Perform a deep copy of the state.
//     This is a simple and extremely versatile solution, but it can become a performance bottleneck if the state is
//     large.
//     Libraries like [github.com/barkimedes/go-deepcopy] use reflection to create a deep copies of even the most
//     complex structures.
//     Tools like [github.com/globusdigital/deep-copy] generate code to allow deep-copying of types.
//     Or trivally, you can just [encoding/json.Marshal] the state to JSON and [encoding/json.Unmarshal] it back to a
//     new instance, which is returned from your [types.Copyable.Copy] method.
//   - Use an immutable data structure, where the state is never modified, and any modifications create a new data
//     structure.
//     [github.com/benbjohnson/immutable] works great for this, however the types themselves do not implement
//     [types.Copyable].
//     You can wrap these immutable data structures in a struct that implements [types.Copyable].
//     As the underlying structure is immutable, you can just return a copy of the wrapper struct.
func NewCopyable[StateT types.Copyable[StateT]](initialState StateT) types.StateSnapshot[StateT] {
	return &copyableSnapshot[StateT]{
		abstract: newAbstract(),
		state:    initialState.Copy(),
	}
}

// copyableSnapshot implements [types.StateSnapshot] for a pointer to a type that implements [types.Copyable].
type copyableSnapshot[StateT types.Copyable[StateT]] struct {
	*abstract
	state StateT
}

// State implements [types.StateSnapshot.State].
func (s *copyableSnapshot[StateT]) State() StateT {
	return s.state.Copy()
}

// Copy implements [types.StateSnapshot.Copy].
func (s *copyableSnapshot[StateT]) Next(state StateT) types.StateSnapshot[StateT] {
	cpy := CopyPtr(s)
	cpy.abstract = s.abstract.Next()
	cpy.state = state.Copy()

	return cpy
}
