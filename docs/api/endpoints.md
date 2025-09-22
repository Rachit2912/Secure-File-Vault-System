# 📌 Authentication Endpoints :

### **POST /api/signup**

**Handler:** `SignupHandler`

- **Request body**

```json
{
  "username": "alice",
  "email": "alice@example.com",
  "password": "mypassword"
}
```

- **Response (200 OK)**

```json
{
  "status": "ok",
  "msg": "user created"
}
```

- **Errors**

  - `400 Bad Request` → invalid JSON, duplicate user, DB error
  - `405 Method Not Allowed` → if not POST

---

### **POST /api/login**

**Handler:** `LoginHandler`

- **Request body**

```json
{
  "username": "alice",
  "password": "mypassword"
}
```

- **Response (200 OK)**

```json
{
  "status": "ok",
  "msg": "login success"
}
```

- JWT is set in a cookie named **`token`**.

- **Errors**

  - `400 Bad Request` → invalid JSON
  - `401 Unauthorized` → user not found or wrong password
  - `417 Expectation Failed` → failed to generate JWT
  - `405 Method Not Allowed` → if not POST

---

### **POST /api/logout**

**Handler:** `LogoutHandler`

- **Request body:** _(none, just requires valid token cookie)_

- **Response (200 OK)**

```json
{
  "message": "Logged out successfully"
}
```

- Clears the `token` cookie.

---

### **GET /api/me**

**Handler:** `RefershHandler`

- **Request:** _(JWT token required in cookie, set by login)_

- **Response (200 OK)**

```json
{
  "id": 1,
  "username": "alice",
  "email": "alice@example.com",
  "role": "user"
}
```

- **Errors**

  - `401 Unauthorized` → missing/invalid token
  - `404 Not Found` → user not found

---

# 📌 File Endpoints

---

### **GET /api/files**

**Handler:** `FilesHandler`

- **Request**

```json
{
  "token": "<jwt-token>"
}
```

- Optional query parameters:
  `search`, `mimeType`, `minSize`, `maxSize`, `startDate`, `endDate`, `uploader`

- **Response**

```json
{
  "files": [
    {
      "id": 12,
      "filename": "report.pdf",
      "size": 204800,
      "uploaded_at": "2025-09-22T12:00:00Z",
      "deduplicated": true,
      "uploader": "alice",
      "is_public": false
    }
  ],
  "dedupSize": 102400,
  "originalSize": 204800,
  "saveSize": 102400
}
```

---

### **POST /api/upload**

**Handler:** `UploadHandler`

- **Request (multipart/form-data)**

```
{
  "token": "<jwt-token>",
  "file": "<binary-file>"
}
```

- **Response (new upload)**

```json
{
  "status": "new-upload",
  "hash": "a7c93f..."
}
```

- **Response (duplicate link)**

```json
{
  "status": "duplicate-linked",
  "hash": "a7c93f..."
}
```

- **Response (quota exceeded, 403)**

```json
{
  "error": "Storage quota exceeded",
  "allowed": "10 MB",
  "used": "12.50 MB"
}
```

---

### **GET /api/fileDelete/{id}**

**Handler:** `FileDeleteHandler`

- **Request**

```json
{
  "token": "<jwt-token>",
  "file_id": 12
}
```

- **Response**

```json
{
  "success": true
}
```

---

### **GET /api/fileDownload/{id}**

**Handler:** `FileDownloadHandler`

- **Request**

```json
{
  "token": "<jwt-token>",
  "file_id": 12
}
```

- **Response:**
  Binary file stream with headers:

```
Content-Disposition: attachment; filename="filename.extension"
Content-Type: application/octet-stream
```

---

### **GET /api/fileTogglePrivacy/{id}**

**Handler:** `FileTogglePrivacyHandler`

- **Request**

```json
{
  "token": "<jwt-token>",
  "file_id": 12
}
```

- **Response**

```json
{
  "success": true,
  "is_public": true
}
```

---

### **GET /api/fileDetails/{id}**

**Handler:** `FileDetailHandler`

- **Request**

```json
{
  "file_id": 12,
  "token": "<jwt-token-or-empty-for-public>"
}
```

- **Response**

```json
{
  "id": 12,
  "filename": "report.pdf",
  "size": 204800,
  "uploaded_at": "2025-09-22T12:00:00Z",
  "uploader_id": 1,
  "is_public": true,
  "deduplicated": false,
  "mime_type": "application/pdf",
  "hash": "a7c93f..."
}
```

---

# 📌 Public Endpoints :

---

### **GET /api/publicFiles**

**Handler:** `PublicFilesHandler`

- **Request**

```json
{}
```

- **Response (200 OK)**

```json
{
  "files": [
    {
      "id": 12,
      "filename": "demo.pdf",
      "size": 204800,
      "uploaded_at": "2025-09-22T12:00:00Z",
      "is_master": true,
      "uploader": "alice",
      "download_count": 5
    },
    {
      "id": 13,
      "filename": "notes.txt",
      "size": 10240,
      "uploaded_at": "2025-09-21T10:30:00Z",
      "is_master": false,
      "uploader": "bob",
      "download_count": 2
    }
  ],
  "total": 2
}
```

- **Errors**

  - `405 Method Not Allowed` → if not GET
  - `500 Internal Server Error` → DB query/scan issues

---

# 📌 Admin Endpoints

---

### **GET /api/adminFiles**

**Handler:** `AdminFilesHandler`

- **Request**

```json
{
  "token": "<jwt-token>",
  "role": "admin"
}
```

Optional query parameters: `search`, `mimeType`, `minSize`, `maxSize`, `startDate`, `endDate`, `uploader`

- **Response**

```json
{
  "files": [
    {
      "id": 12,
      "filename": "report.pdf",
      "size": 204800,
      "uploaded_at": "2025-09-22T12:00:00Z",
      "deduplicated": true,
      "uploader": "alice",
      "is_public": true
    }
  ],
  "dedupSize": 102400,
  "originalSize": 204800,
  "saveSize": 102400
}
```

- **Errors**

  - `403 Forbidden` → if not admin
  - `405 Method Not Allowed` → if not GET
  - `500 Internal Server Error` → DB query issues

---

### **POST /api/makeAdmin**

**Handler:** `MakeAdminHandler`

- **Request**

```json
{
  "token": "<jwt-token>",
  "role": "admin",
  "username": "bob"
}
```

- **Response**

```json
{
  "status": "ok",
  "username": "bob",
  "newRole": "admin"
}
```

- **Errors**

  - `400 Bad Request` → invalid input
  - `403 Forbidden` → if not admin
  - `500 Internal Server Error` → DB error

---

### **POST /api/makeUser**

**Handler:** `MakeUserHandler`

- **Request**

```json
{
  "token": "<jwt-token>",
  "role": "admin",
  "username": "bob"
}
```

- **Response**

```json
{
  "status": "ok",
  "username": "bob",
  "newRole": "user"
}
```

- **Errors**

  - `400 Bad Request` → invalid input
  - `403 Forbidden` → if not admin
  - `500 Internal Server Error` → DB error

---
