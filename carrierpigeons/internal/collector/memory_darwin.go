// Package collector defines the interface for system metric collection.
//go:build darwin

package collector

import (
	"context"
	"fmt"
	"time"
)

// MemoryCollector collects memory metrics on Darwin/macOS.
type MemoryCollector struct {
	config *CollectorConfig
}

// NewMemoryCollector creates a new memory metric collector for macOS.
func NewMemoryCollector(config *CollectorConfig) *MemoryCollector {
	if config == nil {
		config = &CollectorConfig{
			Labels: make(map[string]string),
		}
	}
	return &MemoryCollector{config: config}
}

// Name returns the collector's identifier.
func (m *MemoryCollector) Name() string {
	return "memory_darwin"
}

// Interval returns the recommended collection interval for memory metrics.
func (m *MemoryCollector) Interval() time.Duration {
	return 15 * time.Second
}

// Collect retrieves memory metrics from sysctl.
func (m *MemoryCollector) Collect(ctx context.Context) ([]Metric, error) {
	// macOS does not have /proc/meminfo, returning an error to indicate no implementation
	// This allows the app to compile on Darwin without the Linux-specific code
	return nil, fmt.Errorf("memory collection not implemented for Darwin")
}
