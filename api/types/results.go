package types

import (
	"context"
)

// Emitter is passed to a [types.Producer.Run] and [types.StreamingTask.Execute] methods to propagate results.
type Emitter[T any] interface {
	// Emit publishes an result from the current goroutine to the consumer.
	// If an error is returned, it represents context cancelation.
	// In these cases, the Producer should immediately cleanup, and return ASAP.
	Emit(context.Context, T) error

	// Close closes the emitter and releases any resources.
	Close()
}
