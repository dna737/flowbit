package smoke

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"flowbit/backend/internal/config"
	"flowbit/backend/internal/db"
	"flowbit/backend/internal/kafka"
	"flowbit/backend/internal/queue"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Run(ctx context.Context, cfg config.Config) error {
	if err := checkPostgres(ctx, cfg); err != nil {
		return fmt.Errorf("postgres check failed: %w", err)
	}
	log.Println("smoke: postgres OK (Neon)")
	if err := checkKafka(ctx, cfg); err != nil {
		return fmt.Errorf("kafka check failed: %w", err)
	}
	log.Println("smoke: kafka OK (produce to topic)")
	return nil
}

func checkPostgres(ctx context.Context, cfg config.Config) error {
	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	var one int
	if err := pool.QueryRow(ctx, "SELECT 1").Scan(&one); err != nil {
		return err
	}
	if one != 1 {
		return errors.New("SELECT 1 did not return 1")
	}

	if cfg.ApplyMigrations {
		if err := db.EnsureSchema(ctx, pool); err != nil {
			return fmt.Errorf("apply schema: %w", err)
		}
		log.Println("smoke: schema applied (APPLY_MIGRATIONS=true)")
	}

	if err := verifyCoreTables(ctx, pool, cfg.ApplyMigrations); err != nil {
		return err
	}
	log.Println("smoke: tables jobs + dead_letter_queue present")
	return nil
}

func verifyCoreTables(ctx context.Context, pool *pgxpool.Pool, migrationsRan bool) error {
	var n int
	err := pool.QueryRow(ctx, `
		SELECT count(*)::int FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name IN ('jobs', 'dead_letter_queue')
	`).Scan(&n)
	if err != nil {
		return fmt.Errorf("table check: %w", err)
	}
	if n == 2 {
		return nil
	}
	if migrationsRan {
		return fmt.Errorf("expected public.jobs and public.dead_letter_queue after migrations (found %d/2)", n)
	}
	return fmt.Errorf("missing jobs/dead_letter_queue (found %d/2); set APPLY_MIGRATIONS=true or run backend/internal/db/sql/schema.sql in Neon SQL Editor", n)
}

func checkKafka(ctx context.Context, cfg config.Config) error {
	kafkaCfg := kafka.Config{
		Brokers:  cfg.KafkaBrokers,
		Topic:    cfg.KafkaTopicJobs,
		CertFile: cfg.KafkaCertFile,
		KeyFile:  cfg.KafkaKeyFile,
		CAFile:   cfg.KafkaCAFile,
	}
	if !kafkaCfg.TLSEnabled() {
		log.Println("smoke: kafka skipped (no TLS certs configured — set KAFKA_CERT_FILE/KEY_FILE/CA_FILE for Aiven)")
		return nil
	}

	writer, err := kafka.NewWriter(kafkaCfg)
	if err != nil {
		return fmt.Errorf("failed to create kafka writer: %w", err)
	}
	defer writer.Close()

	log.Printf("smoke: publishing to kafka topic %q", cfg.KafkaTopicJobs)
	testCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	return kafka.PublishJob(testCtx, writer, queueSmokeMessage())
}

func queueSmokeMessage() queue.JobMessage {
	return queue.JobMessage{
		JobID:   uuid.New().String(),
		JobType: "smoke-check",
		Parameters: map[string]any{
			"source": "backend/cmd/smoke",
		},
	}
}
