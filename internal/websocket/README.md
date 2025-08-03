# WebSocket Infrastructure

This package provides comprehensive WebSocket infrastructure for real-time communication in the Auto-Devs project.

## Features

- **Real-time messaging** with JSON message protocol
- **Room-based messaging** for project-specific communications
- **Connection management** with health monitoring and ping/pong heartbeat
- **Authentication and authorization** with token-based security
- **Rate limiting** and error handling middleware
- **Message persistence** for offline delivery
- **Comprehensive testing** with unit and integration tests

## Architecture

### Core Components

- **Hub** (`hub.go`): Central message routing and connection management
- **Connection** (`connection.go`): Individual WebSocket connection handling
- **Message** (`message.go`): Message protocol and data structures
- **Service** (`service.go`): High-level WebSocket service interface
- **Handler** (`handler.go`): HTTP/WebSocket upgrade handling

### Middleware

- **Authentication** (`auth.go`): Token validation and user management
- **Rate Limiting** (`middleware.go`): Request throttling and abuse prevention
- **Error Handling** (`middleware.go`): Graceful error management
- **Logging** (`middleware.go`): Comprehensive request/response logging

### Message Processing

- **Processors** (`processors.go`): Message-type specific handlers
- **Persistence** (`persistence.go`): Offline message storage and delivery

## Message Protocol

All WebSocket messages follow this JSON structure:

```json
{
  "type": "message_type",
  "data": { ... },
  "timestamp": "2023-12-01T10:00:00Z",
  "message_id": "uuid"
}
```

### Message Types

#### Task Events
- `task_created`: New task creation
- `task_updated`: Task modifications
- `task_deleted`: Task removal

#### Project Events
- `project_updated`: Project modifications

#### Status Events
- `status_changed`: Entity status transitions

#### User Presence
- `user_joined`: User joins project
- `user_left`: User leaves project

#### System Messages
- `ping`/`pong`: Connection health checks
- `auth_required`/`auth_success`/`auth_failed`: Authentication flow
- `error`: Error notifications

## Usage

### Basic Setup

```go
// Initialize WebSocket service
wsService := websocket.NewService()

// Get handler for HTTP routes
wsHandler := wsService.GetHandler()

// Setup routes
router.GET("/ws/connect", wsHandler.HandleWebSocket)
```

### Sending Notifications

```go
// Task created notification
err := wsService.NotifyTaskCreated(task, projectID)

// Task updated notification
changes := map[string]interface{}{
    "status": map[string]interface{}{
        "old": "TODO",
        "new": "IN_PROGRESS",
    },
}
err := wsService.NotifyTaskUpdated(taskID, projectID, changes, updatedTask)

// Direct user message
err := wsService.SendDirectMessage(userID, websocket.TaskCreated, taskData)

// Project broadcast
err := wsService.SendProjectMessage(projectID, websocket.ProjectUpdated, projectData)
```

### Client Connection

```javascript
// Connect to WebSocket
const ws = new WebSocket('ws://localhost:8098/ws/connect?token=your-auth-token');

// Handle messages
ws.onmessage = (event) => {
    const message = JSON.parse(event.data);
    console.log('Received:', message.type, message.data);
};

// Subscribe to project
ws.send(JSON.stringify({
    type: 'subscription',
    data: {
        action: 'subscribe',
        project_id: 'project-uuid'
    }
}));
```

## Configuration

### Default Settings

- **Rate Limiting**: 10 requests/second, burst of 20
- **Connection Timeout**: 60 seconds
- **Ping Interval**: 54 seconds
- **Max Message Size**: 512 bytes
- **Message Persistence**: 1000 messages, 24-hour TTL

### Custom Configuration

```go
config := &websocket.ServiceConfig{
    RequestsPerSecond:  5.0,
    BurstSize:         10,
    MaxErrors:         5,
    ErrorResetInterval: 10 * time.Minute,
    MaxStoredMessages:  500,
    MessageTTL:        12 * time.Hour,
}

wsService := websocket.NewServiceWithConfig(config)
```

## Monitoring

### Health Checks

```bash
# WebSocket service health
GET /api/v1/websocket/health

# Connection metrics
GET /api/v1/websocket/metrics

# Connection counts
GET /api/v1/websocket/connections/count
GET /api/v1/websocket/projects/{id}/connections/count
GET /api/v1/websocket/users/{id}/connections/count
```

### Metrics

The service provides comprehensive metrics:

```json
{
  "hub": {
    "active_connections": 42,
    "total_connections": 150,
    "messages_sent": 1234,
    "messages_received": 890,
    "broadcasts_sent": 67
  },
  "middleware": {
    "rate_limiter": {
      "active_limiters": 42,
      "requests_per_second": 10.0
    },
    "error_handler": {
      "connections_with_errors": 3,
      "total_errors": 12
    }
  },
  "offline_messages": {
    "total_messages": 15,
    "users_with_messages": 5
  }
}
```

## Security

### Authentication

- Token-based authentication via query parameter or Authorization header
- User association with connections
- Role-based access control for admin endpoints

### Rate Limiting

- Per-connection request throttling
- Configurable limits and burst allowances
- Automatic cleanup of inactive limiters

### Error Handling

- Connection-specific error tracking
- Automatic disconnection after error threshold
- Graceful degradation and recovery

## Testing

Run the comprehensive test suite:

```bash
# Run all WebSocket tests
go test ./internal/websocket -v

# Run with coverage
go test ./internal/websocket -v -cover

# Run benchmarks
go test ./internal/websocket -bench=.
```

### Test Coverage

- Message creation and parsing
- Hub registration and broadcasting
- Connection management
- Rate limiting
- Error handling
- Message persistence
- Authentication flows

## Integration Examples

See `example_usage.go` for comprehensive usage examples including:

- Basic service usage
- Custom message types
- Message handling patterns
- Connection management
- Error handling strategies
- HTTP handler integration

## WebSocket Routes

### Connection
- `GET /ws/connect` - WebSocket upgrade endpoint

### Management API
- `GET /api/v1/websocket/connections` - List active connections
- `GET /api/v1/websocket/metrics` - Service metrics
- `POST /api/v1/websocket/broadcast` - Broadcast message
- `GET /api/v1/websocket/health` - Health status
- `GET /api/v1/websocket/stats` - Statistics

### Administrative
- `POST /api/v1/websocket/users/{id}/disconnect` - Disconnect user
- `POST /api/v1/websocket/projects/{id}/disconnect` - Disconnect project
- `POST /api/v1/websocket/users/{id}/message` - Send direct message
- `POST /api/v1/websocket/projects/{id}/message` - Send project message

## Error Codes

Common WebSocket error codes:

- `invalid_message`: Malformed message format
- `rate_limited`: Too many requests
- `unauthorized`: Authentication required
- `processing_error`: Message processing failed
- `connection_closed`: Connection terminated

## Future Enhancements

- Database-backed message persistence
- Horizontal scaling with Redis pub/sub
- Message acknowledgment and delivery guarantees
- Advanced presence tracking
- Message history and replay
- File upload support
- Custom message compression