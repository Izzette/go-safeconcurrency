package ecommerce

import (
	"encoding/json"
	"mime"
	"net/http"
	"sync"

	"github.com/Izzette/go-safeconcurrency/api/types"
	"github.com/Izzette/go-safeconcurrency/eventloop"
	"github.com/Izzette/go-safeconcurrency/examples/ecommerce/events"
	"github.com/Izzette/go-safeconcurrency/examples/ecommerce/state"
)

// EcommerceServer manages inventory and users
// It uses event loops to manage product stock and user actions
// It responds to the following API requests:
//
//   - GET api/product/?product_id → Product # Display product details
//   - POST api/product/reserve?product_id&user_id {quantity: int} # Add or remove items from the cart
//   - GET api/user/cart?user_id → map[string]int # Display the content of the cart
//   - POST api/user/order?user_id → Order # Start the checkout process (create order)
//   - GET api/user/order?user_id&order_id → Order # Display order details
//   - POST api/user/order/status?user_id&order_id {status: string} # Updating order status
type EcommerceServer struct {
	stockLoop types.EventLoop[state.ProductStockState]
	userLoops *sync.Map // userID -> user EventLoop
	mux       *http.ServeMux
}

// NewEcommerceServer initializes the server with the initial stock.
func NewEcommerceServer(initialStock map[string]*state.Product) *EcommerceServer {
	stockState := state.NewProductStockState(initialStock)

	server := &EcommerceServer{
		stockLoop: eventloop.New[state.ProductStockState](stockState),
		userLoops: &sync.Map{},
		mux:       http.NewServeMux(),
	}

	server.mux.HandleFunc("/api/product", server.HandleProduct)
	server.mux.HandleFunc("/api/product/reserve", server.HandleProductReservation)
	server.mux.HandleFunc("/api/user/cart", server.HandleUserCart)
	server.mux.HandleFunc("/api/user/order/status", server.HandleUserOrderStatus)
	server.mux.HandleFunc("/api/user/order", server.HandleUserOrder)

	return server
}

// GetStockLoop returns the event loop for managing product stock
func (s *EcommerceServer) GetUserLoop(userID string) types.EventLoop[state.UserState] {
	newLoop := eventloop.New(state.NewUserState())
	loadedLoop, loaded := s.userLoops.LoadOrStore(userID, newLoop)
	if loaded {
		// Close the new loop if it was not stored in the map.
		// This is to avoid leaking contexts and goroutines created when the loop was initialized.
		newLoop.Close()
	} else {
		// Start the new loop if it was stored in the map
		newLoop.Start()
	}

	// Cast the loaded loop to the correct type.
	// We are the only ones who store in the map and we only store types.EventLoop[state.UserState].
	loop, ok := loadedLoop.(types.EventLoop[state.UserState])
	if !ok {
		panic("user loop is not of type EventLoop")
	}

	return loop
}

// Start starts the server and the stock event loop
// You must only call this method once, otherwise it will panic.
func (s *EcommerceServer) Start() {
	s.stockLoop.Start()
}

// Stop stops the server and all event loops
// It is safe to call this method any number of times.
// You may call this method before or after the server is started.
func (s *EcommerceServer) Stop() {
	s.stockLoop.Close()
	s.userLoops.Range(func(_, value interface{}) bool {
		loop, ok := value.(types.EventLoop[state.UserState])
		if !ok {
			panic("user loop is not of type EventLoop")
		}

		loop.Close()

		// Keep iteration on the map to close all loops
		// Returning false would stop the iteration
		return true
	})
}

// ServeHTTP implements the [http.Handler] interface for the server
func (s *EcommerceServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// HandleProduct handles product-related requests
func (s *EcommerceServer) HandleProduct(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	productID := r.URL.Query().Get("product_id")
	if productID == "" {
		http.Error(w, "Missing product_id", http.StatusBadRequest)
		return
	}

	product := s.stockLoop.Snapshot().State().GetProduct(productID)
	if product == nil {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(product)
}

// HandleProductReservation handles product reservation requests
func (s *EcommerceServer) HandleProductReservation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	productID := r.URL.Query().Get("product_id")
	userID := r.URL.Query().Get("user_id")
	if productID == "" || userID == "" {
		http.Error(w, "Missing product_id or user_id", http.StatusBadRequest)
		return
	}

	contentType := r.Header.Get("Content-Type")
	mediatype, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		http.Error(w, "Invalid Content-Type", http.StatusBadRequest)
		return
	} else if mediatype != "application/json" {
		http.Error(w, "Unsupported Content-Type", http.StatusUnsupportedMediaType)
		return
	}

	var req ReserveProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Send the event to reserve stock
	stockEvent := &events.ReserveStockEvent{
		ProductID: productID,
		Quantity:  req.Quantity,
	}
	gen, err := s.stockLoop.Send(r.Context(), stockEvent)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Wait for the event to be processed, we have to make sure the stock is reserved before we add to the cart
	_, err = eventloop.WaitForGeneration(r.Context(), s.stockLoop, gen)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if stockEvent.Err == events.ErrProductNotFound {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	} else if stockEvent.Err != nil {
		http.Error(w, stockEvent.Err.Error(), http.StatusBadRequest)
		return
	}

	// Send the event to add to the user's cart
	userLoop := s.GetUserLoop(userID)
	addToCartEvent := &events.AddToCartEvent{
		ProductID: productID,
		Quantity:  req.Quantity,
	}
	_, err = userLoop.Send(r.Context(), addToCartEvent)
	// No need to wait for this event to be processed, however it may show an inconsistent state if the user checks the
	// cart before the event is processed.

	w.WriteHeader(http.StatusOK)
}

// HandleUserCart handles user cart-related requests
func (s *EcommerceServer) HandleUserCart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "Missing user_id", http.StatusBadRequest)
		return
	}

	// Get the user's cart
	userLoop := s.GetUserLoop(userID)
	snapshot := userLoop.Snapshot()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(snapshot.State().GetCart())
}

// HandleUserOrderStatus updates the order status
func (s *EcommerceServer) HandleUserOrderStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID := r.URL.Query().Get("user_id")
	orderID := r.URL.Query().Get("order_id")
	if userID == "" || orderID == "" {
		http.Error(w, "Missing order_id", http.StatusBadRequest)
		return
	}

	userLoop := s.GetUserLoop(userID)

	// Update order status
	contentType := r.Header.Get("Content-Type")
	mediatype, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		http.Error(w, "Invalid Content-Type", http.StatusBadRequest)
		return
	} else if mediatype != "application/json" {
		http.Error(w, "Unsupported Content-Type", http.StatusUnsupportedMediaType)
		return
	}
	var req UpdateOrderStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	updateOrderStatusEvent := &events.UpdateOrderStatusEvent{
		OrderID: orderID,
		Status:  req.Status,
	}
	_, err = userLoop.Send(r.Context(), updateOrderStatusEvent)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// No need to wait for this event to be processed, however it may show an inconsistent state if the user checks the
	// order status before the event is processed.

	w.WriteHeader(http.StatusOK)
	return
}

// HandleUserOrder handles user order-related requests
func (s *EcommerceServer) HandleUserOrder(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "Missing user_id ", http.StatusBadRequest)
		return
	}

	userLoop := s.GetUserLoop(userID)
	if r.Method == http.MethodPost {
		// Checkout process
		checkoutEvent := &events.CheckoutEvent{}
		gen, err := userLoop.Send(r.Context(), checkoutEvent)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Wait for the event to be processed
		snap, err := eventloop.WaitForGeneration(r.Context(), userLoop, gen)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		} else if checkoutEvent.Err == events.ErrCartEmpty {
			http.Error(w, "Cart is empty", http.StatusBadRequest)
			return
		} else if checkoutEvent.Err != nil {
			http.Error(w, checkoutEvent.Err.Error(), http.StatusInternalServerError)
			return
		}
		order := snap.State().GetOrder(checkoutEvent.OrderID)

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(order)
		return
	}

	orderID := r.URL.Query().Get("order_id")
	if orderID == "" {
		http.Error(w, "Missing order_id", http.StatusBadRequest)
		return
	}
	if r.Method == http.MethodGet {
		snapshot := userLoop.Snapshot()
		order := snapshot.State().GetOrder(orderID)
		if order == nil {
			http.Error(w, "Order not found", http.StatusNotFound)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(order)
		return
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
}

// ReserveProductRequest represents a request to reserve a product
type ReserveProductRequest struct {
	// Quantity is the number of items to reserve
	Quantity int `json:"quantity"`
}

// UpdateOrderStatusRequest represents a request to update the order status
type UpdateOrderStatusRequest struct {
	// Status is the new status for the order
	Status string `json:"status"`
}
