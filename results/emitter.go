package results

import (
	"context"
	"sync"

	"github.com/Izzette/go-safeconcurrency/api/types"
)

// NewEmitter creates a new [types.Emitter] and is used to emit results from a [types.Producer] or
// [types.StreamingTask].
func NewEmitter[T any](results chan<- T) types.Emitter[T] {
	return &emitter[T]{results: results, closeOnce: &sync.Once{}}
}

// emitter implements [types.Emitter].
type emitter[T any] struct {
	results   chan<- T
	closeOnce *sync.Once
}

// Emit implements [types.Emitter.Emit].
// It emits a result to the results channel.
// If the [context.Context] is canceled, it returns an error.
func (e *emitter[T]) Emit(ctx context.Context, value T) error {
	// The `select` statement is non-deterministic, and may still emit a result even if the context has been canceled
	// before Emit is called.
	if err := context.Cause(ctx); err != nil {
		//nolint:wrapcheck
		return err
	}

	select {
	case <-ctx.Done():
		//nolint:wrapcheck
		return context.Cause(ctx)
	case e.results <- value:
		return nil
	}
}

// Close implements [types.Emitter.Close].
// It closes the underlying results channel.
func (e *emitter[T]) Close() {
	e.closeOnce.Do(e.closeResults)
}

// closeResults closes the results channel without synchronizing with [emitter.closeOnce].
func (e *emitter[T]) closeResults() {
	close(e.results)
}
