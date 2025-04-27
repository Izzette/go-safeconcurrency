package examples

import (
	"context"
	"fmt"
	"log"

	"github.com/Izzette/go-safeconcurrency/eventloop"
)

// Example_eventLoop demonstrates basic event loop usage
func Example_eventLoop() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	initialState := &AppState{}

	// Create event loop with 1-event buffer
	el := eventloop.NewBuffered(
		initialState,
		1,
	)
	defer el.Close()
	el.Start()

	// Send our event to the event loop
	gen, err := el.Send(ctx, &IncrementEvent{})
	if err != nil {
		log.Fatal(err)
	}

	// Wait for event processing
	snap, err := eventloop.WaitForGeneration(ctx, el, gen)
	if err != nil {
		log.Fatal(err)
	}

	// Output final state
	fmt.Printf("Counter: %d\n", snap.State().Counter)

	// Output:
	// <IncrementEvent@1> current value: 0, next value: 1
	// Counter: 1
}
