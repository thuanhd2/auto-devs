package worktree

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/auto-devs/auto-devs/internal/service/git"
)

// ExampleWorktreeUsage demonstrates how to use the worktree service
func ExampleWorktreeUsage() {
	// 1. Create worktree manager with custom configuration
	worktreeConfig := &WorktreeConfig{
		BaseDirectory:   "/tmp/example-worktrees",
		MaxPathLength:   2048,
		MinDiskSpace:    100 * 1024 * 1024, // 100MB
		CleanupInterval: 24 * time.Hour,
		EnableLogging:   true,
		LogLevel:        0,
	}

	_, err := NewWorktreeManager(worktreeConfig)
	if err != nil {
		log.Fatalf("Failed to create worktree manager: %v", err)
	}

	// 2. Create integrated service with Git support
	gitConfig := &git.ManagerConfig{
		DefaultTimeout: 30 * time.Second,
		MaxRetries:     3,
		WorkingDir:     "/tmp/example-worktrees",
		EnableLogging:  true,
		LogLevel:       0,
	}

	integratedService, err := NewIntegratedWorktreeService(&IntegratedConfig{
		Worktree: worktreeConfig,
		Git:      gitConfig,
	})
	if err != nil {
		log.Fatalf("Failed to create integrated service: %v", err)
	}

	ctx := context.Background()

	// 3. Create a task worktree
	fmt.Println("=== Creating Task Worktree ===")
	createRequest := &CreateTaskWorktreeRequest{
		ProjectID: "project-123",
		TaskID:    "task-456",
		TaskTitle: "Implement new feature",
	}

	taskInfo, err := integratedService.CreateTaskWorktree(ctx, createRequest)
	if err != nil {
		log.Printf("Failed to create task worktree: %v", err)
		return
	}

	fmt.Printf("Created worktree: %s\n", taskInfo.WorktreePath)
	fmt.Printf("Branch name: %s\n", taskInfo.BranchName)
	fmt.Printf("Created at: %s\n", taskInfo.CreatedAt)

	// 4. Get worktree information
	fmt.Println("\n=== Getting Worktree Info ===")
	info, err := integratedService.GetTaskWorktreeInfo(ctx, "project-123", "task-456")
	if err != nil {
		log.Printf("Failed to get worktree info: %v", err)
	} else {
		fmt.Printf("Worktree path: %s\n", info.WorktreePath)
		fmt.Printf("Branch name: %s\n", info.BranchName)
		if info.WorktreeInfo != nil {
			fmt.Printf("File count: %d\n", info.WorktreeInfo.FileCount)
			fmt.Printf("Size: %d bytes\n", info.WorktreeInfo.Size)
		}
	}

	// 5. List all worktrees for a project
	fmt.Println("\n=== Listing Project Worktrees ===")
	worktrees, err := integratedService.ListProjectWorktrees(ctx, "project-123")
	if err != nil {
		log.Printf("Failed to list worktrees: %v", err)
	} else {
		fmt.Printf("Found %d worktrees:\n", len(worktrees))
		for i, wt := range worktrees {
			fmt.Printf("  %d. %s (branch: %s)\n", i+1, wt.WorktreePath, wt.BranchName)
		}
	}

	// 6. Clean up worktree
	fmt.Println("\n=== Cleaning Up Worktree ===")
	cleanupRequest := &CleanupTaskWorktreeRequest{
		ProjectID:  "project-123",
		TaskID:     "task-456",
		BranchName: taskInfo.BranchName,
	}

	err = integratedService.CleanupTaskWorktree(ctx, cleanupRequest)
	if err != nil {
		log.Printf("Failed to cleanup worktree: %v", err)
	} else {
		fmt.Println("Worktree cleaned up successfully")
	}

	// 7. Verify cleanup
	fmt.Println("\n=== Verifying Cleanup ===")
	_, err = integratedService.GetTaskWorktreeInfo(ctx, "project-123", "task-456")
	if err != nil {
		fmt.Println("Worktree no longer exists (expected)")
	} else {
		fmt.Println("Worktree still exists (unexpected)")
	}
}

// ExampleBasicWorktreeUsage demonstrates basic worktree operations without Git
func ExampleBasicWorktreeUsage() {
	// Create worktree manager
	manager, err := NewWorktreeManager(&WorktreeConfig{
		BaseDirectory: "/tmp/basic-worktrees",
		EnableLogging: true,
	})
	if err != nil {
		log.Fatalf("Failed to create worktree manager: %v", err)
	}

	ctx := context.Background()

	// Generate worktree path
	path, err := manager.GenerateWorktreePath("project-123", "task-456")
	if err != nil {
		log.Fatalf("Failed to generate worktree path: %v", err)
	}
	fmt.Printf("Generated path: %s\n", path)

	// Create worktree
	worktreePath, err := manager.CreateWorktree(ctx, "project-123", "task-456")
	if err != nil {
		log.Fatalf("Failed to create worktree: %v", err)
	}
	fmt.Printf("Created worktree: %s\n", worktreePath)

	// Check if worktree exists
	exists := manager.WorktreeExists(worktreePath)
	fmt.Printf("Worktree exists: %t\n", exists)

	// Get worktree info
	info, err := manager.GetWorktreeInfo(worktreePath)
	if err != nil {
		log.Fatalf("Failed to get worktree info: %v", err)
	}
	fmt.Printf("Worktree info: %+v\n", info)

	// List worktrees for project
	worktrees, err := manager.ListWorktrees("project-123")
	if err != nil {
		log.Fatalf("Failed to list worktrees: %v", err)
	}
	fmt.Printf("Project worktrees: %v\n", worktrees)

	// Clean up worktree
	err = manager.CleanupWorktree(ctx, worktreePath)
	if err != nil {
		log.Fatalf("Failed to cleanup worktree: %v", err)
	}
	fmt.Println("Worktree cleaned up")
}

// ExamplePathValidation demonstrates path validation features
func ExamplePathValidation() {
	manager, err := NewWorktreeManager(&WorktreeConfig{
		BaseDirectory: "/tmp/validation-worktrees",
		EnableLogging: false,
	})
	if err != nil {
		log.Fatalf("Failed to create worktree manager: %v", err)
	}

	// Test various path components
	testCases := []struct {
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

	fmt.Println("=== Path Validation Examples ===")
	for _, tc := range testCases {
		result := manager.cleanPathComponent(tc.input)
		fmt.Printf("Input: %q -> Output: %q (Expected: %q)\n", tc.input, result, tc.expected)
	}
}
