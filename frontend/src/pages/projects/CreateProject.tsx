import { z } from 'zod'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { useNavigate } from '@tanstack/react-router'
import type { CreateProjectRequest } from '@/types/project'
import { ArrowLeft, GitBranch, Info } from 'lucide-react'
import { useCreateProject } from '@/hooks/use-projects'
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

type CreateProjectFormData = z.infer<typeof createProjectSchema>

export function CreateProject() {
  const navigate = useNavigate()
  const createProject = useCreateProject()

  const form = useForm<CreateProjectFormData>({
    resolver: zodResolver(createProjectSchema),
    defaultValues: {
      name: '',
      description: '',
      repo_url: '',
    },
  })

  const onSubmit = async (data: CreateProjectFormData) => {
    const projectData: CreateProjectRequest = {
      name: data.name,
      description: data.description || undefined,
      repo_url: data.repo_url,
    }

    try {
      const project = await createProject.mutateAsync(projectData)
      navigate({
        to: '/projects/$projectId',
        params: { projectId: project.id },
      })
    } catch (error) {
      // Error handling is done in the mutation hook
    }
  }

  return (
    <div className='h-full space-y-6'>
      {/* Header */}
      <div className='flex items-center gap-4'>
        <Button
          variant='ghost'
          size='icon'
          onClick={() => navigate({ to: '/projects' })}
        >
          <ArrowLeft className='h-4 w-4' />
        </Button>
        <div>
          <h1 className='text-3xl font-bold'>Create Project</h1>
          <p className='text-muted-foreground'>
            Set up a new development project to start managing tasks
          </p>
        </div>
      </div>

      <div className='max-w-2xl'>
        <Card>
          <CardHeader>
            <CardTitle>Project Information</CardTitle>
            <CardDescription>
              Provide basic information about your project and repository
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
                    The system will use this repository to create isolated
                    branches for each task. Make sure the repository URL is
                    accessible and you have appropriate permissions.
                  </AlertDescription>
                </Alert>

                <div className='flex gap-3 pt-6'>
                  <Button
                    type='submit'
                    disabled={createProject.isPending}
                    className='flex-1 sm:flex-none'
                  >
                    {createProject.isPending ? 'Creating...' : 'Create Project'}
                  </Button>
                  <Button
                    type='button'
                    variant='outline'
                    onClick={() => navigate({ to: '/projects' })}
                    disabled={createProject.isPending}
                  >
                    Cancel
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
