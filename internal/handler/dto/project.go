package dto

import (
	"time"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/google/uuid"
)

// Project request DTOs
type ProjectCreateRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=255" example:"My Project"`
	Description string `json:"description" binding:"max=1000" example:"Project description"`
	RepoURL     string `json:"repo_url" binding:"required,url,max=500" example:"https://github.com/user/repo"`
}

type ProjectUpdateRequest struct {
	Name        *string `json:"name,omitempty" binding:"omitempty,min=1,max=255" example:"Updated Project Name"`
	Description *string `json:"description,omitempty" binding:"omitempty,max=1000" example:"Updated description"`
	RepoURL     *string `json:"repo_url,omitempty" binding:"omitempty,url,max=500" example:"https://github.com/user/updated-repo"`
}

// Project response DTOs
type ProjectResponse struct {
	ID          uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Name        string    `json:"name" example:"My Project"`
	Description string    `json:"description" example:"Project description"`
	RepoURL     string    `json:"repo_url" example:"https://github.com/user/repo"`
	CreatedAt   time.Time `json:"created_at" example:"2024-01-15T10:30:00Z"`
	UpdatedAt   time.Time `json:"updated_at" example:"2024-01-15T10:30:00Z"`
}

type ProjectWithTasksResponse struct {
	ProjectResponse
	Tasks []TaskResponse `json:"tasks"`
}

type ProjectListResponse struct {
	Projects []ProjectResponse `json:"projects"`
	Total    int               `json:"total"`
}

// Helper functions to convert between entity and DTO
func (p *ProjectResponse) FromEntity(project *entity.Project) {
	p.ID = project.ID
	p.Name = project.Name
	p.Description = project.Description
	p.RepoURL = project.RepoURL
	p.CreatedAt = project.CreatedAt
	p.UpdatedAt = project.UpdatedAt
}

func (p *ProjectWithTasksResponse) FromEntity(project *entity.Project) {
	p.ProjectResponse.FromEntity(project)
	p.Tasks = make([]TaskResponse, len(project.Tasks))
	for i, task := range project.Tasks {
		p.Tasks[i].FromEntity(&task)
	}
}

func ProjectResponseFromEntity(project *entity.Project) ProjectResponse {
	var resp ProjectResponse
	resp.FromEntity(project)
	return resp
}

func ProjectListResponseFromEntities(projects []*entity.Project) ProjectListResponse {
	responses := make([]ProjectResponse, len(projects))
	for i, project := range projects {
		responses[i] = ProjectResponseFromEntity(project)
	}
	return ProjectListResponse{
		Projects: responses,
		Total:    len(projects),
	}
}