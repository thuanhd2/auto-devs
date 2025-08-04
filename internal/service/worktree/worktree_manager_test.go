package worktree

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewWorktreeManager(t *testing.T) {
	// Test with default config but using temp directory
	config := DefaultWorktreeConfig()
	config.BaseDirectory = "/tmp/test-worktrees-default"

	manager, err := NewWorktreeManager(config)
	if err != nil {
		t.Fatalf("Failed to create worktree manager with default config: %v", err)
	}
	if manager == nil {
		t.Fatal("Manager should not be nil")
	}

	// Test with custom config
	customConfig := &WorktreeConfig{
		BaseDirectory:   "/tmp/test-worktrees",
		MaxPathLength:   1024,
		MinDiskSpace:    50 * 1024 * 1024,
		CleanupInterval: 1 * time.Hour,
		EnableLogging:   true,
		LogLevel:        slog.LevelInfo,
	}

	manager, err = NewWorktreeManager(customConfig)
	if err != nil {
		t.Fatalf("Failed to create worktree manager with custom config: %v", err)
	}
	if manager == nil {
		t.Fatal("Manager should not be nil")
	}

	// Clean up test directories
	os.RemoveAll("/tmp/test-worktrees")
	os.RemoveAll("/tmp/test-worktrees-default")
}

func TestGenerateWorktreePath(t *testing.T) {
	manager, err := NewWorktreeManager(&WorktreeConfig{
		BaseDirectory: "/tmp/test-worktrees",
		MaxPathLength: 1024,
		EnableLogging: false,
	})
	if err != nil {
		t.Fatalf("Failed to create worktree manager: %v", err)
	}

	// Test valid inputs
	path, err := manager.GenerateWorktreePath("project-123", "task-456")
	if err != nil {
		t.Fatalf("Failed to generate worktree path: %v", err)
	}
	expectedPath := filepath.Join("/tmp/test-worktrees", "project-project-123", "task-task-456")
	if path != expectedPath {
		t.Errorf("Expected path %s, got %s", expectedPath, path)
	}

	// Test with special characters
	path, err = manager.GenerateWorktreePath("project/with/slashes", "task:with:colons")
	if err != nil {
		t.Fatalf("Failed to generate worktree path with special characters: %v", err)
	}
	expectedPath = filepath.Join("/tmp/test-worktrees", "project-project_with_slashes", "task-task_with_colons")
	if path != expectedPath {
		t.Errorf("Expected path %s, got %s", expectedPath, path)
	}

	// Test empty inputs
	_, err = manager.GenerateWorktreePath("", "task-456")
	if err == nil {
		t.Error("Expected error for empty project_id")
	}

	_, err = manager.GenerateWorktreePath("project-123", "")
	if err == nil {
		t.Error("Expected error for empty task_id")
	}

	// Clean up
	os.RemoveAll("/tmp/test-worktrees")
}

func TestCleanPathComponent(t *testing.T) {
	manager, err := NewWorktreeManager(&WorktreeConfig{
		BaseDirectory: "/tmp/test-worktrees",
		MaxPathLength: 1024,
		EnableLogging: false,
	})
	if err != nil {
		t.Fatalf("Failed to create worktree manager: %v", err)
	}

	tests := []struct {
		input    string
		expected string
	}{
		{"normal-name", "normal-name"},
		{"name with spaces", "name_with_spaces"},
		{"name/with/slashes", "name_with_slashes"},
		{"name:with:colons", "name_with_colons"},
		{"name*with*stars", "name_with_stars"},
		{"name?with?questions", "name_with_questions"},
		{"name\"with\"quotes", "name_with_quotes"},
		{"name<with>brackets", "name_with_brackets"},
		{"name|with|pipes", "name_with_pipes"},
		{"__name__with__underscores__", "name_with_underscores"},
		{"   name with leading/trailing spaces   ", "name_with_leading_trailing_spaces"},
	}

	for _, test := range tests {
		result := manager.cleanPathComponent(test.input)
		if result != test.expected {
			t.Errorf("cleanPathComponent(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

func TestCreateWorktree(t *testing.T) {
	manager, err := NewWorktreeManager(&WorktreeConfig{
		BaseDirectory: "/tmp/test-worktrees",
		MaxPathLength: 1024,
		EnableLogging: false,
	})
	if err != nil {
		t.Fatalf("Failed to create worktree manager: %v", err)
	}

	ctx := context.Background()

	// Test creating worktree
	worktreePath, err := manager.CreateWorktree(ctx, "project-123", "task-456")
	if err != nil {
		t.Fatalf("Failed to create worktree: %v", err)
	}

	// Verify worktree exists
	if !manager.WorktreeExists(worktreePath) {
		t.Error("Created worktree should exist")
	}

	// Test creating duplicate worktree
	_, err = manager.CreateWorktree(ctx, "project-123", "task-456")
	if err == nil {
		t.Error("Expected error when creating duplicate worktree")
	}

	// Clean up
	os.RemoveAll("/tmp/test-worktrees")
}

func TestWorktreeExists(t *testing.T) {
	manager, err := NewWorktreeManager(&WorktreeConfig{
		BaseDirectory: "/tmp/test-worktrees",
		MaxPathLength: 1024,
		EnableLogging: false,
	})
	if err != nil {
		t.Fatalf("Failed to create worktree manager: %v", err)
	}

	// Test non-existent worktree
	exists := manager.WorktreeExists("/tmp/test-worktrees/non-existent")
	if exists {
		t.Error("Non-existent worktree should not exist")
	}

	// Create a worktree and test
	ctx := context.Background()
	worktreePath, err := manager.CreateWorktree(ctx, "project-123", "task-456")
	if err != nil {
		t.Fatalf("Failed to create worktree: %v", err)
	}

	exists = manager.WorktreeExists(worktreePath)
	if !exists {
		t.Error("Created worktree should exist")
	}

	// Clean up
	os.RemoveAll("/tmp/test-worktrees")
}

func TestCleanupWorktree(t *testing.T) {
	manager, err := NewWorktreeManager(&WorktreeConfig{
		BaseDirectory: "/tmp/test-worktrees",
		MaxPathLength: 1024,
		EnableLogging: false,
	})
	if err != nil {
		t.Fatalf("Failed to create worktree manager: %v", err)
	}

	ctx := context.Background()

	// Create a worktree
	worktreePath, err := manager.CreateWorktree(ctx, "project-123", "task-456")
	if err != nil {
		t.Fatalf("Failed to create worktree: %v", err)
	}

	// Verify it exists
	if !manager.WorktreeExists(worktreePath) {
		t.Error("Created worktree should exist")
	}

	// Clean up worktree
	err = manager.CleanupWorktree(ctx, worktreePath)
	if err != nil {
		t.Fatalf("Failed to cleanup worktree: %v", err)
	}

	// Verify it's gone
	if manager.WorktreeExists(worktreePath) {
		t.Error("Cleaned up worktree should not exist")
	}

	// Test cleaning up non-existent worktree (should not error)
	err = manager.CleanupWorktree(ctx, "/tmp/test-worktrees/non-existent")
	if err != nil {
		t.Errorf("Cleaning up non-existent worktree should not error: %v", err)
	}

	// Clean up
	os.RemoveAll("/tmp/test-worktrees")
}

func TestListWorktrees(t *testing.T) {
	manager, err := NewWorktreeManager(&WorktreeConfig{
		BaseDirectory: "/tmp/test-worktrees",
		MaxPathLength: 1024,
		EnableLogging: false,
	})
	if err != nil {
		t.Fatalf("Failed to create worktree manager: %v", err)
	}

	ctx := context.Background()

	// Create multiple worktrees
	worktree1, err := manager.CreateWorktree(ctx, "project-123", "task-456")
	if err != nil {
		t.Fatalf("Failed to create worktree 1: %v", err)
	}

	worktree2, err := manager.CreateWorktree(ctx, "project-123", "task-789")
	if err != nil {
		t.Fatalf("Failed to create worktree 2: %v", err)
	}

	// List worktrees for project
	worktrees, err := manager.ListWorktrees("project-123")
	if err != nil {
		t.Fatalf("Failed to list worktrees: %v", err)
	}

	if len(worktrees) != 2 {
		t.Errorf("Expected 2 worktrees, got %d", len(worktrees))
	}

	// Verify both worktrees are in the list
	found1, found2 := false, false
	for _, wt := range worktrees {
		if wt == worktree1 {
			found1 = true
		}
		if wt == worktree2 {
			found2 = true
		}
	}

	if !found1 || !found2 {
		t.Error("Not all worktrees found in list")
	}

	// Test listing worktrees for non-existent project
	worktrees, err = manager.ListWorktrees("non-existent-project")
	if err != nil {
		t.Fatalf("Failed to list worktrees for non-existent project: %v", err)
	}

	if len(worktrees) != 0 {
		t.Errorf("Expected 0 worktrees for non-existent project, got %d", len(worktrees))
	}

	// Clean up
	os.RemoveAll("/tmp/test-worktrees")
}

func TestGetWorktreeInfo(t *testing.T) {
	manager, err := NewWorktreeManager(&WorktreeConfig{
		BaseDirectory: "/tmp/test-worktrees",
		MaxPathLength: 1024,
		EnableLogging: false,
	})
	if err != nil {
		t.Fatalf("Failed to create worktree manager: %v", err)
	}

	ctx := context.Background()

	// Create a worktree
	worktreePath, err := manager.CreateWorktree(ctx, "project-123", "task-456")
	if err != nil {
		t.Fatalf("Failed to create worktree: %v", err)
	}

	// Get worktree info
	info, err := manager.GetWorktreeInfo(worktreePath)
	if err != nil {
		t.Fatalf("Failed to get worktree info: %v", err)
	}

	if info == nil {
		t.Fatal("Worktree info should not be nil")
	}

	if info.Path != worktreePath {
		t.Errorf("Expected path %s, got %s", worktreePath, info.Path)
	}

	if !info.IsValid {
		t.Error("Worktree should be valid")
	}

	if info.FileCount != 0 {
		t.Errorf("Expected 0 files, got %d", info.FileCount)
	}

	// Test getting info for non-existent worktree
	_, err = manager.GetWorktreeInfo("/tmp/test-worktrees/non-existent")
	if err == nil {
		t.Error("Expected error when getting info for non-existent worktree")
	}

	// Clean up
	os.RemoveAll("/tmp/test-worktrees")
}
