# PigeonCoop - Go-Powered Telemetry Broker

[![Go Version](https://img.shields.io/badge/Go-1.22+-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

A high-performance, event-driven telemetry broker that ingests MQTT payloads from distributed edge nodes and stores them in PostgreSQL with batched insertions for optimal throughput.

## Architecture Overview

```
┌─────────────----┐     ┌─────────────┐     ┌─────────────┐
│ CarrierPigeons  │────▶│ MQTT Broker │────▶│ PigeonCoop  │
└─────────────----┘     └─────────────┘     └─────────────┘
                                               │
                                               ▼
                                        ┌─────────────┐
                                        │ PostgreSQL  │
                                        └─────────────┘
```

## Core Features

- **Interface-Driven Design**: Explicit interfaces for `PubSubClient` and `TelemetryStore`
- **State Machine**: Connection lifecycle management (Connecting, Connected, Backoff, Disconnected)
- **Worker Pool**: Configurable goroutines for batch processing
- **Buffered Channels**: Non-blocking MQTT handler with configurable queue depth
- **Graceful Shutdown**: Context-based cancellation for clean teardown
- **Structured Logging**: Uses `log/slog` with JSON output

## Package Structure

```
pigeoncoop/
├── cmd/                          # Command-line entry point
│   ├── main.go                   # Application entry
│   ├── config.go                 # Configuration loading
│   └── app.go                    # Main orchestrator
├── internal/
│   ├── pubsub/                   # MQTT client implementation
│   │   ├── interface.go          # PubSubClient interface
│   │   └── mqtt.go               # MQTT client implementation
│   ├── storage/                  # PostgreSQL implementation
│   │   ├── interface.go          # TelemetryStore interface
│   │   └── postgres.go           # PostgreSQL store implementation
│   ├── telemetry/                # Data models
│   │   ├── types.go              # TelemetryPayload struct
│   │   └── json.go               # Custom JSON marshaling
│   ├── worker/                   # Worker pool for batch processing
│   │   └── pool.go               # Worker pool implementation
│   └── state/                    # State machine
│       └── machine.go            # Connection state machine
└── go.mod                        # Go module definition
```

## Data Contracts

### TelemetryPayload

```go
type TelemetryPayload struct {
    NodeID      string            `json:"node_id"`
    Timestamp   time.Time         `json:"timestamp"`
    MetricType  string            `json:"metric_type"`
    Value       float64           `json:"value"`
    Metadata    map[string]string `json:"metadata"`
}
```

### PostgreSQL Schema

```sql
CREATE TABLE telemetry (
    id BIGSERIAL PRIMARY KEY,
    node_id TEXT NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL,
    metric_type TEXT NOT NULL,
    value DOUBLE PRECISION NOT NULL,
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_telemetry_node_id ON telemetry(node_id);
CREATE INDEX idx_telemetry_timestamp ON telemetry(timestamp);
CREATE INDEX idx_telemetry_metric_type ON telemetry(metric_type);
CREATE INDEX idx_telemetry_node_timestamp ON telemetry(node_id, timestamp);
```

## Configuration

Create a `config.yaml` file:

```yaml
mqtt:
  broker_url: "tcp://mqtt-broker:1883"
  client_id: "pigeoncoop-01"
  username: "user"
  password: "password"
  keep_alive: 30
  clean_session: true

storage:
  host: "postgres"
  port: 5432
  user: "pigeoncoop"
  password: "secret"
  database: "telemetry"
  ssl_mode: "disable"
  pool_size: 20

worker:
  worker_count: 8
  channel_buffer: 10000
  batch_size: 100
  flush_interval: 1s

metrics:
  prometheus_port: 9090
```

## Usage

```go
package main

import (
    "log"
    "github.com/dvandagriff/Carrier-Pigeons/pigeoncoop/cmd"
)

func main() {
    // Load configuration
    config, err := cmd.LoadConfig("config.yaml")
    if err != nil {
        log.Fatal(err)
    }

    // Create application
    app := cmd.New(config)

    // Start application
    if err := app.Start(); err != nil {
        log.Fatal(err)
    }

    // Application continues running until SIGTERM/SIGINT
}
```

## Concurrency Model

```
MQTT Message Handler
        │
        ▼
  ┌───────────┐
  │ Buffered  │
  │ Channel   │
  │ (10K buf) │
  └─────┬─────┘
        │
        ▼
┌───────────────────┐
│  Worker Pool (N)  │
│  Goroutines       │
└────────┬──────────┘
         │
         ▼
   ┌───────────┐
   │ Batch     │
   │ Insert    │
   │ (pgx)     │
   └───────────┘
```

## State Machine

```
           ┌─────────────┐
           │Disconnected│
           └──────┬──────┘
                  │ Connect()
                  ▼
           ┌─────────────┐
           │Connecting  │
           └──────┬──────┘
                  │
        ┌─────────┴─────────┐
        │                   │
     Connect()          Backoff()
        ▼                   ▼
  ┌─────────────┐    ┌─────────────┐
  │ Connected   │◀───│  Backoff    │
  └──────┬──────┘    └─────────────┘
         │                   │
    Disconnect()          Timeout()
         │                   ▼
         │          ┌─────────────┐
         └─────────▶│Connecting  │
                    └─────────────┘

Shutdown Sequence:
  Connected ▶ Stopping ▶ Stopped
```

## Performance Considerations

1. **Batch Inserts**: Uses pgx batch operations for efficient database writes
2. **Worker Pool**: Configurable concurrency for optimal throughput
3. **Buffered Channels**: Prevents blocking MQTT handlers
4. **Connection Pooling**: pgx connection pooling with health checks
5. **Non-blocking Operations**: MQTT handlers return immediately after queuing

## Graceful Shutdown

The application handles SIGTERM/SIGINT signals to ensure:
1. MQTT subscriptions are unsubscribed
2. Worker pool drains remaining records
3. Database connections are closed gracefully
4. State machine transitions to Stopped state

## License

MIT

## Contributing

1. Fork the repository
2. Create a feature branch
3. Run `go test ./...`
4. Submit a pull request
