package snapshot

// CopyMap creates a shallow copy of the map m and returns it.
func CopyMap[KeyT comparable, ValueT any](m map[KeyT]ValueT) map[KeyT]ValueT {
	cpy := make(map[KeyT]ValueT, len(m))
	for k, v := range m {
		cpy[k] = v
	}

	return cpy
}

// CopySlice creates a shallow copy of the slice s and returns it.
func CopySlice[T any](s []T) []T {
	cpy := make([]T, len(s))
	copy(cpy, s)

	return cpy
}

// CopyPtr creates a shallow copy of value at the pointer ptr and returns a pointer to the copy.
func CopyPtr[T any](ptr *T) *T {
	// If the pointer is nil, return nil.
	if ptr == nil {
		return nil
	}

	// Allocate a new pointer to the type T and copy the value from the original pointer.
	cpy := new(T)
	*cpy = *ptr

	// Return the copy.
	return cpy
}
