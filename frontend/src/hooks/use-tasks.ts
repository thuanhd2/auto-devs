import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { tasksApi } from '@/lib/api/tasks'
import type {
  Task,

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
    onMutate: async (newTask) => {
      // Cancel any outgoing refetches
      await queryClient.cancelQueries({ queryKey: [TASKS_QUERY_KEY, newTask.project_id] })

      // Snapshot the previous value
      const previousTasks = queryClient.getQueryData([TASKS_QUERY_KEY, newTask.project_id])

      // Optimistically update to the new value
      const optimisticTask = {
        ...newTask,
        id: `temp-${Date.now()}`, // Temporary ID
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      }

      queryClient.setQueryData([TASKS_QUERY_KEY, newTask.project_id], (old: any) => {
        if (!old) return { tasks: [optimisticTask], total: 1 }
        return {
          ...old,
          tasks: [optimisticTask, ...old.tasks],
          total: old.total + 1,
        }
      })

      // Return a context object with the snapshotted value
      return { previousTasks, optimisticTask }
    },
    onError: (error: any, newTask, context) => {
      // If the mutation fails, use the context returned from onMutate to roll back
      if (context?.previousTasks) {
        queryClient.setQueryData([TASKS_QUERY_KEY, newTask.project_id], context.previousTasks)
      }
      toast.error(error.response?.data?.message || 'Failed to create task')
    },
    onSuccess: (newTask, variables, context) => {
      // Replace the optimistic task with the real one
      queryClient.setQueryData([TASKS_QUERY_KEY, newTask.project_id], (old: any) => {
        if (!old) return old
        return {
          ...old,
          tasks: old.tasks.map((task: Task) => 
            task.id === context?.optimisticTask.id ? newTask : task
          ),
        }
      })
      toast.success('Task created successfully')
    },
    onSettled: (newTask) => {
      // Always refetch after error or success to ensure consistency
      if (newTask) {
        queryClient.invalidateQueries({
          queryKey: [TASKS_QUERY_KEY, newTask.project_id],
        })
      }
    },
  })
}

export function useUpdateTask() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ taskId, updates }: { taskId: string; updates: UpdateTaskRequest }) =>
      tasksApi.updateTask(taskId, updates),
    onMutate: async ({ taskId, updates }) => {
      // Cancel any outgoing refetches
      await queryClient.cancelQueries({ queryKey: [TASKS_QUERY_KEY, taskId] })

      // Snapshot the previous task
      const previousTask = queryClient.getQueryData([TASKS_QUERY_KEY, taskId])
      
      // Get the project from any tasks query that might contain this task
      const tasksQueries = queryClient.getQueriesData({ queryKey: [TASKS_QUERY_KEY] })
      let projectId: string | null = null
      let previousTasks: any = null

      for (const [queryKey, queryData] of tasksQueries) {
        if (Array.isArray(queryKey) && queryKey.length >= 2 && queryData && typeof queryData === 'object') {
          const data = queryData as any
          if (data.tasks) {
            const task = data.tasks.find((t: Task) => t.id === taskId)
            if (task) {
              projectId = task.project_id
              previousTasks = queryData
              break
            }
          }
        }
      }

      // Optimistically update individual task
      if (previousTask) {
        const updatedTask = { ...previousTask as Task, ...updates, updated_at: new Date().toISOString() }
        queryClient.setQueryData([TASKS_QUERY_KEY, taskId], updatedTask)
      }

      // Optimistically update task in tasks list
      if (projectId && previousTasks) {
        queryClient.setQueryData([TASKS_QUERY_KEY, projectId], (old: any) => {
          if (!old) return old
          return {
            ...old,
            tasks: old.tasks.map((task: Task) =>
              task.id === taskId 
                ? { ...task, ...updates, updated_at: new Date().toISOString() }
                : task
            ),
          }
        })
      }

      return { previousTask, previousTasks, projectId }
    },
    onError: (error: any, { taskId }, context) => {
      // Roll back optimistic updates
      if (context?.previousTask) {
        queryClient.setQueryData([TASKS_QUERY_KEY, taskId], context.previousTask)
      }
      if (context?.previousTasks && context?.projectId) {
        queryClient.setQueryData([TASKS_QUERY_KEY, context.projectId], context.previousTasks)
      }
      toast.error(error.response?.data?.message || 'Failed to update task')
    },
    onSuccess: (updatedTask) => {
      // Update both individual task and tasks list with server response
      queryClient.setQueryData([TASKS_QUERY_KEY, updatedTask.id], updatedTask)
      
      queryClient.setQueryData([TASKS_QUERY_KEY, updatedTask.project_id], (old: any) => {
        if (!old) return old
        return {
          ...old,
          tasks: old.tasks.map((task: Task) =>
            task.id === updatedTask.id ? updatedTask : task
          ),
        }
      })
      
      toast.success('Task updated successfully')
    },
    onSettled: (updatedTask, error, { taskId }) => {
      // Always refetch to ensure consistency
      queryClient.invalidateQueries({ queryKey: [TASKS_QUERY_KEY, taskId] })
      if (updatedTask) {
        queryClient.invalidateQueries({ queryKey: [TASKS_QUERY_KEY, updatedTask.project_id] })
      }
    },
  })
}

export function useDeleteTask() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: tasksApi.deleteTask,
    onMutate: async (taskId: string) => {
      // Cancel any outgoing refetches
      await queryClient.cancelQueries({ queryKey: [TASKS_QUERY_KEY, taskId] })

      // Snapshot the previous task
      const previousTask = queryClient.getQueryData([TASKS_QUERY_KEY, taskId])
      
      // Find and snapshot the tasks list that contains this task
      const tasksQueries = queryClient.getQueriesData({ queryKey: [TASKS_QUERY_KEY] })
      let projectId: string | null = null
      let previousTasks: any = null

      for (const [queryKey, queryData] of tasksQueries) {
        if (Array.isArray(queryKey) && queryKey.length >= 2 && queryData && typeof queryData === 'object') {
          const data = queryData as any
          if (data.tasks) {
            const task = data.tasks.find((t: Task) => t.id === taskId)
            if (task) {
              projectId = task.project_id
              previousTasks = queryData
              break
            }
          }
        }
      }

      // Optimistically remove task from tasks list
      if (projectId && previousTasks) {
        queryClient.setQueryData([TASKS_QUERY_KEY, projectId], (old: any) => {
          if (!old) return old
          return {
            ...old,
            tasks: old.tasks.filter((task: Task) => task.id !== taskId),
            total: Math.max(0, old.total - 1),
          }
        })
      }

      // Remove individual task from cache
      queryClient.removeQueries({ queryKey: [TASKS_QUERY_KEY, taskId] })

      return { previousTask, previousTasks, projectId }
    },
    onError: (error: any, taskId, context) => {
      // Roll back the deletion
      if (context?.previousTask) {
        queryClient.setQueryData([TASKS_QUERY_KEY, taskId], context.previousTask)
      }
      if (context?.previousTasks && context?.projectId) {
        queryClient.setQueryData([TASKS_QUERY_KEY, context.projectId], context.previousTasks)
      }
      toast.error(error.response?.data?.message || 'Failed to delete task')
    },
    onSuccess: () => {
      toast.success('Task deleted successfully')
    },
    onSettled: (_, error, taskId, context) => {
      // Always refetch to ensure consistency
      if (context?.projectId) {
        queryClient.invalidateQueries({ queryKey: [TASKS_QUERY_KEY, context.projectId] })
      }
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

export function useDuplicateTask() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (task: Task) => {
      // Create a new task with similar data but different title
      const duplicatedTask = {
        project_id: task.project_id,
        title: `${task.title} (Copy)`,
        description: task.description,
        status: 'TODO' as Task['status'], // Reset to TODO
        plan: task.plan,
        branch_name: '', // Reset branch name
        pr_url: '', // Reset PR URL
      }
      return tasksApi.createTask(duplicatedTask)
    },
    onSuccess: (newTask) => {
      // Invalidate tasks list for the project
      queryClient.invalidateQueries({
        queryKey: [TASKS_QUERY_KEY, newTask.project_id],
      })
      toast.success('Task duplicated successfully')
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.message || 'Failed to duplicate task')
    },
  })
}