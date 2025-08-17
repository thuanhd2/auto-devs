import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { ArrowRight, Loader2 } from 'lucide-react'
import type { Task, TaskStatus } from '@/types/task'
import { canTransitionTo, getStatusTitle, getStatusColor } from '@/lib/kanban'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

const changeStatusSchema = z.object({
  status: z.enum(['TODO', 'PLANNING', 'PLAN_REVIEWING', 'IMPLEMENTING', 'CODE_REVIEWING', 'DONE', 'CANCELLED']),
})

type ChangeStatusFormData = z.infer<typeof changeStatusSchema>

interface ChangeStatusDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  task: Task
  onStatusChange: (newStatus: TaskStatus) => Promise<void>
}

export function ChangeStatusDialog({
  open,
  onOpenChange,
  task,
  onStatusChange,
}: ChangeStatusDialogProps) {
  const [isSubmitting, setIsSubmitting] = useState(false)

  const form = useForm<ChangeStatusFormData>({
    resolver: zodResolver(changeStatusSchema),
    defaultValues: {
      status: task.status,
    },
  })

  // Get available status transitions from current status
  const availableStatuses: TaskStatus[] = ['TODO', 'PLANNING', 'PLAN_REVIEWING', 'IMPLEMENTING', 'CODE_REVIEWING', 'DONE', 'CANCELLED']
    .filter((status) => status !== task.status && canTransitionTo(task.status, status as TaskStatus)) as TaskStatus[]

  const handleSubmit = async (data: ChangeStatusFormData) => {
    if (data.status === task.status) {
      onOpenChange(false)
      return
    }

    setIsSubmitting(true)
    try {
      await onStatusChange(data.status)
      onOpenChange(false)
      form.reset()
    } catch (error) {
      // Error is handled by the parent component's mutation
      console.error('Failed to change status:', error)
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleCancel = () => {
    form.reset()
    onOpenChange(false)
  }

  const selectedStatus = form.watch('status')
  const currentStatusTitle = getStatusTitle(task.status)
  const currentStatusColor = getStatusColor(task.status)
  const selectedStatusTitle = selectedStatus ? getStatusTitle(selectedStatus) : ''
  const selectedStatusColor = selectedStatus ? getStatusColor(selectedStatus) : ''

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>Change Task Status</DialogTitle>
          <DialogDescription>
            Change the status of "{task.title}" to move it through your workflow.
          </DialogDescription>
        </DialogHeader>

        <Form {...form}>
          <form onSubmit={form.handleSubmit(handleSubmit)} className="space-y-6">
            {/* Current Status Display */}
            <div className="space-y-2">
              <label className="text-sm font-medium">Current Status</label>
              <div className="flex items-center gap-2">
                <Badge className={currentStatusColor} variant="outline">
                  {currentStatusTitle}
                </Badge>
              </div>
            </div>

            {/* Status Selection */}
            <FormField
              control={form.control}
              name="status"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>New Status</FormLabel>
                  <Select
                    onValueChange={field.onChange}
                    value={field.value}
                    disabled={availableStatuses.length === 0}
                  >
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue placeholder="Select new status" />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      {availableStatuses.map((status) => (
                        <SelectItem key={status} value={status}>
                          <div className="flex items-center gap-2">
                            <Badge 
                              className={`${getStatusColor(status)} text-xs`} 
                              variant="outline"
                            >
                              {getStatusTitle(status)}
                            </Badge>
                          </div>
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  <FormMessage />
                </FormItem>
              )}
            />

            {/* Status Transition Preview */}
            {selectedStatus && selectedStatus !== task.status && (
              <div className="flex items-center gap-2 p-3 bg-muted rounded-lg">
                <Badge className={currentStatusColor} variant="outline">
                  {currentStatusTitle}
                </Badge>
                <ArrowRight className="h-4 w-4 text-muted-foreground" />
                <Badge className={selectedStatusColor} variant="outline">
                  {selectedStatusTitle}
                </Badge>
              </div>
            )}

            {/* No Available Transitions Message */}
            {availableStatuses.length === 0 && (
              <div className="p-3 bg-yellow-50 border border-yellow-200 rounded-lg">
                <p className="text-sm text-yellow-800">
                  No status transitions are available from the current status "{currentStatusTitle}".
                </p>
              </div>
            )}

            <DialogFooter>
              <Button
                type="button"
                variant="outline"
                onClick={handleCancel}
                disabled={isSubmitting}
              >
                Cancel
              </Button>
              <Button
                type="submit"
                disabled={
                  isSubmitting || 
                  !selectedStatus || 
                  selectedStatus === task.status ||
                  availableStatuses.length === 0
                }
              >
                {isSubmitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                Change Status
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  )
}