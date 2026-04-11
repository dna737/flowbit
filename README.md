# Flowbit

Backend-first implementation of the Flowbit scheduler using managed services (Neon Postgres + Aiven Kafka).

## Prerequisites

- Go 1.24+
- Managed service credentials for:
  - Neon Postgres
  - Aiven (Kafka - download service.cert, service.key, ca.pem from Aiven Console)

## Environment setup

1. Copy `.env.example` to `.env` at the **repository root** (or in `backend/`; both are loaded).
2. Fill in `DATABASE_URL` and `KAFKA_*` for API, worker, and smoke checks.
3. `go run ./cmd/...` from `backend/` automatically loads `../.env` then `./.env` so you do not need to export variables manually.

PowerShell example:

```powershell
Copy-Item .env.example .env
```

## Block 1 smoke checks

Run connectivity checks for Postgres (`SELECT 1`) and Kafka (produce message):

```powershell
cd backend
go run ./cmd/smoke
```

Expected output contains:
- `smoke checks passed: postgres + kafka`

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

**Kafka TLS:** Aiven Kafka requires TLS with certificate authentication. Place your `service.cert`, `service.key`, and `ca.pem` files in the project root or specify their paths via environment variables (use `../service.key` etc. when `go run` cwd is `backend/`). If smoke fails with **not a PEM private key**, `service.key` is missing or truncated—re-download it from Aiven (Kafka service → **Connection information**), next to `service.cert`.

**Optional integration test (Docker):** Spins up Postgres via Testcontainers, applies schema, and round-trips `CreateJob` / `GetJobByID`.

```powershell
cd backend
$env:INTEGRATION=1
go test -tags=integration -v ./integration/...
```

**Optional managed E2E:** Same manual flow as Block 2, but script it in CI when `DATABASE_URL` and `KAFKA_*` secrets are present; skip the job when secrets are missing so PRs stay green.
