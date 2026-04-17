package realtime

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
)

const (
	notifyChannel = "job_status"
	maxBackoff    = 30 * time.Second
)

type Broadcaster interface {
	Broadcast([]byte)
}

func Listen(ctx context.Context, databaseURL string, hub Broadcaster) {
	if hub == nil {
		return
	}

	backoff := time.Second
	for {
		if ctx.Err() != nil {
			return
		}

		conn, err := pgx.Connect(ctx, databaseURL)
		if err != nil {
			log.Printf("realtime listener connect failed: %v", err)
			if !sleepBackoff(ctx, backoff) {
				return
			}
			backoff = nextBackoff(backoff)
			continue
		}

		if _, err := conn.Exec(ctx, "LISTEN "+notifyChannel); err != nil {
			log.Printf("realtime listener listen failed: %v", err)
			_ = conn.Close(ctx)
			if !sleepBackoff(ctx, backoff) {
				return
			}
			backoff = nextBackoff(backoff)
			continue
		}

		backoff = time.Second
		log.Printf("realtime listener subscribed to %q", notifyChannel)
		err = waitForNotifications(ctx, conn, hub)
		_ = conn.Close(context.Background())
		if ctx.Err() != nil {
			return
		}

		log.Printf("realtime listener reconnecting after error: %v", err)
		if !sleepBackoff(ctx, backoff) {
			return
		}
		backoff = nextBackoff(backoff)
	}
}

func waitForNotifications(ctx context.Context, conn *pgx.Conn, hub Broadcaster) error {
	for {
		notification, err := conn.WaitForNotification(ctx)
		if err != nil {
			return err
		}
		hub.Broadcast([]byte(notification.Payload))
	}
}

func sleepBackoff(ctx context.Context, delay time.Duration) bool {
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func nextBackoff(current time.Duration) time.Duration {
	if current <= 0 {
		return time.Second
	}
	next := current * 2
	if next > maxBackoff {
		return maxBackoff
	}
	return next
}
