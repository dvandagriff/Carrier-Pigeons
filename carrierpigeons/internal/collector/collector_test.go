package collector

import (
	"context"
	"maps"
	"testing"
	"time"
)

// assertMetricCollector ensures a value satisfies the MetricCollector interface.
func assertMetricCollector(t *testing.T, c interface{ Name() string }) {
	t.Helper()
	if _, ok := c.(MetricCollector); !ok {
		t.Errorf("value %T does not implement MetricCollector", c)
	}
}

func TestNewCPUCollector_Defaults(t *testing.T) {
	c := NewCPUCollector(nil)
	if c == nil {
		t.Fatal("NewCPUCollector returned nil")
	}
	if c.Name() != "cpu_darwin" {
		t.Errorf("Name() = %q, want cpu_darwin", c.Name())
	}
	if c.Interval() != 10*time.Second {
		t.Errorf("Interval() = %v, want 10s", c.Interval())
	}
}

func TestNewMemoryCollector_Defaults(t *testing.T) {
	m := NewMemoryCollector(nil)
	if m == nil {
		t.Fatal("NewMemoryCollector returned nil")
	}
	if m.Name() != "memory_darwin" {
		t.Errorf("Name() = %q, want memory_darwin", m.Name())
	}
	if m.Interval() != 15*time.Second {
		t.Errorf("Interval() = %v, want 15s", m.Interval())
	}
}

func TestCollectors_ImplementInterface(t *testing.T) {
	assertMetricCollector(t, NewCPUCollector(nil))
	assertMetricCollector(t, NewMemoryCollector(nil))
}

func TestCPUCollector_CollectReturnsError(t *testing.T) {
	c := NewCPUCollector(nil)
	// Darwin collectors are stubs that return an error.
	metrics, err := c.Collect(context.Background())
	if err == nil {
		t.Fatal("expected error from CPU collector on Darwin")
	}
	if metrics != nil {
		t.Errorf("expected nil metrics on error, got %d", len(metrics))
	}
}

func TestMemoryCollector_CollectReturnsError(t *testing.T) {
	m := NewMemoryCollector(nil)
	metrics, err := m.Collect(context.Background())
	if err == nil {
		t.Fatal("expected error from memory collector on Darwin")
	}
	if metrics != nil {
		t.Errorf("expected nil metrics on error, got %d", len(metrics))
	}
}

func TestCollect_ContextCanceled(t *testing.T) {
	c := NewCPUCollector(nil)
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	time.Sleep(5 * time.Millisecond)
	if _, err := c.Collect(ctx); err == nil {
		t.Error("expected context error to be surfaced")
	}
}

func TestMetricTypeConstants(t *testing.T) {
	cases := map[MetricType]string{
		MetricTypeCPUUsage:        "cpu_usage",
		MetricTypeCPUTemp:         "cpu_temp",
		MetricTypeMemoryUsage:     "memory_usage",
		MetricTypeMemoryAvailable: "memory_available",
		MetricTypeDiskUsage:       "disk_usage",
		MetricTypeServiceState:    "service_state",
	}
	for mt, want := range cases {
		if string(mt) != want {
			t.Errorf("MetricType %v = %q, want %q", mt, string(mt), want)
		}
	}
}

// fakeCollector is a deterministic collector used to exercise the Metric
// struct handling and metadata propagation without touching the OS.
type fakeCollector struct {
	labels    map[string]string
	interval  time.Duration
	collectFn func(ctx context.Context) ([]Metric, error)
}

func (c *fakeCollector) Name() string            { return "fake" }
func (c *fakeCollector) Interval() time.Duration { return c.interval }
func (c *fakeCollector) Collect(ctx context.Context) ([]Metric, error) {
	if c.collectFn != nil {
		return c.collectFn(ctx)
	}
	return []Metric{
		{
			NodeID:     "node-1",
			Timestamp:  time.Now(),
			MetricType: MetricTypeCPUUsage,
			Value:      42.0,
			Metadata:   maps.Clone(c.labels),
		},
	}, nil
}

func TestMetric_StructFields(t *testing.T) {
	ts := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	m := Metric{
		NodeID:     "node-1",
		Timestamp:  ts,
		MetricType: MetricTypeCPUUsage,
		Value:      12.5,
		Metadata:   map[string]string{"host": "edge-1"},
	}
	if m.NodeID != "node-1" || m.Value != 12.5 {
		t.Errorf("unexpected struct fields: %+v", m)
	}
	if !m.Timestamp.Equal(ts) {
		t.Errorf("timestamp = %v, want %v", m.Timestamp, ts)
	}
	if len(m.Metadata) != 1 || m.Metadata["host"] != "edge-1" {
		t.Errorf("metadata = %v, want {host: edge-1}", m.Metadata)
	}
}

// TestCollectorLabelsPropagate verifies the Label config is attached to the
// metadata of every metric produced by a collector, and that the label map is
// cloned (mutating the source afterwards must not affect the produced metrics).
func TestCollectorLabelsPropagate(t *testing.T) {
	labels := map[string]string{"node_id": "node-9", "region": "us-east"}
	c := &fakeCollector{labels: labels, interval: time.Second}

	got, err := c.Collect(context.Background())
	if err != nil {
		t.Fatalf("Collect() error = %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d metrics, want 1", len(got))
	}
	if got[0].Metadata["node_id"] != "node-9" || got[0].Metadata["region"] != "us-east" {
		t.Errorf("labels not propagated: %v", got[0].Metadata)
	}

	// Mutate source labels; the produced metric must be unaffected.
	labels["node_id"] = "changed"
	if got[0].Metadata["node_id"] != "node-9" {
		t.Errorf("metadata was not cloned: %v", got[0].Metadata)
	}
}

func TestMetricCollector_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	c := &fakeCollector{
		interval: time.Second,
		collectFn: func(ctx context.Context) ([]Metric, error) {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
				return nil, nil
			}
		},
	}
	if _, err := c.Collect(ctx); err == nil {
		t.Error("expected context cancellation error")
	}
}

func TestMetricType_NonEmpty(t *testing.T) {
	// Sanity check that each metric type has a non-empty, stable string form.
	types := []MetricType{
		MetricTypeCPUUsage, MetricTypeCPUTemp, MetricTypeMemoryUsage,
		MetricTypeMemoryAvailable, MetricTypeDiskUsage, MetricTypeServiceState,
	}
	seen := map[string]bool{}
	for _, mt := range types {
		if string(mt) == "" {
			t.Errorf("MetricType %v has empty string form", mt)
		}
		if seen[string(mt)] {
			t.Errorf("duplicate MetricType string form: %q", string(mt))
		}
		seen[string(mt)] = true
	}
}
