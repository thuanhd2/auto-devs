import { z } from 'zod'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { useNavigate } from '@tanstack/react-router'
import { Plus } from 'lucide-react'
import { useCreateProject } from '@/hooks/use-projects'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'

const createProjectSchema = z.object({
  name: z
    .string()
    .min(1, 'Project name is required')
    .max(100, 'Project name must be less than 100 characters'),
  description: z
    .string()
    .max(500, 'Description must be less than 500 characters')
    .optional(),
  worktree_base_path: z
    .string()
    .min(1, 'Worktree base path is required')
    .max(500, 'Worktree base path must be less than 500 characters'),
})

type CreateProjectFormData = z.infer<typeof createProjectSchema>

interface ProjectCreateModalProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function ProjectCreateModal({
  open,
  onOpenChange,
}: ProjectCreateModalProps) {
  const navigate = useNavigate()
  const createProject = useCreateProject()

  const form = useForm<CreateProjectFormData>({
    resolver: zodResolver(createProjectSchema),
    defaultValues: {
      name: '',
      description: '',
      worktree_base_path: '',
    },
  })

  const onSubmit = async (data: CreateProjectFormData) => {
    try {
      const project = await createProject.mutateAsync({
        name: data.name,
        description: data.description || undefined,
        worktree_base_path: data.worktree_base_path,
      })

      // Close modal and reset form
      onOpenChange(false)
      form.reset()

      // Navigate to the new project
      navigate({
        to: '/projects/$projectId',
        params: { projectId: project.id },
      })
    } catch (error) {
      // Error handling is done in the mutation hook
    }
  }

  const handleClose = () => {
    onOpenChange(false)
    form.reset()
  }

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent className='sm:max-w-[425px]'>
        <DialogHeader>
          <DialogTitle className='flex items-center gap-2'>
            <Plus className='h-5 w-5' />
            Create New Project
          </DialogTitle>
          <DialogDescription>
            Set up a new development project to start managing tasks
          </DialogDescription>
        </DialogHeader>

        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className='space-y-4'>
            <FormField
              control={form.control}
              name='name'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Project Name</FormLabel>
                  <FormControl>
                    <Input placeholder='My Awesome Project' {...field} />
                  </FormControl>
                  <FormDescription>
                    A descriptive name for your project
                  </FormDescription>
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
                      placeholder='Brief description of what this project does...'
                      className='resize-none'
                      rows={3}
                      {...field}
                    />
                  </FormControl>
                  <FormDescription>
                    Optional description to help identify this project
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name='worktree_base_path'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Worktree Base Path</FormLabel>
                  <FormControl>
                    <Input placeholder='/path/to/your/project' {...field} />
                  </FormControl>
                  <FormDescription>
                    Base path for Git worktree operations
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            <div className='flex justify-end gap-3 pt-4'>
              <Button type='button' variant='outline' onClick={handleClose}>
                Cancel
              </Button>
              <Button type='submit' disabled={createProject.isPending}>
                {createProject.isPending ? 'Creating...' : 'Create Project'}
              </Button>
            </div>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  )
}
