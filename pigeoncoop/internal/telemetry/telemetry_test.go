// Package telemetry_test contains tests for the telemetry package.
package telemetry

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTelemetryPayload_Validate(t *testing.T) {
	tests := []struct {
		name    string
		payload TelemetryPayload
		wantErr bool
	}{
		{
			name: "valid payload",
			payload: TelemetryPayload{
				NodeID:     "node-01",
				Timestamp:  time.Now(),
				MetricType: "temperature",
				Value:      25.5,
				Metadata:   map[string]string{"location": "room1"},
			},
			wantErr: false,
		},
		{
			name: "missing node_id",
			payload: TelemetryPayload{
				NodeID:     "",
				Timestamp:  time.Now(),
				MetricType: "temperature",
				Value:      25.5,
				Metadata:   map[string]string{"location": "room1"},
			},
			wantErr: true,
		},
		{
			name: "missing metric_type",
			payload: TelemetryPayload{
				NodeID:     "node-01",
				Timestamp:  time.Now(),
				MetricType: "",
				Value:      25.5,
				Metadata:   map[string]string{"location": "room1"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.payload.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTelemetryPayload_JSONMarshalUnmarshal(t *testing.T) {
	original := TelemetryPayload{
		NodeID:     "node-01",
		Timestamp:  time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		MetricType: "temperature",
		Value:      25.5,
		Metadata:   map[string]string{"location": "room1", "unit": "celsius"},
	}

	// Marshal
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	// Unmarshal
	var unmarshaled TelemetryPayload
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	// Compare
	if unmarshaled.NodeID != original.NodeID {
		t.Errorf("NodeID after unmarshal = %s, want %s", unmarshaled.NodeID, original.NodeID)
	}
	if unmarshaled.MetricType != original.MetricType {
		t.Errorf("MetricType after unmarshal = %s, want %s", unmarshaled.MetricType, original.MetricType)
	}
	if unmarshaled.Value != original.Value {
		t.Errorf("Value after unmarshal = %f, want %f", unmarshaled.Value, original.Value)
	}
	if len(unmarshaled.Metadata) != len(original.Metadata) {
		t.Errorf("Metadata length after unmarshal = %d, want %d", len(unmarshaled.Metadata), len(original.Metadata))
	}
}
