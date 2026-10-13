package gembid

import (
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
)

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

type CreateListingRequest struct {
	GemID int64 `json:"gem_id"`
}

func scanListing(row pgx.Row, listing *ListingResponse) error {
	return row.Scan(
		&listing.ListingID, &listing.GemID, &listing.SellerID, &listing.SaleType,
		&listing.Status, &listing.FixedPrice, &listing.CreatedAt, &listing.PublishedAt,
		&listing.UpdatedAt, &listing.RelistedFromListingID,
	)
}

func CreateListingHandler(w http.ResponseWriter, r *http.Request) {
	sellerID, _ := r.Context().Value(sellerIDKey).(int64)

	var request CreateListingRequest

	if !decodeJSON(w, r, &request) {
		return
	}

	if request.GemID == 0 {
		http.Error(w, "gem_id is required", http.StatusBadRequest)
		return
	}

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

	err = scanListing(DB.QueryRow(
		r.Context(),
		`
		INSERT INTO public.listings (gem_id, seller_id)
		VALUES ($1, $2)
		RETURNING
			listing_id, gem_id, seller_id, sale_type, status, fixed_price,
			created_at, published_at, updated_at, relisted_from_listing_id
		`,
		request.GemID, sellerID,
	), &listing)

	if err != nil {
		http.Error(w, "Could not create listing", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "Draft listing created",
		"listing": listing,
	})
}

type ListingListItem struct {
	ListingID        int64      `json:"listing_id"`
	GemID            int64      `json:"gem_id"`
	GemName          string     `json:"gem_name"`
	GemType          string     `json:"gem_type"`
	SaleType         *string    `json:"sale_type"`
	Status           string     `json:"status"`
	FixedPrice       *float64   `json:"fixed_price"`
	Description      *string    `json:"description"`
	SellerID         int64      `json:"seller_id"`
	ShopName         string     `json:"shop_name"`
	AuctionID        *int64     `json:"auction_id"`
	CurrentPrice     *float64   `json:"current_price"`
	MinimumIncrement *float64   `json:"minimum_increment"`
	EndTime          *time.Time `json:"end_time"`
	ImageURL         *string    `json:"image_url"`
}

// listListings returns listings with gem, shop and auction details, filtered by where.
func listListings(w http.ResponseWriter, r *http.Request, where string, args ...any) {
	rows, err := DB.Query(
		r.Context(),
		`
		SELECT
			l.listing_id, l.gem_id, g.gem_name, g.gem_type, l.sale_type, l.status, l.fixed_price,
			g.description, l.seller_id, sp.shop_name,
			a.auction_id, a.current_price, a.minimum_increment, a.end_time, gi.image_url
		FROM public.listings l
		INNER JOIN public.gems g ON g.gem_id = l.gem_id
		INNER JOIN public.seller_profiles sp ON sp.seller_id = l.seller_id
		LEFT JOIN public.auctions a ON a.listing_id = l.listing_id
		LEFT JOIN public.gem_images gi ON gi.gem_id = l.gem_id AND gi.is_primary
		WHERE `+where+`
		ORDER BY l.published_at DESC NULLS LAST, l.created_at DESC
		`,
		args...,
	)
	if err != nil {
		http.Error(w, "Could not fetch listings", http.StatusInternalServerError)
		return
	}
	listings, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (ListingListItem, error) {
		var item ListingListItem
		err := row.Scan(
			&item.ListingID, &item.GemID, &item.GemName, &item.GemType,
			&item.SaleType, &item.Status, &item.FixedPrice,
			&item.Description, &item.SellerID, &item.ShopName,
			&item.AuctionID, &item.CurrentPrice, &item.MinimumIncrement, &item.EndTime, &item.ImageURL,
		)
		return item, err
	})
	if err != nil {
		http.Error(w, "Could not read listing data", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, listings)
}

func ListListingsHandler(w http.ResponseWriter, r *http.Request) {
	listListings(w, r, `l.status IN ('PUBLISHED', 'ACTIVE')`)
}

// MySellerListingsHandler returns the verified seller's own listings in every status, drafts included.
func MySellerListingsHandler(w http.ResponseWriter, r *http.Request) {
	sellerID, _ := r.Context().Value(sellerIDKey).(int64)
	listListings(w, r, `l.seller_id = $1`, sellerID)
}

func GetListingHandler(w http.ResponseWriter, r *http.Request) {
	listingID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)

	if err != nil {
		http.Error(w, "Invalid listing ID", http.StatusBadRequest)
		return
	}

	var listing ListingResponse

	err = scanListing(DB.QueryRow(
		r.Context(),
		`
		SELECT
			listing_id, gem_id, seller_id, sale_type, status, fixed_price,
			created_at, published_at, updated_at, relisted_from_listing_id
		FROM public.listings
		WHERE listing_id = $1
		`,
		listingID,
	), &listing)

	if err == pgx.ErrNoRows {
		http.Error(w, "Listing not found", http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, "Could not fetch listing", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, listing)
}

type PublishListingRequest struct {
	SaleType   string   `json:"sale_type"`
	FixedPrice *float64 `json:"fixed_price"`
}

func PublishListingHandler(w http.ResponseWriter, r *http.Request) {
	sellerID, _ := r.Context().Value(sellerIDKey).(int64)

	listingID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)

	if err != nil {
		http.Error(w, "Invalid listing ID", http.StatusBadRequest)
		return
	}

	var request PublishListingRequest

	if !decodeJSON(w, r, &request) {
		return
	}

	var listing ListingResponse

	switch request.SaleType {

	case "FIXED_PRICE":
		if request.FixedPrice == nil || *request.FixedPrice < 0 || *request.FixedPrice > 9999999999.99 {
			http.Error(w, "fixed_price must be between 0 and 9999999999.99", http.StatusBadRequest)
			return
		}

		tx, txErr := DB.Begin(r.Context())
		if txErr != nil {
			http.Error(w, "Could not start listing transaction", http.StatusInternalServerError)
			return
		}
		defer tx.Rollback(r.Context())

		if _, txErr = tx.Exec(r.Context(),
			`SELECT 1 FROM public.seller_profiles WHERE seller_id = $1 FOR UPDATE`,
			sellerID,
		); txErr != nil {
			http.Error(w, "Could not lock seller account", http.StatusInternalServerError)
			return
		}
		if msg, txErr := checkListingQuota(r.Context(), tx, sellerID); txErr != nil {
			http.Error(w, "Could not check listing quota", http.StatusInternalServerError)
			return
		} else if msg != "" {
			http.Error(w, msg, http.StatusForbidden)
			return
		}

		err = scanListing(tx.QueryRow(
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
		), &listing)
		if err == nil {
			err = tx.Commit(r.Context())
		}

	case "AUCTION":
		if request.FixedPrice != nil {
			http.Error(w, "fixed_price must not be set for AUCTION listings", http.StatusBadRequest)
			return
		}

		err = scanListing(DB.QueryRow(
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
		), &listing)

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

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Listing updated",
		"listing": listing,
	})
}

func DeleteListingHandler(w http.ResponseWriter, r *http.Request) {
	sellerID, _ := r.Context().Value(sellerIDKey).(int64)

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
		if isPGError(err, "23503") {
			http.Error(w, "Cannot delete: this listing is referenced elsewhere", http.StatusConflict)
			return
		}

		http.Error(w, "Could not delete listing", http.StatusInternalServerError)
		return
	}

	if tag.RowsAffected() == 0 {
		http.Error(w, "Listing not found, not owned by you, or not in DRAFT status (published listings cannot be deleted)", http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Listing deleted",
	})
}
