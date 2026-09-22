package state

import (
	"context"
	"testing"
	"time"
)

func TestNewMachine(t *testing.T) {
	m := NewMachine()
	
	if m.Current() != StateDisconnected {
		t.Errorf("Expected initial state to be StateDisconnected, got %v", m.Current())
	}
}

func TestTransition(t *testing.T) {
	m := NewMachine()
	
	// Test valid transitions
	if !m.Transition(StateConnecting) {
		t.Error("Expected transition from Disconnected to Connecting to succeed")
	}
	if m.Current() != StateConnecting {
		t.Errorf("Expected state to be StateConnecting, got %v", m.Current())
	}
	
	if !m.Transition(StateConnected) {
		t.Error("Expected transition from Connecting to Connected to succeed")
	}
	if m.Current() != StateConnected {
		t.Errorf("Expected state to be StateConnected, got %v", m.Current())
	}
	
	// Test transitions from Connected
	if !m.Transition(StateBackoff) {
		t.Error("Expected transition from Connected to Backoff to succeed")
	}
	if m.Current() != StateBackoff {
		t.Errorf("Expected state to be StateBackoff, got %v", m.Current())
	}
}

func TestBackoffDuration(t *testing.T) {
	m := NewMachine()
	
	// First backoff should be minBackoff
	duration := m.BackoffDuration()
	if duration != m.minBackoff {
		t.Errorf("Expected backoff duration to be %v, got %v", m.minBackoff, duration)
	}
	
	// Second backoff should be 2x
	duration = m.BackoffDuration()
	expected := m.minBackoff * 2
	if duration != expected {
		t.Errorf("Expected backoff duration to be %v, got %v", expected, duration)
	}
	
	// Thrid backoff should be 4x (capped at maxBackoff)
	duration = m.BackoffDuration()
	expected = m.minBackoff * 4
	if expected > m.maxBackoff {
		expected = m.maxBackoff
	}
	if duration != expected {
		t.Errorf("Expected backoff duration to be %v, got %v", expected, duration)
	}
}

func TestResetBackoff(t *testing.T) {
	m := NewMachine()
	
	// Advance backoff
	m.BackoffDuration()
	m.BackoffDuration()
	
	// Reset
	m.ResetBackoff()
	
	// Next backoff should be minBackoff again
	duration := m.BackoffDuration()
	if duration != m.minBackoff {
		t.Errorf("Expected backoff duration to be %v after reset, got %v", m.minBackoff, duration)
	}
}

func TestReadyChannel(t *testing.T) {
	m := NewMachine()
	
	// Should not be ready initially
	select {
	case <-m.Ready():
		t.Error("Expected Ready channel to not be closed initially")
	default:
	}
	
	// Transition to connected
	m.Transition(StateConnecting)
	m.Transition(StateConnected)
	
	// Should be ready now
	select {
	case <-m.Ready():
		// Expected
	default:
		t.Error("Expected Ready channel to be closed after Connected transition")
	}
}

func TestWaitConnected(t *testing.T) {
	m := NewMachine()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	
	// Should not be connected initially
	if m.WaitConnected(ctx) {
		t.Error("Expected WaitConnected to return false initially")
	}
	
	// Now transition
	m.Transition(StateConnecting)
	m.Transition(StateConnected)
	
	// Should be connected now
	ctx2, cancel2 := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel2()
	if !m.WaitConnected(ctx2) {
		t.Error("Expected WaitConnected to return true after Connected transition")
	}
}

func TestMachineMetrics(t *testing.T) {
	m := NewMachine()
	
	if m.ConnectionAttempts() != 0 {
		t.Errorf("Expected zero connection attempts initially, got %d", m.ConnectionAttempts())
	}
	
	// Simulate connection attempts
	m.Transition(StateConnecting)
	m.Transition(StateBackoff)
	m.Transition(StateConnecting)
	m.Transition(StateConnected)
	
	if m.ConnectionAttempts() != 2 {
		t.Errorf("Expected 2 connection attempts, got %d", m.ConnectionAttempts())
	}
	
	if m.ConnectionFailures() != 0 {
		t.Errorf("Expected zero connection failures after successful connection, got %d", m.ConnectionFailures())
	}
}

func TestStateMachineConcurrentAccess(t *testing.T) {
	m := NewMachine()
	done := make(chan bool)
	
	// Concurrent reads and writes
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				_ = m.Current()
				_ = m.Connected()
				m.Transition(StateConnecting)
				m.Transition(StateConnected)
				m.Transition(StateBackoff)
				m.Transition(StateConnecting)
				m.BackoffDuration()
			}
			done <- true
		}()
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}
