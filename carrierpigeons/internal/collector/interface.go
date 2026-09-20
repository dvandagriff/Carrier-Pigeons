// Package collector defines the interface for system metric collection.
package collector

import (
	"context"
	"time"
)

// MetricType represents the type of metric being collected.
type MetricType string

const (
	MetricTypeCPUUsage       MetricType = "cpu_usage"
	MetricTypeCPUTemp        MetricType = "cpu_temp"
	MetricTypeMemoryUsage    MetricType = "memory_usage"
	MetricTypeMemoryAvailable MetricType = "memory_available"
	MetricTypeDiskUsage      MetricType = "disk_usage"
	MetricTypeServiceState   MetricType = "service_state"
)

// Metric represents a single collected metric value.
type Metric struct {
	NodeID      string
	Timestamp   time.Time
	MetricType  MetricType
	Value       float64
	Metadata    map[string]string
}

// MetricCollector defines the interface for collecting system metrics.
type MetricCollector interface {
	// Name returns the collector's identifier.
	Name() string

	// Collect retrieves the current metrics from the system.
	// The ctx can be used to cancel long-running collection operations.
	Collect(ctx context.Context) ([]Metric, error)

	// Interval returns the recommended collection interval for this metric.
	Interval() time.Duration
}

// CollectorConfig holds configuration for metric collectors.
type CollectorConfig struct {
	// Labels are key-value pairs to attach to all metrics from this collector.
	Labels map[string]string
}
