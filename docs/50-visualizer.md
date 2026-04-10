# Live visualizer

See also: [Architecture](10-architecture.md) · [AI dispatcher](40-ai-dispatcher.md) · [Stack](20-stack-and-deployment.md) · [Build checklist](BUILD-CHECKLIST.md)

---

## Block 5 — Live visualizer

- WebSocket endpoint streams job status changes from the Go server
- React UI shows pipeline stages (queued → running → done / failed) lighting up
- Connect to the AI dispatcher input box

---

## Contract (sketch)

- **Stages:** Model the UI around discrete states that match what the API and worker write to PostgreSQL (and optionally emit over the socket).
- **Transport:** WebSocket by default; if you switch to SSE, update this doc and the architecture note.
