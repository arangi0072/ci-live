# Project Architecture & API Flow

This document explains the structural design of the CI Live Backend and how an API request flows through the various layers of the application.

## 1. Project Structure

The project follows a standard Go layout, organizing code by domain and responsibility. This layered architecture ensures separation of concerns, making the codebase easier to test, maintain, and scale.

```text
ci-live/
├── cmd/
│   └── server/
│       └── main.go       # Application entry point. Loads config, connects to DB, injects dependencies, and starts the server.
├── internal/
│   ├── auth/             # The Auth domain (all authentication logic)
│   │   ├── handler.go    # HTTP handlers (Controller layer) parsing requests and returning JSON responses.
│   │   ├── jwt.go        # JWT creation and validation logic.
│   │   ├── models.go     # Struct definitions for database models and JSON request/response payloads.
│   │   ├── password.go   # Bcrypt password hashing and comparison utilities.
│   │   ├── repository.go # Database interaction (Data Access layer) executing SQL queries.
│   │   ├── service.go    # Core business logic orchestrating rules, validation, and calling the repository.
│   │   └── *_test.go     # Unit tests for the domain logic.
│   ├── config/
│   │   └── config.go     # Loads and parses environment variables from .env.
│   ├── database/
│   │   └── postgres.go   # PostgreSQL connection pool setup and verification.
│   ├── middleware/
│   │   └── ...           # Gin middlewares (e.g., Auth verification, Request Logging, Panic Recovery).
│   ├── response/
│   │   └── ...           # Standardized JSON response formatting helpers.
│   └── router/
│       └── router.go     # Gin router setup, route grouping, and middleware attachment.
├── migrations/
│   ├── embed.go          # Go embed directive to bundle SQL files into the binary.
│   └── *.sql             # Raw SQL files for database schema creation.
├── .env                  # Environment variables configuration.
├── go.mod                # Go module dependencies.
└── README.md             # High-level project documentation.
```

## 2. Layered Architecture

The application is built using a **3-Tier Architecture** pattern:

1. **Handler Layer (`handler.go`)**: 
   - Acts as the HTTP interface (Controller). 
   - Responsible for extracting data from HTTP requests (JSON binding), calling the appropriate Service method, and formatting the HTTP response (status codes and JSON payload).
   - Does *not* contain business logic or database queries.

2. **Service Layer (`service.go`)**: 
   - Contains the core **Business Logic**.
   - Responsible for enforcing rules (e.g., checking if an email already exists, validating password strength, generating tokens).
   - Depends on the Repository layer to fetch or save data.

3. **Repository Layer (`repository.go`)**: 
   - The Data Access layer.
   - Responsible for all interactions with the PostgreSQL database.
   - Executes SQL queries and maps the results to Go structs (`models.go`).

## 3. How an API Request Works (Flow)

Let's trace a typical API request, for example, the **User Signup (`POST /api/v1/auth/signup`)** flow:

### Step 1: Routing (`router.go`)
1. The client sends an HTTP `POST` request to `/api/v1/auth/signup` with a JSON payload (email and password).
2. The Gin router in `router.go` receives the request and matches it to the `authHandler.Signup` function.

### Step 2: Handler (`auth/handler.go`)
1. The `Signup` function is triggered.
2. It attempts to bind the incoming JSON body to a `SignupRequest` struct. If it fails (e.g., malformed JSON), it immediately returns a `400 Bad Request`.
3. If successful, it calls `h.service.Signup(ctx, req)` passing the parsed request data.

### Step 3: Service (`auth/service.go`)
1. The `Signup` method in the Service layer begins executing business rules:
   - Validates that the email is not empty and password is at least 8 characters.
   - Calls `s.repository.GetUserByEmail` to check if a user with this email already exists.
   - Hashes the password using `HashPassword` from `password.go`.
   - Generates a new UUID for the user.
2. It then creates a `User` struct and calls `s.repository.CreateUser(ctx, user)`.
3. After the user is created, it generates an Email Verification Token and calls `s.repository.CreateEmailVerificationToken`.
4. Finally, it uses the `JWTManager` to generate Access and Refresh tokens, stores the refresh token in the database, and returns an `AuthResponse` back to the Handler.

### Step 4: Repository (`auth/repository.go`)
1. Methods like `CreateUser` receive the Go struct and execute raw SQL `INSERT` statements against the PostgreSQL database.
2. It returns any database errors (like unique constraint violations or connection drops) back up to the Service layer.

### Step 5: Response (`auth/handler.go`)
1. The Handler receives the `AuthResponse` (containing the user data and tokens) from the Service.
2. It serializes the response to JSON and sends it back to the client with an HTTP `201 Created` status code.
3. If any error occurred during the process, the Handler passes the error to `handleAuthError`, which translates internal Go errors (like `ErrEmailAlreadyExists`) into appropriate HTTP status codes (like `409 Conflict`) and user-friendly JSON error messages.

## 4. Database Migrations on Startup

To ensure the database is always in sync with the code:
1. When `main.go` runs, it connects to PostgreSQL.
2. It reads the embedded SQL file from `migrations/embed.go` (which contains `live_streaming_prototype_migration.sql`).
3. It executes the SQL directly against the database (`CREATE TABLE IF NOT EXISTS ...`).
4. This guarantees that all required tables and indexes exist before the application starts accepting HTTP traffic.
