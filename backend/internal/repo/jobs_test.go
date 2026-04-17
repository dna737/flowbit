package repo

import (
	"context"
	"errors"
	"testing"
	"time"

	"flowbit/backend/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
)

func TestJobsRepo_CreateJob(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	t0 := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	rows := mock.NewRows([]string{"id", "job_type", "parameters", "status", "attempts", "last_error", "created_at", "updated_at"}).
		AddRow("550e8400-e29b-41d4-a716-446655440000", "echo", []byte(`{"k":"v"}`), "pending", 0, nil, t0, t0)

	mock.ExpectQuery(`INSERT INTO jobs`).
		WithArgs("echo", pgxmock.AnyArg(), models.JobStatusPending).
		WillReturnRows(rows)

	r := NewJobsRepo(mock)
	job, err := r.CreateJob(context.Background(), "echo", map[string]any{"k": "v"}, models.JobStatusPending)
	if err != nil {
		t.Fatal(err)
	}
	if job.ID != "550e8400-e29b-41d4-a716-446655440000" || job.JobType != "echo" || job.Status != "pending" {
		t.Fatalf("unexpected job: %+v", job)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestJobsRepo_GetJobByID_invalidUUID(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	r := NewJobsRepo(mock)
	_, err = r.GetJobByID(context.Background(), "not-a-uuid")
	if !errors.Is(err, ErrJobNotFound) {
		t.Fatalf("want ErrJobNotFound, got %v", err)
	}
}

func TestJobsRepo_GetJobByID_notFound(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	mock.ExpectQuery(`SELECT id::text`).
		WithArgs("550e8400-e29b-41d4-a716-446655440001").
		WillReturnError(pgx.ErrNoRows)

	r := NewJobsRepo(mock)
	_, err = r.GetJobByID(context.Background(), "550e8400-e29b-41d4-a716-446655440001")
	if !errors.Is(err, ErrJobNotFound) {
		t.Fatalf("want ErrJobNotFound, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestJobsRepo_UpdateJobStatus_noRows(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	mock.ExpectQuery(`WITH updated AS`).
		WithArgs("550e8400-e29b-41d4-a716-446655440002", models.JobStatusRunning, pgxmock.AnyArg()).
		WillReturnError(pgx.ErrNoRows)

	r := NewJobsRepo(mock)
	err = r.UpdateJobStatus(context.Background(), "550e8400-e29b-41d4-a716-446655440002", models.JobStatusRunning, nil)
	if !errors.Is(err, ErrJobNotFound) {
		t.Fatalf("want ErrJobNotFound, got %v", err)
	}
}

func TestJobsRepo_UpdateJobStatus_success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	rows := mock.NewRows([]string{"id"}).AddRow("550e8400-e29b-41d4-a716-446655440003")
	mock.ExpectQuery(`WITH updated AS`).
		WithArgs("550e8400-e29b-41d4-a716-446655440003", models.JobStatusSucceeded, pgxmock.AnyArg()).
		WillReturnRows(rows)

	r := NewJobsRepo(mock)
	if err := r.UpdateJobStatus(context.Background(), "550e8400-e29b-41d4-a716-446655440003", models.JobStatusSucceeded, nil); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestJobsRepo_ListJobs(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	now := time.Date(2026, 3, 2, 10, 0, 0, 0, time.UTC)
	lastError := "boom"
	rows := mock.NewRows([]string{"id", "job_type", "parameters", "status", "attempts", "last_error", "created_at", "updated_at"}).
		AddRow("550e8400-e29b-41d4-a716-446655440010", "echo", []byte(`{"message":"newest"}`), models.JobStatusSucceeded, 1, nil, now, now).
		AddRow("550e8400-e29b-41d4-a716-446655440011", "fail", []byte(`{"kind":"older"}`), models.JobStatusFailed, 3, &lastError, now.Add(-time.Minute), now.Add(-time.Minute))

	mock.ExpectQuery(`SELECT id::text, job_type, parameters`).
		WithArgs(2).
		WillReturnRows(rows)

	r := NewJobsRepo(mock)
	jobs, err := r.ListJobs(context.Background(), 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 2 {
		t.Fatalf("want 2 jobs got %d", len(jobs))
	}
	if jobs[0].Parameters["message"] != "newest" {
		t.Fatalf("unexpected first job params: %+v", jobs[0].Parameters)
	}
	if jobs[1].Status != models.JobStatusFailed {
		t.Fatalf("unexpected second job status: %s", jobs[1].Status)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestJobsRepo_WriteToDLQ_invalidJobID(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	r := NewJobsRepo(mock)
	err = r.WriteToDLQ(context.Background(), "bad-id", "echo", []byte(`{}`), "oops")
	if err == nil {
		t.Fatal("expected error")
	}
}
