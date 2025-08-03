package websocket

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestConnection_Close(t *testing.T) {
	// Create a hub
	hub := NewHub()

	// Create a new connection with nil conn (for testing close logic only)
	wsConn := &Connection{
		conn:        nil,
		ID:          "test-connection",
		ProjectIDs:  make(map[uuid.UUID]bool),
		send:        make(chan []byte, messageBufferSize),
		hub:         hub,
		ConnectedAt: time.Now(),
		LastPong:    time.Now(),
		ctx:         context.Background(),
		cancel:      func() {},
		closed:      false,
		closeMu:     sync.Mutex{},
	}

	// Test that connection is not closed initially
	assert.False(t, wsConn.IsClosed())

	// Close the connection
	wsConn.Close()

	// Test that connection is now closed
	assert.True(t, wsConn.IsClosed())

	// Test that calling Close() again doesn't panic
	assert.NotPanics(t, func() {
		wsConn.Close()
	})

	// Test that connection is still closed
	assert.True(t, wsConn.IsClosed())
}

func TestConnection_SafeClose(t *testing.T) {
	// Create a hub
	hub := NewHub()

	// Create a new connection with nil conn (for testing close logic only)
	wsConn := &Connection{
		conn:        nil,
		ID:          "test-connection",
		ProjectIDs:  make(map[uuid.UUID]bool),
		send:        make(chan []byte, messageBufferSize),
		hub:         hub,
		ConnectedAt: time.Now(),
		LastPong:    time.Now(),
		ctx:         context.Background(),
		cancel:      func() {},
		closed:      false,
		closeMu:     sync.Mutex{},
	}

	// Test that connection is not closed initially
	assert.False(t, wsConn.IsClosed())

	// Safely close the connection
	wsConn.SafeClose()

	// Test that connection is now closed
	assert.True(t, wsConn.IsClosed())

	// Test that calling SafeClose() again doesn't panic
	assert.NotPanics(t, func() {
		wsConn.SafeClose()
	})

	// Test that connection is still closed
	assert.True(t, wsConn.IsClosed())
}

func TestConnection_SendMessage_WhenClosed(t *testing.T) {
	// Create a hub
	hub := NewHub()

	// Create a new connection with nil conn (for testing close logic only)
	wsConn := &Connection{
		conn:        nil,
		ID:          "test-connection",
		ProjectIDs:  make(map[uuid.UUID]bool),
		send:        make(chan []byte, messageBufferSize),
		hub:         hub,
		ConnectedAt: time.Now(),
		LastPong:    time.Now(),
		ctx:         context.Background(),
		cancel:      func() {},
		closed:      false,
		closeMu:     sync.Mutex{},
	}

	// Close the connection
	wsConn.Close()

	// Test that sending a message to a closed connection returns an error
	message, _ := NewMessage(Ping, map[string]string{"status": "ok"})
	err := wsConn.SendMessage(message)
	assert.Equal(t, ErrConnectionClosed, err)
}

func TestConnection_SendBytes_WhenClosed(t *testing.T) {
	// Create a hub
	hub := NewHub()

	// Create a new connection with nil conn (for testing close logic only)
	wsConn := &Connection{
		conn:        nil,
		ID:          "test-connection",
		ProjectIDs:  make(map[uuid.UUID]bool),
		send:        make(chan []byte, messageBufferSize),
		hub:         hub,
		ConnectedAt: time.Now(),
		LastPong:    time.Now(),
		ctx:         context.Background(),
		cancel:      func() {},
		closed:      false,
		closeMu:     sync.Mutex{},
	}

	// Close the connection
	wsConn.Close()

	// Test that sending bytes to a closed connection returns an error
	err := wsConn.SendBytes([]byte("test"))
	assert.Equal(t, ErrConnectionClosed, err)
}

func TestConnection_ConcurrentClose(t *testing.T) {
	// Create a hub
	hub := NewHub()

	// Create a new connection with nil conn (for testing close logic only)
	wsConn := &Connection{
		conn:        nil,
		ID:          "test-connection",
		ProjectIDs:  make(map[uuid.UUID]bool),
		send:        make(chan []byte, messageBufferSize),
		hub:         hub,
		ConnectedAt: time.Now(),
		LastPong:    time.Now(),
		ctx:         context.Background(),
		cancel:      func() {},
		closed:      false,
		closeMu:     sync.Mutex{},
	}

	// Test concurrent close operations
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Panic occurred: %v", r)
				}
				done <- true
			}()

			wsConn.Close()
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Test that connection is closed
	assert.True(t, wsConn.IsClosed())
}

func TestConnection_GetConnectionInfo(t *testing.T) {
	// Create a hub
	hub := NewHub()

	// Create a new connection with nil conn (for testing close logic only)
	wsConn := &Connection{
		conn:        nil,
		ID:          "test-connection",
		ProjectIDs:  make(map[uuid.UUID]bool),
		send:        make(chan []byte, messageBufferSize),
		hub:         hub,
		ConnectedAt: time.Now(),
		LastPong:    time.Now(),
		ctx:         context.Background(),
		cancel:      func() {},
		closed:      false,
		closeMu:     sync.Mutex{},
	}

	// Get connection info
	info := wsConn.GetConnectionInfo()

	// Test that info contains expected fields
	assert.NotEmpty(t, info["id"])
	assert.Equal(t, "", info["user_id"])
	assert.False(t, info["is_closed"].(bool))
	assert.Equal(t, 0, info["project_count"].(int))

	// Close the connection
	wsConn.Close()

	// Get connection info again
	info = wsConn.GetConnectionInfo()

	// Test that is_closed is now true
	assert.True(t, info["is_closed"].(bool))
}
