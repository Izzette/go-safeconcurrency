package safeconcurrencysync

import "context"

// ChannelLock is a [sync.Locker] that use a channel to synchronize access allowing [context.Context] cancelation to be
// respected.
type ChannelLock struct {
	channel chan struct{}
}

// NewChannelLock creates a new [channelLock].
func NewChannelLock() *ChannelLock {
	channel := make(chan struct{}, 1)
	channel <- struct{}{} // Put the "talking-stick" in the channel.

	return &ChannelLock{
		channel: channel,
	}
}

// LockWithContext locks the [ChannelLock].
// If the [context.Context] is canceled, it returns the [context.Cause].
// If the [context.Context] is not canceled, it returns nil.
func (l *ChannelLock) LockWithContext(ctx context.Context) error {
	// select is not deterministic, and may still obtain the lock even if the context has been canceled.
	if err := context.Cause(ctx); err != nil {
		//nolint:wrapcheck
		return err
	}

	// Wait for the lock to be available or the context to be canceled, whichever comes first.
	select {
	case <-ctx.Done():
		//nolint:wrapcheck
		return context.Cause(ctx)
	case <-l.channel: // Get the "talking-stick" from the channel.
		return nil
	}
}

// Lock implements [sync.Locker.Lock].
func (l *ChannelLock) Lock() {
	<-l.channel // Get the "talking-stick" from the channel.
}

// Unlock implements [sync.Locker.Unlock].
func (l *ChannelLock) Unlock() {
	select {
	case l.channel <- struct{}{}: // Put the "talking-stick" back in the channel.
	default:
		panic("unlocking a ChannelLock that is not locked")
	}
}
