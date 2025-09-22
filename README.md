[![Review Assignment Due Date](https://classroom.github.com/assets/deadline-readme-button-22041afd0340ce965d47ae6ef1cefeee28c7c493a6346c4f15d667ab976d596c.svg)](https://classroom.github.com/a/2xw7QaEj)

# File Vault — Capstone Internship Task

A secure file storage and sharing application with backend in **Go**, frontend in **React + TypeScript**, and **PostgreSQL** as the database.

---

## ✨ Features

* 🔐 **Authentication** — signup/login with JWT
* 📂 **File upload (multi + drag & drop)**
* 🧮 **File deduplication** — store once, reference multiple times
* 🛡 **MIME type validation** — only valid file types allowed
* 📊 **Storage quotas** — per-user storage limit
* 🚦 **Rate limiting** — control request bursts
* 🔗 **Public file sharing** — share via unique link
* 📉 **Storage statistics** — total, deduplicated, savings

---

## 🛠 Prerequisites

* [Docker](https://docs.docker.com/get-docker/) & Docker Compose (recommended)
* Or (manual mode):

  * Go 1.20+
  * Node.js 18+
  * PostgreSQL 14+

---

## 🚀 Quick Start (Docker)

From the project root:
1) configure .env values 
2) run : 
```bash
docker-compose up --build
```

* Backend: **[http://localhost:8080](http://localhost:8080)**
* Frontend: **[http://localhost:5173](http://localhost:5173)**

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

* Register a new account
* Log in and obtain JWT
* Upload files (single/multi)
* Re-upload same file → deduplication should save storage
* Share file → access via public link
* Check storage stats → used space, dedup savings
* Delete file (only by uploader)

---

## 📌 Notes

This README covers setup, structure, and basic usage.
For deeper details (API endpoints, architecture diagrams, screenshots), please see the **full documentation** in the `docs/` folder.
