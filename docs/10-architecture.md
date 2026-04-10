# Architecture

See also: [Vision and demo](00-vision-and-demo.md) · [Scheduler spec](30-scheduler.md) · [AI dispatcher](40-ai-dispatcher.md) · [Visualizer](50-visualizer.md)

---

## 2. The three layers

| Layer | What it is |
| ----- | ---------- |
| **Scheduler (core)** | The engine. Kafka, workers, PostgreSQL, retries, dead-letter queue. Deployable and testable without AI or a UI. |
| **AI dispatcher** | The brain. Gemini API parses plain English into a structured job payload and forwards it to the scheduler. |
| **Live visualizer** | The face. A React UI connected via WebSocket that shows each job moving through pipeline stages in real time. |

**Build order is fixed:** scheduler first, then AI dispatcher, then visualizer. Each layer proves the one before it is working.

---

## Repository layout (planned)

When you add code, align package or module names with the spec boundaries:

- **`scheduler`** (or equivalent): Kafka produce/consume, PostgreSQL job state, retries, DLQ — see [30-scheduler.md](30-scheduler.md).
- **`dispatcher`** (or equivalent): Gemini client, `POST /dispatch`, validation, forward to `/jobs` — see [40-ai-dispatcher.md](40-ai-dispatcher.md).
- **`realtime`** (or equivalent): WebSocket hub, streaming status to the UI — see [50-visualizer.md](50-visualizer.md).

Process layout is flexible (e.g. `cmd/api` + `cmd/worker` vs a single binary); keep logical boundaries clear in code.
