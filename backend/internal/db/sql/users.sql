-- Clerk-backed user rows. id stores the Clerk user id (JWT sub).
-- dispatch_categories is the user's single source of truth for the job_type
-- values the AI dispatcher may pick. New users start with one generic label;
-- the Settings dialog lets them edit the list.
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    email TEXT,
    first_name TEXT,
    last_name TEXT,
    image_url TEXT,
    dispatch_categories JSONB NOT NULL DEFAULT '["general"]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT dispatch_categories_max_len CHECK (
        jsonb_array_length(COALESCE(dispatch_categories, '[]'::jsonb)) <= 10
    )
);

ALTER TABLE users ADD COLUMN IF NOT EXISTS email TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS first_name TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS last_name TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS image_url TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
ALTER TABLE users ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
