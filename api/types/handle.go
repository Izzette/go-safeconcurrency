package types

import (
	"context"
)

// Handle is passed to a Runner's Run method to propagate results and recoverable errors to the Worker.
type Handle[T any] interface {
	// Publish publishes an result from the Runner.
	// If an error is returned, it represents context cancelation.
	// In these cases, the Runner should immediately cleanup, and return ASAP.
	Publish(context.Context, T) error

	// Error propagates a recoverable error emitted from the Runner.
	// If an error is returned, it represents context cancelation.
	// In these cases, the Runner should immediately cleanup, and return ASAP.
	Error(context.Context, error) error

	// Close closes the handle and releases any resources.
	Close()
}
