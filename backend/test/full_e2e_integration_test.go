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

func TestFullAPIEndToEnd(t *testing.T) {
	if os.Getenv("RUN_AUCTION_INTEGRATION") != "1" {
		t.Skip("set RUN_AUCTION_INTEGRATION=1 to run database integration test")
	}

	ConnectDatabase()
	t.Cleanup(func() { DB.Close() })
	ctx := context.Background()

	var adminID int64
	if err := DB.QueryRow(ctx,
		`SELECT user_id FROM public.users WHERE role = 'ADMIN' ORDER BY user_id LIMIT 1`,
	).Scan(&adminID); err != nil {
		t.Fatalf("need one admin in test database: %v", err)
	}
	adminToken := integrationToken(t, adminID)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/register", RegisterHandler)
	mux.HandleFunc("POST /api/login", LoginHandler)
	mux.HandleFunc("GET /api/me", AuthMiddleware(MeHandler))
	mux.HandleFunc("PUT /api/me", AuthMiddleware(UpdateProfileHandler))
	mux.HandleFunc("PUT /api/me/password", AuthMiddleware(ChangePasswordHandler))
	mux.HandleFunc("POST /api/seller/apply", AuthMiddleware(SellerApplyHandler))
	mux.HandleFunc("GET /api/seller/status", AuthMiddleware(SellerStatusHandler))
	mux.HandleFunc("PUT /api/admin/seller-applications/{id}", AuthMiddleware(AdminMiddleware(AdminReviewSellerHandler)))
	mux.HandleFunc("POST /api/gems", AuthMiddleware(VerifiedSellerMiddleware(CreateGemHandler)))
	mux.HandleFunc("GET /api/gems/{id}", GetGemHandler)
	mux.HandleFunc("POST /api/listings", AuthMiddleware(VerifiedSellerMiddleware(CreateListingHandler)))
	mux.HandleFunc("GET /api/listings/{id}", GetListingHandler)
	mux.HandleFunc("PUT /api/listings/{id}/publish", AuthMiddleware(VerifiedSellerMiddleware(PublishListingHandler)))
	mux.HandleFunc("POST /api/listings/{id}/buy", AuthMiddleware(BuyNowHandler))
	mux.HandleFunc("GET /api/orders", AuthMiddleware(ListOrdersHandler))
	mux.HandleFunc("GET /api/orders/{id}", AuthMiddleware(GetOrderHandler))
	mux.HandleFunc("POST /api/orders/{id}/pay", AuthMiddleware(PayOrderHandler))
	mux.HandleFunc("PUT /api/orders/{id}/ship", AuthMiddleware(VerifiedSellerMiddleware(ShipOrderHandler)))
	mux.HandleFunc("POST /api/orders/{id}/confirm-delivery", AuthMiddleware(ConfirmDeliveryHandler))
	mux.HandleFunc("GET /api/notifications", AuthMiddleware(ListNotificationsHandler))
	mux.HandleFunc("GET /api/subscriptions/me", AuthMiddleware(MySubscriptionHandler))
	mux.HandleFunc("POST /api/subscriptions/upgrade", AuthMiddleware(UpgradeSubscriptionHandler))
	mux.HandleFunc("GET /api/admin/analytics", AuthMiddleware(AdminMiddleware(AdminAnalyticsHandler)))

	server := httptest.NewServer(mux)
	defer server.Close()

	call := func(method, path, token string, body any, want int) []byte {
		t.Helper()
		resp := integrationRequest(t, method, server.URL+path, token, body)
		requireHTTPStatus(t, method+" "+path, resp, want)
		return resp.Body
	}
	login := func(username, password string, want int) (string, UserResponse) {
		t.Helper()
		body := call(http.MethodPost, "/api/login", "", map[string]any{
			"username": username, "password": password,
		}, want)
		if want != http.StatusOK {
			return "", UserResponse{}
		}
		var response struct {
			Token string       `json:"token"`
			User  UserResponse `json:"user"`
		}
		if err := json.Unmarshal(body, &response); err != nil {
			t.Fatal(err)
		}
		if response.Token == "" {
			t.Fatal("login returned empty token")
		}
		return response.Token, response.User
	}
	register := func(username, name, email, password string) UserResponse {
		t.Helper()
		body := call(http.MethodPost, "/api/register", "", map[string]any{
			"username":  username,
			"full_name": name,
			"email":     email,
			"phone":     "0812345678",
			"password":  password,
		}, http.StatusCreated)
		var response struct {
			User UserResponse `json:"user"`
		}
		if err := json.Unmarshal(body, &response); err != nil {
			t.Fatal(err)
		}
		return response.User
	}

	suffix := time.Now().UnixNano()
	sellerUsername := fmt.Sprintf("seller%d", suffix%1000000000)
	buyerUsername := fmt.Sprintf("buyer%d", suffix%1000000000)
	sellerEmail := fmt.Sprintf("e2e-seller-%d@test.local", suffix)
	buyerEmail := fmt.Sprintf("e2e-buyer-%d@test.local", suffix)
	password := "StartPass123"
	newPassword := "ChangedPass123"

	sellerUser := register(sellerUsername, "E2E Seller", sellerEmail, password)
	buyerUser := register(buyerUsername, "E2E Buyer", buyerEmail, password)

	var sellerID, gemID, listingID, orderID int64
	t.Cleanup(func() {
		c := context.Background()
		users := []int64{sellerUser.UserID, buyerUser.UserID}

		_, _ = DB.Exec(c, `DELETE FROM public.email_deliveries
			WHERE notification_id IN (SELECT notification_id FROM public.notifications WHERE user_id = ANY($1))`, users)
		_, _ = DB.Exec(c, `DELETE FROM public.notifications WHERE user_id = ANY($1)`, users)
		_, _ = DB.Exec(c, `DELETE FROM public.transactions WHERE buyer_id = ANY($1)`, users)
		if orderID != 0 {
			_, _ = DB.Exec(c, `DELETE FROM public.orders WHERE order_id = $1`, orderID)
		}
		_, _ = DB.Exec(c, `DELETE FROM public.subscriptions WHERE user_id = ANY($1)`, users)
		if listingID != 0 {
			_, _ = DB.Exec(c, `DELETE FROM public.listings WHERE listing_id = $1`, listingID)
		}
		if gemID != 0 {
			_, _ = DB.Exec(c, `DELETE FROM public.gems WHERE gem_id = $1`, gemID)
		}
		if sellerID != 0 {
			_, _ = DB.Exec(c, `DELETE FROM public.seller_profiles WHERE seller_id = $1`, sellerID)
		}
		_, _ = DB.Exec(c, `DELETE FROM public.users WHERE user_id = ANY($1)`, users)
	})

	call(http.MethodPost, "/api/register", "", map[string]any{
		"username": buyerUsername,
		"password": password,
	}, http.StatusConflict)
	login(buyerUsername, "WrongPass123", http.StatusUnauthorized)

	sellerToken, _ := login(sellerUsername, password, http.StatusOK)
	buyerToken, _ := login(buyerUsername, password, http.StatusOK)

	call(http.MethodGet, "/api/me", "", nil, http.StatusUnauthorized)
	call(http.MethodGet, "/api/me", buyerToken, nil, http.StatusOK)
	call(http.MethodPut, "/api/me", buyerToken, map[string]any{
		"full_name": "E2E Buyer Updated",
		"phone":     "0899999999",
	}, http.StatusOK)
	call(http.MethodPut, "/api/me/password", buyerToken, map[string]any{
		"current_password": "wrong",
		"new_password":     newPassword,
	}, http.StatusUnauthorized)
	call(http.MethodPut, "/api/me/password", buyerToken, map[string]any{
		"current_password": password,
		"new_password":     newPassword,
	}, http.StatusOK)
	login(buyerUsername, password, http.StatusUnauthorized)
	buyerToken, _ = login(buyerUsername, newPassword, http.StatusOK)

	applyBody := call(http.MethodPost, "/api/seller/apply", sellerToken, map[string]any{
		"shop_name":   "E2E Gem Shop",
		"seller_type": "INDIVIDUAL",
		"legal_name":  "E2E Seller",
		"address":     "Bangkok",
		"description": "Integration test seller",
	}, http.StatusCreated)
	var application struct {
		Application SellerApplicationResponse `json:"application"`
	}
	if err := json.Unmarshal(applyBody, &application); err != nil {
		t.Fatal(err)
	}
	sellerID = application.Application.SellerID

	call(http.MethodPost, "/api/seller/apply", sellerToken, map[string]any{
		"shop_name": "Again", "legal_name": "Again",
	}, http.StatusConflict)
	call(http.MethodPut, fmt.Sprintf("/api/admin/seller-applications/%d", sellerID), buyerToken,
		map[string]any{"decision": "APPROVE"}, http.StatusForbidden)
	call(http.MethodPut, fmt.Sprintf("/api/admin/seller-applications/%d", sellerID), adminToken,
		map[string]any{"decision": "APPROVE"}, http.StatusOK)
	call(http.MethodGet, "/api/seller/status", sellerToken, nil, http.StatusOK)

	call(http.MethodPost, "/api/gems", buyerToken, map[string]any{
		"gem_name": "Not Allowed", "gem_type": "Ruby",
	}, http.StatusForbidden)

	gemBody := call(http.MethodPost, "/api/gems", sellerToken, map[string]any{
		"gem_name":    "E2E Blue Sapphire",
		"gem_type":    "Sapphire",
		"carat":       1.25,
		"shape":       "Oval",
		"color":       "Blue",
		"description": "End-to-end test gem",
	}, http.StatusCreated)
	var createdGem struct {
		Gem GemResponse `json:"gem"`
	}
	if err := json.Unmarshal(gemBody, &createdGem); err != nil {
		t.Fatal(err)
	}
	gemID = createdGem.Gem.GemID
	call(http.MethodGet, fmt.Sprintf("/api/gems/%d", gemID), "", nil, http.StatusOK)

	listingBody := call(http.MethodPost, "/api/listings", sellerToken,
		map[string]any{"gem_id": gemID}, http.StatusCreated)
	var createdListing struct {
		Listing ListingResponse `json:"listing"`
	}
	if err := json.Unmarshal(listingBody, &createdListing); err != nil {
		t.Fatal(err)
	}
	listingID = createdListing.Listing.ListingID

	call(http.MethodPut, fmt.Sprintf("/api/listings/%d/publish", listingID), sellerToken, map[string]any{
		"sale_type":   "FIXED_PRICE",
		"fixed_price": 2500.0,
	}, http.StatusOK)
	call(http.MethodGet, fmt.Sprintf("/api/listings/%d", listingID), "", nil, http.StatusOK)

	call(http.MethodPost, fmt.Sprintf("/api/listings/%d/buy", listingID), sellerToken, nil, http.StatusForbidden)
	orderBody := call(http.MethodPost, fmt.Sprintf("/api/listings/%d/buy", listingID), buyerToken, nil, http.StatusCreated)
	var order OrderResponse
	if err := json.Unmarshal(orderBody, &order); err != nil {
		t.Fatal(err)
	}
	orderID = order.OrderID

	call(http.MethodGet, "/api/orders", buyerToken, nil, http.StatusOK)
	call(http.MethodGet, "/api/orders?as=seller", sellerToken, nil, http.StatusOK)
	call(http.MethodPost, fmt.Sprintf("/api/orders/%d/pay", orderID), buyerToken, map[string]any{
		"payment_method":   "MOCK_CARD",
		"shipping_address": "230 Test Road, Bangkok",
	}, http.StatusOK)
	call(http.MethodPut, fmt.Sprintf("/api/orders/%d/ship", orderID), sellerToken, map[string]any{
		"courier":         "Thailand Post",
		"tracking_number": "TEST123456TH",
	}, http.StatusOK)
	call(http.MethodPost, fmt.Sprintf("/api/orders/%d/confirm-delivery", orderID), buyerToken, nil, http.StatusOK)

	call(http.MethodGet, "/api/notifications", buyerToken, nil, http.StatusOK)
	call(http.MethodGet, "/api/notifications", sellerToken, nil, http.StatusOK)

	var subscription struct {
		Plan string `json:"plan"`
	}
	if err := json.Unmarshal(call(http.MethodGet, "/api/subscriptions/me", buyerToken, nil, http.StatusOK), &subscription); err != nil {
		t.Fatal(err)
	}
	if subscription.Plan != "FREE" {
		t.Fatalf("expected FREE before upgrade, got %s", subscription.Plan)
	}
	call(http.MethodPost, "/api/subscriptions/upgrade", buyerToken,
		map[string]any{"payment_method": "MOCK_QR"}, http.StatusOK)
	if err := json.Unmarshal(call(http.MethodGet, "/api/subscriptions/me", buyerToken, nil, http.StatusOK), &subscription); err != nil {
		t.Fatal(err)
	}
	if subscription.Plan != "PRO" {
		t.Fatalf("expected PRO after upgrade, got %s", subscription.Plan)
	}

	call(http.MethodGet, "/api/admin/analytics", buyerToken, nil, http.StatusForbidden)
	call(http.MethodGet, "/api/admin/analytics", adminToken, nil, http.StatusOK)

	var finalStatus string
	if err := DB.QueryRow(ctx, `SELECT status FROM public.orders WHERE order_id = $1`, orderID).Scan(&finalStatus); err != nil {
		t.Fatal(err)
	}
	if finalStatus != "DELIVERED" {
		t.Fatalf("expected final order status DELIVERED, got %s", finalStatus)
	}
}
