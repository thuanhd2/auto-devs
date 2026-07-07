import { useState, useEffect } from 'react'
import { getAIs } from '@/types/task'
import { Bot, Play } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
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

interface ImplementationConfirmationDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  taskTitle: string
  onConfirm: (aiType: string, autoImplement?: boolean) => void
  mode?: 'implementing' | 'planning'
}

export function ImplementationConfirmationDialog({
  open,
  onOpenChange,
  taskTitle,
  onConfirm,
  mode,
}: ImplementationConfirmationDialogProps) {
  const [selectedAIType, setSelectedAIType] = useState<string>('')
  const [autoImplement, setAutoImplement] = useState(false)

  const localStorageKey =
    mode === 'planning' ? 'ai_preference_planning' : 'ai_preference_implementing'

  // Load AI type preference from localStorage
  useEffect(() => {
    const savedAI =
      localStorage.getItem(localStorageKey) ||
      localStorage.getItem('ai_preference_implementing') ||
      localStorage.getItem('ai_preference_planning') ||
      'claude-code'
    setSelectedAIType(savedAI)
  }, [localStorageKey])

  const handleConfirm = () => {
    if (selectedAIType) {
      localStorage.setItem(localStorageKey, selectedAIType)
      onConfirm(selectedAIType, mode === 'planning' ? autoImplement : undefined)
      onOpenChange(false)
      setAutoImplement(false)
    }
  }

  const handleCancel = () => {
    onOpenChange(false)
    setAutoImplement(false)
  }

  const isPlanning = mode === 'planning'
  const ais = getAIs(isPlanning)

  const title = isPlanning ? 'Start Planning' : 'Approve Plan and Start Implementation'
  const description = isPlanning
    ? `Select AI assistant to start planning for task:`
    : `Approve the plan and start implementing for task:`
  const confirmLabel = isPlanning ? 'Start Planning' : 'Approve and Start Implementation'

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='sm:max-w-[425px]'>
        <DialogHeader>
          <DialogTitle className='flex items-center gap-2'>
            <Play className='h-5 w-5' />
            {title}
          </DialogTitle>
          <DialogDescription>
            {description} <strong>{taskTitle}</strong>
          </DialogDescription>
        </DialogHeader>

        <div className='space-y-4'>
          <div className='space-y-2'>
            <label className='flex items-center gap-2 text-sm font-medium'>
              <Bot className='h-4 w-4' />
              Select AI Assistant:
            </label>
            <Select value={selectedAIType} onValueChange={setSelectedAIType}>
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

          {isPlanning && (
            <div className='flex items-center space-x-2 pt-2'>
              <Checkbox
                id='auto-implement-confirmation'
                checked={autoImplement}
                onCheckedChange={(checked) => setAutoImplement(checked === true)}
              />
              <label
                htmlFor='auto-implement-confirmation'
                className='text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70'
              >
                Auto implement after planning complete
              </label>
            </div>
          )}
        </div>

        <DialogFooter>
          <Button variant='outline' onClick={handleCancel}>
            Cancel
          </Button>
          <Button
            onClick={handleConfirm}
            disabled={!selectedAIType}
            className={isPlanning ? 'bg-blue-600 hover:bg-blue-700 text-white' : undefined}
          >
            {confirmLabel}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
