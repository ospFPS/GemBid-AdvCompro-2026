package gembid

import (
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

type SellerApplyRequest struct {
	ShopName    string `json:"shop_name"`
	SellerType  string `json:"seller_type"`
	LegalName   string `json:"legal_name"`
	Address     string `json:"address"`
	Description string `json:"description"`
}

type SellerApplicationResponse struct {
	SellerID           int64      `json:"seller_id"`
	UserID             int64      `json:"user_id"`
	ShopName           string     `json:"shop_name"`
	SellerType         *string    `json:"seller_type"`
	LegalName          string     `json:"legal_name"`
	Address            *string    `json:"address"`
	Description        *string    `json:"description"`
	VerificationStatus string     `json:"verification_status"`
	AppliedAt          time.Time  `json:"applied_at"`
	ReviewedAt         *time.Time `json:"reviewed_at,omitempty"`
	ReviewedBy         *int64     `json:"reviewed_by,omitempty"`
}

func scanSellerApplication(row pgx.Row, application *SellerApplicationResponse) error {
	return row.Scan(
		&application.SellerID, &application.UserID, &application.ShopName,
		&application.SellerType, &application.LegalName, &application.Address,
		&application.Description, &application.VerificationStatus, &application.AppliedAt,
		&application.ReviewedAt, &application.ReviewedBy,
	)
}

func SellerApplyHandler(w http.ResponseWriter, r *http.Request) {

	userID, _ := r.Context().Value(userIDKey).(int64)

	var request SellerApplyRequest

	if !decodeJSON(w, r, &request) {
		return
	}

	request.ShopName = strings.TrimSpace(request.ShopName)
	request.SellerType = strings.TrimSpace(request.SellerType)
	request.LegalName = strings.TrimSpace(request.LegalName)
	request.Address = strings.TrimSpace(request.Address)
	request.Description = strings.TrimSpace(request.Description)

	if request.ShopName == "" || request.LegalName == "" {
		http.Error(w, "shop_name and legal_name are required", http.StatusBadRequest)
		return
	}
	if len(request.ShopName) > 150 || len(request.SellerType) > 50 || len(request.LegalName) > 150 ||
		len(request.Address) > 2000 || len(request.Description) > 5000 {
		http.Error(w, "seller application contains a field that is too long", http.StatusBadRequest)
		return
	}

	tx, err := DB.Begin(r.Context())

	if err != nil {
		http.Error(w, "Could not start database transaction", http.StatusInternalServerError)
		return
	}

	defer tx.Rollback(r.Context())

	var role string
	var sellerStatus string

	err = tx.QueryRow(
		r.Context(),
		`
		SELECT
			role,
			seller_status
		FROM public.users
		WHERE user_id = $1
		FOR UPDATE
		`,
		userID,
	).Scan(&role, &sellerStatus)

	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if role != "USER" {
		http.Error(w, "Only normal users can apply to become sellers", http.StatusForbidden)
		return
	}

	if sellerStatus != "NONE" {
		http.Error(w, "Seller application is not available for this account", http.StatusConflict)
		return
	}

	var application SellerApplicationResponse

	err = scanSellerApplication(tx.QueryRow(
		r.Context(),
		`
		INSERT INTO public.seller_profiles (
			user_id,
			shop_name,
			seller_type,
			legal_name,
			address,
			description
		)
		VALUES (
			$1,
			$2,
			NULLIF($3, ''),
			$4,
			NULLIF($5, ''),
			NULLIF($6, '')
		)
		RETURNING
			seller_id,
			user_id,
			shop_name,
			seller_type,
			legal_name,
			address,
			description,
			verification_status,
			applied_at,
			reviewed_at,
			reviewed_by
		`,
		userID,
		request.ShopName,
		request.SellerType,
		request.LegalName,
		request.Address,
		request.Description,
	), &application)

	if err != nil {
		http.Error(w, "Could not create seller application", http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec(
		r.Context(),
		`
		UPDATE public.users
		SET seller_status = 'PENDING'
		WHERE user_id = $1
		`,
		userID,
	)

	if err != nil {
		http.Error(w, "Could not update seller status", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(r.Context()); err != nil {
		http.Error(w, "Could not complete seller application", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"message":     "Seller application submitted successfully",
		"application": application,
	})
}

func SellerStatusHandler(w http.ResponseWriter, r *http.Request) {

	userID, _ := r.Context().Value(userIDKey).(int64)

	var sellerStatus string

	err := DB.QueryRow(
		r.Context(),
		`
		SELECT seller_status
		FROM public.users
		WHERE user_id = $1
		`,
		userID,
	).Scan(&sellerStatus)

	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if sellerStatus == "NONE" {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"seller_status": "NONE",
			"application":   nil,
		})

		return
	}

	var application SellerApplicationResponse

	err = scanSellerApplication(DB.QueryRow(
		r.Context(),
		`
		SELECT
			seller_id,
			user_id,
			shop_name,
			seller_type,
			legal_name,
			address,
			description,
			verification_status,
			applied_at,
			reviewed_at,
			reviewed_by
		FROM public.seller_profiles
		WHERE user_id = $1
		`,
		userID,
	), &application)

	if err != nil {
		http.Error(w, "Seller profile not found", http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"seller_status": sellerStatus,
		"application":   application,
	})
}
