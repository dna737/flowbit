//go:build e2e

package integration

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"flowbit/backend/internal/config"
	"flowbit/backend/internal/db"
	"flowbit/backend/internal/kafka"
	"flowbit/backend/internal/models"
	"flowbit/backend/internal/queue"
	"flowbit/backend/internal/repo"
	"flowbit/backend/internal/worker"

	"github.com/google/uuid"
	kafkago "github.com/segmentio/kafka-go"
)

// workerPublisher wraps a kafka.Writer to satisfy worker.Publisher in e2e tests.
type workerPublisher struct {
	writer *kafka.Writer
}

func (p workerPublisher) PublishJob(ctx context.Context, msg queue.JobMessage) error {
	return kafka.PublishJob(ctx, p.writer, msg)
}

// TestStack_echoJob_endToEnd exercises Neon + Aiven Kafka + worker.HandleJob in one process (no HTTP).
// Requires DATABASE_URL, KAFKA_BROKERS, topic, and TLS cert paths in .env (same as cmd/smoke).
func TestStack_echoJob_endToEnd(t *testing.T) {
	if os.Getenv("E2E_STACK") != "1" {
		t.Skip(`set E2E_STACK=1 and .env with DATABASE_URL + Aiven KAFKA_*; run: go test -tags=e2e -count=1 ./integration -run TestStack_echoJob_endToEnd`)
	}

	config.LoadDotenv()
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	tlsProbe := kafka.Config{CertFile: cfg.KafkaCertFile, KeyFile: cfg.KafkaKeyFile, CAFile: cfg.KafkaCAFile}
	if !tlsProbe.TLSEnabled() {
		t.Fatal("E2E_STACK requires KAFKA_CERT_FILE, KAFKA_KEY_FILE, KAFKA_CA_FILE (Aiven TLS)")
	}

	ctx := context.Background()
	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { pool.Close() })

	if err := db.EnsureSchema(ctx, pool); err != nil {
		t.Fatal(err)
	}

	store := repo.NewJobsRepo(pool)
	job, err := store.CreateJob(ctx, "echo", map[string]any{"e2e": true}, models.JobStatusPending)
	if err != nil {
		t.Fatal(err)
	}

	group := "flowbit-e2e-" + uuid.NewString()
	reader, err := kafka.NewReader(kafka.Config{
		Brokers:     cfg.KafkaBrokers,
		Topic:       cfg.KafkaTopicJobs,
		GroupID:     group,
		CertFile:    cfg.KafkaCertFile,
		KeyFile:     cfg.KafkaKeyFile,
		CAFile:      cfg.KafkaCAFile,
		StartOffset: kafkago.FirstOffset,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = reader.Close() })

	writer, err := kafka.NewWriter(kafka.Config{
		Brokers:  cfg.KafkaBrokers,
		Topic:    cfg.KafkaTopicJobs,
		CertFile: cfg.KafkaCertFile,
		KeyFile:  cfg.KafkaKeyFile,
		CAFile:   cfg.KafkaCAFile,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = writer.Close() })

	jobMsg := queue.JobMessage{JobID: job.ID, JobType: job.JobType, Parameters: job.Parameters}
	if err := kafka.PublishJob(ctx, writer, jobMsg); err != nil {
		t.Fatal(err)
	}

	// Read from the beginning of the topic and scan for our job ID.
	// Avoids the LastOffset race (group-join on remote Kafka can be slow).
	readCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()

	var got queue.JobMessage
	for {
		kmsg, err := reader.ReadMessage(readCtx)
		if err != nil {
			t.Fatalf("read kafka: %v", err)
		}
		var m queue.JobMessage
		if err := json.Unmarshal(kmsg.Value, &m); err != nil {
			continue // skip unrelated messages
		}
		if m.JobID == job.ID {
			got = m
			break
		}
	}
	if got.JobID != job.ID {
		t.Fatalf("job id mismatch: got %q want %q", got.JobID, job.ID)
	}

	// Pass writer as publisher so HandleJob can re-publish on failure.
	// The echo job succeeds, so the publisher is not exercised here.
	pub := workerPublisher{writer: writer}
	worker.HandleJob(ctx, store, pub, got, t.Logf)

	final, err := store.GetJobByID(ctx, job.ID)
	if err != nil {
		t.Fatal(err)
	}
	if final.Status != models.JobStatusSucceeded {
		t.Fatalf("want status %q got %q last_error=%v", models.JobStatusSucceeded, final.Status, final.LastError)
	}
}
