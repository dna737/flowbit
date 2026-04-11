package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/segmentio/kafka-go"
)

func main() {
	broker := os.Getenv("KAFKA_BROKERS")
	if broker == "" {
		log.Fatal("KAFKA_BROKERS env var required (e.g. your-service.aivencloud.com:PORT)")
	}

	topicName := "jobs"

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
	if !caCertPool.AppendCertsFromPEM(caCert) {
		log.Fatalf("Failed to parse CA certificate file")
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{keypair},
		RootCAs:      caCertPool,
		MinVersion:   tls.VersionTLS12,
	}

	// Init producer
	producer := &kafka.Writer{
		Addr:      kafka.TCP(broker),
		Topic:     topicName,
		Transport: &kafka.Transport{TLS: tlsConfig},
	}
	defer producer.Close()

	// Produce 100 messages
	for i := 0; i < 100; i++ {
		message := fmt.Sprint("Hello from Go using SSL ", i+1, "!")
		err := producer.WriteMessages(context.Background(), kafka.Message{Value: []byte(message)})
		if err != nil {
			log.Printf("Failed to send message: %s", err)
		} else {
			log.Printf("Message sent: %s", message)
		}
		time.Sleep(time.Second)
	}
}
