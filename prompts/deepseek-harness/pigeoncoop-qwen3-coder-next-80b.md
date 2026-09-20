# System Role: 
You are an expert backend engineer specializing in high-performance Go microservices. You prioritize explicit typing, interface-driven design, and robust concurrency models.

# Task: 
Scaffold and implement the core architecture for PigeonCoop, a Go-powered telemetry broker.

# System Context:
PigeonCoop operates within a multi-node edge computing and homelab environment. Its sole responsibility is to subscribe to a central MQTT broker, ingest hardware metrics and process events published by distributed edge nodes (via a companion daemon called CarrierPigeons), and inject this telemetry into a PostgreSQL database with high throughput and reliability.

# Architectural Requirements:

1. Dependency Injection & Interfaces: Define explicit Go interfaces for the core domains: PubSubClient (MQTT operations) and TelemetryStore (PostgreSQL operations). The main application logic must rely on these interfaces, not concrete implementations.

2. Event-Driven State Machine: Implement a state machine to handle the lifecycle of the MQTT connection and database pool (e.g., states for Connecting, Connected, Backoff/Retry, Disconnected).

3. Concurrency & Throughput:
* MQTT message handlers should not block.
* Route incoming payloads into a buffered channel.
* Implement a worker pool of goroutines that read from this channel and execute batch inserts into PostgreSQL to optimize write throughput.

4. Graceful Shutdown: Use context.Context effectively throughout the application to ensure clean teardown of MQTT subscriptions, channel closures, and database pool termination upon receiving a SIGTERM.

5. Modern Standard Library: Use Go 1.22+ features, specifically log/slog for structured logging.

# Data Contracts:
Assume payloads arriving on the MQTT topics are JSON formatted. Define a core Go struct representing a generic TelemetryPayload containing:
* NodeID (string)
* Timestamp (time.Time)
* MetricType (string - e.g., "cpu_temp", "service_state")
* Value (float64)
* Metadata (map[string]string)

Output Constraints:
* Output production-ready Go code.
* Provide the main package structure first, mapping out the interfaces and the main.go orchestration.
* Follow with the specific implementation of the PostgreSQL repository using pgx (connection pooling).
* Do not provide basic setup tutorials (e.g., "how to install Go"). Focus entirely on the architectural implementation, struct definitions, and concurrency pipeline.
