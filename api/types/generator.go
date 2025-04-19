package types

import "context"

// Generator objects produce one or more results asynchronously via a results channel, obtained from .Results().
// Prudence around calls to .Run() or .Wait() should be observed, as if your results channel does not have enough buffer
// to capture ALL results.
// This can lead to a deadlock if all results are not completely consumed before calling .Wait(), or results are
// consumed in another goroutine.
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
	Results() <-chan Result[T]
}
