package events

import (
	"errors"

	"github.com/Izzette/go-safeconcurrency/api/types"
	"github.com/Izzette/go-safeconcurrency/examples/ecommerce/state"
)

// ErrCartEmpty is the error returned when the cart is empty
var ErrCartEmpty = errors.New("cart is empty")

// AddToCartEvent adds items to user's cart
type AddToCartEvent struct {
	// ProductID is the ID of the product to add
	ProductID string

	// Quantity is the number of items to add
	Quantity int
}

// Dispatch implements [types.Event.Dispatch].
func (e *AddToCartEvent) Dispatch(_ types.GenerationID, s *state.UserState) *state.UserState {
	s.UpdateCart(e.ProductID, e.Quantity)
	return s
}

// RemoveFromCartEvent removes items from cart
type RemoveFromCartEvent struct {
	// ProductID is the ID of the product to remove
	ProductID string
}

// Dispatch implements [types.Event.Dispatch].
func (e *RemoveFromCartEvent) Dispatch(_ types.GenerationID, s *state.UserState) {
	// Remove the product from the cart
	s.UpdateCart(e.ProductID, 0)
}
