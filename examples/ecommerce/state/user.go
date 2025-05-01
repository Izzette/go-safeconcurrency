package state

import (
	"fmt"
	"strconv"

	"github.com/Izzette/go-safeconcurrency/api/types"
	"github.com/Izzette/go-safeconcurrency/eventloop/snapshot"
)

// UserState tracks individual user carts and orders
type UserState struct {
	// cart holds the user's cart items
	cart map[string]int // productID -> quantity

	// orders holds the user's orders
	orders map[string]*Order // orderID -> Order

	orderSequence int // used to generate unique order IDs
}

// NewUserSnapshot creates a [types.StateSnapshot] with an empty user state.
func NewUserSnapshot() types.StateSnapshot[*UserState] {
	return snapshot.NewCopyable(&UserState{
		cart:          make(map[string]int),
		orders:        make(map[string]*Order),
		orderSequence: 1, // Start order IDs from 1
	})
}

// NextOrderID generates a new unique order ID.
func (s *UserState) NextOrderID() string {
	orderID := strconv.Itoa(s.orderSequence)
	s.orderSequence++
	return orderID
}

// GetCart retrieves the user's cart.
// GetCart returns a copy of the cart, which can be modified without affecting the original or other copies.
// In order to persist the changes, call [UserState.UpdateCart] with the product ID and the modified cart
// It returns an empty map if the cart is empty.
func (s *UserState) GetCart() map[string]int {
	// Return a copy of the cart to avoid modifying the original
	return snapshot.CopyMap(s.cart)
}

// UpdateCart updates the user's cart with the given product ID and quantity.
// UpdateCart uses a Copy-on-Write strategy to avoid modifying the original cart.
func (s *UserState) UpdateCart(productID string, quantity int) {
	// Copy the cart to avoid modifying the original
	cart := snapshot.CopyMap(s.cart)

	// Update the cart
	if quantity == 0 {
		delete(cart, productID)
	} else {
		cart[productID] = quantity
	}

	// Replace the cart with the updated one
	s.cart = cart
}

// ClearCart clears the user's cart.
// ClearCart creates a new cart to avoid modifying the original cart.
func (s *UserState) ClearCart() {
	// Create a new cart to avoid modifying the original
	s.cart = make(map[string]int)
}

// GetOrder retrieves an order by ID.
// GetOrder returns a copy of the order, which can be modified without affecting the original or other copies.
// In order to persist the changes, call [UserState.UpdateOrder] with the order ID and the modified order
// It returns nil if the order is not found.
func (s *UserState) GetOrder(orderID string) *Order {
	return s.orders[orderID].Copy()
}

// UpdateOrder updates an order in the user state.
// It uses a Copy-on-Write strategy to avoid modifying the original order.
func (s *UserState) UpdateOrder(orderID string, order *Order) {
	// Copy the orders to avoid modifying the originals
	orders := snapshot.CopyMap(s.orders)

	// Update the order
	orders[orderID] = order.Copy()

	// Replace the orders map with the updated one
	s.orders = orders
}

// Copy implements [types.Copyable.Copy].
// A Copy-on-Write strategy is used for updates to avoid modifying the original state, so it's safe to use a shallow
// copy.
func (s *UserState) Copy() *UserState {
	return snapshot.CopyPtr(s)
}

// Order represents a user's order
type Order struct {
	// ID is the unique identifier for the order
	ID string

	// Items is a map of product IDs to quantities
	Items map[string]int

	// Status is the order status (e.g., "created", "shipped", "delivered")
	Status string
}

// String implements the [fmt.Stringer] interface.
func (o *Order) String() string {
	return fmt.Sprintf("Order ID: %s, Status: %s, Items: %v", o.ID, o.Status, o.Items)
}

// Copy implements [types.Copyable.Copy].
func (o *Order) Copy() *Order {
	return snapshot.CopyPtr(o)
}
