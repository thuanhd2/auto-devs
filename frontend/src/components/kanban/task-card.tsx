import { useState } from 'react'
import { MoreHorizontal, Calendar, GitBranch, ExternalLink } from 'lucide-react'
import { formatDistanceToNow } from 'date-fns'
import { Card, CardContent, CardHeader } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { getStatusColor, getStatusTitle } from '@/lib/kanban'
import { GitStatusBadge } from './git-status-badge'
import type { Task } from '@/types/task'

interface TaskCardProps {
  task: Task
  isDragging?: boolean
  onEdit?: (task: Task) => void
  onDelete?: (taskId: string) => void
  onViewDetails?: (task: Task) => void
}

export function TaskCard({ 
  task, 
  isDragging = false, 
  onEdit, 
  onDelete, 
  onViewDetails 
}: TaskCardProps) {
  const [isMenuOpen, setIsMenuOpen] = useState(false)

  const handleEdit = () => {
    setIsMenuOpen(false)
    onEdit?.(task)
  }

  const handleDelete = () => {
    setIsMenuOpen(false)
    onDelete?.(task.id)
  }

  const handleViewDetails = () => {
    setIsMenuOpen(false)
    onViewDetails?.(task)
  }

  const statusColor = getStatusColor(task.status)
  const statusTitle = getStatusTitle(task.status)
  const updatedAgo = formatDistanceToNow(new Date(task.updated_at), { addSuffix: true })

  return (
    <Card 
      className={`
        cursor-pointer transition-all duration-200 hover:shadow-md
        ${isDragging ? 'opacity-50 rotate-2 shadow-lg' : ''}
        ${statusColor} border-l-4
      `}
    >
      <CardHeader className="pb-2">
        <div className="flex items-start justify-between gap-2">
          <div className="flex-1 min-w-0">
            <h3 className="font-medium text-sm line-clamp-2 text-gray-900">
              {task.title}
            </h3>
            <div className="flex items-center gap-1 mt-1">
              <Badge variant="secondary" className="text-xs">
                {statusTitle}
              </Badge>
              {task.git_info && (
                <GitStatusBadge 
                  status={task.git_info.status}
                  branchName={task.git_info.branch_name}
                  variant="compact"
                  className="ml-1"
                />
              )}
            </div>
          </div>
          
          <DropdownMenu open={isMenuOpen} onOpenChange={setIsMenuOpen}>
            <DropdownMenuTrigger asChild>
              <Button
                variant="ghost"
                size="sm"
                className="h-6 w-6 p-0 hover:bg-gray-100"
                onClick={(e) => e.stopPropagation()}
              >
                <MoreHorizontal className="h-3 w-3" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="w-48">
              <DropdownMenuItem onClick={handleViewDetails}>
                View Details
              </DropdownMenuItem>
              <DropdownMenuItem onClick={handleEdit}>
                Edit Task
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem 
                onClick={handleDelete}
                className="text-red-600 focus:text-red-700"
              >
                Delete Task
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </CardHeader>
      
      <CardContent className="pt-0">
        {task.description && (
          <p className="text-xs text-gray-600 line-clamp-2 mb-3">
            {task.description}
          </p>
        )}
        
        <div className="flex items-center justify-between text-xs text-gray-500">
          <div className="flex items-center gap-1">
            <Calendar className="h-3 w-3" />
            <span>{updatedAgo}</span>
          </div>
          
          <div className="flex items-center gap-2">
            {(task.git_info?.branch_name || task.branch_name) && (
              <div className="flex items-center gap-1">
                <GitBranch className="h-3 w-3" />
                <span className="max-w-16 truncate">
                  {task.git_info?.branch_name || task.branch_name}
                </span>
              </div>
            )}
            
            {task.git_info?.worktree_path && (
              <div className="flex items-center gap-1 text-xs text-blue-600" title={task.git_info.worktree_path}>
                <span className="max-w-20 truncate">
                  {task.git_info.worktree_path.split('/').pop()}
                </span>
              </div>
            )}
            
            {(task.git_info?.pr_url || task.pr_url) && (
              <a
                href={task.git_info?.pr_url || task.pr_url}
                target="_blank"
                rel="noopener noreferrer"
                className="flex items-center gap-1 hover:text-blue-600 transition-colors"
                onClick={(e) => e.stopPropagation()}
              >
                <ExternalLink className="h-3 w-3" />
                <span>PR</span>
              </a>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  )
}