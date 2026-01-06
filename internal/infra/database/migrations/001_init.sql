-- +migrate Up

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id uuid PRIMARY KEY,
    name VARCHAR(155) NOT NULL,
    email VARCHAR(155) NOT NULL UNIQUE,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'ACTIVE', 'BLOCKED')),
    password_hash VARCHAR(255) NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME,
    email_confirmed_at DATETIME,
    blocked_at DATETIME
);

CREATE INDEX idx_users_email ON users(email);

-- Sessions table
CREATE TABLE IF NOT EXISTS sessions (
    id uuid PRIMARY KEY,
    token VARCHAR(255) NOT NULL UNIQUE,
    device_name VARCHAR(255) NOT NULL,
    ip_address VARCHAR(45) NOT NULL,
    user_agent TEXT NOT NULL,
    expires_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL,
    user_id uuid NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX idx_sessions_token ON sessions(token);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);

-- Verifications table
CREATE TABLE IF NOT EXISTS verifications (
    id uuid PRIMARY KEY,
    flow VARCHAR(20) NOT NULL CHECK (flow IN ('RESET_PASSWORD', 'VERIFICATION_EMAIL', 'CHANGE_EMAIL')),
    token VARCHAR(255) NOT NULL UNIQUE,
    created_at DATETIME NOT NULL,
    expires_at DATETIME NOT NULL,
    payload TEXT,
    user_id uuid NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX idx_verifications_token ON verifications(token);
CREATE INDEX idx_verifications_expires_at ON verifications(expires_at);

-- +migrate Down
DROP INDEX IF EXISTS idx_verifications_expires_at;
DROP INDEX IF EXISTS idx_verifications_token;
DROP TABLE IF EXISTS verifications;

DROP INDEX IF EXISTS idx_sessions_expires_at;
DROP INDEX IF EXISTS idx_sessions_token;
DROP TABLE IF EXISTS sessions;

DROP INDEX IF EXISTS idx_users_email;
DROP TABLE IF EXISTS users;
