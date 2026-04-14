package kafka

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"flowbit/backend/internal/queue"

	kafkago "github.com/segmentio/kafka-go"
)

type Writer = kafkago.Writer
type Reader = kafkago.Reader

type Config struct {
	Brokers  []string
	Topic    string
	GroupID  string
	// Aiven uses TLS certificate authentication
	CertFile  string      // Path to service.cert
	KeyFile   string      // Path to service.key
	CAFile    string      // Path to ca.pem
	TLSConfig *tls.Config // Pre-built TLS config; if set, skips file loading
	// StartOffset for a new consumer group: 0 = kafka-go default (FirstOffset). Use kafkago.LastOffset to read only new records.
	StartOffset int64
}

// TLSEnabled reports whether TLS is configured (pre-built or via cert files).
func (c Config) TLSEnabled() bool {
	return c.TLSConfig != nil || (c.CertFile != "" && c.KeyFile != "" && c.CAFile != "")
}

// resolveTLS returns the TLS config to use: pre-built if set, otherwise loaded from files.
func resolveTLS(cfg Config) (*tls.Config, error) {
	if cfg.TLSConfig != nil {
		return cfg.TLSConfig, nil
	}
	if cfg.CertFile != "" && cfg.KeyFile != "" && cfg.CAFile != "" {
		return LoadTLSConfig(cfg.CertFile, cfg.KeyFile, cfg.CAFile)
	}
	return nil, nil
}

// LoadTLSConfig creates a TLS configuration for Aiven certificate authentication.
func LoadTLSConfig(certFile, keyFile, caFile string) (*tls.Config, error) {
	certPEM, err := os.ReadFile(certFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read client certificate: %w", err)
	}
	keyPEM, err := os.ReadFile(keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read client key: %w", err)
	}
	if !bytes.Contains(keyPEM, []byte("-----BEGIN ")) {
		return nil, fmt.Errorf("client key %q is not a PEM private key (download service.key from Aiven → Kafka → Connection information)", keyFile)
	}
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to load client certificate: %w", err)
	}

	caCert, err := os.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate: %w", err)
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to parse CA certificate")
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
		MinVersion:   tls.VersionTLS12,
	}, nil
}

func NewWriter(cfg Config) (*kafkago.Writer, error) {
	tlsConfig, err := resolveTLS(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS config: %w", err)
	}

	return &kafkago.Writer{
		Addr:         kafkago.TCP(cfg.Brokers...),
		Topic:        cfg.Topic,
		RequiredAcks: kafkago.RequireAll,
		BatchTimeout: 200 * time.Millisecond,
		Transport:    &kafkago.Transport{TLS: tlsConfig},
	}, nil
}

func NewReader(cfg Config) (*kafkago.Reader, error) {
	tlsConfig, err := resolveTLS(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS config: %w", err)
	}

	dialer := &kafkago.Dialer{
		Timeout:   10 * time.Second,
		DualStack: true,
		TLS:       tlsConfig,
	}

	rc := kafkago.ReaderConfig{
		Brokers:  cfg.Brokers,
		GroupID:  cfg.GroupID,
		Topic:    cfg.Topic,
		MinBytes: 1,
		MaxBytes: 10e6,
		Dialer:   dialer,
	}
	if cfg.StartOffset != 0 {
		rc.StartOffset = cfg.StartOffset
	}
	return kafkago.NewReader(rc), nil
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
