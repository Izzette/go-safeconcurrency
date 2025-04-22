package results

import (
	"context"

	"github.com/Izzette/go-safeconcurrency/api/types"
)

// NewHandle creates a new SimpleHandle, which implements [types.Handle] and is used to publish results from a
// [types.Runner] or [types.Task].
func NewHandle[T any](results chan<- T) types.Handle[T] {
	return &handle[T]{results: results}
}

// handle implements [types.Handle].
type handle[T any] struct {
	results chan<- T
}

// Publish implements [types.Handle.Publish].
// It publishes a result to the results channel.
// If the [context.Context] is canceled, it returns an error.
func (h *handle[T]) Publish(ctx context.Context, value T) error {
	// The `select` statement is non-deterministic, and may still publish a result even if the context has been canceled
	// before Publish is called.
	if err := context.Cause(ctx); err != nil {
		//nolint:wrapcheck
		return err
	}

	select {
	case <-ctx.Done():
		//nolint:wrapcheck
		return context.Cause(ctx)
	case h.results <- value:
		return nil
	}
}

// Close implements [types.Handle.Close].
// It closes the underlying results channel.
func (h *handle[T]) Close() {
	// Close the results channel.
	close(h.results)
}
