package git

import (
	"context"
	"testing"
)

func TestProjectGitService_ValidateGitProjectConfig(t *testing.T) {
	// Setup
	executor, _ := NewDefaultCommandExecutor()
	commands := NewGitCommands(executor)
	validator := NewGitValidator(commands)
	config := &ManagerConfig{}
	manager, _ := NewGitManager(config)
	service := NewProjectGitService(validator, manager, commands)

	ctx := context.Background()

	tests := []struct {
		name    string
		config  *GitProjectConfig
		wantErr bool
	}{
		{
			name: "valid config with Git enabled",
			config: &GitProjectConfig{
				RepositoryURL:    "https://github.com/user/repo.git",
				MainBranch:       "main",
				WorktreeBasePath: "/tmp/test-repo",
				GitAuthMethod:    "https",
				GitEnabled:       true,
			},
			wantErr: false,
		},
		{
			name: "valid config with Git disabled",
			config: &GitProjectConfig{
				GitEnabled: false,
			},
			wantErr: false,
		},
		{
			name: "missing repository URL when Git enabled",
			config: &GitProjectConfig{
				MainBranch:       "main",
				WorktreeBasePath: "/tmp/test-repo",
				GitAuthMethod:    "https",
				GitEnabled:       true,
			},
			wantErr: true,
		},
		{
			name: "missing main branch when Git enabled",
			config: &GitProjectConfig{
				RepositoryURL:    "https://github.com/user/repo.git",
				WorktreeBasePath: "/tmp/test-repo",
				GitAuthMethod:    "https",
				GitEnabled:       true,
			},
			wantErr: true,
		},
		{
			name: "missing worktree base path when Git enabled",
			config: &GitProjectConfig{
				RepositoryURL: "https://github.com/user/repo.git",
				MainBranch:    "main",
				GitAuthMethod: "https",
				GitEnabled:    true,
			},
			wantErr: true,
		},
		{
			name: "missing auth method when Git enabled",
			config: &GitProjectConfig{
				RepositoryURL:    "https://github.com/user/repo.git",
				MainBranch:       "main",
				WorktreeBasePath: "/tmp/test-repo",
				GitEnabled:       true,
			},
			wantErr: true,
		},
		{
			name: "invalid auth method",
			config: &GitProjectConfig{
				RepositoryURL:    "https://github.com/user/repo.git",
				MainBranch:       "main",
				WorktreeBasePath: "/tmp/test-repo",
				GitAuthMethod:    "invalid",
				GitEnabled:       true,
			},
			wantErr: true,
		},
		{
			name: "HTTPS URL with SSH auth method",
			config: &GitProjectConfig{
				RepositoryURL:    "https://github.com/user/repo.git",
				MainBranch:       "main",
				WorktreeBasePath: "/tmp/test-repo",
				GitAuthMethod:    "ssh",
				GitEnabled:       true,
			},
			wantErr: true,
		},
		{
			name: "SSH URL with HTTPS auth method",
			config: &GitProjectConfig{
				RepositoryURL:    "ssh://git@github.com/user/repo.git",
				MainBranch:       "main",
				WorktreeBasePath: "/tmp/test-repo",
				GitAuthMethod:    "https",
				GitEnabled:       true,
			},
			wantErr: true,
		},
		{
			name: "relative worktree path",
			config: &GitProjectConfig{
				RepositoryURL:    "https://github.com/user/repo.git",
				MainBranch:       "main",
				WorktreeBasePath: "relative/path",
				GitAuthMethod:    "https",
				GitEnabled:       true,
			},
			wantErr: true,
		},
		{
			name: "worktree path with ..",
			config: &GitProjectConfig{
				RepositoryURL:    "https://github.com/user/repo.git",
				MainBranch:       "main",
				WorktreeBasePath: "/tmp/../test-repo",
				GitAuthMethod:    "https",
				GitEnabled:       true,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateGitProjectConfig(ctx, tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateGitProjectConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestProjectGitService_validateAuthMethodMatchesURL(t *testing.T) {
	// Setup
	executor, _ := NewDefaultCommandExecutor()
	commands := NewGitCommands(executor)
	validator := NewGitValidator(commands)
	config := &ManagerConfig{}
	manager, _ := NewGitManager(config)
	service := NewProjectGitService(validator, manager, commands)

	tests := []struct {
		name       string
		repoURL    string
		authMethod string
		wantErr    bool
	}{
		{
			name:       "HTTPS URL with HTTPS auth",
			repoURL:    "https://github.com/user/repo.git",
			authMethod: "https",
			wantErr:    false,
		},
		{
			name:       "HTTPS URL with SSH auth",
			repoURL:    "https://github.com/user/repo.git",
			authMethod: "ssh",
			wantErr:    true,
		},
		{
			name:       "SSH URL with SSH auth",
			repoURL:    "ssh://git@github.com/user/repo.git",
			authMethod: "ssh",
			wantErr:    false,
		},
		{
			name:       "SSH URL with HTTPS auth",
			repoURL:    "ssh://git@github.com/user/repo.git",
			authMethod: "https",
			wantErr:    true,
		},
		{
			name:       "SSH URL with @ and : with SSH auth",
			repoURL:    "git@github.com:user/repo.git",
			authMethod: "ssh",
			wantErr:    false,
		},
		{
			name:       "SSH URL with @ and : with HTTPS auth",
			repoURL:    "git@github.com:user/repo.git",
			authMethod: "https",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateAuthMethodMatchesURL(tt.repoURL, tt.authMethod)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateAuthMethodMatchesURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestProjectGitService_validateWorktreeBasePath(t *testing.T) {
	// Setup
	executor, _ := NewDefaultCommandExecutor()
	commands := NewGitCommands(executor)
	validator := NewGitValidator(commands)
	config := &ManagerConfig{}
	manager, _ := NewGitManager(config)
	service := NewProjectGitService(validator, manager, commands)

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "empty path",
			path:    "",
			wantErr: true,
		},
		{
			name:    "absolute path",
			path:    "/tmp/test-repo",
			wantErr: false,
		},
		{
			name:    "relative path",
			path:    "relative/path",
			wantErr: true,
		},
		{
			name:    "path with ..",
			path:    "/tmp/../test-repo",
			wantErr: true,
		},
		{
			name:    "path with .. in middle",
			path:    "/tmp/test/../repo",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateWorktreeBasePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateWorktreeBasePath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestProjectGitService_isDirectoryEmpty(t *testing.T) {
	// Setup
	executor, _ := NewDefaultCommandExecutor()
	commands := NewGitCommands(executor)
	validator := NewGitValidator(commands)
	config := &ManagerConfig{}
	manager, _ := NewGitManager(config)
	service := NewProjectGitService(validator, manager, commands)

	// Test with non-existent directory (should return true)
	if !service.isDirectoryEmpty("/non/existent/path") {
		t.Error("isDirectoryEmpty() should return true for non-existent directory")
	}

	// Note: Testing with actual directories would require file system operations
	// which are better suited for integration tests
}
