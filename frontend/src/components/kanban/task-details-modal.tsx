import { formatDistanceToNow, format } from 'date-fns'
import { Calendar, GitBranch, ExternalLink, Edit, Trash2 } from 'lucide-react'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'
import { getStatusColor, getStatusTitle } from '@/lib/kanban'
import type { Task } from '@/types/task'

interface TaskDetailsModalProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  task: Task | null
  onEdit?: (task: Task) => void
  onDelete?: (taskId: string) => void
}

export function TaskDetailsModal({
  open,
  onOpenChange,
  task,
  onEdit,
  onDelete,
}: TaskDetailsModalProps) {
  if (!task) return null

  const statusTitle = getStatusTitle(task.status)
  const statusColor = getStatusColor(task.status)
  const createdAgo = formatDistanceToNow(new Date(task.created_at), { addSuffix: true })
  const updatedAgo = formatDistanceToNow(new Date(task.updated_at), { addSuffix: true })

  const handleEdit = () => {
    onEdit?.(task)
    onOpenChange(false)
  }

  const handleDelete = () => {
    if (confirm('Are you sure you want to delete this task?')) {
      onDelete?.(task.id)
      onOpenChange(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <div className="flex items-start justify-between gap-4">
            <div className="flex-1">
              <DialogTitle className="text-xl font-semibold mb-2">
                {task.title}
              </DialogTitle>
              <div className="flex items-center gap-2">
                <Badge className={statusColor}>
                  {statusTitle}
                </Badge>
              </div>
            </div>
            
            <div className="flex items-center gap-1">
              <Button
                variant="outline"
                size="sm"
                onClick={handleEdit}
              >
                <Edit className="h-4 w-4" />
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={handleDelete}
                className="text-red-600 hover:text-red-700"
              >
                <Trash2 className="h-4 w-4" />
              </Button>
            </div>
          </div>
        </DialogHeader>

        <div className="space-y-4">
          {/* Description */}
          {task.description && (
            <div>
              <h4 className="font-medium text-sm text-gray-700 mb-2">Description</h4>
              <p className="text-sm text-gray-600 whitespace-pre-wrap">
                {task.description}
              </p>
            </div>
          )}

          {/* Plan */}
          {task.plan && (
            <div>
              <h4 className="font-medium text-sm text-gray-700 mb-2">Plan</h4>
              <div className="text-sm text-gray-600 bg-gray-50 p-3 rounded border whitespace-pre-wrap">
                {task.plan}
              </div>
            </div>
          )}

          <Separator />

          {/* Metadata */}
          <div className="grid grid-cols-2 gap-4 text-sm">
            <div>
              <h4 className="font-medium text-gray-700 mb-1">Created</h4>
              <div className="flex items-center gap-1 text-gray-600">
                <Calendar className="h-3 w-3" />
                <span>{createdAgo}</span>
              </div>
              <p className="text-xs text-gray-500 mt-1">
                {format(new Date(task.created_at), 'PPpp')}
              </p>
            </div>
            
            <div>
              <h4 className="font-medium text-gray-700 mb-1">Last Updated</h4>
              <div className="flex items-center gap-1 text-gray-600">
                <Calendar className="h-3 w-3" />
                <span>{updatedAgo}</span>
              </div>
              <p className="text-xs text-gray-500 mt-1">
                {format(new Date(task.updated_at), 'PPpp')}
              </p>
            </div>
          </div>

          {/* Git Information */}
          {(task.branch_name || task.pr_url) && (
            <>
              <Separator />
              <div>
                <h4 className="font-medium text-sm text-gray-700 mb-2">Git Information</h4>
                <div className="space-y-2">
                  {task.branch_name && (
                    <div className="flex items-center gap-2 text-sm">
                      <GitBranch className="h-4 w-4 text-gray-500" />
                      <span className="font-mono text-gray-600">{task.branch_name}</span>
                    </div>
                  )}
                  
                  {task.pr_url && (
                    <div className="flex items-center gap-2 text-sm">
                      <ExternalLink className="h-4 w-4 text-gray-500" />
                      <a
                        href={task.pr_url}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-blue-600 hover:text-blue-700 hover:underline"
                      >
                        View Pull Request
                      </a>
                    </div>
                  )}
                </div>
              </div>
            </>
          )}

          {/* Completion Date */}
          {task.completed_at && (
            <>
              <Separator />
              <div>
                <h4 className="font-medium text-sm text-gray-700 mb-1">Completed</h4>
                <p className="text-sm text-gray-600">
                  {format(new Date(task.completed_at), 'PPpp')}
                </p>
              </div>
            </>
          )}
        </div>
      </DialogContent>
    </Dialog>
  )
}