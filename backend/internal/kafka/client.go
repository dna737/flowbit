package kafka

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"time"

	"flowbit/backend/internal/queue"

	kafkago "github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
)

type Writer = kafkago.Writer
type Reader = kafkago.Reader

type Config struct {
	Brokers []string
	Topic   string
	GroupID string
	User    string
	Pass    string
	// DisableTLS turns off TLS for plaintext local brokers only. Default false = TLS enabled (safe default).
	DisableTLS bool
}

// TLSEnabled reports whether TLS is used for broker connections.
func (c Config) TLSEnabled() bool {
	return !c.DisableTLS
}

func NewWriter(cfg Config) *kafkago.Writer {
	return &kafkago.Writer{
		Addr:         kafkago.TCP(cfg.Brokers...),
		Topic:        cfg.Topic,
		RequiredAcks: kafkago.RequireAll,
		BatchTimeout: 200 * time.Millisecond,
		Transport:    transport(cfg.User, cfg.Pass, cfg.TLSEnabled()),
	}
}

func NewReader(cfg Config) *kafkago.Reader {
	return kafkago.NewReader(kafkago.ReaderConfig{
		Brokers:   cfg.Brokers,
		GroupID:   cfg.GroupID,
		Topic:     cfg.Topic,
		MinBytes:  1,
		MaxBytes:  10e6,
		Dialer:    dialer(cfg.User, cfg.Pass, cfg.TLSEnabled()),
	})
}

func PublishJob(ctx context.Context, writer *kafkago.Writer, msg queue.JobMessage) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal job message: %w", err)
	}
	return writer.WriteMessages(ctx, kafkago.Message{
		Key:   []byte(msg.JobID),
		Value: body,
		Time:  time.Now().UTC(),
	})
}

var minTLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}

func dialer(user, pass string, useTLS bool) *kafkago.Dialer {
	d := &kafkago.Dialer{
		Timeout:   10 * time.Second,
		DualStack: true,
	}
	if useTLS {
		d.TLS = minTLSConfig
	}
	if user != "" || pass != "" {
		d.SASLMechanism = plain.Mechanism{
			Username: user,
			Password: pass,
		}
	}
	return d
}

func transport(user, pass string, useTLS bool) *kafkago.Transport {
	tr := &kafkago.Transport{}
	if useTLS {
		tr.TLS = minTLSConfig
	}
	if user != "" || pass != "" {
		tr.SASL = plain.Mechanism{
			Username: user,
			Password: pass,
		}
	}
	return tr
}
