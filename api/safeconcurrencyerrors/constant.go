package safeconcurrencyerrors

// constantError is a custom error type that can be used to create constant errors.
type constantError string

// Error implements the error interface for constantError.
func (e constantError) Error() string {
	return string(e)
}

// Unwrap implements the error interface for constantError.
func (e constantError) Unwrap() error {
	return nil
}
