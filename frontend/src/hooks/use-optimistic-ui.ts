import { useState, useCallback } from 'react'
import { useNotifications } from './use-notifications'

export interface OptimisticAction<TData = any, TError = any> {
  id: string
  type: 'create' | 'update' | 'delete'
  entity: string
  data: TData
  timestamp: Date
  isPending: boolean
  error?: TError
}

export interface OptimisticUIOptions {
  successMessage?: string
  errorMessage?: string
  loadingMessage?: string
  showToast?: boolean
  timeout?: number
}

export function useOptimisticUI() {
  const [actions, setActions] = useState<OptimisticAction[]>([])
  const { success, error, loading, dismiss } = useNotifications()

  const addAction = useCallback((
    type: 'create' | 'update' | 'delete',
    entity: string,
    data: any,
    options: OptimisticUIOptions = {}
  ): string => {
    const actionId = `${entity}-${type}-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`
    
    const action: OptimisticAction = {
      id: actionId,
      type,
      entity,
      data,
      timestamp: new Date(),
      isPending: true,
    }

    setActions(prev => [...prev, action])

    // Show loading toast if requested
    if (options.showToast && options.loadingMessage) {
      loading(options.loadingMessage, undefined, { timeout: options.timeout })
    }

    return actionId
  }, [loading])

  const completeAction = useCallback((
    actionId: string,
    result?: any,
    options: OptimisticUIOptions = {}
  ) => {
    setActions(prev => prev.map(action => 
      action.id === actionId 
        ? { ...action, isPending: false, data: result || action.data }
        : action
    ))

    // Show success toast if requested
    if (options.showToast && options.successMessage) {
      success(options.successMessage)
    }

    // Remove action after delay
    setTimeout(() => {
      setActions(prev => prev.filter(action => action.id !== actionId))
    }, 5000)
  }, [success])

  const failAction = useCallback((
    actionId: string,
    errorData: any,
    options: OptimisticUIOptions = {}
  ) => {
    setActions(prev => prev.map(action => 
      action.id === actionId 
        ? { ...action, isPending: false, error: errorData }
        : action
    ))

    // Show error toast if requested
    if (options.showToast && options.errorMessage) {
      error(options.errorMessage)
    }

    // Remove action after delay
    setTimeout(() => {
      setActions(prev => prev.filter(action => action.id !== actionId))
    }, 10000)
  }, [error])

  const removeAction = useCallback((actionId: string) => {
    setActions(prev => prev.filter(action => action.id !== actionId))
  }, [])

  const clearActions = useCallback(() => {
    setActions([])
  }, [])

  const getActionsByEntity = useCallback((entity: string) => {
    return actions.filter(action => action.entity === entity)
  }, [actions])

  const getPendingActions = useCallback(() => {
    return actions.filter(action => action.isPending)
  }, [actions])

  const getFailedActions = useCallback(() => {
    return actions.filter(action => action.error)
  }, [actions])

  const hasPendingAction = useCallback((entity: string, type?: 'create' | 'update' | 'delete') => {
    return actions.some(action => 
      action.entity === entity && 
      action.isPending && 
      (!type || action.type === type)
    )
  }, [actions])

  // Convenience methods for common operations
  const createOptimistic = useCallback((
    entity: string,
    data: any,
    options: OptimisticUIOptions = {}
  ) => {
    return addAction('create', entity, data, {
      loadingMessage: `Creating ${entity}...`,
      successMessage: `${entity} created successfully!`,
      errorMessage: `Failed to create ${entity}`,
      showToast: true,
      ...options,
    })
  }, [addAction])

  const updateOptimistic = useCallback((
    entity: string,
    data: any,
    options: OptimisticUIOptions = {}
  ) => {
    return addAction('update', entity, data, {
      loadingMessage: `Updating ${entity}...`,
      successMessage: `${entity} updated successfully!`,
      errorMessage: `Failed to update ${entity}`,
      showToast: true,
      ...options,
    })
  }, [addAction])

  const deleteOptimistic = useCallback((
    entity: string,
    data: any,
    options: OptimisticUIOptions = {}
  ) => {
    return addAction('delete', entity, data, {
      loadingMessage: `Deleting ${entity}...`,
      successMessage: `${entity} deleted successfully!`,
      errorMessage: `Failed to delete ${entity}`,
      showToast: true,
      ...options,
    })
  }, [addAction])

  return {
    // Core action management
    actions,
    addAction,
    completeAction,
    failAction,
    removeAction,
    clearActions,
    
    // Query methods
    getActionsByEntity,
    getPendingActions,
    getFailedActions,
    hasPendingAction,
    
    // Convenience methods
    createOptimistic,
    updateOptimistic,
    deleteOptimistic,
    
    // Statistics
    totalActions: actions.length,
    pendingCount: actions.filter(a => a.isPending).length,
    failedCount: actions.filter(a => a.error).length,
  }
}

// Hook for managing optimistic state for a specific entity
export function useEntityOptimisticState<T>(
  entity: string,
  currentData: T,
  isLoading: boolean = false
) {
  const { actions, hasPendingAction } = useOptimisticUI()
  
  const entityActions = actions.filter(action => action.entity === entity)
  const latestAction = entityActions.sort((a, b) => 
    b.timestamp.getTime() - a.timestamp.getTime()
  )[0]

  // Determine the optimistic state
  const optimisticData = latestAction && latestAction.isPending 
    ? { ...currentData, ...latestAction.data }
    : currentData

  const isPending = hasPendingAction(entity)
  const isError = latestAction?.error !== undefined

  return {
    data: optimisticData,
    isPending: isPending || isLoading,
    isError,
    error: latestAction?.error,
    lastAction: latestAction,
    actionCount: entityActions.length,
  }
}