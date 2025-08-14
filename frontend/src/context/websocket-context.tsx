import {
  createContext,
  useContext,
  useCallback,
  useEffect,
  useRef,
  useState,
  ReactNode,
} from 'react'
import { optimisticUpdateManager } from '@/services/optimisticUpdates'
import {
  websocketService,
  CentrifugeMessage,
  ConnectionState,
} from '@/services/websocketService'

interface WebSocketContextValue {
  // Connection state
  connectionState: ConnectionState
  isConnected: boolean
  isConnecting: boolean
  isReconnecting: boolean
  lastError: string | undefined

  // Connection control
  connect: () => Promise<void>
  disconnect: () => void
  reconnect: () => Promise<void>

  // Project subscription
  currentProjectId: string | null
  setCurrentProjectId: (projectId: string | null) => void
  subscribeToProject: (projectId: string) => Promise<void>
  unsubscribeFromProject: (projectId: string) => Promise<void>

  // Message handling
  subscribe: (
    messageType: string,
    handler: (message: CentrifugeMessage) => void
  ) => void
  unsubscribe: (
    messageType: string,
    handler: (message: CentrifugeMessage) => void
  ) => void

  // User presence
  onlineUsers: Map<string, { username: string; joinedAt: Date }>
  userCount: number

  // Event handlers
  onTaskCreated?: (task: any) => void
  onTaskUpdated?: (task: any, changes?: any) => void
  onTaskDeleted?: (taskId: string) => void
  onProjectUpdated?: (project: any, changes?: any) => void
  onStatusChanged?: (
    entityType: string,
    entityId: string,
    oldStatus: string,
    newStatus: string
  ) => void
  onUserJoined?: (userId: string, username: string, projectId: string) => void
  onUserLeft?: (userId: string, username: string, projectId: string) => void
  onConnectionError?: (error: string) => void
  onAuthRequired?: () => void

  // Debugging
  messageHistory: CentrifugeMessage[]
  clearMessageHistory: () => void
  isDebugMode: boolean
  setDebugMode: (enabled: boolean) => void
}

const WebSocketContext = createContext<WebSocketContextValue | null>(null)

interface WebSocketProviderProps {
  children: ReactNode
  authToken?: string | null
  autoConnect?: boolean
  maxMessageHistory?: number
  onTaskCreated?: (task: any) => void
  onTaskUpdated?: (task: any, changes?: any) => void
  onTaskDeleted?: (taskId: string) => void
  onProjectUpdated?: (project: any, changes?: any) => void
  onStatusChanged?: (
    entityType: string,
    entityId: string,
    oldStatus: string,
    newStatus: string
  ) => void
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
  const [currentProjectId, setCurrentProjectId] = useState<string | null>(null)
  const [onlineUsers, setOnlineUsers] = useState<
    Map<string, { username: string; joinedAt: Date }>
  >(new Map())
  const [messageHistory, setMessageHistory] = useState<CentrifugeMessage[]>([])
  const [isDebugMode, setDebugMode] = useState(false)

  const connectionListenerRef = useRef<(state: ConnectionState) => void>()
  const messageHandlersRef = useRef<
    Map<string, Set<(message: CentrifugeMessage) => void>>
  >(new Map())

  // Connection state management
  const handleConnectionStateChange = useCallback(
    (state: ConnectionState) => {
      setConnectionState(state)

      if (state.status === 'error' && state.lastError) {
        onConnectionError?.(state.lastError)
      }
    },
    [onConnectionError]
  )

  // Message history management
  const addToMessageHistory = useCallback(
    (message: CentrifugeMessage) => {
      if (isDebugMode) {
        setMessageHistory((prev) => {
          const newHistory = [...prev, message]
          return newHistory.slice(-maxMessageHistory)
        })
      }
    },
    [isDebugMode, maxMessageHistory]
  )

  const clearMessageHistory = useCallback(() => {
    setMessageHistory([])
  }, [])

  const confirmOptimisticPendingUpdates = useCallback(
    (entityType: string, entityId: string) => {
      const pendingUpdates = optimisticUpdateManager.getAllPendingUpdates()
      const matchingUpdate = pendingUpdates.find(
        (update) =>
          update.entityType === entityType && update.entityId === entityId
      )
      if (matchingUpdate) {
        optimisticUpdateManager.confirmUpdate(matchingUpdate.id, entityId)
      }
    },
    []
  )

  // Message handling
  const handleMessage = useCallback(
    (message: CentrifugeMessage) => {
      // Add to message history if debug mode is enabled
      addToMessageHistory(message)

      // Call specific event handlers based on message type
      switch (message.type) {
        case 'task_created':
          onTaskCreated?.(message.data)
          break
        case 'task_updated':
          // Check for optimistic update confirmation
          confirmOptimisticPendingUpdates('task', message.data.task.id)
          console.log('task_updated !!!!!!!!', message.data)
          onTaskUpdated?.(message.data.task, message.data.changes)
          break
        case 'task_deleted':
          confirmOptimisticPendingUpdates('task', message.data.task_id)
          onTaskDeleted?.(message.data.task_id)
          break
        case 'project_updated':
          onProjectUpdated?.(message.data.project, message.data.changes)
          break
        case 'status_changed':
          const { entity_type, entity_id, old_status, new_status } =
            message.data
          console.log('task status_changed !!!!!!!!', message.data)
          onStatusChanged?.(entity_type, entity_id, old_status, new_status)
          break
        case 'user_joined':
          const { user_id, username, project_id } = message.data
          if (project_id === currentProjectId) {
            setOnlineUsers((prev) =>
              new Map(prev).set(user_id, { username, joinedAt: new Date() })
            )
          }
          onUserJoined?.(user_id, username, project_id)
          break
        case 'user_left':
          const {
            user_id: leftUserId,
            username: leftUsername,
            project_id: leftProjectId,
          } = message.data
          if (leftProjectId === currentProjectId) {
            setOnlineUsers((prev) => {
              const newMap = new Map(prev)
              newMap.delete(leftUserId)
              return newMap
            })
          }
          onUserLeft?.(leftUserId, leftUsername, leftProjectId)
          break
        case 'error':
          onConnectionError?.(message.data.error)
          break
        case 'auth_failed':
        case 'auth_required':
          onAuthRequired?.()
          break
      }

      // Call registered message handlers
      const handlers = messageHandlersRef.current.get(message.type)
      if (handlers) {
        handlers.forEach((handler) => handler(message))
      }

      // Call wildcard handlers
      const wildcardHandlers = messageHandlersRef.current.get('*')
      if (wildcardHandlers) {
        wildcardHandlers.forEach((handler) => handler(message))
      }
    },
    [
      addToMessageHistory,
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
      confirmOptimisticPendingUpdates,
    ]
  )

  // Set up connection state listener
  useEffect(() => {
    connectionListenerRef.current = handleConnectionStateChange
    websocketService.subscribeToConnectionState(handleConnectionStateChange)

    return () => {
      if (connectionListenerRef.current) {
        websocketService.unsubscribeFromConnectionState(
          connectionListenerRef.current
        )
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

  // Project subscription management
  const handleSetCurrentProjectId = useCallback(
    async (projectId: string | null) => {
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
    },
    [currentProjectId, connectionState.status]
  )

  // Auto-subscribe to project when connected
  useEffect(() => {
    if (currentProjectId && connectionState.status === 'connected') {
      websocketService.subscribeToProject(currentProjectId).catch(console.error)
    }
  }, [currentProjectId, connectionState.status])

  // Subscribe to project channel messages
  useEffect(() => {
    if (currentProjectId && connectionState.status === 'connected') {
      const channel = `project:${currentProjectId}`
      websocketService.subscribe(channel, handleMessage)

      return () => {
        websocketService.unsubscribe(channel, handleMessage)
      }
    }
  }, [currentProjectId, connectionState.status, handleMessage])

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

  const subscribeToProject = useCallback(async (projectId: string) => {
    await websocketService.subscribeToProject(projectId)
  }, [])

  const unsubscribeFromProject = useCallback(async (projectId: string) => {
    await websocketService.unsubscribeFromProject(projectId)
  }, [])

  const subscribe = useCallback(
    (messageType: string, handler: (message: CentrifugeMessage) => void) => {
      if (!messageHandlersRef.current.has(messageType)) {
        messageHandlersRef.current.set(messageType, new Set())
      }
      messageHandlersRef.current.get(messageType)!.add(handler)
    },
    []
  )

  const unsubscribe = useCallback(
    (messageType: string, handler: (message: CentrifugeMessage) => void) => {
      const handlers = messageHandlersRef.current.get(messageType)
      if (handlers) {
        handlers.delete(handler)
        if (handlers.size === 0) {
          messageHandlersRef.current.delete(messageType)
        }
      }
    },
    []
  )

  const contextValue: WebSocketContextValue = {
    // Connection state
    connectionState,
    isConnected: connectionState.status === 'connected',
    isConnecting: connectionState.status === 'connecting',
    isReconnecting: connectionState.isReconnecting,
    lastError: connectionState.lastError,

    // Connection control
    connect,
    disconnect,
    reconnect,

    // Project subscription
    subscribeToProject,
    unsubscribeFromProject,
    currentProjectId,
    setCurrentProjectId: handleSetCurrentProjectId,

    // Message handling
    subscribe,
    unsubscribe,

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

// eslint-disable-next-line react-refresh/only-export-components
export function useWebSocketContext(): WebSocketContextValue {
  const context = useContext(WebSocketContext)
  if (!context) {
    throw new Error(
      'useWebSocketContext must be used within a WebSocketProvider'
    )
  }
  return context
}

// eslint-disable-next-line react-refresh/only-export-components
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
  }
}

// eslint-disable-next-line react-refresh/only-export-components
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
