// Package publisher provides MQTT implementation of the Publisher interface.
package publisher

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/eclipse/paho.golang/paho"
)

// MQTTPublisher implements Publisher using the Paho MQTT client library.
type MQTTPublisher struct {
	config      *PublisherConfig
	client      *paho.Client
	mu          sync.RWMutex
	connectedCh chan struct{}
	connected   bool
}

// NewMQTTPublisher creates a new MQTT publisher.
func NewMQTTPublisher(cfg *PublisherConfig) *MQTTPublisher {
	if cfg == nil {
		cfg = &PublisherConfig{}
	}
	if cfg.Topic == "" {
		cfg.Topic = "telemetry/+"
	}
	if cfg.KeepAlive == 0 {
		cfg.KeepAlive = 30
	}
	return &MQTTPublisher{
		connectedCh: make(chan struct{}),
		config:      cfg,
	}
}

// Start establishes the MQTT connection and begins the publishing loop.
func (p *MQTTPublisher) Start(ctx context.Context) error {
	slog.Info("starting MQTT publisher", "broker_url", p.config.BrokerURL)

	// Create client options
	opts := paho.NewClientOptions().
		SetClientID(p.config.ClientID).
		SetAutoAcknowledged(true).
		SetCleanSession(p.config.CleanSession).
		SetKeepAlive(time.Duration(p.config.KeepAlive) * time.Second).
		SetConnectTimeout(30 * time.Second).
		SetServerAddresses([]string{p.config.BrokerURL})

	if p.config.Username != "" {
		opts.SetUsername(p.config.Username)
	}
	if p.config.Password != "" {
		opts.SetPassword([]byte(p.config.Password))
	}

	// Set up connection handlers
	opts.SetOnConnectHandler(p.onConnect)
	opts.SetConnectionLostHandler(p.onConnectionLost)

	// Create client
	client, err := paho.NewClient(opts)
	if err != nil {
		return fmt.Errorf("failed to create MQTT client: %w", err)
	}

	// Connect
	connACK, err := client.Connect(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to MQTT broker: %w", err)
	}

	if connACK.ReturnCode != paho.ConnectReturnCodeSuccess {
		return fmt.Errorf("MQTT connection rejected: %v", connACK.ReturnCode)
	}

	p.mu.Lock()
	p.client = client
	p.connected = true
	close(p.connectedCh)
	p.mu.Unlock()

	slog.Info("MQTT publisher connected", "client_id", p.config.ClientID)
	return nil
}

// Stop gracefully closes the MQTT connection.
func (p *MQTTPublisher) Stop(ctx context.Context) error {
	p.mu.Lock()
	if p.client == nil {
		p.mu.Unlock()
		return nil
	}
	p.mu.Unlock()

	slog.Info("disconnecting from MQTT broker")

	// Disconnect with timeout
	dcCtx, dcCancel := context.WithTimeout(ctx, 10*time.Second)
	defer dcCancel()

	opts := paho.NewDisconnect().WithReasonCode(0)
	_, err := p.client.Disconnect(dcCtx, opts)
	
	p.mu.Lock()
	p.connected = false
	p.client = nil
	p.mu.Unlock()

	if err != nil {
		return fmt.Errorf("failed to disconnect: %w", err)
	}

	slog.Info("MQTT publisher disconnected")
	return nil
}

// Publish sends a metric to the MQTT topic.
func (p *MQTTPublisher) Publish(ctx context.Context, metric Metric) error {
	p.mu.RLock()
	client := p.client
	connected := p.connected
	p.mu.RUnlock()

	if !connected || client == nil {
		return fmt.Errorf("not connected to MQTT broker")
	}

	// Convert metric to JSON payload
	payload, err := metric.MarshalJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal metric: %w", err)
	}

	// Publish to topic
	msg := paho.NewPublish(p.config.Topic).
		SetQoS(1).
		SetRetained(false).
		SetPayload(payload)

	pubACK, err := client.Publish(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to publish metric: %w", err)
	}

	// Wait for acknowledgment with timeout
	pubCtx, pubCancel := context.WithTimeout(ctx, 5*time.Second)
	defer pubCancel()

	select {
	case <-pubACK.Delivered:
		slog.Debug("metric published", "topic", p.config.Topic)
		return nil
	case <-pubACK.Error:
		return fmt.Errorf("publish delivery failed")
	case <-pubCtx.Done():
		return fmt.Errorf("publish acknowledgment timeout: %w", pubCtx.Err())
	}
}

// IsConnected returns true if the publisher is connected.
func (p *MQTTPublisher) IsConnected() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.connected
}

// ConnectionReady returns a channel that is closed when connected.
func (p *MQTTPublisher) ConnectionReady() <-chan struct{} {
	return p.connectedCh
}

func (p *MQTTPublisher) onConnect(client *paho.Client, connACK *paho.ConnectAck) {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	p.connected = true
	close(p.connectedCh)
	slog.Info("MQTT connection established")
}

func (p *MQTTPublisher) onConnectionLost(client *paho.Client, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	p.connected = false
	slog.Error("MQTT connection lost", "error", err.Error())
}
