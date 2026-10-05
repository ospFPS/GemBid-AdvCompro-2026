package gembid

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	_ "adcompro-backend/docs"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func Run() {
	ConnectDatabase()
	defer DB.Close()

	jwtSecret := os.Getenv("JWT_SECRET")
	if len([]byte(jwtSecret)) < 32 {
		log.Fatal("JWT_SECRET must be at least 32 bytes")
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /docs", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/docs/index.html", http.StatusMovedPermanently)
	})
	mux.Handle("GET /docs/", httpSwagger.Handler(
		httpSwagger.URL("/docs/doc.json"),
		httpSwagger.PersistAuthorization(true),
	))

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	mux.HandleFunc("POST /api/register", RegisterHandler)

	mux.HandleFunc("POST /api/login", LoginHandler)

	mux.HandleFunc("GET /api/me", AuthMiddleware(MeHandler))
	mux.HandleFunc("PUT /api/me", AuthMiddleware(UpdateProfileHandler))
	mux.HandleFunc("PUT /api/me/password", AuthMiddleware(ChangePasswordHandler))
	mux.HandleFunc("POST /api/seller/apply", AuthMiddleware(SellerApplyHandler))

	mux.HandleFunc("GET /api/seller/status", AuthMiddleware(SellerStatusHandler))

	mux.HandleFunc("GET /api/admin/seller-applications", AuthMiddleware(AdminMiddleware(AdminSellerApplicationsHandler)))

	mux.HandleFunc("PUT /api/admin/seller-applications/{id}", AuthMiddleware(AdminMiddleware(AdminReviewSellerHandler)))

	mux.HandleFunc("POST /api/gems", AuthMiddleware(VerifiedSellerMiddleware(CreateGemHandler)))

	mux.HandleFunc("GET /api/gems", ListGemsHandler)

	mux.HandleFunc("GET /api/gems/{id}", GetGemHandler)

	mux.HandleFunc("PUT /api/gems/{id}", AuthMiddleware(VerifiedSellerMiddleware(UpdateGemHandler)))

	mux.HandleFunc("DELETE /api/gems/{id}", AuthMiddleware(VerifiedSellerMiddleware(DeleteGemHandler)))

	mux.HandleFunc("POST /api/listings", AuthMiddleware(VerifiedSellerMiddleware(CreateListingHandler)))

	mux.HandleFunc("GET /api/listings", ListListingsHandler)

	mux.HandleFunc("GET /api/listings/{id}", GetListingHandler)

	mux.HandleFunc("PUT /api/listings/{id}/publish", AuthMiddleware(VerifiedSellerMiddleware(PublishListingHandler)))

	mux.HandleFunc("DELETE /api/listings/{id}", AuthMiddleware(VerifiedSellerMiddleware(DeleteListingHandler)))

	mux.HandleFunc("POST /api/auctions", AuthMiddleware(VerifiedSellerMiddleware(CreateAuctionHandler)))

	mux.HandleFunc("GET /api/auctions", ListAuctionsHandler)

	mux.HandleFunc("GET /api/auctions/{id}", GetAuctionHandler)

	mux.HandleFunc("GET /api/auctions/{id}/bids", ListAuctionBidsHandler)

	mux.HandleFunc("POST /api/auctions/{id}/bids", AuthMiddleware(PlaceBidHandler))

	mux.HandleFunc("POST /api/auctions/{id}/close", AuthMiddleware(VerifiedSellerMiddleware(CloseAuctionHandler)))

	mux.HandleFunc("POST /api/listings/{id}/buy", AuthMiddleware(BuyNowHandler))
	mux.HandleFunc("GET /api/orders", AuthMiddleware(ListOrdersHandler))
	mux.HandleFunc("GET /api/orders/{id}", AuthMiddleware(GetOrderHandler))
	mux.HandleFunc("POST /api/orders/{id}/pay", AuthMiddleware(PayOrderHandler))
	mux.HandleFunc("POST /api/orders/{id}/cancel", AuthMiddleware(CancelOrderHandler))
	mux.HandleFunc("PUT /api/orders/{id}/ship", AuthMiddleware(VerifiedSellerMiddleware(ShipOrderHandler)))
	mux.HandleFunc("POST /api/orders/{id}/confirm-delivery", AuthMiddleware(ConfirmDeliveryHandler))

	mux.HandleFunc("GET /api/notifications", AuthMiddleware(ListNotificationsHandler))
	mux.HandleFunc("PUT /api/notifications/read-all", AuthMiddleware(MarkAllNotificationsReadHandler))
	mux.HandleFunc("PUT /api/notifications/{id}/read", AuthMiddleware(MarkNotificationReadHandler))

	mux.HandleFunc("GET /api/subscriptions/plans", ListPlansHandler)
	mux.HandleFunc("GET /api/subscriptions/me", AuthMiddleware(MySubscriptionHandler))
	mux.HandleFunc("POST /api/subscriptions/upgrade", AuthMiddleware(UpgradeSubscriptionHandler))

	mux.HandleFunc("GET /api/admin/analytics", AuthMiddleware(AdminMiddleware(AdminAnalyticsHandler)))

	mux.HandleFunc("POST /api/reports", AuthMiddleware(CreateReportHandler))
	mux.HandleFunc("GET /api/reports", AuthMiddleware(ListMyReportsHandler))
	mux.HandleFunc("GET /api/admin/reports", AuthMiddleware(AdminMiddleware(AdminListReportsHandler)))
	mux.HandleFunc("PUT /api/admin/reports/{id}", AuthMiddleware(AdminMiddleware(AdminReviewReportHandler)))
	mux.HandleFunc("GET /api/admin/moderation-actions", AuthMiddleware(AdminMiddleware(AdminListModerationActionsHandler)))

	StartAuctionCloser()
	StartEmailWorker()

	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	log.Println("GemBid backend running on http://localhost:" + port)

	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}
