//go:build integration

package integration

import (
	"context"
	"os"
	"testing"

	"flowbit/backend/internal/db"
	"flowbit/backend/internal/models"
	"flowbit/backend/internal/repo"

	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func TestPostgresRepo_createJob_roundTrip(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("set INTEGRATION=1 and ensure Docker is running")
	}
	ctx := context.Background()
	container, err := postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithDatabase("flowbit"),
		postgres.WithUsername("flowbit"),
		postgres.WithPassword("flowbit"),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}

	pool, err := db.NewPool(ctx, connStr)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	if err := db.EnsureSchema(ctx, pool); err != nil {
		t.Fatal(err)
	}

	r := repo.NewJobsRepo(pool)
	job, err := r.CreateJob(ctx, "echo", map[string]any{"x": 1.0}, models.JobStatusPending)
	if err != nil {
		t.Fatal(err)
	}
	got, err := r.GetJobByID(ctx, job.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.JobType != "echo" || got.Status != models.JobStatusPending {
		t.Fatalf("unexpected job: %+v", got)
	}
}
