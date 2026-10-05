package main

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

// Monetization: sellers on FREE can have a limited number of live listings.
// PRO is a 30-day pass paid with a mock payment; paying again extends it.
const (
	proPriceTHB              = 299.00
	proPeriodDays            = 30
	freeLiveListingLimit     = 5
	activeProSubscriptionSQL = `
		plan = 'PRO' AND status = 'ACTIVE'
		AND (end_date IS NULL OR end_date >= CURRENT_DATE)`
)

type SubscriptionPlan struct {
	Plan             string   `json:"plan"`
	PriceTHB         float64  `json:"price_thb"`
	PeriodDays       int      `json:"period_days,omitempty"`
	LiveListingLimit int      `json:"live_listing_limit,omitempty"`
	Features         []string `json:"features"`
}

type SubscriptionResponse struct {
	SubscriptionID int64      `json:"subscription_id"`
	Plan           string     `json:"plan"`
	Status         string     `json:"status"`
	StartDate      time.Time  `json:"start_date"`
	EndDate        *time.Time `json:"end_date"`
	CreatedAt      time.Time  `json:"created_at"`
}

type UpgradeSubscriptionRequest struct {
	PaymentMethod string `json:"payment_method" example:"MOCK_CARD"`
}

var plans = []SubscriptionPlan{
	{Plan: "FREE", PriceTHB: 0, LiveListingLimit: freeLiveListingLimit,
		Features: []string{"Browse, bid and buy", "Up to " + strconv.Itoa(freeLiveListingLimit) + " live listings as a seller"}},
	{Plan: "PRO", PriceTHB: proPriceTHB, PeriodDays: proPeriodDays,
		Features: []string{"Everything in FREE", "Unlimited live listings"}},
}

func scanSubscription(row pgx.Row) (SubscriptionResponse, error) {
	var s SubscriptionResponse
	err := row.Scan(&s.SubscriptionID, &s.Plan, &s.Status, &s.StartDate, &s.EndDate, &s.CreatedAt)
	return s, err
}

// checkListingQuota returns a non-empty message when a FREE seller is at the live listing limit.
// ponytail: count-then-insert is not race-proof; two simultaneous publishes can exceed the limit by one.
func checkListingQuota(ctx context.Context, q interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}, sellerID int64) (string, error) {
	var live int
	var isPro bool
	err := q.QueryRow(
		ctx,
		`
		SELECT
			(SELECT count(*) FROM public.listings
			 WHERE seller_id = $1 AND status IN ('PUBLISHED', 'ACTIVE')),
			EXISTS (
				SELECT 1 FROM public.subscriptions s
				INNER JOIN public.seller_profiles sp ON sp.user_id = s.user_id
				WHERE sp.seller_id = $1 AND `+activeProSubscriptionSQL+`
			)
		`,
		sellerID,
	).Scan(&live, &isPro)
	if err != nil || isPro || live < freeLiveListingLimit {
		return "", err
	}
	return "FREE plan allows " + strconv.Itoa(freeLiveListingLimit) +
		" live listings. Upgrade to PRO for unlimited listings.", nil
}

// ListPlansHandler godoc
// @Summary List subscription plans
// @Tags Subscriptions
// @Produce json
// @Success 200 {array} SubscriptionPlan
// @Router /api/subscriptions/plans [get]
func ListPlansHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, plans)
}

// MySubscriptionHandler godoc
// @Summary Get my current plan
// @Tags Subscriptions
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {string} string
// @Router /api/subscriptions/me [get]
func MySubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(userIDKey).(int64)

	s, err := scanSubscription(DB.QueryRow(
		r.Context(),
		`
		SELECT subscription_id, plan, status, start_date, end_date, created_at
		FROM public.subscriptions
		WHERE user_id = $1 AND `+activeProSubscriptionSQL,
		userID,
	))

	if err == pgx.ErrNoRows {
		writeJSON(w, http.StatusOK, map[string]any{"plan": "FREE", "subscription": nil})
		return
	}
	if err != nil {
		http.Error(w, "Could not load subscription", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"plan": "PRO", "subscription": s})
}

// UpgradeSubscriptionHandler godoc
// @Summary Upgrade to PRO (mock payment)
// @Description Pays for 30 days of PRO. If PRO is already active, the end date is extended by 30 days.
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body UpgradeSubscriptionRequest true "MOCK_CARD or MOCK_QR"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {string} string
// @Failure 401 {string} string
// @Router /api/subscriptions/upgrade [post]
func UpgradeSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(userIDKey).(int64)
	ctx := r.Context()

	var request UpgradeSubscriptionRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	request.PaymentMethod = strings.ToUpper(strings.TrimSpace(request.PaymentMethod))
	if !validPaymentMethods[request.PaymentMethod] {
		http.Error(w, "payment_method must be MOCK_CARD or MOCK_QR", http.StatusBadRequest)
		return
	}

	tx, err := DB.Begin(ctx)
	if err != nil {
		http.Error(w, "Could not start subscription transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(ctx)

	// Serialize upgrades per user.
	if _, err := tx.Exec(ctx, `SELECT 1 FROM public.users WHERE user_id = $1 FOR UPDATE`, userID); err != nil {
		http.Error(w, "Could not lock user", http.StatusInternalServerError)
		return
	}

	// Lazily expire a lapsed PRO so the one-ACTIVE-per-user index allows a new one.
	if _, err := tx.Exec(ctx,
		`UPDATE public.subscriptions SET status = 'EXPIRED'
		 WHERE user_id = $1 AND status = 'ACTIVE' AND end_date < CURRENT_DATE`, userID); err != nil {
		http.Error(w, "Could not update subscriptions", http.StatusInternalServerError)
		return
	}

	const returning = ` RETURNING subscription_id, plan, status, start_date, end_date, created_at`

	sub, err := scanSubscription(tx.QueryRow(ctx,
		`UPDATE public.subscriptions SET end_date = end_date + $2::int
		 WHERE user_id = $1 AND status = 'ACTIVE' AND plan = 'PRO'`+returning,
		userID, proPeriodDays))
	if err == pgx.ErrNoRows {
		sub, err = scanSubscription(tx.QueryRow(ctx,
			`INSERT INTO public.subscriptions (user_id, plan, status, start_date, end_date)
			 VALUES ($1, 'PRO', 'ACTIVE', CURRENT_DATE, CURRENT_DATE + $2::int)`+returning,
			userID, proPeriodDays))
	}
	if err != nil {
		http.Error(w, "Could not save subscription", http.StatusInternalServerError)
		return
	}

	var t TransactionResponse
	err = tx.QueryRow(
		ctx,
		`
		INSERT INTO public.transactions
			(subscription_id, buyer_id, amount, payment_method, status, reference_number, paid_at)
		VALUES ($1, $2, $3, $4, 'SUCCESS', $5, NOW())
		RETURNING transaction_id, order_id, subscription_id, amount, payment_method, status,
			reference_number, created_at, paid_at
		`,
		sub.SubscriptionID, userID, proPriceTHB, request.PaymentMethod, mockPayment(),
	).Scan(&t.TransactionID, &t.OrderID, &t.SubscriptionID, &t.Amount, &t.PaymentMethod,
		&t.Status, &t.ReferenceNumber, &t.CreatedAt, &t.PaidAt)
	if err != nil {
		http.Error(w, "Could not record payment", http.StatusInternalServerError)
		return
	}

	if err := notify(ctx, tx, userID, "SUBSCRIPTION_ACTIVATED", "GemBid PRO activated",
		"Your PRO plan is active until "+sub.EndDate.Format("2 Jan 2006")+".", 0, 0, 0); err != nil {
		http.Error(w, "Could not create notification", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(ctx); err != nil {
		http.Error(w, "Could not commit subscription", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"message":      "PRO plan active",
		"subscription": sub,
		"transaction":  t,
	})
}
