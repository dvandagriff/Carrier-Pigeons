// Package telemetry provides JSON marshaling/unmarshaling for telemetry payloads.
package telemetry

import (
	"encoding/json"
	"time"
)

// MarshalJSON implements custom JSON marshaling for TelemetryPayload.
func (p TelemetryPayload) MarshalJSON() ([]byte, error) {
	type Alias TelemetryPayload
	return json.Marshal(&struct {
		Timestamp string `json:"timestamp"`
		*Alias
	}{
		Timestamp: p.Timestamp.UTC().Format(time.RFC3339Nano),
		Alias:     (*Alias)(&p),
	})
}

// UnmarshalJSON implements custom JSON unmarshaling for TelemetryPayload.
func (p *TelemetryPayload) UnmarshalJSON(data []byte) error {
	type Alias TelemetryPayload
	aux := &struct {
		Timestamp string `json:"timestamp"`
		*Alias
	}{
		Alias: (*Alias)(p),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Parse timestamp
	ts, err := time.Parse(time.RFC3339Nano, aux.Timestamp)
	if err != nil {
		return err
	}
	p.Timestamp = ts

	return nil
}
