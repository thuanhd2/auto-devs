package websocket

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/centrifugal/centrifuge"
	"github.com/google/uuid"
)

// Hub maintains the set of active connections and broadcasts messages to them
type Hub struct {
	node *centrifuge.Node

	// Metrics
	metrics *HubMetrics

	// Mutex for thread-safe operations
	mu sync.RWMutex
}

// BroadcastMessage represents a message to be broadcasted
type BroadcastMessage struct {
	Message     *Message
	ProjectID   *uuid.UUID  // If set, only broadcast to this project
	UserID      *string     // If set, only broadcast to this user
	ExcludeConn *Connection // Exclude this connection from broadcast
}

// MessageProcessor defines the interface for processing specific message types
type MessageProcessor interface {
	ProcessMessage(conn *Connection, message *Message) error
}

// HubMetrics tracks hub statistics
type HubMetrics struct {
	TotalConnections   int64
	ActiveConnections  int64
	MessagesSent       int64
	MessagesReceived   int64
	BroadcastsSent     int64
	ConnectionsCreated int64
	ConnectionsClosed  int64
	mu                 sync.RWMutex
}

// NewHub creates a new Hub
func NewHub(node *centrifuge.Node) *Hub {
	hub := &Hub{
		node:    node,
		metrics: &HubMetrics{},
	}

	return hub
}

func generatePrivateChannel(_ *string, projectID *uuid.UUID) string {
	// hardCodeUserID := "123"
	// theUserID := hardCodeUserID
	// if userID != nil {
	// 	theUserID = *userID
	// }
	// if projectID != nil {
	// 	// $:<user_id>
	// 	return fmt.Sprintf("$:%s", theUserID)
	// }
	// // $:<user_id>:project:<project_id>
	// return fmt.Sprintf("$:%s:project:%s", theUserID, projectID)
	if projectID == nil {
		// TODO: do nothing now
		log.Printf("No project ID provided, skipping broadcast")
		return "dummy_channel"
	}
	return fmt.Sprintf("project:%s", projectID)
}

// Broadcast sends a message to all relevant connections
func (h *Hub) Broadcast(message *Message, projectID *uuid.UUID, userID *string, excludeConn *Connection) {
	h.metrics.incrementBroadcastsSent()

	channel := generatePrivateChannel(userID, projectID)

	messageBytes, err := message.ToBytes()
	if err != nil {
		log.Printf("Error converting message to bytes: %v", err)
		return
	}
	h.node.Publish(channel, messageBytes)
}

// BroadcastToProject sends a message to all connections subscribed to a project
func (h *Hub) BroadcastToProject(message *Message, projectID uuid.UUID, excludeConn *Connection) {
	h.Broadcast(message, &projectID, nil, excludeConn)
}

// BroadcastToUser sends a message to all connections of a specific user
func (h *Hub) BroadcastToUser(message *Message, userID string, excludeConn *Connection) {
	h.Broadcast(message, nil, &userID, excludeConn)
}

// BroadcastToAll sends a message to all connected clients
func (h *Hub) BroadcastToAll(message *Message, excludeConn *Connection) {
	h.Broadcast(message, nil, nil, excludeConn)
}

// GetMetrics returns hub metrics
func (h *Hub) GetMetrics() HubMetrics {
	h.metrics.mu.RLock()
	defer h.metrics.mu.RUnlock()
	return *h.metrics
}

// Shutdown gracefully shuts down the hub and closes all connections
func (h *Hub) Shutdown() {
	h.node.Shutdown(context.Background())
	log.Printf("Hub shutdown complete")
}

// Metrics methods
func (m *HubMetrics) incrementActiveConnections() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ActiveConnections++
}

func (m *HubMetrics) decrementActiveConnections() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ActiveConnections--
}

func (m *HubMetrics) incrementConnectionsCreated() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ConnectionsCreated++
	m.TotalConnections++
}

func (m *HubMetrics) incrementConnectionsClosed() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ConnectionsClosed++
}

func (m *HubMetrics) incrementMessagesSent() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.MessagesSent++
}

func (m *HubMetrics) incrementMessagesReceived() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.MessagesReceived++
}

func (m *HubMetrics) incrementBroadcastsSent() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.BroadcastsSent++
}
