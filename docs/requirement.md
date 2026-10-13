# Design Document 2026 — Detailed Guide
This document explains what each section of the **Template of Design Document_2026** is asking for and what should be included in your project documentation.
---
# 1. Project Overview
This section gives the reader a quick understanding of the whole project.
## Project Name
Write the official name of your web application.
Examples:
- **Item Allocator for Student Projects**
- **SafeDeal**
Try to use the same project name everywhere:
- Design document
- GitHub repository
- README
- User interface
- Presentation
## Team Members
List everyone working on the project.
Recommended information:
- Full name
- Student ID
- Main responsibility
Example:
| Name | Student ID | Main Responsibility |
|---|---|---|
| Student A | 660xxxxx | Frontend |
| Student B | 660xxxxx | Backend |
| Student C | 660xxxxx | Database & Testing |
This is important because Section 11 also asks the team to explain how work is divided fairly.
## Objective
The objective answers:
> Why are we building this application?
It should explain:
- The problem
- What the application does
- Who it helps
- What improvement it provides
Example:
> The objective of this project is to develop a web-based resource allocation system that allows students to request tools and components for their projects. The system helps laboratory administrators manage requests, check component availability, prioritize requests based on pickup date, and notify students when their requested items are ready.
Do not explain too many technical details here. Keep the focus on the purpose of the project.
## Scope
The scope explains:
> What exactly are we building?
### Included
Possible examples:
- User registration
- Login
- User dashboard
- Create requests
- View requests
- Edit requests
- Delete requests
- Admin dashboard
- Inventory checking
- Notifications
- Subscription system
- Payment simulation
- Analytics
### Excluded
It is useful to explain what is not included.
Examples:
> The system will not process real bank payments. Payment will be simulated using Stripe Test Mode.
> The system will not automatically purchase unavailable laboratory components.
Defining exclusions prevents the project from becoming too broad.
## Target User Persona
This section should be more specific than simply saying:
> Our users are students.
Explain:
### Who they are
Examples:
- University students
- Engineering students
- Laboratory staff
- Small businesses
- First-jobbers
### Their pain points
Example:
> Students currently need to manually contact laboratory staff to check whether components are available.
### What they need
Example:
> They need a simple system where they can request items and know when those items are ready.
### How the app solves the problem
Example:
> The application allows students to submit requests online and automatically checks inventory availability.
Example persona:
> **Persona:** Engineering students aged 18–24 who regularly borrow laboratory components for academic projects. They often do not know whether required items are available before visiting the laboratory. The application allows them to submit requests in advance and receive confirmation when their items are ready.
---
# 2. Requirements
This section defines what the system must do and how well it should operate.
---
## 2.1 Functional Requirements
Functional requirements describe actual features.
A useful writing style is:
> The user can...
or
> The system shall...
### Authentication
Examples:
- Users can register.
- Users can log in.
- Users can log out.
- Users can reset their password.
### User Functions
Examples:
- Users can create a request.
- Users can view their requests.
- Users can edit requests.
- Users can cancel requests.
### Admin Functions
Examples:
- Admins can view all requests.
- Admins can approve or reject requests.
- Admins can check available inventory.
- Admins can update inventory.
### Subscription Functions
Examples:
- Users can view available subscription plans.
- Users can upgrade from Free to Pro.
- Pro users can access premium features.
### Analytics Functions
Examples:
- Admins can view total users.
- Admins can see active subscriptions.
- Admins can see transaction volume.
A good format is:
| ID | Requirement |
|---|---|
| FR-01 | Users can register using email and password. |
| FR-02 | Users can log in to the application. |
| FR-03 | Users can create a project request. |
| FR-04 | Admins can approve requests. |
| FR-05 | Admins can view analytics. |
Using IDs such as **FR-01** also makes testing easier later.
---
## 2.2 Non-functional Requirements
Non-functional requirements describe how well the system should operate.
For example, instead of saying:
> Users can log in.
A non-functional requirement could say:
> Login credentials must be protected securely.
### Security
Possible requirements:
- Passwords must be hashed.
- Authentication should use JWT or secure sessions.
- Sensitive credentials must not be stored directly in source code.
- HTTPS should be used in production.
### Performance
Example:
> Normal API requests should respond within approximately 2 seconds under typical usage.
### Usability
Example:
> Navigation should be simple enough for first-time users to understand without training.
### Responsive Design
The application should work on:
- Desktop
- Tablet
- Mobile
### Reliability
Example:
> Failed requests should display understandable error messages instead of crashing the application.
### Accessibility
Possible requirements:
- Clear text contrast
- Keyboard navigation
- Proper form labels
- Readable button sizes
---
# 3. UI/UX Design
This section shows what the application looks like and how users move through it.
---
## 3.1 Wireframes
The template specifically recommends using **Figma**.
At minimum, the design should include the following major pages.
### Page 1 — Public Landing Page
This is what visitors see before logging in.
Possible content:
- Logo
- Navigation
- Main headline
- Short description
- Benefits
- Call-to-action buttons
- Feature highlights
- Pricing
- FAQ
Example button labels:
- Get Started
- Login
### Page 2 — Authentication Page
Possible login fields:
- Email
- Password
- Login button
- Register link
Possible registration fields:
- Name
- Email
- Password
- Confirm password
- Register button
The login and registration screens may be separate or combined.
### Page 3 — Main Application Workspace
This is usually the most important page because it contains the main CRUD functions.
CRUD means:
- **Create**
- **Read**
- **Update**
- **Delete**
Example:
- Create a request
- View requests
- Edit a request
- Delete or cancel a request
### Page 4 — Pricing & Checkout
The template expects some kind of monetization.
Example:
| Feature | Free | Pro |
|---|---:|---:|
| Projects | 3 | Unlimited |
| Basic Dashboard | Yes | Yes |
| Advanced Analytics | No | Yes |
| Priority Support | No | Yes |
| Price | ฿0 | ฿99/month |
The payment does not need to use real money.
Possible methods:
- Stripe Test Mode
- Mock QR payment
- Simulated checkout
### Page 5 — User Profile & Account Settings
Possible content:
- Name
- Email
- Password settings
- Subscription plan
- Account role
- Profile picture
- Notification preferences
Example:
```text
Name: John Smith
Email: john@email.com
Plan: PRO
Role: USER
[Change Password]
[Manage Subscription]
```
### Page 6 — Admin Analytics Dashboard
This page should show business information, not only CRUD controls.
Possible cards:
```text
Total Users: 1,250
Pro Users: 220
Free Users: 1,030
Monthly Revenue: ฿21,780
Conversion Rate: 17.6%
```
Possible visualizations:
- User growth
- Monthly revenue
- Free vs Pro users
- Number of transactions
---
## 3.2 User Flow Diagram
This diagram shows the path users follow through the system.
Example user flow:
```text
START
  ↓
Landing Page
  ↓
Register / Login
  ↓
Dashboard
  ↓
Create Request
  ↓
Submit
  ↓
Database
  ↓
Admin Review
  ↓
Approved
  ↓
Notification
  ↓
User Receives Item
  ↓
END
```
### Normal User Flow
```text
Landing Page
  ↓
Register
  ↓
Login
  ↓
Dashboard
  ↓
Create Request
  ↓
View Status
```
### Admin Flow
```text
Login
  ↓
Admin Dashboard
  ↓
View Requests
  ↓
Check Inventory
  ↓
Approve / Reject
  ↓
Notify User
```
---
# 4. System Design
This section explains how the application works internally.
---
## 4.1 System Architecture
A typical architecture using Next.js, FastAPI, and PostgreSQL could look like this:
```text
               USER
                 │
                 ▼
        ┌─────────────────┐
        │     Next.js     │
        │    Frontend     │
        └─────────────────┘
                 │
               HTTP
                 │
                 ▼
        ┌─────────────────┐
        │     FastAPI     │
        │     Backend     │
        └─────────────────┘
                 │
                 ▼
        ┌─────────────────┐
        │   PostgreSQL    │
        │    Database     │
        └─────────────────┘
```
If the project uses an external service:
```text
                    External API
                         ↑
                         │
User → Next.js → FastAPI → PostgreSQL
```
### Frontend Responsibilities
The frontend handles:
- UI
- Forms
- Buttons
- Displaying data
- Sending requests to APIs
### Backend Responsibilities
The backend handles:
- Business logic
- Authentication
- Input validation
- CRUD operations
- Authorization
### Database Responsibilities
The database stores:
- Users
- Projects
- Requests
- Products
- Transactions
- Subscription data
---
## 4.2 Database Design
The design document should include an **ER Diagram**.
ER means **Entity Relationship**.
Example:
```text
USERS
-----
user_id PK
name
email
password_hash
role
subscription_id
        │
        │ 1:M
        ▼
PROJECT_REQUESTS
----------------
request_id PK
user_id FK
project_name
pickup_date
status
        │
        │ M:N
        ▼
REQUEST_ITEMS
-------------
request_id FK
item_id FK
quantity
        │
        ▼
ITEMS
-----
item_id PK
name
type
stock
```
Definitions:
- **PK** = Primary Key
- **FK** = Foreign Key
### Users Table
Stores user account information.
Relationship example:
> One user can create multiple requests.
### Requests Table
Stores project requests submitted by users.
Relationship example:
> Each request belongs to one user.
### Items Table
Stores inventory information.
Possible attributes:
- item_id
- item_name
- item_type
- quantity
- availability
---
## 4.3 Application Architecture
System Architecture explains what components exist.
Application Architecture explains how data moves between them.
Example when a user submits a request:
```text
Request Form
    ↓
Frontend validates input
    ↓
POST /api/requests
    ↓
FastAPI receives JSON
    ↓
Backend validates request
    ↓
Database INSERT
    ↓
Database returns request ID
    ↓
FastAPI sends JSON response
    ↓
Frontend updates screen
```
This is the application data flow.
---
# 5. Security Design
This section explains how accounts and data are protected.
---
## Authentication
Example:
> The application uses JWT authentication.
Login flow:
```text
User enters email/password
        ↓
Backend checks database
        ↓
Password hash verified
        ↓
JWT generated
        ↓
JWT sent to user
        ↓
JWT included in future requests
```
---
## Password Security
Plain-text passwords should never be stored.
Bad example:
```text
password123
```
Instead, use hashing such as bcrypt.
Example:
```text
$2b$12$...
```
A good explanation is:
> User passwords will be hashed using bcrypt before being stored in PostgreSQL. Plain-text passwords will never be stored.
---
## Data Protection
Possible methods:
- HTTPS
- Environment variables
- Secure API communication
- SQL parameterization
- Input validation
- Rate limiting
---
## Role-Based Access Control
RBAC means **Role-Based Access Control**.
### Free User
Can:
- Create requests
- View own requests
- Edit profile
Cannot:
- Access admin tools
- Use Pro-only features
### Pro User
Can:
- Use all Free features
- Access premium functionality
- Use higher limits
### Admin
Can:
- View users
- Manage requests
- Manage inventory
- View analytics
Example table:
| Feature | Free | Pro | Admin |
|---|---:|---:|---:|
| Create request | Yes | Yes | Yes |
| View own requests | Yes | Yes | Yes |
| Advanced analytics | No | Yes | Yes |
| User management | No | No | Yes |
| Inventory management | No | No | Yes |
---
# 6. Monetization & Feature Gating
This section explains how the application makes money and how paid features are controlled.
---
## Business Model
Possible models:
### Subscription
Example:
> ฿99/month
### One-time Purchase
Example:
> ฿499 lifetime Pro upgrade
### Usage-based Pricing
Example:
> First 10 reports are free, then ฿10 per additional report.
Subscription is often the easiest model to explain and demonstrate.
---
## Tier Definitions
Example:
| Feature | Free | Pro |
|---|---:|---:|
| Projects | 3 | Unlimited |
| Basic Dashboard | Yes | Yes |
| Advanced Analytics | No | Yes |
| Export PDF | No | Yes |
| Priority Support | No | Yes |
| Price | ฿0 | ฿99/month |
---
## Payment Flow
Example:
```text
User presses Upgrade
       ↓
Selects Pro
       ↓
Checkout Page
       ↓
Stripe Test Mode / Mock Payment
       ↓
Payment Successful
       ↓
Backend receives result
       ↓
Database updated
tier = "PRO"
       ↓
User gains Pro access
```
---
## Enforcement
Feature access should not only be hidden in the frontend.
Incorrect approach:
```text
if user.plan == "free":
    hide button
```
The backend should also verify permissions.
Correct concept:
```text
User requests Premium Feature
        ↓
Backend checks user's tier
        ↓
IF tier = PRO
    Allow request
ELSE
    Return 403 Forbidden
```
This is true feature gating.
---
# 7. Business Analytics Dashboard
This dashboard is intended for administrators.
Its purpose is to answer business questions such as:
> Is the application growing?
> How many users are paying?
> How much revenue is being generated?
---
## Key Metrics
### Total Users
Example:
```text
COUNT(users)
```
### Active Users
Users who were active during a defined time period, such as the last 30 days.
### Free Users
Example:
```text
WHERE tier = 'free'
```
### Pro Users
Example:
```text
WHERE tier = 'pro'
```
### Conversion Rate
Formula:
[
Conversion Rate =
rac{Pro Users}{Total Users}
	imes 100
]
Example:
- Total users = 1,000
- Pro users = 150
[
rac{150}{1000}	imes100 = 15%
]
### MRR
MRR means **Monthly Recurring Revenue**.
Example:
200 Pro users × ฿99/month
[
MRR = 200 	imes 99
]
[
MRR = ฿19,800
]
### ARR
ARR means **Annual Recurring Revenue**.
[
ARR = MRR 	imes 12
]
Example:
[
19,800 	imes 12 = ฿237,600
]
---
## Visualizations
Example:
| Metric | Visualization |
|---|---|
| Monthly Revenue | Line chart |
| Free vs Pro | Pie or donut chart |
| New Users | Bar chart |
| Daily Transactions | Line chart |
| Requests by Category | Bar chart |
---
## Data Sources
Example:
| Metric | Source |
|---|---|
| Total users | users table |
| Pro users | users.subscription |
| Revenue | payments table |
| Requests | requests table |
| Active users | login_activity table |
---
# 8. External System Integration
This section explains any external service, API, hardware, robot, or third-party platform used by the project.
---
## 8.1 External System Description
Possible external systems include:
- Payment API
- Email API
- AI service
- Robot
- IoT device
- External database
Example:
> The application integrates with Stripe Test Mode to simulate subscription payments.
Another example:
> The system integrates with an existing inventory database used by the laboratory.
Explain:
- What the external system is
- Why it is needed
- What data is sent
- What data is received
---
## 8.2 API Design
List important API endpoints.
Example:
| Method | Endpoint | Purpose |
|---|---|---|
| POST | `/api/register` | Register user |
| POST | `/api/login` | Authenticate user |
| GET | `/api/requests` | Get requests |
| POST | `/api/requests` | Create request |
| PUT | `/api/requests/{id}` | Update request |
| DELETE | `/api/requests/{id}` | Delete request |
### Example JSON Request
```json
{
  "project_name": "Gem Cutting Machine",
  "pickup_date": "2026-09-20"
}
```
### Example JSON Response
```json
{
  "id": 15,
  "status": "pending"
}
```
### API Security
A good explanation is:
> Protected endpoints require a valid JWT token in the Authorization header.
---
# 9. Testing Plan
The testing plan should explain what will be tested and how.
---
## 9.1 Unit Testing
Unit testing checks small individual functions.
Examples:
```text
calculate_total_price()
check_inventory()
validate_password()
```
Example description:
> The inventory checking function will be tested using available, partially available, and unavailable stock cases.
---
## 9.2 Integration Testing
Integration testing checks whether multiple parts of the system work together.
Example:
```text
Frontend
   ↓
API
   ↓
Database
```
Possible integration tests:
- Submit a request from the frontend and verify that it appears in PostgreSQL.
- Login through the frontend, receive a JWT from FastAPI, and verify access to a protected endpoint.
- Complete a simulated payment and verify that the user's plan changes to Pro.
---
## 9.3 User Acceptance Testing
User Acceptance Testing checks whether real users can use the system successfully.
Example tasks:
1. Create an account.
2. Log in.
3. Submit a request.
4. Check request status.
5. Upgrade to Pro.
Observe:
- Errors
- User confusion
- Completion time
- User feedback
Example survey question:
> Was it easy to create a request?
Scale:
- 1 = Very difficult
- 5 = Very easy
---
## Testing Against Requirements
Functional requirement IDs can be linked directly to tests.
Example:
| Requirement | Test | Expected Result |
|---|---|---|
| FR-01 | Register user | Account is created |
| FR-02 | Login | JWT is generated |
| FR-03 | Create request | Request is saved |
| FR-04 | Admin approval | Status becomes Approved |
This proves that the completed system satisfies the original requirements.
---
# 10. Timeline and Milestones
This section explains how development will be organized.
Example:
| Week | Phase | Deliverable |
|---|---|---|
| Week 1 | Requirements | Requirements document |
| Week 2 | UI & Database Design | Figma + ER Diagram |
| Week 3 | Backend | Authentication + APIs |
| Week 4 | Frontend | CRUD interface |
| Week 5 | Subscription & Analytics | Pro features + dashboard |
| Week 6 | Testing | Test report |
| Week 7 | Deployment | Live website |
---
## 10.1 Milestones
Milestones are major checkpoints.
Possible milestones:
- Requirements completed
- Figma completed
- Database completed
- Authentication working
- CRUD completed
- Payment simulation completed
- Admin dashboard completed
- Testing completed
- Deployment completed
These represent major project achievements rather than small development tasks.
---
# 11. Deployment Plan
The application must be deployed online.
A local address such as:
```text
http://localhost:3000
```
does not satisfy the live deployment requirement.
---
## 11.1 Hosting and Environment
### Frontend Hosting
For Next.js, possible platforms include:
- Vercel
- Cloud VM
- Docker-based hosting
Example:
```text
https://your-project.vercel.app
```
### Backend Hosting
FastAPI could be hosted on:
- Render
- Railway
- Cloud VM
Example:
```text
https://your-api.onrender.com
```
### Database Hosting
PostgreSQL could be hosted using:
- Supabase
- Neon
- Railway
- Managed PostgreSQL
- Dockerized PostgreSQL on a VM
Example architecture:
```text
Vercel
Next.js
   ↓
Render
FastAPI
   ↓
Neon
PostgreSQL
```
---
## Environment Variables
Sensitive information should never be committed to GitHub.
Examples:
```text
SECRET_KEY
DATABASE_PASSWORD
API_KEY
```
These should be stored in a `.env` file during local development.
Example:
```text
DATABASE_URL=...
JWT_SECRET=...
STRIPE_SECRET_KEY=...
```
The `.gitignore` file should include:
```text
.env
```
A good explanation is:
> Environment variables containing sensitive credentials are excluded from the Git repository using `.gitignore`. Production secrets are configured directly through the hosting platform.
---
## 11.2 Version Control & Git Workflow
A useful Git branch structure is:
```text
main
 │
 └── development
       │
       ├── feature/login
       ├── feature/dashboard
       ├── feature/payment
       └── feature/admin
```
### Main Branch
Contains stable, production-ready code.
### Development Branch
Contains features currently being integrated and tested.
### Feature Branches
Each feature is developed separately.
Example:
```text
feature/authentication
```
Typical flow:
```text
feature/authentication
        ↓
Pull Request
        ↓
development
        ↓
Testing
        ↓
main
```
---
## Commit Practices
Avoid unclear commits such as:
```text
update
fix
stuff
```
Better examples:
```text
feat: add user registration API
fix: prevent duplicate email registration
feat: add admin analytics dashboard
test: add authentication tests
```
---
## Pull Requests
Before merging, another team member should review the code.
Example:
```text
Student A
develops login
      ↓
pushes feature/login
      ↓
creates Pull Request
      ↓
Student B reviews
      ↓
merge into development
```
---
## Contribution Balance
The project template states that individual grades may be adjusted based on commit history.
Therefore, one team member should not perform almost all development work.
Example division:
| Member | Main Area |
|---|---|
| Member 1 | Frontend + UI |
| Member 2 | Backend + API |
| Member 3 | Database + Analytics |
| Member 4 | Testing + Deployment |
Even with assigned areas, every team member should make meaningful contributions to the repository.
---
# Overall Project Structure
When all sections are combined, the project should roughly demonstrate the following system:
```text
                    WEB APPLICATION
Visitor
   │
   ▼
Landing Page
   │
   ├──────────────┐
   ▼              ▼
Register         Login
   │              │
   └──────┬───────┘
          ▼
      User Dashboard
          │
    ┌─────┴─────┐
    ▼           ▼
 CRUD          Profile
    │
    ▼
Database
          +
Pricing
   ↓
Checkout
   ↓
Free → Pro
          +
Admin Login
   ↓
Admin Dashboard
   ↓
Users / Transactions / Analytics
          +
Security
JWT + bcrypt + RBAC
          +
Testing
          +
Deployment
Frontend + Backend + PostgreSQL
          +
GitHub
main → development → feature branches
```
---
# What the Template Is Really Asking
The whole design document can be understood through these questions:
1. **What are you building?**
2. **Who needs it?**
3. **What can it do?**
4. **What does it look like?**
5. **How is it built?**
6. **How is the data stored?**
7. **How is the system secured?**
8. **How does it make money?**
9. **How does the admin measure business performance?**
10. **How do you test it?**
11. **How will your team build and manage it?**
12. **How will it be deployed online?**
If all of these questions are answered clearly, the design document will cover the important requirements in the **Template of Design Document_2026**.
