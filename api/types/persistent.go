package types

// Copyable is an interface which represents objects that can be "copied".
// A copy may mean many things:
//
// - Shallow copy for Copy-on-Write / Persistent data-structures
// - Partial copy of sensitive fields
// - Deep copy of the full object
//
// If the object is a nil pointer, the Copy() method should return nil.
type Copyable[T any] interface {
	Copy() T
}

// Value is a type constraint for all type primitives that are "passed by value".
// These types are not pointers, slices, maps, channels, or functions.
// structs are not included, as they can contain types that are "passed by reference".
type Value interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uintptr |
		~float32 | ~float64 |
		~complex64 | ~complex128 |
		~bool | ~string
}
