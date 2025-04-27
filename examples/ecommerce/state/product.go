package state

import "fmt"

// ProductStockState tracks available inventory
type ProductStockState struct {
	products map[string]*Product
}

// NewProductStockState creates a new ProductStockState
func NewProductStockState(products map[string]*Product) *ProductStockState {
	return &ProductStockState{
		products: products,
	}
}

// GetProduct gets a product by ID.
// GetProduct returns a copy of the product, which can be modified without affecting the original or other copies.
// In order to persist the changes, call [ProductStockState.UpdateProduct] with the product ID and the modified product
// It returns nil if the product is not found.
func (s *ProductStockState) GetProduct(productID string) *Product {
	product, _ := s.products[productID]
	if product == nil {
		return nil
	}
	// Return a copy of the product to avoid modifying the original
	productCopy := *product
	return &productCopy
}

// UpdateProduct updates a product in the stock state.
// It uses a Copy-on-Write strategy to avoid modifying the original product.
func (s *ProductStockState) UpdateProduct(productID string, product *Product) {
	// Copy the products to avoid modifying the originals
	products := make(map[string]*Product)
	for k, v := range s.products {
		products[k] = v
	}

	// Update the product
	products[productID] = product

	// Replace the products map with the updated one
	s.products = products
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

func (p *Product) String() string {
	return fmt.Sprintf("Description: %s, Price: %.2f, Available: %d", p.Description, p.Price, p.Stock-p.Reserved)
}

// Available returns the number of items available for sale
func (p *Product) Available() int {
	return p.Stock - p.Reserved
}
