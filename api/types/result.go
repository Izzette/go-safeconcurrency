package types

// Result represents a result, or error to produce a result of work.
type Result[T any] interface {
	// Get returns a pointer to the result if the work was successful, otherwise an error.
	Get() (T, error)
}
