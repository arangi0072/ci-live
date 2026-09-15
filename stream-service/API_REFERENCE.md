# CI Live Stream Service API

Version: v1
Service: `stream-service`
Base Path: `/api/v1`

---

# 1. Overview

The Stream Service is responsible for the core live-streaming domain of CI Live.

It manages:

- Streams
- Stream sessions
- Stream endpoints
- Stream recordings
- Live chat
- Stream likes
- Stream lifecycle
- Stream-related statistics

The service uses PostgreSQL as its primary persistent data store.

Media ingestion and playback are handled by MediaMTX, while object storage and CDN delivery are handled separately.

---

# 2. Base URL

Development:

    http://localhost:8080/api/v1

Production:

    https://api.cloudignite.in/api/v1

---

# 3. Authentication

Protected endpoints require:

    Authorization: Bearer <access_token>

The authentication middleware extracts the authenticated user's UUID and stores it in the request context as:

    user_id

The Stream Service does not own the authentication user table.

Authentication is handled by the Auth Service.

---

# 4. Standard Response Format

Successful response:

```json
{
  "success": true,
  "data": {}
}
```

Error response:

```json
{
  "success": false,
  "error": {
    "message": "stream not found",
    "code": "NOT_FOUND"
  }
}
```

---

# 5. Health Endpoints

## GET /health

Returns the basic service health status.

### Response

```json
{
  "success": true,
  "data": {
    "status": "ok"
  }
}
```

---

## GET /ready

Returns the service readiness status.

### Response

```json
{
  "success": true,
  "data": {
    "status": "ok"
  }
}
```

---

# 6. Streams

## GET /streams

Returns a list of streams.

### Authentication

Public.

### Query Parameters

| Parameter     | Type    | Description        |
| ------------- | ------- | ------------------ |
| `page`        | integer | Page number        |
| `limit`       | integer | Number of records  |
| `status`      | string  | Stream status      |
| `visibility`  | string  | Stream visibility  |
| `category_id` | UUID    | Filter by category |
| `channel_id`  | UUID    | Filter by channel  |

### Stream Status

```text
CREATED
STARTING
LIVE
RECONNECTING
ENDING
ENDED
FAILED
CANCELLED
```

### Visibility

```text
PUBLIC
UNLISTED
PRIVATE
```

---

## GET /streams/:id

Returns a stream by UUID.

### Authentication

Public.

### Path Parameters

| Parameter | Type |
| --------- | ---- |
| `id`      | UUID |

---

## POST /streams

Creates a new stream.

### Authentication

Required.

### Request

```json
{
  "channel_id": "UUID",
  "title": "My Live Stream",
  "description": "Live gaming session",
  "category_id": "UUID",
  "visibility": "PUBLIC"
}
```

---

## PATCH /streams/:id

Updates a stream.

### Authentication

Required.

### Path Parameters

`id` — stream UUID.

### Request

```json
{
  "title": "Updated Stream Title",
  "description": "Updated description",
  "thumbnail_url": "https://..."
}
```

---

## DELETE /streams/:id

Deletes/cancels a stream.

### Authentication

Required.

---

## PATCH /streams/:id/status

Updates the lifecycle state of a stream.

### Authentication

Required.

### Request

```json
{
  "status": "LIVE"
}
```

---

# 7. Stream Sessions

A stream session represents an individual media-server session associated with a stream.

This allows one stream to have multiple sessions because of reconnects or media-server changes.

---

## POST /streams/:stream_id/sessions

Creates a stream session.

### Authentication

Required.

### Request

```json
{
  "media_server_id": "mediamtx-01",
  "media_path": "stream/abc123"
}
```

---

## GET /sessions/:id

Returns a session.

### Authentication

Required.

---

## GET /streams/:stream_id/sessions

Returns sessions belonging to a stream.

### Authentication

Required.

---

## PATCH /sessions/:id

Updates a stream session.

### Authentication

Required.

### Example

```json
{
  "ended_at": "2026-09-15T12:30:00Z",
  "duration_seconds": 3600,
  "peak_viewers": 1250
}
```

---

## DELETE /sessions/:id

Deletes a stream session.

### Authentication

Required.

---

# 8. Stream Endpoints

Stream endpoints contain the media connection information associated with a stream.

---

## POST /streams/:stream_id/endpoints

Creates an endpoint.

### Authentication

Required.

### Example

```json
{
  "type": "RTMP",
  "url": "rtmp://stream.cloudignite.in/live",
  "is_primary": true
}
```

---

## GET /endpoints/:id

Returns an endpoint.

### Authentication

Required.

---

## GET /streams/:stream_id/endpoints

Lists endpoints for a stream.

### Authentication

Required.

---

## PATCH /endpoints/:id

Updates an endpoint.

### Authentication

Required.

---

## DELETE /endpoints/:id

Deletes an endpoint.

### Authentication

Required.

---

# 9. Recordings

Recordings represent persistent video files generated from stream sessions.

---

## GET /recordings/:id

Returns a recording.

### Authentication

Public.

---

## GET /streams/:stream_id/recordings

Returns recordings belonging to a stream.

### Authentication

Public.

---

## GET /sessions/:session_id/recordings

Returns recordings belonging to a session.

### Authentication

Public.

---

## POST /recordings

Creates a recording record.

### Authentication

Required.

### Example

```json
{
  "stream_id": "UUID",
  "session_id": "UUID",
  "storage_provider": "minio",
  "storage_bucket": "ci-live",
  "storage_key": "recordings/stream-id/file.mp4"
}
```

---

## PATCH /recordings/:id

Updates recording metadata.

### Authentication

Required.

---

## PATCH /recordings/:id/status

Updates recording status.

### Authentication

Required.

---

## DELETE /recordings/:id

Deletes a recording.

### Authentication

Required.

---

# 10. Chat

Chat messages belong to a stream.

Maximum message length:

```
500 characters
```

Messages use soft deletion.

---

## POST /streams/:stream_id/chat/messages

Creates a chat message.

### Authentication

Required.

### Request

```json
{
  "message": "Hello everyone!"
}
```

---

## GET /chat/messages/:id

Returns a chat message.

### Authentication

Required.

---

## GET /streams/:stream_id/chat/messages

Returns stream chat history.

### Authentication

Public.

### Query Parameters

| Parameter         | Type    |
| ----------------- | ------- |
| `page`            | integer |
| `limit`           | integer |
| `user_id`         | UUID    |
| `include_deleted` | boolean |

---

## DELETE /chat/messages/:id

Soft-deletes a chat message.

### Authentication

Required.

Only the message author or authorized stream owner should be permitted to delete the message according to service authorization rules.

---

## GET /streams/:stream_id/chat/stats

Returns chat statistics.

### Authentication

Required.

### Example

```json
{
  "stream_id": "UUID",
  "total_messages": 12500,
  "deleted_messages": 20,
  "active_messages": 12480
}
```

---

# 11. Likes

Likes use a composite key:

```
(stream_id, user_id)
```

A user can therefore like a stream only once.

---

## POST /streams/:stream_id/like

Likes a stream.

### Authentication

Required.

### Response

```json
{
  "success": true,
  "data": {
    "stream_id": "UUID",
    "user_id": "UUID",
    "created_at": "2026-09-15T11:30:00Z"
  }
}
```

---

## DELETE /streams/:stream_id/like

Removes the authenticated user's like.

### Authentication

Required.

---

## GET /streams/:stream_id/like

Returns the authenticated user's like.

### Authentication

Required.

---

## GET /streams/:stream_id/like/check

Checks whether the authenticated user has liked the stream.

### Authentication

Required.

### Example

```json
{
  "success": true,
  "data": {
    "liked": true
  }
}
```

---

## GET /streams/:stream_id/likes

Lists likes for a stream.

### Authentication

Required.

### Query Parameters

| Parameter | Type    |
| --------- | ------- |
| `page`    | integer |
| `limit`   | integer |

---

## GET /streams/:stream_id/likes/stats

Returns stream like statistics.

### Authentication

Required.

### Example

```json
{
  "success": true,
  "data": {
    "stream_id": "UUID",
    "like_count": 1234
  }
}
```

---

# 12. Authorization Model

The service follows an ownership model.

Typical flow:

```text
Authenticated User
        |
        v
Auth Middleware
        |
        v
user_id
        |
        v
Handler
        |
        v
Service
        |
        v
Repository
        |
        v
PostgreSQL
```

Services must verify ownership before allowing protected mutations.

---

# 13. UUID Requirements

The following identifiers use UUIDs:

* Stream ID
* Channel ID
* Session ID
* Endpoint ID
* Recording ID
* User ID
* Category ID

Chat message IDs use:

```
BIGSERIAL
```

and therefore are represented as:

```
int64
```

---

# 14. Pagination

List endpoints support pagination.

Recommended defaults:

```text
page = 1
limit = 20
```

Maximum:

```text
limit = 100
```

Typical response:

```json
{
  "success": true,
  "data": {
    "items": [],
    "total": 0,
    "page": 1,
    "limit": 20
  }
}
```

---

# 15. Error Categories

Common service errors include:

```text
BAD_REQUEST
UNAUTHORIZED
FORBIDDEN
NOT_FOUND
CONFLICT
TOO_MANY_REQUESTS
INTERNAL_SERVER_ERROR
```

Domain-specific errors include:

```text
stream not found
like already exists
stream is not liked
chat message not found
recording not found
session not found
endpoint not found
unauthorized
forbidden
invalid UUID
```

---

# 16. Media Architecture

The Stream Service does not directly transport live video.

Media flow:

```text
Streamer
   |
   | RTMP
   v
MediaMTX
   |
   +---- HLS
   |
   +---- Recording
   |
   v
CDN
   |
   v
Viewer
```

The Stream Service manages the metadata and lifecycle around this media flow.

---

# 17. Storage

MinIO is used for persistent object storage.

Typical objects include:

```text
avatars/
channels/
banners/
thumbnails/
recordings/
```

The database stores metadata and storage references rather than the actual binary media.

---

# 18. Service Responsibilities

The Stream Service owns:

* Stream metadata
* Stream lifecycle
* Stream sessions
* Media endpoint metadata
* Recording metadata
* Chat messages
* Stream likes
* Stream statistics

The Auth Service owns:

* Users
* Passwords
* Refresh tokens
* Email verification
* Password reset

MediaMTX owns:

* RTMP ingest
* Live media processing
* HLS playback
* Media sessions
* Media recording

MinIO owns:

* Object storage

CDN owns:

* Edge delivery
* Cache delivery