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
	"time"
)

const (
	cpuTempPath     = "/sys/class/thermal"
	thermalZonePath = "/sys/class/thermal/thermal_zone"
	cpuFreqPath     = "/sys/devices/system/cpu/cpu0/cpufreq"
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
	currentFreqPath := filepath.Join(cpuFreqPath, " scaling_cur_freq")
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
	maxFreqPath := filepath.Join(cpuFreqPath, " scaling_max_freq")
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

// MemoryCollector collects memory usage metrics.
type MemoryCollector struct {
	config *CollectorConfig
}

// NewMemoryCollector creates a new memory metric collector for Linux.
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
	memInfo := make(map[string]int64)

	for scanner.Scan() {
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

	// Convert to bytes for precise values
	totalBytes := float64(totalKB) * 1024
	availableBytes := float64(availableMemoryKB) * 1024
	usedBytes := float64(usedMemoryKB) * 1024

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
