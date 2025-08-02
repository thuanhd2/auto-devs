# Repository Layer Implementation

This document describes the repository layer implementation for the Auto-Devs project.

## Overview

The repository layer follows the Clean Architecture pattern and provides data access abstractions for the domain entities. It includes:

- Entity models with validation tags
- Repository interfaces defining contracts
- PostgreSQL implementations with comprehensive error handling
- Unit tests using testcontainers for integration testing

## Entity Models

### Project Entity
- `ID`: UUID primary key
- `Name`: Project name (required, max 255 chars)
- `Description`: Project description (max 1000 chars)
- `RepoURL`: Git repository URL (required, valid URL, max 500 chars)
- `CreatedAt`, `UpdatedAt`: Timestamps with automatic management

### Task Entity
- `ID`: UUID primary key
- `ProjectID`: Foreign key reference to Project
- `Title`: Task title (required, max 255 chars)
- `Description`: Task description (max 1000 chars)
- `Status`: Task status with predefined enum values
- `BranchName`: Optional Git branch name
- `PullRequest`: Optional pull request reference
- `CreatedAt`, `UpdatedAt`: Timestamps with automatic management

### Task Status Enum
- `TODO`: Initial state
- `PLANNING`: Task planning phase
- `PLAN_REVIEWING`: Plan review phase
- `IMPLEMENTING`: Implementation phase
- `CODE_REVIEWING`: Code review phase
- `DONE`: Completed state
- `CANCELLED`: Cancelled state

## Repository Interfaces

### ProjectRepository
- `Create(ctx, project)`: Create new project
- `GetByID(ctx, id)`: Retrieve project by ID
- `GetAll(ctx)`: Retrieve all projects
- `Update(ctx, project)`: Update existing project
- `Delete(ctx, id)`: Delete project
- `GetWithTaskCount(ctx, id)`: Retrieve project with task count

### TaskRepository
- `Create(ctx, task)`: Create new task
- `GetByID(ctx, id)`: Retrieve task by ID
- `GetByProjectID(ctx, projectID)`: Retrieve tasks by project
- `Update(ctx, task)`: Update existing task
- `Delete(ctx, id)`: Delete task
- `UpdateStatus(ctx, id, status)`: Update task status
- `GetByStatus(ctx, status)`: Retrieve tasks by status

## PostgreSQL Implementation

### Features
- **Error Handling**: Comprehensive error handling with PostgreSQL-specific error codes
- **Validation**: Database constraint validation with meaningful error messages
- **Transactions**: Support for database transactions via `WithTransaction` helper
- **Connection Pooling**: Configurable connection pool settings
- **Timestamps**: Automatic timestamp management via database triggers
- **Foreign Key Constraints**: Proper referential integrity

### Database Connection Configuration
```go
type Config struct {
    Host            string
    Port            string
    Username        string
    Password        string
    DBName          string
    SSLMode         string
    MaxOpenConns    int           // Default: 25
    MaxIdleConns    int           // Default: 5
    ConnMaxLifetime time.Duration // Default: 15 minutes
    ConnMaxIdleTime time.Duration // Default: 5 minutes
    ConnTimeout     time.Duration // Optional connection timeout
}
```

### Error Handling
The implementation provides specific error handling for:
- **Unique Violations**: Duplicate entries
- **Foreign Key Violations**: Invalid references
- **Check Violations**: Constraint violations (e.g., invalid task status)
- **Not Found**: Entity not found errors
- **Connection Issues**: Database connectivity problems

## Testing

### Test Coverage
- **Current Coverage**: 74.4% of statements
- **Test Framework**: Go testing with testify assertions
- **Integration Testing**: testcontainers with PostgreSQL 15
- **Test Database**: Isolated PostgreSQL container per test

### Test Structure
Each repository has comprehensive tests covering:
- CRUD operations (Create, Read, Update, Delete)
- Error conditions (not found, invalid data, constraint violations)
- Edge cases (nullable fields, default values)
- Complex queries (joins, aggregations)

### Running Tests
```bash
# Run all repository tests
go test ./internal/repository/postgres -v

# Run with coverage
go test ./internal/repository/postgres -cover

# Generate coverage report
go test ./internal/repository/postgres -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Usage Example

```go
// Initialize database
config := database.Config{
    Host:     "localhost",
    Port:     "5432",
    Username: "postgres",
    Password: "password",
    DBName:   "autodevs",
    SSLMode:  "disable",
}

db, err := database.NewConnection(config)
if err != nil {
    log.Fatal(err)
}
defer db.Close()

// Create repositories
projectRepo := postgres.NewProjectRepository(db)
taskRepo := postgres.NewTaskRepository(db)

// Use repositories
ctx := context.Background()

project := &entity.Project{
    Name:        "My Project",
    Description: "Project description",
    RepoURL:     "https://github.com/user/repo.git",
}

err = projectRepo.Create(ctx, project)
if err != nil {
    log.Fatal(err)
}
```

## Migration Support

The implementation works with the existing database schema defined in:
- `migrations/000001_init.up.sql`: Database schema creation
- `migrations/000001_init.down.sql`: Database schema rollback

## Next Steps

The repository layer is ready for integration with:
1. **Use Case Layer**: Business logic implementation
2. **Handler Layer**: HTTP API endpoints
3. **Dependency Injection**: Wire setup for dependency management

## Files Structure

```
internal/repository/
├── README.md                          # This documentation
├── project.go                         # Project repository interface
├── task.go                           # Task repository interface
└── postgres/
    ├── project_repository.go         # PostgreSQL project implementation
    ├── project_repository_test.go    # Project repository tests
    ├── task_repository.go            # PostgreSQL task implementation
    └── task_repository_test.go       # Task repository tests
```