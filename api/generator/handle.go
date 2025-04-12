package generator

import (
	"context"
	"sync/atomic"

	"github.com/Izzette/go-safeconcurrency/api/results"
	"github.com/Izzette/go-safeconcurrency/api/types"
)

type handle[T any] struct {
	results  chan<- types.Result[T]
	canceled atomic.Bool
}

func (h *handle[T]) Publish(ctx context.Context, value T) error {
	return h.result(ctx, results.NewSimpleResult(value))
}

func (h *handle[T]) Error(ctx context.Context, err error) error {
	return h.result(ctx, results.NewSimpleError[T](err))
}

func (h *handle[T]) result(ctx context.Context, result types.Result[T]) error {
	if h.canceled.Load() {
		panic("attempt to publish to a canceled worker")
	}

	// The `select` statement is non-deterministic, and may still publish a result
	// even if the context has been canceled before Publish is called.
	if err := ctx.Err(); err != nil {
		h.canceled.Store(true)

		//nolint:wrapcheck
		return err
	}

	select {
	case <-ctx.Done():
		h.canceled.Store(true)

		//nolint:wrapcheck
		return ctx.Err()
	case h.results <- result:
		return nil
	}
}
