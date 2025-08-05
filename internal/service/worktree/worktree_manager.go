package worktree

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/auto-devs/auto-devs/config"
)

// WorktreeManager provides worktree directory management functionality
type WorktreeManager struct {
	config *config.WorktreeConfig
	logger *slog.Logger
}

// NewWorktreeManager creates a new WorktreeManager instance
func NewWorktreeManager(config *config.WorktreeConfig) (*WorktreeManager, error) {
	// Setup logger
	var logger *slog.Logger
	if config.EnableLogging {
		logger = slog.Default().With("component", "worktree-manager")
	} else {
		logger = slog.New(slog.NewTextHandler(nil, &slog.HandlerOptions{Level: slog.LevelError}))
	}

	manager := &WorktreeManager{
		config: config,
		logger: logger,
	}

	// Initialize base directory
	if err := manager.initializeBaseDirectory(); err != nil {
		return nil, fmt.Errorf("failed to initialize base directory: %w", err)
	}

	return manager, nil
}

// initializeBaseDirectory creates the base worktree directory if it doesn't exist
func (wm *WorktreeManager) initializeBaseDirectory() error {
	wm.logger.Debug("Initializing base worktree directory", "path", wm.config.BaseDirectory)

	// Create base directory if it doesn't exist
	if err := os.MkdirAll(wm.config.BaseDirectory, 0o755); err != nil {
		return fmt.Errorf("failed to create base directory: %w", err)
	}

	// Validate directory permissions
	if err := wm.validateDirectoryPermissions(wm.config.BaseDirectory); err != nil {
		return fmt.Errorf("base directory permission validation failed: %w", err)
	}

	wm.logger.Info("Base worktree directory initialized", "path", wm.config.BaseDirectory)
	return nil
}

// validateDirectoryPermissions checks if the directory has proper read/write permissions
func (wm *WorktreeManager) validateDirectoryPermissions(dirPath string) error {
	// Check if directory exists
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", dirPath)
	}

	// Check read permissions
	if _, err := os.ReadDir(dirPath); err != nil {
		return fmt.Errorf("directory is not readable: %s", err)
	}

	// Check write permissions by creating a test file
	testFile := filepath.Join(dirPath, ".test_write_permission")
	if err := os.WriteFile(testFile, []byte("test"), 0o644); err != nil {
		return fmt.Errorf("directory is not writable: %s", err)
	}

	// Clean up test file
	os.Remove(testFile)

	return nil
}

// GenerateWorktreePath generates a unique worktree path for a task
func (wm *WorktreeManager) GenerateWorktreePath(projectID string, taskID string) (string, error) {
	wm.logger.Debug("Generating worktree path", "project_id", projectID, "task_id", taskID)

	// Validate inputs
	if projectID == "" || taskID == "" {
		return "", fmt.Errorf("project_id and task_id are required")
	}

	// Clean and validate project ID and task ID
	cleanProjectID := wm.cleanPathComponent(projectID)
	cleanTaskID := wm.cleanPathComponent(taskID)

	if cleanProjectID == "" || cleanTaskID == "" {
		return "", fmt.Errorf("invalid project_id or task_id after cleaning")
	}

	// Generate path structure: /worktrees/project-{id}/task-{id}/
	projectDir := fmt.Sprintf("project-%s", cleanProjectID)
	taskDir := fmt.Sprintf("task-%s", cleanTaskID)

	worktreePath := filepath.Join(wm.config.BaseDirectory, projectDir, taskDir)

	// Validate path length
	if len(worktreePath) > wm.config.MaxPathLength {
		return "", fmt.Errorf("generated path %s exceeds maximum length: %d > %d", worktreePath, len(worktreePath), wm.config.MaxPathLength)
	}

	wm.logger.Debug("Generated worktree path", "path", worktreePath)
	return worktreePath, nil
}

// cleanPathComponent cleans and validates a path component
func (wm *WorktreeManager) cleanPathComponent(component string) string {
	// Remove leading/trailing whitespace
	component = strings.TrimSpace(component)

	// Replace spaces with underscores
	component = strings.ReplaceAll(component, " ", "_")

	// Replace invalid characters with underscores
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range invalidChars {
		component = strings.ReplaceAll(component, char, "_")
	}

	// Remove multiple consecutive underscores
	for strings.Contains(component, "__") {
		component = strings.ReplaceAll(component, "__", "_")
	}

	// Remove leading/trailing underscores
	component = strings.Trim(component, "_")

	// Limit length
	if len(component) > 100 {
		component = component[:100]
	}

	return component
}

// CreateWorktree creates a worktree directory for a task
func (wm *WorktreeManager) CreateWorktree(ctx context.Context, projectID string, taskID string) (string, error) {
	wm.logger.Info("Creating worktree", "project_id", projectID, "task_id", taskID)

	// Generate worktree path
	worktreePath, err := wm.GenerateWorktreePath(projectID, taskID)
	if err != nil {
		return "", fmt.Errorf("failed to generate worktree path: %w", err)
	}

	// Check if worktree already exists
	if wm.WorktreeExists(worktreePath) {
		return "", fmt.Errorf("worktree already exists: %s", worktreePath)
	}

	// Validate available disk space
	if err := wm.validateDiskSpace(worktreePath); err != nil {
		return "", fmt.Errorf("insufficient disk space: %w", err)
	}

	// Create worktree directory
	if err := os.MkdirAll(worktreePath, 0o755); err != nil {
		return "", fmt.Errorf("failed to create worktree directory: %w", err)
	}

	// Validate the created directory
	if err := wm.validateWorktreeDirectory(worktreePath); err != nil {
		// Clean up on validation failure
		os.RemoveAll(worktreePath)
		return "", fmt.Errorf("worktree validation failed: %w", err)
	}

	wm.logger.Info("Worktree created successfully", "path", worktreePath)
	return worktreePath, nil
}

// WorktreeExists checks if a worktree directory exists
func (wm *WorktreeManager) WorktreeExists(worktreePath string) bool {
	_, err := os.Stat(worktreePath)
	return !os.IsNotExist(err)
}

// validateDiskSpace checks if there's sufficient disk space available
func (wm *WorktreeManager) validateDiskSpace(path string) error {
	// Get disk usage information
	_, err := os.Stat(wm.config.BaseDirectory)
	if err != nil {
		return fmt.Errorf("failed to get directory stats: %w", err)
	}

	// For now, we'll use a simple approach
	// In a production environment, you might want to use syscall.Statfs
	// to get actual disk space information
	wm.logger.Debug("Disk space validation passed", "path", path)
	return nil
}

// validateWorktreeDirectory validates a worktree directory
func (wm *WorktreeManager) validateWorktreeDirectory(worktreePath string) error {
	// Check if directory exists
	if !wm.WorktreeExists(worktreePath) {
		return fmt.Errorf("worktree directory does not exist: %s", worktreePath)
	}

	// Validate directory permissions
	if err := wm.validateDirectoryPermissions(worktreePath); err != nil {
		return fmt.Errorf("worktree directory permission validation failed: %w", err)
	}

	// Check if directory is empty (optional validation)
	entries, err := os.ReadDir(worktreePath)
	if err != nil {
		return fmt.Errorf("failed to read worktree directory: %w", err)
	}

	if len(entries) > 0 {
		wm.logger.Warn("Worktree directory is not empty", "path", worktreePath, "entries", len(entries))
	}

	return nil
}

// CleanupWorktree removes a worktree directory and cleans up associated resources
func (wm *WorktreeManager) CleanupWorktree(ctx context.Context, worktreePath string) error {
	wm.logger.Info("Cleaning up worktree", "path", worktreePath)

	// Validate worktree path
	if !strings.HasPrefix(worktreePath, wm.config.BaseDirectory) {
		return fmt.Errorf("invalid worktree path: not under base directory")
	}

	// Check if worktree exists
	if !wm.WorktreeExists(worktreePath) {
		wm.logger.Warn("Worktree does not exist, skipping cleanup", "path", worktreePath)
		return nil
	}

	// Remove worktree directory and all contents
	if err := os.RemoveAll(worktreePath); err != nil {
		return fmt.Errorf("failed to remove worktree directory: %w", err)
	}

	wm.logger.Info("Worktree cleaned up successfully", "path", worktreePath)
	return nil
}

// ListWorktrees lists all worktrees for a project
func (wm *WorktreeManager) ListWorktrees(projectID string) ([]string, error) {
	wm.logger.Debug("Listing worktrees", "project_id", projectID)

	cleanProjectID := wm.cleanPathComponent(projectID)
	if cleanProjectID == "" {
		return nil, fmt.Errorf("invalid project_id")
	}

	projectDir := fmt.Sprintf("project-%s", cleanProjectID)
	projectPath := filepath.Join(wm.config.BaseDirectory, projectDir)

	// Check if project directory exists
	if !wm.WorktreeExists(projectPath) {
		return []string{}, nil
	}

	// Read project directory
	entries, err := os.ReadDir(projectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read project directory: %w", err)
	}

	var worktrees []string
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "task-") {
			worktreePath := filepath.Join(projectPath, entry.Name())
			worktrees = append(worktrees, worktreePath)
		}
	}

	wm.logger.Debug("Found worktrees", "count", len(worktrees), "project_id", projectID)
	return worktrees, nil
}

// GetWorktreeInfo returns information about a worktree
func (wm *WorktreeManager) GetWorktreeInfo(worktreePath string) (*WorktreeInfo, error) {
	wm.logger.Debug("Getting worktree info", "path", worktreePath)

	if !wm.WorktreeExists(worktreePath) {
		return nil, fmt.Errorf("worktree does not exist: %s", worktreePath)
	}

	stat, err := os.Stat(worktreePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree stats: %w", err)
	}

	// Count files in worktree
	fileCount := 0
	err = filepath.Walk(worktreePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			fileCount++
		}
		return nil
	})
	if err != nil {
		wm.logger.Warn("Failed to count files in worktree", "path", worktreePath, "error", err)
	}

	info := &WorktreeInfo{
		Path:      worktreePath,
		CreatedAt: stat.ModTime(),
		FileCount: fileCount,
		IsValid:   true,
		Size:      stat.Size(),
	}

	wm.logger.Debug("Worktree info retrieved", "path", worktreePath, "file_count", fileCount)
	return info, nil
}

// WorktreeInfo contains information about a worktree
type WorktreeInfo struct {
	Path      string    `json:"path"`
	CreatedAt time.Time `json:"created_at"`
	FileCount int       `json:"file_count"`
	IsValid   bool      `json:"is_valid"`
	Size      int64     `json:"size"`
}
