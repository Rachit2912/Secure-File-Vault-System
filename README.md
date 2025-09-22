[![Review Assignment Due Date](https://classroom.github.com/assets/deadline-readme-button-22041afd0340ce965d47ae6ef1cefeee28c7c493a6346c4f15d667ab976d596c.svg)](https://classroom.github.com/a/2xw7QaEj)

---

# File Vault — Capstone Internship Task

A secure file storage and sharing application with backend in **Go**, frontend in **React + TypeScript**, and **PostgreSQL** as the database.

---

## ✨ Features

- 🔐 **Authentication** — signup/login with JWT
- 📂 **File upload (multi + drag & drop)**
- 🧮 **File deduplication** — store once, reference multiple times
- 🛡 **MIME type validation** — only valid file types allowed
- 📊 **Storage quotas** — per-user storage limit
- 🚦 **Rate limiting** — control request bursts
- 🔗 **Public file sharing** — share via unique link
- 📉 **Storage statistics** — total, deduplicated, savings

---

## 🛠 Tech Stack

| Layer          | Technology                 |
| -------------- | -------------------------- |
| **Backend**    | Go 1.20+ with Gorilla Mux  |
| **Frontend**   | React 19, Vite, TypeScript |
| **Database**   | PostgreSQL 15              |
| **Deployment** | Docker & Docker Compose    |

---

## 🛠 Prerequisites

- [Docker](https://docs.docker.com/get-docker/) & Docker Compose (recommended)

Or (manual mode):

- Go 1.20+
- Node.js 18+
- PostgreSQL 14+

---

## 🚀 Quick Start with Docker (Recommended)

From the project root:

1. Configure `.env` values (see `backend/.env.example` if present)
2. Run:

   ```bash
   docker-compose up --build
   ```

- Backend: **[http://localhost:8080](http://localhost:8080)**
- Frontend: **[http://localhost:5173](http://localhost:5173)**

### Default Admin Account

The initial migration creates a default admin for quick testing:

```
username: root_rachit
password: rachit
```

---

## 📂 Project Structure

```
backend/                  → Go backend
  cmd/                    → Application entrypoint
  internal/               → Core code (config, handlers, middleware, services, utils)
  db/                     → Database setup and migrations
  models/                 → Database models
  uploads/                → File storage on disk

frontend/                 → React + TypeScript frontend
  public/                 → Static assets
  src/
    api/                  → API service calls (auth, files, etc.)
    components/           → Reusable UI components (forms, uploads, stats, etc.)
    contexts/             → React contexts (auth, errors)
    pages/                → Main pages (login, signup, dashboard, admin, etc.)
    routes/               → Routing setup (AppRoutes)
    styles/               → Global and component styles
    utils/                → Helper functions
    App.tsx               → Main app component
    main.tsx              → Entry point

docker-compose.yml        → Docker setup for backend + frontend + DB
```

---

## ✅ Usage Checklist

- Register a new account
- Log in and obtain JWT
- Upload files (single/multi)
- Re-upload same file → deduplication should save storage
- Share file → access via public link
- Check storage stats → used space, dedup savings
- Delete file (only by uploader)

---

## 📖 Full Documentation

For deeper details, see the docs folder:

- [Getting Started](docs/getting-started.md)
- [API Endpoints](docs/api/endpoints.md)
- [Postman Collection](docs/api/endpoints.yml)
- [Database Schema](docs/db-schema.md)
- [Architecture & Design](docs/architecture.md)
