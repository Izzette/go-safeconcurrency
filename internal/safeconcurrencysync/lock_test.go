package safeconcurrencysync

import (
	"context"
	"sync"
	"testing"
)

func TestChannelLock(t *testing.T) {
	// Create a new ChannelLock
	lock := NewChannelLock()

	// Lock the CtxLock
	lock.Lock()

	// Unlock the CtxLock
	//nolint:staticcheck // Empty critical section
	lock.Unlock()
}

func TestChannelLockDoubleUnlock(t *testing.T) {
	// Create a new ChannelLock
	lock := NewChannelLock()

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic, but got none")
		}
	}()
	// Unlock the ctxLock without locking it first
	lock.Unlock()
}

func TestChannelLockCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a new ChannelLock
	lock := NewChannelLock()

	// Cancel the context
	cancel()

	// Lock the CtxLock
	if err := lock.LockWithContext(ctx); err == nil {
		t.Errorf("Expected error, but got nil")
	}
}

func TestChannelLockContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a new ChannelLock
	lock := NewChannelLock()

	// Lock the CtxLock
	if err := lock.LockWithContext(ctx); err != nil {
		t.Errorf("Expected nil error, but got %v", err)
	}

	// Unlock the CtxLock
	lock.Unlock()
}

func TestChannelLockBlock(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a new ChannelLock
	lock := NewChannelLock()

	// Lock the CtxLock
	if err := lock.LockWithContext(ctx); err != nil {
		t.Errorf("Expected nil error, but got %v", err)
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		// Try to lock the CtxLock again
		if err := lock.LockWithContext(ctx); err != nil {
			t.Errorf("Expected no error, but got %v", err)
		}
		lock.Unlock()
	}()

	lock.Unlock()
	wg.Wait()
}
