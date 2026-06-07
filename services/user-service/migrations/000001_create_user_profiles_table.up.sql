-- Create user_profiles table for the user-service.
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS user_profiles (
    id           UUID PRIMARY KEY,  -- Same UUID as auth-service users.id
    email        VARCHAR(255) NOT NULL UNIQUE,
    display_name VARCHAR(100) NOT NULL,
    avatar_url   TEXT DEFAULT '',
    bio          TEXT DEFAULT '',
    status       VARCHAR(20) NOT NULL DEFAULT 'offline',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_user_profiles_email ON user_profiles(email);
CREATE INDEX idx_user_profiles_display_name ON user_profiles USING gin(to_tsvector('english', display_name));
