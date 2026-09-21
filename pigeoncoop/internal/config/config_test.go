// Package config_test contains tests for the config package.
package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	configContent := `mqtt:
  broker_url: "tcp://localhost:1883"
  client_id: "test-client"
  username: "user"
  password: "pass"
  keep_alive: 30
  clean_session: true

storage:
  host: "localhost"
  port: 5432
  user: "test"
  password: "test"
  database: "test"
  ssl_mode: "disable"
  pool_size: 10

worker:
  worker_count: 4
  channel_buffer: 1000
  batch_size: 50
  flush_interval: 1s

metrics:
  prometheus_port: 9090
`

	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(configContent)); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	cfg, err := LoadConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.MQTT.BrokerURL != "tcp://localhost:1883" {
		t.Errorf("Expected broker_url to be 'tcp://localhost:1883', got '%s'", cfg.MQTT.BrokerURL)
	}
	if cfg.Worker.WorkerCount != 4 {
		t.Errorf("Expected worker_count to be 4, got %d", cfg.Worker.WorkerCount)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				MQTT: MQTTConfig{
					BrokerURL: "tcp://localhost:1883",
				},
				Storage: StorageConfig{
					Host:     "localhost",
					Database: "test",
				},
			},
			wantErr: false,
		},
		{
			name: "missing broker URL",
			config: Config{
				MQTT: MQTTConfig{
					BrokerURL: "",
				},
				Storage: StorageConfig{
					Host:     "localhost",
					Database: "test",
				},
			},
			wantErr: true,
		},
		{
			name: "missing host",
			config: Config{
				MQTT: MQTTConfig{
					BrokerURL: "tcp://localhost:1883",
				},
				Storage: StorageConfig{
					Host:     "",
					Database: "test",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestToClientConfig(t *testing.T) {
	cfg := MQTTConfig{
		BrokerURL: "tcp://localhost:1883",
		ClientID:  "test-client",
		Username:  "user",
		Password:  "pass",
		KeepAlive: 30,
	}

	clientCfg := cfg.ToClientConfig()

	if clientCfg.BrokerURL != "tcp://localhost:1883" {
		t.Errorf("Expected broker_url to be 'tcp://localhost:1883', got '%s'", clientCfg.BrokerURL)
	}
	if clientCfg.ClientID != "test-client" {
		t.Errorf("Expected client_id to be 'test-client', got '%s'", clientCfg.ClientID)
	}
}
