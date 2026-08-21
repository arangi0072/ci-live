# CI Live Backend

A robust backend service for the Live Streaming Platform Prototype, written in Go using the Gin web framework and PostgreSQL. This repository currently provides a comprehensive Authentication Service.

## Features

- **User Authentication**: Secure signup, login, and logout.
- **JWT Management**: Access tokens for short-lived authorization and refresh tokens for secure sessions.
- **Password Security**: Passwords hashed securely using `bcrypt`.
- **Account Recovery**: Forgot password and reset password functionality.
- **Email Verification**: Token-based email verification flow.
- **Automatic Migrations**: Database tables are automatically created/migrated upon application startup using Go's `//go:embed` feature.

## Requirements

- Go 1.20 or higher
- PostgreSQL database

## Getting Started

### 1. Environment Configuration
The application is configured using environment variables. Create a `.env` file in the root directory (or use the one provided) with the following variables:

```env
PORT=8080
ENVIRONMENT=development
DATABASE_URL=postgres://user:password@localhost:5432/dbname?sslmode=disable
JWT_SECRET=your_32_character_long_secret_key
ACCESS_TOKEN_EXPIRY=15m
REFRESH_TOKEN_EXPIRY=720h
```

*(Note: `DATABASE_URL` and `JWT_SECRET` are required. `JWT_SECRET` must be at least 32 characters long).*

### 2. Database
Ensure your PostgreSQL server is running and matches the credentials supplied in your `DATABASE_URL`. The application will automatically run the required database migrations (found in `migrations/live_streaming_prototype_migration.sql`) when the server starts.

### 3. Running the Server
To start the backend service, run:
```bash
go run ./cmd/server/main.go
```
Or to build and run the binary:
```bash
go build -o server ./cmd/server
./server
```

### 4. Running Tests
Unit tests are included for the core logic (password hashing and JWT management). To run the tests:
```bash
go test -v ./internal/auth/...
```

## API Endpoints

All endpoints are prefixed with `/api/v1`.

### Public Routes
- `POST /auth/signup`: Register a new user.
- `POST /auth/login`: Authenticate a user and receive JWT tokens.
- `POST /auth/refresh`: Refresh an expired access token using a refresh token.
- `POST /auth/verify-email`: Verify a user's email address.
- `POST /auth/forgot-password`: Request a password reset link.
- `POST /auth/reset-password`: Reset the password using the token.

### Protected Routes (Requires Bearer Token)
- `GET /auth/me`: Get the currently authenticated user's profile.
- `POST /auth/logout`: Revoke the user's refresh token and log out.

## Project Structure

- `cmd/server/`: The entry point of the application (`main.go`).
- `internal/auth/`: Core authentication logic, handlers, JWT management, and tests.
- `internal/config/`: Configuration loader.
- `internal/database/`: PostgreSQL connection setup and pooling.
- `internal/middleware/`: Gin middlewares (e.g., Auth, Logger, Recovery).
- `internal/router/`: Application routing and endpoint registration.
- `migrations/`: Embedded SQL scripts for database schemas.
