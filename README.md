<h1 align="center">Redpanda 101</h1>
<p align="center">
  <img alt="GitHub Actions Workflow Status" src="https://img.shields.io/github/actions/workflow/status/1995parham-learning/redpanda101/test.yaml?style=for-the-badge&logo=github">
</p>

## Introduction

A demonstration project exploring [Redpanda](https://redpanda.com/) as a lightweight Kafka alternative with Go. This project implements an event-driven order management and **matching** system: orders are published to Redpanda, and a matching engine materialises an in-memory limit order book straight from the log to cross orders into trades.

For Kafka/Redpanda integration in Go, this project uses [franz-go](https://github.com/twmb/franz-go). Other alternatives include [Confluent Kafka](https://github.com/confluentinc/confluent-kafka-go) and [Sarama](https://github.com/IBM/sarama).

## Architecture

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────────┐
│   HTTP Client   │────▶│    Producer     │────▶│      Redpanda       │
│  (k6 / curl)    │     │  POST /orders/  │     │  orders (by symbol) │
└─────────────────┘     └─────────────────┘     └─────┬───────────┬───┘
                                                      │           │
                          consumer (history)         ▼           ▼      matcher (single writer)
                  ┌─────────────────┐     ┌─────────────────┐  ┌──────────────────────┐
                  │   PostgreSQL    │◀────│    Consumer     │  │       Matcher        │
                  │  orders/trades  │     │  (30 workers)   │  │  in-memory CLOB per  │
                  └────────▲────────┘     └─────────────────┘  │  symbol (price-time) │
                           │                                   └─────────┬────────────┘
                           │ trades                                      │ trades
                           └─────────────────────────────────◀──────────┘
                                                                         ▼
                                                            ┌─────────────────────┐
                                                            │      Redpanda       │
                                                            │    trades topic     │
                                                            └─────────────────────┘
```

The **consumer** persists raw order history (worker pool, order-independent). The
**matcher** is a separate single-writer service: it replays the `orders` log from
the start to rebuild every order book in memory, then continuously crosses new
orders into trades, publishing them to the `trades` topic and PostgreSQL. Orders
are keyed by **symbol** (`srcCurrency-dstCurrency`) so each market is totally
ordered on its partition — the property a matching engine depends on.

## Features

- **Producer**: HTTP API that accepts orders and publishes them to Redpanda, keyed by symbol
- **Matcher**: Single-writer matching engine with a price-time-priority limit order book held in memory and materialised from the Redpanda log; executes at the maker price and emits trades
- **Consumer**: Worker pool that persists order history to PostgreSQL
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

### 2. Create Topics

Open Redpanda Console at http://127.0.0.1:8085 and create the `orders` and
`trades` topics (or use `just redpanda-migrate`).

### 3. Run Database Migrations

```bash
go build -o redpanda101 ./cmd/redpanda101
./redpanda101 migrate
```

### 4. Start Producer, Matcher & Consumer

```bash
# Terminal 1: Producer (HTTP API on port 1378)
./redpanda101 -c configs/producer.toml produce

# Terminal 2: Matcher (in-memory order book → trades)
./redpanda101 -c configs/consumer.toml match

# Terminal 3: Consumer (order history → PostgreSQL, optional)
./redpanda101 -c configs/consumer.toml consume
```

### 5. Test the API

Submit a resting sell, then a buy that crosses it:

```bash
# A sell of 5 units at price 100 (rests in the book)
curl -X POST http://127.0.0.1:1378/orders/ \
  -H "Content-Type: application/json" \
  -d '{"src_currency": 1, "dst_currency": 2, "side": "sell", "price": 100, "quantity": 5, "channel": 1}'

# A buy of 5 units at price 105 → matches at the maker price of 100
curl -X POST http://127.0.0.1:1378/orders/ \
  -H "Content-Type: application/json" \
  -d '{"src_currency": 1, "dst_currency": 2, "side": "buy", "price": 105, "quantity": 5, "channel": 1}'
```

The matcher logs each fill and the resulting top-of-book, and publishes a trade
to the `trades` topic. Required order fields: `side` (`buy`/`sell`), `price`
(> 0), `quantity` (> 0), and a valid `channel`.

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
| Redpanda Console | http://127.0.0.1:8085 | - |
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

**Matcher:**
- `orders_matched_total` - Orders processed by the engine
- `trades_produced_total` - Trades generated from crossings
- `match_latency_seconds` - Time to match a single order

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
│   ├── cmd/                  # CLI commands (produce, consume, match, migrate)
│   ├── domain/
│   │   ├── model/            # Domain models (Order, Trade, Side, Channel)
│   │   └── orderbook/        # Pure CLOB matching engine (price-time priority) + tests
│   └── infra/                # Infrastructure layer
│       ├── config/           # Configuration loading
│       ├── consumer/         # Kafka consumer + metrics
│       ├── producer/         # Kafka producer + metrics
│       ├── matcher/          # Matching service (orders → engine → trades)
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

## Matching Engine

The order book lives in `internal/domain/orderbook` as pure, deterministic
domain logic (no I/O, clocks, or randomness), which makes it exhaustively
unit-testable:

- **Continuous limit order book (CLOB)** per symbol with **price-time priority**:
  best price first, ties broken by arrival order (FIFO).
- **Maker-price execution**: a crossing taker fills at the resting order's price,
  giving the taker any price improvement.
- **Engine** routes each order to its symbol's book, lazily creating markets.

The `matcher` service is the in-memory store made real. It is a **group-less
(direct) consumer** that reads every partition of `orders` from offset 0 on every
boot, processing orders **one at a time** (never on a worker pool) so the book
observes them in log order. There are no committed offsets to resume from — the
book is purely a materialised view of the log, so a restart rebuilds it exactly.

To make that full-replay safe, trade side effects are **idempotent**: each trade
gets a deterministic UUID derived from its unique `(buy, sell)` order pair and a
timestamp taken from the order record (not the wall clock). Re-running the log
therefore produces byte-identical trades — `ON CONFLICT (id) DO NOTHING` dedupes
them in PostgreSQL, and downstream consumers can dedupe the `trades` topic by id.

> **Single-writer model:** a direct consumer reads *all* partitions itself, so run
> exactly **one** `match` instance — it is the sole writer for the whole topic.
> The trade-off is replay cost (every boot reprocesses the full log and re-emits
> historical trades). Scaling to one matcher per market — with snapshots so a
> reassigned book recovers without a full replay — is the production follow-up.

## Design Decisions

- **Redpanda over Kafka**: Easier to run locally with Docker, fully Kafka-compatible
- **franz-go**: Modern, performant Kafka client with excellent Redpanda support
- **Symbol-keyed orders**: Each market is totally ordered on one partition, which the matcher relies on
- **Integer prices/quantities**: Avoids floating-point rounding bugs in money math
- **Worker Pool Pattern**: Configurable concurrency for order *history* (consumer); the matcher is deliberately sequential
- **Retry with Backoff**: Exponential backoff (100ms, 200ms, 400ms) for transient DB failures
- **UUID for Order/Trade IDs**: Guaranteed uniqueness without coordination
