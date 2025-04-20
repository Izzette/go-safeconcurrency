package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Izzette/go-safeconcurrency/api/generator"
	"github.com/Izzette/go-safeconcurrency/api/types"
)

func main() {
	// Create a new generator with a buffer size of 1.
	w := generator.NewGeneratorBuffered[int](&myJob{}, 1)

	// Start the generator with a context that has a deadline of 2 seconds.
	ctx, cancelFunc := context.WithDeadline(context.Background(), time.Now().Add(2*time.Second))
	w.Start(ctx)

	// Use a deferred function to ensure that the generator is stopped and any errors are handled.
	defer func() {
		// Cancel the context to stop the generator.
		cancelFunc()

		if err := w.Wait(); err != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] Generator produced a fatal error: %q\n", err)
		}
	}()

	// Read results from the generator.
	for value := range w.Results() {
		fmt.Printf("%d\n", value)
	}
}

// myJob is a simple implementation of [types.Runner] that generates integers from 0 to 9.
type myJob struct{}

// Run implements [types.Runner.Run].
func (r *myJob) Run(ctx context.Context, h types.Handle[int]) error {
	for i := 0; i < 10; i += 1 {
		if cancelation := h.Publish(ctx, i); cancelation != nil {
			return cancelation
		}
	}

	return nil
}
