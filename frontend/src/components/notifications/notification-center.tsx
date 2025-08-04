import { useState } from 'react'
import { Bell, X, CheckCircle, AlertCircle, AlertTriangle, Info, Loader2, Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetTrigger } from '@/components/ui/sheet'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Separator } from '@/components/ui/separator'
import { useNotifications, type NotificationItem } from '@/hooks/use-notifications'
import { cn } from '@/lib/utils'
import { formatDistanceToNow } from 'date-fns'

const NotificationIcon = ({ type }: { type: NotificationItem['type'] }) => {
  const iconProps = { className: 'h-4 w-4' }
  
  switch (type) {
    case 'success':
      return <CheckCircle {...iconProps} className="h-4 w-4 text-green-500" />
    case 'error':
      return <AlertCircle {...iconProps} className="h-4 w-4 text-red-500" />
    case 'warning':
      return <AlertTriangle {...iconProps} className="h-4 w-4 text-yellow-500" />
    case 'info':
      return <Info {...iconProps} className="h-4 w-4 text-blue-500" />
    case 'loading':
      return <Loader2 {...iconProps} className="h-4 w-4 text-gray-500 animate-spin" />
    default:
      return <Info {...iconProps} className="h-4 w-4 text-gray-500" />
  }
}

interface NotificationItemComponentProps {
  notification: NotificationItem
  onRemove: (id: string) => void
}

function NotificationItemComponent({ notification, onRemove }: NotificationItemComponentProps) {
  return (
    <div className="group relative flex items-start gap-3 rounded-lg border p-3 transition-colors hover:bg-accent/50">
      <NotificationIcon type={notification.type} />
      
      <div className="flex-1 space-y-1">
        <div className="flex items-start justify-between gap-2">
          <p className="text-sm font-medium leading-5">{notification.title}</p>
          <div className="flex items-center gap-1">
            <span className="text-xs text-muted-foreground">
              {formatDistanceToNow(notification.timestamp, { addSuffix: true })}
            </span>
            <Button
              variant="ghost"
              size="sm"
              className="h-6 w-6 p-0 opacity-0 transition-opacity group-hover:opacity-100"
              onClick={() => onRemove(notification.id)}
            >
              <X className="h-3 w-3" />
              <span className="sr-only">Remove notification</span>
            </Button>
          </div>
        </div>
        
        {notification.description && (
          <p className="text-xs text-muted-foreground">{notification.description}</p>
        )}
        
        {notification.action && (
          <Button
            variant="outline"
            size="sm"
            className="mt-2 h-7 px-2 text-xs"
            onClick={notification.action.onClick}
          >
            {notification.action.label}
          </Button>
        )}
      </div>
    </div>
  )
}

interface NotificationCenterProps {
  className?: string
}

export function NotificationCenter({ className }: NotificationCenterProps) {
  const [isOpen, setIsOpen] = useState(false)
  const { history, clearHistory, removeFromHistory } = useNotifications()
  
  const unreadCount = history.length
  const hasNotifications = history.length > 0

  const handleClearAll = () => {
    clearHistory()
  }

  return (
    <Sheet open={isOpen} onOpenChange={setIsOpen}>
      <SheetTrigger asChild>
        <Button
          variant="ghost"
          size="sm"
          className={cn('relative h-8 w-8 p-0', className)}
        >
          <Bell className="h-4 w-4" />
          {unreadCount > 0 && (
            <Badge
              variant="destructive"
              className="absolute -right-1 -top-1 h-5 w-5 rounded-full p-0 text-xs flex items-center justify-center"
            >
              {unreadCount > 99 ? '99+' : unreadCount}
            </Badge>
          )}
          <span className="sr-only">Open notifications</span>
        </Button>
      </SheetTrigger>
      
      <SheetContent className="w-full sm:max-w-md">
        <SheetHeader className="space-y-4">
          <div className="flex items-center justify-between">
            <SheetTitle>Notifications</SheetTitle>
            {hasNotifications && (
              <Button
                variant="ghost"
                size="sm"
                onClick={handleClearAll}
                className="h-8 px-2 text-xs"
              >
                <Trash2 className="h-3 w-3 mr-1" />
                Clear all
              </Button>
            )}
          </div>
          
          {hasNotifications && (
            <div className="flex items-center gap-2 text-sm text-muted-foreground">
              <Badge variant="secondary" className="text-xs">
                {unreadCount} notification{unreadCount === 1 ? '' : 's'}
              </Badge>
            </div>
          )}
        </SheetHeader>

        <Separator className="my-4" />

        <ScrollArea className="h-[calc(100vh-200px)]">
          {!hasNotifications ? (
            <div className="flex flex-col items-center justify-center py-12 text-center">
              <Bell className="h-12 w-12 text-muted-foreground/30" />
              <h3 className="mt-4 text-sm font-medium">No notifications</h3>
              <p className="text-sm text-muted-foreground">
                You're all caught up! New notifications will appear here.
              </p>
            </div>
          ) : (
            <div className="space-y-3">
              {history.map((notification) => (
                <NotificationItemComponent
                  key={notification.id}
                  notification={notification}
                  onRemove={removeFromHistory}
                />
              ))}
            </div>
          )}
        </ScrollArea>
      </SheetContent>
    </Sheet>
  )
}