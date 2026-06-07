-- Create chat_members table for chat membership and per-chat RBAC.
CREATE TABLE IF NOT EXISTS chat_members (
    id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chat_id   UUID NOT NULL,
    user_id   UUID NOT NULL, -- auth user UUID
    role      VARCHAR(20) NOT NULL DEFAULT 'member',
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_chat_members_chat FOREIGN KEY (chat_id) REFERENCES chats(id) ON DELETE CASCADE,
    CONSTRAINT uq_chat_members UNIQUE (chat_id, user_id),
    CONSTRAINT chk_chat_members_role CHECK (role IN ('owner', 'admin', 'member'))
);

CREATE INDEX idx_chat_members_user_id ON chat_members(user_id);
CREATE INDEX idx_chat_members_chat_id ON chat_members(chat_id);
