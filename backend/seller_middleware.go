package gembid

import (
	"context"
	"net/http"

	"github.com/jackc/pgx/v5"
)

const sellerIDKey contextKey = "seller_id"

func VerifiedSellerMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		userID, ok := r.Context().Value(userIDKey).(int64)

		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var sellerID int64
		err := DB.QueryRow(
			r.Context(),
			`
			SELECT sp.seller_id
			FROM public.users u
			INNER JOIN public.seller_profiles sp ON sp.user_id = u.user_id
			WHERE u.user_id = $1
			  AND u.seller_status = 'VERIFIED'
			  AND sp.verification_status = 'VERIFIED'
			`,
			userID,
		).Scan(&sellerID)

		if err == pgx.ErrNoRows {
			http.Error(w, "Verified seller account required", http.StatusForbidden)
			return
		}

		if err != nil {
			http.Error(w, "Could not verify seller account", http.StatusInternalServerError)
			return
		}

		next(w, r.WithContext(context.WithValue(r.Context(), sellerIDKey, sellerID)))
	}
}
