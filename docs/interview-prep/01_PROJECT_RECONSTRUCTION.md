# 01. Project Reconstruction

> **Source of Truth Notice**: This document is generated purely by reverse-engineering the source code of this repository (`backend/`, `frontend/`, `docker-compose.yml`, SQL migrations). No non-existent features or fake metrics have been created.

---

## A. What Does This System Actually Do?

**Secure File Vault** is a multi-tenant file management web application. It allows users to register, log in, upload files, manage file privacy settings (public vs. private), download files, and view storage statistics.

Key capabilities verified directly in code:
1. **JWT-Based Authentication**: Registration (`POST /api/signup`), Login (`POST /api/login`), and Cookie-based session validation (`backend/internal/middleware/auth.go`).
2. **Content-Based File Deduplication**: Before storing new file bytes, the backend calculates the SHA-256 hash of uploaded content. If a master file with the same MIME type, size, and hash exists in PostgreSQL, the backend increments `reference_count` and links the new file record to the existing object without duplicating storage (`backend/internal/handlers/files.go:UploadHandler`).
3. **MIME Type Validation**: Prevents file extension spoofing by inspecting the file header with `http.DetectContentType` (`backend/internal/utils/mime.go`).
4. **Storage Quota Enforcement**: Limits total storage per user based on configured `USER_QUOTA_MB` (`backend/internal/utils/quota.go`).
5. **Per-User In-Memory Rate Limiting**: Caps API request rates per user ID using `golang.org/x/time/rate` (`backend/internal/middleware/rateLimit.go`).
6. **Public vs. Private Sharing & Soft-Auth**: Users can toggle file privacy (`GET /api/fileTogglePrivacy/{id}`). Public files can be downloaded by anyone (`SoftAuthMiddleware` allows guest downloads), while private files require JWT verification of owner identity (`backend/internal/handlers/files.go:FileDownloadHandler`).
7. **Admin Privileges**: Users with role `admin` can view all files across the platform (`GET /api/adminFiles`) and promote/demote user roles (`POST /api/makeAdmin`, `POST /api/makeUser`).
8. **Cloud Object Storage Integration**: Uses Supabase Storage SDK (`github.com/supabase-community/storage-go`) via Service Key to store master file objects in bucket `file-vault` (`backend/internal/storage/supabase.go`).

---

## B. What Problem Was It Trying To Solve?

1. **Storage Storage Cost & Redundancy**: If multiple users upload identical files (e.g., common PDFs, images, dataset downloads), storing separate physical copies wastes disk/cloud storage space. FileVault solves this by storing the physical file **once** (master file) while creating separate metadata pointers for each user.
2. **Access Control & Secure Sharing**: File storage services need granular permission boundaries. FileVault enforces user ownership for file deletion and private access, while enabling tokenless/guest downloads for explicitly shared public files.
3. **Resource Protection & Abuse Prevention**: Unrestricted file uploads can lead to Denial of Service (DoS) or disk exhaustion. FileVault enforces per-user storage quotas and per-user request rate limits.

---

## C. Main Technologies Used & WHERE Each Is Used

| Technology | Role / Purpose | Exact Code Location |
| :--- | :--- | :--- |
| **Go (1.20+)** | Backend REST API server | `backend/cmd/server/main.go`, `backend/internal/` |
| **Gorilla Mux (`github.com/gorilla/mux`)** | HTTP router & path variable extraction | `backend/cmd/server/main.go`, `backend/internal/handlers/` |
| **PostgreSQL 15 (`github.com/lib/pq`)** | Relational database storing users and file metadata | `backend/internal/db/db.go`, `backend/internal/db/migrations/` |
| **Supabase Storage SDK (`github.com/supabase-community/storage-go`)** | Remote cloud object storage bucket client | `backend/internal/storage/supabase.go` |
| **Bcrypt (`golang.org/x/crypto/bcrypt`)** | Password hashing during signup & verification on login | `backend/internal/handlers/auth.go:SignupHandler`, `LoginHandler` |
| **JWT (`github.com/golang-jwt/jwt/v5`)** | Token generation and signature verification for auth | `backend/internal/services/auth.go`, `backend/internal/middleware/auth.go` |
| **Rate Limiter (`golang.org/x/time/rate`)** | In-memory token bucket rate limiter | `backend/internal/middleware/rateLimit.go` |
| **Dotenv (`github.com/joho/godotenv`)** | Environment variable loader from `.env` file | `backend/internal/config/config.go` |
| **React 19 + TypeScript + Vite** | Frontend single-page application (SPA) | `frontend/src/` |
| **Docker & Docker Compose** | Containerized deployment orchestration | `docker-compose.yml`, `backend/Dockerfile`, `frontend/Dockerfile` |

---

## D. Repository Structure

```
.
├── docker-compose.yml              # Orchestrates db (Postgres 15), backend (Go), frontend (Vite)
├── README.md                       # Project summary & user guide
├── backend/
│   ├── cmd/server/main.go          # Application entrypoint: loads config, storage, DB, sets up router & starts HTTP server
│   ├── internal/
│   │   ├── config/config.go        # Configuration loader using godotenv & env fallbacks
│   │   ├── db/
│   │   │   ├── db.go               # PostgreSQL database connection pool setup (SetMaxOpenConns, SetMaxIdleConns)
│   │   │   └── migrations/         # Plain SQL schema migrations (001_init to 005_add_file_description)
│   │   ├── handlers/               # HTTP Handlers (controllers)
│   │   │   ├── admin.go            # AdminFilesHandler, MakeAdminHandler, MakeUserHandler
│   │   │   ├── auth.go             # SignupHandler, LoginHandler, LogoutHandler, RefershHandler
│   │   │   ├── files.go            # FilesHandler, UploadHandler, FileDeleteHandler, FileDownloadHandler, FileTogglePrivacyHandler, FileDetailHandler
│   │   │   └── public.go           # PublicFilesHandler
│   │   ├── middleware/             # HTTP Middleware
│   │   │   ├── auth.go             # AuthMiddleware (strict JWT requirement)
│   │   │   ├── cors.go             # CORS header handling & preflight response
│   │   │   ├── rateLimit.go        # RateLimitMiddleware (per-user rate limiting)
│   │   │   └── soft_auth.go        # SoftAuthMiddleware (optional JWT parsing for public endpoints)
│   │   ├── models/                 # Data structs & DB helper methods
│   │   │   ├── claims.go           # JWT Claims struct
│   │   │   ├── file.go             # File model struct
│   │   │   └── user.go             # User model struct & GetUserByID
│   │   ├── services/               # Business logic
│   │   │   ├── auth.go             # GenerateJWT function
│   │   │   ├── dedup.go            # ComputeHash (SHA-256) & FindDuplicate algorithms
│   │   │   └── files.go            # GetFileByID SQL query service
│   │   ├── storage/                # Storage client wrapper
│   │   │   ├── files.go            # UploadFile, DownloadFile, DeleteFile helpers
│   │   │   └── supabase.go         # Supabase client initialization (`storage.Init`)
│   │   └── utils/                  # Helper utilities
│   │       ├── format.go           # FormatBytes human-readable string converter
│   │       ├── mime.go             # ValidateMIME type verification
│   │       └── quota.go            # GetUserQuotaBytes helper
│   ├── Dockerfile                  # Go multi-stage Docker build
│   └── go.mod / go.sum             # Go module dependencies
└── frontend/                       # React 19 SPA frontend
    ├── src/
    │   ├── api/                    # API wrappers (fetch client using API_BASE)
    │   ├── components/             # Reusable React components (FileUpload, FileList, StorageStats, Filters)
    │   ├── contexts/               # React Contexts (AuthContext, ErrorContext)
    │   ├── pages/                  # Page components (Dashboard, FileDetail, AdminDashboard, LoginPage, SignupPage)
    │   └── routes/                 # Routing definition (AppRoutes)
    ├── package.json                # Frontend dependencies
    └── vite.config.ts              # Vite configuration
```

---

## Architecture Diagram

```
+-----------------------------------------------------------------------------------+
|                                 CLIENT (Browser / React)                          |
+-----------------------------------------------------------------------------------+
                                          |
                                    HTTP Requests
                                (Cookies: token=JWT)
                                          |
                                          v
+-----------------------------------------------------------------------------------+
|                                 GO BACKEND SERVER                                 |
|                                                                                   |
|  +-----------------------------------------------------------------------------+  |
|  |                             MIDDLEWARE LAYER                                |  |
|  |  CORS Middleware -> Auth / SoftAuth Middleware -> RateLimit Middleware      |  |
|  +-----------------------------------------------------------------------------+  |
|                                         |                                         |
|                                         v                                         |
|  +-----------------------------------------------------------------------------+  |
|  |                             HANDLERS / CONTROLLERS                          |  |
|  |  auth.go, files.go, admin.go, public.go                                     |  |
|  +-----------------------------------------------------------------------------+  |
|                                   /           \                                   |
|                                  /             \                                  |
|                                 v               v                                 |
|  +-----------------------------------+     +-----------------------------------+  |
|  |          SERVICES & UTILS         |     |          STORAGE CLIENT           |  |
|  |  dedup.go (SHA-256), mime.go,     |     |  storage/files.go                 |  |
|  |  quota.go, auth.go (JWT)          |     |  Supabase Storage SDK Client      |  |
|  +-----------------------------------+     +-----------------------------------+  |
+-----------------------------------------------------------------------------------+
                   |                                           |
             SQL Queries                                Cloud Storage API
             ($1, $2 params)                              (HTTPS REST)
                   |                                           |
                   v                                           v
+-------------------------------------+     +---------------------------------------+
|        POSTGRESQL 15 DATABASE       |     |        SUPABASE STORAGE BUCKET        |
|  Tables: users, files               |     |        Bucket: file-vault             |
+-------------------------------------+     +---------------------------------------+
```
