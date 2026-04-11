# Observability and day-zero runbook

See also: [Stack and deployment](20-stack-and-deployment.md) · [Build checklist](BUILD-CHECKLIST.md)

---

## Block 1 — Cloud services setup

- Create Neon project → copy connection string → run SQL to create `jobs` and `dead_letter_queue` tables
- Create Aiven Kafka service at console.aiven.io → create topic `jobs` → download TLS certificates (service.cert, service.key, ca.pem) from the Overview page → copy Service URI to `.env`
- Create Grafana Cloud account → spin up hosted Prometheus endpoint → copy `remote_write` URL
- Add all credentials to a local `.env` file → verify Go can connect to each service

If you defer Grafana wiring until late in the project, note that explicitly here so Block 1 still has a clear “observability prereq” state.

---

## Block 6 — Observability + polish

- Prometheus metrics endpoint on the Go server
- Grafana Cloud dashboard: queue depth, job throughput, failure rate
- Deploy all services to cloud targets
- Record the demo (link or file reference below)

### Demo artifact

| Recording / notes | Link or path |
| ----------------- | ------------ |
| *(add when ready)* | |
