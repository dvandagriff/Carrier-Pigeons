// Package publisher_test contains integration tests that bind the real
// MQTTPublisher to a genuine embedded Mochi MQTT broker.
package publisher_test

import (
	"context"
	"log/slog"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/dvantagriff/Carrier-Pigeons/carrierpigeons/internal/publisher"

	"github.com/eclipse/paho.golang/paho"
	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/listeners"
)

// embeddedBroker is a thin wrapper around the Mochi MQTT server that manages
// its lifecycle for tests.
type embeddedBroker struct {
	server *mqtt.Server
	addr   string
}

// startEmbeddedBroker boots an in-process Mochi broker on a random free port
// and returns it. The returned stop function cleanly shuts the broker down.
func startEmbeddedBroker(t *testing.T) *embeddedBroker {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to allocate listener port: %v", err)
	}
	addr := ln.Addr().String()

	s := mqtt.New(&mqtt.Options{
		// Silence the broker's logger unless a test asks to see it.
		Logger: slog.New(slog.NewTextHandler(nil, &slog.HandlerOptions{Level: slog.LevelError})),
	})

	tc := listeners.Config{
		Type:    listeners.TypeTCP,
		ID:      "test-tcp",
		Address: addr,
	}
	if err := s.AddListener(listeners.NewTCP(tc)); err != nil {
		ln.Close()
		t.Fatalf("failed to add TCP listener: %v", err)
	}

	// We need a real listener to serve on; swap in the one bound to our socket.
	// mochi creates its own listener from Config.Address on Serve, so close the
	// temp socket and let mochi bind `addr` itself.
	ln.Close()
	if err := s.Serve(); err != nil {
		t.Fatalf("broker failed to serve: %v", err)
	}

	// Give the event loop a moment to bind the listener.
	for i := 0; i < 100; i++ {
		// Best-effort: confirm the port is accepting connections.
		conn, cerr := net.DialTimeout("tcp", addr, 200*time.Millisecond)
		if cerr == nil {
			conn.Close()
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	return &embeddedBroker{server: s, addr: addr}
}

func (b *embeddedBroker) addrNow() string { return b.addr }

func (b *embeddedBroker) stop(t *testing.T) {
	t.Helper()
	// Mochi's Server.Close closes listeners and the done channel.
	if err := b.server.Close(); err != nil {
		t.Logf("broker close returned error (ignored): %v", err)
	}
}

// pubSubRoundTrip verifies that a metric published by a real MQTTPublisher
// bound to the embedded broker is delivered to a real paho subscriber on the
// same topic.
func TestMQTTPublisher_RealBrokerRoundTrip(t *testing.T) {
	broker := startEmbeddedBroker(t)
	defer broker.stop(t)

	pubCfg := &publisher.PublisherConfig{
		BrokerURL: "tcp://" + broker.addrNow(),
		ClientID:  "cp-roundtrip",
		Topic:     "telemetry/test",
	}
	pub := publisher.NewMQTTPublisher(pubCfg)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := pub.Start(ctx); err != nil {
		t.Fatalf("publisher failed to start against embedded broker: %v", err)
	}
	defer func() {
		stopCtx, c := context.WithTimeout(context.Background(), 5*time.Second)
		defer c()
		_ = pub.Stop(stopCtx)
	}()

	if !pub.IsConnected() {
		t.Fatal("expected publisher to be connected after Start")
	}

	// Real paho subscriber to receive the published metric.
	sub := newPahoClient(t, "cp-subscriber")
	defer sub.disconnect(t)

	if err := sub.connect(ctx); err != nil {
		t.Fatalf("subscriber failed to connect: %v", err)
	}

	received := make(chan publisher.Metric, 1)
	sub.subscribe(ctx, "telemetry/test", 1, func(_ context.Context, p *paho.Publish) {
		var m publisher.Metric
		if err := mustUnmarshal(p.Payload, &m); err == nil {
			select {
			case received <- m:
			default:
			}
		}
	})

	pubMetric := publisher.Metric{
		NodeID:     "node-int-1",
		Timestamp:  time.Now().UTC().Format(time.RFC3339Nano),
		MetricType: "cpu_usage",
		Value:      77.5,
		Metadata:   map[string]string{"host": "edge-int"},
	}
	if err := pub.Publish(ctx, pubMetric); err != nil {
		t.Fatalf("publish failed: %v", err)
	}

	select {
	case got := <-received:
		if got.NodeID != pubMetric.NodeID || got.Value != pubMetric.Value || got.MetricType != "cpu_usage" {
			t.Errorf("received metric %+v does not match published %+v", got, pubMetric)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for metric to be delivered to subscriber")
	}
}

// TestMQTTPublisher_StartConnectionTimeout verifies Start returns an error
// when the broker URL points at a closed port.
func TestMQTTPublisher_StartConnectionTimeout(t *testing.T) {
	cfg := &publisher.PublisherConfig{
		BrokerURL: "tcp://127.0.0.1:1", // port 1 is not connectable
		ClientID:  "cp-timeout",
		Topic:     "telemetry/test",
	}
	pub := publisher.NewMQTTPublisher(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := pub.Start(ctx); err == nil {
		t.Fatal("expected Start to fail against unreachable broker")
	}
}

// startFreeBroker starts an embedded broker and blocks until a client can dial
// the bound address, returning the address. Used by other helper-based tests.
func startFreeBroker(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listener alloc: %v", err)
	}
	addr := ln.Addr().String()
	ln.Close()

	s := mqtt.New(nil)
	if err := s.AddListener(listeners.NewTCP(listeners.Config{
		Type:    listeners.TypeTCP,
		ID:      "t",
		Address: addr,
	})); err != nil {
		t.Fatalf("add listener: %v", err)
	}
	if err := s.Serve(); err != nil {
		t.Fatalf("serve: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	for i := 0; i < 100; i++ {
		conn, cerr := net.DialTimeout("tcp", addr, 100*time.Millisecond)
		if cerr == nil {
			conn.Close()
			return addr
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("broker never became ready at %s", addr)
	return ""
}

// ---- minimal paho test helper ----

type pahoTestClient struct {
	client *paho.Client
}

func newPahoClient(t *testing.T, id string) *pahoTestClient {
	t.Helper()
	conf := paho.ClientConfig{ClientID: id, PingHandler: paho.NewDefaultPinger()}
	return &pahoTestClient{client: paho.NewClient(conf)}
}

func (c *pahoTestClient) connect(ctx context.Context) error {
	resp, err := c.client.Connect(ctx, &paho.Connect{KeepAlive: 5, CleanStart: true})
	if err != nil {
		return err
	}
	return resp.Error()
}

func (c *pahoTestClient) subscribe(ctx context.Context, topic string, qos byte, h func(ctx context.Context, p *paho.Publish)) {
	_ = c.client.AddOnPublishReceived(func(pr paho.PublishReceived) (bool, error) {
		p := pr.Packet
		h(context.Background(), &p)
		return true, nil
	})
	_, err := c.client.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{{Topic: topic, QoS: qos}},
	})
	if err != nil {
		// best effort; tests that care will fail on delivery
	}
}

func (c *pahoTestClient) disconnect(t *testing.T) {
	t.Helper()
	_ = c.client.Disconnect(&paho.Disconnect{})
}

// mustUnmarshal decodes JSON into v; returns error on failure.
func mustUnmarshal(data []byte, v any) error {
	var mu sync.Mutex
	return jsonUnmarshal(data, v, mu)
}
