[Back to Home Page](../README.md)

# Getting Started

This guide explains how to set up and run the File Vault project locally for development and testing.

---

## 📦 Prerequisites

- Docker & Docker Compose (recommended)
- OR:

  - Go 1.20+
  - Node.js 18+
  - PostgreSQL 14+

- `psql` CLI for manual DB setup
- (Optional) `migrate` CLI for migrations

---

## 🚀 Quick Start with Docker (Recommended)

From the project root:

```bash
docker-compose up --build
```

Services:

- Backend → [http://localhost:8080](http://localhost:8080)
- Frontend → [http://localhost:5173](http://localhost:5173)

---

## ⚡ Manual Setup (without Docker)

### 1. Database

Create database:

```bash
createdb filevault
```

Run migrations:

```bash
psql -U postgres -d filevault -f backend/db/migrations/001_init.up.sql
```

### 2. Backend

```bash
cd backend
cp .env.example .env   # edit .env values if needed
go mod tidy
go run ./cmd/server
```

### 3. Frontend

```bash
cd frontend
npm install
npm run dev
```

---

## ✅ Quick Verification

1. **Signup** → create a new user account
2. **Login** → obtain JWT token
3. **Upload file** → verify deduplication (re-upload same file, storage shouldn’t increase)
4. **List files** → see uploaded metadata
5. **Public share** → open public link and download

---

## 🛠 Troubleshooting

- If backend cannot connect to DB, check `.env` values and database is running.
- If frontend cannot reach backend, confirm `VITE_API_URL` matches backend URL.
- If migrations fail, recheck SQL files in `backend/db/migrations/`.
