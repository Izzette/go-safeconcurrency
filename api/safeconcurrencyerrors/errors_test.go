package safeconcurrencyerrors

import (
	"errors"
	"testing"
)

func TestStopError(t *testing.T) {
	// Check the error message.
	expectedMsg := "stop"
	if Stop.Error() != expectedMsg {
		t.Errorf("expected %q, got %q", expectedMsg, Stop.Error())
	}

	// Check the Unwrap method.
	if errors.Unwrap(Stop) != nil {
		t.Error("expected nil unwrapped error")
	}

	if !errors.Is(Stop, Stop) {
		t.Error("This would be rather silly if it didn't work")
	}
}
