package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// SetupSwaggerRoutes configures swagger documentation routes
func SetupSwaggerRoutes(router *gin.Engine) {
	// Serve swagger.json file with correct content type
	router.GET("/swagger.json", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		c.File("docs/swagger.json")
	})

	// Serve swagger.yaml file with correct content type
	router.GET("/swagger.yaml", func(c *gin.Context) {
		c.Header("Content-Type", "application/x-yaml")
		c.File("docs/swagger.yaml")
	})

	// Redirect root to swagger UI
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})
}
