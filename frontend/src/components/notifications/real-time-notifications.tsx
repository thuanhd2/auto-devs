import { useEffect, useRef } from 'react'
import type { Task } from '@/types/task'
import { toast } from 'sonner'
import { useTaskAnimations } from '@/utils/animations'
import { useWebSocketContext } from '@/context/websocket-context'

interface RealTimeNotificationsProps {
  projectId?: string
  enableSound?: boolean
  enableBrowserNotifications?: boolean
  enableToastNotifications?: boolean
}

/**
 * Component that handles real-time notifications for task and project updates
 */
export function RealTimeNotifications({
  projectId,
  enableSound = false,
  enableBrowserNotifications = false,
  enableToastNotifications = true,
}: RealTimeNotificationsProps) {
  const audioRef = useRef<HTMLAudioElement | null>(null)
  const {
    animateTaskCreated,
    animateTaskUpdated,
    animateTaskStatusChanged,
    animateTaskDeleted,
    animateColumnCountUpdate,
  } = useTaskAnimations()

  // Request browser notification permission
  useEffect(() => {
    if (enableBrowserNotifications && 'Notification' in window) {
      if (Notification.permission === 'default') {
        Notification.requestPermission()
      }
    }
  }, [enableBrowserNotifications])

  // Initialize audio for sound notifications
  useEffect(() => {
    if (enableSound) {
      audioRef.current = new Audio('/notification-sound.mp3')
      audioRef.current.volume = 0.3
    }
  }, [enableSound])

  const playNotificationSound = () => {
    if (enableSound && audioRef.current) {
      audioRef.current.currentTime = 0
      audioRef.current.play().catch(console.warn)
    }
  }

  const showBrowserNotification = (
    title: string,
    body: string,
    icon?: string
  ) => {
    if (
      enableBrowserNotifications &&
      'Notification' in window &&
      Notification.permission === 'granted'
    ) {
      new Notification(title, {
        body,
        icon: icon || '/favicon.ico',
        badge: '/favicon.ico',
        tag: 'task-update',
        requireInteraction: false,
      })
    }
  }

  const showToastNotification = (
    message: string,
    type: 'success' | 'info' | 'warning' | 'error' = 'info'
  ) => {
    if (enableToastNotifications) {
      toast[type](message)
    }
  }

  // Handle task created
  const handleTaskCreated = (task: Task) => {
    if (!projectId || task.project_id === projectId) {
      const message = `New task created: "${task.title}"`

      showToastNotification(message, 'success')
      showBrowserNotification('New Task', message)
      playNotificationSound()

      // Animate the new task
      setTimeout(() => {
        animateTaskCreated(task.id, {
          showToast: false, // Already shown above
        })
        animateColumnCountUpdate(task.status)
      }, 100)
    }
  }

  // Handle task updated
  const handleTaskUpdated = (task: Task, changes?: any) => {
    console.log('handleTaskUpdated !!!!!!!!', task, changes)
    if (!projectId || task.project_id === projectId) {
      let message = `Task "${task.title}" updated`

      if (changes?.status) {
        const { old: oldStatus, new: newStatus } = changes.status
        message = `Task "${task.title}" moved from ${oldStatus} to ${newStatus}`

        showToastNotification(message, 'info')
        animateTaskStatusChanged(task.id, newStatus, {
          showToast: false,
        })

        // Update column counts
        animateColumnCountUpdate(oldStatus)
        animateColumnCountUpdate(newStatus)
      } else {
        showToastNotification(message, 'info')
        animateTaskUpdated(task.id, {
          showToast: false,
        })
      }

      showBrowserNotification('Task Updated', message)
      playNotificationSound()
    }
  }

  // Handle task deleted
  const handleTaskDeleted = (taskId: string) => {
    const message = 'Task deleted'

    showToastNotification(message, 'info')
    showBrowserNotification('Task Deleted', message)
    playNotificationSound()

    // Animate task removal
    animateTaskDeleted(taskId, {
      showToast: false,
    })
  }

  // Handle project updated
  const handleProjectUpdated = (project: any) => {
    if (!projectId || project.id === projectId) {
      const message = `Project "${project.name}" updated`

      showToastNotification(message, 'info')
      showBrowserNotification('Project Updated', message)
      playNotificationSound()
    }
  }

  // Handle user presence
  const handleUserJoined = (username: string, userProjectId: string) => {
    if (!projectId || userProjectId === projectId) {
      const message = `${username} joined the project`

      showToastNotification(message, 'info')
    }
  }

  const handleUserLeft = (username: string, userProjectId: string) => {
    if (!projectId || userProjectId === projectId) {
      const message = `${username} left the project`

      showToastNotification(message, 'info')
    }
  }

  // Handle connection errors
  const handleConnectionError = (error: string) => {
    showToastNotification(`Connection error: ${error}`, 'error')
  }

  // Handle auth failures
  const handleAuthRequired = () => {
    showToastNotification(
      'Authentication required. Please log in again.',
      'warning'
    )
  }

  return (
    <WebSocketNotificationHandler
      onTaskCreated={handleTaskCreated}
      onTaskUpdated={handleTaskUpdated}
      onTaskDeleted={handleTaskDeleted}
      onProjectUpdated={handleProjectUpdated}
      onUserJoined={handleUserJoined}
      onUserLeft={handleUserLeft}
      onConnectionError={handleConnectionError}
      onAuthRequired={handleAuthRequired}
    />
  )
}

/**
 * Internal component that registers WebSocket event handlers
 */
function WebSocketNotificationHandler({
  onTaskCreated,
  onTaskUpdated,
  onTaskDeleted,
  onProjectUpdated,
  onUserJoined,
  onUserLeft,
  onConnectionError,
  onAuthRequired,
}: {
  onTaskCreated?: (task: Task) => void
  onTaskUpdated?: (task: Task, changes?: any) => void
  onTaskDeleted?: (taskId: string) => void
  onProjectUpdated?: (project: any) => void
  onUserJoined?: (username: string, projectId: string) => void
  onUserLeft?: (username: string, projectId: string) => void
  onConnectionError?: (error: string) => void
  onAuthRequired?: () => void
}) {
  const { subscribe, unsubscribe } = useWebSocketContext()

  useEffect(() => {
    const handlers = [
      {
        type: 'task_created',
        handler: (message: any) => onTaskCreated?.(message.data),
      },
      {
        type: 'task_updated',
        handler: (message: any) =>
          onTaskUpdated?.(message.data.task, message.data.changes),
      },
      {
        type: 'task_deleted',
        handler: (message: any) =>
          onTaskDeleted?.(message.data.task_id || message.data.id),
      },
      {
        type: 'project_updated',
        handler: (message: any) => onProjectUpdated?.(message.data.project),
      },
      {
        type: 'user_joined',
        handler: (message: any) =>
          onUserJoined?.(message.data.username, message.data.project_id),
      },
      {
        type: 'user_left',
        handler: (message: any) =>
          onUserLeft?.(message.data.username, message.data.project_id),
      },
      {
        type: 'error',
        handler: (message: any) =>
          onConnectionError?.(message.data.message || message.data.error),
      },
      {
        type: 'auth_required',
        handler: () => onAuthRequired?.(),
      },
      {
        type: 'auth_failed',
        handler: () => onAuthRequired?.(),
      },
    ]

    // Subscribe to all handlers
    handlers.forEach(({ type, handler }) => {
      subscribe(type, handler)
    })

    // Cleanup on unmount
    return () => {
      handlers.forEach(({ type, handler }) => {
        unsubscribe(type, handler)
      })
    }
  }, [
    subscribe,
    unsubscribe,
    onTaskCreated,
    onTaskUpdated,
    onTaskDeleted,
    onProjectUpdated,
    onUserJoined,
    onUserLeft,
    onConnectionError,
    onAuthRequired,
  ])

  return null
}
