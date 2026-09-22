// Package collector defines the interface for system metric collection.
//go:build linux

package collector

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// MemoryCollector collects memory usage metrics on Linux.
type MemoryCollector struct {
	config     *CollectorConfig
	bufferPool *sync.Pool
}

// NewMemoryCollector creates a new memory metric collector for Linux.
func NewMemoryCollector(config *CollectorConfig) *MemoryCollector {
	if config == nil {
		config = &CollectorConfig{
			Labels: make(map[string]string),
		}
	}
	// Create a buffer pool for reader operations to reduce GC pressure
	bufferPool := &sync.Pool{
		New: func() interface{} {
			buf := make([]byte, 128)
			return &buf
		},
	}
	return &MemoryCollector{config: config, bufferPool: bufferPool}
}

// Name returns the collector's identifier.
func (m *MemoryCollector) Name() string {
	return "memory_linux"
}

// Interval returns the recommended collection interval for memory metrics.
func (m *MemoryCollector) Interval() time.Duration {
	return 15 * time.Second
}

// Collect retrieves memory metrics from /proc/meminfo.
func (m *MemoryCollector) Collect(ctx context.Context) ([]Metric, error) {
	var metrics []Metric

	meminfoPath := "/proc/meminfo"
	file, err := os.Open(meminfoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open meminfo: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// Set larger buffer for reading /proc/meminfo
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 128*1024)

	memInfo := make(map[string]int64)

	for scanner.Scan() {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return metrics, ctx.Err()
		default:
		}

		fields := strings.Split(scanner.Text(), ":")
		if len(fields) != 2 {
			continue
		}

		name := strings.TrimSpace(fields[0])
		valueStr := strings.TrimSpace(fields[1])
		valueStr = strings.TrimSuffix(valueStr, " kB")

		value, err := strconv.ParseInt(strings.TrimSpace(valueStr), 10, 64)
		if err != nil {
			continue
		}

		memInfo[name] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading meminfo: %w", err)
	}

	// Calculate memory metrics
	totalKB := memInfo["MemTotal"]
	freeKB := memInfo["MemFree"]
	availableKB := memInfo["MemAvailable"]
	buffersKB := memInfo["Buffers"]
	cachedKB := memInfo["Cached"]
	reclaimableKB := memInfo["SReclaimable"]
	shmemKB := memInfo["Shmem"]

	// Available memory (from /proc/meminfo definition)
	availableMemoryKB := availableKB
	if availableKB == 0 {
		availableMemoryKB = freeKB + buffersKB + cachedKB - reclaimableKB - shmemKB
	}

	// Total used memory
	usedMemoryKB := totalKB - availableMemoryKB

	// Memory usage percentage
	memUsagePercent := float64(usedMemoryKB) / float64(totalKB) * 100.0

	// Memory available percentage
	memAvailablePercent := float64(availableMemoryKB) / float64(totalKB) * 100.0

	metadata := make(map[string]string)
	for k, v := range m.config.Labels {
		metadata[k] = v
	}

	metrics = append(metrics, Metric{
		NodeID:      "carrier-pigeon",
		Timestamp:   time.Now(),
		MetricType:  MetricTypeMemoryUsage,
		Value:       memUsagePercent,
		Metadata:    metadata,
	})

	metrics = append(metrics, Metric{
		NodeID:      "carrier-pigeon",
		Timestamp:   time.Now(),
		MetricType:  MetricTypeMemoryAvailable,
		Value:       memAvailablePercent,
		Metadata:    metadata,
	})

	return metrics, nil
}
