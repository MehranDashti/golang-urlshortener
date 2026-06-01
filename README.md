# 🔗 Go URL Shortener

A production-grade URL shortener built with Go, Gin, GORM, and MySQL. Built as a learning project to master Go through real-world application development.

---

## ✨ Features

- **URL Shortening** — Create short links with optional expiry dates
- **Click Tracking** — Track how many times each link is clicked
- **JWT Authentication** — Secure access and refresh token pair
- **Role-based Access** — Separate client and admin route groups
- **Admin Dashboard** — Manage all users and links
- **Link Expiry** — Links automatically return `410 Gone` after expiry
- **Standardized API Responses** — Consistent JSON shape for all endpoints
- **Full Test Suite** — Unit tests + integration tests with a real test DB

---

## 🏗 Project Structure

```
golang-urlshortener/
├── cmd/
│   └── server/
│       └── main.go               # Entry point — wires all dependencies
├── internal/                     # Private packages (compiler enforced)
│   ├── apperror/                 # Custom error types with HTTP codes
│   ├── config/                   # Env loading, Config struct
│   ├── database/                 # GORM connection + AutoMigrate
│   ├── handler/                  # HTTP handlers + request validation
│   │   ├── auth_handler.go
│   │   ├── url_handler.go
│   │   ├── admin_handler.go
│   │   ├── request.go            # Validation helper (bindAndValidate)
│   │   ├── requests.go           # Request structs with validate tags
│   │   └── response.go           # Standard APIResponse shape
│   ├── middleware/               # JWT auth + admin role middleware
│   ├── model/                    # GORM models (User, URL)
│   ├── repository/               # DB queries, one file per model
│   ├── router/                   # Gin router + route groups
│   ├── service/                  # Business logic layer
│   └── util/                     # Short code generator
├── pkg/
│   └── token/                    # JWT Manager (reusable)
├── tests/                        # Integration tests
│   ├── testserver/               # Full test server setup
│   ├── testhelper/               # HTTP request helpers
│   ├── auth_test.go
│   └── url_test.go
├── .env.example
├── .env.testing.example
├── .gitignore
└── go.mod
```

---

## 🚀 Getting Started

### Prerequisites

- Go 1.21+
- MySQL 8+

### Installation

```bash
git clone https://github.com/yourusername/golang-urlshortener.git
cd golang-urlshortener
go mod download
```

### Configuration

```bash
cp .env.example .env
```

Edit `.env` with your values:

```env
DB_USER=root
DB_PASS=yourpassword
DB_HOST=localhost
DB_PORT=3306
DB_NAME=urlshortener

APP_PORT=8080
APP_BASE_URL=http://localhost:8080

JWT_SECRET=your-super-secret-key
ACCESS_TOKEN_DURATION_MINUTES=15
REFRESH_TOKEN_DURATION_DAYS=7
```

### Run

```bash
go run cmd/server/main.go
```

The server starts on `http://localhost:8080`. GORM auto-migrates the database on startup.

---

## 📡 API Reference

All endpoints are prefixed with `/api/v1`.

### Response Shape

Every response follows a consistent structure:

**Success:**
```json
{
  "success": true,
  "code": 200,
  "message": "عملیات با موفقیت انجام شد",
  "data": {}
}
```

**Error:**
```json
{
  "success": false,
  "code": 400,
  "message": "validation failed",
  "error": {
    "Email": "Email is required"
  }
}
```

---

### Auth Routes

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| POST | `/api/v1/auth/signup` | Create a new account | Public |
| POST | `/api/v1/auth/login` | Login and get token pair | Public |
| POST | `/api/v1/auth/refresh` | Refresh access token | Public |

**Signup**
```bash
curl -X POST http://localhost:8080/api/v1/auth/signup \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "123456"}'
```

**Login**
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "123456"}'
```

Response:
```json
{
  "success": true,
  "code": 200,
  "message": "ورود با موفقیت انجام شد",
  "data": {
    "access_token": "eyJhbGci...",
    "refresh_token": "eyJhbGci..."
  }
}
```

**Refresh Token**
```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token": "eyJhbGci..."}'
```

---

### Client Routes (JWT Required)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/client/shorten` | Create a short link |
| GET | `/api/v1/client/links` | List your links |

**Create Short Link**
```bash
curl -X POST http://localhost:8080/api/v1/client/shorten \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <access_token>" \
  -d '{"url": "https://example.com", "expires_at": "2027-01-01T00:00:00Z"}'
```

`expires_at` is optional. Format: RFC3339 (`2006-01-02T15:04:05Z`).

Response:
```json
{
  "success": true,
  "code": 201,
  "message": "لینک کوتاه با موفقیت ساخته شد",
  "data": {
    "short_url": "http://localhost:8080/api/v1/abc123",
    "short_code": "abc123",
    "original_url": "https://example.com"
  }
}
```

**List My Links**
```bash
curl http://localhost:8080/api/v1/client/links \
  -H "Authorization: Bearer <access_token>"
```

---

### Public Routes

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/:code` | Redirect to original URL |

Expired links return `410 Gone`.

---

### Admin Routes (JWT + Admin Role Required)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/admin/links` | List all links |
| DELETE | `/api/v1/admin/links/:id` | Delete a link |
| GET | `/api/v1/admin/users` | List all users |

To make a user admin, update directly in MySQL:
```sql
UPDATE users SET role = 'admin' WHERE email = 'admin@example.com';
```

---

## 🔐 Authentication Flow

```
POST /auth/login
    → access token  (15 min) — use on every protected request
    → refresh token (7 days) — use only to get new tokens

Access token expires?
    → POST /auth/refresh with refresh token
    → get new access token + new refresh token (rotation)
    → old refresh token is now invalid
```

Using a refresh token as an access token returns `401`. Using an access token as a refresh token returns `401`. This is enforced by the `token_type` claim embedded in the JWT.

---

## 🧪 Testing

### Unit Tests

```bash
go test ./internal/... -v
```

Handler tests use `httptest` with functional mocks — no database needed.

### Integration Tests

```bash
cp .env.testing.example .env.testing
# fill in your test DB credentials
```

Create the test database:
```sql
CREATE DATABASE urlshortener_test;
```

Run:
```bash
go test ./tests/... -v
```

Integration tests spin up a full server against a real test database. Tables are truncated between tests with `CleanDB()`.

---

## 🏛 Architecture

```
HTTP Request
    ↓
Middleware (JWT Auth → Role Check)
    ↓
Handler  (parse request, validate, write response)
    ↓
Service  (business logic — expiry check, bcrypt, token generation)
    ↓
Repository (GORM queries)
    ↓
MySQL
```

Each layer depends on the one below it through an **interface** — defined in the consumer package. This makes every layer independently testable.

---

## 🛠 Tech Stack

| | |
|---|---|
| Language | Go 1.21+ |
| Web Framework | Gin v1.9.1 |
| ORM | GORM |
| Database | MySQL 8 |
| Auth | golang-jwt/jwt v5 |
| Password Hashing | bcrypt (cost 12) |
| Validation | go-playground/validator v10 |
| Testing | testify (assert + require) |
| Config | godotenv |

---

## 📝 Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_USER` | MySQL username | — |
| `DB_PASS` | MySQL password | — |
| `DB_HOST` | MySQL host | `localhost` |
| `DB_PORT` | MySQL port | `3306` |
| `DB_NAME` | Database name | — |
| `APP_PORT` | Server port | `8080` |
| `APP_BASE_URL` | Base URL for short links | — |
| `JWT_SECRET` | Secret key for JWT signing | — |
| `ACCESS_TOKEN_DURATION_MINUTES` | Access token lifetime | `15` |
| `REFRESH_TOKEN_DURATION_DAYS` | Refresh token lifetime | `7` |

---

## 📄 License

MIT