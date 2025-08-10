package github

import (
	"testing"
)

func TestNewGitHubServiceV2(t *testing.T) {
	config := &GitHubConfig{
		Token:     "test-token",
		BaseURL:   "https://api.github.com",
		UserAgent: "test-agent",
		Timeout:   30,
	}

	service := NewGitHubServiceV2(config)
	if service == nil {
		t.Fatal("Expected service to be created, got nil")
	}

	if service.config != config {
		t.Error("Expected config to be set correctly")
	}

	if service.client == nil {
		t.Error("Expected GitHub client to be created")
	}

	if service.rateLimiter == nil {
		t.Error("Expected rate limiter to be created")
	}
}

func TestGitHubServiceV2_ValidateRepository(t *testing.T) {
	service := &GitHubServiceV2{}

	tests := []struct {
		name    string
		repo    string
		wantErr bool
	}{
		{"valid repo", "owner/repo", false},
		{"empty repo", "", true},
		{"invalid format", "invalid", true},
		{"missing owner", "/repo", true},
		{"missing repo", "owner/", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateRepository(tt.repo)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateRepository() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGitHubServiceV2_IsValidMergeMethod(t *testing.T) {
	service := &GitHubServiceV2{}

	tests := []struct {
		name   string
		method string
		want   bool
	}{
		{"merge method", "merge", true},
		{"squash method", "squash", true},
		{"rebase method", "rebase", true},
		{"invalid method", "invalid", false},
		{"empty method", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := service.isValidMergeMethod(tt.method); got != tt.want {
				t.Errorf("isValidMergeMethod() = %v, want %v", got, tt.want)
			}
		})
	}
}
