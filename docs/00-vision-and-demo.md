# Flowbit — Vision and demo

**AI-Powered Distributed Job Scheduler — Build Document**  
Go · Kafka · PostgreSQL · Gemini API · React

See also: [Architecture](10-architecture.md) · [Stack and deployment](20-stack-and-deployment.md) · [Build checklist](BUILD-CHECKLIST.md)

---

## 1. What is this project?

A backend infrastructure system where a user types a plain English command, an AI parses it into a structured job, and a real distributed pipeline processes it reliably — with the entire lifecycle visible in a live visualizer.

This is not a CRUD app. It is the category of tool that powers Celery, Sidekiq, AWS SQS, and Temporal — built from scratch, deployed on real cloud infrastructure, and accessible via a public URL.

**One-line pitch (use once in talks and README):** “I built Flowbit — a distributed task queue from scratch using Go, Kafka, and PostgreSQL — with an AI dispatcher that converts plain English into jobs, and a live pipeline visualizer. It’s deployed. Here’s the link.”

---

## 3. The demo flow

1. User opens the public URL and types: “send an email to john@example.com about the Q3 report”
2. The AI dispatcher (Gemini) parses the input and returns structured JSON: `{ job_type, parameters, priority }`
3. The API server validates the payload, writes it to PostgreSQL with `status: pending`, and publishes to Kafka
4. A worker consumes the job from Kafka and executes the task logic
5. On failure: exponential backoff retry. After 3 attempts: dead-letter queue
6. The React UI shows each stage lighting up via WebSocket in real time
7. Grafana dashboard shows queue depth, throughput, and failure rate

---

## 10. The interview story

This project answers three questions interviewers actually care about:

- Can you build infrastructure, not just features?
- Do you understand how distributed systems fail and recover?
- Can you ship to real production, not just localhost?

The demo is the differentiator. A live URL where someone types a plain English command and watches a distributed pipeline process it in real time is a conversation that interviewers remember.

Use the **same one-line pitch** as in [§1](#1-what-is-this-project) for interviews; optionally add “on real infrastructure” if the conversation needs emphasis.

Most engineers wire up third-party services. You’re building what those services are.
