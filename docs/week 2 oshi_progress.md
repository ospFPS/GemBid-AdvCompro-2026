# GemBid Backend Progress — Week 2 (Oshi)

## Date

**5 October 2026**

## Project

**GemBid — Gem Marketplace & Auction Platform**

This document summarizes the work completed today for the backend, database, testing, security, and frontend preparation.

---

# 1. Main Goal Today

The main goal was to finish the remaining backend features, test the complete system, improve security, organize the project structure, and prepare the backend for frontend integration.

The backend roadmap had 10 remaining major topics:

```text
1. Order + Buy Now
2. Payment / Transaction
3. Shipping
4. Notification + Email
5. Admin Analytics
6. Report + Moderation
7. User Profile / Account Settings
8. Subscription / Monetization
9. Automatic Auction Closing
10. Testing + Security + Final Validation
```

Current result:

```text
1. Order + Buy Now                   ✅
2. Payment / Transaction             ✅
3. Shipping                          ✅
4. Notification + Email              ✅
5. Admin Analytics                   ✅
6. Report + Moderation               ✅
7. User Profile / Account Settings   ✅
8. Subscription / Monetization       ✅
9. Automatic Auction Closing         ✅
10. Testing + Security + Validation  ✅
```

The agreed backend roadmap is now feature-complete.

---

# 2. New Backend Features Added

## 2.1 Order + Buy Now

Fixed-price listings can now be purchased.

Main flow:

```text
Buyer
↓
Published Fixed-Price Listing
↓
Buy Now
↓
Order Created
↓
PENDING_PAYMENT
```

Main endpoint:

```text
POST /api/listings/{id}/buy
```

Order endpoints:

```text
GET  /api/orders
GET  /api/orders/{id}
POST /api/orders/{id}/pay
POST /api/orders/{id}/cancel
PUT  /api/orders/{id}/ship
POST /api/orders/{id}/confirm-delivery
```

The backend prevents the seller from buying their own listing.

---

## 2.2 Payment / Transaction System

A mock payment system was added for the university project.

Supported test payment methods include:

```text
MOCK_CARD
MOCK_QR
```

Payment flow:

```text
PENDING_PAYMENT
↓
Mock Payment
↓
Transaction = SUCCESS
↓
Order = PAID
```

The system also prevents the same paid order from being paid twice.

---

## 2.3 Shipping System

After payment, the seller can add shipping information.

Flow:

```text
PAID
↓
Seller Ships
↓
SHIPPED
↓
Buyer Confirms Delivery
↓
DELIVERED
```

Shipping data includes:

```text
courier
tracking_number
shipping_address
shipping_status
shipped_at
delivered_at
```

---

# 3. Notification + Email System

An in-app notification system was added.

Examples of notification events:

```text
OUTBID
AUCTION_WON
AUCTION_LOST
PAYMENT_SUCCESS
ORDER_SHIPPED
REPORT_REVIEWED
MODERATION_WARNING
```

Notification endpoints:

```text
GET /api/notifications
PUT /api/notifications/read-all
PUT /api/notifications/{id}/read
```

An email queue was also added.

Flow:

```text
Backend Event
↓
Create Notification
↓
Create Email Delivery
↓
Email Worker
↓
SMTP
```

If SMTP is not configured, emails stay queued.

Environment variables:

```env
SMTP_HOST=
SMTP_PORT=587
SMTP_FROM=
SMTP_USER=
SMTP_PASS=
```

Email is optional. In-app notifications continue to work without SMTP.

---

# 4. Admin Analytics

A new admin analytics endpoint was added:

```text
GET /api/admin/analytics
```

It can return information such as:

```text
total users
new users
seller status summary
total gems
listing status summary
auction status summary
total bids
order status summary
sales value
average order value
active PRO subscriptions
subscription revenue
email status
top sold gem types
daily sales
```

The endpoint is protected by:

```text
AuthMiddleware
↓
AdminMiddleware
↓
AdminAnalyticsHandler
```

---

# 5. Report + Moderation System

A complete report and moderation system was added.

New main file:

```text
backend/report.go
```

Users can report one target at a time:

```text
User
OR
Listing
OR
Auction
```

User endpoints:

```text
POST /api/reports
GET  /api/reports
```

Admin endpoints:

```text
GET /api/admin/reports
PUT /api/admin/reports/{id}
GET /api/admin/moderation-actions
```

Admin decisions:

```text
REJECT
OR
RESOLVE
```

Supported moderation actions:

```text
NONE
WARN_USER
SUSPEND_LISTING
SUSPEND_SELLER
```

Examples:

```text
Report Listing
↓
Admin Reviews
↓
RESOLVE
↓
SUSPEND_LISTING
↓
Listing = SUSPENDED
↓
Moderation action saved
↓
Reporter notified
```

The system also blocks:

```text
reporting yourself
reporting your own listing
reporting your own auction
multiple report targets in one request
reviewing the same report twice
```

---

# 6. User Profile / Account Settings

User profile support was completed.

Endpoints:

```text
GET /api/me
PUT /api/me
PUT /api/me/password
```

Users can update:

```text
username
full_name
phone
```

Users can also change their password after entering the current password.

Existing JWT tokens are not revoked immediately after a password change. They expire normally after 24 hours.

---

# 7. Username Authentication Added

The authentication system was changed from:

```text
email + password
```

to:

```text
username + password
```

This change was made to match the frontend design.

Registration now supports:

```json
{
  "username": "testuser",
  "password": "Test1234"
}
```

Optional registration fields:

```text
full_name
email
phone
```

If `full_name` is not provided, the backend uses the username as the default full name.

Login now uses:

```json
{
  "username": "testuser",
  "password": "Test1234"
}
```

Username rules:

```text
3–30 characters
lowercase
allowed:
a-z
0-9
.
_
-
```

Usernames are unique.

Existing users were automatically given usernames in this format:

```text
user_<user_id>
```

Example:

```text
user_3
user_4
```

The exact username-only registration and login flow was tested against Supabase and passed.

---

# 8. Subscription / Monetization

A Free / PRO subscription system was added.

Subscription endpoints:

```text
GET  /api/subscriptions/plans
GET  /api/subscriptions/me
POST /api/subscriptions/upgrade
```

Current logic:

```text
FREE
↓
Limited live listings
↓
Upgrade
↓
PRO
↓
Higher / unlocked listing limit
```

The FREE plan currently allows:

```text
5 live listings
```

The listing limit is protected against simultaneous publish requests.

Test case:

```text
Seller has 4 live listings

Request A → publish
Request B → publish
same time
```

Expected result:

```text
one request → success
one request → rejected
final live listings = 5
```

This concurrency test passed.

---

# 9. Automatic Auction Closing

Automatic auction closing was added.

Flow:

```text
Auction reaches end_time
↓
Background closer finds due auction
↓
Lock auction
↓
Check highest bid
↓
Check reserve
↓
Winner OR UNSOLD
↓
Listing status updated
↓
Winning order created
↓
Notifications created
```

Successful auction:

```text
Auction = ENDED
Winning Bid = WON
Listing = SOLD
Order = PENDING_PAYMENT
Winner notified
Seller notified
Other bidders notified
```

No winner / reserve not reached:

```text
Auction = ENDED
Listing = UNSOLD
```

---

# 10. Auction Concurrency Validation

The auction engine was tested again with simultaneous bids.

The backend uses:

```sql
SELECT ... FOR UPDATE
```

inside a PostgreSQL transaction.

This protects the auction from two requests updating the same price at the same time.

Test example:

```text
Bidder A → 5750
Bidder B → 6000
same time
```

Final result:

```text
highest price = 6000
highest bidder = bidder B
older accepted bid = OUTBID
```

The real Supabase concurrency test passed.

---

# 11. Database Migrations Added

New migrations added today:

```text
019_orders_payments_shipping_notifications_subscriptions.sql
020_reports_moderation_constraints.sql
021_add_username.sql
```

## Migration 019

Added or updated support for:

```text
orders
transactions
shipping
email deliveries
subscriptions
status constraints
indexes
```

## Migration 020

Added report/moderation validation:

```text
report status check
exactly one report target
moderation action type check
```

## Migration 021

Added:

```text
users.username
unique username index
username format validation
optional email
```

All required new migrations were validated and applied to the current Supabase database.

---

# 12. Security Improvements

A full security review was completed.

## JSON Validation

Requests now reject:

```text
unknown JSON fields
multiple JSON objects in one request
extra trailing JSON data
```

Example attack:

```json
{
  "username": "test",
  "password": "Password123",
  "role": "ADMIN"
}
```

Result:

```text
400 Bad Request
```

## JWT Security

JWT validation checks:

```text
valid token
HS256 only
expiration
valid user ID
```

Invalid or expired tokens return:

```text
401 Unauthorized
```

## Password Rules

Current password rules:

```text
minimum = 8
maximum = 72 bytes
bcrypt
```

## JWT Secret

Backend startup now requires:

```text
JWT_SECRET >= 32 bytes
```

Current local JWT secret was checked and is 64 characters.

The real `.env` is not tracked by Git.

---

# 13. Input Validation Improvements

Extra validation was added for database-sized fields.

Examples:

```text
username
full_name
email
phone
seller information
gem fields
price values
bid values
shipping address
courier
tracking number
report reason
report description
```

This prevents invalid or extremely large values from reaching PostgreSQL.

---

# 14. Full End-to-End API Test

A complete API journey was added and tested.

Main test file:

```text
backend/test/full_e2e_integration_test.go
```

Test flow:

```text
Register
↓
Login
↓
Profile
↓
Change Password
↓
Login With New Password
↓
Apply as Seller
↓
Admin Approval
↓
Verified Seller
↓
Create Gem
↓
Create Listing
↓
Publish Fixed Price
↓
Buy Now
↓
Order
↓
Payment
↓
Shipping
↓
Delivery
↓
Notifications
↓
PRO Subscription
↓
Admin Analytics
```

This full test passed against the real Supabase database.

---

# 15. Full Backend Test Suite

The integration tests are now organized under:

```text
backend/test/
```

Current test files:

```text
auction_integration_test.go
concurrency_integration_test.go
full_e2e_integration_test.go
migration_integration_test.go
orders_integration_test.go
report_integration_test.go
security_validation_test.go
```

Test coverage includes:

```text
Auction engine
Simultaneous bidding
Automatic auction closing
Subscription listing limit concurrency
Full API E2E
Orders
Payment
Shipping
Notifications
Subscriptions
Reports
Moderation
Migration validation
JWT validation
Invalid JSON
Username validation
Password validation
```

Full test result:

```text
PASS
```

---

# 16. Backend Project Structure Reorganized

The Go backend was reorganized so the test files could live inside a dedicated `test` folder.

Previous design:

```text
backend/
├── main.go
├── auth.go
├── auction.go
└── *_test.go
```

Current design:

```text
backend/
├── admin.go
├── auction.go
├── auth.go
├── database.go
├── gem.go
├── gem_handlers.go
├── listings.go
├── middleware.go
├── notifications.go
├── orders.go
├── report.go
├── seller.go
├── seller_middleware.go
├── server.go
├── subscriptions.go
│
├── cmd/
│   └── server/
│       └── main.go
│
├── test/
│   ├── auction_integration_test.go
│   ├── concurrency_integration_test.go
│   ├── full_e2e_integration_test.go
│   ├── migration_integration_test.go
│   ├── orders_integration_test.go
│   ├── report_integration_test.go
│   └── security_validation_test.go
│
└── db/
    └── migrations/
```

The backend package is now:

```go
package gembid
```

The executable entry point is:

```text
backend/cmd/server/main.go
```

New backend run command:

```powershell
cd "D:\RAI\adcompro proj\backend"
go run ./cmd/server
```

Normal tests:

```powershell
go test ./...
```

Full Supabase integration tests:

```powershell
$env:RUN_AUCTION_INTEGRATION="1"
go test ./test -count=1 -v
```

---

# 17. Docker Update

The Dockerfile was updated for the new Go structure.

It now builds:

```text
./cmd/server
```

instead of building the backend root directly.

Docker Compose configuration validation passed.

---

# 18. Swagger

Swagger remains available at:

```text
http://localhost:8080/docs
```

Swagger was updated for:

```text
User Profile
Password Change
Orders
Payment
Shipping
Notifications
Subscriptions
Admin Analytics
Reports
Moderation
Username Login
Username Registration
```

Important note:

The Go handler comments were removed from the backend source files.

The existing generated Swagger specification in:

```text
backend/docs/docs.go
```

still works.

Do not run:

```text
swag init
```

unless Swagger annotations are restored first.

---

# 19. Code Cleanup

Backend Go comments were removed to keep the source files cleaner.

The subscription file was checked again and already contained no comments.

Cleanup did not change backend behavior.

Validation after cleanup:

```text
go test ./...      ✅
go vet ./...       ✅
go build           ✅
git diff --check   ✅
```

---

# 20. Dependency / Vulnerability Checks

Go vulnerability scan:

```text
govulncheck ./...
```

Result:

```text
0 reachable vulnerabilities
```

Frontend security scan:

```text
npm audit
```

Result:

```text
0 vulnerabilities
```

---

# 21. Frontend Branch Preparation

The old frontend folder was removed.

Two teammate frontend versions were extracted from GitHub branches without changing the backend.

Current frontend folders:

```text
frontend-william
frontend-ryu
```

Sources:

```text
william-frontend branch
→ frontend-william

Frontend-Frien branch
→ frontend-ryu
```

The two folders do not contain separate nested Git repositories.

Both login/signup designs were aligned to use:

```text
username
password
```

William already used username.

Ryu was changed from email to username in the form labels/field names.

The frontend still needs full API wiring where the current pages are only UI forms.

---

# 22. Runtime Architecture

Current intended architecture:

```text
Browser
↓
Next.js Frontend
↓
Go REST Backend
↓
Supabase PostgreSQL
```

For a deployed version:

```text
Internet
↓
HTTPS
↓
Frontend / Reverse Proxy
↓
Go Backend
↓
Supabase
```

---

# 23. Current Run Commands

## Backend

```powershell
cd "D:\RAI\adcompro proj\backend"
go run ./cmd/server
```

Backend:

```text
http://localhost:8080
```

## William Frontend

```powershell
cd "D:\RAI\adcompro proj\frontend-william"
npm install
npm run dev -- -p 3000
```

Open:

```text
http://localhost:3000
```

## Ryu Frontend

```powershell
cd "D:\RAI\adcompro proj\frontend-ryu"
npm install
npm run dev -- -p 3001
```

Open:

```text
http://localhost:3001
```

---

# 24. Validation Completed Today

The backend passed:

```text
go test ./...                       ✅
go vet ./...                        ✅
go build                            ✅
Full Supabase integration tests     ✅
Auction concurrency test            ✅
Automatic auction closer test       ✅
Subscription quota race test        ✅
Full API E2E test                   ✅
Order / payment / shipping test     ✅
Report / moderation test            ✅
Security validation test            ✅
Migration validation                ✅
Swagger runtime check               ✅
Docker Compose validation           ✅
```

---

# 25. Current Backend Completion

For the agreed university project scope:

```text
Backend feature completion: 100%
```

Main backend systems:

```text
Authentication                       ✅
Username Login                       ✅
JWT                                  ✅
User Profile                         ✅
Password Change                      ✅
Seller Application                   ✅
Admin Seller Approval                ✅
Verified Seller Middleware           ✅
Gem CRUD                             ✅
Listings                             ✅
Fixed Price                          ✅
Auction Engine                       ✅
Concurrent Bidding                   ✅
Automatic Auction Closing            ✅
Buy Now                              ✅
Orders                               ✅
Payments                             ✅
Shipping                             ✅
Notifications                        ✅
Email Queue                          ✅
Subscriptions                        ✅
Admin Analytics                      ✅
Reports                              ✅
Moderation                           ✅
Swagger                              ✅
Docker Configuration                 ✅
Security Validation                  ✅
Integration Tests                    ✅
```

---

# 26. What Is Still Not Production-Level

The backend is complete for the course project, but some production-level work remains.

```text
Real payment gateway        not implemented
Real SMTP credentials       not configured
Public HTTPS/domain         not configured
Production monitoring       not configured
Rate limiting               not implemented
JWT token revocation        not implemented
Production backup plan      not configured
```

These are deployment/production improvements, not missing main project features.

---

# 27. Immediate Next Steps

The backend should now be treated as mostly frozen.

Recommended next work:

```text
1. Connect the selected frontend to the Go API
2. Test registration with username + password
3. Test login with username + password
4. Connect marketplace and auction pages
5. Connect seller pages
6. Connect orders and notifications
7. Deploy to teammate/friend server
8. Test with several real users
9. Run simultaneous auction test from different devices
10. Fix only bugs found during real use
```

Do not add unnecessary backend features unless the frontend or real-user test shows a clear missing requirement.

---

# 28. Final Week 2 Snapshot

```text
USERNAME REGISTER / LOGIN
        ↓
JWT
        ↓
NORMAL USER
        ↓
PROFILE
        ↓
SELLER APPLICATION
        ↓
ADMIN APPROVAL
        ↓
VERIFIED SELLER
        ↓
CREATE GEM
        ↓
CREATE LISTING
        ↓
┌───────────────────────┐
↓                       ↓
FIXED PRICE           AUCTION
↓                       ↓
BUY NOW               BIDDING
↓                       ↓
ORDER                 AUTO CLOSE
↓                       ↓
PAYMENT               WINNER
↓                       ↓
SHIPPING              ORDER
↓                       ↓
DELIVERY              PAYMENT
        ↓
NOTIFICATIONS
        ↓
REPORT / MODERATION
        ↓
ADMIN ANALYTICS
        ↓
TESTED / VALIDATED
```

---

# 29. Short Handoff Summary

Today the backend was moved from a mostly complete marketplace/auction backend to a complete university-project backend.

The main new work was:

```text
Order system
Payment system
Shipping system
Notification system
Email queue
Subscription system
Admin analytics
Report system
Moderation system
Automatic auction closing
Winner order creation
User profile
Username authentication
Security improvements
Full E2E testing
Concurrency testing
Test folder organization
Database migration updates
Swagger updates
Frontend branch preparation
```

The next main project phase is now:

```text
Frontend Integration
↓
Server Deployment
↓
Real Multi-User Testing
↓
Final Demo
```

