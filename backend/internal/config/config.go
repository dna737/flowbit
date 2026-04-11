package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	APIAddr          string
	DatabaseURL      string
	KafkaBrokers     []string
	KafkaTopicJobs   string
	// Aiven uses TLS certificate authentication (service.cert, service.key, ca.pem)
	KafkaCertFile    string // Path to service.cert
	KafkaKeyFile     string // Path to service.key
	KafkaCAFile      string // Path to ca.pem
	KafkaConsumerGrp string
	ApplyMigrations  bool
}

func Load() (Config, error) {
	cfg := Config{
		APIAddr:          getenv("API_ADDR", ":8080"),
		DatabaseURL:      getenv("DATABASE_URL", ""),
		KafkaBrokers:     splitCSV(os.Getenv("KAFKA_BROKERS")),
		KafkaTopicJobs:   getenv("KAFKA_TOPIC_JOBS", "jobs"),
		// Aiven TLS certificate files
		KafkaCertFile:    getenv("KAFKA_CERT_FILE", "service.cert"),
		KafkaKeyFile:     getenv("KAFKA_KEY_FILE", "service.key"),
		KafkaCAFile:      getenv("KAFKA_CA_FILE", "ca.pem"),
		KafkaConsumerGrp: getenv("KAFKA_CONSUMER_GROUP", "flowbit-workers"),
		ApplyMigrations:  getenvBool("APPLY_MIGRATIONS", true),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}
	if len(cfg.KafkaBrokers) == 0 {
		return Config{}, errors.New("KAFKA_BROKERS is required")
	}
	return cfg, nil
}

func getenv(key, fallback string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	return v
}

func getenvBool(key string, fallback bool) bool {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}

func splitCSV(value string) []string {
	raw := strings.Split(value, ",")
	out := make([]string, 0, len(raw))
	for _, item := range raw {
		part := strings.TrimSpace(item)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}
