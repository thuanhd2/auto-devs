package git

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
)

// BranchNamingConfig defines the configuration for branch naming conventions
type BranchNamingConfig struct {
	Prefix    string // e.g., "task"
	IncludeID bool   // Include task ID in name
	Separator string // e.g., "-"
	MaxLength int    // Maximum branch name length (default: 255)
	UseSlug   bool   // Use slugified title
}

// DefaultBranchNamingConfig returns the default branch naming configuration
func DefaultBranchNamingConfig() *BranchNamingConfig {
	return &BranchNamingConfig{
		Prefix:    "task",
		IncludeID: true,
		Separator: "-",
		MaxLength: 255,
		UseSlug:   true,
	}
}

// BranchManager provides branch management functionality
type BranchManager struct {
	commands  *GitCommands
	validator *GitValidator
	config    *BranchNamingConfig
	logger    *slog.Logger
}

// NewBranchManager creates a new BranchManager instance
func NewBranchManager(commands *GitCommands, validator *GitValidator, config *BranchNamingConfig) *BranchManager {
	if config == nil {
		config = DefaultBranchNamingConfig()
	}

	return &BranchManager{
		commands:  commands,
		validator: validator,
		config:    config,
		logger:    slog.Default().With("component", "branch-manager"),
	}
}

// GenerateBranchName generates a branch name based on task information
func (bm *BranchManager) GenerateBranchName(taskID string, title string) (string, error) {
	bm.logger.Debug("Generating branch name", "task_id", taskID, "title", title)

	var parts []string

	// Add prefix if configured
	if bm.config.Prefix != "" {
		parts = append(parts, bm.config.Prefix)
	}

	// Add task ID if configured
	if bm.config.IncludeID && taskID != "" {
		parts = append(parts, taskID)
	}

	// Add title (slugified if configured)
	if title != "" {
		if bm.config.UseSlug {
			slug := bm.slugifyTitle(title)
			if slug != "" {
				parts = append(parts, slug)
			}
		} else {
			// Use simple title processing
			simpleTitle := bm.simpleTitleProcess(title)
			if simpleTitle != "" {
				parts = append(parts, simpleTitle)
			}
		}
	}

	// Join parts with separator
	branchName := strings.Join(parts, bm.config.Separator)

	// Validate and clean the branch name
	branchName = bm.cleanBranchName(branchName)

	// Check length limit
	if len(branchName) > bm.config.MaxLength {
		branchName = branchName[:bm.config.MaxLength]
		// Ensure we don't cut in the middle of a word if possible
		if lastSep := strings.LastIndex(branchName, bm.config.Separator); lastSep > 0 {
			branchName = branchName[:lastSep]
		}
	}

	// Final validation
	if err := bm.validator.ValidateBranchName(branchName); err != nil {
		return "", fmt.Errorf("generated branch name validation failed: %w", err)
	}

	bm.logger.Debug("Generated branch name", "branch_name", branchName)
	return branchName, nil
}

// slugifyTitle converts a title to a URL-friendly slug
func (bm *BranchManager) slugifyTitle(title string) string {
	// Convert to lowercase
	slug := strings.ToLower(title)

	// Replace spaces and special characters with separator
	slug = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(slug, bm.config.Separator)

	// Remove leading and trailing separators
	slug = strings.Trim(slug, bm.config.Separator)

	// Limit length to prevent overly long slugs
	if len(slug) > 50 {
		slug = slug[:50]
		// Don't cut in the middle of a word if possible
		if lastSep := strings.LastIndex(slug, bm.config.Separator); lastSep > 0 {
			slug = slug[:lastSep]
		}
	}

	return slug
}

// simpleTitleProcess provides basic title processing without slugification
func (bm *BranchManager) simpleTitleProcess(title string) string {
	// Remove special characters and convert spaces to separators
	processed := regexp.MustCompile(`[^a-zA-Z0-9\s]+`).ReplaceAllString(title, "")
	processed = regexp.MustCompile(`\s+`).ReplaceAllString(processed, bm.config.Separator)
	processed = strings.Trim(processed, bm.config.Separator)

	// Limit length
	if len(processed) > 30 {
		processed = processed[:30]
		if lastSep := strings.LastIndex(processed, bm.config.Separator); lastSep > 0 {
			processed = processed[:lastSep]
		}
	}

	return processed
}

// cleanBranchName cleans and validates a branch name
func (bm *BranchManager) cleanBranchName(name string) string {
	// Remove leading/trailing separators
	name = strings.Trim(name, bm.config.Separator)

	// Replace multiple consecutive separators with single separator
	name = regexp.MustCompile(bm.config.Separator+`+`).ReplaceAllString(name, bm.config.Separator)

	// Ensure it doesn't start with a dot
	if strings.HasPrefix(name, ".") {
		name = "branch" + bm.config.Separator + name
	}

	// Ensure it doesn't end with .lock
	if strings.HasSuffix(name, ".lock") {
		name = strings.TrimSuffix(name, ".lock")
	}

	return name
}

// CreateBranchFromMain creates a new branch from the main/default branch
func (bm *BranchManager) CreateBranchFromMain(ctx context.Context, workingDir, branchName string) error {
	bm.logger.Info("Creating branch from main", "branch_name", branchName, "working_dir", workingDir)

	// Validate branch name
	if err := bm.validator.ValidateBranchName(branchName); err != nil {
		return fmt.Errorf("branch name validation failed: %w", err)
	}

	// Check for branch conflicts
	exists, err := bm.validator.CheckBranchExists(ctx, workingDir, branchName)
	if err != nil {
		return fmt.Errorf("failed to check branch existence: %w", err)
	}

	if exists {
		return fmt.Errorf("%w: branch '%s' already exists", ErrBranchAlreadyExists, branchName)
	}

	// Get current branch to ensure we're on main
	currentBranch, err := bm.commands.CurrentBranch(ctx, workingDir)
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	// Create the branch from current branch
	err = bm.commands.CreateBranch(ctx, workingDir, branchName, currentBranch)
	if err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}

	bm.logger.Info("Branch created successfully", "branch_name", branchName, "from_branch", currentBranch)
	return nil
}

// SwitchToBranch switches to the specified branch
func (bm *BranchManager) SwitchToBranch(ctx context.Context, workingDir, branchName string) error {
	bm.logger.Info("Switching to branch", "branch_name", branchName, "working_dir", workingDir)

	// Validate branch name
	if err := bm.validator.ValidateBranchName(branchName); err != nil {
		return fmt.Errorf("branch name validation failed: %w", err)
	}

	// Check if branch exists
	exists, err := bm.validator.CheckBranchExists(ctx, workingDir, branchName)
	if err != nil {
		return fmt.Errorf("failed to check branch existence: %w", err)
	}

	if !exists {
		return fmt.Errorf("%w: branch '%s' does not exist", ErrBranchNotFound, branchName)
	}

	// Check working directory status
	status, err := bm.validator.ValidateWorkingDirectory(ctx, workingDir)
	if err != nil {
		return fmt.Errorf("failed to validate working directory: %w", err)
	}

	if !status.IsClean {
		return fmt.Errorf("%w: cannot switch branches with uncommitted changes", ErrWorkingDirDirty)
	}

	// Switch to the branch
	err = bm.commands.Checkout(ctx, workingDir, branchName, false)
	if err != nil {
		return fmt.Errorf("failed to switch to branch: %w", err)
	}

	bm.logger.Info("Successfully switched to branch", "branch_name", branchName)
	return nil
}

// DeleteBranch deletes a branch with proper cleanup
func (bm *BranchManager) DeleteBranch(ctx context.Context, workingDir, branchName string, force bool) error {
	bm.logger.Info("Deleting branch", "branch_name", branchName, "working_dir", workingDir, "force", force)

	// Validate branch name
	if err := bm.validator.ValidateBranchName(branchName); err != nil {
		return fmt.Errorf("branch name validation failed: %w", err)
	}

	// Check if branch exists
	exists, err := bm.validator.CheckBranchExists(ctx, workingDir, branchName)
	if err != nil {
		return fmt.Errorf("failed to check branch existence: %w", err)
	}

	if !exists {
		return fmt.Errorf("%w: branch '%s' does not exist", ErrBranchNotFound, branchName)
	}

	// Get current branch
	currentBranch, err := bm.commands.CurrentBranch(ctx, workingDir)
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	// Cannot delete current branch unless forced
	if currentBranch == branchName && !force {
		return fmt.Errorf("%w: cannot delete current branch '%s'", ErrCannotDeleteBranch, branchName)
	}

	// Delete the branch
	err = bm.deleteBranchCommand(ctx, workingDir, branchName, force)
	if err != nil {
		return fmt.Errorf("failed to delete branch: %w", err)
	}

	bm.logger.Info("Branch deleted successfully", "branch_name", branchName)
	return nil
}

// deleteBranchCommand executes the actual branch deletion command
func (bm *BranchManager) deleteBranchCommand(ctx context.Context, workingDir, branchName string, force bool) error {
	args := []string{"branch"}
	if force {
		args = append(args, "-D")
	} else {
		args = append(args, "-d")
	}
	args = append(args, branchName)

	result, err := bm.commands.executor.Execute(ctx, workingDir, args...)
	if err != nil {
		return err
	}

	if result.ExitCode != 0 {
		return NewGitError("delete-branch", result.ExitCode, result.Command, result.Stdout, result.Stderr, nil)
	}

	return nil
}

// CheckBranchConflict checks for potential branch naming conflicts
func (bm *BranchManager) CheckBranchConflict(ctx context.Context, workingDir, branchName string) (*BranchConflictInfo, error) {
	bm.logger.Debug("Checking branch conflict", "branch_name", branchName, "working_dir", workingDir)

	conflictInfo := &BranchConflictInfo{
		BranchName:  branchName,
		HasConflict: false,
	}

	// Check if branch already exists
	exists, err := bm.validator.CheckBranchExists(ctx, workingDir, branchName)
	if err != nil {
		return nil, fmt.Errorf("failed to check branch existence: %w", err)
	}

	if exists {
		conflictInfo.HasConflict = true
		conflictInfo.ConflictType = "branch_exists"
		conflictInfo.ConflictMessage = fmt.Sprintf("Branch '%s' already exists", branchName)
		return conflictInfo, nil
	}

	// Check for similar branch names
	similarBranches, err := bm.findSimilarBranches(ctx, workingDir, branchName)
	if err != nil {
		bm.logger.Warn("Failed to check for similar branches", "error", err)
	} else if len(similarBranches) > 0 {
		conflictInfo.HasConflict = true
		conflictInfo.ConflictType = "similar_branches"
		conflictInfo.ConflictMessage = fmt.Sprintf("Found similar branches: %s", strings.Join(similarBranches, ", "))
		conflictInfo.SimilarBranches = similarBranches
	}

	return conflictInfo, nil
}

// findSimilarBranches finds branches with similar names
func (bm *BranchManager) findSimilarBranches(ctx context.Context, workingDir, branchName string) ([]string, error) {
	// Get all branches
	branches, err := bm.commands.ListBranches(ctx, workingDir, &ListBranchesOptions{})
	if err != nil {
		return nil, err
	}

	var similarBranches []string
	branchNameLower := strings.ToLower(branchName)

	for _, branch := range branches {
		branchLower := strings.ToLower(branch)

		// Check for exact prefix match
		if strings.HasPrefix(branchLower, branchNameLower) && branchLower != branchNameLower {
			similarBranches = append(similarBranches, branch)
		}

		// Check for high similarity (80% or more)
		if bm.calculateSimilarity(branchLower, branchNameLower) >= 0.8 {
			similarBranches = append(similarBranches, branch)
		}
	}

	return similarBranches, nil
}

// calculateSimilarity calculates similarity between two strings using Levenshtein distance
func (bm *BranchManager) calculateSimilarity(s1, s2 string) float64 {
	if s1 == s2 {
		return 1.0
	}

	len1, len2 := len(s1), len(s2)
	if len1 == 0 {
		return 0.0
	}
	if len2 == 0 {
		return 0.0
	}

	// Use simple similarity calculation for performance
	// Count common characters
	common := 0
	for _, char := range s1 {
		if strings.ContainsRune(s2, char) {
			common++
		}
	}

	maxLen := len1
	if len2 > maxLen {
		maxLen = len2
	}

	return float64(common) / float64(maxLen)
}

// ValidateBranchNameFormat validates branch name format according to Git rules
func (bm *BranchManager) ValidateBranchNameFormat(branchName string) (*BranchValidationResult, error) {
	result := &BranchValidationResult{
		BranchName: branchName,
		IsValid:    true,
		Issues:     []string{},
	}

	// Check length
	if len(branchName) > bm.config.MaxLength {
		result.IsValid = false
		result.Issues = append(result.Issues, fmt.Sprintf("Branch name too long (max %d characters)", bm.config.MaxLength))
	}

	if len(branchName) == 0 {
		result.IsValid = false
		result.Issues = append(result.Issues, "Branch name cannot be empty")
	}

	// Check for invalid characters and patterns
	invalidPatterns := map[string]string{
		`^\.`:       "Cannot start with dot",
		`\.\.$`:     "Cannot end with two dots",
		`^/`:        "Cannot start with slash",
		`/$`:        "Cannot end with slash",
		`//`:        "Cannot contain double slash",
		`\.lock$`:   "Cannot end with .lock",
		`@{`:        "Cannot contain @{",
		`\^`:        "Cannot contain ^",
		`~`:         "Cannot contain ~",
		`:`:         "Cannot contain :",
		`\?`:        "Cannot contain ?",
		`\*`:        "Cannot contain *",
		`\[`:        "Cannot contain [",
		`\\`:        "Cannot contain backslash",
		`\s`:        "Cannot contain whitespace",
		`\x00-\x1f`: "Cannot contain control characters",
		`\x7f`:      "Cannot contain DEL character",
	}

	for pattern, message := range invalidPatterns {
		matched, err := regexp.MatchString(pattern, branchName)
		if err != nil {
			continue // Skip invalid regex patterns
		}
		if matched {
			result.IsValid = false
			result.Issues = append(result.Issues, message)
		}
	}

	// Check for consecutive dots
	if strings.Contains(branchName, "..") {
		result.IsValid = false
		result.Issues = append(result.Issues, "Cannot contain consecutive dots")
	}

	// Check for reserved names
	reservedNames := []string{"HEAD", "ORIG_HEAD", "FETCH_HEAD", "MERGE_HEAD", "CHERRY_PICK_HEAD"}
	for _, reserved := range reservedNames {
		if strings.EqualFold(branchName, reserved) {
			result.IsValid = false
			result.Issues = append(result.Issues, fmt.Sprintf("Cannot use reserved name '%s'", reserved))
		}
	}

	return result, nil
}

// BranchConflictInfo represents information about branch conflicts
type BranchConflictInfo struct {
	BranchName      string   `json:"branch_name"`
	HasConflict     bool     `json:"has_conflict"`
	ConflictType    string   `json:"conflict_type,omitempty"`
	ConflictMessage string   `json:"conflict_message,omitempty"`
	SimilarBranches []string `json:"similar_branches,omitempty"`
}

// BranchValidationResult represents the result of branch name validation
type BranchValidationResult struct {
	BranchName string   `json:"branch_name"`
	IsValid    bool     `json:"is_valid"`
	Issues     []string `json:"issues,omitempty"`
}
