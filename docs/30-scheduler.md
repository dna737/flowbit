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

- **Retries:** Exponential backoff with optional jitter between attempts; document parameters here or in code.
- **DLQ:** After three failed attempts, the job must not retry indefinitely; store enough context to inspect failures later.

Exact JSON schemas for job payloads and DB columns can be added when you implement.
