package gembid

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type execer interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func pathID(r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	return id, err == nil && id > 0
}

func notify(
	ctx context.Context,
	db execer,
	userID int64,
	kind, title, message string,
	auctionID, listingID, orderID int64,
) error {
	_, err := db.Exec(
		ctx,
		`
		WITH n AS (
			INSERT INTO public.notifications
				(user_id, type, title, message, auction_id, listing_id, order_id)
			VALUES ($1, $2, $3, $4, NULLIF($5::bigint, 0), NULLIF($6::bigint, 0), NULLIF($7::bigint, 0))
			RETURNING notification_id, user_id
		)
		INSERT INTO public.email_deliveries (notification_id, recipient_email)
		SELECT n.notification_id, u.email
		FROM n
		INNER JOIN public.users u ON u.user_id = n.user_id
		WHERE u.email IS NOT NULL AND u.email <> ''
		`,
		userID, kind, title, message, auctionID, listingID, orderID,
	)
	return err
}

func StartEmailWorker() {
	host := os.Getenv("SMTP_HOST")
	if host == "" {
		log.Println("SMTP_HOST not set: emails stay queued in email_deliveries")
		return
	}

	port := os.Getenv("SMTP_PORT")
	if port == "" {
		port = "587"
	}
	from := os.Getenv("SMTP_FROM")
	auth := smtp.PlainAuth("", os.Getenv("SMTP_USER"), os.Getenv("SMTP_PASS"), host)

	send := func(to, subject, body string) error {
		msg := "From: " + from + "\r\nTo: " + to + "\r\nSubject: " + subject +
			"\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n" + body
		return smtp.SendMail(host+":"+port, auth, from, []string{to}, []byte(msg))
	}

	go func() {
		for {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			if err := sendPendingEmails(ctx, send); err != nil {
				log.Println("email worker:", err)
			}
			cancel()
			time.Sleep(30 * time.Second)
		}
	}()
}

func sendPendingEmails(ctx context.Context, send func(to, subject, body string) error) error {
	rows, err := DB.Query(
		ctx,
		`
		SELECT ed.email_delivery_id, ed.recipient_email, n.title, n.message
		FROM public.email_deliveries ed
		INNER JOIN public.notifications n ON n.notification_id = ed.notification_id
		WHERE ed.status = 'PENDING'
		ORDER BY ed.email_delivery_id
		LIMIT 20
		`,
	)
	if err != nil {
		return err
	}

	type email struct {
		id                int64
		to, subject, body string
	}
	emails, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (email, error) {
		var e email
		err := row.Scan(&e.id, &e.to, &e.subject, &e.body)
		return e, err
	})
	if err != nil {
		return err
	}

	for _, e := range emails {
		if sendErr := send(e.to, "[GemBid] "+e.subject, e.body); sendErr != nil {
			_, err = DB.Exec(ctx,
				`UPDATE public.email_deliveries SET status = 'FAILED', failed_at = NOW(), error_message = $2 WHERE email_delivery_id = $1`,
				e.id, sendErr.Error())
		} else {
			_, err = DB.Exec(ctx,
				`UPDATE public.email_deliveries SET status = 'SENT', sent_at = NOW() WHERE email_delivery_id = $1`,
				e.id)
		}
		if err != nil {
			return fmt.Errorf("update email %d: %w", e.id, err)
		}
	}
	return nil
}

type NotificationResponse struct {
	NotificationID int64     `json:"notification_id"`
	Type           string    `json:"type"`
	Title          string    `json:"title"`
	Message        string    `json:"message"`
	AuctionID      *int64    `json:"auction_id"`
	ListingID      *int64    `json:"listing_id"`
	OrderID        *int64    `json:"order_id"`
	IsRead         bool      `json:"is_read"`
	CreatedAt      time.Time `json:"created_at"`
}

func ListNotificationsHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(userIDKey).(int64)
	unreadOnly := r.URL.Query().Get("unread") == "true"

	rows, err := DB.Query(
		r.Context(),
		`
		SELECT notification_id, type, title, message, auction_id, listing_id,
			order_id, is_read, created_at
		FROM public.notifications
		WHERE user_id = $1 AND (NOT $2 OR is_read = FALSE)
		ORDER BY created_at DESC, notification_id DESC
		LIMIT 100
		`,
		userID, unreadOnly,
	)
	if err != nil {
		http.Error(w, "Could not load notifications", http.StatusInternalServerError)
		return
	}

	notifications, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (NotificationResponse, error) {
		var n NotificationResponse
		err := row.Scan(&n.NotificationID, &n.Type, &n.Title, &n.Message, &n.AuctionID,
			&n.ListingID, &n.OrderID, &n.IsRead, &n.CreatedAt)
		return n, err
	})
	if err != nil {
		http.Error(w, "Could not read notifications", http.StatusInternalServerError)
		return
	}

	var unread int
	_ = DB.QueryRow(r.Context(),
		`SELECT count(*) FROM public.notifications WHERE user_id = $1 AND is_read = FALSE`,
		userID).Scan(&unread)

	writeJSON(w, http.StatusOK, map[string]any{
		"notifications": notifications,
		"unread_count":  unread,
	})
}

func MarkNotificationReadHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(userIDKey).(int64)
	id, ok := pathID(r)
	if !ok {
		http.Error(w, "Invalid notification ID", http.StatusBadRequest)
		return
	}

	tag, err := DB.Exec(r.Context(),
		`UPDATE public.notifications SET is_read = TRUE WHERE notification_id = $1 AND user_id = $2`,
		id, userID)
	if err != nil {
		http.Error(w, "Could not update notification", http.StatusInternalServerError)
		return
	}
	if tag.RowsAffected() == 0 {
		http.Error(w, "Notification not found", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "Notification marked as read"})
}

func MarkAllNotificationsReadHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(userIDKey).(int64)

	tag, err := DB.Exec(r.Context(),
		`UPDATE public.notifications SET is_read = TRUE WHERE user_id = $1 AND is_read = FALSE`,
		userID)
	if err != nil {
		http.Error(w, "Could not update notifications", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"message": "All notifications marked as read",
		"updated": tag.RowsAffected(),
	})
}
