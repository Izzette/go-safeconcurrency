package results

import (
	"errors"
	"testing"

	"github.com/Izzette/go-safeconcurrency/api/types"
)

func TestDrainResultChannel(t *testing.T) {
	// Create a buffered channel with 3 items
	ch := make(chan types.Result[int], 3)
	ch <- types.Result[int](&simpleResult[int]{42})
	ch <- types.Result[int](&simpleResult[int]{24})
	ch <- types.Result[int](&simpleError[int]{errors.New("test")})
	close(ch)

	// Drain should remove all 3 items
	DrainResultChannel(ch)

	// Verify channel is empty
	if len(ch) != 0 {
		t.Errorf("Expected drained channel, found %d items", len(ch))
	}
}
