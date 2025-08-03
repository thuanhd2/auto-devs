import React, { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Badge } from '@/components/ui/badge'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { ConnectionStatus } from '@/components/ui/connection-status'
import { 
  useWebSocketContext, 
  useWebSocketConnection, 
  useWebSocketProject,
  useWebSocketTaskUpdates 
} from '@/context/websocket-context'
import { taskOptimisticUpdates } from '@/services/optimisticUpdates'
import { 
  Wifi, 
  Send, 
  Users, 
  Activity,
  MessageSquare,
  CheckCircle,
  AlertCircle
} from 'lucide-react'

export function WebSocketExample() {
  const [customMessage, setCustomMessage] = useState('')
  const [testProjectId, setTestProjectId] = useState('test-project-123')
  const [taskUpdates, setTaskUpdates] = useState<any[]>([])
  
  // WebSocket hooks
  const { send, onlineUsers, userCount } = useWebSocketContext()
  const { 
    isConnected, 
    isConnecting, 
    connectionState, 
    connect, 
    disconnect, 
    reconnect,
    queuedMessageCount,
    clearMessageQueue 
  } = useWebSocketConnection()
  
  const { 
    currentProjectId, 
    setCurrentProjectId,
    subscribeToProject,
    unsubscribeFromProject 
  } = useWebSocketProject()

  // Listen for task updates
  useWebSocketTaskUpdates(
    currentProjectId || undefined,
    (task, changes) => {
      setTaskUpdates(prev => [...prev.slice(-9), { 
        timestamp: new Date().toLocaleTimeString(),
        type: changes?.created ? 'created' : changes?.deleted ? 'deleted' : 'updated',
        task,
        changes 
      }])
    }
  )

  const handleSendCustomMessage = async () => {
    if (!customMessage.trim()) return
    
    try {
      const message = JSON.parse(customMessage)
      await send(message)
      setCustomMessage('')
    } catch (error) {
      console.error('Invalid JSON:', error)
    }
  }

  const handleSubscribeToProject = async () => {
    if (!testProjectId.trim()) return
    
    try {
      await setCurrentProjectId(testProjectId)
    } catch (error) {
      console.error('Failed to subscribe to project:', error)
    }
  }

  const handleUnsubscribeFromProject = async () => {
    try {
      await setCurrentProjectId(null)
    } catch (error) {
      console.error('Failed to unsubscribe from project:', error)
    }
  }

  const handleSimulateOptimisticUpdate = () => {
    const mockTask = {
      id: `task-${Date.now()}`,
      title: 'Test Task',
      status: 'TODO',
      project_id: currentProjectId
    }

    taskOptimisticUpdates.updateTaskStatus(
      mockTask.id,
      'IN_PROGRESS',
      mockTask,
      (updatedTask) => {
        setTaskUpdates(prev => [...prev.slice(-9), {
          timestamp: new Date().toLocaleTimeString(),
          type: 'optimistic_update',
          task: updatedTask,
          changes: { status: { old: 'TODO', new: 'IN_PROGRESS' } }
        }])
      },
      (confirmedTask) => {
        console.log('Optimistic update confirmed:', confirmedTask)
      },
      (originalTask) => {
        console.log('Optimistic update reverted:', originalTask)
      }
    )
  }

  return (
    <div className="space-y-6 p-6 max-w-4xl mx-auto">
      <div className="text-center">
        <h1 className="text-3xl font-bold mb-2">WebSocket Integration Example</h1>
        <p className="text-muted-foreground">
          Test and demonstrate the WebSocket client functionality
        </p>
      </div>

      {/* Connection Status */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Wifi className="h-5 w-5" />
            Connection Status
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <ConnectionStatus
            connectionState={connectionState}
            queuedMessageCount={queuedMessageCount}
            onReconnect={reconnect}
            onClearQueue={clearMessageQueue}
            variant="full"
            showDetails={true}
          />
          
          <div className="flex gap-2">
            <Button 
              onClick={connect} 
              disabled={isConnected || isConnecting}
              size="sm"
            >
              Connect
            </Button>
            <Button 
              onClick={disconnect} 
              disabled={!isConnected}
              variant="outline"
              size="sm"
            >
              Disconnect
            </Button>
            <Button 
              onClick={reconnect} 
              disabled={isConnecting}
              variant="outline"
              size="sm"
            >
              Reconnect
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Project Subscription */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Users className="h-5 w-5" />
            Project Subscription
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex gap-2">
            <div className="flex-1">
              <Label htmlFor="projectId">Project ID</Label>
              <Input
                id="projectId"
                value={testProjectId}
                onChange={(e) => setTestProjectId(e.target.value)}
                placeholder="Enter project ID"
              />
            </div>
          </div>
          
          <div className="flex gap-2">
            <Button 
              onClick={handleSubscribeToProject}
              disabled={!isConnected || !testProjectId.trim()}
              size="sm"
            >
              Subscribe
            </Button>
            <Button 
              onClick={handleUnsubscribeFromProject}
              disabled={!isConnected || !currentProjectId}
              variant="outline"
              size="sm"
            >
              Unsubscribe
            </Button>
          </div>

          {currentProjectId && (
            <Alert>
              <CheckCircle className="h-4 w-4" />
              <AlertDescription>
                Subscribed to project: <code>{currentProjectId}</code>
              </AlertDescription>
            </Alert>
          )}

          <div className="text-sm text-muted-foreground">
            <div>Online users: {userCount}</div>
            {userCount > 0 && (
              <div className="mt-1">
                {Array.from(onlineUsers.entries()).map(([userId, user]) => (
                  <Badge key={userId} variant="secondary" className="mr-1">
                    {user.username}
                  </Badge>
                ))}
              </div>
            )}
          </div>
        </CardContent>
      </Card>

      {/* Custom Message */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Send className="h-5 w-5" />
            Send Custom Message
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div>
            <Label htmlFor="customMessage">JSON Message</Label>
            <Textarea
              id="customMessage"
              value={customMessage}
              onChange={(e) => setCustomMessage(e.target.value)}
              placeholder='{"type": "test", "data": {"message": "Hello WebSocket!"}}'
              className="font-mono text-sm"
              rows={4}
            />
          </div>
          
          <Button 
            onClick={handleSendCustomMessage}
            disabled={!isConnected || !customMessage.trim()}
            className="w-full"
          >
            <Send className="h-4 w-4 mr-2" />
            Send Message
          </Button>
        </CardContent>
      </Card>

      {/* Optimistic Updates */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Activity className="h-5 w-5" />
            Optimistic Updates
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <Button 
            onClick={handleSimulateOptimisticUpdate}
            disabled={!currentProjectId}
            className="w-full"
          >
            Simulate Task Status Update
          </Button>
          
          <div className="text-sm text-muted-foreground">
            This will create an optimistic update that changes a mock task status.
            The update will automatically revert after 10 seconds if not confirmed.
          </div>
        </CardContent>
      </Card>

      {/* Task Updates Log */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <MessageSquare className="h-5 w-5" />
            Real-time Task Updates
          </CardTitle>
        </CardHeader>
        <CardContent>
          {taskUpdates.length === 0 ? (
            <div className="text-center text-muted-foreground py-8">
              No task updates received yet
            </div>
          ) : (
            <div className="space-y-2 max-h-64 overflow-y-auto">
              {taskUpdates.map((update, index) => (
                <div 
                  key={index}
                  className="p-3 border rounded bg-muted/50"
                >
                  <div className="flex items-center justify-between mb-2">
                    <Badge 
                      variant={
                        update.type === 'created' ? 'default' :
                        update.type === 'deleted' ? 'destructive' :
                        update.type === 'optimistic_update' ? 'secondary' :
                        'outline'
                      }
                    >
                      {update.type}
                    </Badge>
                    <span className="text-xs text-muted-foreground">
                      {update.timestamp}
                    </span>
                  </div>
                  <div className="text-sm">
                    <div className="font-medium">{update.task.title}</div>
                    {update.changes && (
                      <div className="text-muted-foreground">
                        Changes: {JSON.stringify(update.changes, null, 2)}
                      </div>
                    )}
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}