package types

import "context"

// Worker represents asynchronous work.
type Worker[T any] interface {
	// Start launches a goroutine to produce the results.
	Start(context.Context)

	// Run launches a goroutine, and than blocks until it's completion or fatal error.
	Run(context.Context) error

	// Wait blocks until the completion of goroutine launched by .Start() or .Run().
	// It will block indefinitely if .Start() is not called.
	Wait() error
}
