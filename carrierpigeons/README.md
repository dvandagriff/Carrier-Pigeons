# CarrierPigeons

[![Go Version](https://img.shields.io/badge/Go-1.22+-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

A high-performance, edge-targeted telemetry collection daemon that gathers system metrics (CPU temperature, memory usage) and publishes them to an MQTT broker for centralized ingestion.

## Overview

`CarrierPigeons` operates as a distributed daemon running on edge nodes (Linux Mint servers, Proxmox VE hosts, RISC-V single-board computers). It collects local system metrics at configured intervals and publishes standardized JSON payloads to an MQTT topic.

```
┌─────────────────┐     ┌─────────────┐     ┌─────────────┐
│  Edge Node 1    │────▶│  MQTT       │────▶│  PigeonCoop │
│  (CarrierPigeon)│     │  Broker     │     │  (Ingestion)│
└─────────────────┘     └─────────────┘     └─────────────┘
┌─────────────────┐            |
│  Edge Node 2    │────--------+
│  (CarrierPigeon)│     
└─────────────────┘     
```

## Features

- **Interface-Driven Design**: Explicit interfaces for `MetricCollector` and `Publisher`
- **Ticker-Based Scheduler**: Configurable collection intervals without blocking
- **Event-Driven State Machine**: Robust MQTT connection lifecycle (Connecting, Connected, Backoff)
- **Graceful Shutdown**: Context-based cancellation with buffered metric flushing
- **Structured Logging**: Uses `log/slog` with JSON output
- **Cross-Platform Collectors**: Linux-specific implementations for CPU/memory metrics

## Data Contracts

### Telemetry Payload

```go
type TelemetryPayload struct {
    NodeID      string            `json:"node_id"`
    Timestamp   string            `json:"timestamp"`
    MetricType  string            `json:"metric_type"`
    Value       float64           `json:"value"`
    Metadata    map[string]string `json:"metadata"`
}
```

Example JSON payload:

```json
{
  "node_id": "edge-node-01",
  "timestamp": "2024-01-15T10:30:45.123456789Z",
  "metric_type": "cpu_temp",
  "value": 65.5,
  "metadata": {
    "node_id": "edge-node-01",
    "zone": "thermal_zone0",
    "zone_type": "platform"
  }
}
```

### Metric Types

- `cpu_temp` - CPU temperature in Celsius
- `cpu_usage` - CPU frequency in MHz
- `memory_usage` - Memory usage percentage
- `memory_available` - Available memory percentage

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Main (main.go)                          │
│  ┌─────────────┐  ┌─────────────┐  ┌──────────────────┐    │
│  │   State     │  │  Collector  │  │     Scheduler    │    │
│  │  Machine    │  │   Pool      │  │   Ticker Loop    │    │
│  └─────────────┘  └─────────────┘  └──────────────────┘    │
│                                      │                      │
│                               ┌────▼────┐                  │
│                               │Channel  │                  │
│                               │(Buffered)│                 │
│                               └────┬────┘                  │
│                                      │                      │
│                               ┌────▼────┐                  │
│                               │Publisher│                  │
│                               │ (MQTT)  │                  │
│                               └─────────┘                  │
└─────────────────────────────────────────────────────────────┘
```

### Components

1. **State Machine** (`internal/state/machine.go`)
   - Manages MQTT connection lifecycle
   - Handles connection failures with exponential backoff

2. **Scheduler** (`internal/scheduler/scheduler.go`)
   - Ticker-based metric collection
   - Non-blocking channel-based dispatch

3. **Collectors** (`internal/collector/`)
   - `cpu_linux.go` - CPU temperature and frequency
   - `memory_linux.go` - Memory usage statistics

4. **Publisher** (`internal/publisher/mqtt.go`)
   - MQTT client using paho.golang
   - Automatic reconnection support

## Installation

```bash
cd carrierpigeons

# Build
go build -o carrierpigeons ./cmd/main.go

# Or cross-compile for target platform
GOOS=linux GOARCH=arm64 go build -o carrierpigeons-arm64 ./cmd/main.go
```

## Configuration

Create a `config.yaml` file:

```yaml
node_id: "edge-node-01"

metrics:
  enabled_collectors:
    - cpu
    - memory

publisher:
  broker_url: "tcp://mqtt-broker:1883"
  client_id: "carrier-pigeon-01"
  username: ""
  password: ""
  topic: "telemetry"
  keep_alive: 30
  clean_session: true

scheduler:
  metric_channel_size: 1000
  outbound_channel_size: 10000
```

## Usage

```bash
# Run with default config
./carrierpigeons

# Specify custom config path
./carrierpigeons -config /etc/carrierpigeons/config.yaml
```

## Signal Handling

- `SIGTERM` - Graceful shutdown with buffered metric flush
- `SIGINT` - Same as SIGTERM (Ctrl+C)

## Performance Considerations

1. **Buffered Channels**: Non-blocking metric collection
2. **Ticker Efficiency**: Single goroutine per collector
3. **Connection Pooling**: Reuses MQTT connections
4. **Backpressure Control**: Configurable channel sizes prevent memory exhaustion

## State Machine Transitions

```
        ┌─────────────┐
        │DISCONNECTED │
        └──────┬──────┘
               │ Start()
               ▼
        ┌─────────────┐
        │ CONNECTING  │
        └──────┬──────┘
               │
        ┌─────────┴─────────┐
        │                   │
     Connect()          Backoff()
        ▼                   ▼
  ┌─────────────┐    ┌─────────────┐
  │ CONNECTED   │◀───│   BACKOFF   │
  └──────┬──────┘    └─────────────┘
         │                   │
    Disconnect()         Timeout()
         │                   ▼
         │          ┌─────────────┐
         └─────────▶│ CONNECTING  │
                    └─────────────┘
```

## License

MIT

## Contributing

1. Fork the repository
2. Create a feature branch
3. Run `go test ./...`
4. Submit a pull request
