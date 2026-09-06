CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    current_token_hash TEXT NOT NULL,
    previous_token_hash TEXT,
    previous_rotated_at TIMESTAMPTZ,
    user_agent TEXT,
    ip_address TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ
);

CREATE INDEX idx_sessions_user_id_active ON sessions(user_id) WHERE revoked_at IS NULL;
CREATE INDEX idx_sessions_current_token ON sessions(current_token_hash);

DROP TABLE IF EXISTS refresh_tokens;
