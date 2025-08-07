# WebSocket Infrastructure

This package provides comprehensive WebSocket infrastructure for real-time communication in the Auto-Devs project.

## Features

- **Real-time messaging** with JSON message protocol
- **Room-based messaging** for project-specific communications
- **Connection management** with health monitoring and ping/pong heartbeat
- **Authentication and authorization** with token-based security
- **Rate limiting** and error handling middleware
- **Message persistence** for offline delivery
- **Cross-process messaging** via Redis Pub/Sub broker
- **Comprehensive testing** with unit and integration tests

## Architecture

### Core Components

- **Hub** (`hub.go`): Central message routing and connection management
- **Connection** (`connection.go`): Individual WebSocket connection handling
- **Message** (`message.go`): Message protocol and data structures
- **Service** (`service.go`): High-level WebSocket service interface
- **Handler** (`handler.go`): HTTP/WebSocket upgrade handling
- **RedisBroker** (`redis_broker.go`): Redis Pub/Sub for cross-process messaging

### Middleware

- **Authentication** (`auth.go`): Token validation and user management
- **Rate Limiting** (`middleware.go`): Request throttling and abuse prevention
- **Error Handling** (`middleware.go`): Graceful error management
- **Logging** (`middleware.go`): Comprehensive request/response logging

### Message Processing

- **Processors** (`processors.go`): Message-type specific handlers
- **Persistence** (`persistence.go`): Offline message storage and delivery

## Cross-Process Messaging with Redis

### Problem

WebSocket Hub is **in-memory managed**, meaning it only exists in the server process. When worker processes need to send messages to WebSocket clients, they can't directly access the Hub.

### Solution

Redis Pub/Sub broker enables cross-process messaging:

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Frontend  │    │    Server   │    │   Worker    │    │    Redis    │
│   Client    │    │             │    │             │    │    Broker   │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
       │                   │                   │                   │
       │ 1. WebSocket      │                   │                   │
       │    Connection     │                   │                   │
       │ ──────────────────►│                   │                   │
       │                   │                   │                   │
       │                   │ 2. Job Enqueued   │                   │
       │                   │ ─────────────────────────────────────►│
       │                   │                   │                   │
       │                   │                   │ 3. Process Job    │
       │                   │                   │ ◄─────────────────│
       │                   │                   │                   │
       │                   │                   │ 4. Update DB      │
       │                   │                   │                   │
       │                   │                   │ 5. Redis Pub      │
       │                   │                   │ ──────────────────►│
       │                   │                   │                   │
       │                   │ 6. Redis Sub      │                   │
       │                   │ ◄─────────────────│                   │
       │                   │                   │                   │
       │ 7. Real-time      │                   │                   │
       │    Update         │                   │                   │
       │ ◄─────────────────│                   │                   │
       │                   │                   │                   │
```

### Implementation

#### Server Setup

```go
// Create WebSocket service with Redis broker
wsService := websocket.NewServiceWithRedisBroker("localhost:6379", "", 0)

// Start Redis broker
if err := wsService.StartRedisBroker(); err != nil {
    log.Fatal("Failed to start Redis broker:", err)
}
defer wsService.StopRedisBroker()

// Setup routes
handler.SetupRoutes(router, wsService)
```

#### Worker Setup

```go
// Create Redis broker client
redisClient := jobs.NewRedisBrokerClient("localhost:6379", "", 0)
defer redisClient.Close()

// Test connection
if err := redisClient.TestConnection(); err != nil {
    log.Fatal("Failed to connect to Redis:", err)
}

// Create processor with Redis broker
processor := jobs.NewProcessorWithRedisBroker(
    taskUsecase,
    projectUsecase,
    worktreeUsecase,
    planningService,
    executionService,
    planRepo,
    wsService,
    redisClient,
)
```

#### Message Publishing

```go
// Worker publishes task update
changes := map[string]interface{}{
    "status": map[string]interface{}{
        "old": "TODO",
        "new": "IN_PROGRESS",
    },
}

taskResponse := map[string]interface{}{
    "id":         taskID.String(),
    "project_id": projectID.String(),
    "title":      "Task title",
    "status":     "IN_PROGRESS",
    "updated_at": time.Now(),
}

// Publish via Redis broker
err := redisClient.PublishTaskUpdated(taskID, projectID, changes, taskResponse)
if err != nil {
    log.Error("Failed to publish task update:", err)
}
```

### Message Format

Redis broker messages follow this JSON structure:

```json
{
  "type": "task_updated",
  "data": {
    "task_id": "uuid",
    "project_id": "uuid",
    "changes": {
      "status": {
        "old": "TODO",
        "new": "IN_PROGRESS"
      }
    },
    "task": {
      "id": "uuid",
      "title": "Task title",
      "status": "IN_PROGRESS"
    }
  },
  "project_id": "uuid",
  "timestamp": "2023-12-01T10:00:00Z",
  "message_id": "uuid",
  "source": "worker"
}
```

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
const ws = new WebSocket(
  "ws://localhost:8098/ws/connect?token=your-auth-token"
);

// Handle messages
ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  console.log("Received:", message.type, message.data);
};

// Subscribe to project
ws.send(
  JSON.stringify({
    type: "subscription",
    data: {
      action: "subscribe",
      project_id: "project-uuid",
    },
  })
);
```

## Configuration

### Default Settings

- **Rate Limiting**: 10 requests/second, burst of 20
- **Connection Timeout**: 60 seconds
- **Ping Interval**: 54 seconds
- **Max Message Size**: 512 bytes
- **Message Persistence**: 1000 messages, 24-hour TTL
- **Redis Channel**: `websocket:broadcast`

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

## Examples

See `examples/redis_broker_example.go` for complete examples of:

- Server setup with Redis broker
- Worker setup with Redis broker client
- Cross-process messaging
- Integration patterns

## Error Handling

The system includes comprehensive error handling:

1. **Connection failures** are logged but don't stop job processing
2. **Message sending failures** are logged with fallback options
3. **Automatic reconnection** for Redis connections
4. **Graceful degradation** when Redis is unavailable
5. **Fallback to in-memory** when Redis broker fails

This provides a robust, fault-tolerant system for real-time notifications across multiple processes.
