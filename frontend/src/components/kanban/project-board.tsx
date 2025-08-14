import { useState, useEffect, useCallback } from 'react'
import { useNavigate, useSearch, useParams } from '@tanstack/react-router'
import { taskOptimisticUpdates } from '@/services/optimisticUpdates'
import { CentrifugeMessage } from '@/services/websocketService'
import type { Task, TaskFilters } from '@/types/task'
import { toast } from 'sonner'
import {
  useWebSocketProject,
  useWebSocketContext,
} from '@/context/websocket-context'
import {
  useTasks,
  useDeleteTask,
  useDuplicateTask,
  useStartPlanning,
  useApprovePlan,
} from '@/hooks/use-tasks'
import { BoardFilters } from './board-filters'
import { KanbanBoard } from './kanban-board'
import { TaskDetailSheet } from './task-detail-sheet'
import { TaskFormModal } from './task-form-modal'

interface ProjectBoardProps {
  projectId: string
}

export function ProjectBoard({ projectId }: ProjectBoardProps) {
  const navigate = useNavigate()
  const searchParams = useSearch({ strict: false }) as { taskId?: string }
  const routeParams = useParams({ strict: false }) as { taskId?: string }

  const [filters, setFilters] = useState<TaskFilters>({})
  const [searchQuery, setSearchQuery] = useState('')
  const [localTasks, setLocalTasks] = useState<Task[]>([])

  // Get taskId from either route params or search params
  const currentTaskId = routeParams.taskId || searchParams.taskId

  // Modal states
  const [taskFormModal, setTaskFormModal] = useState<{
    open: boolean
    mode: 'create' | 'edit'
    task?: Task | null
  }>({ open: false, mode: 'create', task: null })

  const [taskDetailSheet, setTaskDetailSheet] = useState<{
    open: boolean
    task?: Task | null
  }>({ open: false, task: null })

  const { data: tasksResponse, refetch } = useTasks(projectId, filters)
  const deleteTaskMutation = useDeleteTask()
  const duplicateTaskMutation = useDuplicateTask()
  const startPlanningMutation = useStartPlanning()
  const approvePlanAndStartImplementMutation = useApprovePlan()
  // WebSocket integration
  const { setCurrentProjectId } = useWebSocketProject(projectId)
  const { subscribe, unsubscribe } = useWebSocketContext()
  const onTaskUpdated = useCallback((message: CentrifugeMessage) => {
    const { task_id: taskId, project_id: projectId, changes } = message.data
    // do nothing if current project id is not the same as the task's project id
    if (projectId !== projectId) {
      return
    }
    // update the task in the local tasks array
    setLocalTasks((prev) =>
      prev.map((t) => {
        if (t.id !== taskId) {
          return t
        }
        const changedValues = {}
        for (const key in changes) {
          if (t[key] !== changes[key].new) {
            changedValues[key] = changes[key].new
          }
        }
        return { ...t, ...changedValues }
      })
    )
  }, [])
  useEffect(() => {
    subscribe('task_updated', onTaskUpdated)
    return () => {
      unsubscribe('task_updated', onTaskUpdated)
    }
  }, [subscribe, unsubscribe, onTaskUpdated])

  // Keep local tasks in sync with server data
  useEffect(() => {
    // console.log('tasks changed', tasks)
    const tasks = tasksResponse?.tasks || []
    setLocalTasks(tasks)
  }, [tasksResponse])

  // Auto-open task detail sheet when taskId is in URL
  useEffect(() => {
    if (currentTaskId && localTasks.length > 0) {
      const task = localTasks.find((t) => t.id === currentTaskId)
      if (task) {
        setTaskDetailSheet({ open: true, task })
      } else {
        // Task not found, navigate back to project
        navigate({
          to: '/projects/$projectId',
          params: { projectId },
          replace: true,
        })
      }
    }
  }, [currentTaskId, localTasks, navigate, projectId])

  // Set current project for WebSocket subscriptions
  useEffect(() => {
    if (projectId) {
      setCurrentProjectId(projectId)
    }
  }, [projectId, setCurrentProjectId])

  // Handle real-time task updates from WebSocket

  const handleDeleteTask = async (taskId: string) => {
    if (confirm('Are you sure you want to delete this task?')) {
      const task = localTasks.find((t) => t.id === taskId)
      if (!task) return

      // Apply optimistic delete
      taskOptimisticUpdates.deleteTask(
        taskId,
        task,
        () => setLocalTasks((prev) => prev.filter((t) => t.id !== taskId)),
        () => {
          // Task deletion confirmed by server
          console.log('Task deletion confirmed')
        },
        (originalTask) => {
          // Revert deletion if failed
          if (originalTask) {
            setLocalTasks((prev) => [...prev, originalTask])
            toast.error('Failed to delete task')
          }
        }
      )

      try {
        await deleteTaskMutation.mutateAsync(taskId)
        // Confirm the optimistic update
        // Note: This will be handled by WebSocket message handler
      } catch (error) {
        // Error is handled by the mutation and optimistic update revert
      }
    }
  }

  const handleCreateTask = () => {
    setTaskFormModal({ open: true, mode: 'create', task: null })
  }

  const handleEditTask = (task: Task) => {
    setTaskFormModal({ open: true, mode: 'edit', task })
  }

  const handleViewTaskDetails = (task: Task) => {
    setTaskDetailSheet({ open: true, task })

    // Update URL to include task ID
    navigate({
      to: '/projects/$projectId/tasks/$taskId',
      params: { projectId, taskId: task.id },
      replace: true,
    })
  }

  const handleStartPlanning = async (
    taskId: string,
    branchName: string,
    aiType: string
  ) => {
    try {
      await startPlanningMutation.mutateAsync({
        taskId,
        request: { branch_name: branchName, ai_type: aiType },
      })
    } catch (error) {
      // Error is handled by the mutation hook
    }
  }

  const handleApprovePlanAndStartImplement = async (
    taskId: string,
    aiType: string
  ) => {
    try {
      await approvePlanAndStartImplementMutation.mutateAsync({
        taskId,
        request: { ai_type: aiType },
      })
    } catch (error) {
      // Error is handled by the mutation hook
    }
  }

  return (
    <div className='flex h-full flex-col'>
      {/* <BoardToolbar
        onCreateTask={handleCreateTask}
        onRefresh={handleRefresh}
        isCompactView={isCompactView}
        onToggleCompactView={() => setIsCompactView(!isCompactView)}
        isLoading={isLoading}
        projectId={projectId}
      /> */}

      <BoardFilters
        filters={filters}
        onFiltersChange={setFilters}
        searchQuery={searchQuery}
        onSearchChange={setSearchQuery}
        taskCount={localTasks.length}
      />

      <div className='flex-1'>
        <KanbanBoard
          tasks={localTasks}
          projectId={projectId}
          onCreateTask={handleCreateTask}
          onEditTask={handleEditTask}
          onDeleteTask={handleDeleteTask}
          onViewTaskDetails={handleViewTaskDetails}
          searchQuery={searchQuery}
        />
      </div>

      {/* Modals */}
      <TaskFormModal
        open={taskFormModal.open}
        onOpenChange={(open) => setTaskFormModal((prev) => ({ ...prev, open }))}
        projectId={projectId}
        task={taskFormModal.task}
        mode={taskFormModal.mode}
      />

      <TaskDetailSheet
        open={taskDetailSheet.open}
        onOpenChange={(open) => {
          setTaskDetailSheet((prev) => ({ ...prev, open }))

          // If closing the sheet, navigate back to project
          if (!open) {
            navigate({
              to: '/projects/$projectId',
              params: { projectId },
              replace: true,
            })
          }
        }}
        task={taskDetailSheet.task || null}
        onEdit={handleEditTask}
        onDelete={handleDeleteTask}
        onDuplicate={async (task) => {
          try {
            await duplicateTaskMutation.mutateAsync(task)
          } catch (error) {
            // Error is handled by the mutation hook
          }
        }}
        onStatusChange={(taskId, newStatus) => {
          // Find the task to apply optimistic update
          const task = localTasks.find((t) => t.id === taskId)
          if (task) {
            // Apply optimistic status change with WebSocket confirmation
            taskOptimisticUpdates.updateTaskStatus(
              taskId,
              newStatus,
              task,
              (updatedTask) => {
                setLocalTasks((prev) =>
                  prev.map((t) => (t.id === taskId ? updatedTask : t))
                )
              },
              (confirmedTask) => {
                // Status change confirmed by WebSocket
                console.log(
                  'Status change confirmed by WebSocket:',
                  confirmedTask
                )
              },
              (originalTask) => {
                // Revert status change if failed
                setLocalTasks((prev) =>
                  prev.map((t) => (t.id === taskId ? originalTask : t))
                )
                toast.error('Failed to update task status')
              }
            )
          }
        }}
        onStartPlanning={handleStartPlanning}
        onApprovePlanAndStartImplement={handleApprovePlanAndStartImplement}
      />
    </div>
  )
}
