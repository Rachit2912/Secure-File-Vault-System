# API Reference

This document lists all API endpoints, grouped into Public, Protected (user), and Admin routes.

---

## 🔓 Public Routes

### POST /api/signup

**Description:** Register a new user.
**Request body:**

```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Response:**

```json
{
  "message": "User created successfully"
}
```

### POST /api/login

**Description:** Log in and receive JWT token.
**Request body:**

```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Response:**

```json
{
  "token": "<jwt-token>"
}
```

### POST /api/logout

**Description:** Logout user (invalidate session/token).

### GET /api/publicFiles

**Description:** List all files marked as public.
**Response:**

```json
[
  {
    "id": 12,
    "filename": "demo.pdf",
    "uploader": "alice",
    "created_at": "2025-09-21T10:00:00Z"
  }
]
```

### GET /api/fileDetails/{id}

**Description:** Get details of a file by ID. Works for guests (soft auth) and logged-in users.

---

## 🔑 Protected User Routes (Auth Required)

### GET /api/me

**Description:** Refresh session / get current user info.
**Auth:** Bearer token required.
**Response:**

```json
{
  "id": 1,
  "email": "user@example.com",
  "role": "user"
}
```

### POST /api/upload

**Description:** Upload one or multiple files.
**Auth:** Bearer token required.
**Request:** `multipart/form-data` with key `files[]`.
**Response:**

```json
{
  "uploaded": [
    {
      "filename": "demo.pdf",
      "hash": "abc123...",
      "deduplicated": false
    }
  ]
}
```

### GET /api/files

**Description:** List files uploaded by the logged-in user.
**Auth:** Bearer token required.

### GET /api/fileDownload/{id}

**Description:** Download a file by ID (owned or shared with user).
**Auth:** Bearer token required.
**Response:** File binary stream.

### GET /api/fileDelete/{id}

**Description:** Delete a file (only by uploader).
**Auth:** Bearer token required.
**Response:**

```json
{ "message": "File deleted" }
```

### GET /api/fileTogglePrivacy/{id}

**Description:** Toggle file privacy (public ↔ private).
**Auth:** Bearer token required.
**Response:**

```json
{ "message": "File privacy updated" }
```

---

## 🛡 Admin Routes (Auth + Admin Role)

### GET /api/adminFiles

**Description:** View all files uploaded by all users.
**Auth:** Bearer token required, admin role.

### POST /api/makeAdmin

**Description:** Promote a user to admin role.
**Auth:** Bearer token required, admin role.

### POST /api/makeUser

**Description:** Demote an admin back to normal user.
**Auth:** Bearer token required, admin role.
