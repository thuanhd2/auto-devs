import { useEffect } from 'react'
import { z } from 'zod'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { useNavigate, useParams } from '@tanstack/react-router'
import type { UpdateProjectRequest } from '@/types/project'
import { ArrowLeft, GitBranch, Trash2, Info } from 'lucide-react'
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
  repo_url: z
    .string()
    .min(1, 'Repository URL is required')
    .refine((url) => {
      try {
        const parsedUrl = new URL(url)
        return ['http:', 'https:', 'git:', 'ssh:'].includes(parsedUrl.protocol)
      } catch {
        return false
      }
    }, 'Please enter a valid repository URL'),
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
      repo_url: '',
    },
  })

  // Update form when project data loads
  useEffect(() => {
    if (project) {
      form.reset({
        name: project.name,
        description: project.description || '',
        repo_url: project.repo_url,
      })
    }
  }, [project, form])

  const onSubmit = async (data: UpdateProjectFormData) => {
    const updates: UpdateProjectRequest = {
      name: data.name,
      description: data.description || undefined,
      repo_url: data.repo_url,
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
        <div className='text-center'>
          <h3 className='text-lg font-semibold'>Error loading project</h3>
          <p className='text-muted-foreground mb-4'>
            {error instanceof Error
              ? error.message
              : 'An unexpected error occurred'}
          </p>
          <Button
            variant='outline'
            onClick={() => navigate({ to: '/projects' })}
          >
            <ArrowLeft className='mr-2 h-4 w-4' />
            Back to Projects
          </Button>
        </div>
      </div>
    )
  }

  if (isLoading) {
    return <EditProjectSkeleton />
  }

  if (!project) {
    return (
      <div className='flex h-full items-center justify-center'>
        <div className='text-center'>
          <h3 className='text-lg font-semibold'>Project not found</h3>
          <p className='text-muted-foreground mb-4'>
            The project you're trying to edit doesn't exist or has been deleted.
          </p>
          <Button
            variant='outline'
            onClick={() => navigate({ to: '/projects' })}
          >
            <ArrowLeft className='mr-2 h-4 w-4' />
            Back to Projects
          </Button>
        </div>
      </div>
    )
  }

  return (
    <div className='h-full space-y-6'>
      {/* Header */}
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

      <div className='max-w-2xl space-y-6'>
        <Card>
          <CardHeader>
            <CardTitle>Project Information</CardTitle>
            <CardDescription>
              Update basic information about your project and repository
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
                  name='repo_url'
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Repository URL</FormLabel>
                      <FormControl>
                        <div className='relative'>
                          <GitBranch className='text-muted-foreground absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2' />
                          <Input
                            placeholder='https://github.com/username/repo.git'
                            className='pl-9'
                            {...field}
                          />
                        </div>
                      </FormControl>
                      <FormDescription>
                        Git repository URL (HTTPS, SSH, or Git protocol)
                      </FormDescription>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <Alert>
                  <Info className='h-4 w-4' />
                  <AlertDescription>
                    Changing the repository URL or branch may affect existing
                    tasks and their associated branches.
                  </AlertDescription>
                </Alert>

                <div className='flex gap-3 pt-6'>
                  <Button
                    type='submit'
                    disabled={updateProject.isPending}
                    className='flex-1 sm:flex-none'
                  >
                    {updateProject.isPending ? 'Updating...' : 'Update Project'}
                  </Button>
                  <Button
                    type='button'
                    variant='outline'
                    onClick={() =>
                      navigate({
                        to: '/projects/$projectId',
                        params: { projectId },
                      })
                    }
                    disabled={updateProject.isPending}
                  >
                    Cancel
                  </Button>
                </div>
              </form>
            </Form>
          </CardContent>
        </Card>

        {/* Danger Zone */}
        <Card className='border-destructive'>
          <CardHeader>
            <CardTitle className='text-destructive'>Danger Zone</CardTitle>
            <CardDescription>
              Irreversible and destructive actions
            </CardDescription>
          </CardHeader>

          <CardContent>
            <div className='flex items-center justify-between rounded-lg border p-4'>
              <div>
                <h4 className='font-medium'>Delete Project</h4>
                <p className='text-muted-foreground text-sm'>
                  Permanently delete this project and all associated tasks
                </p>
              </div>
              <SimpleConfirmDialog
                title='Delete Project'
                description={`Are you sure you want to delete "${project.name}"? This action cannot be undone and will permanently delete all tasks, branches, and associated data.`}
                onConfirm={onDelete}
                destructive
              >
                <Button
                  variant='destructive'
                  size='sm'
                  disabled={deleteProject.isPending}
                >
                  <Trash2 className='mr-2 h-4 w-4' />
                  {deleteProject.isPending ? 'Deleting...' : 'Delete'}
                </Button>
              </SimpleConfirmDialog>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}

function EditProjectSkeleton() {
  return (
    <div className='h-full space-y-6'>
      <div className='flex items-center gap-4'>
        <Skeleton className='h-10 w-10' />
        <div className='space-y-2'>
          <Skeleton className='h-8 w-48' />
          <Skeleton className='h-4 w-64' />
        </div>
      </div>

      <div className='max-w-2xl space-y-6'>
        <Card>
          <CardHeader>
            <Skeleton className='h-6 w-40' />
            <Skeleton className='h-4 w-64' />
          </CardHeader>

          <CardContent className='space-y-6'>
            {Array.from({ length: 4 }).map((_, i) => (
              <div key={i} className='space-y-2'>
                <Skeleton className='h-4 w-24' />
                <Skeleton className='h-10 w-full' />
                <Skeleton className='h-4 w-48' />
              </div>
            ))}

            <div className='flex gap-3 pt-6'>
              <Skeleton className='h-10 w-32' />
              <Skeleton className='h-10 w-20' />
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
