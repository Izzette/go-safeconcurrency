package generator

import (
	"context"
	"testing"

	"github.com/Izzette/go-safeconcurrency/api/types"
)

// benchRunner is a simple implementation of the [types.Runner] interface.
type benchRunner struct {
	n int
}

// Run implements the [types.Runner.Run] method.
func (r *benchRunner) Run(ctx context.Context, h types.Handle[struct{}]) error {
	s := struct{}{}
	for i := 0; i < r.n; i++ {
		if err := h.Publish(ctx, s); err != nil {
			return err
		}
	}

	return nil
}

// BenchmarkGenerator measures the performance of generating and consuming values using a buffer size of 1.
func BenchmarkGenerator(b *testing.B) {
	runner := &benchRunner{n: b.N}
	gen := NewGeneratorBuffered[struct{}](runner, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	b.ResetTimer()
	gen.Start(ctx)

	count := 0
	for range gen.Results() {
		count++
	}

	// Stop timer before closing the generator to avoid measuring cleanup time.
	b.StopTimer()

	// Verify completion
	if err := gen.Wait(); err != nil {
		b.Error(err)
	}
	if count != b.N {
		b.Fatalf("expected %d values, got %d", b.N, count)
	}
}
