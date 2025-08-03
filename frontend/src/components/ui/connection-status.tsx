import React from 'react'
import { cn } from '@/lib/utils'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'
import { Progress } from '@/components/ui/progress'
import { 
  Wifi, 
  WifiOff, 
  AlertCircle, 
  CheckCircle, 
  Loader2, 
  RefreshCw,
  MessageSquare,
  Clock
} from 'lucide-react'
import { ConnectionState } from '@/services/websocketService'

export interface ConnectionStatusProps {
  connectionState: ConnectionState
  queuedMessageCount?: number
  onReconnect?: () => void
  onClearQueue?: () => void
  className?: string
  variant?: 'badge' | 'full' | 'compact'
  showDetails?: boolean
}

export function ConnectionStatus({
  connectionState,
  queuedMessageCount = 0,
  onReconnect,
  onClearQueue,
  className,
  variant = 'badge',
  showDetails = false,
}: ConnectionStatusProps) {
  const getStatusIcon = () => {
    switch (connectionState.status) {
      case 'connected':
        return <CheckCircle className="h-4 w-4 text-green-600" />
      case 'connecting':
        return <Loader2 className="h-4 w-4 animate-spin text-yellow-600" />
      case 'disconnected':
        return <WifiOff className="h-4 w-4 text-gray-600" />
      case 'error':
        return <AlertCircle className="h-4 w-4 text-red-600" />
      default:
        return <WifiOff className="h-4 w-4 text-gray-600" />
    }
  }

  const getStatusText = () => {
    if (connectionState.isReconnecting) {
      return `Reconnecting... (${connectionState.reconnectAttempts})`
    }
    
    switch (connectionState.status) {
      case 'connected':
        return 'Connected'
      case 'connecting':
        return 'Connecting...'
      case 'disconnected':
        return 'Disconnected'
      case 'error':
        return connectionState.lastError || 'Connection Error'
      default:
        return 'Unknown'
    }
  }

  const getStatusColor = () => {
    switch (connectionState.status) {
      case 'connected':
        return 'bg-green-100 text-green-800 border-green-200'
      case 'connecting':
        return 'bg-yellow-100 text-yellow-800 border-yellow-200'
      case 'disconnected':
        return 'bg-gray-100 text-gray-800 border-gray-200'
      case 'error':
        return 'bg-red-100 text-red-800 border-red-200'
      default:
        return 'bg-gray-100 text-gray-800 border-gray-200'
    }
  }

  if (variant === 'badge') {
    return (
      <TooltipProvider>
        <Tooltip>
          <TooltipTrigger asChild>
            <Badge
              variant="outline"
              className={cn(
                'flex items-center gap-1 px-2 py-1',
                getStatusColor(),
                className
              )}
            >
              {getStatusIcon()}
              <span className="text-xs font-medium">{getStatusText()}</span>
              {queuedMessageCount > 0 && (
                <Badge variant="secondary" className="ml-1 px-1 text-xs">
                  {queuedMessageCount}
                </Badge>
              )}
            </Badge>
          </TooltipTrigger>
          <TooltipContent>
            <div className="text-sm">
              <div>Status: {getStatusText()}</div>
              {connectionState.connectedAt && (
                <div>Connected: {connectionState.connectedAt.toLocaleTimeString()}</div>
              )}
              {connectionState.disconnectedAt && connectionState.status === 'disconnected' && (
                <div>Disconnected: {connectionState.disconnectedAt.toLocaleTimeString()}</div>
              )}
              {queuedMessageCount > 0 && (
                <div>Queued messages: {queuedMessageCount}</div>
              )}
            </div>
          </TooltipContent>
        </Tooltip>
      </TooltipProvider>
    )
  }

  if (variant === 'compact') {
    return (
      <div className={cn('flex items-center gap-2', className)}>
        {getStatusIcon()}
        <span className="text-sm text-muted-foreground">{getStatusText()}</span>
        {queuedMessageCount > 0 && (
          <Badge variant="secondary" className="px-1 text-xs">
            <MessageSquare className="h-3 w-3 mr-1" />
            {queuedMessageCount}
          </Badge>
        )}
      </div>
    )
  }

  // Full variant
  return (
    <div className={cn('space-y-3', className)}>
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          {getStatusIcon()}
          <span className="font-medium">{getStatusText()}</span>
        </div>
        
        {(connectionState.status === 'disconnected' || connectionState.status === 'error') && onReconnect && (
          <Button
            variant="outline"
            size="sm"
            onClick={onReconnect}
            disabled={connectionState.status === 'connecting'}
          >
            <RefreshCw className="h-4 w-4 mr-1" />
            Reconnect
          </Button>
        )}
      </div>

      {connectionState.isReconnecting && (
        <div className="space-y-2">
          <div className="flex items-center gap-2 text-sm text-muted-foreground">
            <Loader2 className="h-4 w-4 animate-spin" />
            <span>Attempting to reconnect... (Attempt {connectionState.reconnectAttempts})</span>
          </div>
          <Progress value={(connectionState.reconnectAttempts / 10) * 100} className="h-2" />
        </div>
      )}

      {showDetails && (
        <div className="space-y-2 text-sm text-muted-foreground">
          {connectionState.connectedAt && connectionState.status === 'connected' && (
            <div className="flex items-center gap-2">
              <Clock className="h-4 w-4" />
              <span>Connected since {connectionState.connectedAt.toLocaleString()}</span>
            </div>
          )}
          
          {connectionState.disconnectedAt && connectionState.status !== 'connected' && (
            <div className="flex items-center gap-2">
              <Clock className="h-4 w-4" />
              <span>Disconnected at {connectionState.disconnectedAt.toLocaleString()}</span>
            </div>
          )}
          
          {queuedMessageCount > 0 && (
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <MessageSquare className="h-4 w-4" />
                <span>{queuedMessageCount} queued messages</span>
              </div>
              {onClearQueue && (
                <Button variant="ghost" size="sm" onClick={onClearQueue}>
                  Clear Queue
                </Button>
              )}
            </div>
          )}
        </div>
      )}

      {connectionState.status === 'error' && connectionState.lastError && (
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>
            {connectionState.lastError}
          </AlertDescription>
        </Alert>
      )}
    </div>
  )
}

export interface ConnectionIndicatorProps {
  isConnected: boolean
  className?: string
  size?: 'sm' | 'md' | 'lg'
}

export function ConnectionIndicator({ 
  isConnected, 
  className,
  size = 'md' 
}: ConnectionIndicatorProps) {
  const sizeClasses = {
    sm: 'h-2 w-2',
    md: 'h-3 w-3',
    lg: 'h-4 w-4'
  }

  return (
    <div
      className={cn(
        'rounded-full',
        sizeClasses[size],
        isConnected 
          ? 'bg-green-500 animate-pulse' 
          : 'bg-red-500',
        className
      )}
      title={isConnected ? 'Connected' : 'Disconnected'}
    />
  )
}

export interface ReconnectionProgressProps {
  attempt: number
  maxAttempts: number
  isReconnecting: boolean
  className?: string
}

export function ReconnectionProgress({
  attempt,
  maxAttempts,
  isReconnecting,
  className
}: ReconnectionProgressProps) {
  if (!isReconnecting) {
    return null
  }

  const progress = (attempt / maxAttempts) * 100

  return (
    <div className={cn('space-y-2', className)}>
      <div className="flex items-center justify-between text-sm">
        <span className="text-muted-foreground">Reconnecting...</span>
        <span className="text-muted-foreground">{attempt}/{maxAttempts}</span>
      </div>
      <Progress value={progress} className="h-2" />
    </div>
  )
}

export interface ConnectionErrorProps {
  error: string
  onRetry?: () => void
  onDismiss?: () => void
  className?: string
}

export function ConnectionError({
  error,
  onRetry,
  onDismiss,
  className
}: ConnectionErrorProps) {
  return (
    <Alert variant="destructive" className={className}>
      <AlertCircle className="h-4 w-4" />
      <AlertDescription className="flex items-center justify-between">
        <span>{error}</span>
        <div className="flex gap-2">
          {onRetry && (
            <Button variant="outline" size="sm" onClick={onRetry}>
              <RefreshCw className="h-4 w-4 mr-1" />
              Retry
            </Button>
          )}
          {onDismiss && (
            <Button variant="ghost" size="sm" onClick={onDismiss}>
              Dismiss
            </Button>
          )}
        </div>
      </AlertDescription>
    </Alert>
  )
}