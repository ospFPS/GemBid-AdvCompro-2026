# GemBid Backend Progress Handoff

## Project
**GemBid — Gem Marketplace & Auction Platform**

This document is a teammate handoff for the backend work completed so far.

> **Latest status (27 September 2026):** Gem CRUD and fixed-price listings are now implemented, and the core **Auction Engine** has also been built and integration-tested.  
> Older sections below that say to stop before the Auction Engine are historical snapshots. See **Section 39 — Latest Update** for the current state.

---

# 1. Current Architecture

The original proposal allows a `FastAPI/GO Backend`, but the current implementation uses **Go**.

```text
Next.js Frontend
        ↓
HTTPS / REST API
        ↓
Go Backend
        ↓
Supabase PostgreSQL
```

Development is currently done on Windows.

Main project path:

```text
D:\RAI\adcompro proj
```

Backend path:

```text
D:\RAI\adcompro proj\backend
```

Environment file:

```text
D:\RAI\adcompro proj\.env
```

Database migration path:

```text
D:\RAI\adcompro proj\backend\db\migrations
```

Planned deployment later:

```text
Windows development PC
        ↓
Docker
        ↓
Ubuntu server
        ↓
Tailscale
        ↓
Supabase PostgreSQL
```

---

# 2. Main Technology Stack Being Used

| Layer | Current Technology |
|---|---|
| Frontend | Next.js / React |
| Backend | Go |
| Database | PostgreSQL |
| Database hosting | Supabase |
| DB driver | pgx |
| Authentication | JWT |
| Password hashing | bcrypt |
| Environment variables | godotenv |
| Deployment later | Docker + Ubuntu + Tailscale |
| API testing | PowerShell / later Postman if needed |

Important Go packages already used:

```text
github.com/jackc/pgx/v5
github.com/jackc/pgx/v5/pgxpool
github.com/joho/godotenv
github.com/golang-jwt/jwt/v5
golang.org/x/crypto/bcrypt
```

---

# 3. Current Backend Folder Structure

Current/expected structure:

```text
D:\RAI\adcompro proj\
│
├── .env
├── .gitignore
│
└── backend\
    ├── main.go
    ├── database.go
    ├── auth.go
    ├── middleware.go
    ├── seller.go
    ├── seller_middleware.go
    ├── admin.go
    ├── gem.go
    ├── go.mod
    ├── go.sum
    │
    └── db\
        └── migrations\
            ├── 001_create_users.sql
            ├── 002_create_seller_profiles.sql
            ├── 003_create_seller_documents.sql
            ├── 004_create_gems.sql
            ├── 005_create_gem_images.sql
            ├── 006_create_certificates.sql
            ├── 007_create_listings.sql
            ├── 008_create_auctions.sql
            ├── 009_create_bids.sql
            ├── 010_create_orders.sql
            ├── 011_create_transactions.sql
            ├── 012_create_shipping.sql
            ├── 013_create_notifications.sql
            ├── 014_create_email_deliveries.sql
            ├── 015_create_reports.sql
            ├── 016_create_moderation_actions.sql
            ├── 017_create_subscriptions.sql
            ├── 018_create_indexes_and_triggers.sql
            └── 003_to_018_combined.sql
```

---

# 4. Environment Variables

The project uses:

```env
DATABASE_URL=...
PORT=8080
JWT_SECRET=...
```

Important:

- Do not commit `.env`.
- `.gitignore` should include:

```text
.env
```

- Do not send the real `DATABASE_URL` or `JWT_SECRET` in chat or commit them to GitHub.
- `JWT_SECRET` is used to sign and verify login tokens.

---

# 5. Database Status

The project database is hosted in Supabase PostgreSQL.

A Go connection test was completed successfully before moving to the real backend.

Confirmed connection flow:

```text
Go
 ↓
pgx / pgxpool
 ↓
Supabase PostgreSQL
```

The backend connection uses `DATABASE_URL`.

## Main tables

The schema currently includes:

```text
users
seller_profiles
seller_documents
gems
gem_images
certificates
listings
auctions
bids
notifications
email_deliveries
orders
transactions
shipping
reports
moderation_actions
subscriptions
```

Total main application tables:

```text
17
```

Migration `018` creates indexes, triggers, and the `updated_at` trigger function rather than another application table.

---

# 6. Important Database Relationships

Simplified relationship:

```text
USERS
│
├── SELLER_PROFILES
│   │
│   ├── SELLER_DOCUMENTS
│   │
│   └── GEMS
│       │
│       ├── GEM_IMAGES
│       ├── CERTIFICATES
│       └── LISTINGS
│           │
│           ├── AUCTIONS
│           │   └── BIDS
│           │
│           └── ORDERS
│               ├── TRANSACTIONS
│               └── SHIPPING
│
├── BIDS
├── ORDERS
├── NOTIFICATIONS
├── REPORTS
└── SUBSCRIPTIONS
```

---

# 7. Authentication System

Authentication is implemented with:

```text
email + password
      ↓
bcrypt password hash
      ↓
public.users
      ↓
login
      ↓
bcrypt verification
      ↓
JWT
```

## Main authentication endpoints

```text
POST /api/register
POST /api/login
GET  /api/me
```

## Register

Expected request:

```json
{
  "full_name": "Test User",
  "email": "test@example.com",
  "phone": "0812345678",
  "password": "Test1234"
}
```

The backend:

1. Parses JSON.
2. Trims input.
3. Converts email to lowercase.
4. Checks required fields.
5. Checks password length.
6. Checks for duplicate email.
7. Hashes password using bcrypt.
8. Inserts the account into `public.users`.
9. Lets the database default to:

```text
role = USER
seller_status = NONE
```

Important:

The registration API does **not** accept a client-provided role.  
A user cannot register themselves as `ADMIN`.

## Login

Expected request:

```json
{
  "email": "test@example.com",
  "password": "Test1234"
}
```

The backend:

1. Finds user by email.
2. Loads stored `password_hash`.
3. Compares password using bcrypt.
4. Creates JWT.
5. Returns token and basic user data.

JWT includes data similar to:

```text
sub
user_id
role
seller_status
iat
exp
```

Current token lifetime:

```text
24 hours
```

## Protected requests

Protected requests use:

```text
Authorization: Bearer <JWT>
```

`AuthMiddleware` verifies the JWT and places:

```text
user_id
```

into the Go request context.

The backend then gets the authenticated user ID from the context rather than trusting a `user_id` sent by the frontend.

---

# 8. Seller System

Current seller flow:

```text
Normal User
   ↓
Apply to Become Seller
   ↓
PENDING
   ↓
Admin Review
   ↓
APPROVE / REJECT
   ↓
VERIFIED / REJECTED
```

## Seller endpoints

```text
POST /api/seller/apply
GET  /api/seller/status
```

## Admin seller endpoints

```text
GET /api/admin/seller-applications
PUT /api/admin/seller-applications/{id}
```

---

# 9. Seller Application

`POST /api/seller/apply`

Expected request:

```json
{
  "shop_name": "Gem Test Shop",
  "seller_type": "Individual",
  "legal_name": "Seller Test",
  "address": "Bangkok, Thailand",
  "description": "Test gemstone seller"
}
```

The backend:

1. Requires a valid JWT.
2. Gets `user_id` from authentication context.
3. Starts a database transaction.
4. Loads the current user's role and seller status.
5. Requires:

```text
role = USER
seller_status = NONE
```

6. Creates a row in `seller_profiles`.
7. Updates:

```text
users.seller_status = PENDING
```

8. Commits the transaction.

The database transaction is used so both operations succeed together.

Expected result:

```text
users.seller_status = PENDING

seller_profiles.verification_status = PENDING
```

Duplicate applications are blocked because a user whose status is no longer `NONE` cannot apply again.

---

# 10. Seller Status

`GET /api/seller/status`

Requires JWT.

Possible seller states:

```text
NONE
PENDING
VERIFIED
REJECTED
SUSPENDED
```

If the user has never applied:

```json
{
  "seller_status": "NONE",
  "application": null
}
```

If the user has applied, the API returns the status and seller profile.

---

# 11. Admin Authorization

Admin routes are protected by two middleware layers:

```text
AuthMiddleware
      ↓
AdminMiddleware
      ↓
Admin Handler
```

`AdminMiddleware`:

1. Gets authenticated `user_id`.
2. Queries the database.
3. Loads the user's current role.
4. Requires:

```text
role = ADMIN
```

5. Returns:

```text
403 Forbidden
```

for non-admin users.

The admin role is checked from the database rather than relying only on the role that was placed in the JWT when the user logged in.

---

# 12. Admin Seller Review

Admin list endpoint:

```text
GET /api/admin/seller-applications
```

Returns seller applications with fields such as:

```text
seller_id
user_id
full_name
email
shop_name
seller_type
legal_name
address
description
verification_status
applied_at
reviewed_at
reviewed_by
```

Review endpoint:

```text
PUT /api/admin/seller-applications/{id}
```

Approve request:

```json
{
  "decision": "APPROVE"
}
```

Reject request:

```json
{
  "decision": "REJECT"
}
```

Backend conversion:

```text
APPROVE → VERIFIED
REJECT  → REJECTED
```

The review uses a database transaction and updates both:

```text
seller_profiles.verification_status
users.seller_status
```

For approval:

```text
seller_profiles.verification_status = VERIFIED
users.seller_status                  = VERIFIED
```

For rejection:

```text
seller_profiles.verification_status = REJECTED
users.seller_status                  = REJECTED
```

It also sets:

```text
reviewed_at = NOW()
reviewed_by = admin user_id
```

Only applications currently in:

```text
PENDING
```

can be reviewed.

---

# 13. Mock Seller Documents Decision

The proposal originally includes seller verification documents.

For the current mock/university version, we decided:

```text
Do not implement actual seller document upload now.
```

The `seller_documents` table is kept in the database but is currently unused.

Nothing is deleted.

This lets the mock seller flow remain:

```text
Apply
↓
PENDING
↓
Admin Approve / Reject
```

without needing real identity-file storage.

---

# 14. Current Main Routes

The current `main.go` should expose:

```text
GET  /health

POST /api/register
POST /api/login
GET  /api/me

POST /api/seller/apply
GET  /api/seller/status

GET  /api/admin/seller-applications
PUT  /api/admin/seller-applications/{id}

POST /api/gems
```

---

# 15. Gem CRUD — Current Work

We have now started the **Gem CRUD** stage.

The proposal expects:

```text
POST   /api/gems
GET    /api/gems/{id}
PUT    /api/gems/{id}
DELETE /api/gems/{id}
```

Current status:

```text
POST   /api/gems       STARTED / IMPLEMENTED IN CURRENT CODE
GET    /api/gems/{id}  NOT YET DONE
PUT    /api/gems/{id}  NOT YET DONE
DELETE /api/gems/{id}  NOT YET DONE
```

---

# 16. Verified Seller Middleware

A dedicated middleware is used for seller-only functionality.

Security chain:

```text
POST /api/gems
      ↓
AuthMiddleware
      ↓
JWT valid?
      ↓
VerifiedSellerMiddleware
      ↓
Verified seller?
      ↓
CreateGemHandler
```

The middleware checks both:

```text
users.seller_status = VERIFIED
```

and:

```text
seller_profiles.verification_status = VERIFIED
```

If both are valid, it retrieves:

```text
seller_id
```

and places it into the request context.

This prevents a normal user from directly calling seller APIs.

---

# 17. Create Gem Endpoint

Current route:

```text
POST /api/gems
```

Only verified sellers can call it.

Expected request:

```json
{
  "gem_name": "Blue Sapphire",
  "gem_type": "Sapphire",
  "carat": 2.15,
  "length": 8.2,
  "width": 6.5,
  "height": 4.1,
  "shape": "Oval",
  "color": "Blue",
  "clarity": "VS",
  "origin": "Thailand",
  "treatment": "Heated",
  "description": "Natural blue sapphire"
}
```

The backend:

1. Gets `seller_id` from `VerifiedSellerMiddleware`.
2. Parses JSON.
3. Cleans text values.
4. Requires:

```text
gem_name
gem_type
```

5. Checks that optional numeric values are greater than zero when supplied.
6. Inserts into `public.gems`.
7. Returns the created gem.

Current gem fields:

```text
gem_id
seller_id
gem_name
gem_type
carat
length
width
height
shape
color
clarity
origin
treatment
description
created_at
updated_at
```

---

# 18. Gem Table

Current `gems` table is designed to store:

```text
gem_id
seller_id
gem_name
gem_type
carat
length
width
height
shape
color
clarity
origin
treatment
description
created_at
updated_at
```

A gem belongs to one seller.

A gem may later have:

```text
many images
one or more certificates
multiple listings over time
```

---

# 19. What Has NOT Been Built Yet

The following work is still pending.

## Remaining Gem CRUD

```text
GET /api/gems/{id}
PUT /api/gems/{id}
DELETE /api/gems/{id}
```

Important future ownership rule:

A seller should only be allowed to edit/delete a gem they own.

Expected check:

```text
Authenticated seller_id
        ↓
Gem.seller_id
        ↓
Must match
```

---

# 20. Listings — Not Yet Built

After Gem CRUD, the next stage is Listings.

Planned flow:

```text
Verified Seller
      ↓
Create Gem
      ↓
Create Listing
      ↓
Save Draft
      ↓
Choose Sale Type
      ↓
FIXED_PRICE
```

For now we are only planning to build the fixed-price part.

Relevant planned endpoints:

```text
POST /api/listings
PUT  /api/listings/{id}
POST /api/listings/{id}/publish
POST /api/listings/{id}/relist
```

Current listing concepts:

```text
sale_type:
FIXED_PRICE
AUCTION

status:
DRAFT
PUBLISHED
ACTIVE
SOLD
UNSOLD
EXPIRED
CANCELLED
SUSPENDED
```

For the current scope we will first focus on:

```text
DRAFT
FIXED_PRICE
PUBLISH
```

---

# 21. Auction Engine — Intentionally Deferred

We have explicitly decided to stop before the Auction Engine for now.

Do not continue into these features yet:

```text
Auction creation
Bidding
Bid validation
Minimum increment
Reserve price
Current highest bidder
Highest bid update
Outbid notification
Outbid email
Auction closing
Winner selection
UNSOLD handling
Auction-generated orders
```

These will be implemented later.

Current project stopping target:

```text
Authentication
↓
Seller System
↓
Gem CRUD
↓
Fixed-Price Listings
↓
STOP BEFORE AUCTION ENGINE
```

---

# 22. What the Auction Engine Will Eventually Be

For future reference only:

```text
Seller publishes auction
      ↓
Auction starts
      ↓
User places bid
      ↓
Backend validates
      ↓
Save bid
      ↓
Update highest bidder
      ↓
Notify previous bidder
      ↓
Auction ends
      ↓
Check reserve
      ↓
Winner OR UNSOLD
```

Again: **not part of the current work session**.

---

# 23. How to Run the Backend

Open PowerShell:

```powershell
cd "D:\RAI\adcompro proj\backend"
```

Format changed Go files if needed:

```powershell
gofmt -w *.go
```

Compile:

```powershell
go build .
```

If no error is printed, compilation succeeded.

Run:

```powershell
go run .
```

Expected output:

```text
Connected to Supabase PostgreSQL
GemBid backend running on http://localhost:8080
```

---

# 24. Basic Health Check

With the backend running:

```powershell
Invoke-RestMethod `
    -Method Get `
    -Uri "http://localhost:8080/health"
```

Expected:

```text
status
------
ok
```

---

# 25. Authentication Test Flow

Register:

```powershell
$userBody = @{
    full_name = "Test User"
    email = "test@example.com"
    phone = "0812345678"
    password = "Test1234"
} | ConvertTo-Json

Invoke-RestMethod `
    -Method Post `
    -Uri "http://localhost:8080/api/register" `
    -ContentType "application/json" `
    -Body $userBody
```

Login:

```powershell
$loginBody = @{
    email = "test@example.com"
    password = "Test1234"
} | ConvertTo-Json

$login = Invoke-RestMethod `
    -Method Post `
    -Uri "http://localhost:8080/api/login" `
    -ContentType "application/json" `
    -Body $loginBody

$token = $login.token
```

Protected request:

```powershell
Invoke-RestMethod `
    -Method Get `
    -Uri "http://localhost:8080/api/me" `
    -Headers @{
        Authorization = "Bearer $token"
    }
```

---

# 26. Seller System Test Flow

Normal user logs in:

```text
USER
seller_status = NONE
```

Apply:

```powershell
$sellerBody = @{
    shop_name = "Gem Test Shop"
    seller_type = "Individual"
    legal_name = "Test User"
    address = "Bangkok, Thailand"
    description = "Gemstone seller"
} | ConvertTo-Json

Invoke-RestMethod `
    -Method Post `
    -Uri "http://localhost:8080/api/seller/apply" `
    -Headers @{
        Authorization = "Bearer $token"
    } `
    -ContentType "application/json" `
    -Body $sellerBody
```

Expected database state:

```text
users.seller_status = PENDING
seller_profiles.verification_status = PENDING
```

Status check:

```powershell
Invoke-RestMethod `
    -Method Get `
    -Uri "http://localhost:8080/api/seller/status" `
    -Headers @{
        Authorization = "Bearer $token"
    }
```

---

# 27. Creating a Test Admin

Normal registration never accepts an admin role.

For development/testing only, a test account may be promoted directly in Supabase:

```sql
UPDATE public.users
SET role = 'ADMIN'
WHERE email = 'admin@test.com';
```

Verify:

```sql
SELECT
    user_id,
    email,
    role,
    seller_status
FROM public.users
ORDER BY user_id;
```

Do not expose an API that allows normal users to set their own role.

---

# 28. Admin Seller Approval Test

Admin logs in and saves:

```text
$adminToken
```

List seller applications:

```powershell
Invoke-RestMethod `
    -Method Get `
    -Uri "http://localhost:8080/api/admin/seller-applications" `
    -Headers @{
        Authorization = "Bearer $adminToken"
    }
```

Approve example seller ID `1`:

```powershell
$reviewBody = @{
    decision = "APPROVE"
} | ConvertTo-Json

Invoke-RestMethod `
    -Method Put `
    -Uri "http://localhost:8080/api/admin/seller-applications/1" `
    -Headers @{
        Authorization = "Bearer $adminToken"
    } `
    -ContentType "application/json" `
    -Body $reviewBody
```

Expected:

```text
users.seller_status = VERIFIED
seller_profiles.verification_status = VERIFIED
```

---

# 29. Create Gem Test

Login as the approved seller and save:

```text
$sellerToken
```

Example:

```powershell
$gemBody = @{
    gem_name = "Blue Sapphire"
    gem_type = "Sapphire"
    carat = 2.15
    length = 8.2
    width = 6.5
    height = 4.1
    shape = "Oval"
    color = "Blue"
    clarity = "VS"
    origin = "Thailand"
    treatment = "Heated"
    description = "Natural blue sapphire"
} | ConvertTo-Json

Invoke-RestMethod `
    -Method Post `
    -Uri "http://localhost:8080/api/gems" `
    -Headers @{
        Authorization = "Bearer $sellerToken"
    } `
    -ContentType "application/json" `
    -Body $gemBody
```

Expected:

```text
Gem created successfully
```

Verify with SQL:

```sql
SELECT
    gem_id,
    seller_id,
    gem_name,
    gem_type,
    carat,
    length,
    width,
    height,
    shape,
    color,
    clarity,
    origin,
    treatment,
    description
FROM public.gems
ORDER BY gem_id;
```

---

# 30. Useful Database Verification Queries

Check users:

```sql
SELECT
    user_id,
    full_name,
    email,
    role,
    seller_status,
    created_at
FROM public.users
ORDER BY user_id;
```

Check seller profiles:

```sql
SELECT
    seller_id,
    user_id,
    shop_name,
    verification_status,
    applied_at,
    reviewed_at,
    reviewed_by
FROM public.seller_profiles
ORDER BY seller_id;
```

Check gems:

```sql
SELECT *
FROM public.gems
ORDER BY gem_id;
```

List all public application tables:

```sql
SELECT table_name
FROM information_schema.tables
WHERE table_schema = 'public'
  AND table_type = 'BASE TABLE'
ORDER BY table_name;
```

---

# 31. Security Rules Already Followed

Current backend design follows these rules:

```text
Passwords are hashed with bcrypt.
Plain-text passwords are not stored.

JWT protects private APIs.

Authenticated user_id comes from JWT context.

SQL uses parameters such as:
$1
$2
$3

Normal registration cannot create admins.

Admin APIs check role in the backend.

Seller APIs check VERIFIED status in the backend.

Seller identity comes from authentication/database state,
not from a frontend-provided seller_id.
```

---

# 32. Important Development Notes

## `package main`

All Go files currently in the same backend package should begin with:

```go
package main
```

Files include:

```text
main.go
database.go
auth.go
middleware.go
seller.go
seller_middleware.go
admin.go
gem.go
```

## Common compile check

Before running:

```powershell
go build .
```

If the command returns to PowerShell without errors, compile is successful.

## Formatting

Use:

```powershell
gofmt -w *.go
```

---

# 33. Known Change From Original Proposal

The proposal still contains some sections that say:

```text
Backend = FastAPI + Python
```

The actual implementation has been changed to:

```text
Backend = Go
```

The REST API design, database schema, authentication design, seller flow, listing flow, and future auction behavior remain based on the proposal.

The final design document should eventually be updated so all technology-stack sections consistently say:

```text
Go Backend
```

instead of FastAPI/Python.

---

# 34. Current Progress Estimate

Approximate overall project progress at the current stopping point:

```text
~30–35%
```

Completed or mostly completed:

```text
✅ Requirements / design
✅ Main database schema
✅ Go ↔ Supabase connection
✅ Authentication
✅ JWT middleware
✅ Seller application
✅ Seller status
✅ Admin seller review
✅ Seller approval / rejection
🟡 Gem CRUD started
```

Still pending:

```text
⬜ Remaining Gem CRUD
⬜ Fixed-price listings
⬜ Frontend integration
⬜ Auction Engine
⬜ Notifications
⬜ Orders / checkout
⬜ Shipping
⬜ Free / Pro subscription logic
⬜ Admin analytics
⬜ Full testing
⬜ Docker deployment
⬜ Ubuntu deployment
```

---

# 35. Immediate Next Steps for Teammate

Continue in this exact order:

```text
1. Verify POST /api/gems works
2. Implement GET /api/gems/{id}
3. Implement PUT /api/gems/{id}
4. Implement DELETE /api/gems/{id}
5. Add ownership checks for seller-owned gems
6. Build POST /api/listings
7. Support listing status = DRAFT
8. Build fixed-price listing fields
9. Build publish endpoint
10. Make fixed-price listing visible to marketplace API
11. STOP before Auction Engine
```

---

# 36. Current Scope Boundary

For this phase, the intended completed flow is:

```text
Visitor
↓
Register
↓
Login
↓
Normal User
↓
Apply as Seller
↓
Admin Approves
↓
Verified Seller
↓
Create Gem
↓
Edit / View / Delete Gem
↓
Create Draft Listing
↓
Choose Fixed Price
↓
Publish
↓
Marketplace
```

Do **not** continue into:

```text
Auction
Bids
Reserve Price
Highest Bidder
Outbid Notifications
Auction Close
Winner Selection
```

until the team decides to start the Auction Engine phase.

---

# 37. Short Handoff Summary

```text
Backend language:
Go

Database:
Supabase PostgreSQL

Authentication:
JWT + bcrypt

Current authentication:
Register / Login / Me working design implemented

Seller system:
Application + status + admin approval/rejection implemented

Documents:
Skipped intentionally for mock version

Gem CRUD:
Create Gem started/implemented
Read/Update/Delete still required

Listings:
Not built yet

Auction Engine:
Explicitly postponed

Current stopping goal:
Finish Gem CRUD + Fixed-Price Listing only
```

---

# 38. Final Workflow Snapshot

```text
REGISTER / LOGIN
      ↓
JWT
      ↓
NORMAL USER
      ↓
SELLER APPLICATION
      ↓
PENDING
      ↓
ADMIN REVIEW
      ↓
VERIFIED SELLER
      ↓
CREATE GEM
      ↓
GEM CRUD
      ↓
DRAFT LISTING
      ↓
FIXED PRICE
      ↓
PUBLISH
      ↓
MARKETPLACE

==============================
OLDER SNAPSHOT — SUPERSEDED
See Section 39 for latest status
==============================
```
---

# 39. Latest Update — 27 September 2026

> **This section is the newest project status and supersedes older notes in this handoff that say Gem CRUD, Listings, or the Auction Engine are still pending.**

Today we completed and tested a large part of the backend.

## 39.1 Seller Approval and Authorization Testing

A fresh seller application was tested end-to-end.

Confirmed flow:

```text
Register / Login
      ↓
Apply as Seller
      ↓
PENDING
      ↓
Admin Login
      ↓
Admin Approves Application
      ↓
VERIFIED
```

Confirmed result:

```text
seller_id = 2
user_id = 3
seller status = VERIFIED
```

The admin protection was also tested:

```text
Normal USER calling admin endpoint
→ 403 Forbidden
```

So both seller verification and admin authorization are working.

---

## 39.2 Gem CRUD Status

Gem CRUD is now implemented beyond the earlier snapshot.

Current routes:

```text
GET    /api/gems
POST   /api/gems
GET    /api/gems/{id}
PUT    /api/gems/{id}
DELETE /api/gems/{id}
```

The following were tested successfully:

```text
POST /api/gems       → 201 Created
GET  /api/gems       → 200 OK
GET  /api/gems/{id}  → 200 OK
PUT  /api/gems/{id}  → 200 OK
```

The update test confirmed that the changed values were stored and returned correctly.

The DELETE handler is implemented, but the API DELETE route has not yet been fully integration-tested through the same end-to-end flow.

Important ownership rule is implemented:

```text
Authenticated seller_id
        ↓
Gem seller_id
        ↓
Must match
```

---

## 39.3 Listings Implemented

New file:

```text
backend/listings.go
```

Current listing routes:

```text
POST   /api/listings
GET    /api/listings
GET    /api/listings/{id}
PUT    /api/listings/{id}/publish
DELETE /api/listings/{id}
```

Fixed-price listing flow was tested successfully:

```text
Gem
 ↓
Create Listing
 ↓
DRAFT
 ↓
Publish as FIXED_PRICE
 ↓
PUBLISHED
 ↓
Visible in public marketplace
```

A test listing was published with:

```text
fixed_price = 12500
status = PUBLISHED
```

The public marketplace correctly excluded the listing while it was still DRAFT and included it after publishing.

### Known listing issue found

```text
GET /api/listings/{id}
```

currently allows a public request to retrieve a DRAFT listing by ID.

Recommended fix:

```text
Public user:
PUBLISHED / ACTIVE → return listing
DRAFT              → 404

Owner seller:
should still be able to view own draft through an authenticated seller flow
```

---

## 39.4 Swagger Added

Swagger / Swaggo is now integrated.

Swagger URL:

```text
http://localhost:8080/docs
```

Generated files:

```text
backend/docs/docs.go
backend/docs/swagger.json
backend/docs/swagger.yaml
```

Swagger currently documents:

```text
Authentication
Seller
Admin
Gems
Listings
Auctions
```

Protected endpoints use:

```text
Authorization: Bearer <JWT>
```

Swagger authorization and protected API testing were confirmed working.

---

## 39.5 Docker Added

New files:

```text
backend/Dockerfile
docker-compose.yml
```

Docker Compose validation passed:

```powershell
docker compose config --quiet
```

The backend remains exposed on:

```text
8080
```

---

## 39.6 Auction Engine Started and Core Logic Implemented

The previous decision to stop before the Auction Engine has now been superseded.

New main file:

```text
backend/auction.go
```

Current auction routes:

```text
POST /api/auctions
GET  /api/auctions
GET  /api/auctions/{id}

GET  /api/auctions/{id}/bids
POST /api/auctions/{id}/bids

POST /api/auctions/{id}/close
```

Auction flow:

```text
Gem
 ↓
DRAFT Listing
 ↓
Set sale_type = AUCTION
 ↓
Create Auction
 ↓
SCHEDULED / ACTIVE
 ↓
Users Bid
 ↓
Highest Bidder Updated
 ↓
Auction End Time
 ↓
Close Auction
 ↓
Reserve Reached?
 ├── YES → Winner + SOLD
 └── NO  → UNSOLD
```

Implemented auction rules:

```text
✅ Starting price
✅ Minimum increment
✅ Scheduled / active auction states
✅ Current price
✅ Current highest bidder
✅ Winning bid tracking
✅ Seller cannot bid on own auction
✅ Too-low bids are rejected
✅ Late bids are rejected
✅ Previous highest bid becomes OUTBID
✅ Reserve price check
✅ Winner selection
✅ Winning bid becomes WON
✅ Listing becomes SOLD or UNSOLD
✅ Auction cannot be closed twice
```

---

## 39.7 Concurrency Protection for Bidding

Bid placement uses a PostgreSQL transaction and:

```sql
FOR UPDATE
```

This locks the auction row while a bid is being processed.

Example:

```text
Current = 5000

Bidder A → 5500
Bidder B → 5300
```

The first request updates the current price while the row is locked.

The second request then checks the new price instead of using stale data.

This prevents two simultaneous requests from both being accepted as the highest bid.

---

## 39.8 Automated Auction Integration Test

New file:

```text
backend/auction_integration_test.go
```

This test creates temporary database records and performs a real two-bidder auction flow against Supabase.

Test sequence:

```text
Seller tries to bid 5000
→ 403 Forbidden ✅

Bidder A → 5000
→ ACCEPTED ✅

Bidder B → 5100
→ REJECTED because minimum is 5250 ✅

Bidder B → 5250
→ ACCEPTED ✅
Bidder A old bid → OUTBID ✅

Bidder A → 5500
→ ACCEPTED ✅
Bidder B old bid → OUTBID ✅

Auction reaches end_time

Bidder B → 5750
→ REJECTED because auction ended ✅

Seller closes auction

Auction = ENDED ✅
Winner = Bidder A ✅
Final price = 5500 ✅
Winning bid = WON ✅
Listing = SOLD ✅
```

Actual test result:

```text
PASS auction=1 bidderA=5 bidderB=6 final_price=5500 winner=5 listing=SOLD
--- PASS: TestAuctionEngineTwoBidderFlow
PASS
```

Temporary integration-test records are cleaned up automatically after the test.

Run the full integration test with:

```powershell
cd "D:\RAI\adcompro proj\backend"

$env:RUN_AUCTION_INTEGRATION="1"
go test -run TestAuctionEngineTwoBidderFlow -v
```

Normal test command:

```powershell
go test ./...
```

---

## 39.9 Build and Static Checks

The current code passed:

```text
gofmt              ✅
go build .          ✅
go test ./...       ✅
go vet ./...        ✅
Swagger generation  ✅
Docker config       ✅
```

---

## 39.10 Files Added Today

Main new files:

```text
backend/gem_handlers.go
backend/listings.go
backend/auction.go
backend/auction_integration_test.go
backend/Dockerfile
backend/docs/docs.go
backend/docs/swagger.json
backend/docs/swagger.yaml
docker-compose.yml
```

Existing files modified include:

```text
backend/main.go
backend/auth.go
backend/admin.go
backend/gem.go
backend/middleware.go
backend/seller.go
backend/seller_middleware.go
backend/go.mod
backend/go.sum
```

No project file was intentionally deleted.

---

## 39.11 Git / GitHub Notes

The real `.env` file is confirmed ignored by Git.

```text
.env → ignored
```

Do not commit:

```text
.env
database passwords
JWT secrets
private credentials
```

A short repository `README.md` has also been prepared/added for GitHub.

Before the backend push, line-ending-only migration changes and the generated local `.exe` build change were removed from the pending Git diff so they do not need to be included in the backend commit.

The backend changes still need to be committed/pushed from the local project if that push has not already been run.

---

## 39.12 Current Backend Status

Latest status:

```text
Database                    ✅
Go ↔ Supabase               ✅
Authentication              ✅
JWT Middleware              ✅
Seller Application          ✅
Admin Seller Approval       ✅
Verified Seller Middleware  ✅
Gem Create                  ✅
Gem Read                    ✅
Gem Update                  ✅
Gem Delete Handler          ✅
Draft Listings              ✅
Fixed-Price Listings        ✅
Marketplace Browse          ✅
Swagger                     ✅
Docker Configuration        ✅
Auction Creation            ✅
Scheduled Auctions          ✅
Active Auctions             ✅
Bid Validation              ✅
Minimum Increment           ✅
Concurrency Protection      ✅
Highest Bidder Tracking     ✅
Outbid Handling             ✅
Reserve Price               ✅
Winner Selection            ✅
SOLD / UNSOLD               ✅
Auction Integration Test    ✅
```

Still pending:

```text
⬜ Automatic auction closing at end_time
⬜ Automatic order creation for auction winner
⬜ Notification system
⬜ Payment / transaction workflow
⬜ Shipping workflow
⬜ Frontend integration
⬜ Fix public access to DRAFT listing detail
⬜ Full API test of DELETE /api/gems/{id}
⬜ Ubuntu deployment
```

---

## 39.13 New Immediate Next Steps

Continue in this order:

```text
1. Implement automatic auction closing
2. Automatically create an order for a successful auction winner
3. Fix public GET /api/listings/{id} so DRAFT listings are private
4. Add notifications / outbid notifications
5. Continue order and transaction workflow
6. Connect Next.js frontend
7. Full end-to-end testing
8. Docker / Ubuntu deployment
```

---

# 40. Latest Workflow Snapshot

```text
REGISTER / LOGIN
      ↓
JWT
      ↓
NORMAL USER
      ↓
SELLER APPLICATION
      ↓
PENDING
      ↓
ADMIN REVIEW
      ↓
VERIFIED SELLER
      ↓
CREATE GEM
      ↓
GEM CRUD
      ↓
DRAFT LISTING
      ↓
      ├──────────────────────┐
      ↓                      ↓
FIXED PRICE               AUCTION
      ↓                      ↓
PUBLISH                  CREATE AUCTION
      ↓                      ↓
MARKETPLACE          SCHEDULED / ACTIVE
                             ↓
                           BIDS
                             ↓
                     HIGHEST BIDDER
                             ↓
                        AUCTION END
                             ↓
                       RESERVE CHECK
                         ↓        ↓
                       SOLD     UNSOLD

NEXT:
Automatic Auction Closing
→ Winner Order Creation
→ Notifications
→ Frontend
```

