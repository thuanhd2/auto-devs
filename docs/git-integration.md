# Git Integration for Projects

## Tổng quan

Hệ thống đã được mở rộng để hỗ trợ tích hợp Git cơ bản cho các project. Tính năng này cho phép:

- Cấu hình repository Git cho project
- Validation cấu hình Git
- Kiểm tra trạng thái repository
- Setup tự động repository

## Cấu trúc Database

### Bảng `projects` - Các trường Git mới

```sql
-- Các trường Git mới được thêm vào bảng projects
ALTER TABLE projects
ADD COLUMN repository_url VARCHAR(500),
ADD COLUMN main_branch VARCHAR(100) DEFAULT 'main',
ADD COLUMN worktree_base_path VARCHAR(500),
ADD COLUMN git_auth_method VARCHAR(20),
ADD COLUMN git_enabled BOOLEAN DEFAULT false;
```

### Indexes

```sql
-- Indexes cho các truy vấn Git
CREATE INDEX idx_projects_git_enabled ON projects(git_enabled);
CREATE INDEX idx_projects_repository_url ON projects(repository_url) WHERE repository_url IS NOT NULL;
CREATE INDEX idx_projects_git_auth_method ON projects(git_auth_method) WHERE git_auth_method IS NOT NULL;
```

## Entity Model

### Project Entity

```go
type Project struct {
    // Existing fields...

    // Git-related fields
    RepositoryURL    string `json:"repository_url" gorm:"column:repository_url;size:500"`
    MainBranch       string `json:"main_branch" gorm:"column:main_branch;default:main;size:100"`
    WorktreeBasePath string `json:"worktree_base_path" gorm:"column:worktree_base_path;size:500"`
    GitAuthMethod    string `json:"git_auth_method" gorm:"column:git_auth_method;size:20"` // "ssh" or "https"
    GitEnabled       bool   `json:"git_enabled" gorm:"column:git_enabled;default:false"`
}
```

## DTOs

### Request DTOs

```go
type ProjectCreateRequest struct {
    // Existing fields...

    // Git-related fields
    RepositoryURL    string `json:"repository_url,omitempty" binding:"omitempty,url,max=500"`
    MainBranch       string `json:"main_branch,omitempty" binding:"omitempty,max=100"`
    WorktreeBasePath string `json:"worktree_base_path,omitempty" binding:"omitempty,max=500"`
    GitAuthMethod    string `json:"git_auth_method,omitempty" binding:"omitempty,oneof=ssh https"`
    GitEnabled       bool   `json:"git_enabled,omitempty"`
}

type GitProjectValidationRequest struct {
    RepositoryURL    string `json:"repository_url" binding:"required,url,max=500"`
    MainBranch       string `json:"main_branch" binding:"required,max=100"`
    WorktreeBasePath string `json:"worktree_base_path" binding:"required,max=500"`
    GitAuthMethod    string `json:"git_auth_method" binding:"required,oneof=ssh https"`
    GitEnabled       bool   `json:"git_enabled"`
}
```

### Response DTOs

```go
type ProjectResponse struct {
    // Existing fields...

    // Git-related fields
    RepositoryURL    string `json:"repository_url,omitempty"`
    MainBranch       string `json:"main_branch,omitempty"`
    WorktreeBasePath string `json:"worktree_base_path,omitempty"`
    GitAuthMethod    string `json:"git_auth_method,omitempty"`
    GitEnabled       bool   `json:"git_enabled"`
}

type GitProjectValidationResponse struct {
    Valid   bool     `json:"valid"`
    Message string   `json:"message,omitempty"`
    Errors  []string `json:"errors,omitempty"`
}

type GitProjectStatusResponse struct {
    GitEnabled       bool                `json:"git_enabled"`
    WorktreeExists   bool                `json:"worktree_exists"`
    RepositoryValid  bool                `json:"repository_valid"`
    CurrentBranch    string              `json:"current_branch,omitempty"`
    RemoteURL        string              `json:"remote_url,omitempty"`
    OnMainBranch     bool                `json:"on_main_branch"`
    WorkingDirStatus *WorkingDirStatus   `json:"working_dir_status,omitempty"`
    Status           string              `json:"status"`
}
```

## Services

### ProjectGitService

Service chính để xử lý các thao tác Git cho project:

```go
type ProjectGitService struct {
    validator *GitValidator
    manager   *GitManager
    commands  *GitCommands
}
```

#### Các method chính:

1. **ValidateGitProjectConfig** - Validation cấu hình Git
2. **TestGitConnection** - Kiểm tra kết nối repository
3. **SetupGitProject** - Setup repository cho project
4. **GetGitProjectStatus** - Lấy trạng thái Git của project

## Validation Rules

### Repository URL Validation

- Hỗ trợ HTTPS và SSH URLs
- Validation format URL
- Kiểm tra hostname hợp lệ

### Authentication Method Validation

- `https`: Cho HTTPS URLs
- `ssh`: Cho SSH URLs
- Validation tương thích giữa URL và auth method

### Worktree Base Path Validation

- Phải là absolute path
- Không được chứa `..`
- Validation quyền truy cập

## Migration

### Migration Up (000007_add_git_fields.up.sql)

```sql
-- Add Git-related fields to projects table
ALTER TABLE projects
ADD COLUMN repository_url VARCHAR(500),
ADD COLUMN main_branch VARCHAR(100) DEFAULT 'main',
ADD COLUMN worktree_base_path VARCHAR(500),
ADD COLUMN git_auth_method VARCHAR(20),
ADD COLUMN git_enabled BOOLEAN DEFAULT false;

-- Create indexes for Git-related queries
CREATE INDEX idx_projects_git_enabled ON projects(git_enabled);
CREATE INDEX idx_projects_repository_url ON projects(repository_url) WHERE repository_url IS NOT NULL;
CREATE INDEX idx_projects_git_auth_method ON projects(git_auth_method) WHERE git_auth_method IS NOT NULL;

-- Migrate existing repo_url data to repository_url for backward compatibility
UPDATE projects
SET repository_url = repo_url
WHERE repo_url IS NOT NULL AND repository_url IS NULL;
```

### Migration Down (000007_add_git_fields.down.sql)

```sql
-- Drop indexes
DROP INDEX IF EXISTS idx_projects_git_enabled;
DROP INDEX IF EXISTS idx_projects_repository_url;
DROP INDEX IF EXISTS idx_projects_git_auth_method;

-- Drop Git-related columns from projects table
ALTER TABLE projects
DROP COLUMN IF EXISTS repository_url,
DROP COLUMN IF EXISTS main_branch,
DROP COLUMN IF EXISTS worktree_base_path,
DROP COLUMN IF EXISTS git_auth_method,
DROP COLUMN IF EXISTS git_enabled;
```

## Testing

### Unit Tests

Tất cả các chức năng Git đều có unit tests:

- `TestProjectGitService_ValidateGitProjectConfig`
- `TestProjectGitService_validateAuthMethodMatchesURL`
- `TestProjectGitService_validateWorktreeBasePath`
- `TestProjectGitService_isDirectoryEmpty`

### Test Coverage

Tests bao gồm:

- Validation cấu hình hợp lệ/không hợp lệ
- Authentication method matching
- Path validation
- Error handling

## Usage Examples

### Tạo Project với Git Integration

```go
// Tạo project với Git enabled
project := &entity.Project{
    Name: "My Git Project",
    Description: "Project with Git integration",
    RepoURL: "https://github.com/user/repo",
    RepositoryURL: "https://github.com/user/repo.git",
    MainBranch: "main",
    WorktreeBasePath: "/tmp/projects/repo",
    GitAuthMethod: "https",
    GitEnabled: true,
}
```

### Validation Git Configuration

```go
config := &GitProjectConfig{
    RepositoryURL: "https://github.com/user/repo.git",
    MainBranch: "main",
    WorktreeBasePath: "/tmp/projects/repo",
    GitAuthMethod: "https",
    GitEnabled: true,
}

err := projectGitService.ValidateGitProjectConfig(ctx, config)
if err != nil {
    // Handle validation error
}
```

### Test Git Connection

```go
err := projectGitService.TestGitConnection(ctx, config)
if err != nil {
    // Handle connection error
}
```

## Security Considerations

1. **Authentication**: Hỗ trợ cả SSH và HTTPS authentication
2. **Path Validation**: Ngăn chặn path traversal attacks
3. **URL Validation**: Validation chặt chẽ repository URLs
4. **Temporary Files**: Sử dụng temporary directories an toàn cho testing

## Future Enhancements

1. **Git Credentials Management**: Quản lý credentials an toàn
2. **Branch Management**: Tích hợp với task branches
3. **Webhook Integration**: Tự động sync với repository changes
4. **Conflict Resolution**: Xử lý merge conflicts
5. **Git History Integration**: Hiển thị commit history cho tasks
