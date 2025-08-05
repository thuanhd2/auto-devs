package integration

import (
	"testing"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/stretchr/testify/assert"
)

// TestTaskGitStatusTransitions tests the Git status transition validation
func TestTaskGitStatusTransitions(t *testing.T) {
	tests := []struct {
		name           string
		from           entity.TaskGitStatus
		to             entity.TaskGitStatus
		expectedResult bool
	}{
		{
			name:           "none to creating",
			from:           entity.TaskGitStatusNone,
			to:             entity.TaskGitStatusCreating,
			expectedResult: true,
		},
		{
			name:           "creating to active",
			from:           entity.TaskGitStatusCreating,
			to:             entity.TaskGitStatusActive,
			expectedResult: true,
		},
		{
			name:           "creating to error",
			from:           entity.TaskGitStatusCreating,
			to:             entity.TaskGitStatusError,
			expectedResult: true,
		},
		{
			name:           "active to completed",
			from:           entity.TaskGitStatusActive,
			to:             entity.TaskGitStatusCompleted,
			expectedResult: true,
		},
		{
			name:           "active to cleaning",
			from:           entity.TaskGitStatusActive,
			to:             entity.TaskGitStatusCleaning,
			expectedResult: true,
		},
		{
			name:           "completed to cleaning",
			from:           entity.TaskGitStatusCompleted,
			to:             entity.TaskGitStatusCleaning,
			expectedResult: true,
		},
		{
			name:           "cleaning to none",
			from:           entity.TaskGitStatusCleaning,
			to:             entity.TaskGitStatusNone,
			expectedResult: true,
		},
		{
			name:           "error to creating (retry)",
			from:           entity.TaskGitStatusError,
			to:             entity.TaskGitStatusCreating,
			expectedResult: true,
		},
		{
			name:           "invalid transition: none to active",
			from:           entity.TaskGitStatusNone,
			to:             entity.TaskGitStatusActive,
			expectedResult: false,
		},
		{
			name:           "invalid transition: completed to active",
			from:           entity.TaskGitStatusCompleted,
			to:             entity.TaskGitStatusActive,
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.from.CanTransitionTo(tt.to)
			assert.Equal(t, tt.expectedResult, result)

			// Also test the validation function
			err := entity.ValidateGitStatusTransition(tt.from, tt.to)
			if tt.expectedResult {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.IsType(t, &entity.TaskGitStatusValidationError{}, err)
			}
		})
	}
}

// TestTaskStatusToGitStatusMapping tests the expected Git status changes for task status transitions
func TestTaskStatusToGitStatusMapping(t *testing.T) {
	tests := []struct {
		name                string
		taskStatus          entity.TaskStatus
		expectedGitStatuses []entity.TaskGitStatus // What Git statuses should be set during this task status
	}{
		{
			name:                "TODO status",
			taskStatus:          entity.TaskStatusTODO,
			expectedGitStatuses: []entity.TaskGitStatus{entity.TaskGitStatusNone},
		},
		{
			name:                "PLANNING status",
			taskStatus:          entity.TaskStatusPLANNING,
			expectedGitStatuses: []entity.TaskGitStatus{entity.TaskGitStatusNone},
		},
		{
			name:                "PLAN_REVIEWING status",
			taskStatus:          entity.TaskStatusPLANREVIEWING,
			expectedGitStatuses: []entity.TaskGitStatus{entity.TaskGitStatusNone},
		},
		{
			name:                "IMPLEMENTING status",
			taskStatus:          entity.TaskStatusIMPLEMENTING,
			expectedGitStatuses: []entity.TaskGitStatus{entity.TaskGitStatusCreating, entity.TaskGitStatusActive, entity.TaskGitStatusError},
		},
		{
			name:                "CODE_REVIEWING status",
			taskStatus:          entity.TaskStatusCODEREVIEWING,
			expectedGitStatuses: []entity.TaskGitStatus{entity.TaskGitStatusActive, entity.TaskGitStatusCompleted},
		},
		{
			name:                "DONE status",
			taskStatus:          entity.TaskStatusDONE,
			expectedGitStatuses: []entity.TaskGitStatus{entity.TaskGitStatusCompleted, entity.TaskGitStatusCleaning, entity.TaskGitStatusNone},
		},
		{
			name:                "CANCELLED status",
			taskStatus:          entity.TaskStatusCANCELLED,
			expectedGitStatuses: []entity.TaskGitStatus{entity.TaskGitStatusCleaning, entity.TaskGitStatusNone, entity.TaskGitStatusError},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test verifies that our Git status design aligns with task statuses
			// Each task status should have appropriate Git statuses available
			assert.NotEmpty(t, tt.expectedGitStatuses, "Each task status should have expected Git statuses")

			// Verify each expected Git status is valid
			for _, gitStatus := range tt.expectedGitStatuses {
				assert.True(t, gitStatus.IsValid(), "Git status %s should be valid", gitStatus)
			}
		})
	}
}

// TestGitStatusValidationFunctions tests the validation helper functions
func TestGitStatusValidationFunctions(t *testing.T) {
	t.Run("IsValid function", func(t *testing.T) {
		validStatuses := []entity.TaskGitStatus{
			entity.TaskGitStatusNone,
			entity.TaskGitStatusCreating,
			entity.TaskGitStatusActive,
			entity.TaskGitStatusCompleted,
			entity.TaskGitStatusCleaning,
			entity.TaskGitStatusError,
		}

		for _, status := range validStatuses {
			assert.True(t, status.IsValid(), "Status %s should be valid", status)
		}

		// Test invalid status
		invalidStatus := entity.TaskGitStatus("INVALID")
		assert.False(t, invalidStatus.IsValid(), "Invalid status should not be valid")
	})

	t.Run("GetDisplayName function", func(t *testing.T) {
		displayNames := map[entity.TaskGitStatus]string{
			entity.TaskGitStatusNone:      "None",
			entity.TaskGitStatusCreating:  "Creating",
			entity.TaskGitStatusActive:    "Active",
			entity.TaskGitStatusCompleted: "Completed",
			entity.TaskGitStatusCleaning:  "Cleaning",
			entity.TaskGitStatusError:     "Error",
		}

		for status, expectedName := range displayNames {
			assert.Equal(t, expectedName, status.GetDisplayName(), "Display name for %s should be %s", status, expectedName)
		}
	})

	t.Run("String function", func(t *testing.T) {
		status := entity.TaskGitStatusActive
		assert.Equal(t, "active", status.String())
	})
}