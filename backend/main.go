package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

func main() {
	ConnectDatabase()
	defer DB.Close()

	if os.Getenv("JWT_SECRET") == "" {
		log.Fatal("JWT_SECRET is not set")
	}

	mux := http.NewServeMux()

	mux.HandleFunc(
		"GET /health",
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			json.NewEncoder(w).Encode(map[string]string{
				"status": "ok",
			})
		},
	)

	mux.HandleFunc(
		"POST /api/register",
		RegisterHandler,
	)

	mux.HandleFunc(
		"POST /api/login",
		LoginHandler,
	)

	mux.HandleFunc(
		"GET /api/me",
		AuthMiddleware(MeHandler),
	)
	// Normal user submits seller application.
	mux.HandleFunc(
		"POST /api/seller/apply",
		AuthMiddleware(SellerApplyHandler),
	)

	// Logged-in user checks seller application status.
	mux.HandleFunc(
		"GET /api/seller/status",
		AuthMiddleware(SellerStatusHandler),
	)

	// Admin views all seller applications.
	mux.HandleFunc(
		"GET /api/admin/seller-applications",
		AuthMiddleware(
			AdminMiddleware(
				AdminSellerApplicationsHandler,
			),
		),
	)

	// Admin approves or rejects a seller application.
	mux.HandleFunc(
		"PUT /api/admin/seller-applications/{id}",
		AuthMiddleware(
			AdminMiddleware(
				AdminReviewSellerHandler,
			),
		),
	)
	mux.HandleFunc(
		"POST /api/gems",
		AuthMiddleware(
			VerifiedSellerMiddleware(
				CreateGemHandler,
			),
		),
	)

	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	log.Println(
		"GemBid backend running on http://localhost:" + port,
	)

	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}
