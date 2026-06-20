CREATE TABLE users (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name              TEXT NOT NULL,
  email             TEXT NOT NULL,
  password_hash     TEXT NOT NULL,
  role              TEXT NOT NULL DEFAULT 'member' CHECK (role IN ('superadmin', 'admin', 'staff', 'member')),
  image_url         TEXT,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);