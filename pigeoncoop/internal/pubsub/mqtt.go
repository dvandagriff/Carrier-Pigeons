// Package pubsub provides MQTT implementation of the PubSubClient interface.
package pubsub

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/eclipse/paho.golang/paho"
)

// MQTTClient implements PubSubClient using paho.golang MQTT library.
type MQTTClient struct {
	config   *ClientConfig
	client   *paho.Client
	connDetails *paho.Connect
	mu       sync.RWMutex

	// For tracking connection state
	connectedCh chan struct{}
}

// NewMQTTClient creates a new MQTT client with the given config.
func NewMQTTClient(cfg *ClientConfig) *MQTTClient {
	return &MQTTClient{
		config:      cfg,
		connectedCh: make(chan struct{}),
	}
}

// Connect establishes a connection to the MQTT broker.
func (m *MQTTClient) Connect(ctx context.Context) error {
	// Create client options
	opts := paho.NewClientOptions().
		SetClientID(m.config.ClientID).
		SetAutoAcknowledged(true).
		SetCleanSession(m.config.CleanSession).
		SetKeepAlive(m.config.KeepAlive * time.Second).
		SetConnectTimeout(30 * time.Second).
		SetServerAddresses([]string{m.config.BrokerURL})

	if m.config.Username != "" {
		opts.SetUsername(m.config.Username)
	}
	if m.config.Password != "" {
		opts.SetPassword([]byte(m.config.Password))
	}

	// Set up connection handler
	opts.SetOnConnectHandler(m.onConnect)
	opts.SetConnectionLostHandler(m.onConnectionLost)

	// Create client
	client, err := paho.NewClient(opts)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	// Connect
	connACK, err := client.Connect(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	if connACK.ReturnCode != paho.ConnectReturnCodeSuccess {
		return fmt.Errorf("connection rejected: %v", connACK.ReturnCode)
	}

	m.client = client
	return nil
}

// Disconnect gracefully closes the connection.
func (m *MQTTClient) Disconnect(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.client == nil {
		return nil
	}

	// Unsubscribe from all topics
	// Note: In a real implementation, you'd track subscriptions

	// Disconnect
	opts := paho.NewDisconnect().WithReasonCode(0)
	_, err := m.client.Disconnect(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to disconnect: %w", err)
	}

	m.client = nil
	return nil
}

// Subscribe registers a callback for messages on a given topic.
func (m *MQTTClient) Subscribe(ctx context.Context, topic string, qos byte, callback func(context.Context, []byte)) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.client == nil {
		return fmt.Errorf("not connected")
	}

	// Create subscription
	subOpts := paho.NewSubscribeOptions(qos)
	subOpts.SetOnMessage(func(_ context.Context, msg *paho.Publish) {
		callback(ctx, msg.Payload)
	})

	// Subscribe
	resp, err := m.client.Subscribe(ctx, paho.NewSubscribeRequest(topic, subOpts))
	if err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}

	// Check subscription response
	for _, sub := range resp.Subscriptions {
		if sub.ReturnCode > 2 {
			return fmt.Errorf("subscription rejected: %v", sub.ReturnCode)
		}
	}

	return nil
}

// Publish sends a message to a topic.
func (m *MQTTClient) Publish(ctx context.Context, topic string, qos byte, payload []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.client == nil {
		return fmt.Errorf("not connected")
	}

	msg := paho.NewPublish(topic).
		SetQoS(qos).
		SetRetained(false).
		SetPayload(payload)

	_, err := m.client.Publish(ctx, msg)
	return err
}

// IsConnected returns true if the client is connected.
func (m *MQTTClient) IsConnected() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.client != nil
}

// ConnectionReady returns a channel that is closed when connected.
func (m *MQTTClient) ConnectionReady() <-chan struct{} {
	return m.connectedCh
}

func (m *MQTTClient) onConnect(client *paho.Client, connACK *paho.ConnectAck) {
	slog.Info("MQTT connected")
	close(m.connectedCh)
}

func (m *MQTTClient) onConnectionLost(client *paho.Client, err error) {
	slog.Error("MQTT connection lost", slog.String("error", err.Error()))
	// Reconnect logic would be handled by the state machine
}
