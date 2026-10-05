package test

import (
	. "adcompro-backend"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type httpTestResponse struct {
	Status int
	Body   []byte
}

func requireHTTPStatus(t *testing.T, action string, resp httpTestResponse, want int) {
	t.Helper()
	if resp.Status != want {
		t.Fatalf("%s: expected %d, got %d: %s", action, want, resp.Status, resp.Body)
	}
}

func TestAuctionEngineTwoBidderFlow(t *testing.T) {
	if os.Getenv("RUN_AUCTION_INTEGRATION") != "1" {
		t.Skip("set RUN_AUCTION_INTEGRATION=1 to run Supabase integration test")
	}

	ConnectDatabase()
	t.Cleanup(func() { DB.Close() })

	ctx := context.Background()

	var sellerID, sellerUserID int64
	err := DB.QueryRow(
		ctx,
		`
		SELECT sp.seller_id, sp.user_id
		FROM public.seller_profiles sp
		INNER JOIN public.users u ON u.user_id = sp.user_id
		WHERE sp.verification_status = 'VERIFIED'
		  AND u.seller_status = 'VERIFIED'
		ORDER BY sp.seller_id
		LIMIT 1
		`,
	).Scan(&sellerID, &sellerUserID)
	if err != nil {
		t.Fatalf("need one verified seller in test database: %v", err)
	}

	suffix := time.Now().UnixNano()
	bidderIDs := make([]int64, 2)
	for i, label := range []string{"A", "B"} {
		err = DB.QueryRow(
			ctx,
			`
			INSERT INTO public.users (username, full_name, email, password_hash)
			VALUES ($1, $2, $3, $4)
			RETURNING user_id
			`,
			fmt.Sprintf("ait%d_%d", i, suffix%1000000000),
			"Auction Integration Bidder "+label,
			fmt.Sprintf("auction-it-%d-%d@test.local", i, suffix),
			"integration-only",
		).Scan(&bidderIDs[i])
		if err != nil {
			t.Fatalf("create bidder %s: %v", label, err)
		}
	}
	bidderAID, bidderBID := bidderIDs[0], bidderIDs[1]

	var gemID int64
	err = DB.QueryRow(
		ctx,
		`
		INSERT INTO public.gems (
			seller_id, gem_name, gem_type, carat, shape, color, description
		)
		VALUES ($1, $2, 'Sapphire', 1.25, 'Oval', 'Blue', $3)
		RETURNING gem_id
		`,
		sellerID,
		fmt.Sprintf("Auction Integration Gem %d", suffix),
		"Temporary record created by automated auction integration test",
	).Scan(&gemID)
	if err != nil {
		t.Fatalf("create test gem: %v", err)
	}

	var listingID int64
	err = DB.QueryRow(
		ctx,
		`
		INSERT INTO public.listings (gem_id, seller_id, sale_type, status)
		VALUES ($1, $2, 'AUCTION', 'DRAFT')
		RETURNING listing_id
		`,
		gemID,
		sellerID,
	).Scan(&listingID)
	if err != nil {
		t.Fatalf("create test listing: %v", err)
	}

	var auctionID int64

	t.Cleanup(func() {
		cleanupCtx := context.Background()

		_, _ = DB.Exec(cleanupCtx, `DELETE FROM public.notifications WHERE auction_id = $1 OR listing_id = $2 OR user_id = ANY($3)`,
			auctionID, listingID, []int64{bidderAID, bidderBID})
		_, _ = DB.Exec(cleanupCtx, `DELETE FROM public.transactions WHERE order_id IN (SELECT order_id FROM public.orders WHERE listing_id = $1)`,
			listingID)
		_, _ = DB.Exec(cleanupCtx, `DELETE FROM public.orders WHERE listing_id = $1`, listingID)
		if auctionID != 0 {
			_, _ = DB.Exec(cleanupCtx, `DELETE FROM public.bids WHERE auction_id = $1`, auctionID)
			_, _ = DB.Exec(cleanupCtx, `DELETE FROM public.auctions WHERE auction_id = $1`, auctionID)
		}
		_, _ = DB.Exec(cleanupCtx, `DELETE FROM public.listings WHERE listing_id = $1`, listingID)
		_, _ = DB.Exec(cleanupCtx, `DELETE FROM public.gems WHERE gem_id = $1`, gemID)
		_, _ = DB.Exec(cleanupCtx, `DELETE FROM public.users WHERE user_id = ANY($1)`, []int64{bidderAID, bidderBID})
	})

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/auctions", AuthMiddleware(VerifiedSellerMiddleware(CreateAuctionHandler)))
	mux.HandleFunc("GET /api/auctions", ListAuctionsHandler)
	mux.HandleFunc("GET /api/auctions/{id}", GetAuctionHandler)
	mux.HandleFunc("GET /api/auctions/{id}/bids", ListAuctionBidsHandler)
	mux.HandleFunc("POST /api/auctions/{id}/bids", AuthMiddleware(PlaceBidHandler))
	mux.HandleFunc("POST /api/auctions/{id}/close", AuthMiddleware(VerifiedSellerMiddleware(CloseAuctionHandler)))

	server := httptest.NewServer(mux)
	defer server.Close()

	sellerToken := integrationToken(t, sellerUserID)
	bidderAToken := integrationToken(t, bidderAID)
	bidderBToken := integrationToken(t, bidderBID)

	startTime := time.Now().Add(-time.Minute)
	endTime := time.Now().Add(6 * time.Second)

	createResp := integrationRequest(
		t,
		http.MethodPost,
		server.URL+"/api/auctions",
		sellerToken,
		map[string]any{
			"listing_id":        listingID,
			"starting_price":    5000.0,
			"minimum_increment": 250.0,
			"reserve_price":     5500.0,
			"start_time":        startTime,
			"end_time":          endTime,
		},
	)
	requireHTTPStatus(t, "create auction", createResp, http.StatusCreated)

	var created struct {
		Auction AuctionResponse `json:"auction"`
	}
	if err := json.Unmarshal(createResp.Body, &created); err != nil {
		t.Fatal(err)
	}
	auctionID = created.Auction.AuctionID

	if created.Auction.Status != "ACTIVE" {
		t.Fatalf("expected ACTIVE auction, got %s", created.Auction.Status)
	}

	var listingStatus string
	if err := DB.QueryRow(
		ctx,
		`SELECT status FROM public.listings WHERE listing_id = $1`,
		listingID,
	).Scan(&listingStatus); err != nil {
		t.Fatal(err)
	}
	if listingStatus != "ACTIVE" {
		t.Fatalf("expected ACTIVE listing, got %s", listingStatus)
	}

	auctionURL := fmt.Sprintf("%s/api/auctions/%d", server.URL, auctionID)
	bidURL := auctionURL + "/bids"
	closeURL := auctionURL + "/close"
	for _, tc := range []struct {
		name   string
		token  string
		amount float64
		status int
	}{
		{"seller self-bid", sellerToken, 5000, http.StatusForbidden},
		{"bidder A first bid", bidderAToken, 5000, http.StatusCreated},
		{"too-low bid", bidderBToken, 5100, http.StatusConflict},
		{"bidder B valid bid", bidderBToken, 5250, http.StatusCreated},
		{"bidder A final bid", bidderAToken, 5500, http.StatusCreated},
	} {
		resp := integrationRequest(t, http.MethodPost, bidURL, tc.token, map[string]any{"bid_amount": tc.amount})
		requireHTTPStatus(t, tc.name, resp, tc.status)
	}

	type bidResult struct {
		amount float64
		resp   httpTestResponse
		err    error
	}
	results := make(chan bidResult, 2)
	start := make(chan struct{})
	for _, tc := range []struct {
		token  string
		amount float64
	}{
		{bidderBToken, 5750},
		{bidderAToken, 6000},
	} {
		go func(token string, amount float64) {
			<-start
			resp, err := integrationRequestRaw(http.MethodPost, bidURL, token, map[string]any{"bid_amount": amount})
			results <- bidResult{amount: amount, resp: resp, err: err}
		}(tc.token, tc.amount)
	}
	close(start)

	for range 2 {
		result := <-results
		if result.err != nil {
			t.Fatal(result.err)
		}
		if result.amount == 6000 && result.resp.Status != http.StatusCreated {
			t.Fatalf("highest concurrent bid: expected %d, got %d: %s", http.StatusCreated, result.resp.Status, result.resp.Body)
		}
		if result.amount == 5750 && result.resp.Status != http.StatusCreated && result.resp.Status != http.StatusConflict {
			t.Fatalf("lower concurrent bid: expected %d or %d, got %d: %s", http.StatusCreated, http.StatusConflict, result.resp.Status, result.resp.Body)
		}
	}

	auctionResp := integrationRequest(
		t,
		http.MethodGet,
		auctionURL,
		"",
		nil,
	)
	requireHTTPStatus(t, "get auction", auctionResp, http.StatusOK)

	var auction AuctionResponse
	if err := json.Unmarshal(auctionResp.Body, &auction); err != nil {
		t.Fatal(err)
	}
	if auction.CurrentPrice != 6000 {
		t.Fatalf("expected current price 6000 after concurrent bids, got %.2f", auction.CurrentPrice)
	}
	if auction.CurrentHighestBidderID == nil || *auction.CurrentHighestBidderID != bidderAID {
		t.Fatalf("expected bidder A (%d) to be highest, got %v", bidderAID, auction.CurrentHighestBidderID)
	}

	bidsResp := integrationRequest(
		t,
		http.MethodGet,
		bidURL,
		"",
		nil,
	)
	requireHTTPStatus(t, "get bids", bidsResp, http.StatusOK)

	var bids []BidResponse
	if err := json.Unmarshal(bidsResp.Body, &bids); err != nil {
		t.Fatal(err)
	}
	if len(bids) < 4 || len(bids) > 5 {
		t.Fatalf("expected 4 or 5 accepted bids after concurrent burst, got %d", len(bids))
	}
	if bids[0].BidderID != bidderAID || bids[0].BidAmount != 6000 || bids[0].BidStatus != "ACCEPTED" {
		t.Fatalf("unexpected winning concurrent bid: %+v", bids[0])
	}
	for _, bid := range bids[1:] {
		if bid.BidStatus == "ACCEPTED" {
			t.Fatalf("only the highest bid should remain accepted: %+v", bid)
		}
	}

	if wait := time.Until(endTime) + 300*time.Millisecond; wait > 0 {
		time.Sleep(wait)
	}

	lateBid := integrationRequest(t, http.MethodPost, bidURL, bidderBToken, map[string]any{"bid_amount": 6250.0})
	requireHTTPStatus(t, "late bid", lateBid, http.StatusConflict)

	closeResp := integrationRequest(
		t,
		http.MethodPost,
		closeURL,
		sellerToken,
		nil,
	)
	requireHTTPStatus(t, "close auction", closeResp, http.StatusOK)

	var closed struct {
		Auction AuctionResponse `json:"auction"`
	}
	if err := json.Unmarshal(closeResp.Body, &closed); err != nil {
		t.Fatal(err)
	}
	if closed.Auction.Status != "ENDED" {
		t.Fatalf("expected ENDED auction, got %s", closed.Auction.Status)
	}
	if closed.Auction.WinnerID == nil || *closed.Auction.WinnerID != bidderAID {
		t.Fatalf("expected winner bidder A (%d), got %v", bidderAID, closed.Auction.WinnerID)
	}
	if closed.Auction.WinningBidID == nil {
		t.Fatal("expected winning_bid_id after successful auction")
	}

	if err := DB.QueryRow(
		ctx,
		`SELECT status FROM public.listings WHERE listing_id = $1`,
		listingID,
	).Scan(&listingStatus); err != nil {
		t.Fatal(err)
	}
	if listingStatus != "SOLD" {
		t.Fatalf("expected SOLD listing, got %s", listingStatus)
	}

	var winningBidStatus string
	if err := DB.QueryRow(
		ctx,
		`SELECT bid_status FROM public.bids WHERE bid_id = $1`,
		*closed.Auction.WinningBidID,
	).Scan(&winningBidStatus); err != nil {
		t.Fatal(err)
	}
	if winningBidStatus != "WON" {
		t.Fatalf("expected winning bid status WON, got %s", winningBidStatus)
	}

	closeAgain := integrationRequest(
		t,
		http.MethodPost,
		closeURL,
		sellerToken,
		nil,
	)
	requireHTTPStatus(t, "close ended auction", closeAgain, http.StatusConflict)

	publicResp := integrationRequest(
		t,
		http.MethodGet,
		server.URL+"/api/auctions",
		"",
		nil,
	)
	requireHTTPStatus(t, "public auction list", publicResp, http.StatusOK)

	var publicAuctions []AuctionListItem
	if err := json.Unmarshal(publicResp.Body, &publicAuctions); err != nil {
		t.Fatal(err)
	}
	for _, item := range publicAuctions {
		if item.AuctionID == auctionID {
			t.Fatalf("ended auction %d should not remain in scheduled/active public list", auctionID)
		}
	}

}

func integrationToken(t *testing.T, userID int64) string {
	t.Helper()

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		t.Fatal("JWT_SECRET is not loaded")
	}

	now := time.Now()
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.MapClaims{
			"sub": strconv.FormatInt(userID, 10),
			"iat": now.Unix(),
			"exp": now.Add(time.Hour).Unix(),
		},
	)

	value, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func integrationRequest(
	t *testing.T,
	method string,
	url string,
	token string,
	body any,
) httpTestResponse {
	t.Helper()
	resp, err := integrationRequestRaw(method, url, token, body)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func integrationRequestRaw(method, url, token string, body any) (httpTestResponse, error) {
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return httpTestResponse{}, err
		}
		reader = bytes.NewReader(raw)
	}

	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		return httpTestResponse{}, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return httpTestResponse{}, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return httpTestResponse{}, err
	}

	return httpTestResponse{Status: resp.StatusCode, Body: raw}, nil
}
