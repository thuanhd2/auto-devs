package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// RedisBrokerClient represents a Redis broker client for worker
type RedisBrokerClient struct {
	client  *redis.Client
	logger  *slog.Logger
	ctx     context.Context
	channel string
}

// BrokerMessage represents a message sent through Redis broker
type BrokerMessage struct {
	Type      string          `json:"type"`
	Data      json.RawMessage `json:"data"`
	ProjectID *uuid.UUID      `json:"project_id,omitempty"`
	UserID    *string         `json:"user_id,omitempty"`
	Timestamp time.Time       `json:"timestamp"`
	MessageID string          `json:"message_id"`
	Source    string          `json:"source"` // "worker", "server", etc.
}

// NewRedisBrokerClient creates a new Redis broker client
func NewRedisBrokerClient(redisAddr, redisPassword string, db int) *RedisBrokerClient {
	client := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       db,
	})

	return &RedisBrokerClient{
		client:  client,
		logger:  slog.Default().With("component", "redis-broker-client"),
		ctx:     context.Background(),
		channel: "websocket:broadcast",
	}
}

// Close closes the Redis client
func (c *RedisBrokerClient) Close() error {
	return c.client.Close()
}

// PublishMessage publishes a message to Redis
func (c *RedisBrokerClient) PublishMessage(message *BrokerMessage) error {
	// Marshal message
	messageBytes, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Publish to Redis
	result := c.client.Publish(c.ctx, c.channel, messageBytes)
	if err := result.Err(); err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	c.logger.Debug("Published message to Redis",
		"message_id", message.MessageID,
		"type", message.Type,
		"recipients", result.Val())

	return nil
}

// PublishTaskUpdated publishes a task updated message
func (c *RedisBrokerClient) PublishTaskUpdated(taskID, projectID uuid.UUID, changes map[string]interface{}, task interface{}) error {
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
		Type:      "task_updated",
		Data:      dataBytes,
		ProjectID: &projectID,
		Timestamp: time.Now(),
		MessageID: uuid.New().String(),
		Source:    "worker",
	}

	return c.PublishMessage(message)
}

// PublishStatusChanged publishes a status changed message
func (c *RedisBrokerClient) PublishStatusChanged(entityID, projectID uuid.UUID, entityType, oldStatus, newStatus string) error {
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
		Type:      "status_changed",
		Data:      dataBytes,
		ProjectID: &projectID,
		Timestamp: time.Now(),
		MessageID: uuid.New().String(),
		Source:    "worker",
	}

	return c.PublishMessage(message)
}

// TestConnection tests the Redis connection
func (c *RedisBrokerClient) TestConnection() error {
	return c.client.Ping(c.ctx).Err()
}
