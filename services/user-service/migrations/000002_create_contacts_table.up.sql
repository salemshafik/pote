-- Create contacts table for user-to-user contact relationships.
CREATE TABLE IF NOT EXISTS contacts (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id   UUID NOT NULL,
    contact_id UUID NOT NULL,
    nickname   VARCHAR(100) DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_contacts_owner FOREIGN KEY (owner_id) REFERENCES user_profiles(id) ON DELETE CASCADE,
    CONSTRAINT fk_contacts_contact FOREIGN KEY (contact_id) REFERENCES user_profiles(id) ON DELETE CASCADE,
    CONSTRAINT uq_contacts_pair UNIQUE (owner_id, contact_id),
    CONSTRAINT chk_contacts_no_self CHECK (owner_id != contact_id)
);

CREATE INDEX idx_contacts_owner_id ON contacts(owner_id);
