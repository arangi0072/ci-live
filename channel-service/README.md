# CI Live Channel Service

A microservice for the Live Streaming Platform Prototype that handles broadcasting channels, categories, topics, and user followings. Written in Go using the Gin web framework and PostgreSQL.

## Features

- **Channel Management**: Users can create and update their streaming channels, customize their channel avatars and banners.
- **Categories & Topics**: Public browsing of streaming categories (e.g., Gaming, Just Chatting) and specific topics (e.g., specific game names).
- **Follow System**: Users can follow/unfollow channels, and fetch their followed channels or their follower counts.
- **JWT Authentication**: Validates tokens issued by the Auth Service to secure private channel operations.
- **Automatic Migrations**: Database schemas are applied on startup.

## Requirements

- Go 1.20 or higher
- PostgreSQL database

## Getting Started

### 1. Environment Configuration
Create a `.env` file or provide the following variables:

```env
PORT=8082
ENVIRONMENT=development
DATABASE_URL=postgres://user:password@localhost:5432/dbname?sslmode=disable
JWT_SECRET=your_32_character_long_secret_key
```

*(Note: Ensure the `JWT_SECRET` matches the one used by the Auth Service to correctly validate tokens).*

### 2. Database
Ensure your PostgreSQL server is running. The application will automatically run the required database migrations on startup.

### 3. Running the Server
```bash
go run ./cmd/server/main.go
```

## API Endpoints Overview

All endpoints are prefixed with `/api/v1`.

### Public Routes
- `GET /channels/:channel_id`: Get channel by ID.
- `GET /channels/username/:username`: Get channel by username.
- `GET /categories/`: List categories.
- `GET /categories/:category_id`: Get category by ID.
- `GET /topics/`: List topics.
- `GET /topics/:topic_id`: Get topic by ID.

### Protected Routes (Requires Bearer Token)
- `POST /channels`: Create your channel.
- `GET /channels/me`: Get your channel.
- `PATCH /channels/me`: Update your channel.
- `POST /channels/:channel_id/follow`: Follow a channel.
- `DELETE /channels/:channel_id/follow`: Unfollow a channel.
- `GET /channels/:channel_id/follow`: Check if you follow a channel.
- `GET /users/me/following`: List channels you follow.

See `API_REFERENCE.md` for detailed payload and response documentation.
