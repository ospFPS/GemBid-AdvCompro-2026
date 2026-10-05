package gembid

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const userIDKey contextKey = "user_id"

// decodeJSON reads exactly one strict JSON value; on failure it writes 400 and returns false.
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if decoder.Decode(dst) != nil || decoder.Decode(&struct{}{}) != io.EOF {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return false
	}
	return true
}

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		scheme, tokenString, found := strings.Cut(authHeader, " ")
		if !found || !strings.EqualFold(scheme, "Bearer") {
			http.Error(w, "Invalid authorization header", http.StatusUnauthorized)
			return
		}

		jwtSecret := os.Getenv("JWT_SECRET")

		token, err := jwt.Parse(
			tokenString,
			func(_ *jwt.Token) (any, error) { return []byte(jwtSecret), nil },
			jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		)

		if err != nil || !token.Valid {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		subject, err := token.Claims.GetSubject()

		if err != nil {
			http.Error(w, "Invalid token subject", http.StatusUnauthorized)
			return
		}

		userID, err := strconv.ParseInt(subject, 10, 64)

		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusUnauthorized)
			return
		}

		next(w, r.WithContext(context.WithValue(r.Context(), userIDKey, userID)))
	}
}

func MeHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(userIDKey).(int64)

	user, err := scanUser(DB.QueryRow(r.Context(),
		`SELECT `+userColumns+` FROM public.users WHERE user_id = $1`, userID))

	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, user)
}
