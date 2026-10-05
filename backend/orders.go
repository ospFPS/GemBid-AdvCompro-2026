package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Order lifecycle:
//
//	Buy Now / Auction win -> PENDING_PAYMENT -> PAID -> SHIPPED -> DELIVERED
//	PENDING_PAYMENT -> CANCELLED (buyer cancels, listing is released)

type OrderResponse struct {
	OrderID      int64     `json:"order_id"`
	ListingID    int64     `json:"listing_id"`
	BuyerID      int64     `json:"buyer_id"`
	SellerID     int64     `json:"seller_id"`
	AuctionID    *int64    `json:"auction_id"`
	WinningBidID *int64    `json:"winning_bid_id"`
	ItemPrice    float64   `json:"item_price"`
	ShippingFee  float64   `json:"shipping_fee"`
	TotalAmount  float64   `json:"total_amount"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type ShippingResponse struct {
	ShippingID      int64      `json:"shipping_id"`
	Courier         *string    `json:"courier"`
	TrackingNumber  *string    `json:"tracking_number"`
	ShippingAddress string     `json:"shipping_address"`
	Status          string     `json:"status"`
	ShippedAt       *time.Time `json:"shipped_at"`
	DeliveredAt     *time.Time `json:"delivered_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type TransactionResponse struct {
	TransactionID   int64      `json:"transaction_id"`
	OrderID         *int64     `json:"order_id"`
	SubscriptionID  *int64     `json:"subscription_id"`
	Amount          float64    `json:"amount"`
	PaymentMethod   *string    `json:"payment_method"`
	Status          string     `json:"status"`
	ReferenceNumber *string    `json:"reference_number"`
	CreatedAt       time.Time  `json:"created_at"`
	PaidAt          *time.Time `json:"paid_at"`
}

type OrderDetailResponse struct {
	Order        OrderResponse         `json:"order"`
	Shipping     *ShippingResponse     `json:"shipping"`
	Transactions []TransactionResponse `json:"transactions"`
}

type PayOrderRequest struct {
	PaymentMethod   string `json:"payment_method" example:"MOCK_CARD"`
	ShippingAddress string `json:"shipping_address" example:"123 Sukhumvit Rd, Bangkok 10110"`
}

type ShipOrderRequest struct {
	Courier        string `json:"courier" example:"Kerry Express"`
	TrackingNumber string `json:"tracking_number" example:"KEX123456789TH"`
}

const orderColumns = `
	o.order_id, o.listing_id, o.buyer_id, o.seller_id, o.auction_id, o.winning_bid_id,
	o.item_price, o.shipping_fee, o.total_amount, o.status, o.created_at, o.updated_at`

func scanOrder(row pgx.Row) (OrderResponse, error) {
	var o OrderResponse
	err := row.Scan(&o.OrderID, &o.ListingID, &o.BuyerID, &o.SellerID, &o.AuctionID,
		&o.WinningBidID, &o.ItemPrice, &o.ShippingFee, &o.TotalAmount, &o.Status,
		&o.CreatedAt, &o.UpdatedAt)
	return o, err
}

var validPaymentMethods = map[string]bool{"MOCK_CARD": true, "MOCK_QR": true}

// mockPayment simulates a successful payment gateway call.
func mockPayment() string {
	return "MOCK-" + rand.Text()[:16]
}

// createOrder inserts a PENDING_PAYMENT order. auctionID / winningBidID may be 0.
func createOrder(
	ctx context.Context,
	tx pgx.Tx,
	listingID, buyerID, sellerID, auctionID, winningBidID int64,
	price float64,
) (int64, error) {
	var orderID int64
	err := tx.QueryRow(
		ctx,
		`
		INSERT INTO public.orders (
			listing_id, buyer_id, seller_id, auction_id, winning_bid_id,
			item_price, shipping_fee, total_amount, status
		)
		VALUES ($1, $2, $3, NULLIF($4::bigint, 0), NULLIF($5::bigint, 0), $6, 0, $6, 'PENDING_PAYMENT')
		RETURNING order_id
		`,
		listingID, buyerID, sellerID, auctionID, winningBidID, price,
	).Scan(&orderID)
	return orderID, err
}

// ============================================================
// POST /api/listings/{id}/buy
// ============================================================

// BuyNowHandler godoc
// @Summary Buy Now
// @Description Buy a PUBLISHED fixed-price listing. Creates a PENDING_PAYMENT order and reserves the listing (SOLD).
// @Tags Orders
// @Produce json
// @Security BearerAuth
// @Param id path int true "Listing ID"
// @Success 201 {object} OrderResponse
// @Failure 400 {string} string
// @Failure 401 {string} string
// @Failure 403 {string} string
// @Failure 404 {string} string
// @Failure 409 {string} string
// @Router /api/listings/{id}/buy [post]
func BuyNowHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(userIDKey).(int64)
	listingID, ok := pathID(r)
	if !ok {
		http.Error(w, "Invalid listing ID", http.StatusBadRequest)
		return
	}
	ctx := r.Context()

	tx, err := DB.Begin(ctx)
	if err != nil {
		http.Error(w, "Could not start order transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(ctx)

	var sellerID, sellerUserID int64
	var saleType, status string
	var price *float64
	err = tx.QueryRow(
		ctx,
		`
		SELECT l.seller_id, sp.user_id, COALESCE(l.sale_type, ''), l.status, l.fixed_price
		FROM public.listings l
		INNER JOIN public.seller_profiles sp ON sp.seller_id = l.seller_id
		WHERE l.listing_id = $1
		FOR UPDATE OF l
		`,
		listingID,
	).Scan(&sellerID, &sellerUserID, &saleType, &status, &price)
	if err == pgx.ErrNoRows {
		http.Error(w, "Listing not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Could not load listing", http.StatusInternalServerError)
		return
	}
	if saleType != "FIXED_PRICE" || status != "PUBLISHED" || price == nil {
		http.Error(w, "Listing is not available for Buy Now", http.StatusConflict)
		return
	}
	if userID == sellerUserID {
		http.Error(w, "You cannot buy your own listing", http.StatusForbidden)
		return
	}

	if _, err := tx.Exec(ctx,
		`UPDATE public.listings SET status = 'SOLD' WHERE listing_id = $1`, listingID); err != nil {
		http.Error(w, "Could not reserve listing", http.StatusInternalServerError)
		return
	}

	orderID, err := createOrder(ctx, tx, listingID, userID, sellerID, 0, 0, *price)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			http.Error(w, "Listing already has an order", http.StatusConflict)
			return
		}
		http.Error(w, "Could not create order", http.StatusInternalServerError)
		return
	}

	if err := notify(ctx, tx, sellerUserID, "ORDER_CREATED", "New order received",
		"A buyer placed a Buy Now order on your listing. Waiting for payment.",
		0, listingID, orderID); err != nil {
		http.Error(w, "Could not create notification", http.StatusInternalServerError)
		return
	}

	order, err := scanOrder(tx.QueryRow(ctx,
		`SELECT `+orderColumns+` FROM public.orders o WHERE o.order_id = $1`, orderID))
	if err != nil {
		http.Error(w, "Could not load order", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(ctx); err != nil {
		http.Error(w, "Could not commit order", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, order)
}

// ============================================================
// GET /api/orders, GET /api/orders/{id}
// ============================================================

// ListOrdersHandler godoc
// @Summary List my orders
// @Description Orders I bought. Use ?as=seller for orders on my shop.
// @Tags Orders
// @Produce json
// @Security BearerAuth
// @Param as query string false "buyer (default) or seller"
// @Success 200 {array} OrderResponse
// @Failure 401 {string} string
// @Router /api/orders [get]
func ListOrdersHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(userIDKey).(int64)

	where := `o.buyer_id = $1`
	if r.URL.Query().Get("as") == "seller" {
		where = `o.seller_id = (SELECT seller_id FROM public.seller_profiles WHERE user_id = $1)`
	}

	rows, err := DB.Query(r.Context(),
		`SELECT `+orderColumns+` FROM public.orders o WHERE `+where+` ORDER BY o.created_at DESC`,
		userID)
	if err != nil {
		http.Error(w, "Could not load orders", http.StatusInternalServerError)
		return
	}
	orders, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (OrderResponse, error) {
		return scanOrder(row)
	})
	if err != nil {
		http.Error(w, "Could not read orders", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, orders)
}

// GetOrderHandler godoc
// @Summary Get order detail
// @Description Order with shipping and payment transactions. Visible to the buyer, the seller and admins.
// @Tags Orders
// @Produce json
// @Security BearerAuth
// @Param id path int true "Order ID"
// @Success 200 {object} OrderDetailResponse
// @Failure 401 {string} string
// @Failure 404 {string} string
// @Router /api/orders/{id} [get]
func GetOrderHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(userIDKey).(int64)
	orderID, ok := pathID(r)
	if !ok {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	// Buyer, seller or admin only; anyone else gets 404.
	order, err := scanOrder(DB.QueryRow(
		r.Context(),
		`
		SELECT `+orderColumns+`
		FROM public.orders o
		INNER JOIN public.seller_profiles sp ON sp.seller_id = o.seller_id
		INNER JOIN public.users me ON me.user_id = $2
		WHERE o.order_id = $1
		  AND (o.buyer_id = me.user_id OR sp.user_id = me.user_id OR me.role = 'ADMIN')
		`,
		orderID, userID,
	))
	if err == pgx.ErrNoRows {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Could not load order", http.StatusInternalServerError)
		return
	}

	detail, err := loadOrderDetail(r.Context(), order)
	if err != nil {
		http.Error(w, "Could not load order detail", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, detail)
}

func loadOrderDetail(ctx context.Context, order OrderResponse) (OrderDetailResponse, error) {
	detail := OrderDetailResponse{Order: order}

	var s ShippingResponse
	err := DB.QueryRow(
		ctx,
		`
		SELECT shipping_id, courier, tracking_number, shipping_address, status,
			shipped_at, delivered_at, updated_at
		FROM public.shipping WHERE order_id = $1
		`,
		order.OrderID,
	).Scan(&s.ShippingID, &s.Courier, &s.TrackingNumber, &s.ShippingAddress, &s.Status,
		&s.ShippedAt, &s.DeliveredAt, &s.UpdatedAt)
	if err == nil {
		detail.Shipping = &s
	} else if err != pgx.ErrNoRows {
		return detail, err
	}

	rows, err := DB.Query(
		ctx,
		`
		SELECT transaction_id, order_id, subscription_id, amount, payment_method, status,
			reference_number, created_at, paid_at
		FROM public.transactions WHERE order_id = $1 ORDER BY transaction_id
		`,
		order.OrderID,
	)
	if err != nil {
		return detail, err
	}
	detail.Transactions, err = pgx.CollectRows(rows, scanTransaction)
	return detail, err
}

func scanTransaction(row pgx.CollectableRow) (TransactionResponse, error) {
	var t TransactionResponse
	err := row.Scan(&t.TransactionID, &t.OrderID, &t.SubscriptionID, &t.Amount,
		&t.PaymentMethod, &t.Status, &t.ReferenceNumber, &t.CreatedAt, &t.PaidAt)
	return t, err
}

// ============================================================
// ORDER STATE TRANSITIONS
// ============================================================

type lockedOrder struct {
	OrderResponse
	SellerUserID int64
}

type httpErr struct {
	status int
	msg    string
}

// transitionOrder locks the order row, runs step inside the transaction,
// commits, and responds with the updated order detail.
func transitionOrder(
	w http.ResponseWriter,
	r *http.Request,
	step func(ctx context.Context, tx pgx.Tx, o lockedOrder) *httpErr,
) {
	orderID, ok := pathID(r)
	if !ok {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}
	ctx := r.Context()

	tx, err := DB.Begin(ctx)
	if err != nil {
		http.Error(w, "Could not start order transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(ctx)

	var o lockedOrder
	err = tx.QueryRow(
		ctx,
		`
		SELECT `+orderColumns+`, sp.user_id
		FROM public.orders o
		INNER JOIN public.seller_profiles sp ON sp.seller_id = o.seller_id
		WHERE o.order_id = $1
		FOR UPDATE OF o
		`,
		orderID,
	).Scan(&o.OrderID, &o.ListingID, &o.BuyerID, &o.SellerID, &o.AuctionID,
		&o.WinningBidID, &o.ItemPrice, &o.ShippingFee, &o.TotalAmount, &o.Status,
		&o.CreatedAt, &o.UpdatedAt, &o.SellerUserID)
	if err == pgx.ErrNoRows {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Could not load order", http.StatusInternalServerError)
		return
	}

	if e := step(ctx, tx, o); e != nil {
		http.Error(w, e.msg, e.status)
		return
	}
	if err := tx.Commit(ctx); err != nil {
		http.Error(w, "Could not commit order update", http.StatusInternalServerError)
		return
	}

	order, err := scanOrder(DB.QueryRow(ctx,
		`SELECT `+orderColumns+` FROM public.orders o WHERE o.order_id = $1`, orderID))
	if err != nil {
		http.Error(w, "Order updated but could not be reloaded", http.StatusInternalServerError)
		return
	}
	detail, err := loadOrderDetail(ctx, order)
	if err != nil {
		http.Error(w, "Order updated but could not be reloaded", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, detail)
}

func dbFail(msg string) *httpErr {
	return &httpErr{http.StatusInternalServerError, msg}
}

// PayOrderHandler godoc
// @Summary Pay for an order (mock payment)
// @Description Buyer pays a PENDING_PAYMENT order with MOCK_CARD or MOCK_QR and gives the shipping address. Order becomes PAID and a shipping record is created.
// @Tags Payments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Order ID"
// @Param request body PayOrderRequest true "Payment method and shipping address"
// @Success 200 {object} OrderDetailResponse
// @Failure 400 {string} string
// @Failure 403 {string} string
// @Failure 404 {string} string
// @Failure 409 {string} string
// @Router /api/orders/{id}/pay [post]
func PayOrderHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(userIDKey).(int64)

	var request PayOrderRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	request.PaymentMethod = strings.ToUpper(strings.TrimSpace(request.PaymentMethod))
	request.ShippingAddress = strings.TrimSpace(request.ShippingAddress)
	if !validPaymentMethods[request.PaymentMethod] {
		http.Error(w, "payment_method must be MOCK_CARD or MOCK_QR", http.StatusBadRequest)
		return
	}
	if request.ShippingAddress == "" {
		http.Error(w, "shipping_address is required", http.StatusBadRequest)
		return
	}

	transitionOrder(w, r, func(ctx context.Context, tx pgx.Tx, o lockedOrder) *httpErr {
		if o.BuyerID != userID {
			return &httpErr{http.StatusForbidden, "Only the buyer can pay for this order"}
		}
		if o.Status != "PENDING_PAYMENT" {
			return &httpErr{http.StatusConflict, "Order is not awaiting payment"}
		}

		if _, err := tx.Exec(
			ctx,
			`
			INSERT INTO public.transactions
				(order_id, buyer_id, amount, payment_method, status, reference_number, paid_at)
			VALUES ($1, $2, $3, $4, 'SUCCESS', $5, NOW())
			`,
			o.OrderID, userID, o.TotalAmount, request.PaymentMethod, mockPayment(),
		); err != nil {
			return dbFail("Could not record payment")
		}
		if _, err := tx.Exec(ctx,
			`UPDATE public.orders SET status = 'PAID' WHERE order_id = $1`, o.OrderID); err != nil {
			return dbFail("Could not update order")
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO public.shipping (order_id, shipping_address) VALUES ($1, $2)`,
			o.OrderID, request.ShippingAddress); err != nil {
			return dbFail("Could not create shipping record")
		}
		if err := notify(ctx, tx, o.SellerUserID, "PAYMENT_RECEIVED", "Payment received",
			"The buyer has paid. Please ship the gem and add the tracking number.",
			0, o.ListingID, o.OrderID); err != nil {
			return dbFail("Could not create notification")
		}
		if err := notify(ctx, tx, o.BuyerID, "PAYMENT_SUCCESS", "Payment successful",
			"Your payment was successful. The seller will ship your gem soon.",
			0, o.ListingID, o.OrderID); err != nil {
			return dbFail("Could not create notification")
		}
		return nil
	})
}

// CancelOrderHandler godoc
// @Summary Cancel an unpaid order
// @Description Buyer cancels a PENDING_PAYMENT order. A fixed-price listing goes back to PUBLISHED; an auction listing becomes UNSOLD.
// @Tags Orders
// @Produce json
// @Security BearerAuth
// @Param id path int true "Order ID"
// @Success 200 {object} OrderDetailResponse
// @Failure 403 {string} string
// @Failure 404 {string} string
// @Failure 409 {string} string
// @Router /api/orders/{id}/cancel [post]
func CancelOrderHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(userIDKey).(int64)

	transitionOrder(w, r, func(ctx context.Context, tx pgx.Tx, o lockedOrder) *httpErr {
		if o.BuyerID != userID {
			return &httpErr{http.StatusForbidden, "Only the buyer can cancel this order"}
		}
		if o.Status != "PENDING_PAYMENT" {
			return &httpErr{http.StatusConflict, "Only unpaid orders can be cancelled"}
		}

		listingStatus := "PUBLISHED"
		if o.AuctionID != nil {
			listingStatus = "UNSOLD"
		}
		if _, err := tx.Exec(ctx,
			`UPDATE public.orders SET status = 'CANCELLED' WHERE order_id = $1`, o.OrderID); err != nil {
			return dbFail("Could not cancel order")
		}
		if _, err := tx.Exec(ctx,
			`UPDATE public.listings SET status = $1 WHERE listing_id = $2`,
			listingStatus, o.ListingID); err != nil {
			return dbFail("Could not release listing")
		}
		if err := notify(ctx, tx, o.SellerUserID, "ORDER_CANCELLED", "Order cancelled",
			"The buyer cancelled the unpaid order. Your listing is now "+listingStatus+".",
			0, o.ListingID, o.OrderID); err != nil {
			return dbFail("Could not create notification")
		}
		return nil
	})
}

// ShipOrderHandler godoc
// @Summary Mark an order as shipped
// @Description Seller adds courier and tracking number to a PAID order.
// @Tags Shipping
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Order ID"
// @Param request body ShipOrderRequest true "Courier and tracking number"
// @Success 200 {object} OrderDetailResponse
// @Failure 400 {string} string
// @Failure 403 {string} string
// @Failure 404 {string} string
// @Failure 409 {string} string
// @Router /api/orders/{id}/ship [put]
func ShipOrderHandler(w http.ResponseWriter, r *http.Request) {
	sellerID, _ := r.Context().Value(sellerIDKey).(int64)

	var request ShipOrderRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	request.Courier = strings.TrimSpace(request.Courier)
	request.TrackingNumber = strings.TrimSpace(request.TrackingNumber)
	if request.Courier == "" || request.TrackingNumber == "" {
		http.Error(w, "courier and tracking_number are required", http.StatusBadRequest)
		return
	}

	transitionOrder(w, r, func(ctx context.Context, tx pgx.Tx, o lockedOrder) *httpErr {
		if o.SellerID != sellerID {
			return &httpErr{http.StatusForbidden, "You do not own this order"}
		}
		if o.Status != "PAID" {
			return &httpErr{http.StatusConflict, "Only PAID orders can be shipped"}
		}

		if _, err := tx.Exec(
			ctx,
			`
			UPDATE public.shipping
			SET courier = $1, tracking_number = $2, status = 'SHIPPED', shipped_at = NOW()
			WHERE order_id = $3
			`,
			request.Courier, request.TrackingNumber, o.OrderID,
		); err != nil {
			return dbFail("Could not update shipping")
		}
		if _, err := tx.Exec(ctx,
			`UPDATE public.orders SET status = 'SHIPPED' WHERE order_id = $1`, o.OrderID); err != nil {
			return dbFail("Could not update order")
		}
		if err := notify(ctx, tx, o.BuyerID, "ORDER_SHIPPED", "Your gem has been shipped",
			"Courier: "+request.Courier+", tracking number: "+request.TrackingNumber+".",
			0, o.ListingID, o.OrderID); err != nil {
			return dbFail("Could not create notification")
		}
		return nil
	})
}

// ConfirmDeliveryHandler godoc
// @Summary Confirm delivery
// @Description Buyer confirms a SHIPPED order has arrived. Order and shipping become DELIVERED.
// @Tags Shipping
// @Produce json
// @Security BearerAuth
// @Param id path int true "Order ID"
// @Success 200 {object} OrderDetailResponse
// @Failure 403 {string} string
// @Failure 404 {string} string
// @Failure 409 {string} string
// @Router /api/orders/{id}/confirm-delivery [post]
func ConfirmDeliveryHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(userIDKey).(int64)

	transitionOrder(w, r, func(ctx context.Context, tx pgx.Tx, o lockedOrder) *httpErr {
		if o.BuyerID != userID {
			return &httpErr{http.StatusForbidden, "Only the buyer can confirm delivery"}
		}
		if o.Status != "SHIPPED" {
			return &httpErr{http.StatusConflict, "Only SHIPPED orders can be confirmed"}
		}

		if _, err := tx.Exec(ctx,
			`UPDATE public.shipping SET status = 'DELIVERED', delivered_at = NOW() WHERE order_id = $1`,
			o.OrderID); err != nil {
			return dbFail("Could not update shipping")
		}
		if _, err := tx.Exec(ctx,
			`UPDATE public.orders SET status = 'DELIVERED' WHERE order_id = $1`, o.OrderID); err != nil {
			return dbFail("Could not update order")
		}
		if err := notify(ctx, tx, o.SellerUserID, "ORDER_DELIVERED", "Order delivered",
			"The buyer confirmed the gem was delivered. The order is complete.",
			0, o.ListingID, o.OrderID); err != nil {
			return dbFail("Could not create notification")
		}
		return nil
	})
}
