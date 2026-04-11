package main

import (
	"context"
	"log"
	"time"

	"flowbit/backend/internal/config"
	"flowbit/backend/internal/smoke"
)

func main() {
	config.LoadDotenv()
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	if err := smoke.Run(ctx, cfg); err != nil {
		log.Fatalf("smoke checks failed: %v", err)
	}

	log.Println("smoke checks passed: postgres + kafka")
}
