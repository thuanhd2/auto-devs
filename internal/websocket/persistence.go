package websocket

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// MessageStatus represents the status of a persisted message
type MessageStatus string

const (
	MessageStatusPending   MessageStatus = "pending"
	MessageStatusDelivered MessageStatus = "delivered"
	MessageStatusExpired   MessageStatus = "expired"
	MessageStatusFailed    MessageStatus = "failed"
)

// PersistedMessage represents a message stored for offline delivery
type PersistedMessage struct {
	ID         string        `json:"id"`
	UserID     string        `json:"user_id"`
	ProjectID  *uuid.UUID    `json:"project_id,omitempty"`
	Message    *Message      `json:"message"`
	Status     MessageStatus `json:"status"`
	CreatedAt  time.Time     `json:"created_at"`
	ExpiresAt  time.Time     `json:"expires_at"`
	Attempts   int           `json:"attempts"`
	MaxAttempts int          `json:"max_attempts"`
	LastAttempt *time.Time   `json:"last_attempt,omitempty"`
	Error      string        `json:"error,omitempty"`
}

// MessagePersistence defines the interface for message persistence
type MessagePersistence interface {
	Store(userID string, projectID *uuid.UUID, message *Message, ttl time.Duration) error
	GetPendingMessages(userID string) ([]*PersistedMessage, error)
	MarkAsDelivered(messageID string) error
	MarkAsFailed(messageID string, errorMsg string) error
	CleanupExpired() error
	GetStats() map[string]interface{}
}

// InMemoryPersistence provides an in-memory implementation of message persistence
type InMemoryPersistence struct {
	messages   map[string]*PersistedMessage
	userIndex  map[string][]string // userID -> []messageID
	mu         sync.RWMutex
	maxMessages int
	defaultTTL time.Duration
}

// NewInMemoryPersistence creates a new in-memory persistence store
func NewInMemoryPersistence(maxMessages int, defaultTTL time.Duration) *InMemoryPersistence {
	imp := &InMemoryPersistence{
		messages:    make(map[string]*PersistedMessage),
		userIndex:   make(map[string][]string),
		maxMessages: maxMessages,
		defaultTTL:  defaultTTL,
	}
	
	// Start cleanup goroutine
	go imp.cleanupRoutine()
	
	return imp
}

// Store stores a message for offline delivery
func (imp *InMemoryPersistence) Store(userID string, projectID *uuid.UUID, message *Message, ttl time.Duration) error {
	imp.mu.Lock()
	defer imp.mu.Unlock()
	
	// Check if we're at capacity
	if len(imp.messages) >= imp.maxMessages {
		// Remove oldest message
		imp.removeOldestMessage()
	}
	
	if ttl == 0 {
		ttl = imp.defaultTTL
	}
	
	persistedMsg := &PersistedMessage{
		ID:          uuid.New().String(),
		UserID:      userID,
		ProjectID:   projectID,
		Message:     message,
		Status:      MessageStatusPending,
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(ttl),
		Attempts:    0,
		MaxAttempts: 3,
	}
	
	// Store message
	imp.messages[persistedMsg.ID] = persistedMsg
	
	// Add to user index
	if imp.userIndex[userID] == nil {
		imp.userIndex[userID] = make([]string, 0)
	}
	imp.userIndex[userID] = append(imp.userIndex[userID], persistedMsg.ID)
	
	return nil
}

// GetPendingMessages retrieves pending messages for a user
func (imp *InMemoryPersistence) GetPendingMessages(userID string) ([]*PersistedMessage, error) {
	imp.mu.RLock()
	defer imp.mu.RUnlock()
	
	messageIDs, exists := imp.userIndex[userID]
	if !exists {
		return []*PersistedMessage{}, nil
	}
	
	var pendingMessages []*PersistedMessage
	for _, messageID := range messageIDs {
		if msg, exists := imp.messages[messageID]; exists {
			if msg.Status == MessageStatusPending && time.Now().Before(msg.ExpiresAt) {
				pendingMessages = append(pendingMessages, msg)
			}
		}
	}
	
	return pendingMessages, nil
}

// MarkAsDelivered marks a message as delivered
func (imp *InMemoryPersistence) MarkAsDelivered(messageID string) error {
	imp.mu.Lock()
	defer imp.mu.Unlock()
	
	if msg, exists := imp.messages[messageID]; exists {
		msg.Status = MessageStatusDelivered
		return nil
	}
	
	return fmt.Errorf("message not found: %s", messageID)
}

// MarkAsFailed marks a message as failed
func (imp *InMemoryPersistence) MarkAsFailed(messageID string, errorMsg string) error {
	imp.mu.Lock()
	defer imp.mu.Unlock()
	
	if msg, exists := imp.messages[messageID]; exists {
		msg.Attempts++
		msg.LastAttempt = &[]time.Time{time.Now()}[0]
		msg.Error = errorMsg
		
		if msg.Attempts >= msg.MaxAttempts {
			msg.Status = MessageStatusFailed
		}
		
		return nil
	}
	
	return fmt.Errorf("message not found: %s", messageID)
}

// CleanupExpired removes expired messages
func (imp *InMemoryPersistence) CleanupExpired() error {
	imp.mu.Lock()
	defer imp.mu.Unlock()
	
	now := time.Now()
	var expiredIDs []string
	
	// Find expired messages
	for id, msg := range imp.messages {
		if now.After(msg.ExpiresAt) || msg.Status == MessageStatusDelivered {
			expiredIDs = append(expiredIDs, id)
		}
	}
	
	// Remove expired messages
	for _, id := range expiredIDs {
		imp.removeMessage(id)
	}
	
	return nil
}

// GetStats returns persistence statistics
func (imp *InMemoryPersistence) GetStats() map[string]interface{} {
	imp.mu.RLock()
	defer imp.mu.RUnlock()
	
	stats := map[string]interface{}{
		"total_messages": len(imp.messages),
		"max_messages":   imp.maxMessages,
		"default_ttl":    imp.defaultTTL.String(),
		"users_with_messages": len(imp.userIndex),
	}
	
	// Count by status
	statusCounts := make(map[MessageStatus]int)
	for _, msg := range imp.messages {
		statusCounts[msg.Status]++
	}
	stats["status_counts"] = statusCounts
	
	return stats
}

// removeOldestMessage removes the oldest message to make space
func (imp *InMemoryPersistence) removeOldestMessage() {
	var oldestID string
	var oldestTime time.Time
	
	for id, msg := range imp.messages {
		if oldestID == "" || msg.CreatedAt.Before(oldestTime) {
			oldestID = id
			oldestTime = msg.CreatedAt
		}
	}
	
	if oldestID != "" {
		imp.removeMessage(oldestID)
	}
}

// removeMessage removes a message and updates indices
func (imp *InMemoryPersistence) removeMessage(messageID string) {
	if msg, exists := imp.messages[messageID]; exists {
		// Remove from user index
		userMessages := imp.userIndex[msg.UserID]
		for i, id := range userMessages {
			if id == messageID {
				imp.userIndex[msg.UserID] = append(userMessages[:i], userMessages[i+1:]...)
				break
			}
		}
		
		// Remove from messages
		delete(imp.messages, messageID)
	}
}

// cleanupRoutine runs periodic cleanup
func (imp *InMemoryPersistence) cleanupRoutine() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		imp.CleanupExpired()
	}
}

// OfflineMessageManager manages offline message delivery
type OfflineMessageManager struct {
	persistence MessagePersistence
	hub         *Hub
	mu          sync.RWMutex
}

// NewOfflineMessageManager creates a new offline message manager
func NewOfflineMessageManager(persistence MessagePersistence, hub *Hub) *OfflineMessageManager {
	return &OfflineMessageManager{
		persistence: persistence,
		hub:         hub,
	}
}

// StoreMessage stores a message for offline delivery
func (omm *OfflineMessageManager) StoreMessage(userID string, projectID *uuid.UUID, message *Message, ttl time.Duration) error {
	return omm.persistence.Store(userID, projectID, message, ttl)
}

// DeliverPendingMessages delivers pending messages to a newly connected user
func (omm *OfflineMessageManager) DeliverPendingMessages(userID string, conn *Connection) error {
	pendingMessages, err := omm.persistence.GetPendingMessages(userID)
	if err != nil {
		return err
	}
	
	for _, persistedMsg := range pendingMessages {
		// Check if user is still subscribed to the project
		if persistedMsg.ProjectID != nil && !conn.IsSubscribedToProject(*persistedMsg.ProjectID) {
			continue
		}
		
		// Try to deliver the message
		if err := conn.SendMessage(persistedMsg.Message); err != nil {
			// Mark as failed
			omm.persistence.MarkAsFailed(persistedMsg.ID, err.Error())
		} else {
			// Mark as delivered
			omm.persistence.MarkAsDelivered(persistedMsg.ID)
		}
	}
	
	return nil
}

// BroadcastWithPersistence broadcasts a message and stores it for offline users
func (omm *OfflineMessageManager) BroadcastWithPersistence(message *Message, projectID *uuid.UUID, userID *string, excludeConn *Connection, ttl time.Duration) {
	// Regular broadcast
	omm.hub.Broadcast(message, projectID, userID, excludeConn)
	
	// Store for offline users
	if projectID != nil {
		// Get all users who should receive this message but are not connected
		// This would require integration with a user service to know who is subscribed to the project
		// For now, we'll skip offline storage for project broadcasts
	} else if userID != nil {
		// Check if user is connected
		if omm.hub.GetUserConnectionCount(*userID) == 0 {
			// User is offline, store the message
			omm.persistence.Store(*userID, projectID, message, ttl)
		}
	}
}

// GetStats returns offline message manager statistics
func (omm *OfflineMessageManager) GetStats() map[string]interface{} {
	return omm.persistence.GetStats()
}

// MessageQueue represents a queue of messages for delivery
type MessageQueue struct {
	messages chan *Message
	stop     chan struct{}
	conn     *Connection
	omm      *OfflineMessageManager
}

// NewMessageQueue creates a new message queue for a connection
func NewMessageQueue(conn *Connection, omm *OfflineMessageManager) *MessageQueue {
	mq := &MessageQueue{
		messages: make(chan *Message, 100),
		stop:     make(chan struct{}),
		conn:     conn,
		omm:      omm,
	}
	
	go mq.processingLoop()
	return mq
}

// Enqueue adds a message to the queue
func (mq *MessageQueue) Enqueue(message *Message) {
	select {
	case mq.messages <- message:
	default:
		// Queue is full, store for offline delivery if user is authenticated
		if userID := mq.conn.GetUserID(); userID != "" {
			mq.omm.StoreMessage(userID, nil, message, 24*time.Hour)
		}
	}
}

// Stop stops the message queue
func (mq *MessageQueue) Stop() {
	close(mq.stop)
}

// processingLoop processes messages from the queue
func (mq *MessageQueue) processingLoop() {
	for {
		select {
		case message := <-mq.messages:
			if err := mq.conn.SendMessage(message); err != nil {
				// Store for offline delivery if user is authenticated
				if userID := mq.conn.GetUserID(); userID != "" {
					mq.omm.StoreMessage(userID, nil, message, 24*time.Hour)
				}
			}
		case <-mq.stop:
			return
		}
	}
}