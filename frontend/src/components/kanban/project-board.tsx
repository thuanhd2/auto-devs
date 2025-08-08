import { useState, useEffect, useCallback } from 'react'
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
import { BoardToolbar } from './board-toolbar'
import { KanbanBoard } from './kanban-board'
import { TaskDetailSheet } from './task-detail-sheet'
import { TaskFormModal } from './task-form-modal'

interface ProjectBoardProps {
  projectId: string
}

export function ProjectBoard({ projectId }: ProjectBoardProps) {
  const [filters, setFilters] = useState<TaskFilters>({})
  const [searchQuery, setSearchQuery] = useState('')
  const [isCompactView, setIsCompactView] = useState(false)
  const [localTasks, setLocalTasks] = useState<Task[]>([])

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

  const {
    data: tasksResponse,
    isLoading,
    refetch,
  } = useTasks(projectId, filters)
  const deleteTaskMutation = useDeleteTask()
  const duplicateTaskMutation = useDuplicateTask()
  const startPlanningMutation = useStartPlanning()
  const approvePlanAndStartImplementMutation = useApprovePlan()
  // WebSocket integration
  const { setCurrentProjectId } = useWebSocketProject(projectId)
  const { isConnected, subscribe, unsubscribe } = useWebSocketContext()
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

  // Set current project for WebSocket subscriptions
  useEffect(() => {
    if (projectId) {
      setCurrentProjectId(projectId)
    }
  }, [projectId, setCurrentProjectId])

  // Handle real-time task updates from WebSocket
  const handleTaskCreated = useCallback(
    (task: Task) => {
      if (task.project_id === projectId) {
        setLocalTasks((prev) => {
          // Check if task already exists to avoid duplicates
          if (prev.some((t) => t.id === task.id)) {
            return prev
          }
          return [...prev, task]
        })
      }
    },
    [projectId]
  )

  const handleTaskUpdated = useCallback(
    (task: Task, changes?: any) => {
      if (task.project_id === projectId) {
        // Confirm any pending optimistic updates for this task
        taskOptimisticUpdates.confirmTaskUpdate(task.id, task)

        setLocalTasks((prev) => prev.map((t) => (t.id === task.id ? task : t)))
      }
    },
    [projectId]
  )

  const handleTaskDeleted = useCallback((taskId: string) => {
    setLocalTasks((prev) => {
      const task = prev.find((t) => t.id === taskId)
      if (task) {
        return prev.filter((t) => t.id !== taskId)
      }
      return prev
    })
  }, [])

  const handleDeleteTask = async (taskId: string) => {
    if (confirm('Are you sure you want to delete this task?')) {
      const task = localTasks.find((t) => t.id === taskId)
      if (!task) return

      // Apply optimistic delete
      const updateId = taskOptimisticUpdates.deleteTask(
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

  const handleRefresh = () => {
    refetch()
    toast.success('Board refreshed')
  }

  const handleCreateTask = () => {
    setTaskFormModal({ open: true, mode: 'create', task: null })
  }

  const handleEditTask = (task: Task) => {
    setTaskFormModal({ open: true, mode: 'edit', task })
  }

  const handleViewTaskDetails = (task: Task) => {
    setTaskDetailSheet({ open: true, task })
  }

  const handleStartPlanning = async (taskId: string, branchName: string) => {
    try {
      await startPlanningMutation.mutateAsync({
        taskId,
        request: { branch_name: branchName },
      })
    } catch (error) {
      // Error is handled by the mutation hook
    }
  }

  const handleApprovePlanAndStartImplement = async (taskId: string) => {
    if (
      confirm(
        'Are you sure you want to approve the plan and start implementing?'
      )
    ) {
      try {
        await approvePlanAndStartImplementMutation.mutateAsync(taskId)
      } catch (error) {
        // Error is handled by the mutation hook
      }
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
          isCompactView={isCompactView}
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
        onOpenChange={(open) =>
          setTaskDetailSheet((prev) => ({ ...prev, open }))
        }
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
