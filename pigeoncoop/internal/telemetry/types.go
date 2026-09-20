// Package telemetry contains core data structures and interfaces for PigeonCoop.
package telemetry

import "time"

// TelemetryPayload represents a generic telemetry event from an edge node.
type TelemetryPayload struct {
	NodeID      string            `json:"node_id"`
	Timestamp   time.Time         `json:"timestamp"`
	MetricType  string            `json:"metric_type"`
	Value       float64           `json:"value"`
	Metadata    map[string]string `json:"metadata"`
}

// Validate checks if the payload has all required fields.
func (p *TelemetryPayload) Validate() error {
	if p.NodeID == "" {
		return ErrMissingNodeID
	}
	if p.MetricType == "" {
		return ErrMissingMetricType
	}
	return nil
}

// Error constants for telemetry validation.
var (
	ErrMissingNodeID    = &TelemetryError{Message: "missing node_id"}
	ErrMissingMetricType = &TelemetryError{Message: "missing metric_type"}
)

// TelemetryError represents a validation or processing error.
type TelemetryError struct {
	Message string
}

func (e *TelemetryError) Error() string {
	return e.Message
}
