import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import type {
  Task,
  UpdateTaskRequest,
  TaskFilters,
  StartPlanningRequest,
} from '@/types/task'
import { toast } from 'sonner'
import { tasksApi } from '@/lib/api/tasks'

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
    mutationFn: ({
      taskId,
      updates,
    }: {
      taskId: string
      updates: UpdateTaskRequest
    }) => tasksApi.updateTask(taskId, updates),
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

export function useStartPlanning() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({
      taskId,
      request,
    }: {
      taskId: string
      request: StartPlanningRequest
    }) => tasksApi.startPlanning(taskId, request),
    onMutate: async ({ taskId, request }) => {
      // Cancel any outgoing refetches
      await queryClient.cancelQueries({ queryKey: [TASKS_QUERY_KEY] })

      // Snapshot the previous value
      const previousTasks = queryClient.getQueryData([TASKS_QUERY_KEY])

      // Optimistically update task status to PLANNING
      queryClient.setQueryData([TASKS_QUERY_KEY], (old: any) => {
        if (!old) return old
        return {
          ...old,
          tasks: old.tasks.map((task: Task) =>
            task.id === taskId
              ? { ...task, status: 'PLANNING' as Task['status'] }
              : task
          ),
        }
      })

      // Return a context object with the snapshotted value
      return { previousTasks }
    },
    onSuccess: (response, { taskId }) => {
      toast.success(`Planning started successfully. Job ID: ${response.job_id}`)
    },
    onError: (error: any, { taskId }, context) => {
      // Revert optimistic update on error
      if (context?.previousTasks) {
        queryClient.setQueryData([TASKS_QUERY_KEY], context.previousTasks)
      }
      toast.error(error.response?.data?.message || 'Failed to start planning')
    },
    onSettled: () => {
      // Always refetch after error or success
      queryClient.invalidateQueries({ queryKey: [TASKS_QUERY_KEY] })
    },
  })
}

export function useApprovePlan() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (taskId: string) => tasksApi.approvePlan(taskId),
    onMutate: async (taskId) => {
      // Cancel any outgoing refetches
      await queryClient.cancelQueries({ queryKey: [TASKS_QUERY_KEY] })

      // Snapshot the previous value
      const previousTasks = queryClient.getQueryData([TASKS_QUERY_KEY])

      // Optimistically update task status to IMPLEMENTING
      queryClient.setQueryData([TASKS_QUERY_KEY], (old: any) => {
        if (!old) return old
        return {
          ...old,
          tasks: old.tasks.map((task: Task) =>
            task.id === taskId
              ? { ...task, status: 'IMPLEMENTING' as Task['status'] }
              : task
          ),
        }
      })

      // Return a context object with the snapshotted value
      return { previousTasks }
    },
    onSuccess: (response, taskId) => {
      toast.success(
        `Plan approved! Implementation job enqueued. Job ID: ${response.job_id}`
      )
    },
    onError: (error: any, taskId, context) => {
      // Revert optimistic update on error
      if (context?.previousTasks) {
        queryClient.setQueryData([TASKS_QUERY_KEY], context.previousTasks)
      }
      toast.error(error.response?.data?.message || 'Failed to approve plan')
    },
    onSettled: () => {
      // Always refetch after error or success
      queryClient.invalidateQueries({ queryKey: [TASKS_QUERY_KEY] })
    },
  })
}
