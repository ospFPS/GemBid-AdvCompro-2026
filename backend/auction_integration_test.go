package main

import (
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
	"golang.org/x/crypto/bcrypt"
)

type httpTestResponse struct {
	Status int
	Body   []byte
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

	passwordHash, err := bcrypt.GenerateFromPassword([]byte("integration-only"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}

	suffix := time.Now().UnixNano()
	var bidderAID, bidderBID int64

	err = DB.QueryRow(
		ctx,
		`
		INSERT INTO public.users (full_name, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING user_id
		`,
		"Auction Integration Bidder A",
		fmt.Sprintf("auction-it-a-%d@test.local", suffix),
		string(passwordHash),
	).Scan(&bidderAID)
	if err != nil {
		t.Fatalf("create bidder A: %v", err)
	}

	err = DB.QueryRow(
		ctx,
		`
		INSERT INTO public.users (full_name, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING user_id
		`,
		"Auction Integration Bidder B",
		fmt.Sprintf("auction-it-b-%d@test.local", suffix),
		string(passwordHash),
	).Scan(&bidderBID)
	if err != nil {
		t.Fatalf("create bidder B: %v", err)
	}

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

		if auctionID != 0 {
			_, _ = DB.Exec(cleanupCtx, `DELETE FROM public.notifications WHERE auction_id = $1 OR listing_id = $2`, auctionID, listingID)
			_, _ = DB.Exec(cleanupCtx, `DELETE FROM public.orders WHERE auction_id = $1`, auctionID)
			_, _ = DB.Exec(cleanupCtx, `DELETE FROM public.bids WHERE auction_id = $1`, auctionID)
			_, _ = DB.Exec(cleanupCtx, `DELETE FROM public.auctions WHERE auction_id = $1`, auctionID)
		}
		_, _ = DB.Exec(cleanupCtx, `DELETE FROM public.listings WHERE listing_id = $1`, listingID)
		_, _ = DB.Exec(cleanupCtx, `DELETE FROM public.gems WHERE gem_id = $1`, gemID)
		_, _ = DB.Exec(cleanupCtx, `DELETE FROM public.users WHERE user_id = ANY($1)`, []int64{bidderAID, bidderBID})
	})

	mux := http.NewServeMux()
	mux.HandleFunc(
		"POST /api/auctions",
		AuthMiddleware(VerifiedSellerMiddleware(CreateAuctionHandler)),
	)
	mux.HandleFunc("GET /api/auctions", ListAuctionsHandler)
	mux.HandleFunc("GET /api/auctions/{id}", GetAuctionHandler)
	mux.HandleFunc("GET /api/auctions/{id}/bids", ListAuctionBidsHandler)
	mux.HandleFunc(
		"POST /api/auctions/{id}/bids",
		AuthMiddleware(PlaceBidHandler),
	)
	mux.HandleFunc(
		"POST /api/auctions/{id}/close",
		AuthMiddleware(VerifiedSellerMiddleware(CloseAuctionHandler)),
	)

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
	if createResp.Status != http.StatusCreated {
		t.Fatalf("create auction: expected 201, got %d: %s", createResp.Status, createResp.Body)
	}

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

	sellerBid := integrationRequest(
		t,
		http.MethodPost,
		fmt.Sprintf("%s/api/auctions/%d/bids", server.URL, auctionID),
		sellerToken,
		map[string]any{"bid_amount": 5000.0},
	)
	if sellerBid.Status != http.StatusForbidden {
		t.Fatalf("seller self-bid should be 403, got %d: %s", sellerBid.Status, sellerBid.Body)
	}

	aFirst := integrationRequest(
		t,
		http.MethodPost,
		fmt.Sprintf("%s/api/auctions/%d/bids", server.URL, auctionID),
		bidderAToken,
		map[string]any{"bid_amount": 5000.0},
	)
	if aFirst.Status != http.StatusCreated {
		t.Fatalf("bidder A first bid: expected 201, got %d: %s", aFirst.Status, aFirst.Body)
	}

	bTooLow := integrationRequest(
		t,
		http.MethodPost,
		fmt.Sprintf("%s/api/auctions/%d/bids", server.URL, auctionID),
		bidderBToken,
		map[string]any{"bid_amount": 5100.0},
	)
	if bTooLow.Status != http.StatusConflict {
		t.Fatalf("too-low bid should be 409, got %d: %s", bTooLow.Status, bTooLow.Body)
	}

	bValid := integrationRequest(
		t,
		http.MethodPost,
		fmt.Sprintf("%s/api/auctions/%d/bids", server.URL, auctionID),
		bidderBToken,
		map[string]any{"bid_amount": 5250.0},
	)
	if bValid.Status != http.StatusCreated {
		t.Fatalf("bidder B valid bid: expected 201, got %d: %s", bValid.Status, bValid.Body)
	}

	aFinal := integrationRequest(
		t,
		http.MethodPost,
		fmt.Sprintf("%s/api/auctions/%d/bids", server.URL, auctionID),
		bidderAToken,
		map[string]any{"bid_amount": 5500.0},
	)
	if aFinal.Status != http.StatusCreated {
		t.Fatalf("bidder A final bid: expected 201, got %d: %s", aFinal.Status, aFinal.Body)
	}

	auctionResp := integrationRequest(
		t,
		http.MethodGet,
		fmt.Sprintf("%s/api/auctions/%d", server.URL, auctionID),
		"",
		nil,
	)
	if auctionResp.Status != http.StatusOK {
		t.Fatalf("get auction: expected 200, got %d: %s", auctionResp.Status, auctionResp.Body)
	}

	var auction AuctionResponse
	if err := json.Unmarshal(auctionResp.Body, &auction); err != nil {
		t.Fatal(err)
	}
	if auction.CurrentPrice != 5500 {
		t.Fatalf("expected current price 5500, got %.2f", auction.CurrentPrice)
	}
	if auction.CurrentHighestBidderID == nil || *auction.CurrentHighestBidderID != bidderAID {
		t.Fatalf("expected bidder A (%d) to be highest, got %v", bidderAID, auction.CurrentHighestBidderID)
	}

	bidsResp := integrationRequest(
		t,
		http.MethodGet,
		fmt.Sprintf("%s/api/auctions/%d/bids", server.URL, auctionID),
		"",
		nil,
	)
	if bidsResp.Status != http.StatusOK {
		t.Fatalf("get bids: expected 200, got %d: %s", bidsResp.Status, bidsResp.Body)
	}

	var bids []BidResponse
	if err := json.Unmarshal(bidsResp.Body, &bids); err != nil {
		t.Fatal(err)
	}
	if len(bids) != 3 {
		t.Fatalf("expected 3 accepted bids in history, got %d", len(bids))
	}
	if bids[0].BidderID != bidderAID || bids[0].BidAmount != 5500 || bids[0].BidStatus != "ACCEPTED" {
		t.Fatalf("unexpected latest bid: %+v", bids[0])
	}
	if bids[1].BidderID != bidderBID || bids[1].BidAmount != 5250 || bids[1].BidStatus != "OUTBID" {
		t.Fatalf("unexpected second bid: %+v", bids[1])
	}
	if bids[2].BidderID != bidderAID || bids[2].BidAmount != 5000 || bids[2].BidStatus != "OUTBID" {
		t.Fatalf("unexpected first bid: %+v", bids[2])
	}

	if wait := time.Until(endTime) + 300*time.Millisecond; wait > 0 {
		time.Sleep(wait)
	}

	lateBid := integrationRequest(
		t,
		http.MethodPost,
		fmt.Sprintf("%s/api/auctions/%d/bids", server.URL, auctionID),
		bidderBToken,
		map[string]any{"bid_amount": 5750.0},
	)
	if lateBid.Status != http.StatusConflict {
		t.Fatalf("late bid should be 409, got %d: %s", lateBid.Status, lateBid.Body)
	}

	closeResp := integrationRequest(
		t,
		http.MethodPost,
		fmt.Sprintf("%s/api/auctions/%d/close", server.URL, auctionID),
		sellerToken,
		nil,
	)
	if closeResp.Status != http.StatusOK {
		t.Fatalf("close auction: expected 200, got %d: %s", closeResp.Status, closeResp.Body)
	}

	var closed struct {
		Auction AuctionResponse `json:"auction"`
		OrderID int64           `json:"order_id"`
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

	var orderBuyer int64
	var orderStatus string
	var orderTotal float64
	if err := DB.QueryRow(
		ctx,
		`SELECT buyer_id, status, total_amount FROM public.orders WHERE order_id = $1 AND auction_id = $2`,
		closed.OrderID, auctionID,
	).Scan(&orderBuyer, &orderStatus, &orderTotal); err != nil {
		t.Fatalf("expected winner order %d: %v", closed.OrderID, err)
	}
	if orderBuyer != bidderAID || orderStatus != "PENDING_PAYMENT" || orderTotal != 5500 {
		t.Fatalf("unexpected winner order: buyer=%d status=%s total=%.2f", orderBuyer, orderStatus, orderTotal)
	}

	var outbidCount int
	if err := DB.QueryRow(
		ctx,
		`SELECT count(*) FROM public.notifications WHERE auction_id = $1 AND type = 'OUTBID'`,
		auctionID,
	).Scan(&outbidCount); err != nil {
		t.Fatal(err)
	}
	if outbidCount != 2 {
		t.Fatalf("expected 2 OUTBID notifications, got %d", outbidCount)
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
		fmt.Sprintf("%s/api/auctions/%d/close", server.URL, auctionID),
		sellerToken,
		nil,
	)
	if closeAgain.Status != http.StatusConflict {
		t.Fatalf("closing ended auction should be 409, got %d: %s", closeAgain.Status, closeAgain.Body)
	}

	publicResp := integrationRequest(
		t,
		http.MethodGet,
		server.URL+"/api/auctions",
		"",
		nil,
	)
	if publicResp.Status != http.StatusOK {
		t.Fatalf("public auction list: expected 200, got %d: %s", publicResp.Status, publicResp.Body)
	}

	var publicAuctions []AuctionListItem
	if err := json.Unmarshal(publicResp.Body, &publicAuctions); err != nil {
		t.Fatal(err)
	}
	for _, item := range publicAuctions {
		if item.AuctionID == auctionID {
			t.Fatalf("ended auction %d should not remain in scheduled/active public list", auctionID)
		}
	}

	t.Logf(
		"PASS auction=%d bidderA=%d bidderB=%d final_price=5500 winner=%d listing=SOLD",
		auctionID,
		bidderAID,
		bidderBID,
		bidderAID,
	)
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

	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		reader = bytes.NewReader(raw)
	}

	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		t.Fatal(err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	return httpTestResponse{
		Status: resp.StatusCode,
		Body:   raw,
	}
}
