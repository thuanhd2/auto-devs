package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTaskStatus_IsValid(t *testing.T) {
	validStatuses := []TaskStatus{
		TaskStatusTODO,
		TaskStatusPLANNING,
		TaskStatusPLANREVIEWING,
		TaskStatusIMPLEMENTING,
		TaskStatusCODEREVIEWING,
		TaskStatusDONE,
		TaskStatusCANCELLED,
	}

	for _, status := range validStatuses {
		t.Run(string(status), func(t *testing.T) {
			assert.True(t, status.IsValid(), "Status %s should be valid", status)
		})
	}

	invalidStatuses := []TaskStatus{
		"INVALID",
		"",
		"todo", // lowercase
		"RANDOM_STATUS",
	}

	for _, status := range invalidStatuses {
		t.Run(string(status), func(t *testing.T) {
			assert.False(t, status.IsValid(), "Status %s should be invalid", status)
		})
	}
}

func TestTaskStatus_String(t *testing.T) {
	status := TaskStatusTODO
	assert.Equal(t, "TODO", status.String())
}

func TestTaskStatus_GetDisplayName(t *testing.T) {
	testCases := []struct {
		status      TaskStatus
		displayName string
	}{
		{TaskStatusTODO, "To Do"},
		{TaskStatusPLANNING, "Planning"},
		{TaskStatusPLANREVIEWING, "Plan Review"},
		{TaskStatusIMPLEMENTING, "Implementing"},
		{TaskStatusCODEREVIEWING, "Code Review"},
		{TaskStatusDONE, "Done"},
		{TaskStatusCANCELLED, "Cancelled"},
		{"INVALID", "INVALID"}, // fallback case
	}

	for _, tc := range testCases {
		t.Run(string(tc.status), func(t *testing.T) {
			assert.Equal(t, tc.displayName, tc.status.GetDisplayName())
		})
	}
}

func TestTaskStatus_CanTransitionTo(t *testing.T) {
	testCases := []struct {
		from     TaskStatus
		to       TaskStatus
		expected bool
	}{
		// Valid transitions from TODO
		{TaskStatusTODO, TaskStatusPLANNING, true},
		{TaskStatusTODO, TaskStatusCANCELLED, true},
		{TaskStatusTODO, TaskStatusIMPLEMENTING, false}, // Skip planning
		
		// Valid transitions from PLANNING
		{TaskStatusPLANNING, TaskStatusPLANREVIEWING, true},
		{TaskStatusPLANNING, TaskStatusTODO, true}, // Back to TODO
		{TaskStatusPLANNING, TaskStatusCANCELLED, true},
		{TaskStatusPLANNING, TaskStatusDONE, false}, // Can't skip to done
		
		// Valid transitions from PLAN_REVIEWING
		{TaskStatusPLANREVIEWING, TaskStatusIMPLEMENTING, true},
		{TaskStatusPLANREVIEWING, TaskStatusPLANNING, true}, // Back to planning
		{TaskStatusPLANREVIEWING, TaskStatusCANCELLED, true},
		{TaskStatusPLANREVIEWING, TaskStatusTODO, false}, // Can't go back to TODO directly
		
		// Valid transitions from IMPLEMENTING
		{TaskStatusIMPLEMENTING, TaskStatusCODEREVIEWING, true},
		{TaskStatusIMPLEMENTING, TaskStatusPLANREVIEWING, true}, // Back to plan review
		{TaskStatusIMPLEMENTING, TaskStatusCANCELLED, true},
		{TaskStatusIMPLEMENTING, TaskStatusDONE, false}, // Can't skip code review
		
		// Valid transitions from CODE_REVIEWING
		{TaskStatusCODEREVIEWING, TaskStatusDONE, true},
		{TaskStatusCODEREVIEWING, TaskStatusIMPLEMENTING, true}, // Back to implementing
		{TaskStatusCODEREVIEWING, TaskStatusCANCELLED, true},
		{TaskStatusCODEREVIEWING, TaskStatusTODO, false}, // Can't go back to TODO
		
		// Valid transitions from DONE
		{TaskStatusDONE, TaskStatusTODO, true}, // Allow reopening
		{TaskStatusDONE, TaskStatusPLANNING, false}, // Can't go to planning from done
		
		// Valid transitions from CANCELLED
		{TaskStatusCANCELLED, TaskStatusTODO, true}, // Allow reactivating
		{TaskStatusCANCELLED, TaskStatusDONE, false}, // Can't go to done from cancelled
	}

	for _, tc := range testCases {
		t.Run(string(tc.from)+"_to_"+string(tc.to), func(t *testing.T) {
			result := tc.from.CanTransitionTo(tc.to)
			assert.Equal(t, tc.expected, result, 
				"Transition from %s to %s should be %v", tc.from, tc.to, tc.expected)
		})
	}
}

func TestGetAllTaskStatuses(t *testing.T) {
	statuses := GetAllTaskStatuses()
	
	expectedStatuses := []TaskStatus{
		TaskStatusTODO,
		TaskStatusPLANNING,
		TaskStatusPLANREVIEWING,
		TaskStatusIMPLEMENTING,
		TaskStatusCODEREVIEWING,
		TaskStatusDONE,
		TaskStatusCANCELLED,
	}
	
	assert.Len(t, statuses, len(expectedStatuses))
	
	for _, expectedStatus := range expectedStatuses {
		assert.Contains(t, statuses, expectedStatus)
	}
}

func TestValidateStatusTransition(t *testing.T) {
	// Valid transition
	err := ValidateStatusTransition(TaskStatusTODO, TaskStatusPLANNING)
	assert.NoError(t, err)
	
	// Invalid transition
	err = ValidateStatusTransition(TaskStatusTODO, TaskStatusDONE)
	assert.Error(t, err)
	assert.IsType(t, &TaskStatusValidationError{}, err)
	
	// Invalid from status
	err = ValidateStatusTransition("INVALID", TaskStatusPLANNING)
	assert.Error(t, err)
	assert.IsType(t, &TaskStatusValidationError{}, err)
	assert.Contains(t, err.Error(), "invalid current status")
	
	// Invalid to status
	err = ValidateStatusTransition(TaskStatusTODO, "INVALID")
	assert.Error(t, err)
	assert.IsType(t, &TaskStatusValidationError{}, err)
	assert.Contains(t, err.Error(), "invalid target status")
}

func TestTaskStatusValidationError_Error(t *testing.T) {
	// Error with custom message
	err := &TaskStatusValidationError{
		CurrentStatus: TaskStatusTODO,
		TargetStatus:  TaskStatusDONE,
		Message:       "Custom error message",
	}
	assert.Equal(t, "Custom error message", err.Error())
	
	// Error without custom message
	err = &TaskStatusValidationError{
		CurrentStatus: TaskStatusTODO,
		TargetStatus:  TaskStatusDONE,
	}
	assert.Equal(t, "invalid status transition from TODO to DONE", err.Error())
}