package websocket

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test helper functions
func setupTestServer() (*httptest.Server, *Handler) {
	handler := NewHandler()
	
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		upgrader.Upgrade(w, r, nil)
	})
	
	server := httptest.NewServer(mux)
	return server, handler
}

func connectWebSocket(t *testing.T, url string) *websocket.Conn {
	u := "ws" + strings.TrimPrefix(url, "http") + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(u, nil)
	require.NoError(t, err)
	return conn
}

func TestMessageCreation(t *testing.T) {
	t.Run("Create message with valid data", func(t *testing.T) {
		data := TaskData{
			TaskID:    uuid.New(),
			ProjectID: uuid.New(),
			Changes:   map[string]interface{}{"status": "completed"},
		}
		
		message, err := NewMessage(TaskUpdated, data)
		assert.NoError(t, err)
		assert.Equal(t, TaskUpdated, message.Type)
		assert.NotEmpty(t, message.MessageID)
		assert.False(t, message.Timestamp.IsZero())
	})
	
	t.Run("Parse message data", func(t *testing.T) {
		originalData := TaskData{
			TaskID:    uuid.New(),
			ProjectID: uuid.New(),
			Changes:   map[string]interface{}{"status": "completed"},
		}
		
		message, err := NewMessage(TaskUpdated, originalData)
		require.NoError(t, err)
		
		var parsedData TaskData
		err = message.ParseData(&parsedData)
		assert.NoError(t, err)
		assert.Equal(t, originalData.TaskID, parsedData.TaskID)
		assert.Equal(t, originalData.ProjectID, parsedData.ProjectID)
	})
	
	t.Run("Convert message to bytes and back", func(t *testing.T) {
		data := TaskData{TaskID: uuid.New(), ProjectID: uuid.New()}
		message, err := NewMessage(TaskCreated, data)
		require.NoError(t, err)
		
		bytes, err := message.ToBytes()
		assert.NoError(t, err)
		
		parsedMessage, err := FromBytes(bytes)
		assert.NoError(t, err)
		assert.Equal(t, message.Type, parsedMessage.Type)
		assert.Equal(t, message.MessageID, parsedMessage.MessageID)
	})
}

func TestHub(t *testing.T) {
	t.Run("Register and unregister connections", func(t *testing.T) {
		hub := NewHub()
		go hub.Run()
		defer func() {
			// Clean shutdown would be better, but for testing this is sufficient
		}()
		
		// Create mock connection
		conn := &Connection{
			ID:         uuid.New().String(),
			ProjectIDs: make(map[uuid.UUID]bool),
			send:       make(chan []byte, 10),
		}
		
		// Register connection
		hub.Register(conn)
		time.Sleep(10 * time.Millisecond) // Give time for registration
		
		// Check metrics
		metrics := hub.GetMetrics()
		assert.Equal(t, int64(1), metrics.ActiveConnections)
		
		// Unregister connection
		hub.Unregister(conn)
		time.Sleep(10 * time.Millisecond) // Give time for unregistration
		
		metrics = hub.GetMetrics()
		assert.Equal(t, int64(0), metrics.ActiveConnections)
	})
	
	t.Run("Broadcast to all connections", func(t *testing.T) {
		hub := NewHub()
		go hub.Run()
		
		// Create mock connections
		conn1 := &Connection{
			ID:         uuid.New().String(),
			ProjectIDs: make(map[uuid.UUID]bool),
			send:       make(chan []byte, 10),
		}
		conn2 := &Connection{
			ID:         uuid.New().String(),
			ProjectIDs: make(map[uuid.UUID]bool),
			send:       make(chan []byte, 10),
		}
		
		hub.Register(conn1)
		hub.Register(conn2)
		time.Sleep(10 * time.Millisecond)
		
		// Create and broadcast message
		message, err := NewMessage(TaskCreated, TaskData{
			TaskID:    uuid.New(),
			ProjectID: uuid.New(),
		})
		require.NoError(t, err)
		
		hub.BroadcastToAll(message, nil)
		time.Sleep(10 * time.Millisecond)
		
		// Check that both connections received the message
		assert.Len(t, conn1.send, 1)
		assert.Len(t, conn2.send, 1)
	})
	
	t.Run("Broadcast to project subscribers", func(t *testing.T) {
		hub := NewHub()
		go hub.Run()
		
		projectID := uuid.New()
		
		// Create connections
		conn1 := &Connection{
			ID:         uuid.New().String(),
			ProjectIDs: make(map[uuid.UUID]bool),
			send:       make(chan []byte, 10),
		}
		conn2 := &Connection{
			ID:         uuid.New().String(),
			ProjectIDs: make(map[uuid.UUID]bool),
			send:       make(chan []byte, 10),
		}
		
		hub.Register(conn1)
		hub.Register(conn2)
		time.Sleep(10 * time.Millisecond)
		
		// Subscribe only conn1 to project
		hub.SubscribeConnectionToProject(conn1, projectID)
		time.Sleep(10 * time.Millisecond)
		
		// Broadcast to project
		message, err := NewMessage(TaskCreated, TaskData{
			TaskID:    uuid.New(),
			ProjectID: projectID,
		})
		require.NoError(t, err)
		
		hub.BroadcastToProject(message, projectID, nil)
		time.Sleep(10 * time.Millisecond)
		
		// Only conn1 should receive the message
		assert.Len(t, conn1.send, 1)
		assert.Len(t, conn2.send, 0)
	})
}

func TestRateLimiter(t *testing.T) {
	t.Run("Allow requests within limit", func(t *testing.T) {
		rl := NewRateLimiter(10.0, 5) // 10 req/sec, burst of 5
		connID := "test-conn"
		
		// Should allow burst requests
		for i := 0; i < 5; i++ {
			assert.True(t, rl.Allow(connID))
		}
		
		// Should block after burst
		assert.False(t, rl.Allow(connID))
	})
	
	t.Run("Remove connection", func(t *testing.T) {
		rl := NewRateLimiter(1.0, 1)
		connID := "test-conn"
		
		// Use up the limit
		assert.True(t, rl.Allow(connID))
		assert.False(t, rl.Allow(connID))
		
		// Remove connection
		rl.RemoveConnection(connID)
		
		// Should allow again after removal
		assert.True(t, rl.Allow(connID))
	})
}

func TestErrorHandler(t *testing.T) {
	t.Run("Handle errors within limit", func(t *testing.T) {
		eh := NewErrorHandler(nil, 3, time.Minute)
		conn := &Connection{ID: "test-conn"}
		
		// Should not close connection for first few errors
		assert.False(t, eh.HandleError(conn, assert.AnError, "test"))
		assert.False(t, eh.HandleError(conn, assert.AnError, "test"))
		
		// Should close connection after max errors
		assert.True(t, eh.HandleError(conn, assert.AnError, "test"))
	})
}

func TestMessagePersistence(t *testing.T) {
	t.Run("Store and retrieve messages", func(t *testing.T) {
		persistence := NewInMemoryPersistence(100, time.Hour)
		userID := "test-user"
		projectID := uuid.New()
		
		message, err := NewMessage(TaskCreated, TaskData{
			TaskID:    uuid.New(),
			ProjectID: projectID,
		})
		require.NoError(t, err)
		
		// Store message
		err = persistence.Store(userID, &projectID, message, time.Hour)
		assert.NoError(t, err)
		
		// Retrieve pending messages
		pending, err := persistence.GetPendingMessages(userID)
		assert.NoError(t, err)
		assert.Len(t, pending, 1)
		assert.Equal(t, message.Type, pending[0].Message.Type)
	})
	
	t.Run("Mark message as delivered", func(t *testing.T) {
		persistence := NewInMemoryPersistence(100, time.Hour)
		userID := "test-user"
		
		message, err := NewMessage(TaskCreated, TaskData{})
		require.NoError(t, err)
		
		err = persistence.Store(userID, nil, message, time.Hour)
		require.NoError(t, err)
		
		// Get pending messages
		pending, err := persistence.GetPendingMessages(userID)
		require.NoError(t, err)
		require.Len(t, pending, 1)
		
		// Mark as delivered
		err = persistence.MarkAsDelivered(pending[0].ID)
		assert.NoError(t, err)
		
		// Should have no pending messages
		pending, err = persistence.GetPendingMessages(userID)
		assert.NoError(t, err)
		assert.Len(t, pending, 0)
	})
	
	t.Run("Cleanup expired messages", func(t *testing.T) {
		persistence := NewInMemoryPersistence(100, time.Hour)
		userID := "test-user"
		
		message, err := NewMessage(TaskCreated, TaskData{})
		require.NoError(t, err)
		
		// Store message with very short TTL
		err = persistence.Store(userID, nil, message, time.Nanosecond)
		require.NoError(t, err)
		
		// Wait for expiration
		time.Sleep(time.Millisecond)
		
		// Cleanup
		err = persistence.CleanupExpired()
		assert.NoError(t, err)
		
		// Should have no pending messages
		pending, err := persistence.GetPendingMessages(userID)
		assert.NoError(t, err)
		assert.Len(t, pending, 0)
	})
}

func TestProcessors(t *testing.T) {
	t.Run("Task event processor", func(t *testing.T) {
		hub := NewHub()
		processor := NewTaskEventProcessor(hub)
		
		conn := &Connection{
			ID:         "test-conn",
			ProjectIDs: make(map[uuid.UUID]bool),
			send:       make(chan []byte, 10),
		}
		
		data := TaskData{
			TaskID:    uuid.New(),
			ProjectID: uuid.New(),
		}
		
		message, err := NewMessage(TaskCreated, data)
		require.NoError(t, err)
		
		err = processor.ProcessMessage(conn, message)
		assert.NoError(t, err)
	})
	
	t.Run("Auth processor", func(t *testing.T) {
		hub := NewHub()
		authService := NewMockAuthService()
		processor := NewAuthProcessor(authService, hub)
		
		conn := &Connection{
			ID:         "test-conn",
			ProjectIDs: make(map[uuid.UUID]bool),
			send:       make(chan []byte, 10),
		}
		
		authData := AuthData{
			Token:  "valid-token",
			UserID: "test-user",
		}
		
		message, err := NewMessage(AuthRequired, authData)
		require.NoError(t, err)
		
		err = processor.ProcessMessage(conn, message)
		assert.NoError(t, err)
		
		// Check that connection received auth success message
		assert.Len(t, conn.send, 1)
		
		// Check that user ID was set
		assert.Equal(t, "user-123", conn.GetUserID())
	})
}

func TestAuthService(t *testing.T) {
	t.Run("Mock auth service validates tokens", func(t *testing.T) {
		authService := NewMockAuthService()
		
		// Valid token
		userInfo, err := authService.ValidateToken("valid-token")
		assert.NoError(t, err)
		assert.NotNil(t, userInfo)
		assert.Equal(t, "user-123", userInfo.ID)
		
		// Invalid token
		userInfo, err = authService.ValidateToken("")
		assert.Error(t, err)
		assert.Nil(t, userInfo)
	})
}

func BenchmarkMessageCreation(b *testing.B) {
	data := TaskData{
		TaskID:    uuid.New(),
		ProjectID: uuid.New(),
		Changes:   map[string]interface{}{"status": "completed"},
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := NewMessage(TaskUpdated, data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMessageSerialization(b *testing.B) {
	data := TaskData{
		TaskID:    uuid.New(),
		ProjectID: uuid.New(),
		Changes:   map[string]interface{}{"status": "completed"},
	}
	
	message, err := NewMessage(TaskUpdated, data)
	if err != nil {
		b.Fatal(err)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := message.ToBytes()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkHubBroadcast(b *testing.B) {
	hub := NewHub()
	go hub.Run()
	
	// Create multiple connections
	connections := make([]*Connection, 100)
	for i := 0; i < 100; i++ {
		conn := &Connection{
			ID:         uuid.New().String(),
			ProjectIDs: make(map[uuid.UUID]bool),
			send:       make(chan []byte, 100),
		}
		connections[i] = conn
		hub.Register(conn)
	}
	
	time.Sleep(10 * time.Millisecond) // Let registrations complete
	
	message, err := NewMessage(TaskCreated, TaskData{
		TaskID:    uuid.New(),
		ProjectID: uuid.New(),
	})
	if err != nil {
		b.Fatal(err)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hub.BroadcastToAll(message, nil)
	}
}