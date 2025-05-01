package eventloop

import (
	"context"
	"errors"
	"testing"

	"github.com/Izzette/go-safeconcurrency/api/safeconcurrencyerrors"
	"github.com/Izzette/go-safeconcurrency/api/types"
	"github.com/Izzette/go-safeconcurrency/eventloop/snapshot"
)

type testEvent struct {
	fn func(gen types.GenerationID, s *testState) *testState
}

// Dispatch implements [types.Event] for testEvent.
func (e *testEvent) Dispatch(gen types.GenerationID, s *testState) *testState {
	if e.fn == nil {
		return s
	}

	return e.fn(gen, s)
}

type testState struct {
	counter int
}

// Copy implements [types.Copyable] for testState.
func (s *testState) Copy() *testState {
	if s == nil {
		return nil
	}

	return &testState{
		counter: s.counter,
	}
}

func TestNewEventLoop(t *testing.T) {
	el := New[int](snapshot.NewZeroValue[int]())
	defer el.Close()

	if el == nil {
		t.Fatal("expected non-nil EventLoop")
	}
}

func TestEventLoopBasicOperation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	initialSnapshot := snapshot.NewCopyable(&testState{counter: 0})
	el := NewBuffered[*testState](initialSnapshot, 10)
	defer el.Close()
	el.Start()

	// Send increment events
	for i := 0; i < 5; i++ {
		if gen, err := el.Send(ctx, &testEvent{
			fn: func(_ types.GenerationID, s *testState) *testState {
				s.counter++

				return s
			},
		}); err != nil {
			t.Fatalf("failed to send event: %v", err)
		} else if gen != types.GenerationID(i+1) {
			t.Errorf("expected generation %d, got %d", i+1, gen)
		}
	}
	// Wait for events to process
	snap, err := WaitForGeneration(ctx, el, 5)
	if err != nil {
		t.Fatalf("failed to wait for event: %v", err)
	}
	if snap.Generation() != 5 {
		t.Fatalf("expected generation 5, got %d", snap.Generation())
	}
	if snap.State().counter != 5 {
		t.Fatalf("expected counter to be 5, got %d", snap.State().counter)
	}
}

func TestEventLoopSnapshotGeneration(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	initialSnapshot := snapshot.NewCopyable(&testState{counter: 0})
	el := NewBuffered[*testState](initialSnapshot, 0)
	defer el.Close()
	el.Start()

	// Get initial snapshot
	snap0 := el.Snapshot()
	if snap0.Generation() != 0 {
		t.Fatalf("expected initial counter to be 0, got %d", snap0.Generation())
	}

	// Send first event
	if gen, err := el.Send(
		ctx, &testEvent{fn: func(gen types.GenerationID, s *testState) *testState {
			if gen != 1 {
				t.Errorf("expected generation 1, got %d", gen)
			}
			s.counter++

			return s
		}},
	); err != nil {
		t.Fatalf("failed to send event: %v", err)
	} else if gen != 1 {
		t.Errorf("expected generation 1, got %d", gen)
	}

	snap1, err := WaitForGeneration(ctx, el, 1)
	if err != nil {
		t.Errorf("failed to wait for event: %v", err)
	} else if snap1.Generation() != 1 {
		t.Errorf("expected counter to be 1, got %d", snap1.Generation())
	}
}

func TestEventLoopContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	initialSnapshot := snapshot.NewCopyable(&testState{counter: 0})
	el := NewBuffered[*testState](initialSnapshot, 0)
	defer el.Close()
	el.Start()

	// Cancel context before sending
	cancel()

	_, err := el.Send(ctx, &testEvent{})
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled error, got %v", err)
	}
}

func TestWaitForCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	initialSnapshot := snapshot.NewCopyable(&testState{counter: 0})
	el := New[*testState](initialSnapshot)
	defer el.Close()
	el.Start()

	// Cancel context before waiting
	cancel()

	snap, err := WaitForGeneration(ctx, el, 1)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled error, got %v", err)
	}
	if snap != nil {
		t.Errorf("expected nil snapshot, got %v", snap)
	}
}

func TestWaitForLoopClosed(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	initialSnapshot := snapshot.NewCopyable(&testState{counter: 0})
	el := NewBuffered[*testState](initialSnapshot, 0)
	defer el.Close()
	el.Start()

	// Close the event loop
	el.Close()

	snap, err := WaitForGeneration(ctx, el, 1)
	if !errors.Is(err, safeconcurrencyerrors.ErrEventLoopClosed) {
		t.Errorf("expected ErrEventLoopClosed error, got %v", err)
	}
	if snap != nil {
		t.Errorf("expected nil snapshot, got %v", snap)
	}
}

func TestSendAndWait(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	initialSnapshot := snapshot.NewCopyable(&testState{counter: 0})
	el := NewBuffered[*testState](initialSnapshot, 0)
	defer el.Close()
	el.Start()

	// Send an event and wait for it
	snap, err := SendAndWait[*testState](ctx, el, &testEvent{
		fn: func(gen types.GenerationID, s *testState) *testState {
			if gen != 1 {
				t.Errorf("expected generation 1, got %d", gen)
			}
			s.counter++

			return s
		},
	})
	if err != nil {
		t.Errorf("failed to send and wait for event: %v", err)
	}
	if snap.State().counter != 1 {
		t.Errorf("expected counter to be 1, got %d", snap.State().counter)
	}
	if snap.Generation() != 1 {
		t.Errorf("expected generation 1, got %d", snap.Generation())
	}
}

func TestEventLoopClosed(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	initialSnapshot := snapshot.NewCopyable(&testState{})
	el := NewBuffered[*testState](initialSnapshot, 1)
	defer el.Close()
	el.Start()

	// Close the loop after sending
	el.Close()

	snap, err := WaitForGeneration(ctx, el, 3)
	if !errors.Is(err, safeconcurrencyerrors.ErrEventLoopClosed) {
		t.Errorf("expected ErrEventLoopClosed error, got %v", err)
	}
	if snap != nil {
		t.Errorf("expected nil snapshot, got %v", snap)
	}
	if el.Snapshot().Generation() != 0 {
		t.Errorf("expected generation 0, got %d", el.Snapshot().Generation())
	}
}

func TestSendAndWaitSendCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	initialSnapshot := snapshot.NewCopyable(&testState{})
	el := New[*testState](initialSnapshot)
	defer el.Close()
	el.Start()

	// Cancel the context before sending
	cancel()
	_, err := SendAndWait[*testState](ctx, el, &testEvent{})
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled error, got %v", err)
	}
	// Flush the event loop
	el.Close()
	if el.Snapshot().Generation() != 0 {
		t.Errorf("expected generation 0, got %d", el.Snapshot().Generation())
	}
}

func TestSendAndWaitWaitCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	initialSnapshot := snapshot.NewCopyable(&testState{})
	el := New[*testState](initialSnapshot)
	defer el.Close()
	el.Start()

	// Send an event which will cancel the context and wait
	waitCtx, waitCancel := context.WithCancel(context.Background())
	defer waitCancel()

	_, err := SendAndWait[*testState](ctx, el, &testEvent{fn: func(gen types.GenerationID, s *testState) *testState {
		cancel()
		<-waitCtx.Done()

		return s
	}})
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled error, got %v", err)
	}
	waitCancel()
	// Flush the event loop
	el.Close()
	if el.Snapshot().Generation() != 1 {
		t.Errorf("expected generation 1, got %d", el.Snapshot().Generation())
	}
}
