package state

import (
	"fmt"
	"strconv"
)

// UserState tracks individual user carts and orders
type UserState struct {
	// cart holds the user's cart items
	cart map[string]int // productID -> quantity

	// orders holds the user's orders
	orders map[string]*Order // orderID -> Order

	orderSequence int // used to generate unique order IDs
}

func NewUserState() *UserState {
	return &UserState{
		cart:          make(map[string]int),
		orders:        make(map[string]*Order),
		orderSequence: 1, // Start order IDs from 1
	}
}

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
	cartCopy := make(map[string]int)
	for k, v := range s.cart {
		cartCopy[k] = v
	}
	return cartCopy
}

// UpdateCart updates the user's cart with the given product ID and quantity.
// UpdateCart uses a Copy-on-Write strategy to avoid modifying the original cart.
func (s *UserState) UpdateCart(productID string, quantity int) {
	// Copy the cart to avoid modifying the original
	cart := make(map[string]int)
	for k, v := range s.cart {
		cart[k] = v
	}

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
	order, exists := s.orders[orderID]
	if !exists {
		return nil
	}
	// Return a copy of the order to avoid modifying the original
	orderCopy := *order
	return &orderCopy
}

// UpdateOrder updates an order in the user state.
// It uses a Copy-on-Write strategy to avoid modifying the original order.
func (s *UserState) UpdateOrder(orderID string, order *Order) {
	// Copy the orders to avoid modifying the originals
	orders := make(map[string]*Order)
	for k, v := range s.orders {
		orders[k] = v
	}

	// Update the order
	orders[orderID] = order

	// Replace the orders map with the updated one
	s.orders = orders
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

func (o *Order) String() string {
	return fmt.Sprintf("Order ID: %s, Status: %s, Items: %v", o.ID, o.Status, o.Items)
}
