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
	Brokers []string
	Topic   string
	GroupID string
	// Aiven uses TLS certificate authentication
	CertFile string // Path to service.cert
	KeyFile  string // Path to service.key
	CAFile   string // Path to ca.pem
}

// TLSEnabled reports whether TLS certificate files are configured.
func (c Config) TLSEnabled() bool {
	return c.CertFile != "" && c.KeyFile != "" && c.CAFile != ""
}

// loadTLSConfig creates a TLS configuration for Aiven certificate authentication.
func loadTLSConfig(certFile, keyFile, caFile string) (*tls.Config, error) {
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

	// Load CA certificate (ca.pem)
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
	var tlsConfig *tls.Config
	var err error

	if cfg.TLSEnabled() {
		tlsConfig, err = loadTLSConfig(cfg.CertFile, cfg.KeyFile, cfg.CAFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load TLS config: %w", err)
		}
	}

	transport := &kafkago.Transport{
		TLS: tlsConfig,
	}

	return &kafkago.Writer{
		Addr:         kafkago.TCP(cfg.Brokers...),
		Topic:        cfg.Topic,
		RequiredAcks: kafkago.RequireAll,
		BatchTimeout: 200 * time.Millisecond,
		Transport:    transport,
	}, nil
}

func NewReader(cfg Config) (*kafkago.Reader, error) {
	dialer := &kafkago.Dialer{
		Timeout:   10 * time.Second,
		DualStack: true,
	}

	if cfg.TLSEnabled() {
		tlsConfig, err := loadTLSConfig(cfg.CertFile, cfg.KeyFile, cfg.CAFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load TLS config: %w", err)
		}
		dialer.TLS = tlsConfig
	}

	return kafkago.NewReader(kafkago.ReaderConfig{
		Brokers:  cfg.Brokers,
		GroupID:  cfg.GroupID,
		Topic:    cfg.Topic,
		MinBytes: 1,
		MaxBytes: 10e6,
		Dialer:   dialer,
	}), nil
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
