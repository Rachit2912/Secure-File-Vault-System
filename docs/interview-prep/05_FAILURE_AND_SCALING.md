# 05. Concurrency, Failure Analysis & Scaling

> **Source of Truth Notice**: This document combines Part 6 (Concurrency / Failure Analysis) and Part 7 (System Evolution & Scaling). It identifies vulnerabilities in the existing Go/Postgres codebase and presents a step-by-step roadmap to scale the architecture from 1,000 to 1,000,000 users.

---

# PART 6 — CONCURRENCY & FAILURE ANALYSIS

This section analyzes real edge cases and failure scenarios in this codebase, detailing current behavior, risks, detection methods, and production fixes.

---

## Scenario 1: Concurrent Duplicate File Upload Race Condition
- **SCENARIO**: Two users upload the exact same new file (SHA-256 hash `X`) at the exact same millisecond.
- **WHAT CURRENT CODE DOES**: Both requests run `UploadHandler` concurrently. Both execute `SELECT ... WHERE mime_type=$1 AND size=$2 AND is_master=TRUE`. Since no master record exists yet, both queries return 0 candidates. Both calculate SHA-256 hash `X`, both check quota, both call `storage.UploadFile()`, and both execute `INSERT INTO files (..., is_master=TRUE)`.
- **WHAT COULD GO WRONG**: Two master file records with the exact same hash `X` get inserted into PostgreSQL, and duplicate objects are written to Supabase Storage. Deduplication completely fails for concurrent uploads.
- **HOW TO DETECT IT**: Query PostgreSQL for duplicate hashes: `SELECT hash, COUNT(*) FROM files WHERE is_master=TRUE GROUP BY hash HAVING COUNT(*) > 1;`.
- **HOW TO FIX IT IN PRODUCTION**:
  - Add a database-level unique partial index: `CREATE UNIQUE INDEX idx_master_hash ON files(hash) WHERE is_master = TRUE;`.
  - Wrap candidate lookup and insertion inside an isolated SQL transaction (`BEGIN TRANSACTION ISOLATION LEVEL SERIALIZABLE`).
  - Handle DB unique key violation by retrying the duplicate-linking flow.
- **TRADE-OFF**: Serializable transactions and DB locks add minor write latency but guarantee absolute deduplication integrity.

---

## Scenario 2: Concurrent Quota Bypass
- **SCENARIO**: A user with 1MB remaining storage quota initiates 10 parallel HTTP POST upload requests for 9MB files.
- **WHAT CURRENT CODE DOES**: All 10 requests hit `UploadHandler` simultaneously. Each executes `SELECT COALESCE(SUM(size),0) FROM files WHERE user_id=$1`. Because none of the 10 inserts have committed yet, all 10 read `used = 9MB` (quota = 10MB). All 10 requests pass `used + size <= quota` (9MB + 9MB = 18MB? No, 1MB + 9MB = 10MB). All 10 uploads complete!
- **WHAT COULD GO WRONG**: User uploads 90MB of files despite only having a 10MB quota (900% quota overrun).
- **HOW TO DETECT IT**: Compare `SUM(size)` per user against `GetUserQuotaBytes()`: `SELECT user_id, SUM(size) FROM files GROUP BY user_id HAVING SUM(size) > 10485760;`.
- **HOW TO FIX IT IN PRODUCTION**:
  - Maintain a `storage_used_bytes` column in the `users` table.
  - Acquire a row lock prior to checking: `SELECT storage_used_bytes FROM users WHERE id=$1 FOR UPDATE;`.
  - Atomically update: `UPDATE users SET storage_used_bytes = storage_used_bytes + $1 WHERE id=$2;`.
- **TRADE-OFF**: Row locking serializes file uploads per user (one upload at a time per user account).

---

## Scenario 3: Database Crash Mid-Upload Operation
- **SCENARIO**: A new master file is successfully uploaded to Supabase Storage, but the database crashes or disconnects right before `INSERT INTO files` executes.
- **WHAT CURRENT CODE DOES**: In `UploadHandler`, `storage.UploadFile()` succeeds. Next, `db.DB.Exec("INSERT INTO files...")` fails with a connection error. The handler enters the error branch `if err != nil` and executes `storage.DeleteFile(storagePath)`.
- **WHAT COULD GO WRONG**: If the application server or network crashes *during* the compensation call `storage.DeleteFile()`, an orphan object remains in Supabase Storage forever, accumulating cloud storage costs.
- **HOW TO DETECT IT**: Run a nightly reconciliation cron job comparing storage bucket object paths against `filepath` strings in PostgreSQL.
- **HOW TO FIX IT IN PRODUCTION**:
  - Use an asynchronous Outbox Pattern or background cleanup worker.
  - Mark file records as `status = 'pending'` in DB before storage upload, then update to `'completed'` upon success. Orphan pending files older than 1 hour are cleaned up by a worker.
- **TRADE-OFF**: Requires a background task runner (e.g., Redis queue / Celery / Asynq).

---

## Scenario 4: Partial Failure During Master Deletion
- **SCENARIO**: User deletes a master file (`is_master = TRUE`) with `reference_count = 1`.
- **WHAT CURRENT CODE DOES**: `FileDeleteHandler` executes `DELETE FROM files WHERE id=$1` FIRST. If DB deletion succeeds, it calls `storage.DeleteFile(filepathOnDisk)`.
- **WHAT COULD GO WRONG**: If Supabase Storage is temporarily down or returns a network error, the DB record is ALREADY deleted! The user receives an error response or success response, but the file object remains in cloud storage.
- **HOW TO DETECT IT**: Server log prints: `Warning: failed to delete storage object <path>: <err>`.
- **HOW TO FIX IT IN PRODUCTION**:
  - Soft-delete DB records (`deleted_at = NOW()`).
  - Use an asynchronous deletion queue to delete cloud storage objects with retries and exponential backoff.
- **TRADE-OFF**: Storage deletion is non-instantaneous (eventual consistency).

---

## Scenario 5: Rate Limiter State Loss on App Restart / Horizontal Scaling
- **SCENARIO**: Server restarts or multiple backend nodes run behind a round-robin load balancer.
- **WHAT CURRENT CODE DOES**: `RateLimitMiddleware` keeps rate limiters in an in-memory Go map `limiters = make(map[int]*userLimiter)`.
- **WHAT COULD GO WRONG**: On server restart, all rate limit buckets reset to full capacity. In a multi-node setup, an attacker can send N times the allowed rate limit by fanning out requests across N backend instances.
- **HOW TO DETECT IT**: High traffic surges hitting DB/storage despite rate limiting enabled.
- **HOW TO FIX IT IN PRODUCTION**: Shift rate limit state store to a centralized Redis cluster using a sliding-window algorithm executed via Lua scripts.
- **TRADE-OFF**: Introduces Redis network hop latency (~1-2ms) on every API request.

---

# PART 7 — SCALE THIS SYSTEM (100x TO 1000x TRAFFIC)

Suppose an interviewer asks: *"How would you scale this system from handling 1,000 files/day to 1,000,000 files/day?"*

Here is the step-by-step architectural evolution, addressing bottlenecks **in order of occurrence**.

---

## Step 1: Bottleneck — In-Memory State & Single-Instance Backend
- **CURRENT BOTTLENECK**: The Go server is stateful (in-memory rate limiting map `limiters`) and runs as a single instance.
- **WHY CHANGE**: A single server process cannot handle 10,000+ concurrent requests and fails completely if the process crashes.
- **PROPOSED EVOLUTION**:
  1. Move rate limiting state to a **Redis** cluster.
  2. Deploy multiple stateless Go backend instances behind an Application Load Balancer (Nginx / AWS ALB).
- **WHEN TO ADD**: At >5,000 active concurrent users or when running >1 server instance.
- **WHEN NOT TO ADD**: Low traffic single-server deployment.
- **TRADE-OFF**: Adds Redis infrastructure overhead and network hop latency.

---

## Step 2: Bottleneck — Synchronous File Payload Streaming Through API Server
- **CURRENT BOTTLENECK**: File bytes stream through the Go API server (`UploadHandler` reads multipart form, computes SHA-256 synchronously, and uploads bytes to Supabase).
- **WHY CHANGE**: Passing large binary streams through backend API memory saturates network interfaces, consumes Go heap memory, and ties up HTTP worker goroutines for long durations.
- **PROPOSED EVOLUTION**: **Presigned S3 / Cloud Storage Direct Uploads**
  1. Client sends file metadata (`filename`, `size`, `client_hash`) to `POST /api/files/presign`.
  2. Backend validates quota and checks DB for deduplication.
     - If duplicate exists -> Returns `{"status": "linked"}` instantly!
     - If new file -> Backend generates a Presigned S3 Upload URL and returns it to client.
  3. Client uploads file bytes directly to S3 bucket from browser.
  4. S3 fires an event notification (AWS SQS / EventBridge) to a worker service that verifies upload completion and updates PostgreSQL.
- **WHEN TO ADD**: When file uploads exceed 100 GB/day or media/video files >20MB are supported.
- **WHEN NOT TO ADD**: Small file uploads (<1MB) where presigning overhead exceeds direct upload.
- **TRADE-OFF**: Frontend must handle multi-step upload logic (Presign -> S3 Upload -> Confirm).

---

## Step 3: Bottleneck — Un-Indexed Database Queries & Full Table Scans
- **CURRENT BOTTLENECK**: PostgreSQL lacks indexes on `mime_type`, `size`, `is_master`, `hash`, `user_id`, and `is_public`.
- **WHY CHANGE**: As the `files` table grows beyond 100,000 rows, `SELECT ... WHERE mime_type=$1 AND size=$2 AND is_master=TRUE` will require scanning millions of rows on disk, causing query timeouts.
- **PROPOSED EVOLUTION**:
  1. Add composite index: `CREATE INDEX idx_files_dedup ON files(mime_type, size) WHERE is_master = TRUE;`.
  2. Add foreign key index: `CREATE INDEX idx_files_user ON files(user_id);`.
  3. Add Redis caching layer for read-heavy public file listings (`GET /api/publicFiles`).
  4. Implement database read replicas (Primary for writes, Read Replicas for `GET` queries).
- **WHEN TO ADD**: Immediately (DB indexes) and at >100,000 DB rows (Caching & Read Replicas).
- **TRADE-OFF**: Indexes slightly slow down write speed (`INSERT`) and consume additional database disk space.

---

## Step 4: Bottleneck — Synchronous Heavy Computation (SHA-256 Hashing)
- **CURRENT BOTTLENECK**: The backend server computes SHA-256 hashes synchronously during the HTTP request lifecycle.
- **WHY CHANGE**: Computing SHA-256 for a 500MB file takes several CPU seconds, blocking HTTP request handling.
- **PROPOSED EVOLUTION**: **Asynchronous Background Processing Queue**
  1. Client uploads file via presigned URL.
  2. Background worker pool (using Redis / NATS / Asynq) picks up hashing tasks asynchronously.
  3. Worker updates deduplication indexes and links records post-upload.
- **WHEN TO ADD**: High-volume uploads with file sizes >50MB.
- **TRADE-OFF**: File deduplication becomes eventually consistent rather than strictly synchronous.

---

## Architecture Evolution Diagram (Target Production State)

```
                              +-------------------------+
                              |   Client (Browser SPA)  |
                              +-------------------------+
                                 /                   \
                 1. Request Presign                2. Direct Upload
                 & Metadata Check                  Binary Payload
                               /                       \
                              v                         v
              +-----------------------+     +-----------------------+
              | Load Balancer (ALB)   |     | Cloud Object Storage  |
              +-----------------------+     |  (AWS S3 / Supabase)  |
               /          |          \      +-----------------------+
              v           v           v                 |
          +-------+   +-------+   +-------+             | 3. Object Created
          | Go    |   | Go    |   | Go    |             |    Event Notification
          | Node  |   | Node  |   | Node  |             v
          +-------+   +-------+   +-------+     +-----------------------+
              |           |           |         | AWS SQS / NATS Queue  |
              +-----------+-----------+         +-----------------------+
                          |                                 |
                          v                                 v
              +-----------------------+         +-----------------------+
              |     Redis Cluster     |         |  Background Workers   |
              | (Rate Limit & Cache)  |         |  (Async Dedup Check)  |
              +-----------------------+         +-----------------------+
                          |                                 |
                          +----------------+----------------+
                                           |
                                           v
                              +-------------------------+
                              | PostgreSQL (Primary)    |
                              | + Read Replicas         |
                              +-------------------------+
```
