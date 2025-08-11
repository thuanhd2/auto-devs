package websocket

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// MessageType represents the type of WebSocket message
type MessageType string

const (
	// Task related messages
	TaskCreated MessageType = "task_created"
	TaskUpdated MessageType = "task_updated"
	TaskDeleted MessageType = "task_deleted"

	// Project related messages
	ProjectUpdated MessageType = "project_updated"

	// Status related messages
	StatusChanged MessageType = "status_changed"

	// User presence messages
	UserJoined MessageType = "user_joined"
	UserLeft   MessageType = "user_left"

	// Connection management messages
	Ping MessageType = "ping"
	Pong MessageType = "pong"

	// Subscription messages
	Subscription MessageType = "subscription"

	// Error messages
	Error MessageType = "error"

	// Pull Request related messages
	MessageTypePRUpdate MessageType = "pr_update"

	// Authentication messages
	AuthRequired MessageType = "auth_required"
	AuthSuccess  MessageType = "auth_success"
	AuthFailed   MessageType = "auth_failed"

	// Execution logs updated
	ExecutionLogsCreated MessageType = "execution_logs_created"
)

// Message represents a WebSocket message
type Message struct {
	Type      MessageType     `json:"type"`
	Data      json.RawMessage `json:"data"`
	Timestamp time.Time       `json:"timestamp"`
	MessageID string          `json:"message_id"`
}

// TaskData represents task-related message data
type TaskData struct {
	TaskID    uuid.UUID              `json:"task_id"`
	ProjectID uuid.UUID              `json:"project_id"`
	Changes   map[string]interface{} `json:"changes,omitempty"`
	Task      interface{}            `json:"task,omitempty"`
}

// ProjectData represents project-related message data
type ProjectData struct {
	ProjectID uuid.UUID              `json:"project_id"`
	Changes   map[string]interface{} `json:"changes,omitempty"`
	Project   interface{}            `json:"project,omitempty"`
}

// StatusData represents status change message data
type StatusData struct {
	EntityID   uuid.UUID `json:"entity_id"`
	EntityType string    `json:"entity_type"` // "task" or "project"
	OldStatus  string    `json:"old_status"`
	NewStatus  string    `json:"new_status"`
	ProjectID  uuid.UUID `json:"project_id"`
}

// UserPresenceData represents user presence message data
type UserPresenceData struct {
	UserID    string    `json:"user_id"`
	ProjectID uuid.UUID `json:"project_id"`
	Action    string    `json:"action"` // "joined" or "left"
}

// ErrorData represents error message data
type ErrorData struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// AuthData represents authentication message data
type AuthData struct {
	Token   string `json:"token,omitempty"`
	UserID  string `json:"user_id,omitempty"`
	Message string `json:"message,omitempty"`
}

// NewMessage creates a new WebSocket message
func NewMessage(msgType MessageType, data interface{}) (*Message, error) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return &Message{
		Type:      msgType,
		Data:      dataBytes,
		Timestamp: time.Now(),
		MessageID: uuid.New().String(),
	}, nil
}

// ParseData parses the message data into the provided struct
func (m *Message) ParseData(v interface{}) error {
	return json.Unmarshal(m.Data, v)
}

// ToBytes converts the message to JSON bytes
func (m *Message) ToBytes() ([]byte, error) {
	return json.Marshal(m)
}

// FromBytes creates a message from JSON bytes
func FromBytes(data []byte) (*Message, error) {
	var msg Message
	err := json.Unmarshal(data, &msg)
	return &msg, err
}
