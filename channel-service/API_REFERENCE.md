# API Reference (Channel Service)

This document details all available API endpoints for the Channel Service.

---

## 1. Public Endpoints (No Authentication Required)

### 1.1 Channels
Retrieves public channel profiles.

- **GET Channel by ID**: `GET /api/v1/channels/:channel_id`
- **GET Channel by Username**: `GET /api/v1/channels/username/:username`
- **Success Response (200 OK)**:
  ```json
  {
    "id": "uuid",
    "user_id": "uuid",
    "username": "gamer123",
    "name": "Gamer's Channel",
    "description": "Daily streams!",
    "avatar_url": "https://...",
    "banner_url": "https://...",
    "is_verified": false,
    "follower_count": 1500,
    "total_views": 10000,
    "created_at": "2026-08-20T10:00:00Z",
    "updated_at": "2026-08-20T10:00:00Z"
  }
  ```

### 1.2 Categories
Browse top-level streaming categories (e.g., Gaming, Just Chatting).

- **List Categories**: `GET /api/v1/categories/`
- **GET Category by ID**: `GET /api/v1/categories/:category_id`
- **GET Category by Slug**: `GET /api/v1/categories/slug/:slug`
- **Success Response (200 OK)** (For single item):
  ```json
  {
    "id": "uuid",
    "name": "Gaming",
    "slug": "gaming",
    "description": "Video games",
    "icon_url": "https://...",
    "stream_count": 42,
    "created_at": "2026-08-20T10:00:00Z",
    "updated_at": "2026-08-20T10:00:00Z"
  }
  ```

### 1.3 Topics
Browse specific topics/games within categories.

- **List Topics**: `GET /api/v1/topics/`
- **GET Topic by ID**: `GET /api/v1/topics/:topic_id`
- **GET Topic by Slug**: `GET /api/v1/topics/slug/:slug`
- **Success Response (200 OK)** (For single item):
  ```json
  {
    "id": "uuid",
    "category_id": "uuid",
    "name": "Action Game 5",
    "slug": "action-game-5",
    "description": "High octane action.",
    "stream_count": 10,
    "created_at": "2026-08-20T10:00:00Z",
    "updated_at": "2026-08-20T10:00:00Z"
  }
  ```

---

## 2. Protected Endpoints (Requires Access Token)

**Important**: For all protected routes, you must pass the Access Token in the `Authorization` header: `Authorization: Bearer <token>`

### 2.1 My Channel Management

- **Create Channel**: `POST /api/v1/channels`
  ```json
  {
    "username": "gamer123", // required, 3-50 chars
    "name": "My Stream",    // required, 1-100 chars
    "description": "Bio",   // optional
    "avatar_url": "...",    // optional
    "banner_url": "..."     // optional
  }
  ```
- **Get My Channel**: `GET /api/v1/channels/me`
- **Update My Channel**: `PATCH /api/v1/channels/me`
  ```json
  {
    "name": "New Stream Name",
    "description": "New bio",
    "avatar_url": "...",
    "banner_url": "..."
  }
  ```
- **Success Responses**: Return the full Channel JSON object (see 1.1).

### 2.2 Follow System

- **Follow a Channel**: `POST /api/v1/channels/:channel_id/follow`
- **Unfollow a Channel**: `DELETE /api/v1/channels/:channel_id/follow`
- **Check if Following**: `GET /api/v1/channels/:channel_id/follow`
  Returns:
  ```json
  {
    "following": true
  }
  ```
- **Get Channel Follower Count**: `GET /api/v1/channels/:channel_id/followers/count`
  Returns:
  ```json
  {
    "count": 1500
  }
  ```
- **List Channels I Follow**: `GET /api/v1/users/me/following`
  Returns a list of channels the current user is following.
