// Package app provides the main application orchestrator.
package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dvandagriff/Carrier-Pigeons/pigeoncoop/internal/config"
	"github.com/dvandagriff/Carrier-Pigeons/pigeoncoop/internal/pubsub"
	"github.com/dvandagriff/Carrier-Pigeons/pigeoncoop/internal/state"
	"github.com/dvandagriff/Carrier-Pigeons/pigeoncoop/internal/storage"
	"github.com/dvandagriff/Carrier-Pigeons/pigeoncoop/internal/telemetry"
	"github.com/dvandagriff/Carrier-Pigeons/pigeoncoop/internal/worker"
)

// App is the main application orchestrator.
type App struct {
	config         config.Config
	state          *state.Machine
	pubSub         *pubsub.MQTTClient
	telemetryStore storage.TelemetryStore
	workerPool     *worker.WorkerPool
	ctx            context.Context
	cancel         context.CancelFunc
}

// New creates a new App instance.
func New(cfg config.Config) *App {
	ctx, cancel := context.WithCancel(context.Background())
	return &App{
		config: cfg,
		state:  state.NewMachine(),
		ctx:    ctx,
		cancel: cancel,
	}
}

// Start begins the application.
func (a *App) Start() error {
	// Initialize telemetry store
	slog.Info("initializing telemetry store")
	if err := a.initStorage(); err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	// Initialize MQTT client
	slog.Info("initializing MQTT client")
	if err := a.initPubSub(); err != nil {
		return fmt.Errorf("failed to initialize MQTT: %w", err)
	}

	// Initialize worker pool
	slog.Info("initializing worker pool")
	if err := a.initWorkerPool(); err != nil {
		return fmt.Errorf("failed to initialize worker pool: %w", err)
	}

	// Start worker pool
	if err := a.workerPool.Start(a.ctx); err != nil {
		return fmt.Errorf("failed to start worker pool: %w", err)
	}

	// Subscribe to topics
	slog.Info("subscribing to MQTT topics")
	if err := a.subscribeToTopics(); err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}

	// Transition to connected state
	a.state.Transition(state.StateConnected)

	slog.Info("application started successfully")

	// Wait for shutdown signal
	a.waitForShutdown()

	return nil
}

// Shutdown gracefully stops the application.
func (a *App) Shutdown() error {
	slog.Info("shutting down application")
	a.cancel()

	// Stop state machine
	a.state.Transition(state.StateStopping)

	// Stop MQTT client
	if a.pubSub != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := a.pubSub.Disconnect(ctx); err != nil {
			slog.Error("failed to disconnect MQTT", "error", err.Error())
		}
	}

	// Stop worker pool
	if a.workerPool != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := a.workerPool.Stop(ctx); err != nil {
			slog.Error("failed to stop worker pool", "error", err.Error())
		}
	}

	// Close storage
	if a.telemetryStore != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := a.telemetryStore.Close(ctx); err != nil {
			slog.Error("failed to close storage", "error", err.Error())
		}
	}

	a.state.Transition(state.StateStopped)

	slog.Info("application shutdown complete")
	return nil
}

func (a *App) initStorage() error {
	store, err := storage.NewPostgresStore(a.config.Storage.ToStoreConfig())
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(a.ctx, 30*time.Second)
	defer cancel()

	if err := store.Init(ctx); err != nil {
		return err
	}

	a.telemetryStore = store
	return nil
}

func (a *App) initPubSub() error {
	client := pubsub.NewMQTTClient(a.config.MQTT.ToClientConfig())

	ctx, cancel := context.WithTimeout(a.ctx, 30*time.Second)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		return err
	}

	a.pubSub = client

	return nil
}

func (a *App) initWorkerPool() error {
	a.workerPool = worker.NewWorkerPool(a.config.Worker.ToWorkerConfig(), a.telemetryStore)
	return nil
}

func (a *App) subscribeToTopics() error {
	topics := []string{
		"telemetry/#",
		"events/#",
	}

	for _, topic := range topics {
		err := a.pubSub.Subscribe(a.ctx, topic, 1, a.handleMQTTMessage)
		if err != nil {
			return fmt.Errorf("failed to subscribe to %s: %w", topic, err)
		}
		slog.Info("subscribed to topic", "topic", topic)
	}

	return nil
}

func (a *App) handleMQTTMessage(ctx context.Context, payload []byte) {
	// Parse the telemetry payload
	var telemetryPayload telemetry.TelemetryPayload
	if err := telemetryPayload.UnmarshalJSON(payload); err != nil {
		slog.Error("failed to parse telemetry payload", "error", err.Error())
		return
	}

	// Validate the payload
	if err := telemetryPayload.Validate(); err != nil {
		slog.Error("invalid telemetry payload", "error", err.Error())
		return
	}

	// Send to worker pool (non-blocking)
	select {
	case a.workerPool.Input() <- worker.TelemetryRecord{
		NodeID:     telemetryPayload.NodeID,
		Timestamp:  telemetryPayload.Timestamp,
		MetricType: telemetryPayload.MetricType,
		Value:      telemetryPayload.Value,
		Metadata:   telemetryPayload.Metadata,
	}:
		slog.Debug("telemetry record queued", "node_id", telemetryPayload.NodeID)
	default:
		slog.Error("worker pool channel full, dropping record")
	}
}

func (a *App) waitForShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	slog.Info("received shutdown signal", "signal", sig.String())
	a.cancel()
}
