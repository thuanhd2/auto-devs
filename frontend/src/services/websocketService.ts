import { API_CONFIG } from '@/config/api'
import {
  Centrifuge,
  PublicationContext,
  SubscriptionDataContext,
} from 'centrifuge'

export interface CentrifugeMessage {
  type: string
  data: any
  timestamp: string
  message_id: string
}

export interface CentrifugeConfig {
  url: string
  token?: string
  debug: boolean
  maxReconnectDelay: number
  timeout: number
}

export interface ConnectionState {
  status: 'connecting' | 'connected' | 'disconnected' | 'error'
  isReconnecting: boolean
  reconnectAttempts: number
  lastError?: string
  connectedAt?: Date
  disconnectedAt?: Date
}

type EventListener = (message: CentrifugeMessage) => void
type ConnectionListener = (state: ConnectionState) => void

export class CentrifugeService {
  private centrifuge: Centrifuge | null = null
  private config: CentrifugeConfig
  private connectionState: ConnectionState
  private eventListeners: Map<string, Set<EventListener>> = new Map()
  private connectionListeners: Set<ConnectionListener> = new Set()
  private subscriptions: Map<string, any> = new Map() // channel -> subscription
  private authToken: string | null = null

  constructor(config?: Partial<CentrifugeConfig>) {
    this.config = {
      url: API_CONFIG.WS_URL.replace('ws://', '').replace('wss://', ''), // Remove protocol prefix
      debug: process.env.NODE_ENV === 'development',
      maxReconnectDelay: 30000,
      timeout: 5000,
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

    // If connected and token changes, reconnect
    if (this.centrifuge && this.connectionState.status === 'connected') {
      this.disconnect()
      this.connect()
    }
  }

  connect(): Promise<void> {
    return new Promise((resolve, reject) => {
      if (this.centrifuge && this.connectionState.status === 'connected') {
        resolve()
        return
      }

      this.updateConnectionState({
        status: 'connecting',
        isReconnecting: this.connectionState.reconnectAttempts > 0,
      })

      // Build WebSocket URL
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
      const wsUrl = `${protocol}//${this.config.url}/connect`

      // Create Centrifuge instance
      this.centrifuge = new Centrifuge(wsUrl, {
        token: this.authToken || undefined,
        debug: this.config.debug,
        timeout: this.config.timeout,
        maxReconnectDelay: this.config.maxReconnectDelay,
      })

      // Set up event handlers
      this.centrifuge.on('connecting', (ctx) => {
        if (this.config.debug) {
          // eslint-disable-next-line no-console
          console.log('Centrifuge connecting:', ctx)
        }
        this.updateConnectionState({
          status: 'connecting',
        })
      })

      this.centrifuge.on('connected', (ctx) => {
        if (this.config.debug) {
          // eslint-disable-next-line no-console
          console.log('Centrifuge connected:', ctx)
        }
        this.updateConnectionState({
          status: 'connected',
          isReconnecting: false,
          reconnectAttempts: 0,
          connectedAt: new Date(),
          lastError: undefined,
        })
        this.resubscribeToChannels()
        resolve()
      })

      this.centrifuge.on('disconnected', (ctx) => {
        if (this.config.debug) {
          // eslint-disable-next-line no-console
          console.log('Centrifuge disconnected:', ctx)
        }
        this.updateConnectionState({
          status: 'disconnected',
          disconnectedAt: new Date(),
        })
      })

      this.centrifuge.on('error', (ctx) => {
        // eslint-disable-next-line no-console
        console.error('Centrifuge error:', ctx)
        this.updateConnectionState({
          status: 'error',
          lastError: ctx.error?.message || 'Connection error occurred',
        })
        reject(ctx.error)
      })

      // Start connection
      this.centrifuge.connect()
    })
  }

  disconnect() {
    if (this.centrifuge) {
      this.centrifuge.disconnect()
      this.centrifuge = null
    }

    // Clear all subscriptions
    this.subscriptions.clear()

    this.updateConnectionState({
      status: 'disconnected',
      isReconnecting: false,
      disconnectedAt: new Date(),
    })
  }

  isConnected(): boolean {
    return (
      this.connectionState.status === 'connected' && this.centrifuge !== null
    )
  }

  // Channel subscription methods
  subscribe(channel: string, listener: EventListener) {
    if (!this.eventListeners.has(channel)) {
      this.eventListeners.set(channel, new Set())
    }
    this.eventListeners.get(channel)!.add(listener)

    // If we're connected, subscribe to the channel immediately
    if (this.isConnected()) {
      this.subscribeToChannel(channel)
    }
  }

  unsubscribe(channel: string, listener: EventListener) {
    const listeners = this.eventListeners.get(channel)
    if (listeners) {
      listeners.delete(listener)
      if (listeners.size === 0) {
        this.eventListeners.delete(channel)
        // Unsubscribe from the channel if no more listeners
        this.unsubscribeFromChannel(channel)
      }
    }
  }

  subscribeToConnectionState(listener: ConnectionListener) {
    this.connectionListeners.add(listener)
  }

  unsubscribeFromConnectionState(listener: ConnectionListener) {
    this.connectionListeners.delete(listener)
  }

  // Project-specific methods for backward compatibility
  subscribeToProject(projectId: string): Promise<void> {
    return this.subscribeToChannel(`project:${projectId}`)
  }

  unsubscribeFromProject(projectId: string): Promise<void> {
    return this.unsubscribeFromChannel(`project:${projectId}`)
  }

  // Send message (for compatibility - Centrifuge uses RPC)
  send(message: any): Promise<void> {
    return new Promise((resolve, reject) => {
      if (!this.centrifuge || !this.isConnected()) {
        reject(new Error('Not connected'))
        return
      }

      // Convert legacy subscription messages to RPC calls
      if (message.type === 'subscription') {
        const { action, project_id } = message.data
        const method =
          action === 'subscribe' ? 'subscribe_project' : 'unsubscribe_project'

        this.centrifuge
          .rpc(method, { project_id })
          .then(() => resolve())
          .catch(reject)
      } else {
        // For other message types, we don't send them directly in Centrifuge
        // The server handles all publishing
        resolve()
      }
    })
  }

  getConnectionState(): ConnectionState {
    return { ...this.connectionState }
  }

  // Make this method public for enhanced service
  subscribeToChannel(channel: string): Promise<void> {
    return new Promise((resolve, reject) => {
      if (!this.centrifuge || !this.isConnected()) {
        reject(new Error('Not connected'))
        return
      }

      if (this.subscriptions.has(channel)) {
        resolve()
        return
      }

      try {
        const subscription = this.centrifuge.newSubscription(channel)

        subscription.on('publication', (ctx: PublicationContext) => {
          if (this.config.debug) {
            // eslint-disable-next-line no-console
            console.log('Received message on channel', channel, ':', ctx.data)
          }

          // Parse the message
          let message: CentrifugeMessage
          try {
            // The data might already be parsed or might be a string
            if (typeof ctx.data === 'string') {
              message = JSON.parse(ctx.data)
            } else {
              message = ctx.data as CentrifugeMessage
            }

            this.handleMessage(channel, message)
          } catch (error) {
            // eslint-disable-next-line no-console
            console.error('Failed to parse message:', error, ctx.data)
          }
        })

        subscription.on('subscribing', (ctx: SubscriptionDataContext) => {
          if (this.config.debug) {
            // eslint-disable-next-line no-console
            console.log('Subscribing to channel:', channel, ctx)
          }
        })

        subscription.on('subscribed', (ctx: SubscriptionDataContext) => {
          if (this.config.debug) {
            // eslint-disable-next-line no-console
            console.log('Subscribed to channel:', channel, ctx)
          }
          resolve()
        })

        subscription.on('unsubscribed', (ctx: SubscriptionDataContext) => {
          if (this.config.debug) {
            // eslint-disable-next-line no-console
            console.log('Unsubscribed from channel:', channel, ctx)
          }
          this.subscriptions.delete(channel)
        })

        subscription.on('error', (ctx) => {
          // eslint-disable-next-line no-console
          console.error('Subscription error for channel', channel, ':', ctx)
          reject(ctx.error)
        })

        this.subscriptions.set(channel, subscription)
        subscription.subscribe()
      } catch (error) {
        reject(error)
      }
    })
  }

  private unsubscribeFromChannel(channel: string): Promise<void> {
    return new Promise((resolve) => {
      const subscription = this.subscriptions.get(channel)
      if (subscription) {
        subscription.unsubscribe()
        this.subscriptions.delete(channel)
      }
      resolve()
    })
  }

  private resubscribeToChannels() {
    // Resubscribe to all channels that have listeners
    for (const channel of this.eventListeners.keys()) {
      this.subscribeToChannel(channel).catch((error) => {
        // eslint-disable-next-line no-console
        console.error('Failed to resubscribe to channel:', channel, error)
      })
    }
  }

  private handleMessage(channel: string, message: CentrifugeMessage) {
    // Handle specific channel messages
    const listeners = this.eventListeners.get(channel)
    if (listeners) {
      listeners.forEach((listener) => {
        try {
          listener(message)
        } catch (error) {
          // eslint-disable-next-line no-console
          console.error('Error in message listener:', error)
        }
      })
    }

    // Also handle wildcard listeners (for backward compatibility)
    const wildcardListeners = this.eventListeners.get('*')
    if (wildcardListeners) {
      wildcardListeners.forEach((listener) => {
        try {
          listener(message)
        } catch (error) {
          // eslint-disable-next-line no-console
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
        // eslint-disable-next-line no-console
        console.error('Error in connection state listener:', error)
      }
    })
  }

  // Compatibility methods for migration
  getQueuedMessageCount(): number {
    // Centrifuge handles queuing internally
    return 0
  }

  clearMessageQueue() {
    // No-op for Centrifuge
  }

  // Debug methods
  getActiveSubscriptions(): string[] {
    return Array.from(this.subscriptions.keys())
  }

  getConnectionInfo(): object {
    return {
      backend: 'centrifuge',
      status: this.connectionState.status,
      subscriptions: this.getActiveSubscriptions(),
      timestamp: new Date().toISOString(),
    }
  }
}

// Export singleton instance
export const websocketService = new CentrifugeService()
