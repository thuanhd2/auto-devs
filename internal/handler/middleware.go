package handler

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/auto-devs/auto-devs/internal/handler/dto"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"golang.org/x/time/rate"
)

// CORSMiddleware configures CORS settings
func CORSMiddleware() gin.HandlerFunc {
	config := cors.Config{
		AllowOrigins: []string{
			"http://localhost:9000",
			"http://localhost:9001",
			"http://localhost:9002",
			"http://localhost:9003",
			"http://localhost:9004",
			"http://localhost:9005",
			"http://localhost:9006",
			"http://localhost:9007",
			"http://localhost:9008",
			"http://localhost:9009",
			"http://localhost:9010",
			"*", // Allow all origins for WebSocket
		}, // React dev servers
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With", "Upgrade", "Connection", "Sec-WebSocket-Key", "Sec-WebSocket-Version", "Sec-WebSocket-Protocol"},
		ExposeHeaders:    []string{"Content-Length", "X-Total-Count"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	return cors.New(config)
}

// RequestLoggingMiddleware logs API requests and responses
func RequestLoggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// Skip logging for WebSocket endpoints to reduce noise
		if param.Path == "/ws" {
			return ""
		}

		// Default logging format
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}

// ErrorHandlingMiddleware handles panics and errors
func ErrorHandlingMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		// Skip error handling for WebSocket endpoints
		if c.Request.URL.Path == "/ws" {
			c.Next()
			return
		}

		if err, ok := recovered.(string); ok {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Error:   err,
				Message: "Internal server error",
				Code:    http.StatusInternalServerError,
			})
		}
		c.AbortWithStatus(http.StatusInternalServerError)
	})
}

// ValidationErrorMiddleware handles validation errors
func ValidationErrorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip validation for WebSocket endpoints
		if c.Request.URL.Path == "/ws" {
			c.Next()
			return
		}

		c.Next()

		// Check if there are any errors from validation
		if len(c.Errors) > 0 {
			err := c.Errors.Last()

			// If it's a validation error, format it properly
			if validationErrors, ok := err.Err.(validator.ValidationErrors); ok {
				details := make(map[string]string)
				for _, fieldErr := range validationErrors {
					details[fieldErr.Field()] = getValidationErrorMessage(fieldErr)
				}

				c.JSON(http.StatusBadRequest, dto.NewValidationErrorResponse(details))
				c.Abort()
				return
			}
		}
	}
}

// getValidationErrorMessage returns a user-friendly validation error message
func getValidationErrorMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "This field is required"
	case "min":
		return "This field must be at least " + fe.Param() + " characters long"
	case "max":
		return "This field must be at most " + fe.Param() + " characters long"
	case "email":
		return "This field must be a valid email address"
	case "url":
		return "This field must be a valid URL"
	case "uuid":
		return "This field must be a valid UUID"
	case "oneof":
		return "This field must be one of: " + fe.Param()
	default:
		return "This field is invalid"
	}
}

// RateLimitMiddleware implements basic rate limiting
func RateLimitMiddleware() gin.HandlerFunc {
	// Create a rate limiter that allows 100 requests per minute
	limiter := rate.NewLimiter(rate.Every(time.Minute/100), 100)

	return func(c *gin.Context) {
		// Skip rate limiting for WebSocket endpoints
		if c.Request.URL.Path == "/ws" {
			c.Next()
			return
		}

		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, dto.ErrorResponse{
				Error:   "Rate limit exceeded",
				Message: "Too many requests, please try again later",
				Code:    http.StatusTooManyRequests,
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// SecurityHeadersMiddleware adds security headers
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip security headers for WebSocket endpoints to avoid conflicts
		if c.Request.URL.Path == "/ws" {
			c.Next()
			return
		}

		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Next()
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

		// Log WebSocket connection attempt
		log.Printf("WebSocket middleware: %s %s from %s", c.Request.Method, c.Request.URL.Path, c.ClientIP())

		c.Next()
	})
}
