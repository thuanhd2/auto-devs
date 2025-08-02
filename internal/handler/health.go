package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
}

func SetupHealthRoutes(router *gin.Engine) {
	v1 := router.Group("/api/v1")
	{
		v1.GET("/health", healthCheck)
	}
}

func healthCheck(c *gin.Context) {
	response := HealthResponse{
		Status:    "ok",
		Timestamp: time.Now(),
		Version:   "1.0.0",
	}

	c.JSON(http.StatusOK, response)
}