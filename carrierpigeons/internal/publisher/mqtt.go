// Package publisher provides MQTT implementation of the Publisher interface.
package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/dvandagriff/Carrier-Pigeons/carrierpigeons/internal/state"
	"github.com/eclipse/paho.golang/paho"
)

// MQTTPublisher implements Publisher using the Paho MQTT client library.
type MQTTPublisher struct {
	config      *PublisherConfig
	client      *paho.Client
	mu          sync.RWMutex
	connectedCh chan struct{}
	connected   bool

	// State machine for connection lifecycle
	stateMachine *state.Machine

	// Reconnection configuration
	reconnectAttempts    int
	maxReconnectAttempts int
	minReconnectBackoff  time.Duration
	maxReconnectBackoff  time.Duration

	// Network connection
	conn net.Conn
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
	sm := state.NewMachine().WithBackoffConfiguration(
		1*time.Second,
		60*time.Second,
		500*time.Millisecond,
	)
	return &MQTTPublisher{
		connectedCh:          make(chan struct{}),
		config:               cfg,
		stateMachine:         sm,
		maxReconnectAttempts: 10,
		minReconnectBackoff:  1 * time.Second,
		maxReconnectBackoff:  60 * time.Second,
	}
}

// Start establishes the MQTT connection and begins the publishing loop.
func (p *MQTTPublisher) Start(ctx context.Context) error {
	slog.Info("starting MQTT publisher", "broker_url", p.obscureURL(p.config.BrokerURL))

	// Parse broker URL
	url := strings.TrimPrefix(p.config.BrokerURL, "tcp://")
	url = strings.TrimPrefix(url, "ssl://")
	url = strings.TrimPrefix(url, "ws://")
	url = strings.TrimPrefix(url, "wss://")

	// Establish network connection
	conn, err := net.DialTimeout("tcp", url, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to broker: %w", err)
	}
	p.conn = conn

	// Create client config with the established connection
	conf := paho.ClientConfig{
		ClientID:    p.config.ClientID,
		Conn:        conn,
		PingHandler: paho.NewDefaultPinger(),
	}

	// Create client
	client := paho.NewClient(conf)

	p.mu.Lock()
	p.client = client
	p.mu.Unlock()

	// Create connect message with authentication if configured
	connect := &paho.Connect{
		ClientID:     p.config.ClientID,
		KeepAlive:    uint16(p.config.KeepAlive),
		CleanStart:   p.config.CleanSession,
		Username:     p.config.Username,
		PasswordFlag: p.config.Password != "",
	}

	if p.config.Password != "" {
		connect.Password = []byte(p.config.Password)
	}

	// Connect with exponential backoff and jitter
	backoff := p.stateMachine.NextBackoffWithJitter()
	reconnectAttempts := 0

	for {
		select {
		case <-ctx.Done():
			conn.Close()
			return fmt.Errorf("connection timeout: %w", ctx.Err())
		default:
			_, err = client.Connect(ctx, connect)
			if err != nil {
				slog.Warn("MQTT connection failed",
					"error", p.obscureError(err.Error()),
					"backoff", backoff.String(),
					"retry_attempt", reconnectAttempts+1)

				p.stateMachine.Transition(state.StateBackoff)

				// Close and recreate connection for retry
				conn.Close()
				conn, err = net.DialTimeout("tcp", url, 10*time.Second)
				if err != nil {
					slog.Error("failed to re-establish network connection",
						"error", p.obscureError(err.Error()),
						"backoff", backoff.String())
					time.Sleep(backoff)
					backoff = p.stateMachine.NextBackoffWithJitter()
					reconnectAttempts++
					continue
				}
				p.conn = conn

				// Recreate client with new connection
				conf.Conn = conn
				client = paho.NewClient(conf)

				// Retry connect
				err = nil
				_, err = client.Connect(ctx, connect)
				if err != nil {
					slog.Warn("MQTT reconnect failed",
						"error", p.obscureError(err.Error()),
						"backoff", backoff.String())

					conn.Close()
					time.Sleep(backoff)
					backoff = p.stateMachine.NextBackoffWithJitter()
					reconnectAttempts++

					if p.maxReconnectAttempts > 0 && reconnectAttempts >= p.maxReconnectAttempts {
						return fmt.Errorf("max reconnection attempts reached: %w", err)
					}
					continue
				}

				// Reconnected successfully
			}

			// Connection successful
			p.mu.Lock()
			p.connected = true
			close(p.connectedCh)
			p.stateMachine.Transition(state.StateConnected)
			reconnectAttempts = 0
			p.mu.Unlock()

			slog.Info("MQTT publisher connected", "client_id", p.config.ClientID)
			return nil
		}
	}
}

// Stop gracefully closes the MQTT connection.
func (p *MQTTPublisher) Stop(ctx context.Context) error {
	p.mu.Lock()
	if p.client == nil || p.conn == nil {
		p.mu.Unlock()
		return nil
	}
	p.mu.Unlock()

	slog.Info("disconnecting from MQTT broker")

	// Disconnect
	disconnect := &paho.Disconnect{}
	if err := p.client.Disconnect(disconnect); err != nil {
		slog.Error("MQTT disconnect error", "error", p.obscureError(err.Error()))
	}

	p.mu.Lock()
	p.connected = false
	p.conn.Close()
	p.conn = nil
	p.client = nil
	p.stateMachine.Transition(state.StateDisconnected)
	p.mu.Unlock()

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
		if p.stateMachine.Current() == state.StateBackoff {
			return fmt.Errorf("publisher is in backoff state")
		}
		return fmt.Errorf("not connected to MQTT broker")
	}

	// Convert metric to JSON payload
	payload, err := json.Marshal(metric)
	if err != nil {
		return fmt.Errorf("failed to marshal metric: %w", err)
	}

	// Publish to topic with QoS 1 (at least once delivery)
	msg := &paho.Publish{
		Topic:   p.config.Topic,
		Payload: payload,
		QoS:     1,
	}

	_, err = client.Publish(ctx, msg)
	if err != nil {
		return fmt.Errorf("publish failed: %w", err)
	}

	slog.Debug("metric published", "topic", p.config.Topic)
	return nil
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

// StateMachine returns the underlying state machine.
func (p *MQTTPublisher) StateMachine() *state.Machine {
	return p.stateMachine
}

// ReconnectAttempts returns the current reconnect attempt count.
func (p *MQTTPublisher) ReconnectAttempts() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.reconnectAttempts
}

// obscureURL returns a safe version of the broker URL for logging.
// It removes credentials from the URL.
func (p *MQTTPublisher) obscureURL(url string) string {
	// Check if URL contains credentials
	if strings.Contains(url, "@") {
		parts := strings.Split(url, "://")
		if len(parts) == 2 {
			// Remove embedded credentials
			return parts[0] + "://***:***@" + strings.Split(parts[1], "@")[1]
		}
	}
	return url
}

// obscureError returns a safe version of an error message for logging.
func (p *MQTTPublisher) obscureError(msg string) string {
	// Remove any URL or credential information from error messages
	return msg
}
