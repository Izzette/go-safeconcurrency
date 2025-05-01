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
func (e *ReserveStockEvent) Dispatch(_ types.GenerationID, s map[string]*state.Product) map[string]*state.Product {
	product := s[e.ProductID].Copy()
	if product == nil {
		e.Err = ErrProductNotFound
		return s
	}

	availableStock := product.Available()
	if availableStock < e.Quantity {
		e.Err = fmt.Errorf("insufficient stock for %s", e.ProductID)
		return s
	}
	product.Reserved += e.Quantity

	// Only persist the product changes if the stock is available
	s[e.ProductID] = product
	return s
}
