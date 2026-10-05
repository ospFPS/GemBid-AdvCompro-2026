package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

// ============================================================
// ADMIN SELLER APPLICATION RESPONSE
// ============================================================

type AdminSellerApplication struct {
	SellerID           int64      `json:"seller_id"`
	UserID             int64      `json:"user_id"`
	FullName           string     `json:"full_name"`
	Email              string     `json:"email"`
	ShopName           string     `json:"shop_name"`
	SellerType         *string    `json:"seller_type"`
	LegalName          string     `json:"legal_name"`
	Address            *string    `json:"address"`
	Description        *string    `json:"description"`
	VerificationStatus string     `json:"verification_status"`
	AppliedAt          time.Time  `json:"applied_at"`
	ReviewedAt         *time.Time `json:"reviewed_at"`
	ReviewedBy         *int64     `json:"reviewed_by"`
}

// ============================================================
// ADMIN REVIEW REQUEST
// ============================================================

type SellerReviewRequest struct {
	Decision string `json:"decision"`
}

// ============================================================
// ADMIN MIDDLEWARE
// ============================================================

// AdminMiddleware checks that the logged-in user is actually an ADMIN.
//
// AuthMiddleware must run before this middleware because
// AuthMiddleware places user_id into the request context.
func AdminMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Get logged-in user ID from JWT context.
		userID, ok := r.Context().Value(userIDKey).(int64)

		if !ok {
			http.Error(
				w,
				"Unauthorized",
				http.StatusUnauthorized,
			)
			return
		}

		var role string

		// Load the current role directly from the database.
		// We do not rely only on the role stored inside the JWT.
		err := DB.QueryRow(
			r.Context(),
			`
			SELECT role
			FROM public.users
			WHERE user_id = $1
			`,
			userID,
		).Scan(&role)

		if err == pgx.ErrNoRows {
			http.Error(
				w,
				"User not found",
				http.StatusUnauthorized,
			)
			return
		}

		if err != nil {
			http.Error(
				w,
				"Could not verify admin account",
				http.StatusInternalServerError,
			)
			return
		}

		// Only ADMIN accounts can continue.
		if role != "ADMIN" {
			http.Error(
				w,
				"Admin access required",
				http.StatusForbidden,
			)
			return
		}

		// Admin verified.
		next(w, r)
	}
}

// ============================================================
// GET /api/admin/seller-applications
// ============================================================

// AdminSellerApplicationsHandler returns seller applications
// for the admin dashboard.
// AdminSellerApplicationsHandler godoc
// @Summary List seller applications
// @Description Admin-only endpoint to view seller applications.
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {string} string
// @Failure 403 {string} string
// @Failure 500 {string} string
// @Router /api/admin/seller-applications [get]
func AdminSellerApplicationsHandler(
	w http.ResponseWriter,
	r *http.Request,
) {

	rows, err := DB.Query(
		r.Context(),
		`
		SELECT
			sp.seller_id,
			sp.user_id,
			u.full_name,
			u.email,
			sp.shop_name,
			sp.seller_type,
			sp.legal_name,
			sp.address,
			sp.description,
			sp.verification_status,
			sp.applied_at,
			sp.reviewed_at,
			sp.reviewed_by
		FROM public.seller_profiles sp
		INNER JOIN public.users u
			ON u.user_id = sp.user_id
		ORDER BY sp.applied_at DESC
		`,
	)

	if err != nil {
		http.Error(
			w,
			"Could not load seller applications",
			http.StatusInternalServerError,
		)
		return
	}

	defer rows.Close()

	applications := make([]AdminSellerApplication, 0)

	for rows.Next() {

		var application AdminSellerApplication

		err := rows.Scan(
			&application.SellerID,
			&application.UserID,
			&application.FullName,
			&application.Email,
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
				"Could not read seller applications",
				http.StatusInternalServerError,
			)
			return
		}

		applications = append(
			applications,
			application,
		)
	}

	// Check if PostgreSQL produced an error while reading rows.
	if err := rows.Err(); err != nil {
		http.Error(
			w,
			"Could not finish reading seller applications",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	json.NewEncoder(w).Encode(
		map[string]interface{}{
			"applications": applications,
			"count":        len(applications),
		},
	)
}

// ============================================================
// PUT /api/admin/seller-applications/{id}
// ============================================================

// AdminReviewSellerHandler allows an admin to:
//
// APPROVE -> VERIFIED
//
// or
//
// REJECT -> REJECTED
// AdminReviewSellerHandler godoc
// @Summary Review a seller application
// @Description Admin approves or rejects a pending seller application.
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Seller application ID"
// @Param request body SellerReviewRequest true "APPROVE or REJECT"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {string} string
// @Failure 401 {string} string
// @Failure 403 {string} string
// @Failure 404 {string} string
// @Failure 409 {string} string
// @Router /api/admin/seller-applications/{id} [put]
func AdminReviewSellerHandler(
	w http.ResponseWriter,
	r *http.Request,
) {

	// --------------------------------------------------------
	// Get admin user ID
	// --------------------------------------------------------

	adminUserID, ok := r.Context().Value(userIDKey).(int64)

	if !ok {
		http.Error(
			w,
			"Unauthorized",
			http.StatusUnauthorized,
		)
		return
	}

	// --------------------------------------------------------
	// Read seller application ID from URL
	// --------------------------------------------------------

	sellerIDText := r.PathValue("id")

	sellerID, err := strconv.ParseInt(
		sellerIDText,
		10,
		64,
	)

	if err != nil || sellerID <= 0 {
		http.Error(
			w,
			"Invalid seller application ID",
			http.StatusBadRequest,
		)
		return
	}

	// --------------------------------------------------------
	// Read JSON request
	// --------------------------------------------------------

	var request SellerReviewRequest

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

	request.Decision = strings.ToUpper(
		strings.TrimSpace(request.Decision),
	)

	// --------------------------------------------------------
	// Convert decision into database status
	// --------------------------------------------------------

	var newStatus string

	switch request.Decision {

	case "APPROVE":
		newStatus = "VERIFIED"

	case "REJECT":
		newStatus = "REJECTED"

	default:
		http.Error(
			w,
			"decision must be APPROVE or REJECT",
			http.StatusBadRequest,
		)
		return
	}

	// --------------------------------------------------------
	// Begin transaction
	// --------------------------------------------------------

	tx, err := DB.Begin(r.Context())

	if err != nil {
		http.Error(
			w,
			"Could not start database transaction",
			http.StatusInternalServerError,
		)
		return
	}

	// Rollback safely does nothing after successful Commit.
	defer tx.Rollback(r.Context())

	// --------------------------------------------------------
	// Load and lock seller application
	// --------------------------------------------------------

	var sellerUserID int64
	var currentStatus string

	err = tx.QueryRow(
		r.Context(),
		`
		SELECT
			user_id,
			verification_status
		FROM public.seller_profiles
		WHERE seller_id = $1
		FOR UPDATE
		`,
		sellerID,
	).Scan(
		&sellerUserID,
		&currentStatus,
	)

	if err == pgx.ErrNoRows {
		http.Error(
			w,
			"Seller application not found",
			http.StatusNotFound,
		)
		return
	}

	if err != nil {
		http.Error(
			w,
			"Could not load seller application",
			http.StatusInternalServerError,
		)
		return
	}

	// --------------------------------------------------------
	// Only PENDING applications may be reviewed
	// --------------------------------------------------------

	if currentStatus != "PENDING" {
		http.Error(
			w,
			"Seller application has already been reviewed",
			http.StatusConflict,
		)
		return
	}

	// --------------------------------------------------------
	// Update seller_profiles
	// --------------------------------------------------------

	_, err = tx.Exec(
		r.Context(),
		`
		UPDATE public.seller_profiles
		SET
			verification_status = $1,
			reviewed_at = NOW(),
			reviewed_by = $2
		WHERE seller_id = $3
		`,
		newStatus,
		adminUserID,
		sellerID,
	)

	if err != nil {
		http.Error(
			w,
			"Could not update seller application",
			http.StatusInternalServerError,
		)
		return
	}

	// --------------------------------------------------------
	// Update users.seller_status
	// --------------------------------------------------------

	_, err = tx.Exec(
		r.Context(),
		`
		UPDATE public.users
		SET seller_status = $1
		WHERE user_id = $2
		`,
		newStatus,
		sellerUserID,
	)

	if err != nil {
		http.Error(
			w,
			"Could not update user seller status",
			http.StatusInternalServerError,
		)
		return
	}

	// --------------------------------------------------------
	// Commit transaction
	// --------------------------------------------------------

	if err := tx.Commit(r.Context()); err != nil {
		http.Error(
			w,
			"Could not complete seller review",
			http.StatusInternalServerError,
		)
		return
	}

	// --------------------------------------------------------
	// Return response
	// --------------------------------------------------------

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	json.NewEncoder(w).Encode(
		map[string]interface{}{
			"message":             "Seller application reviewed successfully",
			"seller_id":           sellerID,
			"user_id":             sellerUserID,
			"decision":            request.Decision,
			"verification_status": newStatus,
			"reviewed_by":         adminUserID,
		},
	)
}

// ============================================================
// GET /api/admin/analytics
// ============================================================

// AdminAnalyticsHandler godoc
// @Summary Platform analytics
// @Description Admin dashboard numbers: users, sellers, listings, auctions, bids, orders, sales (GMV), subscription revenue, email delivery and 30-day daily sales.
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {string} string
// @Failure 403 {string} string
// @Failure 500 {string} string
// @Router /api/admin/analytics [get]
func AdminAnalyticsHandler(w http.ResponseWriter, r *http.Request) {
	var report []byte

	// One round trip; PostgreSQL builds the JSON.
	err := DB.QueryRow(
		r.Context(),
		`
		SELECT json_build_object(
			'generated_at', NOW(),
			'users', (
				SELECT json_build_object(
					'total', count(*),
					'admins', count(*) FILTER (WHERE role = 'ADMIN'),
					'new_last_30_days', count(*) FILTER (WHERE created_at >= NOW() - INTERVAL '30 days'),
					'sellers_by_status', (
						SELECT COALESCE(json_object_agg(seller_status, n), '{}')
						FROM (SELECT seller_status, count(*) n FROM public.users GROUP BY seller_status) s
					)
				)
				FROM public.users
			),
			'gems_total', (SELECT count(*) FROM public.gems),
			'listings_by_status', (
				SELECT COALESCE(json_object_agg(status, n), '{}')
				FROM (SELECT status, count(*) n FROM public.listings GROUP BY status) s
			),
			'auctions_by_status', (
				SELECT COALESCE(json_object_agg(status, n), '{}')
				FROM (SELECT status, count(*) n FROM public.auctions GROUP BY status) s
			),
			'bids_total', (SELECT count(*) FROM public.bids),
			'orders_by_status', (
				SELECT COALESCE(json_object_agg(status, n), '{}')
				FROM (SELECT status, count(*) n FROM public.orders GROUP BY status) s
			),
			'sales', (
				SELECT json_build_object(
					'paid_orders', count(*),
					'gross_merchandise_value', COALESCE(sum(amount), 0),
					'average_order_value', COALESCE(round(avg(amount), 2), 0)
				)
				FROM public.transactions
				WHERE order_id IS NOT NULL AND status = 'SUCCESS'
			),
			'subscriptions', json_build_object(
				'active_pro', (
					SELECT count(*) FROM public.subscriptions WHERE `+activeProSubscriptionSQL+`
				),
				'revenue', (
					SELECT COALESCE(sum(amount), 0) FROM public.transactions
					WHERE subscription_id IS NOT NULL AND status = 'SUCCESS'
				)
			),
			'emails_by_status', (
				SELECT COALESCE(json_object_agg(status, n), '{}')
				FROM (SELECT status, count(*) n FROM public.email_deliveries GROUP BY status) s
			),
			'top_gem_types_sold', (
				SELECT COALESCE(json_agg(t), '[]')
				FROM (
					SELECT g.gem_type, count(*) AS orders, sum(o.total_amount) AS revenue
					FROM public.orders o
					INNER JOIN public.listings l ON l.listing_id = o.listing_id
					INNER JOIN public.gems g ON g.gem_id = l.gem_id
					WHERE o.status IN ('PAID', 'SHIPPED', 'DELIVERED')
					GROUP BY g.gem_type
					ORDER BY revenue DESC
					LIMIT 5
				) t
			),
			'daily_sales_last_30_days', (
				SELECT COALESCE(json_agg(d ORDER BY d.day), '[]')
				FROM (
					SELECT paid_at::date AS day, count(*) AS orders, sum(amount) AS revenue
					FROM public.transactions
					WHERE order_id IS NOT NULL AND status = 'SUCCESS'
					  AND paid_at >= NOW() - INTERVAL '30 days'
					GROUP BY paid_at::date
				) d
			)
		)
		`,
	).Scan(&report)
	if err != nil {
		http.Error(w, "Could not load analytics", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(report)
}
