package config

import "testing"

func TestLoad_KafkaUseTLS_defaultsTrue(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://u:p@localhost/db?sslmode=disable")
	t.Setenv("KAFKA_BROKERS", "localhost:9092")
	t.Setenv("KAFKA_USE_TLS", "")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.KafkaUseTLS {
		t.Fatal("KAFKA_USE_TLS should default to true when unset")
	}
}

func TestLoad_KafkaUseTLS_false(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://u:p@localhost/db?sslmode=disable")
	t.Setenv("KAFKA_BROKERS", "localhost:9092")
	t.Setenv("KAFKA_USE_TLS", "false")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.KafkaUseTLS {
		t.Fatal("expected KafkaUseTLS false")
	}
}
