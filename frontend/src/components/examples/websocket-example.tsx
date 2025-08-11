import { useState, useEffect } from 'react'
import type { CentrifugeMessage } from '@/services/websocketService'
import { useWebSocketConnection } from '@/context/websocket-context'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'

export function WebSocketExample() {
  const [messages, setMessages] = useState<CentrifugeMessage[]>([])
  const [inputMessage, setInputMessage] = useState('')
  const [isConnected, setIsConnected] = useState(false)
  const [connectionStatus, setConnectionStatus] = useState('disconnected')

  const {
    isConnected: wsConnected,
    sendMessage,
    subscribe,
    unsubscribe,
  } = useWebSocketConnection()

  useEffect(() => {
    setIsConnected(wsConnected)
    setConnectionStatus(wsConnected ? 'connected' : 'disconnected')
  }, [wsConnected])

  useEffect(() => {
    const handleMessage = (message: CentrifugeMessage) => {
      setMessages((prev) => [...prev, message])
    }

    subscribe('test_channel', handleMessage)

    return () => {
      unsubscribe('test_channel', handleMessage)
    }
  }, [subscribe, unsubscribe])

  const handleSendMessage = () => {
    if (inputMessage.trim()) {
      sendMessage('test_channel', { text: inputMessage })
      setInputMessage('')
    }
  }

  const handleClearMessages = () => {
    setMessages([])
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'connected':
        return 'bg-green-500'
      case 'connecting':
        return 'bg-yellow-500'
      case 'disconnected':
        return 'bg-red-500'
      default:
        return 'bg-gray-500'
    }
  }

  return (
    <div className='space-y-4'>
      <Card>
        <CardHeader>
          <CardTitle className='flex items-center gap-2'>
            WebSocket Connection Status
            <div
              className={`h-3 w-3 rounded-full ${getStatusColor(connectionStatus)}`}
            />
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className='space-y-2'>
            <div className='flex items-center gap-2'>
              <span className='text-sm font-medium'>Status:</span>
              <Badge variant={isConnected ? 'default' : 'secondary'}>
                {connectionStatus}
              </Badge>
            </div>
            <div className='flex items-center gap-2'>
              <span className='text-sm font-medium'>Connected:</span>
              <Badge variant={isConnected ? 'default' : 'destructive'}>
                {isConnected ? 'Yes' : 'No'}
              </Badge>
            </div>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Send Message</CardTitle>
        </CardHeader>
        <CardContent>
          <div className='flex gap-2'>
            <Input
              value={inputMessage}
              onChange={(e) => setInputMessage(e.target.value)}
              placeholder='Type your message...'
              onKeyPress={(e) => e.key === 'Enter' && handleSendMessage()}
            />
            <Button onClick={handleSendMessage} disabled={!isConnected}>
              Send
            </Button>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className='flex items-center justify-between'>
            Messages
            <Button variant='outline' size='sm' onClick={handleClearMessages}>
              Clear
            </Button>
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className='max-h-64 space-y-2 overflow-y-auto'>
            {messages.length === 0 ? (
              <p className='text-muted-foreground text-sm'>No messages yet</p>
            ) : (
              messages.map((message, index) => (
                <div key={index} className='rounded border p-2 text-sm'>
                  <div className='font-medium'>{message.channel}</div>
                  <div className='text-muted-foreground'>
                    {JSON.stringify(message.data, null, 2)}
                  </div>
                </div>
              ))
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
