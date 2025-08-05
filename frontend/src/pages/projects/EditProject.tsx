import { useEffect } from 'react'
import { z } from 'zod'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { useNavigate, useParams } from '@tanstack/react-router'
import type { UpdateProjectRequest } from '@/types/project'
import { ArrowLeft, GitBranch, Trash2, Info, GitFork } from 'lucide-react'
import {
  useProject,
  useUpdateProject,
  useDeleteProject,
} from '@/hooks/use-projects'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
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
import { Separator } from '@/components/ui/separator'
import { Skeleton } from '@/components/ui/skeleton'
import { Textarea } from '@/components/ui/textarea'
import { SimpleConfirmDialog } from '@/components/simple-confirm-dialog'

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

export function EditProject() {
  const navigate = useNavigate()
  const { projectId } = useParams({
    from: '/_authenticated/projects/$projectId/edit',
  })
  const { data: project, isLoading, error } = useProject(projectId)
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
      navigate({ to: '/projects/$projectId', params: { projectId } })
    } catch (error) {
      // Error handling is done in the mutation hook
    }
  }

  const onDelete = async () => {
    try {
      await deleteProject.mutateAsync(projectId)
      navigate({ to: '/projects' })
    } catch (error) {
      // Error handling is done in the mutation hook
    }
  }

  if (error) {
    return (
      <div className='flex h-full items-center justify-center'>
        <Alert className='max-w-md'>
          <Info className='h-4 w-4' />
          <AlertDescription>
            Failed to load project. Please try again.
          </AlertDescription>
        </Alert>
      </div>
    )
  }

  if (isLoading) {
    return (
      <div className='h-full space-y-6'>
        <div className='flex items-center gap-4'>
          <Skeleton className='h-10 w-10' />
          <div className='space-y-2'>
            <Skeleton className='h-8 w-64' />
            <Skeleton className='h-4 w-96' />
          </div>
        </div>
        <div className='max-w-2xl'>
          <Card>
            <CardHeader>
              <Skeleton className='h-6 w-48' />
              <Skeleton className='h-4 w-96' />
            </CardHeader>
            <CardContent className='space-y-6'>
              <Skeleton className='h-10 w-full' />
              <Skeleton className='h-20 w-full' />
              <Skeleton className='h-10 w-full' />
            </CardContent>
          </Card>
        </div>
      </div>
    )
  }

  return (
    <div className='h-full space-y-6'>
      {/* Header */}
      <div className='flex items-center justify-between'>
        <div className='flex items-center gap-4'>
          <Button
            variant='ghost'
            size='icon'
            onClick={() =>
              navigate({ to: '/projects/$projectId', params: { projectId } })
            }
          >
            <ArrowLeft className='h-4 w-4' />
          </Button>
          <div>
            <h1 className='text-3xl font-bold'>Edit Project</h1>
            <p className='text-muted-foreground'>
              Update project settings and configuration
            </p>
          </div>
        </div>

        <SimpleConfirmDialog
          title='Delete Project'
          description='Are you sure you want to delete this project? This action cannot be undone.'
          onConfirm={onDelete}
        >
          <Button variant='destructive' disabled={deleteProject.isPending}>
            {deleteProject.isPending ? (
              'Deleting...'
            ) : (
              <>
                <Trash2 className='mr-2 h-4 w-4' />
                Delete Project
              </>
            )}
          </Button>
        </SimpleConfirmDialog>
      </div>

      <div className='max-w-2xl'>
        <Card>
          <CardHeader>
            <CardTitle className='flex items-center gap-2'>
              <GitBranch className='h-5 w-5' />
              Project Details
            </CardTitle>
            <CardDescription>
              Update your project settings and Git integration configuration
            </CardDescription>
          </CardHeader>

          <CardContent>
            <Form {...form}>
              <form
                onSubmit={form.handleSubmit(onSubmit)}
                className='space-y-6'
              >
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

                <div className='flex gap-4'>
                  <Button
                    type='button'
                    variant='outline'
                    onClick={() =>
                      navigate({
                        to: '/projects/$projectId',
                        params: { projectId },
                      })
                    }
                  >
                    Cancel
                  </Button>
                  <Button type='submit' disabled={updateProject.isPending}>
                    {updateProject.isPending ? 'Saving...' : 'Save Changes'}
                  </Button>
                </div>
              </form>
            </Form>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
