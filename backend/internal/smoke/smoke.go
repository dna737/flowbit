package smoke

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"flowbit/backend/internal/config"
	"flowbit/backend/internal/db"
	"flowbit/backend/internal/kafka"
	"flowbit/backend/internal/queue"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
)

func Run(ctx context.Context, cfg config.Config) error {
	if strings.TrimSpace(cfg.RedisURL) == "" {
		return errors.New("REDIS_URL is required for smoke checks (api and worker do not need Redis)")
	}
	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		if err := checkPostgres(ctx, cfg); err != nil {
			return fmt.Errorf("postgres check failed: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		if err := checkKafka(ctx, cfg); err != nil {
			return fmt.Errorf("kafka check failed: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		if err := checkRedis(ctx, cfg); err != nil {
			return fmt.Errorf("redis check failed: %w", err)
		}
		return nil
	})
	return g.Wait()
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
	writer := kafka.NewWriter(kafka.Config{
		Brokers:    cfg.KafkaBrokers,
		Topic:      cfg.KafkaTopicJobs,
		User:       cfg.KafkaUsername,
		Pass:       cfg.KafkaPassword,
		DisableTLS: !cfg.KafkaUseTLS,
	})
	defer writer.Close()

	testCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	return kafka.PublishJob(testCtx, writer, queueSmokeMessage())
}

func checkRedis(ctx context.Context, cfg config.Config) error {
	opt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		return err
	}
	client := redis.NewClient(opt)
	defer client.Close()

	if err := client.Ping(ctx).Err(); err != nil {
		return err
	}
	return nil
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
