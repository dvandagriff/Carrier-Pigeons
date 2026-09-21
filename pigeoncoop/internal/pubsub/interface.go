// Package pubsub defines the interface for message broker operations (MQTT).
package pubsub

import "context"

// PubSubClient defines the interface for publish/subscribe operations.
type PubSubClient interface {
	// Connect establishes a connection to the message broker.
	Connect(ctx context.Context) error

	// Disconnect gracefully closes the connection to the broker.
	Disconnect(ctx context.Context) error

	// Subscribe registers a callback for messages on a given topic.
	// Returns a subscription handle.
	Subscribe(ctx context.Context, topic string, qos byte, callback func(context.Context, []byte)) error

	// Publish sends a message to a topic.
	Publish(ctx context.Context, topic string, qos byte, payload []byte) error

	// IsConnected returns true if the client is currently connected.
	IsConnected() bool

	// ConnectionReady returns a channel that is closed when connected.
	ConnectionReady() <-chan struct{}
}

// ClientConfig holds configuration for creating a PubSubClient.
type ClientConfig struct {
	BrokerURL    string
	ClientID     string
	Username     string
	Password     string
	KeepAlive    int
	CleanSession bool
}
