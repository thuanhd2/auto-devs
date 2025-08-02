# GORM Migration Documentation

## Tổng quan

Dự án sử dụng GORM (Go Object Relational Mapper) để cải thiện khả năng bảo trì và giảm boilerplate code.

## Database Connection

### GormDB (`pkg/database/gorm.go`)

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

### Project Repository (`internal/repository/postgres/project_repository.go`)

```go
// After (GORM)
result := r.db.WithContext(ctx).Create(project)
```

#### Task Repository (`internal/repository/postgres/task_repository.go`)

- Tương tự như Project Repository
- Sử dụng GORM relationships và constraints
- Simplified error handling

### 4. Migration System

#### AutoMigrate (`pkg/database/migration.go`)

```go
func RunMigrations(db *GormDB) error {
    return db.AutoMigrate(
        &entity.Project{},
        &entity.Task{},
    )
}
```

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
