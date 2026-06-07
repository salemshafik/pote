-- Create invites table for email invitations.
CREATE TABLE IF NOT EXISTS invites (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    inviter_id UUID NOT NULL,
    email      VARCHAR(255) NOT NULL,
    status     VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL DEFAULT (NOW() + INTERVAL '7 days'),

    CONSTRAINT fk_invites_inviter FOREIGN KEY (inviter_id) REFERENCES user_profiles(id) ON DELETE CASCADE
);

CREATE INDEX idx_invites_inviter_id ON invites(inviter_id);
CREATE INDEX idx_invites_email ON invites(email);
CREATE INDEX idx_invites_status ON invites(status);
