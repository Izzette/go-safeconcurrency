package types

import (
	"context"
)

// Handle is passed to a [types.Runner.Run] and [types.MultiResultTask.Execute] methods to propagate results.
type Handle[T any] interface {
	// Publish publishes an result from the Runner.
	// If an error is returned, it represents context cancelation.
	// In these cases, the Runner should immediately cleanup, and return ASAP.
	Publish(context.Context, T) error

	// Close closes the handle and releases any resources.
	Close()
}
