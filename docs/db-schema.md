[Back to Home Page](../README.md)

# 📦 Database Schema

This document explains the **database schema** for the project, including all tables, relationships, and relevant notes.
The database is **PostgreSQL**, and migrations are managed via `.sql` files in the `db/migrations/` directory.

---

## 🗂️ Tables Overview

- **users** → stores user accounts, roles, and profile information.
- **files** → stores uploaded files metadata, with deduplication support.

Relationship:

- A **user** can upload multiple **files** (`1:N` relationship).
- If a user is deleted, all their files are automatically deleted (`ON DELETE CASCADE`).

---

## 👤 `users` Table

Holds user credentials, profile data, and activity info.

| Column            | Type        | Constraints                 | Description                          |
| ----------------- | ----------- | --------------------------- | ------------------------------------ |
| `id`              | SERIAL      | PRIMARY KEY                 | Unique user ID                       |
| `username`        | TEXT        | UNIQUE, NOT NULL            | Username (used for login)            |
| `email`           | TEXT        | UNIQUE, NOT NULL            | User email                           |
| `password`        | TEXT        | NOT NULL                    | Bcrypt hashed password               |
| `created_at`      | TIMESTAMP   | DEFAULT `CURRENT_TIMESTAMP` | Account creation time                |
| `role`            | VARCHAR(20) | NOT NULL, DEFAULT `'user'`  | Either `user` or `admin`             |
| `last_login`      | TIMESTAMP   | NULLABLE                    | Last login timestamp                 |
| `profile_picture` | TEXT        | NULLABLE                    | File path or URL for profile picture |
| `is_active`       | BOOLEAN     | NOT NULL, DEFAULT `TRUE`    | Marks if user is active              |

---

## 📂 `files` Table

Stores all file metadata and deduplication info.

| Column            | Type      | Constraints                                 | Description                         |
| ----------------- | --------- | ------------------------------------------- | ----------------------------------- |
| `id`              | SERIAL    | PRIMARY KEY                                 | Unique file ID                      |
| `filename`        | TEXT      | NOT NULL                                    | Original filename                   |
| `filepath`        | TEXT      | NOT NULL                                    | Path where file is stored on server |
| `hash`            | TEXT      | NOT NULL                                    | File hash (used for deduplication)  |
| `size`            | BIGINT    | NOT NULL                                    | File size in bytes                  |
| `uploaded_at`     | TIMESTAMP | DEFAULT `CURRENT_TIMESTAMP`                 | When file was uploaded              |
| `user_id`         | INT       | NOT NULL, FK → `users.id` ON DELETE CASCADE | Uploader user                       |
| `reference_count` | BIGINT    | NOT NULL, DEFAULT `1`                       | Number of references to this file   |
| `is_master`       | BOOLEAN   | NOT NULL, DEFAULT `TRUE`                    | True if master/original file        |
| `mime_type`       | TEXT      | NULLABLE                                    | Detected MIME type                  |
| `is_public`       | BOOLEAN   | NOT NULL, DEFAULT `FALSE`                   | Whether file is public              |
| `download_count`  | INT       | NOT NULL, DEFAULT `0`                       | Number of times downloaded          |
| `description`     | TEXT      | NULLABLE                                    | Optional description of file        |

---

## 🔑 Relationships

- **users → files**

  - `users.id` is referenced by `files.user_id`.
  - If a user is deleted, their files are deleted automatically.
  - Relationship type: **One-to-Many**.

- **files (deduplication)**

  - Deduplication is handled by comparing `hash`, `size`, and `mime_type`.
  - Master file → `is_master = TRUE`
  - Duplicates → linked to master file by hash, `is_master = FALSE`, increment `reference_count`.

---

## 🛠️ Migrations Summary

The schema is built through SQL migrations (`db/migrations/`):

1. **`001_init.up.sql`**

   - Creates `users` and `files` tables.
   - Inserts default **admin user** (`root_rachit`).

2. **`002_add_last_login_to_users.up.sql`**

   - Adds `last_login` column to `users`.

3. **`003_add_profile_picture_to_users.up.sql`**

   - Adds `profile_picture` column to `users`.

4. **`004_add_is_active_to_users.up.sql`**

   - Adds `is_active` column to `users`.

5. **`005_add_file_description_to_files.up.sql`**

   - Adds `description` column to `files`.

Each `.down.sql` file drops or removes the corresponding column, allowing rollback.

---

## 👑 Default Admin User

On initialization, a default admin account is created:

```sql
INSERT INTO users (username, email, password, role)
VALUES (
    'root_rachit',
    'rachit@root.com',
    '$2a$10$0kDc0hnX/v6s.X0sV3hSUujJTppCN2l/88sCP/RTFNu2WKEGGt7Iu',
    'admin'
);
```

- **Username**: `root_rachit`
- **Email**: `rachit@root.com`
- **Password**: `rachit` (bcrypt-hashed in DB)
- **Role**: `admin`

---

## 📊 ERD (Entity Relationship Diagram)

```
┌────────────────┐        ┌─────────────────────────┐
│   users        │        │         files           │
│─────────────── │        │──────────────────────── │
│ id (PK)        │◄───────┤ user_id (FK → users.id) │
│ username       │        │ id (PK)                 │
│ email          │        │ filename                │
│ password       │        │ filepath                │
│ created_at     │        │ hash                    │
│ role           │        │ size                    │
│ last_login     │        │ uploaded_at             │
│ profile_picture│        │ reference_count         │
│ is_active      │        │ is_master               │
└────────────────┘        │ mime_type               │
                          │ is_public               │
                          │ download_count          │
                          │ description             │
                          └─────────────────────────┘
```

---

## ✅ Key Notes

- **Deduplication**: Prevents duplicate files being stored; instead, links to master file.
- **Soft deletes**: Users can be deactivated with `is_active`.
- **Public/Private**: Files can be toggled with `is_public`.
- **Download tracking**: Each download increments `download_count`.
- **Extensibility**: Profile pictures and file descriptions are optional fields for extra metadata.

---

[Back to Home Page](../README.md)
