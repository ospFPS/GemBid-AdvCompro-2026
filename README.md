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
- Frontend

## Project

**Gem Marketplace & Auction Platform**  
Advanced Computer Programming  
Academic Year 2026/1
