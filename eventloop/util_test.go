package eventloop

import (
	"context"
	"testing"

	"github.com/Izzette/go-safeconcurrency/api/types"
	"github.com/Izzette/go-safeconcurrency/eventloop/snapshot"
)

func TestEventFromFunc(t *testing.T) {
	// Create a new event from a function
	event := EventFromFunc(func(gen types.GenerationID, s *testState) *testState {
		s.counter++

		return s
	})

	// Check if the event's function works as expected
	s := &testState{counter: 0}
	s = event.Dispatch(0, s)
	if s.counter != 1 {
		t.Fatalf("expected counter to be 1, got %d", s.counter)
	}
}

func TestSendFunc(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a new event loop
	el := New[*testState](snapshot.NewCopyable[*testState](&testState{}))
	defer el.Close()

	// Start the event loop
	el.Start()

	// Send a function event to the event loop
	gen, err := SendFunc(ctx, el, func(gen types.GenerationID, s *testState) *testState {
		s.counter++

		return s
	})
	if err != nil {
		t.Fatalf("failed to send event: %v", err)
	}

	// Wait for the event to be processed
	snap, err := WaitForGeneration(ctx, el, gen)
	if err != nil {
		t.Fatalf("failed to wait for event: %v", err)
	}

	// Check if the state is updated correctly
	if snap.State().counter != 1 {
		t.Fatalf("expected counter to be 1, got %d", snap.State().counter)
	}
}

func TestSendFuncAndWait(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a new event loop
	el := New[*testState](snapshot.NewCopyable[*testState](&testState{}))
	defer el.Close()

	// Start the event loop
	el.Start()

	// Send a function event to the event loop and wait for it to be processed
	snap, err := SendFuncAndWait(ctx, el, func(gen types.GenerationID, s *testState) *testState {
		s.counter++

		return s
	})
	if err != nil {
		t.Fatalf("failed to send and wait for event: %v", err)
	}

	// Check if the state is updated correctly
	if snap.State().counter != 1 {
		t.Fatalf("expected counter to be 1, got %d", snap.State().counter)
	}
}
