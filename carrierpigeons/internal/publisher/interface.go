// Package publisher defines the interface for telemetry data publishing.
package publisher

import (
	"context"
)

// Publisher defines the interface for publishing telemetry data.
type Publisher interface {
	// Start begins the publisher's connection and publishing loop.
	Start(ctx context.Context) error

	// Stop gracefully shuts down the publisher.
	Stop(ctx context.Context) error

	// Publish sends a metric to the telemetry system.
	// Returns an error if the publisher is not connected or the message failed.
	Publish(ctx context.Context, metric Metric) error

	// IsConnected returns true if the publisher is connected to the broker.
	IsConnected() bool

	// ConnectionReady returns a channel that is closed when connected.
	ConnectionReady() <-chan struct{}
}

// Metric represents a telemetry metric to be published.
type Metric struct {
	NodeID     string            `json:"node_id"`
	Timestamp  string            `json:"timestamp"`
	MetricType string            `json:"metric_type"`
	Value      float64           `json:"value"`
	Metadata   map[string]string `json:"metadata"`
}

// PublisherConfig holds configuration for the publisher.
type PublisherConfig struct {
	BrokerURL    string
	ClientID     string
	Username     string
	Password     string
	Topic        string
	KeepAlive    int
	CleanSession bool
}

// Metric has no custom JSON marshaling: the embedded fields marshal/unmarshal
// with their standard representations. Note that a MarshalJSON method that
// calls json.Marshal on the same type would recurse infinitely, so none is
// defined here.

// ToCollectorMetric converts a publisher Metric to a collector Metric.
func (m *Metric) ToCollectorMetric() (collectorMetric Metric) {
	return collectorMetric
}
