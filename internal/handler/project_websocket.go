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

	// Git-related fields
	if req.RepositoryURL != nil && *req.RepositoryURL != originalProject.RepositoryURL {
		usecaseReq.RepositoryURL = *req.RepositoryURL
		changes["repository_url"] = map[string]interface{}{
			"old": originalProject.RepositoryURL,
			"new": *req.RepositoryURL,
		}
	}
	if req.MainBranch != nil && *req.MainBranch != originalProject.MainBranch {
		usecaseReq.MainBranch = *req.MainBranch
		changes["main_branch"] = map[string]interface{}{
			"old": originalProject.MainBranch,
			"new": *req.MainBranch,
		}
	}
	if req.WorktreeBasePath != nil && *req.WorktreeBasePath != originalProject.WorktreeBasePath {
		usecaseReq.WorktreeBasePath = *req.WorktreeBasePath
		changes["worktree_base_path"] = map[string]interface{}{
			"old": originalProject.WorktreeBasePath,
			"new": *req.WorktreeBasePath,
		}
	}
	if req.GitAuthMethod != nil && *req.GitAuthMethod != originalProject.GitAuthMethod {
		usecaseReq.GitAuthMethod = *req.GitAuthMethod
		changes["git_auth_method"] = map[string]interface{}{
			"old": originalProject.GitAuthMethod,
			"new": *req.GitAuthMethod,
		}
	}
	if req.GitEnabled != nil && *req.GitEnabled != originalProject.GitEnabled {
		usecaseReq.GitEnabled = *req.GitEnabled
		changes["git_enabled"] = map[string]interface{}{
			"old": originalProject.GitEnabled,
			"new": *req.GitEnabled,
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
