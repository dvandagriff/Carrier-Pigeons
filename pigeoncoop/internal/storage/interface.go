// Package storage defines the interface for telemetry storage operations (PostgreSQL).
package storage

import (
	"context"
	"time"
)

// TelemetryStore defines the interface for telemetry data storage operations.
type TelemetryStore interface {
	// Init initializes the database connection and performs migrations.
	Init(ctx context.Context) error

	// Close gracefully terminates the database connection pool.
	Close(ctx context.Context) error

	// InsertBatch inserts multiple telemetry records in a batch operation.
	InsertBatch(ctx context.Context, payloads []TelemetryRecord) error

	// InsertSingle inserts a single telemetry record.
	InsertSingle(ctx context.Context, record TelemetryRecord) error

	// HealthCheck returns true if the store is reachable.
	HealthCheck(ctx context.Context) error
}

// TelemetryRecord represents a database record for telemetry data.
type TelemetryRecord struct {
	NodeID      string
	Timestamp   time.Time
	MetricType  string
	Value       float64
	Metadata    map[string]string
}

// StoreConfig holds configuration for creating a TelemetryStore.
type StoreConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string
	PoolSize int
}
