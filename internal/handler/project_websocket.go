package handler

import (
	"log"
	"net/http"

	"github.com/auto-devs/auto-devs/internal/handler/dto"
	"github.com/auto-devs/auto-devs/internal/usecase"
	"github.com/auto-devs/auto-devs/internal/websocket"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ProjectHandlerWithWebSocket extends the basic project handler with WebSocket notifications
type ProjectHandlerWithWebSocket struct {
	*ProjectHandler
	wsService *websocket.Service
}

// NewProjectHandlerWithWebSocket creates a new project handler with WebSocket support
func NewProjectHandlerWithWebSocket(projectUsecase usecase.ProjectUsecase, wsService *websocket.Service) *ProjectHandlerWithWebSocket {
	return &ProjectHandlerWithWebSocket{
		ProjectHandler: NewProjectHandler(projectUsecase),
		wsService:      wsService,
	}
}

// UpdateProject updates a project and sends WebSocket notification
func (h *ProjectHandlerWithWebSocket) UpdateProject(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err, http.StatusBadRequest, "Invalid project ID"))
		return
	}

	var req dto.ProjectUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err, http.StatusBadRequest, "Invalid request data"))
		return
	}

	// Get the original project to track changes
	originalProject, err := h.projectUsecase.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.NewErrorResponse(err, http.StatusNotFound, "Project not found"))
		return
	}

	usecaseReq := usecase.UpdateProjectRequest{}
	changes := make(map[string]interface{})
	
	if req.Name != nil && *req.Name != originalProject.Name {
		usecaseReq.Name = *req.Name
		changes["name"] = map[string]interface{}{
			"old": originalProject.Name,
			"new": *req.Name,
		}
	}
	if req.Description != nil && *req.Description != originalProject.Description {
		usecaseReq.Description = *req.Description
		changes["description"] = map[string]interface{}{
			"old": originalProject.Description,
			"new": *req.Description,
		}
	}
	if req.RepoURL != nil && *req.RepoURL != originalProject.RepoURL {
		usecaseReq.RepoURL = *req.RepoURL
		changes["repo_url"] = map[string]interface{}{
			"old": originalProject.RepoURL,
			"new": *req.RepoURL,
		}
	}

	project, err := h.projectUsecase.Update(c.Request.Context(), id, usecaseReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err, http.StatusInternalServerError, "Failed to update project"))
		return
	}

	response := dto.ProjectResponseFromEntity(project)
	
	// Send WebSocket notification if there were changes
	if len(changes) > 0 {
		if err := h.wsService.NotifyProjectUpdated(project.ID, changes, response); err != nil {
			log.Printf("Failed to send WebSocket notification for project update: %v", err)
		}
	}

	c.JSON(http.StatusOK, response)
}