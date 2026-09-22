package publisher

import (
	"testing"

	"github.com/dvandagriff/Carrier-Pigeons/carrierpigeons/internal/state"
)

func TestNewMQTTPublisher(t *testing.T) {
	cfg := &PublisherConfig{
		BrokerURL: "tcp://localhost:1883",
		ClientID:  "test-client",
		Topic:     "test",
	}

	p := NewMQTTPublisher(cfg)

	if p == nil {
		t.Fatal("NewMQTTPublisher returned nil")
	}

	if p.config.ClientID != "test-client" {
		t.Errorf("Expected ClientID to be 'test-client', got '%s'", p.config.ClientID)
	}

	if p.config.Topic != "test" {
		t.Errorf("Expected Topic to be 'test', got '%s'", p.config.Topic)
	}
}

func TestNewMQTTPublisherWithDefaults(t *testing.T) {
	cfg := &PublisherConfig{}

	p := NewMQTTPublisher(cfg)

	if p == nil {
		t.Fatal("NewMQTTPublisher returned nil")
	}

	if p.config.Topic != "telemetry/+" {
		t.Errorf("Expected default Topic to be 'telemetry/+', got '%s'", p.config.Topic)
	}

	if p.config.KeepAlive != 30 {
		t.Errorf("Expected default KeepAlive to be 30, got %d", p.config.KeepAlive)
	}
}

func TestMQTTPublisherStateMachine(t *testing.T) {
	cfg := &PublisherConfig{
		BrokerURL: "tcp://localhost:1883",
		ClientID:  "test-client",
	}

	p := NewMQTTPublisher(cfg)

	if p.StateMachine() == nil {
		t.Fatal("StateMachine() returned nil")
	}

	// Verify initial state
	if p.StateMachine().Current() != state.StateDisconnected {
		t.Errorf("Expected initial state to be StateDisconnected, got %v", p.StateMachine().Current())
	}
}
