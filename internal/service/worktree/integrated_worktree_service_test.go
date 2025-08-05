package worktree

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/auto-devs/auto-devs/config"
	"github.com/auto-devs/auto-devs/internal/service/git"
	"github.com/google/uuid"
)

func TestNewIntegratedWorktreeService(t *testing.T) {
	// Test with default config
	config := &IntegratedConfig{
		Worktree: &config.WorktreeConfig{
			BaseDirectory: "/tmp/test-integrated-worktrees",
			MaxPathLength: 1024,
			EnableLogging: false,
		},
		Git: &git.ManagerConfig{
			WorkingDir:    "/tmp/test-integrated-worktrees",
			EnableLogging: false,
		},
	}

	service, err := NewIntegratedWorktreeService(config)
	if err != nil {
		t.Fatalf("Failed to create integrated worktree service: %v", err)
	}
	if service == nil {
		t.Fatal("Service should not be nil")
	}

	// Clean up
	os.RemoveAll("/tmp/test-integrated-worktrees")
}

func TestCreateTaskWorktree(t *testing.T) {
	// get absolute path of this project
	projectPath, err := os.Getwd()
	projectRootPath := filepath.Join(projectPath, "../../../")
	if err != nil {
		t.Fatalf("Failed to get project path: %v", err)
	}

	baseWorktreeDir := filepath.Join(projectRootPath, "bin")
	projectWorkDir := projectRootPath
	config := &IntegratedConfig{
		Worktree: &config.WorktreeConfig{
			BaseDirectory: baseWorktreeDir,
			MaxPathLength: 1024,
			EnableLogging: false,
		},
		Git: &git.ManagerConfig{
			WorkingDir:    projectWorkDir,
			EnableLogging: false,
		},
	}

	service, err := NewIntegratedWorktreeService(config)
	if err != nil {
		t.Fatalf("Failed to create integrated worktree service: %v", err)
	}

	ctx := context.Background()

	dummyTaskID := fmt.Sprintf("task-%s", uuid.New().String())
	// Test creating task worktree
	request := &CreateTaskWorktreeRequest{
		ProjectID:         "project-123",
		TaskID:            dummyTaskID,
		TaskTitle:         "Test Task",
		ProjectWorkDir:    projectWorkDir,
		ProjectMainBranch: "main",
	}

	info, err := service.CreateTaskWorktree(ctx, request)
	if err != nil {
		t.Fatalf("Failed to create task worktree: %v", err)
	}

	if info == nil {
		t.Fatal("Task worktree info should not be nil")
	}

	if info.ProjectID != request.ProjectID {
		t.Errorf("Expected project ID %s, got %s", request.ProjectID, info.ProjectID)
	}

	if info.TaskID != request.TaskID {
		t.Errorf("Expected task ID %s, got %s", request.TaskID, info.TaskID)
	}

	if info.TaskTitle != request.TaskTitle {
		t.Errorf("Expected task title %s, got %s", request.TaskTitle, info.TaskTitle)
	}

	if info.WorktreePath == "" {
		t.Error("Worktree path should not be empty")
	}

	if info.BranchName == "" {
		t.Error("Branch name should not be empty")
	}

	// Clean up
	service.CleanupTaskWorktree(ctx, &CleanupTaskWorktreeRequest{
		ProjectID: request.ProjectID,
		TaskID:    request.TaskID,
	})
}

func TestGetTaskWorktreeInfo(t *testing.T) {
	projectPath, err := os.Getwd()
	projectRootPath := filepath.Join(projectPath, "../../../")
	if err != nil {
		t.Fatalf("Failed to get project path: %v", err)
	}

	baseWorktreeDir := filepath.Join(projectRootPath, "bin")
	config := &IntegratedConfig{
		Worktree: &config.WorktreeConfig{
			BaseDirectory: baseWorktreeDir,
			MaxPathLength: 1024,
			EnableLogging: false,
		},
		Git: &git.ManagerConfig{
			WorkingDir:    projectRootPath,
			EnableLogging: false,
		},
	}

	service, err := NewIntegratedWorktreeService(config)
	if err != nil {
		t.Fatalf("Failed to create integrated worktree service: %v", err)
	}

	ctx := context.Background()

	dummyTaskID := fmt.Sprintf("task-%s", uuid.New().String())
	// Create a task worktree first
	createRequest := &CreateTaskWorktreeRequest{
		ProjectID:         "project-123",
		TaskID:            dummyTaskID,
		TaskTitle:         "Test Task",
		ProjectWorkDir:    projectRootPath,
		ProjectMainBranch: "main",
	}

	createdInfo, err := service.CreateTaskWorktree(ctx, createRequest)
	if err != nil {
		t.Fatalf("Failed to create task worktree: %v", err)
	}

	// Get task worktree info
	info, err := service.GetTaskWorktreeInfo(ctx, "project-123", dummyTaskID)
	if err != nil {
		t.Fatalf("Failed to get task worktree info: %v", err)
	}

	if info == nil {
		t.Fatal("Task worktree info should not be nil")
	}

	if info.ProjectID != "project-123" {
		t.Errorf("Expected project ID project-123, got %s", info.ProjectID)
	}

	if info.TaskID != dummyTaskID {
		t.Errorf("Expected task ID %s, got %s", dummyTaskID, info.TaskID)
	}

	if info.WorktreePath != createdInfo.WorktreePath {
		t.Errorf("Expected worktree path %s, got %s", createdInfo.WorktreePath, info.WorktreePath)
	}

	// Test getting info for non-existent worktree
	_, err = service.GetTaskWorktreeInfo(ctx, "project-123", "non-existent-task")
	if err == nil {
		t.Error("Expected error when getting info for non-existent worktree")
	}

	// Clean up
	service.CleanupTaskWorktree(ctx, &CleanupTaskWorktreeRequest{
		ProjectID: "project-123",
		TaskID:    dummyTaskID,
	})
}

func TestListProjectWorktrees(t *testing.T) {
	projectPath, err := os.Getwd()
	projectRootPath := filepath.Join(projectPath, "../../../")
	if err != nil {
		t.Fatalf("Failed to get project path: %v", err)
	}

	baseWorktreeDir := filepath.Join(projectRootPath, "bin")
	config := &IntegratedConfig{
		Worktree: &config.WorktreeConfig{
			BaseDirectory: baseWorktreeDir,
			MaxPathLength: 1024,
			EnableLogging: false,
		},
		Git: &git.ManagerConfig{
			WorkingDir:    projectRootPath,
			EnableLogging: false,
		},
	}

	service, err := NewIntegratedWorktreeService(config)
	if err != nil {
		t.Fatalf("Failed to create integrated worktree service: %v", err)
	}

	ctx := context.Background()

	dummyTaskID1 := fmt.Sprintf("task-%s", uuid.New().String())
	dummyTaskID2 := fmt.Sprintf("task-%s", uuid.New().String())
	// Create multiple task worktrees
	createRequest1 := &CreateTaskWorktreeRequest{
		ProjectID:         "project-123",
		TaskID:            dummyTaskID1,
		TaskTitle:         "Test Task 1",
		ProjectWorkDir:    projectRootPath,
		ProjectMainBranch: "main",
	}

	createRequest2 := &CreateTaskWorktreeRequest{
		ProjectID:         "project-123",
		TaskID:            dummyTaskID2,
		TaskTitle:         "Test Task 2",
		ProjectWorkDir:    projectRootPath,
		ProjectMainBranch: "main",
	}

	_, err = service.CreateTaskWorktree(ctx, createRequest1)
	if err != nil {
		t.Fatalf("Failed to create task worktree 1: %v", err)
	}

	_, err = service.CreateTaskWorktree(ctx, createRequest2)
	if err != nil {
		t.Fatalf("Failed to create task worktree 2: %v", err)
	}

	// List project worktrees
	worktrees, err := service.ListProjectWorktrees(ctx, "project-123")
	if err != nil {
		t.Fatalf("Failed to list project worktrees: %v", err)
	}

	if len(worktrees) < 2 {
		t.Errorf("Expected at least 2 worktrees, got %d", len(worktrees))
	}

	// Test listing worktrees for non-existent project
	worktrees, err = service.ListProjectWorktrees(ctx, "non-existent-project")
	if err != nil {
		t.Fatalf("Failed to list worktrees for non-existent project: %v", err)
	}

	if len(worktrees) != 0 {
		t.Errorf("Expected 0 worktrees for non-existent project, got %d", len(worktrees))
	}

	// Clean up
	service.CleanupTaskWorktree(ctx, &CleanupTaskWorktreeRequest{
		ProjectID: "project-123",
		TaskID:    dummyTaskID1,
	})
	service.CleanupTaskWorktree(ctx, &CleanupTaskWorktreeRequest{
		ProjectID: "project-123",
		TaskID:    dummyTaskID2,
	})
}
