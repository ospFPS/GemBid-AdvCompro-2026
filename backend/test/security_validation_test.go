package test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	. "adcompro-backend"
	"github.com/golang-jwt/jwt/v5"
)

func TestSecurityValidationGuards(t *testing.T) {
	t.Run("JWT accepts only valid HS256 unexpired tokens", func(t *testing.T) {
		t.Setenv("JWT_SECRET", "security-test-secret")
		handler := AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})

		request := func(token string) int {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if token != "" {
				req.Header.Set("Authorization", "Bearer "+token)
			}
			rec := httptest.NewRecorder()
			handler(rec, req)
			return rec.Code
		}

		if got := request(""); got != http.StatusUnauthorized {
			t.Fatalf("missing token: got %d", got)
		}

		valid := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub": "42",
			"exp": time.Now().Add(time.Hour).Unix(),
		})
		validString, err := valid.SignedString([]byte("security-test-secret"))
		if err != nil {
			t.Fatal(err)
		}
		if got := request(validString); got != http.StatusNoContent {
			t.Fatalf("valid token: got %d", got)
		}

		expired := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub": "42",
			"exp": time.Now().Add(-time.Hour).Unix(),
		})
		expiredString, _ := expired.SignedString([]byte("security-test-secret"))
		if got := request(expiredString); got != http.StatusUnauthorized {
			t.Fatalf("expired token: got %d", got)
		}

		wrongAlg := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
			"sub": "42",
			"exp": time.Now().Add(time.Hour).Unix(),
		})
		wrongAlgString, _ := wrongAlg.SignedString([]byte("security-test-secret"))
		if got := request(wrongAlgString); got != http.StatusUnauthorized {
			t.Fatalf("wrong JWT algorithm: got %d", got)
		}
	})

	t.Run("registration rejects invalid and extra JSON", func(t *testing.T) {
		for name, body := range map[string]string{
			"bad email":      `{"username":"testuser","email":"not-email","password":"Password123"}`,
			"bad username":   `{"username":"x!","password":"Password123"}`,
			"short password": `{"username":"testuser","password":"short"}`,
			"long password":  `{"username":"testuser","password":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`,
			"unknown field":  `{"username":"testuser","password":"Password123","role":"ADMIN"}`,
			"trailing JSON":  `{"username":"testuser","password":"Password123"}{"role":"ADMIN"}`,
		} {
			t.Run(name, func(t *testing.T) {
				rec := httptest.NewRecorder()
				RegisterHandler(rec, httptest.NewRequest(http.MethodPost, "/api/register", strings.NewReader(body)))
				if rec.Code != http.StatusBadRequest {
					t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
				}
			})
		}
	})
}
