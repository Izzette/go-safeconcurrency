package examples

import (
	"fmt"

	"github.com/Izzette/go-safeconcurrency/api/types"
)

// AppState holds our application state
type AppState struct {
	Counter int
}

// IncrementEvent implements types.Event to modify our state
type IncrementEvent struct{}

// Dispatch implements [types.Event.Dispatch].
func (e *IncrementEvent) Dispatch(gen types.GenerationID, s *AppState) {
	fmt.Printf("<IncrementEvent@%d> current value: %d, next value: %d\n", gen, s.Counter, s.Counter+1)
	s.Counter++
}
