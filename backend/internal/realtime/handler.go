package realtime

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"flowbit/backend/internal/models"
	"flowbit/backend/internal/session"

	"github.com/coder/websocket"
)

const (
	snapshotLimit = 100
	writeTimeout  = 5 * time.Second
)

// JobLister returns the most recent jobs owned by a single user.
type JobLister interface {
	ListJobsByUser(ctx context.Context, userID string, limit int) ([]models.Job, error)
}

type ClientHub interface {
	Register(*Client)
	Unregister(*Client)
}

type snapshotMessage struct {
	Type string       `json:"type"`
	Jobs []models.Job `json:"jobs"`
}

// Handler returns the GET /ws handler. The user identity is resolved from the
// server-issued session cookie already attached to the request context.
func Handler(hub ClientHub, lister JobLister, allowedOrigins []string) http.HandlerFunc {
	originPatterns := allowedOriginPatterns(allowedOrigins)

	return func(w http.ResponseWriter, r *http.Request) {
		if hub == nil || lister == nil {
			http.NotFound(w, r)
			return
		}

		userID, ok := session.UserIDFromContext(r.Context())
		if !ok {
			http.Error(w, "session user id is missing", http.StatusUnauthorized)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		jobs, err := lister.ListJobsByUser(ctx, userID, snapshotLimit)
		cancel()
		if err != nil {
			http.Error(w, "failed to load jobs snapshot", http.StatusInternalServerError)
			return
		}
		if jobs == nil {
			jobs = []models.Job{}
		}

		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			OriginPatterns: originPatterns,
		})
		if err != nil {
			return
		}

		client := NewClient(userID, conn)
		hub.Register(client)
		defer hub.Unregister(client)

		runCtx, runCancel := context.WithCancel(r.Context())
		defer runCancel()

		go writePump(runCtx, client)

		snapshot, err := json.Marshal(snapshotMessage{
			Type: "snapshot",
			Jobs: jobs,
		})
		if err != nil {
			log.Printf("marshal snapshot: %v", err)
			return
		}

		select {
		case client.send <- snapshot:
		case <-runCtx.Done():
			return
		}

		readPump(runCtx, conn)
	}
}

func writePump(ctx context.Context, client *Client) {
	defer client.conn.Close(websocket.StatusNormalClosure, "")

	for {
		select {
		case <-ctx.Done():
			return
		case payload, ok := <-client.send:
			if !ok {
				return
			}

			writeCtx, cancel := context.WithTimeout(ctx, writeTimeout)
			err := client.conn.Write(writeCtx, websocket.MessageText, payload)
			cancel()
			if err != nil {
				return
			}
		}
	}
}

func readPump(ctx context.Context, conn *websocket.Conn) {
	for {
		if _, _, err := conn.Read(ctx); err != nil {
			return
		}
	}
}

func allowedOriginPatterns(origins []string) []string {
	patterns := make([]string, 0, len(origins))
	seen := make(map[string]struct{}, len(origins))
	for _, origin := range origins {
		u, err := url.Parse(origin)
		if err != nil || u.Host == "" {
			continue
		}
		host := strings.TrimSpace(u.Host)
		if host == "" {
			continue
		}
		if _, ok := seen[host]; ok {
			continue
		}
		seen[host] = struct{}{}
		patterns = append(patterns, host)
	}
	return patterns
}
