package test

import (
	. "adcompro-backend"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestOrderPaymentShippingFlow(t *testing.T) {
	if os.Getenv("RUN_AUCTION_INTEGRATION") != "1" {
		t.Skip("set RUN_AUCTION_INTEGRATION=1 to run database integration test")
	}

	ConnectDatabase()
	t.Cleanup(func() { DB.Close() })
	ctx := context.Background()

	var sellerID, sellerUserID int64
	if err := DB.QueryRow(ctx, `
		SELECT sp.seller_id, sp.user_id
		FROM public.seller_profiles sp
		INNER JOIN public.users u ON u.user_id = sp.user_id
		WHERE sp.verification_status = 'VERIFIED' AND u.seller_status = 'VERIFIED'
		ORDER BY sp.seller_id LIMIT 1`,
	).Scan(&sellerID, &sellerUserID); err != nil {
		t.Fatalf("need one verified seller in test database: %v", err)
	}

	suffix := time.Now().UnixNano()
	newUser := func(name, role string) int64 {
		var id int64
		if err := DB.QueryRow(ctx,
			`INSERT INTO public.users (username, full_name, email, password_hash, role) VALUES ($1, $2, $3, 'x', $4) RETURNING user_id`,
			fmt.Sprintf("ord_%s_%d", name, suffix%1000000000), name, fmt.Sprintf("order-it-%s-%d@test.local", name, suffix), role,
		).Scan(&id); err != nil {
			t.Fatalf("create %s: %v", name, err)
		}
		return id
	}
	buyerID := newUser("buyer", "USER")
	adminID := newUser("admin", "ADMIN")

	var gemID int64
	if err := DB.QueryRow(ctx,
		`INSERT INTO public.gems (seller_id, gem_name, gem_type) VALUES ($1, $2, 'Ruby') RETURNING gem_id`,
		sellerID, fmt.Sprintf("Order Integration Gem %d", suffix),
	).Scan(&gemID); err != nil {
		t.Fatal(err)
	}
	newListing := func() int64 {
		var id int64
		if err := DB.QueryRow(ctx, `
			INSERT INTO public.listings (gem_id, seller_id, sale_type, status, fixed_price, published_at)
			VALUES ($1, $2, 'FIXED_PRICE', 'PUBLISHED', 12500, NOW()) RETURNING listing_id`,
			gemID, sellerID,
		).Scan(&id); err != nil {
			t.Fatal(err)
		}
		return id
	}
	listingA, listingB := newListing(), newListing()

	t.Cleanup(func() {
		c := context.Background()
		listings := []int64{listingA, listingB}
		users := []int64{buyerID, adminID}
		_, _ = DB.Exec(c, `DELETE FROM public.notifications WHERE listing_id = ANY($1) OR user_id = ANY($2)`, listings, users)
		_, _ = DB.Exec(c, `DELETE FROM public.transactions WHERE buyer_id = ANY($1)`, users)
		_, _ = DB.Exec(c, `DELETE FROM public.orders WHERE listing_id = ANY($1)`, listings)
		_, _ = DB.Exec(c, `DELETE FROM public.listings WHERE listing_id = ANY($1)`, listings)
		_, _ = DB.Exec(c, `DELETE FROM public.gems WHERE gem_id = $1`, gemID)
		_, _ = DB.Exec(c, `DELETE FROM public.users WHERE user_id = ANY($1)`, users)
	})

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/listings/{id}/buy", AuthMiddleware(BuyNowHandler))
	mux.HandleFunc("GET /api/orders/{id}", AuthMiddleware(GetOrderHandler))
	mux.HandleFunc("POST /api/orders/{id}/pay", AuthMiddleware(PayOrderHandler))
	mux.HandleFunc("POST /api/orders/{id}/cancel", AuthMiddleware(CancelOrderHandler))
	mux.HandleFunc("PUT /api/orders/{id}/ship", AuthMiddleware(VerifiedSellerMiddleware(ShipOrderHandler)))
	mux.HandleFunc("POST /api/orders/{id}/confirm-delivery", AuthMiddleware(ConfirmDeliveryHandler))
	mux.HandleFunc("GET /api/notifications", AuthMiddleware(ListNotificationsHandler))
	mux.HandleFunc("PUT /api/notifications/read-all", AuthMiddleware(MarkAllNotificationsReadHandler))
	mux.HandleFunc("GET /api/subscriptions/me", AuthMiddleware(MySubscriptionHandler))
	mux.HandleFunc("POST /api/subscriptions/upgrade", AuthMiddleware(UpgradeSubscriptionHandler))
	mux.HandleFunc("GET /api/admin/analytics", AuthMiddleware(AdminMiddleware(AdminAnalyticsHandler)))
	server := httptest.NewServer(mux)
	defer server.Close()

	seller := integrationToken(t, sellerUserID)
	buyer := integrationToken(t, buyerID)
	admin := integrationToken(t, adminID)

	call := func(method, path, token string, body any, want int) []byte {
		t.Helper()
		resp := integrationRequest(t, method, server.URL+path, token, body)
		if resp.Status != want {
			t.Fatalf("%s %s: expected %d, got %d: %s", method, path, want, resp.Status, resp.Body)
		}
		return resp.Body
	}
	orderStatus := func(body []byte) string {
		t.Helper()
		var d OrderDetailResponse
		if err := json.Unmarshal(body, &d); err != nil {
			t.Fatal(err)
		}
		return d.Order.Status
	}

	call("POST", fmt.Sprintf("/api/listings/%d/buy", listingA), seller, nil, http.StatusForbidden)
	var order OrderResponse
	if err := json.Unmarshal(call("POST", fmt.Sprintf("/api/listings/%d/buy", listingA), buyer, nil, http.StatusCreated), &order); err != nil {
		t.Fatal(err)
	}
	if order.Status != "PENDING_PAYMENT" || order.TotalAmount != 12500 || order.BuyerID != buyerID {
		t.Fatalf("unexpected order: %+v", order)
	}
	call("POST", fmt.Sprintf("/api/listings/%d/buy", listingA), buyer, nil, http.StatusConflict)
	orderPath := fmt.Sprintf("/api/orders/%d", order.OrderID)

	pay := map[string]any{"payment_method": "MOCK_CARD", "shipping_address": "1 Test Rd, Bangkok"}
	call("POST", orderPath+"/pay", seller, pay, http.StatusForbidden)
	call("POST", orderPath+"/pay", buyer, map[string]any{"payment_method": "CASH", "shipping_address": "x"}, http.StatusBadRequest)
	paid := call("POST", orderPath+"/pay", buyer, pay, http.StatusOK)
	var detail OrderDetailResponse
	if err := json.Unmarshal(paid, &detail); err != nil {
		t.Fatal(err)
	}
	if detail.Order.Status != "PAID" || detail.Shipping == nil || detail.Shipping.Status != "PENDING" ||
		len(detail.Transactions) != 1 || detail.Transactions[0].Status != "SUCCESS" {
		t.Fatalf("unexpected paid order: %s", paid)
	}
	call("POST", orderPath+"/pay", buyer, pay, http.StatusConflict)
	call("POST", orderPath+"/cancel", buyer, nil, http.StatusConflict)

	ship := map[string]any{"courier": "Thailand Post", "tracking_number": "EF123456789TH"}
	call("PUT", orderPath+"/ship", buyer, ship, http.StatusForbidden)
	call("POST", orderPath+"/confirm-delivery", buyer, nil, http.StatusConflict)
	if s := orderStatus(call("PUT", orderPath+"/ship", seller, ship, http.StatusOK)); s != "SHIPPED" {
		t.Fatalf("expected SHIPPED, got %s", s)
	}
	if s := orderStatus(call("POST", orderPath+"/confirm-delivery", buyer, nil, http.StatusOK)); s != "DELIVERED" {
		t.Fatalf("expected DELIVERED, got %s", s)
	}
	call("GET", orderPath, admin, nil, http.StatusOK)

	var orderB OrderResponse
	if err := json.Unmarshal(call("POST", fmt.Sprintf("/api/listings/%d/buy", listingB), buyer, nil, http.StatusCreated), &orderB); err != nil {
		t.Fatal(err)
	}
	if s := orderStatus(call("POST", fmt.Sprintf("/api/orders/%d/cancel", orderB.OrderID), buyer, nil, http.StatusOK)); s != "CANCELLED" {
		t.Fatalf("expected CANCELLED, got %s", s)
	}
	var listingStatus string
	_ = DB.QueryRow(ctx, `SELECT status FROM public.listings WHERE listing_id = $1`, listingB).Scan(&listingStatus)
	if listingStatus != "PUBLISHED" {
		t.Fatalf("cancelled order should republish listing, got %s", listingStatus)
	}

	var inbox struct {
		Notifications []NotificationResponse `json:"notifications"`
		UnreadCount   int                    `json:"unread_count"`
	}
	if err := json.Unmarshal(call("GET", "/api/notifications", buyer, nil, http.StatusOK), &inbox); err != nil {
		t.Fatal(err)
	}
	types := map[string]bool{}
	for _, n := range inbox.Notifications {
		types[n.Type] = true
	}
	if !types["PAYMENT_SUCCESS"] || !types["ORDER_SHIPPED"] || inbox.UnreadCount == 0 {
		t.Fatalf("unexpected buyer notifications: %+v", inbox)
	}
	var queued int
	_ = DB.QueryRow(ctx, `
		SELECT count(*) FROM public.email_deliveries ed
		INNER JOIN public.notifications n ON n.notification_id = ed.notification_id
		WHERE n.user_id = $1`, buyerID).Scan(&queued)
	if queued != len(inbox.Notifications) {
		t.Fatalf("expected one queued email per notification, got %d for %d", queued, len(inbox.Notifications))
	}
	call("PUT", "/api/notifications/read-all", buyer, nil, http.StatusOK)

	upgrade := map[string]any{"payment_method": "MOCK_QR"}
	var first, second struct {
		Subscription SubscriptionResponse `json:"subscription"`
	}
	_ = json.Unmarshal(call("POST", "/api/subscriptions/upgrade", buyer, upgrade, http.StatusOK), &first)
	_ = json.Unmarshal(call("POST", "/api/subscriptions/upgrade", buyer, upgrade, http.StatusOK), &second)
	if first.Subscription.SubscriptionID != second.Subscription.SubscriptionID ||
		second.Subscription.EndDate.Sub(*first.Subscription.EndDate) != 30*24*time.Hour {
		t.Fatalf("second upgrade should extend the same subscription: %+v / %+v", first, second)
	}
	var me struct {
		Plan string `json:"plan"`
	}
	_ = json.Unmarshal(call("GET", "/api/subscriptions/me", buyer, nil, http.StatusOK), &me)
	if me.Plan != "PRO" {
		t.Fatalf("expected PRO, got %s", me.Plan)
	}

	call("GET", "/api/admin/analytics", buyer, nil, http.StatusForbidden)
	var report map[string]any
	if err := json.Unmarshal(call("GET", "/api/admin/analytics", admin, nil, http.StatusOK), &report); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"users", "listings_by_status", "orders_by_status", "sales", "subscriptions", "daily_sales_last_30_days"} {
		if _, ok := report[key]; !ok {
			t.Fatalf("analytics missing %q: %v", key, report)
		}
	}

	t.Logf("PASS order=%d delivered, order=%d cancelled, buyer=%d PRO", order.OrderID, orderB.OrderID, buyerID)
}
