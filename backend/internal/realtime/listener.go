package realtime

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"flowbit/backend/internal/models"

	"github.com/jackc/pgx/v5"
)

const (
	notifyChannel = "job_status"
	maxBackoff    = 30 * time.Second
)

// UserBroadcaster fans a payload out to clients owned by userID.
type UserBroadcaster interface {
	BroadcastToUser(userID string, payload []byte)
}

// JobByID fetches a single job (unscoped — listener trusts the notify payload).
type JobByID interface {
	GetJobByID(ctx context.Context, id string) (models.Job, error)
}

// notifyPayload is the JSON written by jobs.UpdateJobStatus's pg_notify call.
// Kept tiny ({id,user_id}) so we never bump into Postgres' ~8000-byte NOTIFY cap.
type notifyPayload struct {
	ID     string `json:"id"`
	UserID string `json:"user_id"`
}

// Listen connects to Postgres, subscribes to the job_status NOTIFY channel,
// and re-broadcasts each event as a fully-hydrated job to the owning user.
func Listen(ctx context.Context, databaseURL string, hub UserBroadcaster, fetcher JobByID) {
	if hub == nil || fetcher == nil {
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
		err = waitForNotifications(ctx, conn, hub, fetcher)
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

func waitForNotifications(ctx context.Context, conn *pgx.Conn, hub UserBroadcaster, fetcher JobByID) error {
	for {
		notification, err := conn.WaitForNotification(ctx)
		if err != nil {
			return err
		}

		var p notifyPayload
		if err := json.Unmarshal([]byte(notification.Payload), &p); err != nil {
			log.Printf("realtime listener: bad NOTIFY payload: %v", err)
			continue
		}
		if p.ID == "" || p.UserID == "" {
			continue
		}

		// Hydrate the full row so the WS client can update its UI without an
		// extra round-trip. Use a short timeout per fetch so a slow DB doesn't
		// stall the notification loop.
		fetchCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		job, err := fetcher.GetJobByID(fetchCtx, p.ID)
		cancel()
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				log.Printf("realtime listener: fetch %s: %v", p.ID, err)
			}
			continue
		}

		payload, err := json.Marshal(job)
		if err != nil {
			log.Printf("realtime listener: marshal job %s: %v", p.ID, err)
			continue
		}

		hub.BroadcastToUser(p.UserID, payload)
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
