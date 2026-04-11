# Kafka Examples for Aiven

Standalone producer and consumer examples for connecting to Aiven Kafka using TLS certificate authentication.

## Prerequisites

1. Download your TLS certificates from Aiven Console:
   - `service.cert` - Client certificate
   - `service.key` - Client key
   - `ca.pem` - CA certificate

2. Place these files in this directory (`examples/kafka/`)

## Usage

### Producer

```powershell
cd examples/kafka
# Initialize go modules if needed
go mod tidy
# Run producer
go run producer/main.go
```

The producer will send 100 messages to the `jobs` topic.

### Consumer

```powershell
cd examples/kafka
go run consumer/main.go
```

The consumer will continuously read messages from the `jobs` topic.

## Configuration

Update the broker address in the source files if your Aiven service has a different endpoint:

```go
Brokers: []string{"your-service.aivencloud.com:26159"},
```

## How It Works

Aiven Kafka uses mutual TLS (mTLS) authentication instead of username/password. The examples demonstrate:

1. Loading the client certificate and key pair
2. Loading the CA certificate to verify the server
3. Configuring TLS with both certificates for mutual authentication
4. Creating a Kafka dialer with the TLS configuration
5. Using the dialer with the segmentio/kafka-go writer/reader
