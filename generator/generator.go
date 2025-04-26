package generator

import (
	"context"
	"sync/atomic"

	"github.com/Izzette/go-safeconcurrency/api/types"
	"github.com/Izzette/go-safeconcurrency/results"
)

// New creates (but does not start) a basic implementation of [types.Generator] with no results buffering.
// If you would like results buffering, use [NewBuffered] instead.
// This is equivalent to calling [NewBuffered] with a buffer size of 0.
func New[T any](producer types.Producer[T]) types.Generator[T] {
	return NewBuffered(producer, 0)
}

// NewBuffered creates (but does not start) a basic implementation of [types.Generator] with the specified
// results buffer size.
// It is not re-startable, and thus [types.Generator.Start] or [types.Generator.Run] must only be called exactly once.
func NewBuffered[T any](producer types.Producer[T], buffer uint) types.Generator[T] {
	return &generator[T]{
		producer: producer,
		results:  make(chan T, buffer),
		done:     make(chan struct{}),
		started:  &atomic.Bool{},
	}
}

// generator implements [types.Generator].
type generator[T any] struct {
	producer types.Producer[T]
	results  chan T
	done     chan struct{}
	err      error
	started  *atomic.Bool
}

// Start implements [types.Generator.Start].
func (gen *generator[T]) Start(ctx context.Context) {
	if gen.started.Swap(true) {
		panic("attempt to start previously started generator.Generator")
	}

	h := results.NewEmitter(gen.results)
	go gen.startInner(ctx, h)
}

// Run implements [types.Generator.Run].
func (gen *generator[T]) Run(ctx context.Context) error {
	gen.Start(ctx)

	return gen.Wait()
}

// Wait implements [types.Generator.Wait].
func (gen *generator[T]) Wait() error {
	// The done channel is always closed when the Producer completes, after setting w.err.
	<-gen.done

	return gen.err
}

// Results implements [types.Generator.Results].
func (gen *generator[T]) Results() <-chan T {
	return gen.results
}

// startInner is the started in the goroutine launched by [*generator.Start].
func (gen *generator[T]) startInner(ctx context.Context, h types.Emitter[T]) {
	defer h.Close()
	defer close(gen.done)

	gen.err = gen.producer.Run(ctx, h)
}
