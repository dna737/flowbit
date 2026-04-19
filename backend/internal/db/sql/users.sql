-- Anonymous user rows keyed by client-generated id (e.g. word-word-word).
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    dispatch_categories JSONB NOT NULL DEFAULT '[]'::jsonb,
    CONSTRAINT dispatch_categories_max_len CHECK (
        jsonb_array_length(COALESCE(dispatch_categories, '[]'::jsonb)) <= 10
    )
);
