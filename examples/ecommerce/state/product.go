package state

import (
	"fmt"

	"github.com/Izzette/go-safeconcurrency/api/types"
	"github.com/Izzette/go-safeconcurrency/eventloop/snapshot"
)

// NewProductStockSnapshot creates a types.StateSnapshot with the provided product stock state.
func NewProductStockSnapshot(products map[string]*Product) types.StateSnapshot[map[string]*Product] {
	return snapshot.NewMap[string, *Product](products)
}

// Product represents an item in the store
type Product struct {
	// Description is the product description
	Description string

	// Price is the price of the product
	Price float64

	// Stock is the number of items available for sale
	Stock int

	// Reserved is the number of items reserved (in cart) but not yet purchased
	Reserved int
}

// String implements the [fmt.Stringer] interface.
func (p *Product) String() string {
	return fmt.Sprintf("Description: %s, Price: %.2f, Available: %d", p.Description, p.Price, p.Stock-p.Reserved)
}

// Available returns the number of items available for sale
func (p *Product) Available() int {
	return p.Stock - p.Reserved
}

// Copy implements [types.Copyable.Copy].
func (p *Product) Copy() *Product {
	if p == nil {
		return nil
	}

	return &Product{
		Description: p.Description,
		Price:       p.Price,
		Stock:       p.Stock,
		Reserved:    p.Reserved,
	}
}
