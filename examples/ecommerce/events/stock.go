package events

import (
	"errors"
	"fmt"

	"github.com/Izzette/go-safeconcurrency/api/types"
	"github.com/Izzette/go-safeconcurrency/examples/ecommerce/state"
)

// ErrProductNotFound is the error returned when the product is not found
var ErrProductNotFound = errors.New("product not found")

// ReserveStockEvent attempts to reserve product inventory
type ReserveStockEvent struct {
	// ProductID is the ID of the product to reserve
	ProductID string

	// Quantity is the number of items to reserve
	Quantity int

	// Err is the error returned from the event
	Err error
}

// Dispatch implements [types.Event.Dispatch].
func (e *ReserveStockEvent) Dispatch(_ types.GenerationID, s *state.ProductStockState) {
	product := s.GetProduct(e.ProductID)
	if product == nil {
		e.Err = ErrProductNotFound
		return
	}

	availableStock := product.Available()
	if availableStock < e.Quantity {
		e.Err = fmt.Errorf("insufficient stock for %s", e.ProductID)
		return
	}
	product.Reserved += e.Quantity

	// Only persist the product changes if the stock is available
	s.UpdateProduct(e.ProductID, product)
}

// ReleaseStockEvent returns inventory to available stock
type ReleaseStockEvent struct {
	ProductID string
	Quantity  int
}

// Dispatch implements [types.Event.Dispatch].
func (e *ReleaseStockEvent) Execute(_ types.GenerationID, s *state.ProductStockState) {
	product := s.GetProduct(e.ProductID)
	if product == nil {
		fmt.Printf("ERROR: %v\n", ErrProductNotFound)
		return
	}

	if product.Reserved < e.Quantity {
		panic("cannot release more stock than reserved")
	}
	product.Reserved -= e.Quantity

	// Save back to the state in a Copy-on-Write fashion
	s.UpdateProduct(e.ProductID, product)
}
