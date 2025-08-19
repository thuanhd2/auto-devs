import { useState, useEffect, useCallback } from 'react'
import { getAIs } from '@/types/task'
import { Loader2, GitBranch, Bot } from 'lucide-react'
import { projectsApi } from '@/lib/api/projects'
import { Alert, AlertDescription } from '@/components/ui/alert'
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
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

interface GitBranch {
  name: string
  is_current: boolean
  last_commit: string
  last_updated: string
}

interface BranchSelectionDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  projectId: string
  taskTitle: string
  onBranchSelected: (branchName: string, aiType: string) => void
}

export function BranchSelectionDialog({
  open,
  onOpenChange,
  projectId,
  taskTitle,
  onBranchSelected,
}: BranchSelectionDialogProps) {
  const [branches, setBranches] = useState<GitBranch[]>([])
  const [selectedBranch, setSelectedBranch] = useState<string>('')
  const [selectedAIType, setSelectedAIType] = useState<string>('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string>('')

  // Load AI type preference from localStorage
  useEffect(() => {
    const savedPlanningAI = localStorage.getItem('ai_preference_planning')
    setSelectedAIType(savedPlanningAI || 'claude-code') // Default to claude-code
  }, [])

  const fetchBranches = useCallback(async () => {
    setLoading(true)
    setError('')

    try {
      const data = await projectsApi.getProjectBranches(projectId)
      setBranches(data.branches || [])

      // Auto-select current branch if available
      const currentBranch = data.branches?.find(
        (branch: GitBranch) => branch.is_current
      )
      if (currentBranch) {
        setSelectedBranch(currentBranch.name)
      }
    } catch (err) {
      setError('Failed to load branches. Please try again.')
      console.error('Error fetching branches:', err)
    } finally {
      setLoading(false)
    }
  }, [projectId])

  useEffect(() => {
    if (open && projectId) {
      fetchBranches()
    }
  }, [open, projectId, fetchBranches])

  const handleConfirm = () => {
    if (selectedBranch && selectedAIType) {
      // Save AI type preference to localStorage
      localStorage.setItem('ai_preference_planning', selectedAIType)

      onBranchSelected(selectedBranch, selectedAIType)
      onOpenChange(false)
      setSelectedBranch('')
      setSelectedAIType(
        localStorage.getItem('ai_preference_planning') || 'claude-code'
      )
    }
  }

  const handleCancel = () => {
    onOpenChange(false)
    setSelectedBranch('')
    setSelectedAIType(
      localStorage.getItem('ai_preference_planning') || 'claude-code'
    )
    setError('')
  }

  const ais = getAIs(true)

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='sm:max-w-[425px]'>
        <DialogHeader>
          <DialogTitle className='flex items-center gap-2'>
            <GitBranch className='h-5 w-5' />
            Start Planning
          </DialogTitle>
          <DialogDescription>
            Select a branch to start planning for task:{' '}
            <strong>{taskTitle}</strong>
          </DialogDescription>
        </DialogHeader>

        <div className='space-y-4'>
          {error && (
            <Alert variant='destructive'>
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}

          {loading ? (
            <div className='flex items-center justify-center py-8'>
              <Loader2 className='h-6 w-6 animate-spin' />
              <span className='ml-2'>Loading branches...</span>
            </div>
          ) : (
            <>
              <div className='space-y-2'>
                <label className='text-sm font-medium'>Select Branch:</label>
                <Select
                  value={selectedBranch}
                  onValueChange={setSelectedBranch}
                >
                  <SelectTrigger>
                    <SelectValue placeholder='Select a branch' />
                  </SelectTrigger>
                  <SelectContent>
                    {branches.map((branch) => (
                      <SelectItem key={branch.name} value={branch.name}>
                        <div className='flex items-center gap-2'>
                          <span>{branch.name}</span>
                          {branch.is_current && (
                            <span className='text-muted-foreground text-xs'>
                              (current)
                            </span>
                          )}
                        </div>
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>

                {branches.length === 0 && !loading && (
                  <p className='text-muted-foreground text-sm'>
                    No branches found in the repository.
                  </p>
                )}
              </div>

              <div className='space-y-2'>
                <label className='flex items-center gap-2 text-sm font-medium'>
                  <Bot className='h-4 w-4' />
                  Select AI Assistant:
                </label>
                <Select
                  value={selectedAIType}
                  onValueChange={setSelectedAIType}
                >
                  <SelectTrigger>
                    <SelectValue placeholder='Select AI type' />
                  </SelectTrigger>
                  <SelectContent>
                    {ais.map((ai) => (
                      <SelectItem key={ai.value} value={ai.value}>
                        <div className='flex items-center gap-2'>
                          <span>{ai.name}</span>
                          <span className='text-muted-foreground text-xs'>
                            ({ai.description})
                          </span>
                        </div>
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </>
          )}
        </div>

        <DialogFooter>
          <Button variant='outline' onClick={handleCancel}>
            Cancel
          </Button>
          <Button
            onClick={handleConfirm}
            disabled={!selectedBranch || !selectedAIType || loading}
          >
            Start Planning
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
