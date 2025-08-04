package handler

import (
	"net/http"
	"strconv"

	"github.com/auto-devs/auto-devs/internal/handler/dto"
	"github.com/auto-devs/auto-devs/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ProjectHandler struct {
	projectUsecase usecase.ProjectUsecase
}

func NewProjectHandler(projectUsecase usecase.ProjectUsecase) *ProjectHandler {
	return &ProjectHandler{
		projectUsecase: projectUsecase,
	}
}

// CreateProject godoc
// @Summary Create a new project
// @Description Create a new project with the provided details
// @Tags projects
// @Accept json
// @Produce json
// @Param project body dto.ProjectCreateRequest true "Project creation data"
// @Success 201 {object} dto.ProjectResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/projects [post]
func (h *ProjectHandler) CreateProject(c *gin.Context) {
	var req dto.ProjectCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err, http.StatusBadRequest, "Invalid request data"))
		return
	}

	usecaseReq := usecase.CreateProjectRequest{
		Name:        req.Name,
		Description: req.Description,
		RepoURL:     req.RepoURL,

		// Git-related fields
		RepositoryURL:    req.RepositoryURL,
		MainBranch:       req.MainBranch,
		WorktreeBasePath: req.WorktreeBasePath,
		GitAuthMethod:    req.GitAuthMethod,
		GitEnabled:       req.GitEnabled,
	}

	project, err := h.projectUsecase.Create(c.Request.Context(), usecaseReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err, http.StatusInternalServerError, "Failed to create project"))
		return
	}

	response := dto.ProjectResponseFromEntity(project)
	c.JSON(http.StatusCreated, response)
}

// GetProject godoc
// @Summary Get a project by ID
// @Description Get a single project by its ID
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Success 200 {object} dto.ProjectResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/projects/{id} [get]
func (h *ProjectHandler) GetProject(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err, http.StatusBadRequest, "Invalid project ID"))
		return
	}

	project, err := h.projectUsecase.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.NewErrorResponse(err, http.StatusNotFound, "Project not found"))
		return
	}

	response := dto.ProjectResponseFromEntity(project)
	c.JSON(http.StatusOK, response)
}

// GetProjectWithTasks godoc
// @Summary Get a project with its tasks
// @Description Get a single project by its ID including all associated tasks
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Success 200 {object} dto.ProjectWithTasksResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/projects/{id}/tasks [get]
func (h *ProjectHandler) GetProjectWithTasks(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err, http.StatusBadRequest, "Invalid project ID"))
		return
	}

	project, err := h.projectUsecase.GetWithTasks(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.NewErrorResponse(err, http.StatusNotFound, "Project not found"))
		return
	}

	var response dto.ProjectWithTasksResponse
	response.FromEntity(project)
	c.JSON(http.StatusOK, response)
}

// ListProjects godoc
// @Summary List all projects
// @Description Get a list of all projects
// @Tags projects
// @Accept json
// @Produce json
// @Success 200 {object} dto.ProjectListResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/projects [get]
func (h *ProjectHandler) ListProjects(c *gin.Context) {
	// Parse query parameters
	search := c.Query("search")
	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortOrder := c.DefaultQuery("sort_order", "desc")
	page := 1
	pageSize := 10

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if sizeStr := c.Query("page_size"); sizeStr != "" {
		if s, err := strconv.Atoi(sizeStr); err == nil && s > 0 && s <= 100 {
			pageSize = s
		}
	}

	var archived *bool
	if archivedStr := c.Query("archived"); archivedStr != "" {
		if archVal, err := strconv.ParseBool(archivedStr); err == nil {
			archived = &archVal
		}
	}

	params := usecase.GetProjectsParams{
		Search:    search,
		SortBy:    sortBy,
		SortOrder: sortOrder,
		Page:      page,
		PageSize:  pageSize,
		Archived:  archived,
	}

	result, err := h.projectUsecase.GetAll(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err, http.StatusInternalServerError, "Failed to fetch projects"))
		return
	}

	response := dto.ProjectListResponseFromResult(result)
	c.JSON(http.StatusOK, response)
}

// UpdateProject godoc
// @Summary Update a project
// @Description Update a project with the provided details
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param project body dto.ProjectUpdateRequest true "Project update data"
// @Success 200 {object} dto.ProjectResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/projects/{id} [put]
func (h *ProjectHandler) UpdateProject(c *gin.Context) {
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

	usecaseReq := usecase.UpdateProjectRequest{}
	if req.Name != nil {
		usecaseReq.Name = *req.Name
	}
	if req.Description != nil {
		usecaseReq.Description = *req.Description
	}
	if req.RepoURL != nil {
		usecaseReq.RepoURL = *req.RepoURL
	}

	// Git-related fields
	if req.RepositoryURL != nil {
		usecaseReq.RepositoryURL = *req.RepositoryURL
	}
	if req.MainBranch != nil {
		usecaseReq.MainBranch = *req.MainBranch
	}
	if req.WorktreeBasePath != nil {
		usecaseReq.WorktreeBasePath = *req.WorktreeBasePath
	}
	if req.GitAuthMethod != nil {
		usecaseReq.GitAuthMethod = *req.GitAuthMethod
	}
	if req.GitEnabled != nil {
		usecaseReq.GitEnabled = *req.GitEnabled
	}

	project, err := h.projectUsecase.Update(c.Request.Context(), id, usecaseReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err, http.StatusInternalServerError, "Failed to update project"))
		return
	}

	response := dto.ProjectResponseFromEntity(project)
	c.JSON(http.StatusOK, response)
}

// DeleteProject godoc
// @Summary Delete a project
// @Description Delete a project by its ID
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Success 204
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/projects/{id} [delete]
func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err, http.StatusBadRequest, "Invalid project ID"))
		return
	}

	err = h.projectUsecase.Delete(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err, http.StatusInternalServerError, "Failed to delete project"))
		return
	}

	c.Status(http.StatusNoContent)
}

// GetProjectStatistics godoc
// @Summary Get project statistics
// @Description Get task statistics and completion data for a project
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Success 200 {object} dto.ProjectStatisticsResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/projects/{id}/statistics [get]
func (h *ProjectHandler) GetProjectStatistics(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err, http.StatusBadRequest, "Invalid project ID"))
		return
	}

	stats, err := h.projectUsecase.GetStatistics(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.NewErrorResponse(err, http.StatusNotFound, "Project not found or failed to get statistics"))
		return
	}

	response := dto.ProjectStatisticsResponseFromUsecase(stats)
	c.JSON(http.StatusOK, response)
}

// ArchiveProject godoc
// @Summary Archive a project
// @Description Archive a project (soft delete)
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Success 204
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/projects/{id}/archive [post]
func (h *ProjectHandler) ArchiveProject(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err, http.StatusBadRequest, "Invalid project ID"))
		return
	}

	err = h.projectUsecase.Archive(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err, http.StatusInternalServerError, "Failed to archive project"))
		return
	}

	c.Status(http.StatusNoContent)
}

// RestoreProject godoc
// @Summary Restore an archived project
// @Description Restore an archived project (undelete)
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Success 204
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/projects/{id}/restore [post]
func (h *ProjectHandler) RestoreProject(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err, http.StatusBadRequest, "Invalid project ID"))
		return
	}

	err = h.projectUsecase.Restore(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err, http.StatusInternalServerError, "Failed to restore project"))
		return
	}

	c.Status(http.StatusNoContent)
}

// GetProjectSettings godoc
// @Summary Get project settings
// @Description Get configuration settings for a project
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Success 200 {object} dto.ProjectSettingsResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/projects/{id}/settings [get]
func (h *ProjectHandler) GetProjectSettings(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err, http.StatusBadRequest, "Invalid project ID"))
		return
	}

	settings, err := h.projectUsecase.GetSettings(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.NewErrorResponse(err, http.StatusNotFound, "Project settings not found"))
		return
	}

	response := dto.ProjectSettingsResponseFromEntity(settings)
	c.JSON(http.StatusOK, response)
}

// UpdateProjectSettings godoc
// @Summary Update project settings
// @Description Update configuration settings for a project
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param settings body dto.ProjectSettingsUpdateRequest true "Project settings data"
// @Success 200 {object} dto.ProjectSettingsResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/projects/{id}/settings [put]
func (h *ProjectHandler) UpdateProjectSettings(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err, http.StatusBadRequest, "Invalid project ID"))
		return
	}

	var req dto.ProjectSettingsUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err, http.StatusBadRequest, "Invalid request data"))
		return
	}

	settings := req.ToEntity()
	settings.ProjectID = id

	updatedSettings, err := h.projectUsecase.UpdateSettings(c.Request.Context(), id, settings)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err, http.StatusInternalServerError, "Failed to update project settings"))
		return
	}

	response := dto.ProjectSettingsResponseFromEntity(updatedSettings)
	c.JSON(http.StatusOK, response)
}

// ValidateGitProject godoc
// @Summary Validate Git project configuration
// @Description Validate Git configuration for a project
// @Tags projects
// @Accept json
// @Produce json
// @Param request body dto.GitProjectValidationRequest true "Git validation request"
// @Success 200 {object} dto.GitProjectValidationResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/projects/validate-git [post]
func (h *ProjectHandler) ValidateGitProject(c *gin.Context) {
	var req dto.GitProjectValidationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err, http.StatusBadRequest, "Invalid request data"))
		return
	}

	// TODO: Implement Git validation logic
	// For now, return a basic validation response
	response := dto.GitProjectValidationResponse{
		Valid:   true,
		Message: "Git configuration is valid",
	}

	c.JSON(http.StatusOK, response)
}

// GetGitProjectStatus godoc
// @Summary Get Git project status
// @Description Get the current Git status for a project
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Success 200 {object} dto.GitProjectStatusResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/projects/{id}/git-status [get]
func (h *ProjectHandler) GetGitProjectStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err, http.StatusBadRequest, "Invalid project ID"))
		return
	}

	project, err := h.projectUsecase.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.NewErrorResponse(err, http.StatusNotFound, "Project not found"))
		return
	}

	// TODO: Implement Git status logic
	// For now, return a basic status response
	response := dto.GitProjectStatusResponse{
		GitEnabled:      project.GitEnabled,
		WorktreeExists:  false,
		RepositoryValid: false,
		Status:          "Git status not implemented yet",
	}

	c.JSON(http.StatusOK, response)
}

// TestGitConnection godoc
// @Summary Test Git connection
// @Description Test the Git connection for a project
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/projects/{id}/test-git-connection [post]
func (h *ProjectHandler) TestGitConnection(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err, http.StatusBadRequest, "Invalid project ID"))
		return
	}

	project, err := h.projectUsecase.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.NewErrorResponse(err, http.StatusNotFound, "Project not found"))
		return
	}

	// TODO: Implement Git connection test logic
	// For now, return a basic response
	response := map[string]interface{}{
		"message":    "Git connection test not implemented yet",
		"project_id": project.ID,
	}

	c.JSON(http.StatusOK, response)
}

// SetupGitProject godoc
// @Summary Setup Git project
// @Description Setup Git integration for a project
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/projects/{id}/setup-git [post]
func (h *ProjectHandler) SetupGitProject(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err, http.StatusBadRequest, "Invalid project ID"))
		return
	}

	project, err := h.projectUsecase.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.NewErrorResponse(err, http.StatusNotFound, "Project not found"))
		return
	}

	// TODO: Implement Git setup logic
	// For now, return a basic response
	response := map[string]interface{}{
		"message":    "Git setup not implemented yet",
		"project_id": project.ID,
	}

	c.JSON(http.StatusOK, response)
}
