import { useState } from 'react'
import { toast } from 'sonner'
import { KanbanBoard } from './kanban-board'
import { BoardToolbar } from './board-toolbar'
import { BoardFilters } from './board-filters'
import { TaskFormModal } from './task-form-modal'
import { TaskDetailsModal } from './task-details-modal'
import { useTasks, useDeleteTask } from '@/hooks/use-tasks'
import type { Task, TaskFilters } from '@/types/task'

interface ProjectBoardProps {
  projectId: string
}

export function ProjectBoard({ projectId }: ProjectBoardProps) {
  const [filters, setFilters] = useState<TaskFilters>({})
  const [searchQuery, setSearchQuery] = useState('')
  const [isCompactView, setIsCompactView] = useState(false)
  
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

  const tasks = tasksResponse?.tasks || []

  const handleDeleteTask = async (taskId: string) => {
    if (confirm('Are you sure you want to delete this task?')) {
      try {
        await deleteTaskMutation.mutateAsync(taskId)
      } catch (error) {
        // Error is handled by the mutation
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
          tasks={tasks}
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