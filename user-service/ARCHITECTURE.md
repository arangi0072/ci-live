# Project Architecture & API Flow

This document explains the structural design of the CI Live User Service and how an API request flows through the various layers of the application.

## 1. Project Structure

The project follows a standard Go layout, organizing code by domain and responsibility.

```text
ci-live/user-service/
├── cmd/
│   └── server/
│       └── main.go       # Application entry point. Loads config, connects to DB, injects dependencies, and starts the server.
├── internal/
│   ├── config/           # Loads and parses environment variables.
│   ├── database/         # PostgreSQL connection pool setup and verification.
│   ├── device/           # Device Management domain logic.
│   │   ├── handler.go    # HTTP handlers (Controller).
│   │   ├── models.go     # Struct definitions for DB and JSON payloads.
│   │   ├── repository.go # Database interaction (Data Access).
│   │   └── service.go    # Core business logic.
│   ├── middleware/       # Gin middlewares (e.g., Auth JWT verification, Logging).
│   ├── response/         # Standardized JSON response formatting helpers.
│   ├── router/           # Gin router setup and route grouping.
│   └── user/             # User Profile domain logic (Handlers, Services, Repositories).
├── migrations/           # Embedded SQL scripts for database schemas.
├── .env                  # Environment variables configuration.
├── go.mod                # Go module dependencies.
├── README.md             # High-level project documentation.
└── API_REFERENCE.md      # Detailed API Request/Response docs.
```

## 2. Layered Architecture

The application is built using a **3-Tier Architecture** pattern for both the `user` and `device` domains:

1. **Handler Layer (`handler.go`)**: 
   - Acts as the HTTP interface (Controller). 
   - Extracts data from HTTP requests, calls the Service method, and formats the HTTP response.
   - Does *not* contain business logic or database queries.

2. **Service Layer (`service.go`)**: 
   - Contains the core **Business Logic**.
   - Responsible for enforcing rules (e.g., checking if a username is taken, verifying device ownership).
   - Depends on the Repository layer to fetch or save data.

3. **Repository Layer (`repository.go`)**: 
   - The Data Access layer.
   - Executes SQL queries against the PostgreSQL database and maps the results to Go structs.

## 3. How an API Request Works (Flow)

Example flow for **Register Device (`POST /api/v1/devices`)**:

1. **Routing (`router.go`)**: The request hits `/api/v1/devices` and passes through the `Auth` middleware, which validates the JWT Bearer token and extracts the `user_id`.
2. **Handler (`device/handler.go`)**: The `Register` handler binds the JSON payload to a `RegisterDeviceRequest` struct. It extracts the `user_id` from the Gin context and passes both to the Service.
3. **Service (`device/service.go`)**: The Service validates the payload (e.g., checking if the platform is valid). It checks if the device is already registered. If it's a new device, it generates a new UUID and builds a `Device` model.
4. **Repository (`device/repository.go`)**: The Repository executes a SQL `INSERT` statement to save the device in PostgreSQL.
5. **Response (`device/handler.go`)**: The Handler receives the newly created device from the Service and returns it as JSON with a `201 Created` status code.
