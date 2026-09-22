package state

import (
	"testing"
)

func BenchmarkStateTransition(b *testing.B) {
	m := NewMachine()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		m.Transition(StateConnecting)
		m.Transition(StateConnected)
		m.Transition(StateBackoff)
	}
}

func BenchmarkStateBackoffCalculation(b *testing.B) {
	m := NewMachine()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = m.BackoffDuration()
	}
}

func BenchmarkStateBackoffWithJitter(b *testing.B) {
	m := NewMachine()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = m.NextBackoffWithJitter()
	}
}

func BenchmarkStateMetricsConcurrent(b *testing.B) {
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
				m.BackoffDuration()
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	b.ReportAllocs()
}
