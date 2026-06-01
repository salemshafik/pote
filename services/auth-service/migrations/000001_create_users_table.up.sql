-- Create users table for the auth-service.
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS users (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email       VARCHAR(255) NOT NULL UNIQUE,
    display_name VARCHAR(100) NOT NULL,
    password_hash TEXT,                          -- NULL for OAuth-only users
    avatar_url  TEXT DEFAULT '',
    provider    VARCHAR(20) NOT NULL DEFAULT 'email', -- 'email' or 'google'
    provider_id VARCHAR(255) DEFAULT '',         -- External OAuth provider user ID
    role        VARCHAR(20) NOT NULL DEFAULT 'user',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_provider ON users(provider, provider_id);
