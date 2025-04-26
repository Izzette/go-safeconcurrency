package safeconcurrencyerrors

// Stop is a special error that can be returned from the
// [github.com/Izzette/go-safeconcurrency/api/types.ResultCallback] to stop processing results.
//
//nolint:errname
const Stop = constantError("stop")
