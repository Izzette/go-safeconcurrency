package eventloop

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/Izzette/go-safeconcurrency/api/types"
	"github.com/Izzette/go-safeconcurrency/internal/safeconcurrencysync"
	"github.com/Izzette/go-safeconcurrency/workpool"
)

// NewBuffered creates (but does not start) a basic implementation of [types.EventLoop].
//
// # Parameters
//
//   - initialSnapshot: A [types.StateSnapshot] that will be used as the initial state of the event loop.
//   - buffer: The size of the event queue. This is the number of events that can be queued before blocking on
//     publication.
//
// # State
//
// The event loop is designed to be used with a shared state, which is passed to each event when it is dispatched.
// At each event execution, the event loop will make a shallow copy of the state and pass it to the event.
// The event can then modify the state or return a new state, and the changes will be persisted to the snapshot after
// the event is finished and will be visible to future events.
//
// Use [github.com/Izzette/go-safeconcurrency/eventloop/snapshot.NewValue],
// [github.com/Izzette/go-safeconcurrency/eventloop/snapshot.NewMap],
// [github.com/Izzette/go-safeconcurrency/eventloop/snapshot.NewSlice],
// or [github.com/Izzette/go-safeconcurrency/eventloop/snapshot.NewCopyable] to create the initial state snapshot.
//
// # Generation
//
// Each state snapshot will be assigned a unique monotonically increasing generation ID, starting at 0.
// This generation ID is incremented each time after an event is processed in the event loop, and the new snapshot is
// available.
// When submitting an event, the [types.GenerationID] that will be assigned to the state snapshot after the event is
// processed is returned.
//
// # Snapshot
//
// The [types.EventLoop.Snapshot] method will return a [types.StateSnapshot], allowing access to:
//
//   - A copy of the state at the time of the snapshot.
//   - The generation ID of the snapshot.
//   - A channel that is closed when the state is no longer valid (as soon as the next event is processed).
//
// # Starting and stopping the Event Loop
//
// The [types.EventLoop.Start] method must be called to start the event loop.
// It may be called after the [types.EventLoop.Close] or [types.EventLoop.Send] methods have been called.
// It is recommended to defer the call to [types.EventLoop.Close] immediately after creating the event loop to avoid
// leaking the goroutines used to process events and any references they may prevent from being garbage collected.
func NewBuffered[StateT any](initialSnapshot types.StateSnapshot[StateT], buffer uint) types.EventLoop[StateT] {
	snapshotPtr := &atomic.Pointer[types.StateSnapshot[StateT]]{}
	snapshotPtr.Store(&initialSnapshot)
	eventPool := workpool.NewBuffered[*atomic.Pointer[types.StateSnapshot[StateT]]](snapshotPtr, 1, buffer)

	return &eventLoop[StateT]{
		done:      make(chan struct{}),
		closeOnce: &sync.Once{},

		eventPool: eventPool,

		generation:     initialSnapshot.Generation(),
		generationLock: safeconcurrencysync.NewChannelLock(),

		snapshotPtr: snapshotPtr,
	}
}

// New creates (but does not start) a basic implementation of [types.EventLoop].
// It is equivalent to calling [NewBuffered] with a buffer size of 0.
// If you would like to use a buffer size, use [NewBuffered] instead.
func New[StateT any](initialSnapshot types.StateSnapshot[StateT]) types.EventLoop[StateT] {
	return NewBuffered(initialSnapshot, 0)
}

// eventLoop is an implementation of [types.EventLoop].
type eventLoop[StateT any] struct {
	done      chan struct{}
	closeOnce *sync.Once

	eventPool types.WorkerPool[*atomic.Pointer[types.StateSnapshot[StateT]]]

	generation     types.GenerationID
	generationLock *safeconcurrencysync.ChannelLock

	snapshotPtr *atomic.Pointer[types.StateSnapshot[StateT]]
}

// Start implements [types.EventLoop.Start].
func (l *eventLoop[StateT]) Start() {
	l.eventPool.Start()
}

// Close implements [types.EventLoop.Close].
func (l *eventLoop[StateT]) Close() {
	// Close the pools and wait for all tasks to complete.
	l.eventPool.Close()
	l.closeOnce.Do(l.closeDone)
}

// Done implements [types.EventLoop.Done].
func (l *eventLoop[StateT]) Done() <-chan struct{} {
	return l.done
}

// Send implements [types.EventLoop.Send].
func (l *eventLoop[StateT]) Send(ctx context.Context, event types.Event[StateT]) (types.GenerationID, error) {
	// select is not deterministic, and may still send tasks even if the context has been canceled.
	if err := context.Cause(ctx); err != nil {
		//nolint:wrapcheck
		return 0, err
	}

	// The eventTask will be submitted to the l.pool by the submitEventTask.
	eventTask := &eventWrapper[StateT]{
		event: event,
	}

	// Obtain the lock to ensure that the generation ID is incremented once for each event and that the order of events is
	// aligned with the order of generation IDs.
	if err := l.generationLock.LockWithContext(ctx); err != nil {
		//nolint:wrapcheck
		return 0, err
	}
	defer l.generationLock.Unlock()

	select {
	case <-ctx.Done():
		//nolint:wrapcheck
		return 0, context.Cause(ctx)
	case l.eventPool.Requests() <- eventTask:
		l.generation++ // increment the generation ID only if the event task is successfully submitted.

		return l.generation, nil
	}
}

// Snapshot implements [types.EventLoop.Snapshot].
func (l *eventLoop[StateT]) Snapshot() types.StateSnapshot[StateT] {
	return *l.snapshotPtr.Load()
}

// closeDone closes the done channel.
// it does not synchronize with [eventLoop.closeOnce].
func (l *eventLoop[StateT]) closeDone() {
	close(l.done)
}

// eventWrapper is a wrapper for a [types.Event] implementing [types.ValuelessTask].
type eventWrapper[StateT any] struct {
	event types.Event[StateT]
}

// Execute implements [types.ValuelessTask.Execute].
func (e *eventWrapper[StateT]) Execute(resource *atomic.Pointer[types.StateSnapshot[StateT]]) {
	snapshot := *resource.Load()
	// Close the previous snapshot expiration channel signalling that a new state is available.
	defer snapshot.Expire()

	// Call the event's Dispatch method with the resource.
	state := e.event.Dispatch(snapshot.Generation()+1, snapshot.State())

	// Create a new stateGeneration with the new state and increment the generation.
	nextSnapshot := snapshot.Next(state)

	resource.Store(&nextSnapshot)
}
