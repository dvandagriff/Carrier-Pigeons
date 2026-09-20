// Package cmd contains command-line utilities and configuration loading.
package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/dvandagriff/Carrier-Pigeons/pigeoncoop/internal/pubsub"
	"github.com/dvandagriff/Carrier-Pigeons/pigeoncoop/internal/storage"
	"github.com/dvandagriff/Carrier-Pigeons/pigeoncoop/internal/worker"
	"gopkg.in/yaml.v3"
)

// Config holds the full application configuration.
type Config struct {
	MQTT      MQTTConfig      `yaml:"mqtt"`
	Storage   StorageConfig   `yaml:"storage"`
	Worker    WorkerConfig    `yaml:"worker"`
	Metrics   MetricsConfig   `yaml:"metrics"`
}

// MQTTConfig holds MQTT-specific configuration.
type MQTTConfig struct {
	BrokerURL    string `yaml:"broker_url"`
	ClientID     string `yaml:"client_id"`
	Username     string `yaml:"username"`
	Password     string `yaml:"password"`
	KeepAlive    int    `yaml:"keep_alive"`
	CleanSession bool   `yaml:"clean_session"`
}

// StorageConfig holds database configuration.
type StorageConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	SSLMode  string `yaml:"ssl_mode"`
	PoolSize int    `yaml:"pool_size"`
}

// WorkerConfig holds worker pool configuration.
type WorkerConfig struct {
	WorkerCount   int           `yaml:"worker_count"`
	ChannelBuffer int           `yaml:"channel_buffer"`
	BatchSize     int           `yaml:"batch_size"`
	FlushInterval time.Duration `yaml:"flush_interval"`
}

// MetricsConfig holds metrics configuration.
type MetricsConfig struct {
	PrometheusPort int `yaml:"prometheus_port"`
}

// ToClientConfig converts MQTTConfig to pubsub.ClientConfig.
func (m MQTTConfig) ToClientConfig() pubsub.ClientConfig {
	return pubsub.ClientConfig{
		BrokerURL:    m.BrokerURL,
		ClientID:     m.ClientID,
		Username:     m.Username,
		Password:     m.Password,
		KeepAlive:    m.KeepAlive,
		CleanSession: m.CleanSession,
	}
}

// ToStoreConfig converts StorageConfig to storage.StoreConfig.
func (s StorageConfig) ToStoreConfig() storage.StoreConfig {
	return storage.StoreConfig{
		Host:     s.Host,
		Port:     s.Port,
		User:     s.User,
		Password: s.Password,
		Database: s.Database,
		SSLMode:  s.SSLMode,
		PoolSize: s.PoolSize,
	}
}

// ToWorkerConfig converts WorkerConfig to worker.Config.
func (w WorkerConfig) ToWorkerConfig() worker.Config {
	return worker.Config{
		WorkerCount:   w.WorkerCount,
		ChannelBuffer: w.ChannelBuffer,
		BatchSize:     w.BatchSize,
		FlushInterval: w.FlushInterval,
	}
}

// LoadConfig loads configuration from a YAML file.
func LoadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return Config{}, fmt.Errorf("failed to parse config: %w", err)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return Config{}, err
	}

	return config, nil
}

// Validate checks that required configuration values are present.
func (c Config) Validate() error {
	if c.MQTT.BrokerURL == "" {
		return fmt.Errorf("mqtt.broker_url is required")
	}
	if c.Storage.Host == "" {
		return fmt.Errorf("storage.host is required")
	}
	if c.Storage.Database == "" {
		return fmt.Errorf("storage.database is required")
	}
	if c.Worker.WorkerCount == 0 {
		c.Worker.WorkerCount = 4 // default
	}
	if c.Worker.ChannelBuffer == 0 {
		c.Worker.ChannelBuffer = 10000 // default
	}
	if c.Worker.BatchSize == 0 {
		c.Worker.BatchSize = 100 // default
	}
	if c.Worker.FlushInterval == 0 {
		c.Worker.FlushInterval = 1 * time.Second // default
	}
	return nil
}
