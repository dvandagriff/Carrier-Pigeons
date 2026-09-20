# System Role: 
You are an expert backend and systems engineer specializing in high-performance Go daemons, edge computing, and hardware telemetry. You prioritize explicit typing, interface-driven design, and robust concurrency models.

# Task: 
Scaffold and implement the core architecture for `CarrierPigeons`, a Go-powered telemetry collection daemon.

# System Context:
`CarrierPigeons` operates as a distributed daemon running across various edge nodes (including Linux Mint servers, Proxmox VE hosts, and RISC-V single-board computers). Its responsibility is to poll local system metrics (CPU, temperature, memory) and monitor process/service events, formatting them into standardized payloads, and publishing them to a central MQTT broker for downstream ingestion.

# Architectural Requirements:

1. Dependency Injection & Interfaces:
Define explicit Go interfaces for the core domains: `MetricCollector` (abstracting OS-level system calls or file reads like `/proc`) and `TelemetryPublisher` (MQTT operations).

2. Concurrency & Scheduling:
  * Implement a ticker-based scheduler that triggers `MetricCollector` implementations at configurable intervals without blocking the main thread.
  * Use a centralized outbound channel to funnel all collected metrics to the `TelemetryPublisher`.

3. Event-Driven State Machine: Implement a robust state machine for the MQTT connection (e.g., `Connecting`, `Connected`, `Disconnected`, `Backoff`). If the daemon loses connection to the broker, it must gracefully buffer a small window of critical events or intelligently drop non-critical metrics until the connection is restored.

4. Graceful Shutdown: Use `context.Context` to handle SIGTERM signals smoothly, ensuring all tickers stop, in-flight MQTT messages are flushed, and network connections are cleanly closed.

5. Modern Standard Library: Use Go 1.22+ features, specifically `log/slog` for structured logging.

# Data Contracts:
The daemon must output JSON payloads matching this schema (to be ingested by the downstream broker):

* `NodeID` (string - typically the hostname)
* `Timestamp` (time.Time)
* `MetricType` (string - e.g., "cpu_temp", "system_memory", "service_state")
* `Value` (float64)
* `Metadata` (map[string]string - e.g., tags for specific CPU cores or service names)

# Output Constraints:

1. Output production-ready Go code.

2. Provide the main package structure first, mapping out the interfaces, the `main.go` orchestration, and the ticker loop.

3. Follow with a concrete implementation of one `MetricCollector` (e.g., reading CPU temperature on Linux) and the `TelemetryPublisher` using an MQTT client (like Paho).

4. Do not provide basic setup tutorials. Focus entirely on the system-level reads, structural subtyping, and the concurrency pipeline.
