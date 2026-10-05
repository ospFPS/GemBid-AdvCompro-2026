package gembid

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

type CreateReportRequest struct {
	ReportedUserID *int64  `json:"reported_user_id"`
	ListingID      *int64  `json:"listing_id"`
	AuctionID      *int64  `json:"auction_id"`
	Reason         string  `json:"reason"`
	Description    *string `json:"description"`
}

type ReviewReportRequest struct {
	Decision string `json:"decision"`
	Action   string `json:"action"`
	Reason   string `json:"reason"`
}

type ReportResponse struct {
	ReportID       int64      `json:"report_id"`
	ReporterUserID int64      `json:"reporter_user_id"`
	ReportedUserID *int64     `json:"reported_user_id"`
	ListingID      *int64     `json:"listing_id"`
	AuctionID      *int64     `json:"auction_id"`
	Reason         string     `json:"reason"`
	Description    *string    `json:"description"`
	Status         string     `json:"status"`
	ReviewedBy     *int64     `json:"reviewed_by"`
	CreatedAt      time.Time  `json:"created_at"`
	ReviewedAt     *time.Time `json:"reviewed_at"`
}

type ModerationActionResponse struct {
	ActionID     int64     `json:"action_id"`
	AdminUserID  int64     `json:"admin_user_id"`
	TargetUserID *int64    `json:"target_user_id"`
	ListingID    *int64    `json:"listing_id"`
	AuctionID    *int64    `json:"auction_id"`
	ActionType   string    `json:"action_type"`
	Reason       *string   `json:"reason"`
	CreatedAt    time.Time `json:"created_at"`
}

const reportColumns = `report_id, reporter_user_id, reported_user_id, listing_id, auction_id,
	reason, description, status, reviewed_by, created_at, reviewed_at`

func scanReport(row pgx.Row) (ReportResponse, error) {
	var report ReportResponse
	err := row.Scan(
		&report.ReportID, &report.ReporterUserID, &report.ReportedUserID,
		&report.ListingID, &report.AuctionID, &report.Reason, &report.Description,
		&report.Status, &report.ReviewedBy, &report.CreatedAt, &report.ReviewedAt,
	)
	return report, err
}

func CreateReportHandler(w http.ResponseWriter, r *http.Request) {
	reporterID, _ := r.Context().Value(userIDKey).(int64)

	var request CreateReportRequest
	if err := decodeJSON(r, &request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	request.Reason = strings.TrimSpace(request.Reason)
	if request.Description != nil {
		value := strings.TrimSpace(*request.Description)
		request.Description = &value
	}
	if request.Reason == "" || len(request.Reason) > 150 {
		http.Error(w, "reason is required and must be at most 150 characters", http.StatusBadRequest)
		return
	}
	if request.Description != nil && len(*request.Description) > 2000 {
		http.Error(w, "description must be at most 2000 characters", http.StatusBadRequest)
		return
	}

	targets := 0
	for _, id := range []*int64{request.ReportedUserID, request.ListingID, request.AuctionID} {
		if id != nil {
			if *id <= 0 {
				http.Error(w, "target IDs must be positive", http.StatusBadRequest)
				return
			}
			targets++
		}
	}
	if targets != 1 {
		http.Error(w, "report exactly one target: reported_user_id, listing_id, or auction_id", http.StatusBadRequest)
		return
	}

	var ownerUserID int64
	switch {
	case request.ReportedUserID != nil:
		if *request.ReportedUserID == reporterID {
			http.Error(w, "You cannot report yourself", http.StatusConflict)
			return
		}
		if err := DB.QueryRow(r.Context(),
			`SELECT user_id FROM public.users WHERE user_id = $1`, *request.ReportedUserID,
		).Scan(&ownerUserID); err == pgx.ErrNoRows {
			http.Error(w, "Reported user not found", http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, "Could not verify reported user", http.StatusInternalServerError)
			return
		}

	case request.ListingID != nil:
		if err := DB.QueryRow(r.Context(), `
			SELECT sp.user_id
			FROM public.listings l
			INNER JOIN public.seller_profiles sp ON sp.seller_id = l.seller_id
			WHERE l.listing_id = $1
		`, *request.ListingID).Scan(&ownerUserID); err == pgx.ErrNoRows {
			http.Error(w, "Listing not found", http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, "Could not verify listing", http.StatusInternalServerError)
			return
		}
		if ownerUserID == reporterID {
			http.Error(w, "You cannot report your own listing", http.StatusConflict)
			return
		}

	case request.AuctionID != nil:
		if err := DB.QueryRow(r.Context(), `
			SELECT sp.user_id
			FROM public.auctions a
			INNER JOIN public.listings l ON l.listing_id = a.listing_id
			INNER JOIN public.seller_profiles sp ON sp.seller_id = l.seller_id
			WHERE a.auction_id = $1
		`, *request.AuctionID).Scan(&ownerUserID); err == pgx.ErrNoRows {
			http.Error(w, "Auction not found", http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, "Could not verify auction", http.StatusInternalServerError)
			return
		}
		if ownerUserID == reporterID {
			http.Error(w, "You cannot report your own auction", http.StatusConflict)
			return
		}
	}

	report, err := scanReport(DB.QueryRow(r.Context(), `
		INSERT INTO public.reports
			(reporter_user_id, reported_user_id, listing_id, auction_id, reason, description)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING `+reportColumns,
		reporterID, request.ReportedUserID, request.ListingID, request.AuctionID,
		request.Reason, request.Description,
	))
	if err != nil {
		http.Error(w, "Could not create report", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, report)
}

func ListMyReportsHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(userIDKey).(int64)
	rows, err := DB.Query(r.Context(),
		`SELECT `+reportColumns+` FROM public.reports WHERE reporter_user_id = $1 ORDER BY created_at DESC LIMIT 200`,
		userID,
	)
	if err != nil {
		http.Error(w, "Could not load reports", http.StatusInternalServerError)
		return
	}
	reports, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (ReportResponse, error) {
		return scanReport(row)
	})
	if err != nil {
		http.Error(w, "Could not read reports", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, reports)
}

func AdminListReportsHandler(w http.ResponseWriter, r *http.Request) {
	status := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("status")))
	if status == "" {
		status = "PENDING"
	}
	if status != "ALL" && status != "PENDING" && status != "RESOLVED" && status != "REJECTED" {
		http.Error(w, "status must be PENDING, RESOLVED, REJECTED, or ALL", http.StatusBadRequest)
		return
	}

	query := `SELECT ` + reportColumns + ` FROM public.reports`
	args := []any{}
	if status != "ALL" {
		query += ` WHERE status = $1`
		args = append(args, status)
	}
	query += ` ORDER BY created_at DESC LIMIT 200`

	rows, err := DB.Query(r.Context(), query, args...)
	if err != nil {
		http.Error(w, "Could not load reports", http.StatusInternalServerError)
		return
	}
	reports, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (ReportResponse, error) {
		return scanReport(row)
	})
	if err != nil {
		http.Error(w, "Could not read reports", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"reports": reports, "count": len(reports)})
}

func resolveReportTarget(ctx context.Context, tx pgx.Tx, report ReportResponse) (targetUserID, sellerID, listingID, auctionID int64, err error) {
	if report.ReportedUserID != nil {
		targetUserID = *report.ReportedUserID
		_ = tx.QueryRow(ctx,
			`SELECT seller_id FROM public.seller_profiles WHERE user_id = $1`,
			targetUserID,
		).Scan(&sellerID)
		return
	}

	if report.ListingID != nil {
		listingID = *report.ListingID
		err = tx.QueryRow(ctx, `
			SELECT l.seller_id, sp.user_id
			FROM public.listings l
			INNER JOIN public.seller_profiles sp ON sp.seller_id = l.seller_id
			WHERE l.listing_id = $1
		`, listingID).Scan(&sellerID, &targetUserID)
		return
	}

	auctionID = *report.AuctionID
	err = tx.QueryRow(ctx, `
		SELECT a.listing_id, l.seller_id, sp.user_id
		FROM public.auctions a
		INNER JOIN public.listings l ON l.listing_id = a.listing_id
		INNER JOIN public.seller_profiles sp ON sp.seller_id = l.seller_id
		WHERE a.auction_id = $1
	`, auctionID).Scan(&listingID, &sellerID, &targetUserID)
	return
}

func AdminListModerationActionsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := DB.Query(r.Context(), `
		SELECT action_id, admin_user_id, target_user_id, listing_id, auction_id,
			action_type, reason, created_at
		FROM public.moderation_actions
		ORDER BY created_at DESC, action_id DESC
		LIMIT 200
	`)
	if err != nil {
		http.Error(w, "Could not load moderation actions", http.StatusInternalServerError)
		return
	}
	actions, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (ModerationActionResponse, error) {
		var action ModerationActionResponse
		err := row.Scan(&action.ActionID, &action.AdminUserID, &action.TargetUserID,
			&action.ListingID, &action.AuctionID, &action.ActionType, &action.Reason,
			&action.CreatedAt)
		return action, err
	})
	if err != nil {
		http.Error(w, "Could not read moderation actions", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, actions)
}

func AdminReviewReportHandler(w http.ResponseWriter, r *http.Request) {
	adminID, _ := r.Context().Value(userIDKey).(int64)
	reportID, ok := pathID(r)
	if !ok {
		http.Error(w, "Invalid report ID", http.StatusBadRequest)
		return
	}

	var request ReviewReportRequest
	if err := decodeJSON(r, &request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	request.Decision = strings.ToUpper(strings.TrimSpace(request.Decision))
	request.Action = strings.ToUpper(strings.TrimSpace(request.Action))
	request.Reason = strings.TrimSpace(request.Reason)
	if request.Action == "" {
		request.Action = "NONE"
	}
	if request.Decision != "RESOLVE" && request.Decision != "REJECT" {
		http.Error(w, "decision must be RESOLVE or REJECT", http.StatusBadRequest)
		return
	}
	if request.Decision == "REJECT" && request.Action != "NONE" {
		http.Error(w, "rejected reports cannot have a moderation action", http.StatusBadRequest)
		return
	}
	if request.Action != "NONE" && request.Action != "WARN_USER" &&
		request.Action != "SUSPEND_LISTING" && request.Action != "SUSPEND_SELLER" {
		http.Error(w, "action must be NONE, WARN_USER, SUSPEND_LISTING, or SUSPEND_SELLER", http.StatusBadRequest)
		return
	}
	if len(request.Reason) > 2000 {
		http.Error(w, "reason must be at most 2000 characters", http.StatusBadRequest)
		return
	}

	tx, err := DB.Begin(r.Context())
	if err != nil {
		http.Error(w, "Could not start review transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(r.Context())

	report, err := scanReport(tx.QueryRow(r.Context(),
		`SELECT `+reportColumns+` FROM public.reports WHERE report_id = $1 FOR UPDATE`,
		reportID,
	))
	if err == pgx.ErrNoRows {
		http.Error(w, "Report not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Could not load report", http.StatusInternalServerError)
		return
	}
	if report.Status != "PENDING" {
		http.Error(w, "Report has already been reviewed", http.StatusConflict)
		return
	}

	targetUserID, sellerID, listingID, auctionID, err := resolveReportTarget(r.Context(), tx, report)
	if err != nil && err != pgx.ErrNoRows {
		http.Error(w, "Could not resolve report target", http.StatusInternalServerError)
		return
	}

	if request.Decision == "RESOLVE" {
		switch request.Action {
		case "WARN_USER":
			if targetUserID == 0 {
				http.Error(w, "Report target has no user to warn", http.StatusConflict)
				return
			}
			if err := notify(r.Context(), tx, targetUserID, "MODERATION_WARNING", "Moderation warning",
				"An administrator reviewed a report involving your account. "+request.Reason,
				auctionID, listingID, 0); err != nil {
				http.Error(w, "Could not create warning notification", http.StatusInternalServerError)
				return
			}

		case "SUSPEND_LISTING":
			if listingID == 0 {
				http.Error(w, "This report does not target a listing or auction", http.StatusConflict)
				return
			}
			if auctionID == 0 {
				_ = tx.QueryRow(r.Context(),
					`SELECT auction_id FROM public.auctions WHERE listing_id = $1`, listingID,
				).Scan(&auctionID)
			}
			if auctionID != 0 {
				var winningBidID *int64
				if err := tx.QueryRow(r.Context(),
					`SELECT winning_bid_id FROM public.auctions WHERE auction_id = $1 FOR UPDATE`,
					auctionID,
				).Scan(&winningBidID); err != nil && err != pgx.ErrNoRows {
					http.Error(w, "Could not lock auction", http.StatusInternalServerError)
					return
				}
				if winningBidID != nil {
					if _, err := tx.Exec(r.Context(),
						`UPDATE public.bids SET bid_status = 'RESERVE_NOT_MET' WHERE bid_id = $1`,
						*winningBidID,
					); err != nil {
						http.Error(w, "Could not close auction bid", http.StatusInternalServerError)
						return
					}
				}
				if _, err := tx.Exec(r.Context(), `
					UPDATE public.auctions
					SET status = 'ENDED', winner_id = NULL, winning_bid_id = NULL, closed_at = NOW()
					WHERE auction_id = $1 AND status IN ('SCHEDULED','ACTIVE')
				`, auctionID); err != nil {
					http.Error(w, "Could not stop auction", http.StatusInternalServerError)
					return
				}
			}
			tag, err := tx.Exec(r.Context(),
				`UPDATE public.listings SET status = 'SUSPENDED', updated_at = NOW()
				 WHERE listing_id = $1 AND status IN ('DRAFT','PUBLISHED','ACTIVE')`,
				listingID,
			)
			if err != nil {
				http.Error(w, "Could not suspend listing", http.StatusInternalServerError)
				return
			}
			if tag.RowsAffected() == 0 {
				http.Error(w, "Listing cannot be suspended in its current state", http.StatusConflict)
				return
			}

		case "SUSPEND_SELLER":
			if sellerID == 0 || targetUserID == 0 {
				http.Error(w, "Report target is not a seller", http.StatusConflict)
				return
			}
			if _, err := tx.Exec(r.Context(),
				`UPDATE public.users SET seller_status = 'SUSPENDED', updated_at = NOW() WHERE user_id = $1`,
				targetUserID,
			); err != nil {
				http.Error(w, "Could not suspend seller user", http.StatusInternalServerError)
				return
			}
			if _, err := tx.Exec(r.Context(),
				`UPDATE public.seller_profiles SET verification_status = 'SUSPENDED' WHERE seller_id = $1`,
				sellerID,
			); err != nil {
				http.Error(w, "Could not suspend seller profile", http.StatusInternalServerError)
				return
			}
			if _, err := tx.Exec(r.Context(), `
				UPDATE public.bids b
				SET bid_status = 'RESERVE_NOT_MET'
				FROM public.auctions a
				INNER JOIN public.listings l ON l.listing_id = a.listing_id
				WHERE a.winning_bid_id = b.bid_id
				  AND l.seller_id = $1
				  AND a.status IN ('SCHEDULED','ACTIVE')
			`, sellerID); err != nil {
				http.Error(w, "Could not finalize seller auctions", http.StatusInternalServerError)
				return
			}
			if _, err := tx.Exec(r.Context(), `
				UPDATE public.auctions a
				SET status = 'ENDED', winner_id = NULL, winning_bid_id = NULL, closed_at = NOW()
				FROM public.listings l
				WHERE a.listing_id = l.listing_id
				  AND l.seller_id = $1
				  AND a.status IN ('SCHEDULED','ACTIVE')
			`, sellerID); err != nil {
				http.Error(w, "Could not stop seller auctions", http.StatusInternalServerError)
				return
			}
			if _, err := tx.Exec(r.Context(),
				`UPDATE public.listings SET status = 'SUSPENDED', updated_at = NOW()
				 WHERE seller_id = $1 AND status IN ('PUBLISHED','ACTIVE')`,
				sellerID,
			); err != nil {
				http.Error(w, "Could not suspend seller listings", http.StatusInternalServerError)
				return
			}
		}

		if request.Action != "NONE" {
			if _, err := tx.Exec(r.Context(), `
				INSERT INTO public.moderation_actions
					(admin_user_id, target_user_id, listing_id, auction_id, action_type, reason)
				VALUES ($1, NULLIF($2::bigint,0), NULLIF($3::bigint,0), NULLIF($4::bigint,0), $5, NULLIF($6,''))
			`, adminID, targetUserID, listingID, auctionID, request.Action, request.Reason); err != nil {
				http.Error(w, "Could not record moderation action", http.StatusInternalServerError)
				return
			}
		}
	}

	newStatus := "REJECTED"
	if request.Decision == "RESOLVE" {
		newStatus = "RESOLVED"
	}
	report, err = scanReport(tx.QueryRow(r.Context(), `
		UPDATE public.reports
		SET status = $1, reviewed_by = $2, reviewed_at = NOW()
		WHERE report_id = $3
		RETURNING `+reportColumns,
		newStatus, adminID, reportID,
	))
	if err != nil {
		http.Error(w, "Could not update report", http.StatusInternalServerError)
		return
	}

	if err := notify(r.Context(), tx, report.ReporterUserID, "REPORT_REVIEWED", "Your report was reviewed",
		"Your report was "+strings.ToLower(newStatus)+".", auctionID, listingID, 0); err != nil {
		http.Error(w, "Could not notify reporter", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(r.Context()); err != nil {
		http.Error(w, "Could not complete report review", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"message": "Report reviewed",
		"report":  report,
		"action":  request.Action,
	})
}
