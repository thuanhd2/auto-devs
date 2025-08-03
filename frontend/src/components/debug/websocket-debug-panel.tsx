import React, { useState, useEffect, useRef } from 'react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Separator } from '@/components/ui/separator'
import { Switch } from '@/components/ui/switch'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { 
  Bug, 
  Trash2, 
  Download, 
  Send, 
  Activity, 
  MessageSquare, 
  Clock, 
  Zap,
  AlertTriangle,
  CheckCircle,
  XCircle,
  Loader2
} from 'lucide-react'
import { useWebSocketContext, useWebSocketDebug } from '@/context/websocket-context'
import { WebSocketMessage } from '@/services/websocketService'
import { optimisticUpdateManager } from '@/services/optimisticUpdates'

export interface WebSocketDebugPanelProps {
  className?: string
  defaultOpen?: boolean
}

export function WebSocketDebugPanel({ 
  className,
  defaultOpen = false 
}: WebSocketDebugPanelProps) {
  const [isOpen, setIsOpen] = useState(defaultOpen)
  const [selectedMessage, setSelectedMessage] = useState<WebSocketMessage | null>(null)
  const [customMessage, setCustomMessage] = useState('')
  const [messageFilter, setMessageFilter] = useState('')
  const [autoScroll, setAutoScroll] = useState(true)
  const [connectionMetrics, setConnectionMetrics] = useState({
    messagesReceived: 0,
    messagesSent: 0,
    connectionTime: 0,
    reconnectCount: 0,
    lastMessageTime: null as Date | null,
  })

  const scrollAreaRef = useRef<HTMLDivElement>(null)
  const {
    connectionState,
    send,
    queuedMessageCount,
    optimisticUpdateCount,
    onlineUsers,
  } = useWebSocketContext()

  const {
    messageHistory,
    clearMessageHistory,
    isDebugMode,
    setDebugMode,
    clearOptimisticUpdates,
  } = useWebSocketDebug()

  // Auto-scroll to bottom when new messages arrive
  useEffect(() => {
    if (autoScroll && scrollAreaRef.current) {
      scrollAreaRef.current.scrollTop = scrollAreaRef.current.scrollHeight
    }
  }, [messageHistory, autoScroll])

  // Update connection metrics
  useEffect(() => {
    setConnectionMetrics(prev => ({
      ...prev,
      messagesReceived: messageHistory.length,
      lastMessageTime: messageHistory.length > 0 ? new Date(messageHistory[messageHistory.length - 1].timestamp) : null,
      reconnectCount: connectionState.reconnectAttempts,
      connectionTime: connectionState.connectedAt ? Date.now() - connectionState.connectedAt.getTime() : 0,
    }))
  }, [messageHistory, connectionState])

  const filteredMessages = messageHistory.filter(message => {
    if (!messageFilter) return true
    return (
      message.type.toLowerCase().includes(messageFilter.toLowerCase()) ||
      JSON.stringify(message.data).toLowerCase().includes(messageFilter.toLowerCase())
    )
  })

  const sendCustomMessage = () => {
    try {
      const parsedMessage = JSON.parse(customMessage)
      send(parsedMessage)
      setConnectionMetrics(prev => ({ ...prev, messagesSent: prev.messagesSent + 1 }))
      setCustomMessage('')
    } catch (error) {
      console.error('Invalid JSON message:', error)
    }
  }

  const exportMessageHistory = () => {
    const data = {
      timestamp: new Date().toISOString(),
      connectionState,
      metrics: connectionMetrics,
      messages: messageHistory,
      optimisticUpdates: optimisticUpdateManager.getAllPendingUpdates(),
    }
    
    const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `websocket-debug-${Date.now()}.json`
    a.click()
    URL.revokeObjectURL(url)
  }

  const getMessageTypeColor = (type: string) => {
    if (type.startsWith('task_')) return 'bg-blue-100 text-blue-800'
    if (type.startsWith('project_')) return 'bg-green-100 text-green-800'
    if (type.startsWith('user_')) return 'bg-purple-100 text-purple-800'
    if (type === 'error' || type === 'auth_failed') return 'bg-red-100 text-red-800'
    if (type === 'ping' || type === 'pong') return 'bg-gray-100 text-gray-800'
    return 'bg-yellow-100 text-yellow-800'
  }

  const getConnectionStatusIcon = () => {
    switch (connectionState.status) {
      case 'connected':
        return <CheckCircle className="h-4 w-4 text-green-600" />
      case 'connecting':
        return <Loader2 className="h-4 w-4 animate-spin text-yellow-600" />
      case 'disconnected':
        return <XCircle className="h-4 w-4 text-gray-600" />
      case 'error':
        return <AlertTriangle className="h-4 w-4 text-red-600" />
      default:
        return <XCircle className="h-4 w-4 text-gray-600" />
    }
  }

  if (!isOpen) {
    return (
      <Button
        variant="outline"
        size="sm"
        onClick={() => setIsOpen(true)}
        className={cn('fixed bottom-4 right-4 z-50', className)}
      >
        <Bug className="h-4 w-4 mr-2" />
        Debug
      </Button>
    )
  }

  return (
    <Card className={cn('fixed bottom-4 right-4 w-96 max-h-[600px] z-50 shadow-lg', className)}>
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <CardTitle className="text-lg flex items-center gap-2">
            <Bug className="h-5 w-5" />
            WebSocket Debug
          </CardTitle>
          <div className="flex items-center gap-2">
            <Switch
              checked={isDebugMode}
              onCheckedChange={setDebugMode}
              size="sm"
            />
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setIsOpen(false)}
            >
              ×
            </Button>
          </div>
        </div>
      </CardHeader>

      <CardContent className="p-0">
        <Tabs defaultValue="messages" className="w-full">
          <TabsList className="grid w-full grid-cols-3 mx-4 mb-2">
            <TabsTrigger value="messages">Messages</TabsTrigger>
            <TabsTrigger value="connection">Connection</TabsTrigger>
            <TabsTrigger value="tools">Tools</TabsTrigger>
          </TabsList>

          <TabsContent value="messages" className="px-4 pb-4 space-y-3">
            <div className="flex items-center gap-2">
              <Input
                placeholder="Filter messages..."
                value={messageFilter}
                onChange={(e) => setMessageFilter(e.target.value)}
                className="flex-1"
              />
              <Switch
                checked={autoScroll}
                onCheckedChange={setAutoScroll}
                size="sm"
              />
            </div>

            <div className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground">
                {filteredMessages.length} messages
              </span>
              <Button
                variant="ghost"
                size="sm"
                onClick={clearMessageHistory}
              >
                <Trash2 className="h-4 w-4" />
              </Button>
            </div>

            <ScrollArea 
              ref={scrollAreaRef}
              className="h-64 border rounded"
            >
              <div className="p-2 space-y-2">
                {filteredMessages.map((message, index) => (
                  <div
                    key={`${message.message_id}-${index}`}
                    className={cn(
                      'p-2 rounded border cursor-pointer hover:bg-accent',
                      selectedMessage?.message_id === message.message_id && 'bg-accent'
                    )}
                    onClick={() => setSelectedMessage(message)}
                  >
                    <div className="flex items-center justify-between mb-1">
                      <Badge 
                        variant="outline" 
                        className={cn('text-xs', getMessageTypeColor(message.type))}
                      >
                        {message.type}
                      </Badge>
                      <span className="text-xs text-muted-foreground">
                        {new Date(message.timestamp).toLocaleTimeString()}
                      </span>
                    </div>
                    <div className="text-xs font-mono bg-muted p-1 rounded overflow-hidden">
                      {JSON.stringify(message.data).substring(0, 50)}...
                    </div>
                  </div>
                ))}
              </div>
            </ScrollArea>

            {selectedMessage && (
              <div className="border rounded p-2">
                <div className="text-sm font-medium mb-2">Message Details</div>
                <pre className="text-xs bg-muted p-2 rounded overflow-auto max-h-32">
                  {JSON.stringify(selectedMessage, null, 2)}
                </pre>
              </div>
            )}
          </TabsContent>

          <TabsContent value="connection" className="px-4 pb-4 space-y-3">
            <div className="grid grid-cols-2 gap-3">
              <Card className="p-3">
                <div className="flex items-center gap-2 mb-2">
                  {getConnectionStatusIcon()}
                  <span className="text-sm font-medium">Status</span>
                </div>
                <div className="text-xs text-muted-foreground">
                  {connectionState.status}
                  {connectionState.isReconnecting && ' (reconnecting)'}
                </div>
              </Card>

              <Card className="p-3">
                <div className="flex items-center gap-2 mb-2">
                  <Clock className="h-4 w-4" />
                  <span className="text-sm font-medium">Uptime</span>
                </div>
                <div className="text-xs text-muted-foreground">
                  {Math.floor(connectionMetrics.connectionTime / 1000)}s
                </div>
              </Card>

              <Card className="p-3">
                <div className="flex items-center gap-2 mb-2">
                  <MessageSquare className="h-4 w-4" />
                  <span className="text-sm font-medium">Messages</span>
                </div>
                <div className="text-xs text-muted-foreground">
                  ↓{connectionMetrics.messagesReceived} ↑{connectionMetrics.messagesSent}
                </div>
              </Card>

              <Card className="p-3">
                <div className="flex items-center gap-2 mb-2">
                  <Activity className="h-4 w-4" />
                  <span className="text-sm font-medium">Queue</span>
                </div>
                <div className="text-xs text-muted-foreground">
                  {queuedMessageCount} queued
                </div>
              </Card>
            </div>

            <Separator />

            <div className="space-y-2">
              <div className="text-sm font-medium">Optimistic Updates</div>
              <div className="flex items-center justify-between">
                <span className="text-sm text-muted-foreground">
                  {optimisticUpdateCount} pending
                </span>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={clearOptimisticUpdates}
                  disabled={optimisticUpdateCount === 0}
                >
                  Clear
                </Button>
              </div>
            </div>

            <Separator />

            <div className="space-y-2">
              <div className="text-sm font-medium">Online Users</div>
              <div className="text-sm text-muted-foreground">
                {onlineUsers.size} users online
              </div>
              {onlineUsers.size > 0 && (
                <div className="text-xs">
                  {Array.from(onlineUsers.values()).map((user, index) => (
                    <div key={index}>{user.username}</div>
                  ))}
                </div>
              )}
            </div>

            {connectionState.lastError && (
              <Alert variant="destructive">
                <AlertTriangle className="h-4 w-4" />
                <AlertDescription className="text-xs">
                  {connectionState.lastError}
                </AlertDescription>
              </Alert>
            )}
          </TabsContent>

          <TabsContent value="tools" className="px-4 pb-4 space-y-3">
            <div className="space-y-2">
              <div className="text-sm font-medium">Send Custom Message</div>
              <Textarea
                placeholder='{"type": "test", "data": {}}'
                value={customMessage}
                onChange={(e) => setCustomMessage(e.target.value)}
                className="font-mono text-xs"
                rows={3}
              />
              <Button
                onClick={sendCustomMessage}
                disabled={!customMessage.trim() || connectionState.status !== 'connected'}
                size="sm"
                className="w-full"
              >
                <Send className="h-4 w-4 mr-2" />
                Send Message
              </Button>
            </div>

            <Separator />

            <div className="space-y-2">
              <Button
                onClick={exportMessageHistory}
                variant="outline"
                size="sm"
                className="w-full"
              >
                <Download className="h-4 w-4 mr-2" />
                Export Debug Data
              </Button>

              <Button
                onClick={clearMessageHistory}
                variant="outline"
                size="sm"
                className="w-full"
              >
                <Trash2 className="h-4 w-4 mr-2" />
                Clear Message History
              </Button>
            </div>

            <Separator />

            <div className="text-xs text-muted-foreground space-y-1">
              <div>Debug Mode: {isDebugMode ? 'Enabled' : 'Disabled'}</div>
              <div>Message History: {messageHistory.length} messages</div>
              <div>Last Message: {connectionMetrics.lastMessageTime?.toLocaleTimeString() || 'None'}</div>
              <div>Reconnect Attempts: {connectionMetrics.reconnectCount}</div>
            </div>
          </TabsContent>
        </Tabs>
      </CardContent>
    </Card>
  )
}

export interface WebSocketPerformanceMonitorProps {
  onMetricsUpdate?: (metrics: any) => void
}

export function WebSocketPerformanceMonitor({ 
  onMetricsUpdate 
}: WebSocketPerformanceMonitorProps) {
  const [metrics, setMetrics] = useState({
    messageRate: 0,
    avgResponseTime: 0,
    errorRate: 0,
    connectionUptime: 0,
  })

  const { connectionState, messageHistory } = useWebSocketContext()
  const metricsRef = useRef({
    messageCount: 0,
    errorCount: 0,
    responseTimes: [] as number[],
    lastUpdate: Date.now(),
  })

  useEffect(() => {
    const interval = setInterval(() => {
      const now = Date.now()
      const timeDiff = now - metricsRef.current.lastUpdate
      const newMessageCount = messageHistory.length
      const messagesDelta = newMessageCount - metricsRef.current.messageCount

      const newMetrics = {
        messageRate: messagesDelta / (timeDiff / 1000),
        avgResponseTime: metricsRef.current.responseTimes.length > 0 
          ? metricsRef.current.responseTimes.reduce((a, b) => a + b, 0) / metricsRef.current.responseTimes.length 
          : 0,
        errorRate: metricsRef.current.errorCount / Math.max(newMessageCount, 1),
        connectionUptime: connectionState.connectedAt 
          ? now - connectionState.connectedAt.getTime() 
          : 0,
      }

      setMetrics(newMetrics)
      onMetricsUpdate?.(newMetrics)

      metricsRef.current.messageCount = newMessageCount
      metricsRef.current.lastUpdate = now
    }, 1000)

    return () => clearInterval(interval)
  }, [messageHistory, connectionState, onMetricsUpdate])

  // Track error messages
  useEffect(() => {
    const errorMessages = messageHistory.filter(msg => 
      msg.type === 'error' || msg.type === 'auth_failed'
    )
    metricsRef.current.errorCount = errorMessages.length
  }, [messageHistory])

  return null // This is a monitoring component, no UI
}