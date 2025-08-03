package websocket

import (
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Hub maintains the set of active connections and broadcasts messages to them
type Hub struct {
	// Registered connections
	connections map[*Connection]bool

	// Connections grouped by project ID for efficient broadcasting
	projectConnections map[uuid.UUID]map[*Connection]bool

	// Connections grouped by user ID
	userConnections map[string]map[*Connection]bool

	// Inbound messages from the connections
	broadcast chan *BroadcastMessage

	// Register requests from the connections
	register chan *Connection

	// Unregister requests from connections
	unregister chan *Connection

	// Message processors
	processors map[MessageType]MessageProcessor

	// Metrics
	metrics *HubMetrics

	// Mutex for thread-safe operations
	mu sync.RWMutex

	// Cleanup ticker
	cleanupTicker *time.Ticker
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
func NewHub() *Hub {
	hub := &Hub{
		connections:        make(map[*Connection]bool),
		projectConnections: make(map[uuid.UUID]map[*Connection]bool),
		userConnections:    make(map[string]map[*Connection]bool),
		broadcast:          make(chan *BroadcastMessage, 256),
		register:           make(chan *Connection),
		unregister:         make(chan *Connection),
		processors:         make(map[MessageType]MessageProcessor),
		metrics:            &HubMetrics{},
		cleanupTicker:      time.NewTicker(30 * time.Second),
	}

	return hub
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	defer h.cleanupTicker.Stop()

	for {
		select {
		case conn := <-h.register:
			h.registerConnection(conn)

		case conn := <-h.unregister:
			h.unregisterConnection(conn)

		case broadcastMsg := <-h.broadcast:
			h.broadcastMessage(broadcastMsg)

		case <-h.cleanupTicker.C:
			h.cleanupUnhealthyConnections()
		}
	}
}

// Register registers a new connection with the hub
func (h *Hub) Register(conn *Connection) {
	h.register <- conn
}

// Unregister unregisters a connection from the hub
func (h *Hub) Unregister(conn *Connection) {
	h.unregister <- conn
}

// Broadcast sends a message to all relevant connections
func (h *Hub) Broadcast(message *Message, projectID *uuid.UUID, userID *string, excludeConn *Connection) {
	broadcastMsg := &BroadcastMessage{
		Message:     message,
		ProjectID:   projectID,
		UserID:      userID,
		ExcludeConn: excludeConn,
	}

	select {
	case h.broadcast <- broadcastMsg:
		h.metrics.incrementBroadcastsSent()
	default:
		log.Printf("Warning: Broadcast channel full, dropping message")
	}
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

// ProcessMessage processes an incoming message from a connection
func (h *Hub) ProcessMessage(conn *Connection, message *Message) {
	h.metrics.incrementMessagesReceived()

	processor, exists := h.processors[message.Type]
	if exists {
		if err := processor.ProcessMessage(conn, message); err != nil {
			log.Printf("Error processing message type %s: %v", message.Type, err)
			conn.sendError("processing_error", "Failed to process message")
		}
	} else {
		log.Printf("No processor found for message type: %s", message.Type)
		conn.sendError("unsupported_message", "Message type not supported")
	}
}

// RegisterProcessor registers a message processor for a specific message type
func (h *Hub) RegisterProcessor(msgType MessageType, processor MessageProcessor) {
	h.processors[msgType] = processor
}

// registerConnection registers a new connection
func (h *Hub) registerConnection(conn *Connection) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.connections[conn] = true
	h.metrics.incrementConnectionsCreated()
	h.metrics.incrementActiveConnections()

	log.Printf("Connection registered: %s", conn.ID)
}

// unregisterConnection unregisters a connection
func (h *Hub) unregisterConnection(conn *Connection) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Check if connection is already closed
	if conn.IsClosed() {
		return
	}

	if _, ok := h.connections[conn]; ok {
		delete(h.connections, conn)

		// Remove from project connections
		for projectID := range conn.ProjectIDs {
			if projectConns, exists := h.projectConnections[projectID]; exists {
				delete(projectConns, conn)
				if len(projectConns) == 0 {
					delete(h.projectConnections, projectID)
				}
			}
		}

		// Remove from user connections
		userID := conn.GetUserID()
		if userID != "" {
			if userConns, exists := h.userConnections[userID]; exists {
				delete(userConns, conn)
				if len(userConns) == 0 {
					delete(h.userConnections, userID)
				}
			}
		}

		h.metrics.decrementActiveConnections()
		h.metrics.incrementConnectionsClosed()

		log.Printf("Connection unregistered: %s", conn.ID)
	}
}

// broadcastMessage broadcasts a message to the appropriate connections
func (h *Hub) broadcastMessage(broadcastMsg *BroadcastMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var targetConnections []*Connection

	switch {
	case broadcastMsg.ProjectID != nil:
		// Broadcast to project subscribers
		if projectConns, exists := h.projectConnections[*broadcastMsg.ProjectID]; exists {
			for conn := range projectConns {
				if conn != broadcastMsg.ExcludeConn {
					targetConnections = append(targetConnections, conn)
				}
			}
		}

	case broadcastMsg.UserID != nil:
		// Broadcast to specific user
		if userConns, exists := h.userConnections[*broadcastMsg.UserID]; exists {
			for conn := range userConns {
				if conn != broadcastMsg.ExcludeConn {
					targetConnections = append(targetConnections, conn)
				}
			}
		}

	default:
		// Broadcast to all connections
		for conn := range h.connections {
			if conn != broadcastMsg.ExcludeConn {
				targetConnections = append(targetConnections, conn)
			}
		}
	}

	// Send message to target connections
	for _, conn := range targetConnections {
		// Check if connection is closed before sending
		if conn.IsClosed() {
			continue
		}

		if err := conn.SendMessage(broadcastMsg.Message); err != nil {
			log.Printf("Error sending message to connection %s: %v", conn.ID, err)
			// Don't unregister here as it could cause deadlock
			go h.Unregister(conn)
		} else {
			h.metrics.incrementMessagesSent()
		}
	}
}

// SubscribeConnectionToProject subscribes a connection to a project
func (h *Hub) SubscribeConnectionToProject(conn *Connection, projectID uuid.UUID) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Add to connection's project list
	conn.SubscribeToProject(projectID)

	// Add to hub's project connections map
	if h.projectConnections[projectID] == nil {
		h.projectConnections[projectID] = make(map[*Connection]bool)
	}
	h.projectConnections[projectID][conn] = true

	log.Printf("Connection %s subscribed to project %s", conn.ID, projectID)
}

// UnsubscribeConnectionFromProject unsubscribes a connection from a project
func (h *Hub) UnsubscribeConnectionFromProject(conn *Connection, projectID uuid.UUID) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Remove from connection's project list
	conn.UnsubscribeFromProject(projectID)

	// Remove from hub's project connections map
	if projectConns, exists := h.projectConnections[projectID]; exists {
		delete(projectConns, conn)
		if len(projectConns) == 0 {
			delete(h.projectConnections, projectID)
		}
	}

	log.Printf("Connection %s unsubscribed from project %s", conn.ID, projectID)
}

// AssociateConnectionWithUser associates a connection with a user
func (h *Hub) AssociateConnectionWithUser(conn *Connection, userID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Remove from old user association if exists
	oldUserID := conn.GetUserID()
	if oldUserID != "" && oldUserID != userID {
		if userConns, exists := h.userConnections[oldUserID]; exists {
			delete(userConns, conn)
			if len(userConns) == 0 {
				delete(h.userConnections, oldUserID)
			}
		}
	}

	// Set new user ID
	conn.SetUserID(userID)

	// Add to new user association
	if h.userConnections[userID] == nil {
		h.userConnections[userID] = make(map[*Connection]bool)
	}
	h.userConnections[userID][conn] = true

	log.Printf("Connection %s associated with user %s", conn.ID, userID)
}

// cleanupUnhealthyConnections removes unhealthy connections
func (h *Hub) cleanupUnhealthyConnections() {
	h.mu.RLock()
	var unhealthyConnections []*Connection
	for conn := range h.connections {
		if !conn.IsHealthy() {
			unhealthyConnections = append(unhealthyConnections, conn)
		}
	}
	h.mu.RUnlock()

	for _, conn := range unhealthyConnections {
		log.Printf("Cleaning up unhealthy connection: %s", conn.ID)
		h.Unregister(conn)
	}
}

// GetMetrics returns hub metrics
func (h *Hub) GetMetrics() HubMetrics {
	h.metrics.mu.RLock()
	defer h.metrics.mu.RUnlock()
	return *h.metrics
}

// GetConnectionsInfo returns information about all connections
func (h *Hub) GetConnectionsInfo() []map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var info []map[string]interface{}
	for conn := range h.connections {
		info = append(info, conn.GetConnectionInfo())
	}
	return info
}

// GetProjectConnectionCount returns the number of connections for a project
func (h *Hub) GetProjectConnectionCount(projectID uuid.UUID) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if projectConns, exists := h.projectConnections[projectID]; exists {
		return len(projectConns)
	}
	return 0
}

// GetUserConnectionCount returns the number of connections for a user
func (h *Hub) GetUserConnectionCount(userID string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if userConns, exists := h.userConnections[userID]; exists {
		return len(userConns)
	}
	return 0
}

// Shutdown gracefully shuts down the hub and closes all connections
func (h *Hub) Shutdown() {
	h.mu.Lock()
	defer h.mu.Unlock()

	log.Printf("Shutting down hub, closing %d connections", len(h.connections))

	// Close all connections safely
	for conn := range h.connections {
		conn.SafeClose()
	}

	// Clear all maps
	h.connections = make(map[*Connection]bool)
	h.projectConnections = make(map[uuid.UUID]map[*Connection]bool)
	h.userConnections = make(map[string]map[*Connection]bool)

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
