# 08. Final Interview Revision Cheat Sheet

> **10-15 Minute Interview Warm-Up Guide**: Read this document right before your interview to refresh architecture, key flows, technology justifications, failure scenarios, and code locations.

---

## 0. WHITEBOARD HIGH-LEVEL DESIGN (HLD) DIAGRAM

Draw this on a whiteboard during your interview when asked: *"Can you draw the system architecture?"*

### Mermaid Graph Code
```mermaid
graph TD
    Client[Client Browser / React SPA] -->|HTTP Requests + JWT Cookie| Router[Gorilla Mux Router]

    subgraph Go Backend Server
        Router --> CORS[CORS Middleware]
        CORS --> Auth[Auth / SoftAuth Middleware]
        Auth --> RateLimit[Rate Limit Middleware]

        RateLimit --> Handlers[API Handlers]
        Handlers --> Dedup[Dedup Service: SHA-256]
        Handlers --> Quota[Quota Utils]
        Handlers --> StorageClient[Storage Wrapper]
    end

    Handlers -->|SQL Queries ($1, $2)| Postgres[(PostgreSQL 15 Database)]
    StorageClient -->|HTTPS REST SDK| Supabase[(Supabase Storage Bucket)]

    classDef primary fill:#2563eb,color:#fff,stroke:#1d4ed8;
    classDef secondary fill:#059669,color:#fff,stroke:#047857;
    classDef db fill:#d97706,color:#fff,stroke:#b45309;
    class Client primary;
    class Router,CORS,Auth,RateLimit,Handlers,Dedup,Quota,StorageClient secondary;
    class Postgres,Supabase db;
```

### Simple Whiteboard ASCII Diagram (How to draw step-by-step on whiteboard)

```
STEP 1: Draw Client (React SPA)
STEP 2: Draw Go API Server Box containing Middleware -> Handlers -> Services
STEP 3: Draw Database (PostgreSQL) and Object Storage (Supabase) at the bottom

+-------------------------------------------------------------------------------+
|                            CLIENT (React SPA / Browser)                       |
+-------------------------------------------------------------------------------+
                                        |
                          HTTP / REST Requests
                      (Cookies: token = JWT)
                                        |
                                        v
+-------------------------------------------------------------------------------+
|                            GO BACKEND API SERVER                              |
|                                                                               |
|  +-------------------------------------------------------------------------+  |
|  |                           MIDDLEWARE PIPELINE                           |  |
|  |  CORS Middleware -> Auth / SoftAuth Middleware -> RateLimit Middleware  |  |
|  +-------------------------------------------------------------------------+  |
|                                       |                                       |
|                                       v                                       |
|  +-------------------------------------------------------------------------+  |
|  |                            HANDLERS LAYER                               |  |
|  |  auth.go | files.go (Upload/Download/Delete) | admin.go | public.go      |  |
|  +-------------------------------------------------------------------------+  |
|                                 /           \                                 |
|                                /             \                                |
|                               v               v                               |
|  +---------------------------------+     +---------------------------------+  |
|  |        SERVICES & UTILS         |     |         STORAGE CLIENT          |  |
|  |  dedup.go (SHA-256 Hash Engine)   |     |  storage/files.go               |  |
|  |  quota.go & mime.go             |     |  Supabase Storage SDK Client    |  |
|  +---------------------------------+     +---------------------------------+  |
+-------------------------------------------------------------------------------+
                 |                                           |
           SQL Queries                                Cloud Storage API
        ($1, $2 Parameters)                             (HTTPS REST)
                 |                                           |
                 v                                           v
+-----------------------------------+     +-------------------------------------+
|      POSTGRESQL 15 DATABASE       |     |       SUPABASE STORAGE BUCKET       |
|  Tables: users, files             |     |       Bucket: 'file-vault'          |
|  - Relational metadata & FKs      |     |       - Physical master binary      |
|  - Reference count tracking       |     |         file objects                |
+-----------------------------------+     +-------------------------------------+
```

---

## 1. High-Level Architecture Summary
- **Stack**: Go 1.20+ (Gorilla Mux) + PostgreSQL 15 + Supabase Cloud Storage + React 19 SPA + Docker Compose.
- **Auth**: Stateless JWT in HTTP-Only, SameSite cookies (5-min TTL). Passwords hashed with Bcrypt.
- **Storage Strategy**: Metadata in Postgres (`files` table); physical binary objects in Supabase bucket `file-vault`.
- **Key Features**: Content-based SHA-256 deduplication, storage quota enforcement, magic-byte MIME validation, per-user rate limiting, soft-auth public sharing, admin management.

---

## 2. 5 Key Request Flows (Quick Mental Map)

1. **Signup / Login**: `POST /api/signup` (Bcrypt hash) -> `POST /api/login` (Bcrypt compare -> Issue signed JWT cookie).
2. **File Upload**: `POST /api/upload` -> Auth -> RateLimit -> Parse multipart -> Magic-byte MIME check -> Query master candidates by `(mime_type, size)` -> SHA-256 Hash ->
   - *If Match*: `INSERT (is_master=FALSE)`, `UPDATE ref_count++` (skip storage write).
   - *If New*: Check quota (`SUM(size) <= 10MB`) -> Upload to Supabase -> `INSERT (is_master=TRUE, ref_count=1)`.
3. **Download**: `GET /api/fileDownload/{id}` -> SoftAuth -> Verify privacy (`is_public` or `userID == owner`) -> Stream bytes from Supabase -> Increment `download_count`.
4. **Delete**: `GET /api/fileDelete/{id}` -> Auth -> If master with `ref_count > 1`, promote a duplicate to master; if `ref_count == 1`, delete DB row and delete cloud storage object.
5. **Admin List**: `GET /api/adminFiles` -> Auth (Role == admin) -> Dynamic parameterized SQL filter -> Calculate savings (`originalSize - dedupSize`).

---

## 3. Database Schema at a Glance

- **`users`**: `id (PK)`, `username (UQ)`, `email (UQ)`, `password`, `role`, `created_at`, `last_login`, `profile_picture`, `is_active`.
- **`files`**: `id (PK)`, `filename`, `filepath`, `hash`, `size`, `uploaded_at`, `user_id (FK -> users.id ON DELETE CASCADE)`, `reference_count`, `is_master`, `mime_type`, `is_public`, `download_count`, `description`.
- **Schema Weakness**: Missing secondary database indexes; master deletion executed outside explicit SQL transactions.

---

## 4. Key Technologies & Justification

| Tech | Why Chosen? | Interview Key Point |
| :--- | :--- | :--- |
| **Go** | High concurrency via goroutines | Lightweight M:N scheduling, single compiled binary |
| **Gorilla Mux** | URL path parameters (`{id}`) | Native Go `http.Handler` interface compliance |
| **PostgreSQL** | Relational integrity & ACID | Native FK cascade, atomic query capability |
| **Supabase Storage** | Cloud object storage API | Decouples binary storage from API server memory |
| **Bcrypt** | Password hashing | Intentionally CPU-slow cipher resisting GPU brute-force |
| **x/time/rate** | Per-user rate limiting | Token bucket algorithm in middleware |

---

## 5. Top 5 Concurrency & Failure Scenarios to Quote

1. **Concurrent Upload Race**: Two identical uploads at the same millisecond both insert master records. *(Fix: Partial unique index on master hash).*
2. **Parallel Quota Bypass**: Concurrent uploads read uncommitted storage totals. *(Fix: SELECT FOR UPDATE on user storage row).*
3. **Database Crash Mid-Upload**: Supabase write succeeds but DB insert fails. *(Fix: Outbox pattern / async status tracking).*
4. **Master Delete Failure**: DB row deleted before Supabase delete fails. *(Fix: Soft-delete + background deletion queue).*
5. **Rate Limiter Restart Reset**: In-memory map clears on restart. *(Fix: Shift state to Redis cluster).*

---

## 6. Top 5 Interview Questions & Core Responses

1. **"Why SHA-256 for deduplication?"** -> Collision-resistant hash ($1 \text{ in } 2^{256}$ probability). We match MIME and size first to minimize candidate lookups.
2. **"Why store JWT in cookies instead of localStorage?"** -> HTTP-Only cookies prevent token theft via XSS attacks.
3. **"How does master promotion work on delete?"** -> If master `ref_count > 1`, we select a non-master duplicate, promote it to `is_master = TRUE`, and delete the user's row without deleting the cloud object.
4. **"How would you scale this to 100x traffic?"** -> Presigned S3 direct uploads, Redis rate limiting, DB indexes & read replicas, and async worker queues.
5. **"What is one design decision you'd change today?"** -> Normalize schema into separate `storage_objects` and `user_files` tables to simplify reference counting and avoid master promotion logic.

---

## 7. Important Code Locations

- Server Bootstrap & Routes: `backend/cmd/server/main.go`
- Authentication & Handlers: `backend/internal/handlers/auth.go`
- Upload & Deduplication Logic: `backend/internal/handlers/files.go:UploadHandler`
- Delete & Master Promotion: `backend/internal/handlers/files.go:FileDeleteHandler`
- SHA-256 Hashing Service: `backend/internal/services/dedup.go`
- Supabase Storage Client: `backend/internal/storage/files.go`
- Rate Limiting Middleware: `backend/internal/middleware/rateLimit.go`
- SQL Schema Migrations: `backend/internal/db/migrations/`
