package publisher

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/dvandagriff/Carrier-Pigeons/carrierpigeons/internal/state"
)

// errNotConnected is returned by testPublisher.Publish when not connected.
var errNotConnected = context.Canceled

// testPublisher is a zero-dependency mock implementing Publisher for
// testing the scheduler and other dependents without an MQTT broker.
type testPublisher struct {
	ctx          context.Context
	cancel       context.CancelFunc
	connected    bool
	Published    []Metric
	PublishErr   error
	failNext     int
	connectedCh  chan struct{}
	once         bool
	stateMachine *state.Machine
}

func newTestPublisher() *testPublisher {
	ctx, cancel := context.WithCancel(context.Background())
	return &testPublisher{
		ctx:          ctx,
		cancel:       cancel,
		connectedCh:  make(chan struct{}),
		stateMachine: state.NewMachine(),
	}
}

func (p *testPublisher) Start(ctx context.Context) error { return nil }
func (p *testPublisher) Stop(ctx context.Context) error  { return nil }

func (p *testPublisher) Publish(ctx context.Context, metric Metric) error {
	if !p.connected {
		return errNotConnected
	}
	// Emit the pre-programmed number of transient failures first, then recover.
	if p.failNext > 0 {
		p.failNext--
		if p.PublishErr != nil {
			return p.PublishErr
		}
		return context.DeadlineExceeded
	}
	p.Published = append(p.Published, metric)
	return nil
}

func (p *testPublisher) IsConnected() bool                { return p.connected }
func (p *testPublisher) ConnectionReady() <-chan struct{} { return p.connectedCh }

func (p *testPublisher) StateMachine() *state.Machine { return p.stateMachine }

func (p *testPublisher) connect() {
	if !p.once {
		p.connected = true
		p.once = true
		close(p.connectedCh)
	}
}

func (p *testPublisher) waitReady(t *testing.T) {
	t.Helper()
	select {
	case <-p.connectedCh:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for publisher to be connected")
	}
}

func TestMetric_MarshalJSON(t *testing.T) {
	m := Metric{
		NodeID:     "node-42",
		Timestamp:  "2024-05-01T12:00:00Z",
		MetricType: "cpu_usage",
		Value:      12.5,
		Metadata:   map[string]string{"host": "edge-1"},
	}

	data, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal round-trip error = %v", err)
	}

	if got["node_id"] != "node-42" {
		t.Errorf("node_id = %v, want node-42", got["node_id"])
	}
	if got["metric_type"] != "cpu_usage" {
		t.Errorf("metric_type = %v, want cpu_usage", got["metric_type"])
	}
	if got["value"] != float64(12.5) {
		t.Errorf("value = %v, want 12.5", got["value"])
	}
	meta, ok := got["metadata"].(map[string]any)
	if !ok || meta["host"] != "edge-1" {
		t.Errorf("metadata = %v, want {host: edge-1}", got["metadata"])
	}
}

func TestMetric_UnmarshalJSON(t *testing.T) {
	src := `{
		"node_id": "node-7",
		"timestamp": "2024-05-01T12:00:00Z",
		"metric_type": "memory_usage",
		"value": 88.1,
		"metadata": {"region": "us-east"}
	}`

	var got Metric
	if err := json.Unmarshal([]byte(src), &got); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	want := Metric{
		NodeID:     "node-7",
		Timestamp:  "2024-05-01T12:00:00Z",
		MetricType: "memory_usage",
		Value:      88.1,
		Metadata:   map[string]string{"region": "us-east"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Metric = %+v, want %+v", got, want)
	}
}

func TestNewMQTTPublisher_Tools(t *testing.T) {
	p := NewMQTTPublisher(&PublisherConfig{})
	if p == nil {
		t.Fatal("NewMQTTPublisher returned nil")
	}
	if p.config.Topic != "telemetry/+" {
		t.Errorf("default Topic = %q, want telemetry/+", p.config.Topic)
	}
	if p.config.KeepAlive != 30 {
		t.Errorf("default KeepAlive = %d, want 30", p.config.KeepAlive)
	}
	if p.ConnectionReady() == nil {
		t.Error("ConnectionReady() returned nil channel")
	}
	if p.StateMachine() == nil {
		t.Error("StateMachine() returned nil")
	}
	if got := p.StateMachine().Current(); got != state.StateDisconnected {
		t.Errorf("initial state = %d, want StateDisconnected", got)
	}
}

func TestObscureURL(t *testing.T) {
	cases := []struct {
		name   string
		in     string
		expect string
	}{
		{"no credentials", "tcp://localhost:1883", "tcp://localhost:1883"},
		{"with credentials", "tcp://user:pass@localhost:1883", "tcp://***:***@localhost:1883"},
		{"empty", "", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := NewMQTTPublisher(&PublisherConfig{})
			got := p.obscureURL(tc.in)
			if got != tc.expect {
				t.Errorf("obscureURL(%q) = %q, want %q", tc.in, got, tc.expect)
			}
		})
	}
}

func TestMockPublisher_PublishRecords(t *testing.T) {
	p := newTestPublisher()
	p.connect()
	p.waitReady(t)

	m := Metric{NodeID: "n1", MetricType: "cpu_usage", Value: 1}
	if err := p.Publish(context.Background(), m); err != nil {
		t.Fatalf("Publish() error = %v", err)
	}
	if len(p.Published) != 1 || p.Published[0].NodeID != "n1" {
		t.Fatalf("Published = %+v, want 1 metric node-1", p.Published)
	}
}

func TestMockPublisher_UnconnectedFails(t *testing.T) {
	p := newTestPublisher()
	if err := p.Publish(context.Background(), Metric{}); err == nil {
		t.Fatal("expected error when unconnected")
	}
}

func TestMockPublisher_FlakyPublishRecovers(t *testing.T) {
	p := newTestPublisher()
	p.connect()
	p.failNext = 2
	p.PublishErr = context.DeadlineExceeded
	p.waitReady(t)

	if err := p.Publish(context.Background(), Metric{}); err == nil {
		t.Fatal("expected first flaky publish to fail")
	}
	if err := p.Publish(context.Background(), Metric{}); err == nil {
		t.Fatal("expected second flaky publish to fail")
	}
	if err := p.Publish(context.Background(), Metric{}); err != nil {
		t.Fatalf("expected third publish to succeed, got %v", err)
	}
}
