# Git Service - Branch Management System

## Overview

The Branch Management System provides comprehensive branch naming and lifecycle management for Git repositories. It integrates with the existing Git service infrastructure to provide automated branch operations with proper validation and conflict detection.

## Features

### 1. Branch Naming Conventions

The system supports configurable branch naming conventions through the `BranchNamingConfig` struct:

```go
type BranchNamingConfig struct {
    Prefix       string // e.g., "task"
    IncludeID    bool   // Include task ID in name
    Separator    string // e.g., "-"
    MaxLength    int    // Maximum branch name length (default: 255)
    UseSlug      bool   // Use slugified title
}
```

**Default Configuration:**

- Prefix: "task"
- IncludeID: true
- Separator: "-"
- MaxLength: 255
- UseSlug: true

**Example Branch Names:**

- `task-123-implement-user-authentication`
- `feature-fix-bug` (without ID)
- `task-456-add-api-endpoint-for-user-example-com`

### 2. Branch Name Generation

The system automatically generates branch names based on task information:

```go
branchName, err := gitManager.GenerateBranchName("123", "Implement user authentication")
// Result: "task-123-implement-user-authentication"
```

**Features:**

- Automatic slugification of titles
- Special character handling
- Length limit enforcement
- Git naming rule compliance

### 3. Branch Lifecycle Management

#### Create Branch from Main

```go
err := gitManager.CreateBranchFromMain(ctx, workingDir, "task-123-feature")
```

#### Switch to Branch

```go
err := gitManager.SwitchToBranch(ctx, workingDir, "task-123-feature")
```

#### Delete Branch

```go
err := gitManager.DeleteBranch(ctx, workingDir, "task-123-feature", false)
```

### 4. Conflict Detection

The system detects potential branch naming conflicts:

```go
conflictInfo, err := gitManager.CheckBranchConflict(ctx, workingDir, "task-123-feature")
if conflictInfo.HasConflict {
    // Handle conflict
    fmt.Println(conflictInfo.ConflictMessage)
}
```

**Conflict Types:**

- `branch_exists`: Branch with the same name already exists
- `similar_branches`: Found branches with similar names

### 5. Branch Name Validation

Comprehensive validation against Git naming rules:

```go
result, err := gitManager.ValidateBranchNameFormat("task-123-feature")
if !result.IsValid {
    for _, issue := range result.Issues {
        fmt.Println(issue)
    }
}
```

**Validation Rules:**

- Cannot be empty
- Maximum length: 255 characters
- Cannot start with dot (.)
- Cannot end with .lock
- Cannot contain @{, ^, ~, :, ?, \*, [, \, whitespace
- Cannot contain control characters
- Cannot contain consecutive dots (..)
- Cannot use reserved names (HEAD, ORIG_HEAD, etc.)

## Usage Examples

### Basic Usage

```go
// Initialize Git Manager
config := &git.ManagerConfig{
    DefaultTimeout: 30 * time.Second,
    MaxRetries:     3,
    EnableLogging:  true,
}

gitManager, err := git.NewGitManager(config)
if err != nil {
    log.Fatal(err)
}

// Generate branch name for a task
branchName, err := gitManager.GenerateBranchName("123", "Implement user authentication")
if err != nil {
    log.Fatal(err)
}

// Check for conflicts
conflictInfo, err := gitManager.CheckBranchConflict(ctx, "/path/to/repo", branchName)
if err != nil {
    log.Fatal(err)
}

if conflictInfo.HasConflict {
    log.Printf("Conflict detected: %s", conflictInfo.ConflictMessage)
    return
}

// Create branch
err = gitManager.CreateBranchFromMain(ctx, "/path/to/repo", branchName)
if err != nil {
    log.Fatal(err)
}

// Switch to branch
err = gitManager.SwitchToBranch(ctx, "/path/to/repo", branchName)
if err != nil {
    log.Fatal(err)
}
```

### Custom Branch Naming Configuration

```go
// Create custom branch naming configuration
branchConfig := &git.BranchNamingConfig{
    Prefix:       "feature",
    IncludeID:    false,
    Separator:    "_",
    MaxLength:    100,
    UseSlug:      true,
}

// Initialize branch manager with custom config
branchManager := git.NewBranchManager(commands, validator, branchConfig)

// Generate branch name
branchName, err := branchManager.GenerateBranchName("123", "Add new API endpoint")
// Result: "feature_add_new_api_endpoint"
```

### Error Handling

```go
// Handle different types of errors
branchName, err := gitManager.GenerateBranchName("123", "Invalid@Branch#Name")
if err != nil {
    if git.IsBranchError(err) {
        // Handle branch-specific errors
        log.Printf("Branch error: %v", err)
    } else {
        // Handle other errors
        log.Printf("Unexpected error: %v", err)
    }
}

// Validate branch name format
result, err := gitManager.ValidateBranchNameFormat(branchName)
if err != nil {
    log.Fatal(err)
}

if !result.IsValid {
    log.Printf("Invalid branch name: %s", branchName)
    for _, issue := range result.Issues {
        log.Printf("  - %s", issue)
    }
}
```

## Integration with Git Manager

The Branch Management System is fully integrated with the Git Manager:

```go
// All branch management methods are available through Git Manager
gitManager := git.NewGitManager(config)

// Branch name generation
branchName, err := gitManager.GenerateBranchName(taskID, title)

// Branch operations
err = gitManager.CreateBranchFromMain(ctx, workingDir, branchName)
err = gitManager.SwitchToBranch(ctx, workingDir, branchName)
err = gitManager.DeleteBranch(ctx, workingDir, branchName, force)

// Conflict detection
conflictInfo, err := gitManager.CheckBranchConflict(ctx, workingDir, branchName)

// Validation
result, err := gitManager.ValidateBranchNameFormat(branchName)
```

## Testing

The system includes comprehensive tests covering:

- Branch naming configuration
- Branch name generation with various inputs
- Title slugification and processing
- Branch name cleaning and validation
- Conflict detection
- Similarity calculation
- Integration tests

Run tests with:

```bash
go test ./internal/service/git -v -run "TestBranchManager"
```

## Performance

The system is optimized for performance:

- Efficient string processing algorithms
- Minimal memory allocations
- Fast similarity calculations
- Optimized regex patterns

Benchmark results:

```bash
go test ./internal/service/git -bench "BenchmarkBranchManager"
```

## Error Handling

The system provides detailed error information:

- Structured error types
- Contextual error messages
- Suggested solutions
- Error classification functions

## Security Considerations

- Input validation for all branch names
- Sanitization of special characters
- Prevention of Git injection attacks
- Safe handling of user-provided data

## Future Enhancements

- Support for custom branch naming patterns
- Advanced conflict resolution strategies
- Branch name suggestions
- Integration with external naming conventions
- Support for branch templates
