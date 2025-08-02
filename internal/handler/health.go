package handler

import (
	"net/http"
	"time"

	"github.com/auto-devs/auto-devs/pkg/database"
	"github.com/gin-gonic/gin"
)

type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Version   string            `json:"version"`
	Database  DatabaseHealth    `json:"database"`
}

type DatabaseHealth struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

func SetupHealthRoutes(router *gin.Engine, db *database.DB) {
	v1 := router.Group("/api/v1")
	{
		v1.GET("/health", healthCheck(db))
	}
}

func healthCheck(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		dbHealth := DatabaseHealth{
			Status: "ok",
		}

		// Check database connectivity
		if err := db.HealthCheck(); err != nil {
			dbHealth.Status = "error"
			dbHealth.Error = err.Error()
		}

		overallStatus := "ok"
		if dbHealth.Status == "error" {
			overallStatus = "degraded"
		}

		response := HealthResponse{
			Status:    overallStatus,
			Timestamp: time.Now(),
			Version:   "1.0.0",
			Database:  dbHealth,
		}

		statusCode := http.StatusOK
		if overallStatus == "degraded" {
			statusCode = http.StatusServiceUnavailable
		}

		c.JSON(statusCode, response)
	}
}