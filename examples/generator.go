package examples

import (
	"context"

	"github.com/Izzette/go-safeconcurrency/api/types"
)

// IteratorJob is a simple implementation of [types.Runner] that generates integer ranges.
type IteratorJob struct {
	Start        int
	EndExclusive int
}

// Run implements [types.Runner.Run].
// It generates integers from Start to EndExclusive and publishes them to the provided handle.
func (r *IteratorJob) Run(ctx context.Context, h types.Handle[int]) error {
	// Iterate over the range of integers from Start to EndExclusive.
	for i := r.Start; i < r.EndExclusive; i += 1 {
		// Publish the current integer to the handle.
		if err := h.Publish(ctx, i); err != nil {
			// The context is canceled, we should stop publishing results.
			return err
		}
	}

	return nil
}
