# 06. Interview Q&A, Tech Deep-Dives & Personal Pitches

> **Source of Truth Notice**: This document combines Part 8 (Technology Questions), Part 11 (Realistic Interview Q&A across 5 levels), and Part 12 (Personal Interview Explanations & Pitches). Grounded entirely in `backend/` and `frontend/` source code.

---

# PART 8 — TECHNOLOGY DEEP DIVES

## 1. Go (1.20+) & Net/HTTP Standard Library
1. **Why was it used here?** Go provides high-performance, low-latency, concurrent HTTP handling out of the box via lightweight goroutines. It compiles to a single binary, making Docker deployment simple.
2. **Internal behavior**: `net/http` spawns a new goroutine for every incoming TCP connection (`go c.serve(connCtx)`), allowing thousands of concurrent HTTP requests without operating system thread overhead.
3. **Why not Node.js or Python?** Go provides strong static typing, native concurrency without event-loop blocking, and significantly lower memory utilization (~15MB RAM per node vs 150MB+ for Node/Python).
4. **Limitations**: Manual error handling (`if err != nil`) increases boilerplate.
5. **If it fails**: Process panics or exits (`log.Fatal`); managed by Docker restart policies (`restart: unless-stopped`).
6. **Follow-Up**: "How does Go handle goroutine scheduling?" -> M:N scheduler mapping M goroutines onto N OS threads using work-stealing queues.

---

## 2. Gorilla Mux (`github.com/gorilla/mux`)
1. **Why was it used here?** Standard Go `net/http` prior to Go 1.22 lacked URL path parameter parsing (`/api/fileDownload/{id}`). Gorilla Mux provides clean route matching and regex route parameters.
2. **Internal behavior**: Builds a tree of route matchers and iterates match rules sequentially or via path prefix routing.
3. **Why not Gin or Fiber?** Gorilla Mux adheres strictly to Go's standard `http.Handler` and `http.HandlerFunc` interfaces without introducing framework lock-in or custom context types.
4. **Limitations**: Gorilla Mux is currently archived / in maintenance mode.
5. **Follow-Up**: "How would you upgrade this today?" -> Migrate to Go 1.22+ native `net/http` route matching (`http.HandleFunc("GET /api/fileDownload/{id}", ...)`) to remove third-party dependencies.

---

## 3. PostgreSQL 15 & `lib/pq`
1. **Why was it used here?** PostgreSQL provides strict ACID compliance, foreign key constraints (`ON DELETE CASCADE`), and complex dynamic SQL filtering capabilities.
2. **Internal behavior**: `database/sql` maintains a thread-safe connection pool. `lib/pq` communicates with PostgreSQL via its binary wire protocol.
3. **Why not MongoDB or MySQL?** FileVault requires strict relational integrity (mapping files to users, atomic quota checks, and cascade deletions) which MongoDB lacks natively. PostgreSQL offers superior JSON handling and indexing over MySQL.
4. **Limitations**: Relational tables require explicit schema migrations.
5. **Follow-Up**: "What happens if DB connections hit `MaxOpenConns(20)`?" -> Extra goroutines block and wait on a condition variable until a connection is returned to the pool.

---

## 4. Supabase Storage SDK (`github.com/supabase-community/storage-go`)
1. **Why was it used here?** Provides S3-compatible cloud object storage with a simple REST API and SDK wrapper.
2. **Internal behavior**: Sends HTTPS REST requests using `SUPABASE_SERVICE_KEY` authorization headers to manage bucket objects.
3. **Why not AWS S3 SDK?** Supabase provides a unified developer backend ecosystem with minimal configuration overhead.
4. **Limitations**: Adds an external REST HTTP call latency (~50-150ms) during uploads and downloads.
5. **Follow-Up**: "What happens if Supabase Storage is down?" -> Upload and download handlers return HTTP 500 Internal Server Error.

---

## 5. Bcrypt (`golang.org/x/crypto/bcrypt`)
1. **Why was it used here?** Standard for secure password hashing.
2. **Internal behavior**: Uses the Eksblowfish cipher with configurable cost factor (`bcrypt.DefaultCost = 10`), generating a salt internally and embedding it in the output string (`$2a$10$...`).
3. **Why not SHA-256 or MD5?** SHA-256 and MD5 are fast cryptographic hashes that allow hardware GPUs to execute billions of brute-force guesses per second. Bcrypt is intentionally CPU/memory slow.

---

# PART 11 — INTERVIEW QUESTIONS (5 LEVELS)

## LEVEL 1 — Basic Understanding

### Q1: "What does this application do and how is it structured?"
- **Core Points**: File management system with auth, deduplication, quotas, rate limits, and public sharing. Go backend + React TS frontend + Postgres + Supabase Storage.
- **Good Answer**: "Secure File Vault is a Go and React web application that manages user file uploads with content-based deduplication and storage quotas. The backend uses Gorilla Mux for HTTP routing, PostgreSQL for user and file metadata, and Supabase Storage for physical cloud objects. Authentication is handled statelessly via JWT cookies."
- **Follow-up**: "Why separate metadata from file binary storage?" -> Database handles fast structured SQL queries and joins, while object stores efficiently handle large binary payloads.

### Q2: "How does user registration and login work?"
- **Core Points**: Bcrypt password hashing, JWT claims, HTTP-Only cookies.
- **Good Answer**: "On signup, passwords are hashed with Bcrypt before saving to PostgreSQL. On login, `bcrypt.CompareHashAndPassword` verifies credentials. On success, a 5-minute signed JWT is issued in an HTTP-Only, SameSite cookie."
- **Follow-up**: "Why store JWT in a cookie instead of returning it in JSON?" -> Prevents token theft via XSS vulnerabilities.

---

## LEVEL 2 — Implementation & Code

### Q3: "Walk me through how file deduplication is implemented."
- **Core Points**: `mime_type` + `size` lookup -> SHA-256 computation -> candidate match -> DB insert with `is_master = FALSE` + `reference_count++`.
- **Good Answer**: "When a file is uploaded, the backend first queries PostgreSQL for candidate master files with the same MIME type and size. It then streams the uploaded file through SHA-256 hashing. If a candidate matches the calculated hash, the backend inserts a duplicate metadata row pointing to the master's storage path, increments the master's reference count, and skips cloud storage upload."
- **Follow-up**: "What is the time complexity of deduplication?" -> Hash calculation is $O(N)$ where $N$ is file byte length; candidate matching is $O(C)$ where $C$ is candidate count.

### Q4: "How does file deletion handle shared deduplicated files?"
- **Core Points**: Master promotion logic vs. duplicate deletion vs. single reference deletion.
- **Good Answer**: "If a non-master duplicate is deleted, we decrement the master's reference count and delete the DB record without touching cloud storage. If a master file with `reference_count > 1` is deleted, we select an existing duplicate, promote it to `is_master = TRUE`, and delete the old record. Cloud storage is only deleted when `reference_count == 1`."
- **Follow-up**: "What happens if master promotion fails?" -> The current code does not execute inside a transaction, which is a known flaw I would fix with `BEGIN...COMMIT` SQL transactions.

---

## LEVEL 3 — System Design & Trade-offs

### Q5: "How would you optimize the current deduplication algorithm?"
- **Core Points**: Database indexing, SHA-256 pre-calculation on client, presigned URLs.
- **Good Answer**: "Currently, candidates are queried on un-indexed `mime_type` and `size` columns. I would add a composite index `(mime_type, size)` on master files. Additionally, clients could compute the SHA-256 hash locally before uploading to perform instant deduplication checks."
- **Follow-up**: "Can you trust client-computed hashes?" -> No, the backend must verify the hash during background processing before committing the master record.

### Q6: "Why did you use PostgreSQL instead of MongoDB?"
- **Core Points**: Foreign keys, ACID, relational consistency for quota and deduplication.
- **Good Answer**: "PostgreSQL provides relational foreign key cascades (`ON DELETE CASCADE`) and ACID transaction guarantees necessary for atomic quota updates and reference-count decrements. MongoDB would require complex multi-document transaction handling or application-level joins."

---

## LEVEL 4 — Failure, Concurrency & Scaling

### Q7: "What happens if two users upload the exact same file simultaneously?"
- **Core Points**: Race condition, dual master upload, partial unique index fix.
- **Good Answer**: "Currently, both concurrent requests execute `SELECT candidates` simultaneously before either inserts a master record. Both read 0 candidates, both upload physical files to cloud storage, and both insert master records with identical hashes. To fix this, I would add a database-level unique constraint on `hash` where `is_master = TRUE` and handle conflicts gracefully."

### Q8: "How would this system handle 100x traffic scaling?"
- **Core Points**: Redis rate limiting, presigned S3 uploads, read replicas, async queue.
- **Good Answer**: "First, I would decouple file payload streaming from API servers using Presigned S3 URLs so clients upload binary data directly to cloud storage. Second, I would shift rate limiting state to a Redis cluster and deploy multiple stateless Go backend nodes behind a load balancer. Third, I would add DB indexes and read replicas."

---

## LEVEL 5 — Interviewer Probing / Authenticity

### Q9: "What is one architectural weakness in this codebase and how would you fix it?"
- **Core Points**: Non-atomic file deletion, un-indexed DB queries, stateful in-memory rate limiter.
- **Good Answer**: "The biggest weakness is that multi-step operations like master file deletion and upload deduplication are executed outside of database transactions. If the process crashes mid-operation, reference counts or cloud storage objects become out of sync. I would fix this by wrapping DB queries inside `db.Begin()` transactions and adopting soft deletes with background reconciliation workers."

---

# PART 12 — MY INTERVIEW EXPLANATIONS & PITCHES

## A. 30-Second Elevator Pitch
> "Secure File Vault is a multi-tenant backend system written in Go and PostgreSQL with a React frontend. It features content-based file deduplication using SHA-256 hashing, per-user storage quota management, token bucket rate limiting, and role-based access control. Physical files are stored in Supabase Cloud Storage while PostgreSQL manages relational metadata and reference counting."

## B. 2-Minute Technical Summary
> "I built FileVault to explore storage optimization and access control patterns in backend systems. On file upload, the Go backend streams file bytes, validates MIME types using magic-byte sniffing, and computes a SHA-256 content hash. It checks PostgreSQL for candidate files matching the MIME type and size. If a match exists, it increments the master record's reference count and links the new record without writing duplicate cloud objects.
> Authentication is stateless using signed JWT cookies, and rate limiting uses Go's token bucket library per user. For public files, a custom Soft-Auth middleware allows guest access while maintaining user context when available. The whole stack is containerized with Docker Compose."

## C. Whiteboard Architecture Script
1. **Draw Client Box**: "Requests originate from the React SPA carrying HTTP-Only JWT cookies."
2. **Draw Middleware Column**: "Requests pass through CORS, Auth / SoftAuth validation, and per-user Rate Limiting."
3. **Draw Handlers & Services**: "Handlers route business logic: `UploadHandler` calls `dedup.go` for SHA-256 hashing and checks `quota.go`."
4. **Draw PostgreSQL**: "Metadata table `files` maps `user_id` FK to `users` and tracks `hash`, `reference_count`, and `is_master`."
5. **Draw Supabase Object Storage**: "Master file bytes are saved to cloud bucket `file-vault` via SDK."

## D. Biggest Technical Challenge & Personal Learnings
> "The biggest technical challenge was handling file deletion lifecycle for deduplicated files. When a user deletes a file, we can't simply remove the cloud object if other users reference it. Implementing the master promotion logic—where a duplicate is promoted to master when the original master is deleted—taught me the importance of reference counting, transactional isolation, and event-driven cleanup workers."
