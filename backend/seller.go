package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// ============================================================
// SELLER APPLICATION REQUEST
// ============================================================

type SellerApplyRequest struct {
	ShopName    string `json:"shop_name"`
	SellerType  string `json:"seller_type"`
	LegalName   string `json:"legal_name"`
	Address     string `json:"address"`
	Description string `json:"description"`
}

// ============================================================
// SELLER APPLICATION RESPONSE
// ============================================================

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

// ============================================================
// POST /api/seller/apply
// ============================================================

// SellerApplyHandler godoc
// @Summary Apply to become a seller
// @Description Submit a seller application for the authenticated normal user.
// @Tags Seller
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body SellerApplyRequest true "Seller application"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {string} string
// @Failure 401 {string} string
// @Failure 403 {string} string
// @Failure 409 {string} string
// @Router /api/seller/apply [post]
func SellerApplyHandler(w http.ResponseWriter, r *http.Request) {

	// Get logged-in user ID from JWT middleware.
	userID, ok := r.Context().Value(userIDKey).(int64)

	if !ok {
		http.Error(
			w,
			"Unauthorized",
			http.StatusUnauthorized,
		)
		return
	}

	// Read JSON body.
	var request SellerApplyRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&request); err != nil {
		http.Error(
			w,
			"Invalid JSON",
			http.StatusBadRequest,
		)
		return
	}

	// Clean input.
	request.ShopName = strings.TrimSpace(request.ShopName)
	request.SellerType = strings.TrimSpace(request.SellerType)
	request.LegalName = strings.TrimSpace(request.LegalName)
	request.Address = strings.TrimSpace(request.Address)
	request.Description = strings.TrimSpace(request.Description)

	// Required fields.
	if request.ShopName == "" || request.LegalName == "" {
		http.Error(
			w,
			"shop_name and legal_name are required",
			http.StatusBadRequest,
		)
		return
	}

	// Start transaction.
	tx, err := DB.Begin(r.Context())

	if err != nil {
		http.Error(
			w,
			"Could not start database transaction",
			http.StatusInternalServerError,
		)
		return
	}

	defer tx.Rollback(r.Context())

	// ========================================================
	// CHECK USER STATUS
	// ========================================================

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
	).Scan(
		&role,
		&sellerStatus,
	)

	if err != nil {
		http.Error(
			w,
			"User not found",
			http.StatusNotFound,
		)
		return
	}

	// Admins do not submit normal seller applications.
	if role != "USER" {
		http.Error(
			w,
			"Only normal users can apply to become sellers",
			http.StatusForbidden,
		)
		return
	}

	// User must not already have an application.
	if sellerStatus != "NONE" {
		http.Error(
			w,
			"Seller application is not available for this account",
			http.StatusConflict,
		)
		return
	}

	// ========================================================
	// CREATE SELLER PROFILE
	// ========================================================

	var application SellerApplicationResponse

	err = tx.QueryRow(
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
	).Scan(
		&application.SellerID,
		&application.UserID,
		&application.ShopName,
		&application.SellerType,
		&application.LegalName,
		&application.Address,
		&application.Description,
		&application.VerificationStatus,
		&application.AppliedAt,
		&application.ReviewedAt,
		&application.ReviewedBy,
	)

	if err != nil {
		http.Error(
			w,
			"Could not create seller application",
			http.StatusInternalServerError,
		)
		return
	}

	// ========================================================
	// UPDATE USER SELLER STATUS
	// ========================================================

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
		http.Error(
			w,
			"Could not update seller status",
			http.StatusInternalServerError,
		)
		return
	}

	// ========================================================
	// COMMIT
	// ========================================================

	if err := tx.Commit(r.Context()); err != nil {
		http.Error(
			w,
			"Could not complete seller application",
			http.StatusInternalServerError,
		)
		return
	}

	// Return response.
	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(
		map[string]interface{}{
			"message":     "Seller application submitted successfully",
			"application": application,
		},
	)
}

// ============================================================
// GET /api/seller/status
// ============================================================

// SellerStatusHandler godoc
// @Summary Get seller application status
// @Description Return the authenticated user's seller status and application details when available.
// @Tags Seller
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {string} string
// @Failure 404 {string} string
// @Router /api/seller/status [get]
func SellerStatusHandler(w http.ResponseWriter, r *http.Request) {

	// Get logged-in user ID.
	userID, ok := r.Context().Value(userIDKey).(int64)

	if !ok {
		http.Error(
			w,
			"Unauthorized",
			http.StatusUnauthorized,
		)
		return
	}

	var sellerStatus string

	// Get seller status from users table.
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
		http.Error(
			w,
			"User not found",
			http.StatusNotFound,
		)
		return
	}

	// User has never applied.
	if sellerStatus == "NONE" {
		w.Header().Set(
			"Content-Type",
			"application/json",
		)

		json.NewEncoder(w).Encode(
			map[string]interface{}{
				"seller_status": "NONE",
				"application":   nil,
			},
		)

		return
	}

	// ========================================================
	// LOAD SELLER PROFILE
	// ========================================================

	var application SellerApplicationResponse

	err = DB.QueryRow(
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
	).Scan(
		&application.SellerID,
		&application.UserID,
		&application.ShopName,
		&application.SellerType,
		&application.LegalName,
		&application.Address,
		&application.Description,
		&application.VerificationStatus,
		&application.AppliedAt,
		&application.ReviewedAt,
		&application.ReviewedBy,
	)

	if err != nil {
		http.Error(
			w,
			"Seller profile not found",
			http.StatusNotFound,
		)
		return
	}

	// ========================================================
	// RETURN STATUS
	// ========================================================

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	json.NewEncoder(w).Encode(
		map[string]interface{}{
			"seller_status": sellerStatus,
			"application":   application,
		},
	)
}
