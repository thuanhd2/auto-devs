import { useEffect } from 'react'
import { z } from 'zod'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { Settings, Trash2 } from 'lucide-react'
import type { UpdateProjectRequest } from '@/types/project'
import { useProject, useUpdateProject, useDeleteProject } from '@/hooks/use-projects'
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
import { SimpleConfirmDialog } from '@/components/simple-confirm-dialog'
import { Separator } from '@/components/ui/separator'

const updateProjectSchema = z.object({
  name: z
    .string()
    .min(1, 'Project name is required')
    .max(100, 'Project name must be less than 100 characters'),
  description: z
    .string()
    .max(500, 'Description must be less than 500 characters')
    .optional(),
  repository_url: z
    .string()
    .refine((url) => {
      if (!url) return true // Optional field
      try {
        const parsedUrl = new URL(url)
        return ['http:', 'https:', 'git:', 'ssh:'].includes(parsedUrl.protocol)
      } catch {
        return false
      }
    }, 'Please enter a valid repository URL')
    .optional(),
  worktree_base_path: z.string().optional(),
})

type UpdateProjectFormData = z.infer<typeof updateProjectSchema>

interface ProjectEditModalProps {
  projectId: string
  open: boolean
  onOpenChange: (open: boolean) => void
  onDelete?: () => void
}

export function ProjectEditModal({
  projectId,
  open,
  onOpenChange,
  onDelete,
}: ProjectEditModalProps) {
  const { data: project } = useProject(projectId)
  const updateProject = useUpdateProject()
  const deleteProject = useDeleteProject()

  const form = useForm<UpdateProjectFormData>({
    resolver: zodResolver(updateProjectSchema),
    defaultValues: {
      name: '',
      description: '',
      repository_url: '',
      worktree_base_path: '',
    },
  })

  // Update form when project data loads
  useEffect(() => {
    if (project) {
      form.reset({
        name: project.name,
        description: project.description || '',
        repository_url: project.repository_url || '',
        worktree_base_path: project.worktree_base_path || '',
      })
    }
  }, [project, form])

  const onSubmit = async (data: UpdateProjectFormData) => {
    const updates: UpdateProjectRequest = {
      name: data.name,
      description: data.description || undefined,
      repository_url: data.repository_url || undefined,
      worktree_base_path: data.worktree_base_path || undefined,
    }

    try {
      await updateProject.mutateAsync({ projectId, updates })
      onOpenChange(false)
    } catch (error) {
      // Error handling is done in the mutation hook
    }
  }

  const handleDelete = async () => {
    try {
      await deleteProject.mutateAsync(projectId)
      onOpenChange(false)
      onDelete?.()
    } catch (error) {
      // Error handling is done in the mutation hook
    }
  }

  const handleClose = () => {
    onOpenChange(false)
    if (project) {
      form.reset({
        name: project.name,
        description: project.description || '',
        repository_url: project.repository_url || '',
        worktree_base_path: project.worktree_base_path || '',
      })
    }
  }

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent className='sm:max-w-[500px] max-h-[80vh] overflow-y-auto'>
        <DialogHeader>
          <DialogTitle className='flex items-center gap-2'>
            <Settings className='h-5 w-5' />
            Edit Project
          </DialogTitle>
          <DialogDescription>
            Update project settings and configuration
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
              name='repository_url'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Repository URL</FormLabel>
                  <FormControl>
                    <Input
                      placeholder='https://github.com/user/repo.git'
                      {...field}
                    />
                  </FormControl>
                  <FormDescription>
                    The Git repository URL (HTTPS or SSH)
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
                    <Input placeholder='/tmp/projects/repo' {...field} />
                  </FormControl>
                  <FormDescription>
                    Base path for Git worktree operations
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            <div className='flex justify-between items-center pt-4'>
              <SimpleConfirmDialog
                title='Delete Project'
                description='Are you sure you want to delete this project? This action cannot be undone.'
                onConfirm={handleDelete}
                destructive={true}
                confirmText='Delete Project'
                cancelText='Cancel'
              >
                <Button 
                  type='button' 
                  variant='destructive' 
                  disabled={deleteProject.isPending}
                >
                  {deleteProject.isPending ? (
                    'Deleting...'
                  ) : (
                    <>
                      <Trash2 className='mr-2 h-4 w-4' />
                      Delete
                    </>
                  )}
                </Button>
              </SimpleConfirmDialog>

              <div className='flex gap-3'>
                <Button type='button' variant='outline' onClick={handleClose}>
                  Cancel
                </Button>
                <Button type='submit' disabled={updateProject.isPending}>
                  {updateProject.isPending ? 'Saving...' : 'Save Changes'}
                </Button>
              </div>
            </div>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  )
}