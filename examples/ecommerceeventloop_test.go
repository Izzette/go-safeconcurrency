package examples

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/Izzette/go-safeconcurrency/examples/ecommerce"
	"github.com/Izzette/go-safeconcurrency/examples/ecommerce/state"
)

// Example_ecommerceEventLoop demonstrates the usage of the event loop to manage inventory and user actions
// within an e-commerce application.
// It's implementation is complex, see [github.com/Izzette/go-safeconcurrency/examples/ecommerce] for details.
func Example_ecommerceEventLoop() {
	// Initialize server with initial stock
	server := ecommerce.NewEcommerceServer(map[string]*state.Product{
		"mittens": {
			Description: "Warm mittens",
			Price:       19.99,
			Stock:       10,
		},
		"golden_eggs": {
			Description: "Golden eggs",
			Price:       49.99,
			Stock:       5,
		},
		"unobtainium": {
			Description: "In this economy?!",
			Price:       999.99,
			Stock:       0,
		},
	})
	defer server.Stop() // Ensure the server is stopped when done along with all the event loops
	server.Start()      // Start the server (main product event loop)

	// Create a test server to simulate API requests
	ts := httptest.NewServer(server)
	defer ts.Close()

	// Query product
	resp, err := http.Get(ts.URL + "/api/product?product_id=mittens")
	if err != nil {
		fmt.Printf("Error fetching product: %v\n", err)
		return
	} else if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error fetching product: %s\n", resp.Status)
		return
	}
	product := &state.Product{}
	if err := json.NewDecoder(resp.Body).Decode(product); err != nil {
		fmt.Printf("Error decoding product: %v\n", err)
		return
	}

	// Display product details
	fmt.Printf("Product: %v\n", product)

	// Adding to cart
	userID := "bob"
	reserveRequest := &ecommerce.ReserveProductRequest{
		Quantity: 3,
	}
	reserveRequestBytes, err := json.Marshal(reserveRequest)
	if err != nil {
		fmt.Printf("Error marshaling request: %v\n", err)
		return
	}
	resp, err = http.Post(
		ts.URL+"/api/product/reserve?product_id=mittens&user_id="+url.QueryEscape(userID),
		"application/json",
		bytes.NewBuffer(reserveRequestBytes),
	)
	if err != nil {
		fmt.Printf("Error reserving product: %v\n", err)
	} else if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error reserving product: %s\n", resp.Status)
	} else {
		fmt.Println("Product reserved successfully")
	}

	// Check remaining stock
	resp, err = http.Get(ts.URL + "/api/product?product_id=mittens")
	if err != nil {
		fmt.Printf("Error fetching product: %v\n", err)
		return
	}
	json.NewDecoder(resp.Body).Decode(&product)
	fmt.Printf("Product after reservation: %v\n", product)

	// Checkout
	resp, err = http.Post(
		ts.URL+"/api/user/order?user_id="+url.QueryEscape(userID),
		"application/x-empty",
		bytes.NewReader([]byte{}),
	)
	if err != nil {
		fmt.Printf("Error checking out: %v\n", err)
		return
	} else if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error checking out: %s\n", resp.Status)
		return
	}
	order := &state.Order{}
	if err := json.NewDecoder(resp.Body).Decode(order); err != nil {
		fmt.Printf("Error decoding order: %v\n", err)
		return
	}
	fmt.Printf("Checkout started successfully: %v\n", order)

	// Display cart
	resp, err = http.Get(ts.URL + "/api/user/cart?user_id=" + url.QueryEscape(userID))
	if err != nil {
		fmt.Printf("Error fetching cart: %v\n", err)
		return
	} else if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error fetching cart: %s\n", resp.Status)
		return
	}
	cart := make(map[string]int)
	if err := json.NewDecoder(resp.Body).Decode(&cart); err != nil {
		fmt.Printf("Error decoding cart: %v\n", err)
		return
	}
	fmt.Printf("Cart: %v\n", cart)

	// Update order status
	updateOrderStatusRequest := &ecommerce.UpdateOrderStatusRequest{
		Status: "shipped",
	}
	updateOrderStatusRequestBytes, err := json.Marshal(updateOrderStatusRequest)
	if err != nil {
		fmt.Printf("Error marshaling request: %v\n", err)
		return
	}
	resp, err = http.Post(
		ts.URL+"/api/user/order/status?user_id="+url.QueryEscape(userID)+"&order_id="+url.QueryEscape(order.ID),
		"application/json",
		bytes.NewBuffer(updateOrderStatusRequestBytes),
	)

	// Display order status
	resp, err = http.Get(
		ts.URL + "/api/user/order?user_id=" + url.QueryEscape(userID) + "&order_id=" + url.QueryEscape(order.ID),
	)
	if err != nil {
		fmt.Printf("Error fetching order status: %v\n", err)
		return
	} else if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error fetching order status: %s\n", resp.Status)
		return
	}
	orderStatus := &state.Order{}
	if err := json.NewDecoder(resp.Body).Decode(orderStatus); err != nil {
		fmt.Printf("Error decoding order status: %v\n", err)
		return
	}
	fmt.Printf("Order: %v\n", orderStatus)

	// Output:
	// Product: Description: Warm mittens, Price: 19.99, Available: 10
	// Product reserved successfully
	// Product after reservation: Description: Warm mittens, Price: 19.99, Available: 7
	// Checkout started successfully: Order ID: 1, Status: created, Items: map[mittens:3]
	// Cart: map[]
	// Order: Order ID: 1, Status: shipped, Items: map[mittens:3]
}
