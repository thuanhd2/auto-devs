package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// WebSocketTestSuite provides comprehensive WebSocket testing
type WebSocketTestSuite struct {
	server   *httptest.Server
	hub      *Hub
	upgrader websocket.Upgrader
}

// SetupWebSocketTestSuite creates a test suite for WebSocket testing
func SetupWebSocketTestSuite(t *testing.T) (*WebSocketTestSuite, func()) {
	// Create hub
	hub := NewHub()
	go hub.Run()

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// WebSocket upgrader
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins for testing
		},
	}

	// WebSocket endpoint
	router.GET("/ws", func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			t.Logf("WebSocket upgrade failed: %v", err)
			return
		}

		// Create connection wrapper
		wsConn := &Connection{
			ID:         uuid.New().String(),
			conn:       conn,
			send:       make(chan []byte, 256),
			ProjectIDs: make(map[uuid.UUID]bool),
			hub:        hub,
		}

		// Register connection
		hub.Register(wsConn)

		// Start connection handlers
		go wsConn.writePump()
		go wsConn.readPump()
	})

	// Start test server
	server := httptest.NewServer(router)

	suite := &WebSocketTestSuite{
		server:   server,
		hub:      hub,
		upgrader: upgrader,
	}

	return suite, func() {
		server.Close()
		hub.Stop()
		gin.SetMode(gin.ReleaseMode)
	}
}

// connectWebSocket creates a WebSocket connection to the test server
func (suite *WebSocketTestSuite) connectWebSocket(t *testing.T) *websocket.Conn {
	// Convert http URL to ws URL
	url := "ws" + strings.TrimPrefix(suite.server.URL, "http") + "/ws"
	
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	require.NoError(t, err)
	
	return conn
}

// TestWebSocket_ConnectionHandling tests basic connection handling
func TestWebSocket_ConnectionHandling(t *testing.T) {
	suite, cleanup := SetupWebSocketTestSuite(t)
	defer cleanup()

	t.Run("successful connection", func(t *testing.T) {
		conn := suite.connectWebSocket(t)
		defer conn.Close()

		// Connection should be registered
		time.Sleep(50 * time.Millisecond) // Allow registration
		metrics := suite.hub.GetMetrics()
		assert.Equal(t, int64(1), metrics.ActiveConnections)
	})

	t.Run("multiple connections", func(t *testing.T) {
		const numConnections = 5
		connections := make([]*websocket.Conn, numConnections)

		// Create multiple connections
		for i := 0; i < numConnections; i++ {
			connections[i] = suite.connectWebSocket(t)
		}

		// Allow registration time
		time.Sleep(100 * time.Millisecond)

		// Check metrics
		metrics := suite.hub.GetMetrics()
		assert.Equal(t, int64(numConnections), metrics.ActiveConnections)

		// Close all connections
		for _, conn := range connections {
			conn.Close()
		}

		// Wait for cleanup
		time.Sleep(100 * time.Millisecond)

		// Should be no active connections
		metrics = suite.hub.GetMetrics()
		assert.Equal(t, int64(0), metrics.ActiveConnections)
	})

	t.Run("connection cleanup on close", func(t *testing.T) {
		conn := suite.connectWebSocket(t)
		
		// Wait for registration
		time.Sleep(50 * time.Millisecond)
		
		initialMetrics := suite.hub.GetMetrics()
		assert.Equal(t, int64(1), initialMetrics.ActiveConnections)

		// Close connection
		conn.Close()

		// Wait for cleanup
		time.Sleep(100 * time.Millisecond)

		// Connection should be unregistered
		finalMetrics := suite.hub.GetMetrics()
		assert.Equal(t, int64(0), finalMetrics.ActiveConnections)
	})
}

// TestWebSocket_MessageBroadcasting tests message broadcasting functionality
func TestWebSocket_MessageBroadcasting(t *testing.T) {
	suite, cleanup := SetupWebSocketTestSuite(t)
	defer cleanup()

	t.Run("broadcast to all connections", func(t *testing.T) {
		const numConnections = 3
		connections := make([]*websocket.Conn, numConnections)

		// Create connections
		for i := 0; i < numConnections; i++ {
			connections[i] = suite.connectWebSocket(t)
			defer connections[i].Close()
		}

		// Allow registration
		time.Sleep(100 * time.Millisecond)

		// Create test message
		testData := TaskData{
			TaskID:    uuid.New(),
			ProjectID: uuid.New(),
			Changes:   map[string]interface{}{"status": "completed"},
		}

		message, err := NewMessage(TaskUpdated, testData)
		require.NoError(t, err)

		// Broadcast message
		suite.hub.BroadcastToAll(message, nil)

		// Wait for message delivery
		time.Sleep(50 * time.Millisecond)

		// All connections should receive the message
		for i, conn := range connections {
			conn.SetReadDeadline(time.Now().Add(1 * time.Second))
			
			var receivedMessage Message
			err := conn.ReadJSON(&receivedMessage)
			require.NoError(t, err, "Connection %d should receive message", i)
			
			assert.Equal(t, message.Type, receivedMessage.Type)
			assert.Equal(t, message.MessageID, receivedMessage.MessageID)
		}
	})

	t.Run("broadcast to project subscribers", func(t *testing.T) {
		conn1 := suite.connectWebSocket(t)
		conn2 := suite.connectWebSocket(t)
		defer conn1.Close()
		defer conn2.Close()

		// Allow registration
		time.Sleep(100 * time.Millisecond)

		projectID := uuid.New()

		// Subscribe conn1 to project (this would normally be done via message handling)
		// For testing, we'll access hub directly
		connections := suite.hub.GetConnections()
		require.Len(t, connections, 2)

		// Subscribe first connection to project
		connections[0].ProjectIDs[projectID] = true

		// Create test message
		testData := TaskData{
			TaskID:    uuid.New(),
			ProjectID: projectID,
		}

		message, err := NewMessage(TaskCreated, testData)
		require.NoError(t, err)

		// Broadcast to project
		suite.hub.BroadcastToProject(message, projectID, nil)

		// Wait for delivery
		time.Sleep(50 * time.Millisecond)

		// Only subscribed connection should receive message
		conn1.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
		var receivedMessage Message
		err = conn1.ReadJSON(&receivedMessage)
		require.NoError(t, err, "Subscribed connection should receive message")

		// Non-subscribed connection should not receive message
		conn2.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
		err = conn2.ReadJSON(&receivedMessage)
		assert.Error(t, err, "Non-subscribed connection should not receive message")
	})

	t.Run("exclude sender from broadcast", func(t *testing.T) {
		conn1 := suite.connectWebSocket(t)
		conn2 := suite.connectWebSocket(t)
		defer conn1.Close()
		defer conn2.Close()

		// Allow registration
		time.Sleep(100 * time.Millisecond)

		connections := suite.hub.GetConnections()
		require.Len(t, connections, 2)

		senderConn := connections[0]

		// Create test message
		testData := TaskData{
			TaskID:    uuid.New(),
			ProjectID: uuid.New(),
		}

		message, err := NewMessage(TaskUpdated, testData)
		require.NoError(t, err)

		// Broadcast excluding sender
		suite.hub.BroadcastToAll(message, senderConn)

		// Wait for delivery
		time.Sleep(50 * time.Millisecond)

		// Sender should not receive message
		conn1.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
		var receivedMessage Message
		err = conn1.ReadJSON(&receivedMessage)
		assert.Error(t, err, "Sender should not receive message")

		// Other connection should receive message
		conn2.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
		err = conn2.ReadJSON(&receivedMessage)
		require.NoError(t, err, "Non-sender should receive message")
		assert.Equal(t, message.Type, receivedMessage.Type)
	})
}

// TestWebSocket_MessageTypes tests different message types
func TestWebSocket_MessageTypes(t *testing.T) {
	suite, cleanup := SetupWebSocketTestSuite(t)
	defer cleanup()

	t.Run("task messages", func(t *testing.T) {
		conn := suite.connectWebSocket(t)
		defer conn.Close()

		time.Sleep(50 * time.Millisecond)

		testCases := []struct {
			messageType MessageType
			data        interface{}
		}{
			{
				messageType: TaskCreated,
				data: TaskData{
					TaskID:    uuid.New(),
					ProjectID: uuid.New(),
					Changes:   map[string]interface{}{"title": "New Task"},
				},
			},
			{
				messageType: TaskUpdated,
				data: TaskData{
					TaskID:    uuid.New(),
					ProjectID: uuid.New(),
					Changes:   map[string]interface{}{"status": "in_progress"},
				},
			},
			{
				messageType: TaskDeleted,
				data: TaskData{
					TaskID:    uuid.New(),
					ProjectID: uuid.New(),
				},
			},
		}

		for _, tc := range testCases {
			t.Run(string(tc.messageType), func(t *testing.T) {
				message, err := NewMessage(tc.messageType, tc.data)
				require.NoError(t, err)

				// Send message
				suite.hub.BroadcastToAll(message, nil)

				// Receive and verify
				conn.SetReadDeadline(time.Now().Add(1 * time.Second))
				var receivedMessage Message
				err = conn.ReadJSON(&receivedMessage)
				require.NoError(t, err)

				assert.Equal(t, tc.messageType, receivedMessage.Type)
				assert.NotEmpty(t, receivedMessage.MessageID)
				assert.False(t, receivedMessage.Timestamp.IsZero())
			})
		}
	})

	t.Run("project messages", func(t *testing.T) {
		conn := suite.connectWebSocket(t)
		defer conn.Close()

		time.Sleep(50 * time.Millisecond)

		projectData := ProjectData{
			ProjectID: uuid.New(),
			Changes:   map[string]interface{}{"name": "Updated Project"},
		}

		message, err := NewMessage(ProjectUpdated, projectData)
		require.NoError(t, err)

		suite.hub.BroadcastToAll(message, nil)

		conn.SetReadDeadline(time.Now().Add(1 * time.Second))
		var receivedMessage Message
		err = conn.ReadJSON(&receivedMessage)
		require.NoError(t, err)

		assert.Equal(t, ProjectUpdated, receivedMessage.Type)

		// Parse data back
		var receivedData ProjectData
		err = receivedMessage.ParseData(&receivedData)
		require.NoError(t, err)
		assert.Equal(t, projectData.ProjectID, receivedData.ProjectID)
	})
}

// TestWebSocket_ErrorHandling tests error scenarios
func TestWebSocket_ErrorHandling(t *testing.T) {
	suite, cleanup := SetupWebSocketTestSuite(t)
	defer cleanup()

	t.Run("connection error handling", func(t *testing.T) {
		conn := suite.connectWebSocket(t)
		
		// Allow registration
		time.Sleep(50 * time.Millisecond)
		
		initialCount := suite.hub.GetMetrics().ActiveConnections
		assert.Equal(t, int64(1), initialCount)

		// Force close connection abruptly
		conn.Close()

		// Wait for cleanup
		time.Sleep(100 * time.Millisecond)

		// Connection should be cleaned up
		finalCount := suite.hub.GetMetrics().ActiveConnections
		assert.Equal(t, int64(0), finalCount)
	})

	t.Run("invalid message handling", func(t *testing.T) {
		conn := suite.connectWebSocket(t)
		defer conn.Close()

		time.Sleep(50 * time.Millisecond)

		// Send invalid JSON
		invalidJSON := `{"invalid": json}`
		err := conn.WriteMessage(websocket.TextMessage, []byte(invalidJSON))
		require.NoError(t, err)

		// Connection should remain active (error should be handled gracefully)
		time.Sleep(100 * time.Millisecond)
		
		metrics := suite.hub.GetMetrics()
		assert.Equal(t, int64(1), metrics.ActiveConnections)
	})

	t.Run("large message handling", func(t *testing.T) {
		conn := suite.connectWebSocket(t)
		defer conn.Close()

		time.Sleep(50 * time.Millisecond)

		// Create large message
		largeData := make(map[string]interface{})
		for i := 0; i < 1000; i++ {
			largeData[fmt.Sprintf("key_%d", i)] = fmt.Sprintf("value_%d", i)
		}

		taskData := TaskData{
			TaskID:    uuid.New(),
			ProjectID: uuid.New(),
			Changes:   largeData,
		}

		message, err := NewMessage(TaskUpdated, taskData)
		require.NoError(t, err)

		// Broadcast large message
		suite.hub.BroadcastToAll(message, nil)

		// Should be able to receive large message
		conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		var receivedMessage Message
		err = conn.ReadJSON(&receivedMessage)
		require.NoError(t, err)

		assert.Equal(t, TaskUpdated, receivedMessage.Type)
	})
}

// TestWebSocket_ConcurrentOperations tests concurrent WebSocket operations
func TestWebSocket_ConcurrentOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent WebSocket test in short mode")
	}

	suite, cleanup := SetupWebSocketTestSuite(t)
	defer cleanup()

	t.Run("concurrent connections", func(t *testing.T) {
		const numGoroutines = 20
		var wg sync.WaitGroup
		connections := make(chan *websocket.Conn, numGoroutines)
		errors := make(chan error, numGoroutines)

		// Create connections concurrently
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				
				conn, _, err := websocket.DefaultDialer.Dial(
					"ws"+strings.TrimPrefix(suite.server.URL, "http")+"/ws", 
					nil,
				)
				
				if err != nil {
					errors <- err
				} else {
					connections <- conn
				}
			}(i)
		}

		wg.Wait()
		close(connections)
		close(errors)

		// Check for errors
		errorCount := 0
		for err := range errors {
			t.Logf("Connection error: %v", err)
			errorCount++
		}

		// Count successful connections
		successfulConns := make([]*websocket.Conn, 0)
		for conn := range connections {
			successfulConns = append(successfulConns, conn)
		}

		assert.LessOrEqual(t, errorCount, numGoroutines/4, "Most connections should succeed")
		assert.GreaterOrEqual(t, len(successfulConns), numGoroutines*3/4, "Most connections should succeed")

		// Wait for registration
		time.Sleep(200 * time.Millisecond)

		// Check metrics
		metrics := suite.hub.GetMetrics()
		assert.Equal(t, int64(len(successfulConns)), metrics.ActiveConnections)

		// Close all connections
		for _, conn := range successfulConns {
			conn.Close()
		}
	})

	t.Run("concurrent message broadcasting", func(t *testing.T) {
		// Create connections
		const numConnections = 5
		connections := make([]*websocket.Conn, numConnections)
		
		for i := 0; i < numConnections; i++ {
			connections[i] = suite.connectWebSocket(t)
			defer connections[i].Close()
		}

		time.Sleep(100 * time.Millisecond)

		// Broadcast messages concurrently
		const numMessages = 10
		var wg sync.WaitGroup
		
		for i := 0; i < numMessages; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				
				testData := TaskData{
					TaskID:    uuid.New(),
					ProjectID: uuid.New(),
					Changes:   map[string]interface{}{"index": i},
				}

				message, err := NewMessage(TaskUpdated, testData)
				if err != nil {
					t.Errorf("Failed to create message: %v", err)
					return
				}

				suite.hub.BroadcastToAll(message, nil)
			}(i)
		}

		wg.Wait()

		// Wait for message delivery
		time.Sleep(200 * time.Millisecond)

		// Each connection should receive multiple messages
		for i, conn := range connections {
			messageCount := 0
			conn.SetReadDeadline(time.Now().Add(1 * time.Second))
			
			for {
				var message Message
				err := conn.ReadJSON(&message)
				if err != nil {
					break // Timeout or no more messages
				}
				messageCount++
			}

			assert.GreaterOrEqual(t, messageCount, numMessages/2, 
				"Connection %d should receive at least half of the messages", i)
		}
	})

	t.Run("concurrent subscription changes", func(t *testing.T) {
		conn := suite.connectWebSocket(t)
		defer conn.Close()

		time.Sleep(50 * time.Millisecond)

		connections := suite.hub.GetConnections()
		require.Len(t, connections, 1)
		testConn := connections[0]

		// Modify subscriptions concurrently
		const numGoroutines = 10
		var wg sync.WaitGroup
		projectIDs := make([]uuid.UUID, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			projectIDs[i] = uuid.New()
		}

		// Subscribe concurrently
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				suite.hub.SubscribeConnectionToProject(testConn, projectIDs[i])
			}(i)
		}

		wg.Wait()

		// Wait for processing
		time.Sleep(50 * time.Millisecond)

		// Check subscriptions
		assert.Len(t, testConn.ProjectIDs, numGoroutines)
		for _, projectID := range projectIDs {
			assert.True(t, testConn.ProjectIDs[projectID], 
				"Should be subscribed to project %s", projectID)
		}

		// Unsubscribe concurrently
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				suite.hub.UnsubscribeConnectionFromProject(testConn, projectIDs[i])
			}(i)
		}

		wg.Wait()

		// Wait for processing
		time.Sleep(50 * time.Millisecond)

		// Should have no subscriptions
		assert.Len(t, testConn.ProjectIDs, 0)
	})
}

// TestWebSocket_RateLimiting tests rate limiting functionality
func TestWebSocket_RateLimiting(t *testing.T) {
	suite, cleanup := SetupWebSocketTestSuite(t)
	defer cleanup()

	t.Run("rate limiter functionality", func(t *testing.T) {
		rateLimiter := NewRateLimiter(5.0, 2) // 5 per second, burst of 2
		connID := uuid.New().String()

		// Should allow burst requests
		assert.True(t, rateLimiter.Allow(connID))
		assert.True(t, rateLimiter.Allow(connID))

		// Should block after burst
		assert.False(t, rateLimiter.Allow(connID))

		// Wait for rate limit to reset
		time.Sleep(200 * time.Millisecond)

		// Should allow again
		assert.True(t, rateLimiter.Allow(connID))
	})

	t.Run("rate limiter cleanup", func(t *testing.T) {
		rateLimiter := NewRateLimiter(1.0, 1)
		connID := uuid.New().String()

		// Use up the limit
		assert.True(t, rateLimiter.Allow(connID))
		assert.False(t, rateLimiter.Allow(connID))

		// Remove connection
		rateLimiter.RemoveConnection(connID)

		// Should allow again after removal
		assert.True(t, rateLimiter.Allow(connID))
	})
}

// TestWebSocket_MessagePersistence tests message persistence functionality
func TestWebSocket_MessagePersistence(t *testing.T) {
	suite, cleanup := SetupWebSocketTestSuite(t)
	defer cleanup()

	t.Run("message storage and retrieval", func(t *testing.T) {
		persistence := NewInMemoryPersistence(100, time.Hour)
		userID := "test-user"
		projectID := uuid.New()

		// Store message
		taskData := TaskData{
			TaskID:    uuid.New(),
			ProjectID: projectID,
		}

		message, err := NewMessage(TaskCreated, taskData)
		require.NoError(t, err)

		err = persistence.Store(userID, &projectID, message, time.Hour)
		require.NoError(t, err)

		// Retrieve pending messages
		pending, err := persistence.GetPendingMessages(userID)
		require.NoError(t, err)
		assert.Len(t, pending, 1)
		assert.Equal(t, message.MessageID, pending[0].Message.MessageID)
	})

	t.Run("message delivery tracking", func(t *testing.T) {
		persistence := NewInMemoryPersistence(100, time.Hour)
		userID := "test-user"

		message, err := NewMessage(TaskCreated, TaskData{})
		require.NoError(t, err)

		// Store message
		err = persistence.Store(userID, nil, message, time.Hour)
		require.NoError(t, err)

		// Get pending messages
		pending, err := persistence.GetPendingMessages(userID)
		require.NoError(t, err)
		require.Len(t, pending, 1)

		// Mark as delivered
		err = persistence.MarkAsDelivered(pending[0].ID)
		require.NoError(t, err)

		// Should have no pending messages
		pending, err = persistence.GetPendingMessages(userID)
		require.NoError(t, err)
		assert.Len(t, pending, 0)
	})

	t.Run("message expiration", func(t *testing.T) {
		persistence := NewInMemoryPersistence(100, time.Hour)
		userID := "test-user"

		message, err := NewMessage(TaskCreated, TaskData{})
		require.NoError(t, err)

		// Store message with very short TTL
		err = persistence.Store(userID, nil, message, time.Nanosecond)
		require.NoError(t, err)

		// Wait for expiration
		time.Sleep(time.Millisecond)

		// Cleanup expired messages
		err = persistence.CleanupExpired()
		require.NoError(t, err)

		// Should have no pending messages
		pending, err := persistence.GetPendingMessages(userID)
		require.NoError(t, err)
		assert.Len(t, pending, 0)
	})
}