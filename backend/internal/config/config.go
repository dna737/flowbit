package config

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Config struct {
	APIAddr        string
	DatabaseURL    string
	AllowedOrigins []string
	KafkaBrokers   []string
	KafkaTopicJobs string
	// Aiven uses TLS certificate authentication (service.cert, service.key, ca.pem)
	KafkaCertFile    string // Path to service.cert
	KafkaKeyFile     string // Path to service.key
	KafkaCAFile      string // Path to ca.pem
	KafkaConsumerGrp string
	ApplyMigrations  bool
	GeminiAPIKey     string // from env GEMINI_API_KEY; empty disables POST /dispatch
	GeminiModel      string // from env GEMINI_MODEL; defaults to gemini-3-flash-preview
	// GeminiFallbacks is a CSV list of models to try if the primary returns a retryable
	// error (503 UNAVAILABLE, 429 RESOURCE_EXHAUSTED). Tried in order.
	GeminiFallbacks []string // from env GEMINI_MODEL_FALLBACKS
}

func Load() (Config, error) {
	cfg := Config{
		APIAddr:        getenv("API_ADDR", ":8080"),
		DatabaseURL:    getenv("DATABASE_URL", ""),
		AllowedOrigins: splitCSV(getenv("ALLOWED_ORIGINS", "http://localhost:5173,http://127.0.0.1:5173,http://localhost:3000,http://127.0.0.1:3000")),
		KafkaBrokers:   splitCSV(os.Getenv("KAFKA_BROKERS")),
		KafkaTopicJobs: getenv("KAFKA_TOPIC_JOBS", "jobs"),
		// Aiven TLS certificate files — empty by default (TLS opt-in).
		// Relative paths are resolved against the directory of the loaded .env file.
		KafkaCertFile:    resolveFromDotenv(getenv("KAFKA_CERT_FILE", "")),
		KafkaKeyFile:     resolveFromDotenv(getenv("KAFKA_KEY_FILE", "")),
		KafkaCAFile:      resolveFromDotenv(getenv("KAFKA_CA_FILE", "")),
		KafkaConsumerGrp: getenv("KAFKA_CONSUMER_GROUP", "flowbit-workers"),
		ApplyMigrations:  getenvBool("APPLY_MIGRATIONS", true),
		GeminiAPIKey:     getenv("GEMINI_API_KEY", ""),
		GeminiModel:      getenv("GEMINI_MODEL", "gemini-3-flash-preview"),
		GeminiFallbacks:  splitCSV(getenv("GEMINI_MODEL_FALLBACKS", "gemini-flash-latest,gemini-2.5-flash,gemini-2.5-pro,gemini-2.0-flash,gemini-2.0-flash-lite")),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}
	if len(cfg.KafkaBrokers) == 0 {
		return Config{}, errors.New("KAFKA_BROKERS is required")
	}
	return cfg, nil
}

// resolveFromDotenv resolves a relative path against dotenvDir so that cert
// paths in .env work regardless of the working directory at runtime.
func resolveFromDotenv(path string) string {
	if path == "" || filepath.IsAbs(path) || dotenvDir == "" {
		return path
	}
	return filepath.Join(dotenvDir, path)
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
