-- ============================
-- Users table
-- ============================
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    role VARCHAR(20) NOT NULL DEFAULT 'user'
);

-- ============================
-- Files table
-- ============================
CREATE TABLE IF NOT EXISTS files (
    id SERIAL PRIMARY KEY,
    filename TEXT NOT NULL,
    filepath TEXT NOT NULL,
    hash TEXT NOT NULL,
    size BIGINT NOT NULL,
    uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reference_count BIGINT NOT NULL DEFAULT 1,
    is_master BOOLEAN NOT NULL DEFAULT TRUE,
    mime_type TEXT,
    is_public BOOLEAN NOT NULL DEFAULT FALSE,
    download_count INT NOT NULL DEFAULT 0
);

-- ============================
-- Default admin user :
-- password = "rachit" (bcrypt hashed)
-- ============================
INSERT INTO users (username, email, password, role)
VALUES (
    'root_rachit',
    'rachit@root.com',
    '$2a$10$0kDc0hnX/v6s.X0sV3hSUujJTppCN2l/88sCP/RTFNu2WKEGGt7Iu',
    -- '$2a$10$CwTycUXWue0Thq9StjUM0uJ8.e5U9hSlJvP1H9xTqG2n5t9e3iHKu',
    'admin'
)
ON CONFLICT (username) DO NOTHING;
