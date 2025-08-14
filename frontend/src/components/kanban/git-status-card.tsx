import { GitFork, Loader2 } from 'lucide-react'
import { useReinitGitRepository } from '@/hooks/use-projects'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'

interface GitStatusCardProps {
  projectId: string
}

export function GitStatusCard({ projectId }: GitStatusCardProps) {
  const { mutate: reinitGit, isPending } = useReinitGitRepository()

  return (
    <Card>
      <CardHeader>
        <CardTitle className='flex items-center gap-2'>
          <GitFork className='h-5 w-5' />
          Git Integration Status
        </CardTitle>
        <CardDescription>
          Current status of Git integration for this project
        </CardDescription>
      </CardHeader>

      <CardContent className='space-y-4'>
        <Button
          onClick={() => reinitGit(projectId)}
          disabled={isPending}
          variant='secondary'
          size='sm'
        >
          {isPending ? (
            <>
              <Loader2 className='mr-2 h-4 w-4 animate-spin' />
              Reinitializing...
            </>
          ) : (
            'Reload'
          )}
        </Button>
      </CardContent>
    </Card>
  )
}
