package test

import (
	. "adcompro-backend"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestReportModerationFlow(t *testing.T) {
	if os.Getenv("RUN_AUCTION_INTEGRATION") != "1" {
		t.Skip("set RUN_AUCTION_INTEGRATION=1 to run database integration test")
	}

	ConnectDatabase()
	t.Cleanup(func() { DB.Close() })
	ctx := context.Background()

	var adminID int64
	if err := DB.QueryRow(ctx,
		`SELECT user_id FROM public.users WHERE role = 'ADMIN' ORDER BY user_id LIMIT 1`,
	).Scan(&adminID); err != nil {
		t.Fatalf("need one admin in test database: %v", err)
	}

	suffix := time.Now().UnixNano()
	newUser := func(name string) int64 {
		var id int64
		if err := DB.QueryRow(ctx, `
			INSERT INTO public.users (username, full_name, email, password_hash)
			VALUES ($1, $2, $3, 'integration-only')
			RETURNING user_id
		`, fmt.Sprintf("rep_%s_%d", name, suffix%1000000000), name, fmt.Sprintf("report-it-%s-%d@test.local", name, suffix)).Scan(&id); err != nil {
			t.Fatalf("create %s: %v", name, err)
		}
		return id
	}

	reporterID := newUser("reporter")
	sellerUserID := newUser("seller")

	var sellerID int64
	if err := DB.QueryRow(ctx, `
		INSERT INTO public.seller_profiles
			(user_id, shop_name, legal_name, verification_status)
		VALUES ($1, $2, $3, 'VERIFIED')
		RETURNING seller_id
	`, sellerUserID, "Report Test Shop", "Report Test Seller").Scan(&sellerID); err != nil {
		t.Fatal(err)
	}
	if _, err := DB.Exec(ctx,
		`UPDATE public.users SET seller_status = 'VERIFIED' WHERE user_id = $1`,
		sellerUserID,
	); err != nil {
		t.Fatal(err)
	}

	var gemID int64
	if err := DB.QueryRow(ctx, `
		INSERT INTO public.gems (seller_id, gem_name, gem_type, carat, shape, color)
		VALUES ($1, $2, 'Sapphire', 1.0, 'Oval', 'Blue')
		RETURNING gem_id
	`, sellerID, fmt.Sprintf("Report Integration Gem %d", suffix)).Scan(&gemID); err != nil {
		t.Fatal(err)
	}

	var listingID int64
	if err := DB.QueryRow(ctx, `
		INSERT INTO public.listings (gem_id, seller_id, sale_type, status, fixed_price, published_at)
		VALUES ($1, $2, 'FIXED_PRICE', 'PUBLISHED', 1200, NOW())
		RETURNING listing_id
	`, gemID, sellerID).Scan(&listingID); err != nil {
		t.Fatal(err)
	}

	var reportID int64
	t.Cleanup(func() {
		cleanup := context.Background()
		_, _ = DB.Exec(cleanup, `
			DELETE FROM public.email_deliveries
			WHERE notification_id IN (
				SELECT notification_id FROM public.notifications
				WHERE user_id = ANY($1)
			)
		`, []int64{reporterID, sellerUserID})
		_, _ = DB.Exec(cleanup,
			`DELETE FROM public.notifications WHERE user_id = ANY($1)`,
			[]int64{reporterID, sellerUserID},
		)
		if reportID != 0 {
			_, _ = DB.Exec(cleanup,
				`DELETE FROM public.moderation_actions WHERE listing_id = $1`,
				listingID,
			)
			_, _ = DB.Exec(cleanup,
				`DELETE FROM public.reports WHERE report_id = $1`,
				reportID,
			)
		}
		_, _ = DB.Exec(cleanup, `DELETE FROM public.listings WHERE listing_id = $1`, listingID)
		_, _ = DB.Exec(cleanup, `DELETE FROM public.gems WHERE gem_id = $1`, gemID)
		_, _ = DB.Exec(cleanup, `DELETE FROM public.seller_profiles WHERE seller_id = $1`, sellerID)
		_, _ = DB.Exec(cleanup,
			`DELETE FROM public.users WHERE user_id = ANY($1)`,
			[]int64{reporterID, sellerUserID},
		)
	})

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/reports", AuthMiddleware(CreateReportHandler))
	mux.HandleFunc("GET /api/reports", AuthMiddleware(ListMyReportsHandler))
	mux.HandleFunc("GET /api/admin/reports", AuthMiddleware(AdminMiddleware(AdminListReportsHandler)))
	mux.HandleFunc("PUT /api/admin/reports/{id}", AuthMiddleware(AdminMiddleware(AdminReviewReportHandler)))
	mux.HandleFunc("GET /api/admin/moderation-actions", AuthMiddleware(AdminMiddleware(AdminListModerationActionsHandler)))
	server := httptest.NewServer(mux)
	defer server.Close()

	reporterToken := integrationToken(t, reporterID)
	sellerToken := integrationToken(t, sellerUserID)
	adminToken := integrationToken(t, adminID)

	invalid := integrationRequest(t, http.MethodPost, server.URL+"/api/reports", reporterToken, map[string]any{
		"reported_user_id": sellerUserID,
		"listing_id":       listingID,
		"reason":           "two targets",
	})
	requireHTTPStatus(t, "reject multiple report targets", invalid, http.StatusBadRequest)

	own := integrationRequest(t, http.MethodPost, server.URL+"/api/reports", sellerToken, map[string]any{
		"listing_id": listingID,
		"reason":     "own listing",
	})
	requireHTTPStatus(t, "reject own listing report", own, http.StatusConflict)

	created := integrationRequest(t, http.MethodPost, server.URL+"/api/reports", reporterToken, map[string]any{
		"listing_id":  listingID,
		"reason":      "Misleading listing",
		"description": "Integration test report",
	})
	requireHTTPStatus(t, "create report", created, http.StatusCreated)

	var report ReportResponse
	if err := json.Unmarshal(created.Body, &report); err != nil {
		t.Fatal(err)
	}
	reportID = report.ReportID
	if report.Status != "PENDING" || report.ListingID == nil || *report.ListingID != listingID {
		t.Fatalf("unexpected report: %+v", report)
	}

	requireHTTPStatus(t, "non-admin list reports",
		integrationRequest(t, http.MethodGet, server.URL+"/api/admin/reports", reporterToken, nil),
		http.StatusForbidden,
	)
	requireHTTPStatus(t, "admin list reports",
		integrationRequest(t, http.MethodGet, server.URL+"/api/admin/reports?status=PENDING", adminToken, nil),
		http.StatusOK,
	)

	review := integrationRequest(t, http.MethodPut,
		fmt.Sprintf("%s/api/admin/reports/%d", server.URL, reportID),
		adminToken,
		map[string]any{
			"decision": "RESOLVE",
			"action":   "SUSPEND_LISTING",
			"reason":   "Listing violates marketplace rules",
		},
	)
	requireHTTPStatus(t, "resolve report", review, http.StatusOK)

	var listingStatus string
	if err := DB.QueryRow(ctx,
		`SELECT status FROM public.listings WHERE listing_id = $1`, listingID,
	).Scan(&listingStatus); err != nil {
		t.Fatal(err)
	}
	if listingStatus != "SUSPENDED" {
		t.Fatalf("expected listing SUSPENDED, got %s", listingStatus)
	}

	var actionCount int
	if err := DB.QueryRow(ctx, `
		SELECT count(*) FROM public.moderation_actions
		WHERE listing_id = $1 AND action_type = 'SUSPEND_LISTING'
	`, listingID).Scan(&actionCount); err != nil {
		t.Fatal(err)
	}
	if actionCount != 1 {
		t.Fatalf("expected one moderation action, got %d", actionCount)
	}

	requireHTTPStatus(t, "review report twice",
		integrationRequest(t, http.MethodPut,
			fmt.Sprintf("%s/api/admin/reports/%d", server.URL, reportID),
			adminToken,
			map[string]any{"decision": "REJECT", "action": "NONE"},
		),
		http.StatusConflict,
	)

	myReports := integrationRequest(t, http.MethodGet, server.URL+"/api/reports", reporterToken, nil)
	requireHTTPStatus(t, "list my reports", myReports, http.StatusOK)
	var reports []ReportResponse
	if err := json.Unmarshal(myReports.Body, &reports); err != nil {
		t.Fatal(err)
	}
	if len(reports) == 0 || reports[0].ReportID != reportID || reports[0].Status != "RESOLVED" {
		t.Fatalf("unexpected report history: %+v", reports)
	}

	requireHTTPStatus(t, "admin moderation history",
		integrationRequest(t, http.MethodGet, server.URL+"/api/admin/moderation-actions", adminToken, nil),
		http.StatusOK,
	)
}
