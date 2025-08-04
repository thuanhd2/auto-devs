import { z } from 'zod'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { useNavigate } from '@tanstack/react-router'
import type { CreateProjectRequest } from '@/types/project'
import { ArrowLeft, GitBranch, Info, GitFork } from 'lucide-react'
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Separator } from '@/components/ui/separator'
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'

const createProjectSchema = z
  .object({
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
          return ['http:', 'https:', 'git:', 'ssh:'].includes(
            parsedUrl.protocol
          )
        } catch {
          return false
        }
      }, 'Please enter a valid repository URL'),

    // Git-related fields
    git_enabled: z.boolean(),
    repository_url: z.string().optional(),
    main_branch: z.string().optional(),
    worktree_base_path: z.string().optional(),
    git_auth_method: z.enum(['ssh', 'https']).optional(),
  })
  .refine(
    (data) => {
      // If Git is enabled, validate required Git fields
      if (data.git_enabled) {
        if (!data.repository_url) {
          return false
        }
        if (!data.main_branch) {
          return false
        }
        if (!data.worktree_base_path) {
          return false
        }
        if (!data.git_auth_method) {
          return false
        }
      }
      return true
    },
    {
      message: 'Git configuration is incomplete',
      path: ['git_enabled'],
    }
  )

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
      git_enabled: false,
      repository_url: '',
      main_branch: 'main',
      worktree_base_path: '',
      git_auth_method: 'https',
    },
  })

  const gitEnabled = form.watch('git_enabled')

  const onSubmit = async (data: CreateProjectFormData) => {
    const projectData: CreateProjectRequest = {
      name: data.name,
      description: data.description || undefined,
      repo_url: data.repo_url,
      git_enabled: data.git_enabled,
      repository_url: data.repository_url,
      main_branch: data.main_branch,
      worktree_base_path: data.worktree_base_path,
      git_auth_method: data.git_auth_method,
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
            <CardTitle className='flex items-center gap-2'>
              <GitBranch className='h-5 w-5' />
              Project Details
            </CardTitle>
            <CardDescription>
              Configure your project settings and optionally enable Git
              integration
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
                        <Input
                          placeholder='https://github.com/user/repo'
                          {...field}
                        />
                      </FormControl>
                      <FormDescription>
                        The URL of your project repository
                      </FormDescription>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <Separator />

                {/* Git Integration Section */}
                <div className='space-y-4'>
                  <div className='flex items-center gap-2'>
                    <GitFork className='h-4 w-4' />
                    <h3 className='text-lg font-semibold'>Git Integration</h3>
                  </div>

                  <FormField
                    control={form.control}
                    name='git_enabled'
                    render={({ field }) => (
                      <FormItem className='flex flex-row items-center justify-between rounded-lg border p-4'>
                        <div className='space-y-0.5'>
                          <FormLabel className='text-base'>
                            Enable Git Integration
                          </FormLabel>
                          <FormDescription>
                            Enable advanced Git features for this project
                          </FormDescription>
                        </div>
                        <FormControl>
                          <Switch
                            checked={field.value}
                            onCheckedChange={field.onChange}
                          />
                        </FormControl>
                      </FormItem>
                    )}
                  />

                  {gitEnabled && (
                    <div className='space-y-4 rounded-lg border p-4'>
                      <FormField
                        control={form.control}
                        name='repository_url'
                        render={({ field }) => (
                          <FormItem>
                            <FormLabel>Git Repository URL</FormLabel>
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

                      <div className='grid grid-cols-2 gap-4'>
                        <FormField
                          control={form.control}
                          name='main_branch'
                          render={({ field }) => (
                            <FormItem>
                              <FormLabel>Main Branch</FormLabel>
                              <FormControl>
                                <Input placeholder='main' {...field} />
                              </FormControl>
                              <FormDescription>
                                Default branch name
                              </FormDescription>
                              <FormMessage />
                            </FormItem>
                          )}
                        />

                        <FormField
                          control={form.control}
                          name='git_auth_method'
                          render={({ field }) => (
                            <FormItem>
                              <FormLabel>Authentication Method</FormLabel>
                              <Select
                                onValueChange={field.onChange}
                                defaultValue={field.value}
                              >
                                <FormControl>
                                  <SelectTrigger>
                                    <SelectValue placeholder='Select auth method' />
                                  </SelectTrigger>
                                </FormControl>
                                <SelectContent>
                                  <SelectItem value='https'>HTTPS</SelectItem>
                                  <SelectItem value='ssh'>SSH</SelectItem>
                                </SelectContent>
                              </Select>
                              <FormDescription>
                                Choose authentication method
                              </FormDescription>
                              <FormMessage />
                            </FormItem>
                          )}
                        />
                      </div>

                      <FormField
                        control={form.control}
                        name='worktree_base_path'
                        render={({ field }) => (
                          <FormItem>
                            <FormLabel>Worktree Base Path</FormLabel>
                            <FormControl>
                              <Input
                                placeholder='/tmp/projects/repo'
                                {...field}
                              />
                            </FormControl>
                            <FormDescription>
                              Base path for Git worktree operations
                            </FormDescription>
                            <FormMessage />
                          </FormItem>
                        )}
                      />

                      <Alert>
                        <Info className='h-4 w-4' />
                        <AlertDescription>
                          Git integration requires proper authentication setup.
                          Make sure you have SSH keys configured for SSH
                          authentication or use HTTPS with appropriate
                          credentials.
                        </AlertDescription>
                      </Alert>
                    </div>
                  )}
                </div>

                <div className='flex gap-4'>
                  <Button
                    type='button'
                    variant='outline'
                    onClick={() => navigate({ to: '/projects' })}
                  >
                    Cancel
                  </Button>
                  <Button type='submit' disabled={createProject.isPending}>
                    {createProject.isPending ? 'Creating...' : 'Create Project'}
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
