package results

import "github.com/Izzette/go-safeconcurrency/api/types"

// DrainResultChannel consumes all results from a channel to prevent goroutine leaks or deadlocks.
// Should typically be deferred immediately after getting a result channel.
func DrainResultChannel[T any](resultCh <-chan types.Result[T]) {
	for range resultCh {
		// Drain the channel
	}
	// The worker or task will close the result channel when the task is done.
}
