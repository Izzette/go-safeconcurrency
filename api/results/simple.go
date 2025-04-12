package results

import (
	"github.com/Izzette/go-safeconcurrency/api/types"
)

type simpleResult[T any] struct {
	value T
}

func (r *simpleResult[T]) Get() (T, error) {
	return r.value, nil
}

type simpleError[T any] struct {
	err error
}

func (r *simpleError[T]) Get() (T, error) {
	// Get the zero-value of type T
	var zero T

	return zero, r.err
}

// NewSimpleResult creates a types.Result, representing the successful operation.
// The result is immutable once initialized.
func NewSimpleResult[T any](value T) types.Result[T] {
	return &simpleResult[T]{value: value}
}

// NewSimpleError creates a types.Result, representing a failed operation.
// The result is immutable once initialized.
func NewSimpleError[T any](err error) types.Result[T] {
	return &simpleError[T]{err: err}
}
