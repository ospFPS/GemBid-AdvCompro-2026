package main

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
			http.Error(
				w,
				"Unauthorized",
				http.StatusUnauthorized,
			)
			return
		}

		var sellerID int64
		var userSellerStatus string
		var profileStatus string

		err := DB.QueryRow(
			r.Context(),
			`
			SELECT
				sp.seller_id,
				u.seller_status,
				sp.verification_status
			FROM public.users u
			INNER JOIN public.seller_profiles sp
				ON sp.user_id = u.user_id
			WHERE u.user_id = $1
			`,
			userID,
		).Scan(
			&sellerID,
			&userSellerStatus,
			&profileStatus,
		)

		if err == pgx.ErrNoRows {
			http.Error(
				w,
				"Verified seller account required",
				http.StatusForbidden,
			)
			return
		}

		if err != nil {
			http.Error(
				w,
				"Could not verify seller account",
				http.StatusInternalServerError,
			)
			return
		}

		if userSellerStatus != "VERIFIED" ||
			profileStatus != "VERIFIED" {

			http.Error(
				w,
				"Verified seller account required",
				http.StatusForbidden,
			)
			return
		}

		ctx := context.WithValue(
			r.Context(),
			sellerIDKey,
			sellerID,
		)

		next(
			w,
			r.WithContext(ctx),
		)
	}
}
