package snapshot

import (
	"sync"

	"github.com/Izzette/go-safeconcurrency/api/types"
)

// abstract is an abstract type implementation part of [types.StateSnapshot].
type abstract struct {
	gen        types.GenerationID
	expiration chan struct{}
	expireOnce *sync.Once
}

func newAbstract() *abstract {
	return &abstract{
		expiration: make(chan struct{}),
		expireOnce: &sync.Once{},
	}
}

// Next creates a copy of the snapshot.
// The generation ID is incremented.
// A new expiration channel is created, but the old one is not closed.
func (s *abstract) Next() *abstract {
	// Make a shallow copy of the snapshot.
	cpy := CopyPtr(s)
	// Make a new expiration channel.
	cpy.expiration = make(chan struct{})
	// Make a new expireOnce.
	cpy.expireOnce = &sync.Once{}
	// Increment the generation ID.
	cpy.gen = s.gen + 1

	// Return the copy.
	return cpy
}

// Generation implements [types.StateSnapshot.Generation].
func (s *abstract) Generation() types.GenerationID {
	return s.gen
}

// Expiration implements [types.StateSnapshot.Expiration].
func (s *abstract) Expiration() <-chan struct{} {
	return s.expiration
}

// Expire implements [types.StateSnapshot.Expire].
func (s *abstract) Expire() {
	s.expireOnce.Do(s.expire)
}

// expire closes the expiration channel.
func (s *abstract) expire() {
	close(s.expiration)
}
