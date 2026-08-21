# CI Live User Service

A robust microservice for the Live Streaming Platform Prototype that handles user profiles and device management. Written in Go using the Gin web framework and PostgreSQL.

## Features

- **User Profiles**: Create, read, and update user profiles (display name, bio, avatar).
- **Public Profiles**: Fetch public user data safely.
- **Username Availability**: Check if a username is already taken.
- **Device Management**: Register devices (like iOS, Android, web) for push notifications (FCM tokens), list devices, update app versions, and deregister devices.
- **JWT Authentication**: Validates tokens issued by the Auth Service to secure private routes.
- **Automatic Migrations**: Database schemas are applied on startup using Go's `//go:embed`.

## Requirements

- Go 1.20 or higher
- PostgreSQL database

## Getting Started

### 1. Environment Configuration
The service is configured using environment variables. Check the `.env` file or provide the following variables:

```env
PORT=8081
ENVIRONMENT=development
DATABASE_URL=postgres://user:password@localhost:5432/dbname?sslmode=disable
JWT_SECRET=your_32_character_long_secret_key
```

*(Note: Ensure the `JWT_SECRET` matches the one used by the Auth Service to correctly validate tokens).*

### 2. Database
Ensure your PostgreSQL server is running. The application will automatically run the required database migrations on startup.

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

## API Endpoints Overview

All endpoints are prefixed with `/api/v1`.

### Public Routes
- `GET /users/:user_id`: Get public profile information.
- `GET /users/username/:username/availability`: Check if a username is available.

### Protected Routes (Requires Bearer Token)
- `GET /users/me`: Get current user's profile.
- `POST /users/me`: Create profile for current user.
- `PATCH /users/me`: Update profile for current user.
- `GET /devices`: List user's registered devices.
- `POST /devices`: Register a new device.
- `PATCH /devices/:device_id`: Update device metadata.
- `DELETE /devices/:device_id`: Deregister a device.

See `API_REFERENCE.md` for detailed payload and response documentation.
