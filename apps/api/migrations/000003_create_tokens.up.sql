CREATE TABLE tokens (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  token       TEXT NOT NULL,
  session_id  UUID NOT NULL,
  expires_at  TIMESTAMPTZ NOT NULL,
  revoked_at  TIMESTAMPTZ,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_tokens_token ON tokens (token);
CREATE INDEX idx_tokens_session_id ON tokens (session_id);
CREATE INDEX idx_tokens_user_id ON tokens (user_id);