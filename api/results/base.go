package results

import (
	"context"

	"github.com/Izzette/go-safeconcurrency/api/types"
)

type simpleResult[T any] struct {
	value T
}

func (r *simpleResult[T]) Get() (T, error) {
	return r.value, nil
}

type simpleError[T any] struct {
	err error
}

func (r *simpleError[T]) Get() (T, error) {
	// Get the zero-value of type T
	var zero T

	return zero, r.err
}

// handle implements types.handle, which is used to publish results from a
// types.Runner or types.Task.
type handle[T any] struct {
	results chan<- types.Result[T]
}

// NewHandle creates a new SimpleHandle, which implements types.Handle.
func NewHandle[T any](results chan<- types.Result[T]) types.Handle[T] {
	return &handle[T]{results: results}
}

func (h *handle[T]) Publish(ctx context.Context, value T) error {
	return h.send(ctx, &simpleResult[T]{value})
}

func (h *handle[T]) Error(ctx context.Context, err error) error {
	return h.send(ctx, &simpleError[T]{err})
}

func (h *handle[T]) Close() {
	// Close the results channel.
	close(h.results)
}

// send is a helper function to send a result to the results channel respecting the context.
func (h *handle[T]) send(ctx context.Context, result types.Result[T]) error {
	// The `select` statement is non-deterministic, and may still publish a result even if the context has been canceled
	// before Publish is called.
	if err := ctx.Err(); err != nil {
		//nolint:wrapcheck
		return err
	}

	select {
	case <-ctx.Done():
		//nolint:wrapcheck
		return ctx.Err()
	case h.results <- result:
		return nil
	}
}
