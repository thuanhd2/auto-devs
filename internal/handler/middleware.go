package handler

import (
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
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:5173", "http://localhost:9000"}, // React dev servers
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length", "X-Total-Count"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	return cors.New(config)
}

// RequestLoggingMiddleware logs API requests and responses
func RequestLoggingMiddleware() gin.HandlerFunc {
	return gin.Logger()
}

// ErrorHandlingMiddleware handles panics and errors
func ErrorHandlingMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
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
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Next()
	}
}
