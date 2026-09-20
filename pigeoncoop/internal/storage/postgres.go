// Package storage provides PostgreSQL implementation of the TelemetryStore interface.
package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresStore implements TelemetryStore using pgx connection pooling.
type PostgresStore struct {
	config *pgxpool.Config
	pool   *pgxpool.Pool
}

// NewPostgresStore creates a new PostgreSQL store with the given config.
func NewPostgresStore(cfg *StoreConfig) (*PostgresStore, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.SSLMode)

	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	config.MaxConns = int32(cfg.PoolSize)
	config.MinConns = 1
	config.MaxConnLifetime = 30 * time.Minute
	config.MaxConnIdleTime = 5 * time.Minute
	config.HealthCheckPeriod = 10 * time.Second

	return &PostgresStore{config: config}, nil
}

// Init initializes the database connection and performs migrations.
func (s *PostgresStore) Init(ctx context.Context) error {
	pool, err := pgxpool.NewWithConfig(ctx, s.config)
	if err != nil {
		return fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test the connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Perform migrations
	if err := s.migrate(ctx, pool); err != nil {
		pool.Close()
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	s.pool = pool
	return nil
}

func (s *PostgresStore) migrate(ctx context.Context, pool *pgxpool.Pool) error {
	schema := `
	CREATE TABLE IF NOT EXISTS telemetry (
		id BIGSERIAL PRIMARY KEY,
		node_id TEXT NOT NULL,
		timestamp TIMESTAMPTZ NOT NULL,
		metric_type TEXT NOT NULL,
		value DOUBLE PRECISION NOT NULL,
		metadata JSONB NOT NULL DEFAULT '{}',
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_telemetry_node_id ON telemetry(node_id);
	CREATE INDEX IF NOT EXISTS idx_telemetry_timestamp ON telemetry(timestamp);
	CREATE INDEX IF NOT EXISTS idx_telemetry_metric_type ON telemetry(metric_type);
	CREATE INDEX IF NOT EXISTS idx_telemetry_node_timestamp ON telemetry(node_id, timestamp);
	`

	_, err := pool.Exec(ctx, schema)
	if err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}

	return nil
}

// Close gracefully terminates the database connection pool.
func (s *PostgresStore) Close(ctx context.Context) error {
	if s.pool == nil {
		return nil
	}
	s.pool.Close()
	return nil
}

// InsertBatch inserts multiple telemetry records in a batch operation.
func (s *PostgresStore) InsertBatch(ctx context.Context, payloads []TelemetryRecord) error {
	if len(payloads) == 0 {
		return nil
	}

	if s.pool == nil {
		return fmt.Errorf("store not initialized")
	}

	// Marshal metadata to JSON
	batch := &pgx.Batch{}
	
	for _, p := range payloads {
		metadataJSON, err := json.Marshal(p.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}

		batch.Queue(`
			INSERT INTO telemetry (node_id, timestamp, metric_type, value, metadata)
			VALUES ($1, $2, $3, $4, $5)
		`, p.NodeID, p.Timestamp, p.MetricType, p.Value, metadataJSON)
	}

	br := s.pool.SendBatch(ctx, batch)
	defer br.Close()

	for range payloads {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("failed to insert record: %w", err)
		}
	}

	return nil
}

// InsertSingle inserts a single telemetry record.
func (s *PostgresStore) InsertSingle(ctx context.Context, record TelemetryRecord) error {
	return s.InsertBatch(ctx, []TelemetryRecord{record})
}

// HealthCheck returns true if the store is reachable.
func (s *PostgresStore) HealthCheck(ctx context.Context) error {
	if s.pool == nil {
		return fmt.Errorf("store not initialized")
	}
	return s.pool.Ping(ctx)
}
