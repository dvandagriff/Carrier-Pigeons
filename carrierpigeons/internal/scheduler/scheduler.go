// Package scheduler provides the ticker-based scheduler for metric collection.
package scheduler

import (
	"context"
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
}

// Config holds configuration for the scheduler.
type Config struct {
	MetricChannelSize int
	OutboundChannelSize int
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

	// Send metrics to outbound channel
	for _, metric := range metrics {
		select {
		case s.outboundChan <- metric:
			// Metric queued successfully
		case <-s.ctx.Done():
			// Scheduler stopped, drop metric
			return
		default:
			// Channel full, drop metric
			slog.Warn("outbound channel full, dropping metric", "collector", c.Name())
		}
	}
}

// publishMetrics reads metrics from the outbound channel and publishes them.
func (s *Scheduler) publishMetrics() {
	defer s.wg.Done()

	// Wait for publisher to be ready
	ready := s.publisher.ConnectionReady()
	select {
	case <-ready:
	case <-s.ctx.Done():
		return
	}

	for {
		select {
		case <-s.ctx.Done():
			// Publish any remaining metrics
			s.flushOutbound()
			slog.Info("publisher stopped")
			return

		case metric, ok := <-s.outboundChan:
			if !ok {
				// Channel closed
				return
			}

			// Publish the metric
			pubCtx, pubCancel := context.WithTimeout(s.ctx, 10*time.Second)
			if err := s.publisher.Publish(pubCtx, metric); err != nil {
				slog.Error("failed to publish metric", 
					"metric_type", metric.MetricType,
					"node_id", metric.NodeID,
					"error", err.Error())
			}
			pubCancel()
		}
	}
}

// flushOutbound sends any remaining metrics in the outbound channel.
func (s *Scheduler) flushOutbound() {
	for {
		select {
		case metric, ok := <-s.outboundChan:
			if !ok {
				return
			}
			pubCtx, pubCancel := context.WithTimeout(context.Background(), 5*time.Second)
			if err := s.publisher.Publish(pubCtx, metric); err != nil {
				slog.Error("failed to flush metric", "error", err.Error())
			}
			pubCancel()
		default:
			return
		}
	}
}

// OutboundChannel returns the channel for external metric injection.
func (s *Scheduler) OutboundChannel() chan<- collector.Metric {
	return s.outboundChan
}
