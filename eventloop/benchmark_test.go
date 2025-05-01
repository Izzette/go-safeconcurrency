package eventloop

import (
	"context"
	"testing"

	"github.com/Izzette/go-safeconcurrency/api/types"
	"github.com/Izzette/go-safeconcurrency/eventloop/snapshot"
)

// BenchmarkEventLoopThroughput measures the throughput of the [types.EventLoop].
// It sends a large number of simple [types.Event] instances, checking that the [types.StateSnapshot] is updated
// correctly, and measures the time taken.
// It's performance is heavily tied to the performance of [workpool.Submit].
func BenchmarkEventLoopThroughput(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	el := NewBuffered[int](snapshot.NewZeroValue[int](), 16)
	defer el.Close()
	defer cancel()
	el.Start()

	b.ResetTimer()
	var gen types.GenerationID
	for i := 0; i < b.N; i++ {
		var err error
		gen, err = el.Send(ctx, &benchEvent{})
		if err != nil {
			b.Fatalf("failed to send event: %v", err)
		}
	}

	// Wait for all events to process
	snap, err := WaitForGeneration(ctx, el, gen)
	if err != nil {
		b.Fatalf("failed to wait for event: %v", err)
	}

	// Stop timer to avoid including the teardown time in the benchmark
	b.StopTimer()

	if snap.State() != b.N {
		b.Errorf("expected state %d, got %d", b.N, snap.State())
	}
}

// benchEvent is an implementation of [types.Event] that increments a counter.
type benchEvent struct{}

// Dispatch implements [types.Event.Dispatch].
func (e *benchEvent) Dispatch(_ types.GenerationID, s int) int {
	s++

	return s
}
