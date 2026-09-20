// Package worker provides a worker pool for batch inserting telemetry records.
package worker

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/dvandagriff/Carrier-Pigeons/pigeoncoop/internal/storage"
)

// Config holds configuration for the worker pool.
type Config struct {
	WorkerCount   int
	ChannelBuffer int
	BatchSize     int
	FlushInterval time.Duration
}

// WorkerPool manages a pool of workers that process telemetry records.
type WorkerPool struct {
	config      Config
	storage     storage.TelemetryStore
	inputChan   chan TelemetryRecord
	wg          sync.WaitGroup
	stopChan    chan struct{}
	stoppedChan chan struct{}
}

// TelemetryRecord is a record ready for storage.
type TelemetryRecord struct {
	NodeID      string
	Timestamp   time.Time
	MetricType  string
	Value       float64
	Metadata    map[string]string
}

// NewWorkerPool creates a new worker pool with the given configuration.
func NewWorkerPool(config Config, store storage.TelemetryStore) *WorkerPool {
	return &WorkerPool{
		config:      config,
		storage:     store,
		inputChan:   make(chan TelemetryRecord, config.ChannelBuffer),
		stopChan:    make(chan struct{}),
		stoppedChan: make(chan struct{}),
	}
}

// Start begins the worker pool processing.
func (w *WorkerPool) Start(ctx context.Context) error {
	if w.config.WorkerCount == 0 {
		return fmt.Errorf("worker count must be greater than 0")
	}

	for i := 0; i < w.config.WorkerCount; i++ {
		w.wg.Add(1)
		go w.worker(i, ctx)
	}

	// Start flush timer
	w.wg.Add(1)
	go w.flushTimer(ctx)

	return nil
}

// worker processes records from the input channel and batches them for storage.
func (w *WorkerPool) worker(id int, ctx context.Context) {
	defer w.wg.Done()

	slog.Info("worker started", "id", id)

	batch := make([]storage.TelemetryRecord, 0, w.config.BatchSize)
	ticker := time.NewTicker(w.config.FlushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// Process remaining records
			if len(batch) > 0 {
				w.insertBatch(ctx, batch)
			}
			slog.Info("worker stopped", "id", id)
			return

		case <-w.stopChan:
			// Process remaining records
			if len(batch) > 0 {
				w.insertBatch(ctx, batch)
			}
			slog.Info("worker stopped", "id", id)
			close(w.stoppedChan)
			return

		case record, ok := <-w.inputChan:
			if !ok {
				// Channel closed, process remaining records
				if len(batch) > 0 {
					w.insertBatch(ctx, batch)
				}
				slog.Info("worker stopped", "id", id)
				return
			}

			batch = append(batch, storage.TelemetryRecord{
				NodeID:     record.NodeID,
				Timestamp:  record.Timestamp,
				MetricType: record.MetricType,
				Value:      record.Value,
				Metadata:   record.Metadata,
			})

			// Flush if batch is full
			if len(batch) >= w.config.BatchSize {
				w.insertBatch(ctx, batch)
				batch = make([]storage.TelemetryRecord, 0, w.config.BatchSize)
			}

		case <-ticker.C:
			// Flush on timer if batch is not empty
			if len(batch) > 0 {
				w.insertBatch(ctx, batch)
				batch = make([]storage.TelemetryRecord, 0, w.config.BatchSize)
			}
		}
	}
}

// flushTimer sends periodic flush signals.
func (w *WorkerPool) flushTimer(ctx context.Context) {
	defer w.wg.Done()

	ticker := time.NewTicker(w.config.FlushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.stopChan:
			return
		case <-ticker.C:
			// Timer just triggers the worker's flush logic
		}
	}
}

// insertBatch inserts a batch of records into the storage.
func (w *WorkerPool) insertBatch(ctx context.Context, records []storage.TelemetryRecord) {
	if len(records) == 0 {
		return
	}

	err := w.storage.InsertBatch(ctx, records)
	if err != nil {
		slog.Error("failed to insert batch", "error", err.Error())
		// In production, you might want to implement retry logic or dead-letter queue
		return
	}

	slog.Debug("batch inserted", "count", len(records))
}

// Input returns the input channel for sending telemetry records.
func (w *WorkerPool) Input() chan<- TelemetryRecord {
	return w.inputChan
}

// Stop gracefully stops the worker pool.
func (w *WorkerPool) Stop(ctx context.Context) error {
	close(w.stopChan)

	// Close input channel to signal workers
	close(w.inputChan)

	// Wait for all workers to finish
	done := make(chan struct{})
	go func() {
		w.wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("stop timed out: %w", ctx.Err())
	case <-done:
		return nil
	}
}
