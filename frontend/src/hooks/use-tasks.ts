import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import type {
  Task,
  UpdateTaskRequest,
  StartPlanningRequest,
  ApprovePlanRequest,
} from '@/types/task'
import { toast } from 'sonner'
import { tasksApi } from '@/lib/api/tasks'

const TASKS_QUERY_KEY = 'tasks'

export function useTasks(projectId: string) {
  return useQuery({
    queryKey: [TASKS_QUERY_KEY, projectId],
    queryFn: () => tasksApi.getTasks(projectId),
    enabled: !!projectId,
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
    onMutate: async ({ taskId }) => {
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
    onSuccess: (response) => {
      toast.success(`Planning started successfully. Job ID: ${response.job_id}`)
    },
    onError: (error: any, context: any) => {
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
    mutationFn: ({
      taskId,
      request,
    }: {
      taskId: string
      request: ApprovePlanRequest
    }) => tasksApi.approvePlan(taskId, request),
    onMutate: async (mutatedTask) => {
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
            task.id === mutatedTask.taskId
              ? { ...task, status: 'IMPLEMENTING' as Task['status'] }
              : task
          ),
        }
      })

      // Return a context object with the snapshotted value
      return { previousTasks }
    },
    onSuccess: (response) => {
      toast.success(
        `Plan approved! Implementation job enqueued. Job ID: ${response.job_id}`
      )
    },
    onError: (error: any, context: any) => {
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

export function useChangeTaskStatus() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({
      taskId,
      status,
    }: {
      taskId: string
      status: Task['status']
    }) => tasksApi.changeTaskStatus(taskId, status),
    onMutate: async ({ taskId, status }) => {
      // Cancel any outgoing refetches
      await queryClient.cancelQueries({ queryKey: [TASKS_QUERY_KEY] })

      // Snapshot the previous value
      const previousTasks = queryClient.getQueryData([TASKS_QUERY_KEY])

      // Optimistically update task status
      queryClient.setQueryData([TASKS_QUERY_KEY], (old: any) => {
        if (!old) return old
        return {
          ...old,
          tasks: old.tasks.map((task: Task) =>
            task.id === taskId ? { ...task, status } : task
          ),
        }
      })

      // Return a context object with the snapshotted value
      return { previousTasks }
    },
    onSuccess: (updatedTask) => {
      // Update individual task query
      queryClient.setQueryData([TASKS_QUERY_KEY, updatedTask.id], updatedTask)

      // Invalidate tasks list for the project
      queryClient.invalidateQueries({
        queryKey: [TASKS_QUERY_KEY, updatedTask.project_id],
      })

      toast.success('Task status updated successfully')
    },
    onError: (error: any, variables, context: any) => {
      // Revert optimistic update on error
      if (context?.previousTasks) {
        queryClient.setQueryData([TASKS_QUERY_KEY], context.previousTasks)
      }
      toast.error(error.response?.data?.message || 'Failed to update task status')
    },
    onSettled: () => {
      // Always refetch after error or success
      queryClient.invalidateQueries({ queryKey: [TASKS_QUERY_KEY] })
    },
  })
}
