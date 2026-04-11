package config

import "testing"

func TestLoad_KafkaCertFiles_defaults(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://u:p@localhost/db?sslmode=disable")
	t.Setenv("KAFKA_BROKERS", "localhost:9092")
	// Clear cert file env vars to test defaults
	t.Setenv("KAFKA_CERT_FILE", "")
	t.Setenv("KAFKA_KEY_FILE", "")
	t.Setenv("KAFKA_CA_FILE", "")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}

	// TLS cert files default to empty (opt-in, not opt-out)
	if cfg.KafkaCertFile != "" {
		t.Fatalf("expected KafkaCertFile default '', got %q", cfg.KafkaCertFile)
	}
	if cfg.KafkaKeyFile != "" {
		t.Fatalf("expected KafkaKeyFile default '', got %q", cfg.KafkaKeyFile)
	}
	if cfg.KafkaCAFile != "" {
		t.Fatalf("expected KafkaCAFile default '', got %q", cfg.KafkaCAFile)
	}
}

func TestLoad_KafkaCertFiles_custom(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://u:p@localhost/db?sslmode=disable")
	t.Setenv("KAFKA_BROKERS", "localhost:9092")
	t.Setenv("KAFKA_CERT_FILE", "/path/to/custom.cert")
	t.Setenv("KAFKA_KEY_FILE", "/path/to/custom.key")
	t.Setenv("KAFKA_CA_FILE", "/path/to/custom-ca.pem")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}

	if cfg.KafkaCertFile != "/path/to/custom.cert" {
		t.Fatalf("expected custom cert file, got %q", cfg.KafkaCertFile)
	}
	if cfg.KafkaKeyFile != "/path/to/custom.key" {
		t.Fatalf("expected custom key file, got %q", cfg.KafkaKeyFile)
	}
	if cfg.KafkaCAFile != "/path/to/custom-ca.pem" {
		t.Fatalf("expected custom CA file, got %q", cfg.KafkaCAFile)
	}
}
