import { useEffect } from 'react'
import * as z from 'zod'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import type { Task } from '@/types/task'
import { getStatusTitle } from '@/lib/kanban'
import { useUpdateTask } from '@/hooks/use-tasks'
import { Button } from '@/components/ui/button'
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
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'

const taskEditSchema = z.object({
  title: z.string().min(1, 'Title is required').max(200, 'Title too long'),
  description: z.string().max(1000, 'Description too long').optional(),
  status: z.enum([
    'TODO',
    'PLANNING',
    'PLAN_REVIEWING',
    'IMPLEMENTING',
    'CODE_REVIEWING',
    'DONE',
    'CANCELLED',
  ]),
  plan: z.string().max(2000, 'Plan too long').optional(),
  branch_name: z.string().max(100, 'Branch name too long').optional(),
  pr_url: z.string().url('Invalid URL').optional().or(z.literal('')),
})

type TaskEditFormValues = z.infer<typeof taskEditSchema>

interface TaskEditFormProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  task: Task
  onSave?: (task: Task) => void
}

export function TaskEditForm({
  open,
  onOpenChange,
  task,
  onSave,
}: TaskEditFormProps) {
  const updateTaskMutation = useUpdateTask()

  const form = useForm<TaskEditFormValues>({
    resolver: zodResolver(taskEditSchema),
    defaultValues: {
      title: '',
      description: '',
      status: 'TODO',
      plan: '',
      branch_name: '',
      pr_url: '',
    },
  })

  // Reset form when task changes or modal opens/closes
  useEffect(() => {
    if (open && task) {
      form.reset({
        title: task.title,
        description: task.description || '',
        status: task.status,
        plan: task.plan || '',
        branch_name: task.branch_name || '',
        pr_url: task.pr_url || '',
      })
    }
  }, [open, task, form])

  const isLoading = updateTaskMutation.isPending

  const onSubmit = async (values: TaskEditFormValues) => {
    try {
      const updatedTask = await updateTaskMutation.mutateAsync({
        taskId: task.id,
        updates: {
          title: values.title,
          description: values.description || '',
          status: values.status,
          plan: values.plan || '',
          branch_name: values.branch_name || '',
          pr_url: values.pr_url || '',
        },
      })

      onSave?.(updatedTask)
      onOpenChange(false)
    } catch (error) {
      // Error is handled by the mutation hooks
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='max-h-[90vh] overflow-y-auto sm:max-w-[600px]'>
        <DialogHeader>
          <DialogTitle>Edit Task</DialogTitle>
          <DialogDescription>
            Update the task details and configuration.
          </DialogDescription>
        </DialogHeader>

        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className='space-y-4'>
            <FormField
              control={form.control}
              name='title'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Title</FormLabel>
                  <FormControl>
                    <Input
                      placeholder='Enter task title...'
                      {...field}
                      disabled={isLoading}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name='status'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Status</FormLabel>
                  <Select
                    onValueChange={field.onChange}
                    defaultValue={field.value}
                  >
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue placeholder='Select status' />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      {(
                        [
                          'TODO',
                          'PLANNING',
                          'PLAN_REVIEWING',
                          'IMPLEMENTING',
                          'CODE_REVIEWING',
                          'DONE',
                          'CANCELLED',
                        ] as const
                      ).map((status) => (
                        <SelectItem key={status} value={status}>
                          {getStatusTitle(status)}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name='description'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Description</FormLabel>
                  <FormControl>
                    <Textarea
                      placeholder='Enter task description...'
                      className='min-h-[100px]'
                      {...field}
                      disabled={isLoading}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name='plan'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Plan</FormLabel>
                  <FormControl>
                    <Textarea
                      placeholder='Enter implementation plan...'
                      className='min-h-[120px]'
                      {...field}
                      disabled={isLoading}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <div className='grid grid-cols-2 gap-4'>
              <FormField
                control={form.control}
                name='branch_name'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Branch Name</FormLabel>
                    <FormControl>
                      <Input
                        placeholder='feature/task-name'
                        {...field}
                        disabled={isLoading}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name='pr_url'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Pull Request URL</FormLabel>
                    <FormControl>
                      <Input
                        placeholder='https://github.com/...'
                        {...field}
                        disabled={isLoading}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>

            <DialogFooter>
              <Button
                type='button'
                variant='outline'
                onClick={() => onOpenChange(false)}
                disabled={isLoading}
              >
                Cancel
              </Button>
              <Button type='submit' disabled={isLoading}>
                {isLoading ? 'Saving...' : 'Update Task'}
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  )
}
