package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"flowbit/backend/internal/api"
	"flowbit/backend/internal/config"
	"flowbit/backend/internal/db"
	"flowbit/backend/internal/dispatcher"
	"flowbit/backend/internal/kafka"
	"flowbit/backend/internal/repo"
)

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

	var aiDispatcher api.AIDispatcher
	if cfg.GeminiAPIKey != "" {
		d, err := dispatcher.NewGeminiDispatcher(cfg.GeminiAPIKey, cfg.GeminiModel)
		if err != nil {
			log.Fatalf("gemini dispatcher error: %v", err)
		}
		aiDispatcher = d
	} else {
		log.Printf("GEMINI_API_KEY not set — POST /dispatch will return 501")
	}

	srv := &api.Server{
		Store:        jobsRepo,
		Publisher:    kafkaJobPublisher{writer: writer},
		AIDispatcher: aiDispatcher,
	}
	mux := http.NewServeMux()
	srv.Mount(mux)

	server := &http.Server{
		Addr:              cfg.APIAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		<-ctx.Done()
		stop()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("server shutdown: %v", err)
		}
	}()

	log.Printf("api listening on %s", cfg.APIAddr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server error: %v", err)
	}
}
