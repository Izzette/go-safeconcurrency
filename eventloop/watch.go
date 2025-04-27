package eventloop

import (
	"context"

	"github.com/Izzette/go-safeconcurrency/api/types"
)

// WatchStateFunc is a function that is called each time a new [types.StateSnapshot] is available.
// If the function returns false, the watch will be cancelled.
type WatchStateFunc[StateT any] func(context.Context, types.StateSnapshot[StateT]) bool

// WatchState is a convenience function that watches the [types.StateSnapshot] instances of the [types.EventLoop] and
// runs the provided function each time a new snapshot is available.
//
//   - ctx: the [context.Context] to use for the watch.
//   - eventLoop: the [types.EventLoop] to watch for snapshots.
//   - watcher: the [WatchStateFunc] to call each time a new snapshot is available.
//
// It will only run one invocation of the watcher at a time.
// It returns a channel that is closed when the watch is cancelled.
//
// # Snapshots
//
// It is not guaranteed to see every [types.StateSnapshot], but the watcher will be run with the latest
// [types.StateSnapshot] available since the previous snapshot expiration was observed.
// If you have a high frequency of events or a slow watcher, you may miss some snapshots between invocations of the
// watcher.
// The watcher will never be called with the same snapshot twice.
//
// # Cancellation
//
// If the context is cancelled, the watch will be cancelled as well.
// If the function returns false on any invocation, the watch will be cancelled.
// In the case that the [types.EventLoop] is closed, the watch is guaranteed to be called with the last
// [types.StateSnapshot] before cancelling the watch.
// Whenever the watch to snapshots is cancelled, the context passed to the watcher will be cancelled, and the returned
// channel will be closed.
func WatchState[StateT any](
	ctx context.Context,
	eventLoop types.EventLoop[StateT],
	watcher WatchStateFunc[StateT],
) <-chan struct{} {
	done := make(chan struct{})
	go watchLoop(ctx, eventLoop, watcher, done)

	return done
}

// watchLoop will call the watch function each time an event is executed in the event loop.
func watchLoop[StateT any](
	ctx context.Context,
	eventLoop types.EventLoop[StateT],
	watch WatchStateFunc[StateT],
	done chan<- struct{},
) {
	defer close(done)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	snap := eventLoop.Snapshot()
	for {
		// Call the watch with the new snapshot.
		if cont := watch(ctx, snap); !cont {
			break
		}

		select {
		case <-ctx.Done():
			return
		case <-eventLoop.Done():
			// The event loop is closed, but we may have not seen the last snapshot yet because
			// select is not guaranteed to select on the first case that is ready if multiple cases
			// are ready when it is woken up.
			previousGen := snap.Generation()
			snap = eventLoop.Snapshot()
			if snap.Generation() == previousGen {
				// The event loop is closed and we have already seen the last snapshot.
				return
			}
		case <-snap.Expiration():
			// The snapshot is no longer valid, so we need to get a new one.
			snap = eventLoop.Snapshot()
		}
	}
}
