-- Create chats table for the chat-service.
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS chats (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type       VARCHAR(20) NOT NULL DEFAULT 'group',
    name       VARCHAR(150) NOT NULL DEFAULT '',
    created_by UUID NOT NULL, -- auth user UUID of the creator
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_chats_type CHECK (type IN ('direct', 'group'))
);

CREATE INDEX idx_chats_created_by ON chats(created_by);
