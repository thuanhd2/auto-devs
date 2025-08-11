import { useState, useEffect } from 'react'
import type { Project } from '@/types/project'
import type { Task, TaskStatus } from '@/types/task'
import { useParams } from 'react-router-dom'
import { useWebSocketConnection } from '@/context/websocket-context'
import { useProjects } from '@/hooks/use-projects'
import { useTasks } from '@/hooks/use-tasks'
import { RealTimeNotifications } from '@/components/notifications/real-time-notifications'
import { RealTimeProjectStats } from '@/components/stats/real-time-project-stats'
import { GitOperationControls } from './git-operation-controls'
import { GitStatusCard } from './git-status-card'
import { KanbanBoard } from './kanban-board'
import { TaskDetailModal } from './task-detail-modal'
import { TaskEditModal } from './task-edit-modal'
import { TaskFormModal } from './task-form-modal'
import { TaskHistoryModal } from './task-history-modal'

interface ProjectBoardProps {
  projectId: string
}

export function ProjectBoard({ projectId }: ProjectBoardProps) {
  const [filters, setFilters] = useState({})
  const [searchQuery, setSearchQuery] = useState('')
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

  const { data: tasksResponse, refetch } = useTasks(projectId, filters)

  // WebSocket integration
  const { setCurrentProjectId } = useWebSocketConnection()

  // Keep local tasks in sync with server data
  useEffect(() => {
    const tasks = tasksResponse?.tasks || []
    setLocalTasks(tasks)
  }, [tasksResponse])

  // Set current project for WebSocket subscriptions
  useEffect(() => {
    if (projectId) {
      setCurrentProjectId(projectId)
    }
  }, [projectId, setCurrentProjectId])

  const handleCreateTask = () => {
    setTaskFormModal({ open: true, mode: 'create', task: null })
  }

  const handleEditTask = (task: Task) => {
    setTaskFormModal({ open: true, mode: 'edit', task })
  }

  const handleViewTaskDetails = (task: Task) => {
    setTaskDetailSheet({ open: true, task })
  }

  const handleCloseTaskFormModal = () => {
    setTaskFormModal({ open: false, mode: 'create', task: null })
  }

  const handleCloseTaskDetailSheet = () => {
    setTaskDetailSheet({ open: false, task: null })
  }

  const handleCloseTaskEditModal = () => {
    setTaskFormModal({ open: false, mode: 'create', task: null })
  }

  const handleCloseTaskHistoryModal = () => {
    // Close history modal logic
  }

  const handleTaskFormSubmit = async (taskData: Partial<Task>) => {
    try {
      await refetch()
      handleCloseTaskFormModal()
    } catch (err) {
      // Handle error
    }
  }

  const handleTaskEditSubmit = async (taskData: Partial<Task>) => {
    try {
      await refetch()
      handleCloseTaskEditModal()
    } catch (err) {
      // Handle error
    }
  }

  const handleTaskDeleted = async (taskId: string) => {
    try {
      await refetch()
    } catch (err) {
      // Handle error
    }
  }

  const handleTaskStatusChange = async (
    taskId: string,
    newStatus: TaskStatus
  ) => {
    try {
      await refetch()
    } catch (err) {
      // Handle error
    }
  }

  const handleTaskMoved = async (
    taskId: string,
    newStatus: TaskStatus,
    newIndex: number
  ) => {
    try {
      await refetch()
    } catch (err) {
      // Handle error
    }
  }

  const handleRefresh = async () => {
    try {
      await refetch()
    } catch (err) {
      // Handle error
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
