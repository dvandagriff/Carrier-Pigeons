// Package main is the entry point for the CarrierPigeons daemon.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dvandagriff/Carrier-Pigeons/carrierpigeons/internal/collector"
	"github.com/dvandagriff/Carrier-Pigeons/carrierpigeons/internal/publisher"
	"github.com/dvandagriff/Carrier-Pigeons/carrierpigeons/internal/state"
	"github.com/dvandagriff/Carrier-Pigeons/carrierpigeons/internal/scheduler"
	"gopkg.in/yaml.v3"
)

// Config holds the application configuration.
type Config struct {
	NodeID        string        `yaml:"node_id"`
	Metrics       MetricsConfig `yaml:"metrics"`
	Publisher     PublisherConfig `yaml:"publisher"`
	Scheduler     SchedulerConfig `yaml:"scheduler"`
}

// MetricsConfig holds metric collection configuration.
type MetricsConfig struct {
	EnabledCollectors []string `yaml:"enabled_collectors"`
}

// PublisherConfig holds MQTT publisher configuration.
type PublisherConfig struct {
	BrokerURL    string `yaml:"broker_url"`
	ClientID     string `yaml:"client_id"`
	Username     string `yaml:"username"`
	Password     string `yaml:"password"`
	Topic        string `yaml:"topic"`
	KeepAlive    int    `yaml:"keep_alive"`
	CleanSession bool   `yaml:"clean_session"`
}

// SchedulerConfig holds scheduler configuration.
type SchedulerConfig struct {
	MetricChannelSize   int `yaml:"metric_channel_size"`
	OutboundChannelSize int `yaml:"outbound_channel_size"`
}

func main() {
	// Load configuration
	config, err := LoadConfig("config.yaml")
	if err != nil {
		slog.Error("failed to load configuration", "error", err.Error())
		os.Exit(1)
	}

	// Set up logging
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	// Create state machine
	stateMachine := state.NewMachine()

	// Create collectors
	collectors, err := createCollectors(config)
	if err != nil {
		slog.Error("failed to create collectors", "error", err.Error())
		os.Exit(1)
	}

	// Create publisher
	publisher := publisher.NewMQTTPublisher(&publisher.PublisherConfig{
		BrokerURL:    config.Publisher.BrokerURL,
		ClientID:     config.Publisher.ClientID,
		Username:     config.Publisher.Username,
		Password:     config.Publisher.Password,
		Topic:        config.Publisher.Topic,
		KeepAlive:    config.Publisher.KeepAlive,
		CleanSession: config.Publisher.CleanSession,
	})

	// Create scheduler
	sched := scheduler.NewScheduler(
		collectors,
		publisher,
		&scheduler.Config{
			MetricChannelSize:   config.Scheduler.MetricChannelSize,
			OutboundChannelSize: config.Scheduler.OutboundChannelSize,
		},
	)

	// Start publisher
	if err := publisher.Start(context.Background()); err != nil {
		slog.Error("failed to start publisher", "error", err.Error())
		os.Exit(1)
	}

	// Start scheduler
	if err := sched.Start(); err != nil {
		slog.Error("failed to start scheduler", "error", err.Error())
		os.Exit(1)
	}

	// Transition to connected state
	stateMachine.Transition(state.StateConnected)

	slog.Info("CarrierPigeons daemon started", "node_id", config.NodeID)

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	slog.Info("received shutdown signal", "signal", sig.String())

	// Transition to disconnected state
	stateMachine.Transition(state.StateDisconnected)

	// Stop scheduler
	sched.Stop()

	// Stop publisher with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := publisher.Stop(ctx); err != nil {
		slog.Error("failed to stop publisher", "error", err.Error())
	}

	slog.Info("CarrierPigeons daemon stopped")
}

// createCollectors creates metric collectors based on configuration.
func createCollectors(config *Config) ([]collector.MetricCollector, error) {
	var collectors []collector.MetricCollector
	collectorConfig := &collector.CollectorConfig{
		Labels: map[string]string{
			"node_id": config.NodeID,
		},
	}

	// Create default collectors
	collectors = append(collectors, collector.NewCPUCollector(collectorConfig))
	collectors = append(collectors, collector.NewMemoryCollector(collectorConfig))

	return collectors, nil
}

// LoadConfig loads configuration from a YAML file.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, err
	}

	// Set defaults
	if config.NodeID == "" {
		hostname, err := os.Hostname()
		if err != nil {
			config.NodeID = "carrier-pigeon"
		} else {
			config.NodeID = hostname
		}
	}

	if config.Publisher.Topic == "" {
		config.Publisher.Topic = "telemetry"
	}

	if config.Scheduler.MetricChannelSize == 0 {
		config.Scheduler.MetricChannelSize = 1000
	}

	if config.Scheduler.OutboundChannelSize == 0 {
		config.Scheduler.OutboundChannelSize = 10000
	}

	return &config, nil
}

// Validate checks that required configuration values are present.
func (c *Config) Validate() error {
	if c.Publisher.BrokerURL == "" {
		return fmt.Errorf("publisher.broker_url is required")
	}
	if c.Publisher.ClientID == "" {
		return fmt.Errorf("publisher.client_id is required")
	}
	if c.Scheduler.MetricChannelSize <= 0 {
		return fmt.Errorf("scheduler.metric_channel_size must be positive")
	}
	if c.Scheduler.OutboundChannelSize <= 0 {
		return fmt.Errorf("scheduler.outbound_channel_size must be positive")
	}
	return nil
}
