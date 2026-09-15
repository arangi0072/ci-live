# CI Live Stream Service Architecture

Service: `stream-service`

Version: 1.0

---

# 1. Purpose

`stream-service` is the core backend service responsible for the live-streaming domain of CI Live.

It provides APIs for:

- Stream creation and management
- Stream lifecycle
- Stream sessions
- Media endpoints
- Recordings
- Live chat
- Likes
- Stream statistics

The service is designed around a layered architecture:

```text
HTTP
 |
 v
Router
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

External infrastructure such as MediaMTX, MinIO, Redis and CDN integrates with the service where required.

---

# 2. High-Level Architecture

```text
                         ┌─────────────────────┐
                         │      Clients        │
                         │ Android / Web / iOS │
                         └──────────┬──────────┘
                                    │
                                    │ HTTPS
                                    v
                         ┌─────────────────────┐
                         │ Cloudflare / Gateway│
                         └──────────┬──────────┘
                                    │
                                    v
                    ┌──────────────────────────────┐
                    │        stream-service        │
                    │                              │
                    │  Router                      │
                    │     ↓                        │
                    │  Middleware                  │
                    │     ↓                        │
                    │  Handlers                    │
                    │     ↓                        │
                    │  Services                    │
                    │     ↓                        │
                    │  Repositories                │
                    └───────┬──────────┬───────────┘
                            │          │
                            │          │
                            v          v
                    ┌────────────┐  ┌────────────┐
                    │ PostgreSQL │  │   Redis    │
                    └────────────┘  └────────────┘

                            │
                            │ Media Control / Metadata
                            v
                    ┌────────────────┐
                    │    MediaMTX    │
                    └───────┬────────┘
                            │
                  ┌─────────┴─────────┐
                  │                   │
                 HLS               Recording
                  │                   │
                  v                   v
                CDN                 MinIO
                  │
                  v
               Viewers
```

---

# 3. Project Structure

```text
stream-service/
│
├── cmd/
│   └── server/
│       └── main.go
│
├── internal/
│   │
│   ├── category/
│   │   ├── handler.go
│   │   ├── models.go
│   │   ├── service.go
│   │   └── repository.go
│   │
│   ├── channel/
│   │   ├── handler.go
│   │   ├── models.go
│   │   ├── service.go
│   │   └── repository.go
│   │
│   ├── follow/
│   │   ├── handler.go
│   │   ├── models.go
│   │   ├── service.go
│   │   └── repository.go
│   │
│   ├── topic/
│   │   ├── handler.go
│   │   ├── models.go
│   │   ├── service.go
│   │   └── repository.go
│   │
│   ├── stream/
│   │   ├── handler.go
│   │   ├── models.go
│   │   ├── service.go
│   │   └── repository.go
│   │
│   ├── session/
│   │   ├── handler.go
│   │   ├── models.go
│   │   ├── service.go
│   │   └── repository.go
│   │
│   ├── endpoint/
│   │   ├── handler.go
│   │   ├── models.go
│   │   ├── service.go
│   │   └── repository.go
│   │
│   ├── recording/
│   │   ├── handler.go
│   │   ├── models.go
│   │   ├── service.go
│   │   └── repository.go
│   │
│   ├── chat/
│   │   ├── handler.go
│   │   ├── models.go
│   │   ├── service.go
│   │   └── repository.go
│   │
│   ├── like/
│   │   ├── handler.go
│   │   ├── models.go
│   │   ├── service.go
│   │   └── repository.go
│   │
│   ├── middleware/
│   │   ├── auth.go
│   │   ├── logger.go
│   │   ├── recovery.go
│   │   └── response.go
│   │
│   ├── router/
│   │   └── router.go
│   │
│   ├── database/
│   │   ├── postgres.go
│   │   └── migrate.go
│   │
│   └── config/
│       └── config.go
│
├── migrations/
│   └── 001_init.sql
│
├── .env
├── config.yaml
├── go.mod
├── go.sum
├── API.md
└── ARCHITECTURE.md
```

---

# 4. Layered Architecture

## Handler Layer

Responsible for HTTP concerns.

Responsibilities:

* Read path parameters
* Read query parameters
* Decode JSON
* Validate basic HTTP input
* Retrieve authenticated user ID
* Call service
* Convert service errors to HTTP responses

Handlers must not contain database queries.

---

# 5. Service Layer

The service layer contains business logic.

Responsibilities:

* Business validation
* Authorization
* Ownership checks
* Domain rules
* Pagination rules
* Duplicate checks
* State validation
* Calling repositories

Example:

```text
Create Like
    |
    v
Validate user
    |
    v
Validate stream
    |
    v
Check ownership/business rules
    |
    v
Check existing like
    |
    v
Repository.CreateLike()
```

---

# 6. Repository Layer

Repositories are responsible for PostgreSQL access.

Responsibilities:

* SQL queries
* Inserts
* Updates
* Deletes
* SELECT queries
* Aggregations
* Transactions where required

Repositories should not contain HTTP logic.

---

# 7. Dependency Flow

The application is composed in `cmd/server/main.go`.

```text
Database
   |
   v
Repository
   |
   v
Service
   |
   v
Handler
   |
   v
Router
```

Example:

```text
PostgreSQL
    ↓
stream.Repository
    ↓
stream.Service
    ↓
stream.Handler
    ↓
router
```

---

# 8. Database Architecture

The Stream Service uses PostgreSQL.

Database:

```text
ci_live
```

The database contains the streaming domain.

Important tables include:

```text
user_profiles
channels
channel_follows
categories
topics
streams
stream_sessions
stream_endpoints
stream_recordings
stream_likes
chat_messages
viewer_sessions
user_devices
notifications
```

The `users` table belongs to the Auth Service.

The streaming database references users using UUIDs but does not rely on cross-database foreign keys.

---

# 9. Stream Model

The central domain object is:

```text
Channel
   |
   └── Stream
          |
          ├── Sessions
          |
          ├── Endpoints
          |
          ├── Recordings
          |
          ├── Likes
          |
          └── Chat Messages
```

---

# 10. Stream Lifecycle

A stream follows the database-defined lifecycle:

```text
CREATED
   |
   v
STARTING
   |
   v
LIVE
   |
   +----> RECONNECTING
   |          |
   |          v
   |         LIVE
   |
   v
ENDING
   |
   v
ENDED
```

Failure path:

```text
STARTING ─────> FAILED

LIVE ─────────> FAILED
```

Cancellation path:

```text
CREATED ──────> CANCELLED
```

The service should enforce valid lifecycle transitions according to the stream business rules.

---

# 11. Stream Sessions

A stream can contain multiple sessions.

```text
Stream
 │
 ├── Session 1
 │      └── MediaMTX instance A
 │
 ├── Session 2
 │      └── MediaMTX instance A
 │
 └── Session 3
        └── MediaMTX instance B
```

This allows reconnects and media-server changes without creating a completely new stream entity.

---

# 12. Media Architecture

Media transport is intentionally separated from API metadata.

## Publishing

```text
Streamer
   |
   | RTMP
   v
MediaMTX
   |
   v
Stream Session
   |
   v
stream-service
```

The API service stores stream/session metadata while MediaMTX handles the actual media pipeline.

---

# 13. Playback

For standard HLS playback:

```text
MediaMTX
    |
    v
CDN
    |
    v
Viewer
```

The Stream Service provides the metadata required by the client to discover the stream and playback information.

---

# 14. Recording Architecture

```text
Streamer
    |
    v
MediaMTX
    |
    v
Recording
    |
    v
MinIO
    |
    v
CDN / API
    |
    v
Viewer
```

The database stores recording metadata.

MinIO stores the actual recording object.

---

# 15. Chat Architecture

Persistent chat:

```text
Client
  |
  | POST message
  v
stream-service
  |
  v
Chat Service
  |
  v
PostgreSQL
```

Realtime chat infrastructure can use Redis for:

* Presence
* Pub/Sub
* Realtime state
* Stream chat coordination

The PostgreSQL database remains the persistent source for chat messages.

---

# 16. Likes Architecture

Likes are persisted in:

```text
stream_likes
```

with:

```text
PRIMARY KEY (stream_id, user_id)
```

Therefore:

```text
User A + Stream X = one like
```

Duplicate likes are rejected at both the service/business layer and database constraint level.

---

# 17. Redis

Redis is intended for realtime state and high-frequency operations.

Example keys:

```text
stream:{id}:viewer_count
stream:{id}:presence
stream:{id}:chat
stream:{id}:likes
```

Redis should not replace PostgreSQL for durable stream metadata.

---

# 18. Authentication Architecture

Authentication belongs to the Auth Service.

```text
Client
   |
   v
Auth Service
   |
   v
JWT
   |
   v
stream-service
   |
   v
Auth Middleware
   |
   v
user_id
```

The Stream Service trusts the authenticated identity supplied through the authentication mechanism.

---

# 19. Authorization

Authentication answers:

```text
Who is the user?
```

Authorization answers:

```text
Is this user allowed to perform this operation?
```

Authorization belongs primarily in the service layer.

Examples:

```text
Create Stream
    ↓
Authenticated user

Update Stream
    ↓
Stream owner

Delete Chat Message
    ↓
Message owner / authorized stream owner

Like Stream
    ↓
Authenticated user
```

---

# 20. Configuration

Configuration is loaded through the config package.

Major configuration areas:

```text
App
Server
PostgreSQL
Redis
JWT
MediaMTX
MinIO
CDN
```

Configuration is provided through environment variables and optional configuration files.

---

# 21. Logging

Zap is used for structured logging.

Important events include:

```text
Service startup
Database connection
Migration
HTTP server startup
HTTP errors
Unexpected shutdown
Database errors
```

Logs should avoid sensitive information such as:

* Passwords
* JWT private keys
* Database credentials
* Refresh tokens

---

# 22. Error Handling

Errors flow upward:

```text
Repository
    ↓
Service
    ↓
Handler
    ↓
HTTP Response
```

Repository errors should be converted into meaningful domain/service errors where appropriate.

Handlers should not expose raw PostgreSQL errors to clients.

---

# 23. Graceful Shutdown

The server listens for:

```text
SIGINT
SIGTERM
```

Shutdown sequence:

```text
Signal
  |
  v
Stop accepting new work
  |
  v
HTTP Server Shutdown
  |
  v
Database Pool Close
  |
  v
Process Exit
```

The HTTP server currently uses a bounded graceful shutdown timeout.

---

# 24. Security Principles

The service should follow:

* HTTPS in production
* JWT authentication
* Authorization checks
* Parameterized SQL queries
* UUID validation
* Request validation
* Maximum chat message length
* Structured logging
* No credential logging
* Database constraints
* Soft deletion where required

---

# 25. Scalability

The service is designed to be horizontally scalable.

```text
                 Load Balancer
                      |
          ┌───────────┼───────────┐
          v           v           v
     stream-01   stream-02   stream-03
          |           |           |
          └───────────┼───────────┘
                      |
               PostgreSQL
                      |
                   Redis
```

Application instances should remain stateless wherever possible.

Realtime state can be shared through Redis.

---

# 26. Media Scaling

Media traffic should not be forced through the Go API service.

Preferred architecture:

```text
API traffic
    ↓
Go Stream Service

Media traffic
    ↓
MediaMTX
    ↓
CDN
    ↓
Viewers
```

This prevents large video bandwidth from consuming API server resources.

---

# 27. Failure Isolation

The architecture separates responsibilities:

```text
Auth Failure
    ≠
API Failure
    ≠
Database Failure
    ≠
Media Server Failure
    ≠
Object Storage Failure
```

For example, a MediaMTX failure should not require the entire API service to restart.

---

# 28. Deployment Architecture

Production deployment can be structured as:

```text
                         Internet
                            |
                            v
                       Cloudflare
                            |
              ┌─────────────┴─────────────┐
              │                           │
              v                           v
        API Gateway                    CDN
              |                           |
              v                           v
       stream-service                  HLS
              |
       ┌──────┼────────┐
       │      │        │
       v      v        v
   PostgreSQL Redis  MediaMTX
                       |
                       v
                     MinIO
```

---

# 29. Source of Truth

Persistent domain data:

```text
PostgreSQL
```

Realtime state:

```text
Redis
```

Live media:

```text
MediaMTX
```

Object files:

```text
MinIO
```

Edge delivery:

```text
CDN
```

Authentication:

```text
Auth Service
```

---

# 30. Design Principles

The Stream Service follows these principles:

### Separation of concerns

HTTP, business logic and database access remain separated.

### Database integrity

Important domain constraints are enforced at the database level.

### Stateless API

Application instances should not depend on local memory for persistent state.

### Media/API separation

Video traffic is handled by the media infrastructure instead of the Go API.

### Horizontal scalability

Multiple Stream Service instances can run behind a load balancer.

### Explicit ownership

User ownership is checked before protected mutations.

### Durable persistence

PostgreSQL is used for durable domain state.

### Realtime optimization

Redis is used for high-frequency realtime operations.

---

# 31. Request Lifecycle

Example: creating a chat message.

```text
Client
  |
  | POST /api/v1/streams/{id}/chat/messages
  | Authorization: Bearer JWT
  v
Gin Router
  |
  v
Auth Middleware
  |
  | user_id
  v
Chat Handler
  |
  v
Chat Service
  |
  ├── Validate message
  ├── Validate stream
  └── Apply authorization/business rules
  |
  v
Chat Repository
  |
  v
PostgreSQL
  |
  v
Chat Message
  |
  v
Service
  |
  v
Handler
  |
  v
JSON Response
```

---

# 32. Application Startup

Startup sequence:

```text
main()
  |
  ├── Load configuration
  |
  ├── Initialize logger
  |
  ├── Create application context
  |
  ├── Register shutdown signals
  |
  ├── Connect PostgreSQL
  |
  ├── Run migrations
  |
  ├── Create repositories
  |
  ├── Create services
  |
  ├── Create handlers
  |
  ├── Create router
  |
  └── Start HTTP server
```

---

# 33. Technology Stack

| Component         | Technology |
| ----------------- | ---------- |
| Language          | Go         |
| HTTP Framework    | Gin        |
| Database          | PostgreSQL |
| PostgreSQL Driver | pgx/v5     |
| Connection Pool   | pgxpool    |
| Cache / Realtime  | Redis      |
| Logging           | Zap        |
| Configuration     | Viper      |
| Authentication    | JWT        |
| Password Hashing  | bcrypt     |
| Media Server      | MediaMTX   |
| Object Storage    | MinIO      |
| Delivery          | CDN        |

---

# 34. Architectural Boundary

The Stream Service should remain focused on streaming-domain responsibilities.

It should NOT become responsible for:

* Password management
* User authentication storage
* Email delivery
* Payment processing
* Video transcoding
* CDN implementation
* Object-storage implementation

Those concerns belong to dedicated services or infrastructure components.

---

# 35. Final Architecture

```text
                              CI LIVE
                                 |
             ┌───────────────────┼───────────────────┐
             |                   |                   |
             v                   v                   v
        Auth Service       Stream Service          CDN
             |                   |                   |
             v                   |                   |
         PostgreSQL              |                   |
                                 |
            ┌────────────────────┼────────────────────┐
            |                    |                    |
            v                    v                    v
       PostgreSQL             Redis                MediaMTX
                                                     |
                                                     |
                                           ┌─────────┴─────────┐
                                           |                   |
                                           v                   v
                                         HLS               Recordings
                                           |                   |
                                           v                   v
                                         CDN                 MinIO
                                           |
                                           v
                                        Viewers
```

This architecture keeps the **API/control plane**, **realtime state**, **media plane**, and **storage layer** separated, which is important as CI Live grows from a prototype into a production streaming platform.