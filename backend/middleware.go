package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const userIDKey contextKey = "user_id"

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			http.Error(
				w,
				"Authorization header required",
				http.StatusUnauthorized,
			)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)

		if len(parts) != 2 ||
			!strings.EqualFold(parts[0], "Bearer") {

			http.Error(
				w,
				"Invalid authorization header",
				http.StatusUnauthorized,
			)
			return
		}

		tokenString := parts[1]

		jwtSecret := os.Getenv("JWT_SECRET")

		token, err := jwt.Parse(
			tokenString,
			func(token *jwt.Token) (interface{}, error) {
				return []byte(jwtSecret), nil
			},
			jwt.WithValidMethods(
				[]string{jwt.SigningMethodHS256.Alg()},
			),
		)

		if err != nil || !token.Valid {
			http.Error(
				w,
				"Invalid or expired token",
				http.StatusUnauthorized,
			)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)

		if !ok {
			http.Error(
				w,
				"Invalid token",
				http.StatusUnauthorized,
			)
			return
		}

		subject, err := claims.GetSubject()

		if err != nil {
			http.Error(
				w,
				"Invalid token subject",
				http.StatusUnauthorized,
			)
			return
		}

		userID, err := strconv.ParseInt(subject, 10, 64)

		if err != nil {
			http.Error(
				w,
				"Invalid user ID",
				http.StatusUnauthorized,
			)
			return
		}

		ctx := context.WithValue(
			r.Context(),
			userIDKey,
			userID,
		)

		next(w, r.WithContext(ctx))
	}
}

func MeHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(userIDKey).(int64)

	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var user UserResponse

	err := DB.QueryRow(
		r.Context(),
		`
		SELECT
			user_id,
			full_name,
			email,
			phone,
			role,
			seller_status,
			created_at
		FROM public.users
		WHERE user_id = $1
		`,
		userID,
	).Scan(
		&user.UserID,
		&user.FullName,
		&user.Email,
		&user.Phone,
		&user.Role,
		&user.SellerStatus,
		&user.CreatedAt,
	)

	if err != nil {
		http.Error(
			w,
			"User not found",
			http.StatusNotFound,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}