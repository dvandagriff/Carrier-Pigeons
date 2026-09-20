// Package state provides the state machine for MQTT connection lifecycle management.
package state

import (
	"context"
	"sync"
	"time"
)

// State represents the current state of the MQTT connection.
type State int

const (
	StateDisconnected State = iota
	StateConnecting
	StateConnected
	StateReconnecting
	StateBackoff
)

func (s State) String() string {
	switch s {
	case StateDisconnected:
		return "DISCONNECTED"
	case StateConnecting:
		return "CONNECTING"
	case StateConnected:
		return "CONNECTED"
	case StateReconnecting:
		return "RECONNECTING"
	case StateBackoff:
		return "BACKOFF"
	default:
		return "UNKNOWN"
	}
}

// TransitionFn is a function that can be called when a state transition occurs.
type TransitionFn func(from, to State)

// Machine manages the lifecycle states for MQTT connections.
type Machine struct {
	mu      sync.RWMutex
	current State
	readyCh chan struct{}
	stopped chan struct{}

	// Retry configuration
	minBackoff      time.Duration
	maxBackoff      time.Duration
	currentBackoff  time.Duration

	// Callbacks
	onTransition TransitionFn
}

// NewMachine creates a new state machine with default settings.
func NewMachine() *Machine {
	return &Machine{
		current:        StateDisconnected,
		readyCh:        make(chan struct{}),
		stopped:        make(chan struct{}),
		minBackoff:     1 * time.Second,
		maxBackoff:     60 * time.Second,
		currentBackoff: 1 * time.Second,
	}
}

// WithTransitionCallback sets a callback for state transitions.
func (m *Machine) WithTransitionCallback(fn TransitionFn) *Machine {
	m.onTransition = fn
	return m
}

// Current returns the current state.
func (m *Machine) Current() State {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.current
}

// Connected returns true if the machine is in the Connected state.
func (m *Machine) Connected() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.current == StateConnected
}

// Ready returns a channel that is closed when the machine becomes connected.
func (m *Machine) Ready() <-chan struct{} {
	return m.readyCh
}

// Stopped returns a channel that is closed when the machine has stopped.
func (m *Machine) Stopped() <-chan struct{} {
	return m.stopped
}

// Transition attempts to change the state. Returns true if transition was valid.
func (m *Machine) Transition(to State) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Validate transitions
	if !m.isValidTransition(m.current, to) {
		return false
	}

	from := m.current
	m.current = to

	// Update ready channel on Connected state
	if to == StateConnected && from != StateConnected {
		select {
		case <-m.readyCh:
		default:
			close(m.readyCh)
		}
	}

	// Update stopped channel
	if to == StateDisconnected || to == StateBackoff {
		select {
		case <-m.stopped:
		default:
			close(m.stopped)
		}
	}

	// Execute callback
	if m.onTransition != nil {
		go m.onTransition(from, to)
	}

	return true
}

func (m *Machine) isValidTransition(from, to State) bool {
	switch from {
	case StateDisconnected:
		return to == StateConnecting
	case StateConnecting:
		return to == StateConnected || to == StateBackoff
	case StateConnected:
		return to == StateDisconnected || to == StateBackoff || to == StateReconnecting
	case StateReconnecting:
		return to == StateConnecting || to == StateBackoff
	case StateBackoff:
		return to == StateConnecting || to == StateDisconnected
	}
	return false
}

// BackoffDuration returns the current backoff duration and updates it exponentially.
func (m *Machine) BackoffDuration() time.Duration {
	m.mu.Lock()
	defer m.mu.Unlock()

	duration := m.currentBackoff
	m.currentBackoff *= 2
	if m.currentBackoff > m.maxBackoff {
		m.currentBackoff = m.maxBackoff
	}
	return duration
}

// ResetBackoff resets the backoff duration to the minimum.
func (m *Machine) ResetBackoff() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.currentBackoff = m.minBackoff
}

// WaitTransition waits for the state machine to reach a specific state.
func (m *Machine) WaitTransition(ctx context.Context, target State) bool {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return false
		case <-ticker.C:
			if m.Current() == target {
				return true
			}
		}
	}
}

// WaitConnected waits for the connection to be established.
func (m *Machine) WaitConnected(ctx context.Context) bool {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return false
		case <-m.Ready():
			return m.Connected()
		}
	}
}
