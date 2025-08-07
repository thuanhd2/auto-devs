package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// RedisBroker handles Redis Pub/Sub for WebSocket messaging
type RedisBroker struct {
	client    *redis.Client
	pubsub    *redis.PubSub
	hub       *Hub
	logger    *slog.Logger
	ctx       context.Context
	cancel    context.CancelFunc
	channel   string
	isRunning bool
}

// BrokerMessage represents a message sent through Redis broker
type BrokerMessage struct {
	Type      MessageType     `json:"type"`
	Data      json.RawMessage `json:"data"`
	ProjectID *uuid.UUID      `json:"project_id,omitempty"`
	UserID    *string         `json:"user_id,omitempty"`
	Timestamp time.Time       `json:"timestamp"`
	MessageID string          `json:"message_id"`
	Source    string          `json:"source"` // "worker", "server", etc.
}

// NewRedisBroker creates a new Redis broker
func NewRedisBroker(redisAddr, redisPassword string, db int, hub *Hub) *RedisBroker {
	ctx, cancel := context.WithCancel(context.Background())

	client := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       db,
	})

	return &RedisBroker{
		client:  client,
		hub:     hub,
		logger:  slog.Default().With("component", "redis-broker"),
		ctx:     ctx,
		cancel:  cancel,
		channel: "websocket:broadcast",
	}
}

// Start starts the Redis broker
func (b *RedisBroker) Start() error {
	if b.isRunning {
		return fmt.Errorf("broker is already running")
	}

	b.logger.Info("Starting Redis broker", "channel", b.channel)

	// Create pubsub
	b.pubsub = b.client.Subscribe(b.ctx, b.channel)

	// Test connection
	if err := b.pubsub.Ping(b.ctx); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	b.isRunning = true

	// Start listening for messages
	go b.listenForMessages()

	b.logger.Info("Redis broker started successfully")
	return nil
}

// Stop stops the Redis broker
func (b *RedisBroker) Stop() error {
	if !b.isRunning {
		return nil
	}

	b.logger.Info("Stopping Redis broker")

	b.cancel()

	if b.pubsub != nil {
		if err := b.pubsub.Close(); err != nil {
			b.logger.Error("Failed to close pubsub", "error", err)
		}
	}

	if err := b.client.Close(); err != nil {
		b.logger.Error("Failed to close Redis client", "error", err)
	}

	b.isRunning = false
	b.logger.Info("Redis broker stopped")
	return nil
}

// PublishMessage publishes a message to Redis
func (b *RedisBroker) PublishMessage(message *BrokerMessage) error {
	if !b.isRunning {
		return fmt.Errorf("broker is not running")
	}

	// Marshal message
	messageBytes, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Publish to Redis
	result := b.client.Publish(b.ctx, b.channel, messageBytes)
	if err := result.Err(); err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	b.logger.Debug("Published message to Redis",
		"message_id", message.MessageID,
		"type", message.Type,
		"recipients", result.Val())

	return nil
}

// PublishTaskUpdated publishes a task updated message
func (b *RedisBroker) PublishTaskUpdated(taskID, projectID uuid.UUID, changes map[string]interface{}, task interface{}) error {
	data := map[string]interface{}{
		"task_id":    taskID.String(),
		"project_id": projectID.String(),
		"changes":    changes,
		"task":       task,
	}

	dataBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal task data: %w", err)
	}

	message := &BrokerMessage{
		Type:      TaskUpdated,
		Data:      dataBytes,
		ProjectID: &projectID,
		Timestamp: time.Now(),
		MessageID: uuid.New().String(),
		Source:    "worker",
	}

	return b.PublishMessage(message)
}

// PublishStatusChanged publishes a status changed message
func (b *RedisBroker) PublishStatusChanged(entityID, projectID uuid.UUID, entityType, oldStatus, newStatus string) error {
	data := map[string]interface{}{
		"entity_id":   entityID.String(),
		"project_id":  projectID.String(),
		"entity_type": entityType,
		"old_status":  oldStatus,
		"new_status":  newStatus,
	}

	dataBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal status data: %w", err)
	}

	message := &BrokerMessage{
		Type:      StatusChanged,
		Data:      dataBytes,
		ProjectID: &projectID,
		Timestamp: time.Now(),
		MessageID: uuid.New().String(),
		Source:    "worker",
	}

	return b.PublishMessage(message)
}

// listenForMessages listens for messages from Redis
func (b *RedisBroker) listenForMessages() {
	b.logger.Info("Listening for messages on channel", "channel", b.channel)

	ch := b.pubsub.Channel()
	for {
		select {
		case <-b.ctx.Done():
			b.logger.Info("Stopping message listener")
			return
		case msg := <-ch:
			b.handleRedisMessage(msg)
		}
	}
}

// handleRedisMessage handles incoming Redis messages
func (b *RedisBroker) handleRedisMessage(redisMsg *redis.Message) {
	b.logger.Debug("Received Redis message", "channel", redisMsg.Channel, "payload", redisMsg.Payload)

	// Parse broker message
	var brokerMsg BrokerMessage
	if err := json.Unmarshal([]byte(redisMsg.Payload), &brokerMsg); err != nil {
		b.logger.Error("Failed to unmarshal broker message", "error", err)
		return
	}

	// Convert to WebSocket message
	wsMessage := &Message{
		Type:      brokerMsg.Type,
		Data:      brokerMsg.Data,
		Timestamp: brokerMsg.Timestamp,
		MessageID: brokerMsg.MessageID,
	}

	// Broadcast to hub
	switch {
	case brokerMsg.ProjectID != nil:
		b.hub.BroadcastToProject(wsMessage, *brokerMsg.ProjectID, nil)
	case brokerMsg.UserID != nil:
		b.hub.BroadcastToUser(wsMessage, *brokerMsg.UserID, nil)
	default:
		b.hub.BroadcastToAll(wsMessage, nil)
	}

	b.logger.Debug("Broadcasted message from Redis",
		"message_id", brokerMsg.MessageID,
		"type", brokerMsg.Type,
		"source", brokerMsg.Source)
}

// IsRunning returns true if the broker is running
func (b *RedisBroker) IsRunning() bool {
	return b.isRunning
}

// GetStats returns broker statistics
func (b *RedisBroker) GetStats() map[string]interface{} {
	stats := map[string]interface{}{
		"is_running": b.isRunning,
		"channel":    b.channel,
	}

	if b.client != nil {
		stats["redis_connected"] = b.client.Ping(b.ctx).Err() == nil
	}

	return stats
}
