package jobs

import (
	"context"
	"fmt"
	"log/slog"
	"testing"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/auto-devs/auto-devs/internal/usecase"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeKanbanClient records calls for assertions
type fakeKanbanClient struct {
	enabled      bool
	commentErr   error
	unblockErr   error
	comments     []string
	commentIDs   []string
	unblockedIDs []string
}

func (f *fakeKanbanClient) CommentTask(ctx context.Context, kanbanTaskID string, body string) error {
	if f.commentErr != nil {
		return f.commentErr
	}
	f.commentIDs = append(f.commentIDs, kanbanTaskID)
	f.comments = append(f.comments, body)
	return nil
}

func (f *fakeKanbanClient) UnblockTask(ctx context.Context, kanbanTaskID string) error {
	if f.unblockErr != nil {
		return f.unblockErr
	}
	f.unblockedIDs = append(f.unblockedIDs, kanbanTaskID)
	return nil
}

func (f *fakeKanbanClient) Enabled() bool {
	return f.enabled
}

func newKanbanTestProcessor(taskUsecase usecase.TaskUsecase, kanbanClient *fakeKanbanClient) *Processor {
	return &Processor{
		taskUsecase:  taskUsecase,
		kanbanClient: kanbanClient,
		logger:       slog.Default().With("component", "job-processor-test"),
	}
}

func TestProcessKanbanNotify_Success(t *testing.T) {
	taskID := uuid.New()
	kanbanID := "card-123"
	prURL := "https://github.com/user/repo/pull/42"

	taskUsecaseMock := usecase.NewTaskUsecaseMock(t)
	taskUsecaseMock.EXPECT().GetByID(context.Background(), taskID).Return(&entity.Task{
		ID:          taskID,
		Title:       "Add login page",
		Status:      entity.TaskStatusCODEREVIEWING,
		PullRequest: &prURL,
	}, nil).Once()

	kanbanClient := &fakeKanbanClient{enabled: true}
	processor := newKanbanTestProcessor(taskUsecaseMock, kanbanClient)

	job, err := NewKanbanNotifyTask(KanbanNotifyPayload{
		TaskID:       taskID,
		KanbanTaskID: kanbanID,
		OldStatus:    entity.TaskStatusIMPLEMENTING,
		NewStatus:    entity.TaskStatusCODEREVIEWING,
	})
	require.NoError(t, err)

	err = processor.ProcessKanbanNotify(context.Background(), job)
	require.NoError(t, err)

	require.Len(t, kanbanClient.comments, 1)
	comment := kanbanClient.comments[0]
	assert.Contains(t, comment, "[auto-devs] status=CODE_REVIEWING")
	assert.Contains(t, comment, fmt.Sprintf("task: %s — Add login page", taskID))
	assert.Contains(t, comment, "old_status: IMPLEMENTING")
	assert.Contains(t, comment, fmt.Sprintf("plans: GET /api/v1/tasks/%s/plans", taskID))
	assert.Contains(t, comment, "pr: "+prURL)
	assert.Contains(t, comment, "error: none")

	assert.Equal(t, []string{kanbanID}, kanbanClient.commentIDs)
	assert.Equal(t, []string{kanbanID}, kanbanClient.unblockedIDs)
}

func TestProcessKanbanNotify_CancelledIncludesError(t *testing.T) {
	taskID := uuid.New()

	taskUsecaseMock := usecase.NewTaskUsecaseMock(t)
	taskUsecaseMock.EXPECT().GetByID(context.Background(), taskID).Return(&entity.Task{
		ID:              taskID,
		Title:           "Broken task",
		Status:          entity.TaskStatusCANCELLED,
		ErrorLogEntries: []string{"first error", "quota limit exceeded"},
	}, nil).Once()

	kanbanClient := &fakeKanbanClient{enabled: true}
	processor := newKanbanTestProcessor(taskUsecaseMock, kanbanClient)

	job, err := NewKanbanNotifyTask(KanbanNotifyPayload{
		TaskID:       taskID,
		KanbanTaskID: "card-9",
		OldStatus:    entity.TaskStatusIMPLEMENTING,
		NewStatus:    entity.TaskStatusCANCELLED,
	})
	require.NoError(t, err)

	err = processor.ProcessKanbanNotify(context.Background(), job)
	require.NoError(t, err)

	require.Len(t, kanbanClient.comments, 1)
	assert.Contains(t, kanbanClient.comments[0], "error: quota limit exceeded")
	assert.Contains(t, kanbanClient.comments[0], "pr: none")
}

func TestProcessKanbanNotify_CommentErrorReturnsError(t *testing.T) {
	taskID := uuid.New()

	taskUsecaseMock := usecase.NewTaskUsecaseMock(t)
	taskUsecaseMock.EXPECT().GetByID(context.Background(), taskID).Return(&entity.Task{
		ID:     taskID,
		Title:  "Test task",
		Status: entity.TaskStatusDONE,
	}, nil).Once()

	kanbanClient := &fakeKanbanClient{enabled: true, commentErr: fmt.Errorf("dashboard down")}
	processor := newKanbanTestProcessor(taskUsecaseMock, kanbanClient)

	job, err := NewKanbanNotifyTask(KanbanNotifyPayload{
		TaskID:       taskID,
		KanbanTaskID: "card-1",
		OldStatus:    entity.TaskStatusCODEREVIEWING,
		NewStatus:    entity.TaskStatusDONE,
	})
	require.NoError(t, err)

	err = processor.ProcessKanbanNotify(context.Background(), job)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "dashboard down")
	assert.Empty(t, kanbanClient.unblockedIDs)
}

func TestProcessKanbanNotify_UnblockErrorReturnsError(t *testing.T) {
	taskID := uuid.New()

	taskUsecaseMock := usecase.NewTaskUsecaseMock(t)
	taskUsecaseMock.EXPECT().GetByID(context.Background(), taskID).Return(&entity.Task{
		ID:     taskID,
		Title:  "Test task",
		Status: entity.TaskStatusDONE,
	}, nil).Once()

	kanbanClient := &fakeKanbanClient{enabled: true, unblockErr: fmt.Errorf("409 conflict")}
	processor := newKanbanTestProcessor(taskUsecaseMock, kanbanClient)

	job, err := NewKanbanNotifyTask(KanbanNotifyPayload{
		TaskID:       taskID,
		KanbanTaskID: "card-1",
		OldStatus:    entity.TaskStatusCODEREVIEWING,
		NewStatus:    entity.TaskStatusDONE,
	})
	require.NoError(t, err)

	err = processor.ProcessKanbanNotify(context.Background(), job)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "409")
	require.Len(t, kanbanClient.comments, 1)
}

func TestProcessKanbanNotify_StaleStatusSkips(t *testing.T) {
	taskID := uuid.New()

	// Task has already moved on to DONE; the retried PLAN_REVIEWING
	// callback is stale and must be dropped without touching the card.
	taskUsecaseMock := usecase.NewTaskUsecaseMock(t)
	taskUsecaseMock.EXPECT().GetByID(context.Background(), taskID).Return(&entity.Task{
		ID:     taskID,
		Title:  "Test task",
		Status: entity.TaskStatusDONE,
	}, nil).Once()

	kanbanClient := &fakeKanbanClient{enabled: true}
	processor := newKanbanTestProcessor(taskUsecaseMock, kanbanClient)

	job, err := NewKanbanNotifyTask(KanbanNotifyPayload{
		TaskID:       taskID,
		KanbanTaskID: "card-1",
		OldStatus:    entity.TaskStatusPLANNING,
		NewStatus:    entity.TaskStatusPLANREVIEWING,
	})
	require.NoError(t, err)

	err = processor.ProcessKanbanNotify(context.Background(), job)
	require.NoError(t, err)
	assert.Empty(t, kanbanClient.comments)
	assert.Empty(t, kanbanClient.unblockedIDs)
}

func TestProcessKanbanNotify_DisabledSkips(t *testing.T) {
	kanbanClient := &fakeKanbanClient{enabled: false}
	processor := newKanbanTestProcessor(nil, kanbanClient)

	job, err := NewKanbanNotifyTask(KanbanNotifyPayload{
		TaskID:       uuid.New(),
		KanbanTaskID: "card-1",
		OldStatus:    entity.TaskStatusIMPLEMENTING,
		NewStatus:    entity.TaskStatusDONE,
	})
	require.NoError(t, err)

	err = processor.ProcessKanbanNotify(context.Background(), job)
	require.NoError(t, err)
	assert.Empty(t, kanbanClient.comments)
	assert.Empty(t, kanbanClient.unblockedIDs)
}
