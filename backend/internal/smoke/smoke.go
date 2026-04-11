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
	return nil
}

func checkKafka(ctx context.Context, cfg config.Config) error {
	writer, err := kafka.NewWriter(kafka.Config{
		Brokers:  cfg.KafkaBrokers,
		Topic:    cfg.KafkaTopicJobs,
		CertFile: cfg.KafkaCertFile,
		KeyFile:  cfg.KafkaKeyFile,
		CAFile:   cfg.KafkaCAFile,
	})
	if err != nil {
		return fmt.Errorf("failed to create kafka writer: %w", err)
	}
	defer writer.Close()

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
