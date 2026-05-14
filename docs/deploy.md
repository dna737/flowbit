# Deploy

The [Dockerfile](../backend/Dockerfile) builds a multi-stage, distroless static image containing both the **API** (`/api`) and **worker** (`/worker`) binaries (Go 1.25, CGO off, non-root user).

Two Cloud Run services are deployed from the same image:
- `flowbit` — API server (HTTPS + WebSocket), default entrypoint `/api`
- `flowbit-worker` — Kafka consumer, entrypoint `/worker`

The worker service uses `--min-instances=1 --cpu-always` so it never scales to zero and can process Kafka messages in the background.

## Cloud Run wiring

- **Container port:** `8080` (both services)
- **Environment variables (shared):**
  - `API_ADDR=:8080`
  - `APPLY_MIGRATIONS=false`
  - `DATABASE_URL` — from GitHub secret
  - `KAFKA_BROKERS` — from GitHub secret
  - `KAFKA_TOPIC_JOBS=jobs`
- **API additional env:** `ALLOWED_ORIGINS`, `GEMINI_API_KEY`
- **Worker additional env:** `KAFKA_CONSUMER_GROUP=flowbit-workers`
- **Secret Manager → mounted files** at `/secrets/`: `service-cert`, `service-key`, `ca-pem`. Then set:
  - `KAFKA_CERT_FILE=/secrets/service-cert`
  - `KAFKA_KEY_FILE=/secrets/service-key`
  - `KAFKA_CA_FILE=/secrets/ca-pem`

## One-time setup

### 1. Grant the runtime service account access to cert secrets

```powershell
$PROJECT_ID = gcloud config get-value project
$PROJECT_NUMBER = gcloud projects describe $PROJECT_ID --format="value(projectNumber)"
$SA = "$PROJECT_NUMBER-compute@developer.gserviceaccount.com"

foreach ($secret in @("service-cert","service-key","ca-pem")) {
  gcloud secrets add-iam-policy-binding $secret `
    --member="serviceAccount:$SA" `
    --role="roles/secretmanager.secretAccessor"
}
```

### 2. Add GitHub Actions secrets

In your repo settings (**Settings → Secrets and variables → Actions**), add:
- `DATABASE_URL` — your Neon Postgres connection string
- `ALLOWED_ORIGINS` — your UI origin (e.g. `https://flowbit.vercel.app`)
- `KAFKA_BROKERS` — your Aiven Kafka broker address
- `GEMINI_API_KEY` (optional) — for the AI dispatch feature
- `GCP_PROJECT_ID` — your GCP project ID
- `GCP_PROJECT_NUMBER` — your GCP project number
- `GCP_SERVICE_ACCOUNT_EMAIL` — the service account used by Workload Identity Federation

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
