import React, { 
  createContext, 
  useContext, 
  useCallback, 
  useEffect, 
  useRef, 
  useState,
  ReactNode 
} from 'react'
import { websocketService, WebSocketMessage, ConnectionState } from '@/services/websocketService'
import { 
  messageHandlerRegistry, 
  messageAggregator,
  createTaskCreatedHandler,
  createTaskUpdatedHandler,
  createTaskDeletedHandler,
  createProjectUpdatedHandler,
  createStatusChangedHandler,
  createUserJoinedHandler,
  createUserLeftHandler,
  createErrorHandler,
  createAuthFailedHandler
} from '@/services/messageHandlers'
import { 
  optimisticUpdateManager,
  taskOptimisticUpdates,
  projectOptimisticUpdates
} from '@/services/optimisticUpdates'

export interface WebSocketContextValue {
  // Connection state
  connectionState: ConnectionState
  isConnected: boolean
  isConnecting: boolean
  isReconnecting: boolean
  lastError: string | undefined
  queuedMessageCount: number

  // Connection control
  connect: () => Promise<void>
  disconnect: () => void
  reconnect: () => Promise<void>

  // Messaging
  send: (message: any) => Promise<void>
  subscribe: (messageType: string, handler: (message: WebSocketMessage) => void) => void
  unsubscribe: (messageType: string, handler: (message: WebSocketMessage) => void) => void

  // Project subscriptions
  subscribeToProject: (projectId: string) => Promise<void>
  unsubscribeFromProject: (projectId: string) => Promise<void>
  currentProjectId: string | null
  setCurrentProjectId: (projectId: string | null) => void

  // Optimistic updates
  optimisticUpdateCount: number
  clearOptimisticUpdates: () => void

  // Message queue
  clearMessageQueue: () => void

  // User presence
  onlineUsers: Map<string, { username: string; joinedAt: Date }>
  userCount: number

  // Event handlers (can be overridden by components)
  onTaskCreated?: (task: any) => void
  onTaskUpdated?: (task: any, changes?: any) => void
  onTaskDeleted?: (taskId: string) => void
  onProjectUpdated?: (project: any, changes?: any) => void
  onStatusChanged?: (entityType: string, entityId: string, oldStatus: string, newStatus: string) => void
  onUserJoined?: (userId: string, username: string, projectId: string) => void
  onUserLeft?: (userId: string, username: string, projectId: string) => void
  onConnectionError?: (error: string) => void
  onAuthRequired?: () => void

  // Debugging
  messageHistory: WebSocketMessage[]
  clearMessageHistory: () => void
  isDebugMode: boolean
  setDebugMode: (enabled: boolean) => void
}

const WebSocketContext = createContext<WebSocketContextValue | null>(null)

export interface WebSocketProviderProps {
  children: ReactNode
  authToken?: string | null
  autoConnect?: boolean
  maxMessageHistory?: number
  onTaskCreated?: (task: any) => void
  onTaskUpdated?: (task: any, changes?: any) => void
  onTaskDeleted?: (taskId: string) => void
  onProjectUpdated?: (project: any, changes?: any) => void
  onStatusChanged?: (entityType: string, entityId: string, oldStatus: string, newStatus: string) => void
  onUserJoined?: (userId: string, username: string, projectId: string) => void
  onUserLeft?: (userId: string, username: string, projectId: string) => void
  onConnectionError?: (error: string) => void
  onAuthRequired?: () => void
}

export function WebSocketProvider({
  children,
  authToken,
  autoConnect = true,
  maxMessageHistory = 100,
  onTaskCreated,
  onTaskUpdated,
  onTaskDeleted,
  onProjectUpdated,
  onStatusChanged,
  onUserJoined,
  onUserLeft,
  onConnectionError,
  onAuthRequired,
}: WebSocketProviderProps) {
  const [connectionState, setConnectionState] = useState<ConnectionState>(
    websocketService.getConnectionState()
  )
  const [queuedMessageCount, setQueuedMessageCount] = useState(0)
  const [optimisticUpdateCount, setOptimisticUpdateCount] = useState(0)
  const [currentProjectId, setCurrentProjectId] = useState<string | null>(null)
  const [onlineUsers, setOnlineUsers] = useState<Map<string, { username: string; joinedAt: Date }>>(new Map())
  const [messageHistory, setMessageHistory] = useState<WebSocketMessage[]>([])
  const [isDebugMode, setDebugMode] = useState(false)

  const connectionListenerRef = useRef<(state: ConnectionState) => void>()
  const queueCheckIntervalRef = useRef<NodeJS.Timeout>()
  const optimisticCheckIntervalRef = useRef<NodeJS.Timeout>()

  // Connection state management
  const handleConnectionStateChange = useCallback((state: ConnectionState) => {
    setConnectionState(state)
    
    if (state.status === 'error' && state.lastError) {
      onConnectionError?.(state.lastError)
    }
  }, [onConnectionError])

  // Message history management
  const addToMessageHistory = useCallback((message: WebSocketMessage) => {
    if (isDebugMode) {
      setMessageHistory(prev => {
        const newHistory = [...prev, message]
        return newHistory.slice(-maxMessageHistory)
      })
    }
  }, [isDebugMode, maxMessageHistory])

  const clearMessageHistory = useCallback(() => {
    setMessageHistory([])
  }, [])

  // Set up message handlers
  useEffect(() => {
    const handlers = [
      // Task handlers
      {
        type: 'task_created',
        handler: createTaskCreatedHandler((task) => {
          onTaskCreated?.(task)
        })
      },
      {
        type: 'task_updated', 
        handler: createTaskUpdatedHandler((task, changes) => {
          // Check for optimistic update confirmation
          const pendingUpdates = optimisticUpdateManager.getAllPendingUpdates()
          const matchingUpdate = pendingUpdates.find(
            update => update.entityType === 'task' && update.entityId === task.id
          )
          
          if (matchingUpdate) {
            optimisticUpdateManager.confirmUpdate(matchingUpdate.id, task)
          }
          
          onTaskUpdated?.(task, changes)
        })
      },
      {
        type: 'task_deleted',
        handler: createTaskDeletedHandler((taskId) => {
          onTaskDeleted?.(taskId)
        })
      },
      
      // Project handlers
      {
        type: 'project_updated',
        handler: createProjectUpdatedHandler((project, changes) => {
          // Check for optimistic update confirmation
          const pendingUpdates = optimisticUpdateManager.getAllPendingUpdates()
          const matchingUpdate = pendingUpdates.find(
            update => update.entityType === 'project' && update.entityId === project.id
          )
          
          if (matchingUpdate) {
            optimisticUpdateManager.confirmUpdate(matchingUpdate.id, project)
          }
          
          onProjectUpdated?.(project, changes)
        })
      },

      // Status handlers
      {
        type: 'status_changed',
        handler: createStatusChangedHandler((entityType, entityId, oldStatus, newStatus) => {
          onStatusChanged?.(entityType, entityId, oldStatus, newStatus)
        })
      },

      // User presence handlers
      {
        type: 'user_joined',
        handler: createUserJoinedHandler((userId, username, projectId) => {
          if (projectId === currentProjectId) {
            setOnlineUsers(prev => new Map(prev).set(userId, { username, joinedAt: new Date() }))
          }
          onUserJoined?.(userId, username, projectId)
        })
      },
      {
        type: 'user_left',
        handler: createUserLeftHandler((userId, username, projectId) => {
          if (projectId === currentProjectId) {
            setOnlineUsers(prev => {
              const newMap = new Map(prev)
              newMap.delete(userId)
              return newMap
            })
          }
          onUserLeft?.(userId, username, projectId)
        })
      },

      // System handlers
      {
        type: 'error',
        handler: createErrorHandler((error) => {
          onConnectionError?.(error)
        })
      },
      {
        type: 'auth_failed',
        handler: createAuthFailedHandler(() => {
          onAuthRequired?.()
        })
      },
      {
        type: 'auth_required',
        handler: createAuthFailedHandler(() => {
          onAuthRequired?.()
        })
      }
    ]

    // Register all handlers
    handlers.forEach(({ type, handler }) => {
      messageHandlerRegistry.register(type, handler)
    })

    // Register wildcard handler for message history
    const wildcardHandler = (message: WebSocketMessage) => {
      addToMessageHistory(message)
    }
    messageHandlerRegistry.register('*', wildcardHandler)

    // Set up message processing
    const messageListener = (message: WebSocketMessage) => {
      messageHandlerRegistry.handle(message)
    }

    websocketService.subscribe('*', messageListener)

    return () => {
      // Cleanup handlers
      handlers.forEach(({ type, handler }) => {
        messageHandlerRegistry.unregister(type, handler)
      })
      messageHandlerRegistry.unregister('*', wildcardHandler)
      websocketService.unsubscribe('*', messageListener)
    }
  }, [
    currentProjectId,
    onTaskCreated,
    onTaskUpdated, 
    onTaskDeleted,
    onProjectUpdated,
    onStatusChanged,
    onUserJoined,
    onUserLeft,
    onConnectionError,
    onAuthRequired,
    addToMessageHistory
  ])

  // Set up connection state listener
  useEffect(() => {
    connectionListenerRef.current = handleConnectionStateChange
    websocketService.subscribeToConnectionState(handleConnectionStateChange)

    return () => {
      if (connectionListenerRef.current) {
        websocketService.unsubscribeFromConnectionState(connectionListenerRef.current)
      }
    }
  }, [handleConnectionStateChange])

  // Set up auth token
  useEffect(() => {
    websocketService.setAuthToken(authToken || null)
  }, [authToken])

  // Auto-connect
  useEffect(() => {
    if (autoConnect && authToken) {
      websocketService.connect().catch(console.error)
    }
  }, [autoConnect, authToken])

  // Set up periodic state checks
  useEffect(() => {
    const updateCounts = () => {
      setQueuedMessageCount(websocketService.getQueuedMessageCount())
      setOptimisticUpdateCount(optimisticUpdateManager.getPendingCount())
    }

    updateCounts()
    queueCheckIntervalRef.current = setInterval(updateCounts, 1000)

    return () => {
      if (queueCheckIntervalRef.current) {
        clearInterval(queueCheckIntervalRef.current)
      }
    }
  }, [])

  // Project subscription management
  const handleSetCurrentProjectId = useCallback(async (projectId: string | null) => {
    if (currentProjectId && currentProjectId !== projectId) {
      try {
        await websocketService.unsubscribeFromProject(currentProjectId)
      } catch (error) {
        console.error('Failed to unsubscribe from previous project:', error)
      }
    }

    setCurrentProjectId(projectId)
    setOnlineUsers(new Map()) // Clear user list when changing projects

    if (projectId && connectionState.status === 'connected') {
      try {
        await websocketService.subscribeToProject(projectId)
      } catch (error) {
        console.error('Failed to subscribe to new project:', error)
      }
    }
  }, [currentProjectId, connectionState.status])

  // Auto-subscribe to project when connected
  useEffect(() => {
    if (currentProjectId && connectionState.status === 'connected') {
      websocketService.subscribeToProject(currentProjectId).catch(console.error)
    }
  }, [currentProjectId, connectionState.status])

  // API methods
  const connect = useCallback(async () => {
    await websocketService.connect()
  }, [])

  const disconnect = useCallback(() => {
    websocketService.disconnect()
  }, [])

  const reconnect = useCallback(async () => {
    websocketService.disconnect()
    await websocketService.connect()
  }, [])

  const send = useCallback(async (message: any) => {
    await websocketService.send(message)
  }, [])

  const subscribe = useCallback((messageType: string, handler: (message: WebSocketMessage) => void) => {
    websocketService.subscribe(messageType, handler)
  }, [])

  const unsubscribe = useCallback((messageType: string, handler: (message: WebSocketMessage) => void) => {
    websocketService.unsubscribe(messageType, handler)
  }, [])

  const subscribeToProject = useCallback(async (projectId: string) => {
    await websocketService.subscribeToProject(projectId)
  }, [])

  const unsubscribeFromProject = useCallback(async (projectId: string) => {
    await websocketService.unsubscribeFromProject(projectId)
  }, [])

  const clearOptimisticUpdates = useCallback(() => {
    optimisticUpdateManager.clearAll()
    setOptimisticUpdateCount(0)
  }, [])

  const clearMessageQueue = useCallback(() => {
    websocketService.clearMessageQueue()
    setQueuedMessageCount(0)
  }, [])

  const contextValue: WebSocketContextValue = {
    // Connection state
    connectionState,
    isConnected: connectionState.status === 'connected',
    isConnecting: connectionState.status === 'connecting',
    isReconnecting: connectionState.isReconnecting,
    lastError: connectionState.lastError,
    queuedMessageCount,

    // Connection control
    connect,
    disconnect,
    reconnect,

    // Messaging
    send,
    subscribe,
    unsubscribe,

    // Project subscriptions
    subscribeToProject,
    unsubscribeFromProject,
    currentProjectId,
    setCurrentProjectId: handleSetCurrentProjectId,

    // Optimistic updates
    optimisticUpdateCount,
    clearOptimisticUpdates,

    // Message queue
    clearMessageQueue,

    // User presence
    onlineUsers,
    userCount: onlineUsers.size,

    // Event handlers
    onTaskCreated,
    onTaskUpdated,
    onTaskDeleted,
    onProjectUpdated,
    onStatusChanged,
    onUserJoined,
    onUserLeft,
    onConnectionError,
    onAuthRequired,

    // Debugging
    messageHistory,
    clearMessageHistory,
    isDebugMode,
    setDebugMode,
  }

  return (
    <WebSocketContext.Provider value={contextValue}>
      {children}
    </WebSocketContext.Provider>
  )
}

export function useWebSocketContext(): WebSocketContextValue {
  const context = useContext(WebSocketContext)
  if (!context) {
    throw new Error('useWebSocketContext must be used within a WebSocketProvider')
  }
  return context
}

// Helper hooks for specific use cases
export function useWebSocketConnection() {
  const { 
    isConnected, 
    isConnecting, 
    isReconnecting,
    connectionState,
    connect,
    disconnect,
    reconnect,
    lastError,
    queuedMessageCount,
    clearMessageQueue
  } = useWebSocketContext()

  return {
    isConnected,
    isConnecting,
    isReconnecting,
    connectionState,
    connect,
    disconnect,
    reconnect,
    lastError,
    queuedMessageCount,
    clearMessageQueue,
  }
}

export function useWebSocketProject(projectId?: string) {
  const {
    currentProjectId,
    setCurrentProjectId,
    subscribeToProject,
    unsubscribeFromProject,
    onlineUsers,
    userCount,
  } = useWebSocketContext()

  useEffect(() => {
    if (projectId && projectId !== currentProjectId) {
      setCurrentProjectId(projectId)
    }
  }, [projectId, currentProjectId, setCurrentProjectId])

  return {
    currentProjectId,
    setCurrentProjectId,
    subscribeToProject,
    unsubscribeFromProject,
    onlineUsers,
    userCount,
  }
}

export function useWebSocketDebug() {
  const {
    messageHistory,
    clearMessageHistory,
    isDebugMode,
    setDebugMode,
    optimisticUpdateCount,
    clearOptimisticUpdates,
  } = useWebSocketContext()

  return {
    messageHistory,
    clearMessageHistory,
    isDebugMode,
    setDebugMode,
    optimisticUpdateCount,
    clearOptimisticUpdates,
  }
}