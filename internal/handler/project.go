package handler

import (
	"net/http"

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
	projects, err := h.projectUsecase.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err, http.StatusInternalServerError, "Failed to fetch projects"))
		return
	}

	response := dto.ProjectListResponseFromEntities(projects)
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