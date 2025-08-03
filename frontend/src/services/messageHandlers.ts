import { WebSocketMessage } from './websocketService'

export interface TaskUpdateData {
  task: any
  changes?: Record<string, any>
}

export interface ProjectUpdateData {
  project: any
  changes?: Record<string, any>
}

export interface StatusChangeData {
  entity_type: string
  entity_id: string
  old_status: string
  new_status: string
  timestamp: string
}

export interface UserPresenceData {
  user_id: string
  username: string
  project_id: string
  timestamp: string
}

export interface MessageHandler<T = any> {
  type: string
  handler: (data: T, message: WebSocketMessage) => void
}

export class MessageHandlerRegistry {
  private handlers: Map<string, Set<(data: any, message: WebSocketMessage) => void>> = new Map()

  register<T>(type: string, handler: (data: T, message: WebSocketMessage) => void) {
    if (!this.handlers.has(type)) {
      this.handlers.set(type, new Set())
    }
    this.handlers.get(type)!.add(handler)
  }

  unregister<T>(type: string, handler: (data: T, message: WebSocketMessage) => void) {
    const handlers = this.handlers.get(type)
    if (handlers) {
      handlers.delete(handler)
      if (handlers.size === 0) {
        this.handlers.delete(type)
      }
    }
  }

  handle(message: WebSocketMessage) {
    const handlers = this.handlers.get(message.type)
    if (handlers) {
      handlers.forEach((handler) => {
        try {
          handler(message.data, message)
        } catch (error) {
          console.error(`Error handling message type ${message.type}:`, error)
        }
      })
    }
  }

  clear() {
    this.handlers.clear()
  }

  getHandlerCount(type?: string): number {
    if (type) {
      return this.handlers.get(type)?.size || 0
    }
    return Array.from(this.handlers.values()).reduce((total, handlers) => total + handlers.size, 0)
  }

  getRegisteredTypes(): string[] {
    return Array.from(this.handlers.keys())
  }
}

// Task event handlers
export const createTaskCreatedHandler = (
  onTaskCreated: (task: any) => void
) => {
  return (data: any, message: WebSocketMessage) => {
    console.log('Task created:', data)
    onTaskCreated(data)
  }
}

export const createTaskUpdatedHandler = (
  onTaskUpdated: (task: any, changes?: Record<string, any>) => void
) => {
  return (data: TaskUpdateData, message: WebSocketMessage) => {
    console.log('Task updated:', data)
    onTaskUpdated(data.task, data.changes)
  }
}

export const createTaskDeletedHandler = (
  onTaskDeleted: (taskId: string) => void
) => {
  return (data: any, message: WebSocketMessage) => {
    console.log('Task deleted:', data)
    if (data.task_id || data.id) {
      onTaskDeleted(data.task_id || data.id)
    }
  }
}

// Project event handlers
export const createProjectUpdatedHandler = (
  onProjectUpdated: (project: any, changes?: Record<string, any>) => void
) => {
  return (data: ProjectUpdateData, message: WebSocketMessage) => {
    console.log('Project updated:', data)
    onProjectUpdated(data.project, data.changes)
  }
}

// Status change handlers
export const createStatusChangedHandler = (
  onStatusChanged: (entityType: string, entityId: string, oldStatus: string, newStatus: string) => void
) => {
  return (data: StatusChangeData, message: WebSocketMessage) => {
    console.log('Status changed:', data)
    onStatusChanged(data.entity_type, data.entity_id, data.old_status, data.new_status)
  }
}

// User presence handlers
export const createUserJoinedHandler = (
  onUserJoined: (userId: string, username: string, projectId: string) => void
) => {
  return (data: UserPresenceData, message: WebSocketMessage) => {
    console.log('User joined:', data)
    onUserJoined(data.user_id, data.username, data.project_id)
  }
}

export const createUserLeftHandler = (
  onUserLeft: (userId: string, username: string, projectId: string) => void
) => {
  return (data: UserPresenceData, message: WebSocketMessage) => {
    console.log('User left:', data)
    onUserLeft(data.user_id, data.username, data.project_id)
  }
}

// System message handlers
export const createErrorHandler = (
  onError: (error: string, code?: string) => void
) => {
  return (data: any, message: WebSocketMessage) => {
    console.error('WebSocket error message:', data)
    onError(data.message || data.error || 'Unknown error', data.code)
  }
}

export const createAuthFailedHandler = (
  onAuthFailed: () => void
) => {
  return (data: any, message: WebSocketMessage) => {
    console.error('Authentication failed:', data)
    onAuthFailed()
  }
}

// Utility functions for message processing
export const extractTaskFromMessage = (message: WebSocketMessage): any | null => {
  switch (message.type) {
    case 'task_created':
      return message.data
    case 'task_updated':
      return message.data.task
    case 'task_deleted':
      return message.data
    default:
      return null
  }
}

export const extractProjectFromMessage = (message: WebSocketMessage): any | null => {
  switch (message.type) {
    case 'project_updated':
      return message.data.project || message.data
    default:
      return null
  }
}

export const extractChangesFromMessage = (message: WebSocketMessage): Record<string, any> | null => {
  switch (message.type) {
    case 'task_updated':
    case 'project_updated':
      return message.data.changes || null
    default:
      return null
  }
}

// Message validation
export const validateMessage = (message: WebSocketMessage): boolean => {
  if (!message.type || !message.timestamp || !message.message_id) {
    console.warn('Invalid message format:', message)
    return false
  }
  return true
}

export const isTaskMessage = (message: WebSocketMessage): boolean => {
  return ['task_created', 'task_updated', 'task_deleted'].includes(message.type)
}

export const isProjectMessage = (message: WebSocketMessage): boolean => {
  return ['project_updated'].includes(message.type)
}

export const isStatusMessage = (message: WebSocketMessage): boolean => {
  return message.type === 'status_changed'
}

export const isUserPresenceMessage = (message: WebSocketMessage): boolean => {
  return ['user_joined', 'user_left'].includes(message.type)
}

export const isSystemMessage = (message: WebSocketMessage): boolean => {
  return ['ping', 'pong', 'auth_required', 'auth_success', 'auth_failed', 'error'].includes(message.type)
}

// Message aggregation for bulk updates
export class MessageAggregator {
  private pendingMessages: Map<string, WebSocketMessage[]> = new Map()
  private flushTimer: NodeJS.Timeout | null = null
  private flushDelay: number = 100 // ms

  constructor(flushDelay = 100) {
    this.flushDelay = flushDelay
  }

  aggregate(message: WebSocketMessage, key: string, onFlush: (messages: WebSocketMessage[]) => void) {
    if (!this.pendingMessages.has(key)) {
      this.pendingMessages.set(key, [])
    }
    
    this.pendingMessages.get(key)!.push(message)

    // Clear existing timer
    if (this.flushTimer) {
      clearTimeout(this.flushTimer)
    }

    // Set new timer to flush messages
    this.flushTimer = setTimeout(() => {
      const messages = this.pendingMessages.get(key)
      if (messages && messages.length > 0) {
        onFlush([...messages])
        this.pendingMessages.delete(key)
      }
      this.flushTimer = null
    }, this.flushDelay)
  }

  flush(key: string, onFlush: (messages: WebSocketMessage[]) => void) {
    const messages = this.pendingMessages.get(key)
    if (messages && messages.length > 0) {
      onFlush([...messages])
      this.pendingMessages.delete(key)
    }

    if (this.flushTimer) {
      clearTimeout(this.flushTimer)
      this.flushTimer = null
    }
  }

  clear() {
    this.pendingMessages.clear()
    if (this.flushTimer) {
      clearTimeout(this.flushTimer)
      this.flushTimer = null
    }
  }
}

// Export singleton instances
export const messageHandlerRegistry = new MessageHandlerRegistry()
export const messageAggregator = new MessageAggregator()