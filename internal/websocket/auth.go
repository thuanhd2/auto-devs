package websocket

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// AuthService defines the interface for authentication
type AuthService interface {
	ValidateToken(token string) (*UserInfo, error)
	ValidateWebSocketAuth(token string) (*UserInfo, error)
}

// UserInfo represents authenticated user information
type UserInfo struct {
	ID       string            `json:"id"`
	Username string            `json:"username"`
	Email    string            `json:"email"`
	Roles    []string          `json:"roles"`
	Metadata map[string]string `json:"metadata"`
}

// MockAuthService provides a mock authentication service for development
type MockAuthService struct{}

// NewMockAuthService creates a new mock authentication service
func NewMockAuthService() *MockAuthService {
	return &MockAuthService{}
}

// ValidateToken validates a JWT token (mock implementation)
func (m *MockAuthService) ValidateToken(token string) (*UserInfo, error) {
	// Mock implementation - in production, implement proper JWT validation
	if token == "" {
		return nil, ErrUnauthorized
	}
	
	// For development, accept any non-empty token
	return &UserInfo{
		ID:       "user-123",
		Username: "developer",
		Email:    "dev@example.com",
		Roles:    []string{"user"},
		Metadata: map[string]string{
			"authenticated_at": time.Now().Format(time.RFC3339),
		},
	}, nil
}

// ValidateWebSocketAuth validates authentication for WebSocket connections
func (m *MockAuthService) ValidateWebSocketAuth(token string) (*UserInfo, error) {
	return m.ValidateToken(token)
}

// AuthMiddleware provides authentication middleware for HTTP requests
func AuthMiddleware(authService AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip authentication for certain paths
		if shouldSkipAuth(c.Request.URL.Path) {
			c.Next()
			return
		}
		
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}
		
		// Parse Bearer token
		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}
		
		token := strings.TrimPrefix(authHeader, bearerPrefix)
		
		// Validate token
		userInfo, err := authService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}
		
		// Store user info in context
		c.Set("user", userInfo)
		c.Next()
	}
}

// WebSocketAuthMiddleware provides authentication for WebSocket connections
func WebSocketAuthMiddleware(authService AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from query parameter or header
		token := extractWebSocketToken(c.Request)
		
		if token != "" {
			// Validate token
			userInfo, err := authService.ValidateWebSocketAuth(token)
			if err == nil {
				// Store user info in context for WebSocket handler
				c.Set("user", userInfo)
			}
		}
		
		// Continue regardless - authentication can also happen via WebSocket messages
		c.Next()
	}
}

// extractWebSocketToken extracts authentication token from WebSocket request
func extractWebSocketToken(r *http.Request) string {
	// Try query parameter first
	if token := r.URL.Query().Get("token"); token != "" {
		return token
	}
	
	// Try Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		const bearerPrefix = "Bearer "
		if strings.HasPrefix(authHeader, bearerPrefix) {
			return strings.TrimPrefix(authHeader, bearerPrefix)
		}
	}
	
	// Try Sec-WebSocket-Protocol header (for token-based auth)
	protocols := r.Header.Get("Sec-WebSocket-Protocol")
	if protocols != "" {
		for _, protocol := range strings.Split(protocols, ",") {
			protocol = strings.TrimSpace(protocol)
			if strings.HasPrefix(protocol, "auth-") {
				return strings.TrimPrefix(protocol, "auth-")
			}
		}
	}
	
	return ""
}

// shouldSkipAuth checks if authentication should be skipped for the given path
func shouldSkipAuth(path string) bool {
	skipPaths := []string{
		"/health",
		"/healthz",
		"/ping",
		"/swagger",
		"/docs",
	}
	
	for _, skipPath := range skipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	
	return false
}

// GetUserFromContext extracts user information from Gin context
func GetUserFromContext(c *gin.Context) (*UserInfo, bool) {
	user, exists := c.Get("user")
	if !exists {
		return nil, false
	}
	
	userInfo, ok := user.(*UserInfo)
	return userInfo, ok
}

// GetUserFromGinContext extracts user information from Gin context
func GetUserFromGinContext(c *gin.Context) *UserInfo {
	if user, exists := GetUserFromContext(c); exists {
		return user
	}
	return nil
}

// AuthProcessor handles authentication-related WebSocket messages
type AuthProcessor struct {
	authService AuthService
	hub         *Hub
}

// NewAuthProcessor creates a new authentication processor
func NewAuthProcessor(authService AuthService, hub *Hub) *AuthProcessor {
	return &AuthProcessor{
		authService: authService,
		hub:         hub,
	}
}

// ProcessMessage processes authentication messages
func (p *AuthProcessor) ProcessMessage(conn *Connection, message *Message) error {
	switch message.Type {
	case AuthRequired:
		return p.handleAuthRequest(conn, message)
	default:
		return ErrProcessingFailed
	}
}

// handleAuthRequest processes authentication requests
func (p *AuthProcessor) handleAuthRequest(conn *Connection, message *Message) error {
	var authData AuthData
	if err := message.ParseData(&authData); err != nil {
		return err
	}
	
	// Validate token
	userInfo, err := p.authService.ValidateWebSocketAuth(authData.Token)
	if err != nil {
		// Send auth failed message
		failMessage, _ := NewMessage(AuthFailed, AuthData{
			Message: "Authentication failed",
		})
		return conn.SendMessage(failMessage)
	}
	
	// Associate connection with user
	p.hub.AssociateConnectionWithUser(conn, userInfo.ID)
	
	// Send auth success message
	successMessage, _ := NewMessage(AuthSuccess, AuthData{
		UserID:  userInfo.ID,
		Message: "Authentication successful",
	})
	
	return conn.SendMessage(successMessage)
}

// RequireAuth middleware that requires authentication
func RequireAuth(authService AuthService) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		user := GetUserFromGinContext(c)
		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}
		c.Next()
	})
}

// RequireRole middleware that requires specific roles
func RequireRole(roles ...string) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		user := GetUserFromGinContext(c)
		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}
		
		// Check if user has any of the required roles
		hasRole := false
		for _, requiredRole := range roles {
			for _, userRole := range user.Roles {
				if userRole == requiredRole {
					hasRole = true
					break
				}
			}
			if hasRole {
				break
			}
		}
		
		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}
		
		c.Next()
	})
}

// ContextKey type for context keys
type ContextKey string

const (
	// UserContextKey is the context key for user information
	UserContextKey ContextKey = "user"
)

// WithUser adds user information to context
func WithUser(ctx context.Context, user *UserInfo) context.Context {
	return context.WithValue(ctx, UserContextKey, user)
}

// UserFromContext extracts user information from context
func UserFromContext(ctx context.Context) (*UserInfo, bool) {
	user, ok := ctx.Value(UserContextKey).(*UserInfo)
	return user, ok
}