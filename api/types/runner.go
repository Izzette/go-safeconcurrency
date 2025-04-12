package types

import "context"

// Runner represents a piece of runnable code.
type Runner[T any] interface {
	// Run should start the Runner and block until all work is completed, optionally returning a error.
	// Recoverable errors should be sent instead to Handle.Error(), rather than returning.
	// Recoverable errors are those that do not necessitate stopping the work and returning.
	// Fatal errors are those that necessitate stopping the work, in these cases an error should be returned.
	// The same context passed to Run() should be passed to Handle.Publish() and Handle.Error().
	// Run MUST cleanup and return as soon as possible after a call to Handle.Publish() or Handle.Error() return a
	// cancellation error.
	Run(context.Context, Handle[T]) error
}
