package examples_test

import (
	"context"
	"fmt"
	"time"

	"github.com/Izzette/go-safeconcurrency/examples"
	"github.com/Izzette/go-safeconcurrency/generator"
)

// Demonstrates how to use the [github.com/Izzette/go-safeconcurrency/generator] package to create a simple
// [types.Generator] that produces integers from 0 to 9.
// See [IteratorJob] for the implementation of the job that generates the integers.
func Example_generator() {
	// Create a new generator with a buffer size of 1.
	// The IteratorJob generates integers from 0 to 9.
	w := generator.NewBuffered[int](&examples.IteratorJob{EndExclusive: 10}, 1)

	// Start the generator with a context that has a deadline of 2 seconds.
	ctx, _ := context.WithDeadline(context.Background(), time.Now().Add(2*time.Second))
	w.Start(ctx)

	// Read results from the generator.
	for value := range w.Results() {
		fmt.Println(value)
	}
	// Output:
	// 0
	// 1
	// 2
	// 3
	// 4
	// 5
	// 6
	// 7
	// 8
	// 9
}
