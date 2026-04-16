# Scheduler (core)

See also: [Architecture](10-architecture.md) · [Stack](20-stack-and-deployment.md) · [AI dispatcher](40-ai-dispatcher.md) · [Build checklist](BUILD-CHECKLIST.md)

---

## Block 2 — Core scheduler

- Go API server with `POST /jobs` and `GET /jobs/:id`
- Kafka producer in the API server
- Worker that consumes from Kafka and updates job status in PostgreSQL
- Test end-to-end with a hardcoded dummy job type (no AI)

---

## Block 3 — Retry + dead-letter queue

- Failed jobs re-publish to Kafka with attempt count
- After 3 attempts: write to `dead_letter_queue` table
- Simulate failures deliberately to verify recovery works

---

## Implementation notes

- **Retries:** `maxAttempts = 3` (0-based `Attempt` field in `queue.JobMessage`). On failure, if `Attempt < maxAttempts-1`, the worker marks the job `retrying` in Postgres and re-publishes with `Attempt+1`. On the final attempt (`Attempt == maxAttempts-1`), the job is marked `failed` and written to `dead_letter_queue`.
- **Backoff (`ReadBackoff`):** `200ms × 2^n`, capped at 30s. Applied to Kafka *read* errors in the worker loop. Re-publish retries go back into the Kafka topic; the consumer group's natural re-delivery timing provides spacing between delivery attempts.
- **DLQ:** After three failed attempts, the job lands in `dead_letter_queue` with `job_id`, `job_type`, `payload` (JSONB), and `error_message`. It is never retried from there automatically.
- **Test job type:** `"fail"` always returns an error, used to drive the retry and DLQ paths in unit tests and demos without needing real infrastructure.

Exact JSON schemas for job payloads and DB columns can be added when you implement.
