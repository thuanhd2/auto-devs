import { websocketService, WebSocketMessage, ConnectionState as LegacyConnectionState } from './websocketService'
import { centrifugeService, CentrifugeMessage, ConnectionState as CentrifugeConnectionState } from './centrifugeService'

// Unified types
export type ConnectionState = LegacyConnectionState | CentrifugeConnectionState
export type Message = WebSocketMessage | CentrifugeMessage
export type EventListener = (message: Message) => void
export type ConnectionListener = (state: ConnectionState) => void

export interface WebSocketServiceConfig {
  backend: 'legacy' | 'centrifuge' | 'auto'
  fallbackToLegacy: boolean
  debug: boolean
}

export class EnhancedWebSocketService {
  private config: WebSocketServiceConfig
  private currentBackend: 'legacy' | 'centrifuge'
  private eventListeners: Map<string, Set<EventListener>> = new Map()
  private connectionListeners: Set<ConnectionListener> = new Set()

  constructor(config?: Partial<WebSocketServiceConfig>) {
    this.config = {
      backend: 'auto',
      fallbackToLegacy: true,
      debug: false,
      ...config,
    }

    // Determine which backend to use
    this.currentBackend = this.determineBackend()
    
    if (this.config.debug) {
      console.log('Enhanced WebSocket Service initialized with backend:', this.currentBackend)
    }

    this.setupBackendListeners()
  }

  private determineBackend(): 'legacy' | 'centrifuge' {
    if (this.config.backend === 'legacy') {
      return 'legacy'
    }
    if (this.config.backend === 'centrifuge') {
      return 'centrifuge'
    }
    
    // Auto-detect: Check environment variables or server capability
    const useLegacy = localStorage.getItem('websocket_use_legacy')
    if (useLegacy === 'true') {
      return 'legacy'
    }
    if (useLegacy === 'false') {
      return 'centrifuge'
    }
    
    // Default to legacy for safety during migration
    return 'legacy'
  }

  private setupBackendListeners() {
    if (this.currentBackend === 'centrifuge') {
      centrifugeService.subscribeToConnectionState((state) => {
        this.notifyConnectionListeners(state)
      })
    } else {
      websocketService.subscribeToConnectionState((state) => {
        this.notifyConnectionListeners(state)
      })
    }
  }

  // Public API methods
  setAuthToken(token: string | null) {
    if (this.currentBackend === 'centrifuge') {
      centrifugeService.setAuthToken(token)
    } else {
      websocketService.setAuthToken(token)
    }
  }

  async connect(): Promise<void> {
    try {
      if (this.currentBackend === 'centrifuge') {
        await centrifugeService.connect()
      } else {
        await websocketService.connect()
      }
    } catch (error) {
      if (this.config.fallbackToLegacy && this.currentBackend === 'centrifuge') {
        console.warn('Centrifuge connection failed, falling back to legacy WebSocket:', error)
        this.switchBackend('legacy')
        return this.connect()
      }
      throw error
    }
  }

  disconnect() {
    if (this.currentBackend === 'centrifuge') {
      centrifugeService.disconnect()
    } else {
      websocketService.disconnect()
    }
  }

  isConnected(): boolean {
    if (this.currentBackend === 'centrifuge') {
      return centrifugeService.isConnected()
    } else {
      return websocketService.isConnectionReadyToSend()
    }
  }

  async send(message: any): Promise<void> {
    if (this.currentBackend === 'centrifuge') {
      return centrifugeService.send(message)
    } else {
      return websocketService.send(message)
    }
  }

  subscribe(messageType: string, listener: EventListener) {
    // Add to our own listener registry
    if (!this.eventListeners.has(messageType)) {
      this.eventListeners.set(messageType, new Set())
    }
    this.eventListeners.get(messageType)!.add(listener)

    // Subscribe to the appropriate backend
    if (this.currentBackend === 'centrifuge') {
      // For Centrifuge, we need to map message types to channels
      const channel = this.mapMessageTypeToChannel(messageType)
      centrifugeService.subscribe(channel, listener as any)
    } else {
      websocketService.subscribe(messageType, listener as any)
    }
  }

  unsubscribe(messageType: string, listener: EventListener) {
    const listeners = this.eventListeners.get(messageType)
    if (listeners) {
      listeners.delete(listener)
      if (listeners.size === 0) {
        this.eventListeners.delete(messageType)
      }
    }

    if (this.currentBackend === 'centrifuge') {
      const channel = this.mapMessageTypeToChannel(messageType)
      centrifugeService.unsubscribe(channel, listener as any)
    } else {
      websocketService.unsubscribe(messageType, listener as any)
    }
  }

  subscribeToConnectionState(listener: ConnectionListener) {
    this.connectionListeners.add(listener)
  }

  unsubscribeFromConnectionState(listener: ConnectionListener) {
    this.connectionListeners.delete(listener)
  }

  async subscribeToProject(projectId: string): Promise<void> {
    if (this.currentBackend === 'centrifuge') {
      return centrifugeService.subscribeToProject(projectId)
    } else {
      return websocketService.subscribeToProject(projectId)
    }
  }

  async unsubscribeFromProject(projectId: string): Promise<void> {
    if (this.currentBackend === 'centrifuge') {
      return centrifugeService.unsubscribeFromProject(projectId)
    } else {
      return websocketService.unsubscribeFromProject(projectId)
    }
  }

  getConnectionState(): ConnectionState {
    if (this.currentBackend === 'centrifuge') {
      return centrifugeService.getConnectionState()
    } else {
      return websocketService.getConnectionState()
    }
  }

  getQueuedMessageCount(): number {
    if (this.currentBackend === 'centrifuge') {
      return centrifugeService.getQueuedMessageCount()
    } else {
      return websocketService.getQueuedMessageCount()
    }
  }

  clearMessageQueue() {
    if (this.currentBackend === 'centrifuge') {
      centrifugeService.clearMessageQueue()
    } else {
      websocketService.clearMessageQueue()
    }
  }

  // Backend management
  getCurrentBackend(): 'legacy' | 'centrifuge' {
    return this.currentBackend
  }

  async switchBackend(backend: 'legacy' | 'centrifuge'): Promise<void> {
    if (this.currentBackend === backend) {
      return
    }

    if (this.config.debug) {
      console.log('Switching WebSocket backend from', this.currentBackend, 'to', backend)
    }

    // Disconnect current backend
    this.disconnect()

    // Clear current listeners
    this.clearBackendListeners()

    // Switch backend
    this.currentBackend = backend

    // Setup new listeners
    this.setupBackendListeners()

    // Resubscribe to all channels/message types
    await this.resubscribeAll()

    // Store preference
    localStorage.setItem('websocket_use_legacy', backend === 'legacy' ? 'true' : 'false')

    // Reconnect
    return this.connect()
  }

  // Utility methods
  getServiceInfo(): object {
    return {
      backend: this.currentBackend,
      config: this.config,
      connectionState: this.getConnectionState(),
      messageTypes: Array.from(this.eventListeners.keys()),
      timestamp: new Date().toISOString(),
    }
  }

  // Private methods
  private mapMessageTypeToChannel(messageType: string): string {
    // Map legacy message types to Centrifuge channels
    // For now, we'll use a simple mapping - in practice you might want more sophisticated routing
    
    if (messageType === '*') {
      // Wildcard listeners need special handling
      return 'system:all'
    }
    
    // For most message types, we'll use the global system channel
    // The server will need to publish to appropriate channels
    return 'system:messages'
  }

  private clearBackendListeners() {
    if (this.currentBackend === 'centrifuge') {
      // Clear Centrifuge listeners - this would need implementation in centrifugeService
    } else {
      // Clear legacy listeners - this would need implementation in websocketService
    }
  }

  private async resubscribeAll(): Promise<void> {
    // Resubscribe to all message types/channels
    const resubscribePromises: Promise<void>[] = []
    
    for (const [messageType, listeners] of this.eventListeners.entries()) {
      for (const listener of listeners) {
        if (this.currentBackend === 'centrifuge') {
          const channel = this.mapMessageTypeToChannel(messageType)
          resubscribePromises.push(
            centrifugeService.subscribeToChannel(channel).then(() => {
              centrifugeService.subscribe(channel, listener as any)
            }).catch(() => {}) // Ignore errors for now
          )
        } else {
          websocketService.subscribe(messageType, listener as any)
        }
      }
    }

    await Promise.all(resubscribePromises)
  }

  private notifyConnectionListeners(state: ConnectionState) {
    this.connectionListeners.forEach((listener) => {
      try {
        listener(state)
      } catch (error) {
        console.error('Error in connection state listener:', error)
      }
    })
  }

  // Debug methods
  getDebugInfo(): object {
    return {
      backend: this.currentBackend,
      config: this.config,
      eventListeners: Array.from(this.eventListeners.keys()),
      connectionListeners: this.connectionListeners.size,
      backendInfo: this.currentBackend === 'centrifuge' 
        ? centrifugeService.getConnectionInfo()
        : { backend: 'legacy', status: websocketService.getConnectionState().status },
    }
  }
}

// Export singleton instance
export const enhancedWebSocketService = new EnhancedWebSocketService()

// Export for backward compatibility
export type { WebSocketMessage, ConnectionState }