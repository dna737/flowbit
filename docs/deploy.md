# Deploy

The root [Dockerfile](../backend/Dockerfile) builds a multi-stage, distroless static image for `./cmd/api` (Go 1.25, CGO off, non-root user, listens on `8080`).

The worker is **not** deployed to Cloud Run — it needs a steady consumer process (run `./cmd/worker` on Compute Engine `e2-micro` or similar), because Cloud Run scales to zero.

## Cloud Run wiring

- **Build type:** Dockerfile, **source location:** `/Dockerfile` (repo root)
- **Container port:** `8080`
- **Plain env vars:**
  - `API_ADDR=:8080`
  - `APPLY_MIGRATIONS=false` (apply schema once out-of-band; multiple instances racing `EnsureSchema` on cold start is bad)
  - `ALLOWED_ORIGINS=<prod UI origin>`
  - `KAFKA_TOPIC_JOBS=jobs`
  - `GEMINI_MODEL` (optional)
- **Secret Manager → env secrets:** `DATABASE_URL`, `KAFKA_BROKERS`, `GEMINI_API_KEY`
- **Secret Manager → mounted files** at `/secrets/`: `service.cert`, `service.key`, `ca.pem`. Then set:
  - `KAFKA_CERT_FILE=/secrets/service.cert`
  - `KAFKA_KEY_FILE=/secrets/service.key`
  - `KAFKA_CA_FILE=/secrets/ca.pem`

Absolute cert paths are used as-is by the config loader (no `.env` in the container).

## Migrations

Set `APPLY_MIGRATIONS=false` on Cloud Run. Apply schema once from a trusted host (local dev, or a one-off Cloud Run Job) against Neon before deploying, and again only when the schema changes.

## WebSocket caveat

Cloud Run supports WebSockets but caps a single request at 60 minutes. The `/ws` visualizer stream will reconnect past that — fine for Flowbit, but worth knowing.

## Verify the image locally

```powershell
docker build -t flowbit-api .
docker run --rm -p 8080:8080 `
  -e API_ADDR=:8080 `
  -e APPLY_MIGRATIONS=false `
  -e DATABASE_URL="<neon-url>" `
  -e KAFKA_BROKERS="<broker:port>" `
  flowbit-api
curl http://localhost:8080/healthz
```

## Kafka TLS troubleshooting

Aiven Kafka requires TLS with certificate authentication. Place `service.cert`, `service.key`, and `ca.pem` in the project root, or specify their paths via `KAFKA_*_FILE` environment variables (use `../service.key` etc. when `go run` cwd is `backend/`).

If smoke fails with **"not a PEM private key"**, `service.key` is missing or truncated — re-download it from Aiven (Kafka service → **Connection information**), next to `service.cert`.
