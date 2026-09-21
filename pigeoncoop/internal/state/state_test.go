// Package state_test contains tests for the state package.
package state

import (
	"context"
	"testing"
	"time"
)

func TestMachine_Transition(t *testing.T) {
	tests := []struct {
		name      string
		steps     []State
		wantValid []bool
	}{
		{
			name:      "disconnected -> connecting",
			steps:     []State{StateConnecting},
			wantValid: []bool{true},
		},
		{
			name:      "connecting -> connected -> stopping",
			steps:     []State{StateConnecting, StateConnected, StateStopping},
			wantValid: []bool{true, true, true},
		},
		{
			name:      "connecting -> backoff",
			steps:     []State{StateConnecting, StateBackoff},
			wantValid: []bool{true, true},
		},
		{
			name:      "backoff -> connecting",
			steps:     []State{StateConnecting, StateBackoff, StateConnecting},
			wantValid: []bool{true, true, true},
		},
		{
			name:      "connected -> stopping -> stopped",
			steps:     []State{StateConnecting, StateConnected, StateStopping, StateStopped},
			wantValid: []bool{true, true, true, true},
		},
		{
			name:      "invalid: connected -> connecting",
			steps:     []State{StateConnecting, StateConnected, StateConnecting},
			wantValid: []bool{true, true, false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMachine()

			// Machine starts in StateDisconnected
			for i, toState := range tt.steps {
				if i >= len(tt.wantValid) {
					t.Fatalf("Not enough wantValid values for step %d", i)
				}
				valid := m.Transition(toState)
				if valid != tt.wantValid[i] {
					t.Errorf("Transition(%s) = %v, want %v", toState, valid, tt.wantValid[i])
				}
			}
		})
	}
}

func TestMachine_Backoff(t *testing.T) {
	m := NewMachine()

	// First backoff should return minBackoff
	duration1 := m.BackoffDuration()
	if duration1 != 1*time.Second {
		t.Errorf("First backoff duration = %v, want 1s", duration1)
	}

	// Second backoff should double
	duration2 := m.BackoffDuration()
	expected := 2 * time.Second
	if duration2 != expected {
		t.Errorf("Second backoff duration = %v, want %v", duration2, expected)
	}

	// Reset backoff
	m.ResetBackoff()
	duration3 := m.BackoffDuration()
	if duration3 != 1*time.Second {
		t.Errorf("After reset, duration = %v, want 1s", duration3)
	}
}

func TestMachine_WaitForState(t *testing.T) {
	m := NewMachine()

	// Machine starts in Disconnected state, so waiting for Connected should timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	result := m.WaitForState(ctx, StateConnected)
	if result {
		t.Error("WaitForState should have timed out (machine is in Disconnected)")
	}

	// Transition through sequence to connected: disconnected -> connecting -> connected
	m.Transition(StateConnecting)
	m.Transition(StateConnected)

	// Should now return true since we're already in Connected state
	ctx2, cancel2 := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel2()
	result = m.WaitForState(ctx2, StateConnected)
	if !result {
		t.Error("WaitForState should have succeeded immediately")
	}
}

func TestMachine_Ready(t *testing.T) {
	m := NewMachine()

	// Should not be ready initially (channel is open)
	select {
	case <-m.Ready():
		t.Error("Machine should not be ready initially")
	default:
		// Expected
	}

	// Transition through sequence to connected
	m.Transition(StateConnecting)
	m.Transition(StateConnected)

	// Channel should be closed now
	select {
	case <-m.Ready():
		// Expected - channel was closed
	default:
		t.Error("Machine should be ready after transition to Connected")
	}
}
