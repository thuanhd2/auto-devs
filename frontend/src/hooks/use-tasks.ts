import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { tasksApi } from '@/lib/api/tasks'
import type {
  Task,
  CreateTaskRequest,
  UpdateTaskRequest,
  TaskFilters,
} from '@/types/task'

export const TASKS_QUERY_KEY = 'tasks'

export function useTasks(projectId: string, filters?: TaskFilters) {
  return useQuery({
    queryKey: [TASKS_QUERY_KEY, projectId, filters],
    queryFn: () => tasksApi.getTasks(projectId, filters),
    enabled: !!projectId,
  })
}

export function useTask(taskId: string) {
  return useQuery({
    queryKey: [TASKS_QUERY_KEY, taskId],
    queryFn: () => tasksApi.getTask(taskId),
    enabled: !!taskId,
  })
}

export function useCreateTask() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: tasksApi.createTask,
    onSuccess: (newTask) => {
      // Invalidate tasks list for the project
      queryClient.invalidateQueries({
        queryKey: [TASKS_QUERY_KEY, newTask.project_id],
      })
      toast.success('Task created successfully')
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.message || 'Failed to create task')
    },
  })
}

export function useUpdateTask() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ taskId, updates }: { taskId: string; updates: UpdateTaskRequest }) =>
      tasksApi.updateTask(taskId, updates),
    onSuccess: (updatedTask) => {
      // Update individual task query
      queryClient.setQueryData([TASKS_QUERY_KEY, updatedTask.id], updatedTask)
      
      // Invalidate tasks list for the project
      queryClient.invalidateQueries({
        queryKey: [TASKS_QUERY_KEY, updatedTask.project_id],
      })
      
      toast.success('Task updated successfully')
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.message || 'Failed to update task')
    },
  })
}

export function useDeleteTask() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: tasksApi.deleteTask,
    onSuccess: (_, taskId) => {
      // Remove task from cache
      queryClient.removeQueries({ queryKey: [TASKS_QUERY_KEY, taskId] })
      
      // Invalidate all tasks queries
      queryClient.invalidateQueries({ queryKey: [TASKS_QUERY_KEY] })
      
      toast.success('Task deleted successfully')
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.message || 'Failed to delete task')
    },
  })
}

// Optimistic update for drag and drop
export function useOptimisticTaskUpdate() {
  const queryClient = useQueryClient()

  return (projectId: string, taskId: string, newStatus: Task['status']) => {
    queryClient.setQueryData([TASKS_QUERY_KEY, projectId], (old: any) => {
      if (!old) return old
      
      return {
        ...old,
        tasks: old.tasks.map((task: Task) =>
          task.id === taskId ? { ...task, status: newStatus } : task
        ),
      }
    })
  }
}