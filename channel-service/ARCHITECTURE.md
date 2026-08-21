# Project Architecture & API Flow

This document explains the structural design of the CI Live Channel Service and how requests flow through its layers.

## 1. Project Structure

The project follows a standard Go domain-driven layout:

```text
ci-live/channel-service/
├── cmd/
│   └── server/
│       └── main.go       # Application entry point. Loads config, connects to DB, and starts the server.
├── internal/
│   ├── category/         # Category domain (Handlers, Services, Repositories).
│   ├── channel/          # Channel domain (Handlers, Services, Repositories).
│   ├── config/           # Loads and parses environment variables.
│   ├── database/         # PostgreSQL connection pool.
│   ├── follow/           # Follow system domain (Handlers, Services, Repositories).
│   ├── middleware/       # Gin middlewares (e.g., Auth JWT verification).
│   ├── response/         # JSON response formatters.
│   ├── router/           # Gin router setup and endpoint registration.
│   └── topic/            # Topic domain (Handlers, Services, Repositories).
├── migrations/           # Embedded SQL scripts for database schemas.
├── .env                  # Environment variables.
├── go.mod                # Go module dependencies.
├── README.md             # High-level project documentation.
└── API_REFERENCE.md      # Detailed API documentation.
```

## 2. Layered Architecture

The application uses a **3-Tier Architecture** for every domain (`category`, `channel`, `follow`, `topic`):

1. **Handler Layer**: The Controller. Extracts parameters and JSON bodies from HTTP requests, delegates to the Service, and returns formatted JSON HTTP responses.
2. **Service Layer**: Core Business Logic. Enforces domain rules (e.g. checking if a channel name is valid or if a user is already following a channel). Depends on the Repository.
3. **Repository Layer**: Data Access. Executes raw SQL queries against PostgreSQL.

## 3. Request Flow Example

Example flow for **Follow Channel (`POST /api/v1/channels/:channel_id/follow`)**:

1. **Routing (`router.go`)**: The `POST` request hits `/api/v1/channels/:channel_id/follow`. Since it's protected, it passes through the `Auth` middleware which parses the JWT Bearer token to extract the user's ID.
2. **Handler (`follow/handler.go`)**: The `Follow` handler extracts the `channel_id` from the URL params and the `user_id` from the JWT context. It passes them to `service.Follow()`.
3. **Service (`follow/service.go`)**: Checks if the channel actually exists by calling the channel repository. Then checks if the user is already following the channel. If all checks pass, it creates a `Follow` relationship object.
4. **Repository (`follow/repository.go`)**: Executes a SQL `INSERT` to link the `user_id` and `channel_id` in the database, and increments the channel's `follower_count` in a database transaction.
5. **Response (`follow/handler.go`)**: If successful, the handler replies with a `200 OK` JSON success message.

## 4. Database Migrations

On startup, `main.go` reads the embedded `live_streaming_prototype_migration.sql` script via the `migrations` package and runs it against PostgreSQL to guarantee the tables exist before accepting requests.
