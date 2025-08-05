package jobs

import (
	"testing"
	"time"

	"github.com/auto-devs/auto-devs/internal/usecase"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockClient is a mock implementation of the actual Client
type MockClient struct {
	mock.Mock
}

func (m *MockClient) EnqueueTaskPlanningString(payload *TaskPlanningPayload, delay time.Duration) (string, error) {
	args := m.Called(payload, delay)
	return args.String(0), args.Error(1)
}

func (m *MockClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestJobClientAdapter_EnqueueTaskPlanning(t *testing.T) {
	// Setup
	mockClient := &MockClient{}
	adapter := NewJobClientAdapter(mockClient)

	taskID := uuid.New()
	projectID := uuid.New()
	branchName := "feature/test-branch"
	expectedJobID := "job-123"

	// Create payload
	payload := &usecase.TaskPlanningPayload{
		TaskID:     taskID,
		BranchName: branchName,
		ProjectID:  projectID,
	}

	// Setup expectations
	mockClient.On("EnqueueTaskPlanningString", mock.AnythingOfType("*jobs.TaskPlanningPayload"), time.Duration(0)).Return(expectedJobID, nil)

	// Execute
	jobID, err := adapter.EnqueueTaskPlanning(payload, 0)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedJobID, jobID)

	// Verify that the mock was called with correct payload
	mockClient.AssertCalled(t, "EnqueueTaskPlanningString", mock.MatchedBy(func(p *TaskPlanningPayload) bool {
		return p.TaskID == taskID &&
			p.BranchName == branchName &&
			p.ProjectID == projectID
	}), time.Duration(0))
}
