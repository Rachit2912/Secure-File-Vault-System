# 04. System Design Learning

> **Source of Truth Notice**: This document analyzes system design patterns reverse-engineered from this repository. Each concept is structured in the required learning format to prepare you for backend system design interview discussions.

---

## 1. Content-Based File Deduplication (SHA-256)

**CONCEPT**: Content-Based File Deduplication
**WHERE IT APPEARS IN MY CODE**: `backend/internal/services/dedup.go:ComputeHash()`, `FindDuplicate()`, and `backend/internal/handlers/files.go:UploadHandler()`.
**WHY IT EXISTS**: To avoid storing multiple identical copies of files uploaded by different users (or the same user multiple times).
**WHAT PROBLEM IT SOLVES**: Storage bloat, increased cloud bandwidth costs, and slow upload times for existing files.
**REAL FAILURE SCENARIO**:
"If two users upload the exact same 1GB file at the exact same millisecond, I would observe both requests selecting 0 candidates, both passing the deduplication check, and both uploading identical physical files to Supabase Storage. This indicates a concurrent write race condition caused by missing database locks (`SELECT FOR UPDATE`) or a unique hash constraint on master records."
**WHAT I WOULD DO**:
"If low traffic -> choose database-level unique constraint on `hash` with `ON CONFLICT` handling because it enforces atomicity with minimal complexity."
"If high concurrent traffic -> choose a Redis distributed lock (`Redlock`) keyed on `file_hash` because it serializes duplicate uploads across distributed backend instances before cloud storage calls occur."
**TRADE-OFF**: Calculating SHA-256 consumes CPU cycles and streams the entire file payload before determining if it is a duplicate.
**INTERVIEWER FOLLOW-UP QUESTIONS**:
- "Why SHA-256 instead of MD5 or SHA-1?" -> SHA-256 is cryptographically collision-resistant, whereas MD5 and SHA-1 have known collision exploits.
- "What if two different files produce the same SHA-256 hash?" -> SHA-256 collision probability is 1 in 2^256 (virtually zero). To be 100% safe, a byte-level byte-by-byte comparison can be added if hashes match.
**GOOD 2-4 LINE ANSWER**:
> "In FileVault, deduplication works by computing a SHA-256 hash of uploaded file streams after filtering candidate files by MIME type and size. If a master record with the same hash exists, we increment its reference count and link a new database record without writing to cloud storage. This saves up to 100% of storage costs for redundant uploads."

---

## 2. In-Memory Token Bucket Rate Limiting

**CONCEPT**: Rate Limiting (Token Bucket)
**WHERE IT APPEARS IN MY CODE**: `backend/internal/middleware/rateLimit.go:RateLimitMiddleware()`. Uses `golang.org/x/time/rate`.
**WHY IT EXISTS**: To protect backend routes from API abuse, brute-force attacks, and server resource exhaustion.
**WHAT PROBLEM IT SOLVES**: Denial of Service (DoS) attacks, rapid polling, and API quota abuse.
**REAL FAILURE SCENARIO**:
"If I deploy 3 backend server instances behind a load balancer, I would observe a user successfully sending 30 requests per second when the limit is set to 10. This indicates that rate limiters are stored independently in local node memory rather than in a shared state store."
**WHAT I WOULD DO**:
"If single node -> choose `golang.org/x/time/rate` with background cleanup goroutines because it requires zero external dependencies."
"If distributed multi-instance deployment -> choose Redis sliding-window rate limiting using Lua scripts because Redis provides atomic, centralized rate limit tracking across all backend nodes."
**TRADE-OFF**: Local memory rate limiting is extremely fast (microsecond latency) but state is lost when the app restarts and does not synchronize across instances.
**INTERVIEWER FOLLOW-UP QUESTIONS**:
- "What happens if a user is not logged in?" -> The current implementation rejects the request because it relies on `userID` from context. For unauthenticated endpoints, IP-based rate limiting (`r.RemoteAddr` / `X-Forwarded-For`) should be used.
**GOOD 2-4 LINE ANSWER**:
> "I implemented rate limiting using Go's `golang.org/x/time/rate` token bucket algorithm in a custom middleware. It maintains a mutex-protected map of limiters per user ID and cleans up inactive entries every 5 minutes. In a distributed environment, I would transition this state store to Redis."

---

## 3. Storage Quota Enforcement

**CONCEPT**: Multi-Tenant Quota Management
**WHERE IT APPEARS IN MY CODE**: `backend/internal/utils/quota.go:GetUserQuotaBytes()` and `backend/internal/handlers/files.go:UploadHandler()`.
**WHY IT EXISTS**: To prevent single users from consuming all available disk/cloud storage space.
**WHAT PROBLEM IT SOLVES**: Storage exhaustion and runaway cloud storage bills.
**REAL FAILURE SCENARIO**:
"If a user opens 10 parallel browser tabs and uploads ten 9MB files simultaneously when their quota is 10MB, I would observe all 10 requests querying `COALESCE(SUM(size), 0)` at the same time, seeing 0MB used, and all 10 uploads succeeding (storing 90MB). This indicates a race condition due to uncommitted reads and missing row locks."
**WHAT I WOULD DO**:
"If standard relational DB -> choose SQL row locking (`SELECT total_bytes_used FROM user_quotas WHERE user_id=$1 FOR UPDATE`) before checking allowance because it serializes quota checks."
"If high-throughput event-driven architecture -> choose atomic Redis increment (`INCRBY user:quota:<id> <size>`) before processing the upload payload."
**TRADE-OFF**: Summing file sizes on every upload (`SUM(size) WHERE user_id=$1`) without an index leads to slower uploads as user file counts grow.
**INTERVIEWER FOLLOW-UP QUESTIONS**:
- "How would you optimize the quota query?" -> Denormalize by adding `storage_used_bytes` column to the `users` table and updating it atomically on upload/delete.
**GOOD 2-4 LINE ANSWER**:
> "Quota enforcement checks the sum of a user's uploaded files against an environment-configured limit (e.g., 10MB) before accepting new uploads. To prevent race conditions in production, I would track user usage in a denormalized column with atomic database updates or Redis counters."

---

## 4. JWT Authentication & HTTP-Only Cookie Session Management

**CONCEPT**: Stateless Authentication via JWT in HTTP-Only Cookies
**WHERE IT APPEARS IN MY CODE**: `backend/internal/services/auth.go:GenerateJWT()`, `backend/internal/middleware/auth.go:AuthMiddleware()`, and `backend/internal/handlers/auth.go:LoginHandler()`.
**WHY IT EXISTS**: To authenticate users statelessly across HTTP requests without querying the database for session data on every request.
**WHAT PROBLEM IT SOLVES**: Database bottleneck from session table lookups and XSS token theft.
**REAL FAILURE SCENARIO**:
"If a user logs out or changes their password, I would observe their old JWT token remaining valid until its 5-minute expiration time. This indicates that stateless JWTs cannot be instantly revoked without a token revocation blacklist."
**WHAT I WOULD DO**:
"If absolute security / immediate revocation is required -> choose Redis-backed token revocation list or short-lived JWTs (15 min) with sliding refresh tokens."
"If pure stateless performance is prioritized -> choose short token TTLs (e.g., 5 minutes) as implemented in this repository."
**TRADE-OFF**: Stateless JWTs avoid database lookups but make instant revocation impossible without adding state.
**INTERVIEWER FOLLOW-UP QUESTIONS**:
- "Why put the JWT in a cookie instead of LocalStorage?" -> LocalStorage is vulnerable to XSS scripts. HTTP-Only cookies cannot be accessed via JavaScript.
- "How do you defend against CSRF with cookies?" -> Use `SameSite=Lax` or `SameSite=Strict` cookies and validate `Origin`/`Referer` headers.
**GOOD 2-4 LINE ANSWER**:
> "Authentication uses HMAC-SHA256 signed JWTs stored in HTTP-Only, SameSite cookies. The backend middleware validates the signature and extracts claims statelessly. This prevents XSS access while keeping the backend decoupled from a centralized session database."

---

## 5. Soft-Authentication & Optional Security Context

**CONCEPT**: Soft-Auth (Optional Authentication Propagation)
**WHERE IT APPEARS IN MY CODE**: `backend/internal/middleware/soft_auth.go:SoftAuthMiddleware()` and `backend/internal/handlers/files.go:FileDownloadHandler()`.
**WHY IT EXISTS**: To support endpoints that serve different behavior for guest users vs. authenticated owners.
**WHAT PROBLEM IT SOLVES**: Duplicating routes for public vs. private access.
**REAL FAILURE SCENARIO**:
"If a guest user accesses `/api/fileDownload/123` for a public file, I would observe the file download succeeding; if they access it for a private file, I would observe a 401 Unauthorized response. This indicates soft-auth correctly populated user context when available while allowing guest access."
**WHAT I WOULD DO**:
"Always use soft-auth for hybrid endpoints where public items are accessible to everyone, but additional details/actions depend on owner identity."
**TRADE-OFF**: Handlers must explicitly handle `nil` user context without crashing or raising panics.
**INTERVIEWER FOLLOW-UP QUESTIONS**:
- "How do you prevent `nil` pointer panics in soft-auth handlers?" -> Safely type-assert context values (e.g., `uidVal, ok := ctx.Value(...).(int)` and check `!ok || uidVal == 0`).
**GOOD 2-4 LINE ANSWER**:
> "Soft-Auth parses JWT tokens if present in request cookies but allows unauthenticated requests to proceed as guests. Handlers then inspect the request context to enforce privacy rules—allowing guest downloads for public files while requiring ownership verification for private files."

---

## 6. Security: File Content MIME Type Detection

**CONCEPT**: Defense-in-Depth File Validation (Magic Bytes Inspection)
**WHERE IT APPEARS IN MY CODE**: `backend/internal/utils/mime.go:ValidateMIME()` and `backend/internal/handlers/files.go:UploadHandler()`.
**WHY IT EXISTS**: Users can rename malicious files (e.g., `shell.php` -> `document.pdf`) to bypass extension-based upload filters.
**WHAT PROBLEM IT SOLVES**: Extension spoofing, arbitrary file upload vulnerabilities, and Remote Code Execution (RCE).
**REAL FAILURE SCENARIO**:
"If an attacker renames `malicious.exe` to `image.png` and uploads it, I would observe `http.DetectContentType` reading the initial 512 magic bytes, detecting `application/x-dosexec`, comparing it with `.png`, and returning HTTP 412 Precondition Failed. This indicates successful magic-byte MIME validation."
**WHAT I WOULD DO**:
"Always inspect raw file header bytes (`http.DetectContentType`) rather than trusting user-supplied `Content-Type` headers or file extensions."
**TRADE-OFF**: Inspecting file headers requires buffering the first 512 bytes and seeking the stream back to index 0 before processing.
**INTERVIEWER FOLLOW-UP QUESTIONS**:
- "Is magic byte detection 100% exploit-proof?" -> No, attackers can construct polyglot files that match valid magic bytes while containing executable payloads. Fully secure systems process files inside isolated sandboxes or run virus scanners (e.g., ClamAV).
**GOOD 2-4 LINE ANSWER**:
> "FileVault performs MIME validation by sniffing the first 512 bytes of the uploaded file payload using `http.DetectContentType`. It then compares the sniffed MIME type against the file's extension to reject spoofed uploads before any bytes reach cloud storage."

---

## 7. Master-Duplicate Promotion & Reference Counting

**CONCEPT**: Copy-on-Write / Reference Count Lifecycle Management
**WHERE IT APPEARS IN MY CODE**: `backend/internal/handlers/files.go:FileDeleteHandler()`.
**WHY IT EXISTS**: When multiple users share a deduplicated file, deleting one user's file must not destroy the cloud storage object for other users.
**WHAT PROBLEM IT SOLVES**: Premature file deletion and broken file references.
**REAL FAILURE SCENARIO**:
"If User A uploads File X (master), User B uploads File X (duplicate), and User A deletes File X, I would observe the system selecting User B's duplicate record, promoting User B's record to `is_master = TRUE`, deleting User A's row, and leaving the physical cloud object intact. This indicates correct master promotion logic."
**WHAT I WOULD DO**:
"If single table schema -> execute master promotion within an isolated SQL transaction (`BEGIN...COMMIT`)."
"If decoupled schema -> separate file metadata (`user_files`) from physical storage records (`storage_objects`) so that deleting a user file simply decrements `storage_objects.reference_count` without needing master promotion."
**TRADE-OFF**: Complex promotion logic in application code increases bug risks compared to a normalized schema with atomic foreign key reference counts.
**GOOD 2-4 LINE ANSWER**:
> "When a file is deleted, FileVault checks if it is a master record with active duplicate references. If reference_count > 1, it promotes an existing duplicate to master and deletes the user's record without touching cloud storage. Only when reference_count reaches 1 is the physical cloud object deleted."

---

## 8. Dynamic SQL Query Construction & Injection Defense

**CONCEPT**: Safe Parameterized Dynamic Query Building
**WHERE IT APPEARS IN MY CODE**: `backend/internal/handlers/admin.go:AdminFilesHandler()` and `backend/internal/handlers/files.go:FilesHandler()`.
**WHY IT EXISTS**: To support multi-field filtering (search string, size ranges, MIME types, date ranges) while preventing SQL injection.
**WHAT PROBLEM IT SOLVES**: SQL Injection attacks and query syntax errors from string concatenation.
**REAL FAILURE SCENARIO**:
"If a user inputs `search = "test' OR '1'='1"`, I would observe the backend generating SQL placeholder `AND f.filename ILIKE $1` with parameter binding `"test' OR '1'='1"`. PostgreSQL treats the input strictly as a string literal. This indicates successful SQL injection defense."
**WHAT I WOULD DO**:
"Always append condition clauses with explicit positional placeholders (`$1, $2, ...`) and pass arguments as a separate slice to `db.Query()`."
**TRADE-OFF**: Manual string building requires carefully tracking argument slice positions (`argPos++`).
**GOOD 2-4 LINE ANSWER**:
> "Dynamic filtering builds SQL queries by conditionally appending `WHERE` clauses with positional parameters (`$1`, `$2`) and storing values in an argument slice. Passing this slice directly to PostgreSQL's prepared driver guarantees complete immunity to SQL injection."

---

## 9. Relational Database Connection Pooling

**CONCEPT**: Database Connection Pool Tuning
**WHERE IT APPEARS IN MY CODE**: `backend/internal/db/db.go:Connect()`.
**WHY IT EXISTS**: Creating a new TCP connection to PostgreSQL for every HTTP request adds significant latency and degrades database performance.
**WHAT PROBLEM IT SOLVES**: Connection overhead, port exhaustion, and database memory starvation.
**REAL FAILURE SCENARIO**:
"If 100 concurrent HTTP requests hit the server, I would observe up to 20 requests executing queries simultaneously while remaining requests queue up for an available connection from the pool. This indicates `SetMaxOpenConns(20)` is constraining active DB connections."
**WHAT I WOULD DO**:
"Configure connection pool limits based on DB server CPU/RAM and expected concurrency (`SetMaxOpenConns(20)`, `SetMaxIdleConns(5)`, `SetConnMaxLifetime(30m)`)."
**TRADE-OFF**: Setting `MaxOpenConns` too low causes request latency under traffic spikes; setting it too high exhausts PostgreSQL process memory.
**GOOD 2-4 LINE ANSWER**:
> "I configured Go's SQL connection pool with `SetMaxOpenConns(20)` and `SetMaxIdleConns(5)`. This reuses active TCP connections across HTTP goroutines, keeping database resource usage predictable and preventing connection exhaustion under load."

---

## 10. CORS & SameSite Cookie Protection

**CONCEPT**: Cross-Origin Resource Sharing (CORS) & Browser Cookie Security
**WHERE IT APPEARS IN MY CODE**: `backend/internal/middleware/cors.go:CORS()` and `backend/internal/handlers/auth.go:LoginHandler()`.
**WHY IT EXISTS**: To allow frontend applications running on different origins (e.g., `localhost:5173`) to send credentials securely while blocking unauthorized cross-origin requests.
**WHAT PROBLEM IT SOLVES**: Cross-Site Request Forgery (CSRF) and unauthorized cross-domain data access.
**REAL FAILURE SCENARIO**:
"If a malicious site `evil.com` sends a fetch request to `/api/files`, I would observe `cors.go` inspecting the `Origin` header, failing to match `evil.com` in the allowlist, and omitting `Access-Control-Allow-Origin`. The browser blocks `evil.com` from reading the response."
**WHAT I WOULD DO**:
"Dynamically validate request `Origin` against a strict allowlist map and return `Vary: Origin` to avoid proxy caching issues."
**TRADE-OFF**: Strictly matching origins requires explicitly configuring all deployed frontend URLs in environment settings.
**GOOD 2-4 LINE ANSWER**:
> "CORS middleware dynamically matches the request `Origin` against an explicitly configured allowlist before issuing `Access-Control-Allow-Credentials: true`. Combined with `SameSite` HTTP-Only cookies, this prevents unauthorized cross-origin API access and CSRF attacks."
