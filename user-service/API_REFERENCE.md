# API Reference (User Service)

This document details all available API endpoints, their expected JSON request parameters, and the resulting JSON responses for the User Service.

---

## 1. Public Endpoints (No Authentication Required)

### 1.1 Get Public Profile
Retrieves the public profile information for a specific user.

- **Endpoint**: `GET /api/v1/users/:user_id`
- **Request Body**: None
- **Success Response (200 OK)**: Returns a [ProfileResponse](#profileresponse-object).

### 1.2 Check Username Availability
Checks if a chosen username is available to use.

- **Endpoint**: `GET /api/v1/users/username/:username/availability`
- **Request Body**: None
- **Success Response (200 OK)**:
  ```json
  {
    "username": "john_doe",
    "available": true
  }
  ```

---

## 2. Protected Endpoints (Requires Access Token)

**Important**: For all protected routes, you must pass the Access Token in the `Authorization` header:
`Authorization: Bearer <your_access_token>`

### 2.1 Get Current User Profile
Retrieves the profile of the currently authenticated user.

- **Endpoint**: `GET /api/v1/users/me`
- **Request Body**: None
- **Success Response (200 OK)**: Returns a [ProfileResponse](#profileresponse-object).

### 2.2 Create User Profile
Creates the profile for the currently authenticated user.

- **Endpoint**: `POST /api/v1/users/me`
- **Request Body**:
  ```json
  {
    "username": "john_doe",     // string, required, 3-50 chars
    "display_name": "John Doe", // string, required, 1-100 chars
    "bio": "Hello world!",      // string, optional, max 500 chars
    "avatar_url": "https://..." // string, optional, max 2048 chars
  }
  ```
- **Success Response (201 Created)**: Returns a [ProfileResponse](#profileresponse-object).

### 2.3 Update User Profile
Updates the profile for the currently authenticated user.

- **Endpoint**: `PATCH /api/v1/users/me`
- **Request Body** (All fields optional):
  ```json
  {
    "username": "new_john_doe",
    "display_name": "John",
    "bio": "Updated bio",
    "avatar_url": "https://..."
  }
  ```
- **Success Response (200 OK)**: Returns a [ProfileResponse](#profileresponse-object).

---

## 3. Device Management (Requires Access Token)

### 3.1 List Devices
Lists all devices registered to the currently authenticated user.

- **Endpoint**: `GET /api/v1/devices`
- **Request Body**: None
- **Success Response (200 OK)**:
  ```json
  {
    "devices": [ /* Array of DeviceResponse objects */ ],
    "total": 1
  }
  ```

### 3.2 Register Device
Registers a new device (or updates an existing one) for push notifications.

- **Endpoint**: `POST /api/v1/devices`
- **Request Body**:
  ```json
  {
    "device_id": "unique-hardware-id", // string, required
    "platform": "ios",                 // string, required (android|ios|web)
    "fcm_token": "firebase-token",     // string, required
    "app_version": "1.0.0"             // string, optional
  }
  ```
- **Success Response (201 Created)**: Returns a [DeviceResponse](#deviceresponse-object).

### 3.3 Update Device
Updates metadata for an existing device.

- **Endpoint**: `PATCH /api/v1/devices/:device_id`
- **Request Body** (All fields optional):
  ```json
  {
    "fcm_token": "new-firebase-token",
    "app_version": "1.0.1",
    "is_active": true
  }
  ```
- **Success Response (200 OK)**: Returns a [DeviceResponse](#deviceresponse-object).

### 3.4 Deregister Device
Removes a device so it no longer receives push notifications.

- **Endpoint**: `DELETE /api/v1/devices/:device_id`
- **Request Body**: None
- **Success Response (200 OK)**:
  ```json
  {
    "message": "device deleted successfully"
  }
  ```

---

## 4. Shared Response Objects

### ProfileResponse Object
```json
{
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "username": "john_doe",
  "display_name": "John Doe",
  "bio": "Hello world!",
  "avatar_url": "https://...",
  "created_at": "2026-08-20T10:00:00Z"
}
```

### DeviceResponse Object
Note: The `fcm_token` is explicitly excluded from responses for security.
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "device_id": "unique-hardware-id",
  "platform": "ios",
  "app_version": "1.0.0",
  "is_active": true,
  "last_seen_at": "2026-08-20T10:00:00Z",
  "created_at": "2026-08-20T10:00:00Z",
  "updated_at": "2026-08-20T10:00:00Z"
}
```
