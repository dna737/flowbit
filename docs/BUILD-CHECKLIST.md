# Build checklist

Ordered execution spine for Flowbit. **Do not start the next block until the previous block’s checkpoints are green.**

Related docs: [Vision and demo](00-vision-and-demo.md) · [Stack](20-stack-and-deployment.md) · [Scheduler](30-scheduler.md) · [AI dispatcher](40-ai-dispatcher.md) · [Visualizer](50-visualizer.md) · [Observability](60-observability-and-runbook.md)

---

## Block 1 — Cloud services setup

- [ ] **Neon:** `jobs` and `dead_letter_queue` tables exist per schema you will use in code; a one-off SQL or migration has been applied.
- [ ] **Postgres connectivity:** From Go (or your chosen verifier), open a connection using `DATABASE_URL` and run a trivial query (e.g. `SELECT 1`).
- [ ] **Aiven Kafka:** Topic `jobs` exists; Service URI + TLS certificates (service.cert, service.key, ca.pem) in `.env`; a minimal producer can publish a test message to `jobs` and succeed.
- [ ] **Grafana Cloud:** Account created; hosted Prometheus `remote_write` URL (or equivalent) captured in `.env` if you will push metrics in Block 6 — or explicitly deferred with a note in `60-observability-and-runbook.md` so Block 1 still has a clear “observability prereq” state.
- [ ] **Secrets:** All of the above live in a local `.env` (or secret store); nothing required for Block 2 is still “TODO in dashboard.”

---

## Block 2 — Core scheduler

- [ ] **API:** `POST /jobs` accepts a validated payload, persists a row with `status` appropriate for “pending” (or your chosen initial state), and publishes to Kafka `jobs`.
- [ ] **API:** `GET /jobs/:id` returns the job record (including status) for an existing id; 404 for unknown id.
- [ ] **Worker:** Consumes from the same topic/partition strategy you defined; processes a **hardcoded dummy** job type (no AI).
- [ ] **E2E:** Submit job via API → message in Kafka → worker runs → PostgreSQL shows terminal success state for that job id **without** manual DB edits.
- [ ] **Automated tests:** From `backend/`, `go test ./...` passes (unit tests; no managed secrets required).

---

## Block 3 — Retry + dead-letter queue

- [ ] **Retry:** On failure, job is re-published or re-driven with an incremented attempt count (visible in DB or message metadata).
- [ ] **Backoff:** Exponential backoff (with optional jitter) is applied between attempts — document the parameters in `30-scheduler.md` or code comments.
- [ ] **DLQ:** After **3** failed attempts, the job lands in `dead_letter_queue` (or equivalent) and is **not** infinitely retried.
- [ ] **Proof:** You can **simulate** a failing task and observe: attempts 1–3 → retry path; attempt 3 failure → DLQ row; a success path still completes without touching DLQ.

---

## Block 4 — AI dispatcher

- [ ] **Endpoint:** `POST /dispatch` accepts plain English (body shape documented).
- [ ] **Gemini:** Server calls Gemini (e.g. `gemini-2.5-flash`), parses **structured** output into `job_type`, `parameters`, `priority` (or your schema).
- [ ] **Integration:** Successful parse results in the same flow as Block 2: valid `POST /jobs` payload and job execution through the worker.
- [ ] **Spot checks:** At least three distinct natural-language examples (e.g. email, image resize, URL scrape) each produce a sensible structured job and complete or fail predictably — not only one happy-path string.

---

## Block 5 — Live visualizer

- [ ] **WebSocket:** Server exposes a WebSocket (or SSE if you change the design — then update docs) that emits job status transitions in **near real time** for subscribed or relevant jobs.
- [ ] **UI:** React app shows pipeline stages (e.g. queued → running → done / failed) and updates when the backend changes state.
- [ ] **Wiring:** User can drive the flow from the UI or the same entry point the demo uses (plain English → dispatch) and **see** stages update without refreshing the page manually.
- [ ] **Deploy path:** UI build is deployable to Vercel (or doc-updated target); URL or env is documented in `20-stack-and-deployment.md`.

---

## Block 6 — Observability + polish

- [ ] **Metrics:** Go process exposes a **Prometheus** scrape endpoint; key series exist for at least queue depth, job throughput, and failure rate (names can match Grafana queries).
- [ ] **Grafana:** Dashboard in Grafana Cloud shows those three dimensions with live or recently scraped data from a running deployment.
- [ ] **Deploy:** API + worker + any required workers are deployed to the chosen hosts (Railway/Render per doc); public URL reachable.
- [ ] **Demo:** Short screen recording or scripted demo notes captured (link or file referenced in `60-observability-and-runbook.md` or `00-vision-and-demo.md`).

---

## Cross-cutting — project “ready to show”

- [ ] **Story:** You can repeat the §3 demo flow on the **deployed** stack end-to-end: English → job → visible stages → Grafana not empty — matching the interview line in §10 of [00-vision-and-demo.md](00-vision-and-demo.md).

---

## Documentation checkpoints (splitting the original Word doc)

| ID | Done when |
| -- | --------- |
| **D1** | `00-vision-and-demo.md` contains §1, §3, §10; one-line pitch appears once. |
| **D2** | `10-architecture.md` states three layers and fixed build order. |
| **D3** | `20-stack-and-deployment.md` has stack + deployment tables and a single env-var checklist. |
| **D4** | `30-scheduler.md` covers Blocks 2–3; `40-ai-dispatcher.md` Block 4; `50-visualizer.md` Block 5; `60-observability-and-runbook.md` Blocks 1 & 6. |
| **D5** | `90-study-guide.md` holds §6, §7, §9; cross-links from other docs removed or minimized to avoid drift. |
| **D6** | This file lists Blocks 1–6 with implementation checkpoints (above). |
