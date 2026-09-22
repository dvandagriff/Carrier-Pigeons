// Package collector defines the interface for system metric collection.
//go:build linux

package collector

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// CPUCollector collects CPU temperature and frequency metrics on Linux.
type CPUCollector struct {
	config *CollectorConfig
}

// NewCPUCollector creates a new CPU metric collector for Linux systems.
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
	return "cpu_linux"
}

// Interval returns the recommended collection interval for CPU metrics.
func (c *CPUCollector) Interval() time.Duration {
	return 10 * time.Second
}

// Collect retrieves CPU temperature metrics from /sys/class/thermal.
func (c *CPUCollector) Collect(ctx context.Context) ([]Metric, error) {
	var metrics []Metric

	// Check if thermal zone exists
	if _, err := os.Stat(thermalZonePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("thermal zone not found: %w", err)
	}

	// Collect CPU temperature from thermal zones
	thermalZones, err := c.getThermalZones(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get thermal zones: %w", err)
	}
	metrics = append(metrics, thermalZones...)

	// Collect CPU frequency metrics
	freqMetrics, err := c.getCPUFrequencies(ctx)
	if err != nil {
		// Log but don't fail the entire collection
		fmt.Printf("Warning: failed to get CPU frequencies: %v\n", err)
	}
	metrics = append(metrics, freqMetrics...)

	return metrics, nil
}

// getThermalZones collects temperature from all thermal zones.
func (c *CPUCollector) getThermalZones(ctx context.Context) ([]Metric, error) {
	var metrics []Metric

	// Read thermal zones from the directory
	entries, err := os.ReadDir(thermalZonePath)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return metrics, ctx.Err()
		default:
		}

		if !entry.IsDir() || !strings.HasPrefix(entry.Name(), "thermal_zone") {
			continue
		}

		zonePath := filepath.Join(thermalZonePath, entry.Name())
		typePath := filepath.Join(zonePath, "type")
		tempPath := filepath.Join(zonePath, "temp")

		// Read temperature
		temp, err := os.ReadFile(tempPath)
		if err != nil {
			continue
		}

		tempValue, err := strconv.ParseFloat(strings.TrimSpace(string(temp)), 64)
		if err != nil {
			continue
		}

		// Convert millidegree Celsius to Celsius
		tempCelsius := tempValue / 1000.0

		// Read zone type
		typeData, err := os.ReadFile(typePath)
		zoneType := strings.TrimSpace(string(typeData))

		metadata := make(map[string]string)
		for k, v := range c.config.Labels {
			metadata[k] = v
		}
		metadata["zone"] = entry.Name()
		metadata["zone_type"] = zoneType

		metrics = append(metrics, Metric{
			NodeID:      "carrier-pigeon",
			Timestamp:   time.Now(),
			MetricType:  MetricTypeCPUTemp,
			Value:       tempCelsius,
			Metadata:    metadata,
		})
	}

	return metrics, nil
}

// getCPUFrequencies collects CPU frequency metrics.
func (c *CPUCollector) getCPUFrequencies(ctx context.Context) ([]Metric, error) {
	var metrics []Metric

	// Check if cpufreq directory exists
	if _, err := os.Stat(cpuFreqPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("cpufreq not found")
	}

	// Read current frequency
	currentFreqPath := filepath.Join(cpuFreqPath, "scaling_cur_freq")
	if freqData, err := os.ReadFile(currentFreqPath); err == nil {
		freqValue, _ := strconv.ParseFloat(strings.TrimSpace(string(freqData)), 64)

		metadata := make(map[string]string)
		for k, v := range c.config.Labels {
			metadata[k] = v
		}
		metadata["cpu"] = "0"
		metadata["freq_type"] = "scaling_current"

		metrics = append(metrics, Metric{
			NodeID:      "carrier-pigeon",
			Timestamp:   time.Now(),
			MetricType:  MetricTypeCPUUsage,
			Value:       freqValue / 1000.0, // Convert to MHz
			Metadata:    metadata,
		})
	}

	// Read maximum frequency
	maxFreqPath := filepath.Join(cpuFreqPath, "scaling_max_freq")
	if freqData, err := os.ReadFile(maxFreqPath); err == nil {
		freqValue, _ := strconv.ParseFloat(strings.TrimSpace(string(freqData)), 64)

		metadata := make(map[string]string)
		for k, v := range c.config.Labels {
			metadata[k] = v
		}
		metadata["cpu"] = "0"
		metadata["freq_type"] = "scaling_max"

		metrics = append(metrics, Metric{
			NodeID:      "carrier-pigeon",
			Timestamp:   time.Now(),
			MetricType:  MetricTypeCPUUsage,
			Value:       freqValue / 1000.0, // Convert to MHz
			Metadata:    metadata,
		})
	}

	return metrics, nil
}

// Pool for Metric instances to reduce GC pressure
var metricPool = sync.Pool{
	New: func() interface{} {
		return &Metric{
			Metadata: make(map[string]string),
		}
	},
}
