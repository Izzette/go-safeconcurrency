package results

import (
	"errors"
	"testing"
)

func TestNewSimpleResult(t *testing.T) {
	expected := 42
	result := &simpleResult[int]{expected}
	val, err := result.Get()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if val != expected {
		t.Errorf("expected %d, got %d", expected, val)
	}
}

func TestNewSimpleError(t *testing.T) {
	expectedErr := errors.New("test error")
	result := &simpleError[int]{expectedErr}
	val, err := result.Get()
	if val != 0 {
		t.Errorf("expected zero value, got %d", val)
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
}
