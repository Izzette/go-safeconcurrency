package eventloop

import (
	"context"

	"github.com/Izzette/go-safeconcurrency/api/safeconcurrencyerrors"
	"github.com/Izzette/go-safeconcurrency/api/types"
)

// WaitForGeneration waits for the [types.EventLoop] to process all [types.Events] up to (and including) the specified
// [types.GenerationID].
// It returns a [types.StateSnapshot] at or later than the [types.GenerationID].
// If the [context.Context] was cancelled before the [types.EventLoop] reaches that [types.GenerationID], an error is
// returned instead.
// If the [types.EventLoop] is closed before the [types.GenerationID] is reached, an error is returned instead.
func WaitForGeneration[StateT any](
	ctx context.Context,
	eventLoop types.EventLoop[StateT],
	gen types.GenerationID,
) (types.StateSnapshot[StateT], error) {
	for {
		snapshot := eventLoop.Snapshot()
		if snapshot.Generation() >= gen {
			// The snapshot is later than the requested generation, so we can return it.
			return snapshot, nil
		}

		// Wait for the expiration channel to be closed.
		select {
		case <-ctx.Done():
			//nolint:wrapcheck
			return nil, context.Cause(ctx)
		case <-eventLoop.Done():
			// The event loop is closed, but we may have not seen the last snapshot yet because the select is not
			// guaranteed to select on the first case that is ready if multiple cases are ready when it is woken up.
			snapshot = eventLoop.Snapshot()
			if snapshot.Generation() >= gen {
				// The snapshot is later than the requested generation, so we can return it.
				return snapshot, nil
			}
			// The event loop is closed before the requested generation is reached.
			return nil, safeconcurrencyerrors.ErrEventLoopClosed
		case <-snapshot.Expiration():
		}
	}
}

// SendAndWait is a convenience function that sends an event to the [types.EventLoop] and waits for it to be processed.
// It is equivalent to calling [types.EventLoop.Send] and then [WaitForGeneration] with the returned generation ID.
func SendAndWait[StateT any](
	ctx context.Context,
	eventLoop types.EventLoop[StateT],
	event types.Event[StateT],
) (types.StateSnapshot[StateT], error) {
	gen, err := eventLoop.Send(ctx, event)
	if err != nil {
		//nolint:wrapcheck
		return nil, err
	}

	// Wait for the event to be processed.
	snap, err := WaitForGeneration(ctx, eventLoop, gen)
	if err != nil {
		return nil, err
	}

	return snap, nil
}

// copyPtr creates a shallow copy of value at the pointer p and returns a pointer to the copy.
func copyPtr[T any](ptr *T) *T {
	// If the pointer is nil, return nil.
	if ptr == nil {
		return nil
	}

	// Allocate a new pointer to the type T and copy the value from the original pointer.
	cpy := new(T)
	*cpy = *ptr

	// Return the copy.
	return cpy
}
