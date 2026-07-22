package usecase

import (
	"context"
	"fmt"
	"testing"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/auto-devs/auto-devs/internal/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newKanbanTestUsecase(t *testing.T) (*taskUsecase, *repository.TaskRepositoryMock, *JobClientInterfaceMock) {
	taskRepo := repository.NewTaskRepositoryMock(t)
	jobClient := NewJobClientInterfaceMock(t)
	uc := &taskUsecase{
		taskRepo:  taskRepo,
		jobClient: jobClient,
	}
	return uc, taskRepo, jobClient
}

func kanbanTestTask(id uuid.UUID, status entity.TaskStatus, kanbanTaskID *string) *entity.Task {
	return &entity.Task{
		ID:           id,
		ProjectID:    uuid.New(),
		Title:        "Test task",
		Status:       status,
		KanbanTaskID: kanbanTaskID,
	}
}

func TestUpdateStatus_EnqueuesKanbanNotify(t *testing.T) {
	uc, taskRepo, jobClient := newKanbanTestUsecase(t)
	taskID := uuid.New()
	kanbanID := "card-123"

	oldTask := kanbanTestTask(taskID, entity.TaskStatusIMPLEMENTING, &kanbanID)
	newTask := kanbanTestTask(taskID, entity.TaskStatusCODEREVIEWING, &kanbanID)

	taskRepo.EXPECT().GetByID(context.Background(), taskID).Return(oldTask, nil).Once()
	taskRepo.EXPECT().UpdateStatus(context.Background(), taskID, entity.TaskStatusCODEREVIEWING).Return(nil).Once()
	taskRepo.EXPECT().GetByID(context.Background(), taskID).Return(newTask, nil).Once()

	jobClient.EXPECT().EnqueueKanbanNotify(&KanbanNotifyPayload{
		TaskID:       taskID,
		KanbanTaskID: kanbanID,
		OldStatus:    entity.TaskStatusIMPLEMENTING,
		NewStatus:    entity.TaskStatusCODEREVIEWING,
	}).Return("job-1", nil).Once()

	task, err := uc.UpdateStatus(context.Background(), taskID, entity.TaskStatusCODEREVIEWING)
	require.NoError(t, err)
	assert.Equal(t, entity.TaskStatusCODEREVIEWING, task.Status)
}

func TestUpdateStatus_NoEnqueueForNonCallbackStatus(t *testing.T) {
	uc, taskRepo, _ := newKanbanTestUsecase(t)
	taskID := uuid.New()
	kanbanID := "card-123"

	oldTask := kanbanTestTask(taskID, entity.TaskStatusTODO, &kanbanID)
	newTask := kanbanTestTask(taskID, entity.TaskStatusPLANNING, &kanbanID)

	taskRepo.EXPECT().GetByID(context.Background(), taskID).Return(oldTask, nil).Once()
	taskRepo.EXPECT().UpdateStatus(context.Background(), taskID, entity.TaskStatusPLANNING).Return(nil).Once()
	taskRepo.EXPECT().GetByID(context.Background(), taskID).Return(newTask, nil).Once()

	_, err := uc.UpdateStatus(context.Background(), taskID, entity.TaskStatusPLANNING)
	require.NoError(t, err)
	// jobClient mock asserts no unexpected calls on cleanup
}

func TestUpdateStatus_NoEnqueueWithoutKanbanTaskID(t *testing.T) {
	uc, taskRepo, _ := newKanbanTestUsecase(t)
	taskID := uuid.New()

	oldTask := kanbanTestTask(taskID, entity.TaskStatusIMPLEMENTING, nil)
	newTask := kanbanTestTask(taskID, entity.TaskStatusCODEREVIEWING, nil)

	taskRepo.EXPECT().GetByID(context.Background(), taskID).Return(oldTask, nil).Once()
	taskRepo.EXPECT().UpdateStatus(context.Background(), taskID, entity.TaskStatusCODEREVIEWING).Return(nil).Once()
	taskRepo.EXPECT().GetByID(context.Background(), taskID).Return(newTask, nil).Once()

	_, err := uc.UpdateStatus(context.Background(), taskID, entity.TaskStatusCODEREVIEWING)
	require.NoError(t, err)
}

func TestUpdateStatus_NoEnqueueWhenStatusUnchanged(t *testing.T) {
	uc, taskRepo, _ := newKanbanTestUsecase(t)
	taskID := uuid.New()
	kanbanID := "card-123"

	task := kanbanTestTask(taskID, entity.TaskStatusDONE, &kanbanID)

	taskRepo.EXPECT().GetByID(context.Background(), taskID).Return(task, nil).Times(2)
	taskRepo.EXPECT().UpdateStatus(context.Background(), taskID, entity.TaskStatusDONE).Return(nil).Once()

	_, err := uc.UpdateStatus(context.Background(), taskID, entity.TaskStatusDONE)
	require.NoError(t, err)
}

func TestUpdateStatus_EnqueueFailureDoesNotFailTransition(t *testing.T) {
	uc, taskRepo, jobClient := newKanbanTestUsecase(t)
	taskID := uuid.New()
	kanbanID := "card-123"

	oldTask := kanbanTestTask(taskID, entity.TaskStatusCODEREVIEWING, &kanbanID)
	newTask := kanbanTestTask(taskID, entity.TaskStatusDONE, &kanbanID)

	taskRepo.EXPECT().GetByID(context.Background(), taskID).Return(oldTask, nil).Once()
	taskRepo.EXPECT().UpdateStatus(context.Background(), taskID, entity.TaskStatusDONE).Return(nil).Once()
	taskRepo.EXPECT().GetByID(context.Background(), taskID).Return(newTask, nil).Once()

	jobClient.EXPECT().EnqueueKanbanNotify(&KanbanNotifyPayload{
		TaskID:       taskID,
		KanbanTaskID: kanbanID,
		OldStatus:    entity.TaskStatusCODEREVIEWING,
		NewStatus:    entity.TaskStatusDONE,
	}).Return("", fmt.Errorf("redis down")).Once()

	task, err := uc.UpdateStatus(context.Background(), taskID, entity.TaskStatusDONE)
	require.NoError(t, err)
	assert.Equal(t, entity.TaskStatusDONE, task.Status)
}

func TestUpdateStatusWithHistory_EnqueuesKanbanNotify(t *testing.T) {
	uc, taskRepo, jobClient := newKanbanTestUsecase(t)
	taskID := uuid.New()
	kanbanID := "card-456"

	oldTask := kanbanTestTask(taskID, entity.TaskStatusPLANNING, &kanbanID)
	newTask := kanbanTestTask(taskID, entity.TaskStatusPLANREVIEWING, &kanbanID)

	taskRepo.EXPECT().GetByID(context.Background(), taskID).Return(oldTask, nil).Once()
	taskRepo.EXPECT().UpdateStatusWithHistory(context.Background(), taskID, entity.TaskStatusPLANREVIEWING, (*string)(nil), (*string)(nil)).Return(nil).Once()
	taskRepo.EXPECT().GetByID(context.Background(), taskID).Return(newTask, nil).Once()

	jobClient.EXPECT().EnqueueKanbanNotify(&KanbanNotifyPayload{
		TaskID:       taskID,
		KanbanTaskID: kanbanID,
		OldStatus:    entity.TaskStatusPLANNING,
		NewStatus:    entity.TaskStatusPLANREVIEWING,
	}).Return("job-2", nil).Once()

	task, err := uc.UpdateStatusWithHistory(context.Background(), UpdateStatusRequest{
		TaskID: taskID,
		Status: entity.TaskStatusPLANREVIEWING,
	})
	require.NoError(t, err)
	assert.Equal(t, entity.TaskStatusPLANREVIEWING, task.Status)
}
