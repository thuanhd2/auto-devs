import { API_CONFIG } from '@/config/api'

export interface WebSocketMessage {
  type: string
  data: any
  timestamp: string
  message_id: string
}

export interface WebSocketConfig {
  url: string
  reconnectInterval: number
  maxReconnectAttempts: number
  heartbeatInterval: number
  messageTimeout: number
  maxQueueSize: number
}

export interface ConnectionState {
  status: 'connecting' | 'connected' | 'disconnected' | 'error'
  isReconnecting: boolean
  reconnectAttempts: number
  lastError?: string
  connectedAt?: Date
  disconnectedAt?: Date
}

export interface QueuedMessage {
  id: string
  message: any
  timestamp: Date
  attempts: number
}

type EventListener = (message: WebSocketMessage) => void
type ConnectionListener = (state: ConnectionState) => void

export class WebSocketService {
  private ws: WebSocket | null = null
  private config: WebSocketConfig
  private connectionState: ConnectionState
  private eventListeners: Map<string, Set<EventListener>> = new Map()
  private connectionListeners: Set<ConnectionListener> = new Set()
  private messageQueue: QueuedMessage[] = []
  private reconnectTimer: NodeJS.Timeout | null = null
  private heartbeatTimer: NodeJS.Timeout | null = null
  private subscriptions: Set<string> = new Set()
  private authToken: string | null = null

  constructor(config?: Partial<WebSocketConfig>) {
    this.config = {
      url: API_CONFIG.WS_URL,
      reconnectInterval: 1000,
      maxReconnectAttempts: 10,
      heartbeatInterval: 30000,
      messageTimeout: 5000,
      maxQueueSize: 100,
      ...config,
    }

    this.connectionState = {
      status: 'disconnected',
      isReconnecting: false,
      reconnectAttempts: 0,
    }
  }

  setAuthToken(token: string | null) {
    this.authToken = token
    if (this.ws && this.connectionState.status === 'connected') {
      // Reconnect with new token
      this.disconnect()
      this.connect()
    }
  }

  connect(): Promise<void> {
    return new Promise((resolve, reject) => {
      if (this.ws && this.connectionState.status === 'connected') {
        resolve()
        return
      }

      this.updateConnectionState({
        status: 'connecting',
        isReconnecting: this.connectionState.reconnectAttempts > 0,
      })

      const wsUrl = this.buildWebSocketUrl()
      this.ws = new WebSocket(wsUrl)

      const connectTimeout = setTimeout(() => {
        if (this.ws && this.ws.readyState === WebSocket.CONNECTING) {
          this.ws.close()
          reject(new Error('Connection timeout'))
        }
      }, this.config.messageTimeout)

      this.ws.onopen = () => {
        clearTimeout(connectTimeout)
        this.updateConnectionState({
          status: 'connected',
          isReconnecting: false,
          reconnectAttempts: 0,
          connectedAt: new Date(),
          lastError: undefined,
        })

        this.startHeartbeat()
        this.resubscribeToTopics()
        this.processMessageQueue()
        resolve()
      }

      this.ws.onmessage = (event) => {
        try {
          const message: WebSocketMessage = JSON.parse(event.data)
          this.handleMessage(message)
        } catch (error) {
          console.error('Failed to parse WebSocket message:', error)
          console.error('WebSocket message:', event.data)
        }
      }

      this.ws.onclose = (event) => {
        clearTimeout(connectTimeout)
        this.stopHeartbeat()

        this.updateConnectionState({
          status: 'disconnected',
          disconnectedAt: new Date(),
        })

        if (!event.wasClean && this.shouldReconnect()) {
          this.scheduleReconnect()
        }
      }

      this.ws.onerror = (error) => {
        clearTimeout(connectTimeout)
        console.error('WebSocket error:', error)

        this.updateConnectionState({
          status: 'error',
          lastError: 'Connection error occurred',
        })

        reject(error)
      }
    })
  }

  disconnect() {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
      this.reconnectTimer = null
    }

    this.stopHeartbeat()

    if (this.ws) {
      this.ws.close(1000, 'Client disconnect')
      this.ws = null
    }

    this.updateConnectionState({
      status: 'disconnected',
      isReconnecting: false,
      disconnectedAt: new Date(),
    })
  }

  isConnectionReadyToSend(): boolean {
    return (
      this.connectionState.status === 'connected' &&
      this.ws &&
      this.ws.readyState !== WebSocket.CONNECTING
    )
  }

  send(message: any): Promise<void> {
    return new Promise((resolve, reject) => {
      if (!this.isConnectionReadyToSend()) {
        // Queue message for later delivery
        if (this.messageQueue.length < this.config.maxQueueSize) {
          this.messageQueue.push({
            id: this.generateMessageId(),
            message,
            timestamp: new Date(),
            attempts: 0,
          })
          resolve()
        } else {
          reject(new Error('Message queue is full'))
        }
        return
      }

      try {
        this.ws.send(JSON.stringify(message))
        resolve()
      } catch (error) {
        // Queue message and trigger reconnection
        if (this.messageQueue.length < this.config.maxQueueSize) {
          this.messageQueue.push({
            id: this.generateMessageId(),
            message,
            timestamp: new Date(),
            attempts: 0,
          })
        }
        reject(error)
      }
    })
  }

  subscribe(messageType: string, listener: EventListener) {
    if (!this.eventListeners.has(messageType)) {
      this.eventListeners.set(messageType, new Set())
    }
    this.eventListeners.get(messageType)!.add(listener)
  }

  unsubscribe(messageType: string, listener: EventListener) {
    const listeners = this.eventListeners.get(messageType)
    if (listeners) {
      listeners.delete(listener)
      if (listeners.size === 0) {
        this.eventListeners.delete(messageType)
      }
    }
  }

  subscribeToConnectionState(listener: ConnectionListener) {
    this.connectionListeners.add(listener)
  }

  unsubscribeFromConnectionState(listener: ConnectionListener) {
    this.connectionListeners.delete(listener)
  }

  subscribeToProject(projectId: string): Promise<void> {
    // check if projectId is already subscribed
    if (this.subscriptions.has(projectId)) {
      return Promise.resolve()
    }

    this.subscriptions.add(projectId)
    return this.send({
      type: 'subscription',
      data: {
        action: 'subscribe',
        project_id: projectId,
      },
    })
  }

  unsubscribeFromProject(projectId: string): Promise<void> {
    this.subscriptions.delete(projectId)
    return this.send({
      type: 'subscription',
      data: {
        action: 'unsubscribe',
        project_id: projectId,
      },
    })
  }

  getConnectionState(): ConnectionState {
    return { ...this.connectionState }
  }

  getQueuedMessageCount(): number {
    return this.messageQueue.length
  }

  clearMessageQueue() {
    this.messageQueue = []
  }

  private buildWebSocketUrl(): string {
    const url = new URL(this.config.url + '/connect')
    if (this.authToken) {
      url.searchParams.set('token', this.authToken)
    }
    return url.toString()
  }

  private handleMessage(message: WebSocketMessage) {
    // Handle system messages
    if (message.type === 'pong') {
      // Heartbeat response
      return
    }

    if (message.type === 'auth_required') {
      this.updateConnectionState({
        status: 'error',
        lastError: 'Authentication required',
      })
      return
    }

    if (message.type === 'auth_failed') {
      this.updateConnectionState({
        status: 'error',
        lastError: 'Authentication failed',
      })
      return
    }

    if (message.type === 'error') {
      console.error('WebSocket server error:', message.data)
      return
    }

    // Dispatch to registered listeners
    const listeners = this.eventListeners.get(message.type)
    if (listeners) {
      listeners.forEach((listener) => {
        try {
          listener(message)
        } catch (error) {
          console.error('Error in message listener:', error)
        }
      })
    }

    // Also dispatch to wildcard listeners
    const wildcardListeners = this.eventListeners.get('*')
    if (wildcardListeners) {
      wildcardListeners.forEach((listener) => {
        try {
          listener(message)
        } catch (error) {
          console.error('Error in wildcard listener:', error)
        }
      })
    }
  }

  private updateConnectionState(updates: Partial<ConnectionState>) {
    this.connectionState = { ...this.connectionState, ...updates }
    this.connectionListeners.forEach((listener) => {
      try {
        listener(this.connectionState)
      } catch (error) {
        console.error('Error in connection state listener:', error)
      }
    })
  }

  private shouldReconnect(): boolean {
    return (
      this.connectionState.reconnectAttempts < this.config.maxReconnectAttempts
    )
  }

  private scheduleReconnect() {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
    }

    const backoffDelay = this.calculateBackoffDelay()

    this.updateConnectionState({
      isReconnecting: true,
      reconnectAttempts: this.connectionState.reconnectAttempts + 1,
    })

    this.reconnectTimer = setTimeout(() => {
      this.connect().catch((error) => {
        console.error('Reconnection failed:', error)
        if (this.shouldReconnect()) {
          this.scheduleReconnect()
        } else {
          this.updateConnectionState({
            status: 'error',
            isReconnecting: false,
            lastError: 'Max reconnection attempts reached',
          })
        }
      })
    }, backoffDelay)
  }

  private calculateBackoffDelay(): number {
    // Exponential backoff with jitter
    const baseDelay = this.config.reconnectInterval
    const exponentialDelay =
      baseDelay * Math.pow(2, this.connectionState.reconnectAttempts)
    const maxDelay = 30000 // 30 seconds max
    const delay = Math.min(exponentialDelay, maxDelay)

    // Add jitter (±25%)
    const jitter = delay * 0.25 * (Math.random() - 0.5)
    return Math.max(1000, delay + jitter) // Minimum 1 second
  }

  private startHeartbeat() {
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer)
    }

    this.heartbeatTimer = setInterval(() => {
      if (this.connectionState.status === 'connected' && this.ws) {
        this.send({ type: 'ping' }).catch((error) => {
          console.error('Failed to send heartbeat:', error)
        })
      }
    }, this.config.heartbeatInterval)
  }

  private stopHeartbeat() {
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer)
      this.heartbeatTimer = null
    }
  }

  private resubscribeToTopics() {
    // Resubscribe to all previously subscribed projects
    this.subscriptions.forEach((projectId) => {
      this.subscribeToProject(projectId).catch((error) => {
        console.error('Failed to resubscribe to project:', projectId, error)
      })
    })
  }

  private processMessageQueue() {
    if (this.connectionState.status !== 'connected' || !this.ws) {
      return
    }

    const messagesToProcess = [...this.messageQueue]
    this.messageQueue = []

    messagesToProcess.forEach((queuedMessage) => {
      if (queuedMessage.attempts < 3) {
        this.send(queuedMessage.message).catch((error) => {
          console.error('Failed to send queued message:', error)
          // Re-queue with increased attempt count
          if (this.messageQueue.length < this.config.maxQueueSize) {
            this.messageQueue.push({
              ...queuedMessage,
              attempts: queuedMessage.attempts + 1,
            })
          }
        })
      }
    })
  }

  private generateMessageId(): string {
    return `msg_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
  }
}

// Export singleton instance
export const websocketService = new WebSocketService()
