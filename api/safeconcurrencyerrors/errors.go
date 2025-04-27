package safeconcurrencyerrors

// Stop is a special error that can be returned from the
// [github.com/Izzette/go-safeconcurrency/workpool.ResultCallback] to stop processing results.
//
//nolint:errname
const Stop = constantError("stop")

// ErrEventLoopClosed is returned when the event loop is closed and no more snapshots will be available.
const ErrEventLoopClosed = constantError("event loop closed")
