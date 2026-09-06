# 07. Resume & README Claim Verification Matrix

> **Source of Truth Notice**: This document audits claims found in `README.md`, `backend/backend_readme.txt`, and `docs/architecture.md` against actual code implementation. It provides safe phrasing to ensure you present your project accurately without making unverified claims.

---

## Claim Verification Matrix

| Claim in Docs / README | Verified in Code? | Code Evidence | Safe Interview Wording | Dangerous / Unsupported Wording to Avoid |
| :--- | :--- | :--- | :--- | :--- |
| **"Authentication — signup/login with JWT"** | **YES** | `backend/internal/handlers/auth.go`, `services/auth.go`, `middleware/auth.go` | "Implemented JWT authentication stored in HTTP-Only cookies with Bcrypt password hashing." | "Built OAuth2 / SSO multi-provider authentication framework." |
| **"File deduplication — store once, reference multiple times"** | **YES** | `backend/internal/services/dedup.go`, `handlers/files.go:UploadHandler` | "Implemented content-based deduplication using SHA-256 hashing to link duplicate files to a master storage object." | "Built enterprise distributed deduplication with block-level chunking and rolling hash algorithms." |
| **"Storage quotas — per-user storage limit"** | **YES** | `backend/internal/utils/quota.go`, `handlers/files.go:UploadHandler` | "Enforced per-user storage limits by checking total user file sizes against configurable environment variables before accepting uploads." | "Built distributed multi-tier storage quota engine with real-time billing integration." |
| **"Rate limiting — control request bursts"** | **PARTIAL** | `backend/internal/middleware/rateLimit.go` | "Implemented per-user in-memory rate limiting using Go's token bucket library (`x/time/rate`)." | "Built a distributed Redis sliding-window rate limiter." *(It is local in-memory, not distributed!)* |
| **"Public file sharing — share via unique link"** | **YES** | `backend/internal/handlers/public.go`, `handlers/files.go:FileTogglePrivacyHandler` | "Supported public file sharing via public flag toggles and tokenless guest download endpoints." | "Implemented encrypted presigned temporary share URLs with time-to-live expiration." |
| **"Storage statistics — total, deduplicated, savings"** | **YES** | `backend/internal/handlers/admin.go:AdminFilesHandler`, `handlers/files.go:FilesHandler` | "Calculated platform-wide and per-user storage savings by comparing original file sizes against master file sizes." | "Built real-time telemetry streaming pipelines for storage analytics." |
| **"Storage: Local `./uploads` folder"** | **NO** *(Documentation Mismatch)* | `docs/architecture.md` claims local `./uploads`, but code in `internal/storage/supabase.go` uses Supabase Cloud Storage SDK! | "Files are stored in cloud object storage via Supabase Storage SDK." | "Files are stored locally on backend container disk in `./uploads`." *(Code uses Supabase!)* |
| **"MIME type validation — only valid file types allowed"** | **YES** | `backend/internal/utils/mime.go` | "Validated file content against extensions using magic-byte content sniffing (`http.DetectContentType`)." | "Built a virus scanning pipeline for uploaded files." |

---

## Critical Documentation Discrepancy to Explain in Interviews

If an interviewer inspects `docs/architecture.md` and notices it mentions `./uploads` while the code uses `storage.UploadFile` with Supabase Storage:

**How to Explain Truthfully**:
> "In the initial design, the application stored file binaries on the local filesystem inside `./uploads`. Later, I refactored the storage layer to integrate cloud object storage using the Supabase Storage SDK for better persistence in cloud environments. The codebase reflects the updated Supabase implementation, though `docs/architecture.md` retained legacy documentation."
