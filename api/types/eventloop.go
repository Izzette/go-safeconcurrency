package types

import "context"

// GenerationID is a monotonically increasing ID that is used to identify the state of the event loop at a given point
// in time.
type GenerationID uint64

// EventLoop dispatches [Event] instances in a loop, processing them one at a time.
//
// # Synchronization
//
// [types.EventLoop.Send] does not wait for the event to be processed, and returns immediately upon enqueuing the
// [Event].
// The returned [GenerationID] will be the ID of the [StateSnapshot] that will be available after the [Event] is
// processed.
// If you want to wait for the event to be processed, you can use the
// [github.com/Izzette/go-safeconcurrency/eventloop.WaitForGeneration] or
// [github.com/Izzette/go-safeconcurrency/eventloop.SendAndWait] helper functions.
//
// See [github.com/Izzette/go-safeconcurrency/eventloop.NewBuffered] to create event loops and for more details
// on how to use them.
type EventLoop[StateT any] interface {
	// Start initializes the event loop and prepares it for event subscription.
	Start()

	// Close closes the requests channel and waits for all events to complete.
	// Any calls to Send after Close will panic.
	// It is safe to call Close multiple times, or to call close before Start.
	Close()

	// Done returns a channel that is closed when the event loop is closed and all events have been processed.
	Done() <-chan struct{}

	// Send enqueues an event to the event loop for processing.
	// If the context is canceled, the event will not be enqueued and an error will be returned.
	// The returned GenerationID will be the ID of the StateSnapshot available after the event is processed.
	Send(context.Context, Event[StateT]) (GenerationID, error)

	// Snapshot returns a StateSnapshot with a copy of the current state of the event loop.
	// Do not modify the state pointer returned by this method, as it will be shared with all other snapshots of the same
	// GenerationID.
	Snapshot() StateSnapshot[StateT]
}

// Event represents a unit of work that can be dispatched in an event loop.
type Event[StateT any] interface {
	// Dispatch is called when the event is processed in the loop.
	//
	//  - GenerationID: the ID of the state snapshot that will be available after the event is processed.
	//  - *StateT: a pointer to a shallow copy of the state.  Updates to the state will be persisted after the event is
	//    processed.
	//
	// The event loop will be blocked until the event is processed.
	// If the state passed to the dispatch method is modified after the event is processed, the changes will not be
	// persisted (for example, if the event dispatches a goroutine).
	Dispatch(GenerationID, *StateT)
}

// StateSnapshot is an interface that represents a snapshot of the state at a given point in time.
// It is returned by the [types.EventLoop.Snapshot] method.
// Snapshots are used to provide a consistent view of the state from outside the event loop.
type StateSnapshot[StateT any] interface {
	// This returns the state at the time of the snapshot.
	// Do not modify the state pointer returned by this method, as it will be shared with all other snapshots of the same
	// GenerationID.
	State() *StateT

	// Generation returns the monotonically increasing generation ID of the state snapshot.
	// This ID is incremented each time an event is processed in the event loop.
	Generation() GenerationID

	// Expiration returns a channel which is closed when the state snapshot is no longer valid.
	// This will occur each time an event finishes processing in the event loop.
	// If the event loop is closed, this channel will never be closed as the snapshot is still valid.
	// If you would like to also check if the event loop is closed, you can use a select statement to wait for both
	// EventLoop.Done() and StateSnapshot.Expiration() channels.
	// The helper function eventloop.WaitForGeneration and eventloop.SendAndWait will do this for you.
	Expiration() <-chan struct{}
}
