# GORM Migration Documentation

## Tổng quan

Dự án đã được refactor từ raw SQL sang sử dụng GORM (Go Object Relational Mapper) để cải thiện khả năng bảo trì và giảm boilerplate code.

## Những thay đổi chính

### 1. Entity Updates

#### Project Entity (`internal/entity/project.go`)

- Thay thế `db` tags bằng `gorm` tags
- Thêm `gorm.DeletedAt` cho soft delete
- Thêm relationship với Task entity
- Sử dụng `autoCreateTime` và `autoUpdateTime` cho timestamps

```go
type Project struct {
    ID          uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
    Name        string         `json:"name" gorm:"size:255;not null"`
    Description string         `json:"description" gorm:"size:1000"`
    RepoURL     string         `json:"repo_url" gorm:"size:500;not null"`
    CreatedAt   time.Time      `json:"created_at" gorm:"autoCreateTime"`
    UpdatedAt   time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
    DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

    // Relationships
    Tasks []Task `json:"tasks,omitempty" gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE"`
}
```

#### Task Entity (`internal/entity/task.go`)

- Tương tự như Project, thay thế `db` tags bằng `gorm` tags
- Thêm relationship với Project entity
- Sử dụng default value cho Status

```go
type Task struct {
    ID          uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
    ProjectID   uuid.UUID      `json:"project_id" gorm:"type:uuid;not null"`
    Title       string         `json:"title" gorm:"size:255;not null"`
    Description string         `json:"description" gorm:"size:1000"`
    Status      TaskStatus     `json:"status" gorm:"size:50;not null;default:'TODO'"`
    BranchName  *string        `json:"branch_name,omitempty" gorm:"size:255"`
    PullRequest *string        `json:"pull_request,omitempty" gorm:"size:255"`
    CreatedAt   time.Time      `json:"created_at" gorm:"autoCreateTime"`
    UpdatedAt   time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
    DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

    // Relationships
    Project Project `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
}
```

### 2. Database Connection

#### GormDB (`pkg/database/gorm.go`)

- Tạo wrapper cho GORM database connection
- Cấu hình connection pool
- Tích hợp với config system
- Hỗ trợ AutoMigrate

```go
type GormDB struct {
    *gorm.DB
}

func NewGormDB(cfg *config.Config) (*GormDB, error) {
    // Connection setup with config
    // Connection pool configuration
    // Logger configuration
}
```

### 3. Repository Refactoring

#### Project Repository (`internal/repository/postgres/project_repository.go`)

- Thay thế raw SQL queries bằng GORM methods
- Sử dụng `Create`, `First`, `Find`, `Save`, `Delete`
- Tự động handle timestamps và UUID generation
- Soft delete support

```go
// Before (Raw SQL)
query := `INSERT INTO projects (id, name, description, repo_url, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`
_, err := r.db.ExecContext(ctx, query, ...)

// After (GORM)
result := r.db.WithContext(ctx).Create(project)
```

#### Task Repository (`internal/repository/postgres/task_repository.go`)

- Tương tự như Project Repository
- Sử dụng GORM relationships và constraints
- Simplified error handling

### 4. Migration System

#### AutoMigrate (`pkg/database/migration.go`)

- Thay thế file-based migrations bằng GORM AutoMigrate
- Tự động tạo tables, indexes, foreign keys dựa trên struct tags
- Simplified migration process

```go
func RunMigrations(db *GormDB) error {
    return db.AutoMigrate(
        &entity.Project{},
        &entity.Task{},
    )
}
```

### 5. Dependency Injection

#### Wire Configuration (`internal/di/wire.go`)

- Cập nhật để sử dụng GormDB thay vì raw DB
- Thêm repository providers
- Simplified dependency tree

## Lợi ích của việc sử dụng GORM

### 1. Giảm Boilerplate Code

- Không cần viết raw SQL queries
- Tự động handle timestamps, UUIDs, relationships
- Simplified CRUD operations

### 2. Type Safety

- Compile-time checking cho database operations
- Better IDE support với autocomplete
- Reduced runtime errors

### 3. Relationships

- Easy handling of foreign keys
- Automatic joins và eager loading
- Cascade operations

### 4. Migration

- AutoMigrate dựa trên struct definitions
- Automatic schema updates
- Version control friendly

### 5. Performance

- Connection pooling
- Query optimization
- Lazy loading support

## Cách sử dụng

### 1. Tạo mới record

```go
project := &entity.Project{
    Name: "New Project",
    Description: "Project description",
    RepoURL: "https://github.com/user/repo",
}
err := projectRepo.Create(ctx, project)
```

### 2. Query với relationships

```go
// GORM sẽ tự động handle joins
var project entity.Project
db.Preload("Tasks").First(&project, projectID)
```

### 3. Soft Delete

```go
// GORM tự động set DeletedAt timestamp
err := projectRepo.Delete(ctx, projectID)
```

### 4. AutoMigrate

```go
// Tự động tạo/update schema
err := db.AutoMigrate(&entity.Project{}, &entity.Task{})
```

## Migration từ Raw SQL

### 1. Thay thế SQL queries

- `INSERT` → `Create()`
- `SELECT` → `First()`, `Find()`
- `UPDATE` → `Save()`, `Updates()`
- `DELETE` → `Delete()`

### 2. Error handling

- GORM cung cấp standardized error types
- Simplified error checking
- Better error messages

### 3. Transactions

```go
err := db.Transaction(func(tx *gorm.DB) error {
    // Transaction operations
    return nil
})
```

## Testing

Repository tests cần được cập nhật để sử dụng GORM test database:

```go
func TestProjectRepository_Create(t *testing.T) {
    // Setup GORM test database
    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    require.NoError(t, err)

    // Run migrations
    err = db.AutoMigrate(&entity.Project{})
    require.NoError(t, err)

    // Test repository operations
    repo := postgres.NewProjectRepository(&database.GormDB{DB: db})
    // ... test logic
}
```

## Kết luận

Việc refactor sang GORM đã mang lại những cải thiện đáng kể:

1. **Code maintainability**: Giảm boilerplate code, tăng readability
2. **Type safety**: Compile-time checking, better IDE support
3. **Relationships**: Easy handling of complex relationships
4. **Migration**: Simplified schema management
5. **Performance**: Built-in optimizations và connection pooling

GORM là một lựa chọn tuyệt vời cho Go applications cần ORM capabilities với PostgreSQL.
