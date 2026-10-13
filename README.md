# GemBid

GemBid is a gemstone marketplace and auction platform developed for the **Advanced Computer Programming** course.

Users can register, apply to become verified sellers, create gemstone listings, sell gems at fixed prices, and participate in auctions.

## Features

- User registration and JWT login
- Seller application and admin approval
- Gemstone CRUD
- Fixed-price listings
- Auction creation and bidding
- Minimum bid increment validation
- Highest bidder tracking
- Reserve price and winner selection
- Swagger API documentation
- Docker support

## Tech Stack

- Go
- PostgreSQL / Supabase
- JWT
- Swagger
- Docker
- Next.js frontend planned

## Run the Backend

```bash
cd backend
go run .
```

Backend:

```text
http://localhost:8080
```

Swagger API:

```text
http://localhost:8080/docs
```

## Environment Setup

Create a `.env` file in the project root:

```env
DATABASE_URL=YOUR_DATABASE_URL
PORT=8080
JWT_SECRET=YOUR_JWT_SECRET
```

Do not commit the real `.env` file to GitHub.

## Current Status

Completed:

- Authentication
- Seller verification
- Gem CRUD
- Fixed-price listings
- Auction bidding engine

Next:

- Automatic auction closing
- Winner order creation
- Notifications
- log in UI
- register UI
- marketplace UI
- auction function UI
- API connection
- Frontend

## Project

**Gem Marketplace & Auction Platform**  
Advanced Computer Programming  
Academic Year 2026/1
**# GemBid

GemBid is a gemstone marketplace and auction platform developed for the **Advanced Computer Programming** course.

Users can register with a username, apply to become verified sellers, create gemstone listings, sell gems at fixed prices, participate in auctions, place orders, make mock payments, and manage shipping.

## Features

- Username registration and JWT login
- bcrypt password hashing
- User profile and password management
- Seller application and admin approval
- Verified seller authorization
- Gemstone CRUD
- Draft and fixed-price listings
- Buy Now and order management
- Mock payment and transaction flow
- Shipping and delivery confirmation
- Auction creation and bidding
- Minimum bid increment validation
- Highest bidder and outbid tracking
- Reserve price and winner selection
- Automatic auction closing
- Automatic winner order creation
- In-app notifications and email queue
- Free / PRO subscription system
- Admin analytics
- Report and moderation system
- Swagger API documentation
- Docker support
- Integration, E2E, concurrency, and security tests

## Tech Stack

- Go
- PostgreSQL / Supabase
- pgx
- JWT
- bcrypt
- Swagger / Swaggo
- Docker
- Next.js / React frontend

## Run the Backend

```bash
cd backend
go run ./cmd/server
```

Backend:

```text
http://localhost:8080
```

Swagger API:

```text
http://localhost:8080/docs
```

## Environment Setup

Create a `.env` file in the project root:

```env
DATABASE_URL=YOUR_DATABASE_URL
PORT=8080
JWT_SECRET=YOUR_JWT_SECRET
SMTP_HOST=
SMTP_PORT=587
SMTP_FROM=
SMTP_USER=
SMTP_PASS=
```

Use a JWT secret with at least 32 bytes.

SMTP is optional. If SMTP is not configured, in-app notifications still work and email deliveries remain queued.

Do not commit the real `.env` file to GitHub.

## Authentication

Registration:

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

Login:

```json
{
  "username": "testuser",
  "password": "Test1234"
}
```

## Payment

The current project uses a mock payment system for testing.

Supported mock methods:

```text
MOCK_CARD
MOCK_QR
```

A valid mock payment is recorded as successful and changes the order from:

```text
PENDING_PAYMENT
↓
PAID
```

No real money or payment gateway is used.

## Testing

Normal tests:

```bash
cd backend
go test ./...
```

Full Supabase integration tests in PowerShell:

```powershell
cd "D:\RAI\adcompro proj\backend"
$env:RUN_AUCTION_INTEGRATION="1"
go test ./test -count=1 -v
```

The current test suite covers:

- Auction bidding
- Simultaneous bidding
- Automatic auction closing
- Orders and payments
- Shipping
- Subscription concurrency
- Reports and moderation
- Full API end-to-end flow
- Migration validation
- Security and input validation

## Current Status

Backend feature scope:

```text
Authentication                 ✅
Seller System                  ✅
Gem CRUD                       ✅
Listings                       ✅
Fixed Price                    ✅
Auction Engine                 ✅
Automatic Auction Closing      ✅
Buy Now / Orders               ✅
Mock Payments                  ✅
Shipping                       ✅
Notifications                  ✅
Subscriptions                  ✅
Admin Analytics                ✅
Reports / Moderation           ✅
Swagger                        ✅
Security Validation            ✅
Integration Testing            ✅
```

The backend is feature-complete for the current project scope.

Current next phase:

```text
Frontend Integration
↓
Server Deployment
↓
Real Multi-User Testing
↓
Final Demo
```

## Project

**Gem Marketplace & Auction Platform**  
Advanced Computer Programming  
Academic Year 2026/1
