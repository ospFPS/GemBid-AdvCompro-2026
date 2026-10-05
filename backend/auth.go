package main

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserResponse struct {
	UserID       int64     `json:"user_id"`
	FullName     string    `json:"full_name"`
	Email        string    `json:"email"`
	Phone        *string   `json:"phone"`
	Role         string    `json:"role"`
	SellerStatus string    `json:"seller_status"`
	CreatedAt    time.Time `json:"created_at"`
}

// RegisterHandler godoc
// @Summary Register a user
// @Description Create a normal GemBid user account. Passwords are hashed before storage.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Registration details"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {string} string
// @Failure 409 {string} string
// @Failure 500 {string} string
// @Router /api/register [post]
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var request RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	request.FullName = strings.TrimSpace(request.FullName)
	request.Email = strings.ToLower(strings.TrimSpace(request.Email))
	request.Phone = strings.TrimSpace(request.Phone)

	if request.FullName == "" ||
		request.Email == "" ||
		request.Password == "" {

		http.Error(
			w,
			"full_name, email and password are required",
			http.StatusBadRequest,
		)
		return
	}

	if len(request.Password) < 8 {
		http.Error(
			w,
			"Password must be at least 8 characters",
			http.StatusBadRequest,
		)
		return
	}

	// Check duplicate email
	var exists bool

	err := DB.QueryRow(
		r.Context(),
		`
		SELECT EXISTS (
			SELECT 1
			FROM public.users
			WHERE email = $1
		)
		`,
		request.Email,
	).Scan(&exists)

	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if exists {
		http.Error(
			w,
			"Email already registered",
			http.StatusConflict,
		)
		return
	}

	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte(request.Password),
		bcrypt.DefaultCost,
	)

	if err != nil {
		http.Error(
			w,
			"Could not secure password",
			http.StatusInternalServerError,
		)
		return
	}

	var user UserResponse

	err = DB.QueryRow(
		r.Context(),
		`
		INSERT INTO public.users (
			full_name,
			email,
			password_hash,
			phone
		)
		VALUES ($1, $2, $3, NULLIF($4, ''))
		RETURNING
			user_id,
			full_name,
			email,
			phone,
			role,
			seller_status,
			created_at
		`,
		request.FullName,
		request.Email,
		string(passwordHash),
		request.Phone,
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
			"Could not create user",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "User registered successfully",
		"user":    user,
	})
}

// LoginHandler godoc
// @Summary Log in
// @Description Authenticate a user and return a JWT bearer token.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {string} string
// @Failure 401 {string} string
// @Failure 500 {string} string
// @Router /api/login [post]
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var request LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	request.Email = strings.ToLower(strings.TrimSpace(request.Email))

	var (
		user         UserResponse
		passwordHash string
	)

	err := DB.QueryRow(
		r.Context(),
		`
		SELECT
			user_id,
			full_name,
			email,
			password_hash,
			phone,
			role,
			seller_status,
			created_at
		FROM public.users
		WHERE email = $1
		`,
		request.Email,
	).Scan(
		&user.UserID,
		&user.FullName,
		&user.Email,
		&passwordHash,
		&user.Phone,
		&user.Role,
		&user.SellerStatus,
		&user.CreatedAt,
	)

	if err == pgx.ErrNoRows {
		http.Error(
			w,
			"Invalid email or password",
			http.StatusUnauthorized,
		)
		return
	}

	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(passwordHash),
		[]byte(request.Password),
	)

	if err != nil {
		http.Error(
			w,
			"Invalid email or password",
			http.StatusUnauthorized,
		)
		return
	}

	jwtSecret := os.Getenv("JWT_SECRET")

	if jwtSecret == "" {
		http.Error(
			w,
			"JWT configuration missing",
			http.StatusInternalServerError,
		)
		return
	}

	now := time.Now()

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.MapClaims{
			"sub":           strconv.FormatInt(user.UserID, 10),
			"user_id":       user.UserID,
			"role":          user.Role,
			"seller_status": user.SellerStatus,
			"iat":           now.Unix(),
			"exp":           now.Add(24 * time.Hour).Unix(),
		},
	)

	tokenString, err := token.SignedString([]byte(jwtSecret))

	if err != nil {
		http.Error(
			w,
			"Could not create token",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Login successful",
		"token":   tokenString,
		"user":    user,
	})
}
