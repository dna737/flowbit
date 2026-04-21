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

Run connectivity checks for Postgres (`SELECT 1`), optional schema apply when `APPLY_MIGRATIONS=true`, verification that `jobs` and `dead_letter_queue` exist, and Kafka (produce message when TLS certs are set):

```powershell
cd backend
go run ./cmd/smoke
```

Expected output contains:
- `smoke: tables jobs + dead_letter_queue present`
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

Create a dummy job (`general` is the seeded default label for a new user) from a third terminal. The `X-User-Id` header picks your row in `users`; the `job_type` must match one of the labels stored in your `dispatch_categories` (edit them from the Settings dialog in the UI, or via `PUT /settings/dispatch-categories`):

```powershell
curl -X POST http://localhost:8080/jobs `
  -H "Content-Type: application/json" `
  -H "X-User-Id: demo" `
  -d "{\"job_type\":\"general\",\"parameters\":{\"message\":\"hello flowbit\"}}"
```

Then fetch status:

```powershell
curl http://localhost:8080/jobs/<job-id>
```

A successful end-to-end run transitions job status to `succeeded`.

## Block 5 live visualizer

Start the API, worker, and frontend in separate terminals:

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

Terminal 3:

```powershell
cd frontend
npm run dev
```

Open `http://localhost:5173`, submit a prompt, and the board will update over WebSocket as the worker moves the job through each state. Configure `ALLOWED_ORIGINS` in the root `.env` when the UI is hosted anywhere other than the default localhost dev ports.

## Automated tests

From `backend/`:

```powershell
go test ./...
```

This runs unit tests only (HTTP handlers with fakes, repo with pgxmock, worker job logic, Kafka TLS defaults, config defaults). No cloud credentials required.

## Contribution workflow

Flowbit uses a PR-first process:

1. Create a dedicated branch from `main` for each change.
2. Run relevant checks before pushing (backend minimum: `cd backend && go test ./...`).
3. Open a PR targeting `main` and include summary + test plan.

Detailed guidance: see `CONTRIBUTING.md`.

**Kafka TLS:** Aiven Kafka requires TLS with certificate authentication. Place your `service.cert`, `service.key`, and `ca.pem` files in the project root or specify their paths via environment variables (use `../service.key` etc. when `go run` cwd is `backend/`). If smoke fails with **not a PEM private key**, `service.key` is missing or truncated—re-download it from Aiven (Kafka service → **Connection information**), next to `service.cert`.

**Optional integration test (Docker):** Spins up Postgres via Testcontainers, applies schema, and round-trips `CreateJob` / `GetJobByID`.

```powershell
cd backend
$env:INTEGRATION=1
go test -tags=integration -v ./integration/...
```

**Optional managed stack test (Neon + Kafka + worker, no HTTP):** Uses the same `.env` as smoke. Creates a `general` job row, publishes to Kafka, consumes with a one-off group at `LastOffset`, runs `worker.HandleJob`, asserts `succeeded`.

```powershell
cd backend
$env:E2E_STACK = "1"
go test -tags=e2e -count=1 ./integration -run TestStack_genericJob_endToEnd -v
```

Skip when `E2E_STACK` is unset so `go test ./...` stays credential-free.

**Block 2 manual E2E:** With `go run ./cmd/api` and `go run ./cmd/worker`, use the `curl` flow above; CI can script the same against a deployed URL when secrets exist.
