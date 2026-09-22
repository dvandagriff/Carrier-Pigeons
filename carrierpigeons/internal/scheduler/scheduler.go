// Package scheduler provides the ticker-based scheduler for metric collection.
package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/dvandagriff/Carrier-Pigeons/carrierpigeons/internal/collector"
	"github.com/dvandagriff/Carrier-Pigeons/carrierpigeons/internal/publisher"
)

// Scheduler manages the collection and publishing of metrics.
type Scheduler struct {
	collectors   []collector.MetricCollector
	publisher    publisher.Publisher
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	metricChan   chan collector.Metric
	outboundChan chan collector.Metric

	// Metrics for monitoring
	totalMetricsCollected uint64
	totalMetricsPublished uint64
	totalMetricsDropped   uint64
	metricsMu            sync.Mutex

	// Configuration
	maxBatchSize int
}

// Config holds configuration for the scheduler.
type Config struct {
	MetricChannelSize   int
	OutboundChannelSize int
	MaxBatchSize        int
}

// NewScheduler creates a new scheduler with the given collectors and publisher.
func NewScheduler(
	collectors []collector.MetricCollector,
	publisher publisher.Publisher,
	config *Config,
) *Scheduler {
	if config == nil {
		config = &Config{
			MetricChannelSize:   1000,
			OutboundChannelSize: 10000,
			MaxBatchSize:        100,
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &Scheduler{
		collectors:    collectors,
		publisher:     publisher,
		ctx:           ctx,
		cancel:        cancel,
		metricChan:    make(chan collector.Metric, config.MetricChannelSize),
		outboundChan:  make(chan collector.Metric, config.OutboundChannelSize),
		maxBatchSize:  config.MaxBatchSize,
	}
}

// Start begins the scheduler's collection and publishing loop.
func (s *Scheduler) Start() error {
	slog.Info("starting scheduler", "collector_count", len(s.collectors))

	// Start collector goroutines
	for _, c := range s.collectors {
		s.wg.Add(1)
		go s.collectMetrics(c)
	}

	// Start publisher goroutine
	s.wg.Add(1)
	go s.publishMetrics()

	slog.Info("scheduler started")
	return nil
}

// Stop gracefully shuts down the scheduler.
func (s *Scheduler) Stop() {
	slog.Info("stopping scheduler")

	// Cancel context to stop collection
	s.cancel()

	// Wait for all collectors to finish
	s.wg.Wait()

	slog.Info("scheduler stopped")
}

// collectMetrics runs the collection loop for a single collector.
func (s *Scheduler) collectMetrics(c collector.MetricCollector) {
	defer s.wg.Done()

	slog.Info("collector started", "collector", c.Name())

	ticker := time.NewTicker(c.Interval())
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			slog.Info("collector stopped", "collector", c.Name())
			return

		case <-ticker.C:
			s.collectOnce(c)
		}
	}
}

// collectOnce performs a single collection cycle.
func (s *Scheduler) collectOnce(c collector.MetricCollector) {
	ctx, cancel := context.WithTimeout(s.ctx, 5*time.Second)
	defer cancel()

	metrics, err := c.Collect(ctx)
	if err != nil {
		slog.Error("failed to collect metrics", "collector", c.Name(), "error", err.Error())
		return
	}

	slog.Debug("collected metrics", "collector", c.Name(), "count", len(metrics))

	// Send metrics to outbound channel with backpressure handling
	for _, metric := range metrics {
		select {
		case s.outboundChan <- metric:
			// Metric queued successfully
			s.metricsMu.Lock()
			s.totalMetricsCollected++
			s.metricsMu.Unlock()
		case <-s.ctx.Done():
			// Scheduler stopped, drop metric
			slog.Debug("scheduler stopped during metric collection")
			return
		default:
			// Channel full, drop metric with backpressure logging
			s.metricsMu.Lock()
			s.totalMetricsDropped++
			droppedCount := s.totalMetricsDropped
			s.metricsMu.Unlock()
			
			if droppedCount%100 == 0 {
				slog.Warn("backpressure: dropping metrics due to channel full",
					"collector", c.Name(),
					"dropped_total", droppedCount)
			}
		}
	}
}

// publishMetrics reads metrics from the outbound channel and publishes them with batching.
func (s *Scheduler) publishMetrics() {
	defer s.wg.Done()

	// Wait for publisher to be ready
	ready := s.publisher.ConnectionReady()
	select {
	case <-ready:
	case <-s.ctx.Done():
		return
	}

	batch := make([]collector.Metric, 0, s.maxBatchSize)
	ticker := time.NewTicker(100 * time.Millisecond) // Flush interval
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			// Final flush of remaining metrics
			s.flushBatch(batch)
			slog.Info("publisher stopped")
			return

		case metric, ok := <-s.outboundChan:
			if !ok {
				// Channel closed, flush remaining batch
				s.flushBatch(batch)
				return
			}

			batch = append(batch, metric)

			// Flush when batch is full
			if len(batch) >= s.maxBatchSize {
				s.flushBatch(batch)
				batch = make([]collector.Metric, 0, s.maxBatchSize)
			}

		case <-ticker.C:
			// Periodic flush of accumulated metrics
			if len(batch) > 0 {
				s.flushBatch(batch)
				batch = make([]collector.Metric, 0, s.maxBatchSize)
			}
		}
	}
}

// flushBatch publishes a batch of metrics.
func (s *Scheduler) flushBatch(metrics []collector.Metric) {
	if len(metrics) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(s.ctx, 10*time.Second)
	defer cancel()

	// Create publisher metrics
	pbMetrics := make([]publisher.Metric, 0, len(metrics))
	for _, m := range metrics {
		pubMetric := publisher.Metric{
			NodeID:      m.NodeID,
			Timestamp:   m.Timestamp.Format(time.RFC3339Nano),
			MetricType:  string(m.MetricType),
			Value:       m.Value,
			Metadata:    m.Metadata,
		}
		pbMetrics = append(pbMetrics, pubMetric)
	}

	// Publish batch
	for _, metric := range pbMetrics {
		select {
		case <-ctx.Done():
			return
		default:
			if err := s.publisher.Publish(ctx, metric); err != nil {
				slog.Error("failed to publish metric in batch",
					"metric_type", metric.MetricType,
					"node_id", metric.NodeID,
					"error", err.Error())
			} else {
				s.metricsMu.Lock()
				s.totalMetricsPublished++
				s.metricsMu.Unlock()
			}
		}
	}
}

// OutboundChannel returns the channel for external metric injection.
func (s *Scheduler) OutboundChannel() chan<- collector.Metric {
	return s.outboundChan
}

// Metrics returns scheduler metrics.
func (s *Scheduler) Metrics() (collected, published, dropped uint64) {
	s.metricsMu.Lock()
	defer s.metricsMu.Unlock()
	return s.totalMetricsCollected, s.totalMetricsPublished, s.totalMetricsDropped
}

// PublishMetricsBatch publishes a batch of metrics with improved performance.
func (s *Scheduler) PublishMetricsBatch(ctx context.Context, metrics []collector.Metric) error {
	// Use context-aware publish with retry
	for i, metric := range metrics {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled during batch publish: %w", ctx.Err())
		default:
			pubMetric := publisher.Metric{
				NodeID:      metric.NodeID,
				Timestamp:   metric.Timestamp.Format(time.RFC3339Nano),
				MetricType:  string(metric.MetricType),
				Value:       metric.Value,
				Metadata:    metric.Metadata,
			}
			if err := s.publisher.Publish(ctx, pubMetric); err != nil {
				// Log error but continue with remaining metrics
				slog.Error("failed to publish metric in batch",
					"index", i,
					"metric_type", metric.MetricType,
					"error", err.Error())
			} else {
				s.metricsMu.Lock()
				s.totalMetricsPublished++
				s.metricsMu.Unlock()
			}
		}
	}
	return nil
}
