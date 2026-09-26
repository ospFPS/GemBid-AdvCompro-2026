package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// ============================================================
// LISTING RESPONSE
// ============================================================

type ListingResponse struct {
	ListingID             int64      `json:"listing_id"`
	GemID                 int64      `json:"gem_id"`
	SellerID              int64      `json:"seller_id"`
	SaleType              *string    `json:"sale_type"`
	Status                string     `json:"status"`
	FixedPrice            *float64   `json:"fixed_price"`
	CreatedAt             time.Time  `json:"created_at"`
	PublishedAt           *time.Time `json:"published_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
	RelistedFromListingID *int64     `json:"relisted_from_listing_id"`
}

// ============================================================
// POST /api/listings  (seller-owned — create a DRAFT listing for a gem)
// ============================================================

type CreateListingRequest struct {
	GemID int64 `json:"gem_id"`
}

func CreateListingHandler(w http.ResponseWriter, r *http.Request) {
	sellerID, ok := r.Context().Value(sellerIDKey).(int64)

	if !ok {
		http.Error(w, "Verified seller account required", http.StatusForbidden)
		return
	}

	var request CreateListingRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if request.GemID == 0 {
		http.Error(w, "gem_id is required", http.StatusBadRequest)
		return
	}

	// Confirm the gem exists AND belongs to this seller before creating a listing for it.
	var gemOwnerID int64

	err := DB.QueryRow(
		r.Context(),
		`SELECT seller_id FROM public.gems WHERE gem_id = $1`,
		request.GemID,
	).Scan(&gemOwnerID)

	if err == pgx.ErrNoRows {
		http.Error(w, "Gem not found", http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, "Could not verify gem ownership", http.StatusInternalServerError)
		return
	}

	if gemOwnerID != sellerID {
		http.Error(w, "You do not own this gem", http.StatusForbidden)
		return
	}

	var listing ListingResponse

	err = DB.QueryRow(
		r.Context(),
		`
		INSERT INTO public.listings (gem_id, seller_id)
		VALUES ($1, $2)
		RETURNING
			listing_id, gem_id, seller_id, sale_type, status, fixed_price,
			created_at, published_at, updated_at, relisted_from_listing_id
		`,
		request.GemID, sellerID,
	).Scan(
		&listing.ListingID, &listing.GemID, &listing.SellerID, &listing.SaleType,
		&listing.Status, &listing.FixedPrice, &listing.CreatedAt, &listing.PublishedAt,
		&listing.UpdatedAt, &listing.RelistedFromListingID,
	)

	if err != nil {
		http.Error(w, "Could not create listing", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Draft listing created",
		"listing": listing,
	})
}

// ============================================================
// GET /api/listings  (public — browse published/active listings)
// ============================================================

type ListingListItem struct {
	ListingID  int64    `json:"listing_id"`
	GemID      int64    `json:"gem_id"`
	GemName    string   `json:"gem_name"`
	GemType    string   `json:"gem_type"`
	SaleType   *string  `json:"sale_type"`
	Status     string   `json:"status"`
	FixedPrice *float64 `json:"fixed_price"`
}

func ListListingsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := DB.Query(
		r.Context(),
		`
		SELECT
			l.listing_id, l.gem_id, g.gem_name, g.gem_type, l.sale_type, l.status, l.fixed_price
		FROM public.listings l
		INNER JOIN public.gems g ON g.gem_id = l.gem_id
		WHERE l.status IN ('PUBLISHED', 'ACTIVE')
		ORDER BY l.published_at DESC NULLS LAST
		`,
	)

	if err != nil {
		http.Error(w, "Could not fetch listings", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	listings := []ListingListItem{}

	for rows.Next() {
		var item ListingListItem

		if err := rows.Scan(
			&item.ListingID, &item.GemID, &item.GemName, &item.GemType,
			&item.SaleType, &item.Status, &item.FixedPrice,
		); err != nil {
			http.Error(w, "Could not read listing data", http.StatusInternalServerError)
			return
		}

		listings = append(listings, item)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(listings)
}

// ============================================================
// GET /api/listings/{id}  (public — single listing detail)
// ============================================================

func GetListingHandler(w http.ResponseWriter, r *http.Request) {
	listingID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)

	if err != nil {
		http.Error(w, "Invalid listing ID", http.StatusBadRequest)
		return
	}

	var listing ListingResponse

	err = DB.QueryRow(
		r.Context(),
		`
		SELECT
			listing_id, gem_id, seller_id, sale_type, status, fixed_price,
			created_at, published_at, updated_at, relisted_from_listing_id
		FROM public.listings
		WHERE listing_id = $1
		`,
		listingID,
	).Scan(
		&listing.ListingID, &listing.GemID, &listing.SellerID, &listing.SaleType,
		&listing.Status, &listing.FixedPrice, &listing.CreatedAt, &listing.PublishedAt,
		&listing.UpdatedAt, &listing.RelistedFromListingID,
	)

	if err == pgx.ErrNoRows {
		http.Error(w, "Listing not found", http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, "Could not fetch listing", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(listing)
}

// ============================================================
// PUT /api/listings/{id}/publish  (seller-owned — DRAFT -> chosen sale type)
// ============================================================

type PublishListingRequest struct {
	SaleType   string   `json:"sale_type"` // "FIXED_PRICE" or "AUCTION"
	FixedPrice *float64 `json:"fixed_price"`
}

func PublishListingHandler(w http.ResponseWriter, r *http.Request) {
	sellerID, ok := r.Context().Value(sellerIDKey).(int64)

	if !ok {
		http.Error(w, "Verified seller account required", http.StatusForbidden)
		return
	}

	listingID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)

	if err != nil {
		http.Error(w, "Invalid listing ID", http.StatusBadRequest)
		return
	}

	var request PublishListingRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	var listing ListingResponse

	switch request.SaleType {

	case "FIXED_PRICE":
		if request.FixedPrice == nil || *request.FixedPrice < 0 {
			http.Error(w, "fixed_price is required and must be >= 0 for FIXED_PRICE listings", http.StatusBadRequest)
			return
		}

		err = DB.QueryRow(
			r.Context(),
			`
			UPDATE public.listings
			SET sale_type = 'FIXED_PRICE', fixed_price = $1, status = 'PUBLISHED',
				published_at = NOW(), updated_at = NOW()
			WHERE listing_id = $2 AND seller_id = $3 AND status = 'DRAFT'
			RETURNING
				listing_id, gem_id, seller_id, sale_type, status, fixed_price,
				created_at, published_at, updated_at, relisted_from_listing_id
			`,
			*request.FixedPrice, listingID, sellerID,
		).Scan(
			&listing.ListingID, &listing.GemID, &listing.SellerID, &listing.SaleType,
			&listing.Status, &listing.FixedPrice, &listing.CreatedAt, &listing.PublishedAt,
			&listing.UpdatedAt, &listing.RelistedFromListingID,
		)

	case "AUCTION":
		if request.FixedPrice != nil {
			http.Error(w, "fixed_price must not be set for AUCTION listings", http.StatusBadRequest)
			return
		}

		// Sale type is recorded here, but the listing stays in DRAFT status —
		// creating the actual auction (starting price, times, etc.) is a
		// separate step that will move it to ACTIVE. Not built yet.
		err = DB.QueryRow(
			r.Context(),
			`
			UPDATE public.listings
			SET sale_type = 'AUCTION', updated_at = NOW()
			WHERE listing_id = $1 AND seller_id = $2 AND status = 'DRAFT'
			RETURNING
				listing_id, gem_id, seller_id, sale_type, status, fixed_price,
				created_at, published_at, updated_at, relisted_from_listing_id
			`,
			listingID, sellerID,
		).Scan(
			&listing.ListingID, &listing.GemID, &listing.SellerID, &listing.SaleType,
			&listing.Status, &listing.FixedPrice, &listing.CreatedAt, &listing.PublishedAt,
			&listing.UpdatedAt, &listing.RelistedFromListingID,
		)

	default:
		http.Error(w, "sale_type must be FIXED_PRICE or AUCTION", http.StatusBadRequest)
		return
	}

	if err == pgx.ErrNoRows {
		http.Error(w, "Listing not found, not owned by you, or not in DRAFT status", http.StatusConflict)
		return
	}

	if err != nil {
		http.Error(w, "Could not publish listing", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Listing updated",
		"listing": listing,
	})
}

// ============================================================
// DELETE /api/listings/{id}  (seller-owned — only while still DRAFT)
// ============================================================

func DeleteListingHandler(w http.ResponseWriter, r *http.Request) {
	sellerID, ok := r.Context().Value(sellerIDKey).(int64)

	if !ok {
		http.Error(w, "Verified seller account required", http.StatusForbidden)
		return
	}

	listingID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)

	if err != nil {
		http.Error(w, "Invalid listing ID", http.StatusBadRequest)
		return
	}

	tag, err := DB.Exec(
		r.Context(),
		`
		DELETE FROM public.listings
		WHERE listing_id = $1 AND seller_id = $2 AND status = 'DRAFT'
		`,
		listingID, sellerID,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			http.Error(w, "Cannot delete: this listing is referenced elsewhere", http.StatusConflict)
			return
		}

		http.Error(w, "Could not delete listing", http.StatusInternalServerError)
		return
	}

	if tag.RowsAffected() == 0 {
		http.Error(
			w,
			"Listing not found, not owned by you, or not in DRAFT status (published listings cannot be deleted)",
			http.StatusNotFound,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Listing deleted",
	})
}
