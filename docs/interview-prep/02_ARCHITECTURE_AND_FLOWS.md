# 02. Architecture & Data Flows

> **Source of Truth Notice**: This document details the step-by-step request and data execution flows through the backend based directly on `cmd/server/main.go`, `internal/middleware/`, `internal/handlers/`, `internal/services/`, and `internal/storage/`.

---

## Overview of Core Use Cases

Below are the 6 primary end-to-end use cases implemented in Secure File Vault:
1. **User Authentication (Signup & Login)**
2. **File Upload with Content-Based Deduplication & Quota Enforcement**
3. **File Download with Privacy Enforcement (Soft-Auth vs. Strict Ownership)**
4. **File Privacy Toggle (Public <-> Private)**
5. **File Deletion with Reference Counting & Master Promotion**
6. **Admin Management (Global File Filter & User Role Escalation)**

---

## Use Case 1: User Signup & Login

```
Client (React Form)
  ── POST /api/signup ──> CORS Middleware ──> SignupHandler ──> Bcrypt Hash ──> Postgres INSERT
  ── POST /api/login  ──> CORS Middleware ──> LoginHandler  ──> Bcrypt Compare ──> Generate JWT ──> Set-Cookie: token
```

### Flow 1A: Signup (`POST /api/signup`)
1. **Request Hits**: `POST /api/signup` with JSON body `{"username": "...", "email": "...", "password": "..."}`.
2. **Route Handling**: Defined in `backend/cmd/server/main.go:main()`.
3. **CORS Middleware**: `middleware.CORS` in `backend/internal/middleware/cors.go` checks `Origin` header and sets `Access-Control-Allow-Origin`, `Credentials`, and methods. Handles `OPTIONS` preflight if present.
4. **Handler Execution**: `handlers.SignupHandler` in `backend/internal/handlers/auth.go`:
   - Validates HTTP method (`POST`).
   - Decodes JSON body into `models.User` struct.
   - Hashes password using `bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)` (`golang.org/x/crypto/bcrypt`).
5. **DB Operation**: Executes `INSERT INTO users (username, email, password, role) VALUES ($1, $2, $3, $4)` with `$4 = "user"`.
6. **Response**: Returns JSON `{"status": "ok", "msg": "user created"}` with HTTP 200 OK.
7. **Failure Handling**:
   - Non-POST method -> 450/405 Method Not Allowed (`http.StatusMethodNotAllowed`).
   - Invalid JSON -> 400 Bad Request.
   - Duplicate username/email constraint violation -> 400 Bad Request with `"User already exists or DB error..."`.

### Flow 1B: Login (`POST /api/login`)
1. **Request Hits**: `POST /api/login` with JSON body `{"username": "...", "password": "..."}`.
2. **Handler Execution**: `handlers.LoginHandler` in `backend/internal/handlers/auth.go`:
   - Decodes JSON into `models.User`.
   - Queries DB: `SELECT id, password, role FROM users WHERE username=$1`.
   - Compares hashes: `bcrypt.CompareHashAndPassword([]byte(hashedPwd), []byte(u.Password))`.
3. **Token Generation**: Calls `services.GenerateJWT(id, u.Username, role)` in `backend/internal/services/auth.go`, signing `models.Claims` (expiration 5 minutes) with `HS256` using secret `config.AppConfig.JWTKey`.
4. **Cookie Issuance**: Sets HTTP-Only cookie `token` with `Expires = time.Now().Add(5 * time.Minute)`, `Path = "/"`, `HttpOnly = true`, `SameSite` mode (Lax for HTTP, None for HTTPS), and `Secure` flag.
5. **Response**: Returns `{"status": "ok", "msg": "login success"}`.

---

## Use Case 2: File Upload with Deduplication & Quota Check (`POST /api/upload`)

```
Client (FormData with 'file')
  ── POST /api/upload ──> CORS ──> AuthMiddleware ──> RateLimitMiddleware ──> UploadHandler
                                                                                │
   ┌────────────────────────────────────────────────────────────────────────────┘
   ├── 1. Parse Multipart (10MB limit)
   ├── 2. MIME Validation (http.DetectContentType vs. ext)
   ├── 3. Candidates Lookup (SELECT FROM files WHERE mime_type=$1 AND size=$2 AND is_master=TRUE)
   ├── 4. SHA-256 Hash Computation & Candidate Matching (services.FindDuplicate)
   │     ├─ IF DUPLICATE FOUND:
   │     │    ├─ INSERT INTO files (is_master=FALSE, filepath=dup.filepath)
   │     │    └─ UPDATE files SET reference_count = reference_count + 1 WHERE id=dup.id
   │     │    └─ RETURN {"status": "duplicate-linked"}
   │     └─ IF NEW FILE:
   │          ├─ Quota Check (COALESCE(SUM(size), 0) FROM files WHERE user_id=$1)
   │          ├─ Upload to Supabase Storage (storage.UploadFile)
   │          ├─ INSERT INTO files (is_master=TRUE, reference_count=1)
   │          └─ RETURN {"status": "new-upload"}
```

### Step-by-Step Logic (`backend/internal/handlers/files.go:UploadHandler`):
1. **Request Hits**: `POST /api/upload` containing `multipart/form-data` with form key `file`. Cookie `token` attached.
2. **Middleware Pipeline**:
   - `CORS`: Handles CORS headers.
   - `AuthMiddleware` (`backend/internal/middleware/auth.go`): Parses `token` cookie via `ParseJWTFromRequest`. Validates JWT signature. Extracts `claims.UserID` and `claims.Role` and places them into Request Context under `ContextUserIDKey` & `ContextUserRoleKey`.
   - `RateLimitMiddleware` (`backend/internal/middleware/rateLimit.go`): Obtains/creates token bucket for `userID`. Checks `limiter.Allow()`. If exceeded, returns HTTP 429 Too Many Requests with JSON `{"error": "rate limit exceeded..."}`.
3. **Form Parsing & File Extraction**:
   - Calls `r.ParseMultipartForm(10 << 20)` (10 MB max memory buffer).
   - Reads form file `file, handler, err := r.FormFile("file")`.
4. **MIME Validation**:
   - Reads first 512 bytes into buffer: `buf := make([]byte, 512)`.
   - Calls `utils.ValidateMIME(handler.Filename, buf)` (`backend/internal/utils/mime.go`).
   - Uses `http.DetectContentType(buf)` to read actual file bytes and checks if the file extension matches allowed MIME extensions. If mismatch, returns 412 Precondition Failed.
   - Resets file seek pointer to 0.
5. **Deduplication Candidate Search**:
   - Queries DB for candidate master files: `SELECT id, user_id, filename, filepath, hash, size, mime_type, reference_count, is_master FROM files WHERE mime_type=$1 AND size=$2 AND is_master=TRUE`.
6. **SHA-256 Hash Calculation**:
   - Calls `services.FindDuplicate(file, candidates)` (`backend/internal/services/dedup.go`).
   - `ComputeHash` streams file bytes through `crypto/sha256` and returns hex-encoded SHA-256 hash string.
   - Compares calculated hash against candidate hashes.
7. **Branch A: Duplicate Found (`dup != nil`)**:
   - Inserts duplicate metadata record into DB with `is_master = FALSE`:
     `INSERT INTO files (user_id, filename, filepath, hash, size, mime_type, is_master) VALUES ($1, $2, $3, $4, $5, $6, FALSE)`.
   - Increments master file reference count:
     `UPDATE files SET reference_count = reference_count + 1 WHERE id=$1`.
   - Skips cloud storage upload completely!
   - Returns JSON `{"status": "duplicate-linked", "hash": "..."}` with HTTP 200.
8. **Branch B: New Unique File**:
   - **Quota Enforcement**:
     - Queries user's current total storage: `SELECT COALESCE(SUM(size),0) FROM files WHERE user_id=$1`.
     - Fetches allowed quota bytes via `utils.GetUserQuotaBytes()` (reads `USER_QUOTA_MB` env var, default 10MB).
     - If `used + size > quota`, returns HTTP 430/403 Forbidden with JSON `{"error": "Storage quota exceeded", "allowed": "...", "used": "..."}`.
   - **Cloud Storage Upload**:
     - Generates timestamped path: `storagePath := fmt.Sprintf("files/%d_%s", time.Now().UnixNano(), handler.Filename)`.
     - Calls `storage.UploadFile(storagePath, file, mimeType)` (`backend/internal/storage/files.go`), uploading bytes to Supabase bucket `file-vault`.
   - **DB Master Record Insertion**:
     - Inserts record into DB with `is_master = TRUE` and `reference_count = 1`:
       `INSERT INTO files (user_id, filename, filepath, hash, size, mime_type, reference_count, is_master) VALUES ($1, $2, $3, $4, $5, $6, 1, TRUE)`.
     - **Rollback on Error**: If DB insert fails, executes compensation action `storage.DeleteFile(storagePath)` to prevent orphan cloud objects.
   - Returns JSON `{"status": "new-upload", "hash": "..."}` with HTTP 200.

---

## Use Case 3: Private & Public File Download (`GET /api/fileDownload/{id}`)

```
Client
  ── GET /api/fileDownload/{id} ──> CORS ──> SoftAuthMiddleware ──> RateLimitMiddleware ──> FileDownloadHandler
                                                                                                 │
   ┌─────────────────────────────────────────────────────────────────────────────────────────────┘
   ├── 1. Extract path variable 'id' via mux.Vars(r)
   ├── 2. Query DB: SELECT user_id, filename, filepath, size, mime_type, is_public FROM files WHERE id=$1
   ├── 3. Visibility / Ownership Check:
   │     ├─ IF is_public == FALSE:
   │     │    └─ Verify request context has userID AND userID == fileOwnerID
   │     │    └─ IF not owner or guest: Return 401/403
   │     └─ IF is_public == TRUE:
   │          └─ Allow guest download
   ├── 4. Download file bytes from Supabase Storage: storage.DownloadFile(storagePath)
   ├── 5. Increment Download Counter: UPDATE files SET download_count = download_count + 1 WHERE id=$1
   └── 6. Return binary stream with headers: Content-Disposition attachment, Content-Type, Content-Length
```

### Exact Code References:
- Route: `/api/fileDownload/{id}` (`backend/cmd/server/main.go`).
- Middleware: `SoftAuthMiddleware` -> `RateLimitMiddleware`.
  - `SoftAuthMiddleware` (`backend/internal/middleware/soft_auth.go`) checks if `token` cookie is valid. If valid, injects `userID` and `role` into context. If invalid or missing, allows request to proceed as guest (`uidVal = nil`).
- Handler: `handlers.FileDownloadHandler` in `backend/internal/handlers/files.go`.
- Storage: `storage.DownloadFile(storagePath)` in `backend/internal/storage/files.go`.

---

## Use Case 4: File Privacy Toggle (`GET /api/fileTogglePrivacy/{id}`)

```
Client
  ── GET /api/fileTogglePrivacy/{id} ──> CORS ──> AuthMiddleware ──> RateLimitMiddleware ──> FileTogglePrivacyHandler
                                                                                                   │
   ┌───────────────────────────────────────────────────────────────────────────────────────────────┘
   ├── 1. Extract userID from Context (AuthMiddleware enforced)
   ├── 2. Query DB: SELECT user_id, is_public FROM files WHERE id=$1
   ├── 3. Ownership Check: IF uploaderID != userID -> Return 403 Forbidden
   ├── 4. Update DB: UPDATE files SET is_public = NOT is_public WHERE id=$1
   └── 5. Return JSON: {"success": true, "is_public": newPrivacy}
```

### Exact Code References:
- Route: `/api/fileTogglePrivacy/{id}` (`backend/cmd/server/main.go`).
- Handler: `FileTogglePrivacyHandler` in `backend/internal/handlers/files.go`.

---

## Use Case 5: File Deletion with Reference Counting & Master Promotion (`GET /api/fileDelete/{id}`)

```
Client
  ── GET /api/fileDelete/{id} ──> CORS ──> AuthMiddleware ──> RateLimitMiddleware ──> FileDeleteHandler
                                                                                            │
   ┌────────────────────────────────────────────────────────────────────────────────────────┘
   ├── 1. Extract file ID & verify ownership (uploaderID == userID)
   ├── 2. Query DB: SELECT user_id, filepath, is_master, reference_count FROM files WHERE id=$1
   ├── 3. Branching Logic based on file role:
   │
   │   CASE A: File is NOT Master (is_master == FALSE)
   │     ├─ Decrement master's reference count:
   │     │    UPDATE files SET reference_count = reference_count - 1
   │     │    WHERE hash = (SELECT hash FROM files WHERE id=$1) AND is_master = TRUE
   │     ├─ Delete duplicate DB record: DELETE FROM files WHERE id=$1
   │     └─ Physical storage object remains untouched!
   │
   │   CASE B: File IS Master AND reference_count > 1 (Shared Duplicate Exists)
   │     ├─ Find a non-master duplicate to promote:
   │     │    SELECT id FROM files WHERE hash = (SELECT hash FROM files WHERE id=$1) AND is_master = FALSE LIMIT 1
   │     ├─ Promote duplicate to new master:
   │     │    UPDATE files SET is_master = TRUE, filepath = $1 WHERE id = $newMasterID
   │     ├─ Delete old master DB record: DELETE FROM files WHERE id=$1
   │     └─ Physical storage object remains untouched!
   │
   │   CASE C: File IS Master AND reference_count == 1 (Only Reference)
   │     ├─ Delete DB record: DELETE FROM files WHERE id=$1
   │     └─ Delete physical object from Supabase Storage: storage.DeleteFile(filepathOnDisk)
   │
   └── 4. Return JSON: {"success": true}
```

### Exact Code References:
- Handler: `FileDeleteHandler` in `backend/internal/handlers/files.go`.
- Storage Delete: `storage.DeleteFile` in `backend/internal/storage/files.go`.

---

## Use Case 6: Admin Management (`GET /api/adminFiles`, `POST /api/makeAdmin`)

```
Admin Client
  ── GET /api/adminFiles ──> CORS ──> AuthMiddleware ──> RateLimitMiddleware ──> AdminFilesHandler
                                                                                        │
   ┌────────────────────────────────────────────────────────────────────────────────────┘
   ├── 1. Context Role Check: r.Context().Value(middleware.ContextUserRoleKey) == "admin"
   │      └─ IF role != "admin": Return 403 Forbidden
   ├── 2. Parse Query Filters (search, mimeType, minSize, maxSize, startDate, endDate, uploader)
   ├── 3. Build Dynamic Parameterized SQL Query ($1, $2, ...) JOINing files and users
   ├── 4. Execute Query & Calculate Storage Metrics:
   │      ├─ originalSize = sum(all file sizes)
   │      ├─ dedupSize = sum(master file sizes)
   │      └─ saveSize = originalSize - dedupSize
   └── 5. Return JSON: {"files": [...], "dedupSize": X, "originalSize": Y, "saveSize": Z}
```

### Exact Code References:
- Handlers: `AdminFilesHandler`, `MakeAdminHandler`, `MakeUserHandler` in `backend/internal/handlers/admin.go`.
