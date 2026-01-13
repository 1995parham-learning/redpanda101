<h1 align="center">Redpanda 101</h1>
<p align="center">
  <img alt="GitHub Actions Workflow Status" src="https://img.shields.io/github/actions/workflow/status/1995parham-learning/redpanda101/test.yaml?style=for-the-badge&logo=github">
</p>

## Introduction

A demonstration project exploring [Redpanda](https://redpanda.com/) as a lightweight Kafka alternative with Go. This project implements an event-driven order management system using the producer-consumer pattern.

For Kafka/Redpanda integration in Go, this project uses [franz-go](https://github.com/twmb/franz-go). Other alternatives include [Confluent Kafka](https://github.com/confluentinc/confluent-kafka-go) and [Sarama](https://github.com/IBM/sarama).

## Architecture

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   HTTP Client   │────▶│    Producer     │────▶│    Redpanda     │
│  (k6 / curl)    │     │  POST /orders/  │     │  (orders topic) │
└─────────────────┘     └─────────────────┘     └────────┬────────┘
                                                         │
                                                         ▼
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   PostgreSQL    │◀────│    Consumer     │◀────│  Worker Pool    │
│   (orders)      │     │  (30 workers)   │     │  (concurrent)   │
└─────────────────┘     └─────────────────┘     └─────────────────┘
```

## Features

- **Producer**: HTTP API that accepts orders and publishes them to Redpanda
- **Consumer**: Worker pool that processes messages and persists to PostgreSQL
- **Observability**: OpenTelemetry tracing (Jaeger) + Prometheus metrics + Grafana dashboards
- **Resilience**: Retry logic with exponential backoff for database operations
- **Configuration**: TOML files with environment variable overrides

## Prerequisites

- Go 1.25+
- Docker & Docker Compose
- [k6](https://k6.io/) (for load testing)
- [just](https://github.com/casey/just) (optional, for task automation)

## Quick Start

### 1. Start Infrastructure

```bash
cd deployments
docker compose up -d
```

### 2. Create Topic

Open Redpanda Console at http://127.0.0.1:8080 and create the `orders` topic.

### 3. Run Database Migrations

```bash
go build -o redpanda101 ./cmd/redpanda101
./redpanda101 migrate
```

### 4. Start Producer & Consumer

```bash
# Terminal 1: Producer (HTTP API on port 1378)
./redpanda101 -c configs/producer.toml produce

# Terminal 2: Consumer
./redpanda101 -c configs/consumer.toml consume
```

### 5. Test the API

```bash
curl -X POST http://127.0.0.1:1378/orders/ \
  -H "Content-Type: application/json" \
  -d '{"src_currency": 1, "dst_currency": 2, "description": "test order", "channel": 1}'
```

Or use the requests in `demo.http` with your REST client.

## Configuration

Configuration is loaded in order (later sources override earlier):

1. Default values
2. TOML config file (`-c` flag)
3. Environment variables (prefix: `redpanda101_`, nested with `__`)

Example environment override:

```bash
export redpanda101_kafka__seeds="kafka1:9092,kafka2:9092"
export redpanda101_consumer__concurrency=50
```

### Key Configuration Options

| Option | Default | Description |
|--------|---------|-------------|
| `kafka.seeds` | `127.0.0.1:19092` | Redpanda/Kafka broker addresses |
| `kafka.consumer_group` | `koochooloo-group` | Consumer group name |
| `consumer.concurrency` | `30` | Number of worker goroutines |
| `database.url` | `postgres://...` | PostgreSQL connection string |
| `telemetry.trace.enabled` | `true` | Enable Jaeger tracing |
| `telemetry.meter.enabled` | `true` | Enable Prometheus metrics |

## Monitoring

| Service | URL | Credentials |
|---------|-----|-------------|
| Redpanda Console | http://127.0.0.1:8080 | - |
| Prometheus | http://127.0.0.1:9090 | - |
| Jaeger | http://127.0.0.1:16686 | - |
| Grafana | http://127.0.0.1:3000 | `parham.alvani@gmail.com` / `P@ssw0rd` |

### Available Metrics

**Producer:**
- `messages_produced_total` - Total messages published
- `produce_latency_seconds` - Message publish latency
- `produce_errors_total` - Publish error count

**Consumer:**
- `message_delay_seconds` - Time between publish and consume
- `database_insertion_time_seconds` - PostgreSQL insert latency

## Load Testing

Run the k6 load test:

```bash
k6 run api/k6/script.js
```

### Sample Results

| Metric | Value |
|--------|-------|
| Total Requests | 606,848 |
| Throughput | ~5,059 req/s |
| Avg Latency | 3.4ms |
| P95 Latency | 5.68ms |
| Error Rate | 0.00% |

<details>
<summary>Full k6 Output</summary>

```
     scenarios: (100.00%) 1 scenario, 35 max VUs, 2m30s max duration

     █ publish

       ✓ success

     checks.........................: 100.00% 606848 out of 606848
     http_req_duration..............: avg=3.4ms   min=8µs   med=3.29ms max=87.75ms p(90)=4.87ms p(95)=5.68ms
     http_req_failed................: 0.00%   0 out of 606848
     http_reqs......................: 606848  5059.115018/s
     vus............................: 34      min=1  max=34

running (2m00.0s), 00/35 VUs, 606848 complete and 0 interrupted iterations
```

</details>

## Project Structure

```
.
├── cmd/redpanda101/          # Application entry point
├── internal/
│   ├── cmd/                  # CLI commands (produce, consume, migrate)
│   ├── domain/model/         # Domain models (Order, Channel)
│   └── infra/                # Infrastructure layer
│       ├── config/           # Configuration loading
│       ├── consumer/         # Kafka consumer + metrics
│       ├── producer/         # Kafka producer + metrics
│       ├── database/         # PostgreSQL connection
│       ├── http/             # HTTP server + controllers
│       ├── kafka/            # Kafka client setup
│       ├── telemetry/        # OpenTelemetry + Prometheus
│       └── logger/           # Zap logger setup
├── configs/                  # TOML configuration files
├── deployments/              # Docker Compose + Grafana dashboards
├── migrations/               # Database migrations
└── api/k6/                   # Load testing scripts
```

## Development

```bash
# Build
just build

# Lint
just lint

# Update dependencies
just update

# Start/stop infrastructure
just dev up
just dev down

# Create new migration
just migrate create <name>
```

## Design Decisions

- **Redpanda over Kafka**: Easier to run locally with Docker, fully Kafka-compatible
- **franz-go**: Modern, performant Kafka client with excellent Redpanda support
- **Worker Pool Pattern**: Configurable concurrency for message processing
- **Retry with Backoff**: Exponential backoff (100ms, 200ms, 400ms) for transient DB failures
- **UUID for Order IDs**: Guaranteed uniqueness without coordination
