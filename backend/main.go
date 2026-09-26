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

	// ============================================================
	// GEMS
	// ============================================================

	mux.HandleFunc(
		"POST /api/gems",
		AuthMiddleware(
			VerifiedSellerMiddleware(
				CreateGemHandler,
			),
		),
	)

	// Public — browse all gems.
	mux.HandleFunc(
		"GET /api/gems",
		ListGemsHandler,
	)

	// Public — single gem detail.
	mux.HandleFunc(
		"GET /api/gems/{id}",
		GetGemHandler,
	)

	mux.HandleFunc(
		"PUT /api/gems/{id}",
		AuthMiddleware(
			VerifiedSellerMiddleware(
				UpdateGemHandler,
			),
		),
	)

	mux.HandleFunc(
		"DELETE /api/gems/{id}",
		AuthMiddleware(
			VerifiedSellerMiddleware(
				DeleteGemHandler,
			),
		),
	)

	// ============================================================
	// LISTINGS
	// ============================================================

	mux.HandleFunc(
		"POST /api/listings",
		AuthMiddleware(
			VerifiedSellerMiddleware(
				CreateListingHandler,
			),
		),
	)

	// Public — browse published/active listings.
	mux.HandleFunc(
		"GET /api/listings",
		ListListingsHandler,
	)

	// Public — single listing detail.
	mux.HandleFunc(
		"GET /api/listings/{id}",
		GetListingHandler,
	)

	// Seller-owned — choose sale type (FIXED_PRICE or AUCTION) for a DRAFT listing.
	mux.HandleFunc(
		"PUT /api/listings/{id}/publish",
		AuthMiddleware(
			VerifiedSellerMiddleware(
				PublishListingHandler,
			),
		),
	)

	// Seller-owned — only while still DRAFT.
	mux.HandleFunc(
		"DELETE /api/listings/{id}",
		AuthMiddleware(
			VerifiedSellerMiddleware(
				DeleteListingHandler,
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
