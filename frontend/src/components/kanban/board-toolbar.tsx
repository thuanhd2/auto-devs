import { useState } from 'react'
import {
  Plus,
  RefreshCw,
  MoreHorizontal,
  LayoutGrid,
  List,
  Settings,
  Download,
} from 'lucide-react'
import { useWebSocketConnection } from '@/context/websocket-context'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
  DropdownMenuCheckboxItem,
} from '@/components/ui/dropdown-menu'
import { Separator } from '@/components/ui/separator'
import { UserPresence } from '@/components/collaboration/user-presence'

interface BoardToolbarProps {
  onCreateTask?: () => void
  onRefresh?: () => void
  isCompactView?: boolean
  onToggleCompactView?: () => void
  isLoading?: boolean
  projectId?: string
}

export function BoardToolbar({
  onCreateTask,
  onRefresh,
  isCompactView = false,
  onToggleCompactView,
  isLoading = false,
  projectId,
}: BoardToolbarProps) {
  const [showHiddenColumns, setShowHiddenColumns] = useState(false)
  const { isConnected, queuedMessageCount } = useWebSocketConnection()

  return (
    <div className='flex items-center justify-between border-b p-4'>
      <div className='flex items-center gap-3'>
        <h1 className='text-2xl font-bold'>Task Board</h1>
        <Separator orientation='vertical' className='h-6' />

        {/* Connection Status */}
        <div className='flex items-center gap-2'>
          <div
            className={`h-2 w-2 rounded-full ${
              isConnected ? 'bg-green-500' : 'bg-red-500'
            } animate-pulse`}
          />
          <span className='text-sm'>{isConnected ? 'Live' : 'Offline'}</span>
          {queuedMessageCount > 0 && (
            <Badge variant='outline' className='text-xs'>
              {queuedMessageCount} queued
            </Badge>
          )}
        </div>

        <Separator orientation='vertical' className='h-6' />

        <Button
          variant='outline'
          size='sm'
          onClick={onRefresh}
          disabled={isLoading}
        >
          <RefreshCw
            className={`mr-2 h-4 w-4 ${isLoading ? 'animate-spin' : ''}`}
          />
          Refresh
        </Button>
      </div>

      <div className='flex items-center gap-2'>
        {/* User Presence */}
        {projectId && (
          <>
            <UserPresence
              projectId={projectId}
              showDetails={false}
              maxAvatars={3}
            />
            <Separator orientation='vertical' className='h-6' />
          </>
        )}

        {/* Create Task */}
        <Button onClick={onCreateTask} variant='default' size='sm'>
          <Plus className='mr-2 h-4 w-4' />
          New Task
        </Button>

        {/* More Options */}
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant='outline' size='sm'>
              <MoreHorizontal className='h-4 w-4' />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align='end' className='w-48'>
            <DropdownMenuItem>
              <Settings className='mr-2 h-4 w-4' />
              Board Settings
            </DropdownMenuItem>

            <DropdownMenuSeparator />

            <DropdownMenuCheckboxItem
              checked={showHiddenColumns}
              onCheckedChange={setShowHiddenColumns}
            >
              Show Hidden Columns
            </DropdownMenuCheckboxItem>

            <DropdownMenuSeparator />

            <DropdownMenuItem>
              <Download className='mr-2 h-4 w-4' />
              Export Tasks
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </div>
  )
}
