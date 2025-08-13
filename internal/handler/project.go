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
		Name:                req.Name,
		Description:         req.Description,
		WorktreeBasePath:    req.WorktreeBasePath,
		InitWorkspaceScript: req.InitWorkspaceScript,
		ExecutorType:        req.ExecutorType,
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
	if req.RepositoryURL != nil {
		usecaseReq.RepositoryURL = *req.RepositoryURL
	}
	if req.WorktreeBasePath != nil {
		usecaseReq.WorktreeBasePath = *req.WorktreeBasePath
	}
	if req.InitWorkspaceScript != nil {
		usecaseReq.InitWorkspaceScript = *req.InitWorkspaceScript
	}
	if req.ExecutorType != nil {
		usecaseReq.ExecutorType = *req.ExecutorType
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







// ReinitGitRepository godoc
// @Summary Reinitialize Git repository for a project
// @Description Reinitialize and reassign Git repository and GitHub repository URL for a project
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Success 200 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/projects/{id}/git/reinit [post]
func (h *ProjectHandler) ReinitGitRepository(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err, http.StatusBadRequest, "Invalid project ID"))
		return
	}

	err = h.projectUsecase.ReinitGitRepository(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err, http.StatusInternalServerError, "Failed to reinitialize Git repository"))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Git repository reinitialized successfully", nil))
}



// ListBranches godoc
// @Summary List Git branches for a project
// @Description Get all Git branches available in the project repository
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Success 200 {object} dto.ListBranchesResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/projects/{id}/branches [get]
func (h *ProjectHandler) ListBranches(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err, http.StatusBadRequest, "Invalid project ID"))
		return
	}

	branches, err := h.projectUsecase.ListBranches(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err, http.StatusInternalServerError, "Failed to list branches"))
		return
	}

	// Convert usecase.GitBranch to dto.GitBranchResponse
	branchResponses := make([]dto.GitBranchResponse, len(branches))
	for i, branch := range branches {
		branchResponses[i] = dto.GitBranchResponse{
			Name:        branch.Name,
			IsCurrent:   branch.IsCurrent,
			LastCommit:  branch.LastCommit,
			LastUpdated: branch.LastUpdated,
		}
	}

	response := dto.ListBranchesResponse{
		Branches: branchResponses,
		Total:    len(branchResponses),
	}
	c.JSON(http.StatusOK, response)
}
