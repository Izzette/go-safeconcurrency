package examples

import (
	"context"
	"fmt"

	"github.com/Izzette/go-safeconcurrency/api/types"
	"github.com/Izzette/go-safeconcurrency/eventloop"
	"github.com/Izzette/go-safeconcurrency/eventloop/snapshot"
)

// Example_eventLoop demonstrates basic event loop usage
func Example_eventLoop() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create an event loop
	el := eventloop.New(snapshot.NewValue(0))
	defer el.Close()
	el.Start()

	// Send our event to the event loop, and wait for it to be processed
	snap, err := eventloop.SendFuncAndWait(ctx, el, func(gen types.GenerationID, state int) int {
		fmt.Printf("<IncrementEvent@%d> current value: %d, next value: %d\n", gen, state, state+1)
		return state + 1
	})
	if err != nil {
		panic(err)
	}

	// Output final state
	fmt.Printf("Counter: %d\n", snap.State())

	// Output:
	// <IncrementEvent@1> current value: 0, next value: 1
	// Counter: 1
}
