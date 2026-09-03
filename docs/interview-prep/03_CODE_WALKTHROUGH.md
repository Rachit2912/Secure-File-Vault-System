# 03. Code Walkthrough, Data Model & APIs

> **Source of Truth Notice**: This document combines Part 3 (Data Model & Database Design), Part 4 (API Inventory & Internals), and Part 9 (Code-Level Walkthroughs & Interview Defence). All technical claims are grounded in `backend/internal/` and SQL migrations.

---

# PART 3 — DATA MODEL & DATABASE DESIGN

## Database Schema (PostgreSQL 15)

The database schema is defined in plain SQL migrations in `backend/internal/db/migrations/`:
1. `001_init.up.sql` -> Creates `users` & `files` tables, inserts default admin (`root_rachit`).
2. `002_add_last_login_to_users.up.sql` -> Adds `last_login TIMESTAMP`.
3. `003_add_profile_picture_to_users.up.sql` -> Adds `profile_picture TEXT`.
4. `004_add_is_active_to_users.up.sql` -> Adds `is_active BOOLEAN NOT NULL DEFAULT TRUE`.
5. `005_add_file_description_to_files.up.sql` -> Adds `description TEXT`.

### Table 1: `users`
| Column | Type | Constraints | Description |
| :--- | :--- | :--- | :--- |
| `id` | `SERIAL` | `PRIMARY KEY` | Auto-incrementing integer user ID |
| `username` | `TEXT` | `NOT NULL UNIQUE` | Unique handle for auth and filtering |
| `email` | `TEXT` | `NOT NULL UNIQUE` | Unique email address |
| `password` | `TEXT` | `NOT NULL` | Bcrypt-hashed password string |
| `created_at` | `TIMESTAMP` | `DEFAULT CURRENT_TIMESTAMP` | Registration timestamp |
| `role` | `VARCHAR(20)` | `NOT NULL DEFAULT 'user'` | Role (`'user'` or `'admin'`) |
| `last_login` | `TIMESTAMP` | `NULLABLE` | Updated when user logs in (Migration 002) |
| `profile_picture`| `TEXT` | `NULLABLE` | URL or path for avatar (Migration 003) |
| `is_active` | `BOOLEAN` | `NOT NULL DEFAULT TRUE` | Soft-delete / account status (Migration 004) |

### Table 2: `files`
| Column | Type | Constraints | Description |
| :--- | :--- | :--- | :--- |
| `id` | `SERIAL` | `PRIMARY KEY` | Auto-incrementing integer file ID |
| `filename` | `TEXT` | `NOT NULL` | Original display name of uploaded file |
| `filepath` | `TEXT` | `NOT NULL` | Relative path in Supabase bucket (`files/<timestamp>_<name>`) |
| `hash` | `TEXT` | `NOT NULL` | SHA-256 hex hash of file content |
| `size` | `BIGINT` | `NOT NULL` | File size in bytes |
| `uploaded_at` | `TIMESTAMP` | `DEFAULT CURRENT_TIMESTAMP` | Upload timestamp |
| `user_id` | `INT` | `NOT NULL REFERENCES users(id) ON DELETE CASCADE` | Foreign Key pointing to uploader |
| `reference_count`| `BIGINT` | `NOT NULL DEFAULT 1` | Count of records referencing this object |
| `is_master` | `BOOLEAN` | `NOT NULL DEFAULT TRUE` | `TRUE` for physical storage master, `FALSE` for dup |
| `mime_type` | `TEXT` | `NULLABLE` | Detected HTTP content type |
| `is_public` | `BOOLEAN` | `NOT NULL DEFAULT FALSE` | `TRUE` if file is publicly viewable/downloadable |
| `download_count` | `INT` | `NOT NULL DEFAULT 0` | Incrementing counter on each file download |
| `description` | `TEXT` | `NULLABLE` | Optional user description (Migration 005) |

---

## Relationships & Schema Analysis

- **`users` -> `files`**: One-to-Many (`1:N`). Enforced by Foreign Key `files.user_id REFERENCES users(id) ON DELETE CASCADE`.
- **Deduplication Pattern**:
  - Master File: `is_master = TRUE`, `reference_count = N` (where N is total duplicates pointing to it), `filepath = "files/..."`.
  - Linked Duplicates: `is_master = FALSE`, `reference_count = 1` (unused), `filepath = dup.filepath` (points to master's cloud path).
  - Deduplication matching is queried via:
    `SELECT ... FROM files WHERE mime_type=$1 AND size=$2 AND is_master=TRUE`.

### Schema Design Critique & Interview Defence

1. **Why was it designed this way?**
   - Simple master/duplicate model stored in a single table avoids needing a separate `file_objects` or `blobs` table. Quick to implement for a prototype.
2. **What are the weaknesses?**
   - **Missing Database Indexes**: There are NO secondary indexes created in migrations! Queries filtering by `mime_type`, `size`, `is_master`, `hash`, `user_id`, or `is_public` perform full table scans (`O(N)`).
   - **Data Redundancy in Deduplication**: Non-master duplicate rows copy the master's `filepath` string and duplicate `hash`.
   - **Race Conditions**: Two simultaneous uploads of the same duplicate file can both select 0 candidates and both write duplicate master objects to cloud storage.
   - **Non-Atomic Master Promotion**: Deleting a master file performs multi-step un-isolated `SELECT` and `UPDATE` statements without an explicit SQL Transaction (`BEGIN...COMMIT`).
3. **How would I redesign it for production?**
   - Normalize into two distinct tables:
     - `storage_objects(id, hash, size, mime_type, storage_path, reference_count, created_at)`
     - `user_files(id, user_id, storage_object_id, display_filename, is_public, download_count, created_at)`
   - Add B-Tree Indexes:
     - `CREATE UNIQUE INDEX idx_storage_objects_hash ON storage_objects(hash);`
     - `CREATE INDEX idx_user_files_user_id ON user_files(user_id);`
     - `CREATE INDEX idx_user_files_public ON user_files(is_public) WHERE is_public = TRUE;`
4. **SQL vs. NoSQL Reasoning**:
   - **SQL (PostgreSQL)** is the ideal choice here because file metadata requires strict relational consistency (foreign keys mapping files to users, ACID guarantees for quota updates, and transactional reference counting during file deletion). A NoSQL document store (like MongoDB) would require manual application-level join handling and lacks multi-document atomic constraints out of the box.

---

# PART 4 — API INVENTORY

An inventory of all backend API endpoints reverse-engineered directly from `cmd/server/main.go`:

| Method | Path | Purpose | Auth Required? | Request Payload | Response Payload | Code Location |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **POST** | `/api/signup` | Register new user account | No | JSON: `{username, email, password}` | JSON: `{status, msg}` | `handlers/auth.go:SignupHandler` |
| **POST** | `/api/login` | Authenticate & issue JWT cookie | No | JSON: `{username, password}` | JSON: `{status, msg}` + Cookie `token` | `handlers/auth.go:LoginHandler` |
| **POST** | `/api/logout` | Clear auth cookie | No | None | JSON: `{message}` (Cookie expired) | `handlers/auth.go:LogoutHandler` |
| **GET** | `/api/me` | Fetch authenticated session user | Yes (AuthMiddleware) | None (Cookie `token`) | JSON: `{id, username, email, role}` | `handlers/auth.go:RefershHandler` |
| **POST** | `/api/upload` | Upload file (dedup + quota check) | Yes (Auth + RateLimit) | `multipart/form-data` (`file`) | JSON: `{status, hash}` | `handlers/files.go:UploadHandler` |
| **GET** | `/api/files` | List current user's files + stats | Yes (Auth + RateLimit) | Query params (filters) | JSON: `{files: [...], dedupSize, ...}` | `handlers/files.go:FilesHandler` |
| **GET** | `/api/fileDetails/{id}`| Fetch file metadata & uploader info| SoftAuth (Guest / User)| Path var `id` | JSON: FileMeta struct | `handlers/files.go:FileDetailHandler` |
| **GET** | `/api/fileDownload/{id}`| Download binary file content | SoftAuth + RateLimit | Path var `id` | Binary octet-stream + Attachment header | `handlers/files.go:FileDownloadHandler` |
| **GET** | `/api/fileDelete/{id}`| Delete file / decrement ref count | Yes (Auth + RateLimit) | Path var `id` | JSON: `{success: true}` | `handlers/files.go:FileDeleteHandler` |
| **GET** | `/api/fileTogglePrivacy/{id}`| Toggle file public/private state | Yes (Auth + RateLimit) | Path var `id` | JSON: `{success, is_public}` | `handlers/files.go:FileTogglePrivacyHandler`|
| **GET** | `/api/publicFiles` | List all public files globally | No | None | JSON: `{files: [...], total}` | `handlers/public.go:PublicFilesHandler` |
| **GET** | `/api/adminFiles` | Admin list all files across app | Yes (Auth Admin + Rate) | Query params (filters) | JSON: `{files: [...], dedupSize, ...}` | `handlers/admin.go:AdminFilesHandler` |
| **POST** | `/api/makeAdmin` | Promote user to admin | Yes (Auth Admin + Rate) | JSON: `{username}` | JSON: `{status, username, newRole}` | `handlers/admin.go:MakeAdminHandler` |
| **POST** | `/api/makeUser` | Demote admin to normal user | Yes (Auth Admin + Rate) | JSON: `{username}` | JSON: `{status, username, newRole}` | `handlers/admin.go:MakeUserHandler` |

---

# PART 9 — CODE-LEVEL INTERVIEW DEFENCE (15 KEY FUNCTIONS)

Below are the 15 most important code functions in the project with complete logic breakdowns, design weaknesses, and interview defence points.

---

### 1. `cmd/server/main.go: main()`
- **Purpose**: Server entry point that bootstraps application configuration, storage SDK, database connection pool, Gorilla Mux routes, middleware wrappers, and HTTP listener.
- **Input**: Environment variables (`PORT`, `DB_URL`, `JWT_KEY`, etc.).
- **Output**: Running HTTP server listening on configured port (default 8080).
- **Step-by-Step Logic**:
  1. Calls `config.LoadConfig()`.
  2. Verifies `JWT_KEY` is set; logs fatal error if empty.
  3. Initializes Supabase storage client via `storage.Init()`.
  4. Connects to PostgreSQL via `db.Connect()`.
  5. Instantiates `gorilla/mux.NewRouter()`.
  6. Registers public routes (`/api/signup`, `/api/login`, `/api/logout`, `/api/publicFiles`).
  7. Registers soft-auth routes (`/api/fileDetails/{id}`).
  8. Registers protected routes wrapped with `middleware.AuthMiddleware` and `middleware.RateLimitMiddleware`.
  9. Starts server via `http.ListenAndServe(":"+port, middleware.CORS(r))`.
- **Failure Cases**: Missing `JWT_KEY` halts app. Invalid DB connection string or DB downtime causes `log.Fatal`.
- **Design Weaknesses**: Uses `log.Fatal` instead of graceful shutdown handling (`os.Signal` listening for `SIGTERM`/`SIGINT`).
- **How to Improve**: Add `http.Server` with `Shutdown(ctx)` context timeout to safely drain connections during deployment restarts.

---

### 2. `internal/middleware/auth.go: AuthMiddleware()`
- **Purpose**: Enforces strict JWT authentication for protected API endpoints.
- **Input**: `http.Handler` next. Checks `token` cookie in `r.Cookie("token")`.
- **Output**: Passes request to `next` with `userID` and `role` stored in `r.Context()`, or returns HTTP 401 Unauthorized.
- **Step-by-Step Logic**:
  1. Reads `JWT_KEY` from `AppConfig`.
  2. Reads `token` cookie from request.
  3. Parses JWT with claims struct `models.Claims{}` and validates signing method `HS256`.
  4. If valid, injects `claims.UserID` into context key `ContextUserIDKey` and `claims.Role` into `ContextUserRoleKey`.
  5. Calls `next.ServeHTTP(w, r.WithContext(ctx))`.
- **Design Problems**: Token expiration is hardcoded to 5 minutes without a refresh token mechanism.
- **Interview Question**: "Why use HTTP-Only cookies instead of Authorization Bearer header?"
- **Answer**: HTTP-Only cookies protect tokens from XSS (JavaScript cannot access `document.cookie`). To protect against CSRF, the cookie uses `SameSite` policies and CORS allowlists.

---

### 3. `internal/middleware/soft_auth.go: SoftAuthMiddleware()`
- **Purpose**: Enables optional authentication on public endpoints (e.g., file detail & download). If a token is provided, user context is attached; if not, request proceeds as a guest.
- **Input**: HTTP request.
- **Output**: Context enriched with `userID`/`role` if cookie present & valid; otherwise continues unmodified.
- **Step-by-Step Logic**:
  1. Calls `ParseJWTFromRequest(r)`.
  2. If `err == nil`, attaches `userID` and `role` to context.
  3. Calls `next.ServeHTTP`.
- **Interview Benefit**: Demonstrates understanding of optional context propagation for shared resources.

---

### 4. `internal/middleware/rateLimit.go: RateLimitMiddleware()`
- **Purpose**: Rate limits API requests per authenticated user.
- **Input**: HTTP request with `userID` in context.
- **Output**: Calls `next.ServeHTTP` if under limit; otherwise HTTP 429 Too Many Requests.
- **Step-by-Step Logic**:
  1. Reads `ApiRateLimit` config (default 2 req/sec).
  2. Extracts `userID` from context.
  3. Locks global mutex `mu`.
  4. Looks up or instantiates `rate.NewLimiter(rate.Limit(rateLimit), rateLimit)` in map `limiters[userID]`.
  5. Updates `lastSeen` timestamp and unlocks mutex.
  6. Checks `ul.limiter.Allow()`. If false, returns JSON 429 error.
  7. Background goroutine runs every 5 minutes to clean up limiters inactive for >5 mins.
- **Design Problems**: In-memory map fails when scaling backend across multiple server instances (each node tracks rate limits independently). Also, mutex contention under high traffic.
- **Improvement**: Replace with Redis fixed-window or sliding-window rate limiter.

---

### 5. `internal/middleware/cors.go: CORS()`
- **Purpose**: Handles Cross-Origin Resource Sharing (CORS) headers and preflight `OPTIONS` requests.
- **Input**: HTTP request.
- **Output**: Writes CORS headers (`Access-Control-Allow-Origin`, `Credentials`, `Methods`, `Headers`).
- **Step-by-Step Logic**:
  1. Reads `Origin` header.
  2. Checks allowlist (`localhost:3000`, `localhost:5173`, `FRONTEND_URL`).
  3. If matched, sets `Access-Control-Allow-Origin: origin` and `Vary: Origin`.
  4. Sets `Access-Control-Allow-Credentials: true`.
  5. If `r.Method == "OPTIONS"`, responds HTTP 200 OK immediately.
- **Security Check**: Dynamically matches origins rather than wildcard `*` (which is incompatible with `Allow-Credentials: true`).

---

### 6. `internal/handlers/auth.go: SignupHandler()`
- **Purpose**: Registers a new user.
- **Input**: JSON payload `{username, email, password}`.
- **Output**: JSON status message or HTTP 400 error.
- **Step-by-Step Logic**:
  1. Decodes request body into `models.User`.
  2. Calls `bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)`.
  3. Executes `INSERT INTO users (username, email, password, role) VALUES ($1, $2, $3, 'user')`.
  4. Returns success JSON.
- **Complexity**: Time: `O(2^cost)` for bcrypt hashing. Space: `O(1)`.

---

### 7. `internal/handlers/auth.go: LoginHandler()`
- **Purpose**: Authenticates user credentials and sets signed JWT cookie.
- **Input**: JSON payload `{username, password}`.
- **Output**: HTTP-Only cookie `token` and success response.
- **Step-by-Step Logic**:
  1. Decodes request body.
  2. Queries DB: `SELECT id, password, role FROM users WHERE username=$1`.
  3. Executes `bcrypt.CompareHashAndPassword`.
  4. Calls `services.GenerateJWT(id, u.Username, role)`.
  5. Sets cookie `token` with 5-minute expiry, `HttpOnly: true`.
  6. Returns success JSON.

---

### 8. `internal/handlers/files.go: UploadHandler()`
- **Purpose**: Uploads file, checks for deduplication, enforces storage quota, uploads to Supabase, and updates DB.
- **Input**: `multipart/form-data` with `file`.
- **Output**: JSON status (`"duplicate-linked"` or `"new-upload"`).
- **Step-by-Step Logic**: (Detailed in Part 2, Use Case 2).
- **Complexity**: Time: `O(C)` where C is candidate master files hashed; Space: `O(B)` where B is 10MB memory buffer.
- **Failure Cases**:
  - Storage upload succeeds but DB insert fails -> Compensation logic `storage.DeleteFile(storagePath)` is invoked.
  - Concurrent duplicate upload -> Both think file is new, uploading twice.
- **Improvement**: Add SQL transaction and DB UNIQUE constraint on `hash` for master files.

---

### 9. `internal/services/dedup.go: ComputeHash()` & `FindDuplicate()`
- **Purpose**: Calculates SHA-256 hash of multipart file stream and compares against candidates.
- **Input**: `multipart.File` stream and slice of candidate `*models.File` structs.
- **Output**: SHA-256 hex hash string and matching candidate pointer (or `nil`).
- **Step-by-Step Logic**:
  1. `ComputeHash`: Creates `sha256.New()`, copies file stream using `io.Copy`, encodes to hex string.
  2. `FindDuplicate`: Iterates candidates. If `candidate.Hash == calculatedHash`, returns candidate.

---

### 10. `internal/handlers/files.go: FileDeleteHandler()`
- **Purpose**: Handles file deletion, reference count decrementing, master promotion, and storage object cleanup.
- **Input**: File ID path parameter.
- **Output**: JSON success status.
- **Step-by-Step Logic**: (Detailed in Part 2, Use Case 5).
- **Design Problems**: Multi-step DB queries executed outside of SQL transactions can lead to race conditions if multiple users delete linked files concurrently.

---

### 11. `internal/handlers/files.go: FileDownloadHandler()`
- **Purpose**: Enforces access control and streams file bytes from Supabase Storage to client.
- **Input**: File ID path parameter.
- **Output**: File bytes with headers (`Content-Disposition: attachment`, `Content-Type`, `Content-Length`).
- **Step-by-Step Logic**:
  1. Reads file metadata from DB (`user_id`, `filename`, `filepath`, `size`, `mime_type`, `is_public`).
  2. If `!is_public`, verifies `userID` from context matches `fileOwnerID`.
  3. Calls `storage.DownloadFile(storagePath)`.
  4. Increments `download_count` in DB (`UPDATE files SET download_count = download_count + 1`).
  5. Writes binary response headers and data bytes to `w`.

---

### 12. `internal/handlers/files.go: FileTogglePrivacyHandler()`
- **Purpose**: Toggles `is_public` boolean column for a file owned by the requesting user.
- **Input**: File ID path parameter.
- **Output**: JSON `{success: true, is_public: newPrivacy}`.

---

### 13. `internal/handlers/admin.go: AdminFilesHandler()`
- **Purpose**: Allows admin users to list all files across all users with dynamic SQL filtering and platform-wide deduplication savings calculation.
- **Input**: Query parameters (`search`, `mimeType`, `minSize`, `maxSize`, `startDate`, `endDate`, `uploader`).
- **Output**: JSON with file list, `originalSize`, `dedupSize`, and `saveSize`.
- **Dynamic Query Logic**: Uses parameterized placeholders `$1, $2` dynamically appended to prevent SQL Injection.

---

### 14. `internal/utils/mime.go: ValidateMIME()`
- **Purpose**: Validates that file extension matches detected file content type.
- **Input**: `filename string`, `header []byte` (first 512 bytes).
- **Output**: `nil` if valid; error if mismatch.
- **Step-by-Step Logic**:
  1. Detects content type via `http.DetectContentType(header)`.
  2. Extracts extension via `filepath.Ext(filename)`.
  3. Obtains allowed extensions via `mime.ExtensionsByType(detected)`.
  4. Returns error if file extension does not match detected extensions.

---

### 15. `internal/utils/quota.go: GetUserQuotaBytes()`
- **Purpose**: Calculates maximum storage quota per user in bytes.
- **Input**: None (reads `config.AppConfig.UserQuotaMB`).
- **Output**: `int64` bytes (`MB * 1024 * 1024`). Default: 10 MB (10,485,760 bytes).
