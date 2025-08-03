import { websocketService } from './websocketService'

export interface OptimisticUpdate<T = any> {
  id: string
  type: 'create' | 'update' | 'delete'
  entityType: string
  entityId: string
  data: T
  originalData?: T
  timestamp: Date
  timeout: number
  onConfirm?: (data: T) => void
  onRevert?: (originalData?: T) => void
  onTimeout?: () => void
}

export interface OptimisticUpdateOptions<T = any> {
  timeout?: number
  onConfirm?: (data: T) => void
  onRevert?: (originalData?: T) => void
  onTimeout?: () => void
}

export class OptimisticUpdateManager {
  private pendingUpdates: Map<string, OptimisticUpdate> = new Map()
  private timeouts: Map<string, NodeJS.Timeout> = new Map()
  private defaultTimeout: number = 10000 // 10 seconds

  constructor(defaultTimeout = 10000) {
    this.defaultTimeout = defaultTimeout
  }

  // Apply an optimistic update
  applyUpdate<T>(
    entityType: string,
    entityId: string,
    type: 'create' | 'update' | 'delete',
    data: T,
    originalData?: T,
    options: OptimisticUpdateOptions<T> = {}
  ): string {
    const updateId = this.generateUpdateId()
    console.log('applyUpdate!!!', updateId)
    const timeout = options.timeout || this.defaultTimeout

    const update: OptimisticUpdate<T> = {
      id: updateId,
      type,
      entityType,
      entityId,
      data,
      originalData,
      timestamp: new Date(),
      timeout,
      onConfirm: options.onConfirm,
      onRevert: options.onRevert,
      onTimeout: options.onTimeout,
    }

    this.pendingUpdates.set(updateId, update)

    // Set timeout for automatic revert
    const timeoutHandle = setTimeout(() => {
      this.revertUpdate(updateId, 'timeout')
    }, timeout)

    this.timeouts.set(updateId, timeoutHandle)

    console.log(
      `Applied optimistic update ${updateId} for ${entityType}:${entityId}`
    )
    return updateId
  }

  // Confirm an optimistic update (call when server confirms)
  confirmUpdate(updateId: string, confirmedData?: any): boolean {
    const update = this.pendingUpdates.get(updateId)
    if (!update) {
      return false
    }

    this.clearTimeout(updateId)
    this.pendingUpdates.delete(updateId)

    if (update.onConfirm) {
      update.onConfirm(confirmedData || update.data)
    }

    console.log(`Confirmed optimistic update ${updateId}`)
    return true
  }

  // Revert an optimistic update
  revertUpdate(
    updateId: string,
    reason: 'manual' | 'conflict' | 'timeout' | 'error' = 'manual'
  ): boolean {
    const update = this.pendingUpdates.get(updateId)
    if (!update) {
      return false
    }

    this.clearTimeout(updateId)
    this.pendingUpdates.delete(updateId)

    if (update.onRevert) {
      update.onRevert(update.originalData)
    }

    if (reason === 'timeout' && update.onTimeout) {
      update.onTimeout()
    }

    console.log(`Reverted optimistic update ${updateId} (reason: ${reason})`)
    return true
  }

  // Get pending update by ID
  getUpdate(updateId: string): OptimisticUpdate | undefined {
    return this.pendingUpdates.get(updateId)
  }

  // Get all pending updates for an entity
  getUpdatesForEntity(
    entityType: string,
    entityId: string
  ): OptimisticUpdate[] {
    return Array.from(this.pendingUpdates.values()).filter(
      (update) =>
        update.entityType === entityType && update.entityId === entityId
    )
  }

  // Get all pending updates
  getAllPendingUpdates(): OptimisticUpdate[] {
    return Array.from(this.pendingUpdates.values())
  }

  // Check if an entity has pending updates
  hasPendingUpdates(entityType: string, entityId: string): boolean {
    return this.getUpdatesForEntity(entityType, entityId).length > 0
  }

  // Clear all pending updates
  clearAll(): void {
    this.timeouts.forEach((timeout) => clearTimeout(timeout))
    this.timeouts.clear()
    this.pendingUpdates.clear()
    console.log('Cleared all optimistic updates')
  }

  // Get count of pending updates
  getPendingCount(): number {
    return this.pendingUpdates.size
  }

  private clearTimeout(updateId: string): void {
    const timeoutHandle = this.timeouts.get(updateId)
    if (timeoutHandle) {
      clearTimeout(timeoutHandle)
      this.timeouts.delete(updateId)
    }
  }

  private generateUpdateId(): string {
    return `opt_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
  }
}

// Task-specific optimistic update helpers
export class TaskOptimisticUpdates {
  private updateManager: OptimisticUpdateManager

  constructor(updateManager: OptimisticUpdateManager) {
    this.updateManager = updateManager
  }

  createTask(
    task: any,
    onLocalUpdate: (task: any) => void,
    onConfirm?: (confirmedTask: any) => void,
    onRevert?: () => void
  ): string {
    // Apply optimistic create
    onLocalUpdate(task)

    return this.updateManager.applyUpdate(
      'task',
      task.id,
      'create',
      task,
      undefined,
      {
        onConfirm: (confirmedTask) => {
          // Update with server-confirmed data
          if (confirmedTask && confirmedTask.id !== task.id) {
            onLocalUpdate(confirmedTask)
          }
          onConfirm?.(confirmedTask)
        },
        onRevert: () => {
          // Remove the optimistically created task
          onRevert?.()
        },
      }
    )
  }

  updateTask(
    taskId: string,
    updates: Partial<any>,
    originalTask: any,
    onLocalUpdate: (updatedTask: any) => void,
    onConfirm?: (confirmedTask: any) => void,
    onRevert?: (originalTask: any) => void
  ): string {
    const optimisticTask = { ...originalTask, ...updates }

    // Apply optimistic update
    onLocalUpdate(optimisticTask)

    return this.updateManager.applyUpdate(
      'task',
      taskId,
      'update',
      optimisticTask,
      originalTask,
      {
        onConfirm: (confirmedTask) => {
          // Update with server-confirmed data
          if (confirmedTask) {
            onLocalUpdate(confirmedTask)
          }
          onConfirm?.(confirmedTask)
        },
        onRevert: (original) => {
          // Revert to original data
          if (original) {
            onLocalUpdate(original)
          }
          onRevert?.(original)
        },
      }
    )
  }

  deleteTask(
    taskId: string,
    task: any,
    onLocalUpdate: () => void,
    onConfirm?: () => void,
    onRevert?: (task: any) => void
  ): string {
    // Apply optimistic delete
    onLocalUpdate()

    return this.updateManager.applyUpdate(
      'task',
      taskId,
      'delete',
      null,
      task,
      {
        onConfirm: () => {
          onConfirm?.()
        },
        onRevert: (originalTask) => {
          // Restore the deleted task
          if (originalTask) {
            onRevert?.(originalTask)
          }
        },
      }
    )
  }

  updateTaskStatus(
    taskId: string,
    newStatus: string,
    originalTask: any,
    onLocalUpdate: (updatedTask: any) => void,
    onConfirm?: (confirmedTask: any) => void,
    onRevert?: (originalTask: any) => void
  ): string {
    return this.updateTask(
      taskId,
      { status: newStatus },
      originalTask,
      onLocalUpdate,
      onConfirm,
      onRevert
    )
  }
}

// Project-specific optimistic update helpers
export class ProjectOptimisticUpdates {
  private updateManager: OptimisticUpdateManager

  constructor(updateManager: OptimisticUpdateManager) {
    this.updateManager = updateManager
  }

  updateProject(
    projectId: string,
    updates: Partial<any>,
    originalProject: any,
    onLocalUpdate: (updatedProject: any) => void,
    onConfirm?: (confirmedProject: any) => void,
    onRevert?: (originalProject: any) => void
  ): string {
    const optimisticProject = { ...originalProject, ...updates }

    // Apply optimistic update
    onLocalUpdate(optimisticProject)

    return this.updateManager.applyUpdate(
      'project',
      projectId,
      'update',
      optimisticProject,
      originalProject,
      {
        onConfirm: (confirmedProject) => {
          if (confirmedProject) {
            onLocalUpdate(confirmedProject)
          }
          onConfirm?.(confirmedProject)
        },
        onRevert: (original) => {
          if (original) {
            onLocalUpdate(original)
          }
          onRevert?.(original)
        },
      }
    )
  }
}

// Conflict resolution helpers
export class ConflictResolver {
  static resolveTaskConflict(
    localTask: any,
    serverTask: any,
    strategy: 'server-wins' | 'client-wins' | 'merge-latest' = 'server-wins'
  ): any {
    switch (strategy) {
      case 'server-wins':
        return serverTask
      case 'client-wins':
        return localTask
      case 'merge-latest':
        // Merge based on timestamps or other criteria
        const localTime = new Date(localTask.updated_at || localTask.created_at)
        const serverTime = new Date(
          serverTask.updated_at || serverTask.created_at
        )
        return serverTime > localTime
          ? serverTask
          : { ...serverTask, ...localTask }
      default:
        return serverTask
    }
  }

  static detectConflicts(localData: any, serverData: any): string[] {
    const conflicts: string[] = []

    if (!localData || !serverData) {
      return conflicts
    }

    // Check for timestamp conflicts
    const localTime = new Date(localData.updated_at || 0)
    const serverTime = new Date(serverData.updated_at || 0)

    if (Math.abs(localTime.getTime() - serverTime.getTime()) > 1000) {
      conflicts.push('timestamp_mismatch')
    }

    // Check for field-level conflicts
    const fields = ['title', 'description', 'status', 'assignee', 'priority']
    fields.forEach((field) => {
      if (localData[field] !== serverData[field]) {
        conflicts.push(`${field}_conflict`)
      }
    })

    return conflicts
  }
}

// Export singleton instances
export const optimisticUpdateManager = new OptimisticUpdateManager()
export const taskOptimisticUpdates = new TaskOptimisticUpdates(
  optimisticUpdateManager
)
export const projectOptimisticUpdates = new ProjectOptimisticUpdates(
  optimisticUpdateManager
)
