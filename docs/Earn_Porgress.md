# GemBid Backend Progress — Earn

**Project:** GemBid — Gem Marketplace & Auction Platform
**Date:** 5 October 2026
**Branch:** `Earn_TTB`
**Builds on:** Oshi's week 1 backend (`main` @ `8e83949`), merged into `Earn_TTB`

---

# 1. Summary

This update adds the remaining commerce features on top of the existing
authentication, seller, gem, listing and auction engine.

| Feature | Status |
|---|---|
| Order + Buy Now System | ✅ |
| Payment / Transaction System (mock) | ✅ |
| Shipping System | ✅ |
| Notification + Email System | ✅ (email sending needs SMTP settings) |
| Admin Analytics | ✅ |
| Subscription / Monetization (FREE / PRO) | ✅ |
| Auction Finalization / Automatic Closing | ✅ |
| Deploy script for the Ubuntu server | ✅ |

All features passed automated integration tests against the project database.

---

# 2. Commits on `Earn_TTB`

| Commit | Description |
|---|---|
| `98fa7f2` | Merge `main` into `Earn_TTB` (brings in auction engine, Swagger, Docker) |
| `f7a0171` | Orders, payments, shipping, notifications, subscriptions, analytics, auto auction close |
| `0a087c8` | `deploy.sh` for the server |
| `5a61f32` | Code simplification (shorter, same behaviour, re-tested) |

---

# 3. Files

New:

```text
backend/orders.go                    Orders, Buy Now, payment, shipping
backend/notifications.go             Notifications, email outbox + SMTP worker, shared helpers
backend/subscriptions.go             FREE / PRO plans, upgrade, listing quota
backend/orders_integration_test.go   End-to-end test for the new features
backend/db/migrations/019_orders_payments_shipping_notifications_subscriptions.sql
deploy.sh                            One-command server deployment
```

Changed:

```text
backend/auction.go                   Shared finalizeAuction, auto closer, outbid notification, quota
backend/admin.go                     Admin analytics endpoint
backend/listings.go                  FREE listing quota on publish
backend/main.go                      New routes, background workers
backend/auction_integration_test.go  Checks winner order + outbid notifications
backend/docs/*                       Regenerated Swagger
.gitignore                           Ignore *.exe build output
```

---

# 4. Database — Migration 019

The tables (`orders`, `transactions`, `shipping`, `notifications`,
`email_deliveries`, `subscriptions`) already existed from migrations 010–017.
Migration 019 adds the rules they were missing. It runs inside one transaction
(all-or-nothing) and **has already been applied** to the project database.

| Table | Change |
|---|---|
| `orders` | Status limited to `PENDING_PAYMENT`, `PAID`, `SHIPPED`, `DELIVERED`, `CANCELLED` (default `PENDING_PAYMENT`) |
| `orders` | Unique index: one non-cancelled order per listing (prevents double Buy Now) |
| `transactions` | `order_id` now nullable, new `subscription_id`; each payment belongs to exactly one order **or** one subscription |
| `transactions` | Status limited to `PENDING`, `SUCCESS`, `FAILED`, `REFUNDED` |
| `shipping` | Status limited to `PENDING`, `SHIPPED`, `DELIVERED` |
| `email_deliveries` | Status limited to `PENDING`, `SENT`, `FAILED`; index for pending emails |
| `subscriptions` | Status limited to `ACTIVE`, `CANCELLED`, `EXPIRED`; one ACTIVE subscription per user |

---

# 5. Order + Buy Now System

```text
Buy Now / Auction win
        ↓
PENDING_PAYMENT ──(buyer cancels)──→ CANCELLED
        ↓ pay
      PAID
        ↓ seller ships
     SHIPPED
        ↓ buyer confirms
    DELIVERED
```

| Endpoint | Who | Description |
|---|---|---|
| `POST /api/listings/{id}/buy` | Logged-in user | Buy a PUBLISHED fixed-price listing. Listing becomes `SOLD`, order `PENDING_PAYMENT` |
| `GET /api/orders` | Logged-in user | My orders as buyer; `?as=seller` for orders on my shop |
| `GET /api/orders/{id}` | Buyer, seller, admin | Order + shipping + payment transactions |
| `POST /api/orders/{id}/cancel` | Buyer | Cancel unpaid order. Fixed-price listing returns to `PUBLISHED`, auction listing becomes `UNSOLD` |

Rules:

- A seller cannot buy their own listing (403).
- The listing row is locked during Buy Now, and the unique index blocks a second order (409).
- Only unpaid orders can be cancelled.

---

# 6. Payment / Transaction System

Payments are **simulated** (no real money), as allowed by the requirements.

| Endpoint | Who | Description |
|---|---|---|
| `POST /api/orders/{id}/pay` | Buyer | Pay with `MOCK_CARD` or `MOCK_QR` and give the shipping address |

Request:

```json
{
  "payment_method": "MOCK_CARD",
  "shipping_address": "123 Sukhumvit Rd, Bangkok 10110"
}
```

In one database transaction the backend:

1. Records a `transactions` row (`SUCCESS`, reference `MOCK-xxxxxxxxxxxxxxxx`, `paid_at`).
2. Sets the order to `PAID`.
3. Creates the `shipping` row with the address (`PENDING`).
4. Notifies the seller (payment received) and the buyer (payment successful).

Paying twice returns 409.

---

# 7. Shipping System

| Endpoint | Who | Description |
|---|---|---|
| `PUT /api/orders/{id}/ship` | Verified seller (owner) | Add courier + tracking number. Order and shipping become `SHIPPED` |
| `POST /api/orders/{id}/confirm-delivery` | Buyer | Order and shipping become `DELIVERED` |

Ship request:

```json
{
  "courier": "Kerry Express",
  "tracking_number": "KEX123456789TH"
}
```

The buyer is notified with the courier and tracking number; the seller is
notified when delivery is confirmed.

---

# 8. Notification + Email System

Every notification is stored in `notifications` **and** queued in
`email_deliveries` (outbox pattern) in the same database transaction as the
action that caused it.

| Endpoint | Description |
|---|---|
| `GET /api/notifications` | Newest first (max 100) + `unread_count`; `?unread=true` for unread only |
| `PUT /api/notifications/{id}/read` | Mark one as read |
| `PUT /api/notifications/read-all` | Mark all as read |

Notification types:

```text
ORDER_CREATED          seller   Buy Now order placed
PAYMENT_RECEIVED       seller   buyer paid
PAYMENT_SUCCESS        buyer    payment done
ORDER_SHIPPED          buyer    courier + tracking number
ORDER_DELIVERED        seller   buyer confirmed delivery
ORDER_CANCELLED        seller   buyer cancelled unpaid order
OUTBID                 bidder   someone placed a higher bid
AUCTION_WON            winner   pay for your order
AUCTION_SOLD           seller   auction ended with a winner
AUCTION_LOST           bidders  auction ended, you did not win
AUCTION_UNSOLD         seller   no winner / reserve not met
AUCTION_RESERVE_NOT_MET bidder  highest bid below reserve
SUBSCRIPTION_ACTIVATED user     PRO active until <date>
```

## Email sending

A background worker sends pending emails over SMTP every 30 seconds and marks
them `SENT` or `FAILED`. It only runs when these are set in `.env`:

```env
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=...
SMTP_PASS=...      # e.g. a Gmail app password
SMTP_FROM=...
```

Without `SMTP_HOST`, emails stay `PENDING` in `email_deliveries` and are sent
once SMTP is configured.

---

# 9. Auction Finalization / Automatic Closing

Previously an auction only closed when the seller called the close endpoint, and
no order was created for the winner.

Now:

- A background job runs **every 30 seconds**. It activates scheduled auctions
  whose start time has passed and finalizes every auction whose `end_time` has
  passed.
- `POST /api/auctions/{id}/close` still works and uses the **same**
  `finalizeAuction` code.
- Both paths lock the auction row, so an auction can never be closed twice.

```text
end_time passed
      ↓
Reserve met and has bids?
 ├── YES → bid WON, listing SOLD, auction ENDED
 │         PENDING_PAYMENT order for winner
 │         notify winner, seller, losing bidders
 └── NO  → listing UNSOLD, auction ENDED
           notify seller (+ highest bidder if reserve not met)
```

Bidding also now sends an `OUTBID` notification to the previous highest bidder.

---

# 10. Subscription / Monetization

| Plan | Price | Live listings | Features |
|---|---|---|---|
| FREE | ฿0 | 5 | Browse, bid, buy, sell up to 5 live listings |
| PRO | ฿299 / 30 days | Unlimited | Everything in FREE + unlimited live listings |

| Endpoint | Description |
|---|---|
| `GET /api/subscriptions/plans` | Public plan list |
| `GET /api/subscriptions/me` | My current plan |
| `POST /api/subscriptions/upgrade` | Mock payment (`MOCK_CARD` / `MOCK_QR`) for 30 days of PRO |

Rules:

- Paying again while PRO is active extends the end date by 30 days.
- The payment is stored in `transactions` with `subscription_id`.
- A lapsed PRO is marked `EXPIRED` on the next upgrade.
- A FREE seller with 5 live (`PUBLISHED` / `ACTIVE`) listings gets **403** when
  publishing a fixed-price listing or creating an auction.

Prices and limits are constants at the top of `backend/subscriptions.go`.

---

# 11. Admin Analytics

`GET /api/admin/analytics` (admin only, 403 for others) returns one JSON report:

```text
users                      total, admins, new in last 30 days, sellers by status
gems_total
listings_by_status
auctions_by_status
bids_total
orders_by_status
sales                      paid orders, gross merchandise value, average order value
subscriptions              active PRO count, subscription revenue
emails_by_status
top_gem_types_sold         top 5 by revenue
daily_sales_last_30_days   orders + revenue per day
```

PostgreSQL builds the whole report in a single query.

---

# 12. Testing

## Automated integration tests (against the project database)

```bash
RUN_AUCTION_INTEGRATION=1 go test -count=1 -v ./...
```

On the server (uses the server `.env`, nothing installed):

```bash
cd /srv/Adv_Compro_Project/GemBid-AdvCompro-2026
docker run --rm --env-file .env -e RUN_AUCTION_INTEGRATION=1 \
  -v "$PWD/backend":/app -w /app golang:1.26-alpine go test -count=1 -v ./...
```

`TestOrderPaymentShippingFlow` (new) checks:

```text
Seller buys own listing           → 403 ✅
Buyer Buy Now                     → 201 PENDING_PAYMENT ✅
Second Buy Now                    → 409 ✅
Seller pays buyer's order         → 403 ✅
Invalid payment method            → 400 ✅
Buyer pays                        → PAID + shipping PENDING + 1 SUCCESS transaction ✅
Pay again / cancel paid order     → 409 ✅
Buyer tries to ship               → 403 ✅
Confirm before shipping           → 409 ✅
Seller ships                      → SHIPPED ✅
Buyer confirms                    → DELIVERED ✅
Admin views order                 → 200 ✅
Buy + cancel second listing       → CANCELLED, listing PUBLISHED again ✅
Buyer notifications               → PAYMENT_SUCCESS, ORDER_SHIPPED, unread > 0 ✅
Email queued per notification     → ✅
Upgrade twice                     → same subscription, +30 days ✅
Plan                              → PRO ✅
Analytics as buyer / admin        → 403 / 200 with all sections ✅
```

`TestAuctionEngineTwoBidderFlow` (extended) now also checks:

```text
Winner order created              → PENDING_PAYMENT, buyer = winner, total 5500 ✅
OUTBID notifications              → 2 ✅
```

Latest result (after the simplification commit):

```text
--- PASS: TestAuctionEngineTwoBidderFlow (13.35s)
--- PASS: TestOrderPaymentShippingFlow (16.43s)
PASS
```

Test records are deleted automatically after each run. A database check after
the first run confirmed no leftover test rows.

## Static checks

```text
go build .     ✅
go vet ./...   ✅
go test ./...  ✅
Swagger regenerated ✅
```

## Manual testing

Swagger UI: `http://thalamus:8080/docs`. Each new endpoint is listed under
the tags Orders, Payments, Shipping, Notifications, Subscriptions and Admin.

---

# 13. Deployment

Production runs on the Ubuntu server (`thalamus`) with Docker Compose in
`/srv/Adv_Compro_Project/GemBid-AdvCompro-2026`, inside the `gembid` tmux
session. ngrok exposes it publicly.

New `deploy.sh`:

```bash
./deploy.sh                                   # pull, rebuild, restart, health check
./deploy.sh backend/db/migrations/020_x.sql   # also apply a migration first
```

It:

1. Pulls the current branch (`git pull --ff-only`).
2. Applies any migration files given (inside a throwaway Postgres container, using `.env`).
3. Runs `docker compose up -d --build` (detached, so it survives closing tmux).
4. Waits up to 30 s for `/health`; prints the last 50 log lines on failure.
5. Starts ngrok only if it is not already running (fixes the `ERR_NGROK_334` duplicate-tunnel error).

Deploy `Earn_TTB`:

```bash
git fetch && git checkout Earn_TTB && ./deploy.sh && docker compose logs -f --tail 20 backend
```

---

# 14. Known Limitations / Not Built Yet

| Item | Note |
|---|---|
| Unpaid order timeout | Unpaid orders do not auto-cancel; buyer must cancel |
| Refunds | Only unpaid orders can be cancelled; no refund flow |
| Cancel subscription | PRO simply expires after 30 days; no cancel endpoint |
| PRO seller analytics | Only admin analytics exists |
| Listing quota race | Two simultaneous publishes can exceed the FREE limit by one |
| Email retries | Failed emails are marked `FAILED` and not retried |
| Real payments | Mock only (`MOCK_CARD`, `MOCK_QR`); Stripe Test Mode not integrated |
| Public DRAFT listing detail | Earlier known issue from week 1, still open |

---

# 15. Next Steps

```text
1. Deploy Earn_TTB to thalamus and test in Swagger
2. Add SMTP settings to .env to start sending emails
3. Merge Earn_TTB into main after team review
4. Connect the Next.js frontend to the new endpoints
5. Optional: unpaid-order timeout, refunds, PRO seller analytics
```
