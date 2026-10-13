# Gem Marketplace & Auction Website

## 1. Project Overview

### Project Idea

This project is a **web-based gemstone marketplace and auction platform**.

The system allows normal users to:

- Register and log in.
- Browse gemstones.
- Search and filter gemstones.
- Buy gemstones at a fixed price.
- Join gemstone auctions.
- Place bids.
- View their bid history.
- View their purchase history.
- Receive notifications when they are outbid.
- Apply to become a seller.

A normal user can apply to become a **verified seller** through a second registration process. After the admin approves the seller application, the same account can both **buy and sell gemstones**.

Verified sellers can:

- Create gemstone listings.
- Save listings as drafts.
- Upload gemstone images and certificates.
- Sell gemstones at a fixed price.
- Create timed auctions.
- Set a starting price.
- Set a minimum bid increment.
- Set an optional reserve price.
- View active listings and auctions.
- Relist unsold gemstones.
- Manage orders and shipping information.

Admins can:

- Manage users.
- Review seller applications.
- Approve or reject sellers.
- Review gemstone listings.
- Moderate auctions.
- Review bid history.
- Manage orders and transactions.
- Review reports.
- View business analytics.

---

## 2. Main System Structure

The project is divided into three main parts:

```text
Frontend
   ↓
Backend
   ↓
Database
```

A simple architecture is:

```text
USER
  ↓
Next.js Frontend
  ↓
Backend API
  ↓
PostgreSQL Database
```

The backend can be implemented with **FastAPI** and **GO** for the current project. The project requirement uses FastAPI as an example, but it does not state that FastAPI is the only backend technology allowed.

---

## 3. Frontend

### 3.1 Purpose

The **Frontend** is everything the user sees and interacts with.

It displays different navigation and features depending on the account type.

Recommended frontend tools:

- **Next.js**
- React
- TypeScript / JavaScript
- HTML / CSS

### 3.2 Normal User Navigation

```text
Home
Marketplace
Auctions
My Bids
My Purchases
Notifications
Profile
Become a Seller
```

Normal users can:

- Browse gemstones.
- Search and filter listings.
- View gemstone details.
- Buy fixed-price gemstones.
- Participate in auctions.
- Place bids.
- View bid history.
- Track purchases.
- Receive notifications.
- Apply to become a seller.

### 3.3 Seller Navigation

A verified seller keeps all normal user functions and receives additional seller features.

```text
Home
Marketplace
Auctions
My Bids
My Purchases

My Listings
My Auctions
Orders
Seller Analytics

Notifications
Profile
```

Seller features include:

- Create new gemstone listings.
- Save listing drafts.
- Publish listings.
- Sell at a fixed price.
- Create auctions.
- Manage active listings.
- Manage auctions.
- View orders.
- Add shipping information.
- Relist unsold gemstones.
- View seller analytics.

### 3.4 Admin Navigation

```text
Dashboard

Users
Seller Applications
Gem Listings
Auctions
Bid History
Orders
Transactions

Reports
Analytics

Notifications
Profile
```

Admin features include:

- Review seller applications.
- Approve or reject sellers.
- Manage users.
- Suspend users if necessary.
- Review gemstone listings.
- Suspend or remove listings.
- Review auction activity.
- View full bid history.
- Review reports.
- Monitor orders and transactions.
- View business analytics.

---

## 4. Role-Based Frontend

After login, the system checks the account information.

```text
Login
  ↓
Backend validates account
  ↓
Return user role and seller status
  ↓
Frontend changes navigation
```

Example:

```text
IF role = ADMIN
    Show Admin Dashboard

ELSE
    Show User Marketplace

    IF seller_status = VERIFIED
        Show Seller Features

    ELSE IF seller_status = PENDING
        Show Seller Application Status

    ELSE
        Show "Become a Seller"
```

Possible seller status values:

```text
NOT_REGISTERED
PENDING
VERIFIED
REJECTED
SUSPENDED
```

---

## 5. Marketplace

The marketplace contains both **fixed-price listings** and **auction listings**.

Example fixed-price listing:

```text
Natural Sapphire
2.40 ct

FIXED PRICE
THB 15,000

[Buy Now]
```

Example auction listing:

```text
Natural Ruby
1.80 ct

AUCTION

Current Bid: THB 4,500
18 Bids
Ends in 43 minutes

[Place Bid]
```

Possible filters:

- Gem type
- Carat
- Price
- Origin
- Treatment
- Certificate
- Fixed Price
- Auction
- Ending Soon
- New Auctions

---

## 6. Seller Registration

All users begin as normal users.

```text
Visitor
  ↓
Register
  ↓
Normal User
  ↓
Become a Seller
  ↓
Seller Registration
  ↓
Upload Seller Information
  ↓
PENDING
  ↓
Admin Review
  ↓
VERIFIED / REJECTED
```

Seller registration information can include:

- Seller or shop name
- Legal name
- Phone number
- Address
- Seller type
- Seller description
- Identification document
- Business or shop document

The seller does not need to create a second account.

The same account becomes:

```text
role = USER
seller_status = VERIFIED
```

---

## 7. Gem Listing System

A seller creates a gemstone record first.

Information can include:

- Gem name
- Gem type
- Carat
- Dimensions
- Shape
- Color
- Clarity
- Origin
- Treatment
- Description
- Certificate issuer
- Certificate number
- Certificate file
- Images

The seller can then:

```text
Create Gem
  ↓
Save Draft
  ↓
Review Information
  ↓
Choose Selling Method
```

Selling methods:

```text
Fixed Price
     OR
Auction
```

---

## 8. Fixed-Price Selling

For a fixed-price listing, the seller enters:

- Price
- Quantity

Flow:

```text
Buyer
  ↓
Buy Now
  ↓
Checkout
  ↓
Order Created
  ↓
Payment
  ↓
Seller Prepares Item
  ↓
Shipping
  ↓
Delivered
  ↓
Completed
```

---

## 9. Auction System

For an auction listing, the seller enters:

- Starting price
- Minimum bid increment
- Optional reserve price
- Start time
- End time

Example:

```text
Starting Price: THB 300
Minimum Increment: THB 100
Reserve Price: THB 1,500
```

Auction flow:

```text
Seller Publishes Auction
  ↓
Auction Becomes Active
  ↓
Users Place Bids
  ↓
Backend Validates Each Bid
  ↓
Highest Bidder Updates
  ↓
Previous Highest Bidder Is Notified
  ↓
Auction Timer Ends
  ↓
Winner Is Selected
```

---

## 10. Outbid Notification

Example:

```text
Person A bids THB 300
  ↓
Person A = Highest Bidder

Person B bids THB 400
  ↓
Backend accepts the new bid
  ↓
Person B = New Highest Bidder
  ↓
Person A = Outbid
```

Person A receives:

```text
You have been outbid!

Current Bid: THB 400

Return to the auction if you want to bid again.
```

Notifications should be available through:

- Website notification
- Email notification

The email should only be sent after the new bid is successfully saved.

```text
Validate Bid
  ↓
Save Bid
  ↓
Update Highest Bidder
  ↓
Create Notification
  ↓
Send Email
```

---

## 11. Reserve Price and Auction Closing

The reserve price is optional.

If the reserve is not met:

```text
Auction Ends
  ↓
Reserve Not Met
  ↓
Auction = UNSOLD
  ↓
Notify Seller
  ↓
Allow Relisting
```

If the reserve is met:

```text
Auction Ends
  ↓
Reserve Met
  ↓
Winner Selected
  ↓
Notify Winner
  ↓
Notify Seller
  ↓
Create Order
  ↓
Auction = SOLD
```

---

## 12. Backend

### 12.1 Purpose

The **Backend** controls the logic, rules, security, and communication between the frontend and database.

Recommended current backend tools:

- **FastAPI**
-**GO**
- Python

FastAPI is suitable for the current university project and can manage all required features.
while GO is use 

### 12.2 Backend Responsibilities

#### Authentication

- User registration
- Login
- Logout
- Password hashing
- JWT authentication
- Protected API routes

#### Role and Permission Checking

- Normal user permissions
- Seller permissions
- Admin permissions

#### Seller System

- Seller application
- Seller verification
- Seller approval
- Seller rejection

#### Gem and Listing System

- Create gemstone
- Edit gemstone
- Delete eligible gemstone
- Upload images
- Upload certificates
- Save draft
- Publish listing
- Fixed-price listing
- Auction listing

#### Auction Engine

- Check if auction is active
- Check auction end time
- Validate bid amount
- Check minimum increment
- Save bid
- Update highest bidder
- Keep bid history
- Check reserve price
- Close auction
- Select winner
- Relist unsold auction

#### Notification System

- Create website notifications
- Send outbid email
- Notify auction winner
- Notify seller
- Notify seller approval or rejection
- Notify shipping updates

#### Order System

- Create order
- Update order status
- Track payment
- Track shipping

#### Admin System

- Review seller applications
- Moderate users
- Moderate listings
- Moderate auctions
- Review reports
- Review bid history
- Provide analytics data

---

## 13. Database

### 13.1 Purpose

The **Database** stores and remembers all system data.

Recommended database tool:

- **PostgreSQL**

### 13.2 Main Database Tables

```text
USERS
SELLER_PROFILES
SELLER_DOCUMENTS
GEMS
GEM_IMAGES
CERTIFICATES
LISTINGS
AUCTIONS
BIDS
NOTIFICATIONS
EMAIL_DELIVERIES
ORDERS
TRANSACTIONS
SHIPPING
REPORTS
MODERATION_ACTIONS
```

A future version may also include:

```text
SUBSCRIPTIONS
```

---

## 14. Database Relationship Tree

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
│       │
│       ├── CERTIFICATES
│       │
│       └── LISTINGS
│           │
│           ├── AUCTIONS
│           │   │
│           │   └── BIDS
│           │
│           └── ORDERS
│               │
│               ├── TRANSACTIONS
│               └── SHIPPING
│
├── BIDS
│
├── ORDERS
│
├── NOTIFICATIONS
│   │
│   └── EMAIL_DELIVERIES
│
└── REPORTS
```

Admin actions connect to:

```text
ADMIN USER
│
├── USERS
├── SELLER_PROFILES
├── LISTINGS
├── AUCTIONS
├── REPORTS
└── MODERATION_ACTIONS
```

---

## 15. Important Database Relationships

### User to Seller

```text
USERS
  1
  │
  │
  0..1
SELLER_PROFILES
```

One user can have one seller profile.

### Seller to Gems

```text
SELLER_PROFILE
      1
      │
      │
      M
     GEMS
```

One seller can own many gemstones.

### Gem to Listings

```text
GEM
 1
 │
 │
 M
LISTINGS
```

One gemstone can have multiple listings over time, which supports relisting.

Example:

```text
GEM #101
  │
  ├── LISTING #1 → UNSOLD
  │
  └── LISTING #2 → ACTIVE
```

### Auction to Bids

```text
AUCTION
   1
   │
   │
   M
 BIDS
```

Every bid is saved to create a complete bid history and audit log.

### Listing to Order

Both fixed-price sales and auction wins can create orders.

```text
FIXED PRICE
     ↓
   ORDER
```

or:

```text
AUCTION WINNER
     ↓
   ORDER
```

---

## 16. Main Tools

| Part | Tool | Purpose |
|---|---|---|
| Frontend | Next.js | Build the website interface |
| Frontend | React | Build reusable UI components |
| Frontend | TypeScript / JavaScript | Frontend programming |
| Backend | FastAPI | REST API and business logic |
| Backend | Python | Backend programming |
| Database | PostgreSQL | Store application data |
| Authentication | JWT | User authentication |
| Password Security | bcrypt | Hash passwords |
| Email | Email API / SMTP Service | Send outbid and system emails |
| File Storage | Cloud/Object Storage | Store gem images and documents |
| API Testing | FastAPI Swagger / Postman | Test backend endpoints |
| Database Management | DBeaver | View and manage PostgreSQL |
| UI Design | Figma | Wireframes and interface design |
| Version Control | Git + GitHub | Team development and source control |

---

## 17. Recommended Current Technology Stack

```text
Frontend
Next.js + React + TypeScript

        ↓ REST API

Backend
FastAPI + Python

        ↓

Database
PostgreSQL
```

Additional services:

```text
Email Service
    ↓
Outbid / Auction / Order Notifications

File Storage
    ↓
Gem Images
Certificates
Seller Documents
```

---

## 18. Example of the Complete System

```text
Person B clicks "Bid THB 400"
          ↓
       FRONTEND
       Next.js
          ↓
POST /api/auctions/{id}/bids
          ↓
       BACKEND
       FastAPI
          ↓
Check:
- User logged in?
- Auction active?
- Bid high enough?
- Minimum increment met?
          ↓
       DATABASE
      PostgreSQL
          ↓
Save Person B's THB 400 bid
Update current highest bidder
          ↓
       BACKEND
Create OUTBID notification
          ↓
      EMAIL SERVICE
          ↓
Person A receives:
"You have been outbid!"
          ↓
       FRONTEND
Person B sees:
"You are currently the highest bidder."
```

---

## 19. Project Goal

The final system should provide a complete gemstone marketplace where:

- Users can browse, buy, and bid on gemstones.
- Users can become verified sellers.
- Sellers can choose between fixed-price sales and auctions.
- Auctions are controlled by reliable backend logic.
- Every bid is recorded.
- Outbid users receive notifications.
- Unsold gems can be relisted.
- Orders, payments, and shipping can be tracked.
- Admins can manage and monitor the platform.

The project demonstrates a complete connection between:

```text
Frontend
+
Backend
+
Database
```

creating a functional full-stack web application.


# Group member:
-sunhanat panaskaroen 68011526 ( Leader, Back-end , DevOps )
Siraphop William Wakeling 68011516 ( Front-end )
-Sakolpat Rungrit 68011507 ( Back-end )
-Sirawich Tangsongsuwan 68011519  ( Front-end )