package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"log"
	"os"
	"time"

	"github.com/segmentio/kafka-go"
)

func main() {
	TOPIC_NAME := "jobs"

	// Load client certificate (service.cert + service.key)
	keypair, err := tls.LoadX509KeyPair("service.cert", "service.key")
	if err != nil {
		log.Fatalf("Failed to load access key and/or access certificate: %s", err)
	}

	// Load CA certificate (ca.pem)
	caCert, err := os.ReadFile("ca.pem")
	if err != nil {
		log.Fatalf("Failed to read CA certificate file: %s", err)
	}

	caCertPool := x509.NewCertPool()
	ok := caCertPool.AppendCertsFromPEM(caCert)
	if !ok {
		log.Fatalf("Failed to parse CA certificate file")
	}

	dialer := &kafka.Dialer{
		Timeout:   10 * time.Second,
		DualStack: true,
		TLS: &tls.Config{
			Certificates: []tls.Certificate{keypair},
			RootCAs:      caCertPool,
			MinVersion:   tls.VersionTLS12,
		},
	}

	// Init consumer
	consumer := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"flowbit-aiven-kafka-flowbit.b.aivencloud.com:26159"},
		Topic:   TOPIC_NAME,
		Dialer:  dialer,
	})

	log.Printf("Consumer started, waiting for messages on topic %q...", TOPIC_NAME)

	for {
		message, err := consumer.ReadMessage(context.Background())

		if err != nil {
			log.Printf("Could not read message: %s", err)
			time.Sleep(time.Second)
			continue
		}

		log.Printf("Got message using SSL: %s", message.Value)
	}
}
