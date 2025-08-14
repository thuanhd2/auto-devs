// Removed unused websocketService import

interface OptimisticUpdate<T = any> {
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

interface OptimisticUpdateOptions<T = any> {
  timeout?: number
  onConfirm?: (data: T) => void
  onRevert?: (originalData?: T) => void
  onTimeout?: () => void
}

class OptimisticUpdateManager {
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
class TaskOptimisticUpdates {
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

// Conflict resolution helpers
class ConflictResolver {
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

// WebSocket Integration for optimistic updates
class WebSocketOptimisticUpdateIntegrator {
  private updateManager: OptimisticUpdateManager
  private entityToUpdateMap: Map<string, Set<string>> = new Map() // entityId -> updateIds

  constructor(updateManager: OptimisticUpdateManager) {
    this.updateManager = updateManager
  }

  // Register an optimistic update for WebSocket confirmation
  registerForConfirmation(
    entityType: string,
    entityId: string,
    updateId: string
  ): void {
    const key = `${entityType}:${entityId}`
    if (!this.entityToUpdateMap.has(key)) {
      this.entityToUpdateMap.set(key, new Set())
    }
    this.entityToUpdateMap.get(key)!.add(updateId)
  }

  // Confirm updates when WebSocket message arrives
  confirmByWebSocket(
    entityType: string,
    entityId: string,
    serverData: any
  ): void {
    const key = `${entityType}:${entityId}`
    const updateIds = this.entityToUpdateMap.get(key)

    if (updateIds && updateIds.size > 0) {
      // Confirm all pending updates for this entity
      updateIds.forEach((updateId) => {
        this.updateManager.confirmUpdate(updateId, serverData)
      })

      // Clear the confirmed updates
      updateIds.clear()
    }
  }

  // Handle conflicts when WebSocket data doesn't match optimistic data
  handleConflict(
    entityType: string,
    entityId: string,
    serverData: any,
    strategy: 'revert' | 'merge' | 'ignore' = 'revert'
  ): void {
    const key = `${entityType}:${entityId}`
    const updateIds = this.entityToUpdateMap.get(key)

    if (updateIds && updateIds.size > 0) {
      updateIds.forEach((updateId) => {
        const update = this.updateManager.getUpdate(updateId)
        if (update) {
          switch (strategy) {
            case 'revert':
              this.updateManager.revertUpdate(updateId, 'conflict')
              break
            case 'merge':
              // Apply merge logic and confirm with merged data
              const mergedData = ConflictResolver.resolveTaskConflict(
                update.data,
                serverData,
                'merge-latest'
              )
              this.updateManager.confirmUpdate(updateId, mergedData)
              break
            case 'ignore':
              // Just confirm with server data
              this.updateManager.confirmUpdate(updateId, serverData)
              break
          }
        }
      })

      updateIds.clear()
    }
  }

  // Cleanup stale registrations
  cleanupStaleRegistrations(): void {
    const keysToDelete: string[] = []

    this.entityToUpdateMap.forEach((updateIds, key) => {
      // Remove confirmed or expired updates
      const validUpdateIds = Array.from(updateIds).filter(
        (updateId) => this.updateManager.getUpdate(updateId) !== undefined
      )

      if (validUpdateIds.length === 0) {
        keysToDelete.push(key)
      } else {
        this.entityToUpdateMap.set(key, new Set(validUpdateIds))
      }
    })

    keysToDelete.forEach((key) => this.entityToUpdateMap.delete(key))
  }
}

// Enhanced Task optimistic updates with WebSocket integration
class EnhancedTaskOptimisticUpdates extends TaskOptimisticUpdates {
  private integrator: WebSocketOptimisticUpdateIntegrator

  constructor(
    updateManager: OptimisticUpdateManager,
    integrator: WebSocketOptimisticUpdateIntegrator
  ) {
    super(updateManager)
    this.integrator = integrator
  }

  updateTaskStatus(
    taskId: string,
    newStatus: string,
    originalTask: any,
    onLocalUpdate: (updatedTask: any) => void,
    onConfirm?: (confirmedTask: any) => void,
    onRevert?: (originalTask: any) => void
  ): string {
    const updateId = super.updateTaskStatus(
      taskId,
      newStatus,
      originalTask,
      onLocalUpdate,
      onConfirm,
      onRevert
    )

    // Register for WebSocket confirmation
    this.integrator.registerForConfirmation('task', taskId, updateId)

    return updateId
  }

  // Handle WebSocket task update confirmation
  confirmTaskUpdate(taskId: string, serverTask: any): void {
    this.integrator.confirmByWebSocket('task', taskId, serverTask)
  }

  // Handle WebSocket task update conflicts
  handleTaskConflict(
    taskId: string,
    serverTask: any,
    strategy: 'revert' | 'merge' | 'ignore' = 'revert'
  ): void {
    this.integrator.handleConflict('task', taskId, serverTask, strategy)
  }
}

// Export enhanced singleton instances
export const optimisticUpdateManager = new OptimisticUpdateManager(15000) // Increase timeout to 15s
const wsOptimisticIntegrator = new WebSocketOptimisticUpdateIntegrator(
  optimisticUpdateManager
)
export const taskOptimisticUpdates = new EnhancedTaskOptimisticUpdates(
  optimisticUpdateManager,
  wsOptimisticIntegrator
)
