import { useEffect, useRef, useCallback, useState } from 'react'
import { websocketService, WebSocketMessage, ConnectionState } from '@/services/websocketService'

export interface UseWebSocketOptions {
  autoConnect?: boolean
  projectId?: string
  authToken?: string | null
}

export interface UseWebSocketReturn {
  connectionState: ConnectionState
  isConnected: boolean
  isConnecting: boolean
  isReconnecting: boolean
  lastError: string | undefined
  queuedMessageCount: number
  connect: () => Promise<void>
  disconnect: () => void
  send: (message: any) => Promise<void>
  subscribe: (messageType: string, listener: (message: WebSocketMessage) => void) => void
  unsubscribe: (messageType: string, listener: (message: WebSocketMessage) => void) => void
  subscribeToProject: (projectId: string) => Promise<void>
  unsubscribeFromProject: (projectId: string) => Promise<void>
  clearMessageQueue: () => void
}

export function useWebSocket(options: UseWebSocketOptions = {}): UseWebSocketReturn {
  const {
    autoConnect = true,
    projectId,
    authToken,
  } = options

  const [connectionState, setConnectionState] = useState<ConnectionState>(
    websocketService.getConnectionState()
  )
  const [queuedMessageCount, setQueuedMessageCount] = useState(0)
  
  const listenersRef = useRef<Map<string, (message: WebSocketMessage) => void>>(new Map())
  const connectionListenerRef = useRef<(state: ConnectionState) => void>()
  const queueCheckIntervalRef = useRef<NodeJS.Timeout>()

  // Update connection state when it changes
  const handleConnectionStateChange = useCallback((state: ConnectionState) => {
    setConnectionState(state)
  }, [])

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

  // Set up queue count monitoring
  useEffect(() => {
    const updateQueueCount = () => {
      setQueuedMessageCount(websocketService.getQueuedMessageCount())
    }

    updateQueueCount()
    queueCheckIntervalRef.current = setInterval(updateQueueCount, 1000)

    return () => {
      if (queueCheckIntervalRef.current) {
        clearInterval(queueCheckIntervalRef.current)
      }
    }
  }, [])

  // Handle auth token changes
  useEffect(() => {
    websocketService.setAuthToken(authToken || null)
  }, [authToken])

  // Auto-connect on mount
  useEffect(() => {
    if (autoConnect && authToken) {
      websocketService.connect().catch((error) => {
        console.error('Auto-connect failed:', error)
      })
    }

    return () => {
      // Clean up listeners on unmount
      listenersRef.current.forEach((listener, messageType) => {
        websocketService.unsubscribe(messageType, listener)
      })
      listenersRef.current.clear()
    }
  }, [autoConnect, authToken])

  // Auto-subscribe to project if provided
  useEffect(() => {
    if (projectId && connectionState.status === 'connected') {
      websocketService.subscribeToProject(projectId).catch((error) => {
        console.error('Failed to subscribe to project:', error)
      })

      return () => {
        websocketService.unsubscribeFromProject(projectId).catch((error) => {
          console.error('Failed to unsubscribe from project:', error)
        })
      }
    }
  }, [projectId, connectionState.status])

  const connect = useCallback(async () => {
    try {
      await websocketService.connect()
    } catch (error) {
      console.error('Connection failed:', error)
      throw error
    }
  }, [])

  const disconnect = useCallback(() => {
    websocketService.disconnect()
  }, [])

  const send = useCallback(async (message: any) => {
    try {
      await websocketService.send(message)
    } catch (error) {
      console.error('Send failed:', error)
      throw error
    }
  }, [])

  const subscribe = useCallback((messageType: string, listener: (message: WebSocketMessage) => void) => {
    // Store reference for cleanup
    listenersRef.current.set(messageType, listener)
    websocketService.subscribe(messageType, listener)
  }, [])

  const unsubscribe = useCallback((messageType: string, listener: (message: WebSocketMessage) => void) => {
    listenersRef.current.delete(messageType)
    websocketService.unsubscribe(messageType, listener)
  }, [])

  const subscribeToProject = useCallback(async (projectId: string) => {
    try {
      await websocketService.subscribeToProject(projectId)
    } catch (error) {
      console.error('Failed to subscribe to project:', error)
      throw error
    }
  }, [])

  const unsubscribeFromProject = useCallback(async (projectId: string) => {
    try {
      await websocketService.unsubscribeFromProject(projectId)
    } catch (error) {
      console.error('Failed to unsubscribe from project:', error)
      throw error
    }
  }, [])

  const clearMessageQueue = useCallback(() => {
    websocketService.clearMessageQueue()
    setQueuedMessageCount(0)
  }, [])

  return {
    connectionState,
    isConnected: connectionState.status === 'connected',
    isConnecting: connectionState.status === 'connecting',
    isReconnecting: connectionState.isReconnecting,
    lastError: connectionState.lastError,
    queuedMessageCount,
    connect,
    disconnect,
    send,
    subscribe,
    unsubscribe,
    subscribeToProject,
    unsubscribeFromProject,
    clearMessageQueue,
  }
}

// Specialized hooks for common use cases

export function useWebSocketMessage<T = any>(
  messageType: string,
  handler: (data: T, message: WebSocketMessage) => void,
  options: UseWebSocketOptions = {}
) {
  const { subscribe, unsubscribe, ...webSocket } = useWebSocket(options)
  
  const handlerRef = useRef(handler)
  handlerRef.current = handler

  useEffect(() => {
    const listener = (message: WebSocketMessage) => {
      handlerRef.current(message.data, message)
    }

    subscribe(messageType, listener)

    return () => {
      unsubscribe(messageType, listener)
    }
  }, [messageType, subscribe, unsubscribe])

  return webSocket
}

export function useWebSocketTaskUpdates(
  projectId: string | undefined,
  onTaskUpdate: (task: any, changes?: any) => void,
  options: UseWebSocketOptions = {}
) {
  const mergedOptions = { ...options, projectId }
  
  const { subscribe, unsubscribe, ...webSocket } = useWebSocket(mergedOptions)
  
  const onTaskUpdateRef = useRef(onTaskUpdate)
  onTaskUpdateRef.current = onTaskUpdate

  useEffect(() => {
    const taskCreatedListener = (message: WebSocketMessage) => {
      onTaskUpdateRef.current(message.data, { created: true })
    }

    const taskUpdatedListener = (message: WebSocketMessage) => {
      onTaskUpdateRef.current(message.data.task, message.data.changes)
    }

    const taskDeletedListener = (message: WebSocketMessage) => {
      onTaskUpdateRef.current(message.data, { deleted: true })
    }

    subscribe('task_created', taskCreatedListener)
    subscribe('task_updated', taskUpdatedListener)
    subscribe('task_deleted', taskDeletedListener)

    return () => {
      unsubscribe('task_created', taskCreatedListener)
      unsubscribe('task_updated', taskUpdatedListener)
      unsubscribe('task_deleted', taskDeletedListener)
    }
  }, [subscribe, unsubscribe])

  return webSocket
}

export function useWebSocketProjectUpdates(
  projectId: string | undefined,
  onProjectUpdate: (project: any) => void,
  options: UseWebSocketOptions = {}
) {
  const mergedOptions = { ...options, projectId }
  
  const { subscribe, unsubscribe, ...webSocket } = useWebSocket(mergedOptions)
  
  const onProjectUpdateRef = useRef(onProjectUpdate)
  onProjectUpdateRef.current = onProjectUpdate

  useEffect(() => {
    const projectUpdatedListener = (message: WebSocketMessage) => {
      onProjectUpdateRef.current(message.data)
    }

    subscribe('project_updated', projectUpdatedListener)

    return () => {
      unsubscribe('project_updated', projectUpdatedListener)
    }
  }, [subscribe, unsubscribe])

  return webSocket
}