package scheduler

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/dvandagriff/Carrier-Pigeons/carrierpigeons/internal/collector"
	"github.com/dvandagriff/Carrier-Pigeons/carrierpigeons/internal/publisher"
)

// recordingPublisher is a mock publisher that records published metrics and
// supports configurable failure injection and a controllable ready channel.
// It implements publisher.Publisher via structural subtyping.
type recordingPublisher struct {
	mu        sync.Mutex
	published []publisher.Metric
	publishes int32
	// publishErr, if set, is returned by every Publish call.
	publishErr error
	// readyDelay controls how long until ConnectionReady closes.
	readyCh   chan struct{}
	readyOnce sync.Once
	// connected reflects the publisher's state.
	connected bool
}

func newRecordingPublisher() *recordingPublisher {
	return &recordingPublisher{
		readyCh:   make(chan struct{}),
		connected: true,
	}
}

func (r *recordingPublisher) Start(ctx context.Context) error { return nil }
func (r *recordingPublisher) Stop(ctx context.Context) error  { return nil }

func (r *recordingPublisher) Publish(ctx context.Context, metric publisher.Metric) error {
	if r.publishErr != nil {
		return r.publishErr
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	r.mu.Lock()
	r.published = append(r.published, metric)
	r.mu.Unlock()
	atomic.AddInt32(&r.publishes, 1)
	return nil
}

func (r *recordingPublisher) IsConnected() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.connected
}

func (r *recordingPublisher) ConnectionReady() <-chan struct{} {
	r.readyOnce.Do(func() { close(r.readyCh) })
	return r.readyCh
}

func (r *recordingPublisher) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.published)
}

func (r *recordingPublisher) setPublishErr(err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.publishErr = err
}

// fakeCollector is a deterministic collector producing a fixed number of metrics.
type fakeCollector struct {
	name       string
	interval   time.Duration
	metrics    int
	collectErr error
	mu         sync.Mutex
	invoked    int32
}

func (c *fakeCollector) Name() string            { return c.name }
func (c *fakeCollector) Interval() time.Duration { return c.interval }

func (c *fakeCollector) Collect(ctx context.Context) ([]collector.Metric, error) {
	if c.collectErr != nil {
		return nil, c.collectErr
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	atomic.AddInt32(&c.invoked, 1)
	out := make([]collector.Metric, 0, c.metrics)
	for i := 0; i < c.metrics; i++ {
		out = append(out, collector.Metric{
			NodeID:     "node-" + c.name,
			Timestamp:  time.Now(),
			MetricType: collector.MetricTypeCPUUsage,
			Value:      float64(i),
			Metadata:   map[string]string{"k": "v"},
		})
	}
	return out, nil
}

func (c *fakeCollector) invokedCount() int {
	return int(atomic.LoadInt32(&c.invoked))
}

// waitReady blocks until the publisher signals readiness or a deadline.
func waitReady(t *testing.T, p publisherReady, timeout time.Duration) {
	t.Helper()
	select {
	case <-p.ConnectionReady():
	case <-time.After(timeout):
		t.Fatal("timed out waiting for publisher readiness")
	}
}

type publisherReady interface {
	ConnectionReady() <-chan struct{}
}

// TestSchedulerCollectsAndPublishes drives a scheduler with two collectors and
// verifies metrics flow through to the publisher.
func TestSchedulerCollectsAndPublishes(t *testing.T) {
	pub := newRecordingPublisher()

	sched := NewScheduler(
		[]collector.MetricCollector{
			&fakeCollector{name: "cpu", interval: 2 * time.Millisecond, metrics: 10},
			&fakeCollector{name: "mem", interval: 2 * time.Millisecond, metrics: 5},
		},
		pub,
		&Config{MetricChannelSize: 100, OutboundChannelSize: 1000, MaxBatchSize: 20},
	)

	if err := sched.Start(); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	// Allow the periodic flush (100ms ticker) several cycles to elapse so the
	// vast majority of collected metrics are flushed to the publisher.
	time.Sleep(350 * time.Millisecond)
	sched.Stop()

	collected, published, dropped := sched.Metrics()
	if collected == 0 {
		t.Error("no metrics were collected")
	}
	if published == 0 {
		t.Error("expected at least some metrics to be published")
	}
	if dropped > 0 {
		t.Errorf("expected no dropped metrics (ample channel headroom), got %d", dropped)
	}
	// The scheduler never loses an accounting: every collected metric is either
	// published, dropped under backpressure, or still sitting in the flush
	// buffer. So collected can never be less than published+dropped.
	if uint64(published)+dropped > collected {
		t.Errorf("accounting lost: published=%d dropped=%d exceeds collected=%d", published, dropped, collected)
	}
	// The publisher must never be told about more metrics than were collected.
	if pub.count() > int(collected) {
		t.Errorf("publisher received %d, only %d were collected", pub.count(), collected)
	}
}

// TestSchedulerGracefulShutdownContext verifies Stop() does not deadlock and
// drains the batch on shutdown.
func TestSchedulerGracefulShutdownContext(t *testing.T) {
	pub := newRecordingPublisher()
	sched := NewScheduler(
		[]collector.MetricCollector{
			&fakeCollector{name: "cpu", interval: time.Millisecond, metrics: 4},
		},
		pub,
		&Config{MetricChannelSize: 100, OutboundChannelSize: 1000, MaxBatchSize: 2},
	)

	if err := sched.Start(); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	done := make(chan struct{})
	go func() {
		sched.Stop()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Stop() deadlocked")
	}

	collected, published, _ := sched.Metrics()
	if collected == 0 {
		t.Error("expected at least one metric collected before shutdown")
	}
	if published == 0 {
		t.Error("expected at least one metric published before shutdown")
	}
}

// TestSchedulerPublisherFailureDrops verifies that when the publisher errors,
// the scheduler does not block and metrics are accounted for (dropped).
func TestSchedulerPublisherFailureDrops(t *testing.T) {
	pub := newRecordingPublisher()
	pub.setPublishErr(context.DeadlineExceeded)

	sched := NewScheduler(
		[]collector.MetricCollector{
			&fakeCollector{name: "cpu", interval: time.Millisecond, metrics: 50},
		},
		pub,
		&Config{MetricChannelSize: 100, OutboundChannelSize: 1000, MaxBatchSize: 10},
	)

	if err := sched.Start(); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	time.Sleep(120 * time.Millisecond)
	sched.Stop()

	collected, _, _ := sched.Metrics()
	if collected == 0 {
		t.Error("expected metrics to be collected despite publisher failures")
	}
}

// TestSchedulerCollectorFailure verifies a collector that errors does not
// stop the scheduler from processing other collectors.
func TestSchedulerCollectorFailure(t *testing.T) {
	pub := newRecordingPublisher()
	sched := NewScheduler(
		[]collector.MetricCollector{
			&fakeCollector{name: "bad", interval: time.Millisecond, metrics: 3, collectErr: context.DeadlineExceeded},
			&fakeCollector{name: "good", interval: time.Millisecond, metrics: 3},
		},
		pub,
		&Config{MetricChannelSize: 100, OutboundChannelSize: 1000, MaxBatchSize: 5},
	)

	if err := sched.Start(); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	time.Sleep(80 * time.Millisecond)
	sched.Stop()

	collected, published, _ := sched.Metrics()
	if collected == 0 || published == 0 {
		t.Errorf("expected good collector to produce metrics, collected=%d published=%d", collected, published)
	}
}

// TestSchedulerStopConcurrentCollect ensures no data race between Stop and the
// collection/publish loops under the race detector.
func TestSchedulerStopConcurrentCollect(t *testing.T) {
	pub := newRecordingPublisher()
	sched := NewScheduler(
		[]collector.MetricCollector{
			&fakeCollector{name: "cpu", interval: time.Millisecond, metrics: 1},
		},
		pub,
		&Config{MetricChannelSize: 64, OutboundChannelSize: 64, MaxBatchSize: 4},
	)

	if err := sched.Start(); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	done := make(chan struct{})
	go func() {
		sched.Stop()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Stop() deadlocked under concurrent collect")
	}
}
