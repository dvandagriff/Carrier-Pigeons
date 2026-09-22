// Package collector defines the interface for system metric collection.
//go:build darwin

package collector

import (
	"context"
	"fmt"
	"time"
)

// CPUCollector collects CPU metrics on Darwin/macOS.
type CPUCollector struct {
	config *CollectorConfig
}

// NewCPUCollector creates a new CPU metric collector for macOS systems.
func NewCPUCollector(config *CollectorConfig) *CPUCollector {
	if config == nil {
		config = &CollectorConfig{
			Labels: make(map[string]string),
		}
	}
	return &CPUCollector{config: config}
}

// Name returns the collector's identifier.
func (c *CPUCollector) Name() string {
	return "cpu_darwin"
}

// Interval returns the recommended collection interval for CPU metrics.
func (c *CPUCollector) Interval() time.Duration {
	return 10 * time.Second
}

// Collect retrieves CPU temperature metrics from sysctl.
func (c *CPUCollector) Collect(ctx context.Context) ([]Metric, error) {
	// macOS does not have /sys/class/thermal, returning an error to indicate no implementation
	// This allows the app to compile on Darwin without the Linux-specific code
	return nil, fmt.Errorf("CPU temperature collection not implemented for Darwin")
}
