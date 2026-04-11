# Stack and deployment

See also: [Observability and runbook](60-observability-and-runbook.md) · [Build checklist](BUILD-CHECKLIST.md)

---

## 4. Full technology stack

| Technology | Role | Where it runs |
| ---------- | ---- | ------------- |
| Go | API server + worker pool | Railway or Render |
| Kafka | Job queue / transport layer | Redpanda Cloud (managed) |
| PostgreSQL | Durable job state store | Neon (managed) |
| Gemini API | NL → structured job parsing | Google AI Studio credits |
| React | Live pipeline visualizer UI | Vercel |
| Prometheus | Metrics collection | Grafana Cloud (free tier) |
| Grafana | Dashboard + alerts | Grafana Cloud (free tier) |

Every service is managed and free-tier. Nothing runs only on localhost.

---

## 8. Deployment targets

| Component | Platform |
| --------- | -------- |
| Go API server + workers | Railway or Render |
| Kafka | Redpanda Cloud — managed, free tier, real URLs |
| PostgreSQL | Neon — managed, free tier |
| React UI | Vercel |
| Prometheus + Grafana | Grafana Cloud — free tier, public dashboard URL |

---

## Environment variables (checklist)

Copy values from each provider into a local `.env` (or your secret store). Check off when verified.

| Variable / secret | Source | Used by |
| ----------------- | ------ | ------- |
| PostgreSQL connection string (Neon) | Neon dashboard | API, worker |
| Kafka bootstrap + credentials | Redpanda Cloud | API producer, worker consumer |
| Gemini API key | Google AI Studio | Dispatcher |
| Grafana Cloud Prometheus `remote_write` (or scrape config) | Grafana Cloud | Block 6 metrics |
| Public URLs / deployment secrets | Railway, Render, Vercel | Deploy |

Adjust names to match your application’s convention (e.g. `DATABASE_URL`, `KAFKA_BROKERS`, `GEMINI_API_KEY`).
