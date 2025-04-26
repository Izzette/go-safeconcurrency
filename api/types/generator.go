package types

import "context"

// Generator objects produce one or more results asynchronously via a results channel.
// Prudence around calls to [types.Generator.Run] or [types.Generator.Wait] should be observed, as if your results
// channel does not have enough buffer to capture ALL results produced by the generator, the generator will block until
// the results channel is drained.
// This can lead to a deadlock if all results are not completely consumed before calling [types.Generator.Wait], and no
// mechanism to drain the results channel from another goroutine is in place.
type Generator[T any] interface {
	// Start launches a goroutine to produce the results.
	Start(context.Context)

	// Run launches a goroutine, and then blocks until it's completion or fatal error.
	Run(context.Context) error

	// Wait blocks until the completion of goroutine launched by .Start() or .Run().
	// It will block indefinitely if .Start() is not called.
	Wait() error

	// Results returns a channel which can be used to consume the results of the producer.
	// It will block indefinitely if .Start() is not called.
	Results() <-chan T
}

// Producer represents a piece of runnable code.
// It is used by [Generator] to produce results.
type Producer[T any] interface {
	// Run should start the producer and block until all work is completed, optionally returning a error.
	// The same context passed to Run() should be passed to Emitter.Emit().
	// Run MUST cleanup and return as soon as possible after a call to Emitter.Emit() returns a cancellation error.
	Run(context.Context, Emitter[T]) error
}
