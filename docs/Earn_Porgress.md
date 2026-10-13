# GemBid — Earn Progress + Frontend API Guide

**Date:** 5 October 2026
**Branch:** `Earn_refactor` (from `main` @ `0ea8465`)

This document has two parts:

1. **Progress** — what changed in the backend this round (a cleanup, no API change).
2. **Frontend API Guide** — everything the frontend team needs to call the backend.

> **Frontend team:** jump to [Part 2](#part-2--frontend-api-guide).

---

# Part 1 — Progress

## 1.1 Summary

| Item | Status |
|---|---|
| Team's `main` (Oshi week 2) deployed to the Ubuntu server | ✅ |
| Migrations 019–021 confirmed applied on the database | ✅ |
| Backend cleanup — less duplicated code, **no API change** | ✅ |
| All 8 integration tests pass | ✅ |
| Live vs. refactored responses compared — identical | ✅ |
| Frontend API guide (this document, Part 2) | ✅ |

## 1.2 Backend cleanup

The same behaviour in **405 fewer lines** (597 removed, 192 added across 12 files).
Endpoints, request bodies, responses, status codes and error messages are unchanged.

| Change | Where |
|---|---|
| `decodeJSON` writes the `400 Invalid JSON` itself; callers are one line | 17 handlers |
| All JSON responses go through `writeJSON` | 27 places |
| `pathID` for `{id}` parsing where the check was already identical | 7 handlers |
| Removed impossible "Unauthorized" checks after the auth middleware | 15 handlers |
| Shared `userColumns` + `scanUser` instead of 4 copies | `auth.go`, `middleware.go` |
| Shared `auctionColumns`, `gemColumns` instead of 7 copies | `auction.go`, `gem*.go` |
| Auction activation SQL reused instead of 2 pasted copies | `auction.go` |
| `pgx.CollectRows` instead of 5 hand-written row loops (also checks row errors) | `auction.go`, `admin.go`, `listings.go`, `gem_handlers.go` |
| Username regex compiled once; `strings.Cut` for the auth header (still accepts `Bearer` in any case); dead JWT checks removed | `auth.go`, `middleware.go` |

Deliberately **not** changed (they would change visible behaviour): report creation and
moderation logic, the two report list response shapes, and the package layout.

## 1.3 Verification

**Integration tests** (real database, test data cleaned up afterwards):

```text
--- PASS: TestAuctionEngineTwoBidderFlow
--- PASS: TestAutomaticAuctionCloser
--- PASS: TestFreeListingQuotaConcurrentPublish
--- PASS: TestFullAPIEndToEnd
--- PASS: TestReportMigration020Applies
--- PASS: TestOrderPaymentShippingFlow
--- PASS: TestReportModerationFlow
--- PASS: TestSecurityValidationGuards
ok  adcompro-backend/test  67.190s
```

**Side-by-side check:** the refactored build ran on port 8081 next to the live server
on 8080. 28 requests (public lists and details, missing and invalid IDs, every auth header
variant, malformed JSON, bad credentials, validation errors) returned **identical status,
headers and body** on both. Logged-in `/api/me` with `Bearer`, `bearer` and `BEARER`
also matched.

Static checks: `go build`, `go vet`, `gofmt` ✅

## 1.4 Deployment state

- The server currently runs `main` @ `0ea8465`.
- `Earn_refactor` is pushed for review. Once merged into `main`, redeploy with
  `git checkout main && git pull && docker compose up -d --build`.
- No database change is needed.

---

# Part 2 — Frontend API Guide

## 2.1 Connecting to the backend

| Environment | Backend URL |
|---|---|
| Local backend | `http://localhost:8080` |
| Team server (Tailscale) | `http://thalamus:8080` |
| Public (ngrok) | `https://angular-snorkel-passive.ngrok-free.dev` |
| Interactive docs (Swagger) | `<backend>/docs` |

### ⚠️ Use a Next.js proxy — the backend does not send CORS headers

A browser `fetch` from `localhost:3000` straight to `:8080` **will be blocked**: the
backend answers the CORS preflight with `405`. Proxy instead, so the browser only
talks to the Next.js origin. In `Frontend/next.config.js`:

```js
const path = require("node:path");

/** @type {import('next').NextConfig} */
const nextConfig = {
  outputFileTracingRoot: path.join(__dirname),
  async rewrites() {
    const backend = process.env.BACKEND_URL ?? "http://localhost:8080";
    return [{ source: "/api/:path*", destination: `${backend}/api/:path*` }];
  },
};

module.exports = nextConfig;
```

Then call **relative** paths from the frontend: `fetch("/api/login", ...)`.
Set `BACKEND_URL` in `.env.local` to switch servers (for example `http://thalamus:8080`).

> Don't create Next.js API routes under `/api/...`; they would shadow the proxy.
> If you call the ngrok URL directly (not through the proxy), add the header
> `ngrok-skip-browser-warning: 1`, or ngrok returns its HTML warning page.

## 2.2 Conventions

| Topic | Rule |
|---|---|
| Request body | JSON. **Unknown fields are rejected** (`400 Invalid JSON`), so send only the documented fields. |
| Success response | JSON |
| **Error response** | **Plain text**, not JSON, e.g. `Listing not found`. Read it with `await res.text()`. |
| Dates | ISO 8601 strings, e.g. `2026-10-05T11:18:26Z`. Send auction times the same way. |
| Money | JSON numbers in THB with 2 decimals, e.g. `12500` or `5250.5` |
| `null` | Optional fields come back as `null` (e.g. `email`, `phone`, `reserve_price`) |
| IDs | Positive integers. `0`, negative or non-numeric IDs → `400` or `404` |

Common status codes:

| Code | Meaning | Frontend action |
|---|---|---|
| 200 / 201 | OK / created | |
| 400 | Invalid input (message says why) | Show the message on the form |
| 401 | Not logged in / token expired | Clear the token, go to login |
| 403 | Logged in but not allowed (not seller/admin/owner, or plan limit) | Show the message |
| 404 | Not found | |
| 409 | Conflict with current state (already sold, already reviewed, bid too low…) | Show the message, refresh the data |

## 2.3 Authentication

1. `POST /api/login` returns `token`.
2. Store it (e.g. `localStorage`) and send it on every protected call:
   `Authorization: Bearer <token>`
3. The token lasts **24 hours**. There is no refresh endpoint, so on `401` send the user to login.
4. `GET /api/me` returns the current user. Use `role` and `seller_status` to decide what to show:

| User type | How to detect | Can use |
|---|---|---|
| Visitor | no token | Public endpoints |
| User | `role = "USER"` | + buying, bidding, orders, notifications, reports, profile, subscription |
| Verified seller | `seller_status = "VERIFIED"` | + gems, listings, auctions, shipping |
| Admin | `role = "ADMIN"` | + admin endpoints |

### Minimal fetch helper

```ts
export async function api<T>(path: string, options: RequestInit = {}): Promise<T> {
  const token = typeof window !== "undefined" ? localStorage.getItem("token") : null;
  const res = await fetch(path, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...options.headers,
    },
  });
  if (res.status === 401) localStorage.removeItem("token");
  if (!res.ok) throw new Error(await res.text()); // errors are plain text
  return res.json();
}

// const { token, user } = await api<{ token: string; user: User }>("/api/login",
//   { method: "POST", body: JSON.stringify({ username, password }) });
```

## 2.4 Endpoint reference

**Auth column:** 🌐 public · 👤 logged in · 💎 verified seller · 🛡️ admin

### Account

| Method | Path | Auth | Body | Response |
|---|---|---|---|---|
| POST | `/api/register` | 🌐 | `{username, password, full_name?, email?, phone?}` | `201 {message, user}` |
| POST | `/api/login` | 🌐 | `{username, password}` | `{message, token, user}` |
| GET | `/api/me` | 👤 | | `User` |
| PUT | `/api/me` | 👤 | `{full_name, username?, phone?}` | `{message, user}` |
| PUT | `/api/me/password` | 👤 | `{current_password, new_password}` | `{message}` |

Rules:
- `username`: 3–30 characters of `a-z 0-9 . _ -`, stored in lowercase.
- `password`: 8–72 characters.
- `full_name` defaults to the username; max lengths are full_name 150, email 255, phone 30.
- On `PUT /api/me`, an empty `phone` clears it and an empty `username` keeps the current one.

```ts
type User = {
  user_id: number; username: string; full_name: string;
  email: string | null; phone: string | null;
  role: "USER" | "ADMIN";
  seller_status: "NONE" | "PENDING" | "VERIFIED" | "REJECTED" | "SUSPENDED";
  created_at: string;
};
```

### Becoming a seller

| Method | Path | Auth | Body | Response |
|---|---|---|---|---|
| POST | `/api/seller/apply` | 👤 | `{shop_name, legal_name, seller_type?, address?, description?}` | `201 {message, application}` |
| GET | `/api/seller/status` | 👤 | | `{seller_status, application}` (`application` is `null` if never applied) |
| GET | `/api/admin/seller-applications` | 🛡️ | | `{applications, count}` |
| PUT | `/api/admin/seller-applications/{id}` | 🛡️ | `{decision: "APPROVE" \| "REJECT"}` | `{message, seller_id, user_id, decision, verification_status, reviewed_by}` |

Flow: apply → `PENDING` → admin approves → `VERIFIED` (seller pages unlock). Only `PENDING` applications can be reviewed.

### Gems (seller inventory)

| Method | Path | Auth | Body | Response |
|---|---|---|---|---|
| GET | `/api/gems` | 🌐 | | `Gem[]` (all gems, newest first) |
| GET | `/api/gems/{id}` | 🌐 | | `Gem` |
| POST | `/api/gems` | 💎 | `GemInput` | `201 {message, gem}` |
| PUT | `/api/gems/{id}` | 💎 owner | `GemInput` (full replace) | `{message, gem}` |
| DELETE | `/api/gems/{id}` | 💎 owner | | `{message}` |

```ts
type GemInput = {
  gem_name: string; gem_type: string;            // required
  carat?: number; length?: number; width?: number; height?: number; // > 0
  shape?: string; color?: string; clarity?: string;
  origin?: string; treatment?: string; description?: string;
};
// Gem = GemInput fields (optional ones may be null) + gem_id, seller_id, created_at, updated_at
```

A gem **cannot be edited** while it has a published, active, sold or suspended listing (`409`),
and cannot be deleted while a listing references it (`409`).

### Listings

| Method | Path | Auth | Body | Response |
|---|---|---|---|---|
| GET | `/api/listings` | 🌐 | | `ListingItem[]` (only `PUBLISHED`/`ACTIVE`) |
| GET | `/api/listings/{id}` | 🌐 | | `Listing` |
| POST | `/api/listings` | 💎 | `{gem_id}` | `201 {message, listing}` (status `DRAFT`) |
| PUT | `/api/listings/{id}/publish` | 💎 owner | see below | `{message, listing}` |
| DELETE | `/api/listings/{id}` | 💎 owner | | `{message}` (only while `DRAFT`) |

Publish body:
- Fixed price: `{"sale_type": "FIXED_PRICE", "fixed_price": 12500}` → listing becomes `PUBLISHED`.
- Auction: `{"sale_type": "AUCTION"}` → stays `DRAFT` until you call `POST /api/auctions`.

```ts
type ListingItem = { listing_id: number; gem_id: number; gem_name: string; gem_type: string;
  sale_type: "FIXED_PRICE" | "AUCTION" | null; status: ListingStatus; fixed_price: number | null };
type Listing = { listing_id: number; gem_id: number; seller_id: number;
  sale_type: "FIXED_PRICE" | "AUCTION" | null; status: ListingStatus; fixed_price: number | null;
  created_at: string; published_at: string | null; updated_at: string;
  relisted_from_listing_id: number | null };
```

**FREE plan limit:** a seller on FREE can have **5 live listings**. The 6th publish (fixed price or
auction) returns `403` with an upgrade message. Show an "Upgrade to PRO" button.

### Auctions and bids

| Method | Path | Auth | Body | Response |
|---|---|---|---|---|
| GET | `/api/auctions` | 🌐 | | `AuctionItem[]` (`SCHEDULED` + `ACTIVE`, active first) |
| GET | `/api/auctions/{id}` | 🌐 | | `Auction` |
| GET | `/api/auctions/{id}/bids` | 🌐 | | `Bid[]` (newest first) |
| POST | `/api/auctions` | 💎 owner | see below | `201 {message, auction}` |
| POST | `/api/auctions/{id}/bids` | 👤 | `{bid_amount}` | `201 {message, bid, current_price, next_minimum_bid}` |
| POST | `/api/auctions/{id}/close` | 💎 owner | | `{message, auction, order_id?}` |

Create body (listing must be a `DRAFT` with `sale_type = "AUCTION"`):

```json
{
  "listing_id": 12,
  "starting_price": 5000,
  "minimum_increment": 250,
  "reserve_price": 6000,
  "start_time": "2026-10-06T09:00:00Z",
  "end_time": "2026-10-07T09:00:00Z"
}
```

`reserve_price` is optional (it can be `null`); if given, it must be ≥ `starting_price`. `end_time` must be after `start_time` and in the future.

Bidding rules:
- First bid ≥ `starting_price`; after that ≥ `current_price + minimum_increment`. Show `next_minimum_bid` from the last bid response.
- A seller cannot bid on their own auction (`403`). Bids before start or after end return `409`.
- The previous leader gets an `OUTBID` notification.

**Closing is automatic.** The backend closes ended auctions every 30 seconds. If the reserve is met, the winner gets a `PENDING_PAYMENT` order and an `AUCTION_WON` notification; otherwise the listing becomes `UNSOLD`. The `close` endpoint is only a manual shortcut.

For a live view, poll `GET /api/auctions/{id}` every few seconds. There is no websocket.

```ts
type Auction = { auction_id: number; listing_id: number; starting_price: number;
  minimum_increment: number; reserve_price: number | null; current_price: number;
  current_highest_bidder_id: number | null; start_time: string; end_time: string;
  status: "SCHEDULED" | "ACTIVE" | "ENDED"; winner_id: number | null;
  winning_bid_id: number | null; created_at: string; closed_at: string | null };
type AuctionItem = { auction_id: number; listing_id: number; gem_id: number; gem_name: string;
  gem_type: string; starting_price: number; minimum_increment: number; current_price: number;
  current_highest_bidder_id: number | null; start_time: string; end_time: string; status: string };
type Bid = { bid_id: number; auction_id: number; bidder_id: number; bid_amount: number;
  bid_time: string; bid_status: "ACCEPTED" | "OUTBID" | "WON" | "RESERVE_NOT_MET" };
```

### Orders, payment and shipping

| Method | Path | Auth | Body | Response |
|---|---|---|---|---|
| POST | `/api/listings/{id}/buy` | 👤 | | `201 Order` (Buy Now on a `PUBLISHED` fixed-price listing) |
| GET | `/api/orders` | 👤 | | `Order[]` I bought; `?as=seller` → orders on my shop |
| GET | `/api/orders/{id}` | 👤 buyer/seller/admin | | `OrderDetail` |
| POST | `/api/orders/{id}/pay` | 👤 buyer | `{payment_method, shipping_address}` | `OrderDetail` |
| POST | `/api/orders/{id}/cancel` | 👤 buyer | | `OrderDetail` (unpaid only) |
| PUT | `/api/orders/{id}/ship` | 💎 seller | `{courier, tracking_number}` | `OrderDetail` |
| POST | `/api/orders/{id}/confirm-delivery` | 👤 buyer | | `OrderDetail` |

```text
Buy Now / auction win → PENDING_PAYMENT ──pay──→ PAID ──ship──→ SHIPPED ──confirm──→ DELIVERED
                              └──cancel──→ CANCELLED (listing becomes available again)
```

- `payment_method` is `"MOCK_CARD"` or `"MOCK_QR"`. Payment is simulated; no real money moves.
- `shipping_address` is required (max 5000 characters); `courier` max 100, `tracking_number` max 150.
- Buying your own listing returns `403`; buying a listing that's already sold returns `409`.

```ts
type Order = { order_id: number; listing_id: number; buyer_id: number; seller_id: number;
  auction_id: number | null; winning_bid_id: number | null; item_price: number;
  shipping_fee: number; total_amount: number;
  status: "PENDING_PAYMENT" | "PAID" | "SHIPPED" | "DELIVERED" | "CANCELLED";
  created_at: string; updated_at: string };
type OrderDetail = {
  order: Order;
  shipping: { shipping_id: number; courier: string | null; tracking_number: string | null;
    shipping_address: string; status: "PENDING" | "SHIPPED" | "DELIVERED";
    shipped_at: string | null; delivered_at: string | null; updated_at: string } | null;
  transactions: { transaction_id: number; order_id: number | null; subscription_id: number | null;
    amount: number; payment_method: string | null; status: "PENDING" | "SUCCESS" | "FAILED" | "REFUNDED";
    reference_number: string | null; created_at: string; paid_at: string | null }[];
};
```

### Notifications

| Method | Path | Auth | Response |
|---|---|---|---|
| GET | `/api/notifications` | 👤 | `{notifications, unread_count}` (newest 100; `?unread=true` for unread only) |
| PUT | `/api/notifications/{id}/read` | 👤 | `{message}` |
| PUT | `/api/notifications/read-all` | 👤 | `{message, updated}` |

```ts
type Notification = { notification_id: number; type: NotificationType; title: string;
  message: string; auction_id: number | null; listing_id: number | null;
  order_id: number | null; is_read: boolean; created_at: string };
```

Use `order_id` / `auction_id` / `listing_id` to link a notification to its page. Poll
`unread_count` for the bell badge.

| `type` | Sent to | When |
|---|---|---|
| `ORDER_CREATED` | seller | Buy Now order placed |
| `PAYMENT_RECEIVED` | seller | Buyer paid |
| `PAYMENT_SUCCESS` | buyer | Payment done |
| `ORDER_SHIPPED` | buyer | Seller shipped (message includes the tracking number) |
| `ORDER_DELIVERED` | seller | Buyer confirmed delivery |
| `ORDER_CANCELLED` | seller | Buyer cancelled an unpaid order |
| `OUTBID` | bidder | Someone bid higher |
| `AUCTION_WON` | winner | Won; link to the order to pay |
| `AUCTION_SOLD` | seller | Auction ended with a winner |
| `AUCTION_LOST` | other bidders | Auction ended |
| `AUCTION_UNSOLD` | seller | No winner / reserve not met |
| `AUCTION_RESERVE_NOT_MET` | top bidder | Highest bid below reserve |
| `SUBSCRIPTION_ACTIVATED` | user | PRO active |
| `MODERATION_WARNING` | reported user | Admin warning |
| `REPORT_REVIEWED` | reporter | Their report was handled |

### Subscription (FREE / PRO)

| Method | Path | Auth | Body | Response |
|---|---|---|---|---|
| GET | `/api/subscriptions/plans` | 🌐 | | `Plan[]` |
| GET | `/api/subscriptions/me` | 👤 | | `{plan: "FREE", subscription: null}` or `{plan: "PRO", subscription}` |
| POST | `/api/subscriptions/upgrade` | 👤 | `{payment_method}` | `{message, subscription, transaction}` |

| Plan | Price | Live listings |
|---|---|---|
| FREE | ฿0 | 5 |
| PRO | ฿299 / 30 days | Unlimited |

Paying again while PRO is active adds 30 days. Render the plans page from `GET /api/subscriptions/plans` rather than hard-coding prices.

### Reports and moderation

| Method | Path | Auth | Body | Response |
|---|---|---|---|---|
| POST | `/api/reports` | 👤 | `{reason, description?, ` **one of** `reported_user_id \| listing_id \| auction_id}` | `201 Report` |
| GET | `/api/reports` | 👤 | | `Report[]` (my reports) |
| GET | `/api/admin/reports` | 🛡️ | | `{reports, count}`; `?status=PENDING` (default) `\|RESOLVED\|REJECTED\|ALL` |
| PUT | `/api/admin/reports/{id}` | 🛡️ | `{decision, action?, reason?}` | `{message, report, action}` |
| GET | `/api/admin/moderation-actions` | 🛡️ | | `ModerationAction[]` |

- `reason`: required, max 150 characters. `description`: max 2000.
- You can't report yourself, your own listing or your own auction (`409`).
- `decision`: `"RESOLVE"` or `"REJECT"`. `action` (only with RESOLVE): `"NONE"` (default), `"WARN_USER"`, `"SUSPEND_LISTING"`, `"SUSPEND_SELLER"`.
- Report `status`: `PENDING`, `RESOLVED` or `REJECTED`.

> Note the shapes differ: `GET /api/reports` returns a **plain array**, while `GET /api/admin/reports` returns `{reports, count}`.

### Admin analytics

`GET /api/admin/analytics` 🛡️ returns one object for the dashboard:

```text
generated_at
users                     { total, admins, new_last_30_days, sellers_by_status: {NONE: n, ...} }
gems_total
listings_by_status        { PUBLISHED: n, SOLD: n, ... }
auctions_by_status        { ACTIVE: n, ENDED: n, ... }
bids_total
orders_by_status          { PAID: n, DELIVERED: n, ... }
sales                     { paid_orders, gross_merchandise_value, average_order_value }
subscriptions             { active_pro, revenue }
emails_by_status          { PENDING: n, SENT: n, ... }
top_gem_types_sold        [ { gem_type, orders, revenue } ]   (top 5)
daily_sales_last_30_days  [ { day, orders, revenue } ]        (only days with sales)
```

The `*_by_status` objects only contain statuses that have rows, so treat missing keys as 0.

### Other

| Method | Path | Auth | Response |
|---|---|---|---|
| GET | `/health` | 🌐 | `{status: "ok"}` |
| GET | `/docs` | 🌐 | Swagger UI |

## 2.5 Status values

```ts
type ListingStatus = "DRAFT" | "PUBLISHED" | "ACTIVE" | "SOLD" | "UNSOLD" | "EXPIRED" | "CANCELLED" | "SUSPENDED";
```

| Listing status | Meaning |
|---|---|
| `DRAFT` | Created, not visible to buyers |
| `PUBLISHED` | Fixed-price listing for sale, or auction scheduled for the future |
| `ACTIVE` | Auction running |
| `SOLD` | Bought (Buy Now) or auction won |
| `UNSOLD` | Auction ended with no winner, or the winner cancelled |
| `SUSPENDED` | Removed by an admin |

## 2.6 Page → endpoint cheat sheet

| Page | Calls |
|---|---|
| Login / Register | `POST /api/login`, `POST /api/register` |
| Navbar | `GET /api/me`, `GET /api/notifications?unread=true` |
| Marketplace | `GET /api/listings`, `GET /api/auctions` |
| Listing detail | `GET /api/listings/{id}`, `GET /api/gems/{gem_id}`, `POST /api/listings/{id}/buy` |
| Auction detail | `GET /api/auctions/{id}`, `GET /api/auctions/{id}/bids`, `POST /api/auctions/{id}/bids` |
| Checkout | `GET /api/orders/{id}`, `POST /api/orders/{id}/pay` |
| My orders | `GET /api/orders`, `POST /api/orders/{id}/confirm-delivery`, `POST /api/orders/{id}/cancel` |
| Profile | `GET/PUT /api/me`, `PUT /api/me/password` |
| Become a seller | `GET /api/seller/status`, `POST /api/seller/apply` |
| Seller dashboard | `POST/PUT/DELETE /api/gems`, `POST /api/listings`, `PUT /api/listings/{id}/publish`, `POST /api/auctions`, `GET /api/orders?as=seller`, `PUT /api/orders/{id}/ship` |
| Pricing | `GET /api/subscriptions/plans`, `GET /api/subscriptions/me`, `POST /api/subscriptions/upgrade` |
| Report button | `POST /api/reports` |
| Admin | `/api/admin/*` endpoints above |

## 2.7 Known gaps (tell the backend team if you need these)

| Gap | Workaround for now |
|---|---|
| No "my listings" endpoint for sellers; `GET /api/listings` only shows published or active ones | Keep the `listing_id` from create/publish responses |
| `GET /api/gems` returns every seller's gems | Filter by `seller_id` on the frontend |
| `GET /api/listings/{id}` also returns `DRAFT` listings to anyone | Don't link to drafts publicly |
| No search, filters or pagination on lists | Filter on the frontend (lists are small) |
| No image upload | Placeholder images |
| No CORS | Use the Next.js proxy (section 2.1) |
| No refresh token (24h expiry) | Re-login on `401` |
