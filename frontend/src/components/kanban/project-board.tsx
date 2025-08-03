import { useState, useEffect, useCallback } from 'react'
import { toast } from 'sonner'
import { KanbanBoard } from './kanban-board'
import { BoardToolbar } from './board-toolbar'
import { BoardFilters } from './board-filters'
import { TaskFormModal } from './task-form-modal'
import { TaskDetailsModal } from './task-details-modal'
import { useTasks, useDeleteTask } from '@/hooks/use-tasks'
import { useWebSocketProject, useWebSocketContext } from '@/context/websocket-context'
import { taskOptimisticUpdates } from '@/services/optimisticUpdates'
import type { Task, TaskFilters } from '@/types/task'

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
  
  const [taskDetailsModal, setTaskDetailsModal] = useState<{
    open: boolean
    task?: Task | null
  }>({ open: false, task: null })

  const { data: tasksResponse, isLoading, refetch } = useTasks(projectId, filters)
  const deleteTaskMutation = useDeleteTask()
  
  // WebSocket integration
  const { setCurrentProjectId } = useWebSocketProject(projectId)
  const { isConnected } = useWebSocketContext()

  const tasks = tasksResponse?.tasks || []

  // Keep local tasks in sync with server data
  useEffect(() => {
    setLocalTasks(tasks)
  }, [tasks])

  // Set current project for WebSocket subscriptions
  useEffect(() => {
    if (projectId) {
      setCurrentProjectId(projectId)
    }
  }, [projectId, setCurrentProjectId])

  // Handle real-time task updates from WebSocket
  const handleTaskCreated = useCallback((task: Task) => {
    if (task.project_id === projectId) {
      setLocalTasks(prev => {
        // Check if task already exists to avoid duplicates
        if (prev.some(t => t.id === task.id)) {
          return prev
        }
        return [...prev, task]
      })
      toast.success(`New task created: ${task.title}`)
    }
  }, [projectId])

  const handleTaskUpdated = useCallback((task: Task, changes?: any) => {
    if (task.project_id === projectId) {
      setLocalTasks(prev => prev.map(t => t.id === task.id ? task : t))
      
      if (changes?.status) {
        toast.info(`Task "${task.title}" moved to ${changes.status.new}`)
      } else {
        toast.info(`Task "${task.title}" updated`)
      }
    }
  }, [projectId])

  const handleTaskDeleted = useCallback((taskId: string) => {
    setLocalTasks(prev => {
      const task = prev.find(t => t.id === taskId)
      if (task) {
        toast.info(`Task "${task.title}" deleted`)
        return prev.filter(t => t.id !== taskId)
      }
      return prev
    })
  }, [])

  const handleDeleteTask = async (taskId: string) => {
    if (confirm('Are you sure you want to delete this task?')) {
      const task = localTasks.find(t => t.id === taskId)
      if (!task) return

      // Apply optimistic delete
      const updateId = taskOptimisticUpdates.deleteTask(
        taskId,
        task,
        () => setLocalTasks(prev => prev.filter(t => t.id !== taskId)),
        () => {
          // Task deletion confirmed by server
          console.log('Task deletion confirmed')
        },
        (originalTask) => {
          // Revert deletion if failed
          if (originalTask) {
            setLocalTasks(prev => [...prev, originalTask])
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
    setTaskDetailsModal({ open: true, task })
  }

  return (
    <div className="h-full flex flex-col">
      <BoardToolbar
        onCreateTask={handleCreateTask}
        onRefresh={handleRefresh}
        isCompactView={isCompactView}
        onToggleCompactView={() => setIsCompactView(!isCompactView)}
        isLoading={isLoading}
      />
      
      <BoardFilters
        filters={filters}
        onFiltersChange={setFilters}
        searchQuery={searchQuery}
        onSearchChange={setSearchQuery}
        taskCount={tasks.length}
      />
      
      <div className="flex-1">
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
        onOpenChange={(open) => setTaskFormModal(prev => ({ ...prev, open }))}
        projectId={projectId}
        task={taskFormModal.task}
        mode={taskFormModal.mode}
      />

      <TaskDetailsModal
        open={taskDetailsModal.open}
        onOpenChange={(open) => setTaskDetailsModal(prev => ({ ...prev, open }))}
        task={taskDetailsModal.task}
        onEdit={handleEditTask}
        onDelete={handleDeleteTask}
      />
    </div>
  )
}