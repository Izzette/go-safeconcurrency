package events

import (
	"fmt"

	"github.com/Izzette/go-safeconcurrency/api/types"
	"github.com/Izzette/go-safeconcurrency/examples/ecommerce/state"
)

// CheckoutEvent converts cart to order
type CheckoutEvent struct {
	// OrderID is the ID of the order after creation
	OrderID string

	// Err is the error returned from the event
	Err error
}

// Dispatch implements [types.Event.Dispatch].
func (e *CheckoutEvent) Dispatch(_ types.GenerationID, s *state.UserState) *state.UserState {
	cart := s.GetCart()
	if len(cart) == 0 {
		e.Err = ErrCartEmpty
		return s
	}

	order := &state.Order{
		ID:     s.NextOrderID(),
		Items:  make(map[string]int),
		Status: "created",
	}
	for pid, qty := range s.GetCart() {
		order.Items[pid] = qty
	}
	e.OrderID = order.ID
	s.UpdateOrder(e.OrderID, order)
	s.ClearCart()

	return s
}

// UpdateOrderStatusEvent changes order status
type UpdateOrderStatusEvent struct {
	// OrderID is the ID of the order to update
	OrderID string

	// Status is the new status for the order
	Status string
}

// Dispatch implements [types.Event.Dispatch].
func (e *UpdateOrderStatusEvent) Dispatch(_ types.GenerationID, s *state.UserState) *state.UserState {
	order := s.GetOrder(e.OrderID)
	if order == nil {
		fmt.Printf("ERROR: order %s not found\n", e.OrderID)
		return s
	}
	order.Status = e.Status
	s.UpdateOrder(e.OrderID, order)
	return s
}
