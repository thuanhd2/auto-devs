package websocket

import (
	"log"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// LoggingMiddleware provides logging for WebSocket connections
type LoggingMiddleware struct {
	logger *log.Logger
}

// NewLoggingMiddleware creates a new logging middleware
func NewLoggingMiddleware(logger *log.Logger) *LoggingMiddleware {
	if logger == nil {
		logger = log.Default()
	}
	
	return &LoggingMiddleware{
		logger: logger,
	}
}

// LogConnection logs connection events
func (m *LoggingMiddleware) LogConnection(conn *Connection, event string, details map[string]interface{}) {
	m.logger.Printf("[WS] Connection %s - %s: %v", conn.ID, event, details)
}

// LogMessage logs message events
func (m *LoggingMiddleware) LogMessage(conn *Connection, direction string, msgType MessageType, size int) {
	m.logger.Printf("[WS] Connection %s - %s message: type=%s, size=%d bytes", 
		conn.ID, direction, msgType, size)
}

// LogError logs error events
func (m *LoggingMiddleware) LogError(conn *Connection, err error, context string) {
	m.logger.Printf("[WS] Connection %s - Error in %s: %v", conn.ID, context, err)
}

// RateLimiter manages rate limiting for WebSocket connections
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	
	// Rate limiting configuration
	requestsPerSecond rate.Limit
	burstSize         int
	
	// Cleanup configuration
	cleanupInterval time.Duration
	maxIdleTime     time.Duration
	
	// Connection timestamps for cleanup
	lastAccess map[string]time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(requestsPerSecond float64, burstSize int) *RateLimiter {
	rl := &RateLimiter{
		limiters:          make(map[string]*rate.Limiter),
		lastAccess:        make(map[string]time.Time),
		requestsPerSecond: rate.Limit(requestsPerSecond),
		burstSize:         burstSize,
		cleanupInterval:   5 * time.Minute,
		maxIdleTime:       10 * time.Minute,
	}
	
	// Start cleanup goroutine
	go rl.cleanup()
	
	return rl
}

// Allow checks if a request is allowed for the given connection
func (rl *RateLimiter) Allow(connID string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	// Get or create limiter for this connection
	limiter, exists := rl.limiters[connID]
	if !exists {
		limiter = rate.NewLimiter(rl.requestsPerSecond, rl.burstSize)
		rl.limiters[connID] = limiter
	}
	
	// Update last access time
	rl.lastAccess[connID] = time.Now()
	
	return limiter.Allow()
}

// RemoveConnection removes rate limiting data for a connection
func (rl *RateLimiter) RemoveConnection(connID string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	delete(rl.limiters, connID)
	delete(rl.lastAccess, connID)
}

// cleanup removes inactive limiters
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.cleanupInterval)
	defer ticker.Stop()
	
	for range ticker.C {
		rl.cleanupInactive()
	}
}

// cleanupInactive removes limiters for inactive connections
func (rl *RateLimiter) cleanupInactive() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	now := time.Now()
	for connID, lastAccess := range rl.lastAccess {
		if now.Sub(lastAccess) > rl.maxIdleTime {
			delete(rl.limiters, connID)
			delete(rl.lastAccess, connID)
		}
	}
}

// GetStats returns rate limiter statistics
func (rl *RateLimiter) GetStats() map[string]interface{} {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	
	return map[string]interface{}{
		"active_limiters":      len(rl.limiters),
		"requests_per_second":  float64(rl.requestsPerSecond),
		"burst_size":          rl.burstSize,
		"cleanup_interval":    rl.cleanupInterval.String(),
		"max_idle_time":       rl.maxIdleTime.String(),
	}
}

// ErrorHandler manages error handling for WebSocket connections
type ErrorHandler struct {
	logger        *log.Logger
	errorCounts   map[string]int
	mu            sync.RWMutex
	maxErrors     int
	resetInterval time.Duration
}

// NewErrorHandler creates a new error handler
func NewErrorHandler(logger *log.Logger, maxErrors int, resetInterval time.Duration) *ErrorHandler {
	if logger == nil {
		logger = log.Default()
	}
	
	eh := &ErrorHandler{
		logger:        logger,
		errorCounts:   make(map[string]int),
		maxErrors:     maxErrors,
		resetInterval: resetInterval,
	}
	
	// Start error count reset goroutine
	go eh.resetErrorCounts()
	
	return eh
}

// HandleError handles errors for a connection
func (eh *ErrorHandler) HandleError(conn *Connection, err error, context string) bool {
	eh.mu.Lock()
	defer eh.mu.Unlock()
	
	// Log the error
	eh.logger.Printf("[WS] Connection %s - Error in %s: %v", conn.ID, context, err)
	
	// Increment error count
	eh.errorCounts[conn.ID]++
	
	// Check if connection should be closed due to too many errors
	if eh.errorCounts[conn.ID] >= eh.maxErrors {
		eh.logger.Printf("[WS] Connection %s - Too many errors (%d), closing connection", 
			conn.ID, eh.errorCounts[conn.ID])
		return true // Should close connection
	}
	
	return false // Don't close connection
}

// RemoveConnection removes error tracking for a connection
func (eh *ErrorHandler) RemoveConnection(connID string) {
	eh.mu.Lock()
	defer eh.mu.Unlock()
	delete(eh.errorCounts, connID)
}

// resetErrorCounts periodically resets error counts
func (eh *ErrorHandler) resetErrorCounts() {
	ticker := time.NewTicker(eh.resetInterval)
	defer ticker.Stop()
	
	for range ticker.C {
		eh.mu.Lock()
		eh.errorCounts = make(map[string]int)
		eh.mu.Unlock()
	}
}

// GetErrorStats returns error statistics
func (eh *ErrorHandler) GetErrorStats() map[string]interface{} {
	eh.mu.RLock()
	defer eh.mu.RUnlock()
	
	totalErrors := 0
	for _, count := range eh.errorCounts {
		totalErrors += count
	}
	
	return map[string]interface{}{
		"connections_with_errors": len(eh.errorCounts),
		"total_errors":           totalErrors,
		"max_errors":             eh.maxErrors,
		"reset_interval":         eh.resetInterval.String(),
	}
}

// MiddlewareManager manages all WebSocket middleware
type MiddlewareManager struct {
	logging     *LoggingMiddleware
	rateLimiter *RateLimiter
	errorHandler *ErrorHandler
}

// NewMiddlewareManager creates a new middleware manager
func NewMiddlewareManager() *MiddlewareManager {
	return &MiddlewareManager{
		logging:      NewLoggingMiddleware(nil),
		rateLimiter:  NewRateLimiter(10.0, 20), // 10 requests per second, burst of 20
		errorHandler: NewErrorHandler(nil, 10, 5*time.Minute), // Max 10 errors, reset every 5 minutes
	}
}

// GetLoggingMiddleware returns the logging middleware
func (mm *MiddlewareManager) GetLoggingMiddleware() *LoggingMiddleware {
	return mm.logging
}

// GetRateLimiter returns the rate limiter
func (mm *MiddlewareManager) GetRateLimiter() *RateLimiter {
	return mm.rateLimiter
}

// GetErrorHandler returns the error handler
func (mm *MiddlewareManager) GetErrorHandler() *ErrorHandler {
	return mm.errorHandler
}

// ProcessMessage processes a message through all middleware
func (mm *MiddlewareManager) ProcessMessage(conn *Connection, messageBytes []byte) bool {
	// Check rate limiting
	if !mm.rateLimiter.Allow(conn.ID) {
		mm.logging.LogError(conn, ErrRateLimited, "rate_limiting")
		conn.sendError("rate_limited", "Too many requests")
		return false
	}
	
	// Log incoming message
	mm.logging.LogMessage(conn, "incoming", MessageType("unknown"), len(messageBytes))
	
	return true
}

// HandleConnectionError handles connection errors through middleware
func (mm *MiddlewareManager) HandleConnectionError(conn *Connection, err error, context string) bool {
	// Log the error
	mm.logging.LogError(conn, err, context)
	
	// Handle the error and check if connection should be closed
	shouldClose := mm.errorHandler.HandleError(conn, err, context)
	
	if shouldClose {
		// Clean up middleware data
		mm.rateLimiter.RemoveConnection(conn.ID)
		mm.errorHandler.RemoveConnection(conn.ID)
	}
	
	return shouldClose
}

// CleanupConnection cleans up middleware data for a connection
func (mm *MiddlewareManager) CleanupConnection(conn *Connection) {
	mm.rateLimiter.RemoveConnection(conn.ID)
	mm.errorHandler.RemoveConnection(conn.ID)
	mm.logging.LogConnection(conn, "cleanup", map[string]interface{}{
		"duration": time.Since(conn.ConnectedAt).String(),
	})
}

// GetMiddlewareStats returns statistics from all middleware
func (mm *MiddlewareManager) GetMiddlewareStats() map[string]interface{} {
	return map[string]interface{}{
		"rate_limiter": mm.rateLimiter.GetStats(),
		"error_handler": mm.errorHandler.GetErrorStats(),
	}
}

// WebSocketMiddleware provides HTTP middleware for WebSocket endpoints
func WebSocketMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Set CORS headers for WebSocket
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, Upgrade, Connection, Sec-WebSocket-Key, Sec-WebSocket-Version, Sec-WebSocket-Protocol")
		c.Header("Access-Control-Allow-Credentials", "true")
		
		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	})
}