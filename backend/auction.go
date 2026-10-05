package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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

// CreateAuctionHandler godoc
// @Summary Create an auction
// @Description Configure an AUCTION draft listing with starting price, increment, optional reserve price, and start/end times.
// @Tags Auctions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateAuctionRequest true "Auction configuration"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {string} string
// @Failure 401 {string} string
// @Failure 403 {string} string
// @Failure 404 {string} string
// @Failure 409 {string} string
// @Router /api/auctions [post]
func CreateAuctionHandler(w http.ResponseWriter, r *http.Request) {
	sellerID, ok := r.Context().Value(sellerIDKey).(int64)
	if !ok {
		http.Error(w, "Verified seller account required", http.StatusForbidden)
		return
	}

	var request CreateAuctionRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
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
	if request.StartingPrice < 0 {
		http.Error(w, "starting_price must be >= 0", http.StatusBadRequest)
		return
	}
	if request.MinimumIncrement <= 0 {
		http.Error(w, "minimum_increment must be > 0", http.StatusBadRequest)
		return
	}
	if request.ReservePrice != nil && *request.ReservePrice < request.StartingPrice {
		http.Error(w, "reserve_price must be >= starting_price", http.StatusBadRequest)
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

	tx, err := DB.BeginTx(r.Context(), pgx.TxOptions{})
	if err != nil {
		http.Error(w, "Could not start auction transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(r.Context())

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
			RETURNING
				auction_id, listing_id, starting_price, minimum_increment,
				reserve_price, current_price, current_highest_bidder_id,
				start_time, end_time, status, winner_id, winning_bid_id,
				created_at, closed_at
			`,
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
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Auction created",
		"auction": auction,
	})
}

// ListAuctionsHandler godoc
// @Summary Browse auctions
// @Description Public endpoint returning scheduled and active auctions.
// @Tags Auctions
// @Produce json
// @Success 200 {array} AuctionListItem
// @Failure 500 {string} string
// @Router /api/auctions [get]
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
	defer rows.Close()

	auctions := []AuctionListItem{}
	for rows.Next() {
		var item AuctionListItem
		if err := rows.Scan(
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
		); err != nil {
			http.Error(w, "Could not read auction data", http.StatusInternalServerError)
			return
		}
		auctions = append(auctions, item)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(auctions)
}

// GetAuctionHandler godoc
// @Summary Get auction details
// @Description Public endpoint returning one auction.
// @Tags Auctions
// @Produce json
// @Param id path int true "Auction ID"
// @Success 200 {object} AuctionResponse
// @Failure 400 {string} string
// @Failure 404 {string} string
// @Router /api/auctions/{id} [get]
func GetAuctionHandler(w http.ResponseWriter, r *http.Request) {
	auctionID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || auctionID <= 0 {
		http.Error(w, "Invalid auction ID", http.StatusBadRequest)
		return
	}

	activateDueAuctions(r.Context(), auctionID)

	var auction AuctionResponse
	err = scanAuction(
		DB.QueryRow(
			r.Context(),
			`
			SELECT
				auction_id, listing_id, starting_price, minimum_increment,
				reserve_price, current_price, current_highest_bidder_id,
				start_time, end_time, status, winner_id, winning_bid_id,
				created_at, closed_at
			FROM public.auctions
			WHERE auction_id = $1
			`,
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(auction)
}

// ListAuctionBidsHandler godoc
// @Summary Get auction bid history
// @Description Public endpoint returning bids for one auction, newest first.
// @Tags Auctions
// @Produce json
// @Param id path int true "Auction ID"
// @Success 200 {array} BidResponse
// @Failure 400 {string} string
// @Failure 404 {string} string
// @Failure 500 {string} string
// @Router /api/auctions/{id}/bids [get]
func ListAuctionBidsHandler(w http.ResponseWriter, r *http.Request) {
	auctionID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || auctionID <= 0 {
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
	defer rows.Close()

	bids := []BidResponse{}
	for rows.Next() {
		var bid BidResponse
		if err := rows.Scan(
			&bid.BidID,
			&bid.AuctionID,
			&bid.BidderID,
			&bid.BidAmount,
			&bid.BidTime,
			&bid.BidStatus,
		); err != nil {
			http.Error(w, "Could not read bid data", http.StatusInternalServerError)
			return
		}
		bids = append(bids, bid)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bids)
}

// PlaceBidHandler godoc
// @Summary Place a bid
// @Description Place a concurrency-safe bid on an active auction. The seller cannot bid on their own auction.
// @Tags Auctions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Auction ID"
// @Param request body PlaceBidRequest true "Bid amount"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {string} string
// @Failure 401 {string} string
// @Failure 403 {string} string
// @Failure 404 {string} string
// @Failure 409 {string} string
// @Router /api/auctions/{id}/bids [post]
func PlaceBidHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(userIDKey).(int64)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	auctionID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || auctionID <= 0 {
		http.Error(w, "Invalid auction ID", http.StatusBadRequest)
		return
	}

	var request PlaceBidRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	request.BidAmount = roundMoney(request.BidAmount)
	if request.BidAmount <= 0 {
		http.Error(w, "bid_amount must be > 0", http.StatusBadRequest)
		return
	}

	tx, err := DB.BeginTx(r.Context(), pgx.TxOptions{})
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
		http.Error(
			w,
			"Bid is below the minimum allowed amount of "+strconv.FormatFloat(minimumBid, 'f', 2, 64),
			http.StatusConflict,
		)
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":          "Bid accepted",
		"bid":              bid,
		"current_price":    request.BidAmount,
		"next_minimum_bid": nextMinimum,
	})
}

// CloseAuctionHandler godoc
// @Summary Close an ended auction
// @Description Seller closes an auction after end_time (the backend also does this automatically every 30s). Reserve price is checked; a winner gets a PENDING_PAYMENT order, otherwise the listing becomes UNSOLD.
// @Tags Auctions
// @Produce json
// @Security BearerAuth
// @Param id path int true "Auction ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {string} string
// @Failure 401 {string} string
// @Failure 403 {string} string
// @Failure 404 {string} string
// @Failure 409 {string} string
// @Router /api/auctions/{id}/close [post]
func CloseAuctionHandler(w http.ResponseWriter, r *http.Request) {
	sellerID, ok := r.Context().Value(sellerIDKey).(int64)
	if !ok {
		http.Error(w, "Verified seller account required", http.StatusForbidden)
		return
	}

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
			`
			SELECT
				auction_id, listing_id, starting_price, minimum_increment,
				reserve_price, current_price, current_highest_bidder_id,
				start_time, end_time, status, winner_id, winning_bid_id,
				created_at, closed_at
			FROM public.auctions
			WHERE auction_id = $1
			`,
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ============================================================
// AUCTION FINALIZATION
// ============================================================

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

// finalizeAuction closes an auction whose end_time has passed. If the reserve
// is met the highest bid WINS, the listing becomes SOLD and a PENDING_PAYMENT
// order is created for the winner; otherwise the listing becomes UNSOLD.
// Seller and bidders are notified. Caller owns the transaction.
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

// activateDueAuctions moves SCHEDULED auctions whose start_time has passed to ACTIVE.
// auctionID 0 means all auctions.
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

// StartAuctionCloser activates and finalizes due auctions every 30 seconds,
// so auctions close on time without the seller pressing "close".
func StartAuctionCloser() {
	go func() {
		for {
			closeDueAuctions(context.Background())
			time.Sleep(30 * time.Second)
		}
	}()
}

func closeDueAuctions(ctx context.Context) {
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
