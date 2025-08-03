package websocket

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512

	// Buffer size for the message channel
	messageBufferSize = 256
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// TODO: Implement proper origin checking for production
		return true
	},
}

// Connection represents a WebSocket connection
type Connection struct {
	// The WebSocket connection
	conn *websocket.Conn

	// Connection ID
	ID string

	// User ID (from authentication)
	UserID string

	// Project IDs this connection is subscribed to
	ProjectIDs map[uuid.UUID]bool

	// Send channel for outbound messages
	send chan []byte

	// Hub reference
	hub *Hub

	// Connection metadata
	ConnectedAt time.Time
	LastPong    time.Time

	// Mutex for thread-safe operations
	mu sync.RWMutex

	// Context for graceful shutdown
	ctx    context.Context
	cancel context.CancelFunc

	// Flag to prevent multiple closes
	closed  bool
	closeMu sync.Mutex
}

// NewConnection creates a new WebSocket connection
func NewConnection(conn *websocket.Conn, hub *Hub) *Connection {
	ctx, cancel := context.WithCancel(context.Background())

	return &Connection{
		conn:        conn,
		ID:          uuid.New().String(),
		ProjectIDs:  make(map[uuid.UUID]bool),
		send:        make(chan []byte, messageBufferSize),
		hub:         hub,
		ConnectedAt: time.Now(),
		LastPong:    time.Now(),
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Start begins the connection's read and write pumps
func (c *Connection) Start() {
	go c.writePump()
	go c.readPump()
}

// Close closes the connection and cleans up resources
func (c *Connection) Close() {
	c.closeMu.Lock()
	defer c.closeMu.Unlock()

	// Check if already closed
	if c.closed {
		return
	}

	c.closed = true
	c.cancel()
	close(c.send)

	// Only close conn if it's not nil
	if c.conn != nil {
		c.conn.Close()
	}
}

// SendMessage sends a message to the connection
func (c *Connection) SendMessage(message *Message) error {
	// Check if connection is closed
	if c.IsClosed() {
		return ErrConnectionClosed
	}

	data, err := message.ToBytes()
	if err != nil {
		return err
	}

	select {
	case c.send <- data:
		return nil
	default:
		// Channel is full, connection is slow or dead
		return ErrConnectionClosed
	}
}

// SendBytes sends raw bytes to the connection
func (c *Connection) SendBytes(data []byte) error {
	// Check if connection is closed
	if c.IsClosed() {
		return ErrConnectionClosed
	}

	select {
	case c.send <- data:
		return nil
	default:
		return ErrConnectionClosed
	}
}

// SubscribeToProject subscribes the connection to a project
func (c *Connection) SubscribeToProject(projectID uuid.UUID) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ProjectIDs[projectID] = true
}

// UnsubscribeFromProject unsubscribes the connection from a project
func (c *Connection) UnsubscribeFromProject(projectID uuid.UUID) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.ProjectIDs, projectID)
}

// IsSubscribedToProject checks if the connection is subscribed to a project
func (c *Connection) IsSubscribedToProject(projectID uuid.UUID) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ProjectIDs[projectID]
}

// GetSubscribedProjects returns a copy of subscribed project IDs
func (c *Connection) GetSubscribedProjects() []uuid.UUID {
	c.mu.RLock()
	defer c.mu.RUnlock()

	projects := make([]uuid.UUID, 0, len(c.ProjectIDs))
	for projectID := range c.ProjectIDs {
		projects = append(projects, projectID)
	}
	return projects
}

// SetUserID sets the user ID for the connection
func (c *Connection) SetUserID(userID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.UserID = userID
}

// GetUserID returns the user ID for the connection
func (c *Connection) GetUserID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.UserID
}

// IsClosed returns true if the connection is closed
func (c *Connection) IsClosed() bool {
	c.closeMu.Lock()
	defer c.closeMu.Unlock()
	return c.closed
}

// SafeClose safely closes the connection if it's not already closed
func (c *Connection) SafeClose() {
	c.closeMu.Lock()
	defer c.closeMu.Unlock()

	if !c.closed {
		c.closed = true
		c.cancel()
		close(c.send)

		// Only close conn if it's not nil
		if c.conn != nil {
			c.conn.Close()
		}
	}
}

// readPump pumps messages from the WebSocket connection to the hub
func (c *Connection) readPump() {
	defer func() {
		c.hub.Unregister(c)
		c.SafeClose()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.LastPong = time.Now()
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			_, messageBytes, err := c.conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket error: %v", err)
				}
				return
			}

			// Parse and handle the message
			c.handleIncomingMessage(messageBytes)
		}
	}
}

// writePump pumps messages from the hub to the WebSocket connection
func (c *Connection) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.SafeClose()
	}()

	for {
		select {
		case <-c.ctx.Done():
			return
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current WebSocket message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleIncomingMessage processes incoming messages from the client
func (c *Connection) handleIncomingMessage(messageBytes []byte) {
	message, err := FromBytes(messageBytes)
	if err != nil {
		log.Printf("Error parsing WebSocket message: %v", err)
		c.sendError("invalid_message", "Failed to parse message")
		return
	}

	switch message.Type {
	case Ping:
		c.handlePing()
	case AuthRequired:
		c.handleAuth(message)
	default:
		// Forward to hub for processing
		c.hub.ProcessMessage(c, message)
	}
}

// handlePing responds to ping messages with pong
func (c *Connection) handlePing() {
	pongMessage, _ := NewMessage(Pong, map[string]string{"status": "ok"})
	c.SendMessage(pongMessage)
}

// handleAuth processes authentication messages
func (c *Connection) handleAuth(message *Message) {
	var authData AuthData
	if err := message.ParseData(&authData); err != nil {
		c.sendError("invalid_auth", "Invalid authentication data")
		return
	}

	// TODO: Implement proper token validation
	// For now, we'll accept any non-empty token
	if authData.Token != "" {
		c.SetUserID(authData.UserID)

		successMessage, _ := NewMessage(AuthSuccess, AuthData{
			UserID:  authData.UserID,
			Message: "Authentication successful",
		})
		c.SendMessage(successMessage)
	} else {
		failMessage, _ := NewMessage(AuthFailed, AuthData{
			Message: "Invalid token",
		})
		c.SendMessage(failMessage)
	}
}

// sendError sends an error message to the connection
func (c *Connection) sendError(code, message string) {
	errorMessage, _ := NewMessage(Error, ErrorData{
		Code:    code,
		Message: message,
	})
	c.SendMessage(errorMessage)
}

// IsHealthy checks if the connection is healthy
func (c *Connection) IsHealthy() bool {
	return time.Since(c.LastPong) < pongWait*2
}

// GetConnectionInfo returns connection metadata
func (c *Connection) GetConnectionInfo() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return map[string]interface{}{
		"id":            c.ID,
		"user_id":       c.UserID,
		"connected_at":  c.ConnectedAt,
		"last_pong":     c.LastPong,
		"is_healthy":    c.IsHealthy(),
		"is_closed":     c.IsClosed(),
		"project_count": len(c.ProjectIDs),
	}
}
