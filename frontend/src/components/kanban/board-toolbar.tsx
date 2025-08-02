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

interface BoardToolbarProps {
  onCreateTask?: () => void
  onRefresh?: () => void
  isCompactView?: boolean
  onToggleCompactView?: () => void
  isLoading?: boolean
}

export function BoardToolbar({
  onCreateTask,
  onRefresh,
  isCompactView = false,
  onToggleCompactView,
  isLoading = false,
}: BoardToolbarProps) {
  const [showHiddenColumns, setShowHiddenColumns] = useState(false)

  return (
    <div className="flex items-center justify-between p-4 bg-white border-b">
      <div className="flex items-center gap-3">
        <h1 className="text-2xl font-bold text-gray-900">Task Board</h1>
        
        <Separator orientation="vertical" className="h-6" />
        
        <Button
          variant="outline"
          size="sm"
          onClick={onRefresh}
          disabled={isLoading}
        >
          <RefreshCw className={`h-4 w-4 mr-2 ${isLoading ? 'animate-spin' : ''}`} />
          Refresh
        </Button>
      </div>

      <div className="flex items-center gap-2">
        {/* View Toggle */}
        <div className="flex items-center border rounded-md">
          <Button
            variant={isCompactView ? "ghost" : "secondary"}
            size="sm"
            onClick={() => onToggleCompactView?.()}
            className="rounded-r-none border-r"
          >
            <LayoutGrid className="h-4 w-4" />
          </Button>
          <Button
            variant={isCompactView ? "secondary" : "ghost"}
            size="sm"
            onClick={() => onToggleCompactView?.()}
            className="rounded-l-none"
          >
            <List className="h-4 w-4" />
          </Button>
        </div>

        {/* Create Task */}
        <Button onClick={onCreateTask}>
          <Plus className="h-4 w-4 mr-2" />
          New Task
        </Button>

        {/* More Options */}
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="outline" size="sm">
              <MoreHorizontal className="h-4 w-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="w-48">
            <DropdownMenuItem>
              <Settings className="h-4 w-4 mr-2" />
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
              <Download className="h-4 w-4 mr-2" />
              Export Tasks
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </div>
  )
}