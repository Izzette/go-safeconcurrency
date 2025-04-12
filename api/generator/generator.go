package generator

import (
	"context"
	"sync/atomic"

	"github.com/Izzette/go-safeconcurrency/api/types"
)

// Generator is a basic implementation of types.Generator.
// It is not re-startable, and thus .Start() or .Run() must only be called exactly once.
type Generator[T any] struct {
	runner  types.Runner[T]
	results chan types.Result[T]
	done    chan struct{}
	err     error
	started atomic.Bool
}

// NewGenerator creates (but does not start) a types.Generator, with no results buffering.
// If you would like results buffering, please use NewGeneratorBuffered.
func NewGenerator[T any](runner types.Runner[T]) *Generator[T] {
	return NewGeneratorBuffered(runner, 0)
}

// NewGeneratorBuffered creates (but does not start) a types.Generator with the specified results buffer size.
func NewGeneratorBuffered[T any](runner types.Runner[T], buffer uint) *Generator[T] {
	return &Generator[T]{
		runner:  runner,
		results: make(chan types.Result[T], buffer),
		done:    make(chan struct{}),
	}
}

// Start implements github.com/Izzette/go-safeconcurrency/api/types.Worker.Start.
func (gen *Generator[T]) Start(ctx context.Context) {
	if gen.started.Swap(true) {
		panic("attempt to start previously started generator.Generator")
	}

	h := &handle[T]{results: gen.results}
	go gen.startInner(ctx, h)
}

// Run implements github.com/Izzette/go-safeconcurrency/api/types.Worker.Run.
func (gen *Generator[T]) Run(ctx context.Context) error {
	gen.Start(ctx)

	return gen.Wait()
}

// Wait implements github.com/Izzette/go-safeconcurrency/api/types.Worker.Wait.
func (gen *Generator[T]) Wait() error {
	// The done channel is always closed when the Runner completes, after setting w.err.
	<-gen.done

	return gen.err
}

// Results implements github.com/Izzette/go-safeconcurrency/api/types.Generator.Results.
func (gen *Generator[T]) Results() <-chan types.Result[T] {
	return gen.results
}

func (gen *Generator[T]) startInner(ctx context.Context, h types.Handle[T]) {
	defer close(gen.results)
	defer close(gen.done)

	gen.err = gen.runner.Run(ctx, h)
}
