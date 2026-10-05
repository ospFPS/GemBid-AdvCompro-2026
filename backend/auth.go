package gembid

import (
	"net/http"
	"net/mail"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

var usernameRE = regexp.MustCompile(`^[a-z0-9._-]{3,30}$`)

const userColumns = `user_id, username, full_name, email, phone, role, seller_status, created_at`

// scanUser scans userColumns, then any extra destinations selected after them.
func scanUser(row pgx.Row, extra ...any) (UserResponse, error) {
	var u UserResponse
	err := row.Scan(append([]any{&u.UserID, &u.Username, &u.FullName, &u.Email, &u.Phone,
		&u.Role, &u.SellerStatus, &u.CreatedAt}, extra...)...)
	return u, err
}

type RegisterRequest struct {
	Username string `json:"username"`
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserResponse struct {
	UserID       int64     `json:"user_id"`
	Username     string    `json:"username"`
	FullName     string    `json:"full_name"`
	Email        *string   `json:"email"`
	Phone        *string   `json:"phone"`
	Role         string    `json:"role"`
	SellerStatus string    `json:"seller_status"`
	CreatedAt    time.Time `json:"created_at"`
}

type UpdateProfileRequest struct {
	Username string `json:"username"`
	FullName string `json:"full_name"`
	Phone    string `json:"phone"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var request RegisterRequest

	if !decodeJSON(w, r, &request) {
		return
	}

	request.Username = strings.ToLower(strings.TrimSpace(request.Username))
	request.FullName = strings.TrimSpace(request.FullName)
	request.Email = strings.ToLower(strings.TrimSpace(request.Email))
	request.Phone = strings.TrimSpace(request.Phone)

	if request.Username == "" || request.Password == "" {
		http.Error(w, "username and password are required", http.StatusBadRequest)
		return
	}
	if !usernameRE.MatchString(request.Username) {
		http.Error(w, "username must be 3-30 characters using a-z, 0-9, dot, underscore or hyphen", http.StatusBadRequest)
		return
	}
	if request.FullName == "" {
		request.FullName = request.Username
	}
	if len(request.FullName) > 150 || len(request.Email) > 255 || len(request.Phone) > 30 {
		http.Error(w, "full_name, email or phone is too long", http.StatusBadRequest)
		return
	}
	if request.Email != "" {
		if parsed, err := mail.ParseAddress(request.Email); err != nil || parsed.Address != request.Email {
			http.Error(w, "Invalid email", http.StatusBadRequest)
			return
		}
	}
	if len(request.Password) < 8 || len(request.Password) > 72 {
		http.Error(w, "Password must be 8 to 72 bytes", http.StatusBadRequest)
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte(request.Password),
		bcrypt.DefaultCost,
	)

	if err != nil {
		http.Error(w, "Could not secure password", http.StatusInternalServerError)
		return
	}

	user, err := scanUser(DB.QueryRow(
		r.Context(),
		`
		INSERT INTO public.users (
			username,
			full_name,
			email,
			password_hash,
			phone
		)
		VALUES ($1, $2, NULLIF($3, ''), $4, NULLIF($5, ''))
		RETURNING `+userColumns,
		request.Username,
		request.FullName,
		request.Email,
		string(passwordHash),
		request.Phone,
	))

	if err != nil {
		if isPGError(err, "23505") {
			http.Error(w, "Username or email already registered", http.StatusConflict)
			return
		}
		http.Error(w, "Could not create user", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "User registered successfully",
		"user":    user,
	})
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var request LoginRequest

	if !decodeJSON(w, r, &request) {
		return
	}

	request.Username = strings.ToLower(strings.TrimSpace(request.Username))

	var passwordHash string
	user, err := scanUser(DB.QueryRow(
		r.Context(),
		`SELECT `+userColumns+`, password_hash FROM public.users WHERE username = $1`,
		request.Username,
	), &passwordHash)

	if err == pgx.ErrNoRows {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
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
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	jwtSecret := os.Getenv("JWT_SECRET")

	now := time.Now()

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.MapClaims{
			"sub": strconv.FormatInt(user.UserID, 10),
			"iat": now.Unix(),
			"exp": now.Add(24 * time.Hour).Unix(),
		},
	)

	tokenString, err := token.SignedString([]byte(jwtSecret))

	if err != nil {
		http.Error(w, "Could not create token", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Login successful",
		"token":   tokenString,
		"user":    user,
	})
}

func UpdateProfileHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(userIDKey).(int64)

	var request UpdateProfileRequest
	if !decodeJSON(w, r, &request) {
		return
	}

	request.Username = strings.ToLower(strings.TrimSpace(request.Username))
	request.FullName = strings.TrimSpace(request.FullName)
	request.Phone = strings.TrimSpace(request.Phone)
	if request.FullName == "" {
		http.Error(w, "full_name is required", http.StatusBadRequest)
		return
	}
	if request.Username != "" && !usernameRE.MatchString(request.Username) {
		http.Error(w, "username must be 3-30 characters using a-z, 0-9, dot, underscore or hyphen", http.StatusBadRequest)
		return
	}
	if len(request.FullName) > 150 || len(request.Phone) > 30 {
		http.Error(w, "full_name or phone is too long", http.StatusBadRequest)
		return
	}

	user, err := scanUser(DB.QueryRow(
		r.Context(),
		`
		UPDATE public.users
		SET full_name = $1,
			username = COALESCE(NULLIF($2, ''), username),
			phone = NULLIF($3, ''),
			updated_at = NOW()
		WHERE user_id = $4
		RETURNING `+userColumns,
		request.FullName, request.Username, request.Phone, userID,
	))

	if err == pgx.ErrNoRows {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	if err != nil {
		if isPGError(err, "23505") {
			http.Error(w, "Username already registered", http.StatusConflict)
			return
		}
		http.Error(w, "Could not update profile", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"message": "Profile updated", "user": user})
}

func ChangePasswordHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(userIDKey).(int64)

	var request ChangePasswordRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	if request.CurrentPassword == "" || len(request.NewPassword) < 8 || len(request.NewPassword) > 72 {
		http.Error(w, "current_password is required and new_password must be 8 to 72 bytes", http.StatusBadRequest)
		return
	}

	var passwordHash string
	err := DB.QueryRow(r.Context(), `SELECT password_hash FROM public.users WHERE user_id = $1`, userID).Scan(&passwordHash)
	if err == pgx.ErrNoRows {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(request.CurrentPassword)) != nil {
		http.Error(w, "Current password is incorrect", http.StatusUnauthorized)
		return
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(request.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Could not secure password", http.StatusInternalServerError)
		return
	}

	if _, err = DB.Exec(
		r.Context(),
		`UPDATE public.users SET password_hash = $1, updated_at = NOW() WHERE user_id = $2`,
		string(newHash), userID,
	); err != nil {
		http.Error(w, "Could not update password", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Password changed successfully"})
}
