package main

import (
	"context"

	"flowbit/backend/internal/kafka"
	"flowbit/backend/internal/queue"
)

type kafkaJobPublisher struct {
	writer *kafka.Writer
}

func (p kafkaJobPublisher) PublishJob(ctx context.Context, msg queue.JobMessage) error {
	return kafka.PublishJob(ctx, p.writer, msg)
}
