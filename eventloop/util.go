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

// EventFromFunc creates a new event from a function that takes a [types.GenerationID] and a state and returns the next
// state.
func EventFromFunc[StateT any](fn types.EventFunc[StateT]) types.Event[StateT] {
	return &eventFunc[StateT]{fn: fn}
}

// SendFunc is a convenience function that sends a function to the [types.EventLoop].
// It is equivalent to calling [EventFromFunc] and then [types.EventLoop.Send].
func SendFunc[StateT any](
	ctx context.Context,
	eventLoop types.EventLoop[StateT],
	fn func(types.GenerationID, StateT) StateT,
) (types.GenerationID, error) {
	event := EventFromFunc(fn)

	//nolint:wrapcheck
	return eventLoop.Send(ctx, event)
}

// SendFuncAndWait is a convenience function that sends a function to the [types.EventLoop] and waits for it to be
// processed.
// It is equivalent to calling [EventFromFunc], [types.EventLoop.Send], and then [WaitForGeneration] with the returned
// generation ID.
func SendFuncAndWait[StateT any](
	ctx context.Context,
	eventLoop types.EventLoop[StateT],
	fn func(types.GenerationID, StateT) StateT,
) (types.StateSnapshot[StateT], error) {
	event := EventFromFunc(fn)

	return SendAndWait(ctx, eventLoop, event)
}

// eventFunc is an implementation of [types.Event] that wraps a function.
type eventFunc[StateT any] struct {
	fn func(types.GenerationID, StateT) StateT
}

// Dispatch implements [types.Event.Dispatch].
func (e *eventFunc[StateT]) Dispatch(gen types.GenerationID, state StateT) StateT {
	return e.fn(gen, state)
}
