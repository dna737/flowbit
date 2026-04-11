# Flowbit

Backend-first implementation of the Flowbit scheduler using managed services (Neon Postgres + Upstash Kafka/Redis).

## Prerequisites

- Go 1.24+
- Managed service credentials for:
  - Neon Postgres
  - Upstash Kafka
  - Upstash Redis (only if you run `go run ./cmd/smoke`; API and worker do not need Redis)

## Environment setup

1. Copy `.env.example` to `.env`.
2. Fill in `DATABASE_URL` and `KAFKA_*` for API/worker. Add `REDIS_URL` when running smoke checks.

PowerShell example:

```powershell
Copy-Item .env.example .env
```

## Block 1 smoke checks

Run connectivity checks for Postgres (`SELECT 1`), Kafka (produce message), and Redis (`PING`):

```powershell
cd backend
go run ./cmd/smoke
```

Expected output contains:
- `smoke checks passed: postgres + kafka + redis`

## Block 2 run flow

Start API and worker in separate terminals:

Terminal 1:

```powershell
cd backend
go run ./cmd/api
```

Terminal 2:

```powershell
cd backend
go run ./cmd/worker
```

Create a dummy job (`echo`) from a third terminal:

```powershell
curl -X POST http://localhost:8080/jobs `
  -H "Content-Type: application/json" `
  -d "{\"job_type\":\"echo\",\"parameters\":{\"message\":\"hello flowbit\"}}"
```

Then fetch status:

```powershell
curl http://localhost:8080/jobs/<job-id>
```

A successful end-to-end run transitions job status to `succeeded`.

## Automated tests

From `backend/`:

```powershell
go test ./...
```

This runs unit tests only (HTTP handlers with fakes, repo with pgxmock, worker job logic, Kafka TLS defaults, config defaults). No cloud credentials required.

**Kafka TLS:** Managed brokers (e.g. Upstash) need TLS. That is the default (`KAFKA_USE_TLS` unset or `true`). Set `KAFKA_USE_TLS=false` only for local plaintext Kafka.

**Optional integration test (Docker):** Spins up Postgres via Testcontainers, applies schema, and round-trips `CreateJob` / `GetJobByID`.

```powershell
cd backend
$env:INTEGRATION=1
go test -tags=integration -v ./integration/...
```

**Optional managed E2E:** Same manual flow as Block 2, but script it in CI when `DATABASE_URL` and `KAFKA_*` secrets are present; skip the job when secrets are missing so PRs stay green.
