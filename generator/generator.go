package generator

import (
	"context"
	"sync/atomic"

	"github.com/Izzette/go-safeconcurrency/api/types"
	"github.com/Izzette/go-safeconcurrency/results"
)

// NewGenerator creates (but does not start) a basic implementation of [types.Generator] with no results buffering.
// If you would like results buffering, use [NewGeneratorBuffered] instead.
// This is equivalent to calling [NewGeneratorBuffered] with a buffer size of 0.
func NewGenerator[T any](runner types.Runner[T]) types.Generator[T] {
	return NewGeneratorBuffered(runner, 0)
}

// NewGeneratorBuffered creates (but does not start) a basic implementation of [types.Generator] with the specified
// results buffer size.
// It is not re-startable, and thus [types.Generator.Start] or [types.Generator.Run] must only be called exactly once.
func NewGeneratorBuffered[T any](runner types.Runner[T], buffer uint) types.Generator[T] {
	return &generator[T]{
		runner:  runner,
		results: make(chan T, buffer),
		done:    make(chan struct{}),
		started: &atomic.Bool{},
	}
}

// generator implements [types.Generator].
type generator[T any] struct {
	runner  types.Runner[T]
	results chan T
	done    chan struct{}
	err     error
	started *atomic.Bool
}

// Start implements [types.Generator.Start].
func (gen *generator[T]) Start(ctx context.Context) {
	if gen.started.Swap(true) {
		panic("attempt to start previously started generator.Generator")
	}

	h := results.NewHandle(gen.results)
	go gen.startInner(ctx, h)
}

// Run implements [types.Generator.Run].
func (gen *generator[T]) Run(ctx context.Context) error {
	gen.Start(ctx)

	return gen.Wait()
}

// Wait implements [types.Generator.Wait].
func (gen *generator[T]) Wait() error {
	// The done channel is always closed when the Runner completes, after setting w.err.
	<-gen.done

	return gen.err
}

// Results implements [types.Generator.Results].
func (gen *generator[T]) Results() <-chan T {
	return gen.results
}

// startInner is the started in the goroutine launched by [*generator.Start].
func (gen *generator[T]) startInner(ctx context.Context, h types.Handle[T]) {
	defer h.Close()
	defer close(gen.done)

	gen.err = gen.runner.Run(ctx, h)
}
