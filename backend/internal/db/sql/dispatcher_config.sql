-- Global allowed job_type values for Gemini system prompt and default user category seeding.
CREATE TABLE IF NOT EXISTS dispatcher_config (
    id SMALLINT PRIMARY KEY CHECK (id = 1),
    allowed_job_types JSONB NOT NULL
);

INSERT INTO dispatcher_config (id, allowed_job_types)
VALUES (1, '["echo","email","image_resize","url_scrape","fail"]'::jsonb)
ON CONFLICT (id) DO NOTHING;
