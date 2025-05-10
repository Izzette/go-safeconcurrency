package safeconcurrencysync

import (
	"context"
	"runtime"
	"sync"
	"testing"
)

// BenchmarkChannelLock measures the performance of the [ChannelLock.Lock] with a single goroutine and no lock
// / contention.
func BenchmarkChannelLock(b *testing.B) {
	// Create a new ChannelLock
	lock := NewChannelLock()

	// Reset the timer after initializing the context and lock to avoid measuring setup time.
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		lock.Lock()
		//nolint:staticcheck // Empty critical section
		lock.Unlock()
	}

	// Stop timer before canceling the context to avoid measuring cleanup time.
	b.StopTimer()
}

// BenchmarkChannelLockN measures the performance of the [ChannelLock.Lock] with nproc-goroutines all contending for the
// lock before immediately unlocking it.
func BenchmarkChannelLockN(b *testing.B) {
	// Create a new ChannelLock
	lock := NewChannelLock()

	numCPU := runtime.NumCPU()
	doneWg := &sync.WaitGroup{} // doneWg signals when all goroutines are done
	doneWg.Add(numCPU)          // Add the number of goroutines to the wait group

	// Reset the timer after initializing the context and lock to avoid measuring setup time.
	b.ResetTimer()

	for cpu := 0; cpu < numCPU; cpu++ {
		go func(cpu int) {
			defer doneWg.Done()

			for i := cpu; i < b.N; i += numCPU {
				lock.Lock()
				//nolint:staticcheck // Empty critical section
				lock.Unlock()
			}
		}(cpu)
	}

	// Wait for all goroutines to finish
	doneWg.Wait()

	// Stop timer before canceling the context to avoid measuring cleanup time.
	b.StopTimer()
}

// BenchmarkChannelLockWithContext measures the performance of the [ChannelLock.LockWithContextl] with a single
// goroutine and no lock contention.
func BenchmarkChannelLockWithContext(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a new ChannelLock
	lock := NewChannelLock()

	// Reset the timer after initializing the context and lock to avoid measuring setup time.
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := lock.LockWithContext(ctx); err != nil {
			b.Errorf("Expected nil error, but got %v", err)
		}
		lock.Unlock()
	}

	// Stop timer before canceling the context to avoid measuring cleanup time.
	b.StopTimer()
}

// BenchmarkChannelLockWithContextN measures the performance of the [ChannelLock.LockWithContext] with nproc-goroutines
// all contending for the lock before immediately unlocking it.
func BenchmarkChannelLockWithContextN(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a new ChannelLock
	lock := NewChannelLock()

	numCPU := runtime.NumCPU()
	doneWg := &sync.WaitGroup{} // doneWg signals when all goroutines are done
	doneWg.Add(numCPU)          // Add the number of goroutines to the wait group

	// Reset the timer after initializing the context and lock to avoid measuring setup time.
	b.ResetTimer()

	for cpu := 0; cpu < numCPU; cpu++ {
		go func(cpu int) {
			defer doneWg.Done()

			for i := cpu; i < b.N; i += numCPU {
				if err := lock.LockWithContext(ctx); err != nil {
					b.Errorf("Expected nil error, but got %v", err)

					break
				}
				lock.Unlock()
			}
		}(cpu)
	}

	// Wait for all goroutines to finish
	doneWg.Wait()

	// Stop timer before canceling the context to avoid measuring cleanup time.
	b.StopTimer()
}
