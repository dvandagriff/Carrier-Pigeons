// Package pubsub provides MQTT implementation of the PubSubClient interface.
package pubsub

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/eclipse/paho.golang/paho"
)

// MQTTClient implements PubSubClient using paho.golang MQTT library.
type MQTTClient struct {
	config *ClientConfig
	client *paho.Client
	mu     sync.RWMutex

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
	// Parse broker URL
	urls := []string{m.config.BrokerURL}

	// Create a TCP connection
	conn, err := net.DialTimeout("tcp", urls[0], 30*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to broker: %w", err)
	}

	// Create client config
	conf := paho.ClientConfig{
		ClientID: m.config.ClientID,
		Conn:     conn,
	}

	// Set up connection handler
	conf.OnServerDisconnect = func(d *paho.Disconnect) {
		slog.Error("MQTT server disconnected", "reason", d.ReasonCode)
	}

	// Create client
	client := paho.NewClient(conf)

	// Create connect packet
	connect := &paho.Connect{
		ClientID:     m.config.ClientID,
		KeepAlive:    uint16(m.config.KeepAlive),
		CleanStart:   m.config.CleanSession,
		UsernameFlag: m.config.Username != "",
		PasswordFlag: m.config.Password != "",
		Username:     m.config.Username,
		Password:     []byte(m.config.Password),
	}

	if connect.KeepAlive == 0 {
		connect.KeepAlive = 30 // default
	}

	// Connect
	connACK, err := client.Connect(ctx, connect)
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to connect: %w", err)
	}

	if connACK.ReasonCode != 0 {
		conn.Close()
		return fmt.Errorf("connection rejected: %v", connACK.ReasonCode)
	}

	m.client = client
	close(m.connectedCh)
	return nil
}

// Disconnect gracefully closes the connection.
func (m *MQTTClient) Disconnect(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.client == nil {
		return nil
	}

	// Disconnect
	disconnect := &paho.Disconnect{
		ReasonCode: 0,
	}
	if err := m.client.Disconnect(disconnect); err != nil {
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

	// Set up message handler for this subscription
	handler := func(pr paho.PublishReceived) (bool, error) {
		msg := pr.Packet
		callback(ctx, msg.Payload)
		return true, nil
	}

	m.client.AddOnPublishReceived(handler)

	// Subscribe
	sub := &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{
				Topic: topic,
				QoS:   qos,
			},
		},
	}

	resp, err := m.client.Subscribe(ctx, sub)
	if err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}

	// Check subscription response
	for _, reason := range resp.Reasons {
		if reason > 2 {
			return fmt.Errorf("subscription rejected: %v", reason)
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

	msg := &paho.Publish{
		Topic:   topic,
		QoS:     qos,
		Retain:  false,
		Payload: payload,
	}

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
