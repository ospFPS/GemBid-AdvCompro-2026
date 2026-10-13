package gembid

import (
	"context"
	"errors"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
)

type CreateAuctionRequest struct {
	ListingID        int64     `json:"listing_id"`
	StartingPrice    float64   `json:"starting_price"`
	MinimumIncrement float64   `json:"minimum_increment"`
	ReservePrice     *float64  `json:"reserve_price"`
	StartTime        time.Time `json:"start_time"`
	EndTime          time.Time `json:"end_time"`
}

type AuctionResponse struct {
	AuctionID              int64      `json:"auction_id"`
	ListingID              int64      `json:"listing_id"`
	StartingPrice          float64    `json:"starting_price"`
	MinimumIncrement       float64    `json:"minimum_increment"`
	ReservePrice           *float64   `json:"reserve_price"`
	CurrentPrice           float64    `json:"current_price"`
	CurrentHighestBidderID *int64     `json:"current_highest_bidder_id"`
	StartTime              time.Time  `json:"start_time"`
	EndTime                time.Time  `json:"end_time"`
	Status                 string     `json:"status"`
	WinnerID               *int64     `json:"winner_id"`
	WinningBidID           *int64     `json:"winning_bid_id"`
	CreatedAt              time.Time  `json:"created_at"`
	ClosedAt               *time.Time `json:"closed_at"`
}

type AuctionListItem struct {
	AuctionID              int64     `json:"auction_id"`
	ListingID              int64     `json:"listing_id"`
	GemID                  int64     `json:"gem_id"`
	GemName                string    `json:"gem_name"`
	GemType                string    `json:"gem_type"`
	StartingPrice          float64   `json:"starting_price"`
	MinimumIncrement       float64   `json:"minimum_increment"`
	CurrentPrice           float64   `json:"current_price"`
	CurrentHighestBidderID *int64    `json:"current_highest_bidder_id"`
	StartTime              time.Time `json:"start_time"`
	EndTime                time.Time `json:"end_time"`
	Status                 string    `json:"status"`
}

type PlaceBidRequest struct {
	BidAmount float64 `json:"bid_amount"`
}

type BidResponse struct {
	BidID     int64     `json:"bid_id"`
	AuctionID int64     `json:"auction_id"`
	BidderID  int64     `json:"bidder_id"`
	BidAmount float64   `json:"bid_amount"`
	BidTime   time.Time `json:"bid_time"`
	BidStatus string    `json:"bid_status"`
}

func roundMoney(v float64) float64 {
	return math.Round(v*100) / 100
}

const auctionColumns = `auction_id, listing_id, starting_price, minimum_increment,
	reserve_price, current_price, current_highest_bidder_id,
	start_time, end_time, status, winner_id, winning_bid_id,
	created_at, closed_at`

func scanAuction(row pgx.Row, auction *AuctionResponse) error {
	return row.Scan(
		&auction.AuctionID,
		&auction.ListingID,
		&auction.StartingPrice,
		&auction.MinimumIncrement,
		&auction.ReservePrice,
		&auction.CurrentPrice,
		&auction.CurrentHighestBidderID,
		&auction.StartTime,
		&auction.EndTime,
		&auction.Status,
		&auction.WinnerID,
		&auction.WinningBidID,
		&auction.CreatedAt,
		&auction.ClosedAt,
	)
}

func CreateAuctionHandler(w http.ResponseWriter, r *http.Request) {
	sellerID, _ := r.Context().Value(sellerIDKey).(int64)

	var request CreateAuctionRequest
	if !decodeJSON(w, r, &request) {
		return
	}

	request.StartingPrice = roundMoney(request.StartingPrice)
	request.MinimumIncrement = roundMoney(request.MinimumIncrement)
	if request.ReservePrice != nil {
		value := roundMoney(*request.ReservePrice)
		request.ReservePrice = &value
	}

	if request.ListingID <= 0 {
		http.Error(w, "listing_id is required", http.StatusBadRequest)
		return
	}
	if request.StartingPrice < 0 || request.StartingPrice > 9999999999.99 {
		http.Error(w, "starting_price is outside the supported money range", http.StatusBadRequest)
		return
	}
	if request.MinimumIncrement <= 0 || request.MinimumIncrement > 9999999999.99 {
		http.Error(w, "minimum_increment is outside the supported money range", http.StatusBadRequest)
		return
	}
	if request.ReservePrice != nil &&
		(*request.ReservePrice < request.StartingPrice || *request.ReservePrice > 9999999999.99) {
		http.Error(w, "reserve_price must be >= starting_price and within the supported money range", http.StatusBadRequest)
		return
	}
	if request.StartTime.IsZero() || request.EndTime.IsZero() {
		http.Error(w, "start_time and end_time are required", http.StatusBadRequest)
		return
	}
	if !request.EndTime.After(request.StartTime) {
		http.Error(w, "end_time must be after start_time", http.StatusBadRequest)
		return
	}
	if !request.EndTime.After(time.Now()) {
		http.Error(w, "end_time must be in the future", http.StatusBadRequest)
		return
	}

	tx, err := DB.Begin(r.Context())
	if err != nil {
		http.Error(w, "Could not start auction transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(r.Context())

	if _, err := tx.Exec(r.Context(),
		`SELECT 1 FROM public.seller_profiles WHERE seller_id = $1 FOR UPDATE`,
		sellerID,
	); err != nil {
		http.Error(w, "Could not lock seller account", http.StatusInternalServerError)
		return
	}

	var ownerID int64
	var saleType string
	var listingStatus string

	err = tx.QueryRow(
		r.Context(),
		`
		SELECT seller_id, COALESCE(sale_type, ''), status
		FROM public.listings
		WHERE listing_id = $1
		FOR UPDATE
		`,
		request.ListingID,
	).Scan(&ownerID, &saleType, &listingStatus)

	if err == pgx.ErrNoRows {
		http.Error(w, "Listing not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Could not read listing", http.StatusInternalServerError)
		return
	}
	if ownerID != sellerID {
		http.Error(w, "You do not own this listing", http.StatusForbidden)
		return
	}
	if listingStatus != "DRAFT" {
		http.Error(w, "Only a DRAFT listing can be configured as an auction", http.StatusConflict)
		return
	}
	if saleType != "AUCTION" {
		http.Error(w, "Set the listing sale_type to AUCTION before creating the auction", http.StatusConflict)
		return
	}
	if msg, err := checkListingQuota(r.Context(), tx, sellerID); err != nil {
		http.Error(w, "Could not check listing quota", http.StatusInternalServerError)
		return
	} else if msg != "" {
		http.Error(w, msg, http.StatusForbidden)
		return
	}

	auctionStatus := "SCHEDULED"
	newListingStatus := "PUBLISHED"
	if !request.StartTime.After(time.Now()) {
		auctionStatus = "ACTIVE"
		newListingStatus = "ACTIVE"
	}

	var auction AuctionResponse
	err = scanAuction(
		tx.QueryRow(
			r.Context(),
			`
			INSERT INTO public.auctions (
				listing_id,
				starting_price,
				minimum_increment,
				reserve_price,
				current_price,
				start_time,
				end_time,
				status
			)
			VALUES ($1, $2, $3, $4, $2, $5, $6, $7)
			RETURNING `+auctionColumns,
			request.ListingID,
			request.StartingPrice,
			request.MinimumIncrement,
			request.ReservePrice,
			request.StartTime,
			request.EndTime,
			auctionStatus,
		),
		&auction,
	)

	if err != nil {
		if isPGError(err, "23505") {
			http.Error(w, "This listing already has an auction", http.StatusConflict)
			return
		}
		http.Error(w, "Could not create auction", http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec(
		r.Context(),
		`
		UPDATE public.listings
		SET status = $1,
			published_at = COALESCE(published_at, NOW()),
			updated_at = NOW()
		WHERE listing_id = $2
		`,
		newListingStatus,
		request.ListingID,
	)
	if err != nil {
		http.Error(w, "Could not update listing status", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(r.Context()); err != nil {
		http.Error(w, "Could not commit auction", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "Auction created",
		"auction": auction,
	})
}

func ListAuctionsHandler(w http.ResponseWriter, r *http.Request) {
	activateDueAuctions(r.Context(), 0)

	rows, err := DB.Query(
		r.Context(),
		`
		SELECT
			a.auction_id,
			a.listing_id,
			g.gem_id,
			g.gem_name,
			g.gem_type,
			a.starting_price,
			a.minimum_increment,
			a.current_price,
			a.current_highest_bidder_id,
			a.start_time,
			a.end_time,
			a.status
		FROM public.auctions a
		INNER JOIN public.listings l ON l.listing_id = a.listing_id
		INNER JOIN public.gems g ON g.gem_id = l.gem_id
		WHERE a.status IN ('SCHEDULED', 'ACTIVE')
		ORDER BY
			CASE WHEN a.status = 'ACTIVE' THEN 0 ELSE 1 END,
			a.end_time ASC
		`,
	)
	if err != nil {
		http.Error(w, "Could not fetch auctions", http.StatusInternalServerError)
		return
	}
	auctions, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (AuctionListItem, error) {
		var item AuctionListItem
		err := row.Scan(
			&item.AuctionID,
			&item.ListingID,
			&item.GemID,
			&item.GemName,
			&item.GemType,
			&item.StartingPrice,
			&item.MinimumIncrement,
			&item.CurrentPrice,
			&item.CurrentHighestBidderID,
			&item.StartTime,
			&item.EndTime,
			&item.Status,
		)
		return item, err
	})
	if err != nil {
		http.Error(w, "Could not read auction data", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, auctions)
}

func GetAuctionHandler(w http.ResponseWriter, r *http.Request) {
	auctionID, ok := pathID(r)
	if !ok {
		http.Error(w, "Invalid auction ID", http.StatusBadRequest)
		return
	}

	activateDueAuctions(r.Context(), auctionID)

	var auction AuctionResponse
	err := scanAuction(
		DB.QueryRow(
			r.Context(),
			`SELECT `+auctionColumns+` FROM public.auctions WHERE auction_id = $1`,
			auctionID,
		),
		&auction,
	)
	if err == pgx.ErrNoRows {
		http.Error(w, "Auction not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Could not fetch auction", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, auction)
}

func ListAuctionBidsHandler(w http.ResponseWriter, r *http.Request) {
	auctionID, ok := pathID(r)
	if !ok {
		http.Error(w, "Invalid auction ID", http.StatusBadRequest)
		return
	}

	var exists bool
	if err := DB.QueryRow(
		r.Context(),
		`SELECT EXISTS(SELECT 1 FROM public.auctions WHERE auction_id = $1)`,
		auctionID,
	).Scan(&exists); err != nil {
		http.Error(w, "Could not verify auction", http.StatusInternalServerError)
		return
	}
	if !exists {
		http.Error(w, "Auction not found", http.StatusNotFound)
		return
	}

	rows, err := DB.Query(
		r.Context(),
		`
		SELECT bid_id, auction_id, bidder_id, bid_amount, bid_time, bid_status
		FROM public.bids
		WHERE auction_id = $1
		ORDER BY bid_time DESC, bid_id DESC
		`,
		auctionID,
	)
	if err != nil {
		http.Error(w, "Could not fetch bids", http.StatusInternalServerError)
		return
	}
	bids, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (BidResponse, error) {
		var bid BidResponse
		err := row.Scan(
			&bid.BidID,
			&bid.AuctionID,
			&bid.BidderID,
			&bid.BidAmount,
			&bid.BidTime,
			&bid.BidStatus,
		)
		return bid, err
	})
	if err != nil {
		http.Error(w, "Could not read bid data", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, bids)
}

func PlaceBidHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(userIDKey).(int64)

	auctionID, ok := pathID(r)
	if !ok {
		http.Error(w, "Invalid auction ID", http.StatusBadRequest)
		return
	}

	var request PlaceBidRequest
	if !decodeJSON(w, r, &request) {
		return
	}

	request.BidAmount = roundMoney(request.BidAmount)
	if request.BidAmount <= 0 || request.BidAmount > 9999999999.99 {
		http.Error(w, "bid_amount is outside the supported money range", http.StatusBadRequest)
		return
	}

	tx, err := DB.Begin(r.Context())
	if err != nil {
		http.Error(w, "Could not start bid transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(r.Context())

	var status string
	var startTime, endTime time.Time
	var currentPrice, minimumIncrement float64
	var currentWinningBidID *int64
	var listingID int64
	var sellerUserID int64

	err = tx.QueryRow(
		r.Context(),
		`
		SELECT
			a.status,
			a.start_time,
			a.end_time,
			a.current_price,
			a.minimum_increment,
			a.winning_bid_id,
			a.listing_id,
			sp.user_id
		FROM public.auctions a
		INNER JOIN public.listings l ON l.listing_id = a.listing_id
		INNER JOIN public.seller_profiles sp ON sp.seller_id = l.seller_id
		WHERE a.auction_id = $1
		FOR UPDATE OF a
		`,
		auctionID,
	).Scan(
		&status,
		&startTime,
		&endTime,
		&currentPrice,
		&minimumIncrement,
		&currentWinningBidID,
		&listingID,
		&sellerUserID,
	)

	if err == pgx.ErrNoRows {
		http.Error(w, "Auction not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Could not load auction", http.StatusInternalServerError)
		return
	}

	now := time.Now()

	if status == "SCHEDULED" && !now.Before(startTime) && now.Before(endTime) {
		status = "ACTIVE"
		if _, err := tx.Exec(
			r.Context(),
			`UPDATE public.auctions SET status = 'ACTIVE' WHERE auction_id = $1`,
			auctionID,
		); err != nil {
			http.Error(w, "Could not activate auction", http.StatusInternalServerError)
			return
		}
		if _, err := tx.Exec(
			r.Context(),
			`UPDATE public.listings SET status = 'ACTIVE', updated_at = NOW() WHERE listing_id = $1`,
			listingID,
		); err != nil {
			http.Error(w, "Could not activate listing", http.StatusInternalServerError)
			return
		}
	}

	if userID == sellerUserID {
		http.Error(w, "You cannot bid on your own auction", http.StatusForbidden)
		return
	}
	if now.Before(startTime) {
		http.Error(w, "Auction has not started yet", http.StatusConflict)
		return
	}
	if !now.Before(endTime) {
		http.Error(w, "Auction has ended", http.StatusConflict)
		return
	}
	if status != "ACTIVE" {
		http.Error(w, "Auction is not active", http.StatusConflict)
		return
	}

	minimumBid := currentPrice
	if currentWinningBidID != nil {
		minimumBid = roundMoney(currentPrice + minimumIncrement)
	}

	if request.BidAmount < minimumBid {
		http.Error(w, "Bid is below the minimum allowed amount of "+strconv.FormatFloat(minimumBid, 'f', 2, 64), http.StatusConflict)
		return
	}

	if currentWinningBidID != nil {
		var previousBidderID int64
		if err := tx.QueryRow(
			r.Context(),
			`UPDATE public.bids SET bid_status = 'OUTBID' WHERE bid_id = $1 RETURNING bidder_id`,
			*currentWinningBidID,
		).Scan(&previousBidderID); err != nil {
			http.Error(w, "Could not update previous bid", http.StatusInternalServerError)
			return
		}
		if previousBidderID != userID {
			if err := notify(r.Context(), tx, previousBidderID, "OUTBID", "You have been outbid",
				"Someone bid "+formatTHB(request.BidAmount)+" on an auction you were leading. Bid again to win it.",
				auctionID, listingID, 0); err != nil {
				http.Error(w, "Could not create notification", http.StatusInternalServerError)
				return
			}
		}
	}

	var bid BidResponse
	err = tx.QueryRow(
		r.Context(),
		`
		INSERT INTO public.bids (auction_id, bidder_id, bid_amount, bid_status)
		VALUES ($1, $2, $3, 'ACCEPTED')
		RETURNING bid_id, auction_id, bidder_id, bid_amount, bid_time, bid_status
		`,
		auctionID,
		userID,
		request.BidAmount,
	).Scan(
		&bid.BidID,
		&bid.AuctionID,
		&bid.BidderID,
		&bid.BidAmount,
		&bid.BidTime,
		&bid.BidStatus,
	)
	if err != nil {
		http.Error(w, "Could not create bid", http.StatusInternalServerError)
		return
	}

	if _, err := tx.Exec(
		r.Context(),
		`
		UPDATE public.auctions
		SET current_price = $1,
			current_highest_bidder_id = $2,
			winning_bid_id = $3
		WHERE auction_id = $4
		`,
		request.BidAmount,
		userID,
		bid.BidID,
		auctionID,
	); err != nil {
		http.Error(w, "Could not update auction price", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(r.Context()); err != nil {
		http.Error(w, "Could not commit bid", http.StatusInternalServerError)
		return
	}

	nextMinimum := roundMoney(request.BidAmount + minimumIncrement)

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"message":          "Bid accepted",
		"bid":              bid,
		"current_price":    request.BidAmount,
		"next_minimum_bid": nextMinimum,
	})
}

func CloseAuctionHandler(w http.ResponseWriter, r *http.Request) {
	sellerID, _ := r.Context().Value(sellerIDKey).(int64)

	auctionID, ok := pathID(r)
	if !ok {
		http.Error(w, "Invalid auction ID", http.StatusBadRequest)
		return
	}

	tx, err := DB.BeginTx(r.Context(), pgx.TxOptions{})
	if err != nil {
		http.Error(w, "Could not start close transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(r.Context())

	var ownerSellerID int64
	err = tx.QueryRow(
		r.Context(),
		`
		SELECT l.seller_id
		FROM public.auctions a
		INNER JOIN public.listings l ON l.listing_id = a.listing_id
		WHERE a.auction_id = $1
		`,
		auctionID,
	).Scan(&ownerSellerID)
	if err == pgx.ErrNoRows {
		http.Error(w, "Auction not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Could not load auction", http.StatusInternalServerError)
		return
	}
	if ownerSellerID != sellerID {
		http.Error(w, "You do not own this auction", http.StatusForbidden)
		return
	}

	result, err := finalizeAuction(r.Context(), tx, auctionID)
	if errors.Is(err, errAuctionAlreadyClosed) || errors.Is(err, errAuctionNotEnded) {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	if err != nil {
		http.Error(w, "Could not close auction", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(r.Context()); err != nil {
		http.Error(w, "Could not commit auction close", http.StatusInternalServerError)
		return
	}

	var auction AuctionResponse
	err = scanAuction(
		DB.QueryRow(
			r.Context(),
			`SELECT `+auctionColumns+` FROM public.auctions WHERE auction_id = $1`,
			auctionID,
		),
		&auction,
	)
	if err != nil {
		http.Error(w, "Auction closed but response could not be loaded", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"message": "Auction closed without a winner",
		"auction": auction,
	}
	if result.HasWinner {
		response["message"] = "Auction closed with a winner"
		response["order_id"] = result.OrderID
	}

	writeJSON(w, http.StatusOK, response)
}

var (
	errAuctionAlreadyClosed = errors.New("Auction is already closed")
	errAuctionNotEnded      = errors.New("Auction cannot be closed before end_time")
)

type auctionResult struct {
	HasWinner bool
	OrderID   int64
}

func formatTHB(v float64) string {
	return strconv.FormatFloat(v, 'f', 2, 64) + " THB"
}

func finalizeAuction(ctx context.Context, tx pgx.Tx, auctionID int64) (auctionResult, error) {
	var result auctionResult
	var listingID, sellerID, sellerUserID int64
	var status string
	var endTime time.Time
	var reservePrice *float64
	var currentPrice float64
	var highestBidderID, winningBidID *int64

	err := tx.QueryRow(
		ctx,
		`
		SELECT
			l.listing_id,
			l.seller_id,
			sp.user_id,
			a.status,
			a.end_time,
			a.reserve_price,
			a.current_price,
			a.current_highest_bidder_id,
			a.winning_bid_id
		FROM public.auctions a
		INNER JOIN public.listings l ON l.listing_id = a.listing_id
		INNER JOIN public.seller_profiles sp ON sp.seller_id = l.seller_id
		WHERE a.auction_id = $1
		FOR UPDATE OF a, l
		`,
		auctionID,
	).Scan(&listingID, &sellerID, &sellerUserID, &status, &endTime,
		&reservePrice, &currentPrice, &highestBidderID, &winningBidID)
	if err != nil {
		return result, err
	}
	if status == "ENDED" {
		return result, errAuctionAlreadyClosed
	}
	if time.Now().Before(endTime) {
		return result, errAuctionNotEnded
	}

	reserveMet := reservePrice == nil || currentPrice >= *reservePrice
	result.HasWinner = highestBidderID != nil && winningBidID != nil && reserveMet

	if !result.HasWinner {
		if winningBidID != nil {
			if _, err := tx.Exec(ctx,
				`UPDATE public.bids SET bid_status = 'RESERVE_NOT_MET' WHERE bid_id = $1`,
				*winningBidID); err != nil {
				return result, err
			}
		}
		if _, err := tx.Exec(
			ctx,
			`
			UPDATE public.auctions
			SET status = 'ENDED',
				winner_id = NULL,
				winning_bid_id = NULL,
				closed_at = NOW()
			WHERE auction_id = $1
			`,
			auctionID,
		); err != nil {
			return result, err
		}
		if _, err := tx.Exec(ctx,
			`UPDATE public.listings SET status = 'UNSOLD', updated_at = NOW() WHERE listing_id = $1`,
			listingID); err != nil {
			return result, err
		}

		if err := notify(ctx, tx, sellerUserID, "AUCTION_UNSOLD", "Auction ended unsold",
			"Your auction ended without a winning bid that met the reserve price.",
			auctionID, listingID, 0); err != nil {
			return result, err
		}
		if highestBidderID != nil {
			if err := notify(ctx, tx, *highestBidderID, "AUCTION_RESERVE_NOT_MET", "Reserve price not met",
				"The auction ended but your highest bid of "+formatTHB(currentPrice)+" did not meet the reserve price.",
				auctionID, listingID, 0); err != nil {
				return result, err
			}
		}
		return result, nil
	}

	if _, err := tx.Exec(
		ctx,
		`
		UPDATE public.auctions
		SET status = 'ENDED',
			winner_id = $1,
			closed_at = NOW()
		WHERE auction_id = $2
		`,
		*highestBidderID,
		auctionID,
	); err != nil {
		return result, err
	}
	if _, err := tx.Exec(ctx,
		`UPDATE public.bids SET bid_status = 'WON' WHERE bid_id = $1`, *winningBidID); err != nil {
		return result, err
	}
	if _, err := tx.Exec(ctx,
		`UPDATE public.listings SET status = 'SOLD', updated_at = NOW() WHERE listing_id = $1`,
		listingID); err != nil {
		return result, err
	}

	order, err := createOrder(ctx, tx, listingID, *highestBidderID, sellerID,
		auctionID, *winningBidID, currentPrice)
	if err != nil {
		return result, err
	}
	result.OrderID = order.OrderID

	if err := notify(ctx, tx, *highestBidderID, "AUCTION_WON", "You won the auction!",
		"Your bid of "+formatTHB(currentPrice)+" won. Please pay for your order to receive the gem.",
		auctionID, listingID, result.OrderID); err != nil {
		return result, err
	}
	if err := notify(ctx, tx, sellerUserID, "AUCTION_SOLD", "Your auction sold",
		"Your auction ended with a winning bid of "+formatTHB(currentPrice)+". Waiting for the buyer's payment.",
		auctionID, listingID, result.OrderID); err != nil {
		return result, err
	}

	rows, err := tx.Query(ctx,
		`SELECT DISTINCT bidder_id FROM public.bids WHERE auction_id = $1 AND bidder_id <> $2`,
		auctionID, *highestBidderID)
	if err != nil {
		return result, err
	}
	losers, err := pgx.CollectRows(rows, pgx.RowTo[int64])
	if err != nil {
		return result, err
	}
	for _, loserID := range losers {
		if err := notify(ctx, tx, loserID, "AUCTION_LOST", "Auction ended",
			"The auction ended with a winning bid of "+formatTHB(currentPrice)+". You did not win this time.",
			auctionID, listingID, 0); err != nil {
			return result, err
		}
	}

	return result, nil
}

func activateDueAuctions(ctx context.Context, auctionID int64) {
	_, _ = DB.Exec(
		ctx,
		`
		WITH activated AS (
			UPDATE public.auctions
			SET status = 'ACTIVE'
			WHERE ($1::bigint = 0 OR auction_id = $1::bigint)
			  AND status = 'SCHEDULED'
			  AND start_time <= NOW()
			  AND end_time > NOW()
			RETURNING listing_id
		)
		UPDATE public.listings l
		SET status = 'ACTIVE', updated_at = NOW()
		FROM activated a
		WHERE l.listing_id = a.listing_id
		`,
		auctionID,
	)
}

func StartAuctionCloser() {
	go func() {
		for {
			CloseDueAuctions(context.Background())
			time.Sleep(30 * time.Second)
		}
	}()
}

func CloseDueAuctions(ctx context.Context) {
	activateDueAuctions(ctx, 0)

	rows, err := DB.Query(ctx,
		`SELECT auction_id FROM public.auctions
		 WHERE status IN ('SCHEDULED', 'ACTIVE') AND end_time <= NOW()
		 ORDER BY end_time`)
	if err != nil {
		log.Println("auction closer:", err)
		return
	}
	due, err := pgx.CollectRows(rows, pgx.RowTo[int64])
	if err != nil {
		log.Println("auction closer:", err)
		return
	}

	for _, auctionID := range due {
		err := pgx.BeginFunc(ctx, DB, func(tx pgx.Tx) error {
			_, err := finalizeAuction(ctx, tx, auctionID)
			return err
		})
		if err != nil && !errors.Is(err, errAuctionAlreadyClosed) {
			log.Printf("auction closer: auction %d: %v", auctionID, err)
		}
	}
}

type MyBidItem struct {
	AuctionID     int64     `json:"auction_id"`
	ListingID     int64     `json:"listing_id"`
	GemName       string    `json:"gem_name"`
	GemType       string    `json:"gem_type"`
	MyHighestBid  float64   `json:"my_highest_bid"`
	MyBidStatus   string    `json:"my_bid_status"`
	CurrentPrice  float64   `json:"current_price"`
	AuctionStatus string    `json:"auction_status"`
	EndTime       time.Time `json:"end_time"`
	IsLeading     bool      `json:"is_leading"`
	IsWinner      bool      `json:"is_winner"`
	LastBidTime   time.Time `json:"last_bid_time"`
	ImageURL      *string   `json:"image_url"`
}

// MyBidsHandler returns one row per auction the user bid on, with their highest bid.
func MyBidsHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(userIDKey).(int64)

	rows, err := DB.Query(r.Context(), `
		SELECT * FROM (
			SELECT DISTINCT ON (a.auction_id)
				a.auction_id, a.listing_id, g.gem_name, g.gem_type,
				b.bid_amount, b.bid_status, a.current_price, a.status, a.end_time,
				COALESCE(a.current_highest_bidder_id = $1, FALSE),
				COALESCE(a.winner_id = $1, FALSE),
				b.bid_time, gi.image_url
			FROM public.bids b
			INNER JOIN public.auctions a ON a.auction_id = b.auction_id
			INNER JOIN public.listings l ON l.listing_id = a.listing_id
			INNER JOIN public.gems g ON g.gem_id = l.gem_id
			LEFT JOIN public.gem_images gi ON gi.gem_id = g.gem_id AND gi.is_primary
			WHERE b.bidder_id = $1
			ORDER BY a.auction_id, b.bid_amount DESC, b.bid_time DESC
		) mine
		ORDER BY bid_time DESC
	`, userID)
	if err != nil {
		http.Error(w, "Could not load bids", http.StatusInternalServerError)
		return
	}
	bids, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (MyBidItem, error) {
		var b MyBidItem
		err := row.Scan(&b.AuctionID, &b.ListingID, &b.GemName, &b.GemType, &b.MyHighestBid,
			&b.MyBidStatus, &b.CurrentPrice, &b.AuctionStatus, &b.EndTime, &b.IsLeading,
			&b.IsWinner, &b.LastBidTime, &b.ImageURL)
		return b, err
	})
	if err != nil {
		http.Error(w, "Could not read bids", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, bids)
}
