package eventloop

import (
	"context"
	"sync"
	"testing"

	"github.com/Izzette/go-safeconcurrency/api/types"
	"github.com/Izzette/go-safeconcurrency/eventloop/snapshot"
)

func TestWatchState(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a new event loop.
	el := New[*testState](snapshot.NewCopyable(&testState{}))
	defer el.Close()

	// Start the event loop.
	el.Start()

	closeStartedOnce := &sync.Once{}
	started := make(chan struct{})
	done := WatchState(ctx, el, func(ctx context.Context, snap types.StateSnapshot[*testState]) bool {
		// Notify that the watcher has started.
		closeStartedOnce.Do(func() {
			close(started)
		})

		if snap.Generation() > 0 {
			// Break on the second snapshot.
			return false
		}

		return true
	})

	select {
	case <-done:
		t.Errorf("expected done channel to not be closed yet")
	default: // do nothing
	}

	// Wait for the watcher to start before sending an event.
	<-started

	// Send an event to the event loop.
	_, err := el.Send(ctx, &testEvent{})
	if err != nil {
		t.Fatalf("failed to send event: %v", err)
	}

	// The watch should be ended soon
	_, ok := <-done
	if ok {
		t.Errorf("expected done channel to be closed, but found unexpected value in it")
	}
}

func TestWatchStateCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a new event loop.
	el := New[*testState](snapshot.NewCopyable(&testState{}))
	defer el.Close()

	// Start the event loop.
	el.Start()

	done := WatchState(ctx, el, func(ctx context.Context, snap types.StateSnapshot[*testState]) bool {
		if snap.Generation() > 0 {
			// Break on the second snapshot.
			return false
		}

		return true
	})

	cancel()

	_, ok := <-done
	if ok {
		t.Errorf("expected done channel to be closed, but found unexpected value in it")
	}
}

func TestWatchStateEventLoopClosed(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a new event loop.
	el := New[*testState](snapshot.NewCopyable(&testState{}))
	defer el.Close()

	// Start the event loop.
	el.Start()

	done := WatchState(ctx, el, func(ctx context.Context, snap types.StateSnapshot[*testState]) bool {
		if snap.Generation() > 0 {
			// Break on the second snapshot.
			return false
		}

		return true
	})

	el.Close()

	_, ok := <-done
	if ok {
		t.Errorf("expected done channel to be closed, but found unexpected value in it")
	}
}
