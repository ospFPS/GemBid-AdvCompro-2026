package test

import (
	. "adcompro-backend"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestAutomaticAuctionCloser(t *testing.T) {
	if os.Getenv("RUN_AUCTION_INTEGRATION") != "1" {
		t.Skip("set RUN_AUCTION_INTEGRATION=1 to run database integration test")
	}
	ConnectDatabase()
	t.Cleanup(func() { DB.Close() })
	ctx := context.Background()

	var sellerID int64
	if err := DB.QueryRow(ctx, `
		SELECT sp.seller_id
		FROM public.seller_profiles sp
		INNER JOIN public.users u ON u.user_id = sp.user_id
		WHERE sp.verification_status = 'VERIFIED' AND u.seller_status = 'VERIFIED'
		ORDER BY sp.seller_id LIMIT 1
	`).Scan(&sellerID); err != nil {
		t.Fatalf("need verified seller: %v", err)
	}

	suffix := time.Now().UnixNano()
	var bidderID int64
	if err := DB.QueryRow(ctx, `
		INSERT INTO public.users (username, full_name, email, password_hash)
		VALUES ($1, 'Auto Close Bidder', $2, 'integration-only')
		RETURNING user_id
	`, fmt.Sprintf("auto%d", suffix%1000000000), fmt.Sprintf("auto-close-%d@test.local", suffix)).Scan(&bidderID); err != nil {
		t.Fatal(err)
	}

	var gemID, listingID, auctionID, bidID int64
	if err := DB.QueryRow(ctx, `
		INSERT INTO public.gems (seller_id, gem_name, gem_type)
		VALUES ($1, $2, 'Ruby') RETURNING gem_id
	`, sellerID, fmt.Sprintf("Auto Close Gem %d", suffix)).Scan(&gemID); err != nil {
		t.Fatal(err)
	}
	if err := DB.QueryRow(ctx, `
		INSERT INTO public.listings (gem_id, seller_id, sale_type, status)
		VALUES ($1, $2, 'AUCTION', 'ACTIVE') RETURNING listing_id
	`, gemID, sellerID).Scan(&listingID); err != nil {
		t.Fatal(err)
	}
	if err := DB.QueryRow(ctx, `
		INSERT INTO public.auctions
			(listing_id, starting_price, minimum_increment, reserve_price, current_price,
			 current_highest_bidder_id, start_time, end_time, status)
		VALUES ($1, 5000, 250, 5000, 5000, $2, NOW() - INTERVAL '2 minutes',
			NOW() - INTERVAL '1 second', 'ACTIVE')
		RETURNING auction_id
	`, listingID, bidderID).Scan(&auctionID); err != nil {
		t.Fatal(err)
	}
	if err := DB.QueryRow(ctx, `
		INSERT INTO public.bids (auction_id, bidder_id, bid_amount, bid_status)
		VALUES ($1, $2, 5000, 'ACCEPTED') RETURNING bid_id
	`, auctionID, bidderID).Scan(&bidID); err != nil {
		t.Fatal(err)
	}
	if _, err := DB.Exec(ctx,
		`UPDATE public.auctions SET winning_bid_id = $1 WHERE auction_id = $2`,
		bidID, auctionID,
	); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		c := context.Background()
		_, _ = DB.Exec(c, `DELETE FROM public.notifications WHERE auction_id = $1 OR listing_id = $2 OR user_id = $3`,
			auctionID, listingID, bidderID)
		_, _ = DB.Exec(c, `DELETE FROM public.transactions WHERE order_id IN (SELECT order_id FROM public.orders WHERE listing_id = $1)`,
			listingID)
		_, _ = DB.Exec(c, `DELETE FROM public.orders WHERE listing_id = $1`, listingID)
		_, _ = DB.Exec(c, `DELETE FROM public.bids WHERE auction_id = $1`, auctionID)
		_, _ = DB.Exec(c, `DELETE FROM public.auctions WHERE auction_id = $1`, auctionID)
		_, _ = DB.Exec(c, `DELETE FROM public.listings WHERE listing_id = $1`, listingID)
		_, _ = DB.Exec(c, `DELETE FROM public.gems WHERE gem_id = $1`, gemID)
		_, _ = DB.Exec(c, `DELETE FROM public.users WHERE user_id = $1`, bidderID)
	})

	CloseDueAuctions(ctx)

	var status string
	var winnerID *int64
	if err := DB.QueryRow(ctx,
		`SELECT status, winner_id FROM public.auctions WHERE auction_id = $1`,
		auctionID,
	).Scan(&status, &winnerID); err != nil {
		t.Fatal(err)
	}
	if status != "ENDED" || winnerID == nil || *winnerID != bidderID {
		t.Fatalf("automatic closer failed: status=%s winner=%v", status, winnerID)
	}

	var orderStatus string
	if err := DB.QueryRow(ctx,
		`SELECT status FROM public.orders WHERE listing_id = $1 AND buyer_id = $2`,
		listingID, bidderID,
	).Scan(&orderStatus); err != nil {
		t.Fatal(err)
	}
	if orderStatus != "PENDING_PAYMENT" {
		t.Fatalf("expected PENDING_PAYMENT auction order, got %s", orderStatus)
	}
}

func TestFreeListingQuotaConcurrentPublish(t *testing.T) {
	if os.Getenv("RUN_AUCTION_INTEGRATION") != "1" {
		t.Skip("set RUN_AUCTION_INTEGRATION=1 to run database integration test")
	}
	ConnectDatabase()
	t.Cleanup(func() { DB.Close() })
	ctx := context.Background()
	suffix := time.Now().UnixNano()

	var userID, sellerID, gemID int64
	if err := DB.QueryRow(ctx, `
		INSERT INTO public.users (username, full_name, email, password_hash, seller_status)
		VALUES ($1, 'Quota Seller', $2, 'integration-only', 'VERIFIED') RETURNING user_id
	`, fmt.Sprintf("quota%d", suffix%1000000000), fmt.Sprintf("quota-%d@test.local", suffix)).Scan(&userID); err != nil {
		t.Fatal(err)
	}
	if err := DB.QueryRow(ctx, `
		INSERT INTO public.seller_profiles (user_id, shop_name, legal_name, verification_status)
		VALUES ($1, 'Quota Shop', 'Quota Seller', 'VERIFIED') RETURNING seller_id
	`, userID).Scan(&sellerID); err != nil {
		t.Fatal(err)
	}
	if err := DB.QueryRow(ctx, `
		INSERT INTO public.gems (seller_id, gem_name, gem_type)
		VALUES ($1, $2, 'Sapphire') RETURNING gem_id
	`, sellerID, fmt.Sprintf("Quota Gem %d", suffix)).Scan(&gemID); err != nil {
		t.Fatal(err)
	}

	listings := make([]int64, 0, 6)
	for i := 0; i < 6; i++ {
		status := "PUBLISHED"
		saleType := "FIXED_PRICE"
		var id int64
		if i >= 4 {
			status = "DRAFT"
			saleType = ""
		}
		if saleType == "" {
			if err := DB.QueryRow(ctx,
				`INSERT INTO public.listings (gem_id, seller_id) VALUES ($1, $2) RETURNING listing_id`,
				gemID, sellerID,
			).Scan(&id); err != nil {
				t.Fatal(err)
			}
		} else if err := DB.QueryRow(ctx, `
			INSERT INTO public.listings (gem_id, seller_id, sale_type, status, fixed_price, published_at)
			VALUES ($1, $2, $3, $4, 100, NOW()) RETURNING listing_id
		`, gemID, sellerID, saleType, status).Scan(&id); err != nil {
			t.Fatal(err)
		}
		listings = append(listings, id)
	}

	t.Cleanup(func() {
		c := context.Background()
		_, _ = DB.Exec(c, `DELETE FROM public.listings WHERE listing_id = ANY($1)`, listings)
		_, _ = DB.Exec(c, `DELETE FROM public.gems WHERE gem_id = $1`, gemID)
		_, _ = DB.Exec(c, `DELETE FROM public.seller_profiles WHERE seller_id = $1`, sellerID)
		_, _ = DB.Exec(c, `DELETE FROM public.users WHERE user_id = $1`, userID)
	})

	mux := http.NewServeMux()
	mux.HandleFunc("PUT /api/listings/{id}/publish", AuthMiddleware(VerifiedSellerMiddleware(PublishListingHandler)))
	server := httptest.NewServer(mux)
	defer server.Close()
	token := integrationToken(t, userID)

	type result struct {
		resp httpTestResponse
		err  error
	}
	results := make(chan result, 2)
	start := make(chan struct{})
	for _, listingID := range listings[4:] {
		go func(id int64) {
			<-start
			resp, err := integrationRequestRaw(http.MethodPut,
				fmt.Sprintf("%s/api/listings/%d/publish", server.URL, id),
				token,
				map[string]any{"sale_type": "FIXED_PRICE", "fixed_price": 100.0},
			)
			results <- result{resp: resp, err: err}
		}(listingID)
	}
	close(start)

	okCount, forbiddenCount := 0, 0
	for range 2 {
		result := <-results
		if result.err != nil {
			t.Fatal(result.err)
		}
		switch result.resp.Status {
		case http.StatusOK:
			okCount++
		case http.StatusForbidden:
			forbiddenCount++
		default:
			t.Fatalf("unexpected concurrent publish response %d: %s", result.resp.Status, result.resp.Body)
		}
	}
	if okCount != 1 || forbiddenCount != 1 {
		t.Fatalf("expected one publish and one quota rejection, got ok=%d forbidden=%d", okCount, forbiddenCount)
	}

	var live int
	if err := DB.QueryRow(ctx,
		`SELECT count(*) FROM public.listings WHERE seller_id = $1 AND status IN ('PUBLISHED','ACTIVE')`,
		sellerID,
	).Scan(&live); err != nil {
		t.Fatal(err)
	}
	if live != 5 {
		t.Fatalf("expected exactly %d live listings, got %d", 5, live)
	}
}
