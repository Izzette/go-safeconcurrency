package examples

import (
	"context"

	"github.com/Izzette/go-safeconcurrency/api/types"
)

// IteratorJob is a simple implementation of [types.Producer] that generates integer ranges.
type IteratorJob struct {
	Start        int
	EndExclusive int
}

// Run implements [types.Producer.Run].
// It generates integers from Start to EndExclusive and emit them through the provided emitter.
func (r *IteratorJob) Run(ctx context.Context, h types.Emitter[int]) error {
	// Iterate over the range of integers from Start to EndExclusive.
	for i := r.Start; i < r.EndExclusive; i += 1 {
		// Emit the current integer through the emitter.
		if err := h.Emit(ctx, i); err != nil {
			// The context is canceled, we should stop emitting results.
			return err
		}
	}

	return nil
}
