# Study guide

See also: [Vision and demo](00-vision-and-demo.md) · [Architecture](10-architecture.md)

This file holds pre-build learning only; it is not required at runtime.

---

## 6. Concepts to study before building

### Core distributed systems

- Message queues and pub/sub (producer/consumer model)
- At-least-once vs exactly-once delivery
- Idempotency
- Dead-letter queues (DLQ)
- Backpressure
- CAP theorem basics

### Scheduling

- Cron expressions
- Delayed execution and TTL
- Exponential backoff with jitter

### Kafka-specific

- Topics, partitions, consumer groups
- Offsets and commit strategies
- Log compaction

### Reliability and fault tolerance

- Worker heartbeats and leader election
- Distributed locks (Redlock)
- Retry semantics

### Observability

- Metrics vs logs vs traces (the three pillars)
- Prometheus data model: counters, gauges, histograms
- Grafana dashboards

### Go-specific

- Goroutines and channels
- Context cancellation
- `errgroup`

---

## 7. Key reading and resources

| Resource | Why it matters |
| -------- | --------------- |
| Tour of Go ([go.dev/tour](https://go.dev/tour)) | Official interactive Go intro — start here |
| Go by Example ([gobyexample.com](https://gobyexample.com)) | Goroutines, channels, HTTP servers with code |
| Kafka Quickstart ([kafka.apache.org](https://kafka.apache.org)) | Gets Kafka running and understood fast |
| AWS Retry + Backoff blog post | Best practical explanation of backoff with jitter anywhere |
| Gemini API docs ([ai.google.dev](https://ai.google.dev)) | Messages API and structured output for the dispatcher |
| Prometheus Getting Started | Data model and Go client setup |
| Grafana Getting Started | Dashboard creation from Prometheus metrics |
| *Designing Data-Intensive Applications* — Kleppmann | Chapters 1, 3, 11. The single best book on this domain. |
| *The Log* — Jay Kreps (LinkedIn Engineering) | The essay that explains why Kafka exists. Essential. |

---

## 9. Prerequisites to be strong in

- Go basics — goroutines, channels, context, HTTP handlers
- Docker and Docker Compose — running multi-container setups locally
- SQL — basic queries, indexes, transactions
- REST APIs — building and consuming them
- Terminal comfort — logs, process management, env vars

Everything else (Kafka, Prometheus) you learn as you build.
