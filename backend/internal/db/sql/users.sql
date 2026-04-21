-- Anonymous user rows keyed by client-generated id (e.g. word-word-word).
-- dispatch_categories is the user's single source of truth for the job_type
-- values the AI dispatcher may pick. New users start with one generic label;
-- the Settings dialog lets them edit the list.
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    dispatch_categories JSONB NOT NULL DEFAULT '["general"]'::jsonb,
    CONSTRAINT dispatch_categories_max_len CHECK (
        jsonb_array_length(COALESCE(dispatch_categories, '[]'::jsonb)) <= 10
    )
);
