package main

import (
	"context"
	"encoding/json"
	"log"
	"os/signal"
	"syscall"
	"time"

	"flowbit/backend/internal/config"
	"flowbit/backend/internal/db"
	"flowbit/backend/internal/kafka"
	"flowbit/backend/internal/queue"
	"flowbit/backend/internal/repo"
	"flowbit/backend/internal/worker"
)

// kafkaJobPublisher wraps a kafka.Writer so it satisfies worker.Publisher.
type kafkaJobPublisher struct {
	writer *kafka.Writer
}

func (p kafkaJobPublisher) PublishJob(ctx context.Context, msg queue.JobMessage) error {
	return kafka.PublishJob(ctx, p.writer, msg)
}

func main() {
	config.LoadDotenv()
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()
	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("postgres error: %v", err)
	}
	defer pool.Close()

	if cfg.ApplyMigrations {
		if err := db.EnsureSchema(ctx, pool); err != nil {
			log.Fatalf("schema error: %v", err)
		}
	}

	jobsRepo := repo.NewJobsRepo(pool)

	writer, err := kafka.NewWriter(kafka.Config{
		Brokers:  cfg.KafkaBrokers,
		Topic:    cfg.KafkaTopicJobs,
		CertFile: cfg.KafkaCertFile,
		KeyFile:  cfg.KafkaKeyFile,
		CAFile:   cfg.KafkaCAFile,
	})
	if err != nil {
		log.Fatalf("kafka writer error: %v", err)
	}
	defer writer.Close()

	pub := kafkaJobPublisher{writer: writer}

	reader, err := kafka.NewReader(kafka.Config{
		Brokers:  cfg.KafkaBrokers,
		Topic:    cfg.KafkaTopicJobs,
		GroupID:  cfg.KafkaConsumerGrp,
		CertFile: cfg.KafkaCertFile,
		KeyFile:  cfg.KafkaKeyFile,
		CAFile:   cfg.KafkaCAFile,
	})
	if err != nil {
		log.Fatalf("kafka reader error: %v", err)
	}
	defer reader.Close()

	log.Printf("worker consuming topic %q as group %q", cfg.KafkaTopicJobs, cfg.KafkaConsumerGrp)
	var errCount int
	for {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				log.Println("worker shutting down")
				return
			}
			sleep := worker.ReadBackoff(errCount)
			errCount++
			log.Printf("kafka read error (retry in %s): %v", sleep, err)
			select {
			case <-time.After(sleep):
			case <-ctx.Done():
				return
			}
			continue
		}
		errCount = 0

		var jobMsg queue.JobMessage
		if err := json.Unmarshal(msg.Value, &jobMsg); err != nil {
			log.Printf("skip malformed message: %v", err)
			continue
		}
		worker.HandleJob(ctx, jobsRepo, pub, jobMsg, log.Printf)
	}
}
